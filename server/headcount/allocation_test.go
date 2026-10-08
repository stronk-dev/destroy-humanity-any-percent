package headcount

import (
	"encoding/json"
	"errors"
	"math/big"
	"os"
	"reflect"
	"testing"
)

type vector struct {
	Name     string           `json:"name"`
	Seats    int64            `json:"seats"`
	K        int64            `json:"k"`
	Assigned map[string]int64 `json:"assigned"`
	Expected Allocation       `json:"expected"`
}

func vectors(t *testing.T) ([]Role, []vector) {
	t.Helper()
	data, err := os.ReadFile("../../testdata/headcount-allocation.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		SchemaVersion int      `json:"schema_version"`
		Hardcap       int64    `json:"hardcap"`
		Roles         []Role   `json:"roles"`
		Cases         []vector `json:"effective_heads"`
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.SchemaVersion != 1 || fixture.Hardcap != maxSafeInteger || len(fixture.Cases) != 12 {
		t.Fatal("unexpected corpus")
	}
	return fixture.Roles, fixture.Cases
}

func TestSharedEffectiveHeads(t *testing.T) {
	roles, cases := vectors(t)
	for _, v := range cases {
		t.Run(v.Name, func(t *testing.T) {
			input := append([]Role(nil), roles...)
			input[0].HeadsPerConverter = v.K
			before, _ := json.Marshal([]any{input, v.Assigned})
			got, err := Project(maxSafeInteger, v.Seats, input, v.Assigned)
			if err != nil || !reflect.DeepEqual(got, v.Expected) {
				t.Fatalf("got %+v, %v; want %+v", got, err, v.Expected)
			}
			after, _ := json.Marshal([]any{input, v.Assigned})
			if string(before) != string(after) {
				t.Fatal("mutated input")
			}
			input[0], input[2] = input[2], input[0]
			reordered, err := Project(maxSafeInteger, v.Seats, input, v.Assigned)
			if err != nil || !reflect.DeepEqual(got, reordered) {
				t.Fatal("role order changes result")
			}
		})
	}
}

// Independent arbitrary-precision model, not Project's overflow-avoidance branch.
func exactModel(e, m, f, k, seats int64) Allocation {
	product := new(big.Int).Mul(big.NewInt(m), big.NewInt(k))
	if product.Cmp(big.NewInt(e)) > 0 {
		product.SetInt64(e)
	}
	used, remainder := new(big.Int), new(big.Int)
	used.QuoRem(product, big.NewInt(k), remainder)
	if remainder.Sign() != 0 {
		used.Add(used, big.NewInt(1))
	}
	return Allocation{
		ProducingHeads: map[string]int64{"role.producer": e - product.Int64(), "role.feeder": f},
		Diverted:       map[string]int64{"role.converter": product.Int64()},
		IdleConverters: map[string]int64{"role.converter": m - used.Int64()},
		Unassigned:     seats - e - m - f,
	}
}

func TestExactSmallDomain(t *testing.T) {
	roles, _ := vectors(t)
	for e := int64(0); e <= 16; e++ {
		for m := int64(0); m <= 16; m++ {
			for f := int64(0); f <= 2; f++ {
				for k := int64(1); k <= 9; k++ {
					roles[0].HeadsPerConverter = k
					seats := e + m + f + 2
					got, err := Project(100, seats, roles, map[string]int64{"role.producer": e, "role.converter": m, "role.feeder": f})
					if err != nil || !reflect.DeepEqual(got, exactModel(e, m, f, k, seats)) {
						t.Fatalf("E=%d M=%d F=%d k=%d: %+v %v", e, m, f, k, got, err)
					}
				}
			}
		}
	}
}

func TestRefusesInvalidAllocations(t *testing.T) {
	base, _ := vectors(t)
	type input struct {
		cap, seats int64
		roles      []Role
		assigned   map[string]int64
	}
	cases := map[string]func(*input){
		"zero cap":            func(v *input) { v.cap = 0 },
		"unsafe cap":          func(v *input) { v.cap = maxSafeInteger + 1 },
		"negative seats":      func(v *input) { v.seats = -1 },
		"over cap":            func(v *input) { v.seats = 13 },
		"empty roles":         func(v *input) { v.roles = nil },
		"missing assignment":  func(v *input) { delete(v.assigned, "role.feeder") },
		"unknown assignment":  func(v *input) { delete(v.assigned, "role.feeder"); v.assigned["role.other"] = 0 },
		"negative assignment": func(v *input) { v.assigned["role.feeder"] = -1 },
		"unsafe assignment":   func(v *input) { v.assigned["role.feeder"] = maxSafeInteger + 1 },
		"over seat budget":    func(v *input) { v.assigned["role.feeder"] = 2 },
		"duplicate roles":     func(v *input) { v.roles[2] = v.roles[1] },
		"invalid ID":          func(v *input) { v.roles[1].ID = "Bad" },
		"unknown kind":        func(v *input) { v.roles[1].Kind = "unknown" },
		"zero ratio":          func(v *input) { v.roles[0].HeadsPerConverter = 0 },
		"unsafe ratio":        func(v *input) { v.roles[0].HeadsPerConverter = maxSafeInteger + 1 },
		"cross-pool source":   func(v *input) { v.roles[0].FromRoleID = "role.other" },
		"self source":         func(v *input) { v.roles[0].FromRoleID = "role.converter" },
		"feeder source":       func(v *input) { v.roles[0].FromRoleID = "role.feeder" },
		"conversion chain": func(v *input) {
			v.roles[0].FromRoleID = "role.feeder"
			v.roles[1] = Role{ID: "role.feeder", Kind: "convert", FromRoleID: "role.producer", HeadsPerConverter: 1}
		},
		"two converters": func(v *input) {
			v.roles[1] = Role{ID: "role.feeder", Kind: "convert", FromRoleID: "role.producer", HeadsPerConverter: 1}
		},
		"nonconverter conversion metadata": func(v *input) { v.roles[1].HeadsPerConverter = 1 },
		"RFC overflow example exceeds safe seat budget": func(v *input) {
			v.cap, v.seats = maxSafeInteger, maxSafeInteger
			v.roles[0].HeadsPerConverter = 3
			v.assigned["role.converter"], v.assigned["role.producer"] = 3100000000000000, 9000000000000000
		},
	}
	for name, corrupt := range cases {
		t.Run(name, func(t *testing.T) {
			v := input{12, 12, append([]Role(nil), base...), map[string]int64{"role.converter": 3, "role.feeder": 0, "role.producer": 8}}
			corrupt(&v)
			before, _ := json.Marshal([]any{v.roles, v.assigned})
			got, err := Project(v.cap, v.seats, v.roles, v.assigned)
			if !errors.Is(err, ErrInvalidAllocation) || !reflect.DeepEqual(got, Allocation{}) {
				t.Fatalf("invalid input produced %+v, %v", got, err)
			}
			after, _ := json.Marshal([]any{v.roles, v.assigned})
			if string(before) != string(after) {
				t.Fatal("rejection mutated inputs")
			}
		})
	}
}

func TestExactMaximumDomain(t *testing.T) {
	roles, _ := vectors(t)
	for _, e := range []int64{0, 1, 2, 5900000000000000, maxSafeInteger - 100, maxSafeInteger - 1, maxSafeInteger} {
		for _, m := range []int64{0, min(1, maxSafeInteger-e), maxSafeInteger - e} {
			for _, k := range []int64{1, 2, 3, maxSafeInteger} {
				roles[0].HeadsPerConverter = k
				got, err := Project(maxSafeInteger, maxSafeInteger, roles, map[string]int64{"role.producer": e, "role.converter": m, "role.feeder": 0})
				if err != nil || !reflect.DeepEqual(got, exactModel(e, m, 0, k, maxSafeInteger)) {
					t.Fatalf("E=%d M=%d k=%d: %+v %v", e, m, k, got, err)
				}
			}
		}
	}
}

func TestFreshResults(t *testing.T) {
	roles, cases := vectors(t)
	v := cases[0]
	got, err := Project(maxSafeInteger, v.Seats, roles, v.Assigned)
	if err != nil {
		t.Fatal(err)
	}
	got.ProducingHeads["role.producer"] = 999
	next, err := Project(maxSafeInteger, v.Seats, roles, v.Assigned)
	if err != nil || !reflect.DeepEqual(next, v.Expected) {
		t.Fatal("shared mutable output")
	}
}

func TestIndependentConversions(t *testing.T) {
	roles := []Role{
		{ID: "role.convert_z", Kind: "convert", FromRoleID: "role.source_a", HeadsPerConverter: 2},
		{ID: "role.source_b", Kind: "produce"},
		{ID: "role.feed", Kind: "pool_feed"},
		{ID: "role.convert_a", Kind: "convert", FromRoleID: "role.source_b", HeadsPerConverter: 3},
		{ID: "role.source_a", Kind: "produce"},
	}
	assigned := map[string]int64{"role.convert_z": 4, "role.source_b": 7, "role.feed": 2, "role.convert_a": 2, "role.source_a": 3}
	want := Allocation{
		ProducingHeads: map[string]int64{"role.source_a": 0, "role.source_b": 1, "role.feed": 2},
		Diverted:       map[string]int64{"role.convert_a": 6, "role.convert_z": 3},
		IdleConverters: map[string]int64{"role.convert_a": 0, "role.convert_z": 2}, Unassigned: 2,
	}
	got, err := Project(20, 20, roles, assigned)
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("%+v %v", got, err)
	}
}

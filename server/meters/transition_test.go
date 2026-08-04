package meters

import (
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"
)

func testCatalog(t *testing.T) *Catalog {
	t.Helper()
	catalog, err := LoadCatalog(validCatalogBytes(t))
	if err != nil {
		t.Fatal(err)
	}
	return catalog
}

func TestAdvanceIsPartitionInvariantAndOfflineStable(t *testing.T) {
	catalog := testCatalog(t)
	whole := NewState(catalog)
	split := NewState(catalog)
	active := map[string]bool{ContributionKey("upgrades", "generator.example"): true}
	if _, err := Advance(catalog, whole, AdvanceContext{AttendedMS: MillisPerHour, NewFactKinds: map[string]bool{}, ActiveContributions: active}); err != nil {
		t.Fatal(err)
	}
	for _, elapsed := range []int64{1_234_567, MillisPerHour - 1_234_567} {
		if _, err := Advance(catalog, split, AdvanceContext{AttendedMS: elapsed, NewFactKinds: map[string]bool{}, ActiveContributions: active}); err != nil {
			t.Fatal(err)
		}
	}
	if whole.Values["doom.probability"] != split.Values["doom.probability"] ||
		whole.DecayRemainders["doom.probability"] != split.DecayRemainders["doom.probability"] ||
		whole.InputRemainders["doom.probability:1"] != split.InputRemainders["doom.probability:1"] {
		t.Fatalf("partition drift: whole=%+v split=%+v", whole, split)
	}
	offline := NewState(catalog)
	before := offline.Values["doom.probability"]
	if _, err := Advance(catalog, offline, AdvanceContext{AttendedMS: 0, NewFactKinds: map[string]bool{}, ActiveContributions: active}); err != nil {
		t.Fatal(err)
	}
	if offline.Values["doom.probability"] != before {
		t.Fatal("offline evaluation moved a meter")
	}
}

func TestAdvanceAggregatesCausalInputsAndEmitsFinalBand(t *testing.T) {
	catalog := testCatalog(t)
	state := NewState(catalog)
	state.Values["doom.probability"] = 69
	changes, err := Advance(catalog, state, AdvanceContext{
		AttendedMS: 0, NewFactKinds: map[string]bool{"externality.emitted": true}, ActiveContributions: map[string]bool{},
	})
	if err != nil {
		t.Fatal(err)
	}
	if state.Values["doom.probability"] != 72 || len(changes) != 1 || changes[0] != (BandChange{MeterID: "doom.probability", FromBand: "low", ToBand: "high", Direction: "up", ValueBefore: 69, ValueAfter: 72}) {
		t.Fatalf("unexpected transition: value=%d changes=%+v", state.Values["doom.probability"], changes)
	}
	if len(state.Values) != 11 {
		t.Fatal("unrelated meter state changed shape")
	}
}

func TestValidateStateRequiresCompleteExactMaps(t *testing.T) {
	catalog := testCatalog(t)
	state := NewState(catalog)
	delete(state.Values, "trust.users.standing")
	if err := ValidateState(catalog, state); err == nil {
		t.Fatal("missing meter value accepted")
	}
	state = NewState(catalog)
	state.InputRemainders["extra:0"] = 0
	if err := ValidateState(catalog, state); err == nil {
		t.Fatal("extra input remainder accepted")
	}
}

func TestNewRunStateReseedsEveryStandingAxisOnly(t *testing.T) {
	catalog := testCatalog(t)
	state, err := NewRunState(catalog, 100)
	if err != nil {
		t.Fatal(err)
	}
	for _, meter := range catalog.Meters {
		got := state.Values[meter.ID]
		if strings.HasSuffix(meter.ID, ".standing") {
			if got != 55 {
				t.Fatalf("%s = %d, want reseed 55", meter.ID, got)
			}
		} else if got != meter.InitialValue {
			t.Fatalf("%s = %d, want literal initial %d", meter.ID, got, meter.InitialValue)
		}
	}
	maxed, err := NewRunState(catalog, maximumExactInteger)
	if err != nil {
		t.Fatal(err)
	}
	if maxed.Values["trust.users.standing"] != catalog.TrustReseed.FloorValue {
		t.Fatalf("overflow-safe reseed = %d", maxed.Values["trust.users.standing"])
	}
	if _, err := NewRunState(catalog, -1); err == nil {
		t.Fatal("negative notoriety accepted")
	}
}

func TestTransitionSharedCorpus(t *testing.T) {
	type step struct {
		AttendedMS          int64    `json:"attended_ms"`
		NewFactKinds        []string `json:"new_fact_kinds"`
		ActiveContributions []string `json:"active_contributions"`
	}
	type testCase struct {
		Name                   string       `json:"name"`
		InitialValue           int          `json:"initial_value"`
		InitialDecayRemainder  int64        `json:"initial_decay_remainder"`
		InitialInputRemainder  int64        `json:"initial_input_remainder"`
		Steps                  []step       `json:"steps"`
		ExpectedValue          int          `json:"expected_value"`
		ExpectedDecayRemainder int64        `json:"expected_decay_remainder"`
		ExpectedInputRemainder int64        `json:"expected_input_remainder"`
		ExpectedChanges        []BandChange `json:"expected_changes"`
	}
	var corpus struct {
		Version int        `json:"version"`
		Cases   []testCase `json:"cases"`
	}
	data, err := os.ReadFile("../../balance/testdata/meters-transition-v1.json")
	if err != nil || json.Unmarshal(data, &corpus) != nil || corpus.Version != 1 || len(corpus.Cases) == 0 {
		t.Fatalf("load shared corpus: %v", err)
	}
	catalog := testCatalog(t)
	for _, vector := range corpus.Cases {
		t.Run(vector.Name, func(t *testing.T) {
			state := NewState(catalog)
			state.Values["doom.probability"] = vector.InitialValue
			state.DecayRemainders["doom.probability"] = vector.InitialDecayRemainder
			state.InputRemainders["doom.probability:1"] = vector.InitialInputRemainder
			changes := []BandChange{}
			for _, item := range vector.Steps {
				stepChanges, advanceErr := Advance(catalog, state, AdvanceContext{AttendedMS: item.AttendedMS, NewFactKinds: boolSet(item.NewFactKinds), ActiveContributions: boolSet(item.ActiveContributions)})
				if advanceErr != nil {
					t.Fatal(advanceErr)
				}
				changes = append(changes, stepChanges...)
			}
			if state.Values["doom.probability"] != vector.ExpectedValue || state.DecayRemainders["doom.probability"] != vector.ExpectedDecayRemainder ||
				state.InputRemainders["doom.probability:1"] != vector.ExpectedInputRemainder || !reflect.DeepEqual(changes, vector.ExpectedChanges) {
				t.Fatalf("got value=%d decay=%d input=%d changes=%+v", state.Values["doom.probability"], state.DecayRemainders["doom.probability"], state.InputRemainders["doom.probability:1"], changes)
			}
		})
	}
}

func boolSet(values []string) map[string]bool {
	result := make(map[string]bool, len(values))
	for _, value := range values {
		result[value] = true
	}
	return result
}

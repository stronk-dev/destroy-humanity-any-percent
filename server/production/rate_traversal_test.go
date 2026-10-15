package production

import (
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"math"
	"os"
	"reflect"
	"testing"

	"cloud-clicker/server/decimal"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/multiplier"
)

// Controlled ordering fixture, not minted content or a natural player state.
// The literal order is independent of the production traversal, including the
// interleaved all/targeted identities within upgrades.
func rateTraversalFixture(t testing.TB) (*economy.Catalog, []multiplier.Contribution) {
	t.Helper()
	data, err := os.ReadFile("../../balance/testdata/axis-stack/economy-v5-fixture.json")
	if err != nil {
		t.Fatal(err)
	}
	var root map[string]any
	if err := json.Unmarshal(data, &root); err != nil {
		t.Fatal(err)
	}
	slots := []multiplier.Slot{multiplier.SlotUpgrades, multiplier.SlotMilestones,
		multiplier.SlotAxisStack, multiplier.SlotFaction, multiplier.SlotDoctrine,
		multiplier.SlotCommons, multiplier.SlotTrust, multiplier.SlotEventBuffs, multiplier.SlotPrestige}
	factors := []string{"1.00000000001e0", "9.99999999999e0", "1.23456789012e0", "7.65432109876e0",
		"3.45678901234e0", "1.11111111111e0", "8.88888888888e0", "2.34567890123e0", "6.78901234567e0"}
	declarations := root["multiplier_sources"].([]any)
	ordered := []multiplier.Contribution{}
	add := func(id string, slot multiplier.Slot, target, factor string) {
		value, err := decimal.ParseCanonical(factor)
		if err != nil {
			t.Fatal(err)
		}
		found := false
		for _, declaration := range declarations {
			if declaration.(map[string]any)["id"] == id {
				found = true
			}
		}
		if !found {
			declarations = append(declarations, map[string]any{"id": id, "slot": string(slot), "target": target, "provider": "fixture"})
		}
		ordered = append(ordered, multiplier.Contribution{Slot: slot, SourceID: id, Target: target, Factor: value})
	}
	add("a.a", slots[0], "all", "1.00000000001e0")
	add("a0", slots[0], "generator.beige_tower", "9.99999999999e0")
	add("a_a", slots[0], "all", "1.23456789012e0")
	for index, slot := range slots {
		id := fmt.Sprintf("source.slot_%d", index)
		if slot == multiplier.SlotAxisStack || slot == multiplier.SlotCommons {
			id = ""
			for _, value := range declarations {
				declaration := value.(map[string]any)
				if declaration["slot"] == string(slot) && declaration["target"] == "all" {
					id = declaration["id"].(string)
					break
				}
			}
			if id == "" {
				t.Fatal("fixture lacks real owner source", slot)
			}
		}
		add(id, slot, "all", factors[index])
	}
	root["multiplier_sources"] = declarations
	data, err = json.Marshal(root)
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := economy.LoadCatalog(data)
	if err != nil {
		t.Fatal(err)
	}
	return catalog, ordered
}

func TestRateTraversalValidatesContributionsBeforeSkippingInactiveGenerators(t *testing.T) {
	catalog, canonical := rateTraversalFixture(t)
	for _, masked := range []bool{false, true} {
		counts, provisioned, masks := map[string]int64{}, map[string]int64{}, map[string]bool{}
		for _, generator := range catalog.GeneratorClassesForScope(economy.ScopeCompany) {
			counts[generator.ID], provisioned[generator.ID] = 0, 0
			masks[generator.ID] = masked
		}
		if masked {
			counts["generator.beige_tower"] = 1
		}
		for name, edit := range map[string]func([]multiplier.Contribution) []multiplier.Contribution{
			"undeclared": func(v []multiplier.Contribution) []multiplier.Contribution { v[0].SourceID = "undeclared"; return v },
			"wrong slot": func(v []multiplier.Contribution) []multiplier.Contribution {
				v[0].Slot = multiplier.SlotTrust
				return v
			},
			"wrong target": func(v []multiplier.Contribution) []multiplier.Contribution {
				v[0].Target = "generator.missing"
				return v
			},
			"zero factor": func(v []multiplier.Contribution) []multiplier.Contribution { v[0].Factor = decimal.Zero; return v },
			"nan factor":  func(v []multiplier.Contribution) []multiplier.Contribution { v[0].Factor = decimal.NaN; return v },
			"duplicate":   func(v []multiplier.Contribution) []multiplier.Contribution { return append(v, v[0]) },
		} {
			t.Run(fmt.Sprintf("masked=%t/%s", masked, name), func(t *testing.T) {
				policy := &simulationPolicy{generators: masks, observeProduction: true, active: map[string]RoleActivation{}, candidates: map[string]RoleActivation{}}
				for _, contribution := range canonical {
					activation := RoleActivation{GeneratorID: "generator.dot_matrix_queue", Kind: economy.RoleSynergyFeed, TargetID: contribution.SourceID}
					policy.candidates[roleActivationKey(activation)] = activation
				}
				input := edit(append([]multiplier.Contribution(nil), canonical...))
				got, err := ratesWithProvisionedAndPolicy(catalog, counts, provisioned, input, policy)
				if !errors.Is(err, ErrInvalidEngineState) || got != nil || len(policy.active) != 0 {
					t.Fatalf("inactive contribution escaped validation: rates=%+v active=%+v err=%v", got, policy.active, err)
				}
			})
		}
	}
}

func TestRateTraversalPreservesExactOrderTargetsAndActivation(t *testing.T) {
	catalog, canonical := rateTraversalFixture(t)
	discriminatesInputOrder, discriminatesTargetGrouping := false, false
	for _, state := range []struct {
		name                 string
		purchased, generated map[string]int64
		masked               map[string]bool
	}{
		{"both", map[string]int64{"generator.beige_tower": 2, "generator.dot_matrix_queue": 4}, map[string]int64{"generator.beige_tower": 3, "generator.dot_matrix_queue": 1}, nil},
		{"provisioned only", map[string]int64{"generator.beige_tower": 0, "generator.dot_matrix_queue": 0}, map[string]int64{"generator.beige_tower": 3, "generator.dot_matrix_queue": 1}, nil},
		{"zero target", map[string]int64{"generator.beige_tower": 0, "generator.dot_matrix_queue": 4}, map[string]int64{"generator.beige_tower": 0, "generator.dot_matrix_queue": 0}, nil},
		{"masked target", map[string]int64{"generator.beige_tower": 2, "generator.dot_matrix_queue": 4}, map[string]int64{"generator.beige_tower": 3, "generator.dot_matrix_queue": 1}, map[string]bool{"generator.beige_tower": true}},
		{"all zero", map[string]int64{"generator.beige_tower": 0, "generator.dot_matrix_queue": 0}, map[string]int64{"generator.beige_tower": 0, "generator.dot_matrix_queue": 0}, nil},
	} {
		t.Run(state.name, func(t *testing.T) {
			for _, generator := range catalog.GeneratorClassesForScope(economy.ScopeCompany) {
				if _, exists := state.purchased[generator.ID]; !exists {
					state.purchased[generator.ID] = 0
				}
				if _, exists := state.generated[generator.ID]; !exists {
					state.generated[generator.ID] = 0
				}
			}
			wantRates := []decimal.Decimal{}
			wantActive := map[string]RoleActivation{}
			for _, generator := range catalog.GeneratorClassesForScope(economy.ScopeCompany) {
				count := state.purchased[generator.ID] + state.generated[generator.ID]
				if count == 0 || state.masked[generator.ID] {
					continue
				}
				base := generator.Production.BaseRate.Mul(decimal.FromFloat64(float64(count)))
				rate, reversed, grouped := base, base, base
				for _, contribution := range canonical {
					if contribution.Target == "all" || contribution.Target == generator.ID {
						rate = rate.Mul(contribution.Factor)
						activation := RoleActivation{GeneratorID: "generator.dot_matrix_queue", Kind: economy.RoleSynergyFeed, TargetID: contribution.SourceID}
						wantActive[roleActivationKey(activation)] = activation
					}
				}
				for index := len(canonical) - 1; index >= 0; index-- {
					contribution := canonical[index]
					if contribution.Target == "all" || contribution.Target == generator.ID {
						reversed = reversed.Mul(contribution.Factor)
					}
				}
				// A tempting optimization applies all-target factors before specific
				// factors. This is not the existing within-slot order.
				for _, slot := range []multiplier.Slot{multiplier.SlotUpgrades, multiplier.SlotMilestones, multiplier.SlotAxisStack,
					multiplier.SlotFaction, multiplier.SlotDoctrine, multiplier.SlotCommons, multiplier.SlotTrust, multiplier.SlotEventBuffs, multiplier.SlotPrestige} {
					for _, targeted := range []bool{false, true} {
						for _, contribution := range canonical {
							if contribution.Slot == slot && ((!targeted && contribution.Target == "all") || (targeted && contribution.Target == generator.ID)) {
								grouped = grouped.Mul(contribution.Factor)
							}
						}
					}
				}
				discriminatesInputOrder = discriminatesInputOrder || math.Float64bits(rate.Mantissa()) != math.Float64bits(reversed.Mantissa()) || rate.Exponent() != reversed.Exponent()
				discriminatesTargetGrouping = discriminatesTargetGrouping || math.Float64bits(rate.Mantissa()) != math.Float64bits(grouped.Mantissa()) || rate.Exponent() != grouped.Exponent()
				wantRates = append(wantRates, rate)
			}
			for _, shift := range []int{0, 1, len(canonical) - 1} {
				input := append(append([]multiplier.Contribution{}, canonical[shift:]...), canonical[:shift]...)
				policy := &simulationPolicy{generators: state.masked, observeProduction: true, active: map[string]RoleActivation{}, candidates: map[string]RoleActivation{}}
				for _, contribution := range canonical {
					activation := RoleActivation{GeneratorID: "generator.dot_matrix_queue", Kind: economy.RoleSynergyFeed, TargetID: contribution.SourceID}
					policy.candidates[roleActivationKey(activation)] = activation
				}
				before := append([]multiplier.Contribution(nil), input...)
				purchasedBefore, generatedBefore := maps.Clone(state.purchased), maps.Clone(state.generated)
				candidatesBefore := maps.Clone(policy.candidates)
				got, err := ratesWithProvisionedAndPolicy(catalog, state.purchased, state.generated, input, policy)
				wantResources := 0
				if len(wantRates) > 0 {
					wantResources = 1
				}
				if err != nil || len(got["company.cash"]) != len(wantRates) || len(got) != wantResources {
					t.Fatalf("shift=%d rates=%+v err=%v", shift, got, err)
				}
				for index, want := range wantRates {
					value := got["company.cash"][index]
					if math.Float64bits(value.Mantissa()) != math.Float64bits(want.Mantissa()) || value.Exponent() != want.Exponent() {
						t.Fatalf("shift=%d generator=%d raw rate differs: got=%016x/%d want=%016x/%d", shift, index, math.Float64bits(value.Mantissa()), value.Exponent(), math.Float64bits(want.Mantissa()), want.Exponent())
					}
				}
				if !reflect.DeepEqual(input, before) || !reflect.DeepEqual(state.purchased, purchasedBefore) ||
					!reflect.DeepEqual(state.generated, generatedBefore) || !reflect.DeepEqual(policy.candidates, candidatesBefore) ||
					!reflect.DeepEqual(policy.active, wantActive) {
					t.Fatalf("shift=%d changed input or activation: %+v want %+v", shift, policy.active, wantActive)
				}
			}
		})
	}
	if !discriminatesInputOrder || !discriminatesTargetGrouping {
		t.Fatal("ordering fixture cannot distinguish the rejected traversal alternatives")
	}
}

// Manual equivalent-refactor measurement, not a pacing or acceptance gate.
// The same benchmark can run on the active traversal and a compiler overlay;
// no alternative implementation is selected by production or ordinary tests.
func BenchmarkRateTraversal(b *testing.B) {
	catalog, contributions := rateTraversalFixture(b)
	for _, profile := range []struct {
		name   string
		owned  int64
		inputs []multiplier.Contribution
		mask   bool
	}{
		{"no_contributions", 8, nil, false},
		{"all_slots", 8, contributions, false},
		{"inactive", 0, contributions, false},
		{"masked", 8, contributions, true},
	} {
		b.Run(profile.name, func(b *testing.B) {
			counts := map[string]int64{}
			masks := map[string]bool{}
			for _, generator := range catalog.GeneratorClassesForScope(economy.ScopeCompany) {
				counts[generator.ID] = profile.owned
				masks[generator.ID] = profile.mask
			}
			policy := &simulationPolicy{generators: masks}
			b.ReportAllocs()
			b.ResetTimer()
			for range b.N {
				rates, err := ratesWithProvisionedAndPolicy(catalog, counts, nil, profile.inputs, policy)
				if err != nil || profile.owned > 0 && !profile.mask && len(rates) == 0 {
					b.Fatalf("rate traversal failed: rates=%v err=%v", rates, err)
				}
			}
		})
	}
}

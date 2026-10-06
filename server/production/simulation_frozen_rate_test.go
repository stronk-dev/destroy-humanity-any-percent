package production

import (
	"bytes"
	"errors"
	"testing"

	"cloud-clicker/server/decimal"
	"cloud-clicker/server/multiplier"
	"cloud-clicker/server/prestige"
	"cloud-clicker/server/save"
)

func frozenRateState(t *testing.T, bundle CatalogBundle) *save.State {
	t.Helper()
	state, err := prestige.NewRunState(bundle.Economy, &save.State{RunSeq: 2}, &save.State{}, engineCursor)
	if err != nil {
		t.Fatal(err)
	}
	state.GeneratorProvisioned["generator.beige_tower"] = 5
	return state
}

func TestSimulationFrozenResourceRate(t *testing.T) {
	bundle := reputationContentBundle(t)
	row := multiplier.Contribution{SourceID: bundle.ReputationTree.Bonus.SourceID,
		Slot: bundle.ReputationTree.Bonus.Slot, Target: bundle.ReputationTree.Bonus.Target}
	for _, arm := range []string{"none", "unit", "non-unit"} {
		for _, shape := range []string{"ordinary", "masked", "zero-production"} {
			t.Run(arm+"/"+shape, func(t *testing.T) {
				state := frozenRateState(t, bundle)
				var external []multiplier.Contribution
				want := decimal.FromFloat64(5)
				if arm != "none" {
					contribution := row
					contribution.Factor = decimal.One
					if arm == "non-unit" {
						var err error
						contribution.Factor, err = bundle.ReputationTree.BonusFactor(6, 6, 50_000)
						if err != nil || contribution.Factor.String() != "1.003e0" {
							t.Fatalf("fixture lost tree-derived factor: %s %v", contribution.Factor, err)
						}
						want = decimal.FromFloat64(5.015)
					}
					external = []multiplier.Contribution{contribution}
				}
				mask := AblationMask{}
				if shape == "masked" {
					mask.GeneratorIDs = []string{"generator.beige_tower"}
					want = decimal.Zero
				} else if shape == "zero-production" {
					state.GeneratorProvisioned["generator.beige_tower"] = 0
					want = decimal.Zero
				}
				before, err := save.EncodeState(state)
				if err != nil {
					t.Fatal(err)
				}
				got, err := SimulateResourceRateWithContributions(state, bundle.Economy, "company.cash", external, mask)
				if err != nil || !got.Eq(want) {
					t.Fatalf("frozen masked rate=%s want=%s err=%v", got, want, err)
				}
				// Projection has no ablation input. Compare its unmasked cases;
				// the masked arm independently requires exactly zero above.
				if shape != "masked" {
					projection, err := ProjectRates(CatalogBundle{Economy: bundle.Economy}, state, external, 0)
					if err != nil {
						t.Fatal(err)
					}
					found := false
					for _, resource := range projection.Resources {
						if resource.ResourceID == "company.cash" {
							found = true
							if !got.Eq(resource.Rate) {
								t.Fatalf("rate differs from canonical projection: %s vs %s", got, resource.Rate)
							}
						}
					}
					if !found {
						t.Fatal("canonical projection omitted cash")
					}
				}
				legacy, err := SimulateResourceRate(state, bundle.Economy, "company.cash", mask)
				legacyWant := decimal.Zero
				if shape == "ordinary" {
					legacyWant = decimal.FromFloat64(5)
				}
				if err != nil || !legacy.Eq(legacyWant) {
					t.Fatalf("legacy nil-input contract changed: %s want=%s err=%v", legacy, legacyWant, err)
				}
				after, err := save.EncodeState(state)
				if err != nil || !bytes.Equal(before, after) {
					t.Fatalf("rate projection mutated state: %v", err)
				}
			})
		}
	}
}

func TestSimulationFrozenResourceRateRefusesInvalidInputs(t *testing.T) {
	bundle := reputationContentBundle(t)
	row := multiplier.Contribution{SourceID: bundle.ReputationTree.Bonus.SourceID,
		Slot: bundle.ReputationTree.Bonus.Slot, Target: bundle.ReputationTree.Bonus.Target, Factor: decimal.One}
	for _, fault := range []string{"source", "slot", "target", "zero", "negative", "nan", "duplicate", "mask", "resource", "nil-state", "nil-catalog"} {
		t.Run(fault, func(t *testing.T) {
			state := frozenRateState(t, bundle)
			before, err := save.EncodeState(state)
			if err != nil {
				t.Fatal(err)
			}
			external := []multiplier.Contribution{row}
			mask, resource, catalog := AblationMask{}, "company.cash", bundle.Economy
			input := state
			switch fault {
			case "source":
				external[0].SourceID = "undeclared"
			case "slot":
				external[0].Slot = "global"
			case "target":
				external[0].Target = "generator.beige_tower"
			case "zero":
				external[0].Factor = decimal.Zero
			case "negative":
				external[0].Factor = decimal.FromFloat64(-1)
			case "nan":
				external[0].Factor = decimal.NaN
			case "duplicate":
				external = append(external, row)
			case "mask":
				mask.GeneratorIDs = []string{"undeclared"}
			case "resource":
				resource = "founder.clout"
			case "nil-state":
				input = nil
			case "nil-catalog":
				catalog = nil
			}
			got, err := SimulateResourceRateWithContributions(input, catalog, resource, external, mask)
			if !errors.Is(err, ErrInvalidEngineState) || !got.IsNaN() {
				t.Fatalf("invalid %s admitted: rate=%s err=%v", fault, got, err)
			}
			after, encodeErr := save.EncodeState(state)
			if encodeErr != nil || !bytes.Equal(before, after) {
				t.Fatalf("invalid projection mutated state: %v", encodeErr)
			}
		})
	}
}

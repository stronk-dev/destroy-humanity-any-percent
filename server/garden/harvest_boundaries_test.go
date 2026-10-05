package garden

import (
	"encoding/json"
	"math"
	"math/big"
	"testing"
)

func TestGardenHarvestBoundaries(t *testing.T) {
	var corpus struct {
		Version int `json:"version"`
		Cases   []struct {
			Name               string          `json:"name"`
			CatalogOps         []fixtureOp     `json:"catalog_ops"`
			Initial            string          `json:"initial"`
			IntentID           string          `json:"intent_id"`
			Targets            []HarvestTarget `json:"targets"`
			ExpectedResult     string          `json:"expected_result"`
			ExpectedState      string          `json:"expected_state"`
			ScalarHarvestUnits *int64          `json:"scalar_harvest_units"`
			ScalarEffectPPM    *int64          `json:"scalar_effect_ppm"`
			ExpectedProduct    string          `json:"expected_product"`
			ExpectedUnits      *int64          `json:"expected_units"`
		} `json:"cases"`
	}
	if err := json.Unmarshal(readRepo(t, "testdata/garden/harvest-boundaries-v1.json"), &corpus); err != nil || corpus.Version != 1 || len(corpus.Cases) != 18 {
		t.Fatalf("harvest boundary corpus incomplete/invalid: %v", err)
	}
	floatingMismatch := 0
	for _, test := range corpus.Cases {
		t.Run(test.Name, func(t *testing.T) {
			catalog := scenarioCatalog(t, test.CatalogOps)
			state, err := DecodeState([]byte(test.Initial))
			if err != nil || ValidateAgainst(catalog, state) != nil {
				t.Fatalf("boundary input is outside the real catalog/state grammar: %v", err)
			}
			if test.ScalarHarvestUnits != nil {
				if test.ScalarEffectPPM == nil || test.ExpectedUnits == nil {
					t.Fatal("scalar reference is incomplete")
				}
				if len(state.Plots) != 1 || state.Plots[0].SpeciesID != "strain_a" || catalog.Species[0].SpeciesID != "strain_a" ||
					catalog.Species[0].HarvestUnits != *test.ScalarHarvestUnits || state.Plots[0].MaturedEffectPPM == nil || *state.Plots[0].MaturedEffectPPM != *test.ScalarEffectPPM {
					t.Fatal("reference operands are not bound to the actual loaded harvest inputs")
				}
				// Production uses bounded int64; this independent arbitrary-precision
				// reference validates the literal mathematical product and quotient.
				product := new(big.Int).Mul(big.NewInt(*test.ScalarHarvestUnits), big.NewInt(*test.ScalarEffectPPM))
				if product.String() != test.ExpectedProduct || new(big.Int).Quo(product, big.NewInt(PPM)).Int64() != *test.ExpectedUnits {
					t.Fatal("literal quotient/product disagrees with math/big")
				}
				wrong := int64(math.Floor(float64(*test.ScalarHarvestUnits) * float64(*test.ScalarEffectPPM) / float64(PPM)))
				if wrong != *test.ExpectedUnits {
					floatingMismatch++
				}
			}
			harvest, err := catalog.Harvest(state, test.IntentID, test.Targets)
			if err != nil {
				t.Fatal(err)
			}
			result, err := json.Marshal(harvest)
			if err != nil || string(result) != test.ExpectedResult {
				t.Fatalf("harvest bytes differ: got %s want %s error=%v", result, test.ExpectedResult, err)
			}
			if err := ValidateAgainst(catalog, state); err != nil {
				t.Fatalf("harvest produced invalid pinned state: %v", err)
			}
			post, err := EncodeState(state)
			if err != nil || string(post) != test.ExpectedState {
				t.Fatalf("post-state differs: got %s want %s error=%v", post, test.ExpectedState, err)
			}
		})
	}
	if floatingMismatch != 1 {
		t.Fatalf("declared floating-product negative population changed: %d mismatches, expected 1", floatingMismatch)
	}
}

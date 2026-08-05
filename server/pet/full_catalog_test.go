package pet

import (
	"encoding/json"
	"testing"
)

const fullCatalogFixture = `{"schema_version":1,"stat_policy":{"grid_ms":60000,"stats":[{"stat_id":"hunger","initial_ppm":800000,"floor_ppm":100000,"decay_ppm_per_grid":1000},{"stat_id":"energy","initial_ppm":800000,"floor_ppm":100000,"decay_ppm_per_grid":1000},{"stat_id":"cleanliness","initial_ppm":800000,"floor_ppm":100000,"decay_ppm_per_grid":1000},{"stat_id":"affection","initial_ppm":800000,"floor_ppm":100000,"decay_ppm_per_grid":1000}],"diminishing_threshold_ppm":700000,"diminishing_factor_ppm":500000},"actions":[{"action_id":"care.feed","stat_id":"hunger","delta_ppm":100000,"cooldown_attended_ms":60000,"min_eligible_ppm":0}],"trust_policy":{"initial_ppm":500000,"neutral_ppm":500000,"floor_ppm":100000,"cap_ppm":1000000,"gain_ppm_per_effective_action":1000,"decay_ppm_per_grid":100},"mood_policy":[{"mood_member":"withdrawn","floor_ppm":0},{"mood_member":"restless","floor_ppm":250000},{"mood_member":"neutral","floor_ppm":500000},{"mood_member":"engaged","floor_ppm":750000}],"behavior_policy":[{"from_state":"idle","event":"care_applied","to_state":"care_response","duration_grid_ticks":1}]}`

func TestLoadFullCatalogClosesEveryPolicyFamily(t *testing.T) {
	catalog, err := LoadCatalog([]byte(fullCatalogFixture))
	if err != nil {
		t.Fatal(err)
	}
	declarations := catalog.StateDeclarations()
	if len(declarations.ActionIDs) != 1 || declarations.ActionIDs[0] != "care.feed" || len(declarations.BehaviorIDs) != len(behaviorStates) {
		t.Fatalf("declarations=%+v", declarations)
	}
}

func TestLoadFullCatalogRejectsMissingKeysAtEveryLayer(t *testing.T) {
	var baseline map[string]any
	if err := json.Unmarshal([]byte(fullCatalogFixture), &baseline); err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name   string
		mutate func(map[string]any)
	}{
		{"top-level trust policy", func(value map[string]any) { delete(value, "trust_policy") }},
		{"stat-policy diminishing factor", func(value map[string]any) {
			delete(value["stat_policy"].(map[string]any), "diminishing_factor_ppm")
		}},
		{"stat-row decay", func(value map[string]any) {
			delete(value["stat_policy"].(map[string]any)["stats"].([]any)[0].(map[string]any), "decay_ppm_per_grid")
		}},
		{"action minimum", func(value map[string]any) {
			delete(value["actions"].([]any)[0].(map[string]any), "min_eligible_ppm")
		}},
		{"trust decay", func(value map[string]any) {
			delete(value["trust_policy"].(map[string]any), "decay_ppm_per_grid")
		}},
		{"mood floor", func(value map[string]any) {
			delete(value["mood_policy"].([]any)[0].(map[string]any), "floor_ppm")
		}},
		{"behavior duration", func(value map[string]any) {
			delete(value["behavior_policy"].([]any)[0].(map[string]any), "duration_grid_ticks")
		}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			encoded, err := json.Marshal(baseline)
			if err != nil {
				t.Fatal(err)
			}
			var value map[string]any
			if err := json.Unmarshal(encoded, &value); err != nil {
				t.Fatal(err)
			}
			testCase.mutate(value)
			data, err := json.Marshal(value)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := LoadCatalog(data); err == nil {
				t.Fatalf("accepted missing key: %s", data)
			}
		})
	}
}

func TestLoadFullCatalogEnforcesNoDeathAndTotalMoodProjection(t *testing.T) {
	var baseline map[string]any
	if err := json.Unmarshal([]byte(fullCatalogFixture), &baseline); err != nil {
		t.Fatal(err)
	}
	for _, testCase := range []struct {
		name   string
		mutate func(map[string]any)
	}{
		{"recovery unavailable at floor", func(value map[string]any) {
			value["actions"].([]any)[0].(map[string]any)["min_eligible_ppm"] = float64(100001)
		}},
		{"mood projection has no zero floor", func(value map[string]any) {
			value["mood_policy"].([]any)[0].(map[string]any)["floor_ppm"] = float64(1)
		}},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			encoded, _ := json.Marshal(baseline)
			var value map[string]any
			_ = json.Unmarshal(encoded, &value)
			testCase.mutate(value)
			candidate, _ := json.Marshal(value)
			if _, err := LoadCatalog(candidate); err == nil {
				t.Fatal("invalid policy accepted")
			}
		})
	}
}

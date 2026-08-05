package pet

import "testing"

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
	invalid := []byte(fullCatalogFixture)
	for index := 0; index < len(invalid)-1; index++ {
		if string(invalid[index:index+2]) == `1}` {
			invalid[index] = '2'
			break
		}
	}
	if _, err := LoadCatalog(append(invalid, []byte(`{"extra":1}`)...)); err == nil {
		t.Fatal("accepted malformed full catalog")
	}
}

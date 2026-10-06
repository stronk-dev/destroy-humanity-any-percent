package harness

import (
	"encoding/json"
	"reflect"
	"testing"
)

// JSON-side shape compiles before the raw-observation field exists. Synthetic
// source-admitted observations are not actual earned careers or a full cohort.
func TestReputationRelevanceRawObservationRetention(t *testing.T) {
	suite, spec, experiment, config := reputationCareerAdmissionInputs(t)
	var results []reputationRelevanceOutcome
	for index, exclude := range []string{"", "reputation.starter.cash_small"} {
		masked := config
		masked.Exclude = exclude
		encoded, err := json.Marshal(expectedReputationCareerSource(t, suite, spec, 0, experiment, masked))
		if err != nil {
			t.Fatal(err)
		}
		var source ReputationCareerMeasurementSource
		if err := json.Unmarshal(encoded, &source); err != nil {
			t.Fatal(err)
		}
		gate := int64(10 + index)
		results = append(results, reputationRelevanceOutcome{gate: &gate,
			purchased: []string{"reputation.unlock.p05"}, source: source})
	}
	report := newReputationRelevanceReport(results)
	encoded, err := json.Marshal(report)
	if err != nil {
		t.Fatal(err)
	}
	var wire struct {
		Arms []struct {
			Source    ReputationCareerMeasurementSource `json:"measurement_source"`
			Gate      *int64                            `json:"run_three_gate_ms"`
			Purchased []string                          `json:"purchased_node_ids"`
		} `json:"arm_observations"`
	}
	if err := json.Unmarshal(encoded, &wire); err != nil {
		t.Fatal(err)
	}
	if len(wire.Arms) != len(results) {
		t.Fatalf("H5 discarded raw arm observations: got %d want %d", len(wire.Arms), len(results))
	}
	for index, result := range results {
		if wire.Arms[index].Source != result.source || !reflect.DeepEqual(wire.Arms[index].Gate, result.gate) ||
			!reflect.DeepEqual(wire.Arms[index].Purchased, result.purchased) {
			t.Fatal("H5 raw observation differs from the admitted arm")
		}
	}
}

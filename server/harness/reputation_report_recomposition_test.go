package harness

import (
	"bytes"
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
	*results[0].gate = 999
	results[0].purchased[0] = "mutated-input"
	after, err := json.Marshal(report)
	if err != nil || !bytes.Equal(encoded, after) {
		t.Fatal("editing mutable inputs revised retained H5 evidence")
	}
}

func TestReputationCareerReportRecomposition(t *testing.T) {
	clock := func(v int64) *int64 { return &v }
	rows := []reputationCareerSeed{
		{PolicyID: "p", AppliedStarterIDs: []string{"synthetic"}, TreatedGateMS: clock(5), ControlGateMS: clock(10), SavedMS: clock(5)},
		{PolicyID: "p", Seed: 1, AppliedStarterIDs: []string{"synthetic"}, TreatedGateMS: clock(10), ControlGateMS: clock(10), SavedMS: clock(0)},
		{PolicyID: "p", Seed: 2, AppliedStarterIDs: []string{"synthetic"}, Excluded: reputationBothArmsBeyondHorizon},
	}
	report, err := newReputationCareerReport(rows)
	if err != nil || report.GatePassed || report.GatedSeeds != 2 || report.ExcludedSeeds != 1 || len(report.Violations) != 1 ||
		!reflect.DeepEqual(report.SavedMS, map[string][3]int64{"p": {5, 5, 5}}) ||
		!reflect.DeepEqual(report.FiniteSavedMS, map[string][3]int64{"p": {0, 5, 5}}) {
		t.Fatalf("shared H4 composer lost gate or conditional/all-finite statistics: %+v / %v", report, err)
	}
}

func TestReputationRelevanceReportRecomposition(t *testing.T) {
	suite, spec, experiment, config := reputationCareerAdmissionInputs(t)
	nodes := config.Bundle.ReputationTree.Nodes()
	var results []reputationRelevanceOutcome
	for seed := uint64(0); seed < 2; seed++ {
		masks := []string{""}
		for _, node := range nodes {
			masks = append(masks, node.NodeID)
		}
		for _, mask := range masks {
			masked := config
			masked.Exclude = mask
			data, err := json.Marshal(expectedReputationCareerSource(t, suite, spec, seed, experiment, masked))
			if err != nil {
				t.Fatal(err)
			}
			var source ReputationCareerMeasurementSource
			if err := json.Unmarshal(data, &source); err != nil {
				t.Fatal(err)
			}
			purchased := []string{"reputation.unlock.p05", "reputation.starter.cash_small"}
			if mask == "reputation.unlock.p05" {
				purchased = nil
			} else if mask == "reputation.starter.cash_small" {
				purchased = purchased[:1]
			}
			var gate *int64
			if seed == 0 {
				value := int64(10)
				if mask == "reputation.starter.cash_small" {
					value = 20
				}
				gate = &value
			}
			results = append(results, reputationRelevanceOutcome{source: source, gate: gate, purchased: purchased})
		}
	}
	// Synthetic two-seed cohort, not the full ratified 97 careers.
	report, err := composeReputationRelevanceReport(nodes, results)
	if err != nil || len(report.Arms) != 20 || len(report.Sources) != 20 || len(report.Nodes) != len(nodes) {
		t.Fatalf("healthy complete synthetic arm groups failed: %v", err)
	}
	if err := validateReputationRelevanceRecomposition(nodes, report); err != nil {
		t.Fatalf("retained raw H5 report did not recompose: %v", err)
	}
	for _, row := range report.Nodes {
		if row.NodeID == "reputation.starter.cash_small" {
			population := row.Population[spec.PolicyID]
			if row.PurchasedRuns[spec.PolicyID] != 2 || population.BoughtPairs.FinitePairs != 1 ||
				population.BoughtPairs.BothUnreached != 1 || row.DeltaMSP50[spec.PolicyID] != 10 || !row.Relevant {
				t.Fatalf("shared H5 composer changed conditional statistic/classification: %+v", row)
			}
		}
	}
	cloneResults := func() []reputationRelevanceOutcome {
		copy := append([]reputationRelevanceOutcome(nil), results...)
		for index := range copy {
			copy[index].purchased = append([]string(nil), results[index].purchased...)
			if results[index].gate != nil {
				value := *results[index].gate
				copy[index].gate = &value
			}
		}
		return copy
	}
	zeroEffect := cloneResults()
	for index := range zeroEffect {
		if zeroEffect[index].source.ExcludedNodeID == "reputation.starter.cash_small" && zeroEffect[index].gate != nil {
			*zeroEffect[index].gate = 10
		}
	}
	invalidReport, err := composeReputationRelevanceReport(nodes, zeroEffect)
	if err != nil {
		t.Fatal("zero-effect control failed before its H5 criterion")
	}
	if err := validateReputationRelevanceRecomposition(nodes, invalidReport); err == nil {
		t.Fatal("internally consistent bought starter with no effect/exclusion passed H5 report admission")
	}
	for _, name := range []string{"missing-mask", "duplicate-mask", "duplicate-baseline", "missing-baseline", "unknown-mask", "source-horizon", "negative-gate", "unknown-purchase", "duplicate-purchase", "excluded-purchase"} {
		t.Run(name, func(t *testing.T) {
			copy := cloneResults()
			switch name {
			case "missing-mask":
				copy = copy[:len(copy)-1]
			case "duplicate-mask":
				copy = append(copy, copy[1])
			case "duplicate-baseline":
				copy = append(copy, copy[0])
			case "missing-baseline":
				copy = copy[1:]
			case "unknown-mask":
				copy[1].source.ExcludedNodeID = "missing"
			case "source-horizon":
				copy[1].source.HorizonMS++
			case "negative-gate":
				*copy[1].gate = -1
			case "unknown-purchase":
				copy[0].purchased = append(copy[0].purchased, "missing")
			case "duplicate-purchase":
				copy[0].purchased = append(copy[0].purchased, copy[0].purchased[0])
			case "excluded-purchase":
				copy[1].purchased = []string{copy[1].source.ExcludedNodeID}
			}
			if _, err := composeReputationRelevanceReport(nodes, copy); err == nil {
				t.Fatal("shared H5 composer admitted a malformed raw population")
			}
		})
	}
	for _, name := range []string{"median", "population", "classification", "sources", "missing-arms", "schema", "epsilon", "threshold"} {
		t.Run("retained-"+name, func(t *testing.T) {
			data, err := json.Marshal(report)
			if err != nil {
				t.Fatal(err)
			}
			var copy reputationRelevanceReport
			if err := json.Unmarshal(data, &copy); err != nil {
				t.Fatal(err)
			}
			switch name {
			case "median":
				copy.Nodes[0].DeltaMSP50[spec.PolicyID]++
			case "population":
				copy.Nodes[0].Population = nil
			case "classification":
				copy.Nodes[0].Relevant = !copy.Nodes[0].Relevant
			case "sources":
				copy.Sources[0].HorizonMS++
			case "missing-arms":
				copy.Arms = copy.Arms[1:]
			case "schema":
				copy.SchemaVersion++
			case "epsilon":
				copy.EpsilonMS++
			case "threshold":
				copy.Threshold = "1e12"
			}
			if err := validateReputationRelevanceRecomposition(nodes, copy); err == nil {
				t.Fatal("retained H5 aggregate/source metadata escaped raw recomposition")
			}
		})
	}
}

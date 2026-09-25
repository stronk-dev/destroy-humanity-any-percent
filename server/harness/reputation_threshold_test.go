package harness

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"cloud-clicker/server/decimal"
	prestigecore "cloud-clicker/server/prestige"
)

var repositoryRootForReputation = filepath.Clean(filepath.Join("..", ".."))

const reputationMeasurementPath = "planning/reputation-tree-v1/threshold-measurement.v1.json"

func reputationMeasurementInputs(t *testing.T) (FirstHourExperimentReport, *prestigecore.Policy) {
	t.Helper()
	var report FirstHourExperimentReport
	data, err := os.ReadFile(filepath.Join(repositoryRootForReputation, "planning/reputation-tree-v1/first-hour-reputation.v1.json"))
	if err != nil || json.Unmarshal(data, &report) != nil {
		t.Fatalf("first-hour Reputation report: %v", err)
	}
	policyBytes, err := os.ReadFile(filepath.Join(repositoryRootForReputation, "balance/prestige/phase0.json"))
	if err != nil {
		t.Fatal(err)
	}
	policy, err := prestigecore.LoadPolicy(policyBytes)
	if err != nil {
		t.Fatal(err)
	}
	return report, policy
}

func reputationThresholdGrid() []string {
	grid := []string{}
	for _, exponent := range []string{"3", "4", "5", "6", "7", "8", "9", "10", "11", "12"} {
		for _, mantissa := range []string{"1", "2", "5"} {
			grid = append(grid, mantissa+"e"+exponent)
		}
	}
	return grid
}

// TestReputationThresholdMeasurement is H2 and the OD-2 measurement report.
// REPUTATION_UPDATE_MEASUREMENT=1 regenerates the pinned report.
func TestReputationThresholdMeasurement(t *testing.T) {
	report, policy := reputationMeasurementInputs(t)
	measured, err := MeasureReputationThresholds(report, policy, reputationThresholdGrid(), ReputationThresholdEnvelope{Minimum: 3, Maximum: 10},
		[]string{"casual.t0_t1", "chaos.t0_t1"})
	if err != nil {
		t.Fatal(err)
	}
	// H2's demonstrated failing case: the live threshold pays nothing.
	for _, row := range measured.Rows {
		if row.Threshold == policy.Threshold {
			if row.Satisfies {
				t.Fatalf("live threshold %s satisfies the envelope: %+v", row.Threshold, row)
			}
			for _, persona := range row.Personas {
				if persona.Max != 0 {
					t.Fatalf("live threshold pays Reputation: %+v", persona)
				}
			}
		}
	}
	encoded, err := json.MarshalIndent(measured, "", " ")
	if err != nil {
		t.Fatal(err)
	}
	encoded = append(encoded, '\n')
	path := filepath.Join(repositoryRootForReputation, reputationMeasurementPath)
	if os.Getenv("REPUTATION_UPDATE_MEASUREMENT") == "1" {
		if err := os.WriteFile(path, encoded, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	pinned, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(pinned, encoded) {
		t.Fatalf("threshold measurement drifted from its source report (regenerate with REPUTATION_UPDATE_MEASUREMENT=1): %v", err)
	}
}

func TestReputationThresholdMeasurementFailsLoud(t *testing.T) {
	report, policy := reputationMeasurementInputs(t)
	unrecorded := report
	unrecorded.Runs = append([]FirstHourRunResult{}, report.Runs...)
	unrecorded.Runs[0].ReputationExits = nil
	if _, err := MeasureReputationThresholds(unrecorded, policy, []string{"1e6"}, ReputationThresholdEnvelope{Minimum: 3, Maximum: 10}, []string{"chaos.t0_t1"}); !errors.Is(err, ErrReputationMeasurement) {
		t.Fatalf("a run without H1 samples was silently measured: %v", err)
	}
	failed := report
	failed.Runs = append([]FirstHourRunResult{}, report.Runs...)
	failed.Runs[0].Outcome = "failed"
	if _, err := MeasureReputationThresholds(failed, policy, []string{"1e6"}, ReputationThresholdEnvelope{Minimum: 3, Maximum: 10}, []string{"chaos.t0_t1"}); !errors.Is(err, ErrReputationMeasurement) {
		t.Fatalf("a failed run was silently measured: %v", err)
	}
	if _, err := MeasureReputationThresholds(report, policy, []string{"1e6"}, ReputationThresholdEnvelope{Minimum: 3, Maximum: 10}, []string{"persona.missing"}); !errors.Is(err, ErrReputationMeasurement) {
		t.Fatalf("a missing gated persona was not an error: %v", err)
	}
	// Hand check: one Chaos run under 1e6.
	paid, err := PaidReputationAtFirstElectiveExit(report.Runs[len(report.Runs)-1], policy, decimal.FromFloat64(1e6))
	if err != nil || paid < 0 {
		t.Fatalf("paid=%d err=%v", paid, err)
	}
}

// TestFirstHourRecordsReputationAtEachExit is H1 on a live run: both Exits
// carry their Reputation facts, lifetime value really accrues (the recorded
// report is not re-derived from a stale harness), and under the live 1e12
// threshold the payout is zero.
func TestFirstHourRecordsReputationAtEachExit(t *testing.T) {
	suite, err := LoadFirstHourSuite(repositoryRootForReputation, "balance/testdata/t0-t1/harness-scenario-v1.json", "balance/testdata/t0-t1/first-hour-policy-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	experiment := FirstHourExperiment{AcquihirePurchasedMinimum: 200, BurnoutPriceFactor: "2e0", RouteKnowledgeBonus: 50, SeedCapital: "1e4", GeneratedBeigeTowers: 10}
	var chaos RunSpec
	for _, spec := range suite.Scenario.Runs {
		if spec.PolicyID == "chaos.t0_t1" {
			chaos = spec
		}
	}
	result := suite.RunExperiment(chaos, 0, experiment)
	if result.Outcome != "completed" || len(result.ReputationExits) != 2 {
		t.Fatalf("outcome=%s exits=%+v failures=%v", result.Outcome, result.ReputationExits, result.InvariantFailures)
	}
	for _, sample := range result.ReputationExits {
		lifetime, err := decimal.ParseCanonical(sample.LifetimeValue)
		if err != nil || !lifetime.Gt(decimal.FromFloat64(1e5)) || sample.ReputationDelta != 0 {
			t.Fatalf("Exit %s lifetime=%s delta=%d err=%v", sample.ExitType, sample.LifetimeValue, sample.ReputationDelta, err)
		}
	}
	if result.ReputationExits[0].ExitType != "scripted_first" || result.ReputationExits[1].ExitType != "collapse" {
		t.Fatalf("Exit order=%+v", result.ReputationExits)
	}
}

// TestReputationRecordingIsFirstHourNeutral is H3: adding H1 recording and
// the lifetime hook left every epoch-8 milestone, ending, and aggregate
// byte-identical; only reputation_exits differs.
func TestReputationRecordingIsFirstHourNeutral(t *testing.T) {
	load := func(path string) FirstHourExperimentReport {
		var report FirstHourExperimentReport
		data, err := os.ReadFile(filepath.Join(repositoryRootForReputation, path))
		if err != nil || json.Unmarshal(data, &report) != nil {
			t.Fatalf("%s: %v", path, err)
		}
		return report
	}
	before := load("planning/prestige-and-exits/first-hour-epoch8-report.v1.json")
	after := load("planning/reputation-tree-v1/first-hour-reputation.v1.json")
	if before.ConstantsHash != after.ConstantsHash || before.ScenarioHash != after.ScenarioHash || len(before.Runs) != len(after.Runs) {
		t.Fatal("reports are not the same measurement")
	}
	encode := func(value any) string { data, _ := json.Marshal(value); return string(data) }
	if encode(before.Aggregate) != encode(after.Aggregate) {
		t.Fatal("aggregate moved")
	}
	for index := range before.Runs {
		left, right := before.Runs[index], after.Runs[index]
		if encode(left.Milestones) != encode(right.Milestones) || encode(left.Ending) != encode(right.Ending) || encode(left.Key) != encode(right.Key) {
			t.Fatalf("run %d moved", index)
		}
	}
}

// requireReputationExhaustive gates the multi-minute Reputation career and
// relevance recomputations out of the 5-minute push harness. They run in the
// maintenance lane via `make reputation-harness-check`; the skip is explicit
// and named so a fast run never looks like it exercised them.
func requireReputationExhaustive(t *testing.T) {
	t.Helper()
	if os.Getenv("CLOUD_CLICKER_REPUTATION_EXHAUSTIVE") != "1" {
		t.Skip("exhaustive Reputation harness evidence runs in `make reputation-harness-check` (CLOUD_CLICKER_REPUTATION_EXHAUSTIVE=1)")
	}
}

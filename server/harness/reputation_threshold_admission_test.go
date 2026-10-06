package harness

import (
	"encoding/json"
	"fmt"
	"testing"

	"cloud-clicker/server/decimal"
	prestigecore "cloud-clicker/server/prestige"
)

// Complete-study admission belongs to this pinned caller, not the generic
// threshold calculator. This is still historical-source recalculation.
func measureReputationThresholdStudy(suite *FirstHourSuite, experiment FirstHourExperiment, report FirstHourExperimentReport,
	policy *prestigecore.Policy, thresholds []string) (ReputationThresholdReport, error) {
	if err := validateReputationFirstHourPopulation(suite, experiment, report); err != nil {
		return ReputationThresholdReport{}, fmt.Errorf("%w: threshold study source: %v", ErrReputationMeasurement, err)
	}
	if report.Aggregate.SchemaVersion != 1 || report.Aggregate.ScenarioID != suite.Scenario.ID ||
		report.Aggregate.ScenarioHash != suite.ScenarioHash || report.Aggregate.ConstantsHash != suite.ConstantsHash {
		return ReputationThresholdReport{}, fmt.Errorf("%w: threshold aggregate source mismatch", ErrReputationMeasurement)
	}
	return MeasureReputationThresholds(report, policy, thresholds, ReputationThresholdEnvelope{Minimum: 3, Maximum: 10},
		[]string{"casual.t0_t1", "chaos.t0_t1"})
}

func cloneReputationThresholdSource(t *testing.T, report FirstHourExperimentReport) FirstHourExperimentReport {
	t.Helper()
	data, err := json.Marshal(report)
	if err != nil {
		t.Fatal(err)
	}
	var copy FirstHourExperimentReport
	if err := json.Unmarshal(data, &copy); err != nil {
		t.Fatal(err)
	}
	return copy
}

// Synthetic samples exercise admission/arithmetic, not earned player careers.
func TestReputationThresholdExitAdmissionDiagnostic(t *testing.T) {
	_, policy := reputationMeasurementInputs(t)
	if policy.ExitModifiersPPM["scripted_first"] != 1_000_000 || policy.ExitModifiersPPM["collapse"] != 750_000 {
		t.Fatal("hand-calculated control no longer uses the declared policy")
	}
	base := FirstHourRunResult{ReputationExits: []FirstHourReputationSample{
		{RunSeq: 1, ExitType: "scripted_first", LifetimeValue: "1e6"},
		{RunSeq: 2, ExitType: "collapse", LifetimeValue: "1e8"},
	}}
	t.Run("known-paid-two", func(t *testing.T) {
		paid, err := PaidReputationAtFirstElectiveExit(base, policy, decimal.FromFloat64(1e6))
		// floor cube-root levels 1 and 4; floor((4-1)*0.75) = 2.
		if err != nil || paid != 2 {
			t.Fatalf("known payout=%d, want2: %v", paid, err)
		}
	})
	t.Run("legal-zero", func(t *testing.T) {
		copy := cloneReputationThresholdSource(t, FirstHourExperimentReport{Runs: []FirstHourRunResult{base}}).Runs[0]
		for index := range copy.ReputationExits {
			copy.ReputationExits[index].LifetimeValue = "0"
		}
		if paid, err := PaidReputationAtFirstElectiveExit(copy, policy, decimal.FromFloat64(1e6)); err != nil || paid != 0 {
			t.Fatalf("legal zero payout=%d: %v", paid, err)
		}
	})
	mutations := map[string]func(*FirstHourRunResult){
		"reversed": func(r *FirstHourRunResult) {
			r.ReputationExits[0], r.ReputationExits[1] = r.ReputationExits[1], r.ReputationExits[0]
		},
		"scripted-zero-sequence":  func(r *FirstHourRunResult) { r.ReputationExits[0].RunSeq = 0 },
		"scripted-wrong-sequence": func(r *FirstHourRunResult) { r.ReputationExits[0].RunSeq = 2 },
		"elective-wrong-sequence": func(r *FirstHourRunResult) { r.ReputationExits[1].RunSeq = 1 },
		"later-elective":          func(r *FirstHourRunResult) { r.ReputationExits[1].RunSeq = 3 },
		"duplicate-scripted":      func(r *FirstHourRunResult) { r.ReputationExits = append(r.ReputationExits, r.ReputationExits[0]) },
		"duplicate-elective":      func(r *FirstHourRunResult) { r.ReputationExits = append(r.ReputationExits, r.ReputationExits[1]) },
		"unknown-extra": func(r *FirstHourRunResult) {
			r.ReputationExits = append(r.ReputationExits, FirstHourReputationSample{RunSeq: 3, ExitType: "invented", LifetimeValue: "1e8"})
		},
		"missing-scripted":           func(r *FirstHourRunResult) { r.ReputationExits = r.ReputationExits[1:] },
		"missing-elective":           func(r *FirstHourRunResult) { r.ReputationExits = r.ReputationExits[:1] },
		"unknown-scripted-kind":      func(r *FirstHourRunResult) { r.ReputationExits[0].ExitType = "invented" },
		"unknown-elective-kind":      func(r *FirstHourRunResult) { r.ReputationExits[1].ExitType = "invented" },
		"negative-scripted-lifetime": func(r *FirstHourRunResult) { r.ReputationExits[0].LifetimeValue = "-1e6" },
		"negative-elective-lifetime": func(r *FirstHourRunResult) { r.ReputationExits[1].LifetimeValue = "-1e8" },
		"nan-scripted-lifetime":      func(r *FirstHourRunResult) { r.ReputationExits[0].LifetimeValue = "NaN" },
		"nan-elective-lifetime":      func(r *FirstHourRunResult) { r.ReputationExits[1].LifetimeValue = "NaN" },
	}
	for name, mutate := range mutations {
		t.Run(name, func(t *testing.T) {
			copy := cloneReputationThresholdSource(t, FirstHourExperimentReport{Runs: []FirstHourRunResult{base}}).Runs[0]
			mutate(&copy)
			paid, err := PaidReputationAtFirstElectiveExit(copy, policy, decimal.FromFloat64(1e6))
			t.Logf("observed payout=%d error=%v", paid, err)
			if err == nil {
				t.Fatal("malformed first-hour Exit sequence was silently measured")
			}
		})
	}
}

// Full-study admission belongs at the pinned caller, not inside the generic
// calculator. These refusals use the same path as its retained-byte check.
func TestReputationThresholdStudySourceDiagnostic(t *testing.T) {
	report, policy := reputationMeasurementInputs(t)
	suite, _, experiment, _ := reputationCareerAdmissionInputs(t)
	if err := validateReputationFirstHourPopulation(suite, experiment, report); err != nil {
		t.Fatalf("retained source is not the declared historical population: %v", err)
	}
	mutations := map[string]func(*FirstHourExperimentReport){
		"report-schema":           func(r *FirstHourExperimentReport) { r.SchemaVersion++ },
		"report-scenario-id":      func(r *FirstHourExperimentReport) { r.ScenarioID = "invented" },
		"report-scenario-hash":    func(r *FirstHourExperimentReport) { r.ScenarioHash = "invented" },
		"report-policy-hash":      func(r *FirstHourExperimentReport) { r.PolicyHash = "invented" },
		"report-constants-hash":   func(r *FirstHourExperimentReport) { r.ConstantsHash = "invented" },
		"experiment":              func(r *FirstHourExperimentReport) { r.Experiment.GeneratedBeigeTowers++ },
		"run-source":              func(r *FirstHourExperimentReport) { r.Runs[0].Key.ConstantsHash = "invented" },
		"run-policy":              func(r *FirstHourExperimentReport) { r.Runs[0].PolicyHash = "invented" },
		"aggregate-count":         func(r *FirstHourExperimentReport) { r.Aggregate.RunCount-- },
		"aggregate-source":        func(r *FirstHourExperimentReport) { r.Aggregate.ConstantsHash = "invented" },
		"aggregate-schema":        func(r *FirstHourExperimentReport) { r.Aggregate.SchemaVersion++ },
		"aggregate-scenario-id":   func(r *FirstHourExperimentReport) { r.Aggregate.ScenarioID = "invented" },
		"aggregate-scenario-hash": func(r *FirstHourExperimentReport) { r.Aggregate.ScenarioHash = "invented" },
		"missing-run":             func(r *FirstHourExperimentReport) { r.Runs = r.Runs[1:] },
		"extra-run":               func(r *FirstHourExperimentReport) { r.Runs = append(r.Runs, r.Runs[0]) },
		"duplicate-run":           func(r *FirstHourExperimentReport) { r.Runs[0] = r.Runs[1] },
		"unknown-seed":            func(r *FirstHourExperimentReport) { r.Runs[0].Key.Seed = "999999" },
		"unknown-persona":         func(r *FirstHourExperimentReport) { r.Runs[0].Key.PolicyID = "invented" },
		"failed":                  func(r *FirstHourExperimentReport) { r.Runs[0].Outcome = "failed" },
		"invariant-failure":       func(r *FirstHourExperimentReport) { r.Runs[0].InvariantFailures = []string{"guard exhaustion"} },
		"missing-ending":          func(r *FirstHourExperimentReport) { r.Runs[0].Ending = nil },
		"no-transitions":          func(r *FirstHourExperimentReport) { r.Runs[0].TransitionCount = 0 },
		"missing-clock":           func(r *FirstHourExperimentReport) { r.Runs[0].Milestones[0].FirstMS = nil },
		"empty":                   func(r *FirstHourExperimentReport) { r.Runs = nil },
	}
	for name, mutate := range mutations {
		t.Run(name, func(t *testing.T) {
			copy := cloneReputationThresholdSource(t, report)
			mutate(&copy)
			oracleErr := validateReputationFirstHourPopulation(suite, experiment, copy)
			measured, err := measureReputationThresholdStudy(suite, experiment, copy, policy, []string{"1e6"})
			t.Logf("existing H3 oracle=%v; admitted H2 rows=%d error=%v", oracleErr, len(measured.Rows), err)
			if err == nil {
				t.Fatal("pinned-study calculation lacked source/population admission; generic calculator alone is not complete-study proof")
			}
		})
	}
	for _, name := range []string{"historical-full", "generic-subset", "reordered-full"} {
		t.Run(name, func(t *testing.T) {
			copy := cloneReputationThresholdSource(t, report)
			if name == "generic-subset" {
				var chosen []FirstHourRunResult
				for _, id := range []string{"casual.t0_t1", "chaos.t0_t1"} {
					for _, run := range copy.Runs {
						if run.Key.PolicyID == id {
							chosen = append(chosen, run)
							break
						}
					}
				}
				copy.Runs = chosen
			}
			if name == "reordered-full" {
				for left, right := 0, len(copy.Runs)-1; left < right; left, right = left+1, right-1 {
					copy.Runs[left], copy.Runs[right] = copy.Runs[right], copy.Runs[left]
				}
			}
			measure := func() (ReputationThresholdReport, error) {
				if name == "generic-subset" {
					return MeasureReputationThresholds(copy, policy, []string{"1e6"}, ReputationThresholdEnvelope{Minimum: 3, Maximum: 10},
						[]string{"casual.t0_t1", "chaos.t0_t1"})
				}
				return measureReputationThresholdStudy(suite, experiment, copy, policy, []string{"1e6"})
			}
			measured, err := measure()
			if err != nil || len(measured.Rows) != 1 || len(measured.Rows[0].Personas) < 2 {
				t.Fatalf("legal generic calculation rejected: %v", err)
			}
		})
	}
}

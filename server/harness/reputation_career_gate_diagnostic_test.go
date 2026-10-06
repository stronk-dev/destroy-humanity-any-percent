package harness

import (
	"reflect"
	"testing"
)

func reputationGateClock(value int64) *int64 { return &value }

func reputationGateDiagnosticRow(treated, control *int64, excluded string) reputationCareerSeed {
	row := reputationCareerSeed{PolicyID: "diagnostic", Seed: 0, AppliedStarterIDs: []string{"synthetic.starter"},
		TreatedGateMS: treated, ControlGateMS: control, Excluded: excluded}
	if treated != nil && control != nil {
		row.SavedMS = reputationGateClock(*control - *treated)
	}
	return row
}

// These rows test the report oracle, not production or naturally earned careers.
func TestReputationCareerGateExclusionDiagnostic(t *testing.T) {
	const reason = "run3_gate_beyond_ratified_horizon_in_both_arms"
	tests := []struct {
		name       string
		row        reputationCareerSeed
		violations int
		gated      int
		excluded   int
		saved      []int64
	}{
		{"earlier", reputationGateDiagnosticRow(reputationGateClock(9), reputationGateClock(10), ""), 0, 1, 0, []int64{1}},
		{"treated-only", reputationGateDiagnosticRow(reputationGateClock(9), nil, ""), 0, 1, 0, nil},
		{"valid-both-unreached", reputationGateDiagnosticRow(nil, nil, reason), 0, 0, 1, nil},
		{"tie", reputationGateDiagnosticRow(reputationGateClock(10), reputationGateClock(10), ""), 1, 1, 0, nil},
		{"slower", reputationGateDiagnosticRow(reputationGateClock(11), reputationGateClock(10), ""), 1, 1, 0, nil},
		{"missing-treatment", reputationGateDiagnosticRow(nil, reputationGateClock(10), ""), 1, 1, 0, nil},
		{"no-starter", func() reputationCareerSeed {
			r := reputationGateDiagnosticRow(reputationGateClock(9), reputationGateClock(10), "")
			r.AppliedStarterIDs = nil
			return r
		}(), 0, 0, 0, nil},
		{"no-starter-valid-exclusion", func() reputationCareerSeed {
			r := reputationGateDiagnosticRow(nil, nil, reason)
			r.AppliedStarterIDs = nil
			return r
		}(), 0, 0, 0, nil},
		{"false-exclusion-tie", reputationGateDiagnosticRow(reputationGateClock(10), reputationGateClock(10), reason), 1, 0, 0, nil},
		{"false-exclusion-slower", reputationGateDiagnosticRow(reputationGateClock(11), reputationGateClock(10), reason), 1, 0, 0, nil},
		{"false-exclusion-earlier", reputationGateDiagnosticRow(reputationGateClock(9), reputationGateClock(10), reason), 1, 0, 0, nil},
		{"false-exclusion-treated-only", reputationGateDiagnosticRow(reputationGateClock(9), nil, reason), 1, 0, 0, nil},
		{"false-exclusion-control-only", reputationGateDiagnosticRow(nil, reputationGateClock(10), reason), 1, 0, 0, nil},
		{"unknown-exclusion", reputationGateDiagnosticRow(nil, nil, "owner_exempt:invented"), 1, 0, 0, nil},
		{"unlabelled-both-unreached", reputationGateDiagnosticRow(nil, nil, ""), 1, 0, 0, nil},
		{"no-starter-unknown-exclusion", func() reputationCareerSeed {
			r := reputationGateDiagnosticRow(nil, nil, "owner_exempt:invented")
			r.AppliedStarterIDs = nil
			return r
		}(), 1, 0, 0, nil},
		{"missing-saving", func() reputationCareerSeed {
			r := reputationGateDiagnosticRow(reputationGateClock(9), reputationGateClock(10), "")
			r.SavedMS = nil
			return r
		}(), 1, 0, 0, nil},
		{"wrong-saving", func() reputationCareerSeed {
			r := reputationGateDiagnosticRow(reputationGateClock(9), reputationGateClock(10), "")
			r.SavedMS = reputationGateClock(7)
			return r
		}(), 1, 0, 0, nil},
		{"wrong-saving-tie", func() reputationCareerSeed {
			r := reputationGateDiagnosticRow(reputationGateClock(10), reputationGateClock(10), "")
			r.SavedMS = reputationGateClock(7)
			return r
		}(), 1, 0, 0, nil},
		{"finite-saving-treated-only", func() reputationCareerSeed {
			r := reputationGateDiagnosticRow(reputationGateClock(9), nil, "")
			r.SavedMS = reputationGateClock(7)
			return r
		}(), 1, 0, 0, nil},
		{"finite-saving-control-only", func() reputationCareerSeed {
			r := reputationGateDiagnosticRow(nil, reputationGateClock(10), "")
			r.SavedMS = reputationGateClock(7)
			return r
		}(), 1, 0, 0, nil},
		{"finite-saving-both-unreached", func() reputationCareerSeed {
			r := reputationGateDiagnosticRow(nil, nil, "")
			r.SavedMS = reputationGateClock(7)
			return r
		}(), 1, 0, 0, nil},
		{"finite-saving-excluded", func() reputationCareerSeed {
			r := reputationGateDiagnosticRow(nil, nil, reason)
			r.SavedMS = reputationGateClock(7)
			return r
		}(), 1, 0, 0, nil},
		{"negative-treatment", reputationGateDiagnosticRow(reputationGateClock(-1), reputationGateClock(10), ""), 1, 0, 0, nil},
		{"negative-control", reputationGateDiagnosticRow(reputationGateClock(9), reputationGateClock(-1), ""), 1, 0, 0, nil},
		{"no-starter-wrong-saving", func() reputationCareerSeed {
			r := reputationGateDiagnosticRow(reputationGateClock(9), reputationGateClock(10), "")
			r.SavedMS, r.AppliedStarterIDs = reputationGateClock(7), nil
			return r
		}(), 1, 0, 0, nil},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			defer func() {
				if failure := recover(); failure != nil {
					t.Errorf("oracle panicked on a report row rather than refusing it: %v", failure)
				}
			}()
			report := reputationCareerReport{}
			violations, saved := evaluateReputationCareerGate([]reputationCareerSeed{test.row}, &report)
			t.Logf("observed violations=%v gated=%d excluded=%d saving=%v", violations, report.GatedSeeds, report.ExcludedSeeds, saved)
			if len(violations) != test.violations || report.GatedSeeds != test.gated || report.ExcludedSeeds != test.excluded ||
				!reflect.DeepEqual(saved[test.row.PolicyID], test.saved) {
				t.Fatalf("H4 oracle admitted/suppressed invalid evidence or changed a healthy control; expected violations/gated/excluded=%d/%d/%d saving=%v",
					test.violations, test.gated, test.excluded, test.saved)
			}
		})
	}
	t.Run("mixed-false-exclusion", func(t *testing.T) {
		rows := []reputationCareerSeed{tests[0].row, tests[2].row, tests[8].row, tests[6].row}
		report := reputationCareerReport{}
		violations, saved := evaluateReputationCareerGate(rows, &report)
		if len(violations) != 1 || report.GatedSeeds != 1 || report.ExcludedSeeds != 1 || !reflect.DeepEqual(saved["diagnostic"], []int64{1}) {
			t.Fatalf("a false exclusion hid a failed row inside healthy evidence: violations=%v gated=%d excluded=%d saved=%v", violations, report.GatedSeeds, report.ExcludedSeeds, saved)
		}
	})
	t.Run("fresh-census-on-every-invocation", func(t *testing.T) {
		rows := []reputationCareerSeed{tests[0].row, tests[2].row, tests[3].row, tests[6].row}
		report := reputationCareerReport{GatedSeeds: 99, ExcludedSeeds: 77}
		for invocation := 0; invocation < 2; invocation++ {
			violations, saved := evaluateReputationCareerGate(rows, &report)
			if len(violations) != 1 || report.GatedSeeds != 2 || report.ExcludedSeeds != 1 || !reflect.DeepEqual(saved["diagnostic"], []int64{1}) {
				t.Fatalf("stale census changed the current measurement at invocation%d: violations=%v gated=%d excluded=%d saved=%v", invocation, violations, report.GatedSeeds, report.ExcludedSeeds, saved)
			}
		}
	})
}

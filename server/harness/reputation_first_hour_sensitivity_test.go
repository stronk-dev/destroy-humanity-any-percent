package harness

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"testing"

	"cloud-clicker/server/decimal"
	"cloud-clicker/server/multiplier"
	"cloud-clicker/server/production"
	"cloud-clicker/server/save"
)

var reputationFirstHourObservation = flag.String("reputation-first-hour-observe", "", "explicit full H3 study; synthetic inputs, never updates retained reports")

func validateReputationFirstHourPopulation(suite *FirstHourSuite, experiment FirstHourExperiment, report FirstHourExperimentReport) error {
	if report.SchemaVersion != 1 || report.ScenarioID != suite.Scenario.ID || report.ScenarioHash != suite.ScenarioHash ||
		report.PolicyHash != suite.PolicyHash || report.ConstantsHash != suite.ConstantsHash || report.Experiment != experiment {
		return fmt.Errorf("H3 population source or experiment mismatch")
	}
	expected := map[RunKey]bool{}
	for _, spec := range suite.Scenario.Runs {
		start, err := strconv.ParseUint(spec.SeedStart, 10, 64)
		if err != nil || spec.SeedCount < 1 {
			return fmt.Errorf("H3 population has invalid declared seeds")
		}
		for offset := 0; offset < spec.SeedCount; offset++ {
			expected[suite.RunKey(spec, start+uint64(offset))] = true
		}
	}
	if len(report.Runs) != len(expected) || report.Aggregate.RunCount != len(expected) {
		return fmt.Errorf("H3 population incomplete: got%d expected%d aggregate%d", len(report.Runs), len(expected), report.Aggregate.RunCount)
	}
	for _, run := range report.Runs {
		if !expected[run.Key] {
			return fmt.Errorf("H3 missing, duplicate or unknown run %s/%s", run.Key.PolicyID, run.Key.Seed)
		}
		delete(expected, run.Key)
		if run.PolicyHash != suite.PolicyHash || run.Outcome != "completed" || len(run.InvariantFailures) != 0 || run.TransitionCount < 1 || run.Ending == nil {
			return fmt.Errorf("H3 invalid run %s/%s", run.Key.PolicyID, run.Key.Seed)
		}
		if len(run.Milestones) != len(suite.Scenario.Milestones) {
			return fmt.Errorf("H3 incomplete milestone population")
		}
		for index, definition := range suite.Scenario.Milestones {
			marker := run.Milestones[index]
			if marker.ID != definition.ID || marker.FirstMS == nil || *marker.FirstMS < 0 {
				return fmt.Errorf("H3 absent or invalid milestone %s/%s/%s", run.Key.PolicyID, run.Key.Seed, definition.ID)
			}
		}
		if len(run.ReputationExits) != 2 {
			return fmt.Errorf("H3 missing Exit observations")
		}
		for index, exit := range run.ReputationExits {
			lifetime, err := decimal.ParseCanonical(exit.LifetimeValue)
			kind := "scripted_first"
			if index == 1 {
				kind = "collapse"
			}
			if err != nil || !lifetime.Gt(decimal.Zero) || exit.RunSeq != int64(index+1) || exit.ExitType != kind ||
				exit.ReputationDelta != 0 || exit.LevelBefore != 0 || exit.LevelAfter != 0 || exit.AvailableAfter != 0 {
				return fmt.Errorf("H3 invalid empty-plan Exit observation %s/%s", run.Key.PolicyID, run.Key.Seed)
			}
		}
	}
	return nil
}

func reputationFirstHourDistributions(report FirstHourExperimentReport) map[string][]int64 {
	values := map[string][]int64{}
	for _, run := range report.Runs {
		for _, marker := range run.Milestones {
			key := run.Key.PolicyID + "/" + marker.ID
			values[key] = append(values[key], *marker.FirstMS)
		}
	}
	for _, times := range values {
		sort.Slice(times, func(i, j int) bool { return times[i] < times[j] })
	}
	return values
}

func requireReputationFirstHourSensitivity(before, after FirstHourExperimentReport) error {
	if reflect.DeepEqual(reputationFirstHourDistributions(before), reputationFirstHourDistributions(after)) {
		return fmt.Errorf("H3 declared factor moved no milestone distribution")
	}
	return nil
}

func reputationFirstHourBySeed(report FirstHourExperimentReport) map[string]FirstHourRunResult {
	rows := map[string]FirstHourRunResult{}
	for _, run := range report.Runs {
		rows[run.Key.PolicyID+"/"+run.Key.Seed] = run
	}
	return rows
}

func reputationFirstHourChanged(before, after FirstHourExperimentReport) []string {
	rows := reputationFirstHourBySeed(before)
	changes := []string{}
	for _, right := range after.Runs {
		left := rows[right.Key.PolicyID+"/"+right.Key.Seed]
		for index, marker := range right.Milestones {
			previous := left.Milestones[index]
			if *previous.FirstMS != *marker.FirstMS {
				changes = append(changes, fmt.Sprintf("%s/%s/%s:%d->%d", right.Key.PolicyID, right.Key.Seed, marker.ID, *previous.FirstMS, *marker.FirstMS))
			}
		}
	}
	return changes
}

func TestReputationFirstHourSensitivityOracle(t *testing.T) {
	suite, _, experiment, _ := reputationCareerAdmissionInputs(t)
	// Explicitly synthetic rows test the oracle, not production behavior.
	report := FirstHourExperimentReport{SchemaVersion: 1, ScenarioID: suite.Scenario.ID, ScenarioHash: suite.ScenarioHash,
		PolicyHash: suite.PolicyHash, ConstantsHash: suite.ConstantsHash, Experiment: experiment}
	for _, spec := range suite.Scenario.Runs {
		for seed := 0; seed < spec.SeedCount; seed++ {
			run := FirstHourRunResult{Key: suite.RunKey(spec, uint64(seed)), PolicyHash: suite.PolicyHash, Outcome: "completed", TransitionCount: 1,
				Ending: &FirstHourEndingSample{}, ReputationExits: []FirstHourReputationSample{
					{RunSeq: 1, ExitType: "scripted_first", LifetimeValue: "1e6"}, {RunSeq: 2, ExitType: "collapse", LifetimeValue: "1e6"}}}
			for _, marker := range suite.Scenario.Milestones {
				value := int64(1000)
				run.Milestones = append(run.Milestones, TimedMilestone{ID: marker.ID, FirstMS: &value})
			}
			report.Runs = append(report.Runs, run)
		}
	}
	report.Aggregate.RunCount = len(report.Runs)
	clone := func() FirstHourExperimentReport {
		data, err := json.Marshal(report)
		if err != nil {
			t.Fatal(err)
		}
		var result FirstHourExperimentReport
		if err := json.Unmarshal(data, &result); err != nil {
			t.Fatal(err)
		}
		return result
	}
	if err := validateReputationFirstHourPopulation(suite, experiment, report); err != nil {
		t.Fatal(err)
	}
	mutations := map[string]func(*FirstHourExperimentReport){
		"empty":             func(r *FirstHourExperimentReport) { r.Runs = nil },
		"missing":           func(r *FirstHourExperimentReport) { r.Runs = r.Runs[1:] },
		"duplicate":         func(r *FirstHourExperimentReport) { r.Runs[0] = r.Runs[1] },
		"extra":             func(r *FirstHourExperimentReport) { r.Runs = append(r.Runs, r.Runs[0]) },
		"aggregate":         func(r *FirstHourExperimentReport) { r.Aggregate.RunCount-- },
		"source":            func(r *FirstHourExperimentReport) { r.ConstantsHash = "wrong" },
		"run-source":        func(r *FirstHourExperimentReport) { r.Runs[0].Key.ConstantsHash = "wrong" },
		"policy":            func(r *FirstHourExperimentReport) { r.Runs[0].PolicyHash = "wrong" },
		"experiment":        func(r *FirstHourExperimentReport) { r.Experiment.GeneratedBeigeTowers++ },
		"failed":            func(r *FirstHourExperimentReport) { r.Runs[0].Outcome = "failed" },
		"truncated":         func(r *FirstHourExperimentReport) { r.Runs[0].InvariantFailures = []string{"guard exhausted"} },
		"no-transitions":    func(r *FirstHourExperimentReport) { r.Runs[0].TransitionCount = 0 },
		"no-ending":         func(r *FirstHourExperimentReport) { r.Runs[0].Ending = nil },
		"null-clock":        func(r *FirstHourExperimentReport) { r.Runs[0].Milestones[0].FirstMS = nil },
		"negative-clock":    func(r *FirstHourExperimentReport) { *r.Runs[0].Milestones[0].FirstMS = -1 },
		"missing-clock":     func(r *FirstHourExperimentReport) { r.Runs[0].Milestones = r.Runs[0].Milestones[1:] },
		"marker":            func(r *FirstHourExperimentReport) { r.Runs[0].Milestones[0].ID = "wrong" },
		"missing-exits":     func(r *FirstHourExperimentReport) { r.Runs[0].ReputationExits = nil },
		"lifetime":          func(r *FirstHourExperimentReport) { r.Runs[0].ReputationExits[0].LifetimeValue = "NaN" },
		"paid":              func(r *FirstHourExperimentReport) { r.Runs[0].ReputationExits[0].ReputationDelta = 1 },
		"run-sequence":      func(r *FirstHourExperimentReport) { r.Runs[0].ReputationExits[0].RunSeq = 2 },
		"exit-kind":         func(r *FirstHourExperimentReport) { r.Runs[0].ReputationExits[0].ExitType = "collapse" },
		"founder-ownership": func(r *FirstHourExperimentReport) { r.Runs[0].ReputationExits[0].AvailableAfter = 1 },
	}
	for name, mutate := range mutations {
		t.Run(name, func(t *testing.T) {
			broken := clone()
			mutate(&broken)
			if err := validateReputationFirstHourPopulation(suite, experiment, broken); err == nil {
				t.Fatal("H3 oracle silently admitted malformed population")
			}
		})
	}
	t.Run("unchanged-is-not-sensitive", func(t *testing.T) {
		if err := requireReputationFirstHourSensitivity(report, clone()); err == nil {
			t.Fatal("H3 oracle fabricated sensitivity")
		}
	})
	t.Run("changed-valid-clock", func(t *testing.T) {
		changed := clone()
		*changed.Runs[0].Milestones[0].FirstMS++
		if err := validateReputationFirstHourPopulation(suite, experiment, changed); err != nil {
			t.Fatal(err)
		}
		if err := requireReputationFirstHourSensitivity(report, changed); err != nil || len(reputationFirstHourChanged(report, changed)) != 1 {
			t.Fatalf("H3 oracle missed the changed valid clock: %v", err)
		}
	})
	for _, mode := range []string{"career", "tier2"} {
		t.Run("input-refuses-"+mode, func(t *testing.T) {
			copy := *suite
			copy.diagnosticExternal = []multiplier.Contribution{}
			var career *careerRuntime
			var tier2 *tier2Runtime
			if mode == "career" {
				career = &careerRuntime{}
			} else {
				tier2 = &tier2Runtime{}
			}
			result, _, _, _ := copy.runWithModes(suite.Scenario.Runs[0], 0, experiment, false, career, tier2)
			if result.Outcome != "failed" || result.TransitionCount != 0 || len(result.InvariantFailures) == 0 {
				t.Fatal("diagnostic input overwrote a real career/Tier2 mode")
			}
		})
	}
}

// This one Reference row checks the injected input reaches the real consumer.
// It is NOT the all-seed H3 study and does not satisfy the fired tiny criterion.
func TestReputationFirstHourDiagnosticConsumerControl(t *testing.T) {
	suite, _, experiment, _ := reputationCareerAdmissionInputs(t)
	suite.Bundle = reputationCareerBundle(t, suite)
	suite.ConstantsHash = suite.Bundle.ConstantsHash
	var spec RunSpec
	for _, candidate := range suite.Scenario.Runs {
		if candidate.PolicyID == "reference.greedy" {
			spec = candidate
		}
	}
	if spec.PolicyVersion != 1 || spec.SeedCount != 1 || spec.HorizonMS != 7_200_000 {
		t.Fatal("single-row control lost its actual Reference population")
	}
	for _, arm := range []struct {
		name, factor, firstLifetime, secondLifetime string
		transitions                                 int64
		clocks                                      []int64
	}{
		{"none", "", "1.4605083614e6", "3.54431965065e6", 11916, []int64{0, 10000, 66992, 356000, 900000, 346000, 2700000}},
		{"unit", "1e0", "1.4605083614e6", "3.54431965065e6", 11916, []int64{0, 10000, 66992, 356000, 900000, 346000, 2700000}},
		{"tiny", "1.000001e0", "1.46050991148e6", "3.54432330408e6", 11915, []int64{0, 10000, 66992, 356000, 900000, 346000, 2700000}},
		{"strong", "2e0", "3.5053281185e6", "7.66194816889e6", 8947, []int64{0, 10000, 44072, 182000, 900000, 177000, 2700000}},
	} {
		t.Run(arm.name, func(t *testing.T) {
			copy := *suite
			if arm.factor != "" {
				bonus := copy.Bundle.ReputationTree.Bonus
				var err error
				copy.diagnosticExternal, err = production.ResolveFrozenContributions(copy.Bundle.Economy, []save.FrozenContribution{{
					SourceID: bonus.SourceID, Slot: bonus.Slot, Target: bonus.Target, Factor: arm.factor,
				}})
				if err != nil {
					t.Fatal(err)
				}
			}
			result := copy.RunExperiment(spec, 0, experiment)
			if result.Outcome != "completed" || len(result.InvariantFailures) != 0 || len(result.Milestones) != 7 || len(result.ReputationExits) != 2 {
				t.Fatalf("single-row control failed or truncated: %+v", result)
			}
			clocks := []int64{}
			for _, marker := range result.Milestones {
				if marker.FirstMS == nil {
					t.Fatal("single-row control has a missing clock")
				}
				clocks = append(clocks, *marker.FirstMS)
			}
			if !reflect.DeepEqual(clocks, arm.clocks) || result.ReputationExits[0].LifetimeValue != arm.firstLifetime ||
				result.ReputationExits[1].LifetimeValue != arm.secondLifetime || result.TransitionCount != arm.transitions {
				t.Fatalf("diagnostic input lost measured consumer behavior: clocks=%v lifetime=%s/%s transitions=%d", clocks,
					result.ReputationExits[0].LifetimeValue, result.ReputationExits[1].LifetimeValue, result.TransitionCount)
			}
		})
	}
}

func TestReputationFirstHourSensitivityObservation(t *testing.T) {
	if *reputationFirstHourObservation == "" {
		t.Skip("full H3 synthetic sensitivity study requires -reputation-first-hour-observe=all; no retained report is updated")
	}
	if *reputationFirstHourObservation != "all" {
		t.Fatal("H3 selector must name the full all-seed study")
	}
	suite, _, experiment, _ := reputationCareerAdmissionInputs(t)
	data, err := os.ReadFile(filepath.Join(repositoryRootForReputation, "planning/prestige-and-exits/first-hour-epoch8-report.v1.json"))
	if err != nil {
		t.Fatal(err)
	}
	var retained FirstHourExperimentReport
	if err := json.Unmarshal(data, &retained); err != nil || retained.Experiment != experiment {
		t.Fatalf("retained epoch8 experiment mismatch: %v", err)
	}
	paired := *suite
	paired.Bundle = reputationCareerBundle(t, suite)
	paired.ConstantsHash = paired.Bundle.ConstantsHash
	if paired.Bundle.Prestige.Threshold != "1e12" {
		t.Fatal("H3 accidentally used the paid-career fixture threshold")
	}
	reports := map[string]FirstHourExperimentReport{}
	for _, arm := range []string{"epoch8", "none", "unit", "tiny", "strong"} {
		copy := paired
		factor := "absent"
		if arm == "epoch8" {
			copy = *suite
		} else if arm != "none" {
			factor = map[string]string{"unit": "1e0", "tiny": "1.000001e0", "strong": "2e0"}[arm]
			bonus := paired.Bundle.ReputationTree.Bonus
			copy.diagnosticExternal, err = production.ResolveFrozenContributions(copy.Bundle.Economy, []save.FrozenContribution{{
				SourceID: bonus.SourceID, Slot: bonus.Slot, Target: bonus.Target, Factor: factor,
			}})
			if err != nil {
				t.Fatal(err)
			}
		}
		report, err := copy.RunAllExperiments(experiment, 8)
		if err != nil {
			t.Fatal(err)
		}
		if err := validateReputationFirstHourPopulation(&copy, experiment, report); err != nil {
			t.Fatalf("arm%s invalid: %v", arm, err)
		}
		reports[arm] = report
		t.Logf("H3 arm=%s factor=%s source=%s scenario=%s policy=%s threshold=%s runs=%d markers=%d", arm, factor,
			copy.ConstantsHash, copy.ScenarioHash, copy.PolicyHash, copy.Bundle.Prestige.Threshold, len(report.Runs), len(report.Runs)*len(copy.Scenario.Milestones))
		for _, run := range report.Runs {
			clocks := []int64{}
			for _, marker := range run.Milestones {
				clocks = append(clocks, *marker.FirstMS)
			}
			t.Logf("%s %s/%s clocks=%v ending=%s lifetime=%s/%s paid=%d/%d transitions=%d", arm, run.Key.PolicyID, run.Key.Seed, clocks,
				run.Ending.Branch, run.ReputationExits[0].LifetimeValue, run.ReputationExits[1].LifetimeValue,
				run.ReputationExits[0].ReputationDelta, run.ReputationExits[1].ReputationDelta, run.TransitionCount)
		}
	}
	// Historical epoch8 has no H1 fields: compare fresh clocks/ending and its
	// original source coordinates, not fabricated Exit observations.
	base, none, unit := reports["epoch8"], reports["none"], reports["unit"]
	if retained.ConstantsHash != base.ConstantsHash || retained.ScenarioHash != base.ScenarioHash || retained.PolicyHash != base.PolicyHash || len(retained.Runs) != 97 {
		t.Fatal("epoch8 report has incompatible source coordinates or population")
	}
	for _, comparison := range []struct {
		name        string
		left, right FirstHourExperimentReport
		full        bool
	}{{"retained-epoch8", retained, base, false}, {"paired-nil", base, none, false}, {"explicit-unit", none, unit, true}} {
		left := reputationFirstHourBySeed(comparison.left)
		for _, right := range comparison.right.Runs {
			previous, ok := left[right.Key.PolicyID+"/"+right.Key.Seed]
			if !ok || !reflect.DeepEqual(previous.Milestones, right.Milestones) || !reflect.DeepEqual(previous.Ending, right.Ending) ||
				comparison.full && (!reflect.DeepEqual(previous.ReputationExits, right.ReputationExits) || previous.TransitionCount != right.TransitionCount) {
				t.Errorf("H3 %s moved %s/%s", comparison.name, right.Key.PolicyID, right.Key.Seed)
			}
		}
		if !reflect.DeepEqual(reputationFirstHourDistributions(comparison.left), reputationFirstHourDistributions(comparison.right)) {
			t.Errorf("H3 %s distributions moved", comparison.name)
		}
	}
	for _, arm := range []string{"tiny", "strong"} {
		changes := reputationFirstHourChanged(none, reports[arm])
		t.Logf("H3 %s changed-markers=%d changes=%v", arm, len(changes), changes)
		if err := requireReputationFirstHourSensitivity(none, reports[arm]); err != nil || len(changes) == 0 {
			t.Errorf("H3 %s sensitivity criterion fired: %v", arm, err)
		}
	}
}

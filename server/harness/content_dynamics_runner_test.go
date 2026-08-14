package harness

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCandidateContentDynamicsScenarioAndReport(t *testing.T) {
	root := filepath.Join("..", "..")
	suite, err := LoadCandidateContentDynamicsSuite(root,
		"testdata/harness/content-dynamics/scenarios/epoch-7-candidate.v1.json",
		"planning/t0-t1-content/promotion-manifest.candidate.v1.json")
	if err != nil {
		t.Fatal(err)
	}
	report, err := suite.Run()
	if err != nil {
		t.Fatal(err)
	}
	if report.DeclaredRuns != 69 || report.ExecutedRuns != 69 || report.DeclaredTransitions != 4215 ||
		report.ExecutedTransitions != 786 || len(report.Observations) != 27 {
		t.Fatalf("content-dynamics report=%+v", report)
	}
	values := map[string]string{}
	for _, row := range report.Observations {
		values[row.RunID+"/"+row.MetricID+"/"+row.Statistic] = row.Value
	}
	if values["active_play.window/active_play.bonus_output/exact"] == "" ||
		values["active_play.window/active_play.spawned_attended_ms/exact"] != "3558" ||
		values["fiscal.one_period/fiscal.credit_after/exact"] != "3" ||
		values["fiscal.four_periods/fiscal.credit_after/exact"] != "12" ||
		values["permit.zero/permits.time_to_12_ms/p50"] != "11979232" ||
		values["permit.hoard/permits.time_to_12_ms/p50"] != "5989616" ||
		values["pitch.seed_sweep/pitch.final_round/p50"] != "3" ||
		values["pitch.seed_sweep/pitch.final_round/p95"] != "4" {
		t.Fatalf("content-dynamics observations=%v", values)
	}
	encoded, err := CanonicalJSON(report)
	if err != nil {
		t.Fatal(err)
	}
	var decoded ContentDynamicsReport
	if json.Unmarshal(encoded, &decoded) != nil || ValidateContentDynamicsReport(decoded, suite.Scenario) != nil {
		t.Fatalf("content-dynamics report does not round-trip: %s", encoded)
	}
	missing := decoded
	missing.Observations = append([]ContentDynamicsObservation(nil), decoded.Observations[1:]...)
	if ValidateContentDynamicsReport(missing, suite.Scenario) == nil {
		t.Fatal("content-dynamics report accepted a missing observation")
	}
	wrongMetric := decoded
	wrongMetric.Observations = append([]ContentDynamicsObservation(nil), decoded.Observations...)
	wrongMetric.Observations[0].MetricID = "active_play.unruled"
	if ValidateContentDynamicsReport(wrongMetric, suite.Scenario) == nil {
		t.Fatal("content-dynamics report accepted an unruled observation")
	}
	repeated, err := suite.Run()
	if err != nil {
		t.Fatal(err)
	}
	repeatedBytes, err := CanonicalJSON(repeated)
	if err != nil || !bytes.Equal(encoded, repeatedBytes) {
		t.Fatalf("content-dynamics rerun drift: err=%v\nfirst=%s\nsecond=%s", err, encoded, repeatedBytes)
	}
}

func TestContentDynamicsScenarioRejectsInexactAndVacuousInputs(t *testing.T) {
	root := filepath.Join("..", "..")
	path := filepath.Join(root, "testdata", "harness", "content-dynamics", "scenarios", "epoch-7-candidate.v1.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	mutations := map[string]func(map[string]any){
		"unknown root": func(root map[string]any) { root["extra"] = true },
		"wrong budget": func(root map[string]any) { root["transition_budget"] = float64(4214) },
		"unsafe seed": func(root map[string]any) {
			root["runs"].([]any)[5].(map[string]any)["seed_start"] = "18446744073709551600"
		},
		"duplicate id": func(root map[string]any) { root["runs"].([]any)[1].(map[string]any)["id"] = "active_play.window" },
		"unknown policy key": func(root map[string]any) {
			root["runs"].([]any)[0].(map[string]any)["policy"].(map[string]any)["extra"] = true
		},
		"vacuous draw": func(root map[string]any) {
			root["runs"].([]any)[0].(map[string]any)["policy"].(map[string]any)["effect_row_id"] = "active.lucky"
		},
	}
	for name, mutate := range mutations {
		t.Run(name, func(t *testing.T) {
			var rootValue map[string]any
			if json.Unmarshal(data, &rootValue) != nil {
				t.Fatal("fixture decode")
			}
			mutate(rootValue)
			changed, _ := json.Marshal(rootValue)
			scenario, loadErr := loadContentDynamicsScenario(changed)
			if name == "vacuous draw" {
				if loadErr != nil {
					t.Fatal(loadErr)
				}
				suite, suiteErr := LoadCandidateContentDynamicsSuite(root,
					"testdata/harness/content-dynamics/scenarios/epoch-7-candidate.v1.json",
					"planning/t0-t1-content/promotion-manifest.candidate.v1.json")
				if suiteErr != nil {
					t.Fatal(suiteErr)
				}
				suite.Scenario = scenario
				if _, runErr := suite.Run(); runErr == nil || !strings.Contains(runErr.Error(), "pinned first draw changed") {
					t.Fatalf("vacuous active draw run error=%v", runErr)
				}
				return
			}
			if loadErr == nil {
				t.Fatalf("mutation accepted: %s", bytes.TrimSpace(changed))
			}
		})
	}
}

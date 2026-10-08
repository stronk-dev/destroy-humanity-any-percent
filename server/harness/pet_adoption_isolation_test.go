package harness

import (
	"bytes"
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"testing"
	"time"

	"cloud-clicker/server/economy"
	"cloud-clicker/server/production"
	"cloud-clicker/server/replaycatalog"
	"cloud-clicker/server/save"
)

var petAdoptionPacingPopulation = flag.String("pet-adoption-pacing", "representative", "adoption isolation population: representative (one seed of every policy) or all (97 seeds)")

// Reuse the recorded, pinned adoption case rather than creating a second
// catalog or draw authority. The optional callback injects an economic leak.
func petAdoptionPacingInputs(t *testing.T, violate func(*save.State)) (*FirstHourSuite, []save.FrozenContribution, []save.FrozenContribution) {
	t.Helper()
	var corpus struct {
		Bundles map[string]struct {
			Hash      string            `json:"constants_hash"`
			Artifacts map[string]string `json:"artifacts"`
		} `json:"bundles"`
		Cases []struct {
			Name     string          `json:"name"`
			Bundle   string          `json:"bundle"`
			Version  int             `json:"state_version"`
			PreState json.RawMessage `json:"pre_state"`
			Payload  json.RawMessage `json:"canonical_payload"`
			Inputs   json.RawMessage `json:"replay_inputs"`
		} `json:"cases"`
	}
	data, err := os.ReadFile(filepath.Join(repositoryRootForReputation, "testdata/replay/pet-adoption-v1.json"))
	if err != nil {
		t.Fatalf("adoption corpus: %v", err)
	}
	if err := json.Unmarshal(data, &corpus); err != nil {
		t.Fatal(err)
	}
	for _, row := range corpus.Cases {
		if row.Name != "applies-starter-adoption" {
			continue
		}
		source, ok := corpus.Bundles[row.Bundle]
		if !ok {
			t.Fatal("adoption case has no pinned bundle")
		}
		artifacts := make(map[string][]byte, len(source.Artifacts))
		for name, raw := range source.Artifacts {
			artifacts[name] = []byte(raw)
		}
		bundle, err := replaycatalog.Load(source.Hash, artifacts)
		if err != nil {
			t.Fatal(err)
		}
		founder, err := save.RestoreState(row.PreState, row.Version, bundle.Economy, economy.ScopeFounder, time.Time{})
		if err != nil {
			t.Fatal(err)
		}
		if len(founder.Pets) != 0 || len(founder.PetIdentities) != 0 {
			t.Fatal("control Founder already has a pet")
		}
		// An admitted, non-unit economic input ensures this is not merely two
		// empty contribution lists taking an unconnected simulation path.
		founder.FiscalGeneratorLevels["generator.beige_tower"] = 1
		before, err := production.FrozenFounderContributions(bundle, founder)
		if err != nil {
			t.Fatal(err)
		}
		nonunit := false
		for _, value := range before {
			nonunit = nonunit || value.Factor != "1e0"
		}
		if !nonunit {
			t.Fatal("pacing control has no non-unit economic input")
		}
		var payload map[string]json.RawMessage
		if err := json.Unmarshal(row.Payload, &payload); err != nil {
			t.Fatal(err)
		}
		canonical, err := json.Marshal(payload)
		if err != nil {
			t.Fatal(err)
		}
		transition, err := production.ApplyFounderLogged(founder, canonical, bundle, row.Inputs)
		if err != nil || transition.Outcome != save.IntentApplied || len(founder.Pets) != 1 || len(founder.PetIdentities) != 1 {
			t.Fatalf("actual adoption did not apply: outcome=%s err=%v", transition.Outcome, err)
		}
		if violate != nil {
			violate(founder)
		}
		after, err := production.FrozenFounderContributions(bundle, founder)
		if err != nil {
			t.Fatal(err)
		}
		suite, err := LoadFirstHourSuite(repositoryRootForReputation, "balance/testdata/t0-t1/harness-scenario-v1.json", "balance/testdata/t0-t1/first-hour-policy-v1.json")
		if err != nil {
			t.Fatal(err)
		}
		suite.Bundle, suite.ConstantsHash = bundle, bundle.ConstantsHash
		return suite, before, after
	}
	t.Fatal("applied adoption case is absent")
	return nil, nil, nil
}

func petAdoptionPacingReport(t *testing.T, suite *FirstHourSuite, values []save.FrozenContribution) FirstHourExperimentReport {
	t.Helper()
	copy := *suite
	var err error
	// This is the existing bounded diagnostic seam, not a new earned bonus,
	// production policy, content mint or change to the retained pacing report.
	copy.diagnosticExternal, err = production.ResolveFrozenContributions(copy.Bundle.Economy, values)
	if err != nil {
		t.Fatal(err)
	}
	experiment := FirstHourExperiment{AcquihirePurchasedMinimum: 200, BurnoutPriceFactor: "2e0", RouteKnowledgeBonus: 50, SeedCapital: "1e4", GeneratedBeigeTowers: 10}
	report, err := copy.RunAllExperiments(experiment, 4)
	if err != nil {
		t.Fatal(err)
	}
	for _, run := range report.Runs {
		if run.Outcome != "completed" || len(run.InvariantFailures) != 0 || run.Ending == nil || run.TransitionCount < 1 {
			t.Fatalf("incomplete pacing run %s/%s: %+v", run.Key.PolicyID, run.Key.Seed, run)
		}
		for _, milestone := range run.Milestones {
			if milestone.FirstMS == nil {
				t.Fatalf("missing pacing milestone %s/%s/%s", run.Key.PolicyID, run.Key.Seed, milestone.ID)
			}
		}
	}
	return report
}

func TestPetAdoptionPacingOutputsAreUnchanged(t *testing.T) {
	suite, before, after := petAdoptionPacingInputs(t, nil)
	expected := 97
	switch *petAdoptionPacingPopulation {
	case "representative":
		// Ordinary CI checks every policy, not a five-minute repeat of the
		// whole study. The explicit all selector retains the full population.
		expected = 3
		suite.Scenario.Runs = append([]RunSpec(nil), suite.Scenario.Runs...)
		for i := range suite.Scenario.Runs {
			suite.Scenario.Runs[i].SeedCount = 1
		}
	case "all":
	default:
		t.Fatalf("unknown adoption pacing population %q", *petAdoptionPacingPopulation)
	}
	policies := map[string]bool{}
	for _, spec := range suite.Scenario.Runs {
		policies[spec.PolicyID] = true
	}
	if len(suite.Scenario.Runs) != 3 || !policies["reference.greedy"] || !policies["casual.t0_t1"] || !policies["chaos.t0_t1"] {
		t.Fatal("adoption comparison lost a declared pacing policy")
	}
	left, right := petAdoptionPacingReport(t, suite, before), petAdoptionPacingReport(t, suite, after)
	if len(left.Runs) != expected || len(right.Runs) != expected {
		t.Fatalf("incomplete declared pacing population: %d/%d", len(left.Runs), len(right.Runs))
	}
	a, errA := json.Marshal(left)
	b, errB := json.Marshal(right)
	if errA != nil || errB != nil || !bytes.Equal(a, b) {
		t.Fatalf("adoption changed pacing outputs: encode errors %v/%v", errA, errB)
	}
	t.Logf("all %d selected paired reference/casual/chaos outputs and their aggregates are byte-identical after actual adoption", expected)
	t.Logf("fixture pacing-envelope failures (retained, not a release pacing verdict): %v", left.Aggregate.Failures)
}

func TestPetAdoptionPacingDetectsAnEconomicLeak(t *testing.T) {
	suite, before, after := petAdoptionPacingInputs(t, func(founder *save.State) {
		founder.FiscalGeneratorLevels["generator.beige_tower"]++
	})
	// The explicit failure control uses the real reference policy/consumer;
	// the successful comparison above checks all three policies.
	for _, spec := range suite.Scenario.Runs {
		if spec.PolicyID == "reference.greedy" {
			suite.Scenario.Runs = []RunSpec{spec}
			break
		}
	}
	left, right := petAdoptionPacingReport(t, suite, before), petAdoptionPacingReport(t, suite, after)
	if len(left.Runs) != 1 || len(right.Runs) != 1 {
		t.Fatal("economic failure control lost its reference population")
	}
	a, _ := json.Marshal(left.Runs[0])
	b, _ := json.Marshal(right.Runs[0])
	if bytes.Equal(a, b) {
		t.Fatal("pacing failed to detect an adoption arm leaking a real Fiscal economic input")
	}
}

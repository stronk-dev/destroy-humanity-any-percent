package harness

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"

	"cloud-clicker/server/production"
	"cloud-clicker/server/replaycatalog"
	"cloud-clicker/server/save"
)

// tier2PacingBundle is the ratified suite bundle with the Tier 2 candidates
// (balance/testdata/t2) swapped in; gateAmount, when set, replaces the
// candidate gate.t1_to_t2 literal for the §P2 calibration sweep.
func tier2PacingBundle(t *testing.T, suite *FirstHourSuite, gateAmount string) production.CatalogBundle {
	t.Helper()
	artifacts := map[string][]byte{}
	for name, data := range suite.Bundle.Artifacts {
		artifacts[name] = data
	}
	for name, file := range map[string]string{"economy": "economy-candidate-v1.json", "routes": "routes-candidate-v1.json", "categories": "categories-candidate-v1.json"} {
		data, err := os.ReadFile(filepath.Join(repositoryRootForReputation, "balance/testdata/t2", file))
		if err != nil {
			t.Fatal(err)
		}
		artifacts[name] = data
	}
	if gateAmount != "" {
		const literal = `"gate_id": "gate.t1_to_t2",
      "requirement": [{ "resource_id": "company.cash", "amount": "1e7" }]`
		if !bytes.Contains(artifacts["routes"], []byte(literal)) {
			t.Fatal("candidate gate literal moved")
		}
		artifacts["routes"] = bytes.Replace(artifacts["routes"], []byte(literal), []byte(strings.Replace(literal, `"1e7"`, `"`+gateAmount+`"`, 1)), 1)
	}
	hash, err := save.ConstantsHashArtifacts(artifacts)
	if err != nil {
		t.Fatal(err)
	}
	if gateAmount == "" {
		pinned, err := os.ReadFile(filepath.Join(repositoryRootForReputation, "balance/testdata/t2/candidate-bundle-hash.txt"))
		if err != nil || string(bytes.TrimSpace(pinned)) != hash {
			t.Fatalf("candidate bundle hash %s differs from the pinned candidate identity", hash)
		}
	}
	bundle, err := replaycatalog.Load(hash, artifacts)
	if err != nil {
		t.Fatal(err)
	}
	return bundle
}

// tier2Experiment is the epoch-8 curriculum tuple the ratified first-hour
// evidence and the composed first-hour run use.
var tier2Experiment = FirstHourExperiment{AcquihirePurchasedMinimum: 200, BurnoutPriceFactor: "2e0", RouteKnowledgeBonus: 50, SeedCapital: "1e4", GeneratedBeigeTowers: 10}

func tier2PacingSuite(t *testing.T) *FirstHourSuite {
	t.Helper()
	suite, err := LoadFirstHourSuite(repositoryRootForReputation, "balance/testdata/t0-t1/harness-scenario-v1.json", "balance/testdata/t0-t1/first-hour-policy-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	return suite
}

func tier2Spec(suite *FirstHourSuite, policyID string, horizonMS int64) RunSpec {
	for _, run := range suite.Scenario.Runs {
		if run.PolicyID == policyID {
			run.HorizonMS = horizonMS
			return run
		}
	}
	return RunSpec{}
}

// TestTier2PacingReachesTheITCompanyGate is the §P mechanics witness on a
// handful of seeds: the exit-once rule lets run 3 continue to gate.t1_to_t2.
// Its failing case is the RFC's own claim: without exit-once, the shipped rule
// Exits every later run at the Garage gate, so the epoch-8 bundle's plain
// first-hour run never crosses gate.t1_to_t2 (it does not even exist there),
// and a pacing run on a bundle without the gate is refused.
func TestTier2PacingReachesTheITCompanyGate(t *testing.T) {
	suite := tier2PacingSuite(t)
	bundle := tier2PacingBundle(t, suite, "")
	result, err := suite.RunTier2Pacing(tier2Spec(suite, "chaos.t0_t1", 21_600_000), 0, Tier2PacingConfig{Bundle: bundle, Experiment: tier2Experiment})
	if err != nil || result.Outcome != "completed" || result.GateFounderAttendedMS == nil || result.GateRunSeq != 3 || result.ElectiveExits != 1 {
		t.Fatalf("chaos seed 0 result=%+v err=%v", result, err)
	}
	if _, err := suite.RunTier2Pacing(tier2Spec(suite, "chaos.t0_t1", 21_600_000), 0, Tier2PacingConfig{Bundle: suite.Bundle, Experiment: tier2Experiment}); err == nil {
		t.Fatal("pacing ran on a bundle without gate.t1_to_t2")
	}
	if _, err := suite.RunTier2Pacing(tier2Spec(suite, "reference.greedy", 21_600_000), 0, Tier2PacingConfig{Bundle: bundle, Experiment: tier2Experiment}); err == nil {
		t.Fatal("held reference.greedy v2 ran without its allocation arms")
	}
	// A horizon that ends before run 3 reaches the gate is a visible must_reach failure, never a dropped sample.
	short, err := suite.RunTier2Pacing(tier2Spec(suite, "chaos.t0_t1", 3_600_000), 0, Tier2PacingConfig{Bundle: bundle, Experiment: tier2Experiment})
	if err != nil || short.Outcome != "failed" || short.GateFounderAttendedMS != nil || len(short.InvariantFailures) != 1 || short.InvariantFailures[0] != "must_reach:milestone.it_company_gate" {
		t.Fatalf("short horizon result=%+v err=%v", short, err)
	}
}

type tier2PacingRow struct {
	GateAmount string              `json:"gate_amount"`
	PolicyID   string              `json:"policy_id"`
	Seeds      int                 `json:"seeds"`
	Reached    int                 `json:"reached"`
	P50MS      *int64              `json:"p50_founder_attended_ms"`
	InEnvelope bool                `json:"in_envelope_2h_3h"`
	Results    []Tier2PacingResult `json:"results"`
}

type tier2PacingReport struct {
	SchemaVersion int              `json:"schema_version"`
	Envelope      [2]int64         `json:"envelope_ms"`
	Note          string           `json:"note"`
	Rows          []tier2PacingRow `json:"rows"`
}

// TestTier2PacingCalibration is the §P2/OD-7 measurement: Chaos (64 seeds,
// 6 h horizon) and Casual (32 seeds, 12 h wall horizon) p50 Founder-attended
// time to gate.t1_to_t2 for a sweep of gate literals, against the [2 h, 3 h]
// envelope. It is exhaustive and runs only under CLOUD_CLICKER_T2_PACING=1;
// T2_PACING_UPDATE=1 rewrites the committed report.
func TestTier2PacingCalibration(t *testing.T) {
	if os.Getenv("CLOUD_CLICKER_T2_PACING") != "1" {
		t.Skip("exhaustive Tier 2 pacing: CLOUD_CLICKER_T2_PACING=1 make test-go GO_PACKAGES=./harness GO_TEST_FLAGS='-count=1 -run TestTier2PacingCalibration -timeout 90m'")
	}
	suite := tier2PacingSuite(t)
	amounts := strings.Fields(os.Getenv("T2_GATE_AMOUNTS"))
	if len(amounts) == 0 {
		amounts = []string{"3e6", "1e7", "3e7", "1e8", "2e8", "3e8", "5e8"}
	}
	policies := []struct {
		id      string
		seeds   int
		horizon int64
	}{{"chaos.t0_t1", 64, 21_600_000}, {"casual.t0_t1", 32, 43_200_000}}
	report := tier2PacingReport{SchemaVersion: 1, Envelope: [2]int64{7_200_000, 10_800_000},
		Note: "Fixture-first §P2 subset: T0–T1 policy literals + exit-once + gate.t1_to_t2; allocation arms and reference.greedy v2 held on OD-1. p50 over all seeds; an unreached seed counts as +inf (never dropped)."}
	for _, amount := range amounts {
		bundle := tier2PacingBundle(t, suite, amount)
		for _, policy := range policies {
			spec := tier2Spec(suite, policy.id, policy.horizon)
			results := make([]Tier2PacingResult, policy.seeds)
			var wait sync.WaitGroup
			errs := make([]error, policy.seeds)
			sem := make(chan struct{}, 8)
			for seed := 0; seed < policy.seeds; seed++ {
				wait.Add(1)
				sem <- struct{}{}
				go func(seed int) {
					defer wait.Done()
					defer func() { <-sem }()
					results[seed], errs[seed] = suite.RunTier2Pacing(spec, uint64(seed), Tier2PacingConfig{Bundle: bundle, Experiment: tier2Experiment})
				}(seed)
			}
			wait.Wait()
			row := tier2PacingRow{GateAmount: amount, PolicyID: policy.id, Seeds: policy.seeds, Results: results}
			values := []int64{}
			for index, result := range results {
				if errs[index] != nil {
					t.Fatalf("%s %s seed %d: %v", amount, policy.id, index, errs[index])
				}
				for _, failure := range result.InvariantFailures {
					if failure != "must_reach:milestone.it_company_gate" {
						t.Fatalf("%s %s seed %d invalid measurement: %v", amount, policy.id, index, result.InvariantFailures)
					}
				}
				if result.GateFounderAttendedMS != nil {
					row.Reached++
					values = append(values, *result.GateFounderAttendedMS)
				}
			}
			sort.Slice(values, func(left, right int) bool { return values[left] < values[right] })
			// p50 = the ceil(n/2)-th smallest over all seeds; unreached seeds sort last as +inf.
			if index := (policy.seeds+1)/2 - 1; index < len(values) {
				p50 := values[index]
				row.P50MS = &p50
				row.InEnvelope = p50 >= report.Envelope[0] && p50 <= report.Envelope[1]
			}
			report.Rows = append(report.Rows, row)
			p50 := "unreached"
			if row.P50MS != nil {
				p50 = strconv.FormatInt(*row.P50MS, 10)
			}
			t.Logf("gate %s %s: reached %d/%d p50=%s ms in_envelope=%v", amount, policy.id, row.Reached, row.Seeds, p50, row.InEnvelope)
		}
	}
	encoded, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	encoded = append(encoded, '\n')
	path := filepath.Join(repositoryRootForReputation, "balance/testdata/t2/pacing-calibration-v1.json")
	if os.Getenv("T2_PACING_UPDATE") == "1" {
		if err := os.WriteFile(path, encoded, 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	committed, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(committed, encoded) {
		t.Fatal(fmt.Sprintf("Tier 2 pacing report drift (err=%v); regenerate with T2_PACING_UPDATE=1", err))
	}
}

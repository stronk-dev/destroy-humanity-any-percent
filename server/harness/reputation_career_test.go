package harness

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"sync"
	"testing"

	"cloud-clicker/server/copykeys"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/production"
	"cloud-clicker/server/reputation"
)

const reputationCareerFixtureThreshold = "1e5"
const reputationCareerReportPath = "planning/reputation-tree-v1/career-h4.v1.json"

func reputationCareerBundle(t *testing.T, suite *FirstHourSuite) production.CatalogBundle {
	t.Helper()
	var root map[string]any
	if err := json.Unmarshal(suite.Bundle.Artifacts["economy"], &root); err != nil {
		t.Fatal(err)
	}
	root["multiplier_sources"] = append(root["multiplier_sources"].([]any), map[string]any{"id": "reputation.founder_bonus", "slot": "prestige", "target": "all", "provider": reputation.Provider})
	data, _ := json.Marshal(root)
	catalog, err := economy.LoadCatalog(data)
	if err != nil {
		t.Fatal(err)
	}
	keys := map[string]struct{}{}
	for _, key := range copykeys.All() {
		keys[key] = struct{}{}
	}
	treeBytes, err := os.ReadFile(filepath.Join(repositoryRootForReputation, "balance/testdata/reputation-tree/fixture-v1.json"))
	if err != nil {
		t.Fatal(err)
	}
	bundle := suite.Bundle
	bundle.Economy = catalog
	if bundle.ReputationTree, err = reputation.LoadTree(treeBytes, reputation.Declarations{Economy: catalog, Curriculum: bundle.Curriculum, CopyKeys: keys}); err != nil {
		t.Fatal(err)
	}
	return bundle
}

type reputationCareerSeed struct {
	PolicyID          string   `json:"policy_id"`
	Seed              uint64   `json:"seed"`
	CareerPolicy      string   `json:"career_policy"`
	PurchasedNodeIDs  []string `json:"purchased_node_ids"`
	AppliedStarterIDs []string `json:"applied_starter_node_ids"`
	BonusFactor       string   `json:"bonus_factor"`
	TreatedGateMS     *int64   `json:"run_three_gate_ms"`
	ControlGateMS     *int64   `json:"control_run_three_gate_ms"`
	SavedMS           *int64   `json:"saved_ms"`
	Excluded          string   `json:"excluded"`
}

type reputationCareerReport struct {
	SchemaVersion int                    `json:"schema_version"`
	Threshold     string                 `json:"fixture_threshold"`
	Seeds         []reputationCareerSeed `json:"seeds"`
	GatedSeeds    int                    `json:"gated_seeds"`
	ExcludedSeeds int                    `json:"excluded_seeds"`
	GatePassed    bool                   `json:"h4_gate_passed"`
	Violations    []string               `json:"h4_violations"`
	SavedMS       map[string][3]int64    `json:"saved_ms_min_p50_max_by_policy"`
	Note          string                 `json:"note"`
}

// TestReputationCareerStartersShortenRunThree is H4: at every seed where run
// 3 starts with at least one starter node, the run-3 Garage gate is strictly
// sooner than the same seed's control career (no purchases). Reference and
// Casual buy the cheapest available node; Chaos draws uniformly (R10 H4).
// REPUTATION_UPDATE_CAREER=1 regenerates the pinned report.
func TestReputationCareerStartersShortenRunThree(t *testing.T) {
	requireReputationExhaustive(t)
	suite, err := LoadFirstHourSuite(repositoryRootForReputation, "balance/testdata/t0-t1/harness-scenario-v1.json", "balance/testdata/t0-t1/first-hour-policy-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	bundle := reputationCareerBundle(t, suite)
	experiment := FirstHourExperiment{AcquihirePurchasedMinimum: 200, BurnoutPriceFactor: "2e0", RouteKnowledgeBonus: 50, SeedCapital: "1e4", GeneratedBeigeTowers: 10}
	type job struct {
		spec RunSpec
		seed uint64
	}
	jobs := []job{}
	for _, spec := range suite.Scenario.Runs {
		start, err := strconv.ParseUint(spec.SeedStart, 10, 64)
		if err != nil {
			t.Fatal(err)
		}
		for offset := uint64(0); offset < uint64(spec.SeedCount); offset++ {
			jobs = append(jobs, job{spec, start + offset})
		}
	}
	results := make([]reputationCareerSeed, len(jobs))
	errs := make([]error, len(jobs))
	var group sync.WaitGroup
	limit := make(chan struct{}, 8)
	for index, current := range jobs {
		group.Add(1)
		go func(index int, current job) {
			defer group.Done()
			limit <- struct{}{}
			defer func() { <-limit }()
			policy := CareerCheapest
			if current.spec.PolicyID == "chaos.t0_t1" {
				policy = CareerSeededUniform
			}
			treated, err := suite.RunReputationCareer(current.spec, current.seed, experiment, ReputationCareerConfig{Bundle: bundle, Threshold: reputationCareerFixtureThreshold, Policy: policy})
			if err != nil {
				errs[index] = err
				return
			}
			control, err := suite.RunReputationCareer(current.spec, current.seed, experiment, ReputationCareerConfig{Bundle: bundle, Threshold: reputationCareerFixtureThreshold, Policy: CareerNone})
			if err != nil {
				errs[index] = err
				return
			}
			row := reputationCareerSeed{PolicyID: current.spec.PolicyID, Seed: current.seed, CareerPolicy: string(policy),
				PurchasedNodeIDs: treated.PurchasedNodeIDs, AppliedStarterIDs: treated.AppliedStarterIDs, BonusFactor: treated.BonusFactor,
				TreatedGateMS: treated.RunThreeGateMS, ControlGateMS: control.RunThreeGateMS}
			if row.TreatedGateMS != nil && row.ControlGateMS != nil {
				value := *row.ControlGateMS - *row.TreatedGateMS
				row.SavedMS = &value
			} else if row.TreatedGateMS == nil && row.ControlGateMS == nil {
				row.Excluded = "run3_gate_beyond_ratified_horizon_in_both_arms"
			}
			results[index] = row
		}(index, current)
	}
	group.Wait()
	for index, err := range errs {
		if err != nil {
			t.Fatalf("seed %d: %v", jobs[index].seed, err)
		}
	}
	report := reputationCareerReport{SchemaVersion: 1, Threshold: reputationCareerFixtureThreshold, Seeds: results, SavedMS: map[string][3]int64{},
		Note: "fixture-first H4: fixture threshold from the OD-2 measurement's satisfying set; nothing is ratified or minted"}
	violations, saved := evaluateReputationCareerGate(results, &report)
	report.Violations = violations
	report.GatePassed = len(report.Violations) == 0
	for _, violation := range report.Violations {
		t.Logf("H4 violation (recorded, not loosened): %s", violation)
	}
	for id, values := range saved {
		sort.Slice(values, func(left, right int) bool { return values[left] < values[right] })
		report.SavedMS[id] = [3]int64{values[0], values[len(values)/2], values[len(values)-1]}
	}
	if report.GatedSeeds == 0 {
		t.Fatal("no seed started run 3 with a starter node: the H4 gate would be vacuous")
	}
	encoded, _ := json.MarshalIndent(report, "", " ")
	encoded = append(encoded, '\n')
	path := filepath.Join(repositoryRootForReputation, reputationCareerReportPath)
	if os.Getenv("REPUTATION_UPDATE_CAREER") == "1" {
		if err := os.WriteFile(path, encoded, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	pinned, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(pinned, encoded) {
		t.Fatalf("career report drifted (regenerate with REPUTATION_UPDATE_CAREER=1): %v", err)
	}
}

// evaluateReputationCareerGate is the H4 gate: every seed whose run 3 starts
// with a starter must reach the run-3 Garage gate strictly sooner than its
// control. Excluded seeds are counted, never dropped.
func evaluateReputationCareerGate(results []reputationCareerSeed, report *reputationCareerReport) ([]string, map[string][]int64) {
	violations := []string{}
	saved := map[string][]int64{}
	for _, row := range results {
		if len(row.AppliedStarterIDs) == 0 {
			continue
		}
		if row.Excluded != "" {
			report.ExcludedSeeds++
			continue
		}
		report.GatedSeeds++
		switch {
		case row.TreatedGateMS == nil:
			violations = append(violations, fmt.Sprintf("%s seed %d: control reached the run-3 gate but the starter career did not", row.PolicyID, row.Seed))
		case row.ControlGateMS == nil:
			// Treated is inside the horizon and control beyond it: strictly sooner.
		case *row.TreatedGateMS >= *row.ControlGateMS:
			violations = append(violations, fmt.Sprintf("%s seed %d: run-3 gate %d is not sooner than control %d with starters %v", row.PolicyID, row.Seed, *row.TreatedGateMS, *row.ControlGateMS, row.AppliedStarterIDs))
		default:
			saved[row.PolicyID] = append(saved[row.PolicyID], *row.SavedMS)
		}
	}
	return violations, saved
}

func TestReputationCareerGateRejectsTiesAndMissingTreatment(t *testing.T) {
	gate := func(treated, control *int64) []string {
		violations, _ := evaluateReputationCareerGate([]reputationCareerSeed{{PolicyID: "p", AppliedStarterIDs: []string{"s"}, TreatedGateMS: treated, ControlGateMS: control,
			SavedMS: func() *int64 {
				if treated == nil || control == nil {
					return nil
				}
				v := *control - *treated
				return &v
			}()}}, &reputationCareerReport{})
		return violations
	}
	value := func(v int64) *int64 { return &v }
	if len(gate(value(10), value(10))) != 1 || len(gate(value(11), value(10))) != 1 || len(gate(nil, value(10))) != 1 {
		t.Fatal("H4 gate accepted a tie, a slower treatment, or a missing treatment")
	}
	if len(gate(value(9), value(10))) != 0 || len(gate(value(9), nil)) != 0 {
		t.Fatal("H4 gate rejected a strictly sooner treatment")
	}
}

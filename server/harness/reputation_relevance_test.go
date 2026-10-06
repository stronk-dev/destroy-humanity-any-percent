package harness

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"sync"
	"testing"

	"cloud-clicker/server/reputation"
)

const reputationRelevancePath = "planning/reputation-tree-v1/relevance-h5.v1.json"

type reputationNodeRelevance struct {
	Population    map[string]reputationRelevancePopulation `json:"population_by_policy"`
	NodeID        string                                   `json:"node_id"`
	Kind          string                                   `json:"kind"`
	DeltaMSP50    map[string]int64                         `json:"delta_ms_p50_by_policy"`
	PurchasedRuns map[string]int                           `json:"purchased_runs_by_policy"`
	Relevant      bool                                     `json:"relevant"`
	Excluded      string                                   `json:"excluded"`
}

type reputationRelevanceReport struct {
	Arms          []reputationRelevanceArm            `json:"arm_observations"`
	Sources       []ReputationCareerMeasurementSource `json:"measurement_sources"`
	SchemaVersion int                                 `json:"schema_version"`
	Threshold     string                              `json:"fixture_threshold"`
	EpsilonMS     int64                               `json:"epsilon_ms"`
	Nodes         []reputationNodeRelevance           `json:"nodes"`
	Note          string                              `json:"note"`
}

type reputationRelevanceArm struct {
	Source    ReputationCareerMeasurementSource `json:"measurement_source"`
	Gate      *int64                            `json:"run_three_gate_ms"`
	Purchased []string                          `json:"purchased_node_ids"`
}

type reputationRelevanceOutcome struct {
	gate      *int64
	purchased []string
	source    ReputationCareerMeasurementSource
}

func projectReputationRelevanceOutcome(suite *FirstHourSuite, spec RunSpec, seed uint64, experiment FirstHourExperiment,
	config ReputationCareerConfig, result ReputationCareerResult) (reputationRelevanceOutcome, error) {
	if err := validateReputationCareerResultSource(suite, spec, seed, experiment, config, result); err != nil {
		return reputationRelevanceOutcome{}, err
	}
	return reputationRelevanceOutcome{gate: result.RunThreeGateMS, purchased: result.PurchasedNodeIDs, source: result.MeasurementSource}, nil
}

func newReputationRelevanceReport(results []reputationRelevanceOutcome) reputationRelevanceReport {
	report := reputationRelevanceReport{SchemaVersion: 1, Threshold: reputationCareerFixtureThreshold, EpsilonMS: reputationRelevanceEpsilonMS,
		Note: "fixture-first H5: leave-one-out on the following run's Garage gate; delta_ms_p50_by_policy is conditional on the node being bought and both gate clocks being finite, not all purchased careers; unreached clocks are not imputed; the elective-Exit dimension needs a run-4 horizon (DESIGN-GAP RT-DG-F)"}
	for _, result := range results {
		report.Sources = append(report.Sources, result.source)
		arm := reputationRelevanceArm{Source: result.source, Purchased: slices.Clone(result.purchased)}
		if result.gate != nil {
			gate := *result.gate
			arm.Gate = &gate
		}
		report.Arms = append(report.Arms, arm)
	}
	return report
}

// reputationRelevanceEpsilonMS is H5's per-node epsilon on the following
// run's Garage gate. DESIGN-GAP RT-DG-F: R10 names an epsilon without a value;
// 1 ms (any strict improvement at p50) is the weakest non-vacuous choice and
// is recorded for the RFC author to replace.
const reputationRelevanceEpsilonMS = 1

// classifyReputationRelevance is H5's rule: relevant when some persona's
// p50 delta reaches epsilon; otherwise excluded with a reason, and an
// exclusion without a reason is an invalid report.
func classifyReputationRelevance(row *reputationNodeRelevance) {
	for _, delta := range row.DeltaMSP50 {
		if delta >= reputationRelevanceEpsilonMS {
			row.Relevant = true
		}
	}
	if row.Relevant {
		return
	}
	purchased := 0
	for _, count := range row.PurchasedRuns {
		purchased += count
	}
	switch {
	case purchased == 0:
		row.Excluded = "unreachable_in_horizon"
	case row.Kind == reputation.KindBonusUnlock:
		row.Excluded = "owner_exempt:OD-5"
	default:
		row.Excluded = ""
	}
}

// TestReputationTreeRelevance is H5: per node, leave-one-out Δ on the
// following run's Garage gate across the career personas. Every node is
// either relevant or excluded with a reason; a starter node that is bought
// yet moves nothing fails. REPUTATION_UPDATE_RELEVANCE=1 regenerates.
func TestReputationTreeRelevance(t *testing.T) {
	requireReputationExhaustive(t)
	report := measureReputationRelevanceReport(t)
	encoded, _ := json.MarshalIndent(report, "", " ")
	encoded = append(encoded, '\n')
	path := filepath.Join(repositoryRootForReputation, reputationRelevancePath)
	if os.Getenv("REPUTATION_UPDATE_RELEVANCE") == "1" {
		if err := os.WriteFile(path, encoded, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	pinned, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(pinned, encoded) {
		t.Fatalf("relevance report drifted (regenerate with REPUTATION_UPDATE_RELEVANCE=1): %v", err)
	}
}

func measureReputationRelevanceReport(t *testing.T) reputationRelevanceReport {
	t.Helper()
	suite, err := LoadFirstHourSuite(repositoryRootForReputation, "balance/testdata/t0-t1/harness-scenario-v1.json", "balance/testdata/t0-t1/first-hour-policy-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	bundle := reputationCareerBundle(t, suite)
	experiment := FirstHourExperiment{AcquihirePurchasedMinimum: 200, BurnoutPriceFactor: "2e0", RouteKnowledgeBonus: 50, SeedCapital: "1e4", GeneratedBeigeTowers: 10}
	type job struct {
		spec    RunSpec
		seed    uint64
		exclude string
	}
	nodes := bundle.ReputationTree.Nodes()
	jobs := []job{}
	for _, spec := range suite.Scenario.Runs {
		start, err := strconv.ParseUint(spec.SeedStart, 10, 64)
		if err != nil {
			t.Fatal(err)
		}
		for offset := uint64(0); offset < uint64(spec.SeedCount); offset++ {
			jobs = append(jobs, job{spec, start + offset, ""})
			for _, node := range nodes {
				jobs = append(jobs, job{spec, start + offset, node.NodeID})
			}
		}
	}
	results := make([]reputationRelevanceOutcome, len(jobs))
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
			config := ReputationCareerConfig{Bundle: bundle, Threshold: reputationCareerFixtureThreshold, Policy: policy, Exclude: current.exclude}
			result, err := suite.RunReputationCareer(current.spec, current.seed, experiment, config)
			if err != nil {
				errs[index] = err
				return
			}
			results[index], errs[index] = projectReputationRelevanceOutcome(suite, current.spec, current.seed, experiment, config, result)
		}(index, current)
	}
	group.Wait()
	for index, err := range errs {
		if err != nil {
			t.Fatalf("seed %d exclude %q: %v", jobs[index].seed, jobs[index].exclude, err)
		}
	}
	report, err := composeReputationRelevanceReport(nodes, results)
	if err != nil {
		t.Fatal(err)
	}
	if err := validateReputationRelevanceRecomposition(nodes, report); err != nil {
		t.Fatal(err)
	}
	t.Logf("H5 raw observations retained and recomposed: %d", len(report.Arms))
	t.Logf("H5 source admission complete: arms=%d retained_sources=%d", len(results), len(report.Sources))
	for _, row := range report.Nodes {
		populationJSON, err := json.Marshal(row.Population)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("H5 current bought/not-bought/finite populations: node=%s population=%s conditional_p50=%v", row.NodeID, populationJSON, row.DeltaMSP50)
		if !row.Relevant && row.Excluded == "" {
			t.Errorf("node %s is bought but moves nothing and has no exclusion reason: %+v", row.NodeID, row)
		}
	}
	return report
}

// Recomposition checks internal evidence consistency, not earned-player or
// producer authenticity. A dated artifact additionally needs exact cohort and
// committed software/input identity plus executed full replay.
func composeReputationRelevanceReport(nodes []reputation.Node, results []reputationRelevanceOutcome) (reputationRelevanceReport, error) {
	known := map[string]bool{}
	for _, node := range nodes {
		if node.NodeID == "" || known[node.NodeID] {
			return reputationRelevanceReport{}, fmt.Errorf("invalid H5 node inventory")
		}
		known[node.NodeID] = true
	}
	baseline := map[RunKey]reputationRelevanceOutcome{}
	type armKey struct {
		key  RunKey
		mask string
	}
	seen := map[armKey]bool{}
	for _, result := range results {
		key := armKey{result.source.RunKey, result.source.ExcludedNodeID}
		if seen[key] || len(nodes) == 0 || key.key.PolicyID == "" || key.key.Seed == "" ||
			key.mask != "" && !known[key.mask] || result.gate != nil && *result.gate < 0 {
			return reputationRelevanceReport{}, fmt.Errorf("invalid or duplicate H5 arm")
		}
		seen[key] = true
		purchased := map[string]bool{}
		for _, id := range result.purchased {
			if !known[id] || purchased[id] || id == key.mask {
				return reputationRelevanceReport{}, fmt.Errorf("invalid H5 purchased-node observation")
			}
			purchased[id] = true
		}
		if key.mask == "" {
			baseline[key.key] = result
		}
	}
	if len(baseline) == 0 || len(results) != len(baseline)*(len(nodes)+1) {
		return reputationRelevanceReport{}, fmt.Errorf("incomplete H5 baseline/mask population")
	}
	for _, result := range results {
		base, ok := baseline[result.source.RunKey]
		expectedSource := base.source
		expectedSource.ExcludedNodeID = result.source.ExcludedNodeID
		if !ok || expectedSource != result.source {
			return reputationRelevanceReport{}, fmt.Errorf("H5 masked arm lacks its matching baseline source")
		}
	}
	report := newReputationRelevanceReport(results)
	for _, node := range nodes {
		row := reputationNodeRelevance{NodeID: node.NodeID, Kind: node.Kind, DeltaMSP50: map[string]int64{}, PurchasedRuns: map[string]int{},
			Population: map[string]reputationRelevancePopulation{}}
		deltas := map[string][]int64{}
		for _, result := range results {
			if result.source.ExcludedNodeID != node.NodeID {
				continue
			}
			base := baseline[result.source.RunKey]
			bought := slices.Contains(base.purchased, node.NodeID)
			observeReputationRelevancePair(&row, deltas, result.source.RunKey.PolicyID, bought, base, result)
		}
		for id, values := range deltas {
			sort.Slice(values, func(left, right int) bool { return values[left] < values[right] })
			row.DeltaMSP50[id] = values[len(values)/2]
		}
		classifyReputationRelevance(&row)
		report.Nodes = append(report.Nodes, row)
	}
	return report, nil
}

func validateReputationRelevanceRecomposition(nodes []reputation.Node, report reputationRelevanceReport) error {
	results := make([]reputationRelevanceOutcome, len(report.Arms))
	for index, arm := range report.Arms {
		results[index] = reputationRelevanceOutcome{source: arm.Source, gate: arm.Gate, purchased: arm.Purchased}
	}
	rebuilt, err := composeReputationRelevanceReport(nodes, results)
	if err != nil {
		return err
	}
	for _, node := range rebuilt.Nodes {
		if !node.Relevant && node.Excluded == "" {
			return fmt.Errorf("H5 bought node %s has no qualifying effect or exclusion reason", node.NodeID)
		}
	}
	expected, err := CanonicalJSON(rebuilt)
	if err != nil {
		return err
	}
	actual, err := CanonicalJSON(report)
	if err != nil || !bytes.Equal(expected, actual) {
		return fmt.Errorf("H5 retained report differs from raw-arm recomposition: %v", err)
	}
	return nil
}

func TestReputationRelevanceClassification(t *testing.T) {
	cases := []struct {
		row  reputationNodeRelevance
		want string
		rel  bool
	}{
		{reputationNodeRelevance{Kind: "starter", DeltaMSP50: map[string]int64{"a": 5}, PurchasedRuns: map[string]int{"a": 3}}, "", true},
		{reputationNodeRelevance{Kind: "starter", DeltaMSP50: map[string]int64{}, PurchasedRuns: map[string]int{}}, "unreachable_in_horizon", false},
		{reputationNodeRelevance{Kind: "bonus_unlock", DeltaMSP50: map[string]int64{"a": 0}, PurchasedRuns: map[string]int{"a": 3}}, "owner_exempt:OD-5", false},
		{reputationNodeRelevance{Kind: "starter", DeltaMSP50: map[string]int64{"a": 0}, PurchasedRuns: map[string]int{"a": 3}}, "", false},
	}
	for index, test := range cases {
		row := test.row
		classifyReputationRelevance(&row)
		if row.Relevant != test.rel || row.Excluded != test.want {
			t.Fatalf("case %d: relevant=%t excluded=%q", index, row.Relevant, row.Excluded)
		}
	}
}

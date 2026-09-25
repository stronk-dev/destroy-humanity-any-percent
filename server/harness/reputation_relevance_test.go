package harness

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"sync"
	"testing"

	"cloud-clicker/server/reputation"
)

const reputationRelevancePath = "planning/reputation-tree-v1/relevance-h5.v1.json"

type reputationNodeRelevance struct {
	NodeID        string           `json:"node_id"`
	Kind          string           `json:"kind"`
	DeltaMSP50    map[string]int64 `json:"delta_ms_p50_by_policy"`
	PurchasedRuns map[string]int   `json:"purchased_runs_by_policy"`
	Relevant      bool             `json:"relevant"`
	Excluded      string           `json:"excluded"`
}

type reputationRelevanceReport struct {
	SchemaVersion int                       `json:"schema_version"`
	Threshold     string                    `json:"fixture_threshold"`
	EpsilonMS     int64                     `json:"epsilon_ms"`
	Nodes         []reputationNodeRelevance `json:"nodes"`
	Note          string                    `json:"note"`
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
	type outcome struct {
		gate      *int64
		purchased []string
	}
	results := make([]outcome, len(jobs))
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
			result, err := suite.RunReputationCareer(current.spec, current.seed, experiment, ReputationCareerConfig{Bundle: bundle,
				Threshold: reputationCareerFixtureThreshold, Policy: policy, Exclude: current.exclude})
			results[index], errs[index] = outcome{result.RunThreeGateMS, result.PurchasedNodeIDs}, err
		}(index, current)
	}
	group.Wait()
	for index, err := range errs {
		if err != nil {
			t.Fatalf("seed %d exclude %q: %v", jobs[index].seed, jobs[index].exclude, err)
		}
	}
	baseline := map[string]outcome{}
	for index, current := range jobs {
		if current.exclude == "" {
			baseline[current.spec.PolicyID+"/"+strconv.FormatUint(current.seed, 10)] = results[index]
		}
	}
	report := reputationRelevanceReport{SchemaVersion: 1, Threshold: reputationCareerFixtureThreshold, EpsilonMS: reputationRelevanceEpsilonMS,
		Note: "fixture-first H5: leave-one-out on the following run's Garage gate; the elective-Exit dimension needs a run-4 horizon (DESIGN-GAP RT-DG-F)"}
	for _, node := range nodes {
		row := reputationNodeRelevance{NodeID: node.NodeID, Kind: node.Kind, DeltaMSP50: map[string]int64{}, PurchasedRuns: map[string]int{}}
		deltas := map[string][]int64{}
		for index, current := range jobs {
			if current.exclude != node.NodeID {
				continue
			}
			base := baseline[current.spec.PolicyID+"/"+strconv.FormatUint(current.seed, 10)]
			bought := false
			for _, id := range base.purchased {
				bought = bought || id == node.NodeID
			}
			if !bought {
				continue
			}
			row.PurchasedRuns[current.spec.PolicyID]++
			if base.gate != nil && results[index].gate != nil {
				deltas[current.spec.PolicyID] = append(deltas[current.spec.PolicyID], *results[index].gate-*base.gate)
			}
		}
		for id, values := range deltas {
			sort.Slice(values, func(left, right int) bool { return values[left] < values[right] })
			row.DeltaMSP50[id] = values[len(values)/2]
		}
		classifyReputationRelevance(&row)
		if !row.Relevant && row.Excluded == "" {
			t.Errorf("node %s is bought but moves nothing and has no exclusion reason: %+v", node.NodeID, row)
		}
		report.Nodes = append(report.Nodes, row)
	}
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

package harness

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"testing"
)

// Counts describe observed clocks, not an imputation or a censoring rule.
// For H5, treated means the unmasked baseline and control the masked career.
type reputationGatePairPopulation struct {
	Pairs         int `json:"pairs"`
	FinitePairs   int `json:"finite_pairs"`
	Faster        int `json:"treated_faster_finite_pairs"`
	Tied          int `json:"tied_finite_pairs"`
	Slower        int `json:"treated_slower_finite_pairs"`
	TreatedOnly   int `json:"treated_only_reached"`
	ControlOnly   int `json:"control_only_reached"`
	BothUnreached int `json:"both_unreached"`
}

func (population *reputationGatePairPopulation) observe(treated, control *int64) {
	population.Pairs++
	switch {
	case treated == nil && control == nil:
		population.BothUnreached++
	case control == nil:
		population.TreatedOnly++
	case treated == nil:
		population.ControlOnly++
	default:
		population.FinitePairs++
		switch {
		case *treated < *control:
			population.Faster++
		case *treated == *control:
			population.Tied++
		default:
			population.Slower++
		}
	}
}

type reputationCareerPopulation struct {
	Rows         int                          `json:"rows"`
	NoStarter    int                          `json:"no_starter_rows"`
	StarterPairs reputationGatePairPopulation `json:"starter_gate_pairs"`
}

// Finite savings are descriptive package comparisons, not per-node effects.
// The strict acceptance oracle remains evaluateReputationCareerGate.
func observeReputationCareerPopulation(rows []reputationCareerSeed) (map[string]reputationCareerPopulation, map[string][3]int64, error) {
	populations := map[string]reputationCareerPopulation{}
	finiteSavings := map[string][]int64{}
	for _, row := range rows {
		if err := validateReputationCareerGateRow(row); err != nil {
			return nil, nil, fmt.Errorf("invalid H4 population row %s/%d: %w", row.PolicyID, row.Seed, err)
		}
		population := populations[row.PolicyID]
		population.Rows++
		if len(row.AppliedStarterIDs) == 0 {
			population.NoStarter++
		} else {
			population.StarterPairs.observe(row.TreatedGateMS, row.ControlGateMS)
			if row.SavedMS != nil {
				finiteSavings[row.PolicyID] = append(finiteSavings[row.PolicyID], *row.SavedMS)
			}
		}
		populations[row.PolicyID] = population
	}
	summaries := map[string][3]int64{}
	for policy, savings := range finiteSavings {
		sort.Slice(savings, func(i, j int) bool { return savings[i] < savings[j] })
		summaries[policy] = [3]int64{savings[0], savings[len(savings)/2], savings[len(savings)-1]}
	}
	return populations, summaries, nil
}

type reputationRelevancePopulation struct {
	Careers      int                          `json:"baseline_careers"`
	NotPurchased int                          `json:"not_purchased_careers"`
	BoughtPairs  reputationGatePairPopulation `json:"purchased_gate_pairs"`
}

// Retain the original H5 estimator/classifier inputs; add its actual denominator.
func observeReputationRelevancePair(row *reputationNodeRelevance, deltas map[string][]int64,
	policy string, bought bool, baseline, masked reputationRelevanceOutcome) {
	population := row.Population[policy]
	population.Careers++
	if !bought {
		population.NotPurchased++
	} else {
		row.PurchasedRuns[policy]++
		population.BoughtPairs.observe(baseline.gate, masked.gate)
		if baseline.gate != nil && masked.gate != nil {
			deltas[policy] = append(deltas[policy], *masked.gate-*baseline.gate)
		}
	}
	row.Population[policy] = population
}

func TestReputationReportPopulation(t *testing.T) {
	clock := func(value int64) *int64 { return &value }
	rows := []reputationCareerSeed{}
	for index, pair := range [][2]*int64{{clock(8), clock(10)}, {clock(10), clock(10)}, {clock(11), clock(10)},
		{clock(8), nil}, {nil, clock(10)}, {nil, nil}, {clock(1), clock(10)}} {
		row := reputationCareerSeed{PolicyID: "p", Seed: uint64(index), AppliedStarterIDs: []string{"package"},
			TreatedGateMS: pair[0], ControlGateMS: pair[1]}
		if pair[0] != nil && pair[1] != nil {
			saving := *pair[1] - *pair[0]
			row.SavedMS = &saving
		} else if pair[0] == nil && pair[1] == nil {
			row.Excluded = reputationBothArmsBeyondHorizon
		}
		if index == 6 {
			row.AppliedStarterIDs = nil
		}
		rows = append(rows, row)
	}
	populations, finite, err := observeReputationCareerPopulation(rows)
	wantPair := reputationGatePairPopulation{Pairs: 6, FinitePairs: 3, Faster: 1, Tied: 1, Slower: 1,
		TreatedOnly: 1, ControlOnly: 1, BothUnreached: 1}
	if err != nil || !reflect.DeepEqual(populations, map[string]reputationCareerPopulation{
		"p": {Rows: 7, NoStarter: 1, StarterPairs: wantPair}}) ||
		!reflect.DeepEqual(finite, map[string][3]int64{"p": {-1, 0, 2}}) {
		t.Fatalf("H4 concealed a population or omitted finite failures: populations=%+v finite=%v err=%v", populations, finite, err)
	}
	report, err := newReputationCareerReport(rows)
	if err != nil || !reflect.DeepEqual(report.Population, populations) || !reflect.DeepEqual(report.FiniteSavedMS, finite) {
		t.Fatalf("H4 report dropped observed populations/savings: %+v / %v", report, err)
	}
	encoded, err := json.Marshal(report)
	var wire struct {
		Population map[string]reputationCareerPopulation `json:"population_by_policy"`
		Finite     map[string][3]int64                   `json:"finite_starter_pair_saved_ms_min_p50_max_by_policy"`
	}
	if err != nil || json.Unmarshal(encoded, &wire) != nil || !reflect.DeepEqual(wire.Population, populations) || !reflect.DeepEqual(wire.Finite, finite) {
		t.Fatal("H4 serialized report lost observed populations/savings")
	}
	legacy := reputationCareerReport{}
	violations, passing := evaluateReputationCareerGate(rows, &legacy)
	if len(violations) != 3 || legacy.GatedSeeds != 5 || legacy.ExcludedSeeds != 1 ||
		!reflect.DeepEqual(passing, map[string][]int64{"p": {2}}) {
		t.Fatalf("observation changed strict H4 acceptance: violations=%v passing=%v report=%+v", violations, passing, legacy)
	}
	bad := rows[0]
	bad.SavedMS = clock(999)
	if _, _, err := observeReputationCareerPopulation([]reputationCareerSeed{bad}); err == nil {
		t.Fatal("H4 census admitted inconsistent clocks/savings")
	}

	row := reputationNodeRelevance{Kind: "starter", DeltaMSP50: map[string]int64{}, PurchasedRuns: map[string]int{},
		Population: map[string]reputationRelevancePopulation{}}
	deltas := map[string][]int64{}
	// One finite positive comparison, three bought censored pairs, one not bought.
	for index, pair := range [][2]*int64{{clock(10), clock(20)}, {clock(10), nil}, {nil, clock(20)}, {nil, nil}, {clock(1), clock(99)}} {
		observeReputationRelevancePair(&row, deltas, "p", index != 4,
			reputationRelevanceOutcome{gate: pair[0]}, reputationRelevanceOutcome{gate: pair[1]})
	}
	observeReputationRelevancePair(&row, deltas, "never-bought", false, reputationRelevanceOutcome{}, reputationRelevanceOutcome{})
	for _, pair := range [][2]*int64{{clock(10), clock(20)}, {clock(10), clock(10)}, {clock(11), clock(10)}} {
		observeReputationRelevancePair(&row, deltas, "finite", true, reputationRelevanceOutcome{gate: pair[0]}, reputationRelevanceOutcome{gate: pair[1]})
	}
	if !reflect.DeepEqual(row.Population, map[string]reputationRelevancePopulation{
		"p": {Careers: 5, NotPurchased: 1, BoughtPairs: reputationGatePairPopulation{Pairs: 4, FinitePairs: 1,
			Faster: 1, TreatedOnly: 1, ControlOnly: 1, BothUnreached: 1}},
		"never-bought": {Careers: 1, NotPurchased: 1},
		"finite":       {Careers: 3, BoughtPairs: reputationGatePairPopulation{Pairs: 3, FinitePairs: 3, Faster: 1, Tied: 1, Slower: 1}}}) ||
		!reflect.DeepEqual(row.PurchasedRuns, map[string]int{"p": 4, "finite": 3}) ||
		!reflect.DeepEqual(deltas, map[string][]int64{"p": {10}, "finite": {10, 0, -1}}) {
		t.Fatalf("H5 concealed purchased/censored/zero-finite populations: row=%+v deltas=%v", row, deltas)
	}
	encoded, err = json.Marshal(row)
	var nodeWire struct {
		Population map[string]reputationRelevancePopulation `json:"population_by_policy"`
	}
	if err != nil || json.Unmarshal(encoded, &nodeWire) != nil || !reflect.DeepEqual(nodeWire.Population, row.Population) {
		t.Fatal("H5 serialized node report lost observed populations")
	}
	row.DeltaMSP50["p"] = deltas["p"][0]
	classifyReputationRelevance(&row)
	if !row.Relevant || row.Excluded != "" {
		t.Fatal("observation changed existing H5 conditional-median classification")
	}
	// Exact census expectations above detect omission; also spell out conservation.
	for policy, population := range populations {
		if population.Rows != population.NoStarter+population.StarterPairs.Pairs {
			t.Fatalf("H4 row conservation failed for %s", policy)
		}
	}
	for policy, population := range row.Population {
		if population.Careers != population.NotPurchased+population.BoughtPairs.Pairs ||
			population.BoughtPairs.Pairs != row.PurchasedRuns[policy] {
			t.Fatalf("H5 career conservation failed for %s", policy)
		}
	}
}

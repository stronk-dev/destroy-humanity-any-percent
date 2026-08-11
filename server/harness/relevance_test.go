package harness

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"cloud-clicker/server/decimal"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/epochseed"
	"cloud-clicker/server/production"
)

func TestRelevanceRunBudgetAndReachedEncoding(t *testing.T) {
	budget, err := ComputeRelevanceRunBudget(2, 1, 3, 4, true)
	if err != nil || budget != 28 {
		t.Fatalf("budget=%d err=%v", budget, err)
	}
	baseline, ablated := int64(100), int64(175)
	both := MakeRelevanceDelta("milestone.test", &baseline, &ablated, 500)
	unreached := MakeRelevanceDelta("milestone.test", &baseline, nil, 500)
	if both.Status != "both_reached" || both.DeltaMS == nil || *both.DeltaMS != 75 ||
		unreached.Status != "ablated_unreached" || unreached.DeltaMS == nil || *unreached.DeltaMS != 400 {
		t.Fatalf("both=%+v unreached=%+v", both, unreached)
	}
	if value, err := ceilDecimalRatio(decimal.New(11, 0), decimal.New(2, 0)); err != nil || value != 6 {
		t.Fatalf("ceil ratio=%d err=%v", value, err)
	}
	if value, err := ceilDecimalRatio(decimal.Zero, decimal.New(2, 0)); err != nil || value != 0 {
		t.Fatalf("zero ratio=%d err=%v", value, err)
	}
}

func TestRelevanceFixtureRunsDeterministicallyThroughProduction(t *testing.T) {
	suite, err := LoadRelevanceSuite("../..", "testdata/harness/relevance/scenario-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	first, err := suite.RunRelevance()
	if err != nil {
		t.Fatal(err)
	}
	second, err := suite.RunRelevance()
	if err != nil {
		t.Fatal(err)
	}
	firstBytes, _ := CanonicalJSON(first)
	secondBytes, _ := CanonicalJSON(second)
	if !reflect.DeepEqual(firstBytes, secondBytes) {
		t.Fatal("relevance report is not byte deterministic")
	}
	if first.RunBudget.DeclaredRuns != 14 || first.RunBudget.ExecutedRuns != 14 ||
		first.RunBudget.DeclaredTransitions != first.RunBudget.ExecutedTransitions || len(first.Items) != 4 || len(first.Groups) != 4 {
		t.Fatalf("report cardinality=%+v items=%d groups=%d", first.RunBudget, len(first.Items), len(first.Groups))
	}
	if first.GreedyOracle == nil || !first.GreedyOracle.Passed {
		t.Fatalf("greedy oracle=%+v", first.GreedyOracle)
	}
	joined := strings.Join(first.Failures, ",")
	for _, expected := range []string{"relevance_floor:upgrade.dead", "role_floor:generator.alpha", "trap_floor:generator.alpha"} {
		if !strings.Contains(joined, expected) {
			t.Fatalf("fixture did not discriminate %q: %v", expected, first.Failures)
		}
	}
	var supported bool
	for _, item := range first.Items {
		if item.PurchasableID == "upgrade.alpha" && item.Support == "group_supported" && item.RelevancePassed {
			supported = true
		}
	}
	if !supported {
		t.Fatalf("fixture did not demonstrate group-supported substitute: %+v", first.Items)
	}
	if err := ValidateRelevanceReport(first); err != nil {
		t.Fatal(err)
	}
	invalid := first
	invalid.Groups = nil
	if err := ValidateRelevanceReport(invalid); err == nil {
		t.Fatal("report accepted a missing row family")
	}
	invalid = first
	invalid.Items = append([]RelevanceItemReport(nil), first.Items...)
	invalid.Items[0], invalid.Items[1] = invalid.Items[1], invalid.Items[0]
	if err := ValidateRelevanceReport(invalid); err == nil {
		t.Fatal("report accepted unsorted item rows")
	}
	invalid = first
	invalid.Items = append([]RelevanceItemReport(nil), first.Items...)
	invalid.Items[0].IndividualDeltas = append([]RelevanceDelta(nil), first.Items[0].IndividualDeltas...)
	invalid.Items[0].IndividualDeltas[0].Status = "both_reached"
	invalid.Items[0].IndividualDeltas[0].AblatedMS = nil
	if err := ValidateRelevanceReport(invalid); err == nil {
		t.Fatal("report accepted an illegal delta null union")
	}
	invalid = first
	invalid.TierContributions = append([]RelevanceTierContribution(nil), first.TierContributions...)
	invalid.TierContributions[0], invalid.TierContributions[1] = invalid.TierContributions[1], invalid.TierContributions[0]
	if err := ValidateRelevanceReport(invalid); err == nil {
		t.Fatal("report accepted unsorted tier contributions")
	}
	invalid = first
	invalid.RoleActivations = append([]RoleActivationCount(nil), first.RoleActivations...)
	invalid.RoleActivations = append(invalid.RoleActivations, invalid.RoleActivations[0])
	if err := ValidateRelevanceReport(invalid); err == nil {
		t.Fatal("report accepted duplicate role activations")
	}
	invalid = first
	invalid.GreedyOracle = &RelevanceGreedyOracle{MilestoneID: first.GreedyOracle.MilestoneID,
		GreedyMS: first.GreedyOracle.GreedyMS, BeamMS: first.GreedyOracle.BeamMS, GapPPM: first.GreedyOracle.GapPPM + 1,
		MaximumPPM: first.GreedyOracle.MaximumPPM, Passed: first.GreedyOracle.Passed}
	if err := ValidateRelevanceReport(invalid); err == nil {
		t.Fatal("report accepted a non-reconciling greedy gap")
	}
	invalid = first
	invalid.Items = append([]RelevanceItemReport(nil), first.Items...)
	invalid.Items[0].IndividualDeltas = append([]RelevanceDelta(nil), first.Items[0].IndividualDeltas...)
	badDelta := *invalid.Items[0].IndividualDeltas[0].DeltaMS + 1
	invalid.Items[0].IndividualDeltas[0].DeltaMS = &badDelta
	if err := ValidateRelevanceReport(invalid); err == nil {
		t.Fatal("report accepted a non-reconciling delta")
	}
}

func TestRelevanceFailsBeforeDispatchWhenRunBudgetIsTooSmall(t *testing.T) {
	suite, err := LoadRelevanceSuite("../..", "testdata/harness/relevance/scenario-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	suite.Scenario.RelevanceBudgetMaxRuns = 11
	if _, err := suite.RunRelevance(); err == nil || !strings.Contains(err.Error(), "run budget") {
		t.Fatalf("budget error=%v", err)
	}
}

func TestRelevanceRuntimeTransitionBudgetAbortsAtActualWork(t *testing.T) {
	suite, err := LoadRelevanceSuite("../..", "testdata/harness/relevance/scenario-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	suite.Scenario.RelevanceBudgetMaxTransitions = 1
	if _, err := suite.RunRelevance(); err == nil || !strings.Contains(err.Error(), "executed 1, maximum 1") {
		t.Fatalf("runtime transition budget error=%v", err)
	}
}

func TestT0T1ReferenceBootstrapMakesNonEmptyPurchaseThroughPinnedManualClamp(t *testing.T) {
	suite, err := LoadRelevanceSuite("../..", "balance/testdata/t0-t1/relevance-scenario-v2.json")
	if err != nil {
		t.Fatal(err)
	}
	state, err := suite.newRelevanceState()
	if err != nil {
		t.Fatal(err)
	}
	counter := &relevanceCounter{limit: suite.Scenario.RelevanceBudgetMaxTransitions}
	manual, err := suite.applyReferenceManual(state, 1, 0, production.AblationMask{}, counter)
	if err != nil {
		t.Fatal(err)
	}
	cash, _ := state.Ledger.Balance("company.cash")
	if manual.Applied != 10 || !cash.Eq(decimal.New(1, 1)) || state.ManualTokenMilli != 40000 || counter.value != 1 {
		t.Fatalf("manual bootstrap=%+v cash=%s tokens=%d transitions=%d", manual, cash, state.ManualTokenMilli, counter.value)
	}

	counter = &relevanceCounter{limit: suite.Scenario.RelevanceBudgetMaxTransitions}
	reference, err := suite.runReference(production.AblationMask{}, counter)
	if err != nil {
		t.Fatal(err)
	}
	if len(reference.Purchases) == 0 || counter.value > suite.Scenario.RelevanceBudgetMaxTransitions {
		t.Fatalf("reference purchases=%v transitions=%d", reference.Purchases, counter.value)
	}
}

func TestT0T1BeamSelectsCheapCandidatesBeforeSimulation(t *testing.T) {
	suite, err := LoadRelevanceSuite("../..", "balance/testdata/t0-t1/relevance-scenario-v2.json")
	if err != nil {
		t.Fatal(err)
	}
	state, err := suite.newRelevanceState()
	if err != nil {
		t.Fatal(err)
	}
	ids, err := suite.beamCandidateIDs(state)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"generator.beige_tower"}
	if !reflect.DeepEqual(ids, want) {
		t.Fatalf("cheap beam candidates=%v want=%v", ids, want)
	}
	if state.EvaluatedThrough != Epoch || state.GeneratorPurchasedTotal != 0 {
		t.Fatalf("cheap selection simulated or mutated state: evaluated=%s purchases=%d", state.EvaluatedThrough, state.GeneratorPurchasedTotal)
	}
}

func TestBeamChildBoundReducesWorkBeforeRollout(t *testing.T) {
	load := func(children int64) (*RelevanceSuite, *relevanceCounter) {
		t.Helper()
		suite, err := LoadRelevanceSuite("../..", "testdata/harness/relevance/scenario-v1.json")
		if err != nil {
			t.Fatal(err)
		}
		suite.Scenario.BeamChildren = children
		return suite, &relevanceCounter{limit: suite.Scenario.RelevanceBudgetMaxTransitions}
	}

	bounded, boundedCounter := load(1)
	if _, err := bounded.runBeam(boundedCounter); err != nil {
		t.Fatal(err)
	}
	exhaustive, exhaustiveCounter := load(4)
	if _, err := exhaustive.runBeam(exhaustiveCounter); err != nil {
		t.Fatal(err)
	}
	if boundedCounter.value >= exhaustiveCounter.value {
		t.Fatalf("pre-rollout child bound did not reduce work: bounded=%d exhaustive=%d", boundedCounter.value, exhaustiveCounter.value)
	}
}

func TestRelevanceWindowsBindEveryItemToAnInWindowMilestone(t *testing.T) {
	suite, err := LoadRelevanceSuite("../..", "testdata/harness/relevance/scenario-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := validateRelevanceWindows(suite.Scenario, suite.Policy, suite.Routes); err != nil {
		t.Fatal(err)
	}
	invalid := suite.Scenario
	invalid.Segments = append([]RelevanceSegment(nil), suite.Scenario.Segments...)
	fromGate := "gate.t1_to_t2"
	invalid.Segments[0].FromGate = &fromGate
	if err := validateRelevanceWindows(invalid, suite.Policy, suite.Routes); err == nil || !strings.Contains(err.Error(), "no milestone") {
		t.Fatalf("out-of-window milestone accepted: %v", err)
	}
}

func TestRelevanceV2PreservesRunGenesisWindowInReport(t *testing.T) {
	suite, err := LoadRelevanceSuite("../..", "testdata/harness/relevance/scenario-v2.json")
	if err != nil {
		t.Fatal(err)
	}
	report, err := suite.RunRelevance()
	if err != nil {
		t.Fatal(err)
	}
	if report.SchemaVersion != 2 || len(report.Items) == 0 || report.Items[0].AvailabilityWindow.FromGate != nil {
		t.Fatalf("v2 report=%+v", report)
	}
	encoded, err := json.Marshal(report.Items[0].AvailabilityWindow)
	if err != nil || !bytes.Equal(encoded, []byte(`{"from_gate":null,"to_gate":"gate.t0_to_t1"}`)) {
		t.Fatalf("v2 report did not preserve null window: %s err=%v", encoded, err)
	}
}

func TestRelevanceV2BindsMultipleOrderedSegmentsToOneMilestone(t *testing.T) {
	suite, err := LoadRelevanceSuite("../..", "testdata/harness/relevance/scenario-v2-multisegment.json")
	if err != nil {
		t.Fatal(err)
	}
	if len(suite.Scenario.Segments) != 2 || suite.Scenario.Segments[0].MilestoneID != suite.Scenario.Milestone.ID ||
		suite.Scenario.Segments[1].MilestoneID != suite.Scenario.Milestone.ID {
		t.Fatalf("multi-segment scenario=%+v", suite.Scenario.Segments)
	}
	report, err := suite.RunRelevance()
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range report.Items {
		if len(item.IndividualDeltas) != 1 || item.IndividualDeltas[0].MilestoneID != suite.Scenario.Milestone.ID {
			t.Fatalf("item %s changed single-milestone report bytes: %+v", item.PurchasableID, item.IndividualDeltas)
		}
	}
}

func TestRelevanceSegmentValidationRejectsRuledV2Matrix(t *testing.T) {
	suite, err := LoadRelevanceSuite("../..", "testdata/harness/relevance/scenario-v2-multisegment.json")
	if err != nil {
		t.Fatal(err)
	}
	clone := func() RelevanceScenario {
		result := suite.Scenario
		result.Segments = append([]RelevanceSegment(nil), suite.Scenario.Segments...)
		return result
	}
	t0, t1 := "gate.t0_to_t1", "gate.t1_to_t2"
	unknown := "gate.unknown"
	tests := []struct {
		name string
		edit func(*RelevanceScenario)
		want string
	}{
		{name: "v1_multiple", edit: func(value *RelevanceScenario) { value.SchemaVersion = 1 }, want: "completely bind"},
		{name: "unordered", edit: func(value *RelevanceScenario) {
			value.Segments[0], value.Segments[1] = value.Segments[1], value.Segments[0]
		}, want: "ordered"},
		{name: "overlapping", edit: func(value *RelevanceScenario) { value.Segments[1].FromGate = nil }, want: "ordered"},
		{name: "duplicate", edit: func(value *RelevanceScenario) { value.Segments[1] = value.Segments[0] }, want: "duplicate"},
		{name: "unknown_from", edit: func(value *RelevanceScenario) { value.Segments[1].FromGate = &unknown }, want: "unknown from_gate"},
		{name: "unknown_to", edit: func(value *RelevanceScenario) { value.Segments[1].ToGate = &unknown }, want: "unknown to_gate"},
		{name: "invalid_pair", edit: func(value *RelevanceScenario) {
			value.Segments[1] = RelevanceSegment{MilestoneID: value.Milestone.ID, FromGate: &t1, ToGate: &t0}
		}, want: "invalid boundary"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			value := clone()
			test.edit(&value)
			if err := validateRelevanceSegments(value, suite.Routes); err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("error=%v want=%q", err, test.want)
			}
		})
	}
	unbound := clone()
	unbound.Segments = []RelevanceSegment{{MilestoneID: unbound.Milestone.ID, FromGate: &t0, ToGate: &t1}}
	if err := validateRelevanceSegments(unbound, suite.Routes); err != nil {
		t.Fatal(err)
	}
	if err := validateRelevanceWindows(unbound, suite.Policy, suite.Routes); err == nil || !strings.Contains(err.Error(), "no milestone") {
		t.Fatalf("unbound item error=%v", err)
	}
}

func TestT0T1RelevanceCandidateBindsThreeHonestIntervals(t *testing.T) {
	suite, err := LoadRelevanceSuite("../..", "balance/testdata/t0-t1/relevance-scenario-v2.json")
	if err != nil {
		t.Fatal(err)
	}
	if suite.Scenario.SchemaVersion != 2 || len(suite.Scenario.Segments) != 3 || len(suite.Policy.Items) != 19 || len(suite.Policy.Groups) != 0 {
		t.Fatalf("candidate segments=%d items=%d groups=%d", len(suite.Scenario.Segments), len(suite.Policy.Items), len(suite.Policy.Groups))
	}
	want := [][2]string{{"", "gate.t0_to_t1"}, {"gate.t0_to_t1", "gate.t2_to_t3"}, {"gate.t2_to_t3", "gate.t3_to_t4"}}
	for index, segment := range suite.Scenario.Segments {
		from := ""
		if segment.FromGate != nil {
			from = *segment.FromGate
		}
		if segment.MilestoneID != suite.Scenario.Milestone.ID || segment.ToGate == nil || from != want[index][0] || *segment.ToGate != want[index][1] {
			t.Fatalf("candidate segment[%d]=%+v", index, segment)
		}
	}
}

func TestRelevanceScheduleCardinalityIsBoundedBeforeMaterialization(t *testing.T) {
	for _, policy := range []string{"casual.phase0", "chaos.phase0"} {
		for _, horizon := range []int64{0, 299_999, 300_000, 86_400_000, 172_800_123} {
			got, err := actionCount(policy, horizon)
			if err != nil {
				t.Fatal(err)
			}
			if want := int64(len(actionTimes(policy, horizon))); got != want {
				t.Fatalf("policy=%s horizon=%d count=%d materialized=%d", policy, horizon, got, want)
			}
		}
	}
	suite, err := LoadRelevanceSuite("../..", "testdata/harness/relevance/scenario-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	suite.Scenario.HorizonMS = relevanceMaxSafeInteger
	if _, err := suite.RunRelevance(); err == nil || !strings.Contains(err.Error(), "runaway preflight") {
		t.Fatalf("huge horizon reached schedule materialization: %v", err)
	}
}

func TestRelevanceReducesSeedsBeforePersonaAnyAndPrunesDominatedState(t *testing.T) {
	value := func(input int64) *int64 { return &input }
	matrix := map[string][]relevancePairedResult{
		"casual.phase0": {
			{baseline: value(100), ablated: value(110)},
			{baseline: value(100), ablated: value(120)},
		},
		"reference.greedy": {
			{baseline: value(100), ablated: value(200)},
			{baseline: value(100), ablated: value(210)},
		},
	}
	reduced, err := reduceRelevancePairMatrix(matrix, "worst", "milestone.test", 500)
	if err != nil || reduced.DeltaMS == nil || *reduced.DeltaMS != 100 {
		t.Fatalf("persona ANY reduction=%+v err=%v", reduced, err)
	}
	suite, err := LoadRelevanceSuite("../..", "testdata/harness/relevance/scenario-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	right, err := suite.newRelevanceState()
	if err != nil {
		t.Fatal(err)
	}
	left, err := cloneState(suite.Catalog, right)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := left.Ledger.ApplyAccrual(economy.Transaction{Entries: []economy.Entry{{ResourceID: "company.cash", Delta: decimal.One}}}); err != nil {
		t.Fatal(err)
	}
	if !relevanceStateDominates(left, right, suite.Catalog) || relevanceStateDominates(right, left, suite.Catalog) {
		t.Fatal("componentwise resource dominance was not strict and directional")
	}
	counter := &relevanceCounter{limit: 10_000}
	persona, err := suite.runPersona(RelevanceRunSpec{PolicyID: "casual.phase0", SeedStart: "7", SeedCount: 1}, 7,
		production.AblationMask{}, counter)
	if err != nil || persona.MilestoneMS == nil || counter.value == 0 {
		t.Fatalf("casual relevance persona=%+v transitions=%d err=%v", persona, counter.value, err)
	}
	suite.Scenario.MaxDecisions = relevanceMaxSafeInteger
	if _, err := suite.preflightTransitionCeiling(1, 0); err == nil {
		t.Fatal("transition preflight overflow failed open")
	}
}

func TestRelevanceRegistryIsFailClosedForActiveCatalogs(t *testing.T) {
	entries, err := LoadRelevanceRegistry("../..")
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Active {
		t.Fatalf("fixture registry entries=%+v", entries)
	}
	report := RelevanceReport{Failures: []string{"relevance_floor:upgrade.dead"}}
	if err := ValidateActiveRelevanceReport(entries[0], report); err != nil {
		t.Fatalf("inactive fixture unexpectedly gated: %v", err)
	}
	entries[0].Active = true
	if err := ValidateActiveRelevanceReport(entries[0], report); err == nil || !strings.Contains(err.Error(), "relevance_floor:upgrade.dead") {
		t.Fatalf("active fixture failed open: %v", err)
	}
}

func TestActiveRelevanceAuthorityUsesEpochArtifactsAndAcceptedHash(t *testing.T) {
	bundle, err := epochseed.Load("../..")
	if err != nil {
		t.Fatal(err)
	}
	policyPath := "balance/relevance/phase0.json"
	bundle.Seed.Artifacts = append(bundle.Seed.Artifacts, epochseed.Artifact{Name: "relevance_policy", Path: policyPath})
	entry := RelevanceRegistryEntry{Scenario: "scenario.active", RelevancePolicy: policyPath,
		JustificationChangelog: epochseed.Current(bundle.Seed).ChangelogRef, Active: true}
	routesPath, _ := epochseed.ArtifactPath(bundle.Seed, "routes")
	if err := bindActiveRelevanceAuthority(&entry, RelevanceScenario{RoutesCatalog: routesPath}, bundle); err != nil {
		t.Fatal(err)
	}
	if entry.ConstantsHash != bundle.Hash {
		t.Fatalf("entry hash=%s bundle hash=%s", entry.ConstantsHash, bundle.Hash)
	}
	invalid := entry
	invalid.RelevancePolicy = "balance/relevance/other.json"
	if err := bindActiveRelevanceAuthority(&invalid, RelevanceScenario{RoutesCatalog: routesPath}, bundle); err == nil {
		t.Fatal("active policy outside the epoch manifest was accepted")
	}
	invalid = entry
	invalid.JustificationChangelog = "changelog/other.md"
	if err := bindActiveRelevanceAuthority(&invalid, RelevanceScenario{RoutesCatalog: routesPath}, bundle); err == nil {
		t.Fatal("active trap evidence outside the current epoch changelog was accepted")
	}
}

func TestTrapExemptionsRequireChangelogEvidence(t *testing.T) {
	root := t.TempDir()
	path := "changelog/test.md"
	if err := os.MkdirAll(filepath.Join(root, "changelog"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, path), []byte("- `relevance.present` — evidence\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	present, missing := "relevance.present", "relevance.missing"
	entry := RelevanceRegistryEntry{JustificationChangelog: path}
	policy := &RelevancePolicy{Items: []RelevancePolicyItem{{TrapExempt: true, JustificationKey: &present}}}
	if err := validateTrapJustifications(root, entry, policy); err != nil {
		t.Fatal(err)
	}
	policy.Items[0].JustificationKey = &missing
	if err := validateTrapJustifications(root, entry, policy); err == nil {
		t.Fatal("missing trap-exemption evidence was accepted")
	}
}

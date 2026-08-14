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
	"cloud-clicker/server/save"
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
	if first.RunBudget.DeclaredRuns != 23 || first.RunBudget.ExecutedRuns != 23 ||
		first.RunBudget.DeclaredTransitions != first.RunBudget.ExecutedTransitions || len(first.Items) != 4 || len(first.Groups) != 4 {
		t.Fatalf("report cardinality=%+v items=%d groups=%d", first.RunBudget, len(first.Items), len(first.Groups))
	}
	if first.GreedyOracle != nil || first.DeviationOracle == nil || first.DeviationOracle.Status != "counterexample" ||
		first.DeviationOracle.MaximumForcedDeviations != 1 || first.DeviationOracle.UnprobedCoordinates != 0 ||
		first.DeviationOracle.ReachedProbes != first.DeviationOracle.ExecutedProbes || first.DeviationOracle.UnreachedProbes != 0 ||
		first.DeviationOracle.StarvedProbes != 0 ||
		first.DeviationOracle.Witness == nil || first.DeviationOracle.Witness.DecisionOrdinal != 8 ||
		first.DeviationOracle.Witness.ForcedArm != "generator.beta" || first.DeviationOracle.Witness.GapPPM != 29_068 ||
		!containsString(first.Failures, "greedy_oracle:deviation_gap") {
		t.Fatalf("deviation oracle=%+v witness=%+v", first.DeviationOracle, first.DeviationOracle.Witness)
	}
	joined := strings.Join(first.Failures, ",")
	for _, expected := range []string{"relevance_floor:upgrade.dead", "trap_floor:upgrade.dead"} {
		if !strings.Contains(joined, expected) {
			t.Fatalf("fixture did not discriminate %q: %v", expected, first.Failures)
		}
	}
	var supported, deliberateTrap bool
	for _, item := range first.Items {
		if item.Support == "group_supported" && item.RelevancePassed {
			supported = true
		}
		if item.PurchasableID == "upgrade.alpha" {
			deliberateTrap = item.TrapExempt && item.JustificationKey != nil && *item.JustificationKey == "relevance.intentional_trap"
		}
	}
	if !supported || !deliberateTrap {
		t.Fatalf("fixture did not demonstrate group-supported deliberate trap: %+v", first.Items)
	}
	if err := ValidateRelevanceReport(first); err != nil {
		t.Fatal(err)
	}
	if first.InstrumentExcludedIDs == nil {
		t.Fatal("schema-v4 report omitted instrument exclusions")
	}
	if !reflect.DeepEqual(first.InstrumentExcludedIDs, []string{"generator.alpha"}) ||
		!containsString(first.Failures, "instrument_affected:exact_id:generator.alpha:trap_floor:generator.alpha") {
		t.Fatalf("instrument exclusions=%v failures=%v", first.InstrumentExcludedIDs, first.Failures)
	}
	var alphaAffected bool
	for _, item := range first.Items {
		if item.PurchasableID == "generator.alpha" {
			alphaAffected = item.InstrumentAffected
		}
	}
	if !alphaAffected {
		t.Fatal("instrument exclusion did not reach the item report")
	}
	invalid := first
	invalid.InstrumentExcludedIDs = nil
	if err := ValidateRelevanceReport(invalid); err == nil {
		t.Fatal("schema-v4 report accepted missing instrument exclusions")
	}
	invalid = first
	invalid.InstrumentExcludedIDs = []string{}
	if err := ValidateRelevanceReport(invalid); err == nil {
		t.Fatal("schema-v4 report accepted an undisclosed instrument exclusion")
	}
	invalid = first
	invalid.Items = append([]RelevanceItemReport(nil), first.Items...)
	invalid.Items[0].ExcludedPersonaIDs = nil
	if err := ValidateRelevanceReport(invalid); err == nil {
		t.Fatal("schema-v3 report accepted a missing persona-exclusion row")
	}
	legacy := invalid
	legacy.SchemaVersion = 2
	legacy.DeviationOracle = nil
	legacy.GreedyOracle = &RelevanceGreedyOracle{MilestoneID: first.DeviationOracle.MilestoneID,
		GreedyMS: 5_558, BeamMS: 5_558, GapPPM: 0, MaximumPPM: 25_000, Passed: true}
	legacy.Failures = append([]string(nil), first.Failures...)
	for index, failure := range legacy.Failures {
		if failure == "greedy_oracle:deviation_gap" {
			legacy.Failures = append(legacy.Failures[:index], legacy.Failures[index+1:]...)
			break
		}
	}
	if err := ValidateRelevanceReport(legacy); err != nil {
		t.Fatalf("schema-v2 report lost backward compatibility: %v", err)
	}
	invalid = first
	invalid.Groups = append([]RelevanceGroupReport(nil), first.Groups...)
	invalid.Groups[0].ExcludedPersonaIDs = nil
	if err := ValidateRelevanceReport(invalid); err == nil {
		t.Fatal("schema-v3 report accepted a missing group persona-exclusion row")
	}
	invalid = first
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
	invalid.DeviationOracle = nil
	if err := ValidateRelevanceReport(invalid); err == nil {
		t.Fatal("schema-v5 report accepted a missing deviation oracle")
	}
	invalid = first
	invalid.Failures = append([]string(nil), first.Failures...)
	for index, failure := range invalid.Failures {
		if failure == "greedy_oracle:deviation_gap" {
			invalid.Failures = append(invalid.Failures[:index], invalid.Failures[index+1:]...)
			break
		}
	}
	if err := ValidateRelevanceReport(invalid); err == nil {
		t.Fatal("report accepted a missing deviation-oracle failure")
	}
	invalid = first
	invalid.DeviationOracle = &RelevanceDeviationOracle{Kind: "deviation.v1", MilestoneID: first.DeviationOracle.MilestoneID,
		Status: "passed", EligibleCoordinates: first.DeviationOracle.EligibleCoordinates,
		SelectedCoordinates: first.DeviationOracle.SelectedCoordinates, ExecutedProbes: first.DeviationOracle.ExecutedProbes,
		ReachedProbes: first.DeviationOracle.ReachedProbes, UnreachedProbes: first.DeviationOracle.UnreachedProbes,
		StarvedProbes:       first.DeviationOracle.StarvedProbes,
		UnprobedCoordinates: first.DeviationOracle.UnprobedCoordinates, MaximumForcedDeviations: 1,
		MaximumPPM: first.DeviationOracle.MaximumPPM, Witness: first.DeviationOracle.Witness}
	if err := ValidateRelevanceReport(invalid); err == nil {
		t.Fatal("passing deviation oracle accepted a counterexample witness")
	}
	invalid = first
	invalid.DeviationOracle = &RelevanceDeviationOracle{Kind: "deviation.v1", MilestoneID: first.DeviationOracle.MilestoneID,
		Status: "incomplete", EligibleCoordinates: 1, SelectedCoordinates: 1, ExecutedProbes: 1,
		ReachedProbes: 0, UnreachedProbes: 0, StarvedProbes: 1, UnprobedCoordinates: 0,
		MaximumForcedDeviations: 1, MaximumPPM: first.DeviationOracle.MaximumPPM}
	invalid.Failures = make([]string, 0, len(first.Failures))
	for _, failure := range first.Failures {
		if failure != "greedy_oracle:deviation_gap" {
			invalid.Failures = append(invalid.Failures, failure)
		}
	}
	invalid.Failures = sortedUniqueStrings(append(invalid.Failures, "greedy_oracle:deviation_incomplete"))
	if err := ValidateRelevanceReport(invalid); err != nil {
		t.Fatalf("schema-v6 report rejected disclosed starved probe: %v", err)
	}
	invalid.DeviationOracle.StarvedProbes = 0
	if err := ValidateRelevanceReport(invalid); err == nil {
		t.Fatal("schema-v6 report accepted undisclosed probe outcome")
	}
	invalid = first
	invalid.Items = append([]RelevanceItemReport(nil), first.Items...)
	deltaItem := -1
	for index := range invalid.Items {
		if invalid.Items[index].IndividualDeltas[0].DeltaMS != nil {
			deltaItem = index
			break
		}
	}
	if deltaItem < 0 {
		t.Fatal("fixture has no reconcilable delta to mutate")
	}
	invalid.Items[deltaItem].IndividualDeltas = append([]RelevanceDelta(nil), first.Items[deltaItem].IndividualDeltas...)
	badDelta := *invalid.Items[deltaItem].IndividualDeltas[0].DeltaMS + 1
	invalid.Items[deltaItem].IndividualDeltas[0].DeltaMS = &badDelta
	invalid.Items[deltaItem].IndividualDeltas[0].Status = "both_reached"
	if err := ValidateRelevanceReport(invalid); err == nil {
		t.Fatal("report accepted a non-reconciling delta")
	}
}

func TestCandidateInstrumentDependenciesInvalidateTargetedUpgradeFindings(t *testing.T) {
	tests := []struct {
		scenario  string
		upgradeID string
		removedID string
	}{
		{scenario: "balance/testdata/t0-t1/relevance-scenario-v2.json", upgradeID: "upgrade.continuous_feed_paper", removedID: "generator.dot_matrix_queue"},
		{scenario: "balance/testdata/t0-t1/relevance-scenario-t1-v2.json", upgradeID: "upgrade.refurbished_sticker", removedID: "generator.beige_tower_v2"},
	}
	for _, test := range tests {
		t.Run(test.upgradeID, func(t *testing.T) {
			suite, err := LoadRelevanceSuite("../..", test.scenario)
			if err != nil {
				t.Fatal(err)
			}
			reasons := instrumentAffectedReasons(production.AblationMask{RemovedGeneratorIDs: []string{test.removedID}}, test.upgradeID, suite.Catalog)
			want := []instrumentAffectedReason{{Kind: "effect_target", RemovedID: test.removedID}}
			if !reflect.DeepEqual(reasons, want) {
				t.Fatalf("reasons=%+v want=%+v", reasons, want)
			}
			failure := relevanceFloorFailure("relevance_floor", test.upgradeID, reasons)
			wantFailure := "instrument_affected:effect_target:" + test.removedID + ":relevance_floor:" + test.upgradeID
			if failure != wantFailure {
				t.Fatalf("failure=%q want=%q", failure, wantFailure)
			}
		})
	}
}

func TestUpgradeBranchProofUsesLegalRankedPrefixAndMaskedControl(t *testing.T) {
	suite, err := LoadRelevanceSuite("../..", "testdata/harness/relevance/scenario-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	catalogBytes, err := os.ReadFile("../../testdata/harness/relevance/economy-v4.json")
	if err != nil {
		t.Fatal(err)
	}
	catalogBytes = bytes.Replace(catalogBytes, []byte(`"amount": "9e2"`), []byte(`"amount": "1e2"`), 1)
	catalogBytes = bytes.Replace(catalogBytes, []byte(`"value": "9e2"`), []byte(`"value": "1e2"`), 1)
	suite.Catalog, err = economy.LoadCatalog(catalogBytes)
	if err != nil {
		t.Fatal(err)
	}
	upgrade, ok := suite.Catalog.Upgrade("upgrade.alpha")
	if !ok {
		t.Fatal("fixture upgrade is absent")
	}
	proof, err := suite.runUpgradeBranchProof(upgrade, 1, &relevanceCounter{limit: 2_000_000})
	if err != nil {
		t.Fatal(err)
	}
	if proof.PurchasableKind != relevanceBranchUpgrade || proof.PurchasableID != upgrade.ID ||
		proof.SelectedAtMS == nil || proof.BaselineMS == nil || proof.MaskedMS == nil || proof.DeltaMS == nil || !proof.Passed ||
		*proof.DeltaMS < proof.EpsilonMS || !containsString(proof.RemovedGeneratorIDs, "generator.beta") ||
		!containsString(proof.RemovedUpgradeIDs, "upgrade.dead") {
		t.Fatalf("branch proof=%+v", proof)
	}

	invalid := proof
	invalid.MaskedMS = invalid.BaselineMS
	zero := int64(0)
	invalid.DeltaMS = &zero
	invalid.Passed = true
	report := RelevanceBranchReport{SchemaVersion: RelevanceBranchReportSchemaVersion, ScenarioID: suite.Scenario.ID, ScenarioHash: suite.ScenarioHash,
		ConstantsHash: suite.ConstantsHash, RelevancePolicyHash: suite.Policy.Hash, MainReferenceMS: 1,
		Proofs: []RelevanceBranchProof{invalid}, Failures: []string{}}
	if err := ValidateRelevanceBranchReport(report); err == nil {
		t.Fatal("branch report accepted a non-discriminating masked control")
	}
}

func TestGeneratorBranchProofUsesLegalRankedPrefixAndMaskedControl(t *testing.T) {
	suite, err := LoadRelevanceSuite("../..", "testdata/harness/relevance/scenario-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	generator, ok := suite.Catalog.GeneratorClass("generator.beta")
	if !ok {
		t.Fatal("fixture generator is absent")
	}
	proof, err := suite.runGeneratorBranchProof(generator, 1, &relevanceCounter{limit: 2_000_000})
	if err != nil {
		t.Fatal(err)
	}
	if proof.PurchasableKind != relevanceBranchGenerator || proof.PurchasableID != generator.ID ||
		proof.EffectTargetID != generator.ID || proof.SelectedAtMS == nil || proof.BaselineMS == nil ||
		proof.MaskedMS == nil || proof.DeltaMS == nil || !proof.Passed || *proof.DeltaMS < proof.EpsilonMS ||
		len(proof.RemovedGeneratorIDs) != 0 ||
		!containsString(proof.RemovedUpgradeIDs, "upgrade.alpha") ||
		!containsString(proof.RemovedUpgradeIDs, "upgrade.dead") {
		t.Fatalf("generator branch proof=%+v", proof)
	}

	invalid := proof
	invalid.MaskedMS = invalid.BaselineMS
	zero := int64(0)
	invalid.DeltaMS = &zero
	invalid.Passed = true
	report := RelevanceBranchReport{SchemaVersion: RelevanceBranchReportSchemaVersion, ScenarioID: suite.Scenario.ID,
		ScenarioHash: suite.ScenarioHash, ConstantsHash: suite.ConstantsHash, RelevancePolicyHash: suite.Policy.Hash,
		MainReferenceMS: 1, Proofs: []RelevanceBranchProof{invalid}, Failures: []string{}}
	if err := ValidateRelevanceBranchReport(report); err == nil {
		t.Fatal("branch report accepted a non-discriminating generator-masked control")
	}
}

func TestUpgradeBranchReportDerivesEveryUpgradeMissedByMainReference(t *testing.T) {
	suite, err := LoadRelevanceSuite("../..", "testdata/harness/relevance/scenario-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	report, err := suite.RunRelevanceBranchProofs(nil)
	if err != nil {
		t.Fatal(err)
	}
	ids := make([]string, 0, len(report.Proofs))
	for _, proof := range report.Proofs {
		ids = append(ids, proof.PurchasableID)
	}
	if !reflect.DeepEqual(ids, []string{"upgrade.alpha", "upgrade.dead"}) ||
		!containsString(report.Failures, "branch_unselected:upgrade.dead") {
		t.Fatalf("proof ids=%v failures=%v", ids, report.Failures)
	}
}

func TestBranchReportDerivesFailedUninstrumentedGeneratorFromValidatedMeasurement(t *testing.T) {
	suite, err := LoadRelevanceSuite("../..", "testdata/harness/relevance/scenario-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	measurement, err := suite.RunRelevance()
	if err != nil {
		t.Fatal(err)
	}
	measurement.InstrumentExcludedIDs = []string{}
	cleanFailures := []string{}
	for _, failure := range measurement.Failures {
		if !strings.HasPrefix(failure, "instrument_affected:") {
			cleanFailures = append(cleanFailures, failure)
		}
	}
	measurement.Failures = cleanFailures
	found := false
	for index := range measurement.Items {
		item := &measurement.Items[index]
		item.InstrumentAffected = false
		if item.PurchasableID != "generator.alpha" {
			continue
		}
		item.RelevancePassed = false
		item.Support = "failed"
		item.SupportingGroupID = nil
		item.NearestPassingEpsilonMS = item.EpsilonMS
		measurement.Failures = sortedUniqueStrings(append(measurement.Failures,
			"relevance_floor:generator.alpha"))
		found = true
	}
	if !found {
		t.Fatal("fixture measurement omitted generator.alpha")
	}
	report, err := suite.RunRelevanceBranchProofs(&measurement)
	if err != nil {
		t.Fatal(err)
	}
	proofFound := false
	for _, proof := range report.Proofs {
		if proof.PurchasableID == "generator.alpha" {
			proofFound = proof.PurchasableKind == relevanceBranchGenerator
		}
	}
	if !proofFound {
		t.Fatalf("branch proofs=%+v", report.Proofs)
	}

	stale := measurement
	stale.ConstantsHash = "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	if _, err := suite.RunRelevanceBranchProofs(&stale); err == nil || !strings.Contains(err.Error(), "identity mismatch") {
		t.Fatalf("stale measurement error=%v", err)
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

func TestRelevanceScenarioRequiresTwoBeamChildren(t *testing.T) {
	suite, err := LoadRelevanceSuite("../..", "testdata/harness/relevance/scenario-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	invalid := suite.Scenario
	invalid.BeamChildren = 1
	if err := validateRelevanceScenario(invalid); err == nil {
		t.Fatal("single-child beam accepted even though it cannot explore the runner-up")
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

func TestRelevanceOracleTreatsEqualityAsHealthyAndFailsOnRegression(t *testing.T) {
	gap, passed, failure, err := relevanceOracleOutcome(100, 100, 50_000)
	if err != nil || gap != 0 || !passed || failure != "" {
		t.Fatalf("equal oracle gap=%d passed=%v failure=%q err=%v", gap, passed, failure, err)
	}
	gap, passed, failure, err = relevanceOracleOutcome(100, 101, 50_000)
	if err != nil || gap != 0 || passed || failure != "greedy_oracle:beam_not_better" {
		t.Fatalf("worse beam gap=%d passed=%v failure=%q err=%v", gap, passed, failure, err)
	}
}

func TestRelevanceMilestonePreflightFailsBeforeTheAblationMatrix(t *testing.T) {
	suite, err := LoadRelevanceSuite("../..", "testdata/harness/relevance/scenario-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	suite.Scenario.MaxDecisions = 1
	suite.Scenario.HorizonMS = 1
	suite.Scenario.DecisionHorizonsMS = []int64{1}
	report, err := suite.RunRelevance()
	if err != nil || !reflect.DeepEqual(report.Failures, []string{"reference_decision_starved"}) || report.RunBudget.ExecutedRuns != 0 {
		t.Fatalf("decision-starved preflight report=%+v err=%v", report, err)
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

}

func TestT1SegmentProgressionUsesTheRealGateTransition(t *testing.T) {
	suite, err := LoadRelevanceSuite("../..", "balance/testdata/t0-t1/relevance-scenario-t1-v2.json")
	if err != nil {
		t.Fatal(err)
	}
	state, err := suite.newRelevanceState()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := state.Ledger.ApplyAccrual(economy.Transaction{Entries: []economy.Entry{{
		ResourceID: "company.cash", Delta: decimal.FromString("2.0836e7"),
	}}}); err != nil {
		t.Fatal(err)
	}
	before, err := cloneState(suite.Catalog, state)
	if err != nil {
		t.Fatal(err)
	}
	request := candidateIntent("upgrade.rack_rail_standardization", 1)
	rejected, err := production.SimulateTransition(request, before, suite.Catalog,
		production.SimulationDependencies{Routes: suite.Routes}, relevanceRevision(1, suite.ConstantsHash),
		production.ModeOnline, relevanceNow(0), nil, nil, production.AblationMask{})
	if err != nil || rejected.Decision.Outcome != save.IntentRejected || before.UpgradesOwned[request.UpgradeID] {
		t.Fatalf("pre-gate decision=%+v owned=%v err=%v", rejected.Decision, before.UpgradesOwned, err)
	}

	revision := int64(1)
	counter := &relevanceCounter{limit: 100}
	atMS, _, progressed, err := suite.progressSegmentGates(state, &revision, 0, 0,
		production.ModeOnline, production.AblationMask{}, counter)
	if err != nil || !progressed || atMS != 0 || revision != 2 || !state.GatesCrossed["gate.t0_to_t1"] ||
		state.GatesCrossed["gate.t2_to_t3"] || state.Tier != 1 || counter.value != 1 {
		t.Fatalf("progressed=%v at=%d revision=%d gates=%v tier=%d work=%d err=%v",
			progressed, atMS, revision, state.GatesCrossed, state.Tier, counter.value, err)
	}
	request = candidateIntent("upgrade.rack_rail_standardization", revision)
	applied, err := production.SimulateTransition(request, state, suite.Catalog,
		production.SimulationDependencies{Routes: suite.Routes}, relevanceRevision(revision, suite.ConstantsHash),
		production.ModeOnline, relevanceNow(0), nil, nil, production.AblationMask{})
	if err != nil || applied.Decision.Outcome != save.IntentApplied || !state.UpgradesOwned[request.UpgradeID] {
		t.Fatalf("post-gate decision=%+v owned=%v err=%v", applied.Decision, state.UpgradesOwned, err)
	}
}

func TestT1SegmentProgressionFindsTheFirstEligibleDecisionBoundary(t *testing.T) {
	suite, err := LoadRelevanceSuite("../..", "balance/testdata/t0-t1/relevance-scenario-t1-v2.json")
	if err != nil {
		t.Fatal(err)
	}
	state, err := suite.newRelevanceState()
	if err != nil {
		t.Fatal(err)
	}
	state.GeneratorCounts["generator.beige_tower"] = 1
	state.GeneratorPurchasedTotal = 1
	if _, err := state.Ledger.ApplyAccrual(economy.Transaction{Entries: []economy.Entry{{
		ResourceID: "company.cash", Delta: decimal.FromString("9.9999e4"),
	}}}); err != nil {
		t.Fatal(err)
	}
	revision := int64(1)
	counter := &relevanceCounter{limit: 100}
	atMS, _, progressed, err := suite.progressSegmentGates(state, &revision, 0, 60_000,
		production.ModeOnline, production.AblationMask{}, counter)
	if err != nil || !progressed || atMS != 1_000 || !state.GatesCrossed["gate.t0_to_t1"] ||
		state.GatesCrossed["gate.t2_to_t3"] {
		t.Fatalf("progressed=%v at=%d gates=%v work=%d err=%v", progressed, atMS, state.GatesCrossed, counter.value, err)
	}
}

func TestT0T1BeamUsesTheReferenceProjectedMilestoneRanking(t *testing.T) {
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
	if manual.Applied == 0 {
		t.Fatal("reference bootstrap applied no manual actions")
	}
	want, err := suite.rankCandidates(state, 2, 0, suite.Scenario.DecisionHorizonsMS[0], production.AblationMask{}, counter)
	if err != nil {
		t.Fatal(err)
	}
	got, err := suite.rankBeamCandidates(state, 2, 0, suite.Scenario.DecisionHorizonsMS[0], production.AblationMask{}, counter)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) == 0 || len(got) > int(suite.Scenario.BeamChildren) || len(want) < len(got) {
		t.Fatalf("beam candidates=%+v reference candidates=%+v", got, want)
	}
	for index := range got {
		if got[index].ID != want[index].ID || compareRelevanceProjections(got[index].Projection, want[index].Projection) != 0 || got[index].AtMS != want[index].AtMS {
			t.Fatalf("beam candidate[%d]=%+v reference=%+v", index, got[index], want[index])
		}
	}
	if state.GeneratorPurchasedTotal != 0 {
		t.Fatalf("ranking mutated source state: purchases=%d", state.GeneratorPurchasedTotal)
	}
}

func TestProjectedMilestoneRankingMakesBankFirstClassAndPurchasesWinTies(t *testing.T) {
	suite, err := LoadRelevanceSuite("../..", "balance/testdata/t0-t1/relevance-scenario-v2.json")
	if err != nil {
		t.Fatal(err)
	}
	stateWithCash := func(amount decimal.Decimal) *save.State {
		state, stateErr := suite.newRelevanceState()
		if stateErr != nil {
			t.Fatal(stateErr)
		}
		state.GeneratorCounts["generator.beige_tower"] = 1
		state.GeneratorPurchasedTotal = 1
		if _, stateErr = state.Ledger.ApplyAccrual(economy.Transaction{Entries: []economy.Entry{{ResourceID: "company.cash", Delta: amount}}}); stateErr != nil {
			t.Fatal(stateErr)
		}
		return state
	}
	counter := &relevanceCounter{limit: 100_000}
	next := suite.Scenario.DecisionHorizonsMS[0]
	bankState := stateWithCash(decimal.FromString("9.9999e4"))
	purchases, bank, bankAtMS, err := suite.rankDecisionOptions(bankState, 1, 0, next, 1, production.AblationMask{}, counter)
	if err != nil || !bank || len(purchases) != 0 || bankAtMS != 1_000 {
		t.Fatalf("bank option purchases=%+v bank=%v at=%d err=%v", purchases, bank, bankAtMS, err)
	}
	tieProjection := relevanceProjection{Numerator: decimal.New(1, 4), Denominator: decimal.One}
	tied := []relevanceCandidate{{ID: "generator.zulu", Projection: tieProjection}, {ID: "generator.alpha", Projection: tieProjection}}
	sortRelevanceCandidates(tied)
	purchases, bank, _, err = selectProjectedDecisionOptions(tied, tieProjection, true, 0, next, 1)
	if err != nil || bank || len(purchases) != 1 || purchases[0].ID != "generator.alpha" {
		t.Fatalf("tie option purchases=%+v bank=%v err=%v", purchases, bank, err)
	}
}

func TestT0T1OpportunityCostCorrectsWitnessRankings(t *testing.T) {
	suite, err := LoadRelevanceSuite("../..", "balance/testdata/t0-t1/relevance-scenario-v2.json")
	if err != nil {
		t.Fatal(err)
	}
	counter := &relevanceCounter{limit: 10_000_000}
	legacy, err := suite.runDirectFromGenesis(production.AblationMask{}, counter)
	if err != nil {
		t.Fatal(err)
	}
	replyRemoved := production.AblationMask{RemovedUpgradeIDs: []string{"upgrade.reply_all_macro"}}
	withoutReply, err := suite.runDirectFromGenesis(replyRemoved, counter)
	if err != nil {
		t.Fatal(err)
	}
	value := func(input *int64) any {
		if input == nil {
			return nil
		}
		return *input
	}
	if legacy.MilestoneMS == nil || withoutReply.MilestoneMS == nil || legacy.DecisionStarved || withoutReply.DecisionStarved ||
		*legacy.MilestoneMS != 451_703 || *withoutReply.MilestoneMS != 452_861 || *legacy.MilestoneMS >= *withoutReply.MilestoneMS {
		t.Fatalf("non-starved reply-all witness=%v removed=%v decisions=%d/%d starved=%v/%v",
			value(legacy.MilestoneMS), value(withoutReply.MilestoneMS), legacy.Decisions, withoutReply.Decisions,
			legacy.DecisionStarved, withoutReply.DecisionStarved)
	}
	mask, diagnostics, err := suite.opportunityAwareMask(production.AblationMask{}, counter)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]bool{
		"generator.dot_matrix_queue": true,
		"generator.nephew_intern":    true,
	}
	for _, diagnostic := range diagnostics {
		if !want[diagnostic.PurchasableID] || diagnostic.BaselineMS == nil || diagnostic.RemovedMS == nil {
			continue
		}
		if *diagnostic.RemovedMS >= *diagnostic.BaselineMS {
			t.Fatalf("opportunity screen kept harmful purchase %s=%d->%d", diagnostic.PurchasableID,
				*diagnostic.BaselineMS, *diagnostic.RemovedMS)
		}
		delete(want, diagnostic.PurchasableID)
	}
	if len(want) != 0 || !reflect.DeepEqual(instrumentExcludedIDs(mask), []string{"generator.dot_matrix_queue", "generator.nephew_intern"}) {
		t.Fatalf("opportunity witnesses missing=%v mask=%+v diagnostics=%+v", want, mask, diagnostics)
	}
	screened, err := suite.runReferenceWithOpportunity(production.AblationMask{}, mask, counter)
	if err != nil {
		t.Fatal(err)
	}
	if screened.DecisionStarved || screened.MilestoneMS == nil || *screened.MilestoneMS != 419_462 || screened.Decisions != 51 {
		t.Fatalf("screened reference=%+v", screened)
	}
}

func TestT1ProjectedMilestoneMeasurement(t *testing.T) {
	suite, err := LoadRelevanceSuite("../..", "balance/testdata/t0-t1/relevance-scenario-t1-v2.json")
	if err != nil {
		t.Fatal(err)
	}
	counter := &relevanceCounter{limit: 2_000_000}
	result, err := suite.runDirectFromGenesis(production.AblationMask{}, counter)
	if err != nil {
		t.Fatal(err)
	}
	if result.DecisionStarved || result.MilestoneMS == nil || *result.MilestoneMS != 4_149_073 || result.Decisions != 289 {
		t.Fatalf("T1 projected measurement milestone=%v decisions=%d at=%d transitions=%d starved=%v",
			result.MilestoneMS, result.Decisions, result.FinalVirtualMS, counter.value, result.DecisionStarved)
	}
}

func TestRegisteredBeamOracleCanFalsifyTheReferenceAtDeclaredParameters(t *testing.T) {
	suite, err := LoadRelevanceSuite("../..", "testdata/harness/relevance/scenario-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	if suite.Scenario.BeamWidth != 8 || suite.Scenario.Milestone.Amount != "1e3" || suite.Scenario.GreedyGapMaximumPPM != 25_000 {
		t.Fatalf("registered oracle parameters drifted: %+v", suite.Scenario)
	}
	diagnostic, err := suite.RunBeamDiagnostic()
	if err != nil {
		t.Fatal(err)
	}
	if diagnostic.Oracle.Passed || diagnostic.Oracle.GapPPM <= diagnostic.Oracle.MaximumPPM {
		t.Fatalf("declared-parameter manual beam did not fire: %+v", diagnostic.Oracle)
	}
}

func TestRegisteredBeamResultIsMonotonicWithWidth(t *testing.T) {
	suite, err := LoadRelevanceSuite("../..", "testdata/harness/relevance/scenario-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	maskCounter := &relevanceCounter{limit: suite.Scenario.RelevanceBudgetMaxTransitions}
	mask, _, err := suite.opportunityAwareMask(production.AblationMask{}, maskCounter)
	if err != nil {
		t.Fatal(err)
	}
	prior := int64(relevanceMaxSafeInteger)
	for _, width := range []int64{1, 8, 32} {
		suite.Scenario.BeamWidth = width
		counter := &relevanceCounter{limit: suite.Scenario.RelevanceBudgetMaxTransitions}
		milestone, runErr := suite.runBeamWithOpportunity(mask, counter)
		if runErr != nil || milestone == nil {
			t.Fatalf("width %d milestone=%v transitions=%d err=%v", width, milestone, counter.value, runErr)
		}
		if *milestone > prior {
			t.Fatalf("beam regressed as width grew: width=%d milestone=%d prior=%d", width, *milestone, prior)
		}
		prior = *milestone
	}
}

func TestBeamTerminalCompletionCacheReusesAnIdenticalSuffix(t *testing.T) {
	suite, err := LoadRelevanceSuite("../..", "testdata/harness/relevance/scenario-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	firstState, err := suite.newRelevanceState()
	if err != nil {
		t.Fatal(err)
	}
	secondState, err := cloneState(suite.Catalog, firstState)
	if err != nil {
		t.Fatal(err)
	}
	cache := map[string]relevanceCompletionCacheEntry{}
	counter := &relevanceCounter{limit: suite.Scenario.RelevanceBudgetMaxTransitions}
	first, err := suite.runRankedCompletionCached(firstState, 1, 0, suite.Scenario.MaxDecisions,
		production.AblationMask{}, counter, cache)
	if err != nil || first.MilestoneMS == nil {
		t.Fatalf("first completion=%+v err=%v", first, err)
	}
	afterFirst := counter.value
	second, err := suite.runRankedCompletionCached(secondState, 1, 0, suite.Scenario.MaxDecisions,
		production.AblationMask{}, counter, cache)
	if err != nil || second.MilestoneMS == nil || *second.MilestoneMS != *first.MilestoneMS || counter.value != afterFirst {
		t.Fatalf("cached completion first=%v second=%v before=%d after=%d err=%v",
			first.MilestoneMS, second.MilestoneMS, afterFirst, counter.value, err)
	}
}

func TestProjectedMilestoneRejectsAbsentResource(t *testing.T) {
	suite, err := LoadRelevanceSuite("../..", "testdata/harness/relevance/scenario-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	state, err := suite.newRelevanceState()
	if err != nil {
		t.Fatal(err)
	}
	suite.Scenario.Milestone.ResourceID = "company.absent"
	if _, _, err := suite.projectedMilestone(state, production.AblationMask{}); err == nil || !strings.Contains(err.Error(), "absent from the save ledger") {
		t.Fatalf("absent milestone resource failed open: %v", err)
	}
}

func TestReferenceDecisionGuardNeverCoasts(t *testing.T) {
	suite, err := LoadRelevanceSuite("../..", "balance/testdata/t0-t1/relevance-scenario-v2.json")
	if err != nil {
		t.Fatal(err)
	}
	state, err := suite.newRelevanceState()
	if err != nil {
		t.Fatal(err)
	}
	counter := &relevanceCounter{limit: suite.Scenario.RelevanceBudgetMaxTransitions}
	rollout, err := suite.runReferenceFrom(state, 1, 0, 1, counter)
	if err != nil {
		t.Fatal(err)
	}
	if rollout.Decisions != 1 || !rollout.DecisionStarved || rollout.MilestoneMS != nil || rollout.FinalVirtualMS >= suite.Scenario.HorizonMS {
		t.Fatalf("reference guard did not stop without coasting: %+v", rollout)
	}
}

func TestTracedReferenceDecisionGuardFailsLoudWithoutCoasting(t *testing.T) {
	suite, err := LoadRelevanceSuite("../..", "balance/testdata/t0-t1/relevance-scenario-v2.json")
	if err != nil {
		t.Fatal(err)
	}
	suite.Scenario.MaxDecisions = 1
	traces := []relevanceDecisionTrace{}
	result, err := suite.runReferenceWithOpportunityTrace(production.AblationMask{}, production.AblationMask{},
		&relevanceCounter{limit: suite.Scenario.RelevanceBudgetMaxTransitions}, &traces)
	if err != nil {
		t.Fatal(err)
	}
	if !result.DecisionStarved || result.MilestoneMS != nil || result.FinalVirtualMS >= suite.Scenario.HorizonMS || len(traces) != 1 {
		t.Fatalf("traced reference guard did not fail loud: result=%+v traces=%d", result, len(traces))
	}
}

func TestBeamChildBoundReducesWorkBeforeRollout(t *testing.T) {
	suite, err := LoadRelevanceSuite("../..", "balance/testdata/t0-t1/relevance-scenario-v2.json")
	if err != nil {
		t.Fatal(err)
	}
	state, err := suite.newRelevanceState()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := state.Ledger.ApplyAccrual(economy.Transaction{Entries: []economy.Entry{{ResourceID: "company.cash", Delta: decimal.New(1, 6)}}}); err != nil {
		t.Fatal(err)
	}
	counter := &relevanceCounter{limit: suite.Scenario.RelevanceBudgetMaxTransitions}
	all, err := suite.rankCandidates(state, 1, 0, suite.Scenario.DecisionHorizonsMS[0], production.AblationMask{}, counter)
	if err != nil {
		t.Fatal(err)
	}
	if len(all) < 3 {
		t.Fatalf("fixture exposes only %d ranked children", len(all))
	}
	suite.Scenario.BeamChildren = 2
	bounded, err := suite.rankBeamCandidates(state, 1, 0, suite.Scenario.DecisionHorizonsMS[0], production.AblationMask{}, counter)
	if err != nil {
		t.Fatal(err)
	}
	if len(bounded) != 2 || bounded[0].ID != all[0].ID || bounded[1].ID != all[1].ID {
		t.Fatalf("pre-rollout bound=%v full=%v", bounded, all)
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
	if report.SchemaVersion != RelevanceReportSchemaVersion || len(report.Items) == 0 || report.Items[0].AvailabilityWindow.FromGate != nil {
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

func TestT0T1RelevanceCandidatesBindTheirMeasuredWindows(t *testing.T) {
	suite, err := LoadRelevanceSuite("../..", "balance/testdata/t0-t1/relevance-scenario-v2.json")
	if err != nil {
		t.Fatal(err)
	}
	if suite.Scenario.SchemaVersion != 2 || len(suite.Scenario.Segments) != 1 || len(suite.Policy.Items) != 9 || len(suite.Policy.Groups) != 0 {
		t.Fatalf("candidate segments=%d items=%d groups=%d", len(suite.Scenario.Segments), len(suite.Policy.Items), len(suite.Policy.Groups))
	}
	want := [][2]string{{"", "gate.t0_to_t1"}}
	for index, segment := range suite.Scenario.Segments {
		from := ""
		if segment.FromGate != nil {
			from = *segment.FromGate
		}
		if segment.MilestoneID != suite.Scenario.Milestone.ID || segment.ToGate == nil || from != want[index][0] || *segment.ToGate != want[index][1] {
			t.Fatalf("candidate segment[%d]=%+v", index, segment)
		}
	}
	t1, err := LoadRelevanceSuite("../..", "balance/testdata/t0-t1/relevance-scenario-t1-v2.json")
	if err != nil {
		t.Fatal(err)
	}
	if len(t1.Scenario.Segments) != 2 || len(t1.Policy.Items) != 18 {
		t.Fatalf("T1 candidate segments=%d items=%d", len(t1.Scenario.Segments), len(t1.Policy.Items))
	}
	measuredItems, measuredGroups, err := t1.measuredPolicyRows()
	if err != nil {
		t.Fatal(err)
	}
	if len(measuredItems) != 9 || len(measuredGroups) != 0 {
		t.Fatalf("T1 evidence scope items=%d groups=%d", len(measuredItems), len(measuredGroups))
	}
	for _, item := range measuredItems {
		if item.Availability.FromGate == nil || *item.Availability.FromGate != "gate.t0_to_t1" ||
			item.Availability.ToGate == nil || *item.Availability.ToGate != "gate.t2_to_t3" {
			t.Fatalf("T1 report retained out-of-window item %+v", item)
		}
	}
	if t1.Scenario.BeamWidth != 1 || t1.Scenario.BeamChildren != 2 || t1.Scenario.MaxDecisions != 384 || t1.Scenario.RelevanceBudgetMaxTransitions != 5_000_000 {
		t.Fatalf("T1 measured branch-B envelope=%+v", t1.Scenario)
	}
	for _, item := range t1.Policy.Items {
		if item.PurchasableID == "generator.legal_dept" {
			t.Fatal("T1 policy included a generator that opens at its terminal milestone")
		}
	}
	trimmed := *suite.Policy
	trimmed.Items = append([]RelevancePolicyItem(nil), suite.Policy.Items[:len(suite.Policy.Items)-1]...)
	if err := validateScopedRelevancePolicy(suite.Scenario, &trimmed, suite.Catalog, suite.Routes); err == nil || !strings.Contains(err.Error(), "item set is incomplete") {
		t.Fatalf("incomplete scoped policy err=%v", err)
	}
	fullFixture, err := LoadRelevanceSuite("../..", "testdata/harness/relevance/scenario-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	invalidFull := *fullFixture.Policy
	invalidFull.Items = append([]RelevancePolicyItem(nil), fullFixture.Policy.Items...)
	invalidFull.Items[0].PurchasableID = "generator.unknown"
	if err := validateScopedRelevancePolicy(fullFixture.Scenario, &invalidFull, fullFixture.Catalog, fullFixture.Routes); err == nil ||
		!strings.Contains(err.Error(), "unknown item") {
		t.Fatalf("full-cardinality policy bypassed identity validation: %v", err)
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
	suite.Scenario.PreflightCeiling = 1
	if _, err := suite.RunRelevance(); err == nil || !strings.Contains(err.Error(), "runaway preflight") {
		t.Fatalf("declarative preflight ceiling failed open: %v", err)
	}
	suite.Scenario.PreflightCeiling = 1_000_000_000_000
	suite.Scenario.HorizonMS = relevanceMaxSafeInteger
	if _, err := suite.RunRelevance(); err == nil || !strings.Contains(err.Error(), "runaway preflight") {
		t.Fatalf("huge horizon reached schedule materialization: %v", err)
	}
}

func TestRelevanceReducesSeedsBeforePersonaAnyAndPrunesDominatedState(t *testing.T) {
	value := func(input int64) *int64 { return &input }
	matrix := map[string][]relevancePairedResult{
		"casual.phase0": {
			{baseline: value(100), ablated: value(110), baselinePurchases: 1},
			{baseline: value(100), ablated: value(120), baselinePurchases: 1},
		},
		"reference.greedy": {
			{baseline: value(100), ablated: value(200), baselinePurchases: 1},
			{baseline: value(100), ablated: value(210), baselinePurchases: 1},
		},
		"spectator.phase0": {
			{baseline: value(100), ablated: value(100), baselinePurchases: 0},
			{baseline: value(100), ablated: value(100), baselinePurchases: 0},
		},
		"sometimes.phase0": {
			{baseline: value(100), ablated: value(150), baselinePurchases: 0},
			{baseline: value(100), ablated: value(150), baselinePurchases: 1},
		},
	}
	reduced, excluded, err := reduceRelevancePairMatrix(matrix, "worst", "milestone.test", 500)
	if err != nil || reduced.DeltaMS == nil || *reduced.DeltaMS != 100 || !reflect.DeepEqual(excluded, []string{"spectator.phase0"}) {
		t.Fatalf("persona ANY reduction=%+v excluded=%v err=%v", reduced, excluded, err)
	}
	if purchases := reduceRelevancePairPurchases(matrix, "worst"); purchases != 1 {
		t.Fatalf("persona ANY purchase count=%d", purchases)
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

func TestRelevanceFixturePinsTrapFloorAndTheRegisteredDeviationWitness(t *testing.T) {
	suite, err := LoadRelevanceSuite("../..", "testdata/harness/relevance/scenario-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	report, err := suite.RunRelevance()
	if err != nil {
		t.Fatal(err)
	}
	if report.DeviationOracle == nil || report.DeviationOracle.Status != "counterexample" ||
		!containsString(report.Failures, "greedy_oracle:deviation_gap") ||
		!containsString(report.Failures, "instrument_affected:effect_target:generator.alpha:trap_floor:upgrade.dead") {
		t.Fatalf("registered deviation witness=%+v failures=%v", report.DeviationOracle, report.Failures)
	}
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func TestRelevanceRegistryIsFailClosedForActiveCatalogs(t *testing.T) {
	entries, err := LoadRelevanceRegistry("../..")
	if err != nil {
		t.Fatal(err)
	}
	var fixture *RelevanceRegistryEntry
	for index := range entries {
		if entries[index].EconomyCatalog == "testdata/harness/relevance/economy-v4.json" {
			fixture = &entries[index]
			break
		}
	}
	if len(entries) != 2 || fixture == nil || fixture.Active {
		t.Fatalf("registry entries=%+v fixture=%+v", entries, fixture)
	}
	report := RelevanceReport{Failures: []string{"relevance_floor:upgrade.dead"}}
	if err := ValidateActiveRelevanceReport(*fixture, report); err != nil {
		t.Fatalf("inactive fixture unexpectedly gated: %v", err)
	}
	fixture.Active = true
	if err := ValidateActiveRelevanceReport(*fixture, report); err == nil || !strings.Contains(err.Error(), "relevance_floor:upgrade.dead") {
		t.Fatalf("active fixture failed open: %v", err)
	}
}

func TestRelevanceRegistryUpdateCanMaterializeMissingEvidence(t *testing.T) {
	entries, err := LoadRelevanceRegistryForUpdate("../..")
	if err != nil {
		t.Fatal(err)
	}
	active := 0
	for _, entry := range entries {
		if entry.Active {
			active++
			if entry.GoldenReport == "" || entry.BranchReport == "" {
				t.Fatalf("active update entry lacks evidence coordinates: %+v", entry)
			}
		}
	}
	if active != 1 {
		t.Fatalf("active entries=%d, want 1", active)
	}
}

func TestRelevanceRegistryUpdateRelaxesOnlyMissingEvidence(t *testing.T) {
	root := t.TempDir()
	const missing = "generated/missing.json"
	if err := validateRelevanceRegistryPath(root, missing, true, true); err != nil {
		t.Fatalf("update rejected missing generated evidence: %v", err)
	}
	if err := validateRelevanceRegistryPath(root, missing, true, false); err == nil {
		t.Fatal("strict check accepted missing generated evidence")
	}
	if err := validateRelevanceRegistryPath(root, missing, false, true); err == nil {
		t.Fatal("update accepted missing source artifact")
	}
}

func TestActiveRelevanceEvidenceRequiresExactBranchCoverage(t *testing.T) {
	entries, err := LoadRelevanceRegistry("../..")
	if err != nil {
		t.Fatal(err)
	}
	var active *RelevanceRegistryEntry
	for index := range entries {
		if entries[index].Active {
			active = &entries[index]
			break
		}
	}
	if active == nil {
		t.Fatal("active relevance registry entry is missing")
	}
	mainData, err := os.ReadFile(filepath.Join("../..", filepath.FromSlash(active.GoldenReport)))
	if err != nil {
		t.Fatal(err)
	}
	var main RelevanceReport
	if err := decodeRelevanceStrict(mainData, &main); err != nil {
		t.Fatal(err)
	}
	branch, err := LoadRegisteredRelevanceBranchReport("../..", *active)
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateActiveRelevanceEvidence(*active, main, branch); err != nil {
		t.Fatal(err)
	}

	missing := branch
	missing.Proofs = append([]RelevanceBranchProof(nil), branch.Proofs[1:]...)
	if err := ValidateActiveRelevanceEvidence(*active, main, missing); err == nil || !strings.Contains(err.Error(), "coverage mismatch") {
		t.Fatalf("missing branch proof accepted: %v", err)
	}

	wrongIdentity := branch
	wrongIdentity.ConstantsHash = "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	if err := ValidateActiveRelevanceEvidence(*active, main, wrongIdentity); err == nil || !strings.Contains(err.Error(), "identity mismatch") {
		t.Fatalf("mismatched branch identity accepted: %v", err)
	}

	uncovered := main
	uncovered.Failures = append(append([]string(nil), main.Failures...), "zzz_unrelated:finding")
	if err := ValidateActiveRelevanceEvidence(*active, uncovered, branch); err == nil || !strings.Contains(err.Error(), "no branch evidence") {
		t.Fatalf("uncovered whole-path finding accepted: %v", err)
	}
}

func TestActiveRelevanceAuthorityUsesEpochArtifactsAndAcceptedHash(t *testing.T) {
	bundle, err := epochseed.Load("../..")
	if err != nil {
		t.Fatal(err)
	}
	policyPath := "balance/relevance/phase0.json"
	for index := range bundle.Seed.Artifacts {
		if bundle.Seed.Artifacts[index].Name == "relevance" {
			bundle.Seed.Artifacts[index].Path = policyPath
		}
	}
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

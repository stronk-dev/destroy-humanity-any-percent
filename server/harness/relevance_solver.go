package harness

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"

	"cloud-clicker/server/decimal"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/production"
	"cloud-clicker/server/save"
)

type relevanceCandidate struct {
	ID              string
	Projection      relevanceProjection
	AtMS            int64
	State           *save.State
	Revision        int64
	RoleActivations []production.RoleActivation
}

// relevanceProjection is an exact comparison ratio in milliseconds. Keeping
// numerator and denominator separate avoids rounding or overflowing projected
// times that lie beyond the safe-integer domain early in a run.
type relevanceProjection struct {
	Numerator   decimal.Decimal
	Denominator decimal.Decimal
}

type relevanceOpportunityDecision struct {
	PurchasableID string
	BaselineMS    *int64
	RemovedMS     *int64
}

type relevanceCounter struct {
	value int64
	limit int64
}

var errReferenceDecisionStarved = errors.New("reference_decision_starved")

func (counter *relevanceCounter) add() error {
	if counter.value >= counter.limit {
		return fmt.Errorf("relevance transition budget exceeded: executed %d, maximum %d", counter.value, counter.limit)
	}
	counter.value++
	return nil
}

func (suite *RelevanceSuite) RunRelevance() (RelevanceReport, error) {
	if suite.Scenario.HorizonMS > suite.Catalog.OfflinePolicy().AccrualCapMS {
		return RelevanceReport{}, fmt.Errorf("relevance runaway preflight horizon %d exceeds online ceiling %d",
			suite.Scenario.HorizonMS, suite.Catalog.OfflinePolicy().AccrualCapMS)
	}
	measuredItems, measuredGroups, err := suite.measuredPolicyRows()
	if err != nil {
		return RelevanceReport{}, err
	}
	nonReferenceSeeds, referenceSeeds := int64(0), int64(0)
	nonReferenceTransitions := int64(0)
	for _, run := range suite.Scenario.Runs {
		if run.Reference {
			if referenceSeeds > relevanceMaxSafeInteger-run.SeedCount {
				return RelevanceReport{}, errors.New("relevance seed cardinality overflow")
			}
			referenceSeeds += run.SeedCount
		} else {
			if nonReferenceSeeds > relevanceMaxSafeInteger-run.SeedCount {
				return RelevanceReport{}, errors.New("relevance seed cardinality overflow")
			}
			nonReferenceSeeds += run.SeedCount
			actions, countErr := actionCount(run.PolicyID, suite.Scenario.HorizonMS)
			if countErr != nil {
				return RelevanceReport{}, countErr
			}
			factor := int64(1 + len(measuredItems) + len(measuredGroups))
			if actions != 0 && run.SeedCount > relevanceMaxSafeInteger/actions ||
				actions*run.SeedCount != 0 && factor > relevanceMaxSafeInteger/(actions*run.SeedCount) ||
				nonReferenceTransitions > relevanceMaxSafeInteger-actions*run.SeedCount*factor {
				return RelevanceReport{}, errors.New("relevance transition preflight overflow")
			}
			nonReferenceTransitions += actions * run.SeedCount * factor
		}
	}
	declaredRuns, err := ComputeRelevanceRunBudget(nonReferenceSeeds, referenceSeeds, int64(len(measuredItems)), int64(len(measuredGroups)), true)
	if err != nil || declaredRuns > suite.Scenario.RelevanceBudgetMaxRuns {
		return RelevanceReport{}, errors.New("relevance run budget exceeds scenario limit")
	}
	transitionCeiling, err := suite.preflightTransitionCeiling(referenceSeeds, nonReferenceTransitions)
	if err != nil {
		return RelevanceReport{}, fmt.Errorf("relevance transition budget preflight: %w", err)
	}
	if transitionCeiling > suite.Scenario.PreflightCeiling {
		return RelevanceReport{}, fmt.Errorf("relevance runaway preflight requires %d, ceiling %d",
			transitionCeiling, suite.Scenario.PreflightCeiling)
	}
	report := RelevanceReport{SchemaVersion: RelevanceReportSchemaVersion, ScenarioID: suite.Scenario.ID,
		ScenarioHash: suite.ScenarioHash, ConstantsHash: suite.ConstantsHash, RelevancePolicyHash: suite.Policy.Hash,
		Items: []RelevanceItemReport{}, Groups: []RelevanceGroupReport{}, TierContributions: []RelevanceTierContribution{},
		RoleActivations: []RoleActivationCount{}, InstrumentExcludedIDs: []string{}, Failures: []string{}}
	counter := &relevanceCounter{limit: suite.Scenario.RelevanceBudgetMaxTransitions}
	opportunityMask, _, err := suite.opportunityAwareMask(production.AblationMask{}, counter)
	if err != nil {
		if errors.Is(err, errReferenceDecisionStarved) {
			return relevanceStarvationDiagnostic(report, declaredRuns, 0, suite.Scenario.RelevanceBudgetMaxTransitions, counter.value), nil
		}
		return RelevanceReport{}, fmt.Errorf("relevance opportunity preflight: %w", err)
	}
	report.InstrumentExcludedIDs = instrumentExcludedIDs(opportunityMask)
	preflight, err := suite.runReferenceWithOpportunity(production.AblationMask{}, opportunityMask, counter)
	if err != nil {
		return RelevanceReport{}, fmt.Errorf("relevance milestone preflight: %w", err)
	}
	if preflight.DecisionStarved {
		return relevanceStarvationDiagnostic(report, declaredRuns, 0,
			suite.Scenario.RelevanceBudgetMaxTransitions, counter.value), nil
	}
	if preflight.MilestoneMS == nil {
		return RelevanceReport{}, fmt.Errorf("milestone_unreachable:%s", suite.Scenario.Milestone.ID)
	}
	itemPairs := map[string]map[string][]relevancePairedResult{}
	removalPairs := map[string]map[string][]relevancePairedResult{}
	groupPairs := map[string]map[string][]relevancePairedResult{}
	baselinePurchases := map[string]int64{}
	baselineRoleRuns := map[string][]map[string]RoleActivationCount{}
	executedRuns := int64(0)
	type oraclePair struct{ greedy, beam *int64 }
	oraclePairs := []oraclePair{}
	baselineUnreached := false
	for _, spec := range suite.Scenario.Runs {
		start, _ := parseSeed(spec.SeedStart)
		for offset := int64(0); offset < spec.SeedCount; offset++ {
			seed := start + uint64(offset)
			baseline, runErr := suite.runPersonaWithOpportunity(spec, seed, production.AblationMask{}, opportunityMask, counter)
			if runErr != nil {
				return RelevanceReport{}, fmt.Errorf("relevance run %s seed %d baseline: %w", spec.PolicyID, seed, runErr)
			}
			executedRuns++
			if spec.Reference && baseline.DecisionStarved {
				return relevanceStarvationDiagnostic(report, declaredRuns, executedRuns,
					suite.Scenario.RelevanceBudgetMaxTransitions, counter.value), nil
			}
			if baseline.MilestoneMS == nil {
				baselineUnreached = true
			}
			baselineRoleRuns[spec.PolicyID] = append(baselineRoleRuns[spec.PolicyID], baseline.Roles)
			for _, item := range measuredItems {
				masked, runErr := suite.runPersonaWithOpportunity(spec, seed, effectMask(item.PurchasableID, suite.Catalog), opportunityMask, counter)
				if runErr != nil {
					return RelevanceReport{}, fmt.Errorf("relevance run %s seed %d effect mask %s: %w", spec.PolicyID, seed, item.PurchasableID, runErr)
				}
				executedRuns++
				if spec.Reference && masked.DecisionStarved {
					return relevanceStarvationDiagnostic(report, declaredRuns, executedRuns,
						suite.Scenario.RelevanceBudgetMaxTransitions, counter.value), nil
				}
				appendRelevancePair(itemPairs, item.PurchasableID, spec.PolicyID, baseline.MilestoneMS, masked.MilestoneMS, baseline.Purchases[item.PurchasableID])
				if spec.Reference {
					removed, removeErr := suite.runPersonaWithOpportunity(spec, seed, removalMask(item.PurchasableID, suite.Catalog), opportunityMask, counter)
					if removeErr != nil {
						return RelevanceReport{}, fmt.Errorf("relevance run %s seed %d removal mask %s: %w", spec.PolicyID, seed, item.PurchasableID, removeErr)
					}
					executedRuns++
					if removed.DecisionStarved {
						return relevanceStarvationDiagnostic(report, declaredRuns, executedRuns,
							suite.Scenario.RelevanceBudgetMaxTransitions, counter.value), nil
					}
					appendRelevancePair(removalPairs, item.PurchasableID, spec.PolicyID, baseline.MilestoneMS, removed.MilestoneMS, baseline.Purchases[item.PurchasableID])
				}
			}
			for _, group := range measuredGroups {
				masked, runErr := suite.runPersonaWithOpportunity(spec, seed, groupMask(group.MemberIDs, suite.Catalog), opportunityMask, counter)
				if runErr != nil {
					return RelevanceReport{}, fmt.Errorf("relevance run %s seed %d group mask %s: %w", spec.PolicyID, seed, group.GroupID, runErr)
				}
				executedRuns++
				if spec.Reference && masked.DecisionStarved {
					return relevanceStarvationDiagnostic(report, declaredRuns, executedRuns,
						suite.Scenario.RelevanceBudgetMaxTransitions, counter.value), nil
				}
				purchases := int64(0)
				for _, memberID := range group.MemberIDs {
					purchases += baseline.Purchases[memberID]
				}
				appendRelevancePair(groupPairs, group.GroupID, spec.PolicyID, baseline.MilestoneMS, masked.MilestoneMS, purchases)
			}
			if spec.Reference {
				beam, beamErr := suite.runBeamWithOpportunity(opportunityMask, counter)
				if beamErr != nil {
					return RelevanceReport{}, fmt.Errorf("relevance run %s seed %d beam: %w", spec.PolicyID, seed, beamErr)
				}
				oraclePairs = append(oraclePairs, oraclePair{greedy: cloneInt64(baseline.MilestoneMS), beam: cloneInt64(beam)})
				executedRuns++ // one beam invocation is one R14 run, regardless of internal transitions.
			}
		}
	}
	if executedRuns != declaredRuns || len(oraclePairs) == 0 {
		return RelevanceReport{}, errors.New("relevance run cardinality mismatch")
	}
	if baselineUnreached {
		report.Failures = append(report.Failures, "baseline_unreached:"+suite.Scenario.Milestone.ID)
	}
	for _, item := range measuredItems {
		baselinePurchases[item.PurchasableID] = reduceRelevancePairPurchases(itemPairs[item.PurchasableID], suite.Scenario.Reducer)
	}
	baselineRoles, roleReduceErr := reduceRelevanceRoles(baselineRoleRuns, suite.Scenario.Reducer)
	if roleReduceErr != nil {
		return RelevanceReport{}, roleReduceErr
	}
	groupReduced := map[string]RelevanceDelta{}
	for _, group := range measuredGroups {
		reduced, excluded, reduceErr := reduceRelevancePairMatrix(groupPairs[group.GroupID], suite.Scenario.Reducer,
			suite.Scenario.Milestone.ID, suite.Scenario.HorizonMS)
		if reduceErr != nil {
			return RelevanceReport{}, reduceErr
		}
		passed := reduced.DeltaMS != nil && *reduced.DeltaMS >= group.EpsilonMS
		report.Groups = append(report.Groups, RelevanceGroupReport{GroupID: group.GroupID, Axis: group.Axis,
			ExcludedPersonaIDs: excluded, Deltas: []RelevanceDelta{reduced}, Passed: passed})
		groupReduced[group.GroupID] = reduced
		if group.Axis == "tier" {
			report.TierContributions = append(report.TierContributions, RelevanceTierContribution{GroupID: group.GroupID, Deltas: []RelevanceDelta{reduced}})
		}
	}
	for _, item := range measuredItems {
		individual, excluded, reduceErr := reduceRelevancePairMatrix(itemPairs[item.PurchasableID], suite.Scenario.Reducer,
			suite.Scenario.Milestone.ID, suite.Scenario.HorizonMS)
		if reduceErr != nil {
			return RelevanceReport{}, reduceErr
		}
		removal, _, reduceErr := reduceRelevancePairMatrix(removalPairs[item.PurchasableID], suite.Scenario.Reducer,
			suite.Scenario.Milestone.ID, suite.Scenario.HorizonMS)
		if reduceErr != nil {
			return RelevanceReport{}, reduceErr
		}
		relevancePassed := individual.DeltaMS != nil && *individual.DeltaMS >= item.EpsilonMS
		support := "failed"
		var supportingGroup *string
		if relevancePassed {
			support = "individual"
		} else {
			for _, groupID := range item.GroupIDs {
				group := groupReduced[groupID]
				threshold := policyGroup(suite.Policy, groupID).EpsilonMS
				if group.DeltaMS != nil && *group.DeltaMS >= threshold {
					relevancePassed, support = true, "group_supported"
					copy := groupID
					supportingGroup = &copy
					break
				}
			}
		}
		nearest := int64(0)
		if !relevancePassed {
			nearest = item.EpsilonMS
			if individual.DeltaMS != nil {
				nearest -= *individual.DeltaMS
				if nearest < 0 {
					nearest = 0
				}
			}
		}
		trapPassed := baselinePurchases[item.PurchasableID] > 0 || item.TrapExempt
		instrumentAffected := maskRemoves(opportunityMask, item.PurchasableID, suite.Catalog)
		report.Items = append(report.Items, RelevanceItemReport{PurchasableID: item.PurchasableID,
			AvailabilityWindow: item.Availability, EpsilonMS: item.EpsilonMS, TrapExempt: item.TrapExempt,
			JustificationKey: item.JustificationKey, BaselinePurchaseCount: baselinePurchases[item.PurchasableID], ExcludedPersonaIDs: excluded,
			InstrumentAffected: instrumentAffected,
			IndividualDeltas:   []RelevanceDelta{individual}, ActionRemovalDeltas: []RelevanceDelta{removal}, Support: support,
			SupportingGroupID: supportingGroup, NearestPassingEpsilonMS: nearest, RelevancePassed: relevancePassed, TrapPassed: trapPassed})
		if !relevancePassed {
			report.Failures = append(report.Failures, relevanceFloorFailure("relevance_floor", item.PurchasableID, instrumentAffected))
		}
		if !trapPassed {
			report.Failures = append(report.Failures, relevanceFloorFailure("trap_floor", item.PurchasableID, instrumentAffected))
		}
	}
	report.RoleActivations = sortedRoleActivations(baselineRoles)
	roleGenerators := map[string]bool{}
	for _, role := range report.RoleActivations {
		if role.Count > 0 {
			roleGenerators[role.GeneratorID] = true
		}
	}
	for _, item := range measuredItems {
		if _, generator := suite.Catalog.GeneratorClass(item.PurchasableID); generator && !roleGenerators[item.PurchasableID] {
			report.Failures = append(report.Failures, "role_floor:"+item.PurchasableID)
		}
	}
	var selected *RelevanceGreedyOracle
	for _, pair := range oraclePairs {
		if pair.greedy == nil || pair.beam == nil {
			selected = nil
			break
		}
		gap := int64(0)
		gap, gapErr := relevanceGapPPM(*pair.greedy, *pair.beam)
		if gapErr != nil {
			return RelevanceReport{}, gapErr
		}
		candidate := &RelevanceGreedyOracle{MilestoneID: suite.Scenario.Milestone.ID,
			GreedyMS: *pair.greedy, BeamMS: *pair.beam, GapPPM: gap, MaximumPPM: suite.Scenario.GreedyGapMaximumPPM,
			Passed: gap <= suite.Scenario.GreedyGapMaximumPPM}
		if selected == nil || candidate.GapPPM > selected.GapPPM {
			selected = candidate
		}
	}
	if selected == nil {
		report.Failures = append(report.Failures, "greedy_oracle:milestone_unreached")
	} else {
		report.GreedyOracle = selected
		if !selected.Passed {
			report.Failures = append(report.Failures, "greedy_oracle:gap")
		}
	}
	report.RunBudget = RelevanceRunBudget{DeclaredRuns: declaredRuns, ExecutedRuns: executedRuns,
		DeclaredTransitions: counter.value, ExecutedTransitions: counter.value}
	report.Failures = sortedUniqueStrings(report.Failures)
	if err := ValidateRelevanceReport(report); err != nil {
		return RelevanceReport{}, err
	}
	return report, nil
}

func relevanceFloorFailure(kind, id string, instrumentAffected bool) string {
	prefix := ""
	if instrumentAffected {
		prefix = "instrument_affected:"
	}
	return prefix + kind + ":" + id
}

func relevanceStarvationDiagnostic(report RelevanceReport, declaredRuns, executedRuns, declaredTransitions, executedTransitions int64) RelevanceReport {
	report.RunBudget = RelevanceRunBudget{DeclaredRuns: declaredRuns, ExecutedRuns: executedRuns,
		DeclaredTransitions: declaredTransitions, ExecutedTransitions: executedTransitions}
	report.Failures = []string{errReferenceDecisionStarved.Error()}
	return report
}

func instrumentExcludedIDs(mask production.AblationMask) []string {
	values := append([]string{}, mask.RemovedGeneratorIDs...)
	values = append(values, mask.RemovedUpgradeIDs...)
	return sortedUniqueStrings(values)
}

// measuredPolicyRows keeps optimization candidates cumulative while limiting
// ablation verdicts to the availability window that closes at the scenario's
// target gate. Earlier-window rows are measured by their earlier scenario;
// their downstream effects cannot earn evidence after that window closes.
func (suite *RelevanceSuite) measuredPolicyRows() ([]RelevancePolicyItem, []RelevancePolicyGroup, error) {
	target, err := decimal.ParseCanonical(suite.Scenario.Milestone.Amount)
	if err != nil {
		return nil, nil, err
	}
	gateOrder := map[string]int{}
	targetBoundary := -1
	targetGateID := ""
	targetMatches := 0
	for index, gate := range suite.Routes.Gates() {
		gateOrder[gate.ID] = index
		for _, requirement := range gate.Requirement {
			if requirement.ResourceID == suite.Scenario.Milestone.ResourceID && requirement.Amount.Eq(target) {
				targetBoundary = index
				targetGateID = gate.ID
				targetMatches++
			}
		}
	}
	if targetMatches != 1 {
		return nil, nil, errors.New("relevance evidence scope requires one target gate coordinate")
	}
	closesAtTarget := false
	for _, segment := range suite.Scenario.Segments {
		if segment.ToGate != nil && *segment.ToGate == targetGateID {
			closesAtTarget = true
			break
		}
	}
	if !closesAtTarget {
		return append([]RelevancePolicyItem(nil), suite.Policy.Items...), append([]RelevancePolicyGroup(nil), suite.Policy.Groups...), nil
	}
	measured := make([]RelevancePolicyItem, 0, len(suite.Policy.Items))
	measuredIDs := map[string]bool{}
	for _, item := range suite.Policy.Items {
		fromBoundary := -1
		if item.Availability.FromGate != nil {
			fromBoundary = gateOrder[*item.Availability.FromGate]
		}
		toBoundary := len(gateOrder)
		if item.Availability.ToGate != nil {
			toBoundary = gateOrder[*item.Availability.ToGate]
		}
		if fromBoundary < targetBoundary && targetBoundary <= toBoundary {
			measured = append(measured, item)
			measuredIDs[item.PurchasableID] = true
		}
	}
	groups := make([]RelevancePolicyGroup, 0, len(suite.Policy.Groups))
	for _, group := range suite.Policy.Groups {
		copy := group
		copy.MemberIDs = nil
		for _, id := range group.MemberIDs {
			if measuredIDs[id] {
				copy.MemberIDs = append(copy.MemberIDs, id)
			}
		}
		if len(copy.MemberIDs) > 0 {
			groups = append(groups, copy)
		}
	}
	if len(measured) == 0 {
		return nil, nil, errors.New("relevance scenario has no evidence rows at its target window")
	}
	return measured, groups, nil
}

func (suite *RelevanceSuite) preflightTransitionCeiling(referenceSeeds, nonReferenceTransitions int64) (int64, error) {
	items := int64(len(suite.Policy.Items))
	binarySteps := int64(1)
	for span := suite.Scenario.HorizonMS; span > 1; span = (span + 1) / 2 {
		binarySteps++
	}
	checkedMul := func(values ...int64) (int64, error) {
		result := int64(1)
		for _, value := range values {
			if value < 0 || value != 0 && result > relevanceMaxSafeInteger/value {
				return 0, errors.New("relevance transition preflight overflow")
			}
			result *= value
		}
		return result, nil
	}
	// Per candidate: one affordability lower bound, two marginal-output lower
	// bounds, the purchase, and conservative endpoint probes.
	perCandidate, err := checkedMul(items, 3*binarySteps+6)
	if err != nil || perCandidate >= relevanceMaxSafeInteger {
		return 0, errors.New("relevance transition preflight overflow")
	}
	perDecision := perCandidate + 1
	perRun, err := checkedMul(suite.Scenario.MaxDecisions, perDecision)
	if err != nil {
		return 0, err
	}
	beamPerSeed, err := checkedMul(suite.Scenario.MaxDecisions, suite.Scenario.BeamWidth, suite.Scenario.BeamChildren+1, perRun+perDecision)
	if err != nil {
		return 0, err
	}
	beam, err := checkedMul(referenceSeeds, beamPerSeed)
	if err != nil {
		return 0, err
	}
	referenceRuns, err := checkedMul(referenceSeeds, 1+2*items+int64(len(suite.Policy.Groups)))
	if err != nil {
		return 0, err
	}
	runWork, err := checkedMul(referenceRuns, perRun)
	if err != nil || runWork > relevanceMaxSafeInteger-beam || runWork+beam > relevanceMaxSafeInteger-nonReferenceTransitions {
		return 0, errors.New("relevance transition preflight overflow")
	}
	return runWork + beam + nonReferenceTransitions, nil
}

func (suite *RelevanceSuite) runPersona(spec RelevanceRunSpec, seed uint64, mask production.AblationMask, counter *relevanceCounter) (relevanceRunResult, error) {
	if spec.Reference {
		return suite.runReference(mask, counter)
	}
	return suite.runPersonaWithOpportunity(spec, seed, mask, production.AblationMask{}, counter)
}

func (suite *RelevanceSuite) runPersonaWithOpportunity(spec RelevanceRunSpec, seed uint64, mask, opportunityMask production.AblationMask, counter *relevanceCounter) (relevanceRunResult, error) {
	if spec.Reference {
		return suite.runReferenceWithOpportunity(mask, opportunityMask, counter)
	}
	state, err := suite.newRelevanceState()
	if err != nil {
		return relevanceRunResult{}, err
	}
	result := relevanceRunResult{Purchases: map[string]int64{}, Roles: map[string]RoleActivationCount{}, FinalState: state}
	revision := int64(1)
	random, uuids := NewSplitMix64(seed), NewUUIDStream(seed)
	policySuite := &Suite{Catalog: suite.Catalog, RoutesCatalog: suite.Routes, ConstantsHash: suite.ConstantsHash}
	previousSession := int64(-1)
	for _, offsetMS := range actionTimes(spec.PolicyID, suite.Scenario.HorizonMS) {
		if reached, reachErr := suite.relevanceMilestoneReached(state); reachErr != nil {
			return relevanceRunResult{}, reachErr
		} else if reached {
			copy := offsetMS
			result.MilestoneMS = &copy
			break
		}
		now := relevanceNow(offsetMS)
		mode := production.ModeOnline
		if spec.PolicyID == "casual.phase0" {
			session := casualSession(offsetMS)
			if previousSession >= 0 && session != previousSession {
				mode = production.ModeOffline
			}
			previousSession = session
		}
		intentBytes, _, intentErr := policySuite.intentBytes(spec.PolicyID, state, revision, now, random, uuids)
		if intentErr != nil {
			return relevanceRunResult{}, intentErr
		}
		request, parseErr := production.ParseIntent(intentBytes)
		if parseErr != nil {
			return relevanceRunResult{}, parseErr
		}
		candidate, cloneErr := cloneState(suite.Catalog, state)
		if cloneErr != nil {
			return relevanceRunResult{}, cloneErr
		}
		beforeCounts := make(map[string]int64, len(state.GeneratorCounts))
		for id, count := range state.GeneratorCounts {
			beforeCounts[id] = count
		}
		beforeUpgrades := make(map[string]bool, len(state.UpgradesOwned))
		for id, owned := range state.UpgradesOwned {
			beforeUpgrades[id] = owned
		}
		if err := counter.add(); err != nil {
			return relevanceRunResult{}, err
		}
		simulation, simulationErr := production.SimulateTransition(request, candidate, suite.Catalog,
			production.SimulationDependencies{Routes: suite.Routes}, save.Revision{Number: revision, ConstantsHash: suite.ConstantsHash},
			mode, now, nil, nil, mask)
		if simulationErr != nil {
			return relevanceRunResult{}, simulationErr
		}
		if simulation.Decision.Outcome != save.IntentApplied {
			continue
		}
		state, revision = candidate, revision+1
		for id, count := range state.GeneratorCounts {
			result.Purchases[id] += count - beforeCounts[id]
		}
		for id, owned := range state.UpgradesOwned {
			if owned && !beforeUpgrades[id] {
				result.Purchases[id]++
			}
		}
		mergeRoleActivations(result.Roles, simulation.RoleActivations)
		if reached, reachErr := suite.relevanceMilestoneReached(state); reachErr != nil {
			return relevanceRunResult{}, reachErr
		} else if reached {
			copy := offsetMS
			result.MilestoneMS = &copy
			break
		}
	}
	result.FinalState, result.Transitions = state, counter.value
	if result.MilestoneMS != nil {
		result.FinalVirtualMS = *result.MilestoneMS
	}
	return result, nil
}

func (suite *RelevanceSuite) runReference(mask production.AblationMask, counter *relevanceCounter) (relevanceRunResult, error) {
	opportunityMask, _, err := suite.opportunityAwareMask(production.AblationMask{}, counter)
	if err != nil {
		return relevanceRunResult{}, err
	}
	return suite.runReferenceWithOpportunity(mask, opportunityMask, counter)
}

func (suite *RelevanceSuite) runReferenceWithOpportunity(mask, opportunityMask production.AblationMask, counter *relevanceCounter) (relevanceRunResult, error) {
	effectiveMask := mergeAblationMasks(mask, opportunityMask)
	state, err := suite.newRelevanceState()
	if err != nil {
		return relevanceRunResult{}, err
	}
	result := relevanceRunResult{Purchases: map[string]int64{}, Roles: map[string]RoleActivationCount{}, FinalState: state}
	revision, nowMS := int64(1), int64(0)
	decisionGuardExhausted := true
	for decision := int64(0); decision < suite.Scenario.MaxDecisions; decision++ {
		if reached, reachErr := suite.relevanceMilestoneReached(state); reachErr != nil {
			return relevanceRunResult{}, reachErr
		} else if reached {
			copy := nowMS
			result.MilestoneMS = &copy
			decisionGuardExhausted = false
			break
		}
		if suite.needsReferenceBootstrap(state) {
			manual, manualErr := suite.applyReferenceManual(state, revision, nowMS, effectiveMask, counter)
			if manualErr != nil {
				return relevanceRunResult{}, manualErr
			}
			revision++
			mergeRoleActivations(result.Roles, manual.RoleActivations)
		}
		result.Decisions = decision + 1
		next, ok := suite.nextDecisionHorizon(nowMS)
		if !ok {
			decisionGuardExhausted = false
			break
		}
		candidates, bank, bankAtMS, rankErr := suite.rankDecisionOptions(state, revision, nowMS, next, 1, effectiveMask, counter)
		if rankErr != nil {
			return relevanceRunResult{}, rankErr
		}
		if bank || len(candidates) == 0 {
			advanceAt := next
			if bank {
				advanceAt = bankAtMS
			}
			advanced, advanceErr := suite.advance(state, revision, advanceAt, effectiveMask, counter)
			if advanceErr != nil {
				return relevanceRunResult{}, advanceErr
			}
			mergeRoleActivations(result.Roles, advanced.RoleActivations)
			nowMS = advanceAt
			continue
		}
		chosen := candidates[0]
		state, revision, nowMS = chosen.State, chosen.Revision, chosen.AtMS
		result.Purchases[chosen.ID]++
		mergeRoleActivations(result.Roles, chosen.RoleActivations)
	}
	if result.MilestoneMS == nil {
		if reached, reachErr := suite.relevanceMilestoneReached(state); reachErr != nil {
			return relevanceRunResult{}, reachErr
		} else if reached {
			copy := nowMS
			result.MilestoneMS = &copy
		}
	}
	result.DecisionStarved = result.MilestoneMS == nil && decisionGuardExhausted
	if result.MilestoneMS == nil && !result.DecisionStarved && nowMS < suite.Scenario.HorizonMS {
		finishedMS, reachedMS, activations, advanceErr := suite.finishToMilestone(state, revision, nowMS, effectiveMask, counter)
		if advanceErr != nil {
			return relevanceRunResult{}, advanceErr
		}
		mergeRoleActivations(result.Roles, activations)
		nowMS, result.MilestoneMS = finishedMS, reachedMS
	}
	if result.MilestoneMS == nil {
		if reached, reachErr := suite.relevanceMilestoneReached(state); reachErr != nil {
			return relevanceRunResult{}, reachErr
		} else if reached {
			copy := nowMS
			result.MilestoneMS = &copy
		}
	}
	result.FinalState, result.FinalVirtualMS, result.Transitions = state, nowMS, counter.value
	return result, nil
}

func (suite *RelevanceSuite) rankCandidates(state *save.State, revision, nowMS, horizonMS int64, mask production.AblationMask, counter *relevanceCounter) ([]relevanceCandidate, error) {
	return suite.rankDirectCandidateIDs(state, revision, nowMS, horizonMS, suite.purchasableIDs(), mask, counter)
}

func (suite *RelevanceSuite) rankBeamCandidates(state *save.State, revision, nowMS, horizonMS int64, mask production.AblationMask, counter *relevanceCounter) ([]relevanceCandidate, error) {
	candidates, _, _, err := suite.rankDecisionOptions(state, revision, nowMS, horizonMS,
		suite.Scenario.BeamChildren, mask, counter)
	return candidates, err
}

// rankDecisionOptions orders purchases and the first-class bank option by the
// same projected milestone time. Purchases win equal scores and then tie-break
// by raw-byte ID. bankAtMS is the next declared boundary or the milestone,
// whichever the closed form reaches first.
func (suite *RelevanceSuite) rankDecisionOptions(state *save.State, revision, nowMS, horizonMS, limit int64,
	mask production.AblationMask, counter *relevanceCounter) ([]relevanceCandidate, bool, int64, error) {
	candidates, err := suite.rankCandidates(state, revision, nowMS, horizonMS, mask, counter)
	if err != nil {
		return nil, false, 0, err
	}
	bankProjection, bankReachable, err := suite.projectedMilestone(state, mask)
	if err != nil {
		return nil, false, 0, err
	}
	return selectProjectedDecisionOptions(candidates, bankProjection, bankReachable, nowMS, horizonMS, limit)
}

func selectProjectedDecisionOptions(candidates []relevanceCandidate, bankProjection relevanceProjection, bankReachable bool,
	nowMS, horizonMS, limit int64) ([]relevanceCandidate, bool, int64, error) {
	bankAtMS := horizonMS
	if bankReachable {
		bankMS, ratioErr := ceilDecimalRatio(bankProjection.Numerator, bankProjection.Denominator)
		if ratioErr == nil && bankMS <= relevanceMaxSafeInteger-nowMS && nowMS+bankMS < bankAtMS {
			bankAtMS = nowMS + bankMS
		} else if ratioErr != nil && !errors.Is(ratioErr, errRelevanceRatioOutsideExactDomain) {
			return nil, false, 0, ratioErr
		}
	}
	if limit < 1 {
		return nil, false, 0, errors.New("relevance decision option limit must be positive")
	}
	if !bankReachable {
		if int64(len(candidates)) > limit {
			candidates = candidates[:limit]
		}
		return candidates, false, bankAtMS, nil
	}
	// Purchases win exact ties, so bank follows every purchase with the same
	// projected milestone time.
	bankIndex := sort.Search(len(candidates), func(index int) bool {
		return compareRelevanceProjections(candidates[index].Projection, bankProjection) > 0
	})
	if int64(bankIndex) >= limit {
		return candidates[:limit], false, bankAtMS, nil
	}
	purchaseLimit := limit - 1
	if int64(len(candidates)) > purchaseLimit {
		candidates = candidates[:purchaseLimit]
	}
	return candidates, true, bankAtMS, nil
}

func (suite *RelevanceSuite) rankDirectCandidateIDs(state *save.State, revision, nowMS, horizonMS int64, ids []string, mask production.AblationMask, counter *relevanceCounter) ([]relevanceCandidate, error) {
	result := make([]relevanceCandidate, 0, len(ids))
	for _, id := range ids {
		if state.UpgradesOwned[id] {
			continue
		}
		candidate, ok, err := suite.rankCandidate(state, revision, nowMS, horizonMS, id, mask, counter)
		if err != nil {
			return nil, err
		}
		if ok {
			result = append(result, candidate)
		}
	}
	sortRelevanceCandidates(result)
	return result, nil
}

func sortRelevanceCandidates(result []relevanceCandidate) {
	sort.Slice(result, func(left, right int) bool {
		if comparison := compareRelevanceProjections(result[left].Projection, result[right].Projection); comparison != 0 {
			return comparison < 0
		}
		return result[left].ID < result[right].ID
	})
}

func compareRelevanceProjections(left, right relevanceProjection) int {
	return int(left.Numerator.Mul(right.Denominator).Cmp(right.Numerator.Mul(left.Denominator)))
}

func (suite *RelevanceSuite) rankCandidate(state *save.State, revision, nowMS, horizonMS int64, id string, mask production.AblationMask, counter *relevanceCounter) (relevanceCandidate, bool, error) {
	resourceID, cost, err := suite.candidateCost(state, id)
	if err != nil || resourceID != suite.Scenario.Milestone.ResourceID {
		return relevanceCandidate{}, false, err
	}
	earliest, affordable, err := suite.earliestAffordable(state, revision, nowMS, horizonMS, resourceID, cost, mask, counter)
	if err != nil || !affordable {
		return relevanceCandidate{}, false, err
	}
	baseAtBuy, err := cloneState(suite.Catalog, state)
	if err != nil {
		return relevanceCandidate{}, false, err
	}
	var activations []production.RoleActivation
	if earliest > nowMS {
		advanced, advanceErr := suite.advance(baseAtBuy, revision, earliest, mask, counter)
		if advanceErr != nil {
			return relevanceCandidate{}, false, advanceErr
		}
		activations = append(activations, advanced.RoleActivations...)
	}
	candidateAtBuy, err := cloneState(suite.Catalog, baseAtBuy)
	if err != nil {
		return relevanceCandidate{}, false, err
	}
	request := candidateIntent(id, revision)
	if err := counter.add(); err != nil {
		return relevanceCandidate{}, false, err
	}
	simulation, err := production.SimulateTransition(request, candidateAtBuy, suite.Catalog,
		production.SimulationDependencies{Routes: suite.Routes}, save.Revision{Number: revision, ConstantsHash: suite.ConstantsHash},
		production.ModeOnline, relevanceNow(earliest), nil, nil, mask)
	if err != nil || simulation.Decision.Outcome != save.IntentApplied {
		return relevanceCandidate{}, false, err
	}
	activations = append(activations, simulation.RoleActivations...)
	projection, reachable, err := suite.projectedMilestone(candidateAtBuy, mask)
	if err != nil {
		return relevanceCandidate{}, false, fmt.Errorf("relevance candidate %s projected milestone: %w", id, err)
	}
	if !reachable {
		return relevanceCandidate{}, false, nil
	}
	waitMS := earliest - nowMS
	if waitMS > 0 {
		projection.Numerator = projection.Numerator.Add(projection.Denominator.Mul(decimal.FromString(strconv.FormatInt(waitMS, 10))))
	}
	return relevanceCandidate{ID: id, Projection: projection,
		AtMS: earliest, State: candidateAtBuy, Revision: revision + 1,
		RoleActivations: activations}, true, nil
}

// projectedMilestone evaluates the T01-C20 closed form as a comparison ratio
// against the production engine's canonical masked rate. The rate is per
// second, hence the exact ×1000 numerator for millisecond scores.
func (suite *RelevanceSuite) projectedMilestone(state *save.State, mask production.AblationMask) (relevanceProjection, bool, error) {
	balance, ok := state.Ledger.Balance(suite.Scenario.Milestone.ResourceID)
	target, err := decimal.ParseCanonical(suite.Scenario.Milestone.Amount)
	if err != nil || !ok {
		return relevanceProjection{}, false, err
	}
	if balance.Gte(target) {
		return relevanceProjection{Numerator: decimal.Zero, Denominator: decimal.One}, true, nil
	}
	rate, err := production.SimulateResourceRate(state, suite.Catalog, suite.Scenario.Milestone.ResourceID, mask)
	if err != nil {
		return relevanceProjection{}, false, err
	}
	if !rate.Gt(decimal.Zero) {
		return relevanceProjection{}, false, nil
	}
	return relevanceProjection{Numerator: target.Sub(balance).Mul(decimal.New(1, 3)), Denominator: rate}, true, nil
}

func (suite *RelevanceSuite) earliestAffordable(state *save.State, revision, nowMS, horizonMS int64, resourceID string, cost decimal.Decimal, mask production.AblationMask, counter *relevanceCounter) (int64, bool, error) {
	balance, _ := state.Ledger.Balance(resourceID)
	if balance.Gte(cost) {
		return nowMS, true, nil
	}
	affordableAt := func(at int64) (bool, error) {
		candidate, err := cloneState(suite.Catalog, state)
		if err != nil {
			return false, err
		}
		if _, err := suite.advance(candidate, revision, at, mask, counter); err != nil {
			return false, err
		}
		value, _ := candidate.Ledger.Balance(resourceID)
		return value.Gte(cost), nil
	}
	ok, err := affordableAt(horizonMS)
	if err != nil || !ok {
		return 0, false, err
	}
	low, high := nowMS+1, horizonMS
	for low < high {
		middle := low + (high-low)/2
		ok, err := affordableAt(middle)
		if err != nil {
			return 0, false, err
		}
		if ok {
			high = middle
		} else {
			low = middle + 1
		}
	}
	return low, true, nil
}

func (suite *RelevanceSuite) advance(state *save.State, revision, atMS int64, mask production.AblationMask, counter *relevanceCounter) (production.AdvanceSimulationResult, error) {
	if err := counter.add(); err != nil {
		return production.AdvanceSimulationResult{}, err
	}
	return production.SimulateAdvance(state, suite.Catalog, production.SimulationDependencies{Routes: suite.Routes},
		save.Revision{Number: revision, ConstantsHash: suite.ConstantsHash}, production.ModeOnline, relevanceNow(atMS), nil, mask)
}

func (suite *RelevanceSuite) candidateCost(state *save.State, id string) (string, decimal.Decimal, error) {
	if generator, ok := suite.Catalog.GeneratorClass(id); ok {
		cost, err := suite.Catalog.BulkCost(id, state.GeneratorCounts[id], 1)
		return generator.Price.ResourceID, cost, err
	}
	if upgrade, ok := suite.Catalog.Upgrade(id); ok {
		return upgrade.Cost.ResourceID, upgrade.Cost.Amount, nil
	}
	return "", decimal.NaN, errors.New("unknown relevance candidate")
}

func candidateIntent(id string, revision int64) production.IntentRequest {
	intentID := relevanceIntentID(revision)
	if len(id) >= len("generator.") && id[:len("generator.")] == "generator." {
		return production.IntentRequest{IntentID: intentID, Kind: production.IntentBuyGenerator, ExpectedRevision: revision,
			GeneratorID: id, CountMode: "exact", Count: 1}
	}
	return production.IntentRequest{IntentID: intentID, Kind: production.IntentBuyUpgrade, ExpectedRevision: revision, UpgradeID: id}
}

type relevanceManualResult struct {
	Applied         int64
	RoleActivations []production.RoleActivation
}

func (suite *RelevanceSuite) applyReferenceManual(state *save.State, revision, nowMS int64, mask production.AblationMask, counter *relevanceCounter) (relevanceManualResult, error) {
	actions := suite.Catalog.ManualActions()
	sort.Slice(actions, func(left, right int) bool { return actions[left].ID < actions[right].ID })
	matching := make([]economy.ManualActionDefinition, 0, 1)
	for _, action := range actions {
		if action.Output.ResourceID == suite.Scenario.Milestone.ResourceID {
			matching = append(matching, action)
		}
	}
	if len(matching) != 1 {
		return relevanceManualResult{}, fmt.Errorf("relevance reference requires exactly one milestone-resource manual action, got %d", len(matching))
	}
	policy := suite.Catalog.ManualPolicy()
	maximumCount := policy.BucketCapMilli / 1000
	balance, _ := state.Ledger.Balance(suite.Scenario.Milestone.ResourceID)
	var cheapest decimal.Decimal
	foundCheapest := false
	for _, id := range suite.purchasableIDs() {
		resourceID, cost, costErr := suite.candidateCost(state, id)
		if costErr != nil {
			return relevanceManualResult{}, costErr
		}
		if resourceID == suite.Scenario.Milestone.ResourceID && cost.Gt(balance) && (!foundCheapest || cost.Lt(cheapest)) {
			cheapest = cost
			foundCheapest = true
		}
	}
	if !foundCheapest {
		return relevanceManualResult{}, errors.New("relevance reference has no milestone-resource bootstrap purchase")
	}
	count, countErr := ceilDecimalRatio(cheapest.Sub(balance), matching[0].Output.AmountPerAction)
	if countErr != nil {
		return relevanceManualResult{}, fmt.Errorf("relevance reference manual bootstrap count: %w", countErr)
	}
	if count > maximumCount {
		count = maximumCount
	}
	if count < 1 || policy.RefillMilliPerMS < 1 || count > relevanceMaxSafeInteger/1000 {
		return relevanceManualResult{}, errors.New("relevance reference manual cadence is invalid")
	}
	windowMS := (count*1000 + policy.RefillMilliPerMS - 1) / policy.RefillMilliPerMS
	request := production.IntentRequest{IntentID: relevanceIntentID(revision), Kind: production.IntentPerformManualBatch,
		ExpectedRevision: revision, ActionID: matching[0].ID, Count: count, WindowMS: windowMS}
	if err := counter.add(); err != nil {
		return relevanceManualResult{}, err
	}
	simulation, err := production.SimulateTransition(request, state, suite.Catalog,
		production.SimulationDependencies{Routes: suite.Routes}, save.Revision{Number: revision, ConstantsHash: suite.ConstantsHash},
		production.ModeOnline, relevanceNow(nowMS), nil, nil, mask)
	if err != nil {
		return relevanceManualResult{}, err
	}
	if simulation.Decision.Outcome != save.IntentApplied {
		return relevanceManualResult{}, errors.New("relevance reference manual action rejected")
	}
	var receipt struct {
		AppliedCount int64 `json:"applied_count"`
	}
	if err := json.Unmarshal(simulation.Decision.Receipt, &receipt); err != nil || receipt.AppliedCount < 0 || receipt.AppliedCount > count {
		return relevanceManualResult{}, errors.New("relevance reference manual receipt is invalid")
	}
	return relevanceManualResult{Applied: receipt.AppliedCount, RoleActivations: simulation.RoleActivations}, nil
}

func (suite *RelevanceSuite) needsReferenceBootstrap(state *save.State) bool {
	for _, generator := range suite.Catalog.GeneratorClassesForScope(economy.ScopeCompany) {
		if state.GeneratorCounts[generator.ID] > 0 {
			return false
		}
	}
	for _, upgrade := range suite.Catalog.Upgrades() {
		if state.UpgradesOwned[upgrade.ID] {
			return false
		}
	}
	return true
}

func relevanceIntentID(revision int64) string {
	intentID := fmt.Sprintf("00000000-0000-7000-8000-%012x", revision)
	if len(intentID) > 36 {
		return "00000000-0000-7000-8000-000000000001"
	}
	return intentID
}

func (suite *RelevanceSuite) newRelevanceState() (*save.State, error) {
	ledger, err := economy.NewLedger(suite.Catalog, economy.ScopeCompany)
	if err != nil {
		return nil, err
	}
	counts, provisioned, remainders := map[string]int64{}, map[string]int64{}, map[string]int64{}
	for _, generator := range suite.Catalog.GeneratorClassesForScope(economy.ScopeCompany) {
		counts[generator.ID], provisioned[generator.ID] = 0, 0
		if generator.Provision != nil {
			remainders[generator.Provision.GeneratorID] = 0
		}
	}
	return &save.State{Ledger: ledger, GeneratorCounts: counts, UpgradesOwned: map[string]bool{}, GeneratorProvisioned: provisioned,
		ProvisionRemaindersPPM: remainders, EvaluatedThrough: Epoch, RunStartedAt: Epoch, RunSeq: 1,
		ManualTokenMilli: suite.Catalog.ManualPolicy().BucketCapMilli, ManualTokenRefilledAt: Epoch,
		GatesCrossed: map[string]bool{}, DoctrinesByTransition: map[string]string{}, LedgerFactKinds: map[string]bool{},
		MeterBands: map[string]int{}, RegionTraits: map[string]bool{}, HintsUnlocked: map[string]bool{}}, nil
}

func (suite *RelevanceSuite) relevanceMilestoneReached(state *save.State) (bool, error) {
	value, ok := state.Ledger.Balance(suite.Scenario.Milestone.ResourceID)
	target, err := decimal.ParseCanonical(suite.Scenario.Milestone.Amount)
	return ok && err == nil && value.Gte(target), err
}

func (suite *RelevanceSuite) nextDecisionHorizon(nowMS int64) (int64, bool) {
	index := sort.Search(len(suite.Scenario.DecisionHorizonsMS), func(index int) bool { return suite.Scenario.DecisionHorizonsMS[index] > nowMS })
	if index == len(suite.Scenario.DecisionHorizonsMS) {
		return 0, false
	}
	return suite.Scenario.DecisionHorizonsMS[index], true
}

func (suite *RelevanceSuite) purchasableIDs() []string {
	result := make([]string, 0, len(suite.Policy.Items))
	for _, item := range suite.Policy.Items {
		result = append(result, item.PurchasableID)
	}
	return result
}

func effectMask(id string, catalog *economy.Catalog) production.AblationMask {
	if _, ok := catalog.GeneratorClass(id); ok {
		return production.AblationMask{GeneratorIDs: []string{id}}
	}
	return production.AblationMask{UpgradeIDs: []string{id}}
}

func removalMask(id string, catalog *economy.Catalog) production.AblationMask {
	if _, ok := catalog.GeneratorClass(id); ok {
		return production.AblationMask{RemovedGeneratorIDs: []string{id}}
	}
	return production.AblationMask{RemovedUpgradeIDs: []string{id}}
}

func cloneAblationMask(mask production.AblationMask) production.AblationMask {
	return production.AblationMask{
		GeneratorIDs:        append([]string(nil), mask.GeneratorIDs...),
		UpgradeIDs:          append([]string(nil), mask.UpgradeIDs...),
		RemovedGeneratorIDs: append([]string(nil), mask.RemovedGeneratorIDs...),
		RemovedUpgradeIDs:   append([]string(nil), mask.RemovedUpgradeIDs...),
		RemovedActionIDs:    append([]string(nil), mask.RemovedActionIDs...),
	}
}

func mergeAblationMasks(left, right production.AblationMask) production.AblationMask {
	merged := cloneAblationMask(left)
	merged.GeneratorIDs = appendUniqueStrings(merged.GeneratorIDs, right.GeneratorIDs...)
	merged.UpgradeIDs = appendUniqueStrings(merged.UpgradeIDs, right.UpgradeIDs...)
	merged.RemovedGeneratorIDs = appendUniqueStrings(merged.RemovedGeneratorIDs, right.RemovedGeneratorIDs...)
	merged.RemovedUpgradeIDs = appendUniqueStrings(merged.RemovedUpgradeIDs, right.RemovedUpgradeIDs...)
	merged.RemovedActionIDs = appendUniqueStrings(merged.RemovedActionIDs, right.RemovedActionIDs...)
	return merged
}

func appendUniqueStrings(values []string, additions ...string) []string {
	for _, addition := range additions {
		found := false
		for _, value := range values {
			if value == addition {
				found = true
				break
			}
		}
		if !found {
			values = append(values, addition)
		}
	}
	return values
}

func addRemoval(mask *production.AblationMask, id string, catalog *economy.Catalog) {
	if _, ok := catalog.GeneratorClass(id); ok {
		mask.RemovedGeneratorIDs = append(mask.RemovedGeneratorIDs, id)
		return
	}
	mask.RemovedUpgradeIDs = append(mask.RemovedUpgradeIDs, id)
}

func maskRemoves(mask production.AblationMask, id string, catalog *economy.Catalog) bool {
	values := mask.RemovedUpgradeIDs
	if _, ok := catalog.GeneratorClass(id); ok {
		values = mask.RemovedGeneratorIDs
	}
	for _, value := range values {
		if value == id {
			return true
		}
	}
	return false
}

// opportunityAwareMask evaluates the reference payback trajectory against the
// counterfactual use of its cash and finite decision slots. A purchase whose
// removal makes the declared milestone earlier is excluded, then the screen is
// repeated to a monotonic fixed point. Both the reference arm and beam rollouts
// use the resulting mask before applying their shared payback ordering.
func (suite *RelevanceSuite) opportunityAwareMask(base production.AblationMask, counter *relevanceCounter) (production.AblationMask, []relevanceOpportunityDecision, error) {
	effective := cloneAblationMask(base)
	diagnostics := []relevanceOpportunityDecision{}
	for {
		baseline, err := suite.runDirectFromGenesis(effective, counter)
		if err != nil {
			return production.AblationMask{}, nil, err
		}
		if baseline.DecisionStarved {
			return production.AblationMask{}, nil, errReferenceDecisionStarved
		}
		changed := false
		for _, id := range suite.purchasableIDs() {
			if baseline.Purchases[id] == 0 || maskRemoves(effective, id, suite.Catalog) {
				continue
			}
			removed := cloneAblationMask(effective)
			addRemoval(&removed, id, suite.Catalog)
			without, err := suite.runDirectFromGenesis(removed, counter)
			if err != nil {
				return production.AblationMask{}, nil, err
			}
			if without.DecisionStarved {
				return production.AblationMask{}, nil, errReferenceDecisionStarved
			}
			diagnostics = append(diagnostics, relevanceOpportunityDecision{PurchasableID: id,
				BaselineMS: cloneInt64(baseline.MilestoneMS), RemovedMS: cloneInt64(without.MilestoneMS)})
			if without.MilestoneMS != nil && (baseline.MilestoneMS == nil || *without.MilestoneMS < *baseline.MilestoneMS) {
				effective = removed
				changed = true
				break
			}
		}
		if !changed {
			return effective, diagnostics, nil
		}
	}
}

func (suite *RelevanceSuite) runDirectFromGenesis(mask production.AblationMask, counter *relevanceCounter) (relevanceRunResult, error) {
	state, err := suite.newRelevanceState()
	if err != nil {
		return relevanceRunResult{}, err
	}
	return suite.runReferenceFromRanked(state, 1, 0, suite.Scenario.MaxDecisions, mask, counter)
}

func groupMask(ids []string, catalog *economy.Catalog) production.AblationMask {
	mask := production.AblationMask{}
	for _, id := range ids {
		if _, ok := catalog.GeneratorClass(id); ok {
			mask.GeneratorIDs = append(mask.GeneratorIDs, id)
		} else {
			mask.UpgradeIDs = append(mask.UpgradeIDs, id)
		}
	}
	return mask
}

func mergeRoleActivations(destination map[string]RoleActivationCount, values []production.RoleActivation) {
	for _, value := range values {
		key := value.GeneratorID + "\x00" + string(value.Kind) + "\x00" + value.TargetID
		entry := destination[key]
		entry.GeneratorID, entry.Kind, entry.TargetID, entry.Count = value.GeneratorID, value.Kind, value.TargetID, entry.Count+1
		destination[key] = entry
	}
}

func policyGroup(policy *RelevancePolicy, id string) RelevancePolicyGroup {
	index := sort.Search(len(policy.Groups), func(index int) bool { return policy.Groups[index].GroupID >= id })
	if index < len(policy.Groups) && policy.Groups[index].GroupID == id {
		return policy.Groups[index]
	}
	return RelevancePolicyGroup{}
}

func appendRelevancePair(destination map[string]map[string][]relevancePairedResult, targetID, policyID string, baseline, ablated *int64, baselinePurchases int64) {
	if destination[targetID] == nil {
		destination[targetID] = map[string][]relevancePairedResult{}
	}
	destination[targetID][policyID] = append(destination[targetID][policyID], relevancePairedResult{
		baseline: cloneInt64(baseline), ablated: cloneInt64(ablated), baselinePurchases: baselinePurchases,
	})
}

// reduceRelevancePairMatrix applies the scenario's conservative reducer inside
// each persona first, then the ANY gate selects the strongest persona result.
func reduceRelevancePairMatrix(byPolicy map[string][]relevancePairedResult, reducer, milestoneID string, horizonMS int64) (RelevanceDelta, []string, error) {
	if len(byPolicy) == 0 {
		return RelevanceDelta{}, nil, errors.New("missing relevance pair matrix")
	}
	policies := make([]string, 0, len(byPolicy))
	for policyID := range byPolicy {
		policies = append(policies, policyID)
	}
	sort.Strings(policies)
	var selected *RelevanceDelta
	var fallback *RelevanceDelta
	excluded := []string{}
	for _, policyID := range policies {
		pairs := make([]relevancePairedResult, 0, len(byPolicy[policyID]))
		for _, pair := range byPolicy[policyID] {
			if pair.baselinePurchases > 0 {
				pairs = append(pairs, pair)
			}
		}
		if len(pairs) == 0 {
			excluded = append(excluded, policyID)
			continue
		}
		rows := make([]RelevanceDelta, 0, len(pairs))
		for _, pair := range pairs {
			rows = append(rows, MakeRelevanceDelta(milestoneID, pair.baseline, pair.ablated, horizonMS))
		}
		reduced, err := reduceRelevanceDeltas(rows, reducer)
		if err != nil {
			return RelevanceDelta{}, nil, err
		}
		if fallback == nil {
			copy := reduced
			fallback = &copy
		}
		if reduced.DeltaMS != nil && (selected == nil || *reduced.DeltaMS > *selected.DeltaMS) {
			copy := reduced
			selected = &copy
		}
	}
	if selected != nil {
		return *selected, excluded, nil
	}
	if fallback != nil {
		return *fallback, excluded, nil
	}
	return RelevanceDelta{MilestoneID: milestoneID, Status: "both_unreached"}, excluded, nil
}

func reduceRelevanceCounts(values []int64, reducer string) int64 {
	if len(values) == 0 {
		return 0
	}
	ordered := append([]int64(nil), values...)
	sort.Slice(ordered, func(left, right int) bool { return ordered[left] < ordered[right] })
	index := 0
	if reducer == "p05" {
		index = (len(ordered) - 1) * 5 / 100
	}
	return ordered[index]
}

func reduceRelevancePairPurchases(byPolicy map[string][]relevancePairedResult, reducer string) int64 {
	result := int64(0)
	for _, pairs := range byPolicy {
		values := make([]int64, 0, len(pairs))
		for _, pair := range pairs {
			if pair.baselinePurchases > 0 {
				values = append(values, pair.baselinePurchases)
			}
		}
		if reduced := reduceRelevanceCounts(values, reducer); reduced > result {
			result = reduced
		}
	}
	return result
}

func reduceRelevanceRoles(byPolicy map[string][]map[string]RoleActivationCount, reducer string) (map[string]RoleActivationCount, error) {
	result := map[string]RoleActivationCount{}
	for _, runs := range byPolicy {
		keys := map[string]RoleActivationCount{}
		for _, run := range runs {
			for key, role := range run {
				keys[key] = role
			}
		}
		for key, identity := range keys {
			values := make([]int64, 0, len(runs))
			for _, run := range runs {
				values = append(values, run[key].Count)
			}
			entry := result[key]
			entry.GeneratorID, entry.Kind, entry.TargetID = identity.GeneratorID, identity.Kind, identity.TargetID
			count := reduceRelevanceCounts(values, reducer)
			if count > relevanceMaxSafeInteger-entry.Count {
				return nil, errors.New("relevance role count overflow")
			}
			entry.Count += count
			result[key] = entry
		}
	}
	for key, role := range result {
		if role.Count == 0 {
			delete(result, key)
		}
	}
	return result, nil
}

func sortedUniqueStrings(values []string) []string {
	sort.Strings(values)
	result := values[:0]
	for _, value := range values {
		if len(result) == 0 || result[len(result)-1] != value {
			result = append(result, value)
		}
	}
	return result
}

func parseSeed(value string) (uint64, error) { return strconv.ParseUint(value, 10, 64) }

func stateDigest(state *save.State) (string, error) {
	data, err := save.EncodeState(state)
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256(data)
	return hex.EncodeToString(digest[:]), nil
}

// runBeam performs the ruled declared-width search. Each child is scored by
// the same T01-C20 projected-milestone metric as the reference arm; equal-state
// nodes deduplicate on state hash + virtual time, with raw-byte paths last.
func (suite *RelevanceSuite) runBeam(counter *relevanceCounter) (*int64, error) {
	effectiveMask, _, err := suite.opportunityAwareMask(production.AblationMask{}, counter)
	if err != nil {
		return nil, err
	}
	return suite.runBeamWithOpportunity(effectiveMask, counter)
}

func (suite *RelevanceSuite) runBeamWithOpportunity(effectiveMask production.AblationMask, counter *relevanceCounter) (*int64, error) {
	type node struct {
		state          *save.State
		revision       int64
		atMS           int64
		path           string
		reached        *int64
		score          relevanceProjection
		scoreReachable bool
	}
	initial, err := suite.newRelevanceState()
	if err != nil {
		return nil, err
	}
	frontier := []node{{state: initial, revision: 1}}
	var best *int64
	for depth := int64(0); depth < suite.Scenario.MaxDecisions && len(frontier) > 0; depth++ {
		children := []node{}
		for _, current := range frontier {
			if reached, _ := suite.relevanceMilestoneReached(current.state); reached {
				copy := current.atMS
				if best == nil || copy < *best {
					best = &copy
				}
				continue
			}
			if suite.needsReferenceBootstrap(current.state) {
				manual, manualErr := suite.applyReferenceManual(current.state, current.revision, current.atMS, effectiveMask, counter)
				if manualErr != nil {
					return nil, manualErr
				}
				current.revision++
				current.path += fmt.Sprintf("/~manual:%d", manual.Applied)
			}
			next, ok := suite.nextDecisionHorizon(current.atMS)
			if !ok {
				continue
			}
			candidates, bank, bankAtMS, rankErr := suite.rankDecisionOptions(current.state, current.revision,
				current.atMS, next, suite.Scenario.BeamChildren, effectiveMask, counter)
			if rankErr != nil {
				return nil, rankErr
			}
			for _, candidate := range candidates {
				children = append(children, node{state: candidate.State, revision: candidate.Revision, atMS: candidate.AtMS,
					path: current.path + "/" + candidate.ID})
			}
			if bank {
				waitState, cloneErr := cloneState(suite.Catalog, current.state)
				if cloneErr != nil {
					return nil, cloneErr
				}
				if _, advanceErr := suite.advance(waitState, current.revision, bankAtMS, effectiveMask, counter); advanceErr != nil {
					return nil, advanceErr
				}
				children = append(children, node{state: waitState, revision: current.revision, atMS: bankAtMS, path: current.path + "/~bank"})
			}
		}
		pruned := make([]node, 0, len(children))
		for index := range children {
			dominated := false
			for other := range children {
				if index != other && children[index].atMS == children[other].atMS &&
					relevanceStateDominates(children[other].state, children[index].state, suite.Catalog) {
					dominated = true
					break
				}
			}
			if !dominated {
				pruned = append(pruned, children[index])
			}
		}
		// R11 defines node identity as canonical state hash + virtual time.
		// Collapse identical children before scoring them; identical nodes have
		// identical projections, so the raw-byte path tie-break stays canonical.
		dedup := map[string]node{}
		for index := range pruned {
			child := pruned[index]
			digest, digestErr := stateDigest(child.state)
			if digestErr != nil {
				return nil, digestErr
			}
			key := fmt.Sprintf("%s:%d", digest, child.atMS)
			prior, exists := dedup[key]
			if !exists || child.path < prior.path {
				dedup[key] = child
			}
		}
		children = children[:0]
		for _, child := range dedup {
			children = append(children, child)
		}
		sort.Slice(children, func(left, right int) bool { return children[left].path < children[right].path })
		for index := range children {
			child := &children[index]
			if reached, reachErr := suite.relevanceMilestoneReached(child.state); reachErr != nil {
				return nil, reachErr
			} else if reached {
				copy := child.atMS
				child.reached, child.scoreReachable = &copy, true
				child.score = relevanceProjection{Numerator: decimal.Zero, Denominator: decimal.One}
				if best == nil || copy < *best {
					best = &copy
				}
				continue
			}
			projection, reachable, projectionErr := suite.projectedMilestone(child.state, effectiveMask)
			if projectionErr != nil {
				return nil, projectionErr
			}
			child.score, child.scoreReachable = projection, reachable
			if reachable && child.atMS > 0 {
				child.score.Numerator = child.score.Numerator.Add(
					child.score.Denominator.Mul(decimal.FromString(strconv.FormatInt(child.atMS, 10))))
			}
		}
		frontier = frontier[:0]
		for _, child := range children {
			frontier = append(frontier, child)
		}
		sort.Slice(frontier, func(left, right int) bool {
			if frontier[left].scoreReachable != frontier[right].scoreReachable {
				return frontier[left].scoreReachable
			}
			if frontier[left].scoreReachable {
				if comparison := compareRelevanceProjections(frontier[left].score, frontier[right].score); comparison != 0 {
					return comparison < 0
				}
			}
			return frontier[left].path < frontier[right].path
		})
		if int64(len(frontier)) > suite.Scenario.BeamWidth {
			frontier = frontier[:suite.Scenario.BeamWidth]
		}
	}
	return best, nil
}

// relevanceStateDominates implements R11's componentwise relation at an equal
// virtual time. A strict improvement is required; otherwise equal states are
// left to the canonical digest deduplicator. Provisioned counts and carried
// remainders are included because they affect future production even though
// the public dominance description abbreviates them as milestone progress.
func relevanceStateDominates(left, right *save.State, catalog *economy.Catalog) bool {
	strict := false
	for _, resource := range catalog.Resources() {
		leftValue, leftOK := left.Ledger.Balance(resource.ID)
		rightValue, rightOK := right.Ledger.Balance(resource.ID)
		if leftOK != rightOK || leftOK && leftValue.Lt(rightValue) {
			return false
		}
		if leftOK && leftValue.Gt(rightValue) {
			strict = true
		}
	}
	for _, generator := range catalog.GeneratorClassesForScope(economy.ScopeCompany) {
		if left.GeneratorCounts[generator.ID] < right.GeneratorCounts[generator.ID] ||
			left.GeneratorProvisioned[generator.ID] < right.GeneratorProvisioned[generator.ID] {
			return false
		}
		if left.GeneratorCounts[generator.ID] > right.GeneratorCounts[generator.ID] ||
			left.GeneratorProvisioned[generator.ID] > right.GeneratorProvisioned[generator.ID] {
			strict = true
		}
	}
	for _, upgrade := range catalog.Upgrades() {
		if right.UpgradesOwned[upgrade.ID] && !left.UpgradesOwned[upgrade.ID] {
			return false
		}
		if left.UpgradesOwned[upgrade.ID] && !right.UpgradesOwned[upgrade.ID] {
			strict = true
		}
	}
	for key, rightValue := range right.ProvisionRemaindersPPM {
		leftValue := left.ProvisionRemaindersPPM[key]
		if leftValue < rightValue {
			return false
		}
		if leftValue > rightValue {
			strict = true
		}
	}
	return strict
}

func (suite *RelevanceSuite) runReferenceFrom(state *save.State, revision, nowMS, maxDecisions int64, counter *relevanceCounter) (relevanceRunResult, error) {
	return suite.runReferenceFromRanked(state, revision, nowMS, maxDecisions, production.AblationMask{}, counter)
}

func (suite *RelevanceSuite) runReferenceFromRanked(state *save.State, revision, nowMS, maxDecisions int64, mask production.AblationMask, counter *relevanceCounter) (relevanceRunResult, error) {
	return suite.runRankedCompletion(state, revision, nowMS, maxDecisions, mask, counter)
}

func (suite *RelevanceSuite) runRankedCompletion(state *save.State, revision, nowMS, maxDecisions int64, mask production.AblationMask, counter *relevanceCounter) (relevanceRunResult, error) {
	result := relevanceRunResult{Purchases: map[string]int64{}, Roles: map[string]RoleActivationCount{}, FinalState: state}
	decisionGuardExhausted := true
	for decision := int64(0); decision < maxDecisions; decision++ {
		if reached, err := suite.relevanceMilestoneReached(state); err != nil {
			return relevanceRunResult{}, err
		} else if reached {
			copy := nowMS
			result.MilestoneMS = &copy
			decisionGuardExhausted = false
			break
		}
		if suite.needsReferenceBootstrap(state) {
			manual, err := suite.applyReferenceManual(state, revision, nowMS, mask, counter)
			if err != nil {
				return relevanceRunResult{}, err
			}
			revision++
			mergeRoleActivations(result.Roles, manual.RoleActivations)
		}
		result.Decisions = decision + 1
		next, ok := suite.nextDecisionHorizon(nowMS)
		if !ok {
			decisionGuardExhausted = false
			break
		}
		candidates, bank, bankAtMS, err := suite.rankDecisionOptions(state, revision, nowMS, next, 1, mask, counter)
		if err != nil {
			return relevanceRunResult{}, err
		}
		if bank || len(candidates) == 0 {
			advanceAt := next
			if bank {
				advanceAt = bankAtMS
			}
			if _, err := suite.advance(state, revision, advanceAt, mask, counter); err != nil {
				return relevanceRunResult{}, err
			}
			nowMS = advanceAt
			continue
		}
		chosen := candidates[0]
		state, revision, nowMS = chosen.State, chosen.Revision, chosen.AtMS
		result.Purchases[chosen.ID]++
		mergeRoleActivations(result.Roles, chosen.RoleActivations)
	}
	if result.MilestoneMS == nil {
		if reached, reachErr := suite.relevanceMilestoneReached(state); reachErr != nil {
			return relevanceRunResult{}, reachErr
		} else if reached {
			copy := nowMS
			result.MilestoneMS = &copy
		}
	}
	result.DecisionStarved = result.MilestoneMS == nil && decisionGuardExhausted
	if result.MilestoneMS == nil && !result.DecisionStarved && nowMS < suite.Scenario.HorizonMS {
		finishedMS, reachedMS, activations, err := suite.finishToMilestone(state, revision, nowMS, mask, counter)
		if err != nil {
			return relevanceRunResult{}, err
		}
		mergeRoleActivations(result.Roles, activations)
		nowMS, result.MilestoneMS = finishedMS, reachedMS
	}
	result.FinalState, result.FinalVirtualMS, result.Transitions = state, nowMS, counter.value
	return result, nil
}

func (suite *RelevanceSuite) finishToMilestone(state *save.State, revision, nowMS int64, mask production.AblationMask, counter *relevanceCounter) (int64, *int64, []production.RoleActivation, error) {
	reachedAt := func(atMS int64) (bool, error) {
		candidate, err := cloneState(suite.Catalog, state)
		if err != nil {
			return false, err
		}
		if _, err := suite.advance(candidate, revision, atMS, mask, counter); err != nil {
			return false, err
		}
		return suite.relevanceMilestoneReached(candidate)
	}
	reached, err := reachedAt(suite.Scenario.HorizonMS)
	if err != nil {
		return 0, nil, nil, err
	}
	target := suite.Scenario.HorizonMS
	if reached {
		low, high := nowMS+1, suite.Scenario.HorizonMS
		for low < high {
			middle := low + (high-low)/2
			at, searchErr := reachedAt(middle)
			if searchErr != nil {
				return 0, nil, nil, searchErr
			}
			if at {
				high = middle
			} else {
				low = middle + 1
			}
		}
		target = low
	}
	advanced, err := suite.advance(state, revision, target, mask, counter)
	if err != nil {
		return 0, nil, nil, err
	}
	if !reached {
		return target, nil, advanced.RoleActivations, nil
	}
	copy := target
	return target, &copy, advanced.RoleActivations, nil
}

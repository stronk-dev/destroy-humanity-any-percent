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
	ID                      string
	PaybackMS               int64
	EarliestPositiveDeltaMS int64
	AtMS                    int64
	State                   *save.State
	Revision                int64
	RoleActivations         []production.RoleActivation
}

type relevanceCounter struct {
	value int64
	limit int64
}

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
			factor := int64(1 + len(suite.Policy.Items) + len(suite.Policy.Groups))
			if actions != 0 && run.SeedCount > relevanceMaxSafeInteger/actions ||
				actions*run.SeedCount != 0 && factor > relevanceMaxSafeInteger/(actions*run.SeedCount) ||
				nonReferenceTransitions > relevanceMaxSafeInteger-actions*run.SeedCount*factor {
				return RelevanceReport{}, errors.New("relevance transition preflight overflow")
			}
			nonReferenceTransitions += actions * run.SeedCount * factor
		}
	}
	declaredRuns, err := ComputeRelevanceRunBudget(nonReferenceSeeds, referenceSeeds, int64(len(suite.Policy.Items)), int64(len(suite.Policy.Groups)), true)
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
	counter := &relevanceCounter{limit: suite.Scenario.RelevanceBudgetMaxTransitions}
	report := RelevanceReport{SchemaVersion: suite.Scenario.SchemaVersion, ScenarioID: suite.Scenario.ID,
		ScenarioHash: suite.ScenarioHash, ConstantsHash: suite.ConstantsHash, RelevancePolicyHash: suite.Policy.Hash,
		Items: []RelevanceItemReport{}, Groups: []RelevanceGroupReport{}, TierContributions: []RelevanceTierContribution{},
		RoleActivations: []RoleActivationCount{}, Failures: []string{}}
	itemPairs := map[string]map[string][]relevancePairedResult{}
	removalPairs := map[string]map[string][]relevancePairedResult{}
	groupPairs := map[string]map[string][]relevancePairedResult{}
	baselinePurchases := map[string]int64{}
	referencePurchaseSamples := map[string][]int64{}
	baselineRoleRuns := map[string][]map[string]RoleActivationCount{}
	executedRuns := int64(0)
	type oraclePair struct{ greedy, beam *int64 }
	oraclePairs := []oraclePair{}
	baselineUnreached := false
	for _, spec := range suite.Scenario.Runs {
		start, _ := parseSeed(spec.SeedStart)
		for offset := int64(0); offset < spec.SeedCount; offset++ {
			seed := start + uint64(offset)
			baseline, runErr := suite.runPersona(spec, seed, production.AblationMask{}, counter)
			if runErr != nil {
				return RelevanceReport{}, fmt.Errorf("relevance run %s seed %d baseline: %w", spec.PolicyID, seed, runErr)
			}
			executedRuns++
			if baseline.MilestoneMS == nil {
				baselineUnreached = true
			}
			if spec.Reference {
				for _, item := range suite.Policy.Items {
					referencePurchaseSamples[item.PurchasableID] = append(referencePurchaseSamples[item.PurchasableID], baseline.Purchases[item.PurchasableID])
				}
			}
			baselineRoleRuns[spec.PolicyID] = append(baselineRoleRuns[spec.PolicyID], baseline.Roles)
			for _, item := range suite.Policy.Items {
				masked, runErr := suite.runPersona(spec, seed, effectMask(item.PurchasableID, suite.Catalog), counter)
				if runErr != nil {
					return RelevanceReport{}, fmt.Errorf("relevance run %s seed %d effect mask %s: %w", spec.PolicyID, seed, item.PurchasableID, runErr)
				}
				executedRuns++
				appendRelevancePair(itemPairs, item.PurchasableID, spec.PolicyID, baseline.MilestoneMS, masked.MilestoneMS)
				if spec.Reference {
					removed, removeErr := suite.runPersona(spec, seed, removalMask(item.PurchasableID, suite.Catalog), counter)
					if removeErr != nil {
						return RelevanceReport{}, fmt.Errorf("relevance run %s seed %d removal mask %s: %w", spec.PolicyID, seed, item.PurchasableID, removeErr)
					}
					executedRuns++
					appendRelevancePair(removalPairs, item.PurchasableID, spec.PolicyID, baseline.MilestoneMS, removed.MilestoneMS)
				}
			}
			for _, group := range suite.Policy.Groups {
				masked, runErr := suite.runPersona(spec, seed, groupMask(group.MemberIDs, suite.Catalog), counter)
				if runErr != nil {
					return RelevanceReport{}, fmt.Errorf("relevance run %s seed %d group mask %s: %w", spec.PolicyID, seed, group.GroupID, runErr)
				}
				executedRuns++
				appendRelevancePair(groupPairs, group.GroupID, spec.PolicyID, baseline.MilestoneMS, masked.MilestoneMS)
			}
			if spec.Reference {
				beam, beamErr := suite.runBeam(counter)
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
	for _, item := range suite.Policy.Items {
		baselinePurchases[item.PurchasableID] = reduceRelevanceCounts(referencePurchaseSamples[item.PurchasableID], suite.Scenario.Reducer)
	}
	baselineRoles, roleReduceErr := reduceRelevanceRoles(baselineRoleRuns, suite.Scenario.Reducer)
	if roleReduceErr != nil {
		return RelevanceReport{}, roleReduceErr
	}
	groupReduced := map[string]RelevanceDelta{}
	for _, group := range suite.Policy.Groups {
		reduced, reduceErr := reduceRelevancePairMatrix(groupPairs[group.GroupID], suite.Scenario.Reducer,
			suite.Scenario.Milestone.ID, suite.Scenario.HorizonMS)
		if reduceErr != nil {
			return RelevanceReport{}, reduceErr
		}
		passed := reduced.DeltaMS != nil && *reduced.DeltaMS >= group.EpsilonMS
		report.Groups = append(report.Groups, RelevanceGroupReport{GroupID: group.GroupID, Axis: group.Axis, Deltas: []RelevanceDelta{reduced}, Passed: passed})
		groupReduced[group.GroupID] = reduced
		if group.Axis == "tier" {
			report.TierContributions = append(report.TierContributions, RelevanceTierContribution{GroupID: group.GroupID, Deltas: []RelevanceDelta{reduced}})
		}
	}
	for _, item := range suite.Policy.Items {
		individual, reduceErr := reduceRelevancePairMatrix(itemPairs[item.PurchasableID], suite.Scenario.Reducer,
			suite.Scenario.Milestone.ID, suite.Scenario.HorizonMS)
		if reduceErr != nil {
			return RelevanceReport{}, reduceErr
		}
		removal, reduceErr := reduceRelevancePairMatrix(removalPairs[item.PurchasableID], suite.Scenario.Reducer,
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
		report.Items = append(report.Items, RelevanceItemReport{PurchasableID: item.PurchasableID,
			AvailabilityWindow: item.Availability, EpsilonMS: item.EpsilonMS, TrapExempt: item.TrapExempt,
			JustificationKey: item.JustificationKey, BaselinePurchaseCount: baselinePurchases[item.PurchasableID],
			IndividualDeltas: []RelevanceDelta{individual}, ActionRemovalDeltas: []RelevanceDelta{removal}, Support: support,
			SupportingGroupID: supportingGroup, NearestPassingEpsilonMS: nearest, RelevancePassed: relevancePassed, TrapPassed: trapPassed})
		if !relevancePassed {
			report.Failures = append(report.Failures, "relevance_floor:"+item.PurchasableID)
		}
		if !trapPassed {
			report.Failures = append(report.Failures, "trap_floor:"+item.PurchasableID)
		}
	}
	report.RoleActivations = sortedRoleActivations(baselineRoles)
	roleGenerators := map[string]bool{}
	for _, role := range report.RoleActivations {
		if role.Count > 0 {
			roleGenerators[role.GeneratorID] = true
		}
	}
	for _, generator := range suite.Catalog.GeneratorClassesForScope(economy.ScopeCompany) {
		if !roleGenerators[generator.ID] {
			report.Failures = append(report.Failures, "role_floor:"+generator.ID)
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
	state, err := suite.newRelevanceState()
	if err != nil {
		return relevanceRunResult{}, err
	}
	result := relevanceRunResult{Purchases: map[string]int64{}, Roles: map[string]RoleActivationCount{}, FinalState: state}
	revision, nowMS := int64(1), int64(0)
	for decision := int64(0); decision < suite.Scenario.MaxDecisions; decision++ {
		if reached, reachErr := suite.relevanceMilestoneReached(state); reachErr != nil {
			return relevanceRunResult{}, reachErr
		} else if reached {
			copy := nowMS
			result.MilestoneMS = &copy
			break
		}
		if suite.needsReferenceBootstrap(state) {
			manual, manualErr := suite.applyReferenceManual(state, revision, nowMS, mask, counter)
			if manualErr != nil {
				return relevanceRunResult{}, manualErr
			}
			revision++
			mergeRoleActivations(result.Roles, manual.RoleActivations)
		}
		result.Decisions = decision + 1
		next, ok := suite.nextDecisionHorizon(nowMS)
		if !ok {
			break
		}
		candidates, rankErr := suite.rankCandidates(state, revision, nowMS, next, mask, counter)
		if rankErr != nil {
			return relevanceRunResult{}, rankErr
		}
		if len(candidates) == 0 {
			advanced, advanceErr := suite.advance(state, revision, next, mask, counter)
			if advanceErr != nil {
				return relevanceRunResult{}, advanceErr
			}
			mergeRoleActivations(result.Roles, advanced.RoleActivations)
			nowMS = next
			continue
		}
		chosen := candidates[0]
		state, revision, nowMS = chosen.State, chosen.Revision, chosen.AtMS
		result.Purchases[chosen.ID]++
		mergeRoleActivations(result.Roles, chosen.RoleActivations)
	}
	if result.MilestoneMS == nil && nowMS < suite.Scenario.HorizonMS {
		finishedMS, reachedMS, activations, advanceErr := suite.finishToMilestone(state, revision, nowMS, mask, counter)
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
	return suite.rankCandidateIDs(state, revision, nowMS, horizonMS, suite.purchasableIDs(), mask, counter)
}

func (suite *RelevanceSuite) rankCandidateIDs(state *save.State, revision, nowMS, horizonMS int64, ids []string, mask production.AblationMask, counter *relevanceCounter) ([]relevanceCandidate, error) {
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
	sort.Slice(result, func(left, right int) bool {
		if result[left].PaybackMS != result[right].PaybackMS {
			return result[left].PaybackMS < result[right].PaybackMS
		}
		if result[left].EarliestPositiveDeltaMS != result[right].EarliestPositiveDeltaMS {
			return result[left].EarliestPositiveDeltaMS < result[right].EarliestPositiveDeltaMS
		}
		return result[left].ID < result[right].ID
	})
	return result, nil
}

// beamCandidateIDs applies T01-C17's cheap, deterministic selection before
// any affordability search, marginal-output probe, or greedy rollout. The
// real one-unit quote is the ordering key and the raw-byte ID is the tie-break.
func (suite *RelevanceSuite) beamCandidateIDs(state *save.State) ([]string, error) {
	type cheapCandidate struct {
		id   string
		cost decimal.Decimal
	}
	values := make([]cheapCandidate, 0, len(suite.Policy.Items))
	for _, id := range suite.purchasableIDs() {
		if state.UpgradesOwned[id] {
			continue
		}
		resourceID, cost, err := suite.candidateCost(state, id)
		if err != nil {
			return nil, err
		}
		if resourceID == suite.Scenario.Milestone.ResourceID {
			values = append(values, cheapCandidate{id: id, cost: cost})
		}
	}
	sort.Slice(values, func(left, right int) bool {
		if !values[left].cost.Eq(values[right].cost) {
			return values[left].cost.Lt(values[right].cost)
		}
		return values[left].id < values[right].id
	})
	if int64(len(values)) > suite.Scenario.BeamChildren {
		values = values[:suite.Scenario.BeamChildren]
	}
	result := make([]string, len(values))
	for index := range values {
		result[index] = values[index].id
	}
	return result, nil
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
	grossAt := func(atMS int64) (decimal.Decimal, error) {
		baseline, cloneErr := cloneState(suite.Catalog, baseAtBuy)
		if cloneErr != nil {
			return decimal.NaN, cloneErr
		}
		candidate, cloneErr := cloneState(suite.Catalog, candidateAtBuy)
		if cloneErr != nil {
			return decimal.NaN, cloneErr
		}
		if atMS > earliest {
			if _, advanceErr := suite.advance(baseline, revision, atMS, mask, counter); advanceErr != nil {
				return decimal.NaN, advanceErr
			}
			if _, advanceErr := suite.advance(candidate, revision+1, atMS, mask, counter); advanceErr != nil {
				return decimal.NaN, advanceErr
			}
		}
		baseBalance, _ := baseline.Ledger.Balance(resourceID)
		candidateBalance, _ := candidate.Ledger.Balance(resourceID)
		return candidateBalance.Add(cost).Sub(baseBalance), nil
	}
	gross, err := grossAt(horizonMS)
	if err != nil {
		return relevanceCandidate{}, false, err
	}
	if !gross.Gt(decimal.Zero) {
		return relevanceCandidate{}, false, nil
	}
	positiveLow, positiveHigh := earliest, horizonMS
	for positiveLow < positiveHigh {
		middle := positiveLow + (positiveHigh-positiveLow)/2
		value, valueErr := grossAt(middle)
		if valueErr != nil {
			return relevanceCandidate{}, false, valueErr
		}
		if value.Gt(decimal.Zero) {
			positiveHigh = middle
		} else {
			positiveLow = middle + 1
		}
	}
	duration := horizonMS - earliest
	value, err := ceilDecimalRatio(cost.Mul(decimal.FromFloat64(float64(duration))), gross)
	if err != nil {
		if errors.Is(err, errRelevanceRatioOutsideExactDomain) {
			return relevanceCandidate{}, false, nil
		}
		return relevanceCandidate{}, false, fmt.Errorf("relevance candidate %s payback: %w", id, err)
	}
	if value > relevanceMaxSafeInteger-(earliest-nowMS) {
		return relevanceCandidate{}, false, fmt.Errorf("relevance candidate %s payback exceeds exact integer domain", id)
	}
	return relevanceCandidate{ID: id, PaybackMS: earliest - nowMS + value, EarliestPositiveDeltaMS: positiveLow,
		AtMS: earliest, State: candidateAtBuy, Revision: revision + 1, RoleActivations: activations}, true, nil
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

func appendRelevancePair(destination map[string]map[string][]relevancePairedResult, targetID, policyID string, baseline, ablated *int64) {
	if destination[targetID] == nil {
		destination[targetID] = map[string][]relevancePairedResult{}
	}
	destination[targetID][policyID] = append(destination[targetID][policyID], relevancePairedResult{
		baseline: cloneInt64(baseline), ablated: cloneInt64(ablated),
	})
}

// reduceRelevancePairMatrix applies the scenario's conservative reducer inside
// each persona first, then the ANY gate selects the strongest persona result.
func reduceRelevancePairMatrix(byPolicy map[string][]relevancePairedResult, reducer, milestoneID string, horizonMS int64) (RelevanceDelta, error) {
	if len(byPolicy) == 0 {
		return RelevanceDelta{}, errors.New("missing relevance pair matrix")
	}
	policies := make([]string, 0, len(byPolicy))
	for policyID := range byPolicy {
		policies = append(policies, policyID)
	}
	sort.Strings(policies)
	var selected *RelevanceDelta
	var fallback *RelevanceDelta
	for _, policyID := range policies {
		pairs := byPolicy[policyID]
		rows := make([]RelevanceDelta, 0, len(pairs))
		for _, pair := range pairs {
			rows = append(rows, MakeRelevanceDelta(milestoneID, pair.baseline, pair.ablated, horizonMS))
		}
		reduced, err := reduceRelevanceDeltas(rows, reducer)
		if err != nil {
			return RelevanceDelta{}, err
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
		return *selected, nil
	}
	return *fallback, nil
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

// runBeam performs the ruled width-eight search. Each child is scored by a
// deterministic greedy rollout; equal-state nodes deduplicate on state hash +
// virtual time, with raw-byte paths as the final ordering key.
func (suite *RelevanceSuite) runBeam(counter *relevanceCounter) (*int64, error) {
	type node struct {
		state    *save.State
		revision int64
		atMS     int64
		path     string
		reached  *int64
		score    int64
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
				manual, manualErr := suite.applyReferenceManual(current.state, current.revision, current.atMS, production.AblationMask{}, counter)
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
			candidateIDs, selectionErr := suite.beamCandidateIDs(current.state)
			if selectionErr != nil {
				return nil, selectionErr
			}
			candidates, rankErr := suite.rankCandidateIDs(current.state, current.revision, current.atMS, next, candidateIDs, production.AblationMask{}, counter)
			if rankErr != nil {
				return nil, rankErr
			}
			for _, candidate := range candidates {
				children = append(children, node{state: candidate.State, revision: candidate.Revision, atMS: candidate.AtMS,
					path: current.path + "/" + candidate.ID})
			}
			waitState, cloneErr := cloneState(suite.Catalog, current.state)
			if cloneErr != nil {
				return nil, cloneErr
			}
			if _, advanceErr := suite.advance(waitState, current.revision, next, production.AblationMask{}, counter); advanceErr != nil {
				return nil, advanceErr
			}
			children = append(children, node{state: waitState, revision: current.revision, atMS: next, path: current.path + "/~wait"})
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
		// Collapse identical children before running their expensive greedy
		// completions; identical nodes have identical scores, so the raw-byte
		// path tie-break can be applied without evaluating duplicate rollouts.
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
			rolloutState, cloneErr := cloneState(suite.Catalog, child.state)
			if cloneErr != nil {
				return nil, cloneErr
			}
			rolloutSuite := *suite
			remainingDecisions := suite.Scenario.MaxDecisions - depth - 1
			rollout, rolloutErr := rolloutSuite.runReferenceFrom(rolloutState, child.revision, child.atMS, remainingDecisions, counter)
			if rolloutErr != nil {
				return nil, rolloutErr
			}
			child.reached = rollout.MilestoneMS
			child.score = suite.Scenario.HorizonMS + 1
			if child.reached != nil {
				child.score = *child.reached
				if best == nil || *child.reached < *best {
					copy := *child.reached
					best = &copy
				}
			}
		}
		frontier = frontier[:0]
		for _, child := range children {
			frontier = append(frontier, child)
		}
		sort.Slice(frontier, func(left, right int) bool {
			if frontier[left].score != frontier[right].score {
				return frontier[left].score < frontier[right].score
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
	result := relevanceRunResult{Purchases: map[string]int64{}, Roles: map[string]RoleActivationCount{}, FinalState: state}
	for decision := int64(0); decision < maxDecisions; decision++ {
		if reached, err := suite.relevanceMilestoneReached(state); err != nil {
			return relevanceRunResult{}, err
		} else if reached {
			copy := nowMS
			result.MilestoneMS = &copy
			break
		}
		if suite.needsReferenceBootstrap(state) {
			manual, err := suite.applyReferenceManual(state, revision, nowMS, production.AblationMask{}, counter)
			if err != nil {
				return relevanceRunResult{}, err
			}
			revision++
			mergeRoleActivations(result.Roles, manual.RoleActivations)
		}
		result.Decisions = decision + 1
		next, ok := suite.nextDecisionHorizon(nowMS)
		if !ok {
			break
		}
		candidates, err := suite.rankCandidates(state, revision, nowMS, next, production.AblationMask{}, counter)
		if err != nil {
			return relevanceRunResult{}, err
		}
		if len(candidates) == 0 {
			if _, err := suite.advance(state, revision, next, production.AblationMask{}, counter); err != nil {
				return relevanceRunResult{}, err
			}
			nowMS = next
			continue
		}
		state, revision, nowMS = candidates[0].State, candidates[0].Revision, candidates[0].AtMS
	}
	if result.MilestoneMS == nil && nowMS < suite.Scenario.HorizonMS {
		finishedMS, reachedMS, _, err := suite.finishToMilestone(state, revision, nowMS, production.AblationMask{}, counter)
		if err != nil {
			return relevanceRunResult{}, err
		}
		nowMS, result.MilestoneMS = finishedMS, reachedMS
	}
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

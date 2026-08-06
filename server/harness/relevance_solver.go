package harness

import (
	"crypto/sha256"
	"encoding/hex"
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
		return errors.New("relevance transition budget exceeded")
	}
	counter.value++
	return nil
}

func (suite *RelevanceSuite) RunRelevance() (RelevanceReport, error) {
	nonReferenceSeeds, referenceSeeds := int64(0), int64(0)
	for _, run := range suite.Scenario.Runs {
		if run.Reference {
			referenceSeeds += run.SeedCount
		} else {
			nonReferenceSeeds += run.SeedCount
		}
	}
	declaredRuns, err := ComputeRelevanceRunBudget(nonReferenceSeeds, referenceSeeds, int64(len(suite.Policy.Items)), int64(len(suite.Policy.Groups)), true)
	if err != nil || declaredRuns > suite.Scenario.RelevanceBudgetMaxRuns {
		return RelevanceReport{}, errors.New("relevance run budget exceeds scenario limit")
	}
	transitionCeiling, err := suite.preflightTransitionCeiling(declaredRuns, referenceSeeds)
	if err != nil || transitionCeiling > suite.Scenario.RelevanceBudgetMaxTransitions {
		return RelevanceReport{}, errors.New("relevance transition budget exceeds scenario limit")
	}
	counter := &relevanceCounter{limit: suite.Scenario.RelevanceBudgetMaxTransitions}
	report := RelevanceReport{SchemaVersion: RelevanceReportSchemaVersion, ScenarioID: suite.Scenario.ID,
		ScenarioHash: suite.ScenarioHash, ConstantsHash: suite.ConstantsHash, RelevancePolicyHash: suite.Policy.Hash,
		Items: []RelevanceItemReport{}, Groups: []RelevanceGroupReport{}, TierContributions: []RelevanceTierContribution{},
		RoleActivations: []RoleActivationCount{}, Failures: []string{}}
	type paired struct{ baseline, ablated *int64 }
	itemPairs := map[string][]paired{}
	removalPairs := map[string][]paired{}
	groupPairs := map[string][]paired{}
	baselinePurchases := map[string]int64{}
	baselineRoles := map[string]RoleActivationCount{}
	executedRuns := int64(0)
	var firstReference *relevanceRunResult
	for _, spec := range suite.Scenario.Runs {
		start, _ := parseSeed(spec.SeedStart)
		for offset := int64(0); offset < spec.SeedCount; offset++ {
			baseline, runErr := suite.runReference(production.AblationMask{}, counter)
			if runErr != nil {
				return RelevanceReport{}, runErr
			}
			executedRuns++
			_ = start + uint64(offset) // Seed remains part of the matrix even while the fixture transition is deterministic.
			if spec.Reference {
				if firstReference == nil {
					copy := baseline
					firstReference = &copy
				}
				for id, count := range baseline.Purchases {
					baselinePurchases[id] += count
				}
			}
			for key, role := range baseline.Roles {
				entry := baselineRoles[key]
				entry.GeneratorID, entry.Kind, entry.TargetID, entry.Count = role.GeneratorID, role.Kind, role.TargetID, entry.Count+role.Count
				baselineRoles[key] = entry
			}
			for _, item := range suite.Policy.Items {
				masked, runErr := suite.runReference(effectMask(item.PurchasableID, suite.Catalog), counter)
				if runErr != nil {
					return RelevanceReport{}, runErr
				}
				executedRuns++
				itemPairs[item.PurchasableID] = append(itemPairs[item.PurchasableID], paired{baseline.MilestoneMS, masked.MilestoneMS})
				if spec.Reference {
					removed, removeErr := suite.runReference(removalMask(item.PurchasableID, suite.Catalog), counter)
					if removeErr != nil {
						return RelevanceReport{}, removeErr
					}
					executedRuns++
					removalPairs[item.PurchasableID] = append(removalPairs[item.PurchasableID], paired{baseline.MilestoneMS, removed.MilestoneMS})
				}
			}
			for _, group := range suite.Policy.Groups {
				masked, runErr := suite.runReference(groupMask(group.MemberIDs, suite.Catalog), counter)
				if runErr != nil {
					return RelevanceReport{}, runErr
				}
				executedRuns++
				groupPairs[group.GroupID] = append(groupPairs[group.GroupID], paired{baseline.MilestoneMS, masked.MilestoneMS})
			}
			if spec.Reference {
				executedRuns++ // one beam invocation is one R14 run, regardless of internal transitions.
			}
		}
	}
	if executedRuns != declaredRuns || firstReference == nil {
		return RelevanceReport{}, errors.New("relevance run cardinality mismatch")
	}
	groupReduced := map[string]RelevanceDelta{}
	for _, group := range suite.Policy.Groups {
		rows := make([]RelevanceDelta, 0, len(groupPairs[group.GroupID]))
		for _, pair := range groupPairs[group.GroupID] {
			rows = append(rows, MakeRelevanceDelta(suite.Scenario.Milestone.ID, pair.baseline, pair.ablated, suite.Scenario.HorizonMS))
		}
		reduced, reduceErr := reduceRelevanceDeltas(rows, suite.Scenario.Reducer)
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
		individualRows := make([]RelevanceDelta, 0, len(itemPairs[item.PurchasableID]))
		for _, pair := range itemPairs[item.PurchasableID] {
			individualRows = append(individualRows, MakeRelevanceDelta(suite.Scenario.Milestone.ID, pair.baseline, pair.ablated, suite.Scenario.HorizonMS))
		}
		individual, reduceErr := reduceRelevanceDeltas(individualRows, suite.Scenario.Reducer)
		if reduceErr != nil {
			return RelevanceReport{}, reduceErr
		}
		removalRows := make([]RelevanceDelta, 0, len(removalPairs[item.PurchasableID]))
		for _, pair := range removalPairs[item.PurchasableID] {
			removalRows = append(removalRows, MakeRelevanceDelta(suite.Scenario.Milestone.ID, pair.baseline, pair.ablated, suite.Scenario.HorizonMS))
		}
		removal, reduceErr := reduceRelevanceDeltas(removalRows, suite.Scenario.Reducer)
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
		nearest := item.EpsilonMS
		if individual.DeltaMS != nil {
			nearest -= *individual.DeltaMS
			if nearest < 0 {
				nearest = 0
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
	beamMS, beamErr := suite.runBeam(counter)
	if beamErr != nil {
		return RelevanceReport{}, beamErr
	}
	if firstReference.MilestoneMS == nil || beamMS == nil {
		report.Failures = append(report.Failures, "greedy_oracle:milestone_unreached")
	} else {
		gap := int64(0)
		if *firstReference.MilestoneMS > *beamMS {
			gap = (*firstReference.MilestoneMS - *beamMS) * 1_000_000 / maxInt64(*beamMS, 1)
		}
		passed := gap <= suite.Scenario.GreedyGapMaximumPPM
		report.GreedyOracle = &RelevanceGreedyOracle{MilestoneID: suite.Scenario.Milestone.ID,
			GreedyMS: *firstReference.MilestoneMS, BeamMS: *beamMS, GapPPM: gap, MaximumPPM: suite.Scenario.GreedyGapMaximumPPM, Passed: passed}
		if !passed {
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

func (suite *RelevanceSuite) preflightTransitionCeiling(runs, referenceSeeds int64) (int64, error) {
	items := int64(len(suite.Policy.Items))
	binarySteps := int64(1)
	for span := suite.Scenario.HorizonMS; span > 1; span = (span + 1) / 2 {
		binarySteps++
	}
	perDecision := items*(binarySteps+4) + 1
	perRun := suite.Scenario.MaxDecisions * perDecision
	beam := referenceSeeds * suite.Scenario.MaxDecisions * suite.Scenario.BeamWidth * (items + 1) * (perRun + perDecision)
	if perRun <= 0 || runs > relevanceMaxSafeInteger/perRun || beam < 0 || runs*perRun > relevanceMaxSafeInteger-beam {
		return 0, errors.New("relevance transition preflight overflow")
	}
	return runs*perRun + beam, nil
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
	ids := suite.purchasableIDs()
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
	baseAtHorizon, err := cloneState(suite.Catalog, baseAtBuy)
	if err != nil {
		return relevanceCandidate{}, false, err
	}
	candidateAtHorizon, err := cloneState(suite.Catalog, candidateAtBuy)
	if err != nil {
		return relevanceCandidate{}, false, err
	}
	if _, err := suite.advance(baseAtHorizon, revision, horizonMS, mask, counter); err != nil {
		return relevanceCandidate{}, false, err
	}
	if _, err := suite.advance(candidateAtHorizon, revision+1, horizonMS, mask, counter); err != nil {
		return relevanceCandidate{}, false, err
	}
	baseBalance, _ := baseAtHorizon.Ledger.Balance(resourceID)
	candidateBalance, _ := candidateAtHorizon.Ledger.Balance(resourceID)
	gross := candidateBalance.Add(cost).Sub(baseBalance)
	if !gross.Gt(decimal.Zero) {
		return relevanceCandidate{}, false, nil
	}
	duration := horizonMS - earliest
	value, err := ceilDecimalRatio(cost.Mul(decimal.FromFloat64(float64(duration))), gross)
	if err != nil || value > relevanceMaxSafeInteger-(earliest-nowMS) {
		return relevanceCandidate{}, false, err
	}
	return relevanceCandidate{ID: id, PaybackMS: earliest - nowMS + value, EarliestPositiveDeltaMS: horizonMS,
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
	intentID := fmt.Sprintf("00000000-0000-7000-8000-%012x", revision)
	if len(intentID) > 36 {
		intentID = "00000000-0000-7000-8000-000000000001"
	}
	if len(id) >= len("generator.") && id[:len("generator.")] == "generator." {
		return production.IntentRequest{IntentID: intentID, Kind: production.IntentBuyGenerator, ExpectedRevision: revision,
			GeneratorID: id, CountMode: "exact", Count: 1}
	}
	return production.IntentRequest{IntentID: intentID, Kind: production.IntentBuyUpgrade, ExpectedRevision: revision, UpgradeID: id}
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

func maxInt64(left, right int64) int64 {
	if left > right {
		return left
	}
	return right
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
			next, ok := suite.nextDecisionHorizon(current.atMS)
			if !ok {
				continue
			}
			candidates, rankErr := suite.rankCandidates(current.state, current.revision, current.atMS, next, production.AblationMask{}, counter)
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
		dedup := map[string]node{}
		for index := range children {
			child := &children[index]
			rolloutState, cloneErr := cloneState(suite.Catalog, child.state)
			if cloneErr != nil {
				return nil, cloneErr
			}
			rolloutSuite := *suite
			rollout, rolloutErr := rolloutSuite.runReferenceFrom(rolloutState, child.revision, child.atMS, counter)
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
			digest, digestErr := stateDigest(child.state)
			if digestErr != nil {
				return nil, digestErr
			}
			key := fmt.Sprintf("%s:%d", digest, child.atMS)
			prior, exists := dedup[key]
			if !exists || child.score < prior.score || child.score == prior.score && child.path < prior.path {
				dedup[key] = *child
			}
		}
		frontier = frontier[:0]
		for _, child := range dedup {
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

func (suite *RelevanceSuite) runReferenceFrom(state *save.State, revision, nowMS int64, counter *relevanceCounter) (relevanceRunResult, error) {
	result := relevanceRunResult{Purchases: map[string]int64{}, Roles: map[string]RoleActivationCount{}, FinalState: state}
	for decision := int64(0); decision < suite.Scenario.MaxDecisions; decision++ {
		if reached, err := suite.relevanceMilestoneReached(state); err != nil {
			return relevanceRunResult{}, err
		} else if reached {
			copy := nowMS
			result.MilestoneMS = &copy
			break
		}
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

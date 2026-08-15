package harness

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"sync"
	"time"

	"cloud-clicker/server/decimal"
	"cloud-clicker/server/economy"
	prestigecore "cloud-clicker/server/prestige"
	"cloud-clicker/server/production"
	"cloud-clicker/server/save"
)

type FirstHourExperiment struct {
	AcquihirePurchasedMinimum int64  `json:"acquihire_purchased_minimum"`
	BurnoutPriceFactor        string `json:"burnout_price_factor"`
	RouteKnowledgeBonus       int64  `json:"route_knowledge_bonus"`
	SeedCapital               string `json:"seed_capital"`
	GeneratedBeigeTowers      int64  `json:"generated_beige_towers"`
}

type FirstHourEndingSample struct {
	PolicyID                 string `json:"policy_id"`
	Seed                     string `json:"seed"`
	Branch                   string `json:"branch"`
	GeneratorPurchasedTotal  int64  `json:"generator_purchased_total"`
	UpgradeCount             int64  `json:"upgrade_count"`
	Cash                     string `json:"cash"`
	CheapestUnownedGenerator string `json:"cheapest_unowned_generator"`
	CheapestUnownedPrice     string `json:"cheapest_unowned_price"`
	RunOneGateMS             int64  `json:"run_one_gate_ms"`
	RunTwoGateMS             int64  `json:"run_two_gate_ms"`
}

type FirstHourRunResult struct {
	Key               RunKey                 `json:"key"`
	PolicyHash        string                 `json:"policy_hash"`
	Outcome           string                 `json:"outcome"`
	Milestones        []TimedMilestone       `json:"milestones"`
	Ending            *FirstHourEndingSample `json:"ending,omitempty"`
	TransitionCount   int64                  `json:"transition_count"`
	InvariantFailures []string               `json:"invariant_failures"`
}

// FirstHourScriptCommand is one player decision emitted by the ratified
// first-hour policy. Persistence revisions and intent IDs are coordinates the
// composed executor supplies; they are not policy decisions.
type FirstHourScriptCommand struct {
	AtMS    int64                     `json:"at_ms"`
	Mode    production.EvaluationMode `json:"mode"`
	Request production.IntentRequest  `json:"-"`
}

type FirstHourExperimentReport struct {
	SchemaVersion int                  `json:"schema_version"`
	ScenarioID    string               `json:"scenario_id"`
	ScenarioHash  string               `json:"scenario_hash"`
	PolicyHash    string               `json:"policy_hash"`
	ConstantsHash string               `json:"constants_hash"`
	Experiment    FirstHourExperiment  `json:"experiment"`
	Runs          []FirstHourRunResult `json:"runs"`
	Aggregate     AggregateReport      `json:"aggregate"`
}

type firstHourBoundary struct {
	atMS         int64
	sessionIndex int64
	first        bool
}

type firstHourRuntime struct {
	suite             *FirstHourSuite
	spec              RunSpec
	policy            FirstHourPolicy
	seed              uint64
	experiment        FirstHourExperiment
	company           *save.State
	founder           *save.State
	revision          int64
	decisionOrdinal   int64
	transitions       int64
	founderAttendedMS int64
	milestones        map[string]*int64
	ending            *FirstHourEndingSample
	lastSession       int64
	commands          *[]FirstHourScriptCommand
}

func (suite *FirstHourSuite) RunExperiment(spec RunSpec, seed uint64, experiment FirstHourExperiment) FirstHourRunResult {
	result, _ := suite.runExperiment(spec, seed, experiment, false)
	return result
}

// RunExperimentScript returns the exact ordered player decisions consumed by
// the headless measurement. It is intentionally a single-run proof seam; the
// scenario and policy registry remain the only script authority.
func (suite *FirstHourSuite) RunExperimentScript(spec RunSpec, seed uint64, experiment FirstHourExperiment) (FirstHourRunResult, []FirstHourScriptCommand) {
	return suite.runExperiment(spec, seed, experiment, true)
}

func (suite *FirstHourSuite) runExperiment(spec RunSpec, seed uint64, experiment FirstHourExperiment, capture bool) (FirstHourRunResult, []FirstHourScriptCommand) {
	result := FirstHourRunResult{Key: suite.RunKey(spec, seed), PolicyHash: suite.PolicyHash, Outcome: "completed", InvariantFailures: []string{}}
	commands := []FirstHourScriptCommand{}
	if err := validateFirstHourExperiment(experiment); err != nil {
		return failFirstHour(result, err), commands
	}
	policy, ok := suite.Policy.Policy(spec.PolicyID, spec.PolicyVersion)
	if !ok {
		return failFirstHour(result, fmt.Errorf("unknown first-hour policy %s v%d", spec.PolicyID, spec.PolicyVersion)), commands
	}
	company, err := newFirstHourCompany(suite.Bundle.Economy)
	if err != nil {
		return failFirstHour(result, err), commands
	}
	founder := &save.State{ReputationLevel: 0, RouteKnowledgeBalance: 0, LedgerFactKinds: map[string]bool{}, NetworkSlots: []save.NetworkSlot{}, ExitHistory: []save.ExitRecord{}}
	runtime := firstHourRuntime{suite: suite, spec: spec, policy: policy, seed: seed, experiment: experiment,
		company: company, founder: founder, revision: 1, milestones: map[string]*int64{}, lastSession: -1}
	if capture {
		runtime.commands = &commands
	}
	boundaries, err := firstHourBoundaries(policy, seed, spec.HorizonMS)
	if err != nil {
		return failFirstHour(result, err), commands
	}
	for _, boundary := range boundaries {
		if runtime.transitions >= suite.Scenario.TransitionBudget {
			return failFirstHour(result, fmt.Errorf("first-hour transition budget exceeded: executed %d, maximum %d", runtime.transitions, suite.Scenario.TransitionBudget)), commands
		}
		if err := runtime.step(boundary); err != nil {
			return failFirstHour(result, err), commands
		}
		if runtime.milestones["milestone.first_elective_exit"] != nil {
			break
		}
	}
	result.Milestones = firstHourMilestoneReport(suite.Scenario.Milestones, runtime.milestones, &result.InvariantFailures)
	result.Ending, result.TransitionCount = runtime.ending, runtime.transitions
	if len(result.InvariantFailures) != 0 {
		result.Outcome = "failed"
	}
	return result, commands
}

func (suite *FirstHourSuite) RunAllExperiments(experiment FirstHourExperiment, workerLimit int) (FirstHourExperimentReport, error) {
	if workerLimit < 1 {
		return FirstHourExperimentReport{}, errors.New("first-hour worker limit must be positive")
	}
	if err := validateFirstHourExperiment(experiment); err != nil {
		return FirstHourExperimentReport{}, err
	}
	type task struct {
		spec RunSpec
		seed uint64
	}
	tasks := []task{}
	for _, spec := range suite.Scenario.Runs {
		start, err := strconv.ParseUint(spec.SeedStart, 10, 64)
		if err != nil || spec.SeedCount < 1 || uint64(spec.SeedCount-1) > ^uint64(0)-start {
			return FirstHourExperimentReport{}, errors.New("invalid first-hour run specification")
		}
		for offset := 0; offset < spec.SeedCount; offset++ {
			tasks = append(tasks, task{spec: spec, seed: start + uint64(offset)})
		}
	}
	runs := make([]FirstHourRunResult, len(tasks))
	work := make(chan int)
	workers := workerLimit
	if len(tasks) < workers {
		workers = len(tasks)
	}
	var wait sync.WaitGroup
	for worker := 0; worker < workers; worker++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			for index := range work {
				runs[index] = suite.RunExperiment(tasks[index].spec, tasks[index].seed, experiment)
			}
		}()
	}
	for index := range tasks {
		work <- index
	}
	close(work)
	wait.Wait()
	sort.Slice(runs, func(left, right int) bool { return lessKey(runs[left].Key, runs[right].Key) })
	aggregate := suite.aggregateFirstHour(runs)
	return FirstHourExperimentReport{SchemaVersion: 1, ScenarioID: suite.Scenario.ID, ScenarioHash: suite.ScenarioHash,
		PolicyHash: suite.PolicyHash, ConstantsHash: suite.ConstantsHash, Experiment: experiment, Runs: runs, Aggregate: aggregate}, nil
}

func validateFirstHourExperiment(experiment FirstHourExperiment) error {
	factor, factorErr := decimal.ParseCanonical(experiment.BurnoutPriceFactor)
	capital, capitalErr := decimal.ParseCanonical(experiment.SeedCapital)
	if experiment.AcquihirePurchasedMinimum < 1 || experiment.AcquihirePurchasedMinimum > decimal.MaxExactInteger ||
		factorErr != nil || !factor.Gt(decimal.Zero) || capitalErr != nil || !capital.Gt(decimal.Zero) ||
		experiment.RouteKnowledgeBonus < 0 || experiment.RouteKnowledgeBonus > decimal.MaxExactInteger || experiment.RouteKnowledgeBonus%2 != 0 ||
		experiment.GeneratedBeigeTowers < 1 || experiment.GeneratedBeigeTowers > decimal.MaxExactInteger {
		return errors.New("invalid first-hour experiment tuple")
	}
	return nil
}

func (suite *FirstHourSuite) aggregateFirstHour(runs []FirstHourRunResult) AggregateReport {
	aggregate := AggregateReport{SchemaVersion: 1, ScenarioID: suite.Scenario.ID, ScenarioHash: suite.ScenarioHash,
		ConstantsHash: suite.ConstantsHash, RunCount: len(runs), Values: []AggregateValue{}, Warnings: []string{}, Failures: []string{}}
	for _, run := range runs {
		for _, failure := range run.InvariantFailures {
			aggregate.Failures = append(aggregate.Failures, formatRunKey(run.Key)+":"+failure)
		}
	}
	for _, envelope := range suite.Scenario.Envelopes {
		values := []int64{}
		for _, run := range runs {
			if run.Key.PolicyID != envelope.PolicyID {
				continue
			}
			for _, milestone := range run.Milestones {
				if milestone.ID == envelope.Milestone && milestone.FirstMS != nil {
					values = append(values, *milestone.FirstMS)
				}
			}
		}
		if len(values) == 0 {
			aggregate.Failures = append(aggregate.Failures, "envelope has no reached values: "+envelope.Milestone)
			continue
		}
		sort.Slice(values, func(left, right int) bool { return values[left] < values[right] })
		value := statistic(values, envelope.Statistic)
		aggregate.Values = append(aggregate.Values, AggregateValue{PolicyID: envelope.PolicyID,
			Milestone: envelope.Milestone, Statistic: envelope.Statistic, ValueMS: value})
		if envelope.MinimumMS != nil && value < *envelope.MinimumMS || envelope.MaximumMS != nil && value > *envelope.MaximumMS {
			aggregate.Failures = append(aggregate.Failures, fmt.Sprintf("envelope %s/%s/%s=%d outside bounds", envelope.PolicyID, envelope.Milestone, envelope.Statistic, value))
		}
	}
	for _, relation := range suite.Scenario.Relations {
		for _, run := range runs {
			if run.Key.PolicyID != relation.PolicyID {
				continue
			}
			left, right := milestoneValue(run.Milestones, relation.LeftMilestone), milestoneValue(run.Milestones, relation.RightMilestone)
			if left == nil || right == nil || relation.Kind == "less_than" && *left >= *right {
				aggregate.Failures = append(aggregate.Failures, fmt.Sprintf("relation %s/%s/%s/seed=%s failed", relation.PolicyID, relation.LeftMilestone, relation.RightMilestone, run.Key.Seed))
			}
		}
	}
	sort.Strings(aggregate.Failures)
	return aggregate
}

func milestoneValue(values []TimedMilestone, id string) *int64 {
	for _, value := range values {
		if value.ID == id {
			return value.FirstMS
		}
	}
	return nil
}

func newFirstHourCompany(catalog *economy.Catalog) (*save.State, error) {
	ledger, err := economy.NewLedger(catalog, economy.ScopeCompany)
	if err != nil {
		return nil, err
	}
	counts, provisioned, remainders := map[string]int64{}, map[string]int64{}, map[string]int64{}
	for _, generator := range catalog.GeneratorClassesForScope(economy.ScopeCompany) {
		counts[generator.ID], provisioned[generator.ID] = 0, 0
		if generator.Provision != nil {
			remainders[generator.Provision.GeneratorID] = 0
		}
	}
	return &save.State{Ledger: ledger, GeneratorCounts: counts, GeneratorProvisioned: provisioned, ProvisionRemaindersPPM: remainders,
		UpgradesOwned: map[string]bool{}, EvaluatedThrough: Epoch, RunStartedAt: Epoch, RunSeq: 1,
		ManualTokenMilli: catalog.ManualPolicy().BucketCapMilli, ManualTokenRefilledAt: Epoch,
		GatesCrossed: map[string]bool{}, DoctrinesByTransition: map[string]string{}, LedgerFactKinds: map[string]bool{},
		MeterBands: map[string]int{}, RegionTraits: map[string]bool{}, HintsUnlocked: map[string]bool{}, OfflineSpans: []save.OfflineSpan{},
		NetworkSlots: []save.NetworkSlot{}, ExitHistory: []save.ExitRecord{}, LifetimeValue: decimal.Zero}, nil
}

func firstHourBoundaries(policy FirstHourPolicy, seed uint64, horizonMS int64) ([]firstHourBoundary, error) {
	if horizonMS < 0 || len(policy.Sessions) != 1 {
		return nil, errors.New("invalid first-hour boundary horizon")
	}
	base := policy.Sessions[0]
	if base.RepeatEveryMS == 0 {
		result := make([]firstHourBoundary, 0, horizonMS/policy.ActionCadenceMS+1)
		for at := base.StartMS; at < base.StartMS+base.DurationMS && at <= horizonMS; at += policy.ActionCadenceMS {
			result = append(result, firstHourBoundary{atMS: at, sessionIndex: 0, first: at == base.StartMS})
		}
		return result, nil
	}
	result := []firstHourBoundary{}
	for sessionIndex := int64(0); ; sessionIndex++ {
		baseStart := base.StartMS + sessionIndex*base.RepeatEveryMS
		if baseStart > horizonMS {
			break
		}
		start, duration := baseStart, base.DurationMS
		if policy.SessionJitter != nil {
			draw := firstHourDraw(policy.SessionJitter.DomainTag, policy.PolicyID, policy.PolicyVersion, seed, 0, sessionIndex)
			startRange := policy.SessionJitter.StartOffsetMS[1] - policy.SessionJitter.StartOffsetMS[0]
			durationRange := policy.SessionJitter.DurationMS[1] - policy.SessionJitter.DurationMS[0]
			start += policy.SessionJitter.StartOffsetMS[0] + int64(draw%uint64(startRange))
			duration = policy.SessionJitter.DurationMS[0] + int64(draw%uint64(durationRange))
		}
		for at := start; at < start+duration && at <= horizonMS; at += policy.ActionCadenceMS {
			result = append(result, firstHourBoundary{atMS: at, sessionIndex: sessionIndex, first: at == start})
		}
	}
	sort.Slice(result, func(left, right int) bool {
		if result[left].atMS != result[right].atMS {
			return result[left].atMS < result[right].atMS
		}
		return result[left].sessionIndex < result[right].sessionIndex
	})
	return result, nil
}

func (runtime *firstHourRuntime) step(boundary firstHourBoundary) error {
	now := Epoch.Add(time.Duration(boundary.atMS) * time.Millisecond)
	if now.Before(runtime.company.EvaluatedThrough) {
		return nil
	}
	mode := production.ModeOnline
	firstSessionGap := boundary.first && runtime.lastSession < 0 && boundary.atMS > 0
	laterSessionGap := boundary.first && runtime.lastSession >= 0 && boundary.sessionIndex != runtime.lastSession
	if firstSessionGap || laterSessionGap {
		mode = production.ModeOffline
		if err := prestigecore.RecordOfflineSpan(runtime.company, runtime.company.EvaluatedThrough, now, runtime.suite.Bundle.Prestige.CatchupCeilingMS); err != nil {
			return err
		}
	}
	runtime.lastSession = boundary.sessionIndex
	advanced, err := production.SimulateAdvance(runtime.company, runtime.suite.Bundle.Economy,
		production.SimulationDependencies{Routes: runtime.suite.Bundle.Routes}, runtime.companyRevision(), mode, now, nil, production.AblationMask{})
	if err != nil {
		return err
	}
	_ = advanced
	runtime.transitions++
	attended, err := prestigecore.AttendedMS(runtime.company, now)
	if err != nil {
		return err
	}
	founderAttended := runtime.founderAttendedMS + attended
	if runtime.scriptedFailureDue(attended) {
		runtime.recordCommand(boundary.atMS, mode, runtime.manualRequest())
		return runtime.applyEnding(now, boundary.atMS, attended)
	}
	if runtime.electiveExitReady(founderAttended) {
		runtime.recordCommand(boundary.atMS, mode, production.IntentRequest{Kind: production.IntentWindDown})
		runtime.observeRunEnded(boundary.atMS, "collapse")
		return nil
	}
	if gateID, ok := runtime.nextReadyGate(); ok {
		request := production.IntentRequest{IntentID: firstHourIntentID(runtime.revision), Kind: production.IntentCrossGate,
			ExpectedRevision: runtime.revision, GateID: gateID}
		return runtime.apply(request, now, boundary.atMS, mode)
	}
	if runtime.policy.Decision.Kind == "t01_c20_projected_time_ranker" {
		runtime.decisionOrdinal++
		return runtime.applyReferenceChoice(boundary.atMS)
	}
	request, wait, err := runtime.choose(now, boundary.atMS)
	if err != nil {
		return err
	}
	runtime.decisionOrdinal++
	if wait {
		return nil
	}
	return runtime.apply(request, now, boundary.atMS, mode)
}

func (runtime *firstHourRuntime) companyRevision() save.Revision {
	return save.Revision{StreamID: relevanceStreamID, OwnerID: relevanceOwnerID, Number: runtime.revision, ConstantsHash: runtime.suite.ConstantsHash}
}

func (runtime *firstHourRuntime) scriptedFailureDue(attended int64) bool {
	return runtime.company.RunSeq == 1 && len(runtime.founder.ExitHistory) == 0 && runtime.company.GatesCrossed["gate.t0_to_t1"] && attended >= 900_000
}

func (runtime *firstHourRuntime) electiveExitReady(founderAttended int64) bool {
	if runtime.company.RunSeq < 2 || !runtime.company.GatesCrossed["gate.t0_to_t1"] || founderAttended < 2_700_000 {
		return false
	}
	terms, err := prestigecore.ComputeTerms(runtime.company, runtime.founder, runtime.suite.Bundle.Prestige, "collapse")
	return err == nil && (terms.ReputationDelta > 0 || terms.RouteKnowledge > 0 || len(terms.NetworkSlotUnlocks) > 0)
}

func (runtime *firstHourRuntime) nextReadyGate() (string, bool) {
	for _, gateID := range runtime.suite.GateIDs() {
		if runtime.company.GatesCrossed[gateID] {
			continue
		}
		gate, ok := runtime.suite.Bundle.Routes.Gate(gateID)
		if ok && gateRequirementsMet(runtime.company, gate) {
			return gateID, true
		}
	}
	return "", false
}

func (runtime *firstHourRuntime) choose(now time.Time, nowMS int64) (production.IntentRequest, bool, error) {
	switch runtime.policy.Decision.Kind {
	case "seeded_uniform_over_legal":
		legal, err := runtime.legalCommands(now)
		if err != nil {
			return production.IntentRequest{}, false, err
		}
		choice, err := firstHourBoundedDraw("decision", runtime.policy.PolicyID, runtime.policy.PolicyVersion, runtime.seed,
			runtime.company.RunSeq, runtime.decisionOrdinal, uint64(len(legal)))
		if err != nil {
			return production.IntentRequest{}, false, err
		}
		selected := legal[choice]
		return selected.request, selected.wait, nil
	case "cheapest_affordable_then_manual":
		legal, err := runtime.legalCommands(now)
		if err != nil {
			return production.IntentRequest{}, false, err
		}
		purchases := make([]firstHourLegalCommand, 0)
		for _, command := range legal {
			if command.request.Kind == production.IntentBuyGenerator || command.request.Kind == production.IntentBuyUpgrade {
				purchases = append(purchases, command)
			}
		}
		if len(purchases) > 0 {
			sort.Slice(purchases, func(left, right int) bool {
				if comparison := purchases[left].cost.Cmp(purchases[right].cost); comparison != 0 {
					return comparison < 0
				}
				return purchases[left].id < purchases[right].id
			})
			return purchases[0].request, false, nil
		}
		return runtime.manualRequest(), false, nil
	default:
		return production.IntentRequest{}, false, fmt.Errorf("unknown first-hour decision kind %q", runtime.policy.Decision.Kind)
	}
}

type firstHourLegalCommand struct {
	id      string
	request production.IntentRequest
	wait    bool
	cost    decimal.Decimal
}

func (runtime *firstHourRuntime) legalCommands(now time.Time) ([]firstHourLegalCommand, error) {
	result := []firstHourLegalCommand{{id: "perform_manual_batch", request: runtime.manualRequest()}}
	generators := runtime.suite.Bundle.Economy.GeneratorClassesForScope(economy.ScopeCompany)
	sort.Slice(generators, func(left, right int) bool { return generators[left].ID < generators[right].ID })
	for _, generator := range generators {
		cost, err := runtime.suite.Bundle.Economy.BulkCost(generator.ID, runtime.company.GeneratorCounts[generator.ID], 1)
		if err != nil {
			return nil, err
		}
		request := candidateIntent(generator.ID, runtime.revision)
		if runtime.commandApplies(request, now) {
			result = append(result, firstHourLegalCommand{id: generator.ID, request: request, cost: cost})
		}
	}
	upgrades := runtime.suite.Bundle.Economy.Upgrades()
	sort.Slice(upgrades, func(left, right int) bool { return upgrades[left].ID < upgrades[right].ID })
	for _, upgrade := range upgrades {
		if runtime.company.UpgradesOwned[upgrade.ID] {
			continue
		}
		request := candidateIntent(upgrade.ID, runtime.revision)
		if runtime.commandApplies(request, now) {
			result = append(result, firstHourLegalCommand{id: upgrade.ID, request: request, cost: upgrade.Cost.Amount})
		}
	}
	rate, err := production.SimulateResourceRate(runtime.company, runtime.suite.Bundle.Economy, "company.cash", production.AblationMask{})
	if err != nil {
		return nil, err
	}
	if rate.Gt(decimal.Zero) {
		result = append(result, firstHourLegalCommand{id: "wait", wait: true})
	}
	return result, nil
}

func (runtime *firstHourRuntime) commandApplies(request production.IntentRequest, now time.Time) bool {
	clone, err := cloneState(runtime.suite.Bundle.Economy, runtime.company)
	if err != nil {
		return false
	}
	result, err := production.SimulateTransition(request, clone, runtime.suite.Bundle.Economy,
		production.SimulationDependencies{Routes: runtime.suite.Bundle.Routes}, runtime.companyRevision(), production.ModeOnline,
		now, nil, nil, production.AblationMask{})
	return err == nil && result.Decision.Outcome == save.IntentApplied
}

func (runtime *firstHourRuntime) manualRequest() production.IntentRequest {
	actions := runtime.suite.Bundle.Economy.ManualActions()
	sort.Slice(actions, func(left, right int) bool { return actions[left].ID < actions[right].ID })
	return production.IntentRequest{IntentID: firstHourIntentID(runtime.revision), Kind: production.IntentPerformManualBatch,
		ExpectedRevision: runtime.revision, ActionID: actions[0].ID, Count: 1, WindowMS: runtime.policy.ActionCadenceMS}
}

func (runtime *firstHourRuntime) applyReferenceChoice(nowMS int64) error {
	if runtime.company.GatesCrossed["gate.t0_to_t1"] {
		return nil
	}
	ranker, err := runtime.referenceRanker()
	if err != nil {
		return err
	}
	counter := &relevanceCounter{limit: runtime.suite.Scenario.TransitionBudget - runtime.transitions}
	decisionHorizon := nowMS + runtime.policy.ActionCadenceMS
	if decisionHorizon > runtime.spec.HorizonMS {
		decisionHorizon = runtime.spec.HorizonMS
	}
	candidates, bank, bankAtMS, err := ranker.rankDecisionOptions(runtime.company, runtime.revision, nowMS,
		decisionHorizon, 1, production.AblationMask{}, counter)
	if err != nil {
		return err
	}
	runtime.transitions += counter.value
	if len(candidates) > 0 {
		candidate := candidates[0]
		runtime.company, runtime.revision = candidate.State, candidate.Revision
		if err := validateFirstHourCompany(runtime.suite.Bundle.Economy, runtime.company); err != nil {
			return err
		}
		runtime.observeReferencePurchase(candidate.AtMS, candidate.ID)
		return nil
	}
	if bank && bankAtMS > nowMS {
		advanced, advanceErr := production.SimulateAdvance(runtime.company, runtime.suite.Bundle.Economy,
			production.SimulationDependencies{Routes: runtime.suite.Bundle.Routes}, runtime.companyRevision(), production.ModeOnline,
			relevanceNow(bankAtMS), nil, production.AblationMask{})
		if advanceErr != nil {
			return advanceErr
		}
		_ = advanced
		runtime.transitions++
		return validateFirstHourCompany(runtime.suite.Bundle.Economy, runtime.company)
	}
	// No producer and no affordable purchase: one ratified manual action is the
	// only legal bootstrap edge. This is the same base case as T01-C16, without
	// the relevance runner's batched-click optimization.
	return runtime.apply(runtime.manualRequest(), relevanceNow(nowMS), nowMS, production.ModeOnline)
}

func (runtime *firstHourRuntime) referenceRanker() (*RelevanceSuite, error) {
	gate, ok := runtime.suite.Bundle.Routes.Gate("gate.t0_to_t1")
	if !ok || len(gate.Requirement) != 1 {
		return nil, errors.New("first-hour reference objective is not a single-resource gate")
	}
	policy := &RelevancePolicy{Items: []RelevancePolicyItem{}}
	for _, generator := range runtime.suite.Bundle.Economy.GeneratorClassesForScope(economy.ScopeCompany) {
		if generator.Tier == 0 {
			policy.Items = append(policy.Items, RelevancePolicyItem{PurchasableID: generator.ID})
		}
	}
	for _, upgrade := range runtime.suite.Bundle.Economy.Upgrades() {
		if upgrade.Window.FromGate == "" && upgrade.Window.ToGate == "gate.t0_to_t1" {
			policy.Items = append(policy.Items, RelevancePolicyItem{PurchasableID: upgrade.ID})
		}
	}
	sort.Slice(policy.Items, func(left, right int) bool {
		return policy.Items[left].PurchasableID < policy.Items[right].PurchasableID
	})
	return &RelevanceSuite{Scenario: RelevanceScenario{HorizonMS: runtime.spec.HorizonMS,
		Milestone: RelevanceMilestone{ID: "milestone.garage_gate", Kind: "resource_at_least", ResourceID: gate.Requirement[0].ResourceID, Amount: gate.Requirement[0].Amount.String()}},
		Catalog: runtime.suite.Bundle.Economy, Routes: runtime.suite.Bundle.Routes, Policy: policy, ConstantsHash: runtime.suite.ConstantsHash}, nil
}

func (runtime *firstHourRuntime) observeReferencePurchase(atMS int64, id string) {
	kind := production.IntentBuyUpgrade
	eventKind := save.EventUpgradePurchased
	if _, ok := runtime.suite.Bundle.Economy.GeneratorClass(id); ok {
		kind, eventKind = production.IntentBuyGenerator, save.EventGeneratorPurchased
	}
	runtime.observeDecision(atMS, kind, save.IntentDecision{Events: []save.EventWrite{{Kind: eventKind}}})
}

func (runtime *firstHourRuntime) apply(request production.IntentRequest, now time.Time, atMS int64, mode production.EvaluationMode) error {
	runtime.recordCommand(atMS, mode, request)
	result, err := production.SimulateTransition(request, runtime.company, runtime.suite.Bundle.Economy,
		production.SimulationDependencies{Routes: runtime.suite.Bundle.Routes}, runtime.companyRevision(), mode, now, nil, nil, production.AblationMask{})
	if err != nil {
		return err
	}
	runtime.transitions++
	if result.Decision.Outcome != save.IntentApplied {
		return fmt.Errorf("first-hour ruled command %s rejected: %s", request.Kind, result.Decision.Receipt)
	}
	runtime.revision++
	runtime.observeDecision(atMS, request.Kind, result.Decision)
	return validateFirstHourCompany(runtime.suite.Bundle.Economy, runtime.company)
}

func (runtime *firstHourRuntime) recordCommand(atMS int64, mode production.EvaluationMode, request production.IntentRequest) {
	if runtime.commands != nil {
		*runtime.commands = append(*runtime.commands, FirstHourScriptCommand{AtMS: atMS, Mode: mode, Request: request})
	}
}

func (runtime *firstHourRuntime) applyEnding(now time.Time, wallMS, attended int64) error {
	terms, err := prestigecore.ComputeTerms(runtime.company, runtime.founder, runtime.suite.Bundle.Prestige, "scripted_first")
	if err != nil {
		return err
	}
	branch, cheapestID, cheapestPrice, err := runtime.selectBranch()
	if err != nil {
		return err
	}
	sample := &FirstHourEndingSample{PolicyID: runtime.policy.PolicyID, Seed: strconv.FormatUint(runtime.seed, 10), Branch: branch,
		GeneratorPurchasedTotal: runtime.company.GeneratorPurchasedTotal, UpgradeCount: countOwnedUpgrades(runtime.company),
		CheapestUnownedGenerator: cheapestID, CheapestUnownedPrice: cheapestPrice.String()}
	if cash, ok := runtime.company.Ledger.Balance("company.cash"); ok {
		sample.Cash = cash.String()
	}
	if gate := runtime.milestones["milestone.garage_gate"]; gate != nil {
		sample.RunOneGateMS = *gate
	}
	bonus := int64(0)
	if branch == "burnout" {
		bonus = runtime.experiment.RouteKnowledgeBonus
	} else if branch == "pivot" {
		bonus = runtime.experiment.RouteKnowledgeBonus / 2
	}
	if bonus > decimal.MaxExactInteger-runtime.founder.RouteKnowledgeBalance ||
		terms.RouteKnowledge > decimal.MaxExactInteger-runtime.founder.RouteKnowledgeBalance-bonus {
		return errors.New("first-hour Route Knowledge overflow")
	}
	runtime.founder.ReputationLevel += terms.ReputationDelta
	runtime.founder.RouteKnowledgeBalance += terms.RouteKnowledge + bonus
	runtime.founder.AgeMS += attended
	runtime.founder.ExitHistory = append(runtime.founder.ExitHistory, save.ExitRecord{RunID: 1, ExitType: "scripted_first", OccurredAt: now, ReputationDelta: terms.ReputationDelta})
	runtime.founderAttendedMS += attended
	runtime.observeRunEnded(wallMS, "scripted_first")
	next, err := prestigecore.NewRunState(runtime.suite.Bundle.Economy, runtime.company, runtime.founder, now)
	if err != nil {
		return err
	}
	if err := applyFirstHourExperimentStarter(next, branch, runtime.experiment); err != nil {
		return err
	}
	runtime.company, runtime.ending = next, sample
	runtime.revision++
	return validateFirstHourCompany(runtime.suite.Bundle.Economy, runtime.company)
}

func validateFirstHourCompany(catalog *economy.Catalog, state *save.State) error {
	if _, err := save.EncodeState(state); err != nil {
		return fmt.Errorf("first-hour state_encodes: %w", err)
	}
	if err := validateStateDomain(catalog, state); err != nil {
		return fmt.Errorf("first-hour numeric/resource domain: %w", err)
	}
	return nil
}

func (runtime *firstHourRuntime) selectBranch() (string, string, decimal.Decimal, error) {
	id, price, err := cheapestUnownedGenerator(runtime.suite.Bundle.Economy, runtime.company)
	if err != nil {
		return "", "", decimal.NaN, err
	}
	if runtime.company.GeneratorPurchasedTotal >= runtime.experiment.AcquihirePurchasedMinimum && countOwnedUpgrades(runtime.company) > 0 {
		return "acquihire", id, price, nil
	}
	factor, err := decimal.ParseCanonical(runtime.experiment.BurnoutPriceFactor)
	if err != nil || factor.Lt(decimal.Zero) {
		return "", "", decimal.NaN, errors.New("invalid first-hour burnout factor")
	}
	cash, _ := runtime.company.Ledger.Balance("company.cash")
	threshold := price.Mul(factor).Quantize(decimal.CanonicalSignificantDigits)
	if !threshold.IsStateValue() {
		return "", "", decimal.NaN, errors.New("first-hour burnout threshold exceeds Decimal state domain")
	}
	if cash.Lt(threshold) {
		return "burnout", id, price, nil
	}
	return "pivot", id, price, nil
}

func cheapestUnownedGenerator(catalog *economy.Catalog, state *save.State) (string, decimal.Decimal, error) {
	id, price := "", decimal.NaN
	for _, generator := range catalog.GeneratorClassesForScope(economy.ScopeCompany) {
		cost, err := catalog.BulkCost(generator.ID, state.GeneratorCounts[generator.ID], 1)
		if err != nil {
			return "", decimal.NaN, err
		}
		if id == "" || cost.Lt(price) || cost.Eq(price) && generator.ID < id {
			id, price = generator.ID, cost
		}
	}
	if id == "" {
		return "", decimal.NaN, errors.New("first-hour economy has no generator")
	}
	return id, price, nil
}

func applyFirstHourExperimentStarter(state *save.State, branch string, experiment FirstHourExperiment) error {
	switch branch {
	case "acquihire":
		amount, err := decimal.ParseCanonical(experiment.SeedCapital)
		if err != nil || !amount.Gt(decimal.Zero) {
			return errors.New("invalid first-hour seed capital")
		}
		_, err = state.Ledger.Apply(economy.Transaction{Entries: []economy.Entry{{ResourceID: "company.cash", Delta: amount}}})
		return err
	case "burnout":
		if experiment.GeneratedBeigeTowers < 1 || experiment.GeneratedBeigeTowers > decimal.MaxExactInteger {
			return errors.New("invalid first-hour generated tower count")
		}
		state.GeneratorProvisioned["generator.beige_tower"] = experiment.GeneratedBeigeTowers
		return nil
	case "pivot":
		state.UpgradesOwned["upgrade.reply_all_macro"] = true
		return nil
	default:
		return errors.New("invalid first-hour branch")
	}
}

func countOwnedUpgrades(state *save.State) int64 {
	count := int64(0)
	for _, owned := range state.UpgradesOwned {
		if owned {
			count++
		}
	}
	return count
}

func (runtime *firstHourRuntime) observeDecision(atMS int64, intentKind string, decision save.IntentDecision) {
	runAttended, _ := prestigecore.AttendedMS(runtime.company, Epoch.Add(time.Duration(atMS)*time.Millisecond))
	for _, milestone := range runtime.suite.Scenario.Milestones {
		if runtime.milestones[milestone.ID] != nil {
			continue
		}
		reached := milestone.Kind == "intent_applied" && milestone.IntentKind == intentKind
		if milestone.Kind == "event_seen" {
			for _, event := range decision.Events {
				reached = reached || string(event.Kind) == milestone.EventKind
			}
		}
		if milestone.Kind == "gate_crossed" {
			reached = runtime.company.RunSeq == milestone.RunSeq && runtime.company.GatesCrossed[milestone.GateID]
		}
		if reached {
			value := runAttended
			if milestone.Clock == "founder_attended_ms" {
				value += runtime.founderAttendedMS
			}
			runtime.milestones[milestone.ID] = &value
			if milestone.ID == "milestone.run2_garage_gate" && runtime.ending != nil {
				runtime.ending.RunTwoGateMS = value
			}
		}
	}
}

func (runtime *firstHourRuntime) observeRunEnded(atMS int64, exitType string) {
	attended, _ := prestigecore.AttendedMS(runtime.company, Epoch.Add(time.Duration(atMS)*time.Millisecond))
	for _, milestone := range runtime.suite.Scenario.Milestones {
		if runtime.milestones[milestone.ID] != nil || milestone.Kind != "run_ended" {
			continue
		}
		if milestone.RunSeq != 0 && milestone.RunSeq != runtime.company.RunSeq || milestone.ExitType != "" && milestone.ExitType != exitType || milestone.ExitTypeNot != "" && milestone.ExitTypeNot == exitType {
			continue
		}
		value := attended
		if milestone.Clock == "founder_attended_ms" {
			value += runtime.founderAttendedMS
		}
		runtime.milestones[milestone.ID] = &value
	}
}

func firstHourMilestoneReport(definitions []FirstHourMilestone, found map[string]*int64, failures *[]string) []TimedMilestone {
	result := make([]TimedMilestone, 0, len(definitions))
	for _, definition := range definitions {
		value := found[definition.ID]
		result = append(result, TimedMilestone{ID: definition.ID, FirstMS: cloneInt64(value)})
		if definition.MustReach && value == nil {
			*failures = append(*failures, "must_reach:"+definition.ID)
		}
	}
	return result
}

func failFirstHour(report FirstHourRunResult, err error) FirstHourRunResult {
	report.Outcome = "failed"
	report.InvariantFailures = append(report.InvariantFailures, err.Error())
	return report
}

func firstHourIntentID(revision int64) string {
	return fmt.Sprintf("0198aaaa-0000-7000-8000-%012d", revision%1_000_000_000_000)
}

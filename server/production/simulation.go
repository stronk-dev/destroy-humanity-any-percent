package production

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"cloud-clicker/server/accrualhook"
	"cloud-clicker/server/decimal"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/faction"
	"cloud-clicker/server/multiplier"
	"cloud-clicker/server/routes"
	"cloud-clicker/server/save"
)

// AblationMask is the closed, simulation-only policy consumed by the balance
// harness. It is deliberately absent from CatalogBundle and replay inputs.
type AblationMask struct {
	GeneratorIDs        []string
	UpgradeIDs          []string
	RemovedGeneratorIDs []string
	RemovedUpgradeIDs   []string
	RemovedActionIDs    []string
}

// SimulationDependencies are the real, already-owned mechanics needed by a
// simulated transition. They never enter CatalogBundle, save state, or replay
// inputs. Relevance supplies the same Routes catalog and accrual hook used by
// the route it simulates; the Phase-0 pacing harness needs Routes only.
type SimulationDependencies struct {
	Routes      *routes.Catalog
	CompactBand *CompactTitheBand
	Factions    *faction.Catalog
	Hook        AccrualHook
}

// RoleActivation is emitted only when a typed generator role participates in
// a non-neutral mechanic during an applied simulated transition.
type RoleActivation struct {
	GeneratorID string                    `json:"generator_id"`
	Kind        economy.GeneratorRoleKind `json:"kind"`
	TargetID    string                    `json:"target_id"`
}

type SimulationResult struct {
	Decision        save.IntentDecision
	RoleActivations []RoleActivation
}

// AdvanceSimulationResult is the action-free counterpart to SimulationResult.
// Evaluation is returned for harness accounting only; no intent receipt,
// event envelope, or revision is created by SimulateAdvance.
type AdvanceSimulationResult struct {
	Evaluation      EvaluationResult
	RoleActivations []RoleActivation
}

// SimulateResourceRate projects one resource's canonical current per-second
// production rate under a harness-only ablation mask. It shares the exact
// contribution and generator-rate assembly used by SimulateAdvance; callers
// must not reconstruct production arithmetic from catalog rows.
func SimulateResourceRate(state *save.State, catalog *economy.Catalog, resourceID string, mask AblationMask) (decimal.Decimal, error) {
	if state == nil || state.Ledger == nil || state.Ledger.Scope() != economy.ScopeCompany || catalog == nil {
		return decimal.NaN, ErrInvalidEngineState
	}
	resource, exists := catalog.Resource(resourceID)
	if !exists || resource.Scope != economy.ScopeCompany {
		return decimal.NaN, ErrInvalidEngineState
	}
	policy, err := simulationPolicyFor(catalog, mask)
	if err != nil {
		return decimal.NaN, err
	}
	contributions, err := assembleContributionsWithPolicy(state, catalog, nil, policy)
	if err != nil {
		return decimal.NaN, err
	}
	purchased, provisioned, err := projectionGeneratorCounts(catalog, state)
	if err != nil {
		return decimal.NaN, err
	}
	rates, err := ratesWithProvisionedAndPolicy(catalog, purchased, provisioned, contributions, policy)
	if err != nil {
		return decimal.NaN, err
	}
	return decimal.SumDeterministic(rates[resourceID]), nil
}

type simulationPolicy struct {
	generators        map[string]bool
	upgrades          map[string]bool
	removedGenerators map[string]bool
	removedUpgrades   map[string]bool
	removedActions    map[string]bool
	observeProduction bool
	active            map[string]RoleActivation
	candidates        map[string]RoleActivation
}

func simulationPolicyFor(catalog *economy.Catalog, mask AblationMask) (*simulationPolicy, error) {
	if catalog == nil {
		return nil, ErrInvalidEngineState
	}
	policy := &simulationPolicy{
		generators: map[string]bool{}, upgrades: map[string]bool{}, removedGenerators: map[string]bool{},
		removedUpgrades: map[string]bool{}, removedActions: map[string]bool{}, active: map[string]RoleActivation{}, candidates: map[string]RoleActivation{},
	}
	validate := func(values []string, exists func(string) bool, target map[string]bool, label string) error {
		ordered := append([]string(nil), values...)
		sort.Strings(ordered)
		for index, id := range ordered {
			if id == "" || index > 0 && ordered[index-1] == id || !exists(id) {
				return fmt.Errorf("%w: invalid %s mask %q", ErrInvalidEngineState, label, id)
			}
			target[id] = true
		}
		return nil
	}
	if err := validate(mask.GeneratorIDs, func(id string) bool { _, ok := catalog.GeneratorClass(id); return ok }, policy.generators, "generator"); err != nil {
		return nil, err
	}
	if err := validate(mask.UpgradeIDs, func(id string) bool { _, ok := catalog.Upgrade(id); return ok }, policy.upgrades, "upgrade"); err != nil {
		return nil, err
	}
	if err := validate(mask.RemovedGeneratorIDs, func(id string) bool { _, ok := catalog.GeneratorClass(id); return ok }, policy.removedGenerators, "removed generator"); err != nil {
		return nil, err
	}
	if err := validate(mask.RemovedUpgradeIDs, func(id string) bool { _, ok := catalog.Upgrade(id); return ok }, policy.removedUpgrades, "removed upgrade"); err != nil {
		return nil, err
	}
	if err := validate(mask.RemovedActionIDs, func(id string) bool { _, ok := catalog.ManualAction(id); return ok }, policy.removedActions, "action"); err != nil {
		return nil, err
	}
	return policy, nil
}

// SimulateTransition is the only entrypoint that can apply an ablation mask.
// A repository source guard restricts non-test callers to server/harness.
func SimulateTransition(request IntentRequest, state *save.State, catalog *economy.Catalog, dependencies SimulationDependencies, revision save.Revision, mode EvaluationMode, now time.Time, external []multiplier.Contribution, sink InvariantSink, mask AblationMask) (SimulationResult, error) {
	policy, err := simulationPolicyFor(catalog, mask)
	if err != nil {
		return SimulationResult{}, err
	}
	contributions, err := assembleContributionsWithPolicy(state, catalog, external, policy)
	if err != nil {
		return SimulationResult{}, err
	}
	decision, err := transitionWithSimulationPolicy(request, state, catalog, dependencies.Routes, nil, dependencies.CompactBand, dependencies.Factions, revision, mode, now, contributions, sink, simulationHook(dependencies.Hook, policy), policy)
	if err != nil {
		return SimulationResult{}, err
	}
	if decision.Outcome != save.IntentApplied {
		return SimulationResult{Decision: decision, RoleActivations: []RoleActivation{}}, nil
	}
	if request.Kind == IntentPerformManualBatch && appliedCountFromDecision(decision) > 0 {
		policy.activateManualRoles(state, catalog, request.ActionID)
		policy.promoteManualPoolCandidates(catalog, request.ActionID)
	}
	return SimulationResult{Decision: decision, RoleActivations: policy.activations()}, nil
}

// SimulateAdvance advances a cloned harness state without applying an action.
// It is the sole authoritative wait edge for relevance search and is restricted
// by the repository source guard to server/harness and tests.
func SimulateAdvance(state *save.State, catalog *economy.Catalog, dependencies SimulationDependencies, revision save.Revision, mode EvaluationMode, now time.Time, external []multiplier.Contribution, mask AblationMask) (AdvanceSimulationResult, error) {
	policy, err := simulationPolicyFor(catalog, mask)
	if err != nil {
		return AdvanceSimulationResult{}, err
	}
	contributions, err := assembleContributionsWithPolicy(state, catalog, external, policy)
	if err != nil {
		return AdvanceSimulationResult{}, err
	}
	result, err := evaluateWithSimulationPolicy(state, catalog, now, mode, contributions, policy)
	if err != nil {
		return AdvanceSimulationResult{}, err
	}
	if dependencies.Hook != nil && result.ElapsedMS > 0 {
		events, hookErr := simulationHook(dependencies.Hook, policy).AfterAccrual(state, catalog, revision,
			accrualhook.Result{Receipt: result.Receipt, ElapsedMS: result.ElapsedMS, ProductionMS: result.ProductionMS,
				BankedCreditMS: result.BankedCreditMS, ProgressDeltaPPM: result.ProgressDeltaPPM}, contributions)
		if hookErr != nil {
			return AdvanceSimulationResult{}, hookErr
		}
		// An action-free edge has no intent/event envelope. A hook may update its
		// replay-owned state, but any attempted event emission is invalid here.
		if len(events) != 0 {
			return AdvanceSimulationResult{}, ErrInvalidEngineState
		}
	}
	return AdvanceSimulationResult{Evaluation: result, RoleActivations: policy.activations()}, nil
}

func (policy *simulationPolicy) masksGenerator(id string) bool {
	return policy != nil && policy.generators[id]
}
func (policy *simulationPolicy) masksUpgrade(id string) bool {
	return policy != nil && policy.upgrades[id]
}
func (policy *simulationPolicy) removesAction(id string) bool {
	return policy != nil && policy.removedActions[id]
}

func (policy *simulationPolicy) removesGenerator(id string) bool {
	return policy != nil && policy.removedGenerators[id]
}

func (policy *simulationPolicy) removesUpgrade(id string) bool {
	return policy != nil && policy.removedUpgrades[id]
}

func roleActivationKey(value RoleActivation) string {
	return value.GeneratorID + "\x00" + string(value.Kind) + "\x00" + value.TargetID
}

func (policy *simulationPolicy) activate(value RoleActivation) {
	if policy == nil {
		return
	}
	policy.active[roleActivationKey(value)] = value
}

func (policy *simulationPolicy) candidate(value RoleActivation) {
	if policy == nil {
		return
	}
	policy.candidates[roleActivationKey(value)] = value
}

func (policy *simulationPolicy) activateSynergySource(sourceID string) {
	if policy == nil || !policy.observeProduction {
		return
	}
	for key, value := range policy.candidates {
		if value.Kind == economy.RoleSynergyFeed && value.TargetID == sourceID {
			policy.active[key] = value
		}
	}
}

func (policy *simulationPolicy) setProductionObservation(enabled bool) {
	if policy != nil {
		policy.observeProduction = enabled
	}
}

func (policy *simulationPolicy) activateManualRoles(state *save.State, catalog *economy.Catalog, actionID string) {
	if policy == nil || state == nil || catalog == nil {
		return
	}
	for _, generator := range catalog.GeneratorClassesForScope(economy.ScopeCompany) {
		if policy.masksGenerator(generator.ID) || state.GeneratorCounts[generator.ID] <= 0 {
			continue
		}
		for _, role := range generator.Roles {
			if role.Kind == economy.RoleManualOutput && role.ActionID == actionID {
				policy.activate(RoleActivation{GeneratorID: generator.ID, Kind: role.Kind, TargetID: role.ActionID})
			}
		}
	}
}

func (policy *simulationPolicy) promoteManualPoolCandidates(catalog *economy.Catalog, actionID string) {
	if policy == nil || catalog == nil {
		return
	}
	for key, value := range policy.candidates {
		pool, ok := catalog.SynergyPool(value.TargetID)
		if !ok {
			continue
		}
		source, ok := catalog.MultiplierSource(pool.ID)
		if ok && source.Target == actionID {
			policy.active[key] = value
		}
	}
}

func (policy *simulationPolicy) activations() []RoleActivation {
	result := make([]RoleActivation, 0, len(policy.active))
	for _, value := range policy.active {
		result = append(result, value)
	}
	sort.Slice(result, func(left, right int) bool { return roleActivationKey(result[left]) < roleActivationKey(result[right]) })
	return result
}

func appliedCountFromDecision(decision save.IntentDecision) int64 {
	var receipt struct {
		AppliedCount int64 `json:"applied_count"`
	}
	if json.Unmarshal(decision.Receipt, &receipt) != nil {
		return 0
	}
	return receipt.AppliedCount
}

type simulationAccrualHook struct {
	inner  AccrualHook
	policy *simulationPolicy
}

func simulationHook(inner AccrualHook, policy *simulationPolicy) AccrualHook {
	if inner == nil {
		return nil
	}
	return simulationAccrualHook{inner: inner, policy: policy}
}

func (hook simulationAccrualHook) AfterAccrual(state *save.State, catalog *economy.Catalog, revision save.Revision, result accrualhook.Result, contributions []multiplier.Contribution) ([]save.EventWrite, error) {
	if state == nil || catalog == nil {
		return nil, ErrInvalidEngineState
	}
	encoded, err := save.EncodeState(state)
	if err != nil {
		return nil, err
	}
	events, err := runSimulationHookWithMasks(hook.inner, state, catalog, revision, result, contributions, hook.policy.generators)
	if err != nil {
		return events, err
	}
	for _, generator := range catalog.GeneratorClassesForScope(economy.ScopeCompany) {
		if hook.policy.masksGenerator(generator.ID) || state.GeneratorCounts[generator.ID] <= 0 ||
			!generatorDeclaresRole(generator, economy.RoleStockRate, "faction.stock") {
			continue
		}
		counterfactual, restoreErr := save.RestoreState(encoded, save.CurrentVersion, catalog, economy.ScopeCompany, time.Time{})
		if restoreErr != nil {
			return nil, restoreErr
		}
		masks := make(map[string]bool, len(hook.policy.generators)+1)
		for id := range hook.policy.generators {
			masks[id] = true
		}
		masks[generator.ID] = true
		if _, counterfactualErr := runSimulationHookWithMasks(hook.inner, counterfactual, catalog, revision, result, contributions, masks); counterfactualErr != nil {
			return nil, counterfactualErr
		}
		if stockStateDiffers(state, counterfactual) {
			hook.policy.activate(RoleActivation{GeneratorID: generator.ID, Kind: economy.RoleStockRate, TargetID: "faction.stock"})
		}
	}
	return events, nil
}

func runSimulationHookWithMasks(hook AccrualHook, state *save.State, catalog *economy.Catalog, revision save.Revision, result accrualhook.Result, contributions []multiplier.Contribution, masks map[string]bool) ([]save.EventWrite, error) {
	original := make(map[string]int64, len(masks))
	for id := range masks {
		original[id] = state.GeneratorCounts[id]
		state.GeneratorCounts[id] = 0
	}
	events, err := hook.AfterAccrual(state, catalog, revision, result, contributions)
	for id, count := range original {
		state.GeneratorCounts[id] = count
	}
	return events, err
}

func stockStateDiffers(left, right *save.State) bool {
	return left.StockUnits != right.StockUnits || left.StockProgressMS != right.StockProgressMS ||
		left.StockRateRemainderPPM != right.StockRateRemainderPPM
}

package production

import (
	"fmt"
	"sort"
	"time"

	"cloud-clicker/server/economy"
	"cloud-clicker/server/multiplier"
	"cloud-clicker/server/save"
)

// AblationMask is the closed, simulation-only policy consumed by the balance
// harness. It is deliberately absent from CatalogBundle and replay inputs.
type AblationMask struct {
	GeneratorIDs     []string
	UpgradeIDs       []string
	RemovedActionIDs []string
}

type simulationPolicy struct {
	generators     map[string]bool
	upgrades       map[string]bool
	removedActions map[string]bool
}

func simulationPolicyFor(catalog *economy.Catalog, mask AblationMask) (*simulationPolicy, error) {
	if catalog == nil {
		return nil, ErrInvalidEngineState
	}
	policy := &simulationPolicy{generators: map[string]bool{}, upgrades: map[string]bool{}, removedActions: map[string]bool{}}
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
	if err := validate(mask.RemovedActionIDs, func(id string) bool { _, ok := catalog.ManualAction(id); return ok }, policy.removedActions, "action"); err != nil {
		return nil, err
	}
	return policy, nil
}

// SimulateTransition is the only entrypoint that can apply an ablation mask.
// A repository source guard restricts non-test callers to server/harness.
func SimulateTransition(request IntentRequest, state *save.State, catalog *economy.Catalog, revision save.Revision, mode EvaluationMode, now time.Time, external []multiplier.Contribution, sink InvariantSink, mask AblationMask) (save.IntentDecision, error) {
	policy, err := simulationPolicyFor(catalog, mask)
	if err != nil {
		return save.IntentDecision{}, err
	}
	contributions, err := assembleContributionsWithPolicy(state, catalog, external, policy)
	if err != nil {
		return save.IntentDecision{}, err
	}
	return transitionWithSimulationPolicy(request, state, catalog, nil, nil, nil, revision, mode, now, contributions, sink, nil, policy)
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

package prestige

import (
	"time"

	"cloud-clicker/server/accrualhook"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/multiplier"
	"cloud-clicker/server/save"
)

type PolicyResolver interface {
	ResolvePrestige(constantsHash string) (*Policy, bool)
}

type AccrualHook struct {
	Policies         PolicyResolver
	CatchupCeilingMS int64
}

func (hook AccrualHook) AfterAccrual(state *save.State, _ *economy.Catalog, revision save.Revision, result accrualhook.Result, _ []multiplier.Contribution) ([]save.EventWrite, error) {
	if hook.Policies == nil || hook.CatchupCeilingMS <= 0 || state == nil || result.ElapsedMS <= 0 {
		return nil, ErrInvalidPolicy
	}
	policy, ok := hook.Policies.ResolvePrestige(revision.ConstantsHash)
	if !ok {
		return nil, ErrInvalidPolicy
	}
	if err := AccumulateLifetimeValue(state, result.Receipt, policy.ValueResourceID); err != nil {
		return nil, err
	}
	to := state.EvaluatedThrough
	from := to.Add(-time.Duration(result.ElapsedMS) * time.Millisecond)
	if err := RecordOfflineSpan(state, from, to, hook.CatchupCeilingMS); err != nil {
		return nil, err
	}
	return nil, nil
}

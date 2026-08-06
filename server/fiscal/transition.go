package fiscal

import (
	"math/big"
	"sort"

	"cloud-clicker/server/decimal"
)

type State struct {
	Credit             int64
	PeriodOpenedWallMS int64
	PeriodSequence     int64
	GeneratorLevels    map[string]int64
	Unlocks            []string
}

type SweepResult struct {
	Periods          int64
	CreditBefore     int64
	Credited         int64
	CreditAfter      int64
	OpenedBeforeMS   int64
	OpenedAfterMS    int64
	SequenceBefore   int64
	SequenceAfter    int64
	Saturated        bool
	HardcapReasonKey string
}

type HarvestOutcome string

const (
	HarvestAutoReported   HarvestOutcome = "auto_reported"
	HarvestEarlySucceeded HarvestOutcome = "early_succeeded"
	HarvestEarlyFailed    HarvestOutcome = "early_failed"
	HarvestGuaranteed     HarvestOutcome = "guaranteed"
	HarvestConsumedByAuto HarvestOutcome = "consumed_by_auto"
)

type HarvestResult struct {
	Sweep                    *SweepResult
	NowWallMS                int64
	PeriodOpenedBeforeWallMS int64
	PeriodsSwept             int64
	SequenceBefore           int64
	DrawPPM                  *int64
	Outcome                  HarvestOutcome
	CreditBefore             int64
	CreditAfter              int64
	Saturated                bool
}

type SpendTarget struct {
	Kind        string
	GeneratorID string
	Levels      int64
	UnlockID    string
}

type SpendResult struct {
	Sweep        *SweepResult
	Target       SpendTarget
	ResolvedCost int64
	CreditBefore int64
	CreditAfter  int64
}

func (catalog *Catalog) ValidateState(state State) error {
	if catalog == nil || state.Credit < 0 || state.Credit > catalog.Credit.Hardcap || state.PeriodOpenedWallMS < 0 || state.PeriodOpenedWallMS > decimal.MaxExactInteger || state.PeriodSequence < 0 || state.PeriodSequence > decimal.MaxExactInteger || len(state.GeneratorLevels) != len(catalog.generatorRows) {
		return ErrInvalidState
	}
	for _, row := range catalog.generatorRows {
		level, ok := state.GeneratorLevels[row.GeneratorID]
		if !ok || level < 0 || level > row.LevelHardcap {
			return ErrInvalidState
		}
	}
	for id := range state.GeneratorLevels {
		if _, ok := catalog.generatorByID[id]; !ok {
			return ErrInvalidState
		}
	}
	if !sort.StringsAreSorted(state.Unlocks) {
		return ErrInvalidState
	}
	for index, id := range state.Unlocks {
		if _, ok := catalog.unlockByID[id]; !ok || index > 0 && state.Unlocks[index-1] == id {
			return ErrInvalidState
		}
	}
	return nil
}

func (catalog *Catalog) Sweep(state *State, nowWallMS int64) (*SweepResult, error) {
	if state == nil || catalog.ValidateState(*state) != nil || nowWallMS < 0 || nowWallMS > decimal.MaxExactInteger {
		return nil, ErrInvalidState
	}
	if nowWallMS < state.PeriodOpenedWallMS {
		return nil, ErrClockRegression
	}
	periods := (nowWallMS - state.PeriodOpenedWallMS) / catalog.Clock.AutoMS
	if periods == 0 {
		return nil, nil
	}
	if periods > decimal.MaxExactInteger-state.PeriodSequence {
		return nil, ErrInvalidState
	}
	beforeCredit, beforeOpened, beforeSequence := state.Credit, state.PeriodOpenedWallMS, state.PeriodSequence
	minted := new(big.Int).Mul(big.NewInt(periods), big.NewInt(catalog.Credit.CreditPerPeriod))
	headroom := big.NewInt(catalog.Credit.Hardcap - state.Credit)
	credited := new(big.Int).Set(minted)
	saturated := credited.Cmp(headroom) > 0
	if saturated {
		credited.Set(headroom)
	}
	state.Credit += credited.Int64()
	state.PeriodOpenedWallMS += periods * catalog.Clock.AutoMS
	state.PeriodSequence += periods
	return &SweepResult{Periods: periods, CreditBefore: beforeCredit, Credited: credited.Int64(), CreditAfter: state.Credit,
		OpenedBeforeMS: beforeOpened, OpenedAfterMS: state.PeriodOpenedWallMS, SequenceBefore: beforeSequence, SequenceAfter: state.PeriodSequence,
		Saturated: saturated, HardcapReasonKey: catalog.Credit.HardcapReasonKey}, nil
}

func (catalog *Catalog) Harvest(state *State, founderID string, nowWallMS int64) (HarvestResult, error) {
	if state == nil {
		return HarvestResult{}, ErrInvalidState
	}
	openedBefore, sequenceBefore, creditBefore := state.PeriodOpenedWallMS, state.PeriodSequence, state.Credit
	sweep, err := catalog.Sweep(state, nowWallMS)
	if err != nil {
		return HarvestResult{}, err
	}
	if sweep != nil {
		return HarvestResult{Sweep: sweep, NowWallMS: nowWallMS, PeriodOpenedBeforeWallMS: openedBefore, PeriodsSwept: sweep.Periods,
			SequenceBefore: sequenceBefore, Outcome: HarvestConsumedByAuto, CreditBefore: creditBefore, CreditAfter: state.Credit, Saturated: sweep.Saturated}, nil
	}
	elapsed := nowWallMS - state.PeriodOpenedWallMS
	if elapsed < catalog.Clock.EarlyMS {
		return HarvestResult{}, ErrNotRipe
	}
	result := HarvestResult{NowWallMS: nowWallMS, PeriodOpenedBeforeWallMS: openedBefore, SequenceBefore: sequenceBefore, CreditBefore: creditBefore}
	success := true
	if elapsed < catalog.Clock.GuaranteedMS {
		draw, drawErr := EarlyHarvestDraw(founderID, state.PeriodSequence)
		if drawErr != nil {
			return HarvestResult{}, drawErr
		}
		result.DrawPPM = &draw
		success = draw < catalog.Clock.EarlySuccessPPM
		if success {
			result.Outcome = HarvestEarlySucceeded
		} else {
			result.Outcome = HarvestEarlyFailed
		}
	} else {
		result.Outcome = HarvestGuaranteed
	}
	if state.PeriodSequence == decimal.MaxExactInteger {
		return HarvestResult{}, ErrInvalidState
	}
	state.PeriodOpenedWallMS = nowWallMS
	state.PeriodSequence++
	if success {
		headroom := catalog.Credit.Hardcap - state.Credit
		credited := minInt64(catalog.Credit.CreditPerPeriod, headroom)
		result.Saturated = credited < catalog.Credit.CreditPerPeriod
		state.Credit += credited
	}
	result.CreditAfter = state.Credit
	return result, nil
}

func (catalog *Catalog) Spend(state *State, nowWallMS int64, target SpendTarget) (SpendResult, error) {
	if state == nil {
		return SpendResult{}, ErrInvalidState
	}
	before := cloneState(*state)
	sweep, err := catalog.Sweep(state, nowWallMS)
	if err != nil {
		return SpendResult{}, err
	}
	result := SpendResult{Sweep: sweep, Target: target, CreditBefore: state.Credit}
	switch target.Kind {
	case "generator_level":
		if target.UnlockID != "" || target.GeneratorID == "" || target.Levels <= 0 {
			*state = before
			return SpendResult{}, ErrUnknownTarget
		}
		cost, costErr := catalog.GeneratorLevelCost(target.GeneratorID, state.GeneratorLevels[target.GeneratorID], target.Levels)
		if costErr != nil {
			*state = before
			return SpendResult{}, costErr
		}
		result.ResolvedCost = cost
		if state.Credit < cost {
			*state = before
			return SpendResult{}, ErrUnaffordable
		}
		state.Credit -= cost
		state.GeneratorLevels[target.GeneratorID] += target.Levels
	case "unlock":
		if target.GeneratorID != "" || target.Levels != 0 || target.UnlockID == "" {
			*state = before
			return SpendResult{}, ErrUnknownTarget
		}
		row, ok := catalog.Unlock(target.UnlockID)
		if !ok {
			*state = before
			return SpendResult{}, ErrUnknownTarget
		}
		index := sort.SearchStrings(state.Unlocks, target.UnlockID)
		if index < len(state.Unlocks) && state.Unlocks[index] == target.UnlockID {
			*state = before
			return SpendResult{}, ErrAlreadyUnlocked
		}
		result.ResolvedCost = row.Cost
		if state.Credit < row.Cost {
			*state = before
			return SpendResult{}, ErrUnaffordable
		}
		state.Credit -= row.Cost
		state.Unlocks = append(state.Unlocks, "")
		copy(state.Unlocks[index+1:], state.Unlocks[index:])
		state.Unlocks[index] = target.UnlockID
	default:
		*state = before
		return SpendResult{}, ErrUnknownTarget
	}
	result.CreditAfter = state.Credit
	return result, nil
}

func cloneState(source State) State {
	result := source
	result.GeneratorLevels = make(map[string]int64, len(source.GeneratorLevels))
	for key, value := range source.GeneratorLevels {
		result.GeneratorLevels[key] = value
	}
	result.Unlocks = append([]string(nil), source.Unlocks...)
	return result
}

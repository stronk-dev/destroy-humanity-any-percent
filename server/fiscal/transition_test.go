package fiscal

import (
	"errors"
	"testing"
)

func initialState() State {
	return State{PeriodOpenedWallMS: 1000, GeneratorLevels: map[string]int64{"generator.beige_tower": 0}, Unlocks: []string{}}
}

func TestSweepIsPhasePreservingAndSaturating(t *testing.T) {
	catalog, err := LoadCatalog(loadSharedFixture(t).Baseline, fiscalEconomy(t))
	if err != nil {
		t.Fatal(err)
	}
	state := initialState()
	state.Credit = 998
	result, err := catalog.Sweep(&state, 1901)
	if err != nil {
		t.Fatal(err)
	}
	if result.Periods != 3 || result.Credited != 2 || !result.Saturated || state.Credit != 1000 || state.PeriodOpenedWallMS != 1900 || state.PeriodSequence != 3 {
		t.Fatalf("unexpected sweep: %#v state=%#v", result, state)
	}
	before := cloneState(state)
	if _, err := catalog.Sweep(&state, 1899); !errors.Is(err, ErrClockRegression) || state.Credit != before.Credit {
		t.Fatalf("clock regression mutated state: %v %#v", err, state)
	}
}

func TestHarvestBoundariesAndRisk(t *testing.T) {
	catalog, err := LoadCatalog(loadSharedFixture(t).Baseline, fiscalEconomy(t))
	if err != nil {
		t.Fatal(err)
	}
	state := initialState()
	if _, err := catalog.Harvest(&state, "founder.one", 1099); !errors.Is(err, ErrNotRipe) {
		t.Fatalf("early boundary error=%v", err)
	}
	state = initialState()
	result, err := catalog.Harvest(&state, "founder.one", 1200)
	if err != nil || result.Outcome != HarvestGuaranteed || state.Credit != 3 || state.PeriodOpenedWallMS != 1200 || state.PeriodSequence != 1 {
		t.Fatalf("guaranteed result=%#v state=%#v err=%v", result, state, err)
	}
	state = initialState()
	result, err = catalog.Harvest(&state, "founder.one", 1300)
	if err != nil || result.Outcome != HarvestConsumedByAuto || result.PeriodsSwept != 1 || state.Credit != 3 {
		t.Fatalf("auto result=%#v state=%#v err=%v", result, state, err)
	}
	state = initialState()
	result, err = catalog.Harvest(&state, "founder.one", 1100)
	if err != nil || result.DrawPPM == nil || result.Outcome != HarvestEarlySucceeded && result.Outcome != HarvestEarlyFailed || state.PeriodSequence != 1 || state.PeriodOpenedWallMS != 1100 {
		t.Fatalf("risky result=%#v state=%#v err=%v", result, state, err)
	}
}

func TestRejectedSpendRollsBackSweep(t *testing.T) {
	catalog, err := LoadCatalog(loadSharedFixture(t).Baseline, fiscalEconomy(t))
	if err != nil {
		t.Fatal(err)
	}
	state := initialState()
	before := cloneState(state)
	if _, err := catalog.Spend(&state, 1300, SpendTarget{Kind: "unlock", UnlockID: "unlock.arcade"}); !errors.Is(err, ErrUnaffordable) {
		t.Fatalf("error=%v", err)
	}
	if state.Credit != before.Credit || state.PeriodOpenedWallMS != before.PeriodOpenedWallMS || state.PeriodSequence != before.PeriodSequence {
		t.Fatalf("rejected spend committed sweep: %#v", state)
	}
	state.Credit = 20
	result, err := catalog.Spend(&state, 1600, SpendTarget{Kind: "generator_level", GeneratorID: "generator.beige_tower", Levels: 3})
	if err != nil || result.ResolvedCost != 6 || state.Credit != 20 {
		t.Fatalf("swept spend result=%#v state=%#v err=%v", result, state, err)
	}
}

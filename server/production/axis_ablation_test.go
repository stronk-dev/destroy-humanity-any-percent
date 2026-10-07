package production

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	"cloud-clicker/server/achievements"
	"cloud-clicker/server/save"
)

type axisAblationRow struct {
	name    string
	input   int64
	owned   int
	rates   [4]string
	offline [4]string
}

// Independent literals: beige base=1, pool=1.001; at x8 PR1=1.2,
// at x12 PR1=1.3/PR2=1.24, at x44 PR1=2.1/PR2=1.88. Masks
// remove contributions, not ownership or the production base/pool.
func axisAblationRows() []axisAblationRow {
	return []axisAblationRow{
		{"empty", 0, 0, [4]string{"1.001e0", "1.001e0", "1.001e0", "1.001e0"}, [4]string{"9.009e-1", "9.009e-1", "9.009e-1", "9.009e-1"}},
		{"one", 8, 1, [4]string{"1.2012e0", "1.001e0", "1.2012e0", "1.001e0"}, [4]string{"1.08108e0", "9.009e-1", "1.08108e0", "9.009e-1"}},
		{"two", 12, 2, [4]string{"1.613612e0", "1.24124e0", "1.3013e0", "1.001e0"}, [4]string{"1.4522508e0", "1.117116e0", "1.17117e0", "9.009e-1"}},
		{"cap", 44, 2, [4]string{"3.951948e0", "1.88188e0", "2.1021e0", "1.001e0"}, [4]string{"3.5567532e0", "1.693692e0", "1.89189e0", "9.009e-1"}},
	}
}

func axisAblationState(t *testing.T, bundle CatalogBundle, start time.Time, row axisAblationRow) *save.State {
	t.Helper()
	state := axisTimingState(t, bundle, start)
	state.UpgradesOwned = map[string]bool{}
	if row.owned >= 1 {
		state.UpgradesOwned["upgrade.pr_intern_1"] = true
	}
	if row.owned == 2 {
		state.UpgradesOwned["upgrade.pr_intern_2"] = true
	}
	state.AchievementsAttainedRun = map[string]bool{}
	if row.input > 0 {
		for _, id := range []string{"achievement.first_gate", "achievement.generators_purchased_1", "achievement.generators_purchased_25"} {
			state.AchievementsAttainedRun[id] = true
		}
	}
	if row.input == 12 {
		state.AchievementsAttainedRun["achievement.generators_owned_100"] = true
	}
	if row.input == 44 {
		for _, definition := range bundle.Achievements.Definitions {
			if definition.ConditionScope == achievements.ScopeRun {
				state.AchievementsAttainedRun[definition.ID] = true
			}
		}
	}
	state.AttainmentScoreRun = row.input
	setCash(t, state, "0")
	if err := bundle.ValidateFoundationState(state); err != nil {
		t.Fatalf("ablation full pinned admission: %v", err)
	}
	return state
}

func axisAblationEncoded(t *testing.T, state *save.State) []byte {
	t.Helper()
	encoded, err := save.EncodeState(state)
	if err != nil {
		t.Fatal(err)
	}
	return encoded
}

func TestAxisSimulationEffectMask(t *testing.T) {
	bundle := axisContentBundle(t)
	start := time.Date(2026, 10, 7, 12, 0, 0, 0, time.UTC)
	masks := []AblationMask{
		{}, {UpgradeIDs: []string{"upgrade.pr_intern_1"}},
		{UpgradeIDs: []string{"upgrade.pr_intern_2"}},
		{UpgradeIDs: []string{"upgrade.pr_intern_2", "upgrade.pr_intern_1"}},
	}
	for _, row := range axisAblationRows() {
		for index, mask := range masks {
			t.Run(fmt.Sprintf("%s/mask%d", row.name, index), func(t *testing.T) {
				state := axisAblationState(t, bundle, start, row)
				before := axisAblationEncoded(t, state)
				rate, err := SimulateResourceRate(state, bundle.Economy, "company.cash", mask)
				if err != nil || rate.String() != row.rates[index] {
					// Keep the independent advance arms running even if the rate
					// observer detects a fault; neither oracle can mask the other.
					t.Errorf("masked actual rate=%s want=%s err=%v", rate, row.rates[index], err)
				}
				if !bytes.Equal(before, axisAblationEncoded(t, state)) {
					t.Fatal("rate projection mutated full encoded Company")
				}
				for _, mode := range []EvaluationMode{ModeOnline, ModeOffline} {
					t.Run(string(mode), func(t *testing.T) {
						actual := axisAblationState(t, bundle, start, row)
						expected := axisAblationState(t, bundle, start, row)
						cash := row.rates[index]
						if mode == ModeOffline {
							cash = row.offline[index]
						}
						setCash(t, expected, cash)
						at := start.Add(time.Second)
						expected.EvaluatedThrough = at
						result, err := SimulateAdvance(actual, bundle.Economy, SimulationDependencies{}, save.Revision{Number: 1}, mode, at, nil, mask)
						if err != nil {
							t.Fatal(err)
						}
						if !bytes.Equal(axisAblationEncoded(t, actual), axisAblationEncoded(t, expected)) {
							t.Fatalf("masked full state mismatch: cash=%s want=%s actual_sha=%x expected_sha=%x", actual.Ledger.Snapshot()["company.cash"], cash,
								sha256.Sum256(axisAblationEncoded(t, actual)), sha256.Sum256(axisAblationEncoded(t, expected)))
						}
						if result.Evaluation.ElapsedMS != 1000 || result.Evaluation.ProductionMS != 1000 || result.Evaluation.BankedCreditMS != 0 {
							t.Fatalf("unexpected simulation interval %+v", result.Evaluation)
						}
						if index == 0 {
							live := axisAblationState(t, bundle, start, row)
							contributions, err := assembleContributions(live, bundle.Economy, nil)
							if err != nil {
								t.Fatal(err)
							}
							liveResult, err := Evaluate(live, bundle.Economy, at, mode, contributions)
							if err != nil || !reflect.DeepEqual(liveResult, result.Evaluation) || !bytes.Equal(axisAblationEncoded(t, live), axisAblationEncoded(t, actual)) {
								t.Fatalf("unmasked simulation diverged from live Evaluate: %v", err)
							}
						}
					})
				}
			})
		}
	}
}

func TestAxisSimulationInvalidEffectMasks(t *testing.T) {
	bundle := axisContentBundle(t)
	start := time.Date(2026, 10, 7, 12, 0, 0, 0, time.UTC)
	for _, ids := range [][]string{{"upgrade.absent"}, {"upgrade.pr_intern_1", "upgrade.pr_intern_1"}} {
		t.Run(fmt.Sprint(ids), func(t *testing.T) {
			state := axisAblationState(t, bundle, start, axisAblationRows()[2])
			before := axisAblationEncoded(t, state)
			mask := AblationMask{UpgradeIDs: ids}
			if _, err := SimulateResourceRate(state, bundle.Economy, "company.cash", mask); !errors.Is(err, ErrInvalidEngineState) {
				t.Fatalf("invalid rate mask admitted: %v", err)
			}
			if _, err := SimulateAdvance(state, bundle.Economy, SimulationDependencies{}, save.Revision{Number: 1}, ModeOnline, start.Add(time.Second), nil, mask); !errors.Is(err, ErrInvalidEngineState) {
				t.Fatalf("invalid advance mask admitted: %v", err)
			}
			if !bytes.Equal(before, axisAblationEncoded(t, state)) {
				t.Fatal("invalid mask mutated state")
			}
		})
	}
}

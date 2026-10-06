package production

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"cloud-clicker/server/decimal"
	"cloud-clicker/server/meters"
	"cloud-clicker/server/save"
)

// Admitted Company state, not a synthetic factor-only object. Historical
// attainment is latched; the first purchase in this population re-attains the
// still-unattained purchase_1 row and changes the input from 6 to 8.
func axisTimingState(t *testing.T, bundle CatalogBundle, now time.Time) *save.State {
	t.Helper()
	state := replayFixtureState(t, bundle.Economy, now)
	state.WireVersion, state.RunSeq, state.Tier = 19, 2, 1
	state.GatesCrossed["gate.t0_to_t1"] = true
	state.GeneratorCounts["generator.beige_tower"] = 1
	state.UpgradesOwned["upgrade.pr_intern_1"] = true
	state.AchievementsEarnedRun = map[string]bool{}
	state.AchievementsAttainedRun = map[string]bool{
		"achievement.first_gate": true, "achievement.generators_purchased_25": true,
	}
	state.AttainmentScoreRun = 6
	meterState, err := meters.NewRunState(bundle.Meters, 0)
	if err != nil {
		t.Fatal(err)
	}
	state.MeterBands = nil
	state.MeterValues, state.MeterDecayRemainders, state.MeterInputRemainders = meterState.Values, meterState.DecayRemainders, meterState.InputRemainders
	setCash(t, state, "1e4")
	if _, err := initializeActivePlayState(state, bundle.Opportunities, "01986666-2000-7000-8000-000000000001"); err != nil {
		t.Fatal(err)
	}
	if err := bundle.ValidateFoundationState(state); err != nil {
		t.Fatalf("axis timing admission: %v", err)
	}
	return state
}

func axisTimingPurchase(t *testing.T, bundle CatalogBundle, state *save.State, now time.Time, mode EvaluationMode) LoggedTransition {
	t.Helper()
	request, err := ParseIntent([]byte(`{"intent_id":"01986666-0901-7000-8000-000000000901","kind":"buy_generator","expected_revision":1,"generator_id":"generator.beige_tower","count":{"mode":"exact","value":1}}`))
	if err != nil {
		t.Fatal(err)
	}
	command := save.ReplayCommand{IntentID: request.IntentID, CompanyStreamID: "01986666-1000-7000-8000-000000000001",
		FounderID: "01986666-2000-7000-8000-000000000001", Revision: 1, RunSeq: 2, RunLogSeq: 1}
	carry := tier2FounderCarry(t, bundle, state.RunStartedAt)
	// A fully awarded veteran prevents first-earn (including career) output
	// from being mistaken for Company attainment in this diagnostic arm.
	carry.AchievementsEarnedLifetime = []string{}
	carry.AchievementScoreLifetime = 0
	for _, definition := range bundle.Achievements.Definitions {
		carry.AchievementsEarnedLifetime = append(carry.AchievementsEarnedLifetime, definition.ID)
		carry.AchievementScoreLifetime += definition.ScoreGrant
	}
	active, err := resolveActivePlaySchedule(state, bundle.Opportunities, bundle.Prestige, command.FounderID, now)
	if err != nil {
		t.Fatal(err)
	}
	inputs, err := buildReplayInputs(replayBuild{Command: command, Mode: mode, Now: now, IntentKind: request.Kind,
		RouteContextVersion: bundle.Routes.ContextVersion(), FounderCarry: &carry, ActivePlay: &active})
	if err != nil {
		t.Fatal(err)
	}
	transition, err := ApplyLogged(state, request.CanonicalPayload, bundle, inputs)
	if err != nil || transition.Outcome != save.IntentApplied {
		t.Fatalf("axis timing purchase outcome=%s receipt=%s error=%v", transition.Outcome, transition.Receipt, err)
	}
	return transition
}

func TestAxisAttainmentChangesOnlyFutureAccrual(t *testing.T) {
	bundle := axisContentBundle(t)
	start := time.Date(2026, 10, 7, 12, 0, 0, 0, time.UTC)
	random := rand.New(rand.NewSource(307))
	for policy := 0; policy < 32; policy++ {
		// Whole seconds keep this timing witness independent of the separately
		// tested ledger-quantization partition property.
		beforeSeconds, afterSeconds := int64(1+random.Intn(4)), int64(1+random.Intn(4))
		for _, mode := range []EvaluationMode{ModeOnline, ModeOffline} {
			t.Run(fmt.Sprintf("policy_%02d_%s", policy, mode), func(t *testing.T) {
				state := axisTimingState(t, bundle, start)
				at := start.Add(time.Duration(beforeSeconds) * time.Second)
				transition := axisTimingPurchase(t, bundle, state, at, mode)
				// Independent fixture arithmetic: 1 unit * (1+.001 pool) *
				// (1+6*.025 axis), then the 11.3 second-unit purchase debit.
				efficiency := decimal.One
				if mode == ModeOffline {
					efficiency = mustDecimal(t, "9e-1")
				}
				want := mustDecimal(t, "1e4").Add(mustDecimal(t, "1.15115e0").Mul(decimal.FromFloat64(float64(beforeSeconds))).Mul(efficiency)).Sub(mustDecimal(t, "1.13e1")).Quantize(decimal.CanonicalSignificantDigits)
				if got := state.Ledger.Snapshot()["company.cash"]; got != want.String() {
					t.Fatalf("triggering interval used future factor: cash=%s want=%s", got, want)
				}
				if state.AttainmentScoreRun != 8 || state.AchievementScoreRun != 0 || state.GeneratorCounts["generator.beige_tower"] != 2 {
					t.Fatalf("purchase did not re-attain: input=%d earned=%d count=%d", state.AttainmentScoreRun, state.AchievementScoreRun, state.GeneratorCounts["generator.beige_tower"])
				}
				var receipt struct {
					Revision int64 `json:"new_revision"`
					Snapshot struct {
						Attainment int64 `json:"attainment_score_run"`
					} `json:"snapshot"`
				}
				if err := json.Unmarshal(transition.Receipt, &receipt); err != nil || receipt.Revision != 2 || receipt.Snapshot.Attainment != 8 {
					t.Fatalf("receipt did not expose committed attainment: %s error=%v", transition.Receipt, err)
				}
				var reattained []string
				for _, event := range transition.Events {
					if event.Kind == save.EventAchievementReattained {
						var payload struct {
							ID string `json:"achievement_id"`
						}
						if err := json.Unmarshal(event.Payload, &payload); err != nil {
							t.Fatal(err)
						}
						reattained = append(reattained, payload.ID)
					}
					if event.Kind == save.EventAchievementEarned {
						t.Fatal("veteran purchase newly earned an achievement")
					}
				}
				if len(reattained) != 1 || reattained[0] != "achievement.generators_purchased_1" {
					t.Fatalf("reattainment order=%v", reattained)
				}
				contributions, err := assembleContributions(state, bundle.Economy, nil)
				if err != nil {
					t.Fatal(err)
				}
				if _, err := Evaluate(state, bundle.Economy, at.Add(time.Duration(afterSeconds)*time.Second), mode, contributions); err != nil {
					t.Fatal(err)
				}
				// New interval: 2 units * (1+.002 pool) * (1+8*.025 axis).
				want = want.Add(mustDecimal(t, "2.4048e0").Mul(decimal.FromFloat64(float64(afterSeconds))).Mul(efficiency)).Quantize(decimal.CanonicalSignificantDigits)
				if got := state.Ledger.Snapshot()["company.cash"]; got != want.String() {
					t.Fatalf("future interval ignored new input: cash=%s want=%s", got, want)
				}
			})
		}
	}
}

func TestAxisAccrualPartitionProperty(t *testing.T) {
	bundle := axisContentBundle(t)
	start := time.Date(2026, 10, 7, 12, 0, 0, 0, time.UTC)
	random := rand.New(rand.NewSource(307))
	for policy := 0; policy < 64; policy++ {
		endMS := int64(2000 + random.Intn(2000))
		cutMS := int64(1 + random.Intn(int(endMS-1)))
		if policy == 0 {
			endMS, cutMS = 2000, 1000
		} // whole-second control
		if policy == 1 {
			endMS, cutMS = 2001, 1
		} // tiny-cut control
		for _, mode := range []EvaluationMode{ModeOnline, ModeOffline} {
			t.Run(fmt.Sprintf("policy_%02d_%s_end%d_cut%d", policy, mode, endMS, cutMS), func(t *testing.T) {
				one, split := axisTimingState(t, bundle, start), axisTimingState(t, bundle, start)
				contributions, err := assembleContributions(one, bundle.Economy, nil)
				if err != nil {
					t.Fatal(err)
				}
				if _, err := Evaluate(one, bundle.Economy, start.Add(time.Duration(endMS)*time.Millisecond), mode, contributions); err != nil {
					t.Fatal(err)
				}
				for _, atMS := range []int64{cutMS, endMS} {
					if _, err := Evaluate(split, bundle.Economy, start.Add(time.Duration(atMS)*time.Millisecond), mode, contributions); err != nil {
						t.Fatal(err)
					}
				}
				if a, b := mustEncodeState(t, one), mustEncodeState(t, split); !bytes.Equal(a, b) {
					t.Fatalf("AC6 partition divergence: one=%s split=%s (full encoded states differ)", one.Ledger.Snapshot()["company.cash"], split.Ledger.Snapshot()["company.cash"])
				}
			})
		}
	}
}

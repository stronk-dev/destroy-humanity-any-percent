package production

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
	"time"

	"cloud-clicker/server/decimal"
	"cloud-clicker/server/guild"
	"cloud-clicker/server/save"
)

func TestAxisCalculationClockBoundariesResearch(t *testing.T) {
	bundle := axisContentBundle(t)
	start := time.Date(2026, 10, 15, 12, 0, 0, 0, time.UTC)
	const capMS = int64(86_400_000)
	if bundle.Prestige.CatchupCeilingMS != 5000 || bundle.Economy.OfflinePolicy().AccrualCapMS != capMS {
		t.Fatal("the two pinned clock thresholds changed")
	}
	accrual, err := makeReplayAccrual(nil, nil, guild.SettlementBatch{}, bundle.Routes.ContextVersion())
	if err != nil {
		t.Fatal(err)
	}
	cells := 0
	for _, buffed := range []bool{true, false} {
		for _, mode := range []EvaluationMode{ModeOnline, ModeOffline} {
			cash := map[int64]decimal.Decimal{}
			for _, offset := range []int64{999, 1000, 4999, 5000, 5001, capMS - 1, capMS, capMS + 1, capMS + 2} {
				t.Run(fmt.Sprintf("buffed=%t/%s/%dms", buffed, mode, offset), func(t *testing.T) {
					initial := axisTimingState(t, bundle, start)
					initial.RunStartedAt = start.Add(-4 * time.Second)
					effect, ok := bundle.Opportunities.Effect("active.production")
					if !ok || effect.DurationMS != 5000 || effect.Factor.String() != "7e0" {
						t.Fatal("pinned production effect changed")
					}
					initial.ActiveBuffs = []save.ActiveBuff{}
					if buffed {
						initial.ActiveBuffs = []save.ActiveBuff{{BuffInstanceID: "01985555-0000-7000-8000-000000000001",
							EffectRowID: effect.ID, ActivatedAttendedMS: 0, ExpiresAttendedMS: 5000}}
					}
					if err := bundle.ValidateFoundationState(initial); err != nil {
						t.Fatal(err)
					}
					origin := mustEncodeState(t, initial)
					at := start.Add(time.Duration(offset) * time.Millisecond)
					baseline := activationRestore(t, bundle, origin, 19, start)
					request, inputs := originManualInputs(t, bundle, baseline, at, mode, accrual, "manual.click")
					var wire replayInputsWire
					if err := json.Unmarshal(inputs, &wire); err != nil {
						t.Fatal(err)
					}
					wire.OfflineCatchup, err = buildOfflineCatchup(initial, bundle.Economy, mode, at)
					wantCatchup := mode == ModeOnline && offset > capMS
					if err != nil || (wire.OfflineCatchup != nil) != wantCatchup {
						t.Fatalf("wrong catchup boundary: %v", err)
					}
					inputs, err = json.Marshal(wire)
					if err != nil {
						t.Fatal(err)
					}
					var resolved replayAccrualResolved
					if err := json.Unmarshal(wire.Resolved, &resolved); err != nil || resolved.ActivePlay == nil {
						t.Fatalf("missing schedule: %v", err)
					}
					attended := int64(4000)
					if offset <= 5000 {
						attended += offset
					}
					if resolved.ActivePlay.AttendedNowMS != attended {
						t.Fatalf("attendance boundary: got%d want%d", resolved.ActivePlay.AttendedNowMS, attended)
					}
					ordinary, err := ApplyLogged(baseline, request.CanonicalPayload, bundle, inputs)
					if err != nil || ordinary.Outcome != save.IntentApplied {
						t.Fatalf("actual logged transition: %v outcome%s", err, ordinary.Outcome)
					}
					var rows []calculationStageObservationResearch
					observedBundle := bundle
					observedBundle.Commons = calculationStageObserverResearch{ReplayCommonsPolicy: bundle.Commons, t: t, rows: &rows}
					observed := activationRestore(t, bundle, origin, 19, start)
					result, err := ApplyLogged(observed, request.CanonicalPayload, observedBundle, inputs)
					if err != nil || result.Outcome != save.IntentApplied || len(rows) != 1 ||
						!bytes.Equal(mustEncodeState(t, observed), mustEncodeState(t, baseline)) ||
						!bytes.Equal(result.Receipt, ordinary.Receipt) || !reflect.DeepEqual(result.Events, ordinary.Events) {
						t.Fatalf("forwarding observer changed the real command: %v rows%d", err, len(rows))
					}
					productionMS, bankedMS := offset, int64(0)
					if mode == ModeOffline || wantCatchup {
						if productionMS > capMS {
							productionMS = capMS
						}
						bankedMS = (offset - productionMS) / 2
					}
					factor, efficiencyTenths := int64(1), int64(10)
					if buffed && (offset < 1000 || offset > 5000) {
						factor = 7
					}
					if mode == ModeOffline || wantCatchup {
						efficiencyTenths = 9
					}
					if wantCatchup {
						factor = 1 // Existing catchup resolves base, not endpoint buff inputs.
					}
					// Exact nanocash before the existing canonical state quantization:
					// 10000 + (115115/100000 per second) * ms/1000 * factor * efficiency.
					nanocash := int64(10000*1_000_000_000) + 115115*productionMS*factor*efficiencyTenths
					want := decimal.New(float64(nanocash), -9).Quantize(decimal.CanonicalSignificantDigits)
					stage := activationRestore(t, bundle, rows[0].State, 19, at)
					value := mustDecimal(t, stage.Ledger.Snapshot()["company.cash"])
					if !value.Eq(want) || rows[0].Result.ElapsedMS != offset || rows[0].Result.ProductionMS != productionMS ||
						rows[0].Result.BankedCreditMS != bankedMS || stage.ComputeCreditMS != bankedMS ||
						!bytes.Equal(origin, mustEncodeState(t, initial)) {
						t.Fatalf("wrong stage payout/clock or consumed origin: cash%s want%s result%+v", value, want, rows[0].Result)
					}
					cash[offset] = value
					cells++
					t.Logf("attended%d catchup%t cash%s production%d bank%d", attended, wantCatchup, value, productionMS, bankedMS)
				})
			}
			if len(cash) != 9 {
				t.Fatal("incomplete mode population")
			}
			if !cash[5001].Gt(cash[5000]) {
				t.Fatal("attendance-boundary context change was lost")
			}
			if mode == ModeOnline && !cash[capMS+1].Lt(cash[capMS]) ||
				mode == ModeOffline && !cash[capMS+1].Eq(cash[capMS]) {
				t.Fatal("online catchup drop or direct-offline equal control was lost")
			}
		}
	}
	if cells != 36 {
		t.Fatalf("incomplete boundary census: %d", cells)
	}
}

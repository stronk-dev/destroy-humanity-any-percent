package production

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
	"time"

	"cloud-clicker/server/decimal"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/guild"
	"cloud-clicker/server/save"
)

func TestAxisCalculationClockSearchResearch(t *testing.T) {
	bundle := axisContentBundle(t)
	start := time.Date(2026, 10, 15, 12, 0, 0, 0, time.UTC)
	const capMS = int64(86_400_000)
	const horizonMS = capMS + 2
	target := mustDecimal(t, "1e5") // Hypothetical threshold, not a catalog price.
	cells, onlineMisses, offlineControls, probeCount := 0, 0, 0, 0
	for _, buffed := range []bool{true, false} {
		for _, mode := range []EvaluationMode{ModeOnline, ModeOffline} {
			t.Run(fmt.Sprintf("buffed=%t/%s", buffed, mode), func(t *testing.T) {
				initial := axisTimingState(t, bundle, start)
				initial.RunStartedAt = start.Add(-4 * time.Second)
				initial.ActiveBuffs = []save.ActiveBuff{}
				if buffed {
					initial.ActiveBuffs = []save.ActiveBuff{{BuffInstanceID: "01985555-0000-7000-8000-000000000001",
						EffectRowID: "active.production", ActivatedAttendedMS: 0, ExpiresAttendedMS: 5000}}
				}
				if err := bundle.ValidateFoundationState(initial); err != nil {
					t.Fatal(err)
				}
				origin := mustEncodeState(t, initial)
				accrual, err := makeReplayAccrual(nil, nil, guild.SettlementBatch{}, bundle.Routes.ContextVersion())
				if err != nil {
					t.Fatal(err)
				}
				queries, rejectedPayoutPredictions := 0, 0
				qualifies := func(offset int64) bool {
					queries++ // No memoization: all old/new/control work is counted.
					if offset == 0 {
						balance, _ := initial.Ledger.Balance("company.cash")
						return balance.Gte(target)
					}
					state := activationRestore(t, bundle, origin, 19, start)
					at := start.Add(time.Duration(offset) * time.Millisecond)
					request, inputs := originManualInputs(t, bundle, state, at, mode, accrual, "manual.click")
					var wire replayInputsWire
					if err := json.Unmarshal(inputs, &wire); err != nil {
						t.Fatal(err)
					}
					wire.OfflineCatchup, err = buildOfflineCatchup(initial, bundle.Economy, mode, at)
					if err != nil {
						t.Fatal(err)
					}
					inputs, err = json.Marshal(wire)
					if err != nil {
						t.Fatal(err)
					}
					var rows []calculationStageObservationResearch
					observedBundle := bundle
					observedBundle.Commons = calculationStageObserverResearch{ReplayCommonsPolicy: bundle.Commons, t: t, rows: &rows}
					result, err := ApplyLogged(state, request.CanonicalPayload, observedBundle, inputs)
					if err != nil || result.Outcome != save.IntentApplied || len(rows) != 1 {
						t.Fatalf("real logged probe at%d: %v outcome%s rows%d", offset, err, result.Outcome, len(rows))
					}
					stage := activationRestore(t, bundle, rows[0].State, 19, at)
					cash := mustDecimal(t, stage.Ledger.Snapshot()["company.cash"])
					productionMS, factor, efficiencyTenths := offset, int64(1), int64(10)
					if buffed && (offset < 1000 || offset > 5000) {
						factor = 7
					}
					if mode == ModeOffline || wire.OfflineCatchup != nil {
						efficiencyTenths = 9
						if productionMS > capMS {
							productionMS = capMS
						}
					}
					if wire.OfflineCatchup != nil {
						factor = 1
					}
					nanocash := int64(10000*1_000_000_000) + 115115*productionMS*factor*efficiencyTenths
					// This alternative conversion/rounding path is a predicate
					// reference here, not an exact-rational payout oracle.
					want := decimal.New(float64(nanocash), -9).Quantize(decimal.CanonicalSignificantDigits)
					if !cash.Eq(want) {
						rejectedPayoutPredictions++
						t.Logf("flat payout model rejected at%d: actual%s ideal%s", offset, cash, want)
					}
					if cash.Gte(target) != want.Gte(target) || !bytes.Equal(origin, mustEncodeState(t, initial)) {
						t.Fatalf("probe%d qualification differs from integer oracle or consumes origin: cash%s ideal%s", offset, cash, want)
					}
					return cash.Gte(target)
				}
				starts := calculationEndpointStartsResearch(t, bundle, initial, mode, horizonMS)
				wantStarts := []int64{0}
				if buffed {
					wantStarts = append(wantStarts, 1000)
				}
				wantStarts = append(wantStarts, 5001)
				if mode == ModeOnline {
					wantStarts = append(wantStarts, capMS+1)
				}
				if !reflect.DeepEqual(starts, wantStarts) {
					t.Fatalf("wrong independently declared boundaries: %v want%v", starts, wantStarts)
				}
				wantFound := buffed || mode == ModeOnline
				wantAt := int64(78_182_687)
				if buffed {
					wantAt = 11_168_956
					if mode == ModeOffline {
						wantAt = 12_409_951
					}
				}
				at, found := calculationRegimeSearchResearch(qualifies, starts, horizonMS)
				if found != wantFound || found && (at != wantAt || !qualifies(at) || qualifies(at-1)) {
					t.Fatalf("wrong earliest result: found%t at%d want found%t at%d", found, at, wantFound, wantAt)
				}
				wholeAt, wholeFound := calculationMonotoneSearchResearch(qualifies, 0, horizonMS)
				withoutCap := []int64{}
				for _, boundary := range starts {
					if boundary != capMS+1 {
						withoutCap = append(withoutCap, boundary)
					}
				}
				cutAt, cutFound := calculationRegimeSearchResearch(qualifies, withoutCap, horizonMS)
				if mode == ModeOnline {
					if wholeFound || cutFound {
						t.Fatal("missing-cap negative control did not miss the real earlier window")
					}
					onlineMisses++
				} else {
					if wholeFound != wantFound || cutFound != wantFound || wantFound && (wholeAt != wantAt || cutAt != wantAt) {
						t.Fatal("offline positive/unreachable controls changed")
					}
					offlineControls++
				}
				cells++
				probeCount += queries
				t.Logf("starts%v found%t earliest%d queries%d rejected flat payouts%d (all candidate/negative/neighbor work)", starts, found, at, queries, rejectedPayoutPredictions)
			})
		}
	}
	if cells != 4 || onlineMisses != 2 || offlineControls != 2 || probeCount == 0 {
		t.Fatalf("incomplete search population: cells%d misses%d controls%d probes%d", cells, onlineMisses, offlineControls, probeCount)
	}
}

func TestAxisCalculationWholeRateReferenceResearch(t *testing.T) {
	bundle := axisContentBundle(t)
	start := time.Date(2026, 10, 15, 12, 0, 0, 0, time.UTC)
	for _, row := range []struct {
		mode                  EvaluationMode
		buffed                bool
		offset                int64
		actual, whole, scaled string
	}{
		{ModeOnline, true, 11_141_857, "9.97816407989e4", "9.97816407988e4", "9.97816407988e4"},
		{ModeOffline, true, 12_407_410, "9.99815771355e4", "9.99815771354e4", "9.99815771354e4"},
		{ModeOnline, false, 43_202_501, "5.97325590261e4", "5.97325590261e4", "5.97325590262e4"},
	} {
		t.Run(fmt.Sprintf("%s/%d", row.mode, row.offset), func(t *testing.T) {
			initial := axisTimingState(t, bundle, start)
			initial.RunStartedAt = start.Add(-4 * time.Second)
			initial.ActiveBuffs = []save.ActiveBuff{}
			if row.buffed {
				initial.ActiveBuffs = []save.ActiveBuff{{BuffInstanceID: "01985555-0000-7000-8000-000000000001",
					EffectRowID: "active.production", ActivatedAttendedMS: 0, ExpiresAttendedMS: 5000}}
			}
			origin := mustEncodeState(t, initial)
			state := activationRestore(t, bundle, origin, 19, start)
			at := start.Add(time.Duration(row.offset) * time.Millisecond)
			schedule, err := resolveActivePlaySchedule(state, bundle.Opportunities, bundle.Prestige, "01986666-2000-7000-8000-000000000001", at)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := applyActivePlaySchedule(state, bundle.Opportunities, bundle.Prestige, "01986666-2000-7000-8000-000000000001", at, schedule); err != nil {
				t.Fatal(err)
			}
			active, err := activePlayContributions(state, bundle.Opportunities, schedule.AttendedNowMS)
			if err != nil {
				t.Fatal(err)
			}
			contributions, err := assembleContributions(state, bundle.Economy, active)
			if err != nil {
				t.Fatal(err)
			}
			rates, err := ratesWithProvisionedAndPolicy(bundle.Economy, state.GeneratorCounts, state.GeneratorProvisioned, contributions, nil)
			if err != nil {
				t.Fatal(err)
			}
			efficiency := decimal.One
			if row.mode == ModeOffline {
				efficiency = mustDecimal(t, "9e-1")
			}
			delta, err := AccrueConstant(rates["company.cash"], row.offset, efficiency)
			if err != nil {
				t.Fatal(err)
			}
			whole := activationRestore(t, bundle, origin, 19, start)
			if _, err := whole.Ledger.ApplyAccrual(economy.Transaction{Entries: []economy.Entry{{ResourceID: "company.cash", Delta: delta}}}); err != nil {
				t.Fatal(err)
			}
			if _, err := Evaluate(state, bundle.Economy, at, row.mode, contributions); err != nil {
				t.Fatal(err)
			}
			actual, wholeRate := state.Ledger.Snapshot()["company.cash"], whole.Ledger.Snapshot()["company.cash"]
			if actual != row.actual || wholeRate != row.whole || actual == row.scaled ||
				!reflect.DeepEqual(state.GeneratorProvisioned, initial.GeneratorProvisioned) || !bytes.Equal(origin, mustEncodeState(t, initial)) {
				t.Fatalf("lost recorded rejected-reference pair: actual%s want%s scaled%s", actual, row.actual, row.scaled)
			}
			t.Logf("Evaluate%s whole-rate%s rejected scaled-integer%s", actual, wholeRate, row.scaled)
		})
	}
}

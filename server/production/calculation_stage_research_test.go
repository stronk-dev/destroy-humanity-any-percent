package production

import (
	"bytes"
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"cloud-clicker/server/accrualhook"
	"cloud-clicker/server/commons"
	"cloud-clicker/server/decimal"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/guild"
	"cloud-clicker/server/multiplier"
	"cloud-clicker/server/save"
)

type calculationStageObservationResearch struct {
	State         []byte
	Result        accrualhook.Result
	Contributions []multiplier.Contribution
}

// Forward the actual last hook, recording its completed state without replacing
// its arithmetic or events. Tests compare the entire observed/unobserved command.
type calculationStageObserverResearch struct {
	ReplayCommonsPolicy
	t    *testing.T
	rows *[]calculationStageObservationResearch
}

func (observer calculationStageObserverResearch) ResolvedAccrualHook(weight *int64) accrualhook.Hook {
	return calculationStageHookResearch{inner: observer.ReplayCommonsPolicy.ResolvedAccrualHook(weight), t: observer.t, rows: observer.rows}
}

type calculationStageHookResearch struct {
	inner accrualhook.Hook
	t     *testing.T
	rows  *[]calculationStageObservationResearch
}

func (hook calculationStageHookResearch) AfterAccrual(state *save.State, catalog *economy.Catalog, revision save.Revision,
	result accrualhook.Result, contributions []multiplier.Contribution) ([]save.EventWrite, error) {
	events, err := hook.inner.AfterAccrual(state, catalog, revision, result, contributions)
	if err == nil {
		*hook.rows = append(*hook.rows, calculationStageObservationResearch{State: mustEncodeState(hook.t, state), Result: result,
			Contributions: append([]multiplier.Contribution(nil), contributions...)})
	}
	return events, err
}

func TestAxisCalculationLongAbsenceStagesResearch(t *testing.T) {
	bundle := axisContentBundle(t)
	start := time.Date(2026, 10, 15, 12, 0, 0, 0, time.UTC)
	at := start.Add(48 * time.Hour)
	commonsCatalog, err := commons.LoadCatalog(bundle.Artifacts["commons"])
	if err != nil {
		t.Fatal(err)
	}
	const guildID = "01986666-3000-7000-8000-000000000001"
	for _, mode := range []EvaluationMode{ModeOnline, ModeOffline} {
		t.Run(string(mode), func(t *testing.T) {
			initial := axisTimingState(t, bundle, start)
			initial.RunStartedAt = start.Add(-4 * time.Second)
			initial.FactionID, initial.IncorporatedAt = "bootstrapper", start
			initial.StockUnits, initial.ConsumedStockUnits = 20, 3
			initial.GuildBoundaryGuildID, initial.GuildTitheCarryPPM = guildID, 999_999
			initial.CompactMember, initial.CompactTithePPM = true, commonsCatalog.DefaultTithePPM
			effect, ok := bundle.Opportunities.Effect("active.production")
			if !ok || effect.DurationMS != 5000 || effect.Factor.String() != "7e0" {
				t.Fatal("pinned production effect changed")
			}
			initial.ActiveBuffs = []save.ActiveBuff{{BuffInstanceID: "01985555-0000-7000-8000-000000000001",
				EffectRowID: effect.ID, ActivatedAttendedMS: 0, ExpiresAttendedMS: 5000}}
			if err := bundle.ValidateFoundationState(initial); err != nil {
				t.Fatal(err)
			}
			origin := mustEncodeState(t, initial)
			batch := guild.SettlementBatch{GuildID: guildID, Settlements: []guild.Settlement{{GuildID: guildID,
				BoundarySeq: 1, DebitUnits: 3, CreditUnits: 7}}}
			external := []multiplier.Contribution{{SourceID: guild.StockConsumptionSourceID, Slot: multiplier.SlotFaction,
				Target: "all", Factor: decimal.One}}
			weight := int64(500_000)
			accrual, err := makeReplayAccrual(external, &weight, batch, bundle.Routes.ContextVersion())
			if err != nil {
				t.Fatal(err)
			}
			baseline := activationRestore(t, bundle, origin, 19, start)
			request, inputs := originManualInputs(t, bundle, baseline, at, mode, accrual, "manual.click")
			var wire replayInputsWire
			if err := json.Unmarshal(inputs, &wire); err != nil {
				t.Fatal(err)
			}
			wire.OfflineCatchup, err = buildOfflineCatchup(initial, bundle.Economy, mode, at)
			if err != nil || (wire.OfflineCatchup != nil) != (mode == ModeOnline) {
				t.Fatalf("unexpected actual catchup envelope: %v", err)
			}
			inputs, err = json.Marshal(wire)
			if err != nil {
				t.Fatal(err)
			}
			ordinary, err := ApplyLogged(baseline, request.CanonicalPayload, bundle, inputs)
			if err != nil || ordinary.Outcome != save.IntentApplied {
				t.Fatalf("actual complete logged baseline failed: %v outcome%s", err, ordinary.Outcome)
			}
			var resolved replayAccrualResolved
			if err := json.Unmarshal(wire.Resolved, &resolved); err != nil || resolved.ActivePlay == nil {
				t.Fatalf("missing actual schedule: %v", err)
			}
			var observedRows []calculationStageObservationResearch
			observedBundle := bundle
			observedBundle.Commons = calculationStageObserverResearch{ReplayCommonsPolicy: bundle.Commons, t: t, rows: &observedRows}
			observed := activationRestore(t, bundle, origin, 19, start)
			observedResult, err := ApplyLogged(observed, request.CanonicalPayload, observedBundle, inputs)
			if err != nil || observedResult.Outcome != save.IntentApplied || len(observedRows) != 1 ||
				!bytes.Equal(mustEncodeState(t, observed), mustEncodeState(t, baseline)) || !bytes.Equal(observedResult.Receipt, ordinary.Receipt) ||
				!reflect.DeepEqual(observedResult.Events, ordinary.Events) {
				t.Fatalf("observer changed the actual complete transition or hook population: %v rows%d", err, len(observedRows))
			}
			project := func(omitClearing bool) ([]calculationStageObservationResearch, []EvaluationResult, []save.EventWrite) {
				t.Helper()
				state := activationRestore(t, bundle, origin, 19, start)
				if err := deriveFactionStockResource(state, bundle.Faction); err != nil {
					t.Fatal(err)
				}
				if !omitClearing {
					if err := applyReplayGuildSettlements(state, accrual.GuildSettlementBatch, bundle.Faction.StockCap); err != nil {
						t.Fatal(err)
					}
				}
				var rows []calculationStageObservationResearch
				projectBundle := bundle
				projectBundle.Commons = calculationStageObserverResearch{ReplayCommonsPolicy: bundle.Commons, t: t, rows: &rows}
				hook := closedReplayAccrualHook(projectBundle, &weight)
				revision := save.Revision{Number: wire.Command.Revision, StreamID: wire.Command.CompanyStreamID,
					OwnerID: wire.Command.FounderID, ConstantsHash: bundle.ConstantsHash}
				var stages []EvaluationResult
				var events []save.EventWrite
				evaluate := func(evaluationMode EvaluationMode, contributions []multiplier.Contribution) {
					result, err := Evaluate(state, bundle.Economy, at, evaluationMode, contributions)
					if err != nil {
						t.Fatal(err)
					}
					stages = append(stages, result)
					values, err := runAccrualHook(hook, request.IntentID, state, bundle.Economy, revision, result, contributions)
					if err != nil {
						t.Fatal(err)
					}
					events = append(events, values...)
				}
				if wire.OfflineCatchup != nil {
					base, err := assembleContributions(state, bundle.Economy, external)
					if err != nil {
						t.Fatal(err)
					}
					evaluate(ModeOffline, base)
				}
				scheduleEvents, err := applyActivePlaySchedule(state, bundle.Opportunities, bundle.Prestige, wire.Command.FounderID, at, *resolved.ActivePlay)
				if err != nil {
					t.Fatal(err)
				}
				for i := range scheduleEvents {
					scheduleEvents[i].IntentID = request.IntentID
				}
				events = append(events, scheduleEvents...)
				active, err := activePlayContributions(state, bundle.Opportunities, resolved.ActivePlay.AttendedNowMS)
				if err != nil {
					t.Fatal(err)
				}
				contributions, err := assembleContributions(state, bundle.Economy, append(append([]multiplier.Contribution{}, external...), active...))
				if err != nil {
					t.Fatal(err)
				}
				evaluate(mode, contributions)
				return rows, stages, events
			}
			projected, stages, events := project(false)
			if !reflect.DeepEqual(projected, observedRows) || len(events) == 0 || len(ordinary.Events) < len(events) ||
				!reflect.DeepEqual(events, ordinary.Events[:len(events)]) {
				t.Fatal("private stage reconstruction lost complete actual hook state/result/contributions or ordered prefix events")
			}
			wantStages, buffFactor := 1, int64(7)
			if mode == ModeOnline {
				wantStages, buffFactor = 2, 1
			}
			// Independent integer arithmetic: rate115115/100000 *86400s *9/10,
			// plus initial10000 cash. All microcash intermediates are exact integers.
			wantMicrocash := int64(10000*1_000_000) + int64(115115*86400*9)*buffFactor
			wantCash := decimal.New(float64(wantMicrocash), -6).String()
			stage := projected[0]
			state := activationRestore(t, bundle, stage.State, 19, at)
			if len(stages) != wantStages || stage.Result.ElapsedMS != 172_800_000 || stage.Result.ProductionMS != 86_400_000 ||
				stage.Result.BankedCreditMS != 43_200_000 || state.ComputeCreditMS != 43_200_000 ||
				state.StockUnits != 17 || state.ConsumedStockUnits != 10 || state.GuildConsumedWindow != 7 || state.GuildBoundarySeq != 1 ||
				state.Ledger.Snapshot()["company.cash"] != wantCash {
				t.Fatalf("wrong stage/cap/bank/clearing/cash: stages%d result%+v cash%s", len(stages), stage.Result, state.Ledger.Snapshot()["company.cash"])
			}
			if mode == ModeOnline && !reflect.DeepEqual(stages[1], EvaluationResult{}) {
				t.Fatal("ordinary stage after catchup is not the actual zero-elapsed evaluation")
			}
			missing, _, _ := project(true)
			missingState := activationRestore(t, bundle, missing[0].State, 19, at)
			if bytes.Equal(missing[0].State, stage.State) || missingState.Ledger.Snapshot()["company.cash"] != wantCash ||
				!bytes.Equal(origin, mustEncodeState(t, initial)) {
				t.Fatal("omitted-clearing control did not expose non-cash state loss or origin was consumed")
			}
			t.Logf("%d stages, one full hook call, cash%s, bank43200000; missing clearing preserves cash but loses state", wantStages, wantCash)
		})
	}
}

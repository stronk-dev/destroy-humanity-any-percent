package production

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"cloud-clicker/server/commons"
	"cloud-clicker/server/commonsbinding"
	"cloud-clicker/server/decimal"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/faction"
	"cloud-clicker/server/meters"
	"cloud-clicker/server/minigame"
	prestigecore "cloud-clicker/server/prestige"
	"cloud-clicker/server/routeprojection"
	"cloud-clicker/server/routes"
	"cloud-clicker/server/save"
)

// R8/AC15: one persisted career, never independently seeded later runs.
// Initial level6 is a diagnostic budget, not the OD-2 pacing/default-user proof.
func TestReputationCareerTwoExitsIntegration(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL not set; composed SQL career NOT EXECUTED")
	}
	ctx := context.Background()
	db, err := save.OpenPostgres(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := save.Migrate(ctx, db); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `TRUNCATE accounts,save_streams,catalog_sets,epochs RESTART IDENTITY CASCADE`); err != nil {
		t.Fatal(err)
	}
	bundle := reputationContentBundle(t)
	seedProductionEpoch(t, db, bundle.ConstantsHash, bundle.Artifacts)
	resolver := integrationCatalogs{economy: map[string]*economy.Catalog{bundle.ConstantsHash: bundle.Economy},
		routes: map[string]*routes.Catalog{bundle.ConstantsHash: bundle.Routes}, prestige: map[string]*prestigecore.Policy{bundle.ConstantsHash: bundle.Prestige},
		factions: map[string]*faction.Catalog{bundle.ConstantsHash: bundle.Faction}}
	set := ReplayCatalogSet{bundle.ConstantsHash: bundle}
	store, err := save.NewStore(db, reputationTaxonomyCatalogs{resolver, set}, nil)
	if err != nil {
		t.Fatal(err)
	}
	repository, err := minigame.NewRepository(db)
	if err != nil {
		t.Fatal(err)
	}
	projector, err := routeprojection.New(db, resolver)
	if err != nil {
		t.Fatal(err)
	}
	service, err := NewService(store, resolver, FrozenContributionProvider{DB: db}, nil, nil,
		WithRouteCatalogs(resolver), WithRouteProjector(projector), WithProgressionRuntime(resolver),
		WithCurrentConstantsHash(bundle.ConstantsHash), WithReplayCatalogs(set), WithGuildSettlements(emptyGuildSettlements{}),
		WithMinigameActivity(repository), WithCompactPolicies(commons.CatalogSet{bundle.ConstantsHash: bundle.Commons.(commonsbinding.ReplayPolicy).Catalog}),
		WithCommonsWeightResolver(integrationWeight(1_000_000)))
	if err != nil {
		t.Fatal(err)
	}
	now := save.CanonicalServerTime(time.Now().UTC())
	const accountID = "01986666-b100-4000-8000-000000000001"
	const founderID = "01986666-b100-4000-8000-000000000002"
	if _, err := db.ExecContext(ctx, `INSERT INTO accounts(account_id,recovery_hash) VALUES($1,'test')`, accountID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO account_founders(account_id,founder_id) VALUES($1,$2)`, accountID, founderID); err != nil {
		t.Fatal(err)
	}
	founder := reputationFounderState(t, bundle, 22, now, 6)
	started := now.Add(-time.Duration(bundle.Curriculum.FirstFailure.AttendedMS) * time.Millisecond)
	company := replayFixtureState(t, bundle.Economy, started)
	company.WireVersion, company.Tier, company.MeterBands = 18, 1, nil
	company.GatesCrossed[bundle.Curriculum.FirstFailure.GateID] = true
	metersState, err := meters.NewRunState(bundle.Meters, 0)
	if err != nil {
		t.Fatal(err)
	}
	company.MeterValues, company.MeterDecayRemainders, company.MeterInputRemainders = metersState.Values, metersState.DecayRemainders, metersState.InputRemainders
	company.AchievementsEarnedRun = map[string]bool{}
	if _, err := initializeActivePlayState(company, bundle.Opportunities, founderID); err != nil {
		t.Fatal(err)
	}
	advanceActivePlayFixtureAttendance(t, company, bundle.Opportunities, bundle.Prestige, founderID, now)
	company.ManualTokenRefilledAt = now
	founderRevision, err := store.CreateStream(ctx, save.StreamKey{OwnerKind: save.OwnerFounder, OwnerID: founderID, Scope: economy.ScopeFounder}, bundle.ConstantsHash, founder, save.WriteContext{Cause: "reputation.career.integration"})
	if err != nil {
		t.Fatal(err)
	}
	companyRevision, err := store.CreateStream(ctx, save.StreamKey{OwnerKind: save.OwnerFounder, OwnerID: founderID, Scope: economy.ScopeCompany}, bundle.ConstantsHash, company, save.WriteContext{Cause: "reputation.career.integration"})
	if err != nil {
		t.Fatal(err)
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()
	if _, err := save.PinRunWithGenesisTx(ctx, tx, companyRevision.StreamID, founderID, 1, bundle.ConstantsHash, companyRevision.Version, mustEncodeState(t, company)); err != nil {
		t.Fatal(err)
	}
	frozen, err := FrozenFounderContributions(bundle, founder)
	if err != nil {
		t.Fatal(err)
	}
	if err := save.InsertRunFrozenContributionsTx(ctx, tx, companyRevision.StreamID, 1, frozen); err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}

	load := func() (save.Loaded, save.Loaded) {
		t.Helper()
		owner, err := store.LoadLatest(ctx, founderRevision.StreamID)
		if err != nil {
			t.Fatal(err)
		}
		current, err := store.LoadLatest(ctx, companyRevision.StreamID)
		if err != nil {
			t.Fatal(err)
		}
		return owner, current
	}
	intentIndex := 0
	handle := func(at time.Time, kind string, fields map[string]any) ([]byte, HandleResult) {
		t.Helper()
		owner, current := load()
		intentIndex++
		payload := map[string]any{"intent_id": fmt.Sprintf("01986666-b101-7000-8000-%012d", intentIndex), "kind": kind, "expected_revision": current.Revision.Number}
		if kind == IntentPurchaseReputationNode {
			payload["expected_revision"] = owner.Revision.Number
		}
		if kind == IntentWindDown {
			payload["expected_founder_revision"] = owner.Revision.Number
		}
		for key, value := range fields {
			payload[key] = value
		}
		wire := reputationShapeJSON(t, payload)
		result, err := service.Handle(ctx, companyRevision.StreamID, ModeOnline, at, wire)
		var receipt struct {
			Outcome string `json:"outcome"`
		}
		if err != nil || result.Replay || json.Unmarshal(result.Receipt, &receipt) != nil || receipt.Outcome != "applied" {
			t.Fatalf("career %s: receipt=%s replay=%v err=%v", kind, result.Receipt, result.Replay, err)
		}
		return wire, result
	}
	manual := map[string]any{"action_id": "manual.click", "count": 1, "window_ms": 1}
	handle(now, IntentPerformManualBatch, manual) // Real scripted replacement.
	owner, current := load()
	var branch string
	if err := db.QueryRowContext(ctx, `SELECT replay_inputs->'resolved'->>'selected_branch' FROM run_log WHERE company_stream_id=$1 AND run_seq=1 AND seq=1`, companyRevision.StreamID).Scan(&branch); err != nil || branch != "burnout" {
		t.Fatalf("scripted stored branch=%q err=%v", branch, err)
	}
	if current.State.RunSeq != 2 || len(owner.State.ExitHistory) != 1 || owner.State.ExitHistory[0].ExitType != "scripted_first" ||
		current.State.GeneratorProvisioned["generator.beige_tower"] != 10 {
		t.Fatalf("scripted first did not produce real burnout run2: founder=%+v company=%+v", owner.State, current.State)
	}
	frozenBefore, err := save.LoadRunFrozenContributions(ctx, db, companyRevision.StreamID, 2)
	if err != nil {
		t.Fatal(err)
	}
	unitRows := 0
	for _, row := range frozenBefore {
		if row.SourceID == "reputation.founder_bonus" {
			if row.Factor != "1e0" {
				t.Fatalf("run2 must retain the pre-purchase unit bonus: %s", row.Factor)
			}
			unitRows++
		}
	}
	if unitRows != 1 {
		t.Fatalf("run2 frozen Reputation population=%d", unitRows)
	}
	wire, purchase := handle(now.Add(time.Second), IntentPurchaseReputationNode, map[string]any{"node_id": "reputation.unlock.p05"})
	retry, err := service.Handle(ctx, companyRevision.StreamID, ModeOnline, now.Add(2*time.Second), wire)
	if err != nil || !retry.Replay || !bytes.Equal(retry.Receipt, purchase.Receipt) {
		t.Fatalf("career purchase retry changed: %v", err)
	}
	frozenAfter, err := save.LoadRunFrozenContributions(ctx, db, companyRevision.StreamID, 2)
	if err != nil || canonicalFixtureValue(t, frozenBefore) != canonicalFixtureValue(t, frozenAfter) {
		t.Fatal("mid-run purchase rewrote frozen contributions")
	}
	handle(now.Add(10000*time.Second), IntentPerformManualBatch, manual)
	handle(now.Add(10001*time.Second), IntentCrossGate, map[string]any{"gate_id": "gate.t0_to_t1", "route_id": nil})
	_, elective := handle(now.Add(10002*time.Second), IntentWindDown, map[string]any{"reputation_plan": []string{"reputation.starter.cash_small", "reputation.starter.generated_beige_tower"}})
	owner, current = load()
	if current.State.RunSeq != 3 || len(owner.State.ExitHistory) != 2 || owner.State.ExitHistory[1].ExitType != "collapse" ||
		owner.State.ReputationLevel != 6 || owner.State.ReputationSpent != 6 || owner.State.ReputationUnlockPPM != 50000 ||
		fmt.Sprint(owner.State.ReputationNodesOwned) != "[reputation.starter.cash_small reputation.starter.generated_beige_tower reputation.unlock.p05]" {
		t.Fatalf("elective plan did not persist exact career accounting: founder=%+v run=%d receipt=%s", owner.State, current.State.RunSeq, elective.Receipt)
	}
	if current.State.GeneratorProvisioned["generator.beige_tower"] != 5 || current.State.GeneratorCounts["generator.beige_tower"] != 0 {
		t.Fatal("run3 starter inventory differs")
	}
	cash, _ := current.State.Ledger.Balance("company.cash")
	if cash.String() != "1e3" {
		t.Fatalf("run3 starter cash=%s", cash)
	}
	var pins, geneses, planEvents int
	if err := db.QueryRowContext(ctx, `SELECT (SELECT count(*) FROM run_epochs WHERE company_stream_id=$1),
		(SELECT count(*) FROM run_genesis WHERE company_stream_id=$1),
		(SELECT count(*) FROM events WHERE stream_id=$2 AND kind='reputation_node_purchased.v1' AND payload->>'source'='exit_plan')`, companyRevision.StreamID, founderRevision.StreamID).Scan(&pins, &geneses, &planEvents); err != nil || pins != 3 || geneses != 3 || planEvents != 2 {
		t.Fatalf("career population pins=%d genesis=%d planEvents=%d err=%v", pins, geneses, planEvents, err)
	}
	var wrongPins int
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM run_epochs WHERE company_stream_id=$1 AND constants_hash<>$2`, companyRevision.StreamID, bundle.ConstantsHash).Scan(&wrongPins); err != nil || wrongPins != 0 {
		t.Fatalf("career mismatched pins=%d err=%v", wrongPins, err)
	}
	var factor string
	if err := db.QueryRowContext(ctx, `SELECT factor FROM run_frozen_contributions WHERE company_stream_id=$1 AND run_seq=3 AND source_id='reputation.founder_bonus'`, companyRevision.StreamID).Scan(&factor); err != nil || factor != "1.003e0" {
		t.Fatalf("run3 frozen factor=%q err=%v", factor, err)
	}
	for _, run := range []int64{1, 2} {
		genesis, version, entries := reputationCareerReplay(t, db, companyRevision.StreamID, founderRevision.StreamID, run, &bundle)
		want := map[int64]int{1: 1, 2: 3}[run]
		if len(entries) != want || !entries[len(entries)-1].Terminal {
			t.Fatalf("completed run%d population=%d", run, len(entries))
		}
		if verdict := VerifyReplayRun(genesis, version, bundle, entries, bundle.ConstantsHash, false); verdict != ReplayVerified {
			t.Fatalf("career run%d verdict=%s", run, verdict)
		}
		if run == 2 {
			forged := append([]ReplayLogEntry(nil), entries...)
			object := reputationShapeObject(t, forged[0].ReplayInputs)
			resolved := reputationShapeObject(t, object["resolved"])
			accrual := reputationShapeObject(t, resolved["accrual"])
			var contributions []replayContribution
			if err := json.Unmarshal(accrual["contributions"], &contributions); err != nil {
				t.Fatal(err)
			}
			changed := 0
			for index := range contributions {
				if contributions[index].SourceID == "reputation.founder_bonus" {
					if contributions[index].Factor != "1e0" {
						t.Fatal("factor negative lacks its unit target")
					}
					contributions[index].Factor = "2e0" // One existing frozen factor byte.
					changed++
				}
			}
			if changed != 1 {
				t.Fatal("factor negative lacks exactly one persisted target")
			}
			accrual["contributions"] = reputationShapeJSON(t, contributions)
			resolved["accrual"] = reputationShapeJSON(t, accrual)
			object["resolved"] = reputationShapeJSON(t, resolved)
			forged[0].ReplayInputs = reputationShapeJSON(t, object)
			if verdict := VerifyReplayRun(genesis, version, bundle, forged, bundle.ConstantsHash, false); verdict != ReplayStateDivergence {
				t.Fatalf("corrupt frozen factor verdict=%s", verdict)
			}
		}
	}
	history, err := store.LoadFounderHistory(ctx, founderRevision.StreamID)
	if err != nil || len(history.Entries) != 3 || VerifyFounderHistory(history, set) != ReplayVerified {
		t.Fatalf("complete Founder career entries=%d err=%v", len(history.Entries), err)
	}
	forgedHistory := cloneReputationHistory(t, history)
	head := reputationShapeObject(t, forgedHistory.HeadState)
	head["reputation_spent"] = json.RawMessage("5")
	forgedHistory.HeadState = reputationShapeJSON(t, head)
	if verdict := VerifyFounderHistory(forgedHistory, set); verdict != ReplayStateDivergence {
		t.Fatalf("corrupt Founder career verdict=%s", verdict)
	}
	// Run3 is unfinished, so use its actual logged transition, not a false
	// completed-run verdict. Its non-unit bonus must affect real production.
	handle(now.Add(17002*time.Second), IntentPerformManualBatch, manual)
	_, current = load()
	cash, _ = current.State.Ledger.Balance("company.cash")
	wantCash := decimal.FromFloat64(5).Mul(decimal.FromFloat64(1.003)).Mul(decimal.FromFloat64(7000)).Add(decimal.FromFloat64(1000)).Add(decimal.One)
	if !cash.Eq(wantCash) {
		t.Fatalf("run3 independent frozen starter production cash=%s want=%s", cash, wantCash)
	}
	genesis, version, entries := reputationCareerReplay(t, db, companyRevision.StreamID, founderRevision.StreamID, 3, &bundle)
	if len(entries) != 1 || entries[0].Terminal {
		t.Fatal("in-progress run3 population differs")
	}
	before, err := save.RestoreState(genesis, version, bundle.Economy, economy.ScopeCompany, time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	transition, err := ApplyLogged(before, entries[0].CanonicalPayload, bundle, entries[0].ReplayInputs)
	if err != nil || !canonicalJSONEqual(transition.Receipt, entries[0].ReceiptJSON) || !canonicalJSONEqual(marshalReplayEvents(transition.Events), entries[0].EventsJSON) || !bytes.Equal(mustEncodeState(t, transition.State), mustEncodeState(t, current.State)) {
		t.Fatalf("run3 continuation replay differs: %v", err)
	}
	t.Log("real SQL career: two completed Company runs and three-entry Founder history verified; run3 continuation matches full head; copied factor and Founder-head corruptions return state_divergence")
}

// Company replay owns Company events and the terminal's base Founder/plan
// events, not automatic Fiscal prefixes (the full Founder verifier owns those).
func reputationCareerReplay(t *testing.T, db *sql.DB, company, founder string, run int64, bundle *CatalogBundle) ([]byte, int, []ReplayLogEntry) {
	t.Helper()
	genesis, version, entries := persistedPrestigeReplay(t, db, company, founder, run, bundle)
	for index := range entries {
		wire, err := parseReplayInputs(entries[index].ReplayInputs)
		if err != nil {
			t.Fatal(err)
		}
		rows, err := db.Query(`SELECT kind,schema_version,intent_id,payload FROM events WHERE intent_id=$3
			AND (stream_id=$1 OR ($4 AND stream_id=$2 AND kind<>'fiscal_period_harvested.v1'))
			ORDER BY CASE WHEN stream_id=$1 THEN 1 ELSE 0 END,event_seq,event_id`, company, founder, wire.Command.IntentID, entries[index].Terminal)
		if err != nil {
			t.Fatal(err)
		}
		events := []save.EventWrite{}
		for rows.Next() {
			var event save.EventWrite
			if err := rows.Scan(&event.Kind, &event.SchemaVersion, &event.IntentID, &event.Payload); err != nil {
				rows.Close()
				t.Fatal(err)
			}
			events = append(events, event)
		}
		err = rows.Err()
		rows.Close()
		if err != nil {
			t.Fatal(err)
		}
		entries[index].EventsJSON = marshalReplayEvents(events)
	}
	return genesis, version, entries
}

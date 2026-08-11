package save

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"cloud-clicker/server/decimal"
	"cloud-clicker/server/economy"
)

func TestApplyExitTransactionAtomicFaultsAndReplayIntegration(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}
	ctx := context.Background()
	db, err := OpenPostgres(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := Migrate(ctx, db); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `TRUNCATE epochs,catalog_sets,save_streams CASCADE`); err != nil {
		t.Fatal(err)
	}
	catalog := stateCatalog(t)
	hash := ConstantsHash([]byte(stateCatalogJSON))
	if _, err := db.Exec(`INSERT INTO catalog_sets(constants_hash) VALUES($1)`, hash); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO catalog_artifacts(constants_hash,artifact_name,bytes) VALUES($1,'economy',$2)`, hash, []byte(stateCatalogJSON)); err != nil {
		t.Fatal(err)
	}
	var epochID int64
	if err := db.QueryRow(`INSERT INTO epochs(name,started_at,changelog_ref) VALUES('Phase 0',now(),'changelog/epoch-1.md') RETURNING epoch_id`).Scan(&epochID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO epoch_hashes(epoch_id,constants_hash) VALUES($1,$2)`, epochID, hash); err != nil {
		t.Fatal(err)
	}
	store, err := NewStore(db, catalogMap{hash: catalog}, nil)
	if err != nil {
		t.Fatal(err)
	}
	ownerID := "01985555-0000-7000-8000-000000000001"
	now := time.Date(2026, 7, 29, 14, 0, 0, 0, time.UTC)
	founderState := exitTestState(t, catalog, economy.ScopeFounder, now, 0)
	companyState := exitTestState(t, catalog, economy.ScopeCompany, now, 1)
	founderRevision, err := store.CreateStream(ctx, StreamKey{OwnerKind: OwnerFounder, OwnerID: ownerID, Scope: economy.ScopeFounder}, hash, founderState, WriteContext{Cause: "exit_test"})
	if err != nil {
		t.Fatal(err)
	}
	companyRevision, err := store.CreateStream(ctx, StreamKey{OwnerKind: OwnerFounder, OwnerID: ownerID, Scope: economy.ScopeCompany}, hash, companyState, WriteContext{Cause: "exit_test"})
	if err != nil {
		t.Fatal(err)
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	companyGenesis, err := EncodeState(companyState)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := PinRunWithGenesisTx(ctx, tx, companyRevision.StreamID, ownerID, 1, hash, CurrentVersion, companyGenesis); err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}

	loggedPayload := []byte(`{"expected_founder_revision":1,"expected_revision":1,"kind":"wind_down"}`)
	loggedDigest := sha256.Sum256(loggedPayload)
	loggedHash := "sha256:" + hex.EncodeToString(loggedDigest[:])
	for index, failStep := range []string{"founder_genesis", "run_epoch", "run_genesis", "run_log", "founder_log"} {
		loggedIntentID := fmt.Sprintf("01985555-009%d-7000-8000-00000000000%d", index, index)
		_, err = store.ApplyExitTransactionLogged(ctx, companyRevision.StreamID, 1, 1, loggedIntentID, loggedHash, loggedPayload,
			loggedExitTestMutation(t, ownerID, companyRevision.StreamID, loggedIntentID, hash, now), func(step string) error {
				if step == failStep {
					return errors.New("injected " + failStep + " fault")
				}
				return nil
			})
		if err == nil {
			t.Fatalf("%s fault committed", failStep)
		}
		var runTwoPins, runTwoGenesis int
		if err := db.QueryRowContext(ctx, `SELECT count(*) FROM run_epochs WHERE company_stream_id=$1 AND run_seq=2`, companyRevision.StreamID).Scan(&runTwoPins); err != nil || runTwoPins != 0 {
			t.Fatalf("%s rollback pins=%d err=%v", failStep, runTwoPins, err)
		}
		if err := db.QueryRowContext(ctx, `SELECT count(*) FROM run_genesis WHERE company_stream_id=$1 AND run_seq=2`, companyRevision.StreamID).Scan(&runTwoGenesis); err != nil || runTwoGenesis != 0 {
			t.Fatalf("%s rollback genesis=%d err=%v", failStep, runTwoGenesis, err)
		}
		var founderGenesis, founderLogs int
		if err := db.QueryRowContext(ctx, `SELECT count(*) FROM founder_genesis WHERE founder_stream_id=$1`, founderRevision.StreamID).Scan(&founderGenesis); err != nil || founderGenesis != 0 {
			t.Fatalf("%s rollback Founder genesis=%d err=%v", failStep, founderGenesis, err)
		}
		if err := db.QueryRowContext(ctx, `SELECT count(*) FROM founder_log WHERE founder_stream_id=$1`, founderRevision.StreamID).Scan(&founderLogs); err != nil || founderLogs != 0 {
			t.Fatalf("%s rollback Founder log=%d err=%v", failStep, founderLogs, err)
		}
	}
	var loggedRows, outboxRows int
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM run_log WHERE company_stream_id=$1`, companyRevision.StreamID).Scan(&loggedRows); err != nil || loggedRows != 0 {
		t.Fatalf("run-log rollback rows=%d err=%v", loggedRows, err)
	}
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM transport_player_outbox WHERE stream_id=$1`, companyRevision.StreamID).Scan(&outboxRows); err != nil || outboxRows != 0 {
		t.Fatalf("outbox rollback rows=%d err=%v", outboxRows, err)
	}
	assertLatestRevision(t, ctx, store, founderRevision.StreamID, 1)
	assertLatestRevision(t, ctx, store, companyRevision.StreamID, 1)

	steps := []string{"founder_revision", "founder_events", "company_final_revision", "company_ended_events", "company_started_revision", "company_started_events", "intent_record", "retention"}
	for index, failStep := range steps {
		intentID := []string{
			"01985555-0001-7000-8000-000000000001", "01985555-0002-7000-8000-000000000002",
			"01985555-0003-7000-8000-000000000003", "01985555-0004-7000-8000-000000000004",
			"01985555-0005-7000-8000-000000000005", "01985555-0006-7000-8000-000000000006",
			"01985555-0007-7000-8000-000000000007", "01985555-0008-7000-8000-000000000008",
		}[index]
		_, err := store.ApplyExitTransaction(ctx, companyRevision.StreamID, 1, 1, intentID, "sha256:1111111111111111111111111111111111111111111111111111111111111111", exitTestMutation(ownerID, companyRevision.StreamID, intentID, hash, now), func(step string) error {
			if step == failStep {
				return errors.New("injected exit fault")
			}
			return nil
		})
		if err == nil {
			t.Fatalf("step %s committed", failStep)
		}
		assertLatestRevision(t, ctx, store, founderRevision.StreamID, 1)
		assertLatestRevision(t, ctx, store, companyRevision.StreamID, 1)
	}

	intentID := "01985555-0010-7000-8000-000000000010"
	requestHash := loggedHash
	result, err := store.ApplyExitTransactionLogged(ctx, companyRevision.StreamID, 1, 1, intentID, requestHash, loggedPayload,
		loggedExitTestMutation(t, ownerID, companyRevision.StreamID, intentID, hash, now), nil)
	if err != nil || result.Outcome != IntentApplied || result.Replay || len(result.Events) != 4 {
		t.Fatalf("result=%+v err=%v", result, err)
	}
	wantKinds := []EventKind{EventFounderAdvanced, EventExitOfferDeclined, EventRunEnded, EventRunStarted}
	for index, want := range wantKinds {
		if result.Events[index].Kind != want {
			t.Fatalf("event %d kind=%s want=%s", index, result.Events[index].Kind, want)
		}
	}
	if count, err := store.CountRunEvents(ctx, companyRevision.StreamID, EventExitOfferDeclined, 1); err != nil || count != 1 {
		t.Fatalf("run 1 decline count=%d err=%v", count, err)
	}
	if count, err := store.CountRunEvents(ctx, companyRevision.StreamID, EventExitOfferDeclined, 2); err != nil || count != 0 {
		t.Fatalf("run 2 decline count=%d err=%v", count, err)
	}
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM transport_player_outbox WHERE founder_id=$1`, ownerID).Scan(&outboxRows); err != nil || outboxRows != 5 {
		t.Fatalf("exit outbox rows=%d err=%v", outboxRows, err)
	}
	rows, err := db.QueryContext(ctx, `SELECT message_kind,scope,revision,COALESCE(payload->>'kind','') FROM transport_player_outbox WHERE founder_id=$1 ORDER BY outbox_id`, ownerID)
	if err != nil {
		t.Fatal(err)
	}
	var delivered []string
	for rows.Next() {
		var kind, scope, eventKind string
		var revision int64
		if err := rows.Scan(&kind, &scope, &revision, &eventKind); err != nil {
			rows.Close()
			t.Fatal(err)
		}
		delivered = append(delivered, fmt.Sprintf("%s:%s:%d:%s", kind, scope, revision, eventKind))
	}
	if err := rows.Close(); err != nil {
		t.Fatal(err)
	}
	wantDelivery := []string{"event:founder:2:founder_advanced", "event:company:2:exit_offer_declined", "event:company:2:run_ended", "event:company:3:run_started", "receipt:company:3:"}
	if fmt.Sprint(delivered) != fmt.Sprint(wantDelivery) {
		t.Fatalf("player delivery order=%v want=%v", delivered, wantDelivery)
	}
	assertLatestRevision(t, ctx, store, founderRevision.StreamID, 2)
	assertLatestRevision(t, ctx, store, companyRevision.StreamID, 3)
	genesis, err := store.LoadRunGenesis(ctx, companyRevision.StreamID, 2)
	if err != nil {
		t.Fatal(err)
	}
	var firstRunTwoState []byte
	if err := db.QueryRowContext(ctx, `SELECT state FROM save_revisions WHERE stream_id=$1 AND revision=3`, companyRevision.StreamID).Scan(&firstRunTwoState); err != nil {
		t.Fatal(err)
	}
	if genesis.Version != CurrentVersion || genesis.ConstantsHash != hash || !bytes.Equal(genesis.State, firstRunTwoState) {
		t.Fatalf("run-2 genesis mismatch: %+v bytes_equal=%t", genesis, bytes.Equal(genesis.State, firstRunTwoState))
	}
	if _, err := db.ExecContext(ctx, `UPDATE run_genesis SET version=version+1 WHERE company_stream_id=$1 AND run_seq=2`, companyRevision.StreamID); err == nil {
		t.Fatal("run genesis update succeeded")
	}
	replayed, err := store.ApplyExitTransaction(ctx, companyRevision.StreamID, 1, 1, intentID, requestHash, func(*State, Revision, *State, Revision) (ExitDecision, error) {
		t.Fatal("replay mutation ran")
		return ExitDecision{}, nil
	}, nil)
	if err != nil || !replayed.Replay || len(replayed.Events) != 4 || string(replayed.Receipt) != string(result.Receipt) {
		t.Fatalf("replay=%+v err=%v", replayed, err)
	}
	for index := range result.Events {
		if replayed.Events[index].EventID != result.Events[index].EventID || replayed.Events[index].Kind != result.Events[index].Kind {
			t.Fatalf("replay event %d=%+v want=%+v", index, replayed.Events[index], result.Events[index])
		}
	}
}

func exitTestState(t *testing.T, catalog *economy.Catalog, scope economy.Scope, now time.Time, runSeq int64) *State {
	t.Helper()
	ledger, err := economy.NewLedger(catalog, scope)
	if err != nil {
		t.Fatal(err)
	}
	state := &State{Ledger: ledger, GeneratorCounts: map[string]int64{}, EvaluatedThrough: now, ManualTokenRefilledAt: now,
		GatesCrossed: map[string]bool{}, DoctrinesByTransition: map[string]string{}, LedgerFactKinds: map[string]bool{},
		MeterBands: map[string]int{}, RegionTraits: map[string]bool{}, HintsUnlocked: map[string]bool{}, CompactSamples: []CompactSample{},
		LifetimeValue: decimal.Zero, OfflineSpans: []OfflineSpan{}, NetworkSlots: []NetworkSlot{}, ExitHistory: []ExitRecord{}, RunSeq: runSeq}
	for _, generator := range catalog.GeneratorClassesForScope(scope) {
		state.GeneratorCounts[generator.ID] = 0
	}
	if scope == economy.ScopeCompany {
		state.ManualTokenMilli = catalog.ManualPolicy().BucketCapMilli
		state.RunStartedAt = now.Add(-time.Hour)
		state.Tier = 2
		state.LifetimeValue = decimal.New(8, 12)
	}
	return state
}

func exitTestMutation(ownerID, companyStreamID, intentID, hash string, now time.Time) ExitMutation {
	return func(founder *State, _ Revision, company *State, _ Revision) (ExitDecision, error) {
		founder.ReputationLevel = 2
		founder.RouteKnowledgeBalance = 25
		founder.ExitHistory = append(founder.ExitHistory, ExitRecord{RunID: company.RunSeq, ExitType: "collapse", OccurredAt: now, ReputationDelta: 2})
		newCompany := &State{Ledger: company.Ledger, GeneratorCounts: map[string]int64{"generator.example": 0}, EvaluatedThrough: now,
			ManualTokenMilli: 50_000, ManualTokenRefilledAt: now, GatesCrossed: map[string]bool{}, RunSeq: company.RunSeq + 1,
			DoctrinesByTransition: map[string]string{}, LedgerFactKinds: map[string]bool{}, MeterBands: map[string]int{}, RegionTraits: map[string]bool{},
			HintsUnlocked: map[string]bool{}, CompactSamples: []CompactSample{}, LifetimeValue: decimal.Zero, RunStartedAt: now,
			OfflineSpans: []OfflineSpan{}, NetworkSlots: []NetworkSlot{}, ExitHistory: []ExitRecord{}}
		terms := map[string]any{"reputation_delta": 2, "network_slot_unlocks": []any{}, "route_knowledge": 25, "clout_reach_note": "reach.none"}
		runID := map[string]any{"company_stream_id": companyStreamID, "run_seq": company.RunSeq}
		ended, _ := json.Marshal(map[string]any{"founder_id": ownerID, "run_id": runID, "exit_type": "collapse", "started_at_ms": company.RunStartedAt.UnixMilli(), "ended_at_ms": now.UnixMilli(), "rta_ms": now.Sub(company.RunStartedAt).Milliseconds(), "attended_ms": now.Sub(company.RunStartedAt).Milliseconds(), "terminal_seq": 1, "payout": terms, "tier": company.Tier, "lifetime_value": company.LifetimeValue.String(), "ledger_fact_kinds": []string{}, "executed_routes": []string{}, "gates_crossed": []string{}, "generators_purchased_total": company.GeneratorPurchasedTotal, "assisted": map[string]bool{"commons": false, "advisor": false}})
		started, _ := json.Marshal(map[string]any{"founder_id": ownerID, "run_id": map[string]any{"company_stream_id": companyStreamID, "run_seq": company.RunSeq + 1}, "started_at_ms": now.UnixMilli(), "assisted": map[string]bool{"commons": false, "advisor": false}})
		advanced, _ := json.Marshal(map[string]any{"founder_id": ownerID, "run_id": runID, "exit_type": "collapse", "reputation_delta": 2, "route_knowledge": 25, "occurred_at_ms": now.UnixMilli()})
		declined, _ := json.Marshal(map[string]any{"offer_id": "01985555-0011-7000-8000-000000000011", "run_seq": company.RunSeq})
		return ExitDecision{Outcome: IntentApplied, Receipt: json.RawMessage(`{"outcome":"applied"}`), FinalCompanyState: company, NewCompanyState: newCompany, NewConstantsHash: hash,
			VersionFloors: ExitVersionFloors{CurrentFounder: CurrentVersion, CurrentCompany: CurrentVersion, NextFounder: CurrentVersion, NextCompany: CurrentVersion},
			FounderEvents: []EventWrite{{Kind: EventFounderAdvanced, SchemaVersion: 1, IntentID: intentID, Payload: advanced}},
			CompanyEndedEvents: []EventWrite{
				{Kind: EventExitOfferDeclined, SchemaVersion: 1, IntentID: intentID, Payload: declined},
				{Kind: EventRunEnded, SchemaVersion: 2, IntentID: intentID, Payload: ended},
			},
			CompanyStartedEvents: []EventWrite{{Kind: EventRunStarted, SchemaVersion: 1, IntentID: intentID, Payload: started}}}, nil
	}
}

func loggedExitTestMutation(t *testing.T, ownerID, companyStreamID, intentID, hash string, now time.Time) LoggedExitMutation {
	base := exitTestMutation(ownerID, companyStreamID, intentID, hash, now)
	return func(founder *State, founderRevision Revision, company *State, companyRevision Revision, command ReplayCommand) (ExitDecision, json.RawMessage, error) {
		decision, err := base(founder, founderRevision, company, companyRevision)
		decision.FounderReplayResolved, _ = json.Marshal(map[string]any{"kind": "exit.v1",
			"company_stream_id": command.CompanyStreamID, "run_seq": command.RunSeq,
			"run_log_seq": command.RunLogSeq, "result_constants_hash": hash})
		decision.FounderAuditReceipt, _ = json.Marshal(map[string]any{"intent_id": command.IntentID,
			"outcome": "applied", "founder_revision": founderRevision.Number + 1,
			"result_constants_hash": hash})
		return decision, testReplayInputs(t, command, now), err
	}
}

func assertLatestRevision(t *testing.T, ctx context.Context, store *Store, streamID string, want int64) {
	t.Helper()
	loaded, err := store.LoadLatest(ctx, streamID)
	if err != nil || loaded.Revision.Number != want {
		t.Fatalf("stream=%s revision=%d want=%d err=%v", streamID, loaded.Revision.Number, want, err)
	}
}

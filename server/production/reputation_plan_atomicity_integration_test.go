package production

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"cloud-clicker/server/save"
)

func TestReputationExitPlanWriteFaultsIntegration(t *testing.T) {
	reputationExitPlanIntegration(t, true)
}

// This is the existing persistence boundary plus the live Exit callback,
// not a replacement HTTP/actor/guard test. Normal Handle is the success control.
func reputationPlanWriteFaults(t *testing.T, ctx context.Context, db *sql.DB, store *save.Store, service *Service,
	bundle CatalogBundle, founderStream, companyStream string, now time.Time) {
	t.Helper()
	// Diagnostic revision fixture only, before any Founder log/genesis. No
	// gameplay/history claim is made for these byte-identical old revisions.
	for _, stream := range []string{founderStream, companyStream} {
		for revision := 2; revision <= 8; revision++ {
			result, err := db.ExecContext(ctx, `INSERT INTO save_revisions(stream_id,revision,version,state,constants_hash)
				SELECT stream_id,$2,version,state,constants_hash FROM save_revisions WHERE stream_id=$1 AND revision=1`, stream, revision)
			if err != nil {
				t.Fatal(err)
			}
			count, err := result.RowsAffected()
			if err != nil || count != 1 {
				t.Fatalf("diagnostic revision rows=%d err=%v", count, err)
			}
		}
	}
	owner, err := store.LoadLatest(ctx, founderStream)
	if err != nil {
		t.Fatal(err)
	}
	company, err := store.LoadLatest(ctx, companyStream)
	if err != nil {
		t.Fatal(err)
	}
	if owner.Revision.Number != 8 || company.Revision.Number != 8 {
		t.Fatal("revision/retention fixture did not load")
	}
	ownerBefore, companyBefore := mustEncodeState(t, owner.State), mustEncodeState(t, company.State)
	requestBytes := []byte(`{"intent_id":"01986666-7f09-7000-8000-000000000009","kind":"wind_down","expected_revision":8,"expected_founder_revision":8,"reputation_plan":["reputation.unlock.p05","reputation.starter.cash_small","reputation.starter.generated_beige_tower"]}`)
	request, err := ParseIntent(requestBytes)
	if err != nil || request.InvalidDetail != "" {
		t.Fatalf("plan request invalid: %v/%s", err, request.InvalidDetail)
	}
	before := reputationPlanDBSnapshot(t, ctx, db)
	stages := []string{"founder_genesis", "founder_revision", "founder_events", "company_final_revision", "company_ended_events", "company_started_revision", "run_epoch", "run_frozen_contributions", "run_genesis", "company_started_events", "founder_log", "run_log", "intent_record", "retention"}
	firedCount := 0
	for _, stage := range stages {
		t.Run(stage, func(t *testing.T) {
			injected := errors.New("injected Reputation plan " + stage)
			fired := false
			_, err := store.ApplyExitTransactionLogged(ctx, companyStream, 8, 8, request.IntentID, request.RequestHash, request.CanonicalPayload,
				func(founder *save.State, founderRevision save.Revision, company *save.State, companyRevision save.Revision, command save.ReplayCommand, founderCommand save.FounderReplayCommand) (save.ExitDecision, json.RawMessage, error) {
					return service.applyLoggedExit(ctx, request, founder, founderRevision, company, companyRevision, command, founderCommand, ModeOnline, now.Add(time.Second), []string{})
				}, func(actual string) error {
					if actual == stage {
						fired = true
						firedCount++
						return injected
					}
					return nil
				})
			if !fired || !errors.Is(err, injected) {
				t.Fatalf("fault not reached/exact: fired=%v err=%v", fired, err)
			}
			after := reputationPlanDBSnapshot(t, ctx, db)
			for table, rows := range before {
				if after[table] != rows {
					t.Errorf("%s fault left changed persisted rows in %s", stage, table)
				}
			}
			ownerAfter, err := store.LoadLatest(ctx, founderStream)
			if err != nil || ownerAfter.Revision.Number != 8 || !bytes.Equal(mustEncodeState(t, ownerAfter.State), ownerBefore) {
				t.Fatalf("Founder full rollback differs: %v", err)
			}
			companyAfter, err := store.LoadLatest(ctx, companyStream)
			if err != nil || companyAfter.Revision.Number != 8 || !bytes.Equal(mustEncodeState(t, companyAfter.State), companyBefore) {
				t.Fatalf("Company full rollback differs: %v", err)
			}
		})
	}
	if firedCount != 14 {
		t.Errorf("complete plan fault population: fired=%d want14", firedCount)
	}
	if t.Failed() {
		return
	} // Do not disguise a failed fault population with a positive control.
	result, err := service.Handle(ctx, companyStream, ModeOnline, now.Add(time.Second), requestBytes)
	if err != nil || result.Replay || !bytes.Contains(result.Receipt, []byte(`"outcome":"applied"`)) {
		t.Fatalf("normal Handle plan: receipt=%s replay=%v err=%v", result.Receipt, result.Replay, err)
	}
	owner, err = store.LoadLatest(ctx, founderStream)
	if err != nil {
		t.Fatal(err)
	}
	company, err = store.LoadLatest(ctx, companyStream)
	if err != nil {
		t.Fatal(err)
	}
	if owner.Revision.Number != 9 || company.Revision.Number != 10 || owner.State.ReputationLevel != 6 || owner.State.ReputationSpent != 6 || owner.State.ReputationUnlockPPM != 50000 || len(owner.State.ReputationNodesOwned) != 3 || company.State.RunSeq != 3 || company.State.GeneratorProvisioned["generator.beige_tower"] != 5 {
		t.Fatal("positive plan state/accounting differs")
	}
	var oldFounder, oldCompany, founderRows, companyRows, purchases int
	if err := db.QueryRowContext(ctx, `SELECT
		(SELECT count(*) FROM save_revisions WHERE stream_id=$1 AND revision<=4),
		(SELECT count(*) FROM save_revisions WHERE stream_id=$2 AND revision<=5),
		(SELECT count(*) FROM save_revisions WHERE stream_id=$1),
		(SELECT count(*) FROM save_revisions WHERE stream_id=$2),
		(SELECT count(*) FROM events WHERE stream_id=$1 AND kind='reputation_node_purchased.v1' AND payload->>'source'='exit_plan')`, founderStream, companyStream).Scan(&oldFounder, &oldCompany, &founderRows, &companyRows, &purchases); err != nil || oldFounder != 0 || oldCompany != 0 || founderRows != 5 || companyRows != 5 || purchases != 3 {
		t.Fatalf("positive retention/population: old=%d/%d rows=%d/%d purchases=%d err=%v", oldFounder, oldCompany, founderRows, companyRows, purchases, err)
	}
	var factor string
	if err := db.QueryRowContext(ctx, `SELECT factor FROM run_frozen_contributions WHERE company_stream_id=$1 AND run_seq=3 AND source_id='reputation.founder_bonus'`, companyStream).Scan(&factor); err != nil || factor != "1.003e0" {
		t.Fatalf("positive frozen factor=%s err=%v", factor, err)
	}
	history, err := store.LoadFounderHistory(ctx, founderStream)
	if err != nil || len(history.Entries) != 1 || history.Genesis.Revision != 8 || VerifyFounderHistory(history, ReplayCatalogSet{bundle.ConstantsHash: bundle}) != ReplayVerified {
		t.Fatalf("positive Founder replay: entries=%d err=%v", len(history.Entries), err)
	}
	genesis, version, entries := reputationCareerReplay(t, db, companyStream, founderStream, 2, &bundle)
	if len(entries) != 1 || VerifyReplayRun(genesis, version, bundle, entries, bundle.ConstantsHash, false) != ReplayVerified {
		t.Fatal("positive Company replay differs")
	}
	committed := reputationPlanDBSnapshot(t, ctx, db)
	retry, err := service.Handle(ctx, companyStream, ModeOnline, now.Add(2*time.Second), requestBytes)
	if err != nil || !retry.Replay || !bytes.Equal(retry.Receipt, result.Receipt) {
		t.Fatalf("positive identical retry differs: %v", err)
	}
	retried := reputationPlanDBSnapshot(t, ctx, db)
	for table, rows := range committed {
		if retried[table] != rows {
			t.Fatalf("retry rewrote %s", table)
		}
	}
	t.Log("all14 plan faults fired exact sentinel; complete rows/heads restored; normal Handle commits/replays once and actually prunes old revisions")
}

// Complete row values, not just counts. The static test-owned table list covers
// writes in this Exit path; serial disposable Postgres, never production data.
func reputationPlanDBSnapshot(t *testing.T, ctx context.Context, db *sql.DB) map[string]string {
	t.Helper()
	tables := []string{"save_streams", "save_revisions", "events", "run_epochs", "run_genesis", "run_frozen_contributions", "run_log", "founder_genesis", "founder_log", "intent_records", "transport_player_outbox", "verification_queue"}
	result := map[string]string{}
	for _, table := range tables {
		var rows string
		query := fmt.Sprintf(`SELECT COALESCE(jsonb_agg(to_jsonb(row) ORDER BY to_jsonb(row)::text),'[]'::jsonb)::text FROM %s row`, table)
		if err := db.QueryRowContext(ctx, query).Scan(&rows); err != nil {
			t.Fatal(err)
		}
		result[table] = rows
	}
	return result
}

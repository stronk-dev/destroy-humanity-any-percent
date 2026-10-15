package replayverify

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"testing"
	"time"

	"cloud-clicker/server/kernel"
	"cloud-clicker/server/production"
	"cloud-clicker/server/save"
)

const integrationHash = "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

type recordingSink struct {
	reports []VerificationInvariant
}

func (sink *recordingSink) ReportVerificationInvariant(report VerificationInvariant) {
	sink.reports = append(sink.reports, report)
}

type recordingProjector struct {
	calls int
	err   error
}

func (projector *recordingProjector) ProjectVerifiedRun(context.Context, *sql.Tx, string, int64) error {
	projector.calls++
	return projector.err
}

func TestVerificationClaimTokenLeaseAndTerminalImmutabilityIntegration(t *testing.T) {
	db := replayIntegrationDB(t)
	ctx := context.Background()
	streamID := "11111111-1111-4111-8111-111111111111"
	seedVerificationRun(t, db, streamID, "21111111-1111-4111-8111-111111111111", 1, kernel.Version)
	repository, err := NewRepository(db, &recordingSink{})
	if err != nil {
		t.Fatal(err)
	}

	first, err := repository.claimNext(ctx)
	if err != nil || first.Attempts != 1 || !uuidPattern.MatchString(first.Token) {
		t.Fatalf("first claim=%+v err=%v", first, err)
	}
	if _, err := db.ExecContext(ctx, `UPDATE verification_queue SET claimed_at=clock_timestamp()-interval '6 minutes' WHERE company_stream_id=$1 AND run_seq=1`, streamID); err != nil {
		t.Fatal(err)
	}
	second, err := repository.claimNext(ctx)
	if err != nil || second.Attempts != 2 || second.Token == first.Token {
		t.Fatalf("second claim=%+v err=%v", second, err)
	}

	staleTx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := ownClaim(ctx, staleTx, first); !errors.Is(err, ErrVerificationClaimLost) {
		t.Fatalf("stale owner err=%v", err)
	}
	_ = staleTx.Rollback()

	winnerTx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := ownClaim(ctx, winnerTx, second); err != nil {
		t.Fatal(err)
	}
	if err := markVerified(ctx, winnerTx, second); err != nil {
		t.Fatal(err)
	}
	if err := winnerTx.Commit(); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `UPDATE verification_queue SET last_error='rewritten' WHERE company_stream_id=$1 AND run_seq=1`, streamID); err == nil {
		t.Fatal("terminal verification row was mutable")
	}
}

func TestVerificationTransientCatalogFailureRetriesThenPoisonsIntegration(t *testing.T) {
	db := replayIntegrationDB(t)
	ctx := context.Background()
	streamID := "12222222-2222-4222-8222-222222222222"
	seedVerificationRun(t, db, streamID, "22222222-2222-4222-8222-222222222222", 1, kernel.Version)
	sink := &recordingSink{}
	repository, err := NewRepository(db, sink)
	if err != nil {
		t.Fatal(err)
	}
	injected := errors.New("temporary catalog connection failure")
	repository.fault = func(step string) error {
		if step == "catalog_query" {
			return injected
		}
		return nil
	}
	projector := &recordingProjector{}
	for attempt := 1; attempt <= verificationFailureLimit; attempt++ {
		worked, err := repository.ProcessNext(ctx, projector)
		if !worked || !errors.Is(err, injected) {
			t.Fatalf("attempt %d worked=%v err=%v", attempt, worked, err)
		}
		if attempt < verificationFailureLimit {
			if _, err := db.ExecContext(ctx, `UPDATE verification_queue SET available_at=clock_timestamp()-interval '1 millisecond' WHERE company_stream_id=$1 AND run_seq=1`, streamID); err != nil {
				t.Fatal(err)
			}
		}
	}
	var status string
	var attempts int
	var verdict sql.NullString
	if err := db.QueryRowContext(ctx, `SELECT status,attempts,verdict FROM verification_queue WHERE company_stream_id=$1 AND run_seq=1`, streamID).Scan(&status, &attempts, &verdict); err != nil {
		t.Fatal(err)
	}
	if status != "dead" || attempts != verificationFailureLimit || verdict.Valid || projector.calls != 0 {
		t.Fatalf("status=%s attempts=%d verdict=%v projector=%d", status, attempts, verdict, projector.calls)
	}
	var poisonDetail string
	if err := db.QueryRowContext(ctx, `SELECT detail FROM verification_poison_dead_letters WHERE company_stream_id=$1 AND run_seq=1`, streamID).Scan(&poisonDetail); err != nil || poisonDetail != injected.Error() {
		t.Fatalf("poison detail=%q err=%v", poisonDetail, err)
	}
	if len(sink.reports) != 1 || sink.reports[0].Kind != "verification_poison_dead_letter" || sink.reports[0].Attempts != verificationFailureLimit {
		t.Fatalf("invariants=%+v", sink.reports)
	}
}

func TestVerificationVersionSkewDefersWithoutSpendingAttemptIntegration(t *testing.T) {
	db := replayIntegrationDB(t)
	ctx := context.Background()
	streamID := "13333333-3333-4333-8333-333333333333"
	seedVerificationRun(t, db, streamID, "23333333-3333-4333-8333-333333333333", 1, "9.9.9")
	repository, err := NewRepository(db, &recordingSink{})
	if err != nil {
		t.Fatal(err)
	}
	projector := &recordingProjector{}
	worked, err := repository.ProcessNext(ctx, projector)
	if !worked || err != nil {
		t.Fatalf("worked=%v err=%v", worked, err)
	}
	var status string
	var attempts int
	var available time.Time
	var verdict sql.NullString
	if err := db.QueryRowContext(ctx, `SELECT status,attempts,available_at,verdict FROM verification_queue WHERE company_stream_id=$1 AND run_seq=1`, streamID).Scan(&status, &attempts, &available, &verdict); err != nil {
		t.Fatal(err)
	}
	if status != "pending" || attempts != 0 || verdict.Valid || !available.After(time.Now()) || projector.calls != 0 {
		t.Fatalf("status=%s attempts=%d available=%v verdict=%v projector=%d", status, attempts, available, verdict, projector.calls)
	}
}

func TestVerificationExpiredFinalClaimPoisonsWithoutSixthAttemptIntegration(t *testing.T) {
	db := replayIntegrationDB(t)
	ctx := context.Background()
	streamID := "17777777-7777-4777-8777-777777777777"
	seedVerificationRun(t, db, streamID, "27777777-7777-4777-8777-777777777777", 1, kernel.Version)
	sink := &recordingSink{}
	repository, err := NewRepository(db, sink)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `UPDATE verification_queue SET attempts=$2,last_error='worker crashed' WHERE company_stream_id=$1 AND run_seq=1`, streamID, verificationFailureLimit-1); err != nil {
		t.Fatal(err)
	}
	claimed, err := repository.claimNext(ctx)
	if err != nil || claimed.Attempts != verificationFailureLimit {
		t.Fatalf("claim=%+v err=%v", claimed, err)
	}
	if _, err := db.ExecContext(ctx, `UPDATE verification_queue SET claimed_at=clock_timestamp()-interval '6 minutes' WHERE company_stream_id=$1 AND run_seq=1`, streamID); err != nil {
		t.Fatal(err)
	}
	worked, err := repository.ProcessNext(ctx, &recordingProjector{})
	if !worked || err != nil {
		t.Fatalf("worked=%v err=%v", worked, err)
	}
	var status string
	var attempts int
	if err := db.QueryRowContext(ctx, `SELECT status,attempts FROM verification_queue WHERE company_stream_id=$1 AND run_seq=1`, streamID).Scan(&status, &attempts); err != nil {
		t.Fatal(err)
	}
	if status != "dead" || attempts != verificationFailureLimit || len(sink.reports) != 1 {
		t.Fatalf("status=%s attempts=%d reports=%+v", status, attempts, sink.reports)
	}
}

func TestVerificationProjectionFailureRollsBackThenPoisonsIntegration(t *testing.T) {
	db := replayIntegrationDB(t)
	ctx := context.Background()
	streamID := "1aaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"
	seedVerificationRun(t, db, streamID, "2aaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa", 1, kernel.Version)
	sink := &recordingSink{}
	repository, err := NewRepository(db, sink)
	if err != nil {
		t.Fatal(err)
	}
	repository.verify = func(context.Context, string, int64) (production.ReplayVerdict, error) {
		return production.ReplayVerified, nil
	}
	injected := errors.New("deterministic projection payload failure")
	projector := &recordingProjector{err: injected}
	for attempt := 1; attempt <= verificationFailureLimit; attempt++ {
		worked, err := repository.ProcessNext(ctx, projector)
		if !worked || !errors.Is(err, injected) {
			t.Fatalf("attempt %d worked=%v err=%v", attempt, worked, err)
		}
		if attempt < verificationFailureLimit {
			if _, err := db.ExecContext(ctx, `UPDATE verification_queue SET available_at=clock_timestamp()-interval '1 millisecond' WHERE company_stream_id=$1 AND run_seq=1`, streamID); err != nil {
				t.Fatal(err)
			}
		}
	}
	var status string
	var verdict sql.NullString
	if err := db.QueryRowContext(ctx, `SELECT status,verdict FROM verification_queue WHERE company_stream_id=$1 AND run_seq=1`, streamID).Scan(&status, &verdict); err != nil {
		t.Fatal(err)
	}
	if status != "dead" || verdict.Valid || projector.calls != verificationFailureLimit || len(sink.reports) != 1 {
		t.Fatalf("status=%s verdict=%v calls=%d reports=%+v", status, verdict, projector.calls, sink.reports)
	}
}

func TestVerificationClaimsCompanyRunsInSequenceIntegration(t *testing.T) {
	db := replayIntegrationDB(t)
	ctx := context.Background()
	streamID := "1bbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb"
	ownerID := "2bbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb"
	seedVerificationRun(t, db, streamID, ownerID, 1, kernel.Version)
	seedAdditionalRun(t, db, streamID, 2, kernel.Version)
	if _, err := db.ExecContext(ctx, `UPDATE verification_queue SET available_at=clock_timestamp()+interval '1 hour' WHERE company_stream_id=$1 AND run_seq=1`, streamID); err != nil {
		t.Fatal(err)
	}
	repository, err := NewRepository(db, &recordingSink{})
	if err != nil {
		t.Fatal(err)
	}
	if claimed, err := repository.claimNext(ctx); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("later run bypassed head: claim=%+v err=%v", claimed, err)
	}
	if _, err := db.ExecContext(ctx, `UPDATE verification_queue SET available_at=clock_timestamp()-interval '1 millisecond' WHERE company_stream_id=$1 AND run_seq=1`, streamID); err != nil {
		t.Fatal(err)
	}
	claimed, err := repository.claimNext(ctx)
	if err != nil || claimed.RunSeq != 1 {
		t.Fatalf("head claim=%+v err=%v", claimed, err)
	}
}

func TestVerificationLegacyGapAndVersionDriftAreDatabaseVerdictsIntegration(t *testing.T) {
	db := replayIntegrationDB(t)
	ctx := context.Background()
	legacyStream := "18888888-8888-4888-8888-888888888888"
	seedVerificationRun(t, db, legacyStream, "28888888-8888-4888-8888-888888888888", 1, kernel.Version)
	if _, err := db.ExecContext(ctx, `ALTER TABLE run_log DISABLE TRIGGER run_log_replay_inputs_required`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO run_log(company_stream_id,run_seq,seq,intent_id,canonical_payload,replay_inputs,receipt,server_ts_ms)
		VALUES($1,1,1,'01989999-0002-7000-8000-000000000002','{}',NULL,'{}',1)`, legacyStream); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `ALTER TABLE run_log ENABLE TRIGGER run_log_replay_inputs_required`); err != nil {
		t.Fatal(err)
	}
	repository, err := NewRepository(db, &recordingSink{})
	if err != nil {
		t.Fatal(err)
	}
	if verdict, err := repository.VerifyStoredRun(ctx, legacyStream, 1); err != nil || verdict != "log_gap" {
		t.Fatalf("legacy verdict=%s err=%v", verdict, err)
	}

	driftStream := "19999999-9999-4999-8999-999999999999"
	seedVerificationRun(t, db, driftStream, "29999999-9999-4999-8999-999999999999", 1, kernel.Version)
	if _, err := db.ExecContext(ctx, `INSERT INTO run_version_drift(company_stream_id,run_seq,observed_version) VALUES($1,1,'9.9.9')`, driftStream); err != nil {
		t.Fatal(err)
	}
	if verdict, err := repository.VerifyStoredRun(ctx, driftStream, 1); err != nil || verdict != "engine_mismatch" {
		t.Fatalf("drift verdict=%s err=%v", verdict, err)
	}
}

func TestVerificationDeterministicCatalogEvidenceDeadLettersIntegration(t *testing.T) {
	db := replayIntegrationDB(t)
	ctx := context.Background()
	streamID := "14444444-4444-4444-8444-444444444444"
	seedVerificationRun(t, db, streamID, "24444444-4444-4444-8444-444444444444", 1, kernel.Version)
	repository, err := NewRepository(db, &recordingSink{})
	if err != nil {
		t.Fatal(err)
	}
	worked, err := repository.ProcessNext(ctx, &recordingProjector{})
	if !worked || err != nil {
		t.Fatalf("worked=%v err=%v", worked, err)
	}
	var status, verdict string
	if err := db.QueryRowContext(ctx, `SELECT status,verdict FROM verification_queue WHERE company_stream_id=$1 AND run_seq=1`, streamID).Scan(&status, &verdict); err != nil {
		t.Fatal(err)
	}
	if status != "dead" || verdict != "constants_mismatch" {
		t.Fatalf("status=%s verdict=%s", status, verdict)
	}
}

func TestVerificationEventsAreStreamPairScopedAndIterationFailsClosedIntegration(t *testing.T) {
	db := replayIntegrationDB(t)
	ctx := context.Background()
	victimCompany := "15555555-5555-4555-8555-555555555555"
	victimOwner := "25555555-5555-4555-8555-555555555555"
	attackerCompany := "16666666-6666-4666-8666-666666666666"
	attackerOwner := "26666666-6666-4666-8666-666666666666"
	seedVerificationRun(t, db, victimCompany, victimOwner, 1, kernel.Version)
	seedVerificationRun(t, db, attackerCompany, attackerOwner, 1, kernel.Version)
	victimFounder := "35555555-5555-4555-8555-555555555555"
	if _, err := db.ExecContext(ctx, `INSERT INTO save_streams(id,owner_kind,owner_id,scope) VALUES($1,'founder',$2,'founder')`, victimFounder, victimOwner); err != nil {
		t.Fatal(err)
	}
	intentID := "01989999-0001-7000-8000-000000000001"
	for _, row := range []struct {
		streamID string
		kind     string
	}{
		{victimCompany, "generator_purchased"},
		{victimFounder, "founder_advanced"},
		{victimFounder, "fiscal_period_harvested.v1"},
		{attackerCompany, "generator_purchased"},
	} {
		if _, err := db.ExecContext(ctx, `INSERT INTO events(stream_id,revision,schema_version,kind,intent_id,constants_hash,payload)
			VALUES($1,1,1,$2,$3,$4,'{}')`, row.streamID, row.kind, intentID, integrationHash); err != nil {
			t.Fatal(err)
		}
	}
	repository, err := NewRepository(db, &recordingSink{})
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := repository.events(ctx, victimCompany, intentID)
	if err != nil {
		t.Fatal(err)
	}
	var events []map[string]any
	if err := json.Unmarshal(encoded, &events); err != nil || len(events) != 2 {
		t.Fatalf("events=%s count=%d err=%v", encoded, len(events), err)
	}
	injected := errors.New("event cursor failed during iteration")
	repository.fault = func(step string) error {
		if step == "event_rows" {
			return injected
		}
		return nil
	}
	if _, err := repository.events(ctx, victimCompany, intentID); !errors.Is(err, injected) {
		t.Fatalf("iteration err=%v", err)
	}
}

func TestVerifiedArchiveCompactionIsDeterministicAtomicAndImmutableIntegration(t *testing.T) {
	db := replayIntegrationDB(t)
	ctx := context.Background()
	streamID := "1ccccccc-cccc-4ccc-8ccc-cccccccccccc"
	ownerID := "2ccccccc-cccc-4ccc-8ccc-cccccccccccc"
	founderID := "3ccccccc-cccc-4ccc-8ccc-cccccccccccc"
	genesis := json.RawMessage(`{"v":1,"resources":{"company.cash":"1e6"},"nested":{"ids":["original",null],"exact":9007199254740993}}`)
	seedVerificationRunWithGenesis(t, db, streamID, ownerID, 1, kernel.Version, genesis)
	if _, err := db.ExecContext(ctx, `INSERT INTO save_streams(id,owner_kind,owner_id,scope) VALUES($1,'founder',$2,'founder')`, founderID, ownerID); err != nil {
		t.Fatal(err)
	}
	otherCompany := "1ddddddd-dddd-4ddd-8ddd-dddddddddddd"
	seedVerificationRun(t, db, otherCompany, "2ddddddd-dddd-4ddd-8ddd-dddddddddddd", 1, kernel.Version)
	intentID := "01989999-0003-7000-8000-000000000003"
	secondIntentID := "01989999-0004-7000-8000-000000000004"
	command := json.RawMessage(`{"count":2,"generator":"garage_pc","kind":"buy_generator"}`)
	inputs := json.RawMessage(`{"v":1,"resolved":{"price":"2e3","nested":[true,null,"<source> Ω"]}}`)
	receipt := json.RawMessage(`{"outcome":"applied","snapshot":{"balances":{"company.cash":"9.98e5"},"exact":9007199254740993}}`)
	secondCommand := json.RawMessage(`{"count":3,"generator":"garage_pc","kind":"buy_generator"}`)
	secondInputs := json.RawMessage(`{"v":1,"resolved":{"price":"3e3","nested":[false,{"reason":"unaffordable"}]}}`)
	secondReceipt := json.RawMessage(`{"outcome":"rejected","error":{"code":"unaffordable","detail":{"required":"3e3"}}}`)
	// Reverse insertion order so preservation includes the declared log sequence,
	// an applied revision and a rejected command's SQL NULL revision.
	for _, row := range []struct {
		seq                      int64
		intent                   string
		command, inputs, receipt json.RawMessage
		revision                 any
		stamp                    int64
	}{
		{2, secondIntentID, secondCommand, secondInputs, secondReceipt, nil, 200},
		{1, intentID, command, inputs, receipt, int64(7), 100},
	} {
		if _, err := db.ExecContext(ctx, `INSERT INTO run_log(company_stream_id,run_seq,seq,intent_id,canonical_payload,replay_inputs,receipt,applied_revision,server_ts_ms)
			VALUES($1,1,$2,$3,$4,$5,$6,$7,$8)`, streamID, row.seq, row.intent, []byte(row.command), string(row.inputs), string(row.receipt), row.revision, row.stamp); err != nil {
			t.Fatal(err)
		}
	}
	companyPayload := json.RawMessage(`{"generator":"garage_pc","count":2,"nested":{"owner":"original","exact":9007199254740993}}`)
	founderPayload := json.RawMessage(`{"founder_id":"2ccccccc-cccc-4ccc-8ccc-cccccccccccc","soul":5,"history":["α",null]}`)
	for _, row := range []struct {
		id, stream, intent, kind string
		seq                      int64
		payload                  json.RawMessage
	}{
		{"4ccccccc-cccc-4ccc-8ccc-ccccccccccc1", streamID, intentID, "generator_purchased", 120, companyPayload},
		{"4ccccccc-cccc-4ccc-8ccc-ccccccccccc2", founderID, intentID, "founder_advanced", 100, founderPayload},
		{"4ccccccc-cccc-4ccc-8ccc-ccccccccccc3", streamID, secondIntentID, "generator_purchased", 220, secondReceipt},
		{"4ccccccc-cccc-4ccc-8ccc-ccccccccccc4", founderID, secondIntentID, "founder_advanced", 200, secondInputs},
		// Same intent on another Company's stream and another intent on this
		// pair must not become evidence for either archived command.
		{"4ddddddd-dddd-4ddd-8ddd-ddddddddddd1", otherCompany, intentID, "generator_purchased", 95, json.RawMessage(`{"exclude":"other-owner"}`)},
		{"4ddddddd-dddd-4ddd-8ddd-ddddddddddd2", founderID, "01989999-0005-7000-8000-000000000005", "founder_advanced", 90, json.RawMessage(`{"exclude":"other-intent"}`)},
	} {
		if _, err := db.ExecContext(ctx, `INSERT INTO events(event_id,stream_id,revision,schema_version,kind,intent_id,constants_hash,payload,event_seq)
			VALUES($1,$2,1,1,$3,$4,$5,$6,$7)`, row.id, row.stream, row.kind, row.intent, integrationHash, string(row.payload), row.seq); err != nil {
			t.Fatal(err)
		}
	}
	var epochID int64
	if err := db.QueryRowContext(ctx, `SELECT epoch_id FROM run_epochs WHERE company_stream_id=$1 AND run_seq=1`, streamID).Scan(&epochID); err != nil {
		t.Fatal(err)
	}
	// Explicit JSON keys, not runArchive's production marshal/unmarshal types:
	// dropping a field/tag on both production sides must not erase the oracle.
	want, err := json.Marshal(map[string]any{
		"schema_version": 1, "company_stream_id": streamID, "run_seq": 1,
		"pin":     map[string]any{"epoch_id": epochID, "constants_hash": integrationHash, "engine_version": kernel.Version, "build_vcs_hash": "integration", "seed": "1"},
		"genesis": map[string]any{"version": 1, "state": genesis},
		"entries": []any{
			map[string]any{"seq": 1, "intent_id": intentID, "canonical_payload": command, "replay_inputs": inputs, "receipt": receipt, "applied_revision": 7, "server_ts_ms": 100,
				"events": []any{
					map[string]any{"event_id": "4ccccccc-cccc-4ccc-8ccc-ccccccccccc2", "stream_id": founderID, "event_seq": 100, "kind": "founder_advanced", "schema_version": 1, "intent_id": intentID, "payload": founderPayload},
					map[string]any{"event_id": "4ccccccc-cccc-4ccc-8ccc-ccccccccccc1", "stream_id": streamID, "event_seq": 120, "kind": "generator_purchased", "schema_version": 1, "intent_id": intentID, "payload": companyPayload},
				}},
			map[string]any{"seq": 2, "intent_id": secondIntentID, "canonical_payload": secondCommand, "replay_inputs": secondInputs, "receipt": secondReceipt, "applied_revision": nil, "server_ts_ms": 200,
				"events": []any{
					map[string]any{"event_id": "4ccccccc-cccc-4ccc-8ccc-ccccccccccc4", "stream_id": founderID, "event_seq": 200, "kind": "founder_advanced", "schema_version": 1, "intent_id": secondIntentID, "payload": secondInputs},
					map[string]any{"event_id": "4ccccccc-cccc-4ccc-8ccc-ccccccccccc3", "stream_id": streamID, "event_seq": 220, "kind": "generator_purchased", "schema_version": 1, "intent_id": secondIntentID, "payload": secondReceipt},
				}},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `DELETE FROM run_log WHERE company_stream_id=$1 AND run_seq=1`, streamID); err == nil {
		t.Fatal("active run log deleted without archive")
	}

	firstTx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := archiveVerifiedRun(ctx, firstTx, streamID, 1); err != nil {
		t.Fatal(err)
	}
	var firstBytes []byte
	var firstHash string
	if err := firstTx.QueryRowContext(ctx, `SELECT bytes,sha256 FROM run_log_archive WHERE company_stream_id=$1 AND run_seq=1`, streamID).Scan(&firstBytes, &firstHash); err != nil {
		t.Fatal(err)
	}
	if err := firstTx.Rollback(); err != nil {
		t.Fatal(err)
	}
	var live int
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM run_log WHERE company_stream_id=$1 AND run_seq=1`, streamID).Scan(&live); err != nil || live != 2 {
		t.Fatalf("rollback live=%d err=%v", live, err)
	}
	var rolledBackArchives int
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM run_log_archive WHERE company_stream_id=$1 AND run_seq=1`, streamID).Scan(&rolledBackArchives); err != nil || rolledBackArchives != 0 {
		t.Fatalf("rollback archives=%d err=%v", rolledBackArchives, err)
	}

	secondTx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := archiveVerifiedRun(ctx, secondTx, streamID, 1); err != nil {
		t.Fatal(err)
	}
	if err := secondTx.Commit(); err != nil {
		t.Fatal(err)
	}
	var secondBytes []byte
	var secondHash, encoding string
	var terminalSeq int64
	if err := db.QueryRowContext(ctx, `SELECT bytes,sha256,encoding,terminal_seq FROM run_log_archive WHERE run_id=$1`, streamID+":1").Scan(&secondBytes, &secondHash, &encoding, &terminalSeq); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(firstBytes, secondBytes) || firstHash != secondHash || encoding != "gzip+json.v1" || terminalSeq != 2 {
		t.Fatalf("archive drift first=%s second=%s encoding=%s terminal=%d", firstHash, secondHash, encoding, terminalSeq)
	}
	digest := sha256.Sum256(secondBytes)
	if secondHash != "sha256:"+hex.EncodeToString(digest[:]) {
		t.Fatal("stored archive digest does not cover stored bytes")
	}
	reader, err := gzip.NewReader(bytes.NewReader(secondBytes))
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := io.ReadAll(reader) // Consume the trailer too: verify the gzip checksum.
	if err != nil {
		t.Fatal(err)
	}
	_ = reader.Close()
	if err := compareArchiveJSON(encoded, want); err != nil {
		t.Fatal(err)
	}
	for _, canonical := range []json.RawMessage{command, secondCommand} {
		if !bytes.Contains(encoded, canonical) {
			t.Fatal("canonical command bytes changed")
		}
	}
	for _, corruption := range []struct{ name, before, after string }{
		{"command", `"count":2`, `"count":9`},
		{"replay", `"price":"2e3"`, `"price":"9e3"`},
		{"receipt", `"company.cash":"9.98e5"`, `"company.cash":"0e0"`},
		{"exact integer", `9007199254740993`, `9007199254740992`},
		{"founder event", `4ccccccc-cccc-4ccc-8ccc-ccccccccccc2`, `4ccccccc-cccc-4ccc-8ccc-ccccccccccc9`},
		{"company event", `"owner":"original"`, `"owner":"lost"`},
		{"rejected revision", `"applied_revision":null`, `"applied_revision":1`},
	} {
		t.Run("reject corrupted "+corruption.name, func(t *testing.T) {
			if !bytes.Contains(encoded, []byte(corruption.before)) {
				t.Fatal("corruption target absent")
			}
			corrupted := bytes.Replace(encoded, []byte(corruption.before), []byte(corruption.after), 1)
			if !json.Valid(corrupted) {
				t.Fatal("corruption is not valid JSON")
			}
			if err := compareArchiveJSON(corrupted, want); err == nil {
				t.Fatal("archive content loss escaped comparison")
			}
		})
	}
	for _, field := range []string{"canonical_payload", "replay_inputs", "receipt", "events"} {
		t.Run("reject omitted "+field, func(t *testing.T) {
			var altered map[string]json.RawMessage
			if err := json.Unmarshal(encoded, &altered); err != nil {
				t.Fatal(err)
			}
			var entries []map[string]json.RawMessage
			if err := json.Unmarshal(altered["entries"], &entries); err != nil {
				t.Fatal(err)
			}
			delete(entries[0], field)
			altered["entries"], err = json.Marshal(entries)
			if err != nil {
				t.Fatal(err)
			}
			corrupted, err := json.Marshal(altered)
			if err != nil {
				t.Fatal(err)
			}
			if err := compareArchiveJSON(corrupted, want); err == nil {
				t.Fatal("omitted evidence escaped comparison")
			}
		})
	}
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM run_log WHERE company_stream_id=$1 AND run_seq=1`, streamID).Scan(&live); err != nil || live != 0 {
		t.Fatalf("compacted live=%d err=%v", live, err)
	}
	if _, err := db.ExecContext(ctx, `UPDATE run_log_archive SET sha256=$2 WHERE company_stream_id=$1 AND run_seq=1`, streamID, firstHash); err == nil {
		t.Fatal("archive history was mutable")
	}
	retryTx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := archiveVerifiedRun(ctx, retryTx, streamID, 1); err != nil {
		t.Fatal(err)
	}
	if err := retryTx.Commit(); err != nil {
		t.Fatal(err)
	}
	var retryBytes []byte
	var retryHash string
	if err := db.QueryRowContext(ctx, `SELECT bytes,sha256 FROM run_log_archive WHERE company_stream_id=$1 AND run_seq=1`, streamID).Scan(&retryBytes, &retryHash); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(retryBytes, secondBytes) || retryHash != secondHash {
		t.Fatal("idempotent retry changed the archive")
	}
}

func compareArchiveJSON(actual, expected []byte) error {
	decode := func(encoded []byte) (any, error) {
		decoder := json.NewDecoder(bytes.NewReader(encoded))
		decoder.UseNumber() // Do not hide corruption above 2^53 through float rounding.
		var value any
		if err := decoder.Decode(&value); err != nil {
			return nil, err
		}
		if err := decoder.Decode(new(any)); err != io.EOF {
			return nil, fmt.Errorf("archive trailing data: %v", err)
		}
		return value, nil
	}
	got, err := decode(actual)
	if err != nil {
		return err
	}
	want, err := decode(expected)
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(got, want) {
		return errors.New("archive JSON does not preserve the complete fixture")
	}
	return nil
}

func replayIntegrationDB(t *testing.T) *sql.DB {
	t.Helper()
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}
	ctx := context.Background()
	db, err := save.OpenPostgres(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := save.Migrate(ctx, db); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `TRUNCATE save_streams,catalog_sets,epochs RESTART IDENTITY CASCADE`); err != nil {
		t.Fatal(err)
	}
	return db
}

func seedVerificationRun(t *testing.T, db *sql.DB, streamID, ownerID string, runSeq int64, engineVersion string) {
	t.Helper()
	seedVerificationRunWithGenesis(t, db, streamID, ownerID, runSeq, engineVersion, json.RawMessage(`{}`))
}

func seedVerificationRunWithGenesis(t *testing.T, db *sql.DB, streamID, ownerID string, runSeq int64, engineVersion string, genesis json.RawMessage) {
	t.Helper()
	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `INSERT INTO save_streams(id,owner_kind,owner_id,scope) VALUES($1,'founder',$2,'company')`, streamID, ownerID); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO catalog_sets(constants_hash) VALUES($1) ON CONFLICT DO NOTHING`, integrationHash); err != nil {
		t.Fatal(err)
	}
	var epochID int64
	if err := tx.QueryRowContext(ctx, `SELECT epoch_id FROM epochs WHERE ended_at IS NULL`).Scan(&epochID); errors.Is(err, sql.ErrNoRows) {
		if err := tx.QueryRowContext(ctx, `INSERT INTO epochs(name,started_at,changelog_ref) VALUES('replay test',clock_timestamp(),'changelog/epoch-1.md') RETURNING epoch_id`).Scan(&epochID); err != nil {
			t.Fatal(err)
		}
	} else if err != nil {
		t.Fatal(err)
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO epoch_hashes(epoch_id,constants_hash) VALUES($1,$2) ON CONFLICT DO NOTHING`, epochID, integrationHash); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO run_epochs(company_stream_id,run_seq,epoch_id,constants_hash,engine_version,build_vcs_hash,seed)
		VALUES($1,$2,$3,$4,$5,'integration','1')`, streamID, runSeq, epochID, integrationHash, engineVersion); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO run_genesis(company_stream_id,run_seq,state,version,constants_hash)
		VALUES($1,$2,$4,1,$3)`, streamID, runSeq, integrationHash, []byte(genesis)); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO verification_queue(company_stream_id,run_seq) VALUES($1,$2)`, streamID, runSeq); err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
}

func seedAdditionalRun(t *testing.T, db *sql.DB, streamID string, runSeq int64, engineVersion string) {
	t.Helper()
	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()
	var epochID int64
	if err := tx.QueryRowContext(ctx, `SELECT epoch_id FROM epochs WHERE ended_at IS NULL`).Scan(&epochID); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO run_epochs(company_stream_id,run_seq,epoch_id,constants_hash,engine_version,build_vcs_hash,seed)
		VALUES($1,$2,$3,$4,$5,'integration','2')`, streamID, runSeq, epochID, integrationHash, engineVersion); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO run_genesis(company_stream_id,run_seq,state,version,constants_hash)
		VALUES($1,$2,'{}',1,$3)`, streamID, runSeq, integrationHash); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO verification_queue(company_stream_id,run_seq) VALUES($1,$2)`, streamID, runSeq); err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
}

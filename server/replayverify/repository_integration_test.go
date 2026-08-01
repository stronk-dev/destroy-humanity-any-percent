package replayverify

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"os"
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
	for _, streamID := range []string{victimCompany, victimFounder, attackerCompany} {
		if _, err := db.ExecContext(ctx, `INSERT INTO events(stream_id,revision,schema_version,kind,intent_id,constants_hash,payload)
			VALUES($1,1,1,'generator_purchased',$2,$3,'{}')`, streamID, intentID, integrationHash); err != nil {
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

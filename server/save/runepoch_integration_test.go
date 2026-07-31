package save

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"testing"
	"time"

	"cloud-clicker/server/economy"
	"cloud-clicker/server/kernel"
	"cloud-clicker/server/runidentity"
)

func TestEngineVersionDriftRemainsPlayableAndIsRecordedOnceIntegration(t *testing.T) {
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
	if _, err := db.ExecContext(ctx, `TRUNCATE epochs,catalog_sets,save_streams RESTART IDENTITY CASCADE`); err != nil {
		t.Fatal(err)
	}
	catalog := stateCatalog(t)
	catalogBytes := []byte(stateCatalogJSON)
	hash := ConstantsHash(catalogBytes)
	if _, err := db.ExecContext(ctx, `INSERT INTO catalog_sets(constants_hash) VALUES($1)`, hash); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO catalog_artifacts(constants_hash,artifact_name,bytes) VALUES($1,'economy',$2)`, hash, catalogBytes); err != nil {
		t.Fatal(err)
	}
	var epochID int64
	if err := db.QueryRowContext(ctx, `INSERT INTO epochs(name,started_at,changelog_ref) VALUES('Phase 0',now(),'changelog/epoch-1.md') RETURNING epoch_id`).Scan(&epochID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO epoch_hashes(epoch_id,constants_hash) VALUES($1,$2)`, epochID, hash); err != nil {
		t.Fatal(err)
	}
	store, err := NewStore(db, catalogMap{hash: catalog}, nil)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 7, 30, 12, 0, 0, 0, time.UTC)
	ownerID := "01985555-6100-7000-8000-000000000006"
	ledger, _ := economy.NewLedger(catalog, economy.ScopeCompany)
	state := &State{Ledger: ledger, GeneratorCounts: map[string]int64{"generator.example": 0}, EvaluatedThrough: now,
		ManualTokenMilli: catalog.ManualPolicy().BucketCapMilli, ManualTokenRefilledAt: now, RunSeq: 1, RunStartedAt: now}
	revision, err := store.CreateStream(ctx, StreamKey{OwnerKind: OwnerFounder, OwnerID: ownerID, Scope: economy.ScopeCompany}, hash, state, WriteContext{Cause: "version_drift_test"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `
		INSERT INTO run_epochs(company_stream_id,run_seq,epoch_id,constants_hash,engine_version,build_vcs_hash,seed)
		VALUES($1,1,$2,$3,'0.0.1','review-fixture',$4)`, revision.StreamID, epochID, hash, runidentity.SeedString(ownerID, 1)); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `
		INSERT INTO run_log(company_stream_id,run_seq,seq,intent_id,canonical_payload,receipt,server_ts_ms)
		VALUES($1,1,1,'01985555-6100-7000-8000-000000000007','{}','{}',1)`, revision.StreamID); err == nil {
		t.Fatal("new run-log row without replay_inputs was accepted")
	}
	payload := []byte(`{"kind":"perform_manual_batch"}`)
	digest := sha256.Sum256(payload)
	requestHash := "sha256:" + hex.EncodeToString(digest[:])
	intentID := "01985555-6101-7000-8000-000000000006"
	result, err := store.ApplyIntentLogged(ctx, revision.StreamID, 1, intentID, requestHash, payload, func(state *State, revision Revision, command ReplayCommand) (IntentDecision, json.RawMessage, error) {
		return IntentDecision{Outcome: IntentApplied, Receipt: json.RawMessage(`{"outcome":"applied"}`)}, testReplayInputs(t, command, now), nil
	})
	if err != nil || result.Outcome != IntentApplied {
		t.Fatalf("drifted command=%+v err=%v", result, err)
	}
	var observed string
	var driftRows int
	if err := db.QueryRowContext(ctx, `SELECT count(*),min(observed_version) FROM run_version_drift WHERE company_stream_id=$1 AND run_seq=1`, revision.StreamID).Scan(&driftRows, &observed); err != nil || driftRows != 1 || observed != kernel.Version {
		t.Fatalf("drift rows=%d observed=%s err=%v", driftRows, observed, err)
	}
	secondPayload := []byte(`{"kind":"buy_generator"}`)
	secondDigest := sha256.Sum256(secondPayload)
	if _, err := store.ApplyIntentLogged(ctx, revision.StreamID, 2, "01985555-6102-7000-8000-000000000006", "sha256:"+hex.EncodeToString(secondDigest[:]), secondPayload, func(state *State, revision Revision, command ReplayCommand) (IntentDecision, json.RawMessage, error) {
		return IntentDecision{Outcome: IntentApplied, Receipt: json.RawMessage(`{"outcome":"applied"}`)}, testReplayInputs(t, command, now), nil
	}); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM run_version_drift WHERE company_stream_id=$1 AND run_seq=1`, revision.StreamID).Scan(&driftRows); err != nil || driftRows != 1 {
		t.Fatalf("deduplicated drift rows=%d err=%v", driftRows, err)
	}
	if _, err := db.ExecContext(ctx, `UPDATE run_log SET replay_inputs=NULL WHERE company_stream_id=$1 AND run_seq=1`, revision.StreamID); err == nil {
		t.Fatal("immutable run-log replay inputs were mutable")
	}
	if _, err := db.ExecContext(ctx, `DELETE FROM run_log WHERE company_stream_id=$1 AND run_seq=1`, revision.StreamID); err == nil {
		t.Fatal("immutable run-log row was deletable")
	}
	var replayInputRows int
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM run_log WHERE company_stream_id=$1 AND run_seq=1 AND replay_inputs IS NOT NULL`, revision.StreamID).Scan(&replayInputRows); err != nil || replayInputRows != 2 {
		t.Fatalf("immutable run-log rows=%d err=%v", replayInputRows, err)
	}
}

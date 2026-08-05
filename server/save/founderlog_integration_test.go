package save

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"cloud-clicker/server/economy"
)

func TestApplyFounderLoggedIntegration(t *testing.T) {
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
	if _, err := db.ExecContext(ctx, `TRUNCATE events,intent_records,founder_log,founder_genesis,save_revisions,save_streams CASCADE`); err != nil {
		t.Fatal(err)
	}

	catalog := stateCatalog(t)
	hash := ConstantsHash([]byte(stateCatalogJSON))
	if _, err := db.ExecContext(ctx, `INSERT INTO catalog_sets(constants_hash) VALUES($1) ON CONFLICT DO NOTHING`, hash); err != nil {
		t.Fatal(err)
	}
	store, err := NewStore(db, catalogMap{hash: catalog}, nil)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 8, 5, 10, 0, 0, 0, time.UTC)
	ownerID := "01990000-0000-7000-8000-000000000101"
	founderState := exitTestState(t, catalog, economy.ScopeFounder, now, 0)
	founderRevision, err := store.CreateStream(ctx, StreamKey{OwnerKind: OwnerFounder, OwnerID: ownerID,
		Scope: economy.ScopeFounder}, hash, founderState, WriteContext{Cause: "founder_log_test"})
	if err != nil {
		t.Fatal(err)
	}
	companyState := exitTestState(t, catalog, economy.ScopeCompany, now, 1)
	companyRevision, err := store.CreateStream(ctx, StreamKey{OwnerKind: OwnerFounder, OwnerID: ownerID,
		Scope: economy.ScopeCompany}, hash, companyState, WriteContext{Cause: "founder_log_scope_test"})
	if err != nil {
		t.Fatal(err)
	}

	payload := []byte(`{"kind":"fixture_founder_action"}`)
	digest := sha256.Sum256(payload)
	requestHash := "sha256:" + hex.EncodeToString(digest[:])
	intentID := "01990000-0001-7000-8000-000000000101"
	result, err := store.ApplyFounderLogged(ctx, founderRevision.StreamID, 1, intentID, requestHash, payload,
		func(state *State, _ Revision, command FounderReplayCommand) (IntentDecision, json.RawMessage, error) {
			state.Soul++
			advancedPayload, marshalErr := json.Marshal(map[string]any{"founder_id": ownerID,
				"run_id":    map[string]any{"company_stream_id": companyRevision.StreamID, "run_seq": 1},
				"exit_type": "collapse", "reputation_delta": 0, "route_knowledge": 0,
				"occurred_at_ms": command.ServerTSMS})
			if marshalErr != nil {
				return IntentDecision{}, nil, marshalErr
			}
			return IntentDecision{Outcome: IntentApplied, Receipt: json.RawMessage(`{"outcome":"applied","soul":1}`),
					Events: []EventWrite{{Kind: EventFounderAdvanced, SchemaVersion: 1, IntentID: command.IntentID, Payload: advancedPayload}}},
				testFounderReplayInputs(t, command, "fixture_founder_action"), nil
		})
	if err != nil || result.Outcome != IntentApplied || result.Replay {
		t.Fatalf("applied result=%+v err=%v", result, err)
	}
	genesis, err := store.LoadFounderGenesis(ctx, founderRevision.StreamID)
	if err != nil || genesis.Revision != 1 || genesis.Version != founderRevision.Version || genesis.ConstantsHash != hash {
		t.Fatalf("Founder genesis=%+v err=%v", genesis, err)
	}
	var firstRevision []byte
	if err := db.QueryRowContext(ctx, `SELECT state::text FROM save_revisions WHERE stream_id=$1 AND revision=1`,
		founderRevision.StreamID).Scan(&firstRevision); err != nil || string(genesis.State) != string(firstRevision) {
		t.Fatalf("Founder genesis differs from first command revision err=%v", err)
	}
	if _, err := db.ExecContext(ctx, `UPDATE founder_genesis SET version=version+1 WHERE founder_stream_id=$1`, founderRevision.StreamID); err == nil {
		t.Fatal("Founder genesis was mutable")
	}
	if _, err := db.ExecContext(ctx, `DELETE FROM founder_genesis WHERE founder_stream_id=$1`, founderRevision.StreamID); err == nil {
		t.Fatal("Founder genesis was deletable")
	}
	loaded, err := store.LoadLatest(ctx, founderRevision.StreamID)
	if err != nil || loaded.Revision.Number != 2 || loaded.State.Soul != 1 {
		t.Fatalf("loaded revision=%+v soul=%d err=%v", loaded.Revision, loaded.State.Soul, err)
	}

	replayed, err := store.ApplyFounderLogged(ctx, founderRevision.StreamID, 1, intentID, requestHash, payload,
		func(*State, Revision, FounderReplayCommand) (IntentDecision, json.RawMessage, error) {
			t.Fatal("idempotent replay invoked mutation")
			return IntentDecision{}, nil, nil
		})
	if err != nil || !replayed.Replay || replayed.Outcome != IntentApplied ||
		string(replayed.Receipt) != `{"outcome":"applied","soul":1}` {
		t.Fatalf("replayed result=%+v err=%v", replayed, err)
	}

	rejectedPayload := []byte(`{"kind":"fixture_founder_rejection"}`)
	rejectedDigest := sha256.Sum256(rejectedPayload)
	rejectedID := "01990000-0002-7000-8000-000000000101"
	rejected, err := store.ApplyFounderLogged(ctx, founderRevision.StreamID, 2, rejectedID,
		"sha256:"+hex.EncodeToString(rejectedDigest[:]), rejectedPayload,
		func(_ *State, _ Revision, command FounderReplayCommand) (IntentDecision, json.RawMessage, error) {
			return IntentDecision{Outcome: IntentRejected, Receipt: json.RawMessage(`{"outcome":"rejected","reason":"not_eligible"}`)},
				testFounderReplayInputs(t, command, "fixture_founder_rejection"), nil
		})
	if err != nil || rejected.Outcome != IntentRejected {
		t.Fatalf("rejected result=%+v err=%v", rejected, err)
	}
	loaded, err = store.LoadLatest(ctx, founderRevision.StreamID)
	if err != nil || loaded.Revision.Number != 2 || loaded.State.Soul != 1 {
		t.Fatalf("rejection mutated state revision=%+v soul=%d err=%v", loaded.Revision, loaded.State.Soul, err)
	}

	called := false
	if _, err := store.ApplyFounderLogged(ctx, companyRevision.StreamID, 1,
		"01990000-0003-7000-8000-000000000101", requestHash, payload,
		func(*State, Revision, FounderReplayCommand) (IntentDecision, json.RawMessage, error) {
			called = true
			return IntentDecision{}, nil, nil
		}); !errors.Is(err, ErrInvalidStream) || called {
		t.Fatalf("Company stream reached Founder mutation called=%v err=%v", called, err)
	}
	called = false
	if _, err := store.ApplyIntent(ctx, founderRevision.StreamID, 2,
		"01990000-0004-7000-8000-000000000101", requestHash,
		func(state *State, _ Revision) (IntentDecision, error) {
			called = true
			state.Soul++
			return IntentDecision{Outcome: IntentApplied, Receipt: json.RawMessage(`{"outcome":"applied"}`)}, nil
		}); !errors.Is(err, ErrInvalidStream) || called {
		t.Fatalf("legacy intent bypass called=%v err=%v", called, err)
	}

	rows, err := db.QueryContext(ctx, `SELECT seq,intent_id,applied_revision,constants_hash FROM founder_log
		WHERE founder_stream_id=$1 ORDER BY seq`, founderRevision.StreamID)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	wantIDs := []string{intentID, rejectedID}
	for index, wantID := range wantIDs {
		if !rows.Next() {
			t.Fatalf("missing Founder log row %d", index+1)
		}
		var sequence int64
		var storedID, storedHash string
		var applied *int64
		if err := rows.Scan(&sequence, &storedID, &applied, &storedHash); err != nil ||
			sequence != int64(index+1) || storedID != wantID || storedHash != hash {
			t.Fatalf("Founder log row=%d/%s/%v/%s err=%v", sequence, storedID, applied, storedHash, err)
		}
		if index == 0 && (applied == nil || *applied != 2) || index == 1 && applied != nil {
			t.Fatalf("Founder log applied revision index=%d value=%v", index, applied)
		}
	}
	if rows.Next() || rows.Err() != nil {
		t.Fatalf("unexpected Founder log rows err=%v", rows.Err())
	}
	if _, err := db.ExecContext(ctx, `UPDATE founder_log SET receipt='{}' WHERE founder_stream_id=$1`, founderRevision.StreamID); err == nil {
		t.Fatal("Founder log was mutable")
	}
	if _, err := db.ExecContext(ctx, `DELETE FROM founder_log WHERE founder_stream_id=$1`, founderRevision.StreamID); err == nil {
		t.Fatal("Founder log was deletable")
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO founder_log(founder_stream_id,seq,intent_id,canonical_payload,
		replay_inputs,receipt,constants_hash,server_ts_ms) VALUES($1,1,$2,'{}','{}','{}',$3,1)`,
		companyRevision.StreamID, "01990000-0005-7000-8000-000000000101", hash); err == nil {
		t.Fatal("Company stream accepted a Founder log row")
	}
	genesislessOwner := "01990000-0000-7000-8000-000000000102"
	genesisless, err := store.CreateStream(ctx, StreamKey{OwnerKind: OwnerFounder, OwnerID: genesislessOwner,
		Scope: economy.ScopeFounder}, hash, exitTestState(t, catalog, economy.ScopeFounder, now, 0), WriteContext{Cause: "founder_genesis_guard_test"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO founder_log(founder_stream_id,seq,intent_id,canonical_payload,
		replay_inputs,receipt,constants_hash,server_ts_ms) VALUES($1,1,$2,'{}',$3,'{}',$4,1)`,
		genesisless.StreamID, "01990000-0005-7000-8000-000000000102",
		testFounderReplayInputs(t, FounderReplayCommand{IntentID: "01990000-0005-7000-8000-000000000102",
			FounderStreamID: genesisless.StreamID, FounderID: genesislessOwner, Revision: 1,
			FounderLogSeq: 1, ServerTSMS: 1}, "fixture_genesisless"), hash); err == nil {
		t.Fatal("Founder log committed without immutable genesis")
	}

	var outboxRows int
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM transport_player_outbox
		WHERE stream_id=$1 AND scope='founder' AND revision=2`, founderRevision.StreamID).Scan(&outboxRows); err != nil || outboxRows != 3 {
		t.Fatalf("Founder outbox rows=%d err=%v", outboxRows, err)
	}
	var eventRows, receiptRows int
	if err := db.QueryRowContext(ctx, `SELECT count(*) FILTER (WHERE message_kind='event'),
		count(*) FILTER (WHERE message_kind='receipt') FROM transport_player_outbox WHERE stream_id=$1`,
		founderRevision.StreamID).Scan(&eventRows, &receiptRows); err != nil || eventRows != 1 || receiptRows != 2 {
		t.Fatalf("Founder outbox event=%d receipt=%d err=%v", eventRows, receiptRows, err)
	}

	var concurrentCallbacks atomic.Int64
	type concurrentResult struct {
		result IntentResult
		err    error
	}
	results := make(chan concurrentResult, 2)
	for index, concurrentID := range []string{
		"01990000-0006-7000-8000-000000000101",
		"01990000-0007-7000-8000-000000000101",
	} {
		concurrentPayload := []byte(`{"kind":"fixture_founder_concurrent_` + string(rune('a'+index)) + `"}`)
		concurrentDigest := sha256.Sum256(concurrentPayload)
		go func(id, hash string, body []byte) {
			value, applyErr := store.ApplyFounderLogged(ctx, founderRevision.StreamID, 2, id, hash, body,
				func(state *State, _ Revision, command FounderReplayCommand) (IntentDecision, json.RawMessage, error) {
					concurrentCallbacks.Add(1)
					state.Soul++
					replayInputs, marshalErr := founderReplayInputs(command, "fixture_founder_concurrent")
					if marshalErr != nil {
						return IntentDecision{}, nil, marshalErr
					}
					return IntentDecision{Outcome: IntentApplied, Receipt: json.RawMessage(`{"outcome":"applied"}`)},
						replayInputs, nil
				})
			results <- concurrentResult{result: value, err: applyErr}
		}(concurrentID, "sha256:"+hex.EncodeToString(concurrentDigest[:]), concurrentPayload)
	}
	applied, conflicted := 0, 0
	for range 2 {
		concurrent := <-results
		if concurrent.err == nil && concurrent.result.Outcome == IntentApplied {
			applied++
		} else {
			var conflict *RevisionConflict
			if errors.As(concurrent.err, &conflict) {
				conflicted++
			} else {
				t.Fatalf("concurrent Founder result=%+v err=%v", concurrent.result, concurrent.err)
			}
		}
	}
	if applied != 1 || conflicted != 1 || concurrentCallbacks.Load() != 1 {
		t.Fatalf("concurrent applied=%d conflicted=%d callbacks=%d", applied, conflicted, concurrentCallbacks.Load())
	}

	rollbackPayload := []byte(`{"kind":"fixture_founder_rollback"}`)
	rollbackDigest := sha256.Sum256(rollbackPayload)
	rollbackID := "01990000-0008-7000-8000-000000000101"
	oversizedReceipt := json.RawMessage(`{"outcome":"applied","padding":"` + strings.Repeat("x", MaxPlayerOutboxBytes) + `"}`)
	if _, err := store.ApplyFounderLogged(ctx, founderRevision.StreamID, 3, rollbackID,
		"sha256:"+hex.EncodeToString(rollbackDigest[:]), rollbackPayload,
		func(state *State, _ Revision, command FounderReplayCommand) (IntentDecision, json.RawMessage, error) {
			state.Soul++
			return IntentDecision{Outcome: IntentApplied, Receipt: oversizedReceipt},
				testFounderReplayInputs(t, command, "fixture_founder_rollback"), nil
		}); !errors.Is(err, ErrInvalidStream) {
		t.Fatalf("post-log rollback err=%v", err)
	}
	loaded, err = store.LoadLatest(ctx, founderRevision.StreamID)
	if err != nil || loaded.Revision.Number != 3 || loaded.State.Soul != 2 {
		t.Fatalf("rollback state revision=%d soul=%d err=%v", loaded.Revision.Number, loaded.State.Soul, err)
	}
	var rollbackRows int
	if err := db.QueryRowContext(ctx, `SELECT (SELECT count(*) FROM founder_log WHERE intent_id=$1) +
		(SELECT count(*) FROM intent_records WHERE intent_id=$1) +
		(SELECT count(*) FROM transport_player_outbox WHERE source_id=$1)`, rollbackID).Scan(&rollbackRows); err != nil || rollbackRows != 0 {
		t.Fatalf("post-log rollback rows=%d err=%v", rollbackRows, err)
	}

	archiveTx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer archiveTx.Rollback()
	if _, err := archiveTx.ExecContext(ctx, `UPDATE save_streams SET archived_at=clock_timestamp() WHERE id=$1`, founderRevision.StreamID); err != nil {
		t.Fatal(err)
	}
	insertConn, err := db.Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer insertConn.Close()
	var insertPID int
	if err := insertConn.QueryRowContext(ctx, `SELECT pg_backend_pid()`).Scan(&insertPID); err != nil {
		t.Fatal(err)
	}
	archiveRaceID := "01990000-0009-7000-8000-000000000101"
	insertResult := make(chan error, 1)
	go func() {
		_, insertErr := insertConn.ExecContext(ctx, `INSERT INTO founder_log(founder_stream_id,seq,intent_id,
			canonical_payload,replay_inputs,receipt,constants_hash,server_ts_ms)
			VALUES($1,4,$2,'{}','{}','{}',$3,1)`, founderRevision.StreamID, archiveRaceID, hash)
		insertResult <- insertErr
	}()
	blocked := false
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if err := db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM pg_stat_activity
			WHERE pid=$1 AND wait_event_type='Lock')`, insertPID).Scan(&blocked); err != nil {
			t.Fatal(err)
		}
		if blocked {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if !blocked {
		t.Fatal("Founder log insert did not lock behind archival")
	}
	if err := archiveTx.Commit(); err != nil {
		t.Fatal(err)
	}
	if err := <-insertResult; err == nil {
		t.Fatal("Founder log insert committed after concurrent archival")
	}
	var archiveRaceRows int
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM founder_log WHERE intent_id=$1`, archiveRaceID).Scan(&archiveRaceRows); err != nil || archiveRaceRows != 0 {
		t.Fatalf("archive race rows=%d err=%v", archiveRaceRows, err)
	}
}

func testFounderReplayInputs(t *testing.T, command FounderReplayCommand, kind string) json.RawMessage {
	t.Helper()
	encoded, err := founderReplayInputs(command, kind)
	if err != nil {
		t.Fatal(err)
	}
	return encoded
}

func founderReplayInputs(command FounderReplayCommand, kind string) (json.RawMessage, error) {
	return json.Marshal(map[string]any{"v": FounderReplayInputsVersion, "command": command,
		"evaluated_at_ms": command.ServerTSMS, "resolved": map[string]any{"kind": kind}})
}

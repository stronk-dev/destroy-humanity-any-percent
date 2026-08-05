package save

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
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
	if _, err := db.ExecContext(ctx, `TRUNCATE events,intent_records,founder_log,save_revisions,save_streams CASCADE`); err != nil {
		t.Fatal(err)
	}

	catalog := stateCatalog(t)
	hash := ConstantsHash([]byte(stateCatalogJSON))
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
			return IntentDecision{Outcome: IntentApplied, Receipt: json.RawMessage(`{"outcome":"applied","soul":1}`)},
				testFounderReplayInputs(t, command, "fixture_founder_action"), nil
		})
	if err != nil || result.Outcome != IntentApplied || result.Replay {
		t.Fatalf("applied result=%+v err=%v", result, err)
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
		companyRevision.StreamID, "01990000-0004-7000-8000-000000000101", hash); err == nil {
		t.Fatal("Company stream accepted a Founder log row")
	}
}

func testFounderReplayInputs(t *testing.T, command FounderReplayCommand, kind string) json.RawMessage {
	t.Helper()
	encoded, err := json.Marshal(map[string]any{"v": FounderReplayInputsVersion, "command": command,
		"evaluated_at_ms": command.ServerTSMS, "resolved": map[string]any{"kind": kind}})
	if err != nil {
		t.Fatal(err)
	}
	return encoded
}

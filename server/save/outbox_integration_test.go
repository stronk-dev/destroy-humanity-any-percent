package save

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"cloud-clicker/server/economy"
)

func TestPlayerOutboxOrderingDeadLetterAndSizeIntegration(t *testing.T) {
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
	if _, err := db.ExecContext(ctx, `TRUNCATE save_streams CASCADE`); err != nil {
		t.Fatal(err)
	}
	catalog := stateCatalog(t)
	hash := ConstantsHash([]byte(stateCatalogJSON))
	store, err := NewStore(db, catalogMap{hash: catalog}, nil)
	if err != nil {
		t.Fatal(err)
	}
	ledger, err := economy.RestoreLedger(catalog, economy.ScopeCompany, map[string]string{"company.cash": "1e1"})
	if err != nil {
		t.Fatal(err)
	}
	state := &State{Ledger: ledger, GeneratorCounts: map[string]int64{"generator.example": 0}, EvaluatedThrough: testCursor,
		ManualTokenMilli: 50_000, ManualTokenRefilledAt: testCursor}
	founderA := "11111111-1111-4111-8111-111111111111"
	founderB := "22222222-2222-4222-8222-222222222222"
	streamA, err := store.CreateStream(ctx, StreamKey{OwnerKind: OwnerFounder, OwnerID: founderA, Scope: economy.ScopeCompany}, hash, state, WriteContext{Cause: "outbox.test"})
	if err != nil {
		t.Fatal(err)
	}
	streamB, err := store.CreateStream(ctx, StreamKey{OwnerKind: OwnerFounder, OwnerID: founderB, Scope: economy.ScopeCompany}, hash, state, WriteContext{Cause: "outbox.test"})
	if err != nil {
		t.Fatal(err)
	}
	intents := []string{
		"01985555-1001-7000-8000-000000000001",
		"01985555-1002-7000-8000-000000000002",
		"01985555-1003-7000-8000-000000000003",
	}
	rows := []struct {
		founder, stream, intent string
		revision                int64
	}{{founderA, streamA.StreamID, intents[0], 1}, {founderA, streamA.StreamID, intents[1], 2}, {founderB, streamB.StreamID, intents[2], 1}}
	for _, row := range rows {
		if _, err := db.ExecContext(ctx, `INSERT INTO transport_player_outbox(founder_id,stream_id,message_kind,source_id,scope,revision,constants_hash,payload) VALUES($1,$2,'receipt',$3,'company',$4,$5,$6)`,
			row.founder, row.stream, row.intent, row.revision, hash, json.RawMessage(`{"outcome":"applied"}`)); err != nil {
			t.Fatal(err)
		}
	}

	claimed, err := store.ClaimPlayerOutbox(ctx, 10, 30*time.Second)
	if err != nil || len(claimed) != 2 || claimed[0].SourceID != intents[0] || claimed[1].SourceID != intents[2] {
		t.Fatalf("first claim=%+v err=%v", claimed, err)
	}
	if err := store.MarkPlayerPublished(ctx, claimed[0].ID, claimed[0].ClaimToken); err != nil {
		t.Fatal(err)
	}
	if err := store.DeferPlayerClaim(ctx, claimed[1].ID, claimed[1].ClaimToken, "temporary publisher outage", 100*time.Millisecond); err != nil {
		t.Fatal(err)
	}
	var transientAttempts int
	var transientDetail string
	var deferred bool
	if err := db.QueryRowContext(ctx, `SELECT attempt_count,last_error,claimed_until > clock_timestamp() FROM transport_player_outbox WHERE outbox_id=$1`, claimed[1].ID).Scan(&transientAttempts, &transientDetail, &deferred); err != nil || transientAttempts != 0 || transientDetail != "temporary publisher outage" || !deferred {
		t.Fatalf("transient attempt=%d detail=%q deferred=%v err=%v", transientAttempts, transientDetail, deferred, err)
	}
	if _, err := db.ExecContext(ctx, `UPDATE transport_player_outbox SET claimed_until=clock_timestamp()-interval '1 millisecond' WHERE outbox_id=$1`, claimed[1].ID); err != nil {
		t.Fatal(err)
	}
	for attempt := 1; attempt <= 5; attempt++ {
		claimed, err = store.ClaimPlayerOutbox(ctx, 10, 30*time.Second)
		if err != nil || len(claimed) != 2 || claimed[0].SourceID != intents[1] || claimed[0].AttemptCount != attempt-1 {
			t.Fatalf("attempt %d claim=%+v err=%v", attempt, claimed, err)
		}
		dead, err := store.FailPlayerClaim(ctx, claimed[0].ID, claimed[0].ClaimToken, "deterministic encode failure", 5)
		if err != nil || dead != (attempt == 5) {
			t.Fatalf("attempt %d dead=%v err=%v", attempt, dead, err)
		}
		if err := store.ReleasePlayerClaim(ctx, claimed[1].ID, claimed[1].ClaimToken); err != nil {
			t.Fatal(err)
		}
	}
	claimed, err = store.ClaimPlayerOutbox(ctx, 10, 30*time.Second)
	if err != nil || len(claimed) != 1 || claimed[0].SourceID != intents[2] {
		t.Fatalf("post-dead claim=%+v err=%v", claimed, err)
	}
	if err := store.MarkPlayerPublished(ctx, claimed[0].ID, claimed[0].ClaimToken); err != nil {
		t.Fatal(err)
	}
	if pending, err := store.PendingPlayerCount(ctx); err != nil || pending != 0 {
		t.Fatalf("pending=%d err=%v", pending, err)
	}
	var attempts int
	var detail string
	var deadAt time.Time
	if err := db.QueryRowContext(ctx, `SELECT attempt_count,last_error,dead_lettered_at FROM transport_player_outbox WHERE source_id=$1`, intents[1]).Scan(&attempts, &detail, &deadAt); err != nil || attempts != 5 || detail != "deterministic encode failure" || deadAt.IsZero() {
		t.Fatalf("dead letter attempts=%d detail=%q at=%v err=%v", attempts, detail, deadAt, err)
	}

	oversize := json.RawMessage(`{"padding":"` + strings.Repeat("x", MaxPlayerOutboxBytes) + `"}`)
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := insertReceiptOutbox(ctx, tx, founderA, streamA.StreamID, "01985555-1004-7000-8000-000000000004", economy.ScopeCompany, 3, hash, oversize); !errors.Is(err, ErrInvalidStream) {
		t.Fatalf("application size guard err=%v", err)
	}
	_ = tx.Rollback()
	if _, err := db.ExecContext(ctx, `INSERT INTO transport_player_outbox(founder_id,stream_id,message_kind,source_id,scope,revision,constants_hash,payload) VALUES($1,$2,'receipt',$3,'company',3,$4,$5)`,
		founderA, streamA.StreamID, "01985555-1005-7000-8000-000000000005", hash, oversize); err == nil {
		t.Fatal("database accepted oversized receipt")
	}

	// Projectors append authoritative events directly. An event that exceeds
	// the transport cap must commit and enter the deterministic relay lane; the
	// transport concern may not roll back authoritative history.
	oversizedEvent := json.RawMessage(`{"compensates_event_id":"01985555-1001-7000-8000-000000000001","reason_key":"` + strings.Repeat("x", MaxPlayerOutboxBytes) + `"}`)
	var oversizedEventID string
	tx, err = db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := tx.QueryRowContext(ctx, `
		INSERT INTO events(stream_id,revision,schema_version,kind,intent_id,constants_hash,payload)
		VALUES($1,1,1,'compensation',NULL,$2,$3)
		RETURNING event_id`, streamA.StreamID, hash, oversizedEvent).Scan(&oversizedEventID); err != nil {
		_ = tx.Rollback()
		t.Fatalf("authoritative oversized event insert: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("authoritative oversized event commit: %v", err)
	}
	var cursorEffect string
	var queuedBytes int
	if err := db.QueryRowContext(ctx, `
		SELECT payload->>'cursor_effect',octet_length(payload::text)
		FROM transport_player_outbox
		WHERE message_kind='event' AND source_id=$1`, oversizedEventID).Scan(&cursorEffect, &queuedBytes); err != nil ||
		cursorEffect != "historical" || queuedBytes <= MaxPlayerOutboxBytes {
		t.Fatalf("oversized historical event effect=%q bytes=%d err=%v", cursorEffect, queuedBytes, err)
	}

	spaced := make(map[string]int, 5_500)
	for index := 0; index < 5_500; index++ {
		spaced[fmt.Sprintf("k%04d", index)] = 0
	}
	canonical, err := json.Marshal(spaced)
	if err != nil || len(canonical) >= MaxPlayerOutboxBytes {
		t.Fatalf("jsonb expansion fixture bytes=%d err=%v", len(canonical), err)
	}
	tx, err = db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := insertReceiptOutbox(ctx, tx, founderA, streamA.StreamID, "01985555-1006-7000-8000-000000000006", economy.ScopeCompany, 3, hash, canonical); !errors.Is(err, ErrInvalidStream) {
		t.Fatalf("jsonb text size guard err=%v", err)
	}
	_ = tx.Rollback()

	sharedIntent := "01985555-1007-7000-8000-000000000007"
	for _, row := range []struct{ founder, stream string }{{founderA, streamA.StreamID}, {founderB, streamB.StreamID}} {
		if _, err := db.ExecContext(ctx, `INSERT INTO transport_player_outbox(founder_id,stream_id,message_kind,source_id,scope,revision,constants_hash,payload) VALUES($1,$2,'receipt',$3,'company',3,$4,'{"outcome":"applied"}')`,
			row.founder, row.stream, sharedIntent, hash); err != nil {
			t.Fatalf("same intent id in independent stream rejected: %v", err)
		}
	}
	var sharedRows int
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM transport_player_outbox WHERE message_kind='receipt' AND source_id=$1`, sharedIntent).Scan(&sharedRows); err != nil || sharedRows != 2 {
		t.Fatalf("cross-stream intent identity rows=%d err=%v", sharedRows, err)
	}
}

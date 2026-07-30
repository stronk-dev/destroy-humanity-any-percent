package save

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"cloud-clicker/server/economy"
)

func TestReceiptOutboxOrderingDeadLetterAndSizeIntegration(t *testing.T) {
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
		if _, err := db.ExecContext(ctx, `INSERT INTO transport_receipt_outbox(founder_id,company_stream_id,intent_id,revision,constants_hash,receipt) VALUES($1,$2,$3,$4,$5,$6)`,
			row.founder, row.stream, row.intent, row.revision, hash, json.RawMessage(`{"outcome":"applied"}`)); err != nil {
			t.Fatal(err)
		}
	}

	claimed, err := store.ClaimReceiptOutbox(ctx, 10, 30*time.Second)
	if err != nil || len(claimed) != 2 || claimed[0].IntentID != intents[0] || claimed[1].IntentID != intents[2] {
		t.Fatalf("first claim=%+v err=%v", claimed, err)
	}
	if err := store.MarkReceiptPublished(ctx, claimed[0].ID, claimed[0].ClaimToken); err != nil {
		t.Fatal(err)
	}
	if err := store.ReleaseReceiptClaim(ctx, claimed[1].ID, claimed[1].ClaimToken); err != nil {
		t.Fatal(err)
	}
	for attempt := 1; attempt <= 5; attempt++ {
		claimed, err = store.ClaimReceiptOutbox(ctx, 10, 30*time.Second)
		if err != nil || len(claimed) != 2 || claimed[0].IntentID != intents[1] || claimed[0].AttemptCount != attempt-1 {
			t.Fatalf("attempt %d claim=%+v err=%v", attempt, claimed, err)
		}
		dead, err := store.FailReceiptClaim(ctx, claimed[0].ID, claimed[0].ClaimToken, "deterministic encode failure", 5)
		if err != nil || dead != (attempt == 5) {
			t.Fatalf("attempt %d dead=%v err=%v", attempt, dead, err)
		}
		if err := store.ReleaseReceiptClaim(ctx, claimed[1].ID, claimed[1].ClaimToken); err != nil {
			t.Fatal(err)
		}
	}
	claimed, err = store.ClaimReceiptOutbox(ctx, 10, 30*time.Second)
	if err != nil || len(claimed) != 1 || claimed[0].IntentID != intents[2] {
		t.Fatalf("post-dead claim=%+v err=%v", claimed, err)
	}
	if err := store.MarkReceiptPublished(ctx, claimed[0].ID, claimed[0].ClaimToken); err != nil {
		t.Fatal(err)
	}
	if pending, err := store.PendingReceiptCount(ctx); err != nil || pending != 0 {
		t.Fatalf("pending=%d err=%v", pending, err)
	}
	var attempts int
	var detail string
	var deadAt time.Time
	if err := db.QueryRowContext(ctx, `SELECT attempt_count,last_error,dead_lettered_at FROM transport_receipt_outbox WHERE intent_id=$1`, intents[1]).Scan(&attempts, &detail, &deadAt); err != nil || attempts != 5 || detail != "deterministic encode failure" || deadAt.IsZero() {
		t.Fatalf("dead letter attempts=%d detail=%q at=%v err=%v", attempts, detail, deadAt, err)
	}

	oversize := json.RawMessage(`{"padding":"` + strings.Repeat("x", MaxReceiptOutboxBytes) + `"}`)
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := insertReceiptOutbox(ctx, tx, founderA, streamA.StreamID, "01985555-1004-7000-8000-000000000004", 3, hash, oversize); !errors.Is(err, ErrInvalidStream) {
		t.Fatalf("application size guard err=%v", err)
	}
	_ = tx.Rollback()
	if _, err := db.ExecContext(ctx, `INSERT INTO transport_receipt_outbox(founder_id,company_stream_id,intent_id,revision,constants_hash,receipt) VALUES($1,$2,$3,3,$4,$5)`,
		founderA, streamA.StreamID, "01985555-1005-7000-8000-000000000005", hash, oversize); err == nil {
		t.Fatal("database accepted oversized receipt")
	}
}

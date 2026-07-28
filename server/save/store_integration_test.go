package save

import (
	"context"
	"errors"
	"os"
	"sync"
	"testing"
	"time"

	"cloud-clicker/server/economy"
)

type catalogMap map[string]*economy.Catalog

func (catalogs catalogMap) Resolve(hash string) (*economy.Catalog, bool) {
	catalog, ok := catalogs[hash]
	return catalog, ok
}

func TestStoreIntegrationRevisionLifecycle(t *testing.T) {
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
	if _, err := db.ExecContext(ctx, `TRUNCATE save_revisions, save_streams`); err != nil {
		t.Fatal(err)
	}

	catalogBytes := []byte(stateCatalogJSON)
	catalog := stateCatalog(t)
	hash := ConstantsHash(catalogBytes)
	store, err := NewStore(db, catalogMap{hash: catalog}, nil)
	if err != nil {
		t.Fatal(err)
	}
	ledger, err := economy.RestoreLedger(catalog, economy.ScopeCompany, map[string]string{"company.cash": "1e1"})
	if err != nil {
		t.Fatal(err)
	}
	state := &State{
		Ledger: ledger, GeneratorCounts: map[string]int64{"generator.example": 7}, EvaluatedThrough: testCursor,
	}
	key := StreamKey{OwnerKind: OwnerFounder, OwnerID: "11111111-1111-4111-8111-111111111111", Scope: economy.ScopeCompany}
	revision, err := store.CreateStream(ctx, key, hash, state, WriteContext{Cause: "integration.create"})
	if err != nil {
		t.Fatal(err)
	}
	for expected := int64(1); expected < 7; expected++ {
		revision, err = store.Write(ctx, revision.StreamID, expected, hash, state, WriteContext{Cause: "integration.write"})
		if err != nil {
			t.Fatal(err)
		}
	}
	if revision.Number != 7 {
		t.Fatalf("revision = %d", revision.Number)
	}
	var count, minimum, maximum int64
	if err := db.QueryRowContext(ctx, `SELECT count(*),min(revision),max(revision) FROM save_revisions WHERE stream_id=$1`, revision.StreamID).Scan(&count, &minimum, &maximum); err != nil {
		t.Fatal(err)
	}
	if count != 5 || minimum != 3 || maximum != 7 {
		t.Fatalf("retained count=%d range=%d..%d", count, minimum, maximum)
	}
	if _, err := store.Write(ctx, revision.StreamID, 6, hash, state, WriteContext{Cause: "integration.conflict"}); !errors.Is(err, ErrConflict) {
		t.Fatalf("conflict error = %v", err)
	}
	if err := store.Archive(ctx, revision.StreamID, 7); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Write(ctx, revision.StreamID, 7, hash, state, WriteContext{Cause: "integration.archived"}); !errors.Is(err, ErrArchived) {
		t.Fatalf("archived error = %v", err)
	}
	loaded, err := store.LoadLatest(ctx, revision.StreamID)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.ArchivedAt == nil || loaded.State.Ledger.Snapshot()["company.cash"] != "1e1" ||
		loaded.State.GeneratorCounts["generator.example"] != 7 || !loaded.State.EvaluatedThrough.Equal(testCursor) {
		t.Fatal("archived stream did not remain readable")
	}

	concurrentKey := StreamKey{OwnerKind: OwnerFounder, OwnerID: "22222222-2222-4222-8222-222222222222", Scope: economy.ScopeCompany}
	concurrent, err := store.CreateStream(ctx, concurrentKey, hash, state, WriteContext{Cause: "integration.concurrent-create"})
	if err != nil {
		t.Fatal(err)
	}
	errorsByWriter := make([]error, 2)
	var wait sync.WaitGroup
	for index := range errorsByWriter {
		wait.Add(1)
		go func(index int) {
			defer wait.Done()
			_, errorsByWriter[index] = store.Write(ctx, concurrent.StreamID, 1, hash, state, WriteContext{Cause: "integration.concurrent-write"})
		}(index)
	}
	wait.Wait()
	successes, conflicts := 0, 0
	for _, writeErr := range errorsByWriter {
		if writeErr == nil {
			successes++
		} else if errors.Is(writeErr, ErrConflict) {
			conflicts++
		} else {
			t.Fatalf("unexpected concurrent error: %v", writeErr)
		}
	}
	if successes != 1 || conflicts != 1 {
		t.Fatalf("concurrent results: successes=%d conflicts=%d", successes, conflicts)
	}

	raceKey := StreamKey{OwnerKind: OwnerFounder, OwnerID: "44444444-4444-4444-8444-444444444444", Scope: economy.ScopeCompany}
	raceRevision, err := store.CreateStream(ctx, raceKey, hash, state, WriteContext{Cause: "integration.archive-race-create"})
	if err != nil {
		t.Fatal(err)
	}
	encodedState, err := EncodeState(state)
	if err != nil {
		t.Fatal(err)
	}
	writer, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer writer.Rollback()
	if err := writer.QueryRowContext(ctx,
		`SELECT scope FROM save_streams WHERE id=$1 FOR UPDATE`, raceRevision.StreamID,
	).Scan(new(string)); err != nil {
		t.Fatal(err)
	}
	if _, err := writer.ExecContext(ctx, `
		INSERT INTO save_revisions (stream_id,revision,version,state,constants_hash)
		VALUES ($1,2,$2,$3,$4)`, raceRevision.StreamID, CurrentVersion, encodedState, hash); err != nil {
		t.Fatal(err)
	}
	archiveResult := make(chan error, 1)
	go func() {
		archiveResult <- store.Archive(ctx, raceRevision.StreamID, 1)
	}()
	deadline := time.Now().Add(5 * time.Second)
	for {
		var blocked int
		if err := db.QueryRowContext(ctx, `
			SELECT count(*) FROM pg_stat_activity
			WHERE datname=current_database() AND wait_event_type='Lock'
			  AND query LIKE '%SELECT archived_at FROM save_streams%'`).Scan(&blocked); err != nil {
			t.Fatal(err)
		}
		if blocked > 0 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("archive did not block on writer row lock")
		}
		time.Sleep(10 * time.Millisecond)
	}
	if err := writer.Commit(); err != nil {
		t.Fatal(err)
	}
	if err := <-archiveResult; !errors.Is(err, ErrConflict) {
		t.Fatalf("archive race error = %v", err)
	}
	var archived bool
	if err := db.QueryRowContext(ctx,
		`SELECT archived_at IS NOT NULL FROM save_streams WHERE id=$1`, raceRevision.StreamID,
	).Scan(&archived); err != nil {
		t.Fatal(err)
	}
	if archived {
		t.Fatal("stale archive succeeded after concurrent write advanced the head")
	}

	var legacyStreamID string
	err = db.QueryRowContext(ctx, `
		INSERT INTO save_streams (owner_kind, owner_id, scope)
		VALUES ('founder', '33333333-3333-4333-8333-333333333333', 'company')
		RETURNING id`).Scan(&legacyStreamID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `
		INSERT INTO save_revisions
			(stream_id, revision, version, state, constants_hash, created_at)
		VALUES ($1, 1, 1, '{"balances":{"company.cash":"1e0"}}', $2, $3)`,
		legacyStreamID, hash, testCursor); err != nil {
		t.Fatal(err)
	}
	legacy, err := store.LoadLatest(ctx, legacyStreamID)
	if err != nil {
		t.Fatal(err)
	}
	if legacy.State.GeneratorCounts["generator.example"] != 0 ||
		!legacy.State.EvaluatedThrough.Equal(testCursor.Truncate(time.Microsecond)) {
		t.Fatalf("legacy migration state = %+v", legacy.State)
	}
}

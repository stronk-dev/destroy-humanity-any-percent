package save

import (
	"context"
	"errors"
	"os"
	"sync"
	"testing"

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
	key := StreamKey{OwnerKind: OwnerFounder, OwnerID: "11111111-1111-4111-8111-111111111111", Scope: economy.ScopeCompany}
	revision, err := store.CreateStream(ctx, key, hash, ledger, WriteContext{Cause: "integration.create"})
	if err != nil {
		t.Fatal(err)
	}
	for expected := int64(1); expected < 7; expected++ {
		revision, err = store.Write(ctx, revision.StreamID, expected, hash, ledger, WriteContext{Cause: "integration.write"})
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
	if _, err := store.Write(ctx, revision.StreamID, 6, hash, ledger, WriteContext{Cause: "integration.conflict"}); !errors.Is(err, ErrConflict) {
		t.Fatalf("conflict error = %v", err)
	}
	if err := store.Archive(ctx, revision.StreamID, 7); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Write(ctx, revision.StreamID, 7, hash, ledger, WriteContext{Cause: "integration.archived"}); !errors.Is(err, ErrArchived) {
		t.Fatalf("archived error = %v", err)
	}
	loaded, err := store.LoadLatest(ctx, revision.StreamID)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.ArchivedAt == nil || loaded.Ledger.Snapshot()["company.cash"] != "1e1" {
		t.Fatal("archived stream did not remain readable")
	}

	concurrentKey := StreamKey{OwnerKind: OwnerFounder, OwnerID: "22222222-2222-4222-8222-222222222222", Scope: economy.ScopeCompany}
	concurrent, err := store.CreateStream(ctx, concurrentKey, hash, ledger, WriteContext{Cause: "integration.concurrent-create"})
	if err != nil {
		t.Fatal(err)
	}
	errorsByWriter := make([]error, 2)
	var wait sync.WaitGroup
	for index := range errorsByWriter {
		wait.Add(1)
		go func(index int) {
			defer wait.Done()
			_, errorsByWriter[index] = store.Write(ctx, concurrent.StreamID, 1, hash, ledger, WriteContext{Cause: "integration.concurrent-write"})
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
}

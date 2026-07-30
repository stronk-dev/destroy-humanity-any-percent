package leaderboard

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"cloud-clicker/server/economy"
	"cloud-clicker/server/save"
)

type epochCatalogs map[string]*economy.Catalog

func (catalogs epochCatalogs) Resolve(hash string) (*economy.Catalog, bool) {
	catalog, ok := catalogs[hash]
	return catalog, ok
}

func TestEpochMintHotfixAndRunPinningIntegration(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}
	ctx := context.Background()
	db, err := save.OpenPostgres(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := save.Migrate(ctx, db); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `TRUNCATE verification_projection_events,epochs,catalog_sets,accounts,save_streams RESTART IDENTITY CASCADE`); err != nil {
		t.Fatal(err)
	}
	economyBytes, _ := os.ReadFile("../../balance/catalogs/phase0.json")
	routeBytes, _ := os.ReadFile("../../balance/routes/phase0.json")
	artifacts := []Artifact{{Name: "economy", Bytes: economyBytes}, {Name: "routes", Bytes: routeBytes}}
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "changelog"), 0o755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"epoch-1.md", "epoch-2.md"} {
		if err := os.WriteFile(filepath.Join(root, "changelog", name), []byte("# tested epoch\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	repository, _ := NewRepository(db, root)
	first, err := repository.MintEpoch(ctx, "Phase 0", time.Date(2026, 7, 29, 12, 0, 0, 0, time.UTC), "changelog/epoch-1.md", artifacts)
	if err != nil || first.ID != 1 || len(first.Hashes) != 1 {
		t.Fatalf("first=%+v err=%v", first, err)
	}
	if err := repository.AddHotfix(ctx, first.Hashes[0], artifacts); err == nil {
		t.Fatal("duplicate accepted hash was treated as a new hotfix")
	}
	catalog, err := economy.LoadCatalog(economyBytes)
	if err != nil {
		t.Fatal(err)
	}
	store, err := save.NewStore(db, epochCatalogs{first.Hashes[0]: catalog}, nil)
	if err != nil {
		t.Fatal(err)
	}
	ownerID := "01985555-4000-7000-8000-000000000004"
	ledger, _ := economy.NewLedger(catalog, economy.ScopeCompany)
	now := time.Date(2026, 7, 29, 13, 0, 0, 0, time.UTC)
	state := &save.State{Ledger: ledger, GeneratorCounts: map[string]int64{"generator.beige_tower": 0}, EvaluatedThrough: now, ManualTokenMilli: catalog.ManualPolicy().BucketCapMilli, ManualTokenRefilledAt: now, RunSeq: 1, RunStartedAt: now}
	revision, err := store.CreateStream(ctx, save.StreamKey{OwnerKind: save.OwnerFounder, OwnerID: ownerID, Scope: economy.ScopeCompany}, first.Hashes[0], state, save.WriteContext{Cause: "epoch.test"})
	if err != nil {
		t.Fatal(err)
	}
	if epochID, err := store.PinRunToCurrentEpoch(ctx, revision.StreamID, ownerID, 1, first.Hashes[0]); err != nil || epochID != 1 {
		t.Fatalf("pin epoch=%d err=%v", epochID, err)
	}
	second, err := repository.MintEpoch(ctx, "Phase 0.1", now.Add(time.Hour), "changelog/epoch-2.md", artifacts)
	if err != nil || second.ID != 2 {
		t.Fatalf("second=%+v err=%v", second, err)
	}
	if epochID, err := store.PinRunToCurrentEpoch(ctx, revision.StreamID, ownerID, 2, first.Hashes[0]); err != nil || epochID != 2 {
		t.Fatalf("second pin epoch=%d err=%v", epochID, err)
	}
	var runOne, runTwo int64
	if err := db.QueryRowContext(ctx, `SELECT epoch_id FROM run_epochs WHERE company_stream_id=$1 AND run_seq=1`, revision.StreamID).Scan(&runOne); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT epoch_id FROM run_epochs WHERE company_stream_id=$1 AND run_seq=2`, revision.StreamID).Scan(&runTwo); err != nil || runOne != 1 || runTwo != 2 {
		t.Fatalf("pins run1=%d run2=%d err=%v", runOne, runTwo, err)
	}
	if _, err := db.ExecContext(ctx, `UPDATE catalog_artifacts SET bytes=$3 WHERE constants_hash=$1 AND artifact_name=$2`, first.Hashes[0], "economy", []byte(`{}`)); err == nil {
		t.Fatal("catalog artifact update bypassed immutability trigger")
	}
	accountID := "01985555-4100-7000-8000-000000000004"
	if _, err := db.ExecContext(ctx, `INSERT INTO accounts(account_id,recovery_hash) VALUES($1,'test')`, accountID); err != nil {
		t.Fatal(err)
	}
	founders := []string{
		"01985555-4101-7000-8000-000000000001",
		"01985555-4102-7000-8000-000000000002",
		"01985555-4103-7000-8000-000000000003",
		"01985555-4104-7000-8000-000000000004",
	}
	for index, founderID := range founders {
		if _, err := db.ExecContext(ctx, `INSERT INTO account_founders(account_id,founder_id,imported,archived_at) VALUES($1,$2,$3,CASE WHEN $4 THEN now() ELSE NULL END)`, accountID, founderID, index == 3, index != 0); err != nil {
			t.Fatal(err)
		}
	}
	variables := Variables{}
	keys := []int64{1_000, 1_000, 2_000}
	for index, key := range keys {
		worldFirst, err := repository.ProjectVerifiedRun(ctx, VerifiedRun{EventID: []string{"01985555-4201-7000-8000-000000000001", "01985555-4202-7000-8000-000000000002", "01985555-4203-7000-8000-000000000003"}[index], RunID: founders[index] + ":1", FounderID: founders[index], CategoryID: "category.any", Variables: variables, EpochID: 1, KeyMS: &key, VerifiedAt: now.Add(time.Duration(index) * time.Second)})
		if err != nil || worldFirst != (index == 0) {
			t.Fatalf("project %d worldFirst=%v err=%v", index, worldFirst, err)
		}
	}
	importedKey := int64(500)
	if _, err := repository.ProjectVerifiedRun(ctx, VerifiedRun{EventID: "01985555-4204-7000-8000-000000000004", RunID: founders[3] + ":1", FounderID: founders[3], CategoryID: "category.any", Variables: variables, EpochID: 1, KeyMS: &importedKey, VerifiedAt: now}); !errors.Is(err, ErrInvalidEpoch) {
		t.Fatalf("imported projection err=%v", err)
	}
	board, err := repository.TimeBoard(ctx, "category.any", variables, 1, 0, 10, nil)
	if err != nil || len(board) != 3 || board[0].Rank != 1 || board[1].Rank != 1 || board[2].Rank != 3 || !board[0].WorldFirst {
		t.Fatalf("board=%+v err=%v", board, err)
	}
	page, err := repository.TimeBoard(ctx, "category.any", variables, 1, 0, 2, &Cursor{Key: board[0].Key, RunID: board[0].RunID})
	if err != nil || len(page) != 2 || page[0].RunID != board[1].RunID || page[1].Rank != 3 {
		t.Fatalf("page=%+v err=%v", page, err)
	}
	assertEpochRows(t, db, 2, 2)
}

func assertEpochRows(t *testing.T, db *sql.DB, epochs, pins int) {
	t.Helper()
	var gotEpochs, gotPins int
	if err := db.QueryRow(`SELECT count(*) FROM epochs`).Scan(&gotEpochs); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT count(*) FROM run_epochs`).Scan(&gotPins); err != nil || gotEpochs != epochs || gotPins != pins {
		t.Fatalf("epochs=%d pins=%d err=%v", gotEpochs, gotPins, err)
	}
}

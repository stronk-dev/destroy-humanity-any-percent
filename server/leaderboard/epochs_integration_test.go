package leaderboard

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"cloud-clicker/server/economy"
	"cloud-clicker/server/epochseed"
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
	economyBytes, _ := os.ReadFile("../../balance/testdata/epoch5/economy.json")
	routeBytes, _ := os.ReadFile("../../balance/testdata/epoch5/routes.json")
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
	if _, err := repository.MintEpoch(ctx, "bad ref", time.Date(2026, 7, 29, 11, 0, 0, 0, time.UTC), "changelog/epoch-2.md", artifacts); !errors.Is(err, ErrInvalidEpoch) {
		t.Fatalf("bad first changelog err=%v", err)
	}
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
	if _, err := db.ExecContext(ctx, `UPDATE epochs SET ended_at=ended_at + interval '1 millisecond' WHERE epoch_id=1`); err == nil {
		t.Fatal("closed epoch timestamp update bypassed history guard")
	}
	if _, err := db.ExecContext(ctx, `UPDATE epochs SET ended_at=NULL WHERE epoch_id=1`); err == nil {
		t.Fatal("closed epoch reopen bypassed history guard")
	}
	if _, err := db.ExecContext(ctx, `UPDATE epochs SET name='rewritten' WHERE epoch_id=2`); err == nil {
		t.Fatal("current epoch metadata update bypassed history guard")
	}
	if _, err := db.ExecContext(ctx, `DELETE FROM epochs WHERE epoch_id=1`); err == nil {
		t.Fatal("closed epoch delete bypassed history guard")
	}
	state.RunSeq = 2
	firstRunTwo, err := store.Write(ctx, revision.StreamID, 1, first.Hashes[0], state, save.WriteContext{Cause: "epoch.run2.genesis"})
	if err != nil {
		t.Fatal(err)
	}
	state.ManualTokenMilli--
	if _, err := store.Write(ctx, revision.StreamID, firstRunTwo.Number, first.Hashes[0], state, save.WriteContext{Cause: "epoch.run2.after_genesis"}); err != nil {
		t.Fatal(err)
	}
	if epochID, err := store.PinRunToCurrentEpoch(ctx, revision.StreamID, ownerID, 2, first.Hashes[0]); err != nil || epochID != 2 {
		t.Fatalf("second pin epoch=%d err=%v", epochID, err)
	}
	var pinnedGenesis, firstRunTwoState, latestRunTwoState string
	if err := db.QueryRowContext(ctx, `SELECT convert_from(state,'UTF8') FROM run_genesis WHERE company_stream_id=$1 AND run_seq=2`, revision.StreamID).Scan(&pinnedGenesis); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT state::text FROM save_revisions WHERE stream_id=$1 AND revision=$2`, revision.StreamID, firstRunTwo.Number).Scan(&firstRunTwoState); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT state::text FROM save_revisions WHERE stream_id=$1 ORDER BY revision DESC LIMIT 1`, revision.StreamID).Scan(&latestRunTwoState); err != nil {
		t.Fatal(err)
	}
	if pinnedGenesis != firstRunTwoState || pinnedGenesis == latestRunTwoState {
		t.Fatal("run genesis did not use the first revision of the requested run")
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
	if _, err := db.ExecContext(ctx, `INSERT INTO run_version_drift(company_stream_id,run_seq,observed_version) VALUES($1,1,'9.9.9')`, revision.StreamID); err != nil {
		t.Fatal(err)
	}
	driftedKey := int64(750)
	if _, err := repository.ProjectVerifiedRun(ctx, VerifiedRun{EventID: "01985555-4205-7000-8000-000000000005", RunID: revision.StreamID + ":1", FounderID: founders[0], CategoryID: "category.any", Variables: Variables{}, EpochID: 1, KeyMS: &driftedKey, VerifiedAt: now}); !errors.Is(err, ErrInvalidEpoch) {
		t.Fatalf("drifted projection err=%v", err)
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
	if _, err := db.ExecContext(ctx, `INSERT INTO verified_runs(run_id,event_id,founder_id,category_id,variables,epoch_id,mandate_level,key_ms,verified_at) VALUES('malformed',$1,$2,'category.any',$3,1,0,1,now())`,
		"01985555-4299-7000-8000-000000000099", founders[0], []byte(`{"commons":false,"advisor":false,"glitched":false}`)); err == nil {
		t.Fatal("malformed verified run id bypassed database constraint")
	}
	board, err := repository.TimeBoard(ctx, "category.any", variables, 1, 0, 10, nil)
	if err != nil || len(board) != 3 || board[0].Rank != 1 || board[1].Rank != 1 || board[2].Rank != 3 || !board[0].WorldFirst {
		t.Fatalf("board=%+v err=%v", board, err)
	}
	page, err := repository.TimeBoard(ctx, "category.any", variables, 1, 0, 2, &Cursor{Key: board[0].Key, RunID: board[0].RunID})
	if err != nil || len(page) != 2 || page[0].RunID != board[1].RunID || page[1].Rank != 3 {
		t.Fatalf("page=%+v err=%v", page, err)
	}
	magnitudeKeys := []MagnitudeKey{{Exponent: 15, Mantissa: 125_000_000_000}, {Exponent: 15, Mantissa: 125_000_000_000}, {Exponent: 14, Mantissa: 999_000_000_000}, {Exponent: -1, Mantissa: 900_000_000_000}, {}}
	for index, key := range magnitudeKeys {
		eventIDs := []string{"01985555-4211-7000-8000-000000000011", "01985555-4212-7000-8000-000000000012", "01985555-4213-7000-8000-000000000013", "01985555-4214-7000-8000-000000000014", "01985555-4215-7000-8000-000000000015"}
		founderIndex := index
		if founderIndex >= len(founders)-1 {
			founderIndex = 0
		}
		runSequence := 1
		if index >= len(founders)-1 {
			runSequence = index - len(founders) + 3
		}
		_, err := repository.ProjectVerifiedRun(ctx, VerifiedRun{EventID: eventIDs[index], RunID: fmt.Sprintf("%s:%d", founders[founderIndex], runSequence), FounderID: founders[founderIndex],
			CategoryID: "valuation", Variables: variables, EpochID: 1, KeyMagnitude: &key, VerifiedAt: now.Add(time.Duration(index+10) * time.Second)})
		if err != nil {
			t.Fatalf("magnitude project %d: %v", index, err)
		}
	}
	magnitudeBoard, err := repository.MagnitudeBoard(ctx, "valuation", variables, 1, 0, 10, nil)
	if err != nil || len(magnitudeBoard) != 5 || magnitudeBoard[0].Rank != 1 || magnitudeBoard[1].Rank != 1 || magnitudeBoard[2].Rank != 3 || magnitudeBoard[3].Rank != 4 || magnitudeBoard[4].Rank != 5 ||
		magnitudeBoard[0].Key != magnitudeKeys[0] || magnitudeBoard[3].Key != magnitudeKeys[3] || magnitudeBoard[4].Key != (MagnitudeKey{}) {
		t.Fatalf("magnitude board=%+v err=%v", magnitudeBoard, err)
	}
	magnitudePage, err := repository.MagnitudeBoard(ctx, "valuation", variables, 1, 0, 2, &MagnitudeCursor{Key: magnitudeBoard[0].Key, RunID: magnitudeBoard[0].RunID})
	if err != nil || len(magnitudePage) != 2 || magnitudePage[0].RunID != magnitudeBoard[1].RunID || magnitudePage[1].Rank != 3 {
		t.Fatalf("magnitude page=%+v err=%v", magnitudePage, err)
	}
	assertEpochRows(t, db, 2, 2)
}

func TestEpochSeedReconciliationIntegration(t *testing.T) {
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
	if _, err := db.ExecContext(ctx, `TRUNCATE epochs,catalog_sets RESTART IDENTITY CASCADE`); err != nil {
		t.Fatal(err)
	}
	bundle, err := epochseed.Load(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "changelog"), 0o755); err != nil {
		t.Fatal(err)
	}
	for id := int64(1); id <= bundle.Seed.CurrentEpochID+1; id++ {
		name := fmt.Sprintf("epoch-%d.md", id)
		if err := os.WriteFile(filepath.Join(root, "changelog", name), []byte("# reconciled epoch\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	repository, err := NewRepository(db, root)
	if err != nil {
		t.Fatal(err)
	}
	started := time.Date(2026, 7, 30, 10, 0, 0, 123_456_789, time.UTC)
	if err := repository.ReconcileSeed(ctx, bundle, started); err != nil {
		t.Fatal(err)
	}
	if err := repository.ReconcileSeed(ctx, bundle, started.Add(time.Hour)); err != nil {
		t.Fatalf("idempotent reconcile: %v", err)
	}
	current, err := repository.Current(ctx)
	if err != nil || current.ID != bundle.Seed.CurrentEpochID || len(current.Hashes) != len(bundle.Seed.Epochs[len(bundle.Seed.Epochs)-1].AcceptedHashes) || current.Hashes[len(current.Hashes)-1] != bundle.Hash {
		t.Fatalf("current=%+v err=%v", current, err)
	}
	var artifactCount int
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM catalog_artifacts WHERE constants_hash=$1`, bundle.Hash).Scan(&artifactCount); err != nil || artifactCount != len(bundle.Seed.Artifacts) {
		t.Fatalf("artifact count=%d err=%v", artifactCount, err)
	}

	next := bundle
	next.Seed.CurrentEpochID = bundle.Seed.CurrentEpochID + 1
	nextChangelog := fmt.Sprintf("changelog/epoch-%d.md", next.Seed.CurrentEpochID)
	advancedHotfixHash := "sha256:0000000000000000000000000000000000000000000000000000000000000002"
	next.Seed.Epochs = append(append([]epochseed.Epoch(nil), bundle.Seed.Epochs...), epochseed.Epoch{
		ID: next.Seed.CurrentEpochID, Name: "Phase 0.1", ChangelogRef: nextChangelog, AcceptedHashes: []string{advancedHotfixHash, bundle.Hash},
	})
	if err := epochseed.Validate(next.Seed); err != nil || !epochseed.Accepts(epochseed.Current(next.Seed), next.Hash) {
		t.Fatalf("next seed invalid before reconciliation: %v seed=%+v hash=%s", err, next.Seed, next.Hash)
	}
	if err := repository.ReconcileSeed(ctx, next, started.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	current, err = repository.Current(ctx)
	if err != nil || current.ID != next.Seed.CurrentEpochID || current.Name != "Phase 0.1" || !reflect.DeepEqual(current.Hashes, []string{advancedHotfixHash, bundle.Hash}) {
		t.Fatalf("advanced current=%+v err=%v", current, err)
	}
	var advancedHotfixArtifacts int
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM catalog_artifacts WHERE constants_hash=$1`, advancedHotfixHash).Scan(&advancedHotfixArtifacts); err != nil || advancedHotfixArtifacts != 0 {
		t.Fatalf("advanced hotfix artifacts=%d err=%v", advancedHotfixArtifacts, err)
	}
	missingChangelog := fmt.Sprintf("changelog/epoch-%d.md", next.Seed.CurrentEpochID+1)
	if _, err := repository.MintEpoch(ctx, "Phase 0.2", started.Add(2*time.Hour), missingChangelog, artifactsFromBundle(bundle)); !errors.Is(err, ErrInvalidEpoch) {
		t.Fatalf("missing changelog should fail before sequence allocation: %v", err)
	}

	if _, err := db.ExecContext(ctx, `TRUNCATE epochs,catalog_sets RESTART IDENTITY CASCADE`); err != nil {
		t.Fatal(err)
	}
	historicalHash := "sha256:0000000000000000000000000000000000000000000000000000000000000001"
	history := bundle
	history.Seed.CurrentEpochID = 2
	history.Seed.Epochs = []epochseed.Epoch{
		{ID: 1, Name: "Phase 0", ChangelogRef: "changelog/epoch-1.md", AcceptedHashes: []string{historicalHash, bundle.Hash}},
		{ID: 2, Name: "Phase 0.1", ChangelogRef: "changelog/epoch-2.md", AcceptedHashes: []string{bundle.Hash}},
	}
	if err := repository.ReconcileSeed(ctx, history, started.Add(3*time.Hour)); err != nil {
		t.Fatalf("fresh database full-history reconcile: %v", err)
	}
	if err := repository.ReconcileSeed(ctx, history, started.Add(4*time.Hour)); err != nil {
		t.Fatalf("full-history reconcile idempotency: %v", err)
	}
	rows, err := repository.db.QueryContext(ctx, `SELECT epoch_id,ended_at IS NOT NULL FROM epochs ORDER BY epoch_id`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var epochStates []string
	for rows.Next() {
		var id int64
		var ended bool
		if err := rows.Scan(&id, &ended); err != nil {
			t.Fatal(err)
		}
		epochStates = append(epochStates, fmt.Sprintf("%d:%t", id, ended))
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(epochStates, []string{"1:true", "2:false"}) {
		t.Fatalf("fresh history epoch states=%v", epochStates)
	}
	var accepted, historicalArtifacts int
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM epoch_hashes`).Scan(&accepted); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM catalog_artifacts WHERE constants_hash=$1`, historicalHash).Scan(&historicalArtifacts); err != nil {
		t.Fatal(err)
	}
	if accepted != 3 || historicalArtifacts != 0 {
		t.Fatalf("fresh history accepted=%d historical artifacts=%d", accepted, historicalArtifacts)
	}
}

func artifactsFromBundle(bundle epochseed.Bundle) []Artifact {
	artifacts := make([]Artifact, 0, len(bundle.Seed.Artifacts))
	for _, declaration := range bundle.Seed.Artifacts {
		artifacts = append(artifacts, Artifact{Name: declaration.Name, Bytes: bundle.Artifacts[declaration.Name]})
	}
	return artifacts
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

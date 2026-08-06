package save

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"cloud-clicker/server/economy"
	"cloud-clicker/server/multiplier"
)

func TestRunFrozenContributionsAreCompleteImmutableAndRetryExactIntegration(t *testing.T) {
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
	fiscalArtifact := []byte(`{"schema_version":1,"generator_level_rows":[{"source_id":"fiscal.generator.example"}]}`)
	artifacts := map[string][]byte{"economy": []byte(stateCatalogJSON), "fiscal": fiscalArtifact}
	hash, err := ConstantsHashArtifacts(artifacts)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO catalog_sets(constants_hash) VALUES($1)`, hash); err != nil {
		t.Fatal(err)
	}
	for name, data := range artifacts {
		if _, err := db.ExecContext(ctx, `INSERT INTO catalog_artifacts(constants_hash,artifact_name,bytes) VALUES($1,$2,$3)`, hash, name, data); err != nil {
			t.Fatal(err)
		}
	}
	var epochID int64
	if err := db.QueryRowContext(ctx, `INSERT INTO epochs(name,started_at,changelog_ref) VALUES('Fiscal test',now(),'changelog/epoch-1.md') RETURNING epoch_id`).Scan(&epochID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO epoch_hashes(epoch_id,constants_hash) VALUES($1,$2)`, epochID, hash); err != nil {
		t.Fatal(err)
	}
	catalog := stateCatalog(t)
	store, err := NewStore(db, catalogMap{hash: catalog}, nil)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 8, 6, 12, 0, 0, 0, time.UTC)
	ownerID := "01989999-6500-7000-8000-000000000001"
	ledger, _ := economy.NewLedger(catalog, economy.ScopeCompany)
	state := &State{Ledger: ledger, GeneratorCounts: map[string]int64{"generator.example": 0},
		EvaluatedThrough: now, ManualTokenRefilledAt: now, RunSeq: 1, RunStartedAt: now}
	revision, err := store.CreateStream(ctx, StreamKey{OwnerKind: OwnerFounder, OwnerID: ownerID,
		Scope: economy.ScopeCompany}, hash, state, WriteContext{Cause: "frozen_contribution_test"})
	if err != nil {
		t.Fatal(err)
	}
	encoded, _ := EncodeState(state)
	values := []FrozenContribution{{SourceID: "fiscal.generator.example", Slot: multiplier.SlotPrestige,
		Target: "generator.example", Factor: "1e0"}, {SourceID: "fiscal.hoard", Slot: multiplier.SlotPrestige,
		Target: "all", Factor: "1e0"}}
	tx, _ := db.BeginTx(ctx, nil)
	defer tx.Rollback()
	if _, err := PinRunWithGenesisTx(ctx, tx, revision.StreamID, ownerID, 1, hash, CurrentVersion, encoded); err != nil {
		t.Fatal(err)
	}
	if err := InsertRunFrozenContributionsTx(ctx, tx, revision.StreamID, 1, values); err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
	loaded, err := LoadRunFrozenContributions(ctx, db, revision.StreamID, 1)
	if err != nil || len(loaded) != 2 {
		t.Fatalf("loaded=%+v err=%v", loaded, err)
	}
	retry, _ := db.BeginTx(ctx, nil)
	defer retry.Rollback()
	conflicting := append([]FrozenContribution(nil), values...)
	conflicting[0].Factor = "2e0"
	if err := InsertRunFrozenContributionsTx(ctx, retry, revision.StreamID, 1, conflicting); !errors.Is(err, ErrInvalidState) {
		t.Fatalf("conflicting retry error=%v", err)
	}
	_ = retry.Rollback()
	if _, err := db.ExecContext(ctx, `UPDATE run_frozen_contributions SET factor='2' WHERE company_stream_id=$1`, revision.StreamID); err == nil {
		t.Fatal("frozen contribution was mutable")
	}

	orphan, _ := db.BeginTx(ctx, nil)
	defer orphan.Rollback()
	if _, err := PinRunToCurrentEpochTx(ctx, orphan, revision.StreamID, ownerID, 2, hash); err != nil {
		t.Fatal(err)
	}
	if err := InsertRunGenesisTx(ctx, orphan, RunGenesis{CompanyStreamID: revision.StreamID, RunSeq: 2,
		State: encoded, Version: CurrentVersion, ConstantsHash: hash}); err != nil {
		t.Fatal(err)
	}
	if err := orphan.Commit(); err == nil {
		t.Fatal("Fiscal run committed without its complete frozen contribution set")
	}
}

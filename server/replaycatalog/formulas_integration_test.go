package replaycatalog

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"cloud-clicker/server/epochseed"
	"cloud-clicker/server/leaderboard"
	"cloud-clicker/server/save"
)

func TestStoredFormulaReplayIntegration(t *testing.T) {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}
	ctx := context.Background()
	db, err := save.OpenPostgres(ctx, url)
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
	root := filepath.Join("..", "..")
	seed, err := epochseed.Load(root)
	if err != nil {
		t.Fatal(err)
	}
	staged := t.TempDir()
	if err := os.MkdirAll(filepath.Join(staged, "changelog"), 0o755); err != nil {
		t.Fatal(err)
	}
	for id := int64(1); id <= seed.Seed.CurrentEpochID+2; id++ {
		if err := os.WriteFile(filepath.Join(staged, "changelog", fmt.Sprintf("epoch-%d.md", id)), []byte("# Replay fixture\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	repository, err := leaderboard.NewRepository(db, staged)
	if err != nil {
		t.Fatal(err)
	}
	started := time.Date(2026, 10, 7, 12, 0, 0, 0, time.UTC)
	if err := repository.ReconcileSeed(ctx, seed, started); err != nil {
		t.Fatal(err)
	}
	formulaBytes, err := os.ReadFile(filepath.Join(root, "docs", "generated", "production-formulas.json"))
	if err != nil {
		t.Fatal(err)
	}
	mint := make([]leaderboard.Artifact, 0, len(seed.Artifacts)+1)
	for name, data := range seed.Artifacts {
		mint = append(mint, leaderboard.Artifact{Name: name, Bytes: data})
	}
	mint = append(mint, leaderboard.Artifact{Name: "formulas", Bytes: formulaBytes})
	epoch, err := repository.MintEpoch(ctx, "Stored replay fixture", started.Add(time.Hour), fmt.Sprintf("changelog/epoch-%d.md", seed.Seed.CurrentEpochID+1), mint)
	if err != nil || len(epoch.Hashes) != 1 {
		t.Fatalf("mint=%+v err=%v", epoch, err)
	}
	set, err := LoadDatabase(ctx, db)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(set[epoch.Hashes[0]].Artifacts["formulas"], formulaBytes) {
		t.Fatal("database load rewrote stored formulas")
	}
	if _, ok := set.ResolvePrestige(epoch.Hashes[0]); !ok {
		t.Fatal("stored formula-bearing epoch cannot resolve")
	}
	if _, ok := set.ResolvePrestige(seed.Hash); !ok {
		t.Fatal("older epoch cannot resolve after new mint")
	}
	if _, present := set[seed.Hash].Artifacts["formulas"]; present {
		t.Fatal("older epoch gained current formulas")
	}
	public, err := repository.PublicCatalog(ctx, epoch.Hashes[0])
	if err != nil {
		t.Fatal(err)
	}
	downloaded := make(map[string][]byte, len(public.Artifacts))
	for _, artifact := range public.Artifacts {
		downloaded[artifact.Name] = artifact.Bytes
	}
	loaded, err := Load(public.ConstantsHash, downloaded)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := loaded.ResolvePrestige(public.ConstantsHash); !ok {
		t.Fatal("public source bytes cannot resolve")
	}

	// Valid JSON and a matching new bundle hash must not disguise a malformed
	// formula schema. This is another test-only mint, never a product epoch.
	mint[len(mint)-1].Bytes = bytes.Replace(formulaBytes, []byte(`"schema_version": 14`), []byte(`"schema_version": 15`), 1)
	if _, err := repository.MintEpoch(ctx, "Malformed formula fixture", started.Add(2*time.Hour), fmt.Sprintf("changelog/epoch-%d.md", seed.Seed.CurrentEpochID+2), mint); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadDatabase(ctx, db); err == nil {
		t.Fatal("hash-matching unsupported stored formula accepted")
	}
}

package leaderboard

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"cloud-clicker/server/epochseed"
	"cloud-clicker/server/save"
)

func TestPublicBoardRankingAndPagesIntegration(t *testing.T) {
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
	bundle, err := epochseed.Load(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "changelog"), 0o755); err != nil {
		t.Fatal(err)
	}
	for id := int64(1); id <= bundle.Seed.CurrentEpochID; id++ {
		if err := os.WriteFile(filepath.Join(root, "changelog", fmt.Sprintf("epoch-%d.md", id)), []byte("# epoch\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	repository, err := NewRepository(db, root)
	if err != nil {
		t.Fatal(err)
	}
	if err := repository.ReconcileSeed(ctx, bundle, time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)); err != nil {
		t.Fatal(err)
	}
	epoch := bundle.Seed.CurrentEpochID
	for category, want := range map[string]RankingKind{"any_percent": RankingTimeMS, "ethical_percent": RankingTimeMS, "valuation": RankingMagnitude} {
		if kind, err := repository.PublicBoardRankingKind(ctx, category, epoch); err != nil || kind != want {
			t.Fatalf("%s kind=%q err=%v", category, kind, err)
		}
	}
	if _, err := repository.PublicBoardRankingKind(ctx, "no_such_category", epoch); !errors.Is(err, ErrUnknownPublicCategory) {
		t.Fatalf("unknown category: %v", err)
	}
	if _, err := repository.PublicBoardRankingKind(ctx, "any_percent", epoch+1); !errors.Is(err, ErrUnknownPublicEpoch) {
		t.Fatalf("unknown epoch: %v", err)
	}

	verified := time.Date(2026, 9, 2, 12, 0, 0, 0, time.UTC)
	insert := func(index int, category string, keyMS *int64, exponent, mantissa *int64, variables string) {
		t.Helper()
		eventID := fmt.Sprintf("01986666-0000-7000-8000-%012d", index)
		if _, err := db.ExecContext(ctx, `INSERT INTO verification_projection_events(event_id) VALUES($1)`, eventID); err != nil {
			t.Fatal(err)
		}
		if _, err := db.ExecContext(ctx, `INSERT INTO verified_runs(run_id,event_id,founder_id,category_id,variables,epoch_id,mandate_level,key_ms,key_exponent,key_mantissa,verified_at,world_first)
			VALUES($1,$2,$3,$4,$5,$6,0,$7,$8,$9,$10,false)`,
			fmt.Sprintf("01986666-0000-4000-8000-%012d:1", index), eventID, fmt.Sprintf("01986666-0000-4000-8000-%012d", 100+index),
			category, variables, epoch, keyMS, exponent, mantissa, verified); err != nil {
			t.Fatal(err)
		}
	}
	plain := `{"commons":false,"advisor":false,"glitched":false,"faction":null}`
	key := func(value int64) *int64 { return &value }
	insert(1, "any_percent", key(3000), nil, nil, plain)
	insert(2, "any_percent", key(1000), nil, nil, plain)
	insert(3, "any_percent", key(3000), nil, nil, plain)
	insert(4, "any_percent", key(500), nil, nil, `{"commons":false,"advisor":true,"glitched":false,"faction":null}`)
	insert(5, "valuation", nil, key(12), key(100_000_000_000), plain)
	insert(6, "valuation", nil, key(15), key(250_000_000_000), plain)

	query := PublicBoardQuery{CategoryID: "any_percent", EpochID: epoch, Limit: 2}
	first, more, err := repository.PublicBoardPage(ctx, RankingTimeMS, query)
	if err != nil || !more || len(first) != 2 || first[0].Key != 1000 || first[0].Rank != 1 || first[1].Key != 3000 || first[1].Rank != 2 {
		t.Fatalf("first=%+v more=%v err=%v", first, more, err)
	}
	query.AfterKey, query.AfterRunID = &first[1].Key, first[1].RunID
	second, more, err := repository.PublicBoardPage(ctx, RankingTimeMS, query)
	if err != nil || more || len(second) != 1 || second[0].Key != 3000 || second[0].Rank != 2 || second[0].RunID <= first[1].RunID {
		t.Fatalf("second=%+v more=%v err=%v", second, more, err)
	}
	advisor, _, err := repository.PublicBoardPage(ctx, RankingTimeMS, PublicBoardQuery{CategoryID: "any_percent", EpochID: epoch, Limit: 10,
		Variables: Variables{Advisor: true}})
	if err != nil || len(advisor) != 1 || advisor[0].Key != 500 || advisor[0].Rank != 1 {
		t.Fatalf("the variables filter must partition boards: %+v %v", advisor, err)
	}
	magnitude, more, err := repository.PublicBoardPage(ctx, RankingMagnitude, PublicBoardQuery{CategoryID: "valuation", EpochID: epoch, Limit: 1})
	if err != nil || !more || len(magnitude) != 1 || magnitude[0].Magnitude != (MagnitudeKey{Exponent: 15, Mantissa: 250_000_000_000}) {
		t.Fatalf("magnitude=%+v more=%v err=%v", magnitude, more, err)
	}
	exact, more, err := repository.PublicBoardPage(ctx, RankingTimeMS, PublicBoardQuery{CategoryID: "any_percent", EpochID: epoch, Limit: 3})
	if err != nil || more || len(exact) != 3 {
		t.Fatalf("an exactly-full page must not claim more: %+v %v %v", exact, more, err)
	}

	// A stored catalog in the epoch that cannot load (here: a non-canonical
	// timer, which the category loader rejects) fails every lookup loudly; it
	// is never skipped. Timer disagreement between LOADABLE catalogs is
	// unreachable through the canonical-shape loader and is witnessed on the
	// pure resolver in public_boards_test.go.
	var routesBytes, categoriesBytes []byte
	if err := db.QueryRowContext(ctx, `SELECT bytes FROM catalog_artifacts WHERE constants_hash=$1 AND artifact_name='routes'`, bundle.Hash).Scan(&routesBytes); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT bytes FROM catalog_artifacts WHERE constants_hash=$1 AND artifact_name='categories'`, bundle.Hash).Scan(&categoriesBytes); err != nil {
		t.Fatal(err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(categoriesBytes, &decoded); err != nil {
		t.Fatal(err)
	}
	rewritten := false
	for _, raw := range decoded["categories"].([]any) {
		if category := raw.(map[string]any); category["id"] == "any_percent" {
			category["timer"], rewritten = "none", true
		}
	}
	conflicting, err := json.Marshal(decoded)
	if err != nil || !rewritten {
		t.Fatalf("fixture rewrite did not apply: %v", err)
	}
	foreign := "sha256:" + strings.Repeat("f", 64)
	for _, statement := range []struct {
		query string
		args  []any
	}{
		{`INSERT INTO catalog_sets(constants_hash) VALUES($1)`, []any{foreign}},
		{`INSERT INTO catalog_artifacts(constants_hash,artifact_name,bytes) VALUES($1,'categories',$2),($1,'routes',$3)`, []any{foreign, conflicting, routesBytes}},
		{`INSERT INTO epoch_hashes(epoch_id,constants_hash) VALUES($1,$2)`, []any{epoch, foreign}},
	} {
		if _, err := db.ExecContext(ctx, statement.query, statement.args...); err != nil {
			t.Fatal(err)
		}
	}
	for _, category := range []string{"any_percent", "valuation"} {
		if kind, err := repository.PublicBoardRankingKind(ctx, category, epoch); !errors.Is(err, ErrInvalidEpoch) {
			t.Fatalf("%s resolved kind=%q err=%v past an unloadable stored catalog", category, kind, err)
		}
	}
}

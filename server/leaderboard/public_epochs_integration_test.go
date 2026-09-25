package leaderboard

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"cloud-clicker/server/save"
)

func TestPublicEpochPageIntegration(t *testing.T) {
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
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "changelog"), 0o755); err != nil {
		t.Fatal(err)
	}
	for id := 1; id <= 3; id++ {
		body := "# Epoch " + strings.Repeat("I", id) + "\n\nNotes.\n"
		if err := os.WriteFile(filepath.Join(root, "changelog", "epoch-"+string(rune('0'+id))+".md"), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	hash := func(fill string) string { return "sha256:" + strings.Repeat(fill, 64) }
	started := time.Date(2026, 7, 29, 12, 0, 0, 123456000, time.UTC)
	for _, statement := range []struct {
		query string
		args  []any
	}{
		{`INSERT INTO catalog_sets(constants_hash) VALUES($1),($2),($3),($4)`, []any{hash("a"), hash("b"), hash("c"), hash("0")}},
		{`INSERT INTO epochs(epoch_id,name,started_at,ended_at,changelog_ref) VALUES(1,'Phase 0',$1,$2,'changelog/epoch-1.md')`, []any{started, started.Add(time.Hour)}},
		{`INSERT INTO epochs(epoch_id,name,started_at,ended_at,changelog_ref) VALUES(2,'Second',$1,$2,'changelog/epoch-2.md')`, []any{started.Add(time.Hour), started.Add(2 * time.Hour)}},
		{`INSERT INTO epochs(epoch_id,name,started_at,changelog_ref) VALUES(3,'Current',$1,'changelog/epoch-3.md')`, []any{started.Add(2 * time.Hour)}},
		// Inserted out of byte order so the page must sort, not echo insertion order.
		{`INSERT INTO epoch_hashes(epoch_id,constants_hash) VALUES(1,$1),(3,$2),(3,$3),(3,$4)`, []any{hash("a"), hash("c"), hash("0"), hash("b")}},
	} {
		if _, err := db.ExecContext(ctx, statement.query, statement.args...); err != nil {
			t.Fatal(err)
		}
	}
	repository, err := NewRepository(db, root)
	if err != nil {
		t.Fatal(err)
	}

	first, more, err := repository.PublicEpochPage(ctx, 0, 2)
	if err != nil || !more || len(first) != 2 {
		t.Fatalf("first page=%+v more=%v err=%v", first, more, err)
	}
	if first[0].ID != 3 || first[1].ID != 2 {
		t.Fatalf("page is not newest-first: %d,%d", first[0].ID, first[1].ID)
	}
	if first[0].EndedAt != nil || first[1].EndedAt == nil || !first[1].EndedAt.Equal(started.Add(2*time.Hour)) {
		t.Fatalf("ended_at mismatch: %+v %+v", first[0].EndedAt, first[1].EndedAt)
	}
	if !reflect.DeepEqual(first[0].Hashes, []string{hash("0"), hash("b"), hash("c")}) {
		t.Fatalf("hashes are not byte-sorted: %v", first[0].Hashes)
	}
	if first[1].Hashes == nil || len(first[1].Hashes) != 0 {
		t.Fatalf("hashless epoch must be an empty non-nil list: %#v", first[1].Hashes)
	}
	if string(first[0].Changelog) != "# Epoch III\n\nNotes.\n" || first[0].ChangelogRef != "changelog/epoch-3.md" || first[0].Name != "Current" {
		t.Fatalf("epoch 3 content mismatch: %+v", first[0])
	}
	if !first[0].StartedAt.Equal(started.Add(2 * time.Hour)) {
		t.Fatalf("started_at mismatch: %v", first[0].StartedAt)
	}

	rest, more, err := repository.PublicEpochPage(ctx, first[1].ID, 2)
	if err != nil || more || len(rest) != 1 || rest[0].ID != 1 || !reflect.DeepEqual(rest[0].Hashes, []string{hash("a")}) {
		t.Fatalf("keyset continuation=%+v more=%v err=%v", rest, more, err)
	}
	if empty, more, err := repository.PublicEpochPage(ctx, 1, 2); err != nil || more || len(empty) != 0 || empty == nil {
		t.Fatalf("exhausted page=%#v more=%v err=%v", empty, more, err)
	}
	for _, invalid := range [][2]int64{{0, 0}, {0, 101}, {-1, 5}} {
		if _, _, err := repository.PublicEpochPage(ctx, invalid[0], int(invalid[1])); err == nil {
			t.Fatalf("invalid page request %v accepted", invalid)
		}
	}

	// A changelog that disappeared from the staged content fails closed.
	if err := os.Remove(filepath.Join(root, "changelog", "epoch-3.md")); err != nil {
		t.Fatal(err)
	}
	if _, _, err := repository.PublicEpochPage(ctx, 0, 1); err == nil {
		t.Fatal("missing changelog served")
	}
}

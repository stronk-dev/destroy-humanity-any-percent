package save

import (
	"bytes"
	"context"
	"database/sql"
	"os"
	"testing"
	"time"
)

func TestPlayerOutboxIdentityMigrationDownRemainsStreamScoped(t *testing.T) {
	data, err := embeddedMigrations.ReadFile("migrations/00041_player_outbox_stream_identity.sql")
	if err != nil {
		t.Fatal(err)
	}
	downMarker := []byte("-- +goose Down")
	parts := bytes.Split(data, downMarker)
	if len(parts) != 2 {
		t.Fatalf("migration Down sections=%d", len(parts)-1)
	}
	down := parts[1]
	streamScoped := []byte("UNIQUE (message_kind,stream_id,source_id)")
	global := []byte("UNIQUE (message_kind,source_id)")
	if !bytes.Contains(down, streamScoped) || bytes.Contains(down, global) {
		t.Fatalf("Down must preserve stream-scoped source identity:\n%s", down)
	}
}

func TestHistoricalEventMigrationDownIsDataPreserving(t *testing.T) {
	data, err := embeddedMigrations.ReadFile("migrations/00042_transport_historical_events.sql")
	if err != nil {
		t.Fatal(err)
	}
	parts := bytes.Split(data, []byte("-- +goose Down"))
	if len(parts) != 2 {
		t.Fatalf("migration Down sections=%d", len(parts)-1)
	}
	down := bytes.ToUpper(parts[1])
	if bytes.Contains(down, []byte("DELETE FROM")) || bytes.Contains(down, []byte("TRUNCATE")) ||
		!bytes.Contains(down, []byte("RAISE EXCEPTION")) {
		t.Fatalf("Down must refuse unsafe rollback without deleting event rows:\n%s", parts[1])
	}
}

func TestCommonsSampleRunBackfillDropsPredatingSampleIdentity(t *testing.T) {
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
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `
		CREATE TEMP TABLE company_compact_memberships(
			company_stream_id uuid PRIMARY KEY, run_seq bigint NOT NULL, updated_at timestamptz NOT NULL
		);
		CREATE TEMP TABLE commons_member_samples(
			company_stream_id uuid PRIMARY KEY, run_seq bigint NOT NULL CHECK(run_seq>0), updated_at timestamptz NOT NULL
		)`); err != nil {
		t.Fatal(err)
	}
	membershipAt := time.Date(2026, 8, 2, 12, 0, 0, 0, time.UTC)
	const staleStream = "018f0000-0000-7000-8000-000000000045"
	const currentStream = "018f0000-0000-7000-8000-000000000046"
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO company_compact_memberships(company_stream_id,run_seq,updated_at) VALUES($1,2,$3),($2,2,$3);
		INSERT INTO commons_member_samples(company_stream_id,run_seq,updated_at) VALUES($1,2,$3-interval '1 second'),($2,2,$3+interval '1 second')`,
		staleStream, currentStream, membershipAt); err != nil {
		t.Fatal(err)
	}
	migration, err := embeddedMigrations.ReadFile("migrations/00045_commons_sample_run_backfill.sql")
	if err != nil {
		t.Fatal(err)
	}
	up := bytes.Split(migration, []byte("-- +goose Down"))[0]
	if _, err := tx.ExecContext(ctx, string(up)); err != nil {
		t.Fatal(err)
	}
	var stale, current sql.NullInt64
	if err := tx.QueryRowContext(ctx, `SELECT run_seq FROM commons_member_samples WHERE company_stream_id=$1`, staleStream).Scan(&stale); err != nil {
		t.Fatal(err)
	}
	if err := tx.QueryRowContext(ctx, `SELECT run_seq FROM commons_member_samples WHERE company_stream_id=$1`, currentStream).Scan(&current); err != nil {
		t.Fatal(err)
	}
	if stale.Valid || !current.Valid || current.Int64 != 2 {
		t.Fatalf("backfill stale=%v current=%v", stale, current)
	}
}

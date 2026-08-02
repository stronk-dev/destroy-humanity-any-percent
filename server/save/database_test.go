package save

import (
	"bytes"
	"context"
	"database/sql"
	"os"
	"testing"
	"testing/fstest"
	"time"

	"github.com/pressly/goose/v3"
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

func TestCommonsSampleRunBackfillDropsPredatingSampleIdentityIntegration(t *testing.T) {
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
	db.SetMaxOpenConns(1)
	if _, err := db.ExecContext(ctx, `CREATE TEMP TABLE company_compact_memberships(
			company_stream_id uuid PRIMARY KEY, run_seq bigint NOT NULL, updated_at timestamptz NOT NULL
		)`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `CREATE TEMP TABLE commons_member_samples(
			company_stream_id uuid PRIMARY KEY, run_seq bigint NOT NULL CHECK(run_seq>0), updated_at timestamptz NOT NULL
		)`); err != nil {
		t.Fatal(err)
	}
	membershipAt := time.Date(2026, 8, 2, 12, 0, 0, 0, time.UTC)
	const staleStream = "018f0000-0000-7000-8000-000000000045"
	const currentStream = "018f0000-0000-7000-8000-000000000046"
	if _, err := db.ExecContext(ctx, `INSERT INTO company_compact_memberships(company_stream_id,run_seq,updated_at) VALUES($1,2,$3),($2,2,$3)`,
		staleStream, currentStream, membershipAt); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO commons_member_samples(company_stream_id,run_seq,updated_at) VALUES($1,2,$3::timestamptz-interval '1 second'),($2,2,$3::timestamptz+interval '1 second')`,
		staleStream, currentStream, membershipAt); err != nil {
		t.Fatal(err)
	}
	migration, err := embeddedMigrations.ReadFile("migrations/00045_commons_sample_run_backfill.sql")
	if err != nil {
		t.Fatal(err)
	}
	provider, err := goose.NewProvider(goose.DialectPostgres, db, fstest.MapFS{
		"00001_commons_sample_run_backfill.sql": {Data: migration},
	}, goose.WithDisableVersioning(true))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := provider.ApplyVersion(ctx, 1, true); err != nil {
		t.Fatal(err)
	}
	var stale, current sql.NullInt64
	if err := db.QueryRowContext(ctx, `SELECT run_seq FROM commons_member_samples WHERE company_stream_id=$1`, staleStream).Scan(&stale); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT run_seq FROM commons_member_samples WHERE company_stream_id=$1`, currentStream).Scan(&current); err != nil {
		t.Fatal(err)
	}
	if stale.Valid || !current.Valid || current.Int64 != 2 {
		t.Fatalf("backfill stale=%v current=%v", stale, current)
	}
	if _, err := provider.ApplyVersion(ctx, 1, false); err == nil {
		t.Fatal("Goose rollback relabeled an intentionally invalidated stale sample")
	}
}

func TestGuildClearingMembershipIdentityRollbackFailsClosedIntegration(t *testing.T) {
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
	db.SetMaxOpenConns(1)
	if _, err := db.ExecContext(ctx, `CREATE TEMP TABLE guild_members(membership_id uuid PRIMARY KEY)`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `CREATE TEMP TABLE guild_clearing_results(
		guild_id uuid NOT NULL,
		boundary_seq bigint NOT NULL,
		founder_id uuid,
		company_stream_id uuid,
		run_seq bigint
	)`); err != nil {
		t.Fatal(err)
	}
	migration, err := embeddedMigrations.ReadFile("migrations/00046_guild_clearing_membership_identity.sql")
	if err != nil {
		t.Fatal(err)
	}
	provider, err := goose.NewProvider(goose.DialectPostgres, db, fstest.MapFS{
		"00001_guild_clearing_membership_identity.sql": {Data: migration},
	}, goose.WithDisableVersioning(true))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := provider.ApplyVersion(ctx, 1, true); err != nil {
		t.Fatal(err)
	}
	const membershipID = "018f0000-0000-4000-8000-000000000046"
	if _, err := db.ExecContext(ctx, `INSERT INTO guild_members(membership_id) VALUES($1)`, membershipID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO guild_clearing_results(guild_id,boundary_seq,founder_id,company_stream_id,run_seq,membership_id)
		VALUES('018f0000-0000-7000-8000-000000000046',1,'018f0000-0000-4000-8000-000000000047','018f0000-0000-4000-8000-000000000048',1,$1)`, membershipID); err != nil {
		t.Fatal(err)
	}
	if _, err := provider.ApplyVersion(ctx, 1, false); err == nil {
		t.Fatal("Goose rollback discarded attributed Guild membership identity")
	}
}

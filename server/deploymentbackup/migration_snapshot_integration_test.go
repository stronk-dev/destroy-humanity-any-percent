package deploymentbackup

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"os"
	"testing"
	"time"

	"cloud-clicker/server/save"
	"filippo.io/age"
)

type dumpCheckpointRunner struct {
	beforeDump func(context.Context) error
	dumps      int
}

func (runner *dumpCheckpointRunner) Run(ctx context.Context, name string, args, environment []string, output io.Writer) error {
	if name == "pg_dump" && len(args) > 1 {
		runner.dumps++
		if runner.beforeDump != nil {
			if err := runner.beforeDump(ctx); err != nil {
				return err
			}
		}
	}
	return (ExecRunner{}).Run(ctx, name, args, environment, output)
}

func TestPostgresBackupBindsMigrationAndDumpSnapshotIntegration(t *testing.T) {
	sourceURL, targetURL, adminURL := os.Getenv("TEST_DATABASE_URL"), os.Getenv("TEST_RESTORE_DATABASE_URL"), os.Getenv("TEST_ADMIN_DATABASE_URL")
	if sourceURL == "" || targetURL == "" || adminURL == "" {
		t.Skip("deployment backup database URLs not set")
	}
	// Same completion guard as the existing real snapshot/cancellation witness;
	// it is not a production backup/restore latency bound.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	resetDatabase(t, adminURL, "cloud_clicker_source")
	resetDatabase(t, adminURL, "cloud_clicker_restore")
	source, err := save.OpenPostgres(ctx, sourceURL)
	if err != nil {
		t.Fatal(err)
	}
	defer source.Close()
	if err := save.Migrate(ctx, source); err != nil {
		t.Fatal(err)
	}
	seedBackupEpoch(t, source)
	seedBackupPopulation(t, source)
	var migration int
	if err := source.QueryRowContext(ctx, `SELECT max(version_id) FILTER (WHERE is_applied) FROM goose_db_version`).Scan(&migration); err != nil {
		t.Fatal(err)
	}
	workspace := t.TempDir()
	identity, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatal(err)
	}
	epoch := fixtureEpoch(t, 8)
	manifest := fixtureReleaseManifestForMigration(t, epoch, migration)
	sourceSecret := writeSecret(t, workspace, "source-url", sourceURL)
	before, err := InspectRecoveryIdentity(ctx, RecoveryIdentityInput{DatabaseURLFile: sourceSecret})
	if err != nil {
		t.Fatal(err)
	}
	runner := &dumpCheckpointRunner{beforeDump: func(ctx context.Context) error {
		// This transaction runs only after the backup's migration check, just
		// before its actual pg_dump. It changes real schema, version and rows.
		change, err := source.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		defer change.Rollback()
		for _, statement := range []string{
			`CREATE TABLE public.migration_snapshot_probe(id integer PRIMARY KEY)`,
			`INSERT INTO public.migration_snapshot_probe VALUES(1)`,
			fmt.Sprintf(`INSERT INTO goose_db_version(version_id,is_applied) VALUES(%d,true)`, migration+1),
			`UPDATE accounts SET recovery_hash='after-concurrent-migration'`,
		} {
			if _, err := change.ExecContext(ctx, statement); err != nil {
				return err
			}
		}
		return change.Commit()
	}}
	input := PostgresBackupInput{
		Directory: t.TempDir(), BackupID: "20260822T175900Z-000000000006", ServerID: "server",
		StartedAt: time.Now().Add(-time.Minute), Now: time.Now, Recipient: identity.Recipient().String(),
		DatabaseURLFile: sourceSecret, ReleaseManifest: writeRegular(t, workspace, "manifest.json", manifest),
		EpochDeclaration: writeRegular(t, workspace, "epoch.json", epoch), Runner: runner,
	}
	header, path, err := CreatePostgresBackup(ctx, input)
	if err != nil || header.BackupID == "" || runner.dumps != 1 {
		t.Fatalf("actual backup with concurrent migration failed: error=%v dumps=%d", err, runner.dumps)
	}
	requireBackupSnapshotReleased(t, ctx, source)
	var advanced int
	if err := source.QueryRowContext(ctx, `SELECT max(version_id) FILTER (WHERE is_applied) FROM goose_db_version`).Scan(&advanced); err != nil || advanced != migration+1 {
		t.Fatal("concurrent schema/version transaction did not commit")
	}
	t.Log("actual concurrent DDL/version/row transaction committed after the check; backup reported completion")
	if _, err := RestorePostgresBackup(ctx, PostgresRestoreInput{
		BackupPath: path, ExpectedManifestSHA256: digest(manifest), IdentityFile: writeSecret(t, workspace, "identity", identity.String()),
		TargetDatabaseURLFile: writeSecret(t, workspace, "target-url", targetURL),
	}); err != nil {
		t.Fatalf("completed backup does not restore with its declared migration: %v", err)
	}
	after, err := InspectRecoveryIdentity(ctx, RecoveryIdentityInput{DatabaseURLFile: writeSecret(t, workspace, "restored-url", targetURL)})
	if err != nil || CompareRecoveryIdentity(before, after) != nil {
		t.Fatalf("restored dump differs from the checked source snapshot: %v", err)
	}
	t.Log("real encrypted restore retains pre-migration schema, version and complete row identity")

	t.Run("initial mismatch refuses before dump", func(t *testing.T) {
		refusal := &dumpCheckpointRunner{}
		candidate := input
		candidate.Directory, candidate.Runner = t.TempDir(), refusal
		header, path, err := CreatePostgresBackup(ctx, candidate)
		if !errors.Is(err, ErrInvalid) || header != (Header{}) || path != "" || refusal.dumps != 0 {
			t.Fatal("initial migration mismatch did not refuse before dump/publication")
		}
		if entries, err := os.ReadDir(candidate.Directory); err != nil || len(entries) != 0 {
			t.Fatal("migration mismatch published backup bytes")
		}
		requireBackupSnapshotReleased(t, ctx, source)
	})
	for _, cancelDump := range []bool{false, true} {
		name := "dump failure releases snapshot"
		if cancelDump {
			name = "dump cancellation releases snapshot"
		}
		t.Run(name, func(t *testing.T) {
			attempt, cancel := context.WithCancel(ctx)
			defer cancel()
			want := errors.New("fixture dump failure")
			failure := &dumpCheckpointRunner{beforeDump: func(context.Context) error {
				if cancelDump {
					cancel()
					return context.Canceled
				}
				return want
			}}
			if cancelDump {
				want = context.Canceled
			}
			candidate := input
			candidate.Directory, candidate.Runner = t.TempDir(), failure
			candidate.ReleaseManifest = writeRegular(t, t.TempDir(), "advanced-manifest.json", fixtureReleaseManifestForMigration(t, epoch, migration+1))
			header, path, err := CreatePostgresBackup(attempt, candidate)
			if !errors.Is(err, want) || header != (Header{}) || path != "" || failure.dumps != 1 {
				t.Fatalf("dump failure did not refuse before publication: error=%v dumps=%d", err, failure.dumps)
			}
			if entries, err := os.ReadDir(candidate.Directory); err != nil || len(entries) != 0 {
				t.Fatal("failed dump published backup bytes")
			}
			requireBackupSnapshotReleased(t, ctx, source)
		})
	}
}

func requireBackupSnapshotReleased(t *testing.T, ctx context.Context, database *sql.DB) {
	t.Helper()
	probe, err := database.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer probe.Rollback()
	// Cancellation can close a pgx connection asynchronously. Observe actual
	// server lock release instead of assuming it precedes the returned error.
	if _, err := probe.ExecContext(ctx, `LOCK TABLE public.goose_db_version IN ACCESS EXCLUSIVE MODE`); err != nil {
		t.Fatalf("backup retained its migration snapshot lock after returning: %v", err)
	}
}

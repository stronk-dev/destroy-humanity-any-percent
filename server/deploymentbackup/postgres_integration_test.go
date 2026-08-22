package deploymentbackup

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"cloud-clicker/server/save"
	"filippo.io/age"
)

func TestPostgresBackupRestoreEmptyAndPopulatedIdentityIntegration(t *testing.T) {
	sourceURL := os.Getenv("TEST_DATABASE_URL")
	targetURL := os.Getenv("TEST_RESTORE_DATABASE_URL")
	adminURL := os.Getenv("TEST_ADMIN_DATABASE_URL")
	if sourceURL == "" || targetURL == "" || adminURL == "" {
		t.Skip("deployment backup database URLs not set")
	}
	ctx := context.Background()
	resetDatabase(t, adminURL, "cloud_clicker_source")
	source, err := save.OpenPostgres(ctx, sourceURL)
	if err != nil {
		t.Fatal(err)
	}
	defer source.Close()
	if err := save.Migrate(ctx, source); err != nil {
		t.Fatal(err)
	}
	var migration int
	if err := source.QueryRowContext(ctx, `SELECT max(version_id) FILTER (WHERE is_applied) FROM goose_db_version`).Scan(&migration); err != nil {
		t.Fatal(err)
	}

	workspace := t.TempDir()
	sourceSecret := writeSecret(t, workspace, "source-url", sourceURL)
	targetSecret := writeSecret(t, workspace, "target-url", targetURL)
	identity, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatal(err)
	}
	identityPath := writeSecret(t, workspace, "age-identity", identity.String())
	epoch := []byte("{\"schema_version\":1,\"current_epoch_id\":8}\n")
	manifest := fixtureReleaseManifestForMigration(t, epoch, migration)
	manifestPath := writeRegular(t, workspace, "release-manifest.json", manifest)
	epochPath := writeRegular(t, workspace, "epoch.json", epoch)
	manifestHash := digest(manifest)
	wrongManifest := fixtureReleaseManifestForMigration(t, epoch, migration-1)
	if _, _, err := CreatePostgresBackup(ctx, PostgresBackupInput{
		Directory: t.TempDir(), BackupID: "20260822T175900Z-000000000000", ServerID: "server",
		StartedAt: time.Now().Add(-time.Minute), Now: time.Now, Recipient: identity.Recipient().String(),
		DatabaseURLFile: sourceSecret, ReleaseManifest: writeRegular(t, workspace, "wrong-manifest.json", wrongManifest), EpochDeclaration: epochPath,
	}); !errors.Is(err, ErrInvalid) {
		t.Fatalf("live database/release migration mismatch accepted: %v", err)
	}

	for _, population := range []struct {
		name string
		seed bool
	}{{name: "empty"}, {name: "populated", seed: true}} {
		t.Run(population.name, func(t *testing.T) {
			if population.seed {
				seedBackupPopulation(t, source)
			}
			before := readBackupIdentity(t, source)
			resetDatabase(t, adminURL, "cloud_clicker_restore")
			backupDirectory := t.TempDir()
			now := time.Date(2026, 8, 22, 18, 0, 0, 0, time.UTC)
			backupID := map[bool]string{false: "20260822T175900Z-000000000001", true: "20260822T175900Z-000000000002"}[population.seed]
			header, path, err := CreatePostgresBackup(ctx, PostgresBackupInput{
				Directory: backupDirectory, BackupID: backupID, ServerID: "01986666-b001-4000-8000-000000000001",
				StartedAt: now.Add(-time.Minute), Now: func() time.Time { return now }, Recipient: identity.Recipient().String(),
				DatabaseURLFile: sourceSecret, ReleaseManifest: manifestPath, EpochDeclaration: epochPath,
			})
			if err != nil {
				t.Fatal(err)
			}
			restoredHeader, err := RestorePostgresBackup(ctx, PostgresRestoreInput{
				BackupPath: path, ExpectedManifestSHA256: manifestHash, IdentityFile: identityPath,
				TargetDatabaseURLFile: targetSecret,
			})
			if err != nil || restoredHeader != header {
				t.Fatalf("header=%+v restored=%+v err=%v", header, restoredHeader, err)
			}
			target, err := save.OpenPostgres(ctx, targetURL)
			if err != nil {
				t.Fatal(err)
			}
			after := readBackupIdentity(t, target)
			if err := target.Close(); err != nil {
				t.Fatal(err)
			}
			if before != after {
				t.Fatalf("restore identity mismatch\nbefore=%s\nafter=%s", before, after)
			}
		})
	}
}

func TestPostgresRestoreRefusesNonCleanTargetIntegration(t *testing.T) {
	sourceURL := os.Getenv("TEST_DATABASE_URL")
	targetURL := os.Getenv("TEST_RESTORE_DATABASE_URL")
	adminURL := os.Getenv("TEST_ADMIN_DATABASE_URL")
	if sourceURL == "" || targetURL == "" || adminURL == "" {
		t.Skip("deployment backup database URLs not set")
	}
	ctx := context.Background()
	source, err := save.OpenPostgres(ctx, sourceURL)
	if err != nil {
		t.Fatal(err)
	}
	defer source.Close()
	if err := save.Migrate(ctx, source); err != nil {
		t.Fatal(err)
	}
	var migration int
	if err := source.QueryRowContext(ctx, `SELECT max(version_id) FILTER (WHERE is_applied) FROM goose_db_version`).Scan(&migration); err != nil {
		t.Fatal(err)
	}
	resetDatabase(t, adminURL, "cloud_clicker_restore")
	target, err := save.OpenPostgres(ctx, targetURL)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := target.ExecContext(ctx, `CREATE TABLE occupied(id integer)`); err != nil {
		t.Fatal(err)
	}
	_ = target.Close()

	workspace := t.TempDir()
	identity, _ := age.GenerateX25519Identity()
	epoch := []byte("{\"schema_version\":1,\"current_epoch_id\":8}\n")
	manifest := fixtureReleaseManifestForMigration(t, epoch, migration)
	header, path, err := CreatePostgresBackup(ctx, PostgresBackupInput{
		Directory: t.TempDir(), BackupID: "20260822T175900Z-000000000003", ServerID: "server",
		StartedAt: time.Now().Add(-time.Minute), Now: time.Now, Recipient: identity.Recipient().String(),
		DatabaseURLFile: writeSecret(t, workspace, "source-url", sourceURL),
		ReleaseManifest: writeRegular(t, workspace, "manifest.json", manifest), EpochDeclaration: writeRegular(t, workspace, "epoch.json", epoch),
	})
	if err != nil || header.BackupID == "" {
		t.Fatal(err)
	}
	_, err = RestorePostgresBackup(ctx, PostgresRestoreInput{
		BackupPath: path, ExpectedManifestSHA256: digest(manifest), IdentityFile: writeSecret(t, workspace, "identity", identity.String()),
		TargetDatabaseURLFile: writeSecret(t, workspace, "target-url", targetURL),
	})
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("non-clean target accepted: %v", err)
	}
}

func seedBackupPopulation(t *testing.T, database *sql.DB) {
	t.Helper()
	ctx := context.Background()
	const (
		accountID = "01986666-c001-7000-8000-000000000001"
		founderID = "01986666-c002-7000-8000-000000000002"
		streamID  = "01986666-c003-7000-8000-000000000003"
		eventID   = "01986666-c004-7000-8000-000000000004"
		verifyID  = "01986666-c005-7000-8000-000000000005"
		hash      = "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	)
	statements := []string{
		`INSERT INTO accounts(account_id,recovery_hash) VALUES('` + accountID + `','backup-fixture')`,
		`INSERT INTO account_founders(account_id,founder_id) VALUES('` + accountID + `','` + founderID + `')`,
		`INSERT INTO save_streams(id,owner_kind,owner_id,scope) VALUES('` + streamID + `','founder','` + founderID + `','company')`,
		`INSERT INTO save_revisions(stream_id,revision,version,state,constants_hash) VALUES('` + streamID + `',1,1,'{"company":"backup-fixture"}','` + hash + `')`,
		`INSERT INTO events(event_id,stream_id,revision,schema_version,kind,constants_hash,payload) VALUES('` + eventID + `','` + streamID + `',1,1,'generator_purchased','` + hash + `','{"fixture":true}')`,
		`INSERT INTO catalog_sets(constants_hash) VALUES('` + hash + `')`,
		`INSERT INTO epochs(epoch_id,name,started_at,changelog_ref) VALUES(8,'Backup Epoch','2026-08-22T00:00:00Z','changelog/epoch-8.md')`,
		`INSERT INTO epoch_hashes(epoch_id,constants_hash) VALUES(8,'` + hash + `')`,
		`INSERT INTO verification_projection_events(event_id) VALUES('` + verifyID + `')`,
		`INSERT INTO verified_runs(run_id,event_id,founder_id,category_id,variables,epoch_id,mandate_level,key_ms,verified_at) VALUES('` + streamID + `:1','` + verifyID + `','` + founderID + `','category.any','{"commons":false,"advisor":false,"glitched":false,"faction":null}',8,0,1234,'2026-08-22T01:00:00Z')`,
	}
	for _, statement := range statements {
		if _, err := database.ExecContext(ctx, statement); err != nil {
			t.Fatalf("seed %q: %v", statement, err)
		}
	}
}

func readBackupIdentity(t *testing.T, database *sql.DB) string {
	t.Helper()
	queries := []string{
		`SELECT COALESCE(string_agg(account_id::text || ':' || recovery_hash,',' ORDER BY account_id),'') FROM accounts`,
		`SELECT COALESCE(string_agg(account_id::text || ':' || founder_id::text,',' ORDER BY founder_id),'') FROM account_founders`,
		`SELECT COALESCE(string_agg(s.id::text || ':' || r.revision::text || ':' || r.state::text,',' ORDER BY s.id,r.revision),'') FROM save_streams s LEFT JOIN save_revisions r ON r.stream_id=s.id WHERE s.scope='company'`,
		`SELECT COALESCE(string_agg(event_id::text || ':' || kind || ':' || payload::text,',' ORDER BY event_id),'') FROM events`,
		`SELECT COALESCE(string_agg(run_id || ':' || category_id || ':' || key_ms::text,',' ORDER BY run_id,category_id),'') FROM verified_runs`,
		`SELECT COALESCE(string_agg(epoch_id::text || ':' || name,',' ORDER BY epoch_id),'') FROM epochs`,
	}
	identity := ""
	for index, query := range queries {
		var value string
		if err := database.QueryRow(query).Scan(&value); err != nil {
			t.Fatal(err)
		}
		identity += fmt.Sprintf("%d=%s\n", index, value)
	}
	return identity
}

func resetDatabase(t *testing.T, adminURL, name string) {
	t.Helper()
	if name != "cloud_clicker_source" && name != "cloud_clicker_restore" {
		t.Fatal("invalid test database name")
	}
	database, err := save.OpenPostgres(context.Background(), adminURL)
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	if _, err := database.Exec(`SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname=$1 AND pid <> pg_backend_pid()`, name); err != nil {
		t.Fatal(err)
	}
	if _, err := database.Exec(`DROP DATABASE IF EXISTS ` + name); err != nil {
		t.Fatal(err)
	}
	if _, err := database.Exec(`CREATE DATABASE ` + name + ` OWNER cloud_clicker`); err != nil {
		t.Fatal(err)
	}
}

func writeSecret(t *testing.T, directory, name, value string) string {
	t.Helper()
	return writeFile(t, directory, name, []byte(value+"\n"), 0o600)
}

func writeRegular(t *testing.T, directory, name string, value []byte) string {
	t.Helper()
	return writeFile(t, directory, name, value, 0o600)
}

func writeFile(t *testing.T, directory, name string, value []byte, mode os.FileMode) string {
	t.Helper()
	path := filepath.Join(directory, name)
	if err := os.WriteFile(path, value, mode); err != nil {
		t.Fatal(err)
	}
	return path
}

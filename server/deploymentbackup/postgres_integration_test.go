package deploymentbackup

import (
	"bytes"
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"cloud-clicker/server/account"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/save"
	"filippo.io/age"
)

type rightsRestoreCatalog struct{}

func (rightsRestoreCatalog) Resolve(hash string) (*economy.Catalog, bool) {
	return &economy.Catalog{}, hash == "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
}

func TestRequireCleanTargetRejectsNonRelationalObjectsIntegration(t *testing.T) {
	adminURL := os.Getenv("TEST_DATABASE_URL")
	if adminURL == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}
	ctx := context.Background()
	admin, err := save.OpenPostgres(ctx, adminURL)
	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close()
	for _, arm := range []struct {
		name, create, observe string
	}{
		{"empty", "", "SELECT 1"},
		{"table", "CREATE TABLE public.occupied(id integer)", "SELECT count(*) FROM pg_class WHERE relname='occupied'"},
		{"enum", "CREATE TYPE public.occupied AS ENUM ('kept')", "SELECT count(*) FROM pg_type WHERE typname='occupied'"},
		{"domain", "CREATE DOMAIN public.occupied AS integer CHECK (VALUE > 0)", "SELECT count(*) FROM pg_type WHERE typname='occupied'"},
		{"function", "CREATE FUNCTION public.occupied() RETURNS integer LANGUAGE sql AS 'SELECT 7'", "SELECT public.occupied() / 7"},
		{"system_namespace_function", "CREATE FUNCTION pg_catalog.occupied() RETURNS integer LANGUAGE sql AS 'SELECT 7'", "SELECT pg_catalog.occupied() / 7"},
		{"schema", "CREATE SCHEMA occupied", "SELECT count(*) FROM pg_namespace WHERE nspname='occupied'"},
		{"large_object", "SELECT lo_create(123)", "SELECT count(*) FROM pg_largeobject_metadata WHERE oid=123"},
		{"extension", "CREATE EXTENSION hstore WITH SCHEMA pg_catalog", "SELECT count(*) FROM pg_extension WHERE extname='hstore'"},
		{"collation", "CREATE COLLATION public.occupied (provider=libc, locale='C')", "SELECT count(*) FROM pg_collation WHERE collname='occupied'"},
		{"operator", "CREATE OPERATOR public.=== (LEFTARG=integer, RIGHTARG=integer, FUNCTION=pg_catalog.int4eq)", "SELECT count(*) FROM pg_operator WHERE oprname='==='"},
		{"operator_family", "CREATE OPERATOR FAMILY public.occupied USING btree", "SELECT count(*) FROM pg_opfamily WHERE opfname='occupied'"},
		{"text_search_configuration", "CREATE TEXT SEARCH CONFIGURATION public.occupied (COPY=pg_catalog.simple)", "SELECT count(*) FROM pg_ts_config WHERE cfgname='occupied'"},
		{"text_search_dictionary", "CREATE TEXT SEARCH DICTIONARY public.occupied (TEMPLATE=pg_catalog.simple)", "SELECT count(*) FROM pg_ts_dict WHERE dictname='occupied'"},
		{"foreign_data_wrapper", "CREATE FOREIGN DATA WRAPPER occupied", "SELECT count(*) FROM pg_foreign_data_wrapper WHERE fdwname='occupied'"},
		{"publication", "CREATE PUBLICATION occupied", "SELECT count(*) FROM pg_publication WHERE pubname='occupied'"},
		{"default_privileges", "ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT ON TABLES TO PUBLIC", "SELECT count(*) FROM pg_default_acl"},
		{"language", "CREATE LANGUAGE occupied HANDLER pg_catalog.plpgsql_call_handler", "SELECT count(*) FROM pg_language WHERE lanname='occupied'"},
		{"cast", "CREATE CAST (uuid AS integer) WITH INOUT", "SELECT count(*) FROM pg_cast WHERE castsource='uuid'::regtype AND casttarget='integer'::regtype"},
		{"access_method", "CREATE ACCESS METHOD occupied TYPE TABLE HANDLER pg_catalog.heap_tableam_handler", "SELECT count(*) FROM pg_am WHERE amname='occupied'"},
	} {
		t.Run(arm.name, func(t *testing.T) {
			name := "backup_clean_" + strings.ToLower(rand.Text())
			if !databaseName.MatchString(name) {
				t.Fatal("invalid generated database name")
			}
			if _, err := admin.ExecContext(ctx, "CREATE DATABASE "+name+" TEMPLATE template0"); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				if _, err := admin.ExecContext(context.Background(), "DROP DATABASE "+name); err != nil {
					t.Error(err)
				}
			})
			parsed, err := url.Parse(adminURL)
			if err != nil {
				t.Fatal(err)
			}
			parsed.Path = "/" + name
			database, err := save.OpenPostgres(ctx, parsed.String())
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				if err := database.Close(); err != nil {
					t.Error(err)
				}
			})
			if err := RequireCleanTarget(ctx, database); err != nil {
				t.Fatalf("fresh database refused: %v", err)
			}
			if arm.create != "" {
				if _, err := database.ExecContext(ctx, arm.create); err != nil {
					t.Fatal(err)
				}
			}
			err = RequireCleanTarget(ctx, database)
			if arm.create == "" {
				if err != nil {
					t.Fatalf("empty database refused: %v", err)
				}
			} else if !errors.Is(err, ErrNonCleanTarget) {
				t.Errorf("existing %s accepted as clean: %v", arm.name, err)
			}
			var retained int
			if err := database.QueryRowContext(ctx, arm.observe).Scan(&retained); err != nil || retained != 1 {
				t.Fatalf("existing object changed: observed=%d err=%v", retained, err)
			}
		})
	}
}

type rightsRestoreState struct {
	account, activeLink, archivedLink, liveStreams, archivedStreams, board, bystander int
}

func readRightsRestoreState(t *testing.T, db *sql.DB) rightsRestoreState {
	t.Helper()
	const accountID = "01986666-c001-7000-8000-000000000001"
	const founderID = "01986666-c002-7000-8000-000000000002"
	const bystanderID = "01986666-c001-7000-8000-000000000011"
	var state rightsRestoreState
	err := db.QueryRow(`SELECT
		(SELECT count(*) FROM accounts WHERE account_id=$1),
		(SELECT count(*) FROM account_founders WHERE founder_id=$2 AND account_id=$1 AND archived_at IS NULL),
		(SELECT count(*) FROM account_founders WHERE founder_id=$2 AND account_id IS NULL AND archived_at IS NOT NULL),
		(SELECT count(*) FROM save_streams WHERE owner_kind='founder' AND owner_id=$2 AND archived_at IS NULL),
		(SELECT count(*) FROM save_streams WHERE owner_kind='founder' AND owner_id=$2 AND archived_at IS NOT NULL),
		(SELECT count(*) FROM verified_runs WHERE founder_id=$2),
		(SELECT count(*) FROM accounts WHERE account_id=$3)`, accountID, founderID, bystanderID).
		Scan(&state.account, &state.activeLink, &state.archivedLink, &state.liveStreams, &state.archivedStreams, &state.board, &state.bystander)
	if err != nil {
		t.Fatal(err)
	}
	return state
}

func TestPreDeletionBackupRestoresDeletedAccountIntegration(t *testing.T) {
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
	seedBackupEpoch(t, source)
	seedBackupPopulation(t, source)
	if _, err := source.ExecContext(ctx, `INSERT INTO accounts(account_id,recovery_hash) VALUES('01986666-c001-7000-8000-000000000011','backup-bystander')`); err != nil {
		t.Fatal(err)
	}
	pre := rightsRestoreState{account: 1, activeLink: 1, liveStreams: 2, board: 1, bystander: 1}
	post := rightsRestoreState{archivedLink: 1, archivedStreams: 2, board: 1, bystander: 1}
	if got := readRightsRestoreState(t, source); got != pre {
		t.Fatalf("invalid pre-delete population: got=%+v want=%+v", got, pre)
	}
	workspace := t.TempDir()
	identity, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatal(err)
	}
	epoch := fixtureEpoch(t, 8)
	manifest := fixtureReleaseManifestForMigration(t, epoch, migration)
	manifestHash := digest(manifest)
	sourceSecret := writeSecret(t, workspace, "source-url", sourceURL)
	targetSecret := writeSecret(t, workspace, "target-url", targetURL)
	identityPath := writeSecret(t, workspace, "identity", identity.String())
	manifestPath := writeRegular(t, workspace, "manifest.json", manifest)
	epochPath := writeRegular(t, workspace, "epoch.json", epoch)
	now := time.Date(2026, 8, 22, 18, 0, 0, 0, time.UTC)
	makeBackup := func(id string) string {
		t.Helper()
		_, path, err := CreatePostgresBackup(ctx, PostgresBackupInput{
			Directory: t.TempDir(), BackupID: id, ServerID: "01986666-b001-4000-8000-000000000001",
			StartedAt: now.Add(-time.Minute), Now: func() time.Time { return now }, Recipient: identity.Recipient().String(),
			DatabaseURLFile: sourceSecret, ReleaseManifest: manifestPath, EpochDeclaration: epochPath,
		})
		if err != nil {
			t.Fatal(err)
		}
		return path
	}
	preBackup := makeBackup("20260822T175900Z-000000000004")
	const accountID = "01986666-c001-7000-8000-000000000001"
	const hash = "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	repository, err := account.NewRepository(source, rightsRestoreCatalog{}, hash,
		account.SigningKeys{CurrentID: "test", Current: bytes.Repeat([]byte{1}, 32)}, func() time.Time { return now }, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := repository.DeleteAccount(ctx, accountID); err != nil {
		t.Fatal(err)
	}
	if got := readRightsRestoreState(t, source); got != post {
		t.Fatalf("invalid post-delete source: got=%+v want=%+v", got, post)
	}
	postBackup := makeBackup("20260822T175900Z-000000000005")
	for _, arm := range []struct {
		name, path string
		want       rightsRestoreState
	}{{"pre-deletion", preBackup, pre}, {"post-deletion", postBackup, post}} {
		t.Run(arm.name, func(t *testing.T) {
			resetDatabase(t, adminURL, "cloud_clicker_restore")
			if _, err := RestorePostgresBackup(ctx, PostgresRestoreInput{
				BackupPath: arm.path, ExpectedManifestSHA256: manifestHash, IdentityFile: identityPath,
				TargetDatabaseURLFile: targetSecret,
			}); err != nil {
				t.Fatal(err)
			}
			target, err := save.OpenPostgres(ctx, targetURL)
			if err != nil {
				t.Fatal(err)
			}
			defer target.Close()
			if got := readRightsRestoreState(t, target); got != arm.want {
				t.Fatalf("%s restore state: got=%+v want=%+v", arm.name, got, arm.want)
			}
		})
	}
}

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
	epoch := fixtureEpoch(t, 8)
	manifest := fixtureReleaseManifestForMigration(t, epoch, migration)
	manifestPath := writeRegular(t, workspace, "release-manifest.json", manifest)
	epochPath := writeRegular(t, workspace, "epoch.json", epoch)
	manifestHash := digest(manifest)
	seedBackupEpoch(t, source)
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
			beforeRecovery, err := InspectRecoveryIdentity(ctx, RecoveryIdentityInput{DatabaseURLFile: sourceSecret})
			if err != nil {
				t.Fatal(err)
			}
			if population.seed && beforeRecovery.VerifiedRunRows != 1 {
				t.Fatalf("populated recovery identity has %d verified runs, want one", beforeRecovery.VerifiedRunRows)
			}
			if population.seed {
				err = ValidatePopulatedRecoveryIdentity(beforeRecovery)
			} else {
				err = ValidateEmptyRecoveryIdentity(beforeRecovery)
			}
			if err != nil {
				t.Fatalf("%s recovery population invalid: %+v err=%v", population.name, beforeRecovery, err)
			}
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
			afterRecovery, err := InspectRecoveryIdentity(ctx, RecoveryIdentityInput{DatabaseURLFile: targetSecret})
			if err != nil || CompareRecoveryIdentity(beforeRecovery, afterRecovery) != nil {
				t.Fatalf("semantic recovery identity mismatch\nbefore=%+v\nafter=%+v\nerr=%v", beforeRecovery, afterRecovery, err)
			}
		})
	}
}

func TestRecoveryIdentityRejectsProjectionEventWithoutVerifiedRunIntegration(t *testing.T) {
	sourceURL := os.Getenv("TEST_DATABASE_URL")
	adminURL := os.Getenv("TEST_ADMIN_DATABASE_URL")
	if sourceURL == "" || adminURL == "" {
		t.Skip("deployment backup database URLs not set")
	}
	ctx := context.Background()
	resetDatabase(t, adminURL, "cloud_clicker_source")
	database, err := save.OpenPostgres(ctx, sourceURL)
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	if err := save.Migrate(ctx, database); err != nil {
		t.Fatal(err)
	}
	seedBackupEpoch(t, database)
	seedBackupPopulationWithVerifiedRun(t, database, false)
	identity, err := InspectRecoveryIdentity(ctx, RecoveryIdentityInput{
		DatabaseURLFile: writeSecret(t, t.TempDir(), "source-url", sourceURL)})
	if err != nil {
		t.Fatal(err)
	}
	if identity.Board.Rows < 1 || identity.VerifiedRunRows != 0 || ValidatePopulatedRecoveryIdentity(identity) == nil {
		t.Fatalf("projection-event-only board accepted: board=%d verified=%d", identity.Board.Rows, identity.VerifiedRunRows)
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
	epoch := fixtureEpoch(t, 8)
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
	if !errors.Is(err, ErrNonCleanTarget) {
		t.Fatalf("non-clean target not refused as non-clean: %v", err)
	}
}

func TestRecoveryIdentityDetectsSameCountContentMutationIntegration(t *testing.T) {
	sourceURL := os.Getenv("TEST_DATABASE_URL")
	adminURL := os.Getenv("TEST_ADMIN_DATABASE_URL")
	if sourceURL == "" || adminURL == "" {
		t.Skip("deployment backup database URLs not set")
	}
	ctx := context.Background()
	resetDatabase(t, adminURL, "cloud_clicker_source")
	database, err := save.OpenPostgres(ctx, sourceURL)
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	if err := save.Migrate(ctx, database); err != nil {
		t.Fatal(err)
	}
	seedBackupEpoch(t, database)
	seedBackupPopulation(t, database)
	secret := writeSecret(t, t.TempDir(), "source-url", sourceURL)
	before, err := InspectRecoveryIdentity(ctx, RecoveryIdentityInput{DatabaseURLFile: secret})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := database.ExecContext(ctx, `UPDATE accounts SET recovery_hash='same-count-mutated'`); err != nil {
		t.Fatal(err)
	}
	after, err := InspectRecoveryIdentity(ctx, RecoveryIdentityInput{DatabaseURLFile: secret})
	if err != nil {
		t.Fatal(err)
	}
	if before.Player.Rows != after.Player.Rows || before.Database.Rows != after.Database.Rows {
		t.Fatalf("content mutation changed row counts: before=%+v after=%+v", before, after)
	}
	if before.Player.SHA256 == after.Player.SHA256 || before.Database.SHA256 == after.Database.SHA256 {
		t.Fatalf("same-count content mutation escaped identity: before=%+v after=%+v", before, after)
	}
}

func seedBackupPopulation(t *testing.T, database *sql.DB) {
	seedBackupPopulationWithVerifiedRun(t, database, true)
}

func seedBackupPopulationWithVerifiedRun(t *testing.T, database *sql.DB, verifiedRun bool) {
	t.Helper()
	ctx := context.Background()
	const (
		accountID       = "01986666-c001-7000-8000-000000000001"
		founderID       = "01986666-c002-7000-8000-000000000002"
		streamID        = "01986666-c003-7000-8000-000000000003"
		founderStreamID = "01986666-c006-7000-8000-000000000006"
		eventID         = "01986666-c004-7000-8000-000000000004"
		verifyID        = "01986666-c005-7000-8000-000000000005"
		hash            = "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	)
	statements := []string{
		`INSERT INTO accounts(account_id,recovery_hash) VALUES('` + accountID + `','backup-fixture')`,
		`INSERT INTO account_founders(account_id,founder_id) VALUES('` + accountID + `','` + founderID + `')`,
		`INSERT INTO save_streams(id,owner_kind,owner_id,scope) VALUES('` + founderStreamID + `','founder','` + founderID + `','founder')`,
		`INSERT INTO save_revisions(stream_id,revision,version,state,constants_hash) VALUES('` + founderStreamID + `',1,1,'{"founder":"backup-fixture"}','` + hash + `')`,
		`INSERT INTO save_streams(id,owner_kind,owner_id,scope) VALUES('` + streamID + `','founder','` + founderID + `','company')`,
		`INSERT INTO save_revisions(stream_id,revision,version,state,constants_hash) VALUES('` + streamID + `',1,1,'{"company":"backup-fixture"}','` + hash + `')`,
		`INSERT INTO events(event_id,stream_id,revision,schema_version,kind,constants_hash,payload) VALUES('` + eventID + `','` + streamID + `',1,1,'generator_purchased','` + hash + `','{"fixture":true}')`,
		`INSERT INTO verification_projection_events(event_id) VALUES('` + verifyID + `')`,
	}
	if verifiedRun {
		statements = append(statements, `INSERT INTO verified_runs(run_id,event_id,founder_id,category_id,variables,epoch_id,mandate_level,key_ms,verified_at) VALUES('`+streamID+`:1','`+verifyID+`','`+founderID+`','category.any','{"commons":false,"advisor":false,"glitched":false,"faction":null}',8,0,1234,'2026-08-22T01:00:00Z')`)
	}
	for _, statement := range statements {
		if _, err := database.ExecContext(ctx, statement); err != nil {
			t.Fatalf("seed %q: %v", statement, err)
		}
	}
}

func seedBackupEpoch(t *testing.T, database *sql.DB) {
	t.Helper()
	const hash = "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	for _, statement := range []string{
		`INSERT INTO catalog_sets(constants_hash) VALUES('` + hash + `')`,
		`INSERT INTO epochs(epoch_id,name,started_at,changelog_ref) VALUES(8,'Backup Epoch','2026-08-22T00:00:00Z','changelog/epoch-8.md')`,
		`INSERT INTO epoch_hashes(epoch_id,constants_hash) VALUES(8,'` + hash + `')`,
	} {
		if _, err := database.ExecContext(context.Background(), statement); err != nil {
			t.Fatalf("seed epoch %q: %v", statement, err)
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

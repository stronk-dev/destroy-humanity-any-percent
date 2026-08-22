package deploymentrelease

import (
	"bytes"
	"context"
	"crypto/sha256"
	"crypto/tls"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"cloud-clicker/server/account"
	"cloud-clicker/server/deploymentbackup"
	"cloud-clicker/server/epochseed"
	"cloud-clicker/server/gameserver"
	"cloud-clicker/server/releasepackage"
	"cloud-clicker/server/save"
	"filippo.io/age"
)

func TestCaddyIntegrationDrainReleaseAndExactBackupRollback(t *testing.T) {
	databaseURL, restoreURL := os.Getenv("TEST_DATABASE_URL"), os.Getenv("TEST_RESTORE_DATABASE_URL")
	adminURL, origin := os.Getenv("TEST_ADMIN_DATABASE_URL"), os.Getenv("TEST_CADDY_ORIGIN")
	if databaseURL == "" || restoreURL == "" || adminURL == "" || origin == "" {
		t.Skip("deployment release integration environment is required")
	}
	ctx := context.Background()
	database, err := save.OpenPostgres(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	if _, err := database.ExecContext(ctx, `DROP SCHEMA public CASCADE; CREATE SCHEMA public`); err != nil {
		t.Fatal(err)
	}
	if err := save.Migrate(ctx, database); err != nil {
		t.Fatal(err)
	}
	root := "/workspace"
	if _, err := os.Stat(filepath.Join(root, "balance", "epochs", "phase0.json")); err != nil {
		t.Fatal(err)
	}
	keys := account.SigningKeys{CurrentID: "release-integration", Current: bytes.Repeat([]byte{0x71}, 32)}
	bootstrap := account.BootstrapReceiptKeys{CurrentID: "release-bootstrap", Current: bytes.Repeat([]byte{0x72}, 32)}
	// Caddy's integration-only internal CA is intentionally not installed into
	// the test container's host trust store. Production never uses this client.
	client := &http.Client{Timeout: 5 * time.Second, Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}} //nolint:gosec
	composition, httpServer := startReleaseServer(t, ctx, database, root, origin, keys, bootstrap)
	waitCaddy(t, client, origin+"/readyz", http.StatusNoContent)
	session, err := OpenAuthenticatedSmoke(ctx, client, origin, composition.CurrentHash)
	if err != nil {
		t.Fatal(err)
	}
	backup := createReleaseBackup(t, database, databaseURL, root)
	expectedEpoch, err := epochseed.Load(root)
	if err != nil {
		t.Fatal(err)
	}
	inspection, err := deploymentbackup.InspectPostgres(ctx, deploymentbackup.PostgresInspectionInput{
		DatabaseURLFile: backup.databaseURLFile, ReleaseManifest: backup.manifestPath, ContentRoot: root, RequireIdentity: true,
	})
	if err != nil || inspection.DatabaseMigration < 1 || inspection.EpochID != expectedEpoch.Seed.CurrentEpochID ||
		inspection.ConstantsHash != composition.CurrentHash || inspection.ArtifactsVerified < 1 {
		t.Fatalf("release database inspection=%+v err=%v", inspection, err)
	}
	var backedUpAccounts int
	if err := database.QueryRow(`SELECT count(*) FROM accounts`).Scan(&backedUpAccounts); err != nil || backedUpAccounts < 1 {
		t.Fatalf("backed-up accounts=%d err=%v", backedUpAccounts, err)
	}

	transaction, err := database.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := transaction.QueryRowContext(ctx, `SELECT id FROM save_streams WHERE owner_kind='founder' AND owner_id=$1 AND scope='company' FOR UPDATE`, session.founderID).Scan(new(string)); err != nil {
		t.Fatal(err)
	}
	intentDone := make(chan int, 1)
	go func() {
		body := `{"intent_id":"01985555-d001-7000-8000-000000000001","kind":"perform_manual_batch","expected_revision":1,"action_id":"manual.click","count":1,"window_ms":1}`
		request, _ := http.NewRequest(http.MethodPost, origin+"/api/v1/intents", strings.NewReader(body))
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("Authorization", "Bearer "+session.accessToken)
		response, requestErr := client.Do(request)
		if requestErr != nil {
			intentDone <- 0
			return
		}
		defer response.Body.Close()
		intentDone <- response.StatusCode
	}()
	waitForDatabaseLock(t, database)

	drainCtx, cancelDrain := context.WithTimeout(ctx, 10*time.Second)
	defer cancelDrain()
	drainDone := make(chan error, 1)
	go func() { drainDone <- composition.Server.Drain(drainCtx, time.Now().UTC()) }()
	socketDone := make(chan struct {
		courtesy, closed bool
		err              error
	}, 1)
	go func() {
		courtesy, closed, socketErr := session.WaitForDrain(drainCtx)
		socketDone <- struct {
			courtesy, closed bool
			err              error
		}{courtesy, closed, socketErr}
	}()
	if !waitHTTPState(drainCtx, client, origin+"/readyz", false) {
		t.Fatal("Caddy readiness never fell")
	}
	refused := false
	for deadline := time.Now().Add(2 * time.Second); time.Now().Before(deadline) && !refused; {
		refused, _ = session.IntentRefused(drainCtx)
	}
	if !refused {
		t.Fatal("new intent was not refused while an admitted intent drained")
	}
	if err := transaction.Commit(); err != nil {
		t.Fatal(err)
	}
	if status := <-intentDone; status != http.StatusOK {
		t.Fatalf("admitted intent status=%d", status)
	}
	if err := <-drainDone; err != nil {
		t.Fatal(err)
	}
	socket := <-socketDone
	if !socket.courtesy || !socket.closed {
		t.Fatalf("drain socket evidence=%+v", socket)
	}
	shutdownServer(t, httpServer)

	// Startup migrations and epoch reconciliation run again, and the same
	// authenticated HTTP/WebSocket path must recover through the unchanged
	// Caddy hop.
	second, secondHTTP := startReleaseServer(t, ctx, database, root, origin, keys, bootstrap)
	waitCaddy(t, client, origin+"/readyz", http.StatusNoContent)
	secondSmoke, err := OpenAuthenticatedSmoke(ctx, client, origin, second.CurrentHash)
	if err != nil {
		t.Fatal(err)
	}
	_ = secondSmoke.Close()
	secondDrain, cancelSecond := context.WithTimeout(ctx, 10*time.Second)
	defer cancelSecond()
	if err := second.Server.Drain(secondDrain, time.Now().UTC()); err != nil {
		t.Fatal(err)
	}
	shutdownServer(t, secondHTTP)

	resetReleaseDatabase(t, adminURL, "cloud_clicker_release_restore")
	if _, err := deploymentbackup.RestorePostgresBackup(ctx, deploymentbackup.PostgresRestoreInput{
		BackupPath: backup.path, ExpectedManifestSHA256: backup.manifestSHA256,
		IdentityFile: backup.identityPath, TargetDatabaseURLFile: writeReleaseFile(t, "restore-url", restoreURL, 0o600),
	}); err != nil {
		t.Fatal(err)
	}
	restored, err := save.OpenPostgres(ctx, restoreURL)
	if err != nil {
		t.Fatal(err)
	}
	defer restored.Close()
	var restoredAccounts int
	if err := restored.QueryRow(`SELECT count(*) FROM accounts`).Scan(&restoredAccounts); err != nil || restoredAccounts != backedUpAccounts {
		t.Fatalf("restored accounts=%d want=%d err=%v", restoredAccounts, backedUpAccounts, err)
	}
	previous, previousHTTP := startReleaseServer(t, ctx, restored, root, origin, keys, bootstrap)
	waitCaddy(t, client, origin+"/readyz", http.StatusNoContent)
	previousSmoke, err := OpenAuthenticatedSmoke(ctx, client, origin, previous.CurrentHash)
	if err != nil {
		t.Fatal(err)
	}
	_ = previousSmoke.Close()
	if _, err := restored.ExecContext(ctx, `ALTER TABLE catalog_artifacts DISABLE TRIGGER USER`); err != nil {
		t.Fatal(err)
	}
	if _, err := restored.ExecContext(ctx, `UPDATE catalog_artifacts SET bytes='severed' WHERE constants_hash=$1 AND artifact_name=(SELECT min(artifact_name) FROM catalog_artifacts WHERE constants_hash=$1)`, previous.CurrentHash); err != nil {
		t.Fatal(err)
	}
	if _, err := restored.ExecContext(ctx, `ALTER TABLE catalog_artifacts ENABLE TRIGGER USER`); err != nil {
		t.Fatal(err)
	}
	if _, err := deploymentbackup.InspectPostgres(ctx, deploymentbackup.PostgresInspectionInput{
		DatabaseURLFile: writeReleaseFile(t, "severed-restore-url", restoreURL, 0o600), ReleaseManifest: backup.manifestPath, ContentRoot: root, RequireIdentity: true,
	}); err == nil {
		t.Fatal("severed catalog artifact accepted by release database inspection")
	}
	previousDrain, cancelPrevious := context.WithTimeout(ctx, 10*time.Second)
	defer cancelPrevious()
	if err := previous.Server.Drain(previousDrain, time.Now().UTC()); err != nil {
		t.Fatal(err)
	}
	shutdownServer(t, previousHTTP)
}

type releaseBackupFixture struct {
	path, manifestSHA256, manifestPath, identityPath, databaseURLFile string
}

func createReleaseBackup(t *testing.T, database *sql.DB, databaseURL, root string) releaseBackupFixture {
	t.Helper()
	var migration int
	if err := database.QueryRow(`SELECT max(version_id) FILTER (WHERE is_applied) FROM goose_db_version`).Scan(&migration); err != nil {
		t.Fatal(err)
	}
	epochPath := filepath.Join(root, filepath.FromSlash(epochseed.Path))
	epochBytes, err := os.ReadFile(epochPath)
	if err != nil {
		t.Fatal(err)
	}
	bundle, err := epochseed.Load(root)
	if err != nil {
		t.Fatal(err)
	}
	manifest := releaseBackupManifest(t, epochBytes, bundle, migration)
	manifestPath := writeReleaseBytes(t, "release-manifest.json", manifest, 0o600)
	identity, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatal(err)
	}
	identityPath := writeReleaseFile(t, "age-identity", identity.String(), 0o600)
	directory := t.TempDir()
	databaseURLFile := writeReleaseFile(t, "source-url", databaseURL, 0o600)
	header, path, err := deploymentbackup.CreatePostgresBackup(context.Background(), deploymentbackup.PostgresBackupInput{
		Directory: directory, BackupID: "20260822T180000Z-acde00000001", ServerID: "018f0000-0000-4000-8000-000000000901",
		StartedAt: time.Now().UTC().Add(-time.Second), Now: time.Now, Recipient: identity.Recipient().String(),
		DatabaseURLFile: databaseURLFile, ReleaseManifest: manifestPath,
		EpochDeclaration: epochPath, PreUpgrade: true,
	})
	if err != nil || !header.PreUpgrade {
		t.Fatalf("pre-upgrade backup header=%+v err=%v", header, err)
	}
	return releaseBackupFixture{path: path, manifestSHA256: digestBytes(manifest), manifestPath: manifestPath,
		identityPath: identityPath, databaseURLFile: databaseURLFile}
}

func releaseBackupManifest(t *testing.T, epoch []byte, bundle epochseed.Bundle, migration int) []byte {
	t.Helper()
	defaultHash := "sha256:" + strings.Repeat("d", 64)
	hashes := map[string]string{
		".env.example": defaultHash, "Caddyfile": defaultHash, "Dockerfile.gameserver": defaultHash,
		"LICENSE": defaultHash, "compose.yml": defaultHash, "compose.rotation.yml": defaultHash,
		"config.schema.json": defaultHash, "release-manifest.schema.json": defaultHash, "rehearsal-evidence.schema.json": defaultHash,
		"images/gameserver.docker.tar": defaultHash, "sbom/application.spdx.json": defaultHash,
		"third-party-licenses.txt": defaultHash, "site/index.html": defaultHash,
		"site/third-party-licenses.txt": defaultHash, "gameserver": defaultHash,
		"deployment-backup": defaultHash, "deployment-release": defaultHash, "deployment-operations": defaultHash, "deployment-rehearsal": defaultHash,
		"operations/prometheus.yml": defaultHash, "operations/cloud-clicker-alerts.yml": defaultHash,
		"operations/cloud-clicker-alerts.test.yml": defaultHash, "operations/alertmanager.example.yml": defaultHash,
		"operations/journald.template.conf": defaultHash, "operations/cloud-clicker-observe.service": defaultHash,
		"operations/cloud-clicker-observe.timer": defaultHash, "operations/operations.env.example": defaultHash,
		"content/balance/epochs/phase0.json": digestBytes(epoch),
		"sbom/alertmanager.spdx.json":        "sha256:" + strings.Repeat("1", 64), "sbom/caddy.spdx.json": "sha256:" + strings.Repeat("2", 64),
		"sbom/gameserver.spdx.json": "sha256:" + strings.Repeat("3", 64), "sbom/node-exporter.spdx.json": "sha256:" + strings.Repeat("4", 64),
		"sbom/postgres.spdx.json": "sha256:" + strings.Repeat("5", 64), "sbom/prometheus.spdx.json": "sha256:" + strings.Repeat("6", 64),
	}
	paths := make([]string, 0, len(hashes))
	for path := range hashes {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	artifacts := make([]releasepackage.File, 0, len(paths))
	for _, path := range paths {
		artifacts = append(artifacts, releasepackage.File{Path: path, SHA256: hashes[path]})
	}
	gameConfig := "sha256:" + strings.Repeat("b", 64)
	manifest := releasepackage.ReleaseManifest{SchemaVersion: 1, ReleaseVersion: "0.1.0", SourceCommit: strings.Repeat("a", 40), Platform: "linux/amd64",
		DockerEngineVersion: "28.4.0", DockerComposeVersion: "2.39.4", DatabaseMigration: migration,
		CompanySaveVersion: save.LatestCompanyVersion, FounderSaveVersion: save.LatestFounderVersion,
		EpochID: bundle.Seed.CurrentEpochID, ConstantsHash: bundle.Hash, CopyHash: defaultHash,
		Images: []releasepackage.Image{
			{Name: "alertmanager", Reference: "alertmanager:v1@sha256:" + strings.Repeat("a", 64), RuntimeConfigSHA256: "sha256:" + strings.Repeat("1", 64), SBOMPath: "sbom/alertmanager.spdx.json", SBOMSHA256: hashes["sbom/alertmanager.spdx.json"]},
			{Name: "caddy", Reference: "caddy:2@sha256:" + strings.Repeat("b", 64), RuntimeConfigSHA256: "sha256:" + strings.Repeat("2", 64), SBOMPath: "sbom/caddy.spdx.json", SBOMSHA256: hashes["sbom/caddy.spdx.json"]},
			{Name: "gameserver", Reference: gameConfig, RuntimeConfigSHA256: gameConfig, SBOMPath: "sbom/gameserver.spdx.json", SBOMSHA256: hashes["sbom/gameserver.spdx.json"]},
			{Name: "node-exporter", Reference: "node-exporter:v1@sha256:" + strings.Repeat("d", 64), RuntimeConfigSHA256: "sha256:" + strings.Repeat("4", 64), SBOMPath: "sbom/node-exporter.spdx.json", SBOMSHA256: hashes["sbom/node-exporter.spdx.json"]},
			{Name: "postgres", Reference: "postgres:16@sha256:" + strings.Repeat("e", 64), RuntimeConfigSHA256: "sha256:" + strings.Repeat("5", 64), SBOMPath: "sbom/postgres.spdx.json", SBOMSHA256: hashes["sbom/postgres.spdx.json"]},
			{Name: "prometheus", Reference: "prometheus:v1@sha256:" + strings.Repeat("f", 64), RuntimeConfigSHA256: "sha256:" + strings.Repeat("6", 64), SBOMPath: "sbom/prometheus.spdx.json", SBOMSHA256: hashes["sbom/prometheus.spdx.json"]},
		}, Artifacts: artifacts}
	if err := releasepackage.ValidateReleaseManifest(manifest); err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	return append(data, '\n')
}

func resetReleaseDatabase(t *testing.T, adminURL, name string) {
	t.Helper()
	if name != "cloud_clicker_release_restore" {
		t.Fatal("invalid restore database")
	}
	admin, err := save.OpenPostgres(context.Background(), adminURL)
	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close()
	_, _ = admin.Exec(`SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname=$1 AND pid <> pg_backend_pid()`, name)
	if _, err := admin.Exec(`DROP DATABASE IF EXISTS ` + name); err != nil {
		t.Fatal(err)
	}
	if _, err := admin.Exec(`CREATE DATABASE ` + name + ` OWNER cloud_clicker`); err != nil {
		t.Fatal(err)
	}
}

func writeReleaseFile(t *testing.T, name, value string, mode os.FileMode) string {
	return writeReleaseBytes(t, name, []byte(value+"\n"), mode)
}
func writeReleaseBytes(t *testing.T, name string, value []byte, mode os.FileMode) string {
	t.Helper()
	directory := t.TempDir()
	path := filepath.Join(directory, name)
	if err := os.WriteFile(path, value, mode); err != nil {
		t.Fatal(err)
	}
	return path
}

func digestBytes(value []byte) string {
	digest := sha256.Sum256(value)
	return "sha256:" + hex.EncodeToString(digest[:])
}

func startReleaseServer(t *testing.T, ctx context.Context, database *sql.DB, root, origin string, keys account.SigningKeys, bootstrap account.BootstrapReceiptKeys) (*gameserver.Composition, *http.Server) {
	t.Helper()
	composition, err := gameserver.Compose(ctx, gameserver.CompositionConfig{DB: database, RepositoryRoot: root,
		ServerID: "018f0000-0000-4000-8000-000000000901", ActivityBracket: "activity.standard",
		PublicOrigin: origin, TrustedProxyHops: 1, SigningKeys: keys, BootstrapKeys: bootstrap})
	if err != nil {
		t.Fatal(err)
	}
	serverContext, cancel := context.WithCancel(ctx)
	t.Cleanup(cancel)
	if err := composition.Server.Start(serverContext); err != nil {
		t.Fatal(err)
	}
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		t.Fatal(err)
	}
	httpServer := &http.Server{Handler: composition.Server.Handler(), ReadHeaderTimeout: 5 * time.Second}
	go func() { _ = httpServer.Serve(listener) }()
	return composition, httpServer
}

func shutdownServer(t *testing.T, server *http.Server) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		t.Fatal(err)
	}
}

func waitCaddy(t *testing.T, client *http.Client, endpoint string, status int) {
	t.Helper()
	deadline := time.Now().Add(20 * time.Second)
	for time.Now().Before(deadline) {
		response, err := client.Get(endpoint)
		if err == nil {
			_ = response.Body.Close()
			if response.StatusCode == status {
				return
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Fatalf("Caddy endpoint %s did not reach %d", endpoint, status)
}

func waitForDatabaseLock(t *testing.T, database *sql.DB) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		var waiting bool
		err := database.QueryRow(`SELECT EXISTS(SELECT 1 FROM pg_stat_activity WHERE datname=current_database() AND pid <> pg_backend_pid() AND wait_event_type='Lock')`).Scan(&waiting)
		if err == nil && waiting {
			return
		}
		time.Sleep(25 * time.Millisecond)
	}
	t.Fatal("admitted intent did not reach the locked save row")
}

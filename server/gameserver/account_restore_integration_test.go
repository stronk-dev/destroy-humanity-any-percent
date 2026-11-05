package gameserver

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"cloud-clicker/server/account"
	"cloud-clicker/server/deploymentbackup"
	"cloud-clicker/server/epochseed"
	"cloud-clicker/server/internal/backupfixture"
	"cloud-clicker/server/releasepackage"
	"cloud-clicker/server/save"
	"filippo.io/age"
)

// RP-125 observation, not a requirement that restoring deleted credentials is
// desirable. Unlike the Account constructor seam, this runs actual Compose,
// Founder initialization, Guild deletion participation and protected bootstrap.
func TestComposedBootstrapDeletionRestoreIntegration(t *testing.T) {
	if os.Getenv("TEST_ACCOUNT_RESTORE") != "1" {
		t.Skip("manual composed restore research is not enabled")
	}
	sourceURL, targetURL, adminURL := os.Getenv("TEST_DATABASE_URL"), os.Getenv("TEST_RESTORE_DATABASE_URL"), os.Getenv("TEST_ADMIN_DATABASE_URL")
	if sourceURL == "" || targetURL == "" || adminURL == "" {
		t.Fatal("composed restore requires all declared backup Postgres URLs")
	}
	root := os.Getenv("TEST_REPOSITORY_ROOT")
	if root == "" {
		t.Fatal("composed restore content root missing")
	}
	ctx := t.Context()
	now := time.Now().UTC().Truncate(time.Second)
	clock := func() time.Time { return now }
	seed, err := epochseed.Load(root)
	if err != nil {
		t.Fatal(err)
	}
	resetAccountRestoreDatabase(t, adminURL, "cloud_clicker_source")
	source, err := save.OpenPostgres(ctx, sourceURL)
	if err != nil {
		t.Fatal("source database unavailable")
	}
	t.Cleanup(func() { source.Close() })
	api, stopSource := startAccountRestoreServer(t, source, root, clock)
	request := accountRestoreRequest
	create := func(key string) accountRestorePlayer {
		t.Helper()
		status, body := request(t, api, http.MethodPost, "/api/v1/bootstrap", "", fmt.Sprintf(`{"idempotency_key":%q}`, key))
		var initial bootstrapResponseEnvelope
		if status != http.StatusCreated || json.Unmarshal(body, &initial) != nil || initial.Account.AccountID == "" || initial.Account.RecoveryCode == "" || initial.Session.AccessToken == "" || initial.Session.RefreshToken == "" {
			t.Fatalf("composed bootstrap producer failed: status=%d", status)
		}
		status, founder := accountRestoreProfile(t, api, initial.Session.AccessToken)
		if status != http.StatusOK {
			t.Fatal("bootstrap credential cannot reach its Founder")
		}
		status, state := request(t, api, http.MethodGet, "/api/v1/founder/state", initial.Session.AccessToken, "")
		var snapshot struct {
			ConstantsHash string `json:"constants_hash"`
		}
		if status != http.StatusOK || json.Unmarshal(state, &snapshot) != nil || snapshot.ConstantsHash != seed.Hash {
			t.Fatalf("composed initial Company snapshot failed: status=%d", status)
		}
		return accountRestorePlayer{bootstrap: initial, raw: body, key: key, founder: founder}
	}
	subject, bystander := create(strings.Repeat("ab", 32)), create(strings.Repeat("cd", 32))
	var migration int
	if err := source.QueryRowContext(ctx, `SELECT max(version_id) FILTER (WHERE is_applied) FROM goose_db_version`).Scan(&migration); err != nil {
		t.Fatal(err)
	}
	// Only backup metadata is synthetic; the epoch bytes/content hash and live
	// migration endpoint are the actual repository composition's coordinates.
	epoch, err := os.ReadFile(filepath.Join(root, epochseed.Path))
	if err != nil {
		t.Fatal(err)
	}
	var manifest releasepackage.ReleaseManifest
	if err := json.Unmarshal(backupfixture.Manifest(t, epoch, migration), &manifest); err != nil {
		t.Fatal(err)
	}
	manifest.ConstantsHash, manifest.EpochID = seed.Hash, seed.Seed.CurrentEpochID
	manifestBytes, err := json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	workspace := t.TempDir()
	writePrivate := func(name, value string) string {
		t.Helper()
		path := filepath.Join(workspace, name)
		if err := os.WriteFile(path, []byte(value), 0o600); err != nil {
			t.Fatal(err)
		}
		return path
	}
	identity, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatal(err)
	}
	manifestPath := writePrivate("manifest.json", string(manifestBytes))
	epochPath := writePrivate("epoch.json", string(epoch))
	sourceSecret, targetSecret := writePrivate("source-url", sourceURL), writePrivate("target-url", targetURL)
	identityPath := writePrivate("identity", identity.String())
	backup := func(suffix string) string {
		t.Helper()
		_, path, err := deploymentbackup.CreatePostgresBackup(ctx, deploymentbackup.PostgresBackupInput{
			Directory: t.TempDir(), BackupID: now.Format("20060102T150405Z") + "-" + suffix,
			ServerID: "018f0000-0000-4000-8000-000000000306", StartedAt: now.Add(-time.Minute), Now: clock,
			Recipient: identity.Recipient().String(), DatabaseURLFile: sourceSecret, ReleaseManifest: manifestPath, EpochDeclaration: epochPath,
		})
		if err != nil {
			t.Fatal(err)
		}
		return path
	}
	preRows := accountRestoreCredentialRows(t, source)
	pre := backup("000000000011")
	if status, _ := request(t, api, http.MethodDelete, "/api/v1/account", subject.bootstrap.Session.AccessToken, ""); status != http.StatusNoContent {
		t.Fatalf("composed account deletion failed: status=%d", status)
	}
	observeAccountRestoreCredentials(t, "deleted-source", source, api, subject, bystander, true, 0)
	postRows := accountRestoreCredentialRows(t, source)
	post := backup("000000000012")
	stopSource()
	for _, arm := range []struct {
		name, path string
		rows       []byte
		absent     bool
	}{{"pre-deletion-restore", pre, preRows, false}, {"post-deletion-restore", post, postRows, true}} {
		t.Run(arm.name, func(t *testing.T) {
			for _, lifetime := range []struct {
				name string
				age  time.Duration
			}{{"current", 0}, {"access-expired", 16 * time.Minute}, {"refresh-expired", 31 * 24 * time.Hour}} {
				t.Run(lifetime.name, func(t *testing.T) {
					resetAccountRestoreDatabase(t, adminURL, "cloud_clicker_restore")
					if _, err := deploymentbackup.RestorePostgresBackup(t.Context(), deploymentbackup.PostgresRestoreInput{
						BackupPath: arm.path, ExpectedManifestSHA256: fmt.Sprintf("sha256:%x", sha256.Sum256(manifestBytes)),
						IdentityFile: identityPath, TargetDatabaseURLFile: targetSecret,
					}); err != nil {
						t.Fatal(err)
					}
					target, err := save.OpenPostgres(t.Context(), targetURL)
					if err != nil {
						t.Fatal("restore database unavailable")
					}
					t.Cleanup(func() { target.Close() })
					if !bytes.Equal(arm.rows, accountRestoreCredentialRows(t, target)) {
						t.Fatal("restored six-family credential population differs from selected backup")
					}
					restoredClock := func() time.Time { return now.Add(lifetime.age) }
					restored, _ := startAccountRestoreServer(t, target, root, restoredClock)
					observeAccountRestoreCredentials(t, arm.name, target, restored, subject, bystander, arm.absent, lifetime.age)
				})
			}
		})
	}
	t.Logf("RP-125 composed coordinate: epoch=%d constants_hash=%s migration=%d; fixture manifest, unchanged keys, application clock offsets 0/16min/31days", seed.Seed.CurrentEpochID, seed.Hash, migration)
}

type accountRestorePlayer struct {
	bootstrap bootstrapResponseEnvelope
	raw       []byte
	key       string
	founder   string
}

func startAccountRestoreServer(t *testing.T, database *sql.DB, root string, clock func() time.Time) (*httptest.Server, func()) {
	t.Helper()
	composition, err := Compose(t.Context(), CompositionConfig{
		DB: database, RepositoryRoot: root, ServerID: "018f0000-0000-4000-8000-000000000306", ActivityBracket: "activity.standard",
		Clock: clock, PublicCursorKeys: testPublicCursorKeys(),
		SigningKeys:   account.SigningKeys{CurrentID: "restore-fixture", Current: bytes.Repeat([]byte{0x73}, 32)},
		BootstrapKeys: account.BootstrapReceiptKeys{CurrentID: "restore-fixture", Current: bytes.Repeat([]byte{0x74}, 32)},
	})
	if err != nil {
		t.Fatal(err)
	}
	serverContext, cancel := context.WithCancel(t.Context())
	if err := composition.Server.Start(serverContext); err != nil {
		cancel()
		t.Fatal(err)
	}
	api := httptest.NewServer(composition.Server.Handler())
	var once sync.Once
	stop := func() {
		once.Do(func() {
			defer cancel()
			defer api.Close()
			drain, stopDrain := context.WithTimeout(context.Background(), 5*time.Second)
			defer stopDrain()
			if err := composition.Server.Drain(drain, clock()); err != nil {
				t.Error(err)
			}
		})
	}
	t.Cleanup(stop)
	if status, _ := accountRestoreRequest(t, api, http.MethodGet, "/readyz", "", ""); status != http.StatusNoContent {
		t.Fatalf("composed restore service not ready: status=%d", status)
	}
	return api, stop
}

func accountRestoreRequest(t *testing.T, api *httptest.Server, method, path, token, body string) (int, []byte) {
	t.Helper()
	request, err := http.NewRequestWithContext(t.Context(), method, api.URL+path, strings.NewReader(body))
	if err != nil {
		t.Fatal("credential request construction failed")
	}
	request.Header.Set("Content-Type", "application/json")
	if token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}
	response, err := api.Client().Do(request)
	if err != nil {
		t.Fatal("credential observation transport failed")
	}
	defer response.Body.Close()
	data, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatal("credential observation body incomplete")
	}
	return response.StatusCode, data
}

func accountRestoreProfile(t *testing.T, api *httptest.Server, token string) (int, string) {
	t.Helper()
	status, body := accountRestoreRequest(t, api, http.MethodGet, "/api/v1/founder", token, "")
	var founder account.Founder
	if status == http.StatusOK && (json.Unmarshal(body, &founder) != nil || founder.ID == "") {
		t.Fatal("successful profile has no valid Founder")
	}
	return status, founder.ID
}

func observeAccountRestoreCredentials(t *testing.T, arm string, database *sql.DB, api *httptest.Server, subject, bystander accountRestorePlayer, absent bool, age time.Duration) {
	t.Helper()
	bystanderToken := bystander.bootstrap.Session.AccessToken
	if age > 0 {
		// Original bystander tokens expire too. A fresh recovery login proves
		// the restored composition is healthy at the advanced application time.
		status, body := accountRestoreRequest(t, api, http.MethodPost, "/api/v1/session", "", fmt.Sprintf(`{"account_id":%q,"recovery_code":%q}`, bystander.bootstrap.Account.AccountID, bystander.bootstrap.Account.RecoveryCode))
		var pair account.TokenPair
		if status != http.StatusOK || json.Unmarshal(body, &pair) != nil || pair.AccessToken == "" || pair.RefreshToken == "" {
			t.Fatalf("fresh composed bystander session failed: status=%d", status)
		}
		bystanderToken = pair.AccessToken
	}
	if status, founder := accountRestoreProfile(t, api, bystanderToken); status != http.StatusOK || founder != bystander.founder {
		t.Fatal("composed bystander control is not healthy")
	}
	var count int
	if err := database.QueryRowContext(t.Context(), "SELECT count(*) FROM accounts WHERE account_id=$1", subject.bootstrap.Account.AccountID).Scan(&count); err != nil || (absent && count != 0) || (!absent && count != 1) {
		t.Fatal("composed source/restore account population is wrong")
	}
	// Replay first: rotation below consumes the original refresh token. With no
	// clock advance its original access must reach the Founder; an aged exact
	// receipt can legitimately contain expired access, checked separately below.
	bootstrapStatus, retry := accountRestoreRequest(t, api, http.MethodPost, "/api/v1/bootstrap", "", fmt.Sprintf(`{"idempotency_key":%q}`, subject.key))
	if bootstrapStatus == http.StatusCreated {
		if absent || !bytes.Equal(retry, subject.raw) {
			t.Fatal("restored bootstrap does not reproduce the original account receipt")
		}
		if age == 0 {
			if status, founder := accountRestoreProfile(t, api, subject.bootstrap.Session.AccessToken); status != http.StatusOK || founder != subject.founder {
				t.Fatal("replayed bootstrap credential cannot reach original Founder")
			}
		}
	} else if bootstrapStatus != http.StatusConflict || string(retry) != "{\"category\":\"conflict\",\"detail\":\"bootstrap_expired\"}\n" {
		t.Fatalf("invalid bootstrap retry observation: status=%d", bootstrapStatus)
	}
	accessStatus, founder := accountRestoreProfile(t, api, subject.bootstrap.Session.AccessToken)
	if accessStatus != http.StatusOK && accessStatus != http.StatusUnauthorized {
		t.Fatalf("invalid access observation: status=%d", accessStatus)
	}
	if age > 0 && accessStatus != http.StatusUnauthorized {
		t.Fatal("expired original access token admitted")
	}
	if accessStatus == http.StatusOK && (absent || founder != subject.founder) {
		t.Fatal("restored access reaches a different or absent Founder")
	}
	refreshStatus, refresh := accountRestoreRequest(t, api, http.MethodPost, "/api/v1/session/refresh", "", fmt.Sprintf(`{"refresh_token":%q}`, subject.bootstrap.Session.RefreshToken))
	recoveryStatus, recovery := accountRestoreRequest(t, api, http.MethodPost, "/api/v1/session", "", fmt.Sprintf(`{"account_id":%q,"recovery_code":%q}`, subject.bootstrap.Account.AccountID, subject.bootstrap.Account.RecoveryCode))
	if age >= 30*24*time.Hour && (refreshStatus != http.StatusUnauthorized || bootstrapStatus != http.StatusConflict) {
		t.Fatal("expired original refresh token or bootstrap receipt admitted")
	}
	for _, result := range []struct {
		status int
		body   []byte
	}{{refreshStatus, refresh}, {recoveryStatus, recovery}} {
		if result.status != http.StatusOK && result.status != http.StatusUnauthorized {
			t.Fatalf("invalid credential observation: status=%d", result.status)
		}
		if result.status == http.StatusOK {
			var pair account.TokenPair
			if absent || json.Unmarshal(result.body, &pair) != nil || pair.AccessToken == "" || pair.RefreshToken == "" {
				t.Fatal("invalid restored credential pair")
			}
			if status, founder := accountRestoreProfile(t, api, pair.AccessToken); status != http.StatusOK || founder != subject.founder {
				t.Fatal("restored credential does not bind original Founder")
			}
		}
	}
	wrongCode := "a" + subject.bootstrap.Account.RecoveryCode[1:]
	if wrongCode == subject.bootstrap.Account.RecoveryCode {
		wrongCode = "b" + subject.bootstrap.Account.RecoveryCode[1:]
	}
	if status, _ := accountRestoreRequest(t, api, http.MethodPost, "/api/v1/session", "", fmt.Sprintf(`{"account_id":%q,"recovery_code":%q}`, subject.bootstrap.Account.AccountID, wrongCode)); status != http.StatusUnauthorized {
		t.Fatal("wrong recovery code admitted")
	}
	if status, _ := accountRestoreProfile(t, api, "invalid-token"); status != http.StatusUnauthorized {
		t.Fatal("invalid access token admitted")
	}
	if status, _ := accountRestoreRequest(t, api, http.MethodPost, "/api/v1/bootstrap", "", `{"idempotency_key":"guessable"}`); status != http.StatusBadRequest {
		t.Fatal("malformed bootstrap key admitted")
	}
	if absent && (accessStatus != http.StatusUnauthorized || refreshStatus != http.StatusUnauthorized || recoveryStatus != http.StatusUnauthorized || bootstrapStatus != http.StatusConflict) {
		t.Fatal("absent account authorizes an original credential")
	}
	var after int
	wantAccounts := 2
	if absent {
		wantAccounts--
	}
	if err := database.QueryRowContext(t.Context(), "SELECT count(*) FROM accounts").Scan(&after); err != nil || after != wantAccounts {
		t.Fatal("credential observations unexpectedly created an account")
	}
	t.Logf("RP-125 composed arm=%s age_minutes=%d account_rows=%d access=%d refresh=%d recovery=%d bootstrap_retry=%d bystander=200", arm, int64(age/time.Minute), count, accessStatus, refreshStatus, recoveryStatus, bootstrapStatus)
}

func accountRestoreCredentialRows(t *testing.T, database *sql.DB) []byte {
	t.Helper()
	var rows bytes.Buffer
	for _, table := range []string{"accounts", "account_founders", "session_families", "sessions", "access_tokens", "bootstrap_receipts"} {
		var encoded []byte
		if err := database.QueryRowContext(t.Context(), `SELECT COALESCE(jsonb_agg(to_jsonb(r) ORDER BY to_jsonb(r)::text),'[]'::jsonb) FROM `+table+` r`).Scan(&encoded); err != nil {
			t.Fatal(err)
		}
		fmt.Fprintf(&rows, "%s\n%s\n", table, encoded)
	}
	return rows.Bytes()
}

func resetAccountRestoreDatabase(t *testing.T, adminURL, name string) {
	t.Helper()
	if name != "cloud_clicker_source" && name != "cloud_clicker_restore" {
		t.Fatal("invalid composed restore test database name")
	}
	database, err := save.OpenPostgres(t.Context(), adminURL)
	if err != nil {
		t.Fatal("test database administrator unavailable")
	}
	defer database.Close()
	if _, err := database.ExecContext(t.Context(), `SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname=$1 AND pid <> pg_backend_pid()`, name); err != nil {
		t.Fatal(err)
	}
	if _, err := database.ExecContext(t.Context(), `DROP DATABASE IF EXISTS `+name); err != nil {
		t.Fatal(err)
	}
	if _, err := database.ExecContext(t.Context(), `CREATE DATABASE `+name+` OWNER cloud_clicker`); err != nil {
		t.Fatal(err)
	}
}

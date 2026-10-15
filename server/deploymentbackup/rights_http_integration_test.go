package deploymentbackup

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"cloud-clicker/server/account"
	"cloud-clicker/server/production"
	"cloud-clicker/server/save"
	"filippo.io/age"
)

// The credential/deletion routes below use the real Account repository. No
// gameplay request may obtain a successful result from this constructor seam.
type rightsRestoreNoGameplay struct{}

func (rightsRestoreNoGameplay) Handle(context.Context, string, production.EvaluationMode, time.Time, []byte) (production.HandleResult, error) {
	return production.HandleResult{}, errors.New("gameplay is outside the credential restore fixture")
}

func TestHTTPDeletionBackupRestoresCredentialAccessIntegration(t *testing.T) {
	sourceURL, targetURL, adminURL := os.Getenv("TEST_DATABASE_URL"), os.Getenv("TEST_RESTORE_DATABASE_URL"), os.Getenv("TEST_ADMIN_DATABASE_URL")
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
	seedBackupEpoch(t, source)
	var migration int
	if err := source.QueryRowContext(ctx, `SELECT max(version_id) FILTER (WHERE is_applied) FROM goose_db_version`).Scan(&migration); err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 8, 22, 18, 0, 0, 0, time.UTC)
	clock := func() time.Time { return now }
	keys := account.SigningKeys{CurrentID: "rights-restore", Current: bytes.Repeat([]byte{0x31}, 32)}
	newAPI := func(database *sql.DB) *httptest.Server {
		t.Helper()
		repository, err := account.NewRepository(database, rightsRestoreCatalog{}, "sha256:"+strings.Repeat("a", 64), keys, clock, nil)
		if err != nil {
			t.Fatal(err)
		}
		config := account.Phase0APIConfig(account.BootstrapReceiptKeys{CurrentID: "rights-restore", Current: bytes.Repeat([]byte{0x32}, 32)})
		api, err := account.NewAPI(repository, rightsRestoreNoGameplay{}, config)
		if err != nil {
			t.Fatal(err)
		}
		return httptest.NewServer(api.Router())
	}
	request := func(server *httptest.Server, method, path, token, body string) (int, []byte) {
		t.Helper()
		r, err := http.NewRequest(method, server.URL+path, strings.NewReader(body))
		if err != nil {
			t.Fatal(err)
		}
		r.Header.Set("Content-Type", "application/json")
		if token != "" {
			r.Header.Set("Authorization", "Bearer "+token)
		}
		response, err := server.Client().Do(r)
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
	profile := func(server *httptest.Server, token string) (int, string) {
		t.Helper()
		status, body := request(server, http.MethodGet, "/api/v1/founder", token, "")
		var value account.Founder
		if status == 200 && (json.Unmarshal(body, &value) != nil || value.ID == "") {
			t.Fatal("invalid successful Founder observation")
		}
		return status, value.ID
	}
	type player struct {
		account account.CreatedAccount
		tokens  account.TokenPair
		founder string
	}
	create := func(server *httptest.Server) player {
		t.Helper()
		var p player
		status, body := request(server, http.MethodPost, "/api/v1/account", "", `{}`)
		if status != 201 || json.Unmarshal(body, &p.account) != nil || p.account.AccountID == "" || p.account.RecoveryCode == "" {
			t.Fatal("HTTP account producer failed")
		}
		status, body = request(server, http.MethodPost, "/api/v1/session", "", fmt.Sprintf(`{"account_id":%q,"recovery_code":%q}`, p.account.AccountID, p.account.RecoveryCode))
		if status != 200 || json.Unmarshal(body, &p.tokens) != nil || p.tokens.AccessToken == "" || p.tokens.RefreshToken == "" {
			t.Fatal("HTTP session producer failed")
		}
		status, p.founder = profile(server, p.tokens.AccessToken)
		if status != 200 {
			t.Fatal("new HTTP credential cannot reach its Founder")
		}
		return p
	}
	server := newAPI(source)
	defer server.Close()
	subject, bystander := create(server), create(server)
	workspace := t.TempDir()
	identity, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatal(err)
	}
	epoch := fixtureEpoch(t, 8)
	manifest := fixtureReleaseManifestForMigration(t, epoch, migration)
	manifestPath := writeRegular(t, workspace, "manifest.json", manifest)
	epochPath := writeRegular(t, workspace, "epoch.json", epoch)
	sourceSecret := writeSecret(t, workspace, "source-url", sourceURL)
	targetSecret := writeSecret(t, workspace, "target-url", targetURL)
	identityPath := writeSecret(t, workspace, "identity", identity.String())
	backup := func(id string) string {
		t.Helper()
		_, path, err := CreatePostgresBackup(ctx, PostgresBackupInput{Directory: t.TempDir(), BackupID: id,
			ServerID: "01986666-b001-4000-8000-000000000001", StartedAt: now.Add(-time.Minute), Now: clock,
			Recipient: identity.Recipient().String(), DatabaseURLFile: sourceSecret, ReleaseManifest: manifestPath, EpochDeclaration: epochPath})
		if err != nil {
			t.Fatal(err)
		}
		return path
	}
	observe := func(name string, database *sql.DB, api *httptest.Server) {
		t.Helper()
		if status, founder := profile(api, bystander.tokens.AccessToken); status != 200 || founder != bystander.founder {
			t.Fatal("restored bystander control is not healthy")
		}
		var count int
		if err := database.QueryRowContext(ctx, "SELECT count(*) FROM accounts WHERE account_id=$1", subject.account.AccountID).Scan(&count); err != nil {
			t.Fatal(err)
		}
		accessStatus, founder := profile(api, subject.tokens.AccessToken)
		if accessStatus != 200 && accessStatus != 401 {
			t.Fatalf("invalid access observation status=%d", accessStatus)
		}
		if accessStatus == 200 && (count != 1 || founder != subject.founder) {
			t.Fatal("restored access reaches a different or absent account")
		}
		refreshStatus, refreshBody := request(api, http.MethodPost, "/api/v1/session/refresh", "", fmt.Sprintf(`{"refresh_token":%q}`, subject.tokens.RefreshToken))
		recoveryStatus, recoveryBody := request(api, http.MethodPost, "/api/v1/session", "", fmt.Sprintf(`{"account_id":%q,"recovery_code":%q}`, subject.account.AccountID, subject.account.RecoveryCode))
		for _, result := range []struct {
			status int
			body   []byte
		}{{refreshStatus, refreshBody}, {recoveryStatus, recoveryBody}} {
			if result.status != 200 && result.status != 401 {
				t.Fatalf("invalid credential observation status=%d", result.status)
			}
			if result.status == 200 {
				var pair account.TokenPair
				if json.Unmarshal(result.body, &pair) != nil || pair.AccessToken == "" || pair.RefreshToken == "" {
					t.Fatal("invalid revived credential pair")
				}
				if status, founder := profile(api, pair.AccessToken); count != 1 || status != 200 || founder != subject.founder {
					t.Fatal("revived credential does not bind the original Founder")
				}
			}
		}
		wrongCode := "a" + subject.account.RecoveryCode[1:]
		if wrongCode == subject.account.RecoveryCode {
			wrongCode = "b" + subject.account.RecoveryCode[1:]
		}
		if status, _ := request(api, http.MethodPost, "/api/v1/session", "", fmt.Sprintf(`{"account_id":%q,"recovery_code":%q}`, subject.account.AccountID, wrongCode)); status != 401 {
			t.Fatal("wrong recovery code was admitted")
		}
		if status, _ := profile(api, "invalid-token"); status != 401 {
			t.Fatal("invalid access token was admitted")
		}
		if count == 0 && (accessStatus != 401 || refreshStatus != 401 || recoveryStatus != 401) {
			t.Fatal("absent account still authorizes old credentials")
		}
		t.Logf("RP-125 HTTP arm=%s account_rows=%d access=%d refresh=%d recovery=%d bystander=200", name, count, accessStatus, refreshStatus, recoveryStatus)
	}
	credentialRows := func(database *sql.DB) []byte {
		t.Helper()
		var rows bytes.Buffer
		for _, table := range []string{"accounts", "account_founders", "session_families", "sessions", "access_tokens"} {
			var encoded []byte
			if err := database.QueryRowContext(ctx, `SELECT COALESCE(jsonb_agg(to_jsonb(r) ORDER BY to_jsonb(r)::text),'[]'::jsonb) FROM `+table+` r`).Scan(&encoded); err != nil {
				t.Fatal(err)
			}
			rows.WriteString(table)
			rows.WriteByte('\n')
			rows.Write(encoded)
			rows.WriteByte('\n')
		}
		return rows.Bytes()
	}
	preRows := credentialRows(source)
	pre := backup("20260822T175900Z-000000000006")
	if status, _ := request(server, http.MethodDelete, "/api/v1/account", subject.tokens.AccessToken, ""); status != 204 {
		t.Fatalf("HTTP deletion status=%d", status)
	}
	var remaining int
	if err := source.QueryRowContext(ctx, "SELECT count(*) FROM accounts WHERE account_id=$1", subject.account.AccountID).Scan(&remaining); err != nil || remaining != 0 {
		t.Fatal("HTTP deletion left the account present")
	}
	observe("deleted-source", source, server)
	postRows := credentialRows(source)
	post := backup("20260822T175900Z-000000000007")
	for _, arm := range []struct {
		name, path string
		rows       []byte
	}{{"pre-deletion-restore", pre, preRows}, {"post-deletion-restore", post, postRows}} {
		t.Run(arm.name, func(t *testing.T) {
			resetDatabase(t, adminURL, "cloud_clicker_restore")
			if _, err := RestorePostgresBackup(ctx, PostgresRestoreInput{BackupPath: arm.path, ExpectedManifestSHA256: digest(manifest), IdentityFile: identityPath, TargetDatabaseURLFile: targetSecret}); err != nil {
				t.Fatal(err)
			}
			target, err := save.OpenPostgres(ctx, targetURL)
			if err != nil {
				t.Fatal(err)
			}
			defer target.Close()
			if !bytes.Equal(arm.rows, credentialRows(target)) {
				t.Fatal("restored credential rows do not match the selected backup population")
			}
			restored := newAPI(target)
			defer restored.Close()
			observe(arm.name, target, restored)
		})
	}
}

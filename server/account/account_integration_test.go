package account

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"cloud-clicker/server/economy"
	"cloud-clicker/server/internal/testhttp"
	"cloud-clicker/server/production"
	"cloud-clicker/server/save"
)

type integrationCatalogs map[string]*economy.Catalog

func (catalogs integrationCatalogs) Resolve(hash string) (*economy.Catalog, bool) {
	catalog, ok := catalogs[hash]
	return catalog, ok
}

func TestAccountSessionIntegration(t *testing.T) {
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
	truncateAccountIntegration(t, db)

	catalogBytes, err := os.ReadFile("../../balance/catalogs/phase0.json")
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := economy.LoadCatalog(catalogBytes)
	if err != nil {
		t.Fatal(err)
	}
	hash := save.ConstantsHash(catalogBytes)
	seedAccountEpoch(t, db, hash, catalogBytes)
	resolver := integrationCatalogs{hash: catalog}
	keys := SigningKeys{CurrentID: "test-current", Current: bytes.Repeat([]byte{0x42}, 32), PreviousID: "test-previous", Previous: bytes.Repeat([]byte{0x24}, 32)}
	now := time.Date(2026, 7, 29, 12, 0, 0, 0, time.UTC)
	repository, err := NewRepository(db, resolver, hash, keys, func() time.Time { return now }, nil)
	if err != nil {
		t.Fatal(err)
	}
	saveStore, err := save.NewStore(db, resolver, nil)
	if err != nil {
		t.Fatal(err)
	}
	intentService, err := production.NewService(saveStore, resolver, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	api, err := NewAPI(repository, intentService, APIConfig{UnauthenticatedBurst: 20, UnauthenticatedPerMin: 60, AccountBurst: 100, AccountPerMin: 600, MaxBodyBytes: 64 << 10})
	if err != nil {
		t.Fatal(err)
	}
	server := testhttp.New(api.Router())
	defer server.Close()

	createdResponse := requestJSON(t, server.Client, http.MethodPost, server.URL+"/api/v1/account", "", `{}`)
	if createdResponse.StatusCode != http.StatusCreated {
		t.Fatalf("create status=%d body=%s", createdResponse.StatusCode, readBody(createdResponse))
	}
	var created CreatedAccount
	decodeResponse(t, createdResponse, &created)
	if created.AccountID == "" || created.RecoveryCode == "" || created.FounderID != "" {
		t.Fatalf("created=%+v", created)
	}
	var accountColumns int
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM information_schema.columns WHERE table_schema=current_schema() AND table_name='accounts'`).Scan(&accountColumns); err != nil || accountColumns != 3 {
		t.Fatalf("account columns=%d err=%v", accountColumns, err)
	}

	sessionResponse := requestJSON(t, server.Client, http.MethodPost, server.URL+"/api/v1/session", "", fmt.Sprintf(`{"account_id":%q,"recovery_code":%q}`, created.AccountID, created.RecoveryCode))
	if sessionResponse.StatusCode != http.StatusOK {
		t.Fatalf("session status=%d body=%s", sessionResponse.StatusCode, readBody(sessionResponse))
	}
	var firstPair TokenPair
	decodeResponse(t, sessionResponse, &firstPair)
	claims, err := repository.Authenticate(ctx, firstPair.AccessToken)
	if err != nil || claims.Subject != created.AccountID {
		t.Fatalf("claims=%+v err=%v", claims, err)
	}

	profileResponse := requestJSON(t, server.Client, http.MethodGet, server.URL+"/api/v1/founder", firstPair.AccessToken, "")
	if profileResponse.StatusCode != http.StatusOK {
		t.Fatalf("profile status=%d body=%s", profileResponse.StatusCode, readBody(profileResponse))
	}
	var profile struct {
		ID        string         `json:"id"`
		CreatedAt time.Time      `json:"created_at"`
		Display   map[string]any `json:"display"`
	}
	decodeResponse(t, profileResponse, &profile)
	if profile.ID != claims.FounderID || profile.Display == nil {
		t.Fatalf("profile=%+v claims=%+v", profile, claims)
	}
	oldState, err := repository.ActiveCompanyState(ctx, created.AccountID)
	if err != nil || oldState.Revision != 1 {
		t.Fatalf("initial state=%+v err=%v", oldState, err)
	}

	intent := `{"intent_id":"01985555-1111-7111-8111-111111111111","kind":"perform_manual_batch","expected_revision":1,"action_id":"manual.click","count":1,"window_ms":1}`
	intentResponse := requestJSON(t, server.Client, http.MethodPost, server.URL+"/api/v1/intents", firstPair.AccessToken, intent)
	if intentResponse.StatusCode != http.StatusOK {
		t.Fatalf("intent status=%d body=%s", intentResponse.StatusCode, readBody(intentResponse))
	}
	var receipt map[string]json.RawMessage
	decodeResponse(t, intentResponse, &receipt)
	if string(receipt["outcome"]) != `"applied"` || string(receipt["new_revision"]) != "2" {
		t.Fatalf("receipt=%v", receipt)
	}

	refreshResponse := requestJSON(t, server.Client, http.MethodPost, server.URL+"/api/v1/session/refresh", "", fmt.Sprintf(`{"refresh_token":%q}`, firstPair.RefreshToken))
	if refreshResponse.StatusCode != http.StatusOK {
		t.Fatalf("refresh status=%d body=%s", refreshResponse.StatusCode, readBody(refreshResponse))
	}
	var rotated TokenPair
	decodeResponse(t, refreshResponse, &rotated)
	reuseResponse := requestJSON(t, server.Client, http.MethodPost, server.URL+"/api/v1/session/refresh", "", fmt.Sprintf(`{"refresh_token":%q}`, firstPair.RefreshToken))
	if reuseResponse.StatusCode != http.StatusUnauthorized || !strings.Contains(readBody(reuseResponse), "refresh_reused") {
		t.Fatalf("reuse status=%d", reuseResponse.StatusCode)
	}
	if _, err := repository.Authenticate(ctx, rotated.AccessToken); err == nil {
		t.Fatal("refresh-family reuse did not revoke rotated access token")
	}

	freshSession := requestJSON(t, server.Client, http.MethodPost, server.URL+"/api/v1/session", "", fmt.Sprintf(`{"account_id":%q,"recovery_code":%q}`, created.AccountID, created.RecoveryCode))
	var freshPair TokenPair
	decodeResponse(t, freshSession, &freshPair)
	newFounderResponse := requestJSON(t, server.Client, http.MethodPost, server.URL+"/api/v1/founder", freshPair.AccessToken, `{}`)
	if newFounderResponse.StatusCode != http.StatusCreated {
		t.Fatalf("new founder status=%d body=%s", newFounderResponse.StatusCode, readBody(newFounderResponse))
	}
	var newFounder Founder
	decodeResponse(t, newFounderResponse, &newFounder)
	if newFounder.ID == profile.ID {
		t.Fatal("New Founder reused the archived founder identity")
	}
	loadedOld, err := saveStore.LoadLatest(ctx, oldState.StreamID)
	if err != nil || loadedOld.ArchivedAt == nil || loadedOld.Revision.Number != 2 {
		t.Fatalf("archived old stream=%+v err=%v", loadedOld, err)
	}

	newState, err := repository.ActiveCompanyState(ctx, created.AccountID)
	if err != nil || newState.Revision != 1 {
		t.Fatalf("new state=%+v err=%v", newState, err)
	}
	importBody, _ := json.Marshal(struct {
		Version       int             `json:"version"`
		ConstantsHash string          `json:"constants_hash"`
		State         json.RawMessage `json:"state"`
	}{Version: newState.Version, ConstantsHash: newState.ConstantsHash, State: newState.State})
	importResponse := requestJSON(t, server.Client, http.MethodPost, server.URL+"/api/v1/founder/import", freshPair.AccessToken, string(importBody))
	if importResponse.StatusCode != http.StatusOK {
		t.Fatalf("import status=%d body=%s", importResponse.StatusCode, readBody(importResponse))
	}
	var imported Founder
	decodeResponse(t, importResponse, &imported)
	if !imported.Imported {
		t.Fatal("imported flag did not round-trip")
	}

	deleteResponse := requestJSON(t, server.Client, http.MethodDelete, server.URL+"/api/v1/account", freshPair.AccessToken, "")
	if deleteResponse.StatusCode != http.StatusNoContent {
		t.Fatalf("delete status=%d body=%s", deleteResponse.StatusCode, readBody(deleteResponse))
	}
	var accounts, emails, sessions int
	if err := db.QueryRowContext(ctx, `SELECT (SELECT count(*) FROM accounts),(SELECT count(*) FROM account_emails),(SELECT count(*) FROM sessions)`).Scan(&accounts, &emails, &sessions); err != nil || accounts != 0 || emails != 0 || sessions != 0 {
		t.Fatalf("PII/session rows accounts=%d emails=%d sessions=%d err=%v", accounts, emails, sessions, err)
	}
	loadedImported, err := saveStore.LoadLatest(ctx, newState.StreamID)
	if err != nil || loadedImported.ArchivedAt == nil {
		t.Fatalf("anonymized save missing after account deletion: %+v err=%v", loadedImported, err)
	}
}

func TestAccountUnauthenticatedRateLimitIntegration(t *testing.T) {
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
	truncateAccountIntegration(t, db)
	catalogBytes, _ := os.ReadFile("../../balance/catalogs/phase0.json")
	catalog, _ := economy.LoadCatalog(catalogBytes)
	hash := save.ConstantsHash(catalogBytes)
	seedAccountEpoch(t, db, hash, catalogBytes)
	resolver := integrationCatalogs{hash: catalog}
	repository, err := NewRepository(db, resolver, hash, SigningKeys{CurrentID: "test", Current: bytes.Repeat([]byte{1}, 32)}, time.Now, nil)
	if err != nil {
		t.Fatal(err)
	}
	store, _ := save.NewStore(db, resolver, nil)
	service, _ := production.NewService(store, resolver, nil, nil, nil)
	api, err := NewAPI(repository, service, APIConfig{UnauthenticatedBurst: 1, UnauthenticatedPerMin: 1, AccountBurst: 1, AccountPerMin: 1, MaxBodyBytes: 64 << 10})
	if err != nil {
		t.Fatal(err)
	}
	server := testhttp.New(api.Router())
	defer server.Close()
	first := requestJSON(t, server.Client, http.MethodPost, server.URL+"/api/v1/account", "", `{}`)
	if first.StatusCode != http.StatusCreated {
		t.Fatalf("first status=%d body=%s", first.StatusCode, readBody(first))
	}
	second := requestJSON(t, server.Client, http.MethodPost, server.URL+"/api/v1/account", "", `{}`)
	if second.StatusCode != http.StatusTooManyRequests || !strings.Contains(readBody(second), `"category":"rate_limited"`) {
		t.Fatalf("second status=%d", second.StatusCode)
	}
}

func seedAccountEpoch(t *testing.T, db *sql.DB, hash string, artifact []byte) {
	t.Helper()
	if _, err := db.Exec(`INSERT INTO catalog_sets(constants_hash) VALUES($1)`, hash); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO catalog_artifacts(constants_hash,artifact_name,bytes) VALUES($1,'economy',$2)`, hash, artifact); err != nil {
		t.Fatal(err)
	}
	var epochID int64
	if err := db.QueryRow(`INSERT INTO epochs(name,started_at,changelog_ref) VALUES('Phase 0',now(),'changelog/epoch-1.md') RETURNING epoch_id`).Scan(&epochID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO epoch_hashes(epoch_id,constants_hash) VALUES($1,$2)`, epochID, hash); err != nil {
		t.Fatal(err)
	}
}

func truncateAccountIntegration(t *testing.T, db *sql.DB) {
	t.Helper()
	_, err := db.Exec(`TRUNCATE epochs,catalog_sets,access_tokens,sessions,account_emails,account_founders,accounts,commons_recruitment_offers,commons_health_scopes,commons_member_samples,commons_projection_events,company_compact_memberships,founder_commons_assignments,commons_cohorts,registry_routes,route_hint_projection_events,founder_route_state,founder_route_executions,route_projection_events,events,intent_records,save_revisions,save_streams CASCADE`)
	if err != nil {
		t.Fatal(err)
	}
}

func requestJSON(t *testing.T, client *http.Client, method, url, accessToken, body string) *http.Response {
	t.Helper()
	request, err := http.NewRequest(method, url, strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Content-Type", "application/json")
	if accessToken != "" {
		request.Header.Set("Authorization", "Bearer "+accessToken)
	}
	response, err := client.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	return response
}

func decodeResponse(t *testing.T, response *http.Response, target any) {
	t.Helper()
	defer response.Body.Close()
	decoder := json.NewDecoder(response.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		t.Fatal(err)
	}
	if extra, err := io.ReadAll(response.Body); err != nil || len(bytes.TrimSpace(extra)) != 0 {
		t.Fatalf("trailing response=%q err=%v", extra, err)
	}
}

func readBody(response *http.Response) string {
	defer response.Body.Close()
	data, _ := io.ReadAll(response.Body)
	return string(data)
}

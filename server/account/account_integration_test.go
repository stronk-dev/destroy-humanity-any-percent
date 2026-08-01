package account

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"cloud-clicker/server/economy"
	"cloud-clicker/server/epochseed"
	"cloud-clicker/server/guild"
	"cloud-clicker/server/internal/testhttp"
	"cloud-clicker/server/production"
	"cloud-clicker/server/replaycatalog"
	"cloud-clicker/server/save"
)

type integrationCatalogs map[string]*economy.Catalog

// emptyGuildSettlements is a constructor-only account fixture. It is not the
// production guild settlement resolver and must never be cited as composition.
type emptyGuildSettlements struct{}

func (emptyGuildSettlements) PendingSettlements(context.Context, string, string, int64) (guild.SettlementBatch, error) {
	return guild.SettlementBatch{}, nil
}

type integrationGuildNames struct{}

func (integrationGuildNames) HandleGuild(_ context.Context, _ string, body []byte) (json.RawMessage, bool, error) {
	if !json.Valid(body) {
		return nil, false, ErrInvalidRequest
	}
	return json.RawMessage(`{"intent_id":"018f0000-0000-7000-8000-000000000201","outcome":"applied","new_revision":2,"guild_id":"018f0000-0000-7000-8000-000000000201"}`), false, nil
}
func (integrationGuildNames) IsInvalidGuildIntent(err error) bool {
	return errors.Is(err, ErrInvalidRequest)
}

func (catalogs integrationCatalogs) Resolve(hash string) (*economy.Catalog, bool) {
	catalog, ok := catalogs[hash]
	return catalog, ok
}

func (catalogs integrationCatalogs) ValidateState(hash string, state *save.State) error {
	if _, ok := catalogs.Resolve(hash); !ok || state == nil || state.FactionID != "" {
		return save.ErrInvalidState
	}
	return nil
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

	repositoryBundle, err := epochseed.Load(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	catalogBytes := repositoryBundle.Artifacts["economy"]
	catalog, err := economy.LoadCatalog(catalogBytes)
	if err != nil {
		t.Fatal(err)
	}
	hash := repositoryBundle.Hash
	oldCatalogBytes := append(append([]byte(nil), catalogBytes...), ' ')
	oldCatalog, err := economy.LoadCatalog(oldCatalogBytes)
	if err != nil {
		t.Fatal(err)
	}
	oldHash := save.ConstantsHash(oldCatalogBytes)
	replayBundle, err := replaycatalog.Load(hash, repositoryBundle.Artifacts)
	if err != nil {
		t.Fatal(err)
	}
	seedAccountEpoch(t, db, hash, repositoryBundle.Artifacts)
	resolver := integrationCatalogs{hash: catalog, oldHash: oldCatalog}
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
	replaySet := production.ReplayCatalogSet{hash: replayBundle}
	intentService, err := production.NewService(saveStore, resolver, nil, nil, nil, production.WithReplayCatalogs(replaySet),
		production.WithProgressionRuntime(replaySet), production.WithCurrentConstantsHash(hash), production.WithGuildSettlements(emptyGuildSettlements{}))
	if err != nil {
		t.Fatal(err)
	}
	api, err := NewAPI(repository, intentService, APIConfig{UnauthenticatedBurst: 20, UnauthenticatedPerMin: 60, AccountBurst: 100, AccountPerMin: 600, MaxBodyBytes: 64 << 10})
	if err != nil {
		t.Fatal(err)
	}
	if err := api.AttachGuildIntents(integrationGuildNames{}); err != nil {
		t.Fatalf("guild composition: %v", err)
	}
	server := testhttp.New(api.Router())
	defer server.Close()

	createdResponse := requestJSON(t, server.Client, http.MethodPost, server.URL+"/api/v1/account", "", `{}`)
	if createdResponse.StatusCode != http.StatusCreated {
		t.Fatalf("create status=%d body=%s", createdResponse.StatusCode, readBody(createdResponse))
	}
	if createdResponse.Header.Get("Cache-Control") != "no-store" {
		t.Fatalf("recovery response cache control=%q", createdResponse.Header.Get("Cache-Control"))
	}
	var created CreatedAccount
	decodeResponse(t, createdResponse, &created)
	if created.AccountID == "" || created.RecoveryCode == "" || created.FounderID != "" {
		t.Fatalf("created=%+v", created)
	}
	unknownResponse := requestJSON(t, server.Client, http.MethodGet, server.URL+"/api/v1/unknown", "", "")
	if unknownResponse.StatusCode != http.StatusNotFound || !strings.Contains(readBody(unknownResponse), `"category":"unknown_id"`) {
		t.Fatal("router 404 was not typed")
	}
	methodResponse := requestJSON(t, server.Client, http.MethodGet, server.URL+"/api/v1/account", "", "")
	if methodResponse.StatusCode != http.StatusMethodNotAllowed || !strings.Contains(readBody(methodResponse), `"detail":"method"`) {
		t.Fatal("router 405 was not typed")
	}
	var accountColumns int
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM information_schema.columns WHERE table_schema=current_schema() AND table_name='accounts'`).Scan(&accountColumns); err != nil || accountColumns != 3 {
		t.Fatalf("account columns=%d err=%v", accountColumns, err)
	}
	storedUpgradeHash := recoveryHashForTest(created.RecoveryCode, argonMemoryKiB+1024, argonIterations, argonParallelism)
	if _, err := db.ExecContext(ctx, `UPDATE accounts SET recovery_hash=$2 WHERE account_id=$1`, created.AccountID, storedUpgradeHash); err != nil {
		t.Fatal(err)
	}

	sessionResponse := requestJSON(t, server.Client, http.MethodPost, server.URL+"/api/v1/session", "", fmt.Sprintf(`{"account_id":%q,"recovery_code":%q}`, created.AccountID, created.RecoveryCode))
	if sessionResponse.StatusCode != http.StatusOK {
		t.Fatalf("session status=%d body=%s", sessionResponse.StatusCode, readBody(sessionResponse))
	}
	var firstPair TokenPair
	decodeResponse(t, sessionResponse, &firstPair)
	var upgradedHash string
	if err := db.QueryRowContext(ctx, `SELECT recovery_hash FROM accounts WHERE account_id=$1`, created.AccountID).Scan(&upgradedHash); err != nil ||
		upgradedHash == storedUpgradeHash || !strings.Contains(upgradedHash, "$m=19456,t=2,p=1$") {
		t.Fatalf("credential was not upgraded in login transaction: %s err=%v", upgradedHash, err)
	}
	claims, err := repository.Authenticate(ctx, firstPair.AccessToken)
	if err != nil || claims.Subject != created.AccountID {
		t.Fatalf("claims=%+v err=%v", claims, err)
	}
	guildResponse := requestJSON(t, server.Client, http.MethodPost, server.URL+"/api/v1/guild/intents", firstPair.AccessToken,
		`{"intent_id":"018f0000-0000-7000-8000-000000000201","kind":"create_guild","expected_revision":1,"name":"Small Systems","join_policy":"open"}`)
	if guildResponse.StatusCode != http.StatusOK || !strings.Contains(readBody(guildResponse), `"outcome":"applied"`) {
		t.Fatalf("guild intent status=%d", guildResponse.StatusCode)
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
	initialGenesis, err := saveStore.LoadRunGenesis(ctx, newState.StreamID, 1)
	if err != nil || initialGenesis.Version != newState.Version || initialGenesis.ConstantsHash != newState.ConstantsHash || !bytes.Equal(initialGenesis.State, newState.State) {
		t.Fatalf("initial genesis=%+v state_equal=%t err=%v", initialGenesis, bytes.Equal(initialGenesis.State, newState.State), err)
	}
	importedState, err := save.RestoreState(newState.State, newState.Version, catalog, economy.ScopeCompany, time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	importedState.RunSeq = 3
	importedState.RunStartedAt = now.Add(-time.Hour)
	importedStateBytes, err := save.EncodeState(importedState)
	if err != nil {
		t.Fatal(err)
	}
	importBody, _ := json.Marshal(struct {
		Version       int             `json:"version"`
		ConstantsHash string          `json:"constants_hash"`
		State         json.RawMessage `json:"state"`
	}{Version: newState.Version, ConstantsHash: oldHash, State: importedStateBytes})
	importResponse := requestJSON(t, server.Client, http.MethodPost, server.URL+"/api/v1/founder/import", freshPair.AccessToken, string(importBody))
	if importResponse.StatusCode != http.StatusOK {
		t.Fatalf("import status=%d body=%s", importResponse.StatusCode, readBody(importResponse))
	}
	var imported Founder
	decodeResponse(t, importResponse, &imported)
	if !imported.Imported {
		t.Fatal("imported flag did not round-trip")
	}
	afterImport, err := repository.ActiveCompanyState(ctx, created.AccountID)
	if err != nil || afterImport.Revision != 1 || afterImport.ConstantsHash != hash || afterImport.StreamID == newState.StreamID {
		t.Fatalf("imported state=%+v err=%v", afterImport, err)
	}
	importGenesis, err := saveStore.LoadRunGenesis(ctx, afterImport.StreamID, 1)
	if err != nil || importGenesis.Version != afterImport.Version || importGenesis.ConstantsHash != afterImport.ConstantsHash || !bytes.Equal(importGenesis.State, afterImport.State) {
		t.Fatalf("import genesis=%+v state_equal=%t err=%v", importGenesis, bytes.Equal(importGenesis.State, afterImport.State), err)
	}
	restoredImport, err := save.RestoreState(afterImport.State, afterImport.Version, catalog, economy.ScopeCompany, time.Time{})
	if err != nil || restoredImport.RunSeq != 1 || !restoredImport.RunStartedAt.Equal(now) {
		t.Fatalf("restored import run=%d started=%s err=%v", restoredImport.RunSeq, restoredImport.RunStartedAt, err)
	}
	importIntent := `{"intent_id":"01985555-1112-7111-8111-111111111112","kind":"perform_manual_batch","expected_revision":1,"action_id":"manual.click","count":1,"window_ms":1}`
	importIntentResponse := requestJSON(t, server.Client, http.MethodPost, server.URL+"/api/v1/intents", freshPair.AccessToken, importIntent)
	if importIntentResponse.StatusCode != http.StatusOK {
		t.Fatalf("import intent status=%d body=%s", importIntentResponse.StatusCode, readBody(importIntentResponse))
	}
	var importReceipt map[string]json.RawMessage
	decodeResponse(t, importIntentResponse, &importReceipt)
	if string(importReceipt["outcome"]) != `"applied"` || string(importReceipt["new_revision"]) != "2" {
		t.Fatalf("import receipt=%v", importReceipt)
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
	var founderRows, linkedRows, unarchivedRows, importedRows int
	if err := db.QueryRowContext(ctx, `SELECT count(*),count(account_id),count(*) FILTER (WHERE archived_at IS NULL),count(*) FILTER (WHERE imported) FROM account_founders`).Scan(
		&founderRows, &linkedRows, &unarchivedRows, &importedRows); err != nil || founderRows != 3 || linkedRows != 0 || unarchivedRows != 0 || importedRows != 1 {
		t.Fatalf("anonymized founders rows=%d linked=%d active=%d imported=%d err=%v", founderRows, linkedRows, unarchivedRows, importedRows, err)
	}
}

func TestConcurrentRefreshReplayRevokesEntireFamilyIntegration(t *testing.T) {
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
	seedAccountEpoch(t, db, hash, map[string][]byte{"economy": catalogBytes})
	repository, err := NewRepository(db, integrationCatalogs{hash: catalog}, hash,
		SigningKeys{CurrentID: "test", Current: bytes.Repeat([]byte{0x42}, 32)}, time.Now, nil)
	if err != nil {
		t.Fatal(err)
	}
	created, err := repository.CreateAccount(ctx)
	if err != nil {
		t.Fatal(err)
	}
	first, err := repository.CreateSession(ctx, created.AccountID, created.RecoveryCode)
	if err != nil {
		t.Fatal(err)
	}
	second, err := repository.RefreshSession(ctx, first.RefreshToken)
	if err != nil {
		t.Fatal(err)
	}
	firstHash, _ := opaqueTokenHash(first.RefreshToken)
	var familyID string
	if err := db.QueryRowContext(ctx, `SELECT family_id FROM sessions WHERE token_hash=$1`, firstHash[:]).Scan(&familyID); err != nil {
		t.Fatal(err)
	}
	blocker, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := blocker.ExecContext(ctx, `SELECT family_id FROM session_families WHERE family_id=$1 FOR UPDATE`, familyID); err != nil {
		t.Fatal(err)
	}
	type refreshResult struct {
		pair TokenPair
		err  error
	}
	legitimate := make(chan refreshResult, 1)
	replay := make(chan refreshResult, 1)
	go func() {
		pair, err := repository.RefreshSession(ctx, second.RefreshToken)
		legitimate <- refreshResult{pair: pair, err: err}
	}()
	time.Sleep(25 * time.Millisecond)
	go func() {
		pair, err := repository.RefreshSession(ctx, first.RefreshToken)
		replay <- refreshResult{pair: pair, err: err}
	}()
	time.Sleep(25 * time.Millisecond)
	if err := blocker.Commit(); err != nil {
		t.Fatal(err)
	}
	legitimateResult, replayResult := <-legitimate, <-replay
	if !errors.Is(legitimateResult.err, ErrAuthentication) && !errors.Is(legitimateResult.err, ErrRefreshReuse) && legitimateResult.err != nil {
		t.Fatalf("legitimate err=%v", legitimateResult.err)
	}
	if !errors.Is(replayResult.err, ErrRefreshReuse) && !errors.Is(replayResult.err, ErrAuthentication) {
		t.Fatalf("replay err=%v", replayResult.err)
	}
	if !errors.Is(legitimateResult.err, ErrRefreshReuse) && !errors.Is(replayResult.err, ErrRefreshReuse) {
		t.Fatalf("reuse was not detected legitimate=%v replay=%v", legitimateResult.err, replayResult.err)
	}
	for _, pair := range []TokenPair{legitimateResult.pair, replayResult.pair} {
		if pair.AccessToken != "" {
			if _, err := repository.Authenticate(ctx, pair.AccessToken); !errors.Is(err, ErrAuthentication) {
				t.Fatalf("descendant access token survived family revocation: %v", err)
			}
		}
		if pair.RefreshToken != "" {
			if _, err := repository.RefreshSession(ctx, pair.RefreshToken); err == nil {
				t.Fatal("descendant refresh token survived family revocation")
			}
		}
	}
	var familyRevoked bool
	if err := db.QueryRowContext(ctx, `SELECT revoked_at IS NOT NULL FROM session_families WHERE family_id=$1`, familyID).Scan(&familyRevoked); err != nil || !familyRevoked {
		t.Fatalf("family revoked=%v err=%v", familyRevoked, err)
	}
	var live int
	if err := db.QueryRowContext(ctx, `SELECT (SELECT count(*) FROM sessions WHERE family_id=$1 AND revoked_at IS NULL)+(SELECT count(*) FROM access_tokens WHERE family_id=$1 AND revoked_at IS NULL)`, familyID).Scan(&live); err != nil || live != 0 {
		t.Fatalf("live family rows=%d err=%v", live, err)
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
	bundle, err := epochseed.Load(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := economy.LoadCatalog(bundle.Artifacts["economy"])
	if err != nil {
		t.Fatal(err)
	}
	hash := bundle.Hash
	seedAccountEpoch(t, db, hash, bundle.Artifacts)
	resolver := integrationCatalogs{hash: catalog}
	repository, err := NewRepository(db, resolver, hash, SigningKeys{CurrentID: "test", Current: bytes.Repeat([]byte{1}, 32)}, time.Now, nil)
	if err != nil {
		t.Fatal(err)
	}
	store, _ := save.NewStore(db, resolver, nil)
	replayBundle, err := replaycatalog.Load(bundle.Hash, bundle.Artifacts)
	if err != nil {
		t.Fatal(err)
	}
	replaySet := production.ReplayCatalogSet{bundle.Hash: replayBundle}
	service, err := production.NewService(store, resolver, nil, nil, nil, production.WithReplayCatalogs(replaySet),
		production.WithProgressionRuntime(replaySet), production.WithCurrentConstantsHash(bundle.Hash), production.WithGuildSettlements(emptyGuildSettlements{}))
	if err != nil {
		t.Fatal(err)
	}
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

func seedAccountEpoch(t *testing.T, db *sql.DB, hash string, artifacts map[string][]byte) {
	t.Helper()
	if _, err := db.Exec(`INSERT INTO catalog_sets(constants_hash) VALUES($1)`, hash); err != nil {
		t.Fatal(err)
	}
	for name, artifact := range artifacts {
		if _, err := db.Exec(`INSERT INTO catalog_artifacts(constants_hash,artifact_name,bytes) VALUES($1,$2,$3)`, hash, name, artifact); err != nil {
			t.Fatal(err)
		}
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

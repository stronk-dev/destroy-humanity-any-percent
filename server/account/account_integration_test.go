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

type integrationSoulRecoveries struct {
	start production.StartSoulRecoveryRequest
}

func (handler *integrationSoulRecoveries) StartSoulRecovery(_ context.Context, request production.StartSoulRecoveryRequest, _ time.Time) (production.HandleResult, error) {
	handler.start = request
	return production.HandleResult{Receipt: json.RawMessage(fmt.Sprintf(`{"session_id":%q,"progress_token":"018f0000-0000-4000-8000-000000000202","activity_id":%q,"required_duration_attended_ms":300000,"attended_progress_ms":0,"last_progress_server_ms":1,"started_wall_ms":1}`, request.SessionID, request.ActivityID))}, nil
}
func (*integrationSoulRecoveries) ProgressSoulRecovery(_ context.Context, request production.ProgressSoulRecoveryRequest, _ time.Time, _ save.ExitFaultInjector) (production.HandleResult, error) {
	return production.HandleResult{Receipt: json.RawMessage(fmt.Sprintf(`{"session_id":%q,"attended_progress_ms":0,"required_duration_attended_ms":300000,"last_progress_server_ms":1,"eligible":false}`, request.SessionID))}, nil
}
func (*integrationSoulRecoveries) CancelSoulRecovery(_ context.Context, request production.FinishSoulRecoveryRequest, _ time.Time, _ save.ExitFaultInjector) (production.HandleResult, error) {
	return production.HandleResult{Receipt: json.RawMessage(fmt.Sprintf(`{"session_id":%q,"outcome":"applied","action":"cancel_soul_recovery"}`, request.SessionID))}, nil
}
func (*integrationSoulRecoveries) ResolveSoulRecovery(_ context.Context, request production.FinishSoulRecoveryRequest, _ time.Time, _ save.ExitFaultInjector) (production.HandleResult, error) {
	return production.HandleResult{Receipt: json.RawMessage(fmt.Sprintf(`{"session_id":%q,"outcome":"applied","action":"resolve_soul_recovery"}`, request.SessionID))}, nil
}
func (*integrationSoulRecoveries) SoulRecoveryBeatCeilingMS(context.Context, string, string) (int64, error) {
	return 6000, nil
}

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

	repositoryBundle := epoch5AccountIntegrationBundle(t)
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
	api, err := NewAPI(repository, intentService, APIConfig{UnauthenticatedBurst: 20, UnauthenticatedPerMin: 60, AccountBurst: 100, AccountPerMin: 600, MaxBodyBytes: 64 << 10, BootstrapReceiptKeys: testBootstrapReceiptKeys()})
	if err != nil {
		t.Fatal(err)
	}
	if err := api.AttachGuildIntents(integrationGuildNames{}); err != nil {
		t.Fatalf("guild composition: %v", err)
	}
	recoveries := &integrationSoulRecoveries{}
	if err := api.AttachSoulRecoveries(recoveries); err != nil {
		t.Fatalf("Soul recovery composition: %v", err)
	}
	minigames := &minigameAPIStub{result: json.RawMessage(`{"constants_hash":"` + hash + `","engine_ref":"pitch","engine_version":"1.0.0","minigame_id":"pitch","mode":"solo","revision":1,"session_id":"01986666-ca01-7000-8000-000000000010","snapshot":{"deck_count":17,"funding_target":"1e3","hand":[],"hands_remaining":3,"phase":"playing","pitch_content_hash":"sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb","pitch_schema_version":1,"revision":1,"round":1,"round_best_valuation":"0","run_currency":4,"shop_offers":[],"slotted_hacks":[]},"status":"active"}`)}
	if err := api.AttachMinigames(minigames); err != nil {
		t.Fatalf("minigame composition: %v", err)
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
	minigameCreate := requestJSON(t, server.Client, http.MethodPost, server.URL+"/api/v1/minigames/pitch/sessions", firstPair.AccessToken,
		`{"idempotency_key":"create-1"}`)
	if minigameCreate.StatusCode != http.StatusOK {
		t.Fatalf("minigame create status=%d body=%s", minigameCreate.StatusCode, readBody(minigameCreate))
	}
	minigameBytes := []byte(readBody(minigameCreate))
	if err := api.privateRegistry.ValidateResponse("create_minigame_session", http.StatusOK, minigameBytes); err != nil ||
		minigames.createAccount != created.AccountID || minigames.createGame != "pitch" || minigames.createKey != "create-1" {
		t.Fatalf("registered minigame route response=%s stub=%+v err=%v", minigameBytes, minigames, err)
	}
	oldState, err := repository.ActiveCompanyState(ctx, created.AccountID)
	if err != nil || oldState.Revision != 1 {
		t.Fatalf("initial state=%+v err=%v", oldState, err)
	}
	recoveryStart := requestJSON(t, server.Client, http.MethodPost, server.URL+"/api/v1/soul-recovery/start", firstPair.AccessToken,
		`{"activity_id":"repot"}`)
	if recoveryStart.StatusCode != http.StatusOK {
		t.Fatalf("Soul recovery start status=%d body=%s", recoveryStart.StatusCode, readBody(recoveryStart))
	}
	var recoveryStartReceipt map[string]json.RawMessage
	decodeResponse(t, recoveryStart, &recoveryStartReceipt)
	var recoverySessionID string
	if json.Unmarshal(recoveryStartReceipt["session_id"], &recoverySessionID) != nil || recoveries.start.FounderID != profile.ID ||
		recoveries.start.CompanyStreamID != oldState.StreamID || recoveries.start.ActivityID != "repot" || recoverySessionID == "" {
		t.Fatalf("Soul recovery authority leak/start mismatch: request=%+v receipt=%v", recoveries.start, recoveryStartReceipt)
	}
	strictStart := requestJSON(t, server.Client, http.MethodPost, server.URL+"/api/v1/soul-recovery/start", firstPair.AccessToken,
		fmt.Sprintf(`{"activity_id":"repot","founder_id":%q}`, profile.ID))
	if strictStart.StatusCode != http.StatusBadRequest {
		t.Fatalf("Soul recovery accepted client Founder authority: %d", strictStart.StatusCode)
	}
	progressBody := fmt.Sprintf(`{"session_id":%q,"progress_token":"018f0000-0000-4000-8000-000000000202"}`, recoverySessionID)
	for index := 0; index < 6; index++ {
		progress := requestJSON(t, server.Client, http.MethodPost, server.URL+"/api/v1/soul-recovery/progress", firstPair.AccessToken, progressBody)
		if progress.StatusCode != http.StatusOK {
			t.Fatalf("Soul progress burst %d status=%d body=%s", index, progress.StatusCode, readBody(progress))
		}
	}
	limited := requestJSON(t, server.Client, http.MethodPost, server.URL+"/api/v1/soul-recovery/progress", firstPair.AccessToken, progressBody)
	if limited.StatusCode != http.StatusTooManyRequests || !strings.Contains(readBody(limited), `"detail":"recovery_progress"`) {
		t.Fatalf("Soul progress limiter status=%d", limited.StatusCode)
	}
	cancelled := requestJSON(t, server.Client, http.MethodPost, server.URL+"/api/v1/soul-recovery/cancel", firstPair.AccessToken,
		fmt.Sprintf(`{"session_id":%q}`, recoverySessionID))
	if cancelled.StatusCode != http.StatusOK {
		t.Fatalf("Soul recovery cancel status=%d body=%s", cancelled.StatusCode, readBody(cancelled))
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
	firstNewState := newState
	initialGenesis, err := saveStore.LoadRunGenesis(ctx, newState.StreamID, 1)
	if err != nil || initialGenesis.Version != newState.Version || initialGenesis.ConstantsHash != newState.ConstantsHash || !bytes.Equal(initialGenesis.State, newState.State) {
		t.Fatalf("initial genesis=%+v state_equal=%t err=%v", initialGenesis, bytes.Equal(initialGenesis.State, newState.State), err)
	}
	for replacement := 2; replacement <= 3; replacement++ {
		response := requestJSON(t, server.Client, http.MethodPost, server.URL+"/api/v1/founder", freshPair.AccessToken, `{}`)
		if response.StatusCode != http.StatusCreated {
			t.Fatalf("New Founder replacement %d status=%d body=%s", replacement, response.StatusCode, readBody(response))
		}
		var nextFounder Founder
		decodeResponse(t, response, &nextFounder)
		if nextFounder.ID == newFounder.ID || nextFounder.Imported {
			t.Fatalf("New Founder replacement %d=%+v prior=%+v", replacement, nextFounder, newFounder)
		}
		priorState := newState
		newState, err = repository.ActiveCompanyState(ctx, created.AccountID)
		if err != nil || newState.Revision != 1 || newState.StreamID == priorState.StreamID ||
			newState.Version != firstNewState.Version || newState.ConstantsHash != firstNewState.ConstantsHash ||
			!bytes.Equal(newState.State, firstNewState.State) {
			t.Fatalf("New Founder replacement %d state=%+v exact_initial=%t err=%v", replacement, newState,
				bytes.Equal(newState.State, firstNewState.State), err)
		}
		archivedPrior, loadErr := saveStore.LoadLatest(ctx, priorState.StreamID)
		if loadErr != nil || archivedPrior.ArchivedAt == nil {
			t.Fatalf("New Founder replacement %d prior stream=%+v err=%v", replacement, archivedPrior, loadErr)
		}
		newFounder = nextFounder
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
	streamRows, err := db.QueryContext(ctx, `SELECT s.id FROM save_streams s JOIN account_founders af ON af.founder_id=s.owner_id
		WHERE af.account_id=$1 AND s.owner_kind='founder' ORDER BY s.id`, created.AccountID)
	if err != nil {
		t.Fatal(err)
	}
	var retainedStreamIDs []string
	for streamRows.Next() {
		var streamID string
		if err := streamRows.Scan(&streamID); err != nil {
			streamRows.Close()
			t.Fatal(err)
		}
		retainedStreamIDs = append(retainedStreamIDs, streamID)
	}
	if err := streamRows.Close(); err != nil {
		t.Fatal(err)
	}
	if len(retainedStreamIDs) != 10 {
		t.Fatalf("pre-delete Founder/Company stream count=%d want=10", len(retainedStreamIDs))
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO bootstrap_receipts
		(request_digest,account_id,key_id,nonce,ciphertext,created_at,refresh_expires_at)
		VALUES($1,$2,'account-deletion-witness',$3,$4,$5,$6)`, bytes.Repeat([]byte{0xa1}, 32), created.AccountID,
		bytes.Repeat([]byte{0xb2}, 12), bytes.Repeat([]byte{0xc3}, 17), now, now.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}

	deleteResponse := requestJSON(t, server.Client, http.MethodDelete, server.URL+"/api/v1/account", freshPair.AccessToken, "")
	if deleteResponse.StatusCode != http.StatusNoContent {
		t.Fatalf("delete status=%d body=%s", deleteResponse.StatusCode, readBody(deleteResponse))
	}
	var accounts, emails, sessions, accessTokens, families, liveBootstrapSecrets int
	if err := db.QueryRowContext(ctx, `SELECT (SELECT count(*) FROM accounts),(SELECT count(*) FROM account_emails),
		(SELECT count(*) FROM sessions),(SELECT count(*) FROM access_tokens),(SELECT count(*) FROM session_families),
		(SELECT count(*) FROM bootstrap_receipts WHERE account_id=$1 AND
			(tombstoned_at IS NULL OR key_id IS NOT NULL OR nonce IS NOT NULL OR ciphertext IS NOT NULL))`, created.AccountID).
		Scan(&accounts, &emails, &sessions, &accessTokens, &families, &liveBootstrapSecrets); err != nil ||
		accounts != 0 || emails != 0 || sessions != 0 || accessTokens != 0 || families != 0 || liveBootstrapSecrets != 0 {
		t.Fatalf("account-linked rows accounts=%d emails=%d sessions=%d access=%d families=%d bootstrap_secrets=%d err=%v",
			accounts, emails, sessions, accessTokens, families, liveBootstrapSecrets, err)
	}
	for _, streamID := range retainedStreamIDs {
		retained, loadErr := saveStore.LoadLatest(ctx, streamID)
		if loadErr != nil || retained.ArchivedAt == nil {
			t.Fatalf("anonymized stream %s missing after account deletion: %+v err=%v", streamID, retained, loadErr)
		}
	}
	var founderRows, linkedRows, unarchivedRows, importedRows int
	if err := db.QueryRowContext(ctx, `SELECT count(*),count(account_id),count(*) FILTER (WHERE archived_at IS NULL),count(*) FILTER (WHERE imported) FROM account_founders`).Scan(
		&founderRows, &linkedRows, &unarchivedRows, &importedRows); err != nil || founderRows != 5 || linkedRows != 0 || unarchivedRows != 0 || importedRows != 1 {
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
	companyRevision, err := repository.ActiveCompanyState(ctx, created.AccountID)
	if err != nil {
		t.Fatal(err)
	}
	company, err := save.RestoreState(companyRevision.State, companyRevision.Version, catalog, economy.ScopeCompany, time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	if got, ok := company.ProvisionRemaindersPPM["generator.beige_tower"]; !ok || got != 0 {
		t.Fatalf("new company provision remainder=%d present=%t", got, ok)
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
	bundle := epoch5AccountIntegrationBundle(t)
	catalog, err := economy.LoadCatalog(bundle.Artifacts["economy"])
	if err != nil {
		t.Fatal(err)
	}
	hash := bundle.Hash
	seedAccountEpoch(t, db, hash, bundle.Artifacts)
	resolver := integrationCatalogs{hash: catalog}
	now := time.Date(2026, 8, 21, 12, 0, 0, 0, time.UTC)
	repository, err := NewRepository(db, resolver, hash, SigningKeys{CurrentID: "test", Current: bytes.Repeat([]byte{1}, 32)}, func() time.Time { return now }, nil)
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
	api, err := NewAPI(repository, service, APIConfig{UnauthenticatedBurst: 1, UnauthenticatedPerMin: 1, AccountBurst: 1, AccountPerMin: 1,
		MaxBodyBytes: 64 << 10, TrustedProxyHops: 1, BootstrapReceiptKeys: testBootstrapReceiptKeys()})
	if err != nil {
		t.Fatal(err)
	}
	if err := api.AttachGameUI(rateLimitGameUI{}); err != nil {
		t.Fatal(err)
	}
	server := testhttp.New(api.Router())
	defer server.Close()

	assertLimited := func(response *http.Response, operation string) {
		t.Helper()
		body := readBody(response)
		if response.StatusCode != http.StatusTooManyRequests || body != "{\"category\":\"rate_limited\",\"detail\":\"ip\"}\n" {
			t.Fatalf("%s limiter status=%d body=%s", operation, response.StatusCode, body)
		}
	}
	assertCounts := func(want [7]int, operation string) {
		t.Helper()
		if got := bootstrapCounts(t, db); got != want {
			t.Fatalf("%s mutated state while limited: got=%v want=%v", operation, got, want)
		}
	}
	refill := func() { now = now.Add(time.Minute) }

	accountIP := "192.0.2.10"
	firstAccount := requestJSONFromIP(t, server.Client, http.MethodPost, server.URL+"/api/v1/account", "", `{}`, accountIP)
	if firstAccount.StatusCode != http.StatusCreated {
		t.Fatalf("first account status=%d body=%s", firstAccount.StatusCode, readBody(firstAccount))
	}
	_ = readBody(firstAccount)
	accountCounts := bootstrapCounts(t, db)
	assertLimited(requestJSONFromIP(t, server.Client, http.MethodPost, server.URL+"/api/v1/account", "", `{}`, accountIP), "create account")
	assertCounts(accountCounts, "create account")
	refill()
	refilledAccount := requestJSONFromIP(t, server.Client, http.MethodPost, server.URL+"/api/v1/account", "", `{}`, accountIP)
	if refilledAccount.StatusCode != http.StatusCreated {
		t.Fatalf("refilled account status=%d body=%s", refilledAccount.StatusCode, readBody(refilledAccount))
	}
	_ = readBody(refilledAccount)

	loginAccount, err := repository.CreateAccount(ctx)
	if err != nil {
		t.Fatal(err)
	}
	loginBody := fmt.Sprintf(`{"account_id":%q,"recovery_code":%q}`, loginAccount.AccountID, loginAccount.RecoveryCode)
	loginIP := "192.0.2.20"
	firstLogin := requestJSONFromIP(t, server.Client, http.MethodPost, server.URL+"/api/v1/session", "", loginBody, loginIP)
	if firstLogin.StatusCode != http.StatusOK {
		t.Fatalf("first recovery login status=%d body=%s", firstLogin.StatusCode, readBody(firstLogin))
	}
	_ = readBody(firstLogin)
	loginCounts := bootstrapCounts(t, db)
	assertLimited(requestJSONFromIP(t, server.Client, http.MethodPost, server.URL+"/api/v1/session", "", loginBody, loginIP), "recovery login")
	assertCounts(loginCounts, "recovery login")
	refill()
	refilledLogin := requestJSONFromIP(t, server.Client, http.MethodPost, server.URL+"/api/v1/session", "", loginBody, loginIP)
	if refilledLogin.StatusCode != http.StatusOK {
		t.Fatalf("refilled recovery login status=%d body=%s", refilledLogin.StatusCode, readBody(refilledLogin))
	}
	_ = readBody(refilledLogin)

	refreshAccount, err := repository.CreateAccount(ctx)
	if err != nil {
		t.Fatal(err)
	}
	firstFamily, err := repository.CreateSession(ctx, refreshAccount.AccountID, refreshAccount.RecoveryCode)
	if err != nil {
		t.Fatal(err)
	}
	secondFamily, err := repository.CreateSession(ctx, refreshAccount.AccountID, refreshAccount.RecoveryCode)
	if err != nil {
		t.Fatal(err)
	}
	refreshIP := "192.0.2.30"
	firstRefresh := requestJSONFromIP(t, server.Client, http.MethodPost, server.URL+"/api/v1/session/refresh", "",
		fmt.Sprintf(`{"refresh_token":%q}`, firstFamily.RefreshToken), refreshIP)
	if firstRefresh.StatusCode != http.StatusOK {
		t.Fatalf("first refresh status=%d body=%s", firstRefresh.StatusCode, readBody(firstRefresh))
	}
	_ = readBody(firstRefresh)
	refreshCounts := bootstrapCounts(t, db)
	secondRefreshBody := fmt.Sprintf(`{"refresh_token":%q}`, secondFamily.RefreshToken)
	assertLimited(requestJSONFromIP(t, server.Client, http.MethodPost, server.URL+"/api/v1/session/refresh", "", secondRefreshBody, refreshIP), "session refresh")
	assertCounts(refreshCounts, "session refresh")
	refill()
	refilledRefresh := requestJSONFromIP(t, server.Client, http.MethodPost, server.URL+"/api/v1/session/refresh", "", secondRefreshBody, refreshIP)
	if refilledRefresh.StatusCode != http.StatusOK {
		t.Fatalf("refilled refresh status=%d body=%s", refilledRefresh.StatusCode, readBody(refilledRefresh))
	}
	_ = readBody(refilledRefresh)

	bootstrapIP := "192.0.2.40"
	firstBootstrap := requestJSONFromIP(t, server.Client, http.MethodPost, server.URL+"/api/v1/bootstrap", "",
		fmt.Sprintf(`{"idempotency_key":%q}`, strings.Repeat("ab", 32)), bootstrapIP)
	if firstBootstrap.StatusCode != http.StatusCreated {
		t.Fatalf("first bootstrap status=%d body=%s", firstBootstrap.StatusCode, readBody(firstBootstrap))
	}
	_ = readBody(firstBootstrap)
	bootstrapState := bootstrapCounts(t, db)
	secondBootstrapBody := fmt.Sprintf(`{"idempotency_key":%q}`, strings.Repeat("cd", 32))
	assertLimited(requestJSONFromIP(t, server.Client, http.MethodPost, server.URL+"/api/v1/bootstrap", "", secondBootstrapBody, bootstrapIP), "bootstrap")
	assertCounts(bootstrapState, "bootstrap")
	refill()
	refilledBootstrap := requestJSONFromIP(t, server.Client, http.MethodPost, server.URL+"/api/v1/bootstrap", "", secondBootstrapBody, bootstrapIP)
	if refilledBootstrap.StatusCode != http.StatusCreated {
		t.Fatalf("refilled bootstrap status=%d body=%s", refilledBootstrap.StatusCode, readBody(refilledBootstrap))
	}
	_ = readBody(refilledBootstrap)
}

type rateLimitGameUI struct{ bootstrapSnapshotFixture }

func (rateLimitGameUI) GameUISnapshot(context.Context, string, time.Time) (json.RawMessage, error) {
	return nil, errors.New("unused rate-limit fixture")
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

type accountIntegrationBundle struct {
	Hash      string
	Artifacts map[string][]byte
}

func epoch5AccountIntegrationBundle(t *testing.T) accountIntegrationBundle {
	t.Helper()
	files := map[string]string{"categories": "categories.json", "commons": "commons.json", "economy": "economy.json",
		"factions": "factions.json", "guilds": "guilds.json", "prestige": "prestige.json", "routes": "routes.json"}
	artifacts := make(map[string][]byte, len(files))
	for name, filename := range files {
		data, err := os.ReadFile(filepath.Join("..", "..", "balance", "testdata", "epoch5", filename))
		if err != nil {
			t.Fatalf("read epoch-5 %s: %v", name, err)
		}
		artifacts[name] = data
	}
	hash, err := save.ConstantsHashArtifacts(artifacts)
	if err != nil || hash != "sha256:63ab30c96b5d76b941b053131fcee63c94b6b3ad91322f9160d94973ce8c58fa" {
		t.Fatalf("epoch-5 account fixture hash=%s err=%v", hash, err)
	}
	return accountIntegrationBundle{Hash: hash, Artifacts: artifacts}
}

func truncateAccountIntegration(t *testing.T, db *sql.DB) {
	t.Helper()
	_, err := db.Exec(`TRUNCATE bootstrap_receipts,epochs,catalog_sets,access_tokens,sessions,account_emails,account_founders,accounts,commons_recruitment_offers,commons_health_scopes,commons_member_samples,commons_projection_events,company_compact_memberships,founder_commons_assignments,commons_cohorts,registry_routes,route_hint_projection_events,founder_route_state,founder_route_executions,route_projection_events,events,intent_records,save_revisions,save_streams CASCADE`)
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

func requestJSONFromIP(t *testing.T, client *http.Client, method, url, accessToken, body, clientIP string) *http.Response {
	t.Helper()
	request, err := http.NewRequest(method, url, strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Forwarded-For", clientIP)
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

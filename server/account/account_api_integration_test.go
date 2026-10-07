package account

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"cloud-clicker/server/internal/testhttp"
)

func TestAccountLifecycleOperationRegistryIntegration(t *testing.T) {
	now := time.Date(2026, 10, 8, 0, 0, 0, 0, time.UTC)
	repository, db, _ := bootstrapRepository(t, &now)
	server := newRefreshPersistenceServer(t, repository, 50)
	registry, err := PrivateAPIRegistry()
	if err != nil {
		t.Fatal(err)
	}
	request := func(id, token, body string, status int) []byte {
		t.Helper()
		operation, ok := registry.Operation(id)
		if !ok {
			t.Fatalf("%s is absent from the mounting authority", id)
		}
		response := requestJSON(t, server.Client, operation.Method, server.URL+operation.Path, token, body)
		defer response.Body.Close()
		data, readErr := io.ReadAll(response.Body)
		if readErr != nil || response.StatusCode != status || response.Header.Get("Content-Type") != "application/json" {
			t.Fatalf("%s lost its status/JSON response (body withheld)", id)
		}
		if registry.ValidateResponse(id, status, data) != nil {
			t.Fatalf("%s actual reply violates its descriptor (body withheld)", id)
		}
		if id == "create_account" && status == http.StatusCreated && response.Header.Get("Cache-Control") != "no-store" {
			t.Fatal("one-time recovery response lost its existing no-store header")
		}
		return data
	}
	initial := refreshPersistenceSnapshot(t, db, true)
	for _, body := range []string{`{"email":"private"}`, `null`, `{ }`, `[]`, `{`} {
		request("create_account", "", body, http.StatusBadRequest)
		assertRefreshPersistenceUnchanged(t, db, true, initial)
	}
	var created CreatedAccount
	// Preserve both historical accepted forms. The second caller also exercises
	// time.Time's trimmed fractional encoding rather than a fixed .000Z fixture.
	for index, body := range []string{"", `{}`} {
		if index == 1 {
			now = now.Add(120 * time.Millisecond)
		}
		var value CreatedAccount
		if json.Unmarshal(request("create_account", "", body, http.StatusCreated), &value) != nil || value.AccountID == "" || !validRecoveryCode(value.RecoveryCode) || value.FounderID != "" || !value.CreatedAt.Equal(now) {
			t.Fatal("account creation lost its exact private credential/identity/time contract")
		}
		if index == 0 {
			created = value
		} else if value.AccountID == created.AccountID || value.RecoveryCode == created.RecoveryCode {
			t.Fatal("separate account calls reused an identity or recovery credential")
		}
	}
	var storedHash, firstFounderID string
	if db.QueryRow(`SELECT a.recovery_hash,af.founder_id FROM accounts a JOIN account_founders af USING(account_id) WHERE a.account_id=$1 AND af.archived_at IS NULL`, created.AccountID).Scan(&storedHash, &firstFounderID) != nil || !verifyRecoveryCode(storedHash, created.RecoveryCode) {
		t.Fatal("returned account credential does not authenticate against the stored hash")
	}
	var pair TokenPair
	if json.Unmarshal(request("create_session", "", fmt.Sprintf(`{"account_id":%q,"recovery_code":%q}`, created.AccountID, created.RecoveryCode), http.StatusOK), &pair) != nil {
		t.Fatal("created account cannot establish a session")
	}
	assertProfile := func(id string) {
		t.Helper()
		var profile struct {
			ID        string                     `json:"id"`
			CreatedAt time.Time                  `json:"created_at"`
			Display   map[string]json.RawMessage `json:"display"`
		}
		if json.Unmarshal(request("get_founder", pair.AccessToken, "", http.StatusOK), &profile) != nil || profile.ID != id || profile.Display == nil || len(profile.Display) != 0 {
			t.Fatal("Founder read lost account ownership or its exact empty display object")
		}
	}
	assertProfile(firstFounderID)
	beforeRefusals := refreshPersistenceSnapshot(t, db, true)
	request("get_founder", "", "", http.StatusUnauthorized)
	request("create_founder", "", `{}`, http.StatusUnauthorized)
	request("create_founder", pair.AccessToken, `null`, http.StatusBadRequest)
	assertRefreshPersistenceUnchanged(t, db, true, beforeRefusals)
	var next Founder
	if json.Unmarshal(request("create_founder", pair.AccessToken, `{}`, http.StatusCreated), &next) != nil || next.ID == firstFounderID || next.ID == "" || next.Imported || !next.CreatedAt.Equal(now) {
		t.Fatal("New Founder failed to return a distinct server-created identity")
	}
	var archivedFounders, archivedStreams, activeFounders int
	if db.QueryRow(`SELECT
		(SELECT count(*) FROM account_founders WHERE founder_id=$1 AND archived_at IS NOT NULL),
		(SELECT count(*) FROM save_streams WHERE owner_kind='founder' AND owner_id=$1 AND archived_at IS NOT NULL),
		(SELECT count(*) FROM account_founders WHERE account_id=$2 AND archived_at IS NULL AND founder_id=$3)`, firstFounderID, created.AccountID, next.ID).Scan(&archivedFounders, &archivedStreams, &activeFounders) != nil || archivedFounders != 1 || archivedStreams != 2 || activeFounders != 1 {
		t.Fatal("New Founder lost its archive-not-delete/single-active-Founder behavior")
	}
	assertProfile(next.ID)
	var rotated TokenPair
	if json.Unmarshal(request("refresh_session", "", fmt.Sprintf(`{"refresh_token":%q}`, pair.RefreshToken), http.StatusOK), &rotated) != nil {
		t.Fatal("existing refresh route failed after New Founder")
	}
	claims, err := repository.Authenticate(context.Background(), rotated.AccessToken)
	if err != nil || claims.Subject != created.AccountID || claims.FounderID != next.ID {
		t.Fatal("refresh did not bind the replacement Founder to the same account")
	}
	// Failure to obtain randomness must retain the existing error and commit no
	// account/Founder/gameplay rows. No database outage or credential is logged.
	beforeFault := refreshPersistenceSnapshot(t, db, true)
	originalRandom := repository.random
	repository.random = strings.NewReader("")
	request("create_account", "", `{}`, http.StatusInternalServerError)
	request("create_founder", rotated.AccessToken, `{}`, http.StatusInternalServerError)
	repository.random = originalRandom
	assertRefreshPersistenceUnchanged(t, db, true, beforeFault)
	if _, err := db.Exec(`UPDATE account_founders SET archived_at=$2 WHERE founder_id=$1`, next.ID, now); err != nil {
		t.Fatal(err)
	}
	withoutActiveFounder := refreshPersistenceSnapshot(t, db, true)
	request("get_founder", rotated.AccessToken, "", http.StatusNotFound)
	request("create_founder", rotated.AccessToken, `{}`, http.StatusNotFound)
	assertRefreshPersistenceUnchanged(t, db, true, withoutActiveFounder)
}

func TestFounderRoutesRetainSharedAccountAndFailedAuthLimitersIntegration(t *testing.T) {
	fixture := newRefreshPersistenceFixture(t, 1)
	config := Phase0APIConfig(testBootstrapReceiptKeys())
	config.AccountBurst, config.AccountPerMin = 1, 1
	config.UnauthenticatedBurst, config.UnauthenticatedPerMin = 1, 1
	api, err := NewAPI(fixture.repository, refreshCensusUnreachableIntents{t}, config)
	if err != nil {
		t.Fatal(err)
	}
	registry := api.privateRegistry
	server := testhttp.New(api.Router())
	t.Cleanup(server.Close)
	request := func(id, token, body string, status int, category, detail string) {
		t.Helper()
		operation, ok := registry.Operation(id)
		if !ok {
			t.Fatal("Founder operation missing")
		}
		// The repository is real; testhttp supplies the same HTTP transport fixture
		// as the account population, not a browser or an OS socket claim.
		response := requestJSON(t, server.Client, operation.Method, server.URL+operation.Path, token, body)
		defer response.Body.Close()
		data, readErr := io.ReadAll(response.Body)
		if readErr != nil || response.StatusCode != status || registry.ValidateResponse(id, status, data) != nil || string(data) != string(exactAPIErrorJSON(apiErrorPair{category, detail})[0]) {
			t.Fatal("Founder limiter lost its exact status/error bytes (body withheld)")
		}
	}
	before := refreshPersistenceSnapshot(t, fixture.db, true)
	request("create_founder", fixture.initial.AccessToken, `null`, 400, "invalid", "body")
	request("get_founder", fixture.initial.AccessToken, "", 429, "rate_limited", "account")
	fixture.now = fixture.now.Add(time.Minute)
	request("get_founder", "", "", 401, "unauthorized", "access_token")
	request("create_founder", "", `{}`, 429, "rate_limited", "ip")
	assertRefreshPersistenceUnchanged(t, fixture.db, true, before)
}

package account

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
)

func TestSessionOperationRegistryIntegration(t *testing.T) {
	fixture := newRefreshPersistenceFixture(t, 10)
	registry, err := PrivateAPIRegistry()
	if err != nil {
		t.Fatal(err)
	}
	operation, ok := registry.Operation("create_session")
	if !ok {
		t.Fatal("session creation is absent from the mounting authority")
	}
	gameplay := refreshPersistenceSnapshot(t, fixture.db, false)
	credentialRows := refreshPersistenceSnapshot(t, fixture.db, true)
	request := func(body string, status int) []byte {
		t.Helper()
		response := requestJSON(t, fixture.server.Client, operation.Method, fixture.server.URL+operation.Path, "", body)
		defer response.Body.Close()
		data, readErr := io.ReadAll(response.Body)
		if readErr != nil || response.StatusCode != status || response.Header.Get("Content-Type") != "application/json" {
			t.Fatal("session creation lost its expected status/JSON response (body withheld)")
		}
		if status == http.StatusOK && response.Header.Get("Cache-Control") != "no-store" {
			t.Fatal("issued session credentials are not protected by no-store (body withheld)")
		}
		if registry.ValidateResponse(operation.ID, status, data) != nil {
			t.Fatal("actual session creation response violates its descriptor (body withheld)")
		}
		return data
	}
	request(`{"account_id":7}`, http.StatusBadRequest)
	assertRefreshPersistenceUnchanged(t, fixture.db, true, credentialRows)
	request(fmt.Sprintf(`{"account_id":%q,"recovery_code":"bad"}`, fixture.account.AccountID), http.StatusUnauthorized)
	assertRefreshPersistenceUnchanged(t, fixture.db, true, credentialRows)
	body := fmt.Sprintf(`{"account_id":%q,"recovery_code":%q}`, fixture.account.AccountID, fixture.account.RecoveryCode)
	if registry.ValidateRequest(operation.ID, []byte(body)) != nil {
		t.Fatal("real credential request does not conform to its descriptor (body withheld)")
	}
	var issued TokenPair
	if json.Unmarshal(request(body, http.StatusOK), &issued) != nil {
		t.Fatal("actual session pair cannot decode")
	}
	claims, err := fixture.repository.Authenticate(context.Background(), issued.AccessToken)
	if err != nil || claims.Subject != fixture.account.AccountID || claims.FounderID != fixture.account.FounderID {
		t.Fatal("registered session handler failed to issue authentic account/Founder-bound credentials")
	}
	// Bind the existing persistence oracle to this newly created family, not the
	// setup family's credentials. Then run rotation/reuse through the registry-mounted route.
	hash, valid := opaqueTokenHash(issued.RefreshToken)
	if !valid || fixture.db.QueryRow(`SELECT family_id FROM sessions WHERE token_hash=$1`, hash[:]).Scan(&fixture.familyID) != nil {
		t.Fatal("registered session handler did not persist its own family")
	}
	assertRefreshIssuedPair(t, fixture, issued, fixture.account.FounderID)
	assertRefreshFamilyCounts(t, fixture, [6]int{1, 0, 0, 1, 0, 0})
	rotated := readRefreshPersistencePair(t, fixture.server, issued.RefreshToken)
	assertRefreshIssuedPair(t, fixture, rotated, fixture.account.FounderID)
	assertRefreshFamilyCounts(t, fixture, [6]int{2, 1, 0, 2, 0, 0})
	assertRefreshPersistenceError(t, fixture.server, issued.RefreshToken, http.StatusUnauthorized, "refresh_reused", "session_family_revoked")
	assertRefreshFamilyCounts(t, fixture, [6]int{2, 1, 2, 2, 2, 1})
	assertRefreshPersistenceUnchanged(t, fixture.db, false, gameplay)
}

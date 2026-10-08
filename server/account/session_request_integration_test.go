package account

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
)

// Exercise admission through the mounted router with usable credentials. An
// invalid request must not issue, consume or revoke credentials before refusal.
func TestSessionRequestSchemaAdmissionIntegration(t *testing.T) {
	if os.Getenv("TEST_DATABASE_URL") == "" {
		t.Skip("session request admission requires real Postgres")
	}
	registry, err := PrivateAPIRegistry()
	if err != nil {
		t.Fatal(err)
	}
	for _, operationID := range []string{"create_session", "refresh_session"} {
		t.Run(operationID, func(t *testing.T) {
			operation, ok := registry.Operation(operationID)
			if !ok {
				t.Fatal("session operation is not registered")
			}
			for _, name := range []string{"duplicate", "escaped-duplicate", "wrong-case", "case-alias", "null-root", "missing-field", "null-field", "unknown-field", "trailing-value"} {
				t.Run(name, func(t *testing.T) {
					fixture := newRefreshPersistenceFixture(t, 10)
					key, credential := "refresh_token", fixture.initial.RefreshToken
					prefix, escaped, upper := "{", `refresh_\u0074oken`, "REFRESH_TOKEN"
					if operationID == "create_session" {
						key, credential = "recovery_code", fixture.account.RecoveryCode
						prefix = fmt.Sprintf(`{"account_id":%q,`, fixture.account.AccountID)
						escaped, upper = `recovery_\u0063ode`, "RECOVERY_CODE"
					}
					valid := prefix + fmt.Sprintf(`%q:%q}`, key, credential)
					body := map[string]string{
						"duplicate":         prefix + fmt.Sprintf(`%q:"bad",%q:%q}`, key, key, credential),
						"escaped-duplicate": prefix + fmt.Sprintf(`%q:"bad","%s":%q}`, key, escaped, credential),
						"wrong-case":        prefix + fmt.Sprintf(`%q:%q}`, upper, credential),
						"case-alias":        prefix + fmt.Sprintf(`%q:"bad",%q:%q}`, upper, key, credential),
						"null-root":         "null",
						"missing-field":     "{}",
						"null-field":        prefix + fmt.Sprintf(`%q:null}`, key),
						"unknown-field":     prefix + fmt.Sprintf(`%q:%q,"unexpected":true}`, key, credential),
						"trailing-value":    valid + " {}",
					}[name]
					if registry.ValidateRequest(operationID, []byte(body)) == nil {
						t.Fatal("invalid-input fixture is accepted by the existing descriptor")
					}
					gameplay := refreshPersistenceSnapshot(t, fixture.db, false)
					credentials := refreshPersistenceSnapshot(t, fixture.db, true)
					response := requestJSON(t, fixture.server.Client, operation.Method, fixture.server.URL+operation.Path, "", body)
					data, readErr := io.ReadAll(response.Body)
					response.Body.Close()
					if readErr != nil || response.StatusCode != http.StatusBadRequest ||
						response.Header.Get("Content-Type") != "application/json" || response.Header.Get("Cache-Control") != "no-store" ||
						string(data) != "{\"category\":\"invalid\",\"detail\":\"body\"}\n" ||
						registry.ValidateResponse(operationID, response.StatusCode, data) != nil {
						t.Fatalf("schema-invalid request was not refused with exact uncached invalid/body (status=%d; credentials/body withheld)", response.StatusCode)
					}
					assertRefreshPersistenceUnchanged(t, fixture.db, true, credentials)
					assertRefreshPersistenceUnchanged(t, fixture.db, false, gameplay)
					assertRefreshFamilyCounts(t, fixture, [6]int{1, 0, 0, 1, 0, 0})

					// Escaped spelling of a single canonical key is valid JSON, not a
					// duplicate. The refused request must leave its credential usable.
					valid = prefix + fmt.Sprintf(`"%s":%q}`, escaped, credential)
					if registry.ValidateRequest(operationID, []byte(valid)) != nil {
						t.Fatal("valid escaped-key control does not match the descriptor")
					}
					response = requestJSON(t, fixture.server.Client, operation.Method, fixture.server.URL+operation.Path, "", valid)
					data, readErr = io.ReadAll(response.Body)
					response.Body.Close()
					if readErr != nil || response.StatusCode != http.StatusOK || response.Header.Get("Cache-Control") != "no-store" ||
						registry.ValidateResponse(operationID, response.StatusCode, data) != nil {
						t.Fatal("valid credential no longer issues a conforming uncached session after refusal (body withheld)")
					}
					var pair TokenPair
					if json.Unmarshal(data, &pair) != nil {
						t.Fatal("valid control cannot decode its issued pair")
					}
					claims, authErr := fixture.repository.Authenticate(context.Background(), pair.AccessToken)
					if authErr != nil || claims.Subject != fixture.account.AccountID || claims.FounderID != fixture.account.FounderID {
						t.Fatal("valid control did not issue authentic account/Founder credentials")
					}
					if operationID == "refresh_session" {
						assertRefreshIssuedPair(t, fixture, pair, fixture.account.FounderID)
						assertRefreshFamilyCounts(t, fixture, [6]int{2, 1, 0, 2, 0, 0})
					}
					assertRefreshPersistenceUnchanged(t, fixture.db, false, gameplay)
				})
			}
		})
	}
}

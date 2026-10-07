package account

import (
	"net/http"
	"testing"
	"time"

	"cloud-clicker/server/publicapi"
)

func TestSessionAPIRegistryPinsExistingWire(t *testing.T) {
	registry, err := PrivateAPIRegistry()
	if err != nil {
		t.Fatal(err)
	}
	for _, row := range []struct {
		id, path, request string
	}{
		{"create_session", "/api/v1/session", `{"account_id":"01985555-1111-7111-8111-111111111110","recovery_code":"recovery"}`},
		{"refresh_session", "/api/v1/session/refresh", `{"refresh_token":"refresh"}`},
	} {
		t.Run(row.id, func(t *testing.T) {
			operation, ok := registry.Operation(row.id)
			if !ok || operation.Path != row.path || operation.Method != http.MethodPost || operation.Auth != publicapi.AuthNone || operation.Public || operation.Surface != publicapi.SurfacePrivateV1 {
				t.Fatal("session operation lost its existing unauthenticated private route")
			}
			if err := registry.ValidateRequest(row.id, []byte(row.request)); err != nil {
				t.Fatal(err)
			}
			if err := registry.ValidateResponse(row.id, http.StatusOK, []byte(`{"access_token":"access","refresh_token":"refresh"}`)); err != nil {
				t.Fatal(err)
			}
			for _, body := range []string{
				`{"access_token":"access"}`, `{"access_token":"access","refresh_token":7}`,
				`{"access_token":"access","refresh_token":"refresh","family_id":"private"}`,
			} {
				if registry.ValidateResponse(row.id, http.StatusOK, []byte(body)) == nil {
					t.Fatal("session response admitted missing, mistyped or private fields")
				}
			}
			for _, body := range []string{`{}`, `[]`, `{"refresh_token":"refresh","private":"x"}`} {
				if registry.ValidateRequest(row.id, []byte(body)) == nil {
					t.Fatal("session request admitted missing/extra fields or a nonobject root")
				}
			}
		})
	}
	for _, row := range []struct{ operation, category, detail string }{
		{"create_session", "unauthorized", "credential"},
		{"refresh_session", "unauthorized", "refresh_token"},
		{"refresh_session", "refresh_reused", "session_family_revoked"},
	} {
		body := exactAPIErrorJSON(apiErrorPair{row.category, row.detail})[0]
		if err := registry.ValidateResponse(row.operation, http.StatusUnauthorized, body); err != nil {
			t.Fatal(err)
		}
		other := "create_session"
		if row.operation == other {
			other = "refresh_session"
		}
		if registry.ValidateResponse(other, http.StatusUnauthorized, body) == nil {
			t.Fatal("session operation accepted another operation's error alternative")
		}
	}
}

func TestSessionRoutesRetainTheSharedUnauthenticatedLimiter(t *testing.T) {
	now := time.Date(2026, 10, 7, 0, 0, 0, 0, time.UTC)
	handler := refreshCensusRouter(t, &now)
	assertRefreshCensusResponse(t, handler, http.MethodPost, "/api/v1/session", "{", http.StatusBadRequest, "invalid", "body")
	assertRefreshCensusResponse(t, handler, http.MethodPost, "/api/v1/session/refresh", "{", http.StatusTooManyRequests, "rate_limited", "ip")
	now = now.Add(time.Minute)
	assertRefreshCensusResponse(t, handler, http.MethodPost, "/api/v1/session/refresh", "{", http.StatusBadRequest, "invalid", "body")
	assertRefreshCensusResponse(t, handler, http.MethodPost, "/api/v1/session", "{", http.StatusTooManyRequests, "rate_limited", "ip")
}

package account

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"cloud-clicker/server/production"
)

// This is a parser/router census, not a session integration fixture. No supplied
// token has the opaque-token shape required to reach the intentionally absent DB.
// The unrelated intent dependency exists only to satisfy the real API constructor.
type refreshCensusUnreachableIntents struct{ t *testing.T }

func (fixture refreshCensusUnreachableIntents) Handle(context.Context, string, production.EvaluationMode, time.Time, []byte) (production.HandleResult, error) {
	fixture.t.Helper()
	fixture.t.Fatal("refresh parser census reached an unrelated intent handler")
	return production.HandleResult{}, ErrInvalidRequest
}

func refreshCensusRouter(t *testing.T, now *time.Time) http.Handler {
	t.Helper()
	config := Phase0APIConfig(testBootstrapReceiptKeys())
	config.UnauthenticatedBurst = 1
	config.UnauthenticatedPerMin = 1
	config.MaxBodyBytes = 1024
	api, err := NewAPI(&Repository{clock: func() time.Time { return *now }}, refreshCensusUnreachableIntents{t}, config)
	if err != nil {
		t.Fatal(err)
	}
	return api.Router()
}

func assertRefreshCensusResponse(t *testing.T, handler http.Handler, method, path, body string, status int, category, detail string) {
	t.Helper()
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	request.RemoteAddr = "192.0.2.42:1234"
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != status {
		t.Fatalf("status=%d want=%d", response.Code, status)
	}
	if response.Header().Get("Content-Type") != "application/json" {
		t.Fatal("typed response lost its exact JSON content type")
	}
	want := fmt.Sprintf("{\"category\":%q,\"detail\":%q}\n", category, detail)
	if response.Body.String() != want {
		// Never echo an unexpected response: a broken handler could include tokens.
		t.Fatalf("response differs from exact %s/%s error bytes", category, detail)
	}
	if method == http.MethodPost && path == "/api/v1/session/refresh" {
		registry, err := PrivateAPIRegistry()
		if err != nil {
			t.Fatal(err)
		}
		if registry.ValidateResponse("refresh_session", response.Code, response.Body.Bytes()) != nil {
			t.Fatal("actual refresh refusal is absent from the generated operation (body withheld)")
		}
	}
}

func TestRefreshWireParserCensus(t *testing.T) {
	cases := []struct {
		name, body string
		status     int
	}{
		{"empty", "", http.StatusBadRequest},
		{"whitespace", " \n\t", http.StatusBadRequest},
		{"unfinished-object", "{", http.StatusBadRequest},
		{"unknown-member", `{"refresh_token":"bad","unexpected":true}`, http.StatusBadRequest},
		{"second-json-value", `{} {}`, http.StatusBadRequest},
		{"trailing-non-json", `{} x`, http.StatusBadRequest},
		{"array-root", `[]`, http.StatusBadRequest},
		{"boolean-root", `true`, http.StatusBadRequest},
		{"number-root", `7`, http.StatusBadRequest},
		{"string-root", `"bad"`, http.StatusBadRequest},
		{"number-token", `{"refresh_token":7}`, http.StatusBadRequest},
		{"object-token", `{"refresh_token":{}}`, http.StatusBadRequest},
		{"array-token", `{"refresh_token":[]}`, http.StatusBadRequest},
		{"oversize-body", `{"refresh_token":"` + strings.Repeat("x", 2048) + `"}`, http.StatusBadRequest},
		{"missing-token", `{}`, http.StatusBadRequest},
		{"null-root", `null`, http.StatusBadRequest},
		{"null-token", `{"refresh_token":null}`, http.StatusBadRequest},
		{"empty-token", `{"refresh_token":""}`, http.StatusUnauthorized},
		{"bad-token-encoding", `{"refresh_token":"%not-a-token"}`, http.StatusUnauthorized},
		{"short-token", `{"refresh_token":"YQ"}`, http.StatusUnauthorized},
		{"duplicate-token", `{"refresh_token":"bad","refresh_token":"also-bad"}`, http.StatusBadRequest},
		{"escaped-duplicate-token", `{"refresh_token":"bad","refresh_\u0074oken":"also-bad"}`, http.StatusBadRequest},
		{"case-insensitive-member", `{"REFRESH_TOKEN":"bad"}`, http.StatusBadRequest},
		{"case-alias-member", `{"REFRESH_TOKEN":"bad","refresh_token":"also-bad"}`, http.StatusBadRequest},
		{"valid-escaped-member", `{"refresh_\u0074oken":"bad"}`, http.StatusUnauthorized},
	}
	for _, row := range cases {
		t.Run(row.name, func(t *testing.T) {
			now := time.Date(2026, 10, 6, 0, 0, 0, 0, time.UTC)
			category, detail := "invalid", "body"
			if row.status == http.StatusUnauthorized {
				category, detail = "unauthorized", "refresh_token"
			}
			assertRefreshCensusResponse(t, refreshCensusRouter(t, &now), http.MethodPost, "/api/v1/session/refresh", row.body, row.status, category, detail)
		})
	}
}

func TestRefreshWireLimiterPrecedesParserAndRefills(t *testing.T) {
	now := time.Date(2026, 10, 6, 0, 0, 0, 0, time.UTC)
	handler := refreshCensusRouter(t, &now)
	assertRefreshCensusResponse(t, handler, http.MethodPost, "/api/v1/session/refresh", "{", http.StatusBadRequest, "invalid", "body")
	assertRefreshCensusResponse(t, handler, http.MethodPost, "/api/v1/session/refresh", "{", http.StatusTooManyRequests, "rate_limited", "ip")
	now = now.Add(time.Minute)
	assertRefreshCensusResponse(t, handler, http.MethodPost, "/api/v1/session/refresh", "{", http.StatusBadRequest, "invalid", "body")
}

func TestRefreshWireMethodAndPathCensus(t *testing.T) {
	now := time.Date(2026, 10, 6, 0, 0, 0, 0, time.UTC)
	handler := refreshCensusRouter(t, &now)
	assertRefreshCensusResponse(t, handler, http.MethodGet, "/api/v1/session/refresh", "", http.StatusMethodNotAllowed, "invalid", "method")
	assertRefreshCensusResponse(t, handler, http.MethodPost, "/api/v1/session/refresh-missing", "", http.StatusNotFound, "unknown_id", "route")
}

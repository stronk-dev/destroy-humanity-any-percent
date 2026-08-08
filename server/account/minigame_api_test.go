package account

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"cloud-clicker/server/minigame"
	"cloud-clicker/server/production"

	"github.com/go-chi/chi/v5"
)

type minigameAPIStub struct {
	createAccount string
	createGame    string
	createKey     string
	playCommand   json.RawMessage
	result        json.RawMessage
	err           error
}

func (stub *minigameAPIStub) CreateMinigameSession(_ context.Context, accountID, minigameID, _, idempotencyKey string, _ time.Time) (json.RawMessage, error) {
	stub.createAccount, stub.createGame, stub.createKey = accountID, minigameID, idempotencyKey
	return stub.result, stub.err
}
func (stub *minigameAPIStub) PlayMinigameCommand(_ context.Context, _, _, _ string, _ int64, command json.RawMessage, _ time.Time) (json.RawMessage, error) {
	stub.playCommand = bytes.Clone(command)
	return stub.result, stub.err
}
func (stub *minigameAPIStub) CurrentMinigameSession(context.Context, string) (json.RawMessage, error) {
	return stub.result, stub.err
}
func (stub *minigameAPIStub) ResolveMinigameSession(context.Context, string, string) (json.RawMessage, error) {
	return stub.result, stub.err
}

func minigameAPIRequest(method, path, body string, parameters map[string]string) *http.Request {
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	route := chi.NewRouteContext()
	for key, value := range parameters {
		route.URLParams.Add(key, value)
	}
	ctx := context.WithValue(request.Context(), chi.RouteCtxKey, route)
	ctx = context.WithValue(ctx, claimsContextKey{}, Claims{Subject: "01985555-1111-7111-8111-111111111111"})
	return request.WithContext(ctx)
}

func TestTypedMinigameHandlersKeepIdentityAndTenantCommandOffTheFlatWire(t *testing.T) {
	now := time.Date(2026, 8, 7, 12, 0, 0, 0, time.UTC)
	stub := &minigameAPIStub{result: json.RawMessage(`{"ok":true}`)}
	api := &API{repository: &Repository{clock: func() time.Time { return now }, random: bytes.NewReader(bytes.Repeat([]byte{0xaa}, 32))},
		config: APIConfig{MaxBodyBytes: 64 << 10}, minigames: stub}

	response := httptest.NewRecorder()
	api.createMinigameSession(response, minigameAPIRequest(http.MethodPost, "/api/v1/minigames/pitch/sessions",
		`{"idempotency_key":"create-1"}`, map[string]string{"minigame_id": "pitch"}))
	if response.Code != http.StatusOK || stub.createAccount == "" || stub.createGame != "pitch" || stub.createKey != "create-1" {
		t.Fatalf("create status=%d body=%s stub=%+v", response.Code, response.Body.String(), stub)
	}

	response = httptest.NewRecorder()
	api.playMinigameCommand(response, minigameAPIRequest(http.MethodPost, "/api/v1/minigames/sessions/id/commands",
		`{"command_id":"command-1","expected_revision":1,"command":{"kind":"end_shop"}}`, map[string]string{"session_id": "01986666-ca01-7000-8000-000000000010"}))
	if response.Code != http.StatusOK || string(stub.playCommand) != `{"kind":"end_shop"}` {
		t.Fatalf("command status=%d body=%s command=%s", response.Code, response.Body.String(), stub.playCommand)
	}

	response = httptest.NewRecorder()
	api.createMinigameSession(response, minigameAPIRequest(http.MethodPost, "/api/v1/minigames/pitch/sessions",
		`{"idempotency_key":"create-2","founder_id":"client-controlled"}`, map[string]string{"minigame_id": "pitch"}))
	if response.Code != http.StatusBadRequest {
		t.Fatalf("client identity field accepted: %d %s", response.Code, response.Body.String())
	}

	response = httptest.NewRecorder()
	api.getCurrentMinigameSession(response, minigameAPIRequest(http.MethodGet, "/api/v1/minigames/sessions/current", `{}`, nil))
	if response.Code != http.StatusBadRequest {
		t.Fatalf("current accepted a request body: %d %s", response.Code, response.Body.String())
	}
	response = httptest.NewRecorder()
	api.getCurrentMinigameSession(response, minigameAPIRequest(http.MethodGet, "/api/v1/minigames/sessions/current", ``, nil))
	if response.Code != http.StatusOK {
		t.Fatalf("bodyless current failed: %d %s", response.Code, response.Body.String())
	}
	response = httptest.NewRecorder()
	api.resolveMinigameSession(response, minigameAPIRequest(http.MethodPost, "/api/v1/minigames/sessions/id/resolve", ``, map[string]string{"session_id": "id"}))
	if response.Code != http.StatusBadRequest {
		t.Fatalf("resolve accepted a missing object: %d %s", response.Code, response.Body.String())
	}
	response = httptest.NewRecorder()
	api.resolveMinigameSession(response, minigameAPIRequest(http.MethodPost, "/api/v1/minigames/sessions/id/resolve", `{}`, map[string]string{"session_id": "id"}))
	if response.Code != http.StatusOK {
		t.Fatalf("empty-object resolve failed: %d %s", response.Code, response.Body.String())
	}
}

func TestMinigameDeterministicErrorTableIsClosed(t *testing.T) {
	api := &API{}
	tests := []struct {
		name, action string
		err          error
		status       int
		pair         string
	}{
		{"tenant", "command", &minigame.Rejection{Code: "illegal_phase", Detail: "x"}, 409, `"category":"not_eligible","detail":"illegal_phase"`},
		{"unlock", "create", production.ErrMinigameFiscalUnlockRequired, 409, `"category":"not_eligible","detail":"fiscal_unlock_required"`},
		{"soul", "create", production.ErrMinigameHumanContentLocked, 409, `"category":"not_eligible","detail":"human_content_locked"`},
		{"exclusive", "create", minigame.ErrExclusiveActivity, 409, `"category":"not_eligible","detail":"exclusive_activity"`},
		{"create-idempotency", "create", minigame.ErrAPIIdempotency, 409, `"category":"idempotency_conflict","detail":"minigame_session"`},
		{"command-idempotency", "command", minigame.ErrAPIIdempotency, 409, `"category":"idempotency_conflict","detail":"minigame_command"`},
		{"gone", "resolve", minigame.ErrSessionGone, 404, `"category":"unknown_id","detail":"minigame_session"`},
		{"founder", "current", ErrFounderNotFound, 404, `"category":"unknown_id","detail":"founder"`},
		{"busy", "command", minigame.ErrSessionBusy, 409, `"category":"conflict","detail":"minigame_session"`},
		{"revision", "command", minigame.ErrSessionRevision, 409, `"category":"conflict","detail":"minigame_revision"`},
		{"invalid", "command", minigame.ErrInvalidSession, 400, `"category":"invalid","detail":"minigame_command"`},
		{"tenant-version", "create", minigame.ErrTenantVersion, 404, `"category":"unknown_id","detail":"minigame_tenant"`},
		{"store", "command", errors.New("database unavailable"), 500, `"category":"internal_invariant","detail":"minigame_api"`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			response := httptest.NewRecorder()
			api.writeMinigameResult(response, nil, test.err, test.action)
			if response.Code != test.status || !strings.Contains(response.Body.String(), test.pair) {
				t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
			}
		})
	}
}

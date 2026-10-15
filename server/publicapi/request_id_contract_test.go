package publicapi_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"cloud-clicker/server/account"
	"cloud-clicker/server/httpapi"
	"cloud-clicker/server/publicapi"
	"cloud-clicker/server/publicread"
)

const requestIDError = "{\"category\":\"internal_invariant\",\"detail\":\"request_id\"}\n"

func TestRequestIDFailureConformsToEveryRegisteredOperation(t *testing.T) {
	private, err := account.PrivateAPIRegistry()
	if err != nil {
		t.Fatal(err)
	}
	public, err := publicread.Registry()
	if err != nil {
		t.Fatal(err)
	}
	for _, registry := range []*publicapi.Registry{private, public} {
		for _, operation := range registry.Operations() {
			t.Run(operation.ID, func(t *testing.T) {
				if err := registry.ValidateResponse(operation.ID, 500, []byte(requestIDError)); err != nil {
					t.Fatalf("outer request-ID failure is undeclared: %v", err)
				}
			})
		}
	}
}

func TestPublicRequestIDFailureIsTypedBeforeHandler(t *testing.T) {
	data, err := os.ReadFile("../../balance/api/phase0.json")
	if err != nil {
		t.Fatal(err)
	}
	policy, err := publicapi.LoadPolicy(data)
	if err != nil {
		t.Fatal(err)
	}
	ids, err := httpapi.NewRequestIDs(policy.RequestID.Pattern, policy.RequestID.MaxBytes, bytes.NewReader(nil))
	if err != nil {
		t.Fatal(err)
	}
	runtime, err := publicapi.NewRuntime(policy, func() time.Time { return time.Date(2026, 10, 15, 12, 0, 0, 0, time.UTC) }, ids)
	if err != nil {
		t.Fatal(err)
	}
	called := false
	handler := runtime.WithRequestID(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { called = true }))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/public/v1/epochs", nil))
	if response.Code != 500 || response.Header().Get("Content-Type") != publicapi.ContentJSON ||
		response.Header().Get("Cache-Control") != "no-store" || response.Body.String() != requestIDError || called {
		t.Fatalf("failure status=%d headers=%v body=%q handler_called=%t", response.Code, response.Header(), response.Body.String(), called)
	}
	for _, header := range []string{"X-Request-ID", "ETag", "X-Content-SHA256"} {
		if response.Header().Get(header) != "" {
			t.Fatalf("failure fabricated %s", header)
		}
	}
}

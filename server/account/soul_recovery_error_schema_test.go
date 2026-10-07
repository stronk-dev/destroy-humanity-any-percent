package account

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"cloud-clicker/server/production"
	"cloud-clicker/server/publicapi"
)

func assertSoulErrorContract(t *testing.T, registry *publicapi.Registry, operation string, status int, body []byte, category, detail string) {
	t.Helper()
	want := []byte(fmt.Sprintf("{\"category\":%q,\"detail\":%q}\n", category, detail))
	if !bytes.Equal(body, want) {
		t.Fatalf("%s reply differs from exact %s/%s bytes (body withheld)", operation, category, detail)
	}
	if err := registry.ValidateResponse(operation, status, body); err != nil {
		t.Fatalf("%s descriptor rejects the actual %s/%s reply: %v", operation, category, detail, err)
	}
}

func TestSoulRecoveryErrorRegistryMatchesHandlers(t *testing.T) {
	registry, err := PrivateAPIRegistry()
	if err != nil {
		t.Fatal(err)
	}
	api := &API{}
	cases := []struct {
		operation, action string
		handler           http.HandlerFunc
	}{
		{"start_soul_recovery", "start", api.startSoulRecovery},
		{"progress_soul_recovery", "progress", api.progressSoulRecovery},
		{"resolve_soul_recovery", "resolve", api.resolveSoulRecovery},
		{"cancel_soul_recovery", "cancel", api.cancelSoulRecovery},
	}
	for _, row := range cases {
		t.Run(row.action+"/unavailable", func(t *testing.T) {
			response := httptest.NewRecorder()
			row.handler(response, httptest.NewRequest(http.MethodPost, "/", strings.NewReader("{}")))
			if response.Code != http.StatusServiceUnavailable || response.Header().Get("Content-Type") != "application/json" {
				t.Fatal("unavailable handler lost its 503 JSON response")
			}
			assertSoulErrorContract(t, registry, row.operation, response.Code, response.Body.Bytes(), "not_configured", "soul_recovery")
		})
		t.Run(row.action+"/internal-failure", func(t *testing.T) {
			response := httptest.NewRecorder()
			api.writeSoulRecoveryResult(response, production.HandleResult{}, errors.New("private fixture diagnostic"), row.action)
			if response.Code != http.StatusInternalServerError || response.Header().Get("Content-Type") != "application/json" {
				t.Fatal("internal mapper lost its 500 JSON response")
			}
			assertSoulErrorContract(t, registry, row.operation, response.Code, response.Body.Bytes(), "internal_invariant", "soul_recovery")
		})
	}
	if err := registry.ValidateResponse("start_soul_recovery", http.StatusNotFound, []byte(`{"category":"unknown_id","detail":"invented_company_detail"}`)); err == nil {
		t.Fatal("error descriptor accepted an unregistered detail")
	}
}

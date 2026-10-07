package account

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"cloud-clicker/server/publicapi"
)

func TestAccountLifecycleRegistryPinsExistingWire(t *testing.T) {
	registry, err := PrivateAPIRegistry()
	if err != nil {
		t.Fatal(err)
	}
	for _, row := range []struct {
		id, method, path, body string
		auth                   publicapi.AuthMode
		status                 int
	}{
		{"create_account", http.MethodPost, "/api/v1/account", `{"account_id":"01985555-1111-7111-8111-111111111110","created_at":"2026-10-08T00:00:00Z","recovery_code":"fixture"}`, publicapi.AuthNone, http.StatusCreated},
		{"create_founder", http.MethodPost, "/api/v1/founder", `{"id":"01985555-1111-7111-8111-111111111111","created_at":"2026-10-08T00:00:00Z","imported":false}`, publicapi.AuthAccessToken, http.StatusCreated},
		{"get_founder", http.MethodGet, "/api/v1/founder", `{"id":"01985555-1111-7111-8111-111111111111","created_at":"2026-10-08T00:00:00Z","display":{}}`, publicapi.AuthAccessToken, http.StatusOK},
	} {
		t.Run(row.id, func(t *testing.T) {
			operation, ok := registry.Operation(row.id)
			if !ok || operation.Method != row.method || operation.Path != row.path || operation.Auth != row.auth || operation.Public || operation.Surface != publicapi.SurfacePrivateV1 {
				t.Fatal("account lifecycle route is absent or misrepresented in the mounting authority")
			}
			if row.method == http.MethodPost {
				if operation.Request == "" || registry.ValidateRequest(row.id, []byte(`{}`)) != nil {
					t.Fatal("account lifecycle POST lacks its canonical empty-object request")
				}
				for _, body := range []string{`null`, `[]`, `{"email":"private"}`} {
					if registry.ValidateRequest(row.id, []byte(body)) == nil {
						t.Fatal("account lifecycle descriptor admitted a nonobject or undeclared input")
					}
				}
			} else if operation.Request != "" {
				t.Fatal("Founder read invented a request body")
			}
			if err := registry.ValidateResponse(row.id, row.status, []byte(row.body)); err != nil {
				t.Fatal(err)
			}
			for _, body := range []string{`{}`, row.body[:len(row.body)-1] + `,"email":"private"}`, `{"id":7,"created_at":false}`} {
				if registry.ValidateResponse(row.id, row.status, []byte(body)) == nil {
					t.Fatal("account lifecycle response admitted missing, private or mistyped fields")
				}
			}
		})
	}
	if registry.ValidateResponse("get_founder", http.StatusOK, []byte(`{"id":"01985555-1111-7111-8111-111111111111","created_at":"2026-10-08T00:00:00Z","display":{"email":"private"}}`)) == nil {
		t.Fatal("current empty display object became a free-form private-data schema")
	}
	for _, row := range []struct {
		id, category, detail string
		status               int
	}{
		{"create_account", "invalid", "body", 400},
		{"create_account", "rate_limited", "ip", 429},
		{"create_account", "internal_invariant", "account_create", 500},
		{"create_founder", "invalid", "body", 400},
		{"create_founder", "unauthorized", "access_token", 401},
		{"create_founder", "unknown_id", "account", 404},
		{"create_founder", "rate_limited", "account", 429},
		{"create_founder", "rate_limited", "ip", 429},
		{"create_founder", "internal_invariant", "founder_create", 500},
		{"get_founder", "unauthorized", "access_token", 401},
		{"get_founder", "unknown_id", "founder", 404},
		{"get_founder", "rate_limited", "account", 429},
		{"get_founder", "rate_limited", "ip", 429},
	} {
		if err := registry.ValidateResponse(row.id, row.status, exactAPIErrorJSON(apiErrorPair{row.category, row.detail})[0]); err != nil {
			t.Fatalf("%s %d %s/%s: %v", row.id, row.status, row.category, row.detail, err)
		}
		if registry.ValidateResponse(row.id, row.status, exactAPIErrorJSON(apiErrorPair{"unknown_id", "run"})[0]) == nil {
			t.Fatalf("%s accepted an unrelated schema-valid error pair", row.id)
		}
	}
}

func TestAccountLifecycleTimestampDescriptorMatchesExistingCodec(t *testing.T) {
	registry, err := PrivateAPIRegistry()
	if err != nil {
		t.Fatal(err)
	}
	for _, milliseconds := range []int{0, 100, 120, 123} {
		instant := time.Date(2026, 10, 8, 0, 0, 0, milliseconds*int(time.Millisecond), time.UTC)
		encoded, err := json.Marshal(instant)
		if err != nil {
			t.Fatal(err)
		}
		body := fmt.Sprintf(`{"account_id":"01985555-1111-7111-8111-111111111110","created_at":%s,"recovery_code":"fixture"}`, encoded)
		if err := registry.ValidateResponse("create_account", 201, []byte(body)); err != nil {
			t.Fatalf("actual time.Time JSON rejected at %d ms: %v", milliseconds, err)
		}
	}
	for _, timestamp := range []string{"not a timestamp", "2026-10-08T00:00:00.1234Z", "2026-10-08T00:00:00.000Z", "2026-10-08T00:00:00.120Z", "2026-10-08T00:00:00+00:00"} {
		body := fmt.Sprintf(`{"account_id":"01985555-1111-7111-8111-111111111110","created_at":%q,"recovery_code":"fixture"}`, timestamp)
		if registry.ValidateResponse("create_account", 201, []byte(body)) == nil {
			t.Fatalf("timestamp not emitted by the existing canonical codec was accepted: %s", timestamp)
		}
	}
	// Adding the codec cannot relax the existing fixed-width bootstrap contract.
	definitions, err := publicapi.ValidateSchemaDefinitions([]publicapi.NamedSchema{{Name: "FixedTimestamp", Schema: apiString("date-time-ms")}})
	if err != nil {
		t.Fatal(err)
	}
	if publicapi.ValidateJSON("FixedTimestamp", []byte(`"2026-10-08T00:00:00Z"`), definitions) == nil {
		t.Fatal("fixed-width millisecond timestamps were relaxed")
	}
}

func TestAccountCreationRetainsSharedUnauthenticatedLimiter(t *testing.T) {
	now := time.Date(2026, 10, 8, 0, 0, 0, 0, time.UTC)
	handler := refreshCensusRouter(t, &now)
	assertRefreshCensusResponse(t, handler, http.MethodPost, "/api/v1/account", "{", 400, "invalid", "body")
	assertRefreshCensusResponse(t, handler, http.MethodPost, "/api/v1/session", "{", 429, "rate_limited", "ip")
	now = now.Add(time.Minute)
	assertRefreshCensusResponse(t, handler, http.MethodPost, "/api/v1/session/refresh", "{", 400, "invalid", "body")
	assertRefreshCensusResponse(t, handler, http.MethodPost, "/api/v1/account", "{", 429, "rate_limited", "ip")
}

package publicapi

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"cloud-clicker/server/httpapi"
)

func phase0Policy(t *testing.T) Policy {
	t.Helper()
	data, err := os.ReadFile("../../balance/api/phase0.json")
	if err != nil {
		t.Fatal(err)
	}
	policy, err := LoadPolicy(data)
	if err != nil {
		t.Fatal(err)
	}
	return policy
}

func TestPhase0PolicyIsExactAndSecretsFailClosed(t *testing.T) {
	policy := phase0Policy(t)
	if policy.PublicLimiter != (LimiterPolicy{Burst: 60, MaxIPEntries: 65_536, RefillPerMinute: 300}) || policy.TrustedProxyHops != 1 {
		t.Fatalf("policy=%+v", policy)
	}
	for _, invalid := range []string{
		`{}`,
		`{"cache_max_age_seconds":{"boards":60,"catalogs_epochs":3600,"registry":300,"verification":31536000},"cursor_key_ids":{"current":"k1","previous":"k0"},"public_limiter":{"burst":60,"max_ip_entries":65536,"refill_per_minute":300},"request_id":{"max_bytes":64,"pattern":"^[A-Za-z0-9-]{1,64}$"},"schema_version":1,"trusted_proxy_hops":1,"unknown":true}`,
		`{"cache_max_age_seconds":{"boards":60,"boards":60,"catalogs_epochs":3600,"registry":300,"verification":31536000},"cursor_key_ids":{"current":"k1","previous":"k0"},"public_limiter":{"burst":60,"max_ip_entries":65536,"refill_per_minute":300},"request_id":{"max_bytes":64,"pattern":"^[A-Za-z0-9-]{1,64}$"},"schema_version":1,"trusted_proxy_hops":1}`,
	} {
		if _, err := LoadPolicy([]byte(invalid)); err == nil {
			t.Fatalf("invalid policy accepted: %s", invalid)
		}
	}
	registry := testRegistry(t)
	key := bytes.Repeat([]byte{1}, 32)
	if _, err := ResolveCursorCodec(policy, registry, func(id string) ([]byte, bool) { return key, id == "k1" }); err == nil {
		t.Fatal("startup accepted a missing previous cursor secret")
	}
	codec, err := ResolveCursorCodec(policy, registry, func(id string) ([]byte, bool) { return key, id == "k1" || id == "k0" })
	if err != nil || codec == nil {
		t.Fatalf("first-deploy shared secret rejected: %v", err)
	}
}

func TestCachedMiddlewareSkipsLimiterFor304AndEchoesRequestID(t *testing.T) {
	policy := phase0Policy(t)
	policy.PublicLimiter.Burst = 1
	policy.PublicLimiter.RefillPerMinute = 1
	now := time.Date(2026, 8, 7, 12, 0, 0, 0, time.UTC)
	ids, err := httpapi.NewRequestIDs(requestIDPatternLiteral, 64, bytes.NewReader(bytes.Repeat([]byte{0xbb}, 64)))
	if err != nil {
		t.Fatal(err)
	}
	runtime, err := NewRuntime(policy, func() time.Time { return now }, ids)
	if err != nil {
		t.Fatal(err)
	}
	body := []byte(`{"ok":true}`)
	handler := runtime.WithRequestID(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if err := runtime.WriteCached(response, request, CacheVerification, ContentJSON, body); err != nil {
			t.Error(err)
		}
	}))

	first := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/public/v1/runs/x/1/genesis", nil)
	request.RemoteAddr = "192.0.2.1:1234"
	request.Header.Set("X-Request-ID", "request-1")
	handler.ServeHTTP(first, request)
	if first.Code != http.StatusOK || first.Body.String() != string(body) || first.Header().Get("X-Request-ID") != "request-1" || first.Header().Get("Cache-Control") != "public,max-age=31536000,immutable" {
		t.Fatalf("first status=%d headers=%v body=%q", first.Code, first.Header(), first.Body.String())
	}

	for index := 0; index < 2; index++ {
		cached := httptest.NewRecorder()
		retry := httptest.NewRequest(http.MethodGet, request.URL.String(), nil)
		retry.RemoteAddr = request.RemoteAddr
		retry.Header.Set("If-None-Match", first.Header().Get("ETag"))
		handler.ServeHTTP(cached, retry)
		if cached.Code != http.StatusNotModified || cached.Body.Len() != 0 || cached.Header().Get("ETag") == "" || cached.Header().Get("X-Request-ID") == "" {
			t.Fatalf("cached %d status=%d headers=%v body=%q", index, cached.Code, cached.Header(), cached.Body.String())
		}
	}

	limited := httptest.NewRecorder()
	retry := httptest.NewRequest(http.MethodGet, request.URL.String(), nil)
	retry.RemoteAddr = request.RemoteAddr
	handler.ServeHTTP(limited, retry)
	if limited.Code != http.StatusTooManyRequests || !strings.Contains(limited.Body.String(), `"category":"rate_limited"`) {
		t.Fatalf("limited status=%d body=%q", limited.Code, limited.Body.String())
	}
}

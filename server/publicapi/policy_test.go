package publicapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"cloud-clicker/server/httpapi"
)

func TestCachedRawHashHeaderCannotOverwriteProtocolHeaders(t *testing.T) {
	ids, err := httpapi.NewRequestIDs(requestIDPatternLiteral, 64, bytes.NewReader(bytes.Repeat([]byte{0xbb}, 64)))
	if err != nil {
		t.Fatal(err)
	}
	runtime, err := NewRuntime(phase0Policy(t), time.Now, ids)
	if err != nil {
		t.Fatal(err)
	}
	for _, headers := range [][]string{{""}, {"bad header"}, {"ETag"}, {"cache-control"}, {"Content-Type"}, {"Content-Length"}, {"x-request-id"}, {"X-Hash", "X-Other-Hash"}} {
		t.Run(strings.Join(headers, ","), func(t *testing.T) {
			response := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, "/raw", nil)
			request.Header.Set("X-Request-ID", "reserved-header-test")
			runtime.WithRequestID(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
				if err := runtime.WriteCached(response, request, CacheVerification, ContentGzip, []byte("archive"), headers...); !errors.Is(err, ErrInvalidPolicy) {
					t.Fatalf("invalid hash header accepted: %v", err)
				}
			})).ServeHTTP(response, request)
			if response.Body.Len() != 0 || response.Header().Get("X-Request-ID") != "reserved-header-test" || len(response.Header()) != 1 {
				t.Fatalf("invalid hash header wrote body/protocol metadata: %v %q", response.Header(), response.Body.String())
			}
		})
	}
}

func TestPublicBudgetDiscardsDeniedBodiesAndEntityHeaders(t *testing.T) {
	for _, explicitStatus := range []bool{false, true} {
		t.Run(fmt.Sprint("explicit status=", explicitStatus), func(t *testing.T) {
			policy := phase0Policy(t)
			policy.PublicLimiter.Burst = 1
			ids, err := httpapi.NewRequestIDs(requestIDPatternLiteral, 64, bytes.NewReader(bytes.Repeat([]byte{0xbb}, 64)))
			if err != nil {
				t.Fatal(err)
			}
			now := time.Date(2026, 10, 15, 12, 0, 0, 0, time.UTC)
			runtime, err := NewRuntime(policy, func() time.Time { return now }, ids)
			if err != nil {
				t.Fatal(err)
			}
			handler := runtime.WithRequestID(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
				if request.URL.Path == "/seed" {
					if err := runtime.WriteCached(response, request, CacheCatalogEpoch, ContentJSON, []byte(`{"seed":true}`)); err != nil {
						t.Fatal(err)
					}
					return
				}
				response.Header().Set("Content-Type", ContentGzip)
				response.Header().Set("Content-Length", "999")
				response.Header().Set("Content-Encoding", "gzip")
				response.Header().Set("Cache-Control", "public,max-age=3600")
				response.Header().Set("ETag", `"denied"`)
				if explicitStatus {
					response.WriteHeader(http.StatusInternalServerError)
				}
				for _, body := range []string{"denied body", "second denied chunk"} {
					if n, err := response.Write([]byte(body)); n != len(body) || err != nil {
						t.Fatalf("denied write=%d err=%v", n, err)
					}
				}
			}))
			request := func(path string) *httptest.ResponseRecorder {
				r := httptest.NewRequest(http.MethodGet, path, nil)
				r.RemoteAddr = "192.0.2.1:4000"
				r.Header.Set("X-Request-ID", "response-budget-check")
				response := httptest.NewRecorder()
				handler.ServeHTTP(response, r)
				return response
			}
			if seed := request("/seed"); seed.Code != http.StatusOK {
				t.Fatalf("seed status=%d body=%q", seed.Code, seed.Body.String())
			}
			denied := request("/denied")
			if denied.Code != http.StatusTooManyRequests || denied.Body.String() != "{\"category\":\"rate_limited\",\"detail\":\"ip\"}\n" ||
				denied.Header().Get("Content-Type") != ContentJSON || denied.Header().Get("X-Request-ID") != "response-budget-check" {
				t.Fatalf("denied status=%d headers=%v body=%q", denied.Code, denied.Header(), denied.Body.String())
			}
			for _, header := range []string{"Content-Length", "Content-Encoding", "Cache-Control", "ETag"} {
				if denied.Header().Get(header) != "" {
					t.Fatalf("denied response retained %s: %v", header, denied.Header())
				}
			}
		})
	}
}

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

func TestPolicyRequiresEveryExactNonNullField(t *testing.T) {
	source, err := os.ReadFile("../../balance/api/phase0.json")
	if err != nil {
		t.Fatal(err)
	}
	var root map[string]json.RawMessage
	if err := json.Unmarshal(source, &root); err != nil {
		t.Fatal(err)
	}
	for _, scope := range []string{"", "cache_max_age_seconds", "cursor_key_ids", "public_limiter", "request_id"} {
		fields := root
		if scope != "" {
			fields = nil
			if err := json.Unmarshal(root[scope], &fields); err != nil {
				t.Fatal(err)
			}
		}
		for name := range fields {
			for _, change := range []string{"missing", "null", "case_replacement", "case_collision"} {
				t.Run(scope+"/"+name+"/"+change, func(t *testing.T) {
					var candidate map[string]json.RawMessage
					if err := json.Unmarshal(source, &candidate); err != nil {
						t.Fatal(err)
					}
					object := candidate
					if scope != "" {
						object = nil
						if err := json.Unmarshal(candidate[scope], &object); err != nil {
							t.Fatal(err)
						}
					}
					switch change {
					case "missing":
						delete(object, name)
					case "null":
						object[name] = json.RawMessage("null")
					case "case_replacement":
						object[strings.ToUpper(name)] = object[name]
						delete(object, name)
					case "case_collision":
						object[strings.ToUpper(name)] = object[name]
					}
					if scope != "" {
						encoded, err := json.Marshal(object)
						if err != nil {
							t.Fatal(err)
						}
						candidate[scope] = encoded
					}
					data, err := json.Marshal(candidate)
					if err != nil {
						t.Fatal(err)
					}
					if policy, err := LoadPolicy(data); !errors.Is(err, ErrInvalidPolicy) || policy != (Policy{}) {
						t.Fatalf("invalid configuration admitted or partially returned: err=%v", err)
					}
				})
			}
		}
	}
}

func TestPolicyPreservesExplicitZeroProxyAndEquivalentKeyEncoding(t *testing.T) {
	source, err := os.ReadFile("../../balance/api/phase0.json")
	if err != nil {
		t.Fatal(err)
	}
	want := phase0Policy(t)
	want.TrustedProxyHops = 0
	data := bytes.Replace(source, []byte(`"trusted_proxy_hops": 1`), []byte(`"tr\u0075sted_proxy_hops": 0`), 1)
	data = bytes.Replace(data, []byte(`"boards"`), []byte(`"bo\u0061rds"`), 1)
	policy, err := LoadPolicy(data)
	if err != nil || policy != want {
		t.Fatalf("valid explicit zero/escaped keys refused: err=%v", err)
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

	etag := first.Header().Get("ETag")
	for _, test := range []struct {
		name   string
		values []string
	}{
		{"strong", []string{etag}},
		{"weak", []string{"W/" + etag}},
		{"list matches last", []string{`"other", W/` + etag}},
		{"list matches first", []string{etag + `, "other"`}},
		{"repeated field", []string{`"other"`, etag}},
		{"quoted comma", []string{`"other,tag", ` + etag}},
		{"optional whitespace", []string{" \tW/" + etag + "\t "}},
		{"empty list members", []string{", , " + etag + ", ,"}},
		{"wildcard", []string{" *\t"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			cached := httptest.NewRecorder()
			retry := httptest.NewRequest(http.MethodGet, request.URL.String(), nil)
			retry.RemoteAddr = request.RemoteAddr
			retry.Header.Set("X-Request-ID", "conditional-cache-check")
			for _, value := range test.values {
				retry.Header.Add("If-None-Match", value)
			}
			handler.ServeHTTP(cached, retry)
			if cached.Code != http.StatusNotModified || cached.Body.Len() != 0 || cached.Header().Get("ETag") != etag ||
				cached.Header().Get("X-Request-ID") != "conditional-cache-check" || cached.Header().Get("Cache-Control") != first.Header().Get("Cache-Control") ||
				cached.Header().Get("Content-Type") != "" || cached.Header().Get("Content-Length") != "" {
				t.Fatalf("cached status=%d headers=%v body=%q", cached.Code, cached.Header(), cached.Body.String())
			}
		})
	}
	for _, test := range []struct {
		name   string
		values []string
	}{
		{"wrong tag", []string{`W/"other"`}},
		{"opaque tag is case sensitive", []string{strings.ToUpper(etag)}},
		{"lowercase weakness marker", []string{"w/" + etag}},
		{"space inside weakness marker", []string{"W/ " + etag}},
		{"match then invalid suffix", []string{etag + " garbage"}},
		{"match then invalid field", []string{etag, "garbage"}},
		{"unquoted tag", []string{strings.Trim(etag, `"`)}},
		{"missing list separator", []string{etag + ` "other"`}},
		{"unclosed quote", []string{etag + `, "other`}},
		{"wildcard with tag", []string{"*, " + etag}},
		{"tag with wildcard", []string{etag + ", *"}},
		{"repeated wildcard", []string{"*", "*"}},
		{"control in opaque tag", []string{etag + ", \"a\tb\""}},
		{"empty members only", []string{", ,"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			response := httptest.NewRecorder()
			retry := httptest.NewRequest(http.MethodGet, request.URL.String(), nil)
			retry.RemoteAddr = request.RemoteAddr
			retry.Header.Set("X-Request-ID", "conditional-cache-check")
			for _, value := range test.values {
				retry.Header.Add("If-None-Match", value)
			}
			handler.ServeHTTP(response, retry)
			if response.Code != http.StatusTooManyRequests || response.Header().Get("ETag") != "" || response.Header().Get("Cache-Control") != "" {
				t.Fatalf("invalid/nonmatching condition bypassed exhausted bucket: %d %v %q", response.Code, response.Header(), response.Body.String())
			}
		})
	}

	limited := httptest.NewRecorder()
	retry := httptest.NewRequest(http.MethodGet, request.URL.String(), nil)
	retry.RemoteAddr = request.RemoteAddr
	handler.ServeHTTP(limited, retry)
	if limited.Code != http.StatusTooManyRequests || !strings.Contains(limited.Body.String(), `"category":"rate_limited"`) {
		t.Fatalf("limited status=%d body=%q", limited.Code, limited.Body.String())
	}
}

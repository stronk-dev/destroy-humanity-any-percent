package httpapi

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestSharedTokenBucketsClientIPAndRequestIDs(t *testing.T) {
	now := time.Date(2026, 8, 7, 12, 0, 0, 123_000_000, time.UTC)
	buckets, err := NewTokenBuckets(1, 60, 2)
	if err != nil || !buckets.Allow("client", now) || buckets.Allow("client", now.Add(-time.Hour)) || !buckets.Allow("client", now.Add(time.Second)) {
		t.Fatalf("bucket behavior err=%v", err)
	}
	if !buckets.Allow("second", now.Add(2*time.Second)) || !buckets.Allow("third", now.Add(2*time.Second)) || buckets.Len() != 2 || buckets.Contains("client") {
		t.Fatal("bounded LRU did not evict the oldest bucket")
	}

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.RemoteAddr = "10.0.0.5:443"
	request.Header.Set("X-Forwarded-For", "203.0.113.8, 10.0.0.4")
	if got := ClientIP(request, 2); got != "203.0.113.8" {
		t.Fatalf("trusted client IP=%q", got)
	}
	if got := ClientIP(request, 0); got != "10.0.0.5" {
		t.Fatalf("direct client IP=%q", got)
	}

	ids, err := NewRequestIDs(`^[A-Za-z0-9-]{1,64}$`, 64, bytes.NewReader(bytes.Repeat([]byte{0xaa}, 32)))
	if err != nil {
		t.Fatal(err)
	}
	if got, err := ids.Resolve("caller-9", now); err != nil || got != "caller-9" {
		t.Fatalf("accepted request ID=%q err=%v", got, err)
	}
	generated, err := ids.Resolve("bad id", now)
	if err != nil || generated != "019fdc18-2e7b-7aaa-aaaa-aaaaaaaaaaaa" {
		t.Fatalf("generated UUIDv7=%q err=%v", generated, err)
	}
}

func TestRequestIDMiddlewareEchoesIdentityOnSuccessAndErrors(t *testing.T) {
	now := time.Date(2026, 10, 15, 12, 0, 0, 0, time.UTC)
	ids, err := NewRequestIDs(`^[A-Za-z0-9-]{1,64}$`, 64, bytes.NewReader(nil))
	if err != nil {
		t.Fatal(err)
	}
	for _, status := range []int{200, 400, 401, 404, 429, 500} {
		request := httptest.NewRequest(http.MethodGet, "/api/v1/example", nil)
		request.Header.Set("X-Request-ID", strings.Repeat("x", 64))
		response := httptest.NewRecorder()
		ids.WithRequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if RequestID(r) != request.Header.Get("X-Request-ID") {
				t.Fatal("handler did not receive the resolved identity")
			}
			w.WriteHeader(status)
			_, _ = w.Write([]byte("original body"))
		}), func() time.Time { return now }).ServeHTTP(response, request)
		if response.Code != status || response.Body.String() != "original body" || response.Header().Get("X-Request-ID") != request.Header.Get("X-Request-ID") {
			t.Fatalf("status=%d response=%d headers=%v body=%q", status, response.Code, response.Header(), response.Body.String())
		}
	}
	if RequestID(nil) != "" || RequestID(httptest.NewRequest(http.MethodGet, "/", nil)) != "" {
		t.Fatal("raw requests must not acquire a resolved identity")
	}
}

func TestNestedRequestIDMiddlewareGeneratesOnceWithoutTruncation(t *testing.T) {
	now := time.Date(2026, 10, 15, 12, 0, 0, 0, time.UTC)
	for _, candidate := range []string{"", "invalid id", strings.Repeat("x", 65)} {
		// Exactly one UUID's entropy. The nested middleware must not draw again,
		// even if another handler changes the raw caller header between mounts.
		ids, err := NewRequestIDs(`^[A-Za-z0-9-]{1,64}$`, 64, bytes.NewReader(bytes.Repeat([]byte{0xaa}, 16)))
		if err != nil {
			t.Fatal(err)
		}
		var outerID string
		inner := ids.WithRequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if RequestID(r) == "" || RequestID(r) != outerID {
				t.Fatal("nested mount changed the resolved identity")
			}
			w.WriteHeader(http.StatusNoContent)
		}), func() time.Time { return now })
		outer := ids.WithRequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			outerID = RequestID(r)
			r.Header.Set("X-Request-ID", "changed raw header")
			inner.ServeHTTP(w, r)
		}), func() time.Time { return now })
		request := httptest.NewRequest(http.MethodGet, "/", nil)
		request.Header.Set("X-Request-ID", candidate)
		response := httptest.NewRecorder()
		outer.ServeHTTP(response, request)
		if response.Code != http.StatusNoContent || len(outerID) != 36 || outerID[14] != '7' || response.Header().Get("X-Request-ID") != outerID {
			t.Fatalf("generated identity status=%d identity=%q", response.Code, outerID)
		}
	}
}

func TestRequestIDMiddlewareRefusesEntropyFailureBeforeHandler(t *testing.T) {
	ids, err := NewRequestIDs(`^[A-Za-z0-9-]{1,64}$`, 64, bytes.NewReader(nil))
	if err != nil {
		t.Fatal(err)
	}
	called := false
	handler := ids.WithRequestID(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		called = true
	}), func() time.Time { return time.Date(2026, 10, 15, 12, 0, 0, 0, time.UTC) })
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/", nil))
	if response.Code != http.StatusInternalServerError || response.Header().Get("X-Request-ID") != "" || called ||
		response.Header().Get("Content-Type") != "text/plain; charset=utf-8" || response.Body.String() != "internal invariant\n" {
		t.Fatalf("entropy failure status=%d handler_called=%t", response.Code, called)
	}
}

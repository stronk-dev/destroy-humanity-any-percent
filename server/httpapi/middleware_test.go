package httpapi

import (
	"bytes"
	"net/http"
	"net/http/httptest"
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

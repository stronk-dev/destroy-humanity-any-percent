package httpapi

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestRequestLogUsesOnlyResolvedIdentityForEveryResponseClass(t *testing.T) {
	for _, status := range []int{200, 204, 304, 400, 401, 404, 429, 500} {
		for _, candidate := range []string{strings.Repeat("x", 64), "", "invalid id", strings.Repeat("x", 65)} {
			ids, err := NewRequestIDs(`^[A-Za-z0-9-]{1,64}$`, 64, bytes.NewReader(bytes.Repeat([]byte{0xaa}, 16)))
			if err != nil {
				t.Fatal(err)
			}
			var logs bytes.Buffer
			logger := slog.New(slog.NewJSONHandler(&logs, nil))
			called := 0
			next := LogRequestID(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				called++
				w.WriteHeader(status)
				if status != 204 && status != 304 {
					_, _ = w.Write([]byte("original response"))
				}
			}), logger)
			handler := ids.WithRequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Downstream raw-header mutation cannot replace the context's ID.
				r.Header.Set("X-Request-ID", "must-not-log-replaced-header")
				next.ServeHTTP(w, r)
			}), func() time.Time { return time.Date(2026, 10, 15, 12, 0, 0, 0, time.UTC) })
			request := httptest.NewRequest(http.MethodGet, "/private/account-id?token=must-not-log-query", strings.NewReader("must-not-log-body"))
			request.RemoteAddr = "192.0.2.123:1234"
			request.Header.Set("Authorization", "Bearer must-not-log-token")
			request.Header.Set("Cookie", "must-not-log-cookie")
			request.Header.Set("X-Request-ID", candidate)
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)
			var record map[string]any
			if err := json.Unmarshal(logs.Bytes(), &record); err != nil {
				t.Fatalf("expected exactly one structured request entry: %v", err)
			}
			if len(record) != 4 || record["time"] == nil || record["level"] != "INFO" ||
				record["msg"] != "api request" || record["request_id"] != response.Header().Get("X-Request-ID") {
				t.Fatal("request correlation logs more than the resolved ID or uses the wrong authority")
			}
			if called != 1 || response.Code != status || ((status == 204 || status == 304) && response.Body.Len() != 0) ||
				((status != 204 && status != 304) && response.Body.String() != "original response") {
				t.Fatal("logging changed the original handler response")
			}
		}
	}
}

func TestRequestLogDoesNotInventAnIdentityOrRequireAnEmbeddedSink(t *testing.T) {
	var logs bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logs, nil))
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(204) })
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Header.Set("X-Request-ID", "raw-header-is-not-resolved")
	response := httptest.NewRecorder()
	LogRequestID(next, logger).ServeHTTP(response, request)
	if logs.Len() != 0 || response.Code != 204 {
		t.Fatal("unresolved request fabricated an identity or changed the response")
	}
	ids, err := NewRequestIDs(`^[A-Za-z0-9-]{1,64}$`, 64, bytes.NewReader(nil))
	if err != nil {
		t.Fatal(err)
	}
	clock := func() time.Time { return time.Date(2026, 10, 15, 12, 0, 0, 0, time.UTC) }
	request.Header.Set("X-Request-ID", "")
	response = httptest.NewRecorder()
	ids.WithRequestID(LogRequestID(next, logger), clock).ServeHTTP(response, request)
	if response.Code != 500 || logs.Len() != 0 {
		t.Fatal("failed identity assignment fabricated a request log")
	}
	request.Header.Set("X-Request-ID", "resolved-without-sink")
	response = httptest.NewRecorder()
	ids.WithRequestID(LogRequestID(next, nil), clock).ServeHTTP(response, request)
	if response.Code != 204 || response.Header().Get("X-Request-ID") != "resolved-without-sink" {
		t.Fatal("explicitly omitted sink changed embedded HTTP behavior")
	}
}

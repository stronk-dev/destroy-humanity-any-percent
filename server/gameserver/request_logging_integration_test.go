package gameserver

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"cloud-clicker/server/account"
)

// The real JSON handler and background jobs share this writer. Never print its
// contents: existing non-request diagnostics have their own privacy boundaries.
type requestLogBuffer struct {
	mu   sync.Mutex
	data bytes.Buffer
}

func (buffer *requestLogBuffer) Write(data []byte) (int, error) {
	buffer.mu.Lock()
	defer buffer.mu.Unlock()
	return buffer.data.Write(data)
}

func (buffer *requestLogBuffer) count(t *testing.T, id string) int {
	t.Helper()
	buffer.mu.Lock()
	data := bytes.Clone(buffer.data.Bytes())
	buffer.mu.Unlock()
	decoder := json.NewDecoder(bytes.NewReader(data))
	count := 0
	for {
		var entry map[string]json.RawMessage
		if err := decoder.Decode(&entry); err == io.EOF {
			break
		} else if err != nil {
			t.Fatal("composed logger emitted invalid JSON")
		}
		var message string
		_ = json.Unmarshal(entry["msg"], &message)
		if message != "api request" {
			continue
		}
		// No headers, body, actor IDs, IP, URL/query or error text may sneak
		// into this bounded correlation entry.
		if len(entry) != 4 || entry["time"] == nil || entry["level"] == nil {
			t.Fatal("request log contains unexpected fields")
		}
		var loggedID string
		if json.Unmarshal(entry["request_id"], &loggedID) != nil || loggedID == "" {
			t.Fatal("request log lacks its resolved ID")
		}
		if loggedID == id {
			count++
		}
	}
	return count
}

func assertComposedGeneratedRequestIDLogs(t *testing.T, server *httptest.Server, logs *requestLogBuffer, tokens account.TokenPair) {
	t.Helper()
	for _, path := range []string{"/api/v1/founder", "/api/public/v1/epochs"} {
		for _, candidate := range []string{"", "invalid id", strings.Repeat("x", 65)} {
			request, err := http.NewRequest(http.MethodGet, server.URL+path+"?private_marker=must-not-log-query", strings.NewReader("must-not-log-body"))
			if err != nil {
				t.Fatal(err)
			}
			request.Header.Set("X-Request-ID", candidate)
			request.Header.Set("Authorization", "Bearer "+tokens.AccessToken)
			request.Header.Set("Cookie", "private=must-not-log-cookie")
			request.Header.Set("X-Forwarded-For", "198.51.100.73")
			response, err := server.Client().Do(request)
			if err != nil {
				t.Fatal(err)
			}
			_ = readCompositionBytes(t, response)
			id := response.Header.Get("X-Request-ID")
			logged := logs.count(t, id)
			if response.StatusCode != http.StatusOK || len(id) != 36 || id[14] != '7' || logged != 1 {
				t.Fatalf("generated request-ID logging failed: surface=%s candidate_bytes=%d status=%d logged=%d", path, len(candidate), response.StatusCode, logged)
			}
		}
	}
}

func assertComposedDrainRequestIDs(t *testing.T, db *sql.DB, server *httptest.Server, logs *requestLogBuffer, entropy *composedRequestIDEntropy, token, intent string) {
	t.Helper()
	before := composedAdmissionRows(t, db)
	for _, test := range []struct{ name, candidate string }{
		{"valid", strings.Repeat("d", 64)}, {"missing", ""}, {"invalid", "invalid id"}, {"overlong", strings.Repeat("d", 65)},
	} {
		t.Run(test.name, func(t *testing.T) {
			request, err := http.NewRequest(http.MethodPost, server.URL+"/api/v1/intents", strings.NewReader(intent))
			if err != nil {
				t.Fatal(err)
			}
			request.Header.Set("Content-Type", "application/json")
			request.Header.Set("Authorization", "Bearer "+token)
			request.Header.Set("X-Request-ID", test.candidate)
			response, err := server.Client().Do(request)
			if err != nil {
				t.Fatal(err)
			}
			body := readCompositionBytes(t, response)
			if response.StatusCode != 503 || string(body) != "{\"category\":\"server_draining\",\"detail\":\"retry_same_intent_id\"}\n" {
				t.Fatalf("drain lost its existing refusal: status=%d", response.StatusCode)
			}
			id := response.Header.Get("X-Request-ID")
			valid := id == test.candidate && test.name == "valid"
			if test.name != "valid" {
				valid = len(id) == 36 && id[14] == '7'
			}
			if !valid || logs.count(t, id) != 1 {
				t.Fatal("drain bypassed resolved request-ID echo/logging")
			}
		})
	}
	t.Run("entropy_failure", func(t *testing.T) {
		entropy.setFailure(true)
		defer entropy.setFailure(false)
		response := compositionRequest(t, server.Client(), http.MethodPost, server.URL+"/api/v1/intents", token, intent)
		body := readCompositionBytes(t, response)
		if response.StatusCode != 500 || string(body) != "{\"category\":\"internal_invariant\",\"detail\":\"request_id\"}\n" || response.Header.Get("X-Request-ID") != "" {
			t.Fatal("failed identity assignment reached drain admission or fabricated an ID")
		}
	})
	if !bytes.Equal(before, composedAdmissionRows(t, db)) {
		t.Fatal("unadmitted draining requests changed persisted application data")
	}
	// The existing stop-intent contract does not stop ordinary reads.
	for _, path := range []string{"/api/v1/founder", "/api/public/v1/epochs"} {
		response := compositionRequest(t, server.Client(), http.MethodGet, server.URL+path, token, "")
		_ = readCompositionBytes(t, response)
		if response.StatusCode != 200 || logs.count(t, response.Header.Get("X-Request-ID")) != 1 {
			t.Fatalf("drain changed a permitted read: status=%d", response.StatusCode)
		}
	}
}

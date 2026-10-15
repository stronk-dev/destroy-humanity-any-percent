package gameserver

import (
	"bytes"
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

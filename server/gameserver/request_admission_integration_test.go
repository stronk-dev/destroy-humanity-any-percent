package gameserver

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"

	"cloud-clicker/server/account"
)

// Reuse the composed lifecycle's live account/session/activity. No copied
// handler, engine stub, artificial result or test-only gameplay endpoint.
func assertComposedRequestAdmission(t *testing.T, db *sql.DB, server *httptest.Server, operationID, path, token, valid, field, detail string, extra ...string) {
	t.Helper()
	registry, err := account.PrivateAPIRegistry()
	if err != nil || registry.ValidateRequest(operationID, []byte(valid)) != nil {
		t.Fatal("admission control does not match the registered request")
	}
	cases := map[string]string{"null-root": "null"}
	if field != "" {
		var fields map[string]json.RawMessage
		if json.Unmarshal([]byte(valid), &fields) != nil || fields[field] == nil {
			t.Fatal("admission fixture has no selected field")
		}
		prefix := valid[:len(valid)-1] + ","
		cases = map[string]string{
			"duplicate":         prefix + fmt.Sprintf(`%q:%s}`, field, fields[field]),
			"escaped-duplicate": prefix + fmt.Sprintf(`"\u%04x%s":%s}`, field[0], field[1:], fields[field]),
			"case-alias":        prefix + fmt.Sprintf(`%q:%s}`, strings.ToUpper(field), fields[field]),
			"unknown-field":     prefix + `"unexpected":true}`,
		}
	}
	for index, body := range extra {
		cases[fmt.Sprintf("nested-%d", index)] = body
	}
	// Stable order includes every supplied control, including nested commands.
	names := make([]string, 0, len(cases))
	for name := range cases {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		body := cases[name]
		t.Run(operationID+"/"+name, func(t *testing.T) {
			if registry.ValidateRequest(operationID, []byte(body)) == nil {
				t.Fatal("invalid admission fixture unexpectedly matches the descriptor")
			}
			before := composedAdmissionRows(t, db)
			response := compositionRequest(t, server.Client(), http.MethodPost, server.URL+path, token, body)
			data := readCompositionBytes(t, response)
			if response.StatusCode != http.StatusBadRequest || response.Header.Get("Content-Type") != "application/json" ||
				string(data) != fmt.Sprintf("{\"category\":\"invalid\",\"detail\":%q}\n", detail) ||
				registry.ValidateResponse(operationID, response.StatusCode, data) != nil {
				t.Errorf("schema-invalid request lost exact invalid/%s refusal (status=%d; body withheld)", detail, response.StatusCode)
			}
			if !bytes.Equal(before, composedAdmissionRows(t, db)) {
				t.Fatal("schema-invalid request changed credential/activity/gameplay/history rows (contents withheld)")
			}
		})
	}
}

func composedAdmissionRows(t *testing.T, db *sql.DB) []byte {
	t.Helper()
	// Closed test-owned names; snapshots contain private data and stay in memory.
	var result bytes.Buffer
	for _, table := range []string{
		"accounts", "account_founders", "session_families", "sessions", "access_tokens", "bootstrap_receipts",
		"save_streams", "save_revisions", "events", "intent_records", "founder_log", "run_log",
		"founder_genesis", "run_genesis", "run_frozen_contributions", "minigame_sessions",
		"minigame_session_commands", "minigame_create_receipts", "minigame_command_receipts",
		"minigame_faucet_window", "soul_recovery_sessions",
	} {
		var rows []byte
		if err := db.QueryRow(`SELECT COALESCE(jsonb_agg(to_jsonb(r) ORDER BY to_jsonb(r)::text),'[]'::jsonb) FROM ` + table + ` r`).Scan(&rows); err != nil {
			t.Fatal(err)
		}
		result.WriteString(table)
		result.WriteByte('\n')
		result.Write(rows)
		result.WriteByte('\n')
	}
	return result.Bytes()
}

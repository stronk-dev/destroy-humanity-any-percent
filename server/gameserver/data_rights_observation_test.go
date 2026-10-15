package gameserver

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"cloud-clicker/server/account"
)

type rightsRowObservation struct {
	count  int
	digest [32]byte
}

// Diagnostic classification, not an assertion that today's retention is right.
func rightsRowDisposition(before, after rightsRowObservation) string {
	switch {
	case after.count == 0:
		return "removed"
	case after.count < before.count:
		return "partially_removed"
	case after.count > before.count:
		return "expanded"
	case after.digest == before.digest:
		return "retained_identical"
	default:
		return "retained_modified"
	}
}

func TestRightsRowObservationDistinguishesErasureFromModification(t *testing.T) {
	before := rightsRowObservation{count: 2, digest: sha256.Sum256([]byte("before"))}
	for _, test := range []struct {
		after rightsRowObservation
		want  string
	}{
		{rightsRowObservation{}, "removed"},
		{rightsRowObservation{count: 1, digest: before.digest}, "partially_removed"},
		{rightsRowObservation{count: 3, digest: before.digest}, "expanded"},
		{before, "retained_identical"},
		{rightsRowObservation{count: 2, digest: sha256.Sum256([]byte("changed"))}, "retained_modified"},
	} {
		if got := rightsRowDisposition(before, test.after); got != test.want {
			t.Fatalf("disposition=%s want=%s", got, test.want)
		}
	}
}

// RP-122: actual terminal producers, current backend deletion and revoked access.
// All payloads stay in memory. Only counts/dispositions enter logs; SQL row JSON
// equality is not an original HTTP-byte proof or a complete field/privacy map.
func observeComposedSessionAccountDeletion(t *testing.T, db *sql.DB, server *httptest.Server, accountID, founderID, sessionID string, tokens account.TokenPair) {
	t.Helper()
	ctx := context.Background()
	type family struct{ name, predicate, order, id string }
	families := []family{
		{"minigame_sessions", "session_id=$1", "session_id", sessionID},
		{"minigame_session_commands", "session_id=$1", "seq", sessionID},
		{"minigame_create_receipts", "session_id=$1", "idempotency_key", sessionID},
		{"minigame_command_receipts", "session_id=$1", "command_id", sessionID},
		{"minigame_faucet_window", "founder_id=$1 AND minigame_id='pitch'", "attended_day", founderID},
		{"soul_recovery_sessions", "founder_id=$1", "session_id", founderID},
	}
	read := func(f family, id string) rightsRowObservation {
		t.Helper()
		// Identifiers/predicates are the closed test-owned table set above.
		query := fmt.Sprintf("SELECT COALESCE(jsonb_agg(to_jsonb(r) ORDER BY %s),'[]'::jsonb)::text FROM %s r WHERE %s", f.order, f.name, f.predicate)
		var text string
		if err := db.QueryRowContext(ctx, query, id).Scan(&text); err != nil {
			t.Fatalf("rights observation %s: %v", f.name, err)
		}
		var rows []json.RawMessage
		if err := json.Unmarshal([]byte(text), &rows); err != nil {
			t.Fatalf("rights observation %s is not a row population", f.name)
		}
		return rightsRowObservation{count: len(rows), digest: sha256.Sum256([]byte(text))}
	}
	before := make([]rightsRowObservation, len(families))
	terminalTokens := func() int {
		t.Helper()
		var count int
		if err := db.QueryRowContext(ctx, `SELECT count(*) FROM soul_recovery_sessions WHERE founder_id=$1 AND status IN ('resolved','cancelled') AND progress_token IS NOT NULL`, founderID).Scan(&count); err != nil {
			t.Fatal(err)
		}
		return count
	}
	for index, f := range families {
		before[index] = read(f, f.id)
		if before[index].count == 0 {
			t.Fatalf("invalid rights population: real producer left %s empty", f.name)
		}
		if missing := read(f, "01986666-ffff-7000-8000-ffffffffffff"); missing.count != 0 {
			t.Fatalf("rights observation %s ignores the selected identity", f.name)
		}
	}
	tokensBefore := terminalTokens()
	live := compositionRequest(t, server.Client(), http.MethodGet, server.URL+"/api/v1/minigames/sessions/current", tokens.AccessToken, "")
	_ = readCompositionBytes(t, live)
	if live.StatusCode != http.StatusOK {
		t.Fatal("deletion probe begins with invalid/expired access")
	}
	deleted := compositionRequest(t, server.Client(), http.MethodDelete, server.URL+"/api/v1/account", tokens.AccessToken, "")
	_ = readCompositionBytes(t, deleted)
	t.Logf("RP-122 produced delete status=%d", deleted.StatusCode)
	if deleted.StatusCode != http.StatusNoContent {
		t.Fatal("actual account deletion failed")
	}
	for index, f := range families {
		after := read(f, f.id)
		t.Logf("RP-122 produced family=%s before=%d after=%d disposition=%s", f.name, before[index].count, after.count, rightsRowDisposition(before[index], after))
	}
	t.Logf("RP-122 produced terminal_soul_tokens before=%d after=%d", tokensBefore, terminalTokens())
	var remaining int
	if err := db.QueryRowContext(ctx, "SELECT count(*) FROM accounts WHERE account_id=$1", accountID).Scan(&remaining); err != nil || remaining != 0 {
		t.Fatal("account deletion did not remove the account row")
	}
	var unlinked, archived bool
	err := db.QueryRowContext(ctx, "SELECT account_id IS NULL,archived_at IS NOT NULL FROM account_founders WHERE founder_id=$1", founderID).Scan(&unlinked, &archived)
	if err != nil && err != sql.ErrNoRows {
		t.Fatal(err)
	}
	t.Logf("RP-122 produced founder_present=%t unlinked=%t archived=%t", err == nil, unlinked, archived)
	for _, path := range []string{"/api/v1/minigames/sessions/current", "/api/v1/founder"} {
		refused := compositionRequest(t, server.Client(), http.MethodGet, server.URL+path, tokens.AccessToken, "")
		_ = readCompositionBytes(t, refused)
		if refused.StatusCode != http.StatusUnauthorized {
			t.Fatalf("deleted access still reaches %s: status=%d", path, refused.StatusCode)
		}
	}
	refresh := compositionRequest(t, server.Client(), http.MethodPost, server.URL+"/api/v1/session/refresh", "", fmt.Sprintf(`{"refresh_token":%q}`, tokens.RefreshToken))
	_ = readCompositionBytes(t, refresh)
	if refresh.StatusCode != http.StatusUnauthorized {
		t.Fatalf("deleted refresh remains usable: status=%d", refresh.StatusCode)
	}
}

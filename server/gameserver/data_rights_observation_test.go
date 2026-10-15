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
		if before.count == 0 {
			return "absent"
		}
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
	if got := rightsRowDisposition(rightsRowObservation{}, rightsRowObservation{}); got != "absent" {
		t.Fatalf("empty unproduced family falsely labelled %s", got)
	}
}

// RP-122: actual producers, current backend deletion and revoked access.
// All payloads stay in memory. Only counts/dispositions enter logs; SQL row JSON
// equality is not an original HTTP-byte proof or a complete field/privacy map.
func observeComposedSessionAccountDeletion(t *testing.T, db *sql.DB, server *httptest.Server, accountID, founderID, sessionID string, tokens account.TokenPair, active bool) {
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
		mayBeAbsent := active && (f.name == "minigame_faucet_window" || f.name == "soul_recovery_sessions")
		if before[index].count == 0 && !mayBeAbsent {
			t.Fatalf("invalid rights population: real producer left %s empty", f.name)
		}
		if missing := read(f, "01986666-ffff-7000-8000-ffffffffffff"); missing.count != 0 {
			t.Fatalf("rights observation %s ignores the selected identity", f.name)
		}
	}
	tokensBefore := terminalTokens()
	live := compositionRequest(t, server.Client(), http.MethodGet, server.URL+"/api/v1/minigames/sessions/current", tokens.AccessToken, "")
	liveBytes := readCompositionBytes(t, live)
	if live.StatusCode != http.StatusOK {
		t.Fatal("deletion probe begins with invalid/expired access")
	}
	phase := "terminal"
	if active {
		phase = "active"
		var current struct {
			Kind    string `json:"kind"`
			Session struct {
				SessionID string `json:"session_id"`
				Status    string `json:"status"`
			} `json:"session"`
		}
		if json.Unmarshal(liveBytes, &current) != nil || current.Kind != "active" || current.Session.SessionID != sessionID || current.Session.Status != "active" {
			t.Fatal("active deletion observation has no matching live active session")
		}
	}
	deleted := compositionRequest(t, server.Client(), http.MethodDelete, server.URL+"/api/v1/account", tokens.AccessToken, "")
	_ = readCompositionBytes(t, deleted)
	t.Logf("RP-122 produced phase=%s delete status=%d", phase, deleted.StatusCode)
	if deleted.StatusCode != http.StatusNoContent {
		t.Fatal("actual account deletion failed")
	}
	for index, f := range families {
		after := read(f, f.id)
		t.Logf("RP-122 produced phase=%s family=%s before=%d after=%d disposition=%s", phase, f.name, before[index].count, after.count, rightsRowDisposition(before[index], after))
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

// RP-120: capture the population before deletion. A policy that removes a nested
// reference must not also hide that surviving row from the after-denominator.
func observeComposedGuildConsumerDeletion(t *testing.T, db *sql.DB, server *httptest.Server, guildID, producerID string, consumer account.CreatedAccount) {
	t.Helper()
	ctx := context.Background()
	session := compositionRequest(t, server.Client(), http.MethodPost, server.URL+"/api/v1/session", "",
		fmt.Sprintf(`{"account_id":%q,"recovery_code":%q}`, consumer.AccountID, consumer.RecoveryCode))
	if session.StatusCode != http.StatusOK {
		_ = readCompositionBytes(t, session)
		t.Fatalf("consumer session refresh status=%d", session.StatusCode)
	}
	var tokens account.TokenPair
	decodeCompositionResponse(t, session, &tokens)
	type family struct{ name, key, allocations, producerPredicate string }
	families := []family{
		{"guild_clearing_results", "boundary_seq", "allocations", "account_id::text=$2"},
		{"guild_events", "event_seq", "payload->'allocations'", "kind='exchange_cleared' AND payload->>'producer_account_id'=$2"},
	}
	type observation struct {
		rows       rightsRowObservation
		references int
	}
	read := func(f family, keys []int64) observation {
		t.Helper()
		query := fmt.Sprintf(`SELECT COALESCE(jsonb_agg(to_jsonb(r) ORDER BY %[1]s),'[]'::jsonb)::text,
		COALESCE(sum((SELECT count(*) FROM jsonb_array_elements(%[2]s) a WHERE a->>'account_id'=$3)),0)
		FROM %[3]s r WHERE guild_id=$1 AND %[1]s=ANY($2::bigint[])`, f.key, f.allocations, f.name)
		args := []any{guildID, keys, consumer.AccountID}
		if f.name == "guild_clearing_results" {
			// Boundary sequence is unique only with guild+account. Event_seq,
			// unlike this composite key, is already a global immutable identity.
			query += " AND account_id::text=$4"
			args = append(args, producerID)
		}
		var text string
		var result observation
		if err := db.QueryRowContext(ctx, query, args...).Scan(&text, &result.references); err != nil {
			t.Fatal(err)
		}
		var rows []json.RawMessage
		if json.Unmarshal([]byte(text), &rows) != nil {
			t.Fatal("shared-history observation is not a row population")
		}
		result.rows = rightsRowObservation{count: len(rows), digest: sha256.Sum256([]byte(text))}
		return result
	}
	keys := make([][]int64, len(families))
	before := make([]observation, len(families))
	for index, f := range families {
		query := fmt.Sprintf(`SELECT COALESCE(jsonb_agg(%[1]s ORDER BY %[1]s),'[]'::jsonb)::text FROM %[2]s
		WHERE guild_id=$1 AND %[3]s AND EXISTS(SELECT 1 FROM jsonb_array_elements(%[4]s) a WHERE a->>'account_id'=$3)`,
			f.key, f.name, f.producerPredicate, f.allocations)
		var encoded string
		if err := db.QueryRowContext(ctx, query, guildID, producerID, consumer.AccountID).Scan(&encoded); err != nil {
			t.Fatal(err)
		}
		if json.Unmarshal([]byte(encoded), &keys[index]) != nil || len(keys[index]) == 0 {
			t.Fatalf("invalid shared-history population: no produced %s references", f.name)
		}
		before[index] = read(f, keys[index])
		if before[index].rows.count != len(keys[index]) || before[index].references == 0 {
			t.Fatalf("invalid shared-history population for %s", f.name)
		}
		if absent := read(f, []int64{0}); absent.rows.count != 0 || absent.references != 0 {
			t.Fatal("shared-history reader ignores captured keys")
		}
	}
	ownRows := func() int {
		t.Helper()
		var count int
		if err := db.QueryRowContext(ctx, "SELECT count(*) FROM guild_clearing_results WHERE guild_id=$1 AND account_id=$2", guildID, consumer.AccountID).Scan(&count); err != nil {
			t.Fatal(err)
		}
		return count
	}
	ownBefore := ownRows()
	if ownBefore == 0 {
		t.Fatal("consumer's own settlement population is empty")
	}
	live := compositionRequest(t, server.Client(), http.MethodGet, server.URL+"/api/v1/founder", tokens.AccessToken, "")
	_ = readCompositionBytes(t, live)
	if live.StatusCode != 200 {
		t.Fatal("consumer access was invalid before deletion")
	}
	deleted := compositionRequest(t, server.Client(), http.MethodDelete, server.URL+"/api/v1/account", tokens.AccessToken, "")
	_ = readCompositionBytes(t, deleted)
	t.Logf("RP-120 shared-consumer delete status=%d own_settlements before=%d after=%d", deleted.StatusCode, ownBefore, ownRows())
	if deleted.StatusCode != http.StatusNoContent {
		t.Fatal("shared-consumer account deletion failed")
	}
	for index, f := range families {
		after := read(f, keys[index])
		t.Logf("RP-120 shared-consumer family=%s rows_before=%d rows_after=%d disposition=%s references_before=%d references_after=%d", f.name,
			before[index].rows.count, after.rows.count, rightsRowDisposition(before[index].rows, after.rows), before[index].references, after.references)
	}
	refused := compositionRequest(t, server.Client(), http.MethodGet, server.URL+"/api/v1/founder", tokens.AccessToken, "")
	_ = readCompositionBytes(t, refused)
	if refused.StatusCode != http.StatusUnauthorized {
		t.Fatal("deleted consumer access remains usable")
	}
}

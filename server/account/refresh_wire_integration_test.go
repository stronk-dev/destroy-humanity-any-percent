package account

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"cloud-clicker/server/internal/testhttp"
)

// Real Postgres/session storage, with the existing testhttp net.Pipe exchange.
// This fixture does not claim a browser, OS socket or Game UI workflow. The
// unrelated intent constructor dependency deliberately fails if invoked.
type refreshPersistenceFixture struct {
	now        time.Time
	repository *Repository
	db         *sql.DB
	server     *testhttp.Server
	account    CreatedAccount
	initial    TokenPair
	familyID   string
}

func newRefreshPersistenceFixture(t *testing.T, burst int) *refreshPersistenceFixture {
	t.Helper()
	fixture := &refreshPersistenceFixture{now: time.Date(2026, 10, 6, 0, 0, 0, 0, time.UTC)}
	fixture.repository, fixture.db, _ = bootstrapRepository(t, &fixture.now)
	var err error
	fixture.account, err = fixture.repository.CreateAccount(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	fixture.initial, err = fixture.repository.CreateSession(context.Background(), fixture.account.AccountID, fixture.account.RecoveryCode)
	if err != nil {
		t.Fatal(err)
	}
	hash, ok := opaqueTokenHash(fixture.initial.RefreshToken)
	if !ok {
		t.Fatal("initial refresh token lacks the canonical opaque-token shape")
	}
	if err := fixture.db.QueryRow(`SELECT family_id FROM sessions WHERE token_hash=$1`, hash[:]).Scan(&fixture.familyID); err != nil {
		t.Fatal(err)
	}
	fixture.server = newRefreshPersistenceServer(t, fixture.repository, burst)
	return fixture
}

func newRefreshPersistenceHandler(t *testing.T, repository *Repository, burst int) http.Handler {
	t.Helper()
	config := Phase0APIConfig(testBootstrapReceiptKeys())
	config.UnauthenticatedBurst, config.UnauthenticatedPerMin = burst, 1
	api, err := NewAPI(repository, refreshCensusUnreachableIntents{t}, config)
	if err != nil {
		t.Fatal(err)
	}
	return api.Router()
}

func newRefreshPersistenceServer(t *testing.T, repository *Repository, burst int) *testhttp.Server {
	t.Helper()
	server := testhttp.New(newRefreshPersistenceHandler(t, repository, burst))
	t.Cleanup(server.Close)
	return server
}

func refreshPersistenceRequest(t *testing.T, server *testhttp.Server, token string, wantStatus int) []byte {
	t.Helper()
	response := requestJSON(t, server.Client, http.MethodPost, server.URL+"/api/v1/session/refresh", "", fmt.Sprintf(`{"refresh_token":%q}`, token))
	defer response.Body.Close()
	if response.StatusCode != wantStatus {
		t.Fatalf("refresh status=%d want=%d (response body deliberately withheld)", response.StatusCode, wantStatus)
	}
	if response.Header.Get("Content-Type") != "application/json" {
		t.Fatal("refresh response lost exact application/json content type")
	}
	if wantStatus == http.StatusOK && response.Header.Get("Cache-Control") != "no-store" {
		t.Fatal("rotated credentials are not protected by no-store (body withheld)")
	}
	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}
	registry, err := PrivateAPIRegistry()
	if err != nil {
		t.Fatal(err)
	}
	if registry.ValidateResponse("refresh_session", response.StatusCode, body) != nil {
		t.Fatal("actual persisted refresh response does not conform to its registered descriptor (body withheld)")
	}
	return body
}

func assertRefreshPersistenceError(t *testing.T, server *testhttp.Server, token string, status int, category, detail string) {
	t.Helper()
	body := refreshPersistenceRequest(t, server, token, status)
	want := fmt.Sprintf("{\"category\":%q,\"detail\":%q}\n", category, detail)
	if !bytes.Equal(body, []byte(want)) {
		t.Fatalf("refresh response differs from exact %s/%s error bytes", category, detail)
	}
}

func readRefreshPersistencePair(t *testing.T, server *testhttp.Server, token string) TokenPair {
	t.Helper()
	body := refreshPersistenceRequest(t, server, token, http.StatusOK)
	var keys map[string]json.RawMessage
	if err := json.Unmarshal(body, &keys); err != nil || len(keys) != 2 {
		t.Fatal("refresh success is not an exact two-field JSON object")
	}
	var pair TokenPair
	if json.Unmarshal(keys["access_token"], &pair.AccessToken) != nil || pair.AccessToken == "" ||
		json.Unmarshal(keys["refresh_token"], &pair.RefreshToken) != nil || pair.RefreshToken == "" {
		t.Fatal("refresh success does not contain two nonempty string credentials")
	}
	canonical, err := json.Marshal(pair)
	if err != nil || !bytes.Equal(body, append(canonical, '\n')) {
		t.Fatal("refresh success contains duplicate, extra, noncanonical or trailing fields")
	}
	return pair
}

// These snapshots remain in memory and are never logged: they contain hashed
// credentials and private game state. Scope is exactly these named tables.
func refreshPersistenceSnapshot(t *testing.T, db *sql.DB, credentials bool) []byte {
	t.Helper()
	tables := []string{"accounts", "account_emails", "account_founders", "save_streams", "save_revisions", "events", "intent_records", "run_genesis", "founder_genesis", "transport_player_outbox"}
	if credentials {
		tables = []string{"session_families", "sessions", "access_tokens"}
	}
	var snapshot bytes.Buffer
	for _, table := range tables {
		var rows []byte
		// The table names above are closed test-owned literals, never user input.
		query := `SELECT COALESCE(jsonb_agg(to_jsonb(r) ORDER BY to_jsonb(r)::text),'[]'::jsonb) FROM ` + table + ` r`
		if err := db.QueryRow(query).Scan(&rows); err != nil {
			t.Fatal(err)
		}
		snapshot.WriteString(table)
		snapshot.WriteByte('\n')
		snapshot.Write(rows)
		snapshot.WriteByte('\n')
	}
	return snapshot.Bytes()
}

func assertRefreshPersistenceUnchanged(t *testing.T, db *sql.DB, credentials bool, before []byte) {
	t.Helper()
	if !bytes.Equal(before, refreshPersistenceSnapshot(t, db, credentials)) {
		t.Fatalf("refresh changed forbidden persisted rows (credentials=%t; contents withheld)", credentials)
	}
}

func assertRefreshFamilyCounts(t *testing.T, fixture *refreshPersistenceFixture, want [6]int) {
	t.Helper()
	var got [6]int
	err := fixture.db.QueryRow(`SELECT
		(SELECT count(*) FROM sessions WHERE family_id=$1),
		(SELECT count(*) FROM sessions WHERE family_id=$1 AND consumed_at IS NOT NULL),
		(SELECT count(*) FROM sessions WHERE family_id=$1 AND revoked_at IS NOT NULL),
		(SELECT count(*) FROM access_tokens WHERE family_id=$1),
		(SELECT count(*) FROM access_tokens WHERE family_id=$1 AND revoked_at IS NOT NULL),
		(SELECT count(*) FROM session_families WHERE family_id=$1 AND revoked_at IS NOT NULL)`, fixture.familyID).
		Scan(&got[0], &got[1], &got[2], &got[3], &got[4], &got[5])
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("family row counts=%v want=%v", got, want)
	}
}

func assertRefreshIssuedPair(t *testing.T, fixture *refreshPersistenceFixture, pair TokenPair, founderID string) {
	t.Helper()
	claims, err := fixture.repository.Authenticate(context.Background(), pair.AccessToken)
	if err != nil || claims.Subject != fixture.account.AccountID || claims.FounderID != founderID ||
		claims.IssuedAt != fixture.now.Unix() || claims.ExpiresAt != fixture.now.Add(15*time.Minute).Unix() {
		t.Fatal("issued access credential is not authentic with exact account/Founder/15-minute binding")
	}
	hash, ok := opaqueTokenHash(pair.RefreshToken)
	if !ok {
		t.Fatal("issued refresh credential is not a canonical 256-bit opaque token")
	}
	var sameAccount, sameFamily, unconsumed, unrevoked bool
	var created, expires time.Time
	err = fixture.db.QueryRow(`SELECT account_id=$2,family_id=$3,consumed_at IS NULL,revoked_at IS NULL,created_at,expires_at FROM sessions WHERE token_hash=$1`,
		hash[:], fixture.account.AccountID, fixture.familyID).Scan(&sameAccount, &sameFamily, &unconsumed, &unrevoked, &created, &expires)
	if err != nil || !sameAccount || !sameFamily || !unconsumed || !unrevoked || !created.Equal(fixture.now) || !expires.Equal(fixture.now.Add(30*24*time.Hour)) {
		t.Fatal("issued refresh row lost account/family/unused/30-day binding")
	}
}

func TestRefreshWireRotationAndReuseIntegration(t *testing.T) {
	if os.Getenv("TEST_DATABASE_URL") == "" {
		t.Skip("real refresh population requires TEST_DATABASE_URL; no rotation was executed")
	}
	for _, expiredAccess := range []bool{false, true} {
		t.Run(fmt.Sprintf("access-expired=%t", expiredAccess), func(t *testing.T) {
			fixture := newRefreshPersistenceFixture(t, 10)
			if expiredAccess {
				fixture.now = fixture.now.Add(15 * time.Minute)
			}
			_, initialErr := fixture.repository.Authenticate(context.Background(), fixture.initial.AccessToken)
			if expiredAccess && !errors.Is(initialErr, ErrAuthentication) || !expiredAccess && initialErr != nil {
				t.Fatal("initial access validity does not match the declared population")
			}
			gameplay := refreshPersistenceSnapshot(t, fixture.db, false)
			assertRefreshFamilyCounts(t, fixture, [6]int{1, 0, 0, 1, 0, 0})
			rotated := readRefreshPersistencePair(t, fixture.server, fixture.initial.RefreshToken)
			if rotated.AccessToken == fixture.initial.AccessToken || rotated.RefreshToken == fixture.initial.RefreshToken {
				t.Fatal("refresh did not issue distinct credentials")
			}
			assertRefreshIssuedPair(t, fixture, rotated, fixture.account.FounderID)
			assertRefreshFamilyCounts(t, fixture, [6]int{2, 1, 0, 2, 0, 0})
			oldHash, _ := opaqueTokenHash(fixture.initial.RefreshToken)
			var consumedAtNow bool
			if err := fixture.db.QueryRow(`SELECT consumed_at=$2 FROM sessions WHERE token_hash=$1`, oldHash[:], fixture.now).Scan(&consumedAtNow); err != nil || !consumedAtNow {
				t.Fatal("old refresh token was not consumed at the controlled server time")
			}
			assertRefreshPersistenceUnchanged(t, fixture.db, false, gameplay)
			assertRefreshPersistenceError(t, fixture.server, fixture.initial.RefreshToken, http.StatusUnauthorized, "refresh_reused", "session_family_revoked")
			assertRefreshFamilyCounts(t, fixture, [6]int{2, 1, 2, 2, 2, 1})
			if _, err := fixture.repository.Authenticate(context.Background(), rotated.AccessToken); !errors.Is(err, ErrAuthentication) {
				t.Fatal("rotated access token survived consumed-token family revocation")
			}
			credentials := refreshPersistenceSnapshot(t, fixture.db, true)
			assertRefreshPersistenceError(t, fixture.server, rotated.RefreshToken, http.StatusUnauthorized, "unauthorized", "refresh_token")
			assertRefreshPersistenceUnchanged(t, fixture.db, true, credentials)
			assertRefreshPersistenceUnchanged(t, fixture.db, false, gameplay)
		})
	}
}

func TestRefreshWireNewFounderBindingIntegration(t *testing.T) {
	fixture := newRefreshPersistenceFixture(t, 10)
	newFounder, err := fixture.repository.NewFounder(context.Background(), fixture.account.AccountID)
	if err != nil || newFounder.ID == fixture.account.FounderID {
		t.Fatal("New Founder setup did not replace the active Founder")
	}
	oldClaims, err := fixture.repository.Authenticate(context.Background(), fixture.initial.AccessToken)
	if err != nil || oldClaims.FounderID != fixture.account.FounderID {
		t.Fatal("New Founder incorrectly revoked or rebound the old access credential")
	}
	gameplay := refreshPersistenceSnapshot(t, fixture.db, false)
	rotated := readRefreshPersistencePair(t, fixture.server, fixture.initial.RefreshToken)
	assertRefreshIssuedPair(t, fixture, rotated, newFounder.ID)
	assertRefreshFamilyCounts(t, fixture, [6]int{2, 1, 0, 2, 0, 0})
	assertRefreshPersistenceUnchanged(t, fixture.db, false, gameplay)
}

func TestRefreshWireUnknownExpiredAndFaultIntegration(t *testing.T) {
	if os.Getenv("TEST_DATABASE_URL") == "" {
		t.Skip("real refresh population requires TEST_DATABASE_URL; no database refusal was executed")
	}
	for _, arm := range []string{"unknown-canonical", "expired-refresh", "closed-database"} {
		t.Run(arm, func(t *testing.T) {
			fixture := newRefreshPersistenceFixture(t, 10)
			gameplay := refreshPersistenceSnapshot(t, fixture.db, false)
			credentials := refreshPersistenceSnapshot(t, fixture.db, true)
			switch arm {
			case "unknown-canonical":
				token := base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{0xdd}, 32))
				hash, ok := opaqueTokenHash(token)
				var known int
				if !ok || fixture.db.QueryRow(`SELECT count(*) FROM sessions WHERE token_hash=$1`, hash[:]).Scan(&known) != nil || known != 0 {
					t.Fatal("unknown-token control is not canonical and absent")
				}
				assertRefreshPersistenceError(t, fixture.server, token, http.StatusUnauthorized, "unauthorized", "refresh_token")
				assertRefreshPersistenceUnchanged(t, fixture.db, true, credentials)
			case "expired-refresh":
				fixture.now = fixture.now.Add(30 * 24 * time.Hour)
				assertRefreshPersistenceError(t, fixture.server, fixture.initial.RefreshToken, http.StatusUnauthorized, "unauthorized", "refresh_token")
				assertRefreshFamilyCounts(t, fixture, [6]int{1, 0, 1, 1, 1, 1})
			case "closed-database":
				closed, err := sql.Open("pgx", os.Getenv("TEST_DATABASE_URL"))
				if err != nil {
					t.Fatal(err)
				}
				if err := closed.Close(); err != nil {
					t.Fatal(err)
				}
				faultyRepository := *fixture.repository
				faultyRepository.db = closed // Do not close the primary fixture/cleanup DB.
				_, closedErr := closed.BeginTx(context.Background(), nil)
				_, refreshErr := faultyRepository.RefreshSession(context.Background(), fixture.initial.RefreshToken)
				if closedErr == nil || !errors.Is(refreshErr, closedErr) {
					t.Fatal("fault population did not reach the actual closed database error")
				}
				faultServer := newRefreshPersistenceServer(t, &faultyRepository, 10)
				// Census of current generic-error mapping, not endorsement as outage policy.
				assertRefreshPersistenceError(t, faultServer, fixture.initial.RefreshToken, http.StatusUnauthorized, "unauthorized", "refresh_token")
				assertRefreshPersistenceUnchanged(t, fixture.db, true, credentials)
			}
			assertRefreshPersistenceUnchanged(t, fixture.db, false, gameplay)
		})
	}
}

func TestRefreshWireLimiterNonMutationIntegration(t *testing.T) {
	fixture := newRefreshPersistenceFixture(t, 1)
	gameplay := refreshPersistenceSnapshot(t, fixture.db, false)
	second := readRefreshPersistencePair(t, fixture.server, fixture.initial.RefreshToken)
	assertRefreshIssuedPair(t, fixture, second, fixture.account.FounderID)
	assertRefreshFamilyCounts(t, fixture, [6]int{2, 1, 0, 2, 0, 0})
	credentials := refreshPersistenceSnapshot(t, fixture.db, true)
	assertRefreshPersistenceError(t, fixture.server, second.RefreshToken, http.StatusTooManyRequests, "rate_limited", "ip")
	assertRefreshPersistenceUnchanged(t, fixture.db, true, credentials)
	assertRefreshPersistenceUnchanged(t, fixture.db, false, gameplay)
	fixture.now = fixture.now.Add(time.Minute)
	third := readRefreshPersistencePair(t, fixture.server, second.RefreshToken)
	assertRefreshIssuedPair(t, fixture, third, fixture.account.FounderID)
	assertRefreshFamilyCounts(t, fixture, [6]int{3, 2, 0, 3, 0, 0})
	assertRefreshPersistenceUnchanged(t, fixture.db, false, gameplay)
}

// Suppress the HTTP response at the transport boundary, not the transaction.
// Never retain its credential bytes: only count the writes and record headers.
type refreshLostReplyWriter struct {
	header http.Header
	status int
	bytes  int
}

func (writer *refreshLostReplyWriter) Header() http.Header    { return writer.header }
func (writer *refreshLostReplyWriter) WriteHeader(status int) { writer.status = status }
func (writer *refreshLostReplyWriter) Write(data []byte) (int, error) {
	if writer.status == 0 {
		writer.status = http.StatusOK
	}
	writer.bytes += len(data)
	return len(data), nil
}

func TestRefreshWireCommittedReplyLossIntegration(t *testing.T) {
	fixture := newRefreshPersistenceFixture(t, 10)
	gameplay := refreshPersistenceSnapshot(t, fixture.db, false)
	// Healthy control: the same API/DB issues a usable pair before response loss.
	second := readRefreshPersistencePair(t, fixture.server, fixture.initial.RefreshToken)
	assertRefreshIssuedPair(t, fixture, second, fixture.account.FounderID)
	assertRefreshFamilyCounts(t, fixture, [6]int{2, 1, 0, 2, 0, 0})

	type observation struct {
		status, bytes int
		contentType   string
		closeError    error
	}
	completed := make(chan observation, 1)
	handler := newRefreshPersistenceHandler(t, fixture.repository, 10)
	lostReplyServer := testhttp.New(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		writer := &refreshLostReplyWriter{header: make(http.Header)}
		handler.ServeHTTP(writer, request)
		// The real handler has returned from its committed rotation, but none of
		// its status/headers/body have reached the HTTP client. Close that pipe.
		connection, _, err := response.(http.Hijacker).Hijack()
		if err == nil {
			err = connection.Close()
		}
		completed <- observation{writer.status, writer.bytes, writer.header.Get("Content-Type"), err}
	}))
	t.Cleanup(lostReplyServer.Close)
	request, err := http.NewRequest(http.MethodPost, lostReplyServer.URL+"/api/v1/session/refresh",
		bytes.NewBufferString(fmt.Sprintf(`{"refresh_token":%q}`, second.RefreshToken)))
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := lostReplyServer.Client.Do(request)
	if response != nil {
		_ = response.Body.Close()
		t.Fatal("lost-reply client unexpectedly received an HTTP response")
	}
	if !errors.Is(err, io.EOF) {
		t.Fatal("lost-reply client did not observe the deliberately closed HTTP connection")
	}
	observed := <-completed
	if observed.closeError != nil || observed.status != http.StatusOK || observed.bytes == 0 || observed.contentType != "application/json" {
		t.Fatal("lost reply did not follow a successful actual refresh-handler response")
	}
	// A client-side failure is not rollback: the old credential was consumed
	// and an unseen descendant pair was committed in the same session family.
	assertRefreshFamilyCounts(t, fixture, [6]int{3, 2, 0, 3, 0, 0})
	assertRefreshPersistenceUnchanged(t, fixture.db, false, gameplay)
	assertRefreshPersistenceError(t, fixture.server, second.RefreshToken, http.StatusUnauthorized, "refresh_reused", "session_family_revoked")
	assertRefreshFamilyCounts(t, fixture, [6]int{3, 2, 3, 3, 3, 1})
	if _, err := fixture.repository.Authenticate(context.Background(), second.AccessToken); !errors.Is(err, ErrAuthentication) {
		t.Fatal("received access token survived lost-reply retry family revocation")
	}
	credentials := refreshPersistenceSnapshot(t, fixture.db, true)
	assertRefreshPersistenceError(t, fixture.server, second.RefreshToken, http.StatusUnauthorized, "unauthorized", "refresh_token")
	assertRefreshPersistenceUnchanged(t, fixture.db, true, credentials)
	assertRefreshPersistenceUnchanged(t, fixture.db, false, gameplay)
}

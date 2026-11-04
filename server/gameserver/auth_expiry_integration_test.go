package gameserver

import (
	"context"
	"database/sql"
	"io"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"cloud-clicker/server/transport"

	"github.com/coder/websocket"
)

func TestComposedAccessExpiryCancelsBlockedSocketAuthenticationIntegration(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	database := newCredentialCleanupDatabase(t, ctx)
	defer database.Close()
	clock := &mutableClock{now: time.Now().UTC().Add(-14*time.Minute - 30*time.Second)}
	_, composition := newCredentialCleanupComposition(t, ctx, database, clock, io.Discard)
	// Issue an actual normal 15-minute token near the end of its lifetime.
	// Advance only the injected application clock, never the socket scheduler.
	clock.Set(time.Now().UTC().Add(-14*time.Minute - 30*time.Second))
	account, err := composition.Accounts.CreateAccount(ctx)
	if err != nil {
		t.Fatal(err)
	}
	pair, err := composition.Accounts.CreateSession(ctx, account.AccountID, account.RecoveryCode)
	if err != nil {
		t.Fatal(err)
	}
	clock.Set(time.Now().UTC())
	claims, err := composition.Accounts.Authenticate(ctx, pair.AccessToken)
	if err != nil || claims.ExpiresAt-claims.IssuedAt != 900 {
		t.Fatalf("fixture lacks an actual normal-lifetime access token: %v", err)
	}
	expires := time.Unix(claims.ExpiresAt, 0)
	serverCtx, stopServer := context.WithCancel(ctx)
	defer stopServer()
	if err := composition.Server.Start(serverCtx); err != nil {
		t.Fatal(err)
	}
	httpServer := httptest.NewServer(composition.Server.Handler())
	defer httpServer.Close()
	peer := dialCompositionSocket(t, "ws"+strings.TrimPrefix(httpServer.URL, "http")+"/connection/websocket", httpServer.Client(), pair.AccessToken)
	defer peer.connection.CloseNow()
	closed := make(chan websocket.StatusCode, 1)
	go func() { closed <- waitSocketCloseStatus(t, peer.connection, time.Until(expires.Add(2*time.Second))) }()
	lock, err := database.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer lock.Rollback()
	if _, err := lock.ExecContext(ctx, "LOCK TABLE access_tokens IN ACCESS EXCLUSIVE MODE"); err != nil {
		t.Fatal(err)
	}
	const blockedAuthentication = `SELECT a.pid FROM pg_stat_activity a JOIN pg_locks l USING(pid)
		WHERE a.datname=current_database() AND l.relation='access_tokens'::regclass
		AND NOT l.granted AND l.mode='AccessShareLock'
		AND a.query LIKE 'SELECT account_id,founder_id,expires_at,revoked_at FROM access_tokens WHERE jti=%'`
	var authPID int
	for {
		err := database.QueryRowContext(ctx, blockedAuthentication).Scan(&authPID)
		if err == nil {
			break
		}
		if err != sql.ErrNoRows {
			t.Fatal(err)
		}
		select {
		case status := <-closed:
			t.Fatalf("socket closed before an actual blocked recheck: %d", status)
		case <-ctx.Done():
			t.Fatal("actual Account recheck never blocked")
		case <-time.After(10 * time.Millisecond):
		}
	}
	if status := <-closed; status != websocket.StatusCode(transport.CloseAuthExpired) {
		t.Fatalf("expired socket with actual blocked Account SQL close=%d want=%d", status, transport.CloseAuthExpired)
	}
	if time.Now().Before(expires) {
		t.Fatal("live access token expired early")
	}
	// Neither releasing the table nor draining the node may rescue this oracle.
	// The original query must cease waiting while its exclusive lock is held.
	queryCtx, stopQuery := context.WithTimeout(ctx, time.Second)
	defer stopQuery()
	for {
		var waiting bool
		if err := database.QueryRowContext(queryCtx, "SELECT EXISTS(SELECT 1 FROM pg_locks WHERE pid=$1 AND NOT granted)", authPID).Scan(&waiting); err != nil {
			t.Fatal(err)
		}
		if !waiting {
			break
		}
		select {
		case <-queryCtx.Done():
			t.Fatal("Account SQL survived token expiry")
		case <-time.After(10 * time.Millisecond):
		}
	}
	if err := lock.Rollback(); err != nil {
		t.Fatal(err)
	}
	var revoked int
	if err := database.QueryRowContext(ctx, "SELECT count(*) FROM access_tokens WHERE revoked_at IS NOT NULL").Scan(&revoked); err != nil || revoked != 0 {
		t.Fatalf("token expiry changed revocation state: count=%d err=%v", revoked, err)
	}
	freshPair, err := composition.Accounts.CreateSession(ctx, account.AccountID, account.RecoveryCode)
	if err != nil {
		t.Fatal(err)
	}
	freshPeer := dialCompositionSocket(t, "ws"+strings.TrimPrefix(httpServer.URL, "http")+"/connection/websocket", httpServer.Client(), freshPair.AccessToken)
	defer freshPeer.connection.CloseNow()
}

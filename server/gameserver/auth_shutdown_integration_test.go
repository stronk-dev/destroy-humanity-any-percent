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

func TestComposedDrainCancelsBlockedSocketAuthenticationIntegration(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
	defer cancel()
	database := newCredentialCleanupDatabase(t, ctx)
	defer database.Close()
	clock := &mutableClock{now: time.Now().UTC()}
	_, composition := newCredentialCleanupComposition(t, ctx, database, clock, io.Discard)
	serverCtx, stopServer := context.WithCancel(ctx)
	defer stopServer()
	if err := composition.Server.Start(serverCtx); err != nil {
		t.Fatal(err)
	}
	account, err := composition.Accounts.CreateAccount(ctx)
	if err != nil {
		t.Fatal(err)
	}
	pair, err := composition.Accounts.CreateSession(ctx, account.AccountID, account.RecoveryCode)
	if err != nil {
		t.Fatal(err)
	}
	httpServer := httptest.NewServer(composition.Server.Handler())
	defer httpServer.Close()
	peer := dialCompositionSocket(t, "ws"+strings.TrimPrefix(httpServer.URL, "http")+"/connection/websocket", httpServer.Client(), pair.AccessToken)
	defer peer.connection.CloseNow()
	closed := make(chan websocket.StatusCode, 1)
	go func() { closed <- waitSocketCloseStatus(t, peer.connection, 35*time.Second) }()
	lock, err := database.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer lock.Rollback()
	if _, err := lock.ExecContext(ctx, "LOCK TABLE access_tokens IN ACCESS EXCLUSIVE MODE"); err != nil {
		t.Fatal(err)
	}
	// Observe the actual Account SQL from the real periodic callback. No
	// shortened heartbeat, direct callback, mock query or backend kill is used.
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
		case <-ctx.Done():
			t.Fatal("periodic Account authentication never blocked on the held table lock")
		case status := <-closed:
			t.Fatalf("healthy socket closed before blocked authentication: %d", status)
		case <-time.After(10 * time.Millisecond):
		}
	}
	drainCtx, stopDrain := context.WithTimeout(ctx, 2*time.Second)
	defer stopDrain()
	if err := composition.Server.Drain(drainCtx, clock.Time()); err != nil {
		t.Fatalf("drain with actual blocked Account SQL: %v", err)
	}
	if status := <-closed; status != websocket.StatusCode(transport.CloseServerDrain) {
		t.Fatalf("healthy drain close=%d want=%d", status, transport.CloseServerDrain)
	}
	// Keep the exclusive lock held: success must come from cancellation, not
	// the fixture releasing the query or destroying the database.
	for {
		var waiting bool
		if err := database.QueryRowContext(drainCtx, "SELECT EXISTS(SELECT 1 FROM pg_locks WHERE pid=$1 AND NOT granted)", authPID).Scan(&waiting); err != nil {
			t.Fatal(err)
		}
		if !waiting {
			break
		}
		select {
		case <-drainCtx.Done():
			t.Fatal("Account SQL lock wait survived drain")
		case <-time.After(10 * time.Millisecond):
		}
	}
	if err := lock.Rollback(); err != nil {
		t.Fatal(err)
	}
	var revoked int
	if err := database.QueryRowContext(ctx, "SELECT count(*) FROM access_tokens WHERE revoked_at IS NOT NULL").Scan(&revoked); err != nil || revoked != 0 {
		t.Fatalf("drain changed revocation state: count=%d err=%v", revoked, err)
	}
}

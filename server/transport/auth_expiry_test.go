package transport

import (
	"context"
	"strings"
	"testing"
	"time"

	"cloud-clicker/server/account"
	"cloud-clicker/server/internal/testhttp"

	"github.com/coder/websocket"
)

func TestAccessExpiryCancelsBlockedSocketReauthentication(t *testing.T) {
	now := time.Now().UTC()
	// The unchanged first presence callback occurs between 12.5 and 25 seconds.
	// Expiry follows it; no direct callback or accelerated library clock is used.
	expires := now.Add(30 * time.Second).Truncate(time.Second)
	auth := &shutdownAuthenticator{staticAuthenticator: staticAuthenticator{claims: account.Claims{
		Subject: "account", FounderID: "founder", IssuedAt: now.Add(-time.Minute).Unix(), ExpiresAt: expires.Unix(), TokenID: "token",
	}}, entered: make(chan struct{}), finished: make(chan struct{}), release: make(chan struct{}), blockOn: 2}
	node, err := NewNode(phase0Policy(t), auth, testMemberships{})
	if err != nil {
		t.Fatal(err)
	}
	if err := node.Run(); err != nil {
		t.Fatal(err)
	}
	server := testhttp.New(node.Handler())
	defer func() {
		// Release before server cleanup even when the old implementation fails.
		close(auth.release)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = node.Shutdown(ctx)
		server.Close()
	}()
	connection := dialAuthenticated(t, "ws"+strings.TrimPrefix(server.URL, "http"), server.Client)
	defer connection.CloseNow()
	readCtx, cancelRead := context.WithDeadline(context.Background(), expires.Add(2*time.Second))
	defer cancelRead()
	closed := make(chan error, 1)
	go func() {
		for {
			_, data, err := connection.Read(readCtx)
			if err != nil {
				closed <- err
				return
			}
			if strings.TrimSpace(string(data)) == "{}" {
				if err := connection.Write(readCtx, websocket.MessageText, []byte("{}")); err != nil {
					closed <- err
					return
				}
			}
		}
	}()
	select {
	case <-auth.entered:
	case err := <-closed:
		t.Fatalf("socket closed before the periodic recheck: %v", err)
	case <-readCtx.Done():
		t.Fatal("periodic recheck never started")
	}
	if err := <-closed; websocket.CloseStatus(err) != websocket.StatusCode(CloseAuthExpired) {
		t.Fatalf("expired socket with a blocked recheck close=%d want=%d: %v", websocket.CloseStatus(err), CloseAuthExpired, err)
	}
	if time.Now().Before(expires) {
		t.Fatal("live access token expired early")
	}
	select {
	case <-auth.finished:
	default:
		t.Fatal("blocked authentication survived token expiry")
	}
	if node.authCtx.Err() != nil {
		t.Fatal("token expiry shut down the whole node")
	}
}

package transport

import (
	"context"
	"errors"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"cloud-clicker/server/account"
	"cloud-clicker/server/internal/testhttp"

	"github.com/coder/websocket"
)

type shutdownAuthenticator struct {
	staticAuthenticator
	calls    atomic.Int32
	entered  chan struct{}
	finished chan struct{}
	release  chan struct{}
}

func (auth *shutdownAuthenticator) Authenticate(ctx context.Context, token string) (account.Claims, error) {
	claims, err := auth.staticAuthenticator.Authenticate(ctx, token)
	if err != nil || auth.calls.Add(1) != 2 {
		return claims, err
	}
	close(auth.entered)
	defer close(auth.finished)
	select {
	case <-ctx.Done():
		return account.Claims{}, ctx.Err()
	case <-auth.release:
		return claims, nil
	}
}

func TestShutdownCancelsInFlightSocketReauthentication(t *testing.T) {
	now := time.Now().UTC()
	auth := &shutdownAuthenticator{staticAuthenticator: staticAuthenticator{claims: account.Claims{
		Subject: "account", FounderID: "founder", IssuedAt: now.Add(-time.Minute).Unix(), ExpiresAt: now.Add(time.Hour).Unix(), TokenID: "token",
	}}, entered: make(chan struct{}), finished: make(chan struct{}), release: make(chan struct{})}
	node, err := NewNode(phase0Policy(t), auth, testMemberships{})
	if err != nil {
		t.Fatal(err)
	}
	if err := node.Run(); err != nil {
		t.Fatal(err)
	}
	defer func() {
		close(auth.release)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = node.Shutdown(ctx)
	}()
	server := testhttp.New(node.Handler())
	defer server.Close()
	connection := dialAuthenticated(t, "ws"+strings.TrimPrefix(server.URL, "http"), server.Client)
	defer connection.CloseNow()
	readCtx, cancelRead := context.WithTimeout(context.Background(), 35*time.Second)
	defer cancelRead()
	readDone := make(chan error, 1)
	go func() {
		for {
			_, data, err := connection.Read(readCtx)
			if err != nil {
				readDone <- err
				return
			}
			if strings.TrimSpace(string(data)) == "{}" {
				if err := connection.Write(readCtx, websocket.MessageText, []byte("{}")); err != nil {
					readDone <- err
					return
				}
			}
		}
	}()
	// Exercise the real registered OnAlive callback on its unchanged presence
	// cadence; no direct callback invocation or test-only scheduler is used.
	select {
	case <-auth.entered:
	case err := <-readDone:
		t.Fatalf("socket closed before reauthentication: %v", err)
	case <-readCtx.Done():
		t.Fatal("periodic reauthentication did not start")
	}
	node.CloseForDrain()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := node.Shutdown(ctx); err != nil {
		t.Fatalf("shutdown with pending reauthentication: %v", err)
	}
	select {
	case <-auth.finished:
	case <-ctx.Done():
		t.Fatal("reauthentication survived node shutdown")
	}
	if err := <-readDone; errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		t.Fatalf("healthy peer did not observe shutdown: %v", err)
	}
}

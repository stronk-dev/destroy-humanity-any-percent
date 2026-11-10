package transport

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"cloud-clicker/server/account"
	"cloud-clicker/server/internal/testhttp"

	"github.com/centrifugal/centrifuge"
	"github.com/coder/websocket"
)

type shutdownAuthenticator struct {
	staticAuthenticator
	calls              atomic.Int32
	entered            chan struct{}
	finished           chan struct{}
	release            chan struct{}
	blockOn            int32
	ignoreCancellation bool
}

func (auth *shutdownAuthenticator) Authenticate(ctx context.Context, token string) (account.Claims, error) {
	claims, err := auth.staticAuthenticator.Authenticate(ctx, token)
	if err != nil || auth.calls.Add(1) != auth.blockOn {
		return claims, err
	}
	close(auth.entered)
	defer close(auth.finished)
	if auth.ignoreCancellation {
		<-auth.release
		return claims, nil
	}
	select {
	case <-ctx.Done():
		return account.Claims{}, ctx.Err()
	case <-auth.release:
		return claims, nil
	}
}

func TestNodeShutdownUsesServerDrainCloseCode(t *testing.T) {
	now := time.Now().UTC()
	auth := staticAuthenticator{claims: account.Claims{
		Subject: "account", FounderID: "founder", IssuedAt: now.Add(-time.Minute).Unix(), ExpiresAt: now.Add(time.Hour).Unix(), TokenID: "token",
	}}
	node, err := NewNode(phase0Policy(t), auth, testMemberships{})
	if err != nil {
		t.Fatal(err)
	}
	if err := node.Run(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = node.Shutdown(ctx)
	})
	server := testhttp.New(node.Handler())
	defer server.Close()
	connection := dialAuthenticated(t, "ws"+strings.TrimPrefix(server.URL, "http"), server.Client)
	defer connection.CloseNow()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	// Reach the library shutdown path with a live peer, as can happen when
	// it runs before the asynchronous CloseForDrain request takes effect.
	shutdown := make(chan error, 1)
	go func() { shutdown <- node.Shutdown(ctx) }()
	for {
		if _, _, err := connection.Read(ctx); err != nil {
			if status := websocket.CloseStatus(err); status != websocket.StatusCode(CloseServerDrain) {
				t.Errorf("shutdown close=%d want=%d: %v", status, CloseServerDrain, err)
			}
			break
		}
	}
	select {
	case err := <-shutdown:
		if err != nil {
			t.Fatal(err)
		}
	case <-ctx.Done():
		t.Fatal("node shutdown did not finish")
	}
}

func TestShutdownCancelsInFlightSocketReauthentication(t *testing.T) {
	now := time.Now().UTC()
	auth := &shutdownAuthenticator{staticAuthenticator: staticAuthenticator{claims: account.Claims{
		Subject: "account", FounderID: "founder", IssuedAt: now.Add(-time.Minute).Unix(), ExpiresAt: now.Add(time.Hour).Unix(), TokenID: "token",
	}}, entered: make(chan struct{}), finished: make(chan struct{}), release: make(chan struct{}), blockOn: 2}
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

func TestShutdownCancelsPendingSocketConnectAndRejectsLateCredentials(t *testing.T) {
	for _, ignoreCancellation := range []bool{false, true} {
		name := "cancel-aware"
		if ignoreCancellation {
			name = "late-valid-result"
		}
		t.Run(name, func(t *testing.T) {
			now := time.Now().UTC()
			auth := &shutdownAuthenticator{staticAuthenticator: staticAuthenticator{claims: account.Claims{
				Subject: "account", FounderID: "founder", IssuedAt: now.Add(-time.Minute).Unix(), ExpiresAt: now.Add(time.Hour).Unix(), TokenID: "token",
			}}, entered: make(chan struct{}), finished: make(chan struct{}), release: make(chan struct{}), blockOn: 1, ignoreCancellation: ignoreCancellation}
			var releaseOnce sync.Once
			release := func() { releaseOnce.Do(func() { close(auth.release) }) }
			node, err := NewNode(phase0Policy(t), auth, testMemberships{})
			if err != nil {
				t.Fatal(err)
			}
			if err := node.Run(); err != nil {
				t.Fatal(err)
			}
			defer func() {
				release()
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				_ = node.Shutdown(ctx)
			}()
			server := testhttp.New(node.Handler())
			defer server.Close()
			connection, _, err := websocket.Dial(context.Background(), "ws"+strings.TrimPrefix(server.URL, "http"), &websocket.DialOptions{
				HTTPClient: server.Client, HTTPHeader: http.Header{"Origin": []string{"http://localhost:5173"}},
			})
			if err != nil {
				t.Fatal(err)
			}
			defer connection.CloseNow()
			writeCommand(t, connection, map[string]any{"id": 1, "connect": map[string]any{"token": "access.good"}})
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			select {
			case <-auth.entered:
			case <-ctx.Done():
				t.Fatal("connect authentication did not start")
			}
			node.CloseForDrain()
			if err := node.Shutdown(ctx); err != nil {
				t.Fatal(err)
			}
			if !ignoreCancellation {
				select {
				case <-auth.finished:
				case <-ctx.Done():
					t.Error("connect authentication survived node shutdown")
				}
			}
			// A valid result that races cancellation must not register a client
			// after the authenticated-client hub has already been shut down.
			release()
			readCtx, cancelRead := context.WithTimeout(context.Background(), time.Second)
			defer cancelRead()
			for {
				_, data, err := connection.Read(readCtx)
				if err != nil {
					if websocket.CloseStatus(err) != websocket.StatusCode(CloseServerDrain) {
						t.Fatalf("pending connect close=%d want=%d: %v", websocket.CloseStatus(err), CloseServerDrain, err)
					}
					break
				}
				var reply protocolReply
				if err := json.Unmarshal(data, &reply); err != nil {
					t.Fatal(err)
				}
				if reply.Connect != nil {
					t.Fatal("credentials accepted after node shutdown")
				}
			}
		})
	}
}

func TestShutdownClosesLateLibraryRegistration(t *testing.T) {
	node := testNode(t)
	entered, release := make(chan struct{}), make(chan struct{})
	var releaseOnce sync.Once
	finish := func() { releaseOnce.Do(func() { close(release) }) }
	defer finish()
	// Model credentials already granted at the application boundary but not yet
	// registered by Centrifuge. This deliberately replaces OnConnecting; it
	// tests the real OnConnect cleanup guard, not Account authentication.
	node.node.OnConnecting(func(context.Context, centrifuge.ConnectEvent) (centrifuge.ConnectReply, error) {
		close(entered)
		<-release
		info, _ := json.Marshal(Identity{AccountID: "account", FounderID: "founder"})
		return centrifuge.ConnectReply{Credentials: &centrifuge.Credentials{UserID: "account", ExpireAt: time.Now().Add(time.Hour).Unix(), Info: info}}, nil
	})
	server := testhttp.New(node.Handler())
	defer server.Close()
	connection, _, err := websocket.Dial(context.Background(), "ws"+strings.TrimPrefix(server.URL, "http"), &websocket.DialOptions{
		HTTPClient: server.Client, HTTPHeader: http.Header{"Origin": []string{"http://localhost:5173"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer connection.CloseNow()
	writeCommand(t, connection, map[string]any{"id": 1, "connect": map[string]any{"token": "access.good"}})
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	select {
	case <-entered:
	case <-ctx.Done():
		t.Fatal("credential handoff did not start")
	}
	node.CloseForDrain()
	if err := node.Shutdown(ctx); err != nil {
		t.Fatal(err)
	}
	finish()
	readCtx, stopRead := context.WithTimeout(context.Background(), time.Second)
	defer stopRead()
	registered := false
	for {
		_, data, err := connection.Read(readCtx)
		if err != nil {
			if websocket.CloseStatus(err) != websocket.StatusCode(CloseServerDrain) {
				t.Fatalf("late registration survived shutdown: %v", err)
			}
			break
		}
		var reply protocolReply
		if err := json.Unmarshal(data, &reply); err != nil {
			t.Fatal(err)
		}
		registered = registered || reply.Connect != nil
	}
	if !registered || node.ConnectionCount() != 0 {
		t.Fatalf("handoff registered=%v remaining connections=%d", registered, node.ConnectionCount())
	}
	node.queueMu.Lock()
	defer node.queueMu.Unlock()
	if len(node.queues) != 0 {
		t.Fatalf("late registration retained %d queues", len(node.queues))
	}
}

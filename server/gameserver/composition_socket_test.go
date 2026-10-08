package gameserver

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"cloud-clicker/server/transport"
	"github.com/coder/websocket"
)

// Use a real WebSocket and force the application ping, independent of
// Centrifuge's randomized first-ping/OnAlive ordering in the Postgres witness.
func compositionProtocolPeer(t *testing.T, frame, wantReply string, closeCode websocket.StatusCode) (*compositionSocket, <-chan error) {
	t.Helper()
	observed := make(chan error, 1)
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		connection, err := websocket.Accept(response, request, nil)
		if err != nil {
			observed <- err
			return
		}
		defer connection.CloseNow()
		ctx, cancel := context.WithTimeout(request.Context(), 2*time.Second)
		defer cancel()
		if err = connection.Write(ctx, websocket.MessageText, []byte(frame)); err == nil {
			var reply []byte
			_, reply, err = connection.Read(ctx)
			if err == nil && !bytes.Equal(reply, []byte(wantReply)) {
				err = fmt.Errorf("unexpected protocol response: got %q want %q", reply, wantReply)
			}
		}
		observed <- err
		if err == nil {
			_ = connection.Close(closeCode, "protocol test")
		}
	}))
	t.Cleanup(server.Close)
	connection, _, err := websocket.Dial(context.Background(), "ws"+strings.TrimPrefix(server.URL, "http"), nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = connection.CloseNow() })
	return &compositionSocket{connection: connection}, observed
}

func TestCompositionSocketWaitAnswersPingBeforeTypedClose(t *testing.T) {
	for _, code := range []websocket.StatusCode{transport.CloseAuthExpired, transport.CloseServerDrain} {
		t.Run(fmt.Sprint(code), func(t *testing.T) {
			socket, observed := compositionProtocolPeer(t, "{}\n", "{}", code)
			got := waitSocketCloseStatus(t, socket.connection, 3*time.Second)
			if err := <-observed; err != nil {
				t.Fatalf("healthy test peer did not answer the application ping: %v", err)
			}
			if got != code {
				t.Fatalf("typed close changed: got %d want %d", got, code)
			}
		})
	}
}

func TestCompositionSocketReadPreservesRepliesAroundPing(t *testing.T) {
	socket, observed := compositionProtocolPeer(t, "{\"id\":7,\"connect\":{}}\n{}\n{\"id\":8,\"subscribe\":{}}\n", "{}", transport.CloseServerDrain)
	if reply := readWSID(t, socket, 7); reply.Connect == nil {
		t.Fatal("reply before ping was lost")
	}
	if reply := readWSID(t, socket, 8); reply.Subscribe == nil {
		t.Fatal("reply after ping was lost")
	}
	if got := waitSocketCloseStatus(t, socket.connection, 3*time.Second); got != transport.CloseServerDrain {
		t.Fatalf("typed close changed: got %d", got)
	}
	if err := <-observed; err != nil {
		t.Fatalf("batch ping did not receive a pong: %v", err)
	}
}

func TestCompositionSocketDoesNotAnswerNestedEmptyObjectAsPing(t *testing.T) {
	socket, observed := compositionProtocolPeer(t, `{"id":7,"connect":{}}`, `{"id":99}`, transport.CloseServerDrain)
	if reply := readWSID(t, socket, 7); reply.Connect == nil {
		t.Fatal("ordinary reply was lost")
	}
	writeWS(t, socket, map[string]any{"id": 99})
	if got := waitSocketCloseStatus(t, socket.connection, 3*time.Second); got != transport.CloseServerDrain {
		t.Fatalf("typed close changed: got %d", got)
	}
	if err := <-observed; err != nil {
		t.Fatalf("ordinary reply incorrectly caused a pong: %v", err)
	}
}

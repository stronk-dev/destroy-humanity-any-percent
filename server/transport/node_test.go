package transport

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"cloud-clicker/server/account"

	"github.com/coder/websocket"
)

type staticAuthenticator struct {
	claims account.Claims
	err    error
}

func (auth staticAuthenticator) Authenticate(_ context.Context, token string) (account.Claims, error) {
	if auth.err != nil || token != "access.good" {
		return account.Claims{}, errors.New("invalid token")
	}
	return auth.claims, nil
}

func testNode(t *testing.T) *Node {
	t.Helper()
	now := time.Now().UTC()
	node, err := NewNode(phase0Policy(t), staticAuthenticator{claims: account.Claims{
		Subject: "account", FounderID: "founder", IssuedAt: now.Add(-time.Minute).Unix(), ExpiresAt: now.Add(time.Hour).Unix(), TokenID: "token",
	}}, testMemberships{})
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
	return node
}

func TestActualWebsocketAuthOriginSubscriptionsAndNoClientPublish(t *testing.T) {
	node := testNode(t)
	server := httptest.NewServer(node.Handler())
	defer server.Close()
	endpoint := "ws" + strings.TrimPrefix(server.URL, "http")

	if connection, response, err := websocket.Dial(context.Background(), endpoint, &websocket.DialOptions{HTTPHeader: http.Header{"Origin": []string{"https://denied.example"}}}); err == nil {
		_ = connection.CloseNow()
		t.Fatal("disallowed origin upgraded")
	} else if response == nil || response.StatusCode != http.StatusForbidden {
		t.Fatalf("disallowed response=%v err=%v", response, err)
	}

	connection, _, err := websocket.Dial(context.Background(), endpoint, &websocket.DialOptions{HTTPHeader: http.Header{"Origin": []string{"http://localhost:5173"}}})
	if err != nil {
		t.Fatal(err)
	}
	defer connection.CloseNow()
	writeCommand(t, connection, map[string]any{"id": 1, "connect": map[string]any{"token": "access.good"}})
	connect := readReply(t, connection)
	if connect.Error != nil || connect.Connect == nil {
		t.Fatalf("connect reply=%+v", connect)
	}

	writeCommand(t, connection, map[string]any{"id": 2, "subscribe": map[string]any{"channel": "player:founder"}})
	if reply := readReply(t, connection); reply.Error != nil || reply.Subscribe == nil {
		t.Fatalf("private subscribe=%+v", reply)
	}
	writeCommand(t, connection, map[string]any{"id": 3, "subscribe": map[string]any{"channel": "player:other"}})
	if reply := readReply(t, connection); reply.Error == nil || reply.Error.Code != 103 {
		t.Fatalf("unauthorized subscribe=%+v", reply)
	}
	writeCommand(t, connection, map[string]any{"id": 4, "publish": map[string]any{"channel": "player:founder", "data": map[string]any{"forbidden": true}}})
	if reply := readReply(t, connection); reply.Error == nil || reply.Error.Code != 103 {
		t.Fatalf("client publish=%+v", reply)
	}
}

func TestWorldPublishesAreCoalescedAtConfiguredRate(t *testing.T) {
	node := testNode(t)
	server := httptest.NewServer(node.Handler())
	defer server.Close()
	connection, _, err := websocket.Dial(context.Background(), "ws"+strings.TrimPrefix(server.URL, "http"), &websocket.DialOptions{HTTPHeader: http.Header{"Origin": []string{"http://localhost:5173"}}})
	if err != nil {
		t.Fatal(err)
	}
	defer connection.CloseNow()
	writeCommand(t, connection, map[string]any{"id": 1, "connect": map[string]any{"token": "access.good"}})
	_ = readReply(t, connection)
	writeCommand(t, connection, map[string]any{"id": 2, "subscribe": map[string]any{"channel": "world"}})
	_ = readReply(t, connection)

	for revision := int64(1); revision <= 20; revision++ {
		payload, _ := json.Marshal(map[string]any{"scope": "world", "rev": revision, "state": map[string]any{"revision": revision}})
		if err := node.Publish(Envelope{Version: WireVersion, Channel: "world", Kind: "snapshot", Revision: revision,
			ConstantsHash: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", Timestamp: time.Now().UTC(), Payload: payload}); err != nil {
			t.Fatal(err)
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_, data, err := connection.Read(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `"revision":20`) {
		t.Fatalf("latest world state missing: %s", data)
	}

	quiet, quietCancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer quietCancel()
	if _, extra, err := connection.Read(quiet); err == nil {
		t.Fatalf("coalescing emitted extra publication: %s", extra)
	} else if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("unexpected quiet read error: %v", err)
	}
}

type protocolReply struct {
	ID        uint32          `json:"id"`
	Error     *protocolError  `json:"error"`
	Connect   json.RawMessage `json:"connect"`
	Subscribe json.RawMessage `json:"subscribe"`
}

type protocolError struct {
	Code uint32 `json:"code"`
}

func writeCommand(t *testing.T, connection *websocket.Conn, value any) {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := connection.Write(ctx, websocket.MessageText, data); err != nil {
		t.Fatal(err)
	}
}

func readReply(t *testing.T, connection *websocket.Conn) protocolReply {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_, data, err := connection.Read(ctx)
	if err != nil {
		t.Fatal(err)
	}
	var reply protocolReply
	if err := json.Unmarshal(data, &reply); err != nil {
		t.Fatalf("decode %s: %v", data, err)
	}
	return reply
}

package transport

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"cloud-clicker/server/account"
	"cloud-clicker/server/internal/testhttp"

	"github.com/coder/websocket"
)

const (
	soakConnections = 5_000
	soakWorldTicks  = 10
)

type soakAuthenticator struct{}

func (soakAuthenticator) Authenticate(_ context.Context, token string) (account.Claims, error) {
	value, ok := strings.CutPrefix(token, "access.")
	index, err := strconv.Atoi(value)
	if !ok || err != nil || index < 0 || index >= soakConnections {
		return account.Claims{}, errorsForSoak("invalid token")
	}
	now := time.Now().UTC()
	return account.Claims{
		Subject: "account-" + value, FounderID: "founder-" + value, IssuedAt: now.Add(-time.Minute).Unix(), ExpiresAt: now.Add(time.Hour).Unix(), TokenID: "token-" + value,
	}, nil
}

type errorsForSoak string

func (err errorsForSoak) Error() string { return string(err) }

func TestFiveThousandConnectionWorldFanoutSoak(t *testing.T) {
	policy := phase0Policy(t)
	policy.WorldHz = 10
	node, err := NewNode(policy, soakAuthenticator{}, testMemberships{})
	if err != nil {
		t.Fatal(err)
	}
	if err := node.Run(); err != nil {
		t.Fatal(err)
	}
	server := testhttp.New(node.Handler())
	defer server.Close()
	endpoint := "ws" + strings.TrimPrefix(server.URL, "http")

	connections := make([]*websocket.Conn, 0, soakConnections)
	defer func() {
		for _, connection := range connections {
			_ = connection.CloseNow()
		}
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = node.Shutdown(ctx)
	}()
	for index := 0; index < soakConnections; index++ {
		connections = append(connections, dialSoakSubscriber(t, endpoint, server.Client, index))
	}
	if got := node.ConnectionCount(); got != soakConnections {
		t.Fatalf("connected=%d want=%d", got, soakConnections)
	}

	results := make(chan error, soakConnections)
	for _, connection := range connections {
		go sniffWorldPublications(connection, results)
	}
	for revision := int64(1); revision <= soakWorldTicks; revision++ {
		// A click-shaped private receipt must be rejected before it reaches a
		// public queue; the sniffers below would also fail if one leaked.
		privatePayload := json.RawMessage(`{"outcome":"applied"}`)
		if err := node.Publish(Envelope{Version: WireVersion, Channel: "world", Kind: "receipt", Revision: revision,
			ConstantsHash: soakHash, Timestamp: time.Now().UTC(), Payload: privatePayload}); err == nil {
			t.Fatal("public per-click receipt was accepted")
		}
		publishWorld(t, node, revision)
		time.Sleep(110 * time.Millisecond)
	}
	for index := 0; index < soakConnections; index++ {
		if err := <-results; err != nil {
			t.Fatalf("subscriber %d: %v", index, err)
		}
	}
}

const soakHash = "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

func dialSoakSubscriber(t *testing.T, endpoint string, client *http.Client, index int) *websocket.Conn {
	t.Helper()
	connection, _, err := websocket.Dial(context.Background(), endpoint, &websocket.DialOptions{
		HTTPClient: client, HTTPHeader: http.Header{"Origin": []string{"http://localhost:5173"}},
	})
	if err != nil {
		t.Fatalf("dial %d: %v", index, err)
	}
	writeCommand(t, connection, map[string]any{"id": 1, "connect": map[string]any{"token": fmt.Sprintf("access.%d", index)}})
	if reply := readReply(t, connection); reply.Error != nil || reply.Connect == nil {
		t.Fatalf("connect %d: %+v", index, reply)
	}
	writeCommand(t, connection, map[string]any{"id": 2, "subscribe": map[string]any{"channel": "world"}})
	if reply := readReply(t, connection); reply.Error != nil || reply.Subscribe == nil {
		t.Fatalf("subscribe %d: %+v", index, reply)
	}
	return connection
}

func sniffWorldPublications(connection *websocket.Conn, result chan<- error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	for revision := int64(1); revision <= soakWorldTicks; revision++ {
		_, data, err := connection.Read(ctx)
		if err != nil {
			result <- err
			return
		}
		var reply protocolPushReply
		if err := json.Unmarshal(data, &reply); err != nil || reply.Push == nil || reply.Push.Publication == nil || reply.Push.Channel != "world" {
			result <- fmt.Errorf("invalid public push %s: %v", data, err)
			return
		}
		var envelope Envelope
		if err := json.Unmarshal(reply.Push.Publication.Data, &envelope); err != nil || envelope.Channel != "world" || envelope.Kind != "snapshot" || envelope.Revision != revision {
			result <- fmt.Errorf("unexpected public envelope %s: %v", reply.Push.Publication.Data, err)
			return
		}
	}
	result <- nil
}

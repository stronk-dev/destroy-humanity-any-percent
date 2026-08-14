package transport

import (
	"bytes"
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
	var lastRevision int64
	for lastRevision < soakWorldTicks {
		_, data, err := connection.Read(ctx)
		if err != nil {
			result <- err
			return
		}
		revisions, err := worldPublicationRevisions(data)
		if err != nil {
			result <- err
			return
		}
		for _, revision := range revisions {
			if revision <= lastRevision || revision > soakWorldTicks {
				result <- fmt.Errorf("unexpected world revision %d after %d", revision, lastRevision)
				return
			}
			lastRevision = revision
		}
	}
	result <- nil
}

// worldPublicationRevisions mirrors the browser runtime's JSON stream
// boundary: one WebSocket message may contain newline-delimited replies, and a
// Centrifuge batch whose publications were all coalesced away may be empty.
// Protocol keepalives and command replies carry no push and are ignored; any
// actual push must still be a valid world snapshot.
func worldPublicationRevisions(data []byte) ([]int64, error) {
	var revisions []int64
	for _, line := range bytes.Split(data, []byte{'\n'}) {
		line = bytes.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		var reply protocolPushReply
		if err := json.Unmarshal(line, &reply); err != nil {
			return nil, fmt.Errorf("invalid public frame %s: %w", line, err)
		}
		if reply.Push == nil {
			continue
		}
		if reply.Push.Publication == nil || reply.Push.Channel != "world" {
			return nil, fmt.Errorf("invalid public push %s", line)
		}
		var envelope Envelope
		if err := json.Unmarshal(reply.Push.Publication.Data, &envelope); err != nil {
			return nil, fmt.Errorf("invalid public envelope %s: %w", reply.Push.Publication.Data, err)
		}
		if envelope.Channel != "world" || envelope.Kind != "snapshot" || envelope.Revision < 1 {
			return nil, fmt.Errorf("unexpected public envelope %s", reply.Push.Publication.Data)
		}
		revisions = append(revisions, envelope.Revision)
	}
	return revisions, nil
}

func TestWorldPublicationRevisionsAcceptsProtocolFramesAndBatches(t *testing.T) {
	first := `{"push":{"channel":"world","pub":{"data":{"v":2,"ch":"world","kind":"snapshot","rev":1}}}}`
	second := `{"push":{"channel":"world","pub":{"data":{"v":2,"ch":"world","kind":"snapshot","rev":2}}}}`
	revisions, err := worldPublicationRevisions([]byte("\n{}\n{\"id\":7}\n" + first + "\n" + second))
	if err != nil || len(revisions) != 2 || revisions[0] != 1 || revisions[1] != 2 {
		t.Fatalf("protocol stream revisions=%v err=%v", revisions, err)
	}
	if revisions, err = worldPublicationRevisions(nil); err != nil || len(revisions) != 0 {
		t.Fatalf("empty filtered batch revisions=%v err=%v", revisions, err)
	}
	receipt := []byte(`{"push":{"channel":"world","pub":{"data":{"v":2,"ch":"world","kind":"receipt","rev":3}}}}`)
	if _, err = worldPublicationRevisions(receipt); err == nil {
		t.Fatal("private receipt-shaped world push accepted")
	}
}

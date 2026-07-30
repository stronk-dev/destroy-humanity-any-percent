package transport

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"cloud-clicker/server/account"
	"cloud-clicker/server/internal/testhttp"

	"github.com/centrifugal/centrifuge"
	"github.com/centrifugal/protocol"
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
	server := testhttp.New(node.Handler())
	defer server.Close()
	endpoint, httpClient := "ws"+strings.TrimPrefix(server.URL, "http"), server.Client

	if connection, response, err := websocket.Dial(context.Background(), endpoint, &websocket.DialOptions{HTTPClient: httpClient, HTTPHeader: http.Header{"Origin": []string{"https://denied.example"}}}); err == nil {
		_ = connection.CloseNow()
		t.Fatal("disallowed origin upgraded")
	} else if response == nil || response.StatusCode != http.StatusForbidden {
		t.Fatalf("disallowed response=%v err=%v", response, err)
	}

	connection, _, err := websocket.Dial(context.Background(), endpoint, &websocket.DialOptions{HTTPClient: httpClient, HTTPHeader: http.Header{"Origin": []string{"http://localhost:5173"}}})
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
	server := testhttp.New(node.Handler())
	defer server.Close()
	endpoint, httpClient := "ws"+strings.TrimPrefix(server.URL, "http"), server.Client
	connection, _, err := websocket.Dial(context.Background(), endpoint, &websocket.DialOptions{HTTPClient: httpClient, HTTPHeader: http.Header{"Origin": []string{"http://localhost:5173"}}})
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

func TestPublicationEnvelopeMetadata(t *testing.T) {
	envelope := []byte(`{"v":1,"ch":"world","kind":"snapshot","rev":42}`)
	data := []byte(`{"push":{"channel":"world","pub":{"data":` + string(envelope) + `}}}`)
	kind, revision, ok := publicationEnvelopeMetadata(data)
	if !ok || kind != "snapshot" || revision != 42 {
		t.Fatalf("kind=%q revision=%d ok=%v", kind, revision, ok)
	}
	encoded, err := (&protocol.Reply{Push: &protocol.Push{Channel: "world", Pub: &protocol.Publication{Data: envelope}}}).MarshalVT()
	if err != nil {
		t.Fatal(err)
	}
	kind, revision, ok = publicationEnvelopeMetadataForProtocol(centrifuge.ProtocolTypeProtobuf, encoded)
	if !ok || kind != "snapshot" || revision != 42 {
		t.Fatalf("protobuf kind=%q revision=%d ok=%v", kind, revision, ok)
	}
}

func TestActualWebsocketRecoveryReplaysPrivateReceiptsAndLatestWorld(t *testing.T) {
	node := testNode(t)
	server := testhttp.New(node.Handler())
	defer server.Close()
	endpoint, httpClient := "ws"+strings.TrimPrefix(server.URL, "http"), server.Client

	private := dialAuthenticated(t, endpoint, httpClient)
	writeCommand(t, private, map[string]any{"id": 2, "subscribe": map[string]any{"channel": "player:founder"}})
	initialPrivate := decodeSubscribe(t, readReply(t, private).Subscribe)
	if !initialPrivate.Recoverable || !initialPrivate.Positioned || initialPrivate.Epoch == "" {
		t.Fatalf("private position=%+v", initialPrivate)
	}
	publishReceipt(t, node, 1)
	firstPrivate := readPush(t, private)
	if firstPrivate.Channel != "player:founder" || firstPrivate.Publication == nil || firstPrivate.Publication.Offset == 0 {
		t.Fatalf("first private push=%+v", firstPrivate)
	}
	privateOffset := firstPrivate.Publication.Offset
	_ = private.CloseNow()
	publishReceipt(t, node, 2)
	publishReceipt(t, node, 3)

	reconnectedPrivate := dialAuthenticated(t, endpoint, httpClient)
	defer reconnectedPrivate.CloseNow()
	writeCommand(t, reconnectedPrivate, map[string]any{"id": 2, "subscribe": map[string]any{
		"channel": "player:founder", "recover": true, "epoch": initialPrivate.Epoch, "offset": privateOffset,
	}})
	recoveredPrivate := decodeSubscribe(t, readReply(t, reconnectedPrivate).Subscribe)
	if !recoveredPrivate.Recovered || !recoveredPrivate.WasRecovering || len(recoveredPrivate.Publications) != 2 {
		t.Fatalf("private recovery=%+v", recoveredPrivate)
	}
	for index, wantRevision := range []int64{2, 3} {
		publication := recoveredPrivate.Publications[index]
		if publication.Offset != privateOffset+uint64(index+1) || envelopeRevision(t, publication.Data) != wantRevision {
			t.Fatalf("private publication[%d]=%+v", index, publication)
		}
	}

	world := dialAuthenticated(t, endpoint, httpClient)
	writeCommand(t, world, map[string]any{"id": 2, "subscribe": map[string]any{"channel": "world"}})
	initialWorld := decodeSubscribe(t, readReply(t, world).Subscribe)
	if !initialWorld.Recoverable || !initialWorld.Positioned || initialWorld.Epoch == "" {
		t.Fatalf("world position=%+v", initialWorld)
	}
	publishWorld(t, node, 1)
	firstWorld := readPush(t, world)
	if firstWorld.Publication == nil || envelopeRevision(t, firstWorld.Publication.Data) != 1 {
		t.Fatalf("first world push=%+v", firstWorld)
	}
	worldOffset := firstWorld.Publication.Offset
	_ = world.CloseNow()
	for revision := int64(2); revision <= 20; revision++ {
		publishWorld(t, node, revision)
	}
	time.Sleep(300 * time.Millisecond)

	reconnectedWorld := dialAuthenticated(t, endpoint, httpClient)
	defer reconnectedWorld.CloseNow()
	writeCommand(t, reconnectedWorld, map[string]any{"id": 2, "subscribe": map[string]any{
		"channel": "world", "recover": true, "epoch": initialWorld.Epoch, "offset": worldOffset,
	}})
	recoveredWorld := decodeSubscribe(t, readReply(t, reconnectedWorld).Subscribe)
	if !recoveredWorld.Recovered || !recoveredWorld.WasRecovering || len(recoveredWorld.Publications) != 1 ||
		envelopeRevision(t, recoveredWorld.Publications[0].Data) != 20 {
		t.Fatalf("world recovery=%+v", recoveredWorld)
	}
}

func TestDrainCourtesyMessageIsLiveOnlyNotRecoverableHistory(t *testing.T) {
	node := testNode(t)
	server := testhttp.New(node.Handler())
	defer server.Close()
	endpoint := "ws" + strings.TrimPrefix(server.URL, "http")
	connection := dialAuthenticated(t, endpoint, server.Client)
	writeCommand(t, connection, map[string]any{"id": 2, "subscribe": map[string]any{"channel": "player:founder"}})
	initial := decodeSubscribe(t, readReply(t, connection).Subscribe)
	publishReceipt(t, node, 1)
	receipt := readPush(t, connection)
	if receipt.Publication == nil || receipt.Publication.Offset == 0 {
		t.Fatalf("receipt=%+v", receipt)
	}
	if err := node.BroadcastDrain("sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", time.Now().UTC()); err != nil {
		t.Fatal(err)
	}
	courtesy := readPush(t, connection)
	if courtesy.Publication == nil || courtesy.Publication.Offset != 0 || envelopeKind(t, courtesy.Publication.Data) != "system" {
		t.Fatalf("courtesy=%+v", courtesy)
	}
	_ = connection.CloseNow()

	reconnected := dialAuthenticated(t, endpoint, server.Client)
	defer reconnected.CloseNow()
	writeCommand(t, reconnected, map[string]any{"id": 2, "subscribe": map[string]any{
		"channel": "player:founder", "recover": true, "epoch": initial.Epoch, "offset": receipt.Publication.Offset,
	}})
	recovered := decodeSubscribe(t, readReply(t, reconnected).Subscribe)
	if !recovered.Recovered || !recovered.WasRecovering || len(recovered.Publications) != 0 {
		t.Fatalf("courtesy entered recovery history: %+v", recovered)
	}
}

func TestActualSlowPrivateConsumerClosesWithQueueOverflowCode(t *testing.T) {
	policy := phase0Policy(t)
	// Keep the byte bound far above this test's total payload. The disconnect
	// must come from the independent application message-count bound.
	policy.PlayerQueueBytes = 1_048_576
	now := time.Now().UTC()
	node, err := NewNode(policy, staticAuthenticator{claims: account.Claims{
		Subject: "account", FounderID: "founder", IssuedAt: now.Add(-time.Minute).Unix(), ExpiresAt: now.Add(time.Hour).Unix(), TokenID: "token",
	}}, testMemberships{})
	if err != nil {
		t.Fatal(err)
	}
	if centrifuge.DisconnectSlow.Code != CloseQueueOverflow {
		t.Fatalf("library slow code=%d", centrifuge.DisconnectSlow.Code)
	}
	if err := node.Run(); err != nil {
		t.Fatal(err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = node.Shutdown(ctx)
	}()
	server := testhttp.New(node.Handler())
	defer server.Close()
	endpoint := "ws" + strings.TrimPrefix(server.URL, "http")
	connection := dialAuthenticated(t, endpoint, server.Client)
	defer connection.CloseNow()
	connection.SetReadLimit(int64(policy.PlayerQueueBytes * 4))
	writeCommand(t, connection, map[string]any{"id": 2, "subscribe": map[string]any{"channel": "player:founder"}})
	_ = readReply(t, connection)

	padding := strings.Repeat("x", 256)
	for revision := int64(1); revision <= int64(policy.PlayerQueueMessages)+2; revision++ {
		payload, _ := json.Marshal(map[string]any{"outcome": "applied", "padding": padding})
		if err := node.Publish(Envelope{Version: WireVersion, Channel: "player:founder", Kind: "receipt", Revision: revision,
			ConstantsHash: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", Timestamp: time.Now().UTC(), Payload: payload}); err != nil {
			t.Fatal(err)
		}
		if revision == 1 {
			// Let the transport writer take the first publication. The in-flight
			// write is not queued; the next 64 publications fill the declared
			// message-count bound and the final one must disconnect.
			time.Sleep(10 * time.Millisecond)
		}
	}
	deadline := time.Now().Add(time.Second)
	for node.ConnectionCount() != 0 && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	if node.ConnectionCount() != 0 {
		t.Fatal("slow consumer remained connected")
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	for reads := 0; reads < policy.PlayerQueueMessages+3; reads++ {
		_, _, err = connection.Read(ctx)
		if err != nil {
			break
		}
	}
	if websocket.CloseStatus(err) != CloseQueueOverflow {
		t.Fatalf("close status=%d err=%v", websocket.CloseStatus(err), err)
	}
}

func TestActualStalledWorldConsumerSkipsQueuedStaleSnapshots(t *testing.T) {
	policy := phase0Policy(t)
	policy.WorldHz = 10
	policy.PlayerQueueBytes = 1_048_576
	now := time.Now().UTC()
	node, err := NewNode(policy, staticAuthenticator{claims: account.Claims{
		Subject: "account", FounderID: "founder", IssuedAt: now.Add(-time.Minute).Unix(), ExpiresAt: now.Add(time.Hour).Unix(), TokenID: "token",
	}}, testMemberships{})
	if err != nil {
		t.Fatal(err)
	}
	if err := node.Run(); err != nil {
		t.Fatal(err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = node.Shutdown(ctx)
	}()
	server := testhttp.New(node.Handler())
	defer server.Close()
	endpoint := "ws" + strings.TrimPrefix(server.URL, "http")
	connection := dialAuthenticated(t, endpoint, server.Client)
	defer connection.CloseNow()
	connection.SetReadLimit(int64(policy.PlayerQueueBytes * 2))
	writeCommand(t, connection, map[string]any{"id": 2, "subscribe": map[string]any{"channel": "world"}})
	_ = readReply(t, connection)

	padding := strings.Repeat("x", 32_000)
	for revision := int64(1); revision <= 8; revision++ {
		payload, _ := json.Marshal(map[string]any{
			"scope": "world",
			"rev":   revision,
			"state": map[string]any{"revision": revision, "padding": padding},
		})
		if err := node.Publish(Envelope{Version: WireVersion, Channel: "world", Kind: "snapshot", Revision: revision,
			ConstantsHash: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", Timestamp: time.Now().UTC(), Payload: payload}); err != nil {
			t.Fatal(err)
		}
		time.Sleep(110 * time.Millisecond)
	}

	revisions := []int64{}
	for {
		ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
		_, data, readErr := connection.Read(ctx)
		cancel()
		if readErr != nil {
			if errors.Is(readErr, context.DeadlineExceeded) {
				break
			}
			t.Fatal(readErr)
		}
		decoder := json.NewDecoder(bytes.NewReader(data))
		for decoder.More() {
			var reply protocolPushReply
			if err := decoder.Decode(&reply); err != nil || reply.Push == nil || reply.Push.Publication == nil {
				t.Fatalf("decode push batch: %v", err)
			}
			revisions = append(revisions, envelopeRevision(t, reply.Push.Publication.Data))
		}
	}
	if len(revisions) == 0 || revisions[len(revisions)-1] != 8 {
		t.Fatalf("latest world revision missing: %v", revisions)
	}
	if len(revisions) > 2 {
		t.Fatalf("stale queued world revisions leaked: %v", revisions)
	}
	if len(revisions) == 2 && revisions[0] >= 8 {
		t.Fatalf("only the in-flight and newest queued snapshots may arrive: %v", revisions)
	}
}

type protocolReply struct {
	ID        uint32          `json:"id"`
	Error     *protocolError  `json:"error"`
	Connect   json.RawMessage `json:"connect"`
	Subscribe json.RawMessage `json:"subscribe"`
}

type protocolSubscribeResult struct {
	Recoverable   bool                  `json:"recoverable"`
	Recovered     bool                  `json:"recovered"`
	WasRecovering bool                  `json:"was_recovering"`
	Positioned    bool                  `json:"positioned"`
	Epoch         string                `json:"epoch"`
	Offset        uint64                `json:"offset"`
	Publications  []protocolPublication `json:"publications"`
}

type protocolPublication struct {
	Data   json.RawMessage `json:"data"`
	Offset uint64          `json:"offset"`
}

type protocolPush struct {
	Channel     string               `json:"channel"`
	Publication *protocolPublication `json:"pub"`
}

type protocolPushReply struct {
	Push *protocolPush `json:"push"`
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

func dialAuthenticated(t *testing.T, endpoint string, httpClient *http.Client) *websocket.Conn {
	t.Helper()
	connection, _, err := websocket.Dial(context.Background(), endpoint, &websocket.DialOptions{HTTPClient: httpClient, HTTPHeader: http.Header{"Origin": []string{"http://localhost:5173"}}})
	if err != nil {
		t.Fatal(err)
	}
	writeCommand(t, connection, map[string]any{"id": 1, "connect": map[string]any{"token": "access.good"}})
	if reply := readReply(t, connection); reply.Error != nil || reply.Connect == nil {
		_ = connection.CloseNow()
		t.Fatalf("connect reply=%+v", reply)
	}
	return connection
}

func decodeSubscribe(t *testing.T, data json.RawMessage) protocolSubscribeResult {
	t.Helper()
	var result protocolSubscribeResult
	if len(data) == 0 {
		t.Fatal("missing subscribe result")
	}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("decode subscribe %s: %v", data, err)
	}
	return result
}

func readPush(t *testing.T, connection *websocket.Conn) protocolPush {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_, data, err := connection.Read(ctx)
	if err != nil {
		t.Fatal(err)
	}
	var reply protocolPushReply
	if err := json.Unmarshal(data, &reply); err != nil || reply.Push == nil {
		t.Fatalf("decode push %s: %v", data, err)
	}
	return *reply.Push
}

func publishReceipt(t *testing.T, node *Node, revision int64) {
	t.Helper()
	payload, _ := json.Marshal(map[string]any{"intent_id": "01985555-0010-7000-8000-000000000010", "outcome": "applied", "new_revision": revision})
	if err := node.Publish(Envelope{Version: WireVersion, Channel: "player:founder", Kind: "receipt", Revision: revision,
		ConstantsHash: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", Timestamp: time.Now().UTC(), Payload: payload}); err != nil {
		t.Fatal(err)
	}
}

func publishWorld(t *testing.T, node *Node, revision int64) {
	t.Helper()
	payload, _ := json.Marshal(map[string]any{"scope": "world", "rev": revision, "state": map[string]any{"revision": revision}})
	if err := node.Publish(Envelope{Version: WireVersion, Channel: "world", Kind: "snapshot", Revision: revision,
		ConstantsHash: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", Timestamp: time.Now().UTC(), Payload: payload}); err != nil {
		t.Fatal(err)
	}
}

func envelopeRevision(t *testing.T, data json.RawMessage) int64 {
	t.Helper()
	var envelope struct {
		Revision int64 `json:"rev"`
	}
	if err := json.Unmarshal(data, &envelope); err != nil {
		t.Fatalf("decode envelope %s: %v", data, err)
	}
	return envelope.Revision
}

func envelopeKind(t *testing.T, data json.RawMessage) string {
	t.Helper()
	var envelope struct {
		Kind string `json:"kind"`
	}
	if err := json.Unmarshal(data, &envelope); err != nil {
		t.Fatalf("decode envelope %s: %v", data, err)
	}
	return envelope.Kind
}

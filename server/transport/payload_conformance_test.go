package transport

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"cloud-clicker/server/internal/testhttp"
)

func TestKnownPayloadConformanceVectors(t *testing.T) {
	data, err := os.ReadFile("../../testdata/transport/payload-conformance-vectors.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Version int `json:"version"`
		Cases   []struct {
			Name     string   `json:"name"`
			Valid    bool     `json:"valid"`
			Envelope Envelope `json:"envelope"`
		} `json:"cases"`
	}
	if err := json.Unmarshal(data, &fixture); err != nil || fixture.Version != 1 || len(fixture.Cases) != 174 {
		t.Fatalf("invalid complete payload conformance fixture: %v", err)
	}
	for _, vector := range fixture.Cases {
		t.Run(vector.Name, func(t *testing.T) {
			before := bytes.Clone(vector.Envelope.Payload)
			encoded, err := Encode(vector.Envelope, phase0Policy(t).MessageBytes)
			if vector.Valid {
				if err != nil || len(encoded) == 0 {
					t.Fatalf("valid owned shape refused: %v", err)
				}
			} else if !errors.Is(err, ErrInvalidPolicy) || encoded != nil {
				t.Fatalf("invalid owned shape published: err=%v", err)
			}
			if !bytes.Equal(before, vector.Envelope.Payload) {
				t.Fatal("encoder rewrote caller-owned payload bytes")
			}
		})
	}
}

func TestPublisherRejectsMissingPresenceCountBeforeHistoryAndHealthyDelivery(t *testing.T) {
	node := testNode(t)
	server := testhttp.New(node.Handler())
	defer server.Close()
	connection := dialAuthenticated(t, "ws"+strings.TrimPrefix(server.URL, "http"), server.Client)
	defer connection.CloseNow()
	writeCommand(t, connection, map[string]any{"id": 2, "subscribe": map[string]any{"channel": "player:founder"}})
	if reply := readReply(t, connection); reply.Error != nil || reply.Subscribe == nil {
		t.Fatalf("subscribe failed: %+v", reply)
	}
	envelope := Envelope{Version: WireVersion, Channel: "player:founder", Kind: "presence", Revision: 1,
		ConstantsHash: "sha256:" + strings.Repeat("a", 64), Timestamp: time.Now().UTC(),
		Payload: json.RawMessage(`{"joined":[],"left":[]}`)}
	if err := node.Publish(envelope); !errors.Is(err, ErrInvalidPolicy) {
		t.Fatalf("publisher accepted the missing count: %v", err)
	}
	envelope.Payload = json.RawMessage(`{"joined":[],"left":[],"count":0}`)
	if err := node.Publish(envelope); err != nil {
		t.Fatal(err)
	}
	push := readPush(t, connection)
	if push.Channel != envelope.Channel || push.Publication == nil || push.Publication.Offset != 1 {
		t.Fatalf("refused frame entered history or healthy delivery failed: %+v", push)
	}
	var received Envelope
	if err := json.Unmarshal(push.Publication.Data, &received); err != nil || received.Kind != "presence" ||
		!bytes.Equal(received.Payload, envelope.Payload) {
		t.Fatalf("healthy presence bytes changed: err=%v", err)
	}
}

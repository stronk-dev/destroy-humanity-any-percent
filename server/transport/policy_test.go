package transport

import (
	"encoding/json"
	"os"
	"testing"
	"time"
)

func TestPhase0PolicyAndWireLimits(t *testing.T) {
	data, err := os.ReadFile("../../balance/transport/phase0.json")
	if err != nil {
		t.Fatal(err)
	}
	policy, err := LoadPolicy(data)
	if err != nil || policy.WorldHz != 4 || policy.PlayerHistorySize != 512 || policy.PlayerQueueMessages != 256 || policy.MessageBytes != 65_536 || policy.DrainTimeoutMS != 15_000 {
		t.Fatalf("policy=%+v err=%v", policy, err)
	}
	payload, _ := json.Marshal(map[string]any{"code": "server_restarting", "resume_after_ms": 15000})
	wire, err := Encode(Envelope{Version: 1, Channel: "world", Kind: "system", Revision: 0, ConstantsHash: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", Timestamp: time.Date(2026, 7, 29, 12, 0, 0, 0, time.UTC), Payload: payload}, policy.MessageBytes)
	if err != nil || len(wire) == 0 {
		t.Fatalf("wire=%s err=%v", wire, err)
	}
	if _, err := Encode(Envelope{Version: 1, Channel: "world", Kind: "system", ConstantsHash: "bad", Timestamp: time.Now(), Payload: payload}, policy.MessageBytes); err == nil {
		t.Fatal("invalid hash accepted")
	}
}

func TestPolicyRejectsUnknownAndUnsafeOrigin(t *testing.T) {
	for _, source := range []string{
		`{"schema_version":1,"world_hz":4,"feed_history_size":50,"player_history_size":512,"player_history_ttl_ms":600000,"player_queue_messages":256,"player_queue_bytes":1048576,"message_bytes":65536,"subscriptions_per_connection":16,"connections_per_account":3,"drain_timeout_ms":15000,"allowed_origins":["https://game.example/path"]}`,
		`{"schema_version":1,"world_hz":4,"feed_history_size":50,"player_history_size":512,"player_history_ttl_ms":600000,"player_queue_messages":256,"player_queue_bytes":1048576,"message_bytes":65536,"subscriptions_per_connection":16,"connections_per_account":3,"drain_timeout_ms":15000,"allowed_origins":["https://game.example"],"surprise":true}`,
	} {
		if _, err := LoadPolicy([]byte(source)); err == nil {
			t.Fatalf("accepted %s", source)
		}
	}
}

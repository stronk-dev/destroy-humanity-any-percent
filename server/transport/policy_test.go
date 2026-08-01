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
	wire, err := Encode(Envelope{Version: WireVersion, Channel: "world", Kind: "system", Revision: 0, ConstantsHash: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", Timestamp: time.Date(2026, 7, 29, 12, 0, 0, 0, time.UTC), Payload: payload}, policy.MessageBytes)
	if err != nil || len(wire) == 0 {
		t.Fatalf("wire=%s err=%v", wire, err)
	}
	if _, err := Encode(Envelope{Version: WireVersion, Channel: "world", Kind: "system", ConstantsHash: "bad", Timestamp: time.Now(), Payload: payload}, policy.MessageBytes); err == nil {
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

func TestWirePayloadAndChannelContractsAreClosed(t *testing.T) {
	policy := phase0Policy(t)
	now := time.Date(2026, 7, 29, 12, 0, 0, 0, time.UTC)
	hash := "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	tests := []struct {
		name    string
		channel string
		kind    string
		rev     int64
		payload string
		valid   bool
	}{
		{name: "company snapshot", channel: "player:founder", kind: "snapshot", rev: 7, payload: `{"scope":"company","rev":7,"state":{}}`, valid: true},
		{name: "world snapshot", channel: "world", kind: "snapshot", rev: 8, payload: `{"scope":"world","rev":8,"state":{"v":1,"world_rev":8,"planet":{"depletion_ppm":0,"health_ppm":0},"commons":{"server_health_ppm":750000,"active_founders":12,"compact_members":7},"population":{"online":3,"founders_total":20},"milestones":{"active_id":null,"progress_ppm":0},"epoch":{"epoch_id":5,"name":"Phase 0"}}}`, valid: true},
		{name: "event", channel: "feed", kind: "event", rev: 9, payload: `{"event_id":"event-9","kind":"run.ended","scope":"company","rev":9,"cursor_effect":"advance","payload":{}}`, valid: true},
		{name: "historical compensation", channel: "player:founder", kind: "event", rev: 5, payload: `{"event_id":"event-c","kind":"compensation","scope":"company","rev":5,"cursor_effect":"historical","payload":{}}`, valid: true},
		{name: "presence", channel: "guild:g", kind: "presence", rev: 0, payload: `{"joined":[],"left":[],"count":2}`, valid: true},
		{name: "system", channel: "world", kind: "system", rev: 0, payload: `{"code":"server_restarting","resume_after_ms":15000}`, valid: true},
		{name: "public receipt", channel: "world", kind: "receipt", rev: 1, payload: `{"outcome":"applied"}`},
		{name: "receipt is object", channel: "player:founder", kind: "receipt", rev: 1, payload: `[]`},
		{name: "revision mismatch", channel: "world", kind: "snapshot", rev: 8, payload: `{"scope":"world","rev":7,"state":{"v":1,"world_rev":7,"planet":{"depletion_ppm":0,"health_ppm":0},"commons":{"server_health_ppm":0,"active_founders":0,"compact_members":0},"population":{"online":0,"founders_total":0},"milestones":{"active_id":null,"progress_ppm":0},"epoch":{"epoch_id":5,"name":"Phase 0"}}}`},
		{name: "scope mismatch", channel: "world", kind: "snapshot", rev: 8, payload: `{"scope":"company","rev":8,"state":{}}`},
		{name: "unknown snapshot field", channel: "world", kind: "snapshot", rev: 8, payload: `{"scope":"world","rev":8,"state":{},"extra":true}`},
		{name: "event payload scalar", channel: "feed", kind: "event", rev: 9, payload: `{"event_id":"event-9","kind":"run.ended","scope":"company","rev":9,"cursor_effect":"advance","payload":1}`},
		{name: "compensation advances", channel: "player:founder", kind: "event", rev: 5, payload: `{"event_id":"event-c","kind":"compensation","scope":"company","rev":5,"cursor_effect":"advance","payload":{}}`},
		{name: "system extra duration", channel: "world", kind: "system", rev: 0, payload: `{"code":"resync_required","resume_after_ms":1}`},
		{name: "unknown channel", channel: "public", kind: "event", rev: 1, payload: `{"event_id":"event-1","kind":"run.ended","scope":"company","rev":1,"cursor_effect":"advance","payload":{}}`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := Encode(Envelope{Version: WireVersion, Channel: test.channel, Kind: test.kind, Revision: test.rev,
				ConstantsHash: hash, Timestamp: now, Payload: json.RawMessage(test.payload)}, policy.MessageBytes)
			if (err == nil) != test.valid {
				t.Fatalf("valid=%v err=%v", test.valid, err)
			}
		})
	}
}

func TestSharedWireVectors(t *testing.T) {
	data, err := os.ReadFile("../../testdata/transport/wire-vectors.json")
	if err != nil {
		t.Fatal(err)
	}
	var vectors []struct {
		Name     string   `json:"name"`
		Valid    bool     `json:"valid"`
		Envelope Envelope `json:"envelope"`
	}
	if err := json.Unmarshal(data, &vectors); err != nil || len(vectors) < 10 {
		t.Fatalf("vectors=%d err=%v", len(vectors), err)
	}
	for _, vector := range vectors {
		t.Run(vector.Name, func(t *testing.T) {
			_, err := Encode(vector.Envelope, phase0Policy(t).MessageBytes)
			if (err == nil) != vector.Valid {
				t.Fatalf("valid=%v err=%v", vector.Valid, err)
			}
		})
	}
}

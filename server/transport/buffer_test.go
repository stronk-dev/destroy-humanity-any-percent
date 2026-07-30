package transport

import (
	"errors"
	"testing"
	"time"
)

type testMemberships struct{}

func (testMemberships) GuildMember(_, id string) bool      { return id == "guild.allowed" }
func (testMemberships) CohortMember(_, id string) bool     { return id == "cohort.allowed" }
func (testMemberships) MatchParticipant(_, id string) bool { return id == "match.allowed" }

func phase0Policy(t *testing.T) Policy {
	t.Helper()
	policy, err := LoadPolicy([]byte(`{"schema_version":1,"world_hz":4,"feed_history_size":50,"player_history_size":256,"player_history_ttl_ms":60000,"player_queue_messages":64,"player_queue_bytes":65536,"message_bytes":65536,"subscriptions_per_connection":16,"connections_per_account":3,"drain_timeout_ms":15000,"allowed_origins":["http://localhost:5173"]}`))
	if err != nil {
		t.Fatal(err)
	}
	return *policy
}

func TestDropStaleWorldAndLosslessReceiptOverflow(t *testing.T) {
	queue, _ := NewConnectionQueue(phase0Policy(t))
	for index := 0; index < 100; index++ {
		if err := queue.Enqueue(QueuedMessage{Channel: "world", Kind: "snapshot", Data: []byte{byte(index)}}); err != nil {
			t.Fatal(err)
		}
	}
	drained := queue.Drain()
	if len(drained) != 1 || drained[0].Data[0] != 99 {
		t.Fatalf("world queue=%+v", drained)
	}
	for index := 0; index < 64; index++ {
		if err := queue.Enqueue(QueuedMessage{Channel: "player:founder", Kind: "receipt", Data: []byte{1}}); err != nil {
			t.Fatal(err)
		}
	}
	if err := queue.Enqueue(QueuedMessage{Channel: "player:founder", Kind: "receipt", Data: []byte{1}}); !errors.Is(err, ErrQueueOverflow) {
		t.Fatalf("overflow err=%v", err)
	}
}

func TestRecoveryKindsAndAuthorization(t *testing.T) {
	history, _ := NewHistory(phase0Policy(t))
	now := time.Date(2026, 7, 29, 12, 0, 0, 0, time.UTC)
	for index := 0; index < 300; index++ {
		_, _ = history.Publish("player:founder", []byte{byte(index)}, now.Add(time.Duration(index)*time.Millisecond))
		_, _ = history.Publish("world", []byte{byte(index)}, now.Add(time.Duration(index)*time.Millisecond))
	}
	if _, recoverable := history.Recover("player:founder", Position{}, now.Add(time.Second)); recoverable {
		t.Fatal("truncated player history reported recoverable")
	}
	world, recoverable := history.Recover("world", Position{}, now.Add(10*time.Second))
	if !recoverable || len(world) != 1 || world[0].Data[0] != byte(299%256) {
		t.Fatalf("world=%+v recoverable=%v", world, recoverable)
	}
	identity := Identity{AccountID: "account", FounderID: "founder"}
	if !Authorized(identity, "player:founder", testMemberships{}) || Authorized(identity, "player:other", testMemberships{}) ||
		!Authorized(identity, "guild:guild.allowed", testMemberships{}) || Authorized(identity, "match:match.denied", testMemberships{}) {
		t.Fatal("channel authorization mismatch")
	}
}

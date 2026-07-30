package transport

import (
	"errors"
	"testing"
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
	if err := queue.ReservePlayer(0); !errors.Is(err, ErrInvalidReservation) {
		t.Fatalf("invalid reservation err=%v", err)
	}
	for revision := int64(1); revision <= 100; revision++ {
		queue.ReserveWorld(revision)
	}
	for revision := int64(1); revision < 100; revision++ {
		if queue.AllowWorldWrite(revision) {
			t.Fatalf("stale world revision %d was allowed", revision)
		}
	}
	if !queue.AllowWorldWrite(100) {
		t.Fatal("latest world revision was dropped")
	}
	for index := 0; index < 64; index++ {
		if err := queue.ReservePlayer(int64(index%2 + 7)); err != nil {
			t.Fatal(err)
		}
	}
	if err := queue.ReservePlayer(9); !errors.Is(err, ErrQueueOverflow) {
		t.Fatalf("overflow err=%v", err)
	}
	if queue.FinishPlayer(99) {
		t.Fatal("unreserved player frame decremented the queue")
	}
	if !queue.FinishPlayer(7) {
		t.Fatal("reserved player frame did not decrement the queue")
	}
	if err := queue.ReservePlayer(9); err != nil {
		t.Fatalf("finished receipt did not free one slot: %v", err)
	}
	queue.ResetPlayer()
	for index := 0; index < 64; index++ {
		if err := queue.ReservePlayer(10); err != nil {
			t.Fatalf("reset left stale reservations at %d: %v", index, err)
		}
	}
}

func TestWorldReservationRollbackRestoresPriorRevision(t *testing.T) {
	queue, _ := NewConnectionQueue(phase0Policy(t))
	queue.ReserveWorld(10)
	previous := queue.ReserveWorld(11)
	queue.RollbackWorld(11, previous)
	if !queue.AllowWorldWrite(10) || queue.AllowWorldWrite(9) {
		t.Fatal("world rollback did not restore revision 10")
	}
}

func TestAuthorization(t *testing.T) {
	identity := Identity{AccountID: "account", FounderID: "founder"}
	if !Authorized(identity, "player:founder", testMemberships{}) || Authorized(identity, "player:other", testMemberships{}) ||
		!Authorized(identity, "guild:guild.allowed", testMemberships{}) || Authorized(identity, "match:match.denied", testMemberships{}) {
		t.Fatal("channel authorization mismatch")
	}
}

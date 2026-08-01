package save

import (
	"bytes"
	"testing"
)

func TestPlayerOutboxIdentityMigrationDownRemainsStreamScoped(t *testing.T) {
	data, err := embeddedMigrations.ReadFile("migrations/00041_player_outbox_stream_identity.sql")
	if err != nil {
		t.Fatal(err)
	}
	downMarker := []byte("-- +goose Down")
	parts := bytes.Split(data, downMarker)
	if len(parts) != 2 {
		t.Fatalf("migration Down sections=%d", len(parts)-1)
	}
	down := parts[1]
	streamScoped := []byte("UNIQUE (message_kind,stream_id,source_id)")
	global := []byte("UNIQUE (message_kind,source_id)")
	if !bytes.Contains(down, streamScoped) || bytes.Contains(down, global) {
		t.Fatalf("Down must preserve stream-scoped source identity:\n%s", down)
	}
}

package save

import (
	"encoding/json"
	"testing"
	"time"
)

func testReplayInputs(t *testing.T, command ReplayCommand, at time.Time) json.RawMessage {
	t.Helper()
	encoded, err := json.Marshal(map[string]any{
		"v": ReplayInputsVersion, "command": command, "evaluated_at_ms": at.UnixMilli(),
		"evaluation_mode": "online", "resolved": map[string]any{"kind": "test"},
	})
	if err != nil {
		t.Fatal(err)
	}
	return encoded
}

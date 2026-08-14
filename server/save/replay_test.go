package save

import (
	"encoding/json"
	"strings"
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

func TestValidateReplayInputsPinsOfflineCatchupCoordinates(t *testing.T) {
	command := ReplayCommand{IntentID: "01986666-ac02-7000-8000-000000000001", CompanyStreamID: "01986666-ac02-4000-8000-000000000002",
		FounderID: "01986666-ac02-4000-8000-000000000003", Revision: 1, RunSeq: 1, RunLogSeq: 1}
	at := time.Date(2026, 8, 12, 8, 0, 0, 0, time.UTC)
	value := map[string]any{"v": ReplayInputsVersion, "command": command, "evaluated_at_ms": at.UnixMilli(), "evaluation_mode": "online",
		"offline_catchup": map[string]any{"opened_at_ms": at.UnixMilli(), "offline_span": map[string]any{"from_ms": at.Add(-48 * time.Hour).UnixMilli(), "to_ms": at.UnixMilli()}},
		"resolved":        map[string]any{"kind": "test"}}
	encoded, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ValidateReplayInputs(encoded, command); err != nil {
		t.Fatal(err)
	}
	value["offline_catchup"].(map[string]any)["opened_at_ms"] = at.Add(-time.Millisecond).UnixMilli()
	tampered, _ := json.Marshal(value)
	if _, err := ValidateReplayInputs(tampered, command); err == nil || !strings.Contains(err.Error(), "invalid replay inputs") {
		t.Fatalf("drifted offline catchup accepted: %v", err)
	}
}

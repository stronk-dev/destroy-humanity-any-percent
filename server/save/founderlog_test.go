package save

import (
	"encoding/json"
	"testing"
)

func TestValidateFounderReplayInputs(t *testing.T) {
	command := FounderReplayCommand{IntentID: "01990000-1000-7000-8000-000000000001",
		FounderStreamID: "01990000-1000-4000-8000-000000000002",
		FounderID:       "01990000-1000-4000-8000-000000000003",
		Revision:        4, FounderLogSeq: 7, ServerTSMS: 1_786_000_000_000}
	valid, err := json.Marshal(map[string]any{"v": FounderReplayInputsVersion, "command": command,
		"evaluated_at_ms": command.ServerTSMS, "resolved": map[string]any{"kind": "fixture"}})
	if err != nil {
		t.Fatal(err)
	}
	wantNormalized, err := normalizeJSON(valid)
	if err != nil {
		t.Fatal(err)
	}
	if normalized, err := ValidateFounderReplayInputs(valid, command); err != nil || string(normalized) != string(wantNormalized) {
		t.Fatalf("valid normalized=%s err=%v", normalized, err)
	}

	tests := map[string]func(map[string]any){
		"wrong version": func(value map[string]any) { value["v"] = 2 },
		"wrong timestamp": func(value map[string]any) {
			value["evaluated_at_ms"] = command.ServerTSMS + 1
		},
		"unknown key":        func(value map[string]any) { value["ambient_company_state"] = true },
		"nonobject resolved": func(value map[string]any) { value["resolved"] = []any{} },
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			value := map[string]any{"v": FounderReplayInputsVersion, "command": command,
				"evaluated_at_ms": command.ServerTSMS, "resolved": map[string]any{"kind": "fixture"}}
			mutate(value)
			encoded, err := json.Marshal(value)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := ValidateFounderReplayInputs(encoded, command); err == nil {
				t.Fatal("invalid Founder replay envelope accepted")
			}
		})
	}

	wrongCommand := command
	wrongCommand.Revision++
	if _, err := ValidateFounderReplayInputs(valid, wrongCommand); err == nil {
		t.Fatal("mismatched authoritative command accepted")
	}
}

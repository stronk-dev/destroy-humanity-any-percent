package leaderboard

import (
	"encoding/json"
	"os"
	"testing"
)

func TestTerminalFounderAgeVersionAdmission(t *testing.T) {
	data, err := os.ReadFile("../../testdata/replay/run-ended-age-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	var corpus struct {
		Cases []struct {
			Name   string `json:"name"`
			Schema int    `json:"expected_schema_version"`
			Event  string `json:"ended_event_json"`
		} `json:"cases"`
	}
	if err = json.Unmarshal(data, &corpus); err != nil || len(corpus.Cases) != 5 {
		t.Fatal("missing vectors", err)
	}
	for _, row := range corpus.Cases {
		t.Run(row.Name, func(t *testing.T) {
			var event struct {
				Payload json.RawMessage `json:"payload"`
			}
			if err := json.Unmarshal([]byte(row.Event), &event); err != nil {
				t.Fatal(err)
			}
			var terminal terminalRun
			if decodeProjectorStrict(event.Payload, &terminal) != nil || !validTerminalEventSchema(row.Schema, event.Payload, terminal) {
				t.Fatal("honest producer version refused")
			}
			var fields map[string]json.RawMessage
			_ = json.Unmarshal(event.Payload, &fields)
			if row.Schema == 4 {
				honestAge := *terminal.FounderAgeMS
				delete(fields, "founder_age_ms")
				raw, _ := json.Marshal(fields)
				terminal.FounderAgeMS = nil
				if validTerminalEventSchema(4, raw, terminal) {
					t.Fatal("missing age admitted")
				}
				for _, bad := range []int64{-1, 0, 9_007_199_254_740_992} {
					terminal.FounderAgeMS = &bad
					if validTerminalEventSchema(4, event.Payload, terminal) {
						t.Fatal("invalid age admitted")
					}
				}
				terminal.FounderAgeMS = &honestAge
				fields["FOUNDER_AGE_MS"], _ = json.Marshal(honestAge)
				raw, _ = json.Marshal(fields)
				if validTerminalEventSchema(4, raw, terminal) {
					t.Fatal("new age case alias admitted")
				}

			} else {
				fields["FOUNDER_AGE_MS"] = json.RawMessage(`null`)
				raw, _ := json.Marshal(fields)
				if validTerminalEventSchema(row.Schema, raw, terminal) {
					t.Fatal("legacy age alias admitted")
				}
			}
		})
	}
}

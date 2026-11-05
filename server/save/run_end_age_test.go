package save

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestRunEndedFounderAgeEventAdmission(t *testing.T) {
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
		t.Fatal("missing age corpus", err)
	}
	for _, row := range corpus.Cases {
		t.Run(row.Name, func(t *testing.T) {
			var wire struct {
				Kind     EventKind       `json:"kind"`
				Schema   int             `json:"schema_version"`
				IntentID string          `json:"intent_id"`
				Payload  json.RawMessage `json:"payload"`
			}
			if err := json.Unmarshal([]byte(row.Event), &wire); err != nil {
				t.Fatal(err)
			}
			event := EventWrite{Kind: wire.Kind, SchemaVersion: wire.Schema, IntentID: wire.IntentID, Payload: wire.Payload}
			if !validEventSchemaVersion(event) || validateEventPayload(event) != nil {
				t.Fatal("honest producer event refused")
			}
			var fields map[string]json.RawMessage
			if err := json.Unmarshal(event.Payload, &fields); err != nil {
				t.Fatal(err)
			}
			valid := func(raw []byte, version int) bool {
				candidate := event
				candidate.Payload = raw
				candidate.SchemaVersion = version
				return validEventSchemaVersion(candidate) && validateEventPayload(candidate) == nil
			}
			if row.Schema == 4 {
				age := fields["founder_age_ms"]
				for _, bad := range []string{"null", "true", "\"3601000\"", "-1", "9007199254740992", "1.5", "1e3", "0"} {
					fields["founder_age_ms"] = json.RawMessage(bad)
					raw, _ := json.Marshal(fields)
					if valid(raw, 4) {
						t.Fatalf("admitted invalid age %s", bad)
					}
				}
				fields["founder_age_ms"] = age
				raw, _ := json.Marshal(fields)
				duplicate := strings.TrimSuffix(string(raw), "}") + `,"founder_age_ms":` + string(age) + `}`
				if valid([]byte(duplicate), 4) {
					t.Fatal("duplicate age admitted")
				}
				for _, key := range []string{"Founder_Age_MS", "FOUNDER_AGE_MS"} {
					delete(fields, "founder_age_ms")
					fields[key] = age
					raw, _ = json.Marshal(fields)
					if valid(raw, 4) {
						t.Fatal("age alias admitted")
					}
					delete(fields, key)
					fields["founder_age_ms"] = age
				}
				delete(fields, "founder_age_ms")
				raw, _ = json.Marshal(fields)
				if valid(raw, 4) {
					t.Fatal("missing age admitted")
				}
				fields["founder_age_ms"] = age
				raw, _ = json.Marshal(fields)
				for _, legacy := range []int{2, 3} {
					if valid(raw, legacy) {
						t.Fatal("v4 age admitted under legacy schema")
					}
				}
				fields["attended_ms"] = json.RawMessage(`0`)
				for _, boundary := range []string{"0", "9007199254740991"} {
					fields["founder_age_ms"] = json.RawMessage(boundary)
					raw, _ = json.Marshal(fields)
					if !valid(raw, 4) {
						t.Fatalf("valid exact age boundary %s refused", boundary)
					}
				}
			} else {
				fields["founder_age_ms"] = json.RawMessage(`null`)
				raw, _ := json.Marshal(fields)
				if valid(raw, row.Schema) {
					t.Fatal("legacy age extension admitted")
				}
				delete(fields, "founder_age_ms")
				fields["FOUNDER_AGE_MS"] = json.RawMessage(`null`)
				raw, _ = json.Marshal(fields)
				if valid(raw, row.Schema) {
					t.Fatal("legacy null alias admitted")
				}
			}
		})
	}
}

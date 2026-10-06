package save

import (
	"encoding/json"
	"errors"
	"math"
	"os"
	"strings"
	"testing"

	"cloud-clicker/server/decimal"
)

const reputationEventIntent = "01986666-9901-7000-8000-000000000001"

func reputationEventFields() map[string]any {
	return map[string]any{"node_id": "reputation.unlock.p05", "node_kind": "bonus_unlock", "cost": int64(1),
		"reputation_level": int64(3), "reputation_spent_before": int64(0), "reputation_spent_after": int64(1),
		"unlock_ppm_after": int64(0), "source": "direct"}
}

func reputationEventJSON(t *testing.T, value any) []byte {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

// Exercise the decision validator actually called before ordinary/Exit event
// writes. This is not a database execution or a client-request exploit claim.
func reputationEventDecision(payload []byte) IntentDecision {
	return IntentDecision{Outcome: IntentApplied, Receipt: []byte(`{"outcome":"applied"}`), Events: []EventWrite{{
		Kind: EventReputationNodePurchased, SchemaVersion: 1, IntentID: reputationEventIntent, Payload: payload,
	}}}
}

func TestReputationPurchaseEventAdmission(t *testing.T) {
	type row struct {
		name string
		raw  []byte
	}
	negative := []row{}
	add := func(name string, value any) { negative = append(negative, row{name, reputationEventJSON(t, value)}) }
	keys := []string{"node_id", "node_kind", "cost", "reputation_level", "reputation_spent_before", "reputation_spent_after", "unlock_ppm_after", "source"}
	for _, key := range keys {
		value := reputationEventFields()
		delete(value, key)
		add("missing/"+key, value)
		value = reputationEventFields()
		value[key] = nil
		add("null/"+key, value)
		value = reputationEventFields()
		value[strings.ToUpper(key)] = value[key]
		delete(value, key)
		add("alias/"+key, value)
		value = reputationEventFields()
		raw := reputationEventJSON(t, value)
		duplicate := string(raw[:len(raw)-1]) + "," + string(reputationEventJSON(t, key)) + ":" + string(reputationEventJSON(t, value[key])) + "}"
		negative = append(negative, row{"duplicate/" + key, []byte(duplicate)})
	}
	for _, key := range []string{"cost", "reputation_level", "reputation_spent_before", "reputation_spent_after", "unlock_ppm_after"} {
		lower, upper := int64(-1), int64(decimal.MaxExactInteger+1)
		if key == "cost" {
			lower = 0
		}
		if key == "unlock_ppm_after" {
			upper = 1_000_001
		}
		for _, edge := range []struct {
			name  string
			value int64
		}{{"lower", lower}, {"upper", upper}} {
			value := reputationEventFields()
			value[key] = edge.value
			if key == "cost" && edge.name == "lower" {
				value["reputation_spent_after"] = int64(0)
			}
			if key == "reputation_spent_before" && edge.name == "lower" {
				value["reputation_spent_after"] = int64(0)
			}
			add("domain/"+key+"/"+edge.name, value)
		}
	}
	value := reputationEventFields()
	value["reputation_spent_after"] = int64(2)
	add("relationship/wrong_sum", value)
	value = reputationEventFields()
	value["cost"], value["reputation_spent_after"] = int64(4), int64(4)
	add("relationship/overspend", value)
	for _, level := range []int64{3, -1} {
		value = reputationEventFields()
		value["cost"], value["reputation_spent_before"], value["reputation_spent_after"] = int64(math.MaxInt64), int64(1), int64(math.MinInt64)
		value["reputation_level"] = level
		name := "overflow/valid_earned"
		if level < 0 {
			name = "overflow/negative_earned"
		}
		add(name, value)
	}
	for _, bad := range []struct{ key, value string }{{"node_kind", "other"}, {"source", "other"}, {"node_id", "Not Mechanical"}, {"extra", "unknown"}} {
		value = reputationEventFields()
		value[bad.key] = bad.value
		add("closed/"+bad.key, value)
	}
	base := string(reputationEventJSON(t, reputationEventFields()))
	for _, shape := range []struct{ name, raw string }{{"null", "null"}, {"array", "[]"}, {"empty", "{}"}, {"trailing_object", base + "{}"}, {"trailing_garbage", base + "!"}} {
		negative = append(negative, row{"shape/" + shape.name, []byte(shape.raw)})
	}
	if len(negative) != 55 {
		t.Fatal("incomplete predeclared event population")
	}
	for _, row := range negative {
		t.Run(row.name, func(t *testing.T) {
			if err := validateIntentDecision(reputationEventDecision(row.raw), reputationEventIntent); !errors.Is(err, ErrInvalidStream) {
				t.Fatalf("malformed purchase event admitted or wrong error: %v", err)
			}
		})
	}
	for _, kind := range []string{"bonus_unlock", "starter"} {
		for _, source := range []string{"direct", "exit_plan"} {
			t.Run("valid/"+kind+"/"+source, func(t *testing.T) {
				value := reputationEventFields()
				value["node_kind"], value["source"] = kind, source
				if err := validateIntentDecision(reputationEventDecision(reputationEventJSON(t, value)), reputationEventIntent); err != nil {
					t.Fatal(err)
				}
			})
		}
	}
	for _, edge := range []string{"max_cost", "max_spent"} {
		t.Run("valid/"+edge, func(t *testing.T) {
			value := reputationEventFields()
			value["reputation_level"], value["reputation_spent_after"], value["unlock_ppm_after"] = int64(decimal.MaxExactInteger), int64(decimal.MaxExactInteger), int64(1_000_000)
			if edge == "max_cost" {
				value["cost"] = int64(decimal.MaxExactInteger)
			} else {
				value["reputation_spent_before"] = int64(decimal.MaxExactInteger - 1)
			}
			if err := validateIntentDecision(reputationEventDecision(reputationEventJSON(t, value)), reputationEventIntent); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestReputationPurchaseProducerEventsRemainAdmitted(t *testing.T) {
	data, err := os.ReadFile("../../testdata/replay/reputation-tree-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	var corpus struct {
		Cases []struct {
			Name   string `json:"name"`
			Events string `json:"events_json"`
		} `json:"cases"`
	}
	if err := json.Unmarshal(data, &corpus); err != nil {
		t.Fatal(err)
	}
	count := 0
	for _, row := range corpus.Cases {
		var events []struct {
			Kind          EventKind       `json:"kind"`
			SchemaVersion int             `json:"schema_version"`
			IntentID      string          `json:"intent_id"`
			Payload       json.RawMessage `json:"payload"`
		}
		if err := json.Unmarshal([]byte(row.Events), &events); err != nil {
			t.Fatal(err)
		}
		for _, event := range events {
			if event.Kind != EventReputationNodePurchased {
				continue
			}
			count++
			t.Run(row.Name, func(t *testing.T) {
				decision := reputationEventDecision(event.Payload)
				decision.Events[0].SchemaVersion, decision.Events[0].IntentID = event.SchemaVersion, event.IntentID
				if err := validateIntentDecision(decision, event.IntentID); err != nil {
					t.Fatal(err)
				}
			})
		}
	}
	if count != 11 {
		t.Fatalf("producer purchase-event census=%d want11", count)
	}
}

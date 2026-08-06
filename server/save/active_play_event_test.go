package save

import (
	"encoding/json"
	"testing"
)

func TestActivePlayClaimedEventUnionRejectsKeySubstitution(t *testing.T) {
	valid := []map[string]any{
		{
			"opportunity_id":   "01986666-0a01-7000-8000-000000000001",
			"effect_row_id":    "active.production",
			"selected_target":  nil,
			"buff_instance_id": "01986666-0a02-7000-8000-000000000002",
		},
		{
			"opportunity_id":        "01986666-0a01-7000-8000-000000000001",
			"effect_row_id":         "active.lucky",
			"selected_target":       nil,
			"requested_delta":       "1e1",
			"actual_credited_delta": "5e0",
			"saturated":             true,
			"cap_reason_key":        "cap.cash",
		},
	}
	for _, payload := range valid {
		encoded, err := json.Marshal(payload)
		if err != nil {
			t.Fatal(err)
		}
		if err := validateEventPayload(EventWrite{Kind: EventOpportunityClaimed, SchemaVersion: 1, Payload: encoded}); err != nil {
			t.Fatalf("valid payload rejected: %v", err)
		}
	}

	smuggled := map[string]any{
		"opportunity_id":  "01986666-0a01-7000-8000-000000000001",
		"effect_row_id":   "active.production",
		"selected_target": nil,
		"unrelated":       "01986666-0a02-7000-8000-000000000002",
	}
	encoded, err := json.Marshal(smuggled)
	if err != nil {
		t.Fatal(err)
	}
	if err := validateEventPayload(EventWrite{Kind: EventOpportunityClaimed, SchemaVersion: 1, Payload: encoded}); err == nil {
		t.Fatal("same-cardinality key substitution accepted")
	}
}

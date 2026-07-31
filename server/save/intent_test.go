package save

import (
	"encoding/json"
	"errors"
	"testing"
)

const testIntentID = "018f6b7c-9abc-7def-8abc-0123456789ab"

func TestValidateIntentDecisionEventRegistry(t *testing.T) {
	valid := IntentDecision{
		Outcome: IntentApplied,
		Receipt: json.RawMessage(`{"outcome":"applied","new_revision":2}`),
		Events: []EventWrite{{
			Kind: EventGeneratorPurchased, SchemaVersion: 1, IntentID: testIntentID,
			Payload: json.RawMessage(`{"generator_id":"generator.example","count":2,"cost_resource_id":"company.cash","cost":"2e0"}`),
		}},
	}
	if err := validateIntentDecision(valid, testIntentID); err != nil {
		t.Fatal(err)
	}
	tests := []IntentDecision{
		{Outcome: "unknown", Receipt: json.RawMessage(`{}`)},
		{Outcome: IntentApplied, Receipt: json.RawMessage(`[]`)},
		{Outcome: IntentRejected, Receipt: json.RawMessage(`{}`), Events: valid.Events},
		{Outcome: IntentApplied, Receipt: json.RawMessage(`{}`), Events: []EventWrite{{
			Kind: EventGeneratorPurchased, SchemaVersion: 1, IntentID: testIntentID,
			Payload: json.RawMessage(`{"generator_id":"generator.example","count":0,"cost_resource_id":"company.cash","cost":"0"}`),
		}}},
		{Outcome: IntentApplied, Receipt: json.RawMessage(`{}`), Events: []EventWrite{{
			Kind: "manual_batch_applied", SchemaVersion: 1, IntentID: testIntentID, Payload: json.RawMessage(`{}`),
		}}},
	}
	for index, decision := range tests {
		if err := validateIntentDecision(decision, testIntentID); !errors.Is(err, ErrInvalidStream) {
			t.Fatalf("case %d error = %v", index, err)
		}
	}
}

func TestValidateRouteEventPayloads(t *testing.T) {
	run := `{"company_stream_id":"11111111-1111-4111-8111-111111111111","run_seq":1}`
	events := []EventWrite{
		{Kind: EventGateCrossed, SchemaVersion: 1, IntentID: testIntentID, Payload: json.RawMessage(`{"gate_id":"gate.example","route_id":null,"run_id":` + run + `,"founder_id":"22222222-2222-4222-8222-222222222222"}`)},
		{Kind: EventRouteExecuted, SchemaVersion: 1, IntentID: testIntentID, Payload: json.RawMessage(`{"route_id":"route.example","gate_id":"gate.example","run_id":` + run + `,"founder_id":"22222222-2222-4222-8222-222222222222"}`)},
		{Kind: EventRouteHintPurchased, SchemaVersion: 1, IntentID: testIntentID, Payload: json.RawMessage(`{"route_id":"route.example","cost":50}`)},
		{Kind: EventRouteKnowledgeGranted, SchemaVersion: 1, IntentID: testIntentID, Payload: json.RawMessage(`{"route_id":"route.example","amount":25,"source":"founder_first"}`)},
	}
	if err := validateIntentDecision(IntentDecision{Outcome: IntentApplied, Receipt: json.RawMessage(`{}`), Events: events}, testIntentID); err != nil {
		t.Fatal(err)
	}
	events[1].Payload = json.RawMessage(`{"route_id":"route.example","gate_id":"gate.example","run_id":` + run + `,"founder_id":"not-a-uuid"}`)
	if err := validateIntentDecision(IntentDecision{Outcome: IntentApplied, Receipt: json.RawMessage(`{}`), Events: events}, testIntentID); !errors.Is(err, ErrInvalidStream) {
		t.Fatalf("invalid route event error=%v", err)
	}
}

func TestValidateCommonsEventPayloads(t *testing.T) {
	run := `{"company_stream_id":"11111111-1111-4111-8111-111111111111","run_seq":1}`
	founder := `"22222222-2222-4222-8222-222222222222"`
	intentEvents := []EventWrite{
		{Kind: EventCompactSigned, SchemaVersion: 1, IntentID: testIntentID, Payload: json.RawMessage(`{"founder_id":` + founder + `,"run_id":` + run + `,"tithe_ppm":100000,"prior_member":false,"new_member":true}`)},
		{Kind: EventCompactTitheRaised, SchemaVersion: 1, IntentID: testIntentID, Payload: json.RawMessage(`{"founder_id":` + founder + `,"run_id":` + run + `,"prior_tithe_ppm":100000,"new_tithe_ppm":130000}`)},
		{Kind: EventCompactSampled, SchemaVersion: 1, IntentID: testIntentID, Payload: json.RawMessage(`{"founder_id":` + founder + `,"run_id":` + run + `,"weight_ppm":1000000,"compliance_ppm":900000,"enclosure":"1e-1","capacity":"1e3","solidarity_ppm":500000,"sampled_ms":3600000}`)},
	}
	if err := validateIntentDecision(IntentDecision{Outcome: IntentApplied, Receipt: json.RawMessage(`{}`), Events: intentEvents}, testIntentID); err != nil {
		t.Fatal(err)
	}
	intentEvents[1].Payload = json.RawMessage(`{"founder_id":` + founder + `,"run_id":` + run + `,"prior_tithe_ppm":130000,"new_tithe_ppm":100000}`)
	if err := validateIntentDecision(IntentDecision{Outcome: IntentApplied, Receipt: json.RawMessage(`{}`), Events: intentEvents}, testIntentID); !errors.Is(err, ErrInvalidStream) {
		t.Fatalf("decreasing tithe event error=%v", err)
	}
	ambient := []EventWrite{
		{Kind: EventCompactHealthBandChanged, SchemaVersion: 1, Payload: json.RawMessage(`{"scope_kind":"server","scope_id":"aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa","from_band":"collapsed","to_band":"strained","health_ppm":400000}`)},
		{Kind: EventCompactRecovered, SchemaVersion: 1, Payload: json.RawMessage(`{"scope_kind":"server","scope_id":"aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa","health_ppm":400000}`)},
		{Kind: EventCompactRecruitmentOffered, SchemaVersion: 1, Payload: json.RawMessage(`{"founder_id":` + founder + `,"reason_key":"compact.recruitment.mid_t3"}`)},
	}
	for _, event := range ambient {
		if err := validateEventPayload(event); err != nil {
			t.Fatalf("event %s: %v", event.Kind, err)
		}
	}
}

func TestValidateFactionEventPayloads(t *testing.T) {
	run := `{"company_stream_id":"11111111-1111-4111-8111-111111111111","run_seq":1}`
	events := []EventWrite{
		{Kind: EventIncorporated, SchemaVersion: 1, IntentID: testIntentID, Payload: json.RawMessage(`{"founder_id":"22222222-2222-4222-8222-222222222222","run_id":` + run + `,"faction_id":"open_source","stock_resource":"libraries","incorporated_at_ms":1785412800000,"compact_auto_signed":true}`)},
		{Kind: EventFactionStockSaturated, SchemaVersion: 1, IntentID: testIntentID, Payload: json.RawMessage(`{"faction_id":"open_source","stock_resource":"libraries","stock_cap":100000}`)},
	}
	if err := validateIntentDecision(IntentDecision{Outcome: IntentApplied, Receipt: json.RawMessage(`{}`), Events: events}, testIntentID); err != nil {
		t.Fatal(err)
	}
	events[1].Payload = json.RawMessage(`{"faction_id":"open_source","stock_resource":"libraries","stock_cap":0}`)
	if err := validateIntentDecision(IntentDecision{Outcome: IntentApplied, Receipt: json.RawMessage(`{}`), Events: events}, testIntentID); !errors.Is(err, ErrInvalidStream) {
		t.Fatalf("invalid faction event error=%v", err)
	}
}

func TestValidateGuildActivityEventPayloads(t *testing.T) {
	run := `{"company_stream_id":"11111111-1111-4111-8111-111111111111","run_seq":1}`
	founder := `"22222222-2222-4222-8222-222222222222"`
	valid := []EventWrite{
		{Kind: EventGuildTitheAccrued, SchemaVersion: 1, Payload: json.RawMessage(`{"founder_id":` + founder + `,"run_id":` + run + `,"progress_delta_ppm":50,"xp_delta":1}`)},
		{Kind: EventGuildActivityEvaluated, SchemaVersion: 1, Payload: json.RawMessage(`{"founder_id":` + founder + `,"run_id":` + run + `,"progress_delta_ppm":0,"xp_delta":0}`)},
	}
	for _, event := range valid {
		if err := validateEventPayload(event); err != nil {
			t.Fatalf("event %s: %v", event.Kind, err)
		}
	}
	invalid := []EventWrite{
		{Kind: EventGuildTitheAccrued, SchemaVersion: 1, Payload: json.RawMessage(`{"founder_id":` + founder + `,"run_id":` + run + `,"progress_delta_ppm":0,"xp_delta":0}`)},
		{Kind: EventGuildActivityEvaluated, SchemaVersion: 1, Payload: json.RawMessage(`{"founder_id":` + founder + `,"run_id":` + run + `,"progress_delta_ppm":50,"xp_delta":1}`)},
	}
	for _, event := range invalid {
		if err := validateEventPayload(event); !errors.Is(err, ErrInvalidStream) {
			t.Fatalf("invalid event %s: %v", event.Kind, err)
		}
	}
}

func TestUUIDV7AndRequestHashGrammar(t *testing.T) {
	if !uuidV7Pattern.MatchString(testIntentID) {
		t.Fatal("valid UUIDv7 rejected")
	}
	for _, invalid := range []string{
		"018f6b7c-9abc-4def-8abc-0123456789ab",
		"018F6B7C-9ABC-7DEF-8ABC-0123456789AB",
		"not-a-uuid",
	} {
		if uuidV7Pattern.MatchString(invalid) {
			t.Fatalf("invalid UUIDv7 accepted: %s", invalid)
		}
	}
}

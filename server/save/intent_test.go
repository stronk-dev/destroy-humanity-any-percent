package save

import (
	"encoding/json"
	"errors"
	"fmt"
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

func TestValidatePetCareEventRegistry(t *testing.T) {
	decision := IntentDecision{Outcome: IntentApplied, Receipt: json.RawMessage(`{"outcome":"applied"}`), Events: []EventWrite{
		{Kind: EventPetCareApplied, SchemaVersion: 1, IntentID: testIntentID,
			Payload: json.RawMessage(`{"pet_id":"018f6b7c-9abc-7def-8abc-0123456789ac","action_id":"care.feed","stat_id":"hunger","before_ppm":600000,"applied_ppm":50000,"after_ppm":650000,"trust_before_ppm":500000,"trust_after_ppm":501000,"mood":"neutral","status_band":"normal","next_eligible_attended_ms":180000}`)},
		{Kind: EventPetStatusChanged, SchemaVersion: 1, IntentID: testIntentID,
			Payload: json.RawMessage(`{"pet_id":"018f6b7c-9abc-7def-8abc-0123456789ac","from_status_band":"low","to_status_band":"normal"}`)},
	}}
	if err := validateIntentDecision(decision, testIntentID); err != nil {
		t.Fatal(err)
	}
	decision.Events[0].Payload = json.RawMessage(`{"pet_id":"018f6b7c-9abc-7def-8abc-0123456789ac","action_id":"care.feed","stat_id":"hunger","before_ppm":600000,"applied_ppm":50000,"after_ppm":640000,"trust_before_ppm":500000,"trust_after_ppm":501000,"mood":"neutral","status_band":"normal","next_eligible_attended_ms":180000}`)
	if err := validateIntentDecision(decision, testIntentID); err == nil {
		t.Fatal("inconsistent pet care event was accepted")
	}
}

func TestValidateSoulEventRegistryExactPayloads(t *testing.T) {
	const session = "018f6b7c-9abc-7def-8abc-0123456789ac"
	const company = "11111111-1111-4111-8111-111111111111"
	events := []EventWrite{
		{Kind: EventSoulPricePaid, SchemaVersion: 1, IntentID: testIntentID, Payload: json.RawMessage(`{"source_id":"soul.fixture","owner_kind":"fixture","eligibility_ref":"offer.fixture","soul_before":20,"debit":10,"soul_after":10,"band_before":"hollow","band_after":"hollow","curtain_copy_key":"category.valuation"}`)},
		{Kind: EventSoulBandChanged, SchemaVersion: 1, IntentID: testIntentID, Payload: json.RawMessage(`{"soul_before":10,"soul_after":9,"band_before":"hollow","band_after":"near_zero","reason_key":"category.low_percent"}`)},
		{Kind: EventSoulDepleted, SchemaVersion: 1, IntentID: testIntentID, Payload: json.RawMessage(`{"source_id":"soul.fixture","owner_kind":"fixture","eligibility_ref":"offer.fixture","soul_before":10,"soul_after":0,"occurred_at_ms":1786053600000}`)},
		{Kind: EventSoulRecoveryStarted, SchemaVersion: 1, IntentID: testIntentID, Payload: json.RawMessage(`{"session_id":"` + session + `","activity_id":"touch_grass.fixture","company_stream_id":"` + company + `","run_seq":1,"founder_attended_start_ms":100,"required_duration_ms":5000}`)},
		{Kind: EventSoulRecoveryCancelled, SchemaVersion: 1, IntentID: testIntentID, Payload: json.RawMessage(`{"session_id":"` + session + `","activity_id":"touch_grass.fixture","company_stream_id":"` + company + `","run_seq":1,"founder_attended_start_ms":100,"founder_attended_end_ms":200,"soul_before":10,"soul_after":10}`)},
		{Kind: EventSoulRecovered, SchemaVersion: 1, IntentID: testIntentID, Payload: json.RawMessage(`{"session_id":"` + session + `","activity_id":"touch_grass.fixture","company_stream_id":"` + company + `","run_seq":1,"founder_attended_start_ms":100,"founder_attended_end_ms":5100,"soul_before":10,"soul_after":25,"recovery_amount":15,"band_before":"hollow","band_after":"hollow","reason_key":"category.any_percent"}`)},
	}
	if err := validateIntentDecision(IntentDecision{Outcome: IntentApplied, Receipt: json.RawMessage(`{}`), Events: events}, testIntentID); err != nil {
		t.Fatal(err)
	}
	invalid := events[4]
	invalid.Payload = json.RawMessage(`{"session_id":"` + session + `","activity_id":"touch_grass.fixture","company_stream_id":"` + company + `","run_seq":1,"founder_attended_start_ms":100,"founder_attended_end_ms":200,"soul_before":10,"soul_after":10,"recovery_amount":null}`)
	if err := validateEventPayload(invalid); !errors.Is(err, ErrInvalidStream) {
		t.Fatalf("cancel payload with recovered-only key accepted: %v", err)
	}
	invalid = events[5]
	invalid.Payload = json.RawMessage(`{"session_id":"` + session + `","activity_id":"touch_grass.fixture","company_stream_id":"` + company + `","run_seq":1,"founder_attended_start_ms":100,"founder_attended_end_ms":5100,"soul_before":10,"soul_after":25,"recovery_amount":15,"band_before":"hollow","band_after":"hollow"}`)
	if err := validateEventPayload(invalid); !errors.Is(err, ErrInvalidStream) {
		t.Fatalf("recovered payload missing reason key accepted: %v", err)
	}
}

func TestValidateUpgradePurchasedEventPayload(t *testing.T) {
	valid := EventWrite{Kind: EventUpgradePurchased, SchemaVersion: 1, IntentID: testIntentID, Payload: json.RawMessage(`{"upgrade_id":"upgrade.click","cost_resource_id":"company.cash","cost":"1e2"}`)}
	if err := validateEventPayload(valid); err != nil {
		t.Fatal(err)
	}
	invalid := []json.RawMessage{
		json.RawMessage(`{"upgrade_id":"upgrade.click","cost_resource_id":"company.cash","cost":"0"}`),
		json.RawMessage(`{"upgrade_id":"upgrade.click","cost_resource_id":"company.cash","cost":"1e2","extra":true}`),
	}
	for _, payload := range invalid {
		valid.Payload = payload
		if err := validateEventPayload(valid); !errors.Is(err, ErrInvalidStream) {
			t.Fatalf("payload=%s error=%v", payload, err)
		}
	}
}

func TestValidateFoundationEventPayloads(t *testing.T) {
	run := `{"company_stream_id":"11111111-1111-4111-8111-111111111111","run_seq":1}`
	valid := []EventWrite{
		{Kind: EventMeterBandChanged, SchemaVersion: 1, IntentID: testIntentID, Payload: json.RawMessage(`{"run_id":` + run + `,"meter_id":"doom.probability","from_band":"low","to_band":"high","direction":"up","value_before":69,"value_after":71}`)},
		{Kind: EventAchievementEarned, SchemaVersion: 1, IntentID: testIntentID, Payload: json.RawMessage(`{"run_id":` + run + `,"achievement_id":"achievement.first_gate","condition_scope":"run","score_grant":4}`)},
	}
	if err := validateIntentDecision(IntentDecision{Outcome: IntentApplied, Receipt: json.RawMessage(`{}`), Events: valid}, testIntentID); err != nil {
		t.Fatal(err)
	}
	invalid := []EventWrite{
		{Kind: EventMeterBandChanged, SchemaVersion: 1, IntentID: testIntentID, Payload: json.RawMessage(`{"run_id":` + run + `,"meter_id":"doom.probability","from_band":"low","to_band":"high","direction":"down","value_before":69,"value_after":71}`)},
		{Kind: EventAchievementEarned, SchemaVersion: 1, IntentID: testIntentID, Payload: json.RawMessage(`{"run_id":` + run + `,"achievement_id":"achievement.first_gate","condition_scope":"founder","score_grant":4}`)},
	}
	for _, event := range invalid {
		if err := validateEventPayload(event); !errors.Is(err, ErrInvalidStream) {
			t.Fatalf("invalid event %s: %v", event.Kind, err)
		}
	}
}

func TestValidateDoctrineAndComputeCreditEventPayloads(t *testing.T) {
	run := `{"company_stream_id":"11111111-1111-4111-8111-111111111111","run_seq":1}`
	events := []EventWrite{
		{Kind: EventDoctrinePicked, SchemaVersion: 1, IntentID: testIntentID, Payload: json.RawMessage(`{"founder_id":"22222222-2222-4222-8222-222222222222","run_id":` + run + `,"transition_id":"transition.t3_to_t4","doctrine_id":"doctrine.capture"}`)},
		{Kind: EventComputeCreditSpent, SchemaVersion: 1, IntentID: testIntentID, Payload: json.RawMessage(`{"founder_id":"22222222-2222-4222-8222-222222222222","run_id":` + run + `,"amount_ms":1500,"target":"accelerate","burst_duration_ms":1500,"burst_speed":"2e0"}`)},
	}
	if err := validateIntentDecision(IntentDecision{Outcome: IntentApplied, Receipt: json.RawMessage(`{}`), Events: events}, testIntentID); err != nil {
		t.Fatal(err)
	}
	events[1].Payload = json.RawMessage(`{"founder_id":"22222222-2222-4222-8222-222222222222","run_id":` + run + `,"amount_ms":1500,"target":"accelerate","burst_duration_ms":1499,"burst_speed":"2e0"}`)
	if err := validateEventPayload(events[1]); !errors.Is(err, ErrInvalidStream) {
		t.Fatalf("mismatched burst duration error=%v", err)
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

func TestRunEndedV2RequiresTerminalCategoryFacts(t *testing.T) {
	base := `{"founder_id":"22222222-2222-4222-8222-222222222222","run_id":{"company_stream_id":"11111111-1111-4111-8111-111111111111","run_seq":1},"exit_type":"collapse","started_at_ms":100,"ended_at_ms":200,"rta_ms":100,"attended_ms":90,"pre_timer":false,"terminal_seq":1,"payout":{"reputation_delta":0,"network_slot_unlocks":[],"route_knowledge":0,"clout_reach_note":"clout.reach.preserved"},"tier":1,"lifetime_value":"1e0","ledger_fact_kinds":[],"executed_routes":[],%s"assisted":{"commons":false,"advisor":false},"faction":null}`
	valid := EventWrite{Kind: EventRunEnded, SchemaVersion: 2, IntentID: testIntentID,
		Payload: json.RawMessage(fmt.Sprintf(base, `"gates_crossed":[],"generators_purchased_total":0,`))}
	if err := validateIntentDecision(IntentDecision{Outcome: IntentApplied, Receipt: json.RawMessage(`{}`), Events: []EventWrite{valid}}, testIntentID); err != nil {
		t.Fatal(err)
	}
	for _, fields := range []string{`"generators_purchased_total":0,`, `"gates_crossed":[],`} {
		invalid := valid
		invalid.Payload = json.RawMessage(fmt.Sprintf(base, fields))
		if err := validateIntentDecision(IntentDecision{Outcome: IntentApplied, Receipt: json.RawMessage(`{}`), Events: []EventWrite{invalid}}, testIntentID); !errors.Is(err, ErrInvalidStream) {
			t.Fatalf("missing v2 terminal fact accepted: %s err=%v", fields, err)
		}
	}
}

func TestAcceptedExitOfferResolutionPayloadIsExact(t *testing.T) {
	valid := EventWrite{Kind: EventExitOfferResolved, SchemaVersion: 1, IntentID: testIntentID,
		Payload: json.RawMessage(`{"offer_id":"018f6b7c-9abc-7def-8abc-0123456789ac","resolution":"accepted"}`)}
	if err := validateIntentDecision(IntentDecision{Outcome: IntentApplied, Receipt: json.RawMessage(`{}`), Events: []EventWrite{valid}}, testIntentID); err != nil {
		t.Fatal(err)
	}
	for _, payload := range []string{
		`{"offer_id":"018f6b7c-9abc-7def-8abc-0123456789ac","resolution":"declined"}`,
		`{"offer_id":"018f6b7c-9abc-7def-8abc-0123456789ac"}`,
		`{"offer_id":"018f6b7c-9abc-7def-8abc-0123456789ac","resolution":"accepted","run_seq":1}`,
	} {
		invalid := valid
		invalid.Payload = json.RawMessage(payload)
		if err := validateIntentDecision(IntentDecision{Outcome: IntentApplied, Receipt: json.RawMessage(`{}`), Events: []EventWrite{invalid}}, testIntentID); !errors.Is(err, ErrInvalidStream) {
			t.Fatalf("invalid accepted resolution payload accepted: %s err=%v", payload, err)
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

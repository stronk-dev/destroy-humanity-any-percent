package save

import "testing"

func TestFiscalEventPayloadsAreClosed(t *testing.T) {
	automatic := EventWrite{Kind: EventFiscalPeriodHarvested, SchemaVersion: 1, Payload: []byte(`{"source":"automatic","periods":2,"credit_before":1,"credited":6,"credit_after":7,"opened_before_ms":1000,"opened_after_ms":1600,"seq_before":4,"seq_after":6,"saturated":false,"hardcap_reason_key":"fiscal.credit.hardcap"}`)}
	manual := EventWrite{Kind: EventFiscalPeriodHarvested, SchemaVersion: 1, Payload: []byte(`{"source":"manual","outcome":"guaranteed","credit_before":1,"credit_after":4,"period_opened_wall_ms_before":1000,"period_opened_wall_ms_after":1200,"seq_before":4,"seq_after":5,"draw_ppm":null,"saturated":false}`)}
	spend := EventWrite{Kind: EventFiscalCreditSpent, SchemaVersion: 1, Payload: []byte(`{"target":{"kind":"generator_level","generator_id":"generator.example","levels":2},"resolved_cost":3,"fiscal_credit_before":7,"fiscal_credit_after":4}`)}
	for _, event := range []EventWrite{automatic, manual, spend} {
		if err := validateEventPayload(event); err != nil {
			t.Fatalf("valid %s payload rejected: %v", event.Kind, err)
		}
	}
	automatic.Payload = []byte(`{"source":"automatic","periods":2,"credit_before":1,"credited":6,"credit_after":8,"opened_before_ms":1000,"opened_after_ms":1600,"seq_before":4,"seq_after":6,"saturated":false,"hardcap_reason_key":"fiscal.credit.hardcap"}`)
	if err := validateEventPayload(automatic); err == nil {
		t.Fatal("inconsistent automatic credit accepted")
	}
	spend.Payload = []byte(`{"target":{"kind":"unlock","unlock_id":"unlock.arcade","extra":true},"resolved_cost":3,"fiscal_credit_before":7,"fiscal_credit_after":4}`)
	if err := validateEventPayload(spend); err == nil {
		t.Fatal("non-exact Fiscal target accepted")
	}
}

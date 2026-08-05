package production

import (
	"encoding/json"
	"errors"
	"testing"

	"cloud-clicker/server/doctrine"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/save"
)

func doctrineIntentCatalog(t *testing.T) *doctrine.Catalog {
	t.Helper()
	catalog, err := doctrine.LoadCatalog([]byte(`{"schema_version":1,"transitions":[{"transition_id":"transition.t3_to_t4","source_tier":3,"gate_id":"gate.t3_to_t4","doctrine_ids":["doctrine.capture","doctrine.ethical"]}]}`))
	if err != nil {
		t.Fatal(err)
	}
	return catalog
}

func TestDoctrineAndComputeIntentGrammar(t *testing.T) {
	pick, err := ParseIntent([]byte(`{"intent_id":"018f6b7c-9abc-7def-8abc-0123456789ab","kind":"pick_doctrine","expected_revision":1,"transition_id":"transition.t3_to_t4","doctrine_id":"doctrine.capture"}`))
	if err != nil || pick.InvalidDetail != "" || pick.TransitionID != "transition.t3_to_t4" || pick.DoctrineID != "doctrine.capture" {
		t.Fatalf("pick=%+v err=%v", pick, err)
	}
	spend, err := ParseIntent([]byte(`{"intent_id":"018f6b7c-9abc-7def-8abc-0123456789ac","kind":"spend_compute_credit","expected_revision":2,"amount_ms":1500,"target":"accelerate"}`))
	if err != nil || spend.InvalidDetail != "" || spend.AmountMS != 1500 || spend.Target != "accelerate" {
		t.Fatalf("spend=%+v err=%v", spend, err)
	}
	malformed, err := ParseIntent([]byte(`{"intent_id":"018f6b7c-9abc-7def-8abc-0123456789ad","kind":"spend_compute_credit","expected_revision":2,"amount_ms":0,"target":"accelerate"}`))
	if err != nil || malformed.InvalidDetail != "amount_ms" {
		t.Fatalf("malformed=%+v err=%v", malformed, err)
	}
}

func TestDoctrinePickAndComputeSpendTransitions(t *testing.T) {
	catalog := foundationCatalog(t)
	state := foundationState(t, catalog, engineCursor)
	state.WireVersion, state.Tier, state.ComputeCreditMS = 17, 3, 5_000
	revision := save.Revision{StreamID: "018f6b7c-9abc-4def-8abc-0123456789ab", OwnerID: "018f6b7c-9abc-4def-8abc-0123456789ac", Number: 1}
	pick := IntentRequest{IntentID: "018f6b7c-9abc-7def-8abc-0123456789ab", Kind: IntentPickDoctrine, TransitionID: "transition.t3_to_t4", DoctrineID: "doctrine.capture"}
	decision, err := transitionWithSimulationPolicy(pick, state, catalog, nil, doctrineIntentCatalog(t), nil, nil, revision, ModeOnline, engineCursor, nil, nil, nil, nil)
	if err != nil || decision.Outcome != save.IntentApplied || state.DoctrinesByTransition[pick.TransitionID] != pick.DoctrineID || len(decision.Events) != 1 || decision.Events[0].Kind != save.EventDoctrinePicked {
		t.Fatalf("pick decision=%+v state=%+v err=%v", decision, state.DoctrinesByTransition, err)
	}
	repeated, err := transitionWithSimulationPolicy(pick, state, catalog, nil, doctrineIntentCatalog(t), nil, nil, save.Revision{Number: 2, StreamID: revision.StreamID, OwnerID: revision.OwnerID}, ModeOnline, engineCursor, nil, nil, nil, nil)
	if err != nil || repeated.Outcome != save.IntentRejected || !receiptHasRejection(repeated.Receipt, "not_eligible", "doctrine_already_picked") {
		t.Fatalf("repeated=%s err=%v", repeated.Receipt, err)
	}

	spend := IntentRequest{IntentID: "018f6b7c-9abc-7def-8abc-0123456789ac", Kind: IntentSpendComputeCredit, AmountMS: 1_500, Target: "accelerate"}
	decision, err = transitionWithSimulationPolicy(spend, state, catalog, nil, doctrineIntentCatalog(t), nil, nil, save.Revision{Number: 2, StreamID: revision.StreamID, OwnerID: revision.OwnerID}, ModeOnline, engineCursor, nil, nil, nil, nil)
	if err != nil || decision.Outcome != save.IntentApplied || state.ComputeCreditMS != 3_500 || state.ComputeBurstRemainingMS != 1_500 || len(decision.Events) != 1 || decision.Events[0].Kind != save.EventComputeCreditSpent {
		t.Fatalf("spend decision=%+v credit=%d burst=%d err=%v", decision, state.ComputeCreditMS, state.ComputeBurstRemainingMS, err)
	}
	active, err := transitionWithSimulationPolicy(spend, state, catalog, nil, doctrineIntentCatalog(t), nil, nil, save.Revision{Number: 3, StreamID: revision.StreamID, OwnerID: revision.OwnerID}, ModeOnline, engineCursor, nil, nil, nil, nil)
	if err != nil || active.Outcome != save.IntentRejected || !receiptHasRejection(active.Receipt, "not_eligible", "burst_active") {
		t.Fatalf("active=%s err=%v", active.Receipt, err)
	}

	state.ComputeBurstRemainingMS, state.ComputeCreditMS = 0, 100
	insufficient, err := transitionWithSimulationPolicy(spend, state, catalog, nil, doctrineIntentCatalog(t), nil, nil, save.Revision{Number: 3, StreamID: revision.StreamID, OwnerID: revision.OwnerID}, ModeOnline, engineCursor, nil, nil, nil, nil)
	if err != nil || insufficient.Outcome != save.IntentRejected || !receiptHasRejection(insufficient.Receipt, "not_eligible", "compute_credit_balance") {
		t.Fatalf("insufficient=%s err=%v", insufficient.Receipt, err)
	}

	state.Ledger, _ = economy.NewLedger(catalog, economy.ScopeFounder)
	if _, err := transitionWithSimulationPolicy(spend, state, catalog, nil, doctrineIntentCatalog(t), nil, nil, revision, ModeOnline, engineCursor, nil, nil, nil, nil); !errors.Is(err, ErrInvalidEngineState) {
		t.Fatalf("Founder spend error=%v", err)
	}
}

func receiptHasRejection(receipt []byte, category, detail string) bool {
	var value struct {
		Outcome   string `json:"outcome"`
		Rejection struct {
			Category string `json:"category"`
			Detail   string `json:"detail"`
		} `json:"rejection"`
	}
	return json.Unmarshal(receipt, &value) == nil && value.Outcome == "rejected" && value.Rejection.Category == category && value.Rejection.Detail == detail
}

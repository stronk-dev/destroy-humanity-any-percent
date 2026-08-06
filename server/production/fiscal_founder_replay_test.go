package production

import (
	"encoding/json"
	"testing"
	"time"

	"cloud-clicker/server/save"
)

func fiscalReplayState(t *testing.T, bundle CatalogBundle, opened int64) *save.State {
	t.Helper()
	state := replayFounderFixtureState(t, bundle, time.UnixMilli(opened).UTC())
	if err := activateFounderFeatureState(state, bundle, 19, opened); err != nil {
		t.Fatal(err)
	}
	state.WireVersion = 19
	return state
}

func TestApplyFounderLoggedFiscalAutoSweepAndRejectedRollback(t *testing.T) {
	_, foundations := foundationTestBundles(t)
	_, pets := founderFeatureBundles(t, foundations)
	bundle := fiscalFeatureBundle(t, pets)
	founderID := "01987778-1000-7000-8000-000000000001"
	streamID := "01987778-1000-4000-8000-000000000002"
	intentID := "01987778-1000-7000-8000-000000000003"
	opened := int64(1_800_000_000_000)
	state := fiscalReplayState(t, bundle, opened)
	now := opened + bundle.Fiscal.Clock.AutoMS
	command := save.FounderReplayCommand{IntentID: intentID, FounderStreamID: streamID, FounderID: founderID,
		Revision: 1, FounderLogSeq: 1, ServerTSMS: now}
	resolved := founderFiscalHarvestResolved{Kind: IntentHarvestFiscalPeriod, NowWallMS: now,
		PeriodOpenedWallMSBefore: opened, PeriodsSwept: 1, SeqBefore: 0, Outcome: "consumed_by_auto"}
	inputs, err := save.MarshalFounderReplayInputs(command, resolved)
	if err != nil {
		t.Fatal(err)
	}
	payload := []byte(`{"expected_revision":1,"kind":"harvest_fiscal_period"}`)
	transition, err := ApplyFounderLogged(state, payload, bundle, inputs)
	if err != nil {
		t.Fatal(err)
	}
	if transition.Outcome != save.IntentApplied || state.FiscalCredit != bundle.Fiscal.Credit.CreditPerPeriod ||
		state.FiscalPeriodSequence != 1 || state.FiscalPeriodOpenedWallMS != now || len(transition.Events) != 1 ||
		transition.Events[0].Kind != save.EventFiscalPeriodHarvested {
		t.Fatalf("transition=%+v state=%+v", transition, state)
	}
	var receipt map[string]json.RawMessage
	if json.Unmarshal(transition.Receipt, &receipt) != nil || string(receipt["fiscal_sweep"]) == "null" {
		t.Fatalf("missing Fiscal sweep receipt: %s", transition.Receipt)
	}

	rejectedID := "01987778-1000-7000-8000-000000000004"
	command.IntentID, command.Revision, command.FounderLogSeq, command.ServerTSMS = rejectedID, 2, 2, now
	spendResolved := founderFiscalSpendResolved{Kind: IntentSpendFiscalCredit,
		Target: fiscalTargetWire{Kind: "unlock", UnlockID: "unlock.arcade"}, ResolvedCost: 5}
	inputs, err = save.MarshalFounderReplayInputs(command, spendResolved)
	if err != nil {
		t.Fatal(err)
	}
	payload = []byte(`{"expected_revision":2,"kind":"spend_fiscal_credit","target":{"kind":"unlock","unlock_id":"unlock.arcade"}}`)
	creditBefore := state.FiscalCredit
	transition, err = ApplyFounderLogged(state, payload, bundle, inputs)
	if err != nil {
		t.Fatal(err)
	}
	if transition.Outcome != save.IntentRejected || state.FiscalCredit != creditBefore || len(transition.Events) != 0 {
		t.Fatalf("rejected transition mutated state: %+v state=%+v", transition, state)
	}
}

func TestFiscalFounderCommandRejectsClockRegression(t *testing.T) {
	_, foundations := foundationTestBundles(t)
	_, pets := founderFeatureBundles(t, foundations)
	bundle := fiscalFeatureBundle(t, pets)
	opened := time.Now().UTC().UnixMilli()
	state := fiscalReplayState(t, bundle, opened)
	command := save.FounderReplayCommand{IntentID: "01987778-2000-7000-8000-000000000001",
		FounderStreamID: "01987778-2000-4000-8000-000000000002", FounderID: "01987778-2000-4000-8000-000000000003",
		Revision: 1, FounderLogSeq: 1, ServerTSMS: opened - 1}
	inputs, err := save.MarshalFounderReplayInputs(command, founderFiscalHarvestResolved{Kind: IntentHarvestFiscalPeriod,
		NowWallMS: opened - 1, PeriodOpenedWallMSBefore: opened, SeqBefore: 0, Outcome: "rejected"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ApplyFounderLogged(state, []byte(`{"expected_revision":1,"kind":"harvest_fiscal_period"}`), bundle, inputs); err == nil {
		t.Fatal("clock regression accepted")
	}
}

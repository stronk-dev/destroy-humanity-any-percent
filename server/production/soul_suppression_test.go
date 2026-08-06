package production

import (
	"encoding/json"
	"testing"
	"time"

	"cloud-clicker/server/economy"
	"cloud-clicker/server/save"
)

func TestApplySuppressedLoggedConsumesTimeWithoutOutput(t *testing.T) {
	_, foundations := foundationTestBundles(t)
	_, withPets := founderFeatureBundles(t, foundations)
	catalogs := soulFeatureBundle(t, fiscalFeatureBundle(t, withPets))
	now := time.Date(2026, 8, 6, 10, 0, 0, 0, time.UTC)
	state := foundationScopeState(t, catalogs.Economy, economy.ScopeCompany)
	state.WireVersion = 16
	state.RunSeq = 1
	state.RunStartedAt = now
	state.EvaluatedThrough = now
	state.ManualTokenRefilledAt = now
	state.GeneratorCounts = map[string]int64{}
	state.GeneratorProvisioned = map[string]int64{}
	for _, definition := range catalogs.Economy.GeneratorClassesForScope(economy.ScopeCompany) {
		state.GeneratorCounts[definition.ID] = 0
		state.GeneratorProvisioned[definition.ID] = 0
	}
	state.GeneratorCounts["generator.beige_tower"] = 3
	state.ProvisionRemaindersPPM = map[string]int64{}
	state.MeterValues = map[string]int{"trust.customer_standing": 500_000}
	state.MeterDecayRemainders = map[string]int64{"trust.customer_standing": 7}
	state.MeterInputRemainders = map[string]int64{"trust.customer_standing": 11}
	state.AchievementsEarnedRun = map[string]bool{}
	beforeCash, _ := state.Ledger.Balance("company.cash")
	command := save.ReplayCommand{IntentID: "01986666-6800-7000-8000-000000000001",
		CompanyStreamID: "01986666-6800-4000-8000-000000000002", FounderID: "01986666-6800-4000-8000-000000000003",
		Revision: 4, RunSeq: 1, RunLogSeq: 2}
	suppression := soulSuppression{FromEvaluatedMS: now.UnixMilli(), ToEvaluatedMS: now.Add(2 * time.Minute).UnixMilli(),
		FounderAttendedStart: 40_000, FounderAttendedEnd: 160_000, SessionID: command.IntentID}
	inputs, err := buildSoulSuppressionInputs(command, soulRecoveryResolveKind, suppression, nil, nil, catalogs.Routes.ContextVersion(), nil)
	if err != nil {
		t.Fatal(err)
	}
	payload, _ := json.Marshal(soulRecoveryPayload{Kind: soulRecoveryResolveKind, SessionID: command.IntentID})
	result, err := ApplySuppressedLogged(state, payload, catalogs, inputs)
	if err != nil {
		t.Fatal(err)
	}
	afterCash, _ := result.State.Ledger.Balance("company.cash")
	if !afterCash.Eq(beforeCash) || !result.State.EvaluatedThrough.Equal(now.Add(2*time.Minute)) {
		t.Fatalf("cash before=%s after=%s evaluated=%s", beforeCash, afterCash, result.State.EvaluatedThrough)
	}
	if result.State.GeneratorProvisioned["generator.beige_tower"] != 0 ||
		result.State.MeterValues["trust.customer_standing"] != 500_000 ||
		result.State.MeterDecayRemainders["trust.customer_standing"] != 7 ||
		result.State.MeterInputRemainders["trust.customer_standing"] != 11 || len(result.State.AchievementsEarnedRun) != 0 {
		t.Fatalf("suppression leaked production-derived state: %+v", result.State)
	}
}

func TestApplySuppressedLoggedRejectsCoordinateDrift(t *testing.T) {
	_, foundations := foundationTestBundles(t)
	_, pets := founderFeatureBundles(t, foundations)
	catalogs := soulFeatureBundle(t, fiscalFeatureBundle(t, pets))
	now := time.Date(2026, 8, 6, 10, 0, 0, 0, time.UTC)
	state := foundationScopeState(t, catalogs.Economy, economy.ScopeCompany)
	state.WireVersion, state.RunSeq, state.RunStartedAt, state.EvaluatedThrough, state.ManualTokenRefilledAt = 16, 1, now, now, now
	state.GeneratorCounts, state.GeneratorProvisioned = map[string]int64{}, map[string]int64{}
	for _, definition := range catalogs.Economy.GeneratorClassesForScope(economy.ScopeCompany) {
		state.GeneratorCounts[definition.ID] = 0
		state.GeneratorProvisioned[definition.ID] = 0
	}
	state.ProvisionRemaindersPPM = map[string]int64{}
	command := save.ReplayCommand{IntentID: "01986666-6801-7000-8000-000000000001",
		CompanyStreamID: "01986666-6801-4000-8000-000000000002", FounderID: "01986666-6801-4000-8000-000000000003",
		Revision: 4, RunSeq: 1, RunLogSeq: 2}
	suppression := soulSuppression{FromEvaluatedMS: now.Add(time.Millisecond).UnixMilli(), ToEvaluatedMS: now.Add(time.Second).UnixMilli(),
		FounderAttendedStart: 0, FounderAttendedEnd: 1_000, SessionID: command.IntentID}
	inputs, err := buildSoulSuppressionInputs(command, soulRecoveryCancelKind, suppression, nil, nil, catalogs.Routes.ContextVersion(), nil)
	if err != nil {
		t.Fatal(err)
	}
	payload, _ := json.Marshal(soulRecoveryPayload{Kind: soulRecoveryCancelKind, SessionID: command.IntentID})
	if _, err := ApplySuppressedLogged(state, payload, catalogs, inputs); err == nil {
		t.Fatal("coordinate drift accepted")
	}
}

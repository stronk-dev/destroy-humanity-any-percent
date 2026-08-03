package production

import (
	"encoding/json"
	"math/big"
	"os"
	"testing"
	"time"

	"cloud-clicker/server/decimal"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/routes"
	"cloud-clicker/server/save"
)

func foundationCatalog(t *testing.T) *economy.Catalog {
	t.Helper()
	data, err := os.ReadFile("../../testdata/economy-foundation-v4.json")
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := economy.LoadCatalog(data)
	if err != nil {
		t.Fatal(err)
	}
	return catalog
}

func foundationState(t *testing.T, catalog *economy.Catalog, now time.Time) *save.State {
	t.Helper()
	ledger, err := economy.NewLedger(catalog, economy.ScopeCompany)
	if err != nil {
		t.Fatal(err)
	}
	return &save.State{
		Ledger: ledger, GeneratorCounts: map[string]int64{"generator.high": 0, "generator.low": 0},
		UpgradesOwned: map[string]bool{}, GeneratorProvisioned: map[string]int64{"generator.high": 0, "generator.low": 0},
		ProvisionRemaindersPPM: map[string]int64{"generator.low": 0}, EvaluatedThrough: now, ManualTokenRefilledAt: now,
		RunStartedAt: now, RunSeq: 1, GatesCrossed: map[string]bool{}, DoctrinesByTransition: map[string]string{},
		LedgerFactKinds: map[string]bool{}, MeterBands: map[string]int{}, RegionTraits: map[string]bool{}, HintsUnlocked: map[string]bool{},
	}
}

func foundationRoutes(t *testing.T) *routes.Catalog {
	t.Helper()
	data, err := os.ReadFile("../../balance/routes/phase0.json")
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := routes.LoadCatalog(data)
	if err != nil {
		t.Fatal(err)
	}
	return catalog
}

func TestContentContributionsUsePurchasedCountsAndRawSourceOrder(t *testing.T) {
	catalog := foundationCatalog(t)
	state := foundationState(t, catalog, time.Date(2026, 8, 3, 10, 0, 0, 0, time.UTC))
	state.GeneratorCounts["generator.low"] = 10
	state.GeneratorProvisioned["generator.low"] = 900
	state.UpgradesOwned["upgrade.click"] = true
	contributions, err := contentContributions(state, catalog)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{
		"generator.low.ladder.purchased_10":             "2e0",
		"generator.low.role.manual_output.manual.click": "1.01e0",
		"pool.operations":                               "1.012e0",
		"upgrade.click.factor":                          "2e0",
	}
	if len(contributions) != len(want) {
		t.Fatalf("contributions=%+v", contributions)
	}
	last := ""
	for _, contribution := range contributions {
		if got := contributionKey(contribution); got <= last {
			t.Fatalf("contribution order %q after %q", got, last)
		} else {
			last = got
		}
		if contribution.Factor.String() != want[contribution.SourceID] {
			t.Fatalf("source %s factor=%s want=%s", contribution.SourceID, contribution.Factor, want[contribution.SourceID])
		}
	}
}

func TestTwoSynergyPoolsComposeInRawSourceOrder(t *testing.T) {
	data, err := os.ReadFile("../../testdata/economy-foundation-v4.json")
	if err != nil {
		t.Fatal(err)
	}
	var authored map[string]any
	if err := json.Unmarshal(data, &authored); err != nil {
		t.Fatal(err)
	}
	generators := authored["generator_classes"].([]any)
	high := generators[1].(map[string]any)
	high["roles"] = append(high["roles"].([]any), map[string]any{"kind": "synergy_feed", "pool_id": "pool.scale"})
	authored["synergy_pools"] = append(authored["synergy_pools"].([]any), map[string]any{
		"id": "pool.scale", "slot": "upgrades", "curve": "log",
		"sources": []any{map[string]any{"kind": "generator", "id_or_class": "generator.high", "per_count_ppm": float64(1_000_000)}},
	})
	authored["multiplier_sources"] = append(authored["multiplier_sources"].([]any), map[string]any{
		"id": "pool.scale", "slot": "upgrades", "target": "generator.high", "provider": "pool.scale",
	})
	encoded, err := json.Marshal(authored)
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := economy.LoadCatalog(encoded)
	if err != nil {
		t.Fatal(err)
	}
	state := foundationState(t, catalog, time.Date(2026, 8, 3, 10, 0, 0, 0, time.UTC))
	state.GeneratorCounts["generator.low"] = 10
	state.GeneratorCounts["generator.high"] = 5
	contributions, err := contentContributions(state, catalog)
	if err != nil {
		t.Fatal(err)
	}
	var pools []string
	for _, contribution := range contributions {
		if contribution.SourceID == "pool.operations" || contribution.SourceID == "pool.scale" {
			pools = append(pools, contribution.SourceID+"="+contribution.Factor.String())
		}
	}
	want := []string{"pool.operations=1.01e0", "pool.scale=1.77815125038e0"}
	if len(pools) != len(want) || pools[0] != want[0] || pools[1] != want[1] {
		t.Fatalf("pool composition=%v want=%v", pools, want)
	}
}

func TestProvisionBucketsArePartitionInvariantAndNewUnitsProduceNextBucket(t *testing.T) {
	catalog := foundationCatalog(t)
	started := time.Date(2026, 8, 3, 10, 0, 0, 0, time.UTC)
	one := foundationState(t, catalog, started)
	split := foundationState(t, catalog, started)
	one.GeneratorCounts["generator.high"] = 2
	split.GeneratorCounts["generator.high"] = 2
	if _, err := Evaluate(one, catalog, started.Add(180*time.Second), ModeOnline, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := Evaluate(split, catalog, started.Add(61*time.Second), ModeOnline, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := Evaluate(split, catalog, started.Add(180*time.Second), ModeOnline, nil); err != nil {
		t.Fatal(err)
	}
	if one.GeneratorProvisioned["generator.low"] != 3 || split.GeneratorProvisioned["generator.low"] != 3 || one.ProvisionRemaindersPPM["generator.low"] != 0 || split.ProvisionRemaindersPPM["generator.low"] != 0 {
		t.Fatalf("one provision=%v/%v split=%v/%v", one.GeneratorProvisioned, one.ProvisionRemaindersPPM, split.GeneratorProvisioned, split.ProvisionRemaindersPPM)
	}
	oneCash, _ := one.Ledger.Balance("company.cash")
	splitCash, _ := split.Ledger.Balance("company.cash")
	if !oneCash.Eq(splitCash) || oneCash.String() != "3.78e3" {
		t.Fatalf("one cash=%s split cash=%s", oneCash, splitCash)
	}
}

func TestProvisionBoundaryCarriesRemainderAndSaturatesOnlyAtDeclaredCap(t *testing.T) {
	catalog := foundationCatalog(t)
	purchased := map[string]int64{"generator.high": 1, "generator.low": 0}
	provisioned := map[string]int64{"generator.high": 0, "generator.low": 0}
	remainders := map[string]int64{"generator.low": 0}
	if err := materializeProvisionBoundary(catalog, purchased, provisioned, remainders); err != nil {
		t.Fatal(err)
	}
	if provisioned["generator.low"] != 0 || remainders["generator.low"] != 500_000 {
		t.Fatalf("first boundary provision=%v remainder=%v", provisioned, remainders)
	}
	if err := materializeProvisionBoundary(catalog, purchased, provisioned, remainders); err != nil {
		t.Fatal(err)
	}
	if provisioned["generator.low"] != 1 || remainders["generator.low"] != 0 {
		t.Fatalf("second boundary provision=%v remainder=%v", provisioned, remainders)
	}
	provisioned["generator.low"] = 9_007_199_254_740_990
	purchased["generator.high"] = 4
	if err := materializeProvisionBoundary(catalog, purchased, provisioned, remainders); err != nil {
		t.Fatal(err)
	}
	if provisioned["generator.low"] != 9_007_199_254_740_991 {
		t.Fatalf("saturated provision=%d", provisioned["generator.low"])
	}
}

func TestBuyUpgradeAppliesCostOwnershipAndTypedEvent(t *testing.T) {
	catalog := foundationCatalog(t)
	routeCatalog := foundationRoutes(t)
	now := time.Date(2026, 8, 3, 10, 0, 0, 0, time.UTC)
	state := foundationState(t, catalog, now)
	if _, err := state.Ledger.ApplyAccrual(economy.Transaction{Entries: []economy.Entry{{ResourceID: "company.cash", Delta: decimal.FromString("1e3")}}}); err != nil {
		t.Fatal(err)
	}
	request := IntentRequest{IntentID: "01986666-0200-7000-8000-000000000200", Kind: IntentBuyUpgrade, ExpectedRevision: 1, UpgradeID: "upgrade.click"}
	revision := save.Revision{Number: 1, OwnerID: "11111111-1111-4111-8111-111111111111", StreamID: "22222222-2222-4222-8222-222222222222"}
	decision, err := TransitionWithPolicies(request, state, catalog, routeCatalog, nil, nil, revision, ModeOnline, now, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if decision.Outcome != save.IntentApplied || !state.UpgradesOwned["upgrade.click"] || len(decision.Events) != 1 || decision.Events[0].Kind != save.EventUpgradePurchased {
		t.Fatalf("decision=%+v owned=%v", decision, state.UpgradesOwned)
	}
	balance, _ := state.Ledger.Balance("company.cash")
	if balance.String() != "9e2" {
		t.Fatalf("balance=%s", balance)
	}
	duplicate, err := TransitionWithPolicies(request, state, catalog, routeCatalog, nil, nil, save.Revision{Number: 2}, ModeOnline, now, nil, nil, nil)
	var duplicateReceipt struct {
		Rejection struct {
			Category string `json:"category"`
			Detail   string `json:"detail"`
		} `json:"rejection"`
	}
	decodeErr := json.Unmarshal(duplicate.Receipt, &duplicateReceipt)
	if err != nil || decodeErr != nil || duplicate.Outcome != save.IntentRejected || duplicateReceipt.Rejection.Category != "not_eligible" || duplicateReceipt.Rejection.Detail != "owned" {
		t.Fatalf("duplicate=%+v err=%v", duplicate, err)
	}
}

func TestManualOutputRoleUsesPurchasedCountOnly(t *testing.T) {
	catalog := foundationCatalog(t)
	now := time.Date(2026, 8, 3, 10, 0, 0, 0, time.UTC)
	state := foundationState(t, catalog, now)
	state.GeneratorCounts["generator.low"] = 10
	state.GeneratorProvisioned["generator.low"] = 900
	state.ManualTokenMilli = 50_000
	contributions, err := assembleContributions(state, catalog, nil)
	if err != nil {
		t.Fatal(err)
	}
	request := IntentRequest{IntentID: "01986666-0201-7000-8000-000000000201", Kind: IntentPerformManualBatch, ExpectedRevision: 1, ActionID: "manual.click", Count: 1, WindowMS: 1}
	decision, err := TransitionWithPolicies(request, state, catalog, nil, nil, nil, save.Revision{Number: 1}, ModeOnline, now, contributions, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	var receipt struct {
		AppliedCount int64 `json:"applied_count"`
	}
	if err := json.Unmarshal(decision.Receipt, &receipt); err != nil {
		t.Fatal(err)
	}
	if decision.Outcome != save.IntentApplied || receipt.AppliedCount != 1 {
		t.Fatalf("decision=%+v", decision)
	}
	balance, _ := state.Ledger.Balance("company.cash")
	if balance.String() != "1.01e0" {
		t.Fatalf("manual output=%s", balance)
	}
}

func TestSimulationMaskNullsWholeGeneratorOutputWithoutChangingOwnership(t *testing.T) {
	catalog := foundationCatalog(t)
	now := time.Date(2026, 8, 3, 10, 0, 0, 0, time.UTC)
	state := foundationState(t, catalog, now)
	state.GeneratorCounts["generator.low"] = 10
	state.ManualTokenMilli = 50_000
	request := IntentRequest{IntentID: "01986666-0202-7000-8000-000000000202", Kind: IntentPerformManualBatch, ExpectedRevision: 1, ActionID: "manual.click", Count: 1, WindowMS: 1}
	decision, err := SimulateTransition(request, state, catalog, save.Revision{Number: 1}, ModeOnline, now, nil, nil, AblationMask{GeneratorIDs: []string{"generator.low"}})
	if err != nil || decision.Outcome != save.IntentApplied {
		t.Fatalf("decision=%+v err=%v", decision, err)
	}
	balance, _ := state.Ledger.Balance("company.cash")
	if balance.String() != "1e0" || state.GeneratorCounts["generator.low"] != 10 {
		t.Fatalf("balance=%s counts=%v", balance, state.GeneratorCounts)
	}
}

func TestSimulationMaskNullsProvisionEdgeAndRemovedActionRejectsBeforeAccrual(t *testing.T) {
	catalog := foundationCatalog(t)
	started := time.Date(2026, 8, 3, 10, 0, 0, 0, time.UTC)
	state := foundationState(t, catalog, started)
	state.GeneratorCounts["generator.high"] = 2
	state.ManualTokenMilli = 50_000
	request := IntentRequest{IntentID: "01986666-0203-7000-8000-000000000203", Kind: IntentPerformManualBatch, ExpectedRevision: 1, ActionID: "manual.click", Count: 1, WindowMS: 1}
	decision, err := SimulateTransition(request, state, catalog, save.Revision{Number: 1}, ModeOnline, started.Add(180*time.Second), nil, nil, AblationMask{GeneratorIDs: []string{"generator.high"}})
	if err != nil || decision.Outcome != save.IntentApplied || state.GeneratorProvisioned["generator.low"] != 0 {
		t.Fatalf("decision=%+v provisioned=%v err=%v", decision, state.GeneratorProvisioned, err)
	}
	balance, _ := state.Ledger.Balance("company.cash")
	if balance.String() != "1e0" {
		t.Fatalf("masked provision output=%s", balance)
	}

	rejectedState := foundationState(t, catalog, started)
	rejectedState.GeneratorCounts["generator.low"] = 10
	rejectedState.ManualTokenMilli = 50_000
	rejected, err := SimulateTransition(request, rejectedState, catalog, save.Revision{Number: 1}, ModeOnline, started.Add(time.Minute), nil, nil, AblationMask{RemovedActionIDs: []string{"manual.click"}})
	if err != nil || rejected.Outcome != save.IntentRejected || !rejectedState.EvaluatedThrough.Equal(started) {
		t.Fatalf("rejected=%+v evaluated=%s err=%v", rejected, rejectedState.EvaluatedThrough, err)
	}
}

func TestSimulationMaskAppliesAcrossDeclineOfferEvaluation(t *testing.T) {
	catalog := foundationCatalog(t)
	started := time.Date(2026, 8, 3, 10, 0, 0, 0, time.UTC)
	state := foundationState(t, catalog, started)
	state.GeneratorCounts["generator.low"] = 1
	state.OfferState = &save.ExitOfferState{
		OfferID:   "01986666-0204-7000-8000-000000000204",
		ExitType:  "collapse",
		SpawnedAt: started,
		ExpiresAt: started.Add(time.Hour),
	}
	request := IntentRequest{
		IntentID:         "01986666-0205-7000-8000-000000000205",
		Kind:             IntentDeclineExitOffer,
		ExpectedRevision: 1,
		OfferID:          state.OfferState.OfferID,
	}
	decision, err := SimulateTransition(request, state, catalog, save.Revision{Number: 1}, ModeOnline, started.Add(time.Second), nil, nil, AblationMask{GeneratorIDs: []string{"generator.low"}})
	if err != nil || decision.Outcome != save.IntentApplied || state.OfferState != nil {
		t.Fatalf("decision=%+v offer=%+v err=%v", decision, state.OfferState, err)
	}
	balance, _ := state.Ledger.Balance("company.cash")
	if !balance.Eq(decimal.Zero) {
		t.Fatalf("masked decline accrued cash=%s", balance)
	}
}

func TestSynergyCurveGoldenFactors(t *testing.T) {
	linear, err := synergyFactor(economy.SynergyLinear, big.NewInt(12_000))
	if err != nil || linear.String() != "1.012e0" {
		t.Fatalf("linear=%s err=%v", linear, err)
	}
	logarithmic, err := synergyFactor(economy.SynergyLog, big.NewInt(12_000))
	if err != nil || logarithmic.String() != "1.0051805125e0" {
		t.Fatalf("log=%s err=%v", logarithmic, err)
	}
}

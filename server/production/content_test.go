package production

import (
	"encoding/json"
	"math/big"
	"os"
	"testing"
	"time"

	"cloud-clicker/server/decimal"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/faction"
	"cloud-clicker/server/routes"
	"cloud-clicker/server/save"
)

const foundationConstantsHash = "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

type foundationCatchupPolicies map[string]int64

func (policies foundationCatchupPolicies) ResolveCatchupCeilingMS(hash string) (int64, bool) {
	value, ok := policies[hash]
	return value, ok
}

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

func foundationCatalogWithMutation(t *testing.T, mutate func(map[string]any)) *economy.Catalog {
	t.Helper()
	data, err := os.ReadFile("../../testdata/economy-foundation-v4.json")
	if err != nil {
		t.Fatal(err)
	}
	var authored map[string]any
	if err := json.Unmarshal(data, &authored); err != nil {
		t.Fatal(err)
	}
	mutate(authored)
	encoded, err := json.Marshal(authored)
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := economy.LoadCatalog(encoded)
	if err != nil {
		t.Fatal(err)
	}
	return catalog
}

func foundationHighSynergyCatalog(t *testing.T, target string) *economy.Catalog {
	t.Helper()
	return foundationCatalogWithMutation(t, func(authored map[string]any) {
		generators := authored["generator_classes"].([]any)
		low := generators[0].(map[string]any)
		filtered := make([]any, 0)
		for _, value := range low["roles"].([]any) {
			if value.(map[string]any)["kind"] != "synergy_feed" {
				filtered = append(filtered, value)
			}
		}
		low["roles"] = filtered
		high := generators[1].(map[string]any)
		high["roles"] = append(high["roles"].([]any), map[string]any{"kind": "synergy_feed", "pool_id": "pool.operations"})
		pool := authored["synergy_pools"].([]any)[0].(map[string]any)
		pool["sources"] = []any{map[string]any{"kind": "generator", "id_or_class": "generator.high", "per_count_ppm": float64(1000)}}
		declaration := authored["multiplier_sources"].([]any)[0].(map[string]any)
		declaration["target"] = target
	})
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

func foundationFactionCatalog(t *testing.T) *faction.Catalog {
	t.Helper()
	data, err := os.ReadFile("../../balance/factions/phase0.json")
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := faction.LoadCatalog(data, faction.CompactTitheBand{MinimumPPM: 50_000, DefaultPPM: 100_000, MaximumPPM: 150_000})
	if err != nil {
		t.Fatal(err)
	}
	return catalog
}

func rejectionDetail(t *testing.T, decision save.IntentDecision) (string, string) {
	t.Helper()
	var receipt struct {
		Rejection struct {
			Category string `json:"category"`
			Detail   string `json:"detail"`
		} `json:"rejection"`
	}
	if err := json.Unmarshal(decision.Receipt, &receipt); err != nil {
		t.Fatal(err)
	}
	return receipt.Rejection.Category, receipt.Rejection.Detail
}

func hasRoleActivation(values []RoleActivation, want RoleActivation) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
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

func TestBuyUpgradeEligibilityUsesPostAccrualStateAndTypedRejections(t *testing.T) {
	catalog := foundationCatalog(t)
	routeCatalog := foundationRoutes(t)
	started := time.Date(2026, 8, 3, 10, 0, 0, 0, time.UTC)
	request := IntentRequest{IntentID: "01986666-0206-7000-8000-000000000206", Kind: IntentBuyUpgrade, ExpectedRevision: 1, UpgradeID: "upgrade.click"}
	revision := save.Revision{Number: 1}

	requires := foundationState(t, catalog, started)
	decision, err := TransitionWithPolicies(request, requires, catalog, routeCatalog, nil, nil, revision, ModeOnline, started, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	category, detail := rejectionDetail(t, decision)
	if decision.Outcome != save.IntentRejected || category != "not_eligible" || detail != "requires" {
		t.Fatalf("requires decision=%+v category=%s detail=%s", decision, category, detail)
	}

	window := foundationState(t, catalog, started)
	if _, err := window.Ledger.ApplyAccrual(economy.Transaction{Entries: []economy.Entry{{ResourceID: "company.cash", Delta: decimal.FromString("1e3")}}}); err != nil {
		t.Fatal(err)
	}
	window.GatesCrossed["gate.t0_to_t1"] = true
	decision, err = TransitionWithPolicies(request, window, catalog, routeCatalog, nil, nil, revision, ModeOnline, started, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	category, detail = rejectionDetail(t, decision)
	if decision.Outcome != save.IntentRejected || category != "not_eligible" || detail != "window" {
		t.Fatalf("window decision=%+v category=%s detail=%s", decision, category, detail)
	}

	accrued := foundationState(t, catalog, started)
	accrued.GeneratorCounts["generator.high"] = 1
	decision, err = TransitionWithPolicies(request, accrued, catalog, routeCatalog, nil, nil, revision, ModeOnline, started.Add(10*time.Second), nil, nil, nil)
	if err != nil || decision.Outcome != save.IntentApplied || !accrued.UpgradesOwned["upgrade.click"] {
		t.Fatalf("post-accrual decision=%+v owned=%v err=%v", decision, accrued.UpgradesOwned, err)
	}
	balance, _ := accrued.Ledger.Balance("company.cash")
	if !balance.Eq(decimal.Zero) {
		t.Fatalf("post-accrual balance=%s", balance)
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
	result, err := SimulateTransition(request, state, catalog, SimulationDependencies{Routes: foundationRoutes(t)}, save.Revision{Number: 1}, ModeOnline, now, nil, nil, AblationMask{GeneratorIDs: []string{"generator.low"}})
	if err != nil || result.Decision.Outcome != save.IntentApplied {
		t.Fatalf("decision=%+v err=%v", result.Decision, err)
	}
	balance, _ := state.Ledger.Balance("company.cash")
	if balance.String() != "1e0" || state.GeneratorCounts["generator.low"] != 10 {
		t.Fatalf("balance=%s counts=%v", balance, state.GeneratorCounts)
	}
}

func TestSimulationUpgradeEffectMaskPreservesPurchaseAndNullsOutput(t *testing.T) {
	catalog := foundationCatalog(t)
	dependencies := SimulationDependencies{Routes: foundationRoutes(t)}
	now := time.Date(2026, 8, 3, 10, 0, 0, 0, time.UTC)
	state := foundationState(t, catalog, now)
	if _, err := state.Ledger.ApplyAccrual(economy.Transaction{Entries: []economy.Entry{{ResourceID: "company.cash", Delta: decimal.FromString("1e3")}}}); err != nil {
		t.Fatal(err)
	}
	buy := IntentRequest{IntentID: "01986666-0207-7000-8000-000000000207", Kind: IntentBuyUpgrade, ExpectedRevision: 1, UpgradeID: "upgrade.click"}
	mask := AblationMask{UpgradeIDs: []string{"upgrade.click"}}
	result, err := SimulateTransition(buy, state, catalog, dependencies, save.Revision{Number: 1}, ModeOnline, now, nil, nil, mask)
	if err != nil || result.Decision.Outcome != save.IntentApplied || !state.UpgradesOwned["upgrade.click"] {
		t.Fatalf("buy result=%+v owned=%v err=%v", result, state.UpgradesOwned, err)
	}
	state.ManualTokenMilli = 50_000
	manual := IntentRequest{IntentID: "01986666-0208-7000-8000-000000000208", Kind: IntentPerformManualBatch, ExpectedRevision: 2, ActionID: "manual.click", Count: 1, WindowMS: 1}
	result, err = SimulateTransition(manual, state, catalog, dependencies, save.Revision{Number: 2}, ModeOnline, now, nil, nil, mask)
	if err != nil || result.Decision.Outcome != save.IntentApplied {
		t.Fatalf("manual result=%+v err=%v", result, err)
	}
	balance, _ := state.Ledger.Balance("company.cash")
	if balance.String() != "9.01e2" {
		t.Fatalf("masked upgrade balance=%s", balance)
	}
}

func TestSimulationActionRemovalRejectsEveryOwnedActionBeforeAccrual(t *testing.T) {
	catalog := foundationCatalog(t)
	dependencies := SimulationDependencies{Routes: foundationRoutes(t)}
	started := time.Date(2026, 8, 3, 10, 0, 0, 0, time.UTC)
	tests := []struct {
		name    string
		request IntentRequest
		mask    AblationMask
		detail  string
	}{
		{name: "generator", request: IntentRequest{IntentID: "01986666-0209-7000-8000-000000000209", Kind: IntentBuyGenerator, ExpectedRevision: 1, GeneratorID: "generator.low", Count: 1, CountMode: "exact"}, mask: AblationMask{RemovedGeneratorIDs: []string{"generator.low"}}, detail: "generator.low"},
		{name: "upgrade", request: IntentRequest{IntentID: "01986666-0210-7000-8000-000000000210", Kind: IntentBuyUpgrade, ExpectedRevision: 1, UpgradeID: "upgrade.click"}, mask: AblationMask{RemovedUpgradeIDs: []string{"upgrade.click"}}, detail: "upgrade.click"},
		{name: "manual", request: IntentRequest{IntentID: "01986666-0211-7000-8000-000000000211", Kind: IntentPerformManualBatch, ExpectedRevision: 1, ActionID: "manual.click", Count: 1, WindowMS: 1}, mask: AblationMask{RemovedActionIDs: []string{"manual.click"}}, detail: "manual.click"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			state := foundationState(t, catalog, started)
			state.GeneratorCounts["generator.low"] = 10
			state.ManualTokenMilli = 50_000
			if _, err := state.Ledger.ApplyAccrual(economy.Transaction{Entries: []economy.Entry{{ResourceID: "company.cash", Delta: decimal.FromString("1e3")}}}); err != nil {
				t.Fatal(err)
			}
			result, err := SimulateTransition(test.request, state, catalog, dependencies, save.Revision{Number: 1}, ModeOnline, started.Add(time.Minute), nil, nil, test.mask)
			if err != nil {
				t.Fatal(err)
			}
			category, detail := rejectionDetail(t, result.Decision)
			if result.Decision.Outcome != save.IntentRejected || category != "unknown_id" || detail != test.detail || !state.EvaluatedThrough.Equal(started) {
				t.Fatalf("result=%+v category=%s detail=%s evaluated=%s", result, category, detail, state.EvaluatedThrough)
			}
		})
	}
}

func TestSimulationProvisionActivationRequiresMaterializedUnits(t *testing.T) {
	catalog := foundationCatalog(t)
	dependencies := SimulationDependencies{Routes: foundationRoutes(t)}
	started := time.Date(2026, 8, 3, 10, 0, 0, 0, time.UTC)
	state := foundationState(t, catalog, started)
	state.GeneratorCounts["generator.high"] = 1
	state.ManualTokenMilli = 50_000
	want := RoleActivation{GeneratorID: "generator.high", Kind: economy.RoleProvision, TargetID: "generator.low"}
	request := IntentRequest{IntentID: "01986666-0213-7000-8000-000000000213", Kind: IntentPerformManualBatch, ExpectedRevision: 1, ActionID: "manual.click", Count: 1, WindowMS: 1}

	first, err := SimulateTransition(request, state, catalog, dependencies, save.Revision{Number: 1}, ModeOnline, started.Add(time.Minute), nil, nil, AblationMask{})
	if err != nil || first.Decision.Outcome != save.IntentApplied || state.ProvisionRemaindersPPM["generator.low"] != 500_000 {
		t.Fatalf("first=%+v remainder=%v err=%v", first, state.ProvisionRemaindersPPM, err)
	}
	if hasRoleActivation(first.RoleActivations, want) {
		t.Fatalf("zero-unit boundary activated provision: %+v", first.RoleActivations)
	}

	request.IntentID = "01986666-0214-7000-8000-000000000214"
	request.ExpectedRevision = 2
	second, err := SimulateTransition(request, state, catalog, dependencies, save.Revision{Number: 2}, ModeOnline, started.Add(2*time.Minute), nil, nil, AblationMask{})
	if err != nil || second.Decision.Outcome != save.IntentApplied || state.GeneratorProvisioned["generator.low"] != 1 || !hasRoleActivation(second.RoleActivations, want) {
		t.Fatalf("second=%+v provisioned=%v err=%v", second, state.GeneratorProvisioned, err)
	}

	state.GeneratorProvisioned["generator.low"] = 9_007_199_254_740_991
	request.IntentID = "01986666-0215-7000-8000-000000000215"
	request.ExpectedRevision = 3
	capped, err := SimulateTransition(request, state, catalog, dependencies, save.Revision{Number: 3}, ModeOnline, started.Add(3*time.Minute), nil, nil, AblationMask{})
	if err != nil || capped.Decision.Outcome != save.IntentApplied {
		t.Fatalf("capped=%+v err=%v", capped, err)
	}
	if hasRoleActivation(capped.RoleActivations, want) {
		t.Fatalf("capped boundary activated provision: %+v", capped.RoleActivations)
	}
}

func TestSimulationSynergyActivationRequiresDeclaredExercisedTarget(t *testing.T) {
	started := time.Date(2026, 8, 3, 10, 0, 0, 0, time.UTC)
	want := RoleActivation{GeneratorID: "generator.high", Kind: economy.RoleSynergyFeed, TargetID: "pool.operations"}
	request := IntentRequest{IntentID: "01986666-0216-7000-8000-000000000216", Kind: IntentPerformManualBatch, ExpectedRevision: 1, ActionID: "manual.click", Count: 1, WindowMS: 1}

	zeroTargetCatalog := foundationCatalogWithMutation(t, func(authored map[string]any) {
		generators := authored["generator_classes"].([]any)
		low := generators[0].(map[string]any)
		filtered := make([]any, 0)
		for _, value := range low["roles"].([]any) {
			if value.(map[string]any)["kind"] != "synergy_feed" {
				filtered = append(filtered, value)
			}
		}
		low["roles"] = filtered
		high := generators[1].(map[string]any)
		high["roles"] = append(high["roles"].([]any), map[string]any{"kind": "synergy_feed", "pool_id": "pool.operations"})
		pool := authored["synergy_pools"].([]any)[0].(map[string]any)
		pool["sources"] = []any{map[string]any{"kind": "generator", "id_or_class": "generator.high", "per_count_ppm": float64(1000)}}
	})
	state := foundationState(t, zeroTargetCatalog, started)
	state.GeneratorCounts["generator.high"] = 1
	state.ManualTokenMilli = 50_000
	result, err := SimulateTransition(request, state, zeroTargetCatalog, SimulationDependencies{Routes: foundationRoutes(t)}, save.Revision{Number: 1}, ModeOnline, started.Add(time.Second), nil, nil, AblationMask{})
	if err != nil || result.Decision.Outcome != save.IntentApplied {
		t.Fatalf("zero target result=%+v err=%v", result, err)
	}
	if hasRoleActivation(result.RoleActivations, want) {
		t.Fatalf("zero-count target activated synergy: %+v", result.RoleActivations)
	}

	undeclaredCatalog := foundationCatalogWithMutation(t, func(authored map[string]any) {
		low := authored["generator_classes"].([]any)[0].(map[string]any)
		filtered := make([]any, 0)
		for _, value := range low["roles"].([]any) {
			if value.(map[string]any)["kind"] != "synergy_feed" {
				filtered = append(filtered, value)
			}
		}
		low["roles"] = filtered
	})
	undeclaredState := foundationState(t, undeclaredCatalog, started)
	undeclaredState.GeneratorCounts["generator.low"] = 1
	undeclaredState.ManualTokenMilli = 50_000
	undeclared, err := SimulateTransition(request, undeclaredState, undeclaredCatalog, SimulationDependencies{Routes: foundationRoutes(t)}, save.Revision{Number: 1}, ModeOnline, started.Add(time.Second), nil, nil, AblationMask{})
	undeclaredWant := RoleActivation{GeneratorID: "generator.low", Kind: economy.RoleSynergyFeed, TargetID: "pool.operations"}
	if err != nil || undeclared.Decision.Outcome != save.IntentApplied || hasRoleActivation(undeclared.RoleActivations, undeclaredWant) {
		t.Fatalf("undeclared synergy role=%+v err=%v", undeclared, err)
	}

	activeCatalog := foundationCatalog(t)
	activeState := foundationState(t, activeCatalog, started)
	activeState.GeneratorCounts["generator.low"] = 1
	activeState.ManualTokenMilli = 50_000
	active, err := SimulateTransition(request, activeState, activeCatalog, SimulationDependencies{Routes: foundationRoutes(t)}, save.Revision{Number: 1}, ModeOnline, started.Add(time.Second), nil, nil, AblationMask{})
	if err != nil || active.Decision.Outcome != save.IntentApplied || !hasRoleActivation(active.RoleActivations, undeclaredWant) {
		t.Fatalf("exercised synergy role=%+v err=%v", active, err)
	}
}

func TestSimulationSynergyActivationOccursAtRateApplication(t *testing.T) {
	started := time.Date(2026, 8, 3, 10, 0, 0, 0, time.UTC)
	targeted := foundationHighSynergyCatalog(t, "generator.low")
	dependencies := SimulationDependencies{Routes: foundationRoutes(t)}
	want := RoleActivation{GeneratorID: "generator.high", Kind: economy.RoleSynergyFeed, TargetID: "pool.operations"}

	masked := foundationState(t, targeted, started)
	masked.GeneratorCounts["generator.high"] = 1
	masked.GeneratorCounts["generator.low"] = 1
	masked.ManualTokenMilli = 50_000
	manual := IntentRequest{IntentID: "01986666-0221-7000-8000-000000000221", Kind: IntentPerformManualBatch, ExpectedRevision: 1, ActionID: "manual.click", Count: 1, WindowMS: 1}
	result, err := SimulateTransition(manual, masked, targeted, dependencies, save.Revision{Number: 1}, ModeOnline, started.Add(time.Second), nil, nil, AblationMask{GeneratorIDs: []string{"generator.low"}})
	if err != nil || result.Decision.Outcome != save.IntentApplied || hasRoleActivation(result.RoleActivations, want) {
		t.Fatalf("masked target result=%+v err=%v", result, err)
	}

	boughtAfter := foundationState(t, targeted, started)
	boughtAfter.GeneratorCounts["generator.high"] = 1
	buy := IntentRequest{IntentID: "01986666-0222-7000-8000-000000000222", Kind: IntentBuyGenerator, ExpectedRevision: 1, GeneratorID: "generator.low", CountMode: "exact", Count: 1}
	result, err = SimulateTransition(buy, boughtAfter, targeted, dependencies, save.Revision{Number: 1}, ModeOnline, started.Add(time.Second), nil, nil, AblationMask{})
	if err != nil || result.Decision.Outcome != save.IntentApplied || boughtAfter.GeneratorCounts["generator.low"] != 1 || hasRoleActivation(result.RoleActivations, want) {
		t.Fatalf("bought-after result=%+v counts=%v err=%v", result, boughtAfter.GeneratorCounts, err)
	}

	terminalProvision := foundationState(t, targeted, started)
	terminalProvision.GeneratorCounts["generator.high"] = 2
	terminalProvision.ManualTokenMilli = 50_000
	manual.IntentID = "01986666-0223-7000-8000-000000000223"
	result, err = SimulateTransition(manual, terminalProvision, targeted, dependencies, save.Revision{Number: 1}, ModeOnline, started.Add(time.Minute), nil, nil, AblationMask{})
	if err != nil || result.Decision.Outcome != save.IntentApplied || terminalProvision.GeneratorProvisioned["generator.low"] != 1 || hasRoleActivation(result.RoleActivations, want) {
		t.Fatalf("terminal provision result=%+v provisioned=%v err=%v", result, terminalProvision.GeneratorProvisioned, err)
	}

	allCatalog := foundationHighSynergyCatalog(t, "all")
	allState := foundationState(t, allCatalog, started)
	allState.GeneratorCounts["generator.high"] = 1
	allState.ManualTokenMilli = 50_000
	manual.IntentID = "01986666-0224-7000-8000-000000000224"
	result, err = SimulateTransition(manual, allState, allCatalog, dependencies, save.Revision{Number: 1}, ModeOnline, started.Add(time.Second), nil, nil, AblationMask{})
	if err != nil || result.Decision.Outcome != save.IntentApplied || !hasRoleActivation(result.RoleActivations, want) {
		t.Fatalf("all target result=%+v err=%v", result, err)
	}
}

func TestSimulationStockRateActivationRequiresNonNeutralRealHook(t *testing.T) {
	catalog := foundationCatalog(t)
	factionCatalog := foundationFactionCatalog(t)
	hook := faction.AccrualHook{
		Catalogs: faction.CatalogSet{foundationConstantsHash: factionCatalog},
		Policies: foundationCatchupPolicies{foundationConstantsHash: 120_000},
	}
	dependencies := SimulationDependencies{Routes: foundationRoutes(t), Hook: hook}
	started := time.Date(2026, 8, 3, 10, 0, 0, 0, time.UTC)
	request := IntentRequest{IntentID: "01986666-0212-7000-8000-000000000212", Kind: IntentPerformManualBatch, ExpectedRevision: 1, ActionID: "manual.click", Count: 1, WindowMS: 1}

	state := foundationState(t, catalog, started)
	state.FactionID = "bootstrapper"
	state.IncorporatedAt = started
	state.GeneratorCounts["generator.low"] = 10
	state.ManualTokenMilli = 50_000
	result, err := SimulateTransition(request, state, catalog, dependencies, save.Revision{Number: 1, ConstantsHash: foundationConstantsHash}, ModeOnline, started.Add(time.Minute), nil, nil, AblationMask{})
	if err != nil || result.Decision.Outcome != save.IntentApplied {
		t.Fatalf("result=%+v err=%v", result, err)
	}
	want := RoleActivation{GeneratorID: "generator.low", Kind: economy.RoleStockRate, TargetID: "faction.stock"}
	if !hasRoleActivation(result.RoleActivations, want) || state.StockUnits != 1 || state.StockProgressMS != 1_200 {
		t.Fatalf("activations=%+v stock=%d progress=%d", result.RoleActivations, state.StockUnits, state.StockProgressMS)
	}

	masked := foundationState(t, catalog, started)
	masked.FactionID = "bootstrapper"
	masked.IncorporatedAt = started
	masked.GeneratorCounts["generator.low"] = 10
	masked.ManualTokenMilli = 50_000
	result, err = SimulateTransition(request, masked, catalog, dependencies, save.Revision{Number: 1, ConstantsHash: foundationConstantsHash}, ModeOnline, started.Add(time.Minute), nil, nil, AblationMask{GeneratorIDs: []string{"generator.low"}})
	if err != nil || result.Decision.Outcome != save.IntentApplied {
		t.Fatalf("masked result=%+v err=%v", result, err)
	}
	for _, activation := range result.RoleActivations {
		if activation == want {
			t.Fatalf("masked stock role activated: %+v", result.RoleActivations)
		}
	}
	if masked.StockUnits != 1 || masked.StockProgressMS != 0 {
		t.Fatalf("masked stock=%d progress=%d", masked.StockUnits, masked.StockProgressMS)
	}
}

func TestSimulationStockRateActivationUsesPerRoleCounterfactual(t *testing.T) {
	catalog := foundationCatalog(t)
	factionCatalog := foundationFactionCatalog(t)
	hook := faction.AccrualHook{
		Catalogs: faction.CatalogSet{foundationConstantsHash: factionCatalog},
		Policies: foundationCatchupPolicies{foundationConstantsHash: 120_000},
	}
	started := time.Date(2026, 8, 3, 10, 0, 0, 0, time.UTC)
	state := foundationState(t, catalog, started)
	state.FactionID = "bootstrapper"
	state.IncorporatedAt = started
	state.StockUnits = factionCatalog.StockCap
	state.StockProgressMS = 30_000
	state.GeneratorCounts["generator.low"] = 1_000
	state.ManualTokenMilli = 50_000
	request := IntentRequest{IntentID: "01986666-0220-7000-8000-000000000220", Kind: IntentPerformManualBatch, ExpectedRevision: 1, ActionID: "manual.click", Count: 1, WindowMS: 1}
	result, err := SimulateTransition(request, state, catalog, SimulationDependencies{Routes: foundationRoutes(t), Hook: hook}, save.Revision{Number: 1, ConstantsHash: foundationConstantsHash}, ModeOnline, started.Add(30*time.Second), nil, nil, AblationMask{})
	if err != nil || result.Decision.Outcome != save.IntentApplied || state.StockProgressMS != 0 {
		t.Fatalf("result=%+v stock progress=%d err=%v", result, state.StockProgressMS, err)
	}
	want := RoleActivation{GeneratorID: "generator.low", Kind: economy.RoleStockRate, TargetID: "faction.stock"}
	if hasRoleActivation(result.RoleActivations, want) {
		t.Fatalf("counterfactually neutral stock role activated: %+v", result.RoleActivations)
	}
}

func TestSimulationMaskNullsProvisionEdgeAndRemovedActionRejectsBeforeAccrual(t *testing.T) {
	catalog := foundationCatalog(t)
	started := time.Date(2026, 8, 3, 10, 0, 0, 0, time.UTC)
	state := foundationState(t, catalog, started)
	state.GeneratorCounts["generator.high"] = 2
	state.ManualTokenMilli = 50_000
	request := IntentRequest{IntentID: "01986666-0203-7000-8000-000000000203", Kind: IntentPerformManualBatch, ExpectedRevision: 1, ActionID: "manual.click", Count: 1, WindowMS: 1}
	dependencies := SimulationDependencies{Routes: foundationRoutes(t)}
	result, err := SimulateTransition(request, state, catalog, dependencies, save.Revision{Number: 1}, ModeOnline, started.Add(180*time.Second), nil, nil, AblationMask{GeneratorIDs: []string{"generator.high"}})
	if err != nil || result.Decision.Outcome != save.IntentApplied || state.GeneratorProvisioned["generator.low"] != 0 {
		t.Fatalf("decision=%+v provisioned=%v err=%v", result.Decision, state.GeneratorProvisioned, err)
	}
	balance, _ := state.Ledger.Balance("company.cash")
	if balance.String() != "1e0" {
		t.Fatalf("masked provision output=%s", balance)
	}

	rejectedState := foundationState(t, catalog, started)
	rejectedState.GeneratorCounts["generator.low"] = 10
	rejectedState.ManualTokenMilli = 50_000
	rejected, err := SimulateTransition(request, rejectedState, catalog, dependencies, save.Revision{Number: 1}, ModeOnline, started.Add(time.Minute), nil, nil, AblationMask{RemovedActionIDs: []string{"manual.click"}})
	if err != nil || rejected.Decision.Outcome != save.IntentRejected || !rejectedState.EvaluatedThrough.Equal(started) {
		t.Fatalf("rejected=%+v evaluated=%s err=%v", rejected.Decision, rejectedState.EvaluatedThrough, err)
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
	result, err := SimulateTransition(request, state, catalog, SimulationDependencies{Routes: foundationRoutes(t)}, save.Revision{Number: 1}, ModeOnline, started.Add(time.Second), nil, nil, AblationMask{GeneratorIDs: []string{"generator.low"}})
	if err != nil || result.Decision.Outcome != save.IntentApplied || state.OfferState != nil {
		t.Fatalf("decision=%+v offer=%+v err=%v", result.Decision, state.OfferState, err)
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

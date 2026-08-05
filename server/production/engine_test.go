package production

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"cloud-clicker/server/decimal"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/multiplier"
	"cloud-clicker/server/routes"
	"cloud-clicker/server/save"
)

var engineCursor = time.Date(2026, 7, 28, 8, 0, 0, 0, time.UTC)

func mustDecimal(t *testing.T, source string) decimal.Decimal {
	t.Helper()
	value, err := decimal.ParseCanonical(source)
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func phase0Catalog(t *testing.T) *economy.Catalog {
	t.Helper()
	data, err := os.ReadFile("../../balance/catalogs/phase0.json")
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := economy.LoadCatalog(data)
	if err != nil {
		t.Fatal(err)
	}
	return catalog
}

func phase0Routes(t *testing.T) *routes.Catalog {
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

func engineState(t *testing.T, catalog *economy.Catalog, cash string, count int64) *save.State {
	t.Helper()
	ledger, err := economy.RestoreLedger(catalog, economy.ScopeCompany, map[string]string{"company.cash": cash})
	if err != nil {
		t.Fatal(err)
	}
	return &save.State{
		Ledger: ledger, GeneratorCounts: map[string]int64{"generator.beige_tower": count},
		EvaluatedThrough: engineCursor, ManualTokenMilli: 50_000, ManualTokenRefilledAt: engineCursor,
	}
}

func TestEvaluateOnlineAndClockRollback(t *testing.T) {
	catalog := phase0Catalog(t)
	state := engineState(t, catalog, "0", 1)
	result, err := Evaluate(state, catalog, engineCursor.Add(time.Second), ModeOnline, nil)
	if err != nil {
		t.Fatal(err)
	}
	if got := state.Ledger.Snapshot()["company.cash"]; got != "1e0" {
		t.Fatalf("cash = %s", got)
	}
	if result.ElapsedMS != 1000 || result.ProductionMS != 1000 || result.BankedCreditMS != 0 {
		t.Fatalf("result = %+v", result)
	}
	before := state.Ledger.Snapshot()["company.cash"]
	if _, err := Evaluate(state, catalog, engineCursor, ModeOnline, nil); err != nil {
		t.Fatal(err)
	}
	if got := state.Ledger.Snapshot()["company.cash"]; got != before {
		t.Fatalf("clock rollback changed cash: %s -> %s", before, got)
	}
}

func TestEvaluateMillisecondBoundaries(t *testing.T) {
	catalog := phase0Catalog(t)
	state := engineState(t, catalog, "0", 1)

	result, err := Evaluate(state, catalog, engineCursor.Add(999*time.Microsecond), ModeOnline, nil)
	if err != nil {
		t.Fatal(err)
	}
	if result.ElapsedMS != 0 || state.Ledger.Snapshot()["company.cash"] != "0" || !state.EvaluatedThrough.Equal(engineCursor) {
		t.Fatalf("sub-millisecond result=%+v state=%+v", result, state)
	}
	result, err = Evaluate(state, catalog, engineCursor.Add(time.Millisecond), ModeOnline, nil)
	if err != nil {
		t.Fatal(err)
	}
	if result.ElapsedMS != 1 || state.Ledger.Snapshot()["company.cash"] != "1e-3" ||
		!state.EvaluatedThrough.Equal(engineCursor.Add(time.Millisecond)) {
		t.Fatalf("exact-millisecond result=%+v cash=%s cursor=%s", result, state.Ledger.Snapshot()["company.cash"], state.EvaluatedThrough)
	}
	before := state.Ledger.Snapshot()["company.cash"]
	result, err = Evaluate(state, catalog, engineCursor.Add(500*time.Microsecond), ModeOnline, nil)
	if err != nil || result.ElapsedMS != 0 || state.Ledger.Snapshot()["company.cash"] != before {
		t.Fatalf("rollback result=%+v cash=%s err=%v", result, state.Ledger.Snapshot()["company.cash"], err)
	}

	offline := engineState(t, catalog, "0", 1)
	result, err = Evaluate(offline, catalog, engineCursor.Add(86_400_000*time.Millisecond), ModeOffline, nil)
	if err != nil {
		t.Fatal(err)
	}
	if result.ProductionMS != 86_400_000 || result.BankedCreditMS != 0 ||
		offline.Ledger.Snapshot()["company.cash"] != "7.776e4" {
		t.Fatalf("offline boundary result=%+v cash=%s", result, offline.Ledger.Snapshot()["company.cash"])
	}
}

func TestEvaluateOfflineCapsProductionAndBanksExcess(t *testing.T) {
	catalog := phase0Catalog(t)
	state := engineState(t, catalog, "0", 1)
	result, err := Evaluate(state, catalog, engineCursor.Add(48*time.Hour), ModeOffline, nil)
	if err != nil {
		t.Fatal(err)
	}
	if got := state.Ledger.Snapshot()["company.cash"]; got != "7.776e4" {
		t.Fatalf("offline cash = %s", got)
	}
	if result.ProductionMS != 86_400_000 || result.BankedCreditMS != 43_200_000 || state.ComputeCreditMS != 43_200_000 {
		t.Fatalf("offline result=%+v credits=%d", result, state.ComputeCreditMS)
	}
}

func TestComputeBurstIsPartitionInvariantAndOfflineCapBounded(t *testing.T) {
	catalog := phase0Catalog(t)
	oneShot := engineState(t, catalog, "0", 1)
	oneShot.WireVersion, oneShot.ComputeBurstRemainingMS = 17, 1_500
	if _, err := Evaluate(oneShot, catalog, engineCursor.Add(1500*time.Millisecond), ModeOnline, nil); err != nil {
		t.Fatal(err)
	}
	partitioned := engineState(t, catalog, "0", 1)
	partitioned.WireVersion, partitioned.ComputeBurstRemainingMS = 17, 1_500
	if _, err := Evaluate(partitioned, catalog, engineCursor.Add(500*time.Millisecond), ModeOnline, nil); err != nil {
		t.Fatal(err)
	}
	if partitioned.ComputeBurstRemainingMS != 1_000 {
		t.Fatalf("partial burst remaining=%d", partitioned.ComputeBurstRemainingMS)
	}
	if _, err := Evaluate(partitioned, catalog, engineCursor.Add(1500*time.Millisecond), ModeOnline, nil); err != nil {
		t.Fatal(err)
	}
	if got, want := partitioned.Ledger.Snapshot()["company.cash"], oneShot.Ledger.Snapshot()["company.cash"]; got != want || got != "3e0" || partitioned.ComputeBurstRemainingMS != 0 {
		t.Fatalf("partitioned cash=%s one-shot=%s remaining=%d", got, want, partitioned.ComputeBurstRemainingMS)
	}

	provisionCatalog := foundationCatalog(t)
	provisionOneShot := foundationState(t, provisionCatalog, engineCursor)
	provisionOneShot.WireVersion, provisionOneShot.ComputeBurstRemainingMS = 17, 90_000
	provisionOneShot.GeneratorCounts["generator.high"], provisionOneShot.GeneratorPurchasedTotal = 2, 2
	if _, err := Evaluate(provisionOneShot, provisionCatalog, engineCursor.Add(90*time.Second), ModeOnline, nil); err != nil {
		t.Fatal(err)
	}
	provisionSplit := foundationState(t, provisionCatalog, engineCursor)
	provisionSplit.WireVersion, provisionSplit.ComputeBurstRemainingMS = 17, 90_000
	provisionSplit.GeneratorCounts["generator.high"], provisionSplit.GeneratorPurchasedTotal = 2, 2
	if _, err := Evaluate(provisionSplit, provisionCatalog, engineCursor.Add(30*time.Second), ModeOnline, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := Evaluate(provisionSplit, provisionCatalog, engineCursor.Add(90*time.Second), ModeOnline, nil); err != nil {
		t.Fatal(err)
	}
	if got, want := provisionSplit.Ledger.Snapshot()["company.cash"], provisionOneShot.Ledger.Snapshot()["company.cash"]; got != want || got != "3.66e3" || provisionSplit.GeneratorProvisioned["generator.low"] != 1 || provisionSplit.ComputeBurstRemainingMS != 0 {
		t.Fatalf("provision-boundary cash=%s one-shot=%s provisioned=%v remaining=%d", got, want, provisionSplit.GeneratorProvisioned, provisionSplit.ComputeBurstRemainingMS)
	}

	data, err := os.ReadFile("../../balance/catalogs/phase0.json")
	if err != nil {
		t.Fatal(err)
	}
	var root map[string]any
	if err := json.Unmarshal(data, &root); err != nil {
		t.Fatal(err)
	}
	root["offline_policy"].(map[string]any)["burst_max_duration_ms"] = float64(90_000_000)
	data, err = json.Marshal(root)
	if err != nil {
		t.Fatal(err)
	}
	offlineCatalog, err := economy.LoadCatalog(data)
	if err != nil {
		t.Fatal(err)
	}
	offline := engineState(t, offlineCatalog, "0", 1)
	offline.WireVersion, offline.ComputeBurstRemainingMS = 17, 90_000_000
	result, err := Evaluate(offline, offlineCatalog, engineCursor.Add(25*time.Hour), ModeOffline, nil)
	if err != nil {
		t.Fatal(err)
	}
	if result.ProductionMS != 86_400_000 || offline.ComputeBurstRemainingMS != 0 || offline.Ledger.Snapshot()["company.cash"] != "1.5552e5" {
		t.Fatalf("offline burst result=%+v cash=%s remaining=%d", result, offline.Ledger.Snapshot()["company.cash"], offline.ComputeBurstRemainingMS)
	}
}

func TestEvaluateOfflineCreditAndResourceHardcaps(t *testing.T) {
	data, err := os.ReadFile("../../balance/catalogs/phase0.json")
	if err != nil {
		t.Fatal(err)
	}
	var root map[string]any
	if err := json.Unmarshal(data, &root); err != nil {
		t.Fatal(err)
	}
	resource := root["resources"].([]any)[0].(map[string]any)
	resource["hardcap"].(map[string]any)["amount"] = "1e1"
	generator := root["generator_classes"].([]any)[0].(map[string]any)
	generator["production"].(map[string]any)["base_rate"] = "1e2"
	data, _ = json.Marshal(root)
	catalog, err := economy.LoadCatalog(data)
	if err != nil {
		t.Fatal(err)
	}
	state := engineState(t, catalog, "0", 1)
	state.ComputeCreditMS = 250_000_000
	result, err := Evaluate(state, catalog, engineCursor.Add(48*time.Hour), ModeOffline, nil)
	if err != nil {
		t.Fatal(err)
	}
	if got := state.Ledger.Snapshot()["company.cash"]; got != "1e1" {
		t.Fatalf("hardcapped cash = %s", got)
	}
	if result.BankedCreditMS != 9_200_000 || state.ComputeCreditMS != 259_200_000 {
		t.Fatalf("credit cap result=%+v credits=%d", result, state.ComputeCreditMS)
	}
}

func TestEvaluateSaturatesR1VectorAndFollowingPurchaseSucceeds(t *testing.T) {
	data, err := os.ReadFile("../../balance/catalogs/phase0.json")
	if err != nil {
		t.Fatal(err)
	}
	var root map[string]any
	if err := json.Unmarshal(data, &root); err != nil {
		t.Fatal(err)
	}
	resource := root["resources"].([]any)[0].(map[string]any)
	resource["hardcap"].(map[string]any)["amount"] = "9.87256122677e8"
	data, err = json.Marshal(root)
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := economy.LoadCatalog(data)
	if err != nil {
		t.Fatal(err)
	}
	state := engineState(t, catalog, "5.6765610215e6", 1)
	elapsed := 981_579_561_656 * time.Millisecond
	result, err := Evaluate(state, catalog, engineCursor.Add(elapsed), ModeOnline, nil)
	if err != nil {
		t.Fatal(err)
	}
	if got := state.Ledger.Snapshot()["company.cash"]; got != "9.87256122677e8" {
		t.Fatalf("cash = %s, want exact hardcap", got)
	}
	if len(result.Receipt.Changes) != 1 || result.Receipt.Changes[0].Delta != "9.81579561655e8" {
		t.Fatalf("receipt = %+v", result.Receipt)
	}
	change := result.Receipt.Changes[0]
	before := mustDecimal(t, change.Before)
	delta := mustDecimal(t, change.Delta)
	if got := before.Add(delta).Quantize(decimal.CanonicalSignificantDigits).String(); got != change.After {
		t.Fatalf("receipt delta reapplies to %s, want %s", got, change.After)
	}

	service := &Service{}
	decision, err := service.buyGenerator(IntentRequest{
		IntentID: "018f6b7c-9abc-7def-8abc-0123456789ab", Kind: IntentBuyGenerator,
		GeneratorID: "generator.beige_tower", CountMode: "exact", Count: 1,
	}, state, catalog, save.Revision{Number: 1}, ModeOnline, state.EvaluatedThrough, nil, &invariantCollector{}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if decision.Outcome != save.IntentApplied || state.GeneratorCounts["generator.beige_tower"] != 2 {
		t.Fatalf("following purchase = %+v generators=%d", decision, state.GeneratorCounts["generator.beige_tower"])
	}
}

func TestEvaluateAtHardcapAdvancesCursorWithoutLedgerChange(t *testing.T) {
	data, err := os.ReadFile("../../balance/catalogs/phase0.json")
	if err != nil {
		t.Fatal(err)
	}
	var root map[string]any
	if err := json.Unmarshal(data, &root); err != nil {
		t.Fatal(err)
	}
	root["resources"].([]any)[0].(map[string]any)["hardcap"].(map[string]any)["amount"] = "1e2"
	data, err = json.Marshal(root)
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := economy.LoadCatalog(data)
	if err != nil {
		t.Fatal(err)
	}
	state := engineState(t, catalog, "1e2", 1)
	now := engineCursor.Add(time.Second)
	result, err := Evaluate(state, catalog, now, ModeOnline, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Receipt.Changes) != 0 || state.Ledger.Snapshot()["company.cash"] != "1e2" || !state.EvaluatedThrough.Equal(now) {
		t.Fatalf("result=%+v balance=%s cursor=%s", result, state.Ledger.Snapshot()["company.cash"], state.EvaluatedThrough)
	}
}

func TestEvaluationPolicyGoldenVectors(t *testing.T) {
	data, err := os.ReadFile("../../testdata/production-engine.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Version int `json:"version"`
		Cases   []struct {
			Name               string         `json:"name"`
			Mode               EvaluationMode `json:"mode"`
			ElapsedMS          int64          `json:"elapsed_ms"`
			InitialCash        string         `json:"initial_cash"`
			GeneratorCount     int64          `json:"generator_count"`
			InitialCreditMS    int64          `json:"initial_credit_ms"`
			ExpectCash         string         `json:"expect_cash"`
			ExpectProductionMS int64          `json:"expect_production_ms"`
			ExpectBankedMS     int64          `json:"expect_banked_ms"`
			ExpectCreditMS     int64          `json:"expect_credit_ms"`
		} `json:"evaluation_cases"`
	}
	if err := json.Unmarshal(data, &fixture); err != nil || fixture.Version != 1 || len(fixture.Cases) == 0 {
		t.Fatalf("fixture: version=%d cases=%d err=%v", fixture.Version, len(fixture.Cases), err)
	}
	catalog := phase0Catalog(t)
	for _, vector := range fixture.Cases {
		t.Run(vector.Name, func(t *testing.T) {
			state := engineState(t, catalog, vector.InitialCash, vector.GeneratorCount)
			state.ComputeCreditMS = vector.InitialCreditMS
			result, err := Evaluate(state, catalog, engineCursor.Add(time.Duration(vector.ElapsedMS)*time.Millisecond), vector.Mode, nil)
			if err != nil {
				t.Fatal(err)
			}
			if got := state.Ledger.Snapshot()["company.cash"]; got != vector.ExpectCash ||
				result.ProductionMS != vector.ExpectProductionMS || result.BankedCreditMS != vector.ExpectBankedMS ||
				state.ComputeCreditMS != vector.ExpectCreditMS {
				t.Fatalf("cash=%s result=%+v credits=%d", got, result, state.ComputeCreditMS)
			}
		})
	}
}

func TestRatesValidateDeclaredContributions(t *testing.T) {
	data, err := os.ReadFile("../../balance/catalogs/phase0.json")
	if err != nil {
		t.Fatal(err)
	}
	var root map[string]any
	if err := json.Unmarshal(data, &root); err != nil {
		t.Fatal(err)
	}
	root["multiplier_sources"] = []any{
		map[string]any{"id": "upgrade.double", "slot": "upgrades", "target": "all", "provider": "upgrade.double"},
		map[string]any{"id": "milestone.triple", "slot": "milestones", "target": "generator.beige_tower", "provider": "milestone.triple"},
	}
	data, _ = json.Marshal(root)
	catalog, err := economy.LoadCatalog(data)
	if err != nil {
		t.Fatal(err)
	}
	rates, err := Rates(catalog, map[string]int64{"generator.beige_tower": 2}, []multiplier.Contribution{
		{Slot: multiplier.SlotMilestones, SourceID: "milestone.triple", Target: "generator.beige_tower", Factor: mustDecimal(t, "3e0")},
		{Slot: multiplier.SlotUpgrades, SourceID: "upgrade.double", Target: "all", Factor: mustDecimal(t, "2e0")},
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := rates["company.cash"][0].String(); got != "1.2e1" {
		t.Fatalf("rate = %s", got)
	}
	if _, err := Rates(catalog, map[string]int64{"generator.beige_tower": 2}, []multiplier.Contribution{{
		Slot: multiplier.SlotTrust, SourceID: "upgrade.double", Target: "all", Factor: mustDecimal(t, "2e0"),
	}}); err == nil {
		t.Fatal("mismatched contribution was accepted")
	}

	valid := multiplier.Contribution{
		Slot: multiplier.SlotUpgrades, SourceID: "upgrade.double", Target: "all", Factor: mustDecimal(t, "2e0"),
	}
	tests := map[string][]multiplier.Contribution{
		"undeclared":        {{Slot: multiplier.SlotUpgrades, SourceID: "upgrade.missing", Target: "all", Factor: mustDecimal(t, "2e0")}},
		"mismatched target": {{Slot: multiplier.SlotUpgrades, SourceID: "upgrade.double", Target: "generator.beige_tower", Factor: mustDecimal(t, "2e0")}},
		"duplicate source":  {valid, valid},
		"zero factor":       {{Slot: multiplier.SlotUpgrades, SourceID: "upgrade.double", Target: "all", Factor: decimal.Zero}},
		"negative factor":   {{Slot: multiplier.SlotUpgrades, SourceID: "upgrade.double", Target: "all", Factor: mustDecimal(t, "-1e0")}},
	}
	for name, contributions := range tests {
		t.Run(name, func(t *testing.T) {
			if _, err := Rates(catalog, map[string]int64{"generator.beige_tower": 2}, contributions); err == nil {
				t.Fatal("invalid runtime contribution set was accepted")
			}
		})
	}
}

func TestRatesArePermutationInvariantWithinSlot(t *testing.T) {
	data, err := os.ReadFile("../../balance/catalogs/phase0.json")
	if err != nil {
		t.Fatal(err)
	}
	var root map[string]any
	if err := json.Unmarshal(data, &root); err != nil {
		t.Fatal(err)
	}
	root["multiplier_sources"] = []any{
		map[string]any{"id": "a.a", "slot": "upgrades", "target": "all", "provider": "upgrade.a"},
		map[string]any{"id": "a0", "slot": "upgrades", "target": "all", "provider": "upgrade.b"},
		map[string]any{"id": "a_a", "slot": "upgrades", "target": "all", "provider": "upgrade.c"},
	}
	data, err = json.Marshal(root)
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := economy.LoadCatalog(data)
	if err != nil {
		t.Fatal(err)
	}
	contributions := []multiplier.Contribution{
		{Slot: multiplier.SlotUpgrades, SourceID: "a.a", Target: "all", Factor: mustDecimal(t, "1.00000000001e0")},
		{Slot: multiplier.SlotUpgrades, SourceID: "a0", Target: "all", Factor: mustDecimal(t, "9.99999999999e0")},
		{Slot: multiplier.SlotUpgrades, SourceID: "a_a", Target: "all", Factor: mustDecimal(t, "1.23456789012e0")},
	}
	permutations := [][]int{{0, 1, 2}, {0, 2, 1}, {1, 0, 2}, {1, 2, 0}, {2, 0, 1}, {2, 1, 0}}
	want := ""
	for _, order := range permutations {
		candidate := []multiplier.Contribution{contributions[order[0]], contributions[order[1]], contributions[order[2]]}
		rates, err := Rates(catalog, map[string]int64{"generator.beige_tower": 1}, candidate)
		if err != nil {
			t.Fatal(err)
		}
		got := rates["company.cash"][0].String()
		if want == "" {
			want = got
		} else if got != want {
			t.Fatalf("permutation %v rate = %s, want %s", order, got, want)
		}
	}
}

func TestSubProgressSharedFixture(t *testing.T) {
	data, err := os.ReadFile("../../testdata/production-engine.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Version int `json:"version"`
		Cases   []struct {
			Name           string `json:"name"`
			Tier           int    `json:"tier"`
			Cash           string `json:"cash"`
			GeneratorCount int64  `json:"generator_count"`
			Expect         string `json:"expect"`
		} `json:"progress_cases"`
		TargetCases []struct {
			Name        string `json:"name"`
			Target      string `json:"target"`
			ExpectValid bool   `json:"expect_valid"`
		} `json:"resource_log_target_cases"`
		ResourceLogCases []struct {
			Name   string `json:"name"`
			Target string `json:"target"`
			Value  string `json:"value"`
			Expect string `json:"expect"`
		} `json:"resource_log_progress_cases"`
	}
	if err := json.Unmarshal(data, &fixture); err != nil || fixture.Version != 1 ||
		len(fixture.TargetCases) == 0 || len(fixture.ResourceLogCases) == 0 {
		t.Fatalf("fixture: version=%d err=%v", fixture.Version, err)
	}
	catalog := phase0Catalog(t)
	for _, vector := range fixture.Cases {
		t.Run(vector.Name, func(t *testing.T) {
			state := engineState(t, catalog, vector.Cash, vector.GeneratorCount)
			value, err := SubProgressValue(catalog, state, vector.Tier)
			if err != nil {
				t.Fatal(err)
			}
			if got := value.String(); got != vector.Expect {
				t.Fatalf("progress = %s, want %s", got, vector.Expect)
			}
		})
	}
	for _, vector := range fixture.TargetCases {
		for _, composite := range []bool{false, true} {
			position := "top-level"
			if composite {
				position = "composite"
			}
			t.Run(vector.Name+"/"+position, func(t *testing.T) {
				_, err := phase0CatalogWithResourceLogTarget(t, vector.Target, composite)
				if gotValid := err == nil; gotValid != vector.ExpectValid {
					t.Fatalf("target %s valid=%v, want %v, error=%v", vector.Target, gotValid, vector.ExpectValid, err)
				}
			})
		}
	}
	for _, vector := range fixture.ResourceLogCases {
		t.Run(vector.Name, func(t *testing.T) {
			catalog, err := phase0CatalogWithResourceLogTarget(t, vector.Target, false)
			if err != nil {
				t.Fatal(err)
			}
			state := engineState(t, catalog, vector.Value, 0)
			value, err := SubProgressValue(catalog, state, 0)
			if err != nil {
				t.Fatal(err)
			}
			if got := value.String(); got != vector.Expect {
				t.Fatalf("progress = %s, want %s", got, vector.Expect)
			}
		})
	}
}

func phase0CatalogWithResourceLogTarget(t *testing.T, target string, composite bool) (*economy.Catalog, error) {
	t.Helper()
	data, err := os.ReadFile("../../balance/catalogs/phase0.json")
	if err != nil {
		t.Fatal(err)
	}
	var root map[string]any
	if err := json.Unmarshal(data, &root); err != nil {
		t.Fatal(err)
	}
	coordinates := root["progress_coordinates"].([]any)
	if composite {
		terms := coordinates[1].(map[string]any)["terms"].([]any)
		for _, rawTerm := range terms {
			term := rawTerm.(map[string]any)
			if term["kind"] == "resource_log" {
				term["target"] = target
			}
		}
	} else {
		coordinates[0].(map[string]any)["target"] = target
	}
	data, err = json.Marshal(root)
	if err != nil {
		t.Fatal(err)
	}
	return economy.LoadCatalog(data)
}

func TestSubProgressIsMonotoneUnderPureAccrual(t *testing.T) {
	catalog := phase0Catalog(t)
	values := []string{"0", "1e0", "1e3", "1e6", "1e9", "1e12", "1e100"}
	for tier := 0; tier <= 3; tier++ {
		previous := mustDecimal(t, "0")
		for _, cash := range values {
			state := engineState(t, catalog, cash, 10)
			current, err := SubProgressValue(catalog, state, tier)
			if err != nil {
				t.Fatal(err)
			}
			if current.Lt(previous) {
				t.Fatalf("tier %d regressed at cash %s: %s < %s", tier, cash, current.String(), previous.String())
			}
			previous = current
		}
	}
}

func TestResourceLogRuntimeRejectsCollapsedDenominator(t *testing.T) {
	catalog := phase0Catalog(t)
	state := engineState(t, catalog, "1e0", 0)
	if _, err := resourceLog(state, "company.cash", mustDecimal(t, "4e-15")); err != ErrInvalidEngineState {
		t.Fatalf("error = %v, want ErrInvalidEngineState", err)
	}
}

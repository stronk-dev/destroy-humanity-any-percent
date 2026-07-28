package production

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"cloud-clicker/server/decimal"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/multiplier"
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
	}
	if err := json.Unmarshal(data, &fixture); err != nil || fixture.Version != 1 {
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
}

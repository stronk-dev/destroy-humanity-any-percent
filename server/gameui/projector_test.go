package gameui

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"cloud-clicker/server/decimal"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/production"
	"cloud-clicker/server/routes"
	"cloud-clicker/server/save"
)

func TestUpgradeProjectionExplainsFirstFailedPurchaseCheck(t *testing.T) {
	data, err := os.ReadFile("../../balance/testdata/t0-t1/economy-v4.json")
	if err != nil {
		t.Fatal(err)
	}
	var source map[string]any
	if err := json.Unmarshal(data, &source); err != nil {
		t.Fatal(err)
	}
	// Separate the cost from the prerequisite in this candidate so an
	// affordability failure cannot accidentally be a requirements failure.
	for _, value := range source["upgrades"].([]any) {
		row := value.(map[string]any)
		if row["id"] == "upgrade.reply_all_macro" {
			row["cost"].(map[string]any)["amount"] = "8e1"
		}
	}
	data, err = json.Marshal(source)
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := economy.LoadCatalog(data)
	if err != nil {
		t.Fatal(err)
	}
	_, routeCatalog := loadCandidateCatalogs(t)
	for _, test := range []struct {
		name, cash, id string
		owned, crossed bool
		reason         any
	}{
		{"owned-before-window-requirement-and-cost", "0", "upgrade.reply_all_macro", true, true, "owned"},
		{"closed-window-before-requirement-and-cost", "0", "upgrade.reply_all_macro", false, true, "window"},
		{"unopened-window", "1e9", "upgrade.rack_rail_standardization", false, false, "window"},
		{"requirement-before-cost", "0", "upgrade.reply_all_macro", false, false, "requirement"},
		{"unaffordable", "6e1", "upgrade.reply_all_macro", false, false, "unaffordable"},
		{"eligible", "1e2", "upgrade.reply_all_macro", false, false, nil},
	} {
		t.Run(test.name, func(t *testing.T) {
			state := candidateState(t, catalog)
			state.UpgradesOwned[test.id] = test.owned
			state.GatesCrossed["gate.t0_to_t1"] = test.crossed
			cash, _ := state.Ledger.Balance("company.cash")
			if _, err := state.Ledger.Apply(economy.Transaction{Entries: []economy.Entry{{ResourceID: "company.cash", Delta: decimal.FromString(test.cash).Sub(cash)}}}); err != nil {
				t.Fatal(err)
			}
			rows, err := upgradeRows(catalog, routeCatalog, state)
			if err != nil {
				t.Fatal(err)
			}
			encoded, err := json.Marshal(rows)
			if err != nil {
				t.Fatal(err)
			}
			var wire []map[string]any
			if err := json.Unmarshal(encoded, &wire); err != nil {
				t.Fatal(err)
			}
			for _, row := range wire {
				if row["upgrade_id"] != test.id {
					continue
				}
				reason, present := row["ineligible_reason"]
				if !present || reason != test.reason || row["eligible"] != (test.reason == nil) || row["owned"] != test.owned {
					t.Fatalf("upgrade state = %v, want explicit reason %v", row, test.reason)
				}
				return
			}
			t.Fatalf("missing upgrade %s", test.id)
		})
	}
}

func TestUpgradeProjectionExplainsAxisRequirementWithoutClientMath(t *testing.T) {
	data, err := os.ReadFile("../../balance/testdata/axis-stack/economy-v5-fixture.json")
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := economy.LoadCatalog(data)
	if err != nil {
		t.Fatal(err)
	}
	_, routeCatalog := loadCandidateCatalogs(t)
	state := axisProjectionState(t, catalog)
	state.GatesCrossed = map[string]bool{"gate.t0_to_t1": true}
	state.DoctrinesByTransition = map[string]string{}
	state.LedgerFactKinds, state.RegionTraits = map[string]bool{}, map[string]bool{}
	state.MeterBands = map[string]int{}
	if _, err := state.Ledger.Apply(economy.Transaction{Entries: []economy.Entry{{ResourceID: "company.cash", Delta: decimal.FromString("1e9")}}}); err != nil {
		t.Fatal(err)
	}
	for _, input := range []int64{8, 10} {
		state.AttainmentScoreRun = input
		rows, err := upgradeRows(catalog, routeCatalog, state)
		if err != nil {
			t.Fatal(err)
		}
		found := false
		for _, row := range rows {
			if row.UpgradeID != "upgrade.pr_intern_2" {
				continue
			}
			found = true
			if row.Eligible != (input == 10) || input == 8 && (row.IneligibleReason == nil || *row.IneligibleReason != "requirement") || input == 10 && row.IneligibleReason != nil {
				t.Fatalf("axis input %d: %+v", input, row)
			}
		}
		if !found {
			t.Fatal("missing PR intern row")
		}
	}
}

func loadCandidateCatalogs(t *testing.T) (*economy.Catalog, *routes.Catalog) {
	t.Helper()
	economyBytes, err := os.ReadFile("../../balance/testdata/t0-t1/economy-v4.json")
	if err != nil {
		t.Fatal(err)
	}
	economyCatalog, err := economy.LoadCatalog(economyBytes)
	if err != nil {
		t.Fatal(err)
	}
	routeBytes, err := os.ReadFile("../../balance/testdata/t0-t1/routes-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	routeCatalog, err := routes.LoadCatalog(routeBytes)
	if err != nil {
		t.Fatal(err)
	}
	return economyCatalog, routeCatalog
}

func candidateState(t *testing.T, catalog *economy.Catalog) *save.State {
	t.Helper()
	ledger, err := economy.NewLedger(catalog, economy.ScopeCompany)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ledger.Apply(economy.Transaction{Entries: []economy.Entry{{ResourceID: "company.cash", Delta: decimal.FromString("1e9")}}}); err != nil {
		t.Fatal(err)
	}
	counts, provisioned := map[string]int64{}, map[string]int64{}
	for _, definition := range catalog.GeneratorClassesForScope(economy.ScopeCompany) {
		counts[definition.ID], provisioned[definition.ID] = 0, 0
	}
	counts["generator.beige_tower"] = 2
	provisioned["generator.beige_tower"] = 3
	return &save.State{
		Ledger: ledger, GeneratorCounts: counts, GeneratorProvisioned: provisioned,
		ProvisionRemaindersPPM: map[string]int64{"generator.beige_tower": 0},
		UpgradesOwned:          map[string]bool{}, GatesCrossed: map[string]bool{},
		DoctrinesByTransition: map[string]string{}, LedgerFactKinds: map[string]bool{},
		MeterBands: map[string]int{}, RegionTraits: map[string]bool{},
		EvaluatedThrough: time.UnixMilli(1_800_000_000_000).UTC(),
		RunStartedAt:     time.UnixMilli(1_799_999_000_000).UTC(), RunSeq: 1,
	}
}

func TestProjectionRowsAreSortedPinnedCatalogViews(t *testing.T) {
	catalog, routeCatalog := loadCandidateCatalogs(t)
	state := candidateState(t, catalog)
	rates, err := production.ProjectRates(production.CatalogBundle{Economy: catalog}, state, nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	resources, err := resourceRows(catalog, state, rates.Resources)
	if err != nil || len(resources) != 2 || resources[0].ResourceID != "company.cash" || resources[1].ResourceID != "company.permits" ||
		resources[0].Cap == nil || resources[0].Cap.ReasonKey != "resource.company_cash.cap.phase0" {
		t.Fatalf("resources=%+v err=%v", resources, err)
	}
	generators, err := generatorRows(catalog, state, rates.Generators)
	if err != nil || len(generators) != 9 || generators[0].GeneratorID != "generator.answering_machine" ||
		generators[1].GeneratorID != "generator.beige_tower" || generators[1].Owned != 2 || generators[1].Provisioned != 3 ||
		generators[1].RateContribution != "5.01e0" || generators[1].NextCost != "1.2769e1" {
		t.Fatalf("generators=%+v err=%v", generators, err)
	}
	upgrades, err := upgradeRows(catalog, routeCatalog, state)
	if err != nil || len(upgrades) != 10 || upgrades[0].UpgradeID != "upgrade.beige_tower_cache" || !upgrades[0].Eligible {
		t.Fatalf("upgrades=%+v err=%v", upgrades, err)
	}
}

func TestProjectionRejectsIncompleteGeneratorKeySets(t *testing.T) {
	catalog, _ := loadCandidateCatalogs(t)
	state := candidateState(t, catalog)
	delete(state.GeneratorProvisioned, "generator.beige_tower")
	if _, err := production.ProjectRates(production.CatalogBundle{Economy: catalog}, state, nil, 0); err == nil {
		t.Fatal("incomplete provisioned-count set projected")
	}
}

func TestProjectionTreatsPreProvisioningStateAsZeroProvisioned(t *testing.T) {
	catalog, _ := loadCandidateCatalogs(t)
	state := candidateState(t, catalog)
	state.GeneratorProvisioned = nil
	rates, err := production.ProjectRates(production.CatalogBundle{Economy: catalog}, state, nil, 0)
	if err != nil || rates.Generators[1].GeneratorID != "generator.beige_tower" || rates.Generators[1].Rate.String() != "2.004e0" {
		t.Fatalf("legacy rates=%v err=%v", rates.Generators, err)
	}
}

func TestProjectionUsesKernelOwnedSchemaV4RateProjection(t *testing.T) {
	catalog, _ := loadCandidateCatalogs(t)
	state := candidateState(t, catalog)
	state.GeneratorCounts["generator.beige_tower"] = 25
	rates, err := production.ProjectRates(production.CatalogBundle{Economy: catalog}, state, nil, 0)
	if err != nil || rates.Generators[1].GeneratorID != "generator.beige_tower" || rates.Generators[1].Rate.String() != "5.74e1" {
		t.Fatalf("schema-v4 kernel rates=%v err=%v", rates.Generators, err)
	}
}

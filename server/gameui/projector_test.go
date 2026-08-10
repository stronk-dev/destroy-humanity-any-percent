package gameui

import (
	"os"
	"testing"
	"time"

	"cloud-clicker/server/decimal"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/production"
	"cloud-clicker/server/routes"
	"cloud-clicker/server/save"
)

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

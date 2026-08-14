package economy

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"cloud-clicker/server/decimal"
)

type kernelFixture struct {
	Version                int                     `json:"version"`
	Catalog                json.RawMessage         `json:"catalog"`
	CurveVectors           []curveVector           `json:"curve_vectors"`
	InvalidCases           []string                `json:"invalid_cases"`
	MultiplierCatalogCases []multiplierCatalogCase `json:"multiplier_catalog_cases"`
}

type multiplierCatalogCase struct {
	Name        string `json:"name"`
	ExpectValid bool   `json:"expect_valid"`
}

type curveVector struct {
	GeneratorID      string `json:"generator_id"`
	Owned            int64  `json:"owned"`
	Count            int64  `json:"count"`
	Cash             string `json:"cash"`
	ExpectCost       string `json:"expect_cost"`
	ExpectAffordable int64  `json:"expect_affordable"`
}

func loadKernelFixture(t *testing.T) kernelFixture {
	t.Helper()
	data, err := os.ReadFile("../../testdata/economy-kernel.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture kernelFixture
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.Version != 1 {
		t.Fatalf("fixture version = %d, want 1", fixture.Version)
	}
	return fixture
}

func TestSharedCatalogAndCurveVectors(t *testing.T) {
	fixture := loadKernelFixture(t)
	catalog, err := LoadCatalog(fixture.Catalog)
	if err != nil {
		t.Fatal(err)
	}
	if len(catalog.Resources()) != 3 || len(catalog.GeneratorClasses()) != 5 {
		t.Fatalf("unexpected catalog sizes: resources=%d generators=%d", len(catalog.Resources()), len(catalog.GeneratorClasses()))
	}

	for _, vector := range fixture.CurveVectors {
		t.Run(vector.GeneratorID, func(t *testing.T) {
			cost, err := catalog.BulkCost(vector.GeneratorID, vector.Owned, vector.Count)
			if err != nil {
				t.Fatal(err)
			}
			if got := cost.String(); got != vector.ExpectCost {
				t.Fatalf("cost = %s, want %s", got, vector.ExpectCost)
			}

			cash, err := decimal.ParseCanonical(vector.Cash)
			if err != nil {
				t.Fatal(err)
			}
			affordable, err := catalog.MaxAffordable(vector.GeneratorID, cash, vector.Owned)
			if err != nil {
				t.Fatal(err)
			}
			if affordable != vector.ExpectAffordable {
				t.Fatalf("affordable = %d, want %d", affordable, vector.ExpectAffordable)
			}
			atResult, err := catalog.BulkCost(vector.GeneratorID, vector.Owned, affordable)
			if err != nil || !atResult.Lte(cash) {
				t.Fatalf("reported count is not affordable: cost=%s err=%v", atResult.String(), err)
			}
			if affordable < decimal.MaxExactInteger-vector.Owned {
				next, nextErr := catalog.BulkCost(vector.GeneratorID, vector.Owned, affordable+1)
				if nextErr == nil && next.Lte(cash) {
					t.Fatalf("next count is also affordable: cost=%s", next.String())
				}
			}
		})
	}
}

func TestSharedInvalidCatalogCases(t *testing.T) {
	fixture := loadKernelFixture(t)
	for _, name := range fixture.InvalidCases {
		t.Run(name, func(t *testing.T) {
			invalid := mutateCatalog(t, fixture.Catalog, name)
			if _, err := LoadCatalog(invalid); err == nil {
				t.Fatal("invalid catalog was accepted")
			}
		})
	}
}

func TestSharedMultiplierCatalogCases(t *testing.T) {
	fixture := loadKernelFixture(t)
	phase0, err := os.ReadFile("../../balance/testdata/valid/permits-economy-candidate-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	for _, vector := range fixture.MultiplierCatalogCases {
		t.Run(vector.Name, func(t *testing.T) {
			_, err := LoadCatalog(mutateMultiplierCatalog(t, phase0, vector.Name))
			if gotValid := err == nil; gotValid != vector.ExpectValid {
				t.Fatalf("valid=%v, want %v, error=%v", gotValid, vector.ExpectValid, err)
			}
		})
	}
}

func mutateMultiplierCatalog(t *testing.T, source []byte, name string) []byte {
	t.Helper()
	var root map[string]any
	if err := json.Unmarshal(source, &root); err != nil {
		t.Fatal(err)
	}
	sources := []any{
		map[string]any{"id": "upgrade.a", "slot": "upgrades", "target": "all", "provider": "upgrade.a"},
		map[string]any{"id": "upgrade.b", "slot": "upgrades", "target": "generator.beige_tower", "provider": "upgrade.b"},
	}
	switch name {
	case "valid-multiple-upgrades":
	case "duplicate-multiplier-source":
		sources[1].(map[string]any)["id"] = "upgrade.a"
	case "second-commons-provider":
		sources[0].(map[string]any)["slot"] = "commons"
		sources[1].(map[string]any)["slot"] = "commons"
	case "second-trust-provider":
		sources[0].(map[string]any)["slot"] = "trust"
		sources[1].(map[string]any)["slot"] = "trust"
	case "unknown-multiplier-slot":
		sources[0].(map[string]any)["slot"] = "dark_magic"
	case "unknown-multiplier-target":
		sources[0].(map[string]any)["target"] = "generator.missing"
	case "malformed-multiplier-provider":
		sources[0].(map[string]any)["provider"] = "Upgrade A"
	default:
		t.Fatalf("unimplemented multiplier catalog case %q", name)
	}
	root["multiplier_sources"] = sources
	data, err := json.Marshal(root)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func mutateCatalog(t *testing.T, source json.RawMessage, name string) []byte {
	t.Helper()
	var root map[string]any
	if err := json.Unmarshal(source, &root); err != nil {
		t.Fatal(err)
	}
	resources := root["resources"].([]any)
	generators := root["generator_classes"].([]any)
	resource := resources[0].(map[string]any)
	generator := generators[0].(map[string]any)
	price := generator["price"].(map[string]any)
	production := generator["production"].(map[string]any)

	switch name {
	case "unsupported-version":
		root["schema_version"] = float64(4)
	case "missing-root-field":
		delete(root, "resources")
	case "unknown-root-field":
		root["unexpected"] = true
	case "missing-hardcap":
		delete(resource, "hardcap")
	case "unknown-nested-field":
		resource["unexpected"] = true
	case "duplicate-resource-id":
		root["resources"] = append(resources, resource)
	case "duplicate-generator-id":
		root["generator_classes"] = append(generators, generator)
	case "dangling-resource-id":
		price["resource_id"] = "company.missing"
	case "invalid-id":
		resource["id"] = "Company.Cash"
	case "invalid-scope":
		resource["scope"] = "universe"
	case "invalid-numeric-kind":
		resource["numeric_kind"] = "float64"
	case "invalid-decimal":
		resource["initial"] = "100"
	case "invalid-bounds":
		resource["minimum"] = "1e0"
	case "invalid-curve-kind":
		price["curve"] = map[string]any{"kind": "script"}
	case "invalid-curve-parameter":
		price["curve"] = map[string]any{"kind": "geometric", "ratio": "9e-1"}
	case "missing-production":
		delete(generator, "production")
	case "dangling-production-resource":
		production["resource_id"] = "company.missing"
	case "nonpositive-production-rate":
		production["base_rate"] = "0"
	case "cross-scope-production":
		production["resource_id"] = "founder.reputation"
	default:
		t.Fatalf("unimplemented invalid fixture case %q", name)
	}

	data, err := json.Marshal(root)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestCatalogV1RemainsReadableWithoutProduction(t *testing.T) {
	legacy := []byte(`{
  "schema_version": 1,
  "resources": [{
    "id": "company.cash", "scope": "company", "numeric_kind": "decimal",
    "initial": "0", "minimum": "0", "hardcap": null
  }],
  "generator_classes": [{
    "id": "generator.legacy",
    "price": {"resource_id":"company.cash","base":"1e0","curve":{"kind":"constant"}}
  }]
}`)
	catalog, err := LoadCatalog(legacy)
	if err != nil {
		t.Fatal(err)
	}
	generator, ok := catalog.GeneratorClass("generator.legacy")
	if !ok || generator.Production != nil {
		t.Fatal("legacy generator unexpectedly became production-capable")
	}
	if scoped := catalog.GeneratorClassesForScope(ScopeCompany); len(scoped) != 0 {
		t.Fatalf("legacy scoped generators = %d", len(scoped))
	}
}

func TestProductionCatalogV4Contract(t *testing.T) {
	data, err := os.ReadFile("../../balance/catalogs/phase0.json")
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := LoadCatalog(data)
	if err != nil {
		t.Fatal(err)
	}
	action, ok := catalog.ManualAction("manual.click")
	if !ok || action.Output.ResourceID != "company.cash" || action.Output.AmountPerAction.String() != "1e0" {
		t.Fatalf("manual action = %+v, exists=%v", action, ok)
	}
	wantSources := []MultiplierSourceDefinition{
		{ID: "active.building.generator.beige_tower", Slot: SlotEventBuffs, Target: "generator.beige_tower", Provider: "active_play"},
		{ID: "active.click", Slot: SlotEventBuffs, Target: "manual.click", Provider: "active_play"},
		{ID: "active.production", Slot: SlotEventBuffs, Target: "all", Provider: "active_play"},
		{ID: "commons.member", Slot: SlotCommons, Target: "all", Provider: "commons"},
		{ID: "fiscal.generator.beige_tower", Slot: SlotPrestige, Target: "generator.beige_tower", Provider: "fiscal"},
		{ID: "fiscal.hoard", Slot: SlotPrestige, Target: "all", Provider: "fiscal"},
		{ID: "guild.stock_consumption", Slot: SlotFaction, Target: "all", Provider: "faction"},
		{ID: "pool.institutional_knowledge", Slot: SlotUpgrades, Target: "all", Provider: "pool.institutional_knowledge"},
		{ID: "pool.operational_excellence", Slot: SlotUpgrades, Target: "all", Provider: "pool.operational_excellence"},
	}
	if sources := catalog.MultiplierSources(); len(sources) != 46 || !reflect.DeepEqual(sources[:len(wantSources)], wantSources) {
		t.Fatalf("production multiplier sources = %+v, want 46 with prefix %+v", sources, wantSources)
	}
	if got := catalog.ManualPolicy(); got.RefillMilliPerMS != 25 || got.BucketCapMilli != 50_000 {
		t.Fatalf("manual policy = %+v", got)
	}
	if got := catalog.OfflinePolicy(); got.Efficiency.String() != "9e-1" || got.AccrualCapMS != 86_400_000 ||
		got.BankRatioNumerator != 1 || got.BankRatioDenominator != 2 || got.BankCapMS != 259_200_000 {
		t.Fatalf("offline policy = %+v", got)
	}
	for tier := 0; tier <= 3; tier++ {
		if _, ok := catalog.ProgressCoordinate(tier); !ok {
			t.Fatalf("missing progress coordinate tier %d", tier)
		}
	}
	coordinate, _ := catalog.ProgressCoordinate(1)
	coordinate.Terms[0].Required = 999
	again, _ := catalog.ProgressCoordinate(1)
	if again.Terms[0].Required == 999 {
		t.Fatal("progress coordinate was mutated through accessor")
	}
}

func TestCatalogProductionAccessorsAreScopedAndImmutable(t *testing.T) {
	catalog, err := LoadCatalog(loadKernelFixture(t).Catalog)
	if err != nil {
		t.Fatal(err)
	}
	definitions := catalog.GeneratorClassesForScope(ScopeCompany)
	if len(definitions) != 5 || definitions[0].Production == nil {
		t.Fatalf("company production definitions = %d", len(definitions))
	}
	definitions[0].Production.ResourceID = "changed"
	again, _ := catalog.GeneratorClass(definitions[0].ID)
	if again.Production.ResourceID == "changed" {
		t.Fatal("catalog production was mutated through an accessor")
	}
	if founder := catalog.GeneratorClassesForScope(ScopeFounder); len(founder) != 0 {
		t.Fatalf("founder production definitions = %d", len(founder))
	}
}

func TestCatalogAccessorsDoNotExposeMutableHardcaps(t *testing.T) {
	catalog, err := LoadCatalog(loadKernelFixture(t).Catalog)
	if err != nil {
		t.Fatal(err)
	}
	resource, ok := catalog.Resource("company.cash")
	if !ok || resource.Hardcap == nil {
		t.Fatal("capped resource missing")
	}
	resource.Hardcap.ReasonKey = "changed"
	again, _ := catalog.Resource("company.cash")
	if again.Hardcap.ReasonKey == "changed" {
		t.Fatal("catalog hardcap was mutated through an accessor")
	}
}

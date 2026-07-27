package economy

import (
	"encoding/json"
	"os"
	"testing"

	"cloud-clicker/server/decimal"
)

type kernelFixture struct {
	Version      int             `json:"version"`
	Catalog      json.RawMessage `json:"catalog"`
	CurveVectors []curveVector   `json:"curve_vectors"`
	InvalidCases []string        `json:"invalid_cases"`
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

	switch name {
	case "unsupported-version":
		root["schema_version"] = float64(2)
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
	default:
		t.Fatalf("unimplemented invalid fixture case %q", name)
	}

	data, err := json.Marshal(root)
	if err != nil {
		t.Fatal(err)
	}
	return data
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

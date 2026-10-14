package economy

import (
	"bytes"
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"cloud-clicker/server/publicapi"
)

func TestCostCurveSchemaAdmitsAllVersionedCatalogPricesWithoutRewritingBytes(t *testing.T) {
	definitions, err := publicapi.ValidateSchemaDefinitions(CostCurveSchemas())
	if err != nil {
		t.Fatal(err)
	}
	var sources [][]byte
	sources = append(sources, loadKernelFixture(t).Catalog)
	// A derived v1 compatibility control drops v2 production. It is not an
	// archived v1 artifact or evidence of a historical production workflow.
	var legacy map[string]any
	if err := json.Unmarshal(loadKernelFixture(t).Catalog, &legacy); err != nil {
		t.Fatal(err)
	}
	legacy["schema_version"] = 1
	for _, source := range legacy["generator_classes"].([]any) {
		delete(source.(map[string]any), "production")
	}
	v1, err := json.Marshal(legacy)
	if err != nil {
		t.Fatal(err)
	}
	sources = append(sources, v1)
	for _, path := range []string{
		"../../balance/testdata/first-content/economy-v3.json", "../../balance/catalogs/phase0.json",
		"../../balance/testdata/axis-stack/economy-v5-fixture.json",
		"../../testdata/harness/content-dynamics/bundles/sha256:6c7fab29c24fae68e3067c883177bc78fe61b9d91704b6d936b3e4f3cfd8f789/artifacts/economy.json",
		"../../testdata/harness/content-dynamics/bundles/sha256:baa890501b2864d14cc0238d633a562cb8c6fca406190487831e0c447af128f6/artifacts/economy.json",
	} {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		sources = append(sources, data)
	}
	for _, data := range sources {
		before := bytes.Clone(data)
		if _, err := LoadCatalog(data); err != nil {
			t.Fatal(err)
		}
		var source struct {
			Generators []struct {
				Price struct {
					Curve json.RawMessage `json:"curve"`
				} `json:"price"`
			} `json:"generator_classes"`
		}
		if err := json.Unmarshal(data, &source); err != nil {
			t.Fatal(err)
		}
		for _, generator := range source.Generators {
			if err := publicapi.ValidateJSON(CostCurveSchemaName, generator.Price.Curve, definitions); err != nil {
				t.Fatal(err)
			}
		}
		if !bytes.Equal(before, data) {
			t.Fatal("price admission rewrote artifact bytes")
		}
	}
}

func TestCostCurveSchemaClosedArmsAndOwnerSemanticBounds(t *testing.T) {
	definitions, err := publicapi.ValidateSchemaDefinitions(CostCurveSchemas())
	if err != nil {
		t.Fatal(err)
	}
	for _, source := range []string{
		`{"kind":"constant","step":null}`, `{"kind":"linear","step":"0","ratio":null}`,
		`{"kind":"geometric","ratio":"1e0","step":null}`, `{"KIND":"constant"}`,
		`{"kind":"linear"}`, `{"kind":"geometric","ratio":"1.0"}`, `{"kind":"future"}`,
	} {
		if err := publicapi.ValidateJSON(CostCurveSchemaName, []byte(source), definitions); err == nil {
			t.Fatal("descriptor admitted invalid curve", source)
		}
	}
	for _, source := range []string{`{"kind":"linear","step":"-1e0"}`, `{"kind":"geometric","ratio":"5e-1"}`} {
		if err := publicapi.ValidateJSON(CostCurveSchemaName, []byte(source), definitions); err != nil {
			t.Fatal(err)
		}
		var raw rawCurve
		if err := json.Unmarshal([]byte(source), &raw); err != nil {
			t.Fatal(err)
		}
		if _, err := parseCurve(raw); err == nil {
			t.Fatal("schema replaced nonnegative step/unit ratio semantics")
		}
	}
	before, fresh := CostCurveSchemas(), CostCurveSchemas()
	if !reflect.DeepEqual(before, fresh) {
		t.Fatal("nondeterministic exports")
	}
	fresh[0].Schema.Alternates[0].Fields[0].Schema.Enum[0] = "future"
	if !reflect.DeepEqual(before, CostCurveSchemas()) {
		t.Fatal("caller mutated curve authority")
	}
	data := costCurveCandidate(t, map[string]any{"kind": "geometric", "ratio": "1e0"})
	escaped := bytes.Replace(data, []byte(`"kind":"geometric"`), []byte(`"k\u0069nd":"geometric"`), 1)
	if bytes.Equal(data, escaped) {
		t.Fatal("escape control did not change bytes")
	}
	if _, err := LoadCatalog(escaped); err != nil {
		t.Fatal(err)
	}
	duplicate := bytes.Replace(data, []byte(`"kind":"geometric","ratio":"1e0"`), []byte(`"kind":"constant","kind":"geometric","ratio":"1e0"`), 1)
	if bytes.Equal(data, duplicate) {
		t.Fatal("last-value control did not change bytes")
	}
	if _, err := LoadCatalog(duplicate); err != nil {
		t.Fatal("existing last-value member behavior changed", err)
	}
}

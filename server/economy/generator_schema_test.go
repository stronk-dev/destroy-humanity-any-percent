package economy

import (
	"bytes"
	"encoding/json"
	"os"
	"reflect"
	"sort"
	"testing"

	"cloud-clicker/server/publicapi"
)

func generatorPolicyDefinitions(t *testing.T) map[string]*publicapi.Schema {
	t.Helper()
	rows := append(CostCurveSchemas(), GeneratorSchemas()...)
	rows = append(rows, OperatingPolicySchemas()...)
	sort.Slice(rows, func(i, j int) bool { return rows[i].Name < rows[j].Name })
	definitions, err := publicapi.ValidateSchemaDefinitions(rows)
	if err != nil {
		t.Fatal(err)
	}
	return definitions
}

func validateGeneratorPolicyComponents(data []byte, definitions map[string]*publicapi.Schema) error {
	var source struct {
		Version    int               `json:"schema_version"`
		Generators []json.RawMessage `json:"generator_classes"`
		Manual     json.RawMessage   `json:"manual_policy"`
		Offline    json.RawMessage   `json:"offline_policy"`
	}
	if err := json.Unmarshal(data, &source); err != nil {
		return err
	}
	name := "EconomyGeneratorV1"
	if source.Version >= 2 {
		name = "EconomyGeneratorV2"
	}
	if source.Version >= 4 {
		name = "EconomyGeneratorV4"
	}
	for _, row := range source.Generators {
		if err := publicapi.ValidateJSON(name, row, definitions); err != nil {
			return err
		}
	}
	if source.Version >= 3 {
		if err := publicapi.ValidateJSON("EconomyManualPolicyV1", source.Manual, definitions); err != nil {
			return err
		}
		return publicapi.ValidateJSON("EconomyOfflinePolicyV1", source.Offline, definitions)
	}
	return nil
}

func TestGeneratorPolicySchemaVersionedCompatibility(t *testing.T) {
	definitions := generatorPolicyDefinitions(t)
	for _, path := range []string{
		"../../balance/catalogs/phase0.json", "../../balance/testdata/first-content/economy-v3.json", "../../balance/testdata/axis-stack/economy-v5-fixture.json",
		"../../testdata/harness/content-dynamics/bundles/sha256:6c7fab29c24fae68e3067c883177bc78fe61b9d91704b6d936b3e4f3cfd8f789/artifacts/economy.json",
		"../../testdata/harness/content-dynamics/bundles/sha256:baa890501b2864d14cc0238d633a562cb8c6fca406190487831e0c447af128f6/artifacts/economy.json",
	} {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		before := bytes.Clone(data)
		if _, err := LoadCatalog(data); err != nil {
			t.Fatal(path, err)
		}
		if err := validateGeneratorPolicyComponents(data, definitions); err != nil {
			t.Fatal(path, err)
		}
		if !bytes.Equal(before, data) {
			t.Fatal("validation rewrote source")
		}
	}
	data := loadKernelFixture(t).Catalog // shared version-2 fixture
	if err := validateGeneratorPolicyComponents(data, definitions); err != nil {
		t.Fatal(err)
	}
	var legacy map[string]any
	if err := json.Unmarshal(data, &legacy); err != nil {
		t.Fatal(err)
	}
	legacy["schema_version"] = 1
	for _, row := range legacy["generator_classes"].([]any) {
		delete(row.(map[string]any), "production")
	}
	derived, err := json.Marshal(legacy) // derived compatibility control, not a historical artifact
	if err != nil {
		t.Fatal(err)
	}
	if _, err := LoadCatalog(derived); err != nil {
		t.Fatal(err)
	}
	if err := validateGeneratorPolicyComponents(derived, definitions); err != nil {
		t.Fatal(err)
	}
	var producing struct {
		Generators []json.RawMessage `json:"generator_classes"`
	}
	if err := json.Unmarshal(data, &producing); err != nil {
		t.Fatal(err)
	}
	if err := publicapi.ValidateJSON("EconomyGeneratorV1", producing.Generators[0], definitions); err == nil {
		t.Fatal("v1 admitted production")
	}
	if err := validateGeneratorPolicyComponents(foundationCatalogBytes(t), definitions); err != nil {
		t.Fatal(err)
	}
}

func TestGeneratorPolicySchemaClosedShapesAndOwnerRelations(t *testing.T) {
	definitions := generatorPolicyDefinitions(t)
	first := func(v map[string]any) map[string]any { return v["generator_classes"].([]any)[0].(map[string]any) }
	for name, edit := range map[string]func(map[string]any){
		"alias":                func(v map[string]any) { row := first(v); row["ID"] = row["id"]; delete(row, "id") },
		"private member":       func(v map[string]any) { first(v)["owner_id"] = "private" },
		"null production":      func(v map[string]any) { first(v)["production"] = nil },
		"null optional object": func(v map[string]any) { first(v)["provisions"] = nil },
		"null roles":           func(v map[string]any) { first(v)["roles"] = nil },
		"tier range":           func(v map[string]any) { first(v)["tier"] = 4 },
		"price number":         func(v map[string]any) { first(v)["price"].(map[string]any)["base"] = 1 },
		"wrong role arm":       func(v map[string]any) { first(v)["roles"].([]any)[0].(map[string]any)["generator_id"] = nil },
		"zero ladder count":    func(v map[string]any) { first(v)["ladder"].([]any)[0].(map[string]any)["purchased_at"] = 0 },
		"null manual":          func(v map[string]any) { v["manual_policy"] = nil },
		"zero refill":          func(v map[string]any) { v["manual_policy"].(map[string]any)["refill_milli_per_ms"] = 0 },
		"offline alias": func(v map[string]any) {
			row := v["offline_policy"].(map[string]any)
			row["EFFICIENCY"] = row["efficiency"]
			delete(row, "efficiency")
		},
		"unsafe duration": func(v map[string]any) {
			v["offline_policy"].(map[string]any)["accrual_cap_ms"] = int64(9007199254740992)
		},
	} {
		t.Run(name, func(t *testing.T) {
			data := editFoundationCatalog(t, edit)
			if err := validateGeneratorPolicyComponents(data, definitions); err == nil {
				t.Fatal("schema admitted malformed component")
			}
			if _, err := LoadCatalog(data); err == nil {
				t.Fatal("owner admitted malformed component")
			}
		})
	}
	for name, edit := range map[string]func(map[string]any){
		"zero price":           func(v map[string]any) { first(v)["price"].(map[string]any)["base"] = "0" },
		"negative production":  func(v map[string]any) { first(v)["production"].(map[string]any)["base_rate"] = "-1e0" },
		"neutral ladder":       func(v map[string]any) { first(v)["ladder"].([]any)[0].(map[string]any)["multiplier_ppm"] = 1000000 },
		"empty roles":          func(v map[string]any) { first(v)["roles"] = []any{} },
		"efficiency above one": func(v map[string]any) { v["offline_policy"].(map[string]any)["efficiency"] = "2e0" },
		"zero burst":           func(v map[string]any) { v["offline_policy"].(map[string]any)["burst_speed"] = "0" },
		"bank ratio": func(v map[string]any) {
			row := v["offline_policy"].(map[string]any)
			row["bank_ratio_numerator"] = row["bank_ratio_denominator"].(float64) + 1
		},
	} {
		t.Run(name, func(t *testing.T) {
			data := editFoundationCatalog(t, edit)
			if err := validateGeneratorPolicyComponents(data, definitions); err != nil {
				t.Fatal("structural descriptor took over semantic ownership", err)
			}
			if _, err := LoadCatalog(data); err == nil {
				t.Fatal("descriptor replaced required owner check")
			}
		})
	}
	data := editFoundationCatalog(t, func(map[string]any) {})
	for name, replacement := range map[string]string{"escaped": `"ba\u0073e_rate":`, "duplicate": `"base_rate":"0","base_rate":`} {
		changed := bytes.Replace(data, []byte(`"base_rate":`), []byte(replacement), 1)
		if bytes.Equal(data, changed) {
			t.Fatal(name, "did not modify source")
		}
		if _, err := LoadCatalog(changed); err != nil {
			t.Fatal(name, err)
		}
		if err := validateGeneratorPolicyComponents(changed, definitions); err != nil {
			t.Fatal(name, err)
		}
	}
	before, fresh := GeneratorSchemas(), GeneratorSchemas()
	if !reflect.DeepEqual(before, fresh) {
		t.Fatal("unstable export")
	}
	for _, row := range fresh {
		if row.Name == "EconomyGeneratorV4" {
			for _, field := range row.Schema.Fields {
				if field.Name == "tier" {
					*field.Schema.Maximum = 99
				}
			}
		}
	}
	if !reflect.DeepEqual(before, GeneratorSchemas()) {
		t.Fatal("caller mutated generator authority")
	}
	policies, changedPolicies := OperatingPolicySchemas(), OperatingPolicySchemas()
	*changedPolicies[0].Schema.Fields[0].Schema.Minimum = 0
	if !reflect.DeepEqual(policies, OperatingPolicySchemas()) {
		t.Fatal("caller mutated policy authority")
	}
}

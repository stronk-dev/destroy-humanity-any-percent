package economy

import (
	"bytes"
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"cloud-clicker/server/publicapi"
)

func resourceSchemaDefinitions(t *testing.T) map[string]*publicapi.Schema {
	t.Helper()
	definitions, err := publicapi.ValidateSchemaDefinitions(ResourceSchemas())
	if err != nil {
		t.Fatal(err)
	}
	return definitions
}

func TestResourceSchemaAdmitsCurrentAndStoredRowsWithoutRewritingBytes(t *testing.T) {
	definitions := resourceSchemaDefinitions(t)
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
			t.Fatal(err)
		}
		var source struct {
			Resources []json.RawMessage `json:"resources"`
		}
		if err := json.Unmarshal(data, &source); err != nil {
			t.Fatal(err)
		}
		for _, row := range source.Resources {
			if err := publicapi.ValidateJSON(ResourceSchemaName, row, definitions); err != nil {
				t.Fatal(err)
			}
		}
		if !bytes.Equal(before, data) {
			t.Fatal("resource validation changed original bytes")
		}
	}
	for _, scope := range []Scope{ScopeCompany, ScopeFounder, ScopeGuild, ScopeWorld} {
		for _, cap := range []any{nil, map[string]any{"amount": "1e3", "reason_key": "cap.example"}} {
			row, err := json.Marshal(map[string]any{"id": "resource.example", "scope": scope, "numeric_kind": "decimal", "initial": "0", "minimum": "0", "hardcap": cap})
			if err != nil {
				t.Fatal(err)
			}
			if err := publicapi.ValidateJSON(ResourceSchemaName, row, definitions); err != nil {
				t.Fatal(err)
			}
			var raw rawResource
			if err := json.Unmarshal(row, &raw); err != nil {
				t.Fatal(err)
			}
			if _, err := parseResource(raw); err != nil {
				t.Fatal(err)
			}
		}
	}
}

func TestResourceSchemaRequiresClosedNullableCapAndRetainsRelations(t *testing.T) {
	definitions := resourceSchemaDefinitions(t)
	var source struct {
		Resources []json.RawMessage `json:"resources"`
	}
	if err := json.Unmarshal(loadKernelFixture(t).Catalog, &source); err != nil {
		t.Fatal(err)
	}
	for name, edit := range map[string]func(map[string]any){
		"missing cap":     func(v map[string]any) { delete(v, "hardcap") },
		"private row":     func(v map[string]any) { v["owner_id"] = "private" },
		"row alias":       func(v map[string]any) { v["ID"] = v["id"]; delete(v, "id") },
		"invalid scope":   func(v map[string]any) { v["scope"] = "account" },
		"numeric kind":    func(v map[string]any) { v["numeric_kind"] = "float" },
		"initial number":  func(v map[string]any) { v["initial"] = 0 },
		"null minimum":    func(v map[string]any) { v["minimum"] = nil },
		"null cap amount": func(v map[string]any) { v["hardcap"].(map[string]any)["amount"] = nil },
		"cap alias": func(v map[string]any) {
			cap := v["hardcap"].(map[string]any)
			cap["AMOUNT"] = cap["amount"]
			delete(cap, "amount")
		},
		"cap extra": func(v map[string]any) { v["hardcap"].(map[string]any)["label"] = "private" },
	} {
		t.Run(name, func(t *testing.T) {
			var row map[string]any
			if err := json.Unmarshal(source.Resources[0], &row); err != nil {
				t.Fatal(err)
			}
			edit(row)
			data, err := json.Marshal(row)
			if err != nil {
				t.Fatal(err)
			}
			if err := publicapi.ValidateJSON(ResourceSchemaName, data, definitions); err == nil {
				t.Fatal("descriptor admitted invalid row")
			}
			if _, err := LoadCatalog(resourceCandidate(t, func(v map[string]any) {
				for key := range v {
					delete(v, key)
				}
				for key, value := range row {
					v[key] = value
				}
			})); err == nil {
				t.Fatal("owner admitted invalid row")
			}
		})
	}
	for name, edit := range map[string]func(map[string]any){
		"initial below minimum": func(v map[string]any) { v["minimum"] = "1e0" },
		"cap below minimum": func(v map[string]any) {
			v["initial"], v["minimum"] = "1e0", "1e0"
			v["hardcap"].(map[string]any)["amount"] = "0"
		},
		"initial above cap": func(v map[string]any) { v["initial"] = "1e1"; v["hardcap"].(map[string]any)["amount"] = "1e0" },
	} {
		t.Run(name, func(t *testing.T) {
			data := resourceCandidate(t, edit)
			var root struct {
				Resources []json.RawMessage `json:"resources"`
			}
			if err := json.Unmarshal(data, &root); err != nil {
				t.Fatal(err)
			}
			if err := publicapi.ValidateJSON(ResourceSchemaName, root.Resources[0], definitions); err != nil {
				t.Fatal(err)
			}
			if _, err := LoadCatalog(data); err == nil {
				t.Fatal("schema replaced ledger-resource relations")
			}
		})
	}
	data := resourceCandidate(t, func(map[string]any) {})
	escaped := bytes.Replace(data, []byte(`"scope":"company"`), []byte(`"sc\u006fpe":"company"`), 1)
	if bytes.Equal(data, escaped) {
		t.Fatal("escape control did not change source")
	}
	if _, err := LoadCatalog(escaped); err != nil {
		t.Fatal(err)
	}
	duplicate := bytes.Replace(data, []byte(`"numeric_kind":"decimal"`), []byte(`"numeric_kind":"decimal","numeric_kind":"decimal"`), 1)
	if bytes.Equal(data, duplicate) {
		t.Fatal("duplicate control did not change source")
	}
	if _, err := LoadCatalog(duplicate); err != nil {
		t.Fatal("last-value behavior changed", err)
	}
	before, fresh := ResourceSchemas(), ResourceSchemas()
	if !reflect.DeepEqual(before, fresh) {
		t.Fatal("nondeterministic exports")
	}
	for _, field := range fresh[0].Schema.Fields {
		if field.Name == "scope" {
			field.Schema.Enum[0] = "account"
		}
	}
	if !reflect.DeepEqual(before, ResourceSchemas()) {
		t.Fatal("caller changed resource authority")
	}
}

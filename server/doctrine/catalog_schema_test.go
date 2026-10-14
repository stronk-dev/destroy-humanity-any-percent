package doctrine

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"cloud-clicker/server/publicapi"
	"cloud-clicker/server/routes"
)

func doctrineSchemaDefinitions(t *testing.T) map[string]*publicapi.Schema {
	t.Helper()
	definitions, err := publicapi.ValidateSchemaDefinitions(DoctrineSchemas())
	if err != nil {
		t.Fatal(err)
	}
	return definitions
}

func TestDoctrineOwnerSchemaAdmitsCurrentStoredAndFixtureWithoutChangingBytes(t *testing.T) {
	definitions := doctrineSchemaDefinitions(t)
	paths := []string{
		"../../balance/doctrines/first-content.json",
		"../../testdata/harness/content-dynamics/bundles/sha256:6c7fab29c24fae68e3067c883177bc78fe61b9d91704b6d936b3e4f3cfd8f789/artifacts/doctrines.json",
		"../../testdata/harness/content-dynamics/bundles/sha256:baa890501b2864d14cc0238d633a562cb8c6fca406190487831e0c447af128f6/artifacts/doctrines.json",
	}
	for index, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		before := bytes.Clone(data)
		if err := publicapi.ValidateJSON(DoctrineCatalogSchemaName, data, definitions); err != nil {
			t.Fatalf("%s: %v", path, err)
		}
		catalog, err := LoadCatalog(data)
		if err != nil {
			t.Fatal(err)
		}
		routePath := "../../balance/routes/phase0.json"
		if index > 0 {
			routePath = filepath.Join(filepath.Dir(path), "routes.json")
		}
		routeBytes, err := os.ReadFile(routePath)
		if err != nil {
			t.Fatal(err)
		}
		routeCatalog, err := routes.LoadCatalog(routeBytes)
		if err != nil {
			t.Fatal(err)
		}
		if err := catalog.ValidateRoutes(routeCatalog); err != nil {
			t.Fatalf("%s: %v", path, err)
		}
		if !bytes.Equal(before, data) {
			t.Fatal("validation changed original artifact bytes")
		}
	}
	if err := publicapi.ValidateJSON(DoctrineCatalogSchemaName, doctrineFixture(t), definitions); err != nil {
		t.Fatal(err)
	}
}

func TestDoctrineOwnerSchemaClosedShapesAndTierBounds(t *testing.T) {
	type object = map[string]any
	row := func(v object) object { return v["transitions"].([]any)[0].(object) }
	definitions := doctrineSchemaDefinitions(t)
	for name, edit := range map[string]func(object){
		"root missing":     func(v object) { delete(v, "schema_version") },
		"root open":        func(v object) { v["account_id"] = "private" },
		"root alias":       func(v object) { v["Schema_version"] = v["schema_version"]; delete(v, "schema_version") },
		"version":          func(v object) { v["schema_version"] = 2 },
		"null transitions": func(v object) { v["transitions"] = nil },
		"row missing":      func(v object) { delete(row(v), "source_tier") },
		"row open":         func(v object) { row(v)["effect"] = "fictional" },
		"row alias":        func(v object) { row(v)["Source_tier"] = row(v)["source_tier"]; delete(row(v), "source_tier") },
		"null tier":        func(v object) { row(v)["source_tier"] = nil },
		"negative tier":    func(v object) { row(v)["source_tier"] = -1 },
		"tier above":       func(v object) { row(v)["source_tier"] = 9 },
		"fractional tier":  func(v object) { row(v)["source_tier"] = .5 },
		"transition ID":    func(v object) { row(v)["transition_id"] = "bad-id" },
		"gate ID":          func(v object) { row(v)["gate_id"] = "bad-id" },
		"null choices":     func(v object) { row(v)["doctrine_ids"] = nil },
		"choice ID":        func(v object) { row(v)["doctrine_ids"] = []string{"bad-id", "doctrine.ethical"} },
	} {
		t.Run(name, func(t *testing.T) {
			data := editDoctrineFixture(t, edit)
			if err := publicapi.ValidateJSON(DoctrineCatalogSchemaName, data, definitions); err == nil {
				t.Fatal("descriptor admitted invalid shape")
			}
			if _, err := LoadCatalog(data); err == nil {
				t.Fatal("loader admitted invalid shape")
			}
		})
	}
	for _, tier := range []int{0, 8} {
		data := editDoctrineFixture(t, func(v object) {
			r := row(v)
			if tier == 0 {
				r["transition_id"], r["gate_id"] = "transition.t0_to_t1", "gate.t0_to_t1"
			} else {
				r["transition_id"], r["gate_id"] = "transition.t8_to_t9", "gate.t8_to_t9"
			}
			r["source_tier"] = tier
		})
		if err := publicapi.ValidateJSON(DoctrineCatalogSchemaName, data, definitions); err != nil {
			t.Fatal(err)
		}
		if _, err := LoadCatalog(data); err != nil {
			t.Fatal(err)
		}
	}
}

func TestDoctrineOwnerSchemaDoesNotReplaceSemanticLoader(t *testing.T) {
	definitions := doctrineSchemaDefinitions(t)
	for name, edit := range map[string]func(map[string]any){
		"empty transitions":  func(v map[string]any) { v["transitions"] = []any{} },
		"boundary agreement": func(v map[string]any) { v["transitions"].([]any)[0].(map[string]any)["source_tier"] = 2 },
		"branching": func(v map[string]any) {
			v["transitions"].([]any)[0].(map[string]any)["doctrine_ids"] = []string{"doctrine.capture"}
		},
		"ordering": func(v map[string]any) {
			v["transitions"].([]any)[0].(map[string]any)["doctrine_ids"] = []string{"doctrine.ethical", "doctrine.capture"}
		},
	} {
		t.Run(name, func(t *testing.T) {
			data := editDoctrineFixture(t, edit)
			if err := publicapi.ValidateJSON(DoctrineCatalogSchemaName, data, definitions); err != nil {
				t.Fatal(err)
			}
			if _, err := LoadCatalog(data); err == nil {
				t.Fatal("structural descriptor replaced semantic validation")
			}
		})
	}
}

func TestDoctrineOwnerSchemaFreshExportsAndOriginalMemberControls(t *testing.T) {
	before := DoctrineSchemas()
	fresh := DoctrineSchemas()
	if !reflect.DeepEqual(before, fresh) {
		t.Fatal("non-deterministic export")
	}
	*fresh[0].Schema.Fields[0].Schema.Minimum = 2
	if !reflect.DeepEqual(before, DoctrineSchemas()) {
		t.Fatal("caller mutation changed owner authority")
	}
	definitions := doctrineSchemaDefinitions(t)
	source := bytes.TrimSpace(doctrineFixture(t))
	for _, data := range [][]byte{
		bytes.Replace(source, []byte(`"schema_version"`), []byte(`"schema_versi\u006fn"`), 1),
		append(bytes.Clone(source[:len(source)-1]), []byte(`,"schema_version":2,"schema_version":1}`)...),
	} {
		if bytes.Equal(source, data) {
			t.Fatal("member control did not change source")
		}
		if err := publicapi.ValidateJSON(DoctrineCatalogSchemaName, data, definitions); err != nil {
			t.Fatal(err)
		}
		if _, err := LoadCatalog(data); err != nil {
			t.Fatal(err)
		}
	}
	data := append(bytes.Clone(source[:len(source)-1]), []byte(`,"schema_version":null}`)...)
	if err := publicapi.ValidateJSON(DoctrineCatalogSchemaName, data, definitions); err == nil {
		t.Fatal("descriptor admitted final null version")
	}
}

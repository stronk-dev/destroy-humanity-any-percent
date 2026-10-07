package leaderboard

import (
	"bytes"
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"cloud-clicker/server/publicapi"
)

func categorySchemaDefinitions(t *testing.T) map[string]*publicapi.Schema {
	t.Helper()
	definitions, err := publicapi.ValidateSchemaDefinitions(CategorySchemas())
	if err != nil {
		t.Fatal(err)
	}
	return definitions
}

func TestCategoryOwnerSchemaAdmitsStoredAndCandidateCatalogsWithoutChangingBytes(t *testing.T) {
	definitions := categorySchemaDefinitions(t)
	for _, path := range []string{
		"../../balance/categories/phase0.json",
		"../../balance/testdata/epoch5/categories.json",
		"../../balance/testdata/first-content/categories-v1.json",
		"../../balance/testdata/t0-t1/categories-v1.json",
		"../../balance/testdata/t2/categories-candidate-v1.json",
	} {
		t.Run(path, func(t *testing.T) {
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			before := bytes.Clone(data)
			if err := publicapi.ValidateJSON(CategoryCatalogSchemaName, data, definitions); err != nil {
				t.Fatal(err)
			}
			var envelope struct {
				FullGateSet []string `json:"full_gate_set"`
			}
			if err := json.Unmarshal(data, &envelope); err != nil {
				t.Fatal(err)
			}
			if _, err := LoadCategoryCatalog(data, envelope.FullGateSet); err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(data, before) {
				t.Fatal("validation rewrote stored evidence")
			}
		})
	}
}

func TestCategoryOwnerSchemaRejectsUndeclaredAndWrongRowShapes(t *testing.T) {
	definitions := categorySchemaDefinitions(t)
	data, err := os.ReadFile("../../balance/categories/phase0.json")
	if err != nil {
		t.Fatal(err)
	}
	type object = map[string]any
	row := func(v object, index int) object { return v["categories"].([]any)[index].(object) }
	predicate := func(v object, index int) object { return row(v, index)["predicate"].(object) }
	for _, test := range []struct {
		name string
		edit func(object)
	}{
		{"missing root field", func(v object) { delete(v, "full_gate_set") }},
		{"extra root field", func(v object) { v["private_note"] = true }},
		{"unsupported version", func(v object) { v["schema_version"] = 2 }},
		{"null rows", func(v object) { v["categories"] = nil }},
		{"invalid gate id", func(v object) { v["full_gate_set"] = []any{"Not a gate"} }},
		{"missing fact set", func(v object) { delete(v["fact_sets"].(object), "completion_set") }},
		{"open fact sets", func(v object) { v["fact_sets"].(object)["private_set"] = []any{} }},
		{"nonstring fact", func(v object) { v["fact_sets"].(object)["forbidden_set"] = []any{true} }},
		{"unknown category", func(v object) { row(v, 0)["id"] = "future_category" }},
		{"wrong name binding", func(v object) { row(v, 0)["name_key"] = "category.valuation" }},
		{"wrong timer binding", func(v object) { row(v, 0)["timer"] = "none" }},
		{"missing predicate", func(v object) { delete(row(v, 0), "predicate") }},
		{"extra row field", func(v object) { row(v, 0)["account_id"] = "private" }},
		{"open predicate", func(v object) { predicate(v, 0)["script"] = "arbitrary" }},
		{"wrong predicate binding", func(v object) { predicate(v, 0)["kind"] = "all_gates" }},
		{"unknown predicate", func(v object) { predicate(v, 0)["kind"] = "script" }},
		{"missing all", func(v object) { delete(predicate(v, 1), "all") }},
		{"recursive predicate", func(v object) { predicate(v, 1)["all"] = []any{object{"kind": "all_of", "all": []any{}}} }},
		{"extra child field", func(v object) { predicate(v, 1)["all"].([]any)[0].(object)["private_note"] = true }},
		{"wrong set binding", func(v object) { predicate(v, 2)["set_ref"] = "completion_set" }},
		{"wrong count field", func(v object) { predicate(v, 3)["field"] = "account_count" }},
		{"negative count", func(v object) { predicate(v, 3)["literal"] = -1 }},
		{"fractional count", func(v object) { predicate(v, 3)["literal"] = json.Number("1.5") }},
		{"decimal spelled integer", func(v object) { predicate(v, 3)["literal"] = json.Number("40.0") }},
		{"overflow count", func(v object) { predicate(v, 3)["literal"] = json.Number("9223372036854775808") }},
		{"string count", func(v object) { predicate(v, 3)["literal"] = "40" }},
	} {
		t.Run(test.name, func(t *testing.T) {
			var value object
			decoder := json.NewDecoder(bytes.NewReader(data))
			decoder.UseNumber()
			if err := decoder.Decode(&value); err != nil {
				t.Fatal(err)
			}
			test.edit(value)
			malformed, err := json.Marshal(value)
			if err != nil {
				t.Fatal(err)
			}
			before := bytes.Clone(malformed)
			if err := publicapi.ValidateJSON(CategoryCatalogSchemaName, malformed, definitions); err == nil {
				t.Fatal("descriptor admitted an undeclared or wrong row shape")
			}
			if !bytes.Equal(malformed, before) {
				t.Fatal("refusal modified evidence")
			}
		})
	}
}

func TestCategoryOwnerSchemaLeavesSemanticPolicyWithLoader(t *testing.T) {
	data, err := os.ReadFile("../../balance/categories/phase0.json")
	if err != nil {
		t.Fatal(err)
	}
	data = bytes.Replace(data, []byte(`"literal": 40`), []byte(`"literal": 41`), 1)
	if err := publicapi.ValidateJSON(CategoryCatalogSchemaName, data, categorySchemaDefinitions(t)); err != nil {
		t.Fatal(err)
	}
	gates := []string{"gate.t0_to_t1", "gate.t2_to_t3", "gate.t3_to_t4", "gate.t4_to_t5", "gate.t7_to_t8"}
	if _, err := LoadCategoryCatalog(data, gates); err == nil {
		t.Fatal("wire grammar substituted for canonical policy validation")
	}
}

func TestCategoryOwnerSchemasExportFreshDeterministicDefinitions(t *testing.T) {
	first, second := CategorySchemas(), CategorySchemas()
	if !reflect.DeepEqual(first, second) {
		t.Fatal("nondeterministic schema export")
	}
	first[0].Schema.Fields[0].Name = "mutated"
	if !reflect.DeepEqual(second, CategorySchemas()) {
		t.Fatal("caller mutation changed subsequent owner exports")
	}
}

package achievements

import (
	"bytes"
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"cloud-clicker/server/publicapi"
)

func achievementSchemaDefinitions(t *testing.T) map[string]*publicapi.Schema {
	t.Helper()
	definitions, err := publicapi.ValidateSchemaDefinitions(AchievementSchemas())
	if err != nil {
		t.Fatal(err)
	}
	return definitions
}

func TestAchievementOwnerSchemaAdmitsExistingArtifactsAndFixtureWithoutChangingBytes(t *testing.T) {
	definitions := achievementSchemaDefinitions(t)
	paths := []string{
		"../../balance/achievements/first-content.json",
		"../../testdata/harness/content-dynamics/bundles/sha256:6c7fab29c24fae68e3067c883177bc78fe61b9d91704b6d936b3e4f3cfd8f789/artifacts/achievements.json",
		"../../testdata/harness/content-dynamics/bundles/sha256:baa890501b2864d14cc0238d633a562cb8c6fca406190487831e0c447af128f6/artifacts/achievements.json",
	}
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		before := bytes.Clone(data)
		if err := publicapi.ValidateJSON(AchievementCatalogSchemaName, data, definitions); err != nil {
			t.Fatalf("%s: %v", path, err)
		}
		if !bytes.Equal(before, data) {
			t.Fatal("validation changed artifact bytes")
		}
	}
	data := validCatalogBytes(t)
	if err := publicapi.ValidateJSON(AchievementCatalogSchemaName, data, definitions); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadCatalog(data, testRegistry()); err != nil {
		t.Fatal(err)
	}
}

func TestAchievementOwnerSchemaClosedShapesAndConditionDepth(t *testing.T) {
	definitions := achievementSchemaDefinitions(t)
	type object = map[string]any
	first := func(v object) object { return v["achievements"].([]any)[0].(object) }
	for _, test := range []struct {
		name string
		edit func(object)
	}{
		{"missing root", func(v object) { delete(v, "schema_version") }},
		{"private root", func(v object) { v["account_id"] = "private" }},
		{"wrong version", func(v object) { v["schema_version"] = 2 }},
		{"null rows", func(v object) { v["achievements"] = nil }},
		{"missing row", func(v object) { delete(first(v), "proof") }},
		{"private row", func(v object) { first(v)["founder_id"] = "private" }},
		{"wrong scope", func(v object) { first(v)["condition_scope"] = "founder" }},
		{"zero score", func(v object) { first(v)["score_grant"] = 0 }},
		{"unsafe score", func(v object) { first(v)["score_grant"] = json.Number("9007199254740992") }},
		{"fractional score", func(v object) { first(v)["score_grant"] = .5 }},
		{"null copy", func(v object) { first(v)["copy_key"] = nil }},
		{"wrong condition", func(v object) { first(v)["condition"].(object)["kind"] = "script" }},
		{"open condition", func(v object) { first(v)["condition"].(object)["script"] = "arbitrary" }},
		{"wrong proof", func(v object) { first(v)["proof"].(object)["kind"] = "clout" }},
		{"missing proof field", func(v object) { delete(first(v)["proof"].(object), "event_kinds") }},
		{"open proof", func(v object) { first(v)["proof"].(object)["resource_id"] = "company.cash" }},
	} {
		t.Run(test.name, func(t *testing.T) {
			data := mutateCatalog(t, test.edit)
			if err := publicapi.ValidateJSON(AchievementCatalogSchemaName, data, definitions); err == nil {
				t.Fatal("descriptor admitted wrong shape")
			}
		})
	}
	for depth := 1; depth <= maximumConditionDepth+1; depth++ {
		condition := object{"kind": "fact_present", "fact_kind": "gate.tier_1"}
		for level := 1; level < depth; level++ {
			condition = object{"kind": "all_of", "conditions": []any{condition, condition}}
		}
		data := mutateCatalog(t, func(v object) { first(v)["condition"] = condition })
		shapeErr := publicapi.ValidateJSON(AchievementCatalogSchemaName, data, definitions)
		_, ownerErr := LoadCatalog(data, testRegistry())
		if (shapeErr == nil) != (depth <= maximumConditionDepth) || (ownerErr == nil) != (depth <= maximumConditionDepth) {
			t.Fatalf("depth %d: shape=%v owner=%v", depth, shapeErr, ownerErr)
		}
	}
}

func TestAchievementOwnerSchemaDoesNotReplaceSemanticLoader(t *testing.T) {
	data := mutateCatalog(t, func(v map[string]any) { v["achievements"].([]any)[0].(map[string]any)["copy_key"] = "missing.copy" })
	if err := publicapi.ValidateJSON(AchievementCatalogSchemaName, data, achievementSchemaDefinitions(t)); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadCatalog(data, testRegistry()); err == nil {
		t.Fatal("schema validation substituted for copy/registry binding")
	}
}

func TestAchievementOwnerSchemaAllConditionAndProofArms(t *testing.T) {
	type object = map[string]any
	definitions := achievementSchemaDefinitions(t)
	for _, test := range []struct {
		name      string
		condition object
		scope     string
		proof     object
		source    string
	}{
		{"run counter", object{"kind": "counter_at_least", "counter": "tier", "minimum": 0}, "run", object{"kind": "provenance", "event_kinds": []string{"gate_crossed"}}, "counter:run:tier"},
		{"career counter", object{"kind": "counter_at_least", "counter": "age_ms", "minimum": 1}, "career", object{"kind": "provenance", "event_kinds": []string{"gate_crossed"}}, "counter:career:age_ms"},
		{"career exits", object{"kind": "exit_count_at_least", "count": 1}, "career", object{"kind": "provenance", "event_kinds": []string{"gate_crossed"}}, "exit_count"},
		{"burn", object{"kind": "fact_present", "fact_kind": "gate.tier_1"}, "run", object{"kind": "burn", "event_kind": "gate_crossed", "resource_id": "company.cash", "minimum": "1e0"}, ""},
	} {
		t.Run(test.name, func(t *testing.T) {
			registry := testRegistry()
			if test.source != "" {
				registry.ProvenanceSources[test.source] = []string{"gate_crossed"}
			}
			data := mutateCatalog(t, func(v object) {
				row := v["achievements"].([]any)[0].(object)
				row["condition"] = test.condition
				row["condition_scope"] = test.scope
				row["proof"] = test.proof
			})
			if err := publicapi.ValidateJSON(AchievementCatalogSchemaName, data, definitions); err != nil {
				t.Fatal(err)
			}
			if _, err := LoadCatalog(data, registry); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestAchievementOwnerSchemasAreFreshAndDeterministic(t *testing.T) {
	first, second := AchievementSchemas(), AchievementSchemas()
	if !reflect.DeepEqual(first, second) {
		t.Fatal("unstable export")
	}
	first[0].Schema.Fields[0].Name = "mutated"
	if !reflect.DeepEqual(second, AchievementSchemas()) {
		t.Fatal("caller mutation changed later exports")
	}
}

func TestAchievementCatalogPreservesEscapingAndLastValueSemantics(t *testing.T) {
	data := validCatalogBytes(t)
	for _, test := range []struct{ before, after string }{
		{`"id":`, `"i\u0064":`},
		{`"schema_version":1`, `"schema_version":0,"schema_version":1`},
		{`"score_grant":4`, `"score_grant":0,"score_grant":4`},
	} {
		candidate := bytes.Replace(data, []byte(test.before), []byte(test.after), 1)
		if bytes.Equal(data, candidate) {
			t.Fatal("control did not change source")
		}
		if _, err := LoadCatalog(candidate, testRegistry()); err != nil {
			t.Fatal(err)
		}
		if err := publicapi.ValidateJSON(AchievementCatalogSchemaName, candidate, achievementSchemaDefinitions(t)); err != nil {
			t.Fatal(err)
		}
	}
}

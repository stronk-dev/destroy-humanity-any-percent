package meters

import (
	"bytes"
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"cloud-clicker/server/multiplier"
	"cloud-clicker/server/publicapi"
)

func meterSchemaDefinitions(t *testing.T) map[string]*publicapi.Schema {
	t.Helper()
	definitions, err := publicapi.ValidateSchemaDefinitions(MeterSchemas())
	if err != nil {
		t.Fatal(err)
	}
	return definitions
}

func TestMeterOwnerSchemaAdmitsExistingArtifactsWithoutChangingBytes(t *testing.T) {
	definitions := meterSchemaDefinitions(t)
	paths := []string{
		"../../balance/meters/first-content.json",
		"../../testdata/harness/content-dynamics/bundles/sha256:6c7fab29c24fae68e3067c883177bc78fe61b9d91704b6d936b3e4f3cfd8f789/artifacts/meters.json",
		"../../testdata/harness/content-dynamics/bundles/sha256:baa890501b2864d14cc0238d633a562cb8c6fca406190487831e0c447af128f6/artifacts/meters.json",
	}
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		before := bytes.Clone(data)
		if err := publicapi.ValidateJSON(MeterCatalogSchemaName, data, definitions); err != nil {
			t.Fatalf("%s: %v", path, err)
		}
		if _, err := LoadCatalog(data); err != nil {
			t.Fatalf("%s: %v", path, err)
		}
		if !bytes.Equal(before, data) {
			t.Fatal("validation changed artifact bytes")
		}
	}
	if err := publicapi.ValidateJSON(MeterCatalogSchemaName, validCatalogBytes(t), definitions); err != nil {
		t.Fatal(err)
	}
}

func TestMeterOwnerSchemaClosedShapes(t *testing.T) {
	type object = map[string]any
	first := func(root object) object { return root["meters"].([]any)[0].(object) }
	reseed := func(root object) object { return root["trust_reseed"].(object) }
	input := func(root object) object { return first(root)["inputs"].([]any)[0].(object) }
	contribution := func(root object) object { return first(root)["inputs"].([]any)[1].(object) }
	band := func(root object) object { return first(root)["bands"].([]any)[0].(object) }
	decay := func(root object) object { return first(root)["decay"].(object) }
	definitions := meterSchemaDefinitions(t)
	for _, test := range []struct {
		name string
		edit func(object)
	}{
		{"root missing", func(v object) { delete(v, "schema_version") }},
		{"root private", func(v object) { v["founder_id"] = "private" }},
		{"version", func(v object) { v["schema_version"] = 2 }},
		{"null meters", func(v object) { v["meters"] = nil }},
		{"reseed missing", func(v object) { delete(reseed(v), "floor_value") }},
		{"reseed private", func(v object) { reseed(v)["account_id"] = "private" }},
		{"reseed fractional", func(v object) { reseed(v)["base_value"] = .5 }},
		{"reseed denominator zero", func(v object) { reseed(v)["notoriety_denominator"] = 0 }},
		{"reseed unsafe", func(v object) { reseed(v)["notoriety_numerator"] = json.Number("9007199254740992") }},
		{"meter missing", func(v object) { delete(first(v), "decay") }},
		{"meter private", func(v object) { first(v)["account_id"] = "private" }},
		{"meter unregistered", func(v object) { first(v)["id"] = "externality.total" }},
		{"meter scope", func(v object) { first(v)["scope"] = "founder" }},
		{"meter minimum", func(v object) { first(v)["min_value"] = 1 }},
		{"meter maximum", func(v object) { first(v)["max_value"] = 99 }},
		{"initial below", func(v object) { first(v)["initial_value"] = -1 }},
		{"initial above", func(v object) { first(v)["initial_value"] = 101 }},
		{"band missing", func(v object) { delete(band(v), "floor_value") }},
		{"band open", func(v object) { band(v)["script"] = "arbitrary" }},
		{"band ID", func(v object) { band(v)["id"] = "bad-id" }},
		{"band floor", func(v object) { band(v)["floor_value"] = 101 }},
		{"input null", func(v object) { first(v)["inputs"] = nil }},
		{"input arm", func(v object) { input(v)["kind"] = "script" }},
		{"input mixed", func(v object) { input(v)["slot"] = "upgrades" }},
		{"input missing", func(v object) { delete(input(v), "delta") }},
		{"input zero", func(v object) { input(v)["delta"] = 0 }},
		{"input below", func(v object) { input(v)["delta"] = -101 }},
		{"input above", func(v object) { input(v)["delta"] = 101 }},
		{"input fact ID", func(v object) { input(v)["fact_kind"] = "bad-id" }},
		{"slot unknown", func(v object) { contribution(v)["slot"] = "fictional" }},
		{"slot zero", func(v object) { contribution(v)["delta_per_attended_hour"] = 0 }},
		{"slot ID", func(v object) { contribution(v)["source_id"] = "bad-id" }},
		{"decay missing", func(v object) { delete(decay(v), "toward_value") }},
		{"decay private", func(v object) { decay(v)["founder_id"] = "private" }},
		{"decay zero", func(v object) { decay(v)["rate_per_attended_hour"] = 0 }},
		{"decay bound", func(v object) { decay(v)["toward_value"] = 101 }},
	} {
		t.Run(test.name, func(t *testing.T) {
			data := mutateCatalog(t, test.edit)
			if err := publicapi.ValidateJSON(MeterCatalogSchemaName, data, definitions); err == nil {
				t.Fatal("descriptor admitted invalid shape")
			}
			if _, err := LoadCatalog(data); err == nil {
				t.Fatal("loader admitted invalid shape")
			}
		})
	}
}

func TestMeterOwnerSchemaInputAndDecayDomains(t *testing.T) {
	definitions := meterSchemaDefinitions(t)
	for _, slot := range multiplier.Order {
		for _, delta := range []int{-100, -1, 1, 100} {
			for _, noDecay := range []bool{false, true} {
				data := mutateCatalog(t, func(root map[string]any) {
					row := root["meters"].([]any)[0].(map[string]any)
					inputs := row["inputs"].([]any)
					inputs[0].(map[string]any)["delta"] = delta
					inputs[1].(map[string]any)["slot"] = string(slot)
					inputs[1].(map[string]any)["delta_per_attended_hour"] = delta
					if noDecay {
						row["decay"] = nil
					}
				})
				if err := publicapi.ValidateJSON(MeterCatalogSchemaName, data, definitions); err != nil {
					t.Fatalf("slot=%s delta=%d null-decay=%t: %v", slot, delta, noDecay, err)
				}
				if _, err := LoadCatalog(data); err != nil {
					t.Fatal(err)
				}
			}
		}
	}
}

func TestMeterOwnerSchemaDoesNotReplaceSemanticLoader(t *testing.T) {
	definitions := meterSchemaDefinitions(t)
	for name, edit := range map[string]func(map[string]any){
		"missing axis": func(v map[string]any) { v["meters"] = v["meters"].([]any)[:10] },
		"unsorted bands": func(v map[string]any) {
			v["meters"].([]any)[0].(map[string]any)["bands"].([]any)[1].(map[string]any)["floor_value"] = 0
		},
		"reseed relationship": func(v map[string]any) { v["trust_reseed"].(map[string]any)["base_value"] = 0 },
		"duplicate source": func(v map[string]any) {
			row := v["meters"].([]any)[0].(map[string]any)
			row["inputs"] = append(row["inputs"].([]any), row["inputs"].([]any)[0])
		},
	} {
		t.Run(name, func(t *testing.T) {
			data := mutateCatalog(t, edit)
			if err := publicapi.ValidateJSON(MeterCatalogSchemaName, data, definitions); err != nil {
				t.Fatal(err)
			}
			if _, err := LoadCatalog(data); err == nil {
				t.Fatal("structural schema replaced semantic loader")
			}
		})
	}
}

func TestMeterOwnerSchemaFreshExportsAndExistingMemberSemantics(t *testing.T) {
	before := MeterSchemas()
	fresh := MeterSchemas()
	if !reflect.DeepEqual(before, fresh) {
		t.Fatal("non-deterministic exports")
	}
	fresh[0].Schema.Fields[0].Schema.Items.Fields[0].Schema.Items.Fields[0].Name = "changed"
	if !reflect.DeepEqual(before, MeterSchemas()) {
		t.Fatal("caller mutation persisted into owner authority")
	}
	definitions := meterSchemaDefinitions(t)
	source := validCatalogBytes(t)
	for _, data := range [][]byte{
		bytes.Replace(source, []byte(`"schema_version"`), []byte(`"schema_versi\u006fn"`), 1),
		bytes.Replace(source, []byte(`"schema_version":1`), []byte(`"schema_version":2,"schema_version":1`), 1),
	} {
		if bytes.Equal(source, data) {
			t.Fatal("control did not change source")
		}
		if err := publicapi.ValidateJSON(MeterCatalogSchemaName, data, definitions); err != nil {
			t.Fatal(err)
		}
		if _, err := LoadCatalog(data); err != nil {
			t.Fatal(err)
		}
	}
	// Do not globally reject duplicate names, but the final selected object must
	// still contain its complete exact shape, as in the TypeScript reader.
	data := append(bytes.Clone(source[:len(source)-1]), []byte(`,"trust_reseed":{"base_value":90}}`)...)
	if err := publicapi.ValidateJSON(MeterCatalogSchemaName, data, definitions); err == nil {
		t.Fatal("incomplete final reseed object admitted by descriptor")
	}
	if _, err := LoadCatalog(data); err == nil {
		t.Fatal("incomplete final reseed object merged with an earlier object")
	}
}

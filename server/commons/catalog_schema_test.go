package commons

import (
	"bytes"
	"os"
	"reflect"
	"testing"

	"cloud-clicker/server/multiplier"
	"cloud-clicker/server/publicapi"
)

func commonsSchemaDefinitions(t *testing.T) map[string]*publicapi.Schema {
	t.Helper()
	definitions, err := publicapi.ValidateSchemaDefinitions(CommonsSchemas())
	if err != nil {
		t.Fatal(err)
	}
	return definitions
}

func TestCommonsOwnerSchemaAdmitsStoredCatalogsAndEveryRegisteredNonCommonsSlot(t *testing.T) {
	definitions := commonsSchemaDefinitions(t)
	for _, path := range []string{
		"../../balance/commons/phase0.json",
		"../../testdata/harness/content-dynamics/bundles/sha256:6c7fab29c24fae68e3067c883177bc78fe61b9d91704b6d936b3e4f3cfd8f789/artifacts/commons.json",
		"../../testdata/harness/content-dynamics/bundles/sha256:baa890501b2864d14cc0238d633a562cb8c6fca406190487831e0c447af128f6/artifacts/commons.json",
	} {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		before := bytes.Clone(data)
		if err := publicapi.ValidateJSON(CommonsCatalogSchemaName, data, definitions); err != nil {
			t.Fatal(err)
		}
		if _, err := LoadCatalog(data); err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(before, data) {
			t.Fatal("validation changed artifact bytes")
		}
	}
	for _, slot := range multiplier.Order {
		if slot == multiplier.SlotCommons {
			continue
		}
		for _, forsworn := range []bool{false, true} {
			data := editCommonsFixture(t, func(v map[string]any) {
				v["source_weights"] = []any{map[string]any{"source_id": "source.example", "slot": string(slot), "weight_ppm": 1000000, "forsworn": forsworn}}
			})
			if err := publicapi.ValidateJSON(CommonsCatalogSchemaName, data, definitions); err != nil {
				t.Fatal(err)
			}
			if _, err := LoadCatalog(data); err != nil {
				t.Fatal(err)
			}
		}
	}
	data := editCommonsFixture(t, func(v map[string]any) {
		v["collective_exponent_ppm"], v["maximum_bonus"], v["population_tolerance_ppm"] = 2000000, "2e0", 0
	})
	if err := publicapi.ValidateJSON(CommonsCatalogSchemaName, data, definitions); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadCatalog(data); err != nil {
		t.Fatal(err)
	}
}

func TestCommonsOwnerSchemaClosesPolicyAndSourceShapes(t *testing.T) {
	definitions := commonsSchemaDefinitions(t)
	for name, edit := range map[string]func(map[string]any){
		"root alias":         func(v map[string]any) { v["SCHEMA_VERSION"] = v["schema_version"]; delete(v, "schema_version") },
		"private root":       func(v map[string]any) { v["member_ids"] = []any{} },
		"version":            func(v map[string]any) { v["schema_version"] = 2 },
		"missing zero field": func(v map[string]any) { delete(v, "population_tolerance_ppm") },
		"null zero field":    func(v map[string]any) { v["npc_weight_ppm"] = nil },
		"ratio maximum":      func(v map[string]any) { v["default_tithe_ppm"] = 1000001 },
		"exponent minimum":   func(v map[string]any) { v["collective_exponent_ppm"] = 999999 },
		"exponent maximum":   func(v map[string]any) { v["collective_exponent_ppm"] = 10000001 },
		"bonus grammar":      func(v map[string]any) { v["maximum_bonus"] = "5" },
		"window zero":        func(v map[string]any) { v["solidarity_window_ms"] = 0 },
		"unsafe count":       func(v map[string]any) { v["cohort_target_size"] = int64(9007199254740992) },
		"null row":           func(v map[string]any) { v["source_weights"] = []any{nil} },
		"recursive slot": func(v map[string]any) {
			v["source_weights"] = []any{map[string]any{"source_id": "source.example", "slot": "commons", "weight_ppm": 1, "forsworn": false}}
		},
		"missing boolean": func(v map[string]any) {
			v["source_weights"] = []any{map[string]any{"source_id": "source.example", "slot": "upgrades", "weight_ppm": 1}}
		},
		"zero weight": func(v map[string]any) {
			v["source_weights"] = []any{map[string]any{"source_id": "source.example", "slot": "upgrades", "weight_ppm": 0, "forsworn": false}}
		},
		"extra row key": func(v map[string]any) {
			v["source_weights"] = []any{map[string]any{"source_id": "source.example", "slot": "upgrades", "weight_ppm": 1, "forsworn": false, "label": "private"}}
		},
	} {
		t.Run(name, func(t *testing.T) {
			data := editCommonsFixture(t, edit)
			if err := publicapi.ValidateJSON(CommonsCatalogSchemaName, data, definitions); err == nil {
				t.Fatal("schema admitted malformed policy")
			}
			if _, err := LoadCatalog(data); err == nil {
				t.Fatal("owner admitted malformed policy")
			}
		})
	}
}

func TestCommonsOwnerSchemaRetainsRelationshipsAndFreshAuthority(t *testing.T) {
	definitions := commonsSchemaDefinitions(t)
	for name, edit := range map[string]func(map[string]any){
		"tithe order":       func(v map[string]any) { v["minimum_tithe_ppm"] = 150001 },
		"Health blend":      func(v map[string]any) { v["guild_health_weight_ppm"] = 0 },
		"threshold order":   func(v map[string]any) { v["healthy_health_ppm"] = 1 },
		"decay/recovery":    func(v map[string]any) { v["health_recovery_ppm_per_hour"] = 1 },
		"merge floor":       func(v map[string]any) { v["cohort_merge_floor"] = 151 },
		"NPC floor":         func(v map[string]any) { v["npc_population_floor"] = 1 },
		"nonpositive bonus": func(v map[string]any) { v["maximum_bonus"] = "0" },
		"duplicate source": func(v map[string]any) {
			row := map[string]any{"source_id": "source.example", "slot": "upgrades", "weight_ppm": 1, "forsworn": false}
			v["source_weights"] = []any{row, row}
		},
	} {
		t.Run(name, func(t *testing.T) {
			data := editCommonsFixture(t, edit)
			if err := publicapi.ValidateJSON(CommonsCatalogSchemaName, data, definitions); err != nil {
				t.Fatal(err)
			}
			if _, err := LoadCatalog(data); err == nil {
				t.Fatal("schema replaced relational loader checks")
			}
		})
	}
	data, err := os.ReadFile("../../balance/commons/phase0.json")
	if err != nil {
		t.Fatal(err)
	}
	trimmed := bytes.TrimSpace(data)
	for _, candidate := range [][]byte{
		bytes.Replace(data, []byte(`"schema_version"`), []byte(`"schema_versi\u006fn"`), 1),
		append(bytes.Clone(trimmed[:len(trimmed)-1]), []byte(`,"schema_version":2,"schema_version":1}`)...),
	} {
		if bytes.Equal(candidate, data) {
			t.Fatal("control did not change bytes")
		}
		if err := publicapi.ValidateJSON(CommonsCatalogSchemaName, candidate, definitions); err != nil {
			t.Fatal(err)
		}
		if _, err := LoadCatalog(candidate); err != nil {
			t.Fatal(err)
		}
	}
	before, fresh := CommonsSchemas(), CommonsSchemas()
	if !reflect.DeepEqual(before, fresh) {
		t.Fatal("nondeterministic exports")
	}
	for _, field := range fresh[0].Schema.Fields {
		if field.Name == "solidarity_window_ms" {
			*field.Schema.Maximum = 0
		}
	}
	if !reflect.DeepEqual(before, CommonsSchemas()) {
		t.Fatal("caller changed schema authority")
	}
}

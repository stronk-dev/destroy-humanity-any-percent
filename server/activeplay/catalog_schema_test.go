package activeplay

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"cloud-clicker/server/economy"
	"cloud-clicker/server/publicapi"
)

func activePlaySchemaDefinitions(t *testing.T) map[string]*publicapi.Schema {
	t.Helper()
	definitions, err := publicapi.ValidateSchemaDefinitions(ActivePlaySchemas())
	if err != nil {
		t.Fatal(err)
	}
	return definitions
}

func TestActivePlayOwnerSchemaAdmitsCurrentStoredAndAllEffectArms(t *testing.T) {
	definitions := activePlaySchemaDefinitions(t)
	paths := []string{
		"../../balance/opportunities/t0-t1.json",
		"../../testdata/harness/content-dynamics/bundles/sha256:6c7fab29c24fae68e3067c883177bc78fe61b9d91704b6d936b3e4f3cfd8f789/artifacts/opportunities.json",
		"../../testdata/harness/content-dynamics/bundles/sha256:baa890501b2864d14cc0238d633a562cb8c6fca406190487831e0c447af128f6/artifacts/opportunities.json",
	}
	for index, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		before := bytes.Clone(data)
		if err := publicapi.ValidateJSON(ActivePlayCatalogSchemaName, data, definitions); err != nil {
			t.Fatalf("%s: %v", path, err)
		}
		economyCatalog := activeEconomy(t)
		if index > 0 {
			economyBytes, err := os.ReadFile(filepath.Join(filepath.Dir(path), "economy.json"))
			if err != nil {
				t.Fatal(err)
			}
			economyCatalog, err = economy.LoadCatalog(economyBytes)
			if err != nil {
				t.Fatal(err)
			}
		}
		if _, err := LoadCatalog(data, economyCatalog); err != nil {
			t.Fatalf("%s: %v", path, err)
		}
		if !bytes.Equal(before, data) {
			t.Fatal("validation changed artifact bytes")
		}
	}
	if err := publicapi.ValidateJSON(ActivePlayCatalogSchemaName, activeFixture(t), definitions); err != nil {
		t.Fatal(err)
	}
	// Every duration arm accepts both boundaries of its existing exact domain.
	for _, duration := range []json.Number{"1", "9007199254740991"} {
		data := editActiveFixture(t, func(root map[string]any) {
			for _, index := range []int{0, 1, 3} {
				root["effects"].([]any)[index].(map[string]any)["duration_ms"] = duration
			}
		})
		if err := publicapi.ValidateJSON(ActivePlayCatalogSchemaName, data, definitions); err != nil {
			t.Fatal(err)
		}
		if _, err := LoadCatalog(data, activeEconomy(t)); err != nil {
			t.Fatal(err)
		}
	}
}

func TestActivePlayOwnerSchemaClosedBoundsAndFormats(t *testing.T) {
	type object = map[string]any
	row := func(root object, index int) object { return root["effects"].([]any)[index].(object) }
	definitions := activePlaySchemaDefinitions(t)
	for name, edit := range map[string]func(object){
		"root private":       func(v object) { v["founder_id"] = "private" },
		"version":            func(v object) { v["schema_version"] = 2 },
		"missing combo":      func(v object) { delete(v, "combo_policy") },
		"null effects":       func(v object) { v["effects"] = nil },
		"sampler":            func(v object) { v["schedule_policy"].(object)["sampler_version"] = "fictional.v1" },
		"substream":          func(v object) { v["schedule_policy"].(object)["substream_label"] = "fictional.v1" },
		"schedule zero":      func(v object) { v["schedule_policy"].(object)["scale_ms"] = 0 },
		"combo Decimal":      func(v object) { v["combo_policy"].(object)["combo_cap"] = 7 },
		"combo ID":           func(v object) { v["combo_policy"].(object)["hardcap_reason_key"] = "bad-id" },
		"kind":               func(v object) { row(v, 0)["kind"] = "instant_payout" },
		"ID":                 func(v object) { row(v, 0)["effect_row_id"] = "bad-id" },
		"weight":             func(v object) { row(v, 0)["weight"] = 0 },
		"fractional weight":  func(v object) { row(v, 0)["weight"] = .5 },
		"ppm":                func(v object) { row(v, 0)["per_owned_ppm"] = 1_000_001 },
		"generator ID":       func(v object) { row(v, 0)["eligible_generator_ids"] = []string{"bad-id"} },
		"action ID":          func(v object) { row(v, 1)["action_ids"] = []string{"bad-id"} },
		"factor canonical":   func(v object) { row(v, 1)["factor"] = "7" },
		"lucky missing":      func(v object) { delete(row(v, 2), "epsilon") },
		"lucky canonical":    func(v object) { row(v, 2)["lucky_bank_frac"] = "0.15" },
		"lucky resource":     func(v object) { row(v, 2)["resource_id"] = "bad-id" },
		"production targets": func(v object) { row(v, 3)["targets"] = []string{"manual_output"} },
	} {
		t.Run(name, func(t *testing.T) {
			data := editActiveFixture(t, edit)
			if err := publicapi.ValidateJSON(ActivePlayCatalogSchemaName, data, definitions); err == nil {
				t.Fatal("descriptor admitted invalid shape")
			}
			if _, err := LoadCatalog(data, activeEconomy(t)); err == nil {
				t.Fatal("loader admitted invalid shape")
			}
		})
	}
}

func TestActivePlayOwnerSchemaDoesNotReplaceSemanticLoader(t *testing.T) {
	definitions := activePlaySchemaDefinitions(t)
	for name, edit := range map[string]func(map[string]any){
		"unbound action": func(v map[string]any) {
			v["effects"].([]any)[1].(map[string]any)["action_ids"] = []string{"manual.absent"}
		},
		"factor relationship": func(v map[string]any) { v["effects"].([]any)[1].(map[string]any)["factor"] = "1e0" },
		"empty effects":       func(v map[string]any) { v["effects"] = []any{} },
		"unordered effects":   func(v map[string]any) { rows := v["effects"].([]any); rows[0], rows[1] = rows[1], rows[0] },
	} {
		t.Run(name, func(t *testing.T) {
			data := editActiveFixture(t, edit)
			if err := publicapi.ValidateJSON(ActivePlayCatalogSchemaName, data, definitions); err != nil {
				t.Fatal(err)
			}
			if _, err := LoadCatalog(data, activeEconomy(t)); err == nil {
				t.Fatal("structural descriptor replaced semantic owner")
			}
		})
	}
}

func TestActivePlayOwnerSchemaFreshAndExistingMemberSemantics(t *testing.T) {
	before := ActivePlaySchemas()
	fresh := ActivePlaySchemas()
	if !reflect.DeepEqual(before, fresh) {
		t.Fatal("non-deterministic export")
	}
	fresh[0].Schema.Fields[0].Schema.Fields[0].Schema.Format = "mechanical-id"
	if !reflect.DeepEqual(before, ActivePlaySchemas()) {
		t.Fatal("caller mutation changed owner authority")
	}
	definitions := activePlaySchemaDefinitions(t)
	source := activeFixture(t)
	for _, data := range [][]byte{
		bytes.Replace(source, []byte(`"schema_version"`), []byte(`"schema_versi\u006fn"`), 1),
		append(bytes.Clone(source[:len(source)-1]), []byte(`,"schema_version":2,"schema_version":1}`)...),
	} {
		if bytes.Equal(source, data) {
			t.Fatal("member control did not change source")
		}
		if err := publicapi.ValidateJSON(ActivePlayCatalogSchemaName, data, definitions); err != nil {
			t.Fatal(err)
		}
		if _, err := LoadCatalog(data, activeEconomy(t)); err != nil {
			t.Fatal(err)
		}
	}
}

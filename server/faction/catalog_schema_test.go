package faction

import (
	"bytes"
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"cloud-clicker/server/commons"
	"cloud-clicker/server/publicapi"
)

func factionSchemaDefinitions(t *testing.T) map[string]*publicapi.Schema {
	t.Helper()
	definitions, err := publicapi.ValidateSchemaDefinitions(FactionSchemas())
	if err != nil {
		t.Fatal(err)
	}
	return definitions
}

func TestFactionOwnerSchemaAdmitsCurrentStoredAndRetunedCatalogs(t *testing.T) {
	definitions := factionSchemaDefinitions(t)
	for _, prefix := range []string{
		"../../testdata/harness/content-dynamics/bundles/sha256:6c7fab29c24fae68e3067c883177bc78fe61b9d91704b6d936b3e4f3cfd8f789/artifacts/",
		"../../testdata/harness/content-dynamics/bundles/sha256:baa890501b2864d14cc0238d633a562cb8c6fca406190487831e0c447af128f6/artifacts/",
	} {
		data, err := os.ReadFile(prefix + "factions.json")
		if err != nil {
			t.Fatal(err)
		}
		policy, err := os.ReadFile(prefix + "commons.json")
		if err != nil {
			t.Fatal(err)
		}
		compact, err := commons.LoadCatalog(policy)
		if err != nil {
			t.Fatal(err)
		}
		before := bytes.Clone(data)
		if err := publicapi.ValidateJSON(FactionCatalogSchemaName, data, definitions); err != nil {
			t.Fatal(err)
		}
		if _, err := LoadCatalog(data, CompactTitheBand{compact.MinimumTithePPM, compact.DefaultTithePPM, compact.MaximumTithePPM}); err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(before, data) {
			t.Fatal("validation rewrote artifact bytes")
		}
	}
	data, band := factionFixture(t)
	for _, candidate := range [][]byte{data, editFactionJSON(t, data, func(v map[string]any) {
		v["stock_cap"] = json.Number("9007199254740991")
		v["stock_interval_ms"] = 86400000
		v["factions"].([]any)[2].(map[string]any)["compact"].(map[string]any)["tithe_ppm"] = band.MaximumPPM
	})} {
		if err := publicapi.ValidateJSON(FactionCatalogSchemaName, candidate, definitions); err != nil {
			t.Fatal(err)
		}
		if _, err := LoadCatalog(candidate, band); err != nil {
			t.Fatal(err)
		}
	}
}

func TestFactionOwnerSchemaClosesRootRowsAndNullableCompact(t *testing.T) {
	definitions := factionSchemaDefinitions(t)
	data, band := factionFixture(t)
	row := func(v map[string]any) map[string]any { return v["factions"].([]any)[0].(map[string]any) }
	compact := func(v map[string]any) map[string]any {
		return v["factions"].([]any)[2].(map[string]any)["compact"].(map[string]any)
	}
	for name, edit := range map[string]func(map[string]any){
		"missing root":             func(v map[string]any) { delete(v, "stock_cap") },
		"root alias":               func(v map[string]any) { v["STOCK_CAP"] = v["stock_cap"]; delete(v, "stock_cap") },
		"private root":             func(v map[string]any) { v["owner_id"] = "private" },
		"version":                  func(v map[string]any) { v["schema_version"] = 2 },
		"stock zero":               func(v map[string]any) { v["stock_cap"] = 0 },
		"unsafe stock":             func(v map[string]any) { v["stock_cap"] = json.Number("9007199254740992") },
		"interval":                 func(v map[string]any) { v["stock_interval_ms"] = 86400001 },
		"null row":                 func(v map[string]any) { v["factions"].([]any)[0] = nil },
		"unknown faction":          func(v map[string]any) { row(v)["id"] = "future" },
		"unknown resource":         func(v map[string]any) { row(v)["produces"] = "future" },
		"missing nullable compact": func(v map[string]any) { delete(row(v), "compact") },
		"wrong compact type":       func(v map[string]any) { row(v)["compact"] = false },
		"missing auto sign":        func(v map[string]any) { delete(compact(v), "auto_sign") },
		"compact alias": func(v map[string]any) {
			compact(v)["TITHE_PPM"] = compact(v)["tithe_ppm"]
			delete(compact(v), "tithe_ppm")
		},
		"null tithe":         func(v map[string]any) { compact(v)["tithe_ppm"] = nil },
		"tithe wrong type":   func(v map[string]any) { compact(v)["tithe_ppm"] = "130000" },
		"null modifier list": func(v map[string]any) { row(v)["modifier_slots"] = nil },
		"copy grammar":       func(v map[string]any) { row(v)["incorporation_copy_key"] = "incorporate.bad-id" },
	} {
		t.Run(name, func(t *testing.T) {
			candidate := editFactionJSON(t, data, edit)
			if err := publicapi.ValidateJSON(FactionCatalogSchemaName, candidate, definitions); err == nil {
				t.Fatal("descriptor admitted malformed catalog")
			}
			if _, err := LoadCatalog(candidate, band); err == nil {
				t.Fatal("owner admitted malformed catalog")
			}
		})
	}
}

func TestFactionOwnerSchemaRetainsCycleBindingAndEmptyModifierRules(t *testing.T) {
	definitions := factionSchemaDefinitions(t)
	data, band := factionFixture(t)
	for name, edit := range map[string]func(map[string]any){
		"self cycle": func(v map[string]any) {
			row := v["factions"].([]any)[0].(map[string]any)
			row["consumes"] = row["produces"]
		},
		"two cycles": func(v map[string]any) {
			rows := v["factions"].([]any)
			rows[0].(map[string]any)["consumes"] = "hype"
			rows[2].(map[string]any)["consumes"] = "compliance"
		},
		"missing faction": func(v map[string]any) { v["factions"] = v["factions"].([]any)[:3] },
		"copy binding": func(v map[string]any) {
			v["factions"].([]any)[0].(map[string]any)["incorporation_copy_key"] = "incorporate.enterprise"
		},
		"Open Source compact null": func(v map[string]any) { v["factions"].([]any)[2].(map[string]any)["compact"] = nil },
		"ordinary compact": func(v map[string]any) {
			v["factions"].([]any)[0].(map[string]any)["compact"] = map[string]any{"auto_sign": true, "tithe_ppm": band.MaximumPPM}
		},
		"auto sign false": func(v map[string]any) {
			v["factions"].([]any)[2].(map[string]any)["compact"].(map[string]any)["auto_sign"] = false
		},
		"below extra tithe": func(v map[string]any) {
			v["factions"].([]any)[2].(map[string]any)["compact"].(map[string]any)["tithe_ppm"] = band.DefaultPPM
		},
		"above declared tithe": func(v map[string]any) {
			v["factions"].([]any)[2].(map[string]any)["compact"].(map[string]any)["tithe_ppm"] = band.MaximumPPM + 1
		},
		"reserved modifier": func(v map[string]any) {
			v["factions"].([]any)[0].(map[string]any)["modifier_slots"] = []any{map[string]any{"slot": "faction", "ppm": 1000000}}
		},
	} {
		t.Run(name, func(t *testing.T) {
			candidate := editFactionJSON(t, data, edit)
			if err := publicapi.ValidateJSON(FactionCatalogSchemaName, candidate, definitions); err != nil {
				t.Fatal(err)
			}
			if _, err := LoadCatalog(candidate, band); err == nil {
				t.Fatal("descriptor replaced owner/composition checks")
			}
		})
	}
	trimmed := bytes.TrimSpace(data)
	for _, candidate := range [][]byte{
		bytes.Replace(data, []byte(`"stock_cap"`), []byte(`"stock_\u0063ap"`), 1),
		append(bytes.Clone(trimmed[:len(trimmed)-1]), []byte(`,"schema_version":2,"schema_version":1}`)...),
	} {
		if bytes.Equal(data, candidate) {
			t.Fatal("member control did not change bytes")
		}
		if err := publicapi.ValidateJSON(FactionCatalogSchemaName, candidate, definitions); err != nil {
			t.Fatal(err)
		}
		if _, err := LoadCatalog(candidate, band); err != nil {
			t.Fatal(err)
		}
	}
	before, fresh := FactionSchemas(), FactionSchemas()
	if !reflect.DeepEqual(before, fresh) {
		t.Fatal("nondeterministic exports")
	}
	for _, field := range fresh[0].Schema.Fields {
		if field.Name == "stock_cap" {
			*field.Schema.Maximum = 0
		}
		if field.Name == "factions" {
			for _, rowField := range field.Schema.Items.Fields {
				if rowField.Name == "id" {
					rowField.Schema.Enum[0] = "future"
				}
			}
		}
	}
	if !reflect.DeepEqual(before, FactionSchemas()) {
		t.Fatal("caller mutated schema authority")
	}
}

package fiscal

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"cloud-clicker/server/economy"
	"cloud-clicker/server/multiplier"
	"cloud-clicker/server/publicapi"
)

func fiscalSchemaDefinitions(t *testing.T) map[string]*publicapi.Schema {
	t.Helper()
	definitions, err := publicapi.ValidateSchemaDefinitions(FiscalSchemas())
	if err != nil {
		t.Fatal(err)
	}
	return definitions
}

func TestFiscalOwnerSchemaAdmitsCurrentStoredAndFixtureWithoutChangingBytes(t *testing.T) {
	definitions := fiscalSchemaDefinitions(t)
	paths := []string{
		"../../balance/fiscal/first-content.json",
		"../../testdata/harness/content-dynamics/bundles/sha256:6c7fab29c24fae68e3067c883177bc78fe61b9d91704b6d936b3e4f3cfd8f789/artifacts/fiscal.json",
		"../../testdata/harness/content-dynamics/bundles/sha256:baa890501b2864d14cc0238d633a562cb8c6fca406190487831e0c447af128f6/artifacts/fiscal.json",
	}
	for index, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		before := bytes.Clone(data)
		if err := publicapi.ValidateJSON(FiscalCatalogSchemaName, data, definitions); err != nil {
			t.Fatalf("%s: %v", path, err)
		}
		economyCatalog := fiscalEconomy(t)
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
			t.Fatal("validation changed original artifact bytes")
		}
	}
	if err := publicapi.ValidateJSON(FiscalCatalogSchemaName, loadSharedFixture(t).Baseline, definitions); err != nil {
		t.Fatal(err)
	}
}

func TestFiscalOwnerSchemaClosedShapesBoundsAndSlotDomain(t *testing.T) {
	type object = map[string]any
	clock := func(v object) object { return v["clock_policy"].(object) }
	credit := func(v object) object { return v["credit_policy"].(object) }
	hoard := func(v object) object { return v["hoard_policy"].(object) }
	level := func(v object) object { return v["generator_level_rows"].([]any)[0].(object) }
	unlock := func(v object) object { return v["unlock_rows"].([]any)[0].(object) }
	definitions := fiscalSchemaDefinitions(t)
	for name, edit := range map[string]func(object){
		"root missing":       func(v object) { delete(v, "clock_policy") },
		"root private":       func(v object) { v["account_id"] = "private" },
		"version":            func(v object) { v["schema_version"] = 2 },
		"null policy":        func(v object) { v["clock_policy"] = nil },
		"clock alias":        func(v object) { clock(v)["EARLY_MS"] = clock(v)["early_ms"]; delete(clock(v), "early_ms") },
		"clock zero":         func(v object) { clock(v)["auto_ms"] = 0 },
		"clock unsafe":       func(v object) { clock(v)["auto_ms"] = json.Number("9007199254740992") },
		"clock fractional":   func(v object) { clock(v)["guaranteed_ms"] = 1.5 },
		"probability below":  func(v object) { clock(v)["early_success_ppm"] = -1 },
		"probability above":  func(v object) { clock(v)["early_success_ppm"] = PPM + 1 },
		"credit missing":     func(v object) { delete(credit(v), "hardcap") },
		"credit zero":        func(v object) { credit(v)["credit_per_period"] = 0 },
		"credit ID":          func(v object) { credit(v)["hardcap_reason_key"] = "bad-id" },
		"credit null reason": func(v object) { credit(v)["hardcap_reason_key"] = nil },
		"hoard private":      func(v object) { hoard(v)["founder_id"] = "private" },
		"hoard ppm":          func(v object) { hoard(v)["ppm_per_credit"] = 0 },
		"hoard cap":          func(v object) { hoard(v)["cap_credits"] = 0 },
		"hoard target":       func(v object) { hoard(v)["target"] = "manual" },
		"hoard slot":         func(v object) { hoard(v)["slot"] = "fictional" },
		"hoard source":       func(v object) { hoard(v)["source_id"] = "bad-id" },
		"level null":         func(v object) { v["generator_level_rows"] = nil },
		"level missing":      func(v object) { delete(level(v), "slot") },
		"level ppm":          func(v object) { level(v)["ppm_per_level"] = PPM + 1 },
		"level cap":          func(v object) { level(v)["level_hardcap"] = json.Number("9007199254740992") },
		"level ID":           func(v object) { level(v)["generator_id"] = "bad-id" },
		"unlock null":        func(v object) { v["unlock_rows"] = nil },
		"unlock cost":        func(v object) { unlock(v)["cost"] = 0 },
		"unlock ID":          func(v object) { unlock(v)["unlock_id"] = "bad-id" },
	} {
		t.Run(name, func(t *testing.T) {
			data := editFiscalFixture(t, edit)
			if err := publicapi.ValidateJSON(FiscalCatalogSchemaName, data, definitions); err == nil {
				t.Fatal("descriptor admitted invalid shape")
			}
			if _, err := LoadCatalog(data, fiscalEconomy(t)); err == nil {
				t.Fatal("owner admitted invalid shape")
			}
		})
	}
	for _, registered := range multiplier.Order {
		data := editFiscalFixture(t, func(v object) { hoard(v)["slot"] = string(registered); level(v)["slot"] = string(registered) })
		if err := publicapi.ValidateJSON(FiscalCatalogSchemaName, data, definitions); err != nil {
			t.Fatalf("registered slot %s: %v", registered, err)
		}
		// The descriptor's common slot domain does not authorize a new Economy
		// declaration. Actual current sources still bind specifically to prestige.
		_, err := LoadCatalog(data, fiscalEconomy(t))
		if (err == nil) != (registered == multiplier.SlotPrestige) {
			t.Fatalf("slot %s did not respect pinned declaration: %v", registered, err)
		}
	}
	for _, maximum := range []bool{false, true} {
		data := editFiscalFixture(t, func(v object) {
			var value json.Number = "1"
			var probability int64
			if maximum {
				value, probability = "9007199254740991", PPM
			}
			clock(v)["early_ms"] = 2
			clock(v)["guaranteed_ms"], clock(v)["auto_ms"] = 2, 2
			clock(v)["early_success_ppm"] = probability
			credit(v)["credit_per_period"], credit(v)["hardcap"] = value, value
			hoard(v)["cap_credits"] = value
			level(v)["level_hardcap"], unlock(v)["cost"] = value, value
		})
		if err := publicapi.ValidateJSON(FiscalCatalogSchemaName, data, definitions); err != nil {
			t.Fatal(err)
		}
		if _, err := LoadCatalog(data, fiscalEconomy(t)); err != nil {
			t.Fatal(err)
		}
	}
}

func TestFiscalOwnerSchemaDoesNotReplaceSemanticLoader(t *testing.T) {
	definitions := fiscalSchemaDefinitions(t)
	for name, edit := range map[string]func(map[string]any){
		"clock ordering":      func(v map[string]any) { v["clock_policy"].(map[string]any)["early_ms"] = 299 },
		"credit relationship": func(v map[string]any) { v["credit_policy"].(map[string]any)["credit_per_period"] = 1001 },
		"unknown generator": func(v map[string]any) {
			v["generator_level_rows"].([]any)[0].(map[string]any)["generator_id"] = "generator.absent"
		},
		"missing rows": func(v map[string]any) { v["unlock_rows"] = []any{} },
		"sequence horizon": func(v map[string]any) {
			clock := v["clock_policy"].(map[string]any)
			clock["early_ms"], clock["guaranteed_ms"], clock["auto_ms"] = 1, 1, 1
		},
	} {
		t.Run(name, func(t *testing.T) {
			data := editFiscalFixture(t, edit)
			if err := publicapi.ValidateJSON(FiscalCatalogSchemaName, data, definitions); err != nil {
				t.Fatal(err)
			}
			if _, err := LoadCatalog(data, fiscalEconomy(t)); err == nil {
				t.Fatal("structural descriptor replaced semantic owner")
			}
		})
	}
}

func TestFiscalOwnerSchemaFreshExportsAndOriginalMemberControls(t *testing.T) {
	before, fresh := FiscalSchemas(), FiscalSchemas()
	if !reflect.DeepEqual(before, fresh) {
		t.Fatal("non-deterministic export")
	}
	*fresh[0].Schema.Fields[0].Schema.Fields[0].Schema.Minimum = 0
	if !reflect.DeepEqual(before, FiscalSchemas()) {
		t.Fatal("caller mutation changed owner authority")
	}
	definitions := fiscalSchemaDefinitions(t)
	source := bytes.TrimSpace(loadSharedFixture(t).Baseline)
	var root map[string]any
	if err := json.Unmarshal(source, &root); err != nil {
		t.Fatal(err)
	}
	for _, data := range [][]byte{
		bytes.Replace(source, []byte(`"schema_version"`), []byte(`"schema_versi\u006fn"`), 1),
		fiscalPolicyOverride(t, "credit_policy", root["credit_policy"]),
	} {
		if bytes.Equal(source, data) {
			t.Fatal("member control did not change source")
		}
		if err := publicapi.ValidateJSON(FiscalCatalogSchemaName, data, definitions); err != nil {
			t.Fatal(err)
		}
		if _, err := LoadCatalog(data, fiscalEconomy(t)); err != nil {
			t.Fatal(err)
		}
	}
	if err := publicapi.ValidateJSON(FiscalCatalogSchemaName, fiscalPolicyOverride(t, "credit_policy", nil), definitions); err == nil {
		t.Fatal("descriptor admitted final null policy")
	}
}

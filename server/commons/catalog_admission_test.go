package commons

import (
	"bytes"
	"encoding/json"
	"os"
	"sort"
	"strings"
	"testing"

	"cloud-clicker/server/decimal"
)

func editCommonsFixture(t *testing.T, edit func(map[string]any)) []byte {
	t.Helper()
	data, err := os.ReadFile("../../balance/commons/phase0.json")
	if err != nil {
		t.Fatal(err)
	}
	var root map[string]any
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	if err := decoder.Decode(&root); err != nil {
		t.Fatal(err)
	}
	edit(root)
	result, err := json.Marshal(root)
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func TestCommonsCatalogRequiresLiteralNamesAndExplicitFalse(t *testing.T) {
	data := editCommonsFixture(t, func(v map[string]any) {
		v["source_weights"] = []any{map[string]any{"source_id": "source.example", "slot": "upgrades", "weight_ppm": 1000000, "forsworn": false}}
	})
	var root map[string]any
	if err := json.Unmarshal(data, &root); err != nil {
		t.Fatal(err)
	}
	type target struct {
		path string
		row  map[string]any
	}
	for _, target := range []target{{"root", root}, {"weight", root["source_weights"].([]any)[0].(map[string]any)}} {
		keys := make([]string, 0, len(target.row))
		for key := range target.row {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			t.Run(target.path+"/"+key, func(t *testing.T) {
				value := target.row[key]
				delete(target.row, key)
				target.row[strings.ToUpper(key)] = value
				defer func() { delete(target.row, strings.ToUpper(key)); target.row[key] = value }()
				candidate, err := json.Marshal(root)
				if err != nil {
					t.Fatal(err)
				}
				if _, err := LoadCatalog(candidate); err == nil {
					t.Fatal("case alias admitted")
				}
			})
		}
	}
	for _, variant := range []string{"missing", "null"} {
		t.Run("forsworn/"+variant, func(t *testing.T) {
			candidate := editCommonsFixture(t, func(v map[string]any) {
				row := map[string]any{"source_id": "source.example", "slot": "upgrades", "weight_ppm": 1000000}
				if variant == "null" {
					row["forsworn"] = nil
				}
				v["source_weights"] = []any{row}
			})
			if _, err := LoadCatalog(candidate); err == nil {
				t.Fatal("required forsworn became implicit false")
			}
		})
	}
	if _, err := LoadCatalog(data); err != nil {
		t.Fatal("explicit false rejected", err)
	}
}

func TestCommonsCatalogUsesExactSafeCountAndTimeDomain(t *testing.T) {
	for _, key := range []string{"solidarity_window_ms", "cohort_target_size", "cohort_merge_floor", "npc_population_floor"} {
		for _, maximum := range []bool{true, false} {
			name := key + "/maximum"
			value := decimal.MaxExactInteger
			if !maximum {
				name = key + "/first invalid"
				value++
			}
			t.Run(name, func(t *testing.T) {
				candidate := editCommonsFixture(t, func(v map[string]any) {
					v[key] = value
					if key == "cohort_merge_floor" {
						v["cohort_target_size"], v["npc_population_floor"] = value, value
					}
				})
				_, err := LoadCatalog(candidate)
				if maximum && err != nil {
					t.Fatal("legal representation maximum rejected", err)
				}
				if !maximum && err == nil {
					t.Fatal("unsafe count/time admitted")
				}
			})
		}
	}
}

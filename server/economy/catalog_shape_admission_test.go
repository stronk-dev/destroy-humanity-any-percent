package economy

import (
	"bytes"
	"encoding/json"
	"os"
	"sort"
	"strconv"
	"strings"
	"testing"
)

func axisCatalogBytes(t *testing.T) []byte {
	t.Helper()
	data, err := os.ReadFile("../../balance/testdata/axis-stack/economy-v5-fixture.json")
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestRemainingCatalogCompositeAndAxisArmsPreserveValidControls(t *testing.T) {
	composite := editEconomyBytes(t, foundationCatalogBytes(t), func(v map[string]any) {
		v["progress_coordinates"].([]any)[0] = map[string]any{"tier": 0, "kind": "composite", "terms": []any{
			map[string]any{"weight": "5e-1", "kind": "resource_log", "resource": "company.cash", "target": "1e3"},
			map[string]any{"weight": "5e-1", "kind": "count_fraction", "count": "generators.total_owned", "required": 10},
		}}
	})
	if _, err := LoadCatalog(composite); err != nil {
		t.Fatal("valid composite", err)
	}
	for index, key := range []string{"count", "resource"} {
		candidate := editEconomyBytes(t, composite, func(v map[string]any) {
			v["progress_coordinates"].([]any)[0].(map[string]any)["terms"].([]any)[index].(map[string]any)[key] = nil
		})
		if _, err := LoadCatalog(candidate); err == nil {
			t.Fatal("term's wrong arm null admitted")
		}
	}
	axis := editEconomyBytes(t, axisCatalogBytes(t), func(v map[string]any) {
		changed := false
		for _, upgrade := range v["upgrades"].([]any) {
			for _, value := range upgrade.(map[string]any)["effects"].([]any) {
				effect := value.(map[string]any)
				if effect["slot"] == "axis_stack" {
					effect["factor"] = nil
					changed = true
				}
			}
		}
		if !changed {
			t.Fatal("axis control had no axis effect")
		}
	})
	if _, err := LoadCatalog(axis); err == nil {
		t.Fatal("axis wrong-arm null admitted")
	}
	original := editEconomyBytes(t, foundationCatalogBytes(t), func(map[string]any) {})
	for _, candidate := range [][]byte{
		bytes.Replace(original, []byte(`"schema_version":4`), []byte(`"schema_version":1,"schema_version":4`), 1),
		bytes.Replace(original, []byte(`"copy_key":`), []byte(`"copy_k\u0065y":`), 1),
	} {
		if bytes.Equal(candidate, original) {
			t.Fatal("control did not alter bytes")
		}
		if _, err := LoadCatalog(candidate); err != nil {
			t.Fatal("decoded-name/last-value behavior changed", err)
		}
	}
}

func editEconomyBytes(t *testing.T, data []byte, edit func(map[string]any)) []byte {
	t.Helper()
	var root map[string]any
	if err := json.Unmarshal(data, &root); err != nil {
		t.Fatal(err)
	}
	edit(root)
	result, err := json.Marshal(root)
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func TestRemainingCatalogObjectsRequireLiteralMembers(t *testing.T) {
	for name, data := range map[string][]byte{"v4": foundationCatalogBytes(t), "v5": axisCatalogBytes(t)} {
		t.Run(name, func(t *testing.T) {
			var root map[string]any
			if err := json.Unmarshal(data, &root); err != nil {
				t.Fatal(err)
			}
			var inspect func(string, any)
			inspect = func(path string, source any) {
				switch value := source.(type) {
				case map[string]any:
					keys := make([]string, 0, len(value))
					for key := range value {
						keys = append(keys, key)
					}
					sort.Strings(keys)
					for _, key := range keys {
						t.Run(path+"/"+key, func(t *testing.T) {
							original := value[key]
							delete(value, key)
							value[strings.ToUpper(key)] = original
							defer func() { delete(value, strings.ToUpper(key)); value[key] = original }()
							candidate, err := json.Marshal(root)
							if err != nil {
								t.Fatal(err)
							}
							if _, err := LoadCatalog(candidate); err == nil {
								t.Fatal("case alias admitted")
							}
						})
					}
					for _, key := range keys {
						if path == "root" && (key == "resources" || key == "generator_classes" || key == "manual_policy" || key == "offline_policy") {
							continue
						} // covered by preceding component regressions
						inspect(path+"/"+key, value[key])
					}
				case []any:
					for index, child := range value {
						inspect(path+"/"+strconv.Itoa(index), child)
					}
				}
			}
			inspect("root", root)
		})
	}
}

func TestCatalogVersionAndArmPresenceAreNotDefaults(t *testing.T) {
	first := func(v map[string]any) map[string]any { return v["progress_coordinates"].([]any)[0].(map[string]any) }
	for name, edit := range map[string]func(map[string]any){
		"missing tier zero":                func(v map[string]any) { delete(first(v), "tier") },
		"null tier zero":                   func(v map[string]any) { first(v)["tier"] = nil },
		"resource progress wrong count":    func(v map[string]any) { first(v)["count"] = nil },
		"resource progress wrong required": func(v map[string]any) { first(v)["required"] = 0 },
		"resource progress wrong terms":    func(v map[string]any) { first(v)["terms"] = nil },
		"static effect wrong factor": func(v map[string]any) {
			v["upgrades"].([]any)[0].(map[string]any)["effects"].([]any)[0].(map[string]any)["factor_ppm"] = nil
		},
		"v4 axis null": func(v map[string]any) { v["axis_stack"] = nil },
	} {
		t.Run(name, func(t *testing.T) {
			data := editEconomyBytes(t, foundationCatalogBytes(t), edit)
			if _, err := LoadCatalog(data); err == nil {
				t.Fatal("wrong presence treated as absent/default")
			}
		})
	}
	for _, key := range []string{"manual_actions", "multiplier_sources", "progress_coordinates", "manual_policy", "offline_policy", "upgrades", "synergy_pools", "axis_stack"} {
		t.Run("legacy/null/"+key, func(t *testing.T) {
			data := editEconomyBytes(t, loadKernelFixture(t).Catalog, func(v map[string]any) { v[key] = nil })
			if _, err := LoadCatalog(data); err == nil {
				t.Fatal("legacy version admitted later-version key")
			}
		})
	}
	for name, data := range map[string][]byte{"v4": foundationCatalogBytes(t), "v5": axisCatalogBytes(t), "v2": loadKernelFixture(t).Catalog} {
		if _, err := LoadCatalog(data); err != nil {
			t.Fatal(name, "original control refused", err)
		}
	}
}

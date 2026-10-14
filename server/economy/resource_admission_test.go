package economy

import (
	"encoding/json"
	"sort"
	"strings"
	"testing"
)

func resourceCandidate(t *testing.T, edit func(map[string]any)) []byte {
	t.Helper()
	var root map[string]any
	if err := json.Unmarshal(loadKernelFixture(t).Catalog, &root); err != nil {
		t.Fatal(err)
	}
	edit(root["resources"].([]any)[0].(map[string]any))
	data, err := json.Marshal(root)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestResourceCatalogRequiresLiteralRowAndHardcapKeys(t *testing.T) {
	for _, scope := range []string{"resource", "hardcap"} {
		var root map[string]any
		if err := json.Unmarshal(resourceCandidate(t, func(map[string]any) {}), &root); err != nil {
			t.Fatal(err)
		}
		row := root["resources"].([]any)[0].(map[string]any)
		if scope == "hardcap" {
			row = row["hardcap"].(map[string]any)
		}
		keys := make([]string, 0, len(row))
		for key := range row {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			t.Run(scope+"/"+key, func(t *testing.T) {
				data := resourceCandidate(t, func(row map[string]any) {
					if scope == "hardcap" {
						row = row["hardcap"].(map[string]any)
					}
					row[strings.ToUpper(key)] = row[key]
					delete(row, key)
				})
				if _, err := LoadCatalog(data); err == nil {
					t.Fatal("nonliteral resource/cap key admitted")
				}
			})
		}
	}
	for _, edit := range []func(map[string]any){func(map[string]any) {}, func(v map[string]any) { v["hardcap"] = nil }} {
		if _, err := LoadCatalog(resourceCandidate(t, edit)); err != nil {
			t.Fatal("legitimate cap/null control rejected", err)
		}
	}
}

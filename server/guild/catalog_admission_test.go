package guild

import (
	"encoding/json"
	"sort"
	"strings"
	"testing"
)

func TestGuildCatalogRequiresLiteralNonNullMembers(t *testing.T) {
	var root map[string]json.RawMessage
	if err := json.Unmarshal([]byte(phase0Catalog), &root); err != nil {
		t.Fatal(err)
	}
	if len(root) != 11 {
		t.Fatal("changed catalog member population")
	}
	keys := make([]string, 0, len(root))
	for key := range root {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		value := root[key]
		t.Run("alias/"+key, func(t *testing.T) {
			delete(root, key)
			root[strings.ToUpper(key)] = value
			defer func() { delete(root, strings.ToUpper(key)); root[key] = value }()
			candidate, err := json.Marshal(root)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := LoadCatalog(candidate); err == nil {
				t.Fatal("case alias admitted")
			}
		})
	}
	key := "consumption_bonus_ppm_per_unit"
	for _, variant := range []string{"missing", "null"} {
		t.Run("zero field/"+variant, func(t *testing.T) {
			value := root[key]
			defer func() { root[key] = value }()
			if variant == "missing" {
				delete(root, key)
			} else {
				root[key] = json.RawMessage("null")
			}
			candidate, err := json.Marshal(root)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := LoadCatalog(candidate); err == nil {
				t.Fatal("required bonus became implicit zero")
			}
		})
	}
	if _, err := LoadCatalog([]byte(phase0Catalog)); err != nil {
		t.Fatal("explicit zero rejected", err)
	}
}

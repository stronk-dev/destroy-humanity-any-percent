package pitch

import (
	"encoding/json"
	"errors"
	"os"
	"testing"

	"cloud-clicker/server/copykeys"
)

func loadFixture(t *testing.T) ([]byte, *Catalog) {
	t.Helper()
	data, err := os.ReadFile("../../balance/testdata/pitch-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	keys := make(map[string]struct{})
	for _, key := range copykeys.All() {
		keys[key] = struct{}{}
	}
	catalog, err := LoadCatalog(data, Declarations{CopyKeys: keys})
	if err != nil {
		t.Fatal(err)
	}
	return data, catalog
}

func TestLoadCatalogClosesPitchContentSurface(t *testing.T) {
	data, catalog := loadFixture(t)
	if catalog.SchemaVersion != 1 || len(catalog.MetricCards) != 12 || len(catalog.GrowthHacks) != 8 ||
		len(catalog.CardInstances()) != 24 || ContentHash(data)[:7] != "sha256:" {
		t.Fatalf("unexpected catalog: %#v", catalog)
	}
	if hack, ok := catalog.Hack("stealth_mode"); !ok || hack.Effect.Kind != "chain_factor" || hack.Effect.PartnerHackID != "pivot" {
		t.Fatalf("chain hack=%+v found=%v", hack, ok)
	}
}

func TestLoadCatalogRejectsMissingAndAmbiguousRows(t *testing.T) {
	data, _ := loadFixture(t)
	var root map[string]any
	if err := json.Unmarshal(data, &root); err != nil {
		t.Fatal(err)
	}
	cases := map[string]func(map[string]any){
		"missing root":   func(value map[string]any) { delete(value, "funding_curve") },
		"missing policy": func(value map[string]any) { delete(value["policy"].(map[string]any), "hand_size") },
		"missing card":   func(value map[string]any) { delete(value["metric_cards"].([]any)[0].(map[string]any), "copies") },
		"missing hack":   func(value map[string]any) { delete(value["growth_hacks"].([]any)[0].(map[string]any), "draft_weight") },
		"missing effect": func(value map[string]any) {
			delete(value["growth_hacks"].([]any)[0].(map[string]any)["effect"].(map[string]any), "factor")
		},
		"extra effect": func(value map[string]any) {
			value["growth_hacks"].([]any)[0].(map[string]any)["effect"].(map[string]any)["amount"] = "1e0"
		},
		"missing target": func(value map[string]any) {
			delete(value["funding_curve"].([]any)[0].(map[string]any), "funding_target")
		},
	}
	keys := make(map[string]struct{})
	for _, key := range copykeys.All() {
		keys[key] = struct{}{}
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			var value map[string]any
			if err := json.Unmarshal(data, &value); err != nil {
				t.Fatal(err)
			}
			mutate(value)
			encoded, _ := json.Marshal(value)
			if _, err := LoadCatalog(encoded, Declarations{CopyKeys: keys}); !errors.Is(err, ErrInvalidCatalog) {
				t.Fatalf("err=%v", err)
			}
		})
	}
}

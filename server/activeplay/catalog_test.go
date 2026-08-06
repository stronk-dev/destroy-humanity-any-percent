package activeplay

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"cloud-clicker/server/economy"
)

func activeEconomy(t *testing.T) *economy.Catalog {
	t.Helper()
	data, err := os.ReadFile("../../balance/catalogs/phase0.json")
	if err != nil {
		t.Fatal(err)
	}
	var root map[string]any
	if json.Unmarshal(data, &root) != nil {
		t.Fatal("invalid economy fixture")
	}
	values := root["multiplier_sources"].([]any)
	values = append(values,
		map[string]any{"id": "active.building.generator.beige_tower", "slot": "event_buffs", "target": "generator.beige_tower", "provider": "active_play"},
		map[string]any{"id": "active.click", "slot": "event_buffs", "target": "manual.click", "provider": "active_play"},
		map[string]any{"id": "active.production", "slot": "event_buffs", "target": "all", "provider": "active_play"},
	)
	root["multiplier_sources"] = values
	candidate, _ := json.Marshal(root)
	catalog, err := economy.LoadCatalog(candidate)
	if err != nil {
		t.Fatal(err)
	}
	return catalog
}

func activeFixture(t *testing.T) []byte {
	t.Helper()
	data, err := os.ReadFile("../../balance/testdata/active-play-foundation-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	var root struct {
		Baseline json.RawMessage `json:"baseline"`
	}
	if json.Unmarshal(data, &root) != nil || len(root.Baseline) == 0 {
		t.Fatal("invalid active-play fixture")
	}
	return root.Baseline
}

func TestCatalogAndScheduleDeterminism(t *testing.T) {
	catalog, err := LoadCatalog(activeFixture(t), activeEconomy(t))
	if err != nil {
		t.Fatal(err)
	}
	first, err := catalog.Spawn("01985555-0000-7000-8000-000000000001", 1, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	second, err := catalog.Spawn("01985555-0000-7000-8000-000000000001", 1, 0, 0)
	if err != nil || !reflect.DeepEqual(first, second) || first.SampledIntervalMS <= catalog.Schedule.MinimumIntervalMS ||
		first.SpawnedAttendedMS != first.SampledIntervalMS || first.ExpiresAttendedMS-first.SpawnedAttendedMS != catalog.Schedule.LifetimeMS {
		t.Fatalf("non-deterministic spawn: first=%+v second=%+v err=%v", first, second, err)
	}
	if first.OpportunityID[14] != '7' || first.OpportunityID[19] < '8' || first.OpportunityID[19] > 'b' {
		t.Fatalf("not UUIDv7-compatible: %s", first.OpportunityID)
	}
}

func TestCatalogRejectsMissingAndCrossArtifactKeys(t *testing.T) {
	var root map[string]any
	if json.Unmarshal(activeFixture(t), &root) != nil {
		t.Fatal("fixture")
	}
	delete(root["schedule_policy"].(map[string]any), "scale_ms")
	candidate, _ := json.Marshal(root)
	if _, err := LoadCatalog(candidate, activeEconomy(t)); err == nil {
		t.Fatal("missing schedule key accepted")
	}
	if _, err := LoadCatalog(activeFixture(t), func() *economy.Catalog {
		data, _ := os.ReadFile("../../balance/catalogs/phase0.json")
		value, _ := economy.LoadCatalog(data)
		return value
	}()); err == nil {
		t.Fatal("missing multiplier declarations accepted")
	}
}

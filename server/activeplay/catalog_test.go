package activeplay

import (
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"

	"cloud-clicker/server/economy"
	"cloud-clicker/server/publicapi"
)

func editActiveFixture(t *testing.T, edit func(map[string]any)) []byte {
	t.Helper()
	var root map[string]any
	if err := json.Unmarshal(activeFixture(t), &root); err != nil {
		t.Fatal(err)
	}
	edit(root)
	data, err := json.Marshal(root)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func activeShapeObjects(root map[string]any) []map[string]any {
	objects := []map[string]any{root, root["schedule_policy"].(map[string]any), root["combo_policy"].(map[string]any)}
	for _, row := range root["effects"].([]any) {
		objects = append(objects, row.(map[string]any))
	}
	return objects
}

func TestCatalogRejectsCaseAliasedObjectMembers(t *testing.T) {
	definitions := activePlaySchemaDefinitions(t)
	var root map[string]any
	if err := json.Unmarshal(activeFixture(t), &root); err != nil {
		t.Fatal(err)
	}
	for index, object := range activeShapeObjects(root) {
		for key := range object {
			t.Run(key+"/"+string(rune('0'+index)), func(t *testing.T) {
				data := editActiveFixture(t, func(candidate map[string]any) {
					row := activeShapeObjects(candidate)[index]
					row[strings.ToUpper(key)] = row[key]
					delete(row, key)
				})
				if _, err := LoadCatalog(data, activeEconomy(t)); err == nil {
					t.Fatal("case alias admitted")
				}
				if err := publicapi.ValidateJSON(ActivePlayCatalogSchemaName, data, definitions); err == nil {
					t.Fatal("descriptor admitted case alias")
				}
			})
		}
	}
}

func TestCatalogRejectsNullFieldsFromOtherEffectArms(t *testing.T) {
	definitions := activePlaySchemaDefinitions(t)
	for index, key := range []string{"factor", "targets", "duration_ms", "action_ids"} {
		t.Run(key, func(t *testing.T) {
			data := editActiveFixture(t, func(root map[string]any) {
				root["effects"].([]any)[index].(map[string]any)[key] = nil
			})
			if _, err := LoadCatalog(data, activeEconomy(t)); err == nil {
				t.Fatal("null field from another effect arm admitted")
			}
			if err := publicapi.ValidateJSON(ActivePlayCatalogSchemaName, data, definitions); err == nil {
				t.Fatal("descriptor admitted wrong-arm null")
			}
		})
	}
}

func TestCatalogRejectsUnsafeEffectDurations(t *testing.T) {
	definitions := activePlaySchemaDefinitions(t)
	for _, index := range []int{0, 1, 3} {
		data := editActiveFixture(t, func(root map[string]any) {
			root["effects"].([]any)[index].(map[string]any)["duration_ms"] = json.Number("9007199254740992")
		})
		if _, err := LoadCatalog(data, activeEconomy(t)); err == nil {
			t.Errorf("effect %d admitted duration outside shared exact domain", index)
		}
		if err := publicapi.ValidateJSON(ActivePlayCatalogSchemaName, data, definitions); err == nil {
			t.Errorf("descriptor admitted unsafe duration on effect %d", index)
		}
	}
}

func activeEconomy(t *testing.T) *economy.Catalog {
	t.Helper()
	data, err := os.ReadFile("../../balance/catalogs/phase0.json")
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := economy.LoadCatalog(data)
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
		var root map[string]any
		_ = json.Unmarshal(data, &root)
		values := root["multiplier_sources"].([]any)
		kept := make([]any, 0, len(values))
		for _, raw := range values {
			if raw.(map[string]any)["provider"] != "active_play" {
				kept = append(kept, raw)
			}
		}
		root["multiplier_sources"] = kept
		candidate, _ := json.Marshal(root)
		value, _ := economy.LoadCatalog(candidate)
		return value
	}()); err == nil {
		t.Fatal("missing multiplier declarations accepted")
	}
}

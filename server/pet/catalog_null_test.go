package pet

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
)

type petZeroCatalogField struct {
	path string
	key  string
	row  map[string]any
}

func zeroPetCatalogFields(root map[string]any) []petZeroCatalogField {
	var fields []petZeroCatalogField
	add := func(path string, row map[string]any, keys ...string) {
		for _, key := range keys {
			row[key] = 0
			fields = append(fields, petZeroCatalogField{path + "/" + key, key, row})
		}
	}
	stat := root["stat_policy"].(map[string]any)
	add("stat_policy", stat, "diminishing_threshold_ppm", "diminishing_factor_ppm")
	for index, source := range stat["stats"].([]any) {
		add(fmt.Sprintf("stat/%d", index), source.(map[string]any), "initial_ppm", "floor_ppm", "decay_ppm_per_grid")
	}
	for index, source := range root["actions"].([]any) {
		add(fmt.Sprintf("action/%d", index), source.(map[string]any), "cooldown_attended_ms", "min_eligible_ppm")
	}
	add("trust_policy", root["trust_policy"].(map[string]any), "initial_ppm", "neutral_ppm", "floor_ppm", "cap_ppm", "gain_ppm_per_effective_action", "decay_ppm_per_grid")
	add("mood/0", root["mood_policy"].([]any)[0].(map[string]any), "floor_ppm")
	return fields
}

func TestPetCatalogRefusesNullNumericFieldsAndPreservesZero(t *testing.T) {
	data, err := os.ReadFile("../../balance/pets/first-content.json")
	if err != nil {
		t.Fatal(err)
	}
	var root map[string]any
	if err := json.Unmarshal(data, &root); err != nil {
		t.Fatal(err)
	}
	fields := zeroPetCatalogFields(root)
	if len(fields) != 31 {
		t.Fatalf("changed regression population: %d", len(fields))
	}
	if _, err := LoadCatalog([]byte(mustJSON(root))); err != nil {
		t.Fatalf("legal explicit-zero control: %v", err)
	}
	for _, field := range fields {
		t.Run(field.path, func(t *testing.T) {
			field.row[field.key] = nil
			defer func() { field.row[field.key] = 0 }()
			if _, err := LoadCatalog([]byte(mustJSON(root))); err == nil {
				t.Fatal("required null numeric value became an implicit zero")
			}
		})
	}
}

func TestPetCatalogGrammarRefusesNullMoodFloor(t *testing.T) {
	data, err := os.ReadFile("../../testdata/pet/catalog-grammar-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	var root map[string]any
	if err := json.Unmarshal(data, &root); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadCatalogGrammar(data); err != nil {
		t.Fatal(err)
	}
	root["mood_thresholds"].([]any)[0].(map[string]any)["floor_ppm"] = nil
	if _, err := LoadCatalogGrammar([]byte(mustJSON(root))); err == nil {
		t.Fatal("required standalone mood floor defaulted from null")
	}
}

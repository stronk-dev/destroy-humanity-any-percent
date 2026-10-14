package pet

import (
	"bytes"
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"cloud-clicker/server/publicapi"
)

func petSchemaDefinitions(t *testing.T) map[string]*publicapi.Schema {
	t.Helper()
	definitions, err := publicapi.ValidateSchemaDefinitions(PetSchemas())
	if err != nil {
		t.Fatal(err)
	}
	return definitions
}

func editPetCatalog(t *testing.T, edit func(map[string]any)) []byte {
	t.Helper()
	data, err := os.ReadFile("../../balance/pets/first-content.json")
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
	return []byte(mustJSON(root))
}

func TestPetOwnerSchemasAdmitBothVersionsAndStoredBytes(t *testing.T) {
	definitions := petSchemaDefinitions(t)
	data := [][]byte{[]byte(fullCatalogFixture)}
	for _, path := range []string{
		"../../balance/pets/first-content.json",
		"../../testdata/harness/content-dynamics/bundles/sha256:6c7fab29c24fae68e3067c883177bc78fe61b9d91704b6d936b3e4f3cfd8f789/artifacts/pets.json",
		"../../testdata/harness/content-dynamics/bundles/sha256:baa890501b2864d14cc0238d633a562cb8c6fca406190487831e0c447af128f6/artifacts/pets.json",
	} {
		bytes, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		data = append(data, bytes)
	}
	data = append(data, editPetCatalog(t, func(v map[string]any) {
		zeroPetCatalogFields(v)
		v["stat_policy"].(map[string]any)["grid_ms"] = json.Number("9007199254740991")
		v["behavior_policy"].([]any)[0].(map[string]any)["duration_grid_ticks"] = json.Number("9007199254740991")
	}))
	for _, candidate := range data {
		before := bytes.Clone(candidate)
		if err := publicapi.ValidateJSON(PetCatalogSchemaName, candidate, definitions); err != nil {
			t.Fatal(err)
		}
		if _, err := LoadCatalog(candidate); err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(before, candidate) {
			t.Fatal("validation changed artifact bytes")
		}
	}
	for _, gate := range []string{"essential", "ordinary", "recovery"} {
		candidate := editPetCatalog(t, func(v map[string]any) { v["actions"].([]any)[0].(map[string]any)["soul_gate"] = gate })
		if err := publicapi.ValidateJSON(PetCatalogSchemaName, candidate, definitions); err != nil {
			t.Fatal(err)
		}
		if _, err := LoadCatalog(candidate); err != nil {
			t.Fatal(err)
		}
	}
	for _, from := range BehaviorStates() {
		for _, event := range BehaviorEvents() {
			for _, to := range BehaviorStates() {
				candidate := editPetCatalog(t, func(v map[string]any) {
					v["behavior_policy"] = []any{map[string]any{"from_state": from, "event": event, "to_state": to, "duration_grid_ticks": 1}}
				})
				if err := publicapi.ValidateJSON(PetCatalogSchemaName, candidate, definitions); err != nil {
					t.Fatal(err)
				}
				if _, err := LoadCatalog(candidate); err != nil {
					t.Fatal(err)
				}
			}
		}
	}
}

func TestPetOwnerSchemaClosesVersionedActionAndPolicyShapes(t *testing.T) {
	definitions := petSchemaDefinitions(t)
	stat := func(v map[string]any) map[string]any { return v["stat_policy"].(map[string]any) }
	action := func(v map[string]any) map[string]any { return v["actions"].([]any)[0].(map[string]any) }
	for name, edit := range map[string]func(map[string]any){
		"missing root":      func(v map[string]any) { delete(v, "trust_policy") },
		"private root":      func(v map[string]any) { v["pet_id"] = "private" },
		"version":           func(v map[string]any) { v["schema_version"] = 3 },
		"root alias":        func(v map[string]any) { v["STAT_POLICY"] = v["stat_policy"]; delete(v, "stat_policy") },
		"null diminution":   func(v map[string]any) { stat(v)["diminishing_factor_ppm"] = nil },
		"unsafe grid":       func(v map[string]any) { stat(v)["grid_ms"] = json.Number("9007199254740992") },
		"grid zero":         func(v map[string]any) { stat(v)["grid_ms"] = 0 },
		"stat enum":         func(v map[string]any) { stat(v)["stats"].([]any)[0].(map[string]any)["stat_id"] = "health" },
		"stat maximum":      func(v map[string]any) { stat(v)["stats"].([]any)[0].(map[string]any)["initial_ppm"] = 1000001 },
		"missing gate":      func(v map[string]any) { delete(action(v), "soul_gate") },
		"gate enum":         func(v map[string]any) { action(v)["soul_gate"] = "human_hobby" },
		"legacy gate":       func(v map[string]any) { v["schema_version"] = 1 },
		"delta zero":        func(v map[string]any) { action(v)["delta_ppm"] = 0 },
		"null cooldown":     func(v map[string]any) { action(v)["cooldown_attended_ms"] = nil },
		"unsafe cooldown":   func(v map[string]any) { action(v)["cooldown_attended_ms"] = json.Number("9007199254740992") },
		"invalid action ID": func(v map[string]any) { action(v)["action_id"] = "bad-id" },
		"null trust":        func(v map[string]any) { v["trust_policy"].(map[string]any)["floor_ppm"] = nil },
		"negative trust":    func(v map[string]any) { v["trust_policy"].(map[string]any)["gain_ppm_per_effective_action"] = -1 },
		"mood enum":         func(v map[string]any) { v["mood_policy"].([]any)[0].(map[string]any)["mood_member"] = "happy" },
		"behavior state":    func(v map[string]any) { v["behavior_policy"].([]any)[0].(map[string]any)["from_state"] = "missing" },
		"behavior event":    func(v map[string]any) { v["behavior_policy"].([]any)[0].(map[string]any)["event"] = "offline_tick" },
		"duration zero":     func(v map[string]any) { v["behavior_policy"].([]any)[0].(map[string]any)["duration_grid_ticks"] = 0 },
	} {
		t.Run(name, func(t *testing.T) {
			data := editPetCatalog(t, edit)
			if err := publicapi.ValidateJSON(PetCatalogSchemaName, data, definitions); err == nil {
				t.Fatal("schema admitted invalid catalog")
			}
			if _, err := LoadCatalog(data); err == nil {
				t.Fatal("owner admitted invalid catalog")
			}
		})
	}
}

func TestPetOwnerSchemasRetainRelationsAndRawMemberChecks(t *testing.T) {
	definitions := petSchemaDefinitions(t)
	for name, edit := range map[string]func(map[string]any){
		"stat cardinality": func(v map[string]any) {
			stat := v["stat_policy"].(map[string]any)
			stat["stats"] = stat["stats"].([]any)[:3]
		},
		"initial below floor": func(v map[string]any) {
			v["stat_policy"].(map[string]any)["stats"].([]any)[0].(map[string]any)["initial_ppm"] = 0
		},
		"trust order":              func(v map[string]any) { v["trust_policy"].(map[string]any)["floor_ppm"] = 500001 },
		"action floor eligibility": func(v map[string]any) { v["actions"].([]any)[0].(map[string]any)["min_eligible_ppm"] = 100001 },
		"action order":             func(v map[string]any) { rows := v["actions"].([]any); rows[0], rows[1] = rows[1], rows[0] },
		"mood order":               func(v map[string]any) { rows := v["mood_policy"].([]any); rows[0], rows[1] = rows[1], rows[0] },
		"nondeterministic FSM": func(v map[string]any) {
			rows := v["behavior_policy"].([]any)
			first := rows[0].(map[string]any)
			v["behavior_policy"] = append(rows, map[string]any{"from_state": first["from_state"], "event": first["event"], "to_state": "resting", "duration_grid_ticks": 1})
		},
	} {
		t.Run(name, func(t *testing.T) {
			data := editPetCatalog(t, edit)
			if err := publicapi.ValidateJSON(PetCatalogSchemaName, data, definitions); err != nil {
				t.Fatal(err)
			}
			if _, err := LoadCatalog(data); err == nil {
				t.Fatal("descriptor replaced semantic admission")
			}
		})
	}
	source := []byte(fullCatalogFixture)
	escaped := bytes.Replace(source, []byte(`"schema_version"`), []byte(`"schema_versi\u006fn"`), 1)
	if bytes.Equal(source, escaped) {
		t.Fatal("escape control did not change bytes")
	}
	if _, err := LoadCatalog(escaped); err != nil {
		t.Fatal(err)
	}
	if err := publicapi.ValidateJSON(PetCatalogSchemaName, escaped, definitions); err != nil {
		t.Fatal(err)
	}
	duplicate := append(bytes.Clone(source[:len(source)-1]), []byte(`,"schema_version":1}`)...)
	if err := publicapi.ValidateJSON(PetCatalogSchemaName, duplicate, definitions); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadCatalog(duplicate); err == nil {
		t.Fatal("raw unique-member gate disappeared")
	}
	before, fresh := PetSchemas(), PetSchemas()
	if !reflect.DeepEqual(before, fresh) {
		t.Fatal("nondeterministic export")
	}
	for _, field := range fresh[1].Schema.Fields {
		if field.Name == "schema_version" {
			*field.Schema.Maximum = 0
		}
	}
	if !reflect.DeepEqual(before, PetSchemas()) {
		t.Fatal("caller mutation changed owner schema")
	}
}

package economy

import (
	"encoding/json"
	"sort"
	"strings"
	"testing"
)

func editFoundationCatalog(t *testing.T, edit func(map[string]any)) []byte {
	t.Helper()
	var root map[string]any
	if err := json.Unmarshal(foundationCatalogBytes(t), &root); err != nil {
		t.Fatal(err)
	}
	edit(root)
	data, err := json.Marshal(root)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestGeneratorAndOperatingPoliciesRequireLiteralKeys(t *testing.T) {
	definitions := generatorPolicyDefinitions(t)
	var root map[string]any
	if err := json.Unmarshal(foundationCatalogBytes(t), &root); err != nil {
		t.Fatal(err)
	}
	type target struct {
		path string
		row  map[string]any
	}
	var targets []target
	for _, source := range root["generator_classes"].([]any) {
		row := source.(map[string]any)
		id := row["id"].(string)
		targets = append(targets, target{id, row})
		for _, key := range []string{"price", "production", "provisions", "provisioned_hardcap"} {
			if value, present := row[key]; present {
				targets = append(targets, target{id + "/" + key, value.(map[string]any)})
			}
		}
		for _, key := range []string{"ladder", "roles"} {
			for _, value := range row[key].([]any) {
				child := value.(map[string]any)
				name := key
				if key == "roles" {
					name += "/" + child["kind"].(string)
				}
				targets = append(targets, target{id + "/" + name, child})
			}
		}
	}
	for _, key := range []string{"manual_policy", "offline_policy"} {
		targets = append(targets, target{key, root[key].(map[string]any)})
	}
	for _, target := range targets {
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
				data, err := json.Marshal(root)
				if err != nil {
					t.Fatal(err)
				}
				if _, err := LoadCatalog(data); err == nil {
					t.Fatal("nonliteral generator/policy member admitted")
				}
				if err := validateGeneratorPolicyComponents(data, definitions); err == nil {
					t.Fatal("descriptor admitted nonliteral member")
				}
			})
		}
	}
}

func TestGeneratorRolesAndOptionalObjectsRejectWrongArmNulls(t *testing.T) {
	definitions := generatorPolicyDefinitions(t)
	var root map[string]any
	if err := json.Unmarshal(foundationCatalogBytes(t), &root); err != nil {
		t.Fatal(err)
	}
	for index, source := range root["generator_classes"].([]any) {
		row := source.(map[string]any)
		for roleIndex, sourceRole := range row["roles"].([]any) {
			role := sourceRole.(map[string]any)
			for _, key := range []string{"generator_id", "pool_id", "action_id", "per_purchased_ppm"} {
				if _, present := role[key]; present {
					continue
				}
				t.Run(role["kind"].(string)+"/null/"+key, func(t *testing.T) {
					data := editFoundationCatalog(t, func(v map[string]any) {
						v["generator_classes"].([]any)[index].(map[string]any)["roles"].([]any)[roleIndex].(map[string]any)[key] = nil
					})
					if _, err := LoadCatalog(data); err == nil {
						t.Fatal("wrong-arm null looked absent")
					}
					if err := validateGeneratorPolicyComponents(data, definitions); err == nil {
						t.Fatal("descriptor admitted wrong-arm null")
					}
				})
			}
		}
		for _, key := range []string{"provisions", "provisioned_hardcap"} {
			if _, present := row[key]; present {
				continue
			}
			t.Run(row["id"].(string)+"/null/"+key, func(t *testing.T) {
				data := editFoundationCatalog(t, func(v map[string]any) { v["generator_classes"].([]any)[index].(map[string]any)[key] = nil })
				if _, err := LoadCatalog(data); err == nil {
					t.Fatal("explicit null optional object looked absent")
				}
				if err := validateGeneratorPolicyComponents(data, definitions); err == nil {
					t.Fatal("descriptor admitted null optional object")
				}
			})
		}
	}
	if _, err := LoadCatalog(foundationCatalogBytes(t)); err != nil {
		t.Fatal("original role/optional-object controls rejected", err)
	}
}

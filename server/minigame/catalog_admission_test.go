package minigame

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func editMinigameFixture(t *testing.T, edit func(map[string]any)) []byte {
	t.Helper()
	data, err := os.ReadFile("../../testdata/minigame/catalog-v2.json")
	if err != nil {
		t.Fatal(err)
	}
	var root map[string]any
	if err := json.Unmarshal(data, &root); err != nil {
		t.Fatal(err)
	}
	edit(root["minigames"].([]any)[0].(map[string]any))
	result, err := json.Marshal(root)
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func TestMinigameCatalogRefusesNullPolicyValuesAndImplicitCurveFields(t *testing.T) {
	for section, keys := range map[string][]string{
		"payout":          {"sends_per_day", "per_send_cap", "conversion_ppm"},
		"rating_policy":   {"starting_elo", "elo_floor", "provisional_games"},
		"offline_quality": {"decay_ppm_per_grid"},
	} {
		for _, key := range keys {
			t.Run(section+"/"+key, func(t *testing.T) {
				data := editMinigameFixture(t, func(v map[string]any) { v[section].(map[string]any)[key] = nil })
				if _, err := LoadCatalog(data); err == nil {
					t.Fatal("required null policy value became zero")
				}
			})
		}
	}
	for _, key := range []string{"score_threshold", "grade_ppm"} {
		for _, variant := range []string{"alias", "null", "missing"} {
			t.Run("curve/"+key+"/"+variant, func(t *testing.T) {
				data := editMinigameFixture(t, func(v map[string]any) {
					quality := v["offline_quality"].(map[string]any)
					quality["neutral_floor_ppm"] = 0
					row := quality["grade_curve"].([]any)[0].(map[string]any)
					row["grade_ppm"] = 0
					switch variant {
					case "alias":
						row[strings.ToUpper(key)] = row[key]
						delete(row, key)
					case "null":
						row[key] = nil
					case "missing":
						delete(row, key)
					}
				})
				if _, err := LoadCatalog(data); err == nil {
					t.Fatal("malformed exact curve row admitted")
				}
			})
		}
	}
}

func TestMinigameCatalogRefusesIncompleteFallbackIdentity(t *testing.T) {
	for kind, idKey := range map[string]string{"bot": "policy_id", "npc_partner": "profile_id"} {
		for _, variant := range []string{"id alias", "version alias", "null reduction"} {
			t.Run(kind+"/"+variant, func(t *testing.T) {
				data := editMinigameFixture(t, func(v map[string]any) {
					identity := map[string]any{idKey: "fixture.policy", "version": "1.0.0"}
					fallback := map[string]any{"kind": kind, "rate_reduction_ppm": 0}
					if kind == "bot" {
						fallback["bot_ref"] = identity
					} else {
						fallback["npc_profile"] = identity
					}
					switch variant {
					case "id alias":
						identity[strings.ToUpper(idKey)] = identity[idKey]
						delete(identity, idKey)
					case "version alias":
						identity["VERSION"] = identity["version"]
						delete(identity, "version")
					case "null reduction":
						fallback["rate_reduction_ppm"] = nil
					}
					v["fallback"] = fallback
				})
				if _, err := LoadCatalog(data); err == nil {
					t.Fatal("malformed fallback identity/reduction admitted")
				}
			})
		}
	}
	data := editMinigameFixture(t, func(v map[string]any) {
		for _, key := range []string{"sends_per_day", "per_send_cap", "conversion_ppm"} {
			v["payout"].(map[string]any)[key] = 0
		}
		for _, key := range []string{"starting_elo", "elo_floor", "provisional_games"} {
			v["rating_policy"].(map[string]any)[key] = 0
		}
		quality := v["offline_quality"].(map[string]any)
		quality["neutral_floor_ppm"], quality["decay_ppm_per_grid"] = 0, 0
		quality["grade_curve"].([]any)[0].(map[string]any)["grade_ppm"] = 0
	})
	if _, err := LoadCatalog(data); err != nil {
		t.Fatal("explicit-zero policies rejected", err)
	}
}

package minigame

import (
	"bytes"
	"os"
	"reflect"
	"testing"

	"cloud-clicker/server/publicapi"
)

func minigameSchemaDefinitions(t *testing.T) map[string]*publicapi.Schema {
	t.Helper()
	definitions, err := publicapi.ValidateSchemaDefinitions(MinigameSchemas())
	if err != nil {
		t.Fatal(err)
	}
	return definitions
}

func TestMinigameOwnerSchemasAdmitHistoricalCurrentAndNonemptyCatalogs(t *testing.T) {
	definitions := minigameSchemaDefinitions(t)
	for _, path := range []string{
		"../../balance/minigames/first-content.json", "../../testdata/minigame/catalog-v2.json",
		"../../testdata/minigame/pitch-v3.json", "../../testdata/minigame/pitch-typer-v3.json", "../../testdata/minigame/pitch-typer-arcade-v3.json",
		"../../testdata/harness/content-dynamics/bundles/sha256:6c7fab29c24fae68e3067c883177bc78fe61b9d91704b6d936b3e4f3cfd8f789/artifacts/minigames.json",
		"../../testdata/harness/content-dynamics/bundles/sha256:baa890501b2864d14cc0238d633a562cb8c6fca406190487831e0c447af128f6/artifacts/minigames.json",
	} {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		before := bytes.Clone(data)
		if err := publicapi.ValidateJSON(MinigameCatalogSchemaName, data, definitions); err != nil {
			t.Fatalf("%s: %v", path, err)
		}
		if _, err := LoadCatalog(data); err != nil {
			t.Fatalf("%s: %v", path, err)
		}
		if !bytes.Equal(before, data) {
			t.Fatal("validation changed artifact bytes")
		}
	}
	for _, fallback := range []map[string]any{
		{"kind": "solo"}, {"kind": "bot", "bot_ref": map[string]any{"policy_id": "fixture.bot", "version": "1.0.0"}, "rate_reduction_ppm": 0},
		{"kind": "npc_partner", "npc_profile": map[string]any{"profile_id": "fixture.npc", "version": "1.0.0"}, "rate_reduction_ppm": 1000000},
	} {
		for _, unlock := range []map[string]any{
			{"kind": "always"}, {"kind": "fact_equals", "fact_id": "fact.example", "value": false},
			{"kind": "fact_equals", "fact_id": "fact.example", "value": ""}, {"kind": "fact_equals", "fact_id": "fact.example", "value": int64(-9007199254740991)},
			{"kind": "fiscal_unlock", "unlock_id": "fixture.unlock"}, {"kind": "tier_at_least", "tier": 0},
			{"kind": "tier_at_least", "tier": 9, "exit_history_at_least": int64(9007199254740991)},
		} {
			data := editMinigameFixture(t, func(v map[string]any) { v["fallback"], v["unlock_condition"] = fallback, unlock })
			if err := publicapi.ValidateJSON(MinigameCatalogSchemaName, data, definitions); err != nil {
				t.Fatal(err)
			}
			if _, err := LoadCatalog(data); err != nil {
				t.Fatal(err)
			}
		}
	}
}

func TestMinigameOwnerSchemaClosesPolicyUnionsAndVersionedSoulGate(t *testing.T) {
	definitions := minigameSchemaDefinitions(t)
	v3, err := os.ReadFile("../../testdata/minigame/pitch-v3.json")
	if err != nil {
		t.Fatal(err)
	}
	for _, replacement := range []string{`"soul_gate":null`, `"soul_gate":"future"`, `"SOUL_GATE":"human_hobby"`} {
		candidate := bytes.Replace(v3, []byte(`"soul_gate": "human_hobby"`), []byte(replacement), 1)
		if bytes.Equal(candidate, v3) {
			t.Fatal("v3 Soul-gate mutation did not change source")
		}
		if err := publicapi.ValidateJSON(MinigameCatalogSchemaName, candidate, definitions); err == nil {
			t.Fatal("v3 descriptor lost required Soul gate")
		}
		if _, err := LoadCatalog(candidate); err == nil {
			t.Fatal("v3 owner lost required Soul gate")
		}
	}
	for name, edit := range map[string]func(map[string]any){
		"private definition": func(v map[string]any) { v["founder_id"] = "private" },
		"future mode":        func(v map[string]any) { v["modes"] = []string{"live_pvp"} },
		"legacy Soul gate":   func(v map[string]any) { v["soul_gate"] = "human_hobby" },
		"power scaling": func(v map[string]any) {
			v["scaling"].(map[string]any)["scaling_inputs"].([]any)[0].(map[string]any)["destination_class"] = "power"
		},
		"unknown source": func(v map[string]any) {
			v["scaling"].(map[string]any)["scaling_inputs"].([]any)[0].(map[string]any)["source_kind"] = "server_result"
		},
		"unknown op": func(v map[string]any) {
			v["scaling"].(map[string]any)["scaling_inputs"].([]any)[0].(map[string]any)["op"] = "pow"
		},
		"null payout":          func(v map[string]any) { v["payout"].(map[string]any)["per_send_cap"] = nil },
		"unsafe payout":        func(v map[string]any) { v["payout"].(map[string]any)["per_send_cap"] = int64(9007199254740992) },
		"fallback extra":       func(v map[string]any) { v["fallback"] = map[string]any{"kind": "solo", "rate_reduction_ppm": 0} },
		"bot missing identity": func(v map[string]any) { v["fallback"] = map[string]any{"kind": "bot", "rate_reduction_ppm": 0} },
		"NPC malformed identity": func(v map[string]any) {
			v["fallback"] = map[string]any{"kind": "npc_partner", "npc_profile": map[string]any{"profile_id": "fixture.npc", "VERSION": "1.0.0"}, "rate_reduction_ppm": 0}
		},
		"quality row missing": func(v map[string]any) {
			delete(v["offline_quality"].(map[string]any)["grade_curve"].([]any)[0].(map[string]any), "score_threshold")
		},
		"quality unknown field": func(v map[string]any) {
			v["offline_quality"].(map[string]any)["grade_curve"].([]any)[0].(map[string]any)["label"] = "private"
		},
		"null rating":    func(v map[string]any) { v["rating_policy"].(map[string]any)["elo_floor"] = nil },
		"unknown unlock": func(v map[string]any) { v["unlock_condition"] = map[string]any{"kind": "client_result"} },
		"null fact value": func(v map[string]any) {
			v["unlock_condition"] = map[string]any{"kind": "fact_equals", "fact_id": "fact.example", "value": nil}
		},
		"tier maximum": func(v map[string]any) { v["unlock_condition"] = map[string]any{"kind": "tier_at_least", "tier": 10} },
	} {
		t.Run(name, func(t *testing.T) {
			data := editMinigameFixture(t, edit)
			if err := publicapi.ValidateJSON(MinigameCatalogSchemaName, data, definitions); err == nil {
				t.Fatal("descriptor admitted malformed artifact")
			}
			if _, err := LoadCatalog(data); err == nil {
				t.Fatal("owner admitted malformed artifact")
			}
		})
	}
}

func TestMinigameOwnerSchemaRetainsSemanticAndRawIdentityChecks(t *testing.T) {
	definitions := minigameSchemaDefinitions(t)
	for name, edit := range map[string]func(map[string]any){
		"modes empty":    func(v map[string]any) { v["modes"] = []string{} },
		"duplicate mode": func(v map[string]any) { v["modes"] = []string{"solo", "solo"} },
		"source ref": func(v map[string]any) {
			v["scaling"].(map[string]any)["scaling_inputs"].([]any)[0].(map[string]any)["source_ref"] = "03"
		},
		"identity operand": func(v map[string]any) {
			v["scaling"].(map[string]any)["scaling_inputs"].([]any)[0].(map[string]any)["operand"] = 1
		},
		"unknown payout fact": func(v map[string]any) { v["payout"].(map[string]any)["payout_score_fact_id"] = "score.unknown" },
		"empty curve":         func(v map[string]any) { v["offline_quality"].(map[string]any)["grade_curve"] = []any{} },
		"curve order": func(v map[string]any) {
			v["offline_quality"].(map[string]any)["grade_curve"].([]any)[1].(map[string]any)["score_threshold"] = 0
		},
		"curve floor":    func(v map[string]any) { v["offline_quality"].(map[string]any)["neutral_floor_ppm"] = 1 },
		"rating order":   func(v map[string]any) { v["rating_policy"].(map[string]any)["elo_floor"] = 1001 },
		"season binding": func(v map[string]any) { v["rating_policy"].(map[string]any)["season_member"] = "unknown" },
	} {
		t.Run(name, func(t *testing.T) {
			data := editMinigameFixture(t, edit)
			if err := publicapi.ValidateJSON(MinigameCatalogSchemaName, data, definitions); err != nil {
				t.Fatal(err)
			}
			if _, err := LoadCatalog(data); err == nil {
				t.Fatal("descriptor replaced mandatory owner checks")
			}
		})
	}
	data, err := os.ReadFile("../../testdata/minigame/catalog-v2.json")
	if err != nil {
		t.Fatal(err)
	}
	escaped := bytes.Replace(data, []byte(`"schema_version"`), []byte(`"schema_versi\u006fn"`), 1)
	if bytes.Equal(data, escaped) {
		t.Fatal("escape control did not change bytes")
	}
	if _, err := LoadCatalog(escaped); err != nil {
		t.Fatal(err)
	}
	if err := publicapi.ValidateJSON(MinigameCatalogSchemaName, escaped, definitions); err != nil {
		t.Fatal(err)
	}
	trimmed := bytes.TrimSpace(data)
	duplicate := append(bytes.Clone(trimmed[:len(trimmed)-1]), []byte(`,"schema_version":2}`)...)
	if err := publicapi.ValidateJSON(MinigameCatalogSchemaName, duplicate, definitions); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadCatalog(duplicate); err == nil {
		t.Fatal("owner's unique-member guard disappeared")
	}
	before, fresh := MinigameSchemas(), MinigameSchemas()
	if !reflect.DeepEqual(before, fresh) {
		t.Fatal("nondeterministic exports")
	}
	for _, definition := range fresh {
		if definition.Name == "MinigamePayoutPolicyV1" {
			for _, field := range definition.Schema.Fields {
				if field.Name == "per_send_cap" {
					*field.Schema.Maximum = 0
				}
			}
		}
	}
	if !reflect.DeepEqual(before, MinigameSchemas()) {
		t.Fatal("caller mutated owner schema")
	}
}

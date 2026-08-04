package achievements

import (
	"encoding/json"
	"os"
	"testing"
)

type catalogParityCorpus struct {
	Version int `json:"version"`
	Cases   []struct {
		Name       string `json:"name"`
		Valid      bool   `json:"valid"`
		Operations []struct {
			Op    string `json:"op"`
			Path  []any  `json:"path"`
			Value any    `json:"value"`
		} `json:"operations"`
	} `json:"cases"`
}

func testRegistry() Registry {
	return Registry{
		CopyKeys:          map[string]bool{"achievement.first_gate": true, "achievement.possession_warning": true},
		GeneratorIDs:      map[string]bool{"generator.clickfarm": true},
		EventKinds:        map[string]bool{"gate_crossed": true, "generator_purchased": true},
		ResourceIDs:       map[string]bool{"company.cash": true},
		RunCounters:       map[string]bool{"generators_purchased_total": true, "tier": true},
		CareerCounters:    map[string]bool{"age_ms": true, "notoriety": true},
		ProvenanceSources: map[string][]string{"fact:gate.tier_1": {"gate_crossed"}},
	}
}

func validCatalogBytes(t *testing.T) []byte {
	t.Helper()
	root := map[string]any{"schema_version": 1, "achievements": []any{
		map[string]any{"id": "achievement.first_gate", "condition_scope": "run", "condition": map[string]any{"kind": "fact_present", "fact_kind": "gate.tier_1"}, "proof": map[string]any{"kind": "provenance", "event_kinds": []string{"gate_crossed"}}, "score_grant": 4, "copy_key": "achievement.first_gate"},
		map[string]any{"id": "achievement.generator_hoard", "condition_scope": "run", "condition": map[string]any{"kind": "owns_generator_at_least", "generator_id": "generator.clickfarm", "count": 300}, "proof": map[string]any{"kind": "possession", "justification_copy_key": "achievement.possession_warning"}, "score_grant": 8, "copy_key": "achievement.first_gate"},
	}}
	data, err := json.Marshal(root)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func mutateCatalog(t *testing.T, mutate func(map[string]any)) []byte {
	t.Helper()
	var root map[string]any
	if json.Unmarshal(validCatalogBytes(t), &root) != nil {
		t.Fatal("decode fixture")
	}
	mutate(root)
	data, err := json.Marshal(root)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestLoadCatalogAndScore(t *testing.T) {
	catalog, err := LoadCatalog(validCatalogBytes(t), testRegistry())
	if err != nil {
		t.Fatal(err)
	}
	if len(catalog.Definitions) != 2 || catalog.Definitions[0].ScoreGrant != 4 || catalog.Definitions[1].Proof.Kind != ProofPossession {
		t.Fatalf("unexpected catalog: %+v", catalog.Definitions)
	}
	score, err := catalog.Score(map[string]bool{"achievement.first_gate": true, "achievement.generator_hoard": true})
	if err != nil || score != 12 {
		t.Fatalf("score=%d err=%v", score, err)
	}
	if _, err := catalog.Score(map[string]bool{"achievement.unknown": true}); err == nil {
		t.Fatal("unknown earned ID accepted")
	}
}

func TestLoadCatalogRejectsProofAndRegistryDrift(t *testing.T) {
	tests := map[string]func(map[string]any){
		"unsorted IDs": func(root map[string]any) { rows := root["achievements"].([]any); rows[0], rows[1] = rows[1], rows[0] },
		"bare possession": func(root map[string]any) {
			root["achievements"].([]any)[1].(map[string]any)["proof"] = map[string]any{"kind": "possession", "justification_copy_key": "missing.copy"}
		},
		"possession without ownership": func(root map[string]any) {
			root["achievements"].([]any)[0].(map[string]any)["proof"] = map[string]any{"kind": "possession", "justification_copy_key": "achievement.possession_warning"}
		},
		"provenance possession": func(root map[string]any) {
			root["achievements"].([]any)[1].(map[string]any)["proof"] = map[string]any{"kind": "provenance", "event_kinds": []string{"generator_purchased"}}
		},
		"unrelated provenance": func(root map[string]any) {
			root["achievements"].([]any)[0].(map[string]any)["proof"] = map[string]any{"kind": "provenance", "event_kinds": []string{"generator_purchased"}}
		},
		"unknown copy": func(root map[string]any) {
			root["achievements"].([]any)[0].(map[string]any)["copy_key"] = "missing.copy"
		},
		"unknown field": func(root map[string]any) { root["achievements"].([]any)[0].(map[string]any)["clout_grant_ppm"] = 4 },
		"wrong condition null field": func(root map[string]any) {
			root["achievements"].([]any)[0].(map[string]any)["condition"].(map[string]any)["minimum"] = nil
		},
		"wrong proof empty field": func(root map[string]any) {
			root["achievements"].([]any)[0].(map[string]any)["proof"].(map[string]any)["event_kind"] = ""
		},
		"career ownership": func(root map[string]any) {
			root["achievements"].([]any)[1].(map[string]any)["condition_scope"] = "career"
		},
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			if _, err := LoadCatalog(mutateCatalog(t, mutate), testRegistry()); err == nil {
				t.Fatal("invalid catalog accepted")
			}
		})
	}
	invalidRegistry := testRegistry()
	invalidRegistry.EventKinds["INVALID EVENT"] = true
	if _, err := LoadCatalog(validCatalogBytes(t), invalidRegistry); err == nil {
		t.Fatal("non-mechanical registry key accepted")
	}
}

func TestLoadCatalogMatchesSharedParityCorpus(t *testing.T) {
	data, err := os.ReadFile("../../balance/testdata/achievements-catalog-parity-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	var corpus catalogParityCorpus
	if err := json.Unmarshal(data, &corpus); err != nil || corpus.Version != 1 || len(corpus.Cases) == 0 {
		t.Fatalf("invalid parity corpus: version=%d cases=%d err=%v", corpus.Version, len(corpus.Cases), err)
	}
	for _, vector := range corpus.Cases {
		t.Run(vector.Name, func(t *testing.T) {
			var root any
			if err := json.Unmarshal(validCatalogBytes(t), &root); err != nil {
				t.Fatal(err)
			}
			for _, operation := range vector.Operations {
				applyCatalogParityOperation(t, root, operation.Op, operation.Path, operation.Value)
			}
			candidate, err := json.Marshal(root)
			if err != nil {
				t.Fatal(err)
			}
			_, loadErr := LoadCatalog(candidate, testRegistry())
			if (loadErr == nil) != vector.Valid {
				t.Fatalf("valid=%v loadErr=%v", vector.Valid, loadErr)
			}
		})
	}
}

func applyCatalogParityOperation(t *testing.T, root any, operation string, path []any, value any) {
	t.Helper()
	if len(path) == 0 {
		t.Fatal("empty parity operation path")
	}
	current := root
	for _, component := range path[:len(path)-1] {
		switch typed := component.(type) {
		case string:
			current = current.(map[string]any)[typed]
		case float64:
			current = current.([]any)[int(typed)]
		default:
			t.Fatalf("invalid parity path component %T", component)
		}
	}
	last := path[len(path)-1]
	switch typed := last.(type) {
	case string:
		object := current.(map[string]any)
		switch operation {
		case "delete":
			delete(object, typed)
		case "set":
			object[typed] = value
		default:
			t.Fatalf("invalid parity operation %q", operation)
		}
	case float64:
		if operation != "set" {
			t.Fatalf("operation %q is invalid for array path", operation)
		}
		current.([]any)[int(typed)] = value
	default:
		t.Fatalf("invalid final parity path component %T", last)
	}
}

func TestPredicateEvaluationAndLifetimeLatch(t *testing.T) {
	catalog, err := LoadCatalog(validCatalogBytes(t), testRegistry())
	if err != nil {
		t.Fatal(err)
	}
	run := Observation{Facts: map[string]bool{"gate.tier_1": true}, Counters: map[string]int64{}, Generators: map[string]int64{"generator.clickfarm": 300}}
	newly, err := catalog.NewlyEarned(map[string]bool{}, map[string]bool{}, run, Observation{})
	if err != nil || len(newly) != 2 || newly[0].ID != "achievement.first_gate" || newly[1].ID != "achievement.generator_hoard" {
		t.Fatalf("new=%+v err=%v", newly, err)
	}
	newly, err = catalog.NewlyEarned(map[string]bool{"achievement.first_gate": true}, map[string]bool{"achievement.generator_hoard": true}, run, Observation{})
	if err != nil || len(newly) != 0 {
		t.Fatalf("latched=%+v err=%v", newly, err)
	}
}

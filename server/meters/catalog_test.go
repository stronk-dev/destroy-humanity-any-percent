package meters

import (
	"encoding/json"
	"os"
	"testing"
)

type catalogParityCorpus struct {
	Version  int             `json:"version"`
	Baseline json.RawMessage `json:"baseline"`
	Cases    []struct {
		Name       string `json:"name"`
		Valid      bool   `json:"valid"`
		Operations []struct {
			Op    string `json:"op"`
			Path  []any  `json:"path"`
			Value any    `json:"value"`
		} `json:"operations"`
	} `json:"cases"`
}

func validCatalogBytes(t *testing.T) []byte {
	t.Helper()
	meters := make([]map[string]any, 0, len(requiredMeterIDs))
	for _, id := range requiredMeterIDs {
		meters = append(meters, map[string]any{
			"id": id, "scope": "company", "min_value": 0, "max_value": 100, "initial_value": 50,
			"bands":  []any{map[string]any{"id": "low", "floor_value": 0}, map[string]any{"id": "high", "floor_value": 70}},
			"inputs": []any{}, "decay": map[string]any{"toward_value": 50, "rate_per_attended_hour": 2},
		})
	}
	meters[0]["inputs"] = []any{
		map[string]any{"kind": "ledger_fact", "fact_kind": "externality.emitted", "delta": 3},
		map[string]any{"kind": "contribution_slot", "slot": "upgrades", "source_id": "generator.example", "delta_per_attended_hour": -2},
	}
	root := map[string]any{
		"schema_version": 1,
		"trust_reseed":   map[string]any{"base_value": 90, "notoriety_numerator": 35, "notoriety_denominator": 100, "floor_value": 55, "ceiling_value": 90},
		"meters":         meters,
	}
	data, err := json.Marshal(root)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func mutateCatalog(t *testing.T, operation func(map[string]any)) []byte {
	t.Helper()
	var root map[string]any
	if err := json.Unmarshal(validCatalogBytes(t), &root); err != nil {
		t.Fatal(err)
	}
	operation(root)
	data, err := json.Marshal(root)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestLoadCatalogRequiresClosedPhaseASetAndInputs(t *testing.T) {
	catalog, err := LoadCatalog(validCatalogBytes(t))
	if err != nil {
		t.Fatal(err)
	}
	if ids := catalog.MeterIDs(); len(ids) != 11 || ids[0] != "doom.probability" || ids[10] != "trust.users.standing" {
		t.Fatalf("unexpected IDs: %v", ids)
	}
	doom, ok := catalog.Meter("doom.probability")
	if !ok || len(doom.Inputs) != 2 || doom.Inputs[0].Kind != InputLedgerFact || doom.Inputs[1].Kind != InputContributionSlot {
		t.Fatalf("unexpected doom meter: %+v", doom)
	}
	if err := catalog.ValidateResourceSeparation([]string{"company.cash", "company.value"}); err != nil {
		t.Fatal(err)
	}
	if err := catalog.ValidateResourceSeparation([]string{"company.cash", "trust.users.standing"}); err == nil {
		t.Fatal("meter/economy ID collision accepted")
	}
}

func TestLoadCatalogRejectsShapeAndSemanticDrift(t *testing.T) {
	tests := map[string]func(map[string]any){
		"missing axis": func(root map[string]any) {
			root["meters"] = root["meters"].([]any)[:10]
		},
		"public axis": func(root map[string]any) {
			root["meters"].([]any)[9].(map[string]any)["id"] = "trust.public.grievance"
		},
		"externality meter": func(root map[string]any) {
			root["meters"].([]any)[0].(map[string]any)["id"] = "externality.total"
		},
		"unsorted band": func(root map[string]any) {
			root["meters"].([]any)[0].(map[string]any)["bands"] = []any{map[string]any{"id": "low", "floor_value": 0}, map[string]any{"id": "high", "floor_value": 0}}
		},
		"duplicate source": func(root map[string]any) {
			input := map[string]any{"kind": "ledger_fact", "fact_kind": "externality.emitted", "delta": 1}
			root["meters"].([]any)[0].(map[string]any)["inputs"] = []any{input, input}
		},
		"zero input": func(root map[string]any) {
			root["meters"].([]any)[0].(map[string]any)["inputs"] = []any{map[string]any{"kind": "ledger_fact", "fact_kind": "externality.emitted", "delta": 0}}
		},
		"missing zero initial": func(root map[string]any) {
			delete(root["meters"].([]any)[0].(map[string]any), "initial_value")
		},
		"missing zero band floor": func(root map[string]any) {
			delete(root["meters"].([]any)[0].(map[string]any)["bands"].([]any)[0].(map[string]any), "floor_value")
		},
		"missing nullable decay": func(root map[string]any) {
			delete(root["meters"].([]any)[0].(map[string]any), "decay")
		},
		"reseed above shared exact range": func(root map[string]any) {
			root["trust_reseed"].(map[string]any)["notoriety_denominator"] = float64(9_007_199_254_740_992)
		},
		"wrong union empty field": func(root map[string]any) {
			root["meters"].([]any)[0].(map[string]any)["inputs"].([]any)[0].(map[string]any)["slot"] = ""
		},
		"unknown field": func(root map[string]any) {
			root["meters"].([]any)[0].(map[string]any)["spendable"] = false
		},
	}
	for name, operation := range tests {
		t.Run(name, func(t *testing.T) {
			if _, err := LoadCatalog(mutateCatalog(t, operation)); err == nil {
				t.Fatal("invalid catalog accepted")
			}
		})
	}
	if _, err := LoadCatalog(append(validCatalogBytes(t), []byte(` {}`)...)); err == nil {
		t.Fatal("trailing JSON accepted")
	}
}

func TestLoadCatalogMatchesSharedParityCorpus(t *testing.T) {
	data, err := os.ReadFile("../../balance/testdata/meters-catalog-parity-v1.json")
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
			if err := json.Unmarshal(corpus.Baseline, &root); err != nil {
				t.Fatal(err)
			}
			for _, operation := range vector.Operations {
				applyCatalogParityOperation(t, root, operation.Op, operation.Path, operation.Value)
			}
			candidate, err := json.Marshal(root)
			if err != nil {
				t.Fatal(err)
			}
			_, loadErr := LoadCatalog(candidate)
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

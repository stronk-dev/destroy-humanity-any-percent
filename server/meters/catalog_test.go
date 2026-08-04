package meters

import (
	"encoding/json"
	"testing"
)

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

package harness

import (
	"encoding/json"
	"os"
	"testing"

	"cloud-clicker/server/economy"
	"cloud-clicker/server/routes"
)

func relevancePolicyDependencies(t *testing.T) ([]byte, *economy.Catalog, *routes.Catalog) {
	t.Helper()
	policyBytes, err := os.ReadFile("../../testdata/harness/relevance/policy-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	catalogBytes, err := os.ReadFile("../../testdata/economy-foundation-v4.json")
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := economy.LoadCatalog(catalogBytes)
	if err != nil {
		t.Fatal(err)
	}
	routesBytes, err := os.ReadFile("../../balance/routes/phase0.json")
	if err != nil {
		t.Fatal(err)
	}
	routeCatalog, err := routes.LoadCatalog(routesBytes)
	if err != nil {
		t.Fatal(err)
	}
	return policyBytes, catalog, routeCatalog
}

func TestLoadRelevancePolicyExactAndComplete(t *testing.T) {
	data, catalog, routeCatalog := relevancePolicyDependencies(t)
	policy, err := LoadRelevancePolicy(data, catalog, routeCatalog)
	if err != nil {
		t.Fatal(err)
	}
	if len(policy.Items) != 3 || len(policy.Groups) != 4 || len(policy.Hash) != 71 {
		t.Fatalf("policy=%+v", policy)
	}
}

func TestLoadRelevancePolicyRejectsWireAndSemanticMutations(t *testing.T) {
	data, catalog, routeCatalog := relevancePolicyDependencies(t)
	var base map[string]any
	if err := json.Unmarshal(data, &base); err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name   string
		mutate func(map[string]any)
	}{
		{name: "missing item key", mutate: func(value map[string]any) { delete(value["items"].([]any)[0].(map[string]any), "epsilon_ms") }},
		{name: "unsorted items", mutate: func(value map[string]any) { rows := value["items"].([]any); rows[0], rows[1] = rows[1], rows[0] }},
		{name: "dangling group", mutate: func(value map[string]any) {
			value["items"].([]any)[0].(map[string]any)["group_ids"] = []any{"group.missing"}
		}},
		{name: "exemption mismatch", mutate: func(value map[string]any) { value["items"].([]any)[2].(map[string]any)["trap_exempt"] = false }},
		{name: "incomplete derived group", mutate: func(value map[string]any) {
			value["groups"].([]any)[0].(map[string]any)["member_ids"] = []any{"generator.high"}
		}},
		{name: "unknown gate", mutate: func(value map[string]any) {
			value["items"].([]any)[0].(map[string]any)["availability_window"].(map[string]any)["from_gate"] = "gate.unknown"
		}},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			var value map[string]any
			encoded, _ := json.Marshal(base)
			_ = json.Unmarshal(encoded, &value)
			test.mutate(value)
			mutated, _ := json.Marshal(value)
			if _, err := LoadRelevancePolicy(mutated, catalog, routeCatalog); err == nil {
				t.Fatal("mutation accepted")
			}
		})
	}
}

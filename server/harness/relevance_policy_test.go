package harness

import (
	"bytes"
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
	mutationBytes, err := os.ReadFile("../../testdata/harness/relevance/policy-mutations-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	var corpus struct {
		SchemaVersion int `json:"schema_version"`
		Cases         []struct {
			ID       string `json:"id"`
			Accepted bool   `json:"accepted"`
		} `json:"cases"`
	}
	if err := json.Unmarshal(mutationBytes, &corpus); err != nil || corpus.SchemaVersion != 1 || len(corpus.Cases) == 0 {
		t.Fatal("invalid relevance mutation corpus")
	}
	for _, test := range corpus.Cases {
		t.Run(test.ID, func(t *testing.T) {
			if test.ID == "integral_decimal_epsilon" {
				mutated := bytes.Replace(data, []byte(`"epsilon_ms": 1000`), []byte(`"epsilon_ms": 1000.0`), 1)
				if _, err := LoadRelevancePolicy(mutated, catalog, routeCatalog); (err == nil) != test.Accepted {
					t.Fatalf("accepted=%v err=%v", err == nil, err)
				}
				return
			}
			if test.ID == "integral_decimal_schema" {
				mutated := bytes.Replace(data, []byte(`"schema_version": 1`), []byte(`"schema_version": 1.0`), 1)
				if _, err := LoadRelevancePolicy(mutated, catalog, routeCatalog); (err == nil) != test.Accepted {
					t.Fatalf("accepted=%v err=%v", err == nil, err)
				}
				return
			}
			var value map[string]any
			_ = json.Unmarshal(data, &value)
			switch test.ID {
			case "missing_item_epsilon":
				delete(value["items"].([]any)[0].(map[string]any), "epsilon_ms")
			case "unsorted_items":
				rows := value["items"].([]any)
				rows[0], rows[1] = rows[1], rows[0]
			case "dangling_group":
				value["items"].([]any)[0].(map[string]any)["group_ids"] = []any{"group.missing"}
			case "exemption_mismatch":
				value["items"].([]any)[2].(map[string]any)["trap_exempt"] = false
			case "incomplete_derived_group":
				value["groups"].([]any)[0].(map[string]any)["member_ids"] = []any{"generator.high"}
			case "unknown_gate":
				value["items"].([]any)[0].(map[string]any)["availability_window"].(map[string]any)["from_gate"] = "gate.unknown"
			default:
				t.Fatalf("unknown mutation %q", test.ID)
			}
			mutated, _ := json.Marshal(value)
			if _, err := LoadRelevancePolicy(mutated, catalog, routeCatalog); (err == nil) != test.Accepted {
				t.Fatalf("accepted=%v err=%v", err == nil, err)
			}
		})
	}
}

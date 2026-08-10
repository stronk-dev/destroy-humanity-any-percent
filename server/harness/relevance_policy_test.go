package harness

import (
	"bytes"
	"encoding/json"
	"os"
	"strconv"
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

func TestRelevancePolicyV2AcceptsRunGenesisWithoutWeakeningV1(t *testing.T) {
	suite, err := LoadRelevanceSuite("../..", "testdata/harness/relevance/scenario-v2.json")
	if err != nil || suite.Policy.SchemaVersion != 2 || suite.Policy.Items[0].Availability.FromGate != nil {
		t.Fatalf("v2 suite=%+v err=%v", suite, err)
	}
	v2, err := os.ReadFile("../../testdata/harness/relevance/fixture-policy-v2.json")
	if err != nil {
		t.Fatal(err)
	}
	v1Null := bytes.Replace(v2, []byte(`"schema_version": 2`), []byte(`"schema_version": 1`), 1)
	if _, err := LoadRelevancePolicy(v1Null, suite.Catalog, suite.Routes); err == nil {
		t.Fatal("schema-v1 policy accepted a null from_gate")
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
			ID        string   `json:"id"`
			Operation string   `json:"operation"`
			Path      []string `json:"path"`
			ValueJSON *string  `json:"value_json"`
			SwapIndex *int     `json:"swap_index"`
			Accepted  bool     `json:"accepted"`
		} `json:"cases"`
	}
	if err := json.Unmarshal(mutationBytes, &corpus); err != nil || corpus.SchemaVersion != 1 || len(corpus.Cases) == 0 {
		t.Fatal("invalid relevance mutation corpus")
	}
	for _, test := range corpus.Cases {
		t.Run(test.ID, func(t *testing.T) {
			mutated := applyRelevanceMutation(t, data, test.Operation, test.Path, test.ValueJSON, test.SwapIndex)
			if _, err := LoadRelevancePolicy(mutated, catalog, routeCatalog); (err == nil) != test.Accepted {
				t.Fatalf("accepted=%v err=%v", err == nil, err)
			}
		})
	}
}

func applyRelevanceMutation(t *testing.T, data []byte, operation string, path []string, valueJSON *string, swapIndex *int) []byte {
	t.Helper()
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	var root any
	if err := decoder.Decode(&root); err != nil {
		t.Fatal(err)
	}
	if len(path) == 0 {
		t.Fatal("mutation path is empty")
	}
	parent := root
	for _, component := range path[:len(path)-1] {
		parent = relevanceMutationChild(t, parent, component)
	}
	last := path[len(path)-1]
	switch operation {
	case "delete":
		object, ok := parent.(map[string]any)
		if !ok {
			t.Fatal("delete parent is not an object")
		}
		delete(object, last)
	case "replace", "replace_number":
		if valueJSON == nil {
			t.Fatal("replace mutation has no value_json")
		}
		valueDecoder := json.NewDecoder(bytes.NewBufferString(*valueJSON))
		valueDecoder.UseNumber()
		var replacement any
		if err := valueDecoder.Decode(&replacement); err != nil {
			t.Fatal(err)
		}
		relevanceMutationSet(t, parent, last, replacement)
	case "swap":
		rows, ok := parent.([]any)
		if !ok || swapIndex == nil {
			t.Fatal("swap mutation is invalid")
		}
		index, err := strconv.Atoi(last)
		if err != nil || index < 0 || index >= len(rows) || *swapIndex < 0 || *swapIndex >= len(rows) {
			t.Fatal("swap index is invalid")
		}
		rows[index], rows[*swapIndex] = rows[*swapIndex], rows[index]
	case "copy":
		rows, ok := parent.([]any)
		if !ok || swapIndex == nil {
			t.Fatal("copy mutation is invalid")
		}
		index, err := strconv.Atoi(last)
		if err != nil || index < 0 || index >= len(rows) || *swapIndex < 0 || *swapIndex >= len(rows) {
			t.Fatal("copy index is invalid")
		}
		rows[*swapIndex] = rows[index]
	default:
		t.Fatalf("unknown mutation operation %q", operation)
	}
	result, err := json.Marshal(root)
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func relevanceMutationChild(t *testing.T, value any, component string) any {
	t.Helper()
	if object, ok := value.(map[string]any); ok {
		child, present := object[component]
		if !present {
			t.Fatalf("missing mutation path component %q", component)
		}
		return child
	}
	rows, ok := value.([]any)
	index, err := strconv.Atoi(component)
	if !ok || err != nil || index < 0 || index >= len(rows) {
		t.Fatalf("invalid mutation path component %q", component)
	}
	return rows[index]
}

func relevanceMutationSet(t *testing.T, parent any, component string, replacement any) {
	t.Helper()
	if object, ok := parent.(map[string]any); ok {
		object[component] = replacement
		return
	}
	rows, ok := parent.([]any)
	index, err := strconv.Atoi(component)
	if !ok || err != nil || index < 0 || index >= len(rows) {
		t.Fatalf("invalid mutation target %q", component)
	}
	rows[index] = replacement
}

package doctrine

import (
	"encoding/json"
	"os"
	"testing"

	"cloud-clicker/server/routes"
)

type parityCorpus struct {
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

func TestLoadCatalogMatchesSharedParityCorpus(t *testing.T) {
	data, err := os.ReadFile("../../balance/testdata/doctrines-catalog-parity-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	var corpus parityCorpus
	if err := json.Unmarshal(data, &corpus); err != nil || corpus.Version != 1 || len(corpus.Cases) == 0 {
		t.Fatalf("invalid corpus: version=%d cases=%d err=%v", corpus.Version, len(corpus.Cases), err)
	}
	for _, vector := range corpus.Cases {
		t.Run(vector.Name, func(t *testing.T) {
			var root any
			if err := json.Unmarshal(corpus.Baseline, &root); err != nil {
				t.Fatal(err)
			}
			for _, operation := range vector.Operations {
				applyOperation(t, root, operation.Op, operation.Path, operation.Value)
			}
			candidate, err := json.Marshal(root)
			if err != nil {
				t.Fatal(err)
			}
			_, loadErr := LoadCatalog(candidate)
			if (loadErr == nil) != vector.Valid {
				t.Fatalf("valid=%v err=%v", vector.Valid, loadErr)
			}
		})
	}
}

func TestCatalogValidatesRouteDoctrineReferences(t *testing.T) {
	doctrineBytes := []byte(`{"schema_version":1,"transitions":[{"transition_id":"transition.t3_to_t4","source_tier":3,"gate_id":"gate.t3_to_t4","doctrine_ids":["doctrine.capture","doctrine.ethical"]}]}`)
	catalog, err := LoadCatalog(doctrineBytes)
	if err != nil {
		t.Fatal(err)
	}
	routeBytes := []byte(`{"schema_version":1,"context_version":1,"depletion_distinct_routes_required":2,"knowledge":{"registry_first_bonus":100,"founder_first_grant":25,"repeat_grant":5,"hint_cost":50},"gates":[{"gate_id":"gate.t3_to_t4","requirement":[{"resource_id":"company.cash","amount":"1e12"}],"routes":[]},{"gate_id":"gate.t4_to_t5","requirement":[{"resource_id":"company.cash","amount":"1e15"}],"routes":[{"route_id":"route.capture","house_name":"Capture","active":true,"requires_context_version":1,"exclusion_slot":"doctrine:transition.t3_to_t4","exclusion_value":"doctrine.capture","predicate":[{"kind":"doctrine_is","transition":"transition.t3_to_t4","doctrine_id":"doctrine.capture"}],"effect":{"kind":"discount","fraction":"5e-1"}}]}]}`)
	routeCatalog, err := routes.LoadCatalog(routeBytes)
	if err != nil {
		t.Fatal(err)
	}
	if err := catalog.ValidateRoutes(routeCatalog); err != nil {
		t.Fatal(err)
	}
	var raw map[string]any
	if err := json.Unmarshal(routeBytes, &raw); err != nil {
		t.Fatal(err)
	}
	badRoute := raw["gates"].([]any)[1].(map[string]any)["routes"].([]any)[0].(map[string]any)
	badRoute["exclusion_value"] = "doctrine.missing"
	badRoute["predicate"].([]any)[0].(map[string]any)["doctrine_id"] = "doctrine.missing"
	bad, _ := json.Marshal(raw)
	badRoutes, err := routes.LoadCatalog(bad)
	if err != nil {
		t.Fatal(err)
	}
	if err := catalog.ValidateRoutes(badRoutes); err == nil {
		t.Fatal("undeclared route doctrine accepted")
	}
}

func applyOperation(t *testing.T, root any, operation string, path []any, value any) {
	t.Helper()
	current := root
	for _, component := range path[:len(path)-1] {
		switch typed := component.(type) {
		case string:
			current = current.(map[string]any)[typed]
		case float64:
			current = current.([]any)[int(typed)]
		}
	}
	last := path[len(path)-1]
	switch typed := last.(type) {
	case string:
		object := current.(map[string]any)
		if operation == "delete" {
			delete(object, typed)
		} else {
			object[typed] = value
		}
	case float64:
		current.([]any)[int(typed)] = value
	}
}

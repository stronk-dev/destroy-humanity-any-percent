package doctrine

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"cloud-clicker/server/routes"
)

func doctrineFixture(t *testing.T) []byte {
	t.Helper()
	data, err := os.ReadFile("../../balance/testdata/doctrines-catalog-parity-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	var corpus parityCorpus
	if err := json.Unmarshal(data, &corpus); err != nil || corpus.Version != 1 || len(corpus.Baseline) == 0 {
		t.Fatalf("invalid fixture: %v", err)
	}
	return corpus.Baseline
}

func editDoctrineFixture(t *testing.T, edit func(map[string]any)) []byte {
	t.Helper()
	var root map[string]any
	if err := json.Unmarshal(doctrineFixture(t), &root); err != nil {
		t.Fatal(err)
	}
	edit(root)
	data, err := json.Marshal(root)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestLoadCatalogRejectsCaseAliasedMembers(t *testing.T) {
	for _, key := range []string{"schema_version", "transitions", "transition_id", "source_tier", "gate_id", "doctrine_ids"} {
		for _, alias := range []string{strings.ToUpper(key), strings.ToUpper(key[:1]) + key[1:]} {
			t.Run(alias, func(t *testing.T) {
				data := editDoctrineFixture(t, func(root map[string]any) {
					object := root
					if key != "schema_version" && key != "transitions" {
						object = root["transitions"].([]any)[0].(map[string]any)
					}
					object[alias] = object[key]
					delete(object, key)
				})
				if _, err := LoadCatalog(data); err == nil {
					t.Fatal("case alias admitted")
				}
			})
		}
	}
}

func TestLoadCatalogRequiresExplicitNonNullZeroSourceTier(t *testing.T) {
	for _, missing := range []bool{false, true} {
		data := editDoctrineFixture(t, func(root map[string]any) {
			row := root["transitions"].([]any)[0].(map[string]any)
			row["transition_id"], row["gate_id"], row["source_tier"] = "transition.t0_to_t1", "gate.t0_to_t1", nil
			if missing {
				delete(row, "source_tier")
			}
		})
		if _, err := LoadCatalog(data); err == nil {
			t.Errorf("missing=%t: absent/null source_tier silently became zero", missing)
		}
	}
	data := editDoctrineFixture(t, func(root map[string]any) {
		row := root["transitions"].([]any)[0].(map[string]any)
		row["transition_id"], row["gate_id"], row["source_tier"] = "transition.t0_to_t1", "gate.t0_to_t1", 0
	})
	if _, err := LoadCatalog(data); err != nil {
		t.Fatalf("explicit zero refused: %v", err)
	}
}

func TestLoadCatalogRejectsFinalNullSchemaVersion(t *testing.T) {
	source := bytes.TrimSpace(doctrineFixture(t))
	data := append(bytes.Clone(source[:len(source)-1]), []byte(`,"schema_version":null}`)...)
	if _, err := LoadCatalog(data); err == nil {
		t.Fatal("final null version inherited the earlier schema version")
	}
}

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

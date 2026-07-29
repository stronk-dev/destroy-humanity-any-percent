package routes

import (
	"encoding/json"
	"os"
	"testing"

	"cloud-clicker/server/decimal"
)

func loadPhase0(t *testing.T) (*Catalog, []byte) {
	t.Helper()
	data, err := os.ReadFile("../../balance/routes/phase0.json")
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := LoadCatalog(data)
	if err != nil {
		t.Fatal(err)
	}
	return catalog, data
}

func TestCatalogLoadsAndProvesDepletionUnreachable(t *testing.T) {
	catalog, _ := loadPhase0(t)
	if got := catalog.MaxRoutesPerRun(); got != 4 {
		t.Fatalf("maximum routes = %d", got)
	}
	if catalog.DepletionDistinctRequired() != 5 {
		t.Fatalf("depletion count = %d", catalog.DepletionDistinctRequired())
	}
	if len(catalog.Gates()) != 3 {
		t.Fatalf("gates = %d", len(catalog.Gates()))
	}
	policy := catalog.KnowledgePolicy()
	if policy.RegistryFirstBonus != 100 || policy.FounderFirstGrant != 25 || policy.RepeatGrant != 5 || policy.HintCost != 50 {
		t.Fatalf("knowledge = %+v", policy)
	}
}

func TestCatalogRejectsReachableDepletionAndUnavailableActiveRoute(t *testing.T) {
	_, data := loadPhase0(t)
	var root map[string]any
	if err := json.Unmarshal(data, &root); err != nil {
		t.Fatal(err)
	}
	root["depletion_distinct_routes_required"] = float64(4)
	mutated, _ := json.Marshal(root)
	if _, err := LoadCatalog(mutated); err == nil {
		t.Fatal("single-run-reachable catalog accepted")
	}

	if err := json.Unmarshal(data, &root); err != nil {
		t.Fatal(err)
	}
	gates := root["gates"].([]any)
	routes := gates[1].(map[string]any)["routes"].([]any)
	routes[0].(map[string]any)["active"] = true
	mutated, _ = json.Marshal(root)
	if _, err := LoadCatalog(mutated); err == nil {
		t.Fatal("active route with unavailable context accepted")
	}
	if err := json.Unmarshal(data, &root); err != nil {
		t.Fatal(err)
	}
	gates = root["gates"].([]any)
	routes = gates[1].(map[string]any)["routes"].([]any)
	routes[0].(map[string]any)["requires_context_version"] = float64(1)
	mutated, _ = json.Marshal(root)
	if _, err := LoadCatalog(mutated); err == nil {
		t.Fatal("meter route lied about its minimum context version")
	}
}

func TestSharedCatalogFixtures(t *testing.T) {
	valid, err := os.ReadFile("../../balance/routes-testdata/valid/minimal.json")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := LoadCatalog(valid); err != nil {
		t.Fatalf("valid fixture: %v", err)
	}
	for _, filename := range []string{"reachable-depletion.json", "temporal-impossibility.json", "unknown-field.json", "unbound-exclusion.json"} {
		data, err := os.ReadFile("../../balance/routes-testdata/invalid/" + filename)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := LoadCatalog(data); err == nil {
			t.Fatalf("invalid fixture %s accepted", filename)
		}
	}
}

func TestCatalogAcceptsDoctrineRouteAtSameOrLaterBoundary(t *testing.T) {
	_, data := loadPhase0(t)
	var root map[string]any
	if err := json.Unmarshal(data, &root); err != nil {
		t.Fatal(err)
	}
	gates := root["gates"].([]any)
	gates[1].(map[string]any)["gate_id"] = "gate.t3_to_t4"
	mutated, _ := json.Marshal(root)
	if _, err := LoadCatalog(mutated); err != nil {
		t.Fatalf("same-boundary doctrine route rejected: %v", err)
	}
}

func TestRouteResolutionDiscountAndSubstitute(t *testing.T) {
	catalog, _ := loadPhase0(t)
	context := Context{
		ContextVersion:        1,
		Resources:             map[string]decimal.Decimal{"company.cash": decimal.New(4, 8)},
		DoctrinesByTransition: map[string]string{"transition.t3_to_t4": "doctrine.capture"},
		LedgerFactKinds:       map[string]bool{},
	}
	discount, matched, err := catalog.Resolve("gate.t4_to_t5", "route.ipo_sequence_break", context)
	if err != nil || !matched || len(discount.Requirement) != 1 || discount.Requirement[0].Amount.String() != "4e14" {
		t.Fatalf("discount=%+v matched=%v err=%v", discount, matched, err)
	}
	context.StructureID = "structure.nonprofit"
	substitute, matched, err := catalog.Resolve("gate.t4_to_t5", "route.nonprofit_wrapper_zip", context)
	if err != nil || !matched || len(substitute.Requirement) != 0 {
		t.Fatalf("substitute=%+v matched=%v err=%v", substitute, matched, err)
	}
	standard, matched, err := catalog.Resolve("gate.t2_to_t3", "", context)
	if err != nil || !matched || standard.RouteID != "" || standard.Requirement[0].Amount.String() != "1e9" {
		t.Fatalf("standard=%+v matched=%v err=%v", standard, matched, err)
	}
}

type predicateFixture struct {
	SchemaVersion int `json:"schema_version"`
	Vectors       []struct {
		Name      string          `json:"name"`
		Condition json.RawMessage `json:"condition"`
		Context   struct {
			ContextVersion int               `json:"context_version"`
			Resources      map[string]string `json:"resources"`
			Doctrines      map[string]string `json:"doctrines_by_transition"`
			StructureID    string            `json:"structure_id"`
			Facts          []string          `json:"ledger_fact_kinds"`
			Meters         map[string]int    `json:"meter_bands"`
			Traits         []string          `json:"region_traits"`
		} `json:"context"`
		Expected bool `json:"expected"`
	} `json:"vectors"`
}

func TestSharedPredicateVectors(t *testing.T) {
	data, err := os.ReadFile("../../testdata/routes/predicate-vectors.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture predicateFixture
	if err := decodeStrict(data, &fixture); err != nil {
		t.Fatal(err)
	}
	for _, vector := range fixture.Vectors {
		t.Run(vector.Name, func(t *testing.T) {
			var sources []rawCondition
			if len(vector.Condition) > 0 && vector.Condition[0] == '[' {
				if err := json.Unmarshal(vector.Condition, &sources); err != nil {
					t.Fatal(err)
				}
			} else {
				var source rawCondition
				if err := json.Unmarshal(vector.Condition, &source); err != nil {
					t.Fatal(err)
				}
				sources = []rawCondition{source}
			}
			predicate := make([]Condition, 0, len(sources))
			for _, source := range sources {
				condition, err := parseCondition(source)
				if err != nil {
					t.Fatal(err)
				}
				predicate = append(predicate, condition)
			}
			resources := make(map[string]decimal.Decimal, len(vector.Context.Resources))
			for id, raw := range vector.Context.Resources {
				value, err := decimal.ParseCanonical(raw)
				if err != nil {
					t.Fatal(err)
				}
				resources[id] = value
			}
			facts := make(map[string]bool, len(vector.Context.Facts))
			for _, fact := range vector.Context.Facts {
				facts[fact] = true
			}
			traits := make(map[string]bool, len(vector.Context.Traits))
			for _, trait := range vector.Context.Traits {
				traits[trait] = true
			}
			got, err := EvaluatePredicate(predicate, Context{
				ContextVersion: vector.Context.ContextVersion, Resources: resources,
				DoctrinesByTransition: vector.Context.Doctrines, StructureID: vector.Context.StructureID,
				LedgerFactKinds: facts, MeterBands: vector.Context.Meters, RegionTraits: traits,
			})
			if err != nil || got != vector.Expected {
				t.Fatalf("got=%v want=%v err=%v", got, vector.Expected, err)
			}
		})
	}
}

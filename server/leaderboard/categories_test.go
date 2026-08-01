package leaderboard

import (
	"os"
	"testing"
)

func TestPhase0CategoryCatalogAndPredicates(t *testing.T) {
	data, err := os.ReadFile("../../testdata/leaderboards/l7a-fixture.json")
	if err != nil {
		t.Fatal(err)
	}
	gates := []string{"gate.t2_to_t3", "gate.t4_to_t5", "gate.t7_to_t8"}
	catalog, err := LoadCategoryCatalog(data, gates)
	if err != nil {
		t.Fatal(err)
	}
	matching, err := catalog.Matching(TerminalFacts{GatesCrossed: gates, Facts: []string{"fact.completion"}, GeneratorsPurchasedTotal: 40})
	if err != nil {
		t.Fatal(err)
	}
	if ids := categoryIDs(matching); !sameStrings(ids, []string{"any_percent", "ethical_percent", "hundred_percent", "low_percent"}) {
		t.Fatalf("matching=%v", ids)
	}
	matching, err = catalog.Matching(TerminalFacts{GatesCrossed: gates[:2], Facts: []string{"fact.forbidden"}, GeneratorsPurchasedTotal: 41})
	if err != nil {
		t.Fatal(err)
	}
	if ids := categoryIDs(matching); !sameStrings(ids, []string{"any_percent"}) {
		t.Fatalf("restricted matching=%v", ids)
	}
}

func TestCategoryCatalogRejectsDriftAndOpenUnion(t *testing.T) {
	data, err := os.ReadFile("../../testdata/leaderboards/l7a-fixture.json")
	if err != nil {
		t.Fatal(err)
	}
	cases := [][]string{
		{"gate.t2_to_t3", "gate.t7_to_t8"},
		{"gate.t2_to_t3", "gate.t4_to_t5", "gate.t7_to_t8", "gate.unknown"},
	}
	for _, gates := range cases {
		if _, err := LoadCategoryCatalog(data, gates); err == nil {
			t.Fatalf("accepted gate drift %v", gates)
		}
	}
	unknown := []byte(`{"schema_version":1,"full_gate_set":["gate.example"],"fact_sets":{"completion_set":[],"forbidden_set":[]},"categories":[{"id":"any_percent","name_key":"category.any_percent","timer":"rta","predicate":{"kind":"script"}}]}`)
	if _, err := LoadCategoryCatalog(unknown, []string{"gate.example"}); err == nil {
		t.Fatal("accepted open predicate union")
	}
}

func categoryIDs(categories []Category) []string {
	result := make([]string, len(categories))
	for index := range categories {
		result[index] = categories[index].ID
	}
	return result
}

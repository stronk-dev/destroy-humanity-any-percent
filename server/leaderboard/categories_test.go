package leaderboard

import (
	"bytes"
	"os"
	"testing"
)

func TestPhase0CategoryCatalogAndPredicates(t *testing.T) {
	data, err := os.ReadFile("../../balance/categories/phase0.json")
	if err != nil {
		t.Fatal(err)
	}
	gates := []string{"gate.t2_to_t3", "gate.t3_to_t4", "gate.t4_to_t5", "gate.t7_to_t8"}
	catalog, err := LoadCategoryCatalog(data, gates)
	if err != nil {
		t.Fatal(err)
	}
	matching, err := catalog.Matching(TerminalFacts{GatesCrossed: gates, Facts: []string{"exit.acquihire"}, GeneratorsPurchasedTotal: 40})
	if err != nil {
		t.Fatal(err)
	}
	if ids := categoryIDs(matching); !sameStrings(ids, []string{"any_percent", "ethical_percent", "hundred_percent", "low_percent", "valuation"}) {
		t.Fatalf("matching=%v", ids)
	}
	matching, err = catalog.Matching(TerminalFacts{GatesCrossed: gates[:2], Facts: []string{"darkpattern.subscription_trap"}, GeneratorsPurchasedTotal: 41})
	if err != nil {
		t.Fatal(err)
	}
	if ids := categoryIDs(matching); !sameStrings(ids, []string{"any_percent", "valuation"}) {
		t.Fatalf("restricted matching=%v", ids)
	}
	if _, err := catalog.Matching(TerminalFacts{GatesCrossed: gates, Facts: []string{"unregistered.example"}}); err == nil {
		t.Fatal("accepted an unregistered terminal fact namespace")
	}
}

func TestCategoryCatalogRejectsDriftAndOpenUnion(t *testing.T) {
	data, err := os.ReadFile("../../balance/categories/phase0.json")
	if err != nil {
		t.Fatal(err)
	}
	cases := [][]string{
		{"gate.t2_to_t3", "gate.t4_to_t5", "gate.t7_to_t8"},
		{"gate.t2_to_t3", "gate.t3_to_t4", "gate.t4_to_t5", "gate.t7_to_t8", "gate.unknown"},
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
	driftedFacts := bytes.Replace(data, []byte(`"completion_set": []`), []byte(`"completion_set": ["exit.acquihire"]`), 1)
	if _, err := LoadCategoryCatalog(driftedFacts, []string{"gate.t2_to_t3", "gate.t3_to_t4", "gate.t4_to_t5", "gate.t7_to_t8"}); err == nil {
		t.Fatal("accepted Phase-0 fact-set drift")
	}
	driftedName := bytes.Replace(data, []byte(`"name_key": "category.any_percent"`), []byte(`"name_key": "category.wrong"`), 1)
	if _, err := LoadCategoryCatalog(driftedName, []string{"gate.t2_to_t3", "gate.t3_to_t4", "gate.t4_to_t5", "gate.t7_to_t8"}); err == nil {
		t.Fatal("accepted a non-canonical category name key")
	}
}

func categoryIDs(categories []Category) []string {
	result := make([]string, len(categories))
	for index := range categories {
		result[index] = categories[index].ID
	}
	return result
}

func TestMagnitudeKeyUsesCanonicalExponentAndPaddedMantissa(t *testing.T) {
	cases := map[string]MagnitudeKey{
		"0":                     {},
		"1e3":                   {Exponent: 3, Mantissa: 100_000_000_000},
		"1.23456789012e15":      {Exponent: 15, Mantissa: 123_456_789_012},
		"9.9e-4000000000000000": {Exponent: -4_000_000_000_000_000, Mantissa: 990_000_000_000},
	}
	for value, expected := range cases {
		actual, err := magnitudeKeyFromCanonical(value)
		if err != nil || actual != expected {
			t.Fatalf("%s key=%+v want=%+v err=%v", value, actual, expected, err)
		}
	}
	for _, value := range []string{"-1e3", "1.000000000000e3", "Infinity", "1e9000000000000000"} {
		if _, err := magnitudeKeyFromCanonical(value); err == nil {
			t.Fatalf("accepted invalid valuation key %q", value)
		}
	}
}

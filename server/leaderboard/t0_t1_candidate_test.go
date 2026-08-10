package leaderboard

import (
	"os"
	"testing"
)

func TestT0T1CandidateExtendsCanonicalFullGateSet(t *testing.T) {
	data, err := os.ReadFile("../../balance/testdata/t0-t1/categories-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	gates := []string{"gate.t0_to_t1", "gate.t2_to_t3", "gate.t3_to_t4", "gate.t4_to_t5", "gate.t7_to_t8"}
	catalog, err := LoadCategoryCatalog(data, gates)
	if err != nil {
		t.Fatal(err)
	}
	if !sameStrings(catalog.FullGateSet, gates) {
		t.Fatalf("full gate set = %v", catalog.FullGateSet)
	}
}

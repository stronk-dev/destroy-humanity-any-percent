package routes

import (
	"os"
	"testing"
)

func TestT0T1CandidateAddsLiteralFirstGateWithoutChangingLaterRoutes(t *testing.T) {
	data, err := os.ReadFile("../../balance/testdata/t0-t1/routes-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := LoadCatalog(data)
	if err != nil {
		t.Fatal(err)
	}
	gates := catalog.Gates()
	if len(gates) != 5 || gates[0].ID != "gate.t0_to_t1" || len(gates[0].Requirement) != 1 || gates[0].Requirement[0].Amount.String() != "1e5" || len(gates[0].Routes) != 0 {
		t.Fatalf("first gate = %+v", gates)
	}
	if len(gates[3].Routes) != 6 || gates[3].ID != "gate.t4_to_t5" {
		t.Fatalf("later routes changed: %+v", gates[3])
	}
}

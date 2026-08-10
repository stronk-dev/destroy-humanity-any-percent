package economy

import (
	"os"
	"testing"
)

func TestT0T1CandidateLoadsAndBindsEveryGate(t *testing.T) {
	data, err := os.ReadFile("../../balance/testdata/t0-t1/economy-v4.json")
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := LoadCatalog(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(catalog.GeneratorClasses()) != 9 || len(catalog.Upgrades()) != 10 || len(catalog.SynergyPools()) != 2 {
		t.Fatalf("candidate rows generators=%d upgrades=%d pools=%d", len(catalog.GeneratorClasses()), len(catalog.Upgrades()), len(catalog.SynergyPools()))
	}
	if err := catalog.ValidateGateReferences([]string{"gate.t0_to_t1", "gate.t2_to_t3", "gate.t3_to_t4", "gate.t4_to_t5", "gate.t7_to_t8"}); err != nil {
		t.Fatal(err)
	}
}

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
	for id, amount := range map[string]string{
		"upgrade.employee_handbook_v0":  "7.5e8",
		"upgrade.institutional_memory":  "9.5e8",
		"upgrade.nephew_business_cards": "8.64e4",
		"upgrade.refurbished_sticker":   "8.5e8",
	} {
		upgrade, ok := catalog.Upgrade(id)
		if !ok || upgrade.Cost.Amount.String() != amount || len(upgrade.Requires) != 1 || upgrade.Requires[0].Value.String() != amount {
			t.Fatalf("rescaled upgrade %s = %+v", id, upgrade)
		}
	}
	for _, id := range []string{"active.building.generator.beige_tower", "active.click", "active.production"} {
		if _, ok := catalog.MultiplierSource(id); !ok {
			t.Fatalf("missing active-play multiplier source %s", id)
		}
	}
	legal, ok := catalog.GeneratorClass("generator.legal_dept")
	if !ok || len(legal.Roles) != 1 || legal.Roles[0].Kind != RoleSynergyFeed || legal.Roles[0].PoolID != "pool.institutional_knowledge" {
		t.Fatalf("legal-department balance choice = %+v", legal)
	}
}

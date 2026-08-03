package economy

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func foundationCatalogBytes(t *testing.T) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "..", "testdata", "economy-foundation-v4.json"))
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestPurchasableCatalogV4LoadsTypedBindings(t *testing.T) {
	catalog, err := LoadCatalog(foundationCatalogBytes(t))
	if err != nil {
		t.Fatal(err)
	}
	if catalog.ProvisionTickMS() != 60_000 || len(catalog.Upgrades()) != 1 || len(catalog.SynergyPools()) != 1 {
		t.Fatalf("foundation counts tick=%d upgrades=%d pools=%d", catalog.ProvisionTickMS(), len(catalog.Upgrades()), len(catalog.SynergyPools()))
	}
	high, ok := catalog.GeneratorClass("generator.high")
	if !ok || high.Tier != 1 || high.Provision == nil || high.Provision.GeneratorID != "generator.low" || len(high.Roles) != 1 {
		t.Fatalf("high generator = %+v exists=%v", high, ok)
	}
	upgrade, ok := catalog.Upgrade("upgrade.click")
	if !ok || len(upgrade.Effects) != 1 || upgrade.Effects[0].Target != "manual.click" || upgrade.Effects[0].Factor.String() != "2e0" {
		t.Fatalf("upgrade = %+v exists=%v", upgrade, ok)
	}
	if declaration, ok := catalog.MultiplierSource("upgrade.click.factor"); !ok || declaration.Provider != "upgrade.click" || declaration.Target != "manual.click" {
		t.Fatalf("derived upgrade source = %+v exists=%v", declaration, ok)
	}
	if err := catalog.ValidateGateReferences([]string{"gate.t0_to_t1"}); err != nil {
		t.Fatal(err)
	}
	if err := catalog.ValidateGateReferences([]string{"gate.other"}); err == nil {
		t.Fatal("unknown availability gate was accepted")
	}
}

func TestPurchasableCatalogV4RejectsUnboundRolesAndTopology(t *testing.T) {
	cases := map[string]func(map[string]any){
		"roleless-generator": func(root map[string]any) {
			root["generator_classes"].([]any)[0].(map[string]any)["roles"] = []any{}
		},
		"dangling-manual-role": func(root map[string]any) {
			roles := root["generator_classes"].([]any)[0].(map[string]any)["roles"].([]any)
			roles[1].(map[string]any)["action_id"] = "manual.missing"
		},
		"unbound-synergy-role": func(root map[string]any) {
			root["generator_classes"].([]any)[0].(map[string]any)["roles"].([]any)[0].(map[string]any)["pool_id"] = "pool.missing"
		},
		"non-adjacent-provision": func(root map[string]any) {
			tier := float64(2)
			root["generator_classes"].([]any)[1].(map[string]any)["tier"] = tier
		},
		"uncapped-provision-target": func(root map[string]any) {
			delete(root["generator_classes"].([]any)[0].(map[string]any), "provisioned_hardcap")
		},
		"pool-without-source-declaration": func(root map[string]any) {
			root["multiplier_sources"] = []any{}
		},
		"upgrade-effect-duplicates-pool": func(root map[string]any) {
			root["upgrades"].([]any)[0].(map[string]any)["effects"].([]any)[0].(map[string]any)["source_id"] = "pool.operations"
		},
		"scalar-upgrade-requires": func(root map[string]any) {
			upgrade := root["upgrades"].([]any)[0].(map[string]any)
			upgrade["requires"] = upgrade["requires"].([]any)[0]
		},
		"empty-upgrade-requires": func(root map[string]any) {
			root["upgrades"].([]any)[0].(map[string]any)["requires"] = []any{}
		},
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			var root map[string]any
			if err := json.Unmarshal(foundationCatalogBytes(t), &root); err != nil {
				t.Fatal(err)
			}
			mutate(root)
			data, err := json.Marshal(root)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := LoadCatalog(data); err == nil {
				t.Fatal("invalid foundation catalog was accepted")
			}
		})
	}
}

func TestUpgradeResourcePredicatesRequireCompanyScope(t *testing.T) {
	var root map[string]any
	if err := json.Unmarshal(foundationCatalogBytes(t), &root); err != nil {
		t.Fatal(err)
	}
	root["resources"] = append(root["resources"].([]any), map[string]any{
		"id": "founder.clout", "scope": "founder", "numeric_kind": "decimal", "initial": "0", "minimum": "0", "hardcap": nil,
	})
	condition := root["upgrades"].([]any)[0].(map[string]any)["requires"].([]any)[0].(map[string]any)
	condition["resource_id"] = "founder.clout"
	data, err := json.Marshal(root)
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := LoadCatalog(data)
	if err != nil {
		t.Fatal(err)
	}
	if err := catalog.ValidateGateReferences([]string{"gate.t0_to_t1"}); err == nil {
		t.Fatal("non-company upgrade predicate was accepted")
	}
}

func TestPurchasableCatalogAccessorsAreImmutable(t *testing.T) {
	catalog, err := LoadCatalog(foundationCatalogBytes(t))
	if err != nil {
		t.Fatal(err)
	}
	generator, _ := catalog.GeneratorClass("generator.low")
	generator.Roles[0].PoolID = "changed"
	generator.Ladder[0].PurchasedAt = 999
	upgrade, _ := catalog.Upgrade("upgrade.click")
	upgrade.Effects[0].Target = "changed"
	pool, _ := catalog.SynergyPool("pool.operations")
	pool.Sources[0].ID = "changed"
	againGenerator, _ := catalog.GeneratorClass("generator.low")
	againUpgrade, _ := catalog.Upgrade("upgrade.click")
	againPool, _ := catalog.SynergyPool("pool.operations")
	if againGenerator.Roles[0].PoolID == "changed" || againGenerator.Ladder[0].PurchasedAt == 999 || againUpgrade.Effects[0].Target == "changed" || againPool.Sources[0].ID == "changed" {
		t.Fatal("foundation catalog mutated through accessors")
	}
}

package gameui

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	"cloud-clicker/server/economy"
	"cloud-clicker/server/production"
	"cloud-clicker/server/save"
)

const axisArmGoldenPath = "../../testdata/axis-stack/game-ui-arm-v1.json"

func axisProjectionState(t *testing.T, catalog *economy.Catalog) *save.State {
	t.Helper()
	ledger, err := economy.NewLedger(catalog, economy.ScopeCompany)
	if err != nil {
		t.Fatal(err)
	}
	counts := map[string]int64{}
	for _, generator := range catalog.GeneratorClassesForScope(economy.ScopeCompany) {
		counts[generator.ID] = 0
	}
	return &save.State{WireVersion: 19, Ledger: ledger, GeneratorCounts: counts, UpgradesOwned: map[string]bool{"upgrade.pr_intern_1": true},
		AchievementsEarnedRun:   map[string]bool{"achievement.first_gate": true},
		AchievementsAttainedRun: map[string]bool{"achievement.first_gate": true, "achievement.generators_purchased_1": true, "achievement.generators_purchased_25": true},
		AttainmentScoreRun:      8}
}

// TestAxisStackArmIsServerDerived pins the Clout v1 CV9 producer block on the
// fixture economy and writes the golden the browser witness renders
// (UPDATE_AXIS_VECTORS=1 regenerates).
func TestAxisStackArmIsServerDerived(t *testing.T) {
	data, err := os.ReadFile("../../balance/testdata/axis-stack/economy-v5-fixture.json")
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := economy.LoadCatalog(data)
	if err != nil {
		t.Fatal(err)
	}
	arm, err := projectAxisStack(production.CatalogBundle{Economy: catalog}, axisProjectionState(t, catalog))
	if err != nil || arm == nil {
		t.Fatalf("arm=%v err=%v", arm, err)
	}
	if arm.InputValue != 8 || arm.InputCap != 44 || arm.Saturated || arm.Product != "1.2e0" || len(arm.Contributions) != 1 ||
		arm.Contributions[0].Factor != "1.2e0" || len(arm.Interns) != 2 || arm.Interns[1].Owned || arm.Interns[1].Factor != "1.16e0" || arm.Interns[1].Minimum != 10 {
		t.Fatalf("arm=%+v", arm)
	}
	if len(arm.Attained) != 3 || !arm.Attained[0].EarnedThisRun || arm.Attained[1].EarnedThisRun {
		t.Fatalf("attained=%+v", arm.Attained)
	}
	encoded, _ := json.MarshalIndent(arm, "", "  ")
	encoded = append(encoded, '\n')
	if os.Getenv("UPDATE_AXIS_VECTORS") == "1" {
		if err := os.WriteFile(axisArmGoldenPath, encoded, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if committed, err := os.ReadFile(axisArmGoldenPath); err != nil || !bytes.Equal(committed, encoded) {
		t.Fatalf("axis arm golden drifted (err=%v)", err)
	}
	// A saturated input is flagged and clamps every factor.
	state := axisProjectionState(t, catalog)
	state.AttainmentScoreRun = 60
	saturated, err := projectAxisStack(production.CatalogBundle{Economy: catalog}, state)
	if err != nil || !saturated.Saturated || saturated.Contributions[0].Factor != "2.1e0" {
		t.Fatalf("saturated=%+v err=%v", saturated, err)
	}
	// Epoch-8 (no axis_stack) projects null.
	epoch8, err := os.ReadFile("../../balance/catalogs/phase0.json")
	if err != nil {
		t.Fatal(err)
	}
	plain, err := economy.LoadCatalog(epoch8)
	if err != nil {
		t.Fatal(err)
	}
	if none, err := projectAxisStack(production.CatalogBundle{Economy: plain}, &save.State{}); none != nil || err != nil {
		t.Fatalf("epoch-8 arm=%v err=%v", none, err)
	}
}

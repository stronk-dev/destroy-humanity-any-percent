package minigame

import (
	"encoding/json"
	"os"
	"testing"
)

func TestLoadCatalogClosesPinnedPolicySurface(t *testing.T) {
	data, err := os.ReadFile("../../testdata/minigame/catalog-v2.json")
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := LoadCatalog(data)
	definition, found := catalog.Definition("fixture.counter")
	if err != nil || !found || len(catalog.MinigameIDs()) != 1 || !catalog.HasRatingSeason("ranked") ||
		catalog.HasRatingSeason("invented") || definition.Payout.PayoutScoreFactID != "score.total" ||
		definition.Rating.EloCeiling != 3000 || definition.OfflineQuality.NeutralFloorPPM != 500_000 {
		t.Fatalf("unexpected catalog: %#v %#v %v", catalog, definition, err)
	}
	if _, found := catalog.Definition("invented"); found {
		t.Fatal("resolved an undeclared minigame")
	}
}

func TestLoadCatalogV3RequiresSoulGateWithoutReinterpretingV2(t *testing.T) {
	data, err := os.ReadFile("../../testdata/minigame/catalog-v2.json")
	if err != nil {
		t.Fatal(err)
	}
	var root map[string]any
	if err := json.Unmarshal(data, &root); err != nil {
		t.Fatal(err)
	}
	root["schema_version"] = float64(3)
	root["minigames"].([]any)[0].(map[string]any)["soul_gate"] = "human_hobby"
	v3, _ := json.Marshal(root)
	catalog, err := LoadCatalog(v3)
	if err != nil || !catalog.SchemaSupportsSoul() {
		t.Fatalf("v3 Soul catalog=%#v err=%v", catalog, err)
	}
	delete(root["minigames"].([]any)[0].(map[string]any), "soul_gate")
	missing, _ := json.Marshal(root)
	if _, err := LoadCatalog(missing); err == nil {
		t.Fatal("v3 accepted missing soul_gate")
	}
	if _, err := LoadCatalog(data); err != nil {
		t.Fatalf("historical v2 no longer decodes: %v", err)
	}
}

func TestLoadCatalogV3AcceptsFiscalUnlockPitchDefinition(t *testing.T) {
	data, err := os.ReadFile("../../testdata/minigame/pitch-v3.json")
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := LoadCatalog(data)
	definition, found := catalog.Definition("pitch")
	if err != nil || !found || definition.Unlock.Kind != "fiscal_unlock" || definition.Unlock.UnlockID != "minigame.pitch" ||
		definition.Payout.CreditedResourceID != "company.cash" || definition.OfflineQuality.AutomationDestination != "minigame.pitch" {
		t.Fatalf("definition=%+v found=%v err=%v", definition, found, err)
	}
}

func TestLoadCatalogRejectsPartialOrDuplicatedPolicy(t *testing.T) {
	for _, invalid := range []string{
		`{"schema_version":2,"rating_seasons":[],"minigames":[],"extra":0}`,
		`{"schema_version":2,"rating_seasons":["ranked","ranked"],"minigames":[]}`,
		`{"schema_version":1,"rating_seasons":[],"minigames":[]}`,
		`{"schema_version":2,"rating_seasons":[],"minigames":[{"minigame_id":"partial"}]}`,
	} {
		if _, err := LoadCatalog([]byte(invalid)); err == nil {
			t.Fatalf("accepted %s", invalid)
		}
	}
}

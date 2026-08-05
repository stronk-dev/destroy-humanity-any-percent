package minigame

import (
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

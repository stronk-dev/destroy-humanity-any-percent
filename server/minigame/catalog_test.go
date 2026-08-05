package minigame

import "testing"

func TestLoadCatalogClosesActivationDomains(t *testing.T) {
	catalog, err := LoadCatalog([]byte(`{"schema_version":1,"minigame_ids":["combat.duel","market.pitch"],"rating_seasons":["preseason","ranked"]}`))
	if err != nil || len(catalog.MinigameIDs()) != 2 || !catalog.HasRatingSeason("ranked") || catalog.HasRatingSeason("invented") {
		t.Fatalf("unexpected catalog: %#v %v", catalog, err)
	}
	for _, invalid := range []string{
		`{"schema_version":1,"minigame_ids":[],"rating_seasons":[],"extra":0}`,
		`{"schema_version":1,"minigame_ids":["z","a"],"rating_seasons":[]}`,
		`{"schema_version":1,"minigame_ids":[],"rating_seasons":["ranked","ranked"]}`,
		`{"schema_version":2,"minigame_ids":[],"rating_seasons":[]}`,
	} {
		if _, err := LoadCatalog([]byte(invalid)); err == nil {
			t.Fatalf("accepted %s", invalid)
		}
	}
}

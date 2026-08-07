package production

import (
	"os"
	"testing"

	"cloud-clicker/server/minigameapi"
	"cloud-clicker/server/save"
)

func TestMinigameAPIArtifactAloneOwnsFounderV21Activation(t *testing.T) {
	data, err := os.ReadFile("../../balance/testdata/minigame-api-candidate-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := minigameapi.LoadCatalog(data)
	if err != nil {
		t.Fatal(err)
	}
	bundle := CatalogBundle{MinigameAPI: catalog}
	founder, company := bundle.versionFloors()
	if founder != 21 || company != save.CurrentVersion {
		t.Fatalf("floors founder=%d company=%d", founder, company)
	}
	state := &save.State{WireVersion: 20, MinigameSessionSeq: 17}
	if err := activateFounderFeatureState(state, bundle, 21, 1, nil); err != nil {
		t.Fatal(err)
	}
	if state.MinigameSessionSeq != 0 {
		t.Fatalf("activation did not initialize sequence: %d", state.MinigameSessionSeq)
	}
	if err := activateFounderFeatureState(&save.State{WireVersion: 20}, CatalogBundle{}, 21, 1, nil); err == nil {
		t.Fatal("Founder v21 activated without minigame_api artifact")
	}
}

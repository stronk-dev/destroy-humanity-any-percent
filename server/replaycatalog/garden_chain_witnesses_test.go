package replaycatalog

import (
	"testing"

	"cloud-clicker/server/production"
)

func TestGardenTransitiveBundleChain(t *testing.T) {
	complete := gardenArtifacts(t)
	valid, err := loadArtifacts(complete)
	if err != nil || valid.Garden == nil {
		t.Fatalf("positive chain: %v", err)
	}
	// Every scalar predecessor required by Garden, not unrelated optional lanes.
	for _, name := range []string{"meters", "achievements", "minigames", "pets", "fiscal", "soul", "pitch", "minigame_api", "reputation_tree", "pet_species", "cosmetics"} {
		t.Run(name, func(t *testing.T) {
			missing := map[string][]byte{}
			for key, value := range complete {
				if key != name {
					missing[key] = value
				}
			}
			if _, err := loadArtifacts(missing); err == nil {
				t.Fatal("Garden admitted without required predecessor")
			}
		})
	}
	without := map[string][]byte{}
	for key, value := range complete {
		if key != "server_garden" {
			without[key] = value
		}
	}
	reduced, err := loadArtifacts(without)
	if err != nil || reduced.Garden != nil || reduced.ConstantsHash == valid.ConstantsHash {
		t.Fatalf("Garden-absent control: %v", err)
	}
	if _, err := Load(reduced.ConstantsHash, complete); err == nil {
		t.Fatal("wrong label admitted complete Garden artifacts")
	}
	forged := reduced
	forged.Garden = valid.Garden
	if _, ok := (production.ReplayCatalogSet{forged.ConstantsHash: forged}).ResolveReplayCatalogs(forged.ConstantsHash); ok {
		t.Fatal("pointer without pinned bytes resolved")
	}
}

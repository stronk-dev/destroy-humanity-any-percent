package replaycatalog

import (
	"os"
	"path/filepath"
	"testing"

	"cloud-clicker/server/production"
)

// cosmeticsArtifacts is the pet_species chain plus the cosmetics fixture: the
// smallest chain that can pin Founder v24 (Cosmetic Shop v1 §2/§3).
func cosmeticsArtifacts(t *testing.T) map[string][]byte {
	t.Helper()
	artifacts := petSpeciesArtifacts(t)
	data, err := os.ReadFile(filepath.Join("..", "..", "balance", "testdata", "cosmetics", "fixture-v1.json"))
	if err != nil {
		t.Fatal(err)
	}
	artifacts["cosmetics"] = data
	return artifacts
}

// C2 (OD-10): cosmetics joins constants identity and requires pet_species.
func TestLoadCosmeticsRequiresItsChain(t *testing.T) {
	complete := cosmeticsArtifacts(t)
	bundle, err := loadArtifacts(complete)
	if err != nil || bundle.Cosmetics == nil || len(bundle.Cosmetics.Items) != 1 {
		t.Fatalf("complete cosmetics bundle rejected: %v", err)
	}
	if _, ok := (production.ReplayCatalogSet{bundle.ConstantsHash: bundle}).ResolveReplayCatalogs(bundle.ConstantsHash); !ok {
		t.Fatal("complete cosmetics bundle is not valid for replay")
	}
	withoutCosmetics := map[string][]byte{}
	for name, data := range complete {
		if name != "cosmetics" {
			withoutCosmetics[name] = data
		}
	}
	reduced, err := loadArtifacts(withoutCosmetics)
	if err != nil || reduced.ConstantsHash == bundle.ConstantsHash {
		t.Fatalf("cosmetics does not join constants identity: %v", err)
	}
	withoutSpecies := map[string][]byte{}
	for name, data := range complete {
		if name != "pet_species" {
			withoutSpecies[name] = data
		}
	}
	if _, err := loadArtifacts(withoutSpecies); err == nil {
		t.Fatal("cosmetics loaded without pet_species")
	}
	priced := map[string][]byte{}
	for name, data := range complete {
		priced[name] = data
	}
	priced["cosmetics"] = []byte(`{"schema_version":1,"items":[{"cosmetic_id":"horse_armor","slot":"pet","price":0,"unlock":{"kind":"active_company_tier_at_least","tier":1}}]}`)
	if _, err := loadArtifacts(priced); err == nil {
		t.Fatal("a priced cosmetics artifact loaded")
	}
	forged := reduced
	forged.Cosmetics = bundle.Cosmetics
	if _, ok := (production.ReplayCatalogSet{forged.ConstantsHash: forged}).ResolveReplayCatalogs(forged.ConstantsHash); ok {
		t.Fatal("a bundle carrying Cosmetics without its artifact bytes resolved")
	}
}

// gardenArtifacts is the cosmetics chain with the fixture Fiscal artifact
// (adds the minigame.server_garden unlock row) and the fixture server_garden.
func gardenArtifacts(t *testing.T) map[string][]byte {
	t.Helper()
	artifacts := cosmeticsArtifacts(t)
	for name, path := range map[string]string{"fiscal": "fiscal-fixture-v1.json", "server_garden": "fixture-v1.json"} {
		data, err := os.ReadFile(filepath.Join("..", "..", "balance", "testdata", "server-garden", path))
		if err != nil {
			t.Fatal(err)
		}
		artifacts[name] = data
	}
	return artifacts
}

// Server Garden SG1/SG2: server_garden joins constants identity, requires
// cosmetics, and binds rule 10 to the pinned Fiscal rows.
func TestLoadGardenRequiresItsChain(t *testing.T) {
	complete := gardenArtifacts(t)
	bundle, err := loadArtifacts(complete)
	if err != nil || bundle.Garden == nil || len(bundle.Garden.Species) != 5 {
		t.Fatalf("complete garden bundle rejected: %v", err)
	}
	if _, ok := (production.ReplayCatalogSet{bundle.ConstantsHash: bundle}).ResolveReplayCatalogs(bundle.ConstantsHash); !ok {
		t.Fatal("complete garden bundle is not valid for replay")
	}
	without := func(name string) map[string][]byte {
		reduced := map[string][]byte{}
		for key, data := range complete {
			if key != name {
				reduced[key] = data
			}
		}
		return reduced
	}
	if _, err := loadArtifacts(without("cosmetics")); err == nil {
		t.Fatal("server_garden loaded without cosmetics")
	}
	productionFiscal := without("")
	productionFiscal["fiscal"] = cosmeticsArtifacts(t)["fiscal"]
	if _, err := loadArtifacts(productionFiscal); err == nil {
		t.Fatal("server_garden loaded against a Fiscal artifact with no garden unlock row")
	}
	reduced, err := loadArtifacts(without("server_garden"))
	if err != nil || reduced.ConstantsHash == bundle.ConstantsHash || reduced.Garden != nil {
		t.Fatalf("server_garden does not join constants identity: %v", err)
	}
}

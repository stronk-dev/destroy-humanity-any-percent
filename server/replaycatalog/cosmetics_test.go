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

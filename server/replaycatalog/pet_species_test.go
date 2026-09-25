package replaycatalog

import (
	"os"
	"path/filepath"
	"testing"

	"cloud-clicker/server/production"
)

// petSpeciesArtifacts is the live epoch plus the Reputation tree fixture and
// the pet_species fixture: the smallest chain that can pin Founder v23.
func petSpeciesArtifacts(t *testing.T) map[string][]byte {
	t.Helper()
	live := liveEpochArtifacts(t)
	artifacts := map[string][]byte{}
	for name, data := range live {
		artifacts[name] = data
	}
	read := func(parts ...string) []byte {
		data, err := os.ReadFile(filepath.Join(append([]string{"..", ".."}, parts...)...))
		if err != nil {
			t.Fatal(err)
		}
		return data
	}
	artifacts["economy"] = withReputationDeclaration(t, live["economy"])
	artifacts["reputation_tree"] = read("balance", "testdata", "reputation-tree", "fixture-v1.json")
	artifacts["pet_species"] = read("balance", "testdata", "pet-species", "fixture-v1.json")
	return artifacts
}

// P3 (PA2): pet_species joins constants identity and requires pets and, on
// the scalar Founder chain, reputation_tree.
func TestLoadPetSpeciesRequiresItsChain(t *testing.T) {
	complete := petSpeciesArtifacts(t)
	bundle, err := loadArtifacts(complete)
	if err != nil || bundle.PetSpecies == nil || bundle.PetSpecies.MaxPetsPerFounder != 1 {
		t.Fatalf("complete pet_species bundle rejected: %v", err)
	}
	if _, ok := (production.ReplayCatalogSet{bundle.ConstantsHash: bundle}).ResolveReplayCatalogs(bundle.ConstantsHash); !ok {
		t.Fatal("complete pet_species bundle is not valid for replay")
	}
	withoutSpecies := map[string][]byte{}
	for name, data := range complete {
		if name != "pet_species" {
			withoutSpecies[name] = data
		}
	}
	reduced, err := loadArtifacts(withoutSpecies)
	if err != nil || reduced.ConstantsHash == bundle.ConstantsHash {
		t.Fatalf("pet_species does not join constants identity: %v", err)
	}
	for _, missing := range []string{"reputation_tree", "pets"} {
		artifacts := map[string][]byte{}
		for name, data := range complete {
			if name != missing {
				artifacts[name] = data
			}
		}
		if missing == "reputation_tree" {
			artifacts["economy"] = liveEpochArtifacts(t)["economy"]
		}
		if _, err := loadArtifacts(artifacts); err == nil {
			t.Fatalf("pet_species loaded without %s", missing)
		}
	}
	invalid := map[string][]byte{}
	for name, data := range complete {
		invalid[name] = data
	}
	invalid["pet_species"] = []byte(`{"schema_version":1,"max_pets_per_founder":0,"species":[]}`)
	if _, err := loadArtifacts(invalid); err == nil {
		t.Fatal("an invalid pet_species artifact loaded")
	}
	// valid() re-checks the chain for bundles assembled outside the loader.
	forged := reduced
	forged.PetSpecies = bundle.PetSpecies
	if _, ok := (production.ReplayCatalogSet{forged.ConstantsHash: forged}).ResolveReplayCatalogs(forged.ConstantsHash); ok {
		t.Fatal("a bundle carrying PetSpecies without its artifact bytes resolved")
	}
}

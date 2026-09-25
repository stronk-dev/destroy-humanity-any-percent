package gameui

import (
	"testing"
	"time"

	"cloud-clicker/server/pet"
)

// GS4-A2: the pet care surface fact is true only when an adopted pet exists;
// an empty v23 pet map (adoption available, nothing adopted) keeps it false.
func TestPetCareFactRequiresAnAdoptedPet(t *testing.T) {
	bundle := petSpeciesBundle(t)
	company, founder := featureStates(bundle)
	now := time.UnixMilli(1_800_000_000_000).UTC()
	founder.WireVersion, founder.AgeMS = 23, 1_000
	founder.Pets, founder.PetIdentities = map[string]pet.CareState{}, map[string]pet.Identity{}
	fact := func() any {
		features, err := projectFeatures(bundle, company, founder, now, false, 0)
		if err != nil {
			t.Fatal(err)
		}
		for _, row := range featureFacts(features) {
			if row.FactID == "feature.pets" {
				return row.Value
			}
		}
		t.Fatal("feature.pets fact missing")
		return nil
	}
	if fact() != false {
		t.Fatal("feature.pets must be false with no adopted pet")
	}
	const id = "01986666-aaaa-7aaa-8aaa-aaaaaaaaaaaa"
	care, err := pet.InitialCareState(bundle.Pets, 1_000)
	if err != nil {
		t.Fatal(err)
	}
	founder.Pets[id] = care
	founder.PetIdentities[id] = pet.Identity{SpeciesID: "pet_species.server_room_cat", Temperament: "sassy", PaletteID: "pet_palette.fur_02",
		NameKey: "pet.name.server_room_cat.n04", AdoptedAtMS: 1, AdoptedAtAttendedMS: 1_000}
	if fact() != true {
		t.Fatal("feature.pets must be true once a pet is adopted")
	}
}

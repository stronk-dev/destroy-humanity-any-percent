package gameui

import (
	"encoding/json"
	"os"
	"sort"
	"strings"
	"testing"
	"time"

	"cloud-clicker/server/pet"
	"cloud-clicker/server/production"
	"cloud-clicker/server/replaycatalog"
	"cloud-clicker/server/reputation"
	"cloud-clicker/server/save"
)

func petSpeciesBundle(t *testing.T) production.CatalogBundle {
	t.Helper()
	live := pinnedBundle(t)
	artifacts := map[string][]byte{}
	for name, data := range live.Artifacts {
		artifacts[name] = data
	}
	var economy map[string]any
	if err := json.Unmarshal(artifacts["economy"], &economy); err != nil {
		t.Fatal(err)
	}
	economy["multiplier_sources"] = append(economy["multiplier_sources"].([]any), map[string]any{"id": "reputation.founder_bonus", "slot": "prestige", "target": "all", "provider": reputation.Provider})
	artifacts["economy"], _ = json.Marshal(economy)
	for name, path := range map[string]string{"reputation_tree": "../../balance/testdata/reputation-tree/fixture-v1.json", "pet_species": "../../balance/testdata/pet-species/fixture-v1.json"} {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		artifacts[name] = data
	}
	hash, err := save.ConstantsHashArtifacts(artifacts)
	if err != nil {
		t.Fatal(err)
	}
	bundle, err := replaycatalog.Load(hash, artifacts)
	if err != nil {
		t.Fatal(err)
	}
	return bundle
}

// AC11: the pet_adoption arm exposes identity, band and eligible actions only,
// is absent below Founder v23, and projects care at effective attendance.
func TestPetAdoptionArmExposesOnlyIdentityBandAndEligibility(t *testing.T) {
	bundle := petSpeciesBundle(t)
	company, founder := featureStates(bundle)
	now := time.UnixMilli(1_800_000_000_000).UTC()
	founder.WireVersion, founder.AgeMS = 22, 1_000
	features, err := projectFeatures(bundle, company, founder, now, false, 0)
	if err != nil || features.PetAdoption != nil {
		t.Fatalf("pre-v23 pet_adoption=%+v err=%v", features.PetAdoption, err)
	}
	founder.WireVersion = 23
	founder.Pets, founder.PetIdentities = map[string]pet.CareState{}, map[string]pet.Identity{}
	features, err = projectFeatures(bundle, company, founder, now, false, 0)
	if err != nil || features.PetAdoption == nil || features.PetAdoption.PetAdoption.Count != 0 || features.PetAdoption.PetAdoption.Cap != 1 ||
		features.PetAdoption.PetAdoption.StarterSpeciesID != "pet_species.server_room_cat" || len(features.PetAdoption.PetAdoption.NameKeys) != 12 {
		t.Fatalf("empty v23 arm=%+v err=%v", features.PetAdoption, err)
	}
	const id = "01986666-aaaa-7aaa-8aaa-aaaaaaaaaaaa"
	care, err := pet.InitialCareState(bundle.Pets, 1_000)
	if err != nil {
		t.Fatal(err)
	}
	founder.Pets[id] = care
	founder.PetIdentities[id] = pet.Identity{SpeciesID: "pet_species.server_room_cat", Temperament: "sassy", PaletteID: "pet_palette.fur_02",
		NameKey: "pet.name.server_room_cat.n04", AdoptedAtMS: 1, AdoptedAtAttendedMS: 1_000}
	fresh, err := projectFeatures(bundle, company, founder, now, false, 0)
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(fresh.PetAdoption)
	if err != nil {
		t.Fatal(err)
	}
	var arm struct {
		PetAdoption map[string]json.RawMessage   `json:"pet_adoption"`
		Pets        []map[string]json.RawMessage `json:"pets"`
	}
	if err := json.Unmarshal(encoded, &arm); err != nil || len(arm.Pets) != 1 {
		t.Fatalf("arm=%s err=%v", encoded, err)
	}
	keys := []string{}
	for key := range arm.Pets[0] {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	if strings.Join(keys, ",") != "eligible_action_ids,name_key,palette_id,pet_id,species_id,status_band,temperament" {
		t.Fatalf("pet row keys %v", keys)
	}
	// A day of attendance later the projected band decays from the fresh one.
	aged, err := projectFeatures(bundle, company, founder, now, false, 86_400_000)
	if err != nil {
		t.Fatal(err)
	}
	if fresh.PetAdoption.Pets[0].StatusBand == aged.PetAdoption.Pets[0].StatusBand {
		t.Fatalf("projected band ignores attendance: %s", fresh.PetAdoption.Pets[0].StatusBand)
	}
	if care := founder.Pets[id]; care.EvaluatedThroughAttendedMS != 1_000 {
		t.Fatal("projection mutated the care record")
	}
	facts := featureFacts(fresh)
	found := false
	for _, fact := range facts {
		found = found || fact.FactID == "feature.pet_adoption" && fact.Value == true
	}
	if !found {
		t.Fatalf("feature.pet_adoption fact missing: %+v", facts)
	}
}

// RP-388: the persisted care receipt and a later read have different attendance
// samples. At the retained failure's threshold, only two more milliseconds
// change the projected band without changing the stored care record.
func TestPetAdoptionBandCanChangeBetweenReceiptAndRead(t *testing.T) {
	bundle := petSpeciesBundle(t)
	_, founder := featureStates(bundle)
	founder.WireVersion, founder.AgeMS = 23, 0
	care, err := pet.InitialCareState(bundle.Pets, 1742)
	if err != nil {
		t.Fatal(err)
	}
	applied, err := pet.ApplyCareTransition(care, bundle.Pets, pet.CareTransitionInput{
		ActionID: "care.feed", AttendedBeforeMS: 1742, AttendedAfterMS: 1920,
	})
	if err != nil || !applied.Applied || applied.StatusBand != "high" ||
		applied.State.StatDecayRemaindersPPM[pet.StatEnergy] != 356000 {
		t.Fatalf("retained care boundary not reproduced: %+v, %v", applied, err)
	}
	const id = "01986666-aaaa-7aaa-8aaa-aaaaaaaaaaaa"
	founder.Pets = map[string]pet.CareState{id: applied.State}
	founder.PetIdentities = map[string]pet.Identity{id: {SpeciesID: "pet_species.server_room_cat", Temperament: "sassy",
		PaletteID: "pet_palette.fur_02", NameKey: "pet.name.server_room_cat.n00", AdoptedAtMS: 1, AdoptedAtAttendedMS: 1742}}
	before, err := json.Marshal(founder.Pets)
	if err != nil {
		t.Fatal(err)
	}
	for _, sample := range []struct {
		attendedMS int64
		band       pet.StatusBand
	}{{1920, "high"}, {1921, "high"}, {1922, "normal"}} {
		arm, err := projectPets(bundle, founder, sample.attendedMS)
		if err != nil || arm == nil || len(arm.Pets) != 1 {
			t.Fatalf("project at %d: %+v, %v", sample.attendedMS, arm, err)
		}
		if arm.Pets[0].StatusBand != sample.band {
			t.Fatalf("project at %d: band=%s, want %s", sample.attendedMS, arm.Pets[0].StatusBand, sample.band)
		}
		for _, action := range arm.Pets[0].EligibleActionIDs {
			if action == "care.feed" {
				t.Fatalf("project at %d re-enabled the cooling-down feed action", sample.attendedMS)
			}
		}
	}
	after, err := json.Marshal(founder.Pets)
	if err != nil || string(after) != string(before) {
		t.Fatalf("read projection changed stored care: %s -> %s, %v", before, after, err)
	}
}

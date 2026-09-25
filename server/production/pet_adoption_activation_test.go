package production

import (
	"errors"
	"os"
	"testing"
	"time"

	"cloud-clicker/server/copykeys"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/pet"
	"cloud-clicker/server/save"
)

const adoptedPetID = "01986666-aaaa-7aaa-8aaa-aaaaaaaaaaaa"

// petSpeciesContentBundle is the fixture Reputation bundle plus the fixture
// pet_species artifact (fixture-first; no epoch pins either yet).
func petSpeciesContentBundle(t *testing.T) CatalogBundle {
	t.Helper()
	bundle := reputationContentBundle(t)
	artifacts := map[string][]byte{}
	for name, data := range bundle.Artifacts {
		artifacts[name] = data
	}
	var err error
	artifacts["pet_species"], err = os.ReadFile("../../balance/testdata/pet-species/fixture-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	hash, err := save.ConstantsHashArtifacts(artifacts)
	if err != nil {
		t.Fatal(err)
	}
	declarations := pet.SpeciesDeclarations{CopyKeys: map[string]struct{}{}, CompanionKeys: map[string]struct{}{}}
	for _, key := range copykeys.All() {
		declarations.CopyKeys[key] = struct{}{}
	}
	for _, key := range copykeys.CompanionKeys() {
		declarations.CompanionKeys[key] = struct{}{}
	}
	if bundle.PetSpecies, err = pet.LoadSpeciesCatalog(artifacts["pet_species"], declarations); err != nil {
		t.Fatal(err)
	}
	bundle.ConstantsHash, bundle.Artifacts = hash, artifacts
	if !bundle.valid(hash) {
		t.Fatal("pet_species bundle is not valid")
	}
	return bundle
}

func adoptedIdentity() pet.Identity {
	return pet.Identity{SpeciesID: "pet_species.server_room_cat", Temperament: "curious", PaletteID: "pet_palette.fur_03",
		NameKey: "pet.name.server_room_cat.n07", AdoptedAtMS: 1_790_000_000_000, AdoptedAtAttendedMS: 5_400_000}
}

// AC14: v23 activates at the run boundary and New-Founder-forward, never
// synthesizing an identity.
func TestPetSpeciesOwnsFounderV23Activation(t *testing.T) {
	tree := reputationContentBundle(t)
	species := petSpeciesContentBundle(t)
	if founder, _ := species.versionFloors(); founder != 23 {
		t.Fatalf("pet_species bundle founder floor = %d", founder)
	}
	now := time.Date(2026, 9, 25, 12, 0, 0, 0, time.UTC)
	founder := foundationScopeState(t, tree.Economy, economy.ScopeFounder)
	current := foundationScopeState(t, tree.Economy, economy.ScopeCompany)
	current.RunStartedAt = now.Add(-time.Hour)
	if err := settleAndActivateFoundations(CatalogBundle{}, tree, founder, current, current); err != nil {
		t.Fatal(err)
	}
	next := foundationScopeState(t, species.Economy, economy.ScopeCompany)
	next.RunStartedAt = now
	withPet := *founder
	withPet.Pets = map[string]pet.CareState{adoptedPetID: {}}
	if err := settleAndActivateFoundations(tree, species, &withPet, current, next); !errors.Is(err, ErrInvalidEngineState) {
		t.Fatalf("v23 activation fabricated an identity for an existing pet: %v", err)
	}
	if err := settleAndActivateFoundations(tree, species, founder, current, next); err != nil {
		t.Fatal(err)
	}
	if save.VersionForState(founder) != 23 || founder.PetIdentities == nil || len(founder.PetIdentities) != 0 {
		t.Fatalf("v23 activation version=%d identities=%v", save.VersionForState(founder), founder.PetIdentities)
	}
	if err := species.ValidateFoundationState(founder); err != nil {
		t.Fatalf("activated Founder invalid: %v", err)
	}
	fresh := foundationScopeState(t, species.Economy, economy.ScopeFounder)
	company := foundationScopeState(t, species.Economy, economy.ScopeCompany)
	company.RunStartedAt = now
	if err := settleAndActivateFoundations(CatalogBundle{}, species, fresh, company, company); err != nil || save.VersionForState(fresh) != 23 || fresh.PetIdentities == nil {
		t.Fatalf("New-Founder-forward v23: %v %v", err, fresh.PetIdentities)
	}
	// Replay activation mirrors the live boundary.
	replayed := &save.State{WireVersion: 22, Pets: map[string]pet.CareState{}}
	if err := activateFounderFeatureState(replayed, species, 23, 1, nil); err != nil || replayed.PetIdentities == nil {
		t.Fatalf("replay activation: %v %+v", err, replayed.PetIdentities)
	}
	if err := activateFounderFeatureState(&save.State{WireVersion: 22, Pets: map[string]pet.CareState{}}, tree, 23, 1, nil); !errors.Is(err, ErrInvalidReplayInputs) {
		t.Fatalf("v23 activated without pet_species: %v", err)
	}
	if err := activateFounderFeatureState(&save.State{WireVersion: 22, Pets: map[string]pet.CareState{adoptedPetID: {}}}, species, 23, 1, nil); !errors.Is(err, ErrInvalidReplayInputs) {
		t.Fatalf("v23 replay activation fabricated an identity: %v", err)
	}
}

// AC2 (catalog-aware half): identities resolve under the pinned artifact.
func TestPinnedPetSpeciesValidatesIdentities(t *testing.T) {
	species := petSpeciesContentBundle(t)
	care := map[string]pet.CareState{adoptedPetID: {EvaluatedThroughAttendedMS: 5_400_000}}
	valid := &save.State{Pets: care, PetIdentities: map[string]pet.Identity{adoptedPetID: adoptedIdentity()}}
	if err := validateFounderPetIdentities(species.PetSpecies, valid); err != nil {
		t.Fatalf("valid identity rejected: %v", err)
	}
	for name, mutate := range map[string]func(*pet.Identity){
		"unknown species":         func(value *pet.Identity) { value.SpeciesID = "pet_species.robot_vacuum" },
		"unknown palette":         func(value *pet.Identity) { value.PaletteID = "pet_palette.fur_99" },
		"unknown name":            func(value *pet.Identity) { value.NameKey = "pet.name.server_room_cat.n99" },
		"adopted after watermark": func(value *pet.Identity) { value.AdoptedAtAttendedMS = 5_400_001 },
	} {
		identity := adoptedIdentity()
		mutate(&identity)
		candidate := &save.State{Pets: care, PetIdentities: map[string]pet.Identity{adoptedPetID: identity}}
		if err := validateFounderPetIdentities(species.PetSpecies, candidate); !errors.Is(err, ErrInvalidEngineState) {
			t.Fatalf("%s accepted: %v", name, err)
		}
	}
	second := "01986666-bbbb-7bbb-8bbb-bbbbbbbbbbbb"
	overCap := &save.State{Pets: map[string]pet.CareState{adoptedPetID: care[adoptedPetID], second: care[adoptedPetID]},
		PetIdentities: map[string]pet.Identity{adoptedPetID: adoptedIdentity(), second: adoptedIdentity()}}
	if err := validateFounderPetIdentities(species.PetSpecies, overCap); !errors.Is(err, ErrInvalidEngineState) {
		t.Fatalf("over-cap identities accepted: %v", err)
	}
	if err := validateFounderPetIdentities(nil, valid); !errors.Is(err, ErrInvalidEngineState) {
		t.Fatalf("identities accepted without pinned pet_species: %v", err)
	}
}

// AC8: the Company replay carry holds pet_identities at Founder floor >= 23.
func TestFounderCarryHoldsPetIdentitiesAtFloor23(t *testing.T) {
	species := petSpeciesContentBundle(t)
	carry := replayFounderCarry{FounderRevision: 1, ReputationLevel: 3, NetworkSlots: []save.NetworkSlot{}, LedgerFactKinds: []string{},
		AchievementsEarnedLifetime: []string{}, FounderExtensions: &replayFounderExtensions{}}
	spent, unlock, owned := int64(0), int64(0), []string{}
	carry.FounderExtensions.ReputationSpent, carry.FounderExtensions.ReputationUnlockPPM, carry.FounderExtensions.ReputationNodesOwned = &spent, &unlock, &owned
	if _, err := stateFromFounderCarry(carry, species); !errors.Is(err, ErrInvalidReplayInputs) {
		t.Fatalf("v23 carry reconstructed without pet identities: %v", err)
	}
	identities := map[string]pet.Identity{}
	carry.FounderExtensions.PetIdentities = &identities
	if !founderCarryShapeValid(t, carry, species, 10) {
		t.Fatal("v10 carry with identities rejected")
	}
	if founderCarryShapeValid(t, carry, species, 9) {
		t.Fatal("pre-v10 carry accepted pet identities at floor 23")
	}
	tree := reputationContentBundle(t)
	if founderCarryShapeValid(t, carry, tree, 10) {
		t.Fatal("carry with identities accepted below floor 23")
	}
}

func founderCarryShapeValid(t *testing.T, carry replayFounderCarry, catalogs CatalogBundle, wireVersion int) bool {
	t.Helper()
	carry.FounderConstantsHash = catalogs.ConstantsHash
	extensions := *carry.FounderExtensions
	extensions.MinigameRatings = map[string]save.MinigameRatingState{}
	extensions.MinigameOfflineQuality = map[string]save.MinigameOfflineQualityState{}
	extensions.Pets = map[string]pet.CareState{}
	extensions.FiscalGeneratorLevels = map[string]int64{}
	extensions.FiscalUnlocks = []string{}
	extensions.SoulExhaustedSourceIDs = []string{}
	carry.FounderExtensions = &extensions
	return validFounderCarry(carry, wireVersion, catalogs)
}

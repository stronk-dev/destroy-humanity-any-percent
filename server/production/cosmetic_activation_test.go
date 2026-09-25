package production

import (
	"errors"
	"os"
	"testing"
	"time"

	"cloud-clicker/server/cosmetic"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/pet"
	"cloud-clicker/server/save"
)

// cosmeticsContentBundle is the fixture pet_species bundle plus the fixture
// cosmetics artifact (fixture-first; no epoch pins either yet).
func cosmeticsContentBundle(t *testing.T) CatalogBundle {
	t.Helper()
	bundle := petSpeciesContentBundle(t)
	artifacts := map[string][]byte{}
	for name, data := range bundle.Artifacts {
		artifacts[name] = data
	}
	var err error
	artifacts["cosmetics"], err = os.ReadFile("../../balance/testdata/cosmetics/fixture-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	hash, err := save.ConstantsHashArtifacts(artifacts)
	if err != nil {
		t.Fatal(err)
	}
	if bundle.Cosmetics, err = cosmetic.Load(artifacts["cosmetics"]); err != nil {
		t.Fatal(err)
	}
	bundle.ConstantsHash, bundle.Artifacts = hash, artifacts
	if !bundle.valid(hash) {
		t.Fatal("cosmetics bundle is not valid")
	}
	return bundle
}

// AC4: v24 activates only at the run boundary (and New-Founder-forward) with
// empty cosmetics; the replay activation mirrors the live boundary.
func TestCosmeticsOwnsFounderV24Activation(t *testing.T) {
	species := petSpeciesContentBundle(t)
	shop := cosmeticsContentBundle(t)
	if founder, _ := shop.versionFloors(); founder != 24 {
		t.Fatalf("cosmetics bundle founder floor = %d", founder)
	}
	now := time.Date(2026, 9, 25, 12, 0, 0, 0, time.UTC)
	founder := foundationScopeState(t, species.Economy, economy.ScopeFounder)
	current := foundationScopeState(t, species.Economy, economy.ScopeCompany)
	current.RunStartedAt = now.Add(-time.Hour)
	if err := settleAndActivateFoundations(CatalogBundle{}, species, founder, current, current); err != nil {
		t.Fatal(err)
	}
	if save.VersionForState(founder) != 23 || founder.Cosmetics != nil {
		t.Fatalf("pre-activation founder version=%d cosmetics=%v", save.VersionForState(founder), founder.Cosmetics)
	}
	next := foundationScopeState(t, shop.Economy, economy.ScopeCompany)
	next.RunStartedAt = now
	preexisting := *founder
	preexisting.Cosmetics = cosmetic.NewState()
	if err := settleAndActivateFoundations(species, shop, &preexisting, current, next); !errors.Is(err, ErrInvalidEngineState) {
		t.Fatalf("v24 activation accepted pre-existing cosmetics: %v", err)
	}
	if err := settleAndActivateFoundations(species, shop, founder, current, next); err != nil {
		t.Fatal(err)
	}
	if save.VersionForState(founder) != 24 || !founder.Cosmetics.Equal(cosmetic.NewState()) {
		t.Fatalf("v24 activation version=%d cosmetics=%+v", save.VersionForState(founder), founder.Cosmetics)
	}
	if err := shop.ValidateFoundationState(founder); err != nil {
		t.Fatalf("activated Founder invalid: %v", err)
	}
	fresh := foundationScopeState(t, shop.Economy, economy.ScopeFounder)
	company := foundationScopeState(t, shop.Economy, economy.ScopeCompany)
	company.RunStartedAt = now
	if err := settleAndActivateFoundations(CatalogBundle{}, shop, fresh, company, company); err != nil || save.VersionForState(fresh) != 24 || fresh.Cosmetics == nil {
		t.Fatalf("New-Founder-forward v24: %v %v", err, fresh.Cosmetics)
	}
	replayed := &save.State{WireVersion: 23, Pets: map[string]pet.CareState{}, PetIdentities: map[string]pet.Identity{}}
	if err := activateFounderFeatureState(replayed, shop, 24, 1, nil); err != nil || !replayed.Cosmetics.Equal(cosmetic.NewState()) {
		t.Fatalf("replay activation: %v %+v", err, replayed.Cosmetics)
	}
	if err := activateFounderFeatureState(&save.State{WireVersion: 23, Pets: map[string]pet.CareState{}, PetIdentities: map[string]pet.Identity{}}, species, 24, 1, nil); !errors.Is(err, ErrInvalidReplayInputs) {
		t.Fatalf("v24 activated without cosmetics: %v", err)
	}
}

// AC3 (catalog-aware half) and AC8: pinned validation, and the carry.
func TestPinnedCosmeticsValidatesAndCarries(t *testing.T) {
	shop := cosmeticsContentBundle(t)
	pets := map[string]pet.CareState{adoptedPetID: {}}
	valid := &save.State{Pets: pets, Cosmetics: &cosmetic.State{Owned: []string{"horse_armor"}, Equipped: map[string]string{adoptedPetID: "horse_armor"}}}
	if err := validateFounderCosmetics(shop.Cosmetics, valid); err != nil {
		t.Fatalf("valid cosmetics rejected: %v", err)
	}
	for name, candidate := range map[string]*save.State{
		"unknown owned id":     {Pets: pets, Cosmetics: &cosmetic.State{Owned: []string{"zebra_armor"}, Equipped: map[string]string{}}},
		"equipped not owned":   {Pets: pets, Cosmetics: &cosmetic.State{Owned: []string{}, Equipped: map[string]string{adoptedPetID: "horse_armor"}}},
		"equipped unknown pet": {Pets: map[string]pet.CareState{}, Cosmetics: &cosmetic.State{Owned: []string{"horse_armor"}, Equipped: map[string]string{adoptedPetID: "horse_armor"}}},
		"missing at floor 24":  {Pets: pets},
	} {
		if err := validateFounderCosmetics(shop.Cosmetics, candidate); !errors.Is(err, ErrInvalidEngineState) {
			t.Fatalf("%s accepted: %v", name, err)
		}
	}
	if err := validateFounderCosmetics(nil, valid); !errors.Is(err, ErrInvalidEngineState) {
		t.Fatalf("cosmetics accepted without the pinned artifact: %v", err)
	}
	// AC8: the Exit replay output carries cosmetics byte-identically.
	target := &save.State{}
	replayed := &save.State{WireVersion: 24, ExitHistory: []save.ExitRecord{{}}, Cosmetics: valid.Cosmetics.Clone()}
	if err := applyFounderReplayOutput(target, replayed); err != nil || !target.Cosmetics.Equal(valid.Cosmetics) {
		t.Fatalf("Exit replay output dropped cosmetics: %v %+v", err, target.Cosmetics)
	}
	// The Company carry holds cosmetics exactly at Founder floor >= 24.
	carry := replayFounderCarry{FounderRevision: 1, ReputationLevel: 3, NetworkSlots: []save.NetworkSlot{}, LedgerFactKinds: []string{},
		AchievementsEarnedLifetime: []string{}, FounderExtensions: &replayFounderExtensions{}}
	spent, unlock, owned := int64(0), int64(0), []string{}
	identities := map[string]pet.Identity{}
	carry.FounderExtensions.ReputationSpent, carry.FounderExtensions.ReputationUnlockPPM, carry.FounderExtensions.ReputationNodesOwned = &spent, &unlock, &owned
	carry.FounderExtensions.PetIdentities = &identities
	if _, err := stateFromFounderCarry(carry, shop); !errors.Is(err, ErrInvalidReplayInputs) {
		t.Fatalf("v24 carry reconstructed without cosmetics: %v", err)
	}
	carry.FounderExtensions.Cosmetics = cosmetic.NewState()
	if !founderCarryShapeValid(t, carry, shop, 11) {
		t.Fatal("v11 carry with cosmetics rejected")
	}
	if founderCarryShapeValid(t, carry, shop, 10) {
		t.Fatal("pre-v11 carry accepted cosmetics at floor 24")
	}
	if founderCarryShapeValid(t, carry, petSpeciesContentBundle(t), 11) {
		t.Fatal("carry with cosmetics accepted below floor 24")
	}
}

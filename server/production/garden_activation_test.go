package production

import (
	"errors"
	"os"
	"testing"
	"time"

	"cloud-clicker/server/copykeys"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/fiscal"
	"cloud-clicker/server/garden"
	"cloud-clicker/server/minigame"
	"cloud-clicker/server/save"
)

// gardenContentBundle is the fixture cosmetics bundle, with the fixture Fiscal
// artifact (production Fiscal plus the minigame.server_garden unlock row), plus
// the fixture server_garden artifact. Fixture-first: no epoch pins either.
func gardenContentBundle(t *testing.T) CatalogBundle {
	t.Helper()
	bundle := cosmeticsContentBundle(t)
	artifacts := map[string][]byte{}
	for name, data := range bundle.Artifacts {
		artifacts[name] = data
	}
	var err error
	if artifacts["fiscal"], err = os.ReadFile("../../balance/testdata/server-garden/fiscal-fixture-v1.json"); err != nil {
		t.Fatal(err)
	}
	if artifacts[garden.ArtifactName], err = os.ReadFile("../../balance/testdata/server-garden/fixture-v1.json"); err != nil {
		t.Fatal(err)
	}
	if bundle.Fiscal, err = fiscal.LoadCatalog(artifacts["fiscal"], bundle.Economy); err != nil {
		t.Fatal(err)
	}
	if bundle.Garden, err = garden.LoadCatalog(artifacts[garden.ArtifactName], gardenDeclarationsForTest(bundle)); err != nil {
		t.Fatal(err)
	}
	hash, err := save.ConstantsHashArtifacts(artifacts)
	if err != nil {
		t.Fatal(err)
	}
	bundle.ConstantsHash, bundle.Artifacts = hash, artifacts
	if !bundle.valid(hash) {
		t.Fatal("garden bundle is not valid")
	}
	return bundle
}

func gardenDeclarationsForTest(bundle CatalogBundle) garden.Declarations {
	declarations := garden.Declarations{CopyKeys: map[string]struct{}{}, ResourceIDs: map[string]struct{}{},
		FiscalUnlockIDs: map[string]struct{}{}, FiscalGeneratorIDs: map[string]struct{}{}}
	for _, key := range copykeys.All() {
		declarations.CopyKeys[key] = struct{}{}
	}
	for _, resource := range bundle.Economy.Resources() {
		declarations.ResourceIDs[resource.ID] = struct{}{}
	}
	for _, row := range bundle.Fiscal.UnlockRows() {
		declarations.FiscalUnlockIDs[row.UnlockID] = struct{}{}
	}
	for _, row := range bundle.Fiscal.GeneratorLevelRows() {
		declarations.FiscalGeneratorIDs[row.GeneratorID] = struct{}{}
	}
	declarations.ValidatePayout = minigame.GardenPayoutValidator(declarations.ResourceIDs, declarations.CopyKeys)
	return declarations
}

// AC2 (bundle half): server_garden pins Founder v25 and requires cosmetics
// (and so the whole lower chain through minigame_api).
func TestGardenBundleChainAndFloor(t *testing.T) {
	grown := gardenContentBundle(t)
	if founder, _ := grown.versionFloors(); founder != 25 {
		t.Fatalf("garden bundle founder floor = %d", founder)
	}
	withoutCosmetics := grown
	withoutCosmetics.Cosmetics = nil
	artifacts := map[string][]byte{}
	for name, data := range grown.Artifacts {
		if name != "cosmetics" {
			artifacts[name] = data
		}
	}
	hash, err := save.ConstantsHashArtifacts(artifacts)
	if err != nil {
		t.Fatal(err)
	}
	withoutCosmetics.ConstantsHash, withoutCosmetics.Artifacts = hash, artifacts
	if withoutCosmetics.valid(hash) {
		t.Fatal("a garden bundle without cosmetics is valid")
	}
	forged := cosmeticsContentBundle(t)
	forged.Garden = grown.Garden
	if forged.valid(forged.ConstantsHash) {
		t.Fatal("a bundle carrying Garden without its artifact bytes is valid")
	}
}

// AC2/AC11: v25 activates only at the run boundary and at New-Founder
// initialization, with an empty garden holding every starter; a pinned-to-
// pinned boundary (every later Exit) leaves the garden byte-identical.
func TestGardenOwnsFounderV25Activation(t *testing.T) {
	shop := cosmeticsContentBundle(t)
	grown := gardenContentBundle(t)
	now := time.Date(2026, 9, 25, 12, 0, 0, 0, time.UTC)
	founder := foundationScopeState(t, shop.Economy, economy.ScopeFounder)
	current := foundationScopeState(t, shop.Economy, economy.ScopeCompany)
	current.RunStartedAt = now.Add(-time.Hour)
	if err := settleAndActivateFoundations(CatalogBundle{}, shop, founder, current, current); err != nil {
		t.Fatal(err)
	}
	if save.VersionForState(founder) != 24 || founder.ServerGarden != nil {
		t.Fatalf("pre-activation founder version=%d garden=%v", save.VersionForState(founder), founder.ServerGarden)
	}
	next := foundationScopeState(t, grown.Economy, economy.ScopeCompany)
	next.RunStartedAt = now
	preexisting := *founder
	preexisting.ServerGarden = garden.NewState(grown.Garden)
	if err := settleAndActivateFoundations(shop, grown, &preexisting, current, next); !errors.Is(err, ErrInvalidEngineState) {
		t.Fatalf("v25 activation accepted a pre-existing garden: %v", err)
	}
	if err := settleAndActivateFoundations(shop, grown, founder, current, next); err != nil {
		t.Fatal(err)
	}
	if save.VersionForState(founder) != 25 || !founder.ServerGarden.Equal(garden.NewState(grown.Garden)) {
		t.Fatalf("v25 activation version=%d garden=%+v", save.VersionForState(founder), founder.ServerGarden)
	}
	if err := grown.ValidateFoundationState(founder); err != nil {
		t.Fatalf("activated Founder invalid: %v", err)
	}
	// A later pinned→pinned boundary (every subsequent Exit) never touches it.
	salt, anchor, effect := "0123456789abcdef", int64(5_000), int64(1_000_000)
	founder.ServerGarden.SaltHex, founder.ServerGarden.TickAnchorWallMS, founder.ServerGarden.TickSeq = &salt, &anchor, 9
	founder.ServerGarden.Plots = []garden.Plot{{Row: 0, Col: 0, SpeciesID: "strain_a", AgeTicks: 3, MaturedEffectPPM: &effect}}
	before := founder.ServerGarden.Clone()
	later := foundationScopeState(t, grown.Economy, economy.ScopeCompany)
	later.RunStartedAt = now.Add(time.Hour)
	if err := settleAndActivateFoundations(grown, grown, founder, next, later); err != nil {
		t.Fatal(err)
	}
	if !founder.ServerGarden.Equal(before) {
		t.Fatalf("a run boundary changed the garden: %+v", founder.ServerGarden)
	}
	fresh := foundationScopeState(t, grown.Economy, economy.ScopeFounder)
	company := foundationScopeState(t, grown.Economy, economy.ScopeCompany)
	company.RunStartedAt = now
	if err := settleAndActivateFoundations(CatalogBundle{}, grown, fresh, company, company); err != nil || save.VersionForState(fresh) != 25 ||
		!fresh.ServerGarden.Equal(garden.NewState(grown.Garden)) {
		t.Fatalf("New-Founder activation version=%d garden=%+v err=%v", save.VersionForState(fresh), fresh.ServerGarden, err)
	}
	if err := shop.ValidateFoundationState(fresh); err == nil {
		t.Fatal("a v25 Founder validated under a bundle that does not pin server_garden")
	}
}

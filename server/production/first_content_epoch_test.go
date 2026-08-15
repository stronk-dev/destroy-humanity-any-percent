package production

import (
	"testing"
	"time"

	"cloud-clicker/server/achievements"
	"cloud-clicker/server/activeplay"
	"cloud-clicker/server/copykeys"
	"cloud-clicker/server/curriculum"
	"cloud-clicker/server/doctrine"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/epochseed"
	"cloud-clicker/server/fiscal"
	"cloud-clicker/server/meters"
	"cloud-clicker/server/minigame"
	"cloud-clicker/server/minigameapi"
	"cloud-clicker/server/pet"
	"cloud-clicker/server/pitch"
	"cloud-clicker/server/relevancepolicy"
	"cloud-clicker/server/save"
	"cloud-clicker/server/soul"
)

func activeContentBundle(t *testing.T) CatalogBundle {
	t.Helper()
	bundle, err := epochseed.Load("../..")
	if err != nil {
		t.Fatal(err)
	}
	if bundle.Seed.CurrentEpochID != 8 || bundle.Hash != "sha256:baa890501b2864d14cc0238d633a562cb8c6fca406190487831e0c447af128f6" {
		t.Fatalf("active epoch=%d hash=%s", bundle.Seed.CurrentEpochID, bundle.Hash)
	}
	return loadCompleteReplayTestBundle(t, bundle.Hash, bundle.Artifacts)
}

func loadCompleteReplayTestBundle(t *testing.T, hash string, artifacts map[string][]byte) CatalogBundle {
	t.Helper()
	bundle := loadReplayTestBundle(t, hash, artifacts)
	var err error
	bundle.Meters, err = meters.LoadCatalog(artifacts["meters"])
	if err != nil {
		t.Fatal(err)
	}
	resourceIDs := make([]string, 0, len(bundle.Economy.Resources()))
	for _, resource := range bundle.Economy.Resources() {
		resourceIDs = append(resourceIDs, resource.ID)
	}
	if err := bundle.Meters.ValidateResourceSeparation(resourceIDs); err != nil {
		t.Fatal(err)
	}
	bundle.Achievements, err = achievements.LoadCatalog(artifacts["achievements"], FoundationAchievementRegistry(bundle.Economy))
	if err != nil {
		t.Fatal(err)
	}
	bundle.Doctrines, err = doctrine.LoadCatalog(artifacts["doctrines"])
	if err != nil {
		t.Fatal(err)
	}
	if err := bundle.Doctrines.ValidateRoutes(bundle.Routes); err != nil {
		t.Fatalf("doctrine routes: %v", err)
	}
	bundle.Minigames, err = minigame.LoadCatalog(artifacts["minigames"])
	if err != nil {
		t.Fatal(err)
	}
	bundle.Pets, err = pet.LoadCatalog(artifacts["pets"])
	if err != nil {
		t.Fatal(err)
	}
	bundle.Fiscal, err = fiscal.LoadCatalog(artifacts["fiscal"], bundle.Economy)
	if err != nil {
		t.Fatal(err)
	}
	keys := make(map[string]struct{})
	for _, key := range copykeys.All() {
		keys[key] = struct{}{}
	}
	bundle.Soul, err = soul.LoadCatalog(artifacts["soul"], soul.Declarations{CopyKeys: keys, EpochSeeded: true,
		CatchupCeilingMS: bundle.Prestige.CatchupCeilingMS})
	if err != nil {
		t.Fatal(err)
	}
	bundle.Pitch, err = pitch.LoadCatalog(artifacts["pitch"], pitch.Declarations{CopyKeys: keys})
	if err != nil {
		t.Fatal(err)
	}
	bundle.MinigameAPI, err = minigameapi.LoadCatalog(artifacts["minigame_api"])
	if err != nil {
		t.Fatal(err)
	}
	if data := artifacts["opportunities"]; len(data) != 0 {
		bundle.Opportunities, err = activeplay.LoadCatalog(data, bundle.Economy)
		if err != nil {
			t.Fatal(err)
		}
	}
	if data := artifacts["relevance"]; len(data) != 0 {
		bundle.Relevance, err = relevancepolicy.Load(data, bundle.Economy, bundle.Routes, false)
		if err != nil {
			t.Fatal(err)
		}
	}
	if data := artifacts["curriculum"]; len(data) != 0 {
		gateIDs := map[string]struct{}{}
		for _, gate := range bundle.Routes.Gates() {
			gateIDs[gate.ID] = struct{}{}
		}
		bundle.Curriculum, err = curriculum.Load(data, curriculum.Declarations{Economy: bundle.Economy, CopyKeys: keys, GateIDs: gateIDs})
		if err != nil {
			t.Fatal(err)
		}
	}
	definition, ok := bundle.Minigames.Definition("pitch")
	if !ok || !bundle.MinigameAPI.SupportsTenant(definition.MinigameID, definition.EngineRef, definition.EngineVersion) || !bundle.valid(hash) {
		t.Fatal("complete first-content bundle is internally inconsistent")
	}
	return bundle
}

func TestFirstContentEpochActivatesAtNewRunBoundary(t *testing.T) {
	legacy := epoch5TestBundle(t)
	active := activeContentBundle(t)
	now := time.Date(2026, 8, 9, 12, 0, 0, 0, time.UTC)
	founder := foundationScopeState(t, legacy.Economy, economy.ScopeFounder)
	company := foundationScopeState(t, legacy.Economy, economy.ScopeCompany)
	company.RunStartedAt = now.Add(-time.Hour)
	newCompany := foundationScopeState(t, active.Economy, economy.ScopeCompany)
	newCompany.RunStartedAt = now

	if err := settleAndActivateFoundations(legacy, active, founder, company, newCompany); err != nil {
		t.Fatal(err)
	}
	if save.VersionForState(company) != save.CurrentVersion || save.VersionForState(newCompany) != 18 || save.VersionForState(founder) != 21 {
		t.Fatalf("versions old_company=%d new_company=%d founder=%d", save.VersionForState(company), save.VersionForState(newCompany), save.VersionForState(founder))
	}
	if len(newCompany.MeterValues) != 11 || len(newCompany.AchievementsEarnedRun) != 0 || newCompany.AchievementScoreRun != 0 ||
		len(founder.AchievementsEarnedLifetime) != 0 || founder.AchievementScoreLifetime != 0 || len(founder.Pets) != 0 ||
		founder.FiscalPeriodOpenedWallMS != now.UnixMilli() || founder.Soul != active.Soul.Policy.Initial || founder.MinigameSessionSeq != 0 {
		t.Fatalf("activation company=%+v founder=%+v", newCompany, founder)
	}
	if _, ok := founder.MinigameRatings["pitch"]; !ok || len(founder.MinigameOfflineQuality) != 1 || len(founder.FiscalUnlocks) != 0 {
		t.Fatalf("Founder content activation incomplete: %+v", founder)
	}
	if err := active.ValidateFoundationState(founder); err != nil {
		t.Fatalf("Founder activation invalid: %v", err)
	}
	if err := active.ValidateFoundationState(newCompany); err != nil {
		t.Fatalf("Company activation invalid: %v", err)
	}
}

func TestFirstContentEpochInitializesFreshFounderWithFullSet(t *testing.T) {
	active := activeContentBundle(t)
	now := time.Date(2026, 8, 9, 13, 0, 0, 0, time.UTC)
	founder := foundationScopeState(t, active.Economy, economy.ScopeFounder)
	company := foundationScopeState(t, active.Economy, economy.ScopeCompany)
	company.RunStartedAt = now
	company.RunSeq = 1
	initializer := FounderInitializer{Catalogs: fixedReplayBundleResolver{bundle: active}}
	frozen, err := initializer.InitializeNewFounder(active.ConstantsHash, "01986666-f101-7000-8000-000000000002", now, founder, company)
	if err != nil {
		t.Fatal(err)
	}
	if save.VersionForState(founder) != 21 || save.VersionForState(company) != 18 || len(frozen) != len(active.Fiscal.GeneratorLevelRows())+1 ||
		len(founder.Pets) != 0 || founder.Soul != active.Soul.Policy.Initial || len(founder.MinigameRatings) != 1 ||
		company.NextOpportunityAttendedMS <= 0 || company.PendingOpportunity != nil || company.ActiveBuffs == nil {
		t.Fatalf("fresh Founder founder=%+v company=%+v frozen=%+v", founder, company, frozen)
	}
	if err := active.ValidateFoundationState(founder); err != nil {
		t.Fatal(err)
	}
	if err := active.ValidateFoundationState(company); err != nil {
		t.Fatal(err)
	}
}

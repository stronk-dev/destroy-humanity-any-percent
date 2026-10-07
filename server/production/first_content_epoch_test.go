package production

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"cloud-clicker/server/achievements"
	"cloud-clicker/server/activeplay"
	"cloud-clicker/server/copykeys"
	"cloud-clicker/server/curriculum"
	"cloud-clicker/server/decimal"
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

const firstContentConstantsHash = "sha256:1a4463bcf67440ce1ba01e6c6eb850c0614329cac63064ef07725d042c7cf21a"

// Historical acceptance must not follow the deploy-current artifact paths.
// Read the ratified manifest's literal sources, checking each byte pin as well
// as the full bundle's original accepted identity.
func firstContentEpochBundle(t *testing.T) CatalogBundle {
	t.Helper()
	data, err := os.ReadFile("../../planning/first-content-epoch/promotion-manifest.candidate.v1.json")
	if err != nil {
		t.Fatal(err)
	}
	var manifest struct {
		SchemaVersion int    `json:"schema_version"`
		Status        string `json:"status"`
		ConstantsHash string `json:"constants_hash"`
		Artifacts     []struct {
			Name       string `json:"name"`
			SourcePath string `json:"source_path"`
			SHA256     string `json:"sha256"`
		} `json:"artifacts"`
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatal(err)
	}
	if manifest.SchemaVersion != 1 || manifest.Status != "ratified" || manifest.ConstantsHash != firstContentConstantsHash || len(manifest.Artifacts) != 16 {
		t.Fatal("historical First Content manifest identity changed")
	}
	artifacts := make(map[string][]byte, len(manifest.Artifacts))
	for _, row := range manifest.Artifacts {
		data, err := os.ReadFile(filepath.Join("../..", row.SourcePath))
		if err != nil {
			t.Fatal(err)
		}
		if fmt.Sprintf("%x", sha256.Sum256(data)) != row.SHA256 {
			t.Fatalf("historical %s source no longer matches its ratified pin", row.Name)
		}
		artifacts[row.Name] = data
	}
	hash, err := save.ConstantsHashArtifacts(artifacts)
	if err != nil || hash != firstContentConstantsHash {
		t.Fatalf("historical First Content bundle=%s err=%v", hash, err)
	}
	seed, err := epochseed.Load("../..")
	if err != nil {
		t.Fatal(err)
	}
	accepted := false
	for _, epoch := range seed.Seed.Epochs {
		if epoch.ID == 6 {
			accepted = epochseed.Accepts(epoch, hash)
		}
	}
	if !accepted {
		t.Fatal("the original First Content hash is no longer accepted by epoch 6")
	}
	return loadCompleteReplayTestBundle(t, hash, artifacts)
}

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
	testContentActivation(t, firstContentEpochBundle(t), 17)
}

func TestCurrentContentActivatesAtNewRunBoundary(t *testing.T) {
	testContentActivation(t, activeContentBundle(t), 18)
}

func testContentActivation(t *testing.T, active CatalogBundle, companyVersion int) {
	t.Helper()
	legacy := epoch5TestBundle(t)
	now := time.Date(2026, 8, 9, 12, 0, 0, 0, time.UTC)
	founder := foundationScopeState(t, legacy.Economy, economy.ScopeFounder)
	company := replayFixtureState(t, legacy.Economy, now.Add(-time.Hour))
	company.RunStartedAt = now.Add(-time.Hour)
	before := mustEncodeState(t, company)
	newCompany := replayFixtureState(t, active.Economy, now)
	newCompany.RunStartedAt = now

	if err := settleAndActivateFoundations(legacy, active, founder, company, newCompany); err != nil {
		t.Fatal(err)
	}
	if save.VersionForState(company) != save.CurrentVersion || save.VersionForState(newCompany) != companyVersion || save.VersionForState(founder) != 21 {
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
	if !bytes.Equal(before, mustEncodeState(t, company)) {
		t.Fatal("new content rewrote the ending Company's pinned state")
	}
	if _, exists := company.Ledger.Balance("company.permits"); exists {
		t.Fatal("permits retroactively appeared in the epoch-5 run")
	}
	assertNewRunPermits(t, newCompany)
}

func TestFirstContentEpochInitializesFreshFounderWithFullSet(t *testing.T) {
	testFreshContentFounder(t, firstContentEpochBundle(t), 17, false)
}

func TestCurrentContentInitializesFreshFounderWithFullSet(t *testing.T) {
	testFreshContentFounder(t, activeContentBundle(t), 18, true)
}

func testFreshContentFounder(t *testing.T, active CatalogBundle, companyVersion int, activePlay bool) {
	t.Helper()
	now := time.Date(2026, 8, 9, 13, 0, 0, 0, time.UTC)
	founder := foundationScopeState(t, active.Economy, economy.ScopeFounder)
	company := replayFixtureState(t, active.Economy, now)
	company.RunStartedAt = now
	company.RunSeq = 1
	initializer := FounderInitializer{Catalogs: fixedReplayBundleResolver{bundle: active}}
	frozen, err := initializer.InitializeNewFounder(active.ConstantsHash, "01986666-f101-7000-8000-000000000002", now, founder, company)
	if err != nil {
		t.Fatal(err)
	}
	if save.VersionForState(founder) != 21 || save.VersionForState(company) != companyVersion || len(frozen) != len(active.Fiscal.GeneratorLevelRows())+1 ||
		len(founder.Pets) != 0 || founder.Soul != active.Soul.Policy.Initial || len(founder.MinigameRatings) != 1 {
		t.Fatalf("fresh Founder founder=%+v company=%+v frozen=%+v", founder, company, frozen)
	}
	if activePlay {
		if company.NextOpportunityAttendedMS <= 0 || company.PendingOpportunity != nil || company.ActiveBuffs == nil {
			t.Fatal("current content lost its initialized Active Play state")
		}
	} else if active.Opportunities != nil || company.NextOpportunityAttendedMS != 0 || company.PendingOpportunity != nil || company.ActiveBuffs != nil {
		t.Fatal("historical epoch 6 acquired later Active Play content")
	}
	if err := active.ValidateFoundationState(founder); err != nil {
		t.Fatal(err)
	}
	if err := active.ValidateFoundationState(company); err != nil {
		t.Fatal(err)
	}
	assertNewRunPermits(t, company)
}

func assertNewRunPermits(t *testing.T, company *save.State) {
	t.Helper()
	permits, exists := company.Ledger.Balance("company.permits")
	if !exists || permits.String() != "0" {
		t.Fatalf("new-run permits=%s exists=%t", permits.String(), exists)
	}
	legalDepartments, exists := company.GeneratorCounts["generator.legal_dept"]
	if !exists || legalDepartments != 0 {
		t.Fatalf("new-run legal departments=%d exists=%t", legalDepartments, exists)
	}
}

func TestFirstContentEpochExitPreservesOldResourceUniverse(t *testing.T) {
	current, next := epoch5TestBundle(t), firstContentEpochBundle(t)
	current.Next = &next
	now := time.Date(2026, 8, 9, 14, 0, 0, 0, time.UTC)
	company := replayFixtureState(t, current.Economy, now.Add(-20*time.Minute))
	company.Tier = 3
	company.LifetimeValue = decimal.New(27, 12)
	terms := json.RawMessage(`{"market_modifier_ppm":1100000,"payout_preview":{"reputation_delta":5,"network_slot_unlocks":[],"route_knowledge":0,"clout_reach_note":"clout.reach.preserved"}}`)
	company.OfferState = &save.ExitOfferState{OfferID: "01986666-0600-7000-8000-000000000600", ExitType: "acquisition",
		TermsJSON: terms, SpawnedAt: now.Add(-time.Minute), ExpiresAt: now.Add(time.Minute)}
	pre := mustEncodeState(t, company)
	request, err := ParseIntent([]byte(`{"intent_id":"01986666-0600-7000-8000-000000000601","kind":"accept_exit_offer","expected_revision":1,"expected_founder_revision":2,"offer_id":"01986666-0600-7000-8000-000000000600"}`))
	if err != nil {
		t.Fatal(err)
	}
	carry := replayFounderCarry{FounderRevision: 2, FounderConstantsHash: current.ConstantsHash, ReputationLevel: 1, Notoriety: 17,
		NetworkSlots: []save.NetworkSlot{}, LedgerFactKinds: []string{}, ExitHistoryCount: 1}
	command := save.ReplayCommand{IntentID: request.IntentID, CompanyStreamID: "01986666-1600-7000-8000-000000000001",
		FounderID: "01986666-2600-7000-8000-000000000001", Revision: 1, RunSeq: 1, RunLogSeq: 1}
	inputs, err := buildReplayInputs(replayBuild{Command: command, Mode: ModeOnline, Now: now, IntentKind: request.Kind,
		RouteContextVersion: current.Routes.ContextVersion(), FounderCarry: &carry, Terminal: true, ExecutedRouteIDs: []string{},
		SelectedExitType: "acquisition", SelectedTerms: terms, NextConstantsHash: next.ConstantsHash})
	if err != nil {
		t.Fatal(err)
	}
	transition, err := ApplyLoggedExit(company, request.CanonicalPayload, current, inputs)
	if err != nil {
		t.Fatal(err)
	}
	decision := transition.Decision
	if decision.Outcome != save.IntentApplied || decision.NewConstantsHash != firstContentConstantsHash ||
		save.VersionForState(company) != 14 || save.VersionForState(decision.NewCompanyState) != 17 ||
		save.VersionForState(transition.Founder) != 21 || decision.NewCompanyState.RunSeq != 2 {
		t.Fatal("epoch-5 Exit did not activate the exact epoch-6 floors/pin at the next run")
	}
	if _, exists := company.Ledger.Balance("company.permits"); exists || len(company.Ledger.Snapshot()) != 1 {
		t.Fatal("ending run acquired the new resource universe")
	}
	assertNewRunPermits(t, decision.NewCompanyState)
	if err := next.ValidateFoundationState(transition.Founder); err != nil {
		t.Fatal(err)
	}
	if err := next.ValidateFoundationState(decision.NewCompanyState); err != nil {
		t.Fatal(err)
	}
	final := mustEncodeState(t, company)
	if _, err := save.RestoreState(final, 14, current.Economy, economy.ScopeCompany, time.Time{}); err != nil {
		t.Fatalf("ending run cannot restore under its own pinned catalog: %v", err)
	}
	if _, err := save.RestoreState(final, 14, next.Economy, economy.ScopeCompany, time.Time{}); err == nil {
		t.Fatal("old run silently restored under the new resource universe")
	}
	restored, err := save.RestoreState(pre, 14, current.Economy, economy.ScopeCompany, time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	replayed, err := ApplyLoggedExit(restored, request.CanonicalPayload, current, inputs)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(decision.Receipt, replayed.Decision.Receipt) || !bytes.Equal(final, mustEncodeState(t, restored)) ||
		canonicalFixtureValue(t, replayFounderOutput(transition.Founder, carry)) != canonicalFixtureValue(t, replayFounderOutput(replayed.Founder, carry)) ||
		!bytes.Equal(mustEncodeState(t, decision.NewCompanyState), mustEncodeState(t, replayed.Decision.NewCompanyState)) ||
		canonicalFixtureValue(t, fixtureEvents(decision.FounderEvents)) != canonicalFixtureValue(t, fixtureEvents(replayed.Decision.FounderEvents)) ||
		canonicalFixtureValue(t, fixtureEvents(decision.CompanyEndedEvents)) != canonicalFixtureValue(t, fixtureEvents(replayed.Decision.CompanyEndedEvents)) ||
		canonicalFixtureValue(t, fixtureEvents(decision.CompanyStartedEvents)) != canonicalFixtureValue(t, fixtureEvents(replayed.Decision.CompanyStartedEvents)) {
		t.Fatal("historical cross-epoch Exit did not replay byte-identically")
	}
}

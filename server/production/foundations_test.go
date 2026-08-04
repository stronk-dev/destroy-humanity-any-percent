package production

import (
	"encoding/json"
	"os"
	"testing"

	"cloud-clicker/server/achievements"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/epochseed"
	"cloud-clicker/server/meters"
	"cloud-clicker/server/save"
)

type achievementParityFixture struct {
	Baseline json.RawMessage `json:"baseline"`
	Registry struct {
		CopyKeys          []string            `json:"copy_keys"`
		GeneratorIDs      []string            `json:"generator_ids"`
		EventKinds        []string            `json:"event_kinds"`
		ResourceIDs       []string            `json:"resource_ids"`
		RunCounters       []string            `json:"run_counters"`
		CareerCounters    []string            `json:"career_counters"`
		ProvenanceSources map[string][]string `json:"provenance_sources"`
	} `json:"registry"`
}

func foundationAchievementRegistry(fixture achievementParityFixture) achievements.Registry {
	return achievements.Registry{
		CopyKeys: toFoundationSet(fixture.Registry.CopyKeys), GeneratorIDs: toFoundationSet(fixture.Registry.GeneratorIDs),
		EventKinds: toFoundationSet(fixture.Registry.EventKinds), ResourceIDs: toFoundationSet(fixture.Registry.ResourceIDs),
		RunCounters: toFoundationSet(fixture.Registry.RunCounters), CareerCounters: toFoundationSet(fixture.Registry.CareerCounters),
		ProvenanceSources: fixture.Registry.ProvenanceSources,
	}
}

func foundationTestBundles(t *testing.T) (CatalogBundle, CatalogBundle) {
	t.Helper()
	seed, err := epochseed.Load("../..")
	if err != nil {
		t.Fatal(err)
	}
	legacy := loadReplayTestBundle(t, seed.Hash, seed.Artifacts)
	metersFixture, err := os.ReadFile("../../balance/testdata/meters-catalog-parity-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	var meterEnvelope struct {
		Baseline json.RawMessage `json:"baseline"`
	}
	if json.Unmarshal(metersFixture, &meterEnvelope) != nil {
		t.Fatal("decode meter fixture")
	}
	meterCatalog, err := meters.LoadCatalog(meterEnvelope.Baseline)
	if err != nil {
		t.Fatal(err)
	}
	achievementArtifact := []byte(`{"schema_version":1,"achievements":[{"id":"achievement.first_gate","condition_scope":"run","condition":{"kind":"counter_at_least","counter":"tier","minimum":1},"proof":{"kind":"provenance","event_kinds":["gate_crossed"]},"score_grant":4,"copy_key":"category.any_percent"}]}`)
	achievementCatalog, err := achievements.LoadCatalog(achievementArtifact, FoundationAchievementRegistry(legacy.Economy))
	if err != nil {
		t.Fatal(err)
	}
	artifacts := make(map[string][]byte, len(seed.Artifacts)+2)
	for name, data := range seed.Artifacts {
		artifacts[name] = append([]byte(nil), data...)
	}
	artifacts["meters"] = append([]byte(nil), meterEnvelope.Baseline...)
	artifacts["achievements"] = append([]byte(nil), achievementArtifact...)
	hash, err := save.ConstantsHashArtifacts(artifacts)
	if err != nil {
		t.Fatal(err)
	}
	active := legacy
	active.ConstantsHash, active.Artifacts, active.Meters, active.Achievements = hash, artifacts, meterCatalog, achievementCatalog
	if !active.valid(hash) {
		t.Fatal("foundation bundle is not internally valid")
	}
	return legacy, active
}

func retunedAchievementBundle(t *testing.T, active CatalogBundle, scoreGrant int64) CatalogBundle {
	t.Helper()
	var raw map[string]any
	if err := json.Unmarshal(active.Artifacts["achievements"], &raw); err != nil {
		t.Fatal(err)
	}
	rows := raw["achievements"].([]any)
	rows[0].(map[string]any)["score_grant"] = scoreGrant
	artifact, err := json.Marshal(raw)
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := achievements.LoadCatalog(artifact, FoundationAchievementRegistry(active.Economy))
	if err != nil {
		t.Fatal(err)
	}
	artifacts := make(map[string][]byte, len(active.Artifacts))
	for name, data := range active.Artifacts {
		artifacts[name] = append([]byte(nil), data...)
	}
	artifacts["achievements"] = artifact
	hash, err := save.ConstantsHashArtifacts(artifacts)
	if err != nil {
		t.Fatal(err)
	}
	active.Artifacts, active.ConstantsHash, active.Achievements = artifacts, hash, catalog
	if !active.valid(hash) {
		t.Fatal("retuned foundation bundle is not internally valid")
	}
	return active
}

func toFoundationSet(values []string) map[string]bool {
	result := make(map[string]bool, len(values))
	for _, value := range values {
		result[value] = true
	}
	return result
}

func foundationScopeState(t *testing.T, catalog *economy.Catalog, scope economy.Scope) *save.State {
	t.Helper()
	ledger, err := economy.NewLedger(catalog, scope)
	if err != nil {
		t.Fatal(err)
	}
	return &save.State{Ledger: ledger, WireVersion: save.CurrentVersion}
}

func TestFoundationActivationIsAtomicNewRunOnly(t *testing.T) {
	legacy, active := foundationTestBundles(t)
	founder := foundationScopeState(t, legacy.Economy, economy.ScopeFounder)
	founder.Notoriety = 100
	company := foundationScopeState(t, legacy.Economy, economy.ScopeCompany)
	company.MeterBands = map[string]int{"trust.users.standing": 99}
	newCompany := foundationScopeState(t, active.Economy, economy.ScopeCompany)

	if err := settleAndActivateFoundations(legacy, active, founder, company, newCompany); err != nil {
		t.Fatal(err)
	}
	if save.VersionForState(founder) != 16 || save.VersionForState(newCompany) != 16 || save.VersionForState(company) != 14 {
		t.Fatalf("versions founder=%d old=%d new=%d", save.VersionForState(founder), save.VersionForState(company), save.VersionForState(newCompany))
	}
	if newCompany.MeterValues["trust.users.standing"] != 55 || newCompany.MeterValues["trust.users.grievance"] != 50 ||
		newCompany.MeterValues["trust.users.standing"] == company.MeterBands["trust.users.standing"] {
		t.Fatalf("new meter state=%v legacy=%v", newCompany.MeterValues, company.MeterBands)
	}
	if len(newCompany.AchievementsEarnedRun) != 0 || newCompany.AchievementScoreRun != 0 || len(founder.AchievementsEarnedLifetime) != 0 {
		t.Fatalf("activation retroactively earned achievements: founder=%+v company=%+v", founder, newCompany)
	}
	if err := active.ValidateFoundationState(founder); err != nil {
		t.Fatal(err)
	}
	if err := active.ValidateFoundationState(newCompany); err != nil {
		t.Fatal(err)
	}
	context, err := routeContext(newCompany, active.Routes.ContextVersion())
	if err != nil || context.MeterBands["trust.users.standing"] != 55 {
		t.Fatalf("v16 route context meters=%v err=%v", context.MeterBands, err)
	}
}

func TestFoundationExitSettlesAndDerivesAchievementScore(t *testing.T) {
	legacy, active := foundationTestBundles(t)
	founder := foundationScopeState(t, legacy.Economy, economy.ScopeFounder)
	oldCompany := foundationScopeState(t, legacy.Economy, economy.ScopeCompany)
	company := foundationScopeState(t, active.Economy, economy.ScopeCompany)
	if err := settleAndActivateFoundations(legacy, active, founder, oldCompany, company); err != nil {
		t.Fatal(err)
	}
	definition := active.Achievements.Definitions[0]
	company.AchievementsEarnedRun[definition.ID] = true
	company.AchievementScoreRun = definition.ScoreGrant
	nextCompany := foundationScopeState(t, active.Economy, economy.ScopeCompany)
	if err := settleAndActivateFoundations(active, active, founder, company, nextCompany); err != nil {
		t.Fatal(err)
	}
	if !founder.AchievementsEarnedLifetime[definition.ID] || founder.AchievementScoreLifetime != definition.ScoreGrant ||
		len(nextCompany.AchievementsEarnedRun) != 0 || nextCompany.AchievementScoreRun != 0 {
		t.Fatalf("settlement founder=%+v next=%+v", founder, nextCompany)
	}
	company.AchievementScoreRun++
	if err := active.ValidateFoundationState(company); err == nil {
		t.Fatal("non-derived run score was accepted")
	}
}

func TestFoundationExitRederivesLifetimeScoreUnderNextCatalog(t *testing.T) {
	legacy, active := foundationTestBundles(t)
	founder := foundationScopeState(t, legacy.Economy, economy.ScopeFounder)
	oldCompany := foundationScopeState(t, legacy.Economy, economy.ScopeCompany)
	company := foundationScopeState(t, active.Economy, economy.ScopeCompany)
	if err := settleAndActivateFoundations(legacy, active, founder, oldCompany, company); err != nil {
		t.Fatal(err)
	}
	definition := active.Achievements.Definitions[0]
	company.AchievementsEarnedRun[definition.ID] = true
	company.AchievementScoreRun = definition.ScoreGrant
	next := retunedAchievementBundle(t, active, definition.ScoreGrant+1)
	nextCompany := foundationScopeState(t, next.Economy, economy.ScopeCompany)
	if err := settleAndActivateFoundations(active, next, founder, company, nextCompany); err != nil {
		t.Fatalf("honest Exit across score retune failed: %v", err)
	}
	if founder.AchievementScoreLifetime != definition.ScoreGrant+1 {
		t.Fatalf("lifetime score = %d, want next-catalog grant %d", founder.AchievementScoreLifetime, definition.ScoreGrant+1)
	}
	if err := next.ValidateFoundationState(founder); err != nil {
		t.Fatalf("next-catalog Founder state rejected: %v", err)
	}
}

func TestFoundationArtifactPairAndDowngradeFailClosed(t *testing.T) {
	legacy, active := foundationTestBundles(t)
	broken := active
	broken.Achievements = nil
	if broken.valid(active.ConstantsHash) {
		t.Fatal("single foundation artifact was accepted")
	}
	founder := foundationScopeState(t, legacy.Economy, economy.ScopeFounder)
	oldCompany := foundationScopeState(t, legacy.Economy, economy.ScopeCompany)
	company := foundationScopeState(t, active.Economy, economy.ScopeCompany)
	if err := settleAndActivateFoundations(legacy, active, founder, oldCompany, company); err != nil {
		t.Fatal(err)
	}
	if err := settleAndActivateFoundations(active, legacy, founder, company, foundationScopeState(t, legacy.Economy, economy.ScopeCompany)); err == nil {
		t.Fatal("foundation artifact downgrade was accepted")
	}
}

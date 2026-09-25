package production

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"cloud-clicker/server/economy"
	"cloud-clicker/server/faction"
	"cloud-clicker/server/meters"
	"cloud-clicker/server/pet"
	prestigecore "cloud-clicker/server/prestige"
	"cloud-clicker/server/routes"
	"cloud-clicker/server/save"
)

// TestAxisStackIntegrationReattainsAndBuysPRIntern is AC8: through
// Service.Handle → store → Postgres, a veteran's generator purchase re-attains
// a run-scoped achievement (achievement_reattained.v1, not achievement_earned)
// and raises the axis, after which a PR Intern purchase applies; the events
// constraint rejects an unregistered kind and schema version 2.
func TestAxisStackIntegrationReattainsAndBuysPRIntern(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}
	ctx := context.Background()
	db, err := save.OpenPostgres(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := save.Migrate(ctx, db); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `TRUNCATE accounts,save_streams,catalog_sets,epochs RESTART IDENTITY CASCADE`); err != nil {
		t.Fatal(err)
	}
	bundle := axisContentBundle(t)
	seedProductionEpoch(t, db, bundle.ConstantsHash, bundle.Artifacts)
	resolver := integrationCatalogs{economy: map[string]*economy.Catalog{bundle.ConstantsHash: bundle.Economy},
		routes: map[string]*routes.Catalog{bundle.ConstantsHash: bundle.Routes}, prestige: map[string]*prestigecore.Policy{bundle.ConstantsHash: bundle.Prestige},
		factions: map[string]*faction.Catalog{bundle.ConstantsHash: bundle.Faction}}
	store, err := save.NewStore(db, resolver, nil)
	if err != nil {
		t.Fatal(err)
	}
	cursor := save.CanonicalServerTime(time.Date(2026, 9, 25, 12, 0, 0, 0, time.UTC))
	const accountID = "01986666-7a00-4000-8000-000000000001"
	const founderID = "01986666-7a00-7000-8000-000000000002"
	if _, err := db.ExecContext(ctx, `INSERT INTO accounts(account_id,recovery_hash) VALUES($1,'test')`, accountID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO account_founders(account_id,founder_id) VALUES($1,$2)`, accountID, founderID); err != nil {
		t.Fatal(err)
	}

	company := replayFixtureState(t, bundle.Economy, cursor)
	meterState, err := meters.NewRunState(bundle.Meters, 0)
	if err != nil {
		t.Fatal(err)
	}
	company.MeterBands = nil
	company.MeterValues, company.MeterDecayRemainders, company.MeterInputRemainders = meterState.Values, meterState.DecayRemainders, meterState.InputRemainders
	company.AchievementsEarnedRun = map[string]bool{}
	// Pre-seeded run attainment (derivation-valid): first_gate 2 + purchased_25 4 = 6.
	company.AchievementsAttainedRun = map[string]bool{"achievement.first_gate": true, "achievement.generators_purchased_25": true}
	company.AttainmentScoreRun = 6
	company.GatesCrossed["gate.t0_to_t1"], company.Tier = true, 1
	company.RunStartedAt = cursor
	setCash(t, company, "1e8")
	if _, err := initializeActivePlayState(company, bundle.Opportunities, founderID); err != nil {
		t.Fatal(err)
	}
	company.WireVersion = 19
	if err := bundle.ValidateFoundationState(company); err != nil {
		t.Fatal(err)
	}
	companyRevision, err := store.CreateStream(ctx, save.StreamKey{OwnerKind: save.OwnerFounder, OwnerID: founderID, Scope: economy.ScopeCompany},
		bundle.ConstantsHash, company, save.WriteContext{Cause: "axis.integration"})
	if err != nil {
		t.Fatal(err)
	}
	founderFloor, _ := bundle.versionFloors()
	founder := replayFounderFixtureState(t, bundle, cursor)
	founder.WireVersion = founderFloor
	founder.MinigameRatings = map[string]save.MinigameRatingState{"pitch": {Elo: 1000, SeasonMember: "s1"}}
	founder.MinigameOfflineQuality = map[string]save.MinigameOfflineQualityState{"pitch": {GradePPM: 200_000}}
	founder.Pets = map[string]pet.CareState{}
	founder.FiscalPeriodOpenedWallMS, founder.FiscalGeneratorLevels, founder.FiscalUnlocks = cursor.UnixMilli(), map[string]int64{}, map[string]bool{}
	for _, row := range bundle.Fiscal.GeneratorLevelRows() {
		founder.FiscalGeneratorLevels[row.GeneratorID] = 0
	}
	founder.Soul, founder.SoulExhaustedSourceIDs = bundle.Soul.Policy.Initial, []string{}
	// The veteran already owns generators_purchased_1 for life, so the run's
	// first purchase can only re-attain it (CV2), never earn it.
	founder.AchievementsEarnedLifetime = map[string]bool{"achievement.generators_purchased_1": true}
	founder.AchievementScoreLifetime = 2
	if err := bundle.ValidateFoundationState(founder); err != nil {
		t.Fatalf("Founder fixture: %v", err)
	}
	if _, err := store.CreateStream(ctx, save.StreamKey{OwnerKind: save.OwnerFounder, OwnerID: founderID, Scope: economy.ScopeFounder},
		bundle.ConstantsHash, founder, save.WriteContext{Cause: "axis.integration"}); err != nil {
		t.Fatal(err)
	}
	frozen, err := FrozenFounderContributions(bundle, founder)
	if err != nil {
		t.Fatal(err)
	}
	genesis, err := save.EncodeState(company)
	if err != nil {
		t.Fatal(err)
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := save.PinRunWithGenesisTx(ctx, tx, companyRevision.StreamID, founderID, 1, bundle.ConstantsHash, save.VersionForState(company), genesis); err != nil {
		t.Fatal(err)
	}
	if err := save.InsertRunFrozenContributionsTx(ctx, tx, companyRevision.StreamID, 1, frozen); err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
	service, err := NewService(store, resolver, nil, nil, nil, WithProgressionRuntime(resolver), WithCurrentConstantsHash(bundle.ConstantsHash),
		WithReplayCatalogs(ReplayCatalogSet{bundle.ConstantsHash: bundle}), WithGuildSettlements(emptyGuildSettlements{}), WithRouteCatalogs(resolver))
	if err != nil {
		t.Fatal(err)
	}

	buyGenerator := []byte(`{"intent_id":"01986666-7a00-7000-8000-000000000003","kind":"buy_generator","expected_revision":1,"generator_id":"generator.beige_tower","count":{"mode":"exact","value":1}}`)
	result, err := service.Handle(ctx, companyRevision.StreamID, ModeOnline, cursor.Add(time.Second), buyGenerator)
	if err != nil || !strings.Contains(string(result.Receipt), `"outcome":"applied"`) || !strings.Contains(string(result.Receipt), `"attainment_score_run":8`) {
		t.Fatalf("generator purchase receipt=%s err=%v", result.Receipt, err)
	}
	earlyIntern := []byte(`{"intent_id":"01986666-7a00-7000-8000-000000000004","kind":"buy_upgrade","expected_revision":2,"upgrade_id":"upgrade.pr_intern_2"}`)
	early, err := service.Handle(ctx, companyRevision.StreamID, ModeOnline, cursor.Add(2*time.Second), earlyIntern)
	if err != nil || !strings.Contains(string(early.Receipt), `"category":"not_eligible"`) || !strings.Contains(string(early.Receipt), `"detail":"requires"`) {
		t.Fatalf("pr_intern_2 at x=8 < 10 receipt=%s err=%v", early.Receipt, err)
	}
	buyIntern := []byte(`{"intent_id":"01986666-7a00-7000-8000-000000000005","kind":"buy_upgrade","expected_revision":2,"upgrade_id":"upgrade.pr_intern_1"}`)
	intern, err := service.Handle(ctx, companyRevision.StreamID, ModeOnline, cursor.Add(3*time.Second), buyIntern)
	if err != nil || !strings.Contains(string(intern.Receipt), `"outcome":"applied"`) {
		t.Fatalf("pr_intern_1 receipt=%s err=%v", intern.Receipt, err)
	}
	// first_gate is newly earned for this Founder (tier 1) and was already
	// attained, so it emits achievement_earned only; generators_purchased_1 is
	// lifetime-owned, so it is re-attained and never earned (CV5: no ID gets
	// both kinds in one transition).
	var earnedPurchase int
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM events WHERE stream_id=$1 AND kind='achievement_earned.v1' AND payload->>'achievement_id'='achievement.generators_purchased_1'`, companyRevision.StreamID).Scan(&earnedPurchase); err != nil || earnedPurchase != 0 {
		t.Fatalf("lifetime-owned purchase earned again rows=%d err=%v", earnedPurchase, err)
	}
	for kind, want := range map[string]int{"achievement_reattained.v1": 1, "achievement_earned.v1": 1, "upgrade_purchased": 1} {
		var count int
		if err := db.QueryRowContext(ctx, `SELECT count(*) FROM events WHERE stream_id=$1 AND kind=$2`, companyRevision.StreamID, kind).Scan(&count); err != nil || count != want {
			t.Fatalf("%s rows=%d want %d err=%v", kind, count, want, err)
		}
	}
	var payload string
	if err := db.QueryRowContext(ctx, `SELECT payload::text FROM events WHERE stream_id=$1 AND kind='achievement_reattained.v1'`, companyRevision.StreamID).Scan(&payload); err != nil ||
		!strings.Contains(payload, `"achievement_id": "achievement.generators_purchased_1"`) || !strings.Contains(payload, `"score_grant": 2`) {
		t.Fatalf("reattained payload=%s err=%v", payload, err)
	}
	loaded, err := store.LoadLatest(ctx, companyRevision.StreamID)
	if err != nil || save.VersionForState(loaded.State) != 19 || loaded.State.AttainmentScoreRun != 8 || !loaded.State.UpgradesOwned["upgrade.pr_intern_1"] {
		t.Fatalf("Company after purchase v%d score=%d owned=%v err=%v", save.VersionForState(loaded.State), loaded.State.AttainmentScoreRun, loaded.State.UpgradesOwned, err)
	}
	// AC8 failing cases: the constraint rejects an unregistered kind and v2.
	if _, err := db.ExecContext(ctx, `UPDATE events SET kind='achievement_unattained.v1' WHERE stream_id=$1 AND kind='achievement_reattained.v1'`, companyRevision.StreamID); err == nil {
		t.Fatal("events constraint accepted an unregistered kind")
	}
	if _, err := db.ExecContext(ctx, `UPDATE events SET schema_version=2 WHERE stream_id=$1 AND kind='achievement_reattained.v1'`, companyRevision.StreamID); err == nil {
		t.Fatal("events constraint accepted achievement_reattained at schema version 2")
	}
}

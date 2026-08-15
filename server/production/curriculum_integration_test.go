package production

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"cloud-clicker/server/commons"
	"cloud-clicker/server/commonsbinding"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/faction"
	"cloud-clicker/server/meters"
	"cloud-clicker/server/minigame"
	"cloud-clicker/server/pet"
	prestigecore "cloud-clicker/server/prestige"
	"cloud-clicker/server/routes"
	"cloud-clicker/server/save"
)

func TestCurriculumAutomaticFailurePersistsAndReplaysIntegration(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}
	ctx := context.Background()
	db, err := save.OpenPostgres(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := save.Migrate(ctx, db); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `TRUNCATE epochs,catalog_sets,save_streams RESTART IDENTITY CASCADE`); err != nil {
		t.Fatal(err)
	}
	current := activeContentBundle(t)
	if current.Curriculum == nil {
		t.Fatal("minted curriculum artifact is unavailable")
	}
	next := current
	seedProductionEpoch(t, db, current.ConstantsHash, current.Artifacts)
	resolver := integrationCatalogs{
		economy:  map[string]*economy.Catalog{current.ConstantsHash: current.Economy, next.ConstantsHash: next.Economy},
		routes:   map[string]*routes.Catalog{current.ConstantsHash: current.Routes, next.ConstantsHash: next.Routes},
		prestige: map[string]*prestigecore.Policy{current.ConstantsHash: current.Prestige, next.ConstantsHash: next.Prestige},
		factions: map[string]*faction.Catalog{current.ConstantsHash: current.Faction, next.ConstantsHash: next.Faction},
	}
	store, err := save.NewStore(db, resolver, nil)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 8, 15, 18, 0, 0, 0, time.UTC)
	const founderID = "01986666-ad00-7000-8000-000000000001"
	founder := replayFounderFixtureState(t, current, now)
	founder.WireVersion = 21
	if err := activateMinigameState(founder, current.Minigames); err != nil {
		t.Fatal(err)
	}
	founder.Pets = map[string]pet.CareState{}
	founder.FiscalPeriodOpenedWallMS, founder.FiscalGeneratorLevels, founder.FiscalUnlocks = now.UnixMilli(), map[string]int64{}, map[string]bool{}
	for _, row := range current.Fiscal.GeneratorLevelRows() {
		founder.FiscalGeneratorLevels[row.GeneratorID] = 0
	}
	founder.Soul, founder.SoulExhaustedSourceIDs = 80, []string{}
	company := replayFixtureState(t, current.Economy, now.Add(-15*time.Minute))
	company.WireVersion, company.Tier = 18, 1
	company.GatesCrossed["gate.t0_to_t1"] = true
	company.MeterBands = nil
	meterState, meterErr := meters.NewRunState(current.Meters, 0)
	if meterErr != nil {
		t.Fatal(meterErr)
	}
	company.MeterValues, company.MeterDecayRemainders, company.MeterInputRemainders = meterState.Values, meterState.DecayRemainders, meterState.InputRemainders
	company.AchievementsEarnedRun = map[string]bool{}
	if _, err := initializeActivePlayState(company, current.Opportunities, founderID); err != nil {
		t.Fatal(err)
	}
	advanceActivePlayFixtureAttendance(t, company, current.Opportunities, current.Prestige, founderID, now)
	company.ManualTokenRefilledAt = now
	founderRevision, err := store.CreateStream(ctx, save.StreamKey{OwnerKind: save.OwnerFounder, OwnerID: founderID, Scope: economy.ScopeFounder}, current.ConstantsHash, founder, save.WriteContext{Cause: "curriculum.integration"})
	if err != nil {
		t.Fatal(err)
	}
	companyRevision, err := store.CreateStream(ctx, save.StreamKey{OwnerKind: save.OwnerFounder, OwnerID: founderID, Scope: economy.ScopeCompany}, current.ConstantsHash, company, save.WriteContext{Cause: "curriculum.integration"})
	if err != nil {
		t.Fatal(err)
	}
	pinTx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	genesis, _ := save.EncodeState(company)
	_, err = save.PinRunWithGenesisTx(ctx, pinTx, companyRevision.StreamID, founderID, 1, current.ConstantsHash, companyRevision.Version, genesis)
	if err == nil {
		var frozen []save.FrozenContribution
		frozen, err = FrozenFiscalContributions(current.Fiscal, founder)
		if err == nil {
			err = save.InsertRunFrozenContributionsTx(ctx, pinTx, companyRevision.StreamID, 1, frozen)
		}
	}
	if err == nil {
		err = pinTx.Commit()
	} else {
		_ = pinTx.Rollback()
	}
	if err != nil {
		t.Fatal(err)
	}
	minigameRepository, err := minigame.NewRepository(db)
	if err != nil {
		t.Fatal(err)
	}
	commonsCatalogs := commons.CatalogSet{current.ConstantsHash: current.Commons.(commonsbinding.ReplayPolicy).Catalog, next.ConstantsHash: next.Commons.(commonsbinding.ReplayPolicy).Catalog}
	service, err := NewService(store, resolver, FrozenContributionProvider{DB: db}, nil, nil,
		WithProgressionRuntime(resolver), WithCurrentConstantsHash(next.ConstantsHash),
		WithReplayCatalogs(ReplayCatalogSet{current.ConstantsHash: current, next.ConstantsHash: next}), WithGuildSettlements(emptyGuildSettlements{}),
		WithMinigameActivity(minigameRepository), WithCompactPolicies(commonsCatalogs), WithCommonsWeightResolver(integrationWeight(1_000_000)))
	if err != nil {
		t.Fatal(err)
	}
	request := []byte(`{"intent_id":"01986666-ad00-7000-8000-000000000002","kind":"perform_manual_batch","expected_revision":1,"action_id":"manual.click","count":1,"window_ms":1}`)
	result, err := service.Handle(ctx, companyRevision.StreamID, ModeOnline, now, request)
	if err != nil || !strings.Contains(string(result.Receipt), `"outcome":"applied"`) {
		t.Fatalf("receipt=%s err=%v", result.Receipt, err)
	}
	retry, err := service.Handle(ctx, companyRevision.StreamID, ModeOnline, now, request)
	if err != nil || !retry.Replay || string(retry.Receipt) != string(result.Receipt) {
		t.Fatalf("retry=%+v err=%v", retry, err)
	}
	loadedCompany, err := store.LoadLatest(ctx, companyRevision.StreamID)
	if err != nil || loadedCompany.State.RunSeq != 2 || loadedCompany.Revision.ConstantsHash != next.ConstantsHash || loadedCompany.State.GeneratorProvisioned["generator.beige_tower"] != 10 {
		t.Fatalf("company=%+v revision=%+v err=%v", loadedCompany.State, loadedCompany.Revision, err)
	}
	loadedFounder, err := store.LoadLatest(ctx, founderRevision.StreamID)
	if err != nil || loadedFounder.State.RouteKnowledgeBalance != 75 || len(loadedFounder.State.ExitHistory) != 1 {
		t.Fatalf("founder=%+v err=%v", loadedFounder.State, err)
	}
	var schema int
	var payload, replayInputs []byte
	if err := db.QueryRowContext(ctx, `SELECT schema_version,payload FROM events WHERE stream_id=$1 AND kind='run_ended'`, companyRevision.StreamID).Scan(&schema, &payload); err != nil {
		t.Fatal(err)
	}
	var ended struct {
		Branch string `json:"branch"`
	}
	if err := json.Unmarshal(payload, &ended); err != nil || schema != 3 || ended.Branch != "burnout" {
		t.Fatalf("run_ended schema=%d payload=%s err=%v", schema, payload, err)
	}
	if err := db.QueryRowContext(ctx, `SELECT replay_inputs FROM run_log WHERE company_stream_id=$1 AND run_seq=1`, companyRevision.StreamID).Scan(&replayInputs); err != nil {
		t.Fatal(err)
	}
	wire, err := parseReplayInputs(replayInputs)
	var terminal replayExitResolved
	if err == nil {
		err = decodeReplayStrict(wire.Resolved, &terminal)
	}
	if err != nil || terminal.SelectedBranch == nil || *terminal.SelectedBranch != "burnout" || terminal.IntentKind != IntentPerformManualBatch {
		t.Fatalf("terminal=%+v err=%v", terminal, err)
	}
}

package production

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"cloud-clicker/server/economy"
	"cloud-clicker/server/faction"
	"cloud-clicker/server/meters"
	"cloud-clicker/server/minigame"
	"cloud-clicker/server/pet"
	"cloud-clicker/server/pitch"
	prestigecore "cloud-clicker/server/prestige"
	"cloud-clicker/server/routes"
	"cloud-clicker/server/save"
)

func TestPitchComposedIntegrationUnlockReplayAndPayout(t *testing.T) {
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
	bundle := pitchFeatureBundle(t)
	seedProductionEpoch(t, db, bundle.ConstantsHash, bundle.Artifacts)
	resolver := integrationCatalogs{economy: map[string]*economy.Catalog{bundle.ConstantsHash: bundle.Economy},
		routes: map[string]*routes.Catalog{bundle.ConstantsHash: bundle.Routes}, prestige: map[string]*prestigecore.Policy{bundle.ConstantsHash: bundle.Prestige},
		factions: map[string]*faction.Catalog{bundle.ConstantsHash: bundle.Faction}}
	store, err := save.NewStore(db, resolver, nil)
	if err != nil {
		t.Fatal(err)
	}
	now := save.CanonicalServerTime(time.Now().UTC())
	const accountID = "01986666-c100-4000-8000-000000000001"
	const founderID = "01986666-c100-4000-8000-000000000002"
	if _, err := db.ExecContext(ctx, `INSERT INTO accounts(account_id,recovery_hash) VALUES($1,'test')`, accountID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO account_founders(account_id,founder_id) VALUES($1,$2)`, accountID, founderID); err != nil {
		t.Fatal(err)
	}
	company := replayFixtureState(t, bundle.Economy, now)
	company.WireVersion, company.MeterBands = 16, nil
	meterState, err := meters.NewRunState(bundle.Meters, 0)
	if err != nil {
		t.Fatal(err)
	}
	company.MeterValues, company.MeterDecayRemainders, company.MeterInputRemainders = meterState.Values, meterState.DecayRemainders, meterState.InputRemainders
	company.AchievementsEarnedRun = map[string]bool{}
	companyRevision, err := store.CreateStream(ctx, save.StreamKey{OwnerKind: save.OwnerFounder, OwnerID: founderID, Scope: economy.ScopeCompany},
		bundle.ConstantsHash, company, save.WriteContext{Cause: "pitch.integration"})
	if err != nil {
		t.Fatal(err)
	}
	founder := replayFounderFixtureState(t, bundle, now)
	founder.WireVersion = 20
	founder.MinigameRatings = map[string]save.MinigameRatingState{"pitch": {Elo: 1000, SeasonMember: "s1"}}
	founder.MinigameOfflineQuality = map[string]save.MinigameOfflineQualityState{"pitch": {GradePPM: 200_000}}
	founder.Pets = map[string]pet.CareState{}
	founder.FiscalCredit, founder.FiscalPeriodOpenedWallMS, founder.FiscalPeriodSequence = 3, now.UnixMilli(), 0
	founder.FiscalGeneratorLevels = make(map[string]int64, len(bundle.Fiscal.GeneratorLevelRows()))
	for _, row := range bundle.Fiscal.GeneratorLevelRows() {
		founder.FiscalGeneratorLevels[row.GeneratorID] = 0
	}
	founder.FiscalUnlocks, founder.Soul, founder.SoulExhaustedSourceIDs = map[string]bool{}, 50, []string{}
	founderRevision, err := store.CreateStream(ctx, save.StreamKey{OwnerKind: save.OwnerFounder, OwnerID: founderID, Scope: economy.ScopeFounder},
		bundle.ConstantsHash, founder, save.WriteContext{Cause: "pitch.integration"})
	if err != nil {
		t.Fatal(err)
	}
	genesis, err := save.EncodeState(company)
	if err != nil {
		t.Fatal(err)
	}
	frozen, err := FrozenFiscalContributions(bundle.Fiscal, founder)
	if err != nil {
		t.Fatal(err)
	}
	pinTx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = save.PinRunWithGenesisTx(ctx, pinTx, companyRevision.StreamID, founderID, 1,
		bundle.ConstantsHash, save.VersionForState(company), genesis); err == nil {
		err = save.InsertRunFrozenContributionsTx(ctx, pinTx, companyRevision.StreamID, 1, frozen)
	}
	if err == nil {
		err = pinTx.Commit()
	} else {
		_ = pinTx.Rollback()
	}
	if err != nil {
		t.Fatal(err)
	}
	repository, err := minigame.NewRepository(db)
	if err != nil {
		t.Fatal(err)
	}
	registry, err := minigame.NewTenantRegistry(pitch.NewTenant())
	if err != nil {
		t.Fatal(err)
	}
	set := ReplayCatalogSet{bundle.ConstantsHash: bundle}
	platform, err := minigame.NewService(repository, registry, set)
	if err != nil {
		t.Fatal(err)
	}
	production, err := NewService(store, resolver, nil, nil, nil, WithProgressionRuntime(resolver), WithCurrentConstantsHash(bundle.ConstantsHash),
		WithReplayCatalogs(set), WithGuildSettlements(emptyGuildSettlements{}))
	if err != nil {
		t.Fatal(err)
	}
	request := minigame.StartRequest{SessionID: "01986666-c101-7000-8000-000000000001", MinigameID: "pitch", FounderID: founderID,
		CompanyStreamID: companyRevision.StreamID, RunSeq: 1, EngineRef: pitch.EngineRef, EngineVersion: pitch.EngineVersion,
		ConstantsHash: bundle.ConstantsHash, ScalingInputs: map[string]int64{pitch.ScalingDestination: 1}, Seed: "2", Mode: minigame.ModeSolo}
	if _, err := production.StartMinigameSession(ctx, platform, request, now); err == nil || !strings.Contains(err.Error(), "fiscal_unlock_required") {
		t.Fatalf("Pitch start before purchase err=%v", err)
	}
	spend := []byte(`{"intent_id":"01986666-c101-7000-8000-000000000002","kind":"spend_fiscal_credit","expected_revision":1,"target":{"kind":"unlock","unlock_id":"minigame.pitch"}}`)
	spendResult, err := production.Handle(ctx, companyRevision.StreamID, ModeOnline, now, spend)
	if err != nil || !bytes.Contains(spendResult.Receipt, []byte(`"outcome":"applied"`)) {
		t.Fatalf("Fiscal purchase receipt=%s err=%v", spendResult.Receipt, err)
	}
	session, err := production.StartMinigameSession(ctx, platform, request, now)
	if err != nil || session.Revision != 1 {
		t.Fatalf("Pitch start after purchase session=%+v err=%v", session, err)
	}
	corpusData, err := os.ReadFile("../../testdata/pitch/content-gate-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	var corpus struct {
		Scenarios []struct {
			Seed     uint64            `json:"seed"`
			Commands []json.RawMessage `json:"commands"`
		} `json:"scenarios"`
	}
	if json.Unmarshal(corpusData, &corpus) != nil || len(corpus.Scenarios) == 0 || corpus.Scenarios[0].Seed != 2 {
		t.Fatal("invalid Pitch integration corpus")
	}
	var resolution *minigame.CertifiedResolution
	for index, command := range corpus.Scenarios[0].Commands {
		var commandValue any
		if json.Unmarshal(command, &commandValue) != nil {
			t.Fatalf("command %d is not JSON", index)
		}
		command, _ = json.Marshal(commandValue)
		decision, playErr := platform.Play(ctx, minigame.PlayRequest{FounderID: founderID, SessionID: request.SessionID,
			ExpectedRevision: int64(index + 1), Command: command})
		if playErr != nil {
			t.Fatalf("command %d err=%v", index, playErr)
		}
		if decision.Resolution != nil {
			resolution = decision.Resolution
		}
	}
	if resolution == nil {
		t.Fatal("Pitch corpus did not produce a certified result")
	}
	result, err := production.ResolveMinigameSession(ctx, platform, resolution, now.Add(time.Minute), nil)
	if err != nil || result.Replay || !bytes.Contains(result.Receipt, []byte(`"credited_delta":"1e0"`)) {
		t.Fatalf("Pitch resolve receipt=%s replay=%v err=%v", result.Receipt, result.Replay, err)
	}
	retry, err := production.ResolveMinigameSession(ctx, platform, resolution, now.Add(time.Minute), nil)
	if err != nil || !retry.Replay || !bytes.Equal(retry.Receipt, result.Receipt) {
		t.Fatalf("Pitch retry receipt=%s replay=%v err=%v", retry.Receipt, retry.Replay, err)
	}
	loadedCompany, _ := store.LoadLatest(ctx, companyRevision.StreamID)
	loadedFounder, _ := store.LoadLatest(ctx, founderRevision.StreamID)
	cash, _ := loadedCompany.State.Ledger.Balance("company.cash")
	if cash.String() != "1e0" || loadedFounder.State.FiscalUnlocks["minigame.pitch"] != true ||
		loadedFounder.State.MinigameRatings["pitch"].Elo != 1000 || loadedFounder.State.MinigameOfflineQuality["pitch"].GradePPM != 500_000 {
		t.Fatalf("cash=%s founder=%+v", cash, loadedFounder.State)
	}
	history, err := store.LoadFounderHistory(ctx, founderRevision.StreamID)
	verdict := VerifyFounderHistory(history, set)
	if err != nil || verdict != ReplayVerified {
		t.Fatalf("Pitch Founder history verdict=%s err=%v", verdict, err)
	}
	if _, err := production.StartMinigameSession(ctx, platform, minigame.StartRequest{SessionID: "01986666-c101-7000-8000-000000000003",
		MinigameID: "pitch", FounderID: founderID, CompanyStreamID: companyRevision.StreamID, RunSeq: 1, EngineRef: "fixture.counter",
		EngineVersion: pitch.EngineVersion, ConstantsHash: bundle.ConstantsHash, ScalingInputs: request.ScalingInputs, Seed: "2", Mode: minigame.ModeSolo}, now); !errors.Is(err, ErrInvalidIntent) {
		t.Fatalf("definition/tenant identity mismatch err=%v", err)
	}
}

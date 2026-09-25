package production

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"cloud-clicker/server/arcade"
	"cloud-clicker/server/copykeys"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/faction"
	"cloud-clicker/server/meters"
	"cloud-clicker/server/minigame"
	"cloud-clicker/server/minigameapi"
	"cloud-clicker/server/pet"
	"cloud-clicker/server/pitch"
	prestigecore "cloud-clicker/server/prestige"
	"cloud-clicker/server/routes"
	"cloud-clicker/server/save"
	"cloud-clicker/server/typer"
)

// arcadeFeatureBundle pins the complete arcade chain (AR-P1) over the Typer
// fixture bundle: the four-row minigames artifact, the four-tenant
// minigame_api artifact, and the arcade content artifact.
func arcadeFeatureBundle(t *testing.T) CatalogBundle {
	t.Helper()
	bundle := typerFeatureBundle(t)
	bundle.Artifacts = cloneArtifactMap(bundle.Artifacts)
	read := func(path string) []byte {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		return data
	}
	bundle.Artifacts["minigames"] = read("../../testdata/minigame/pitch-typer-arcade-v3.json")
	bundle.Artifacts["minigame_api"] = read("../../balance/testdata/minigame-api-arcade-candidate-v1.json")
	bundle.Artifacts["arcade"] = read("../../balance/testdata/arcade-v1.json")
	var err error
	if bundle.Minigames, err = minigame.LoadCatalog(bundle.Artifacts["minigames"]); err != nil {
		t.Fatal(err)
	}
	if bundle.MinigameAPI, err = minigameapi.LoadCatalog(bundle.Artifacts["minigame_api"]); err != nil {
		t.Fatal(err)
	}
	keys := make(map[string]struct{})
	for _, key := range copykeys.All() {
		keys[key] = struct{}{}
	}
	if bundle.Arcade, err = arcade.LoadCatalog(bundle.Artifacts["arcade"], arcade.Declarations{CopyKeys: keys}); err != nil {
		t.Fatal(err)
	}
	if bundle.ConstantsHash, err = save.ConstantsHashArtifacts(bundle.Artifacts); err != nil {
		t.Fatal(err)
	}
	if !bundle.valid(bundle.ConstantsHash) {
		t.Fatal("arcade fixture bundle is not internally valid")
	}
	return bundle
}

func seedArcadeFounder(t *testing.T, ctx context.Context, db *sql.DB, store *save.Store, bundle CatalogBundle, now time.Time, suffix string, soul int64) typerFounder {
	t.Helper()
	accountID := "01986666-a7" + suffix + "-4000-8000-000000000001"
	founderID := "01986666-a7" + suffix + "-4000-8000-000000000002"
	if _, err := db.ExecContext(ctx, `INSERT INTO accounts(account_id,recovery_hash) VALUES($1,'test')`, accountID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO account_founders(account_id,founder_id) VALUES($1,$2)`, accountID, founderID); err != nil {
		t.Fatal(err)
	}
	company := replayFixtureState(t, bundle.Economy, now)
	company.WireVersion, company.MeterBands, company.Tier = 16, nil, 0
	meterState, err := meters.NewRunState(bundle.Meters, 0)
	if err != nil {
		t.Fatal(err)
	}
	company.MeterValues, company.MeterDecayRemainders, company.MeterInputRemainders = meterState.Values, meterState.DecayRemainders, meterState.InputRemainders
	company.AchievementsEarnedRun = map[string]bool{}
	companyRevision, err := store.CreateStream(ctx, save.StreamKey{OwnerKind: save.OwnerFounder, OwnerID: founderID, Scope: economy.ScopeCompany},
		bundle.ConstantsHash, company, save.WriteContext{Cause: "arcade.integration"})
	if err != nil {
		t.Fatal(err)
	}
	founder := replayFounderFixtureState(t, bundle, now)
	founder.WireVersion = 20
	founder.MinigameRatings = map[string]save.MinigameRatingState{}
	founder.MinigameOfflineQuality = map[string]save.MinigameOfflineQualityState{}
	for _, id := range []string{"arcade.mine_grid", "arcade.snake", "pitch", "typer"} {
		founder.MinigameRatings[id] = save.MinigameRatingState{Elo: 1000, SeasonMember: "s1"}
		founder.MinigameOfflineQuality[id] = save.MinigameOfflineQualityState{GradePPM: 200_000}
	}
	founder.Pets = map[string]pet.CareState{}
	founder.FiscalCredit, founder.FiscalPeriodOpenedWallMS, founder.FiscalPeriodSequence = 0, now.UnixMilli(), 0
	founder.FiscalGeneratorLevels = make(map[string]int64, len(bundle.Fiscal.GeneratorLevelRows()))
	for _, row := range bundle.Fiscal.GeneratorLevelRows() {
		founder.FiscalGeneratorLevels[row.GeneratorID] = 0
	}
	founder.FiscalUnlocks, founder.Soul, founder.SoulExhaustedSourceIDs = map[string]bool{}, soul, []string{}
	founder.ExitHistory = nil
	founderRevision, err := store.CreateStream(ctx, save.StreamKey{OwnerKind: save.OwnerFounder, OwnerID: founderID, Scope: economy.ScopeFounder},
		bundle.ConstantsHash, founder, save.WriteContext{Cause: "arcade.integration"})
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
	return typerFounder{founderID: founderID, companyStreamID: companyRevision.StreamID, founderStreamID: founderRevision.StreamID}
}

// Demo Disc Arcade AC (A5): the composed platform through real Postgres —
// the always unlock at Tier 0 with no Exit, the human_hobby lock at
// near-zero Soul, play to terminal, a zero-credit applied resolution with
// cash, rating, and quality unchanged, idempotent retry, and quit releasing
// the Exit block.
func TestArcadeComposedIntegrationUnlockLockPlayAndZeroCreditResolution(t *testing.T) {
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
	bundle := arcadeFeatureBundle(t)
	seedProductionEpoch(t, db, bundle.ConstantsHash, bundle.Artifacts)
	resolver := integrationCatalogs{economy: map[string]*economy.Catalog{bundle.ConstantsHash: bundle.Economy},
		routes: map[string]*routes.Catalog{bundle.ConstantsHash: bundle.Routes}, prestige: map[string]*prestigecore.Policy{bundle.ConstantsHash: bundle.Prestige},
		factions: map[string]*faction.Catalog{bundle.ConstantsHash: bundle.Faction}}
	store, err := save.NewStore(db, resolver, nil)
	if err != nil {
		t.Fatal(err)
	}
	now := save.CanonicalServerTime(time.Now().UTC())
	repository, err := minigame.NewRepository(db)
	if err != nil {
		t.Fatal(err)
	}
	registry, err := minigame.NewTenantRegistry(pitch.NewTenant(), typer.NewTenant(), arcade.NewMineGridTenant(), arcade.NewSnakeTenant())
	if err != nil {
		t.Fatal(err)
	}
	set := ReplayCatalogSet{bundle.ConstantsHash: bundle}
	platform, err := minigame.NewService(repository, registry, set)
	if err != nil {
		t.Fatal(err)
	}
	service, err := NewService(store, resolver, nil, nil, nil, WithProgressionRuntime(resolver), WithCurrentConstantsHash(bundle.ConstantsHash),
		WithReplayCatalogs(set), WithGuildSettlements(emptyGuildSettlements{}))
	if err != nil {
		t.Fatal(err)
	}
	request := func(founder typerFounder, sessionID, toy, engine, destination string) minigame.StartRequest {
		return minigame.StartRequest{SessionID: sessionID, MinigameID: toy, FounderID: founder.founderID,
			CompanyStreamID: founder.companyStreamID, RunSeq: 1, EngineRef: engine, EngineVersion: arcade.EngineVersion,
			ConstantsHash: bundle.ConstantsHash, ScalingInputs: map[string]int64{destination: 1}, Seed: "1", Mode: minigame.ModeSolo}
	}
	mineRequest := func(founder typerFounder, sessionID string) minigame.StartRequest {
		return request(founder, sessionID, "arcade.mine_grid", arcade.MineGridEngineRef, arcade.MineGridScalingDestination)
	}

	// AR1.5: near-zero Soul locks the hobby toy with the shipped rejection.
	hollow := seedArcadeFounder(t, ctx, db, store, bundle, now, "00", 5)
	if _, err := service.StartMinigameSession(ctx, platform, mineRequest(hollow, "01986666-a700-7000-8000-000000000003"), now); !errors.Is(err, ErrInvalidIntent) ||
		!strings.Contains(err.Error(), "human_content_locked") {
		t.Fatalf("near-zero Soul must lock the arcade: %v", err)
	}

	play := func(founder typerFounder, start minigame.StartRequest, commands []string, outcome string) {
		t.Helper()
		session, err := service.StartMinigameSession(ctx, platform, start, now)
		if err != nil || session.Revision != 1 {
			t.Fatalf("%s start session=%+v err=%v", start.MinigameID, session, err)
		}
		if active, err := repository.ActiveMinigame(ctx, founder.founderID); err != nil || !active {
			t.Fatalf("an open arcade session must hold the Exit block: active=%v err=%v", active, err)
		}
		before, _ := store.LoadLatest(ctx, founder.companyStreamID)
		beforeCash, _ := before.State.Ledger.Balance("company.cash")
		var resolution *minigame.CertifiedResolution
		for index, command := range commands {
			canonical, ok := canonicalCommand(command)
			if !ok {
				t.Fatalf("command %d is not canonical JSON", index)
			}
			decision, playErr := platform.Play(ctx, minigame.PlayRequest{FounderID: founder.founderID, SessionID: start.SessionID,
				ExpectedRevision: int64(index + 1), Command: canonical})
			if playErr != nil {
				t.Fatalf("%s command %d (%s) err=%v", start.MinigameID, index, command, playErr)
			}
			if decision.Resolution != nil {
				resolution = decision.Resolution
			}
		}
		if resolution == nil || resolution.Result().Outcome != outcome || resolution.Result().RatingDelta != nil {
			t.Fatalf("%s did not reach %s", start.MinigameID, outcome)
		}
		result, err := service.ResolveMinigameSession(ctx, platform, resolution, now.Add(time.Minute), nil)
		if err != nil || result.Replay {
			t.Fatalf("%s resolve receipt=%s err=%v", start.MinigameID, result.Receipt, err)
		}
		retry, err := service.ResolveMinigameSession(ctx, platform, resolution, now.Add(time.Minute), nil)
		if err != nil || !retry.Replay || !bytes.Equal(retry.Receipt, result.Receipt) {
			t.Fatalf("%s retry receipt=%s err=%v", start.MinigameID, retry.Receipt, err)
		}
		// AR1.6: an applied resolution that credits and forfeits exactly zero.
		var receipt map[string]json.RawMessage
		if json.Unmarshal(result.Receipt, &receipt) != nil || string(receipt["credited_delta"]) != `"0"` ||
			string(receipt["configured_cap_forfeit_units"]) != "0" || string(receipt["outcome"]) != `"applied"` {
			t.Fatalf("%s receipt must credit exactly zero: %s", start.MinigameID, result.Receipt)
		}
		after, _ := store.LoadLatest(ctx, founder.companyStreamID)
		afterCash, _ := after.State.Ledger.Balance("company.cash")
		if afterCash.String() != beforeCash.String() {
			t.Fatalf("%s moved cash %s -> %s", start.MinigameID, beforeCash, afterCash)
		}
		founderState, _ := store.LoadLatest(ctx, founder.founderStreamID)
		if rating := founderState.State.MinigameRatings[start.MinigameID]; rating.Elo != 1000 {
			t.Fatalf("%s moved rating: %+v", start.MinigameID, rating)
		}
		if quality := founderState.State.MinigameOfflineQuality[start.MinigameID]; quality.GradePPM != 200_000 {
			t.Fatalf("%s moved offline quality: %+v", start.MinigameID, quality)
		}
		history, err := store.LoadFounderHistory(ctx, founder.founderStreamID)
		if verdict := VerifyFounderHistory(history, set); err != nil || verdict != ReplayVerified {
			t.Fatalf("%s Founder history verdict=%s err=%v", start.MinigameID, verdict, err)
		}
		if active, err := repository.ActiveMinigame(ctx, founder.founderID); err != nil || active {
			t.Fatalf("a resolved arcade session must release the Exit block: active=%v err=%v", active, err)
		}
	}

	// AR1.4: always unlocked at Tier 0 with no Exit; quit clears a board.
	player := seedArcadeFounder(t, ctx, db, store, bundle, now, "01", 50)
	play(player, mineRequest(player, "01986666-a701-7000-8000-000000000003"),
		[]string{`{"kind":"choose_board","preset_id":"small"}`, `{"cell":40,"kind":"reveal"}`, `{"kind":"quit"}`}, arcade.MineGridQuit)
	// Snake on the 20x20 candidate: the head starts at x=10 facing east and
	// crashes into the wall at tick 10.
	play(player, request(player, "01986666-a701-7000-8000-000000000004", "arcade.snake", arcade.SnakeEngineRef, arcade.SnakeScalingDestination),
		[]string{`{"kind":"advance","through_tick":10,"turns":[]}`}, arcade.SnakeCrashed)
}

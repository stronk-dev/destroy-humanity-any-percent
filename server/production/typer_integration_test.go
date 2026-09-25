package production

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"os"
	"regexp"
	"testing"
	"time"

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

// typerFeatureBundle pins the complete Typer chain (TT-PA3) on top of the
// Pitch fixture bundle: the pitch+typer minigames artifact, the two-tenant
// minigame_api artifact, and the typer content artifact.
func typerFeatureBundle(t *testing.T) CatalogBundle {
	t.Helper()
	bundle := pitchFeatureBundle(t)
	bundle.Artifacts = cloneArtifactMap(bundle.Artifacts)
	read := func(path string) []byte {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		return data
	}
	bundle.Artifacts["minigames"] = read("../../testdata/minigame/pitch-typer-v3.json")
	bundle.Artifacts["minigame_api"] = read("../../balance/testdata/minigame-api-typer-candidate-v1.json")
	bundle.Artifacts["typer"] = read("../../balance/testdata/typer-v1.json")
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
	if bundle.Typer, err = typer.LoadCatalog(bundle.Artifacts["typer"], typer.Declarations{CopyKeys: keys}); err != nil {
		t.Fatal(err)
	}
	if bundle.ConstantsHash, err = save.ConstantsHashArtifacts(bundle.Artifacts); err != nil {
		t.Fatal(err)
	}
	if !bundle.valid(bundle.ConstantsHash) {
		t.Fatal("Typer fixture bundle is not internally valid")
	}
	return bundle
}

type typerFounder struct {
	founderID       string
	companyStreamID string
	founderStreamID string
}

func seedTyperFounder(t *testing.T, ctx context.Context, db *sql.DB, store *save.Store, bundle CatalogBundle, now time.Time,
	suffix string, tier int64, exits int,
) typerFounder {
	t.Helper()
	accountID := "01986666-c2" + suffix + "-4000-8000-000000000001"
	founderID := "01986666-c2" + suffix + "-4000-8000-000000000002"
	if _, err := db.ExecContext(ctx, `INSERT INTO accounts(account_id,recovery_hash) VALUES($1,'test')`, accountID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO account_founders(account_id,founder_id) VALUES($1,$2)`, accountID, founderID); err != nil {
		t.Fatal(err)
	}
	company := replayFixtureState(t, bundle.Economy, now)
	company.WireVersion, company.MeterBands, company.Tier = 16, nil, tier
	meterState, err := meters.NewRunState(bundle.Meters, 0)
	if err != nil {
		t.Fatal(err)
	}
	company.MeterValues, company.MeterDecayRemainders, company.MeterInputRemainders = meterState.Values, meterState.DecayRemainders, meterState.InputRemainders
	company.AchievementsEarnedRun = map[string]bool{}
	companyRevision, err := store.CreateStream(ctx, save.StreamKey{OwnerKind: save.OwnerFounder, OwnerID: founderID, Scope: economy.ScopeCompany},
		bundle.ConstantsHash, company, save.WriteContext{Cause: "typer.integration"})
	if err != nil {
		t.Fatal(err)
	}
	founder := replayFounderFixtureState(t, bundle, now)
	founder.WireVersion = 20
	founder.MinigameRatings = map[string]save.MinigameRatingState{"pitch": {Elo: 1000, SeasonMember: "s1"}, "typer": {Elo: 1000, SeasonMember: "s1"}}
	founder.MinigameOfflineQuality = map[string]save.MinigameOfflineQualityState{"pitch": {GradePPM: 200_000}, "typer": {GradePPM: 200_000}}
	founder.Pets = map[string]pet.CareState{}
	founder.FiscalCredit, founder.FiscalPeriodOpenedWallMS, founder.FiscalPeriodSequence = 0, now.UnixMilli(), 0
	founder.FiscalGeneratorLevels = make(map[string]int64, len(bundle.Fiscal.GeneratorLevelRows()))
	for _, row := range bundle.Fiscal.GeneratorLevelRows() {
		founder.FiscalGeneratorLevels[row.GeneratorID] = 0
	}
	founder.FiscalUnlocks, founder.Soul, founder.SoulExhaustedSourceIDs = map[string]bool{}, 50, []string{}
	founder.ExitHistory = nil
	for index := 0; index < exits; index++ {
		founder.ExitHistory = append(founder.ExitHistory, save.ExitRecord{RunID: int64(index + 1), ExitType: "scripted_first", OccurredAt: now.Add(-time.Hour)})
	}
	founderRevision, err := store.CreateStream(ctx, save.StreamKey{OwnerKind: save.OwnerFounder, OwnerID: founderID, Scope: economy.ScopeFounder},
		bundle.ConstantsHash, founder, save.WriteContext{Cause: "typer.integration"})
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

var creditedDeltaPattern = regexp.MustCompile(`"credited_delta":"([^"]+)"`)

// Typer AC8/AC9/AC11: a composed platform run through real Postgres.
func TestTyperComposedIntegrationUnlockPlayPayoutAndNeutrality(t *testing.T) {
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
	bundle := typerFeatureBundle(t)
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
	registry, err := minigame.NewTenantRegistry(pitch.NewTenant(), typer.NewTenant())
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
	startRequest := func(founder typerFounder, sessionID string) minigame.StartRequest {
		return minigame.StartRequest{SessionID: sessionID, MinigameID: "typer", FounderID: founder.founderID,
			CompanyStreamID: founder.companyStreamID, RunSeq: 1, EngineRef: typer.EngineRef, EngineVersion: typer.EngineVersion,
			ConstantsHash: bundle.ConstantsHash, ScalingInputs: map[string]int64{typer.ScalingDestination: 1}, Seed: "1", Mode: minigame.ModeSolo}
	}

	// AC9: the tier arm reads only pinned server state.
	tierZero := seedTyperFounder(t, ctx, db, store, bundle, now, "00", 0, 1)
	if _, err := service.StartMinigameSession(ctx, platform, startRequest(tierZero, "01986666-c200-7000-8000-000000000003"), now); !errors.Is(err, ErrMinigameTierRequired) {
		t.Fatalf("Tier-0 Typer start must reject tier_required: %v", err)
	}
	noExit := seedTyperFounder(t, ctx, db, store, bundle, now, "01", 1, 0)
	if _, err := service.StartMinigameSession(ctx, platform, startRequest(noExit, "01986666-c201-7000-8000-000000000003"), now); !errors.Is(err, ErrMinigameCurriculumExitRequired) {
		t.Fatalf("run-1 Typer start must reject curriculum_exit_required: %v", err)
	}

	catalog := bundle.Typer
	order := typer.PromptOrder(catalog, 1, 1)
	play := func(founder typerFounder, sessionID, assist string) (string, []byte) {
		t.Helper()
		session, err := service.StartMinigameSession(ctx, platform, startRequest(founder, sessionID), now)
		if err != nil || session.Revision != 1 {
			t.Fatalf("Typer start session=%+v err=%v", session, err)
		}
		commands := []string{`{"assist_level":"` + assist + `","kind":"begin"}`, `{"kind":"submit_line","text":"not the command"}`}
		for _, prompt := range order {
			encoded, _ := json.Marshal(map[string]string{"kind": "submit_line", "text": prompt.Text})
			commands = append(commands, string(encoded))
		}
		var resolution *minigame.CertifiedResolution
		for index, command := range commands {
			canonical, ok := canonicalCommand(command)
			if !ok {
				t.Fatalf("command %d is not canonical JSON", index)
			}
			decision, playErr := platform.Play(ctx, minigame.PlayRequest{FounderID: founder.founderID, SessionID: sessionID,
				ExpectedRevision: int64(index + 1), Command: canonical})
			if playErr != nil {
				t.Fatalf("%s command %d err=%v", assist, index, playErr)
			}
			if decision.Resolution != nil {
				resolution = decision.Resolution
			}
		}
		if resolution == nil || resolution.Result().Outcome != typer.OutcomeCompleted {
			t.Fatalf("%s run did not complete", assist)
		}
		result, err := service.ResolveMinigameSession(ctx, platform, resolution, now.Add(time.Minute), nil)
		if err != nil || result.Replay {
			t.Fatalf("%s resolve receipt=%s err=%v", assist, result.Receipt, err)
		}
		retry, err := service.ResolveMinigameSession(ctx, platform, resolution, now.Add(time.Minute), nil)
		if err != nil || !retry.Replay || !bytes.Equal(retry.Receipt, result.Receipt) {
			t.Fatalf("%s retry receipt=%s replay=%v err=%v", assist, retry.Receipt, retry.Replay, err)
		}
		match := creditedDeltaPattern.FindSubmatch(result.Receipt)
		if match == nil {
			t.Fatalf("%s receipt has no credited_delta: %s", assist, result.Receipt)
		}
		var facts map[string]int64 = map[string]int64{}
		for _, fact := range resolution.Result().ScoreFacts {
			facts[fact.Kind] = fact.Value
		}
		if facts["typer.clean_lines"] != catalog.Policy.RunLength-1 || facts["typer.lines_cleared"] != catalog.Policy.RunLength || facts["typer.misses"] != 1 {
			t.Fatalf("%s facts=%v", assist, facts)
		}
		loadedCompany, _ := store.LoadLatest(ctx, founder.companyStreamID)
		cash, _ := loadedCompany.State.Ledger.Balance("company.cash")
		if cash.String() != string(match[1]) {
			t.Fatalf("%s cash=%s credited=%s", assist, cash, match[1])
		}
		history, err := store.LoadFounderHistory(ctx, founder.founderStreamID)
		if verdict := VerifyFounderHistory(history, set); err != nil || verdict != ReplayVerified {
			t.Fatalf("%s Founder history verdict=%s err=%v", assist, verdict, err)
		}
		return string(match[1]), result.Receipt
	}
	timed := seedTyperFounder(t, ctx, db, store, bundle, now, "02", 1, 1)
	untimed := seedTyperFounder(t, ctx, db, store, bundle, now, "03", 1, 1)
	timedCredit, _ := play(timed, "01986666-c202-7000-8000-000000000003", "timed")
	untimedCredit, _ := play(untimed, "01986666-c203-7000-8000-000000000003", "untimed")
	// AC8 / TT5: end_run is legal before begin, so an open Typer session (which
	// blocks Exit, MA-C12) always has a reachable exit.
	stalled := seedTyperFounder(t, ctx, db, store, bundle, now, "04", 1, 1)
	stalledSession := "01986666-c204-7000-8000-000000000003"
	if _, err := service.StartMinigameSession(ctx, platform, startRequest(stalled, stalledSession), now); err != nil {
		t.Fatal(err)
	}
	if active, err := repository.ActiveMinigame(ctx, stalled.founderID); err != nil || !active {
		t.Fatalf("open Typer session must hold the Exit block: active=%v err=%v", active, err)
	}
	ended, err := platform.Play(ctx, minigame.PlayRequest{FounderID: stalled.founderID, SessionID: stalledSession, ExpectedRevision: 1,
		Command: json.RawMessage(`{"kind":"end_run"}`)})
	if err != nil || ended.Resolution == nil || ended.Resolution.Result().Outcome != typer.OutcomeEndedEarly {
		t.Fatalf("end_run before begin must end the run: %+v err=%v", ended, err)
	}
	if _, err := service.ResolveMinigameSession(ctx, platform, ended.Resolution, now.Add(time.Minute), nil); err != nil {
		t.Fatalf("ended_early resolution err=%v", err)
	}
	if active, err := repository.ActiveMinigame(ctx, stalled.founderID); err != nil || active {
		t.Fatalf("resolved Typer session must release the Exit block: active=%v err=%v", active, err)
	}
	t.Logf("credited timed=%s untimed=%s", timedCredit, untimedCredit)
	// AC11 (OD-2 as ruled): identical facts pay identically in both modes.
	if timedCredit != untimedCredit || timedCredit == "0" {
		t.Fatalf("payout must be mode-neutral and nonzero: timed=%s untimed=%s", timedCredit, untimedCredit)
	}
}

func canonicalCommand(command string) (json.RawMessage, bool) {
	var value any
	if json.Unmarshal([]byte(command), &value) != nil {
		return nil, false
	}
	encoded, err := json.Marshal(value)
	return encoded, err == nil
}

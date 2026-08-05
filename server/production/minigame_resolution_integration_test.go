package production

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"cloud-clicker/server/economy"
	"cloud-clicker/server/faction"
	"cloud-clicker/server/meters"
	"cloud-clicker/server/minigame"
	prestigecore "cloud-clicker/server/prestige"
	"cloud-clicker/server/routes"
	"cloud-clicker/server/save"
)

type resolutionFixtureTenant struct{}

type resolutionFixtureSnapshot struct {
	Total int64 `json:"total"`
}

type resolutionFixtureCommand struct {
	Add    int64 `json:"add"`
	Finish bool  `json:"finish"`
}

func (resolutionFixtureTenant) Descriptor() minigame.Descriptor {
	return minigame.Descriptor{EngineRef: "fixture.counter", EngineVersion: "1.0.0", CommandSchema: "fixture.command.v1",
		SnapshotSchema: "fixture.snapshot.v1", ResultSchema: "fixture.result.v1", Modes: []minigame.Mode{minigame.ModeSolo},
		ErrorTaxonomy: []string{"invalid_command"}, Destinations: map[string]minigame.DestinationClass{"option.count": minigame.DestinationBreadth}}
}
func (resolutionFixtureTenant) ValidateCommand(data json.RawMessage) error {
	var command resolutionFixtureCommand
	return decodeResolutionFixture(data, &command)
}
func (resolutionFixtureTenant) ValidateSnapshot(data json.RawMessage) error {
	var snapshot resolutionFixtureSnapshot
	return decodeResolutionFixture(data, &snapshot)
}
func (resolutionFixtureTenant) ValidateResult(result *minigame.Result) error {
	if result == nil || result.Outcome != "completed" || result.RatingDelta == nil || len(result.ScoreFacts) != 1 || result.ScoreFacts[0].Kind != "score.total" {
		return minigame.ErrInvalidTenant
	}
	return nil
}
func (resolutionFixtureTenant) Create(minigame.CreateInput) (json.RawMessage, error) {
	return json.RawMessage(`{"total":0}`), nil
}
func (resolutionFixtureTenant) Apply(input minigame.ApplyInput) (minigame.ApplyOutput, error) {
	var snapshot resolutionFixtureSnapshot
	var command resolutionFixtureCommand
	if decodeResolutionFixture(input.Snapshot, &snapshot) != nil || decodeResolutionFixture(input.Command, &command) != nil || command.Add < 0 {
		return minigame.ApplyOutput{}, minigame.ErrTenantRejected
	}
	snapshot.Total += command.Add
	encoded, _ := json.Marshal(snapshot)
	if !command.Finish {
		return minigame.ApplyOutput{Snapshot: encoded}, nil
	}
	delta := int64(25)
	return minigame.ApplyOutput{Snapshot: encoded, Result: &minigame.Result{Outcome: "completed", RatingDelta: &delta,
		ScoreFacts: []minigame.ScoreFact{{Kind: "score.total", Value: snapshot.Total}}}}, nil
}

func decodeResolutionFixture(data []byte, target any) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	if decoder.More() {
		return errors.New("trailing fixture value")
	}
	return nil
}

func TestResolveMinigameSessionIntegrationAtomicReplayAndFaults(t *testing.T) {
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
	_, active := foundationTestBundles(t)
	artifact, err := os.ReadFile("../../testdata/minigame/catalog-v2.json")
	if err != nil {
		t.Fatal(err)
	}
	bundle := active
	bundle.Artifacts = cloneArtifactMap(active.Artifacts)
	bundle.Artifacts["minigames"] = artifact
	bundle.ConstantsHash, err = save.ConstantsHashArtifacts(bundle.Artifacts)
	if err != nil {
		t.Fatal(err)
	}
	bundle.Minigames, err = minigame.LoadCatalog(artifact)
	if err != nil || !bundle.valid(bundle.ConstantsHash) {
		t.Fatalf("content bundle err=%v", err)
	}
	seedProductionEpoch(t, db, bundle.ConstantsHash, bundle.Artifacts)
	resolver := integrationCatalogs{economy: map[string]*economy.Catalog{bundle.ConstantsHash: bundle.Economy},
		routes: map[string]*routes.Catalog{bundle.ConstantsHash: bundle.Routes}, prestige: map[string]*prestigecore.Policy{bundle.ConstantsHash: bundle.Prestige},
		factions: map[string]*faction.Catalog{bundle.ConstantsHash: bundle.Faction}}
	store, err := save.NewStore(db, resolver, nil)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 8, 5, 17, 0, 0, 0, time.UTC)
	const accountID = "01986666-a900-4000-8000-000000000001"
	const founderID = "01986666-a900-4000-8000-000000000002"
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
		bundle.ConstantsHash, company, save.WriteContext{Cause: "minigame.resolve.integration"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.PinRunToCurrentEpoch(ctx, companyRevision.StreamID, founderID, 1, bundle.ConstantsHash); err != nil {
		t.Fatal(err)
	}
	founder := replayFounderFixtureState(t, bundle, now)
	founder.MinigameRatings["fixture.counter"] = save.MinigameRatingState{Elo: 1000, SeasonMember: "ranked", GamesCounted: 0}
	founder.MinigameOfflineQuality["fixture.counter"] = save.MinigameOfflineQualityState{GradePPM: 500_000}
	founderRevision, err := store.CreateStream(ctx, save.StreamKey{OwnerKind: save.OwnerFounder, OwnerID: founderID, Scope: economy.ScopeFounder},
		bundle.ConstantsHash, founder, save.WriteContext{Cause: "minigame.resolve.integration"})
	if err != nil {
		t.Fatal(err)
	}
	repository, err := minigame.NewRepository(db)
	if err != nil {
		t.Fatal(err)
	}
	registry, err := minigame.NewTenantRegistry(resolutionFixtureTenant{})
	if err != nil {
		t.Fatal(err)
	}
	platform, err := minigame.NewService(repository, registry)
	if err != nil {
		t.Fatal(err)
	}
	production, err := NewService(store, resolver, nil, nil, nil, WithProgressionRuntime(resolver), WithCurrentConstantsHash(bundle.ConstantsHash),
		WithReplayCatalogs(ReplayCatalogSet{bundle.ConstantsHash: bundle}), WithGuildSettlements(emptyGuildSettlements{}))
	if err != nil {
		t.Fatal(err)
	}
	makeResolution := func(index int) *minigame.CertifiedResolution {
		t.Helper()
		sessionID := fmt.Sprintf("01986666-a9%02x-7000-8000-%012d", index, index)
		if _, startErr := platform.Start(ctx, minigame.StartRequest{SessionID: sessionID, MinigameID: "fixture.counter", FounderID: founderID,
			CompanyStreamID: companyRevision.StreamID, RunSeq: 1, EngineRef: "fixture.counter", EngineVersion: "1.0.0",
			ConstantsHash: bundle.ConstantsHash, ScalingInputs: map[string]int64{"option.count": 3}, Seed: "1", Mode: minigame.ModeSolo}); startErr != nil {
			t.Fatal(startErr)
		}
		decision, playErr := platform.Play(ctx, minigame.PlayRequest{FounderID: founderID, SessionID: sessionID, ExpectedRevision: 1,
			Command: json.RawMessage(`{"add":400,"finish":true}`)})
		if playErr != nil || decision.Resolution == nil {
			t.Fatalf("play resolution=%v err=%v", decision.Resolution, playErr)
		}
		return decision.Resolution
	}
	genesisFault := makeResolution(99)
	if _, err := production.ResolveMinigameSession(ctx, platform, genesisFault, now.Add(time.Minute), func(step string) error {
		if step == "founder_genesis" {
			return errors.New("injected Founder genesis fault")
		}
		return nil
	}); err == nil {
		t.Fatal("Founder genesis fault committed")
	}
	if current, loadErr := store.LoadLatest(ctx, founderRevision.StreamID); loadErr != nil || current.Revision.Number != 1 {
		t.Fatalf("Founder genesis fault leaked revision=%d err=%v", current.Revision.Number, loadErr)
	}
	resolution := makeResolution(1)
	result, err := production.ResolveMinigameSession(ctx, platform, resolution, now.Add(time.Minute), nil)
	if err != nil || result.Replay || !bytes.Contains(result.Receipt, []byte(`"credited_delta":"5e1"`)) {
		t.Fatalf("resolution receipt=%s replay=%v err=%v", result.Receipt, result.Replay, err)
	}
	retry, err := production.ResolveMinigameSession(ctx, platform, resolution, now.Add(time.Minute), nil)
	if err != nil || !retry.Replay || !bytes.Equal(retry.Receipt, result.Receipt) {
		t.Fatalf("retry receipt=%s replay=%v err=%v", retry.Receipt, retry.Replay, err)
	}
	loadedCompany, _ := store.LoadLatest(ctx, companyRevision.StreamID)
	loadedFounder, _ := store.LoadLatest(ctx, founderRevision.StreamID)
	cash, _ := loadedCompany.State.Ledger.Balance("company.cash")
	if loadedCompany.Revision.Number != 2 || loadedFounder.Revision.Number != 2 || cash.String() != "5e1" ||
		loadedFounder.State.MinigameRatings["fixture.counter"].Elo != 1025 || loadedFounder.State.MinigameOfflineQuality["fixture.counter"].GradePPM != 750_000 {
		t.Fatalf("company=%d cash=%s founder=%d rating=%+v quality=%+v", loadedCompany.Revision.Number, cash, loadedFounder.Revision.Number,
			loadedFounder.State.MinigameRatings["fixture.counter"], loadedFounder.State.MinigameOfflineQuality["fixture.counter"])
	}
	founderHistory, err := store.LoadFounderHistory(ctx, founderRevision.StreamID)
	if err != nil || VerifyFounderHistory(founderHistory, ReplayCatalogSet{bundle.ConstantsHash: bundle}) != ReplayVerified {
		t.Fatalf("Founder minigame history did not verify: %v", err)
	}
	var companyLogs, founderLogs, resolutionEvents, quota int
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM run_log WHERE company_stream_id=$1 AND (convert_from(canonical_payload,'UTF8')::jsonb)->>'kind'='resolve_minigame_session'`, companyRevision.StreamID).Scan(&companyLogs); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM founder_log WHERE founder_stream_id=$1 AND (convert_from(canonical_payload,'UTF8')::jsonb)->>'kind'='resolve_minigame_session'`, founderRevision.StreamID).Scan(&founderLogs); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM events WHERE intent_id=$1`, "01986666-a901-7000-8000-000000000001").Scan(&resolutionEvents); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT quota_used FROM minigame_faucet_window WHERE founder_id=$1 AND minigame_id='fixture.counter'`, founderID).Scan(&quota); err != nil {
		t.Fatal(err)
	}
	if companyLogs != 1 || founderLogs != 1 || resolutionEvents != 2 || quota != 1 {
		t.Fatalf("logs=%d/%d events=%d quota=%d", companyLogs, founderLogs, resolutionEvents, quota)
	}
	faultSteps := []string{"faucet_window", "session_terminal", "founder_revision", "founder_events", "company_revision", "company_events", "run_log", "founder_log", "intent_record", "retention"}
	for index, step := range faultSteps {
		resolution := makeResolution(index + 2)
		beforeCompany, _ := store.LoadLatest(ctx, companyRevision.StreamID)
		beforeFounder, _ := store.LoadLatest(ctx, founderRevision.StreamID)
		_, resolveErr := production.ResolveMinigameSession(ctx, platform, resolution, now.Add(2*time.Minute), func(actual string) error {
			if actual == step {
				return errors.New("injected minigame resolution fault")
			}
			return nil
		})
		if resolveErr == nil {
			t.Fatalf("fault %s committed", step)
		}
		afterCompany, _ := store.LoadLatest(ctx, companyRevision.StreamID)
		afterFounder, _ := store.LoadLatest(ctx, founderRevision.StreamID)
		view, _ := resolution.View()
		session, loadErr := repository.Load(ctx, founderID, view.SessionID)
		if loadErr != nil || afterCompany.Revision.Number != beforeCompany.Revision.Number || afterFounder.Revision.Number != beforeFounder.Revision.Number ||
			session.Status != minigame.StatusClaimed || session.Result != nil {
			t.Fatalf("fault %s leaked company=%d/%d founder=%d/%d session=%+v err=%v", step, beforeCompany.Revision.Number,
				afterCompany.Revision.Number, beforeFounder.Revision.Number, afterFounder.Revision.Number, session, loadErr)
		}
	}
}

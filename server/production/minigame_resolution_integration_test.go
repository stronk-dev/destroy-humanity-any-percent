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

type resolutionFixtureTenant struct {
	observe func(minigame.ApplyInput)
}

type resolutionFixtureSnapshot struct {
	Total int64 `json:"total"`
}

type resolutionFixtureCommand struct {
	Add    int64 `json:"add"`
	Finish bool  `json:"finish"`
}

func (resolutionFixtureTenant) Descriptor() minigame.Descriptor {
	return minigame.Descriptor{EngineRef: "fixture.counter", EngineVersion: "1.0.0", CommandSchema: "fixture.command.v1",
		SnapshotSchema: "fixture.snapshot.v1", ResultSchema: "fixture.result.v1", Modes: []minigame.Mode{minigame.ModeSolo, minigame.ModeAsyncSnapshot},
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
	if result == nil {
		return nil // Nonterminal play has no certified result yet (C14).
	}
	if result.Outcome != "completed" || result.RatingDelta == nil || len(result.ScoreFacts) != 1 || result.ScoreFacts[0].Kind != "score.total" {
		return minigame.ErrInvalidTenant
	}
	return nil
}
func (resolutionFixtureTenant) Create(minigame.CreateInput) (json.RawMessage, error) {
	return json.RawMessage(`{"total":0}`), nil
}
func (tenant resolutionFixtureTenant) Apply(input minigame.ApplyInput) (minigame.ApplyOutput, error) {
	if tenant.observe != nil {
		tenant.observe(input)
	}
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
	// A parent with only skipped children reports PASS. Keep this prerequisite
	// at the observed top level so the composed lane cannot count an absent DB.
	if os.Getenv("TEST_DATABASE_URL") == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}
	for _, mode := range []minigame.Mode{minigame.ModeSolo, minigame.ModeAsyncSnapshot} {
		t.Run(string(mode), func(t *testing.T) {
			testResolveMinigameSessionAtomicReplayAndFaults(t, mode)
		})
	}
}

func testResolveMinigameSessionAtomicReplayAndFaults(t *testing.T, mode minigame.Mode) {
	t.Helper()
	databaseURL := os.Getenv("TEST_DATABASE_URL")
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
	registry, err := minigame.NewTenantRegistry(resolutionFixtureTenant{observe: func(input minigame.ApplyInput) {
		t.Helper()
		if input.Mode != mode || len(input.ScalingInputs) != 1 || input.ScalingInputs["option.count"] != 3 {
			t.Fatalf("tenant lost persisted mode/frozen scaling: mode=%s scaling=%v", input.Mode, input.ScalingInputs)
		}
	}})
	if err != nil {
		t.Fatal(err)
	}
	platform, err := minigame.NewService(repository, registry)
	if err != nil {
		t.Fatal(err)
	}
	projectionFailure := &failNextProjection{}
	production, err := NewService(store, resolver, nil, nil, nil, WithProgressionRuntime(resolver), WithCurrentConstantsHash(bundle.ConstantsHash),
		WithReplayCatalogs(ReplayCatalogSet{bundle.ConstantsHash: bundle}), WithGuildSettlements(emptyGuildSettlements{}), WithEventProjector(projectionFailure))
	if err != nil {
		t.Fatal(err)
	}
	makeResolution := func(index int) *minigame.CertifiedResolution {
		t.Helper()
		sessionID := fmt.Sprintf("01986666-a9%02x-7000-8000-%012d", index, index)
		scaling := map[string]int64{"option.count": 3}
		if _, startErr := platform.Start(ctx, minigame.StartRequest{SessionID: sessionID, MinigameID: "fixture.counter", FounderID: founderID,
			CompanyStreamID: companyRevision.StreamID, RunSeq: 1, EngineRef: "fixture.counter", EngineVersion: "1.0.0",
			ConstantsHash: bundle.ConstantsHash, ScalingInputs: scaling, Seed: "1", Mode: mode}); startErr != nil {
			t.Fatal(startErr)
		}
		// Later source changes cannot rewrite this session's frozen values.
		scaling["option.count"] = 8
		decision, playErr := platform.Play(ctx, minigame.PlayRequest{FounderID: founderID, SessionID: sessionID, ExpectedRevision: 1,
			Command: json.RawMessage(`{"add":100,"finish":false}`)})
		if playErr != nil || decision.Resolution != nil || decision.Session.Revision != 2 {
			t.Fatalf("nonterminal play revision=%d resolution=%v err=%v", decision.Session.Revision, decision.Resolution, playErr)
		}
		// Construct new service/repository objects and resume from the SQL row.
		// This is an in-process reconstruction, not a database/host crash claim.
		resumedRepository, resumeErr := minigame.NewRepository(db)
		if resumeErr != nil {
			t.Fatal(resumeErr)
		}
		platform, resumeErr = minigame.NewService(resumedRepository, registry)
		if resumeErr != nil {
			t.Fatal(resumeErr)
		}
		persisted, resumeErr := platform.Load(ctx, founderID, sessionID)
		if resumeErr != nil || persisted.Mode != mode || persisted.Revision != 2 || persisted.Status != minigame.StatusActive ||
			!bytes.Equal(persisted.State, []byte(`{"total":100}`)) {
			t.Fatalf("persisted resume did not retain its mode/state: session=%+v err=%v", persisted, resumeErr)
		}
		decision, playErr = platform.Play(ctx, minigame.PlayRequest{FounderID: founderID, SessionID: sessionID, ExpectedRevision: 2,
			Command: json.RawMessage(`{"add":300,"finish":true}`)})
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
	if _, err := db.ExecContext(ctx, `DELETE FROM minigame_sessions WHERE session_id=$1`, "01986666-a963-7000-8000-000000000099"); err != nil {
		t.Fatal(err)
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

	apiSessionID := "01986666-a964-7000-8000-000000000100"
	if _, err := platform.Start(ctx, minigame.StartRequest{SessionID: apiSessionID, MinigameID: "fixture.counter", FounderID: founderID,
		CompanyStreamID: companyRevision.StreamID, RunSeq: 1, EngineRef: "fixture.counter", EngineVersion: "1.0.0",
		ConstantsHash: bundle.ConstantsHash, ScalingInputs: map[string]int64{"option.count": 3}, Seed: "2", Mode: mode}); err != nil {
		t.Fatal(err)
	}
	apiCommand := PlayMinigameAPIRequest{FounderID: founderID, SessionID: apiSessionID, CommandID: "terminal-1", ExpectedRevision: 1,
		Command: json.RawMessage(`{"add":400,"finish":true}`)}
	apiResult, err := production.PlayMinigameAPICommand(ctx, platform, apiCommand, now.Add(90*time.Second), nil)
	var apiResponse map[string]json.RawMessage
	if err != nil || apiResult.Replay || json.Unmarshal(apiResult.Receipt, &apiResponse) != nil || len(apiResponse) != 10 ||
		string(apiResponse["status"]) != `"resolved"` || len(apiResponse["resolution_receipt"]) == 0 {
		t.Fatalf("API terminal receipt=%s replay=%v err=%v", apiResult.Receipt, apiResult.Replay, err)
	}
	apiRetry, err := production.PlayMinigameAPICommand(ctx, platform, apiCommand, now.Add(2*time.Minute), nil)
	if err != nil || !apiRetry.Replay || !bytes.Equal(apiRetry.Receipt, apiResult.Receipt) {
		t.Fatalf("API retry receipt=%s replay=%v err=%v", apiRetry.Receipt, apiRetry.Replay, err)
	}
	conflict := apiCommand
	conflict.Command = json.RawMessage(`{"add":401,"finish":true}`)
	if _, err := production.PlayMinigameAPICommand(ctx, platform, conflict, now.Add(2*time.Minute), nil); !errors.Is(err, minigame.ErrAPIIdempotency) {
		t.Fatalf("API command hash conflict err=%v", err)
	}
	apiSession, err := repository.Load(ctx, founderID, apiSessionID)
	if err != nil || apiSession.Status != minigame.StatusResolved || apiSession.Revision != 2 ||
		bytes.Contains(apiSession.ResolutionReceipt, []byte(`"snapshot"`)) {
		t.Fatalf("API terminal session=%+v err=%v", apiSession, err)
	}
	var commandReceipts int
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM minigame_command_receipts WHERE session_id=$1`, apiSessionID).Scan(&commandReceipts); err != nil || commandReceipts != 1 {
		t.Fatalf("API command receipts=%d err=%v", commandReceipts, err)
	}

	for index, failure := range []struct {
		step   string
		cancel bool
	}{{step: "company_revision"}, {step: "retention"}, {step: "company_revision", cancel: true}} {
		t.Run(fmt.Sprintf("API rollback retry %s cancel=%t", failure.step, failure.cancel), func(t *testing.T) {
			faultSessionID := fmt.Sprintf("01986666-a966-7000-8000-%012d", 300+index)
			started, startErr := platform.Start(ctx, minigame.StartRequest{SessionID: faultSessionID, MinigameID: "fixture.counter", FounderID: founderID,
				CompanyStreamID: companyRevision.StreamID, RunSeq: 1, EngineRef: "fixture.counter", EngineVersion: "1.0.0",
				ConstantsHash: bundle.ConstantsHash, ScalingInputs: map[string]int64{"option.count": 3}, Seed: "3", Mode: mode})
			if startErr != nil {
				t.Fatal(startErr)
			}
			beforeCompany, loadErr := store.LoadLatest(ctx, companyRevision.StreamID)
			if loadErr != nil {
				t.Fatal(loadErr)
			}
			beforeFounder, loadErr := store.LoadLatest(ctx, founderRevision.StreamID)
			if loadErr != nil {
				t.Fatal(loadErr)
			}
			window := func() string {
				t.Helper()
				var value string
				if queryErr := db.QueryRowContext(ctx, `SELECT row_to_json(w)::text FROM minigame_faucet_window w WHERE founder_id=$1 AND minigame_id='fixture.counter'`, founderID).Scan(&value); queryErr != nil {
					t.Fatal(queryErr)
				}
				return value
			}
			beforeWindow := window()
			command := PlayMinigameAPIRequest{FounderID: founderID, SessionID: faultSessionID, CommandID: "terminal-fault", ExpectedRevision: 1,
				Command: json.RawMessage(`{"add":400,"finish":true}`)}
			failedCtx, cancel := context.WithCancel(ctx)
			defer cancel()
			injected := errors.New("injected API terminal fault")
			fired := false
			_, playErr := production.PlayMinigameAPICommand(failedCtx, platform, command, now.Add(2*time.Minute), func(step string) error {
				if step != failure.step {
					return nil
				}
				fired = true
				if failure.cancel {
					cancel()
				}
				return injected
			})
			if !fired || !errors.Is(playErr, injected) {
				t.Fatalf("terminal fault fired=%t err=%v", fired, playErr)
			}
			failedSession, loadErr := repository.Load(ctx, founderID, faultSessionID)
			if loadErr != nil || failedSession.Status != minigame.StatusActive || failedSession.ClaimToken != "" || failedSession.ClaimedAt != nil ||
				failedSession.Revision != started.Revision || !bytes.Equal(failedSession.State, started.State) || len(failedSession.ResolutionReceipt) != 0 {
				t.Fatalf("failed terminal command must release its claim without applying: session=%+v err=%v", failedSession, loadErr)
			}
			for _, before := range []save.Loaded{beforeCompany, beforeFounder} {
				after, readErr := store.LoadLatest(ctx, before.Revision.StreamID)
				if readErr != nil || after.Revision.Number != before.Revision.Number || !bytes.Equal(mustEncodeState(t, after.State), mustEncodeState(t, before.State)) {
					t.Fatalf("failed API terminal command changed stream %s: err=%v", before.Revision.StreamID, readErr)
				}
			}
			var receipts, commands int
			if queryErr := db.QueryRowContext(ctx, `SELECT (SELECT count(*) FROM minigame_command_receipts WHERE session_id=$1),(SELECT count(*) FROM minigame_session_commands WHERE session_id=$1)`, faultSessionID).Scan(&receipts, &commands); queryErr != nil || receipts != 0 || commands != 0 || window() != beforeWindow {
				t.Fatalf("rollback leaked receipts=%d commands=%d err=%v", receipts, commands, queryErr)
			}
			// No lease ageing, deletion, replacement command ID or service restart:
			// the exact failed request can complete immediately and then replay.
			applied, retryErr := production.PlayMinigameAPICommand(ctx, platform, command, now.Add(2*time.Minute), nil)
			if retryErr != nil || applied.Replay {
				t.Fatalf("immediate retry replay=%t err=%v", applied.Replay, retryErr)
			}
			appliedHeads := make([]save.Loaded, 0, 2)
			for _, before := range []save.Loaded{beforeCompany, beforeFounder} {
				head, readErr := store.LoadLatest(ctx, before.Revision.StreamID)
				if readErr != nil || head.Revision.Number != before.Revision.Number+1 {
					t.Fatalf("retry did not advance stream %s once: err=%v", before.Revision.StreamID, readErr)
				}
				appliedHeads = append(appliedHeads, head)
			}
			beforeCash, _ := beforeCompany.State.Ledger.Balance("company.cash")
			paidCash, _ := appliedHeads[0].State.Ledger.Balance("company.cash")
			if paidCash.Sub(beforeCash).String() != "5e1" || !bytes.Contains(applied.Receipt, []byte(`"credited_delta":"5e1"`)) {
				t.Fatalf("retry payout cash=%s/%s receipt=%s", beforeCash, paidCash, applied.Receipt)
			}
			appliedWindow := window()
			replayed, retryErr := production.PlayMinigameAPICommand(ctx, platform, command, now.Add(2*time.Minute), nil)
			if retryErr != nil || !replayed.Replay || !bytes.Equal(applied.Receipt, replayed.Receipt) {
				t.Fatalf("committed retry replay=%t equal=%t err=%v", replayed.Replay, bytes.Equal(applied.Receipt, replayed.Receipt), retryErr)
			}
			terminal, loadErr := repository.Load(ctx, founderID, faultSessionID)
			if loadErr != nil || terminal.Status != minigame.StatusResolved || terminal.Revision != started.Revision+1 {
				t.Fatalf("retried terminal session=%+v err=%v", terminal, loadErr)
			}
			for _, appliedHead := range appliedHeads {
				after, readErr := store.LoadLatest(ctx, appliedHead.Revision.StreamID)
				if readErr != nil || after.Revision.Number != appliedHead.Revision.Number || !bytes.Equal(mustEncodeState(t, after.State), mustEncodeState(t, appliedHead.State)) {
					t.Fatalf("durable replay changed stream %s: err=%v", appliedHead.Revision.StreamID, readErr)
				}
			}
			var events int
			if queryErr := db.QueryRowContext(ctx, `SELECT (SELECT count(*) FROM minigame_command_receipts WHERE session_id=$1),(SELECT count(*) FROM minigame_session_commands WHERE session_id=$1),(SELECT count(*) FROM events WHERE intent_id=$1)`, faultSessionID).Scan(&receipts, &commands, &events); queryErr != nil || receipts != 1 || commands != 1 || events != 2 || window() != appliedWindow {
				t.Fatalf("retry duplicated committed work: receipts=%d commands=%d events=%d err=%v", receipts, commands, events, queryErr)
			}
		})
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
		if _, err := db.ExecContext(ctx, `DELETE FROM minigame_sessions WHERE session_id=$1`, view.SessionID); err != nil {
			t.Fatal(err)
		}
	}

	t.Run("API post-commit projection failure preserves terminal receipt", func(t *testing.T) {
		const sessionID = "01986666-a967-7000-8000-000000000400"
		if _, err := platform.Start(ctx, minigame.StartRequest{SessionID: sessionID, MinigameID: "fixture.counter", FounderID: founderID,
			CompanyStreamID: companyRevision.StreamID, RunSeq: 1, EngineRef: "fixture.counter", EngineVersion: "1.0.0",
			ConstantsHash: bundle.ConstantsHash, ScalingInputs: map[string]int64{"option.count": 3}, Seed: "3", Mode: mode}); err != nil {
			t.Fatal(err)
		}
		command := PlayMinigameAPIRequest{FounderID: founderID, SessionID: sessionID, CommandID: "post-commit", ExpectedRevision: 1,
			Command: json.RawMessage(`{"add":400,"finish":true}`)}
		projectionFailure.fail = true
		if _, err := production.PlayMinigameAPICommand(ctx, platform, command, now.Add(3*time.Minute), nil); !errors.Is(err, ErrInvalidEngineState) || errors.Is(err, minigame.ErrClaimLost) || projectionFailure.fail {
			t.Fatalf("post-commit projection error=%v fired=%t", err, !projectionFailure.fail)
		}
		committed, loadErr := repository.Load(ctx, founderID, sessionID)
		if loadErr != nil || committed.Status != minigame.StatusResolved || committed.Revision != 2 || len(committed.ResolutionReceipt) == 0 {
			t.Fatalf("cleanup reopened committed session=%+v err=%v", committed, loadErr)
		}
		var storedResponse []byte
		if err := db.QueryRowContext(ctx, `SELECT response::text FROM minigame_command_receipts WHERE session_id=$1 AND command_id='post-commit'`, sessionID).Scan(&storedResponse); err != nil {
			t.Fatal(err)
		}
		canonicalResponse, err := normalizeReplayJSON(storedResponse)
		if err != nil {
			t.Fatal(err)
		}
		retried, err := production.PlayMinigameAPICommand(ctx, platform, command, now.Add(3*time.Minute), nil)
		if err != nil || !retried.Replay || !bytes.Equal(retried.Receipt, canonicalResponse) {
			t.Fatalf("post-commit receipt retry replay=%t equal=%t err=%v", retried.Replay, bytes.Equal(retried.Receipt, canonicalResponse), err)
		}
	})

	// A resolution whose faucet credits nothing (zero score here; an exhausted
	// daily window or a saturated hardcap reach the same empty ledger receipt)
	// must still commit its terminal receipt. Otherwise the claimed session is
	// stranded and blocks Exit (MA-C12).
	zeroSessionID := "01986666-a965-7000-8000-000000000200"
	if _, err := platform.Start(ctx, minigame.StartRequest{SessionID: zeroSessionID, MinigameID: "fixture.counter", FounderID: founderID,
		CompanyStreamID: companyRevision.StreamID, RunSeq: 1, EngineRef: "fixture.counter", EngineVersion: "1.0.0",
		ConstantsHash: bundle.ConstantsHash, ScalingInputs: map[string]int64{"option.count": 3}, Seed: "4", Mode: mode}); err != nil {
		t.Fatal(err)
	}
	zeroPlay, err := platform.Play(ctx, minigame.PlayRequest{FounderID: founderID, SessionID: zeroSessionID, ExpectedRevision: 1,
		Command: json.RawMessage(`{"add":0,"finish":true}`)})
	if err != nil || zeroPlay.Resolution == nil {
		t.Fatalf("zero play resolution=%v err=%v", zeroPlay.Resolution, err)
	}
	beforeZero, _ := store.LoadLatest(ctx, companyRevision.StreamID)
	beforeCash, _ := beforeZero.State.Ledger.Balance("company.cash")
	zeroResult, err := production.ResolveMinigameSession(ctx, platform, zeroPlay.Resolution, now.Add(4*time.Minute), nil)
	if err != nil || zeroResult.Replay || !bytes.Contains(zeroResult.Receipt, []byte(`"credited_delta":"0"`)) {
		t.Fatalf("zero-credit resolution receipt=%s replay=%v err=%v", zeroResult.Receipt, zeroResult.Replay, err)
	}
	zeroSession, err := repository.Load(ctx, founderID, zeroSessionID)
	afterZero, _ := store.LoadLatest(ctx, companyRevision.StreamID)
	afterCash, _ := afterZero.State.Ledger.Balance("company.cash")
	if err != nil || zeroSession.Status != minigame.StatusResolved || afterZero.Revision.Number != beforeZero.Revision.Number+1 || !afterCash.Eq(beforeCash) {
		t.Fatalf("zero-credit session=%+v company=%d/%d cash=%s/%s err=%v", zeroSession, beforeZero.Revision.Number,
			afterZero.Revision.Number, beforeCash, afterCash, err)
	}
	founderHistory, err = store.LoadFounderHistory(ctx, founderRevision.StreamID)
	if err != nil || VerifyFounderHistory(founderHistory, ReplayCatalogSet{bundle.ConstantsHash: bundle}) != ReplayVerified {
		t.Fatalf("zero-credit Founder history did not verify: %v", err)
	}
}

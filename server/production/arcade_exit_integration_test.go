package production

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"testing"
	"time"

	"cloud-clicker/server/arcade"
	"cloud-clicker/server/commons"
	"cloud-clicker/server/commonsbinding"
	"cloud-clicker/server/copykeys"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/faction"
	"cloud-clicker/server/meters"
	"cloud-clicker/server/minigame"
	"cloud-clicker/server/minigameapi"
	"cloud-clicker/server/pet"
	"cloud-clicker/server/pitch"
	prestigecore "cloud-clicker/server/prestige"
	"cloud-clicker/server/routeprojection"
	"cloud-clicker/server/routes"
	"cloud-clicker/server/save"
)

// Preserve the entire current curriculum/economy; add only the two accepted
// Arcade fixture rows. This is not a minted epoch or a public-wire contract.
func currentArcadeExitBundle(t *testing.T) CatalogBundle {
	t.Helper()
	current := activeContentBundle(t)
	bundle := current
	bundle.Artifacts = cloneArtifactMap(current.Artifacts)
	read := func(path string) []byte {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		return data
	}
	appendRows := func(name, field, idField, path string, include func(map[string]json.RawMessage) bool) {
		var target, candidate map[string]json.RawMessage
		if json.Unmarshal(bundle.Artifacts[name], &target) != nil || json.Unmarshal(read(path), &candidate) != nil {
			t.Fatal("invalid candidate/current artifact")
		}
		var rows, extra []map[string]json.RawMessage
		if json.Unmarshal(target[field], &rows) != nil || json.Unmarshal(candidate[field], &extra) != nil {
			t.Fatal("invalid candidate/current rows")
		}
		original := append([]map[string]json.RawMessage(nil), rows...)
		for _, row := range extra {
			if include(row) {
				rows = append(rows, row)
			}
		}
		sort.Slice(rows, func(i, j int) bool { return string(rows[i][idField]) < string(rows[j][idField]) })
		var err error
		if target[field], err = json.Marshal(rows); err != nil {
			t.Fatal(err)
		}
		if bundle.Artifacts[name], err = json.Marshal(target); err != nil {
			t.Fatal(err)
		}
		if len(rows) != len(original)+2 {
			t.Fatalf("%s must add exactly two Arcade rows", name)
		}
		for _, prior := range original {
			matched := false
			for _, row := range rows {
				if string(row[idField]) == string(prior[idField]) {
					before, _ := json.Marshal(prior)
					after, _ := json.Marshal(row)
					matched = bytes.Equal(before, after)
				}
			}
			if !matched {
				t.Fatalf("%s changed or lost an original current row", name)
			}
		}
	}
	appendRows("minigames", "minigames", "minigame_id", "../../testdata/minigame/pitch-typer-arcade-v3.json", func(row map[string]json.RawMessage) bool {
		return string(row["minigame_id"]) == `"arcade.mine_grid"` || string(row["minigame_id"]) == `"arcade.snake"`
	})
	appendRows("minigame_api", "tenants", "engine_ref", "../../balance/testdata/minigame-api-arcade-candidate-v1.json", func(row map[string]json.RawMessage) bool {
		return string(row["engine_ref"]) == `"mine_grid"` || string(row["engine_ref"]) == `"snake"`
	})
	bundle.Artifacts["arcade"] = read("../../balance/testdata/arcade-v1.json")
	var err error
	if bundle.Minigames, err = minigame.LoadCatalog(bundle.Artifacts["minigames"]); err != nil {
		t.Fatal(err)
	}
	if bundle.MinigameAPI, err = minigameapi.LoadCatalog(bundle.Artifacts["minigame_api"]); err != nil {
		t.Fatal(err)
	}
	keys := map[string]struct{}{}
	for _, key := range copykeys.All() {
		keys[key] = struct{}{}
	}
	if bundle.Arcade, err = arcade.LoadCatalog(bundle.Artifacts["arcade"], arcade.Declarations{CopyKeys: keys}); err != nil {
		t.Fatal(err)
	}
	if bundle.ConstantsHash, err = save.ConstantsHashArtifacts(bundle.Artifacts); err != nil || !bundle.valid(bundle.ConstantsHash) {
		t.Fatalf("current Arcade test bundle invalid: %v", err)
	}
	for name, data := range current.Artifacts {
		if name != "minigames" && name != "minigame_api" && !bytes.Equal(data, bundle.Artifacts[name]) {
			t.Fatalf("test changed current artifact %s", name)
		}
	}
	return bundle
}

func seedCurrentArcadeExitFounder(t *testing.T, ctx context.Context, db *sql.DB, store *save.Store, bundle CatalogBundle, now time.Time, index int) typerFounder {
	t.Helper()
	accountID := fmt.Sprintf("01986666-aa00-4000-8000-%012d", 2*index+1)
	founderID := fmt.Sprintf("01986666-aa00-4000-8000-%012d", 2*index+2)
	if _, err := db.ExecContext(ctx, `INSERT INTO accounts(account_id,recovery_hash) VALUES($1,'test')`, accountID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO account_founders(account_id,founder_id) VALUES($1,$2)`, accountID, founderID); err != nil {
		t.Fatal(err)
	}
	founder := replayFounderFixtureState(t, bundle, now)
	founder.WireVersion = 21
	if err := activateMinigameState(founder, bundle.Minigames); err != nil {
		t.Fatal(err)
	}
	founder.Pets = map[string]pet.CareState{}
	founder.FiscalPeriodOpenedWallMS, founder.FiscalGeneratorLevels, founder.FiscalUnlocks = now.UnixMilli(), map[string]int64{}, map[string]bool{}
	for _, row := range bundle.Fiscal.GeneratorLevelRows() {
		founder.FiscalGeneratorLevels[row.GeneratorID] = 0
	}
	founder.Soul, founder.SoulExhaustedSourceIDs = 80, []string{}
	started := now.Add(-time.Duration(bundle.Curriculum.FirstFailure.AttendedMS) * time.Millisecond)
	company := replayFixtureState(t, bundle.Economy, started)
	company.WireVersion, company.MeterBands = 18, nil
	company.Tier = 1
	company.GatesCrossed[bundle.Curriculum.FirstFailure.GateID] = true
	setCash(t, company, "1e5")
	meterState, err := meters.NewRunState(bundle.Meters, 0)
	if err != nil {
		t.Fatal(err)
	}
	company.MeterValues, company.MeterDecayRemainders, company.MeterInputRemainders = meterState.Values, meterState.DecayRemainders, meterState.InputRemainders
	company.AchievementsEarnedRun = map[string]bool{}
	if _, err := initializeActivePlayState(company, bundle.Opportunities, founderID); err != nil {
		t.Fatal(err)
	}
	advanceActivePlayFixtureAttendance(t, company, bundle.Opportunities, bundle.Prestige, founderID, now)
	company.ManualTokenRefilledAt = now
	founderRevision, err := store.CreateStream(ctx, save.StreamKey{OwnerKind: save.OwnerFounder, OwnerID: founderID, Scope: economy.ScopeFounder}, bundle.ConstantsHash, founder, save.WriteContext{Cause: "arcade.exit.integration"})
	if err != nil {
		t.Fatal(err)
	}
	companyRevision, err := store.CreateStream(ctx, save.StreamKey{OwnerKind: save.OwnerFounder, OwnerID: founderID, Scope: economy.ScopeCompany}, bundle.ConstantsHash, company, save.WriteContext{Cause: "arcade.exit.integration"})
	if err != nil {
		t.Fatal(err)
	}
	genesis := mustEncodeState(t, company)
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()
	if _, err := save.PinRunWithGenesisTx(ctx, tx, companyRevision.StreamID, founderID, 1, bundle.ConstantsHash, companyRevision.Version, genesis); err != nil {
		t.Fatal(err)
	}
	frozen, err := FrozenFiscalContributions(bundle.Fiscal, founder)
	if err != nil {
		t.Fatal(err)
	}
	if err := save.InsertRunFrozenContributionsTx(ctx, tx, companyRevision.StreamID, 1, frozen); err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
	return typerFounder{founderID: founderID, companyStreamID: companyRevision.StreamID, founderStreamID: founderRevision.StreamID, runSeq: 1}
}

func TestArcadeCurrentCurriculumExitIntegration(t *testing.T) {
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
	bundle := currentArcadeExitBundle(t)
	seedProductionEpoch(t, db, bundle.ConstantsHash, bundle.Artifacts)
	resolver := integrationCatalogs{economy: map[string]*economy.Catalog{bundle.ConstantsHash: bundle.Economy},
		routes: map[string]*routes.Catalog{bundle.ConstantsHash: bundle.Routes}, prestige: map[string]*prestigecore.Policy{bundle.ConstantsHash: bundle.Prestige},
		factions: map[string]*faction.Catalog{bundle.ConstantsHash: bundle.Faction}}
	store, err := save.NewStore(db, resolver, nil)
	if err != nil {
		t.Fatal(err)
	}
	repository, err := minigame.NewRepository(db)
	if err != nil {
		t.Fatal(err)
	}
	registry, err := minigame.NewTenantRegistry(pitch.NewTenant(), arcade.NewMineGridTenant(), arcade.NewSnakeTenant())
	if err != nil {
		t.Fatal(err)
	}
	set := ReplayCatalogSet{bundle.ConstantsHash: bundle}
	platform, err := minigame.NewService(repository, registry, set)
	if err != nil {
		t.Fatal(err)
	}
	routeProjector, err := routeprojection.New(db, resolver)
	if err != nil {
		t.Fatal(err)
	}
	service, err := NewService(store, resolver, FrozenContributionProvider{DB: db}, nil, nil,
		WithRouteCatalogs(resolver), WithRouteProjector(routeProjector),
		WithProgressionRuntime(resolver), WithCurrentConstantsHash(bundle.ConstantsHash), WithReplayCatalogs(set),
		WithGuildSettlements(emptyGuildSettlements{}), WithMinigameActivity(repository),
		WithCompactPolicies(commons.CatalogSet{bundle.ConstantsHash: bundle.Commons.(commonsbinding.ReplayPolicy).Catalog}), WithCommonsWeightResolver(integrationWeight(1_000_000)))
	if err != nil {
		t.Fatal(err)
	}
	now := save.CanonicalServerTime(time.Now().UTC())
	index, intentIndex := 0, 0
	seed := func(t *testing.T) typerFounder {
		index++
		return seedCurrentArcadeExitFounder(t, ctx, db, store, bundle, now, index)
	}
	intent := func(t *testing.T, founder typerFounder, kind string) []byte {
		t.Helper()
		company, err := store.LoadLatest(ctx, founder.companyStreamID)
		if err != nil {
			t.Fatal(err)
		}
		owner, err := store.LoadLatest(ctx, founder.founderStreamID)
		if err != nil {
			t.Fatal(err)
		}
		intentIndex++
		payload := map[string]any{"intent_id": fmt.Sprintf("01986666-aa01-7000-8000-%012d", intentIndex), "kind": kind, "expected_revision": company.Revision.Number}
		switch kind {
		case IntentWindDown:
			payload["expected_founder_revision"] = owner.Revision.Number
		case IntentPerformManualBatch:
			payload["action_id"], payload["count"], payload["window_ms"] = "manual.click", 1, 1
		}
		encoded, err := json.Marshal(payload)
		if err != nil {
			t.Fatal(err)
		}
		return encoded
	}
	receiptMatches := func(data json.RawMessage, outcome, category, detail string) bool {
		var receipt struct {
			Outcome   string `json:"outcome"`
			Rejection struct {
				Category string `json:"category"`
				Detail   string `json:"detail"`
			} `json:"rejection"`
		}
		return json.Unmarshal(data, &receipt) == nil && receipt.Outcome == outcome && receipt.Rejection.Category == category && receipt.Rejection.Detail == detail
	}
	for _, kind := range []string{IntentWindDown, IntentPerformManualBatch} {
		t.Run("eligible_no_session/"+kind, func(t *testing.T) {
			founder := seed(t)
			result, err := service.Handle(ctx, founder.companyStreamID, ModeOnline, now, intent(t, founder, kind))
			if err != nil || !receiptMatches(result.Receipt, "applied", "", "") {
				t.Fatalf("eligible no-session control failed: receipt=%s err=%v", result.Receipt, err)
			}
			company, err := store.LoadLatest(ctx, founder.companyStreamID)
			if err != nil || company.State.RunSeq != 2 {
				t.Fatalf("control did not Exit: err=%v", err)
			}
		})
	}
	for _, row := range []struct {
		toy, phase  string
		chooseBoard bool
	}{
		{"arcade.mine_grid", "setup", false},
		{"arcade.mine_grid", "playing", true},
		{"arcade.snake", "playing", false},
	} {
		toy := row.toy
		for _, kind := range []string{IntentWindDown, IntentPerformManualBatch} {
			t.Run(toy+"/"+row.phase+"/"+kind, func(t *testing.T) {
				founder := seed(t)
				sessionID := fmt.Sprintf("01986666-aa02-7000-8000-%012d", index)
				if _, err := service.StartMinigameAPISession(ctx, platform, StartMinigameAPIRequest{FounderID: founder.founderID,
					CompanyStreamID: founder.companyStreamID, SessionID: sessionID, IntentID: fmt.Sprintf("01986666-aa03-7000-8000-%012d", index),
					MinigameID: toy, IdempotencyKey: "arcade-exit-" + sessionID}, now, nil); err != nil {
					t.Fatal(err)
				}
				revision := int64(1)
				if row.chooseBoard {
					chosen, err := platform.Play(ctx, minigame.PlayRequest{FounderID: founder.founderID, SessionID: sessionID,
						ExpectedRevision: revision, Command: json.RawMessage(`{"kind":"choose_board","preset_id":"small"}`)})
					if err != nil || chosen.Resolution != nil {
						t.Fatalf("choose board did not enter playing: err=%v", err)
					}
					revision++
				}
				started, err := repository.Load(ctx, founder.founderID, sessionID)
				var snapshot struct {
					Phase string `json:"phase"`
				}
				if err != nil || json.Unmarshal(started.State, &snapshot) != nil || snapshot.Phase != row.phase || started.Revision != revision {
					t.Fatalf("wrong actual session phase/revision: phase=%s revision=%d err=%v", snapshot.Phase, started.Revision, err)
				}
				assertBlocked := func(status minigame.Status) {
					t.Helper()
					before := []save.Loaded{}
					for _, stream := range []string{founder.companyStreamID, founder.founderStreamID} {
						loaded, err := store.LoadLatest(ctx, stream)
						if err != nil {
							t.Fatal(err)
						}
						before = append(before, loaded)
					}
					session, err := repository.Load(ctx, founder.founderID, sessionID)
					if err != nil || session.Status != status {
						t.Fatalf("expected %s session: err=%v", status, err)
					}
					commandRows := func() string {
						var encoded string
						if err := db.QueryRowContext(ctx, `SELECT COALESCE(jsonb_agg(to_jsonb(c) ORDER BY seq),'[]'::jsonb)::text
FROM minigame_session_commands c WHERE session_id=$1`, sessionID).Scan(&encoded); err != nil {
							t.Fatal(err)
						}
						return encoded
					}
					beforeCommands := commandRows()
					blocked, err := service.Handle(ctx, founder.companyStreamID, ModeOnline, save.CanonicalServerTime(time.Now().UTC()), intent(t, founder, kind))
					if err != nil || !receiptMatches(blocked.Receipt, "rejected", "not_eligible", "minigame_session_active") {
						t.Fatalf("%s %s did not block due Exit: receipt=%s err=%v", status, toy, blocked.Receipt, err)
					}
					for _, loaded := range before {
						after, err := store.LoadLatest(ctx, loaded.Revision.StreamID)
						if err != nil || after.Revision.Number != loaded.Revision.Number || after.Revision.ConstantsHash != loaded.Revision.ConstantsHash || !bytes.Equal(mustEncodeState(t, after.State), mustEncodeState(t, loaded.State)) {
							t.Fatalf("blocked Exit mutated stream %s: err=%v", loaded.Revision.StreamID, err)
						}
					}
					after, err := repository.Load(ctx, founder.founderID, sessionID)
					if err != nil || after.Status != session.Status || after.Revision != session.Revision || after.ClaimToken != session.ClaimToken || !bytes.Equal(after.State, session.State) || !bytes.Equal(after.Genesis, session.Genesis) || !bytes.Equal(after.Result, session.Result) {
						t.Fatalf("blocked Exit changed minigame session: err=%v", err)
					}
					if commandRows() != beforeCommands {
						t.Fatal("blocked Exit changed complete minigame command rows")
					}
				}
				assertBlocked(minigame.StatusActive)
				quit, err := platform.Play(ctx, minigame.PlayRequest{FounderID: founder.founderID, SessionID: sessionID, ExpectedRevision: revision, Command: json.RawMessage(`{"kind":"quit"}`)})
				if err != nil || quit.Resolution == nil || quit.Resolution.Result().Outcome != "quit" {
					t.Fatalf("quit did not reach terminal: err=%v", err)
				}
				assertBlocked(minigame.StatusClaimed)
				if _, err := service.ResolveMinigameSession(ctx, platform, quit.Resolution, save.CanonicalServerTime(time.Now().UTC()), nil); err != nil {
					t.Fatal(err)
				}
				payload := intent(t, founder, kind)
				exited, err := service.Handle(ctx, founder.companyStreamID, ModeOnline, save.CanonicalServerTime(time.Now().UTC()), payload)
				if err != nil || !receiptMatches(exited.Receipt, "applied", "", "") {
					t.Fatalf("resolved quit did not release due Exit: receipt=%s err=%v", exited.Receipt, err)
				}
				retry, err := service.Handle(ctx, founder.companyStreamID, ModeOnline, now, payload)
				if err != nil || !retry.Replay || !bytes.Equal(exited.Receipt, retry.Receipt) {
					t.Fatalf("Exit retry changed result: err=%v", err)
				}
				company, err := store.LoadLatest(ctx, founder.companyStreamID)
				if err != nil || company.State.RunSeq != 2 {
					t.Fatalf("Exit did not start run 2: err=%v", err)
				}
				owner, err := store.LoadLatest(ctx, founder.founderStreamID)
				if err != nil || len(owner.State.ExitHistory) != 1 || owner.State.ExitHistory[0].ExitType != "scripted_first" {
					t.Fatalf("Exit did not persist first ending: err=%v", err)
				}
				history, err := store.LoadFounderHistory(ctx, founder.founderStreamID)
				if err != nil || VerifyFounderHistory(history, set) != ReplayVerified {
					t.Fatalf("Exit Founder replay failed: err=%v", err)
				}
				genesis, version, entries := persistedPrestigeReplay(t, db, founder.companyStreamID, founder.founderStreamID, 1, &bundle)
				// Exit entries own both streams' events; ordinary Company commands
				// (including minigame resolution) own only Company-stream events.
				for i := range entries {
					if entries[i].Terminal {
						continue
					}
					wire, err := parseReplayInputs(entries[i].ReplayInputs)
					if err != nil {
						t.Fatal(err)
					}
					rows, err := db.QueryContext(ctx, `SELECT kind,schema_version,intent_id,payload FROM events
WHERE stream_id=$1 AND intent_id=$2 ORDER BY event_seq,event_id`, founder.companyStreamID, wire.Command.IntentID)
					if err != nil {
						t.Fatal(err)
					}
					events := []save.EventWrite{}
					for rows.Next() {
						var event save.EventWrite
						if err := rows.Scan(&event.Kind, &event.SchemaVersion, &event.IntentID, &event.Payload); err != nil {
							rows.Close()
							t.Fatal(err)
						}
						events = append(events, event)
					}
					err = rows.Err()
					rows.Close()
					if err != nil {
						t.Fatal(err)
					}
					entries[i].EventsJSON = marshalReplayEvents(events)
				}
				if verdict := VerifyReplayRun(genesis, version, bundle, entries, bundle.ConstantsHash, false); verdict != ReplayVerified {
					t.Fatalf("Exit Company replay failed: %s", verdict)
				}
			})
		}
	}
}

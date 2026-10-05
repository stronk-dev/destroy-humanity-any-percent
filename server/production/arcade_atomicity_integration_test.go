package production

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"testing"
	"time"

	"cloud-clicker/server/arcade"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/faction"
	"cloud-clicker/server/minigame"
	"cloud-clicker/server/pitch"
	prestigecore "cloud-clicker/server/prestige"
	"cloud-clicker/server/routes"
	"cloud-clicker/server/save"
	"cloud-clicker/server/typer"
)

func TestArcadeRejectedAdvanceIntegrationPreservesPersistedSessionAndCommands(t *testing.T) {
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
	now := save.CanonicalServerTime(time.Now().UTC())
	founder := seedArcadeFounder(t, ctx, db, store, bundle, now, "02", 50)
	sessionID := "01986666-a702-7000-8000-000000000003"
	started, err := service.StartMinigameSession(ctx, platform, minigame.StartRequest{SessionID: sessionID, MinigameID: "arcade.snake",
		FounderID: founder.founderID, CompanyStreamID: founder.companyStreamID, RunSeq: 1, EngineRef: arcade.SnakeEngineRef,
		EngineVersion: arcade.EngineVersion, ConstantsHash: bundle.ConstantsHash, ScalingInputs: map[string]int64{arcade.SnakeScalingDestination: 1}, Seed: "1", Mode: minigame.ModeSolo}, now)
	if err != nil || started.Revision != 1 {
		t.Fatalf("actual session start: revision=%d err=%v", started.Revision, err)
	}

	var fixture struct {
		Version int `json:"version"`
		Cases   []struct {
			Name    string `json:"name"`
			Through string `json:"through"`
			Turns   []struct {
				Tick      string `json:"tick"`
				Direction string `json:"direction"`
			} `json:"turns"`
			Expected string `json:"expected"`
		} `json:"cases"`
	}
	data, err := os.ReadFile("../../testdata/arcade/snake-rejection-atomicity-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.Version != 1 || len(fixture.Cases) != 6 {
		t.Fatal("unexpected shared rejection population")
	}
	load := func() minigame.Session {
		value, err := repository.Load(ctx, founder.founderID, sessionID)
		if err != nil {
			t.Fatal(err)
		}
		return value
	}
	commands := func() (string, int) {
		var aggregate string
		var count int
		if err := db.QueryRowContext(ctx, `SELECT COALESCE(jsonb_agg(to_jsonb(c) ORDER BY seq),'[]'::jsonb)::text, count(*)
FROM minigame_session_commands c WHERE session_id=$1`, sessionID).Scan(&aggregate, &count); err != nil {
			t.Fatal(err)
		}
		return aggregate, count
	}
	play := func(command string, revision int64) (minigame.PlayDecision, error) {
		canonical, ok := canonicalCommand(command)
		if !ok {
			t.Fatal("test command is not valid canonical JSON")
		}
		return platform.Play(ctx, minigame.PlayRequest{FounderID: founder.founderID, SessionID: sessionID, ExpectedRevision: revision, Command: canonical})
	}
	for _, stage := range []string{"genesis", "moved"} {
		if stage == "moved" {
			before := load()
			var prior arcade.SnakeSnapshot
			if err := json.Unmarshal(before.State, &prior); err != nil {
				t.Fatal(err)
			}
			_, count := commands()
			decision, err := play(`{"kind":"advance","through_tick":1,"turns":[]}`, before.Revision)
			after := load()
			_, afterCount := commands()
			var actual arcade.SnakeSnapshot
			if err != nil || decision.Resolution != nil || json.Unmarshal(after.State, &actual) != nil || actual.Tick != prior.Tick+1 || actual.Body[0] != prior.Body[0]+1 ||
				after.Revision != before.Revision+1 || afterCount != count+1 || bytes.Equal(after.State, before.State) {
				t.Fatalf("accepted advance failed its real DB positive control: err=%v", err)
			}
		}
		for _, row := range fixture.Cases {
			t.Run(stage+"/"+row.Name, func(t *testing.T) {
				before := load()
				beforeRows, beforeCount := commands()
				if beforeCount != int(before.Revision-1) || beforeCount != map[string]int{"genesis": 0, "moved": 1}[stage] {
					t.Fatal("wrong actual command-history population")
				}
				var state arcade.SnakeSnapshot
				if err := json.Unmarshal(before.State, &state); err != nil {
					t.Fatal(err)
				}
				if state.Width != 20 || state.Height != 20 || state.Direction != "right" {
					t.Fatal("wrong candidate/actual direction population")
				}
				times := map[string]int64{"current_tick": state.Tick, "next_tick": state.Tick + 1,
					"second_tick": state.Tick + 2, "after_death": state.Tick + state.Width - state.Body[0]%state.Width + 1}
				resolve := func(key string) int64 {
					value, ok := times[key]
					if !ok {
						t.Fatalf("unknown time marker %s", key)
					}
					return value
				}
				turns := []map[string]any{}
				for _, turn := range row.Turns {
					turns = append(turns, map[string]any{"tick": resolve(turn.Tick), "direction": turn.Direction})
				}
				command, err := json.Marshal(map[string]any{"kind": "advance", "through_tick": resolve(row.Through), "turns": turns})
				if err != nil {
					t.Fatal(err)
				}
				decision, playErr := play(string(command), before.Revision)
				var rejection *minigame.Rejection
				if !errors.As(playErr, &rejection) || rejection.Code != row.Expected || decision.Resolution != nil {
					t.Fatalf("expected %s with no resolution, got resolution=%t err=%v", row.Expected, decision.Resolution != nil, playErr)
				}
				after := load()
				afterRows, afterCount := commands()
				if after.Revision != before.Revision || !bytes.Equal(after.State, before.State) || !bytes.Equal(after.Genesis, before.Genesis) ||
					!bytes.Equal(after.Result, before.Result) || !after.CreatedAt.Equal(before.CreatedAt) || afterRows != beforeRows || afterCount != beforeCount {
					t.Fatalf("rejected advance mutated durable state/history: revision %d→%d commands %d→%d", before.Revision, after.Revision, beforeCount, afterCount)
				}
				if after.Status != minigame.StatusActive || after.ClaimToken != "" || after.ClaimedAt != nil {
					t.Fatal("rejected advance did not release its actual claim")
				}
				if active, err := repository.ActiveMinigame(ctx, founder.founderID); err != nil || !active {
					t.Fatalf("rejected play lost the active session: active=%v err=%v", active, err)
				}
			})
		}
	}
	current := load()
	decision, err := play(`{"kind":"quit"}`, current.Revision)
	if err != nil || decision.Resolution == nil || decision.Resolution.Result().Outcome != arcade.SnakeQuit {
		t.Fatalf("quit unusable after actual rejections: %v", err)
	}
	resolved, err := service.ResolveMinigameSession(ctx, platform, decision.Resolution, now.Add(time.Minute), nil)
	if err != nil || resolved.Replay {
		t.Fatalf("real resolution failed: %v", err)
	}
	if final := load(); final.Status != minigame.StatusResolved || final.Revision != current.Revision+1 {
		t.Fatal("actual terminal was not committed")
	}
	if _, count := commands(); count != 2 {
		t.Fatal("rejected commands entered history or accepted commands disappeared")
	}
	if active, err := repository.ActiveMinigame(ctx, founder.founderID); err != nil || active {
		t.Fatalf("resolved quit did not release exclusivity: active=%v err=%v", active, err)
	}
	t.Log("12 actual rejected advances preserve state/revision/history and release claims; one accepted advance and quit commit exactly two rows")
}

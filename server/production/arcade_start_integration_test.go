package production

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"strconv"
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

// AR-P3's atomic coordinator, not a public handler/socket or a Pitch-less
// catalog. RP-201 routes the unresolved catalog dependency separately.
func TestArcadeAtomicStartIntegrationCreatesBothPinnedTenants(t *testing.T) {
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
		routes:   map[string]*routes.Catalog{bundle.ConstantsHash: bundle.Routes},
		prestige: map[string]*prestigecore.Policy{bundle.ConstantsHash: bundle.Prestige},
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
	service, err := NewService(store, resolver, nil, nil, nil, WithProgressionRuntime(resolver),
		WithCurrentConstantsHash(bundle.ConstantsHash), WithReplayCatalogs(set),
		WithGuildSettlements(emptyGuildSettlements{}), WithMinigameActivity(repository))
	if err != nil {
		t.Fatal(err)
	}
	now := save.CanonicalServerTime(time.Now().UTC())
	for _, row := range []struct{ suffix, toy, engine, destination string }{
		{"20", "arcade.mine_grid", arcade.MineGridEngineRef, arcade.MineGridScalingDestination},
		{"21", "arcade.snake", arcade.SnakeEngineRef, arcade.SnakeScalingDestination},
	} {
		t.Run(row.toy, func(t *testing.T) {
			founder := seedArcadeFounderVersion(t, ctx, db, store, bundle, now, row.suffix, 50, 21)
			request := StartMinigameAPIRequest{SessionID: "01986666-a8" + row.suffix + "-7000-8000-000000000003",
				IntentID: "01986666-a9" + row.suffix + "-7000-8000-000000000004", FounderID: founder.founderID,
				CompanyStreamID: founder.companyStreamID, MinigameID: row.toy, IdempotencyKey: "atomic-arcade-" + row.suffix}
			created, err := service.StartMinigameAPISession(ctx, platform, request, now, nil)
			if err != nil || created.Replay {
				t.Fatalf("atomic start receipt=%s replay=%v err=%v", created.Receipt, created.Replay, err)
			}
			session, err := repository.Load(ctx, founder.founderID, request.SessionID)
			if err != nil {
				t.Fatal(err)
			}
			wantSeed := minigameSessionSeed(founder.founderID, 1, 1)
			if session.Seed != wantSeed || session.Revision != 1 || session.Status != minigame.StatusActive || session.EngineRef != row.engine ||
				session.EngineVersion != arcade.EngineVersion || session.MinigameID != row.toy || session.ConstantsHash != bundle.ConstantsHash {
				t.Fatalf("wrong persisted start identity: %+v", session)
			}
			content, ok := set.ResolveTenantContent(bundle.ConstantsHash, row.engine, arcade.EngineVersion)
			if !ok || !bytes.Equal(content.Bytes, bundle.Artifacts["arcade"]) || content.Hash != arcade.ContentHash(bundle.Artifacts["arcade"]) {
				t.Fatal("start content is not the pinned Arcade artifact")
			}
			seed, err := strconv.ParseUint(wantSeed, 10, 64)
			if err != nil {
				t.Fatal(err)
			}
			genesis, err := registry.Create(row.engine, arcade.EngineVersion, minigame.CreateInput{Mode: minigame.ModeSolo, Seed: seed,
				ScalingInputs: map[string]int64{row.destination: 1}, Content: content.Bytes, ContentHash: content.Hash, ContentSchemaVersion: content.SchemaVersion})
			if err != nil {
				t.Fatal(err)
			}
			canonical, err := normalizeReplayJSON(genesis)
			if err != nil || !bytes.Equal(canonical, session.Genesis) || !bytes.Equal(canonical, session.State) {
				t.Fatalf("persisted genesis differs from actual tenant: %v", err)
			}
			var receipt struct {
				SessionID string          `json:"session_id"`
				Snapshot  json.RawMessage `json:"snapshot"`
			}
			if json.Unmarshal(created.Receipt, &receipt) != nil || receipt.SessionID != request.SessionID || !bytes.Equal(receipt.Snapshot, canonical) {
				t.Fatalf("wrong atomic create receipt: %s", created.Receipt)
			}
			retry := request
			retry.SessionID = "01986666-a8" + row.suffix + "-7000-8000-000000000005"
			retried, err := service.StartMinigameAPISession(ctx, platform, retry, now.Add(time.Second), nil)
			if err != nil || !retried.Replay || !bytes.Equal(retried.Receipt, created.Receipt) {
				t.Fatalf("retry changed create receipt: %s err=%v", retried.Receipt, err)
			}
			latest, err := store.LoadLatest(ctx, founder.founderStreamID)
			if err != nil || latest.Revision.Number != 2 || latest.State.MinigameSessionSeq != 1 || save.VersionForState(latest.State) != 21 {
				t.Fatalf("atomic start/retry advanced Founder incorrectly: %+v err=%v", latest, err)
			}
			company, err := store.LoadLatest(ctx, founder.companyStreamID)
			if err != nil || company.Revision.Number != 1 {
				t.Fatalf("start changed Company: %+v err=%v", company, err)
			}
			var sessions, receipts int
			if err := db.QueryRowContext(ctx, `SELECT
				(SELECT count(*) FROM minigame_sessions WHERE founder_id=$1),
				(SELECT count(*) FROM minigame_create_receipts WHERE founder_id=$1)`, founder.founderID).Scan(&sessions, &receipts); err != nil || sessions != 1 || receipts != 1 {
				t.Fatalf("start/retry cardinality: sessions=%d receipts=%d err=%v", sessions, receipts, err)
			}
			history, err := store.LoadFounderHistory(ctx, founder.founderStreamID)
			if err != nil || VerifyFounderHistory(history, set) != ReplayVerified {
				t.Fatalf("atomic start Founder replay failed: %v", err)
			}
		})
	}
}

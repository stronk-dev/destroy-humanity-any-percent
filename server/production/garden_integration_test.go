package production

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"cloud-clicker/server/economy"
	"cloud-clicker/server/faction"
	"cloud-clicker/server/meters"
	prestigecore "cloud-clicker/server/prestige"
	"cloud-clicker/server/routes"
	"cloud-clicker/server/save"
)

// TestGardenIntegrationPersistsReplayableFounderLog is SG5/SG8 through
// Service.Handle on Postgres. It covers:
//   - the live salt draw and idempotent retry;
//   - a client tick field rejected (AC12);
//   - the advance pre-step on a garden command and on spend_fiscal_credit
//     (server_ms is the database clock, so real-time ticks are ~0 here; tick
//     behaviour is the corpus's job);
//   - harvest failing closed until the SG-P2 coordinator lands;
//   - no salt in any receipt or stored event (AC10's persisted half);
//   - an unregistered event kind refused by the database;
//   - Founder history replaying clean.
func TestGardenIntegrationPersistsReplayableFounderLog(t *testing.T) {
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
	bundle := gardenContentBundle(t)
	seedProductionEpoch(t, db, bundle.ConstantsHash, bundle.Artifacts)
	resolver := integrationCatalogs{economy: map[string]*economy.Catalog{bundle.ConstantsHash: bundle.Economy},
		routes: map[string]*routes.Catalog{bundle.ConstantsHash: bundle.Routes}, prestige: map[string]*prestigecore.Policy{bundle.ConstantsHash: bundle.Prestige},
		factions: map[string]*faction.Catalog{bundle.ConstantsHash: bundle.Faction}}
	store, err := save.NewStore(db, resolver, nil)
	if err != nil {
		t.Fatal(err)
	}
	cursor := save.CanonicalServerTime(time.Date(2026, 9, 25, 12, 0, 0, 0, time.UTC))
	const accountID = "01986666-7f00-4000-8000-000000000001"
	const founderID = "01986666-7f00-7000-8000-000000000002"
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
	company.WireVersion, company.MeterBands = 16, nil
	company.MeterValues, company.MeterDecayRemainders, company.MeterInputRemainders = meterState.Values, meterState.DecayRemainders, meterState.InputRemainders
	company.AchievementsEarnedRun = map[string]bool{}
	company.RunStartedAt = cursor
	companyRevision, err := store.CreateStream(ctx, save.StreamKey{OwnerKind: save.OwnerFounder, OwnerID: founderID, Scope: economy.ScopeCompany},
		bundle.ConstantsHash, company, save.WriteContext{Cause: "garden.integration"})
	if err != nil {
		t.Fatal(err)
	}
	founder := reputationFounderState(t, bundle, 25, cursor, 0)
	founder.FiscalUnlocks[bundle.Garden.UnlockID] = true
	founder.FiscalCredit = 40
	founderRevision, err := store.CreateStream(ctx, save.StreamKey{OwnerKind: save.OwnerFounder, OwnerID: founderID, Scope: economy.ScopeFounder},
		bundle.ConstantsHash, founder, save.WriteContext{Cause: "garden.integration"})
	if err != nil {
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
		WithReplayCatalogs(ReplayCatalogSet{bundle.ConstantsHash: bundle}), WithGuildSettlements(emptyGuildSettlements{}))
	if err != nil {
		t.Fatal(err)
	}
	handle := func(at time.Duration, body string) HandleResult {
		t.Helper()
		result, err := service.Handle(ctx, companyRevision.StreamID, ModeOnline, cursor.Add(at), []byte(body))
		if err != nil {
			t.Fatalf("%s: %v", body, err)
		}
		return result
	}
	plant := `{"intent_id":"01986666-7f00-7000-8000-000000000003","kind":"garden_plant","expected_revision":1,"row":0,"col":0,"species_id":"strain_a"}`
	planted := handle(time.Second, plant)
	if planted.Replay || !strings.Contains(string(planted.Receipt), `"outcome":"applied"`) || !strings.Contains(string(planted.Receipt), `"garden_advance"`) {
		t.Fatalf("plant receipt=%s", planted.Receipt)
	}
	if retry := handle(2*time.Second, plant); !retry.Replay || string(retry.Receipt) != string(planted.Receipt) {
		t.Fatalf("retry=%s replay=%v", retry.Receipt, retry.Replay)
	}
	if forged := handle(3*time.Second, `{"intent_id":"01986666-7f00-7000-8000-000000000004","kind":"garden_plant","expected_revision":2,"row":0,"col":1,"species_id":"strain_b","tick_seq":9}`); !strings.Contains(string(forged.Receipt), `"category":"invalid"`) {
		t.Fatalf("client tick field receipt=%s", forged.Receipt)
	}
	handle(4*time.Second, `{"intent_id":"01986666-7f00-7000-8000-000000000005","kind":"garden_plant","expected_revision":2,"row":1,"col":1,"species_id":"strain_b"}`)
	switched := handle(2*time.Hour, `{"intent_id":"01986666-7f00-7000-8000-000000000006","kind":"garden_set_substrate","expected_revision":3,"substrate_id":"containerized"}`)
	if !strings.Contains(string(switched.Receipt), `"outcome":"applied"`) {
		t.Fatalf("substrate receipt=%s", switched.Receipt)
	}
	spent := handle(3*time.Hour, `{"intent_id":"01986666-7f00-7000-8000-000000000007","kind":"spend_fiscal_credit","expected_revision":4,"target":{"kind":"generator_level","generator_id":"generator.beige_tower","levels":1}}`)
	if !strings.Contains(string(spent.Receipt), `"garden_advance"`) || !strings.Contains(string(spent.Receipt), `"outcome":"applied"`) {
		t.Fatalf("spend_fiscal_credit must carry the garden pre-step: %s", spent.Receipt)
	}
	if _, err := service.Handle(ctx, companyRevision.StreamID, ModeOnline, cursor.Add(4*time.Hour),
		[]byte(`{"intent_id":"01986666-7f00-7000-8000-000000000008","kind":"garden_harvest","expected_revision":5,"plots":[{"row":0,"col":0}]}`)); !errors.Is(err, ErrInvalidIntent) {
		t.Fatalf("harvest must fail closed until the SG-P2 coordinator lands: %v", err)
	}
	loaded, err := store.LoadLatest(ctx, founderRevision.StreamID)
	if err != nil || loaded.Revision.Number != 5 || save.VersionForState(loaded.State) != 25 || loaded.State.ServerGarden.SaltHex == nil || len(loaded.State.ServerGarden.Plots) != 2 ||
		loaded.State.ServerGarden.SubstrateID != "containerized" {
		t.Fatalf("Founder after garden commands revision=%d garden=%+v err=%v", loaded.Revision.Number, loaded.State.ServerGarden, err)
	}
	salt := *loaded.State.ServerGarden.SaltHex
	var leaks int
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM events WHERE stream_id=$1 AND payload::text LIKE '%'||$2||'%'`, founderRevision.StreamID, salt).Scan(&leaks); err != nil || leaks != 0 {
		t.Fatalf("the hidden salt appears in %d stored events (err=%v)", leaks, err)
	}
	for _, receipt := range [][]byte{planted.Receipt, switched.Receipt, spent.Receipt} {
		if strings.Contains(string(receipt), salt) {
			t.Fatalf("the hidden salt appears in a receipt: %s", receipt)
		}
	}
	var gardenEvents int
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM events WHERE stream_id=$1 AND kind LIKE 'garden_%'`, founderRevision.StreamID).Scan(&gardenEvents); err != nil || gardenEvents != 3 {
		t.Fatalf("garden events=%d err=%v", gardenEvents, err)
	}
	if _, err := db.ExecContext(ctx, `UPDATE events SET kind='garden_watered.v1' WHERE stream_id=$1 AND kind='garden_planted.v1'`, founderRevision.StreamID); err == nil {
		t.Fatal("events constraint accepted garden_watered.v1")
	}
	history, err := store.LoadFounderHistory(ctx, founderRevision.StreamID)
	if err != nil {
		t.Fatal(err)
	}
	if verdict := VerifyFounderHistory(history, ReplayCatalogSet{bundle.ConstantsHash: bundle}); verdict != ReplayVerified {
		t.Fatalf("Founder history verdict=%v", verdict)
	}
}

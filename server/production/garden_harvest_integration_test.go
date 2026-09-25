package production

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"cloud-clicker/server/decimal"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/faction"
	"cloud-clicker/server/garden"
	"cloud-clicker/server/meters"
	prestigecore "cloud-clicker/server/prestige"
	"cloud-clicker/server/routes"
	"cloud-clicker/server/save"
)

type gardenHarvestFixture struct {
	ctx             context.Context
	db              *sql.DB
	store           *save.Store
	service         *Service
	bundle          CatalogBundle
	founderID       string
	companyStreamID string
	founderStreamID string
}

// newGardenHarvestFixture seeds a v25 Founder whose garden already holds
// mature plants (Postgres stamps server_ms from its own clock, so plants
// cannot be grown in real time inside a test).
func newGardenHarvestFixture(t *testing.T) gardenHarvestFixture {
	t.Helper()
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}
	ctx := context.Background()
	db, err := save.OpenPostgres(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
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
	cursor := save.CanonicalServerTime(time.Now().Add(-time.Minute))
	const accountID = "01986666-7f10-4000-8000-000000000001"
	const founderID = "01986666-7f10-7000-8000-000000000002"
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
		bundle.ConstantsHash, company, save.WriteContext{Cause: "garden.harvest.integration"})
	if err != nil {
		t.Fatal(err)
	}
	founder := reputationFounderState(t, bundle, 25, cursor, 0)
	founder.FiscalUnlocks[bundle.Garden.UnlockID] = true
	founder.FiscalGeneratorLevels[bundle.Garden.HostGeneratorID] = 8
	salt, anchor := "0123456789abcdef", cursor.UnixMilli()
	founder.ServerGarden.SaltHex, founder.ServerGarden.TickAnchorWallMS = &salt, &anchor
	effect := int64(1_000_000)
	plots := []garden.Plot{}
	for _, row := range []int64{0, 1} {
		for col := int64(0); col < 6; col++ {
			plots = append(plots, garden.Plot{Row: row, Col: col, SpeciesID: "strain_c", AgeTicks: 6, MaturedEffectPPM: &effect})
		}
	}
	plots = append(plots, garden.Plot{Row: 2, Col: 0, SpeciesID: "strain_d", AgeTicks: 8, MaturedEffectPPM: &effect},
		garden.Plot{Row: 2, Col: 1, SpeciesID: "strain_e", AgeTicks: 5, MaturedEffectPPM: &effect},
		garden.Plot{Row: 2, Col: 2, SpeciesID: "strain_a", AgeTicks: 1})
	founder.ServerGarden.Plots = plots
	founderRevision, err := store.CreateStream(ctx, save.StreamKey{OwnerKind: save.OwnerFounder, OwnerID: founderID, Scope: economy.ScopeFounder},
		bundle.ConstantsHash, founder, save.WriteContext{Cause: "garden.harvest.integration"})
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
	return gardenHarvestFixture{ctx: ctx, db: db, store: store, service: service, bundle: bundle, founderID: founderID,
		companyStreamID: companyRevision.StreamID, founderStreamID: founderRevision.StreamID}
}

func (fixture gardenHarvestFixture) harvest(t *testing.T, sequence, expected int, plots string) HandleResult {
	t.Helper()
	body := fmt.Sprintf(`{"intent_id":"01986666-7f10-7000-8000-%012d","kind":"garden_harvest","expected_revision":%d,"plots":[%s]}`, sequence, expected, plots)
	result, err := fixture.service.Handle(fixture.ctx, fixture.companyStreamID, ModeOnline, time.Now(), []byte(body))
	if err != nil {
		t.Fatalf("harvest %d: %v", sequence, err)
	}
	return result
}

func (fixture gardenHarvestFixture) revisions(t *testing.T) (int64, int64) {
	t.Helper()
	founder, err := fixture.store.LoadLatest(fixture.ctx, fixture.founderStreamID)
	if err != nil {
		t.Fatal(err)
	}
	company, err := fixture.store.LoadLatest(fixture.ctx, fixture.companyStreamID)
	if err != nil {
		t.Fatal(err)
	}
	return founder.Revision.Number, company.Revision.Number
}

func (fixture gardenHarvestFixture) quota(t *testing.T) int64 {
	t.Helper()
	var used int64
	err := fixture.db.QueryRowContext(fixture.ctx, `SELECT COALESCE(sum(quota_used),0) FROM minigame_faucet_window WHERE founder_id=$1 AND minigame_id='server_garden'`, fixture.founderID).Scan(&used)
	if err != nil {
		t.Fatal(err)
	}
	return used
}

// TestGardenHarvestIntegration is AC8/AC9 through Service.Handle on Postgres:
// credit, retry, zero-harvest, daily-send forfeit with seeds, rejection
// isolation, cross-log harvest_hash binding, and replay of both logs.
func TestGardenHarvestIntegration(t *testing.T) {
	fixture := newGardenHarvestFixture(t)
	founderRevision, companyRevision := fixture.revisions(t)
	if founderRevision != 1 || companyRevision != 1 {
		t.Fatalf("fixture revisions %d/%d", founderRevision, companyRevision)
	}
	// A rejected harvest is Founder-only: the Company and the window are untouched.
	if rejected := fixture.harvest(t, 1, 1, `{"row":2,"col":2}`); !strings.Contains(string(rejected.Receipt), `"detail":"plant_not_mature"`) {
		t.Fatalf("immature harvest receipt=%s", rejected.Receipt)
	}
	if _, company := fixture.revisions(t); company != 1 || fixture.quota(t) != 0 {
		t.Fatalf("a rejection touched the Company (%d) or the window (%d)", company, fixture.quota(t))
	}
	first := fixture.harvest(t, 2, 1, `{"row":0,"col":0}`)
	var receipt gardenHarvestAPIReceipt
	if err := json.Unmarshal(first.Receipt, &receipt); err != nil || receipt.Outcome != "applied" || receipt.Harvest.TotalUnits != 20 ||
		receipt.Credited != "2e1" || !receipt.FaucetApplied || fmt.Sprint(receipt.Harvest.SeedsDiscovered) != "[strain_c]" {
		t.Fatalf("first harvest receipt=%s err=%v", first.Receipt, err)
	}
	if retry := fixture.harvest(t, 2, 1, `{"row":0,"col":0}`); !retry.Replay || string(retry.Receipt) != string(first.Receipt) {
		t.Fatalf("retry=%s replay=%v", retry.Receipt, retry.Replay)
	}
	if founder, company := fixture.revisions(t); founder != 2 || company != 2 || fixture.quota(t) != 1 {
		t.Fatalf("after one credited harvest (and a retry) revisions=%d/%d quota=%d", founder, company, fixture.quota(t))
	}
	zero := fixture.harvest(t, 3, 2, `{"row":2,"col":1}`)
	if err := json.Unmarshal(zero.Receipt, &receipt); err != nil || receipt.FaucetApplied || receipt.Credited != "0" || receipt.Harvest.TotalUnits != 0 {
		t.Fatalf("zero harvest receipt=%s", zero.Receipt)
	}
	if fixture.quota(t) != 1 {
		t.Fatalf("a zero harvest consumed a daily send: quota=%d", fixture.quota(t))
	}
	expected := 3
	for index := 1; index < 8; index++ {
		fixture.harvest(t, 3+index, expected, fmt.Sprintf(`{"row":%d,"col":%d}`, index%2, index/2))
		expected++
	}
	if fixture.quota(t) != 8 {
		t.Fatalf("eight credited harvests quota=%d", fixture.quota(t))
	}
	late := fixture.harvest(t, 20, expected, `{"row":2,"col":0}`)
	if err := json.Unmarshal(late.Receipt, &receipt); err != nil || receipt.Credited != "0" || receipt.ForfeitedUnits != 40 ||
		receipt.CapReasonKey == nil || *receipt.CapReasonKey != "cap.minigame_faucet" || fmt.Sprint(receipt.Harvest.SeedsDiscovered) != "[strain_d]" {
		t.Fatalf("past-quota harvest must forfeit visibly and still record seeds: %s", late.Receipt)
	}
	loaded, err := fixture.store.LoadLatest(fixture.ctx, fixture.companyStreamID)
	if err != nil {
		t.Fatal(err)
	}
	cash, _ := loaded.State.Ledger.Balance("company.cash")
	if !cash.Eq(decimal.FromString("1.6e2")) {
		t.Fatalf("company.cash=%s, want exactly eight credits of 20", cash)
	}
	founder, err := fixture.store.LoadLatest(fixture.ctx, fixture.founderStreamID)
	if err != nil || !founder.State.ServerGarden.Collected("strain_d") || !founder.State.ServerGarden.Collected("strain_c") {
		t.Fatalf("seed collection after harvests: %+v err=%v", founder.State.ServerGarden.SeedCollection, err)
	}
	// AC9: both logs bind the same harvest_hash for the first credited harvest.
	var founderHash, companyHash string
	if err := fixture.db.QueryRowContext(fixture.ctx, `SELECT payload->>'harvest_hash' FROM events WHERE stream_id=$1 AND kind='garden_harvested.v1' ORDER BY revision LIMIT 1`, fixture.founderStreamID).Scan(&founderHash); err != nil {
		t.Fatal(err)
	}
	if err := fixture.db.QueryRowContext(fixture.ctx, `SELECT convert_from(canonical_payload,'UTF8')::jsonb->>'harvest_hash' FROM run_log WHERE company_stream_id=$1 ORDER BY seq LIMIT 1`, fixture.companyStreamID).Scan(&companyHash); err != nil {
		t.Fatal(err)
	}
	if founderHash == "" || founderHash != companyHash {
		t.Fatalf("harvest_hash binding founder=%q company=%q", founderHash, companyHash)
	}
	history, err := fixture.store.LoadFounderHistory(fixture.ctx, fixture.founderStreamID)
	if err != nil {
		t.Fatal(err)
	}
	if verdict := VerifyFounderHistory(history, ReplayCatalogSet{fixture.bundle.ConstantsHash: fixture.bundle}); verdict != ReplayVerified {
		t.Fatalf("Founder history verdict=%v", verdict)
	}
	replayGardenRunLog(t, fixture)
}

// replayGardenRunLog replays every Company run-log entry from genesis and
// requires identical receipts and Company events (the run is not terminal,
// so VerifyReplayRun's terminal requirement does not apply).
func replayGardenRunLog(t *testing.T, fixture gardenHarvestFixture) {
	t.Helper()
	var genesis []byte
	var version int
	if err := fixture.db.QueryRowContext(fixture.ctx, `SELECT state,version FROM run_genesis WHERE company_stream_id=$1 AND run_seq=1`, fixture.companyStreamID).Scan(&genesis, &version); err != nil {
		t.Fatal(err)
	}
	state, err := save.RestoreState(genesis, version, fixture.bundle.Economy, economy.ScopeCompany, time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	rows, err := fixture.db.QueryContext(fixture.ctx, `SELECT intent_id,canonical_payload,replay_inputs,receipt FROM run_log WHERE company_stream_id=$1 AND run_seq=1 ORDER BY seq`, fixture.companyStreamID)
	if err != nil {
		t.Fatal(err)
	}
	type entry struct {
		intent                   string
		payload, inputs, receipt []byte
	}
	entries := []entry{}
	for rows.Next() {
		var value entry
		if err := rows.Scan(&value.intent, &value.payload, &value.inputs, &value.receipt); err != nil {
			t.Fatal(err)
		}
		entries = append(entries, value)
	}
	rows.Close()
	if len(entries) != 10 {
		t.Fatalf("run log holds %d garden credits, want 10", len(entries))
	}
	for _, value := range entries {
		transition, err := ApplyLogged(state, value.payload, fixture.bundle, value.inputs)
		if err != nil || !canonicalJSONEqual(transition.Receipt, value.receipt) {
			t.Fatalf("run-log entry %s diverged: %v", value.intent, err)
		}
		var stored []byte
		if err := fixture.db.QueryRowContext(fixture.ctx, `SELECT COALESCE(json_agg(json_build_object('kind',kind,'schema_version',schema_version,'intent_id',intent_id,'payload',payload) ORDER BY event_seq),'[]') FROM events WHERE stream_id=$1 AND intent_id=$2`, fixture.companyStreamID, value.intent).Scan(&stored); err != nil {
			t.Fatal(err)
		}
		if !canonicalJSONEqual(marshalReplayEvents(transition.Events), stored) {
			t.Fatalf("run-log entry %s Company events diverged: %s vs %s", value.intent, marshalReplayEvents(transition.Events), stored)
		}
		state = transition.State
	}
}

// TestGardenHarvestFaultsAreAllOrNothing is AC8's fault half: a fault injected
// after every write of the coordinator leaves no trace on either stream, the
// window, the logs, the events, or the intent record.
func TestGardenHarvestFaultsAreAllOrNothing(t *testing.T) {
	fixture := newGardenHarvestFixture(t)
	points := []string{"faucet_window", "company_revision", "company_events", "run_log", "founder_revision", "founder_events", "founder_log", "intent_record", "retention"}
	founderLoaded, err := fixture.store.LoadLatest(fixture.ctx, fixture.founderStreamID)
	if err != nil {
		t.Fatal(err)
	}
	count := func(query string, args ...any) int {
		var value int
		if err := fixture.db.QueryRowContext(fixture.ctx, query, args...).Scan(&value); err != nil {
			t.Fatal(err)
		}
		return value
	}
	for index, point := range points {
		request, err := ParseIntent([]byte(fmt.Sprintf(`{"intent_id":"01986666-7f11-7000-8000-%012d","kind":"garden_harvest","expected_revision":1,"plots":[{"row":0,"col":0}]}`, index+1)))
		if err != nil {
			t.Fatal(err)
		}
		now := time.Now()
		attendance, err := fixture.service.ResolveFounderAttendance(fixture.ctx, fixture.founderStreamID, fixture.companyStreamID, now)
		if err != nil {
			t.Fatal(err)
		}
		injected := errors.New("injected " + point)
		_, err = fixture.service.creditGardenHarvest(fixture.ctx, fixture.companyStreamID, founderLoaded, fixture.bundle, request,
			save.CanonicalServerTime(now).UnixMilli(), attendance, func(name string) error {
				if name == point {
					return injected
				}
				return nil
			})
		if !errors.Is(err, injected) {
			t.Fatalf("%s: fault not raised: %v", point, err)
		}
		founder, company := fixture.revisions(t)
		if founder != 1 || company != 1 || fixture.quota(t) != 0 ||
			count(`SELECT count(*) FROM run_log WHERE company_stream_id=$1`, fixture.companyStreamID) != 0 ||
			count(`SELECT count(*) FROM events WHERE intent_id=$1`, request.IntentID) != 0 ||
			count(`SELECT count(*) FROM intent_records WHERE intent_id=$1`, request.IntentID) != 0 {
			t.Fatalf("%s left partial rows: revisions=%d/%d quota=%d", point, founder, company, fixture.quota(t))
		}
	}
	if result := fixture.harvest(t, 99, 1, `{"row":0,"col":0}`); !strings.Contains(string(result.Receipt), `"outcome":"applied"`) {
		t.Fatalf("clean harvest after faults: %s", result.Receipt)
	}
}

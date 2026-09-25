package production

import (
	"context"
	"encoding/json"
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

// TestPetAdoptionIntegrationPersistsReplayableFounderLog covers the Postgres
// halves of AC9 (event registry + constraint) and AC10 (idempotency: one nonce
// per adoption, byte-identical retry, idempotency_conflict on a changed body),
// through Service.Handle and ApplyFounderLogged.
func TestPetAdoptionIntegrationPersistsReplayableFounderLog(t *testing.T) {
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
	bundle := petSpeciesContentBundle(t)
	seedProductionEpoch(t, db, bundle.ConstantsHash, bundle.Artifacts)
	resolver := integrationCatalogs{economy: map[string]*economy.Catalog{bundle.ConstantsHash: bundle.Economy},
		routes: map[string]*routes.Catalog{bundle.ConstantsHash: bundle.Routes}, prestige: map[string]*prestigecore.Policy{bundle.ConstantsHash: bundle.Prestige},
		factions: map[string]*faction.Catalog{bundle.ConstantsHash: bundle.Faction}}
	store, err := save.NewStore(db, resolver, nil)
	if err != nil {
		t.Fatal(err)
	}
	cursor := save.CanonicalServerTime(time.Date(2026, 9, 25, 12, 0, 0, 0, time.UTC))
	const accountID = "01986666-6f00-4000-8000-000000000001"
	const founderID = "01986666-6f00-7000-8000-000000000002"
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
		bundle.ConstantsHash, company, save.WriteContext{Cause: "pet.adoption.integration"})
	if err != nil {
		t.Fatal(err)
	}
	founder := petFounderState(t, bundle, 23, cursor, 0)
	founderRevision, err := store.CreateStream(ctx, save.StreamKey{OwnerKind: save.OwnerFounder, OwnerID: founderID, Scope: economy.ScopeFounder},
		bundle.ConstantsHash, founder, save.WriteContext{Cause: "pet.adoption.integration"})
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
	draws := 0
	service, err := NewService(store, resolver, nil, nil, nil, WithProgressionRuntime(resolver), WithCurrentConstantsHash(bundle.ConstantsHash),
		WithReplayCatalogs(ReplayCatalogSet{bundle.ConstantsHash: bundle}), WithGuildSettlements(emptyGuildSettlements{}),
		WithAdoptionNonceSource(func() (string, error) { draws++; return nonceA, nil }))
	if err != nil {
		t.Fatal(err)
	}
	body := []byte(`{"intent_id":"01986666-6f00-7000-8000-000000000003","kind":"adopt_pet","expected_revision":1,"species_id":"pet_species.server_room_cat","name_key":"pet.name.server_room_cat.n02"}`)
	result, err := service.Handle(ctx, companyRevision.StreamID, ModeOnline, cursor.Add(time.Second), body)
	if err != nil || result.Replay || !strings.Contains(string(result.Receipt), `"outcome":"applied"`) {
		t.Fatalf("adoption result=%s replay=%v err=%v", result.Receipt, result.Replay, err)
	}
	retry, err := service.Handle(ctx, companyRevision.StreamID, ModeOnline, cursor.Add(2*time.Second), body)
	if err != nil || !retry.Replay || string(retry.Receipt) != string(result.Receipt) || draws != 1 {
		t.Fatalf("retry=%s replay=%v draws=%d err=%v", retry.Receipt, retry.Replay, draws, err)
	}
	changed := []byte(strings.Replace(string(body), "n02", "n03", 1))
	conflict, err := service.Handle(ctx, companyRevision.StreamID, ModeOnline, cursor.Add(2*time.Second), changed)
	if err != nil || !strings.Contains(string(conflict.Receipt), `"category":"idempotency_conflict"`) || draws != 1 {
		t.Fatalf("changed payload receipt=%s draws=%d err=%v", conflict.Receipt, draws, err)
	}
	var receipt founderAdoptionReceipt
	if err := json.Unmarshal(result.Receipt, &receipt); err != nil {
		t.Fatal(err)
	}
	loaded, err := store.LoadLatest(ctx, founderRevision.StreamID)
	if err != nil || loaded.Revision.Number != 2 || save.VersionForState(loaded.State) != 23 || loaded.State.PetIdentities[receipt.PetID].NameKey != "pet.name.server_room_cat.n02" {
		t.Fatalf("Founder after adoption revision=%d identities=%+v err=%v", loaded.Revision.Number, loaded.State.PetIdentities, err)
	}
	var logKind string
	if err := db.QueryRowContext(ctx, `SELECT replay_inputs->'resolved'->>'kind' FROM founder_log WHERE founder_stream_id=$1 AND intent_id=$2`,
		founderRevision.StreamID, "01986666-6f00-7000-8000-000000000003").Scan(&logKind); err != nil || logKind != IntentAdoptPet {
		t.Fatalf("adoption log kind=%q err=%v", logKind, err)
	}
	var eventCount int
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM events WHERE stream_id=$1 AND kind='pet_adopted.v1'`, founderRevision.StreamID).Scan(&eventCount); err != nil || eventCount != 1 {
		t.Fatalf("pet_adopted events=%d err=%v", eventCount, err)
	}
	// AC9 failing case: the events constraint rejects an unregistered kind.
	if _, err := db.ExecContext(ctx, `UPDATE events SET kind='pet_released.v1' WHERE stream_id=$1 AND kind='pet_adopted.v1'`, founderRevision.StreamID); err == nil {
		t.Fatal("events constraint accepted an unregistered kind")
	}
	// AC4 at the service boundary: a second adoption at cap 1 rejects.
	second := []byte(`{"intent_id":"01986666-6f00-7000-8000-000000000004","kind":"adopt_pet","expected_revision":2,"species_id":"pet_species.server_room_cat","name_key":"pet.name.server_room_cat.n03"}`)
	capped, err := service.Handle(ctx, companyRevision.StreamID, ModeOnline, cursor.Add(3*time.Second), second)
	if err != nil || !strings.Contains(string(capped.Receipt), `"adoption_cap_reached"`) || draws != 2 {
		t.Fatalf("second adoption receipt=%s draws=%d err=%v", capped.Receipt, draws, err)
	}
	history, err := store.LoadFounderHistory(ctx, founderRevision.StreamID)
	if err != nil {
		t.Fatal(err)
	}
	if verdict := VerifyFounderHistory(history, ReplayCatalogSet{bundle.ConstantsHash: bundle}); verdict != ReplayVerified {
		t.Fatalf("Founder history verdict=%v", verdict)
	}
}

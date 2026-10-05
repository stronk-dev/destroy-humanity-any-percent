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
	"cloud-clicker/server/pet"
	prestigecore "cloud-clicker/server/prestige"
	"cloud-clicker/server/routes"
	"cloud-clicker/server/save"
	"cloud-clicker/server/soul"
)

// TestCosmeticIntegrationPersistsReplayableFounderLog covers AC9 (the three
// event kinds admitted, atomic state/event/receipt, an unregistered kind
// rejected by the database) and AC5's service-boundary rows (idempotent retry,
// idempotency_conflict, owned, invalid price field), through Service.Handle.
// All three cosmetic intents stay usable during real Soul recovery (§4.5),
// which then resolves without changing cosmetic ownership.
func TestCosmeticIntegrationPersistsReplayableFounderLog(t *testing.T) {
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
	bundle := cosmeticsContentBundle(t)
	seedProductionEpoch(t, db, bundle.ConstantsHash, bundle.Artifacts)
	resolver := integrationCatalogs{economy: map[string]*economy.Catalog{bundle.ConstantsHash: bundle.Economy},
		routes: map[string]*routes.Catalog{bundle.ConstantsHash: bundle.Routes}, prestige: map[string]*prestigecore.Policy{bundle.ConstantsHash: bundle.Prestige},
		factions: map[string]*faction.Catalog{bundle.ConstantsHash: bundle.Faction}}
	store, err := save.NewStore(db, resolver, nil)
	if err != nil {
		t.Fatal(err)
	}
	// Founder ApplyLogged commands use the database clock. Anchor the fixture
	// there too, so a later recovery terminal never moves Fiscal time backwards.
	var databaseNow time.Time
	if err := db.QueryRowContext(ctx, `SELECT clock_timestamp()`).Scan(&databaseNow); err != nil {
		t.Fatal(err)
	}
	cursor := save.CanonicalServerTime(databaseNow)
	const accountID = "01986666-7e00-4000-8000-000000000001"
	const founderID = "01986666-7e00-7000-8000-000000000002"
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
	// §4.2 step 5: the active Company tier the server freezes is 1 (Horse Armor unlocks at 1).
	company.Tier = 1
	company.GatesCrossed["gate.t0_to_t1"] = true
	companyRevision, err := store.CreateStream(ctx, save.StreamKey{OwnerKind: save.OwnerFounder, OwnerID: founderID, Scope: economy.ScopeCompany},
		bundle.ConstantsHash, company, save.WriteContext{Cause: "cosmetic.integration"})
	if err != nil {
		t.Fatal(err)
	}
	founder := reputationFounderState(t, bundle, 24, cursor, 0)
	care, err := pet.InitialCareState(bundle.Pets, founder.AgeMS)
	if err != nil {
		t.Fatal(err)
	}
	identity := adoptedIdentity()
	identity.AdoptedAtMS, identity.AdoptedAtAttendedMS = cursor.UnixMilli(), founder.AgeMS
	founder.Pets[adoptedPetID], founder.PetIdentities[adoptedPetID] = care, identity
	founderRevision, err := store.CreateStream(ctx, save.StreamKey{OwnerKind: save.OwnerFounder, OwnerID: founderID, Scope: economy.ScopeFounder},
		bundle.ConstantsHash, founder, save.WriteContext{Cause: "cosmetic.integration"})
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
	recoveries, err := soul.NewRecoveryRepository(db)
	if err != nil {
		t.Fatal(err)
	}
	service, err := NewService(store, resolver, nil, nil, nil, WithProgressionRuntime(resolver), WithCurrentConstantsHash(bundle.ConstantsHash),
		WithReplayCatalogs(ReplayCatalogSet{bundle.ConstantsHash: bundle}), WithGuildSettlements(emptyGuildSettlements{}), WithSoulRecovery(recoveries))
	if err != nil {
		t.Fatal(err)
	}
	const recoveryID = "01986666-7e00-7000-8000-000000000010"
	started, err := service.StartSoulRecovery(ctx, StartSoulRecoveryRequest{SessionID: recoveryID,
		FounderID: founderID, CompanyStreamID: companyRevision.StreamID, ActivityID: "defrag"}, cursor)
	if err != nil {
		t.Fatalf("start real Soul recovery for cosmetic §4.5: %v", err)
	}
	var recoveryStart soulRecoveryStartReceipt
	if err := json.Unmarshal(started.Receipt, &recoveryStart); err != nil || recoveryStart.RequiredDurationAttendedMS <= 0 || recoveryStart.ProgressToken == "" {
		t.Fatalf("Soul recovery start receipt=%s err=%v", started.Receipt, err)
	}
	ordinary := []byte(`{"intent_id":"01986666-7e00-7000-8000-000000000011","kind":"perform_manual_batch","expected_revision":1,"action_id":"manual.click","count":1,"window_ms":1}`)
	blocked, err := service.Handle(ctx, companyRevision.StreamID, ModeOnline, cursor.Add(time.Second), ordinary)
	if err != nil || !strings.Contains(string(blocked.Receipt), `"detail":"exclusive_activity"`) {
		t.Fatalf("ordinary intent during Soul recovery receipt=%s err=%v", blocked.Receipt, err)
	}
	companyBefore, err := store.LoadLatest(ctx, companyRevision.StreamID)
	if err != nil {
		t.Fatal(err)
	}
	companyBytes, err := save.EncodeState(companyBefore.State)
	if err != nil {
		t.Fatal(err)
	}
	body := []byte(`{"intent_id":"01986666-7e00-7000-8000-000000000003","kind":"acquire_cosmetic","expected_revision":1,"cosmetic_id":"horse_armor"}`)
	result, err := service.Handle(ctx, companyRevision.StreamID, ModeOnline, cursor.Add(time.Second), body)
	if err != nil || result.Replay || !strings.Contains(string(result.Receipt), `"outcome":"applied"`) || !strings.Contains(string(result.Receipt), `"order_number":1`) {
		t.Fatalf("acquire result=%s replay=%v err=%v", result.Receipt, result.Replay, err)
	}
	for _, forbidden := range []string{"price", "amount", "currency", "payment"} {
		if strings.Contains(string(result.Receipt), forbidden) {
			t.Fatalf("receipt carries %q: %s", forbidden, result.Receipt)
		}
	}
	retry, err := service.Handle(ctx, companyRevision.StreamID, ModeOnline, cursor.Add(2*time.Second), body)
	if err != nil || !retry.Replay || string(retry.Receipt) != string(result.Receipt) {
		t.Fatalf("retry=%s replay=%v err=%v", retry.Receipt, retry.Replay, err)
	}
	changed := []byte(strings.Replace(string(body), "horse_armor", "zebra_armor", 1))
	conflict, err := service.Handle(ctx, companyRevision.StreamID, ModeOnline, cursor.Add(2*time.Second), changed)
	if err != nil || !strings.Contains(string(conflict.Receipt), `"category":"idempotency_conflict"`) {
		t.Fatalf("changed payload receipt=%s err=%v", conflict.Receipt, err)
	}
	loaded, err := store.LoadLatest(ctx, founderRevision.StreamID)
	if err != nil || loaded.Revision.Number != 2 || save.VersionForState(loaded.State) != 24 || !loaded.State.Cosmetics.Owns("horse_armor") {
		t.Fatalf("Founder after acquire revision=%d cosmetics=%+v err=%v", loaded.Revision.Number, loaded.State.Cosmetics, err)
	}
	// I2: the Company stream is untouched by a Founder cosmetic intent.
	companyAfter, err := store.LoadLatest(ctx, companyRevision.StreamID)
	if err != nil || companyAfter.Revision.Number != companyBefore.Revision.Number {
		t.Fatalf("Company revision moved: %d -> %d err=%v", companyBefore.Revision.Number, companyAfter.Revision.Number, err)
	}
	if after, _ := save.EncodeState(companyAfter.State); string(after) != string(companyBytes) {
		t.Fatal("Company state bytes changed")
	}
	for _, row := range []struct{ intent, body, want string }{
		{"01986666-7e00-7000-8000-000000000004", `"cosmetic_id":"horse_armor"`, `"detail":"owned"`},
		{"01986666-7e00-7000-8000-000000000005", `"cosmetic_id":"horse_armor","price":0`, `"category":"invalid"`},
	} {
		rejected, err := service.Handle(ctx, companyRevision.StreamID, ModeOnline, cursor.Add(3*time.Second),
			[]byte(`{"intent_id":"`+row.intent+`","kind":"acquire_cosmetic","expected_revision":2,`+row.body+`}`))
		if err != nil || !strings.Contains(string(rejected.Receipt), row.want) {
			t.Fatalf("rejection %s receipt=%s err=%v", row.want, rejected.Receipt, err)
		}
	}
	var eventCount int
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM events WHERE stream_id=$1 AND kind LIKE 'cosmetic_%'`, founderRevision.StreamID).Scan(&eventCount); err != nil || eventCount != 1 {
		t.Fatalf("cosmetic events=%d err=%v (rejections must store no event)", eventCount, err)
	}
	var payload string
	if err := db.QueryRowContext(ctx, `SELECT payload::text FROM events WHERE stream_id=$1 AND kind='cosmetic_acquired.v1'`, founderRevision.StreamID).Scan(&payload); err != nil ||
		!strings.Contains(payload, `"order_number": 1`) {
		t.Fatalf("cosmetic_acquired payload=%s err=%v", payload, err)
	}
	var eventID string
	if err := db.QueryRowContext(ctx, `SELECT event_id FROM events WHERE stream_id=$1 AND kind='cosmetic_acquired.v1'`, founderRevision.StreamID).Scan(&eventID); err != nil {
		t.Fatal(err)
	}
	// AC9 requires the database, not only the Go event decoder, to reject an
	// extra cosmetic payload key. Roll back every probe so the real history
	// remains byte-identical for the verification below.
	for _, row := range []struct{ name, kind, valid, extra, missing string }{
		{"acquired", "cosmetic_acquired.v1", `{"cosmetic_id":"horse_armor","order_number":1}`, `{"cosmetic_id":"horse_armor","order_number":1,"price":0}`, `{"cosmetic_id":"horse_armor"}`},
		{"equipped", "cosmetic_equipped.v1", `{"cosmetic_id":"horse_armor","pet_id":"01986666-aaaa-7aaa-8aaa-aaaaaaaaaaaa","replaced_cosmetic_id":null}`, `{"cosmetic_id":"horse_armor","pet_id":"01986666-aaaa-7aaa-8aaa-aaaaaaaaaaaa","replaced_cosmetic_id":null,"price":0}`, `{"cosmetic_id":"horse_armor","pet_id":"01986666-aaaa-7aaa-8aaa-aaaaaaaaaaaa"}`},
		{"unequipped", "cosmetic_unequipped.v1", `{"cosmetic_id":"horse_armor","pet_id":"01986666-aaaa-7aaa-8aaa-aaaaaaaaaaaa"}`, `{"cosmetic_id":"horse_armor","pet_id":"01986666-aaaa-7aaa-8aaa-aaaaaaaaaaaa","amount":0}`, `{"cosmetic_id":"horse_armor"}`},
	} {
		t.Run("database-payload-keys-"+row.name, func(t *testing.T) {
			tx, err := db.BeginTx(ctx, nil)
			if err != nil {
				t.Fatal(err)
			}
			defer tx.Rollback()
			result, err := tx.ExecContext(ctx, `UPDATE events SET kind=$1,payload=$2::jsonb WHERE event_id=$3`, row.kind, row.valid, eventID)
			if err != nil {
				t.Fatalf("valid %s payload rejected by database: %v", row.name, err)
			}
			if affected, err := result.RowsAffected(); err != nil || affected != 1 {
				t.Fatalf("valid %s payload updated %d events: %v", row.name, affected, err)
			}
			if _, err := tx.ExecContext(ctx, `UPDATE events SET kind=$1,payload=$2::jsonb WHERE event_id=$3`, row.kind, row.extra, eventID); err == nil {
				t.Fatalf("database accepted an extra field on %s payload", row.name)
			}
		})
		t.Run("database-payload-required-"+row.name, func(t *testing.T) {
			tx, err := db.BeginTx(ctx, nil)
			if err != nil {
				t.Fatal(err)
			}
			defer tx.Rollback()
			if _, err := tx.ExecContext(ctx, `UPDATE events SET kind=$1,payload=$2::jsonb WHERE event_id=$3`, row.kind, row.missing, eventID); err == nil {
				t.Fatalf("database accepted a missing field on %s payload", row.name)
			}
		})
	}
	// This new constraint must not narrow the payload grammar of other events.
	tx, err = db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `UPDATE events SET kind='generator_purchased',payload='{"unrelated":true}'::jsonb WHERE event_id=$1`, eventID); err != nil {
		t.Fatalf("non-cosmetic event payload narrowed: %v", err)
	}
	if err := tx.Rollback(); err != nil {
		t.Fatal(err)
	}
	// AC9 failing case: the events constraint rejects an unregistered kind.
	if _, err := db.ExecContext(ctx, `UPDATE events SET kind='cosmetic_purchased.v1' WHERE stream_id=$1 AND kind='cosmetic_acquired.v1'`, founderRevision.StreamID); err == nil {
		t.Fatal("events constraint accepted cosmetic_purchased.v1")
	}
	history, err := store.LoadFounderHistory(ctx, founderRevision.StreamID)
	if err != nil {
		t.Fatal(err)
	}
	if verdict := VerifyFounderHistory(history, ReplayCatalogSet{bundle.ConstantsHash: bundle}); verdict != ReplayVerified {
		t.Fatalf("Founder history verdict=%v", verdict)
	}
	for _, row := range []struct {
		body, event string
	}{
		{`{"intent_id":"01986666-7e00-7000-8000-000000000012","kind":"equip_cosmetic","expected_revision":2,"cosmetic_id":"horse_armor","pet_id":"` + adoptedPetID + `"}`, `"kind":"cosmetic_equipped.v1"`},
		{`{"intent_id":"01986666-7e00-7000-8000-000000000013","kind":"unequip_cosmetic","expected_revision":3,"pet_id":"` + adoptedPetID + `"}`, `"kind":"cosmetic_unequipped.v1"`},
	} {
		applied, err := service.Handle(ctx, companyRevision.StreamID, ModeOnline, cursor.Add(4*time.Second), []byte(row.body))
		if err != nil || !strings.Contains(string(applied.Receipt), `"outcome":"applied"`) || !strings.Contains(string(applied.Receipt), row.event) {
			t.Fatalf("cosmetic equip/unequip during Soul recovery receipt=%s err=%v", applied.Receipt, err)
		}
	}
	// Advancing the Founder for a cosmetic intent must not strand the active
	// recovery coordinator or erase cosmetics when recovery resolves. Each beat
	// uses the pinned policy and server-controlled test clock, not client time.
	recoveryNow := cursor
	remaining := recoveryStart.RequiredDurationAttendedMS
	for remaining > 0 {
		step := min(remaining, bundle.Soul.Policy.RecoveryBeatCeilingMS)
		if step <= 0 {
			t.Fatal("Soul recovery beat ceiling must be positive")
		}
		recoveryNow = recoveryNow.Add(time.Duration(step) * time.Millisecond)
		if _, err := service.ProgressSoulRecovery(ctx, ProgressSoulRecoveryRequest{SessionID: recoveryID, FounderID: founderID,
			ProgressToken: recoveryStart.ProgressToken}, recoveryNow, nil); err != nil {
			t.Fatalf("Soul recovery progress after cosmetic acquisition: %v", err)
		}
		remaining -= step
	}
	resolved, err := service.ResolveSoulRecovery(ctx, FinishSoulRecoveryRequest{SessionID: recoveryID, FounderID: founderID}, recoveryNow, nil)
	if err != nil || !strings.Contains(string(resolved.Receipt), `"outcome":"applied"`) {
		t.Fatalf("Soul recovery resolution after cosmetic acquisition receipt=%s err=%v", resolved.Receipt, err)
	}
	terminal, err := recoveries.Load(ctx, founderID, recoveryID)
	if err != nil || terminal.Status != soul.RecoveryResolved {
		t.Fatalf("Soul recovery did not persist its terminal state: %v", err)
	}
	afterRecovery, err := store.LoadLatest(ctx, founderRevision.StreamID)
	if err != nil || afterRecovery.Revision.Number != 5 || !loaded.State.Cosmetics.Equal(afterRecovery.State.Cosmetics) {
		t.Fatalf("Soul recovery changed cosmetic state or failed to advance Founder: %v", err)
	}
	history, err = store.LoadFounderHistory(ctx, founderRevision.StreamID)
	if err != nil {
		t.Fatal(err)
	}
	if verdict := VerifyFounderHistory(history, ReplayCatalogSet{bundle.ConstantsHash: bundle}); verdict != ReplayVerified {
		t.Fatalf("Founder history after cosmetic acquisition and Soul recovery verdict=%v", verdict)
	}
}

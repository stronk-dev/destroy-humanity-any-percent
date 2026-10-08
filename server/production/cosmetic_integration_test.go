package production

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math/big"
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
	bundle := twoItemCosmeticsBundle(t)
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
	exerciseCosmeticCommandPersistence(t, ctx, db, store, service, bundle, companyRevision.StreamID, founderRevision.StreamID, cursor)
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
	if err != nil || afterRecovery.Revision.Number != 7 || !loaded.State.Cosmetics.Equal(afterRecovery.State.Cosmetics) {
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

// Exercise the real transaction before the simulated recovery clock advances.
// The second fixture item makes not_owned reachable; ownership and equip state
// still come exclusively from Service.Handle.
func exerciseCosmeticCommandPersistence(t *testing.T, ctx context.Context, db *sql.DB, store *save.Store,
	service *Service, bundle CatalogBundle, companyStreamID, founderStreamID string, now time.Time) {
	t.Helper()
	type observation struct {
		founder, company string
		counts           [6]int64
	}
	observe := func(t *testing.T) observation {
		t.Helper()
		var observed observation
		for index, stream := range []string{founderStreamID, companyStreamID} {
			loaded, err := store.LoadLatest(ctx, stream)
			if err != nil {
				t.Fatal(err)
			}
			value := fmt.Sprintf("%d/%s/%s", loaded.Revision.Number, loaded.Revision.ConstantsHash, mustEncodeState(t, loaded.State))
			if index == 0 {
				observed.founder = value
			} else {
				observed.company = value
			}
		}
		if err := db.QueryRowContext(ctx, `SELECT
			(SELECT count(*) FROM save_revisions WHERE stream_id IN ($1,$2)),
			(SELECT count(*) FROM events WHERE stream_id IN ($1,$2)),
			(SELECT count(*) FROM intent_records WHERE stream_id IN ($1,$2)),
			(SELECT count(*) FROM founder_log WHERE founder_stream_id=$1),
			(SELECT count(*) FROM transport_player_outbox WHERE stream_id IN ($1,$2)),
			(SELECT count(*) FROM founder_genesis WHERE founder_stream_id=$1)`, founderStreamID, companyStreamID).
			Scan(&observed.counts[0], &observed.counts[1], &observed.counts[2], &observed.counts[3], &observed.counts[4], &observed.counts[5]); err != nil {
			t.Fatal(err)
		}
		return observed
	}
	assertRows := func(t *testing.T, intentID string, wantCommand, wantEvent int64) {
		t.Helper()
		var commands, logs, receipts, events, eventDeliveries int64
		if err := db.QueryRowContext(ctx, `SELECT
			(SELECT count(*) FROM intent_records WHERE stream_id=$1 AND intent_id=$2),
			(SELECT count(*) FROM founder_log WHERE founder_stream_id=$1 AND intent_id=$2),
			(SELECT count(*) FROM transport_player_outbox WHERE stream_id=$1 AND message_kind='receipt' AND source_id=$2),
			(SELECT count(*) FROM events WHERE stream_id=$1 AND intent_id=$2),
			(SELECT count(*) FROM transport_player_outbox o JOIN events e ON e.stream_id=o.stream_id AND e.event_id=o.source_id
			 WHERE o.stream_id=$1 AND o.message_kind='event' AND e.intent_id=$2)`, founderStreamID, intentID).
			Scan(&commands, &logs, &receipts, &events, &eventDeliveries); err != nil {
			t.Fatal(err)
		}
		if commands != wantCommand || logs != wantCommand || receipts != wantCommand || events != wantEvent || eventDeliveries != wantEvent {
			t.Fatalf("%s persisted commands/logs/receipts/events/event-deliveries = %d/%d/%d/%d/%d, want %d/%d/%d/%d/%d",
				intentID, commands, logs, receipts, events, eventDeliveries, wantCommand, wantCommand, wantCommand, wantEvent, wantEvent)
		}
	}
	expectedAppliedReceipt := func(t *testing.T, intentID string, revision int64, cosmeticJSON string) string {
		t.Helper()
		// A real wall-clock command can cross a Fiscal period. Derive that
		// permitted shared prelude from the previous SQL state, pinned policy
		// and recorded evaluation time; do not discard an unexpected field or
		// make the fixture clock slower just to keep the literal receipt green.
		var credit, opened, sequence, evaluated int64
		if err := db.QueryRowContext(ctx, `SELECT (r.state->>'fiscal_credit')::bigint,
			(r.state->>'fiscal_period_opened_wall_ms')::bigint,(r.state->>'fiscal_period_seq')::bigint,l.server_ts_ms
			FROM save_revisions r JOIN founder_log l ON l.founder_stream_id=r.stream_id
			WHERE r.stream_id=$1 AND r.revision=$2 AND l.intent_id=$3`, founderStreamID, revision-1, intentID).
			Scan(&credit, &opened, &sequence, &evaluated); err != nil {
			t.Fatal(err)
		}
		if evaluated < opened {
			t.Fatal("stored cosmetic command moves Fiscal time backwards")
		}
		periods := (evaluated - opened) / bundle.Fiscal.Clock.AutoMS
		minted := new(big.Int).Mul(big.NewInt(periods), big.NewInt(bundle.Fiscal.Credit.CreditPerPeriod))
		headroom := big.NewInt(bundle.Fiscal.Credit.Hardcap - credit)
		saturated := minted.Cmp(headroom) > 0
		credited := new(big.Int).Set(minted)
		if saturated {
			credited.Set(headroom)
		}
		var fields map[string]json.RawMessage
		if err := json.Unmarshal([]byte(cosmeticJSON), &fields); err != nil {
			t.Fatal(err)
		}
		if periods > 0 {
			sweep := map[string]any{"periods": periods, "credit_before": credit, "credited": credited.Int64(),
				"credit_after": credit + credited.Int64(), "opened_before_ms": opened,
				"opened_after_ms": opened + periods*bundle.Fiscal.Clock.AutoMS, "seq_before": sequence,
				"seq_after": sequence + periods, "saturated": saturated, "hardcap_reason_key": bundle.Fiscal.Credit.HardcapReasonKey}
			encoded, err := json.Marshal(sweep)
			if err != nil {
				t.Fatal(err)
			}
			fields["fiscal_sweep"] = encoded
		}
		var afterCredit, afterOpened, afterSequence int64
		if err := db.QueryRowContext(ctx, `SELECT (state->>'fiscal_credit')::bigint,
			(state->>'fiscal_period_opened_wall_ms')::bigint,(state->>'fiscal_period_seq')::bigint
			FROM save_revisions WHERE stream_id=$1 AND revision=$2`, founderStreamID, revision).
			Scan(&afterCredit, &afterOpened, &afterSequence); err != nil {
			t.Fatal(err)
		}
		if afterCredit != credit+credited.Int64() || afterOpened != opened+periods*bundle.Fiscal.Clock.AutoMS || afterSequence != sequence+periods {
			t.Fatal("applied cosmetic command persisted the wrong Fiscal prelude")
		}
		return canonicalFixtureValue(t, fields)
	}
	assertStored := func(t *testing.T, intentID string, receipt json.RawMessage, revision int64, applied bool) {
		t.Helper()
		var record, log, delivery []byte
		var appliedRevision sql.NullInt64
		var deliveryRevision int64
		if err := db.QueryRowContext(ctx, `SELECT r.receipt,l.receipt,l.applied_revision,o.payload,o.revision
			FROM intent_records r JOIN founder_log l ON l.founder_stream_id=r.stream_id AND l.intent_id=r.intent_id
			JOIN transport_player_outbox o ON o.stream_id=r.stream_id AND o.source_id=r.intent_id AND o.message_kind='receipt'
			WHERE r.stream_id=$1 AND r.intent_id=$2`, founderStreamID, intentID).
			Scan(&record, &log, &appliedRevision, &delivery, &deliveryRevision); err != nil {
			t.Fatal(err)
		}
		for _, value := range [][]byte{record, log, delivery} {
			if canonicalFixtureJSON(t, value) != canonicalFixtureJSON(t, receipt) {
				t.Fatalf("%s persisted receipt differs from returned receipt: %s != %s", intentID, value, receipt)
			}
		}
		if deliveryRevision != revision || appliedRevision.Valid != applied || applied && appliedRevision.Int64 != revision {
			t.Fatalf("%s applied revision=%v delivery revision=%d, want applied=%v revision=%d", intentID, appliedRevision, deliveryRevision, applied, revision)
		}
		var eventCount int64
		if applied {
			eventCount = 1
			var envelope founderCosmeticReceipt
			if err := json.Unmarshal(receipt, &envelope); err != nil {
				t.Fatal(err)
			}
			var kind string
			var payload []byte
			var eventRevision int64
			if err := db.QueryRowContext(ctx, `SELECT kind,payload,revision FROM events WHERE stream_id=$1 AND intent_id=$2 AND kind LIKE 'cosmetic_%'`, founderStreamID, intentID).
				Scan(&kind, &payload, &eventRevision); err != nil {
				t.Fatal(err)
			}
			if kind != string(envelope.Event.Kind) || eventRevision != revision || canonicalFixtureJSON(t, payload) != canonicalFixtureValue(t, envelope.Event.Payload) {
				t.Fatalf("%s persisted event differs from receipt: %s/%d/%s", intentID, kind, eventRevision, payload)
			}
			var fields map[string]json.RawMessage
			if err := json.Unmarshal(receipt, &fields); err != nil {
				t.Fatal(err)
			}
			if sweep, ok := fields["fiscal_sweep"]; ok {
				eventCount++
				var expectedPayload map[string]any
				if err := json.Unmarshal(sweep, &expectedPayload); err != nil {
					t.Fatal(err)
				}
				expectedPayload["source"] = "automatic"
				if err := db.QueryRowContext(ctx, `SELECT payload,revision FROM events WHERE stream_id=$1 AND intent_id=$2 AND kind='fiscal_period_harvested.v1'`, founderStreamID, intentID).
					Scan(&payload, &eventRevision); err != nil {
					t.Fatal(err)
				}
				if eventRevision != revision || canonicalFixtureJSON(t, payload) != canonicalFixtureValue(t, expectedPayload) {
					t.Fatal("persisted automatic Fiscal event differs from the independently checked receipt")
				}
			}
		}
		assertRows(t, intentID, 1, eventCount)
	}
	request := func(id, kind string, revision int64, fields string) []byte {
		return []byte(fmt.Sprintf(`{"intent_id":%q,"kind":%q,"expected_revision":%d,%s}`, id, kind, revision, fields))
	}
	assertRejection := func(t *testing.T, result HandleResult, id string, revision int64, category, detail string) {
		t.Helper()
		want := fmt.Sprintf(`{"intent_id":%q,"outcome":"rejected","current_revision":%d,"rejection":{"category":%q,"detail":%q}}`, id, revision, category, detail)
		if canonicalFixtureJSON(t, result.Receipt) != canonicalFixtureJSON(t, []byte(want)) {
			t.Fatalf("receipt=%s, want %s", result.Receipt, want)
		}
	}
	checkRejection := func(t *testing.T, id, kind string, revision int64, fields, category, detail string) {
		t.Helper()
		body := request(id, kind, revision, fields)
		before := observe(t)
		result, err := service.Handle(ctx, companyStreamID, ModeOnline, now, body)
		if err != nil || result.Replay {
			t.Fatalf("first rejection replay=%v err=%v", result.Replay, err)
		}
		assertRejection(t, result, id, revision, category, detail)
		assertStored(t, id, result.Receipt, revision, false)
		after := observe(t)
		if before.founder != after.founder || before.company != after.company || before.counts[0] != after.counts[0] || before.counts[1] != after.counts[1] ||
			after.counts[2] != before.counts[2]+1 || after.counts[3] != before.counts[3]+1 || after.counts[4] != before.counts[4]+1 || after.counts[5] != before.counts[5] {
			t.Fatalf("rejection changed state/events or persisted unexpected rows: before=%+v after=%+v", before.counts, after.counts)
		}
		retry, err := service.Handle(ctx, companyStreamID, ModeOnline, now, body)
		if err != nil || !retry.Replay || string(retry.Receipt) != string(result.Receipt) || observe(t) != after {
			t.Fatalf("rejected retry changed persisted state: replay=%v receipt=%s err=%v", retry.Replay, retry.Receipt, err)
		}
		changed := request(id, kind, revision+1, fields)
		conflict, err := service.Handle(ctx, companyStreamID, ModeOnline, now, changed)
		if err != nil || conflict.Replay {
			t.Fatalf("changed rejected request replay=%v err=%v", conflict.Replay, err)
		}
		assertRejection(t, conflict, id, revision, "idempotency_conflict", id)
		assertStored(t, id, result.Receipt, revision, false)
		if observe(t) != after {
			t.Fatal("changed payload rewrote a recorded rejection")
		}
	}
	for index, row := range []struct{ name, kind, fields, category, detail string }{
		{"unknown-cosmetic", IntentAcquireCosmetic, `"cosmetic_id":"missing_armor"`, "unknown_id", "cosmetic_id"},
		{"unknown-equip-cosmetic", IntentEquipCosmetic, `"cosmetic_id":"missing_armor","pet_id":"` + adoptedPetID + `"`, "unknown_id", "cosmetic_id"},
		{"unknown-pet", IntentEquipCosmetic, `"cosmetic_id":"horse_armor","pet_id":"01986666-7e00-7000-8000-000000000099"`, "unknown_id", "pet_id"},
		{"unknown-unequip-pet", IntentUnequipCosmetic, `"pet_id":"01986666-7e00-7000-8000-000000000099"`, "unknown_id", "pet_id"},
		{"not-owned", IntentEquipCosmetic, `"cosmetic_id":"zebra_armor","pet_id":"` + adoptedPetID + `"`, "not_eligible", "not_owned"},
		{"nothing-equipped", IntentUnequipCosmetic, `"pet_id":"` + adoptedPetID + `"`, "not_eligible", "nothing_equipped"},
	} {
		t.Run("persisted-rejection-"+row.name, func(t *testing.T) {
			id := fmt.Sprintf("01986666-7e00-7000-8000-%012d", index+100)
			checkRejection(t, id, row.kind, 4, row.fields, row.category, row.detail)
		})
	}

	const equipID = "01986666-7e00-7000-8000-000000000110"
	equip := request(equipID, IntentEquipCosmetic, 4, `"cosmetic_id":"horse_armor","pet_id":"`+adoptedPetID+`"`)
	before := observe(t)
	// Fail the LAST insert in ApplyFounderLogged: state, event, history and
	// intent receipt have already been written inside its transaction. The
	// scoped constraint affects only this fixture command, not other requests.
	if _, err := db.ExecContext(ctx, `ALTER TABLE transport_player_outbox ADD CONSTRAINT cosmetic_test_receipt_failure
		CHECK (source_id <> '01986666-7e00-7000-8000-000000000110'::uuid) NOT VALID`); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if _, err := db.ExecContext(ctx, `ALTER TABLE transport_player_outbox DROP CONSTRAINT IF EXISTS cosmetic_test_receipt_failure`); err != nil {
			t.Errorf("remove test-only outbox fault: %v", err)
		}
	}()
	if _, err := service.Handle(ctx, companyStreamID, ModeOnline, now, equip); err == nil || !strings.Contains(err.Error(), "cosmetic_test_receipt_failure") {
		t.Fatalf("expected actual receipt insert failure, got %v", err)
	}
	if observe(t) != before {
		t.Fatal("failed receipt insert left partial state, events, history or outbox rows")
	}
	assertRows(t, equipID, 0, 0)
	if _, err := db.ExecContext(ctx, `ALTER TABLE transport_player_outbox DROP CONSTRAINT cosmetic_test_receipt_failure`); err != nil {
		t.Fatal(err)
	}

	// Concurrent retries of that same request must commit exactly once, not
	// turn the losers into already_equipped or stale-revision rejections.
	type response struct {
		result HandleResult
		err    error
	}
	const submissions = 8
	start := make(chan struct{})
	responses := make(chan response, submissions)
	for range submissions {
		go func() {
			<-start
			result, err := service.Handle(ctx, companyStreamID, ModeOnline, now, equip)
			responses <- response{result, err}
		}()
	}
	close(start)
	completed := make([]response, 0, submissions)
	for range submissions {
		completed = append(completed, <-responses)
	}
	var original json.RawMessage
	fresh := 0
	for _, response := range completed {
		if response.err != nil {
			t.Fatal(response.err)
		}
		if !response.result.Replay {
			fresh++
		}
		if original == nil {
			original = response.result.Receipt
		} else if string(response.result.Receipt) != string(original) {
			t.Fatalf("concurrent retry receipt=%s differs from %s", response.result.Receipt, original)
		}
	}
	wantEquip := fmt.Sprintf(`{"intent_id":%q,"outcome":"applied","founder_revision":5,"kind":"equip_cosmetic","cosmetics":{"owned":["horse_armor"],"equipped":{%q:"horse_armor"}},"event":{"kind":"cosmetic_equipped.v1","payload":{"cosmetic_id":"horse_armor","pet_id":%q,"replaced_cosmetic_id":null}}}`, equipID, adoptedPetID, adoptedPetID)
	if fresh != 1 || canonicalFixtureJSON(t, original) != expectedAppliedReceipt(t, equipID, 5, wantEquip) {
		t.Fatalf("concurrent commits=%d receipt=%s, want one exact equip", fresh, original)
	}
	assertStored(t, equipID, original, 5, true)
	loaded, err := store.LoadLatest(ctx, founderStreamID)
	if err != nil || loaded.Revision.Number != 5 || loaded.State.Cosmetics.Equipped[adoptedPetID] != "horse_armor" {
		t.Fatalf("concurrent equip did not persist exactly one revision: revision=%d err=%v", loaded.Revision.Number, err)
	}
	if observe(t).company != before.company {
		t.Fatal("cosmetic equip changed Company state")
	}
	t.Run("persisted-rejection-already-equipped", func(t *testing.T) {
		checkRejection(t, "01986666-7e00-7000-8000-000000000112", IntentEquipCosmetic, 5,
			`"cosmetic_id":"horse_armor","pet_id":"`+adoptedPetID+`"`, "not_eligible", "already_equipped")
	})

	const staleID = "01986666-7e00-7000-8000-000000000111"
	stale := request(staleID, IntentUnequipCosmetic, 4, `"pet_id":"`+adoptedPetID+`"`)
	before = observe(t)
	result, err := service.Handle(ctx, companyStreamID, ModeOnline, now, stale)
	if err != nil || result.Replay {
		t.Fatalf("stale command replay=%v err=%v", result.Replay, err)
	}
	assertRejection(t, result, staleID, 5, "revision_conflict", "expected_revision")
	assertRows(t, staleID, 0, 0)
	if observe(t) != before {
		t.Fatal("stale request mutated state or persisted a terminal command")
	}
	// A revision conflict is not a recorded terminal intent: the same UUID with
	// a newly read revision can apply. A true recorded rejection cannot change.
	corrected := request(staleID, IntentUnequipCosmetic, 5, `"pet_id":"`+adoptedPetID+`"`)
	result, err = service.Handle(ctx, companyStreamID, ModeOnline, now, corrected)
	if err != nil || result.Replay {
		t.Fatalf("corrected revision replay=%v err=%v", result.Replay, err)
	}
	wantUnequip := fmt.Sprintf(`{"intent_id":%q,"outcome":"applied","founder_revision":6,"kind":"unequip_cosmetic","cosmetics":{"owned":["horse_armor"],"equipped":{}},"event":{"kind":"cosmetic_unequipped.v1","payload":{"cosmetic_id":"horse_armor","pet_id":%q}}}`, staleID, adoptedPetID)
	if canonicalFixtureJSON(t, result.Receipt) != expectedAppliedReceipt(t, staleID, 6, wantUnequip) {
		t.Fatalf("corrected unequip receipt=%s, want %s", result.Receipt, wantUnequip)
	}
	assertStored(t, staleID, result.Receipt, 6, true)
	loaded, err = store.LoadLatest(ctx, founderStreamID)
	if err != nil || loaded.Revision.Number != 6 || !loaded.State.Cosmetics.Owns("horse_armor") || len(loaded.State.Cosmetics.Equipped) != 0 || observe(t).company != before.company {
		t.Fatalf("corrected unequip lost ownership, changed Company or did not persist: err=%v", err)
	}
}

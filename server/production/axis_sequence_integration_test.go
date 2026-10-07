package production

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"cloud-clicker/server/decimal"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/faction"
	"cloud-clicker/server/minigame"
	prestigecore "cloud-clicker/server/prestige"
	"cloud-clicker/server/routes"
	"cloud-clicker/server/save"

	"github.com/jackc/pgx/v5/pgconn"
)

type axisPersistedCensus struct {
	rows, applied, refused, retries, conflicts, faults, outboxRefusals int
}

type axisPersistedAttempt struct {
	raw     []byte
	mode    EvaluationMode
	receipt []byte
}

// Serial, declared disposable Postgres only. SQL triggers inject errors into
// real writes, not a test-side replacement for Service.Handle or Store logic.
func TestAxisSequencePersistenceIntegration(t *testing.T) {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}
	ctx := context.Background()
	db, err := save.OpenPostgres(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := save.Migrate(ctx, db); err != nil {
		t.Fatal(err)
	}
	// No OR REPLACE/CASCADE: unknown pre-existing objects must not be overwritten.
	if _, err := db.ExecContext(ctx, `CREATE FUNCTION axis_sequence_fault() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'axis_sequence_fault:%',TG_ARGV[0] USING ERRCODE='P0001'; END $$`); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if _, err := db.ExecContext(ctx, `DROP FUNCTION axis_sequence_fault()`); err != nil {
			t.Errorf("owned fault-function cleanup: %v", err)
		}
	}()
	raw, err := os.ReadFile("../../testdata/axis-stack/sequence-research-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	var plan axisSequenceReport
	if err := json.Unmarshal(raw, &plan); err != nil || len(plan.Rows) != 16 || plan.CommandAttempts != 136 || plan.Negatives != 48 {
		t.Fatalf("invalid declared command population: %v", err)
	}
	bundle := axisContentBundle(t)
	if bundle.ConstantsHash != plan.Bundle.ConstantsHash {
		t.Fatal("persisted sequence bundle differs from declared source")
	}
	var census axisPersistedCensus
	for _, row := range plan.Rows {
		t.Run(row.ID, func(t *testing.T) {
			axisPersistedSequence(t, ctx, db, bundle, row, &census)
		})
	}
	if census != (axisPersistedCensus{rows: 16, applied: 120, refused: 48, retries: 168, conflicts: 32, faults: 48, outboxRefusals: 16}) {
		t.Fatalf("incomplete persisted population: %+v", census)
	}
	t.Logf("actual16 Handle/Store/Postgres sequences: %+v; full receipts/outbox/replay/retention; OPEN runs, not completed or natural/minted/AC6", census)
}

func axisPersistedUnchanged(t *testing.T, ctx context.Context, db *sql.DB, before map[string]string) {
	t.Helper()
	after := reputationPlanDBSnapshot(t, ctx, db)
	if len(before) != 12 || len(after) != 12 {
		t.Fatal("rollback snapshot lacks full twelve-table census")
	}
	for table, rows := range before {
		if after[table] != rows {
			t.Fatalf("changed full rows in %s", table)
		}
	}
}

func axisPersistedFault(t *testing.T, ctx context.Context, db *sql.DB, service *Service, stream string, mode EvaluationMode, now time.Time, raw []byte, table, operation string, census *axisPersistedCensus) {
	t.Helper()
	stage := table + "." + operation
	predicate := ""
	if table == "transport_player_outbox" {
		predicate = "WHEN (NEW.message_kind='receipt')"
	}
	query := fmt.Sprintf(`CREATE TRIGGER axis_sequence_fail BEFORE %s ON %s FOR EACH ROW %s EXECUTE FUNCTION axis_sequence_fault('%s')`, operation, table, predicate, stage)
	if _, err := db.ExecContext(ctx, query); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if _, err := db.ExecContext(ctx, fmt.Sprintf(`DROP TRIGGER axis_sequence_fail ON %s`, table)); err != nil {
			t.Errorf("owned fault-trigger cleanup: %v", err)
		}
	}()
	before := reputationPlanDBSnapshot(t, ctx, db)
	_, err := service.Handle(ctx, stream, mode, now, raw)
	var pgError *pgconn.PgError
	if !errors.As(err, &pgError) || pgError.Code != "P0001" || pgError.Message != "axis_sequence_fault:"+stage {
		t.Fatalf("live write fault %s not reached exactly: %v", stage, err)
	}
	axisPersistedUnchanged(t, ctx, db, before)
	census.faults++
}

func axisPersistedRequest(t *testing.T, action axisSequenceAction, revision int64, replacements map[string]any) []byte {
	t.Helper()
	var body map[string]any
	if err := json.Unmarshal(action.Payload, &body); err != nil {
		t.Fatal(err)
	}
	var wire replayInputsWire
	if err := json.Unmarshal(action.Inputs, &wire); err != nil {
		t.Fatal(err)
	}
	body["intent_id"] = wire.Command.IntentID
	body["expected_revision"] = revision
	for key, value := range replacements {
		body[key] = value
	}
	raw, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	request, err := ParseIntent(raw)
	if err != nil || request.InvalidDetail != "" {
		t.Fatalf("invalid persisted action: %v %s", err, request.InvalidDetail)
	}
	return raw
}

type axisOutboxReader interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

func axisPersistedEventOutboxMatches(ctx context.Context, reader axisOutboxReader, stream, founder string) (bool, error) {
	var events, outbox, unmatched int
	if err := reader.QueryRowContext(ctx, `SELECT count(*) FROM events WHERE stream_id=$1`, stream).Scan(&events); err != nil {
		return false, err
	}
	if err := reader.QueryRowContext(ctx, `SELECT count(*) FROM transport_player_outbox WHERE stream_id=$1 AND message_kind='event'`, stream).Scan(&outbox); err != nil {
		return false, err
	}
	if err := reader.QueryRowContext(ctx, `SELECT count(*) FROM events e WHERE e.stream_id=$1 AND NOT EXISTS (
		SELECT 1 FROM transport_player_outbox o WHERE o.stream_id=e.stream_id AND o.message_kind='event' AND o.source_id=e.event_id
		AND o.founder_id=$2 AND o.scope='company' AND o.revision=e.revision AND o.constants_hash=e.constants_hash AND o.occurred_at=e.occurred_at
		AND o.payload=jsonb_build_object('event_id',e.event_id::text,'kind',e.kind::text,'scope','company','rev',e.revision,
		'cursor_effect',CASE WHEN e.kind='compensation' THEN 'historical' ELSE 'advance' END,'payload',e.payload))`, stream, founder).Scan(&unmatched); err != nil {
		return false, err
	}
	return events > 0 && events == outbox && unmatched == 0, nil
}

func axisPersistedSequence(t *testing.T, ctx context.Context, db *sql.DB, bundle CatalogBundle, row axisSequenceRow, census *axisPersistedCensus) {
	t.Helper()
	if _, err := db.ExecContext(ctx, `TRUNCATE accounts,save_streams,catalog_sets,epochs RESTART IDENTITY CASCADE`); err != nil {
		t.Fatal(err)
	}
	seedProductionEpoch(t, db, bundle.ConstantsHash, bundle.Artifacts)
	resolver := integrationCatalogs{economy: map[string]*economy.Catalog{bundle.ConstantsHash: bundle.Economy}, routes: map[string]*routes.Catalog{bundle.ConstantsHash: bundle.Routes}, prestige: map[string]*prestigecore.Policy{bundle.ConstantsHash: bundle.Prestige}, factions: map[string]*faction.Catalog{bundle.ConstantsHash: bundle.Faction}}
	store, err := save.NewStore(db, resolver, nil)
	if err != nil {
		t.Fatal(err)
	}
	const account = "01986666-2100-4000-8000-000000000001"
	if _, err := db.ExecContext(ctx, `INSERT INTO accounts(account_id,recovery_hash) VALUES($1,'test')`, account); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO account_founders(account_id,founder_id) VALUES($1,$2)`, account, row.FounderID); err != nil {
		t.Fatal(err)
	}
	var databaseNow time.Time
	if err := db.QueryRowContext(ctx, `SELECT clock_timestamp()`).Scan(&databaseNow); err != nil {
		t.Fatal(err)
	}
	start := save.CanonicalServerTime(databaseNow.Add(-26 * time.Hour))
	company := axisTimingState(t, bundle, start)
	company.GeneratorCounts["generator.beige_tower"], company.GeneratorCounts["generator.beige_tower_v2"], company.GeneratorCounts["generator.legal_dept"] = 99, 10, 1
	company.GeneratorPurchasedTotal, company.ComputeCreditMS = 110, 5000
	setCash(t, company, "1e9")
	spawn, err := initializeActivePlayState(company, bundle.Opportunities, row.FounderID)
	if err != nil || spawn.EffectRowID != row.Effect || spawn.SpawnedAttendedMS != row.FirstSpawnMS {
		t.Fatalf("declared selected spawn differs: %v %+v", err, spawn)
	}
	founder := reputationFounderState(t, bundle, 21, start, 0)
	founder.ExitHistory = []save.ExitRecord{{RunID: 1, ExitType: "collapse", OccurredAt: start.Add(-time.Hour)}}
	for _, definition := range bundle.Achievements.Definitions {
		founder.AchievementsEarnedLifetime[definition.ID] = true
		founder.AchievementScoreLifetime += definition.ScoreGrant
	}
	cr, err := store.CreateStream(ctx, save.StreamKey{OwnerKind: save.OwnerFounder, OwnerID: row.FounderID, Scope: economy.ScopeCompany}, bundle.ConstantsHash, company, save.WriteContext{Cause: "axis.sequence"})
	if err != nil {
		t.Fatal(err)
	}
	fr, err := store.CreateStream(ctx, save.StreamKey{OwnerKind: save.OwnerFounder, OwnerID: row.FounderID, Scope: economy.ScopeFounder}, bundle.ConstantsHash, founder, save.WriteContext{Cause: "axis.sequence"})
	if err != nil {
		t.Fatal(err)
	}
	frozen, err := FrozenFounderContributions(bundle, founder)
	if err != nil {
		t.Fatal(err)
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()
	if _, err := save.PinRunWithGenesisTx(ctx, tx, cr.StreamID, row.FounderID, 2, bundle.ConstantsHash, 19, mustEncodeState(t, company)); err != nil {
		t.Fatal(err)
	}
	if err := save.InsertRunFrozenContributionsTx(ctx, tx, cr.StreamID, 2, frozen); err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
	// Explicit setup-only padding, not invented gameplay or a false replay log.
	if _, err := db.ExecContext(ctx, `INSERT INTO save_revisions(stream_id,revision,version,state,constants_hash) SELECT stream_id,padding,version,state,constants_hash FROM save_revisions CROSS JOIN generate_series(2,7) padding WHERE stream_id=$1 AND revision=1`, cr.StreamID); err != nil {
		t.Fatal(err)
	}
	activity, err := minigame.NewRepository(db)
	if err != nil {
		t.Fatal(err)
	}
	service, err := NewService(store, resolver, FrozenContributionProvider{DB: db}, nil, nil, WithProgressionRuntime(resolver), WithCurrentConstantsHash(bundle.ConstantsHash), WithReplayCatalogs(ReplayCatalogSet{bundle.ConstantsHash: bundle}), WithGuildSettlements(emptyGuildSettlements{}), WithMinigameActivity(activity), WithRouteCatalogs(resolver))
	if err != nil {
		t.Fatal(err)
	}
	initial := reputationPlanDBSnapshot(t, ctx, db)
	head, err := store.LoadLatest(ctx, cr.StreamID)
	if err != nil || head.Revision.Number != 7 {
		t.Fatalf("initial padded head differs: %v", err)
	}
	attempts := []axisPersistedAttempt{}
	normal := func(raw []byte, mode EvaluationMode, now time.Time, outcome, detail string) {
		t.Helper()
		before, beforeRevision := mustEncodeState(t, head.State), head.Revision.Number
		result, err := service.Handle(ctx, cr.StreamID, mode, now, raw)
		var receipt struct {
			Outcome   string                            `json:"outcome"`
			Rejection struct{ Category, Detail string } `json:"rejection"`
		}
		if err != nil || result.Replay || json.Unmarshal(result.Receipt, &receipt) != nil || receipt.Outcome != outcome || detail != "" && (receipt.Rejection.Category != "not_eligible" || receipt.Rejection.Detail != detail) {
			t.Fatalf("normal Handle differs: %v replay=%v receipt=%s", err, result.Replay, result.Receipt)
		}
		head, err = store.LoadLatest(ctx, cr.StreamID)
		if err != nil || head.Revision.ConstantsHash != bundle.ConstantsHash || head.State.WireVersion != 19 || head.State.RunSeq != 2 {
			t.Fatalf("persisted head/hash/floor differs: %v", err)
		}
		if outcome == "applied" {
			census.applied++
			if head.Revision.Number != beforeRevision+1 {
				t.Fatal("applied command did not advance exactly one revision")
			}
			var count, minimum, maximum int64
			if err := db.QueryRowContext(ctx, `SELECT count(*),min(revision),max(revision) FROM save_revisions WHERE stream_id=$1`, cr.StreamID).Scan(&count, &minimum, &maximum); err != nil || count != 5 || minimum != head.Revision.Number-4 || maximum != head.Revision.Number {
				t.Fatalf("actual retention differs: %d [%d,%d] head%d err%v", count, minimum, maximum, head.Revision.Number, err)
			}
		} else {
			census.refused++
			if head.Revision.Number != beforeRevision || !bytes.Equal(before, mustEncodeState(t, head.State)) {
				t.Fatal("refused command changed complete state/head")
			}
		}
		request, err := ParseIntent(raw)
		if err != nil {
			t.Fatal(err)
		}
		var outboxReceipt []byte
		var outboxHash, outboxFounder, outboxScope string
		var outboxRevision int64
		if err := db.QueryRowContext(ctx, `SELECT payload,constants_hash,founder_id,scope,revision FROM transport_player_outbox WHERE stream_id=$1 AND source_id=$2 AND message_kind='receipt'`, cr.StreamID, request.IntentID).Scan(&outboxReceipt, &outboxHash, &outboxFounder, &outboxScope, &outboxRevision); err != nil || !canonicalJSONEqual(outboxReceipt, result.Receipt) || outboxHash != bundle.ConstantsHash || outboxFounder != row.FounderID || outboxScope != "company" || outboxRevision != head.Revision.Number {
			t.Fatalf("complete outbox receipt binding differs: %v", err)
		}
		attempts = append(attempts, axisPersistedAttempt{raw: raw, mode: mode, receipt: result.Receipt})
		for _, table := range []string{"run_log", "intent_records", "transport_player_outbox"} {
			key := "stream_id"
			if table == "run_log" {
				key = "company_stream_id"
			}
			var count int
			predicate := ""
			if table == "transport_player_outbox" {
				predicate = " AND message_kind='receipt'"
			}
			if err := db.QueryRowContext(ctx, fmt.Sprintf(`SELECT count(*) FROM %s WHERE %s=$1%s`, table, key, predicate), cr.StreamID).Scan(&count); err != nil || count != len(attempts) {
				t.Fatalf("%s exact intent census differs: %d/%d %v", table, count, len(attempts), err)
			}
		}
		if matches, err := axisPersistedEventOutboxMatches(ctx, db, cr.StreamID, row.FounderID); err != nil || !matches {
			t.Fatalf("full event/outbox bijection differs: %v", err)
		}
		if len(attempts) == 1 {
			before := reputationPlanDBSnapshot(t, ctx, db)
			probe, err := db.BeginTx(ctx, nil)
			if err != nil {
				t.Fatal(err)
			}
			defer probe.Rollback()
			result, err := probe.ExecContext(ctx, `UPDATE transport_player_outbox SET payload=jsonb_set(payload,'{kind}','"forged.event"'::jsonb) WHERE outbox_id=(SELECT min(outbox_id) FROM transport_player_outbox WHERE stream_id=$1 AND message_kind='event')`, cr.StreamID)
			if err != nil {
				t.Fatal(err)
			}
			changed, err := result.RowsAffected()
			if err != nil || changed != 1 {
				t.Fatalf("outbox corruption not reached: %d %v", changed, err)
			}
			if matches, err := axisPersistedEventOutboxMatches(ctx, probe, cr.StreamID, row.FounderID); err != nil || matches {
				t.Fatalf("corrupt event payload was not refused: %v", err)
			}
			if err := probe.Rollback(); err != nil {
				t.Fatal(err)
			}
			axisPersistedUnchanged(t, ctx, db, before)
			census.outboxRefusals++
		}
		if outcome == "rejected" {
			var count int
			if err := db.QueryRowContext(ctx, `SELECT count(*) FROM events WHERE stream_id=$1 AND intent_id=$2`, cr.StreamID, request.IntentID).Scan(&count); err != nil || count != 0 {
				t.Fatalf("refusal emitted events: %v count%d", err, count)
			}
		}
	}
	base := time.Date(2026, 10, 7, 12, 0, 0, 0, time.UTC).UnixMilli()
	var final time.Time
	var claimedBuff string
	for actionIndex, action := range row.Actions {
		var wire replayInputsWire
		if err := json.Unmarshal(action.Inputs, &wire); err != nil {
			t.Fatal(err)
		}
		now := start.Add(time.Duration(wire.EvaluatedAtMS-base) * time.Millisecond)
		final = now
		raw := axisPersistedRequest(t, action, head.Revision.Number, nil)
		coreIndex := actionIndex - row.PreludeCount
		if row.Gap == 3114 && row.Modes[0] == ModeOnline && (coreIndex == 1 || coreIndex == 3) {
			for _, fault := range []struct{ table, operation string }{{"save_revisions", "INSERT"}, {"events", "INSERT"}, {"run_log", "INSERT"}, {"intent_records", "INSERT"}, {"transport_player_outbox", "INSERT"}, {"save_revisions", "DELETE"}} {
				t.Run(fmt.Sprintf("core%d/%s/%s", coreIndex, fault.table, fault.operation), func(t *testing.T) {
					axisPersistedFault(t, ctx, db, service, cr.StreamID, wire.EvaluationMode, now, raw, fault.table, fault.operation, census)
				})
			}
			if t.Failed() {
				return
			}
		}
		detail := ""
		if coreIndex == 7 {
			detail = "owned"
		}
		normal(raw, wire.EvaluationMode, now, action.Outcome, detail)
		if coreIndex == 0 && (head.State.AttainmentScoreRun != 12 || head.State.AchievementScoreRun != 0 || head.State.PendingOpportunity == nil || head.State.PendingOpportunity.EffectRowID != row.Effect) {
			t.Fatal("persisted independent attainment/spawn differs")
		}
		if coreIndex == 1 {
			if head.State.PendingOpportunity != nil || (row.Effect == "active.lucky" && len(head.State.ActiveBuffs) != 0) || (row.Effect != "active.lucky" && (len(head.State.ActiveBuffs) != 1 || head.State.ActiveBuffs[0].EffectRowID != row.Effect)) {
				t.Fatal("persisted actual claim/buff differs")
			}
			if len(head.State.ActiveBuffs) == 1 {
				claimedBuff = head.State.ActiveBuffs[0].BuffInstanceID
			}
		}
		if coreIndex == 2 && !head.State.UpgradesOwned["upgrade.pr_intern_2"] {
			t.Fatal("persisted PR2 missing")
		}
		if coreIndex == 3 && (head.State.ComputeCreditMS != 2000 || head.State.ComputeBurstRemainingMS != 3000) {
			t.Fatal("persisted exact compute debit/burst missing")
		}
		if coreIndex == 6 && claimedBuff != "" {
			for _, buff := range head.State.ActiveBuffs {
				if buff.BuffInstanceID == claimedBuff {
					t.Fatal("persisted original buff did not expire")
				}
			}
		}
		if coreIndex == 1 || coreIndex == 3 {
			id := fmt.Sprintf("01986666-1e%02x-7000-8000-000000000001", coreIndex)
			refusal := "opportunity_not_pending"
			if coreIndex == 3 {
				refusal = "burst_active"
			}
			normal(axisPersistedRequest(t, action, head.Revision.Number, map[string]any{"intent_id": id}), wire.EvaluationMode, now, "rejected", refusal)
		}
	}
	permits, ok := head.State.Ledger.Balance("company.permits")
	if !ok || !permits.Gt(decimal.Zero) || head.State.GeneratorCounts["generator.beige_tower"] != 101 || (row.Gap == 3114 && head.State.GeneratorProvisioned["generator.beige_tower"] != 0) || (row.Gap == 90000000 && head.State.GeneratorProvisioned["generator.beige_tower"] != 1440) {
		t.Fatal("persisted final resource population differs")
	}
	if len(attempts) != len(row.Actions)+2 {
		t.Fatal("normal/refusal command population truncated")
	}
	committed := reputationPlanDBSnapshot(t, ctx, db)
	for _, attempt := range attempts {
		result, err := service.Handle(ctx, cr.StreamID, attempt.mode, final.Add(time.Second), attempt.raw)
		if err != nil || !result.Replay || !bytes.Equal(result.Receipt, attempt.receipt) {
			t.Fatalf("later identical retry differs: %v replay=%v receipt=%s", err, result.Replay, result.Receipt)
		}
		axisPersistedUnchanged(t, ctx, db, committed)
		census.retries++
	}
	for _, core := range []int{1, 3} {
		action := row.Actions[row.PreludeCount+core]
		var wire replayInputsWire
		if err := json.Unmarshal(action.Inputs, &wire); err != nil {
			t.Fatal(err)
		}
		originalRevision := int64(0)
		for _, attempt := range attempts {
			request, err := ParseIntent(attempt.raw)
			if err != nil {
				t.Fatal(err)
			}
			if request.IntentID == wire.Command.IntentID {
				originalRevision = request.ExpectedRevision
			}
		}
		if originalRevision == 0 {
			t.Fatal("conflict lacks actual original request")
		}
		replacements := map[string]any{"opportunity_id": "01986666-1999-7000-8000-000000000099"}
		if core == 3 {
			replacements = map[string]any{"amount_ms": 1500}
		}
		changed := axisPersistedRequest(t, action, originalRevision, replacements)
		result, err := service.Handle(ctx, cr.StreamID, ModeOnline, final.Add(time.Second), changed)
		var receipt struct {
			Rejection struct{ Category, Detail string } `json:"rejection"`
		}
		if err != nil || result.Replay || json.Unmarshal(result.Receipt, &receipt) != nil || receipt.Rejection.Category != "idempotency_conflict" {
			t.Fatalf("changed same-ID conflict differs: %v receipt%s", err, result.Receipt)
		}
		axisPersistedUnchanged(t, ctx, db, committed)
		census.conflicts++
	}
	genesis, version, entries := reputationCareerReplay(t, db, cr.StreamID, fr.StreamID, 2, nil)
	if version != 19 || len(entries) != len(attempts) || VerifyReplayRun(genesis, version, bundle, entries, bundle.ConstantsHash, false) != ReplayLogGap {
		t.Fatal("OPEN run/census differs")
	}
	working, err := save.RestoreState(genesis, 19, bundle.Economy, economy.ScopeCompany, time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	for index, entry := range entries {
		if entry.Sequence != int64(index+1) || entry.Terminal {
			t.Fatal("actual recorded log order differs")
		}
		transition, err := ApplyLogged(working, entry.CanonicalPayload, bundle, entry.ReplayInputs)
		if err != nil || !canonicalJSONEqual(transition.Receipt, entry.ReceiptJSON) || !canonicalJSONEqual(marshalReplayEvents(transition.Events), entry.EventsJSON) {
			t.Fatalf("recorded transition%d differs: %v", index, err)
		}
		working = transition.State
	}
	if !bytes.Equal(mustEncodeState(t, working), mustEncodeState(t, head.State)) {
		t.Fatal("complete recorded replay head differs")
	}
	for _, table := range []string{"save_streams", "run_epochs", "run_genesis", "run_frozen_contributions", "founder_genesis", "founder_log", "verification_queue"} {
		if initial[table] != committed[table] {
			t.Fatalf("unowned immutable rows changed: %s", table)
		}
	}
	f, err := store.LoadLatest(ctx, fr.StreamID)
	if err != nil || f.Revision.Number != 1 || f.Revision.ConstantsHash != bundle.ConstantsHash || !bytes.Equal(mustEncodeState(t, f.State), mustEncodeState(t, founder)) {
		t.Fatalf("unchanged Founder differs: %v", err)
	}
	census.rows++
}

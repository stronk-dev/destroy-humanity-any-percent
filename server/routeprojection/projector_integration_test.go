package routeprojection

import (
	"context"
	"database/sql"
	"encoding/json"
	"os"
	"sync"
	"testing"
	"time"

	"cloud-clicker/server/routes"
	"cloud-clicker/server/save"
)

type testCatalogs map[string]*routes.Catalog

func (catalogs testCatalogs) ResolveRoutes(hash string) (*routes.Catalog, bool) {
	catalog, ok := catalogs[hash]
	return catalog, ok
}

func TestProjectorIntegrationConvergesAcrossDeliveryOrder(t *testing.T) {
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
	truncateRouteProjection(t, ctx, db)

	data, err := os.ReadFile("../../balance/routes/phase0.json")
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := routes.LoadCatalog(data)
	if err != nil {
		t.Fatal(err)
	}
	const hash = "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	projector, err := New(db, testCatalogs{hash: catalog})
	if err != nil {
		t.Fatal(err)
	}

	const (
		founderA = "33333333-3333-4333-8333-333333333333"
		founderB = "44444444-4444-4444-8444-444444444444"
		eventAID = "10000000-0000-4000-8000-000000000001"
		eventBID = "20000000-0000-4000-8000-000000000002"
	)
	streamA := insertStream(t, ctx, db, founderA, "company")
	streamB := insertStream(t, ctx, db, founderB, "company")
	founderStreamB := insertStream(t, ctx, db, founderB, "founder")
	timeA := time.Date(2026, 7, 29, 14, 0, 0, 0, time.UTC)
	timeB := time.Date(2026, 7, 29, 15, 0, 0, 0, time.UTC)
	eventA := insertExecution(t, ctx, db, eventAID, streamA, founderA, 2, 1, "route.nonprofit_wrapper_zip", "gate.t4_to_t5", hash, timeA)
	eventB := insertExecution(t, ctx, db, eventBID, streamB, founderB, 2, 1, "route.nonprofit_wrapper_zip", "gate.t4_to_t5", hash, timeB)

	// The later event arrives first and provisionally receives both first grants.
	if err := projector.Project(ctx, []save.EventRecord{eventB}); err != nil {
		t.Fatal(err)
	}
	if err := projector.SubmitName(ctx, "route.nonprofit_wrapper_zip", founderB, "Provisional Name", timeB.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	if err := projector.ResolveName(ctx, "route.nonprofit_wrapper_zip", true); err != nil {
		t.Fatal(err)
	}

	hintPayload := json.RawMessage(`{"route_id":"route.nonprofit_wrapper_zip","cost":50}`)
	hintTime := timeB.Add(2 * time.Hour)
	const hintID = "30000000-0000-4000-8000-000000000003"
	if _, err := db.ExecContext(ctx, `INSERT INTO events(event_id,stream_id,revision,schema_version,kind,intent_id,constants_hash,occurred_at,payload) VALUES($1,$2,2,1,'route_hint_purchased',$3,$4,$5,$6)`, hintID, founderStreamB, "018f6b7c-9abc-7def-8abc-333333333333", hash, hintTime, hintPayload); err != nil {
		t.Fatal(err)
	}
	hint := save.EventRecord{EventID: hintID, StreamID: founderStreamB, OwnerID: founderB, Revision: 2, Kind: save.EventRouteHintPurchased, IntentID: "018f6b7c-9abc-7def-8abc-333333333333", ConstantsHash: hash, OccurredAt: hintTime, Payload: hintPayload}
	if err := projector.Project(ctx, []save.EventRecord{hint}); err != nil {
		t.Fatal(err)
	}

	// Delivering the truly earlier event must move every Registry consequence to A.
	if err := projector.Project(ctx, []save.EventRecord{eventA}); err != nil {
		t.Fatal(err)
	}
	assertRegistry(t, ctx, db, "route.nonprofit_wrapper_zip", eventAID, founderA, 2, "reserved", "Nonprofit Wrapper Zip", timeA.Add(72*time.Hour))
	assertKnowledgeState(t, ctx, db, founderA, 125, 0)
	assertKnowledgeState(t, ctx, db, founderB, 0, 25)

	var compensations, activeRegistryGrants int64
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM events WHERE kind='compensation' AND payload->>'reason_key'='route.registry_first_reordered'`).Scan(&compensations); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM events awarded WHERE awarded.kind='route_knowledge_granted' AND awarded.payload->>'source'='registry_first' AND NOT EXISTS (SELECT 1 FROM events compensation WHERE compensation.kind='compensation' AND compensation.payload->>'compensates_event_id'=awarded.event_id::text)`).Scan(&activeRegistryGrants); err != nil {
		t.Fatal(err)
	}
	if compensations != 1 || activeRegistryGrants != 1 {
		t.Fatalf("compensations=%d active Registry grants=%d", compensations, activeRegistryGrants)
	}

	// Every source event remains replay-safe after the correction.
	for _, replay := range []save.EventRecord{eventB, hint, eventA} {
		if err := projector.Project(ctx, []save.EventRecord{replay}); err != nil {
			t.Fatal(err)
		}
	}
	assertRegistry(t, ctx, db, "route.nonprofit_wrapper_zip", eventAID, founderA, 2, "reserved", "Nonprofit Wrapper Zip", timeA.Add(72*time.Hour))

	// Immutable-history repair reconstructs the spent correction as debt.
	if _, err := db.ExecContext(ctx, `DELETE FROM founder_route_state WHERE founder_id=$1`, founderB); err != nil {
		t.Fatal(err)
	}
	var repaired save.State
	if err := projector.RepairFounder(ctx, founderB, &repaired); err != nil || repaired.RouteKnowledgeBalance != 0 {
		t.Fatalf("repair balance=%d err=%v", repaired.RouteKnowledgeBalance, err)
	}
	assertKnowledgeState(t, ctx, db, founderB, 0, 25)

	// A later legitimate grant repays debt before becoming spendable.
	repeatB := insertExecution(t, ctx, db, "40000000-0000-4000-8000-000000000004", streamB, founderB, 3, 2, "route.nonprofit_wrapper_zip", "gate.t4_to_t5", hash, timeB.Add(3*time.Hour))
	if err := projector.Project(ctx, []save.EventRecord{repeatB}); err != nil {
		t.Fatal(err)
	}
	assertKnowledgeState(t, ctx, db, founderB, 0, 20)

	// Equal timestamps use canonical UUID byte order, not delivery order.
	tieTime := timeB.Add(4 * time.Hour)
	const (
		lowID  = "50000000-0000-4000-8000-000000000005"
		highID = "60000000-0000-4000-8000-000000000006"
	)
	low := insertExecution(t, ctx, db, lowID, streamA, founderA, 4, 3, "route.ipo_sequence_break", "gate.t4_to_t5", hash, tieTime)
	high := insertExecution(t, ctx, db, highID, streamB, founderB, 4, 3, "route.ipo_sequence_break", "gate.t4_to_t5", hash, tieTime)
	if err := projector.Project(ctx, []save.EventRecord{high}); err != nil {
		t.Fatal(err)
	}
	if err := projector.Project(ctx, []save.EventRecord{low}); err != nil {
		t.Fatal(err)
	}
	assertRegistry(t, ctx, db, "route.ipo_sequence_break", lowID, founderA, 2, "reserved", "IPO Sequence Break", tieTime.Add(72*time.Hour))

	// Corrupt historical accounting must fail closed and roll back the incoming claim and count.
	brokenTime := tieTime.Add(time.Hour)
	const (
		brokenEarlierID = "70000000-0000-4000-8000-000000000007"
		brokenLaterID   = "80000000-0000-4000-8000-000000000008"
	)
	brokenEarlier := insertExecution(t, ctx, db, brokenEarlierID, streamA, founderA, 6, 4, "route.acquihire_out_of_bounds", "gate.t4_to_t5", hash, brokenTime)
	brokenLater := insertExecution(t, ctx, db, brokenLaterID, streamB, founderB, 5, 4, "route.acquihire_out_of_bounds", "gate.t4_to_t5", hash, brokenTime.Add(time.Hour))
	if err := projector.Project(ctx, []save.EventRecord{brokenLater}); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `DELETE FROM events WHERE stream_id=$1 AND revision=5 AND kind='route_knowledge_granted' AND payload->>'source'='registry_first'`, streamB); err != nil {
		t.Fatal(err)
	}
	if err := projector.Project(ctx, []save.EventRecord{brokenEarlier}); err == nil {
		t.Fatal("missing provisional grant should reject displacement")
	}
	var winner string
	var executions, projectedEarlier int64
	if err := db.QueryRowContext(ctx, `SELECT first_event_id,execution_count FROM registry_routes WHERE route_id='route.acquihire_out_of_bounds'`).Scan(&winner, &executions); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM route_projection_events WHERE event_id=$1`, brokenEarlierID).Scan(&projectedEarlier); err != nil {
		t.Fatal(err)
	}
	if winner != brokenLaterID || executions != 1 || projectedEarlier != 0 {
		t.Fatalf("failed correction committed winner=%s executions=%d projected=%d", winner, executions, projectedEarlier)
	}
}

func truncateRouteProjection(t *testing.T, ctx context.Context, db *sql.DB) {
	t.Helper()
	if _, err := db.ExecContext(ctx, `TRUNCATE commons_recruitment_offers,commons_health_scopes,commons_member_samples,commons_projection_events,company_compact_memberships,founder_commons_assignments,commons_cohorts,registry_routes,route_hint_projection_events,founder_route_state,founder_route_executions,route_projection_events,events,intent_records,save_revisions,save_streams CASCADE`); err != nil {
		t.Fatal(err)
	}
}

func insertStream(t *testing.T, ctx context.Context, db *sql.DB, founderID, scope string) string {
	t.Helper()
	var streamID string
	if err := db.QueryRowContext(ctx, `INSERT INTO save_streams(owner_kind,owner_id,scope) VALUES('founder',$1,$2) RETURNING id`, founderID, scope).Scan(&streamID); err != nil {
		t.Fatal(err)
	}
	return streamID
}

func insertExecution(t *testing.T, ctx context.Context, db *sql.DB, eventID, streamID, founderID string, revision, runSeq int64, routeID, gateID, hash string, occurred time.Time) save.EventRecord {
	t.Helper()
	payload, err := json.Marshal(map[string]any{
		"route_id": routeID,
		"gate_id":  gateID,
		"run_id": map[string]any{
			"company_stream_id": streamID,
			"run_seq":           runSeq,
		},
		"founder_id": founderID,
	})
	if err != nil {
		t.Fatal(err)
	}
	const intentID = "018f6b7c-9abc-7def-8abc-111111111111"
	if _, err := db.ExecContext(ctx, `INSERT INTO events(event_id,stream_id,revision,schema_version,kind,intent_id,constants_hash,occurred_at,payload) VALUES($1,$2,$3,1,'route_executed',$4,$5,$6,$7)`, eventID, streamID, revision, intentID, hash, occurred, payload); err != nil {
		t.Fatal(err)
	}
	return save.EventRecord{EventID: eventID, StreamID: streamID, OwnerID: founderID, Revision: revision, Kind: save.EventRouteExecuted, IntentID: intentID, ConstantsHash: hash, OccurredAt: occurred, Payload: payload}
}

func assertKnowledgeState(t *testing.T, ctx context.Context, db *sql.DB, founderID string, wantBalance, wantDebt int64) {
	t.Helper()
	var balance, debt int64
	if err := db.QueryRowContext(ctx, `SELECT route_knowledge_balance,route_knowledge_debt FROM founder_route_state WHERE founder_id=$1`, founderID).Scan(&balance, &debt); err != nil {
		t.Fatal(err)
	}
	if balance != wantBalance || debt != wantDebt {
		t.Fatalf("founder %s balance/debt=%d/%d want %d/%d", founderID, balance, debt, wantBalance, wantDebt)
	}
}

func assertRegistry(t *testing.T, ctx context.Context, db *sql.DB, routeID, wantEvent, wantFounder string, wantCount int64, wantState, wantName string, wantDeadline time.Time) {
	t.Helper()
	var eventID, founderID, state, name string
	var count int64
	var deadline time.Time
	if err := db.QueryRowContext(ctx, `SELECT first_event_id,first_founder_id,execution_count,name_state,name,naming_reserved_until FROM registry_routes WHERE route_id=$1`, routeID).Scan(&eventID, &founderID, &count, &state, &name, &deadline); err != nil {
		t.Fatal(err)
	}
	if eventID != wantEvent || founderID != wantFounder || count != wantCount || state != wantState || name != wantName || !deadline.Equal(wantDeadline) {
		t.Fatalf("Registry=%s/%s count=%d state=%s name=%q deadline=%s", eventID, founderID, count, state, name, deadline)
	}
}

func TestProjectorIntegrationFirstExecutorRaceReplayAndRepair(t *testing.T) {
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
	if _, err := db.ExecContext(ctx, `TRUNCATE commons_recruitment_offers,commons_health_scopes,commons_member_samples,commons_projection_events,company_compact_memberships,founder_commons_assignments,commons_cohorts,registry_routes, route_hint_projection_events, founder_route_state, founder_route_executions, route_projection_events, events, intent_records, save_revisions, save_streams CASCADE`); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile("../../balance/routes/phase0.json")
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := routes.LoadCatalog(data)
	if err != nil {
		t.Fatal(err)
	}
	const hash = "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	projector, err := New(db, testCatalogs{hash: catalog})
	if err != nil {
		t.Fatal(err)
	}

	founders := []string{"11111111-1111-4111-8111-111111111111", "22222222-2222-4222-8222-222222222222"}
	streamIDs := make([]string, 2)
	for index, founderID := range founders {
		if err := db.QueryRowContext(ctx, `INSERT INTO save_streams(owner_kind,owner_id,scope) VALUES('founder',$1,'company') RETURNING id`, founderID).Scan(&streamIDs[index]); err != nil {
			t.Fatal(err)
		}
	}
	occurred := time.Date(2026, 7, 29, 11, 0, 0, 0, time.UTC)
	events := make([]save.EventRecord, 2)
	for index := range events {
		payload, _ := json.Marshal(map[string]any{
			"route_id": "route.nonprofit_wrapper_zip", "gate_id": "gate.t4_to_t5",
			"run_id": map[string]any{"company_stream_id": streamIDs[index], "run_seq": 1}, "founder_id": founders[index],
		})
		var eventID string
		if err := db.QueryRowContext(ctx, `INSERT INTO events(stream_id,revision,schema_version,kind,intent_id,constants_hash,occurred_at,payload) VALUES($1,2,1,'route_executed',$2,$3,$4,$5) RETURNING event_id`, streamIDs[index], "018f6b7c-9abc-7def-8abc-111111111111", hash, occurred, payload).Scan(&eventID); err != nil {
			t.Fatal(err)
		}
		events[index] = save.EventRecord{EventID: eventID, StreamID: streamIDs[index], OwnerID: founders[index], Revision: 2, Kind: save.EventRouteExecuted, IntentID: "018f6b7c-9abc-7def-8abc-111111111111", ConstantsHash: hash, OccurredAt: occurred, Payload: payload}
	}

	var wait sync.WaitGroup
	errorsFound := make(chan error, 2)
	for index := range events {
		wait.Add(1)
		go func(event save.EventRecord) {
			defer wait.Done()
			errorsFound <- projector.Project(ctx, []save.EventRecord{event})
		}(events[index])
	}
	wait.Wait()
	close(errorsFound)
	for err := range errorsFound {
		if err != nil {
			t.Fatal(err)
		}
	}
	for _, event := range events {
		if err := projector.Project(ctx, []save.EventRecord{event}); err != nil {
			t.Fatal(err)
		}
	}

	var registryCount, projectedCount, grantEvents int
	var firstFounder, nameState string
	if err := db.QueryRowContext(ctx, `SELECT execution_count,first_founder_id,name_state FROM registry_routes WHERE route_id='route.nonprofit_wrapper_zip'`).Scan(&registryCount, &firstFounder, &nameState); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM route_projection_events`).Scan(&projectedCount); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM events awarded WHERE awarded.kind='route_knowledge_granted' AND NOT EXISTS (SELECT 1 FROM events compensation WHERE compensation.kind='compensation' AND compensation.payload->>'compensates_event_id'=awarded.event_id::text)`).Scan(&grantEvents); err != nil {
		t.Fatal(err)
	}
	if registryCount != 2 || projectedCount != 2 || grantEvents != 3 || nameState != "reserved" {
		t.Fatalf("registry=%d projected=%d grants=%d state=%s", registryCount, projectedCount, grantEvents, nameState)
	}

	balances := map[string]int64{}
	rows, err := db.QueryContext(ctx, `SELECT founder_id,route_knowledge_balance FROM founder_route_state`)
	if err != nil {
		t.Fatal(err)
	}
	for rows.Next() {
		var founder string
		var balance int64
		if err := rows.Scan(&founder, &balance); err != nil {
			t.Fatal(err)
		}
		balances[founder] = balance
	}
	rows.Close()
	if balances[firstFounder] != 125 {
		t.Fatalf("first balance=%d all=%+v", balances[firstFounder], balances)
	}
	for _, founder := range founders {
		if founder != firstFounder && balances[founder] != 25 {
			t.Fatalf("loser balance=%d all=%+v", balances[founder], balances)
		}
	}

	var founderStream string
	if err := db.QueryRowContext(ctx, `INSERT INTO save_streams(owner_kind,owner_id,scope) VALUES('founder',$1,'founder') RETURNING id`, firstFounder).Scan(&founderStream); err != nil {
		t.Fatal(err)
	}
	hintPayload := json.RawMessage(`{"route_id":"route.nonprofit_wrapper_zip","cost":50}`)
	var hintID string
	if err := db.QueryRowContext(ctx, `INSERT INTO events(stream_id,revision,schema_version,kind,intent_id,constants_hash,occurred_at,payload) VALUES($1,2,1,'route_hint_purchased',$2,$3,$4,$5) RETURNING event_id`, founderStream, "018f6b7c-9abc-7def-8abc-333333333333", hash, occurred.Add(time.Second), hintPayload).Scan(&hintID); err != nil {
		t.Fatal(err)
	}
	hint := save.EventRecord{EventID: hintID, StreamID: founderStream, OwnerID: firstFounder, Revision: 2, Kind: save.EventRouteHintPurchased, IntentID: "018f6b7c-9abc-7def-8abc-333333333333", ConstantsHash: hash, OccurredAt: occurred.Add(time.Second), Payload: hintPayload}
	if err := projector.Project(ctx, []save.EventRecord{hint}); err != nil {
		t.Fatal(err)
	}
	if err := projector.Project(ctx, []save.EventRecord{hint}); err != nil {
		t.Fatal(err)
	}
	var repaired save.State
	if err := projector.RepairFounder(ctx, firstFounder, &repaired); err != nil || repaired.RouteKnowledgeBalance != 75 {
		t.Fatalf("repaired=%d err=%v", repaired.RouteKnowledgeBalance, err)
	}
	if _, err := db.ExecContext(ctx, `DELETE FROM founder_route_state WHERE founder_id=$1`, firstFounder); err != nil {
		t.Fatal(err)
	}
	if err := projector.RepairFounder(ctx, firstFounder, &repaired); err != nil || repaired.RouteKnowledgeBalance != 75 {
		t.Fatalf("read repair=%d err=%v", repaired.RouteKnowledgeBalance, err)
	}

	if err := projector.SubmitName(ctx, "route.nonprofit_wrapper_zip", firstFounder, "First Public Name", occurred.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	if err := projector.ResolveName(ctx, "route.nonprofit_wrapper_zip", true); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT name_state FROM registry_routes WHERE route_id='route.nonprofit_wrapper_zip'`).Scan(&nameState); err != nil || nameState != "published" {
		t.Fatalf("name state=%s err=%v", nameState, err)
	}
	if count, err := projector.FounderDistinctRoutes(ctx, firstFounder); err != nil || count != 1 {
		t.Fatalf("distinct=%d err=%v", count, err)
	}
}

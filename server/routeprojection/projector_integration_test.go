package routeprojection

import (
	"context"
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
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM events WHERE kind='route_knowledge_granted'`).Scan(&grantEvents); err != nil {
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

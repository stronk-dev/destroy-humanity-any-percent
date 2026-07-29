package commonsprojection

import (
	"context"
	"encoding/json"
	"os"
	"sync"
	"testing"
	"time"

	"cloud-clicker/server/commons"
	"cloud-clicker/server/save"
)

type fixedAssignments struct{ serverID string }

func (value fixedAssignments) ResolveAssignment(string) (AssignmentContext, bool) {
	return AssignmentContext{ServerID: value.serverID, ActivityBracket: "activity.standard"}, true
}

func TestProjectorIntegrationConcurrentAssignmentReplayAndLeave(t *testing.T) {
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
	if _, err := db.ExecContext(ctx, `TRUNCATE commons_projection_events,company_compact_memberships,founder_commons_assignments,commons_cohorts,registry_routes,route_hint_projection_events,founder_route_state,founder_route_executions,route_projection_events,events,intent_records,save_revisions,save_streams CASCADE`); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile("../../balance/commons/phase0.json")
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := commons.LoadCatalog(data)
	if err != nil {
		t.Fatal(err)
	}
	const hash = "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	const serverID = "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"
	projector, err := New(db, fixedAssignments{serverID: serverID}, commons.CatalogSet{hash: catalog})
	if err != nil {
		t.Fatal(err)
	}
	founders := []string{"11111111-1111-4111-8111-111111111111", "22222222-2222-4222-8222-222222222222"}
	streams := make([]string, len(founders))
	occurred := time.Date(2026, 7, 29, 14, 0, 0, 0, time.UTC)
	events := make([]save.EventRecord, len(founders))
	for index, founder := range founders {
		if err := db.QueryRowContext(ctx, `INSERT INTO save_streams(owner_kind,owner_id,scope) VALUES('founder',$1,'company') RETURNING id`, founder).Scan(&streams[index]); err != nil {
			t.Fatal(err)
		}
		payload, _ := json.Marshal(map[string]any{"founder_id": founder, "run_id": map[string]any{"company_stream_id": streams[index], "run_seq": 1}, "tithe_ppm": 100000, "prior_member": false, "new_member": true})
		var eventID string
		if err := db.QueryRowContext(ctx, `INSERT INTO events(stream_id,revision,schema_version,kind,intent_id,constants_hash,occurred_at,payload) VALUES($1,2,1,'compact_signed',$2,$3,$4,$5) RETURNING event_id`, streams[index], "018f6b7c-9abc-7def-8abc-111111111111", hash, occurred, payload).Scan(&eventID); err != nil {
			t.Fatal(err)
		}
		events[index] = save.EventRecord{EventID: eventID, StreamID: streams[index], OwnerID: founder, Revision: 2, Kind: save.EventCompactSigned, ConstantsHash: hash, OccurredAt: occurred, Payload: payload}
	}
	var wait sync.WaitGroup
	failures := make(chan error, len(events))
	for _, event := range events {
		wait.Add(1)
		go func(event save.EventRecord) {
			defer wait.Done()
			failures <- projector.Project(ctx, []save.EventRecord{event})
		}(event)
	}
	wait.Wait()
	close(failures)
	for err := range failures {
		if err != nil {
			t.Fatal(err)
		}
	}
	for _, event := range events {
		if err := projector.Project(ctx, []save.EventRecord{event}); err != nil {
			t.Fatal(err)
		}
	}
	left, err := projector.FounderCohort(ctx, founders[0])
	if err != nil {
		t.Fatal(err)
	}
	right, err := projector.FounderCohort(ctx, founders[1])
	if err != nil {
		t.Fatal(err)
	}
	var cohorts, members, projected int
	if err := db.QueryRowContext(ctx, `SELECT count(*),COALESCE(sum(member_count),0) FROM commons_cohorts`).Scan(&cohorts, &members); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM commons_projection_events`).Scan(&projected); err != nil {
		t.Fatal(err)
	}
	if left != right || cohorts != 1 || members != 2 || projected != 2 {
		t.Fatalf("cohorts=%d members=%d projected=%d left=%s right=%s", cohorts, members, projected, left, right)
	}
	payload, _ := json.Marshal(map[string]any{"founder_id": founders[0], "run_id": map[string]any{"company_stream_id": streams[0], "run_seq": 1}, "tithe_ppm": 100000, "prior_member": true, "new_member": false})
	var leaveID string
	if err := db.QueryRowContext(ctx, `INSERT INTO events(stream_id,revision,schema_version,kind,intent_id,constants_hash,occurred_at,payload) VALUES($1,3,1,'compact_left',$2,$3,$4,$5) RETURNING event_id`, streams[0], "018f6b7c-9abc-7def-8abc-222222222222", hash, occurred.Add(time.Second), payload).Scan(&leaveID); err != nil {
		t.Fatal(err)
	}
	leave := save.EventRecord{EventID: leaveID, StreamID: streams[0], OwnerID: founders[0], Revision: 3, Kind: save.EventCompactLeft, ConstantsHash: hash, OccurredAt: occurred.Add(time.Second), Payload: payload}
	if err := projector.Project(ctx, []save.EventRecord{leave}); err != nil {
		t.Fatal(err)
	}
	if err := projector.Project(ctx, []save.EventRecord{leave}); err != nil {
		t.Fatal(err)
	}
	var member bool
	if err := db.QueryRowContext(ctx, `SELECT member FROM company_compact_memberships WHERE company_stream_id=$1`, streams[0]).Scan(&member); err != nil || member {
		t.Fatalf("member=%v err=%v", member, err)
	}
	stable, err := projector.FounderCohort(ctx, founders[0])
	if err != nil || stable != left {
		t.Fatalf("stable=%s want=%s err=%v", stable, left, err)
	}
}

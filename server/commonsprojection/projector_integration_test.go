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

func TestHealthBandsUseCatalogThresholds(t *testing.T) {
	data, err := os.ReadFile("../../balance/commons/phase0.json")
	if err != nil { t.Fatal(err) }
	catalog, err := commons.LoadCatalog(data)
	if err != nil { t.Fatal(err) }
	for value, expected := range map[int64]string{349_999: "collapsed", 350_000: "strained", 799_999: "strained", 800_000: "healthy"} {
		if actual := healthBand(catalog, value); actual != expected { t.Fatalf("health %d band=%s want=%s", value, actual, expected) }
	}
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
	if _, err := db.ExecContext(ctx, `TRUNCATE commons_recruitment_offers,commons_health_scopes,commons_member_samples,commons_projection_events,company_compact_memberships,founder_commons_assignments,commons_cohorts,registry_routes,route_hint_projection_events,founder_route_state,founder_route_executions,route_projection_events,events,intent_records,save_revisions,save_streams CASCADE`); err != nil {
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
	recruitFounder := "33333333-3333-4333-8333-333333333333"
	var recruitStream string
	if err := db.QueryRowContext(ctx, `INSERT INTO save_streams(owner_kind,owner_id,scope) VALUES('founder',$1,'company') RETURNING id`, recruitFounder).Scan(&recruitStream); err != nil {
		t.Fatal(err)
	}
	offered, err := projector.OfferRecruitment(ctx, recruitStream, recruitFounder, 1, hash, time.Date(2026, 7, 29, 13, 0, 0, 0, time.UTC))
	if err != nil || !offered {
		t.Fatalf("offered=%v err=%v", offered, err)
	}
	offered, err = projector.OfferRecruitment(ctx, recruitStream, recruitFounder, 1, hash, time.Date(2026, 7, 29, 13, 1, 0, 0, time.UTC))
	if err != nil || offered {
		t.Fatalf("repeat offered=%v err=%v", offered, err)
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
	samplePayload, _ := json.Marshal(map[string]any{"founder_id": founders[0], "run_id": map[string]any{"company_stream_id": streams[0], "run_seq": 1}, "weight_ppm": 1000000, "compliance_ppm": 1000000, "enclosure": "0", "capacity": "1e0", "solidarity_ppm": 1000, "sampled_ms": 3600000})
	var sampleID string
	if err := db.QueryRowContext(ctx, `INSERT INTO events(stream_id,revision,schema_version,kind,intent_id,constants_hash,occurred_at,payload) VALUES($1,3,1,'compact_sampled',$2,$3,$4,$5) RETURNING event_id`, streams[0], "018f6b7c-9abc-7def-8abc-333333333333", hash, occurred.Add(time.Second), samplePayload).Scan(&sampleID); err != nil {
		t.Fatal(err)
	}
	sample := save.EventRecord{EventID: sampleID, StreamID: streams[0], OwnerID: founders[0], Revision: 3, Kind: save.EventCompactSampled, ConstantsHash: hash, OccurredAt: occurred.Add(time.Second), Payload: samplePayload}
	if err := projector.Project(ctx, []save.EventRecord{sample}); err != nil {
		t.Fatal(err)
	}
	if err := projector.Project(ctx, []save.EventRecord{sample}); err != nil {
		t.Fatal(err)
	}
	snapshot, err := projector.Snapshot(ctx, founders[0])
	if err != nil || snapshot.HealthPPM != 762195 || snapshot.CohortCapacity != "1e0" || snapshot.ServerCapacity != "1e0" || snapshot.NPCWeightPPM != 19_500_000 {
		t.Fatalf("snapshot=%+v err=%v", snapshot, err)
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
	var splitID string
	if err := db.QueryRowContext(ctx, `INSERT INTO commons_cohorts(server_id,activity_bracket,cohort_seq,member_count) VALUES($1,'activity.standard',2,1) RETURNING cohort_id`, serverID).Scan(&splitID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `UPDATE founder_commons_assignments SET cohort_id=$1 WHERE founder_id=$2`, splitID, founders[1]); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `UPDATE company_compact_memberships SET cohort_id=$1 WHERE founder_id=$2`, splitID, founders[1]); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `UPDATE commons_cohorts SET member_count=1 WHERE cohort_id=$1`, left); err != nil {
		t.Fatal(err)
	}
	merged, err := projector.MergeCollapsed(ctx, hash, serverID, "activity.standard", occurred.Add(2*time.Second))
	if err != nil || merged != 1 {
		t.Fatalf("merged=%d err=%v", merged, err)
	}
	mergedCohort, err := projector.FounderCohort(ctx, founders[1])
	if err != nil || mergedCohort != left {
		t.Fatalf("merged cohort=%s want=%s err=%v", mergedCohort, left, err)
	}
	var closed bool
	if err := db.QueryRowContext(ctx, `SELECT closed_at IS NOT NULL FROM commons_cohorts WHERE cohort_id=$1`, splitID).Scan(&closed); err != nil || !closed {
		t.Fatalf("closed=%v err=%v", closed, err)
	}
}

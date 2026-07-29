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
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := commons.LoadCatalog(data)
	if err != nil {
		t.Fatal(err)
	}
	for value, expected := range map[int64]string{349_999: "collapsed", 350_000: "strained", 799_999: "strained", 800_000: "healthy"} {
		if actual := healthBand(catalog, value); actual != expected {
			t.Fatalf("health %d band=%s want=%s", value, actual, expected)
		}
	}
}

func TestProjectionOrderUsesRevisionForSameStreamTimestamp(t *testing.T) {
	at := time.Date(2026, 7, 29, 15, 0, 0, 0, time.UTC)
	leave := save.EventRecord{EventID: "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb", StreamID: "stream.same", Revision: 8, Kind: save.EventCompactLeft, OccurredAt: at}
	resign := save.EventRecord{EventID: "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa", StreamID: "stream.same", Revision: 9, Kind: save.EventCompactSigned, OccurredAt: at}
	if !projectionEventBefore(leave, resign) || projectionEventBefore(resign, leave) {
		t.Fatal("same-stream timestamp tie did not follow revision order")
	}
	other := resign
	other.StreamID = "stream.other"
	if !projectionEventBefore(other, leave) {
		t.Fatal("cross-stream timestamp tie did not retain kind priority")
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
	const retunedHash = "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	const serverID = "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"
	retuned := *catalog
	retuned.GuildHealthWeightPPM = 100_000
	retuned.CohortHealthWeightPPM = 100_000
	retuned.ServerHealthWeightPPM = 800_000
	projector, err := New(db, fixedAssignments{serverID: serverID}, commons.CatalogSet{hash: catalog, retunedHash: &retuned})
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
	snapshot, err := projector.Snapshot(ctx, founders[0], hash)
	if err != nil || snapshot.HealthPPM != 762195 || snapshot.CohortCapacity != "1e0" || snapshot.ServerCapacity != "1e0" || snapshot.NPCWeightPPM != 19_500_000 {
		t.Fatalf("snapshot=%+v err=%v", snapshot, err)
	}
	if _, err := db.ExecContext(ctx, `UPDATE commons_health_scopes SET health_ppm=500000 WHERE scope_kind='server' AND scope_id=$1`, serverID); err != nil {
		t.Fatal(err)
	}
	defaultWeighted, err := projector.Snapshot(ctx, founders[0], hash)
	if err != nil {
		t.Fatal(err)
	}
	retunedSnapshot, err := projector.Snapshot(ctx, founders[0], retunedHash)
	if err != nil {
		t.Fatal(err)
	}
	wantRetuned, err := commons.EffectiveHealthPPM(&retuned, 0, retunedSnapshot.CohortHealthPPM, retunedSnapshot.ServerHealthPPM, false)
	if err != nil || retunedSnapshot.HealthPPM != wantRetuned || retunedSnapshot.HealthPPM == defaultWeighted.HealthPPM {
		t.Fatalf("retuned snapshot=%+v want=%d default=%d err=%v", retunedSnapshot, wantRetuned, defaultWeighted.HealthPPM, err)
	}
	secondPayload, _ := json.Marshal(map[string]any{"founder_id": founders[0], "run_id": map[string]any{"company_stream_id": streams[0], "run_seq": 1}, "weight_ppm": 1000000, "compliance_ppm": 1000000, "enclosure": "0", "capacity": "2e0", "solidarity_ppm": 2000, "sampled_ms": 3600000})
	var secondID string
	if err := db.QueryRowContext(ctx, `INSERT INTO events(stream_id,revision,schema_version,kind,intent_id,constants_hash,occurred_at,payload) VALUES($1,4,1,'compact_sampled',$2,$3,$4,$5) RETURNING event_id`, streams[0], "018f6b7c-9abc-7def-8abc-444444444444", hash, occurred.Add(2*time.Second), secondPayload).Scan(&secondID); err != nil {
		t.Fatal(err)
	}
	second := save.EventRecord{EventID: secondID, StreamID: streams[0], OwnerID: founders[0], Revision: 4, Kind: save.EventCompactSampled, ConstantsHash: hash, OccurredAt: occurred.Add(2 * time.Second), Payload: secondPayload}
	if err := projector.Project(ctx, []save.EventRecord{second}); err != nil {
		t.Fatal(err)
	}
	if err := projector.Project(ctx, []save.EventRecord{second}); err != nil {
		t.Fatal(err)
	}
	snapshot, err = projector.Snapshot(ctx, founders[0], hash)
	if err != nil || snapshot.CohortCapacity != "3e0" || snapshot.ServerCapacity != "3e0" {
		t.Fatalf("cumulative snapshot=%+v err=%v", snapshot, err)
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
	if err := projector.Project(ctx, []save.EventRecord{events[0], sample, leave}); err != nil {
		t.Fatalf("replay signed/sample/left after leave: %v", err)
	}
	var replayCapacity string
	if err := db.QueryRowContext(ctx, `SELECT capacity FROM commons_member_samples WHERE company_stream_id=$1`, streams[0]).Scan(&replayCapacity); err != nil || replayCapacity != "3e0" {
		t.Fatalf("replay capacity=%s err=%v", replayCapacity, err)
	}
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM commons_projection_events`).Scan(&projected); err != nil || projected != 5 {
		t.Fatalf("projected after replay=%d err=%v", projected, err)
	}
	var invalidSampleID string
	if err := db.QueryRowContext(ctx, `INSERT INTO events(stream_id,revision,schema_version,kind,intent_id,constants_hash,occurred_at,payload) VALUES($1,5,1,'compact_sampled',$2,$3,$4,$5) RETURNING event_id`, streams[0], "018f6b7c-9abc-7def-8abc-555555555555", hash, occurred.Add(3*time.Second), samplePayload).Scan(&invalidSampleID); err != nil {
		t.Fatal(err)
	}
	invalidSample := save.EventRecord{EventID: invalidSampleID, StreamID: streams[0], OwnerID: founders[0], Revision: 5, Kind: save.EventCompactSampled, ConstantsHash: hash, OccurredAt: occurred.Add(3 * time.Second), Payload: samplePayload}
	if err := projector.Project(ctx, []save.EventRecord{invalidSample}); err == nil {
		t.Fatal("first-delivery sample after leave succeeded")
	}
	var invalidProjected int
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM commons_projection_events WHERE event_id=$1`, invalidSampleID).Scan(&invalidProjected); err != nil || invalidProjected != 0 {
		t.Fatalf("failed sample dedup rows=%d err=%v", invalidProjected, err)
	}

	tiedAt := occurred.Add(4 * time.Second)
	resignPayload, _ := json.Marshal(map[string]any{"founder_id": founders[0], "run_id": map[string]any{"company_stream_id": streams[0], "run_seq": 1}, "tithe_ppm": 100000, "prior_member": false, "new_member": true})
	var resignID string
	if err := db.QueryRowContext(ctx, `INSERT INTO events(stream_id,revision,schema_version,kind,intent_id,constants_hash,occurred_at,payload) VALUES($1,7,1,'compact_signed',$2,$3,$4,$5) RETURNING event_id`, streams[0], "018f6b7c-9abc-7def-8abc-777777777777", hash, tiedAt, resignPayload).Scan(&resignID); err != nil {
		t.Fatal(err)
	}
	staleLeavePayload, _ := json.Marshal(map[string]any{"founder_id": founders[0], "run_id": map[string]any{"company_stream_id": streams[0], "run_seq": 1}, "tithe_ppm": 100000, "prior_member": true, "new_member": false})
	var staleLeaveID string
	if err := db.QueryRowContext(ctx, `INSERT INTO events(stream_id,revision,schema_version,kind,intent_id,constants_hash,occurred_at,payload) VALUES($1,6,1,'compact_left',$2,$3,$4,$5) RETURNING event_id`, streams[0], "018f6b7c-9abc-7def-8abc-666666666666", hash, tiedAt, staleLeavePayload).Scan(&staleLeaveID); err != nil {
		t.Fatal(err)
	}
	resign := save.EventRecord{EventID: resignID, StreamID: streams[0], OwnerID: founders[0], Revision: 7, Kind: save.EventCompactSigned, ConstantsHash: hash, OccurredAt: tiedAt, Payload: resignPayload}
	staleLeave := save.EventRecord{EventID: staleLeaveID, StreamID: streams[0], OwnerID: founders[0], Revision: 6, Kind: save.EventCompactLeft, ConstantsHash: hash, OccurredAt: tiedAt, Payload: staleLeavePayload}
	if err := projector.Project(ctx, []save.EventRecord{resign}); err != nil {
		t.Fatal(err)
	}
	if err := projector.Project(ctx, []save.EventRecord{staleLeave}); err != nil {
		t.Fatalf("stale leave after re-sign: %v", err)
	}
	var projectedRevision int64
	if err := db.QueryRowContext(ctx, `SELECT member,projected_revision FROM company_compact_memberships WHERE company_stream_id=$1`, streams[0]).Scan(&member, &projectedRevision); err != nil || !member || projectedRevision != 7 {
		t.Fatalf("same-time out-of-order membership member=%v revision=%d err=%v", member, projectedRevision, err)
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

	var overflowID string
	if err := db.QueryRowContext(ctx, `INSERT INTO commons_cohorts(server_id,activity_bracket,cohort_seq,member_count) VALUES($1,'activity.standard',3,26) RETURNING cohort_id`, serverID).Scan(&overflowID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `UPDATE founder_commons_assignments SET cohort_id=$1 WHERE founder_id=$2`, overflowID, founders[1]); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `UPDATE company_compact_memberships SET cohort_id=$1 WHERE founder_id=$2`, overflowID, founders[1]); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `UPDATE commons_cohorts SET member_count=200 WHERE cohort_id=$1`, left); err != nil {
		t.Fatal(err)
	}
	merged, err = projector.MergeCollapsed(ctx, hash, serverID, "activity.standard", occurred.Add(3*time.Second))
	if err != nil || merged != 0 {
		t.Fatalf("over-cap merge=%d err=%v", merged, err)
	}
	unchanged, err := projector.FounderCohort(ctx, founders[1])
	if err != nil || unchanged != overflowID {
		t.Fatalf("over-cap source moved cohort=%s want=%s err=%v", unchanged, overflowID, err)
	}
	if err := db.QueryRowContext(ctx, `SELECT closed_at IS NOT NULL FROM commons_cohorts WHERE cohort_id=$1`, overflowID).Scan(&closed); err != nil || closed {
		t.Fatalf("over-cap source closed=%v err=%v", closed, err)
	}
	if _, err := db.ExecContext(ctx, `UPDATE commons_cohorts SET member_count=25 WHERE cohort_id=$1`, overflowID); err != nil {
		t.Fatal(err)
	}
	merged, err = projector.MergeCollapsed(ctx, hash, serverID, "activity.standard", occurred.Add(4*time.Second))
	if err != nil || merged != 1 {
		t.Fatalf("exact-cap merge=%d err=%v", merged, err)
	}
	var targetMembers int
	if err := db.QueryRowContext(ctx, `SELECT member_count FROM commons_cohorts WHERE cohort_id=$1`, left).Scan(&targetMembers); err != nil || targetMembers != 225 {
		t.Fatalf("exact-cap target members=%d err=%v", targetMembers, err)
	}

	var atFloorID string
	if err := db.QueryRowContext(ctx, `INSERT INTO commons_cohorts(server_id,activity_bracket,cohort_seq,member_count) VALUES($1,'activity.standard',4,40) RETURNING cohort_id`, serverID).Scan(&atFloorID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `UPDATE commons_cohorts SET member_count=150 WHERE cohort_id=$1`, left); err != nil {
		t.Fatal(err)
	}
	merged, err = projector.MergeCollapsed(ctx, hash, serverID, "activity.standard", occurred.Add(5*time.Second))
	if err != nil || merged != 0 {
		t.Fatalf("at-floor merge=%d err=%v", merged, err)
	}
	if err := db.QueryRowContext(ctx, `SELECT closed_at IS NOT NULL FROM commons_cohorts WHERE cohort_id=$1`, atFloorID).Scan(&closed); err != nil || closed {
		t.Fatalf("at-floor source closed=%v err=%v", closed, err)
	}
}

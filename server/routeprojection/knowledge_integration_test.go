package routeprojection

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"cloud-clicker/server/decimal"
	"cloud-clicker/server/routes"
	"cloud-clicker/server/save"
)

func TestExitKnowledgeProjectionDeliveryAndRebuildIntegration(t *testing.T) {
	for _, mode := range []string{"missing", "existing", "rollback", "concurrent", "invalid_source"} {
		t.Run(mode, func(t *testing.T) {
			db, projector, founder, companyStream, founderStream := exitKnowledgeFixture(t)
			ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
			defer cancel()
			const hash = "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
			at := time.Date(2026, 11, 5, 12, 0, 0, 0, time.UTC)
			want := int64(25)
			if mode == "existing" {
				execution := insertExecution(t, ctx, db, "01986666-cc00-7000-8000-000000000001", companyStream,
					founder, 1, 1, "route.nonprofit_wrapper_zip", "gate.t4_to_t5", hash, at.Add(-time.Minute))
				if err := projector.Project(ctx, []save.EventRecord{execution}); err != nil {
					t.Fatal(err)
				}
				assertKnowledgeState(t, ctx, db, founder, 125, 0)
				want += 125
			}
			payloadFounder := founder
			if mode == "invalid_source" {
				payloadFounder = "01986666-cc00-7000-8000-000000000002"
			}
			advance := insertKnowledgeFact(t, ctx, db, founderStream, founder, hash, at, save.EventFounderAdvanced,
				map[string]any{"founder_id": payloadFounder, "run_id": map[string]any{"company_stream_id": companyStream, "run_seq": 1},
					"exit_type": "scripted_first", "reputation_delta": 0, "route_knowledge": 75, "occurred_at_ms": at.UnixMilli()})
			hint := insertKnowledgeFact(t, ctx, db, founderStream, founder, hash, at.Add(time.Second), save.EventRouteHintPurchased,
				map[string]any{"route_id": "route.nonprofit_wrapper_zip", "cost": 50})
			if mode == "invalid_source" {
				state := &save.State{RouteKnowledgeBalance: 37}
				if err := projector.RepairFounder(ctx, founder, state); !errors.Is(err, ErrProjection) || state.RouteKnowledgeBalance != 37 {
					t.Fatal("contradictory reward owner admitted", err)
				}
				var rows int
				if err := db.QueryRowContext(ctx, `SELECT (SELECT count(*) FROM exit_knowledge_projection_events)+
					(SELECT count(*) FROM route_hint_projection_events)+(SELECT count(*) FROM founder_route_state)`).Scan(&rows); err != nil || rows != 0 {
					t.Fatalf("invalid source left projection rows=%d err=%v", rows, err)
				}
				return
			}
			if mode == "rollback" {
				if _, err := db.ExecContext(ctx, `CREATE FUNCTION knowledge_projection_fault() RETURNS trigger LANGUAGE plpgsql AS $$
					BEGIN RAISE EXCEPTION 'knowledge_projection_fixture'; END; $$;
					CREATE TRIGGER knowledge_projection_fault BEFORE UPDATE ON founder_route_state
					FOR EACH ROW EXECUTE FUNCTION knowledge_projection_fault()`); err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() {
					cleanup, stop := context.WithTimeout(context.Background(), 5*time.Second)
					defer stop()
					if _, err := db.ExecContext(cleanup, `DROP TRIGGER IF EXISTS knowledge_projection_fault ON founder_route_state;
						DROP FUNCTION IF EXISTS knowledge_projection_fault()`); err != nil {
						t.Error(err)
					}
				})
				state := &save.State{RouteKnowledgeBalance: 37}
				if err := projector.RepairFounder(ctx, founder, state); err == nil || state.RouteKnowledgeBalance != 37 {
					t.Fatal("failed projection consumed claims or changed caller state")
				}
				var rows int
				if err := db.QueryRowContext(ctx, `SELECT (SELECT count(*) FROM exit_knowledge_projection_events)+
					(SELECT count(*) FROM route_hint_projection_events)+(SELECT count(*) FROM founder_route_state)`).Scan(&rows); err != nil || rows != 0 {
					t.Fatalf("rollback rows=%d err=%v", rows, err)
				}
				if _, err := db.ExecContext(ctx, `DROP TRIGGER knowledge_projection_fault ON founder_route_state; DROP FUNCTION knowledge_projection_fault()`); err != nil {
					t.Fatal(err)
				}
			}
			if mode == "concurrent" {
				start := make(chan struct{})
				failures := make(chan error, 8)
				var workers sync.WaitGroup
				for index := range 8 {
					workers.Add(1)
					go func() {
						defer workers.Done()
						<-start
						if index%2 == 0 {
							failures <- projector.Project(ctx, []save.EventRecord{hint, advance})
						} else {
							state := &save.State{}
							err := projector.RepairFounder(ctx, founder, state)
							if err == nil && state.RouteKnowledgeBalance != want {
								err = fmt.Errorf("concurrent repaired balance=%d", state.RouteKnowledgeBalance)
							}
							failures <- err
						}
					}()
				}
				close(start)
				workers.Wait()
				close(failures)
				for err := range failures {
					if err != nil {
						t.Fatal(err)
					}
				}
			}
			state := &save.State{}
			if err := projector.RepairFounder(ctx, founder, state); err != nil || state.RouteKnowledgeBalance != want {
				t.Fatalf("repaired=%d want=%d err=%v", state.RouteKnowledgeBalance, want, err)
			}
			// Delivery after read-repair, including a reversed separate callback,
			// must not credit or debit the source facts again.
			for _, event := range []save.EventRecord{hint, advance, advance, hint} {
				if err := projector.Project(ctx, []save.EventRecord{event}); err != nil {
					t.Fatal(err)
				}
			}
			bad := advance
			bad.Payload = json.RawMessage(`{"route_knowledge":999}`)
			if err := projector.Project(ctx, []save.EventRecord{bad}); !errors.Is(err, ErrProjection) {
				t.Fatal("unbound Exit payload admitted", err)
			}
			assertKnowledgeState(t, ctx, db, founder, want, 0)
			if _, err := db.ExecContext(ctx, `DELETE FROM founder_route_state WHERE founder_id=$1`, founder); err != nil {
				t.Fatal(err)
			}
			if err := projector.RepairFounder(ctx, founder, state); err != nil || state.RouteKnowledgeBalance != want {
				t.Fatalf("rebuilt=%d want=%d err=%v", state.RouteKnowledgeBalance, want, err)
			}
			var exits, hints int
			if err := db.QueryRowContext(ctx, `SELECT (SELECT count(*) FROM exit_knowledge_projection_events WHERE founder_id=$1),
				(SELECT count(*) FROM route_hint_projection_events WHERE founder_id=$1)`, founder).Scan(&exits, &hints); err != nil || exits != 1 || hints != 1 {
				t.Fatalf("exit claims=%d hint claims=%d err=%v", exits, hints, err)
			}
		})
	}
}

func TestExitKnowledgeRepaysRegistryDebtIntegration(t *testing.T) {
	db, projector, founder, company, founderStream := exitKnowledgeFixture(t)
	ctx := t.Context()
	const hash = "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	at := time.Date(2026, 11, 5, 13, 0, 0, 0, time.UTC)
	execution := insertExecution(t, ctx, db, "01986666-cc00-7000-8000-000000000011", company,
		founder, 1, 1, "route.nonprofit_wrapper_zip", "gate.t4_to_t5", hash, at)
	if err := projector.Project(ctx, []save.EventRecord{execution}); err != nil {
		t.Fatal(err)
	}
	advance := func(run int64) save.EventRecord {
		t.Helper()
		stamp := at.Add(time.Duration(run) * time.Second)
		return insertKnowledgeFact(t, ctx, db, founderStream, founder, hash, stamp, save.EventFounderAdvanced,
			map[string]any{"founder_id": founder, "run_id": map[string]any{"company_stream_id": company, "run_seq": run},
				"exit_type": "collapse", "reputation_delta": 0, "route_knowledge": 25, "occurred_at_ms": stamp.UnixMilli()})
	}
	if err := projector.Project(ctx, []save.EventRecord{advance(1)}); err != nil {
		t.Fatal(err)
	}
	assertKnowledgeState(t, ctx, db, founder, 150, 0)
	var routeIDs []string
	catalog, _ := projector.catalogs.ResolveRoutes(hash)
	for _, gate := range catalog.Gates() {
		for _, route := range gate.Routes {
			routeIDs = append(routeIDs, route.RouteID)
		}
	}
	if len(routeIDs) < 3 {
		t.Fatal("debt fixture needs three different hint targets")
	}
	for index, routeID := range routeIDs[:3] {
		hint := insertKnowledgeFact(t, ctx, db, founderStream, founder, hash, at.Add(time.Duration(index+2)*time.Second),
			save.EventRouteHintPurchased, map[string]any{"route_id": routeID, "cost": 50})
		if err := projector.Project(ctx, []save.EventRecord{hint}); err != nil {
			t.Fatal(err)
		}
	}
	assertKnowledgeState(t, ctx, db, founder, 0, 0)
	// An earlier executor displaces the spent Registry bonus; earned Exit credit
	// must repay that debt before making any new hint balance spendable.
	const earlierFounder = "01986666-cc00-7000-8000-000000000002"
	earlierCompany := insertStream(t, ctx, db, earlierFounder, "company")
	earlier := insertExecution(t, ctx, db, "01986666-cc00-7000-8000-000000000012", earlierCompany,
		earlierFounder, 1, 1, "route.nonprofit_wrapper_zip", "gate.t4_to_t5", hash, at.Add(-time.Second))
	if err := projector.Project(ctx, []save.EventRecord{earlier}); err != nil {
		t.Fatal(err)
	}
	assertKnowledgeState(t, ctx, db, founder, 0, 100)
	for run := int64(2); run <= 6; run++ {
		event := advance(run)
		if err := projector.Project(ctx, []save.EventRecord{event, event}); err != nil {
			t.Fatal(err)
		}
		net := 25*(run-1) - 100
		assertKnowledgeState(t, ctx, db, founder, max(net, 0), max(-net, 0))
	}
	if _, err := db.ExecContext(ctx, `DELETE FROM founder_route_state WHERE founder_id=$1`, founder); err != nil {
		t.Fatal(err)
	}
	state := &save.State{}
	if err := projector.RepairFounder(ctx, founder, state); err != nil || state.RouteKnowledgeBalance != 25 {
		t.Fatalf("debt rebuild balance=%d err=%v", state.RouteKnowledgeBalance, err)
	}
	assertKnowledgeState(t, ctx, db, founder, 25, 0)
}

func TestExitKnowledgeProjectionCapacityIntegration(t *testing.T) {
	db, projector, founder, company, stream := exitKnowledgeFixture(t)
	ctx := t.Context()
	const hash = "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	at := time.Date(2026, 11, 5, 12, 0, 0, 0, time.UTC)
	advance := func(run, amount int64) save.EventRecord {
		return insertKnowledgeFact(t, ctx, db, stream, founder, hash, at, save.EventFounderAdvanced,
			map[string]any{"founder_id": founder, "run_id": map[string]any{"company_stream_id": company, "run_seq": run},
				"exit_type": "collapse", "reputation_delta": 0, "route_knowledge": amount, "occurred_at_ms": at.UnixMilli()})
	}
	if err := projector.Project(ctx, []save.EventRecord{advance(1, decimal.MaxExactInteger)}); err != nil {
		t.Fatal(err)
	}
	assertKnowledgeState(t, ctx, db, founder, decimal.MaxExactInteger, 0)
	overflow := advance(2, 1)
	if err := projector.Project(ctx, []save.EventRecord{overflow}); !errors.Is(err, ErrProjection) {
		t.Fatal("out-of-domain balance admitted", err)
	}
	assertKnowledgeState(t, ctx, db, founder, decimal.MaxExactInteger, 0)
	var claims int
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM exit_knowledge_projection_events WHERE founder_id=$1`, founder).Scan(&claims); err != nil || claims != 1 {
		t.Fatalf("overflow committed claims=%d err=%v", claims, err)
	}
	if _, err := db.ExecContext(ctx, `DELETE FROM founder_route_state WHERE founder_id=$1`, founder); err != nil {
		t.Fatal(err)
	}
	state := &save.State{RouteKnowledgeBalance: 37}
	if err := projector.RepairFounder(ctx, founder, state); !errors.Is(err, ErrProjection) || state.RouteKnowledgeBalance != 37 {
		t.Fatal("overflow rebuild changed caller state", err)
	}
	var rows int
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM founder_route_state WHERE founder_id=$1`, founder).Scan(&rows); err != nil || rows != 0 {
		t.Fatalf("overflow rebuild committed cache rows=%d err=%v", rows, err)
	}
}

func exitKnowledgeFixture(t *testing.T) (*sql.DB, *Projector, string, string, string) {
	t.Helper()
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}
	db, err := save.OpenPostgres(t.Context(), databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	if err := save.Migrate(t.Context(), db); err != nil {
		t.Fatal(err)
	}
	truncateRouteProjection(t, t.Context(), db)
	data, err := os.ReadFile("../../balance/routes/phase0.json")
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := routes.LoadCatalog(data)
	if err != nil {
		t.Fatal(err)
	}
	projector, err := New(db, testCatalogs{"sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb": catalog})
	if err != nil {
		t.Fatal(err)
	}
	const founder = "01986666-cc00-7000-8000-000000000001"
	return db, projector, founder, insertStream(t, t.Context(), db, founder, "company"), insertStream(t, t.Context(), db, founder, "founder")
}

func insertKnowledgeFact(t *testing.T, ctx context.Context, db *sql.DB, stream, founder, hash string, at time.Time, kind save.EventKind, value any) save.EventRecord {
	t.Helper()
	payload, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	var id string
	if err := db.QueryRowContext(ctx, `INSERT INTO events(stream_id,revision,schema_version,kind,constants_hash,occurred_at,payload)
		VALUES($1,2,1,$2,$3,$4,$5) RETURNING event_id`, stream, kind, hash, at, payload).Scan(&id); err != nil {
		t.Fatal(err)
	}
	return save.EventRecord{EventID: id, StreamID: stream, OwnerID: founder, Revision: 2, Kind: kind, ConstantsHash: hash, OccurredAt: at, Payload: payload}
}

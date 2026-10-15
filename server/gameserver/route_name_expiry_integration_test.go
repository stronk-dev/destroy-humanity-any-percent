package gameserver

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"cloud-clicker/server/operations"
	"cloud-clicker/server/routeprojection"
	"cloud-clicker/server/save"

	"github.com/jackc/pgx/v5/pgconn"
)

const expiryRouteID = "route.nonprofit_wrapper_zip"

// The discovery is a controlled event fixture fed through the real projector,
// not a claim that current-content players can naturally reach this late gate.
func seedRouteExpiryReservation(t *testing.T, ctx context.Context, db *sql.DB, composition *Composition, deadline time.Time) {
	t.Helper()
	const founder = "01986666-0000-4000-8000-000000000411"
	const eventID = "01986666-0000-4000-8000-000000000412"
	const intentID = "01986666-0000-7000-8000-000000000413"
	var stream string
	if err := db.QueryRowContext(ctx, `INSERT INTO save_streams(owner_kind,owner_id,scope) VALUES('founder',$1,'company') RETURNING id`, founder).Scan(&stream); err != nil {
		t.Fatal(err)
	}
	payload, err := json.Marshal(map[string]any{
		"route_id": expiryRouteID, "gate_id": "gate.t4_to_t5", "founder_id": founder,
		"run_id": map[string]any{"company_stream_id": stream, "run_seq": 1},
	})
	if err != nil {
		t.Fatal(err)
	}
	occurred := deadline.Add(-72 * time.Hour)
	if _, err := db.ExecContext(ctx, `INSERT INTO events(event_id,stream_id,revision,schema_version,kind,intent_id,constants_hash,occurred_at,payload)
		VALUES($1,$2,1,1,'route_executed',$3,$4,$5,$6)`, eventID, stream, intentID, composition.CurrentHash, occurred, payload); err != nil {
		t.Fatal(err)
	}
	projector, err := routeprojection.New(db, composition.Catalogs)
	if err != nil {
		t.Fatal(err)
	}
	if err := projector.Project(ctx, []save.EventRecord{{
		EventID: eventID, StreamID: stream, OwnerID: founder, Revision: 1, Kind: save.EventRouteExecuted,
		IntentID: intentID, ConstantsHash: composition.CurrentHash, OccurredAt: occurred, Payload: payload,
	}}); err != nil {
		t.Fatal(err)
	}
}

func routeExpiryHistory(t *testing.T, ctx context.Context, db *sql.DB) string {
	t.Helper()
	var history string
	if err := db.QueryRowContext(ctx, `SELECT jsonb_build_object(
		'registry',(SELECT jsonb_agg(to_jsonb(r)-'name_state' ORDER BY route_id) FROM registry_routes r),
		'events',(SELECT jsonb_agg(to_jsonb(e) ORDER BY event_id) FROM events e),
		'projection',(SELECT jsonb_agg(to_jsonb(e) ORDER BY event_id) FROM route_projection_events e),
		'executions',(SELECT jsonb_agg(to_jsonb(e) ORDER BY founder_id,route_id) FROM founder_route_executions e),
		'knowledge',(SELECT jsonb_agg(to_jsonb(e) ORDER BY founder_id) FROM founder_route_state e))::text`).Scan(&history); err != nil {
		t.Fatal(err)
	}
	return history
}

func requireRouteExpiryState(t *testing.T, ctx context.Context, db *sql.DB, server *Server, want string) {
	t.Helper()
	var state string
	if err := db.QueryRowContext(ctx, `SELECT name_state FROM registry_routes WHERE route_id=$1`, expiryRouteID).Scan(&state); err != nil || state != want {
		t.Fatalf("stored naming state=%q want=%q err=%v", state, want, err)
	}
	response := httptest.NewRecorder()
	server.Handler().ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/public/v1/registry/routes", nil))
	var page struct {
		Items []struct {
			RouteID      string `json:"route_id"`
			NamingStatus string `json:"naming_status"`
		} `json:"items"`
	}
	if response.Code != http.StatusOK || json.Unmarshal(response.Body.Bytes(), &page) != nil || len(page.Items) != 1 ||
		page.Items[0].RouteID != expiryRouteID || page.Items[0].NamingStatus != want {
		t.Fatalf("mounted public naming status did not follow stored lifecycle: status=%d body=%s", response.Code, response.Body.String())
	}
}

func requireRouteExpiryMetric(t *testing.T, registry *operations.Registry, result string, count int) {
	t.Helper()
	response := httptest.NewRecorder()
	registry.Handler().ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	want := fmt.Sprintf("cloud_clicker_job_runs_total{job_name=\"route_name_expiry\",result=%q} %d\n", result, count)
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), want) {
		t.Fatalf("route expiry observation missing: %s", want)
	}
}

func TestComposedRouteNameExpiryStartupAndFailureIntegration(t *testing.T) {
	for _, subject := range []string{"expiry", "sql_failure"} {
		t.Run(subject, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			db := newCredentialCleanupDatabase(t, ctx)
			defer db.Close()
			defer cancel() // Stop jobs before closing this disposable database.
			clock := &mutableClock{now: time.Date(2026, 10, 15, 12, 0, 0, 0, time.UTC)}
			var logs cleanupRuntimeLog
			registry, composition := newCredentialCleanupComposition(t, ctx, db, clock, &logs)
			seedRouteExpiryReservation(t, ctx, db, composition, clock.Time())
			requireRouteExpiryState(t, ctx, db, composition.Server, "reserved")
			before := routeExpiryHistory(t, ctx, db)
			if subject == "sql_failure" {
				if _, err := db.ExecContext(ctx, `CREATE FUNCTION route_expiry_fault() RETURNS trigger LANGUAGE plpgsql AS $$
					BEGIN RAISE EXCEPTION 'route_expiry_fixture' USING ERRCODE='P0001'; END; $$;
					CREATE TRIGGER route_expiry_fault BEFORE UPDATE ON registry_routes FOR EACH ROW
					WHEN (OLD.name_state='reserved' AND NEW.name_state='house') EXECUTE FUNCTION route_expiry_fault()`); err != nil {
					t.Fatal(err)
				}
			}
			err := composition.Server.Start(ctx)
			if subject == "expiry" {
				if err != nil {
					t.Fatal(err)
				}
				requireRouteExpiryState(t, ctx, db, composition.Server, "house")
				requireRouteExpiryMetric(t, registry, "success", 1)
				requireRouteExpiryMetric(t, registry, "failure", 0)
				drainCredentialCleanupComposition(t, composition, clock.Time())
			} else {
				var fault *pgconn.PgError
				if !errors.As(err, &fault) || fault.Code != "P0001" || fault.Message != "route_expiry_fixture" {
					t.Fatalf("startup did not propagate actual naming-expiry SQL failure: %v", err)
				}
				requireRouteExpiryState(t, ctx, db, composition.Server, "reserved")
				requireRouteExpiryMetric(t, registry, "failure", 1)
				requireRouteExpiryMetric(t, registry, "success", 0)
				ready := httptest.NewRecorder()
				composition.Server.Handler().ServeHTTP(ready, httptest.NewRequest(http.MethodGet, "/readyz", nil))
				if ready.Code != http.StatusServiceUnavailable {
					t.Fatal("failed naming expiry admitted readiness")
				}
				message := logs.String()
				if !strings.Contains(message, `"job_name":"route_name_expiry"`) || strings.Contains(message, "route_expiry_fixture") {
					t.Fatal("expiry failure log absent or includes private SQL detail")
				}
				row, _ := json.Marshal(map[string]string{"CONTAINER_TAG": "cloud-clicker-gameserver", "__CURSOR": "route-expiry", "__REALTIME_TIMESTAMP": "1", "MESSAGE": strings.TrimSpace(message)})
				if cursor, err := operations.CleanupFailureFromJournal(bytes.NewReader(row)); err != nil || cursor == "" {
					t.Fatalf("expiry failure lost durable cleanup observation: %v", err)
				}
				if _, err := db.ExecContext(ctx, "DROP TRIGGER route_expiry_fault ON registry_routes"); err != nil {
					t.Fatal(err)
				}
				var recoveredLog cleanupRuntimeLog
				recoveredRegistry, recovered := newCredentialCleanupComposition(t, ctx, db, clock, &recoveredLog)
				if err := recovered.Server.Start(ctx); err != nil {
					t.Fatal(err)
				}
				requireRouteExpiryState(t, ctx, db, recovered.Server, "house")
				requireRouteExpiryMetric(t, recoveredRegistry, "success", 1)
				drainCredentialCleanupComposition(t, recovered, clock.Time())
			}
			if routeExpiryHistory(t, ctx, db) != before {
				t.Fatal("composed expiry changed name/credit/Knowledge or immutable execution history")
			}
		})
	}
}

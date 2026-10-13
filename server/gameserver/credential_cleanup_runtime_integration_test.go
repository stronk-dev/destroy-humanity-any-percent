package gameserver

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
)

// Timed manual acceptance, deliberately outside push CI. Neither the actual
// job interval nor its callback is replaced. The spy only records the returned
// error/context and returns that exact error unchanged.
func TestComposedCredentialCleanupRuntimeAndDrainIntegration(t *testing.T) {
	if os.Getenv("CLOUD_CLICKER_CLEANUP_RUNTIME_TEST") != "1" {
		t.Skip("real one-minute timer; use make test-deployment-cleanup-runtime")
	}
	for _, subject := range []string{"runtime_sql_failure", "orderly_drain"} {
		t.Run(subject, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 95*time.Second)
			defer cancel()
			database := newCredentialCleanupDatabase(t, ctx)
			defer database.Close()
			// Cancellation must begin before the database is closed on failures.
			defer cancel()
			clock := &mutableClock{now: time.Date(2026, 10, 13, 0, 0, 0, 0, time.UTC)}
			var logs cleanupRuntimeLog
			registry, composition := newCredentialCleanupComposition(t, ctx, database, clock, &logs)
			if len(composition.Server.jobs) != 6 {
				t.Fatal("composed job population changed; observation must be reconciled")
			}
			job, ok := composition.Server.jobs[5].(*PeriodicJob)
			if !ok || job.interval != time.Minute {
				t.Fatal("test must retain the actual composed one-minute cleanup job")
			}
			type outcome struct {
				err      error
				canceled bool
			}
			observed := make(chan outcome, 3)
			original := job.run
			job.run = func(ctx context.Context) error {
				err := original(ctx)
				select {
				case observed <- outcome{err, ctx.Err() != nil}:
				default:
					t.Error("cleanup callback observation overflow")
				}
				return err
			}
			httpServer := httptest.NewServer(composition.Server.Handler())
			defer httpServer.Close()
			response := compositionRequest(t, httpServer.Client(), http.MethodPost, httpServer.URL+"/api/v1/bootstrap", "",
				fmt.Sprintf(`{"idempotency_key":%q}`, strings.Repeat("ab", 32)))
			status := response.StatusCode
			_, readErr := io.Copy(io.Discard, response.Body)
			_ = response.Body.Close()
			if status != http.StatusCreated || readErr != nil {
				t.Fatal("actual bootstrap setup failed")
			}
			if err := composition.Server.Start(ctx); err != nil {
				t.Fatal(err)
			}
			prime := <-observed
			if prime.err != nil || prime.canceled {
				t.Fatal("healthy startup callback was not observed")
			}
			requireCleanupRuntimeHTTP(t, httpServer, "/readyz", http.StatusNoContent)
			requireCredentialCleanupMetric(t, registry, "success", 1)
			requireCredentialCleanupMetric(t, registry, "failure", 0)
			clock.Set(clock.Time().Add(31 * 24 * time.Hour))
			if counts := credentialCleanupCounts(t, ctx, database); counts != [4]int{1, 1, 1, 1} {
				t.Fatal("expired actual credential population missing")
			}
			if subject == "runtime_sql_failure" {
				if _, err := database.ExecContext(ctx, `CREATE FUNCTION operations_runtime_cleanup_fault() RETURNS trigger LANGUAGE plpgsql AS $$
					BEGIN RAISE EXCEPTION 'operations_runtime_cleanup_fixture' USING ERRCODE='P0001'; END; $$;
					CREATE TRIGGER operations_runtime_cleanup_fault BEFORE DELETE ON sessions FOR EACH ROW
					EXECUTE FUNCTION operations_runtime_cleanup_fault()`); err != nil {
					t.Fatal(err)
				}
				select {
				case err := <-composition.Server.Failures():
					var postgresError *pgconn.PgError
					if !errors.As(err, &postgresError) || postgresError.Code != "P0001" || postgresError.Message != "operations_runtime_cleanup_fixture" {
						t.Fatal("runtime did not report the actual cleanup SQL failure")
					}
				case <-ctx.Done():
					t.Fatal("actual one-minute cleanup failure did not reach runtime supervision")
				}
				result := <-observed
				if result.err == nil || result.canceled {
					t.Fatal("runtime SQL failure was mistaken for orderly cancellation")
				}
				requireCleanupRuntimeHTTP(t, httpServer, "/readyz", http.StatusServiceUnavailable)
				requireCredentialCleanupMetric(t, registry, "failure", 1)
				if !strings.Contains(logs.String(), `"job_name":"credential_cleanup"`) || strings.Contains(logs.String(), "operations_runtime_cleanup_fixture") {
					t.Fatal("runtime failure lost bounded observation")
				}
				drainCtx, stop := context.WithTimeout(context.Background(), composition.Node.DrainTimeout())
				defer stop()
				// The failure channel above owns the job error. Drain owns the
				// remaining shutdown, not replaying an already reported error.
				if err := composition.Server.Drain(drainCtx, clock.Time()); err != nil {
					t.Fatalf("shutdown after reported runtime failure: %v", err)
				}
				if counts := credentialCleanupCounts(t, ctx, database); counts != [4]int{1, 1, 1, 1} {
					t.Fatal("failed session prune committed credentials")
				}
				if _, err := database.ExecContext(ctx, "DROP TRIGGER operations_runtime_cleanup_fault ON sessions"); err != nil {
					t.Fatal(err)
				}
				var recoveredLog cleanupRuntimeLog
				recoveredRegistry, recovered := newCredentialCleanupComposition(t, ctx, database, clock, &recoveredLog)
				if err := recovered.Server.Start(ctx); err != nil {
					t.Fatalf("startup after runtime SQL repair: %v", err)
				}
				requireCredentialCleanupMetric(t, recoveredRegistry, "success", 1)
				requireCredentialCleanupMetric(t, recoveredRegistry, "failure", 0)
				if counts := credentialCleanupCounts(t, ctx, database); counts != [4]int{} || recoveredLog.String() != "" {
					t.Fatal("runtime recovery did not collect expired credentials without an alarm")
				}
				drainCredentialCleanupComposition(t, recovered, clock.Time())
				return
			}

			// Hold only this disposable database's session table. The actual
			// scheduled SQL must enter its lock wait before requesting drain.
			lock, err := database.BeginTx(ctx, nil)
			if err != nil {
				t.Fatal(err)
			}
			defer lock.Rollback()
			if _, err := lock.ExecContext(ctx, "LOCK TABLE sessions IN ACCESS EXCLUSIVE MODE"); err != nil {
				t.Fatal(err)
			}
			ticker := time.NewTicker(100 * time.Millisecond)
			defer ticker.Stop()
			for {
				var waiting int
				if err := database.QueryRowContext(ctx, `SELECT count(*) FROM pg_stat_activity
					WHERE datname=current_database() AND wait_event_type='Lock' AND query LIKE '%DELETE FROM sessions token%'`).Scan(&waiting); err != nil {
					t.Fatal(err)
				}
				if waiting == 1 {
					break
				}
				if waiting > 1 {
					t.Fatal("ambiguous actual cleanup lock population")
				}
				select {
				case <-ctx.Done():
					t.Fatal("actual one-minute cleanup never entered the SQL lock wait")
				case <-ticker.C:
				}
			}
			drainCtx, stop := context.WithTimeout(context.Background(), composition.Node.DrainTimeout())
			defer stop()
			if err := composition.Server.Drain(drainCtx, clock.Time()); err != nil {
				t.Fatal("orderly drain failed")
			}
			result := <-observed
			t.Logf("actual scheduled SQL canceled by drain: context_ended=%t cancellation_error=%t", result.canceled, errors.Is(result.err, context.Canceled))
			if !result.canceled || !errors.Is(result.err, context.Canceled) {
				t.Fatal("drain did not cancel the actual cleanup SQL")
			}
			requireCredentialCleanupMetric(t, registry, "failure", 0)
			requireCredentialCleanupMetric(t, registry, "success", 1)
			requireCredentialCleanupMetric(t, registry, "canceled", 1)
			requireCleanupRuntimeHTTP(t, httpServer, "/readyz", http.StatusServiceUnavailable)
			select {
			case <-composition.Server.Failures():
				t.Fatal("orderly cancellation reached fatal runtime supervision")
			default:
			}
			if strings.Contains(logs.String(), `"job_name":"credential_cleanup"`) {
				t.Fatal("orderly canceled cleanup emitted a failure alarm")
			}
			if err := lock.Rollback(); err != nil {
				t.Fatal(err)
			}
			if counts := credentialCleanupCounts(t, ctx, database); counts != [4]int{1, 1, 1, 1} {
				t.Fatal("canceled cleanup committed partial credentials")
			}
		})
	}
}

type cleanupRuntimeLog struct {
	mu   sync.Mutex
	data bytes.Buffer
}

func (log *cleanupRuntimeLog) Write(data []byte) (int, error) {
	log.mu.Lock()
	defer log.mu.Unlock()
	return log.data.Write(data)
}
func (log *cleanupRuntimeLog) String() string {
	log.mu.Lock()
	defer log.mu.Unlock()
	return log.data.String()
}

func requireCleanupRuntimeHTTP(t *testing.T, server *httptest.Server, path string, want int) {
	t.Helper()
	response, err := server.Client().Get(server.URL + path)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != want {
		t.Fatalf("actual HTTP readiness: got=%d want=%d", response.StatusCode, want)
	}
}

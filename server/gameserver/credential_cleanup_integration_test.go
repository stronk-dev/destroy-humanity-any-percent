package gameserver

import (
	"bytes"
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"cloud-clicker/server/account"
	"cloud-clicker/server/operations"
	"cloud-clicker/server/save"

	"github.com/jackc/pgx/v5/pgconn"
)

// This tests the real composed cleanup callback and its fatal startup policy.
// Reading a registry directly here is not a Prometheus scrape or alert proof.
func TestComposedCredentialCleanupFailureAndRecoveryIntegration(t *testing.T) {
	for _, subject := range []string{"healthy", "sessions", "bootstrap_receipts"} {
		t.Run(subject, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			database := newCredentialCleanupDatabase(t, ctx)
			defer database.Close()
			clock := &mutableClock{now: time.Date(2026, 10, 13, 0, 0, 0, 0, time.UTC)}
			registry, composition := newCredentialCleanupComposition(t, ctx, database, clock)
			httpServer := httptest.NewServer(composition.Server.Handler())
			defer httpServer.Close()
			response := compositionRequest(t, httpServer.Client(), http.MethodPost, httpServer.URL+"/api/v1/bootstrap", "",
				fmt.Sprintf(`{"idempotency_key":%q}`, strings.Repeat("ab", 32)))
			status := response.StatusCode
			_, readErr := io.Copy(io.Discard, response.Body) // never log returned credentials
			_ = response.Body.Close()
			httpServer.Close()
			if status != http.StatusCreated || readErr != nil {
				t.Fatalf("actual bootstrap setup failed: status=%d error=%v", status, readErr)
			}
			clock.Set(clock.Time().Add(31 * 24 * time.Hour))
			wantLive := [4]int{1, 1, 1, 1}
			if counts := credentialCleanupCounts(t, ctx, database); counts != wantLive {
				t.Fatalf("actual expired credential population missing: %v", counts)
			}
			if subject == "healthy" {
				if err := composition.Server.Start(ctx); err != nil {
					t.Fatal(err)
				}
				requireCredentialCleanupMetric(t, registry, "success", 1)
				requireCredentialCleanupMetric(t, registry, "failure", 0)
				if counts := credentialCleanupCounts(t, ctx, database); counts != [4]int{} {
					t.Fatalf("healthy composed prime did not collect credentials: %v", counts)
				}
				drainCredentialCleanupComposition(t, composition, clock.Time())
				return
			}

			// Only this disposable database is faulted. The exception is raised by
			// the actual DELETE/UPDATE, not a replacement Go repository or job.
			if _, err := database.ExecContext(ctx, `CREATE FUNCTION operations_cleanup_fault() RETURNS trigger LANGUAGE plpgsql AS $$
				BEGIN RAISE EXCEPTION 'operations_cleanup_fixture' USING ERRCODE='P0001'; END; $$`); err != nil {
				t.Fatal(err)
			}
			trigger := `CREATE TRIGGER operations_cleanup_fault BEFORE DELETE ON sessions FOR EACH ROW EXECUTE FUNCTION operations_cleanup_fault()`
			if subject == "bootstrap_receipts" {
				trigger = `CREATE TRIGGER operations_cleanup_fault BEFORE UPDATE ON bootstrap_receipts FOR EACH ROW
					WHEN (OLD.tombstoned_at IS NULL AND NEW.tombstoned_at IS NOT NULL) EXECUTE FUNCTION operations_cleanup_fault()`
			}
			if _, err := database.ExecContext(ctx, trigger); err != nil {
				t.Fatal(err)
			}
			requireCredentialCleanupMetric(t, registry, "failure", 0)
			err := composition.Server.Start(ctx)
			var postgresError *pgconn.PgError
			if !errors.As(err, &postgresError) || postgresError.Code != "P0001" || postgresError.Message != "operations_cleanup_fixture" {
				t.Fatalf("startup did not propagate the actual cleanup SQL fault: %v", err)
			}
			requireCredentialCleanupMetric(t, registry, "failure", 1)
			requireCredentialCleanupMetric(t, registry, "success", 0)
			ready := httptest.NewRecorder()
			composition.Server.Handler().ServeHTTP(ready, httptest.NewRequest(http.MethodGet, "/readyz", nil))
			if ready.Code != http.StatusServiceUnavailable {
				t.Fatal("failed cleanup prime admitted readiness")
			}
			if subject == "bootstrap_receipts" {
				// The session prune commits before the separate bootstrap prune.
				// Do not misrepresent this callback as one atomic transaction.
				wantLive = [4]int{0, 0, 0, 1}
			}
			if counts := credentialCleanupCounts(t, ctx, database); counts != wantLive {
				t.Fatalf("cleanup failure violated its actual transaction boundary: got=%v want=%v", counts, wantLive)
			}
			freshRegistry, err := operations.NewRegistry(database)
			if err != nil {
				t.Fatal(err)
			}
			requireCredentialCleanupMetric(t, freshRegistry, "failure", 0)
			t.Log("actual composed cleanup failed before readiness; the observed in-process failure counter resets to zero in a fresh registry")

			if _, err := database.ExecContext(ctx, "DROP TRIGGER operations_cleanup_fault ON "+subject); err != nil {
				t.Fatal(err)
			}
			recoveredRegistry, recovered := newCredentialCleanupComposition(t, ctx, database, clock)
			if err := recovered.Server.Start(ctx); err != nil {
				t.Fatalf("removing the actual SQL fault did not restore startup: %v", err)
			}
			requireCredentialCleanupMetric(t, recoveredRegistry, "success", 1)
			requireCredentialCleanupMetric(t, recoveredRegistry, "failure", 0)
			if counts := credentialCleanupCounts(t, ctx, database); counts != [4]int{} {
				t.Fatalf("real startup recovery left expired credentials: %v", counts)
			}
			drainCredentialCleanupComposition(t, recovered, clock.Time())
		})
	}
}

func newCredentialCleanupDatabase(t *testing.T, ctx context.Context) *sql.DB {
	t.Helper()
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}
	admin, err := save.OpenPostgres(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	name := "operations_cleanup_" + strings.ToLower(rand.Text())
	if len(name) > 63 || !regexp.MustCompile(`^operations_cleanup_[a-z0-9]+$`).MatchString(name) {
		_ = admin.Close()
		t.Fatal("invalid generated fixture database name")
	}
	if _, err := admin.ExecContext(ctx, "CREATE DATABASE "+name+" TEMPLATE template0"); err != nil {
		_ = admin.Close()
		t.Fatal(err)
	}
	t.Cleanup(func() {
		cleanup, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if _, err := admin.ExecContext(cleanup, "DROP DATABASE "+name); err != nil {
			t.Error(err)
		}
		_ = admin.Close()
	})
	parsed, err := url.Parse(databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	parsed.Path = "/" + name
	database, err := save.OpenPostgres(ctx, parsed.String())
	if err != nil {
		t.Fatal(err)
	}
	if err := save.Migrate(ctx, database); err != nil {
		_ = database.Close()
		t.Fatal(err)
	}
	return database
}

func newCredentialCleanupComposition(t *testing.T, ctx context.Context, database *sql.DB, clock *mutableClock) (*operations.Registry, *Composition) {
	t.Helper()
	registry, err := operations.NewRegistry(database)
	if err != nil {
		t.Fatal(err)
	}
	composition, err := Compose(ctx, CompositionConfig{
		PublicCursorKeys: testPublicCursorKeys(), DB: database, RepositoryRoot: filepathRoot(t),
		ServerID: "018f0000-0000-4000-8000-000000000305", ActivityBracket: "activity.standard", Clock: clock.Time,
		SigningKeys:   account.SigningKeys{CurrentID: "cleanup-fixture", Current: bytes.Repeat([]byte{0x71}, 32)},
		BootstrapKeys: account.BootstrapReceiptKeys{CurrentID: "cleanup-fixture", Current: bytes.Repeat([]byte{0x72}, 32)},
		Operations:    registry,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := composition.Node.Shutdown(shutdown); err != nil {
			t.Error(err)
		}
	})
	return registry, composition
}

func credentialCleanupCounts(t *testing.T, ctx context.Context, database *sql.DB) [4]int {
	t.Helper()
	var result [4]int
	if err := database.QueryRowContext(ctx, `SELECT (SELECT count(*) FROM sessions),
		(SELECT count(*) FROM access_tokens),(SELECT count(*) FROM session_families),
		(SELECT count(*) FROM bootstrap_receipts WHERE tombstoned_at IS NULL)`).Scan(&result[0], &result[1], &result[2], &result[3]); err != nil {
		t.Fatal(err)
	}
	return result
}

func requireCredentialCleanupMetric(t *testing.T, registry *operations.Registry, result string, want int) {
	t.Helper()
	response := httptest.NewRecorder()
	registry.Handler().ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	metric := fmt.Sprintf("cloud_clicker_job_runs_total{job_name=\"credential_cleanup\",result=%q} %d\n", result, want)
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), metric) {
		t.Fatalf("actual composed cleanup result missing: status=%d want=%s", response.Code, metric)
	}
}

func drainCredentialCleanupComposition(t *testing.T, composition *Composition, now time.Time) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := composition.Server.Drain(ctx, now); err != nil {
		t.Fatal(err)
	}
}

package operations

import (
	"context"
	"crypto/rand"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"cloud-clicker/server/save"
)

func TestDatabaseCollectorIntegrationUsesCurrentPostgresSchema(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	database, err := save.OpenPostgres(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	if err := save.Migrate(ctx, database); err != nil {
		t.Fatal(err)
	}
	registry, err := NewRegistry(database)
	if err != nil {
		t.Fatal(err)
	}
	response := httptest.NewRecorder()
	registry.Handler().ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("metrics status=%d body=%s", response.Code, response.Body.String())
	}
	text := response.Body.String()
	for _, want := range []string{
		"cloud_clicker_postgres_reachable 1",
		"cloud_clicker_database_collection_success 1",
		"cloud_clicker_outbox_pending",
		`cloud_clicker_dead_letters{queue="transport"}`,
		`cloud_clicker_dead_letters{queue="verification"}`,
		`cloud_clicker_dead_letters{queue="verification_poison"}`,
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("current-schema metric missing %q in:\n%s", want, text)
		}
	}
}

func TestDatabaseCollectorIntegrationReportsBlockedQueryAndRecovers(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	database, err := save.OpenPostgres(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	if err := save.Migrate(ctx, database); err != nil {
		t.Fatal(err)
	}
	registry, err := NewRegistry(database)
	if err != nil {
		t.Fatal(err)
	}
	observe := func(wantCollection, wantPopulation bool) {
		t.Helper()
		response := httptest.NewRecorder()
		registry.Handler().ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/metrics", nil))
		if response.Code != http.StatusOK {
			t.Fatalf("metrics status=%d body=%s", response.Code, response.Body.String())
		}
		text := response.Body.String()
		collection := "cloud_clicker_database_collection_success 0\n"
		if wantCollection {
			collection = "cloud_clicker_database_collection_success 1\n"
		}
		if !strings.Contains(text, "cloud_clicker_postgres_reachable 1\n") || !strings.Contains(text, collection) {
			t.Fatalf("actual database health was hidden: %s", text)
		}
		for _, population := range []string{"cloud_clicker_outbox_pending", "cloud_clicker_dead_letters"} {
			if strings.Contains(text, "# TYPE "+population) != wantPopulation {
				t.Fatalf("collection success=%t exposed stale/absent %s: %s", wantCollection, population, text)
			}
		}
	}
	observe(true, true)
	lock, err := database.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer lock.Rollback()
	if _, err := lock.ExecContext(ctx, "LOCK TABLE verification_dead_letters IN ACCESS EXCLUSIVE MODE"); err != nil {
		t.Fatal(err)
	}
	observe(false, false)
	if err := lock.Rollback(); err != nil {
		t.Fatal(err)
	}
	observe(true, true)
}

func TestDatabaseCollectorIntegrationReportsConnectionRefusalAndRecovers(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	admin, err := save.OpenPostgres(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close()
	name := "operations_reachability_" + strings.ToLower(rand.Text())
	if len(name) > 63 || !operationsFaultDatabaseName.MatchString(name) {
		t.Fatal("invalid generated fixture database name")
	}
	if _, err := admin.ExecContext(ctx, "CREATE DATABASE "+name+" TEMPLATE template0"); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if _, err := admin.ExecContext(context.Background(), "DROP DATABASE "+name); err != nil {
			t.Error(err)
		}
	}()
	parsed, err := url.Parse(databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	parsed.Path = "/" + name
	database, err := save.OpenPostgres(ctx, parsed.String())
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	if err := save.Migrate(ctx, database); err != nil {
		t.Fatal(err)
	}
	registry, err := NewRegistry(database)
	if err != nil {
		t.Fatal(err)
	}
	observe := func(healthy bool) {
		t.Helper()
		response := httptest.NewRecorder()
		registry.Handler().ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/metrics", nil))
		value := "0\n"
		if healthy {
			value = "1\n"
		}
		text := response.Body.String()
		if response.Code != http.StatusOK ||
			!strings.Contains(text, "cloud_clicker_postgres_reachable "+value) ||
			!strings.Contains(text, "cloud_clicker_database_collection_success "+value) {
			t.Fatalf("actual connection state hidden: %s", text)
		}
		for _, population := range []string{"cloud_clicker_outbox_pending", "cloud_clicker_dead_letters"} {
			if strings.Contains(text, "# TYPE "+population) != healthy {
				t.Fatalf("unavailable population exposed or recovery missing: %s", text)
			}
		}
	}
	observe(true)
	restore := refuseOperationsDatabaseConnections(t, ctx, admin, name)
	// Restore before the test-owned database is closed/dropped, including when
	// an assertion fails. No existing test/application database is disabled.
	defer func() {
		if err := restore(); err != nil {
			t.Error(err)
		}
	}()
	observe(false)
	if err := restore(); err != nil {
		t.Fatal(err)
	}
	observe(true)
}

var operationsFaultDatabaseName = regexp.MustCompile(`^(cloud_clicker_operations_test|operations_reachability_[a-z0-9]+)$`)

func refuseOperationsDatabaseConnections(t *testing.T, ctx context.Context, admin *sql.DB, name string) func() error {
	t.Helper()
	var controlDatabase string
	if !operationsFaultDatabaseName.MatchString(name) ||
		admin.QueryRowContext(ctx, "SELECT current_database()").Scan(&controlDatabase) != nil || controlDatabase == name {
		t.Fatal("connection refusal requires a test-owned subject and separate control database")
	}
	if _, err := admin.ExecContext(ctx, "ALTER DATABASE "+name+" ALLOW_CONNECTIONS false"); err != nil {
		t.Fatal(err)
	}
	restore := func() error {
		restoreContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_, err := admin.ExecContext(restoreContext, "ALTER DATABASE "+name+" ALLOW_CONNECTIONS true")
		return err
	}
	// Recover before caller defers can close the admin if this helper fails.
	// Once returned, the caller owns restoration. Every terminated connection
	// is selected from this exact subject; the control database remains usable.
	returned := false
	defer func() {
		if !returned {
			if err := restore(); err != nil {
				t.Error(err)
			}
		}
	}()
	rows, err := admin.QueryContext(ctx, "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname=$1 AND pid<>pg_backend_pid()", name)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	terminated := 0
	for rows.Next() {
		var success bool
		if err := rows.Scan(&success); err != nil || !success {
			t.Fatalf("fixture connection termination failed: %v", err)
		}
		terminated++
	}
	if err := rows.Err(); err != nil || terminated == 0 {
		t.Fatalf("no live fixture connection was terminated: count=%d error=%v", terminated, err)
	}
	returned = true
	return restore
}

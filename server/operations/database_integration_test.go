package operations

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
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

package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"log/slog"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"cloud-clicker/server/save"

	"github.com/jackc/pgx/v5/pgconn"
)

// Exercise run's actual configuration, registry/logger wiring and startup.
// This is development-profile command logic, not an OCI process, journald
// driver, production secrets, listener traffic or release-host observation.
func TestCommandCleanupStartupFailureLogsBeforeReturningIntegration(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	admin, err := save.OpenPostgres(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close()
	name := "operations_command_cleanup_" + strings.ToLower(rand.Text())
	if len(name) > 63 || !regexp.MustCompile(`^operations_command_cleanup_[a-z0-9]+$`).MatchString(name) {
		t.Fatal("invalid generated command fixture name")
	}
	if _, err := admin.ExecContext(ctx, "CREATE DATABASE "+name+" TEMPLATE template0"); err != nil {
		t.Fatal(err)
	}
	defer func() {
		cleanup, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if _, err := admin.ExecContext(cleanup, "DROP DATABASE "+name); err != nil {
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
	// A statement trigger faults the real cleanup DELETE even with no user
	// rows. It cannot be mistaken for a setup/key/content failure.
	if _, err := database.ExecContext(ctx, `CREATE FUNCTION operations_command_cleanup_fault() RETURNS trigger LANGUAGE plpgsql AS $$
		BEGIN RAISE EXCEPTION 'operations_command_cleanup_fixture' USING ERRCODE='P0001'; END; $$;
		CREATE TRIGGER operations_command_cleanup_fault BEFORE DELETE ON sessions FOR EACH STATEMENT
		EXECUTE FUNCTION operations_command_cleanup_fault()`); err != nil {
		t.Fatal(err)
	}
	root, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	// Keep the valid address occupied. If startup incorrectly reaches listener
	// binding, it fails promptly instead of leaving this test serving forever.
	reserved, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer reserved.Close()
	for key, value := range map[string]string{
		"DATABASE_URL": parsed.String(), "CLOUD_CLICKER_DEPLOYMENT_MODE": "development",
		"CLOUD_CLICKER_REPOSITORY_ROOT": root, "CLOUD_CLICKER_SERVER_ID": "018f0000-0000-4000-8000-000000000306",
		"CLOUD_CLICKER_JWT_KEY":          base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{0x71}, 32)),
		"CLOUD_CLICKER_BOOTSTRAP_KEY":    base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{0x72}, 32)),
		"CLOUD_CLICKER_BOOTSTRAP_KEY_ID": "command-cleanup-fixture",
		"CLOUD_CLICKER_CURSOR_KEY":       base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{0x73}, 32)),
		"LISTEN_ADDR":                    reserved.Addr().String(),
	} {
		t.Setenv(key, value)
	}
	var logs bytes.Buffer
	err = run(slog.New(slog.NewJSONHandler(&logs, nil)))
	var postgresError *pgconn.PgError
	if !errors.As(err, &postgresError) || postgresError.Code != "P0001" || postgresError.Message != "operations_command_cleanup_fixture" {
		t.Fatal("command did not return the actual cleanup SQL failure")
	}
	message := logs.String()
	if strings.Count(message, `"event_class":"composed_job_failure"`) != 1 ||
		!strings.Contains(message, `"job_name":"credential_cleanup"`) || strings.Contains(message, "operations_command_cleanup_fixture") ||
		strings.Contains(message, parsed.String()) {
		t.Fatal("command startup failure lost bounded logger wiring or disclosed private error")
	}
}

package deploymentbackup

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
)

func TestDatabaseServiceFilesKeepPasswordOutOfEnvironmentAndArguments(t *testing.T) {
	workspace := t.TempDir()
	secret := filepath.Join(workspace, "database-url")
	password := "s3cr:et\\value"
	if err := os.WriteFile(secret, []byte("postgres://cloud_clicker:s3cr%3Aet%5Cvalue@postgres:5432/cloud_clicker?sslmode=disable\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	environment, err := databaseServiceEnvironment(secret, workspace)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(strings.Join(environment, "\n"), password) {
		t.Fatal("database password entered subprocess environment")
	}
	service, err := os.ReadFile(filepath.Join(workspace, "pg_service.conf"))
	if err != nil || strings.Contains(string(service), password) {
		t.Fatalf("service file exposed password: %v", err)
	}
	pass, err := os.ReadFile(filepath.Join(workspace, "pgpass"))
	if err != nil || !strings.Contains(string(pass), `s3cr\:et\\value`) {
		t.Fatalf("pgpass escaping failed: %q err=%v", pass, err)
	}
}

func TestDatabaseURLSecretOutsideComposeMustBeOwnerOnly(t *testing.T) {
	secret := filepath.Join(t.TempDir(), "database-url")
	if err := os.WriteFile(secret, []byte("postgres://user:password@postgres/database\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := databaseServiceEnvironment(secret, t.TempDir()); err == nil {
		t.Fatal("world-readable database URL accepted outside /run/secrets")
	}
}

func TestDatabaseServiceFilesMatchDriverEndpointOverrides(t *testing.T) {
	for _, test := range []struct {
		name  string
		query string
	}{
		{"ordinary", ""},
		{"host", "host=other_postgres"},
		{"port", "port=6543"},
		{"user", "user=other_user"},
		{"dbname", "dbname=other_database"},
		{"database alias", "database=other_database"},
		{"same database aliases", "database=other_database&dbname=other_database"},
		{"password", "password=other%3Apass%5Cword"},
		{"all", "host=other_postgres&port=6543&user=other_user&dbname=other_database&password=other%3Apass%5Cword"},
	} {
		t.Run(test.name, func(t *testing.T) {
			workspace := t.TempDir()
			secret := filepath.Join(t.TempDir(), "database-url")
			connection := "postgres://cloud_clicker:original_password@postgres:5432/cloud_clicker?sslmode=disable"
			if test.query != "" {
				connection += "&" + test.query
			}
			config, err := pgx.ParseConfig(connection)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(secret, []byte(connection+"\n"), 0o600); err != nil {
				t.Fatal(err)
			}
			environment, err := databaseServiceEnvironment(secret, workspace)
			if err != nil {
				t.Fatal(err)
			}
			service, err := os.ReadFile(filepath.Join(workspace, "pg_service.conf"))
			if err != nil {
				t.Fatal(err)
			}
			for _, field := range []string{"host=" + config.Host, "port=" + fmt.Sprint(config.Port), "user=" + config.User, "dbname=" + config.Database} {
				if !strings.Contains(string(service), "\n"+field+"\n") {
					t.Errorf("libpq endpoint differs from the driver: missing %q", field)
				}
			}
			pass, err := os.ReadFile(filepath.Join(workspace, "pgpass"))
			if err != nil {
				t.Fatal(err)
			}
			wantPass := strings.Join([]string{pgpassEscape(config.Host), fmt.Sprint(config.Port), pgpassEscape(config.Database), pgpassEscape(config.User), pgpassEscape(config.Password)}, ":") + "\n"
			if string(pass) != wantPass {
				t.Error("pass file does not bind the driver's endpoint and password")
			}
			if strings.Contains(string(service), config.Password) || strings.Contains(strings.Join(environment, "\n"), config.Password) {
				t.Error("password entered service text or subprocess environment")
			}
		})
	}
}

func TestDatabaseServiceFilesRejectUnsafeOrAmbiguousEndpointOverrides(t *testing.T) {
	for name, query := range map[string]string{
		"host injection":        "host=other_postgres%0Adbname%3Dother_database",
		"host whitespace":       "host=%20other_postgres%20",
		"empty host":            "host=",
		"invalid port":          "port=not_a_port",
		"oversized port":        "port=65536",
		"empty port":            "port=",
		"user injection":        "user=other_user%0Ahost%3Dother_postgres",
		"invalid database":      "dbname=other%2Fdatabase",
		"password injection":    "password=other%0Apassword",
		"contradictory aliases": "database=one_database&dbname=another_database",
	} {
		t.Run(name, func(t *testing.T) {
			workspace := t.TempDir()
			secret := filepath.Join(t.TempDir(), "database-url")
			connection := "postgres://cloud_clicker:original_password@postgres:5432/cloud_clicker?sslmode=disable&" + query
			if err := os.WriteFile(secret, []byte(connection+"\n"), 0o600); err != nil {
				t.Fatal(err)
			}
			if environment, err := databaseServiceEnvironment(secret, workspace); !errors.Is(err, ErrInvalid) || environment != nil {
				t.Fatalf("unsafe endpoint override accepted: environment=%v err=%v", environment, err)
			}
			entries, err := os.ReadDir(workspace)
			if err != nil || len(entries) != 0 {
				t.Fatalf("unsafe connection wrote credential files: count=%d err=%v", len(entries), err)
			}
		})
	}
}

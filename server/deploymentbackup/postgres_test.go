package deploymentbackup

import (
	"errors"
	"fmt"
	"net/url"
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

func TestDatabaseServiceFilesRejectMalformedQueryBeforeWriting(t *testing.T) {
	for _, query := range []string{
		"host=other%zz", "port=65%zz", "user=other%zz",
		"dbname=other%zz", "database=other%zz", "password=other%zz",
		"sslmode=verify%zz", "application_name=backup%zz",
		"host=other;port=6543", "dbname=other;database=another",
		"password=other;sslmode=disable", "%zz=value",
	} {
		t.Run(query, func(t *testing.T) {
			connection := "postgres://cloud_clicker:original_password@postgres:5432/cloud_clicker?sslmode=disable&" + query
			parsed, err := url.Parse(connection)
			if err != nil {
				t.Fatal(err)
			}
			// URL.Query (also used by the inspection driver) silently drops
			// malformed fields. Require the complete query parser's refusal,
			// not a claimed disagreement between the two lenient consumers.
			if _, err := url.ParseQuery(parsed.RawQuery); err == nil {
				t.Fatal("diagnostic query has no syntax error")
			}
			workspace := t.TempDir()
			secret := filepath.Join(t.TempDir(), "database-url")
			if err := os.WriteFile(secret, []byte(connection+"\n"), 0o600); err != nil {
				t.Fatal(err)
			}
			if environment, err := databaseServiceEnvironment(secret, workspace); !errors.Is(err, ErrInvalid) || environment != nil {
				t.Errorf("malformed query produced fallback credentials: environment=%v err=%v", environment, err)
			}
			if entries, err := os.ReadDir(workspace); err != nil || len(entries) != 0 {
				t.Errorf("malformed query wrote credential files: count=%d err=%v", len(entries), err)
			}
		})
	}
}

func TestDatabaseServiceFilesRetainEncodedQueryValues(t *testing.T) {
	for _, query := range []string{
		"password=colon%3Abackslash%5Csemicolon%3Bpercent%25",
		"password=literal%2520value",
		"application_name=backup%3Bordinary",
		"host=other_postgres&dbname=other_database&password=other%26password",
	} {
		t.Run(query, func(t *testing.T) {
			connection := "postgres://cloud_clicker:original_password@postgres:5432/cloud_clicker?sslmode=disable&" + query
			config, err := pgx.ParseConfig(connection)
			if err != nil {
				t.Fatal(err)
			}
			workspace := t.TempDir()
			secret := filepath.Join(t.TempDir(), "database-url")
			if err := os.WriteFile(secret, []byte(connection+"\n"), 0o600); err != nil {
				t.Fatal(err)
			}
			if _, err := databaseServiceEnvironment(secret, workspace); err != nil {
				t.Fatal(err)
			}
			data, err := os.ReadFile(filepath.Join(workspace, "pgpass"))
			want := strings.Join([]string{pgpassEscape(config.Host), fmt.Sprint(config.Port), pgpassEscape(config.Database), pgpassEscape(config.User), pgpassEscape(config.Password)}, ":") + "\n"
			if err != nil || string(data) != want {
				t.Fatalf("encoded query changed endpoint/password binding: err=%v", err)
			}
		})
	}
}

package deploymentbackup

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
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

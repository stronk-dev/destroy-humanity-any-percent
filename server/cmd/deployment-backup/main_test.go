package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestRecoveryIdentityRequiresDatabaseSecretPath(t *testing.T) {
	if err := runRecoveryIdentity(context.Background(), nil); err == nil {
		t.Fatal("recovery identity accepted without database URL secret")
	}
	command, class := boundedFailure("recovery-identity", errors.New("database unavailable"))
	if command != "recovery-identity" || class != "operation_failed" {
		t.Fatalf("command=%q class=%q", command, class)
	}
}

func TestBackupFailureOutputIsBounded(t *testing.T) {
	command, class := boundedFailure("recovery_code=private", errors.New("database_url=private"))
	if command != "unknown" || class != "operation_failed" {
		t.Fatalf("unbounded failure escaped: command=%q class=%q", command, class)
	}
}

func TestBackupPathsAdmitsEveryTargetEntryForValidation(t *testing.T) {
	target := t.TempDir()
	if err := os.WriteFile(filepath.Join(target, "partial.tmp"), []byte("partial"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(target, "unexpected-directory"), 0o700); err != nil {
		t.Fatal(err)
	}
	paths, err := backupPaths(target)
	if err != nil || len(paths) != 2 {
		t.Fatalf("paths=%v err=%v", paths, err)
	}
	if err := runRetention([]string{"--target", target}); err == nil {
		t.Fatal("invalid target entries were silently excluded")
	}
}

func TestCreateFlagsRequireEveryReleaseIdentityInput(t *testing.T) {
	if _, err := parseCreateFlags("create", []string{"--target=/backups"}); err == nil {
		t.Fatal("incomplete backup command accepted")
	}
	flags, err := parseCreateFlags("create", []string{
		"--target=/backups", "--database-url-file=/run/secrets/database-url",
		"--release-manifest=/release-manifest.json", "--epoch=/epoch.json",
		"--age-recipient=age1fixture", "--server-id=server", "--metrics-dir=/operations",
	})
	if err != nil || flags.target != "/backups" {
		t.Fatalf("flags=%+v err=%v", flags, err)
	}
}

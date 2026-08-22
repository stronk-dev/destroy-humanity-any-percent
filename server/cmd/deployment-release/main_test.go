package main

import (
	"context"
	"errors"
	"flag"
	"path/filepath"
	"testing"

	"cloud-clicker/server/deploymentrelease"
)

func TestReleaseFailureOutputIsBounded(t *testing.T) {
	command, class := boundedFailure("recovery_code=private", errors.New("database_url=private"))
	if command != "unknown" || class != "operation_failed" {
		t.Fatalf("unbounded failure escaped: command=%q class=%q", command, class)
	}
}

func TestRuntimeFlagsRequireEveryOperatorBoundary(t *testing.T) {
	set := newFlagSetForTest(t)
	values := addRuntimeFlags(set, true)
	args := []string{"--operator-state=" + t.TempDir(), "--operator=operator-1", "--public-origin=https://game.example",
		"--receiver-health-url=http://alertmanager:9093/-/healthy",
		"--backup-target=/backups", "--metrics-dir=/operations", "--age-recipient=age1fixture", "--age-identity-file=/run/secrets/age-identity"}
	if err := set.Parse(args); err != nil {
		t.Fatal(err)
	}
	if values.identity == "" || values.state == "" {
		t.Fatalf("runtime flags=%+v", values)
	}
	if _, err := filepath.Abs(values.state); err != nil {
		t.Fatal(err)
	}
}

func TestRollbackRequiresOffHostIdentityBeforeControllerConstruction(t *testing.T) {
	err := runRollback(context.Background(), []string{"--failed-bundle=/failed", "--previous-bundle=/previous",
		"--backup-id=20260822T180000Z-acde00000001", "--backup=/backups/20260822T180000Z-acde00000001.ccbackup",
		"--operator-state=" + t.TempDir(), "--operator=operator-1", "--public-origin=https://game.example",
		"--receiver-health-url=http://alertmanager:9093/-/healthy", "--backup-target=/backups", "--age-recipient=age1fixture"})
	if err == nil {
		t.Fatal("rollback without restore identity accepted")
	}
}

func TestInstallRequiresExactBundleAndAllOperatorBoundaries(t *testing.T) {
	if err := runInstall(context.Background(), []string{"--bundle=/candidate"}); err == nil {
		t.Fatal("install without operator boundaries accepted")
	}
	if err := runInstall(context.Background(), []string{"--operator-state=" + t.TempDir(), "--operator=operator-1"}); err == nil {
		t.Fatal("install without exact bundle accepted")
	}
	command, class := boundedFailure("install", deploymentrelease.ErrInvalid)
	if command != "install" || class != "invalid_input" {
		t.Fatalf("install failure not bounded: command=%q class=%q", command, class)
	}
}

func newFlagSetForTest(t *testing.T) *flag.FlagSet {
	t.Helper()
	return flag.NewFlagSet(t.Name(), flag.ContinueOnError)
}

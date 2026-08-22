package main

import (
	"context"
	"flag"
	"path/filepath"
	"testing"
)

func TestRuntimeFlagsRequireEveryOperatorBoundary(t *testing.T) {
	set := newFlagSetForTest(t)
	values := addRuntimeFlags(set, true)
	args := []string{"--operator-state=" + t.TempDir(), "--operator=operator-1", "--public-origin=https://game.example",
		"--receiver-health-url=http://alertmanager:9093/-/healthy",
		"--backup-target=/backups", "--age-recipient=age1fixture", "--age-identity-file=/run/secrets/age-identity"}
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

func newFlagSetForTest(t *testing.T) *flag.FlagSet {
	t.Helper()
	return flag.NewFlagSet(t.Name(), flag.ContinueOnError)
}

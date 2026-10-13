package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestActualBackupWorkerPrivateProfileIntegration(t *testing.T) {
	if os.Getenv("CLOUD_CLICKER_BACKUP_PROFILE_TEST") != "1" {
		t.Skip("requires owned private stack; use make test-deployment-backup-profile")
	}
	root := os.Getenv("CLOUD_CLICKER_OPERATIONS_ROOT")
	if !filepath.IsAbs(root) {
		t.Fatal("use the root Make selector")
	}
	project := "cloud-clicker-backup-profile-" + strings.ToLower(rand.Text())
	composeArgs := []string{"compose", "--project-name", project, "-f", filepath.Join(root, "compose.deployment-operations-test.yml"),
		"-f", filepath.Join(root, "compose.deployment-backup-operations-test.yml")}
	t.Cleanup(func() { cleanupPrivateOperationsProfile(t, composeArgs, os.Environ(), project) })
	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Minute)
	defer cancel()
	command := exec.CommandContext(ctx, "docker", append(append([]string{}, composeArgs...), "run", "--rm", "--use-aliases", "operations-test",
		"-test.run", "^TestPrivateOperationsBackupExecutionIntegration$", "-test.v", "-test.count=1")...)
	var observation bytes.Buffer
	command.Dir, command.Stdout, command.Stderr = root, io.MultiWriter(os.Stdout, &observation), os.Stderr
	if err := command.Run(); err != nil {
		t.Fatalf("actual backup/receiver journey failed: %v", err)
	}
	if !backupProfileExecuted(observation.String()) {
		t.Fatal("backup population did not explicitly pass; empty or skipped selection is not evidence")
	}
}

func backupProfileExecuted(output string) bool {
	for _, line := range strings.Split(output, "\n") {
		if strings.HasPrefix(line, "--- PASS: TestPrivateOperationsBackupExecutionIntegration (") && strings.HasSuffix(line, "s)") {
			return true
		}
	}
	return false
}

func TestBackupProfileRequiresItsExecutedPopulation(t *testing.T) {
	for _, output := range []string{"", "testing: warning: no tests to run\nPASS\n",
		"--- SKIP: TestPrivateOperationsBackupExecutionIntegration (0.00s)\nPASS\n",
		"--- PASS: TestPrivateOperationsStorageIntegration (1.00s)\nPASS\n",
		"    log: --- PASS: TestPrivateOperationsBackupExecutionIntegration (1.00s)\n"} {
		if backupProfileExecuted(output) {
			t.Fatal("missing/skipped/different population accepted")
		}
	}
	if !backupProfileExecuted("--- PASS: TestPrivateOperationsBackupExecutionIntegration (390.12s)\nPASS\n") {
		t.Fatal("actual named population refused")
	}
}

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

func TestActualGameserverOutagePrivateProfileIntegration(t *testing.T) {
	if os.Getenv("CLOUD_CLICKER_GAMESERVER_PROFILE_TEST") != "1" {
		t.Skip("requires owned private stack; use make test-deployment-gameserver-profile")
	}
	root := os.Getenv("CLOUD_CLICKER_OPERATIONS_ROOT")
	if !filepath.IsAbs(root) {
		t.Fatal("use the root Make selector")
	}
	project := "cloud-clicker-gameserver-profile-" + strings.ToLower(rand.Text())
	composeArgs := []string{"compose", "--project-name", project, "-f", filepath.Join(root, "compose.deployment-operations-test.yml"),
		"-f", filepath.Join(root, "compose.deployment-gameserver-operations-test.yml")}
	t.Cleanup(func() { cleanupPrivateOperationsProfile(t, composeArgs, os.Environ(), project) })
	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Minute)
	defer cancel()
	command := exec.CommandContext(ctx, "docker", append(append([]string{}, composeArgs...), "run", "--rm", "--use-aliases", "operations-test",
		"-test.run", "^TestPrivateOperationsGameserverIntegration$", "-test.v", "-test.count=1")...)
	var observation bytes.Buffer
	command.Dir, command.Stdout, command.Stderr = root, io.MultiWriter(os.Stdout, &observation), os.Stderr
	if err := command.Run(); err != nil {
		t.Fatalf("actual gameserver/receiver journey failed: %v", err)
	}
	if !privateProfileExecuted(observation.String(), "TestPrivateOperationsGameserverIntegration") {
		t.Fatal("gameserver population did not explicitly pass; empty/skipped output is not evidence")
	}
}

func TestGameserverProfileRequiresItsExecutedPopulation(t *testing.T) {
	const name = "TestPrivateOperationsGameserverIntegration"
	for _, output := range []string{"", "testing: warning: no tests to run\nPASS\n",
		"--- SKIP: " + name + " (0.00s)\nPASS\n",
		"--- PASS: TestPrivateOperationsIngressIntegration (1.00s)\nPASS\n",
		"    log: --- PASS: " + name + " (1.00s)\n"} {
		if privateProfileExecuted(output, name) {
			t.Fatal("missing/skipped/different population accepted")
		}
	}
	if !privateProfileExecuted("--- PASS: "+name+" (390.12s)\nPASS\n", name) {
		t.Fatal("actual named population refused")
	}
}

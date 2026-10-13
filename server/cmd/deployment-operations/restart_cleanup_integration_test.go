package main

import (
	"context"
	"crypto/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestPrivateRestartStackCancellationCleanupIntegration(t *testing.T) {
	if os.Getenv("CLOUD_CLICKER_RESTART_OBSERVER_TEST") != "1" {
		t.Skip("requires owned Docker fixture; use make test-deployment-restart-profile")
	}
	root := os.Getenv("CLOUD_CLICKER_OPERATIONS_ROOT")
	if !filepath.IsAbs(root) {
		t.Fatal("use the root Make selector")
	}
	project := "cloud-clicker-restart-cancel-" + strings.ToLower(rand.Text())
	composeArgs := []string{"compose", "--project-name", project, "-f", filepath.Join(root, "compose.deployment-operations-test.yml"),
		"-f", filepath.Join(root, "compose.deployment-restart-test.yml")}
	environment := append(os.Environ(), "CLOUD_CLICKER_RESTART_OBSERVATION_DIR="+t.TempDir())
	t.Cleanup(func() { cleanupRestartProfile(t, composeArgs, environment, project) })
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	command := exec.CommandContext(ctx, "docker", append(append([]string{}, composeArgs...), "run", "--rm", "--no-deps",
		"--entrypoint", "/bin/sh", "operations-test", "-c", "sleep 600")...)
	command.Env, command.Dir = environment, root
	command.Stdout, command.Stderr = os.Stdout, os.Stderr
	if err := command.Start(); err != nil {
		t.Fatal(err)
	}
	finished := make(chan error, 1)
	go func() { finished <- command.Wait(); close(finished) }()
	defer func() { cancel(); <-finished }()
	filter := "label=com.docker.compose.project=" + project
	var id string
	for ctx.Err() == nil {
		data, err := exec.CommandContext(ctx, "docker", "ps", "--quiet", "--no-trunc", "--filter", filter).Output()
		if err != nil {
			t.Fatal(err)
		}
		if ids := strings.Fields(string(data)); len(ids) == 1 {
			id = ids[0]
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if id == "" {
		t.Fatal("private one-off runner never became live")
	}
	cancel()
	if err := <-finished; err == nil {
		t.Fatal("canceled Compose command unexpectedly completed successfully")
	}
	check, stop := context.WithTimeout(context.Background(), time.Minute)
	defer stop()
	running, err := exec.CommandContext(check, "docker", "inspect", "--format={{.State.Running}}", id).Output()
	if err != nil || strings.TrimSpace(string(running)) != "true" {
		t.Fatal("negative did not reproduce a live runner after CLI cancellation")
	}
	cleanupRestartProfile(t, composeArgs, environment, project)
	for _, args := range [][]string{
		{"ps", "--all", "--quiet", "--filter", filter},
		{"network", "ls", "--quiet", "--filter", filter},
	} {
		data, err := exec.CommandContext(check, "docker", args...).Output()
		if err != nil || strings.TrimSpace(string(data)) != "" {
			t.Fatal("owned cancellation cleanup left a container or private network")
		}
	}
	t.Log("CLI cancellation left its real one-off runner live; cleanup removed only that project and its network")
}

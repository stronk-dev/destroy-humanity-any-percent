package main

import (
	"bufio"
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"
)

func TestActualCaddyOutagePrivateProfileIntegration(t *testing.T) {
	if os.Getenv("CLOUD_CLICKER_INGRESS_PROFILE_TEST") != "1" {
		t.Skip("requires owned private stack; use make test-deployment-ingress-profile")
	}
	root := os.Getenv("CLOUD_CLICKER_OPERATIONS_ROOT")
	if !filepath.IsAbs(root) {
		t.Fatal("use the root Make selector")
	}
	project := "cloud-clicker-ingress-profile-" + strings.ToLower(rand.Text())
	composeArgs := []string{"compose", "--project-name", project, "-f", filepath.Join(root, "compose.deployment-operations-test.yml")}
	// Cleanup expects the first -f path at its established argument coordinate.
	t.Cleanup(func() { cleanupPrivateOperationsProfile(t, composeArgs, os.Environ(), project) })
	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Minute)
	defer cancel()
	command := exec.CommandContext(ctx, "docker", append(append([]string{}, composeArgs...), "run", "--rm", "--use-aliases", "operations-test",
		"-test.run", "^TestPrivateOperationsIngressIntegration$", "-test.v", "-test.count=1")...)
	command.Dir, command.Stderr = root, os.Stderr
	stdout, err := command.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := command.Start(); err != nil {
		t.Fatal(err)
	}
	var observation strings.Builder
	finished := make(chan error, 1)
	go func() {
		reader := bufio.NewScanner(stdout)
		reader.Buffer(make([]byte, 4096), 64<<10)
		phase := 0
		var observationErr error
		for reader.Scan() {
			line := reader.Text()
			fmt.Fprintln(os.Stdout, line)
			observation.WriteString(line + "\n")
			action := ""
			if strings.HasSuffix(line, ": native_caddy_stop") {
				action = "stop"
			}
			if strings.HasSuffix(line, ": native_caddy_start") {
				action = "start"
			}
			if action == "" {
				continue
			}
			if action == "stop" && phase != 0 || action == "start" && phase != 1 {
				observationErr = errors.New("unexpected proxy fixture control order")
				cancel()
				break
			}
			ids, err := exec.CommandContext(ctx, "docker", "ps", "--all", "--quiet", "--no-trunc", "--filter", "label=com.docker.compose.project="+project,
				"--filter", "label=com.docker.compose.service=caddy").Output()
			id := strings.TrimSpace(string(ids))
			if err != nil || !regexp.MustCompile(`^[0-9a-f]{64}$`).MatchString(id) {
				observationErr = errors.New("owned proxy identity unavailable")
				cancel()
				break
			}
			labels, err := exec.CommandContext(ctx, "docker", "inspect", `--format={{index .Config.Labels "com.docker.compose.project"}} {{index .Config.Labels "com.docker.compose.service"}}`, id).Output()
			if err != nil || strings.TrimSpace(string(labels)) != project+" caddy" {
				observationErr = errors.New("refusing proxy action without exact ownership")
				cancel()
				break
			}
			args := []string{action, id}
			if action == "stop" {
				args = []string{"stop", "--timeout", "10", id}
			}
			if err := exec.CommandContext(ctx, "docker", args...).Run(); err != nil {
				observationErr = err
				cancel()
				break
			}
			running, err := exec.CommandContext(ctx, "docker", "inspect", "--format={{.State.Running}}", id).Output()
			want := "false"
			if action == "start" {
				want = "true"
			}
			if err != nil || strings.TrimSpace(string(running)) != want {
				observationErr = errors.New("actual proxy process did not reach the requested state")
				cancel()
				break
			}
			phase++
		}
		if reader.Err() != nil {
			observationErr = reader.Err()
			cancel()
		}
		waitErr := command.Wait()
		if phase != 2 && observationErr == nil {
			observationErr = errors.New("proxy fixture did not execute both physical actions")
		}
		finished <- errors.Join(observationErr, waitErr)
		close(finished)
	}()
	if err := <-finished; err != nil {
		t.Fatalf("actual proxy outage journey failed: %v", err)
	}
	if !privateProfileExecuted(observation.String(), "TestPrivateOperationsIngressIntegration") {
		t.Fatal("ingress population did not explicitly pass")
	}
}

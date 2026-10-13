package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestHostObservationActualDockerRestartIntegration(t *testing.T) {
	runActualDockerRestart(t, false)
}

func TestHostObservationActualDockerRestartProfileIntegration(t *testing.T) {
	runActualDockerRestart(t, true)
}

func runActualDockerRestart(t *testing.T, profile bool) {
	t.Helper()
	if os.Getenv("CLOUD_CLICKER_RESTART_OBSERVER_TEST") != "1" {
		t.Skip("real owned Docker container; use make test-deployment-restart-observer")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	docker := func(args ...string) []byte {
		t.Helper()
		data, err := exec.CommandContext(ctx, "docker", args...).Output()
		if err != nil {
			t.Fatalf("owned restart fixture command failed: %v", err)
		}
		return data
	}
	name := "cloud-clicker-restart-observer-" + strings.ToLower(rand.Text())
	image := "postgres:16-alpine@sha256:cf78e76683b9ca8c5733cbbdce6c9262b45b6767934dd0a95e671f9a0fc20685"
	create := func(containerName, script string) string {
		t.Helper()
		id := strings.TrimSpace(string(docker("create", "--name", containerName, "--label", "cloud-clicker.restart-observer-test="+containerName,
			"--restart=on-failure:3", "--network=none", "--read-only", "--cap-drop=ALL", "--security-opt=no-new-privileges",
			"--entrypoint=/bin/sh", image, "-c", script)))
		if !regexp.MustCompile(`^[0-9a-f]{64}$`).MatchString(id) {
			t.Fatal("Docker did not return an exact fixture container identity")
		}
		t.Cleanup(func() {
			cleanup, stop := context.WithTimeout(context.Background(), 30*time.Second)
			defer stop()
			label, err := exec.CommandContext(cleanup, "docker", "inspect", `--format={{index .Config.Labels "cloud-clicker.restart-observer-test"}}`, id).Output()
			if err != nil || strings.TrimSpace(string(label)) != containerName {
				t.Error("refusing removal without exact fixture ownership")
				return
			}
			if err := exec.CommandContext(cleanup, "docker", "rm", "--force", "--volumes", id).Run(); err != nil {
				t.Error("owned restart fixture removal failed")
			}
		})
		return id
	}
	cid := create(name, "sleep 15; exit 1")
	docker("start", cid)
	created, err := time.Parse(time.RFC3339Nano, strings.TrimSpace(string(docker("inspect", "--format={{.Created}}", cid))))
	if err != nil {
		t.Fatal("actual container birth unavailable")
	}
	base := t.TempDir()
	for _, child := range []string{"bundle", "metrics", "persistent", "backup", "journal"} {
		if err := os.Mkdir(filepath.Join(base, child), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	budget := filepath.Join(base, "budget")
	if err := os.WriteFile(budget, []byte("1000\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	original, originalJournal := commandOutput, cleanupJournalCursor
	defer func() { commandOutput = original; cleanupJournalCursor = originalJournal }()
	// Only the official-project lookup and journal source are controlled.
	// Inspection, filesystem reads and host publication remain actual code.
	commandOutput = func(program string, args ...string) ([]byte, error) {
		if program == "docker" && len(args) == 9 && args[0] == "compose" && args[5] == "ps" && args[4] == filepath.Join(base, "bundle", "compose.yml") {
			return []byte(cid + "\n"), nil
		}
		return exec.CommandContext(ctx, program, args...).Output()
	}
	cleanupJournalCursor = func(string, time.Time) (string, error) { return "", nil }
	observe := func(want uint64) {
		t.Helper()
		if err := runHostObserve([]string{"--bundle=" + filepath.Join(base, "bundle"), "--metrics-dir=" + filepath.Join(base, "metrics"),
			"--persistent-path=" + filepath.Join(base, "persistent"), "--backup-path=" + filepath.Join(base, "backup"),
			"--journal-path=" + filepath.Join(base, "journal"), "--journal-budget-bytes-file=" + budget}); err != nil {
			t.Fatal(err)
		}
		data, err := os.ReadFile(filepath.Join(base, "metrics", "host.prom"))
		if err != nil || !strings.Contains(string(data), fmt.Sprintf("cloud_clicker_gameserver_restart_count %d\n", want)) ||
			!strings.Contains(string(data), fmt.Sprintf("cloud_clicker_gameserver_container_created_timestamp_seconds %d\n", created.Unix())) {
			t.Fatal("actual inspected count/birth did not reach host publication")
		}
		if strings.Contains(string(data), cid) || strings.Contains(string(data), name) {
			t.Fatal("fixture/container identity escaped as a metric")
		}
	}
	observe(0)
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		// Independent single-field inspection establishes the actual engine
		// population, not the production parser or a typed counter literal.
		count, err := strconv.ParseUint(strings.TrimSpace(string(docker("inspect", "--format={{.RestartCount}}", cid))), 10, 64)
		if err != nil || count > 3 {
			t.Fatal("restart-capped fixture produced an invalid population")
		}
		if count == 3 {
			observe(3)
			if elapsed := time.Since(created); elapsed < 0 || elapsed > 10*time.Minute {
				t.Fatal("actual restarts do not satisfy the recent-container population")
			}
			t.Logf("actual Docker automatic restarts3 observed and published with their bound birth, age=%s", time.Since(created))
			break
		}
		select {
		case <-ctx.Done():
			t.Fatal("real restart population did not reach its objective")
		case <-ticker.C:
		}
	}
	if profile {
		root := os.Getenv("CLOUD_CLICKER_OPERATIONS_ROOT")
		if !filepath.IsAbs(root) {
			t.Fatal("use make test-deployment-restart-profile for the owned private stack")
		}
		project := "cloud-clicker-restart-profile-" + strings.ToLower(rand.Text())
		composeArgs := []string{"compose", "--project-name", project, "-f", filepath.Join(root, "compose.deployment-operations-test.yml"),
			"-f", filepath.Join(root, "compose.deployment-restart-test.yml")}
		metrics := filepath.Join(base, "metrics")
		environment := append(os.Environ(), "CLOUD_CLICKER_RESTART_OBSERVATION_DIR="+metrics)
		t.Cleanup(func() { cleanupPrivateOperationsProfile(t, composeArgs, environment, project) })
		command := exec.CommandContext(ctx, "docker", append(append([]string{}, composeArgs...), "run", "--rm", "--use-aliases", "operations-test",
			"-test.run", "^TestPrivateOperationsRestartIntegration$", "-test.v", "-test.count=1")...)
		command.Env = environment
		command.Dir = root
		command.Stdout, command.Stderr = os.Stdout, os.Stderr
		if err := command.Start(); err != nil {
			t.Fatal(err)
		}
		finished := make(chan error, 1)
		go func() { finished <- command.Wait(); close(finished) }()
		defer func() { cancel(); <-finished }()
		// Monitoring starts only after the real count reaches three. The child
		// requests recovery only after proving first-scrape firing and delivery.
		request := filepath.Join(metrics, "restart-recovery-request")
		for {
			data, err := os.ReadFile(request)
			if err == nil {
				if string(data) != "recover\n" {
					t.Fatal("invalid private restart recovery request")
				}
				break
			}
			if !os.IsNotExist(err) {
				t.Fatal(err)
			}
			select {
			case err := <-finished:
				t.Fatalf("private restart profile ended before requesting actual recovery: %v", err)
			case <-ctx.Done():
				t.Fatal("private restart profile did not reach firing/delivery")
			case <-ticker.C:
			}
		}
		previousBirth := created
		cid = create(name+"-recovered", "sleep 600")
		docker("start", cid)
		created, err = time.Parse(time.RFC3339Nano, strings.TrimSpace(string(docker("inspect", "--format={{.Created}}", cid))))
		if err != nil || !created.After(previousBirth) {
			t.Fatal("replacement container does not have an actual later birth")
		}
		observe(0)
		if err := <-finished; err != nil {
			t.Fatalf("private restart receiver/resolution proof failed: %v", err)
		}
		t.Log("actual replacement count0 reached the same private receiver as matching resolution")
	}
}

func cleanupPrivateOperationsProfile(t *testing.T, composeArgs, environment []string, project string) {
	t.Helper()
	cleanup, stop := context.WithTimeout(context.Background(), time.Minute)
	defer stop()
	// Killing the Compose CLI does not kill its one-off container. Down
	// also omits that runner, leaving the private network occupied.
	ids, err := exec.CommandContext(cleanup, "docker", "ps", "--all", "--quiet", "--no-trunc",
		"--filter", "label=com.docker.compose.project="+project,
		"--filter", "label=com.docker.compose.service=operations-test").Output()
	if err != nil {
		t.Errorf("cannot enumerate owned private runner: %v", err)
		return
	}
	for _, id := range strings.Fields(string(ids)) {
		if !regexp.MustCompile(`^[0-9a-f]{64}$`).MatchString(id) {
			t.Error("invalid owned private runner identity")
			return
		}
		labels, err := exec.CommandContext(cleanup, "docker", "inspect",
			`--format={{index .Config.Labels "com.docker.compose.project"}} {{index .Config.Labels "com.docker.compose.service"}} {{index .Config.Labels "com.docker.compose.oneoff"}}`, id).Output()
		if err != nil || strings.TrimSpace(string(labels)) != project+" operations-test True" {
			t.Error("refusing removal without exact private runner ownership")
			return
		}
		if err := exec.CommandContext(cleanup, "docker", "rm", "--force", "--volumes", id).Run(); err != nil {
			t.Errorf("owned private runner removal: %v", err)
			return
		}
	}
	command := exec.CommandContext(cleanup, "docker", append(append([]string{}, composeArgs...), "down", "--volumes")...)
	command.Env = environment
	command.Dir = filepath.Dir(composeArgs[4])
	command.Stdout, command.Stderr = os.Stdout, os.Stderr
	if err := command.Run(); err != nil {
		t.Errorf("owned private operations stack cleanup: %v", err)
	}
}

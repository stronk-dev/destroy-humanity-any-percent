package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"cloud-clicker/server/operations"
)

// This is journal-file compatibility evidence, not a host logging-driver,
// retention, capacity-planning or clean-host release witness.
func TestActualJournalCollectionIntegration(t *testing.T) {
	if os.Getenv("CLOUD_CLICKER_JOURNAL_TEST") != "1" {
		t.Skip("requires owned Linux journald fixture; use make test-deployment-journal")
	}
	root := os.Getenv("CLOUD_CLICKER_OPERATIONS_ROOT")
	if !filepath.IsAbs(root) {
		t.Fatal("use the root Make selector")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	name := "cloud-clicker-journal-" + strings.ToLower(rand.Text())
	command := exec.CommandContext(ctx, "docker", "create", "--name", name,
		"--label", "cloud-clicker.journal-test="+name, "--network=none", "--read-only",
		"--mount", "type=volume,destination=/etc", "--tmpfs", "/run",
		"--tmpfs", "/var/log/journal:size=64g", "--tmpfs", "/tmp",
		"--mount", "type=bind,source="+filepath.Join(root, ".cache/bin/journal-operations.test")+",destination=/journal.test,readonly",
		"--env", "CLOUD_CLICKER_JOURNAL_CHILD=1", "--entrypoint=/journal.test",
		"mcr.microsoft.com/playwright:v1.62.0-noble", "-test.run=^TestNativeJournalCollectionIntegration$", "-test.v", "-test.count=1", "-test.timeout=90s")
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("create owned journal fixture: %v: %s", err, output)
	}
	t.Cleanup(func() {
		cleanup, stop := context.WithTimeout(context.Background(), 30*time.Second)
		defer stop()
		label, err := exec.CommandContext(cleanup, "docker", "inspect", `--format={{index .Config.Labels "cloud-clicker.journal-test"}}`, name).Output()
		if err != nil || strings.TrimSpace(string(label)) != name {
			t.Error("refusing removal without exact journal-fixture ownership")
			return
		}
		if err := exec.CommandContext(cleanup, "docker", "rm", "--force", "--volumes", name).Run(); err != nil {
			t.Error("owned journal-fixture removal failed")
		}
	})
	var observation bytes.Buffer
	command = exec.CommandContext(ctx, "docker", "start", "--attach", name)
	command.Stdout, command.Stderr = io.MultiWriter(os.Stdout, &observation), os.Stderr
	if err := command.Run(); err != nil {
		t.Fatalf("actual journal collection failed: %v", err)
	}
	if !privateProfileExecuted(observation.String(), "TestNativeJournalCollectionIntegration") {
		t.Fatal("journal population did not explicitly pass; skipped or absent tests are not evidence")
	}
}

func TestNativeJournalCollectionIntegration(t *testing.T) {
	if os.Getenv("CLOUD_CLICKER_JOURNAL_CHILD") != "1" {
		t.Skip("runs only in the disposable journal fixture")
	}
	if runtime.GOOS != "linux" {
		t.Fatal("native journal fixture requires Linux")
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	if err := os.WriteFile("/etc/machine-id", []byte("42e6c46f32034c51a0e449c52b7b62f4\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var diagnostics bytes.Buffer
	daemon := exec.CommandContext(ctx, "/usr/lib/systemd/systemd-journald")
	daemon.Stdout, daemon.Stderr = &diagnostics, &diagnostics
	if err := daemon.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = daemon.Process.Signal(os.Interrupt)
		if err := daemon.Wait(); err != nil {
			t.Errorf("journald did not shut down cleanly: %v: %s", err, diagnostics.String())
		}
	}()
	for {
		if info, err := os.Stat("/run/systemd/journal/socket"); err == nil && info.Mode()&os.ModeSocket != 0 {
			break
		}
		select {
		case <-ctx.Done():
			t.Fatal("journald socket did not become ready")
		case <-time.After(20 * time.Millisecond):
		}
	}
	journal := "/var/log/journal"
	run := func(input string, program string, args ...string) []byte {
		t.Helper()
		command := exec.CommandContext(ctx, program, args...)
		command.Stdin = strings.NewReader(input)
		var diagnostic bytes.Buffer
		command.Stderr = &diagnostic
		data, err := command.Output()
		if err != nil || diagnostic.Len() != 0 {
			t.Fatalf("native journal command %s failed: %v: %s", program, err, diagnostic.String())
		}
		return data
	}
	flush := func() {
		t.Helper()
		run("", "journalctl", "--sync")
		run("", "journalctl", "--flush")
	}
	flush()
	if cursor, err := readCleanupJournal(journal, time.Now().UTC()); err != nil || cursor != "" {
		t.Fatalf("healthy empty native journal refused: cursor=%q error=%v", cursor, err)
	}
	started := time.Now().UTC().Truncate(time.Second)
	emit := func(service, tag, message string) {
		t.Helper()
		run("SYSLOG_IDENTIFIER=cloud-clicker-"+service+"\nCONTAINER_TAG="+tag+"\nMESSAGE="+message+"\n", "logger", "--journald")
	}
	for _, service := range []string{"caddy", "gameserver", "postgres", "backup", "prometheus", "alertmanager", "node-exporter"} {
		emit(service, "cloud-clicker-"+service, `{"event_class":"fixture"}`)
	}
	emit("gameserver", "cloud-clicker-gameserver", `{"event_class":"composed_job_failure","job_name":"credential_cleanup"}`)
	// Neither collector may import an unrelated service's invalid message.
	emit("unrelated", "cloud-clicker-unrelated", "not a cleanup JSON message")
	flush()
	// Both collectors use a whole-second upper query bound. Cross that actual
	// boundary so newly persisted microsecond timestamps belong to the window.
	select {
	case <-ctx.Done():
		t.Fatal("journal collection window did not complete")
	case <-time.After(time.Until(time.Unix(time.Now().Unix()+1, 0))):
	}
	data := run("", "journalctl", "--directory="+journal, "--output=json", "--all", "--no-pager", "CONTAINER_TAG=cloud-clicker-gameserver")
	var rows []struct {
		Cursor  string `json:"__CURSOR"`
		Message string `json:"MESSAGE"`
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	for {
		var row struct {
			Cursor  string `json:"__CURSOR"`
			Message string `json:"MESSAGE"`
		}
		if err := decoder.Decode(&row); err == io.EOF {
			break
		} else if err != nil {
			t.Fatal(err)
		}
		rows = append(rows, row)
	}
	if len(rows) != 2 || rows[1].Cursor == "" || !strings.Contains(rows[1].Message, `"credential_cleanup"`) {
		t.Fatal("native journal did not retain the exact two gameserver events in order")
	}
	wantCursor := fmt.Sprintf("%x", sha256.Sum256([]byte(rows[1].Cursor)))
	if cursor, err := readCleanupJournal(journal, time.Now().UTC()); err != nil || cursor != wantCursor {
		t.Fatalf("native cleanup failure cursor lost: got=%q want=%q error=%v", cursor, wantCursor, err)
	}
	// Use the actual short population's JSON byte rate to supply the fixture's
	// reserve; this is not a release-workload capacity observation. The private
	// tmpfs limit allocates no 64GiB backing memory and is removed with its owner.
	all := run("", "journalctl", "--directory="+journal, "--since=@"+fmt.Sprint(started.Unix()), "--output=json", "--all", "--no-pager")
	budget := uint64(math.Ceil(float64(len(all))*float64(24*time.Hour)/float64(time.Since(started)))) * 14
	observe := func(directory string) ([]byte, error) {
		t.Helper()
		output, err := os.CreateTemp(t.TempDir(), "observation-")
		if err != nil {
			t.Fatal(err)
		}
		defer output.Close()
		previous := os.Stdout
		os.Stdout = output
		defer func() { os.Stdout = previous }()
		err = runJournalObserve([]string{"--population=native_journal_compatibility", "--started-at=" + started.Format(time.RFC3339),
			"--journal-path=" + directory, "--journal-max-use-bytes=" + fmt.Sprint(budget)})
		data, readErr := os.ReadFile(output.Name())
		if readErr != nil {
			t.Fatal(readErr)
		}
		return data, err
	}
	output, err := observe(journal)
	observation, decodeErr := operations.DecodeJournalObservation(output)
	if err != nil || decodeErr != nil || observation.Samples != 8 {
		t.Fatalf("native workload population lost: samples=%d error=%v decode=%v", observation.Samples, err, decodeErr)
	}
	empty := t.TempDir()
	if cursor, err := readCleanupJournal(empty, time.Now().UTC()); err != nil || cursor != "" {
		t.Fatalf("collector read another directory: cursor=%q error=%v", cursor, err)
	}
	if output, err := observe(empty); err == nil || len(output) != 0 {
		t.Fatal("empty requested directory produced workload evidence")
	}
	if cursor, err := readCleanupJournal(filepath.Join(empty, "absent"), time.Now().UTC()); err == nil || cursor != "" {
		t.Fatal("missing native journal directory was silently accepted")
	}
	if output, err := observe(filepath.Join(empty, "absent")); err == nil || len(output) != 0 {
		t.Fatal("missing native journal directory produced workload evidence")
	}
	emit("gameserver", "cloud-clicker-gameserver", `{"event_class":"composed_job_failure","job_name":"unknown_fixture_job"}`)
	flush()
	if _, err := readCleanupJournal(journal, time.Now().UTC().Add(time.Second)); err == nil {
		t.Fatal("native unknown cleanup job was accepted")
	}
	t.Log("actual journald files: empty collection, eight rows from seven selected services, failure cursor, directory isolation and unknown-job refusal passed; no host driver/retention credit")
}

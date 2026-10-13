package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestOperationsFailureOutputIsBounded(t *testing.T) {
	command, class := boundedFailure("recovery_code=private", errors.New("database_url=private"))
	if command != "unknown" || class != "operation_failed" {
		t.Fatalf("unbounded failure escaped: command=%q class=%q", command, class)
	}
}

func TestAlertObserveRequiresExactBundleAndOutput(t *testing.T) {
	if err := runAlertObserve(nil); err == nil {
		t.Fatal("empty alert observation inputs accepted")
	}
}

func TestRecordAndJournalRenderCommands(t *testing.T) {
	directory := t.TempDir()
	if err := runRecord([]string{"--metrics-dir=" + directory, "--operation=restore", "--result=success", "--at=2026-08-22T12:00:00Z"}); err != nil {
		t.Fatal(err)
	}
	metric, _ := os.ReadFile(filepath.Join(directory, "restore-success.prom"))
	if !strings.Contains(string(metric), "cloud_clicker_restore_last_success_timestamp_seconds") {
		t.Fatalf("missing restore metric: %q", metric)
	}
	observation := `{"schema_version":1,"population":"r006_release_workload","started_at":"2026-08-22T10:00:00Z","completed_at":"2026-08-22T11:00:00Z","objective_completed":true,"guard_exhausted":false,"samples":2,"observed_bytes":1000,"peak_bytes_per_day":24000,"filesystem_bytes":1000000,"journal_max_use_bytes":336000,"journal_retention_seconds":1209600,"storage_alert_fraction":0.8,"raw_ip_enabled":false,"security_sink_separate":false,"security_retention_seconds":0}`
	observationPath := filepath.Join(directory, "observation.json")
	templatePath := filepath.Join(directory, "journal.template")
	output := filepath.Join(directory, "rendered", "cloud-clicker.conf")
	budgetOutput := filepath.Join(directory, "rendered", "journal-budget")
	if err := os.WriteFile(observationPath, []byte(observation), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(templatePath, []byte("[Journal]\nMaxRetentionSec=@@MAX_RETENTION_SEC@@\nSystemMaxUse=@@SYSTEM_MAX_USE@@\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := runJournalRender([]string{"--observation=" + observationPath, "--template=" + templatePath, "--output=" + output, "--budget-output=" + budgetOutput}); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(output)
	if string(data) != "[Journal]\nMaxRetentionSec=1209600\nSystemMaxUse=336000\n" {
		t.Fatalf("wrong rendered policy: %q", data)
	}
	budget, _ := os.ReadFile(budgetOutput)
	if string(budget) != "336000\n" {
		t.Fatalf("wrong measured budget: %q", budget)
	}
	if err := runJournalRender([]string{"--observation=" + observationPath, "--template=" + templatePath, "--output=" + output, "--budget-output=" + budgetOutput}); err == nil {
		t.Fatal("existing policy overwritten")
	}
}

func TestAlertDeliveryRequiresNotificationCounterIncrease(t *testing.T) {
	requests := 0
	alertmanager := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/api/v2/alerts":
			requests++
			response.WriteHeader(http.StatusOK)
		case "/metrics":
			fmt.Fprintf(response, "alertmanager_notification_requests_total{integration=\"webhook\"} %d\nalertmanager_notification_requests_failed_total{integration=\"webhook\",reason=\"other\"} 0\n", 4+requests)
		default:
			response.WriteHeader(http.StatusNotFound)
		}
	}))
	defer alertmanager.Close()
	receiver := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) { response.WriteHeader(http.StatusNoContent) }))
	defer receiver.Close()
	if err := runAlertTest([]string{"--alertmanager-url=" + alertmanager.URL, "--receiver-health-url=" + receiver.URL}); err != nil {
		t.Fatal(err)
	}
	requests = 0
	stalled := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.URL.Path == "/metrics" {
			fmt.Fprintln(response, "alertmanager_notification_requests_total{integration=\"webhook\"} 4\nalertmanager_notification_requests_failed_total{integration=\"webhook\",reason=\"other\"} 0")
			return
		}
		response.WriteHeader(http.StatusOK)
	}))
	defer stalled.Close()
	started := time.Now()
	if err := runAlertTest([]string{"--alertmanager-url=" + stalled.URL, "--receiver-health-url=" + receiver.URL}); err == nil || time.Since(started) < 5*time.Second {
		t.Fatalf("health-only receiver path accepted too early: err=%v elapsed=%v", err, time.Since(started))
	}
}

func TestHostObservationWritesExactRestartAndStorageSignals(t *testing.T) {
	base := t.TempDir()
	for _, name := range []string{"bundle", "metrics", "persistent", "backup", "journal"} {
		if err := os.Mkdir(filepath.Join(base, name), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(base, "bundle", "compose.yml"), []byte("fixture"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(base, "journal", "entry"), []byte("journal"), 0o600); err != nil {
		t.Fatal(err)
	}
	budget := filepath.Join(base, "budget")
	if err := os.WriteFile(budget, []byte("1000\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	original := commandOutput
	originalJournal := cleanupJournalCursor
	cleanupJournalCursor = func(path string, _ time.Time) (string, error) {
		if path != filepath.Join(base, "journal") {
			t.Fatal("collector did not read the configured journal subject")
		}
		return strings.Repeat("a", 64), nil
	}
	commandOutput = func(_ string, args ...string) ([]byte, error) {
		for _, value := range args {
			if value == "inspect" {
				return []byte("3 2026-08-22T12:00:00Z\n"), nil
			}
		}
		return []byte("container-id\n"), nil
	}
	defer func() { commandOutput = original; cleanupJournalCursor = originalJournal }()
	if err := runHostObserve([]string{"--bundle=" + filepath.Join(base, "bundle"), "--metrics-dir=" + filepath.Join(base, "metrics"),
		"--persistent-path=" + filepath.Join(base, "persistent"), "--backup-path=" + filepath.Join(base, "backup"),
		"--journal-path=" + filepath.Join(base, "journal"), "--journal-budget-bytes-file=" + budget}); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(filepath.Join(base, "metrics", "host.prom"))
	wantCreated := fmt.Sprintf("cloud_clicker_gameserver_container_created_timestamp_seconds %d", time.Date(2026, 8, 22, 12, 0, 0, 0, time.UTC).Unix())
	if !strings.Contains(string(data), "cloud_clicker_gameserver_restart_count 3") || !strings.Contains(string(data), wantCreated) || !strings.Contains(string(data), "cloud_clicker_journal_usage_bytes 7") {
		t.Fatalf("host observation missing: %s", data)
	}
	data, err := os.ReadFile(filepath.Join(base, "metrics", "cleanup.prom"))
	if err != nil || !strings.Contains(string(data), "cloud_clicker_cleanup_failure_detected_timestamp_seconds ") {
		t.Fatal("host observer did not publish the durable cleanup failure")
	}
	cleanupJournalCursor = func(string, time.Time) (string, error) { return "", errors.New("journal unavailable") }
	before := string(data)
	if err := runHostObserve([]string{"--bundle=" + filepath.Join(base, "bundle"), "--metrics-dir=" + filepath.Join(base, "metrics"),
		"--persistent-path=" + filepath.Join(base, "persistent"), "--backup-path=" + filepath.Join(base, "backup"),
		"--journal-path=" + filepath.Join(base, "journal"), "--journal-budget-bytes-file=" + budget}); err == nil {
		t.Fatal("failed journal command reported successful host observation")
	}
	data, _ = os.ReadFile(filepath.Join(base, "metrics", "cleanup.prom"))
	if string(data) != before {
		t.Fatal("failed journal command refreshed its heartbeat")
	}
}

func TestRestartObservationRequiresExactNonfutureBirthAndCount(t *testing.T) {
	at := time.Date(2026, 10, 13, 0, 0, 0, 0, time.UTC)
	for _, input := range []string{
		"3", "3 unknown", "-1 2026-10-12T23:59:00Z", "3.0 2026-10-12T23:59:00Z",
		"18446744073709551616 2026-10-12T23:59:00Z", "3 0001-01-01T00:00:00Z",
		"3 1970-01-01T00:00:00Z", "3 2026-10-13T00:00:01Z", "3 2026-10-12T23:59:00Z extra",
	} {
		if _, _, err := parseRestartObservation([]byte(input), at); err == nil {
			t.Fatalf("invalid restart observation accepted: %q", input)
		}
	}
	created := at.Add(-time.Minute)
	count, birth, err := parseRestartObservation([]byte("3 "+created.Format(time.RFC3339Nano)+"\n"), at)
	if err != nil || count != 3 || !birth.Equal(created) {
		t.Fatal("valid restart count lost its actual container birth")
	}
	if _, _, err := parseRestartObservation([]byte("3 "+created.Format(time.RFC3339Nano)), time.Time{}); err == nil {
		t.Fatal("missing collection time accepted")
	}
}

func TestCleanupJournalCommandConsumesCompleteOutputAndExitStatus(t *testing.T) {
	before := journalCommandContext
	defer func() { journalCommandContext = before }()
	at := time.Date(2026, 10, 13, 0, 0, 0, 0, time.UTC)
	for _, mode := range []string{"valid", "empty", "truncated", "failed", "warning"} {
		t.Run(mode, func(t *testing.T) {
			journalCommandContext = func(ctx context.Context, name string, args ...string) *exec.Cmd {
				joined := strings.Join(args, " ")
				if name != "journalctl" || !strings.Contains(joined, "--directory=/fixture/journal") ||
					!strings.Contains(joined, fmt.Sprintf("--since=@%d", at.Add(-14*24*time.Hour).Unix())) ||
					!strings.Contains(joined, "CONTAINER_TAG=cloud-clicker-gameserver") ||
					!strings.Contains(joined, "--quiet") || strings.Contains(joined, "--grep") || strings.Contains(joined, "--lines") {
					t.Fatal("journal command omitted its subject, full retained window or JSON/no-entry contract")
				}
				return exec.CommandContext(ctx, os.Args[0], "-test.run=^TestCleanupJournalCommandProcess$", "--", mode)
			}
			cursor, err := readCleanupJournal("/fixture/journal", at)
			if mode == "valid" && (err != nil || len(cursor) != 64) ||
				mode == "empty" && (err != nil || cursor != "") ||
				(mode == "truncated" || mode == "failed" || mode == "warning") && err == nil {
				t.Fatalf("command output/exit did not discriminate mode=%s cursor=%q error=%v", mode, cursor, err)
			}
		})
	}
}

// A controlled subprocess, not a systemd or Docker-driver claim.
func TestCleanupJournalCommandProcess(t *testing.T) {
	if len(os.Args) < 2 || os.Args[len(os.Args)-2] != "--" {
		return
	}
	mode := os.Args[len(os.Args)-1]
	if mode != "empty" {
		fmt.Println(`{"CONTAINER_TAG":"cloud-clicker-gameserver","__CURSOR":"fixture","__REALTIME_TIMESTAMP":"1","MESSAGE":"{\"event_class\":\"composed_job_failure\",\"job_name\":\"credential_cleanup\"}"}`)
	}
	if mode == "truncated" {
		fmt.Print(`{"MESSAGE":`)
	}
	if mode == "failed" {
		os.Exit(1)
	}
	if mode == "warning" {
		fmt.Fprintln(os.Stderr, "fixture exclusion")
	}
	os.Exit(0)
}

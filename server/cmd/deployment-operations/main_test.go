package main

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
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
			fmt.Fprintf(response, "alertmanager_notifications_total{integration=\"webhook\"} %d\n", 4+requests)
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
			fmt.Fprintln(response, `alertmanager_notifications_total{integration="webhook"} 4`)
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
	commandOutput = func(_ string, args ...string) ([]byte, error) {
		for _, value := range args {
			if value == "inspect" {
				return []byte("3\n"), nil
			}
		}
		return []byte("container-id\n"), nil
	}
	defer func() { commandOutput = original }()
	if err := runHostObserve([]string{"--bundle=" + filepath.Join(base, "bundle"), "--metrics-dir=" + filepath.Join(base, "metrics"),
		"--persistent-path=" + filepath.Join(base, "persistent"), "--backup-path=" + filepath.Join(base, "backup"),
		"--journal-path=" + filepath.Join(base, "journal"), "--journal-budget-bytes-file=" + budget}); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(filepath.Join(base, "metrics", "host.prom"))
	if !strings.Contains(string(data), "cloud_clicker_gameserver_restart_count 3") || !strings.Contains(string(data), "cloud_clicker_journal_usage_bytes 7") {
		t.Fatalf("host observation missing: %s", data)
	}
}

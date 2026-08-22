package operations

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRecordOperationPreservesSuccessAcrossLaterFailure(t *testing.T) {
	directory := t.TempDir()
	succeeded := time.Unix(1_700_000_000, 0)
	failed := succeeded.Add(time.Hour)
	if err := RecordOperation(directory, "backup", "success", succeeded); err != nil {
		t.Fatal(err)
	}
	if err := RecordOperation(directory, "backup", "failure", failed); err != nil {
		t.Fatal(err)
	}
	success, err := os.ReadFile(filepath.Join(directory, "backup-success.prom"))
	if err != nil || !strings.Contains(string(success), "cloud_clicker_backup_last_success_timestamp_seconds 1700000000") {
		t.Fatalf("success observation lost: %q err=%v", success, err)
	}
	failure, err := os.ReadFile(filepath.Join(directory, "backup-failure.prom"))
	if err != nil || !strings.Contains(string(failure), "cloud_clicker_backup_last_failure_timestamp_seconds 1700003600") {
		t.Fatalf("failure observation missing: %q err=%v", failure, err)
	}
	entries, _ := os.ReadDir(directory)
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".tmp") {
			t.Fatalf("partial metric survived: %s", entry.Name())
		}
	}
}

func TestRecordOperationRejectsUnboundedNamesAndSymlinkTarget(t *testing.T) {
	directory := t.TempDir()
	if err := RecordOperation(directory, "player-controlled", "success", time.Now()); !errors.Is(err, ErrInvalid) {
		t.Fatalf("unbounded operation accepted: %v", err)
	}
	target := filepath.Join(directory, "outside")
	if err := os.WriteFile(target, []byte("protected"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, filepath.Join(directory, "release-success.prom")); err != nil {
		t.Fatal(err)
	}
	if err := RecordOperation(directory, "release", "success", time.Now()); !errors.Is(err, ErrInvalid) {
		t.Fatalf("symlink metric target accepted: %v", err)
	}
	data, _ := os.ReadFile(target)
	if string(data) != "protected" {
		t.Fatalf("symlink target changed: %q", data)
	}
}

func TestRecordHostUsesOnlyBoundedStorageLabels(t *testing.T) {
	directory := t.TempDir()
	if err := RecordHost(directory, 0.75, 0.5, 3, 80, 100); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(directory, "host.prom"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`storage="backup"} 0.5`, `storage="persistent"} 0.75`, "cloud_clicker_gameserver_restart_count 3", "cloud_clicker_journal_usage_bytes 80"} {
		if !strings.Contains(string(data), want) {
			t.Fatalf("missing host metric %q in %s", want, data)
		}
	}
	if err := RecordHost(directory, 0, 0, 0, 101, 100); err != nil {
		t.Fatalf("over-budget journal pressure was hidden: %v", err)
	}
	data, _ = os.ReadFile(filepath.Join(directory, "host.prom"))
	if !strings.Contains(string(data), "cloud_clicker_journal_usage_bytes 101") {
		t.Fatalf("over-budget journal pressure absent: %s", data)
	}
}

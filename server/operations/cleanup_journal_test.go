package operations

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCleanupJournalConsumesActualBoundedJobFailure(t *testing.T) {
	var logs bytes.Buffer
	registry, err := NewRegistryWithLogger(fakeDatabase{ping: errors.New("unused")}, slog.New(slog.NewJSONHandler(&logs, nil)))
	if err != nil {
		t.Fatal(err)
	}
	private := errors.New("recovery_code=private database_url=private account_id=private")
	if err := registry.ObserveJob("credential_cleanup", time.Now(), func() error { return private }); !errors.Is(err, private) {
		t.Fatal("logging changed the job error")
	}
	message := strings.TrimSpace(logs.String())
	if strings.Contains(message, "private") || !strings.Contains(message, `"event_class":"composed_job_failure"`) {
		t.Fatal("failure event absent or includes underlying private error")
	}
	row := map[string]string{"CONTAINER_TAG": "cloud-clicker-gameserver", "__CURSOR": "fixture-cursor", "__REALTIME_TIMESTAMP": "1", "MESSAGE": message}
	data, _ := json.Marshal(row)
	cursor, err := CleanupFailureFromJournal(bytes.NewReader(data))
	want := fmt.Sprintf("%x", sha256.Sum256([]byte("fixture-cursor")))
	if err != nil || cursor != want {
		t.Fatalf("actual logger event did not reach journal consumer: cursor=%q error=%v", cursor, err)
	}
	if err := registry.ObserveJob("credential_cleanup", time.Now(), func() error { return nil }); err != nil || strings.TrimSpace(logs.String()) != message {
		t.Fatal("successful cleanup emitted a failure")
	}
	if err := registry.ObserveJob("unknown", time.Now(), func() error { return private }); err == nil || strings.TrimSpace(logs.String()) != message {
		t.Fatal("unknown job escaped closed-set logging")
	}
}

func TestCleanupJournalRejectsIncompleteAndForeignPopulations(t *testing.T) {
	valid := `{"CONTAINER_TAG":"cloud-clicker-gameserver","__CURSOR":"cursor","__REALTIME_TIMESTAMP":"1","MESSAGE":"{\"event_class\":\"composed_job_failure\",\"job_name\":\"credential_cleanup\"}"}`
	for name, data := range map[string]string{
		"truncated":      valid + `{"MESSAGE":`,
		"foreign_tag":    strings.Replace(valid, "cloud-clicker-gameserver", "other", 1),
		"missing_cursor": strings.Replace(valid, `"__CURSOR":"cursor",`, "", 1),
		"bad_time":       strings.Replace(valid, `"__REALTIME_TIMESTAMP":"1"`, `"__REALTIME_TIMESTAMP":"bad"`, 1),
		"partial":        strings.Replace(valid, `"MESSAGE":`, `"CONTAINER_PARTIAL_MESSAGE":"true","MESSAGE":`, 1),
		"unknown_job":    strings.Replace(valid, "credential_cleanup", "private_name", 1),
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := CleanupFailureFromJournal(strings.NewReader(data)); err == nil {
				t.Fatal("invalid journal population reported complete")
			}
		})
	}
	for _, data := range []string{"", strings.Replace(valid, "credential_cleanup", "guild_sweep", 1), strings.Replace(valid, "composed_job_failure", "runtime_failure", 1)} {
		if cursor, err := CleanupFailureFromJournal(strings.NewReader(data)); cursor != "" || err != nil {
			t.Fatalf("healthy/non-cleanup population rejected: %q %v", cursor, err)
		}
	}
}

func TestCleanupJournalSurvivesRestartAndDoesNotRefreshSeenFailure(t *testing.T) {
	directory := t.TempDir()
	at := time.Date(2026, 10, 13, 12, 0, 0, 0, time.UTC)
	first, second := strings.Repeat("a", 64), strings.Repeat("b", 64)
	for _, subject := range []struct {
		cursor        string
		now, detected time.Time
	}{
		{"", at, time.Unix(0, 0)},
		{first, at.Add(time.Minute), at.Add(time.Minute)},
		{first, at.Add(7 * time.Minute), at.Add(time.Minute)}, // reread after observer restart
		{"", at.Add(8 * time.Minute), at.Add(time.Minute)},    // aged-out journal entry
		{second, at.Add(9 * time.Minute), at.Add(9 * time.Minute)},
	} {
		if err := RecordCleanupJournal(directory, subject.cursor, subject.now); err != nil {
			t.Fatal(err)
		}
		data, err := os.ReadFile(filepath.Join(directory, "cleanup.prom"))
		if err != nil || !strings.Contains(string(data), fmt.Sprintf("cloud_clicker_cleanup_failure_detected_timestamp_seconds %d\n", subject.detected.Unix())) ||
			!strings.Contains(string(data), fmt.Sprintf("cloud_clicker_cleanup_journal_observed_timestamp_seconds %d\n", subject.now.Unix())) || strings.Contains(string(data), first) {
			t.Fatalf("wrong durable cleanup observation: %s error=%v", data, err)
		}
	}
	for _, invalid := range []string{`{`, `{}`, fmt.Sprintf(`{"cursor_hash":%q,"detected_at":123}`, first)} {
		if err := os.WriteFile(filepath.Join(directory, "cleanup-state.json"), []byte(invalid), 0o600); err != nil {
			t.Fatal(err)
		}
		before, _ := os.ReadFile(filepath.Join(directory, "cleanup.prom"))
		if err := RecordCleanupJournal(directory, second, at.Add(10*time.Minute)); err == nil {
			t.Fatalf("invalid durable state silently reset: %s", invalid)
		}
		after, _ := os.ReadFile(filepath.Join(directory, "cleanup.prom"))
		if !bytes.Equal(before, after) {
			t.Fatal("failed observation refreshed healthy metrics")
		}
	}
}

func TestCleanupJournalFailedMetricWriteDoesNotConsumeFailure(t *testing.T) {
	directory := t.TempDir()
	if err := os.Mkdir(filepath.Join(directory, "cleanup.prom"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := RecordCleanupJournal(directory, strings.Repeat("a", 64), time.Now()); err == nil {
		t.Fatal("blocked metric publication reported success")
	}
	if _, err := os.Stat(filepath.Join(directory, "cleanup-state.json")); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("unpublished failure consumed its journal cursor")
	}
}

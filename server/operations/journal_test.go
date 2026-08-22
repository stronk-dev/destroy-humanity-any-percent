package operations

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func validJournalObservation() JournalObservation {
	return JournalObservation{SchemaVersion: 1, Population: "r006_release_workload", StartedAt: time.Unix(1_700_000_000, 0),
		CompletedAt: time.Unix(1_700_003_600, 0), ObjectiveCompleted: true, Samples: 12, ObservedBytes: 1_000,
		PeakBytesPerDay: 24_000, FilesystemBytes: 1_000_000, JournalMaxUseBytes: 336_000,
		JournalRetentionSeconds: int64(JournalRetention / time.Second), StorageAlertFraction: 0.8}
}

func TestJournalObservationRequiresMeasuredFourteenDayCapacity(t *testing.T) {
	valid := validJournalObservation()
	if err := ValidateJournalObservation(valid); err != nil {
		t.Fatal(err)
	}
	for name, mutate := range map[string]func(*JournalObservation){
		"guard exhausted":       func(value *JournalObservation) { value.GuardExhausted = true },
		"short retention":       func(value *JournalObservation) { value.JournalRetentionSeconds-- },
		"long retention":        func(value *JournalObservation) { value.JournalRetentionSeconds++ },
		"insufficient capacity": func(value *JournalObservation) { value.JournalMaxUseBytes-- },
		"late alert":            func(value *JournalObservation) { value.StorageAlertFraction = 1 },
		"unbounded raw IP":      func(value *JournalObservation) { value.RawIPEnabled = true },
	} {
		t.Run(name, func(t *testing.T) {
			candidate := valid
			mutate(&candidate)
			if err := ValidateJournalObservation(candidate); !errors.Is(err, ErrInvalid) {
				t.Fatalf("invalid observation accepted: %v", err)
			}
		})
	}
}

func TestJournalObservationRefusesUnimplementedRawIPSecuritySink(t *testing.T) {
	observation := validJournalObservation()
	observation.RawIPEnabled = true
	observation.SecuritySinkSeparate = true
	observation.SecurityRetentionSeconds = int64(SecurityRetention / time.Second)
	if err := ValidateJournalObservation(observation); !errors.Is(err, ErrInvalid) {
		t.Fatalf("self-asserted security sink accepted: %v", err)
	}
}

func TestJournalPolicyRendersExactMeasuredValuesAndRejectsUnknownFields(t *testing.T) {
	observation := validJournalObservation()
	template := []byte("[Journal]\nMaxRetentionSec=@@MAX_RETENTION_SEC@@\nSystemMaxUse=@@SYSTEM_MAX_USE@@\n")
	data, err := RenderJournalPolicy(template, observation)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "[Journal]\nMaxRetentionSec=1209600\nSystemMaxUse=336000\n" {
		t.Fatalf("wrong rendered policy: %q", data)
	}
	encoded, _ := json.Marshal(observation)
	encoded = append(encoded[:len(encoded)-1], []byte(`,"forged":true}`)...)
	if _, err := DecodeJournalObservation(encoded); !errors.Is(err, ErrInvalid) {
		t.Fatalf("unknown observation field accepted: %v", err)
	}
	if strings.Contains(string(data), "@@") {
		t.Fatal("unresolved policy token")
	}
}

func TestJournalPolicyRefusesExistingOrSymlinkTarget(t *testing.T) {
	directory := t.TempDir()
	existing := filepath.Join(directory, "existing")
	if err := os.WriteFile(existing, []byte("preserve"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := WriteJournalPolicy(existing, []byte("replace")); !errors.Is(err, ErrInvalid) {
		t.Fatalf("existing policy accepted: %v", err)
	}
	target := filepath.Join(directory, "target")
	if err := os.WriteFile(target, []byte("protected"), 0o600); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(directory, "link")
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}
	if err := WriteJournalPolicy(link, []byte("replace")); !errors.Is(err, ErrInvalid) {
		t.Fatalf("symlink policy accepted: %v", err)
	}
	data, _ := os.ReadFile(target)
	if string(data) != "protected" {
		t.Fatalf("symlink target changed: %q", data)
	}
}

package deploymentbackup

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"filippo.io/age"
)

func TestEncryptedBackupRoundTripAndFailClosedInputs(t *testing.T) {
	identity, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 8, 22, 12, 30, 0, 0, time.UTC)
	manifest := "sha256:" + strings.Repeat("a", 64)
	dump := []byte("PGDMP\x01fixture custom-format bytes")
	header, path, err := Create(CreateInput{Directory: t.TempDir(), BackupID: "20260822T120000Z-abcdef123456",
		ServerID: "01986666-b001-4000-8000-000000000001", ReleaseManifestSHA256: manifest, EpochID: 8,
		StartedAt: now.Add(-time.Minute), Now: func() time.Time { return now }, Recipient: identity.Recipient().String(), Dump: bytes.NewReader(dump)})
	if err != nil || header.PayloadBytes == 0 {
		t.Fatalf("header=%+v err=%v", header, err)
	}
	var restored bytes.Buffer
	if _, err := Restore(path, manifest, identity, &restored); err != nil || !bytes.Equal(restored.Bytes(), dump) {
		t.Fatalf("restore=%q err=%v", restored.Bytes(), err)
	}
	wrong, _ := age.GenerateX25519Identity()
	if _, err := Restore(path, manifest, wrong, &bytes.Buffer{}); !errors.Is(err, ErrInvalid) {
		t.Fatalf("wrong identity accepted: %v", err)
	}
	if _, err := Restore(path, "sha256:"+strings.Repeat("b", 64), identity, &bytes.Buffer{}); !errors.Is(err, ErrInvalid) {
		t.Fatalf("wrong manifest accepted: %v", err)
	}
	data, _ := os.ReadFile(path)
	data[len(data)-1] ^= 0xff
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Restore(path, manifest, identity, &bytes.Buffer{}); !errors.Is(err, ErrInvalid) {
		t.Fatalf("corruption accepted: %v", err)
	}
}

func TestCreateFailureLeavesNoValidOrPartialBackup(t *testing.T) {
	identity, _ := age.GenerateX25519Identity()
	directory := t.TempDir()
	failing := &errorReader{remaining: 4}
	_, _, err := Create(CreateInput{Directory: directory, BackupID: "20260822T120000Z-abcdef123456", ServerID: "server",
		ReleaseManifestSHA256: "sha256:" + strings.Repeat("a", 64), EpochID: 8, StartedAt: time.Now().Add(-time.Minute),
		Now: time.Now, Recipient: identity.Recipient().String(), Dump: failing})
	if err == nil {
		t.Fatal("interrupted dump accepted")
	}
	entries, _ := os.ReadDir(directory)
	if len(entries) != 0 {
		t.Fatalf("partial files survived: %v", entries)
	}
}

func TestRetentionPreservesNewestRecentDailyAndUnresolvedUpgrade(t *testing.T) {
	identity, _ := age.GenerateX25519Identity()
	directory := t.TempDir()
	now := time.Date(2026, 8, 22, 12, 0, 0, 0, time.UTC)
	paths := []string{}
	for index, ageHours := range []int{1, 8 * 24, 8*24 + 6, 31 * 24} {
		completed := now.Add(-time.Duration(ageHours) * time.Hour)
		id := completed.Format("20060102T150405Z") + "-abcdef12345" + string(rune('0'+index))
		header, path, err := Create(CreateInput{Directory: directory, BackupID: id, ServerID: "server", ReleaseManifestSHA256: "sha256:" + strings.Repeat("a", 64), EpochID: 8,
			StartedAt: completed.Add(-time.Minute), Now: func() time.Time { return completed }, Recipient: identity.Recipient().String(), Dump: strings.NewReader("PGDMP"), PreUpgrade: index == 3})
		if err != nil || header.BackupID != id {
			t.Fatal(err)
		}
		paths = append(paths, path)
	}
	invalid := filepath.Join(directory, "invalid.ccbackup")
	_ = os.WriteFile(invalid, []byte("broken"), 0o600)
	paths = append(paths, invalid)
	plan := PlanRetention(paths, now)
	if len(plan.Keep) != 3 || len(plan.Delete) != 1 || len(plan.Invalid) != 1 {
		t.Fatalf("plan=%+v", plan)
	}
}

type errorReader struct{ remaining int }

func (reader *errorReader) Read(target []byte) (int, error) {
	if reader.remaining <= 0 {
		return 0, errors.New("interrupted")
	}
	n := reader.remaining
	if n > len(target) {
		n = len(target)
	}
	reader.remaining -= n
	copy(target[:n], bytes.Repeat([]byte{'x'}, n))
	return n, nil
}

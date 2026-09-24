package deploymentbackup

import (
	"bytes"
	"encoding/json"
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

func TestBackupHeaderDecoderRejectsUnknownTrailingAndBackwardTime(t *testing.T) {
	now := time.Date(2026, 8, 23, 12, 0, 0, 0, time.UTC)
	header := Header{SchemaVersion: 1, BackupID: "20260823T120000Z-abcdef123456", ServerID: "server",
		ReleaseManifestSHA256: "sha256:" + strings.Repeat("a", 64), EpochID: 8, StartedAt: now,
		CompletedAt: now.Add(time.Minute), PayloadSHA256: "sha256:" + strings.Repeat("b", 64), PayloadBytes: 100, PreUpgrade: true}
	data, _ := json.Marshal(header)
	if decoded, err := DecodeHeader(data); err != nil || decoded.BackupID != header.BackupID {
		t.Fatalf("decoded=%+v err=%v", decoded, err)
	}
	unknown := append(data[:len(data)-1], []byte(`,"unknown":true}`)...)
	if _, err := DecodeHeader(unknown); !errors.Is(err, ErrInvalid) {
		t.Fatalf("unknown header field accepted: %v", err)
	}
	if _, err := DecodeHeader(append(data, []byte(`{}`)...)); !errors.Is(err, ErrInvalid) {
		t.Fatalf("trailing header accepted: %v", err)
	}
	header.CompletedAt = header.StartedAt.Add(-time.Second)
	data, _ = json.Marshal(header)
	if _, err := DecodeHeader(data); !errors.Is(err, ErrInvalid) {
		t.Fatalf("backward header interval accepted: %v", err)
	}
}

func TestPlaintextEnvelopeIdentityCannotBeRewritten(t *testing.T) {
	identity, _ := age.GenerateX25519Identity()
	manifest := "sha256:" + strings.Repeat("a", 64)
	_, path, err := Create(CreateInput{Directory: t.TempDir(), BackupID: "20260822T120000Z-abcdef123456",
		ServerID: "server-a", ReleaseManifestSHA256: manifest, EpochID: 8, StartedAt: time.Now().Add(-time.Minute),
		Now: time.Now, Recipient: identity.Recipient().String(), Dump: strings.NewReader("PGDMP")})
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	data = bytes.Replace(data, []byte(`"server_id":"server-a"`), []byte(`"server_id":"server-b"`), 1)
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Restore(path, manifest, identity, &bytes.Buffer{}); !errors.Is(err, ErrInvalid) {
		t.Fatalf("rewritten plaintext identity accepted: %v", err)
	}
}

func TestAuthenticatedIdentityComparesInstantsAcrossTimeLocations(t *testing.T) {
	instant := time.Date(2026, 8, 22, 12, 0, 0, 0, time.UTC)
	left := authenticatedHeader{BackupID: "id", ServerID: "server", StartedAt: instant, CompletedAt: instant}
	right := left
	right.StartedAt = instant.In(time.FixedZone("local-zero", 0))
	right.CompletedAt = right.StartedAt
	if left == right {
		t.Fatal("fixture did not exercise distinct time locations")
	}
	if !authenticatedEqual(left, right) {
		t.Fatal("equal instants were rejected because their location pointers differed")
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
	status := EvaluateSchedule(paths, now)
	if status.Missing || status.Late || len(status.Invalid) != 1 || !status.LatestCompletedAt.Equal(now.Add(-time.Hour)) {
		t.Fatalf("status=%+v", status)
	}
	if _, err := ApplyRetention(paths, now); !errors.Is(err, ErrInvalid) {
		t.Fatalf("invalid population was deleted: %v", err)
	}
}

func TestScheduleReportsMissingAndLateWithoutCoasting(t *testing.T) {
	now := time.Date(2026, 8, 22, 12, 0, 0, 0, time.UTC)
	if status := EvaluateSchedule(nil, now); !status.Missing || status.Late {
		t.Fatalf("missing status=%+v", status)
	}
	identity, _ := age.GenerateX25519Identity()
	completed := now.Add(-6*time.Hour - time.Second)
	_, path, err := Create(CreateInput{Directory: t.TempDir(), BackupID: "20260822T055959Z-abcdef123456", ServerID: "server",
		ReleaseManifestSHA256: "sha256:" + strings.Repeat("a", 64), EpochID: 8, StartedAt: completed.Add(-time.Minute),
		Now: func() time.Time { return completed }, Recipient: identity.Recipient().String(), Dump: strings.NewReader("PGDMP")})
	if err != nil {
		t.Fatal(err)
	}
	if status := EvaluateSchedule([]string{path}, now); status.Missing || !status.Late {
		t.Fatalf("late status=%+v", status)
	}
}

func TestApplyRetentionDeletesOnlyPlannedExpiredBackup(t *testing.T) {
	identity, _ := age.GenerateX25519Identity()
	directory := t.TempDir()
	now := time.Date(2026, 8, 22, 12, 0, 0, 0, time.UTC)
	paths := []string{}
	for index, ageHours := range []int{1, 31 * 24} {
		completed := now.Add(-time.Duration(ageHours) * time.Hour)
		id := completed.Format("20060102T150405Z") + "-11111111111" + string(rune('0'+index))
		_, path, err := Create(CreateInput{Directory: directory, BackupID: id, ServerID: "server",
			ReleaseManifestSHA256: "sha256:" + strings.Repeat("a", 64), EpochID: 8,
			StartedAt: completed.Add(-time.Minute), Now: func() time.Time { return completed },
			Recipient: identity.Recipient().String(), Dump: strings.NewReader("PGDMP")})
		if err != nil {
			t.Fatal(err)
		}
		paths = append(paths, path)
	}
	plan := PlanRetention(paths, now)
	if len(plan.Keep) != 1 || len(plan.Delete) != 1 {
		t.Fatalf("plan=%+v", plan)
	}
	applied, err := ApplyRetention(paths, now)
	if err != nil {
		t.Fatal(err)
	}
	if len(applied.Keep) != 1 || len(applied.Delete) != 1 {
		t.Fatalf("applied=%+v", applied)
	}
	if _, err := os.Stat(plan.Keep[0]); err != nil {
		t.Fatalf("protected backup removed: %v", err)
	}
	if _, err := os.Stat(plan.Delete[0]); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expired backup survived: %v", err)
	}
}

func TestRetentionNeverDeletesTheOnlyNewestBackupEvenAfterThirtyDays(t *testing.T) {
	identity, _ := age.GenerateX25519Identity()
	now := time.Date(2026, 8, 22, 12, 0, 0, 0, time.UTC)
	completed := now.Add(-31 * 24 * time.Hour)
	_, path, err := Create(CreateInput{Directory: t.TempDir(), BackupID: "20260722T120000Z-222222222222", ServerID: "server",
		ReleaseManifestSHA256: "sha256:" + strings.Repeat("a", 64), EpochID: 8,
		StartedAt: completed.Add(-time.Minute), Now: func() time.Time { return completed },
		Recipient: identity.Recipient().String(), Dump: strings.NewReader("PGDMP")})
	if err != nil {
		t.Fatal(err)
	}
	plan := PlanRetention([]string{path}, now)
	if len(plan.Keep) != 1 || len(plan.Delete) != 0 {
		t.Fatalf("newest expired backup was not protected: %+v", plan)
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

func TestHeaderChecksumGuardsScheduleRetentionAndRestore(t *testing.T) {
	identity, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 9, 24, 12, 0, 0, 0, time.UTC)
	manifest := "sha256:" + strings.Repeat("a", 64)
	create := func(t *testing.T, directory string, dump []byte) (Header, string) {
		t.Helper()
		header, path, err := Create(CreateInput{Directory: directory, BackupID: "20260924T110000Z-abcdef123456",
			ServerID: "01986666-b001-4000-8000-000000000001", ReleaseManifestSHA256: manifest, EpochID: 8,
			StartedAt: now.Add(-time.Minute), Now: func() time.Time { return now }, Recipient: identity.Recipient().String(), Dump: bytes.NewReader(dump)})
		if err != nil {
			t.Fatal(err)
		}
		return header, path
	}
	splitEnvelope := func(t *testing.T, path string) ([]byte, []byte) {
		t.Helper()
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		first := bytes.IndexByte(data, '\n')
		second := bytes.IndexByte(data[first+1:], '\n') + first + 1
		return data[:second+1], data[second+1:]
	}

	t.Run("truncated payload with intact header", func(t *testing.T) {
		directory := t.TempDir()
		_, path := create(t, directory, []byte("PGDMP\x01truncation fixture"))
		prefix, payload := splitEnvelope(t, path)
		if err := os.WriteFile(path, append(append([]byte(nil), prefix...), payload[:len(payload)-1]...), 0o600); err != nil {
			t.Fatal(err)
		}
		if _, err := ReadHeader(path); !errors.Is(err, ErrInvalid) {
			t.Fatalf("truncated payload header accepted: %v", err)
		}
		if plan := PlanRetention([]string{path}, now); len(plan.Invalid) != 1 || len(plan.Keep) != 0 {
			t.Fatalf("truncated backup counted as retainable: %+v", plan)
		}
		if status := EvaluateSchedule([]string{path}, now); !status.Missing || len(status.Invalid) != 1 {
			t.Fatalf("truncated backup counted as newest valid: %+v", status)
		}
	})

	t.Run("swapped authenticated payload", func(t *testing.T) {
		header, original := create(t, t.TempDir(), []byte("PGDMP\x01original"))
		_, other := create(t, t.TempDir(), []byte("PGDMP\x01substituted dump of the same length"))
		prefix, _ := splitEnvelope(t, original)
		_, substitute := splitEnvelope(t, other)
		if err := os.WriteFile(original, append(append([]byte(nil), prefix...), substitute...), 0o600); err != nil {
			t.Fatal(err)
		}
		// Age authentication and the encrypted header copy both still verify;
		// only the plaintext header's payload length/checksum binds the bytes.
		if _, err := Restore(original, manifest, identity, &bytes.Buffer{}); !errors.Is(err, ErrInvalid) {
			t.Fatalf("substituted payload restored under header %s: %v", header.PayloadSHA256, err)
		}
		if _, err := ReadHeader(original); !errors.Is(err, ErrInvalid) {
			t.Fatalf("substituted payload header accepted: %v", err)
		}
	})
}

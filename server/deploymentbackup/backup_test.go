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

func TestBackupEnvelopeReadersEnforceSingleBoundedHeader(t *testing.T) {
	identity, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 10, 12, 12, 0, 0, 0, time.UTC)
	manifest := "sha256:" + strings.Repeat("a", 64)
	dump := []byte("PGDMP\x01envelope framing control")
	_, original, err := Create(CreateInput{Directory: t.TempDir(), BackupID: "20261012T120000Z-abcdef123456",
		ServerID: "server-a", ReleaseManifestSHA256: manifest, EpochID: 8,
		StartedAt: now.Add(-time.Minute), Now: func() time.Time { return now }, Recipient: identity.Recipient().String(), Dump: bytes.NewReader(dump)})
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(original)
	if err != nil {
		t.Fatal(err)
	}
	parts := bytes.SplitN(data, []byte{'\n'}, 3)
	if len(parts) != 3 {
		t.Fatal("healthy backup has no complete envelope framing")
	}
	for _, row := range []struct {
		name, suffix string
		valid        bool
	}{
		{"healthy", "", true},
		{"json-whitespace", " \t\r", true},
		{"maximum-line", "", true},
		{"second-object", "{}", false},
		{"second-value", "true", false},
		{"garbage", "!", false},
		{"oversized-line", strings.Repeat(" ", 64<<10), false},
	} {
		t.Run(row.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "candidate.ccbackup")
			changed := append([]byte(magic), parts[1]...)
			suffix := row.suffix
			if row.name == "maximum-line" {
				suffix = strings.Repeat(" ", (64<<10)-len(parts[1])-1)
			}
			changed = append(changed, suffix...)
			changed = append(append(changed, '\n'), parts[2]...)
			if err := os.WriteFile(path, changed, 0o600); err != nil {
				t.Fatal(err)
			}
			_, readErr := ReadHeader(path)
			var output bytes.Buffer
			_, restoreErr := Restore(path, manifest, identity, &output)
			if row.valid {
				if readErr != nil || restoreErr != nil || !bytes.Equal(output.Bytes(), dump) {
					t.Fatalf("valid envelope refused: read=%v restore=%v", readErr, restoreErr)
				}
				return
			}
			if !errors.Is(readErr, ErrInvalid) || !errors.Is(restoreErr, ErrInvalid) || output.Len() != 0 {
				t.Fatalf("malformed header admitted: read=%v restore=%v output_bytes=%d", readErr, restoreErr, output.Len())
			}
			if status := EvaluateSchedule([]string{path}, now); !status.Missing || len(status.Invalid) != 1 {
				t.Fatal("malformed envelope counted as a scheduled backup")
			}
			if plan := PlanRetention([]string{path}, now); len(plan.Invalid) != 1 || len(plan.Keep)+len(plan.Delete) != 0 {
				t.Fatal("malformed envelope entered the retention population")
			}
		})
	}
}

func TestBackupRestoreEnforcesSingleBoundedAuthenticatedMetadata(t *testing.T) {
	identity, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 10, 12, 12, 0, 0, 0, time.UTC)
	dump := []byte("PGDMP\x01authenticated framing control")
	for _, row := range []struct {
		name, suffix string
		valid        bool
	}{
		{"healthy", "", true},
		{"json-whitespace", " \t\r", true},
		{"maximum-line", "", true},
		{"second-object", "{}", false},
		{"second-value", "true", false},
		{"garbage", "!", false},
		{"oversized-line", strings.Repeat(" ", 64<<10), false},
	} {
		t.Run(row.name, func(t *testing.T) {
			header := Header{SchemaVersion: 1, BackupID: "20261012T120000Z-abcdef123456", ServerID: "server-a",
				ReleaseManifestSHA256: "sha256:" + strings.Repeat("a", 64), EpochID: 8,
				StartedAt: now.Add(-time.Minute), CompletedAt: now}
			metadata, err := json.Marshal(authenticatedFrom(header))
			if err != nil {
				t.Fatal(err)
			}
			var payload bytes.Buffer
			encrypted, err := age.Encrypt(&payload, identity.Recipient())
			if err != nil {
				t.Fatal(err)
			}
			suffix := row.suffix
			if row.name == "maximum-line" {
				suffix = strings.Repeat(" ", (64<<10)-len(metadata)-1)
			}
			plain := append(append(append(metadata, suffix...), '\n'), dump...)
			if _, err := encrypted.Write(plain); err != nil {
				t.Fatal(err)
			}
			if err := encrypted.Close(); err != nil {
				t.Fatal(err)
			}
			header.PayloadSHA256, header.PayloadBytes = digest(payload.Bytes()), int64(payload.Len())
			headerBytes, err := json.Marshal(header)
			if err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(t.TempDir(), "candidate.ccbackup")
			data := append(append(append([]byte(magic), headerBytes...), '\n'), payload.Bytes()...)
			if err := os.WriteFile(path, data, 0o600); err != nil {
				t.Fatal(err)
			}
			// The encrypted body is genuine and checksummed: only metadata parsing
			// can distinguish these cases, not a corrupt-payload shortcut.
			if actual, err := ReadHeader(path); err != nil || actual != header {
				t.Fatalf("authenticated control has an invalid outer envelope: %v", err)
			}
			var output bytes.Buffer
			_, restoreErr := Restore(path, header.ReleaseManifestSHA256, identity, &output)
			if row.valid {
				if restoreErr != nil || !bytes.Equal(output.Bytes(), dump) {
					t.Fatalf("valid metadata refused: %v", restoreErr)
				}
			} else if !errors.Is(restoreErr, ErrInvalid) || output.Len() != 0 {
				t.Fatalf("malformed authenticated metadata admitted: err=%v output_bytes=%d", restoreErr, output.Len())
			}
		})
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
		_, other := create(t, t.TempDir(), []byte("PGDMP\x01altered!"))
		prefix, originalPayload := splitEnvelope(t, original)
		_, substitute := splitEnvelope(t, other)
		if int64(len(originalPayload)) != header.PayloadBytes || len(substitute) != len(originalPayload) || bytes.Equal(substitute, originalPayload) {
			t.Fatalf("checksum fixture must substitute different equal-length ciphertext: original=%d substitute=%d header=%d", len(originalPayload), len(substitute), header.PayloadBytes)
		}
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

func TestCreateVerifiesTheCommittedEnvelopeBytes(t *testing.T) {
	identity, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	beforeEnvelopeCommit = func(path string) error {
		info, err := os.Stat(path)
		if err != nil {
			return err
		}
		// A torn write: the envelope lost its final bytes after close.
		return os.Truncate(path, info.Size()-1)
	}
	t.Cleanup(func() { beforeEnvelopeCommit = nil })
	now := time.Date(2026, 9, 24, 12, 0, 0, 0, time.UTC)
	_, path, err := Create(CreateInput{Directory: directory, BackupID: "20260924T120000Z-abcdef123456",
		ServerID: "01986666-b001-4000-8000-000000000001", ReleaseManifestSHA256: "sha256:" + strings.Repeat("a", 64), EpochID: 8,
		StartedAt: now.Add(-time.Minute), Now: func() time.Time { return now }, Recipient: identity.Recipient().String(),
		Dump: bytes.NewReader([]byte("PGDMP\x01torn envelope"))})
	if !errors.Is(err, ErrInvalid) || path != "" {
		t.Fatalf("torn envelope committed: path=%q err=%v", path, err)
	}
	entries, err := os.ReadDir(directory)
	if err != nil || len(entries) != 0 {
		t.Fatalf("torn envelope left files behind: %v err=%v", entries, err)
	}
}

func TestCreateDoesNotOverwriteAConcurrentCompletedBackup(t *testing.T) {
	identity, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	now := time.Date(2026, 10, 8, 12, 0, 0, 0, time.UTC)
	manifest := "sha256:" + strings.Repeat("a", 64)
	input := CreateInput{Directory: directory, BackupID: "20261008T120000Z-abcdef123456",
		ServerID: "server", ReleaseManifestSHA256: manifest, EpochID: 8,
		StartedAt: now.Add(-time.Minute), Now: func() time.Time { return now }, Recipient: identity.Recipient().String()}
	firstDump := []byte("PGDMP\x01first writer")
	winnerDump := []byte("PGDMP\x01completed concurrent writer")
	var winnerPath string
	var winnerBytes []byte
	var winnerHeader Header
	// Deterministically finish another real encrypted backup after the first
	// writer's destination check, without sleeps, goroutines or fake envelopes.
	beforeEnvelopeCommit = func(_ string) error {
		beforeEnvelopeCommit = nil
		other := input
		other.Dump = bytes.NewReader(winnerDump)
		winnerHeader, winnerPath, err = Create(other)
		if err != nil {
			return err
		}
		winnerBytes, err = os.ReadFile(winnerPath)
		return err
	}
	t.Cleanup(func() { beforeEnvelopeCommit = nil })
	input.Dump = bytes.NewReader(firstDump)
	header, path, createErr := Create(input)
	if len(winnerBytes) == 0 {
		t.Fatalf("concurrent writer did not commit a valid fixture: %v", createErr)
	}
	current, err := os.ReadFile(winnerPath)
	if err != nil || !bytes.Equal(current, winnerBytes) {
		t.Fatalf("completed concurrent backup was overwritten: read=%v create=%v", err, createErr)
	}
	if !errors.Is(createErr, ErrInvalid) || path != "" || header != (Header{}) {
		t.Fatalf("colliding writer reported success: path=%q err=%v", path, createErr)
	}
	var restored bytes.Buffer
	if actual, err := Restore(winnerPath, manifest, identity, &restored); err != nil || actual != winnerHeader || !bytes.Equal(restored.Bytes(), winnerDump) {
		t.Fatalf("winning backup did not retain its exact restored identity: %v", err)
	}
	entries, err := os.ReadDir(directory)
	if err != nil || len(entries) != 1 || entries[0].Name() != filepath.Base(winnerPath) {
		t.Fatalf("collision left partial backup files: %v err=%v", entries, err)
	}
	// An already-existing destination still refuses without touching its bytes.
	input.Dump = bytes.NewReader(firstDump)
	if _, _, err := Create(input); !errors.Is(err, ErrInvalid) {
		t.Fatalf("existing completed backup accepted: %v", err)
	}
	if current, err := os.ReadFile(winnerPath); err != nil || !bytes.Equal(current, winnerBytes) {
		t.Fatalf("existing backup changed after refusal: %v", err)
	}
}

func TestEnvelopeRenameRefusesEveryExistingDestination(t *testing.T) {
	for _, kind := range []string{"absent", "file", "symlink", "directory"} {
		t.Run(kind, func(t *testing.T) {
			directory := t.TempDir()
			source := filepath.Join(directory, ".backup-envelope-123.tmp")
			destination := filepath.Join(directory, "completed.ccbackup")
			victim := filepath.Join(directory, "symlink-target")
			if err := os.WriteFile(source, []byte("verified envelope"), 0o600); err != nil {
				t.Fatal(err)
			}
			var err error
			switch kind {
			case "file":
				err = os.WriteFile(destination, []byte("existing envelope"), 0o600)
			case "symlink":
				err = os.WriteFile(victim, []byte("unchanged target"), 0o600)
				if err == nil {
					err = os.Symlink(victim, destination)
				}
			case "directory":
				err = os.Mkdir(destination, 0o700)
			}
			if err != nil {
				t.Fatal(err)
			}
			err = renameEnvelopeNoReplace(source, destination)
			if kind == "absent" {
				if err != nil {
					t.Fatal(err)
				}
				if _, err := os.Lstat(source); !errors.Is(err, os.ErrNotExist) {
					t.Fatalf("successful rename retained source: %v", err)
				}
				if data, err := os.ReadFile(destination); err != nil || string(data) != "verified envelope" {
					t.Fatalf("successful rename changed bytes: %v", err)
				}
				return
			}
			if !errors.Is(err, os.ErrExist) {
				t.Fatalf("existing %s was not refused as a collision: %v", kind, err)
			}
			if data, err := os.ReadFile(source); err != nil || string(data) != "verified envelope" {
				t.Fatalf("refused rename consumed or changed source: %v", err)
			}
			switch kind {
			case "file":
				if data, err := os.ReadFile(destination); err != nil || string(data) != "existing envelope" {
					t.Fatalf("existing envelope changed: %v", err)
				}
			case "symlink":
				if target, err := os.Readlink(destination); err != nil || target != victim {
					t.Fatalf("existing symlink changed: %v", err)
				}
				if data, err := os.ReadFile(victim); err != nil || string(data) != "unchanged target" {
					t.Fatalf("symlink target changed: %v", err)
				}
			case "directory":
				if info, err := os.Lstat(destination); err != nil || !info.IsDir() {
					t.Fatalf("existing directory changed: %v", err)
				}
			}
		})
	}
}

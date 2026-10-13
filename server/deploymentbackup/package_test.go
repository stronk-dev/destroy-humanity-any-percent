package deploymentbackup

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"cloud-clicker/server/internal/backupfixture"
	"filippo.io/age"
)

func TestPackageRoundTripBindsDumpManifestEpochAndHeader(t *testing.T) {
	identity, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatal(err)
	}
	epoch := fixtureEpoch(t, 8)
	manifest := fixtureReleaseManifest(t, epoch)
	dump := []byte("PGDMP\x01custom database fixture")
	now := time.Date(2026, 8, 22, 15, 0, 0, 0, time.UTC)
	header, path, err := CreatePackage(PackageInput{
		Directory: t.TempDir(), BackupID: "20260822T145900Z-abcdef123456", ServerID: "01986666-b001-4000-8000-000000000001",
		StartedAt: now.Add(-time.Minute), Now: func() time.Time { return now }, Recipient: identity.Recipient().String(),
		Dump: bytes.NewReader(dump), ReleaseManifest: manifest, EpochDeclaration: epoch, PreUpgrade: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	destination := t.TempDir()
	extracted, err := ExtractPackage(path, digest(manifest), identity, destination)
	if err != nil {
		t.Fatal(err)
	}
	restored, err := os.ReadFile(extracted.DumpPath)
	if err != nil || !bytes.Equal(restored, dump) || extracted.Header != header || extracted.Metadata.DatabaseDumpSHA256 != digest(dump) {
		t.Fatalf("restored=%q header=%+v metadata=%+v err=%v", restored, extracted.Header, extracted.Metadata, err)
	}
}

func TestPackageCreationUsesTheCompleteDumpRegardlessOfReaderPosition(t *testing.T) {
	identity, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatal(err)
	}
	epoch := fixtureEpoch(t, 8)
	manifest := fixtureReleaseManifest(t, epoch)
	dump := []byte("PGDMP complete dump with distinct prefix and suffix")
	now := time.Date(2026, 10, 12, 15, 0, 0, 0, time.UTC)
	for _, position := range []int64{0, 5, int64(len(dump))} {
		t.Run(fmt.Sprintf("position_%d", position), func(t *testing.T) {
			reader := bytes.NewReader(dump)
			if _, err := reader.Seek(position, io.SeekStart); err != nil {
				t.Fatal(err)
			}
			header, path, err := CreatePackage(PackageInput{
				Directory: t.TempDir(), BackupID: "20261012T145900Z-abcdef123456", ServerID: "server",
				StartedAt: now.Add(-time.Minute), Now: func() time.Time { return now }, Recipient: identity.Recipient().String(),
				Dump: reader, ReleaseManifest: manifest, EpochDeclaration: epoch,
			})
			if err != nil {
				t.Fatalf("complete readable dump refused: %v", err)
			}
			extracted, err := ExtractPackage(path, digest(manifest), identity, t.TempDir())
			if err != nil {
				t.Fatalf("creation reported success for an unrestorable package: %v", err)
			}
			actual, err := os.ReadFile(extracted.DumpPath)
			if err != nil || !bytes.Equal(actual, dump) || extracted.Header != header ||
				extracted.Metadata.DatabaseDumpSHA256 != digest(dump) || extracted.Metadata.DatabaseDumpBytes != int64(len(dump)) {
				t.Fatalf("completed package did not bind the complete dump: bytes=%q metadata=%+v err=%v", actual, extracted.Metadata, err)
			}
		})
	}
}

// Replace the test input only when creation rewinds after reading its first
// pass. This models a changed source between hashing and packaging without
// altering ciphertext, metadata, clocks, or the production code.
type packageChangingDump struct {
	Reader      *bytes.Reader
	replacement []byte
	read        bool
	changes     int
}

func (dump *packageChangingDump) Read(data []byte) (int, error) {
	dump.read = true
	return dump.Reader.Read(data)
}

func (dump *packageChangingDump) Seek(offset int64, whence int) (int64, error) {
	if dump.read && offset == 0 && whence == io.SeekStart {
		dump.Reader = bytes.NewReader(dump.replacement)
		dump.read = false
		dump.changes++
	}
	return dump.Reader.Seek(offset, whence)
}

func TestPackageCreationRefusesChangedDumpBeforeCommitting(t *testing.T) {
	identity, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatal(err)
	}
	epoch := fixtureEpoch(t, 8)
	manifest := fixtureReleaseManifest(t, epoch)
	dump := []byte("PGDMP original dump bytes")
	now := time.Date(2026, 10, 12, 15, 0, 0, 0, time.UTC)
	for _, fixture := range []struct {
		name        string
		replacement []byte
		valid       bool
	}{
		{"unchanged", dump, true},
		{"same_length_changed", []byte("PGDMP modified dump bytes"), false},
		{"truncated", dump[:len(dump)-1], false},
		{"appended", append(bytes.Clone(dump), 'x'), false},
	} {
		t.Run(fixture.name, func(t *testing.T) {
			directory := t.TempDir()
			source := &packageChangingDump{Reader: bytes.NewReader(dump), replacement: fixture.replacement}
			header, path, err := CreatePackage(PackageInput{
				Directory: directory, BackupID: "20261012T145900Z-abcdef123456", ServerID: "server",
				StartedAt: now.Add(-time.Minute), Now: func() time.Time { return now }, Recipient: identity.Recipient().String(),
				Dump:            source,
				ReleaseManifest: manifest, EpochDeclaration: epoch,
			})
			if source.changes != 1 {
				t.Fatalf("invalid fixture: expected one between-pass replacement, got %d", source.changes)
			}
			if fixture.valid {
				extracted, extractErr := ExtractPackage(path, digest(manifest), identity, t.TempDir())
				if err != nil || extractErr != nil || extracted.Header != header {
					t.Fatalf("intact two-pass input failed: create=%v extract=%v", err, extractErr)
				}
				actual, readErr := os.ReadFile(extracted.DumpPath)
				if readErr != nil || !bytes.Equal(actual, dump) {
					t.Fatalf("intact dump changed: %v", readErr)
				}
				return
			}
			if !errors.Is(err, ErrInvalid) || path != "" || header != (Header{}) {
				t.Fatalf("changed input reported a completed backup: header=%+v path=%q err=%v", header, path, err)
			}
			entries, readErr := os.ReadDir(directory)
			if readErr != nil || len(entries) != 0 {
				t.Fatalf("failed creation retained files: entries=%v err=%v", entries, readErr)
			}
		})
	}
}

func TestPackageRejectsWrongEpochArtifactAndNonEmptyExtractionTarget(t *testing.T) {
	identity, _ := age.GenerateX25519Identity()
	epoch := fixtureEpoch(t, 8)
	manifest := fixtureReleaseManifest(t, epoch)
	wrongEpoch := fixtureEpoch(t, 7)
	input := PackageInput{Directory: t.TempDir(), BackupID: "20260822T145900Z-abcdef123456", ServerID: "server",
		StartedAt: time.Now().Add(-time.Minute), Now: time.Now, Recipient: identity.Recipient().String(),
		Dump: bytes.NewReader([]byte("PGDMP")), ReleaseManifest: manifest, EpochDeclaration: wrongEpoch}
	if _, _, err := CreatePackage(input); !errors.Is(err, ErrInvalid) {
		t.Fatalf("wrong epoch accepted: %v", err)
	}
	input.EpochDeclaration = []byte("{\"schema_version\":1,\"current_epoch_id\":8}\n")
	if _, _, err := CreatePackage(input); !errors.Is(err, ErrInvalid) {
		t.Fatalf("reduced noncanonical epoch accepted: %v", err)
	}
	input.EpochDeclaration = epoch
	_, path, err := CreatePackage(input)
	if err != nil {
		t.Fatal(err)
	}
	nonEmpty := t.TempDir()
	if err := os.WriteFile(filepath.Join(nonEmpty, "existing"), []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := ExtractPackage(path, digest(manifest), identity, nonEmpty); !errors.Is(err, ErrInvalid) {
		t.Fatalf("non-empty target accepted: %v", err)
	}
}

func TestPackageRejectsTruncatedEnvelope(t *testing.T) {
	identity, _ := age.GenerateX25519Identity()
	epoch := fixtureEpoch(t, 8)
	manifest := fixtureReleaseManifest(t, epoch)
	_, path, err := CreatePackage(PackageInput{Directory: t.TempDir(), BackupID: "20260822T145900Z-abcdef123456", ServerID: "server",
		StartedAt: time.Now().Add(-time.Minute), Now: time.Now, Recipient: identity.Recipient().String(),
		Dump: bytes.NewReader([]byte("PGDMP payload")), ReleaseManifest: manifest, EpochDeclaration: epoch})
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data[:len(data)-8], 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := ExtractPackage(path, digest(manifest), identity, t.TempDir()); !errors.Is(err, ErrInvalid) {
		t.Fatalf("truncated envelope accepted: %v", err)
	}
}

func TestPackageConsumesAuthenticatedStreamAfterTarEnd(t *testing.T) {
	identity, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatal(err)
	}
	epoch := fixtureEpoch(t, 8)
	manifest := fixtureReleaseManifest(t, epoch)
	now := time.Date(2026, 10, 12, 15, 0, 0, 0, time.UTC)
	dump := []byte("PGDMP trailing-stream fixture")
	header, path, err := CreatePackage(PackageInput{
		Directory: t.TempDir(), BackupID: "20261012T145900Z-abcdef123456", ServerID: "server",
		StartedAt: now.Add(-time.Minute), Now: func() time.Time { return now }, Recipient: identity.Recipient().String(),
		Dump: bytes.NewReader(dump), ReleaseManifest: manifest, EpochDeclaration: epoch,
	})
	if err != nil {
		t.Fatal(err)
	}
	var original bytes.Buffer
	if _, err := Restore(path, digest(manifest), identity, &original); err != nil {
		t.Fatal(err)
	}
	for _, fixture := range []struct {
		name string
		tail []byte
	}{
		{"exact_package", nil},
		{"one_extra_byte", []byte("x")},
		{"extra_zero_block", make([]byte, 512)},
		{"second_archive", original.Bytes()},
	} {
		t.Run(fixture.name, func(t *testing.T) {
			plain := append(append([]byte(nil), original.Bytes()...), fixture.tail...)
			_, path, err := Create(CreateInput{
				Directory: t.TempDir(), BackupID: header.BackupID, ServerID: header.ServerID,
				ReleaseManifestSHA256: header.ReleaseManifestSHA256, EpochID: header.EpochID,
				StartedAt: header.StartedAt, Now: func() time.Time { return header.CompletedAt },
				Recipient: identity.Recipient().String(), Dump: bytes.NewReader(plain),
			})
			if err != nil {
				t.Fatal(err)
			}
			// This is a valid age stream and matching outer checksum: only exact
			// package framing can reject the additional plaintext after tar EOF.
			var decrypted bytes.Buffer
			if _, err := Restore(path, digest(manifest), identity, &decrypted); err != nil || !bytes.Equal(decrypted.Bytes(), plain) {
				t.Fatalf("encrypted fixture not intact: %v", err)
			}
			destination := t.TempDir()
			type outcome struct {
				result ExtractedPackage
				err    error
			}
			completed := make(chan outcome, 1)
			go func() {
				result, err := ExtractPackage(path, digest(manifest), identity, destination)
				completed <- outcome{result, err}
			}()
			select {
			case actual := <-completed:
				if fixture.tail != nil {
					if !errors.Is(actual.err, ErrInvalid) || actual.result.DumpPath != "" || actual.result.Header != (Header{}) {
						t.Fatalf("trailing package data accepted or returned as usable: %+v err=%v", actual.result, actual.err)
					}
					return
				}
				restored, readErr := os.ReadFile(actual.result.DumpPath)
				if actual.err != nil || readErr != nil || !bytes.Equal(restored, dump) {
					t.Fatalf("exact package refused: restore=%v read=%v", actual.err, readErr)
				}
			case <-time.After(5 * time.Second):
				t.Fatal("package extraction stalled after tar EOF instead of completing/refusing")
			}
		})
	}
}

func TestObjectiveMeasurementFailsIncompleteLateAndSlowObservations(t *testing.T) {
	now := time.Date(2026, 8, 22, 16, 0, 0, 0, time.UTC)
	header := Header{SchemaVersion: 1, BackupID: "20260822T145900Z-abcdef123456", ServerID: "server",
		ReleaseManifestSHA256: "sha256:" + strings.Repeat("a", 64), EpochID: 8,
		StartedAt: now.Add(-time.Hour), CompletedAt: now.Add(-30 * time.Minute),
		PayloadSHA256: "sha256:" + strings.Repeat("b", 64), PayloadBytes: 1}
	for name, observation := range map[string]ObjectiveObservation{
		"incomplete": {},
		"late":       {IncidentAt: now.Add(6 * time.Hour), RestoreStartedAt: now.Add(6 * time.Hour), SmokePassedAt: now.Add(6*time.Hour + time.Minute)},
		"slow":       {IncidentAt: now, RestoreStartedAt: now, SmokePassedAt: now.Add(4*time.Hour + time.Second)},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := MeasureObjectives(header, observation); !errors.Is(err, ErrInvalid) {
				t.Fatalf("invalid observation accepted: %+v err=%v", observation, err)
			}
		})
	}
	measurement, err := MeasureObjectives(header, ObjectiveObservation{IncidentAt: now, RestoreStartedAt: now, SmokePassedAt: now.Add(time.Minute)})
	if err != nil || measurement.RPOSeconds != 1800 || measurement.RTOSeconds != 60 {
		t.Fatalf("measurement=%+v err=%v", measurement, err)
	}
}

func TestObjectiveMeasurementEnforcesExactDurationBounds(t *testing.T) {
	completed := time.Date(2026, 10, 12, 12, 0, 0, 0, time.UTC)
	header := Header{SchemaVersion: 1, BackupID: "20261012T115900Z-abcdef123456", ServerID: "server",
		ReleaseManifestSHA256: "sha256:" + strings.Repeat("a", 64), EpochID: 8,
		StartedAt: completed.Add(-time.Minute), CompletedAt: completed,
		PayloadSHA256: "sha256:" + strings.Repeat("b", 64), PayloadBytes: 1}
	for _, objective := range []struct {
		name  string
		bound time.Duration
	}{
		{"rpo", 6 * time.Hour},
		{"rto", 4 * time.Hour},
	} {
		for _, boundary := range []struct {
			name     string
			duration time.Duration
			valid    bool
		}{
			{"zero", 0, true},
			{"one_ns_below", objective.bound - time.Nanosecond, true},
			{"exact", objective.bound, true},
			{"one_ns_over", objective.bound + time.Nanosecond, false},
			{"999ms_over", objective.bound + 999*time.Millisecond, false},
			{"one_second_over", objective.bound + time.Second, false},
		} {
			t.Run(objective.name+"_"+boundary.name, func(t *testing.T) {
				rpo, rto := time.Minute, time.Minute
				if objective.name == "rpo" {
					rpo = boundary.duration
				} else {
					rto = boundary.duration
				}
				incident := completed.Add(rpo)
				restore := incident.Add(time.Minute)
				measurement, err := MeasureObjectives(header, ObjectiveObservation{
					IncidentAt: incident, RestoreStartedAt: restore, SmokePassedAt: restore.Add(rto)})
				if !boundary.valid {
					if !errors.Is(err, ErrInvalid) || measurement != (ObjectiveMeasurement{}) {
						t.Fatalf("out-of-bound duration accepted or partial result returned: duration=%s measurement=%+v err=%v", boundary.duration, measurement, err)
					}
					return
				}
				if err != nil || measurement.RPOSeconds != int64(rpo/time.Second) || measurement.RTOSeconds != int64(rto/time.Second) {
					t.Fatalf("valid duration refused or existing whole-second report changed: measurement=%+v err=%v", measurement, err)
				}
			})
		}
	}
}

func fixtureReleaseManifest(t *testing.T, epoch []byte) []byte {
	return fixtureReleaseManifestForMigration(t, epoch, 1)
}

func fixtureEpoch(t *testing.T, current int64) []byte { return backupfixture.Epoch(t, current) }

func fixtureReleaseManifestForMigration(t *testing.T, epoch []byte, migration int) []byte {
	return backupfixture.Manifest(t, epoch, migration)
}

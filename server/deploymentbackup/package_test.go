package deploymentbackup

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"cloud-clicker/server/epochseed"
	"cloud-clicker/server/releasepackage"
	"cloud-clicker/server/save"
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

func fixtureEpoch(t *testing.T, current int64) []byte {
	t.Helper()
	hash := "sha256:" + strings.Repeat("c", 64)
	epochs := make([]epochseed.Epoch, current)
	for index := range epochs {
		id := int64(index + 1)
		epochs[index] = epochseed.Epoch{ID: id, Name: fmt.Sprintf("Epoch %d", id),
			ChangelogRef: fmt.Sprintf("changelog/epoch-%d.md", id), AcceptedHashes: []string{hash}}
	}
	seed := epochseed.Seed{SchemaVersion: 1, CurrentEpochID: current,
		Artifacts: []epochseed.Artifact{{Name: "fixture", Path: "balance/fixture.json"}}, Epochs: epochs}
	if err := epochseed.Validate(seed); err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(seed)
	if err != nil {
		t.Fatal(err)
	}
	return append(data, '\n')
}

func fixtureReleaseManifestForMigration(t *testing.T, epoch []byte, migration int) []byte {
	t.Helper()
	defaultHash := "sha256:" + strings.Repeat("d", 64)
	artifactHashes := map[string]string{
		".env.example": defaultHash, "Caddyfile": defaultHash, "Dockerfile.gameserver": defaultHash,
		"LICENSE": defaultHash, "compose.yml": defaultHash, "compose.rotation.yml": defaultHash,
		"config.schema.json": defaultHash, "release-manifest.schema.json": defaultHash, "rehearsal-evidence.schema.json": defaultHash,
		"images/gameserver.docker.tar": defaultHash, "sbom/application.spdx.json": defaultHash,
		"third-party-licenses.txt": defaultHash, "site/index.html": defaultHash,
		"site/third-party-licenses.txt": defaultHash, "gameserver": defaultHash,
		"deployment-backup": defaultHash, "deployment-release": defaultHash, "deployment-operations": defaultHash, "deployment-rehearsal": defaultHash,
		"operations/prometheus.yml": defaultHash, "operations/cloud-clicker-alerts.yml": defaultHash,
		"operations/cloud-clicker-alerts.test.yml": defaultHash, "operations/alertmanager.example.yml": defaultHash,
		"operations/journald.template.conf": defaultHash, "operations/cloud-clicker-observe.service": defaultHash,
		"operations/cloud-clicker-observe.timer": defaultHash, "operations/operations.env.example": defaultHash,
		"content/balance/epochs/phase0.json": digest(epoch),
		"sbom/alertmanager.spdx.json":        "sha256:" + strings.Repeat("1", 64),
		"sbom/caddy.spdx.json":               "sha256:" + strings.Repeat("2", 64),
		"sbom/gameserver.spdx.json":          "sha256:" + strings.Repeat("3", 64),
		"sbom/node-exporter.spdx.json":       "sha256:" + strings.Repeat("4", 64),
		"sbom/postgres.spdx.json":            "sha256:" + strings.Repeat("5", 64),
		"sbom/prometheus.spdx.json":          "sha256:" + strings.Repeat("6", 64),
	}
	paths := make([]string, 0, len(artifactHashes))
	for path := range artifactHashes {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	artifacts := make([]releasepackage.File, 0, len(paths))
	for _, path := range paths {
		artifacts = append(artifacts, releasepackage.File{Path: path, SHA256: artifactHashes[path]})
	}
	gameConfig := "sha256:" + strings.Repeat("b", 64)
	manifest := releasepackage.ReleaseManifest{
		SchemaVersion: 1, ReleaseVersion: "0.1.0", SourceCommit: strings.Repeat("a", 40), Platform: "linux/amd64",
		DockerEngineVersion: "28.3.3", DockerComposeVersion: "2.39.1", DatabaseMigration: migration,
		CompanySaveVersion: save.LatestCompanyVersion, FounderSaveVersion: save.LatestFounderVersion,
		EpochID: 8, ConstantsHash: "sha256:" + strings.Repeat("c", 64), CopyHash: defaultHash,
		Images: []releasepackage.Image{
			{Name: "alertmanager", Reference: "alertmanager:v1@sha256:" + strings.Repeat("a", 64), RuntimeConfigSHA256: "sha256:" + strings.Repeat("1", 64), SBOMPath: "sbom/alertmanager.spdx.json", SBOMSHA256: artifactHashes["sbom/alertmanager.spdx.json"]},
			{Name: "caddy", Reference: "caddy:2@sha256:" + strings.Repeat("b", 64), RuntimeConfigSHA256: "sha256:" + strings.Repeat("2", 64), SBOMPath: "sbom/caddy.spdx.json", SBOMSHA256: artifactHashes["sbom/caddy.spdx.json"]},
			{Name: "gameserver", Reference: gameConfig, RuntimeConfigSHA256: gameConfig, SBOMPath: "sbom/gameserver.spdx.json", SBOMSHA256: artifactHashes["sbom/gameserver.spdx.json"]},
			{Name: "node-exporter", Reference: "node-exporter:v1@sha256:" + strings.Repeat("d", 64), RuntimeConfigSHA256: "sha256:" + strings.Repeat("4", 64), SBOMPath: "sbom/node-exporter.spdx.json", SBOMSHA256: artifactHashes["sbom/node-exporter.spdx.json"]},
			{Name: "postgres", Reference: "postgres:16@sha256:" + strings.Repeat("e", 64), RuntimeConfigSHA256: "sha256:" + strings.Repeat("5", 64), SBOMPath: "sbom/postgres.spdx.json", SBOMSHA256: artifactHashes["sbom/postgres.spdx.json"]},
			{Name: "prometheus", Reference: "prometheus:v1@sha256:" + strings.Repeat("f", 64), RuntimeConfigSHA256: "sha256:" + strings.Repeat("6", 64), SBOMPath: "sbom/prometheus.spdx.json", SBOMSHA256: artifactHashes["sbom/prometheus.spdx.json"]},
		}, Artifacts: artifacts,
	}
	if err := releasepackage.ValidateReleaseManifest(manifest); err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	return append(data, '\n')
}

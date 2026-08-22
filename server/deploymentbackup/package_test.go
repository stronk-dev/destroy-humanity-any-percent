package deploymentbackup

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"cloud-clicker/server/releasepackage"
	"cloud-clicker/server/save"
	"filippo.io/age"
)

func TestPackageRoundTripBindsDumpManifestEpochAndHeader(t *testing.T) {
	identity, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatal(err)
	}
	epoch := []byte("{\"schema_version\":1,\"current_epoch_id\":8}\n")
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
	epoch := []byte("{\"schema_version\":1,\"current_epoch_id\":8}\n")
	manifest := fixtureReleaseManifest(t, epoch)
	wrongEpoch := []byte("{\"schema_version\":1,\"current_epoch_id\":7}\n")
	input := PackageInput{Directory: t.TempDir(), BackupID: "20260822T145900Z-abcdef123456", ServerID: "server",
		StartedAt: time.Now().Add(-time.Minute), Now: time.Now, Recipient: identity.Recipient().String(),
		Dump: bytes.NewReader([]byte("PGDMP")), ReleaseManifest: manifest, EpochDeclaration: wrongEpoch}
	if _, _, err := CreatePackage(input); !errors.Is(err, ErrInvalid) {
		t.Fatalf("wrong epoch accepted: %v", err)
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
	epoch := []byte("{\"schema_version\":1,\"current_epoch_id\":8}\n")
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

func fixtureReleaseManifest(t *testing.T, epoch []byte) []byte {
	return fixtureReleaseManifestForMigration(t, epoch, 1)
}

func fixtureReleaseManifestForMigration(t *testing.T, epoch []byte, migration int) []byte {
	t.Helper()
	defaultHash := "sha256:" + strings.Repeat("d", 64)
	artifactHashes := map[string]string{
		".env.example": defaultHash, "Caddyfile": defaultHash, "Dockerfile.gameserver": defaultHash,
		"LICENSE": defaultHash, "compose.yml": defaultHash, "compose.rotation.yml": defaultHash,
		"config.schema.json": defaultHash, "release-manifest.schema.json": defaultHash,
		"images/gameserver.docker.tar": defaultHash, "sbom/application.spdx.json": defaultHash,
		"third-party-licenses.txt": defaultHash, "site/index.html": defaultHash,
		"site/third-party-licenses.txt": defaultHash, "gameserver": defaultHash,
		"deployment-backup":                  defaultHash,
		"content/balance/epochs/phase0.json": digest(epoch),
		"sbom/caddy.spdx.json":               "sha256:" + strings.Repeat("1", 64),
		"sbom/gameserver.spdx.json":          "sha256:" + strings.Repeat("2", 64),
		"sbom/postgres.spdx.json":            "sha256:" + strings.Repeat("3", 64),
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
			{Name: "caddy", Reference: "caddy:2@sha256:" + strings.Repeat("a", 64), RuntimeConfigSHA256: "sha256:" + strings.Repeat("1", 64), SBOMPath: "sbom/caddy.spdx.json", SBOMSHA256: artifactHashes["sbom/caddy.spdx.json"]},
			{Name: "gameserver", Reference: gameConfig, RuntimeConfigSHA256: gameConfig, SBOMPath: "sbom/gameserver.spdx.json", SBOMSHA256: artifactHashes["sbom/gameserver.spdx.json"]},
			{Name: "postgres", Reference: "postgres:16@sha256:" + strings.Repeat("c", 64), RuntimeConfigSHA256: "sha256:" + strings.Repeat("3", 64), SBOMPath: "sbom/postgres.spdx.json", SBOMSHA256: artifactHashes["sbom/postgres.spdx.json"]},
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

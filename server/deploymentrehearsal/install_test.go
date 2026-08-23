package deploymentrehearsal

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"cloud-clicker/server/deploymentrelease"
	"cloud-clicker/server/releasepackage"
)

func TestCandidateInstallRetainsOnlyExactControllerAuthority(t *testing.T) {
	config, manifestBytes := candidateInstallFixture(t)
	execute := func(_ context.Context, value ScenarioConfig) error {
		record := authorityInstallRecord(time.Date(2026, 8, 23, 12, 0, 0, 0, time.UTC), "1.0.0",
			hashBytes(manifestBytes), authorityImageReferences("a"))
		return os.WriteFile(filepath.Join(value.InstallOperatorState, "release-ledger.jsonl"), encodeReleaseRecords([]deploymentrelease.ReleaseRecord{record}), 0o600)
	}
	dependencies := candidateInstallDependencies{validateBundle: func(string) error { return nil }, execute: execute}
	if err := installCandidate(context.Background(), config, dependencies); err != nil {
		t.Fatal(err)
	}
	retained := filepath.Join(config.ArtifactsDirectory, requiredRunArtifactFiles["install_ledger"])
	records, err := deploymentrelease.DecodeReleaseLedger(mustRead(t, retained))
	if err != nil || len(records) != 1 || records[0].ManifestSHA256 != hashBytes(manifestBytes) {
		t.Fatalf("records=%+v err=%v", records, err)
	}
	if err := installCandidate(context.Background(), config, dependencies); err == nil {
		t.Fatal("retained install authority overwrite accepted")
	}
}

func TestCandidateInstallRejectsControllerAndLedgerFailures(t *testing.T) {
	for name, execute := range map[string]func(context.Context, ScenarioConfig) error{
		"controller failure": func(context.Context, ScenarioConfig) error { return errors.New("install failed") },
		"missing ledger":     func(context.Context, ScenarioConfig) error { return nil },
		"wrong manifest": func(_ context.Context, value ScenarioConfig) error {
			record := authorityInstallRecord(time.Now(), "1.0.0", hashForBuild("f"), authorityImageReferences("a"))
			return os.WriteFile(filepath.Join(value.InstallOperatorState, "release-ledger.jsonl"), encodeReleaseRecords([]deploymentrelease.ReleaseRecord{record}), 0o600)
		},
		"extra success": func(_ context.Context, value ScenarioConfig) error {
			manifestBytes := mustRead(t, filepath.Join(value.CandidateBundle, releasepackage.ReleaseManifestPath))
			first := authorityInstallRecord(time.Now(), "1.0.0", hashBytes(manifestBytes), authorityImageReferences("a"))
			second := first
			second.StartedAt = first.CompletedAt.Add(time.Second)
			second.CompletedAt = second.StartedAt.Add(time.Second)
			return os.WriteFile(filepath.Join(value.InstallOperatorState, "release-ledger.jsonl"),
				encodeReleaseRecords([]deploymentrelease.ReleaseRecord{first, second}), 0o600)
		},
	} {
		t.Run(name, func(t *testing.T) {
			config, _ := candidateInstallFixture(t)
			dependencies := candidateInstallDependencies{validateBundle: func(string) error { return nil }, execute: execute}
			if err := installCandidate(context.Background(), config, dependencies); err == nil {
				t.Fatal("invalid candidate install authority accepted")
			}
			if _, err := os.Lstat(filepath.Join(config.ArtifactsDirectory, requiredRunArtifactFiles["install_ledger"])); !errors.Is(err, os.ErrNotExist) {
				t.Fatalf("failed install retained authority: %v", err)
			}
		})
	}
}

func candidateInstallFixture(t *testing.T) (ScenarioConfig, []byte) {
	t.Helper()
	config := validScenarioConfig(t)
	manifestBytes := authorityManifestFixture(strings.Repeat("c", 40), "1.0.0", "a")
	if err := os.WriteFile(filepath.Join(config.CandidateBundle, releasepackage.ReleaseManifestPath), manifestBytes, 0o444); err != nil {
		t.Fatal(err)
	}
	return config, manifestBytes
}

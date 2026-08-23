package deploymentrehearsal

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"cloud-clicker/server/releasepackage"
)

func TestObserveSupplyChainDerivesExactValidatedClosure(t *testing.T) {
	fixture := supplyChainFixture(t)
	result, err := ObserveSupplyChain(fixture.request)
	if err != nil || result.ProductionImages != 6 || result.SBOMDocuments != 8 {
		t.Fatalf("result=%+v err=%v", result, err)
	}
	data, err := os.ReadFile(fixture.request.Output)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := DecodeSupplyChainResult(data); err != nil {
		t.Fatal(err)
	}
	if _, err := ObserveSupplyChain(fixture.request); err == nil {
		t.Fatal("supply-chain result overwritten")
	}
}

func TestObserveSupplyChainRejectsUnvalidatedOrMismatchedInputs(t *testing.T) {
	t.Run("bundle validator", func(t *testing.T) {
		fixture := supplyChainFixture(t)
		validateReleaseBundle = func(string) error { return releasepackage.ErrInvalidContent }
		t.Cleanup(func() { validateReleaseBundle = releasepackage.ValidateBundle })
		if _, err := ObserveSupplyChain(fixture.request); !errors.Is(err, ErrInvalid) {
			t.Fatalf("unvalidated bundle accepted: %v", err)
		}
	})
	t.Run("secret identity", func(t *testing.T) {
		fixture := supplyChainFixture(t)
		secret := validSecretScanResultForSupply(fixture.candidateManifest, strings.Repeat("e", 40), fixture.started)
		if err := releasepackage.WriteSecretScanResult(fixture.request.SecretScan+"-wrong", secret); err != nil {
			t.Fatal(err)
		}
		fixture.request.SecretScan += "-wrong"
		if _, err := ObserveSupplyChain(fixture.request); !errors.Is(err, ErrInvalid) {
			t.Fatalf("wrong source identity accepted: %v", err)
		}
	})
}

type supplyFixture struct {
	request           SupplyChainRequest
	candidateManifest []byte
	started           time.Time
}

func supplyChainFixture(t *testing.T) supplyFixture {
	t.Helper()
	root := t.TempDir()
	candidateBundle := filepath.Join(root, "candidate")
	previousBundle := filepath.Join(root, "previous")
	if err := os.Mkdir(candidateBundle, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(previousBundle, 0o700); err != nil {
		t.Fatal(err)
	}
	candidateManifest := supplyManifestBytes(t, strings.Repeat("c", 40), true)
	previousManifest := supplyManifestBytes(t, strings.Repeat("d", 40), false)
	if err := os.WriteFile(filepath.Join(candidateBundle, releasepackage.ReleaseManifestPath), candidateManifest, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(previousBundle, releasepackage.ReleaseManifestPath), previousManifest, 0o600); err != nil {
		t.Fatal(err)
	}
	candidateBuild := validBuildRecord()
	candidateBuild.SchemaVersion = 2
	candidateBuild.Role = "candidate"
	candidateBuild.SourceCommit = strings.Repeat("c", 40)
	candidateBuild.ManifestSHA256 = hashBytes(candidateManifest)
	candidateBuild.RebuildManifestSHA256 = candidateBuild.ManifestSHA256
	candidateBuild.RehearsalImages = []BuildImage{validRehearsalImage()}
	previousBuild := validBuildRecord()
	previousBuild.SourceCommit = strings.Repeat("d", 40)
	previousBuild.ManifestSHA256 = hashBytes(previousManifest)
	previousBuild.RebuildManifestSHA256 = previousBuild.ManifestSHA256
	candidateBuildPath := filepath.Join(root, "candidate-build.json")
	previousBuildPath := filepath.Join(root, "previous-build.json")
	writeJSONFile(t, candidateBuildPath, candidateBuild)
	writeJSONFile(t, previousBuildPath, previousBuild)
	started := time.Date(2026, 8, 23, 12, 0, 0, 0, time.UTC)
	secretPath := filepath.Join(root, "secret.json")
	if err := releasepackage.WriteSecretScanResult(secretPath, validSecretScanResultForSupply(candidateManifest, strings.Repeat("c", 40), started)); err != nil {
		t.Fatal(err)
	}
	original := validateReleaseBundle
	validateReleaseBundle = func(string) error { return nil }
	t.Cleanup(func() { validateReleaseBundle = original })
	clock := started.Add(2 * time.Second)
	now := func() time.Time {
		clock = clock.Add(time.Second)
		return clock
	}
	return supplyFixture{candidateManifest: candidateManifest, started: started,
		request: SupplyChainRequest{CandidateBundle: candidateBundle, PreviousBundle: previousBundle,
			CandidateBuild: candidateBuildPath, PreviousBuild: previousBuildPath, SecretScan: secretPath,
			Output: filepath.Join(root, "supply-chain.json"), Now: now}}
}

func supplyManifestBytes(t *testing.T, commit string, rehearsal bool) []byte {
	t.Helper()
	manifest := map[string]any{"source_commit": commit, "images": []any{1, 2, 3, 4, 5, 6},
		"artifacts": []map[string]string{{"path": "LICENSE"}, {"path": "third-party-licenses.txt"}, {"path": "site/third-party-licenses.txt"}}}
	if rehearsal {
		manifest["rehearsal_images"] = []any{1}
	} else {
		manifest["rehearsal_images"] = []any{}
	}
	data, err := json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	return append(data, '\n')
}

func validSecretScanResultForSupply(manifest []byte, commit string, started time.Time) releasepackage.SecretScanResult {
	return releasepackage.SecretScanResult{SchemaVersion: 1, ManifestSHA256: hashBytes(manifest), SourceCommit: commit,
		StartedAt: started, CompletedAt: started.Add(time.Second), TrackedFiles: 100, ImageArchiveScanned: true, ObjectiveCompleted: true}
}

func writeJSONFile(t *testing.T, path string, value any) {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, append(data, '\n'), 0o600); err != nil {
		t.Fatal(err)
	}
}

package harness

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"cloud-clicker/server/epochseed"
	"cloud-clicker/server/save"
)

func TestContentDynamicsProductionRegistryIsHonestlyEmpty(t *testing.T) {
	entries, err := LoadContentDynamicsRegistry("../..")
	if err != nil || len(entries) != 0 {
		t.Fatalf("entries=%+v err=%v", entries, err)
	}
	if err := GenerateRegisteredContentSnapshots("../.."); err != nil {
		t.Fatal(err)
	}
}

func TestContentDynamicsGoldenPathDiscoveryIsStrict(t *testing.T) {
	empty, err := registeredContentGoldenPaths([]byte(`{"schema_version":1,"entries":[]}`))
	if err != nil || len(empty) != 0 {
		t.Fatalf("empty=%v err=%v", empty, err)
	}
	data := []byte(`{"schema_version":1,"entries":[{"epoch_seed_path":"balance/epochs/phase0.json","epoch_id":7,"bundle_snapshot_manifest":"testdata/harness/content-dynamics/bundles/sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa/manifest.v1.json","scenario":"testdata/harness/content-dynamics/scenarios/epoch-7.json","golden_report":"testdata/harness/content-dynamics/goldens/epoch-7.json"}]}`)
	paths, err := registeredContentGoldenPaths(data)
	if err != nil || len(paths) != 1 || paths[0] != "testdata/harness/content-dynamics/goldens/epoch-7.json" {
		t.Fatalf("paths=%v err=%v", paths, err)
	}
	if _, err := registeredContentGoldenPaths([]byte(strings.Replace(string(data), "goldens/epoch-7.json", "../escape.json", 1))); err == nil {
		t.Fatal("escaping golden path accepted")
	}
}

func TestContentDynamicsSnapshotRoundTripAndTamperRejection(t *testing.T) {
	for _, mutation := range []string{"none", "missing", "extra", "tampered", "manifest_hash"} {
		t.Run(mutation, func(t *testing.T) {
			root, bundle, entry := contentSnapshotFixture(t)
			if err := GenerateContentBundleSnapshot(root, bundle, 1, entry.BundleSnapshotManifest); err != nil {
				t.Fatal(err)
			}
			artifactDirectory := filepath.Join(root, filepath.FromSlash(filepath.Dir(entry.BundleSnapshotManifest)), "artifacts")
			switch mutation {
			case "missing":
				if err := os.Remove(filepath.Join(artifactDirectory, "economy.json")); err != nil {
					t.Fatal(err)
				}
			case "extra":
				if err := os.WriteFile(filepath.Join(artifactDirectory, "extra.json"), []byte("{}"), 0o644); err != nil {
					t.Fatal(err)
				}
			case "tampered":
				if err := os.WriteFile(filepath.Join(artifactDirectory, "routes.json"), []byte(`{"tampered":true}`), 0o644); err != nil {
					t.Fatal(err)
				}
			case "manifest_hash":
				path := filepath.Join(root, filepath.FromSlash(entry.BundleSnapshotManifest))
				data, err := os.ReadFile(path)
				if err != nil {
					t.Fatal(err)
				}
				data = []byte(strings.Replace(string(data), bundle.Hash, "sha256:"+strings.Repeat("f", 64), 1))
				if err := os.WriteFile(path, data, 0o644); err != nil {
					t.Fatal(err)
				}
			}
			loaded, err := LoadContentBundleSnapshot(root, entry)
			if mutation == "none" {
				if err != nil || loaded.Hash != bundle.Hash || len(loaded.Artifacts) != 2 {
					t.Fatalf("loaded=%+v err=%v", loaded, err)
				}
				if err := GenerateContentBundleSnapshot(root, bundle, 1, entry.BundleSnapshotManifest); err != nil {
					t.Fatalf("byte-identical regeneration failed: %v", err)
				}
			} else if err == nil {
				t.Fatalf("mutation %s was accepted", mutation)
			}
		})
	}
}

func TestContentDynamicsSnapshotRefusesRewriteAndHandAuthoredSubset(t *testing.T) {
	root, bundle, entry := contentSnapshotFixture(t)
	if err := GenerateContentBundleSnapshot(root, bundle, 1, entry.BundleSnapshotManifest); err != nil {
		t.Fatal(err)
	}
	changed := bundle
	changed.Artifacts = map[string][]byte{"economy": []byte(`{"changed":true}`), "routes": bundle.Artifacts["routes"]}
	changed.Hash, _ = save.ConstantsHashArtifacts(changed.Artifacts)
	changed.Seed.Epochs[0].AcceptedHashes = []string{changed.Hash}
	if err := GenerateContentBundleSnapshot(root, changed, 1, entry.BundleSnapshotManifest); err == nil || !strings.Contains(err.Error(), "refusing to rewrite") {
		t.Fatalf("rewrite err=%v", err)
	}
	subset := bundle
	subset.Artifacts = map[string][]byte{"economy": bundle.Artifacts["economy"]}
	subset.Hash, _ = save.ConstantsHashArtifacts(subset.Artifacts)
	subset.Seed.Epochs[0].AcceptedHashes = []string{subset.Hash}
	if err := GenerateContentBundleSnapshot(root, subset, 1, "testdata/harness/content-dynamics/bundles/subset/manifest.v1.json"); err == nil {
		t.Fatal("hand-authored artifact subset accepted")
	}
}

func contentSnapshotFixture(t *testing.T) (string, epochseed.Bundle, ContentDynamicsRegistryEntry) {
	t.Helper()
	root := t.TempDir()
	artifacts := map[string][]byte{"economy": []byte(`{"schema_version":1}`), "routes": []byte(`{"schema_version":1}`)}
	hash, err := save.ConstantsHashArtifacts(artifacts)
	if err != nil {
		t.Fatal(err)
	}
	seed := epochseed.Seed{SchemaVersion: 1, CurrentEpochID: 1,
		Artifacts: []epochseed.Artifact{{Name: "economy", Path: "balance/catalogs/economy.json"}, {Name: "routes", Path: "balance/routes/routes.json"}},
		Epochs:    []epochseed.Epoch{{ID: 1, Name: "fixture", ChangelogRef: "changelog/epoch-1.md", AcceptedHashes: []string{hash}}}}
	seedBytes, err := CanonicalJSON(seed)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "balance", "epochs"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, filepath.FromSlash(epochseed.Path)), seedBytes, 0o644); err != nil {
		t.Fatal(err)
	}
	entry := ContentDynamicsRegistryEntry{EpochSeedPath: epochseed.Path, EpochID: 1,
		BundleSnapshotManifest: "testdata/harness/content-dynamics/bundles/" + hash + "/manifest.v1.json",
		Scenario:               "testdata/harness/content-dynamics/scenarios/fixture.json", GoldenReport: "testdata/harness/content-dynamics/goldens/fixture.json"}
	return root, epochseed.Bundle{Seed: seed, Artifacts: artifacts, Hash: hash}, entry
}

func TestContentDynamicsHashFixtureUsesFullSHA256Coordinate(t *testing.T) {
	digest := sha256.Sum256([]byte("fixture"))
	value := "sha256:" + hex.EncodeToString(digest[:])
	if !relevanceHashPattern.MatchString(value) {
		t.Fatalf("invalid full hash %s", value)
	}
}

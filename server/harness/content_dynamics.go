package harness

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"cloud-clicker/server/epochseed"
	"cloud-clicker/server/save"
)

const contentDynamicsRegistryPath = "testdata/harness/content-dynamics/registry-v1.json"

type ContentDynamicsRegistryEntry struct {
	EpochSeedPath          string `json:"epoch_seed_path"`
	EpochID                int64  `json:"epoch_id"`
	BundleSnapshotManifest string `json:"bundle_snapshot_manifest"`
	Scenario               string `json:"scenario"`
	GoldenReport           string `json:"golden_report"`
}

type contentDynamicsSnapshotArtifact struct {
	Name           string `json:"name"`
	ProductionPath string `json:"production_path"`
	SnapshotPath   string `json:"snapshot_path"`
	SHA256         string `json:"sha256"`
}

type contentDynamicsSnapshotManifest struct {
	SchemaVersion int                               `json:"schema_version"`
	EpochSeedPath string                            `json:"epoch_seed_path"`
	EpochID       int64                             `json:"epoch_id"`
	ConstantsHash string                            `json:"constants_hash"`
	Artifacts     []contentDynamicsSnapshotArtifact `json:"artifacts"`
}

type contentDynamicsRegistryWireEntry struct {
	EpochSeedPath          *string `json:"epoch_seed_path"`
	EpochID                *int64  `json:"epoch_id"`
	BundleSnapshotManifest *string `json:"bundle_snapshot_manifest"`
	Scenario               *string `json:"scenario"`
	GoldenReport           *string `json:"golden_report"`
}

func LoadContentDynamicsRegistry(repositoryRoot string) ([]ContentDynamicsRegistryEntry, error) {
	data, err := os.ReadFile(filepath.Join(repositoryRoot, filepath.FromSlash(contentDynamicsRegistryPath)))
	if err != nil {
		return nil, err
	}
	return decodeContentDynamicsRegistry(data)
}

func decodeContentDynamicsRegistry(data []byte) ([]ContentDynamicsRegistryEntry, error) {
	var wire struct {
		SchemaVersion *int                               `json:"schema_version"`
		Entries       []contentDynamicsRegistryWireEntry `json:"entries"`
	}
	if err := decodeRelevanceStrict(data, &wire); err != nil || wire.SchemaVersion == nil || *wire.SchemaVersion != 1 || wire.Entries == nil {
		return nil, errors.New("invalid content-dynamics registry")
	}
	entries := make([]ContentDynamicsRegistryEntry, 0, len(wire.Entries))
	owned, prior := map[string]bool{}, ""
	for index, row := range wire.Entries {
		if row.EpochSeedPath == nil || row.EpochID == nil || row.BundleSnapshotManifest == nil || row.Scenario == nil || row.GoldenReport == nil ||
			*row.EpochID < 1 || *row.Scenario == "" || prior != "" && prior >= *row.Scenario {
			return nil, fmt.Errorf("invalid content-dynamics registry entry %d", index)
		}
		entry := ContentDynamicsRegistryEntry{EpochSeedPath: *row.EpochSeedPath, EpochID: *row.EpochID,
			BundleSnapshotManifest: *row.BundleSnapshotManifest, Scenario: *row.Scenario, GoldenReport: *row.GoldenReport}
		if !validRepositoryPath(entry.EpochSeedPath) {
			return nil, fmt.Errorf("invalid content-dynamics epoch seed path %q", entry.EpochSeedPath)
		}
		// Every historical entry reads the same append-only epoch authority. Only
		// the generated snapshot, scenario, and report are entry-owned paths.
		for _, path := range []string{entry.BundleSnapshotManifest, entry.Scenario, entry.GoldenReport} {
			if !validRepositoryPath(path) || owned[path] {
				return nil, fmt.Errorf("invalid or duplicate content-dynamics path %q", path)
			}
			owned[path] = true
		}
		if entry.EpochSeedPath != epochseed.Path || !strings.HasPrefix(entry.BundleSnapshotManifest, "testdata/harness/content-dynamics/bundles/") ||
			!strings.HasPrefix(entry.Scenario, "testdata/harness/content-dynamics/scenarios/") ||
			!strings.HasPrefix(entry.GoldenReport, "testdata/harness/content-dynamics/goldens/") {
			return nil, fmt.Errorf("content-dynamics entry %q owns an invalid path class", entry.Scenario)
		}
		entries, prior = append(entries, entry), entry.Scenario
	}
	return entries, nil
}

func GenerateRegisteredContentSnapshots(repositoryRoot string) error {
	entries, err := LoadContentDynamicsRegistry(repositoryRoot)
	if err != nil {
		return err
	}
	if len(entries) == 0 {
		return nil
	}
	bundle, err := epochseed.Load(repositoryRoot)
	if err != nil {
		return err
	}
	current := epochseed.Current(bundle.Seed)
	for _, entry := range entries {
		manifestPath := filepath.Join(repositoryRoot, filepath.FromSlash(entry.BundleSnapshotManifest))
		if _, statErr := os.Stat(manifestPath); errors.Is(statErr, os.ErrNotExist) {
			if entry.EpochID != current.ID {
				return fmt.Errorf("historical content-dynamics snapshot %d is missing", entry.EpochID)
			}
			if err := GenerateContentBundleSnapshot(repositoryRoot, bundle, entry.EpochID, entry.BundleSnapshotManifest); err != nil {
				return err
			}
		} else if statErr != nil {
			return statErr
		}
		if _, err := LoadContentBundleSnapshot(repositoryRoot, entry); err != nil {
			return err
		}
	}
	return nil
}

func GenerateRegisteredContentBaselines(repositoryRoot string) error {
	if err := GenerateRegisteredContentSnapshots(repositoryRoot); err != nil {
		return err
	}
	entries, err := LoadContentDynamicsRegistry(repositoryRoot)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		suite, loadErr := LoadRegisteredContentDynamicsSuite(repositoryRoot, entry)
		if loadErr != nil {
			return loadErr
		}
		report, runErr := suite.Run()
		if runErr != nil {
			return runErr
		}
		data, encodeErr := CanonicalJSON(report)
		if encodeErr != nil {
			return encodeErr
		}
		if err := writeImmutableFile(filepath.Join(repositoryRoot, filepath.FromSlash(entry.GoldenReport)), data); err != nil {
			return err
		}
	}
	return nil
}

func ValidateContentDynamicsRegistry(repositoryRoot string) error {
	entries, err := LoadContentDynamicsRegistry(repositoryRoot)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		suite, err := LoadRegisteredContentDynamicsSuite(repositoryRoot, entry)
		if err != nil {
			return err
		}
		report, err := suite.Run()
		if err != nil {
			return err
		}
		actual, err := CanonicalJSON(report)
		if err != nil {
			return err
		}
		golden, err := os.ReadFile(filepath.Join(repositoryRoot, filepath.FromSlash(entry.GoldenReport)))
		if err != nil {
			return err
		}
		if !bytes.Equal(actual, golden) {
			return fmt.Errorf("content-dynamics golden drift for %q", entry.Scenario)
		}
	}
	return nil
}

func GenerateContentBundleSnapshot(repositoryRoot string, bundle epochseed.Bundle, epochID int64, manifestRelativePath string) error {
	if repositoryRoot == "" || !validRepositoryPath(manifestRelativePath) || epochseed.Validate(bundle.Seed) != nil ||
		epochID != epochseed.Current(bundle.Seed).ID || !epochseed.Accepts(epochseed.Current(bundle.Seed), bundle.Hash) {
		return errors.New("invalid content-dynamics snapshot source")
	}
	recomputed, err := save.ConstantsHashArtifacts(bundle.Artifacts)
	if err != nil || recomputed != bundle.Hash || len(bundle.Artifacts) != len(bundle.Seed.Artifacts) {
		return errors.New("content-dynamics snapshot source identity mismatch")
	}
	base := filepath.ToSlash(filepath.Dir(manifestRelativePath))
	manifest := contentDynamicsSnapshotManifest{SchemaVersion: 1, EpochSeedPath: epochseed.Path, EpochID: epochID,
		ConstantsHash: bundle.Hash, Artifacts: make([]contentDynamicsSnapshotArtifact, 0, len(bundle.Seed.Artifacts))}
	sources := append([]epochseed.Artifact(nil), bundle.Seed.Artifacts...)
	sort.Slice(sources, func(left, right int) bool { return sources[left].Name < sources[right].Name })
	for _, source := range sources {
		data, ok := bundle.Artifacts[source.Name]
		if !ok || len(data) == 0 {
			return fmt.Errorf("content-dynamics snapshot missing artifact %q", source.Name)
		}
		digest := sha256.Sum256(data)
		snapshotPath := base + "/artifacts/" + source.Name + filepath.Ext(source.Path)
		manifest.Artifacts = append(manifest.Artifacts, contentDynamicsSnapshotArtifact{Name: source.Name,
			ProductionPath: source.Path, SnapshotPath: snapshotPath, SHA256: "sha256:" + hex.EncodeToString(digest[:])})
		if err := writeImmutableFile(filepath.Join(repositoryRoot, filepath.FromSlash(snapshotPath)), data); err != nil {
			return err
		}
	}
	manifestBytes, err := CanonicalJSON(manifest)
	if err != nil {
		return err
	}
	return writeImmutableFile(filepath.Join(repositoryRoot, filepath.FromSlash(manifestRelativePath)), manifestBytes)
}

func LoadContentBundleSnapshot(repositoryRoot string, entry ContentDynamicsRegistryEntry) (epochseed.Bundle, error) {
	manifestBytes, err := os.ReadFile(filepath.Join(repositoryRoot, filepath.FromSlash(entry.BundleSnapshotManifest)))
	if err != nil {
		return epochseed.Bundle{}, err
	}
	manifest, err := decodeContentDynamicsSnapshotManifest(manifestBytes)
	if err != nil || manifest.EpochSeedPath != entry.EpochSeedPath || manifest.EpochID != entry.EpochID {
		return epochseed.Bundle{}, errors.New("invalid content-dynamics snapshot manifest")
	}
	seedBytes, err := os.ReadFile(filepath.Join(repositoryRoot, filepath.FromSlash(entry.EpochSeedPath)))
	if err != nil {
		return epochseed.Bundle{}, err
	}
	seed, err := epochseed.Decode(seedBytes)
	if err != nil || entry.EpochID > int64(len(seed.Epochs)) || !epochseed.Accepts(seed.Epochs[entry.EpochID-1], manifest.ConstantsHash) {
		return epochseed.Bundle{}, errors.New("content-dynamics snapshot is not accepted by its epoch")
	}
	artifacts := make(map[string][]byte, len(manifest.Artifacts))
	declaredFiles := make(map[string]bool, len(manifest.Artifacts))
	productionPaths := make(map[string]bool, len(manifest.Artifacts))
	snapshotPaths := make(map[string]bool, len(manifest.Artifacts))
	prior := ""
	for _, row := range manifest.Artifacts {
		if row.Name == "" || prior != "" && prior >= row.Name || !validRepositoryPath(row.ProductionPath) || !validRepositoryPath(row.SnapshotPath) ||
			!strings.HasPrefix(row.SnapshotPath, filepath.ToSlash(filepath.Dir(entry.BundleSnapshotManifest))+"/artifacts/") || !relevanceHashPattern.MatchString(row.SHA256) ||
			productionPaths[row.ProductionPath] || snapshotPaths[row.SnapshotPath] {
			return epochseed.Bundle{}, fmt.Errorf("invalid content-dynamics snapshot artifact %q", row.Name)
		}
		productionPaths[row.ProductionPath], snapshotPaths[row.SnapshotPath] = true, true
		data, readErr := os.ReadFile(filepath.Join(repositoryRoot, filepath.FromSlash(row.SnapshotPath)))
		if readErr != nil {
			return epochseed.Bundle{}, readErr
		}
		digest := sha256.Sum256(data)
		if row.SHA256 != "sha256:"+hex.EncodeToString(digest[:]) {
			return epochseed.Bundle{}, fmt.Errorf("content-dynamics snapshot artifact %q hash mismatch", row.Name)
		}
		artifacts[row.Name], declaredFiles[filepath.Base(row.SnapshotPath)], prior = data, true, row.Name
	}
	artifactDirectory := filepath.Join(repositoryRoot, filepath.FromSlash(filepath.ToSlash(filepath.Dir(entry.BundleSnapshotManifest))+"/artifacts"))
	files, err := os.ReadDir(artifactDirectory)
	if err != nil || len(files) != len(declaredFiles) {
		return epochseed.Bundle{}, errors.New("content-dynamics snapshot artifact set mismatch")
	}
	for _, file := range files {
		if file.IsDir() || !declaredFiles[file.Name()] {
			return epochseed.Bundle{}, errors.New("content-dynamics snapshot contains an undeclared artifact")
		}
	}
	recomputed, err := save.ConstantsHashArtifacts(artifacts)
	if err != nil || recomputed != manifest.ConstantsHash {
		return epochseed.Bundle{}, errors.New("content-dynamics snapshot bundle hash mismatch")
	}
	snapshotSeed := seed
	snapshotSeed.CurrentEpochID = entry.EpochID
	snapshotSeed.Epochs = append([]epochseed.Epoch(nil), seed.Epochs[:entry.EpochID]...)
	snapshotSeed.Artifacts = make([]epochseed.Artifact, 0, len(manifest.Artifacts))
	for _, row := range manifest.Artifacts {
		snapshotSeed.Artifacts = append(snapshotSeed.Artifacts, epochseed.Artifact{Name: row.Name, Path: row.ProductionPath})
	}
	return epochseed.Bundle{Seed: snapshotSeed, Artifacts: artifacts, Hash: manifest.ConstantsHash}, nil
}

func decodeContentDynamicsSnapshotManifest(data []byte) (contentDynamicsSnapshotManifest, error) {
	var manifest contentDynamicsSnapshotManifest
	if err := decodeRelevanceStrict(data, &manifest); err != nil || manifest.SchemaVersion != 1 ||
		!validRepositoryPath(manifest.EpochSeedPath) || !relevanceHashPattern.MatchString(manifest.ConstantsHash) || len(manifest.Artifacts) == 0 {
		return contentDynamicsSnapshotManifest{}, errors.New("invalid content-dynamics snapshot manifest")
	}
	prior, productionPaths, snapshotPaths := "", map[string]bool{}, map[string]bool{}
	for _, row := range manifest.Artifacts {
		if row.Name == "" || prior != "" && prior >= row.Name || !validRepositoryPath(row.ProductionPath) || !validRepositoryPath(row.SnapshotPath) ||
			!relevanceHashPattern.MatchString(row.SHA256) || productionPaths[row.ProductionPath] || snapshotPaths[row.SnapshotPath] {
			return contentDynamicsSnapshotManifest{}, errors.New("invalid content-dynamics snapshot artifact set")
		}
		prior, productionPaths[row.ProductionPath], snapshotPaths[row.SnapshotPath] = row.Name, true, true
	}
	return manifest, nil
}

func validRepositoryPath(value string) bool {
	return value != "" && filepath.ToSlash(filepath.Clean(value)) == value && value != "." && !strings.HasPrefix(value, "../") && !filepath.IsAbs(value)
}

func writeImmutableFile(path string, data []byte) error {
	if existing, err := os.ReadFile(path); err == nil {
		if bytes.Equal(existing, data) {
			return nil
		}
		return fmt.Errorf("refusing to rewrite immutable content-dynamics snapshot %q", path)
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

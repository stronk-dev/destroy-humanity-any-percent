// Package epochseed owns the repository manifest that defines which catalog
// artifacts compose a constants hash. Runtime, harness, and deployment code
// must consume this package instead of spelling artifact sets independently.
package epochseed

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"cloud-clicker/server/save"
)

const Path = "balance/epochs/phase0.json"

var (
	hashPattern         = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)
	artifactNamePattern = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)
	ErrInvalidSeed      = errors.New("invalid epoch seed")
)

type Artifact struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

type Epoch struct {
	ID             int64    `json:"epoch_id"`
	Name           string   `json:"name"`
	ChangelogRef   string   `json:"changelog_ref"`
	AcceptedHashes []string `json:"accepted_hashes"`
}

type Seed struct {
	SchemaVersion  int        `json:"schema_version"`
	CurrentEpochID int64      `json:"current_epoch_id"`
	Artifacts      []Artifact `json:"artifacts"`
	Epochs         []Epoch    `json:"epochs"`
}

type Bundle struct {
	Seed      Seed
	Artifacts map[string][]byte
	Hash      string
}

func Decode(data []byte) (Seed, error) {
	var seed Seed
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&seed); err != nil {
		return Seed{}, errors.Join(ErrInvalidSeed, err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return Seed{}, fmt.Errorf("%w: seed must contain exactly one JSON value", ErrInvalidSeed)
	}
	if err := Validate(seed); err != nil {
		return Seed{}, err
	}
	return seed, nil
}

func Validate(seed Seed) error {
	if seed.SchemaVersion != 1 || seed.CurrentEpochID < 1 || len(seed.Artifacts) == 0 || len(seed.Epochs) == 0 {
		return fmt.Errorf("%w: invalid root", ErrInvalidSeed)
	}
	seenNames, seenPaths := map[string]bool{}, map[string]bool{}
	for _, artifact := range seed.Artifacts {
		if !artifactNamePattern.MatchString(artifact.Name) || path.Clean(artifact.Path) != artifact.Path ||
			seenNames[artifact.Name] || seenPaths[artifact.Path] || !strings.HasPrefix(artifact.Path, "balance/") {
			return fmt.Errorf("%w: invalid or duplicate artifact %q", ErrInvalidSeed, artifact.Name)
		}
		seenNames[artifact.Name], seenPaths[artifact.Path] = true, true
	}
	for index, epoch := range seed.Epochs {
		if epoch.ID != int64(index+1) || strings.TrimSpace(epoch.Name) == "" ||
			epoch.ChangelogRef != fmt.Sprintf("changelog/epoch-%d.md", epoch.ID) || !SortedUniqueHashes(epoch.AcceptedHashes) {
			return fmt.Errorf("%w: invalid epoch %d", ErrInvalidSeed, epoch.ID)
		}
	}
	if seed.CurrentEpochID != seed.Epochs[len(seed.Epochs)-1].ID {
		return fmt.Errorf("%w: current epoch is not last", ErrInvalidSeed)
	}
	return nil
}

func Load(repositoryRoot string) (Bundle, error) {
	if repositoryRoot == "" {
		return Bundle{}, ErrInvalidSeed
	}
	data, err := os.ReadFile(filepath.Join(repositoryRoot, filepath.FromSlash(Path)))
	if err != nil {
		return Bundle{}, err
	}
	seed, err := Decode(data)
	if err != nil {
		return Bundle{}, err
	}
	artifacts, err := ReadArtifacts(repositoryRoot, seed)
	if err != nil {
		return Bundle{}, err
	}
	hash, err := save.ConstantsHashArtifacts(artifacts)
	if err != nil {
		return Bundle{}, err
	}
	return Bundle{Seed: seed, Artifacts: artifacts, Hash: hash}, nil
}

func ReadArtifacts(repositoryRoot string, seed Seed) (map[string][]byte, error) {
	if repositoryRoot == "" || Validate(seed) != nil {
		return nil, ErrInvalidSeed
	}
	artifacts := make(map[string][]byte, len(seed.Artifacts))
	for _, artifact := range seed.Artifacts {
		data, err := os.ReadFile(filepath.Join(repositoryRoot, filepath.FromSlash(artifact.Path)))
		if err != nil {
			return nil, fmt.Errorf("read artifact %s: %w", artifact.Name, err)
		}
		artifacts[artifact.Name] = data
	}
	return artifacts, nil
}

func ArtifactPath(seed Seed, name string) (string, bool) {
	for _, artifact := range seed.Artifacts {
		if artifact.Name == name {
			return artifact.Path, true
		}
	}
	return "", false
}

func Current(seed Seed) Epoch { return seed.Epochs[len(seed.Epochs)-1] }

func Accepts(epoch Epoch, hash string) bool {
	index := sort.SearchStrings(epoch.AcceptedHashes, hash)
	return index < len(epoch.AcceptedHashes) && epoch.AcceptedHashes[index] == hash
}

func SortedUniqueHashes(hashes []string) bool {
	if len(hashes) == 0 {
		return false
	}
	for index, hash := range hashes {
		if !hashPattern.MatchString(hash) || index > 0 && hashes[index-1] >= hash {
			return false
		}
	}
	return true
}

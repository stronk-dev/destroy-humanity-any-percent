// Package releasepackage derives and stages the repository-independent files
// consumed by the production gameserver. The epoch declaration remains the
// sole authority for catalog membership.
package releasepackage

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"cloud-clicker/server/epochseed"
)

var (
	ErrInvalidContent = errors.New("invalid release runtime content")
	hashPattern       = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)
)

type File struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
}

type Closure struct {
	ConstantsHash string `json:"constants_hash"`
	CopyHash      string `json:"copy_hash"`
	EpochID       int64  `json:"epoch_id"`
	Files         []File `json:"files"`
}

type contentIdentity struct {
	SchemaVersion int    `json:"schema_version"`
	ConstantsHash string `json:"constants_hash"`
	CopyHash      string `json:"copy_hash"`
}

func DeriveRuntimeClosure(root string) (Closure, error) {
	if root == "" {
		return Closure{}, ErrInvalidContent
	}
	bundle, err := epochseed.Load(root)
	if err != nil || !epochseed.Accepts(epochseed.Current(bundle.Seed), bundle.Hash) {
		return Closure{}, errors.Join(ErrInvalidContent, err)
	}
	paths := []string{epochseed.Path, "balance/transport/phase0.json", "moderation/guild-names.txt", "deployment/content-manifest.v1.json"}
	for _, artifact := range bundle.Seed.Artifacts {
		paths = append(paths, artifact.Path)
	}
	for _, epoch := range bundle.Seed.Epochs {
		paths = append(paths, epoch.ChangelogRef)
	}
	sort.Strings(paths)
	files := make([]File, 0, len(paths))
	seen := map[string]bool{}
	for _, path := range paths {
		if !validRelativePath(path) || seen[path] {
			return Closure{}, fmt.Errorf("%w: duplicate or invalid path %q", ErrInvalidContent, path)
		}
		seen[path] = true
		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(path)))
		if err != nil || len(data) == 0 {
			return Closure{}, fmt.Errorf("%w: required path %q", ErrInvalidContent, path)
		}
		files = append(files, File{Path: path, SHA256: digest(data)})
	}
	identity, err := loadContentIdentity(root)
	if err != nil || identity.ConstantsHash != bundle.Hash {
		return Closure{}, errors.Join(ErrInvalidContent, err)
	}
	return Closure{ConstantsHash: bundle.Hash, CopyHash: identity.CopyHash,
		EpochID: bundle.Seed.CurrentEpochID, Files: files}, nil
}

func StageRuntimeContent(root, destination string) (Closure, error) {
	closure, err := DeriveRuntimeClosure(root)
	if err != nil {
		return Closure{}, err
	}
	if err := requireEmptyDestination(destination); err != nil {
		return Closure{}, err
	}
	for _, file := range closure.Files {
		source := filepath.Join(root, filepath.FromSlash(file.Path))
		target := filepath.Join(destination, filepath.FromSlash(file.Path))
		data, err := os.ReadFile(source)
		if err != nil {
			return Closure{}, fmt.Errorf("%w: read %q", ErrInvalidContent, file.Path)
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return Closure{}, err
		}
		if err := os.WriteFile(target, data, 0o644); err != nil {
			return Closure{}, err
		}
	}
	if err := ValidateStagedContent(destination, closure); err != nil {
		return Closure{}, err
	}
	return closure, nil
}

func ValidateStagedContent(root string, closure Closure) error {
	if root == "" || !validClosure(closure) {
		return ErrInvalidContent
	}
	expected := make(map[string]string, len(closure.Files))
	for _, file := range closure.Files {
		expected[file.Path] = file.SHA256
	}
	seen := map[string]bool{}
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		relative = filepath.ToSlash(relative)
		want, ok := expected[relative]
		if !ok || seen[relative] || entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("%w: undeclared path %q", ErrInvalidContent, relative)
		}
		data, err := os.ReadFile(path)
		if err != nil || digest(data) != want {
			return fmt.Errorf("%w: byte mismatch %q", ErrInvalidContent, relative)
		}
		seen[relative] = true
		return nil
	})
	if err != nil {
		return errors.Join(ErrInvalidContent, err)
	}
	if len(seen) != len(expected) {
		return fmt.Errorf("%w: staged file set is incomplete", ErrInvalidContent)
	}
	return nil
}

func validClosure(closure Closure) bool {
	if closure.EpochID < 1 || !hashPattern.MatchString(closure.ConstantsHash) || !hashPattern.MatchString(closure.CopyHash) || len(closure.Files) == 0 {
		return false
	}
	prior := ""
	for _, file := range closure.Files {
		if !validRelativePath(file.Path) || !hashPattern.MatchString(file.SHA256) || prior >= file.Path {
			return false
		}
		prior = file.Path
	}
	return true
}

func loadContentIdentity(root string) (contentIdentity, error) {
	data, err := os.ReadFile(filepath.Join(root, "deployment", "content-manifest.v1.json"))
	if err != nil {
		return contentIdentity{}, err
	}
	var identity contentIdentity
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&identity) != nil || decoder.Decode(&struct{}{}) != io.EOF || identity.SchemaVersion != 1 ||
		!hashPattern.MatchString(identity.ConstantsHash) || !hashPattern.MatchString(identity.CopyHash) {
		return contentIdentity{}, ErrInvalidContent
	}
	return identity, nil
}

func requireEmptyDestination(destination string) error {
	if destination == "" {
		return ErrInvalidContent
	}
	entries, err := os.ReadDir(destination)
	if errors.Is(err, os.ErrNotExist) {
		return os.MkdirAll(destination, 0o755)
	}
	if err != nil || len(entries) != 0 {
		return ErrInvalidContent
	}
	return nil
}

func validRelativePath(value string) bool {
	return value != "" && filepath.IsLocal(filepath.FromSlash(value)) && filepath.ToSlash(filepath.Clean(filepath.FromSlash(value))) == value
}

func digest(data []byte) string {
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}

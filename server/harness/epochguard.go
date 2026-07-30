package harness

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strings"

	"cloud-clicker/server/economy"
	"cloud-clicker/server/save"
)

const epochSeedPath = "balance/epochs/phase0.json"

var epochHashPattern = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)
var epochArtifactNamePattern = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

type epochSeed struct {
	SchemaVersion  int             `json:"schema_version"`
	CurrentEpochID int64           `json:"current_epoch_id"`
	Artifacts      []epochArtifact `json:"artifacts"`
	Epochs         []epochRecord   `json:"epochs"`
}

type epochArtifact struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

type epochRecord struct {
	ID             int64    `json:"epoch_id"`
	Name           string   `json:"name"`
	ChangelogRef   string   `json:"changelog_ref"`
	AcceptedHashes []string `json:"accepted_hashes"`
}

// ComputeEpochSeedHash returns the exact constants identity described by the
// worktree seed. It is a maintenance aid: registration still requires a
// reviewed seed diff and the history guard validates the committed bytes.
func ComputeEpochSeedHash(root string) (string, error) {
	data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(epochSeedPath)))
	if err != nil {
		return "", err
	}
	seed, err := decodeEpochSeed(data)
	if err != nil {
		return "", err
	}
	artifacts := make(map[string][]byte, len(seed.Artifacts))
	for _, artifact := range seed.Artifacts {
		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(artifact.Path)))
		if err != nil {
			return "", err
		}
		artifacts[artifact.Name] = data
	}
	return save.ConstantsHashArtifacts(artifacts)
}

// ValidateRepositoryEpochChanges enforces Leaderboards L8 from the first
// committed epoch seed onward. The first seed is the history boundary; every
// later constants-artifact revision must register its exact resulting bytes.
func ValidateRepositoryEpochChanges(root string) error {
	shallow, err := gitOutput(root, "rev-parse", "--is-shallow-repository")
	if err != nil {
		return fmt.Errorf("epoch guard cannot determine history completeness: %w", err)
	}
	if string(shallow) != "false" {
		return fmt.Errorf("epoch guard requires complete git history; shallow state %q", shallow)
	}

	seedBytes, err := gitBlob(root, "HEAD", epochSeedPath)
	if err != nil {
		return fmt.Errorf("epoch guard cannot read current seed: %w", err)
	}
	seed, err := decodeEpochSeed(seedBytes)
	if err != nil {
		return fmt.Errorf("epoch guard current seed: %w", err)
	}
	paths := append([]string{epochSeedPath}, artifactPaths(seed)...)
	dirty, err := gitOutput(root, append([]string{"status", "--porcelain", "--untracked-files=all", "--"}, paths...)...)
	if err != nil {
		return fmt.Errorf("epoch guard cannot inspect worktree: %w", err)
	}
	if len(dirty) != 0 {
		return fmt.Errorf("epoch seed or constants artifacts have uncommitted changes")
	}
	history, err := gitOutput(root, "log", "--reverse", "--format=%H", "HEAD", "--", epochSeedPath)
	if err != nil {
		return fmt.Errorf("epoch guard cannot inspect seed history: %w", err)
	}
	seedCommits := strings.Fields(string(history))
	if len(seedCommits) == 0 {
		return fmt.Errorf("epoch guard found no committed seed")
	}
	commits, err := gitLines(root, "rev-list", "--reverse", seedCommits[0]+"..HEAD")
	if err != nil {
		return fmt.Errorf("epoch guard cannot inspect revisions: %w", err)
	}
	for _, commit := range commits {
		if strings.TrimSpace(commit) == "" {
			continue
		}
		if err := validateEpochRevision(root, commit); err != nil {
			return fmt.Errorf("invalid epoch registration commit %s: %w", commit, err)
		}
	}
	currentHash, err := hashArtifactsAt(root, "HEAD", seed)
	if err != nil {
		return err
	}
	if !currentEpoch(seed).accepts(currentHash) {
		return fmt.Errorf("current epoch %d does not accept resulting constants hash %s", seed.CurrentEpochID, currentHash)
	}
	return nil
}

func validateEpochRevision(root, commit string) error {
	parent, err := firstParent(root, commit)
	if err != nil {
		return err
	}
	beforeBytes, beforeErr := gitBlob(root, parent, epochSeedPath)
	afterBytes, afterErr := gitBlob(root, commit, epochSeedPath)
	if beforeErr != nil || afterErr != nil {
		return fmt.Errorf("epoch seed must remain present after its initial commit")
	}
	before, err := decodeEpochSeed(beforeBytes)
	if err != nil {
		return fmt.Errorf("parent seed: %w", err)
	}
	after, err := decodeEpochSeed(afterBytes)
	if err != nil {
		return fmt.Errorf("new seed: %w", err)
	}
	changed, err := gitLines(root, "diff-tree", "--no-commit-id", "--name-only", "-r", "-M", parent, commit)
	if err != nil {
		return err
	}
	artifactChanged := intersects(changed, artifactPaths(before)) || intersects(changed, artifactPaths(after))
	seedChanged := containsPath(changed, epochSeedPath)
	if !artifactChanged {
		if seedChanged {
			return fmt.Errorf("epoch seed changed without a constants artifact")
		}
		return nil
	}
	if !seedChanged {
		return fmt.Errorf("constants artifact changed without %s", epochSeedPath)
	}
	if !reflect.DeepEqual(before.Artifacts, after.Artifacts) {
		return fmt.Errorf("artifact identity changes require a successor RFC")
	}

	subjectBytes, err := gitOutput(root, "show", "-s", "--format=%s", commit)
	if err != nil {
		return err
	}
	balanceChange := strings.HasPrefix(string(subjectBytes), "BALANCE-CHANGE:")
	resultingHash, err := hashArtifactsAt(root, commit, after)
	if err != nil {
		return err
	}
	if !currentEpoch(after).accepts(resultingHash) {
		return fmt.Errorf("resulting constants hash %s is absent from current epoch", resultingHash)
	}
	lowered, err := loweredHardcaps(root, parent, commit, before, after)
	if err != nil {
		return err
	}
	if balanceChange {
		if err := validateMint(root, parent, commit, before, after, changed, resultingHash); err != nil {
			return err
		}
		if len(lowered) > 0 {
			changelog, err := gitBlob(root, commit, currentEpoch(after).ChangelogRef)
			if err != nil || !bytes.Contains(bytes.ToLower(changelog), []byte("cap migration:")) {
				return fmt.Errorf("lowered hardcaps %s require a 'Cap migration:' policy in the new epoch changelog", strings.Join(lowered, ","))
			}
		}
		return nil
	}
	if len(lowered) > 0 {
		return fmt.Errorf("hotfix lowers hardcaps %s; mint a BALANCE-CHANGE epoch", strings.Join(lowered, ","))
	}
	return validateHotfix(before, after, resultingHash)
}

func validateMint(root, parent, commit string, before, after epochSeed, changed []string, resultingHash string) error {
	if len(after.Epochs) != len(before.Epochs)+1 || after.CurrentEpochID != before.CurrentEpochID+1 ||
		!reflect.DeepEqual(after.Epochs[:len(before.Epochs)], before.Epochs) {
		return fmt.Errorf("BALANCE-CHANGE must append exactly one immutable epoch")
	}
	added := currentEpoch(after)
	if added.ID != after.CurrentEpochID || !added.accepts(resultingHash) {
		return fmt.Errorf("new epoch does not own resulting constants hash")
	}
	if !containsPath(changed, added.ChangelogRef) {
		return fmt.Errorf("new epoch changelog %s must be added in the same commit", added.ChangelogRef)
	}
	if _, err := gitBlob(root, parent, added.ChangelogRef); err == nil {
		return fmt.Errorf("new epoch changelog %s already existed before the mint", added.ChangelogRef)
	}
	if _, err := gitBlob(root, commit, added.ChangelogRef); err != nil {
		return fmt.Errorf("new epoch changelog %s is absent from the mint", added.ChangelogRef)
	}
	return nil
}

func validateHotfix(before, after epochSeed, resultingHash string) error {
	if before.CurrentEpochID != after.CurrentEpochID || len(before.Epochs) != len(after.Epochs) || len(after.Epochs) == 0 ||
		!reflect.DeepEqual(before.Epochs[:len(before.Epochs)-1], after.Epochs[:len(after.Epochs)-1]) {
		return fmt.Errorf("hotfix may only extend the current epoch accepted set")
	}
	oldCurrent, newCurrent := currentEpoch(before), currentEpoch(after)
	if oldCurrent.ID != newCurrent.ID || oldCurrent.Name != newCurrent.Name || oldCurrent.ChangelogRef != newCurrent.ChangelogRef {
		return fmt.Errorf("hotfix changed current epoch identity")
	}
	if !isAppendOnlySet(oldCurrent.AcceptedHashes, newCurrent.AcceptedHashes, resultingHash) {
		return fmt.Errorf("hotfix must append only the resulting constants hash")
	}
	return nil
}

func decodeEpochSeed(data []byte) (epochSeed, error) {
	var seed epochSeed
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&seed); err != nil {
		return seed, err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return seed, fmt.Errorf("epoch seed must contain exactly one JSON value")
	}
	if seed.SchemaVersion != 1 || seed.CurrentEpochID < 1 || len(seed.Artifacts) == 0 || len(seed.Epochs) == 0 {
		return seed, fmt.Errorf("invalid epoch seed root")
	}
	seenNames, seenPaths := map[string]bool{}, map[string]bool{}
	for _, artifact := range seed.Artifacts {
		if !epochArtifactNamePattern.MatchString(artifact.Name) || path.Clean(artifact.Path) != artifact.Path ||
			seenNames[artifact.Name] || seenPaths[artifact.Path] || !strings.HasPrefix(artifact.Path, "balance/") {
			return seed, fmt.Errorf("invalid or duplicate artifact %q", artifact.Name)
		}
		seenNames[artifact.Name], seenPaths[artifact.Path] = true, true
	}
	for index, epoch := range seed.Epochs {
		if epoch.ID != int64(index+1) || epoch.Name == "" || epoch.ChangelogRef != fmt.Sprintf("changelog/epoch-%d.md", epoch.ID) || len(epoch.AcceptedHashes) == 0 {
			return seed, fmt.Errorf("invalid epoch %d", epoch.ID)
		}
		if !sortedUniqueHashes(epoch.AcceptedHashes) {
			return seed, fmt.Errorf("epoch %d hashes must be sorted and unique", epoch.ID)
		}
	}
	if seed.CurrentEpochID != seed.Epochs[len(seed.Epochs)-1].ID {
		return seed, fmt.Errorf("current epoch is not the last epoch")
	}
	return seed, nil
}

func loweredHardcaps(root, parent, commit string, before, after epochSeed) ([]string, error) {
	beforePath, beforeOK := artifactPath(before, "economy")
	afterPath, afterOK := artifactPath(after, "economy")
	if !beforeOK || !afterOK {
		return nil, fmt.Errorf("epoch artifacts must include economy")
	}
	beforeBytes, err := gitBlob(root, parent, beforePath)
	if err != nil {
		return nil, err
	}
	afterBytes, err := gitBlob(root, commit, afterPath)
	if err != nil {
		return nil, err
	}
	oldCatalog, err := economy.LoadCatalog(beforeBytes)
	if err != nil {
		return nil, fmt.Errorf("parent economy catalog: %w", err)
	}
	newCatalog, err := economy.LoadCatalog(afterBytes)
	if err != nil {
		return nil, fmt.Errorf("new economy catalog: %w", err)
	}
	var lowered []string
	for _, oldResource := range oldCatalog.Resources() {
		newResource, ok := newCatalog.Resource(oldResource.ID)
		if !ok || oldResource.Hardcap == nil || newResource.Hardcap == nil {
			continue
		}
		if newResource.Hardcap.Amount.Lt(oldResource.Hardcap.Amount) {
			lowered = append(lowered, oldResource.ID)
		}
	}
	sort.Strings(lowered)
	return lowered, nil
}

func hashArtifactsAt(root, commit string, seed epochSeed) (string, error) {
	artifacts := make(map[string][]byte, len(seed.Artifacts))
	for _, artifact := range seed.Artifacts {
		data, err := gitBlob(root, commit, artifact.Path)
		if err != nil {
			return "", fmt.Errorf("cannot read artifact %s at %s: %w", artifact.Path, commit, err)
		}
		artifacts[artifact.Name] = data
	}
	return save.ConstantsHashArtifacts(artifacts)
}

func gitBlob(root, commit, path string) ([]byte, error) {
	command := exec.Command("git", "show", commit+":"+path)
	command.Dir = root
	output, err := command.Output()
	if err != nil {
		return nil, fmt.Errorf("git show %s:%s: %w", commit, path, err)
	}
	return output, nil
}

func artifactPaths(seed epochSeed) []string {
	paths := make([]string, 0, len(seed.Artifacts))
	for _, artifact := range seed.Artifacts {
		paths = append(paths, artifact.Path)
	}
	return paths
}

func artifactPath(seed epochSeed, name string) (string, bool) {
	for _, artifact := range seed.Artifacts {
		if artifact.Name == name {
			return artifact.Path, true
		}
	}
	return "", false
}

func currentEpoch(seed epochSeed) epochRecord { return seed.Epochs[len(seed.Epochs)-1] }

func (epoch epochRecord) accepts(hash string) bool {
	index := sort.SearchStrings(epoch.AcceptedHashes, hash)
	return index < len(epoch.AcceptedHashes) && epoch.AcceptedHashes[index] == hash
}

func sortedUniqueHashes(hashes []string) bool {
	for index, hash := range hashes {
		if !epochHashPattern.MatchString(hash) || index > 0 && hashes[index-1] >= hash {
			return false
		}
	}
	return true
}

func isAppendOnlySet(before, after []string, required string) bool {
	if len(after) != len(before)+1 || !sortedUniqueHashes(after) {
		return false
	}
	seenRequired := false
	for _, hash := range after {
		if hash == required {
			seenRequired = true
			continue
		}
		if !containsPath(before, hash) {
			return false
		}
	}
	return seenRequired
}

func intersects(changed, watched []string) bool {
	for _, path := range watched {
		if containsPath(changed, path) {
			return true
		}
	}
	return false
}

func containsPath(paths []string, wanted string) bool {
	for _, path := range paths {
		if strings.TrimSpace(path) == wanted {
			return true
		}
	}
	return false
}

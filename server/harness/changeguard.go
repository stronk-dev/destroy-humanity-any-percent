package harness

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"reflect"
	"strings"
)

const (
	baselinePath        = "testdata/harness/pacing-baseline.json"
	goldenPath          = "testdata/harness/golden-seed.json"
	relevanceGoldenPath = "testdata/harness/relevance/golden-report-v1.json"
)

var baselineArtifactPaths = map[string]struct{}{
	baselinePath:        {},
	goldenPath:          {},
	relevanceGoldenPath: {},
}

// ValidateBaselineCommit enforces the separate-commit review protocol for one
// non-initial baseline revision. inputsBefore contains paths changed after the
// previous baseline revision and before this artifact commit.
func ValidateBaselineCommit(commitPaths, inputsBefore []string, subject string) error {
	baselineChanged := false
	for _, path := range commitPaths {
		path = strings.TrimSpace(path)
		if path == "" {
			continue
		}
		if path == baselinePath || path == relevanceGoldenPath {
			baselineChanged = true
		}
		if _, allowed := baselineArtifactPaths[path]; !allowed {
			return fmt.Errorf("baseline artifact commit also changes %s", path)
		}
	}
	if !baselineChanged {
		return fmt.Errorf("baseline artifact commit does not change a governed report")
	}
	if !strings.HasPrefix(subject, "BALANCE-CHANGE:") {
		if strings.HasPrefix(subject, "CONSTANTS-IDENTITY:") {
			return nil
		}
		return fmt.Errorf("baseline rewrite commit subject must begin BALANCE-CHANGE: or CONSTANTS-IDENTITY:")
	}
	for _, path := range inputsBefore {
		path = strings.TrimSpace(path)
		if strings.HasPrefix(path, "balance/catalogs/") || strings.HasPrefix(path, "balance/categories/") || strings.HasPrefix(path, "balance/routes/") || strings.HasPrefix(path, "balance/commons/") || strings.HasPrefix(path, "balance/factions/") || strings.HasPrefix(path, "balance/prestige/") || strings.HasPrefix(path, "balance/guilds/") || strings.HasPrefix(path, "testdata/harness/scenarios/") || strings.HasPrefix(path, "testdata/harness/relevance/") {
			return nil
		}
	}
	return fmt.Errorf("baseline rewrite has no changed balance artifact or scenario before its artifact commit")
}

// ValidateRepositoryBaselineChange validates every reachable non-initial
// baseline revision. It intentionally uses no CI-provider metadata, so local
// and hosted checks enforce the same repository history.
func ValidateRepositoryBaselineChange(root string) error {
	shallow, err := gitOutput(root, "rev-parse", "--is-shallow-repository")
	if err != nil {
		return fmt.Errorf("baseline guard cannot determine history completeness: %w", err)
	}
	switch string(shallow) {
	case "false":
	case "true":
		return fmt.Errorf("baseline guard requires complete git history; shallow repository detected")
	default:
		return fmt.Errorf("baseline guard received ambiguous shallow-repository result %q", shallow)
	}

	dirty, err := gitOutput(root, "status", "--porcelain", "--untracked-files=all", "--", baselinePath, goldenPath, relevanceGoldenPath)
	if err != nil {
		return fmt.Errorf("baseline guard cannot inspect artifact worktree: %w", err)
	}
	if len(dirty) != 0 {
		return fmt.Errorf("baseline artifacts have uncommitted changes")
	}

	history, err := gitOutput(root, "log", "--reverse", "--format=%H", "HEAD", "--", baselinePath, relevanceGoldenPath)
	if err != nil {
		return fmt.Errorf("baseline guard requires complete baseline history: %w", err)
	}
	commits := strings.Fields(string(history))
	if len(commits) == 0 {
		return fmt.Errorf("baseline guard found no committed baseline")
	}
	for index := 1; index < len(commits); index++ {
		commit := commits[index]
		parent, err := firstParent(root, commit)
		if err != nil {
			return err
		}
		commitPaths, err := gitLines(root, "diff-tree", "--no-commit-id", "--name-only", "-r", "-M", parent, commit)
		if err != nil {
			return fmt.Errorf("baseline guard cannot inspect artifact commit %s: %w", commit, err)
		}
		inputPaths, err := gitLines(root, "diff", "--name-only", commits[index-1], parent)
		if err != nil {
			return fmt.Errorf("baseline guard cannot inspect inputs before %s: %w", commit, err)
		}
		subject, err := gitOutput(root, "show", "-s", "--format=%s", commit)
		if err != nil {
			return fmt.Errorf("baseline guard cannot inspect subject for %s: %w", commit, err)
		}
		if err := ValidateBaselineCommit(commitPaths, inputPaths, string(subject)); err != nil {
			return fmt.Errorf("invalid baseline commit %s: %w", commit, err)
		}
		if strings.HasPrefix(string(subject), "CONSTANTS-IDENTITY:") {
			if err := validateConstantsIdentityCommit(root, commits[index-1], parent, commit); err != nil {
				return fmt.Errorf("invalid constants-identity commit %s: %w", commit, err)
			}
		}
	}
	return nil
}

func validateConstantsIdentityCommit(root, previousBaseline, parent, commit string) error {
	seedBytes, err := gitBlob(root, commit, epochSeedPath)
	if err != nil {
		return err
	}
	seed, err := decodeEpochSeed(seedBytes)
	if err != nil {
		return err
	}
	if err := validateConstantsIdentityArtifactBytes(root, previousBaseline, commit, seed); err != nil {
		return err
	}
	expectedHash, err := hashArtifactsAt(root, commit, seed)
	if err != nil {
		return err
	}
	beforeBaseline, err := gitBlob(root, parent, baselinePath)
	if err != nil {
		return err
	}
	afterBaseline, err := gitBlob(root, commit, baselinePath)
	if err != nil {
		return err
	}
	beforeGolden, err := gitBlob(root, parent, goldenPath)
	if err != nil {
		return err
	}
	afterGolden, err := gitBlob(root, commit, goldenPath)
	if err != nil {
		return err
	}
	return validateConstantsIdentityBlobs(beforeBaseline, afterBaseline, beforeGolden, afterGolden, expectedHash)
}

func validateConstantsIdentityArtifactBytes(root, previousBaseline, commit string, current epochSeed) error {
	previousSeed, present, err := epochSeedAt(root, previousBaseline)
	if err != nil {
		return err
	}
	// The repository's first identity repair predates the manifest at its previous
	// baseline. There is no declared bundle to pin across that one migration.
	if !present {
		return nil
	}
	if !reflect.DeepEqual(previousSeed.Artifacts, current.Artifacts) {
		return fmt.Errorf("constants-identity commit changes the seed artifact set")
	}
	for _, artifact := range current.Artifacts {
		before, err := gitBlob(root, previousBaseline, artifact.Path)
		if err != nil {
			return err
		}
		after, err := gitBlob(root, commit, artifact.Path)
		if err != nil {
			return err
		}
		if !bytes.Equal(before, after) {
			return fmt.Errorf("constants-identity commit changes artifact bytes for %s", artifact.Name)
		}
	}
	return nil
}

func epochSeedAt(root, commit string) (epochSeed, bool, error) {
	history, err := gitOutput(root, "log", "-1", "--format=%H", commit, "--", epochSeedPath)
	if err != nil {
		return epochSeed{}, false, err
	}
	if len(history) == 0 {
		return epochSeed{}, false, nil
	}
	data, err := gitBlob(root, commit, epochSeedPath)
	if err != nil {
		return epochSeed{}, false, err
	}
	seed, err := decodeEpochSeed(data)
	return seed, true, err
}

func validateConstantsIdentityBlobs(beforeBaseline, afterBaseline, beforeGolden, afterGolden []byte, expectedHash string) error {
	var oldBaseline, newBaseline AggregateReport
	if err := decodeStrictJSON(beforeBaseline, &oldBaseline); err != nil {
		return err
	}
	if err := decodeStrictJSON(afterBaseline, &newBaseline); err != nil {
		return err
	}
	if newBaseline.ConstantsHash != expectedHash {
		return fmt.Errorf("baseline hash %s differs from epoch manifest %s", newBaseline.ConstantsHash, expectedHash)
	}
	oldBaseline.ConstantsHash, newBaseline.ConstantsHash = "", ""
	if !reflect.DeepEqual(oldBaseline, newBaseline) {
		return fmt.Errorf("constants-identity commit changes pacing baseline content")
	}
	var oldGolden, newGolden GoldenReport
	if err := decodeStrictJSON(beforeGolden, &oldGolden); err != nil {
		return err
	}
	if err := decodeStrictJSON(afterGolden, &newGolden); err != nil {
		return err
	}
	for index := range newGolden.Runs {
		if newGolden.Runs[index].Key.ConstantsHash != expectedHash {
			return fmt.Errorf("golden run %d hash differs from epoch manifest", index)
		}
		newGolden.Runs[index].Key.ConstantsHash = ""
	}
	for index := range oldGolden.Runs {
		oldGolden.Runs[index].Key.ConstantsHash = ""
	}
	if !reflect.DeepEqual(oldGolden, newGolden) {
		return fmt.Errorf("constants-identity commit changes golden behavior")
	}
	return nil
}

func decodeStrictJSON(data []byte, target any) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return fmt.Errorf("artifact must contain exactly one JSON value")
	}
	return nil
}

func firstParent(root, commit string) (string, error) {
	line, err := gitOutput(root, "rev-list", "--parents", "-n", "1", commit)
	if err != nil {
		return "", fmt.Errorf("baseline guard cannot inspect parent for %s: %w", commit, err)
	}
	fields := strings.Fields(string(line))
	if len(fields) < 2 || fields[0] != commit {
		return "", fmt.Errorf("baseline guard found no parent for non-initial baseline commit %s", commit)
	}
	return fields[1], nil
}

func gitLines(root string, arguments ...string) ([]string, error) {
	output, err := gitOutput(root, arguments...)
	if err != nil || len(output) == 0 {
		return nil, err
	}
	return strings.Split(string(output), "\n"), nil
}

func gitOutput(root string, arguments ...string) ([]byte, error) {
	command := exec.Command("git", arguments...)
	command.Dir = root
	output, err := command.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("git %s: %w: %s", strings.Join(arguments, " "), err, strings.TrimSpace(string(output)))
	}
	return bytes.TrimSpace(output), nil
}

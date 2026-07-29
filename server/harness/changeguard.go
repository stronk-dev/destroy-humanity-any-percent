package harness

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

const (
	baselinePath = "testdata/harness/pacing-baseline.json"
	goldenPath   = "testdata/harness/golden-seed.json"
)

var baselineArtifactPaths = map[string]struct{}{
	baselinePath: {},
	goldenPath:   {},
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
		if path == baselinePath {
			baselineChanged = true
		}
		if _, allowed := baselineArtifactPaths[path]; !allowed {
			return fmt.Errorf("baseline artifact commit also changes %s", path)
		}
	}
	if !baselineChanged {
		return fmt.Errorf("baseline artifact commit does not change %s", baselinePath)
	}
	if !strings.HasPrefix(subject, "BALANCE-CHANGE:") {
		return fmt.Errorf("baseline rewrite commit subject must begin BALANCE-CHANGE:")
	}
	for _, path := range inputsBefore {
		path = strings.TrimSpace(path)
		if strings.HasPrefix(path, "balance/catalogs/") || strings.HasPrefix(path, "testdata/harness/scenarios/") {
			return nil
		}
	}
	return fmt.Errorf("baseline rewrite has no changed catalog or scenario before its artifact commit")
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

	dirty, err := gitOutput(root, "status", "--porcelain", "--untracked-files=all", "--", baselinePath, goldenPath)
	if err != nil {
		return fmt.Errorf("baseline guard cannot inspect artifact worktree: %w", err)
	}
	if len(dirty) != 0 {
		return fmt.Errorf("baseline artifacts have uncommitted changes")
	}

	history, err := gitOutput(root, "log", "--reverse", "--format=%H", "HEAD", "--", baselinePath)
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

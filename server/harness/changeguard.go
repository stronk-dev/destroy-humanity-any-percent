package harness

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

const baselinePath = "testdata/harness/pacing-baseline.json"

// ValidateBaselineChange enforces the review protocol independently of CI
// provider metadata. The first checked-in baseline is allowed; later rewrites
// require both changed inputs and a BALANCE-CHANGE commit subject.
func ValidateBaselineChange(changes []string, subject string) error {
	baselineStatus := ""
	inputsChanged := false
	for _, line := range changes {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		path := fields[len(fields)-1]
		if path == baselinePath {
			baselineStatus = fields[0]
		}
		if strings.HasPrefix(path, "balance/catalogs/") || strings.HasPrefix(path, "testdata/harness/scenarios/") {
			inputsChanged = true
		}
	}
	if strings.HasPrefix(baselineStatus, "A") || baselineStatus == "" {
		return nil
	}
	if !inputsChanged {
		return fmt.Errorf("baseline rewrite has no changed catalog or scenario")
	}
	if !strings.HasPrefix(subject, "BALANCE-CHANGE:") {
		return fmt.Errorf("baseline rewrite commit subject must begin BALANCE-CHANGE:")
	}
	return nil
}

func ValidateRepositoryBaselineChange(root string) error {
	headChanges, err := gitOutput(root, "diff", "--name-status", "HEAD^", "HEAD")
	if err != nil {
		// A worktree without a parent is the initial repository state. Other git
		// failures are loud because a shallow CI checkout could bypass the gate.
		count, headErr := gitOutput(root, "rev-list", "--count", "HEAD")
		if headErr == nil && string(count) == "1" {
			return nil
		}
		return fmt.Errorf("baseline guard requires HEAD parent: %w", err)
	}
	if !changesPath(headChanges, baselinePath) {
		return nil
	}
	// Baseline regeneration is deliberately a separate reviewable commit.
	// Inspect inputs changed since the prior baseline revision instead of
	// requiring the input and its review artifact in the same commit.
	history, err := gitOutput(root, "log", "--format=%H", "--", baselinePath)
	if err != nil {
		return fmt.Errorf("baseline guard requires history: %w", err)
	}
	commits := strings.Fields(string(history))
	if len(commits) < 2 {
		return nil
	}
	changes, err := gitOutput(root, "diff", "--name-status", commits[1], "HEAD")
	if err != nil {
		return fmt.Errorf("baseline guard cannot inspect accepted inputs: %w", err)
	}
	subject, err := gitOutput(root, "log", "-1", "--format=%s")
	if err != nil {
		return err
	}
	lines := strings.Split(strings.TrimSpace(string(changes)), "\n")
	return ValidateBaselineChange(lines, strings.TrimSpace(string(subject)))
}

func changesPath(changes []byte, path string) bool {
	for _, line := range strings.Split(strings.TrimSpace(string(changes)), "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[len(fields)-1] == path {
			return true
		}
	}
	return false
}

func gitOutput(root string, arguments ...string) ([]byte, error) {
	command := exec.Command("git", arguments...)
	command.Dir = root
	output, err := command.Output()
	if err != nil {
		return nil, err
	}
	return bytes.TrimSpace(output), nil
}

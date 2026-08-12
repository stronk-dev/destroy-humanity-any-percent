package harness

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestRepositoryGuardChecksCoveredBaselineCommit(t *testing.T) {
	root := newGuardRepository(t)
	writeGuardCommit(t, root, "economy: retune", map[string]string{
		"balance/catalogs/phase0.json": `{"value":2}`,
	})
	writeGuardCommit(t, root, "ordinary baseline rewrite", map[string]string{
		baselinePath: `{"baseline":2}`,
		goldenPath:   `{"golden":2}`,
	})
	writeGuardCommit(t, root, "docs: cover", map[string]string{"README.md": "cover\n"})
	if err := ValidateRepositoryBaselineChange(root); err == nil || !strings.Contains(err.Error(), "subject") {
		t.Fatalf("covered invalid baseline err=%v", err)
	}
}

func TestRepositoryGuardRejectsSmuggledPath(t *testing.T) {
	root := newGuardRepository(t)
	writeGuardCommit(t, root, "economy: retune", map[string]string{
		"balance/catalogs/phase0.json": `{"value":2}`,
	})
	writeGuardCommit(t, root, "BALANCE-CHANGE: smuggled code", map[string]string{
		baselinePath:     `{"baseline":2}`,
		goldenPath:       `{"golden":2}`,
		"server/code.go": "package server\n\nconst Smuggled = true\n",
	})
	if err := ValidateRepositoryBaselineChange(root); err == nil || !strings.Contains(err.Error(), "server/code.go") {
		t.Fatalf("smuggled path err=%v", err)
	}
}

func TestRepositoryGuardAcceptsNamedPublishedMixedCommitCorrection(t *testing.T) {
	root := newGuardRepository(t)
	writeGuardCommit(t, root, "harness: implementation input", map[string]string{
		"testdata/harness/relevance/scenario-v1.json": `{"version":2}`,
	})
	writeGuardCommit(t, root, "harness: mixed implementation and golden", map[string]string{
		relevanceGoldenPath: `{"schema_version":1}`,
		"server/code.go":    "package server\n\nconst Mixed = true\n",
	})
	offendingBytes, err := gitOutput(root, "rev-parse", "HEAD")
	if err != nil {
		t.Fatal(err)
	}
	offending := string(offendingBytes)
	if err := ValidateRepositoryBaselineChange(root); err == nil || !strings.Contains(err.Error(), "server/code.go") {
		t.Fatalf("mixed commit unexpectedly passed before correction: %v", err)
	}
	runGuardGit(t, root, "update-ref", publishedMainRef, offending)
	writeGuardCommit(t, root, "docs: record published packaging correction", map[string]string{
		"planning/test/log.md": "published correction record\n",
	})
	registry := fmt.Sprintf(`{"schema_version":1,"corrections":[{"offending_commit":%q,"kind":"mixed_artifact_commit","reason":"Published artifact packaging cannot be rewritten.","review_log":"planning/test/log.md"}]}`, offending)
	writeGuardCommit(t, root, "guard: register published baseline correction", map[string]string{
		baselineHistoryCorrectionsPath: registry,
	})
	if err := ValidateRepositoryBaselineChange(root); err != nil {
		t.Fatalf("named published correction failed: %v", err)
	}

	writeGuardCommit(t, root, "guard: illegally remove correction", map[string]string{
		baselineHistoryCorrectionsPath: `{"schema_version":1,"corrections":[]}`,
	})
	if err := ValidateRepositoryBaselineChange(root); err == nil || !strings.Contains(err.Error(), "does not append") {
		t.Fatalf("non-append-only correction registry err=%v", err)
	}
}

func TestRepositoryGuardRejectsUnpublishedMixedCommitCorrection(t *testing.T) {
	root := newGuardRepository(t)
	initialBytes, err := gitOutput(root, "rev-parse", "HEAD")
	if err != nil {
		t.Fatal(err)
	}
	runGuardGit(t, root, "update-ref", publishedMainRef, string(initialBytes))
	writeGuardCommit(t, root, "harness: mixed implementation and golden", map[string]string{
		relevanceGoldenPath: `{"schema_version":1}`,
		"server/code.go":    "package server\n\nconst Mixed = true\n",
	})
	offendingBytes, err := gitOutput(root, "rev-parse", "HEAD")
	if err != nil {
		t.Fatal(err)
	}
	offending := string(offendingBytes)
	writeGuardCommit(t, root, "docs: correction record", map[string]string{
		"planning/test/log.md": "unpublished correction must fail\n",
	})
	registry := fmt.Sprintf(`{"schema_version":1,"corrections":[{"offending_commit":%q,"kind":"mixed_artifact_commit","reason":"This commit is not published.","review_log":"planning/test/log.md"}]}`, offending)
	writeGuardCommit(t, root, "guard: attempt unpublished baseline correction", map[string]string{
		baselineHistoryCorrectionsPath: registry,
	})
	if err := ValidateRepositoryBaselineChange(root); err == nil || !strings.Contains(err.Error(), "unpublished commits must be repackaged, not forgiven") {
		t.Fatalf("unpublished correction err=%v", err)
	}
}

func TestRepositoryGuardRejectsMissingPublicationRefAsConfigurationError(t *testing.T) {
	root := newGuardRepository(t)
	writeGuardCommit(t, root, "harness: mixed implementation and golden", map[string]string{
		relevanceGoldenPath: `{"schema_version":1}`,
		"server/code.go":    "package server\n\nconst Mixed = true\n",
	})
	offendingBytes, err := gitOutput(root, "rev-parse", "HEAD")
	if err != nil {
		t.Fatal(err)
	}
	writeGuardCommit(t, root, "docs: correction record", map[string]string{
		"planning/test/log.md": "missing publication ref must fail as configuration\n",
	})
	registry := fmt.Sprintf(`{"schema_version":1,"corrections":[{"offending_commit":%q,"kind":"mixed_artifact_commit","reason":"Publication cannot be established.","review_log":"planning/test/log.md"}]}`, string(offendingBytes))
	writeGuardCommit(t, root, "guard: attempt correction without publication ref", map[string]string{
		baselineHistoryCorrectionsPath: registry,
	})
	err = ValidateRepositoryBaselineChange(root)
	if err == nil || !strings.Contains(err.Error(), "publication check is not configured") || !strings.Contains(err.Error(), publishedMainRef) {
		t.Fatalf("missing publication ref err=%v", err)
	}
}

func TestRepositoryGuardRejectsCorrectionForUngovernedCommit(t *testing.T) {
	root := newGuardRepository(t)
	writeGuardCommit(t, root, "docs: ordinary unpublished-independent change", map[string]string{
		"README.md": "ordinary change\n",
	})
	offendingBytes, err := gitOutput(root, "rev-parse", "HEAD")
	if err != nil {
		t.Fatal(err)
	}
	offending := string(offendingBytes)
	runGuardGit(t, root, "update-ref", publishedMainRef, offending)
	writeGuardCommit(t, root, "docs: correction record", map[string]string{
		"planning/test/log.md": "blanket amnesty must fail\n",
	})
	registry := fmt.Sprintf(`{"schema_version":1,"corrections":[{"offending_commit":%q,"kind":"mixed_artifact_commit","reason":"This commit changed no governed report.","review_log":"planning/test/log.md"}]}`, offending)
	writeGuardCommit(t, root, "guard: attempt blanket baseline correction", map[string]string{
		baselineHistoryCorrectionsPath: registry,
	})
	if err := ValidateRepositoryBaselineChange(root); err == nil || !strings.Contains(err.Error(), "ledger entry was never consumed") {
		t.Fatalf("blanket correction err=%v", err)
	}
}

func TestRepositoryGuardRejectsCorrectionCommitWithSmuggledPath(t *testing.T) {
	root := newGuardRepository(t)
	writeGuardCommit(t, root, "harness: mixed implementation and golden", map[string]string{
		relevanceGoldenPath: `{"schema_version":1}`,
		"server/code.go":    "package server\n\nconst Mixed = true\n",
	})
	offendingBytes, err := gitOutput(root, "rev-parse", "HEAD")
	if err != nil {
		t.Fatal(err)
	}
	writeGuardCommit(t, root, "docs: correction record", map[string]string{"planning/test/log.md": "record\n"})
	registry := fmt.Sprintf(`{"schema_version":1,"corrections":[{"offending_commit":%q,"kind":"mixed_artifact_commit","reason":"Published artifact packaging cannot be rewritten.","review_log":"planning/test/log.md"}]}`, string(offendingBytes))
	writeGuardCommit(t, root, "guard: smuggle with correction", map[string]string{
		baselineHistoryCorrectionsPath: registry,
		"README.md":                    "smuggled\n",
	})
	if err := ValidateRepositoryBaselineChange(root); err == nil || !strings.Contains(err.Error(), "must change only") {
		t.Fatalf("smuggled correction commit err=%v", err)
	}
}

func TestRepositoryGuardAcceptsSeparateInputAndArtifacts(t *testing.T) {
	root := newGuardRepository(t)
	writeGuardCommit(t, root, "economy: retune", map[string]string{
		"balance/catalogs/phase0.json": `{"value":2}`,
	})
	writeGuardCommit(t, root, "BALANCE-CHANGE: reviewed retune", map[string]string{
		baselinePath: `{"baseline":2}`,
		goldenPath:   `{"golden":2}`,
	})
	writeGuardCommit(t, root, "docs: explain retune", map[string]string{"README.md": "documented\n"})
	if err := ValidateRepositoryBaselineChange(root); err != nil {
		t.Fatal(err)
	}
}

func TestRepositoryGuardAcceptsSeparateRelevanceFixtureAndGolden(t *testing.T) {
	root := newGuardRepository(t)
	writeGuardCommit(t, root, "harness: add relevance fixture", map[string]string{
		"testdata/harness/relevance/scenario-v1.json": `{"version":1}`,
	})
	writeGuardCommit(t, root, "BALANCE-CHANGE: add relevance golden", map[string]string{
		relevanceGoldenPath: `{"schema_version":1}`,
	})
	if err := ValidateRepositoryBaselineChange(root); err != nil {
		t.Fatal(err)
	}
}

func TestRepositoryGuardDiscoversEveryRegisteredRelevanceGolden(t *testing.T) {
	root := newGuardRepository(t)
	secondGolden := "testdata/harness/relevance/golden-report-v2.json"
	registry := `{"schema_version":1,"entries":[` +
		`{"economy_catalog":"balance/catalogs/phase0.json","scenario":"testdata/harness/relevance/scenario-v1.json","relevance_policy":"testdata/harness/relevance/policy-v1.json","golden_report":"` + relevanceGoldenPath + `","justification_changelog":"testdata/harness/relevance/CHANGELOG.md"},` +
		`{"economy_catalog":"balance/catalogs/phase1.json","scenario":"testdata/harness/relevance/scenario-v2.json","relevance_policy":"testdata/harness/relevance/policy-v2.json","golden_report":"` + secondGolden + `","justification_changelog":"testdata/harness/relevance/CHANGELOG.md"}]}`
	writeGuardCommit(t, root, "harness: register second relevance scenario", map[string]string{
		relevanceRegistryPath:                         registry,
		"testdata/harness/relevance/scenario-v2.json": `{"version":2}`,
	})
	writeGuardCommit(t, root, "BALANCE-CHANGE: add second relevance golden", map[string]string{
		secondGolden: `{"schema_version":1}`,
	})
	if err := ValidateRepositoryBaselineChange(root); err != nil {
		t.Fatal(err)
	}
	writeGuardFile(t, root, secondGolden, `{"dirty":true}`)
	if err := ValidateRepositoryBaselineChange(root); err == nil || !strings.Contains(err.Error(), "uncommitted") {
		t.Fatalf("dynamic dirty relevance golden err=%v", err)
	}
	writeGuardFile(t, root, secondGolden, `{"schema_version":1}`)
	writeGuardCommit(t, root, "harness: retire second relevance scenario", map[string]string{
		relevanceRegistryPath: `{"schema_version":1,"entries":[{"economy_catalog":"balance/catalogs/phase0.json","scenario":"testdata/harness/relevance/scenario-v1.json","relevance_policy":"testdata/harness/relevance/policy-v1.json","golden_report":"testdata/harness/relevance/golden-report-v1.json","justification_changelog":"testdata/harness/relevance/CHANGELOG.md"}]}`,
	})
	writeGuardFile(t, root, secondGolden, `{"dirty":"hidden-after-retirement"}`)
	if err := ValidateRepositoryBaselineChange(root); err == nil || !strings.Contains(err.Error(), "uncommitted") {
		t.Fatalf("historical relevance golden escaped governance: %v", err)
	}
}

func TestRepositoryGuardRejectsNoPriorInputAndDirtyArtifacts(t *testing.T) {
	root := newGuardRepository(t)
	writeGuardCommit(t, root, "BALANCE-CHANGE: artifact only", map[string]string{
		baselinePath: `{"baseline":2}`,
	})
	if err := ValidateRepositoryBaselineChange(root); err == nil || !strings.Contains(err.Error(), "no changed balance artifact or scenario") {
		t.Fatalf("artifact-only err=%v", err)
	}

	clean := newGuardRepository(t)
	writeGuardFile(t, clean, baselinePath, `{"dirty":true}`)
	if err := ValidateRepositoryBaselineChange(clean); err == nil || !strings.Contains(err.Error(), "uncommitted") {
		t.Fatalf("dirty artifact err=%v", err)
	}
}

func TestRepositoryGuardRejectsShallowHistory(t *testing.T) {
	source := newGuardRepository(t)
	writeGuardCommit(t, source, "economy: retune", map[string]string{
		"balance/catalogs/phase0.json": `{"value":2}`,
	})
	writeGuardCommit(t, source, "BALANCE-CHANGE: reviewed retune", map[string]string{
		baselinePath: `{"baseline":2}`,
		goldenPath:   `{"golden":2}`,
	})
	writeGuardCommit(t, source, "docs: cover", map[string]string{"README.md": "cover\n"})
	destination := filepath.Join(t.TempDir(), "shallow")
	sourceURL := (&url.URL{Scheme: "file", Path: source}).String()
	runGuardGit(t, "", "clone", "--quiet", "--depth=2", sourceURL, destination)
	if err := ValidateRepositoryBaselineChange(destination); err == nil || !strings.Contains(err.Error(), "shallow") {
		t.Fatalf("shallow history err=%v", err)
	}
}

func newGuardRepository(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	runGuardGit(t, root, "init", "--quiet", "--initial-branch=main")
	runGuardGit(t, root, "config", "user.name", "Harness Test")
	runGuardGit(t, root, "config", "user.email", "harness@example.invalid")
	writeGuardFile(t, root, baselinePath, `{"baseline":1}`)
	writeGuardFile(t, root, goldenPath, `{"golden":1}`)
	writeGuardFile(t, root, relevanceGoldenPath, `{"schema_version":0}`)
	writeGuardFile(t, root, relevanceRegistryPath, `{"schema_version":1,"entries":[{"economy_catalog":"balance/catalogs/phase0.json","scenario":"testdata/harness/relevance/scenario-v1.json","relevance_policy":"testdata/harness/relevance/policy-v1.json","golden_report":"testdata/harness/relevance/golden-report-v1.json","justification_changelog":"testdata/harness/relevance/CHANGELOG.md"}]}`)
	writeGuardFile(t, root, contentDynamicsRegistryPath, `{"schema_version":1,"entries":[]}`)
	writeGuardFile(t, root, "balance/catalogs/phase0.json", `{"value":1}`)
	writeGuardFile(t, root, "server/code.go", "package server\n")
	runGuardGit(t, root, "add", ".")
	runGuardGit(t, root, "commit", "--quiet", "-m", "harness: initial baseline")
	return root
}

func writeGuardCommit(t *testing.T, root, subject string, files map[string]string) {
	t.Helper()
	for path, contents := range files {
		writeGuardFile(t, root, path, contents)
	}
	runGuardGit(t, root, "add", ".")
	runGuardGit(t, root, "commit", "--quiet", "-m", subject)
}

func writeGuardFile(t *testing.T, root, path, contents string) {
	t.Helper()
	absolute := filepath.Join(root, filepath.FromSlash(path))
	if err := os.MkdirAll(filepath.Dir(absolute), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(absolute, []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}
}

func runGuardGit(t *testing.T, root string, arguments ...string) {
	t.Helper()
	command := exec.Command("git", arguments...)
	if root != "" {
		command.Dir = root
	}
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatal(fmt.Errorf("git %s: %w: %s", strings.Join(arguments, " "), err, strings.TrimSpace(string(output))))
	}
}

package harness

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
)

const (
	baselinePath                   = "testdata/harness/pacing-baseline.json"
	goldenPath                     = "testdata/harness/golden-seed.json"
	relevanceGoldenPath            = "testdata/harness/relevance/golden-report-v1.json"
	baselineHistoryCorrectionsPath = "kernel/baseline-history-corrections.json"
	publishedMainRef               = "refs/remotes/origin/main"
)

type baselineHistoryCorrection struct {
	OffendingCommit string `json:"offending_commit"`
	Kind            string `json:"kind"`
	Reason          string `json:"reason"`
	ReviewLog       string `json:"review_log"`
}

type baselineHistoryCorrections struct {
	SchemaVersion int                         `json:"schema_version"`
	Corrections   []baselineHistoryCorrection `json:"corrections"`
}

// ValidateBaselineCommit enforces the separate-commit review protocol for one
// non-initial baseline revision. inputsBefore contains paths changed after the
// previous baseline revision and before this artifact commit.
func ValidateBaselineCommit(commitPaths, inputsBefore []string, subject string) error {
	governed := map[string]struct{}{baselinePath: {}, goldenPath: {}, relevanceGoldenPath: {}}
	reports := map[string]struct{}{baselinePath: {}, relevanceGoldenPath: {}}
	return validateBaselineCommit(commitPaths, inputsBefore, subject, governed, reports)
}

func validateBaselineCommit(commitPaths, inputsBefore []string, subject string, governed, reports map[string]struct{}) error {
	baselineChanged := false
	for _, path := range commitPaths {
		path = strings.TrimSpace(path)
		if path == "" {
			continue
		}
		if _, report := reports[path]; report {
			baselineChanged = true
		}
		if _, allowed := governed[path]; !allowed {
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
		if strings.HasPrefix(path, "balance/catalogs/") || strings.HasPrefix(path, "balance/categories/") || strings.HasPrefix(path, "balance/routes/") || strings.HasPrefix(path, "balance/commons/") || strings.HasPrefix(path, "balance/factions/") || strings.HasPrefix(path, "balance/prestige/") || strings.HasPrefix(path, "balance/guilds/") || strings.HasPrefix(path, "balance/relevance/") || strings.HasPrefix(path, "changelog/") || strings.HasPrefix(path, "testdata/harness/scenarios/") || strings.HasPrefix(path, "testdata/harness/relevance/") || strings.HasPrefix(path, "testdata/harness/content-dynamics/") {
			return nil
		}
	}
	return fmt.Errorf("baseline rewrite has no changed balance artifact or scenario before its artifact commit")
}

// ValidateRepositoryBaselineChange validates every reachable non-initial
// baseline revision. It intentionally uses no CI-provider metadata, so local
// and hosted checks enforce the same repository history.
func ValidateRepositoryBaselineChange(root string) error {
	corrections, err := repositoryBaselineHistoryCorrections(root)
	if err != nil {
		return fmt.Errorf("baseline guard cannot load history corrections: %w", err)
	}
	relevanceReports, err := repositoryRelevanceGoldenPaths(root)
	if err != nil {
		return fmt.Errorf("baseline guard cannot load relevance registry: %w", err)
	}
	contentReports, err := repositoryContentGoldenPaths(root)
	if err != nil {
		return fmt.Errorf("baseline guard cannot load content-dynamics registry: %w", err)
	}
	governed := map[string]struct{}{baselinePath: {}, goldenPath: {}}
	reports := map[string]struct{}{baselinePath: {}}
	for _, report := range relevanceReports {
		governed[report] = struct{}{}
		reports[report] = struct{}{}
	}
	for _, report := range contentReports {
		governed[report] = struct{}{}
		reports[report] = struct{}{}
	}
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

	artifactPaths := []string{baselinePath, goldenPath}
	artifactPaths = append(artifactPaths, relevanceReports...)
	artifactPaths = append(artifactPaths, contentReports...)
	dirtyArguments := []string{"status", "--porcelain", "--untracked-files=all", "--", baselineHistoryCorrectionsPath}
	dirtyArguments = append(dirtyArguments, artifactPaths...)
	dirty, err := gitOutput(root, dirtyArguments...)
	if err != nil {
		return fmt.Errorf("baseline guard cannot inspect artifact worktree: %w", err)
	}
	if len(dirty) != 0 {
		return fmt.Errorf("baseline artifacts have uncommitted changes")
	}

	historyArguments := []string{"log", "--reverse", "--format=%H", "HEAD", "--", baselinePath}
	historyArguments = append(historyArguments, relevanceReports...)
	historyArguments = append(historyArguments, contentReports...)
	history, err := gitOutput(root, historyArguments...)
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
		if err := validateBaselineCommit(commitPaths, inputPaths, string(subject), governed, reports); err != nil {
			correction, corrected := corrections[commit]
			if !corrected || correction.Kind != "mixed_artifact_commit" || !isMixedArtifactCommit(commitPaths, governed, reports) {
				return fmt.Errorf("invalid baseline commit %s: %w", commit, err)
			}
			delete(corrections, commit)
		}
		if strings.HasPrefix(string(subject), "CONSTANTS-IDENTITY:") {
			if err := validateConstantsIdentityCommit(root, commits[index-1], parent, commit, relevanceReports); err != nil {
				return fmt.Errorf("invalid constants-identity commit %s: %w", commit, err)
			}
		}
	}
	if len(corrections) != 0 {
		return fmt.Errorf("baseline history correction targets a commit that does not change a governed report")
	}
	return nil
}

func isMixedArtifactCommit(paths []string, governed, reports map[string]struct{}) bool {
	hasReport, hasOutsidePath := false, false
	for _, path := range paths {
		path = strings.TrimSpace(path)
		if _, ok := reports[path]; ok {
			hasReport = true
		}
		if _, ok := governed[path]; !ok {
			hasOutsidePath = true
		}
	}
	return hasReport && hasOutsidePath
}

func repositoryBaselineHistoryCorrections(root string) (map[string]baselineHistoryCorrection, error) {
	history, err := gitOutput(root, "log", "--reverse", "--format=%H", "HEAD", "--", baselineHistoryCorrectionsPath)
	if err != nil {
		return nil, err
	}
	commits := strings.Fields(string(history))
	result := map[string]baselineHistoryCorrection{}
	for _, commit := range commits {
		parent, err := firstParent(root, commit)
		if err != nil {
			return nil, err
		}
		paths, err := gitLines(root, "diff-tree", "--no-commit-id", "--name-only", "-r", "-M", parent, commit)
		if err != nil {
			return nil, err
		}
		if len(paths) != 1 || paths[0] != baselineHistoryCorrectionsPath {
			return nil, fmt.Errorf("history-correction commit %s must change only %s", commit, baselineHistoryCorrectionsPath)
		}
		data, err := gitBlob(root, commit, baselineHistoryCorrectionsPath)
		if err != nil {
			return nil, err
		}
		parsed, err := decodeBaselineHistoryCorrections(data)
		if err != nil {
			return nil, fmt.Errorf("history corrections at %s: %w", commit, err)
		}
		if len(parsed) <= len(result) {
			return nil, fmt.Errorf("history-correction commit %s does not append an entry", commit)
		}
		for offending, prior := range result {
			if current, ok := parsed[offending]; !ok || current != prior {
				return nil, fmt.Errorf("history-correction commit %s changes or removes an existing entry", commit)
			}
		}
		for offending, correction := range parsed {
			if _, exists := result[offending]; exists {
				continue
			}
			if err := gitAncestor(root, offending, commit); err != nil {
				return nil, fmt.Errorf("history correction %s does not follow its offending commit: %w", offending, err)
			}
			if err := gitAncestor(root, offending, publishedMainRef); err != nil {
				return nil, fmt.Errorf("history correction %s targets an unpublished commit; unpublished commits must be repackaged, not forgiven", offending)
			}
			if _, present, err := gitBlobIfPresent(root, commit, correction.ReviewLog); err != nil || !present {
				if err != nil {
					return nil, err
				}
				return nil, fmt.Errorf("history correction %s references missing review log %s", offending, correction.ReviewLog)
			}
		}
		result = parsed
	}
	return result, nil
}

func decodeBaselineHistoryCorrections(data []byte) (map[string]baselineHistoryCorrection, error) {
	var registry baselineHistoryCorrections
	if err := decodeStrictJSON(data, &registry); err != nil {
		return nil, err
	}
	if registry.SchemaVersion != 1 || registry.Corrections == nil {
		return nil, errors.New("invalid baseline history-corrections envelope")
	}
	result := make(map[string]baselineHistoryCorrection, len(registry.Corrections))
	prior := ""
	for _, correction := range registry.Corrections {
		decoded, err := hex.DecodeString(correction.OffendingCommit)
		if err != nil || len(decoded) != 20 || hex.EncodeToString(decoded) != correction.OffendingCommit ||
			correction.OffendingCommit <= prior || correction.Kind != "mixed_artifact_commit" || strings.TrimSpace(correction.Reason) == "" ||
			!validRepositoryPath(correction.ReviewLog) || !strings.HasPrefix(correction.ReviewLog, "planning/") || !strings.HasSuffix(correction.ReviewLog, "/log.md") {
			return nil, errors.New("invalid or unsorted baseline history correction")
		}
		result[correction.OffendingCommit] = correction
		prior = correction.OffendingCommit
	}
	return result, nil
}

func gitAncestor(root, ancestor, descendant string) error {
	command := exec.Command("git", "merge-base", "--is-ancestor", ancestor, descendant)
	command.Dir = root
	if output, err := command.CombinedOutput(); err != nil {
		return fmt.Errorf("git merge-base --is-ancestor %s %s: %w: %s", ancestor, descendant, err, strings.TrimSpace(string(output)))
	}
	return nil
}

func repositoryContentGoldenPaths(root string) ([]string, error) {
	history, err := gitOutput(root, "log", "--reverse", "--format=%H", "HEAD", "--", contentDynamicsRegistryPath)
	if err != nil {
		return nil, err
	}
	commits := strings.Fields(string(history))
	seen := map[string]bool{}
	if len(commits) == 0 {
		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(contentDynamicsRegistryPath)))
		if err != nil {
			return nil, err
		}
		paths, err := registeredContentGoldenPaths(data)
		if err != nil {
			return nil, err
		}
		for _, path := range paths {
			seen[path] = true
		}
	} else {
		for _, commit := range commits {
			data, err := gitBlob(root, commit, contentDynamicsRegistryPath)
			if err != nil {
				return nil, err
			}
			paths, err := registeredContentGoldenPaths(data)
			if err != nil {
				return nil, fmt.Errorf("content-dynamics registry at %s: %w", commit, err)
			}
			for _, path := range paths {
				seen[path] = true
			}
		}
	}
	paths := make([]string, 0, len(seen))
	for path := range seen {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	return paths, nil
}

func registeredContentGoldenPaths(data []byte) ([]string, error) {
	var registry struct {
		SchemaVersion *int                               `json:"schema_version"`
		Entries       []contentDynamicsRegistryWireEntry `json:"entries"`
	}
	if err := decodeRelevanceStrict(data, &registry); err != nil || registry.SchemaVersion == nil || *registry.SchemaVersion != 1 || registry.Entries == nil {
		return nil, errors.New("invalid content-dynamics registry")
	}
	result := make([]string, 0, len(registry.Entries))
	for index, entry := range registry.Entries {
		if entry.GoldenReport == nil || !validRepositoryPath(*entry.GoldenReport) ||
			!strings.HasPrefix(*entry.GoldenReport, "testdata/harness/content-dynamics/goldens/") {
			return nil, fmt.Errorf("invalid content-dynamics golden path at entry %d", index)
		}
		result = append(result, *entry.GoldenReport)
	}
	return result, nil
}

func repositoryRelevanceGoldenPaths(root string) ([]string, error) {
	history, err := gitOutput(root, "log", "--reverse", "--format=%H", "HEAD", "--", relevanceRegistryPath)
	if err != nil {
		return nil, err
	}
	commits := strings.Fields(string(history))
	if len(commits) == 0 {
		return nil, errors.New("no committed relevance registry")
	}
	seen := map[string]bool{}
	for _, commit := range commits {
		data, err := gitBlob(root, commit, relevanceRegistryPath)
		if err != nil {
			return nil, err
		}
		paths, err := registeredRelevanceGoldenPaths(data)
		if err != nil {
			return nil, fmt.Errorf("relevance registry at %s: %w", commit, err)
		}
		for _, path := range paths {
			seen[path] = true
		}
	}
	paths := make([]string, 0, len(seen))
	for path := range seen {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	return paths, nil
}

func registeredRelevanceGoldenPaths(data []byte) ([]string, error) {
	var registry struct {
		SchemaVersion *int                         `json:"schema_version"`
		Entries       []relevanceRegistryWireEntry `json:"entries"`
	}
	if err := decodeRelevanceStrict(data, &registry); err != nil || registry.SchemaVersion == nil || *registry.SchemaVersion != 1 || registry.Entries == nil {
		return nil, errors.New("invalid relevance scenario registry")
	}
	paths := make([]string, 0, len(registry.Entries))
	seen := map[string]bool{}
	for index, entry := range registry.Entries {
		if entry.GoldenReport == nil || filepath.ToSlash(filepath.Clean(*entry.GoldenReport)) != *entry.GoldenReport || seen[*entry.GoldenReport] {
			return nil, fmt.Errorf("invalid relevance registry golden path at entry %d", index)
		}
		seen[*entry.GoldenReport] = true
		paths = append(paths, *entry.GoldenReport)
	}
	sort.Strings(paths)
	return paths, nil
}

func validateConstantsIdentityCommit(root, previousBaseline, parent, commit string, relevanceReports []string) error {
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
	if err := validateConstantsIdentityBlobs(beforeBaseline, afterBaseline, beforeGolden, afterGolden, expectedHash); err != nil {
		return err
	}
	return validateRelevanceIdentityReports(root, parent, commit, seed, expectedHash, relevanceReports)
}

func validateRelevanceIdentityReports(root, parent, commit string, seed epochSeed, expectedHash string, reports []string) error {
	activeReports, err := activeRelevanceReportsAt(root, commit, seed)
	if err != nil {
		return err
	}
	for _, reportPath := range reports {
		if reportPath == baselinePath || reportPath == goldenPath {
			continue
		}
		before, beforePresent, err := gitBlobIfPresent(root, parent, reportPath)
		if err != nil {
			return err
		}
		after, afterPresent, err := gitBlobIfPresent(root, commit, reportPath)
		if err != nil {
			return err
		}
		if beforePresent != afterPresent {
			return fmt.Errorf("constants-identity commit adds or removes relevance report %s", reportPath)
		}
		if !beforePresent {
			continue
		}
		if !activeReports[reportPath] {
			if !bytes.Equal(before, after) {
				return fmt.Errorf("constants-identity commit changes inactive relevance report %s", reportPath)
			}
			continue
		}
		var oldReport, newReport RelevanceReport
		if err := decodeStrictJSON(before, &oldReport); err != nil {
			return fmt.Errorf("decode previous relevance report %s: %w", reportPath, err)
		}
		if err := decodeStrictJSON(after, &newReport); err != nil {
			return fmt.Errorf("decode relevance report %s: %w", reportPath, err)
		}
		if newReport.ConstantsHash != expectedHash {
			return fmt.Errorf("relevance report %s hash differs from epoch manifest", reportPath)
		}
		oldReport.ConstantsHash, newReport.ConstantsHash = "", ""
		if !reflect.DeepEqual(oldReport, newReport) {
			return fmt.Errorf("constants-identity commit changes relevance behavior in %s", reportPath)
		}
	}
	return nil
}

func activeRelevanceReportsAt(root, commit string, seed epochSeed) (map[string]bool, error) {
	result := map[string]bool{}
	activeEconomy, present := epochArtifactPath(seed, "economy")
	if !present {
		return nil, errors.New("epoch seed has no economy artifact")
	}
	economyBytes, err := gitBlob(root, commit, activeEconomy)
	if err != nil {
		return nil, err
	}
	var economyEnvelope struct {
		SchemaVersion int `json:"schema_version"`
	}
	if err := json.Unmarshal(economyBytes, &economyEnvelope); err != nil {
		return nil, err
	}
	if economyEnvelope.SchemaVersion < 4 {
		return result, nil
	}
	activeRoutes, routesPresent := epochArtifactPath(seed, "routes")
	activePolicy, policyPresent := epochArtifactPath(seed, "relevance_policy")
	if !routesPresent || !policyPresent {
		return nil, errors.New("active relevance epoch is missing routes or relevance_policy")
	}
	registryBytes, err := gitBlob(root, commit, relevanceRegistryPath)
	if err != nil {
		return nil, err
	}
	var registry struct {
		SchemaVersion *int                         `json:"schema_version"`
		Entries       []relevanceRegistryWireEntry `json:"entries"`
	}
	if err := decodeRelevanceStrict(registryBytes, &registry); err != nil || registry.SchemaVersion == nil || *registry.SchemaVersion != 1 {
		return nil, errors.New("invalid relevance registry in constants-identity commit")
	}
	for _, entry := range registry.Entries {
		if entry.EconomyCatalog == nil || *entry.EconomyCatalog != activeEconomy || entry.Scenario == nil || entry.RelevancePolicy == nil || entry.GoldenReport == nil {
			continue
		}
		scenarioBytes, err := gitBlob(root, commit, *entry.Scenario)
		if err != nil {
			return nil, err
		}
		var scenario RelevanceScenario
		if err := decodeRelevanceStrict(scenarioBytes, &scenario); err != nil {
			return nil, err
		}
		if scenario.Catalog != activeEconomy || scenario.RoutesCatalog != activeRoutes || scenario.Policy != activePolicy || *entry.RelevancePolicy != activePolicy {
			return nil, errors.New("active relevance registry does not match epoch artifacts")
		}
		result[*entry.GoldenReport] = true
	}
	if len(result) != 1 {
		return nil, errors.New("active relevance epoch does not have exactly one report")
	}
	return result, nil
}

func epochArtifactPath(seed epochSeed, name string) (string, bool) {
	for _, artifact := range seed.Artifacts {
		if artifact.Name == name {
			return artifact.Path, true
		}
	}
	return "", false
}

func gitBlobIfPresent(root, commit, path string) ([]byte, bool, error) {
	command := exec.Command("git", "cat-file", "-e", commit+":"+path)
	command.Dir = root
	if output, err := command.CombinedOutput(); err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("git cat-file -e %s:%s: %w: %s", commit, path, err, strings.TrimSpace(string(output)))
	}
	data, err := gitBlob(root, commit, path)
	return data, true, err
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

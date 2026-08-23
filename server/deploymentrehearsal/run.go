package deploymentrehearsal

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"sort"
	"time"

	"cloud-clicker/server/deploymentbrowser"
	"cloud-clicker/server/operations"
	"cloud-clicker/server/releasepackage"
)

var requiredRunArtifactFiles = map[string]string{
	"candidate_manifest":  "candidate-manifest.json",
	"previous_manifest":   "previous-manifest.json",
	"release_ledger":      "release-ledger.jsonl",
	"rotation_ledger":     "rotation-ledger.jsonl",
	"backup_header":       "backup-header.json",
	"browser_result":      "browser-result.json",
	"alert_delivery":      "alert-delivery.json",
	"journal_observation": "journal-observation.json",
	"secret_scan":         "secret-scan.json",
	"supply_chain":        "supply-chain.json",
}

func LoadAndValidateRun(evidencePath, planPath, resultsDirectory, artifactsDirectory, candidateBundleDirectory string) (Evidence, error) {
	evidence, err := Load(evidencePath)
	if err != nil {
		return Evidence{}, err
	}
	planBytes, err := os.ReadFile(planPath)
	if err != nil {
		return Evidence{}, err
	}
	plan, err := LoadExecutionPlan(planPath)
	if err != nil {
		return Evidence{}, err
	}
	if err := ValidateRunBindings(evidence, plan, planBytes, resultsDirectory, artifactsDirectory, candidateBundleDirectory); err != nil {
		return Evidence{}, err
	}
	return evidence, nil
}

func ValidateRunBindings(evidence Evidence, plan ExecutionPlan, planBytes []byte, resultsDirectory, artifactsDirectory, candidateBundleDirectory string) error {
	if Validate(evidence) != nil || ValidateExecutionPlan(plan) != nil || len(planBytes) == 0 || resultsDirectory == "" || artifactsDirectory == "" || candidateBundleDirectory == "" ||
		evidence.RunID != plan.RunID || evidence.ManifestSHA256 != plan.ManifestSHA256 ||
		evidence.PreviousManifestSHA256 != plan.PreviousManifestSHA256 || artifactDigest(evidence.Artifacts, "rehearsal_plan") != hashBytes(planBytes) {
		return ErrInvalid
	}
	if err := validateToolBindings(evidence.Tools, candidateBundleDirectory); err != nil {
		return err
	}
	if err := validateRunArtifacts(evidence, artifactsDirectory); err != nil {
		return err
	}
	entries, err := os.ReadDir(resultsDirectory)
	if err != nil || len(entries) != len(plan.Checks) {
		return ErrInvalid
	}
	populations := make(map[string]Population, len(evidence.Populations))
	for _, population := range evidence.Populations {
		populations[population.Name] = population
	}
	stepCommandHashes := map[string][]string{}
	stepResultHashes := map[string][]string{}
	stepStarts := map[string]time.Time{}
	stepEnds := map[string]time.Time{}
	for index, check := range plan.Checks {
		entry := entries[index]
		if entry.Name() != check.Name+".json" || entry.Type()&os.ModeSymlink != 0 {
			return ErrInvalid
		}
		info, err := entry.Info()
		if err != nil || !info.Mode().IsRegular() || info.Mode().Perm() != 0o600 || info.Size() <= 0 || info.Size() > maximumCommandOutput {
			return ErrInvalid
		}
		data, err := os.ReadFile(filepath.Join(resultsDirectory, entry.Name()))
		if err != nil {
			return ErrInvalid
		}
		result, err := decodeCheckResult(data)
		resultHash := hashBytes(data)
		if err != nil || result.SchemaVersion != 1 || result.Name != check.Name || result.Kind != check.Kind || result.Step != check.Step ||
			result.CommandSHA256 != hashJSON(check.Command) || result.ExitCode != check.ExpectedExit || result.Result != "passed" || result.GuardExhausted ||
			result.StartedAt.Before(evidence.StartedAt) || result.CompletedAt.After(evidence.CompletedAt) || !result.CompletedAt.After(result.StartedAt) ||
			populations[check.Name].EvidenceSHA256 != resultHash {
			return ErrInvalid
		}
		stepCommandHashes[check.Step] = append(stepCommandHashes[check.Step], result.CommandSHA256)
		stepResultHashes[check.Step] = append(stepResultHashes[check.Step], resultHash)
		if stepStarts[check.Step].IsZero() || result.StartedAt.Before(stepStarts[check.Step]) {
			stepStarts[check.Step] = result.StartedAt
		}
		if result.CompletedAt.After(stepEnds[check.Step]) {
			stepEnds[check.Step] = result.CompletedAt
		}
	}
	for _, step := range evidence.Steps {
		if len(stepCommandHashes[step.Name]) == 0 || step.StartedAt != stepStarts[step.Name] || step.CompletedAt != stepEnds[step.Name] ||
			step.InputSHA256 != hashJSON(stepCommandHashes[step.Name]) || step.OutputSHA256 != hashJSON(stepResultHashes[step.Name]) {
			return ErrInvalid
		}
	}
	return nil
}

func validateToolBindings(tools []Tool, candidateBundleDirectory string) error {
	paths := map[string]string{
		"deployment-rehearsal": "deployment-rehearsal",
		"deployment-release":   "deployment-release",
		"browser-driver":       "deployment-browser",
	}
	toolHashes := make(map[string]string, len(tools))
	for _, tool := range tools {
		toolHashes[tool.Name] = tool.SHA256
	}
	for name, path := range paths {
		info, err := os.Lstat(filepath.Join(candidateBundleDirectory, path))
		if err != nil || !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 || info.Mode().Perm()&0o111 == 0 || info.Size() <= 0 {
			return ErrInvalid
		}
		data, err := os.ReadFile(filepath.Join(candidateBundleDirectory, path))
		if err != nil || toolHashes[name] != hashBytes(data) {
			return ErrInvalid
		}
	}
	return nil
}

func validateRunArtifacts(evidence Evidence, directory string) error {
	entries, err := os.ReadDir(directory)
	if err != nil || len(entries) != len(requiredRunArtifactFiles) {
		return ErrInvalid
	}
	candidateManifestBytes, err := os.ReadFile(filepath.Join(directory, requiredRunArtifactFiles["candidate_manifest"]))
	if err != nil {
		return ErrInvalid
	}
	candidateSourceCommit, err := manifestSourceCommit(candidateManifestBytes)
	if err != nil {
		return ErrInvalid
	}
	fileToName := map[string]string{}
	for name, file := range requiredRunArtifactFiles {
		fileToName[file] = name
	}
	wantFiles := make([]string, 0, len(fileToName))
	for file := range fileToName {
		wantFiles = append(wantFiles, file)
	}
	sort.Strings(wantFiles)
	for index, entry := range entries {
		if entry.Name() != wantFiles[index] || entry.Type()&os.ModeSymlink != 0 {
			return ErrInvalid
		}
		info, err := entry.Info()
		if err != nil || !info.Mode().IsRegular() || info.Mode().Perm() != 0o600 || info.Size() <= 0 || info.Size() > maximumCommandOutput {
			return ErrInvalid
		}
		data, err := os.ReadFile(filepath.Join(directory, entry.Name()))
		name := fileToName[entry.Name()]
		if err != nil || artifactDigest(evidence.Artifacts, name) != hashBytes(data) {
			return ErrInvalid
		}
		if (name == "candidate_manifest" && evidence.ManifestSHA256 != hashBytes(data)) ||
			(name == "previous_manifest" && evidence.PreviousManifestSHA256 != hashBytes(data)) {
			return ErrInvalid
		}
		if err := validateTypedRunArtifact(name, data, evidence, candidateSourceCommit); err != nil {
			return err
		}
	}
	return nil
}

func validateTypedRunArtifact(name string, data []byte, evidence Evidence, candidateSourceCommit string) error {
	switch name {
	case "browser_result":
		result, err := deploymentbrowser.DecodeResult(data)
		if err != nil || result.ManifestSHA256 != evidence.ManifestSHA256 || result.StartedAt.Before(evidence.StartedAt) || result.CompletedAt.After(evidence.CompletedAt) {
			return ErrInvalid
		}
	case "journal_observation":
		observation, err := operations.DecodeJournalObservation(data)
		if err != nil || observation.StartedAt.Before(evidence.StartedAt) || observation.CompletedAt.After(evidence.CompletedAt) {
			return ErrInvalid
		}
	case "secret_scan":
		result, err := releasepackage.DecodeSecretScanResult(data)
		if err != nil || result.ManifestSHA256 != evidence.ManifestSHA256 || result.SourceCommit != candidateSourceCommit ||
			result.StartedAt.Before(evidence.StartedAt) || result.CompletedAt.After(evidence.CompletedAt) {
			return ErrInvalid
		}
	}
	return nil
}

func manifestSourceCommit(data []byte) (string, error) {
	var manifest struct {
		SourceCommit string `json:"source_commit"`
	}
	if json.Unmarshal(data, &manifest) != nil || !commitPattern.MatchString(manifest.SourceCommit) {
		return "", ErrInvalid
	}
	return manifest.SourceCommit, nil
}

func decodeCheckResult(data []byte) (CheckResult, error) {
	if containsForbiddenEvidenceKey(data) {
		return CheckResult{}, ErrInvalid
	}
	var result CheckResult
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&result) != nil || decoder.Decode(&struct{}{}) != io.EOF {
		return CheckResult{}, ErrInvalid
	}
	return result, nil
}

func artifactDigest(artifacts []Artifact, name string) string {
	for _, artifact := range artifacts {
		if artifact.Name == name {
			return artifact.SHA256
		}
	}
	return ""
}

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
	"candidate_manifest":    "candidate-manifest.json",
	"previous_manifest":     "previous-manifest.json",
	"candidate_build":       "candidate-build.json",
	"previous_build":        "previous-build.json",
	"install_ledger":        "install-ledger.jsonl",
	"release_ledger":        "release-ledger.jsonl",
	"rotation_ledger":       "rotation-ledger.jsonl",
	"backup_header":         "backup-header.json",
	"host_observation":      "host-observation.json",
	"objective_observation": "objective-observation.json",
	"browser_result":        "browser-result.json",
	"alert_delivery":        "alert-delivery.json",
	"journal_observation":   "journal-observation.json",
	"secret_scan":           "secret-scan.json",
	"supply_chain":          "supply-chain.json",
}

func LoadAndValidateRun(evidencePath, planPath, resultsDirectory, artifactsDirectory, candidateBundleDirectory, sealDirectory string) (Evidence, error) {
	return loadAndValidateBoundRun(evidencePath, planPath, resultsDirectory, artifactsDirectory, candidateBundleDirectory, sealDirectory, false)
}

func LoadAndValidateBaseRun(evidencePath, planPath, resultsDirectory, artifactsDirectory, candidateBundleDirectory string) (Evidence, error) {
	return loadAndValidateBoundRun(evidencePath, planPath, resultsDirectory, artifactsDirectory, candidateBundleDirectory, "", true)
}

func loadAndValidateBoundRun(evidencePath, planPath, resultsDirectory, artifactsDirectory, candidateBundleDirectory, sealDirectory string, base bool) (Evidence, error) {
	evidence, err := Load(evidencePath)
	if err != nil {
		if !base {
			return Evidence{}, err
		}
		data, readErr := os.ReadFile(evidencePath)
		if readErr != nil {
			return Evidence{}, readErr
		}
		evidence, err = decodeBaseEvidence(data)
		if err != nil {
			return Evidence{}, err
		}
	}
	planBytes, err := os.ReadFile(planPath)
	if err != nil {
		return Evidence{}, err
	}
	plan, err := LoadExecutionPlan(planPath)
	if err != nil {
		return Evidence{}, err
	}
	if err := validateRunBindings(evidence, plan, planBytes, resultsDirectory, artifactsDirectory, candidateBundleDirectory, base); err != nil {
		return Evidence{}, err
	}
	if !base && ValidateFinalSeal(evidence, plan, planBytes, resultsDirectory, artifactsDirectory, candidateBundleDirectory, sealDirectory) != nil {
		return Evidence{}, ErrInvalid
	}
	return evidence, nil
}

func ValidateRunBindings(evidence Evidence, plan ExecutionPlan, planBytes []byte, resultsDirectory, artifactsDirectory, candidateBundleDirectory, sealDirectory string) error {
	if err := validateRunBindings(evidence, plan, planBytes, resultsDirectory, artifactsDirectory, candidateBundleDirectory, false); err != nil {
		return err
	}
	return ValidateFinalSeal(evidence, plan, planBytes, resultsDirectory, artifactsDirectory, candidateBundleDirectory, sealDirectory)
}

func ValidateBaseRunBindings(evidence Evidence, plan ExecutionPlan, planBytes []byte, resultsDirectory, artifactsDirectory, candidateBundleDirectory string) error {
	return validateRunBindings(evidence, plan, planBytes, resultsDirectory, artifactsDirectory, candidateBundleDirectory, true)
}

func validateRunBindings(evidence Evidence, plan ExecutionPlan, planBytes []byte, resultsDirectory, artifactsDirectory, candidateBundleDirectory string, base bool) error {
	validateEvidence := Validate
	if base {
		validateEvidence = ValidateBaseEvidence
	}
	if validateEvidence(evidence) != nil || ValidateExecutionPlan(plan) != nil || len(planBytes) == 0 || resultsDirectory == "" || artifactsDirectory == "" || candidateBundleDirectory == "" ||
		evidence.RunID != plan.RunID || evidence.ManifestSHA256 != plan.ManifestSHA256 ||
		evidence.PreviousManifestSHA256 != plan.PreviousManifestSHA256 || artifactDigest(evidence.Artifacts, "rehearsal_plan") != hashBytes(planBytes) {
		return ErrInvalid
	}
	if err := validateToolBindings(evidence.Tools, candidateBundleDirectory); err != nil {
		return err
	}
	if err := validateRunArtifacts(evidence, artifactsDirectory, candidateBundleDirectory); err != nil {
		return err
	}
	entries, err := os.ReadDir(resultsDirectory)
	if err != nil || len(entries) != len(plan.Checks) {
		return ErrInvalid
	}
	resultEntries := make(map[string]os.DirEntry, len(entries))
	for _, entry := range entries {
		resultEntries[entry.Name()] = entry
	}
	populations := make(map[string]Population, len(evidence.Populations))
	for _, population := range evidence.Populations {
		populations[population.Name] = population
	}
	stepCommandHashes := map[string][]string{}
	stepResultHashes := map[string][]string{}
	stepStarts := map[string]time.Time{}
	stepEnds := map[string]time.Time{}
	for _, check := range plan.Checks {
		entry, ok := resultEntries[check.Name+".json"]
		if !ok || entry.Type()&os.ModeSymlink != 0 {
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

func decodeBaseEvidence(data []byte) (Evidence, error) {
	if containsForbiddenEvidenceKey(data) {
		return Evidence{}, ErrInvalid
	}
	var evidence Evidence
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&evidence) != nil || decoder.Decode(&struct{}{}) != io.EOF || ValidateBaseEvidence(evidence) != nil {
		return Evidence{}, ErrInvalid
	}
	return evidence, nil
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

func validateRunArtifacts(evidence Evidence, directory, candidateBundleDirectory string) error {
	entries, err := os.ReadDir(directory)
	if err != nil || len(entries) != len(requiredRunArtifactFiles) {
		return ErrInvalid
	}
	candidateManifestBytes, err := os.ReadFile(filepath.Join(directory, requiredRunArtifactFiles["candidate_manifest"]))
	if err != nil {
		return ErrInvalid
	}
	bundledCandidateManifest, err := os.ReadFile(filepath.Join(candidateBundleDirectory, releasepackage.ReleaseManifestPath))
	if err != nil || !bytes.Equal(candidateManifestBytes, bundledCandidateManifest) {
		return ErrInvalid
	}
	candidateSourceCommit, err := manifestSourceCommit(candidateManifestBytes)
	if err != nil {
		return ErrInvalid
	}
	previousManifestBytes, err := os.ReadFile(filepath.Join(directory, requiredRunArtifactFiles["previous_manifest"]))
	if err != nil {
		return ErrInvalid
	}
	previousSourceCommit, err := manifestSourceCommit(previousManifestBytes)
	if err != nil {
		return ErrInvalid
	}
	candidateBuildBytes, err := os.ReadFile(filepath.Join(directory, requiredRunArtifactFiles["candidate_build"]))
	if err != nil {
		return ErrInvalid
	}
	candidateBuild, err := DecodeBuildRecord(candidateBuildBytes)
	if err != nil || candidateBuild.Role != "candidate" || candidateBuild.SchemaVersion != 2 || candidateBuild.SourceCommit != candidateSourceCommit || candidateBuild.ManifestSHA256 != hashBytes(candidateManifestBytes) {
		return ErrInvalid
	}
	previousBuildBytes, err := os.ReadFile(filepath.Join(directory, requiredRunArtifactFiles["previous_build"]))
	if err != nil {
		return ErrInvalid
	}
	previousBuild, err := DecodeBuildRecord(previousBuildBytes)
	if err != nil || previousBuild.Role != "previous" || previousBuild.SourceCommit != previousSourceCommit || previousBuild.ManifestSHA256 != hashBytes(previousManifestBytes) {
		return ErrInvalid
	}
	fileToName := map[string]string{}
	artifactData := make(map[string][]byte, len(requiredRunArtifactFiles))
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
		artifactData[name] = data
	}
	return validateOperatorRecords(evidence, artifactData)
}

func validateTypedRunArtifact(name string, data []byte, evidence Evidence, candidateSourceCommit string) error {
	switch name {
	case "host_observation":
		observation, err := DecodeHostObservation(data)
		if err != nil || observation.Host != evidence.Host || observation.StartedAt.Before(evidence.StartedAt) || observation.CompletedAt.After(evidence.CompletedAt) {
			return ErrInvalid
		}
	case "objective_observation":
		observation, err := DecodeObjectiveObservation(data)
		if err != nil || !sameObjectives(observation.Objectives, evidence.Objectives) || observation.StartedAt.Before(evidence.StartedAt) || observation.CompletedAt.After(evidence.CompletedAt) {
			return ErrInvalid
		}
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
	case "alert_delivery":
		observation, err := operations.DecodeAlertDeliveryObservation(data)
		if err != nil || observation.ManifestSHA256 != evidence.ManifestSHA256 || observation.StartedAt.Before(evidence.StartedAt) || observation.CompletedAt.After(evidence.CompletedAt) {
			return ErrInvalid
		}
	case "secret_scan":
		result, err := releasepackage.DecodeSecretScanResult(data)
		if err != nil || result.ManifestSHA256 != evidence.ManifestSHA256 || result.SourceCommit != candidateSourceCommit ||
			result.StartedAt.Before(evidence.StartedAt) || result.CompletedAt.After(evidence.CompletedAt) {
			return ErrInvalid
		}
	case "supply_chain":
		result, err := DecodeSupplyChainResult(data)
		if err != nil || result.CandidateManifestSHA256 != evidence.ManifestSHA256 || result.PreviousManifestSHA256 != evidence.PreviousManifestSHA256 ||
			result.CandidateBuildSHA256 != artifactDigest(evidence.Artifacts, "candidate_build") ||
			result.PreviousBuildSHA256 != artifactDigest(evidence.Artifacts, "previous_build") ||
			result.SecretScanSHA256 != artifactDigest(evidence.Artifacts, "secret_scan") ||
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

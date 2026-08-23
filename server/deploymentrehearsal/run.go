package deploymentrehearsal

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
)

func LoadAndValidateRun(evidencePath, planPath, resultsDirectory string) (Evidence, error) {
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
	if err := ValidateRunBindings(evidence, plan, planBytes, resultsDirectory); err != nil {
		return Evidence{}, err
	}
	return evidence, nil
}

func ValidateRunBindings(evidence Evidence, plan ExecutionPlan, planBytes []byte, resultsDirectory string) error {
	if Validate(evidence) != nil || ValidateExecutionPlan(plan) != nil || len(planBytes) == 0 || resultsDirectory == "" ||
		evidence.RunID != plan.RunID || evidence.ManifestSHA256 != plan.ManifestSHA256 ||
		evidence.PreviousManifestSHA256 != plan.PreviousManifestSHA256 || artifactDigest(evidence.Artifacts, "rehearsal_plan") != hashBytes(planBytes) {
		return ErrInvalid
	}
	entries, err := os.ReadDir(resultsDirectory)
	if err != nil || len(entries) != len(plan.Checks) {
		return ErrInvalid
	}
	populations := make(map[string]Population, len(evidence.Populations))
	for _, population := range evidence.Populations {
		populations[population.Name] = population
	}
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
		if err != nil || result.SchemaVersion != 1 || result.Name != check.Name || result.Kind != check.Kind || result.Step != check.Step ||
			result.CommandSHA256 != hashJSON(check.Command) || result.ExitCode != check.ExpectedExit || result.Result != "passed" || result.GuardExhausted ||
			result.StartedAt.Before(evidence.StartedAt) || result.CompletedAt.After(evidence.CompletedAt) || !result.CompletedAt.After(result.StartedAt) ||
			populations[check.Name].EvidenceSHA256 != hashBytes(data) {
			return ErrInvalid
		}
	}
	return nil
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

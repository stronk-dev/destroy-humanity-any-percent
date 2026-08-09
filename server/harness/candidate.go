package harness

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"cloud-clicker/server/replaycatalog"
)

const candidateManifestStatus = "ratified"

type CandidateIdentity struct {
	ManifestPath  string
	ArtifactNames []string
	ConstantsHash string
}

type CandidatePacingDelta struct {
	PolicyID         string `json:"policy_id"`
	Milestone        string `json:"milestone_id"`
	Statistic        string `json:"statistic"`
	BaselineMS       int64  `json:"baseline_ms"`
	CandidateMS      int64  `json:"candidate_ms"`
	DeltaMS          int64  `json:"delta_ms"`
	RelativeDeltaPPM *int64 `json:"relative_delta_ppm"`
}

type CandidatePacingReport struct {
	SchemaVersion          int                    `json:"schema_version"`
	ManifestPath           string                 `json:"manifest_path"`
	ArtifactCount          int                    `json:"artifact_count"`
	ArtifactNames          []string               `json:"artifact_names"`
	CandidateConstantsHash string                 `json:"candidate_constants_hash"`
	BaselineConstantsHash  string                 `json:"baseline_constants_hash"`
	ScenarioID             string                 `json:"scenario_id"`
	ScenarioHash           string                 `json:"scenario_hash"`
	RunCount               int                    `json:"run_count"`
	Values                 []CandidatePacingDelta `json:"values"`
	PacingWarnings         []string               `json:"pacing_warnings"`
	PacingFindings         []string               `json:"pacing_findings"`
	InvariantFailures      []string               `json:"invariant_failures"`
}

type candidateManifest struct {
	SchemaVersion int                    `json:"schema_version"`
	Status        string                 `json:"status"`
	ConstantsHash string                 `json:"constants_hash"`
	Artifacts     []candidateManifestRow `json:"artifacts"`
}

type candidateManifestRow struct {
	Name            string `json:"name"`
	SchemaVersion   int    `json:"schema_version"`
	SourcePath      string `json:"source_path"`
	ProductionPath  string `json:"production_path"`
	SHA256          string `json:"sha256"`
	ContentGate     string `json:"content_gate"`
	ConsumedVerdict string `json:"consumed_verdict"`
}

func LoadCandidateSuite(repositoryRoot, scenarioPath, manifestPath string) (*Suite, CandidateIdentity, error) {
	scenario, scenarioBytes, err := loadScenario(repositoryRoot, scenarioPath)
	if err != nil {
		return nil, CandidateIdentity{}, err
	}
	manifest, artifacts, productionPaths, err := loadCandidateManifest(repositoryRoot, manifestPath)
	if err != nil {
		return nil, CandidateIdentity{}, err
	}
	for name, scenarioPath := range map[string]string{"commons": scenario.CommonsCatalog, "economy": scenario.Catalog, "routes": scenario.RoutesCatalog} {
		if productionPaths[name] != scenarioPath {
			return nil, CandidateIdentity{}, fmt.Errorf("scenario %s path %q differs from candidate production path %q", name, scenarioPath, productionPaths[name])
		}
	}
	if _, err := replaycatalog.Load(manifest.ConstantsHash, artifacts); err != nil {
		return nil, CandidateIdentity{}, fmt.Errorf("candidate replay bundle: %w", err)
	}
	suite, err := newSuite(scenario, scenarioBytes, artifacts["economy"], artifacts["routes"], artifacts["commons"], manifest.ConstantsHash)
	if err != nil {
		return nil, CandidateIdentity{}, err
	}
	names := make([]string, 0, len(manifest.Artifacts))
	for _, row := range manifest.Artifacts {
		names = append(names, row.Name)
	}
	return suite, CandidateIdentity{ManifestPath: manifestPath, ArtifactNames: names, ConstantsHash: manifest.ConstantsHash}, nil
}

func BuildCandidatePacingReport(identity CandidateIdentity, current, baseline AggregateReport) CandidatePacingReport {
	warnings := make([]string, 0)
	findings := make([]string, 0)
	baselineWarnings, baselineFindings := CompareBaseline(current, baseline)
	warnings = append(warnings, baselineWarnings...)
	findings = append(findings, baselineFindings...)
	invariants := make([]string, 0)
	for _, failure := range current.Failures {
		if strings.HasPrefix(failure, "envelope ") {
			findings = append(findings, failure)
		} else {
			invariants = append(invariants, failure)
		}
	}
	warnings = append(warnings, current.Warnings...)
	sort.Strings(warnings)
	sort.Strings(findings)
	sort.Strings(invariants)
	baselineByKey := make(map[string]int64, len(baseline.Values))
	for _, value := range baseline.Values {
		baselineByKey[pacingKey(value)] = value.ValueMS
	}
	deltas := make([]CandidatePacingDelta, 0, len(current.Values))
	for _, value := range current.Values {
		prior := baselineByKey[pacingKey(value)]
		delta := value.ValueMS - prior
		var relative *int64
		if prior != 0 {
			ppm := delta * 1_000_000 / prior
			relative = &ppm
		} else if delta == 0 {
			ppm := int64(0)
			relative = &ppm
		}
		deltas = append(deltas, CandidatePacingDelta{PolicyID: value.PolicyID, Milestone: value.Milestone,
			Statistic: value.Statistic, BaselineMS: prior, CandidateMS: value.ValueMS, DeltaMS: delta, RelativeDeltaPPM: relative})
	}
	return CandidatePacingReport{SchemaVersion: 1, ManifestPath: identity.ManifestPath, ArtifactCount: len(identity.ArtifactNames),
		ArtifactNames: append([]string(nil), identity.ArtifactNames...), CandidateConstantsHash: identity.ConstantsHash,
		BaselineConstantsHash: baseline.ConstantsHash, ScenarioID: current.ScenarioID, ScenarioHash: current.ScenarioHash,
		RunCount: current.RunCount, Values: deltas, PacingWarnings: warnings, PacingFindings: findings, InvariantFailures: invariants}
}

func loadCandidateManifest(repositoryRoot, manifestPath string) (candidateManifest, map[string][]byte, map[string]string, error) {
	data, err := os.ReadFile(filepath.Join(repositoryRoot, filepath.FromSlash(manifestPath)))
	if err != nil {
		return candidateManifest{}, nil, nil, err
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var manifest candidateManifest
	if err := decoder.Decode(&manifest); err != nil {
		return candidateManifest{}, nil, nil, err
	}
	var trailing any
	if !errors.Is(decoder.Decode(&trailing), io.EOF) || manifest.SchemaVersion != 1 || manifest.Status != candidateManifestStatus ||
		len(manifest.Artifacts) != 16 || manifest.ConstantsHash == "" {
		return candidateManifest{}, nil, nil, errors.New("invalid ratified candidate manifest")
	}
	artifacts := make(map[string][]byte, len(manifest.Artifacts))
	productionPaths := make(map[string]string, len(manifest.Artifacts))
	seenProduction := make(map[string]bool, len(manifest.Artifacts))
	prior := ""
	for _, row := range manifest.Artifacts {
		if row.Name <= prior || row.SchemaVersion < 1 || row.SourcePath == "" || row.ProductionPath == "" || row.ContentGate == "" ||
			row.ConsumedVerdict == "" || strings.HasPrefix(strings.ToLower(row.ConsumedVerdict), "pending") || seenProduction[row.ProductionPath] {
			return candidateManifest{}, nil, nil, fmt.Errorf("invalid candidate manifest row %q", row.Name)
		}
		artifact, err := os.ReadFile(filepath.Join(repositoryRoot, filepath.FromSlash(row.SourcePath)))
		if err != nil {
			return candidateManifest{}, nil, nil, err
		}
		digest := sha256.Sum256(artifact)
		if hex.EncodeToString(digest[:]) != row.SHA256 {
			return candidateManifest{}, nil, nil, fmt.Errorf("candidate artifact %q SHA-256 mismatch", row.Name)
		}
		var envelope struct {
			SchemaVersion int `json:"schema_version"`
		}
		if err := json.Unmarshal(artifact, &envelope); err != nil || envelope.SchemaVersion != row.SchemaVersion {
			return candidateManifest{}, nil, nil, fmt.Errorf("candidate artifact %q schema version does not match manifest", row.Name)
		}
		artifacts[row.Name] = artifact
		productionPaths[row.Name] = row.ProductionPath
		seenProduction[row.ProductionPath] = true
		prior = row.Name
	}
	return manifest, artifacts, productionPaths, nil
}

func pacingKey(value AggregateValue) string {
	return value.PolicyID + "\x00" + value.Milestone + "\x00" + value.Statistic
}

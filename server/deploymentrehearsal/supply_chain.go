package deploymentrehearsal

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"slices"
	"time"

	"cloud-clicker/server/releasepackage"
)

type SupplyChainResult struct {
	SchemaVersion            int       `json:"schema_version"`
	CandidateManifestSHA256  string    `json:"candidate_manifest_sha256"`
	PreviousManifestSHA256   string    `json:"previous_manifest_sha256"`
	CandidateBuildSHA256     string    `json:"candidate_build_sha256"`
	PreviousBuildSHA256      string    `json:"previous_build_sha256"`
	SecretScanSHA256         string    `json:"secret_scan_sha256"`
	StartedAt                time.Time `json:"started_at"`
	CompletedAt              time.Time `json:"completed_at"`
	ProductionImages         int       `json:"production_images"`
	RehearsalImages          int       `json:"rehearsal_images"`
	SBOMDocuments            int       `json:"sbom_documents"`
	RootAttributionPresent   bool      `json:"root_attribution_present"`
	SiteAttributionPresent   bool      `json:"site_attribution_present"`
	SourceAndImageProvenance bool      `json:"source_and_image_provenance"`
	ObjectiveCompleted       bool      `json:"objective_completed"`
	GuardExhausted           bool      `json:"guard_exhausted"`
}

type SupplyChainRequest struct {
	CandidateBundle string
	PreviousBundle  string
	CandidateBuild  string
	PreviousBuild   string
	SecretScan      string
	Output          string
	Now             func() time.Time
}

var validateReleaseBundle = releasepackage.ValidateBundle

func ObserveSupplyChain(request SupplyChainRequest) (SupplyChainResult, error) {
	for _, path := range []string{request.CandidateBundle, request.PreviousBundle, request.CandidateBuild, request.PreviousBuild, request.SecretScan, request.Output} {
		if !filepath.IsAbs(path) || filepath.Clean(path) != path {
			return SupplyChainResult{}, ErrInvalid
		}
	}
	if request.Now == nil || validateReleaseBundle(request.CandidateBundle) != nil || validateReleaseBundle(request.PreviousBundle) != nil {
		return SupplyChainResult{}, ErrInvalid
	}
	started := request.Now().UTC()
	candidateManifest, err := os.ReadFile(filepath.Join(request.CandidateBundle, releasepackage.ReleaseManifestPath))
	if err != nil {
		return SupplyChainResult{}, ErrInvalid
	}
	previousManifest, err := os.ReadFile(filepath.Join(request.PreviousBundle, releasepackage.ReleaseManifestPath))
	if err != nil {
		return SupplyChainResult{}, ErrInvalid
	}
	candidateBuildBytes, candidateBuild, err := loadBuildBytes(request.CandidateBuild)
	if err != nil {
		return SupplyChainResult{}, err
	}
	previousBuildBytes, previousBuild, err := loadBuildBytes(request.PreviousBuild)
	if err != nil {
		return SupplyChainResult{}, err
	}
	secretBytes, err := os.ReadFile(request.SecretScan)
	if err != nil {
		return SupplyChainResult{}, ErrInvalid
	}
	secret, err := releasepackage.DecodeSecretScanResult(secretBytes)
	if err != nil {
		return SupplyChainResult{}, ErrInvalid
	}
	candidateIdentity, err := decodeSupplyManifest(candidateManifest)
	if err != nil {
		return SupplyChainResult{}, err
	}
	previousIdentity, err := decodeSupplyManifest(previousManifest)
	if err != nil || candidateBuild.Role != "candidate" || previousBuild.Role != "previous" || candidateBuild.SchemaVersion != 2 ||
		candidateBuild.ManifestSHA256 != hashBytes(candidateManifest) || previousBuild.ManifestSHA256 != hashBytes(previousManifest) ||
		candidateBuild.SourceCommit != candidateIdentity.SourceCommit || previousBuild.SourceCommit != previousIdentity.SourceCommit ||
		secret.ManifestSHA256 != hashBytes(candidateManifest) || secret.SourceCommit != candidateIdentity.SourceCommit ||
		len(candidateIdentity.Images) != 6 || len(candidateIdentity.RehearsalImages) != 1 ||
		!slices.Contains(candidateIdentity.Artifacts, "LICENSE") || !slices.Contains(candidateIdentity.Artifacts, "third-party-licenses.txt") ||
		!slices.Contains(candidateIdentity.Artifacts, "site/third-party-licenses.txt") {
		return SupplyChainResult{}, ErrInvalid
	}
	result := SupplyChainResult{SchemaVersion: 1, CandidateManifestSHA256: hashBytes(candidateManifest), PreviousManifestSHA256: hashBytes(previousManifest),
		CandidateBuildSHA256: hashBytes(candidateBuildBytes), PreviousBuildSHA256: hashBytes(previousBuildBytes), SecretScanSHA256: hashBytes(secretBytes),
		StartedAt: started, CompletedAt: request.Now().UTC(), ProductionImages: len(candidateIdentity.Images),
		RehearsalImages: len(candidateIdentity.RehearsalImages), SBOMDocuments: 1 + len(candidateIdentity.Images) + len(candidateIdentity.RehearsalImages),
		RootAttributionPresent: true, SiteAttributionPresent: true, SourceAndImageProvenance: true, ObjectiveCompleted: true}
	if ValidateSupplyChainResult(result) != nil {
		return SupplyChainResult{}, ErrInvalid
	}
	data, err := json.Marshal(result)
	if err != nil {
		return SupplyChainResult{}, err
	}
	if err := writeExclusiveResult(request.Output, append(data, '\n')); err != nil {
		return SupplyChainResult{}, err
	}
	return result, nil
}

func ValidateSupplyChainResult(result SupplyChainResult) error {
	if result.SchemaVersion != 1 || !hashPattern.MatchString(result.CandidateManifestSHA256) || !hashPattern.MatchString(result.PreviousManifestSHA256) ||
		result.CandidateManifestSHA256 == result.PreviousManifestSHA256 || !hashPattern.MatchString(result.CandidateBuildSHA256) ||
		!hashPattern.MatchString(result.PreviousBuildSHA256) || !hashPattern.MatchString(result.SecretScanSHA256) || result.StartedAt.IsZero() ||
		!result.CompletedAt.After(result.StartedAt) || result.ProductionImages != 6 || result.RehearsalImages != 1 || result.SBOMDocuments != 8 ||
		!result.RootAttributionPresent || !result.SiteAttributionPresent || !result.SourceAndImageProvenance || !result.ObjectiveCompleted || result.GuardExhausted {
		return ErrInvalid
	}
	return nil
}

func DecodeSupplyChainResult(data []byte) (SupplyChainResult, error) {
	var result SupplyChainResult
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&result) != nil || decoder.Decode(&struct{}{}) != io.EOF || ValidateSupplyChainResult(result) != nil {
		return SupplyChainResult{}, ErrInvalid
	}
	return result, nil
}

type supplyManifest struct {
	SourceCommit    string `json:"source_commit"`
	Images          []any  `json:"images"`
	RehearsalImages []any  `json:"rehearsal_images"`
	Artifacts       []struct {
		Path string `json:"path"`
	} `json:"artifacts"`
}

type supplyManifestIdentity struct {
	SourceCommit    string
	Images          []any
	RehearsalImages []any
	Artifacts       []string
}

func decodeSupplyManifest(data []byte) (supplyManifestIdentity, error) {
	var manifest supplyManifest
	if json.Unmarshal(data, &manifest) != nil || !commitPattern.MatchString(manifest.SourceCommit) {
		return supplyManifestIdentity{}, ErrInvalid
	}
	paths := make([]string, len(manifest.Artifacts))
	for index, artifact := range manifest.Artifacts {
		paths[index] = artifact.Path
	}
	return supplyManifestIdentity{SourceCommit: manifest.SourceCommit, Images: manifest.Images,
		RehearsalImages: manifest.RehearsalImages, Artifacts: paths}, nil
}

func loadBuildBytes(path string) ([]byte, BuildRecord, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, BuildRecord{}, ErrInvalid
	}
	record, err := DecodeBuildRecord(data)
	return data, record, err
}

func writeExclusiveResult(path string, data []byte) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return err
	}
	written := false
	defer func() {
		_ = file.Close()
		if !written {
			_ = os.Remove(path)
		}
	}()
	if _, err = file.Write(data); err == nil {
		err = file.Sync()
	}
	if err != nil {
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	written = true
	return nil
}

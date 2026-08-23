package deploymentrehearsal

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"time"
)

const ForgeryMutation = "browser_result_hash"

type ForgeryProof struct {
	SchemaVersion      int       `json:"schema_version"`
	RunID              string    `json:"run_id"`
	ManifestSHA256     string    `json:"manifest_sha256"`
	BaseEvidenceSHA256 string    `json:"base_evidence_sha256"`
	Mutation           string    `json:"mutation"`
	StartedAt          time.Time `json:"started_at"`
	CompletedAt        time.Time `json:"completed_at"`
	RejectionObserved  bool      `json:"rejection_observed"`
	ObjectiveCompleted bool      `json:"objective_completed"`
	GuardExhausted     bool      `json:"guard_exhausted"`
}

type ForgeryRequest struct {
	BaseEvidence    string
	Plan            string
	Results         string
	Artifacts       string
	CandidateBundle string
	Output          string
	Now             func() time.Time
}

func ProduceForgeryProof(request ForgeryRequest) (ForgeryProof, error) {
	for _, path := range []string{request.BaseEvidence, request.Plan, request.Results, request.Artifacts, request.CandidateBundle, request.Output} {
		if !filepath.IsAbs(path) || filepath.Clean(path) != path {
			return ForgeryProof{}, ErrInvalid
		}
	}
	if request.Now == nil {
		return ForgeryProof{}, ErrInvalid
	}
	baseBytes, err := os.ReadFile(request.BaseEvidence)
	if err != nil {
		return ForgeryProof{}, err
	}
	base, err := decodeBaseEvidence(baseBytes)
	if err != nil {
		return ForgeryProof{}, err
	}
	planBytes, err := os.ReadFile(request.Plan)
	if err != nil {
		return ForgeryProof{}, err
	}
	plan, err := LoadExecutionPlan(request.Plan)
	if err != nil || ValidateBaseRunBindings(base, plan, planBytes, request.Results, request.Artifacts, request.CandidateBundle) != nil {
		return ForgeryProof{}, ErrInvalid
	}
	started := request.Now().UTC()
	forged := forgeBaseEvidence(base)
	if ValidateBaseRunBindings(forged, plan, planBytes, request.Results, request.Artifacts, request.CandidateBundle) == nil {
		return ForgeryProof{}, ErrInvalid
	}
	proof := ForgeryProof{SchemaVersion: 1, RunID: base.RunID, ManifestSHA256: base.ManifestSHA256,
		BaseEvidenceSHA256: hashBytes(baseBytes), Mutation: ForgeryMutation, StartedAt: started,
		CompletedAt: request.Now().UTC(), RejectionObserved: true, ObjectiveCompleted: true}
	if ValidateForgeryProof(proof) != nil {
		return ForgeryProof{}, ErrInvalid
	}
	data, err := json.Marshal(proof)
	if err != nil || writeNewEvidence(request.Output, append(data, '\n')) != nil {
		return ForgeryProof{}, ErrInvalid
	}
	return proof, nil
}

func ValidateForgeryProof(proof ForgeryProof) error {
	if proof.SchemaVersion != 1 || !identifierPattern.MatchString(proof.RunID) || !hashPattern.MatchString(proof.ManifestSHA256) ||
		!hashPattern.MatchString(proof.BaseEvidenceSHA256) || proof.Mutation != ForgeryMutation || proof.StartedAt.IsZero() ||
		!proof.CompletedAt.After(proof.StartedAt) || !proof.RejectionObserved || !proof.ObjectiveCompleted || proof.GuardExhausted {
		return ErrInvalid
	}
	return nil
}

func DecodeForgeryProof(data []byte) (ForgeryProof, error) {
	var proof ForgeryProof
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&proof) != nil || decoder.Decode(&struct{}{}) != io.EOF || ValidateForgeryProof(proof) != nil {
		return ForgeryProof{}, ErrInvalid
	}
	return proof, nil
}

func ValidateFinalSeal(evidence Evidence, plan ExecutionPlan, planBytes []byte, resultsDirectory, artifactsDirectory, candidateBundleDirectory, sealDirectory string) error {
	if sealDirectory == "" {
		return ErrInvalid
	}
	files := []string{"base-evidence.json", "forgery-proof.json"}
	entries, err := os.ReadDir(sealDirectory)
	if err != nil || len(entries) != len(files) {
		return ErrInvalid
	}
	sort.Strings(files)
	data := map[string][]byte{}
	for index, entry := range entries {
		if entry.Name() != files[index] || entry.Type()&os.ModeSymlink != 0 {
			return ErrInvalid
		}
		info, err := entry.Info()
		if err != nil || !info.Mode().IsRegular() || info.Mode().Perm() != 0o600 || info.Size() <= 0 || info.Size() > maximumCommandOutput {
			return ErrInvalid
		}
		data[entry.Name()], err = os.ReadFile(filepath.Join(sealDirectory, entry.Name()))
		if err != nil {
			return ErrInvalid
		}
	}
	baseBytes, proofBytes := data["base-evidence.json"], data["forgery-proof.json"]
	if artifactDigest(evidence.Artifacts, "base_evidence") != hashBytes(baseBytes) ||
		artifactDigest(evidence.Artifacts, "forgery_proof") != hashBytes(proofBytes) || populationDigest(evidence.Populations, "forged_successful_evidence") != hashBytes(proofBytes) {
		return ErrInvalid
	}
	base, err := decodeBaseEvidence(baseBytes)
	if err != nil || !finalExtendsBase(evidence, base) || ValidateBaseRunBindings(base, plan, planBytes, resultsDirectory, artifactsDirectory, candidateBundleDirectory) != nil {
		return ErrInvalid
	}
	proof, err := DecodeForgeryProof(proofBytes)
	if err != nil || proof.RunID != base.RunID || proof.ManifestSHA256 != base.ManifestSHA256 || proof.BaseEvidenceSHA256 != hashBytes(baseBytes) ||
		proof.StartedAt.Before(base.CompletedAt) || proof.CompletedAt.After(evidence.CompletedAt) {
		return ErrInvalid
	}
	if ValidateBaseRunBindings(forgeBaseEvidence(base), plan, planBytes, resultsDirectory, artifactsDirectory, candidateBundleDirectory) == nil {
		return ErrInvalid
	}
	return nil
}

func forgeBaseEvidence(base Evidence) Evidence {
	data, _ := json.Marshal(base)
	var forged Evidence
	_ = json.Unmarshal(data, &forged)
	for index := range forged.Artifacts {
		if forged.Artifacts[index].Name == "browser_result" {
			forged.Artifacts[index].SHA256 = hashForMutation(forged.Artifacts[index].SHA256)
			break
		}
	}
	return forged
}

func hashForMutation(current string) string {
	value := "sha256:" + string(bytes.Repeat([]byte{'f'}, 64))
	if current == value {
		return "sha256:" + string(bytes.Repeat([]byte{'e'}, 64))
	}
	return value
}

func finalExtendsBase(final, base Evidence) bool {
	stripped := final
	stripped.Populations = removeNamedPopulation(final.Populations, "forged_successful_evidence")
	stripped.Artifacts = filterNamedArtifacts(final.Artifacts, BaseRequiredArtifacts)
	stripped.CompletedAt = base.CompletedAt
	return reflect.DeepEqual(stripped, base)
}

func removeNamedPopulation(populations []Population, name string) []Population {
	result := make([]Population, 0, len(populations)-1)
	for _, population := range populations {
		if population.Name != name {
			result = append(result, population)
		}
	}
	return result
}

func filterNamedArtifacts(artifacts []Artifact, names []string) []Artifact {
	result := make([]Artifact, 0, len(names))
	for _, artifact := range artifacts {
		if slicesContains(names, artifact.Name) {
			result = append(result, artifact)
		}
	}
	return result
}

func populationDigest(populations []Population, name string) string {
	for _, population := range populations {
		if population.Name == name {
			return population.EvidenceSHA256
		}
	}
	return ""
}

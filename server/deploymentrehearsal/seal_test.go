package deploymentrehearsal

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestProduceForgeryProofRequiresValidBaseThenObservedRejection(t *testing.T) {
	fixture := boundRunFixture(t)
	basePath := filepath.Join(fixture.sealDirectory, "base-evidence.json")
	baseBytes, err := os.ReadFile(basePath)
	if err != nil {
		t.Fatal(err)
	}
	base, err := decodeBaseEvidence(baseBytes)
	if err != nil {
		t.Fatal(err)
	}
	clock := base.CompletedAt
	now := func() time.Time {
		clock = clock.Add(time.Second)
		return clock
	}
	output := filepath.Join(t.TempDir(), "forgery-proof.json")
	proof, err := ProduceForgeryProof(ForgeryRequest{BaseEvidence: basePath, Plan: fixture.planPath,
		Results: fixture.resultsDirectory, Artifacts: fixture.artifactsDirectory, CandidateBundle: fixture.candidateBundleDirectory,
		Output: output, Now: now})
	if err != nil || proof.BaseEvidenceSHA256 != hashBytes(baseBytes) || !proof.RejectionObserved {
		t.Fatalf("proof=%+v err=%v", proof, err)
	}
	data, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := DecodeForgeryProof(data); err != nil {
		t.Fatal(err)
	}

	base.Artifacts[0].SHA256 = hashForBuild("f")
	invalidBytes, _ := json.Marshal(base)
	invalidPath := filepath.Join(t.TempDir(), "invalid-base.json")
	if err := os.WriteFile(invalidPath, append(invalidBytes, '\n'), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := ProduceForgeryProof(ForgeryRequest{BaseEvidence: invalidPath, Plan: fixture.planPath,
		Results: fixture.resultsDirectory, Artifacts: fixture.artifactsDirectory, CandidateBundle: fixture.candidateBundleDirectory,
		Output: filepath.Join(t.TempDir(), "invalid-proof.json"), Now: now}); !errors.Is(err, ErrInvalid) {
		t.Fatalf("pre-invalid base accepted: %v", err)
	}
}

func TestFinalSealRejectsClaimOnlyOrChangedProof(t *testing.T) {
	fixture := boundRunFixture(t)
	proofPath := filepath.Join(fixture.sealDirectory, "forgery-proof.json")
	data, err := os.ReadFile(proofPath)
	if err != nil {
		t.Fatal(err)
	}
	var proof ForgeryProof
	if json.Unmarshal(data, &proof) != nil {
		t.Fatal("decode proof")
	}
	proof.Mutation = "claimed_rejection"
	changed, _ := json.Marshal(proof)
	changed = append(changed, '\n')
	if err := os.WriteFile(proofPath, changed, 0o600); err != nil {
		t.Fatal(err)
	}
	setArtifactDigest(&fixture.evidence, "forgery_proof", hashBytes(changed))
	setPopulationDigest(&fixture.evidence, "forged_successful_evidence", hashBytes(changed))
	writeEvidenceFixture(t, fixture.evidencePath, fixture.evidence)
	if _, err := LoadAndValidateRun(fixture.evidencePath, fixture.planPath, fixture.resultsDirectory,
		fixture.artifactsDirectory, fixture.candidateBundleDirectory, fixture.sealDirectory); !errors.Is(err, ErrInvalid) {
		t.Fatalf("claim-only forgery proof accepted: %v", err)
	}
}

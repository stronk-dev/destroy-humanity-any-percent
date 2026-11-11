package deploymentrehearsal

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"cloud-clicker/server/releasepackage"
)

func TestRehearsalReadersRequireExplicitGuardObservations(t *testing.T) {
	evidence := validEvidence()
	base := validEvidence()
	base.Populations = removeNamedPopulation(base.Populations, "forged_successful_evidence")
	base.Artifacts = filterNamedArtifacts(base.Artifacts, BaseRequiredArtifacts)
	host := HostObservation{SchemaVersion: 1, StartedAt: evidence.StartedAt,
		CompletedAt: evidence.CompletedAt, Host: evidence.Host, ObjectiveCompleted: true}
	objective := ObjectiveObservation{SchemaVersion: 1, StartedAt: evidence.StartedAt,
		CompletedAt: evidence.CompletedAt, Objectives: evidence.Objectives, ObjectiveCompleted: true}
	proof := ForgeryProof{SchemaVersion: 1, RunID: evidence.RunID, ManifestSHA256: evidence.ManifestSHA256,
		BaseEvidenceSHA256: evidence.PreviousManifestSHA256, Mutation: ForgeryMutation,
		StartedAt: evidence.StartedAt, CompletedAt: evidence.CompletedAt,
		RejectionObserved: true, ObjectiveCompleted: true}
	supply := SupplyChainResult{SchemaVersion: 1, CandidateManifestSHA256: evidence.ManifestSHA256,
		PreviousManifestSHA256: evidence.PreviousManifestSHA256, CandidateBuildSHA256: hashForBuild("a"),
		PreviousBuildSHA256: hashForBuild("b"), SecretScanSHA256: hashForBuild("c"),
		StartedAt: evidence.StartedAt, CompletedAt: evidence.CompletedAt,
		ProductionImages: 6, RehearsalImages: 1, SBOMDocuments: 8,
		RootAttributionPresent: true, SiteAttributionPresent: true,
		SourceAndImageProvenance: true, ObjectiveCompleted: true}
	scan := validSecretScanResultForSupply([]byte("bound manifest"), "cccccccccccccccccccccccccccccccccccccccc", evidence.StartedAt)
	readers := []struct {
		name   string
		value  any
		decode func([]byte) error
	}{
		{"final evidence", evidence, func(data []byte) error { _, err := Decode(data); return err }},
		{"base evidence", base, func(data []byte) error { _, err := decodeBaseEvidence(data); return err }},
		{"build", validBuildRecord(), func(data []byte) error { _, err := DecodeBuildRecord(data); return err }},
		{"host", host, func(data []byte) error { _, err := DecodeHostObservation(data); return err }},
		{"objectives", objective, func(data []byte) error { _, err := DecodeObjectiveObservation(data); return err }},
		{"forgery proof", proof, func(data []byte) error { _, err := DecodeForgeryProof(data); return err }},
		{"supply chain", supply, func(data []byte) error { _, err := DecodeSupplyChainResult(data); return err }},
		{"secret scan", scan, func(data []byte) error { _, err := releasepackage.DecodeSecretScanResult(data); return err }},
	}
	for _, reader := range readers {
		t.Run(reader.name, func(t *testing.T) {
			data, err := json.Marshal(reader.value)
			if err != nil {
				t.Fatal(err)
			}
			if err := reader.decode(data); err != nil {
				t.Fatalf("complete control rejected: %v", err)
			}
			for name, mutate := range map[string]func([]byte) []byte{
				"missing": func(data []byte) []byte { return bytes.Replace(data, []byte(`,"guard_exhausted":false`), nil, 1) },
				"null": func(data []byte) []byte {
					return bytes.Replace(data, []byte(`"guard_exhausted":false`), []byte(`"guard_exhausted":null`), 1)
				},
				"case alias": func(data []byte) []byte {
					return bytes.Replace(data, []byte(`"guard_exhausted"`), []byte(`"GUARD_EXHAUSTED"`), 1)
				},
				"overwritten failure": func(data []byte) []byte {
					return append([]byte(`{"guard_exhausted":true,`), data[1:]...)
				},
			} {
				t.Run(name, func(t *testing.T) {
					mutated := mutate(data)
					if bytes.Equal(data, mutated) {
						t.Fatal("mutation did not reach its subject")
					}
					if err := reader.decode(mutated); err == nil {
						t.Fatal("unrecorded or contradictory guard observation accepted")
					}
				})
			}
		})
	}
}

func TestRehearsalReadersRequireNestedZeroObservations(t *testing.T) {
	for _, name := range []string{"rpo_seconds", "rto_seconds"} {
		t.Run(name, func(t *testing.T) {
			evidence := validEvidence()
			evidence.Objectives.NewestValidBackupAt = evidence.Objectives.IncidentAt
			evidence.Objectives.AuthenticatedSmokeAt = evidence.Objectives.RestoreStartedAt
			evidence.Objectives.RPOSeconds, evidence.Objectives.RTOSeconds = 0, 0
			data, err := json.Marshal(evidence)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := Decode(data); err != nil {
				t.Fatal("explicit zero objective control rejected")
			}
			data = bytes.Replace(data, []byte(`,"`+name+`":0`), nil, 1)
			if _, err := Decode(data); err == nil {
				t.Fatal("missing nested objective measurement became zero")
			}
		})
	}
	plan := validExecutionPlan()
	encoded, err := json.Marshal(plan)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "plan.json")
	if err := os.WriteFile(path, encoded, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadExecutionPlan(path); err != nil {
		t.Fatalf("complete plan rejected: %v", err)
	}
	mutated := bytes.Replace(encoded, []byte(`,"expected_exit":0`), nil, 1)
	if bytes.Equal(encoded, mutated) {
		t.Fatal("positive expected-exit mutation missed its subject")
	}
	if err := os.WriteFile(path, mutated, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadExecutionPlan(path); err == nil {
		t.Fatal("missing planned positive exit became zero")
	}
}

func TestRehearsalReadersRequireExplicitSecretFindings(t *testing.T) {
	scan := validSecretScanResultForSupply([]byte("bound manifest"), "cccccccccccccccccccccccccccccccccccccccc", validEvidence().StartedAt)
	data, err := json.Marshal(scan)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := releasepackage.DecodeSecretScanResult(data); err != nil {
		t.Fatalf("explicit zero-findings control rejected: %v", err)
	}
	for name, mutated := range map[string][]byte{
		"missing findings":    bytes.Replace(data, []byte(`,"findings":0`), nil, 1),
		"null findings":       bytes.Replace(data, []byte(`"findings":0`), []byte(`"findings":null`), 1),
		"case alias":          bytes.Replace(data, []byte(`"findings"`), []byte(`"FINDINGS"`), 1),
		"overwritten finding": append([]byte(`{"findings":1,`), data[1:]...),
	} {
		t.Run(name, func(t *testing.T) {
			if bytes.Equal(data, mutated) {
				t.Fatal("findings mutation missed its subject")
			}
			if _, err := releasepackage.DecodeSecretScanResult(mutated); err == nil {
				t.Fatal("unrecorded or contradictory scan findings accepted")
			}
		})
	}
}

func TestBaseRunRejectsUnrecordedHostGuard(t *testing.T) {
	for _, test := range []struct{ name, from, to string }{
		{"complete rebound control", `"guard_exhausted":false`, `"guard_exhausted": false`},
		{"missing", `,"guard_exhausted":false`, ``},
		{"null", `"guard_exhausted":false`, `"guard_exhausted":null`},
		{"case alias", `"guard_exhausted"`, `"GUARD_EXHAUSTED"`},
		{"overwritten failure", `"guard_exhausted":false`, `"guard_exhausted":true,"guard_exhausted":false`},
	} {
		t.Run(test.name, func(t *testing.T) {
			fixture := boundRunFixture(t)
			basePath := filepath.Join(fixture.sealDirectory, "base-evidence.json")
			base, err := decodeBaseEvidence(mustRead(t, basePath))
			if err != nil {
				t.Fatal(err)
			}
			artifactPath := filepath.Join(fixture.artifactsDirectory, requiredRunArtifactFiles["host_observation"])
			var compact bytes.Buffer
			if err := json.Compact(&compact, mustRead(t, artifactPath)); err != nil {
				t.Fatal(err)
			}
			changed := bytes.Replace(compact.Bytes(), []byte(test.from), []byte(test.to), 1)
			if bytes.Equal(compact.Bytes(), changed) || !json.Valid(changed) {
				t.Fatal("guard mutation missed its actual artifact")
			}
			mustRewrite(t, artifactPath, changed)
			setArtifactDigest(&base, "host_observation", hashBytes(changed))
			writeEvidenceFixture(t, basePath, base)
			_, err = LoadAndValidateBaseRun(basePath, fixture.planPath, fixture.resultsDirectory,
				fixture.artifactsDirectory, fixture.candidateBundleDirectory)
			if test.name == "complete rebound control" {
				if err != nil {
					t.Fatalf("complete rebound control rejected: %v", err)
				}
			} else if err == nil {
				t.Fatal("bound dossier admitted an unrecorded or contradictory host guard")
			}
		})
	}
}

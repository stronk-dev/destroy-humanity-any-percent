package deploymentrehearsal

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"cloud-clicker/server/deploymentbrowser"
	"cloud-clicker/server/operations"
	"cloud-clicker/server/releasepackage"
)

func TestRunValidationBindsPlanAndEveryExactResultByte(t *testing.T) {
	fixture := boundRunFixture(t)
	validated, err := LoadAndValidateRun(fixture.evidencePath, fixture.planPath, fixture.resultsDirectory, fixture.artifactsDirectory, fixture.candidateBundleDirectory, fixture.sealDirectory)
	if err != nil || validated.RunID != fixture.evidence.RunID {
		t.Fatalf("bound run rejected: run=%s err=%v", validated.RunID, err)
	}

	structural := validEvidence()
	if err := Validate(structural); err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(structural)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(fixture.evidencePath, append(data, '\n'), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadAndValidateRun(fixture.evidencePath, fixture.planPath, fixture.resultsDirectory, fixture.artifactsDirectory, fixture.candidateBundleDirectory, fixture.sealDirectory); !errors.Is(err, ErrInvalid) {
		t.Fatalf("structurally valid forged evidence accepted: %v", err)
	}
}

func TestBaseRunValidationAcceptsOnlyTheFortyTwoCompletedInputs(t *testing.T) {
	fixture := boundRunFixture(t)
	base := fixture.evidence
	base.Populations = removePopulation(base.Populations, "forged_successful_evidence")
	base.Artifacts = filterNamedArtifacts(base.Artifacts, BaseRequiredArtifacts)
	writeEvidenceFixture(t, fixture.evidencePath, base)
	validated, err := LoadAndValidateBaseRun(fixture.evidencePath, fixture.planPath, fixture.resultsDirectory,
		fixture.artifactsDirectory, fixture.candidateBundleDirectory)
	if err != nil || len(validated.Populations) != len(planPopulations()) {
		t.Fatalf("base run rejected: populations=%d err=%v", len(validated.Populations), err)
	}
	if _, err := LoadAndValidateRun(fixture.evidencePath, fixture.planPath, fixture.resultsDirectory,
		fixture.artifactsDirectory, fixture.candidateBundleDirectory, fixture.sealDirectory); !errors.Is(err, ErrInvalid) {
		t.Fatalf("unsealed base accepted as final: %v", err)
	}
}

func TestRunValidationRejectsRewrittenMissingAndUnsafeResults(t *testing.T) {
	for name, mutate := range map[string]func(*testing.T, boundFixture){
		"rewritten result": func(t *testing.T, fixture boundFixture) {
			path := filepath.Join(fixture.resultsDirectory, fixture.plan.Checks[0].Name+".json")
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			var result CheckResult
			if json.Unmarshal(data, &result) != nil {
				t.Fatal("decode result")
			}
			result.ExitCode++
			encoded, _ := json.Marshal(result)
			if err := os.WriteFile(path, append(encoded, '\n'), 0o600); err != nil {
				t.Fatal(err)
			}
		},
		"missing result": func(t *testing.T, fixture boundFixture) {
			if err := os.Remove(filepath.Join(fixture.resultsDirectory, fixture.plan.Checks[0].Name+".json")); err != nil {
				t.Fatal(err)
			}
		},
		"extra result": func(t *testing.T, fixture boundFixture) {
			if err := os.WriteFile(filepath.Join(fixture.resultsDirectory, "summary.json"), []byte("{}\n"), 0o600); err != nil {
				t.Fatal(err)
			}
		},
		"unsafe mode": func(t *testing.T, fixture boundFixture) {
			if err := os.Chmod(filepath.Join(fixture.resultsDirectory, fixture.plan.Checks[0].Name+".json"), 0o644); err != nil {
				t.Fatal(err)
			}
		},
	} {
		t.Run(name, func(t *testing.T) {
			fixture := boundRunFixture(t)
			mutate(t, fixture)
			if _, err := LoadAndValidateRun(fixture.evidencePath, fixture.planPath, fixture.resultsDirectory, fixture.artifactsDirectory, fixture.candidateBundleDirectory, fixture.sealDirectory); !errors.Is(err, ErrInvalid) {
				t.Fatalf("invalid result population accepted: %v", err)
			}
		})
	}
}

func TestRunValidationRejectsForgedPopulationEvidenceHash(t *testing.T) {
	fixture := boundRunFixture(t)
	fixture.evidence.Populations[0].EvidenceSHA256 = hashForBuild("f")
	data, err := json.MarshalIndent(fixture.evidence, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(fixture.evidencePath, append(data, '\n'), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadAndValidateRun(fixture.evidencePath, fixture.planPath, fixture.resultsDirectory, fixture.artifactsDirectory, fixture.candidateBundleDirectory, fixture.sealDirectory); !errors.Is(err, ErrInvalid) {
		t.Fatalf("forged population evidence hash accepted: %v", err)
	}
}

func TestBaseRunRejectsRehashedDifferentHostAndObjectives(t *testing.T) {
	for name, mutate := range map[string]func(*testing.T, *boundFixture, *Evidence){
		"host": func(t *testing.T, fixture *boundFixture, base *Evidence) {
			path := filepath.Join(fixture.artifactsDirectory, requiredRunArtifactFiles["host_observation"])
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			observation, err := DecodeHostObservation(data)
			if err != nil {
				t.Fatal(err)
			}
			observation.Host.Kernel = "6.12.1"
			data, _ = json.Marshal(observation)
			data = append(data, '\n')
			if err := os.WriteFile(path, data, 0o600); err != nil {
				t.Fatal(err)
			}
			setArtifactDigest(base, "host_observation", hashBytes(data))
		},
		"objectives": func(t *testing.T, fixture *boundFixture, base *Evidence) {
			path := filepath.Join(fixture.artifactsDirectory, requiredRunArtifactFiles["objective_observation"])
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			observation, err := DecodeObjectiveObservation(data)
			if err != nil {
				t.Fatal(err)
			}
			observation.Objectives.NewestValidBackupAt = observation.Objectives.NewestValidBackupAt.Add(-time.Minute)
			observation.Objectives.RPOSeconds += 60
			data, _ = json.Marshal(observation)
			data = append(data, '\n')
			if err := os.WriteFile(path, data, 0o600); err != nil {
				t.Fatal(err)
			}
			setArtifactDigest(base, "objective_observation", hashBytes(data))
		},
	} {
		t.Run(name, func(t *testing.T) {
			fixture := boundRunFixture(t)
			basePath := filepath.Join(fixture.sealDirectory, "base-evidence.json")
			data, err := os.ReadFile(basePath)
			if err != nil {
				t.Fatal(err)
			}
			base, err := decodeBaseEvidence(data)
			if err != nil {
				t.Fatal(err)
			}
			mutate(t, &fixture, &base)
			writeEvidenceFixture(t, basePath, base)
			if _, err := LoadAndValidateBaseRun(basePath, fixture.planPath, fixture.resultsDirectory,
				fixture.artifactsDirectory, fixture.candidateBundleDirectory); !errors.Is(err, ErrInvalid) {
				t.Fatalf("rehashed different %s accepted: %v", name, err)
			}
		})
	}
}

func TestRunValidationRejectsForgedArtifactsAndStepAggregation(t *testing.T) {
	for name, mutate := range map[string]func(*testing.T, *boundFixture){
		"changed artifact": func(t *testing.T, fixture *boundFixture) {
			path := filepath.Join(fixture.artifactsDirectory, requiredRunArtifactFiles["browser_result"])
			if err := os.WriteFile(path, []byte("forged browser result\n"), 0o600); err != nil {
				t.Fatal(err)
			}
		},
		"missing artifact": func(t *testing.T, fixture *boundFixture) {
			if err := os.Remove(filepath.Join(fixture.artifactsDirectory, requiredRunArtifactFiles["alert_delivery"])); err != nil {
				t.Fatal(err)
			}
		},
		"forged step output": func(t *testing.T, fixture *boundFixture) {
			fixture.evidence.Steps[0].OutputSHA256 = hashForBuild("f")
			writeEvidenceFixture(t, fixture.evidencePath, fixture.evidence)
		},
		"rehashed different host": func(t *testing.T, fixture *boundFixture) {
			path := filepath.Join(fixture.artifactsDirectory, requiredRunArtifactFiles["host_observation"])
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			observation, err := DecodeHostObservation(data)
			if err != nil {
				t.Fatal(err)
			}
			observation.Host.Kernel = "6.12.1"
			data, _ = json.Marshal(observation)
			data = append(data, '\n')
			if err := os.WriteFile(path, data, 0o600); err != nil {
				t.Fatal(err)
			}
			setArtifactDigest(&fixture.evidence, "host_observation", hashBytes(data))
			writeEvidenceFixture(t, fixture.evidencePath, fixture.evidence)
		},
		"rehashed different objectives": func(t *testing.T, fixture *boundFixture) {
			path := filepath.Join(fixture.artifactsDirectory, requiredRunArtifactFiles["objective_observation"])
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			observation, err := DecodeObjectiveObservation(data)
			if err != nil {
				t.Fatal(err)
			}
			observation.Objectives.NewestValidBackupAt = observation.Objectives.NewestValidBackupAt.Add(-time.Minute)
			observation.Objectives.RPOSeconds += 60
			data, _ = json.Marshal(observation)
			data = append(data, '\n')
			if err := os.WriteFile(path, data, 0o600); err != nil {
				t.Fatal(err)
			}
			setArtifactDigest(&fixture.evidence, "objective_observation", hashBytes(data))
			writeEvidenceFixture(t, fixture.evidencePath, fixture.evidence)
		},
		"rehashed invalid browser": func(t *testing.T, fixture *boundFixture) {
			path := filepath.Join(fixture.artifactsDirectory, requiredRunArtifactFiles["browser_result"])
			data := []byte(`{"schema_version":1,"summary":"passed"}` + "\n")
			if err := os.WriteFile(path, data, 0o600); err != nil {
				t.Fatal(err)
			}
			setArtifactDigest(&fixture.evidence, "browser_result", hashBytes(data))
			writeEvidenceFixture(t, fixture.evidencePath, fixture.evidence)
		},
		"rehashed invalid journal": func(t *testing.T, fixture *boundFixture) {
			path := filepath.Join(fixture.artifactsDirectory, requiredRunArtifactFiles["journal_observation"])
			data := []byte(`{"schema_version":1,"objective_completed":true}` + "\n")
			if err := os.WriteFile(path, data, 0o600); err != nil {
				t.Fatal(err)
			}
			setArtifactDigest(&fixture.evidence, "journal_observation", hashBytes(data))
			writeEvidenceFixture(t, fixture.evidencePath, fixture.evidence)
		},
		"rehashed invalid secret scan": func(t *testing.T, fixture *boundFixture) {
			path := filepath.Join(fixture.artifactsDirectory, requiredRunArtifactFiles["secret_scan"])
			data := []byte(`{"schema_version":1,"objective_completed":true}` + "\n")
			if err := os.WriteFile(path, data, 0o600); err != nil {
				t.Fatal(err)
			}
			setArtifactDigest(&fixture.evidence, "secret_scan", hashBytes(data))
			writeEvidenceFixture(t, fixture.evidencePath, fixture.evidence)
		},
		"rehashed invalid supply chain": func(t *testing.T, fixture *boundFixture) {
			path := filepath.Join(fixture.artifactsDirectory, requiredRunArtifactFiles["supply_chain"])
			data := []byte(`{"schema_version":1,"objective_completed":true}` + "\n")
			if err := os.WriteFile(path, data, 0o600); err != nil {
				t.Fatal(err)
			}
			setArtifactDigest(&fixture.evidence, "supply_chain", hashBytes(data))
			writeEvidenceFixture(t, fixture.evidencePath, fixture.evidence)
		},
		"rehashed health-only alert": func(t *testing.T, fixture *boundFixture) {
			path := filepath.Join(fixture.artifactsDirectory, requiredRunArtifactFiles["alert_delivery"])
			data := []byte(`{"schema_version":1,"objective_completed":true}` + "\n")
			if err := os.WriteFile(path, data, 0o600); err != nil {
				t.Fatal(err)
			}
			setArtifactDigest(&fixture.evidence, "alert_delivery", hashBytes(data))
			writeEvidenceFixture(t, fixture.evidencePath, fixture.evidence)
		},
	} {
		t.Run(name, func(t *testing.T) {
			fixture := boundRunFixture(t)
			mutate(t, &fixture)
			if _, err := LoadAndValidateRun(fixture.evidencePath, fixture.planPath, fixture.resultsDirectory, fixture.artifactsDirectory, fixture.candidateBundleDirectory, fixture.sealDirectory); !errors.Is(err, ErrInvalid) {
				t.Fatalf("forged artifact/step accepted: %v", err)
			}
		})
	}
}

func TestRunValidationBindsExactCandidateToolBytes(t *testing.T) {
	for name, mutate := range map[string]func(*testing.T, boundFixture){
		"changed tool": func(t *testing.T, fixture boundFixture) {
			path := filepath.Join(fixture.candidateBundleDirectory, "deployment-browser")
			if err := os.Chmod(path, 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, []byte("forged\n"), 0o555); err != nil {
				t.Fatal(err)
			}
			if err := os.Chmod(path, 0o555); err != nil {
				t.Fatal(err)
			}
		},
		"non executable": func(t *testing.T, fixture boundFixture) {
			if err := os.Chmod(filepath.Join(fixture.candidateBundleDirectory, "deployment-release"), 0o444); err != nil {
				t.Fatal(err)
			}
		},
		"symlinked tool": func(t *testing.T, fixture boundFixture) {
			path := filepath.Join(fixture.candidateBundleDirectory, "deployment-rehearsal")
			if err := os.Remove(path); err != nil {
				t.Fatal(err)
			}
			if err := os.Symlink("deployment-release", path); err != nil {
				t.Fatal(err)
			}
		},
	} {
		t.Run(name, func(t *testing.T) {
			fixture := boundRunFixture(t)
			mutate(t, fixture)
			if _, err := LoadAndValidateRun(fixture.evidencePath, fixture.planPath, fixture.resultsDirectory, fixture.artifactsDirectory, fixture.candidateBundleDirectory, fixture.sealDirectory); !errors.Is(err, ErrInvalid) {
				t.Fatalf("unbound candidate tool accepted: %v", err)
			}
		})
	}
}

type boundFixture struct {
	evidence                 Evidence
	plan                     ExecutionPlan
	evidencePath             string
	planPath                 string
	resultsDirectory         string
	artifactsDirectory       string
	candidateBundleDirectory string
	sealDirectory            string
}

func boundRunFixture(t *testing.T) boundFixture {
	t.Helper()
	root := t.TempDir()
	results := filepath.Join(root, "results")
	artifactsDirectory := filepath.Join(root, "artifacts")
	candidateBundleDirectory := filepath.Join(root, "candidate-bundle")
	sealDirectory := filepath.Join(root, "seal")
	if err := os.Mkdir(results, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(artifactsDirectory, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(candidateBundleDirectory, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(sealDirectory, 0o700); err != nil {
		t.Fatal(err)
	}
	evidence := validEvidence()
	candidateManifestBytes := []byte(`{"source_commit":"` + strings.Repeat("c", 40) + `"}` + "\n")
	previousManifestBytes := []byte(`{"source_commit":"` + strings.Repeat("d", 40) + `"}` + "\n")
	candidateBuild := validBuildRecord()
	candidateBuild.SchemaVersion = 2
	candidateBuild.Role = "candidate"
	candidateBuild.SourceCommit = strings.Repeat("c", 40)
	candidateBuild.ManifestSHA256 = hashBytes(candidateManifestBytes)
	candidateBuild.RebuildManifestSHA256 = candidateBuild.ManifestSHA256
	candidateBuild.RehearsalImages = []BuildImage{validRehearsalImage()}
	previousBuild := validBuildRecord()
	previousBuild.SourceCommit = strings.Repeat("d", 40)
	previousBuild.ManifestSHA256 = hashBytes(previousManifestBytes)
	previousBuild.RebuildManifestSHA256 = previousBuild.ManifestSHA256
	candidateBuildBytes, _ := json.Marshal(candidateBuild)
	candidateBuildBytes = append(candidateBuildBytes, '\n')
	previousBuildBytes, _ := json.Marshal(previousBuild)
	previousBuildBytes = append(previousBuildBytes, '\n')
	toolFiles := map[string]string{"deployment-rehearsal": "deployment-rehearsal", "deployment-release": "deployment-release", "browser-driver": "deployment-browser"}
	for index := range evidence.Tools {
		data := []byte(evidence.Tools[index].Name + " binary\n")
		if err := os.WriteFile(filepath.Join(candidateBundleDirectory, toolFiles[evidence.Tools[index].Name]), data, 0o555); err != nil {
			t.Fatal(err)
		}
		evidence.Tools[index].SHA256 = hashBytes(data)
	}
	artifactBytes := map[string][]byte{}
	for name, file := range requiredRunArtifactFiles {
		artifactBytes[name] = []byte(name + "\n")
		if name == "candidate_manifest" {
			artifactBytes[name] = candidateManifestBytes
		}
		if name == "previous_manifest" {
			artifactBytes[name] = previousManifestBytes
		}
		if name == "candidate_build" {
			artifactBytes[name] = candidateBuildBytes
		}
		if name == "previous_build" {
			artifactBytes[name] = previousBuildBytes
		}
		if name == "host_observation" {
			observation := HostObservation{SchemaVersion: 1, StartedAt: evidence.StartedAt,
				CompletedAt: evidence.StartedAt.Add(time.Second), Host: evidence.Host, ObjectiveCompleted: true}
			artifactBytes[name], _ = json.Marshal(observation)
			artifactBytes[name] = append(artifactBytes[name], '\n')
		}
		if name == "objective_observation" {
			observation := ObjectiveObservation{SchemaVersion: 1, StartedAt: evidence.StartedAt,
				CompletedAt: evidence.Objectives.AuthenticatedSmokeAt, Objectives: evidence.Objectives, ObjectiveCompleted: true}
			artifactBytes[name], _ = json.Marshal(observation)
			artifactBytes[name] = append(artifactBytes[name], '\n')
		}
		if name == "browser_result" {
			result := deploymentbrowser.Result{SchemaVersion: 1, ManifestSHA256: hashBytes(candidateManifestBytes),
				StartedAt: evidence.StartedAt.Add(time.Second), CompletedAt: evidence.StartedAt.Add(2 * time.Second), Surface: "desk",
				BootstrapCommitted: true, CredentialsPresent: true, WebSocketObserved: true, ManualIntentObserved: true,
				ManualIntentStatus: 200, ObjectiveCompleted: true}
			artifactBytes[name], _ = json.Marshal(result)
			artifactBytes[name] = append(artifactBytes[name], '\n')
		}
		if name == "secret_scan" {
			result := releasepackage.SecretScanResult{SchemaVersion: 1, ManifestSHA256: hashBytes(candidateManifestBytes), SourceCommit: strings.Repeat("c", 40),
				StartedAt: evidence.StartedAt.Add(5 * time.Second), CompletedAt: evidence.StartedAt.Add(6 * time.Second), TrackedFiles: 100,
				ImageArchiveScanned: true, ObjectiveCompleted: true}
			artifactBytes[name], _ = json.Marshal(result)
			artifactBytes[name] = append(artifactBytes[name], '\n')
		}
		if name == "supply_chain" {
			result := SupplyChainResult{SchemaVersion: 1, CandidateManifestSHA256: hashBytes(candidateManifestBytes),
				PreviousManifestSHA256: hashBytes(previousManifestBytes), CandidateBuildSHA256: hashBytes(candidateBuildBytes),
				PreviousBuildSHA256: hashBytes(previousBuildBytes), StartedAt: evidence.StartedAt.Add(7 * time.Second),
				CompletedAt: evidence.StartedAt.Add(8 * time.Second), ProductionImages: 6, RehearsalImages: 1, SBOMDocuments: 8,
				RootAttributionPresent: true, SiteAttributionPresent: true, SourceAndImageProvenance: true, ObjectiveCompleted: true}
			artifactBytes[name], _ = json.Marshal(result)
			artifactBytes[name] = append(artifactBytes[name], '\n')
		}
		if name == "journal_observation" {
			observation := operations.JournalObservation{SchemaVersion: 1, Population: "r006-standard-workload",
				StartedAt: evidence.StartedAt.Add(3 * time.Second), CompletedAt: evidence.StartedAt.Add(4 * time.Second),
				ObjectiveCompleted: true, Samples: 2, ObservedBytes: 100, PeakBytesPerDay: 100,
				FilesystemBytes: 10_000, JournalMaxUseBytes: 1_400, JournalRetentionSeconds: int64(operations.JournalRetention / time.Second),
				StorageAlertFraction: 0.8}
			artifactBytes[name], _ = json.Marshal(observation)
			artifactBytes[name] = append(artifactBytes[name], '\n')
		}
		if name == "alert_delivery" {
			alerts := make([]operations.AlertObservation, len(operations.ReleaseFloorAlertNames))
			for index, alertName := range operations.ReleaseFloorAlertNames {
				alerts[index] = operations.AlertObservation{Name: alertName, FiringDelivered: true, ResolvedDelivered: true}
			}
			observation := operations.AlertDeliveryObservation{SchemaVersion: 1, ManifestSHA256: hashBytes(candidateManifestBytes),
				StartedAt: evidence.StartedAt.Add(9 * time.Second), CompletedAt: evidence.StartedAt.Add(10 * time.Second),
				RuleFixturesPassed: true, Alerts: alerts, ObjectiveCompleted: true}
			artifactBytes[name], _ = json.Marshal(observation)
			artifactBytes[name] = append(artifactBytes[name], '\n')
		}
		if err := os.WriteFile(filepath.Join(artifactsDirectory, file), artifactBytes[name], 0o600); err != nil {
			t.Fatal(err)
		}
	}
	evidence.ManifestSHA256 = hashBytes(artifactBytes["candidate_manifest"])
	evidence.PreviousManifestSHA256 = hashBytes(artifactBytes["previous_manifest"])
	secretScanDigest := hashBytes(artifactBytes["secret_scan"])
	var supplyResult SupplyChainResult
	if json.Unmarshal(artifactBytes["supply_chain"], &supplyResult) != nil {
		t.Fatal("decode supply fixture")
	}
	supplyResult.SecretScanSHA256 = secretScanDigest
	artifactBytes["supply_chain"], _ = json.Marshal(supplyResult)
	artifactBytes["supply_chain"] = append(artifactBytes["supply_chain"], '\n')
	if err := os.WriteFile(filepath.Join(artifactsDirectory, requiredRunArtifactFiles["supply_chain"]), artifactBytes["supply_chain"], 0o600); err != nil {
		t.Fatal(err)
	}
	plan := validExecutionPlan()
	plan.RunID = evidence.RunID
	plan.ManifestSHA256 = evidence.ManifestSHA256
	plan.PreviousManifestSHA256 = evidence.PreviousManifestSHA256
	planBytes, err := json.MarshalIndent(plan, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	planBytes = append(planBytes, '\n')
	for index := range evidence.Artifacts {
		if evidence.Artifacts[index].Name == "rehearsal_plan" {
			evidence.Artifacts[index].SHA256 = hashBytes(planBytes)
		} else if data, ok := artifactBytes[evidence.Artifacts[index].Name]; ok {
			evidence.Artifacts[index].SHA256 = hashBytes(data)
		}
	}
	populationIndex := map[string]int{}
	for index, population := range evidence.Populations {
		populationIndex[population.Name] = index
	}
	for index, check := range plan.Checks {
		started := evidence.StartedAt.Add(time.Duration(index+1) * time.Millisecond)
		result := CheckResult{SchemaVersion: 1, Name: check.Name, Kind: check.Kind, Step: check.Step,
			StartedAt: started, CompletedAt: started.Add(time.Microsecond), CommandSHA256: hashJSON(check.Command),
			OutputSHA256: hashBytes([]byte(check.Name)), ExitCode: check.ExpectedExit, Result: "passed"}
		data, err := json.Marshal(result)
		if err != nil {
			t.Fatal(err)
		}
		data = append(data, '\n')
		if err := os.WriteFile(filepath.Join(results, check.Name+".json"), data, 0o600); err != nil {
			t.Fatal(err)
		}
		evidence.Populations[populationIndex[check.Name]].EvidenceSHA256 = hashBytes(data)
	}
	for index := range evidence.Steps {
		step := &evidence.Steps[index]
		commandHashes, resultHashes := []string{}, []string{}
		var started, completed time.Time
		for _, check := range plan.Checks {
			if check.Step != step.Name {
				continue
			}
			data, err := os.ReadFile(filepath.Join(results, check.Name+".json"))
			if err != nil {
				t.Fatal(err)
			}
			result, err := decodeCheckResult(data)
			if err != nil {
				t.Fatal(err)
			}
			commandHashes = append(commandHashes, result.CommandSHA256)
			resultHashes = append(resultHashes, hashBytes(data))
			if started.IsZero() || result.StartedAt.Before(started) {
				started = result.StartedAt
			}
			if result.CompletedAt.After(completed) {
				completed = result.CompletedAt
			}
		}
		step.StartedAt, step.CompletedAt = started, completed
		step.InputSHA256, step.OutputSHA256 = hashJSON(commandHashes), hashJSON(resultHashes)
	}
	planPath := filepath.Join(root, "plan.json")
	if err := os.WriteFile(planPath, planBytes, 0o600); err != nil {
		t.Fatal(err)
	}
	base := evidence
	base.Populations = removePopulation(base.Populations, "forged_successful_evidence")
	base.Artifacts = filterNamedArtifacts(base.Artifacts, BaseRequiredArtifacts)
	base.CompletedAt = evidence.CompletedAt.Add(-2 * time.Second)
	baseBytes, err := json.MarshalIndent(base, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	baseBytes = append(baseBytes, '\n')
	if err := os.WriteFile(filepath.Join(sealDirectory, "base-evidence.json"), baseBytes, 0o600); err != nil {
		t.Fatal(err)
	}
	proof := ForgeryProof{SchemaVersion: 1, RunID: base.RunID, ManifestSHA256: base.ManifestSHA256,
		BaseEvidenceSHA256: hashBytes(baseBytes), Mutation: ForgeryMutation, StartedAt: base.CompletedAt,
		CompletedAt: base.CompletedAt.Add(time.Second), RejectionObserved: true, ObjectiveCompleted: true}
	proofBytes, err := json.Marshal(proof)
	if err != nil {
		t.Fatal(err)
	}
	proofBytes = append(proofBytes, '\n')
	if err := os.WriteFile(filepath.Join(sealDirectory, "forgery-proof.json"), proofBytes, 0o600); err != nil {
		t.Fatal(err)
	}
	setArtifactDigest(&evidence, "base_evidence", hashBytes(baseBytes))
	setArtifactDigest(&evidence, "forgery_proof", hashBytes(proofBytes))
	setPopulationDigest(&evidence, "forged_successful_evidence", hashBytes(proofBytes))
	evidenceBytes, err := json.MarshalIndent(evidence, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	evidencePath := filepath.Join(root, "evidence.json")
	if err := os.WriteFile(evidencePath, append(evidenceBytes, '\n'), 0o600); err != nil {
		t.Fatal(err)
	}
	return boundFixture{evidence: evidence, plan: plan, evidencePath: evidencePath, planPath: planPath,
		resultsDirectory: results, artifactsDirectory: artifactsDirectory, candidateBundleDirectory: candidateBundleDirectory, sealDirectory: sealDirectory}
}

func setPopulationDigest(evidence *Evidence, name, digest string) {
	for index := range evidence.Populations {
		if evidence.Populations[index].Name == name {
			evidence.Populations[index].EvidenceSHA256 = digest
			return
		}
	}
}

func setArtifactDigest(evidence *Evidence, name, digest string) {
	for index := range evidence.Artifacts {
		if evidence.Artifacts[index].Name == name {
			evidence.Artifacts[index].SHA256 = digest
			return
		}
	}
}

func removePopulation(populations []Population, name string) []Population {
	result := make([]Population, 0, len(populations)-1)
	for _, population := range populations {
		if population.Name != name {
			result = append(result, population)
		}
	}
	return result
}

func writeEvidenceFixture(t *testing.T, path string, evidence Evidence) {
	t.Helper()
	data, err := json.MarshalIndent(evidence, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, append(data, '\n'), 0o600); err != nil {
		t.Fatal(err)
	}
}

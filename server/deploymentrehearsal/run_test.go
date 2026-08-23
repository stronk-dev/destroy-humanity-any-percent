package deploymentrehearsal

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRunValidationBindsPlanAndEveryExactResultByte(t *testing.T) {
	fixture := boundRunFixture(t)
	validated, err := LoadAndValidateRun(fixture.evidencePath, fixture.planPath, fixture.resultsDirectory)
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
	if _, err := LoadAndValidateRun(fixture.evidencePath, fixture.planPath, fixture.resultsDirectory); !errors.Is(err, ErrInvalid) {
		t.Fatalf("structurally valid forged evidence accepted: %v", err)
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
			if _, err := LoadAndValidateRun(fixture.evidencePath, fixture.planPath, fixture.resultsDirectory); !errors.Is(err, ErrInvalid) {
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
	if _, err := LoadAndValidateRun(fixture.evidencePath, fixture.planPath, fixture.resultsDirectory); !errors.Is(err, ErrInvalid) {
		t.Fatalf("forged population evidence hash accepted: %v", err)
	}
}

type boundFixture struct {
	evidence         Evidence
	plan             ExecutionPlan
	evidencePath     string
	planPath         string
	resultsDirectory string
}

func boundRunFixture(t *testing.T) boundFixture {
	t.Helper()
	root := t.TempDir()
	results := filepath.Join(root, "results")
	if err := os.Mkdir(results, 0o700); err != nil {
		t.Fatal(err)
	}
	evidence := validEvidence()
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
	planPath := filepath.Join(root, "plan.json")
	if err := os.WriteFile(planPath, planBytes, 0o600); err != nil {
		t.Fatal(err)
	}
	evidenceBytes, err := json.MarshalIndent(evidence, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	evidencePath := filepath.Join(root, "evidence.json")
	if err := os.WriteFile(evidencePath, append(evidenceBytes, '\n'), 0o600); err != nil {
		t.Fatal(err)
	}
	return boundFixture{evidence: evidence, plan: plan, evidencePath: evidencePath, planPath: planPath, resultsDirectory: results}
}

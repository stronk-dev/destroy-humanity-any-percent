package deploymentrehearsal

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestEvidenceValidatesOnlyCompleteExactRehearsal(t *testing.T) {
	evidence := validEvidence()
	data, err := json.Marshal(evidence)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := Decode(data)
	if err != nil || decoded.RunID != evidence.RunID {
		t.Fatalf("valid evidence rejected: decoded=%+v err=%v", decoded, err)
	}
	path := filepath.Join(t.TempDir(), "r006.json")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(path); err != nil {
		t.Fatalf("persisted valid evidence rejected: %v", err)
	}
}

func TestBaseEvidenceSeparatesNonCircularForgeryPopulation(t *testing.T) {
	base := validEvidence()
	populations := make([]Population, 0, len(base.Populations)-1)
	for _, population := range base.Populations {
		if population.Name != "forged_successful_evidence" {
			populations = append(populations, population)
		}
	}
	base.Populations = populations
	if err := ValidateBaseEvidence(base); err != nil {
		t.Fatalf("valid 42-population base rejected: %v", err)
	}
	if err := Validate(base); !errors.Is(err, ErrInvalid) {
		t.Fatalf("base accepted as final evidence: %v", err)
	}
}

func TestEvidenceRejectsIncompleteForgedAndGuardedClaims(t *testing.T) {
	mutations := map[string]func(*Evidence){
		"objective incomplete": func(value *Evidence) { value.ObjectiveCompleted = false },
		"guard exhausted":      func(value *Evidence) { value.GuardExhausted = true },
		"unsupported host":     func(value *Evidence) { value.Host.Architecture = "arm64" },
		"dirty host":           func(value *Evidence) { value.Host.SourceCheckoutAbsent = false },
		"provider credential":  func(value *Evidence) { value.Host.ProviderCredentialsAbsent = false },
		"missing tool":         func(value *Evidence) { value.Tools = value.Tools[:2] },
		"missing step":         func(value *Evidence) { value.Steps = value.Steps[:len(value.Steps)-1] },
		"reordered step": func(value *Evidence) {
			value.Steps[0], value.Steps[1] = value.Steps[1], value.Steps[0]
		},
		"failed step":        func(value *Evidence) { value.Steps[0].Result = "failed" },
		"missing population": func(value *Evidence) { value.Populations = value.Populations[:len(value.Populations)-1] },
		"vacuous negative": func(value *Evidence) {
			for index := range value.Populations {
				if value.Populations[index].Kind == "negative" {
					value.Populations[index].SeveringCaught = false
					return
				}
			}
		},
		"forged kind": func(value *Evidence) {
			for index := range value.Populations {
				if value.Populations[index].Kind == "negative" {
					value.Populations[index].Kind = "positive"
					value.Populations[index].SeveringCaught = false
					return
				}
			}
		},
		"rpo above bound": func(value *Evidence) {
			value.Objectives.NewestValidBackupAt = value.Objectives.IncidentAt.Add(-MaximumRPO - time.Second)
			value.Objectives.RPOSeconds = int64((MaximumRPO + time.Second) / time.Second)
		},
		"forged rto number": func(value *Evidence) { value.Objectives.RTOSeconds-- },
		"identity mismatch": func(value *Evidence) { value.Objectives.RestoredIdentityMatch = false },
		"missing artifact":  func(value *Evidence) { value.Artifacts = value.Artifacts[:len(value.Artifacts)-1] },
		"missing exclusion": func(value *Evidence) { value.Exclusions = value.Exclusions[:len(value.Exclusions)-1] },
	}
	for name, mutate := range mutations {
		t.Run(name, func(t *testing.T) {
			value := validEvidence()
			mutate(&value)
			if err := Validate(value); !errors.Is(err, ErrInvalid) {
				t.Fatalf("invalid evidence accepted: %v", err)
			}
		})
	}
}

func TestEvidenceDecoderRejectsUnknownPrivateAndTrailingFields(t *testing.T) {
	data, err := json.Marshal(validEvidence())
	if err != nil {
		t.Fatal(err)
	}
	for name, mutate := range map[string]func([]byte) []byte{
		"unknown": func(data []byte) []byte {
			return []byte(strings.TrimSuffix(string(data), "}") + `,"summary":"looks good"}`)
		},
		"private": func(data []byte) []byte {
			return []byte(strings.TrimSuffix(string(data), "}") + `,"account_id":"private"}`)
		},
		"trailing": func(data []byte) []byte { return append(data, []byte(` {}`)...) },
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := Decode(mutate(data)); !errors.Is(err, ErrInvalid) {
				t.Fatalf("%s evidence accepted: %v", name, err)
			}
		})
	}
}

func validEvidence() Evidence {
	start := time.Date(2026, 8, 23, 12, 0, 0, 0, time.UTC)
	hash := func(fill string) string { return "sha256:" + strings.Repeat(fill, 64) }
	steps := make([]Step, len(RequiredSteps))
	for index, name := range RequiredSteps {
		stepStart := start.Add(time.Duration(index) * time.Minute)
		steps[index] = Step{Name: name, CommandClass: requiredCommandClass[name], StartedAt: stepStart,
			CompletedAt: stepStart.Add(30 * time.Second), InputSHA256: hash("a"), OutputSHA256: hash("b"), Result: "passed"}
	}
	populations := make([]Population, 0, len(RequiredPopulations))
	for name, kind := range RequiredPopulations {
		populations = append(populations, Population{Name: name, Kind: kind, Result: "passed", SeveringCaught: kind == "negative", EvidenceSHA256: hash("c")})
	}
	tools := []Tool{{Name: "deployment-rehearsal", SHA256: hash("d")}, {Name: "deployment-release", SHA256: hash("e")}, {Name: "browser-driver", SHA256: hash("f")}}
	artifacts := make([]Artifact, len(RequiredArtifacts))
	for index, name := range RequiredArtifacts {
		artifacts[index] = Artifact{Name: name, SHA256: hash("1")}
	}
	incident := start.Add(12 * time.Minute)
	restore := start.Add(13 * time.Minute)
	return Evidence{SchemaVersion: SchemaVersion, RunID: "r006-20260823-001", ManifestSHA256: hash("2"),
		PreviousManifestSHA256: hash("3"), ReleaseVersion: "1.0.0", PreviousReleaseVersion: "0.9.0",
		StartedAt: start, CompletedAt: start.Add(20 * time.Minute), Host: Host{OS: "linux", Architecture: "amd64",
			Distribution: "debian-13", Kernel: "6.12.0", DockerEngine: "28.3.3", DockerCompose: "2.39.1",
			CleanStart: true, SourceCheckoutAbsent: true, ProviderCredentialsAbsent: true}, Tools: tools, Steps: steps,
		Populations: populations, Objectives: Objectives{IncidentAt: incident, NewestValidBackupAt: incident.Add(-2 * time.Minute),
			RestoreStartedAt: restore, AuthenticatedSmokeAt: restore.Add(2 * time.Minute), RPOSeconds: 120, RTOSeconds: 120,
			RestoredIdentityMatch: true}, Artifacts: artifacts, Exclusions: slicesClone(RequiredExclusions), ObjectiveCompleted: true}
}

func slicesClone(values []string) []string { return append([]string(nil), values...) }

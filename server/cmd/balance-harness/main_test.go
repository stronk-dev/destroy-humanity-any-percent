package main

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"cloud-clicker/server/epochseed"
	"cloud-clicker/server/harness"
)

func TestWriteRelevanceReportRejectsFailuresBeforeWriting(t *testing.T) {
	output := filepath.Join(t.TempDir(), "report.json")
	report := harness.RelevanceReport{Failures: []string{"relevance_floor:upgrade.dead"}}

	err := writeRelevanceReport(output, report)
	if err == nil || !strings.Contains(err.Error(), "relevance_floor:upgrade.dead") {
		t.Fatalf("writeRelevanceReport error = %v", err)
	}
	if _, statErr := os.Stat(output); !os.IsNotExist(statErr) {
		t.Fatalf("failing report was written: %v", statErr)
	}
	diagnosticBytes, readErr := os.ReadFile(relevanceDiagnosticPath(output))
	if readErr != nil {
		t.Fatal(readErr)
	}
	diagnostic := string(diagnosticBytes)
	if !strings.Contains(diagnostic, `"kind": "non_authoritative_relevance_diagnostic"`) ||
		!strings.Contains(diagnostic, `"authoritative": false`) || !strings.Contains(diagnostic, `"relevance_floor:upgrade.dead"`) {
		t.Fatalf("non-authoritative diagnostic=%s", diagnosticBytes)
	}
}

func TestRegisteredRelevanceSelectorUsesRegistryAuthority(t *testing.T) {
	root := filepath.Clean("../../..")
	output := filepath.Join(t.TempDir(), "report.json")
	observation := filepath.Join(t.TempDir(), "observation.json")
	epochID := int64(8)
	bundleHash := mustEpochHash(t, root)
	var err error
	observationRecorder, err = harness.NewHarnessObservationRecorder(observation, "relevance-registered", &epochID, bundleHash)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { observationRecorder = nil })
	selector := "testdata/harness/relevance/scenario-v1.json"
	if err := runRegisteredRelevance(root, output, selector); err != nil {
		t.Fatal(err)
	}
	if err := observationRecorder.Complete(); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	want, err := os.ReadFile(filepath.Join(root, "testdata/harness/relevance/golden-report-v1.json"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(want) {
		t.Fatal("registered selector output differs from registered golden")
	}
	artifact, err := harness.LoadHarnessObservation(observation)
	if err != nil {
		t.Fatal(err)
	}
	if err := harness.ValidateCompleteHarnessObservation(artifact); err != nil {
		t.Fatal(err)
	}
	if len(artifact.Objectives) != 1 || artifact.Objectives[0].Identity.ScenarioPath != selector ||
		artifact.Objectives[0].Work.DeclaredRuns == nil || *artifact.Objectives[0].Work.DeclaredRuns != 23 {
		t.Fatalf("registered observation=%+v", artifact)
	}
}

func TestRegisteredRelevanceSelectorRejectsUnregisteredPathBeforeSimulation(t *testing.T) {
	root := filepath.Clean("../../..")
	observation := filepath.Join(t.TempDir(), "observation.json")
	epochID := int64(8)
	var err error
	observationRecorder, err = harness.NewHarnessObservationRecorder(observation, "relevance-registered", &epochID, mustEpochHash(t, root))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { observationRecorder = nil })
	err = runRegisteredRelevance(root, filepath.Join(t.TempDir(), "report.json"), "balance/testdata/t0-t1/relevance-scenario-t1-v2.json")
	if err == nil || !strings.Contains(err.Error(), "unregistered relevance selector") {
		t.Fatalf("selector error=%v", err)
	}
	artifact, loadErr := harness.LoadHarnessObservation(observation)
	if loadErr != nil {
		t.Fatal(loadErr)
	}
	if len(artifact.Objectives) != 0 || artifact.State != harness.ObservationStateRunning {
		t.Fatalf("unregistered selector began simulation: %+v", artifact)
	}
}

func TestObservationSignalSubprocess(t *testing.T) {
	if os.Getenv("CLOUD_CLICKER_OBSERVATION_SIGNAL_HELPER") == "1" {
		path := os.Getenv("CLOUD_CLICKER_OBSERVATION_SIGNAL_PATH")
		epochID, index, active := int64(8), 0, true
		var err error
		observationRecorder, err = harness.NewHarnessObservationRecorder(path, "relevance-registered", &epochID,
			"sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
		if err != nil {
			os.Exit(2)
		}
		if err := observationRecorder.DeclareObjectives([]string{"relevance:0:scenario.json"}); err != nil {
			os.Exit(2)
		}
		if err := observationRecorder.StartObjective(harness.HarnessObservationObjectiveSpec{
			ID: "relevance:0:scenario.json", Kind: "registered_relevance",
			Identity: harness.HarnessObservationIdentity{RegistryIndex: &index, ScenarioPath: "scenario.json",
				EconomyCatalogPath: "economy.json", RelevancePolicyPath: "policy.json", GoldenReportPath: "golden.json",
				ScenarioHash:        "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
				RelevancePolicyHash: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
				ConstantsHash:       "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", Active: &active}}); err != nil {
			os.Exit(2)
		}
		installObservationSignalHandler()
		select {}
	}
	path := filepath.Join(t.TempDir(), "observation.json")
	command := exec.Command(os.Args[0], "-test.run=^TestObservationSignalSubprocess$")
	command.Env = append(os.Environ(), "CLOUD_CLICKER_OBSERVATION_SIGNAL_HELPER=1", "CLOUD_CLICKER_OBSERVATION_SIGNAL_PATH="+path)
	if err := command.Start(); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(5 * time.Second)
	for {
		artifact, err := harness.LoadHarnessObservation(path)
		if err == nil && artifact.CurrentObjective != nil {
			break
		}
		if time.Now().After(deadline) {
			_ = command.Process.Kill()
			t.Fatalf("signal helper did not start: %v", err)
		}
		time.Sleep(10 * time.Millisecond)
	}
	if err := command.Process.Signal(syscall.SIGTERM); err != nil {
		t.Fatal(err)
	}
	err := command.Wait()
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) || exitErr.ExitCode() == 0 {
		t.Fatalf("signal helper exit=%v", err)
	}
	artifact, err := harness.LoadHarnessObservation(path)
	if err != nil {
		t.Fatal(err)
	}
	if artifact.State != harness.ObservationStateIncomplete || artifact.Termination == nil || *artifact.Termination != "signal" ||
		harness.ValidateCompleteHarnessObservation(artifact) == nil {
		t.Fatalf("signal artifact=%+v", artifact)
	}
}

func mustEpochHash(t *testing.T, root string) string {
	t.Helper()
	bundle, err := epochseed.Load(root)
	if err != nil {
		t.Fatal(err)
	}
	return bundle.Hash
}

func TestWriteRelevanceReportRemovesStaleDiagnosticOnSuccess(t *testing.T) {
	output := filepath.Join(t.TempDir(), "report.json")
	diagnostic := relevanceDiagnosticPath(output)
	if err := os.WriteFile(diagnostic, []byte("stale"), 0o644); err != nil {
		t.Fatal(err)
	}
	report := harness.RelevanceReport{}
	if err := writeRelevanceReport(output, report); err != nil {
		t.Fatal(err)
	}
	if _, statErr := os.Stat(diagnostic); !os.IsNotExist(statErr) {
		t.Fatalf("stale diagnostic survived success: %v", statErr)
	}
}

package harness

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const observationTestHash = "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

func TestHarnessObservationRecorderCompletesAtomicArtifact(t *testing.T) {
	directory := t.TempDir()
	path := filepath.Join(directory, "observation.json")
	epochID := int64(8)
	nowValue := time.Date(2026, 8, 21, 10, 0, 0, 0, time.UTC)
	now := func() time.Time {
		value := nowValue
		nowValue = nowValue.Add(time.Second)
		return value
	}
	recorder, err := newHarnessObservationRecorder(path, "relevance-registered", &epochID, observationTestHash, now)
	if err != nil {
		t.Fatal(err)
	}
	running, err := LoadHarnessObservation(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateCompleteHarnessObservation(running); err == nil {
		t.Fatal("running checkpoint was accepted as complete")
	}
	index, active := 0, true
	if err := recorder.DeclareObjectives([]string{"relevance:0:scenario.json"}); err != nil {
		t.Fatal(err)
	}
	if err := recorder.StartObjective(HarnessObservationObjectiveSpec{ID: "relevance:0:scenario.json", Kind: "registered_relevance",
		Identity: HarnessObservationIdentity{RegistryIndex: &index, ScenarioPath: "scenario.json", EconomyCatalogPath: "economy.json",
			RelevancePolicyPath: "policy.json", GoldenReportPath: "golden.json", ScenarioHash: observationTestHash,
			RelevancePolicyHash: observationTestHash, ConstantsHash: observationTestHash, Active: &active}}); err != nil {
		t.Fatal(err)
	}
	declared, partial, transitions := int64(2), int64(1), int64(9)
	if err := recorder.Progress(HarnessObservationProgress{Work: HarnessObservationWork{DeclaredRuns: &declared,
		ExecutedRuns: &partial, ExecutedTransitions: &transitions}}); err != nil {
		t.Fatal(err)
	}
	executed, declaredTransitions := int64(2), int64(9)
	if err := recorder.CompleteObjective(HarnessObservationProgress{Work: HarnessObservationWork{DeclaredRuns: &declared,
		ExecutedRuns: &executed, DeclaredTransitions: &declaredTransitions, ExecutedTransitions: &transitions},
		GuardState: ObservationConditionClear, PopulationExclusions: ObservationConditionClear,
		TruncationState: ObservationConditionClear}); err != nil {
		t.Fatal(err)
	}
	if err := recorder.Complete(); err != nil {
		t.Fatal(err)
	}
	complete, err := LoadHarnessObservation(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateCompleteHarnessObservation(complete); err != nil {
		t.Fatal(err)
	}
	if complete.Authoritative || complete.State != ObservationStateComplete || len(complete.Objectives) != 1 {
		t.Fatalf("complete observation=%+v", complete)
	}
	entries, err := os.ReadDir(directory)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Name() != "observation.json" {
		t.Fatalf("atomic writer leaked files: %v", entries)
	}
}

func TestValidateCompleteHarnessObservationRejectsIncompleteEvidence(t *testing.T) {
	valid := validHarnessObservationFixture()
	tests := []struct {
		name   string
		mutate func(*HarnessObservation)
		want   string
	}{
		{"running", func(value *HarnessObservation) { value.State = ObservationStateRunning }, "incomplete"},
		{"unknown mode", func(value *HarnessObservation) { value.Mode = "other" }, "incomplete"},
		{"negative elapsed", func(value *HarnessObservation) { value.Objectives[0].ElapsedMS = -1 }, "objective"},
		{"missing exclusions", func(value *HarnessObservation) { value.Objectives[0].InstrumentExcluded = nil }, "objective"},
		{"guard fired", func(value *HarnessObservation) { value.Objectives[0].GuardState = ObservationConditionFired }, "objective"},
		{"population excluded", func(value *HarnessObservation) { value.Objectives[0].PopulationExclusions = ObservationConditionFired }, "objective"},
		{"truncated", func(value *HarnessObservation) { value.Objectives[0].TruncationState = ObservationConditionFired }, "objective"},
		{"run mismatch", func(value *HarnessObservation) { *value.Objectives[0].Work.ExecutedRuns = 1 }, "run cardinality mismatch"},
		{"transition mismatch", func(value *HarnessObservation) { *value.Objectives[0].Work.ExecutedTransitions = 8 }, "transition cardinality mismatch"},
		{"active constants severed", func(value *HarnessObservation) {
			value.Objectives[0].Identity.ConstantsHash = "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
		}, "identity"},
		{"completion severed", func(value *HarnessObservation) { value.Objectives[0].State = ObservationStateRunning }, "objective"},
		{"declared objective severed", func(value *HarnessObservation) {
			value.DeclaredObjectiveIDs = append(value.DeclaredObjectiveIDs, "relevance:1:missing.json")
		}, "incomplete"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			candidate := valid
			candidate.Objectives = append([]HarnessObservationObjective(nil), valid.Objectives...)
			work := valid.Objectives[0].Work
			work.DeclaredRuns = copiedInt64(work.DeclaredRuns)
			work.ExecutedRuns = copiedInt64(work.ExecutedRuns)
			work.DeclaredTransitions = copiedInt64(work.DeclaredTransitions)
			work.ExecutedTransitions = copiedInt64(work.ExecutedTransitions)
			candidate.Objectives[0].Work = work
			test.mutate(&candidate)
			if err := ValidateCompleteHarnessObservation(candidate); err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("validation error=%v, want %q", err, test.want)
			}
		})
	}
}

func TestHarnessObservationFailureAndMissingArtifactAreNotComplete(t *testing.T) {
	path := filepath.Join(t.TempDir(), "observation.json")
	if _, err := LoadHarnessObservation(path); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("missing artifact error=%v", err)
	}
	epochID := int64(8)
	recorder, err := NewHarnessObservationRecorder(path, "relevance-registered", &epochID, observationTestHash)
	if err != nil {
		t.Fatal(err)
	}
	index, active := 0, true
	if err := recorder.DeclareObjectives([]string{"relevance:0:scenario.json"}); err != nil {
		t.Fatal(err)
	}
	if err := recorder.StartObjective(HarnessObservationObjectiveSpec{ID: "relevance:0:scenario.json", Kind: "registered_relevance",
		Identity: HarnessObservationIdentity{RegistryIndex: &index, ScenarioPath: "scenario.json", EconomyCatalogPath: "economy.json",
			RelevancePolicyPath: "policy.json", GoldenReportPath: "golden.json", ScenarioHash: observationTestHash,
			RelevancePolicyHash: observationTestHash, ConstantsHash: observationTestHash, Active: &active}}); err != nil {
		t.Fatal(err)
	}
	if err := recorder.Fail("signal", errors.New("terminated")); err != nil {
		t.Fatal(err)
	}
	artifact, err := LoadHarnessObservation(path)
	if err != nil {
		t.Fatal(err)
	}
	if artifact.State != ObservationStateIncomplete || artifact.Termination == nil || *artifact.Termination != "signal" ||
		ValidateCompleteHarnessObservation(artifact) == nil {
		t.Fatalf("failed artifact=%+v", artifact)
	}
}

func validHarnessObservationFixture() HarnessObservation {
	finished, termination := "2026-08-21T10:00:01Z", "objective"
	epochID, index, active := int64(8), 0, true
	runs, transitions := int64(2), int64(9)
	return HarnessObservation{SchemaVersion: 1, Kind: "harness_observation.v1", Authoritative: false,
		Mode: "relevance-registered", State: ObservationStateComplete, Termination: &termination,
		StartedAt: "2026-08-21T10:00:00Z", UpdatedAt: finished, FinishedAt: &finished,
		ActiveEpochID: &epochID, ActiveConstantsHash: observationTestHash,
		DeclaredObjectiveIDs: []string{"relevance:0:scenario.json"}, Errors: []string{},
		Objectives: []HarnessObservationObjective{{ID: "relevance:0:scenario.json", Kind: "registered_relevance",
			State: ObservationStateComplete, StartedAt: "2026-08-21T10:00:00Z", UpdatedAt: finished, FinishedAt: &finished,
			Identity: HarnessObservationIdentity{RegistryIndex: &index, ScenarioPath: "scenario.json", EconomyCatalogPath: "economy.json",
				RelevancePolicyPath: "policy.json", GoldenReportPath: "golden.json", ScenarioHash: observationTestHash,
				RelevancePolicyHash: observationTestHash, ConstantsHash: observationTestHash, Active: &active},
			Work: HarnessObservationWork{DeclaredRuns: &runs, ExecutedRuns: &runs, DeclaredTransitions: &transitions,
				ExecutedTransitions: &transitions}, GuardState: ObservationConditionClear,
			PopulationExclusions: ObservationConditionClear, TruncationState: ObservationConditionClear,
			InstrumentExcluded: []string{}, Errors: []string{}}}}
}

func copiedInt64(value *int64) *int64 {
	if value == nil {
		return nil
	}
	copy := *value
	return &copy
}

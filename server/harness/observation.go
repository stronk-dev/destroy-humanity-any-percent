package harness

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const HarnessObservationSchemaVersion = 1

const (
	ObservationStateRunning    = "running"
	ObservationStateComplete   = "complete"
	ObservationStateIncomplete = "incomplete"

	ObservationConditionUnknown = "unknown"
	ObservationConditionClear   = "clear"
	ObservationConditionFired   = "fired"
)

type HarnessObservationWork struct {
	DeclaredRuns        *int64 `json:"declared_runs"`
	ExecutedRuns        *int64 `json:"executed_runs"`
	DeclaredTransitions *int64 `json:"declared_transitions"`
	ExecutedTransitions *int64 `json:"executed_transitions"`
}

type HarnessObservationIdentity struct {
	RegistryIndex       *int   `json:"registry_index"`
	ScenarioPath        string `json:"scenario_path"`
	EconomyCatalogPath  string `json:"economy_catalog_path"`
	RelevancePolicyPath string `json:"relevance_policy_path"`
	GoldenReportPath    string `json:"golden_report_path"`
	ScenarioHash        string `json:"scenario_hash"`
	RelevancePolicyHash string `json:"relevance_policy_hash"`
	ConstantsHash       string `json:"constants_hash"`
	Active              *bool  `json:"active"`
}

type HarnessObservationObjective struct {
	ID                   string                     `json:"id"`
	Kind                 string                     `json:"kind"`
	State                string                     `json:"state"`
	StartedAt            string                     `json:"started_at"`
	UpdatedAt            string                     `json:"updated_at"`
	FinishedAt           *string                    `json:"finished_at"`
	ElapsedMS            int64                      `json:"elapsed_ms"`
	Identity             HarnessObservationIdentity `json:"identity"`
	Work                 HarnessObservationWork     `json:"work"`
	GuardState           string                     `json:"guard_state"`
	PopulationExclusions string                     `json:"population_exclusions"`
	TruncationState      string                     `json:"truncation_state"`
	InstrumentExcluded   []string                   `json:"instrument_excluded_ids"`
	Errors               []string                   `json:"errors"`
}

type HarnessObservation struct {
	SchemaVersion        int                           `json:"schema_version"`
	Kind                 string                        `json:"kind"`
	Authoritative        bool                          `json:"authoritative"`
	Mode                 string                        `json:"mode"`
	State                string                        `json:"state"`
	Termination          *string                       `json:"termination"`
	StartedAt            string                        `json:"started_at"`
	UpdatedAt            string                        `json:"updated_at"`
	FinishedAt           *string                       `json:"finished_at"`
	ElapsedMS            int64                         `json:"elapsed_ms"`
	CurrentObjective     *string                       `json:"current_objective"`
	ActiveEpochID        *int64                        `json:"active_epoch_id"`
	ActiveConstantsHash  string                        `json:"active_constants_hash"`
	DeclaredObjectiveIDs []string                      `json:"declared_objective_ids"`
	Objectives           []HarnessObservationObjective `json:"objectives"`
	Errors               []string                      `json:"errors"`
}

type HarnessObservationObjectiveSpec struct {
	ID       string
	Kind     string
	Identity HarnessObservationIdentity
	Work     HarnessObservationWork
}

type HarnessObservationProgress struct {
	Work                 HarnessObservationWork
	GuardState           string
	PopulationExclusions string
	TruncationState      string
	InstrumentExcluded   []string
}

type HarnessObservationRecorder struct {
	mu               sync.Mutex
	path             string
	started          time.Time
	now              func() time.Time
	artifact         HarnessObservation
	objectiveStarted map[string]time.Time
}

func NewHarnessObservationRecorder(path, mode string, activeEpochID *int64, activeConstantsHash string) (*HarnessObservationRecorder, error) {
	return newHarnessObservationRecorder(path, mode, activeEpochID, activeConstantsHash, time.Now)
}

func newHarnessObservationRecorder(path, mode string, activeEpochID *int64, activeConstantsHash string, now func() time.Time) (*HarnessObservationRecorder, error) {
	if path == "" || mode == "" || now == nil {
		return nil, errors.New("invalid harness observation recorder")
	}
	started := now()
	stamp := observationTimestamp(started)
	recorder := &HarnessObservationRecorder{path: path, started: started, now: now, objectiveStarted: map[string]time.Time{},
		artifact: HarnessObservation{SchemaVersion: HarnessObservationSchemaVersion,
			Kind: "harness_observation.v1", Authoritative: false, Mode: mode,
			State: ObservationStateRunning, StartedAt: stamp, UpdatedAt: stamp,
			ActiveEpochID: activeEpochID, ActiveConstantsHash: activeConstantsHash,
			DeclaredObjectiveIDs: []string{}, Objectives: []HarnessObservationObjective{}, Errors: []string{}}}
	if err := recorder.writeLocked(); err != nil {
		return nil, err
	}
	return recorder, nil
}

func (recorder *HarnessObservationRecorder) DeclareObjectives(ids []string) error {
	recorder.mu.Lock()
	defer recorder.mu.Unlock()
	if recorder.artifact.State != ObservationStateRunning || len(recorder.artifact.Objectives) != 0 ||
		len(recorder.artifact.DeclaredObjectiveIDs) != 0 || len(ids) == 0 {
		return errors.New("invalid harness observation objective declaration")
	}
	seen := map[string]bool{}
	for _, id := range ids {
		if id == "" || seen[id] {
			return errors.New("invalid harness observation objective declaration")
		}
		seen[id] = true
	}
	recorder.artifact.DeclaredObjectiveIDs = append([]string(nil), ids...)
	recorder.touchLocked(recorder.now())
	return recorder.writeLocked()
}

func (recorder *HarnessObservationRecorder) StartObjective(spec HarnessObservationObjectiveSpec) error {
	recorder.mu.Lock()
	defer recorder.mu.Unlock()
	if recorder.artifact.State != ObservationStateRunning || recorder.artifact.CurrentObjective != nil || spec.ID == "" || spec.Kind == "" {
		return errors.New("invalid harness observation objective start")
	}
	next := len(recorder.artifact.Objectives)
	if next >= len(recorder.artifact.DeclaredObjectiveIDs) || recorder.artifact.DeclaredObjectiveIDs[next] != spec.ID {
		return errors.New("undeclared or out-of-order harness observation objective")
	}
	for _, objective := range recorder.artifact.Objectives {
		if objective.ID == spec.ID {
			return fmt.Errorf("duplicate harness observation objective %q", spec.ID)
		}
	}
	now := recorder.now()
	stamp := observationTimestamp(now)
	recorder.artifact.Objectives = append(recorder.artifact.Objectives, HarnessObservationObjective{
		ID: spec.ID, Kind: spec.Kind, State: ObservationStateRunning, StartedAt: stamp, UpdatedAt: stamp,
		Identity: spec.Identity, Work: spec.Work, GuardState: ObservationConditionUnknown,
		PopulationExclusions: ObservationConditionUnknown, TruncationState: ObservationConditionUnknown,
		InstrumentExcluded: []string{}, Errors: []string{},
	})
	recorder.objectiveStarted[spec.ID] = now
	recorder.artifact.CurrentObjective = stringPointer(spec.ID)
	recorder.touchLocked(now)
	return recorder.writeLocked()
}

func (recorder *HarnessObservationRecorder) Progress(progress HarnessObservationProgress) error {
	recorder.mu.Lock()
	defer recorder.mu.Unlock()
	index, err := recorder.currentObjectiveIndexLocked()
	if err != nil {
		return err
	}
	objective := &recorder.artifact.Objectives[index]
	mergeObservationWork(&objective.Work, progress.Work)
	if progress.GuardState != "" {
		objective.GuardState = progress.GuardState
	}
	if progress.PopulationExclusions != "" {
		objective.PopulationExclusions = progress.PopulationExclusions
	}
	if progress.TruncationState != "" {
		objective.TruncationState = progress.TruncationState
	}
	if progress.InstrumentExcluded != nil {
		objective.InstrumentExcluded = append([]string(nil), progress.InstrumentExcluded...)
	}
	now := recorder.now()
	objective.UpdatedAt = observationTimestamp(now)
	objective.ElapsedMS = observationElapsedMS(recorder.objectiveStarted[objective.ID], now)
	recorder.touchLocked(now)
	return recorder.writeLocked()
}

func (recorder *HarnessObservationRecorder) CompleteObjective(progress HarnessObservationProgress) error {
	recorder.mu.Lock()
	defer recorder.mu.Unlock()
	index, err := recorder.currentObjectiveIndexLocked()
	if err != nil {
		return err
	}
	objective := &recorder.artifact.Objectives[index]
	mergeObservationWork(&objective.Work, progress.Work)
	if progress.GuardState != "" {
		objective.GuardState = progress.GuardState
	}
	if progress.PopulationExclusions != "" {
		objective.PopulationExclusions = progress.PopulationExclusions
	}
	if progress.TruncationState != "" {
		objective.TruncationState = progress.TruncationState
	}
	if progress.InstrumentExcluded != nil {
		objective.InstrumentExcluded = append([]string(nil), progress.InstrumentExcluded...)
	}
	now := recorder.now()
	stamp := observationTimestamp(now)
	objective.State = ObservationStateComplete
	objective.UpdatedAt = stamp
	objective.FinishedAt = stringPointer(stamp)
	objective.ElapsedMS = observationElapsedMS(recorder.objectiveStarted[objective.ID], now)
	delete(recorder.objectiveStarted, objective.ID)
	recorder.artifact.CurrentObjective = nil
	recorder.touchLocked(now)
	return recorder.writeLocked()
}

func (recorder *HarnessObservationRecorder) Fail(termination string, cause error) error {
	recorder.mu.Lock()
	defer recorder.mu.Unlock()
	if recorder.artifact.State == ObservationStateComplete {
		return errors.New("cannot fail completed harness observation")
	}
	if termination != "error" && termination != "signal" && termination != "interrupted" {
		return errors.New("invalid harness observation termination")
	}
	now := recorder.now()
	stamp := observationTimestamp(now)
	if recorder.artifact.CurrentObjective != nil {
		for index := range recorder.artifact.Objectives {
			objective := &recorder.artifact.Objectives[index]
			if objective.ID == *recorder.artifact.CurrentObjective {
				objective.State = ObservationStateIncomplete
				objective.UpdatedAt = stamp
				objective.FinishedAt = stringPointer(stamp)
				objective.ElapsedMS = observationElapsedMS(recorder.objectiveStarted[objective.ID], now)
				delete(recorder.objectiveStarted, objective.ID)
				if cause != nil {
					objective.Errors = append(objective.Errors, cause.Error())
				}
				break
			}
		}
	}
	recorder.artifact.State = ObservationStateIncomplete
	recorder.artifact.Termination = stringPointer(termination)
	recorder.artifact.FinishedAt = stringPointer(stamp)
	recorder.artifact.CurrentObjective = nil
	if cause != nil {
		recorder.artifact.Errors = append(recorder.artifact.Errors, cause.Error())
	}
	recorder.touchLocked(now)
	return recorder.writeLocked()
}

func (recorder *HarnessObservationRecorder) Complete() error {
	recorder.mu.Lock()
	defer recorder.mu.Unlock()
	if recorder.artifact.State != ObservationStateRunning || recorder.artifact.CurrentObjective != nil {
		return errors.New("cannot complete harness observation with running objective")
	}
	now := recorder.now()
	stamp := observationTimestamp(now)
	recorder.artifact.State = ObservationStateComplete
	recorder.artifact.Termination = stringPointer("objective")
	recorder.artifact.FinishedAt = stringPointer(stamp)
	recorder.touchLocked(now)
	if err := ValidateCompleteHarnessObservation(recorder.artifact); err != nil {
		recorder.artifact.State = ObservationStateIncomplete
		recorder.artifact.Termination = stringPointer("error")
		recorder.artifact.Errors = append(recorder.artifact.Errors, err.Error())
		if writeErr := recorder.writeLocked(); writeErr != nil {
			return errors.Join(err, writeErr)
		}
		return err
	}
	return recorder.writeLocked()
}

func LoadHarnessObservation(path string) (HarnessObservation, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return HarnessObservation{}, err
	}
	var observation HarnessObservation
	decoder := json.NewDecoder(bytesReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&observation); err != nil {
		return HarnessObservation{}, err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return HarnessObservation{}, errors.New("harness observation must contain exactly one JSON value")
	}
	return observation, nil
}

func ValidateCompleteHarnessObservation(observation HarnessObservation) error {
	if observation.SchemaVersion != HarnessObservationSchemaVersion || observation.Kind != "harness_observation.v1" ||
		observation.Authoritative || observation.Mode != "check" && observation.Mode != "relevance-registered" ||
		observation.State != ObservationStateComplete ||
		observation.Termination == nil || *observation.Termination != "objective" || observation.StartedAt == "" ||
		observation.UpdatedAt == "" || observation.FinishedAt == nil || observation.CurrentObjective != nil ||
		observation.ElapsedMS < 0 || !validObservationTimestamp(observation.StartedAt) ||
		!validObservationTimestamp(observation.UpdatedAt) || !validObservationTimestamp(*observation.FinishedAt) ||
		observation.ActiveEpochID == nil || *observation.ActiveEpochID < 1 || !relevanceHashPattern.MatchString(observation.ActiveConstantsHash) ||
		observation.DeclaredObjectiveIDs == nil || len(observation.DeclaredObjectiveIDs) == 0 ||
		observation.Objectives == nil || len(observation.Objectives) != len(observation.DeclaredObjectiveIDs) ||
		observation.Errors == nil || len(observation.Errors) != 0 {
		return errors.New("incomplete harness observation")
	}
	seen := map[string]bool{}
	activeRelevanceObjectives := 0
	for index, objective := range observation.Objectives {
		if objective.ID != observation.DeclaredObjectiveIDs[index] {
			return errors.New("harness observation objective declaration mismatch")
		}
		if objective.ID == "" || objective.Kind == "" || seen[objective.ID] || objective.State != ObservationStateComplete ||
			objective.StartedAt == "" || objective.UpdatedAt == "" || objective.ElapsedMS < 0 ||
			!validObservationTimestamp(objective.StartedAt) || !validObservationTimestamp(objective.UpdatedAt) ||
			objective.FinishedAt == nil || objective.Errors == nil || len(objective.Errors) != 0 ||
			!validObservationTimestamp(*objective.FinishedAt) || objective.InstrumentExcluded == nil ||
			!sortedUniqueObservationStrings(objective.InstrumentExcluded) ||
			objective.GuardState != ObservationConditionClear || objective.PopulationExclusions != ObservationConditionClear ||
			objective.TruncationState != ObservationConditionClear {
			return fmt.Errorf("incomplete harness observation objective %q", objective.ID)
		}
		seen[objective.ID] = true
		if err := validateObservationIdentity(objective, observation.ActiveConstantsHash); err != nil {
			return fmt.Errorf("harness observation objective %q: %w", objective.ID, err)
		}
		if err := validateObservationWork(objective.Work); err != nil {
			return fmt.Errorf("harness observation objective %q: %w", objective.ID, err)
		}
		if observation.Mode == "check" {
			if index == 0 && objective.Kind != "repository_guards" || index == 1 && objective.Kind != "standard_pacing" ||
				index > 1 && objective.Kind != "registered_relevance" {
				return errors.New("invalid complete-check objective order")
			}
			if index > 1 && (objective.Identity.RegistryIndex == nil || *objective.Identity.RegistryIndex != index-2) {
				return errors.New("invalid complete-check registry order")
			}
		}
		if objective.Kind == "registered_relevance" && objective.Identity.Active != nil && *objective.Identity.Active {
			activeRelevanceObjectives++
		}
	}
	if observation.Mode == "check" && (len(observation.Objectives) < 3 || activeRelevanceObjectives != 1) {
		return errors.New("complete check has missing or invalid objectives")
	}
	if observation.Mode == "relevance-registered" && (len(observation.Objectives) != 1 || observation.Objectives[0].Kind != "registered_relevance") {
		return errors.New("registered relevance observation has invalid objectives")
	}
	return nil
}

func validObservationTimestamp(value string) bool {
	_, err := time.Parse(time.RFC3339Nano, value)
	return err == nil
}

func sortedUniqueObservationStrings(values []string) bool {
	for index, value := range values {
		if value == "" || index > 0 && values[index-1] >= value {
			return false
		}
	}
	return true
}

func validateObservationIdentity(objective HarnessObservationObjective, activeConstantsHash string) error {
	switch objective.Kind {
	case "repository_guards":
		return nil
	case "standard_pacing":
		if objective.Identity.ScenarioPath == "" || !relevanceHashPattern.MatchString(objective.Identity.ScenarioHash) ||
			objective.Identity.ConstantsHash != activeConstantsHash {
			return errors.New("invalid standard pacing identity")
		}
	case "registered_relevance":
		identity := objective.Identity
		if identity.RegistryIndex == nil || *identity.RegistryIndex < 0 || identity.ScenarioPath == "" ||
			identity.EconomyCatalogPath == "" || identity.RelevancePolicyPath == "" || identity.GoldenReportPath == "" ||
			!relevanceHashPattern.MatchString(identity.ScenarioHash) || !relevanceHashPattern.MatchString(identity.RelevancePolicyHash) ||
			!relevanceHashPattern.MatchString(identity.ConstantsHash) || identity.Active == nil || *identity.Active && identity.ConstantsHash != activeConstantsHash {
			return errors.New("invalid registered relevance identity")
		}
	default:
		return errors.New("unknown harness observation objective kind")
	}
	return nil
}

func validateObservationWork(work HarnessObservationWork) error {
	for _, value := range []*int64{work.DeclaredRuns, work.ExecutedRuns, work.DeclaredTransitions, work.ExecutedTransitions} {
		if value != nil && *value < 0 {
			return errors.New("negative work cardinality")
		}
	}
	if (work.DeclaredRuns == nil) != (work.ExecutedRuns == nil) ||
		(work.DeclaredTransitions == nil) != (work.ExecutedTransitions == nil) {
		return errors.New("partial work cardinality")
	}
	if work.DeclaredRuns != nil && *work.DeclaredRuns != *work.ExecutedRuns {
		return errors.New("run cardinality mismatch")
	}
	if work.DeclaredTransitions != nil && *work.DeclaredTransitions != *work.ExecutedTransitions {
		return errors.New("transition cardinality mismatch")
	}
	return nil
}

func (recorder *HarnessObservationRecorder) currentObjectiveIndexLocked() (int, error) {
	if recorder.artifact.State != ObservationStateRunning || recorder.artifact.CurrentObjective == nil {
		return 0, errors.New("no running harness observation objective")
	}
	for index := range recorder.artifact.Objectives {
		if recorder.artifact.Objectives[index].ID == *recorder.artifact.CurrentObjective {
			return index, nil
		}
	}
	return 0, errors.New("running harness observation objective is missing")
}

func (recorder *HarnessObservationRecorder) touchLocked(now time.Time) {
	recorder.artifact.UpdatedAt = observationTimestamp(now)
	recorder.artifact.ElapsedMS = observationElapsedMS(recorder.started, now)
}

func (recorder *HarnessObservationRecorder) writeLocked() error {
	data, err := CanonicalJSON(recorder.artifact)
	if err != nil {
		return err
	}
	directory := filepath.Dir(recorder.path)
	file, err := os.CreateTemp(directory, ".harness-observation-*")
	if err != nil {
		return err
	}
	temporary := file.Name()
	remove := true
	defer func() {
		_ = file.Close()
		if remove {
			_ = os.Remove(temporary)
		}
	}()
	if err := file.Chmod(0o644); err != nil {
		return err
	}
	if _, err := file.Write(data); err != nil {
		return err
	}
	if err := file.Sync(); err != nil {
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	if err := os.Rename(temporary, recorder.path); err != nil {
		return err
	}
	remove = false
	directoryHandle, err := os.Open(directory)
	if err != nil {
		return err
	}
	defer directoryHandle.Close()
	return directoryHandle.Sync()
}

func mergeObservationWork(current *HarnessObservationWork, update HarnessObservationWork) {
	if update.DeclaredRuns != nil {
		current.DeclaredRuns = int64Pointer(*update.DeclaredRuns)
	}
	if update.ExecutedRuns != nil {
		current.ExecutedRuns = int64Pointer(*update.ExecutedRuns)
	}
	if update.DeclaredTransitions != nil {
		current.DeclaredTransitions = int64Pointer(*update.DeclaredTransitions)
	}
	if update.ExecutedTransitions != nil {
		current.ExecutedTransitions = int64Pointer(*update.ExecutedTransitions)
	}
}

func observationTimestamp(value time.Time) string {
	return value.UTC().Format(time.RFC3339Nano)
}

func observationElapsedMS(started, now time.Time) int64 {
	elapsed := now.Sub(started).Milliseconds()
	if elapsed < 0 {
		return 0
	}
	return elapsed
}

func stringPointer(value string) *string { return &value }
func int64Pointer(value int64) *int64    { return &value }

// bytesReader is kept here so strict decoding has one obvious exact-value path.
func bytesReader(data []byte) *bytes.Reader { return bytes.NewReader(data) }

package deploymentrehearsal

import (
	"bytes"
	"encoding/json"
	"io"
	"time"
)

type HostObservation struct {
	SchemaVersion      int       `json:"schema_version"`
	StartedAt          time.Time `json:"started_at"`
	CompletedAt        time.Time `json:"completed_at"`
	Host               Host      `json:"host"`
	ObjectiveCompleted bool      `json:"objective_completed"`
	GuardExhausted     bool      `json:"guard_exhausted"`
}

type ObjectiveObservation struct {
	SchemaVersion      int        `json:"schema_version"`
	StartedAt          time.Time  `json:"started_at"`
	CompletedAt        time.Time  `json:"completed_at"`
	Objectives         Objectives `json:"objectives"`
	ObjectiveCompleted bool       `json:"objective_completed"`
	GuardExhausted     bool       `json:"guard_exhausted"`
}

func DecodeHostObservation(data []byte) (HostObservation, error) {
	var observation HostObservation
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&observation) != nil || decoder.Decode(&struct{}{}) != io.EOF || ValidateHostObservation(observation) != nil {
		return HostObservation{}, ErrInvalid
	}
	return observation, nil
}

func ValidateHostObservation(observation HostObservation) error {
	if observation.SchemaVersion != 1 || observation.StartedAt.IsZero() || !observation.CompletedAt.After(observation.StartedAt) ||
		!observation.ObjectiveCompleted || observation.GuardExhausted || validateHost(observation.Host) != nil {
		return ErrInvalid
	}
	return nil
}

func WriteHostObservation(path string, observation HostObservation) error {
	if ValidateHostObservation(observation) != nil {
		return ErrInvalid
	}
	data, err := json.Marshal(observation)
	if err != nil {
		return err
	}
	return writeNewEvidence(path, append(data, '\n'))
}

func DecodeObjectiveObservation(data []byte) (ObjectiveObservation, error) {
	var observation ObjectiveObservation
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&observation) != nil || decoder.Decode(&struct{}{}) != io.EOF || ValidateObjectiveObservation(observation) != nil {
		return ObjectiveObservation{}, ErrInvalid
	}
	return observation, nil
}

func ValidateObjectiveObservation(observation ObjectiveObservation) error {
	if observation.SchemaVersion != 1 || observation.StartedAt.IsZero() || !observation.CompletedAt.After(observation.StartedAt) ||
		!observation.ObjectiveCompleted || observation.GuardExhausted ||
		validateObjectives(observation.Objectives, observation.StartedAt, observation.CompletedAt) != nil {
		return ErrInvalid
	}
	return nil
}

func WriteObjectiveObservation(path string, observation ObjectiveObservation) error {
	if ValidateObjectiveObservation(observation) != nil {
		return ErrInvalid
	}
	data, err := json.Marshal(observation)
	if err != nil {
		return err
	}
	return writeNewEvidence(path, append(data, '\n'))
}

func sameObjectives(left, right Objectives) bool {
	return left.IncidentAt.Equal(right.IncidentAt) && left.NewestValidBackupAt.Equal(right.NewestValidBackupAt) &&
		left.RestoreStartedAt.Equal(right.RestoreStartedAt) && left.AuthenticatedSmokeAt.Equal(right.AuthenticatedSmokeAt) &&
		left.RPOSeconds == right.RPOSeconds && left.RTOSeconds == right.RTOSeconds && left.RestoredIdentityMatch == right.RestoredIdentityMatch
}

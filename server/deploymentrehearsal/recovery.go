package deploymentrehearsal

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"time"

	"cloud-clicker/server/deploymentbackup"
	"cloud-clicker/server/deploymentrelease"
)

const recoveryCheckpointName = "recovery-checkpoint.json"

type recoveryBackup struct {
	ID     string                  `json:"id"`
	Path   string                  `json:"path"`
	Header deploymentbackup.Header `json:"header"`
}

type RecoveryCheckpoint struct {
	SchemaVersion     int                               `json:"schema_version"`
	ManifestSHA256    string                            `json:"manifest_sha256"`
	IncidentAt        time.Time                         `json:"incident_at"`
	PopulatedBackup   recoveryBackup                    `json:"populated_backup"`
	PopulatedIdentity deploymentbackup.RecoveryIdentity `json:"populated_identity"`
	EmptyBackup       recoveryBackup                    `json:"empty_backup"`
	EmptyIdentity     deploymentbackup.RecoveryIdentity `json:"empty_identity"`
}

type recoveryRuntime interface {
	InspectRecoveryIdentity(context.Context, deploymentrelease.Bundle) (deploymentbackup.RecoveryIdentity, error)
	CreateRecoveryBackup(context.Context, deploymentrelease.Bundle) (deploymentrelease.BackupReference, error)
	StopFailed(context.Context, deploymentrelease.Bundle) error
	ResetDatabase(context.Context, deploymentrelease.Bundle) error
	StartRecoveryCore(context.Context, deploymentrelease.Bundle) error
	VerifyIdentity(context.Context, deploymentrelease.Bundle) error
	RestoreRecoveryBackup(context.Context, deploymentrelease.Bundle, deploymentrelease.BackupReference) error
	AuthenticatedSmoke(context.Context, deploymentrelease.Bundle) error
}

type recoveryDependencies struct {
	loadBundle func(string) (deploymentrelease.Bundle, error)
	readHeader func(string) (deploymentbackup.Header, error)
	runtime    recoveryRuntime
	now        func() time.Time
}

func RunEmptyRecovery(ctx context.Context, config ScenarioConfig) (RecoveryCheckpoint, error) {
	runtime := deploymentrelease.DockerRuntime{PublicOrigin: config.PublicOrigin, ReceiverHealthURL: config.ReceiverHealthURL,
		BackupTarget: config.BackupTarget, MetricsDirectory: config.MetricsDirectory, AgeRecipient: config.AgeRecipient,
		AgeIdentityFile: config.AgeIdentityFile, ServerID: config.ServerID, DrainTimeout: 20 * time.Second}
	return runEmptyRecovery(ctx, config, recoveryDependencies{loadBundle: deploymentrelease.LoadBundle,
		readHeader: deploymentbackup.ReadHeader, runtime: runtime, now: time.Now})
}

func RunPopulatedRecovery(ctx context.Context, config ScenarioConfig) (ObjectiveObservation, error) {
	runtime := deploymentrelease.DockerRuntime{PublicOrigin: config.PublicOrigin, ReceiverHealthURL: config.ReceiverHealthURL,
		BackupTarget: config.BackupTarget, MetricsDirectory: config.MetricsDirectory, AgeRecipient: config.AgeRecipient,
		AgeIdentityFile: config.AgeIdentityFile, ServerID: config.ServerID, DrainTimeout: 20 * time.Second}
	return runPopulatedRecovery(ctx, config, recoveryDependencies{loadBundle: deploymentrelease.LoadBundle,
		readHeader: deploymentbackup.ReadHeader, runtime: runtime, now: time.Now})
}

func runEmptyRecovery(ctx context.Context, config ScenarioConfig, dependencies recoveryDependencies) (RecoveryCheckpoint, error) {
	bundle, err := validateRecoveryDependencies(config, dependencies)
	if err != nil {
		return RecoveryCheckpoint{}, err
	}
	checkpointPath := filepath.Join(config.WorkDirectory, recoveryCheckpointName)
	if _, err := os.Lstat(checkpointPath); !errors.Is(err, os.ErrNotExist) {
		return RecoveryCheckpoint{}, ErrInvalid
	}
	populated, err := dependencies.runtime.InspectRecoveryIdentity(ctx, bundle)
	if err != nil || deploymentbackup.ValidatePopulatedRecoveryIdentity(populated) != nil {
		return RecoveryCheckpoint{}, errors.Join(ErrInvalid, err)
	}
	populatedBackup, err := captureRecoveryBackup(ctx, config, bundle, dependencies)
	if err != nil {
		return RecoveryCheckpoint{}, err
	}
	incidentAt := dependencies.now().UTC()
	if incidentAt.Before(populatedBackup.Header.CompletedAt) {
		return RecoveryCheckpoint{}, ErrInvalid
	}
	if err := resetRecoveryDatabase(ctx, dependencies.runtime, bundle); err != nil {
		return RecoveryCheckpoint{}, err
	}
	empty, err := dependencies.runtime.InspectRecoveryIdentity(ctx, bundle)
	if err != nil || deploymentbackup.ValidateEmptyRecoveryIdentity(empty) != nil {
		return RecoveryCheckpoint{}, errors.Join(ErrInvalid, err)
	}
	emptyBackup, err := captureRecoveryBackup(ctx, config, bundle, dependencies)
	if err != nil {
		return RecoveryCheckpoint{}, err
	}
	if err := dependencies.runtime.StopFailed(ctx, bundle); err != nil {
		return RecoveryCheckpoint{}, err
	}
	if err := dependencies.runtime.ResetDatabase(ctx, bundle); err != nil {
		return RecoveryCheckpoint{}, err
	}
	if err := dependencies.runtime.RestoreRecoveryBackup(ctx, bundle, deploymentrelease.BackupReference{ID: emptyBackup.ID, Path: emptyBackup.Path}); err != nil {
		return RecoveryCheckpoint{}, err
	}
	if err := dependencies.runtime.StartRecoveryCore(ctx, bundle); err != nil {
		return RecoveryCheckpoint{}, err
	}
	if err := dependencies.runtime.VerifyIdentity(ctx, bundle); err != nil {
		return RecoveryCheckpoint{}, err
	}
	restoredEmpty, err := dependencies.runtime.InspectRecoveryIdentity(ctx, bundle)
	if err != nil || deploymentbackup.CompareRecoveryIdentity(empty, restoredEmpty) != nil {
		return RecoveryCheckpoint{}, errors.Join(ErrInvalid, err)
	}
	checkpoint := RecoveryCheckpoint{SchemaVersion: 1, ManifestSHA256: bundle.ManifestSHA256, IncidentAt: incidentAt,
		PopulatedBackup: populatedBackup, PopulatedIdentity: populated, EmptyBackup: emptyBackup, EmptyIdentity: empty}
	if validateRecoveryCheckpoint(checkpoint, config, bundle) != nil {
		return RecoveryCheckpoint{}, ErrInvalid
	}
	data, _ := json.Marshal(checkpoint)
	if err := writeNewEvidence(checkpointPath, append(data, '\n')); err != nil {
		return RecoveryCheckpoint{}, err
	}
	return checkpoint, nil
}

func runPopulatedRecovery(ctx context.Context, config ScenarioConfig, dependencies recoveryDependencies) (ObjectiveObservation, error) {
	bundle, err := validateRecoveryDependencies(config, dependencies)
	if err != nil {
		return ObjectiveObservation{}, err
	}
	objectivePath := filepath.Join(config.ArtifactsDirectory, requiredRunArtifactFiles["objective_observation"])
	if _, err := os.Lstat(objectivePath); !errors.Is(err, os.ErrNotExist) {
		return ObjectiveObservation{}, ErrInvalid
	}
	checkpoint, err := loadRecoveryCheckpoint(filepath.Join(config.WorkDirectory, recoveryCheckpointName))
	if err != nil || validateRecoveryCheckpoint(checkpoint, config, bundle) != nil {
		return ObjectiveObservation{}, ErrInvalid
	}
	if err := dependencies.runtime.StopFailed(ctx, bundle); err != nil {
		return ObjectiveObservation{}, err
	}
	if err := dependencies.runtime.ResetDatabase(ctx, bundle); err != nil {
		return ObjectiveObservation{}, err
	}
	restoreStarted := dependencies.now().UTC()
	reference := deploymentrelease.BackupReference{ID: checkpoint.PopulatedBackup.ID, Path: checkpoint.PopulatedBackup.Path}
	if err := dependencies.runtime.RestoreRecoveryBackup(ctx, bundle, reference); err != nil {
		return ObjectiveObservation{}, err
	}
	if err := dependencies.runtime.StartRecoveryCore(ctx, bundle); err != nil {
		return ObjectiveObservation{}, err
	}
	if err := dependencies.runtime.VerifyIdentity(ctx, bundle); err != nil {
		return ObjectiveObservation{}, err
	}
	restored, err := dependencies.runtime.InspectRecoveryIdentity(ctx, bundle)
	if err != nil || deploymentbackup.CompareRecoveryIdentity(checkpoint.PopulatedIdentity, restored) != nil {
		return ObjectiveObservation{}, errors.Join(ErrInvalid, err)
	}
	if err := dependencies.runtime.AuthenticatedSmoke(ctx, bundle); err != nil {
		return ObjectiveObservation{}, err
	}
	smokePassed := dependencies.now().UTC()
	measurement, err := deploymentbackup.MeasureObjectives(checkpoint.PopulatedBackup.Header, deploymentbackup.ObjectiveObservation{
		IncidentAt: checkpoint.IncidentAt, RestoreStartedAt: restoreStarted, SmokePassedAt: smokePassed})
	if err != nil {
		return ObjectiveObservation{}, err
	}
	observation := ObjectiveObservation{SchemaVersion: 1, StartedAt: checkpoint.IncidentAt, CompletedAt: smokePassed,
		Objectives: Objectives{IncidentAt: checkpoint.IncidentAt, NewestValidBackupAt: checkpoint.PopulatedBackup.Header.CompletedAt,
			RestoreStartedAt: restoreStarted, AuthenticatedSmokeAt: smokePassed, RPOSeconds: measurement.RPOSeconds,
			RTOSeconds: measurement.RTOSeconds, RestoredIdentityMatch: true}, ObjectiveCompleted: true}
	if err := WriteObjectiveObservation(objectivePath, observation); err != nil {
		return ObjectiveObservation{}, err
	}
	return observation, nil
}

func validateRecoveryDependencies(config ScenarioConfig, dependencies recoveryDependencies) (deploymentrelease.Bundle, error) {
	if ValidateScenarioConfig(config) != nil || validateScenarioFilesystem(config) != nil || dependencies.loadBundle == nil ||
		dependencies.readHeader == nil || dependencies.runtime == nil || dependencies.now == nil {
		return deploymentrelease.Bundle{}, ErrInvalid
	}
	bundle, err := dependencies.loadBundle(config.CandidateBundle)
	if err != nil || bundle.Root != config.CandidateBundle {
		return deploymentrelease.Bundle{}, errors.Join(ErrInvalid, err)
	}
	return bundle, nil
}

func resetRecoveryDatabase(ctx context.Context, runtime recoveryRuntime, bundle deploymentrelease.Bundle) error {
	if err := runtime.StopFailed(ctx, bundle); err != nil {
		return err
	}
	if err := runtime.ResetDatabase(ctx, bundle); err != nil {
		return err
	}
	if err := runtime.StartRecoveryCore(ctx, bundle); err != nil {
		return err
	}
	return runtime.VerifyIdentity(ctx, bundle)
}

func captureRecoveryBackup(ctx context.Context, config ScenarioConfig, bundle deploymentrelease.Bundle, dependencies recoveryDependencies) (recoveryBackup, error) {
	reference, err := dependencies.runtime.CreateRecoveryBackup(ctx, bundle)
	if err != nil {
		return recoveryBackup{}, err
	}
	header, err := dependencies.readHeader(reference.Path)
	result := recoveryBackup{ID: reference.ID, Path: reference.Path, Header: header}
	if err != nil || validateRecoveryBackup(result, config, bundle) != nil {
		return recoveryBackup{}, errors.Join(ErrInvalid, err)
	}
	return result, nil
}

func validateRecoveryBackup(backup recoveryBackup, config ScenarioConfig, bundle deploymentrelease.Bundle) error {
	headerBytes, _ := json.Marshal(backup.Header)
	decoded, headerErr := deploymentbackup.DecodeHeader(headerBytes)
	info, fileErr := os.Lstat(backup.Path)
	if backup.ID == "" || backup.Path != filepath.Join(config.BackupTarget, backup.ID+".ccbackup") || backup.Header.BackupID != backup.ID ||
		backup.Header.ReleaseManifestSHA256 != bundle.ManifestSHA256 || backup.Header.EpochID != bundle.Manifest.EpochID ||
		backup.Header.ServerID != config.ServerID || backup.Header.PreUpgrade || backup.Header.UpgradeResolved || headerErr != nil || decoded != backup.Header ||
		fileErr != nil || !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 || info.Size() < 1 {
		return ErrInvalid
	}
	return nil
}

func validateRecoveryCheckpoint(checkpoint RecoveryCheckpoint, config ScenarioConfig, bundle deploymentrelease.Bundle) error {
	if checkpoint.SchemaVersion != 1 || checkpoint.ManifestSHA256 != bundle.ManifestSHA256 || checkpoint.IncidentAt.IsZero() ||
		validateRecoveryBackup(checkpoint.PopulatedBackup, config, bundle) != nil || validateRecoveryBackup(checkpoint.EmptyBackup, config, bundle) != nil ||
		deploymentbackup.ValidatePopulatedRecoveryIdentity(checkpoint.PopulatedIdentity) != nil ||
		deploymentbackup.ValidateEmptyRecoveryIdentity(checkpoint.EmptyIdentity) != nil || checkpoint.IncidentAt.Before(checkpoint.PopulatedBackup.Header.CompletedAt) ||
		checkpoint.EmptyBackup.Header.StartedAt.Before(checkpoint.IncidentAt) {
		return ErrInvalid
	}
	return nil
}

func loadRecoveryCheckpoint(path string) (RecoveryCheckpoint, error) {
	info, err := os.Lstat(path)
	if err != nil || !info.Mode().IsRegular() || info.Mode().Perm() != 0o600 {
		return RecoveryCheckpoint{}, ErrInvalid
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return RecoveryCheckpoint{}, err
	}
	var checkpoint RecoveryCheckpoint
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&checkpoint) != nil || decoder.Decode(&struct{}{}) != io.EOF {
		return RecoveryCheckpoint{}, ErrInvalid
	}
	return checkpoint, nil
}

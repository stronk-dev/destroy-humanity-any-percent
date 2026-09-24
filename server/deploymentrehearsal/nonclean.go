package deploymentrehearsal

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"time"

	"cloud-clicker/server/deploymentbackup"
	"cloud-clicker/server/deploymentrelease"
	"cloud-clicker/server/releasepackage"
)

type nonCleanRuntime interface {
	CreateRecoveryBackup(context.Context, deploymentrelease.Bundle) (deploymentrelease.BackupReference, error)
	RestoreRecoveryBackup(context.Context, deploymentrelease.Bundle, deploymentrelease.BackupReference) error
	InspectRecoveryIdentity(context.Context, deploymentrelease.Bundle) (deploymentbackup.RecoveryIdentity, error)
}

type nonCleanDependencies struct {
	loadBundle func(string) (deploymentrelease.Bundle, error)
	runtime    func(ScenarioConfig) nonCleanRuntime
	refusal    func(error) bool
}

// ProbeNonCleanRestore produces non_clean_restore_target on the installed
// candidate stack. A recovery backup of the live, populated database must be
// creatable (the input is real); restoring it in place, without a clean
// volume, must then be refused by the backup tool's clean-target gate and
// leave the live database identity byte-for-byte unchanged. Any other failure
// is a setup error, never a rejection.
func ProbeNonCleanRestore(ctx context.Context, config ScenarioConfig) (ProbeOutcome, error) {
	return probeNonCleanRestore(ctx, config, nonCleanDependencies{loadBundle: deploymentrelease.LoadBundle,
		refusal: deploymentrelease.IsNonCleanRestoreRefusal,
		runtime: func(config ScenarioConfig) nonCleanRuntime {
			return deploymentrelease.DockerRuntime{PublicOrigin: config.PublicOrigin, ReceiverHealthURL: config.ReceiverHealthURL,
				BackupTarget: config.BackupTarget, MetricsDirectory: config.MetricsDirectory, AgeRecipient: config.AgeRecipient,
				AgeIdentityFile: config.AgeIdentityFile, ServerID: config.ServerID, DrainTimeout: 20 * time.Second,
				RotationLedgerPath: filepath.Join(config.InstallOperatorState, "rotation-ledger.jsonl")}
		}})
}

func probeNonCleanRestore(ctx context.Context, config ScenarioConfig, dependencies nonCleanDependencies) (ProbeOutcome, error) {
	if ValidateScenarioConfig(config) != nil || validateScenarioFilesystem(config) != nil || dependencies.loadBundle == nil ||
		dependencies.runtime == nil || dependencies.refusal == nil {
		return ProbeAccepted, ErrInvalid
	}
	candidate, err := dependencies.loadBundle(config.CandidateBundle)
	if err != nil {
		return ProbeAccepted, err
	}
	manifestBytes, err := os.ReadFile(filepath.Join(config.CandidateBundle, releasepackage.ReleaseManifestPath))
	if err != nil {
		return ProbeAccepted, err
	}
	authority, err := decodeManifestAuthority(manifestBytes)
	if err != nil {
		return ProbeAccepted, err
	}
	install, err := readLedger(filepath.Join(config.InstallOperatorState, "release-ledger.jsonl"))
	if err != nil || len(install) != 1 || !matchesInstall(install[0], authority, hashBytes(manifestBytes)) {
		return ProbeAccepted, errors.Join(ErrInvalid, err)
	}
	runtime := dependencies.runtime(config)
	before, err := runtime.InspectRecoveryIdentity(ctx, candidate)
	if err != nil {
		return ProbeAccepted, err
	}
	backup, err := runtime.CreateRecoveryBackup(ctx, candidate)
	if err != nil {
		return ProbeAccepted, err
	}
	defer os.Remove(backup.Path)
	restoreErr := runtime.RestoreRecoveryBackup(ctx, candidate, backup)
	after, err := runtime.InspectRecoveryIdentity(ctx, candidate)
	if err != nil {
		return ProbeAccepted, err
	}
	if restoreErr == nil {
		return ProbeAccepted, nil
	}
	if !dependencies.refusal(restoreErr) || after != before {
		return ProbeAccepted, errors.Join(ErrInvalid, restoreErr)
	}
	return ProbeRejected, nil
}

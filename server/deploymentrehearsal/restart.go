package deploymentrehearsal

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"time"

	"cloud-clicker/server/deploymentrelease"
	"cloud-clicker/server/releasepackage"
)

type restartRuntime interface {
	KillGameserver(context.Context, deploymentrelease.Bundle) (deploymentrelease.DrainEvidence, string, error)
	Start(context.Context, deploymentrelease.Bundle) error
	AuthenticatedSmoke(context.Context, deploymentrelease.Bundle) error
}

type restartDependencies struct {
	loadBundle func(string) (deploymentrelease.Bundle, error)
	runtime    func(ScenarioConfig) restartRuntime
}

// ProbeRestartDuringAdmittedWork produces gameserver_restart_during_admitted_work
// on the installed candidate stack. An ungraceful SIGKILL restart must be
// classified by the release drain derivation as not a bounded drain (invalid
// evidence, non-zero exit); the stack must then be restored to readiness and a
// passing authenticated smoke so later rehearsal steps start from service.
func ProbeRestartDuringAdmittedWork(ctx context.Context, config ScenarioConfig) (ProbeOutcome, error) {
	return probeRestartDuringAdmittedWork(ctx, config, restartDependencies{loadBundle: deploymentrelease.LoadBundle,
		runtime: func(config ScenarioConfig) restartRuntime {
			return deploymentrelease.DockerRuntime{PublicOrigin: config.PublicOrigin, ReceiverHealthURL: config.ReceiverHealthURL,
				BackupTarget: config.BackupTarget, MetricsDirectory: config.MetricsDirectory, AgeRecipient: config.AgeRecipient,
				ServerID: config.ServerID, DrainTimeout: 20 * time.Second,
				RotationLedgerPath: filepath.Join(config.InstallOperatorState, "rotation-ledger.jsonl")}
		}})
}

func probeRestartDuringAdmittedWork(ctx context.Context, config ScenarioConfig, dependencies restartDependencies) (ProbeOutcome, error) {
	if ValidateScenarioConfig(config) != nil || validateScenarioFilesystem(config) != nil || dependencies.loadBundle == nil || dependencies.runtime == nil {
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
	evidence, exitCode, killErr := runtime.KillGameserver(ctx, candidate)
	// Service is restored whatever the observation, so a failed probe never
	// leaves the rehearsal host down for its next step.
	restoreErr := errors.Join(runtime.Start(ctx, candidate), runtime.AuthenticatedSmoke(ctx, candidate))
	if killErr != nil || restoreErr != nil {
		return ProbeAccepted, errors.Join(ErrInvalid, killErr, restoreErr)
	}
	if evidence.Valid() || exitCode == "0" {
		return ProbeAccepted, nil
	}
	return ProbeRejected, nil
}

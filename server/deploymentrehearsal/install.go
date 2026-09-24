package deploymentrehearsal

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"cloud-clicker/server/deploymentrelease"
	"cloud-clicker/server/releasepackage"
)

type candidateInstallDependencies struct {
	validateBundle func(string) error
	execute        func(context.Context, ScenarioConfig) error
}

func InstallCandidate(ctx context.Context, config ScenarioConfig) error {
	return installCandidate(ctx, config, candidateInstallDependencies{validateBundle: releasepackage.ValidateBundle, execute: executeCandidateInstall})
}

func installCandidate(ctx context.Context, config ScenarioConfig, dependencies candidateInstallDependencies) error {
	if ValidateScenarioConfig(config) != nil || validateScenarioFilesystem(config) != nil || dependencies.validateBundle == nil || dependencies.execute == nil ||
		dependencies.validateBundle(config.CandidateBundle) != nil {
		return ErrInvalid
	}
	if err := dependencies.execute(ctx, config); err != nil {
		return err
	}
	manifestBytes, err := os.ReadFile(filepath.Join(config.CandidateBundle, releasepackage.ReleaseManifestPath))
	if err != nil {
		return err
	}
	manifest, err := decodeManifestAuthority(manifestBytes)
	if err != nil {
		return err
	}
	ledgerPath := filepath.Join(config.InstallOperatorState, "release-ledger.jsonl")
	ledgerBytes, err := os.ReadFile(ledgerPath)
	if err != nil {
		return err
	}
	records, err := deploymentrelease.DecodeReleaseLedger(ledgerBytes)
	if err != nil || len(records) != 1 || !matchesInstall(records[0], manifest, hashBytes(manifestBytes)) {
		return ErrInvalid
	}
	return writeNewEvidence(filepath.Join(config.ArtifactsDirectory, requiredRunArtifactFiles["install_ledger"]), ledgerBytes)
}

func executeCandidateInstall(ctx context.Context, config ScenarioConfig) error {
	runtime := deploymentrelease.DockerRuntime{PublicOrigin: config.PublicOrigin, ReceiverHealthURL: config.ReceiverHealthURL,
		BackupTarget: config.BackupTarget, MetricsDirectory: config.MetricsDirectory, AgeRecipient: config.AgeRecipient,
		ServerID: config.ServerID, DrainTimeout: 20 * time.Second,
		RotationLedgerPath: filepath.Join(config.InstallOperatorState, "rotation-ledger.jsonl")}
	controller := deploymentrelease.Controller{Runtime: runtime,
		LedgerPath: filepath.Join(config.InstallOperatorState, "release-ledger.jsonl"), Operator: config.Operator, Now: time.Now}
	return controller.Install(ctx, deploymentrelease.InstallRequest{Bundle: config.CandidateBundle})
}

package deploymentrehearsal

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"

	"cloud-clicker/server/deploymentbackup"
	"cloud-clicker/server/deploymentrelease"
	"cloud-clicker/server/releasepackage"
)

// lifecycleDependencies isolates the host runtime so the ordering and ledger
// contracts are testable; production uses the exact DockerRuntime.
type lifecycleDependencies struct {
	validateBundle func(string) error
	loadBundle     func(string) (deploymentrelease.Bundle, error)
	runtime        func(ScenarioConfig, string) deploymentrelease.InstallRuntime
	now            func() time.Time
}

func productionLifecycle() lifecycleDependencies {
	return lifecycleDependencies{validateBundle: releasepackage.ValidateBundle, loadBundle: deploymentrelease.LoadBundle, now: time.Now,
		runtime: func(config ScenarioConfig, state string) deploymentrelease.InstallRuntime {
			return deploymentrelease.DockerRuntime{PublicOrigin: config.PublicOrigin, ReceiverHealthURL: config.ReceiverHealthURL,
				BackupTarget: config.BackupTarget, MetricsDirectory: config.MetricsDirectory, AgeRecipient: config.AgeRecipient,
				AgeIdentityFile: config.AgeIdentityFile, ServerID: config.ServerID, DrainTimeout: 20 * time.Second,
				RotationLedgerPath: filepath.Join(state, "rotation-ledger.jsonl")}
		}}
}

// ReleaseLifecycle produces bounded_drain_and_restart. It runs only after the
// exact candidate install succeeded (install ledger), removes that candidate
// stack, installs the exact previous bundle into the separate lifecycle
// operator state, then releases the candidate over it through the production
// controller: preflight, pre-upgrade backup, drain, migration, identity and
// authenticated smoke. The lifecycle ledger must then hold exactly the
// previous install and the succeeded candidate release.
func ReleaseLifecycle(ctx context.Context, config ScenarioConfig) error {
	return releaseLifecycle(ctx, config, productionLifecycle())
}

func releaseLifecycle(ctx context.Context, config ScenarioConfig, dependencies lifecycleDependencies) error {
	candidate, previous, err := lifecycleInputs(config, dependencies)
	if err != nil {
		return err
	}
	install, err := readLedger(filepath.Join(config.InstallOperatorState, "release-ledger.jsonl"))
	if err != nil || len(install) != 1 || !matchesInstall(install[0], candidate.authority, candidate.digest) {
		return errors.Join(ErrInvalid, err)
	}
	if existing, err := readLedger(lifecycleLedger(config)); !errors.Is(err, os.ErrNotExist) || len(existing) != 0 {
		return errors.Join(ErrInvalid, err)
	}
	installed := dependencies.runtime(config, config.InstallOperatorState)
	if err := installed.AbortInstall(ctx, candidate.bundle); err != nil {
		return err
	}
	lifecycle := lifecycleController(config, dependencies)
	if err := lifecycle.Install(ctx, deploymentrelease.InstallRequest{Bundle: config.PreviousBundle}); err != nil {
		return err
	}
	if err := lifecycle.Release(ctx, deploymentrelease.ReleaseRequest{CurrentBundle: config.PreviousBundle, CandidateBundle: config.CandidateBundle}); err != nil {
		return err
	}
	records, err := readLedger(lifecycleLedger(config))
	if err != nil || len(records) != 2 || !matchesInstall(records[0], previous.authority, previous.digest) ||
		!matchesTransition(records[1], "release", candidate.authority, candidate.digest, previous.authority, previous.digest) ||
		records[0].Operator != records[1].Operator {
		return errors.Join(ErrInvalid, err)
	}
	return nil
}

// RollbackLifecycle produces exact_previous_release_rollback. It rolls the
// released candidate back to the exact previous bundle with that release's own
// pre-upgrade backup, then retains the three-row lifecycle ledger and the
// decoded pre-upgrade backup header as the operator evidence the final
// validator binds (validateOperatorRecords).
func RollbackLifecycle(ctx context.Context, config ScenarioConfig) error {
	return rollbackLifecycle(ctx, config, productionLifecycle())
}

func rollbackLifecycle(ctx context.Context, config ScenarioConfig, dependencies lifecycleDependencies) error {
	candidate, previous, err := lifecycleInputs(config, dependencies)
	if err != nil {
		return err
	}
	records, err := readLedger(lifecycleLedger(config))
	if err != nil || len(records) != 2 ||
		!matchesTransition(records[1], "release", candidate.authority, candidate.digest, previous.authority, previous.digest) {
		return errors.Join(ErrInvalid, err)
	}
	backup := deploymentrelease.BackupReference{ID: records[1].BackupID, Path: filepath.Join(config.BackupTarget, records[1].BackupID+".ccbackup")}
	header, err := deploymentbackup.ReadHeader(backup.Path)
	if err != nil || header.BackupID != backup.ID || !header.PreUpgrade || header.ReleaseManifestSHA256 != previous.digest {
		return errors.Join(ErrInvalid, err)
	}
	lifecycle := lifecycleController(config, dependencies)
	if err := lifecycle.Rollback(ctx, deploymentrelease.RollbackRequest{FailedBundle: config.CandidateBundle,
		PreviousBundle: config.PreviousBundle, Backup: backup}); err != nil {
		return err
	}
	ledgerBytes, err := os.ReadFile(lifecycleLedger(config))
	if err != nil {
		return err
	}
	final, err := deploymentrelease.DecodeReleaseLedger(ledgerBytes)
	if err != nil || len(final) != 3 ||
		!matchesTransition(final[2], "rollback", previous.authority, previous.digest, candidate.authority, candidate.digest) ||
		final[2].BackupID != final[1].BackupID || final[2].Operator != final[1].Operator || final[2].StartedAt.After(final[1].RollbackUntil) {
		return errors.Join(ErrInvalid, err)
	}
	headerBytes, err := json.Marshal(header)
	if err != nil {
		return err
	}
	if err := writeNewEvidence(filepath.Join(config.ArtifactsDirectory, requiredRunArtifactFiles["release_ledger"]), ledgerBytes); err != nil {
		return err
	}
	return writeNewEvidence(filepath.Join(config.ArtifactsDirectory, requiredRunArtifactFiles["backup_header"]), append(headerBytes, '\n'))
}

type lifecycleBundle struct {
	bundle    deploymentrelease.Bundle
	authority manifestAuthority
	digest    string
}

func lifecycleInputs(config ScenarioConfig, dependencies lifecycleDependencies) (lifecycleBundle, lifecycleBundle, error) {
	if ValidateScenarioConfig(config) != nil || validateScenarioFilesystem(config) != nil || dependencies.validateBundle == nil ||
		dependencies.loadBundle == nil || dependencies.runtime == nil || dependencies.now == nil {
		return lifecycleBundle{}, lifecycleBundle{}, ErrInvalid
	}
	load := func(root string) (lifecycleBundle, error) {
		if dependencies.validateBundle(root) != nil {
			return lifecycleBundle{}, ErrInvalid
		}
		data, err := os.ReadFile(filepath.Join(root, releasepackage.ReleaseManifestPath))
		if err != nil {
			return lifecycleBundle{}, err
		}
		authority, err := decodeManifestAuthority(data)
		if err != nil {
			return lifecycleBundle{}, err
		}
		var manifest releasepackage.ReleaseManifest
		if err := json.Unmarshal(data, &manifest); err != nil {
			return lifecycleBundle{}, ErrInvalid
		}
		return lifecycleBundle{bundle: deploymentrelease.Bundle{Root: root, Manifest: manifest, ManifestSHA256: hashBytes(data)},
			authority: authority, digest: hashBytes(data)}, nil
	}
	candidate, err := load(config.CandidateBundle)
	if err != nil {
		return lifecycleBundle{}, lifecycleBundle{}, err
	}
	previous, err := load(config.PreviousBundle)
	if err != nil {
		return lifecycleBundle{}, lifecycleBundle{}, err
	}
	return candidate, previous, nil
}

func lifecycleController(config ScenarioConfig, dependencies lifecycleDependencies) deploymentrelease.Controller {
	return deploymentrelease.Controller{Runtime: dependencies.runtime(config, config.LifecycleOperatorState),
		LedgerPath: lifecycleLedger(config), Operator: config.Operator, Now: dependencies.now,
		Load: dependencies.loadBundle}
}

func lifecycleLedger(config ScenarioConfig) string {
	return filepath.Join(config.LifecycleOperatorState, "release-ledger.jsonl")
}

func readLedger(path string) ([]deploymentrelease.ReleaseRecord, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return deploymentrelease.DecodeReleaseLedger(data)
}

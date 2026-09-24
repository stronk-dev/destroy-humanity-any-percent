package deploymentrehearsal

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"

	"filippo.io/age"

	"cloud-clicker/server/deploymentbackup"
	"cloud-clicker/server/deploymentrelease"
)

// errProbeRuntimeReached marks a guard runtime call. Release negatives must be
// refused before any runtime mutation; the positive control proves the same
// inputs otherwise reach the runtime.
var errProbeRuntimeReached = errors.New("probe runtime reached")

type guardRuntime struct{ calls []string }

func (runtime *guardRuntime) reach(name string) error {
	runtime.calls = append(runtime.calls, name)
	return errProbeRuntimeReached
}
func (runtime *guardRuntime) Preflight(context.Context, deploymentrelease.Bundle) error {
	return runtime.reach("preflight")
}
func (runtime *guardRuntime) CreatePreUpgradeBackup(context.Context, deploymentrelease.Bundle) (deploymentrelease.BackupReference, error) {
	return deploymentrelease.BackupReference{}, runtime.reach("backup")
}
func (runtime *guardRuntime) Prepare(context.Context, deploymentrelease.Bundle) error {
	return runtime.reach("prepare")
}
func (runtime *guardRuntime) DrainCurrent(context.Context, deploymentrelease.Bundle) (deploymentrelease.DrainEvidence, error) {
	return deploymentrelease.DrainEvidence{}, runtime.reach("drain")
}
func (runtime *guardRuntime) Start(context.Context, deploymentrelease.Bundle) error {
	return runtime.reach("start")
}
func (runtime *guardRuntime) VerifyIdentity(context.Context, deploymentrelease.Bundle) error {
	return runtime.reach("identity")
}
func (runtime *guardRuntime) AuthenticatedSmoke(context.Context, deploymentrelease.Bundle) error {
	return runtime.reach("smoke")
}
func (runtime *guardRuntime) StopFailed(context.Context, deploymentrelease.Bundle) error {
	return runtime.reach("stop_failed")
}
func (runtime *guardRuntime) ResetDatabase(context.Context, deploymentrelease.Bundle) error {
	return runtime.reach("reset_database")
}
func (runtime *guardRuntime) Restore(context.Context, deploymentrelease.Bundle, deploymentrelease.BackupReference) error {
	return runtime.reach("restore")
}
func (runtime *guardRuntime) VerifyRestoreInputs(context.Context, deploymentrelease.Bundle, deploymentrelease.BackupReference) error {
	return runtime.reach("restore_inputs")
}

// runIrreversibleMigrationProbe drives the production release controller with
// the exact candidate and previous bundles. The unmodified pair must pass the
// compatibility gate and reach the runtime; a candidate whose database
// migration is below the running release (a schema that could only be reached
// by a Down migration) must be refused at `compatibility` before any runtime
// step, with a failed ledger row.
func runIrreversibleMigrationProbe(request ProbeRequest) (ProbeOutcome, error) {
	candidate, err := deploymentrelease.LoadBundle(request.CandidateBundle)
	if err != nil {
		return ProbeAccepted, err
	}
	previous, err := deploymentrelease.LoadBundle(request.PreviousBundle)
	if err != nil {
		return ProbeAccepted, err
	}
	release := func(candidate deploymentrelease.Bundle) (*guardRuntime, []deploymentrelease.ReleaseRecord, error) {
		root, err := os.MkdirTemp(request.WorkDirectory, "release-probe-")
		if err != nil {
			return nil, nil, err
		}
		defer os.RemoveAll(root)
		runtime := &guardRuntime{}
		ledger := filepath.Join(root, "release-ledger.jsonl")
		controller := deploymentrelease.Controller{Runtime: runtime, LedgerPath: ledger, Operator: "r006-probe",
			Now: time.Now, Load: func(path string) (deploymentrelease.Bundle, error) {
				if path == request.CandidateBundle {
					return candidate, nil
				}
				return previous, nil
			}}
		releaseErr := controller.Release(context.Background(), deploymentrelease.ReleaseRequest{
			CurrentBundle: request.PreviousBundle, CandidateBundle: request.CandidateBundle})
		records, readErr := deploymentrelease.ReadReleaseLedger(ledger)
		if releaseErr == nil || readErr != nil {
			return nil, nil, errors.Join(ErrInvalid, readErr)
		}
		return runtime, records, nil
	}
	control, _, err := release(candidate)
	if err != nil || len(control.calls) == 0 {
		return ProbeAccepted, errors.Join(ErrInvalid, err)
	}
	backward := candidate
	backward.Manifest.DatabaseMigration = previous.Manifest.DatabaseMigration - 1
	guard, records, err := release(backward)
	if err != nil {
		return ProbeAccepted, err
	}
	if len(guard.calls) == 0 && len(records) == 1 && records[0].FailureStage == "compatibility" {
		return ProbeRejected, nil
	}
	return ProbeAccepted, nil
}

// runMissingPreviousInputProbe proves rollback refuses a missing pre-upgrade
// backup and a missing previous image at the production gates that run before
// anything destructive: VerifyRestoreInputs (host envelope, identity) and
// Prepare (image load and config-ID inspection). Each gate first accepts the
// intact input so a rejection is attributable to the removed input alone.
func runMissingPreviousInputProbe(request ProbeRequest) (ProbeOutcome, error) {
	previous, err := deploymentrelease.LoadBundle(request.PreviousBundle)
	if err != nil {
		return ProbeAccepted, err
	}
	root, err := os.MkdirTemp(request.WorkDirectory, "rollback-input-probe-")
	if err != nil {
		return ProbeAccepted, err
	}
	defer os.RemoveAll(root)
	target, metrics := filepath.Join(root, "backups"), filepath.Join(root, "metrics")
	for _, directory := range []string{target, metrics} {
		if err := os.Mkdir(directory, 0o700); err != nil {
			return ProbeAccepted, err
		}
	}
	identity, err := age.GenerateX25519Identity()
	if err != nil {
		return ProbeAccepted, err
	}
	identityPath := filepath.Join(root, "identity")
	if err := os.WriteFile(identityPath, []byte(identity.String()+"\n"), 0o600); err != nil {
		return ProbeAccepted, err
	}
	const serverID = "01986666-b001-4000-8000-000000000006"
	now := time.Now().UTC()
	header, backupPath, err := deploymentbackup.Create(deploymentbackup.CreateInput{Directory: target, BackupID: now.Format("20060102T150405Z") + "-abcdef123456",
		ServerID: serverID, ReleaseManifestSHA256: previous.ManifestSHA256, EpochID: previous.Manifest.EpochID, StartedAt: now.Add(-time.Second),
		Now: func() time.Time { return now }, Recipient: identity.Recipient().String(), Dump: bytes.NewReader([]byte("PGDMP\x01rollback-input probe")), PreUpgrade: true})
	if err != nil {
		return ProbeAccepted, err
	}
	imageMissing := false
	runner := probeRunner(func(args []string) ([]byte, error) {
		joined := strings.Join(args, " ")
		switch {
		case strings.HasPrefix(joined, "load ") || strings.HasPrefix(joined, "pull "):
			return nil, nil
		case strings.HasPrefix(joined, "image inspect "):
			if imageMissing {
				return nil, errors.New("Error: No such image")
			}
			for _, image := range previous.Manifest.Images {
				if strings.HasSuffix(joined, " "+image.Reference) {
					return []byte(image.RuntimeConfigSHA256 + "\n"), nil
				}
			}
		}
		return nil, errors.New("unexpected rollback probe command: " + joined)
	})
	runtime := deploymentrelease.DockerRuntime{Runner: runner, PublicOrigin: "https://probe.example", ReceiverHealthURL: "http://alertmanager:9093/-/healthy",
		BackupTarget: target, MetricsDirectory: metrics, AgeRecipient: identity.Recipient().String(), AgeIdentityFile: identityPath,
		ServerID: serverID, DrainTimeout: 20 * time.Second, RotationLedgerPath: filepath.Join(root, "rotation-ledger.jsonl")}
	backup := deploymentrelease.BackupReference{ID: header.BackupID, Path: backupPath}
	if runtime.VerifyRestoreInputs(context.Background(), previous, backup) != nil || runtime.Prepare(context.Background(), previous) != nil {
		return ProbeAccepted, ErrInvalid
	}
	if err := os.Remove(backupPath); err != nil {
		return ProbeAccepted, err
	}
	imageMissing = true
	if runtime.VerifyRestoreInputs(context.Background(), previous, backup) != nil && runtime.Prepare(context.Background(), previous) != nil {
		return ProbeRejected, nil
	}
	return ProbeAccepted, nil
}

type probeRunner func([]string) ([]byte, error)

func (runner probeRunner) Run(_ context.Context, _ string, name string, args ...string) ([]byte, error) {
	if name != "docker" {
		return nil, errors.New("unexpected rollback probe executable")
	}
	return runner(args)
}

// changeCurrentEpoch rewrites the packaged epoch declaration to claim a
// different current epoch, which no longer derives the manifest's identity.
func changeCurrentEpoch(data []byte) ([]byte, error) {
	var declaration map[string]any
	if err := json.Unmarshal(data, &declaration); err != nil {
		return nil, ErrInvalid
	}
	current, ok := declaration["current_epoch_id"].(float64)
	if !ok {
		return nil, ErrInvalid
	}
	declaration["current_epoch_id"] = current - 1
	changed, err := json.MarshalIndent(declaration, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(changed, '\n'), nil
}

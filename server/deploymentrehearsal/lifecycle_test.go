package deploymentrehearsal

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"cloud-clicker/server/deploymentbackup"
	"cloud-clicker/server/deploymentrelease"
	"cloud-clicker/server/releasepackage"
)

type lifecycleRuntime struct {
	calls     []string
	config    ScenarioConfig
	previous  string
	recipient string
	epoch     int64
}

func (runtime *lifecycleRuntime) call(name string) error {
	runtime.calls = append(runtime.calls, name)
	return nil
}
func (runtime *lifecycleRuntime) Preflight(context.Context, deploymentrelease.Bundle) error {
	return runtime.call("preflight")
}
func (runtime *lifecycleRuntime) CreatePreUpgradeBackup(_ context.Context, bundle deploymentrelease.Bundle) (deploymentrelease.BackupReference, error) {
	runtime.calls = append(runtime.calls, "backup")
	now := time.Now().UTC()
	header, path, err := deploymentbackup.Create(deploymentbackup.CreateInput{Directory: runtime.config.BackupTarget,
		BackupID: now.Format("20060102T150405Z") + "-abcdef123456", ServerID: runtime.config.ServerID,
		ReleaseManifestSHA256: bundle.ManifestSHA256, EpochID: runtime.epoch, StartedAt: now.Add(-time.Second),
		Now: func() time.Time { return now }, Recipient: runtime.recipient, Dump: bytes.NewReader([]byte("PGDMP\x01lifecycle")), PreUpgrade: true})
	return deploymentrelease.BackupReference{ID: header.BackupID, Path: path}, err
}
func (runtime *lifecycleRuntime) Prepare(context.Context, deploymentrelease.Bundle) error {
	return runtime.call("prepare")
}
func (runtime *lifecycleRuntime) DrainCurrent(context.Context, deploymentrelease.Bundle) (deploymentrelease.DrainEvidence, error) {
	return deploymentrelease.DrainEvidence{ReadinessDown: true, CourtesyFrame: true, IntentsRefused: true, AdmittedComplete: true,
		JobsFlushed: true, SocketsClosed: true, WithinBound: true}, runtime.call("drain")
}
func (runtime *lifecycleRuntime) Start(context.Context, deploymentrelease.Bundle) error {
	return runtime.call("start")
}
func (runtime *lifecycleRuntime) VerifyIdentity(context.Context, deploymentrelease.Bundle) error {
	return runtime.call("identity")
}
func (runtime *lifecycleRuntime) AuthenticatedSmoke(context.Context, deploymentrelease.Bundle) error {
	return runtime.call("smoke")
}
func (runtime *lifecycleRuntime) StopFailed(context.Context, deploymentrelease.Bundle) error {
	return runtime.call("stop_failed")
}
func (runtime *lifecycleRuntime) ResetDatabase(context.Context, deploymentrelease.Bundle) error {
	return runtime.call("reset_database")
}
func (runtime *lifecycleRuntime) Restore(context.Context, deploymentrelease.Bundle, deploymentrelease.BackupReference) error {
	return runtime.call("restore")
}
func (runtime *lifecycleRuntime) VerifyRestoreInputs(_ context.Context, _ deploymentrelease.Bundle, backup deploymentrelease.BackupReference) error {
	runtime.calls = append(runtime.calls, "restore_inputs")
	_, err := deploymentbackup.ReadHeader(backup.Path)
	return err
}
func (runtime *lifecycleRuntime) PreflightInstall(context.Context, deploymentrelease.Bundle) error {
	return runtime.call("install_preflight")
}
func (runtime *lifecycleRuntime) StartInstall(context.Context, deploymentrelease.Bundle) error {
	return runtime.call("install_start")
}
func (runtime *lifecycleRuntime) AbortInstall(context.Context, deploymentrelease.Bundle) error {
	return runtime.call("abort_install")
}

type lifecycleFixture struct {
	config       ScenarioConfig
	runtime      *lifecycleRuntime
	dependencies lifecycleDependencies
}

func newLifecycleFixture(t *testing.T) lifecycleFixture {
	t.Helper()
	config := validScenarioConfig(t)
	write := func(root, version, commit string) {
		t.Helper()
		images := make([]map[string]string, 6)
		for index, name := range []string{"alertmanager", "caddy", "gameserver", "node-exporter", "postgres", "prometheus"} {
			images[index] = map[string]string{"name": name, "reference": name + ":" + version + "@sha256:" + strings.Repeat(string(rune('1'+index)), 64)}
		}
		data, err := json.Marshal(map[string]any{"schema_version": 1, "release_version": version, "source_commit": commit, "epoch_id": 8,
			"database_migration": 74, "images": images})
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, releasepackage.ReleaseManifestPath), append(data, '\n'), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	write(config.PreviousBundle, "0.1.0-preview.3", strings.Repeat("a", 40))
	write(config.CandidateBundle, "0.1.0-preview.4", strings.Repeat("b", 40))
	runtime := &lifecycleRuntime{config: config, recipient: config.AgeRecipient, epoch: 8}
	load := func(root string) (deploymentrelease.Bundle, error) {
		data, err := os.ReadFile(filepath.Join(root, releasepackage.ReleaseManifestPath))
		if err != nil {
			return deploymentrelease.Bundle{}, err
		}
		var manifest releasepackage.ReleaseManifest
		if err := json.Unmarshal(data, &manifest); err != nil {
			return deploymentrelease.Bundle{}, err
		}
		return deploymentrelease.Bundle{Root: root, Manifest: manifest, ManifestSHA256: hashBytes(data)}, nil
	}
	clock := time.Now().UTC()
	dependencies := lifecycleDependencies{validateBundle: func(string) error { return nil }, loadBundle: load,
		now:     func() time.Time { clock = clock.Add(time.Second); return clock },
		runtime: func(ScenarioConfig, string) deploymentrelease.InstallRuntime { return runtime }}
	return lifecycleFixture{config: config, runtime: runtime, dependencies: dependencies}
}

func (fixture lifecycleFixture) installCandidate(t *testing.T) {
	t.Helper()
	controller := deploymentrelease.Controller{Runtime: fixture.runtime, LedgerPath: filepath.Join(fixture.config.InstallOperatorState, "release-ledger.jsonl"),
		Operator: fixture.config.Operator, Now: fixture.dependencies.now, Load: fixture.dependencies.loadBundle}
	if err := controller.Install(context.Background(), deploymentrelease.InstallRequest{Bundle: fixture.config.CandidateBundle}); err != nil {
		t.Fatal(err)
	}
	fixture.runtime.calls = nil
}

func TestLifecycleProducersRunInstallReleaseAndRollbackInContractOrder(t *testing.T) {
	fixture := newLifecycleFixture(t)
	fixture.installCandidate(t)
	if err := releaseLifecycle(context.Background(), fixture.config, fixture.dependencies); err != nil {
		t.Fatal(err)
	}
	wantRelease := []string{"abort_install", "prepare", "install_preflight", "install_start", "identity", "smoke",
		"prepare", "preflight", "backup", "drain", "start", "identity", "smoke"}
	if !slices.Equal(fixture.runtime.calls, wantRelease) {
		t.Fatalf("release lifecycle calls=%v", fixture.runtime.calls)
	}
	fixture.runtime.calls = nil
	if err := rollbackLifecycle(context.Background(), fixture.config, fixture.dependencies); err != nil {
		t.Fatal(err)
	}
	wantRollback := []string{"prepare", "preflight", "restore_inputs", "stop_failed", "reset_database", "restore", "start", "identity", "smoke"}
	if !slices.Equal(fixture.runtime.calls, wantRollback) {
		t.Fatalf("rollback lifecycle calls=%v", fixture.runtime.calls)
	}
	ledger, err := deploymentrelease.DecodeReleaseLedger(mustRead(t, filepath.Join(fixture.config.ArtifactsDirectory, requiredRunArtifactFiles["release_ledger"])))
	if err != nil || len(ledger) != 3 || ledger[0].Action != "install" || ledger[1].Action != "release" || ledger[2].Action != "rollback" {
		t.Fatalf("retained lifecycle ledger=%+v err=%v", ledger, err)
	}
	// The produced rows must satisfy the exact predicates validateOperatorRecords applies.
	previousBytes := mustRead(t, filepath.Join(fixture.config.PreviousBundle, releasepackage.ReleaseManifestPath))
	candidateBytes := mustRead(t, filepath.Join(fixture.config.CandidateBundle, releasepackage.ReleaseManifestPath))
	previous, _ := decodeManifestAuthority(previousBytes)
	candidate, _ := decodeManifestAuthority(candidateBytes)
	if !matchesInstall(ledger[0], previous, hashBytes(previousBytes)) ||
		!matchesTransition(ledger[1], "release", candidate, hashBytes(candidateBytes), previous, hashBytes(previousBytes)) ||
		!matchesTransition(ledger[2], "rollback", previous, hashBytes(previousBytes), candidate, hashBytes(candidateBytes)) {
		t.Fatalf("produced lifecycle rows do not satisfy the evidence validator: %+v", ledger)
	}
	header, err := deploymentbackup.DecodeHeader(mustRead(t, filepath.Join(fixture.config.ArtifactsDirectory, requiredRunArtifactFiles["backup_header"])))
	if err != nil || header.BackupID != ledger[1].BackupID || !header.PreUpgrade {
		t.Fatalf("retained backup header=%+v err=%v", header, err)
	}
}

func TestLifecycleProducersRefuseOutOfOrderOrMissingAuthority(t *testing.T) {
	t.Run("release before candidate install", func(t *testing.T) {
		fixture := newLifecycleFixture(t)
		if err := releaseLifecycle(context.Background(), fixture.config, fixture.dependencies); !errors.Is(err, ErrInvalid) {
			t.Fatalf("release without candidate install accepted: %v", err)
		}
		if len(fixture.runtime.calls) != 0 {
			t.Fatalf("out-of-order release reached runtime: %v", fixture.runtime.calls)
		}
	})
	t.Run("release over an existing lifecycle", func(t *testing.T) {
		fixture := newLifecycleFixture(t)
		fixture.installCandidate(t)
		if err := os.WriteFile(filepath.Join(fixture.config.LifecycleOperatorState, "release-ledger.jsonl"), nil, 0o600); err != nil {
			t.Fatal(err)
		}
		if err := releaseLifecycle(context.Background(), fixture.config, fixture.dependencies); !errors.Is(err, ErrInvalid) {
			t.Fatalf("second lifecycle accepted: %v", err)
		}
		if len(fixture.runtime.calls) != 0 {
			t.Fatalf("occupied lifecycle reached runtime: %v", fixture.runtime.calls)
		}
	})
	t.Run("rollback before release", func(t *testing.T) {
		fixture := newLifecycleFixture(t)
		if err := rollbackLifecycle(context.Background(), fixture.config, fixture.dependencies); err == nil {
			t.Fatal("rollback without a release accepted")
		}
		if len(fixture.runtime.calls) != 0 {
			t.Fatalf("rollback without release reached runtime: %v", fixture.runtime.calls)
		}
	})
	t.Run("rollback with missing pre-upgrade backup", func(t *testing.T) {
		fixture := newLifecycleFixture(t)
		fixture.installCandidate(t)
		if err := releaseLifecycle(context.Background(), fixture.config, fixture.dependencies); err != nil {
			t.Fatal(err)
		}
		entries, err := os.ReadDir(fixture.config.BackupTarget)
		if err != nil || len(entries) != 1 {
			t.Fatalf("backup target=%v err=%v", entries, err)
		}
		if err := os.Remove(filepath.Join(fixture.config.BackupTarget, entries[0].Name())); err != nil {
			t.Fatal(err)
		}
		fixture.runtime.calls = nil
		if err := rollbackLifecycle(context.Background(), fixture.config, fixture.dependencies); err == nil {
			t.Fatal("rollback without its pre-upgrade backup accepted")
		}
		if slices.Contains(fixture.runtime.calls, "reset_database") {
			t.Fatalf("missing backup reached destructive rollback: %v", fixture.runtime.calls)
		}
	})
}

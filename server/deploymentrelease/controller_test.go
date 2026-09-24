package deploymentrelease

import (
	"context"
	"errors"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"cloud-clicker/server/releasepackage"
)

type runtimeFixture struct {
	failAt string
	drain  DrainEvidence
	calls  []string
	backup BackupReference
}

func (runtime *runtimeFixture) call(name string) error {
	runtime.calls = append(runtime.calls, name)
	if runtime.failAt == name {
		return errors.New("severed " + name)
	}
	return nil
}
func (runtime *runtimeFixture) Preflight(context.Context, Bundle) error {
	return runtime.call("preflight")
}
func (runtime *runtimeFixture) CreatePreUpgradeBackup(context.Context, Bundle) (BackupReference, error) {
	err := runtime.call("backup")
	return runtime.backup, err
}
func (runtime *runtimeFixture) Prepare(context.Context, Bundle) error { return runtime.call("prepare") }
func (runtime *runtimeFixture) DrainCurrent(context.Context, Bundle) (DrainEvidence, error) {
	err := runtime.call("drain")
	return runtime.drain, err
}
func (runtime *runtimeFixture) Start(context.Context, Bundle) error { return runtime.call("start") }
func (runtime *runtimeFixture) VerifyIdentity(context.Context, Bundle) error {
	return runtime.call("identity")
}
func (runtime *runtimeFixture) AuthenticatedSmoke(context.Context, Bundle) error {
	return runtime.call("smoke")
}
func (runtime *runtimeFixture) StopFailed(context.Context, Bundle) error {
	return runtime.call("stop_failed")
}
func (runtime *runtimeFixture) ResetDatabase(context.Context, Bundle) error {
	return runtime.call("reset_database")
}
func (runtime *runtimeFixture) Restore(context.Context, Bundle, BackupReference) error {
	return runtime.call("restore")
}
func (runtime *runtimeFixture) VerifyRestoreInputs(context.Context, Bundle, BackupReference) error {
	return runtime.call("restore_inputs")
}
func (runtime *runtimeFixture) PreflightInstall(context.Context, Bundle) error {
	return runtime.call("install_preflight")
}
func (runtime *runtimeFixture) StartInstall(context.Context, Bundle) error {
	return runtime.call("install_start")
}
func (runtime *runtimeFixture) AbortInstall(context.Context, Bundle) error {
	return runtime.call("abort_install")
}

func TestInstallSequenceAuthorityCleanupAndRetry(t *testing.T) {
	current, candidate, loader := fixtureBundles()
	newController := func(runtime Runtime, ledger string, clock *sequenceClock) Controller {
		return Controller{Runtime: runtime, LedgerPath: ledger, Operator: "operator-1", Now: clock.Time, Load: loader}
	}
	t.Run("exact success", func(t *testing.T) {
		clock := &sequenceClock{now: time.Date(2026, 8, 23, 12, 0, 0, 0, time.UTC)}
		runtime := validRuntime()
		ledger := filepath.Join(t.TempDir(), "release-ledger.jsonl")
		if err := newController(runtime, ledger, clock).Install(context.Background(), InstallRequest{Bundle: candidate}); err != nil {
			t.Fatal(err)
		}
		want := []string{"prepare", "install_preflight", "install_start", "identity", "smoke"}
		if strings.Join(runtime.calls, ",") != strings.Join(want, ",") {
			t.Fatalf("install sequence=%v", runtime.calls)
		}
		records, err := ReadReleaseLedger(ledger)
		if err != nil || len(records) != 1 || records[0].Action != "install" || records[0].Result != "succeeded" ||
			records[0].BackupID != "none" || records[0].PreviousVersion != "" || !records[0].RollbackUntil.IsZero() {
			t.Fatalf("install record=%+v err=%v", records, err)
		}
		runtime.calls = nil
		if err := newController(runtime, ledger, clock).Install(context.Background(), InstallRequest{Bundle: candidate}); err == nil {
			t.Fatal("second install accepted over successful install")
		}
		if len(runtime.calls) != 0 {
			t.Fatalf("closed install authority reached runtime: %v", runtime.calls)
		}
	})

	for _, stage := range []string{"install_preflight", "prepare", "install_start", "identity", "smoke"} {
		t.Run("severed_"+stage, func(t *testing.T) {
			clock := &sequenceClock{now: time.Date(2026, 8, 23, 13, 0, 0, 0, time.UTC)}
			runtime := validRuntime()
			runtime.failAt = stage
			ledger := filepath.Join(t.TempDir(), "release-ledger.jsonl")
			if err := newController(runtime, ledger, clock).Install(context.Background(), InstallRequest{Bundle: candidate}); err == nil {
				t.Fatalf("severed %s accepted", stage)
			}
			started := stage == "install_start" || stage == "identity" || stage == "smoke"
			if slices.Contains(runtime.calls, "abort_install") != started {
				t.Fatalf("stage=%s calls=%v", stage, runtime.calls)
			}
			records, err := ReadReleaseLedger(ledger)
			if err != nil || len(records) != 1 || records[0].Action != "install" || records[0].Result != "failed" {
				t.Fatalf("install records=%+v err=%v", records, err)
			}
		})
	}

	t.Run("failed attempt permits corrected bundle", func(t *testing.T) {
		clock := &sequenceClock{now: time.Date(2026, 8, 23, 14, 0, 0, 0, time.UTC)}
		ledger := filepath.Join(t.TempDir(), "release-ledger.jsonl")
		failed := validRuntime()
		failed.failAt = "install_preflight"
		if err := newController(failed, ledger, clock).Install(context.Background(), InstallRequest{Bundle: candidate}); err == nil {
			t.Fatal("severed initial attempt accepted")
		}
		corrected := validRuntime()
		if err := newController(corrected, ledger, clock).Install(context.Background(), InstallRequest{Bundle: current}); err != nil {
			t.Fatalf("corrected exact bundle retry rejected: %v", err)
		}
	})

	t.Run("ledger failure aborts running install", func(t *testing.T) {
		times := []time.Time{
			time.Date(2026, 8, 23, 15, 0, 1, 0, time.UTC),
			time.Date(2026, 8, 23, 15, 0, 0, 0, time.UTC),
		}
		now := func() time.Time { value := times[0]; times = times[1:]; return value }
		runtime := validRuntime()
		controller := Controller{Runtime: runtime, LedgerPath: filepath.Join(t.TempDir(), "release-ledger.jsonl"), Operator: "operator-1", Now: now, Load: loader}
		if err := controller.Install(context.Background(), InstallRequest{Bundle: candidate}); err == nil {
			t.Fatal("invalid success record accepted")
		}
		if !slices.Contains(runtime.calls, "abort_install") {
			t.Fatalf("ledger failure left install running: %v", runtime.calls)
		}
	})
}

func TestInstallRejectsRuntimeWithoutInstallAuthority(t *testing.T) {
	_, candidate, loader := fixtureBundles()
	runtime := struct{ Runtime }{Runtime: validRuntime()}
	controller := Controller{Runtime: runtime, LedgerPath: filepath.Join(t.TempDir(), "release-ledger.jsonl"), Operator: "operator-1", Now: time.Now, Load: loader}
	if err := controller.Install(context.Background(), InstallRequest{Bundle: candidate}); !errors.Is(err, ErrInvalid) {
		t.Fatalf("runtime without install boundary accepted: %v", err)
	}
}

func TestReleaseSequenceAndEveryRuledSeveringStage(t *testing.T) {
	current, candidate, loader := fixtureBundles()
	for _, stage := range []string{"prepare", "preflight", "backup", "drain", "start", "identity", "smoke"} {
		t.Run(stage, func(t *testing.T) {
			clock := &sequenceClock{now: time.Date(2026, 8, 22, 12, 0, 0, 0, time.UTC)}
			runtime := validRuntime()
			runtime.failAt = stage
			ledger := filepath.Join(t.TempDir(), "release-ledger.jsonl")
			controller := Controller{Runtime: runtime, LedgerPath: ledger, Operator: "operator-1", Now: clock.Time, Load: loader}
			if err := controller.Release(context.Background(), ReleaseRequest{CurrentBundle: current, CandidateBundle: candidate}); err == nil {
				t.Fatalf("severed %s accepted", stage)
			}
			records, err := ReadReleaseLedger(ledger)
			if err != nil || len(records) != 1 || records[0].Result != "failed" || records[0].FailureStage == "" {
				t.Fatalf("failure record=%+v err=%v", records, err)
			}
		})
	}

	clock := &sequenceClock{now: time.Date(2026, 8, 22, 13, 0, 0, 0, time.UTC)}
	runtime := validRuntime()
	ledger := filepath.Join(t.TempDir(), "release-ledger.jsonl")
	controller := Controller{Runtime: runtime, LedgerPath: ledger, Operator: "operator-1", Now: clock.Time, Load: loader}
	if err := controller.Release(context.Background(), ReleaseRequest{CurrentBundle: current, CandidateBundle: candidate}); err != nil {
		t.Fatal(err)
	}
	want := []string{"prepare", "preflight", "backup", "drain", "start", "identity", "smoke"}
	if strings.Join(runtime.calls, ",") != strings.Join(want, ",") {
		t.Fatalf("sequence=%v want=%v", runtime.calls, want)
	}
	records, _ := ReadReleaseLedger(ledger)
	if len(records) != 1 || records[0].Result != "succeeded" || records[0].RollbackUntil.Sub(records[0].CompletedAt) != RollbackWindow {
		t.Fatalf("success record=%+v", records)
	}
}

func TestReleaseRejectsEachMissingDrainProperty(t *testing.T) {
	current, candidate, loader := fixtureBundles()
	for _, sever := range []func(*DrainEvidence){
		func(v *DrainEvidence) { v.ReadinessDown = false }, func(v *DrainEvidence) { v.CourtesyFrame = false },
		func(v *DrainEvidence) { v.IntentsRefused = false }, func(v *DrainEvidence) { v.AdmittedComplete = false },
		func(v *DrainEvidence) { v.JobsFlushed = false }, func(v *DrainEvidence) { v.SocketsClosed = false },
		func(v *DrainEvidence) { v.WithinBound = false },
	} {
		runtime := validRuntime()
		sever(&runtime.drain)
		controller := Controller{Runtime: runtime, LedgerPath: filepath.Join(t.TempDir(), "release-ledger.jsonl"), Operator: "operator-1", Now: time.Now, Load: loader}
		if err := controller.Release(context.Background(), ReleaseRequest{CurrentBundle: current, CandidateBundle: candidate}); err == nil {
			t.Fatal("incomplete drain evidence accepted")
		}
	}
}

func TestRollbackUsesExactPreviousManifestBackupAndSevenDayWindow(t *testing.T) {
	current, candidate, loader := fixtureBundles()
	setup := func(t *testing.T) (*sequenceClock, *runtimeFixture, Controller) {
		t.Helper()
		clock := &sequenceClock{now: time.Date(2026, 8, 22, 12, 0, 0, 0, time.UTC)}
		runtime := validRuntime()
		controller := Controller{Runtime: runtime, LedgerPath: filepath.Join(t.TempDir(), "release-ledger.jsonl"), Operator: "operator-1", Now: clock.Time, Load: loader}
		if err := controller.Release(context.Background(), ReleaseRequest{CurrentBundle: current, CandidateBundle: candidate}); err != nil {
			t.Fatal(err)
		}
		runtime.calls = nil
		return clock, runtime, controller
	}
	t.Run("exact success", func(t *testing.T) {
		clock, runtime, controller := setup(t)
		clock.now = clock.now.Add(6 * 24 * time.Hour)
		if err := controller.Rollback(context.Background(), RollbackRequest{FailedBundle: candidate, PreviousBundle: current, Backup: runtime.backup}); err != nil {
			t.Fatal(err)
		}
		want := []string{"prepare", "preflight", "restore_inputs", "stop_failed", "reset_database", "restore", "start", "identity", "smoke"}
		if strings.Join(runtime.calls, ",") != strings.Join(want, ",") {
			t.Fatalf("rollback sequence=%v", runtime.calls)
		}
	})
	t.Run("expired", func(t *testing.T) {
		clock, runtime, controller := setup(t)
		clock.now = clock.now.Add(8 * 24 * time.Hour)
		if err := controller.Rollback(context.Background(), RollbackRequest{FailedBundle: candidate, PreviousBundle: current, Backup: runtime.backup}); err == nil {
			t.Fatal("rollback after seven days accepted")
		}
		if len(runtime.calls) != 0 {
			t.Fatalf("expired authority reached runtime: %v", runtime.calls)
		}
	})
	t.Run("wrong backup", func(t *testing.T) {
		_, runtime, controller := setup(t)
		wrong := BackupReference{ID: "20260822T180000Z-acde00000002", Path: runtime.backup.Path}
		if err := controller.Rollback(context.Background(), RollbackRequest{FailedBundle: candidate, PreviousBundle: current, Backup: wrong}); err == nil {
			t.Fatal("wrong backup accepted")
		}
		if len(runtime.calls) != 0 {
			t.Fatalf("wrong backup reached runtime: %v", runtime.calls)
		}
	})
	t.Run("wrong previous manifest", func(t *testing.T) {
		_, runtime, controller := setup(t)
		if err := controller.Rollback(context.Background(), RollbackRequest{FailedBundle: candidate, PreviousBundle: candidate, Backup: runtime.backup}); err == nil {
			t.Fatal("wrong previous manifest accepted")
		}
		if len(runtime.calls) != 0 {
			t.Fatalf("wrong previous manifest reached runtime: %v", runtime.calls)
		}
	})
}

func TestRollbackRecordsEverySeveredExecutionStage(t *testing.T) {
	current, candidate, loader := fixtureBundles()
	for _, stage := range []string{"prepare", "preflight", "restore_inputs", "stop_failed", "reset_database", "restore", "start", "identity", "smoke"} {
		t.Run(stage, func(t *testing.T) {
			clock := &sequenceClock{now: time.Date(2026, 8, 22, 12, 0, 0, 0, time.UTC)}
			ledger := filepath.Join(t.TempDir(), "release-ledger.jsonl")
			initial := validRuntime()
			controller := Controller{Runtime: initial, LedgerPath: ledger, Operator: "operator-1", Now: clock.Time, Load: loader}
			if err := controller.Release(context.Background(), ReleaseRequest{CurrentBundle: current, CandidateBundle: candidate}); err != nil {
				t.Fatal(err)
			}
			clock.now = clock.now.Add(time.Hour)
			severed := validRuntime()
			severed.failAt = stage
			controller.Runtime = severed
			if err := controller.Rollback(context.Background(), RollbackRequest{FailedBundle: candidate, PreviousBundle: current, Backup: initial.backup}); err == nil {
				t.Fatalf("severed rollback stage %s accepted", stage)
			}
			records, err := ReadReleaseLedger(ledger)
			if err != nil || len(records) != 2 || records[1].Action != "rollback" || records[1].Result != "failed" || records[1].FailureStage == "" {
				t.Fatalf("rollback records=%+v err=%v", records, err)
			}
			// No input failure may stop the failed release or remove the live database.
			if stage == "prepare" || stage == "preflight" || stage == "restore_inputs" {
				if slices.Contains(severed.calls, "stop_failed") || slices.Contains(severed.calls, "reset_database") {
					t.Fatalf("severed %s reached destructive rollback: %v", stage, severed.calls)
				}
			}
		})
	}
}

func TestReleaseRejectsDowngradeAndRecordsInvalidCandidate(t *testing.T) {
	current, candidate, loader := fixtureBundles()
	clock := &sequenceClock{now: time.Date(2026, 8, 22, 12, 0, 0, 0, time.UTC)}
	ledger := filepath.Join(t.TempDir(), "release-ledger.jsonl")
	downgradeLoader := func(path string) (Bundle, error) {
		bundle, err := loader(path)
		if err == nil && path == candidate {
			bundle.Manifest.DatabaseMigration = 72
		}
		return bundle, err
	}
	controller := Controller{Runtime: validRuntime(), LedgerPath: ledger, Operator: "operator-1", Now: clock.Time, Load: downgradeLoader}
	if err := controller.Release(context.Background(), ReleaseRequest{CurrentBundle: current, CandidateBundle: candidate}); err == nil {
		t.Fatal("backward migration accepted")
	}
	records, _ := ReadReleaseLedger(ledger)
	if len(records) != 1 || records[0].FailureStage != "compatibility" {
		t.Fatalf("downgrade record=%+v", records)
	}

	invalidLedger := filepath.Join(t.TempDir(), "release-ledger.jsonl")
	invalidController := Controller{Runtime: validRuntime(), LedgerPath: invalidLedger, Operator: "operator-1", Now: clock.Time,
		Load: func(string) (Bundle, error) { return Bundle{}, ErrInvalid }}
	if err := invalidController.Release(context.Background(), ReleaseRequest{CurrentBundle: current, CandidateBundle: "/missing"}); err == nil {
		t.Fatal("invalid candidate accepted")
	}
	invalidRecords, err := ReadReleaseLedger(invalidLedger)
	if err != nil || len(invalidRecords) != 1 || invalidRecords[0].FailureStage != "candidate_bundle" || invalidRecords[0].Result != "failed" {
		t.Fatalf("invalid candidate record=%+v err=%v", invalidRecords, err)
	}
}

func TestReleaseRejectsVersionDowngradeEvenWhenMigrationIsForwardCompatible(t *testing.T) {
	current, candidate, loader := fixtureBundles()
	downgradeLoader := func(path string) (Bundle, error) {
		bundle, err := loader(path)
		if err == nil && path == candidate {
			bundle.Manifest.ReleaseVersion = "0.9.0"
		}
		return bundle, err
	}
	controller := Controller{Runtime: validRuntime(), LedgerPath: filepath.Join(t.TempDir(), "release-ledger.jsonl"),
		Operator: "operator-1", Now: time.Now, Load: downgradeLoader}
	if err := controller.Release(context.Background(), ReleaseRequest{CurrentBundle: current, CandidateBundle: candidate}); err == nil {
		t.Fatal("normal release accepted a semantic version downgrade")
	}
}

func TestFailedPostBackupReleaseAuthorizesExactRollback(t *testing.T) {
	current, candidate, loader := fixtureBundles()
	clock := &sequenceClock{now: time.Date(2026, 8, 22, 12, 0, 0, 0, time.UTC)}
	ledger := filepath.Join(t.TempDir(), "release-ledger.jsonl")
	failedRuntime := validRuntime()
	failedRuntime.failAt = "smoke"
	controller := Controller{Runtime: failedRuntime, LedgerPath: ledger, Operator: "operator-1", Now: clock.Time, Load: loader}
	if err := controller.Release(context.Background(), ReleaseRequest{CurrentBundle: current, CandidateBundle: candidate}); err == nil {
		t.Fatal("severed candidate smoke accepted")
	}
	records, err := ReadReleaseLedger(ledger)
	if err != nil || len(records) != 1 || records[0].Result != "failed" || records[0].RollbackUntil.IsZero() {
		t.Fatalf("failed release authority=%+v err=%v", records, err)
	}
	rollbackRuntime := validRuntime()
	controller.Runtime = rollbackRuntime
	clock.now = clock.now.Add(time.Hour)
	if err := controller.Rollback(context.Background(), RollbackRequest{FailedBundle: candidate, PreviousBundle: current, Backup: failedRuntime.backup}); err != nil {
		t.Fatal(err)
	}
}

func TestFailedRollbackCanRetryButSuccessfulRollbackClosesAuthority(t *testing.T) {
	current, candidate, loader := fixtureBundles()
	clock := &sequenceClock{now: time.Date(2026, 8, 22, 12, 0, 0, 0, time.UTC)}
	ledger := filepath.Join(t.TempDir(), "release-ledger.jsonl")
	initial := validRuntime()
	controller := Controller{Runtime: initial, LedgerPath: ledger, Operator: "operator-1", Now: clock.Time, Load: loader}
	if err := controller.Release(context.Background(), ReleaseRequest{CurrentBundle: current, CandidateBundle: candidate}); err != nil {
		t.Fatal(err)
	}
	clock.now = clock.now.Add(time.Hour)
	severed := validRuntime()
	severed.failAt = "restore"
	controller.Runtime = severed
	request := RollbackRequest{FailedBundle: candidate, PreviousBundle: current, Backup: initial.backup}
	if err := controller.Rollback(context.Background(), request); err == nil {
		t.Fatal("severed rollback unexpectedly succeeded")
	}
	controller.Runtime = validRuntime()
	if err := controller.Rollback(context.Background(), request); err != nil {
		t.Fatalf("failed rollback could not be retried: %v", err)
	}
	if err := controller.Rollback(context.Background(), request); err == nil {
		t.Fatal("successful rollback left authority open")
	}
}

type sequenceClock struct{ now time.Time }

func (clock *sequenceClock) Time() time.Time {
	value := clock.now
	clock.now = clock.now.Add(time.Second)
	return value
}

func validRuntime() *runtimeFixture {
	return &runtimeFixture{backup: BackupReference{ID: "20260822T180000Z-acde00000001", Path: "/backups/20260822T180000Z-acde00000001.ccbackup"}, drain: DrainEvidence{
		ReadinessDown: true, CourtesyFrame: true, IntentsRefused: true, AdmittedComplete: true, JobsFlushed: true, SocketsClosed: true, WithinBound: true,
	}}
}

func fixtureBundles() (string, string, func(string) (Bundle, error)) {
	current, candidate := "/bundles/1.0.0", "/bundles/1.1.0"
	makeBundle := func(root, version, fill string, migration int) Bundle {
		images := []releasepackage.Image{
			{Name: "alertmanager", Reference: "alertmanager:v1@sha256:" + strings.Repeat(fill, 64)},
			{Name: "caddy", Reference: "caddy:v1@sha256:" + strings.Repeat(fill, 64)},
			{Name: "gameserver", Reference: "sha256:" + strings.Repeat(fill, 64)},
			{Name: "node-exporter", Reference: "node-exporter:v1@sha256:" + strings.Repeat(fill, 64)},
			{Name: "postgres", Reference: "postgres:v1@sha256:" + strings.Repeat(fill, 64)},
			{Name: "prometheus", Reference: "prometheus:v1@sha256:" + strings.Repeat(fill, 64)},
		}
		return Bundle{Root: root, ManifestSHA256: "sha256:" + strings.Repeat(fill, 64), Manifest: releasepackage.ReleaseManifest{ReleaseVersion: version, DatabaseMigration: migration, Images: images}}
	}
	bundles := map[string]Bundle{current: makeBundle(current, "1.0.0", "a", 73), candidate: makeBundle(candidate, "1.1.0", "b", 74)}
	return current, candidate, func(path string) (Bundle, error) {
		bundle, ok := bundles[path]
		if !ok {
			return Bundle{}, ErrInvalid
		}
		return bundle, nil
	}
}

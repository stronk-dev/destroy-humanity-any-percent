package deploymentrehearsal

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"cloud-clicker/server/deploymentbackup"
	"cloud-clicker/server/deploymentrelease"
)

type nonCleanFake struct {
	restoreErr    error
	changeOnFail  bool
	identity      deploymentbackup.RecoveryIdentity
	backupPath    string
	restoreCalled bool
}

func (fake *nonCleanFake) CreateRecoveryBackup(context.Context, deploymentrelease.Bundle) (deploymentrelease.BackupReference, error) {
	return deploymentrelease.BackupReference{ID: "20260924T120000Z-abcdef123456", Path: fake.backupPath}, os.WriteFile(fake.backupPath, []byte("backup"), 0o600)
}
func (fake *nonCleanFake) RestoreRecoveryBackup(context.Context, deploymentrelease.Bundle, deploymentrelease.BackupReference) error {
	fake.restoreCalled = true
	if fake.changeOnFail {
		fake.identity.DatabaseMigration++
	}
	return fake.restoreErr
}
func (fake *nonCleanFake) InspectRecoveryIdentity(context.Context, deploymentrelease.Bundle) (deploymentbackup.RecoveryIdentity, error) {
	return fake.identity, nil
}

// realRefusal renders the exact structured failure the backup CLI writes for a
// non-clean target, wrapped the way ExecRunner reports a failed command.
func realRefusal(class string) error {
	var stderr bytes.Buffer
	slog.New(slog.NewJSONHandler(&stderr, nil)).Error("deployment backup command failed", "command", "restore", "error_class", class)
	return fmt.Errorf("docker failed: %w: %s", errors.New("exit status 1"), bytes.TrimSpace(stderr.Bytes()))
}

func TestNonCleanRestoreProducerRequiresTheCleanTargetRefusal(t *testing.T) {
	run := func(t *testing.T, fake *nonCleanFake, installed bool) (ProbeOutcome, error) {
		t.Helper()
		fixture := newLifecycleFixture(t)
		if installed {
			fixture.installCandidate(t)
		}
		fake.backupPath = filepath.Join(fixture.config.BackupTarget, "20260924T120000Z-abcdef123456.ccbackup")
		outcome, err := probeNonCleanRestore(context.Background(), fixture.config, nonCleanDependencies{loadBundle: fixture.dependencies.loadBundle,
			refusal: deploymentrelease.IsNonCleanRestoreRefusal, runtime: func(ScenarioConfig) nonCleanRuntime { return fake }})
		if _, statErr := os.Stat(fake.backupPath); !errors.Is(statErr, os.ErrNotExist) && installed {
			t.Fatalf("probe left its recovery backup behind: %v", statErr)
		}
		return outcome, err
	}
	if outcome, err := run(t, &nonCleanFake{restoreErr: realRefusal("non_clean_target")}, true); err != nil || outcome != ProbeRejected {
		t.Fatalf("clean-target refusal outcome=%d err=%v", outcome, err)
	}
	if outcome, err := run(t, &nonCleanFake{}, true); err != nil || outcome != ProbeAccepted {
		t.Fatalf("accepted in-place restore outcome=%d err=%v", outcome, err)
	}
	if outcome, err := run(t, &nonCleanFake{restoreErr: realRefusal("operation_failed")}, true); err == nil || outcome == ProbeRejected {
		t.Fatalf("unrelated restore failure counted as rejection: outcome=%d err=%v", outcome, err)
	}
	if outcome, err := run(t, &nonCleanFake{restoreErr: realRefusal("non_clean_target"), changeOnFail: true}, true); err == nil || outcome == ProbeRejected {
		t.Fatalf("refusal that changed the live database counted as rejection: outcome=%d err=%v", outcome, err)
	}
	fake := &nonCleanFake{restoreErr: realRefusal("non_clean_target")}
	if outcome, err := run(t, fake, false); err == nil || outcome == ProbeRejected || fake.restoreCalled {
		t.Fatalf("probe ran without an installed candidate: outcome=%d err=%v", outcome, err)
	}
}

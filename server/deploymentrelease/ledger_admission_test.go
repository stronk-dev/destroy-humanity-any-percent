package deploymentrelease

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLifecycleRejectsInadmissibleLedgerBeforeRuntimeOperations(t *testing.T) {
	for _, action := range []string{"install", "release", "rollback"} {
		for _, condition := range []string{"malformed", "empty", "directory", "symlink", "public mode", "earlier clock", "equal clock", "valid"} {
			t.Run(action+"/"+condition, func(t *testing.T) {
				current, candidate, loader := fixtureBundles()
				currentBundle, err := loader(current)
				if err != nil {
					t.Fatal(err)
				}
				candidateBundle, err := loader(candidate)
				if err != nil {
					t.Fatal(err)
				}
				base := time.Date(2026, 8, 23, 12, 0, 0, 0, time.UTC)
				started := base.Add(2 * time.Minute)
				path := filepath.Join(t.TempDir(), "release-ledger.jsonl")
				runtime := validRuntime()
				controller := Controller{Runtime: runtime, LedgerPath: path, Operator: "operator-1",
					Now: func() time.Time { return started }, Load: loader}
				record := controller.baseRecord("install", currentBundle, Bundle{}, base)
				record.CompletedAt, record.Result = base.Add(time.Minute), "succeeded"
				if action == "install" {
					// A failed initial install is retryable with a valid ledger.
					record.Result, record.FailureStage = "failed", "preflight"
				} else if action == "rollback" {
					record = controller.baseRecord("release", candidateBundle, currentBundle, base)
					record.CompletedAt, record.Result = base.Add(time.Minute), "succeeded"
					record.BackupID = runtime.backup.ID
					record.RollbackUntil = record.CompletedAt.Add(RollbackWindow)
				}
				if err := AppendReleaseRecord(path, record); err != nil {
					t.Fatal(err)
				}
				switch condition {
				case "malformed":
					if err := os.WriteFile(path, []byte("not a ledger\n"), 0o600); err != nil {
						t.Fatal(err)
					}
				case "empty":
					if err := os.WriteFile(path, nil, 0o600); err != nil {
						t.Fatal(err)
					}
				case "directory":
					if err := os.Rename(path, path+".original"); err != nil {
						t.Fatal(err)
					}
					if err := os.Mkdir(path, 0o700); err != nil {
						t.Fatal(err)
					}
				case "symlink":
					if err := os.Rename(path, path+".original"); err != nil {
						t.Fatal(err)
					}
					if err := os.Symlink(path+".original", path); err != nil {
						t.Fatal(err)
					}
				case "public mode":
					if err := os.Chmod(path, 0o644); err != nil {
						t.Fatal(err)
					}
				case "earlier clock":
					started = record.CompletedAt.Add(-time.Nanosecond)
				case "equal clock":
					started = record.CompletedAt
				}
				before, beforeErr := os.ReadFile(path)
				switch action {
				case "install":
					err = controller.Install(context.Background(), InstallRequest{Bundle: current})
				case "release":
					err = controller.Release(context.Background(), ReleaseRequest{CurrentBundle: current, CandidateBundle: candidate})
				case "rollback":
					err = controller.Rollback(context.Background(), RollbackRequest{FailedBundle: candidate, PreviousBundle: current, Backup: runtime.backup})
				}
				if condition != "valid" {
					if err == nil || len(runtime.calls) != 0 {
						t.Fatalf("inadmissible ledger reached runtime: err=%v calls=%v", err, runtime.calls)
					}
					after, afterErr := os.ReadFile(path)
					if (beforeErr == nil) != (afterErr == nil) || !bytes.Equal(before, after) {
						t.Fatalf("refused operation changed ledger bytes: before=%v after=%v", beforeErr, afterErr)
					}
					return
				}
				if err != nil || len(runtime.calls) == 0 {
					t.Fatalf("admissible lifecycle failed: err=%v calls=%v", err, runtime.calls)
				}
				result, err := ReadReleaseLedger(path)
				if err != nil || len(result) != 2 || result[1].Action != action || result[1].Result != "succeeded" || !result[1].StartedAt.Equal(started) {
					t.Fatalf("admissible operation lost its exact success record: %+v err=%v", result, err)
				}
			})
		}
	}
}

func TestOperatorStateRejectsDirectorySymlinksWithoutChangingTheirTargets(t *testing.T) {
	for _, command := range []string{"lock", "append"} {
		for _, fixture := range []struct {
			name    string
			mode    os.FileMode
			symlink bool
		}{
			{"public target symlink", 0o775, true},
			{"private target symlink", 0o700, true},
			{"real directory", 0o775, false},
		} {
			t.Run(command+"/"+fixture.name, func(t *testing.T) {
				root := t.TempDir()
				target := filepath.Join(root, "target")
				if err := os.Mkdir(target, 0o700); err != nil {
					t.Fatal(err)
				}
				if err := os.Chmod(target, fixture.mode); err != nil {
					t.Fatal(err)
				}
				directory := target
				if fixture.symlink {
					directory = filepath.Join(root, "operator-state")
					if err := os.Symlink(target, directory); err != nil {
						t.Fatal(err)
					}
				}
				var err error
				if command == "lock" {
					lock, lockErr := AcquireOperatorStateLock(directory)
					err = lockErr
					if lockErr == nil {
						if lock == nil {
							t.Fatal("successful lock returned no closer")
						}
						err = lock.Close()
					}
				} else {
					err = AppendReleaseRecord(filepath.Join(directory, "release-ledger.jsonl"),
						fixtureReleaseRecord(time.Date(2026, 8, 23, 12, 0, 0, 0, time.UTC)))
				}
				info, statErr := os.Stat(target)
				if statErr != nil {
					t.Fatal(statErr)
				}
				entries, readErr := os.ReadDir(target)
				if fixture.symlink {
					link, linkErr := os.Lstat(directory)
					if !errors.Is(err, ErrInvalid) || info.Mode().Perm() != fixture.mode || readErr != nil || len(entries) != 0 || linkErr != nil || link.Mode()&os.ModeSymlink == 0 {
						t.Fatalf("refused directory symlink mutated target: err=%v mode=%o/%o entries=%d read=%v link=%v",
							err, info.Mode().Perm(), fixture.mode, len(entries), readErr, linkErr)
					}
					return
				}
				if err != nil || info.Mode().Perm() != 0o700 || readErr != nil || len(entries) != 1 {
					t.Fatalf("real operator directory failed private initialization: err=%v mode=%o entries=%d read=%v",
						err, info.Mode().Perm(), len(entries), readErr)
				}
				if command == "append" {
					if records, err := ReadReleaseLedger(filepath.Join(directory, "release-ledger.jsonl")); err != nil || len(records) != 1 {
						t.Fatalf("real directory lost its readable record: records=%d err=%v", len(records), err)
					}
				}
			})
		}
	}
}

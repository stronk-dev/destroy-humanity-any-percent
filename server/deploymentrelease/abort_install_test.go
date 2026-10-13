package deploymentrelease

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestAbortInstallPreservesNormalizedRotationOverlay(t *testing.T) {
	for _, test := range []struct {
		name      string
		jwt       bool
		bootstrap bool
		removed   bool
		malformed bool
		wantError bool
	}{
		{name: "no ledger"},
		{name: "both open", jwt: true, bootstrap: true},
		{name: "both removed", jwt: true, bootstrap: true, removed: true},
		{name: "JWT only", jwt: true, wantError: true},
		{name: "bootstrap only", bootstrap: true, wantError: true},
		{name: "malformed ledger", malformed: true, wantError: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			bundle := dockerFixtureBundle(t)
			runner := &commandFixture{}
			runtime := dockerFixtureRuntime(runner, t.TempDir(), "")
			seedAbortRotation(t, runtime.RotationLedgerPath, test.jwt, test.bootstrap, test.removed)
			if test.malformed {
				if err := os.WriteFile(runtime.RotationLedgerPath, []byte("not a ledger\n"), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			err := runtime.AbortInstall(context.Background(), bundle)
			if test.wantError {
				if err == nil || len(runner.calls) != 0 {
					t.Fatalf("invalid rotation reached cleanup: calls=%v err=%v", runner.calls, err)
				}
				return
			}
			want := append([]string{"docker"}, abortComposeArgs(bundle, test.jwt && test.bootstrap && !test.removed)...)
			if err != nil || len(runner.calls) != 1 || !reflect.DeepEqual(runner.calls[0], want) {
				t.Fatalf("cleanup lost normalized rotation: calls=%v want=%v err=%v", runner.calls, want, err)
			}
		})
	}
}

func TestAbortInstallUsesDefaultCommandRunnerAndReturnsCommandFailure(t *testing.T) {
	for _, overlap := range []bool{false, true} {
		name := "without overlap"
		if overlap {
			name = "with both overlaps"
		}
		t.Run(name, func(t *testing.T) {
			defer func() {
				if value := recover(); value != nil {
					t.Fatalf("production-shaped cleanup panicked instead of returning a command failure: %v", value)
				}
			}()
			bundle := dockerFixtureBundle(t)
			runtime := dockerFixtureRuntime(nil, t.TempDir(), "")
			seedAbortRotation(t, runtime.RotationLedgerPath, overlap, overlap, false)
			// Exercise ExecRunner's real process boundary, but never Docker or
			// a daemon: the only executable on PATH is this owned recording stub.
			bin := t.TempDir()
			arguments := filepath.Join(bin, "arguments")
			stub := []byte("#!/bin/sh\nprintf '%s\\n' \"$@\" > \"$CLOUD_CLICKER_ABORT_TEST_ARGS\"\nexit 23\n")
			if err := os.WriteFile(filepath.Join(bin, "docker"), stub, 0o700); err != nil {
				t.Fatal(err)
			}
			t.Setenv("PATH", bin)
			t.Setenv("CLOUD_CLICKER_ABORT_TEST_ARGS", arguments)
			err := runtime.AbortInstall(context.Background(), bundle)
			var failure *CommandFailure
			if !errors.As(err, &failure) || failure.Command != "docker" || failure.ExitCode != 23 {
				t.Fatalf("cleanup failed to return the actual command failure: %v", err)
			}
			data, err := os.ReadFile(arguments)
			want := abortComposeArgs(bundle, overlap)
			if err != nil || !reflect.DeepEqual(strings.Split(strings.TrimSuffix(string(data), "\n"), "\n"), want) {
				t.Fatalf("default runner lost cleanup arguments: got=%q want=%v err=%v", data, want, err)
			}
		})
	}
}

func abortComposeArgs(bundle Bundle, overlap bool) []string {
	args := []string{"compose", "--project-name", "cloud-clicker", "--file", filepath.Join(bundle.Root, "compose.yml")}
	if overlap {
		args = append(args, "--file", filepath.Join(bundle.Root, "compose.rotation.yml"))
	}
	return append(args, "down", "--volumes", "--remove-orphans")
}

func seedAbortRotation(t *testing.T, path string, jwt, bootstrap, removed bool) {
	t.Helper()
	now := time.Date(2026, 8, 23, 12, 0, 0, 0, time.UTC)
	for _, family := range []KeyFamily{FamilyJWT, FamilyBootstrap} {
		if family == FamilyJWT && !jwt || family == FamilyBootstrap && !bootstrap {
			continue
		}
		if err := ActivateRotation(path, family, "key-2", "key-1", "operator-1", now); err != nil {
			t.Fatal(err)
		}
	}
	if !removed {
		return
	}
	// Keep the append-only ledger chronological: activate both families
	// before removing either, at a time satisfying both governed minima.
	for _, family := range []KeyFamily{FamilyJWT, FamilyBootstrap} {
		if family == FamilyJWT && !jwt || family == FamilyBootstrap && !bootstrap {
			continue
		}
		if err := RemovePrevious(path, family, "key-2", "key-1", "operator-1", now.Add(BootstrapOverlap)); err != nil {
			t.Fatal(err)
		}
	}
}

package deploymentrehearsal

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestBundleMutationProbesRequirePreparedMutationAndGateRejection(t *testing.T) {
	for name, mutation := range bundleMutations {
		t.Run(name, func(t *testing.T) {
			request, original := probeFixture(t, name, mutation)
			outcome, err := runBundleMutationProbe(request, mutation, func(root string) error {
				path := filepath.Join(root, filepath.FromSlash(mutation.path))
				if mutation.mutate == nil {
					if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
						t.Fatalf("removal fixture remained: %v", err)
					}
				} else {
					changed, err := os.ReadFile(path)
					if err != nil || string(changed) == string(original) {
						t.Fatalf("rewrite fixture not changed: %v", err)
					}
				}
				return errors.New("release gate rejected exact mutation")
			})
			if err != nil || outcome != ProbeRejected {
				t.Fatalf("prepared rejected fixture outcome=%d err=%v", outcome, err)
			}
			unchanged, err := os.ReadFile(filepath.Join(request.CandidateBundle, filepath.FromSlash(mutation.path)))
			if err != nil || string(unchanged) != string(original) {
				t.Fatalf("probe modified retained candidate: %v", err)
			}

			outcome, err = runBundleMutationProbe(request, mutation, func(string) error { return nil })
			if err != nil || outcome != ProbeAccepted {
				t.Fatalf("gate acceptance masqueraded as rejection: outcome=%d err=%v", outcome, err)
			}
		})
	}
}

func TestBundleMutationProbeSetupFailureCannotCountAsRejection(t *testing.T) {
	mutation := bundleMutations["removed_catalog"]
	request, _ := probeFixture(t, "removed_catalog", mutation)
	if err := os.Remove(filepath.Join(request.CandidateBundle, filepath.FromSlash(mutation.path))); err != nil {
		t.Fatal(err)
	}
	outcome, err := runBundleMutationProbe(request, mutation, func(string) error {
		t.Fatal("validator reached after fixture setup failure")
		return nil
	})
	if err == nil || outcome == ProbeRejected {
		t.Fatalf("setup failure satisfied negative row: outcome=%d err=%v", outcome, err)
	}
}

func TestProbeRejectsUnsupportedPopulationAndUnsafePaths(t *testing.T) {
	root := t.TempDir()
	candidate := filepath.Join(root, "candidate")
	previous := filepath.Join(root, "previous")
	work := filepath.Join(root, "work")
	for _, path := range []string{candidate, previous, work} {
		if err := os.Mkdir(path, 0o700); err != nil {
			t.Fatal(err)
		}
	}
	request := ProbeRequest{Population: "not_registered", CandidateBundle: candidate, PreviousBundle: previous, WorkDirectory: work}
	if outcome, err := RunProbe(request); !errors.Is(err, ErrInvalid) || outcome == ProbeRejected {
		t.Fatalf("unsupported population became rejection: outcome=%d err=%v", outcome, err)
	}
	request.WorkDirectory = filepath.Join(candidate, "work")
	if err := os.Mkdir(request.WorkDirectory, 0o700); err != nil {
		t.Fatal(err)
	}
	if outcome, err := RunProbe(request); !errors.Is(err, ErrInvalid) || outcome == ProbeRejected {
		t.Fatalf("nested work path accepted: outcome=%d err=%v", outcome, err)
	}
}

func probeFixture(t *testing.T, name string, mutation bundleMutation) (ProbeRequest, []byte) {
	t.Helper()
	root := t.TempDir()
	candidate := filepath.Join(root, "candidate")
	previous := filepath.Join(root, "previous")
	work := filepath.Join(root, "work")
	for _, path := range []string{candidate, previous, work} {
		if err := os.Mkdir(path, 0o700); err != nil {
			t.Fatal(err)
		}
	}
	original := []byte("fixture")
	if mutation.mutate != nil {
		if name == "changed_sbom" {
			original = []byte("{\"spdx\":true}\n")
		} else {
			original = []byte("{\"reference\": \"sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa\", \"runtime_config_sha256\": \"sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb\"}\n")
		}
	}
	path := filepath.Join(candidate, filepath.FromSlash(mutation.path))
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, original, 0o600); err != nil {
		t.Fatal(err)
	}
	return ProbeRequest{Population: name, CandidateBundle: candidate, PreviousBundle: previous, WorkDirectory: work}, original
}

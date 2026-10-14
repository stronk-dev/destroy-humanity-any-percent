package deploymentbackup

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func writeIdentityTemporary(t *testing.T, path, content string, now time.Time) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	old := now.Add(-2 * StaleTemporaryAge)
	if err := os.Chtimes(path, old, old); err != nil {
		t.Fatal(err)
	}
}

func requireIdentityBytes(t *testing.T, path, want string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil || string(data) != want {
		t.Fatalf("refusal changed test-owned file %s: %v", filepath.Base(path), err)
	}
}

func TestCleanupRejectsReplacementDirectoryIdentity(t *testing.T) {
	for _, kind := range []string{"renamed-directory", "retargeted-alias"} {
		t.Run(kind, func(t *testing.T) {
			base, now := t.TempDir(), time.Now().UTC()
			original := filepath.Join(base, "original")
			replacement := filepath.Join(base, "replacement")
			for _, path := range []string{original, replacement} {
				if err := os.Mkdir(path, 0o700); err != nil {
					t.Fatal(err)
				}
			}
			name := ".backup-envelope-123.tmp"
			writeIdentityTemporary(t, filepath.Join(original, name), "inspected temporary", now)
			writeIdentityTemporary(t, filepath.Join(replacement, name), "not inspected", now)
			target := original
			if kind == "retargeted-alias" {
				target = filepath.Join(base, "alias")
				if err := os.Symlink(original, target); err != nil {
					t.Fatal(err)
				}
			}
			state, err := InspectTarget(target, now)
			if err != nil || len(state.StaleTemporaries) != 1 {
				t.Fatal("invalid inspected population")
			}
			if kind == "renamed-directory" {
				saved := filepath.Join(base, "inspected")
				if err := os.Rename(original, saved); err != nil {
					t.Fatal(err)
				}
				if err := os.Rename(replacement, target); err != nil {
					t.Fatal(err)
				}
				original, replacement = saved, target
			} else {
				if err := os.Remove(target); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(replacement, target); err != nil {
					t.Fatal(err)
				}
			}
			if err := RemoveStaleTemporaries(state, now); !errors.Is(err, ErrInvalid) {
				t.Errorf("replaced inspected directory admitted: %v", err)
			}
			requireIdentityBytes(t, filepath.Join(original, name), "inspected temporary")
			requireIdentityBytes(t, filepath.Join(replacement, name), "not inspected")
		})
	}
}

func TestCleanupRejectsReplacementFileIdentity(t *testing.T) {
	directory, now := t.TempDir(), time.Now().UTC()
	path := filepath.Join(directory, ".backup-payload-123.tmp")
	writeIdentityTemporary(t, path, "inspected temporary", now)
	state, err := InspectTarget(directory, now)
	if err != nil || len(state.StaleTemporaries) != 1 {
		t.Fatal("invalid inspected population")
	}
	saved := filepath.Join(directory, "inspected.txt")
	if err := os.Rename(path, saved); err != nil {
		t.Fatal(err)
	}
	writeIdentityTemporary(t, path, "not inspected", now)
	if err := RemoveStaleTemporaries(state, now); !errors.Is(err, ErrInvalid) {
		t.Errorf("matching name and age concealed replaced inode: %v", err)
	}
	requireIdentityBytes(t, path, "not inspected")
	requireIdentityBytes(t, saved, "inspected temporary")
}

func TestCleanupPreservesUnchangedDirectoryAliases(t *testing.T) {
	for _, alias := range []bool{false, true} {
		t.Run(map[bool]string{false: "real", true: "alias"}[alias], func(t *testing.T) {
			base, now := t.TempDir(), time.Now().UTC()
			directory := filepath.Join(base, "target")
			if err := os.Mkdir(directory, 0o700); err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(directory, ".backup-payload-456.tmp")
			writeIdentityTemporary(t, path, "inspected temporary", now)
			foreign := filepath.Join(directory, "leave-alone.txt")
			if err := os.WriteFile(foreign, []byte("untouched"), 0o600); err != nil {
				t.Fatal(err)
			}
			target := directory
			if alias {
				target = filepath.Join(base, "alias")
				if err := os.Symlink(directory, target); err != nil {
					t.Fatal(err)
				}
			}
			state, err := InspectTarget(target, now)
			if err != nil || len(state.StaleTemporaries) != 1 || len(state.Foreign) != 1 {
				t.Fatal("invalid healthy control")
			}
			if err := RemoveStaleTemporaries(state, now); err != nil {
				t.Fatal(err)
			}
			if _, err := os.Lstat(path); !errors.Is(err, os.ErrNotExist) {
				t.Fatal("unchanged own temporary was not removed")
			}
			requireIdentityBytes(t, foreign, "untouched")
		})
	}
}

// Exercise the actual destructive helper after its directory handle is bound.
// This is a deterministic local filesystem boundary, not a timed race or proof
// of the public entry point's advisory-lock behavior (covered separately).
func TestCleanupBoundRootSurvivesDirectoryReplacement(t *testing.T) {
	base, now := t.TempDir(), time.Now().UTC()
	directory := filepath.Join(base, "target")
	if err := os.Mkdir(directory, 0o700); err != nil {
		t.Fatal(err)
	}
	name := ".backup-payload-789.tmp"
	writeIdentityTemporary(t, filepath.Join(directory, name), "inspected temporary", now)
	state, err := InspectTarget(directory, now)
	if err != nil || len(state.StaleTemporaries) != 1 {
		t.Fatal("invalid inspected population")
	}
	root, err := os.OpenRoot(directory)
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()
	saved := filepath.Join(base, "inspected")
	if err := os.Rename(directory, saved); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(directory, 0o700); err != nil {
		t.Fatal(err)
	}
	writeIdentityTemporary(t, filepath.Join(directory, name), "not inspected", now)
	if err := removeStaleTemporariesInRoot(root, state, now); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Lstat(filepath.Join(saved, name)); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("bound cleanup did not remove the original inspected temporary")
	}
	requireIdentityBytes(t, filepath.Join(directory, name), "not inspected")
}

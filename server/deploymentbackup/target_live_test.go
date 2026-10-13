package deploymentbackup

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"filippo.io/age"
)

type cleanupProbeRead func([]byte) (int, error)

func (read cleanupProbeRead) Read(data []byte) (int, error) { return read(data) }

type cleanupProbeReport struct {
	Stale, Active int
	Refused       bool
	Remains       bool
	PID           int
}

func TestBackupCleanupProcessHelper(t *testing.T) {
	target := os.Getenv("CLOUD_CLICKER_TEST_CLEANUP_TARGET")
	if target == "" {
		t.Skip("child-process fixture only")
	}
	now, err := time.Parse(time.RFC3339Nano, os.Getenv("CLOUD_CLICKER_TEST_CLEANUP_NOW"))
	if err != nil || !filepath.IsAbs(target) {
		t.Fatal("invalid cleanup child fixture")
	}
	state, err := InspectTarget(target, now)
	if err != nil || len(state.StaleTemporaries) != 1 {
		t.Fatal("child did not inspect the actual aged temporary")
	}
	cleanupErr := RemoveStaleTemporaries(state, now)
	_, statErr := os.Stat(state.StaleTemporaries[0])
	if err := json.NewEncoder(os.Stdout).Encode(cleanupProbeReport{Stale: len(state.StaleTemporaries), Active: len(state.ActiveTemporaries),
		Refused: errors.Is(cleanupErr, ErrTargetBusy), Remains: statErr == nil, PID: os.Getpid()}); err != nil {
		t.Fatal(err)
	}
}

func TestCleanupCannotRemoveALiveAgedBackupTemporary(t *testing.T) {
	identity, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	now := time.Now().UTC()
	manifest := "sha256:" + strings.Repeat("a", 64)
	dump := []byte("PGDMP\x01live writer survives cleanup")
	entered, release := make(chan struct{}), make(chan struct{})
	var readOnce, releaseOnce sync.Once
	reader := bytes.NewReader(dump)
	type completion struct {
		header Header
		path   string
		err    error
	}
	completed := make(chan completion, 1)
	go func() {
		header, path, err := Create(CreateInput{Directory: directory, BackupID: now.Format("20060102T150405Z") + "-abcdef123456",
			ServerID: "fixture-server", ReleaseManifestSHA256: manifest, EpochID: 8,
			StartedAt: now.Add(-time.Minute), Now: func() time.Time { return now }, Recipient: identity.Recipient().String(),
			Dump: cleanupProbeRead(func(data []byte) (int, error) {
				readOnce.Do(func() { close(entered); <-release })
				return reader.Read(data)
			})})
		completed <- completion{header, path, err}
	}()
	var finished *completion
	finish := func() completion {
		releaseOnce.Do(func() { close(release) })
		if finished == nil {
			result := <-completed
			finished = &result
		}
		return *finished
	}
	t.Cleanup(func() { finish() })
	select {
	case <-entered:
	case early := <-completed:
		finished = &early
		t.Fatal("actual writer failed before the paused dump read")
	case <-t.Context().Done():
		t.Fatal("actual writer did not reach the dump read")
	}
	state, err := InspectTarget(directory, now)
	if err != nil || len(state.ActiveTemporaries) != 1 || len(state.StaleTemporaries)+len(state.Backups)+len(state.Foreign) != 0 {
		t.Fatal("live writer lacks its exclusive owned temporary")
	}
	payload := state.ActiveTemporaries[0]
	old := now.Add(-StaleTemporaryAge - time.Second)
	if err := os.Chtimes(payload, old, old); err != nil {
		t.Fatal(err)
	}
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	child := exec.CommandContext(t.Context(), executable, "-test.run=^TestBackupCleanupProcessHelper$")
	child.Dir = root
	child.Env = append(os.Environ(), "CLOUD_CLICKER_TEST_CLEANUP_TARGET="+directory,
		"CLOUD_CLICKER_TEST_CLEANUP_NOW="+now.Format(time.RFC3339Nano))
	output, err := child.Output()
	if err != nil {
		t.Fatal("cleanup child failed (private dump and key withheld)")
	}
	var report cleanupProbeReport
	if json.NewDecoder(bytes.NewReader(output)).Decode(&report) != nil || report.PID == os.Getpid() || report.Stale != 1 || report.Active != 0 {
		t.Fatal("cleanup observation is not the actual second-process stale population")
	}
	if !report.Refused || !report.Remains {
		t.Errorf("second-process cleanup removed or admitted the live aged temporary: refused=%t remains=%t", report.Refused, report.Remains)
	}
	result := finish()
	if result.err != nil {
		t.Error("live encrypted writer could not complete after cleanup")
		return
	}
	var restored bytes.Buffer
	if header, err := Restore(result.path, manifest, identity, &restored); err != nil || header != result.header || !bytes.Equal(restored.Bytes(), dump) {
		t.Fatal("writer did not preserve authenticated bytes and identity")
	}
	// Once no creator owns the target, real crash debris still needs removal.
	orphan := filepath.Join(directory, ".backup-envelope-4242.tmp")
	if err := os.WriteFile(orphan, []byte("abandoned"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(orphan, old, old); err != nil {
		t.Fatal(err)
	}
	state, err = InspectTarget(directory, now)
	if err != nil || RemoveStaleTemporaries(state, now) != nil {
		t.Fatal("genuine orphan cleanup failed after the writer completed")
	}
	if _, err := os.Stat(orphan); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("orphan retained after the creator released ownership")
	}
	if actual, err := ReadHeader(result.path); err != nil || actual != result.header {
		t.Fatal("orphan cleanup changed the completed backup")
	}
}

func TestCleanupRechecksInspectedTargetBeforeRemoving(t *testing.T) {
	for _, change := range []string{"freshened", "symlink", "directory", "foreign-name", "outside-target", "unbound-state"} {
		t.Run(change, func(t *testing.T) {
			now := time.Now().UTC()
			directory := t.TempDir()
			path := filepath.Join(directory, ".backup-payload-123.tmp")
			writeOld := func(path string) {
				t.Helper()
				if err := os.WriteFile(path, []byte("crash debris"), 0o600); err != nil {
					t.Fatal(err)
				}
				old := now.Add(-2 * StaleTemporaryAge)
				if err := os.Chtimes(path, old, old); err != nil {
					t.Fatal(err)
				}
			}
			writeOld(path)
			state, err := InspectTarget(directory, now)
			if err != nil || len(state.StaleTemporaries) != 1 {
				t.Fatal("invalid inspected stale population")
			}
			switch change {
			case "freshened":
				err = os.Chtimes(path, now, now)
			case "symlink":
				victim := filepath.Join(t.TempDir(), "victim")
				writeOld(victim)
				if err = os.Remove(path); err == nil {
					err = os.Symlink(victim, path)
				}
			case "directory":
				if err = os.Remove(path); err == nil {
					err = os.Mkdir(path, 0o700)
				}
			case "foreign-name":
				path = filepath.Join(directory, "foreign.tmp")
				writeOld(path)
				state.StaleTemporaries = []string{path}
			case "outside-target":
				path = filepath.Join(t.TempDir(), ".backup-envelope-456.tmp")
				writeOld(path)
				state.StaleTemporaries = []string{path}
			case "unbound-state":
				state.directory = ""
			}
			if err != nil {
				t.Fatal(err)
			}
			if err := RemoveStaleTemporaries(state, now); !errors.Is(err, ErrInvalid) {
				t.Fatalf("changed target admitted for cleanup: %v", err)
			}
			if _, err := os.Lstat(path); err != nil {
				t.Fatal("refused entry was removed")
			}
		})
	}
}

func TestCreateExcludesCleanupOwnershipAndReleasesAfterFailure(t *testing.T) {
	identity, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	now := time.Now().UTC()
	input := CreateInput{Directory: directory, BackupID: now.Format("20060102T150405Z") + "-abcdef123456",
		ServerID: "fixture-server", ReleaseManifestSHA256: "sha256:" + strings.Repeat("a", 64), EpochID: 8,
		StartedAt: now, Now: func() time.Time { return now }, Recipient: identity.Recipient().String()}
	ownership, err := lockTarget(directory, true)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { ownership.Close() })
	read := false
	input.Dump = cleanupProbeRead(func([]byte) (int, error) { read = true; return 0, errors.New("unexpected dump read") })
	if _, _, err := Create(input); !errors.Is(err, ErrTargetBusy) || read {
		t.Fatal("creator admitted while cleanup owns the target")
	}
	if entries, err := os.ReadDir(directory); err != nil || len(entries) != 0 {
		t.Fatal("blocked creator wrote into the cleanup target")
	}
	if err := ownership.Close(); err != nil {
		t.Fatal(err)
	}
	dumpErr := errors.New("fixture dump failure")
	input.Dump = cleanupProbeRead(func([]byte) (int, error) { return 0, dumpErr })
	if _, _, err := Create(input); !errors.Is(err, dumpErr) {
		t.Fatal("failed dump did not return its actual error")
	}
	cleanup, err := lockTarget(directory, true)
	if err != nil {
		t.Fatal("failed creator leaked target ownership")
	}
	if err := cleanup.Close(); err != nil {
		t.Fatal(err)
	}
	input.Dump = bytes.NewReader([]byte("PGDMP\x01healthy retry"))
	if _, _, err := Create(input); err != nil {
		t.Fatal("healthy creator could not retry after failure")
	}
}

func TestPackageCreationUnblocksItsPipeWhenCleanupOwnsTarget(t *testing.T) {
	identity, err := age.GenerateX25519Identity()
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	now := time.Now().UTC()
	epoch := fixtureEpoch(t, 8)
	manifest := fixtureReleaseManifest(t, epoch)
	dump := []byte("PGDMP\x01package retry retains complete dump")
	input := PackageInput{Directory: directory, BackupID: now.Format("20060102T150405Z") + "-abcdef123456",
		ServerID: "fixture-server", StartedAt: now, Now: func() time.Time { return now },
		Recipient: identity.Recipient().String(), Dump: bytes.NewReader(dump),
		ReleaseManifest: manifest, EpochDeclaration: epoch}
	ownership, err := lockTarget(directory, true)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { ownership.Close() })
	// This invokes the real package tar producer and pipe-to-encryption boundary.
	// Early refusal must return rather than leave its tar writer blocked forever.
	if _, _, err := CreatePackage(input); !errors.Is(err, ErrTargetBusy) {
		t.Fatal("package creator did not return its ownership refusal")
	}
	if entries, err := os.ReadDir(directory); err != nil || len(entries) != 0 {
		t.Fatal("refused package left partial backup files")
	}
	if err := ownership.Close(); err != nil {
		t.Fatal(err)
	}
	header, path, err := CreatePackage(input)
	if err != nil {
		t.Fatal("package could not retry after cleanup released ownership")
	}
	extracted, err := ExtractPackage(path, digest(manifest), identity, t.TempDir())
	if err != nil {
		t.Fatal("retried package could not be authenticated and extracted")
	}
	actual, err := os.ReadFile(extracted.DumpPath)
	if err != nil || !bytes.Equal(actual, dump) || extracted.Header != header {
		t.Fatal("package retry changed dump bytes or header identity")
	}
}

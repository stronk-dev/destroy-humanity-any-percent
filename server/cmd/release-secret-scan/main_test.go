package main

import (
	"archive/tar"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"cloud-clicker/server/releasepackage"
)

func TestStructuredSecretScanBindsSourceImageAndManifest(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "tracked.txt"), []byte("clean\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	archive := filepath.Join(root, "gameserver.tar")
	writeCleanArchive(t, archive)
	output := filepath.Join(root, "secret-scan.json")
	originalCommand, originalNow := commandOutput, now
	t.Cleanup(func() { commandOutput, now = originalCommand, originalNow })
	commandOutput = func(_ string, args ...string) ([]byte, error) {
		if len(args) == 2 && args[0] == "ls-files" {
			return []byte("tracked.txt\x00"), nil
		}
		return []byte(strings.Repeat("c", 40) + "\n"), nil
	}
	clock := time.Date(2026, 8, 23, 12, 0, 0, 0, time.UTC)
	now = func() time.Time {
		clock = clock.Add(time.Second)
		return clock
	}
	manifest := "sha256:" + strings.Repeat("a", 64)
	result, err := run([]string{"--root", root, "--gameserver-archive", archive, "--manifest-sha256", manifest, "--output", output})
	if err != nil || result.ManifestSHA256 != manifest || !result.ImageArchiveScanned {
		t.Fatalf("result=%+v err=%v", result, err)
	}
	data, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := releasepackage.DecodeSecretScanResult(data)
	if err != nil || decoded.SourceCommit != strings.Repeat("c", 40) {
		t.Fatalf("decoded=%+v err=%v", decoded, err)
	}
	if _, err := run([]string{"--root", root, "--manifest-sha256", manifest, "--output", filepath.Join(root, "missing-image.json")}); err == nil {
		t.Fatal("structured source-only scan accepted")
	}
}

func writeCleanArchive(t *testing.T, path string) {
	t.Helper()
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	writer := tar.NewWriter(file)
	data := []byte("clean image member\n")
	if err := writer.WriteHeader(&tar.Header{Name: "layer", Mode: 0o644, Typeflag: tar.TypeReg, Size: int64(len(data))}); err != nil {
		t.Fatal(err)
	}
	if _, err := writer.Write(data); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
}

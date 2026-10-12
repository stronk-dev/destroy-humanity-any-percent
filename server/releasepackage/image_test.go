package releasepackage

import (
	"archive/tar"
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateDockerArchiveRejectsMalformedTarWithoutPanic(t *testing.T) {
	archive, imageID := fixtureDockerArchive(t, t.TempDir(), "amd64")
	data, err := os.ReadFile(archive)
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateDockerArchive(archive, imageID, "0.1.0-preview.1", strings.Repeat("d", 40)); err != nil {
		t.Fatalf("intact archive refused: %v", err)
	}
	for name, malformed := range malformedDockerArchives(t, data) {
		t.Run(name, func(t *testing.T) {
			defer func() {
				if recovered := recover(); recovered != nil {
					t.Errorf("malformed archive panicked instead of returning invalid content: %v", recovered)
				}
			}()
			path := filepath.Join(t.TempDir(), "malformed.docker.tar")
			if err := os.WriteFile(path, malformed, 0o644); err != nil {
				t.Fatal(err)
			}
			if err := ValidateDockerArchive(path, imageID, "0.1.0-preview.1", strings.Repeat("d", 40)); !errors.Is(err, ErrInvalidContent) {
				t.Fatalf("malformed archive not refused as invalid content: %v", err)
			}
		})
	}
}

func TestAssembleBundleRejectsMalformedDockerArchiveWithoutManifest(t *testing.T) {
	repositoryRoot := filepath.Join("..", "..")
	intact := bundleInputs(t, repositoryRoot)
	if _, err := AssembleBundle(intact); err != nil {
		t.Fatalf("intact bundle refused: %v", err)
	}
	if err := ValidateBundle(intact.Output); err != nil {
		t.Fatalf("intact assembled bundle did not validate: %v", err)
	}
	data, err := os.ReadFile(intact.GameserverImageArchive)
	if err != nil {
		t.Fatal(err)
	}
	for name, malformed := range malformedDockerArchives(t, data) {
		t.Run(name, func(t *testing.T) {
			defer func() {
				if recovered := recover(); recovered != nil {
					t.Errorf("bundle assembly panicked instead of refusing the archive: %v", recovered)
				}
			}()
			input := bundleInputs(t, repositoryRoot)
			if err := os.WriteFile(input.GameserverImageArchive, malformed, 0o644); err != nil {
				t.Fatal(err)
			}
			manifest, err := AssembleBundle(input)
			if !errors.Is(err, ErrInvalidContent) || manifest.ReleaseVersion != "" {
				t.Fatalf("malformed archive did not refuse bundle assembly: manifest=%+v err=%v", manifest, err)
			}
			if _, err := os.Stat(filepath.Join(input.Output, ReleaseManifestPath)); !errors.Is(err, os.ErrNotExist) {
				t.Fatalf("failed assembly published a release manifest: %v", err)
			}
		})
	}
}

func malformedDockerArchives(t *testing.T, data []byte) map[string][]byte {
	t.Helper()
	first, err := tar.NewReader(bytes.NewReader(data)).Next()
	if err != nil || first.Size <= 1 {
		t.Fatalf("fixture needs a nonempty first regular member: header=%+v err=%v", first, err)
	}
	// The fixture uses short USTAR names without extension headers. Keep the
	// first member intact when corrupting the following header, so the decoder
	// must also handle malformed input after making progress.
	nextHeader := 512 + ((first.Size+511)/512)*512
	if int64(len(data)) < nextHeader+512 {
		t.Fatal("fixture needs a second tar header")
	}
	badFirst := bytes.Clone(data)
	badFirst[0] ^= 1
	badLater := bytes.Clone(data)
	badLater[nextHeader] ^= 1
	return map[string][]byte{
		"empty":                 nil,
		"short first header":    bytes.Clone(data[:16]),
		"bad first checksum":    badFirst,
		"truncated member body": bytes.Clone(data[:513]),
		"short later header":    bytes.Clone(data[:nextHeader+16]),
		"bad later checksum":    badLater,
	}
}

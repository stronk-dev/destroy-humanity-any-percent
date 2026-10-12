package releasepackage

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/iotest"
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

func TestValidateDockerArchiveBindsOrderedLayerContents(t *testing.T) {
	for _, compressed := range []bool{false, true} {
		for _, mutation := range []string{"intact", "changed first layer", "changed second layer", "reversed layers", "missing rootfs", "missing diff ID"} {
			t.Run(fmt.Sprintf("gzip=%t/%s", compressed, mutation), func(t *testing.T) {
				archive, imageID := layeredDockerArchive(t, compressed, mutation)
				err := ValidateDockerArchive(archive, imageID, "0.1.0-preview.1", strings.Repeat("d", 40))
				if mutation == "intact" {
					if err != nil {
						t.Fatalf("intact layered archive refused: %v", err)
					}
				} else if !errors.Is(err, ErrInvalidContent) {
					t.Fatalf("config-bound layer integrity failure accepted: %v", err)
				}
			})
		}
	}
}

func TestArchiveMemberDiffIDRejectsUnreadableCompression(t *testing.T) {
	plain := []byte("layer bytes\n")
	var buffer bytes.Buffer
	writer := gzip.NewWriter(&buffer)
	if _, err := writer.Write(plain); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	compressed := buffer.Bytes()
	for _, input := range [][]byte{plain, compressed} {
		if identity, err := archiveMemberDiffID(bytes.NewReader(input), int64(len(input))); err != nil || identity != digest(plain) {
			t.Fatalf("intact uncompressed identity mismatch: identity=%s err=%v", identity, err)
		}
	}
	badChecksum := bytes.Clone(compressed)
	badChecksum[len(badChecksum)-8] ^= 1
	readFailure := errors.New("injected layer read failure")
	for name, input := range map[string]io.Reader{
		"bad gzip checksum": bytes.NewReader(badChecksum),
		"truncated gzip":    bytes.NewReader(compressed[:len(compressed)-3]),
		"bad gzip header":   bytes.NewReader([]byte{0x1f, 0x8b, 0, 0}),
		"unsupported zstd":  bytes.NewReader(append(bytes.Clone(zstdMagic), plain...)),
		"failed read":       iotest.ErrReader(readFailure),
	} {
		t.Run(name, func(t *testing.T) {
			identity, err := archiveMemberDiffID(input, 512)
			if identity != "" || !errors.Is(err, ErrInvalidContent) {
				t.Fatalf("invalid layer returned an identity: identity=%s err=%v", identity, err)
			}
			if name == "failed read" && !errors.Is(err, readFailure) {
				t.Fatalf("source failure lost its cause: %v", err)
			}
		})
	}
}

func layeredDockerArchive(t *testing.T, compressed bool, mutation string) (string, string) {
	t.Helper()
	base, baseID := fixtureDockerArchive(t, t.TempDir(), "amd64")
	baseBytes, err := os.ReadFile(base)
	if err != nil {
		t.Fatal(err)
	}
	reader := tar.NewReader(bytes.NewReader(baseBytes))
	var config map[string]any
	for {
		header, err := reader.Next()
		if err != nil {
			t.Fatalf("fixture config missing: %v", err)
		}
		if header.Name == strings.TrimPrefix(baseID, "sha256:")+".json" {
			if err := json.NewDecoder(reader).Decode(&config); err != nil {
				t.Fatal(err)
			}
			break
		}
	}
	layer := func(value string) []byte {
		var buffer bytes.Buffer
		writer := tar.NewWriter(&buffer)
		if err := writer.WriteHeader(&tar.Header{Name: "content.txt", Mode: 0o644, Size: int64(len(value)), Typeflag: tar.TypeReg}); err != nil {
			t.Fatal(err)
		}
		if _, err := writer.Write([]byte(value)); err != nil {
			t.Fatal(err)
		}
		if err := writer.Close(); err != nil {
			t.Fatal(err)
		}
		return buffer.Bytes()
	}
	first, second := layer("first layer"), layer("second layer")
	config["rootfs"] = map[string]any{"type": "layers", "diff_ids": []string{digest(first), digest(second)}}
	switch mutation {
	case "changed first layer":
		first = layer("wrong first")
	case "changed second layer":
		second = layer("wrong second")
	case "missing rootfs":
		delete(config, "rootfs")
	case "missing diff ID":
		config["rootfs"] = map[string]any{"type": "layers", "diff_ids": []string{digest(first)}}
	}
	if compressed {
		compress := func(data []byte) []byte {
			var buffer bytes.Buffer
			writer := gzip.NewWriter(&buffer)
			if _, err := writer.Write(data); err != nil {
				t.Fatal(err)
			}
			if err := writer.Close(); err != nil {
				t.Fatal(err)
			}
			return buffer.Bytes()
		}
		first, second = compress(first), compress(second)
	}
	configBytes, err := json.Marshal(config)
	if err != nil {
		t.Fatal(err)
	}
	imageID := digest(configBytes)
	configName := strings.TrimPrefix(imageID, "sha256:") + ".json"
	layers := []string{"layer/first.tar", "layer/second.tar"}
	if mutation == "reversed layers" {
		layers[0], layers[1] = layers[1], layers[0]
	}
	manifestBytes, err := json.Marshal([]dockerArchiveManifest{{Config: configName, Layers: layers}})
	if err != nil {
		t.Fatal(err)
	}
	var buffer bytes.Buffer
	writer := tar.NewWriter(&buffer)
	for name, data := range map[string][]byte{configName: configBytes, "manifest.json": manifestBytes, "layer/first.tar": first, "layer/second.tar": second} {
		if err := writer.WriteHeader(&tar.Header{Name: name, Mode: 0o644, Size: int64(len(data)), Typeflag: tar.TypeReg}); err != nil {
			t.Fatal(err)
		}
		if _, err := writer.Write(data); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "layered.docker.tar")
	if err := os.WriteFile(path, buffer.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
	return path, imageID
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

package releasepackage

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSecretScanRejectsSeededTrackedAndImageMaterial(t *testing.T) {
	seed := strings.Join([]string{"CLOUD_CLICKER_", "SECRET_SCAN_SENTINEL_", "abcdefghijklmnop"}, "")
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "source.go"), []byte("package fixture\n// "+seed+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	findings, err := ScanTrackedFiles(root, []string{"source.go"})
	if err != nil || len(findings) != 1 || findings[0].Rule != "seeded-fixture" {
		t.Fatalf("tracked finding=%+v err=%v", findings, err)
	}
	if err := RequireNoSecrets(findings); !errors.Is(err, ErrInvalidContent) {
		t.Fatalf("seeded tracked secret accepted: %v", err)
	}

	archive, _ := fixtureDockerArchive(t, root, "amd64")
	data, err := os.ReadFile(archive)
	if err != nil {
		t.Fatal(err)
	}
	data = append(data, []byte(seed)...)
	if err := os.WriteFile(archive, data, 0o644); err != nil {
		t.Fatal(err)
	}
	findings, err = ScanDockerArchive(archive)
	if err == nil || len(findings) != 0 {
		// Appending after the tar end marker is deliberately malformed and must
		// fail loud rather than silently exclude the bytes.
		t.Fatalf("malformed seeded archive did not fail closed findings=%+v err=%v", findings, err)
	}
}

func TestCurrentTrackedTreeHasNoRecognizedSecretMaterial(t *testing.T) {
	root := filepath.Join("..", "..")
	paths := []string{"deployment/.env.example", "deployment/compose.template.yml", "deployment/Dockerfile.gameserver"}
	findings, err := ScanTrackedFiles(root, paths)
	if err != nil || len(findings) != 0 {
		t.Fatalf("release inputs findings=%+v err=%v", findings, err)
	}
}

func TestSecretScanOpensCompressedAndNestedImageLayers(t *testing.T) {
	seed := strings.Join([]string{"CLOUD_CLICKER_", "SECRET_SCAN_SENTINEL_", "abcdefghijklmnop"}, "")
	tarBytes := func(files map[string][]byte) []byte {
		t.Helper()
		var buffer bytes.Buffer
		writer := tar.NewWriter(&buffer)
		for name, data := range files {
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
		return buffer.Bytes()
	}
	gzipBytes := func(data []byte) []byte {
		t.Helper()
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
	archive := func(layer []byte) string {
		t.Helper()
		path := filepath.Join(t.TempDir(), "gameserver.docker.tar")
		if err := os.WriteFile(path, tarBytes(map[string][]byte{"manifest.json": []byte("[]\n"),
			"blobs/sha256/" + strings.Repeat("c", 64): layer}), 0o644); err != nil {
			t.Fatal(err)
		}
		return path
	}
	secretFile := []byte("key=" + seed + "\n")
	for name, layer := range map[string][]byte{
		"gzip layer":           gzipBytes(tarBytes(map[string][]byte{"opt/cloud-clicker/content/leaked.txt": secretFile})),
		"plain tar layer":      tarBytes(map[string][]byte{"opt/cloud-clicker/content/leaked.txt": secretFile}),
		"gzip file in layer":   gzipBytes(tarBytes(map[string][]byte{"usr/share/doc/leaked.txt.gz": gzipBytes(secretFile)})),
		"private key in layer": gzipBytes(tarBytes(map[string][]byte{"etc/key.pem": []byte(strings.Join([]string{"-----BEGIN ", "PRIVATE KEY-----\n"}, ""))})),
	} {
		t.Run(name, func(t *testing.T) {
			findings, err := ScanDockerArchive(archive(layer))
			if err != nil || len(findings) == 0 {
				t.Fatalf("secret inside image layer not found: findings=%+v err=%v", findings, err)
			}
			if err := RequireNoSecrets(findings); !errors.Is(err, ErrInvalidContent) {
				t.Fatalf("layer secret accepted: %v", err)
			}
		})
	}
	clean, err := ScanDockerArchive(archive(gzipBytes(tarBytes(map[string][]byte{"opt/cloud-clicker/content/ok.txt": []byte("ordinary\n")}))))
	if err != nil || len(clean) != 0 {
		t.Fatalf("clean gzip layer findings=%+v err=%v", clean, err)
	}
	if _, err := ScanDockerArchive(archive(append([]byte{0x28, 0xb5, 0x2f, 0xfd}, []byte("zstd frame")...))); !errors.Is(err, ErrInvalidContent) {
		t.Fatalf("unscannable zstd layer accepted: %v", err)
	}
	if _, err := ScanDockerArchive(archive([]byte{0x1f, 0x8b, 0x08, 0x00, 0x01})); !errors.Is(err, ErrInvalidContent) {
		t.Fatalf("corrupt gzip layer accepted: %v", err)
	}
}

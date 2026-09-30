package releasepackage

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
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

func TestSecretScanOpensLegacyTarNestedGzip(t *testing.T) {
	secret := strings.Join([]string{"CLOUD_CLICKER_", "SECRET_SCAN_SENTINEL_", "abcdefghijklmnop"}, "")
	var compressed bytes.Buffer
	gz, err := gzip.NewWriterLevel(&compressed, gzip.BestCompression)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := gz.Write([]byte("key=" + secret + "\n")); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(compressed.Bytes(), []byte(secret)) {
		t.Fatal("gzip fixture exposes the sentinel in raw bytes")
	}
	var inner bytes.Buffer
	writer := tar.NewWriter(&inner)
	if err := writer.WriteHeader(&tar.Header{Name: "nested/leak.gz", Mode: 0o644, Size: int64(compressed.Len()), Typeflag: tar.TypeReg}); err != nil {
		t.Fatal(err)
	}
	if _, err := writer.Write(compressed.Bytes()); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	archive := func(layer []byte) string {
		t.Helper()
		var outer bytes.Buffer
		writer := tar.NewWriter(&outer)
		if err := writer.WriteHeader(&tar.Header{Name: "blobs/sha256/layer", Mode: 0o644, Size: int64(len(layer)), Typeflag: tar.TypeReg}); err != nil {
			t.Fatal(err)
		}
		if _, err := writer.Write(layer); err != nil {
			t.Fatal(err)
		}
		if err := writer.Close(); err != nil {
			t.Fatal(err)
		}
		path := filepath.Join(t.TempDir(), "image.tar")
		if err := os.WriteFile(path, outer.Bytes(), 0o600); err != nil {
			t.Fatal(err)
		}
		return path
	}
	if findings, err := ScanDockerArchive(archive(inner.Bytes())); err != nil || len(findings) == 0 {
		t.Fatalf("USTAR control not found: findings=%v err=%v", findings, err)
	}
	legacy := append([]byte(nil), inner.Bytes()...)
	if string(legacy[257:262]) != "ustar" {
		t.Fatal("inner fixture lacks USTAR magic")
	}
	for index := 257; index < 265; index++ {
		legacy[index] = 0
	}
	if _, err := tar.NewReader(bytes.NewReader(legacy)).Next(); err == nil {
		t.Fatal("corrupted-checksum control was unexpectedly valid")
	}
	for index := 148; index < 156; index++ {
		legacy[index] = ' '
	}
	sum := 0
	for _, value := range legacy[:512] {
		sum += int(value)
	}
	copy(legacy[148:156], []byte(fmt.Sprintf("%06o\x00 ", sum)))
	reader := tar.NewReader(bytes.NewReader(legacy))
	if _, err := reader.Next(); err != nil {
		t.Fatalf("V7 tar header rejected by Go reader: %v", err)
	}
	if data, err := io.ReadAll(reader); err != nil || !bytes.Equal(data, compressed.Bytes()) {
		t.Fatalf("V7 tar payload rejected: err=%v", err)
	}
	if findings, err := ScanDockerArchive(archive(legacy)); err != nil || len(findings) == 0 {
		t.Fatalf("Go-readable V7 nested gzip not found: findings=%v err=%v", findings, err)
	}
}

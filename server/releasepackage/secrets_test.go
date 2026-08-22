package releasepackage

import (
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

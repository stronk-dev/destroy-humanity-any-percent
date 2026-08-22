package releasepackage

import (
	"archive/tar"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCurrentMigrationIsContiguous(t *testing.T) {
	current, err := CurrentMigration(filepath.Join("..", ".."))
	if err != nil || current != 74 {
		t.Fatalf("migration=%d err=%v", current, err)
	}
}

func TestReleaseManifestBindsEveryBundleByteAndImageSBOM(t *testing.T) {
	root := releaseBundleFixture(t)
	closure := Closure{EpochID: 8, ConstantsHash: "sha256:" + strings.Repeat("a", 64), CopyHash: "sha256:" + strings.Repeat("b", 64), Files: []File{{Path: "fixture", SHA256: "sha256:" + strings.Repeat("c", 64)}}}
	images := fixtureManifestImages(t, root)
	manifest, encoded, err := BuildReleaseManifest(root, ManifestInput{ReleaseVersion: "0.1.0-preview.1", SourceCommit: strings.Repeat("d", 40),
		DockerEngineVersion: "28.3.3", DockerComposeVersion: "2.39.1", DatabaseMigration: 74, Closure: closure, Images: images})
	if err != nil || len(manifest.Artifacts) == 0 {
		t.Fatalf("manifest=%+v err=%v", manifest, err)
	}
	if err := os.WriteFile(filepath.Join(root, ReleaseManifestPath), encoded, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := ValidateBundle(root); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "site", "index.html"), []byte("tampered"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := ValidateBundle(root); !errors.Is(err, ErrInvalidContent) {
		t.Fatalf("tampered bundle accepted: %v", err)
	}
}

func TestReleaseManifestRejectsMutableImageAndWrongSBOMHash(t *testing.T) {
	root := releaseBundleFixture(t)
	closure := Closure{EpochID: 1, ConstantsHash: "sha256:" + strings.Repeat("a", 64), CopyHash: "sha256:" + strings.Repeat("b", 64), Files: []File{{Path: "fixture", SHA256: "sha256:" + strings.Repeat("c", 64)}}}
	images := fixtureManifestImages(t, root)
	images[0].Reference = "caddy:2-alpine"
	if _, _, err := BuildReleaseManifest(root, ManifestInput{ReleaseVersion: "0.1.0", SourceCommit: strings.Repeat("d", 40), DockerEngineVersion: "28", DockerComposeVersion: "2", DatabaseMigration: 74, Closure: closure, Images: images}); !errors.Is(err, ErrInvalidContent) {
		t.Fatalf("mutable image accepted: %v", err)
	}
	images = fixtureManifestImages(t, root)
	images[1].SBOMSHA256 = "sha256:" + strings.Repeat("f", 64)
	if _, _, err := BuildReleaseManifest(root, ManifestInput{ReleaseVersion: "0.1.0", SourceCommit: strings.Repeat("d", 40), DockerEngineVersion: "28", DockerComposeVersion: "2", DatabaseMigration: 74, Closure: closure, Images: images}); !errors.Is(err, ErrInvalidContent) {
		t.Fatalf("wrong SBOM hash accepted: %v", err)
	}
}

func releaseBundleFixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	files := map[string]string{
		".env.example": "PUBLIC=example\n", "Caddyfile": "fixture\n", "Dockerfile.gameserver": "fixture\n", "LICENSE": "fixture\n",
		"compose.yml": "fixture\n", "compose.rotation.yml": "fixture\n", "config.schema.json": "{}\n", "release-manifest.schema.json": "{}\n",
		"sbom/application.spdx.json": fixtureSPDX("application"), "sbom/caddy.spdx.json": fixtureSPDX("caddy"), "sbom/gameserver.spdx.json": fixtureSPDX("gameserver"), "sbom/postgres.spdx.json": fixtureSPDX("postgres"),
		"third-party-licenses.txt": "fixture\n", "site/index.html": "<html></html>\n", "site/third-party-licenses.txt": "fixture\n",
		"gameserver": "binary\n", "content/balance/epochs/phase0.json": "{}\n",
	}
	for path, value := range files {
		target := filepath.Join(root, filepath.FromSlash(path))
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(target, []byte(value), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	for _, name := range []string{"config.schema.json", "release-manifest.schema.json"} {
		data, err := os.ReadFile(filepath.Join("..", "..", "deployment", name))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, name), data, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	archive, _ := fixtureDockerArchive(t, t.TempDir(), "amd64")
	if err := os.MkdirAll(filepath.Join(root, "images"), 0o755); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(archive)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "images", "gameserver.docker.tar"), data, 0o644); err != nil {
		t.Fatal(err)
	}
	return root
}

func fixtureManifestImages(t *testing.T, root string) []Image {
	t.Helper()
	result := []Image{}
	for index, name := range []string{"caddy", "gameserver", "postgres"} {
		path := "sbom/" + name + ".spdx.json"
		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(path)))
		if err != nil {
			t.Fatal(err)
		}
		result = append(result, Image{Name: name, Reference: name + ":fixture@sha256:" + strings.Repeat(string(rune('a'+index)), 64), SBOMPath: path, SBOMSHA256: digest(data)})
	}
	archive, err := os.Open(filepath.Join(root, "images", "gameserver.docker.tar"))
	if err != nil {
		t.Fatal(err)
	}
	reader := tar.NewReader(archive)
	for {
		header, readErr := reader.Next()
		if readErr != nil {
			break
		}
		if strings.HasSuffix(header.Name, ".json") && header.Name != "manifest.json" {
			result[1].Reference = "sha256:" + strings.TrimSuffix(header.Name, ".json")
			break
		}
	}
	_ = archive.Close()
	return result
}

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
	if err != nil || current != 79 {
		t.Fatalf("migration=%d err=%v", current, err)
	}
}

func TestReleaseManifestBindsEveryBundleByteAndImageSBOM(t *testing.T) {
	root := releaseBundleFixture(t)
	closure, err := DeriveRuntimeClosure(filepath.Join(root, "content"))
	if err != nil {
		t.Fatal(err)
	}
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
	originalCaddy, err := os.ReadFile(filepath.Join(root, "Caddyfile"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "Caddyfile"), append(originalCaddy, []byte("\n# /metrics\n")...), 0o644); err != nil {
		t.Fatal(err)
	}
	_, rebound, err := BuildReleaseManifest(root, ManifestInput{ReleaseVersion: "0.1.0-preview.1", SourceCommit: strings.Repeat("d", 40),
		DockerEngineVersion: "28.3.3", DockerComposeVersion: "2.39.1", DatabaseMigration: 74, Closure: closure, Images: images})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ReleaseManifestPath), rebound, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := ValidateBundle(root); !errors.Is(err, ErrInvalidContent) {
		t.Fatalf("manifest-rebound public metrics route accepted: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "Caddyfile"), originalCaddy, 0o644); err != nil {
		t.Fatal(err)
	}
	_, encoded, err = BuildReleaseManifest(root, ManifestInput{ReleaseVersion: "0.1.0-preview.1", SourceCommit: strings.Repeat("d", 40),
		DockerEngineVersion: "28.3.3", DockerComposeVersion: "2.39.1", DatabaseMigration: 74, Closure: closure, Images: images})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ReleaseManifestPath), encoded, 0o644); err != nil {
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
	images = fixtureManifestImages(t, root)
	images = append(images[:5], images[6:]...)
	if _, _, err := BuildReleaseManifest(root, ManifestInput{ReleaseVersion: "0.1.0", SourceCommit: strings.Repeat("d", 40), DockerEngineVersion: "28", DockerComposeVersion: "2", DatabaseMigration: 74, Closure: closure, Images: images}); !errors.Is(err, ErrInvalidContent) {
		t.Fatalf("missing Prometheus image and SBOM accepted: %v", err)
	}
}

func releaseBundleFixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	files := map[string]string{
		".env.example": "PUBLIC=example\n", "Caddyfile": "fixture\n", "Dockerfile.gameserver": "fixture\n", "LICENSE": "fixture\n",
		"compose.yml": "fixture\n", "compose.rotation.yml": "fixture\n", "config.schema.json": "{}\n", "release-manifest.schema.json": "{}\n", "rehearsal-evidence.schema.json": "{}\n",
		"sbom/application.spdx.json": fixtureSPDX("application"),
		"third-party-licenses.txt":   "fixture\n", "site/index.html": "<html></html>\n", "site/third-party-licenses.txt": "fixture\n",
		"gameserver": "binary\n", "deployment-backup": "binary\n", "deployment-release": "binary\n", "deployment-operations": "binary\n", "deployment-rehearsal": "binary\n",
		"operations/prometheus.yml": "fixture\n", "operations/cloud-clicker-alerts.yml": "fixture\n", "operations/cloud-clicker-alerts.test.yml": "fixture\n",
		"operations/alertmanager.example.yml": "fixture\n", "operations/journald.template.conf": "fixture\n",
		"operations/cloud-clicker-observe.service": "fixture\n", "operations/cloud-clicker-observe.timer": "fixture\n",
		"operations/operations.env.example": "fixture\n",
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
	for _, name := range []string{"config.schema.json", "release-manifest.schema.json", "rehearsal-evidence.schema.json"} {
		data, err := os.ReadFile(filepath.Join("..", "..", "deployment", name))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, name), data, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	for _, name := range []string{"Caddyfile", "Dockerfile.gameserver"} {
		data, err := os.ReadFile(filepath.Join("..", "..", "deployment", name))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, name), data, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.RemoveAll(filepath.Join(root, "operations")); err != nil {
		t.Fatal(err)
	}
	if err := copyTree(filepath.Join("..", "..", "deployment", "operations"), filepath.Join(root, "operations")); err != nil {
		t.Fatal(err)
	}
	if _, err := StageRuntimeContent(filepath.Join("..", ".."), filepath.Join(root, "content")); err != nil {
		t.Fatal(err)
	}
	archive, gameserverID := fixtureDockerArchive(t, t.TempDir(), "amd64")
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
	configIDs := map[string]string{}
	for index, name := range releaseImageNames {
		configIDs[name] = "sha256:" + strings.Repeat(string(rune('1'+index)), 64)
	}
	configIDs["gameserver"] = gameserverID
	for _, name := range releaseImageNames {
		path := filepath.Join(root, "sbom", name+".spdx.json")
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(fixtureSPDX(strings.ReplaceAll(configIDs[name], ":", "-"))), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	images := fixtureManifestImages(t, root)
	imageReferences := map[string]string{}
	for _, image := range images {
		imageReferences[image.Name] = image.Reference
	}
	template, err := os.ReadFile(filepath.Join("..", "..", "deployment", "compose.template.yml"))
	if err != nil {
		t.Fatal(err)
	}
	compose, err := RenderCompose(template, imageReferences)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "compose.yml"), compose, 0o644); err != nil {
		t.Fatal(err)
	}
	return root
}

func fixtureManifestImages(t *testing.T, root string) []Image {
	t.Helper()
	result := []Image{}
	for index, name := range releaseImageNames {
		path := "sbom/" + name + ".spdx.json"
		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(path)))
		if err != nil {
			t.Fatal(err)
		}
		result = append(result, Image{Name: name, Reference: name + ":fixture@sha256:" + strings.Repeat(string(rune('a'+index)), 64), RuntimeConfigSHA256: "sha256:" + strings.Repeat(string(rune('1'+index)), 64), SBOMPath: path, SBOMSHA256: digest(data)})
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
			result[2].Reference = "sha256:" + strings.TrimSuffix(header.Name, ".json")
			result[2].RuntimeConfigSHA256 = result[2].Reference
			break
		}
	}
	_ = archive.Close()
	return result
}

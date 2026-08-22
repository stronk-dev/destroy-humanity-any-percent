package releasepackage

import (
	"archive/tar"
	"encoding/binary"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAssembleBundleBindsBuiltInputsWithoutCheckout(t *testing.T) {
	repositoryRoot := filepath.Join("..", "..")
	inputs := bundleInputs(t, repositoryRoot)
	manifest, err := AssembleBundle(inputs)
	if err != nil {
		t.Fatal(err)
	}
	if manifest.Platform != "linux/amd64" || len(manifest.Images) != 3 || len(manifest.Artifacts) < 30 {
		t.Fatalf("manifest did not bind release inputs: %+v", manifest)
	}
	if err := ValidateBundle(inputs.Output); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(inputs.Output, "third-party-licenses.txt")); err != nil {
		t.Fatal(err)
	}
	if err := ValidateBundle(inputs.Output); !errors.Is(err, ErrInvalidContent) {
		t.Fatalf("missing attribution accepted: %v", err)
	}
	if _, err := AssembleBundle(inputs); !errors.Is(err, ErrInvalidContent) {
		t.Fatalf("nonempty output accepted: %v", err)
	}
}

func TestAssembleBundleRejectsWrongArchitectureAndClientSymlink(t *testing.T) {
	repositoryRoot := filepath.Join("..", "..")
	inputs := bundleInputs(t, repositoryRoot)
	binaryBytes, err := os.ReadFile(inputs.ServerBinary)
	if err != nil {
		t.Fatal(err)
	}
	binary.LittleEndian.PutUint16(binaryBytes[18:20], 183)
	if err := os.WriteFile(inputs.ServerBinary, binaryBytes, 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := AssembleBundle(inputs); !errors.Is(err, ErrInvalidContent) {
		t.Fatalf("arm64 binary accepted: %v", err)
	}

	inputs = bundleInputs(t, repositoryRoot)
	inputs.SourceCommit = strings.Repeat("e", 40)
	if _, err := AssembleBundle(inputs); !errors.Is(err, ErrInvalidContent) {
		t.Fatalf("image from another source commit accepted: %v", err)
	}

	inputs = bundleInputs(t, repositoryRoot)
	if err := os.Symlink("index.html", filepath.Join(inputs.ClientDist, "alias.html")); err != nil {
		t.Fatal(err)
	}
	if _, err := AssembleBundle(inputs); !errors.Is(err, ErrInvalidContent) {
		t.Fatalf("client symlink accepted: %v", err)
	}
}

func bundleInputs(t *testing.T, repositoryRoot string) BundleInput {
	t.Helper()
	base := t.TempDir()
	serverBinary := filepath.Join(base, "gameserver")
	binaryBytes := make([]byte, 64)
	copy(binaryBytes, []byte("\x7fELF"))
	binaryBytes[4], binaryBytes[5] = 2, 1
	binary.LittleEndian.PutUint16(binaryBytes[16:18], 2)
	binary.LittleEndian.PutUint16(binaryBytes[18:20], 62)
	if err := os.WriteFile(serverBinary, binaryBytes, 0o755); err != nil {
		t.Fatal(err)
	}
	client := filepath.Join(base, "client")
	metadata := filepath.Join(base, "metadata")
	if err := os.MkdirAll(client, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(metadata, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(client, "index.html"), []byte("<html>fixture</html>\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	for name, data := range map[string]string{"third-party-licenses.txt": "license fixture\n", "sbom.spdx.json": fixtureSPDX("application")} {
		if err := os.WriteFile(filepath.Join(metadata, name), []byte(data), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	images, configIDs, sboms := map[string]string{}, map[string]string{}, map[string]string{}
	for index, name := range []string{"caddy", "gameserver", "postgres"} {
		images[name] = name + ":fixture@sha256:" + strings.Repeat(string(rune('a'+index)), 64)
		configIDs[name] = "sha256:" + strings.Repeat(string(rune('1'+index)), 64)
	}
	archive, imageID := fixtureDockerArchive(t, base, "amd64")
	images["gameserver"], configIDs["gameserver"] = imageID, imageID
	for _, name := range []string{"caddy", "gameserver", "postgres"} {
		sboms[name] = filepath.Join(base, name+".spdx.json")
		if err := os.WriteFile(sboms[name], []byte(fixtureSPDX(strings.ReplaceAll(configIDs[name], ":", "-"))), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return BundleInput{RepositoryRoot: repositoryRoot, Output: filepath.Join(base, "bundle"), ServerBinary: serverBinary,
		ClientDist: client, MetadataDirectory: metadata, GameserverImageArchive: archive, ReleaseVersion: "0.1.0-preview.1", SourceCommit: strings.Repeat("d", 40),
		DockerEngineVersion: "28.3.3", DockerComposeVersion: "2.39.1", Images: images, ImageConfigIDs: configIDs, ImageSBOMs: sboms}
}

func fixtureSPDX(name string) string {
	return "{\"spdxVersion\":\"SPDX-2.3\",\"dataLicense\":\"CC0-1.0\",\"SPDXID\":\"SPDXRef-DOCUMENT\",\"name\":\"" + name + "\",\"documentNamespace\":\"https://example.test/sbom/" + name + "\",\"creationInfo\":{\"created\":\"2026-08-22T12:00:00Z\",\"creators\":[\"Tool: fixture\"]}}\n"
}

func fixtureDockerArchive(t *testing.T, directory, architecture string) (string, string) {
	t.Helper()
	config := []byte("{\"architecture\":\"" + architecture + "\",\"os\":\"linux\",\"config\":{\"User\":\"65532:65532\",\"Entrypoint\":[\"/usr/local/bin/gameserver\"],\"Labels\":{\"org.opencontainers.image.version\":\"0.1.0-preview.1\",\"org.opencontainers.image.revision\":\"dddddddddddddddddddddddddddddddddddddddd\",\"org.opencontainers.image.source\":\"https://github.com/stronk-dev/destroy-humanity-any-percent\",\"org.opencontainers.image.licenses\":\"MIT\"}}}\n")
	imageID := digest(config)
	configName := strings.TrimPrefix(imageID, "sha256:") + ".json"
	manifest := []byte("[{\"Config\":\"" + configName + "\",\"RepoTags\":[\"cloud-clicker/gameserver:fixture\"],\"Layers\":[\"layer/layer.tar\"]}]\n")
	path := filepath.Join(directory, "gameserver.docker.tar")
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	writer := tar.NewWriter(file)
	for name, data := range map[string][]byte{configName: config, "manifest.json": manifest, "layer/layer.tar": []byte("fixture layer\n")} {
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
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	return path, imageID
}

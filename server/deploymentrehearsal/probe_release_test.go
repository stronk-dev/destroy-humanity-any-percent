package deploymentrehearsal

import (
	"archive/tar"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"cloud-clicker/server/deploymentrelease"
	"cloud-clicker/server/releasepackage"
)

// These are real assembler/validator inputs, but synthetic ELF/image/client
// fixtures, not runnable release artifacts or a clean-host rehearsal.
func TestR18ReleaseNegativeProbesUseValidatedInputs(t *testing.T) {
	for _, population := range []string{"irreversible_or_down_migration", "missing_previous_image_or_backup"} {
		t.Run(population, func(t *testing.T) {
			request := releaseProbeFixture(t, "0.1.0-preview.2", "0.1.0-preview.1")
			request.Population = population
			originals := map[string][]byte{}
			for _, root := range []string{request.CandidateBundle, request.PreviousBundle} {
				originals[root] = readProbeInput(t, filepath.Join(root, releasepackage.ReleaseManifestPath))
			}
			if outcome, err := RunProbe(request); err != nil || outcome != ProbeRejected {
				t.Fatalf("named production negative did not reject: outcome=%d err=%v", outcome, err)
			}
			for root, original := range originals {
				if _, err := deploymentrelease.LoadBundle(root); err != nil {
					t.Fatalf("probe changed a validated bundle: %v", err)
				}
				if !bytes.Equal(original, readProbeInput(t, filepath.Join(root, releasepackage.ReleaseManifestPath))) {
					t.Fatal("probe rebound a retained fixture manifest")
				}
			}
			assertProbeWorkEmpty(t, request.WorkDirectory)

			// Invalid baseline input is not successful execution of a negative.
			roots := []string{request.PreviousBundle}
			if population == "irreversible_or_down_migration" {
				roots = append(roots, request.CandidateBundle)
			}
			for _, root := range roots {
				for _, name := range []string{"missing_manifest", "tampered_client"} {
					t.Run(filepath.Base(root)+"/"+name, func(t *testing.T) {
						path := filepath.Join(root, releasepackage.ReleaseManifestPath)
						if name == "tampered_client" {
							path = filepath.Join(root, "site", "index.html")
						}
						original := readProbeInput(t, path)
						t.Cleanup(func() { writeProbeInput(t, path, original, 0o644) })
						if name == "missing_manifest" {
							if err := os.Remove(path); err != nil {
								t.Fatal(err)
							}
						} else {
							writeProbeInput(t, path, []byte("changed after manifest binding\n"), 0o644)
						}
						outcome, err := RunProbe(request)
						if outcome == ProbeRejected || !errors.Is(err, deploymentrelease.ErrInvalid) {
							t.Fatalf("invalid input masqueraded as named rejection: outcome=%d err=%v", outcome, err)
						}
						assertProbeWorkEmpty(t, request.WorkDirectory)
					})
				}
			}
		})
	}
}

func TestR18MigrationProbeRequiresForwardCompatibleControl(t *testing.T) {
	for _, version := range []string{"0.1.0-preview.1", "0.1.0-preview.0"} {
		t.Run(version, func(t *testing.T) {
			request := releaseProbeFixture(t, version, "0.1.0-preview.1")
			request.Population = "irreversible_or_down_migration"
			outcome, err := RunProbe(request)
			if outcome == ProbeRejected || !errors.Is(err, ErrInvalid) {
				t.Fatalf("non-forward control masqueraded as migration rejection: outcome=%d err=%v", outcome, err)
			}
			assertProbeWorkEmpty(t, request.WorkDirectory)
		})
	}
}

func assertProbeWorkEmpty(t *testing.T, root string) {
	t.Helper()
	entries, err := os.ReadDir(root)
	if err != nil || len(entries) != 0 {
		t.Fatalf("probe work not cleaned: entries=%v err=%v", entries, err)
	}
}

func readProbeInput(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func writeProbeInput(t *testing.T, path string, data []byte, mode os.FileMode) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, mode); err != nil {
		t.Fatal(err)
	}
}

func releaseProbeFixture(t *testing.T, candidateVersion, previousVersion string) ProbeRequest {
	t.Helper()
	request := validProbeDirectories(t, "")
	for index, root := range []string{request.CandidateBundle, request.PreviousBundle} {
		version, commit := candidateVersion, strings.Repeat("c", 40)
		if index == 1 {
			version, commit = previousVersion, strings.Repeat("d", 40)
		}
		inputRoot := t.TempDir()
		binaryPath := filepath.Join(inputRoot, "fixture.elf")
		binaryBytes := make([]byte, 64)
		copy(binaryBytes, "\x7fELF")
		binaryBytes[4], binaryBytes[5] = 2, 1
		binary.LittleEndian.PutUint16(binaryBytes[16:18], 2)
		binary.LittleEndian.PutUint16(binaryBytes[18:20], 62)
		writeProbeInput(t, binaryPath, binaryBytes, 0o755)
		client, metadata := filepath.Join(inputRoot, "client"), filepath.Join(inputRoot, "metadata")
		writeProbeInput(t, filepath.Join(client, "index.html"), []byte("<html>probe fixture</html>\n"), 0o644)
		writeProbeInput(t, filepath.Join(metadata, "third-party-licenses.txt"), []byte("fixture attribution\n"), 0o644)
		writeProbeInput(t, filepath.Join(metadata, "sbom.spdx.json"), probeSPDX("application"), 0o644)

		config, err := json.Marshal(map[string]any{"architecture": "amd64", "os": "linux", "config": map[string]any{
			"User": "65532:65532", "Entrypoint": []string{"/usr/local/bin/gameserver"}, "Labels": map[string]string{
				"org.opencontainers.image.version": version, "org.opencontainers.image.revision": commit,
				"org.opencontainers.image.source":   "https://github.com/stronk-dev/destroy-humanity-any-percent",
				"org.opencontainers.image.licenses": "MIT"}}})
		if err != nil {
			t.Fatal(err)
		}
		imageID := hashBytes(config)
		configName := strings.TrimPrefix(imageID, "sha256:") + ".json"
		archiveManifest, err := json.Marshal([]map[string]any{{"Config": configName, "RepoTags": []string{"cloud-clicker/gameserver:fixture"}, "Layers": []string{"layer/layer.tar"}}})
		if err != nil {
			t.Fatal(err)
		}
		var archive bytes.Buffer
		writer := tar.NewWriter(&archive)
		for _, entry := range []struct {
			name string
			data []byte
		}{{configName, config}, {"manifest.json", archiveManifest}, {"layer/layer.tar", []byte("synthetic fixture layer\n")}} {
			if err := writer.WriteHeader(&tar.Header{Name: entry.name, Mode: 0o644, Size: int64(len(entry.data)), Typeflag: tar.TypeReg}); err != nil {
				t.Fatal(err)
			}
			if _, err := writer.Write(entry.data); err != nil {
				t.Fatal(err)
			}
		}
		if err := writer.Close(); err != nil {
			t.Fatal(err)
		}
		archivePath := filepath.Join(inputRoot, "gameserver.docker.tar")
		writeProbeInput(t, archivePath, archive.Bytes(), 0o644)

		images, configs, sboms := map[string]string{}, map[string]string{}, map[string]string{}
		for imageIndex, name := range []string{"alertmanager", "caddy", "gameserver", "node-exporter", "postgres", "prometheus"} {
			images[name] = name + ":fixture@sha256:" + strings.Repeat(string(rune('a'+imageIndex)), 64)
			configs[name] = "sha256:" + strings.Repeat(string(rune('1'+imageIndex)), 64)
			if name == "gameserver" {
				images[name], configs[name] = imageID, imageID
			}
			sboms[name] = filepath.Join(inputRoot, name+".spdx.json")
			writeProbeInput(t, sboms[name], probeSPDX(strings.ReplaceAll(configs[name], ":", "-")), 0o644)
		}
		browserID := "sha256:" + strings.Repeat("8", 64)
		browserSBOM := filepath.Join(inputRoot, "playwright.spdx.json")
		writeProbeInput(t, browserSBOM, probeSPDX(strings.ReplaceAll(browserID, ":", "-")), 0o644)
		_, err = releasepackage.AssembleBundle(releasepackage.BundleInput{
			RepositoryRoot: filepath.Join("..", ".."), Output: root, ServerBinary: binaryPath,
			BackupBinary: binaryPath, ReleaseBinary: binaryPath, OperationsBinary: binaryPath,
			RehearsalBinary: binaryPath, BrowserBinary: binaryPath, ClientDist: client, MetadataDirectory: metadata,
			GameserverImageArchive: archivePath, ReleaseVersion: version, SourceCommit: commit,
			DockerEngineVersion: "28.3.3", DockerComposeVersion: "2.39.1",
			Images: images, ImageConfigIDs: configs, ImageSBOMs: sboms,
			BrowserImage: "mcr.microsoft.com/playwright:v1@sha256:" + strings.Repeat("9", 64), BrowserConfigID: browserID, BrowserSBOM: browserSBOM})
		if err != nil {
			t.Fatalf("assemble validated probe bundle: %v", err)
		}
		if _, err := deploymentrelease.LoadBundle(root); err != nil {
			t.Fatalf("load probe bundle through production gate: %v", err)
		}
	}
	return request
}

func probeSPDX(name string) []byte {
	return []byte(`{"spdxVersion":"SPDX-2.3","dataLicense":"CC0-1.0","SPDXID":"SPDXRef-DOCUMENT","name":"` + name + `","documentNamespace":"https://example.test/probe/` + name + `","creationInfo":{"created":"2026-10-07T00:00:00Z","creators":["Tool: fixture"]}}` + "\n")
}

package releasepackage

import (
	"archive/tar"
	"encoding/binary"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"cloud-clicker/server/save"
)

func TestAssembleBundleBindsBuiltInputsWithoutCheckout(t *testing.T) {
	repositoryRoot := filepath.Join("..", "..")
	inputs := bundleInputs(t, repositoryRoot)
	manifest, err := AssembleBundle(inputs)
	if err != nil {
		t.Fatal(err)
	}
	if manifest.Platform != "linux/amd64" || len(manifest.Images) != 6 || len(manifest.RehearsalImages) != 1 || len(manifest.Artifacts) < 40 {
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

func TestAssembleBundleRequiresCompleteRehearsalBrowserClosure(t *testing.T) {
	repositoryRoot := filepath.Join("..", "..")
	for name, mutate := range map[string]func(*BundleInput){
		"browser binary": func(input *BundleInput) { input.BrowserBinary = "" },
		"browser image":  func(input *BundleInput) { input.BrowserImage = "playwright:latest" },
		"browser config": func(input *BundleInput) { input.BrowserConfigID = "" },
		"browser SBOM":   func(input *BundleInput) { input.BrowserSBOM = "" },
	} {
		t.Run(name, func(t *testing.T) {
			inputs := bundleInputs(t, repositoryRoot)
			mutate(&inputs)
			if _, err := AssembleBundle(inputs); !errors.Is(err, ErrInvalidContent) {
				t.Fatalf("incomplete browser closure accepted: %v", err)
			}
		})
	}

	inputs := bundleInputs(t, repositoryRoot)
	if _, err := AssembleBundle(inputs); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(inputs.Output, "deployment-browser")); err != nil {
		t.Fatal(err)
	}
	if err := ValidateBundle(inputs.Output); !errors.Is(err, ErrInvalidContent) {
		t.Fatalf("bundle without browser driver accepted: %v", err)
	}
}

func TestValidateBundleRetainsPreBrowserRollbackSchema(t *testing.T) {
	inputs := bundleInputs(t, filepath.Join("..", ".."))
	if _, err := AssembleBundle(inputs); err != nil {
		t.Fatal(err)
	}
	removeRehearsalDeclaration(t, inputs.Output, true)
	if err := ValidateBundle(inputs.Output); err != nil {
		t.Fatalf("historical bundle rejected: %v", err)
	}

	// A historical schema may omit only the later optional browser declaration.
	mutateBundledSchema(t, inputs.Output, func(schema map[string]any) {
		required := schema["required"].([]any)
		schema["required"] = required[1:]
	})
	if err := ValidateBundle(inputs.Output); !errors.Is(err, ErrInvalidContent) {
		t.Fatalf("historical bundle with missing required schema field accepted: %v", err)
	}
}

func TestValidateBundleRequiresBrowserDeclarationInCurrentBundle(t *testing.T) {
	inputs := bundleInputs(t, filepath.Join("..", ".."))
	if _, err := AssembleBundle(inputs); err != nil {
		t.Fatal(err)
	}
	mutateBundledSchema(t, inputs.Output, func(schema map[string]any) {
		delete(schema["properties"].(map[string]any), "rehearsal_images")
	})
	if err := ValidateBundle(inputs.Output); !errors.Is(err, ErrInvalidContent) {
		t.Fatalf("browser bundle without rehearsal_images schema property accepted: %v", err)
	}
}

func TestValidateBundlePreservesPreviousSaveVersions(t *testing.T) {
	for name, mutate := range map[string]func(*ReleaseManifest){
		"previous company and founder": func(manifest *ReleaseManifest) {
			manifest.CompanySaveVersion--
			manifest.FounderSaveVersion--
		},
		"zero company":   func(manifest *ReleaseManifest) { manifest.CompanySaveVersion = 0 },
		"zero founder":   func(manifest *ReleaseManifest) { manifest.FounderSaveVersion = 0 },
		"future company": func(manifest *ReleaseManifest) { manifest.CompanySaveVersion++ },
		"future founder": func(manifest *ReleaseManifest) { manifest.FounderSaveVersion++ },
	} {
		t.Run(name, func(t *testing.T) {
			inputs := bundleInputs(t, filepath.Join("..", ".."))
			manifest, err := AssembleBundle(inputs)
			if err != nil {
				t.Fatal(err)
			}
			if manifest.CompanySaveVersion != save.LatestCompanyVersion || manifest.FounderSaveVersion != save.LatestFounderVersion {
				t.Fatalf("new bundle did not use current save versions: company=%d founder=%d", manifest.CompanySaveVersion, manifest.FounderSaveVersion)
			}
			mutate(&manifest)
			writeFixtureManifest(t, inputs.Output, manifest)
			err = ValidateBundle(inputs.Output)
			if name == "previous company and founder" {
				if err != nil {
					t.Fatalf("exact previous-version bundle rejected: %v", err)
				}
			} else if !errors.Is(err, ErrInvalidContent) {
				t.Fatalf("invalid save-version bundle accepted: %v", err)
			}
		})
	}
}

func removeRehearsalDeclaration(t *testing.T, root string, removeBrowserArtifacts bool) {
	t.Helper()
	manifest, _, err := LoadReleaseManifest(root)
	if err != nil {
		t.Fatal(err)
	}
	manifest.RehearsalImages = nil
	if removeBrowserArtifacts {
		for _, name := range []string{"deployment-browser", "sbom/playwright.spdx.json"} {
			if err := os.Remove(filepath.Join(root, name)); err != nil {
				t.Fatal(err)
			}
		}
		artifacts := manifest.Artifacts[:0]
		for _, artifact := range manifest.Artifacts {
			if artifact.Path != "deployment-browser" && artifact.Path != "sbom/playwright.spdx.json" {
				artifacts = append(artifacts, artifact)
			}
		}
		manifest.Artifacts = artifacts
	}
	writeFixtureManifest(t, root, manifest)
	mutateBundledSchema(t, root, func(schema map[string]any) {
		delete(schema["properties"].(map[string]any), "rehearsal_images")
	})
}

func mutateBundledSchema(t *testing.T, root string, mutate func(map[string]any)) {
	t.Helper()
	path := filepath.Join(root, "release-manifest.schema.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var schema map[string]any
	if err := json.Unmarshal(data, &schema); err != nil {
		t.Fatal(err)
	}
	mutate(schema)
	data, err = json.MarshalIndent(schema, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
	manifest, _, err := LoadReleaseManifest(root)
	if err != nil {
		t.Fatal(err)
	}
	for index := range manifest.Artifacts {
		if manifest.Artifacts[index].Path == "release-manifest.schema.json" {
			manifest.Artifacts[index].SHA256 = digest(data)
		}
	}
	writeFixtureManifest(t, root, manifest)
}

func writeFixtureManifest(t *testing.T, root string, manifest ReleaseManifest) {
	t.Helper()
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ReleaseManifestPath), append(data, '\n'), 0o644); err != nil {
		t.Fatal(err)
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
	backupBytes, err := os.ReadFile(inputs.BackupBinary)
	if err != nil {
		t.Fatal(err)
	}
	binary.LittleEndian.PutUint16(backupBytes[18:20], 183)
	if err := os.WriteFile(inputs.BackupBinary, backupBytes, 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := AssembleBundle(inputs); !errors.Is(err, ErrInvalidContent) {
		t.Fatalf("arm64 backup helper accepted: %v", err)
	}

	inputs = bundleInputs(t, repositoryRoot)
	releaseBytes, err := os.ReadFile(inputs.ReleaseBinary)
	if err != nil {
		t.Fatal(err)
	}
	binary.LittleEndian.PutUint16(releaseBytes[18:20], 183)
	if err := os.WriteFile(inputs.ReleaseBinary, releaseBytes, 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := AssembleBundle(inputs); !errors.Is(err, ErrInvalidContent) {
		t.Fatalf("arm64 release helper accepted: %v", err)
	}

	inputs = bundleInputs(t, repositoryRoot)
	operationsBytes, err := os.ReadFile(inputs.OperationsBinary)
	if err != nil {
		t.Fatal(err)
	}
	binary.LittleEndian.PutUint16(operationsBytes[18:20], 183)
	if err := os.WriteFile(inputs.OperationsBinary, operationsBytes, 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := AssembleBundle(inputs); !errors.Is(err, ErrInvalidContent) {
		t.Fatalf("arm64 operations helper accepted: %v", err)
	}

	inputs = bundleInputs(t, repositoryRoot)
	rehearsalBytes, err := os.ReadFile(inputs.RehearsalBinary)
	if err != nil {
		t.Fatal(err)
	}
	binary.LittleEndian.PutUint16(rehearsalBytes[18:20], 183)
	if err := os.WriteFile(inputs.RehearsalBinary, rehearsalBytes, 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := AssembleBundle(inputs); !errors.Is(err, ErrInvalidContent) {
		t.Fatalf("arm64 rehearsal helper accepted: %v", err)
	}

	inputs = bundleInputs(t, repositoryRoot)
	browserBytes, err := os.ReadFile(inputs.BrowserBinary)
	if err != nil {
		t.Fatal(err)
	}
	binary.LittleEndian.PutUint16(browserBytes[18:20], 183)
	if err := os.WriteFile(inputs.BrowserBinary, browserBytes, 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := AssembleBundle(inputs); !errors.Is(err, ErrInvalidContent) {
		t.Fatalf("arm64 browser helper accepted: %v", err)
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

func TestAssembleBundleRequiresRehearsalBinaryAndEvidenceSchema(t *testing.T) {
	repositoryRoot := filepath.Join("..", "..")
	inputs := bundleInputs(t, repositoryRoot)
	inputs.RehearsalBinary = ""
	if _, err := AssembleBundle(inputs); !errors.Is(err, ErrInvalidContent) {
		t.Fatalf("missing rehearsal helper accepted: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(repositoryRoot, "deployment", "rehearsal-evidence.schema.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `"guard_exhausted"`) {
		t.Fatal("rehearsal schema lost guard field")
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
	for index, name := range releaseImageNames {
		images[name] = name + ":fixture@sha256:" + strings.Repeat(string(rune('a'+index)), 64)
		configIDs[name] = "sha256:" + strings.Repeat(string(rune('1'+index)), 64)
	}
	archive, imageID := fixtureDockerArchive(t, base, "amd64")
	images["gameserver"], configIDs["gameserver"] = imageID, imageID
	for _, name := range releaseImageNames {
		sboms[name] = filepath.Join(base, name+".spdx.json")
		if err := os.WriteFile(sboms[name], []byte(fixtureSPDX(strings.ReplaceAll(configIDs[name], ":", "-"))), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	backupBinary := filepath.Join(base, "deployment-backup")
	if err := os.WriteFile(backupBinary, binaryBytes, 0o755); err != nil {
		t.Fatal(err)
	}
	releaseBinary := filepath.Join(base, "deployment-release")
	if err := os.WriteFile(releaseBinary, binaryBytes, 0o755); err != nil {
		t.Fatal(err)
	}
	operationsBinary := filepath.Join(base, "deployment-operations")
	if err := os.WriteFile(operationsBinary, binaryBytes, 0o755); err != nil {
		t.Fatal(err)
	}
	rehearsalBinary := filepath.Join(base, "deployment-rehearsal")
	if err := os.WriteFile(rehearsalBinary, binaryBytes, 0o755); err != nil {
		t.Fatal(err)
	}
	browserBinary := filepath.Join(base, "deployment-browser")
	if err := os.WriteFile(browserBinary, binaryBytes, 0o755); err != nil {
		t.Fatal(err)
	}
	browserConfigID := "sha256:" + strings.Repeat("8", 64)
	browserSBOM := filepath.Join(base, "playwright.spdx.json")
	if err := os.WriteFile(browserSBOM, []byte(fixtureSPDX(strings.ReplaceAll(browserConfigID, ":", "-"))), 0o644); err != nil {
		t.Fatal(err)
	}
	return BundleInput{RepositoryRoot: repositoryRoot, Output: filepath.Join(base, "bundle"), ServerBinary: serverBinary, BackupBinary: backupBinary, ReleaseBinary: releaseBinary, OperationsBinary: operationsBinary, RehearsalBinary: rehearsalBinary, BrowserBinary: browserBinary,
		ClientDist: client, MetadataDirectory: metadata, GameserverImageArchive: archive, ReleaseVersion: "0.1.0-preview.1", SourceCommit: strings.Repeat("d", 40),
		DockerEngineVersion: "28.3.3", DockerComposeVersion: "2.39.1", Images: images, ImageConfigIDs: configIDs, ImageSBOMs: sboms,
		BrowserImage: "mcr.microsoft.com/playwright:v1@sha256:" + strings.Repeat("9", 64), BrowserConfigID: browserConfigID, BrowserSBOM: browserSBOM}
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

func TestValidateBundleBindsManifestClaimsToComposeAndContent(t *testing.T) {
	repositoryRoot := filepath.Join("..", "..")
	forge := func(t *testing.T, mutate func(*ReleaseManifest, string)) error {
		t.Helper()
		inputs := bundleInputs(t, repositoryRoot)
		if _, err := AssembleBundle(inputs); err != nil {
			t.Fatal(err)
		}
		if err := ValidateBundle(inputs.Output); err != nil {
			t.Fatalf("unforged bundle rejected: %v", err)
		}
		manifest, _, err := LoadReleaseManifest(inputs.Output)
		if err != nil {
			t.Fatal(err)
		}
		mutate(&manifest, inputs.Output)
		encoded, err := json.MarshalIndent(manifest, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(inputs.Output, ReleaseManifestPath), append(encoded, '\n'), 0o644); err != nil {
			t.Fatal(err)
		}
		return ValidateBundle(inputs.Output)
	}
	image := func(name string) func(*ReleaseManifest, string) {
		return func(manifest *ReleaseManifest, _ string) {
			for index := range manifest.Images {
				if manifest.Images[index].Name == name {
					reference := manifest.Images[index].Reference
					manifest.Images[index].Reference = reference[:strings.LastIndex(reference, "@sha256:")] + "@sha256:" + strings.Repeat("0", 64)
				}
			}
		}
	}
	for name, mutate := range map[string]func(*ReleaseManifest, string){
		"caddy digest":      image("caddy"),
		"postgres digest":   image("postgres"),
		"prometheus digest": image("prometheus"),
		"epoch id":          func(manifest *ReleaseManifest, _ string) { manifest.EpochID += 91 },
		"constants hash": func(manifest *ReleaseManifest, _ string) {
			manifest.ConstantsHash = "sha256:" + strings.Repeat("e", 64)
		},
		"copy hash": func(manifest *ReleaseManifest, _ string) { manifest.CopyHash = "sha256:" + strings.Repeat("e", 64) },
		"removed catalog with rebound manifest": func(manifest *ReleaseManifest, root string) {
			const catalog = "content/balance/catalogs/phase0.json"
			if err := os.Remove(filepath.Join(root, catalog)); err != nil {
				t.Fatal(err)
			}
			kept := manifest.Artifacts[:0]
			for _, artifact := range manifest.Artifacts {
				if artifact.Path != catalog {
					kept = append(kept, artifact)
				}
			}
			if len(kept) == len(manifest.Artifacts) {
				t.Fatalf("fixture bundle has no %s", catalog)
			}
			manifest.Artifacts = kept
		},
	} {
		t.Run(name, func(t *testing.T) {
			if err := forge(t, mutate); !errors.Is(err, ErrInvalidContent) {
				t.Fatalf("forged manifest claim accepted: %v", err)
			}
		})
	}
}

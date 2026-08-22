package releasepackage

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"cloud-clicker/server/epochseed"
)

func TestRepositoryRuntimeClosureIsManifestDrivenAndExact(t *testing.T) {
	root := filepath.Join("..", "..")
	closure, err := DeriveRuntimeClosure(root)
	if err != nil {
		t.Fatal(err)
	}
	bundle, err := epochseed.Load(root)
	if err != nil {
		t.Fatal(err)
	}
	// One epoch declaration, every manifest-owned artifact/changelog, and the
	// three non-epoch composition/identity files.
	wantFiles := 1 + len(bundle.Seed.Artifacts) + len(bundle.Seed.Epochs) + 3
	if closure.EpochID != bundle.Seed.CurrentEpochID || len(closure.Files) != wantFiles {
		t.Fatalf("epoch=%d files=%d", closure.EpochID, len(closure.Files))
	}
	seen := map[string]bool{}
	for _, file := range closure.Files {
		seen[file.Path] = true
	}
	for _, required := range []string{
		"balance/epochs/phase0.json", "balance/catalogs/phase0.json", "balance/transport/phase0.json",
		"changelog/epoch-8.md", "moderation/guild-names.txt", "deployment/content-manifest.v1.json",
	} {
		if !seen[required] {
			t.Fatalf("runtime closure omitted %q", required)
		}
	}
	for _, forbidden := range []string{"balance/testdata/valid/minimal.json", "server/go.mod", "planning/CURRENT-STATE.md"} {
		if seen[forbidden] {
			t.Fatalf("runtime closure included non-runtime path %q", forbidden)
		}
	}
}

func TestStagedRuntimeContentRejectsMissingTamperedAndExtraFiles(t *testing.T) {
	root := filepath.Join("..", "..")
	for _, test := range []struct {
		name   string
		mutate func(string, Closure) error
	}{
		{"missing", func(destination string, closure Closure) error {
			return os.Remove(filepath.Join(destination, filepath.FromSlash(closure.Files[0].Path)))
		}},
		{"tampered", func(destination string, closure Closure) error {
			return os.WriteFile(filepath.Join(destination, filepath.FromSlash(closure.Files[0].Path)), []byte("tampered\n"), 0o644)
		}},
		{"extra", func(destination string, _ Closure) error {
			return os.WriteFile(filepath.Join(destination, "undeclared.txt"), []byte("extra\n"), 0o644)
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			destination := filepath.Join(t.TempDir(), "content")
			closure, err := StageRuntimeContent(root, destination)
			if err != nil {
				t.Fatal(err)
			}
			if err := test.mutate(destination, closure); err != nil {
				t.Fatal(err)
			}
			if err := ValidateStagedContent(destination, closure); !errors.Is(err, ErrInvalidContent) {
				t.Fatalf("invalid staged content accepted: %v", err)
			}
		})
	}
}

func TestStageRuntimeContentRefusesAStaleDestination(t *testing.T) {
	destination := t.TempDir()
	if err := os.WriteFile(filepath.Join(destination, "stale"), []byte("stale"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := StageRuntimeContent(filepath.Join("..", ".."), destination); !errors.Is(err, ErrInvalidContent) {
		t.Fatalf("stale destination accepted: %v", err)
	}
}

func TestRuntimeClosureFailsWhenADeclaredArtifactIsAbsent(t *testing.T) {
	root := filepath.Join(t.TempDir(), "repo")
	for _, directory := range []string{"balance/epochs", "balance/transport", "moderation", "deployment", "changelog"} {
		if err := os.MkdirAll(filepath.Join(root, filepath.FromSlash(directory)), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	files := map[string]string{
		"balance/epochs/phase0.json":          `{"schema_version":1,"current_epoch_id":1,"artifacts":[{"name":"missing","path":"balance/missing.json"}],"epochs":[{"epoch_id":1,"name":"fixture","changelog_ref":"changelog/epoch-1.md","accepted_hashes":["sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"]}]}`,
		"balance/transport/phase0.json":       `{}`,
		"moderation/guild-names.txt":          "blocked\n",
		"deployment/content-manifest.v1.json": `{"schema_version":1,"constants_hash":"sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","copy_hash":"sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"}`,
		"changelog/epoch-1.md":                "# fixture\n",
	}
	for path, value := range files {
		if err := os.WriteFile(filepath.Join(root, filepath.FromSlash(path)), []byte(value), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := DeriveRuntimeClosure(root); !errors.Is(err, ErrInvalidContent) {
		t.Fatalf("missing declared artifact accepted: %v", err)
	}
}

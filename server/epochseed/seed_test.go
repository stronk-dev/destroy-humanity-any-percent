package epochseed

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFixtureArtifactSetIsManifestDriven(t *testing.T) {
	root := filepath.Join("..", "..")
	bundle, err := Load(root)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"commons", "economy", "factions", "prestige", "routes"}
	if len(bundle.Artifacts) != len(want) || !Accepts(Current(bundle.Seed), bundle.Hash) {
		t.Fatalf("artifacts=%v hash=%s", bundle.Artifacts, bundle.Hash)
	}
	for _, name := range want {
		if len(bundle.Artifacts[name]) == 0 {
			t.Fatalf("missing %s", name)
		}
	}
}

func TestAddedManifestArtifactChangesCompositionWithoutCode(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "balance", "epochs"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "balance", "extra"), 0o755); err != nil {
		t.Fatal(err)
	}
	seed := []byte(`{"schema_version":1,"current_epoch_id":1,"artifacts":[{"name":"extra","path":"balance/extra/value.json"}],"epochs":[{"epoch_id":1,"name":"fixture","changelog_ref":"changelog/epoch-1.md","accepted_hashes":["sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"]}]}`)
	if err := os.WriteFile(filepath.Join(root, filepath.FromSlash(Path)), seed, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "balance", "extra", "value.json"), []byte(`{"value":1}`), 0o644); err != nil {
		t.Fatal(err)
	}
	bundle, err := Load(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(bundle.Artifacts) != 1 || string(bundle.Artifacts["extra"]) != `{"value":1}` || bundle.Hash == "" {
		t.Fatalf("bundle=%+v", bundle)
	}
}

func TestValidateRejectsForgedInProcessSeed(t *testing.T) {
	bundle, err := Load(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	bundle.Seed.Artifacts = append(bundle.Seed.Artifacts, bundle.Seed.Artifacts[0])
	if err := Validate(bundle.Seed); err == nil {
		t.Fatal("duplicate in-process artifact declaration passed")
	}
}

package replaycatalog

import (
	"path/filepath"
	"testing"

	"cloud-clicker/server/epochseed"
)

func TestLoadRequiresExactSixArtifactBundle(t *testing.T) {
	bundle, err := epochseed.Load(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := Load(bundle.Hash, bundle.Artifacts)
	if err != nil || loaded.ConstantsHash != bundle.Hash || loaded.Commons == nil {
		t.Fatalf("bundle=%+v err=%v", loaded, err)
	}
	missing := make(map[string][]byte, len(bundle.Artifacts)-1)
	for name, data := range bundle.Artifacts {
		if name != "guilds" {
			missing[name] = data
		}
	}
	if _, err := Load(bundle.Hash, missing); err == nil {
		t.Fatal("catalog bundle missing guilds artifact was accepted")
	}
	extra := make(map[string][]byte, len(bundle.Artifacts)+1)
	for name, data := range bundle.Artifacts {
		extra[name] = data
	}
	extra["future"] = []byte(`{}`)
	if _, err := Load(bundle.Hash, extra); err == nil {
		t.Fatal("catalog bundle with unregistered artifact was accepted")
	}
	relabeled := "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	if _, err := Load(relabeled, bundle.Artifacts); err == nil {
		t.Fatal("catalog bundle bytes were accepted under a false constants hash")
	}
	loaded.Artifacts["economy"] = append(loaded.Artifacts["economy"], '\n')
	if _, ok := loaded.ResolvePrestige(bundle.Hash); ok {
		t.Fatal("mutated catalog bundle remained valid under its old constants hash")
	}
}

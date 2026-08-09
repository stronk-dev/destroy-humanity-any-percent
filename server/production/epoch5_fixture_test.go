package production

import (
	"os"
	"path/filepath"
	"testing"

	"cloud-clicker/server/save"
)

const epoch5ConstantsHash = "sha256:63ab30c96b5d76b941b053131fcee63c94b6b3ad91322f9160d94973ce8c58fa"

func epoch5TestBundle(t *testing.T) CatalogBundle {
	t.Helper()
	paths := map[string]string{
		"categories": "categories.json",
		"commons":    "commons.json",
		"economy":    "economy.json",
		"factions":   "factions.json",
		"guilds":     "guilds.json",
		"prestige":   "prestige.json",
		"routes":     "routes.json",
	}
	artifacts := make(map[string][]byte, len(paths))
	for name, filename := range paths {
		data, err := os.ReadFile(filepath.Join("..", "..", "balance", "testdata", "epoch5", filename))
		if err != nil {
			t.Fatalf("read epoch-5 %s fixture: %v", name, err)
		}
		artifacts[name] = data
	}
	hash, err := save.ConstantsHashArtifacts(artifacts)
	if err != nil || hash != epoch5ConstantsHash {
		t.Fatalf("epoch-5 fixture hash=%s err=%v", hash, err)
	}
	return loadReplayTestBundle(t, hash, artifacts)
}

package replaycatalog

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"cloud-clicker/server/production"
	"cloud-clicker/server/reputation"
	"cloud-clicker/server/save"
)

// liveEpochArtifacts is the currently pinned epoch artifact set, read through
// the epoch declaration so this test follows future mints.
func liveEpochArtifacts(t *testing.T) map[string][]byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "..", "balance", "epochs", "phase0.json"))
	if err != nil {
		t.Fatal(err)
	}
	var declaration struct {
		Artifacts []struct{ Name, Path string } `json:"artifacts"`
	}
	if err := json.Unmarshal(data, &declaration); err != nil {
		t.Fatal(err)
	}
	artifacts := map[string][]byte{}
	for _, row := range declaration.Artifacts {
		artifacts[row.Name], err = os.ReadFile(filepath.Join("..", "..", row.Path))
		if err != nil {
			t.Fatal(err)
		}
	}
	return artifacts
}

func withReputationDeclaration(t *testing.T, economyBytes []byte) []byte {
	t.Helper()
	var root map[string]any
	if err := json.Unmarshal(economyBytes, &root); err != nil {
		t.Fatal(err)
	}
	root["multiplier_sources"] = append(root["multiplier_sources"].([]any), map[string]any{"id": "reputation.founder_bonus", "slot": "prestige", "target": "all", "provider": reputation.Provider})
	data, err := json.Marshal(root)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func loadArtifacts(artifacts map[string][]byte) (production.CatalogBundle, error) {
	hash, err := save.ConstantsHashArtifacts(artifacts)
	if err != nil {
		return production.CatalogBundle{}, err
	}
	return Load(hash, artifacts)
}

func TestLoadPairsReputationTreeWithItsDeclarationAndMinigameAPI(t *testing.T) {
	live := liveEpochArtifacts(t)
	tree, err := os.ReadFile(filepath.Join("..", "..", "balance", "testdata", "reputation-tree", "fixture-v1.json"))
	if err != nil {
		t.Fatal(err)
	}
	complete := map[string][]byte{}
	for name, data := range live {
		complete[name] = data
	}
	complete["economy"] = withReputationDeclaration(t, live["economy"])
	complete["reputation_tree"] = tree
	bundle, err := loadArtifacts(complete)
	if err != nil || bundle.ReputationTree == nil || len(bundle.ReputationTree.Nodes()) != 9 {
		t.Fatalf("complete reputation bundle rejected: %v", err)
	}
	set := production.ReplayCatalogSet{bundle.ConstantsHash: bundle}
	if _, ok := set.ResolveReplayCatalogs(bundle.ConstantsHash); !ok {
		t.Fatal("complete reputation bundle is not valid for replay")
	}

	treeWithoutDeclaration := map[string][]byte{}
	for name, data := range live {
		treeWithoutDeclaration[name] = data
	}
	treeWithoutDeclaration["reputation_tree"] = tree
	if _, err := loadArtifacts(treeWithoutDeclaration); err == nil {
		t.Fatal("tree without its economy declaration loaded")
	}

	declarationWithoutTree := map[string][]byte{}
	for name, data := range live {
		declarationWithoutTree[name] = data
	}
	declarationWithoutTree["economy"] = complete["economy"]
	if _, err := loadArtifacts(declarationWithoutTree); !errors.Is(err, reputation.ErrInvalidTree) {
		t.Fatalf("declaration without a tree loaded: %v", err)
	}

	withoutAPI := map[string][]byte{}
	for name, data := range complete {
		if name != "minigame_api" {
			withoutAPI[name] = data
		}
	}
	if _, err := loadArtifacts(withoutAPI); err == nil {
		t.Fatal("tree loaded without the minigame_api chain that owns Founder v21")
	}

	// valid() re-checks the pairing for bundles assembled outside the loader:
	// a live bundle stays valid until only its Economy is swapped for one that
	// declares the Reputation source.
	liveBundle, err := loadArtifacts(live)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := (production.ReplayCatalogSet{liveBundle.ConstantsHash: liveBundle}).ResolveReplayCatalogs(liveBundle.ConstantsHash); !ok {
		t.Fatal("live bundle is not valid")
	}
	forged := liveBundle
	forged.Economy = bundle.Economy
	if _, ok := (production.ReplayCatalogSet{forged.ConstantsHash: forged}).ResolveReplayCatalogs(forged.ConstantsHash); ok {
		t.Fatal("bundle declaring the Reputation source without a tree resolved")
	}
}

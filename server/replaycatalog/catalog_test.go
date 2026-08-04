package replaycatalog

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"cloud-clicker/server/epochseed"
	"cloud-clicker/server/save"
)

func TestLoadRequiresExactSevenArtifactBundle(t *testing.T) {
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
	invalidCategories := make(map[string][]byte, len(bundle.Artifacts))
	for name, data := range bundle.Artifacts {
		invalidCategories[name] = append([]byte(nil), data...)
	}
	invalidCategories["categories"] = []byte(`{"schema_version":1}`)
	invalidHash, err := save.ConstantsHashArtifacts(invalidCategories)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Load(invalidHash, invalidCategories); err == nil {
		t.Fatal("invalid pinned category artifact was accepted")
	}
	loaded.Artifacts["economy"] = append(loaded.Artifacts["economy"], '\n')
	if _, ok := loaded.ResolvePrestige(bundle.Hash); ok {
		t.Fatal("mutated catalog bundle remained valid under its old constants hash")
	}
}

func TestLoadAcceptsOnlyPairedFoundationArtifacts(t *testing.T) {
	seed, err := epochseed.Load(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	fixture, err := os.ReadFile(filepath.Join("..", "..", "balance", "testdata", "meters-catalog-parity-v1.json"))
	if err != nil {
		t.Fatal(err)
	}
	var envelope struct {
		Baseline json.RawMessage `json:"baseline"`
	}
	if err := json.Unmarshal(fixture, &envelope); err != nil {
		t.Fatal(err)
	}
	artifacts := make(map[string][]byte, len(seed.Artifacts)+2)
	for name, data := range seed.Artifacts {
		artifacts[name] = append([]byte(nil), data...)
	}
	artifacts["meters"] = envelope.Baseline
	artifacts["achievements"] = []byte(`{"schema_version":1,"achievements":[{"id":"achievement.first_gate","condition_scope":"run","condition":{"kind":"counter_at_least","counter":"tier","minimum":1},"proof":{"kind":"provenance","event_kinds":["gate_crossed"]},"score_grant":4,"copy_key":"category.any_percent"}]}`)
	hash, err := save.ConstantsHashArtifacts(artifacts)
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := Load(hash, artifacts)
	if err != nil || loaded.Meters == nil || loaded.Achievements == nil {
		t.Fatalf("active bundle=%+v err=%v", loaded, err)
	}
	delete(artifacts, "achievements")
	oneSidedHash, err := save.ConstantsHashArtifacts(artifacts)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Load(oneSidedHash, artifacts); err == nil {
		t.Fatal("one-sided foundation artifact set was accepted")
	}
}

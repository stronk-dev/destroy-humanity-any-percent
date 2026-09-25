package replaycatalog

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"cloud-clicker/server/production"
	"cloud-clicker/server/save"
)

// AR-P1/AR-P2 (Demo Disc Arcade AC1): the arcade artifact, its two
// definition rows, and their minigame_api tenants load together or not at
// all; both toys resolve to the one pinned artifact.
func TestLoadArcadeChainIsAllOrNothing(t *testing.T) {
	read := func(parts ...string) []byte {
		data, err := os.ReadFile(filepath.Join(append([]string{"..", ".."}, parts...)...))
		if err != nil {
			t.Fatal(err)
		}
		return data
	}
	complete := func() map[string][]byte {
		artifacts := pitchChainArtifacts(t)
		artifacts["minigames"] = read("testdata", "minigame", "pitch-typer-arcade-v3.json")
		artifacts["minigame_api"] = read("balance", "testdata", "minigame-api-arcade-candidate-v1.json")
		artifacts["typer"] = read("balance", "testdata", "typer-v1.json")
		artifacts["arcade"] = read("balance", "testdata", "arcade-v1.json")
		return artifacts
	}
	load := func(artifacts map[string][]byte) (production.CatalogBundle, string, error) {
		hash, err := save.ConstantsHashArtifacts(artifacts)
		if err != nil {
			t.Fatal(err)
		}
		bundle, err := Load(hash, artifacts)
		return bundle, hash, err
	}
	bundle, hash, err := load(complete())
	if err != nil || bundle.Arcade == nil || bundle.Typer == nil {
		t.Fatalf("complete arcade chain err=%v", err)
	}
	set := production.ReplayCatalogSet{hash: bundle}
	first, okMine := set.ResolveTenantContent(hash, "mine_grid", "1.0.0")
	second, okSnake := set.ResolveTenantContent(hash, "snake", "1.0.0")
	if !okMine || !okSnake || first.Hash != second.Hash || string(first.Bytes) != string(complete()["arcade"]) {
		t.Fatal("both arcade toys must resolve to the one pinned arcade artifact")
	}
	if _, ok := set.ResolveTenantContent(hash, "mine_grid", "2.0.0"); ok {
		t.Fatal("an unknown engine version must resolve nothing")
	}
	withoutStageToy := func(a map[string][]byte) {
		var value map[string]any
		_ = json.Unmarshal(a["arcade"], &value)
		stage := value["container"].(map[string]any)["stages"].([]any)[0].(map[string]any)
		stage["toys"] = []any{"arcade.mine_grid", "arcade.pinball"}
		a["arcade"], _ = json.Marshal(value)
	}
	cases := map[string]func(map[string][]byte){
		"definitions without artifact": func(a map[string][]byte) { delete(a, "arcade") },
		"artifact without definitions": func(a map[string][]byte) { a["minigames"] = read("testdata", "minigame", "pitch-typer-v3.json") },
		"definitions without tenants": func(a map[string][]byte) {
			a["minigame_api"] = read("balance", "testdata", "minigame-api-typer-candidate-v1.json")
		},
		"artifact without minigame_api": func(a map[string][]byte) { delete(a, "minigame_api"); delete(a, "typer") },
		"stage toy without definition":  withoutStageToy,
	}
	for name, mutate := range cases {
		artifacts := complete()
		mutate(artifacts)
		if _, _, err := load(artifacts); err == nil {
			t.Fatalf("%s: incomplete arcade chain loaded", name)
		}
	}
	// A bundle without the arcade still resolves Typer and nothing arcade.
	typerOnly := complete()
	delete(typerOnly, "arcade")
	typerOnly["minigames"] = read("testdata", "minigame", "pitch-typer-v3.json")
	typerOnly["minigame_api"] = read("balance", "testdata", "minigame-api-typer-candidate-v1.json")
	typerBundle, typerHash, err := load(typerOnly)
	if err != nil || typerBundle.Arcade != nil {
		t.Fatalf("typer-only bundle err=%v", err)
	}
	if _, ok := (production.ReplayCatalogSet{typerHash: typerBundle}).ResolveTenantContent(typerHash, "snake", "1.0.0"); ok {
		t.Fatal("a bundle without the arcade resolved snake content")
	}
}

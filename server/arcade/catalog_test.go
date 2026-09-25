package arcade

import (
	"encoding/json"
	"errors"
	"os"
	"testing"

	"cloud-clicker/server/copykeys"
)

const (
	candidatePath = "../../balance/testdata/arcade-v1.json"
	fixturePath   = "../../testdata/arcade/corpus-fixture-v1.json"
)

func readFile(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func declarations() Declarations {
	keys := map[string]struct{}{}
	for _, key := range copykeys.All() {
		keys[key] = struct{}{}
	}
	return Declarations{CopyKeys: keys}
}

func mutate(t *testing.T, path string, change func(map[string]any)) []byte {
	t.Helper()
	var value map[string]any
	if err := json.Unmarshal(readFile(t, path), &value); err != nil {
		t.Fatal(err)
	}
	change(value)
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestArcadeCatalogLoadsCandidateAndFixture(t *testing.T) {
	candidate, err := LoadCatalog(readFile(t, candidatePath), declarations())
	if err != nil {
		t.Fatalf("candidate must load: %v", err)
	}
	if len(candidate.Container.Stages) != 1 || candidate.Container.Stages[0].StageID != "cover_disc" ||
		len(candidate.Container.Stages[0].Toys) != 2 || candidate.Snake.Width != 20 || len(candidate.MineGrid.Presets) != 3 {
		t.Fatalf("candidate does not carry the AR1.2/AR3.1/AR4.1 v1 rows: %+v", candidate)
	}
	if _, err := LoadCatalog(readFile(t, fixturePath), declarations()); err != nil {
		t.Fatalf("corpus fixture must load: %v", err)
	}
}

func TestArcadeCatalogRejectsEveryLoaderDefect(t *testing.T) {
	stage := func(v map[string]any) map[string]any {
		return v["container"].(map[string]any)["stages"].([]any)[0].(map[string]any)
	}
	preset := func(v map[string]any, index int) map[string]any {
		return v["mine_grid"].(map[string]any)["presets"].([]any)[index].(map[string]any)
	}
	snake := func(v map[string]any) map[string]any { return v["snake"].(map[string]any) }
	cases := map[string]func(map[string]any){
		"unknown root key":      func(v map[string]any) { v["extra"] = 1 },
		"wrong schema":          func(v map[string]any) { v["schema_version"] = 2 },
		"unknown container key": func(v map[string]any) { v["container"].(map[string]any)["extra"] = 1 },
		"unknown stage key":     func(v map[string]any) { stage(v)["extra"] = 1 },
		"empty stages":          func(v map[string]any) { v["container"].(map[string]any)["stages"] = []any{} },
		"unsorted toys":         func(v map[string]any) { stage(v)["toys"] = []any{"arcade.snake", "arcade.mine_grid"} },
		"unknown title key":     func(v map[string]any) { stage(v)["title_copy_key"] = "arcade.stage.missing.title" },
		"stage tier above max":  func(v map[string]any) { stage(v)["min_tier"] = 10 },
		"unknown preset key":    func(v map[string]any) { preset(v, 0)["extra"] = 1 },
		"unsorted presets": func(v map[string]any) {
			p := v["mine_grid"].(map[string]any)["presets"].([]any)
			p[0], p[1] = p[1], p[0]
		},
		"narrow board":            func(v map[string]any) { preset(v, 2)["width"] = 4 },
		"tall board":              func(v map[string]any) { preset(v, 2)["height"] = 31 },
		"too many mines":          func(v map[string]any) { preset(v, 2)["mines"] = 9*9 - 8 },
		"no mines":                func(v map[string]any) { preset(v, 2)["mines"] = 0 },
		"preset copy key drift":   func(v map[string]any) { preset(v, 2)["copy_key"] = "arcade.mine_grid.preset.large" },
		"unknown snake key":       func(v map[string]any) { snake(v)["extra"] = 1 },
		"snake start too long":    func(v map[string]any) { snake(v)["start_length"] = 11 },
		"snake growth zero":       func(v map[string]any) { snake(v)["growth_per_food"] = 0 },
		"snake advance too large": func(v map[string]any) { snake(v)["max_ticks_per_advance"] = 257 },
		"snake tick too fast":     func(v map[string]any) { snake(v)["presentation_tick_ms"] = 49 },
	}
	for name, change := range cases {
		if _, err := LoadCatalog(mutate(t, candidatePath, change), declarations()); !errors.Is(err, ErrInvalidCatalog) {
			t.Fatalf("%s: expected rejection, got %v", name, err)
		}
	}
	// The boundaries themselves are legal.
	boundary := mutate(t, candidatePath, func(v map[string]any) {
		preset(v, 2)["mines"] = 9*9 - 9
		snake(v)["start_length"] = 10
		snake(v)["presentation_tick_ms"] = 50
	})
	if _, err := LoadCatalog(boundary, declarations()); err != nil {
		t.Fatalf("inclusive loader bounds must load: %v", err)
	}
}

package replaycatalog

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"cloud-clicker/server/epochseed"
	"cloud-clicker/server/save"
)

func TestLoadAndResolveOptionalHistoricalFormulas(t *testing.T) {
	seed, err := epochseed.Load(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	legacy, err := Load(seed.Hash, seed.Artifacts)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := legacy.ResolvePrestige(seed.Hash); !ok {
		t.Fatal("existing bundle no longer resolves")
	}
	data, err := os.ReadFile(filepath.Join("..", "..", "docs", "generated", "production-formulas.json"))
	if err != nil {
		t.Fatal(err)
	}
	seed.Artifacts["formulas"] = data
	if _, err := Load(seed.Hash, seed.Artifacts); err == nil {
		t.Fatal("formula bytes accepted under pre-formula hash")
	}
	hash, err := save.ConstantsHashArtifacts(seed.Artifacts)
	if err != nil {
		t.Fatal(err)
	}
	bundle, err := Load(hash, seed.Artifacts)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(bundle.Artifacts["formulas"], data) {
		t.Fatal("formula bytes were rewritten")
	}
	if _, ok := bundle.ResolvePrestige(hash); !ok {
		t.Fatal("formula-bearing bundle loads but cannot resolve")
	}
	data[0] = ' '
	if bundle.Artifacts["formulas"][0] != '{' {
		t.Fatal("loader aliases caller formula bytes")
	}

	fixture, err := os.ReadFile(filepath.Join("..", "..", "balance", "testdata", "formula-artifact-parity-v14.json"))
	if err != nil {
		t.Fatal(err)
	}
	var envelope struct {
		Cases []struct {
			ID      string `json:"id"`
			Replace string `json:"replace"`
			With    string `json:"with"`
			Valid   bool   `json:"valid"`
		} `json:"cases"`
	}
	if err := json.Unmarshal(fixture, &envelope); err != nil {
		t.Fatal(err)
	}
	baseline := bytes.Clone(bundle.Artifacts["formulas"])
	for _, control := range envelope.Cases {
		t.Run(control.ID, func(t *testing.T) {
			if bytes.Count(baseline, []byte(control.Replace)) != 1 {
				t.Fatal("control failed to hit exactly one field")
			}
			bundle.Artifacts["formulas"] = bytes.Replace(baseline, []byte(control.Replace), []byte(control.With), 1)
			identity, err := save.ConstantsHashArtifacts(bundle.Artifacts)
			if err != nil {
				t.Fatal(err)
			}
			loaded, err := Load(identity, bundle.Artifacts)
			if (err == nil) != control.Valid {
				t.Fatalf("loader valid=%v err=%v", control.Valid, err)
			}
			if err == nil {
				if _, ok := loaded.ResolvePrestige(identity); !ok {
					t.Fatal("valid control cannot resolve")
				}
			}
			bundle.ConstantsHash = identity
			if _, ok := bundle.ResolvePrestige(identity); ok != control.Valid {
				t.Fatalf("resolver valid=%v want=%v", ok, control.Valid)
			}
		})
	}
}

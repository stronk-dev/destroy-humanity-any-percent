package fiscal

import (
	"encoding/json"
	"os"
	"testing"

	"cloud-clicker/server/economy"
)

type sharedFixture struct {
	Version          int             `json:"version"`
	Baseline         json.RawMessage `json:"baseline"`
	InvalidMutations []struct {
		Name, Op string
		Path     []any
		Value    any
	} `json:"invalid_mutations"`
	FactorVectors []struct {
		Kind     string
		Count    int64
		Expected string
	} `json:"factor_vectors"`
	CostVectors []struct{ Current, Levels, Expected int64 } `json:"cost_vectors"`
	RNGVectors  []struct {
		FounderID string `json:"founder_id"`
		Sequence  int64  `json:"sequence"`
		Expected  int64  `json:"expected"`
	} `json:"rng_vectors"`
}

func loadSharedFixture(t *testing.T) sharedFixture {
	t.Helper()
	data, err := os.ReadFile("../../balance/testdata/fiscal-foundation-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture sharedFixture
	if json.Unmarshal(data, &fixture) != nil || fixture.Version != 1 {
		t.Fatal("invalid fiscal fixture")
	}
	return fixture
}

func fiscalEconomy(t *testing.T) *economy.Catalog {
	t.Helper()
	data, err := os.ReadFile("../../balance/catalogs/phase0.json")
	if err != nil {
		t.Fatal(err)
	}
	var root map[string]any
	if json.Unmarshal(data, &root) != nil {
		t.Fatal("invalid economy fixture")
	}
	values := root["multiplier_sources"].([]any)
	values = append(values,
		map[string]any{"id": "fiscal.generator.beige_tower", "slot": "prestige", "target": "generator.beige_tower", "provider": "fiscal"},
		map[string]any{"id": "fiscal.hoard", "slot": "prestige", "target": "all", "provider": "fiscal"},
	)
	root["multiplier_sources"] = values
	candidate, _ := json.Marshal(root)
	catalog, err := economy.LoadCatalog(candidate)
	if err != nil {
		t.Fatal(err)
	}
	return catalog
}

func TestCatalogMatchesSharedMutationCorpus(t *testing.T) {
	fixture, economyCatalog := loadSharedFixture(t), fiscalEconomy(t)
	if _, err := LoadCatalog(fixture.Baseline, economyCatalog); err != nil {
		t.Fatal(err)
	}
	for _, vector := range fixture.InvalidMutations {
		t.Run(vector.Name, func(t *testing.T) {
			var root any
			if json.Unmarshal(fixture.Baseline, &root) != nil {
				t.Fatal("invalid baseline")
			}
			applyMutation(t, root, vector.Op, vector.Path, vector.Value)
			candidate, _ := json.Marshal(root)
			if _, err := LoadCatalog(candidate, economyCatalog); err == nil {
				t.Fatal("invalid mutation accepted")
			}
		})
	}
}

func TestFactorsCostsAndRNGMatchSharedVectors(t *testing.T) {
	fixture := loadSharedFixture(t)
	catalog, err := LoadCatalog(fixture.Baseline, fiscalEconomy(t))
	if err != nil {
		t.Fatal(err)
	}
	for _, vector := range fixture.FactorVectors {
		var value string
		if vector.Kind == "hoard" {
			factor, err := catalog.HoardFactor(vector.Count)
			if err != nil {
				t.Fatal(err)
			}
			value = factor.String()
		} else {
			factor, err := catalog.GeneratorLevelFactor("generator.beige_tower", vector.Count)
			if err != nil {
				t.Fatal(err)
			}
			value = factor.String()
		}
		if value != vector.Expected {
			t.Errorf("%s factor(%d)=%s want %s", vector.Kind, vector.Count, value, vector.Expected)
		}
	}
	for _, vector := range fixture.CostVectors {
		value, err := catalog.GeneratorLevelCost("generator.beige_tower", vector.Current, vector.Levels)
		if err != nil || value != vector.Expected {
			t.Errorf("cost(%d,%d)=%d,%v want %d", vector.Current, vector.Levels, value, err, vector.Expected)
		}
	}
	for _, vector := range fixture.RNGVectors {
		value, err := EarlyHarvestDraw(vector.FounderID, vector.Sequence)
		if err != nil || value != vector.Expected {
			t.Errorf("draw(%s,%d)=%d,%v want %d", vector.FounderID, vector.Sequence, value, err, vector.Expected)
		}
	}
}

func applyMutation(t *testing.T, root any, operation string, path []any, value any) {
	t.Helper()
	current := root
	for _, component := range path[:len(path)-1] {
		switch typed := component.(type) {
		case string:
			current = current.(map[string]any)[typed]
		case float64:
			current = current.([]any)[int(typed)]
		}
	}
	last := path[len(path)-1]
	switch typed := last.(type) {
	case string:
		object := current.(map[string]any)
		if operation == "delete" {
			delete(object, typed)
		} else {
			object[typed] = value
		}
	case float64:
		current.([]any)[int(typed)] = value
	}
}

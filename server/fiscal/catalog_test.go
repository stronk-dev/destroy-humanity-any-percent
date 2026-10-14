package fiscal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"

	"cloud-clicker/server/economy"
)

func editFiscalFixture(t *testing.T, edit func(map[string]any)) []byte {
	t.Helper()
	var root map[string]any
	if err := json.Unmarshal(loadSharedFixture(t).Baseline, &root); err != nil {
		t.Fatal(err)
	}
	edit(root)
	data, err := json.Marshal(root)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func fiscalShapeObjects(root map[string]any) []map[string]any {
	objects := []map[string]any{root, root["clock_policy"].(map[string]any), root["credit_policy"].(map[string]any), root["hoard_policy"].(map[string]any)}
	for _, family := range []string{"generator_level_rows", "unlock_rows"} {
		for _, row := range root[family].([]any) {
			objects = append(objects, row.(map[string]any))
		}
	}
	return objects
}

func fiscalPolicyOverride(t *testing.T, key string, override any) []byte {
	t.Helper()
	source := bytes.TrimSpace(loadSharedFixture(t).Baseline)
	data, err := json.Marshal(override)
	if err != nil {
		t.Fatal(err)
	}
	result := append(bytes.Clone(source[:len(source)-1]), []byte(`,"`+key+`":`)...)
	result = append(result, data...)
	result = append(result, '}')
	if !json.Valid(result) {
		t.Fatal("override is not valid JSON")
	}
	return result
}

func TestCatalogRejectsCaseAliasedObjectMembers(t *testing.T) {
	var root map[string]any
	if err := json.Unmarshal(loadSharedFixture(t).Baseline, &root); err != nil {
		t.Fatal(err)
	}
	for index, object := range fiscalShapeObjects(root) {
		for key := range object {
			t.Run(fmt.Sprintf("%d/%s", index, key), func(t *testing.T) {
				data := editFiscalFixture(t, func(candidate map[string]any) {
					row := fiscalShapeObjects(candidate)[index]
					row[strings.ToUpper(key)] = row[key]
					delete(row, key)
				})
				if _, err := LoadCatalog(data, fiscalEconomy(t)); err == nil {
					t.Fatal("case alias admitted")
				}
			})
		}
	}
}

func TestCatalogRejectsIncompleteOrNullFinalPolicies(t *testing.T) {
	var root map[string]any
	if err := json.Unmarshal(loadSharedFixture(t).Baseline, &root); err != nil {
		t.Fatal(err)
	}
	for family, field := range map[string]string{"clock_policy": "early_ms", "credit_policy": "credit_per_period", "hoard_policy": "ppm_per_credit"} {
		for _, nullPolicy := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/null=%t", family, nullPolicy), func(t *testing.T) {
				var override any = map[string]any{field: root[family].(map[string]any)[field]}
				if nullPolicy {
					override = nil
				}
				if _, err := LoadCatalog(fiscalPolicyOverride(t, family, override), fiscalEconomy(t)); err == nil {
					t.Fatal("final partial/null policy inherited earlier fields")
				}
			})
		}
	}
	credit := root["credit_policy"].(map[string]any)
	credit["hardcap_reason_key"] = nil
	if _, err := LoadCatalog(fiscalPolicyOverride(t, "credit_policy", credit), fiscalEconomy(t)); err == nil {
		t.Error("final null reason inherited earlier string")
	}
}

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
	catalog, err := economy.LoadCatalog(data)
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

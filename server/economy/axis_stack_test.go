package economy_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"cloud-clicker/server/economy"
	"cloud-clicker/server/routes"
)

type axisCorpus struct {
	EconomyFixture string          `json:"economy_fixture"`
	RoutesBase     string          `json:"routes_base"`
	EconomyCases   []axisCorpusRow `json:"economy_cases"`
	CrossCases     []axisCorpusRow `json:"cross_cases"`
	RouteCases     []axisCorpusRow `json:"route_cases"`
}

type axisCorpusRow struct {
	Name                 string              `json:"name"`
	Mutations            []axisCorpusMutator `json:"mutations"`
	Expect               string              `json:"expect"`
	AchievementsPinned   bool                `json:"achievements_pinned"`
	MaximumAttainmentRun int64               `json:"maximum_attainment_run"`
	MaximumScoreRun      int64               `json:"maximum_score_run"`
}

type axisCorpusMutator struct {
	Op    string            `json:"op"`
	Path  []json.RawMessage `json:"path"`
	Value json.RawMessage   `json:"value"`
}

func repositoryFile(t *testing.T, relative string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "..", relative))
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func loadAxisCorpus(t *testing.T) axisCorpus {
	t.Helper()
	var corpus axisCorpus
	if err := json.Unmarshal(repositoryFile(t, "testdata/axis-stack/loader-corpus-v1.json"), &corpus); err != nil {
		t.Fatal(err)
	}
	return corpus
}

// mutateJSON applies the corpus mutations. A path element is a key, an index,
// or an {"id": X} selector naming the array element whose "id" is X.
func mutateJSON(t *testing.T, base []byte, mutations []axisCorpusMutator) []byte {
	t.Helper()
	decoder := json.NewDecoder(bytes.NewReader(base))
	decoder.UseNumber()
	var document any
	if err := decoder.Decode(&document); err != nil {
		t.Fatal(err)
	}
	for _, mutation := range mutations {
		var value any
		if mutation.Value != nil {
			valueDecoder := json.NewDecoder(bytes.NewReader(mutation.Value))
			valueDecoder.UseNumber()
			if err := valueDecoder.Decode(&value); err != nil {
				t.Fatal(err)
			}
		}
		document = applyMutation(t, document, mutation.Path, mutation.Op, value)
	}
	encoded, err := json.Marshal(document)
	if err != nil {
		t.Fatal(err)
	}
	return encoded
}

func applyMutation(t *testing.T, node any, path []json.RawMessage, op string, value any) any {
	t.Helper()
	if len(path) == 0 {
		switch op {
		case "set":
			return value
		case "append":
			list, ok := node.([]any)
			if !ok {
				t.Fatalf("append target is not an array")
			}
			return append(list, value)
		}
		t.Fatalf("unsupported terminal op %q", op)
	}
	switch container := node.(type) {
	case map[string]any:
		var key string
		if err := json.Unmarshal(path[0], &key); err != nil {
			t.Fatalf("object path element %s: %v", path[0], err)
		}
		if len(path) == 1 && op == "delete" {
			delete(container, key)
			return container
		}
		container[key] = applyMutation(t, container[key], path[1:], op, value)
		return container
	case []any:
		index := -1
		var numeric int
		if err := json.Unmarshal(path[0], &numeric); err == nil {
			index = numeric
		} else {
			var selector map[string]string
			if err := json.Unmarshal(path[0], &selector); err != nil {
				t.Fatalf("array path element %s: %v", path[0], err)
			}
			for candidate, element := range container {
				if row, ok := element.(map[string]any); ok && row["id"] == selector["id"] {
					index = candidate
				}
			}
		}
		if index < 0 || index >= len(container) {
			t.Fatalf("path element %s selects nothing", path[0])
		}
		if len(path) == 1 && op == "delete" {
			return append(append([]any{}, container[:index]...), container[index+1:]...)
		}
		container[index] = applyMutation(t, container[index], path[1:], op, value)
		return container
	}
	t.Fatalf("path %s descends into a scalar", path[0])
	return nil
}

func TestAxisStackLoaderCorpus(t *testing.T) {
	corpus := loadAxisCorpus(t)
	fixture := repositoryFile(t, corpus.EconomyFixture)
	if len(corpus.EconomyCases) < 20 {
		t.Fatalf("economy corpus shrank to %d cases", len(corpus.EconomyCases))
	}
	for _, row := range corpus.EconomyCases {
		t.Run(row.Name, func(t *testing.T) {
			_, err := economy.LoadCatalog(mutateJSON(t, fixture, row.Mutations))
			if (row.Expect == "accept") != (err == nil) {
				t.Fatalf("expect %s, got err=%v", row.Expect, err)
			}
		})
	}
	for _, row := range corpus.CrossCases {
		t.Run("cross/"+row.Name, func(t *testing.T) {
			catalog, err := economy.LoadCatalog(mutateJSON(t, fixture, row.Mutations))
			if err != nil {
				t.Fatalf("cross rows must load the economy: %v", err)
			}
			err = catalog.ValidateAxisInputs(row.AchievementsPinned, row.MaximumAttainmentRun, row.MaximumScoreRun)
			if (row.Expect == "accept") != (err == nil) {
				t.Fatalf("expect %s, got err=%v", row.Expect, err)
			}
		})
	}
	routesBase := repositoryFile(t, corpus.RoutesBase)
	for _, row := range corpus.RouteCases {
		t.Run("routes/"+row.Name, func(t *testing.T) {
			_, err := routes.LoadCatalog(mutateJSON(t, routesBase, row.Mutations))
			if (row.Expect == "accept") != (err == nil) {
				t.Fatalf("expect %s, got err=%v", row.Expect, err)
			}
		})
	}
}

func TestAxisStackFixtureShape(t *testing.T) {
	catalog, err := economy.LoadCatalog(repositoryFile(t, "balance/testdata/axis-stack/economy-v5-fixture.json"))
	if err != nil {
		t.Fatal(err)
	}
	axis, ok := catalog.AxisStack()
	if !ok || axis.Input != economy.AxisInputAttainmentRun || axis.InputCap != 44 || axis.CapReasonKey != "cap.axis_stack_input" {
		t.Fatalf("axis stack = %+v, %v", axis, ok)
	}
	for _, expected := range []struct {
		id      string
		minimum int64
		ppm     int64
	}{{"upgrade.pr_intern_1", 6, 25_000}, {"upgrade.pr_intern_2", 10, 20_000}} {
		upgrade, ok := catalog.Upgrade(expected.id)
		if !ok || upgrade.AxisMinimum != expected.minimum || len(upgrade.Requires) != 0 || len(upgrade.Effects) != 1 ||
			upgrade.Effects[0].Slot != economy.SlotAxisStack || upgrade.Effects[0].FactorPPM != expected.ppm || upgrade.Effects[0].Target != "all" {
			t.Fatalf("%s = %+v", expected.id, upgrade)
		}
		source, ok := catalog.MultiplierSource(expected.id + ".axis")
		if !ok || source.Provider != economy.AxisStackProvider {
			t.Fatalf("%s declaration = %+v", expected.id, source)
		}
	}
	// The epoch-8 schema-4 catalog still loads unchanged.
	if _, err := economy.LoadCatalog(repositoryFile(t, "balance/catalogs/phase0.json")); err != nil {
		t.Fatal(err)
	}
	if fmt.Sprint(economy.MultiplierSlotOrder) != "[upgrades milestones axis_stack faction doctrine commons trust event_buffs prestige]" {
		t.Fatalf("slot order = %v (OD-6: axis_stack after milestones, before faction)", economy.MultiplierSlotOrder)
	}
}

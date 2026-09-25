package garden

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

	"cloud-clicker/server/copykeys"
	"cloud-clicker/server/fiscal"
	"cloud-clicker/server/minigame"
)

const repositoryRoot = "../../"

type fixtureOp struct {
	Op    string          `json:"op"`
	Path  []any           `json:"path"`
	Value json.RawMessage `json:"value"`
}

type fixtureCase struct {
	Name         string      `json:"name"`
	Ops          []fixtureOp `json:"ops"`
	DuplicateKey string      `json:"duplicate_key"`
}

type fixtureCorpus struct {
	Version     int           `json:"version"`
	Valid       string        `json:"valid"`
	Fiscal      string        `json:"fiscal"`
	ResourceIDs []string      `json:"resource_ids"`
	Cases       []fixtureCase `json:"cases"`
}

func readRepo(t testing.TB, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(repositoryRoot + path)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func loadCorpus(t testing.TB) fixtureCorpus {
	t.Helper()
	var corpus fixtureCorpus
	if err := json.Unmarshal(readRepo(t, "testdata/garden/catalog-fixtures-v1.json"), &corpus); err != nil || corpus.Version != 1 {
		t.Fatalf("garden fixture corpus: %v", err)
	}
	return corpus
}

// fixtureDeclarations resolves the declared authorities from the fixture Fiscal
// artifact's raw rows, the corpus resource list, and the generated copy keys.
func fixtureDeclarations(t testing.TB, corpus fixtureCorpus) Declarations {
	t.Helper()
	var raw struct {
		GeneratorLevelRows []struct {
			GeneratorID string `json:"generator_id"`
		} `json:"generator_level_rows"`
		UnlockRows []struct {
			UnlockID string `json:"unlock_id"`
		} `json:"unlock_rows"`
	}
	if err := json.Unmarshal(readRepo(t, corpus.Fiscal), &raw); err != nil {
		t.Fatal(err)
	}
	declarations := Declarations{CopyKeys: map[string]struct{}{}, ResourceIDs: map[string]struct{}{}, FiscalUnlockIDs: map[string]struct{}{}, FiscalGeneratorIDs: map[string]struct{}{}}
	for _, key := range copykeys.All() {
		declarations.CopyKeys[key] = struct{}{}
	}
	for _, id := range corpus.ResourceIDs {
		declarations.ResourceIDs[id] = struct{}{}
	}
	for _, row := range raw.GeneratorLevelRows {
		declarations.FiscalGeneratorIDs[row.GeneratorID] = struct{}{}
	}
	for _, row := range raw.UnlockRows {
		declarations.FiscalUnlockIDs[row.UnlockID] = struct{}{}
	}
	declarations.ValidatePayout = minigame.GardenPayoutValidator(declarations.ResourceIDs, declarations.CopyKeys)
	return declarations
}

func fixtureCatalog(t testing.TB) *Catalog {
	t.Helper()
	corpus := loadCorpus(t)
	catalog, err := LoadCatalog(readRepo(t, corpus.Valid), fixtureDeclarations(t, corpus))
	if err != nil {
		t.Fatalf("valid garden fixture must load: %v", err)
	}
	return catalog
}

// applyCase mutates the valid fixture through the case's ops (or builds the
// raw duplicate-key text) so Go and TS test identical bytes.
func applyCase(t testing.TB, valid []byte, test fixtureCase) []byte {
	t.Helper()
	if test.DuplicateKey != "" {
		index := bytes.IndexByte(valid, '{')
		return append(append(append([]byte{}, valid[:index+1]...), []byte(fmt.Sprintf("%q: %q,", test.DuplicateKey, "unrelated"))...), valid[index+1:]...)
	}
	var document any
	decoder := json.NewDecoder(bytes.NewReader(valid))
	decoder.UseNumber()
	if err := decoder.Decode(&document); err != nil {
		t.Fatal(err)
	}
	for _, op := range test.Ops {
		document = mutate(t, document, op.Path, op)
	}
	data, err := json.Marshal(document)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func mutate(t testing.TB, node any, path []any, op fixtureOp) any {
	t.Helper()
	if len(path) == 0 {
		var value any
		decoder := json.NewDecoder(bytes.NewReader(op.Value))
		decoder.UseNumber()
		if err := decoder.Decode(&value); err != nil {
			t.Fatal(err)
		}
		return value
	}
	switch typed := node.(type) {
	case map[string]any:
		key := path[0].(string)
		if len(path) == 1 && op.Op == "delete" {
			delete(typed, key)
			return typed
		}
		typed[key] = mutate(t, typed[key], path[1:], op)
		return typed
	case []any:
		index := int(path[0].(float64))
		typed[index] = mutate(t, typed[index], path[1:], op)
		return typed
	}
	t.Fatalf("fixture path %v does not resolve", path)
	return nil
}

func TestGardenFixtureLoads(t *testing.T) {
	catalog := fixtureCatalog(t)
	if catalog.UnlockID != "minigame.server_garden" || catalog.HostGeneratorID != "generator.beige_tower" || len(catalog.Species) != 5 ||
		len(catalog.Substrates) != 4 || len(catalog.Recipes) != 5 || fmt.Sprint(catalog.Starters()) != "[strain_a strain_b]" {
		t.Fatalf("fixture shape %+v", catalog)
	}
	if got := catalog.Dimension(0); got.Width != 2 || got.Height != 2 {
		t.Fatalf("level 0 grid %+v", got)
	}
	if got := catalog.Dimension(100); got.Width != 6 || got.Height != 6 {
		t.Fatalf("hardcap grid %+v", got)
	}
	if got := catalog.Dimension(3); got.Width != 4 || got.Height != 3 {
		t.Fatalf("level 3 grid %+v", got)
	}
}

// TestGardenLoaderRejectsEveryRule is AC1: every SG1 rule rejects its own
// fixture. The production Fiscal artifact alone must also reject the fixture
// garden (no minigame.server_garden unlock row), which is rule 10.
func TestGardenLoaderRejectsEveryRule(t *testing.T) {
	corpus := loadCorpus(t)
	valid := readRepo(t, corpus.Valid)
	declarations := fixtureDeclarations(t, corpus)
	if len(corpus.Cases) < 40 {
		t.Fatalf("rejecting corpus shrank to %d cases", len(corpus.Cases))
	}
	for _, test := range corpus.Cases {
		t.Run(test.Name, func(t *testing.T) {
			data := applyCase(t, valid, test)
			if _, err := LoadCatalog(data, declarations); !errors.Is(err, ErrInvalidCatalog) {
				t.Fatalf("case %s loaded (err=%v):\n%s", test.Name, err, data)
			}
		})
	}
	production := fixtureDeclarations(t, fixtureCorpus{Fiscal: "balance/fiscal/first-content.json", ResourceIDs: corpus.ResourceIDs})
	if _, err := LoadCatalog(valid, production); !errors.Is(err, ErrInvalidCatalog) {
		t.Fatalf("the production Fiscal artifact has no garden unlock row, so the fixture must reject: %v", err)
	}
	if _, err := fiscal.LoadCatalog(readRepo(t, corpus.Fiscal), nil); err == nil {
		t.Fatal("fiscal loader must require an economy catalog")
	}
}

// TestGardenBudgetRuleRejects exercises SG1 rule 8 directly. Within rule 3's
// domains the worst case is ceil(604_800_000/60_000) × 36 = 362_880, so no
// artifact can trip it; this proves the check itself can fail.
func TestGardenBudgetRuleRejects(t *testing.T) {
	catalog := fixtureCatalog(t)
	if err := catalog.validateBudget(); err != nil {
		t.Fatalf("fixture budget: %v", err)
	}
	catalog.CatchupCapMS = EvaluationBudget * catalog.Substrates[0].TickMS
	if err := catalog.validateBudget(); !errors.Is(err, ErrInvalidCatalog) {
		t.Fatalf("an over-budget clock must reject: %v", err)
	}
}

func TestGardenTransitionIsAppendOnly(t *testing.T) {
	corpus := loadCorpus(t)
	declarations := fixtureDeclarations(t, corpus)
	current := fixtureCatalog(t)
	if err := ValidateTransition(current, current); err != nil {
		t.Fatalf("identity transition: %v", err)
	}
	if err := ValidateTransition(nil, current); err != nil {
		t.Fatalf("activation transition: %v", err)
	}
	if err := ValidateTransition(current, nil); !errors.Is(err, ErrInvalidTransition) {
		t.Fatalf("disappearing artifact must reject: %v", err)
	}
	valid := string(readRepo(t, corpus.Valid))
	for name, replacement := range map[string][2]string{
		"species":   {`"species_id": "strain_e"`, `"species_id": "strain_f"`},
		"substrate": {`"substrate_id": "mainframe"`, `"substrate_id": "mainframes"`},
		"recipe":    {`"recipe_id": "r_e_cluster"`, `"recipe_id": "r_e_cluster2"`},
	} {
		mutated := strings.Replace(valid, replacement[0], replacement[1], 1)
		if name == "species" {
			mutated = strings.ReplaceAll(mutated, `"child_species_id": "strain_e"`, `"child_species_id": "strain_f"`)
			mutated = strings.ReplaceAll(mutated, "garden.species.strain_f.", "garden.species.strain_e.")
		}
		next, err := LoadCatalog([]byte(mutated), declarations)
		if err != nil {
			t.Fatalf("%s rename fixture must load: %v", name, err)
		}
		if err := ValidateTransition(current, next); !errors.Is(err, ErrInvalidTransition) {
			t.Fatalf("removing a %s id must reject: %v", name, err)
		}
	}
}

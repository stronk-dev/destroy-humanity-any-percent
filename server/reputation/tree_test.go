package reputation

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"cloud-clicker/server/copykeys"
	"cloud-clicker/server/curriculum"
	"cloud-clicker/server/decimal"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/routes"
)

const repositoryRoot = "../.."

func readRepository(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(repositoryRoot, path))
	if err != nil {
		t.Fatal(err)
	}
	return data
}

// FixtureEconomyBytes is the live economy artifact plus exactly the one
// declaration row R2 adds; the production artifact gains it only at mint.
func fixtureEconomyBytes(t *testing.T) []byte {
	t.Helper()
	var root map[string]any
	if err := json.Unmarshal(readRepository(t, "balance/catalogs/phase0.json"), &root); err != nil {
		t.Fatal(err)
	}
	sources := root["multiplier_sources"].([]any)
	root["multiplier_sources"] = append(sources, map[string]any{"id": "reputation.founder_bonus", "slot": "prestige", "target": "all", "provider": Provider})
	data, err := json.Marshal(root)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func fixtureDeclarations(t *testing.T) Declarations {
	t.Helper()
	economyCatalog, err := economy.LoadCatalog(fixtureEconomyBytes(t))
	if err != nil {
		t.Fatal(err)
	}
	keys := map[string]struct{}{}
	for _, key := range copykeys.All() {
		keys[key] = struct{}{}
	}
	routeCatalog, err := routes.LoadCatalog(readRepository(t, "balance/routes/phase0.json"))
	if err != nil {
		t.Fatal(err)
	}
	gates := map[string]struct{}{}
	for _, gate := range routeCatalog.Gates() {
		gates[gate.ID] = struct{}{}
	}
	curriculumCatalog, err := curriculum.Load(readRepository(t, "balance/curriculum/t0-t1.json"), curriculum.Declarations{Economy: economyCatalog, CopyKeys: keys, GateIDs: gates})
	if err != nil {
		t.Fatal(err)
	}
	return Declarations{Economy: economyCatalog, Curriculum: curriculumCatalog, CopyKeys: keys}
}

type treeCorpus struct {
	SchemaVersion int    `json:"schema_version"`
	Valid         string `json:"valid"`
	Rejections    []struct {
		Name string          `json:"name"`
		Rule int             `json:"rule"`
		Tree json.RawMessage `json:"tree"`
	} `json:"rejections"`
}

func TestTreeCorpusLoadsValidFixtureAndRejectsEveryRule(t *testing.T) {
	var corpus treeCorpus
	if err := json.Unmarshal(readRepository(t, "testdata/reputation/tree-fixtures-v1.json"), &corpus); err != nil || corpus.SchemaVersion != 1 {
		t.Fatalf("corpus: %v", err)
	}
	declarations := fixtureDeclarations(t)
	tree, err := LoadTree(readRepository(t, corpus.Valid), declarations)
	if err != nil {
		t.Fatalf("valid fixture rejected: %v", err)
	}
	if len(tree.Nodes()) != 9 || tree.Bonus.PerLevelPPM != 10_000 {
		t.Fatalf("unexpected fixture shape: %+v", tree.Bonus)
	}
	rules := map[int]bool{}
	for _, row := range corpus.Rejections {
		if _, err := LoadTree(row.Tree, declarations); !errors.Is(err, ErrInvalidTree) {
			t.Errorf("%s (rule %d) loaded: %v", row.Name, row.Rule, err)
		} else if row.Rule > 0 && !strings.Contains(err.Error(), "rule ") {
			t.Errorf("%s: rejection does not name a rule: %v", row.Name, err)
		}
		rules[row.Rule] = true
	}
	for rule := 1; rule <= 8; rule++ {
		if !rules[rule] {
			t.Errorf("corpus has no rejection fixture for rule %d", rule)
		}
	}
}

func TestTreeRejectsBundleWithoutTheDeclarationRow(t *testing.T) {
	declarations := fixtureDeclarations(t)
	live, err := economy.LoadCatalog(readRepository(t, "balance/catalogs/phase0.json"))
	if err != nil {
		t.Fatal(err)
	}
	declarations.Economy = live
	if _, err := LoadTree(readRepository(t, "balance/testdata/reputation-tree/fixture-v1.json"), declarations); !errors.Is(err, ErrInvalidTree) {
		t.Fatalf("tree loaded against an economy lacking reputation.founder_bonus: %v", err)
	}
}

func TestAccountingDerivations(t *testing.T) {
	tree, err := LoadTree(readRepository(t, "balance/testdata/reputation-tree/fixture-v1.json"), fixtureDeclarations(t))
	if err != nil {
		t.Fatal(err)
	}
	if value, err := Available(10, 4); err != nil || value != 6 {
		t.Fatalf("available = %d, %v", value, err)
	}
	for _, invalid := range [][2]int64{{3, 4}, {-1, 0}, {5, -1}, {decimal.MaxExactInteger + 1, 0}} {
		if _, err := Available(invalid[0], invalid[1]); !errors.Is(err, ErrInvalidState) {
			t.Fatalf("available(%v) accepted", invalid)
		}
	}
	owned := []string{"reputation.starter.cash_small", "reputation.unlock.p05", "reputation.unlock.p25", "reputation.unlock.retired"}
	if value, err := tree.UnlockPPM(owned); err != nil || value != 250_000 {
		t.Fatalf("unlock = %d, %v", value, err)
	}
	if value, err := tree.UnlockPPM(nil); err != nil || value != 0 {
		t.Fatalf("empty unlock = %d, %v", value, err)
	}
	if _, err := tree.UnlockPPM([]string{"reputation.unlock.p25", "reputation.unlock.p05"}); !errors.Is(err, ErrInvalidState) {
		t.Fatal("unsorted owned set accepted")
	}
}

type bonusVector struct {
	Level       int64  `json:"level"`
	Spent       int64  `json:"spent"`
	PerLevelPPM int64  `json:"per_level_ppm"`
	UnlockPPM   int64  `json:"unlock_ppm"`
	Factor      string `json:"factor"`
}

var bonusInputs = []bonusVector{
	{Level: 0, Spent: 0, PerLevelPPM: 10_000, UnlockPPM: 0},
	{Level: 0, Spent: 0, PerLevelPPM: 10_000, UnlockPPM: 1_000_000},
	{Level: 7, Spent: 3, PerLevelPPM: 10_000, UnlockPPM: 0},
	{Level: 1, Spent: 1, PerLevelPPM: 10_000, UnlockPPM: 50_000},
	{Level: 4, Spent: 1, PerLevelPPM: 10_000, UnlockPPM: 50_000},
	{Level: 552, Spent: 552, PerLevelPPM: 10_000, UnlockPPM: 1_000_000},
	{Level: 1_000, Spent: 17, PerLevelPPM: 10_000, UnlockPPM: 750_000},
	{Level: 3, Spent: 2, PerLevelPPM: 1, UnlockPPM: 1},
	{Level: decimal.MaxExactInteger, Spent: 5, PerLevelPPM: 1_000_000, UnlockPPM: 1_000_000},
	{Level: decimal.MaxExactInteger, Spent: 0, PerLevelPPM: 10_000, UnlockPPM: 250_000},
}

// TestBonusVectors pins the Go-authored AC5 vectors both runtimes consume.
// REPUTATION_UPDATE_VECTORS=1 regenerates the file from the Go arithmetic.
func TestBonusVectors(t *testing.T) {
	path := filepath.Join(repositoryRoot, "testdata/reputation/bonus-vectors-v1.json")
	computed := make([]bonusVector, len(bonusInputs))
	for index, input := range bonusInputs {
		factor, err := BonusFactor(input.Level, input.Spent, input.PerLevelPPM, input.UnlockPPM)
		if err != nil {
			t.Fatalf("vector %d: %v", index, err)
		}
		computed[index] = input
		computed[index].Factor = factor.String()
	}
	if os.Getenv("REPUTATION_UPDATE_VECTORS") == "1" {
		data, _ := json.MarshalIndent(map[string]any{"schema_version": 1, "vectors": computed}, "", "  ")
		if err := os.WriteFile(path, append(data, '\n'), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	var pinned struct {
		SchemaVersion int           `json:"schema_version"`
		Vectors       []bonusVector `json:"vectors"`
	}
	data, err := os.ReadFile(path)
	if err != nil || json.Unmarshal(data, &pinned) != nil || pinned.SchemaVersion != 1 || len(pinned.Vectors) != len(computed) {
		t.Fatalf("pinned vectors unreadable: %v", err)
	}
	for index := range computed {
		if pinned.Vectors[index] != computed[index] {
			t.Errorf("vector %d: pinned %+v computed %+v", index, pinned.Vectors[index], computed[index])
		}
	}
	if pinned.Vectors[3].Factor != "1.0005e0" || pinned.Vectors[5].Factor != "6.52e0" || pinned.Vectors[0].Factor != "1e0" {
		t.Fatalf("hand-checked vectors drifted: %+v", pinned.Vectors)
	}
	for _, invalid := range []bonusVector{{Level: 1, Spent: 2, PerLevelPPM: 1, UnlockPPM: 1}, {Level: 1, PerLevelPPM: 0}, {Level: 1, PerLevelPPM: 1, UnlockPPM: 1_000_001}} {
		if _, err := BonusFactor(invalid.Level, invalid.Spent, invalid.PerLevelPPM, invalid.UnlockPPM); !errors.Is(err, ErrInvalidState) {
			t.Fatalf("invalid bonus input accepted: %+v", invalid)
		}
	}
}

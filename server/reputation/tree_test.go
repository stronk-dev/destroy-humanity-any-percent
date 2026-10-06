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

func TestTreeStarterClosedWire(t *testing.T) {
	var corpus struct {
		SchemaVersion int `json:"schema_version"`
		Cases         []struct {
			Kind  string          `json:"kind"`
			Key   string          `json:"key"`
			Value json.RawMessage `json:"value"`
		} `json:"cases"`
	}
	if err := json.Unmarshal(readRepository(t, "testdata/reputation/starter-key-rejections-v1.json"), &corpus); err != nil || corpus.SchemaVersion != 1 || len(corpus.Cases) != 20 {
		t.Fatalf("starter-key corpus: %v", err)
	}
	declarations := fixtureDeclarations(t)
	fixture := readRepository(t, "balance/testdata/reputation-tree/fixture-v1.json")
	legal, err := LoadTree(fixture, declarations)
	if err != nil {
		t.Fatal(err)
	}
	kinds := map[string]bool{}
	for _, node := range legal.Nodes() {
		if node.Starter != nil {
			kinds[node.Starter.Kind] = true
		}
	}
	for _, kind := range []string{"resource_grant", "generated_generators", "preowned_upgrade"} {
		if !kinds[kind] {
			t.Fatalf("positive fixture lacks %s", kind)
		}
	}
	for _, row := range corpus.Cases {
		t.Run(row.Kind+"/"+row.Key+"/"+string(row.Value), func(t *testing.T) {
			var root map[string]json.RawMessage
			if err := json.Unmarshal(fixture, &root); err != nil {
				t.Fatal(err)
			}
			var nodes []map[string]json.RawMessage
			if err := json.Unmarshal(root["nodes"], &nodes); err != nil {
				t.Fatal(err)
			}
			matched := false
			for _, node := range nodes {
				if node["starter"] == nil {
					continue
				}
				var starter map[string]json.RawMessage
				if err := json.Unmarshal(node["starter"], &starter); err != nil {
					t.Fatal(err)
				}
				var kind string
				if err := json.Unmarshal(starter["kind"], &kind); err != nil {
					t.Fatal(err)
				}
				if kind != row.Kind {
					continue
				}
				if _, present := starter[row.Key]; present {
					t.Fatalf("test key already legal: %s", row.Key)
				}
				starter[row.Key] = row.Value
				node["starter"], err = json.Marshal(starter)
				if err != nil {
					t.Fatal(err)
				}
				matched = true
				break
			}
			if !matched {
				t.Fatalf("fixture lacks starter %s", row.Kind)
			}
			root["nodes"], err = json.Marshal(nodes)
			if err != nil {
				t.Fatal(err)
			}
			data, err := json.Marshal(root)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := LoadTree(data, declarations); !errors.Is(err, ErrInvalidTree) {
				t.Fatalf("cross-arm starter key admitted: %s %s=%s: %v", row.Kind, row.Key, row.Value, err)
			}
		})
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

type bonusDomainCorpus struct {
	SchemaVersion int     `json:"schema_version"`
	Levels        []int64 `json:"levels"`
	PerLevelPPM   []int64 `json:"per_level_ppm"`
	UnlockPPM     []int64 `json:"unlock_ppm"`
	Invalid       []struct {
		Name string `json:"name"`
		bonusVector
	} `json:"invalid"`
}

func readBonusDomain(t *testing.T) bonusDomainCorpus {
	t.Helper()
	var corpus bonusDomainCorpus
	if err := json.Unmarshal(readRepository(t, "testdata/reputation/bonus-domain-v1.json"), &corpus); err != nil || corpus.SchemaVersion != 1 || len(corpus.Levels) != 7 || len(corpus.PerLevelPPM) != 3 || len(corpus.UnlockPPM) != 7 || len(corpus.Invalid) != 8 {
		t.Fatalf("bonus-domain corpus: %v", err)
	}
	return corpus
}

func TestBonusDoesNotDebitEarnedLevel(t *testing.T) {
	corpus := readBonusDomain(t)
	triples, spends := 0, 0
	for _, level := range corpus.Levels {
		for _, perLevel := range corpus.PerLevelPPM {
			for _, unlock := range corpus.UnlockPPM {
				baseline, err := BonusFactor(level, 0, perLevel, unlock)
				if err != nil {
					t.Fatalf("legal baseline (%d,%d,%d): %v", level, perLevel, unlock, err)
				}
				if (level == 0 || unlock == 0) && baseline.String() != "1e0" {
					t.Fatalf("non-neutral zero input: %s", baseline.String())
				}
				triples++
				seen := map[int64]bool{}
				for _, spent := range []int64{0, 1, level / 2, level} {
					if spent > level || seen[spent] {
						continue
					}
					seen[spent] = true
					available, err := Available(level, spent)
					if err != nil || available != level-spent {
						t.Fatalf("available(%d,%d) = %d: %v", level, spent, available, err)
					}
					factor, err := BonusFactor(level, spent, perLevel, unlock)
					if err != nil || factor.String() != baseline.String() {
						t.Fatalf("spending changed bonus (%d,%d,%d,%d): got %s want %s: %v", level, spent, perLevel, unlock, factor.String(), baseline.String(), err)
					}
					spends++
				}
			}
		}
	}
	if triples != 147 || spends != 462 {
		t.Fatalf("incomplete matrix: %d triples/%d spends", triples, spends)
	}
}

func TestBonusDomainRejections(t *testing.T) {
	for _, row := range readBonusDomain(t).Invalid {
		t.Run(row.Name, func(t *testing.T) {
			if _, err := BonusFactor(row.Level, row.Spent, row.PerLevelPPM, row.UnlockPPM); !errors.Is(err, ErrInvalidState) {
				t.Fatalf("invalid bonus input admitted: %+v: %v", row.bonusVector, err)
			}
			available, err := Available(row.Level, row.Spent)
			if row.Level < 0 || row.Level > decimal.MaxExactInteger || row.Spent < 0 || row.Spent > row.Level {
				if !errors.Is(err, ErrInvalidState) {
					t.Fatalf("invalid accounting admitted: %d: %v", available, err)
				}
			} else if err != nil || available != row.Level-row.Spent {
				t.Fatalf("valid accounting rejected by bonus-only domain: %d: %v", available, err)
			}
		})
	}
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

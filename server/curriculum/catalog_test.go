package curriculum

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"cloud-clicker/server/copykeys"
	"cloud-clicker/server/decimal"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/prestige"
	"cloud-clicker/server/save"
)

func curriculumFixture(t *testing.T) []byte {
	t.Helper()
	data, err := os.ReadFile("../../balance/testdata/t0-t1/curriculum-v2.json")
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func curriculumDeclarations(t *testing.T) Declarations {
	t.Helper()
	_, economyCatalog := loadCandidate(t)
	keys := map[string]struct{}{}
	for _, key := range copykeys.All() {
		keys[key] = struct{}{}
	}
	return Declarations{Economy: economyCatalog, CopyKeys: keys, GateIDs: map[string]struct{}{"gate.t0_to_t1": {}}}
}

func editCurriculumFixture(t *testing.T, edit func(map[string]any)) []byte {
	t.Helper()
	var root map[string]any
	if err := json.Unmarshal(curriculumFixture(t), &root); err != nil {
		t.Fatal(err)
	}
	edit(root)
	data, err := json.Marshal(root)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func curriculumObjects(root map[string]any) []map[string]any {
	failure := root["first_failure"].(map[string]any)
	objects := []map[string]any{root, failure}
	for _, source := range failure["branches"].([]any) {
		branch := source.(map[string]any)
		objects = append(objects, branch, branch["starter_package"].(map[string]any))
	}
	return objects
}

func TestLoadRejectsCaseAliasedObjectMembers(t *testing.T) {
	declarations := curriculumDeclarations(t)
	var root map[string]any
	if err := json.Unmarshal(curriculumFixture(t), &root); err != nil {
		t.Fatal(err)
	}
	for index, object := range curriculumObjects(root) {
		for key := range object {
			t.Run(fmt.Sprintf("%d/%s", index, key), func(t *testing.T) {
				data := editCurriculumFixture(t, func(v map[string]any) {
					row := curriculumObjects(v)[index]
					row[strings.ToUpper(key)] = row[key]
					delete(row, key)
				})
				if _, err := Load(data, declarations); err == nil {
					t.Fatal("case alias admitted")
				}
			})
		}
	}
}

func TestLoadRequiresExplicitZeroFieldsAndClosedBranchStarterShapes(t *testing.T) {
	declarations := curriculumDeclarations(t)
	for name, edit := range map[string]func(map[string]any){
		"missing exit count": func(v map[string]any) { delete(v["first_failure"].(map[string]any), "founder_exit_count") },
		"null exit count":    func(v map[string]any) { v["first_failure"].(map[string]any)["founder_exit_count"] = nil },
		"missing bonus": func(v map[string]any) {
			delete(v["first_failure"].(map[string]any)["branches"].([]any)[0].(map[string]any), "route_knowledge_bonus")
		},
		"null bonus": func(v map[string]any) {
			v["first_failure"].(map[string]any)["branches"].([]any)[0].(map[string]any)["route_knowledge_bonus"] = nil
		},
		"wrong branch field": func(v map[string]any) {
			v["first_failure"].(map[string]any)["branches"].([]any)[1].(map[string]any)["minimum_owned_upgrades"] = 0
		},
		"wrong starter field": func(v map[string]any) {
			v["first_failure"].(map[string]any)["branches"].([]any)[0].(map[string]any)["starter_package"].(map[string]any)["count"] = 0
		},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := Load(editCurriculumFixture(t, edit), declarations); err == nil {
				t.Fatal("missing/null/wrong-arm field admitted")
			}
		})
	}
	for index := range 3 {
		t.Run(fmt.Sprintf("wrong starter kind/%d", index), func(t *testing.T) {
			data := editCurriculumFixture(t, func(v map[string]any) {
				rows := v["first_failure"].(map[string]any)["branches"].([]any)
				rows[index].(map[string]any)["starter_package"] = rows[(index+1)%3].(map[string]any)["starter_package"]
			})
			if _, err := Load(data, declarations); err == nil {
				t.Fatal("starter kind admitted on the wrong ruled branch")
			}
		})
	}
	if catalog, err := Load(curriculumFixture(t), declarations); err != nil || catalog.FirstFailure.FounderExitCount != 0 || catalog.FirstFailure.Branches[0].RouteKnowledgeBonus != 0 {
		t.Fatalf("explicit zero controls refused or changed: %v", err)
	}
}

func TestCandidateSelectsEveryBranchAndAppliesStarter(t *testing.T) {
	catalog, economyCatalog := loadCandidate(t)
	now := time.UnixMilli(1_800_000_000_000)
	founder := &save.State{}
	prior := candidateRunState(t, economyCatalog, now)

	tests := []struct {
		name   string
		mutate func(*save.State)
		branch string
		assert func(*testing.T, *save.State)
	}{
		{name: "acquihire", mutate: func(state *save.State) {
			state.GeneratorPurchasedTotal = 200
			state.UpgradesOwned["upgrade.reply_all_macro"] = true
		}, branch: "acquihire", assert: func(t *testing.T, state *save.State) {
			cash, _ := state.Ledger.Balance("company.cash")
			if cash.String() != "1e4" {
				t.Fatalf("cash=%s", cash.String())
			}
		}},
		{name: "burnout", mutate: func(*save.State) {}, branch: "burnout", assert: func(t *testing.T, state *save.State) {
			if state.GeneratorProvisioned["generator.beige_tower"] != 10 || state.GeneratorCounts["generator.beige_tower"] != 0 {
				t.Fatalf("starter mutated purchased count: %#v", state.GeneratorProvisioned)
			}
		}},
		{name: "pivot", mutate: func(state *save.State) {
			_, err := state.Ledger.Apply(economy.Transaction{Entries: []economy.Entry{{ResourceID: "company.cash", Delta: mustDecimal(t, "1e9")}}})
			if err != nil {
				t.Fatal(err)
			}
		}, branch: "pivot", assert: func(t *testing.T, state *save.State) {
			if !state.UpgradesOwned["upgrade.reply_all_macro"] {
				t.Fatal("pivot starter missing")
			}
		}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			state := cloneCandidateState(t, prior)
			test.mutate(state)
			branch, err := catalog.SelectBranch(state, economyCatalog)
			if err != nil || branch.Branch != test.branch {
				t.Fatalf("branch=%q err=%v", branch.Branch, err)
			}
			newRun, err := prestige.NewRunState(economyCatalog, state, founder, now.Add(time.Second))
			if err != nil {
				t.Fatal(err)
			}
			if err := catalog.ApplyStarter(newRun, branch); err != nil {
				t.Fatal(err)
			}
			test.assert(t, newRun)
		})
	}
}

func TestLoaderPinsGrammarAndReferencesRatherThanCandidateLiterals(t *testing.T) {
	_, economyCatalog := loadCandidate(t)
	data, err := os.ReadFile(filepath.Join("..", "..", "balance", "testdata", "t0-t1", "curriculum-v2.json"))
	if err != nil {
		t.Fatal(err)
	}
	var source map[string]any
	if err := json.Unmarshal(data, &source); err != nil {
		t.Fatal(err)
	}
	failure := source["first_failure"].(map[string]any)
	branches := failure["branches"].([]any)
	acquihire := branches[0].(map[string]any)
	acquihire["minimum_purchased_generators"] = float64(201)
	acquihire["route_knowledge_bonus"] = float64(1)
	acquihire["starter_package"].(map[string]any)["amount"] = "2e4"
	burnout := branches[1].(map[string]any)
	burnout["cheapest_price_factor"] = "3e0"
	burnout["starter_package"].(map[string]any)["count"] = float64(11)
	retuned, err := json.Marshal(source)
	if err != nil {
		t.Fatal(err)
	}
	keys := map[string]struct{}{}
	for _, key := range copykeys.All() {
		keys[key] = struct{}{}
	}
	if _, err := Load(retuned, Declarations{Economy: economyCatalog, CopyKeys: keys, GateIDs: map[string]struct{}{"gate.t0_to_t1": {}}}); err != nil {
		t.Fatalf("valid retune was compiled into the loader: %v", err)
	}

	burnout["starter_package"].(map[string]any)["count"] = float64(0)
	invalid, err := json.Marshal(source)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Load(invalid, Declarations{Economy: economyCatalog, CopyKeys: keys, GateIDs: map[string]struct{}{"gate.t0_to_t1": {}}}); err == nil {
		t.Fatal("zero generated starter count accepted")
	}
}

func loadCandidate(t *testing.T) (*Catalog, *economy.Catalog) {
	t.Helper()
	root := filepath.Join("..", "..")
	economyBytes, err := os.ReadFile(filepath.Join(root, "balance", "catalogs", "phase0.json"))
	if err != nil {
		t.Fatal(err)
	}
	economyCatalog, err := economy.LoadCatalog(economyBytes)
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(root, "balance", "testdata", "t0-t1", "curriculum-v2.json"))
	if err != nil {
		t.Fatal(err)
	}
	keys := map[string]struct{}{}
	for _, key := range copykeys.All() {
		keys[key] = struct{}{}
	}
	catalog, err := Load(data, Declarations{Economy: economyCatalog, CopyKeys: keys, GateIDs: map[string]struct{}{"gate.t0_to_t1": {}}})
	if err != nil {
		t.Fatal(err)
	}
	return catalog, economyCatalog
}

func candidateRunState(t *testing.T, catalog *economy.Catalog, now time.Time) *save.State {
	t.Helper()
	ledger, err := economy.NewLedger(catalog, economy.ScopeCompany)
	if err != nil {
		t.Fatal(err)
	}
	counts, provisioned := map[string]int64{}, map[string]int64{}
	for _, generator := range catalog.GeneratorClassesForScope(economy.ScopeCompany) {
		counts[generator.ID], provisioned[generator.ID] = 0, 0
	}
	return &save.State{Ledger: ledger, GeneratorCounts: counts, GeneratorProvisioned: provisioned, UpgradesOwned: map[string]bool{}, RunSeq: 1, RunStartedAt: now, EvaluatedThrough: now}
}

func cloneCandidateState(t *testing.T, state *save.State) *save.State {
	t.Helper()
	clone := *state
	clone.GeneratorCounts = map[string]int64{}
	clone.GeneratorProvisioned = map[string]int64{}
	clone.UpgradesOwned = map[string]bool{}
	for id, count := range state.GeneratorCounts {
		clone.GeneratorCounts[id] = count
	}
	for id, count := range state.GeneratorProvisioned {
		clone.GeneratorProvisioned[id] = count
	}
	for id, owned := range state.UpgradesOwned {
		clone.UpgradesOwned[id] = owned
	}
	return &clone
}

func mustDecimal(t *testing.T, value string) decimal.Decimal {
	t.Helper()
	parsed, err := decimal.ParseCanonical(value)
	if err != nil {
		t.Fatal(err)
	}
	return parsed
}

package gameui

import (
	"encoding/json"
	"errors"
	"os"
	"testing"

	"cloud-clicker/server/copykeys"
	"cloud-clicker/server/decimal"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/multiplier"
	"cloud-clicker/server/production"
	"cloud-clicker/server/reputation"
	"cloud-clicker/server/save"
)

func reputationTreeForProjection(t *testing.T) *reputation.Tree {
	t.Helper()
	data, err := os.ReadFile("../../balance/catalogs/phase0.json")
	if err != nil {
		t.Fatal(err)
	}
	var root map[string]any
	_ = json.Unmarshal(data, &root)
	root["multiplier_sources"] = append(root["multiplier_sources"].([]any), map[string]any{"id": "reputation.founder_bonus", "slot": "prestige", "target": "all", "provider": reputation.Provider})
	data, _ = json.Marshal(root)
	catalog, err := economy.LoadCatalog(data)
	if err != nil {
		t.Fatal(err)
	}
	keys := map[string]struct{}{}
	for _, key := range copykeys.All() {
		keys[key] = struct{}{}
	}
	treeBytes, err := os.ReadFile("../../balance/testdata/reputation-tree/fixture-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	tree, err := reputation.LoadTree(treeBytes, reputation.Declarations{Economy: catalog, CopyKeys: keys})
	if err != nil {
		t.Fatal(err)
	}
	return tree
}

func TestReputationArmIsServerDerived(t *testing.T) {
	tree := reputationTreeForProjection(t)
	bundle := production.CatalogBundle{ReputationTree: tree}
	founder := &save.State{WireVersion: 22, ReputationLevel: 4, ReputationSpent: 1, ReputationUnlockPPM: 50_000,
		ReputationNodesOwned: []string{"reputation.unlock.p05"}}
	bonus := []multiplier.Contribution{{SourceID: "reputation.founder_bonus", Factor: decimal.One}}
	arm, err := projectReputation(bundle, founder, bonus)
	if err != nil || arm == nil {
		t.Fatalf("arm=%v err=%v", arm, err)
	}
	if arm.Available != 3 || arm.BonusFactorNextRun != "1.002e0" || arm.BonusFactorThisRun == nil || *arm.BonusFactorThisRun != "1e0" {
		t.Fatalf("header=%+v", arm)
	}
	states := map[string]string{}
	for _, row := range arm.Nodes {
		states[row.NodeID] = row.State
	}
	for id, want := range map[string]string{
		"reputation.unlock.p05":                    "owned",
		"reputation.starter.cash_small":            "available",
		"reputation.starter.generated_beige_tower": "locked",
		"reputation.unlock.p25":                    "unaffordable",
		"reputation.unlock.p100":                   "locked",
	} {
		if states[id] != want {
			t.Errorf("%s state=%s want %s", id, states[id], want)
		}
	}
	boundary, err := projectReputation(bundle, &save.State{WireVersion: 22, ReputationLevel: 2, ReputationSpent: 1, ReputationUnlockPPM: 50_000,
		ReputationNodesOwned: []string{"reputation.unlock.p05"}}, nil)
	if err != nil || boundary.Nodes[1].NodeID != "reputation.starter.cash_small" || boundary.Nodes[1].State != "unaffordable" || boundary.BonusFactorThisRun != nil {
		t.Fatalf("cost == available+1 boundary: %+v err=%v", boundary, err)
	}
	if arm, err := projectReputation(production.CatalogBundle{}, founder, nil); arm != nil || err != nil {
		t.Fatalf("tree-less bundle projected %v %v", arm, err)
	}
	if arm, err := projectReputation(bundle, &save.State{WireVersion: 21}, nil); arm != nil || err != nil {
		t.Fatalf("pre-v22 Founder projected %v %v", arm, err)
	}
	if _, err := projectReputation(bundle, &save.State{WireVersion: 22, ReputationLevel: 1, ReputationSpent: 2}, nil); !errors.Is(err, ErrInvalidProjection) {
		t.Fatalf("overspent Founder projected: %v", err)
	}
}

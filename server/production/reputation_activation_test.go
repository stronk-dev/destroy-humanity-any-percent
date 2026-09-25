package production

import (
	"encoding/json"
	"errors"
	"os"
	"testing"
	"time"

	"cloud-clicker/server/copykeys"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/epochseed"
	"cloud-clicker/server/reputation"
	"cloud-clicker/server/save"
)

// reputationContentBundle is the live epoch-8 bundle plus the fixture
// reputation_tree and its one economy declaration row (fixture-first; no
// epoch pins it yet).
func reputationContentBundle(t *testing.T) CatalogBundle {
	t.Helper()
	seed, err := epochseed.Load("../..")
	if err != nil {
		t.Fatal(err)
	}
	artifacts := map[string][]byte{}
	for name, data := range seed.Artifacts {
		artifacts[name] = data
	}
	var root map[string]any
	if err := json.Unmarshal(artifacts["economy"], &root); err != nil {
		t.Fatal(err)
	}
	root["multiplier_sources"] = append(root["multiplier_sources"].([]any), map[string]any{"id": "reputation.founder_bonus", "slot": "prestige", "target": "all", "provider": reputation.Provider})
	artifacts["economy"], _ = json.Marshal(root)
	artifacts["reputation_tree"], err = os.ReadFile("../../balance/testdata/reputation-tree/fixture-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	hash, err := save.ConstantsHashArtifacts(artifacts)
	if err != nil {
		t.Fatal(err)
	}
	bundle := activeContentBundle(t)
	bundle.ConstantsHash, bundle.Artifacts = hash, artifacts
	if bundle.Economy, err = economy.LoadCatalog(artifacts["economy"]); err != nil {
		t.Fatal(err)
	}
	keys := map[string]struct{}{}
	for _, key := range copykeys.All() {
		keys[key] = struct{}{}
	}
	bundle.ReputationTree, err = reputation.LoadTree(artifacts["reputation_tree"], reputation.Declarations{Economy: bundle.Economy, Curriculum: bundle.Curriculum, CopyKeys: keys})
	if err != nil {
		t.Fatal(err)
	}
	if !bundle.valid(hash) {
		t.Fatal("reputation bundle is not valid")
	}
	return bundle
}

func TestReputationTreeOwnsFounderV22Activation(t *testing.T) {
	active := activeContentBundle(t)
	tree := reputationContentBundle(t)
	if founder, _ := tree.versionFloors(); founder != 22 {
		t.Fatalf("reputation bundle founder floor = %d", founder)
	}

	// Run boundary: a v21 Founder under epoch 8 activates v22 under the tree.
	legacy := epoch5TestBundle(t)
	now := time.Date(2026, 9, 25, 12, 0, 0, 0, time.UTC)
	founder := foundationScopeState(t, legacy.Economy, economy.ScopeFounder)
	company := foundationScopeState(t, legacy.Economy, economy.ScopeCompany)
	company.RunStartedAt = now.Add(-2 * time.Hour)
	middle := foundationScopeState(t, active.Economy, economy.ScopeCompany)
	middle.RunStartedAt = now.Add(-time.Hour)
	if err := settleAndActivateFoundations(legacy, active, founder, company, middle); err != nil {
		t.Fatal(err)
	}
	founder.ReputationLevel = 4
	next := foundationScopeState(t, tree.Economy, economy.ScopeCompany)
	next.RunStartedAt = now
	if err := settleAndActivateFoundations(active, tree, founder, middle, next); err != nil {
		t.Fatal(err)
	}
	if save.VersionForState(founder) != 22 || founder.ReputationSpent != 0 || founder.ReputationNodesOwned == nil || len(founder.ReputationNodesOwned) != 0 || founder.ReputationLevel != 4 {
		t.Fatalf("v22 activation founder version=%d spent=%d owned=%v level=%d", save.VersionForState(founder), founder.ReputationSpent, founder.ReputationNodesOwned, founder.ReputationLevel)
	}
	if err := tree.ValidateFoundationState(founder); err != nil {
		t.Fatalf("activated Founder invalid: %v", err)
	}
	// Replay activation (Founder log Exit arm) mirrors the live boundary.
	replayed := &save.State{WireVersion: 21, ReputationLevel: 3}
	if err := activateFounderFeatureState(replayed, tree, 22, 1, nil); err != nil || replayed.ReputationNodesOwned == nil || replayed.ReputationSpent != 0 {
		t.Fatalf("replay activation: %v %+v", err, replayed)
	}
	if err := activateFounderFeatureState(&save.State{WireVersion: 21}, active, 22, 1, nil); !errors.Is(err, ErrInvalidReplayInputs) {
		t.Fatalf("v22 activated without a tree: %v", err)
	}
	if err := activateFounderFeatureState(&save.State{WireVersion: 21, ReputationUnlockPPM: 50_000}, tree, 22, 1, nil); !errors.Is(err, ErrInvalidReplayInputs) {
		t.Fatalf("v22 activation accepted a pre-activation unlock mirror: %v", err)
	}
}

func TestReputationPinnedTreeValidatesTheUnlockMirrorAndAccounting(t *testing.T) {
	tree := reputationContentBundle(t)
	founder := &save.State{ReputationLevel: 10, ReputationSpent: 6, ReputationNodesOwned: []string{"reputation.unlock.p05", "reputation.unlock.p25"}, ReputationUnlockPPM: 250_000}
	if err := validateFounderReputationState(tree.ReputationTree, founder); err != nil {
		t.Fatalf("valid mirror rejected: %v", err)
	}
	for name, mutate := range map[string]func(*save.State){
		"mirror too low":      func(state *save.State) { state.ReputationUnlockPPM = 50_000 },
		"mirror without node": func(state *save.State) { state.ReputationNodesOwned = []string{"reputation.starter.cash_small"} },
		"spent over level":    func(state *save.State) { state.ReputationSpent = 11 },
		"nil owned set":       func(state *save.State) { state.ReputationNodesOwned = nil; state.ReputationUnlockPPM = 0 },
	} {
		candidate := *founder
		candidate.ReputationNodesOwned = append([]string(nil), founder.ReputationNodesOwned...)
		mutate(&candidate)
		if err := validateFounderReputationState(tree.ReputationTree, &candidate); !errors.Is(err, ErrInvalidEngineState) {
			t.Fatalf("%s accepted: %v", name, err)
		}
	}
	if err := validateFounderReputationState(nil, founder); !errors.Is(err, ErrInvalidEngineState) {
		t.Fatalf("tree state accepted without a pinned tree: %v", err)
	}
	if err := validateFounderReputationState(nil, &save.State{}); err != nil {
		t.Fatalf("empty state rejected without a tree: %v", err)
	}
}

func TestFounderCarryFailsClosedForV22UntilTheReplayInputsCarryTreeState(t *testing.T) {
	tree := reputationContentBundle(t)
	carry := replayFounderCarry{FounderRevision: 1, ReputationLevel: 3, NetworkSlots: []save.NetworkSlot{}, LedgerFactKinds: []string{},
		AchievementsEarnedLifetime: []string{}, FounderExtensions: &replayFounderExtensions{}}
	if _, err := stateFromFounderCarry(carry, tree); !errors.Is(err, ErrInvalidReplayInputs) {
		t.Fatalf("v22 carry reconstructed without tree fields: %v", err)
	}
}

func TestReputationFrozenFounderContributionsAddTheBonusRow(t *testing.T) {
	tree := reputationContentBundle(t)
	live := activeContentBundle(t)
	now := time.Date(2026, 9, 25, 12, 0, 0, 0, time.UTC)
	liveFounder := reputationFounderState(t, live, 21, now, 3)
	plain, err := FrozenFounderContributions(live, liveFounder)
	if err != nil || len(plain) != len(live.Fiscal.GeneratorLevelRows())+1 {
		t.Fatalf("tree-less bundle rows=%v err=%v", plain, err)
	}
	founder := reputationFounderState(t, tree, 22, now, 552)
	rows, err := FrozenFounderContributions(tree, founder)
	if err != nil || len(rows) != len(plain)+1 {
		t.Fatalf("tree bundle rows=%v err=%v", rows, err)
	}
	factor := func(values []save.FrozenContribution) string {
		for _, row := range values {
			if row.SourceID == "reputation.founder_bonus" {
				return row.Factor
			}
		}
		return ""
	}
	if factor(rows) != "1e0" {
		t.Fatalf("no unlock node factor = %q", factor(rows))
	}
	founder.ReputationSpent, founder.ReputationNodesOwned = 552, []string{"reputation.starter.cash_large", "reputation.starter.cash_small",
		"reputation.starter.generated_beige_tower", "reputation.starter.upgrade_continuous_feed_paper", "reputation.unlock.p05",
		"reputation.unlock.p100", "reputation.unlock.p25", "reputation.unlock.p50", "reputation.unlock.p75"}
	founder.ReputationUnlockPPM = 1_000_000
	if rows, err = FrozenFounderContributions(tree, founder); err != nil || factor(rows) != "6.52e0" {
		t.Fatalf("full tree factor = %q err=%v (from the earned level, not available)", factor(rows), err)
	}
	founder.ReputationUnlockPPM = 750_000
	if _, err := FrozenFounderContributions(tree, founder); err == nil {
		t.Fatal("a mismatched unlock mirror froze a factor")
	}
}

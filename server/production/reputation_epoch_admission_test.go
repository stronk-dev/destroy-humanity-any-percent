package production

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"slices"
	"sort"
	"testing"
	"time"

	"cloud-clicker/server/economy"
	"cloud-clicker/server/meters"
	"cloud-clicker/server/reputation"
	"cloud-clicker/server/save"
)

func reputationEpochPopulation(t *testing.T, current CatalogBundle) ([]string, []string) {
	t.Helper()
	data, err := os.ReadFile("../../testdata/reputation/epoch-transitions-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	var table struct {
		Version  int      `json:"version"`
		Removed  []string `json:"removed_node_ids"`
		Profiles []string `json:"valid_profiles"`
	}
	if err := json.Unmarshal(data, &table); err != nil {
		t.Fatal(err)
	}
	ids := []string{}
	for _, node := range current.ReputationTree.Nodes() {
		ids = append(ids, node.NodeID)
	}
	if table.Version != 1 || len(table.Removed) != 9 || !slices.Equal(table.Removed, ids) ||
		!slices.Equal(table.Profiles, []string{"unchanged", "retune", "append", "activation", "inactive"}) {
		t.Fatal("incomplete epoch transition population")
	}
	return table.Removed, table.Profiles
}

// The candidate is valid as an independent tree. Rebuild prerequisite links
// and the terminal unlock ladder so only the previous->next ID loss is tested.
func reputationEpochCandidate(t *testing.T, current CatalogBundle, removedID string, appendNode bool) CatalogBundle {
	t.Helper()
	next := current
	next.Next, next.Artifacts = nil, cloneArtifactMap(current.Artifacts)
	var root map[string]any
	if err := json.Unmarshal(next.Artifacts["reputation_tree"], &root); err != nil {
		t.Fatal(err)
	}
	kept := []any{}
	previous := ""
	var lastUnlock, cashNode map[string]any
	for _, raw := range root["nodes"].([]any) {
		node := raw.(map[string]any)
		if node["node_id"] == removedID {
			continue
		}
		requires := []string{}
		for _, requirement := range node["requires"].([]any) {
			if requirement != removedID {
				requires = append(requires, requirement.(string))
			}
		}
		if node["kind"] == reputation.KindBonusUnlock {
			if previous != "" && !slices.Contains(requires, previous) {
				requires = append(requires, previous)
			}
			previous, lastUnlock = node["node_id"].(string), node
		}
		sort.Strings(requires)
		node["requires"] = requires
		if node["node_id"] == "reputation.starter.cash_small" {
			cashNode = node
		}
		kept = append(kept, node)
	}
	lastUnlock["unlock_ppm"] = 1_000_000
	if appendNode {
		added := map[string]any{}
		for key, value := range cashNode {
			added[key] = value
		}
		added["node_id"] = "reputation.starter.appended_cash"
		kept = append(kept, added)
	}
	root["nodes"] = kept
	next.Artifacts["reputation_tree"] = reputationStarterJSON(t, root)
	var err error
	next.ReputationTree, err = reputation.LoadTree(next.Artifacts["reputation_tree"], reputation.Declarations{
		Economy: next.Economy, Curriculum: next.Curriculum, CopyKeys: reputationStarterCopyKeys(),
	})
	if err != nil {
		t.Fatalf("candidate must be valid in isolation: %v", err)
	}
	next.ConstantsHash, err = save.ConstantsHashArtifacts(next.Artifacts)
	if err != nil || !next.valid(next.ConstantsHash) {
		t.Fatalf("invalid candidate hash/bundle: %v", err)
	}
	return next
}

func reputationEpochCompany(t *testing.T, bundle CatalogBundle, now time.Time) *save.State {
	t.Helper()
	state := replayFixtureState(t, bundle.Economy, now.Add(-time.Hour))
	_, state.WireVersion = bundle.versionFloors()
	state.MeterBands = nil
	value, err := meters.NewRunState(bundle.Meters, 0)
	if err != nil {
		t.Fatal(err)
	}
	state.MeterValues, state.MeterDecayRemainders, state.MeterInputRemainders = value.Values, value.DecayRemainders, value.InputRemainders
	state.AchievementsEarnedRun = map[string]bool{}
	if err := bundle.ValidateFoundationState(state); err != nil {
		t.Fatal(err)
	}
	return state
}

func reputationEpochFounderCommand(t *testing.T, state *save.State, nextHash string, now time.Time) (IntentRequest, []byte) {
	t.Helper()
	command := save.FounderReplayCommand{IntentID: "01986666-8d01-7000-8000-000000000001",
		FounderStreamID: "01986666-8c00-4000-8000-000000000001", FounderID: "01986666-8d00-7000-8000-000000000001",
		Revision: 1, FounderLogSeq: 1, ServerTSMS: now.UnixMilli()}
	resolved := founderExitResolvedWire{Kind: founderExitResolvedKind, Outcome: string(save.IntentApplied),
		CompanyStreamID: "01986666-8e00-7000-8000-000000000001", RunSeq: 1, RunLogSeq: 1,
		ResultConstantsHash: nextHash, AgeMSBefore: state.AgeMS, AgeMSAfter: state.AgeMS,
		AddedNetworkSlots: []save.NetworkSlot{}, AddedLedgerFactKinds: []string{}, AddedLifetimeAchievements: []string{},
		ExitRecord: &founderExitRecordWire{RunID: 1, ExitType: "collapse", OccurredAtMS: now.UnixMilli()}, ResultFounderWireVersion: 22}
	inputs, err := save.MarshalFounderReplayInputs(command, resolved)
	if err != nil {
		t.Fatal(err)
	}
	request, err := ParseIntent([]byte(`{"intent_id":"01986666-8d01-7000-8000-000000000001","kind":"wind_down","expected_revision":1,"expected_founder_revision":1}`))
	if err != nil || request.InvalidDetail != "" {
		t.Fatalf("invalid epoch fixture request: %v/%s", err, request.InvalidDetail)
	}
	return request, inputs
}

func TestReputationEpochRetainsEveryNodeID(t *testing.T) {
	current := reputationContentBundle(t)
	ids, profiles := reputationEpochPopulation(t, current)
	now := time.Date(2026, 10, 6, 12, 0, 0, 0, time.UTC)
	for _, id := range ids {
		next := reputationEpochCandidate(t, current, id, false)
		if _, exists := next.ReputationTree.Node(id); exists || len(next.ReputationTree.Nodes()) != 8 {
			t.Fatal("counterfact did not remove exactly one node")
		}
		source := current
		source.Next = &next
		for _, arm := range []string{"linked_loader", "live", "founder_replay"} {
			t.Run(id+"/"+arm, func(t *testing.T) {
				if arm == "linked_loader" {
					if source.valid(source.ConstantsHash) {
						t.Error("linked loader admitted a removed node")
					}
					if _, ok := (ReplayCatalogSet{source.ConstantsHash: source}).ResolveReplayCatalogs(source.ConstantsHash); ok {
						t.Error("linked resolver admitted a removed node")
					}
					return
				}
				state := reputationFounderState(t, current, 22, now, 11)
				state.ReputationSpent, state.ReputationNodesOwned = 4, []string{"reputation.retired.unknown"}
				if err := current.ValidateFoundationState(state); err != nil {
					t.Fatal(err)
				}
				before := mustEncodeState(t, state)
				if arm == "live" {
					company := reputationEpochCompany(t, current, now)
					created := foundationScopeState(t, next.Economy, economy.ScopeCompany)
					created.RunStartedAt = now
					if err := settleAndActivateFoundations(source, next, state, company, created); !errors.Is(err, ErrInvalidEngineState) {
						t.Fatalf("live boundary admitted removal: %v", err)
					}
				} else {
					request, inputs := reputationEpochFounderCommand(t, state, next.ConstantsHash, now)
					result, err := ApplyFounderLogged(state, request.CanonicalPayload, source, inputs)
					if !errors.Is(err, ErrInvalidReplayInputs) || len(result.Receipt) != 0 || len(result.Events) != 0 {
						t.Fatalf("Founder admitted removal: %v/%s events=%d receipt=%s", err, result.Outcome, len(result.Events), result.Receipt)
					}
				}
				if !bytes.Equal(before, mustEncodeState(t, state)) {
					t.Fatal("refused epoch mutated Founder state")
				}
			})
		}
	}
	t.Run("whole_tree_withdrawal", func(t *testing.T) {
		next := activeContentBundle(t)
		source := current
		source.Next = &next
		if source.valid(source.ConstantsHash) {
			t.Fatal("linked loader admitted whole-tree withdrawal")
		}
	})
	for _, profile := range profiles {
		t.Run("valid/"+profile, func(t *testing.T) {
			source, next := current, current
			switch profile {
			case "unchanged":
			case "retune":
				next = reputationRetunedEpoch(t, current, 60_000)
			case "append":
				next = reputationEpochCandidate(t, current, "", true)
			case "activation":
				source = activeContentBundle(t)
			case "inactive":
				source = activeContentBundle(t)
				next = source
			default:
				t.Fatalf("unknown positive epoch profile: %s", profile)
			}
			if !next.valid(next.ConstantsHash) {
				t.Fatal("standalone next artifact must remain loadable")
			}
			source.Next = &next
			if !source.valid(source.ConstantsHash) {
				t.Fatal("valid epoch transition rejected")
			}
		})
	}
}

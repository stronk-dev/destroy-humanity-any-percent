package harness

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
)

// TestTier2RelevanceCandidateAddsExactlyTheSixTierTwoRows loads the §P3
// combined scenario and requires its candidate policy to equal the epoch-8
// policy plus exactly the six Tier-2 rows, each windowed gate.t1_to_t2 →
// gate.t2_to_t3 with epsilon 1000 ms and no trap exemption.
func TestTier2RelevanceCandidateAddsExactlyTheSixTierTwoRows(t *testing.T) {
	suite, err := LoadRelevanceSuite(repositoryRootForReputation, "balance/testdata/t2/relevance-scenario-t1-t2-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	if len(suite.Scenario.Segments) != 3 || *suite.Scenario.Segments[1].ToGate != "gate.t1_to_t2" || *suite.Scenario.Segments[2].FromGate != "gate.t1_to_t2" {
		t.Fatalf("segments = %+v", suite.Scenario.Segments)
	}
	live, candidate := readTier2RelevancePolicy(t, "balance/relevance/t0-t1.json"), readTier2RelevancePolicy(t, "balance/testdata/t2/relevance-candidate-v1.json")
	if err := tier2RelevanceExtension(live, candidate); err != nil {
		t.Fatal(err)
	}
}

func readTier2RelevancePolicy(t *testing.T, path string) RelevancePolicy {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(repositoryRootForReputation, path))
	if err != nil {
		t.Fatal(err)
	}
	var policy RelevancePolicy
	if err := json.Unmarshal(data, &policy); err != nil {
		t.Fatal(err)
	}
	return policy
}

// Kept separate so invalid candidates exercise the same check as the real artifact.
func tier2RelevanceExtension(livePolicy, candidatePolicy RelevancePolicy) error {
	if candidatePolicy.SchemaVersion != livePolicy.SchemaVersion || !reflect.DeepEqual(candidatePolicy.Groups, livePolicy.Groups) {
		return fmt.Errorf("changed inherited policy schema or groups")
	}
	live, candidate := map[string]RelevancePolicyItem{}, map[string]RelevancePolicyItem{}
	for _, item := range livePolicy.Items {
		live[item.PurchasableID] = item
	}
	for _, item := range candidatePolicy.Items {
		if _, exists := candidate[item.PurchasableID]; exists {
			return fmt.Errorf("duplicate candidate row %s", item.PurchasableID)
		}
		candidate[item.PurchasableID] = item
	}
	for id, original := range live {
		if item, exists := candidate[id]; !exists || !reflect.DeepEqual(item, original) {
			return fmt.Errorf("changed or missing inherited row %s", id)
		}
	}
	added := []string{}
	for id, item := range candidate {
		if _, ok := live[id]; ok {
			continue
		}
		if item.Availability.FromGate == nil || *item.Availability.FromGate != "gate.t1_to_t2" || item.Availability.ToGate == nil || *item.Availability.ToGate != "gate.t2_to_t3" || item.EpsilonMS != 1000 || item.TrapExempt || item.JustificationKey != nil || len(item.GroupIDs) != 0 {
			return fmt.Errorf("tier-2 relevance row %s = %+v", id, item)
		}
		added = append(added, id)
	}
	sort.Strings(added)
	want := []string{"generator.hot_desk_program", "generator.managed_services_contract", "generator.open_plan_floor", "upgrade.move_fast_break_things", "upgrade.nap_pod", "upgrade.ping_pong_table"}
	if len(added) != len(want) || len(candidate) != len(live)+len(want) {
		return fmt.Errorf("added rows %v, want %v", added, want)
	}
	for index := range want {
		if added[index] != want[index] {
			return fmt.Errorf("added rows %v, want %v", added, want)
		}
	}
	return nil
}

func TestTier2RelevanceExtensionRejectsInheritedPolicyChanges(t *testing.T) {
	live := readTier2RelevancePolicy(t, "balance/relevance/t0-t1.json")
	for _, test := range []struct {
		name   string
		change func(*RelevancePolicy)
		want   string
	}{
		{"epsilon", func(p *RelevancePolicy) { p.Items[0].EpsilonMS++ }, "changed or missing inherited row"},
		{"window", func(p *RelevancePolicy) { p.Items[0].Availability.ToGate = nil }, "changed or missing inherited row"},
		{"trap", func(p *RelevancePolicy) {
			key := "relevance.intentional_trap"
			p.Items[0].TrapExempt, p.Items[0].JustificationKey = true, &key
		}, "changed or missing inherited row"},
		{"membership", func(p *RelevancePolicy) { p.Items[0].GroupIDs = []string{"group.changed"} }, "changed or missing inherited row"},
		{"missing", func(p *RelevancePolicy) { p.Items = p.Items[1:] }, "changed or missing inherited row"},
		{"duplicate", func(p *RelevancePolicy) { p.Items = append(p.Items, p.Items[0]) }, "duplicate candidate row"},
		{"schema", func(p *RelevancePolicy) { p.SchemaVersion++ }, "changed inherited policy schema or groups"},
		{"groups", func(p *RelevancePolicy) {
			p.Groups = append(p.Groups, RelevancePolicyGroup{GroupID: "group.changed"})
		}, "changed inherited policy schema or groups"},
		{"new-row-threshold", func(p *RelevancePolicy) {
			for index := range p.Items {
				if p.Items[index].PurchasableID == "upgrade.nap_pod" {
					p.Items[index].EpsilonMS++
				}
			}
		}, "tier-2 relevance row upgrade.nap_pod"},
		{"new-row-exemption", func(p *RelevancePolicy) {
			for index := range p.Items {
				if p.Items[index].PurchasableID == "upgrade.nap_pod" {
					key := "relevance.intentional_trap"
					p.Items[index].TrapExempt, p.Items[index].JustificationKey = true, &key
				}
			}
		}, "tier-2 relevance row upgrade.nap_pod"},
		{"missing-new-row", func(p *RelevancePolicy) {
			for index, item := range p.Items {
				if item.PurchasableID == "upgrade.nap_pod" {
					p.Items = append(p.Items[:index], p.Items[index+1:]...)
					break
				}
			}
		}, "added rows"},
	} {
		t.Run(test.name, func(t *testing.T) {
			candidate := readTier2RelevancePolicy(t, "balance/testdata/t2/relevance-candidate-v1.json")
			test.change(&candidate)
			if err := tier2RelevanceExtension(live, candidate); err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("mutation %s: got %v, want %q", test.name, err, test.want)
			}
		})
	}
}

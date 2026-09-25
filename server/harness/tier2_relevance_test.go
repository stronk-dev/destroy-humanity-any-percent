package harness

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
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
	type row struct {
		PurchasableID string `json:"purchasable_id"`
		Window        struct {
			From *string `json:"from_gate"`
			To   *string `json:"to_gate"`
		} `json:"availability_window"`
		EpsilonMS  int64 `json:"epsilon_ms"`
		TrapExempt bool  `json:"trap_exempt"`
	}
	read := func(path string) map[string]row {
		data, err := os.ReadFile(filepath.Join(repositoryRootForReputation, path))
		if err != nil {
			t.Fatal(err)
		}
		var policy struct {
			Items []row `json:"items"`
		}
		if err := json.Unmarshal(data, &policy); err != nil {
			t.Fatal(err)
		}
		result := map[string]row{}
		for _, item := range policy.Items {
			result[item.PurchasableID] = item
		}
		return result
	}
	live, candidate := read("balance/relevance/t0-t1.json"), read("balance/testdata/t2/relevance-candidate-v1.json")
	added := []string{}
	for id, item := range candidate {
		if _, ok := live[id]; ok {
			continue
		}
		if item.Window.From == nil || *item.Window.From != "gate.t1_to_t2" || item.Window.To == nil || *item.Window.To != "gate.t2_to_t3" || item.EpsilonMS != 1000 || item.TrapExempt {
			t.Fatalf("tier-2 relevance row %s = %+v", id, item)
		}
		added = append(added, id)
	}
	sort.Strings(added)
	want := []string{"generator.hot_desk_program", "generator.managed_services_contract", "generator.open_plan_floor", "upgrade.move_fast_break_things", "upgrade.nap_pod", "upgrade.ping_pong_table"}
	if len(added) != len(want) || len(candidate) != len(live)+len(want) {
		t.Fatalf("added rows %v, want %v", added, want)
	}
	for index := range want {
		if added[index] != want[index] {
			t.Fatalf("added rows %v, want %v", added, want)
		}
	}
}

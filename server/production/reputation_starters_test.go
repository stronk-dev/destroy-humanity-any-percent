package production

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"strings"
	"testing"
	"time"

	"cloud-clicker/server/copykeys"
	"cloud-clicker/server/curriculum"
	"cloud-clicker/server/decimal"
	"cloud-clicker/server/economy"
	prestigecore "cloud-clicker/server/prestige"
	"cloud-clicker/server/reputation"
	"cloud-clicker/server/save"
)

type reputationStarterBoundary struct {
	Profile     string   `json:"profile"`
	Branch      string   `json:"branch"`
	PreCash     string   `json:"pre_cash"`
	Cash        string   `json:"cash"`
	Permits     string   `json:"permits"`
	Provisioned int64    `json:"provisioned"`
	Upgrades    []string `json:"upgrades"`
	Applied     []string `json:"applied"`
}

func reputationStarterBoundaryCases(t *testing.T) []reputationStarterBoundary {
	t.Helper()
	data, err := os.ReadFile("../../testdata/reputation/starter-boundaries-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	var table struct {
		Version int                         `json:"version"`
		Cases   []reputationStarterBoundary `json:"cases"`
	}
	if err := json.Unmarshal(data, &table); err != nil {
		t.Fatal(err)
	}
	if table.Version != 1 || len(table.Cases) != 4 {
		t.Fatal("starter boundary population changed")
	}
	return table.Cases
}

// Change only fixture artifacts, then strictly reload the changed catalogs and
// derive their actual pin. This is neither a production retune nor a mint.
func reputationStarterNextBundle(t *testing.T, current CatalogBundle, profile string) CatalogBundle {
	t.Helper()
	next := current
	next.Next = nil
	next.Artifacts = cloneArtifactMap(current.Artifacts)
	var tree map[string]any
	if err := json.Unmarshal(next.Artifacts["reputation_tree"], &tree); err != nil {
		t.Fatal(err)
	}
	nodes := tree["nodes"].([]any)
	switch profile {
	case "retire":
		kept := []any{}
		for _, raw := range nodes {
			id := raw.(map[string]any)["node_id"]
			if id != "reputation.starter.generated_beige_tower" && id != "reputation.starter.upgrade_continuous_feed_paper" {
				kept = append(kept, raw)
			}
		}
		tree["nodes"] = kept
	case "idempotent":
		var content map[string]any
		if err := json.Unmarshal(next.Artifacts["curriculum"], &content); err != nil {
			t.Fatal(err)
		}
		pivot := content["first_failure"].(map[string]any)["branches"].([]any)[2].(map[string]any)
		pivot["starter_package"].(map[string]any)["upgrade_id"] = "upgrade.continuous_feed_paper"
		next.Artifacts["curriculum"] = reputationStarterJSON(t, content)
	case "generated_cap":
		for _, raw := range nodes {
			node := raw.(map[string]any)
			if node["node_id"] == "reputation.starter.generated_beige_tower" {
				node["starter"].(map[string]any)["count"] = decimal.MaxExactInteger - 10
			}
		}
	case "resource_cap":
		for _, raw := range nodes {
			node := raw.(map[string]any)
			if node["node_id"] == "reputation.starter.cash_small" {
				node["starter"] = map[string]any{"kind": "resource_grant", "resource_id": "company.permits", "amount": "2.4e1"}
			}
		}
	default:
		t.Fatalf("unknown boundary profile %s", profile)
	}
	next.Artifacts["reputation_tree"] = reputationStarterJSON(t, tree)
	keys := reputationStarterCopyKeys()
	gateIDs := map[string]struct{}{}
	for _, gate := range next.Routes.Gates() {
		gateIDs[gate.ID] = struct{}{}
	}
	var err error
	next.Curriculum, err = curriculum.Load(next.Artifacts["curriculum"], curriculum.Declarations{Economy: next.Economy, CopyKeys: keys, GateIDs: gateIDs})
	if err != nil {
		t.Fatal(err)
	}
	next.ReputationTree, err = reputation.LoadTree(next.Artifacts["reputation_tree"], reputation.Declarations{Economy: next.Economy, Curriculum: next.Curriculum, CopyKeys: keys})
	if err != nil {
		t.Fatal(err)
	}
	next.ConstantsHash, err = save.ConstantsHashArtifacts(next.Artifacts)
	if err != nil || !next.valid(next.ConstantsHash) || next.ConstantsHash == current.ConstantsHash {
		t.Fatalf("invalid next fixture pin: %v", err)
	}
	return next
}

func reputationStarterJSON(t *testing.T, value any) []byte {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func reputationStarterCopyKeys() map[string]struct{} {
	keys := map[string]struct{}{}
	for _, key := range copykeys.All() {
		keys[key] = struct{}{}
	}
	return keys
}

func TestReputationStarterNextBundles(t *testing.T) {
	now := time.Date(2026, 9, 25, 12, 0, 0, 0, time.UTC)
	fixture := makeReputationExitFixture(t, now)
	base := reputationContentBundle(t)
	for _, row := range reputationStarterBoundaryCases(t) {
		t.Run(row.Profile, func(t *testing.T) {
			current := base
			next := reputationStarterNextBundle(t, current, row.Profile)
			current.Next = &next
			wire, err := parseReplayInputs(fixture.Case.ReplayInputs)
			if err != nil {
				t.Fatal(err)
			}
			var resolved replayExitResolved
			if err := json.Unmarshal(wire.Resolved, &resolved); err != nil {
				t.Fatal(err)
			}
			founder := reputationFounderState(t, current, 22, now, 552)
			founder.ReputationSpent, founder.ReputationUnlockPPM = 552, 1_000_000
			founder.ReputationNodesOwned = []string{"reputation.retired.unknown", "reputation.starter.cash_large", "reputation.starter.cash_small", "reputation.starter.generated_beige_tower", "reputation.starter.upgrade_continuous_feed_paper", "reputation.unlock.p05", "reputation.unlock.p100", "reputation.unlock.p25", "reputation.unlock.p50", "reputation.unlock.p75"}
			carry := founderCarry(founder)
			carry.FounderRevision, carry.FounderConstantsHash = resolved.FounderCarry.FounderRevision, current.ConstantsHash
			resolved.FounderCarry, resolved.NextConstantsHash, resolved.SelectedBranch = carry, next.ConstantsHash, &row.Branch
			wire.Resolved = reputationStarterJSON(t, resolved)
			company := replayFixtureStateFromEncoded(t, current, fixture.Case.PreState)
			cash, err := decimal.ParseCanonical(row.PreCash)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := company.Ledger.Apply(economy.Transaction{Entries: []economy.Entry{{ResourceID: "company.cash", Delta: cash}}}); err != nil {
				t.Fatal(err)
			}
			transition, err := ApplyLoggedExit(company, fixture.Case.CanonicalPayload, current, reputationStarterJSON(t, wire))
			if err != nil {
				t.Fatal(err)
			}
			if transition.Decision.Outcome != save.IntentApplied {
				t.Fatalf("outcome=%s", transition.Decision.Outcome)
			}
			created := transition.Decision.NewCompanyState
			for id, expected := range map[string]string{"company.cash": row.Cash, "company.permits": row.Permits} {
				actual, ok := created.Ledger.Balance(id)
				if !ok || actual.String() != expected {
					t.Fatalf("%s=%s want=%s", id, actual, expected)
				}
			}
			if created.GeneratorProvisioned["generator.beige_tower"] != row.Provisioned || created.GeneratorCounts["generator.beige_tower"] != 0 || created.GeneratorPurchasedTotal != 0 {
				t.Fatalf("generated=%d want=%d purchased=%d total=%d", created.GeneratorProvisioned["generator.beige_tower"], row.Provisioned, created.GeneratorCounts["generator.beige_tower"], created.GeneratorPurchasedTotal)
			}
			if len(created.UpgradesOwned) != len(row.Upgrades) {
				t.Fatalf("upgrades=%v want=%v", created.UpgradesOwned, row.Upgrades)
			}
			for _, id := range row.Upgrades {
				if !created.UpgradesOwned[id] {
					t.Fatalf("missing upgrade %s", id)
				}
			}
			var payload struct {
				ReputationTree reputationRunStarted `json:"reputation_tree"`
			}
			started := transition.Decision.CompanyStartedEvents[0]
			if err := json.Unmarshal(started.Payload, &payload); err != nil {
				t.Fatal(err)
			}
			if started.SchemaVersion != 2 || payload.ReputationTree.BonusFactor != "6.52e0" || !slices.Equal(payload.ReputationTree.AppliedStarterNodeIDs, row.Applied) {
				t.Fatalf("wrong started summary: %+v", payload.ReputationTree)
			}
			if transition.Founder.ReputationLevel != 552 || transition.Founder.ReputationSpent != 552 || !slices.Equal(transition.Founder.ReputationNodesOwned, founder.ReputationNodesOwned) {
				t.Fatal("next tree rewrote Founder ownership/accounting")
			}
		})
	}
}

func TestReputationStarterOverCapArtifacts(t *testing.T) {
	current := reputationContentBundle(t)
	for _, profile := range []string{"generated_cap", "resource_cap"} {
		t.Run(profile, func(t *testing.T) {
			next := reputationStarterNextBundle(t, current, profile)
			var tree map[string]any
			if err := json.Unmarshal(next.Artifacts["reputation_tree"], &tree); err != nil {
				t.Fatal(err)
			}
			for _, raw := range tree["nodes"].([]any) {
				node := raw.(map[string]any)
				if profile == "generated_cap" && node["node_id"] == "reputation.starter.generated_beige_tower" {
					node["starter"].(map[string]any)["count"] = decimal.MaxExactInteger - 9
				}
				if profile == "resource_cap" && node["node_id"] == "reputation.starter.cash_small" {
					node["starter"].(map[string]any)["amount"] = "2.5e1"
				}
			}
			_, err := reputation.LoadTree(reputationStarterJSON(t, tree), reputation.Declarations{Economy: next.Economy, Curriculum: next.Curriculum, CopyKeys: reputationStarterCopyKeys()})
			if !errors.Is(err, reputation.ErrInvalidTree) || !strings.Contains(err.Error(), "rule 7") {
				t.Fatalf("over-cap artifact did not refuse rule7: %v", err)
			}
		})
	}
}

func TestReputationStarterEngineGuards(t *testing.T) {
	now := time.Date(2026, 9, 25, 12, 0, 0, 0, time.UTC)
	base := reputationContentBundle(t)
	for _, profile := range []string{"generated_cap", "resource_cap"} {
		t.Run(profile, func(t *testing.T) {
			bundle := base
			founder := reputationFounderState(t, bundle, 22, now, 6)
			founder.ReputationSpent, founder.ReputationUnlockPPM = 6, 50_000
			founder.ReputationNodesOwned = []string{"reputation.starter.cash_small", "reputation.starter.generated_beige_tower", "reputation.unlock.p05"}
			if profile == "resource_cap" {
				bundle = reputationStarterNextBundle(t, base, profile)
				founder.ReputationLevel, founder.ReputationSpent, founder.ReputationUnlockPPM = 2, 2, 0
				founder.ReputationNodesOwned = []string{"reputation.starter.cash_small"}
			}
			company, err := prestigecore.NewRunState(bundle.Economy, &save.State{RunSeq: 1}, founder, now)
			if err != nil {
				t.Fatal(err)
			}
			if profile == "generated_cap" {
				company.GeneratorProvisioned["generator.beige_tower"] = decimal.MaxExactInteger - 4
			} else {
				if _, err := company.Ledger.Apply(economy.Transaction{Entries: []economy.Entry{{ResourceID: "company.permits", Delta: decimal.One}}}); err != nil {
					t.Fatal(err)
				}
			}
			ids, err := ApplyReputationStarters(bundle, founder, company)
			if !errors.Is(err, ErrInvalidEngineState) || ids != nil {
				t.Fatalf("invalid engine input admitted/clamped: ids=%v err=%v", ids, err)
			}
			if profile == "generated_cap" {
				if company.GeneratorProvisioned["generator.beige_tower"] != decimal.MaxExactInteger-4 || !strings.Contains(err.Error(), "provisioned hardcap") {
					t.Fatal("failed generated credit mutated/clamped its target")
				}
			} else {
				balance, _ := company.Ledger.Balance("company.permits")
				if balance.String() != "1e0" {
					t.Fatalf("failed permit credit mutated/clamped target: %s", balance)
				}
			}
			if company.GeneratorCounts["generator.beige_tower"] != 0 || company.GeneratorPurchasedTotal != 0 {
				t.Fatal("free guard population became purchased")
			}
		})
	}
}

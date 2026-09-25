package production

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"cloud-clicker/server/economy"
	"cloud-clicker/server/multiplier"
	"cloud-clicker/server/reputation"
	"cloud-clicker/server/save"
)

func TestResolveFrozenContributionsAcceptsTheReputationProvider(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "balance", "catalogs", "phase0.json"))
	if err != nil {
		t.Fatal(err)
	}
	var root map[string]any
	if err := json.Unmarshal(data, &root); err != nil {
		t.Fatal(err)
	}
	root["multiplier_sources"] = append(root["multiplier_sources"].([]any),
		map[string]any{"id": "reputation.founder_bonus", "slot": "prestige", "target": "all", "provider": reputation.Provider},
		map[string]any{"id": "other.founder_bonus", "slot": "prestige", "target": "all", "provider": "other_provider"})
	data, _ = json.Marshal(root)
	catalog, err := economy.LoadCatalog(data)
	if err != nil {
		t.Fatal(err)
	}
	resolved, err := ResolveFrozenContributions(catalog, []save.FrozenContribution{
		{SourceID: "fiscal.hoard", Slot: multiplier.Slot("prestige"), Target: "all", Factor: "1e0"},
		{SourceID: "reputation.founder_bonus", Slot: multiplier.Slot("prestige"), Target: "all", Factor: "1.0005e0"},
	})
	if err != nil || len(resolved) != 2 || resolved[1].Factor.String() != "1.0005e0" {
		t.Fatalf("reputation contribution not resolved: %v %+v", err, resolved)
	}
	if _, err := ResolveFrozenContributions(catalog, []save.FrozenContribution{
		{SourceID: "other.founder_bonus", Slot: multiplier.Slot("prestige"), Target: "all", Factor: "1e0"},
	}); !errors.Is(err, ErrInvalidEngineState) {
		t.Fatalf("undeclared provider resolved: %v", err)
	}
}

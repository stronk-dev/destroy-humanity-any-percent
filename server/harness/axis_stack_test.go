package harness

import (
	"errors"
	"os"
	"strings"
	"testing"

	"cloud-clicker/server/economy"
	"cloud-clicker/server/save"
)

func axisFixtureCatalog(t *testing.T) *economy.Catalog {
	t.Helper()
	data, err := os.ReadFile("../../balance/testdata/axis-stack/economy-v5-fixture.json")
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := economy.LoadCatalog(data)
	if err != nil {
		t.Fatal(err)
	}
	return catalog
}

// AC9 failing case: a v5 economy with an axis stack must fail loudly rather
// than run with a silently neutral stack, in both harness runtimes.
func TestHarnessRefusesUnevaluatedAxisStack(t *testing.T) {
	fixture, err := os.ReadFile("../../balance/testdata/axis-stack/economy-v5-fixture.json")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := newSuite(Scenario{}, nil, fixture, nil, nil, ""); !errors.Is(err, ErrAxisStackUnevaluated) {
		t.Fatalf("Phase-0 suite accepted an axis stack: %v", err)
	}
	suite, err := LoadFirstHourSuite(repositoryRootForReputation, "balance/testdata/t0-t1/harness-scenario-v1.json", "balance/testdata/t0-t1/first-hour-policy-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	axisSuite := *suite
	axisSuite.Bundle.Economy = axisFixtureCatalog(t)
	spec := suite.Scenario.Runs[0]
	result := axisSuite.RunExperiment(spec, 1, FirstHourExperiment{AcquihirePurchasedMinimum: 200, BurnoutPriceFactor: "2e0", RouteKnowledgeBonus: 50, SeedCapital: "1e4", GeneratedBeigeTowers: 10})
	if result.Outcome == "completed" || !strings.Contains(strings.Join(result.InvariantFailures, " "), "axis_stack economy requires attainment") {
		t.Fatalf("first-hour runner simulated an axis stack neutrally: outcome=%s failures=%v", result.Outcome, result.InvariantFailures)
	}
	// Control: the pinned epoch-8 economy still runs.
	if control := suite.RunExperiment(spec, 1, FirstHourExperiment{AcquihirePurchasedMinimum: 200, BurnoutPriceFactor: "2e0", RouteKnowledgeBonus: 50, SeedCapital: "1e4", GeneratedBeigeTowers: 10}); control.Outcome != "completed" {
		t.Fatalf("control outcome=%s failures=%v", control.Outcome, control.InvariantFailures)
	}
}

func TestAxisInputWithinCapInvariant(t *testing.T) {
	catalog := axisFixtureCatalog(t)
	state := &save.State{AchievementsAttainedRun: map[string]bool{}, AttainmentScoreRun: 8}
	if err := CheckAxisInputWithinCap(catalog, state); err != nil {
		t.Fatal(err)
	}
	state.AttainmentScoreRun = 60 // saturates: x clamps to the visible cap
	if err := CheckAxisInputWithinCap(catalog, state); err != nil {
		t.Fatal(err)
	}
	state.AttainmentScoreRun = -1
	if err := CheckAxisInputWithinCap(catalog, state); err == nil {
		t.Fatal("negative attainment accepted")
	}
	// Inert without the Company v19 set or without a declared stack.
	if err := CheckAxisInputWithinCap(catalog, &save.State{AttainmentScoreRun: -1}); err != nil {
		t.Fatal(err)
	}
}

package harness

import (
	"errors"
	"slices"
	"testing"
)

func reputationCareerAdmissionInputs(t *testing.T) (*FirstHourSuite, RunSpec, FirstHourExperiment, ReputationCareerConfig) {
	t.Helper()
	suite, err := LoadFirstHourSuite(repositoryRootForReputation, "balance/testdata/t0-t1/harness-scenario-v1.json", "balance/testdata/t0-t1/first-hour-policy-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	for _, spec := range suite.Scenario.Runs {
		if spec.PolicyID == "chaos.t0_t1" {
			if spec.HorizonMS != 7_200_000 || spec.SeedStart != "0" {
				t.Fatal("career admission controls lost their declared population")
			}
			experiment := FirstHourExperiment{AcquihirePurchasedMinimum: 200, BurnoutPriceFactor: "2e0", RouteKnowledgeBonus: 50, SeedCapital: "1e4", GeneratedBeigeTowers: 10}
			return suite, spec, experiment, ReputationCareerConfig{Bundle: reputationCareerBundle(t, suite), Threshold: reputationCareerFixtureThreshold, Policy: CareerCheapest}
		}
	}
	t.Fatal("career admission controls have no Chaos population")
	return nil, RunSpec{}, FirstHourExperiment{}, ReputationCareerConfig{}
}

func TestReputationCareerRefusesInvalidInputs(t *testing.T) {
	suite, spec, experiment, config := reputationCareerAdmissionInputs(t)
	for _, threshold := range []string{reputationCareerFixtureThreshold, "1e12"} {
		for _, mutation := range []string{"empty-policy", "unknown-policy", "unknown-exclusion"} {
			t.Run(threshold+"/"+mutation, func(t *testing.T) {
				copy := config
				copy.Threshold = threshold
				switch mutation {
				case "empty-policy":
					copy.Policy = ""
				case "unknown-policy":
					copy.Policy = "cheepest"
				case "unknown-exclusion":
					copy.Exclude = "reputation.starter.typo"
				}
				if _, err := suite.RunReputationCareer(spec, 0, experiment, copy); !errors.Is(err, ErrReputationCareer) {
					t.Fatalf("invalid career input silently measured: %v", err)
				}
			})
		}
	}
}

func TestReputationCareerPreservesDeclaredInputs(t *testing.T) {
	suite, spec, experiment, config := reputationCareerAdmissionInputs(t)
	for _, input := range []struct {
		name    string
		policy  ReputationCareerPolicy
		exclude string
	}{
		{"cheapest", CareerCheapest, ""},
		{"seeded-uniform", CareerSeededUniform, ""},
		{"none", CareerNone, ""},
		{"known-exclusion", CareerCheapest, "reputation.starter.cash_small"},
	} {
		t.Run(input.name, func(t *testing.T) {
			copy := config
			copy.Policy, copy.Exclude = input.policy, input.exclude
			result, err := suite.RunReputationCareer(spec, 0, experiment, copy)
			if err != nil || result.Run.Outcome != "completed" || len(result.Run.ReputationExits) != 2 || result.RunThreeGateMS == nil {
				t.Fatalf("declared career did not complete: err=%v outcome=%s exits=%d gate=%v", err, result.Run.Outcome, len(result.Run.ReputationExits), result.RunThreeGateMS)
			}
			if input.policy == CareerNone {
				if len(result.PurchasedNodeIDs) != 0 || len(result.AppliedStarterIDs) != 0 || result.BonusFactor != "1e0" {
					t.Fatal("no-purchase control acquired a Reputation effect")
				}
			} else if len(result.PurchasedNodeIDs) == 0 {
				t.Fatal("legal purchasing control bought no nodes")
			}
			if input.exclude != "" && (slices.Contains(result.PurchasedNodeIDs, input.exclude) || slices.Contains(result.AppliedStarterIDs, input.exclude)) {
				t.Fatal("known leave-one-out mask was ignored")
			}
		})
	}
}

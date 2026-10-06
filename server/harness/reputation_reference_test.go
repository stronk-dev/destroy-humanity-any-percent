package harness

import (
	"bytes"
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"cloud-clicker/server/decimal"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/multiplier"
	prestigecore "cloud-clicker/server/prestige"
	"cloud-clicker/server/production"
	"cloud-clicker/server/save"
)

var reputationReferenceObservation = flag.String("reputation-reference-observe", "", "explicit Reference seed0 career observation: baseline or current; never updates reports")

// This is an isolated legal run3 fixture, not a naturally earned/SQL career.
// Every asset and the non-unit factor come from the public tree/assembly paths.
func reputationReferenceFixture(t *testing.T, arm string) *firstHourRuntime {
	t.Helper()
	suite, _, _, config := reputationCareerAdmissionInputs(t)
	suite.Bundle, suite.ConstantsHash = config.Bundle, config.Bundle.ConstantsHash
	tree := config.Bundle.ReputationTree
	founder := &save.State{ReputationLevel: 6, ReputationNodesOwned: []string{}}
	for _, id := range []string{"reputation.unlock.p05", "reputation.starter.cash_small", "reputation.starter.generated_beige_tower"} {
		purchase, rejection, err := tree.Purchase(founder.ReputationLevel, founder.ReputationSpent, founder.ReputationNodesOwned, id)
		if err != nil || rejection != nil {
			t.Fatalf("fixture tree purchase %s: %v rejection=%v", id, err, rejection)
		}
		founder.ReputationSpent, founder.ReputationNodesOwned, founder.ReputationUnlockPPM = purchase.SpentAfter, purchase.OwnedAfter, purchase.UnlockPPMAfter
	}
	prior, err := newFirstHourCompany(config.Bundle.Economy)
	if err != nil {
		t.Fatal(err)
	}
	prior.RunSeq = 2
	company, err := prestigecore.NewRunState(config.Bundle.Economy, prior, founder, Epoch)
	if err != nil {
		t.Fatal(err)
	}
	starters, err := production.ApplyReputationStarters(config.Bundle, founder, company)
	cash, _ := company.Ledger.Balance("company.cash")
	if err != nil || len(starters) != 2 || company.RunSeq != 3 || cash.String() != "1e3" ||
		company.GeneratorProvisioned["generator.beige_tower"] != 5 || company.GeneratorCounts["generator.beige_tower"] != 0 {
		t.Fatalf("fixture lost its assembled run3 assets: err=%v starters=%v cash=%s", err, starters, cash)
	}
	factor, err := tree.BonusFactor(founder.ReputationLevel, founder.ReputationSpent, founder.ReputationUnlockPPM)
	if err != nil || factor.String() != "1.003e0" {
		t.Fatalf("fixture has no legal non-unit bonus: %s %v", factor, err)
	}
	var external []multiplier.Contribution
	if arm != "none" {
		if arm == "unit" {
			factor = decimal.One // Explicit counterfactual; do not alter the tree.
		} else if arm != "tree" {
			t.Fatalf("unknown diagnostic arm %q", arm)
		}
		external, err = production.ResolveFrozenContributions(config.Bundle.Economy, []save.FrozenContribution{{
			SourceID: tree.Bonus.SourceID, Slot: tree.Bonus.Slot, Target: tree.Bonus.Target, Factor: factor.String(),
		}})
		if err != nil {
			t.Fatal(err)
		}
	}
	for _, spec := range suite.Scenario.Runs {
		if spec.PolicyID == "reference.greedy" {
			policy, ok := suite.Policy.Policy(spec.PolicyID, spec.PolicyVersion)
			if !ok || policy.ActionCadenceMS != 1000 {
				t.Fatal("fixture lost its ratified Reference policy")
			}
			return &firstHourRuntime{suite: suite, spec: spec, policy: policy, company: company, founder: founder,
				revision: 1, milestones: map[string]*int64{}, career: &careerRuntime{bonus: external}}
		}
	}
	t.Fatal("fixture has no Reference population")
	return nil
}

func reputationReferenceCashRate(t *testing.T, runtime *firstHourRuntime, state *save.State) decimal.Decimal {
	t.Helper()
	// The first-hour instrument uses legacy v14 states and does not simulate
	// active play. Use that same contract, not a modern active-play projection.
	if save.VersionForState(state) != save.CurrentVersion {
		t.Fatal("diagnostic no longer uses the first-hour harness state contract")
	}
	projection, err := production.ProjectRates(production.CatalogBundle{Economy: runtime.suite.Bundle.Economy}, state, runtime.external(), 0)
	if err != nil {
		t.Fatal(err)
	}
	for _, resource := range projection.Resources {
		if resource.ResourceID == "company.cash" {
			return resource.Rate
		}
	}
	t.Fatal("canonical projection lost company.cash")
	return decimal.NaN
}

func requireReferenceCompanyEqual(t *testing.T, actual, expected *save.State) {
	t.Helper()
	got, gotErr := save.EncodeState(actual)
	want, wantErr := save.EncodeState(expected)
	if gotErr != nil || wantErr != nil {
		t.Fatalf("diagnostic states cannot encode: got=%v want=%v", gotErr, wantErr)
	}
	if !bytes.Equal(got, want) {
		gotCash, _ := actual.Ledger.Balance("company.cash")
		wantCash, _ := expected.Ledger.Balance("company.cash")
		t.Fatalf("Reference diverges from the same frozen production input: cash=%s want=%s", gotCash, wantCash)
	}
}

func TestReputationReferenceFrozenInputDiagnostic(t *testing.T) {
	for _, arm := range []string{"none", "unit", "tree"} {
		t.Run(arm, func(t *testing.T) {
			newFixture := func(t *testing.T) (*firstHourRuntime, *RelevanceSuite) {
				runtime := reputationReferenceFixture(t, arm)
				ranker, err := runtime.referenceRanker()
				if err != nil {
					t.Fatal(err)
				}
				want := "5e0"
				if arm == "tree" {
					want = "5.015e0"
				}
				if rate := reputationReferenceCashRate(t, runtime, runtime.company); rate.String() != want {
					t.Fatalf("independent canonical rate lost the declared fixture: %s want %s", rate, want)
				}
				return runtime, ranker
			}
			t.Run("projection", func(t *testing.T) {
				runtime, ranker := newFixture(t)
				got, reachable, err := ranker.projectedMilestone(runtime.company, production.AblationMask{})
				want := reputationReferenceCashRate(t, runtime, runtime.company)
				if err != nil || !reachable || !got.Denominator.Eq(want) {
					t.Fatalf("Reference rate drops frozen input: got=%s want=%s reachable=%v err=%v", got.Denominator, want, reachable, err)
				}
			})
			t.Run("advance", func(t *testing.T) {
				runtime, ranker := newFixture(t)
				want, err := cloneState(ranker.Catalog, runtime.company)
				if err != nil {
					t.Fatal(err)
				}
				if _, err := production.SimulateAdvance(want, ranker.Catalog, production.SimulationDependencies{Routes: ranker.Routes},
					runtime.companyRevision(), production.ModeOnline, relevanceNow(1000), runtime.external(), production.AblationMask{}); err != nil {
					t.Fatal(err)
				}
				if _, err := ranker.advance(runtime.company, 1, 1000, production.AblationMask{}, &relevanceCounter{limit: runtime.suite.Scenario.TransitionBudget}); err != nil {
					t.Fatal(err)
				}
				requireReferenceCompanyEqual(t, runtime.company, want)
			})
			t.Run("candidate-purchase", func(t *testing.T) {
				runtime, ranker := newFixture(t)
				want, err := cloneState(ranker.Catalog, runtime.company)
				if err != nil {
					t.Fatal(err)
				}
				request := production.IntentRequest{IntentID: relevanceIntentID(1), Kind: production.IntentBuyGenerator,
					ExpectedRevision: 1, GeneratorID: "generator.beige_tower", CountMode: "exact", Count: 1}
				transition, err := production.SimulateTransition(request, want, ranker.Catalog, production.SimulationDependencies{Routes: ranker.Routes},
					runtime.companyRevision(), production.ModeOnline, relevanceNow(1000), runtime.external(), nil, production.AblationMask{})
				if err != nil || transition.Decision.Outcome != save.IntentApplied {
					t.Fatalf("independent candidate purchase: %v", err)
				}
				got, reachable, err := ranker.rankCandidate(runtime.company, 1, 1000, 2000, request.GeneratorID,
					production.AblationMask{}, &relevanceCounter{limit: runtime.suite.Scenario.TransitionBudget})
				if err != nil || !reachable || got.AtMS != 1000 || got.Revision != 2 {
					t.Fatalf("Reference candidate lost its actual purchase: reachable=%v at=%d revision=%d err=%v", reachable, got.AtMS, got.Revision, err)
				}
				requireReferenceCompanyEqual(t, got.State, want)
				if rate := reputationReferenceCashRate(t, runtime, want); !got.Projection.Denominator.Eq(rate) {
					t.Fatalf("candidate projection drops frozen input: got=%s want=%s", got.Projection.Denominator, rate)
				}
			})
			t.Run("actual-bank-branch", func(t *testing.T) {
				runtime, ranker := newFixture(t)
				// Explicit zero-cash counterfactual isolates bank dispatch without
				// changing the retained starter, tree, policy or scenario bytes.
				var err error
				runtime.company.Ledger, err = economy.NewLedger(ranker.Catalog, economy.ScopeCompany)
				if err != nil {
					t.Fatal(err)
				}
				candidates, bank, at, err := ranker.rankDecisionOptions(runtime.company, 1, 0, 1000, 1,
					production.AblationMask{}, &relevanceCounter{limit: runtime.suite.Scenario.TransitionBudget})
				if err != nil || len(candidates) != 0 || !bank || at != 1000 {
					t.Fatalf("fixture does not isolate actual bank dispatch: candidates=%d bank=%v at=%d err=%v", len(candidates), bank, at, err)
				}
				want, err := cloneState(ranker.Catalog, runtime.company)
				if err != nil {
					t.Fatal(err)
				}
				if _, err := production.SimulateAdvance(want, ranker.Catalog, production.SimulationDependencies{Routes: ranker.Routes},
					runtime.companyRevision(), production.ModeOnline, relevanceNow(at), runtime.external(), production.AblationMask{}); err != nil {
					t.Fatal(err)
				}
				if err := runtime.applyReferenceChoice(0); err != nil {
					t.Fatal(err)
				}
				if runtime.company.EvaluatedThrough != relevanceNow(at) || runtime.revision != 1 {
					t.Fatal("actual bank branch did not advance without an intent revision")
				}
				requireReferenceCompanyEqual(t, runtime.company, want)
			})
			t.Run("effect-mask", func(t *testing.T) {
				runtime, ranker := newFixture(t)
				_, reachable, err := ranker.projectedMilestone(runtime.company, production.AblationMask{GeneratorIDs: []string{"generator.beige_tower"}})
				if err != nil || reachable {
					t.Fatalf("effect mask did not remove the only producer: reachable=%v err=%v", reachable, err)
				}
			})
		})
	}
	t.Run("zero-production", func(t *testing.T) {
		runtime := reputationReferenceFixture(t, "tree")
		runtime.company.GeneratorProvisioned["generator.beige_tower"] = 0
		ranker, err := runtime.referenceRanker()
		if err != nil {
			t.Fatal(err)
		}
		_, reachable, err := ranker.projectedMilestone(runtime.company, production.AblationMask{})
		if err != nil || reachable || !reputationReferenceCashRate(t, runtime, runtime.company).Eq(decimal.Zero) {
			t.Fatalf("frozen bonus invented production: reachable=%v err=%v", reachable, err)
		}
	})
}

// This opt-in observation is not H4/H5 acceptance. Baseline mode requires the
// retained Reference row; current mode records any drift without authoring it.
func TestReputationReferenceCareerObservation(t *testing.T) {
	mode := *reputationReferenceObservation
	if mode == "" {
		t.Skip("explicit observation: root make test-go with -args -reputation-reference-observe=baseline|current")
	}
	if mode != "baseline" && mode != "current" {
		t.Fatal("unknown Reference observation mode")
	}
	suite, _, experiment, config := reputationCareerAdmissionInputs(t)
	var reference RunSpec
	for _, spec := range suite.Scenario.Runs {
		if spec.PolicyID == "reference.greedy" {
			reference = spec
		}
	}
	if reference.PolicyID == "" || reference.SeedStart != "0" || reference.SeedCount != 1 || reference.HorizonMS != 7_200_000 {
		t.Fatal("observation lost its ratified Reference population")
	}
	config.Policy = CareerCheapest
	treated, treatedErr := suite.RunReputationCareer(reference, 0, experiment, config)
	config.Policy = CareerNone
	control, controlErr := suite.RunReputationCareer(reference, 0, experiment, config)
	observation := struct {
		Mode          string                 `json:"mode"`
		ConstantsHash string                 `json:"fixture_constants_hash"`
		Threshold     string                 `json:"fixture_threshold"`
		Experiment    FirstHourExperiment    `json:"experiment"`
		Spec          RunSpec                `json:"run_spec"`
		Treated       ReputationCareerResult `json:"cheapest"`
		Control       ReputationCareerResult `json:"none"`
	}{mode, config.Bundle.ConstantsHash, config.Threshold, experiment, reference, treated, control}
	encoded, err := json.Marshal(observation)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Reference career observation: %s", encoded)
	if treatedErr != nil || controlErr != nil || treated.Run.Outcome != "completed" || control.Run.Outcome != "completed" ||
		len(treated.Run.ReputationExits) != 2 || len(control.Run.ReputationExits) != 2 || treated.RunThreeGateMS == nil || control.RunThreeGateMS == nil {
		t.Fatalf("Reference observation did not complete its declared objective: treated=%v control=%v", treatedErr, controlErr)
	}
	data, err := os.ReadFile(filepath.Join(repositoryRootForReputation, reputationCareerReportPath))
	var retained reputationCareerReport
	if err != nil || json.Unmarshal(data, &retained) != nil || retained.Threshold != config.Threshold {
		t.Fatalf("retained Reference comparison is unavailable: %v", err)
	}
	var original *reputationCareerSeed
	for index := range retained.Seeds {
		row := &retained.Seeds[index]
		if row.PolicyID == reference.PolicyID && row.Seed == 0 {
			if original != nil {
				t.Fatal("duplicate retained Reference row")
			}
			original = row
		}
	}
	if original == nil || original.TreatedGateMS == nil || original.ControlGateMS == nil {
		t.Fatal("retained comparison lost the Reference pair")
	}
	same := *original.TreatedGateMS == *treated.RunThreeGateMS && *original.ControlGateMS == *control.RunThreeGateMS &&
		original.BonusFactor == treated.BonusFactor && reflect.DeepEqual(original.PurchasedNodeIDs, treated.PurchasedNodeIDs) &&
		reflect.DeepEqual(original.AppliedStarterIDs, treated.AppliedStarterIDs)
	t.Logf("Reference v1 row reproduced=%v; gate treated=%d (v1=%d), control=%d (v1=%d)", same,
		*treated.RunThreeGateMS, *original.TreatedGateMS, *control.RunThreeGateMS, *original.ControlGateMS)
	if mode == "baseline" && !same {
		t.Fatal("pre-correction Reference observation does not reproduce its retained row")
	}
}

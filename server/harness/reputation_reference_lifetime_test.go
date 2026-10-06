package harness

import (
	"bytes"
	"testing"

	"cloud-clicker/server/decimal"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/production"
	"cloud-clicker/server/save"
)

func referenceLifetimePurchase(t *testing.T, runtime *firstHourRuntime, id string) production.IntentRequest {
	t.Helper()
	request := production.IntentRequest{IntentID: relevanceIntentID(runtime.revision), ExpectedRevision: runtime.revision}
	if _, ok := runtime.suite.Bundle.Economy.GeneratorClass(id); ok {
		request.Kind, request.GeneratorID, request.CountMode, request.Count = production.IntentBuyGenerator, id, "exact", 1
	} else if _, ok := runtime.suite.Bundle.Economy.Upgrade(id); ok {
		request.Kind, request.UpgradeID = production.IntentBuyUpgrade, id
	} else {
		t.Fatalf("Reference selected an undeclared candidate %q", id)
	}
	return request
}

func requireReferenceLifetimeState(t *testing.T, actual, expected *save.State) {
	t.Helper()
	got, gotErr := save.EncodeState(actual)
	want, wantErr := save.EncodeState(expected)
	if gotErr != nil || wantErr != nil {
		t.Fatalf("lifetime states cannot encode: got=%v want=%v", gotErr, wantErr)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("Reference omitted canonical lifetime accrual: got=%s want=%s", actual.LifetimeValue, expected.LifetimeValue)
	}
}

func TestReputationReferenceLifetimeAccounting(t *testing.T) {
	for _, arm := range []string{"none", "unit", "tree"} {
		t.Run(arm, func(t *testing.T) {
			fixture := func(t *testing.T) (*firstHourRuntime, *RelevanceSuite, decimal.Decimal) {
				t.Helper()
				runtime := reputationReferenceFixture(t, arm)
				if !runtime.company.LifetimeValue.Eq(decimal.Zero) {
					t.Fatal("starter assets must not be fabricated as lifetime production")
				}
				ranker, err := runtime.referenceRanker()
				if err != nil {
					t.Fatal(err)
				}
				want := decimal.FromFloat64(5)
				if arm == "tree" {
					want = decimal.FromFloat64(5.015)
				}
				return runtime, ranker, want
			}
			t.Run("actual-candidate-adoption", func(t *testing.T) {
				runtime, ranker, value := fixture(t)
				candidates, _, _, err := ranker.rankDecisionOptions(runtime.company, runtime.revision, 1000, 2000, 1,
					production.AblationMask{}, &relevanceCounter{limit: runtime.suite.Scenario.TransitionBudget})
				if err != nil || len(candidates) == 0 || candidates[0].AtMS != 1000 {
					t.Fatalf("fixture does not isolate one-second candidate adoption: count=%d err=%v", len(candidates), err)
				}
				want, err := cloneState(ranker.Catalog, runtime.company)
				if err != nil {
					t.Fatal(err)
				}
				result, err := production.SimulateTransition(referenceLifetimePurchase(t, runtime, candidates[0].ID), want, ranker.Catalog,
					production.SimulationDependencies{Routes: ranker.Routes, Hook: runtime.lifetimeHook()}, runtime.companyRevision(),
					production.ModeOnline, relevanceNow(1000), runtime.external(), nil, production.AblationMask{})
				if err != nil || result.Decision.Outcome != save.IntentApplied || !want.LifetimeValue.Eq(value) {
					t.Fatalf("canonical purchase lost paid production: lifetime=%s want=%s err=%v", want.LifetimeValue, value, err)
				}
				if err := runtime.applyReferenceChoice(1000); err != nil {
					t.Fatal(err)
				}
				if runtime.revision != 2 || runtime.company.EvaluatedThrough != relevanceNow(1000) {
					t.Fatal("Reference did not actually adopt the candidate")
				}
				requireReferenceLifetimeState(t, runtime.company, want)
			})
			t.Run("actual-bank-dispatch", func(t *testing.T) {
				runtime, ranker, value := fixture(t)
				var err error
				runtime.company.Ledger, err = economy.NewLedger(ranker.Catalog, economy.ScopeCompany)
				if err != nil {
					t.Fatal(err)
				}
				candidates, bank, at, err := ranker.rankDecisionOptions(runtime.company, runtime.revision, 0, 1000, 1,
					production.AblationMask{}, &relevanceCounter{limit: runtime.suite.Scenario.TransitionBudget})
				if err != nil || len(candidates) != 0 || !bank || at != 1000 {
					t.Fatalf("fixture does not isolate bank dispatch: count=%d bank=%v at=%d err=%v", len(candidates), bank, at, err)
				}
				want, err := cloneState(ranker.Catalog, runtime.company)
				if err != nil {
					t.Fatal(err)
				}
				if _, err := production.SimulateAdvance(want, ranker.Catalog,
					production.SimulationDependencies{Routes: ranker.Routes, Hook: runtime.lifetimeHook()}, runtime.companyRevision(),
					production.ModeOnline, relevanceNow(at), runtime.external(), production.AblationMask{}); err != nil || !want.LifetimeValue.Eq(value) {
					t.Fatalf("canonical bank lost paid production: lifetime=%s want=%s err=%v", want.LifetimeValue, value, err)
				}
				if err := runtime.applyReferenceChoice(0); err != nil {
					t.Fatal(err)
				}
				if runtime.revision != 1 || runtime.company.EvaluatedThrough != relevanceNow(at) {
					t.Fatal("Reference did not actually bank without an intent revision")
				}
				requireReferenceLifetimeState(t, runtime.company, want)
			})
			t.Run("zero-elapsed-candidate", func(t *testing.T) {
				runtime, ranker, _ := fixture(t)
				request := referenceLifetimePurchase(t, runtime, "generator.beige_tower")
				want, err := cloneState(ranker.Catalog, runtime.company)
				if err != nil {
					t.Fatal(err)
				}
				result, err := production.SimulateTransition(request, want, ranker.Catalog,
					production.SimulationDependencies{Routes: ranker.Routes, Hook: runtime.lifetimeHook()}, runtime.companyRevision(),
					production.ModeOnline, relevanceNow(0), runtime.external(), nil, production.AblationMask{})
				if err != nil || result.Decision.Outcome != save.IntentApplied || !want.LifetimeValue.Eq(decimal.Zero) {
					t.Fatalf("zero elapsed invented lifetime production: %s %v", want.LifetimeValue, err)
				}
				got, reachable, err := ranker.rankCandidate(runtime.company, runtime.revision, 0, 1000, request.GeneratorID,
					production.AblationMask{}, &relevanceCounter{limit: runtime.suite.Scenario.TransitionBudget})
				if err != nil || !reachable || got.AtMS != 0 || got.Revision != 2 {
					t.Fatalf("zero-elapsed candidate setup failed: reachable=%v err=%v", reachable, err)
				}
				requireReferenceLifetimeState(t, got.State, want)
			})
			t.Run("ordinary-intent-control", func(t *testing.T) {
				runtime, ranker, value := fixture(t)
				request := referenceLifetimePurchase(t, runtime, "generator.beige_tower")
				want, err := cloneState(ranker.Catalog, runtime.company)
				if err != nil {
					t.Fatal(err)
				}
				result, err := production.SimulateTransition(request, want, ranker.Catalog,
					production.SimulationDependencies{Routes: ranker.Routes, Hook: runtime.lifetimeHook()}, runtime.companyRevision(),
					production.ModeOnline, relevanceNow(1000), runtime.external(), nil, production.AblationMask{})
				if err != nil || result.Decision.Outcome != save.IntentApplied || !want.LifetimeValue.Eq(value) {
					t.Fatalf("ordinary-intent oracle lost production: %s %v", want.LifetimeValue, err)
				}
				if err := runtime.apply(request, relevanceNow(1000), 1000, production.ModeOnline); err != nil {
					t.Fatal(err)
				}
				requireReferenceLifetimeState(t, runtime.company, want)
			})
			for _, masked := range []bool{false, true} {
				name := "ranker-advance"
				if masked {
					name = "effect-masked-advance"
				}
				t.Run(name, func(t *testing.T) {
					runtime, ranker, value := fixture(t)
					mask := production.AblationMask{}
					if masked {
						mask.GeneratorIDs = []string{"generator.beige_tower"}
						value = decimal.Zero
					}
					want, err := cloneState(ranker.Catalog, runtime.company)
					if err != nil {
						t.Fatal(err)
					}
					if _, err := production.SimulateAdvance(want, ranker.Catalog,
						production.SimulationDependencies{Routes: ranker.Routes, Hook: runtime.lifetimeHook()}, runtime.companyRevision(),
						production.ModeOnline, relevanceNow(1000), runtime.external(), mask); err != nil || !want.LifetimeValue.Eq(value) {
						t.Fatalf("canonical advance lost masked accounting: lifetime=%s want=%s err=%v", want.LifetimeValue, value, err)
					}
					if _, err := ranker.advance(runtime.company, runtime.revision, 1000, mask,
						&relevanceCounter{limit: runtime.suite.Scenario.TransitionBudget}); err != nil {
						t.Fatal(err)
					}
					requireReferenceLifetimeState(t, runtime.company, want)
				})
			}
		})
	}
}

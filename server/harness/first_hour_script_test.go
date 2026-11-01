package harness

import (
	"bytes"
	"reflect"
	"testing"
	"time"

	"cloud-clicker/server/decimal"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/production"
	"cloud-clicker/server/save"
)

func TestFirstHourReferenceScriptCapturesOnlySelectedPurchase(t *testing.T) {
	for _, arm := range []string{"none", "unit", "tree"} {
		for _, timing := range []string{"immediate", "delayed"} {
			t.Run(arm+"/"+timing, func(t *testing.T) {
				runtime := reputationReferenceFixture(t, arm)
				price, err := runtime.suite.Bundle.Economy.BulkCost("generator.beige_tower", 0, 1)
				if err != nil {
					t.Fatal(err)
				}
				if timing == "delayed" {
					price = price.Sub(decimal.FromFloat64(2.5))
				}
				balances := runtime.company.Ledger.Snapshot()
				balances["company.cash"] = price.String()
				runtime.company.Ledger, err = economy.RestoreLedger(runtime.suite.Bundle.Economy, economy.ScopeCompany, balances)
				if err != nil {
					t.Fatal(err)
				}
				origin, err := save.EncodeState(runtime.company)
				if err != nil {
					t.Fatal(err)
				}
				revision := runtime.companyRevision()
				ranker, err := runtime.referenceRanker()
				if err != nil {
					t.Fatal(err)
				}
				counter := &relevanceCounter{limit: runtime.suite.Scenario.TransitionBudget}
				candidates, _, _, err := ranker.rankDecisionOptions(runtime.company, runtime.revision, 0, runtime.policy.ActionCadenceMS, 1, production.AblationMask{}, counter)
				if err != nil || len(candidates) == 0 {
					t.Fatal("fixture has no real selected purchase", err)
				}
				selected := candidates[0]
				if timing == "delayed" && selected.AtMS <= 0 || timing == "immediate" && selected.AtMS != 0 {
					t.Fatal("fixture lost its timing arm", selected.AtMS)
				}
				commands := []FirstHourScriptCommand{}
				runtime.commands = &commands
				if err := runtime.applyReferenceChoice(0); err != nil {
					t.Fatal(err)
				}
				want := FirstHourScriptCommand{AtMS: selected.AtMS, Mode: production.ModeOnline, Request: candidateIntent(selected.ID, revision.Number)}
				if len(commands) != 1 || !reflect.DeepEqual(commands[0], want) {
					t.Fatalf("selected purchase missing, duplicated or changed: got%+v want%+v", commands, want)
				}
				if runtime.revision != selected.Revision || runtime.transitions != counter.value {
					t.Fatal("capture changed actual choice revision/work")
				}
				requireReferenceCompanyEqual(t, runtime.company, selected.State)
				// Replay exactly the current ranker's existing wait-then-buy path.
				// This does NOT adopt the draft original-interval accounting path.
				replayed, err := save.RestoreState(origin, save.VersionForState(runtime.company), runtime.suite.Bundle.Economy, economy.ScopeCompany, time.Time{})
				if err != nil {
					t.Fatal(err)
				}
				if selected.AtMS > 0 {
					if _, err := production.SimulateAdvance(replayed, ranker.Catalog, production.SimulationDependencies{Routes: ranker.Routes, Hook: runtime.lifetimeHook()}, revision, production.ModeOnline, relevanceNow(selected.AtMS), runtime.external(), production.AblationMask{}); err != nil {
						t.Fatal(err)
					}
				}
				result, err := production.SimulateTransition(commands[0].Request, replayed, ranker.Catalog, production.SimulationDependencies{Routes: ranker.Routes, Hook: runtime.lifetimeHook()}, revision, commands[0].Mode, relevanceNow(commands[0].AtMS), runtime.external(), nil, production.AblationMask{})
				if err != nil || result.Decision.Outcome != save.IntentApplied {
					t.Fatal("captured real command cannot apply", err)
				}
				requireReferenceCompanyEqual(t, runtime.company, replayed)
			})
		}
	}
}

func TestFirstHourReferenceScriptDoesNotRecordBankAsAction(t *testing.T) {
	runtime := reputationReferenceFixture(t, "none")
	var err error
	runtime.company.Ledger, err = economy.NewLedger(runtime.suite.Bundle.Economy, economy.ScopeCompany)
	if err != nil {
		t.Fatal(err)
	}
	for id := range runtime.company.GeneratorProvisioned {
		runtime.company.GeneratorProvisioned[id] = 0
	}
	runtime.company.GeneratorProvisioned["generator.beige_tower"] = 1
	commands := []FirstHourScriptCommand{}
	runtime.commands = &commands
	if err := runtime.applyReferenceChoice(0); err != nil {
		t.Fatal(err)
	}
	if len(commands) != 0 || runtime.revision != 1 || !runtime.company.EvaluatedThrough.Equal(relevanceNow(runtime.policy.ActionCadenceMS)) {
		t.Fatal("bank was omitted or falsely captured as a player action", commands, runtime.revision)
	}
}

func TestFirstHourScriptCoversActualRevisionChanges(t *testing.T) {
	suite, err := LoadFirstHourSuite(repositoryRootForReputation, "balance/testdata/t0-t1/harness-scenario-v1.json", "balance/testdata/t0-t1/first-hour-policy-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	experiment := FirstHourExperiment{AcquihirePurchasedMinimum: 200, BurnoutPriceFactor: "2e0", RouteKnowledgeBonus: 50, SeedCapital: "1e4", GeneratedBeigeTowers: 10}
	for _, spec := range suite.Scenario.Runs {
		t.Run(spec.PolicyID, func(t *testing.T) {
			withCapture, commands := suite.RunExperimentScript(spec, 0, experiment)
			withoutCapture := suite.RunExperiment(spec, 0, experiment)
			if withCapture.Outcome != "completed" || !reflect.DeepEqual(withCapture, withoutCapture) {
				t.Fatal("capture changes complete measured result", withCapture.InvariantFailures, withoutCapture.InvariantFailures)
			}
			policy, ok := suite.Policy.Policy(spec.PolicyID, spec.PolicyVersion)
			if !ok {
				t.Fatal("policy absent")
			}
			company, err := newFirstHourCompany(suite.Bundle.Economy)
			if err != nil {
				t.Fatal(err)
			}
			observed := []FirstHourScriptCommand{}
			runtime := &firstHourRuntime{suite: suite, spec: spec, policy: policy, seed: 0, experiment: experiment, company: company,
				founder: &save.State{LedgerFactKinds: map[string]bool{}, NetworkSlots: []save.NetworkSlot{}, ExitHistory: []save.ExitRecord{}}, revision: 1, lastSession: -1, milestones: map[string]*int64{}, commands: &observed}
			boundaries, err := firstHourBoundaries(policy, 0, spec.HorizonMS)
			if err != nil {
				t.Fatal(err)
			}
			for _, boundary := range boundaries {
				if runtime.transitions >= suite.Scenario.TransitionBudget {
					t.Fatal("unchanged scenario budget exceeded")
				}
				beforeRevision, beforeCommands := runtime.revision, len(observed)
				if err := runtime.step(boundary); err != nil {
					t.Fatal(err)
				}
				terminal := runtime.milestones["milestone.first_elective_exit"] != nil
				wantAdded := runtime.revision - beforeRevision
				if terminal {
					wantAdded++
				} // Terminal measurement records Wind Down, not a new simulated run.
				if int64(len(observed)-beforeCommands) != wantAdded {
					t.Fatalf("actual action/restart missing or duplicated in script at%d: revision_delta%d commands_added%d", boundary.atMS, runtime.revision-beforeRevision, len(observed)-beforeCommands)
				}
				if len(observed) > beforeCommands {
					command := observed[beforeCommands]
					if command.AtMS < boundary.atMS || command.AtMS > spec.HorizonMS || beforeCommands > 0 && command.AtMS < observed[beforeCommands-1].AtMS {
						t.Fatal("script escaped actual choice chronology")
					}
				}
				if terminal {
					break
				}
			}
			if runtime.milestones["milestone.first_elective_exit"] == nil || !reflect.DeepEqual(commands, observed) {
				t.Fatal("public script omits actual observed action stream")
			}
			generators, upgrades := 0, 0
			for _, command := range commands {
				if command.Request.Kind == production.IntentBuyGenerator {
					generators++
				}
				if command.Request.Kind == production.IntentBuyUpgrade {
					upgrades++
				}
			}
			if spec.PolicyID == "reference.greedy" && (generators == 0 || upgrades == 0) {
				t.Fatal("complete reference capture did not include both actual purchase families", generators, upgrades)
			}
			if withCapture.TransitionCount != runtime.transitions || !reflect.DeepEqual(withCapture.Ending, runtime.ending) || !reflect.DeepEqual(withCapture.ReputationExits, runtime.reputationExits) {
				t.Fatal("script observation changed full runner outputs")
			}
			t.Logf("policy%s seed0 commands%d generators%d upgrades%d transitions%d capture-off full result identical", spec.PolicyID, len(commands), generators, upgrades, runtime.transitions)
		})
	}
}

func TestFirstHourReferenceScriptCaptureIsDetached(t *testing.T) {
	runtime := reputationReferenceFixture(t, "tree")
	commands := []FirstHourScriptCommand{}
	runtime.commands = &commands
	if err := runtime.applyReferenceChoice(0); err != nil {
		t.Fatal(err)
	}
	if len(commands) != 1 {
		t.Fatal("selected command absent")
	}
	before, err := save.EncodeState(runtime.company)
	if err != nil {
		t.Fatal(err)
	}
	commands[0].Request.GeneratorID = "generator.not_owned"
	commands[0].Request.UpgradeID = "upgrade.not_owned"
	commands[0].AtMS++
	after, err := save.EncodeState(runtime.company)
	if err != nil || !bytes.Equal(before, after) {
		t.Fatal("editing a captured command changed adopted Company", err)
	}
}

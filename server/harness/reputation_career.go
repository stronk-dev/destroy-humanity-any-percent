package harness

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"cloud-clicker/server/multiplier"
	prestigecore "cloud-clicker/server/prestige"
	"cloud-clicker/server/production"
	"cloud-clicker/server/save"
)

// Reputation Tree v1 H4: a runs 1–3 career on the served transitions. The
// first-hour runner continues past the first elective Exit: at that Exit the
// declared purchase policy buys tree nodes, run 3 is assembled with the
// served starter application, and run 3 production carries the frozen
// Founder bonus. The measurement is fixture-first: it runs under a fixture
// threshold from the OD-2 measurement, never a minted value.

// ReputationCareerPolicy is declared data (R10 H4): "cheapest" buys the
// cheapest available node repeatedly (ties by tree order); "seeded_uniform"
// draws uniformly among available nodes; "none" is the control.
type ReputationCareerPolicy string

const (
	CareerCheapest      ReputationCareerPolicy = "cheapest"
	CareerSeededUniform ReputationCareerPolicy = "seeded_uniform"
	CareerNone          ReputationCareerPolicy = "none"
)

type ReputationCareerConfig struct {
	Bundle    production.CatalogBundle
	Threshold string
	Policy    ReputationCareerPolicy
}

type ReputationCareerResult struct {
	Run                 FirstHourRunResult `json:"run"`
	PurchasedNodeIDs    []string           `json:"purchased_node_ids"`
	AppliedStarterIDs   []string           `json:"applied_starter_node_ids"`
	BonusFactor         string             `json:"bonus_factor"`
	RunThreeGateMS      *int64             `json:"run_three_gate_ms"`
	ReputationAvailable int64              `json:"reputation_available_after_purchases"`
}

type careerRuntime struct {
	config    ReputationCareerConfig
	policy    *prestigecore.Policy
	purchased []string
	applied   []string
	bonus     []multiplier.Contribution
	factor    string
	gateMS    *int64
}

func (career *careerRuntime) done() bool { return career.gateMS != nil }

var ErrReputationCareer = errors.New("invalid reputation career")

// RunReputationCareer runs one seed as a runs 1–3 career.
func (suite *FirstHourSuite) RunReputationCareer(spec RunSpec, seed uint64, experiment FirstHourExperiment, config ReputationCareerConfig) (ReputationCareerResult, error) {
	if config.Bundle.ReputationTree == nil || config.Bundle.Economy == nil || suite.Bundle.Prestige == nil {
		return ReputationCareerResult{}, ErrReputationCareer
	}
	raw, err := json.Marshal(suite.Bundle.Prestige)
	if err != nil {
		return ReputationCareerResult{}, err
	}
	var fields map[string]any
	if err := json.Unmarshal(raw, &fields); err != nil {
		return ReputationCareerResult{}, err
	}
	fields["threshold"] = config.Threshold
	raw, _ = json.Marshal(fields)
	policy, err := prestigecore.LoadPolicy(raw)
	if err != nil {
		return ReputationCareerResult{}, fmt.Errorf("%w: fixture threshold %q: %v", ErrReputationCareer, config.Threshold, err)
	}
	career := &careerRuntime{config: config, policy: policy, purchased: []string{}, applied: []string{}}
	// The whole career runs on the tree bundle: identical to the suite bundle
	// except for the reputation.founder_bonus declaration row R2 adds.
	careerSuite := *suite
	careerSuite.Bundle = config.Bundle
	result, _, career := careerSuite.runWithCareer(spec, seed, experiment, false, career)
	if result.Outcome != "completed" {
		return ReputationCareerResult{}, fmt.Errorf("%w: seed %d outcome %s %v", ErrReputationCareer, seed, result.Outcome, result.InvariantFailures)
	}
	// A career whose run-3 Garage crossing lies beyond the ratified horizon
	// returns a nil RunThreeGateMS; callers must report it, never drop it.
	available := int64(0)
	if count := len(result.ReputationExits); count != 0 {
		available = result.ReputationExits[count-1].LevelAfter
	}
	for _, id := range career.purchased {
		node, _ := config.Bundle.ReputationTree.Node(id)
		available -= node.Cost
	}
	return ReputationCareerResult{Run: result, PurchasedNodeIDs: career.purchased, AppliedStarterIDs: career.applied,
		BonusFactor: career.factor, RunThreeGateMS: career.gateMS, ReputationAvailable: available}, nil
}

func (runtime *firstHourRuntime) prestigePolicy() *prestigecore.Policy {
	if runtime.career != nil {
		return runtime.career.policy
	}
	return runtime.suite.Bundle.Prestige
}

// external is the run's frozen Founder contribution set as production reads
// it; outside a career run 3 it is empty (the first-hour suite is tree-less).
func (runtime *firstHourRuntime) external() []multiplier.Contribution {
	if runtime.career == nil || runtime.company.RunSeq != 3 {
		return nil
	}
	return runtime.career.bonus
}

// applyCareerExit is the run-2 elective Exit in a career: credit Reputation
// under the fixture threshold, buy nodes by the declared policy, and assemble
// run 3 with the served starter application and frozen bonus.
func (runtime *firstHourRuntime) applyCareerExit(now time.Time, wallMS, attended int64) error {
	career := runtime.career
	tree := career.config.Bundle.ReputationTree
	terms, err := prestigecore.ComputeTerms(runtime.company, runtime.founder, career.policy, "collapse")
	if err != nil {
		return err
	}
	runtime.recordReputationExit("collapse", terms.ReputationDelta)
	runtime.founder.ReputationLevel += terms.ReputationDelta
	runtime.founder.AgeMS += attended
	runtime.founder.ExitHistory = append(runtime.founder.ExitHistory, save.ExitRecord{RunID: runtime.company.RunSeq, ExitType: "collapse", OccurredAt: now, ReputationDelta: terms.ReputationDelta})
	runtime.founderAttendedMS += attended
	runtime.observeRunEnded(wallMS, "collapse")
	if runtime.founder.ReputationNodesOwned == nil {
		runtime.founder.ReputationNodesOwned = []string{}
	}
	for ordinal := int64(0); career.config.Policy != CareerNone; ordinal++ {
		available := []string{}
		for _, node := range tree.Nodes() {
			if _, rejection, err := tree.Purchase(runtime.founder.ReputationLevel, runtime.founder.ReputationSpent, runtime.founder.ReputationNodesOwned, node.NodeID); err == nil && rejection == nil {
				available = append(available, node.NodeID)
			}
		}
		if len(available) == 0 {
			break
		}
		choice := available[0]
		switch career.config.Policy {
		case CareerCheapest:
			best := int64(-1)
			for _, id := range available {
				node, _ := tree.Node(id)
				if best < 0 || node.Cost < best {
					best, choice = node.Cost, id
				}
			}
		case CareerSeededUniform:
			index, drawErr := firstHourBoundedDraw("reputation_career", runtime.policy.PolicyID, runtime.policy.PolicyVersion, runtime.seed, runtime.company.RunSeq, ordinal, uint64(len(available)))
			if drawErr != nil {
				return drawErr
			}
			choice = available[index]
		default:
			return fmt.Errorf("%w: unknown career policy %q", ErrReputationCareer, career.config.Policy)
		}
		applied, rejection, err := tree.Purchase(runtime.founder.ReputationLevel, runtime.founder.ReputationSpent, runtime.founder.ReputationNodesOwned, choice)
		if err != nil || rejection != nil {
			return fmt.Errorf("%w: career purchase %s", ErrReputationCareer, choice)
		}
		runtime.founder.ReputationSpent, runtime.founder.ReputationNodesOwned, runtime.founder.ReputationUnlockPPM = applied.SpentAfter, applied.OwnedAfter, applied.UnlockPPMAfter
		career.purchased = append(career.purchased, choice)
	}
	next, err := prestigecore.NewRunState(runtime.suite.Bundle.Economy, runtime.company, runtime.founder, now)
	if err != nil {
		return err
	}
	if career.applied, err = production.ApplyReputationStarters(career.config.Bundle, runtime.founder, next); err != nil {
		return err
	}
	if career.applied == nil {
		career.applied = []string{}
	}
	factor, err := tree.BonusFactor(runtime.founder.ReputationLevel, runtime.founder.ReputationSpent, runtime.founder.ReputationUnlockPPM)
	if err != nil {
		return err
	}
	career.factor = factor.String()
	career.bonus = []multiplier.Contribution{{SourceID: tree.Bonus.SourceID, Slot: tree.Bonus.Slot, Target: tree.Bonus.Target, Factor: factor}}
	if _, err := production.ResolveFrozenContributions(career.config.Bundle.Economy, []save.FrozenContribution{{SourceID: tree.Bonus.SourceID,
		Slot: tree.Bonus.Slot, Target: tree.Bonus.Target, Factor: factor.String()}}); err != nil {
		return err
	}
	runtime.company = next
	runtime.revision++
	return validateFirstHourCompany(runtime.suite.Bundle.Economy, runtime.company)
}

// observeCareerGate records run 3's Garage crossing on the Company attended
// clock, the same clock as the run-2 milestone.
func (runtime *firstHourRuntime) observeCareerGate(request production.IntentRequest, now time.Time) error {
	if runtime.career == nil || runtime.company.RunSeq != 3 || request.Kind != production.IntentCrossGate || runtime.career.gateMS != nil {
		return nil
	}
	attended, err := prestigecore.AttendedMS(runtime.company, now)
	if err != nil {
		return err
	}
	runtime.career.gateMS = &attended
	return nil
}

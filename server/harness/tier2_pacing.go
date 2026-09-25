package harness

import (
	"errors"
	"fmt"
	"sort"
	"time"

	prestigecore "cloud-clicker/server/prestige"
	"cloud-clicker/server/production"
	"cloud-clicker/server/save"
)

// Tier 2 pacing (rfc/tier2-content.md §P1/§P2, fixture-first subset). The
// runner reuses the ratified T0–T1 policy literals with two §P1 changes: the
// elective Exit rule is taken only while Founder Exit history has exactly one
// entry (t01_c32_readiness_once), and gate.t1_to_t2 joins the crossing set.
// Offers are never answered (offer_rule: ignore; the simulated transitions
// spawn none). The allocation arms and reference.greedy v2 are held on the
// unruled headcount seat source (OD-1), so only chaos and casual run.

const tier2GateID = "gate.t1_to_t2"

var ErrTier2Pacing = errors.New("invalid tier 2 pacing measurement")

type Tier2PacingConfig struct {
	// Bundle is the candidate bundle that declares gate.t1_to_t2.
	Bundle production.CatalogBundle
	// Experiment is the curriculum branch tuple; the measurement uses the
	// epoch-8 tuple the ratified first-hour evidence and composed runs use.
	Experiment FirstHourExperiment
}

type Tier2PacingResult struct {
	PolicyID string `json:"policy_id"`
	Seed     uint64 `json:"seed"`
	Outcome  string `json:"outcome"`
	// GateFounderAttendedMS is milestone.it_company_gate: Founder-attended
	// time at the first gate.t1_to_t2 crossing in any run (run_seq null), or
	// nil when the horizon ends first (a must_reach failure, never dropped).
	GateFounderAttendedMS *int64   `json:"gate_founder_attended_ms"`
	GateRunSeq            int64    `json:"gate_run_seq"`
	ElectiveExits         int64    `json:"elective_exits"`
	TransitionCount       int64    `json:"transition_count"`
	InvariantFailures     []string `json:"invariant_failures"`
}

type tier2Runtime struct {
	gateMS        *int64
	gateRunSeq    int64
	electiveExits int64
}

func (tier2 *tier2Runtime) done() bool { return tier2.gateMS != nil }

// gateIDs is the crossing set: the scenario's gates, plus gate.t1_to_t2 in the
// Tier 2 measurement.
func (runtime *firstHourRuntime) gateIDs() []string {
	gates := runtime.suite.GateIDs()
	if runtime.tier2 != nil {
		gates = append(append([]string(nil), gates...), tier2GateID)
		sort.Strings(gates)
	}
	return gates
}

// applyTier2Exit is the single elective collapse Exit of the measurement; the
// next run is assembled exactly as the scripted ending assembles run 2.
func (runtime *firstHourRuntime) applyTier2Exit(now time.Time, wallMS, attended int64) error {
	terms, err := prestigecore.ComputeTerms(runtime.company, runtime.founder, runtime.prestigePolicy(), "collapse")
	if err != nil {
		return err
	}
	runtime.recordReputationExit("collapse", terms.ReputationDelta)
	runtime.recordCommand(wallMS, production.ModeOnline, production.IntentRequest{Kind: production.IntentWindDown})
	runtime.founder.ReputationLevel += terms.ReputationDelta
	runtime.founder.RouteKnowledgeBalance += terms.RouteKnowledge
	runtime.founder.AgeMS += attended
	runtime.founder.ExitHistory = append(runtime.founder.ExitHistory, save.ExitRecord{RunID: runtime.company.RunSeq, ExitType: "collapse",
		OccurredAt: now, ReputationDelta: terms.ReputationDelta})
	runtime.founderAttendedMS += attended
	runtime.observeRunEnded(wallMS, "collapse")
	next, err := prestigecore.NewRunState(runtime.suite.Bundle.Economy, runtime.company, runtime.founder, now)
	if err != nil {
		return err
	}
	runtime.company = next
	runtime.revision++
	runtime.tier2.electiveExits++
	return validateFirstHourCompany(runtime.suite.Bundle.Economy, runtime.company)
}

func (runtime *firstHourRuntime) observeTier2Gate(request production.IntentRequest, now time.Time) error {
	if runtime.tier2 == nil || runtime.tier2.gateMS != nil || request.Kind != production.IntentCrossGate || request.GateID != tier2GateID {
		return nil
	}
	attended, err := prestigecore.AttendedMS(runtime.company, now)
	if err != nil {
		return err
	}
	value := runtime.founderAttendedMS + attended
	runtime.tier2.gateMS, runtime.tier2.gateRunSeq = &value, runtime.company.RunSeq
	return nil
}

// RunTier2Pacing runs one seed of the Tier 2 pacing measurement.
func (suite *FirstHourSuite) RunTier2Pacing(spec RunSpec, seed uint64, config Tier2PacingConfig) (Tier2PacingResult, error) {
	if config.Bundle.Routes == nil || config.Bundle.Economy == nil || config.Bundle.Prestige == nil {
		return Tier2PacingResult{}, ErrTier2Pacing
	}
	if _, ok := config.Bundle.Routes.Gate(tier2GateID); !ok {
		return Tier2PacingResult{}, fmt.Errorf("%w: bundle does not declare %s", ErrTier2Pacing, tier2GateID)
	}
	if spec.PolicyID != "chaos.t0_t1" && spec.PolicyID != "casual.t0_t1" {
		return Tier2PacingResult{}, fmt.Errorf("%w: policy %s is held (reference.greedy v2 needs the §P1 allocation arms)", ErrTier2Pacing, spec.PolicyID)
	}
	measurement := *suite
	measurement.Bundle = config.Bundle
	run, _, _, tier2 := measurement.runWithModes(spec, seed, config.Experiment, false, nil, &tier2Runtime{})
	result := Tier2PacingResult{PolicyID: spec.PolicyID, Seed: seed, Outcome: "completed", InvariantFailures: []string{}, TransitionCount: run.TransitionCount}
	// The ratified first-hour milestones are run-1/run-2 observations; only
	// the pacing runner's own failures (guards, rejected commands) are carried.
	for _, failure := range run.InvariantFailures {
		if len(failure) < len("must_reach:") || failure[:len("must_reach:")] != "must_reach:" {
			result.InvariantFailures = append(result.InvariantFailures, failure)
		}
	}
	if tier2 != nil {
		result.GateFounderAttendedMS, result.GateRunSeq, result.ElectiveExits = cloneInt64(tier2.gateMS), tier2.gateRunSeq, tier2.electiveExits
	}
	if result.GateFounderAttendedMS == nil {
		result.InvariantFailures = append(result.InvariantFailures, "must_reach:milestone.it_company_gate")
	}
	if len(result.InvariantFailures) != 0 {
		result.Outcome = "failed"
	}
	return result, nil
}

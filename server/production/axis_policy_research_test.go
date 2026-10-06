package production

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"testing"
	"time"

	"cloud-clicker/server/economy"
	"cloud-clicker/server/save"
)

// R-012 observation only: these are separate fixed-mode calls, not a newly
// adopted definition of one offline episode or a production anchor format.
type policyResearchProfile struct {
	ID       string         `json:"id"`
	Mode     EvaluationMode `json:"mode"`
	End      int64          `json:"end_ms"`
	Credit   int64          `json:"initial_credit_ms"`
	Burst    int64          `json:"initial_burst_ms"`
	Provider int64          `json:"purchased_provider_count"`
}

type policyResearchPhase struct {
	At     int64            `json:"at_ms"`
	Result EvaluationResult `json:"actual_evaluation_result"`
}

type policyResearchBranch struct {
	Final  json.RawMessage       `json:"final_state"`
	Phases []policyResearchPhase `json:"phases"`
}

type policyResearchObservation struct {
	Profile policyResearchProfile `json:"profile"`
	Initial json.RawMessage       `json:"initial_state"`
	One     policyResearchBranch  `json:"one"`
	Split   policyResearchBranch  `json:"split"`
	Diff    []string              `json:"differing_top_level_fields"`
}

type policyResearchReport struct {
	Version    int                         `json:"version"`
	Acceptance string                      `json:"acceptance_status"`
	BundleHash string                      `json:"fixture_bundle_hash"`
	Sources    map[string]string           `json:"source_sha256"`
	Rows       []policyResearchObservation `json:"rows"`
	Divergent  int                         `json:"differing_profiles"`
}

const policyResearchPath = "../../testdata/axis-stack/policy-research-v1.json"

func policyResearchRestore(t *testing.T, bundle CatalogBundle, state *save.State, start time.Time) (*save.State, []byte) {
	t.Helper()
	if err := bundle.ValidateFoundationState(state); err != nil {
		t.Fatalf("policy diagnostic admission: %v", err)
	}
	encoded := mustEncodeState(t, state)
	restored, err := save.RestoreState(encoded, save.VersionForState(state), bundle.Economy, economy.ScopeCompany, start)
	if err != nil {
		t.Fatalf("policy diagnostic restore: %v", err)
	}
	if err := bundle.ValidateFoundationState(restored); err != nil {
		t.Fatalf("policy restored admission: %v", err)
	}
	if !bytes.Equal(encoded, mustEncodeState(t, restored)) {
		t.Fatal("full Company restoration changed bytes")
	}
	return restored, encoded
}

func policyResearchRun(t *testing.T, bundle CatalogBundle, profile policyResearchProfile, start time.Time, cuts []int64) ([]byte, policyResearchBranch) {
	t.Helper()
	state := axisTimingState(t, bundle, start)
	state.ComputeCreditMS, state.ComputeBurstRemainingMS = profile.Credit, profile.Burst
	state.GeneratorCounts["generator.beige_tower_v2"] = profile.Provider
	state.GeneratorPurchasedTotal = 1 + profile.Provider
	state, initial := policyResearchRestore(t, bundle, state, start)
	branch := policyResearchBranch{}
	for _, at := range cuts {
		before := mustEncodeState(t, state)
		cursor := state.EvaluatedThrough
		oldCredit, oldBurst := state.ComputeCreditMS, state.ComputeBurstRemainingMS
		oldProvision, oldRemainder := state.GeneratorProvisioned["generator.beige_tower"], state.ProvisionRemaindersPPM["generator.beige_tower"]
		contributions, err := assembleContributions(state, bundle.Economy, nil)
		if err != nil {
			t.Fatal(err)
		}
		result, err := Evaluate(state, bundle.Economy, start.Add(time.Duration(at)*time.Millisecond), profile.Mode, contributions)
		if err != nil {
			t.Fatalf("%s at%d invalid observation: %v", profile.ID, at, err)
		}
		elapsed := at - cursor.Sub(start).Milliseconds()
		if elapsed <= 0 {
			if result.ElapsedMS != 0 || result.ProductionMS != 0 || result.BankedCreditMS != 0 || !bytes.Equal(before, mustEncodeState(t, state)) {
				t.Fatalf("%s zero-work changed full state", profile.ID)
			}
		} else {
			policy := bundle.Economy.OfflinePolicy()
			production, bank := elapsed, int64(0)
			if profile.Mode == ModeOffline {
				production = min(elapsed, policy.AccrualCapMS)
				bank = min((elapsed-production)*policy.BankRatioNumerator/policy.BankRatioDenominator, policy.BankCapMS-oldCredit)
			}
			remainingBurst := max(int64(0), oldBurst-elapsed)
			if result.ElapsedMS != elapsed || result.ProductionMS != production || result.BankedCreditMS != bank ||
				state.ComputeCreditMS != oldCredit+bank || state.ComputeBurstRemainingMS != remainingBurst || !state.EvaluatedThrough.Equal(start.Add(time.Duration(at)*time.Millisecond)) {
				t.Fatalf("%s integer policy binding failed at%d: %+v credit%d burst%d", profile.ID, at, result, state.ComputeCreditMS, state.ComputeBurstRemainingMS)
			}
			tick := bundle.Economy.ProvisionTickMS()
			offset := cursor.Sub(state.RunStartedAt).Milliseconds()
			boundaries := (offset+production)/tick - offset/tick
			// Fixture has one provision source: purchased beige-v2, no generated
			// providers. This integer oracle does not call materialization code.
			provider, _ := bundle.Economy.GeneratorClass("generator.beige_tower_v2")
			target, _ := bundle.Economy.GeneratorClass("generator.beige_tower")
			if provider.Provision == nil || provider.Provision.GeneratorID != target.ID || target.ProvisionedHardcap == nil ||
				state.GeneratorProvisioned[provider.ID] != 0 || tick != 60000 {
				t.Fatal("unsupported provision diagnostic population")
			}
			numerator := oldRemainder + boundaries*profile.Provider*provider.Provision.RatePPM
			generated := min(numerator/1000000, target.ProvisionedHardcap.Count-oldProvision)
			if state.GeneratorProvisioned[target.ID] != oldProvision+generated || state.ProvisionRemaindersPPM[target.ID] != numerator%1000000 {
				t.Fatalf("%s provision binding failed at%d", profile.ID, at)
			}
		}
		state, branch.Final = policyResearchRestore(t, bundle, state, start)
		branch.Phases = append(branch.Phases, policyResearchPhase{at, result})
	}
	return initial, branch
}

func policyResearchDifferences(t *testing.T, one, split []byte) []string {
	t.Helper()
	a, b := map[string]json.RawMessage{}, map[string]json.RawMessage{}
	if err := json.Unmarshal(one, &a); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(split, &b); err != nil {
		t.Fatal(err)
	}
	keys := map[string]bool{}
	for key := range a {
		keys[key] = true
	}
	for key := range b {
		keys[key] = true
	}
	different := []string{}
	for key := range keys {
		if !bytes.Equal(a[key], b[key]) {
			different = append(different, key)
		}
	}
	sort.Strings(different)
	return different
}

func TestAxisPolicyBoundaryResearch(t *testing.T) {
	bundle := researchBundleCap(t, axisContentBundle(t), "1e100")
	policy := bundle.Economy.OfflinePolicy()
	var profiles []policyResearchProfile
	for _, end := range []int64{policy.AccrualCapMS - 1, policy.AccrualCapMS, policy.AccrualCapMS + 1, policy.AccrualCapMS + 2, 2 * policy.AccrualCapMS, 4 * policy.AccrualCapMS} {
		for _, credit := range []int64{0, policy.BankCapMS - 1} {
			profiles = append(profiles, policyResearchProfile{fmt.Sprintf("offline/%d/credit%d", end, credit), ModeOffline, end, credit, 0, 0})
		}
	}
	for _, mode := range []EvaluationMode{ModeOnline, ModeOffline} {
		for _, burst := range []int64{0, 1, 1500, policy.BurstMaxDurationMS} {
			for _, end := range []int64{3114, policy.BurstMaxDurationMS + 1} {
				profiles = append(profiles, policyResearchProfile{fmt.Sprintf("burst/%s/%d/end%d", mode, burst, end), mode, end, 0, burst, 0})
			}
		}
		for _, provider := range []int64{1, 10} {
			for _, end := range []int64{59999, 60000, 120001} {
				profiles = append(profiles, policyResearchProfile{fmt.Sprintf("provision/%s/%d/end%d", mode, provider, end), mode, end, 0, 0, provider})
			}
		}
		for _, end := range []int64{0, -1} {
			profiles = append(profiles, policyResearchProfile{fmt.Sprintf("clock/%s/end%d", mode, end), mode, end, 0, 0, 0})
		}
	}
	if len(profiles) != 44 {
		t.Fatal("incomplete policy population")
	}
	report := policyResearchReport{Version: 1, Acceptance: "NOT_PROVEN: production AC6 remains red", BundleHash: bundle.ConstantsHash,
		Sources: sourceResearchHashes(t, []string{"server/production/axis_policy_research_test.go", "server/production/axis_rate_source_research_test.go", "server/production/axis_timing_test.go", "server/production/axis_stack_test.go", "server/production/axis_partition_research_test.go", "balance/testdata/axis-stack/economy-v5-fixture.json", "server/production/engine.go", "server/production/content.go", "server/production/accrual.go", "server/economy/ledger.go", "server/save/state.go"})}
	start := time.Date(2026, 10, 7, 12, 0, 0, 0, time.UTC)
	seen := map[string]bool{}
	for _, profile := range profiles {
		if seen[profile.ID] {
			t.Fatal("duplicate policy profile")
		}
		seen[profile.ID] = true
		initial, one := policyResearchRun(t, bundle, profile, start, []int64{profile.End})
		otherInitial, split := policyResearchRun(t, bundle, profile, start, []int64{profile.End / 2, profile.End})
		if !bytes.Equal(initial, otherInitial) {
			t.Fatal("paired initial context differs")
		}
		different := policyResearchDifferences(t, one.Final, split.Final)
		if len(different) > 0 {
			report.Divergent++
		}
		report.Rows = append(report.Rows, policyResearchObservation{profile, json.RawMessage(initial), one, split, different})
	}
	if len(report.Rows) != 44 {
		t.Fatal("truncated policy population")
	}
	sourceResearchArtifact(t, policyResearchPath, "UPDATE_POLICY_RESEARCH", report)
	t.Logf("R-012 actual policy observer:44 complete full-state/restore pairs, differing%d; AC6 NOT_PROVEN", report.Divergent)
}

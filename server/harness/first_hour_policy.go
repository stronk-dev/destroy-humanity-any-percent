package harness

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

const (
	firstHourPolicyID      = "first_hour_policy.v1"
	firstHourPolicyVersion = 3
)

type FirstHourPolicyRegistry struct {
	SchemaVersion     int                        `json:"schema_version"`
	ID                string                     `json:"id"`
	Version           int                        `json:"version"`
	SeedDerivation    FirstHourSeedDerivation    `json:"seed_derivation"`
	EnumerationOrder  FirstHourEnumerationOrder  `json:"enumeration_order"`
	Policies          []FirstHourPolicy          `json:"policies"`
	ExitRules         []FirstHourExitRule        `json:"exit_rules"`
	WaitLegality      FirstHourWaitLegality      `json:"wait_legality"`
	FirstActionTiming FirstHourFirstActionTiming `json:"first_action_timing"`
}

type FirstHourSeedDerivation struct {
	Kind                  string                         `json:"kind"`
	DomainTags            FirstHourDomainTags            `json:"domain_tags"`
	Material              []string                       `json:"material"`
	IntegerEncoding       string                         `json:"integer_encoding"`
	Separator             string                         `json:"separator"`
	Draw                  string                         `json:"draw"`
	Reduction             string                         `json:"reduction"`
	FixedNonDecisionDraws FirstHourFixedNonDecisionDraws `json:"fixed_material_for_non_decision_draws"`
	RangeSemantics        string                         `json:"range_semantics"`
}

type FirstHourDomainTags struct {
	Decision      string `json:"decision"`
	SessionJitter string `json:"session_jitter"`
}

type FirstHourFixedNonDecisionDraws struct {
	RunSeq          string `json:"run_seq"`
	DecisionOrdinal string `json:"decision_ordinal"`
	Note            string `json:"note"`
}

type FirstHourEnumerationOrder struct {
	Kind        string   `json:"kind"`
	Classes     []string `json:"classes"`
	WithinClass string   `json:"within_class"`
	Note        string   `json:"note"`
}

type FirstHourPolicy struct {
	PolicyID             string                  `json:"policy_id"`
	PolicyVersion        int                     `json:"policy_version"`
	Model                string                  `json:"model"`
	Sessions             []FirstHourSession      `json:"sessions"`
	SessionStartJitterMS *int64                  `json:"session_start_jitter_ms,omitempty"`
	ActionCadenceMS      int64                   `json:"action_cadence_ms"`
	Decision             FirstHourDecision       `json:"decision"`
	GateCrossing         string                  `json:"gate_crossing"`
	RunContinuation      string                  `json:"run_continuation"`
	ExitRule             string                  `json:"exit_rule"`
	SessionsClock        *FirstHourSessionsClock `json:"sessions_clock,omitempty"`
	SessionJitter        *FirstHourSessionJitter `json:"session_jitter,omitempty"`
	Objectives           *FirstHourObjectives    `json:"objectives,omitempty"`
}

type FirstHourSession struct {
	StartMS       int64 `json:"start_ms"`
	DurationMS    int64 `json:"duration_ms"`
	RepeatEveryMS int64 `json:"repeat_every_ms"`
}

type FirstHourDecision struct {
	Kind             string   `json:"kind"`
	LegalCommands    []string `json:"legal_commands"`
	ManualBatchCount int64    `json:"manual_batch_count"`
	Affordability    string   `json:"affordability"`
	WaitAdvancesTo   string   `json:"wait_advances_to"`
}

type FirstHourSessionsClock struct {
	Origin          string `json:"origin"`
	Units           string `json:"units"`
	SpansRuns       bool   `json:"spans_runs"`
	RestartOnRunEnd bool   `json:"restart_on_run_end"`
	GapIsOffline    bool   `json:"gap_is_offline"`
	OnSessionOpen   string `json:"on_session_open"`
}

type FirstHourSessionJitter struct {
	Kind           string   `json:"kind"`
	DomainTag      string   `json:"domain_tag"`
	DrawnPer       []string `json:"drawn_per"`
	StartOffsetMS  []int64  `json:"start_offset_ms"`
	DurationMS     []int64  `json:"duration_ms"`
	RangeSemantics string   `json:"range_semantics"`
	Rationale      string   `json:"rationale"`
}

type FirstHourObjectives struct {
	Kind                  string `json:"kind"`
	GateScope             string `json:"gate_scope"`
	MultiResourceOrdering string `json:"multi_resource_ordering"`
	TerminalObjective     string `json:"terminal_objective"`
	ExitPrecedence        string `json:"exit_precedence"`
	Note                  string `json:"note"`
}

type FirstHourExitRule struct {
	ID     string                   `json:"id"`
	All    []FirstHourExitCondition `json:"all"`
	Action string                   `json:"action"`
}

type FirstHourExitCondition struct {
	Kind       string `json:"kind"`
	RunSeq     int64  `json:"run_seq,omitempty"`
	GateID     string `json:"gate_id,omitempty"`
	AttendedMS int64  `json:"attended_ms,omitempty"`
}

type FirstHourWaitLegality struct {
	Rule      string `json:"rule"`
	Rationale string `json:"rationale"`
}

type FirstHourFirstActionTiming struct {
	Kind string `json:"kind"`
	Note string `json:"note"`
}

func LoadFirstHourPolicy(repositoryRoot, policyPath string) (FirstHourPolicyRegistry, []byte, string, error) {
	data, err := os.ReadFile(filepath.Join(repositoryRoot, filepath.FromSlash(policyPath)))
	if err != nil {
		return FirstHourPolicyRegistry{}, nil, "", err
	}
	var registry FirstHourPolicyRegistry
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&registry); err != nil {
		return FirstHourPolicyRegistry{}, nil, "", fmt.Errorf("first-hour policy: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return FirstHourPolicyRegistry{}, nil, "", errors.New("first-hour policy must contain exactly one JSON value")
	}
	if err := validateFirstHourPolicy(registry); err != nil {
		return FirstHourPolicyRegistry{}, nil, "", err
	}
	digest := sha256.Sum256(data)
	return registry, data, "sha256:" + hex.EncodeToString(digest[:]), nil
}

func validateFirstHourPolicy(registry FirstHourPolicyRegistry) error {
	if registry.SchemaVersion != 1 || registry.ID != firstHourPolicyID || registry.Version != firstHourPolicyVersion {
		return errors.New("invalid first-hour policy envelope")
	}
	seed := registry.SeedDerivation
	if seed.Kind != "sha256_substream" || seed.DomainTags.Decision != "decision" || seed.DomainTags.SessionJitter != "jitter" ||
		!equalStrings(seed.Material, []string{"domain_tag", "policy_id", "policy_version", "seed", "run_seq", "decision_ordinal"}) ||
		seed.IntegerEncoding != "decimal_ascii_no_padding" || seed.Separator != "0x1f" ||
		seed.Draw != "first_8_bytes_big_endian_uint64" || seed.Reduction != "modulo_legal_command_count" ||
		seed.FixedNonDecisionDraws.RunSeq != "0" || seed.FixedNonDecisionDraws.DecisionOrdinal != "session_index" ||
		seed.RangeSemantics != "half_open_lower_inclusive_upper_exclusive" {
		return errors.New("invalid first-hour seed derivation")
	}
	if registry.EnumerationOrder.Kind != "fixed_command_class_then_raw_byte_id" ||
		!equalStrings(registry.EnumerationOrder.Classes, []string{"perform_manual_batch", "buy_generator", "buy_upgrade", "wait"}) ||
		registry.EnumerationOrder.WithinClass != "raw_byte_ascending_purchasable_id" {
		return errors.New("invalid first-hour command enumeration")
	}
	if registry.WaitLegality.Rule != "legal_only_when_production_income_positive" || registry.FirstActionTiming.Kind != "at_session_start" {
		return errors.New("invalid first-hour timing policy")
	}
	if len(registry.Policies) != 3 || len(registry.ExitRules) != 1 {
		return errors.New("invalid first-hour policy registry cardinality")
	}
	expectedIDs := []string{"casual.t0_t1", "chaos.t0_t1", "reference.greedy"}
	ids := make([]string, 0, len(registry.Policies))
	for _, policy := range registry.Policies {
		if err := validateFirstHourPolicyRow(policy); err != nil {
			return err
		}
		ids = append(ids, policy.PolicyID)
	}
	sort.Strings(ids)
	if !equalStrings(ids, expectedIDs) {
		return errors.New("invalid first-hour policy ids")
	}
	rule := registry.ExitRules[0]
	if rule.ID != "t01_c32_readiness" || rule.Action != "wind_down_at_first_true_boundary" || len(rule.All) != 4 ||
		rule.All[0].Kind != "run_seq_at_least" || rule.All[0].RunSeq != 2 ||
		rule.All[1].Kind != "gate_crossed" || rule.All[1].GateID != "gate.t0_to_t1" ||
		rule.All[2].Kind != "founder_attended_ms_at_least" || rule.All[2].AttendedMS != 2_700_000 ||
		rule.All[3].Kind != "previewed_exit_grants_any_persistent_value" {
		return errors.New("invalid first-hour exit rule")
	}
	return nil
}

func validateFirstHourPolicyRow(policy FirstHourPolicy) error {
	if policy.PolicyVersion != 1 || policy.PolicyID == "" || policy.Model == "" || len(policy.Sessions) != 1 ||
		policy.ActionCadenceMS <= 0 || policy.Decision.ManualBatchCount != 1 ||
		!equalStrings(policy.Decision.LegalCommands, []string{"perform_manual_batch", "buy_generator", "buy_upgrade", "wait"}) ||
		policy.GateCrossing != "first_boundary_requirements_met" || policy.RunContinuation != "same_policy_new_run" ||
		policy.ExitRule != "t01_c32_readiness" {
		return fmt.Errorf("invalid first-hour policy row %q", policy.PolicyID)
	}
	session := policy.Sessions[0]
	if session.StartMS < 0 || session.DurationMS <= 0 || session.RepeatEveryMS < 0 {
		return fmt.Errorf("invalid first-hour session row %q", policy.PolicyID)
	}
	switch policy.PolicyID {
	case "chaos.t0_t1":
		if policy.Decision.Kind != "seeded_uniform_over_legal" || policy.Decision.Affordability != "current_balance_only" ||
			policy.Decision.WaitAdvancesTo != "next_action_boundary" || policy.ActionCadenceMS != 2_000 ||
			policy.SessionStartJitterMS == nil || *policy.SessionStartJitterMS != 0 || policy.SessionsClock != nil || policy.SessionJitter != nil || policy.Objectives != nil {
			return errors.New("invalid chaos first-hour policy")
		}
	case "casual.t0_t1":
		if policy.Decision.Kind != "cheapest_affordable_then_manual" || policy.Decision.Affordability != "current_balance_only" ||
			policy.Decision.WaitAdvancesTo != "next_action_boundary" || policy.ActionCadenceMS != 5_000 || policy.SessionsClock == nil || policy.SessionJitter == nil ||
			policy.SessionStartJitterMS != nil || policy.Objectives != nil {
			return errors.New("invalid casual first-hour policy")
		}
		clock, jitter := policy.SessionsClock, policy.SessionJitter
		if clock.Origin != "founder_genesis" || clock.Units != "wall_ms" || !clock.SpansRuns || clock.RestartOnRunEnd || !clock.GapIsOffline ||
			clock.OnSessionOpen != "apply_offline_catchup_then_evaluate_online" || jitter.Kind != "independent_per_session" || jitter.DomainTag != "jitter" ||
			!equalStrings(jitter.DrawnPer, []string{"policy_id", "seed", "session_index"}) ||
			!equalInt64s(jitter.StartOffsetMS, []int64{0, 600_000}) || !equalInt64s(jitter.DurationMS, []int64{600_000, 1_200_000}) ||
			jitter.RangeSemantics != "half_open_lower_inclusive_upper_exclusive" {
			return errors.New("invalid casual first-hour schedule")
		}
	case "reference.greedy":
		if policy.Decision.Kind != "t01_c20_projected_time_ranker" || policy.Decision.Affordability != "earliest_affordable" ||
			policy.Decision.WaitAdvancesTo != "bank_or_next_boundary" || policy.ActionCadenceMS != 1_000 || policy.Objectives == nil ||
			policy.SessionStartJitterMS == nil || *policy.SessionStartJitterMS != 0 || policy.SessionsClock != nil || policy.SessionJitter != nil {
			return errors.New("invalid reference first-hour policy")
		}
		objectives := policy.Objectives
		if objectives.Kind != "ordered_next_unmet_progression_requirement" || objectives.GateScope != "gates_declared_by_the_scenario_segments_only" ||
			objectives.MultiResourceOrdering != "raw_byte_ascending_resource_id_then_largest_remaining_deficit_first" ||
			objectives.TerminalObjective != "exit_readiness_predicate" || objectives.ExitPrecedence != "c32_exit_rule_preempts_the_ranker_at_any_boundary_where_it_is_true" {
			return errors.New("invalid reference first-hour objectives")
		}
	default:
		return fmt.Errorf("unknown first-hour policy %q", policy.PolicyID)
	}
	return nil
}

func (registry FirstHourPolicyRegistry) Policy(id string, version int) (FirstHourPolicy, bool) {
	for _, policy := range registry.Policies {
		if policy.PolicyID == id && policy.PolicyVersion == version {
			return policy, true
		}
	}
	return FirstHourPolicy{}, false
}

func firstHourDraw(domainTag, policyID string, policyVersion int, seed uint64, runSeq, ordinal int64) uint64 {
	material := []string{domainTag, policyID, strconv.Itoa(policyVersion), strconv.FormatUint(seed, 10), strconv.FormatInt(runSeq, 10), strconv.FormatInt(ordinal, 10)}
	digest := sha256.Sum256([]byte(strings.Join(material, "\x1f")))
	return binary.BigEndian.Uint64(digest[:8])
}

func firstHourBoundedDraw(domainTag, policyID string, policyVersion int, seed uint64, runSeq, ordinal int64, bound uint64) (uint64, error) {
	if bound == 0 {
		return 0, errors.New("first-hour draw bound is zero")
	}
	return firstHourDraw(domainTag, policyID, policyVersion, seed, runSeq, ordinal) % bound, nil
}

func equalStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func equalInt64s(left, right []int64) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

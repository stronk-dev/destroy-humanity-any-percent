package prestige

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
	"time"

	"cloud-clicker/server/decimal"
)

var ErrInvalidPolicy = errors.New("invalid prestige policy")
var mechanicalIDPattern = regexp.MustCompile(`^[a-z][a-z0-9_]*(?:\.[a-z][a-z0-9_]*)*$`)

// MaximumOfferDurationMS is a representation limit, not a balance value. P3's
// existing expiry path converts milliseconds to signed int64 nanoseconds.
const MaximumOfferDurationMS = int64((1<<63 - 1) / time.Millisecond)

type Policy struct {
	SchemaVersion          int              `json:"schema_version"`
	ValueResourceID        string           `json:"value_resource_id"`
	Threshold              string           `json:"threshold"`
	ExitModifiersPPM       map[string]int64 `json:"exit_modifiers_ppm"`
	CollapseRouteKnowledge int64            `json:"collapse_route_knowledge"`
	CatchupCeilingMS       int64            `json:"catchup_ceiling_ms"`
	OfferDurationMS        int64            `json:"offer_duration_ms"`
	DeclineDriftPPM        int64            `json:"decline_drift_ppm"`
	SpawnGatePPM           []int64          `json:"spawn_gate_ppm"`
	AdvisorPerRunPPM       int64            `json:"advisor_per_run_ppm"`
	AdvisorCapPPM          int64            `json:"advisor_cap_ppm"`
	threshold              decimal.Decimal
}

func LoadPolicy(data []byte) (*Policy, error) {
	if !prestigePolicyHasExactShape(data) {
		return nil, ErrInvalidPolicy
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var policy Policy
	if err := decoder.Decode(&policy); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidPolicy, err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return nil, ErrInvalidPolicy
	}
	threshold, err := decimal.ParseCanonical(policy.Threshold)
	if err != nil || !threshold.IsStateValue() || !threshold.Gt(decimal.Zero) || policy.SchemaVersion != 1 || len(policy.SpawnGatePPM) != 10 ||
		!mechanicalIDPattern.MatchString(policy.ValueResourceID) ||
		policy.CollapseRouteKnowledge < 0 || policy.CollapseRouteKnowledge > decimal.MaxExactInteger ||
		policy.CatchupCeilingMS <= 0 || policy.CatchupCeilingMS > 86_400_000 ||
		policy.OfferDurationMS <= 0 || policy.OfferDurationMS > MaximumOfferDurationMS ||
		policy.DeclineDriftPPM < 0 || policy.DeclineDriftPPM > 1_000_000 ||
		policy.AdvisorPerRunPPM < 0 || policy.AdvisorPerRunPPM > 1_000_000 ||
		policy.AdvisorCapPPM < 0 || policy.AdvisorCapPPM > 1_000_000 {
		return nil, ErrInvalidPolicy
	}
	exitTypes := []string{"acquihire", "acquisition", "ipo", "collapse", "scripted_first"}
	if len(policy.ExitModifiersPPM) != len(exitTypes) {
		return nil, ErrInvalidPolicy
	}
	for _, kind := range exitTypes {
		value, ok := policy.ExitModifiersPPM[kind]
		if !ok || value < 0 || value > decimal.MaxExactInteger {
			return nil, ErrInvalidPolicy
		}
	}
	for _, value := range policy.SpawnGatePPM {
		if value < 0 || value > 1_000_000 {
			return nil, ErrInvalidPolicy
		}
	}
	policy.threshold = threshold
	return &policy, nil
}

func prestigePolicyHasExactShape(data []byte) bool {
	root, ok := prestigePolicyObject(data, "schema_version", "value_resource_id", "threshold", "exit_modifiers_ppm", "collapse_route_knowledge", "catchup_ceiling_ms", "offer_duration_ms", "decline_drift_ppm", "spawn_gate_ppm", "advisor_per_run_ppm", "advisor_cap_ppm")
	if !ok {
		return false
	}
	if _, ok := prestigePolicyObject(root["exit_modifiers_ppm"], "acquihire", "acquisition", "collapse", "ipo", "scripted_first"); !ok {
		return false
	}
	var gates []json.RawMessage
	if json.Unmarshal(root["spawn_gate_ppm"], &gates) != nil || gates == nil {
		return false
	}
	for _, value := range gates {
		if bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
			return false
		}
	}
	return true
}

func prestigePolicyObject(data []byte, keys ...string) (map[string]json.RawMessage, bool) {
	var object map[string]json.RawMessage
	if json.Unmarshal(data, &object) != nil || len(object) != len(keys) {
		return nil, false
	}
	for _, key := range keys {
		value, ok := object[key]
		if !ok || bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
			return nil, false
		}
	}
	return object, true
}

func (policy *Policy) ThresholdValue() decimal.Decimal { return policy.threshold }

func (policy *Policy) Modifier(exitType string) (int64, bool) {
	if policy == nil {
		return 0, false
	}
	value, ok := policy.ExitModifiersPPM[exitType]
	return value, ok
}

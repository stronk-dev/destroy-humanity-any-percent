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

type Policy struct {
	SchemaVersion          int              `json:"schema_version"`
	ValueResourceID        string           `json:"value_resource_id"`
	Threshold              string           `json:"threshold"`
	ExitModifiersPPM       map[string]int64 `json:"exit_modifiers_ppm"`
	CollapseRouteKnowledge int64            `json:"collapse_route_knowledge"`
	OfferDurationMS        int64            `json:"offer_duration_ms"`
	DeclineDriftPPM        int64            `json:"decline_drift_ppm"`
	SpawnGatePPM           []int64          `json:"spawn_gate_ppm"`
	AdvisorPerRunPPM       int64            `json:"advisor_per_run_ppm"`
	AdvisorCapPPM          int64            `json:"advisor_cap_ppm"`
	threshold              decimal.Decimal
}

func LoadPolicy(data []byte) (*Policy, error) {
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
		policy.OfferDurationMS <= 0 || policy.OfferDurationMS > int64((1<<63-1)/time.Millisecond) ||
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

func (policy *Policy) ThresholdValue() decimal.Decimal { return policy.threshold }

func (policy *Policy) Modifier(exitType string) (int64, bool) {
	if policy == nil {
		return 0, false
	}
	value, ok := policy.ExitModifiersPPM[exitType]
	return value, ok
}

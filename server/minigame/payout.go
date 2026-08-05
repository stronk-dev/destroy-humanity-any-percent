package minigame

import (
	"encoding/json"
	"errors"

	"cloud-clicker/server/decimal"
	"cloud-clicker/server/fixedgrid"
)

var ErrInvalidPayoutPolicy = errors.New("invalid minigame payout policy")

const payoutPPM = int64(1_000_000)

type PayoutPolicy struct {
	CreditedResourceID string `json:"credited_resource_id"`
	SendsPerDay        int64  `json:"sends_per_day"`
	PerSendCap         int64  `json:"per_send_cap"`
	ConversionPPM      int64  `json:"conversion_ppm"`
	PayoutScoreFactID  string `json:"payout_score_fact_id"`
	CapReasonKey       string `json:"cap_reason_key"`
}

type PayoutDeclarations struct {
	ResourceIDs  map[string]struct{}
	ScoreFactIDs map[string]struct{}
	CopyKeys     map[string]struct{}
}

type PayoutConversion struct {
	ReducedScore           int64
	ConvertedUnits         int64
	ConversionRemainderPPM int64
}

func LoadPayoutPolicy(data []byte, declarations PayoutDeclarations) (PayoutPolicy, error) {
	if !hasExactJSONKeys(data, "cap_reason_key", "conversion_ppm", "credited_resource_id", "payout_score_fact_id", "per_send_cap", "sends_per_day") ||
		declarations.ResourceIDs == nil || declarations.ScoreFactIDs == nil || declarations.CopyKeys == nil {
		return PayoutPolicy{}, ErrInvalidPayoutPolicy
	}
	var policy PayoutPolicy
	if decodeExact(data, &policy) != nil || !mechanicalPattern.MatchString(policy.CreditedResourceID) ||
		!mechanicalPattern.MatchString(policy.PayoutScoreFactID) || !mechanicalPattern.MatchString(policy.CapReasonKey) {
		return PayoutPolicy{}, ErrInvalidPayoutPolicy
	}
	if _, ok := declarations.ResourceIDs[policy.CreditedResourceID]; !ok || !validExactNonnegative(policy.SendsPerDay) ||
		!validExactNonnegative(policy.PerSendCap) || policy.ConversionPPM < 0 || policy.ConversionPPM > payoutPPM {
		return PayoutPolicy{}, ErrInvalidPayoutPolicy
	}
	if _, ok := declarations.ScoreFactIDs[policy.PayoutScoreFactID]; !ok {
		return PayoutPolicy{}, ErrInvalidPayoutPolicy
	}
	if _, ok := declarations.CopyKeys[policy.CapReasonKey]; !ok {
		return PayoutPolicy{}, ErrInvalidPayoutPolicy
	}
	return policy, nil
}

func SelectPayoutScore(result *Result, policy PayoutPolicy) (int64, error) {
	if !mechanicalPattern.MatchString(policy.PayoutScoreFactID) || result == nil || !validResult(result) {
		return 0, ErrInvalidPayoutPolicy
	}
	found := false
	var score int64
	for _, fact := range result.ScoreFacts {
		if fact.Kind != policy.PayoutScoreFactID {
			continue
		}
		if found || fact.Value < 0 {
			return 0, ErrInvalidPayoutPolicy
		}
		found, score = true, fact.Value
	}
	if !found {
		return 0, ErrInvalidPayoutPolicy
	}
	return score, nil
}

func ConvertPayout(score, rateReductionPPM, conversionPPM, priorRemainderPPM int64) (PayoutConversion, error) {
	if !validExactNonnegative(score) || rateReductionPPM < 0 || rateReductionPPM > payoutPPM ||
		conversionPPM < 0 || conversionPPM > payoutPPM || priorRemainderPPM < 0 || priorRemainderPPM >= payoutPPM {
		return PayoutConversion{}, ErrInvalidPayoutPolicy
	}
	reduced, err := fixedgrid.Integrate(score, payoutPPM-rateReductionPPM, 0, payoutPPM)
	if err != nil || !reduced.Whole.IsInt64() {
		return PayoutConversion{}, ErrInvalidPayoutPolicy
	}
	converted, err := fixedgrid.Integrate(reduced.Whole.Int64(), conversionPPM, priorRemainderPPM, payoutPPM)
	if err != nil || !converted.Whole.IsInt64() {
		return PayoutConversion{}, ErrInvalidPayoutPolicy
	}
	return PayoutConversion{
		ReducedScore: reduced.Whole.Int64(), ConvertedUnits: converted.Whole.Int64(),
		ConversionRemainderPPM: converted.Remainder,
	}, nil
}

func validExactNonnegative(value int64) bool {
	return value >= 0 && value <= decimal.MaxExactInteger
}

// MarshalJSON keeps policy rows on the exact C30 wire even if callers encode
// a loaded policy again for an artifact or session genesis.
func (policy PayoutPolicy) MarshalJSON() ([]byte, error) {
	type wire PayoutPolicy
	if !mechanicalPattern.MatchString(policy.CreditedResourceID) || !validExactNonnegative(policy.SendsPerDay) ||
		!validExactNonnegative(policy.PerSendCap) || policy.ConversionPPM < 0 || policy.ConversionPPM > payoutPPM ||
		!mechanicalPattern.MatchString(policy.PayoutScoreFactID) || !mechanicalPattern.MatchString(policy.CapReasonKey) {
		return nil, ErrInvalidPayoutPolicy
	}
	return json.Marshal(wire(policy))
}

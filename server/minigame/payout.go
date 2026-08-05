package minigame

import (
	"encoding/json"
	"errors"
	"math/big"

	"cloud-clicker/server/decimal"
)

var ErrInvalidPayoutPolicy = errors.New("invalid minigame payout policy")

const payoutPPM = int64(1_000_000)

type PayoutPolicy struct {
	CreditedResourceID string `json:"credited_resource_id"`
	SendsPerDay        int64  `json:"sends_per_day"`
	PerSendCap         int64  `json:"per_send_cap"`
	ConversionPPM      int64  `json:"conversion_ppm"`
}

type PayoutConversion struct {
	ReducedScore           int64
	ConvertedUnits         int64
	ConversionRemainderPPM int64
}

func LoadPayoutPolicy(data []byte, declaredResources map[string]struct{}) (PayoutPolicy, error) {
	if !hasExactJSONKeys(data, "conversion_ppm", "credited_resource_id", "per_send_cap", "sends_per_day") || declaredResources == nil {
		return PayoutPolicy{}, ErrInvalidPayoutPolicy
	}
	var policy PayoutPolicy
	if decodeExact(data, &policy) != nil || !mechanicalPattern.MatchString(policy.CreditedResourceID) {
		return PayoutPolicy{}, ErrInvalidPayoutPolicy
	}
	if _, ok := declaredResources[policy.CreditedResourceID]; !ok || !validExactNonnegative(policy.SendsPerDay) ||
		!validExactNonnegative(policy.PerSendCap) || policy.ConversionPPM < 0 || policy.ConversionPPM > payoutPPM {
		return PayoutPolicy{}, ErrInvalidPayoutPolicy
	}
	return policy, nil
}

func ConvertPayout(score, rateReductionPPM, conversionPPM, priorRemainderPPM int64) (PayoutConversion, error) {
	if !validExactNonnegative(score) || rateReductionPPM < 0 || rateReductionPPM > payoutPPM ||
		conversionPPM < 0 || conversionPPM > payoutPPM || priorRemainderPPM < 0 || priorRemainderPPM >= payoutPPM {
		return PayoutConversion{}, ErrInvalidPayoutPolicy
	}
	reducedNumerator := new(big.Int).Mul(big.NewInt(score), big.NewInt(payoutPPM-rateReductionPPM))
	reduced := new(big.Int).Quo(reducedNumerator, big.NewInt(payoutPPM))
	conversionNumerator := new(big.Int).Mul(reduced, big.NewInt(conversionPPM))
	conversionNumerator.Add(conversionNumerator, big.NewInt(priorRemainderPPM))
	converted, remainder := new(big.Int), new(big.Int)
	converted.QuoRem(conversionNumerator, big.NewInt(payoutPPM), remainder)
	return PayoutConversion{
		ReducedScore: reduced.Int64(), ConvertedUnits: converted.Int64(),
		ConversionRemainderPPM: remainder.Int64(),
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
		!validExactNonnegative(policy.PerSendCap) || policy.ConversionPPM < 0 || policy.ConversionPPM > payoutPPM {
		return nil, ErrInvalidPayoutPolicy
	}
	return json.Marshal(wire(policy))
}

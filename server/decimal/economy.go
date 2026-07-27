package decimal

import (
	"errors"
	"math"
)

var ErrInvalidEconomyInput = errors.New("invalid economy input")

// SumGeometricSeries returns the price of count purchases after owned existing
// purchases: base * ratio^owned * (1-ratio^count) / (1-ratio).
func SumGeometricSeries(count int64, base, ratio Decimal, owned int64) Decimal {
	if count < 0 || count > MaxExactInteger || owned < 0 || owned > MaxExactInteger ||
		!base.IsStateValue() || !ratio.IsStateValue() || base.Lt(Zero) || ratio.Lt(One) {
		return NaN
	}
	if count == 0 || base.Eq(Zero) {
		return Zero
	}
	if ratio.Eq(One) {
		return base.Mul(FromFloat64(float64(count)))
	}
	return base.
		Mul(ratio.Pow(float64(owned))).
		Mul(One.Sub(ratio.Pow(float64(count)))).
		Div(One.Sub(ratio))
}

// AffordGeometricSeries returns the exact, verified maximum affordable count.
// The closed-form inverse is only an estimate; the result is corrected until
// the two RFC-0001 affordability postconditions hold, with binary-search
// fallback when local correction is insufficient.
func AffordGeometricSeries(cash, base, ratio Decimal, owned int64) (int64, error) {
	if owned < 0 || owned > MaxExactInteger || !cash.IsStateValue() || !base.IsStateValue() ||
		!ratio.IsStateValue() || cash.Lt(Zero) || !base.Gt(Zero) || ratio.Lt(One) {
		return 0, ErrInvalidEconomyInput
	}
	if cash.Lt(SumGeometricSeries(1, base, ratio, owned)) {
		return 0, nil
	}
	if ratio.Eq(One) {
		candidate := cash.Div(base).Floor().toFloat64()
		if !isExactCount(candidate) {
			return MaxExactInteger, nil
		}
		return minInt64(int64(candidate), MaxExactInteger), nil
	}

	actualStart := base.Mul(ratio.Pow(float64(owned)))
	estimate := cash.
		Div(actualStart).
		Mul(ratio.Sub(One)).
		Add(One).
		Log10().
		Div(ratio.Log10()).
		Floor().
		toFloat64()

	if !isExactCount(estimate) {
		return binarySearchAffordable(cash, base, ratio, owned), nil
	}
	candidate := minInt64(int64(estimate), MaxExactInteger)

	for correction := 0; correction < 8 && candidate > 0 && !isAffordable(candidate, cash, base, ratio, owned); correction++ {
		candidate--
	}
	for correction := 0; correction < 8 && candidate < MaxExactInteger && isAffordable(candidate+1, cash, base, ratio, owned); correction++ {
		candidate++
	}
	if !affordabilityPostconditions(candidate, cash, base, ratio, owned) {
		candidate = binarySearchAffordable(cash, base, ratio, owned)
	}
	return candidate, nil
}

func isExactCount(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0) && value >= 0 && value <= float64(MaxExactInteger) && math.Trunc(value) == value
}

func isAffordable(count int64, cash, base, ratio Decimal, owned int64) bool {
	cost := SumGeometricSeries(count, base, ratio, owned)
	return cost.IsStateValue() && cost.Lte(cash)
}

func affordabilityPostconditions(count int64, cash, base, ratio Decimal, owned int64) bool {
	if !isAffordable(count, cash, base, ratio, owned) {
		return false
	}
	return count == MaxExactInteger || !isAffordable(count+1, cash, base, ratio, owned)
}

func binarySearchAffordable(cash, base, ratio Decimal, owned int64) int64 {
	if isAffordable(MaxExactInteger, cash, base, ratio, owned) {
		return MaxExactInteger
	}
	low, high := int64(0), MaxExactInteger
	for low < high {
		mid := low + (high-low+1)/2
		if isAffordable(mid, cash, base, ratio, owned) {
			low = mid
		} else {
			high = mid - 1
		}
	}
	return low
}

func minInt64(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

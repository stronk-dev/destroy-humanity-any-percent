// Package production implements shared closed-form production primitives.
package production

import (
	"errors"
	"sort"

	"cloud-clicker/server/decimal"
)

var ErrInvalidAccrual = errors.New("invalid production accrual input")

// AccrueConstant integrates non-negative per-second rates over an exact number
// of milliseconds. Sources are sorted on a copy so input order cannot affect
// the single authoritative quantization boundary.
func AccrueConstant(rates []decimal.Decimal, elapsedMilliseconds int64, efficiency decimal.Decimal) (decimal.Decimal, error) {
	if elapsedMilliseconds < 0 || elapsedMilliseconds > decimal.MaxExactInteger ||
		!efficiency.IsStateValue() || efficiency.Lt(decimal.Zero) {
		return decimal.NaN, ErrInvalidAccrual
	}
	ordered := append([]decimal.Decimal(nil), rates...)
	for _, rate := range ordered {
		if !rate.IsStateValue() || rate.Lt(decimal.Zero) {
			return decimal.NaN, ErrInvalidAccrual
		}
	}
	sort.Slice(ordered, func(left, right int) bool {
		if ordered[left].Exponent() != ordered[right].Exponent() {
			return ordered[left].Exponent() < ordered[right].Exponent()
		}
		return ordered[left].String() < ordered[right].String()
	})
	totalRate := decimal.Zero
	for _, rate := range ordered {
		totalRate = totalRate.Add(rate)
		if !totalRate.IsStateValue() {
			return decimal.NaN, ErrInvalidAccrual
		}
	}
	seconds := decimal.FromFloat64(float64(elapsedMilliseconds)).Div(decimal.FromFloat64(1000))
	delta := totalRate.Mul(seconds).Mul(efficiency).Quantize(decimal.CanonicalSignificantDigits)
	if !delta.IsStateValue() {
		return decimal.NaN, ErrInvalidAccrual
	}
	return delta, nil
}

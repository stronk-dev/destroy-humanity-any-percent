// Package production implements shared closed-form production primitives.
package production

import (
	"errors"

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
	for _, rate := range rates {
		if !rate.IsStateValue() || rate.Lt(decimal.Zero) {
			return decimal.NaN, ErrInvalidAccrual
		}
	}
	totalRate := decimal.SumDeterministic(rates)
	if !totalRate.IsStateValue() {
		return decimal.NaN, ErrInvalidAccrual
	}
	seconds := decimal.FromFloat64(float64(elapsedMilliseconds)).Div(decimal.FromFloat64(1000))
	delta := totalRate.Mul(seconds).Mul(efficiency).Quantize(decimal.CanonicalSignificantDigits)
	if !delta.IsStateValue() {
		return decimal.NaN, ErrInvalidAccrual
	}
	return delta, nil
}

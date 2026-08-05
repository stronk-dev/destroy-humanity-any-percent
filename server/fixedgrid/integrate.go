// Package fixedgrid owns the exact carried-remainder primitive shared by
// attended-time mechanics. Callers own saturation and reset policy.
package fixedgrid

import (
	"errors"
	"math/big"
)

const MaxExactInteger = int64(9_007_199_254_740_991)

var ErrInvalidInput = errors.New("invalid fixed-grid input")

type Result struct {
	Whole     *big.Int
	Remainder int64
}

// Integrate computes floor((units*rate+remainder)/divisor) without a binary
// float or fixed-width intermediate. Every persisted input remains inside the
// cross-runtime exact-integer domain; Whole stays wide until its owner applies
// a declared hardcap.
func Integrate(units, rate, remainder, divisor int64) (Result, error) {
	if units < 0 || units > MaxExactInteger || rate < 0 || rate > MaxExactInteger ||
		remainder < 0 || divisor < 1 || divisor > MaxExactInteger || remainder >= divisor {
		return Result{}, ErrInvalidInput
	}
	numerator := new(big.Int).Mul(big.NewInt(units), big.NewInt(rate))
	numerator.Add(numerator, big.NewInt(remainder))
	whole, carry := new(big.Int), new(big.Int)
	whole.QuoRem(numerator, big.NewInt(divisor), carry)
	return Result{Whole: whole, Remainder: carry.Int64()}, nil
}

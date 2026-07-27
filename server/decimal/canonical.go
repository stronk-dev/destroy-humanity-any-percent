package decimal

import (
	"errors"
	"math"
	"strconv"
	"strings"
)

const (
	// CanonicalSignificantDigits is the authoritative persisted/wire precision.
	CanonicalSignificantDigits = 12
	// MaxExactInteger is shared with JavaScript's Number.MAX_SAFE_INTEGER.
	MaxExactInteger int64 = 9_007_199_254_740_991
)

var ErrInvalidCanonical = errors.New("invalid canonical decimal")

// Quantize rounds a finite Decimal to significantDigits decimal digits using
// round-half-to-even. Non-finite diagnostic values propagate unchanged.
func (d Decimal) Quantize(significantDigits int) Decimal {
	d = d.Normalize()
	if !d.IsFinite() || d.mantissa == 0 {
		return d
	}
	if significantDigits < 1 || significantDigits > 15 {
		return NaN
	}

	factor := math.Pow10(significantDigits - 1)
	rounded := math.RoundToEven(math.Abs(d.mantissa)*factor) / factor
	exponent := d.exponent
	if rounded >= 10 {
		rounded = 1
		if exponent == maxExponent {
			return NaN
		}
		exponent++
	}
	return New(math.Copysign(rounded, d.mantissa), exponent)
}

// IsStateValue reports whether d is valid authoritative gameplay state.
func (d Decimal) IsStateValue() bool {
	return d.IsFinite() && d.exponent >= -maxExponent && d.exponent <= maxExponent
}

// ParseCanonical accepts only the RFC-0001 finite wire grammar. FromString is
// intentionally more permissive for config/import and diagnostic arithmetic.
func ParseCanonical(value string) (Decimal, error) {
	if value == "0" {
		return Zero, nil
	}
	if strings.TrimSpace(value) != value || strings.Contains(value, "E") {
		return NaN, ErrInvalidCanonical
	}
	parts := strings.Split(value, "e")
	if len(parts) != 2 || !validCanonicalCoefficient(parts[0]) || !validCanonicalExponent(parts[1]) {
		return NaN, ErrInvalidCanonical
	}

	mantissa, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return NaN, ErrInvalidCanonical
	}
	exponent, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil || exponent < -maxExponent || exponent > maxExponent {
		return NaN, ErrInvalidCanonical
	}
	d := New(mantissa, exponent)
	if !d.IsStateValue() || d.String() != value {
		return NaN, ErrInvalidCanonical
	}
	return d, nil
}

func validCanonicalCoefficient(value string) bool {
	unsigned := value
	if strings.HasPrefix(unsigned, "-") {
		unsigned = unsigned[1:]
	}
	if len(unsigned) == 0 || unsigned[0] < '1' || unsigned[0] > '9' {
		return false
	}
	if len(unsigned) == 1 {
		return true
	}
	if len(unsigned) < 3 || unsigned[1] != '.' {
		return false
	}
	digits := unsigned[2:]
	if len(digits) < 1 || len(digits) > CanonicalSignificantDigits-1 || digits[len(digits)-1] == '0' {
		return false
	}
	for _, digit := range digits {
		if digit < '0' || digit > '9' {
			return false
		}
	}
	return true
}

func validCanonicalExponent(value string) bool {
	if value == "0" {
		return true
	}
	unsigned := value
	if strings.HasPrefix(unsigned, "-") {
		unsigned = unsigned[1:]
	}
	if len(unsigned) == 0 || unsigned[0] == '0' || strings.HasPrefix(value, "+") {
		return false
	}
	for _, digit := range unsigned {
		if digit < '0' || digit > '9' {
			return false
		}
	}
	return true
}

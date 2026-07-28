// Package decimal implements the layer-0 numeric core specified by RFC-0001.
//
// Decimal ports break_infinity.js 2.2.0's mantissa/exponent arithmetic. The
// JavaScript package remains the golden-vector oracle; this implementation is
// deliberately small enough to audit operation-by-operation against it.
package decimal

import (
	"math"
	"strconv"
	"strings"
)

const (
	maxExponent  = int64(8_999_999_999_999_999)
	jsMaxInteger = float64(9_007_199_254_740_991)
)

// Decimal stores a normalized signed mantissa and a base-10 exponent. A finite,
// non-zero Decimal has |mantissa| in [1, 10). Special values are diagnostic
// only and are never valid authoritative state.
type Decimal struct {
	mantissa float64
	exponent int64
}

var (
	Zero   = Decimal{}
	One    = Decimal{mantissa: 1}
	NaN    = Decimal{mantissa: math.NaN()}
	Inf    = Decimal{mantissa: math.Inf(1)}
	NegInf = Decimal{mantissa: math.Inf(-1)}
)

// New constructs and normalizes a Decimal from mantissa/exponent components.
func New(mantissa float64, exponent int64) Decimal {
	return Decimal{mantissa: mantissa, exponent: exponent}.Normalize()
}

// Mantissa returns the normalized signed mantissa.
func (d Decimal) Mantissa() float64 { return d.mantissa }

// Exponent returns the normalized base-10 exponent.
func (d Decimal) Exponent() int64 { return d.exponent }

// Normalize returns d in canonical mantissa/exponent form, following
// break_infinity.js's normalization order.
func (d Decimal) Normalize() Decimal {
	if math.IsNaN(d.mantissa) || math.IsInf(d.mantissa, 0) {
		d.exponent = 0
		return d
	}
	if d.mantissa >= 1 && d.mantissa < 10 {
		if !validExponent(d.exponent) {
			return NaN
		}
		return d
	}
	if d.mantissa == 0 {
		return Zero
	}

	tempExponent := math.Floor(math.Log10(math.Abs(d.mantissa)))
	if math.IsNaN(tempExponent) || math.IsInf(tempExponent, 0) || tempExponent < math.MinInt64 || tempExponent > math.MaxInt64 {
		return NaN
	}
	shift := int64(tempExponent)
	if shift == -324 {
		d.mantissa = 10 * d.mantissa / 1e-323
	} else {
		d.mantissa /= math.Pow10(int(shift))
	}
	if shift > 0 && d.exponent > math.MaxInt64-shift || shift < 0 && d.exponent < math.MinInt64-shift {
		return NaN
	}
	d.exponent += shift
	// Log10/Pow10 can disagree by one ULP at exact powers of ten. In that
	// case the first scaling pass leaves a boundary mantissa of 10 (or,
	// symmetrically, one just below 1). Correct the carry before validating
	// the representation and exponent domain.
	magnitude := math.Abs(d.mantissa)
	if magnitude >= 10 {
		d.mantissa /= 10
		if d.exponent == math.MaxInt64 {
			return NaN
		}
		d.exponent++
	} else if magnitude < 1 {
		d.mantissa *= 10
		if d.exponent == math.MinInt64 {
			return NaN
		}
		d.exponent--
	}
	if !validExponent(d.exponent) || math.IsNaN(d.mantissa) || math.IsInf(d.mantissa, 0) {
		return NaN
	}
	return d
}

func validExponent(exponent int64) bool {
	return exponent >= -maxExponent && exponent <= maxExponent
}

// FromFloat64 converts an IEEE-754 float64 to a normalized Decimal.
func FromFloat64(value float64) Decimal {
	if math.IsNaN(value) {
		return NaN
	}
	if math.IsInf(value, 1) {
		return Inf
	}
	if math.IsInf(value, -1) {
		return NegInf
	}
	if value == 0 {
		return Zero
	}
	exponent := math.Floor(math.Log10(math.Abs(value)))
	var mantissa float64
	if exponent == -324 {
		mantissa = 10 * value / 1e-323
	} else {
		mantissa = value / math.Pow10(int(exponent))
	}
	return New(mantissa, int64(exponent))
}

// FromString parses decimal input without panicking. Canonical scientific
// strings take the direct mantissa/exponent path used by break_infinity.js.
func FromString(value string) Decimal {
	value = strings.TrimSpace(value)
	switch value {
	case "NaN", "nan":
		return NaN
	case "Infinity", "+Infinity", "infinity", "+infinity":
		return Inf
	case "-Infinity", "-infinity":
		return NegInf
	}

	if strings.Count(strings.ToLower(value), "e") == 1 {
		parts := strings.FieldsFunc(value, func(r rune) bool { return r == 'e' || r == 'E' })
		if len(parts) != 2 {
			return NaN
		}
		mantissa, mantissaErr := strconv.ParseFloat(parts[0], 64)
		exponentFloat, exponentErr := strconv.ParseFloat(parts[1], 64)
		if mantissaErr != nil || exponentErr != nil || math.IsNaN(exponentFloat) || math.IsInf(exponentFloat, 0) || math.Trunc(exponentFloat) != exponentFloat || math.Abs(exponentFloat) > float64(maxExponent) {
			return NaN
		}
		return New(mantissa, int64(exponentFloat))
	}

	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return NaN
	}
	return FromFloat64(parsed)
}

// String returns the RFC-0001 canonical wire representation for finite values.
// Diagnostic non-finite values retain stable tokens, but IsStateValue rejects
// them and they must never be persisted or sent as gameplay state.
func (d Decimal) String() string {
	if d.IsNaN() {
		return "NaN"
	}
	if math.IsInf(d.mantissa, 1) {
		return "Infinity"
	}
	if math.IsInf(d.mantissa, -1) {
		return "-Infinity"
	}
	d = d.Quantize(CanonicalSignificantDigits)
	if d.IsNaN() {
		return "NaN"
	}
	if d.mantissa == 0 {
		return "0"
	}
	coefficient := strconv.FormatFloat(d.mantissa, 'f', CanonicalSignificantDigits-1, 64)
	coefficient = strings.TrimRight(strings.TrimRight(coefficient, "0"), ".")
	return coefficient + "e" + strconv.FormatInt(d.exponent, 10)
}

func formatNumber(value float64) string {
	if math.IsNaN(value) {
		return "NaN"
	}
	if math.IsInf(value, 1) {
		return "Infinity"
	}
	if math.IsInf(value, -1) {
		return "-Infinity"
	}
	return strconv.FormatFloat(value, 'f', -1, 64)
}

// IsNaN reports whether d is not a number.
func (d Decimal) IsNaN() bool { return math.IsNaN(d.mantissa) }

// IsFinite reports whether d is neither NaN nor an infinity.
func (d Decimal) IsFinite() bool {
	return !math.IsNaN(d.mantissa) && !math.IsInf(d.mantissa, 0)
}

// Neg returns -d.
func (d Decimal) Neg() Decimal {
	return Decimal{mantissa: -d.mantissa, exponent: d.exponent}
}

// Abs returns |d|.
func (d Decimal) Abs() Decimal {
	return Decimal{mantissa: math.Abs(d.mantissa), exponent: d.exponent}
}

// Add returns d + other with break_infinity.js's 14-digit alignment step.
func (d Decimal) Add(other Decimal) Decimal {
	if d.IsNaN() || other.IsNaN() {
		return NaN
	}
	if math.IsInf(d.mantissa, 0) || math.IsInf(other.mantissa, 0) {
		if math.IsInf(d.mantissa, 0) && math.IsInf(other.mantissa, 0) && math.Signbit(d.mantissa) != math.Signbit(other.mantissa) {
			return Zero
		}
		if math.IsInf(d.mantissa, 0) {
			return d
		}
		return other
	}
	if d.mantissa == 0 {
		return other
	}
	if other.mantissa == 0 {
		return d
	}

	bigger, smaller := d, other
	if other.exponent > d.exponent {
		bigger, smaller = other, d
	}
	if bigger.exponent-smaller.exponent > 17 {
		return bigger
	}
	aligned := 1e14*bigger.mantissa + 1e14*smaller.mantissa*math.Pow10(int(smaller.exponent-bigger.exponent))
	return New(jsRound(aligned), bigger.exponent-14)
}

func jsRound(value float64) float64 { return math.Floor(value + 0.5) }

// Sub returns d - other.
func (d Decimal) Sub(other Decimal) Decimal { return d.Add(other.Neg()) }

// Mul returns d * other.
func (d Decimal) Mul(other Decimal) Decimal {
	if d.IsNaN() || other.IsNaN() {
		return NaN
	}
	if (math.IsInf(d.mantissa, 0) && other.mantissa == 0) || (d.mantissa == 0 && math.IsInf(other.mantissa, 0)) {
		return Zero
	}
	if math.IsInf(d.mantissa, 0) || math.IsInf(other.mantissa, 0) {
		return Decimal{mantissa: math.Copysign(math.Inf(1), d.mantissa*other.mantissa)}
	}
	return arithmeticResult(d.mantissa*other.mantissa, d.exponent+other.exponent)
}

// Div computes the final quotient directly. Building a reciprocal first can
// leave the valid exponent domain transiently even when the quotient is valid.
func (d Decimal) Div(other Decimal) Decimal {
	if d.IsNaN() || other.IsNaN() {
		return NaN
	}
	if other.mantissa == 0 || math.IsInf(other.mantissa, 0) {
		return Zero
	}
	if math.IsInf(d.mantissa, 0) {
		return Decimal{mantissa: math.Copysign(math.Inf(1), d.mantissa*other.mantissa)}
	}
	return arithmeticResult(d.mantissa/other.mantissa, d.exponent-other.exponent)
}

func arithmeticResult(mantissa float64, exponent int64) Decimal {
	magnitude := math.Abs(mantissa)
	if magnitude >= 10 {
		mantissa /= 10
		exponent++
	} else if magnitude > 0 && magnitude < 1 {
		mantissa *= 10
		exponent--
	}
	if exponent > maxExponent {
		return Decimal{mantissa: math.Copysign(math.Inf(1), mantissa)}
	}
	if exponent < -maxExponent {
		return Zero
	}
	return New(mantissa, exponent)
}

// Log10 returns the base-10 logarithm as a Decimal.
func (d Decimal) Log10() Decimal {
	if d.IsNaN() || d.mantissa < 0 {
		return NaN
	}
	if d.mantissa == 0 {
		return NegInf
	}
	return FromFloat64(float64(d.exponent) + math.Log10(d.mantissa))
}

// Ln returns the natural logarithm as a Decimal.
func (d Decimal) Ln() Decimal {
	if d.IsNaN() || d.mantissa < 0 {
		return NaN
	}
	if d.mantissa == 0 {
		return NegInf
	}
	return FromFloat64(2.302585092994045 * (float64(d.exponent) + math.Log10(d.mantissa)))
}

func pow10(value float64) Decimal {
	if math.IsNaN(value) {
		return NaN
	}
	if math.IsInf(value, 1) || value > float64(maxExponent) {
		return Inf
	}
	if math.IsInf(value, -1) || value < -float64(maxExponent) {
		return Zero
	}
	if math.Trunc(value) == value {
		return Decimal{mantissa: 1, exponent: int64(value)}
	}
	exponent := math.Trunc(value)
	return New(math.Pow(10, math.Mod(value, 1)), int64(exponent))
}

// Pow returns d raised to a float64 power.
func (d Decimal) Pow(power float64) Decimal {
	if math.IsNaN(power) || d.IsNaN() {
		return NaN
	}
	if d.mantissa == 0 {
		if power == 0 {
			return One
		}
		return Zero
	}
	if math.Trunc(power) == power {
		exactExponentProduct := float64(d.exponent) * power
		if math.Abs(exactExponentProduct) <= float64(maxExponent+1) {
			mantissa := math.Pow(d.mantissa, power)
			if !math.IsNaN(mantissa) && !math.IsInf(mantissa, 0) && mantissa != 0 {
				return arithmeticResult(mantissa, int64(exactExponentProduct))
			}
		}
	}
	exponentProduct := float64(d.exponent) * power
	if math.Trunc(exponentProduct) == exponentProduct && math.Abs(exponentProduct) <= jsMaxInteger {
		mantissa := math.Pow(d.mantissa, power)
		if !math.IsNaN(mantissa) && !math.IsInf(mantissa, 0) && mantissa != 0 && math.Abs(exponentProduct) <= float64(maxExponent) {
			return New(mantissa, int64(exponentProduct))
		}
	}

	integerPart := math.Trunc(exponentProduct)
	fractionalPart := exponentProduct - integerPart
	mantissa := math.Pow(10, power*math.Log10(d.mantissa)+fractionalPart)
	if !math.IsNaN(mantissa) && !math.IsInf(mantissa, 0) && mantissa != 0 && math.Abs(integerPart) <= float64(maxExponent) {
		return New(mantissa, int64(integerPart))
	}

	result := pow10(power * (float64(d.exponent) + math.Log10(math.Abs(d.mantissa))))
	if d.mantissa < 0 {
		mod := math.Abs(math.Mod(power, 2))
		if mod == 1 {
			return result.Neg()
		}
		if mod != 0 {
			return NaN
		}
	}
	return result
}

// Exp returns e^d.
func (d Decimal) Exp() Decimal {
	if d.IsNaN() {
		return NaN
	}
	value := d.toFloat64()
	if value > -706 && value < 709 {
		return FromFloat64(math.Exp(value))
	}
	return FromFloat64(math.E).Pow(value)
}

// Floor returns the greatest integer not greater than d.
func (d Decimal) Floor() Decimal {
	if d.IsNaN() || math.IsInf(d.mantissa, 0) {
		return d
	}
	if d.exponent < -1 {
		if d.mantissa >= 0 {
			return Zero
		}
		return FromFloat64(-1)
	}
	if d.exponent < 17 {
		return FromFloat64(math.Floor(d.toFloat64()))
	}
	return d
}

func (d Decimal) toFloat64() float64 {
	if d.IsNaN() || math.IsInf(d.mantissa, 0) {
		return d.mantissa
	}
	if d.exponent > 308 {
		return math.Copysign(math.Inf(1), d.mantissa)
	}
	if d.exponent < -324 {
		return math.Copysign(0, d.mantissa)
	}
	if d.exponent == -324 {
		return math.Copysign(5e-324, d.mantissa)
	}
	value := d.mantissa * math.Pow10(int(d.exponent))
	if math.IsInf(value, 0) || d.exponent < 0 {
		return value
	}
	rounded := jsRound(value)
	if math.Abs(rounded-value) < 1e-10 {
		return rounded
	}
	return value
}

// Cmp returns -1, 0, or 1 according to d's ordering relative to other. It
// returns NaN when either operand is NaN.
func (d Decimal) Cmp(other Decimal) float64 {
	if d.IsNaN() || other.IsNaN() {
		return math.NaN()
	}
	if d.mantissa == 0 {
		if other.mantissa < 0 {
			return 1
		}
		if other.mantissa > 0 {
			return -1
		}
		return 0
	}
	if other.mantissa == 0 {
		if d.mantissa < 0 {
			return -1
		}
		return 1
	}
	if d.mantissa > 0 {
		if other.mantissa < 0 || d.exponent > other.exponent {
			return 1
		}
		if d.exponent < other.exponent {
			return -1
		}
	} else {
		if other.mantissa > 0 || d.exponent > other.exponent {
			return -1
		}
		if d.exponent < other.exponent {
			return 1
		}
	}
	if d.mantissa < other.mantissa {
		return -1
	}
	if d.mantissa > other.mantissa {
		return 1
	}
	return 0
}

// Eq reports exact normalized mantissa/exponent equality.
func (d Decimal) Eq(other Decimal) bool {
	return d.mantissa == other.mantissa && d.exponent == other.exponent
}

func (d Decimal) Lt(other Decimal) bool  { return d.Cmp(other) == -1 }
func (d Decimal) Lte(other Decimal) bool { return !d.Gt(other) }
func (d Decimal) Gt(other Decimal) bool  { return d.Cmp(other) == 1 }
func (d Decimal) Gte(other Decimal) bool { return !d.Lt(other) }

func (d Decimal) Max(other Decimal) Decimal {
	if d.Lt(other) {
		return other
	}
	return d
}

func (d Decimal) Min(other Decimal) Decimal {
	if d.Gt(other) {
		return other
	}
	return d
}

func Max(a, b Decimal) Decimal { return a.Max(b) }
func Min(a, b Decimal) Decimal { return a.Min(b) }

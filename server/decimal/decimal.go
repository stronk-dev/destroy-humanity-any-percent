// Package decimal implements the layer-0 numeric core specified by RFC-0001.
//
// Decimal deliberately follows break_eternity.js normalization and operation
// ordering within the break_infinity-compatible range. Agreement with the
// JavaScript reference is more important here than numerically prettier code.
package decimal

import (
	"math"
	"strconv"
	"strings"
)

const (
	exponentLimit       = 9e15
	firstNegativeLayer  = 1 / exponentLimit
	layerDown           = 15.954242509439325
	maxSignificantDigit = 17
)

// Decimal stores a normalized signed mantissa and a base-10 exponent. A finite,
// non-zero Decimal has |mantissa| in [1, 10). Special values are stored in the
// mantissa with exponent zero.
type Decimal struct {
	mantissa float64
	exponent int64
}

type component struct {
	sign  float64
	layer int
	mag   float64
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

// Normalize returns d in canonical mantissa/exponent form.
func (d Decimal) Normalize() Decimal {
	if math.IsNaN(d.mantissa) || math.IsInf(d.mantissa, 0) {
		d.exponent = 0
		return d
	}
	if d.mantissa == 0 {
		return Zero
	}
	if math.Abs(d.mantissa) >= 1 && math.Abs(d.mantissa) < 10 {
		return d
	}

	tempExponent := math.Floor(math.Log10(math.Abs(d.mantissa)))
	if tempExponent < math.MinInt64 || tempExponent > math.MaxInt64 {
		return NaN
	}
	shift := int64(tempExponent)
	if shift == -324 {
		d.mantissa = d.mantissa * 10 / 1e-323
	} else {
		d.mantissa /= math.Pow10(int(shift))
	}
	if shift > 0 && d.exponent > math.MaxInt64-shift || shift < 0 && d.exponent < math.MinInt64-shift {
		return NaN
	}
	d.exponent += shift
	if math.Abs(d.mantissa) >= 10 {
		d.mantissa /= 10
		d.exponent++
	}
	return d
}

// FromFloat64 converts a float using the same component normalization order as
// break_eternity.js.
func FromFloat64(value float64) Decimal {
	if math.IsNaN(value) {
		return NaN
	}
	if math.IsInf(value, 0) {
		return Decimal{mantissa: value}
	}
	if value == 0 {
		return Zero
	}
	return fromComponent(math.Copysign(1, value), 0, math.Abs(value))
}

// FromString parses the canonical wire strings emitted by break_eternity.js.
// Invalid or out-of-scope layer-2 values become NaN; parsing never panics.
func FromString(value string) Decimal {
	value = strings.TrimSpace(strings.ToLower(strings.Replace(value, ",", "", 1)))
	switch value {
	case "nan":
		return NaN
	case "infinity", "+infinity":
		return Inf
	case "-infinity":
		return NegInf
	}

	parts := strings.Split(value, "e")
	if len(parts) == 1 {
		parsed, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return NaN
		}
		return FromFloat64(parsed)
	}
	if len(parts) != 2 {
		return NaN
	}
	// Canonical wire values already carry break_eternity's rounded mantissa and
	// integer exponent. Preserve those bits instead of re-running transcendental
	// normalization through a different platform libm.
	if mantissa, errMantissa := strconv.ParseFloat(parts[0], 64); errMantissa == nil && math.Abs(mantissa) >= 1 && math.Abs(mantissa) < 10 {
		if exponent, errExponent := strconv.ParseInt(parts[1], 10, 64); errExponent == nil {
			return New(mantissa, exponent)
		}
	}

	if parsed, err := strconv.ParseFloat(value, 64); err == nil && !math.IsNaN(parsed) && !math.IsInf(parsed, 0) && math.Abs(parsed) > 1e-307 {
		return FromFloat64(parsed)
	}

	mantissa, errMantissa := strconv.ParseFloat(parts[0], 64)
	exponent, errExponent := strconv.ParseFloat(parts[1], 64)
	if errMantissa != nil || errExponent != nil || mantissa == 0 || math.IsNaN(exponent) || math.IsInf(exponent, 0) {
		if mantissa == 0 && errMantissa == nil && errExponent == nil {
			return Zero
		}
		return NaN
	}
	return fromComponent(math.Copysign(1, mantissa), 1, exponent+math.Log10(math.Abs(mantissa)))
}

func fromFiniteNumber(value float64) Decimal {
	if value == 0 {
		return Zero
	}
	exponent := math.Floor(math.Log10(math.Abs(value)))
	if exponent < math.MinInt64 || exponent > math.MaxInt64 {
		return NaN
	}
	e := int64(exponent)
	var mantissa float64
	if e == -324 {
		mantissa = value * 10 / 1e-323
	} else {
		mantissa = value / math.Pow10(int(e))
	}
	return New(mantissa, e)
}

func fromComponent(sign float64, layer int, mag float64) Decimal {
	if math.IsNaN(sign) || math.IsNaN(mag) {
		return NaN
	}
	if sign == 0 || layer == 0 && mag == 0 || layer > 0 && math.IsInf(mag, -1) {
		return Zero
	}
	if layer == 0 && mag < 0 {
		mag = -mag
		sign = -sign
	}
	if math.IsInf(mag, 0) {
		return Decimal{mantissa: math.Copysign(math.Inf(1), sign)}
	}
	if layer == 0 && mag < firstNegativeLayer {
		layer++
		mag = math.Log10(mag)
	}

	absMag := math.Abs(mag)
	signMag := math.Copysign(1, mag)
	if absMag >= exponentLimit {
		layer++
		mag = signMag * math.Log10(absMag)
	} else {
		for absMag < layerDown && layer > 0 {
			layer--
			if layer == 0 {
				mag = math.Pow(10, mag)
			} else {
				mag = signMag * math.Pow(10, absMag)
				absMag = math.Abs(mag)
				signMag = math.Copysign(1, mag)
			}
		}
	}

	if layer < 0 || layer > 1 || math.IsNaN(mag) {
		return NaN
	}
	if layer == 0 {
		if mag < 0 {
			mag = -mag
			sign = -sign
		}
		return fromFiniteNumber(sign * mag)
	}

	exponent := math.Floor(mag)
	if exponent < math.MinInt64 || exponent > math.MaxInt64 {
		return NaN
	}
	mantissa := sign * math.Pow(10, mag-exponent)
	return New(mantissa, int64(exponent))
}

func (d Decimal) components() component {
	if d.IsNaN() {
		return component{sign: math.NaN(), layer: -1, mag: math.NaN()}
	}
	if math.IsInf(d.mantissa, 0) {
		return component{sign: math.Copysign(1, d.mantissa), layer: -1, mag: math.Inf(1)}
	}
	if d.mantissa == 0 {
		return component{}
	}

	sign := math.Copysign(1, d.mantissa)
	if d.exponent >= -324 && d.exponent <= 308 {
		value := math.Abs(scaledFloat(d.mantissa, d.exponent))
		if value >= firstNegativeLayer && value < exponentLimit {
			return component{sign: sign, mag: value}
		}
	}
	return component{sign: sign, layer: 1, mag: float64(d.exponent) + math.Log10(math.Abs(d.mantissa))}
}

func scaledFloat(mantissa float64, exponent int64) float64 {
	value, err := strconv.ParseFloat(
		strconv.FormatFloat(mantissa, 'f', -1, 64)+"e"+strconv.FormatInt(exponent, 10),
		64,
	)
	if err != nil {
		if exponent > 0 {
			return math.Copysign(math.Inf(1), mantissa)
		}
		return math.Copysign(0, mantissa)
	}
	return value
}

// String returns the canonical break_eternity.js wire representation.
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
	c := d.components()
	if c.layer == 0 && (c.mag < 1e21 && c.mag > 1e-7 || c.mag == 0) {
		return formatNumber(c.sign * c.mag)
	}

	return formatNumber(d.mantissa) + "e" + strconv.FormatInt(d.exponent, 10)
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
	if value == 0 {
		return "0"
	}
	abs := math.Abs(value)
	if abs >= 1e21 || abs < 1e-6 {
		scientific := strconv.FormatFloat(value, 'e', -1, 64)
		parts := strings.Split(scientific, "e")
		exponent, err := strconv.ParseInt(parts[1], 10, 64)
		if err == nil {
			sign := ""
			if exponent >= 0 {
				sign = "+"
			}
			return parts[0] + "e" + sign + strconv.FormatInt(exponent, 10)
		}
	}
	return strconv.FormatFloat(value, 'f', -1, 64)
}

// IsNaN reports whether d is not a number.
func (d Decimal) IsNaN() bool { return math.IsNaN(d.mantissa) }

// IsFinite reports whether d is neither NaN nor an infinity.
func (d Decimal) IsFinite() bool { return !math.IsNaN(d.mantissa) && !math.IsInf(d.mantissa, 0) }

// Neg returns -d.
func (d Decimal) Neg() Decimal { return Decimal{mantissa: -d.mantissa, exponent: d.exponent} }

// Abs returns |d|.
func (d Decimal) Abs() Decimal { return Decimal{mantissa: math.Abs(d.mantissa), exponent: d.exponent} }

// Add returns d + other using the JavaScript operation order.
func (d Decimal) Add(other Decimal) Decimal {
	if d.IsNaN() || other.IsNaN() {
		return NaN
	}
	if math.IsInf(d.mantissa, 0) || math.IsInf(other.mantissa, 0) {
		if math.IsInf(d.mantissa, 0) && math.IsInf(other.mantissa, 0) && math.Signbit(d.mantissa) != math.Signbit(other.mantissa) {
			return NaN
		}
		if math.IsInf(d.mantissa, 0) {
			return d
		}
		return other
	}

	aDecimal, bDecimal := d, other
	a, b := aDecimal.components(), bDecimal.components()
	if a.sign == 0 {
		return other
	}
	if b.sign == 0 {
		return d
	}
	if a.sign == -b.sign && a.layer == b.layer && a.mag == b.mag {
		return Zero
	}
	if cmpAbsComponent(a, b) < 0 {
		a, b = b, a
		aDecimal, bDecimal = bDecimal, aDecimal
	}
	if a.layer == 0 && b.layer == 0 {
		return FromFloat64(a.sign*a.mag + b.sign*b.mag)
	}

	layerA := a.layer * int(math.Copysign(1, a.mag))
	layerB := b.layer * int(math.Copysign(1, b.mag))
	if layerA-layerB >= 2 {
		return aDecimal
	}
	if layerA == 0 && layerB == -1 {
		if math.Abs(b.mag-math.Log10(a.mag)) > maxSignificantDigit {
			return aDecimal
		}
		magDiff := math.Pow(10, math.Log10(a.mag)-b.mag)
		mantissa := b.sign + a.sign*magDiff
		return fromComponent(math.Copysign(1, mantissa), 1, b.mag+math.Log10(math.Abs(mantissa)))
	}
	if layerA == 1 && layerB == 0 {
		if math.Abs(a.mag-math.Log10(b.mag)) > maxSignificantDigit {
			return aDecimal
		}
		magDiff := math.Pow(10, a.mag-math.Log10(b.mag))
		mantissa := b.sign + a.sign*magDiff
		return fromComponent(math.Copysign(1, mantissa), 1, math.Log10(b.mag)+math.Log10(math.Abs(mantissa)))
	}
	if math.Abs(a.mag-b.mag) > maxSignificantDigit {
		return aDecimal
	}
	magDiff := math.Pow(10, a.mag-b.mag)
	mantissa := b.sign + a.sign*magDiff
	return fromComponent(math.Copysign(1, mantissa), 1, b.mag+math.Log10(math.Abs(mantissa)))
}

// Sub returns d - other.
func (d Decimal) Sub(other Decimal) Decimal { return d.Add(other.Neg()) }

// Mul returns d * other using the JavaScript operation order.
func (d Decimal) Mul(other Decimal) Decimal {
	if d.IsNaN() || other.IsNaN() {
		return NaN
	}
	if (math.IsInf(d.mantissa, 0) && other.mantissa == 0) || (d.mantissa == 0 && math.IsInf(other.mantissa, 0)) {
		return NaN
	}
	if math.IsInf(d.mantissa, 0) || math.IsInf(other.mantissa, 0) {
		sign := math.Copysign(1, d.mantissa) * math.Copysign(1, other.mantissa)
		return Decimal{mantissa: math.Copysign(math.Inf(1), sign)}
	}

	aDecimal, bDecimal := d, other
	a, b := aDecimal.components(), bDecimal.components()
	if a.sign == 0 || b.sign == 0 {
		return Zero
	}
	if a.layer == b.layer && a.mag == -b.mag {
		return fromComponent(a.sign*b.sign, 0, 1)
	}
	if a.layer < b.layer || a.layer == b.layer && math.Abs(a.mag) < math.Abs(b.mag) {
		a, b = b, a
		aDecimal, bDecimal = bDecimal, aDecimal
	}
	if a.layer == 0 && b.layer == 0 {
		return FromFloat64(a.sign * b.sign * a.mag * b.mag)
	}
	if a.layer-b.layer >= 2 {
		if b.sign < 0 {
			return aDecimal.Neg()
		}
		return aDecimal
	}
	if a.layer == 1 && b.layer == 0 {
		return fromComponent(a.sign*b.sign, 1, a.mag+math.Log10(b.mag))
	}
	if a.layer == 1 && b.layer == 1 {
		return fromComponent(a.sign*b.sign, 1, a.mag+b.mag)
	}
	return NaN
}

func (d Decimal) reciprocal() Decimal {
	if d.IsNaN() || d.mantissa == 0 {
		return NaN
	}
	if math.IsInf(d.mantissa, 0) {
		return Zero
	}
	c := d.components()
	if c.layer == 0 {
		return fromComponent(c.sign, 0, 1/c.mag)
	}
	return fromComponent(c.sign, c.layer, -c.mag)
}

// Div returns d / other.
func (d Decimal) Div(other Decimal) Decimal { return d.Mul(other.reciprocal()) }

// Log10 returns the base-10 logarithm as a Decimal.
func (d Decimal) Log10() Decimal {
	if d.IsNaN() || d.mantissa <= 0 {
		return NaN
	}
	c := d.components()
	if c.layer > 0 {
		return fromComponent(math.Copysign(1, c.mag), c.layer-1, math.Abs(c.mag))
	}
	return fromComponent(1, 0, math.Log10(c.mag))
}

// Ln returns the natural logarithm as a Decimal.
func (d Decimal) Ln() Decimal {
	if d.IsNaN() || d.mantissa <= 0 {
		return NaN
	}
	c := d.components()
	if c.layer == 0 {
		return fromComponent(1, 0, math.Log(c.mag))
	}
	return fromComponent(math.Copysign(1, c.mag), 0, math.Abs(c.mag)*2.302585092994046)
}

func (d Decimal) pow10() Decimal {
	if d.IsNaN() {
		return NaN
	}
	if math.IsInf(d.mantissa, 1) {
		return Inf
	}
	if math.IsInf(d.mantissa, -1) {
		return Zero
	}
	x := d.toFloat64()
	if math.IsInf(x, 0) || math.IsNaN(x) {
		return NaN
	}
	return fromComponent(1, 1, x)
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
	if d.Eq(One) || power == 0 {
		return One
	}
	if power == 1 {
		return d
	}

	result := d.Abs().Log10().Mul(FromFloat64(power)).pow10()
	if d.mantissa < 0 {
		mod := math.Abs(math.Mod(power, 2))
		if mod == 1 {
			return result.Neg()
		}
		if mod == 0 {
			return result
		}
		return NaN
	}
	return result
}

// Exp returns e^d.
func (d Decimal) Exp() Decimal {
	if d.IsNaN() {
		return NaN
	}
	c := d.components()
	if c.layer > 0 && c.mag < 0 {
		return One
	}
	if c.layer == 0 && c.mag <= 709.7 {
		return FromFloat64(math.Exp(c.sign * c.mag))
	}
	if c.layer == 0 {
		return fromComponent(1, 1, c.sign*math.Log10(math.E)*c.mag)
	}
	return NaN
}

// Floor returns the greatest integer not greater than d.
func (d Decimal) Floor() Decimal {
	if d.IsNaN() || math.IsInf(d.mantissa, 0) {
		return d
	}
	c := d.components()
	if c.mag < 0 {
		if c.sign < 0 {
			return FromFloat64(-1)
		}
		return Zero
	}
	if c.sign < 0 {
		return d.Neg().ceil().Neg()
	}
	if c.layer == 0 {
		return FromFloat64(math.Floor(c.mag))
	}
	return d
}

func (d Decimal) ceil() Decimal {
	c := d.components()
	if c.mag < 0 {
		if c.sign > 0 {
			return One
		}
		return Zero
	}
	if c.sign < 0 {
		return d.Neg().Floor().Neg()
	}
	if c.layer == 0 {
		return FromFloat64(math.Ceil(c.mag))
	}
	return d
}

func (d Decimal) toFloat64() float64 {
	if d.IsNaN() || math.IsInf(d.mantissa, 0) {
		return d.mantissa
	}
	c := d.components()
	if c.layer == 0 {
		return c.sign * c.mag
	}
	if c.mag > 308 {
		return math.Copysign(math.Inf(1), c.sign)
	}
	if c.mag < -324 {
		return math.Copysign(0, c.sign)
	}
	return c.sign * math.Pow(10, c.mag)
}

func cmpAbsComponent(a, b component) int {
	layerA := a.layer
	if a.mag < 0 {
		layerA = -layerA
	}
	layerB := b.layer
	if b.mag < 0 {
		layerB = -layerB
	}
	if layerA < layerB {
		return -1
	}
	if layerA > layerB {
		return 1
	}
	if a.mag < b.mag {
		return -1
	}
	if a.mag > b.mag {
		return 1
	}
	return 0
}

// Cmp returns -1, 0, or 1 according to d's ordering relative to other. It
// returns NaN when either operand is NaN, matching the JavaScript reference.
func (d Decimal) Cmp(other Decimal) float64 {
	a, b := d.components(), other.components()
	if math.IsNaN(a.sign) || math.IsNaN(b.sign) {
		return math.NaN()
	}
	if a.sign < b.sign {
		return -1
	}
	if a.sign > b.sign {
		return 1
	}
	return a.sign * float64(cmpAbsComponent(a, b))
}

// Eq reports exact Decimal equality under JavaScript component semantics.
func (d Decimal) Eq(other Decimal) bool {
	a, b := d.components(), other.components()
	return a.sign == b.sign && a.layer == b.layer && a.mag == b.mag
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

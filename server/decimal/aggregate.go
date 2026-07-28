package decimal

import (
	"math"
	"sort"
)

// SumDeterministic returns an order-independent n-ary sum without mutating
// values. Same-exponent terms are combined before Decimal normalization so a
// valid cancellation cannot fail only because an intermediate prefix overflowed.
func SumDeterministic(values []Decimal) Decimal {
	ordered := append([]Decimal(nil), values...)
	for _, value := range ordered {
		if !value.IsStateValue() {
			return NaN
		}
	}
	sort.Slice(ordered, func(left, right int) bool {
		if ordered[left].exponent != ordered[right].exponent {
			return ordered[left].exponent < ordered[right].exponent
		}
		leftMagnitude, rightMagnitude := math.Abs(ordered[left].mantissa), math.Abs(ordered[right].mantissa)
		if leftMagnitude != rightMagnitude {
			return leftMagnitude < rightMagnitude
		}
		return ordered[left].mantissa < ordered[right].mantissa
	})

	total := Zero
	for start := 0; start < len(ordered); {
		exponent := ordered[start].exponent
		end := start
		mantissa := float64(0)
		for end < len(ordered) && ordered[end].exponent == exponent {
			mantissa += ordered[end].mantissa
			end++
		}
		if mantissa != 0 {
			term := arithmeticResult(mantissa, exponent)
			if !term.IsStateValue() {
				return NaN
			}
			total = total.Add(term)
			if !total.IsStateValue() {
				return NaN
			}
		}
		start = end
	}
	return total
}

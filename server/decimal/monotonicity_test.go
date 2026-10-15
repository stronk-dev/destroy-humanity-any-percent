package decimal

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
)

// Exact order, not relative-error closeness: a reversed threshold is a failure
// even if both amounts serialize to the same twelve-digit string.
func orderedPair(lower, upper Decimal, strict bool) error {
	if !lower.IsStateValue() || !upper.IsStateValue() {
		return fmt.Errorf("invalid output (%g,e%d), (%g,e%d)", lower.Mantissa(), lower.Exponent(), upper.Mantissa(), upper.Exponent())
	}
	comparison := lower.Cmp(upper)
	if comparison > 0 || strict && comparison == 0 {
		return fmt.Errorf("unordered output (%g,e%d), (%g,e%d), strict=%t", lower.Mantissa(), lower.Exponent(), upper.Mantissa(), upper.Exponent(), strict)
	}
	return nil
}

func TestOrderedArithmeticTransforms(t *testing.T) {
	var fixture struct {
		Version int
		Cases   []struct {
			Name, Lower, Upper, Addend, Scale string
			Terms                             []string
			Equal, Strict                     bool
		}
	}
	data, err := os.ReadFile("../../testdata/numeric-monotonicity-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.Version != 1 || len(fixture.Cases) != 18 {
		t.Fatal("unexpected ordered-input population")
	}
	for _, row := range fixture.Cases {
		t.Run(row.Name, func(t *testing.T) {
			lower, upper := FromString(row.Lower), FromString(row.Upper)
			addend, scale := FromString(row.Addend), FromString(row.Scale)
			terms := make([]Decimal, len(row.Terms))
			for i, source := range row.Terms {
				terms[i] = FromString(source)
			}
			for _, value := range append([]Decimal{lower, upper, addend, scale}, terms...) {
				if !value.IsStateValue() {
					t.Fatal("invalid fixture input")
				}
			}
			wantOrder := -1.0
			if row.Equal {
				wantOrder = 0
			}
			if lower.Cmp(upper) != wantOrder || scale.Cmp(Zero) <= 0 || row.Equal && row.Strict {
				t.Fatal("fixture input order/domain differs from literal declaration")
			}
			operations := []struct {
				name string
				run  func(Decimal) Decimal
			}{
				{"add", func(value Decimal) Decimal { return value.Add(addend) }},
				{"multiply", func(value Decimal) Decimal { return value.Mul(scale) }},
				{"divide", func(value Decimal) Decimal { return value.Div(scale) }},
				{"quantize", func(value Decimal) Decimal { return value.Quantize(CanonicalSignificantDigits) }},
				{"sum", func(value Decimal) Decimal { return SumDeterministic(append([]Decimal{value}, terms...)) }},
			}
			for _, operation := range operations {
				t.Run(operation.name, func(t *testing.T) {
					a, b := operation.run(lower), operation.run(upper)
					if err := orderedPair(a, b, row.Strict); err != nil {
						t.Fatal(err)
					}
					if err := orderedPair(a.Quantize(CanonicalSignificantDigits), b.Quantize(CanonicalSignificantDigits), row.Strict); err != nil {
						t.Fatalf("committed order: %v", err)
					}
					if row.Equal && !a.Eq(b) {
						t.Fatal("equal inputs produced unequal results")
					}
				})
			}
		})
	}
}

func TestOrderedPairControls(t *testing.T) {
	for _, row := range []struct {
		name         string
		lower, upper Decimal
		strict, pass bool
	}{
		{"increasing", One, FromFloat64(2), true, true},
		{"rounding-plateau", One, One, false, true},
		{"reversed", FromFloat64(2), One, false, false},
		{"reversed-below-canonical-resolution", FromString("1.000000000000001"), One, false, false},
		{"constant-shortcut", Zero, Zero, true, false},
		{"nan", NaN, One, false, false},
		{"infinity", One, Inf, false, false},
		{"unnormalized", Decimal{mantissa: 10}, FromFloat64(20), false, false},
	} {
		t.Run(row.name, func(t *testing.T) {
			if err := orderedPair(row.lower, row.upper, row.strict); (err == nil) != row.pass {
				t.Fatalf("pass=%t: %v", row.pass, err)
			}
		})
	}
}

func TestOrderedArithmeticUnderflowBoundary(t *testing.T) {
	// This first attempted ordered-input case leaves the gameplay domain. Keep
	// it explicit, outside the representable-output population. The pinned JS
	// library instead retains an invalid e=-9e15 diagnostic (tested there).
	got := FromString("1e-8999999999999999").Div(FromFloat64(2))
	if !got.Eq(Zero) || !got.IsStateValue() {
		t.Fatalf("Go underflow = (%g,e%d), want canonical zero", got.Mantissa(), got.Exponent())
	}
}

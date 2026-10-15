package decimal

import (
	"encoding/json"
	"math/big"
	"os"
	"slices"
	"testing"
)

// Negative evidence for a proposed proof lemma, not accepted payout semantics
// or proof that an actual production history reaches these bucket values.
func TestCanonicalGroupCarryResearch(t *testing.T) {
	data, err := os.ReadFile("../../testdata/numeric-canonical-group-carry-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Version               int    `json:"version"`
		Count                 int    `json:"repeat_count"`
		Repeated              string `json:"repeat_value"`
		Last                  string `json:"fixed_last"`
		Lower                 string `json:"variable_lower"`
		Upper                 string `json:"variable_upper"`
		LowerPayout           string `json:"lower_payout"`
		UpperPayout           string `json:"upper_payout"`
		LiteralDenominator    string `json:"literal_sum_denominator"`
		LiteralLowerNumerator string `json:"literal_lower_numerator"`
		LiteralUpperNumerator string `json:"literal_upper_numerator"`
	}
	if err := json.Unmarshal(data, &fixture); err != nil || fixture.Version != 1 || fixture.Count != 65535 {
		t.Fatal("incomplete canonical group-carry fixture", err)
	}
	parse := func(raw string) Decimal {
		t.Helper()
		value, err := ParseCanonical(raw)
		if err != nil || !value.Gt(Zero) {
			t.Fatal("invalid positive canonical input", raw, err)
		}
		return value
	}
	fixed := make([]Decimal, fixture.Count+1)
	for index := range fixture.Count {
		fixed[index] = parse(fixture.Repeated)
	}
	fixed[fixture.Count] = parse(fixture.Last)
	sum := func(variable Decimal) Decimal {
		t.Helper()
		values := append(slices.Clone(fixed), variable)
		before := slices.Clone(values)
		result := SumDeterministic(values)
		if !result.IsStateValue() || !slices.Equal(values, before) {
			t.Fatal("invalid output or mutated inputs")
		}
		slices.Reverse(values)
		if SumDeterministic(values) != result {
			t.Fatal("source ordering changes the result")
		}
		return result
	}
	lower, upper := sum(parse(fixture.Lower)), sum(parse(fixture.Upper))
	if !parse(fixture.Upper).Gt(parse(fixture.Lower)) || !lower.Gt(upper) ||
		lower.String() != fixture.LowerPayout || upper.String() != fixture.UpperPayout {
		t.Fatalf("lost canonical-input counterexample: raw(%g,e%d)→(%g,e%d), payout%s→%s", lower.mantissa, lower.exponent, upper.mantissa, upper.exponent, lower, upper)
	}
	// Independent exact rational sums of the literal decimal inputs are ordered;
	// neither the literal nor represented-input sum supports a smaller result.
	literal := func(variable string) *big.Rat {
		rational := func(raw string) *big.Rat {
			value, ok := new(big.Rat).SetString(raw)
			if !ok {
				t.Fatal("invalid exact reference", raw)
			}
			return value
		}
		result := rational(fixture.Repeated)
		result.Mul(result, new(big.Rat).SetInt64(int64(fixture.Count)))
		result.Add(result, rational(fixture.Last))
		return result.Add(result, rational(variable))
	}
	lo, hi := literal(fixture.Lower), literal(fixture.Upper)
	expected, ok := new(big.Rat).SetString(fixture.LiteralLowerNumerator + "/" + fixture.LiteralDenominator)
	if !ok || lo.Cmp(expected) != 0 {
		t.Fatal("wrong literal lower reference")
	}
	expected, ok = new(big.Rat).SetString(fixture.LiteralUpperNumerator + "/" + fixture.LiteralDenominator)
	if !ok || hi.Cmp(expected) != 0 || lo.Cmp(hi) >= 0 {
		t.Fatal("wrong ordered literal references")
	}
	if !sum(FromString("1e0")).Lt(sum(FromString("2e0"))) || sum(parse(fixture.Upper)) != upper {
		t.Fatal("ordinary/equal controls failed")
	}
	t.Logf("canonical65537-term raw(%g,e%d)→(%g,e%d), Q12%s→%s; literal sums%s→%s; reachability NOT proved", lower.mantissa, lower.exponent, upper.mantissa, upper.exponent, lower, upper, lo.RatString(), hi.RatString())
}

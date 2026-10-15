package decimal

import (
	"encoding/json"
	"math"
	"math/big"
	"os"
	"sort"
	"testing"
)

// Unapplied research, not a replacement for the specified grouped reducer.
// Align original terms once and use Neumaier compensation before normalization.
func alignedCompensatedResearch(values []Decimal) Decimal {
	ordered := append([]Decimal(nil), values...)
	exponent, any := int64(0), false
	for _, value := range ordered {
		if !value.IsStateValue() {
			return NaN
		}
		if value.mantissa != 0 && (!any || value.exponent > exponent) {
			exponent, any = value.exponent, true
		}
	}
	if !any {
		return Zero
	}
	sort.Slice(ordered, func(i, j int) bool {
		if ordered[i].exponent != ordered[j].exponent {
			return ordered[i].exponent < ordered[j].exponent
		}
		if math.Abs(ordered[i].mantissa) != math.Abs(ordered[j].mantissa) {
			return math.Abs(ordered[i].mantissa) < math.Abs(ordered[j].mantissa)
		}
		return ordered[i].mantissa < ordered[j].mantissa
	})
	sum, correction := 0.0, 0.0
	for _, value := range ordered {
		x := value.mantissa * math.Pow10(int(value.exponent-exponent))
		next := sum + x
		if math.Abs(sum) >= math.Abs(x) {
			correction += (sum - next) + x
		} else {
			correction += (x - next) + sum
		}
		sum = next
	}
	result := New(sum+correction, exponent)
	if !result.IsStateValue() {
		return NaN
	}
	return result
}

// Exact rational diagnostic of the represented inputs, deliberately bounded to
// this small-exponent experiment. This is not an enormous-domain runtime sum or
// a replacement for the normative float-mantissa Quantize operation.
func exactAggregationResearch(t *testing.T, values []Decimal) *big.Rat {
	t.Helper()
	total := new(big.Rat)
	for _, value := range values {
		if !value.IsStateValue() || value.exponent < -32 || value.exponent > 32 {
			t.Fatal("exact diagnostic does not admit this input")
		}
		term := new(big.Rat).SetFloat64(value.mantissa)
		power := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(math.Abs(float64(value.exponent)))), nil)
		scale := new(big.Rat).SetInt(power)
		if value.exponent < 0 {
			term.Quo(term, scale)
		} else {
			term.Mul(term, scale)
		}
		total.Add(total, term)
	}
	return total
}

func TestAggregationAlignedCandidateResearch(t *testing.T) {
	var fixture struct {
		Fixed    []struct{ Rate string }
		Variable struct {
			Lower string `json:"lower_rate"`
			Upper string `json:"upper_rate"`
		}
	}
	data, err := os.ReadFile("../../testdata/numeric-sum-order-research-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &fixture); err != nil || len(fixture.Fixed) != 4 {
		t.Fatal("invalid existing counterexample fixture", err)
	}
	fixed := []Decimal{}
	for _, row := range fixture.Fixed {
		fixed = append(fixed, FromString(row.Rate))
	}
	for _, variable := range []string{fixture.Variable.Lower, fixture.Variable.Upper} {
		values := append(append([]Decimal(nil), fixed...), FromString(variable))
		exact := exactAggregationResearch(t, values)
		// Both exact represented sums exceed the same canonical midpoint.
		midpoint, _ := new(big.Rat).SetString("10.12459571165")
		if exact.Cmp(midpoint) <= 0 {
			t.Fatal("lost independently specified midpoint control")
		}
		got := alignedCompensatedResearch(values)
		if !got.IsStateValue() || got.String() != "1.01245957117e1" {
			t.Fatalf("candidate does not repair the declared midpoint pair: (%g,e%d)", got.mantissa, got.exponent)
		}
		t.Logf("RP-664 variable%s exact%s current%s candidate%s raw(%g,e%d)", variable, exact.RatString(), SumDeterministic(values), got, got.mantissa, got.exponent)
	}
	// Retain the signed/domain controls, not only positive production examples.
	boundary := FromString("9e8999999999999999")
	for _, values := range [][]Decimal{{boundary, boundary, boundary.Neg(), boundary.Neg()}, {One, One.Neg()}, {}} {
		if !alignedCompensatedResearch(values).Eq(Zero) {
			t.Fatal("candidate fails existing exact cancellation/empty control")
		}
	}
	if !alignedCompensatedResearch([]Decimal{boundary, boundary}).IsNaN() || !alignedCompensatedResearch([]Decimal{NaN}).IsNaN() {
		t.Fatal("candidate fails existing domain refusal controls")
	}
}

func TestAggregationAlignedSeedResearch(t *testing.T) {
	const population = 100_000
	seed := uint64(663)
	next := func() uint64 {
		seed ^= seed << 13
		seed ^= seed >> 7
		seed ^= seed << 17
		return seed
	}
	currentRaw, currentCommitted, candidateRaw, candidateCommitted := 0, 0, 0, 0
	for index := 0; index < population; index++ {
		fixed := make([]Decimal, 4)
		for i := range fixed {
			fixed[i] = New(float64(100_000_000_000_000+next()%900_000_000_000_000)/1e14, int64(next()%14)-5)
		}
		lower := append(append([]Decimal(nil), fixed...), FromString("9.999999999999998e0"))
		upper := append(append([]Decimal(nil), fixed...), FromString("1e1"))
		for arm, sum := range []func([]Decimal) Decimal{SumDeterministic, alignedCompensatedResearch} {
			a, b := sum(lower), sum(upper)
			if !a.IsStateValue() || !b.IsStateValue() {
				t.Fatal("incomplete measurement: invalid admitted positive output", index, arm)
			}
			if a.Gt(b) {
				if arm == 0 {
					currentRaw++
				} else {
					candidateRaw++
				}
			}
			if a.Quantize(12).Gt(b.Quantize(12)) {
				if arm == 0 {
					currentCommitted++
				} else {
					candidateCommitted++
				}
				t.Logf("committed reversal index%d arm%d fixed%+v lower(%g,e%d) upper(%g,e%d)", index, arm, fixed, a.mantissa, a.exponent, b.mantissa, b.exponent)
			}
		}
	}
	t.Logf("completed%d seed663: current raw%d committed%d; candidate raw%d committed%d; diagnostic only, not a monotonicity proof", population, currentRaw, currentCommitted, candidateRaw, candidateCommitted)
}

func TestAggregationAlignedCandidateLosesCancellationTailResearch(t *testing.T) {
	// Exact same-exponent cancellation is an existing reason for grouping. The
	// tiny surviving term must not disappear just because the experiment aligns
	// every original term to the largest exponent before those terms cancel.
	tiny, large := FromString("1e-300"), FromString("1e300")
	values := []Decimal{large, large.Neg(), tiny}
	if got := SumDeterministic(values); !got.Eq(tiny) {
		t.Fatalf("lost current signed-tail control: %s", got)
	}
	if got := alignedCompensatedResearch(values); !got.Eq(Zero) {
		t.Fatalf("candidate's rejected cancellation behavior changed: %s; investigate", got)
	}
	t.Log("REJECTED candidate: current1e-300, aligned0 after exact large-term cancellation")
}

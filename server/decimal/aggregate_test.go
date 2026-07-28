package decimal

import "testing"

func TestSumDeterministicPermutationAndInputSafety(t *testing.T) {
	boundary := FromString("9e8999999999999999")
	inputs := [][]Decimal{
		{boundary, boundary, boundary.Neg(), boundary.Neg()},
		{boundary, boundary.Neg(), boundary, boundary.Neg()},
		{boundary.Neg(), boundary.Neg(), boundary, boundary},
	}
	for _, values := range inputs {
		before := append([]Decimal(nil), values...)
		if got := SumDeterministic(values); !got.Eq(Zero) {
			t.Fatalf("sum = %s, want zero", got.String())
		}
		for index := range values {
			if !values[index].Eq(before[index]) {
				t.Fatal("input mutated")
			}
		}
	}
	if got := SumDeterministic([]Decimal{boundary, boundary}); !got.IsNaN() {
		t.Fatalf("out-of-domain sum = %s, want NaN", got.String())
	}
}

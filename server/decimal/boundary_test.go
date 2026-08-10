package decimal

import "testing"

func TestExponentBoundaryArithmetic(t *testing.T) {
	tests := []struct {
		name string
		got  Decimal
		want string
	}{
		{"division avoids transient reciprocal overflow", FromString("1e8999999999999999").Div(FromString("2.5e8999999999999999")), "4e-1"},
		{"multiplication carry rescues lower bound", FromString("9e-8999999999999999").Mul(FromString("2e-1")), "1.8e-8999999999999999"},
		{"integer power carry rescues lower bound", FromString("4.44849805906e-4500000000000000").Pow(2), "1.97891349815e-8999999999999999"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := test.got.String(); got != test.want {
				t.Fatalf("got %s (mantissa=%g exponent=%d), want %s", got, test.got.Mantissa(), test.got.Exponent(), test.want)
			}
		})
	}
}

func TestInt64ExactAcceptsCanonicalIntegerAcrossArchitectures(t *testing.T) {
	// linux/amd64 reconstructs this normalized Decimal just below 926157 in
	// toFloat64 even though Floor returned the canonical representation of the
	// exact integer. Int64Exact must classify the Decimal representation, not a
	// second lossy float reconstruction.
	progress := FromString("9.26157482632e-1")
	got, ok := progress.Mul(FromFloat64(1_000_000)).Floor().Int64Exact()
	if !ok || got != 926_157 {
		t.Fatalf("got %d ok=%v, want 926157 true", got, ok)
	}

	for _, fractional := range []string{"9.261571e5", "-9.261571e5"} {
		if got, ok := FromString(fractional).Int64Exact(); ok {
			t.Fatalf("%s accepted as exact integer %d", fractional, got)
		}
	}
}

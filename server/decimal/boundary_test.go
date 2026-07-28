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

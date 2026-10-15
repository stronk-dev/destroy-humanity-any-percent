package decimal

import (
	"encoding/json"
	"os"
	"strconv"
	"testing"
)

// These raw diagnostics cannot use the canonical-string golden schema: their
// whole purpose is to fail the state/wire boundary, not to serialize as zero.
func TestArithmeticLowerDomainDiagnostics(t *testing.T) {
	var fixture struct {
		Version int
		Cases   []struct {
			Name, Op, A, B, Exponent string
			Mantissa                 float64
			State                    bool
		}
	}
	data, err := os.ReadFile("../../testdata/numeric-lower-domain-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.Version != 1 || len(fixture.Cases) != 12 {
		t.Fatal("unexpected lower-domain population")
	}
	for _, row := range fixture.Cases {
		t.Run(row.Name, func(t *testing.T) {
			a, b := FromString(row.A), FromString(row.B)
			if !a.IsStateValue() || !b.IsStateValue() {
				t.Fatal("invalid fixture input")
			}
			var got Decimal
			switch row.Op {
			case "mul":
				got = a.Mul(b)
			case "div":
				got = a.Div(b)
			case "pow":
				power, err := strconv.ParseFloat(row.B, 64)
				if err != nil {
					t.Fatal(err)
				}
				got = a.Pow(power)
			default:
				t.Fatalf("unknown operation %q", row.Op)
			}
			if got.Mantissa() != row.Mantissa || strconv.FormatInt(got.Exponent(), 10) != row.Exponent {
				t.Fatalf("raw result (%g,e%d), want (%g,e%s)", got.Mantissa(), got.Exponent(), row.Mantissa, row.Exponent)
			}
			if got.IsNaN() || !got.IsFinite() || got.IsStateValue() != row.State {
				t.Fatal("wrong finite/state classification")
			}
			if row.State {
				parsed, err := ParseCanonical(got.String())
				if err != nil || !parsed.Eq(got.Quantize(CanonicalSignificantDigits)) {
					t.Fatal("valid control failed wire round trip")
				}
			} else {
				if got.Eq(Zero) || got.Quantize(CanonicalSignificantDigits).IsStateValue() {
					t.Fatal("diagnostic became valid saved state")
				}
				if _, err := ParseCanonical(got.String()); err == nil {
					t.Fatal("diagnostic became a canonical wire value")
				}
				if New(row.Mantissa, got.Exponent()).IsStateValue() {
					t.Fatal("constructor expanded gameplay domain")
				}
			}
		})
	}
}

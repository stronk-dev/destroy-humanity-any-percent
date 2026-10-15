package decimal

import (
	"encoding/json"
	"math"
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
			Mantissa                 *float64
			State                    bool
			ClassificationOnly       bool `json:"classification_only"`
			Sign                     int
		}
	}
	data, err := os.ReadFile("../../testdata/numeric-lower-domain-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.Version != 1 || len(fixture.Cases) != 24 {
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
			case "exp":
				got = a.Exp()
			case "pow":
				power, err := strconv.ParseFloat(row.B, 64)
				if err != nil {
					t.Fatal(err)
				}
				got = a.Pow(power)
			default:
				t.Fatalf("unknown operation %q", row.Op)
			}
			if row.ClassificationOnly {
				// Transcendental intermediates need not be bit-identical. These
				// finite lower-domain diagnostics must still agree on rejection,
				// not silently become zero. No valid-value tolerance is widened.
				if row.State || row.Mantissa != nil || row.Exponent != "" || got.Exponent() >= -maxExponent ||
					(row.Sign != -1 && row.Sign != 1) || math.Signbit(got.Mantissa()) != (row.Sign < 0) ||
					math.Abs(got.Mantissa()) < 1 || math.Abs(got.Mantissa()) >= 10 {
					t.Fatalf("invalid lower-domain classification case/result (%g,e%d)", got.Mantissa(), got.Exponent())
				}
			} else {
				if row.Mantissa == nil {
					t.Fatal("missing literal mantissa")
				}
				if got.Mantissa() != *row.Mantissa || strconv.FormatInt(got.Exponent(), 10) != row.Exponent {
					t.Fatalf("raw result (%g,e%d), want (%g,e%s)", got.Mantissa(), got.Exponent(), *row.Mantissa, row.Exponent)
				}
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
				if New(got.Mantissa(), got.Exponent()).IsStateValue() {
					t.Fatal("constructor expanded gameplay domain")
				}
			}
		})
	}
}

func TestPowerDiagnosticExponentArithmeticDoesNotWrapIntoState(t *testing.T) {
	// A valid input can produce a finite diagnostic at int64's lower edge.
	// Applying another operation must not wrap its exponent back into state.
	base := FromString("1e-8999999999999999")
	value := base.Pow(float64(math.MinInt64) / -float64(maxExponent))
	if value.Exponent() != math.MinInt64 || !value.IsFinite() || math.Abs(value.Mantissa()) < 1 || math.Abs(value.Mantissa()) >= 10 || value.IsStateValue() {
		t.Fatalf("missing actual power diagnostic: (%g,e%d)", value.Mantissa(), value.Exponent())
	}
	for _, result := range []Decimal{value.Mul(value), value.Div(FromString("1e8999999999999999")), One.Div(value)} {
		if !result.IsNaN() || result.IsStateValue() {
			t.Fatal("unrepresentable exponent must refuse, not wrap into state")
		}
	}
	if !value.Div(value).Eq(One) || !value.Mul(Zero).Eq(Zero) || !Zero.Mul(value).Eq(Zero) || !Zero.Div(value).Eq(Zero) {
		t.Fatal("representable quotient/zero controls changed")
	}
	boundary := FromString("1e8999999999999999")
	for _, result := range []Decimal{boundary.Add(value), value.Add(boundary), boundary.Sub(value)} {
		if !result.Eq(boundary) || !result.IsStateValue() {
			t.Fatal("diagnostic exponent gap overflow corrupted a representable sum")
		}
	}
}

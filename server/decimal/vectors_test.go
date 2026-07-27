package decimal

import (
	"encoding/json"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

type goldenVector struct {
	Assert      string `json:"assert"`
	A           string `json:"a"`
	B           string `json:"b"`
	Op          string `json:"op"`
	Ratio       string `json:"ratio"`
	Owned       string `json:"owned"`
	Expect      string `json:"expect"`
	ExpectClass string `json:"expectClass"`
}

type goldenFile struct {
	Version int            `json:"version"`
	Seed    uint32         `json:"seed"`
	Vectors []goldenVector `json:"vectors"`
}

func loadGoldenVectors(t testing.TB) goldenFile {
	t.Helper()
	path := filepath.Join("..", "..", "testdata", "decimal-vectors.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var fixture goldenFile
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	return fixture
}

func vectorDecimal(t testing.TB, vector goldenVector) Decimal {
	t.Helper()
	a := FromString(vector.A)
	b := FromString(vector.B)
	switch vector.Op {
	case "add", "commit-add":
		return a.Add(b)
	case "sub":
		return a.Sub(b)
	case "mul", "commit-mul":
		return a.Mul(b)
	case "div":
		return a.Div(b)
	case "pow":
		power, err := strconv.ParseFloat(vector.B, 64)
		if err != nil {
			t.Fatal(err)
		}
		return a.Pow(power)
	case "log10":
		return a.Log10()
	case "ln":
		return a.Ln()
	case "exp":
		return a.Exp()
	case "floor":
		return a.Floor()
	case "sum":
		owned := parseVectorInt(t, vector.Owned)
		count := parseVectorInt(t, vector.A)
		return SumGeometricSeries(count, b, FromString(vector.Ratio), owned)
	default:
		t.Fatalf("operation %q does not return Decimal", vector.Op)
		return NaN
	}
}

func parseVectorInt(t testing.TB, value string) int64 {
	t.Helper()
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		t.Fatal(err)
	}
	return parsed
}

func decimalClass(value Decimal) string {
	if value.IsNaN() {
		return "nan"
	}
	if math.IsInf(value.mantissa, 1) {
		return "positive-infinity"
	}
	if math.IsInf(value.mantissa, -1) {
		return "negative-infinity"
	}
	return "finite"
}

func assertApprox(t testing.TB, got Decimal, vector goldenVector) {
	t.Helper()
	if class := decimalClass(got); class != vector.ExpectClass {
		t.Fatalf("%s(%q, %q) classification: got %q, want %q", vector.Op, vector.A, vector.B, class, vector.ExpectClass)
	}
	if vector.ExpectClass != "finite" {
		return
	}
	want := FromString(vector.Expect)
	if want.Eq(Zero) {
		if !got.Eq(Zero) {
			t.Fatalf("%s(%q, %q): got %q, want exact zero", vector.Op, vector.A, vector.B, got.String())
		}
		return
	}
	relative := symmetricRelativeError(got, want)
	if math.IsNaN(relative) || relative > 1e-12 {
		t.Fatalf("%s(%q, %q): got %q, reference %q, relative error %.3g", vector.Op, vector.A, vector.B, got.String(), vector.Expect, relative)
	}
}

// symmetricRelativeError works directly on normalized components. Computing
// got.Sub(want) through Decimal would deliberately discard close differences
// once their logarithmic representations cross break_eternity's precision
// threshold, making the test metric less accurate than the values under test.
func symmetricRelativeError(a, b Decimal) float64 {
	if math.Signbit(a.mantissa) != math.Signbit(b.mantissa) {
		return 1
	}
	maxExponent := max(a.exponent, b.exponent)
	aScaled := math.Abs(a.mantissa) * math.Pow(10, float64(a.exponent-maxExponent))
	bScaled := math.Abs(b.mantissa) * math.Pow(10, float64(b.exponent-maxExponent))
	return math.Abs(aScaled-bScaled) / math.Max(aScaled, bScaled)
}

func assertExact(t testing.TB, vector goldenVector) {
	t.Helper()
	var got string
	switch vector.Op {
	case "canonical":
		got = FromString(vector.A).String()
	case "parse-valid":
		_, err := ParseCanonical(vector.A)
		got = strconv.FormatBool(err == nil)
	case "state-valid":
		got = strconv.FormatBool(FromString(vector.A).IsStateValue())
	case "cmp":
		got = formatNumber(FromString(vector.A).Cmp(FromString(vector.B)))
	case "commit-add", "commit-mul":
		got = vectorDecimal(t, vector).String()
	default:
		t.Fatalf("unknown exact operation %q", vector.Op)
	}
	if got != vector.Expect {
		t.Fatalf("%s(%q, %q): got %q, want %q", vector.Op, vector.A, vector.B, got, vector.Expect)
	}
}

func assertDecision(t testing.TB, vector goldenVector) {
	t.Helper()
	owned := parseVectorInt(t, vector.Owned)
	cash := FromString(vector.A)
	base := FromString(vector.B)
	ratio := FromString(vector.Ratio)
	got, err := AffordGeometricSeries(cash, base, ratio, owned)
	if err != nil {
		t.Fatal(err)
	}
	want := parseVectorInt(t, vector.Expect)
	if got != want {
		t.Fatalf("afford(%q, %q): got %d, want %d", vector.A, vector.B, got, want)
	}
	if !affordabilityPostconditions(got, cash, base, ratio, owned) {
		t.Fatalf("afford(%q, %q): result %d violates postconditions", vector.A, vector.B, got)
	}
}

func TestGoldenVectors(t *testing.T) {
	fixture := loadGoldenVectors(t)
	if fixture.Version != 2 {
		t.Fatalf("got vector schema %d, want 2", fixture.Version)
	}
	if len(fixture.Vectors) < 5_000 {
		t.Fatalf("got %d vectors, want at least 5000", len(fixture.Vectors))
	}
	for index, vector := range fixture.Vectors {
		t.Run(strconv.Itoa(index)+"/"+vector.Op, func(t *testing.T) {
			switch vector.Assert {
			case "exact":
				assertExact(t, vector)
			case "approx":
				assertApprox(t, vectorDecimal(t, vector), vector)
			case "decision":
				assertDecision(t, vector)
			default:
				t.Fatalf("unknown assertion category %q", vector.Assert)
			}
		})
	}
}

func TestCanonicalRoundTripProperties(t *testing.T) {
	rng := rand.New(rand.NewSource(0xc10dc1))
	for index := 0; index < 10_000; index++ {
		mantissa := 1 + rng.Float64()*8.999999999
		if rng.Intn(2) == 0 {
			mantissa = -mantissa
		}
		exponent := int64(rng.Intn(2_000_001) - 1_000_000)
		value := New(mantissa, exponent).Quantize(CanonicalSignificantDigits)
		parsed, err := ParseCanonical(value.String())
		if err != nil {
			t.Fatalf("parse %q: %v", value.String(), err)
		}
		if parsed.String() != value.String() {
			t.Fatalf("round trip changed %q to %q", value.String(), parsed.String())
		}
		if value.mantissa != 0 && (math.Abs(value.mantissa) < 1 || math.Abs(value.mantissa) >= 10) {
			t.Fatalf("not normalized: %#v", value)
		}
	}
}

func TestNonFiniteValuesAreDiagnosticOnly(t *testing.T) {
	for _, value := range []Decimal{NaN, Inf, NegInf} {
		if value.IsStateValue() {
			t.Fatalf("%q must not be valid state", value.String())
		}
		if _, err := ParseCanonical(value.String()); err == nil {
			t.Fatalf("%q must not parse as canonical state", value.String())
		}
	}
}

func FuzzCanonicalRoundTrip(f *testing.F) {
	for _, seed := range []string{"0", "1e0", "-4.25e-7", "9.87654321012e123456", "NaN", "Infinity"} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, source string) {
		value := FromString(source)
		_ = value.Add(One).Mul(value).Div(One).String()
		if !value.IsStateValue() {
			return
		}
		canonical := value.String()
		parsed, err := ParseCanonical(canonical)
		if err != nil || parsed.String() != canonical {
			t.Fatalf("canonical round trip failed for %q via %q", source, canonical)
		}
	})
}

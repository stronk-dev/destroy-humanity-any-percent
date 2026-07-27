package decimal

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

type goldenVector struct {
	A      string `json:"a"`
	B      string `json:"b"`
	Op     string `json:"op"`
	Ratio  string `json:"ratio"`
	Owned  string `json:"owned"`
	Expect string `json:"expect"`
}

type goldenFile struct {
	Seed    uint32         `json:"seed"`
	Vectors []goldenVector `json:"vectors"`
}

func loadGoldenVectors(t *testing.T) goldenFile {
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

func evaluateGolden(t *testing.T, vector goldenVector) string {
	t.Helper()
	a := FromString(vector.A)
	b := FromString(vector.B)
	switch vector.Op {
	case "add":
		return a.Add(b).String()
	case "sub":
		return a.Sub(b).String()
	case "mul":
		return a.Mul(b).String()
	case "div":
		return a.Div(b).String()
	case "pow":
		power, err := strconv.ParseFloat(vector.B, 64)
		if err != nil {
			t.Fatal(err)
		}
		return a.Pow(power).String()
	case "log10":
		return a.Log10().String()
	case "ln":
		return a.Ln().String()
	case "exp":
		return a.Exp().String()
	case "floor":
		return a.Floor().String()
	case "cmp":
		return formatNumber(a.Cmp(b))
	case "sum", "afford":
		owned, err := strconv.ParseInt(vector.Owned, 10, 64)
		if err != nil {
			t.Fatal(err)
		}
		ratio := FromString(vector.Ratio)
		if vector.Op == "sum" {
			count, err := strconv.ParseInt(vector.A, 10, 64)
			if err != nil {
				t.Fatal(err)
			}
			return SumGeometricSeries(count, b, ratio, owned).String()
		}
		return AffordGeometricSeries(a, b, ratio, owned).String()
	default:
		t.Fatalf("unknown vector operation %q", vector.Op)
		return ""
	}
}

func TestGoldenVectors(t *testing.T) {
	fixture := loadGoldenVectors(t)
	if len(fixture.Vectors) < 5_000 {
		t.Fatalf("got %d vectors, want at least 5000", len(fixture.Vectors))
	}
	for index, vector := range fixture.Vectors {
		got := evaluateGolden(t, vector)
		if got != vector.Expect {
			t.Fatalf("vector %d %s(%q, %q): got %q, want %q", index, vector.Op, vector.A, vector.B, got, vector.Expect)
		}
	}
}

func TestSpecialValuesRoundTripWithoutPanics(t *testing.T) {
	for _, value := range []string{"0", "-1", "NaN", "Infinity", "-Infinity"} {
		parsed := FromString(value)
		if got := parsed.String(); got != value && !(value == "-1" && got == "-1") {
			t.Fatalf("round trip %q: got %q", value, got)
		}
		for _, operation := range []func() Decimal{
			func() Decimal { return parsed.Add(One) },
			func() Decimal { return parsed.Sub(One) },
			func() Decimal { return parsed.Mul(Zero) },
			func() Decimal { return parsed.Div(Zero) },
			func() Decimal { return parsed.Pow(2) },
			func() Decimal { return parsed.Log10() },
			func() Decimal { return parsed.Ln() },
			func() Decimal { return parsed.Exp() },
		} {
			_ = operation().String()
		}
	}
	if !math.IsNaN(NaN.Cmp(One)) {
		t.Fatal("NaN comparison must return NaN")
	}
}


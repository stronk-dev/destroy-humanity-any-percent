package economy

import (
	"math/rand"
	"testing"

	"cloud-clicker/server/decimal"
)

var affordableBenchmarkSink int64

func TestGeometricMaxAffordableBoundaries(t *testing.T) {
	tests := []struct {
		name  string
		price PriceDefinition
		cash  decimal.Decimal
		owned int64
		want  int64
	}{
		{
			name:  "zero cash",
			price: geometricPrice(decimal.FromFloat64(10), decimal.FromFloat64(1.13)),
			cash:  decimal.Zero,
			want:  0,
		},
		{
			name:  "ratio one respects remaining exact integer capacity",
			price: geometricPrice(decimal.FromFloat64(2), decimal.One),
			cash:  decimal.FromFloat64(100),
			owned: decimal.MaxExactInteger - 3,
			want:  3,
		},
		{
			name:  "owned adjacent to exact integer ceiling",
			price: geometricPrice(decimal.One, decimal.FromFloat64(1.13)),
			cash:  decimal.New(1, 1_000_000_000_000_000),
			owned: decimal.MaxExactInteger - 1,
			want:  1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := MaxAffordable(test.price, test.cash, test.owned)
			if err != nil {
				t.Fatal(err)
			}
			if got != test.want {
				t.Fatalf("MaxAffordable() = %d, want %d", got, test.want)
			}
			assertAffordablePostconditions(t, test.price, test.cash, test.owned, got)
		})
	}
}

func TestGeometricMaxAffordableHugeExponent(t *testing.T) {
	price := geometricPrice(
		decimal.New(1, 1_000_000),
		decimal.FromFloat64(1.13),
	)
	const owned = int64(2_000)
	const purchased = int64(250)
	cash, err := BulkCost(price, owned, purchased)
	if err != nil {
		t.Fatal(err)
	}

	got, err := MaxAffordable(price, cash, owned)
	if err != nil {
		t.Fatal(err)
	}
	if got != purchased {
		t.Fatalf("MaxAffordable() = %d, want %d", got, purchased)
	}
	assertAffordablePostconditions(t, price, cash, owned, got)
}

func TestMaxAffordableGeneratedPostconditions(t *testing.T) {
	random := rand.New(rand.NewSource(20260728))
	for index := 0; index < 500; index++ {
		owned := random.Int63n(1_000_000)
		target := random.Int63n(10_000)
		base := decimal.FromFloat64(float64(random.Intn(10_000)+1) / 10)
		ratio := decimal.FromFloat64(1 + float64(random.Intn(200))/1_000)
		price := geometricPrice(base, ratio)
		cash, err := BulkCost(price, owned, target)
		if err != nil {
			t.Fatalf("case %d setup: %v", index, err)
		}

		got, err := MaxAffordable(price, cash, owned)
		if err != nil {
			t.Fatalf("case %d: %v", index, err)
		}
		assertAffordablePostconditions(t, price, cash, owned, got)
	}
}

func TestConstantAndLinearMaxAffordableRegression(t *testing.T) {
	tests := []struct {
		name  string
		price PriceDefinition
		cash  decimal.Decimal
		owned int64
		want  int64
	}{
		{
			name: "constant",
			price: PriceDefinition{
				Base:  decimal.FromFloat64(5),
				Curve: CostCurve{Kind: CurveConstant},
			},
			cash: decimal.FromFloat64(52),
			want: 10,
		},
		{
			name: "linear",
			price: PriceDefinition{
				Base:  decimal.FromFloat64(10),
				Curve: CostCurve{Kind: CurveLinear, Step: decimal.FromFloat64(2)},
			},
			cash:  decimal.FromFloat64(54),
			owned: 3,
			want:  3,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := MaxAffordable(test.price, test.cash, test.owned)
			if err != nil {
				t.Fatal(err)
			}
			if got != test.want {
				t.Fatalf("MaxAffordable() = %d, want %d", got, test.want)
			}
			assertAffordablePostconditions(t, test.price, test.cash, test.owned, got)
		})
	}
}

func TestGeometricFastPathBenchmarkRatio(t *testing.T) {
	price, cash, owned := affordabilityBenchmarkInputs()
	public := testing.Benchmark(func(b *testing.B) {
		for index := 0; index < b.N; index++ {
			affordableBenchmarkSink, _ = MaxAffordable(price, cash, owned)
		}
	})
	helper := testing.Benchmark(func(b *testing.B) {
		for index := 0; index < b.N; index++ {
			affordableBenchmarkSink, _ = decimal.AffordGeometricSeries(
				cash,
				price.Base,
				price.Curve.Ratio,
				owned,
			)
		}
	})
	if helper.NsPerOp() == 0 {
		t.Fatal("helper benchmark resolution is zero nanoseconds per operation")
	}
	ratio := float64(public.NsPerOp()) / float64(helper.NsPerOp())
	t.Logf("public=%d ns/op helper=%d ns/op ratio=%.2fx", public.NsPerOp(), helper.NsPerOp(), ratio)
	if ratio >= 10 {
		t.Fatalf("public geometric path is %.2fx helper; want <10x", ratio)
	}
}

func BenchmarkGeometricMaxAffordable(b *testing.B) {
	price, cash, owned := affordabilityBenchmarkInputs()
	b.Run("economy", func(b *testing.B) {
		for index := 0; index < b.N; index++ {
			affordableBenchmarkSink, _ = MaxAffordable(price, cash, owned)
		}
	})
	b.Run("decimal-helper", func(b *testing.B) {
		for index := 0; index < b.N; index++ {
			affordableBenchmarkSink, _ = decimal.AffordGeometricSeries(
				cash,
				price.Base,
				price.Curve.Ratio,
				owned,
			)
		}
	})
}

func affordabilityBenchmarkInputs() (PriceDefinition, decimal.Decimal, int64) {
	return geometricPrice(
		decimal.FromFloat64(100),
		decimal.FromFloat64(1.13),
	), decimal.New(1, 1_000), 10_000
}

func geometricPrice(base, ratio decimal.Decimal) PriceDefinition {
	return PriceDefinition{
		Base: base,
		Curve: CostCurve{
			Kind:  CurveGeometric,
			Ratio: ratio,
		},
	}
}

func assertAffordablePostconditions(t *testing.T, price PriceDefinition, cash decimal.Decimal, owned, count int64) {
	t.Helper()
	maximum := decimal.MaxExactInteger - owned
	if count < 0 || count > maximum {
		t.Fatalf("count %d outside 0..%d", count, maximum)
	}
	cost, err := BulkCost(price, owned, count)
	if err != nil || !cost.Lte(cash) {
		t.Fatalf("count %d is not affordable: cost=%s err=%v cash=%s", count, cost.String(), err, cash.String())
	}
	if count == maximum {
		return
	}
	next, err := BulkCost(price, owned, count+1)
	if err == nil && next.Lte(cash) {
		t.Fatalf("count %d is not maximal: next cost=%s cash=%s", count, next.String(), cash.String())
	}
}

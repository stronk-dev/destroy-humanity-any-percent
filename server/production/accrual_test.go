package production

import (
	"encoding/json"
	"os"
	"testing"

	"cloud-clicker/server/decimal"
)

type accrualFixture struct {
	Version int             `json:"version"`
	Vectors []accrualVector `json:"vectors"`
}

type accrualVector struct {
	Name        string   `json:"name"`
	Rates       []string `json:"rates"`
	ElapsedMS   int64    `json:"elapsed_ms"`
	Efficiency  string   `json:"efficiency"`
	Expect      string   `json:"expect"`
	ExpectError bool     `json:"expect_error"`
}

func TestAccrualVectors(t *testing.T) {
	data, err := os.ReadFile("../../testdata/production-accrual.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture accrualFixture
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.Version != 1 {
		t.Fatalf("version = %d", fixture.Version)
	}
	for _, vector := range fixture.Vectors {
		t.Run(vector.Name, func(t *testing.T) {
			rates := make([]decimal.Decimal, len(vector.Rates))
			for index, source := range vector.Rates {
				rates[index] = decimal.FromString(source)
			}
			before := append([]decimal.Decimal(nil), rates...)
			got, err := AccrueConstant(rates, vector.ElapsedMS, decimal.FromString(vector.Efficiency))
			if vector.ExpectError {
				if err == nil {
					t.Fatalf("got %s, want error", got.String())
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got.String() != vector.Expect {
				t.Fatalf("got %s, want %s", got.String(), vector.Expect)
			}
			for index := range rates {
				if !rates[index].Eq(before[index]) {
					t.Fatal("input rates mutated")
				}
			}
		})
	}
}

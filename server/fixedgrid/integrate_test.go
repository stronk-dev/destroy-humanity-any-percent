package fixedgrid

import (
	"encoding/json"
	"math/big"
	"os"
	"testing"
)

type fixture struct {
	Cases []struct {
		Name          string `json:"name"`
		Units         int64  `json:"units"`
		Rate          int64  `json:"rate"`
		Remainder     int64  `json:"remainder"`
		Divisor       int64  `json:"divisor"`
		Valid         bool   `json:"valid"`
		Whole         string `json:"whole"`
		NextRemainder int64  `json:"next_remainder"`
	} `json:"cases"`
}

func TestSharedVectors(t *testing.T) {
	raw, err := os.ReadFile("../../testdata/fixed-grid-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	var vectors fixture
	if err := json.Unmarshal(raw, &vectors); err != nil || len(vectors.Cases) == 0 {
		t.Fatalf("fixture cases=%d err=%v", len(vectors.Cases), err)
	}
	for _, testCase := range vectors.Cases {
		t.Run(testCase.Name, func(t *testing.T) {
			actual, err := Integrate(testCase.Units, testCase.Rate, testCase.Remainder, testCase.Divisor)
			if (err == nil) != testCase.Valid {
				t.Fatalf("valid=%v err=%v", testCase.Valid, err)
			}
			if testCase.Valid && (actual.Whole.String() != testCase.Whole || actual.Remainder != testCase.NextRemainder) {
				t.Fatalf("actual=%s/%d want=%s/%d", actual.Whole, actual.Remainder, testCase.Whole, testCase.NextRemainder)
			}
		})
	}
}

func TestPartitionInvariant(t *testing.T) {
	const rate, divisor = int64(333_333), int64(1_000_003)
	combined, err := Integrate(9_000_000_000_000, rate, 777_777, divisor)
	if err != nil {
		t.Fatal(err)
	}
	first, err := Integrate(4_000_000_000_000, rate, 777_777, divisor)
	if err != nil || !first.Whole.IsInt64() {
		t.Fatalf("first=%+v err=%v", first, err)
	}
	second, err := Integrate(5_000_000_000_000, rate, first.Remainder, divisor)
	if err != nil {
		t.Fatal(err)
	}
	splitWhole := new(big.Int).Add(first.Whole, second.Whole)
	if splitWhole.Cmp(combined.Whole) != 0 || second.Remainder != combined.Remainder {
		t.Fatalf("combined=%s/%d split=%s/%d", combined.Whole, combined.Remainder, splitWhole, second.Remainder)
	}
}

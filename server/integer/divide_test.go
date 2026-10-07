package integer

import (
	"errors"
	"math"
	"testing"
)

func TestDivideFloor(t *testing.T) {
	for _, row := range []struct{ value, divisor, expected int64 }{
		{0, 1, 0}, {7, 2, 3}, {1, 2, 0}, {math.MaxInt64, 1, math.MaxInt64}, {math.MaxInt64, 2, 4611686018427387903},
	} {
		actual, err := DivideFloor(row.value, row.divisor)
		if err != nil || actual != row.expected {
			t.Fatalf("%d/%d=%d err=%v want=%d", row.value, row.divisor, actual, err, row.expected)
		}
	}
	for _, row := range [][2]int64{{-1, 1}, {1, 0}, {1, -1}, {math.MinInt64, 1}} {
		if _, err := DivideFloor(row[0], row[1]); !errors.Is(err, ErrInvalidDivision) {
			t.Fatalf("invalid pair %v: %v", row, err)
		}
	}
}

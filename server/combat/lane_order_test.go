package combat

import (
	"encoding/json"
	"math"
	"os"
	"reflect"
	"slices"
	"testing"
)

func TestLaneOrderVectors(t *testing.T) {
	data, err := os.ReadFile("../../testdata/combat/lane-order-vectors.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Version     int `json:"version"`
		Comparisons []struct {
			Name     string   `json:"name"`
			Left     [3]int32 `json:"left"`
			Right    [3]int32 `json:"right"`
			Expected int      `json:"expected"`
		} `json:"comparisons"`
	}
	if err := json.Unmarshal(data, &fixture); err != nil || fixture.Version != 1 || len(fixture.Comparisons) != 10 {
		t.Fatalf("incomplete shared lane-order vectors: %v", err)
	}
	for _, row := range fixture.Comparisons {
		t.Run(row.Name, func(t *testing.T) {
			left := LaneOrderKey{row.Left[0], row.Left[1], row.Left[2]}
			right := LaneOrderKey{row.Right[0], row.Right[1], row.Right[2]}
			if got := CompareLaneOrder(left, right); got != row.Expected {
				t.Fatalf("compare(%v,%v)=%d, want %d", left, right, got, row.Expected)
			}
			if CompareLaneOrder(right, left) != -row.Expected {
				t.Fatal("reverse order differs")
			}
		})
	}
}

func TestLaneOrderRankAndSorting(t *testing.T) {
	// Nested enumeration independently assigns canonical ranks. Do not derive
	// the expected order by sorting with the comparator under test.
	keys := make([]LaneOrderKey, 0, 18)
	for _, tick := range []int32{math.MinInt32, 0, math.MaxInt32} {
		for _, slot := range []int32{0, 1} {
			for _, seq := range []int32{math.MinInt32, 0, math.MaxInt32} {
				keys = append(keys, LaneOrderKey{tick, slot, seq})
			}
		}
	}
	for i, left := range keys {
		for j, right := range keys {
			want := 0
			if i < j {
				want = -1
			} else if i > j {
				want = 1
			}
			if got := CompareLaneOrder(left, right); got != want {
				t.Fatalf("rank %d vs %d: got %d want %d", i, j, got, want)
			}
		}
	}
	for shift := range keys {
		input := append(slices.Clone(keys[shift:]), keys[:shift]...)
		slices.Reverse(input)
		before := slices.Clone(input)
		ordered := slices.Clone(input)
		slices.SortFunc(ordered, CompareLaneOrder)
		if !reflect.DeepEqual(ordered, keys) || !reflect.DeepEqual(input, before) {
			t.Fatalf("sorting rotation %d changed source or produced wrong order", shift)
		}
	}
}

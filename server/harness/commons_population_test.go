package harness

import (
	"os"
	"sort"
	"testing"

	"cloud-clicker/server/commons"
)

func TestCommonsPopulationInvariance(t *testing.T) {
	data, err := os.ReadFile("../../balance/commons/phase0.json")
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := commons.LoadCatalog(data)
	if err != nil {
		t.Fatal(err)
	}
	const seedCount = int64(128)
	var leftTotal, rightTotal int64
	leftValues, rightValues := make([]int64, 0, seedCount), make([]int64, 0, seedCount)
	for seed := uint64(0); seed < uint64(seedCount); seed++ {
		left, err := SimulateCommonsPopulation(catalog, 200, seed)
		if err != nil {
			t.Fatal(err)
		}
		right, err := SimulateCommonsPopulation(catalog, 20_000, seed)
		if err != nil {
			t.Fatal(err)
		}
		leftTotal += left.MeanModifierPPM
		rightTotal += right.MeanModifierPPM
		leftValues = append(leftValues, left.MeanModifierPPM)
		rightValues = append(rightValues, right.MeanModifierPPM)
	}
	difference := leftTotal/seedCount - rightTotal/seedCount
	if difference < 0 {
		difference = -difference
	}
	if difference > catalog.PopulationTolerancePPM {
		t.Fatalf("mean modifier differs by %d ppm at 200 vs 20,000 (limit %d)", difference, catalog.PopulationTolerancePPM)
	}
	sort.Slice(leftValues, func(i, j int) bool { return leftValues[i] < leftValues[j] })
	sort.Slice(rightValues, func(i, j int) bool { return rightValues[i] < rightValues[j] })
	lower, upper := int(seedCount*25/1000), int(seedCount*975/1000)
	if leftValues[upper] < rightValues[lower] || rightValues[upper] < leftValues[lower] {
		t.Fatalf("95%% intervals do not overlap: 200=[%d,%d] 20000=[%d,%d]", leftValues[lower], leftValues[upper], rightValues[lower], rightValues[upper])
	}
}

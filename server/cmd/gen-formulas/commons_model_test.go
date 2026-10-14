package main

import (
	"bytes"
	"encoding/json"
	"testing"

	"cloud-clicker/server/commons"
	"cloud-clicker/server/multiplier"
)

func TestCommonsSourceWeightDocumentCopyPreservesBytesAndIsolation(t *testing.T) {
	for _, source := range [][]commons.SourceWeight{
		{}, {{SourceID: "source.clean", Slot: multiplier.SlotAxisStack, WeightPPM: 1_000_000, Forsworn: false},
			{SourceID: "source.dark", Slot: multiplier.SlotUpgrades, WeightPPM: 1, Forsworn: true}},
	} {
		expected, err := json.Marshal(source)
		if err != nil {
			t.Fatal(err)
		}
		model := commonsSourceWeights(source)
		actual, err := json.Marshal(model)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(expected, actual) {
			t.Fatalf("document wire changed: %s != %s", expected, actual)
		}
		if len(model) > 0 {
			model[0].WeightPPM = 0
			if source[0].WeightPPM != 1_000_000 {
				t.Fatal("document retained producer ownership")
			}
		}
	}
	if result := commonsSourceWeights(nil); result == nil || len(result) != 0 {
		t.Fatal("nil input did not preserve canonical empty array")
	}
}

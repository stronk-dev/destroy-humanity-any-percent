package multiplier

import (
	"reflect"
	"testing"
)

func TestSlotOrderIsClosedExactAndUnique(t *testing.T) {
	want := []Slot{
		SlotUpgrades, SlotMilestones, SlotFaction, SlotDoctrine,
		SlotCommons, SlotTrust, SlotEventBuffs, SlotPrestige,
	}
	if !reflect.DeepEqual(Order[:], want) {
		t.Fatalf("order = %v, want %v", Order, want)
	}
	seen := make(map[Slot]bool, len(Order))
	for _, slot := range Order {
		if seen[slot] || !ValidSlot(slot) {
			t.Fatalf("invalid or duplicate slot %q", slot)
		}
		seen[slot] = true
	}
	if ValidSlot("dark_magic") {
		t.Fatal("unknown slot accepted")
	}
}

func TestOrderedSourceIDsUsesRawBytesAndDoesNotMutateInput(t *testing.T) {
	input := []string{"a.a", "a_a", "a0"}
	wantInput := append([]string(nil), input...)
	got := OrderedSourceIDs(input)
	want := []string{"a.a", "a0", "a_a"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ordered = %v, want raw-byte order %v", got, want)
	}
	if !reflect.DeepEqual(input, wantInput) {
		t.Fatalf("input mutated: got %v, want %v", input, wantInput)
	}
	if WithinSlotOrder != "source_id_raw_byte_ascending" {
		t.Fatalf("published order token = %q", WithinSlotOrder)
	}
}

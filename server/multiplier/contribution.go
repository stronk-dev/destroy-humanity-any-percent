// Package multiplier defines the neutral boundary between production and
// state-owning multiplier providers.
package multiplier

import (
	"sort"

	"cloud-clicker/server/decimal"
)

type Slot string

const (
	SlotUpgrades   Slot = "upgrades"
	SlotMilestones Slot = "milestones"
	SlotFaction    Slot = "faction"
	SlotDoctrine   Slot = "doctrine"
	SlotCommons    Slot = "commons"
	SlotTrust      Slot = "trust"
	SlotEventBuffs Slot = "event_buffs"
	SlotPrestige   Slot = "prestige"
)

var Order = [...]Slot{
	SlotUpgrades, SlotMilestones, SlotFaction, SlotDoctrine,
	SlotCommons, SlotTrust, SlotEventBuffs, SlotPrestige,
}

const WithinSlotOrder = "source_id_raw_byte_ascending"

type Contribution struct {
	Slot     Slot
	SourceID string
	Target   string
	Factor   decimal.Decimal
}

func ValidSlot(slot Slot) bool {
	for _, candidate := range Order {
		if slot == candidate {
			return true
		}
	}
	return false
}

// OrderedSourceIDs returns an independently-owned slice in the canonical
// within-slot order. Go string comparison is lexicographic over raw UTF-8
// bytes, which is the published cross-runtime contract.
func OrderedSourceIDs(sourceIDs []string) []string {
	ordered := append([]string(nil), sourceIDs...)
	sort.Slice(ordered, func(left, right int) bool {
		return ordered[left] < ordered[right]
	})
	return ordered
}

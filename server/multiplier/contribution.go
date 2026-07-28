// Package multiplier defines the neutral boundary between production and
// state-owning multiplier providers.
package multiplier

import "cloud-clicker/server/decimal"

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

package combat

// LaneOrderKey is C4's ordering projection, not a placement command or its
// legality check. C3 supplies the signed-int32 representation; the engine owns
// admissible ticks, player slots, sequences and duplicate-command handling.
type LaneOrderKey struct {
	Tick         int32
	PlayerSlot   int32
	PlacementSeq int32
}

// CompareLaneOrder compares (tick, player_slot, placement_seq) without
// subtracting int32 values, which could wrap at the representation edges.
// Equal keys compare equal; this does not choose a duplicate-placement policy.
func CompareLaneOrder(left, right LaneOrderKey) int {
	for _, pair := range [][2]int32{
		{left.Tick, right.Tick},
		{left.PlayerSlot, right.PlayerSlot},
		{left.PlacementSeq, right.PlacementSeq},
	} {
		if pair[0] < pair[1] {
			return -1
		}
		if pair[0] > pair[1] {
			return 1
		}
	}
	return 0
}

package minigame

// GardenPayoutValidator adapts the platform's six-key payout loader for the
// Server Garden persistent tenant (SG1 rule 9, SG-P1): the garden's payout row
// must load through LoadPayoutPolicy unchanged, against the bundle's
// resources and copy keys, with the garden's single declared score fact.
func GardenPayoutValidator(resourceIDs, copyKeys map[string]struct{}) func([]byte, map[string]struct{}) error {
	return func(data []byte, scoreFactIDs map[string]struct{}) error {
		_, err := LoadPayoutPolicy(data, PayoutDeclarations{ResourceIDs: resourceIDs, ScoreFactIDs: scoreFactIDs, CopyKeys: copyKeys})
		return err
	}
}

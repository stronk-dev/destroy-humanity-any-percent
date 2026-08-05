package pet

import "errors"

var ErrInvalidCombatInputs = errors.New("invalid pet combat inputs")

// CombatInputs is the complete Pet Care -> combat C5 boundary. Combat owns
// every formula that consumes these integers; Pet Care only supplies the
// replay-owned Trust value beside the Founder-owned Soul value.
type CombatInputs struct {
	PetTrustPPM int64 `json:"pet_trust_ppm"`
	Soul        int64 `json:"soul"`
}

func NewCombatInputs(petTrustPPM, soul int64) (CombatInputs, error) {
	if petTrustPPM < 0 || petTrustPPM > 1_000_000 || soul < -maxExactInteger || soul > maxExactInteger {
		return CombatInputs{}, ErrInvalidCombatInputs
	}
	return CombatInputs{PetTrustPPM: petTrustPPM, Soul: soul}, nil
}

func CombatInputsForState(state CareState, soul int64) (CombatInputs, error) {
	return NewCombatInputs(state.TrustPPM, soul)
}

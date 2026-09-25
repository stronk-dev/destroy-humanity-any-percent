package pet

import (
	"errors"
	"fmt"
	"regexp"

	"cloud-clicker/server/decimal"
)

// Identity is one adopted pet's immutable identity (Pet Adoption v1 PA3). It
// is keyed by pet ID in Founder save pet_identities, beside the mutable care
// record in pets.
type Identity struct {
	SpeciesID           string `json:"species_id"`
	Temperament         string `json:"temperament"`
	PaletteID           string `json:"palette_id"`
	NameKey             string `json:"name_key"`
	AdoptedAtMS         int64  `json:"adopted_at_ms"`
	AdoptedAtAttendedMS int64  `json:"adopted_at_attended_ms"`
}

var (
	ErrInvalidIdentity = errors.New("invalid pet identity")
	petIDV7Pattern     = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-7[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
)

// ValidateIdentityShape checks the catalog-free PA3 invariants: exact
// biconditional key sets with care, UUIDv7 IDs, well-formed fields, and an
// adoption coordinate no later than the care watermark.
func ValidateIdentityShape(identities map[string]Identity, care map[string]CareState) error {
	if identities == nil || len(identities) != len(care) {
		return fmt.Errorf("%w: identity and care key sets differ", ErrInvalidIdentity)
	}
	for id, identity := range identities {
		record, ok := care[id]
		if !ok || !petIDV7Pattern.MatchString(id) {
			return fmt.Errorf("%w: pet %q has no care record or is not UUIDv7", ErrInvalidIdentity, id)
		}
		if !speciesIDPattern.MatchString(identity.SpeciesID) || !contains(TemperamentOrder, identity.Temperament) ||
			!paletteIDPattern.MatchString(identity.PaletteID) || !nameKeyPattern.MatchString(identity.NameKey) {
			return fmt.Errorf("%w: pet %q has a malformed identity", ErrInvalidIdentity, id)
		}
		if identity.AdoptedAtMS < 0 || identity.AdoptedAtMS > decimal.MaxExactInteger || identity.AdoptedAtAttendedMS < 0 ||
			identity.AdoptedAtAttendedMS > record.EvaluatedThroughAttendedMS {
			return fmt.Errorf("%w: pet %q has an invalid adoption coordinate", ErrInvalidIdentity, id)
		}
	}
	return nil
}

// ValidateIdentitiesAgainst adds the catalog-aware PA3.3/PA3.4 invariants.
func ValidateIdentitiesAgainst(catalog *SpeciesCatalog, identities map[string]Identity) error {
	if catalog == nil {
		return fmt.Errorf("%w: pet_species is not pinned", ErrInvalidIdentity)
	}
	if int64(len(identities)) > catalog.MaxPetsPerFounder {
		return fmt.Errorf("%w: %d pets exceed the cap %d", ErrInvalidIdentity, len(identities), catalog.MaxPetsPerFounder)
	}
	for id, identity := range identities {
		row, ok := catalog.Row(identity.SpeciesID)
		if !ok || !row.HasTemperament(identity.Temperament) || !row.HasPalette(identity.PaletteID) || !row.HasName(identity.NameKey) {
			return fmt.Errorf("%w: pet %q does not resolve under the pinned pet_species", ErrInvalidIdentity, id)
		}
	}
	return nil
}

func CloneIdentities(values map[string]Identity) map[string]Identity {
	if values == nil {
		return nil
	}
	result := make(map[string]Identity, len(values))
	for id, identity := range values {
		result[id] = identity
	}
	return result
}

// InitialCareState is PA4.6.2's canonical initial care record for a pet
// adopted at the frozen attendance total attendedMS: stats at their initial
// values, zero remainders, no cooldowns, Trust at its initial value, and a
// watermark at attendedMS so pre-adoption attendance never decays the pet.
func InitialCareState(catalog *Catalog, attendedMS int64) (CareState, error) {
	if catalog == nil || attendedMS < 0 || attendedMS > decimal.MaxExactInteger {
		return CareState{}, fmt.Errorf("%w: invalid initial care inputs", ErrInvalidIdentity)
	}
	stats, remainders := map[StatID]int64{}, map[StatID]int64{}
	for _, row := range catalog.StatPolicy.Stats {
		stats[row.StatID], remainders[row.StatID] = row.InitialPPM, 0
	}
	return CareState{StatsPPM: stats, StatDecayRemaindersPPM: remainders, CooldownUntilAttendedMS: map[string]int64{},
		TrustPPM: catalog.TrustPolicy.InitialPPM, EvaluatedThroughAttendedMS: attendedMS, BehaviorState: BehaviorIdle,
		BehaviorEnteredAtAttendedMS: attendedMS, BehaviorQueue: []BehaviorQueueEntry{}}, nil
}

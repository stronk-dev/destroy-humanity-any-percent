package cosmetic

import (
	"errors"
	"fmt"
)

// ErrInvalidState is a Founder `cosmetics` object that violates §3.
var ErrInvalidState = errors.New("invalid cosmetics Founder state")

// State is Founder save `cosmetics` (§3): the byte-sorted owned ids and the
// per-pet equip map. Empty collections are canonical [] / {}, never null.
type State struct {
	Owned    []string          `json:"owned"`
	Equipped map[string]string `json:"equipped"`
}

// NewState is the activation value (§6): nothing owned, nothing equipped.
func NewState() *State {
	return &State{Owned: []string{}, Equipped: map[string]string{}}
}

// Clone deep-copies state; nil stays nil.
func (state *State) Clone() *State {
	if state == nil {
		return nil
	}
	clone := &State{Owned: append([]string{}, state.Owned...), Equipped: make(map[string]string, len(state.Equipped))}
	for pet, id := range state.Equipped {
		clone.Equipped[pet] = id
	}
	return clone
}

// Owns reports whether id is in the owned set.
func (state *State) Owns(id string) bool {
	if state == nil {
		return false
	}
	for _, owned := range state.Owned {
		if owned == id {
			return true
		}
	}
	return false
}

// Equal reports byte-level equality of the two objects.
func (state *State) Equal(other *State) bool {
	if state == nil || other == nil {
		return state == other
	}
	if len(state.Owned) != len(other.Owned) || len(state.Equipped) != len(other.Equipped) {
		return false
	}
	for index := range state.Owned {
		if state.Owned[index] != other.Owned[index] {
			return false
		}
	}
	for pet, id := range state.Equipped {
		if other.Equipped[pet] != id {
			return false
		}
	}
	return true
}

// ValidateShape checks the catalog-free §3 invariants: non-nil collections,
// strictly ascending mechanical owned ids, every equip key a pet, and every
// equipped id owned.
func ValidateShape(state *State, pets map[string]struct{}) error {
	if state == nil || state.Owned == nil || state.Equipped == nil {
		return fmt.Errorf("%w: owned and equipped are required and never null", ErrInvalidState)
	}
	for index, id := range state.Owned {
		if !mechanicalPattern.MatchString(id) || index > 0 && state.Owned[index-1] >= id {
			return fmt.Errorf("%w: owned ids must be mechanical, unique, and byte-sorted", ErrInvalidState)
		}
	}
	for pet, id := range state.Equipped {
		if _, ok := pets[pet]; !ok {
			return fmt.Errorf("%w: equipped key %q is not a pet", ErrInvalidState, pet)
		}
		if !state.Owns(id) {
			return fmt.Errorf("%w: pet %q wears unowned %q", ErrInvalidState, pet, id)
		}
	}
	return nil
}

// ValidateAgainst adds the catalog-aware §3 invariants: every owned id is
// pinned, the owned set is bounded by the catalog, and equipped ids use the
// pet slot.
func ValidateAgainst(catalog *Catalog, state *State) error {
	if catalog == nil {
		return fmt.Errorf("%w: cosmetics is not pinned", ErrInvalidState)
	}
	if len(state.Owned) > len(catalog.Items) {
		return fmt.Errorf("%w: owned set exceeds the catalog", ErrInvalidState)
	}
	for _, id := range state.Owned {
		if _, ok := catalog.Item(id); !ok {
			return fmt.Errorf("%w: owned %q is not in the pinned catalog", ErrInvalidState, id)
		}
	}
	for pet, id := range state.Equipped {
		if item, _ := catalog.Item(id); item.Slot != SlotPet {
			return fmt.Errorf("%w: pet %q wears a non-pet cosmetic", ErrInvalidState, pet)
		}
	}
	return nil
}

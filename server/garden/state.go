package garden

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"sort"
)

var (
	ErrInvalidState = errors.New("invalid server_garden state")

	saltPattern = regexp.MustCompile(`^[0-9a-f]{16}$`)
)

// Plot is one occupied rack slot. Stage is derived, never stored:
// mature ⟺ MaturedEffectPPM != nil (SG2).
type Plot struct {
	Row              int64  `json:"row"`
	Col              int64  `json:"col"`
	SpeciesID        string `json:"species_id"`
	AgeTicks         int64  `json:"age_ticks"`
	MaturedEffectPPM *int64 `json:"matured_effect_ppm"`
}

// Mature reports the derived stage.
func (plot Plot) Mature() bool { return plot.MaturedEffectPPM != nil }

// State is the Founder `server_garden` object (SG2). The salt never leaves the
// server (OD-12); read projections and receipts must omit it.
type State struct {
	SaltHex            *string  `json:"salt_hex"`
	TickAnchorWallMS   *int64   `json:"tick_anchor_wall_ms"`
	TickSeq            int64    `json:"tick_seq"`
	SubstrateID        string   `json:"substrate_id"`
	SubstrateSetWallMS *int64   `json:"substrate_set_wall_ms"`
	Plots              []Plot   `json:"plots"`
	SeedCollection     []string `json:"seed_collection"`
}

// NewState is the SG2 activation value: empty plots, every starter collected,
// the default substrate, and a dormant clock.
func NewState(catalog *Catalog) *State {
	return &State{SubstrateID: catalog.DefaultSubstrateID, Plots: []Plot{}, SeedCollection: catalog.Starters()}
}

// Clone deep-copies the state; nil stays nil.
func (state *State) Clone() *State {
	if state == nil {
		return nil
	}
	clone := &State{TickSeq: state.TickSeq, SubstrateID: state.SubstrateID,
		SaltHex: cloneString(state.SaltHex), TickAnchorWallMS: cloneInt(state.TickAnchorWallMS), SubstrateSetWallMS: cloneInt(state.SubstrateSetWallMS),
		Plots: make([]Plot, len(state.Plots)), SeedCollection: append([]string{}, state.SeedCollection...)}
	for index, plot := range state.Plots {
		plot.MaturedEffectPPM = cloneInt(plot.MaturedEffectPPM)
		clone.Plots[index] = plot
	}
	return clone
}

// Equal compares canonical encodings.
func (state *State) Equal(other *State) bool {
	if state == nil || other == nil {
		return state == other
	}
	left, leftErr := json.Marshal(state)
	right, rightErr := json.Marshal(other)
	return leftErr == nil && rightErr == nil && string(left) == string(right)
}

// Plot returns the plot at (row, col).
func (state *State) Plot(row, col int64) (Plot, int, bool) {
	index := sort.Search(len(state.Plots), func(i int) bool { return !plotLess(state.Plots[i].Row, state.Plots[i].Col, row, col) })
	if index < len(state.Plots) && state.Plots[index].Row == row && state.Plots[index].Col == col {
		return state.Plots[index], index, true
	}
	return Plot{}, index, false
}

// Collected reports whether speciesID is in the seed collection.
func (state *State) Collected(speciesID string) bool {
	index := sort.SearchStrings(state.SeedCollection, speciesID)
	return index < len(state.SeedCollection) && state.SeedCollection[index] == speciesID
}

func (state *State) insertPlot(plot Plot) {
	_, index, _ := state.Plot(plot.Row, plot.Col)
	state.Plots = append(state.Plots, Plot{})
	copy(state.Plots[index+1:], state.Plots[index:])
	state.Plots[index] = plot
}

func (state *State) removePlot(index int) {
	state.Plots = append(state.Plots[:index], state.Plots[index+1:]...)
}

func (state *State) collect(speciesID string) bool {
	index := sort.SearchStrings(state.SeedCollection, speciesID)
	if index < len(state.SeedCollection) && state.SeedCollection[index] == speciesID {
		return false
	}
	state.SeedCollection = append(state.SeedCollection, "")
	copy(state.SeedCollection[index+1:], state.SeedCollection[index:])
	state.SeedCollection[index] = speciesID
	return true
}

func plotLess(leftRow, leftCol, rightRow, rightCol int64) bool {
	return leftRow < rightRow || leftRow == rightRow && leftCol < rightCol
}

// ValidateShape is the catalog-free SG2 codec rule set.
func ValidateShape(state *State) error {
	if state == nil {
		return fmt.Errorf("%w: server_garden is required", ErrInvalidState)
	}
	if state.SaltHex != nil && !saltPattern.MatchString(*state.SaltHex) {
		return fmt.Errorf("%w: salt_hex must be 16 lowercase hex characters", ErrInvalidState)
	}
	for _, value := range []*int64{state.TickAnchorWallMS, state.SubstrateSetWallMS} {
		if value != nil && (*value < 0 || *value > maxExactInteger) {
			return fmt.Errorf("%w: wall stamps must be exact non-negative integers", ErrInvalidState)
		}
	}
	if state.TickSeq < 0 || state.TickSeq > maxExactInteger || !idPattern.MatchString(state.SubstrateID) || state.Plots == nil || state.SeedCollection == nil {
		return fmt.Errorf("%w: scalar fields", ErrInvalidState)
	}
	for index, plot := range state.Plots {
		if plot.Row < 0 || plot.Row >= MaxGridSide || plot.Col < 0 || plot.Col >= MaxGridSide || !idPattern.MatchString(plot.SpeciesID) ||
			plot.AgeTicks < 0 || plot.AgeTicks > MaxAgeTicks || plot.MaturedEffectPPM != nil && (*plot.MaturedEffectPPM < 0 || *plot.MaturedEffectPPM > MaxEffectPPM) {
			return fmt.Errorf("%w: plot domain", ErrInvalidState)
		}
		if index > 0 && !plotLess(state.Plots[index-1].Row, state.Plots[index-1].Col, plot.Row, plot.Col) {
			return fmt.Errorf("%w: plots must be sorted by (row, col) and unique", ErrInvalidState)
		}
	}
	for index, species := range state.SeedCollection {
		if !idPattern.MatchString(species) || index > 0 && state.SeedCollection[index-1] >= species {
			return fmt.Errorf("%w: seed_collection must be sorted and unique", ErrInvalidState)
		}
	}
	return nil
}

// ValidateAgainst binds the state to the pinned catalog. An unknown ID is an
// invariant failure, never a silent drop (SG1 rule 11).
func ValidateAgainst(catalog *Catalog, state *State) error {
	if err := ValidateShape(state); err != nil {
		return err
	}
	if _, ok := catalog.Substrate(state.SubstrateID); !ok {
		return fmt.Errorf("%w: unknown substrate %q", ErrInvalidState, state.SubstrateID)
	}
	for _, plot := range state.Plots {
		if _, ok := catalog.SpeciesRow(plot.SpeciesID); !ok {
			return fmt.Errorf("%w: unknown plot species %q", ErrInvalidState, plot.SpeciesID)
		}
	}
	for _, species := range state.SeedCollection {
		if _, ok := catalog.SpeciesRow(species); !ok {
			return fmt.Errorf("%w: unknown collected species %q", ErrInvalidState, species)
		}
	}
	for _, starter := range catalog.Starters() {
		if !state.Collected(starter) {
			return fmt.Errorf("%w: seed_collection must include every starter", ErrInvalidState)
		}
	}
	return nil
}

func cloneString(value *string) *string {
	if value == nil {
		return nil
	}
	copy := *value
	return &copy
}

func cloneInt(value *int64) *int64 {
	if value == nil {
		return nil
	}
	copy := *value
	return &copy
}

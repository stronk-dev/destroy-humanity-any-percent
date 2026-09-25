package garden

import (
	"errors"
	"fmt"
	"strconv"

	"cloud-clicker/server/determinism"
)

var ErrInvalidAdvance = errors.New("invalid server_garden advance")

const (
	founderSubstreamLabel = "garden.founder.v1"
	tickSubstreamLabel    = "garden.tick.v1"
)

// PlotRef names a matured plant in an advance summary.
type PlotRef struct {
	Row       int64  `json:"row"`
	Col       int64  `json:"col"`
	SpeciesID string `json:"species_id"`
}

// Spawn names a seedling placed by spread or mutation.
type Spawn struct {
	Row       int64  `json:"row"`
	Col       int64  `json:"col"`
	SpeciesID string `json:"species_id"`
	RecipeID  string `json:"recipe_id"`
}

// Advance is the SG8 `garden_advanced.v1` payload and the advance summary in
// resolved inputs. CatchupForfeitedMS is never truncated silently (SG3).
type Advance struct {
	TicksApplied       int64     `json:"ticks_applied"`
	TickSeqAfter       int64     `json:"tick_seq_after"`
	CatchupForfeitedMS int64     `json:"catchup_forfeited_ms"`
	CatchupReasonKey   *string   `json:"catchup_reason_key"`
	Matured            []PlotRef `json:"matured"`
	Spawned            []Spawn   `json:"spawned"`
}

// Visible reports whether SG8 emits `garden_advanced.v1` for this advance.
func (advance Advance) Visible() bool {
	return advance.TicksApplied > 0 && (len(advance.Matured) > 0 || len(advance.Spawned) > 0) || advance.CatchupForfeitedMS > 0
}

// AdvanceInput carries everything SG3 reads besides state and catalog. Salt
// is the server-drawn value frozen in resolved inputs; it is required exactly
// when the advance initializes it (NeedsSalt) and must be empty otherwise.
type AdvanceInput struct {
	ServerMS  int64
	HostLevel int64
	Unlocked  bool
	Salt      string
}

// NeedsSalt reports whether an advance at these inputs draws the hidden salt.
func NeedsSalt(state *State, unlocked bool) bool {
	return unlocked && state != nil && state.SaltHex == nil
}

// Advance is SG3's lazy wall-clock advance: exact, integer-only, keyed by the
// absolute tick sequence so it is partition-invariant below the cap.
func (catalog *Catalog) Advance(state *State, input AdvanceInput) (Advance, error) {
	summary := Advance{Matured: []PlotRef{}, Spawned: []Spawn{}}
	if catalog == nil || state == nil || input.ServerMS < 0 || input.ServerMS > maxExactInteger || input.HostLevel < 0 {
		return Advance{}, ErrInvalidAdvance
	}
	if input.Salt != "" != NeedsSalt(state, input.Unlocked) {
		return Advance{}, fmt.Errorf("%w: salt must be supplied exactly when it is initialized", ErrInvalidAdvance)
	}
	summary.TickSeqAfter = state.TickSeq
	if !input.Unlocked {
		return summary, nil
	}
	if state.SaltHex == nil {
		if !saltPattern.MatchString(input.Salt) {
			return Advance{}, fmt.Errorf("%w: salt grammar", ErrInvalidAdvance)
		}
		salt := input.Salt
		state.SaltHex = &salt
	}
	if state.TickAnchorWallMS == nil {
		anchor := input.ServerMS
		state.TickAnchorWallMS = &anchor
		return summary, nil
	}
	anchor := *state.TickAnchorWallMS
	if input.ServerMS <= anchor {
		return summary, nil
	}
	elapsed := input.ServerMS - anchor
	if elapsed > catalog.CatchupCapMS {
		summary.CatchupForfeitedMS = elapsed - catalog.CatchupCapMS
		anchor += summary.CatchupForfeitedMS
		elapsed = catalog.CatchupCapMS
		reason := catalog.CatchupReasonKey
		summary.CatchupReasonKey = &reason
	}
	substrate, ok := catalog.Substrate(state.SubstrateID)
	if !ok {
		return Advance{}, fmt.Errorf("%w: unknown substrate", ErrInvalidAdvance)
	}
	ticks := elapsed / substrate.TickMS
	dimension := catalog.Dimension(input.HostLevel)
	base, err := gardenBase(*state.SaltHex)
	if err != nil {
		return Advance{}, err
	}
	for step := int64(1); step <= ticks; step++ {
		if catalog.fixedPoint(state, dimension) {
			break
		}
		catalog.tick(state, dimension, substrate, base, state.TickSeq+step, &summary)
	}
	state.TickSeq += ticks
	anchor += ticks * substrate.TickMS
	state.TickAnchorWallMS = &anchor
	summary.TicksApplied = ticks
	summary.TickSeqAfter = state.TickSeq
	return summary, nil
}

func gardenBase(saltHex string) (uint64, error) {
	salt, err := strconv.ParseUint(saltHex, 16, 64)
	if err != nil || !saltPattern.MatchString(saltHex) {
		return 0, fmt.Errorf("%w: salt", ErrInvalidAdvance)
	}
	return determinism.Substream(salt, founderSubstreamLabel).Next(), nil
}

func active(dimension DimensionRow, row, col int64) bool {
	return row < dimension.Height && col < dimension.Width
}

// fixedPoint reports SG3's exact skip condition: no growing plant in the
// active region and no empty active plot with an eligible recipe. Such ticks
// consume no draws, so skipping them is exact.
func (catalog *Catalog) fixedPoint(state *State, dimension DimensionRow) bool {
	for _, plot := range state.Plots {
		if active(dimension, plot.Row, plot.Col) && !plot.Mature() {
			return false
		}
	}
	for row := int64(0); row < dimension.Height; row++ {
		for col := int64(0); col < dimension.Width; col++ {
			if _, _, occupied := state.Plot(row, col); occupied {
				continue
			}
			if len(catalog.eligible(census(state, dimension, row, col))) > 0 {
				return false
			}
		}
	}
	return true
}

// census is SG4 step 2 for one empty plot: Moore neighbours holding a mature
// plant, per species, inside the active region only.
func census(state *State, dimension DimensionRow, row, col int64) map[string]int64 {
	counts := map[string]int64{}
	for dr := int64(-1); dr <= 1; dr++ {
		for dc := int64(-1); dc <= 1; dc++ {
			if dr == 0 && dc == 0 {
				continue
			}
			r, c := row+dr, col+dc
			if r < 0 || c < 0 || !active(dimension, r, c) {
				continue
			}
			if plot, _, ok := state.Plot(r, c); ok && plot.Mature() {
				counts[plot.SpeciesID]++
			}
		}
	}
	return counts
}

func (catalog *Catalog) eligible(counts map[string]int64) []Recipe {
	recipes := []Recipe{}
	for _, recipe := range catalog.Recipes {
		ok := true
		for _, parent := range recipe.Parents {
			ok = ok && counts[parent.SpeciesID] >= parent.MinMatureNeighbours
		}
		if ok {
			recipes = append(recipes, recipe)
		}
	}
	return recipes
}

// tick is SG4 for absolute tick s: growth, then one census, then spread and
// mutation walking empty active plots in (row, col) order.
func (catalog *Catalog) tick(state *State, dimension DimensionRow, substrate Substrate, base uint64, s int64, summary *Advance) {
	for index := range state.Plots {
		plot := &state.Plots[index]
		if !active(dimension, plot.Row, plot.Col) || plot.Mature() {
			continue
		}
		plot.AgeTicks = min(plot.AgeTicks+1, MaxAgeTicks)
		species, _ := catalog.SpeciesRow(plot.SpeciesID)
		if plot.AgeTicks >= species.MaturationTicks {
			effect := substrate.EffectPPM
			plot.MaturedEffectPPM = &effect
			summary.Matured = append(summary.Matured, PlotRef{Row: plot.Row, Col: plot.Col, SpeciesID: plot.SpeciesID})
		}
	}
	type candidate struct {
		row, col int64
		recipes  []Recipe
	}
	candidates := []candidate{}
	for row := int64(0); row < dimension.Height; row++ {
		for col := int64(0); col < dimension.Width; col++ {
			if _, _, occupied := state.Plot(row, col); occupied {
				continue
			}
			if recipes := catalog.eligible(census(state, dimension, row, col)); len(recipes) > 0 {
				candidates = append(candidates, candidate{row: row, col: col, recipes: recipes})
			}
		}
	}
	if len(candidates) == 0 {
		return
	}
	random := determinism.Substream(base^uint64(s), tickSubstreamLabel)
	for _, cell := range candidates {
		draw := int64(random.Bound(uint64(PPM)))
		cumulative := int64(0)
		for _, recipe := range cell.recipes {
			effective := min(PPM, recipe.ChancePPM*substrate.ChanceFactorPPM/PPM)
			cumulative = min(PPM, cumulative+effective)
			if draw < cumulative {
				state.insertPlot(Plot{Row: cell.row, Col: cell.col, SpeciesID: recipe.ChildSpeciesID})
				summary.Spawned = append(summary.Spawned, Spawn{Row: cell.row, Col: cell.col, SpeciesID: recipe.ChildSpeciesID, RecipeID: recipe.RecipeID})
				break
			}
		}
	}
}

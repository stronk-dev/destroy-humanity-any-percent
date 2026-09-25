package garden

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
)

// Rejection is a closed SG5 rejection: an existing category with a closed
// garden detail. A rejection mutates nothing.
type Rejection struct {
	Category string
	Detail   string
}

func (rejection *Rejection) Error() string {
	return "garden rejection " + rejection.Category + "/" + rejection.Detail
}

func reject(category, detail string) error { return &Rejection{Category: category, Detail: detail} }

// AsRejection unwraps a closed garden rejection.
func AsRejection(err error) (*Rejection, bool) {
	var rejection *Rejection
	ok := errors.As(err, &rejection)
	return rejection, ok
}

// Closed SG5 details (the category is part of each pair).
const (
	DetailInactive         = "garden_inactive"
	DetailUnlockRequired   = "fiscal_unlock_required"
	DetailHumanLocked      = "human_content_locked"
	DetailPlotDormant      = "plot_dormant"
	DetailPlotOccupied     = "plot_occupied"
	DetailUnknownSpecies   = "garden_species"
	DetailSeedNotCollected = "seed_not_collected"
	DetailUnknownPlot      = "garden_plot"
	DetailNotMature        = "plant_not_mature"
	DetailUnknownSubstrate = "garden_substrate"
	DetailUnchanged        = "substrate_unchanged"
	DetailLockout          = "substrate_lockout"
)

// Gate is SG5 validation steps 2–3, run after the advance: the unlock and,
// under a human_hobby soul gate, near-zero Soul.
func (catalog *Catalog) Gate(unlocked, humanContentLocked bool) error {
	if !unlocked {
		return reject("not_eligible", DetailUnlockRequired)
	}
	if catalog.SoulGate == SoulGateHumanHobby && humanContentLocked {
		return reject("not_eligible", DetailHumanLocked)
	}
	return nil
}

// Planted is the `garden_planted.v1` payload.
type Planted struct {
	Row       int64  `json:"row"`
	Col       int64  `json:"col"`
	SpeciesID string `json:"species_id"`
}

// Plant places a free seedling (OD-9).
func (catalog *Catalog) Plant(state *State, hostLevel, row, col int64, speciesID string) (Planted, error) {
	if !active(catalog.Dimension(hostLevel), row, col) {
		return Planted{}, reject("not_eligible", DetailPlotDormant)
	}
	if _, _, occupied := state.Plot(row, col); occupied {
		return Planted{}, reject("not_eligible", DetailPlotOccupied)
	}
	if _, ok := catalog.SpeciesRow(speciesID); !ok {
		return Planted{}, reject("unknown_id", DetailUnknownSpecies)
	}
	if !state.Collected(speciesID) {
		return Planted{}, reject("not_eligible", DetailSeedNotCollected)
	}
	state.insertPlot(Plot{Row: row, Col: col, SpeciesID: speciesID})
	return Planted{Row: row, Col: col, SpeciesID: speciesID}, nil
}

// Uprooted is the `garden_uprooted.v1` payload.
type Uprooted struct {
	Row       int64  `json:"row"`
	Col       int64  `json:"col"`
	SpeciesID string `json:"species_id"`
	WasMature bool   `json:"was_mature"`
}

// Uproot removes a plant with no payout and no seed; dormant plots allowed.
func (catalog *Catalog) Uproot(state *State, row, col int64) (Uprooted, error) {
	plot, index, ok := state.Plot(row, col)
	if !ok {
		return Uprooted{}, reject("unknown_id", DetailUnknownPlot)
	}
	state.removePlot(index)
	return Uprooted{Row: row, Col: col, SpeciesID: plot.SpeciesID, WasMature: plot.Mature()}, nil
}

// SubstrateSet is the `garden_substrate_set.v1` payload.
type SubstrateSet struct {
	FromSubstrateID        string `json:"from_substrate_id"`
	ToSubstrateID          string `json:"to_substrate_id"`
	PartialTickForfeitedMS int64  `json:"partial_tick_forfeited_ms"`
}

// SetSubstrate switches the whole-garden substrate: re-anchors at serverMS
// (the partial tick is forfeited and reported) and stamps the lockout.
func (catalog *Catalog) SetSubstrate(state *State, serverMS int64, substrateID string) (SubstrateSet, error) {
	if _, ok := catalog.Substrate(substrateID); !ok {
		return SubstrateSet{}, reject("unknown_id", DetailUnknownSubstrate)
	}
	if substrateID == state.SubstrateID {
		return SubstrateSet{}, reject("not_eligible", DetailUnchanged)
	}
	if state.SubstrateSetWallMS != nil && serverMS-*state.SubstrateSetWallMS < catalog.SubstrateLockoutMS {
		return SubstrateSet{}, reject("not_eligible", DetailLockout)
	}
	forfeited := int64(0)
	if state.TickAnchorWallMS != nil && serverMS > *state.TickAnchorWallMS {
		forfeited = serverMS - *state.TickAnchorWallMS
	}
	result := SubstrateSet{FromSubstrateID: state.SubstrateID, ToSubstrateID: substrateID, PartialTickForfeitedMS: forfeited}
	anchor, stamp := serverMS, serverMS
	state.SubstrateID, state.TickAnchorWallMS, state.SubstrateSetWallMS = substrateID, &anchor, &stamp
	return result, nil
}

// HarvestTarget is one requested plot.
type HarvestTarget struct {
	Row int64 `json:"row"`
	Col int64 `json:"col"`
}

// HarvestedPlot is one harvested plant with its frozen-effect units.
type HarvestedPlot struct {
	Col       int64  `json:"col"`
	Row       int64  `json:"row"`
	SpeciesID string `json:"species_id"`
	Units     int64  `json:"units"`
}

// Harvest is SG6's Founder side: remove every listed mature plant, collect new
// seeds, and compute units and the cross-stream harvest hash.
type Harvest struct {
	Plots           []HarvestedPlot `json:"plots"`
	TotalUnits      int64           `json:"total_units"`
	SeedsDiscovered []string        `json:"seeds_discovered"`
	HarvestHash     string          `json:"harvest_hash"`
}

// ValidHarvestTargets is the decode-time rule: 1–36 entries, (row, col)
// sorted, duplicate-free, inside [0,5].
func ValidHarvestTargets(targets []HarvestTarget) bool {
	if len(targets) < 1 || len(targets) > int(MaxGridSide*MaxGridSide) {
		return false
	}
	for index, target := range targets {
		if target.Row < 0 || target.Row >= MaxGridSide || target.Col < 0 || target.Col >= MaxGridSide ||
			index > 0 && !plotLess(targets[index-1].Row, targets[index-1].Col, target.Row, target.Col) {
			return false
		}
	}
	return true
}

// Harvest applies SG6's Founder side. Dormant mature plants may be harvested.
func (catalog *Catalog) Harvest(state *State, intentID string, targets []HarvestTarget) (Harvest, error) {
	if !ValidHarvestTargets(targets) {
		return Harvest{}, fmt.Errorf("%w: harvest targets", ErrInvalidAdvance)
	}
	// SG5 order: any empty listed plot first, then any immature listed plant.
	for _, target := range targets {
		if _, _, ok := state.Plot(target.Row, target.Col); !ok {
			return Harvest{}, reject("unknown_id", DetailUnknownPlot)
		}
	}
	for _, target := range targets {
		if plot, _, _ := state.Plot(target.Row, target.Col); !plot.Mature() {
			return Harvest{}, reject("not_eligible", DetailNotMature)
		}
	}
	harvest := Harvest{Plots: []HarvestedPlot{}, SeedsDiscovered: []string{}}
	for _, target := range targets {
		plot, index, _ := state.Plot(target.Row, target.Col)
		species, _ := catalog.SpeciesRow(plot.SpeciesID)
		units := species.HarvestUnits * *plot.MaturedEffectPPM / PPM
		harvest.Plots = append(harvest.Plots, HarvestedPlot{Row: plot.Row, Col: plot.Col, SpeciesID: plot.SpeciesID, Units: units})
		harvest.TotalUnits += units
		state.removePlot(index)
		if state.collect(plot.SpeciesID) {
			harvest.SeedsDiscovered = insertSorted(harvest.SeedsDiscovered, plot.SpeciesID)
		}
	}
	hash, err := HarvestHash(intentID, harvest.Plots, harvest.TotalUnits)
	if err != nil {
		return Harvest{}, err
	}
	harvest.HarvestHash = hash
	return harvest, nil
}

// HarvestHash is SG6's cross-stream binding: SHA-256 over the canonical
// (byte-sorted key) JSON {intent_id, plots, total_units}.
func HarvestHash(intentID string, plots []HarvestedPlot, totalUnits int64) (string, error) {
	canonical, err := json.Marshal(struct {
		IntentID   string          `json:"intent_id"`
		Plots      []HarvestedPlot `json:"plots"`
		TotalUnits int64           `json:"total_units"`
	}{IntentID: intentID, Plots: plots, TotalUnits: totalUnits})
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256(canonical)
	return "sha256:" + hex.EncodeToString(digest[:]), nil
}

func insertSorted(values []string, value string) []string {
	result := make([]string, 0, len(values)+1)
	inserted := false
	for _, existing := range values {
		if !inserted && value < existing {
			result = append(result, value)
			inserted = true
		}
		result = append(result, existing)
	}
	if !inserted {
		result = append(result, value)
	}
	return result
}

package production

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"cloud-clicker/server/economy"
	"cloud-clicker/server/garden"
	"cloud-clicker/server/save"
)

// Server Garden SG9: the advisory read projection. It runs the SG3 advance on
// a discarded clone and never contains the salt, a draw, or a future tick.

type gardenViewPlot struct {
	Row             int64  `json:"row"`
	Col             int64  `json:"col"`
	SpeciesID       string `json:"species_id"`
	Stage           string `json:"stage"`
	AgeTicks        int64  `json:"age_ticks"`
	MaturationTicks int64  `json:"maturation_ticks"`
	Dormant         bool   `json:"dormant"`
}

type gardenView struct {
	Width                     int64            `json:"width"`
	Height                    int64            `json:"height"`
	MaxWidth                  int64            `json:"max_width"`
	MaxHeight                 int64            `json:"max_height"`
	SubstrateID               string           `json:"substrate_id"`
	SubstrateLockoutUntilMS   *int64           `json:"substrate_lockout_until_ms"`
	NextTickWallMS            *int64           `json:"next_tick_wall_ms"`
	TickSeq                   int64            `json:"tick_seq"`
	PendingCatchupForfeitedMS int64            `json:"pending_catchup_forfeited_ms"`
	SeedCollection            []string         `json:"seed_collection"`
	SpeciesTotal              int64            `json:"species_total"`
	Plots                     []gardenViewPlot `json:"plots"`
}

// projectionSalt is a placeholder used only on the discarded clone. An
// unsalted garden is also unanchored, so the projection applies zero ticks and
// the value never influences any draw.
const projectionSalt = "0000000000000000"

// ProjectGardenView is the pure SG9 projection of a Founder's garden at serverMS.
func ProjectGardenView(bundle CatalogBundle, founder *save.State, founderRevision int64, serverMS int64) (json.RawMessage, error) {
	if !gardenActive(bundle, founder) {
		return json.Marshal(map[string]any{"kind": "inactive"})
	}
	catalog := bundle.Garden
	hostLevel, unlocked := gardenInputs(catalog, founder)
	if !unlocked {
		return json.Marshal(map[string]any{"kind": "locked", "unlock_id": catalog.UnlockID})
	}
	clone := founder.ServerGarden.Clone()
	salt := ""
	if garden.NeedsSalt(clone, true) {
		if clone.TickAnchorWallMS != nil {
			return nil, fmt.Errorf("%w: unsalted garden with a clock anchor", ErrInvalidEngineState)
		}
		salt = projectionSalt
	}
	advance, err := catalog.Advance(clone, garden.AdvanceInput{ServerMS: serverMS, HostLevel: hostLevel, Unlocked: true, Salt: salt})
	if err != nil {
		return nil, err
	}
	dimension := catalog.Dimension(hostLevel)
	view := gardenView{Width: dimension.Width, Height: dimension.Height, MaxWidth: catalog.MaxWidth, MaxHeight: catalog.MaxHeight,
		SubstrateID: clone.SubstrateID, TickSeq: clone.TickSeq, PendingCatchupForfeitedMS: advance.CatchupForfeitedMS,
		SeedCollection: append([]string{}, clone.SeedCollection...), SpeciesTotal: int64(len(catalog.Species)), Plots: []gardenViewPlot{}}
	if clone.SubstrateSetWallMS != nil {
		if until := *clone.SubstrateSetWallMS + catalog.SubstrateLockoutMS; until > serverMS {
			view.SubstrateLockoutUntilMS = &until
		}
	}
	if clone.TickAnchorWallMS != nil {
		substrate, _ := catalog.Substrate(clone.SubstrateID)
		next := *clone.TickAnchorWallMS + substrate.TickMS
		view.NextTickWallMS = &next
	}
	for _, plot := range clone.Plots {
		species, _ := catalog.SpeciesRow(plot.SpeciesID)
		stage := "growing"
		if plot.Mature() {
			stage = "mature"
		}
		view.Plots = append(view.Plots, gardenViewPlot{Row: plot.Row, Col: plot.Col, SpeciesID: plot.SpeciesID, Stage: stage,
			AgeTicks: plot.AgeTicks, MaturationTicks: species.MaturationTicks, Dormant: plot.Row >= dimension.Height || plot.Col >= dimension.Width})
	}
	return json.Marshal(map[string]any{"kind": "active", "founder_revision": founderRevision, "server_ms": serverMS, "garden": view})
}

// GardenView is the SG9 read for the Founder owning companyStreamID.
func (s *Service) GardenView(ctx context.Context, companyStreamID string, now time.Time) (json.RawMessage, error) {
	if s == nil || s.replayCatalogs == nil {
		return nil, fmt.Errorf("%w: Founder replay runtime unavailable", ErrInvalidIntent)
	}
	founder, err := s.store.LoadSiblingLatest(ctx, companyStreamID, economy.ScopeFounder)
	if err != nil {
		return nil, err
	}
	bundle, ok := s.replayCatalogs.ResolveReplayCatalogs(founder.Revision.ConstantsHash)
	if !ok {
		return nil, fmt.Errorf("%w: replay catalog bundle unavailable", ErrInvalidIntent)
	}
	return ProjectGardenView(bundle, founder.State, founder.Revision.Number, save.CanonicalServerTime(now).UnixMilli())
}

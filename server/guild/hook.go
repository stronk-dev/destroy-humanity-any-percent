package guild

import (
	"encoding/json"
	"errors"

	"cloud-clicker/server/accrualhook"
	"cloud-clicker/server/decimal"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/multiplier"
	"cloud-clicker/server/save"
)

const StockConsumptionSourceID = "guild.stock_consumption"

var ErrInvalidTithe = errors.New("invalid guild tithe")

type CatalogResolver interface {
	ResolveGuild(constantsHash string) (*Catalog, bool)
}

type AccrualHook struct{ Catalogs CatalogResolver }

func (hook AccrualHook) AfterAccrual(state *save.State, _ *economy.Catalog, revision save.Revision, result accrualhook.Result, contributions []multiplier.Contribution) ([]save.EventWrite, error) {
	if state == nil || hook.Catalogs == nil || result.ProgressDeltaPPM < 0 || result.ProgressDeltaPPM > 1_000_000 {
		return nil, ErrInvalidTithe
	}
	member := false
	for _, contribution := range contributions {
		if contribution.SourceID == StockConsumptionSourceID {
			if member {
				return nil, ErrInvalidTithe
			}
			member = true
		}
	}
	if !member {
		return nil, nil
	}
	catalog, ok := hook.Catalogs.ResolveGuild(revision.ConstantsHash)
	if !ok || catalog == nil || state.GuildTitheCarryPPM < 0 || state.GuildTitheCarryPPM >= 1_000_000 {
		return nil, ErrInvalidTithe
	}
	numerator := result.ProgressDeltaPPM*catalog.GuildTithePPM + state.GuildTitheCarryPPM
	xp := numerator / 1_000_000
	state.GuildTitheCarryPPM = numerator % 1_000_000
	if xp > decimal.MaxExactInteger {
		return nil, ErrInvalidTithe
	}
	payload, _ := json.Marshal(map[string]any{
		"founder_id":         revision.OwnerID,
		"run_id":             map[string]any{"company_stream_id": revision.StreamID, "run_seq": state.RunSeq},
		"progress_delta_ppm": result.ProgressDeltaPPM,
		"xp_delta":           xp,
	})
	kind := save.EventGuildTitheAccrued
	if xp == 0 {
		kind = save.EventGuildActivityEvaluated
	}
	return []save.EventWrite{{Kind: kind, SchemaVersion: 1, Payload: payload}}, nil
}

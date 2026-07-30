// Package guild owns the account-scoped guild lifecycle and deterministic
// faction-stock exchange. It does not import the production engine.
package guild

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

const CatalogSchemaVersion = 1

var ErrInvalidCatalog = errors.New("invalid guild catalog")

type Catalog struct {
	GuildTithePPM              int64
	ClearingRatePPM            int64
	NPCExchangePPM             int64
	StockIntakeCap             int64
	ConsumptionBonusPPMPerUnit int64
	MaxMembers                 int
	MinMembers                 int
	GraceDays                  int
}

type rawCatalog struct {
	SchemaVersion              int   `json:"schema_version"`
	GuildTithePPM              int64 `json:"guild_tithe_ppm"`
	ClearingRatePPM            int64 `json:"clearing_rate_ppm"`
	NPCExchangePPM             int64 `json:"npc_exchange_ppm"`
	StockIntakeCap             int64 `json:"stock_intake_cap"`
	ConsumptionBonusPPMPerUnit int64 `json:"consumption_bonus_ppm_per_unit"`
	MaxMembers                 int   `json:"max_members"`
	MinMembers                 int   `json:"min_members"`
	GraceDays                  int   `json:"grace_days"`
}

func LoadCatalog(data []byte) (*Catalog, error) {
	var raw rawCatalog
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&raw); err != nil {
		return nil, fmt.Errorf("%w: decode: %v", ErrInvalidCatalog, err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return nil, fmt.Errorf("%w: trailing JSON", ErrInvalidCatalog)
	}
	if raw.SchemaVersion != CatalogSchemaVersion || raw.GuildTithePPM != 20_000 ||
		raw.ClearingRatePPM != 500_000 || raw.NPCExchangePPM != 250_000 ||
		raw.StockIntakeCap != 120 || raw.ConsumptionBonusPPMPerUnit != 0 ||
		raw.MaxMembers != 50 || raw.MinMembers != 2 || raw.GraceDays != 7 ||
		raw.NPCExchangePPM >= raw.ClearingRatePPM || raw.MinMembers > raw.MaxMembers {
		return nil, ErrInvalidCatalog
	}
	return &Catalog{
		GuildTithePPM: raw.GuildTithePPM, ClearingRatePPM: raw.ClearingRatePPM,
		NPCExchangePPM: raw.NPCExchangePPM, StockIntakeCap: raw.StockIntakeCap,
		ConsumptionBonusPPMPerUnit: raw.ConsumptionBonusPPMPerUnit,
		MaxMembers:                 raw.MaxMembers, MinMembers: raw.MinMembers, GraceDays: raw.GraceDays,
	}, nil
}

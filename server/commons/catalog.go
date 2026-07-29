// Package commons implements the pure Mutual Aid Compact arithmetic and
// declarative policy. It deliberately does not import the production engine.
package commons

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"

	"cloud-clicker/server/decimal"
	"cloud-clicker/server/multiplier"
)

const (
	CatalogSchemaVersion = 1
	PPM                  = int64(1_000_000)
)

var (
	ErrInvalidCatalog = errors.New("invalid commons catalog")
	idPattern         = regexp.MustCompile(`^[a-z][a-z0-9_]*(?:\.[a-z][a-z0-9_]*)*$`)
)

type SourceWeight struct {
	SourceID  string
	Slot      multiplier.Slot
	WeightPPM int64
	Forsworn  bool
}

type Catalog struct {
	SourceWeights            []SourceWeight
	DefaultTithePPM          int64
	MinimumTithePPM          int64
	MaximumTithePPM          int64
	GuildHealthWeightPPM     int64
	CohortHealthWeightPPM    int64
	ServerHealthWeightPPM    int64
	CollectiveWeightPPM      int64
	CollapseHealthPPM        int64
	MaximumBonus             decimal.Decimal
	HealthRecoveryPPMPerHour int64
	HealthDecayPPMPerHour    int64
	SolidarityWindowMS       int64
	CohortTargetSize         int
	CohortMergeFloor         int
	NPCPopulationFloor       int
	NPCWeightPPM             int64
	NPCCompliancePPM         int64
	PopulationTolerancePPM   int64
	weightBySource           map[string]SourceWeight
}

type rawCatalog struct {
	SchemaVersion            int               `json:"schema_version"`
	SourceWeights            []rawSourceWeight `json:"source_weights"`
	DefaultTithePPM          int64             `json:"default_tithe_ppm"`
	MinimumTithePPM          int64             `json:"minimum_tithe_ppm"`
	MaximumTithePPM          int64             `json:"maximum_tithe_ppm"`
	GuildHealthWeightPPM     int64             `json:"guild_health_weight_ppm"`
	CohortHealthWeightPPM    int64             `json:"cohort_health_weight_ppm"`
	ServerHealthWeightPPM    int64             `json:"server_health_weight_ppm"`
	CollectiveWeightPPM      int64             `json:"collective_weight_ppm"`
	CollapseHealthPPM        int64             `json:"collapse_health_ppm"`
	MaximumBonus             string            `json:"maximum_bonus"`
	HealthRecoveryPPMPerHour int64             `json:"health_recovery_ppm_per_hour"`
	HealthDecayPPMPerHour    int64             `json:"health_decay_ppm_per_hour"`
	SolidarityWindowMS       int64             `json:"solidarity_window_ms"`
	CohortTargetSize         int               `json:"cohort_target_size"`
	CohortMergeFloor         int               `json:"cohort_merge_floor"`
	NPCPopulationFloor       int               `json:"npc_population_floor"`
	NPCWeightPPM             int64             `json:"npc_weight_ppm"`
	NPCCompliancePPM         int64             `json:"npc_compliance_ppm"`
	PopulationTolerancePPM   int64             `json:"population_tolerance_ppm"`
}

type rawSourceWeight struct {
	SourceID  string `json:"source_id"`
	Slot      string `json:"slot"`
	WeightPPM int64  `json:"weight_ppm"`
	Forsworn  bool   `json:"forsworn"`
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
	bonus, err := decimal.ParseCanonical(raw.MaximumBonus)
	if err != nil || !bonus.Gt(decimal.Zero) {
		return nil, fmt.Errorf("%w: maximum_bonus must be a positive canonical Decimal", ErrInvalidCatalog)
	}
	validRatio := func(value int64) bool { return value >= 0 && value <= PPM }
	if raw.SchemaVersion != CatalogSchemaVersion || raw.SourceWeights == nil ||
		!validRatio(raw.DefaultTithePPM) || !validRatio(raw.MinimumTithePPM) || !validRatio(raw.MaximumTithePPM) ||
		raw.MinimumTithePPM > raw.DefaultTithePPM || raw.DefaultTithePPM > raw.MaximumTithePPM ||
		!validRatio(raw.GuildHealthWeightPPM) || !validRatio(raw.CohortHealthWeightPPM) || !validRatio(raw.ServerHealthWeightPPM) ||
		raw.GuildHealthWeightPPM+raw.CohortHealthWeightPPM+raw.ServerHealthWeightPPM != PPM ||
		!validRatio(raw.CollectiveWeightPPM) || !validRatio(raw.CollapseHealthPPM) ||
		!validRatio(raw.HealthRecoveryPPMPerHour) || !validRatio(raw.HealthDecayPPMPerHour) ||
		raw.HealthRecoveryPPMPerHour <= raw.HealthDecayPPMPerHour || raw.SolidarityWindowMS <= 0 ||
		raw.CohortTargetSize <= 0 || raw.CohortMergeFloor <= 0 || raw.CohortMergeFloor > raw.CohortTargetSize ||
		raw.NPCPopulationFloor < raw.CohortMergeFloor || !validRatio(raw.NPCWeightPPM) || !validRatio(raw.NPCCompliancePPM) ||
		raw.PopulationTolerancePPM < 0 || raw.PopulationTolerancePPM > PPM {
		return nil, ErrInvalidCatalog
	}
	catalog := &Catalog{
		DefaultTithePPM: raw.DefaultTithePPM, MinimumTithePPM: raw.MinimumTithePPM, MaximumTithePPM: raw.MaximumTithePPM,
		GuildHealthWeightPPM: raw.GuildHealthWeightPPM, CohortHealthWeightPPM: raw.CohortHealthWeightPPM, ServerHealthWeightPPM: raw.ServerHealthWeightPPM,
		CollectiveWeightPPM: raw.CollectiveWeightPPM, CollapseHealthPPM: raw.CollapseHealthPPM, MaximumBonus: bonus,
		HealthRecoveryPPMPerHour: raw.HealthRecoveryPPMPerHour, HealthDecayPPMPerHour: raw.HealthDecayPPMPerHour,
		SolidarityWindowMS: raw.SolidarityWindowMS, CohortTargetSize: raw.CohortTargetSize, CohortMergeFloor: raw.CohortMergeFloor,
		NPCPopulationFloor: raw.NPCPopulationFloor, NPCWeightPPM: raw.NPCWeightPPM, NPCCompliancePPM: raw.NPCCompliancePPM,
		PopulationTolerancePPM: raw.PopulationTolerancePPM, weightBySource: map[string]SourceWeight{},
	}
	for index, source := range raw.SourceWeights {
		weight := SourceWeight{SourceID: source.SourceID, Slot: multiplier.Slot(source.Slot), WeightPPM: source.WeightPPM, Forsworn: source.Forsworn}
		if !idPattern.MatchString(weight.SourceID) || !multiplier.ValidSlot(weight.Slot) || weight.Slot == multiplier.SlotCommons || weight.WeightPPM <= 0 || weight.WeightPPM > PPM {
			return nil, fmt.Errorf("%w: source_weights[%d]", ErrInvalidCatalog, index)
		}
		if _, duplicate := catalog.weightBySource[weight.SourceID]; duplicate {
			return nil, fmt.Errorf("%w: duplicate source %q", ErrInvalidCatalog, weight.SourceID)
		}
		catalog.SourceWeights = append(catalog.SourceWeights, weight)
		catalog.weightBySource[weight.SourceID] = weight
	}
	return catalog, nil
}

func (catalog *Catalog) SourceWeight(id string) (SourceWeight, bool) {
	if catalog == nil {
		return SourceWeight{}, false
	}
	value, ok := catalog.weightBySource[id]
	return value, ok
}

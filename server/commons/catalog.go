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
	SourceID  string          `json:"source_id"`
	Slot      multiplier.Slot `json:"slot"`
	WeightPPM int64           `json:"weight_ppm"`
	Forsworn  bool            `json:"forsworn"`
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
	CollectiveExponentPPM    int64
	CollapseHealthPPM        int64
	HealthyHealthPPM         int64
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
	CollectiveExponentPPM    int64             `json:"collective_exponent_ppm"`
	CollapseHealthPPM        int64             `json:"collapse_health_ppm"`
	HealthyHealthPPM         int64             `json:"healthy_health_ppm"`
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
	if !commonsCatalogHasExactShape(data) {
		return nil, ErrInvalidCatalog
	}
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
		!validRatio(raw.CollectiveWeightPPM) || raw.CollectiveExponentPPM < PPM || raw.CollectiveExponentPPM > 10*PPM || !validRatio(raw.CollapseHealthPPM) ||
		!validRatio(raw.HealthyHealthPPM) || raw.HealthyHealthPPM <= raw.CollapseHealthPPM ||
		!validRatio(raw.HealthRecoveryPPMPerHour) || !validRatio(raw.HealthDecayPPMPerHour) ||
		raw.HealthRecoveryPPMPerHour <= raw.HealthDecayPPMPerHour || raw.SolidarityWindowMS <= 0 || raw.SolidarityWindowMS > decimal.MaxExactInteger ||
		raw.CohortTargetSize <= 0 || int64(raw.CohortTargetSize) > decimal.MaxExactInteger || raw.CohortMergeFloor <= 0 ||
		int64(raw.CohortMergeFloor) > decimal.MaxExactInteger || raw.CohortMergeFloor > raw.CohortTargetSize ||
		raw.NPCPopulationFloor < raw.CohortMergeFloor || int64(raw.NPCPopulationFloor) > decimal.MaxExactInteger || !validRatio(raw.NPCWeightPPM) || !validRatio(raw.NPCCompliancePPM) ||
		raw.PopulationTolerancePPM < 0 || raw.PopulationTolerancePPM > PPM {
		return nil, ErrInvalidCatalog
	}
	catalog := &Catalog{
		SourceWeights:   make([]SourceWeight, 0, len(raw.SourceWeights)),
		DefaultTithePPM: raw.DefaultTithePPM, MinimumTithePPM: raw.MinimumTithePPM, MaximumTithePPM: raw.MaximumTithePPM,
		GuildHealthWeightPPM: raw.GuildHealthWeightPPM, CohortHealthWeightPPM: raw.CohortHealthWeightPPM, ServerHealthWeightPPM: raw.ServerHealthWeightPPM,
		CollectiveWeightPPM: raw.CollectiveWeightPPM, CollectiveExponentPPM: raw.CollectiveExponentPPM, CollapseHealthPPM: raw.CollapseHealthPPM, MaximumBonus: bonus,
		HealthyHealthPPM:         raw.HealthyHealthPPM,
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

func commonsCatalogHasExactShape(data []byte) bool {
	var root map[string]json.RawMessage
	if json.Unmarshal(data, &root) != nil || !exactCommonsKeys(root,
		"schema_version", "source_weights", "default_tithe_ppm", "minimum_tithe_ppm", "maximum_tithe_ppm",
		"guild_health_weight_ppm", "cohort_health_weight_ppm", "server_health_weight_ppm", "collective_weight_ppm",
		"collective_exponent_ppm", "collapse_health_ppm", "healthy_health_ppm", "maximum_bonus", "health_recovery_ppm_per_hour",
		"health_decay_ppm_per_hour", "solidarity_window_ms", "cohort_target_size", "cohort_merge_floor", "npc_population_floor",
		"npc_weight_ppm", "npc_compliance_ppm", "population_tolerance_ppm") {
		return false
	}
	var weights []map[string]json.RawMessage
	if json.Unmarshal(root["source_weights"], &weights) != nil || weights == nil {
		return false
	}
	for _, weight := range weights {
		if !exactCommonsKeys(weight, "source_id", "slot", "weight_ppm", "forsworn") {
			return false
		}
	}
	return true
}

func exactCommonsKeys(fields map[string]json.RawMessage, expected ...string) bool {
	if len(fields) != len(expected) {
		return false
	}
	for _, key := range expected {
		value, exists := fields[key]
		if !exists || bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
			return false
		}
	}
	return true
}

func (catalog *Catalog) SourceWeight(id string) (SourceWeight, bool) {
	if catalog == nil {
		return SourceWeight{}, false
	}
	value, ok := catalog.weightBySource[id]
	return value, ok
}

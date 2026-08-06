// Package activeplay owns the optional active-play catalog and deterministic
// opportunity scheduler. It has no store or production dependency.
package activeplay

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"regexp"
	"sort"

	"cloud-clicker/server/decimal"
	"cloud-clicker/server/determinism"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/multiplier"
	"cloud-clicker/server/runidentity"
)

const (
	CatalogSchemaVersion = 1
	SamplerVersion       = "gamma6_exp.v1"
	SpawnSubstream       = "active_play.spawn.v1"
	OpportunityIDStream  = "active_play.opportunity_id.v1"
	BuffIDStream         = "active_play.buff_id.v1"
)

var (
	ErrInvalidCatalog = errors.New("invalid active-play catalog")
	ErrInvalidDraw    = errors.New("invalid active-play draw")
	idPattern         = regexp.MustCompile(`^[a-z][a-z0-9_]*(?:\.[a-z][a-z0-9_]*)*$`)
)

type SchedulePolicy struct {
	SamplerVersion    string
	SubstreamLabel    string
	MinimumIntervalMS int64
	ScaleMS           int64
	LifetimeMS        int64
	MaxDueTransitions int64
}

type Effect struct {
	ID                   string
	Kind                 string
	Weight               int64
	Factor               decimal.Decimal
	DurationMS           int64
	Targets              []string
	ActionIDs            []string
	PerOwnedPPM          int64
	EligibleGeneratorIDs []string
	LuckyBankFrac        decimal.Decimal
	LuckyRateCap         decimal.Decimal
	Epsilon              decimal.Decimal
	ResourceID           string
	HardcapReasonKey     string
}

type ComboPolicy struct {
	Cap              decimal.Decimal
	HardcapReasonKey string
}

type Catalog struct {
	Schedule SchedulePolicy
	Combo    ComboPolicy
	effects  []Effect
	byID     map[string]Effect
	weight   uint64
}

type rawCatalog struct {
	SchemaVersion  *int              `json:"schema_version"`
	SchedulePolicy rawSchedulePolicy `json:"schedule_policy"`
	Effects        []rawEffect       `json:"effects"`
	ComboPolicy    rawComboPolicy    `json:"combo_policy"`
}
type rawSchedulePolicy struct {
	SamplerVersion    string `json:"sampler_version"`
	SubstreamLabel    string `json:"substream_label"`
	MinimumIntervalMS *int64 `json:"minimum_interval_ms"`
	ScaleMS           *int64 `json:"scale_ms"`
	LifetimeMS        *int64 `json:"lifetime_ms"`
	MaxDueTransitions *int64 `json:"max_due_transitions"`
}
type rawEffect struct {
	EffectRowID          string   `json:"effect_row_id"`
	Kind                 string   `json:"kind"`
	Weight               *int64   `json:"weight"`
	Factor               *string  `json:"factor,omitempty"`
	DurationMS           *int64   `json:"duration_ms,omitempty"`
	Targets              []string `json:"targets,omitempty"`
	ActionIDs            []string `json:"action_ids,omitempty"`
	PerOwnedPPM          *int64   `json:"per_owned_ppm,omitempty"`
	EligibleGeneratorIDs []string `json:"eligible_generator_ids,omitempty"`
	LuckyBankFrac        *string  `json:"lucky_bank_frac,omitempty"`
	LuckyRateCap         *string  `json:"lucky_rate_cap,omitempty"`
	Epsilon              *string  `json:"epsilon,omitempty"`
	ResourceID           string   `json:"resource_id,omitempty"`
	HardcapReasonKey     string   `json:"hardcap_reason_key,omitempty"`
}
type rawComboPolicy struct {
	ComboCap         *string `json:"combo_cap"`
	HardcapReasonKey string  `json:"hardcap_reason_key"`
}

func LoadCatalog(data []byte, economyCatalog *economy.Catalog) (*Catalog, error) {
	if economyCatalog == nil {
		return nil, ErrInvalidCatalog
	}
	var raw rawCatalog
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&raw); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidCatalog, err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF || raw.SchemaVersion == nil || *raw.SchemaVersion != CatalogSchemaVersion ||
		raw.SchedulePolicy.MinimumIntervalMS == nil || raw.SchedulePolicy.ScaleMS == nil || raw.SchedulePolicy.LifetimeMS == nil ||
		raw.SchedulePolicy.MaxDueTransitions == nil || raw.ComboPolicy.ComboCap == nil || len(raw.Effects) == 0 {
		return nil, ErrInvalidCatalog
	}
	schedule := SchedulePolicy{raw.SchedulePolicy.SamplerVersion, raw.SchedulePolicy.SubstreamLabel, *raw.SchedulePolicy.MinimumIntervalMS,
		*raw.SchedulePolicy.ScaleMS, *raw.SchedulePolicy.LifetimeMS, *raw.SchedulePolicy.MaxDueTransitions}
	combo, comboErr := decimal.ParseCanonical(*raw.ComboPolicy.ComboCap)
	if schedule.SamplerVersion != SamplerVersion || schedule.SubstreamLabel != SpawnSubstream || schedule.MinimumIntervalMS <= 0 ||
		schedule.ScaleMS <= 0 || schedule.LifetimeMS <= 0 || schedule.MaxDueTransitions <= 0 ||
		schedule.MinimumIntervalMS > decimal.MaxExactInteger || schedule.ScaleMS > decimal.MaxExactInteger || schedule.LifetimeMS > decimal.MaxExactInteger ||
		schedule.MaxDueTransitions > decimal.MaxExactInteger || comboErr != nil || !combo.Gt(decimal.One) ||
		!idPattern.MatchString(raw.ComboPolicy.HardcapReasonKey) {
		return nil, ErrInvalidCatalog
	}
	result := &Catalog{Schedule: schedule, Combo: ComboPolicy{combo, raw.ComboPolicy.HardcapReasonKey}, byID: map[string]Effect{}}
	previous := ""
	for index, source := range raw.Effects {
		if source.Weight == nil || *source.Weight <= 0 || *source.Weight > decimal.MaxExactInteger || !idPattern.MatchString(source.EffectRowID) ||
			previous != "" && previous >= source.EffectRowID {
			return nil, fmt.Errorf("%w: effects[%d]", ErrInvalidCatalog, index)
		}
		previous = source.EffectRowID
		effect := Effect{ID: source.EffectRowID, Kind: source.Kind, Weight: *source.Weight}
		if err := parseEffect(&effect, source, economyCatalog); err != nil {
			return nil, fmt.Errorf("%w: effects[%d]: %v", ErrInvalidCatalog, index, err)
		}
		if result.weight > ^uint64(0)-uint64(effect.Weight) {
			return nil, ErrInvalidCatalog
		}
		result.weight += uint64(effect.Weight)
		result.effects = append(result.effects, effect)
		result.byID[effect.ID] = effect
	}
	return result, nil
}

func parseEffect(out *Effect, raw rawEffect, catalog *economy.Catalog) error {
	switch raw.Kind {
	case "production_frenzy":
		if raw.Factor == nil || raw.DurationMS == nil || !only(raw, "factor", "duration", "targets") {
			return ErrInvalidCatalog
		}
		factor, err := decimal.ParseCanonical(*raw.Factor)
		if err != nil || !factor.Gt(decimal.One) || *raw.DurationMS <= 0 || !sortedUnique(raw.Targets) || len(raw.Targets) == 0 {
			return ErrInvalidCatalog
		}
		for _, target := range raw.Targets {
			if target != "generator_production" {
				return ErrInvalidCatalog
			}
		}
		if !validDeclaration(catalog, out.ID, "all") {
			return ErrInvalidCatalog
		}
		out.Factor, out.DurationMS, out.Targets = factor, *raw.DurationMS, append([]string(nil), raw.Targets...)
	case "click_frenzy":
		if raw.Factor == nil || raw.DurationMS == nil || !only(raw, "factor", "duration", "actions") {
			return ErrInvalidCatalog
		}
		factor, err := decimal.ParseCanonical(*raw.Factor)
		if err != nil || !factor.Gt(decimal.One) || *raw.DurationMS <= 0 || !sortedUnique(raw.ActionIDs) || len(raw.ActionIDs) == 0 {
			return ErrInvalidCatalog
		}
		for _, action := range raw.ActionIDs {
			if _, ok := catalog.ManualAction(action); !ok || !validDeclaration(catalog, out.ID, action) {
				return ErrInvalidCatalog
			}
		}
		out.Factor, out.DurationMS, out.ActionIDs = factor, *raw.DurationMS, append([]string(nil), raw.ActionIDs...)
	case "building_special":
		if raw.PerOwnedPPM == nil || raw.DurationMS == nil || !only(raw, "ppm", "duration", "generators") || *raw.PerOwnedPPM <= 0 ||
			*raw.PerOwnedPPM > 1_000_000 || *raw.DurationMS <= 0 || !sortedUnique(raw.EligibleGeneratorIDs) || len(raw.EligibleGeneratorIDs) == 0 {
			return ErrInvalidCatalog
		}
		for _, generator := range raw.EligibleGeneratorIDs {
			if _, ok := catalog.GeneratorClass(generator); !ok || !validDeclaration(catalog, out.ID+"."+generator, generator) {
				return ErrInvalidCatalog
			}
		}
		out.PerOwnedPPM, out.DurationMS, out.EligibleGeneratorIDs = *raw.PerOwnedPPM, *raw.DurationMS, append([]string(nil), raw.EligibleGeneratorIDs...)
	case "lucky_payout":
		if raw.LuckyBankFrac == nil || raw.LuckyRateCap == nil || raw.Epsilon == nil || !only(raw, "lucky") {
			return ErrInvalidCatalog
		}
		frac, e1 := decimal.ParseCanonical(*raw.LuckyBankFrac)
		cap, e2 := decimal.ParseCanonical(*raw.LuckyRateCap)
		epsilon, e3 := decimal.ParseCanonical(*raw.Epsilon)
		resource, ok := catalog.Resource(raw.ResourceID)
		if e1 != nil || e2 != nil || e3 != nil || !frac.Gt(decimal.Zero) || !frac.Lt(decimal.One) || !cap.Gt(decimal.Zero) ||
			!epsilon.IsStateValue() || epsilon.Lt(decimal.Zero) || !ok || resource.Scope != economy.ScopeCompany || !idPattern.MatchString(raw.HardcapReasonKey) {
			return ErrInvalidCatalog
		}
		out.LuckyBankFrac, out.LuckyRateCap, out.Epsilon, out.ResourceID, out.HardcapReasonKey = frac, cap, epsilon, raw.ResourceID, raw.HardcapReasonKey
	default:
		return ErrInvalidCatalog
	}
	return nil
}

func only(raw rawEffect, mode ...string) bool {
	allowed := map[string]bool{}
	for _, value := range mode {
		allowed[value] = true
	}
	return (raw.Factor == nil) == !allowed["factor"] && (raw.DurationMS == nil) == !allowed["duration"] &&
		(raw.Targets == nil) == !allowed["targets"] && (raw.ActionIDs == nil) == !allowed["actions"] &&
		(raw.PerOwnedPPM == nil) == !allowed["ppm"] && (raw.EligibleGeneratorIDs == nil) == !allowed["generators"] &&
		(raw.LuckyBankFrac == nil && raw.LuckyRateCap == nil && raw.Epsilon == nil && raw.ResourceID == "" && raw.HardcapReasonKey == "") == !allowed["lucky"]
}

func validDeclaration(catalog *economy.Catalog, sourceID, target string) bool {
	value, ok := catalog.MultiplierSource(sourceID)
	return ok && value.Slot == economy.SlotEventBuffs && value.Provider == "active_play" && value.Target == target
}

func sortedUnique(values []string) bool {
	if values == nil {
		return false
	}
	for index, value := range values {
		if !idPattern.MatchString(value) || index > 0 && values[index-1] >= value {
			return false
		}
	}
	return true
}

func (catalog *Catalog) Effects() []Effect { return append([]Effect(nil), catalog.effects...) }
func (catalog *Catalog) Effect(id string) (Effect, bool) {
	value, ok := catalog.byID[id]
	return value, ok
}

type Spawn struct {
	Sequence          int64
	SampledIntervalMS int64
	EffectDraw        uint64
	GeneratorDraw     *uint64
	EffectRowID       string
	SelectedGenerator string
	OpportunityID     string
	SpawnedAttendedMS int64
	ExpiresAttendedMS int64
}

func (catalog *Catalog) Spawn(founderID string, runSeq, sequence, fromAttendedMS int64) (Spawn, error) {
	if catalog == nil || founderID == "" || runSeq < 1 || sequence < 0 || fromAttendedMS < 0 {
		return Spawn{}, ErrInvalidDraw
	}
	base := runidentity.Seed(founderID, runSeq) ^ uint64(sequence)
	random := determinism.Substream(base, SpawnSubstream)
	interval, err := sampleGamma6MS(random, catalog.Schedule.MinimumIntervalMS, catalog.Schedule.ScaleMS)
	if err != nil || interval > decimal.MaxExactInteger-fromAttendedMS {
		return Spawn{}, ErrInvalidDraw
	}
	draw := random.Bound(catalog.weight)
	remaining := draw
	var effect Effect
	for _, candidate := range catalog.effects {
		if remaining < uint64(candidate.Weight) {
			effect = candidate
			break
		}
		remaining -= uint64(candidate.Weight)
	}
	if effect.ID == "" {
		return Spawn{}, ErrInvalidDraw
	}
	var generatorDraw *uint64
	selected := ""
	if effect.Kind == "building_special" {
		value := random.Bound(uint64(len(effect.EligibleGeneratorIDs)))
		generatorDraw = &value
		selected = effect.EligibleGeneratorIDs[value]
	}
	spawned := fromAttendedMS + interval
	if catalog.Schedule.LifetimeMS > decimal.MaxExactInteger-spawned {
		return Spawn{}, ErrInvalidDraw
	}
	return Spawn{Sequence: sequence, SampledIntervalMS: interval, EffectDraw: draw, GeneratorDraw: generatorDraw,
		EffectRowID: effect.ID, SelectedGenerator: selected, OpportunityID: deterministicID(base, OpportunityIDStream, spawned),
		SpawnedAttendedMS: spawned, ExpiresAttendedMS: spawned + catalog.Schedule.LifetimeMS}, nil
}

func sampleGamma6MS(random *determinism.SplitMix64, minimum, scale int64) (int64, error) {
	if random == nil || minimum <= 0 || scale <= 0 {
		return 0, ErrInvalidDraw
	}
	sum := 0.0
	for index := 0; index < 6; index++ {
		// The top 53 bits form a value in (0,1]; zero is deliberately shifted
		// away from log(0). Only the resulting integer interval crosses runtimes.
		u := float64((random.Next()>>11)+1) / float64(uint64(1)<<53)
		sum += -math.Log(u)
	}
	value := float64(minimum) + math.Ceil(sum*float64(scale))
	if math.IsInf(value, 0) || math.IsNaN(value) || value < 1 || value > float64(decimal.MaxExactInteger) {
		return 0, ErrInvalidDraw
	}
	return int64(value), nil
}

func deterministicID(seed uint64, label string, coordinate int64) string {
	random := determinism.Substream(seed^uint64(coordinate), label)
	var raw [16]byte
	binary.BigEndian.PutUint64(raw[:8], random.Next())
	binary.BigEndian.PutUint64(raw[8:], random.Next())
	// UUIDv7-compatible layout; the attended coordinate supplies its timestamp.
	timestamp := uint64(coordinate) & ((uint64(1) << 48) - 1)
	for index := 5; index >= 0; index-- {
		raw[index] = byte(timestamp)
		timestamp >>= 8
	}
	raw[6] = raw[6]&0x0f | 0x70
	raw[8] = raw[8]&0x3f | 0x80
	encoded := hex.EncodeToString(raw[:])
	return encoded[:8] + "-" + encoded[8:12] + "-" + encoded[12:16] + "-" + encoded[16:20] + "-" + encoded[20:]
}

func (catalog *Catalog) BuffID(founderID string, runSeq, sequence, attendedMS int64) string {
	return deterministicID(runidentity.Seed(founderID, runSeq)^uint64(sequence), BuffIDStream, attendedMS)
}

func SortedEffects(values []Effect) {
	sort.Slice(values, func(i, j int) bool { return values[i].ID < values[j].ID })
}

var _ multiplier.Slot = multiplier.SlotEventBuffs

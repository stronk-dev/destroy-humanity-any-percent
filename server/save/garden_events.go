package save

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"

	"cloud-clicker/server/decimal"
	"cloud-clicker/server/garden"
)

// Server Garden SG8 event kinds (schema v1). The Founder log carries the
// first five; `garden_harvest_credited.v1` is the Company run-log side.
const (
	EventGardenAdvanced        EventKind = "garden_advanced.v1"
	EventGardenPlanted         EventKind = "garden_planted.v1"
	EventGardenUprooted        EventKind = "garden_uprooted.v1"
	EventGardenSubstrateSet    EventKind = "garden_substrate_set.v1"
	EventGardenHarvested       EventKind = "garden_harvested.v1"
	EventGardenHarvestCredited EventKind = "garden_harvest_credited.v1"
)

var (
	gardenIDPattern   = regexp.MustCompile(`^[a-z][a-z0-9_]{0,47}$`)
	gardenHashPattern = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)
)

func isGardenEvent(kind EventKind) bool {
	switch kind {
	case EventGardenAdvanced, EventGardenPlanted, EventGardenUprooted, EventGardenSubstrateSet, EventGardenHarvested, EventGardenHarvestCredited:
		return true
	}
	return false
}

func validGardenCell(row, col int64) bool {
	return row >= 0 && row < garden.MaxGridSide && col >= 0 && col < garden.MaxGridSide
}

// validateGardenEventPayload is SG8's strict payload grammar. It never trusts
// a payload: every key must be present (nullable keys included) and exact.
func validateGardenEventPayload(event EventWrite) error {
	invalid := fmt.Errorf("%w: invalid %s payload", ErrInvalidStream, event.Kind)
	fields := gardenEventFields(event.Payload, gardenEventKeys[event.Kind]...)
	if event.SchemaVersion != 1 || fields == nil {
		return invalid
	}
	switch event.Kind {
	case EventGardenAdvanced:
		if !gardenEventItems(fields["matured"], "row", "col", "species_id") ||
			!gardenEventItems(fields["spawned"], "row", "col", "species_id", "recipe_id") {
			return invalid
		}
		var payload garden.Advance
		if decodeStrictJSON(event.Payload, &payload) != nil || payload.TicksApplied < 0 || payload.TickSeqAfter < payload.TicksApplied ||
			payload.TickSeqAfter > decimal.MaxExactInteger || payload.CatchupForfeitedMS < 0 || payload.CatchupForfeitedMS > decimal.MaxExactInteger ||
			(payload.CatchupForfeitedMS > 0) != (payload.CatchupReasonKey != nil) || payload.Matured == nil || payload.Spawned == nil || !payload.Visible() {
			return invalid
		}
		if payload.CatchupReasonKey != nil && !mechanicalIDPattern.MatchString(*payload.CatchupReasonKey) {
			return invalid
		}
		for _, plot := range payload.Matured {
			if !validGardenCell(plot.Row, plot.Col) || !gardenIDPattern.MatchString(plot.SpeciesID) {
				return invalid
			}
		}
		for _, spawn := range payload.Spawned {
			if !validGardenCell(spawn.Row, spawn.Col) || !gardenIDPattern.MatchString(spawn.SpeciesID) || !gardenIDPattern.MatchString(spawn.RecipeID) {
				return invalid
			}
		}
	case EventGardenPlanted:
		var payload garden.Planted
		if decodeStrictJSON(event.Payload, &payload) != nil || !validGardenCell(payload.Row, payload.Col) || !gardenIDPattern.MatchString(payload.SpeciesID) {
			return invalid
		}
	case EventGardenUprooted:
		var payload garden.Uprooted
		if decodeStrictJSON(event.Payload, &payload) != nil || !validGardenCell(payload.Row, payload.Col) || !gardenIDPattern.MatchString(payload.SpeciesID) {
			return invalid
		}
	case EventGardenSubstrateSet:
		var payload garden.SubstrateSet
		if decodeStrictJSON(event.Payload, &payload) != nil || !gardenIDPattern.MatchString(payload.FromSubstrateID) || !gardenIDPattern.MatchString(payload.ToSubstrateID) ||
			payload.FromSubstrateID == payload.ToSubstrateID || payload.PartialTickForfeitedMS < 0 || payload.PartialTickForfeitedMS > garden.MaxTickMS {
			return invalid
		}
	case EventGardenHarvested:
		if !gardenEventItems(fields["plots"], "row", "col", "species_id", "units") {
			return invalid
		}
		var payload garden.Harvest
		if decodeStrictJSON(event.Payload, &payload) != nil || len(payload.Plots) < 1 || len(payload.Plots) > int(garden.MaxGridSide*garden.MaxGridSide) ||
			payload.SeedsDiscovered == nil || !gardenHashPattern.MatchString(payload.HarvestHash) || payload.TotalUnits < 0 || payload.TotalUnits > decimal.MaxExactInteger {
			return invalid
		}
		sum := int64(0)
		for _, plot := range payload.Plots {
			if !validGardenCell(plot.Row, plot.Col) || !gardenIDPattern.MatchString(plot.SpeciesID) || plot.Units < 0 {
				return invalid
			}
			sum += plot.Units
		}
		if sum != payload.TotalUnits {
			return invalid
		}
	case EventGardenHarvestCredited:
		var payload struct {
			HarvestHash    string  `json:"harvest_hash"`
			SelectedUnits  int64   `json:"selected_units"`
			Credited       string  `json:"credited"`
			ForfeitedUnits int64   `json:"forfeited_units"`
			CapReasonKey   *string `json:"cap_reason_key"`
			FaucetApplied  bool    `json:"faucet_applied"`
		}
		if decodeStrictJSON(event.Payload, &payload) != nil || !gardenHashPattern.MatchString(payload.HarvestHash) || payload.SelectedUnits < 0 ||
			payload.ForfeitedUnits < 0 || payload.ForfeitedUnits > payload.SelectedUnits || payload.Credited == "" ||
			payload.CapReasonKey != nil && !mechanicalIDPattern.MatchString(*payload.CapReasonKey) {
			return invalid
		}
		if _, err := decimal.ParseCanonical(payload.Credited); err != nil {
			return invalid
		}
	default:
		return invalid
	}
	return nil
}

var gardenEventKeys = map[EventKind][]string{
	EventGardenAdvanced:        {"ticks_applied", "tick_seq_after", "catchup_forfeited_ms", "catchup_reason_key", "matured", "spawned"},
	EventGardenPlanted:         {"row", "col", "species_id"},
	EventGardenUprooted:        {"row", "col", "species_id", "was_mature"},
	EventGardenSubstrateSet:    {"from_substrate_id", "to_substrate_id", "partial_tick_forfeited_ms"},
	EventGardenHarvested:       {"plots", "total_units", "seeds_discovered", "harvest_hash"},
	EventGardenHarvestCredited: {"harvest_hash", "selected_units", "credited", "forfeited_units", "cap_reason_key", "faucet_applied"},
}

// Check the original tokens before struct/map decoding loses duplicate names,
// accepts case aliases, or turns a missing/null scalar into its zero value.
// Only SG8's two explicitly nullable reason fields admit null. Keep this scoped
// to Garden; exactJSONKeys is also used by the saved Cosmetic state codec.
func gardenEventFields(data []byte, keys ...string) map[string]json.RawMessage {
	decoder := json.NewDecoder(bytes.NewReader(data))
	first, err := decoder.Token()
	if err != nil || first != json.Delim('{') || len(keys) == 0 {
		return nil
	}
	allowed := make(map[string]bool, len(keys))
	for _, key := range keys {
		allowed[key] = true
	}
	fields := make(map[string]json.RawMessage, len(keys))
	for decoder.More() {
		token, err := decoder.Token()
		key, ok := token.(string)
		if err != nil || !ok || !allowed[key] || fields[key] != nil {
			return nil
		}
		var raw json.RawMessage
		if decoder.Decode(&raw) != nil || bytes.Equal(bytes.TrimSpace(raw), []byte("null")) && key != "catchup_reason_key" && key != "cap_reason_key" {
			return nil
		}
		fields[key] = raw
	}
	last, err := decoder.Token()
	if err != nil || last != json.Delim('}') || len(fields) != len(keys) || ensureJSONEnd(decoder) != nil {
		return nil
	}
	return fields
}

func gardenEventItems(data []byte, keys ...string) bool {
	var items []json.RawMessage
	if json.Unmarshal(data, &items) != nil || items == nil {
		return false
	}
	for _, item := range items {
		if gardenEventFields(item, keys...) == nil {
			return false
		}
	}
	return true
}

func exactJSONKeys(data []byte, keys ...string) bool {
	var fields map[string]json.RawMessage
	if json.Unmarshal(data, &fields) != nil || len(fields) != len(keys) || len(keys) == 0 {
		return false
	}
	for _, key := range keys {
		if _, ok := fields[key]; !ok {
			return false
		}
	}
	return true
}

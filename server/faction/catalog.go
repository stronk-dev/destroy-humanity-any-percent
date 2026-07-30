// Package faction owns the run-scoped faction catalog and stock arithmetic.
// It does not import the production engine or economy ledger.
package faction

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
	"sort"

	"cloud-clicker/server/commons"
)

const CatalogSchemaVersion = 1

var (
	ErrInvalidCatalog = errors.New("invalid faction catalog")
	idPattern         = regexp.MustCompile(`^[a-z][a-z0-9_]*(?:\.[a-z][a-z0-9_]*)*$`)
)

var phase0FactionIDs = []string{"bootstrapper", "enterprise", "open_source", "vc_funded"}
var phase0Resources = []string{"compliance", "hype", "libraries", "revenue"}

type CompactBinding struct {
	AutoSign bool
	TithePPM int64
}

type Faction struct {
	ID                   string
	Produces             string
	Consumes             string
	Compact              *CompactBinding
	IncorporationCopyKey string
}

type Catalog struct {
	StockCap        int64
	StockIntervalMS int64
	Factions        []Faction
	byID            map[string]Faction
}

type rawCatalog struct {
	SchemaVersion   int          `json:"schema_version"`
	StockCap        int64        `json:"stock_cap"`
	StockIntervalMS int64        `json:"stock_interval_ms"`
	Factions        []rawFaction `json:"factions"`
}

type rawFaction struct {
	ID                   string        `json:"id"`
	Produces             string        `json:"produces"`
	Consumes             string        `json:"consumes"`
	Compact              *rawCompact   `json:"compact"`
	ModifierSlots        []rawModifier `json:"modifier_slots"`
	IncorporationCopyKey string        `json:"incorporation_copy_key"`
}

type rawCompact struct {
	AutoSign bool  `json:"auto_sign"`
	TithePPM int64 `json:"tithe_ppm"`
}

type rawModifier struct {
	Slot string `json:"slot"`
	PPM  int64  `json:"ppm"`
}

func LoadCatalog(data []byte, commonsCatalog *commons.Catalog) (*Catalog, error) {
	if commonsCatalog == nil {
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
	if raw.SchemaVersion != CatalogSchemaVersion || raw.StockCap < 1 || raw.StockCap > 9_007_199_254_740_991 ||
		raw.StockIntervalMS < 1 || raw.StockIntervalMS > 86_400_000 || len(raw.Factions) != len(phase0FactionIDs) {
		return nil, ErrInvalidCatalog
	}
	catalog := &Catalog{StockCap: raw.StockCap, StockIntervalMS: raw.StockIntervalMS, Factions: make([]Faction, 0, len(raw.Factions)), byID: map[string]Faction{}}
	produced, consumed := map[string]bool{}, map[string]bool{}
	for index, source := range raw.Factions {
		if !idPattern.MatchString(source.ID) || !idPattern.MatchString(source.Produces) || !idPattern.MatchString(source.Consumes) ||
			source.Produces == source.Consumes || source.ModifierSlots == nil || len(source.ModifierSlots) != 0 ||
			source.IncorporationCopyKey != "incorporate."+source.ID || catalog.byID[source.ID].ID != "" || produced[source.Produces] || consumed[source.Consumes] {
			return nil, fmt.Errorf("%w: factions[%d]", ErrInvalidCatalog, index)
		}
		faction := Faction{ID: source.ID, Produces: source.Produces, Consumes: source.Consumes, IncorporationCopyKey: source.IncorporationCopyKey}
		if source.Compact != nil {
			if source.ID != "open_source" || !source.Compact.AutoSign || source.Compact.TithePPM <= commonsCatalog.DefaultTithePPM ||
				source.Compact.TithePPM < commonsCatalog.MinimumTithePPM || source.Compact.TithePPM > commonsCatalog.MaximumTithePPM {
				return nil, fmt.Errorf("%w: factions[%d].compact", ErrInvalidCatalog, index)
			}
			faction.Compact = &CompactBinding{AutoSign: true, TithePPM: source.Compact.TithePPM}
		} else if source.ID == "open_source" {
			return nil, fmt.Errorf("%w: open_source compact", ErrInvalidCatalog)
		}
		catalog.byID[faction.ID] = faction
		produced[faction.Produces], consumed[faction.Consumes] = true, true
		catalog.Factions = append(catalog.Factions, faction)
	}
	if !sameSortedKeys(catalog.byID, phase0FactionIDs) || !sameSortedSet(produced, phase0Resources) || !sameSortedSet(consumed, phase0Resources) || !singleCycle(catalog.Factions) {
		return nil, ErrInvalidCatalog
	}
	sort.Slice(catalog.Factions, func(left, right int) bool { return catalog.Factions[left].ID < catalog.Factions[right].ID })
	return catalog, nil
}

func (catalog *Catalog) Faction(id string) (Faction, bool) {
	if catalog == nil {
		return Faction{}, false
	}
	value, ok := catalog.byID[id]
	return value, ok
}

func sameSortedKeys(values map[string]Faction, expected []string) bool {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return equalStrings(keys, expected)
}

func sameSortedSet(values map[string]bool, expected []string) bool {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return equalStrings(keys, expected)
}

func equalStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func singleCycle(factions []Faction) bool {
	consumerByResource := map[string]string{}
	for _, faction := range factions {
		consumerByResource[faction.Consumes] = faction.ID
	}
	if len(factions) == 0 {
		return false
	}
	visited := map[string]bool{}
	current := factions[0].ID
	byID := map[string]Faction{}
	for _, faction := range factions {
		byID[faction.ID] = faction
	}
	for range factions {
		if visited[current] {
			return false
		}
		visited[current] = true
		current = consumerByResource[byID[current].Produces]
	}
	return current == factions[0].ID && len(visited) == len(factions)
}

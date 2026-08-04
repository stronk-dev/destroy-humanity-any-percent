// Package meters owns the run-scoped Company pressure-meter catalog and exact integer rules.
// It deliberately does not import the economy ledger: meters are consequences, never payment.
package meters

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
	"sort"

	"cloud-clicker/server/multiplier"
)

const (
	CatalogSchemaVersion = 1
	MinimumValue         = 0
	MaximumValue         = 100
)

var (
	ErrInvalidCatalog = errors.New("invalid meter catalog")
	idPattern         = regexp.MustCompile(`^[a-z][a-z0-9_]*(?:\.[a-z][a-z0-9_]*)*$`)
	requiredMeterIDs  = []string{
		"doom.probability",
		"trust.employees.grievance", "trust.employees.standing",
		"trust.investors.grievance", "trust.investors.standing",
		"trust.press.grievance", "trust.press.standing",
		"trust.regulators.grievance", "trust.regulators.standing",
		"trust.users.grievance", "trust.users.standing",
	}
)

type InputKind string

const (
	InputLedgerFact       InputKind = "ledger_fact"
	InputContributionSlot InputKind = "contribution_slot"
)

type Band struct {
	ID         string
	FloorValue int
}

type Input struct {
	Kind                 InputKind
	FactKind             string
	Slot                 multiplier.Slot
	SourceID             string
	Delta                int
	DeltaPerAttendedHour int
}

type Decay struct {
	TowardValue int
	RatePerHour int
}

type Meter struct {
	ID           string
	InitialValue int
	Bands        []Band
	Inputs       []Input
	Decay        *Decay
}

type TrustReseed struct {
	BaseValue            int
	NotorietyNumerator   int64
	NotorietyDenominator int64
	FloorValue           int
	CeilingValue         int
}

type Catalog struct {
	TrustReseed TrustReseed
	Meters      []Meter
	byID        map[string]Meter
}

type rawCatalog struct {
	SchemaVersion int            `json:"schema_version"`
	TrustReseed   rawTrustReseed `json:"trust_reseed"`
	Meters        []rawMeter     `json:"meters"`
}

type rawTrustReseed struct {
	BaseValue            int   `json:"base_value"`
	NotorietyNumerator   int64 `json:"notoriety_numerator"`
	NotorietyDenominator int64 `json:"notoriety_denominator"`
	FloorValue           int   `json:"floor_value"`
	CeilingValue         int   `json:"ceiling_value"`
}

type rawMeter struct {
	ID           string     `json:"id"`
	Scope        string     `json:"scope"`
	MinValue     int        `json:"min_value"`
	MaxValue     int        `json:"max_value"`
	InitialValue int        `json:"initial_value"`
	Bands        []rawBand  `json:"bands"`
	Inputs       []rawInput `json:"inputs"`
	Decay        *rawDecay  `json:"decay"`
}

type rawBand struct {
	ID         string `json:"id"`
	FloorValue int    `json:"floor_value"`
}

type rawInput struct {
	Kind                 InputKind `json:"kind"`
	FactKind             string    `json:"fact_kind,omitempty"`
	Slot                 string    `json:"slot,omitempty"`
	SourceID             string    `json:"source_id,omitempty"`
	Delta                *int      `json:"delta,omitempty"`
	DeltaPerAttendedHour *int      `json:"delta_per_attended_hour,omitempty"`
}

type rawDecay struct {
	TowardValue int `json:"toward_value"`
	RatePerHour int `json:"rate_per_attended_hour"`
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
	if raw.SchemaVersion != CatalogSchemaVersion || len(raw.Meters) != len(requiredMeterIDs) || !validReseed(raw.TrustReseed) {
		return nil, ErrInvalidCatalog
	}
	catalog := &Catalog{
		TrustReseed: TrustReseed{
			BaseValue: raw.TrustReseed.BaseValue, NotorietyNumerator: raw.TrustReseed.NotorietyNumerator,
			NotorietyDenominator: raw.TrustReseed.NotorietyDenominator, FloorValue: raw.TrustReseed.FloorValue,
			CeilingValue: raw.TrustReseed.CeilingValue,
		},
		Meters: make([]Meter, 0, len(raw.Meters)), byID: map[string]Meter{},
	}
	for index, source := range raw.Meters {
		if source.ID != requiredMeterIDs[index] || source.Scope != "company" || source.MinValue != MinimumValue ||
			source.MaxValue != MaximumValue || source.InitialValue < MinimumValue || source.InitialValue > MaximumValue ||
			len(source.Bands) == 0 || source.Inputs == nil || catalog.byID[source.ID].ID != "" {
			return nil, fmt.Errorf("%w: meters[%d]", ErrInvalidCatalog, index)
		}
		meter := Meter{ID: source.ID, InitialValue: source.InitialValue, Bands: make([]Band, 0, len(source.Bands)), Inputs: make([]Input, 0, len(source.Inputs))}
		previousFloor := -1
		seenBands := map[string]bool{}
		for bandIndex, band := range source.Bands {
			if !idPattern.MatchString(band.ID) || seenBands[band.ID] || band.FloorValue < MinimumValue || band.FloorValue > MaximumValue ||
				band.FloorValue <= previousFloor || bandIndex == 0 && band.FloorValue != MinimumValue {
				return nil, fmt.Errorf("%w: meters[%d].bands[%d]", ErrInvalidCatalog, index, bandIndex)
			}
			seenBands[band.ID], previousFloor = true, band.FloorValue
			meter.Bands = append(meter.Bands, Band{ID: band.ID, FloorValue: band.FloorValue})
		}
		seenInputs := map[string]bool{}
		for inputIndex, sourceInput := range source.Inputs {
			input, uniqueness, err := parseInput(sourceInput)
			if err != nil || seenInputs[uniqueness] {
				return nil, fmt.Errorf("%w: meters[%d].inputs[%d]", ErrInvalidCatalog, index, inputIndex)
			}
			seenInputs[uniqueness] = true
			meter.Inputs = append(meter.Inputs, input)
		}
		if source.Decay != nil {
			if source.Decay.TowardValue < MinimumValue || source.Decay.TowardValue > MaximumValue || source.Decay.RatePerHour < 1 || source.Decay.RatePerHour > MaximumValue {
				return nil, fmt.Errorf("%w: meters[%d].decay", ErrInvalidCatalog, index)
			}
			meter.Decay = &Decay{TowardValue: source.Decay.TowardValue, RatePerHour: source.Decay.RatePerHour}
		}
		catalog.byID[meter.ID] = meter
		catalog.Meters = append(catalog.Meters, meter)
	}
	return catalog, nil
}

func validReseed(value rawTrustReseed) bool {
	return value.BaseValue >= MinimumValue && value.BaseValue <= MaximumValue && value.NotorietyNumerator >= 0 &&
		value.NotorietyDenominator > 0 && value.FloorValue >= MinimumValue && value.FloorValue <= value.CeilingValue &&
		value.CeilingValue <= MaximumValue && value.BaseValue >= value.FloorValue && value.BaseValue <= value.CeilingValue
}

func parseInput(source rawInput) (Input, string, error) {
	switch source.Kind {
	case InputLedgerFact:
		if !idPattern.MatchString(source.FactKind) || source.Slot != "" || source.SourceID != "" || source.Delta == nil ||
			source.DeltaPerAttendedHour != nil || *source.Delta == 0 || *source.Delta < -MaximumValue || *source.Delta > MaximumValue {
			return Input{}, "", ErrInvalidCatalog
		}
		return Input{Kind: source.Kind, FactKind: source.FactKind, Delta: *source.Delta}, "fact\x00" + source.FactKind, nil
	case InputContributionSlot:
		if !multiplier.ValidSlot(multiplier.Slot(source.Slot)) || !idPattern.MatchString(source.SourceID) || source.FactKind != "" ||
			source.Delta != nil || source.DeltaPerAttendedHour == nil || *source.DeltaPerAttendedHour == 0 ||
			*source.DeltaPerAttendedHour < -MaximumValue || *source.DeltaPerAttendedHour > MaximumValue {
			return Input{}, "", ErrInvalidCatalog
		}
		return Input{Kind: source.Kind, Slot: multiplier.Slot(source.Slot), SourceID: source.SourceID, DeltaPerAttendedHour: *source.DeltaPerAttendedHour},
			"slot\x00" + source.Slot + "\x00" + source.SourceID, nil
	default:
		return Input{}, "", ErrInvalidCatalog
	}
}

func (catalog *Catalog) Meter(id string) (Meter, bool) {
	if catalog == nil {
		return Meter{}, false
	}
	value, ok := catalog.byID[id]
	return cloneMeter(value), ok
}

func (catalog *Catalog) MeterIDs() []string {
	if catalog == nil {
		return nil
	}
	ids := make([]string, 0, len(catalog.Meters))
	for _, meter := range catalog.Meters {
		ids = append(ids, meter.ID)
	}
	return ids
}

func (catalog *Catalog) ValidateResourceSeparation(resourceIDs []string) error {
	if catalog == nil {
		return ErrInvalidCatalog
	}
	resources := append([]string(nil), resourceIDs...)
	sort.Strings(resources)
	meters := catalog.MeterIDs()
	left, right := 0, 0
	for left < len(resources) && right < len(meters) {
		if resources[left] == meters[right] {
			return fmt.Errorf("%w: meter/economy ID collision %q", ErrInvalidCatalog, resources[left])
		}
		if resources[left] < meters[right] {
			left++
		} else {
			right++
		}
	}
	return nil
}

func cloneMeter(source Meter) Meter {
	result := source
	result.Bands = append([]Band(nil), source.Bands...)
	result.Inputs = append([]Input(nil), source.Inputs...)
	if source.Decay != nil {
		decay := *source.Decay
		result.Decay = &decay
	}
	return result
}

package meters

import (
	"errors"
	"fmt"
)

const (
	MillisPerHour       int64 = 3_600_000
	MaximumAttendedStep int64 = 86_400_000
)

var ErrInvalidState = errors.New("invalid meter state")

type State struct {
	Values          map[string]int
	DecayRemainders map[string]int64
	InputRemainders map[string]int64
}

type AdvanceContext struct {
	AttendedMS          int64
	NewFactKinds        map[string]bool
	ActiveContributions map[string]bool
}

type BandChange struct {
	MeterID     string `json:"meter_id"`
	FromBand    string `json:"from_band"`
	ToBand      string `json:"to_band"`
	Direction   string `json:"direction"`
	ValueBefore int    `json:"value_before"`
	ValueAfter  int    `json:"value_after"`
}

func InputRemainderKey(meterID string, inputIndex int) string {
	return fmt.Sprintf("%s:%d", meterID, inputIndex)
}

func ContributionKey(slot, sourceID string) string {
	return slot + "\x00" + sourceID
}

func NewState(catalog *Catalog) State {
	state := State{Values: map[string]int{}, DecayRemainders: map[string]int64{}, InputRemainders: map[string]int64{}}
	if catalog == nil {
		return state
	}
	for _, meter := range catalog.Meters {
		state.Values[meter.ID] = meter.InitialValue
		state.DecayRemainders[meter.ID] = 0
		for index, input := range meter.Inputs {
			if input.Kind == InputContributionSlot {
				state.InputRemainders[InputRemainderKey(meter.ID, index)] = 0
			}
		}
	}
	return state
}

func ValidateState(catalog *Catalog, state State) error {
	if catalog == nil || state.Values == nil || state.DecayRemainders == nil || state.InputRemainders == nil {
		return ErrInvalidState
	}
	expectedInputs := map[string]bool{}
	if len(state.Values) != len(catalog.Meters) || len(state.DecayRemainders) != len(catalog.Meters) {
		return ErrInvalidState
	}
	for _, meter := range catalog.Meters {
		value, valueOK := state.Values[meter.ID]
		remainder, remainderOK := state.DecayRemainders[meter.ID]
		if !valueOK || !remainderOK || value < MinimumValue || value > MaximumValue || remainder < 0 || remainder >= MillisPerHour {
			return ErrInvalidState
		}
		for index, input := range meter.Inputs {
			if input.Kind != InputContributionSlot {
				continue
			}
			key := InputRemainderKey(meter.ID, index)
			expectedInputs[key] = true
			if remainder, ok := state.InputRemainders[key]; !ok || remainder < 0 || remainder >= MillisPerHour {
				return ErrInvalidState
			}
		}
	}
	if len(expectedInputs) != len(state.InputRemainders) {
		return ErrInvalidState
	}
	return nil
}

func Advance(catalog *Catalog, state State, context AdvanceContext) ([]BandChange, error) {
	if context.AttendedMS < 0 || context.AttendedMS > MaximumAttendedStep || context.NewFactKinds == nil || context.ActiveContributions == nil {
		return nil, ErrInvalidState
	}
	if err := ValidateState(catalog, state); err != nil {
		return nil, err
	}
	changes := make([]BandChange, 0)
	for _, meter := range catalog.Meters {
		before := state.Values[meter.ID]
		value := before
		if meter.Decay != nil && value == meter.Decay.TowardValue {
			state.DecayRemainders[meter.ID] = 0
		} else if meter.Decay != nil && context.AttendedMS > 0 {
			steps, remainder := wholeSteps(int64(meter.Decay.RatePerHour), context.AttendedMS, state.DecayRemainders[meter.ID])
			if value < meter.Decay.TowardValue {
				value += int(steps)
				if value >= meter.Decay.TowardValue {
					value, remainder = meter.Decay.TowardValue, 0
				}
			} else {
				value -= int(steps)
				if value <= meter.Decay.TowardValue {
					value, remainder = meter.Decay.TowardValue, 0
				}
			}
			state.DecayRemainders[meter.ID] = remainder
		}
		factDelta, rateDelta := 0, int64(0)
		for inputIndex, input := range meter.Inputs {
			switch input.Kind {
			case InputLedgerFact:
				if context.NewFactKinds[input.FactKind] {
					factDelta += input.Delta
				}
			case InputContributionSlot:
				if !context.ActiveContributions[ContributionKey(string(input.Slot), input.SourceID)] || context.AttendedMS == 0 {
					continue
				}
				key := InputRemainderKey(meter.ID, inputIndex)
				magnitude := int64(input.DeltaPerAttendedHour)
				sign := int64(1)
				if magnitude < 0 {
					magnitude, sign = -magnitude, -1
				}
				steps, remainder := wholeSteps(magnitude, context.AttendedMS, state.InputRemainders[key])
				state.InputRemainders[key] = remainder
				rateDelta += sign * steps
			}
		}
		value = clampValue(int64(value) + int64(factDelta) + rateDelta)
		state.Values[meter.ID] = value
		from, to := BandFor(meter, before), BandFor(meter, value)
		if from != to {
			direction := "down"
			if value > before {
				direction = "up"
			}
			changes = append(changes, BandChange{MeterID: meter.ID, FromBand: from, ToBand: to, Direction: direction, ValueBefore: before, ValueAfter: value})
		}
	}
	return changes, nil
}

func wholeSteps(rate, elapsedMS, remainder int64) (int64, int64) {
	numerator := rate*elapsedMS + remainder
	return numerator / MillisPerHour, numerator % MillisPerHour
}

func clampValue(value int64) int {
	if value < MinimumValue {
		return MinimumValue
	}
	if value > MaximumValue {
		return MaximumValue
	}
	return int(value)
}

func BandFor(meter Meter, value int) string {
	band := meter.Bands[0].ID
	for _, candidate := range meter.Bands[1:] {
		if value < candidate.FloorValue {
			break
		}
		band = candidate.ID
	}
	return band
}

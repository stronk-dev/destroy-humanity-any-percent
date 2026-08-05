package pet

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"regexp"
)

const maxExactInteger = int64(9_007_199_254_740_991)

var (
	ErrInvalidCareState = errors.New("invalid pet care state")
	petIDPattern        = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[1-8][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
	mechanicalIDPattern = regexp.MustCompile(`^[a-z][a-z0-9_]*(?:\.[a-z][a-z0-9_]*)*$`)
)

type BehaviorQueueEntry struct {
	BehaviorID    string `json:"behavior_id"`
	DueAttendedMS int64  `json:"due_attended_ms"`
}

type CareState struct {
	StatsPPM                    map[StatID]int64     `json:"stats_ppm"`
	StatDecayRemaindersPPM      map[StatID]int64     `json:"stat_decay_remainders_ppm"`
	CooldownUntilAttendedMS     map[string]int64     `json:"cooldown_until_attended_ms"`
	TrustPPM                    int64                `json:"trust_ppm"`
	TrustDecayRemainderPPM      int64                `json:"trust_decay_remainder_ppm"`
	EvaluatedThroughAttendedMS  int64                `json:"evaluated_through_attended_ms"`
	BehaviorState               BehaviorState        `json:"behavior_state"`
	BehaviorEnteredAtAttendedMS int64                `json:"behavior_entered_at_attended_ms"`
	BehaviorQueue               []BehaviorQueueEntry `json:"behavior_queue"`
	BehaviorPRNGCursor          int64                `json:"behavior_prng_cursor"`
}

type StateDeclarations struct {
	ActionIDs   []string
	BehaviorIDs []string
}

func DecodeCareStates(data []byte, declarations StateDeclarations) (map[string]CareState, error) {
	if !uniqueStateJSONKeys(data) || !exactCareStateKeys(data) {
		return nil, ErrInvalidCareState
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var states map[string]CareState
	if err := decoder.Decode(&states); err != nil || states == nil {
		return nil, ErrInvalidCareState
	}
	var trailing any
	if !errors.Is(decoder.Decode(&trailing), io.EOF) || ValidateCareStates(states, declarations) != nil {
		return nil, ErrInvalidCareState
	}
	return cloneCareStates(states), nil
}

func exactCareStateKeys(data []byte) bool {
	var states map[string]json.RawMessage
	if json.Unmarshal(data, &states) != nil || states == nil {
		return false
	}
	stateKeys := []string{"stats_ppm", "stat_decay_remainders_ppm", "cooldown_until_attended_ms", "trust_ppm",
		"trust_decay_remainder_ppm", "evaluated_through_attended_ms", "behavior_state", "behavior_entered_at_attended_ms", "behavior_queue", "behavior_prng_cursor"}
	for _, encoded := range states {
		var state map[string]json.RawMessage
		if json.Unmarshal(encoded, &state) != nil || !hasRawKeys(state, stateKeys) {
			return false
		}
		var queue []json.RawMessage
		if json.Unmarshal(state["behavior_queue"], &queue) != nil {
			return false
		}
		for _, entry := range queue {
			var object map[string]json.RawMessage
			if json.Unmarshal(entry, &object) != nil || !hasRawKeys(object, []string{"behavior_id", "due_attended_ms"}) {
				return false
			}
		}
	}
	return true
}

func hasRawKeys(value map[string]json.RawMessage, expected []string) bool {
	if len(value) != len(expected) {
		return false
	}
	for _, key := range expected {
		if _, exists := value[key]; !exists {
			return false
		}
	}
	return true
}

func ValidateCareStates(states map[string]CareState, declarations StateDeclarations) error {
	actions, ok := declaredIDs(declarations.ActionIDs)
	if !ok {
		return ErrInvalidCareState
	}
	behaviors, ok := declaredIDs(declarations.BehaviorIDs)
	if !ok || states == nil {
		return ErrInvalidCareState
	}
	for petID, state := range states {
		if !petIDPattern.MatchString(petID) || !completeStatMap(state.StatsPPM, 1_000_000) ||
			!completeStatMap(state.StatDecayRemaindersPPM, maxExactInteger) ||
			state.TrustPPM < 0 || state.TrustPPM > 1_000_000 ||
			!exactNonnegative(state.TrustDecayRemainderPPM) || !exactNonnegative(state.EvaluatedThroughAttendedMS) ||
			!ValidBehaviorState(state.BehaviorState) || !exactNonnegative(state.BehaviorEnteredAtAttendedMS) ||
			state.BehaviorEnteredAtAttendedMS > state.EvaluatedThroughAttendedMS || state.BehaviorPRNGCursor != 0 ||
			!ValidBehaviorQueueLength(len(state.BehaviorQueue)) {
			return ErrInvalidCareState
		}
		if state.CooldownUntilAttendedMS == nil || state.BehaviorQueue == nil {
			return ErrInvalidCareState
		}
		for actionID, until := range state.CooldownUntilAttendedMS {
			if _, exists := actions[actionID]; !exists || !exactNonnegative(until) {
				return ErrInvalidCareState
			}
		}
		seenQueuedBehaviors := map[string]bool{}
		for index, entry := range state.BehaviorQueue {
			if _, exists := behaviors[entry.BehaviorID]; !exists || !exactNonnegative(entry.DueAttendedMS) {
				return ErrInvalidCareState
			}
			if seenQueuedBehaviors[entry.BehaviorID] {
				return ErrInvalidCareState
			}
			seenQueuedBehaviors[entry.BehaviorID] = true
			if index > 0 {
				prior := state.BehaviorQueue[index-1]
				if prior.DueAttendedMS > entry.DueAttendedMS ||
					prior.DueAttendedMS == entry.DueAttendedMS && prior.BehaviorID >= entry.BehaviorID {
					return ErrInvalidCareState
				}
			}
		}
	}
	return nil
}

// ValidateCareStatesForCatalog closes the remainder domain that the generic
// v18 codec cannot know without the pinned artifact.
func ValidateCareStatesForCatalog(states map[string]CareState, catalog *Catalog) error {
	if catalog == nil || ValidateCareStates(states, catalog.StateDeclarations()) != nil {
		return ErrInvalidCareState
	}
	for _, state := range states {
		if state.TrustDecayRemainderPPM >= catalog.StatPolicy.GridMS {
			return ErrInvalidCareState
		}
		for _, remainder := range state.StatDecayRemaindersPPM {
			if remainder >= catalog.StatPolicy.GridMS {
				return ErrInvalidCareState
			}
		}
	}
	return nil
}

func completeStatMap(values map[StatID]int64, maximum int64) bool {
	if len(values) != len(statIDs) {
		return false
	}
	for _, id := range statIDs {
		value, ok := values[id]
		if !ok || value < 0 || value > maximum {
			return false
		}
	}
	return true
}

func declaredIDs(values []string) (map[string]struct{}, bool) {
	result := make(map[string]struct{}, len(values))
	for _, value := range values {
		if !mechanicalIDPattern.MatchString(value) {
			return nil, false
		}
		if _, duplicate := result[value]; duplicate {
			return nil, false
		}
		result[value] = struct{}{}
	}
	return result, true
}

func exactNonnegative(value int64) bool { return value >= 0 && value <= maxExactInteger }

func cloneCareStates(states map[string]CareState) map[string]CareState {
	result := make(map[string]CareState, len(states))
	for petID, state := range states {
		cloned := state
		cloned.StatsPPM = cloneStatMap(state.StatsPPM)
		cloned.StatDecayRemaindersPPM = cloneStatMap(state.StatDecayRemaindersPPM)
		cloned.CooldownUntilAttendedMS = make(map[string]int64, len(state.CooldownUntilAttendedMS))
		for key, value := range state.CooldownUntilAttendedMS {
			cloned.CooldownUntilAttendedMS[key] = value
		}
		cloned.BehaviorQueue = make([]BehaviorQueueEntry, len(state.BehaviorQueue))
		copy(cloned.BehaviorQueue, state.BehaviorQueue)
		result[petID] = cloned
	}
	return result
}

func cloneStatMap(values map[StatID]int64) map[StatID]int64 {
	result := make(map[StatID]int64, len(values))
	for key, value := range values {
		result[key] = value
	}
	return result
}

func uniqueStateJSONKeys(data []byte) bool {
	decoder := json.NewDecoder(bytes.NewReader(data))
	var readValue func() bool
	readValue = func() bool {
		token, err := decoder.Token()
		if err != nil {
			return false
		}
		delimiter, compound := token.(json.Delim)
		if !compound {
			return true
		}
		switch delimiter {
		case '{':
			seen := map[string]struct{}{}
			for decoder.More() {
				keyToken, keyErr := decoder.Token()
				key, valid := keyToken.(string)
				if keyErr != nil || !valid {
					return false
				}
				if _, duplicate := seen[key]; duplicate {
					return false
				}
				seen[key] = struct{}{}
				if !readValue() {
					return false
				}
			}
			end, endErr := decoder.Token()
			return endErr == nil && end == json.Delim('}')
		case '[':
			for decoder.More() {
				if !readValue() {
					return false
				}
			}
			end, endErr := decoder.Token()
			return endErr == nil && end == json.Delim(']')
		default:
			return false
		}
	}
	if !readValue() {
		return false
	}
	var trailing any
	return errors.Is(decoder.Decode(&trailing), io.EOF)
}

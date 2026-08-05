package pet

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
)

var ErrInvalidCatalogGrammar = errors.New("invalid pet catalog grammar")

type MoodThreshold struct {
	MoodMember Mood  `json:"mood_member"`
	FloorPPM   int64 `json:"floor_ppm"`
}

type BehaviorCandidate struct {
	FromState         BehaviorState `json:"from_state"`
	Event             BehaviorEvent `json:"event"`
	ToState           BehaviorState `json:"to_state"`
	DurationGridTicks int64         `json:"duration_grid_ticks"`
}

type CatalogGrammar struct {
	MoodThresholds     []MoodThreshold     `json:"mood_thresholds"`
	BehaviorCandidates []BehaviorCandidate `json:"behavior_candidates"`
}

func LoadCatalogGrammar(data []byte) (CatalogGrammar, error) {
	if !uniqueStateJSONKeys(data) {
		return CatalogGrammar{}, ErrInvalidCatalogGrammar
	}
	var raw map[string]json.RawMessage
	if json.Unmarshal(data, &raw) != nil || !hasRawKeys(raw, []string{"mood_thresholds", "behavior_candidates"}) {
		return CatalogGrammar{}, ErrInvalidCatalogGrammar
	}
	if !exactCatalogRows(raw["mood_thresholds"], []string{"mood_member", "floor_ppm"}) ||
		!exactCatalogRows(raw["behavior_candidates"], []string{"from_state", "event", "to_state", "duration_grid_ticks"}) {
		return CatalogGrammar{}, ErrInvalidCatalogGrammar
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var grammar CatalogGrammar
	if err := decoder.Decode(&grammar); err != nil {
		return CatalogGrammar{}, ErrInvalidCatalogGrammar
	}
	var trailing any
	if !errors.Is(decoder.Decode(&trailing), io.EOF) || validateCatalogGrammar(grammar) != nil {
		return CatalogGrammar{}, ErrInvalidCatalogGrammar
	}
	grammar.MoodThresholds = append([]MoodThreshold(nil), grammar.MoodThresholds...)
	grammar.BehaviorCandidates = append([]BehaviorCandidate(nil), grammar.BehaviorCandidates...)
	return grammar, nil
}

func exactCatalogRows(data []byte, keys []string) bool {
	var rows []map[string]json.RawMessage
	if json.Unmarshal(data, &rows) != nil || rows == nil {
		return false
	}
	for _, row := range rows {
		if !hasRawKeys(row, keys) {
			return false
		}
	}
	return true
}

func validateCatalogGrammar(grammar CatalogGrammar) error {
	if len(grammar.MoodThresholds) != len(moods) || grammar.BehaviorCandidates == nil {
		return ErrInvalidCatalogGrammar
	}
	seenMoods := make(map[Mood]struct{}, len(moods))
	for index, row := range grammar.MoodThresholds {
		if !ValidMood(row.MoodMember) || row.FloorPPM < 0 || row.FloorPPM > 1_000_000 ||
			index == 0 && row.FloorPPM != 0 || index > 0 && row.FloorPPM <= grammar.MoodThresholds[index-1].FloorPPM {
			return ErrInvalidCatalogGrammar
		}
		if _, duplicate := seenMoods[row.MoodMember]; duplicate {
			return ErrInvalidCatalogGrammar
		}
		seenMoods[row.MoodMember] = struct{}{}
	}
	seenCandidates := make(map[[3]string]struct{}, len(grammar.BehaviorCandidates))
	for _, row := range grammar.BehaviorCandidates {
		if !ValidBehaviorState(row.FromState) || !ValidBehaviorEvent(row.Event) || !ValidBehaviorState(row.ToState) ||
			row.DurationGridTicks < 1 || !exactNonnegative(row.DurationGridTicks) {
			return ErrInvalidCatalogGrammar
		}
		key := [3]string{string(row.FromState), string(row.Event), string(row.ToState)}
		if _, duplicate := seenCandidates[key]; duplicate {
			return ErrInvalidCatalogGrammar
		}
		seenCandidates[key] = struct{}{}
	}
	return nil
}

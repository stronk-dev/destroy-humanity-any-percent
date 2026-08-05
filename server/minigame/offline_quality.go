package minigame

import (
	"encoding/json"
	"errors"
)

var ErrInvalidOfflineQualityPolicy = errors.New("invalid minigame offline quality policy")

type OfflineQualityGrade struct {
	ScoreThreshold int64 `json:"score_threshold"`
	GradePPM       int64 `json:"grade_ppm"`
}

type OfflineQualityPolicy struct {
	ScoreFact             string                `json:"score_fact"`
	GradeCurve            []OfflineQualityGrade `json:"grade_curve"`
	DecayGridMS           int64                 `json:"decay_grid_ms"`
	DecayPPMPerGrid       int64                 `json:"decay_ppm_per_grid"`
	NeutralFloorPPM       int64                 `json:"neutral_floor_ppm"`
	AutomationDestination string                `json:"automation_destination"`
}

type OfflineQualityState struct {
	GradePPM              int64 `json:"grade_ppm"`
	LastFounderAttendedMS int64 `json:"last_founder_attended_ms"`
	DecayRemainderPPM     int64 `json:"decay_remainder_ppm"`
}

type OfflineQualityDeclarations struct {
	ScoreFactIDs           map[string]struct{}
	AutomationDestinations map[string]struct{}
}

func LoadOfflineQualityPolicy(data []byte, declarations OfflineQualityDeclarations) (OfflineQualityPolicy, error) {
	if !hasExactJSONKeys(data, "automation_destination", "decay_grid_ms", "decay_ppm_per_grid", "grade_curve", "neutral_floor_ppm", "score_fact") ||
		declarations.ScoreFactIDs == nil || declarations.AutomationDestinations == nil {
		return OfflineQualityPolicy{}, ErrInvalidOfflineQualityPolicy
	}
	var policy OfflineQualityPolicy
	if decodeExact(data, &policy) != nil || !mechanicalPattern.MatchString(policy.ScoreFact) ||
		!mechanicalPattern.MatchString(policy.AutomationDestination) || policy.DecayGridMS <= 0 ||
		!validExactNonnegative(policy.DecayGridMS) || policy.DecayPPMPerGrid < 0 || policy.DecayPPMPerGrid > payoutPPM ||
		policy.NeutralFloorPPM < 0 || policy.NeutralFloorPPM > payoutPPM || len(policy.GradeCurve) == 0 {
		return OfflineQualityPolicy{}, ErrInvalidOfflineQualityPolicy
	}
	if _, ok := declarations.ScoreFactIDs[policy.ScoreFact]; !ok {
		return OfflineQualityPolicy{}, ErrInvalidOfflineQualityPolicy
	}
	if _, ok := declarations.AutomationDestinations[policy.AutomationDestination]; !ok {
		return OfflineQualityPolicy{}, ErrInvalidOfflineQualityPolicy
	}
	for index, row := range policy.GradeCurve {
		if row.ScoreThreshold < 0 || !validExactNonnegative(row.ScoreThreshold) || row.GradePPM < policy.NeutralFloorPPM || row.GradePPM > payoutPPM ||
			index > 0 && (row.ScoreThreshold <= policy.GradeCurve[index-1].ScoreThreshold || row.GradePPM < policy.GradeCurve[index-1].GradePPM) {
			return OfflineQualityPolicy{}, ErrInvalidOfflineQualityPolicy
		}
	}
	if policy.GradeCurve[0].GradePPM != policy.NeutralFloorPPM {
		return OfflineQualityPolicy{}, ErrInvalidOfflineQualityPolicy
	}
	return policy, nil
}

func OfflineGradeForScore(policy OfflineQualityPolicy, score int64) (int64, error) {
	if score < 0 || !validExactNonnegative(score) || len(policy.GradeCurve) == 0 {
		return 0, ErrInvalidOfflineQualityPolicy
	}
	grade := policy.NeutralFloorPPM
	for _, row := range policy.GradeCurve {
		if score < row.ScoreThreshold {
			break
		}
		grade = row.GradePPM
	}
	if grade < policy.NeutralFloorPPM || grade > payoutPPM {
		return 0, ErrInvalidOfflineQualityPolicy
	}
	return grade, nil
}

func ValidateOfflineQualityState(state OfflineQualityState) error {
	if state.GradePPM < 0 || state.GradePPM > payoutPPM || !validExactNonnegative(state.LastFounderAttendedMS) ||
		state.DecayRemainderPPM < 0 || state.DecayRemainderPPM >= payoutPPM {
		return ErrInvalidOfflineQualityPolicy
	}
	return nil
}

func (policy OfflineQualityPolicy) MarshalJSON() ([]byte, error) {
	type wire OfflineQualityPolicy
	encoded, err := json.Marshal(wire(policy))
	if err != nil {
		return nil, err
	}
	loaded, err := LoadOfflineQualityPolicy(encoded, OfflineQualityDeclarations{
		ScoreFactIDs: map[string]struct{}{policy.ScoreFact: {}}, AutomationDestinations: map[string]struct{}{policy.AutomationDestination: {}},
	})
	if err != nil || len(loaded.GradeCurve) != len(policy.GradeCurve) {
		return nil, ErrInvalidOfflineQualityPolicy
	}
	return encoded, nil
}

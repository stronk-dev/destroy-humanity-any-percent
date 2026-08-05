package minigame

import (
	"errors"
	"math/big"

	"cloud-clicker/server/decimal"
	"cloud-clicker/server/fixedgrid"
)

var ErrInvalidResolutionTransition = errors.New("invalid minigame resolution transition")

type RatingState struct {
	Elo          int64  `json:"elo"`
	SeasonMember string `json:"season_member"`
	GamesCounted int64  `json:"games_counted"`
}

type FounderResolutionTransition struct {
	RatingBefore  RatingState         `json:"rating_before"`
	RatingAfter   RatingState         `json:"rating_after"`
	QualityBefore OfflineQualityState `json:"quality_before"`
	QualityAfter  OfflineQualityState `json:"quality_after"`
}

// ApplyFounderResolution is the pure C40 transition shared by live composition
// and replay. Tenant-certified rating_delta is authoritative; this package
// never owns or applies an Elo K-factor.
func ApplyFounderResolution(rating RatingState, quality OfflineQualityState, result *Result,
	policy Definition, founderAttendedMS int64,
) (FounderResolutionTransition, error) {
	if result == nil || !validResult(result) || founderAttendedMS < quality.LastFounderAttendedMS ||
		founderAttendedMS > decimal.MaxExactInteger || validateRatingState(rating, policy.Rating) != nil ||
		ValidateOfflineQualityState(quality) != nil {
		return FounderResolutionTransition{}, ErrInvalidResolutionTransition
	}
	selectedScore, err := selectUniqueScore(result, policy.OfflineQuality.ScoreFact)
	if err != nil {
		return FounderResolutionTransition{}, err
	}
	decayed, err := decayOfflineQuality(quality, policy.OfflineQuality, founderAttendedMS)
	if err != nil {
		return FounderResolutionTransition{}, err
	}
	grade, err := OfflineGradeForScore(policy.OfflineQuality, selectedScore)
	if err != nil {
		return FounderResolutionTransition{}, ErrInvalidResolutionTransition
	}
	afterQuality := decayed
	afterQuality.GradePPM = grade
	afterQuality.LastFounderAttendedMS = founderAttendedMS
	afterQuality.DecayRemainderPPM = 0
	afterRating := rating
	if result.RatingDelta != nil {
		if rating.GamesCounted == decimal.MaxExactInteger {
			return FounderResolutionTransition{}, ErrInvalidResolutionTransition
		}
		value := new(big.Int).Add(big.NewInt(rating.Elo), big.NewInt(*result.RatingDelta))
		floor, ceiling := big.NewInt(policy.Rating.EloFloor), big.NewInt(policy.Rating.EloCeiling)
		if value.Cmp(floor) < 0 {
			value.Set(floor)
		} else if value.Cmp(ceiling) > 0 {
			value.Set(ceiling)
		}
		afterRating.Elo = value.Int64()
		afterRating.GamesCounted++
	}
	return FounderResolutionTransition{RatingBefore: rating, RatingAfter: afterRating,
		QualityBefore: quality, QualityAfter: afterQuality}, nil
}

func decayOfflineQuality(state OfflineQualityState, policy OfflineQualityPolicy, attendedMS int64) (OfflineQualityState, error) {
	elapsed := attendedMS - state.LastFounderAttendedMS
	integrated, err := fixedgrid.Integrate(elapsed, policy.DecayPPMPerGrid, state.DecayRemainderPPM, policy.DecayGridMS)
	if err != nil {
		return OfflineQualityState{}, ErrInvalidResolutionTransition
	}
	distance := state.GradePPM - policy.NeutralFloorPPM
	if distance < 0 {
		return OfflineQualityState{}, ErrInvalidResolutionTransition
	}
	if integrated.Whole.Cmp(big.NewInt(distance)) >= 0 {
		state.GradePPM, state.DecayRemainderPPM = policy.NeutralFloorPPM, 0
	} else {
		state.GradePPM -= integrated.Whole.Int64()
		state.DecayRemainderPPM = integrated.Remainder
		if state.DecayRemainderPPM >= payoutPPM {
			return OfflineQualityState{}, ErrInvalidResolutionTransition
		}
	}
	state.LastFounderAttendedMS = attendedMS
	return state, nil
}

func selectUniqueScore(result *Result, kind string) (int64, error) {
	found := false
	var score int64
	for _, fact := range result.ScoreFacts {
		if fact.Kind != kind {
			continue
		}
		if found || fact.Value < 0 {
			return 0, ErrInvalidResolutionTransition
		}
		found, score = true, fact.Value
	}
	if !found {
		return 0, ErrInvalidResolutionTransition
	}
	return score, nil
}

func validateRatingState(state RatingState, policy RatingPolicy) error {
	if state.Elo < policy.EloFloor || state.Elo > policy.EloCeiling || state.GamesCounted < 0 ||
		state.GamesCounted > decimal.MaxExactInteger || state.SeasonMember != policy.SeasonMember {
		return ErrInvalidResolutionTransition
	}
	return nil
}

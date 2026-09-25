package production

import (
	"encoding/json"
	"fmt"

	"cloud-clicker/server/achievements"
	"cloud-clicker/server/decimal"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/multiplier"
	"cloud-clicker/server/save"
)

// AxisInput is the clamped run-local axis value x = min(input, input_cap) and
// whether the clamp engaged (CV3). It reads only Company state: under the
// ruled attainment input it performs no Founder read at all (AC3).
func AxisInput(state *save.State, catalog *economy.Catalog) (x int64, saturated bool, err error) {
	axis, declared := catalog.AxisStack()
	if !declared || state == nil {
		return 0, false, ErrInvalidEngineState
	}
	var input int64
	switch axis.Input {
	case economy.AxisInputAttainmentRun:
		if state.AchievementsAttainedRun == nil {
			return 0, false, fmt.Errorf("%w: attainment input without Company v19 state", ErrInvalidEngineState)
		}
		input = state.AttainmentScoreRun
	case economy.AxisInputScoreRun:
		input = state.AchievementScoreRun
	default:
		return 0, false, ErrInvalidEngineState
	}
	if input < 0 || input > decimal.MaxExactInteger {
		return 0, false, ErrInvalidEngineState
	}
	if input > axis.InputCap {
		return axis.InputCap, true, nil
	}
	return input, false, nil
}

// AxisFactors returns every axis effect's factor at the current input, owned
// or not (the read projection's "projected factor", CV9), through the same
// private arithmetic as live rates.
func AxisFactors(state *save.State, catalog *economy.Catalog) (map[string]decimal.Decimal, error) {
	result := map[string]decimal.Decimal{}
	for _, upgrade := range catalog.Upgrades() {
		for _, effect := range upgrade.Effects {
			if effect.Slot != economy.SlotAxisStack {
				continue
			}
			factor, err := axisFactor(state, catalog, effect.FactorPPM)
			if err != nil {
				return nil, err
			}
			result[effect.SourceID] = factor
		}
	}
	return result, nil
}

// AxisProduct is the owned axis_stack contributions' product as the engine
// folds them for target "all" (CV3 read projection).
func AxisProduct(state *save.State, catalog *economy.Catalog) (decimal.Decimal, error) {
	contributions, err := contentContributions(state, catalog)
	if err != nil {
		return decimal.NaN, err
	}
	axis := make([]multiplier.Contribution, 0)
	for _, contribution := range contributions {
		if contribution.Slot == economy.SlotAxisStack {
			axis = append(axis, contribution)
		}
	}
	return contributionFactorForTarget(catalog, "all", axis)
}

// axisFactor is factor_i = (1,000,000 + x * factor_ppm_i) / 1,000,000 with
// exact integer arithmetic and one Decimal quantization (CV3).
func axisFactor(state *save.State, catalog *economy.Catalog, factorPPM int64) (decimal.Decimal, error) {
	x, _, err := AxisInput(state, catalog)
	if err != nil {
		return decimal.NaN, err
	}
	return countPPMFactor(x, factorPPM)
}

// validateAttainmentState is the CV2 derivation check on a Company v19 state
// under its pinned achievements artifact: every attained ID is a run-scoped
// definition, the stored score is their summed grant, and every run-scoped ID
// earned this run is also attained.
func validateAttainmentState(bundle CatalogBundle, state *save.State) error {
	if state == nil || state.AchievementsAttainedRun == nil || bundle.Achievements == nil {
		return fmt.Errorf("%w: Company v19 attainment state requires the pinned achievements artifact", ErrInvalidEngineState)
	}
	var score int64
	for id, attained := range state.AchievementsAttainedRun {
		definition, ok := bundle.Achievements.Definition(id)
		if !attained || !ok || definition.ConditionScope != achievements.ScopeRun {
			return fmt.Errorf("%w: attained achievement %q is not a run-scoped definition", ErrInvalidEngineState, id)
		}
		score += definition.ScoreGrant
	}
	if score != state.AttainmentScoreRun {
		return fmt.Errorf("%w: attainment score does not derive from attained IDs", ErrInvalidEngineState)
	}
	for id, earned := range state.AchievementsEarnedRun {
		definition, ok := bundle.Achievements.Definition(id)
		if earned && ok && definition.ConditionScope == achievements.ScopeRun && !state.AchievementsAttainedRun[id] {
			return fmt.Errorf("%w: run-scoped earned achievement %q is not attained", ErrInvalidEngineState, id)
		}
	}
	return nil
}

// attainRun is the achievement hook's second pass (CV2). It runs against the
// same pre-achievement observation as NewlyEarned, selects run-scoped
// definitions not yet attained this run whose condition and proof hold, and
// never reads Founder lifetime state. newlyEarned are the IDs the first pass
// staged in this transition; they keep achievement_earned.v1, and only the
// remainder emits achievement_reattained.v1, in catalog (byte) order.
func attainRun(bundle CatalogBundle, state *save.State, run achievements.Observation, newlyEarned map[string]bool,
	proof func(achievements.Definition) bool, runID map[string]any, intentID string, events *[]save.EventWrite) error {
	if state.AchievementsAttainedRun == nil {
		return nil
	}
	staged := make([]achievements.Definition, 0)
	for _, definition := range bundle.Achievements.Definitions {
		if definition.ConditionScope != achievements.ScopeRun || state.AchievementsAttainedRun[definition.ID] {
			continue
		}
		if achievements.Eligible(definition.Condition, run) && proof(definition) {
			staged = append(staged, definition)
		}
	}
	for _, definition := range staged {
		state.AchievementsAttainedRun[definition.ID] = true
		state.AttainmentScoreRun += definition.ScoreGrant
	}
	for _, definition := range staged {
		if newlyEarned[definition.ID] {
			continue
		}
		payload, _ := json.Marshal(map[string]any{"run_id": runID, "achievement_id": definition.ID, "score_grant": definition.ScoreGrant})
		*events = append(*events, save.EventWrite{Kind: save.EventAchievementReattained, SchemaVersion: 1, IntentID: intentID, Payload: payload})
	}
	return validateAttainmentState(bundle, state)
}

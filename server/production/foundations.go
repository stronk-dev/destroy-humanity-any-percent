package production

import (
	"fmt"

	"cloud-clicker/server/economy"
	"cloud-clicker/server/meters"
	"cloud-clicker/server/save"
)

func (bundle CatalogBundle) foundationsActive() bool {
	return bundle.Meters != nil && bundle.Achievements != nil
}

// ValidateFoundationState closes catalog-derived save invariants that the save
// package cannot own without importing feature packages.
func (bundle CatalogBundle) ValidateFoundationState(state *save.State) error {
	if state == nil || state.Ledger == nil {
		return ErrInvalidEngineState
	}
	if !bundle.foundationsActive() {
		if save.VersionForState(state) >= 15 {
			return fmt.Errorf("%w: foundation save without pinned artifacts", ErrInvalidEngineState)
		}
		return nil
	}
	if save.VersionForState(state) != 16 {
		return fmt.Errorf("%w: pinned foundations require save v16", ErrInvalidEngineState)
	}
	switch state.Ledger.Scope() {
	case economy.ScopeCompany:
		meterState := meters.State{Values: state.MeterValues, DecayRemainders: state.MeterDecayRemainders, InputRemainders: state.MeterInputRemainders}
		if err := meters.ValidateState(bundle.Meters, meterState); err != nil {
			return err
		}
		score, err := bundle.Achievements.Score(state.AchievementsEarnedRun)
		if err != nil || score != state.AchievementScoreRun {
			return fmt.Errorf("%w: run achievement score does not derive from earned IDs", ErrInvalidEngineState)
		}
	case economy.ScopeFounder:
		score, err := bundle.Achievements.Score(state.AchievementsEarnedLifetime)
		if err != nil || score != state.AchievementScoreLifetime {
			return fmt.Errorf("%w: lifetime achievement score does not derive from earned IDs", ErrInvalidEngineState)
		}
	}
	return nil
}

func settleAndActivateFoundations(current, next CatalogBundle, founder, company, newCompany *save.State) error {
	if founder == nil || company == nil || newCompany == nil {
		return ErrInvalidEngineState
	}
	currentActive, nextActive := current.foundationsActive(), next.foundationsActive()
	if currentActive {
		if !nextActive || save.VersionForState(founder) != 16 || save.VersionForState(company) != 16 {
			return fmt.Errorf("%w: foundation mechanics cannot disappear between epochs", ErrInvalidEngineState)
		}
		if err := current.ValidateFoundationState(founder); err != nil {
			return err
		}
		if err := current.ValidateFoundationState(company); err != nil {
			return err
		}
		for id := range company.AchievementsEarnedRun {
			if founder.AchievementsEarnedLifetime[id] {
				return fmt.Errorf("%w: run achievement already owned for life", ErrInvalidEngineState)
			}
			founder.AchievementsEarnedLifetime[id] = true
		}
		score, err := current.Achievements.Score(founder.AchievementsEarnedLifetime)
		if err != nil {
			return err
		}
		founder.AchievementScoreLifetime = score
	} else if save.VersionForState(founder) >= 15 || save.VersionForState(company) >= 15 {
		return fmt.Errorf("%w: active save lacks pinned foundation artifacts", ErrInvalidEngineState)
	}

	if !nextActive {
		return nil
	}
	if !currentActive {
		// Activation is deliberately non-retroactive. Pre-foundation history does
		// not synthesize achievements or import the legacy MeterBands placeholder.
		founder.AchievementsEarnedLifetime = map[string]bool{}
		founder.AchievementScoreLifetime = 0
	}
	founder.WireVersion = 16
	newMeters, err := meters.NewRunState(next.Meters, founder.Notoriety)
	if err != nil {
		return err
	}
	newCompany.WireVersion = 16
	newCompany.MeterBands = nil
	newCompany.MeterValues = newMeters.Values
	newCompany.MeterDecayRemainders = newMeters.DecayRemainders
	newCompany.MeterInputRemainders = newMeters.InputRemainders
	newCompany.AchievementsEarnedRun = map[string]bool{}
	newCompany.AchievementScoreRun = 0
	if err := next.ValidateFoundationState(founder); err != nil {
		return err
	}
	return next.ValidateFoundationState(newCompany)
}

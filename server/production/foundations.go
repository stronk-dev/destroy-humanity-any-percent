package production

import (
	"fmt"
	"sort"

	"cloud-clicker/server/achievements"
	"cloud-clicker/server/copykeys"
	"cloud-clicker/server/decimal"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/fiscal"
	"cloud-clicker/server/meters"
	"cloud-clicker/server/minigame"
	"cloud-clicker/server/pet"
	"cloud-clicker/server/save"
)

// FoundationAchievementRegistry composes catalog-bound IDs from the pinned
// economy artifact and structural authorities from the versioned kernel.
func FoundationAchievementRegistry(catalog *economy.Catalog) achievements.Registry {
	registry := achievements.Registry{
		CopyKeys: map[string]bool{}, GeneratorIDs: map[string]bool{}, EventKinds: map[string]bool{}, ResourceIDs: map[string]bool{},
		RunCounters:    map[string]bool{"generators_purchased_total": true, "tier": true},
		CareerCounters: map[string]bool{"age_ms": true, "notoriety": true},
		ProvenanceSources: map[string][]string{
			"counter:run:generators_purchased_total": {string(save.EventGeneratorPurchased)},
			"counter:run:tier":                       {string(save.EventGateCrossed)},
			"counter:career:age_ms":                  {string(save.EventFounderAdvanced)},
			"counter:career:notoriety":               {string(save.EventFounderAdvanced)},
			"exit_count":                             {string(save.EventFounderAdvanced), string(save.EventRunEnded)},
		},
	}
	for _, key := range copykeys.All() {
		registry.CopyKeys[key] = true
	}
	for _, kind := range save.AllEventKinds {
		registry.EventKinds[string(kind)] = true
	}
	if catalog != nil {
		for _, definition := range catalog.GeneratorClasses() {
			registry.GeneratorIDs[definition.ID] = true
		}
		for _, definition := range catalog.Resources() {
			registry.ResourceIDs[definition.ID] = true
		}
	}
	return registry
}

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
	founderFloor, companyFloor := bundle.versionFloors()
	wantVersion := companyFloor
	if state.Ledger.Scope() == economy.ScopeFounder {
		wantVersion = founderFloor
	}
	if save.VersionForState(state) != wantVersion {
		return fmt.Errorf("%w: pinned artifact/save version mismatch", ErrInvalidEngineState)
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
		if err := validateFounderMinigameState(bundle.Minigames, state); err != nil {
			return err
		}
		if bundle.Pets == nil {
			if len(state.Pets) != 0 {
				return fmt.Errorf("%w: pets without pinned artifact", ErrInvalidEngineState)
			}
		} else if pet.ValidateCareStatesForCatalog(state.Pets, bundle.Pets) != nil {
			return fmt.Errorf("%w: invalid pinned pet state", ErrInvalidEngineState)
		}
		if err := validateFounderFiscalState(bundle.Fiscal, state); err != nil {
			return err
		}
	}
	return nil
}

func validateFounderFiscalState(catalog *fiscal.Catalog, state *save.State) error {
	if catalog == nil {
		if state.FiscalCredit != 0 || state.FiscalPeriodOpenedWallMS != 0 || state.FiscalPeriodSequence != 0 ||
			len(state.FiscalGeneratorLevels) != 0 || len(state.FiscalUnlocks) != 0 {
			return fmt.Errorf("%w: fiscal state without pinned artifact", ErrInvalidEngineState)
		}
		return nil
	}
	unlockIDs := make([]string, 0, len(state.FiscalUnlocks))
	for id, unlocked := range state.FiscalUnlocks {
		if !unlocked {
			return fmt.Errorf("%w: invalid fiscal unlock state", ErrInvalidEngineState)
		}
		unlockIDs = append(unlockIDs, id)
	}
	sort.Strings(unlockIDs)
	if err := catalog.ValidateState(fiscal.State{Credit: state.FiscalCredit, PeriodOpenedWallMS: state.FiscalPeriodOpenedWallMS,
		PeriodSequence: state.FiscalPeriodSequence, GeneratorLevels: state.FiscalGeneratorLevels, Unlocks: unlockIDs}); err != nil {
		return fmt.Errorf("%w: invalid pinned fiscal state", ErrInvalidEngineState)
	}
	return nil
}

func validateFounderMinigameState(catalog *minigame.Catalog, state *save.State) error {
	if catalog == nil {
		if len(state.MinigameRatings) != 0 || len(state.MinigameOfflineQuality) != 0 {
			return fmt.Errorf("%w: minigames without pinned artifact", ErrInvalidEngineState)
		}
		return nil
	}
	ids := catalog.MinigameIDs()
	if len(state.MinigameRatings) != len(ids) || len(state.MinigameOfflineQuality) != len(ids) {
		return fmt.Errorf("%w: minigame state key set", ErrInvalidEngineState)
	}
	for _, id := range ids {
		rating, ok := state.MinigameRatings[id]
		quality, qualityOK := state.MinigameOfflineQuality[id]
		if !ok || !qualityOK || !catalog.HasRatingSeason(rating.SeasonMember) || minigame.ValidateOfflineQualityState(minigame.OfflineQualityState{
			GradePPM: quality.GradePPM, LastFounderAttendedMS: quality.LastFounderAttendedMS, DecayRemainderPPM: quality.DecayRemainderPPM,
		}) != nil {
			return fmt.Errorf("%w: invalid pinned minigame state", ErrInvalidEngineState)
		}
	}
	return nil
}

func validateFounderCarryFoundationState(bundle CatalogBundle, state *save.State) error {
	if !bundle.foundationsActive() || state == nil || state.Ledger == nil || state.Ledger.Scope() != economy.ScopeFounder || save.VersionForState(state) != 16 {
		return fmt.Errorf("%w: invalid partial Founder carry", ErrInvalidReplayInputs)
	}
	score, err := bundle.Achievements.Score(state.AchievementsEarnedLifetime)
	if err != nil || score != state.AchievementScoreLifetime {
		return fmt.Errorf("%w: invalid Founder carry achievement score", ErrInvalidReplayInputs)
	}
	return nil
}

func settleAndActivateFoundations(current, next CatalogBundle, founder, company, newCompany *save.State) error {
	if founder == nil || company == nil || newCompany == nil {
		return ErrInvalidEngineState
	}
	currentActive, nextActive := current.foundationsActive(), next.foundationsActive()
	if currentActive {
		currentFounderFloor, currentCompanyFloor := current.versionFloors()
		if !nextActive || save.VersionForState(founder) != currentFounderFloor || save.VersionForState(company) != currentCompanyFloor {
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
		// Founder is persisted under the next run's constants hash. Re-derive the
		// complete lifetime score with that pinned catalog so a balance retune
		// cannot make an otherwise honest Exit fail next-catalog validation.
		score, err := next.Achievements.Score(founder.AchievementsEarnedLifetime)
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
	nextFounderFloor, nextCompanyFloor := next.versionFloors()
	if next.Minigames != nil && current.Minigames == nil {
		founder.MinigameRatings = map[string]save.MinigameRatingState{}
		founder.MinigameOfflineQuality = map[string]save.MinigameOfflineQualityState{}
		// Content rows activate only with a later mint. An empty activation
		// artifact therefore produces complete empty maps without inventing Elo.
	}
	if next.Pets != nil && current.Pets == nil {
		founder.Pets = map[string]pet.CareState{}
	}
	if next.Fiscal != nil && current.Fiscal == nil {
		openedMS := newCompany.RunStartedAt.UnixMilli()
		if openedMS <= 0 || openedMS > decimal.MaxExactInteger {
			return fmt.Errorf("%w: Fiscal activation timestamp", ErrInvalidEngineState)
		}
		founder.FiscalCredit, founder.FiscalPeriodOpenedWallMS, founder.FiscalPeriodSequence = 0, openedMS, 0
		founder.FiscalGeneratorLevels = make(map[string]int64, len(next.Fiscal.GeneratorLevelRows()))
		for _, row := range next.Fiscal.GeneratorLevelRows() {
			founder.FiscalGeneratorLevels[row.GeneratorID] = 0
		}
		founder.FiscalUnlocks = map[string]bool{}
	}
	founder.WireVersion = nextFounderFloor
	newMeters, err := meters.NewRunState(next.Meters, founder.Notoriety)
	if err != nil {
		return err
	}
	newCompany.WireVersion = nextCompanyFloor
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

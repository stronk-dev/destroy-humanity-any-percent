package production

import (
	"fmt"
	"sort"

	"cloud-clicker/server/achievements"
	"cloud-clicker/server/copykeys"
	"cloud-clicker/server/cosmetic"
	"cloud-clicker/server/decimal"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/fiscal"
	"cloud-clicker/server/garden"
	"cloud-clicker/server/meters"
	"cloud-clicker/server/minigame"
	"cloud-clicker/server/pet"
	"cloud-clicker/server/reputation"
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
	if state.Ledger.Scope() == economy.ScopeCompany && companyFloor >= 19 {
		if err := validateAttainmentState(bundle, state); err != nil {
			return err
		}
	}
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
		if err := validateFounderSoulState(bundle, state); err != nil {
			return err
		}
		if bundle.MinigameAPI == nil {
			if state.MinigameSessionSeq != 0 {
				return fmt.Errorf("%w: minigame API state without pinned artifact", ErrInvalidEngineState)
			}
		} else if state.MinigameSessionSeq < 0 || state.MinigameSessionSeq > decimal.MaxExactInteger {
			return fmt.Errorf("%w: invalid minigame API state", ErrInvalidEngineState)
		}
		if err := validateFounderReputationState(bundle.ReputationTree, state); err != nil {
			return err
		}
		if err := validateFounderPetIdentities(bundle.PetSpecies, state); err != nil {
			return err
		}
		if err := validateFounderCosmetics(bundle.Cosmetics, state); err != nil {
			return err
		}
		if err := validateFounderGarden(bundle.Garden, state); err != nil {
			return err
		}
	}
	return nil
}

// validateFounderGarden is Server Garden SG2's pinned half: the object exists
// exactly when server_garden is pinned, and resolves under it.
func validateFounderGarden(catalog *garden.Catalog, state *save.State) error {
	if catalog == nil {
		if state.ServerGarden != nil {
			return fmt.Errorf("%w: server_garden without pinned server_garden artifact", ErrInvalidEngineState)
		}
		return nil
	}
	if err := garden.ValidateAgainst(catalog, state.ServerGarden); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidEngineState, err)
	}
	return nil
}

// validateFounderCosmetics is Cosmetic Shop v1 §3's pinned half: the object
// exists exactly when cosmetics is pinned, and resolves under it.
func validateFounderCosmetics(catalog *cosmetic.Catalog, state *save.State) error {
	if catalog == nil {
		if state.Cosmetics != nil {
			return fmt.Errorf("%w: cosmetics without pinned cosmetics artifact", ErrInvalidEngineState)
		}
		return nil
	}
	pets := make(map[string]struct{}, len(state.Pets))
	for id := range state.Pets {
		pets[id] = struct{}{}
	}
	if err := cosmetic.ValidateShape(state.Cosmetics, pets); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidEngineState, err)
	}
	if err := cosmetic.ValidateAgainst(catalog, state.Cosmetics); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidEngineState, err)
	}
	return nil
}

// validateFounderPetIdentities is Pet Adoption v1 PA3's pinned half: identities
// exist exactly when pet_species is pinned, and resolve under it within the cap.
func validateFounderPetIdentities(species *pet.SpeciesCatalog, state *save.State) error {
	if species == nil {
		if state.PetIdentities != nil {
			return fmt.Errorf("%w: pet identities without pinned pet_species", ErrInvalidEngineState)
		}
		return nil
	}
	if err := pet.ValidateIdentityShape(state.PetIdentities, state.Pets); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidEngineState, err)
	}
	if err := pet.ValidateIdentitiesAgainst(species, state.PetIdentities); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidEngineState, err)
	}
	return nil
}

// validateFounderReputationState is R1's pinned-tree half of the v22
// invariants: available is derivable and reputation_unlock_ppm mirrors the
// owned bonus_unlock nodes of the pinned tree (a checked mirror, never an
// independent authority).
func validateFounderReputationState(tree *reputation.Tree, state *save.State) error {
	if tree == nil {
		if state.ReputationSpent != 0 || state.ReputationNodesOwned != nil || state.ReputationUnlockPPM != 0 {
			return fmt.Errorf("%w: Reputation tree state without pinned artifact", ErrInvalidEngineState)
		}
		return nil
	}
	if _, err := reputation.Available(state.ReputationLevel, state.ReputationSpent); err != nil || state.ReputationNodesOwned == nil {
		return fmt.Errorf("%w: invalid Reputation accounting", ErrInvalidEngineState)
	}
	unlock, err := tree.UnlockPPM(state.ReputationNodesOwned)
	if err != nil || unlock != state.ReputationUnlockPPM {
		return fmt.Errorf("%w: reputation_unlock_ppm does not mirror the owned unlock nodes", ErrInvalidEngineState)
	}
	return nil
}

func validateFounderSoulState(bundle CatalogBundle, state *save.State) error {
	if bundle.Soul == nil {
		if state.Soul != 0 || state.SoulExhaustedSourceIDs != nil {
			return fmt.Errorf("%w: Soul state without pinned artifact", ErrInvalidEngineState)
		}
		return nil
	}
	if state.Soul < bundle.Soul.Policy.Floor || state.Soul > bundle.Soul.Policy.Max || state.SoulExhaustedSourceIDs == nil {
		return fmt.Errorf("%w: Soul value outside pinned policy", ErrInvalidEngineState)
	}
	previous := ""
	for _, id := range state.SoulExhaustedSourceIDs {
		source, ok := bundle.Soul.DebitSource(id)
		if !ok || !source.MayExhaust || id <= previous {
			return fmt.Errorf("%w: invalid Soul exhausted-source set", ErrInvalidEngineState)
		}
		previous = id
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
	// Cosmetic Shop v1 §2: ids are permanent and slots never change across
	// epochs; the artifact cannot disappear once pinned.
	if err := garden.ValidateTransition(current.Garden, next.Garden); err != nil {
		return err
	}
	if err := cosmetic.ValidateTransition(current.Cosmetics, next.Cosmetics); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidEngineState, err)
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
		if err := activateMinigameState(founder, next.Minigames); err != nil {
			return err
		}
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
	if next.Soul != nil && current.Soul == nil {
		founder.Soul = next.Soul.Policy.Initial
		founder.SoulExhaustedSourceIDs = []string{}
	}
	if next.MinigameAPI != nil {
		founder.MinigameSessionSeq = 0
	}
	if next.ReputationTree != nil && current.ReputationTree == nil {
		// R7 v21→v22: Reputation earned before activation stays in
		// reputation_level and is fully spendable; nothing is spent or owned.
		if founder.ReputationUnlockPPM != 0 || founder.ReputationSpent != 0 || founder.ReputationNodesOwned != nil {
			return fmt.Errorf("%w: Reputation tree state before activation", ErrInvalidEngineState)
		}
		founder.ReputationSpent, founder.ReputationNodesOwned = 0, []string{}
	}
	if next.PetSpecies != nil && current.PetSpecies == nil {
		// PA6.2 v22→v23: legal only with no pet; identity is never synthesized.
		if founder.PetIdentities != nil || len(founder.Pets) != 0 {
			return fmt.Errorf("%w: pet state cannot activate pet identities", ErrInvalidEngineState)
		}
		founder.PetIdentities = map[string]pet.Identity{}
	}
	if next.Cosmetics != nil && current.Cosmetics == nil {
		// Cosmetic Shop v1 §6: v23→v24 activates with nothing owned.
		if founder.Cosmetics != nil {
			return fmt.Errorf("%w: cosmetics cannot pre-exist activation", ErrInvalidEngineState)
		}
		founder.Cosmetics = cosmetic.NewState()
	}
	if next.Garden != nil && current.Garden == nil {
		// Server Garden SG2: activation at a new-run boundary or New-Founder
		// initialization initializes an empty garden with every starter.
		if founder.ServerGarden != nil {
			return fmt.Errorf("%w: server_garden cannot pre-exist activation", ErrInvalidEngineState)
		}
		founder.ServerGarden = garden.NewState(next.Garden)
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
	// Clout v1 CV4: attainment is new-run bound and discarded at Exit; it
	// exists only when the next run's pinned economy declares the axis stack.
	newCompany.AchievementsAttainedRun, newCompany.AttainmentScoreRun = nil, 0
	if AxisStackDeclared(next.Economy) {
		newCompany.AchievementsAttainedRun = map[string]bool{}
	}
	if err := next.ValidateFoundationState(founder); err != nil {
		return err
	}
	return next.ValidateFoundationState(newCompany)
}

func activateMinigameState(state *save.State, catalog *minigame.Catalog) error {
	if state == nil || catalog == nil {
		return fmt.Errorf("%w: missing pinned minigame catalog", ErrInvalidEngineState)
	}
	ids := catalog.MinigameIDs()
	state.MinigameRatings = make(map[string]save.MinigameRatingState, len(ids))
	state.MinigameOfflineQuality = make(map[string]save.MinigameOfflineQualityState, len(ids))
	for _, id := range ids {
		definition, ok := catalog.Definition(id)
		if !ok {
			return fmt.Errorf("%w: missing pinned minigame definition", ErrInvalidEngineState)
		}
		state.MinigameRatings[id] = save.MinigameRatingState{
			Elo: definition.Rating.StartingElo, SeasonMember: definition.Rating.SeasonMember,
		}
		state.MinigameOfflineQuality[id] = save.MinigameOfflineQualityState{
			GradePPM: definition.OfflineQuality.NeutralFloorPPM, LastFounderAttendedMS: state.AgeMS,
		}
	}
	return nil
}

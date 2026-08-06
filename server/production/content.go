package production

import (
	"fmt"
	"math/big"
	"sort"

	"cloud-clicker/server/decimal"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/multiplier"
	"cloud-clicker/server/save"
)

var ppmDenominator = decimal.FromFloat64(1_000_000)

func contentContributions(state *save.State, catalog *economy.Catalog) ([]multiplier.Contribution, error) {
	return contentContributionsWithPolicy(state, catalog, nil)
}

func contentContributionsWithPolicy(state *save.State, catalog *economy.Catalog, policy *simulationPolicy) ([]multiplier.Contribution, error) {
	if state == nil || catalog == nil || state.GeneratorCounts == nil {
		return nil, ErrInvalidEngineState
	}
	result := make([]multiplier.Contribution, 0)
	for _, upgrade := range catalog.Upgrades() {
		if !state.UpgradesOwned[upgrade.ID] || policy.masksUpgrade(upgrade.ID) {
			continue
		}
		for _, effect := range upgrade.Effects {
			result = append(result, multiplier.Contribution{Slot: effect.Slot, SourceID: effect.SourceID, Target: effect.Target, Factor: effect.Factor})
		}
	}
	for _, generator := range catalog.GeneratorClassesForScope(economy.ScopeCompany) {
		purchased, exists := state.GeneratorCounts[generator.ID]
		if !exists || purchased < 0 || purchased > decimal.MaxExactInteger {
			return nil, ErrInvalidEngineState
		}
		if policy.masksGenerator(generator.ID) {
			continue
		}
		for _, rung := range generator.Ladder {
			if purchased < rung.PurchasedAt {
				break
			}
			factor := decimal.FromFloat64(float64(rung.MultiplierPPM)).Div(ppmDenominator).Quantize(decimal.CanonicalSignificantDigits)
			if !factor.IsStateValue() || !factor.Gt(decimal.Zero) {
				return nil, ErrInvalidEngineState
			}
			result = append(result, multiplier.Contribution{Slot: multiplier.SlotMilestones, SourceID: economy.LadderSourceID(generator.ID, rung.PurchasedAt), Target: generator.ID, Factor: factor})
		}
		for _, role := range generator.Roles {
			if role.Kind != economy.RoleManualOutput || purchased == 0 {
				continue
			}
			factor, err := countPPMFactor(purchased, role.PerPurchasedPPM)
			if err != nil {
				return nil, err
			}
			result = append(result, multiplier.Contribution{Slot: multiplier.SlotUpgrades, SourceID: economy.ManualRoleSourceID(generator.ID, role.ActionID), Target: role.ActionID, Factor: factor})
		}
	}
	for _, pool := range catalog.SynergyPools() {
		total := new(big.Int)
		for _, source := range pool.Sources {
			var count int64
			switch source.Kind {
			case economy.SynergyGenerator:
				if policy.masksGenerator(source.ID) {
					continue
				}
				var exists bool
				count, exists = state.GeneratorCounts[source.ID]
				if !exists {
					return nil, ErrInvalidEngineState
				}
				generator, declared := catalog.GeneratorClass(source.ID)
				if count > 0 && declared && generatorDeclaresRole(generator, economy.RoleSynergyFeed, pool.ID) {
					policy.candidate(RoleActivation{GeneratorID: source.ID, Kind: economy.RoleSynergyFeed, TargetID: pool.ID})
				}
			case economy.SynergyUpgrade:
				if state.UpgradesOwned[source.ID] && !policy.masksUpgrade(source.ID) {
					count = 1
				}
			default:
				return nil, ErrInvalidEngineState
			}
			term := new(big.Int).Mul(big.NewInt(count), big.NewInt(source.PerCountPPM))
			total.Add(total, term)
		}
		if total.Sign() == 0 {
			continue
		}
		factor, err := synergyFactor(pool.Curve, total)
		if err != nil {
			return nil, err
		}
		declaration, exists := catalog.MultiplierSource(pool.ID)
		if !exists {
			return nil, ErrInvalidEngineState
		}
		result = append(result, multiplier.Contribution{Slot: pool.Slot, SourceID: pool.ID, Target: declaration.Target, Factor: factor})
	}
	sort.Slice(result, func(left, right int) bool { return contributionKey(result[left]) < contributionKey(result[right]) })
	return result, nil
}

func assembleContributions(state *save.State, catalog *economy.Catalog, external []multiplier.Contribution) ([]multiplier.Contribution, error) {
	return assembleContributionsWithPolicy(state, catalog, external, nil)
}

func assembleContributionsWithPolicy(state *save.State, catalog *economy.Catalog, external []multiplier.Contribution, policy *simulationPolicy) ([]multiplier.Contribution, error) {
	derived, err := contentContributionsWithPolicy(state, catalog, policy)
	if err != nil {
		return nil, err
	}
	combined := append(append(make([]multiplier.Contribution, 0, len(external)+len(derived)), external...), derived...)
	sort.Slice(combined, func(left, right int) bool { return contributionKey(combined[left]) < contributionKey(combined[right]) })
	if _, err := validateContributions(catalog, combined); err != nil {
		return nil, err
	}
	return combined, nil
}

func countPPMFactor(count, perPurchasedPPM int64) (decimal.Decimal, error) {
	if count < 0 || count > decimal.MaxExactInteger || perPurchasedPPM <= 0 || perPurchasedPPM > decimal.MaxExactInteger {
		return decimal.NaN, ErrInvalidEngineState
	}
	value := new(big.Int).Mul(big.NewInt(count), big.NewInt(perPurchasedPPM))
	factor := decimal.One.Add(decimal.FromString(value.String()).Div(ppmDenominator)).Quantize(decimal.CanonicalSignificantDigits)
	if !factor.IsStateValue() || !factor.Gt(decimal.Zero) {
		return decimal.NaN, ErrInvalidEngineState
	}
	return factor, nil
}

func synergyFactor(curve economy.SynergyCurve, totalPPM *big.Int) (decimal.Decimal, error) {
	if totalPPM == nil || totalPPM.Sign() < 0 {
		return decimal.NaN, ErrInvalidEngineState
	}
	base := decimal.One.Add(decimal.FromString(totalPPM.String()).Div(ppmDenominator))
	var factor decimal.Decimal
	switch curve {
	case economy.SynergyLinear:
		factor = base
	case economy.SynergyLog:
		factor = decimal.One.Add(base.Log10())
	default:
		return decimal.NaN, ErrInvalidEngineState
	}
	factor = factor.Quantize(decimal.CanonicalSignificantDigits)
	if !factor.IsStateValue() || !factor.Gt(decimal.Zero) {
		return decimal.NaN, fmt.Errorf("%w: invalid synergy factor", ErrInvalidEngineState)
	}
	return factor, nil
}

func contributionKey(value multiplier.Contribution) string {
	return string(value.Slot) + "\x00" + value.SourceID + "\x00" + value.Target
}

func contributionFactorForTarget(catalog *economy.Catalog, target string, contributions []multiplier.Contribution) (decimal.Decimal, error) {
	bySource, err := validateContributions(catalog, contributions)
	if err != nil {
		return decimal.NaN, err
	}
	factor := decimal.One
	for _, slot := range multiplier.Order {
		sources := make([]string, 0)
		for identity, contribution := range bySource {
			if contribution.Slot == slot && contribution.Target == target {
				sources = append(sources, identity)
			}
		}
		for _, identity := range multiplier.OrderedSourceIDs(sources) {
			factor = factor.Mul(bySource[identity].Factor)
		}
	}
	factor = factor.Quantize(decimal.CanonicalSignificantDigits)
	if !factor.IsStateValue() || !factor.Gt(decimal.Zero) {
		return decimal.NaN, ErrInvalidEngineState
	}
	return factor, nil
}

func materializeProvisionBoundary(catalog *economy.Catalog, purchased, provisioned, remainders map[string]int64) error {
	return materializeProvisionBoundaryWithPolicy(catalog, purchased, provisioned, remainders, nil)
}

func materializeProvisionBoundaryWithPolicy(catalog *economy.Catalog, purchased, provisioned, remainders map[string]int64, policy *simulationPolicy) error {
	if catalog == nil || purchased == nil || provisioned == nil || remainders == nil {
		return ErrInvalidEngineState
	}
	staged := make(map[string]int64)
	stagedRemainders := make(map[string]int64)
	denominator := big.NewInt(1_000_000)
	for _, source := range catalog.GeneratorClassesForScope(economy.ScopeCompany) {
		if source.Provision == nil || policy.masksGenerator(source.ID) {
			continue
		}
		target, exists := catalog.GeneratorClass(source.Provision.GeneratorID)
		sourcePurchased, purchasedExists := purchased[source.ID]
		sourceProvisioned, provisionedExists := provisioned[source.ID]
		priorRemainder, remainderExists := remainders[source.Provision.GeneratorID]
		currentTarget, targetExists := provisioned[source.Provision.GeneratorID]
		if !exists || target.ProvisionedHardcap == nil || !purchasedExists || !provisionedExists || !remainderExists || !targetExists ||
			sourcePurchased < 0 || sourceProvisioned < 0 || priorRemainder < 0 || priorRemainder >= 1_000_000 || currentTarget < 0 {
			return ErrInvalidEngineState
		}
		sourceTotal := new(big.Int).Add(big.NewInt(sourcePurchased), big.NewInt(sourceProvisioned))
		numerator := new(big.Int).Mul(sourceTotal, big.NewInt(source.Provision.RatePPM))
		numerator.Add(numerator, big.NewInt(priorRemainder))
		quotient, remainder := new(big.Int), new(big.Int)
		quotient.QuoRem(numerator, denominator, remainder)
		stagedRemainders[source.Provision.GeneratorID] = remainder.Int64()
		headroom := target.ProvisionedHardcap.Count - currentTarget
		if headroom < 0 {
			return ErrInvalidEngineState
		}
		if quotient.Cmp(big.NewInt(headroom)) > 0 {
			staged[source.Provision.GeneratorID] = headroom
		} else if !quotient.IsInt64() {
			return ErrInvalidEngineState
		} else {
			staged[source.Provision.GeneratorID] = quotient.Int64()
		}
		if staged[source.Provision.GeneratorID] > 0 && generatorDeclaresRole(source, economy.RoleProvision, source.Provision.GeneratorID) {
			policy.activate(RoleActivation{GeneratorID: source.ID, Kind: economy.RoleProvision, TargetID: source.Provision.GeneratorID})
		}
	}
	for targetID, remainder := range stagedRemainders {
		remainders[targetID] = remainder
		provisioned[targetID] += staged[targetID]
	}
	return nil
}

func generatorDeclaresRole(generator economy.GeneratorClassDefinition, kind economy.GeneratorRoleKind, targetID string) bool {
	for _, role := range generator.Roles {
		if role.Kind != kind {
			continue
		}
		switch kind {
		case economy.RoleProvision:
			return role.GeneratorID == targetID
		case economy.RoleSynergyFeed:
			return role.PoolID == targetID
		case economy.RoleManualOutput:
			return role.ActionID == targetID
		case economy.RoleStockRate:
			return targetID == "faction.stock"
		}
	}
	return false
}

func cloneInt64Counts(source map[string]int64) map[string]int64 {
	if source == nil {
		return nil
	}
	result := make(map[string]int64, len(source))
	for id, value := range source {
		result[id] = value
	}
	return result
}

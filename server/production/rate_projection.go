package production

import (
	"sort"
	"time"

	"cloud-clicker/server/decimal"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/multiplier"
	prestigecore "cloud-clicker/server/prestige"
	"cloud-clicker/server/save"
)

// GeneratorRate is one byte-sorted generator contribution in a read-only rate
// projection. Rate is the generator's complete contribution after the same
// contribution assembly used by live and replay evaluation.
type GeneratorRate struct {
	GeneratorID string
	ResourceID  string
	Rate        decimal.Decimal
}

// ResourceRate is one byte-sorted aggregate resource rate in a read-only rate
// projection. Resources with no current production are represented by zero.
type ResourceRate struct {
	ResourceID string
	Rate       decimal.Decimal
}

// RateProjection is the presentation-only view of canonical production rates.
// It contains no state cursor and cannot be committed as transition evidence.
type RateProjection struct {
	Generators []GeneratorRate
	Resources  []ResourceRate
}

// ResolveRateProjectionAttendedMS resolves the attended coordinate used by an
// active-play rate projection without mutating replay-owned state. Long gaps
// are classified with the pinned prestige policy exactly as they are at the
// transition boundary.
func ResolveRateProjectionAttendedMS(bundle CatalogBundle, state *save.State, now time.Time) (int64, error) {
	if state == nil || state.Ledger == nil || state.Ledger.Scope() != economy.ScopeCompany || now.IsZero() {
		return 0, ErrInvalidEngineState
	}
	if save.VersionForState(state) < 18 {
		return 0, nil
	}
	if bundle.Opportunities == nil || bundle.Prestige == nil {
		return 0, ErrInvalidEngineState
	}
	clone := *state
	clone.OfflineSpans = append([]save.OfflineSpan(nil), state.OfflineSpans...)
	effectiveNow := save.CanonicalServerTime(now)
	if effectiveNow.Before(clone.EvaluatedThrough) {
		return 0, ErrInvalidEngineState
	}
	if effectiveNow.After(clone.EvaluatedThrough) {
		if err := prestigecore.RecordOfflineSpan(&clone, clone.EvaluatedThrough, effectiveNow, bundle.Prestige.CatchupCeilingMS); err != nil {
			return 0, err
		}
	}
	attended, err := prestigecore.AttendedMS(&clone, effectiveNow)
	if err != nil || attended < 0 {
		return 0, ErrInvalidEngineState
	}
	return attended, nil
}

// ProjectRates returns the canonical per-generator and per-resource production
// rates for a pinned Company state. It is pure: it does not advance a cursor,
// mutate the save, emit events, construct receipts, or create replay inputs.
func ProjectRates(bundle CatalogBundle, state *save.State, external []multiplier.Contribution, attendedMS int64) (RateProjection, error) {
	if state == nil || state.Ledger == nil || state.Ledger.Scope() != economy.ScopeCompany || bundle.Economy == nil || attendedMS < 0 {
		return RateProjection{}, ErrInvalidEngineState
	}
	if (save.VersionForState(state) >= 18) != (bundle.Opportunities != nil) {
		return RateProjection{}, ErrInvalidEngineState
	}
	purchased, provisioned, err := projectionGeneratorCounts(bundle.Economy, state)
	if err != nil {
		return RateProjection{}, err
	}
	active := []multiplier.Contribution{}
	if save.VersionForState(state) >= 18 {
		active, err = activePlayContributions(state, bundle.Opportunities, attendedMS)
		if err != nil {
			return RateProjection{}, err
		}
	}
	inputs := append(append([]multiplier.Contribution(nil), external...), active...)
	contributions, err := assembleContributions(state, bundle.Economy, inputs)
	if err != nil {
		return RateProjection{}, err
	}
	rates, err := ratesWithProvisionedAndPolicy(bundle.Economy, purchased, provisioned, contributions, nil)
	if err != nil {
		return RateProjection{}, err
	}
	result := RateProjection{Generators: []GeneratorRate{}, Resources: []ResourceRate{}}
	for _, resource := range bundle.Economy.Resources() {
		if resource.Scope != economy.ScopeCompany {
			continue
		}
		result.Resources = append(result.Resources, ResourceRate{ResourceID: resource.ID, Rate: decimal.SumDeterministic(rates[resource.ID])})
	}
	definitions := bundle.Economy.GeneratorClassesForScope(economy.ScopeCompany)
	for _, generator := range definitions {
		isolatedPurchased := make(map[string]int64, len(definitions))
		isolatedProvisioned := make(map[string]int64, len(definitions))
		for _, row := range definitions {
			isolatedPurchased[row.ID], isolatedProvisioned[row.ID] = 0, 0
		}
		isolatedPurchased[generator.ID] = purchased[generator.ID]
		isolatedProvisioned[generator.ID] = provisioned[generator.ID]
		isolated, isolatedErr := ratesWithProvisionedAndPolicy(bundle.Economy, isolatedPurchased, isolatedProvisioned, contributions, nil)
		if isolatedErr != nil {
			return RateProjection{}, isolatedErr
		}
		result.Generators = append(result.Generators, GeneratorRate{GeneratorID: generator.ID,
			ResourceID: generator.Production.ResourceID, Rate: decimal.SumDeterministic(isolated[generator.Production.ResourceID])})
	}
	sort.Slice(result.Generators, func(left, right int) bool {
		return result.Generators[left].GeneratorID < result.Generators[right].GeneratorID
	})
	sort.Slice(result.Resources, func(left, right int) bool {
		return result.Resources[left].ResourceID < result.Resources[right].ResourceID
	})
	return result, nil
}

func projectionGeneratorCounts(catalog *economy.Catalog, state *save.State) (map[string]int64, map[string]int64, error) {
	purchased, provisioned := make(map[string]int64), make(map[string]int64)
	definitions := catalog.GeneratorClassesForScope(economy.ScopeCompany)
	if len(state.GeneratorCounts) != len(definitions) || state.GeneratorProvisioned != nil && len(state.GeneratorProvisioned) != len(definitions) {
		return nil, nil, ErrInvalidEngineState
	}
	for _, generator := range definitions {
		owned, ok := state.GeneratorCounts[generator.ID]
		generated := int64(0)
		generatedOK := state.GeneratorProvisioned == nil
		if state.GeneratorProvisioned != nil {
			generated, generatedOK = state.GeneratorProvisioned[generator.ID]
		}
		if !ok || !generatedOK || owned < 0 || generated < 0 || owned > decimal.MaxExactInteger-generated {
			return nil, nil, ErrInvalidEngineState
		}
		purchased[generator.ID], provisioned[generator.ID] = owned, generated
	}
	return purchased, provisioned, nil
}

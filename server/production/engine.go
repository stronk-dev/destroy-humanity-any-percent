package production

import (
	"errors"
	"fmt"
	"math/big"
	"sort"
	"time"

	"cloud-clicker/server/decimal"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/multiplier"
	"cloud-clicker/server/save"
)

var ErrInvalidEngineState = errors.New("invalid production engine state")

type EvaluationMode string

const (
	ModeOnline  EvaluationMode = "online"
	ModeOffline EvaluationMode = "offline"
)

type EvaluationResult struct {
	Receipt        economy.Receipt
	ElapsedMS      int64
	ProductionMS   int64
	BankedCreditMS int64
}

func Evaluate(
	state *save.State,
	catalog *economy.Catalog,
	now time.Time,
	mode EvaluationMode,
	contributions []multiplier.Contribution,
) (EvaluationResult, error) {
	if state == nil || state.Ledger == nil || catalog == nil || state.Ledger.Scope() != economy.ScopeCompany ||
		(mode != ModeOnline && mode != ModeOffline) {
		return EvaluationResult{}, ErrInvalidEngineState
	}
	if !now.After(state.EvaluatedThrough) {
		return EvaluationResult{}, nil
	}
	elapsedMS := now.Sub(state.EvaluatedThrough).Milliseconds()
	if elapsedMS <= 0 {
		return EvaluationResult{}, nil
	}
	if elapsedMS > decimal.MaxExactInteger {
		return EvaluationResult{}, fmt.Errorf("%w: elapsed time exceeds exact integer domain", ErrInvalidEngineState)
	}

	productionMS := elapsedMS
	efficiency := decimal.One
	banked := int64(0)
	if mode == ModeOffline {
		policy := catalog.OfflinePolicy()
		if policy.AccrualCapMS <= 0 {
			return EvaluationResult{}, fmt.Errorf("%w: catalog has no offline policy", ErrInvalidEngineState)
		}
		efficiency = policy.Efficiency
		if productionMS > policy.AccrualCapMS {
			productionMS = policy.AccrualCapMS
		}
		excess := elapsedMS - productionMS
		banked = ratioFloor(excess, policy.BankRatioNumerator, policy.BankRatioDenominator)
		remaining := policy.BankCapMS - state.ComputeCreditMS
		if remaining < 0 {
			return EvaluationResult{}, fmt.Errorf("%w: compute credits exceed catalog cap", ErrInvalidEngineState)
		}
		if banked > remaining {
			banked = remaining
		}
	}

	rates, err := Rates(catalog, state.GeneratorCounts, contributions)
	if err != nil {
		return EvaluationResult{}, err
	}
	resourceIDs := make([]string, 0, len(rates))
	for resourceID := range rates {
		resourceIDs = append(resourceIDs, resourceID)
	}
	sort.Strings(resourceIDs)
	entries := make([]economy.Entry, 0, len(resourceIDs))
	for _, resourceID := range resourceIDs {
		delta, err := AccrueConstant(rates[resourceID], productionMS, efficiency)
		if err != nil {
			return EvaluationResult{}, fmt.Errorf("%w: accrue %s: %v", ErrInvalidEngineState, resourceID, err)
		}
		balance, exists := state.Ledger.Balance(resourceID)
		definition, defined := catalog.Resource(resourceID)
		if !exists || !defined {
			return EvaluationResult{}, fmt.Errorf("%w: missing production resource %q", ErrInvalidEngineState, resourceID)
		}
		if definition.Hardcap != nil {
			headroom := definition.Hardcap.Amount.Sub(balance).Quantize(decimal.CanonicalSignificantDigits)
			if headroom.Lt(decimal.Zero) {
				return EvaluationResult{}, fmt.Errorf("%w: resource %q exceeds hardcap", ErrInvalidEngineState, resourceID)
			}
			if delta.Gt(headroom) {
				delta = headroom
			}
		}
		if !delta.Eq(decimal.Zero) {
			entries = append(entries, economy.Entry{ResourceID: resourceID, Delta: delta})
		}
	}
	receipt, err := state.Ledger.Apply(economy.Transaction{Entries: entries})
	if err != nil {
		return EvaluationResult{}, fmt.Errorf("%w: ledger commit: %v", ErrInvalidEngineState, err)
	}
	state.ComputeCreditMS += banked
	state.EvaluatedThrough = state.EvaluatedThrough.Add(time.Duration(elapsedMS) * time.Millisecond)
	return EvaluationResult{
		Receipt: receipt, ElapsedMS: elapsedMS, ProductionMS: productionMS, BankedCreditMS: banked,
	}, nil
}

func Rates(
	catalog *economy.Catalog,
	counts map[string]int64,
	contributions []multiplier.Contribution,
) (map[string][]decimal.Decimal, error) {
	if catalog == nil || counts == nil {
		return nil, ErrInvalidEngineState
	}
	bySource := make(map[string]multiplier.Contribution, len(contributions))
	for _, contribution := range contributions {
		declaration, exists := catalog.MultiplierSource(contribution.SourceID)
		if !exists || declaration.Slot != contribution.Slot || declaration.Target != contribution.Target ||
			!contribution.Factor.IsStateValue() || !contribution.Factor.Gt(decimal.Zero) {
			return nil, fmt.Errorf("%w: invalid multiplier contribution %q", ErrInvalidEngineState, contribution.SourceID)
		}
		if _, duplicate := bySource[contribution.SourceID]; duplicate {
			return nil, fmt.Errorf("%w: duplicate multiplier contribution %q", ErrInvalidEngineState, contribution.SourceID)
		}
		bySource[contribution.SourceID] = contribution
	}

	generators := catalog.GeneratorClassesForScope(economy.ScopeCompany)
	if len(counts) != len(generators) {
		return nil, fmt.Errorf("%w: generator count set does not match catalog", ErrInvalidEngineState)
	}
	rates := make(map[string][]decimal.Decimal)
	for _, generator := range generators {
		count, exists := counts[generator.ID]
		if !exists || count < 0 || count > decimal.MaxExactInteger || generator.Production == nil {
			return nil, fmt.Errorf("%w: invalid generator count %q", ErrInvalidEngineState, generator.ID)
		}
		if count == 0 {
			continue
		}
		rate := generator.Production.BaseRate.Mul(decimal.FromFloat64(float64(count)))
		for _, slot := range multiplier.Order {
			sources := make([]string, 0)
			for sourceID, contribution := range bySource {
				if contribution.Slot == slot && (contribution.Target == "all" || contribution.Target == generator.ID) {
					sources = append(sources, sourceID)
				}
			}
			sort.Strings(sources)
			for _, sourceID := range sources {
				rate = rate.Mul(bySource[sourceID].Factor)
			}
		}
		if !rate.IsStateValue() || rate.Lt(decimal.Zero) {
			return nil, fmt.Errorf("%w: non-finite rate for %q", ErrInvalidEngineState, generator.ID)
		}
		rates[generator.Production.ResourceID] = append(rates[generator.Production.ResourceID], rate)
	}
	return rates, nil
}

func ratioFloor(value, numerator, denominator int64) int64 {
	if value <= 0 || numerator <= 0 || denominator <= 0 {
		return 0
	}
	product := new(big.Int).Mul(big.NewInt(value), big.NewInt(numerator))
	product.Quo(product, big.NewInt(denominator))
	if !product.IsInt64() {
		return decimal.MaxExactInteger
	}
	return product.Int64()
}

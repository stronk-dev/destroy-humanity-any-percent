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
	Receipt          economy.Receipt
	ElapsedMS        int64
	ProductionMS     int64
	BankedCreditMS   int64
	ProgressDeltaPPM int64
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
	effectiveNow := save.CanonicalServerTime(now)
	if !effectiveNow.After(state.EvaluatedThrough) {
		return EvaluationResult{}, nil
	}
	elapsedMS := effectiveNow.Sub(state.EvaluatedThrough).Milliseconds()
	if elapsedMS <= 0 {
		return EvaluationResult{}, nil
	}
	if elapsedMS > decimal.MaxExactInteger {
		return EvaluationResult{}, fmt.Errorf("%w: elapsed time exceeds exact integer domain", ErrInvalidEngineState)
	}
	beforeProgressPPM, err := tierProgressPPM(catalog, state)
	if err != nil {
		return EvaluationResult{}, err
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

	entries, provisioned, remainders, err := accrueContent(state, catalog, productionMS, efficiency, contributions)
	if err != nil {
		return EvaluationResult{}, err
	}
	receipt, err := state.Ledger.ApplyAccrual(economy.Transaction{Entries: entries})
	if err != nil {
		return EvaluationResult{}, fmt.Errorf("%w: ledger commit: %v", ErrInvalidEngineState, err)
	}
	state.ComputeCreditMS += banked
	state.GeneratorProvisioned = provisioned
	state.ProvisionRemaindersPPM = remainders
	state.EvaluatedThrough = effectiveNow
	afterProgressPPM, err := tierProgressPPM(catalog, state)
	if err != nil {
		return EvaluationResult{}, err
	}
	progressDeltaPPM := afterProgressPPM - beforeProgressPPM
	if progressDeltaPPM < 0 {
		progressDeltaPPM = 0
	}
	return EvaluationResult{
		Receipt: receipt, ElapsedMS: elapsedMS, ProductionMS: productionMS, BankedCreditMS: banked,
		ProgressDeltaPPM: progressDeltaPPM,
	}, nil
}

func accrueContent(state *save.State, catalog *economy.Catalog, productionMS int64, efficiency decimal.Decimal, contributions []multiplier.Contribution) ([]economy.Entry, map[string]int64, map[string]int64, error) {
	provisioned := cloneInt64Counts(state.GeneratorProvisioned)
	remainders := cloneInt64Counts(state.ProvisionRemaindersPPM)
	if provisioned == nil {
		provisioned = make(map[string]int64, len(state.GeneratorCounts))
		for id := range state.GeneratorCounts {
			provisioned[id] = 0
		}
	}
	if remainders == nil {
		remainders = map[string]int64{}
	}
	deltas := map[string][]decimal.Decimal{}
	accrueSegment := func(segmentMS int64) error {
		if segmentMS <= 0 {
			return nil
		}
		rates, err := ratesWithProvisioned(catalog, state.GeneratorCounts, provisioned, contributions)
		if err != nil {
			return err
		}
		for resourceID, values := range rates {
			delta, err := AccrueConstant(values, segmentMS, efficiency)
			if err != nil {
				return fmt.Errorf("%w: accrue %s: %v", ErrInvalidEngineState, resourceID, err)
			}
			if !delta.Eq(decimal.Zero) {
				deltas[resourceID] = append(deltas[resourceID], delta)
			}
		}
		return nil
	}
	tickMS := catalog.ProvisionTickMS()
	if tickMS == 0 || productionMS == 0 {
		if err := accrueSegment(productionMS); err != nil {
			return nil, nil, nil, err
		}
	} else {
		if state.RunStartedAt.IsZero() || state.EvaluatedThrough.Before(state.RunStartedAt) || productionMS/tickMS > catalog.OfflinePolicy().AccrualCapMS/tickMS+1 {
			return nil, nil, nil, fmt.Errorf("%w: invalid provision bucket horizon", ErrInvalidEngineState)
		}
		cursorMS := state.EvaluatedThrough.UnixMilli()
		endMS := cursorMS + productionMS
		runStartMS := state.RunStartedAt.UnixMilli()
		nextBoundary := runStartMS + ((cursorMS-runStartMS)/tickMS+1)*tickMS
		for cursorMS < endMS {
			segmentEnd := endMS
			if nextBoundary < segmentEnd {
				segmentEnd = nextBoundary
			}
			if err := accrueSegment(segmentEnd - cursorMS); err != nil {
				return nil, nil, nil, err
			}
			cursorMS = segmentEnd
			if cursorMS == nextBoundary {
				if err := materializeProvisionBoundary(catalog, state.GeneratorCounts, provisioned, remainders); err != nil {
					return nil, nil, nil, err
				}
				nextBoundary += tickMS
			}
		}
	}
	resourceIDs := make([]string, 0, len(deltas))
	for resourceID := range deltas {
		resourceIDs = append(resourceIDs, resourceID)
	}
	sort.Strings(resourceIDs)
	entries := make([]economy.Entry, 0, len(resourceIDs))
	for _, resourceID := range resourceIDs {
		delta := decimal.SumDeterministic(deltas[resourceID]).Quantize(decimal.CanonicalSignificantDigits)
		if !delta.Eq(decimal.Zero) {
			entries = append(entries, economy.Entry{ResourceID: resourceID, Delta: delta})
		}
	}
	return entries, provisioned, remainders, nil
}

func tierProgressPPM(catalog *economy.Catalog, state *save.State) (int64, error) {
	if _, ok := catalog.ProgressCoordinate(int(state.Tier)); !ok {
		return 0, nil
	}
	progress, err := SubProgressValue(catalog, state, int(state.Tier))
	if err != nil {
		return 0, err
	}
	value, ok := progress.Mul(decimal.FromFloat64(1_000_000)).Floor().Int64Exact()
	if !ok || value < 0 || value > 1_000_000 {
		return 0, ErrInvalidEngineState
	}
	return value, nil
}

func Rates(
	catalog *economy.Catalog,
	counts map[string]int64,
	contributions []multiplier.Contribution,
) (map[string][]decimal.Decimal, error) {
	return ratesWithProvisioned(catalog, counts, nil, contributions)
}

func validateContributions(catalog *economy.Catalog, contributions []multiplier.Contribution) (map[string]multiplier.Contribution, error) {
	if catalog == nil {
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
	return bySource, nil
}

func ratesWithProvisioned(catalog *economy.Catalog, counts, provisioned map[string]int64, contributions []multiplier.Contribution) (map[string][]decimal.Decimal, error) {
	if catalog == nil || counts == nil {
		return nil, ErrInvalidEngineState
	}
	bySource, err := validateContributions(catalog, contributions)
	if err != nil {
		return nil, err
	}

	generators := catalog.GeneratorClassesForScope(economy.ScopeCompany)
	if len(counts) != len(generators) || provisioned != nil && len(provisioned) != len(generators) {
		return nil, fmt.Errorf("%w: generator count set does not match catalog", ErrInvalidEngineState)
	}
	rates := make(map[string][]decimal.Decimal)
	for _, generator := range generators {
		count, exists := counts[generator.ID]
		if !exists || count < 0 || count > decimal.MaxExactInteger || generator.Production == nil {
			return nil, fmt.Errorf("%w: invalid generator count %q", ErrInvalidEngineState, generator.ID)
		}
		generated := int64(0)
		if provisioned != nil {
			var generatedExists bool
			generated, generatedExists = provisioned[generator.ID]
			if !generatedExists || generated < 0 || generated > decimal.MaxExactInteger {
				return nil, fmt.Errorf("%w: invalid provisioned generator count %q", ErrInvalidEngineState, generator.ID)
			}
		}
		if count == 0 && generated == 0 {
			continue
		}
		total := new(big.Int).Add(big.NewInt(count), big.NewInt(generated))
		rate := generator.Production.BaseRate.Mul(decimal.FromString(total.String()))
		for _, slot := range multiplier.Order {
			sources := make([]string, 0)
			for sourceID, contribution := range bySource {
				if contribution.Slot == slot && (contribution.Target == "all" || contribution.Target == generator.ID) {
					sources = append(sources, sourceID)
				}
			}
			for _, sourceID := range multiplier.OrderedSourceIDs(sources) {
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

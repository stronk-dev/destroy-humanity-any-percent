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
	return evaluateWithSimulationPolicy(state, catalog, now, mode, contributions, nil)
}

func evaluateWithSimulationPolicy(state *save.State, catalog *economy.Catalog, now time.Time, mode EvaluationMode, contributions []multiplier.Contribution, policy *simulationPolicy) (EvaluationResult, error) {
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
	if state.ComputeBurstRemainingMS < 0 || state.ComputeBurstRemainingMS > 0 && save.VersionForState(state) < 17 {
		return EvaluationResult{}, fmt.Errorf("%w: invalid compute burst state", ErrInvalidEngineState)
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

	boostedWallMS := state.ComputeBurstRemainingMS
	if boostedWallMS > elapsedMS {
		boostedWallMS = elapsedMS
	}
	boostedProductionMS := boostedWallMS
	if boostedProductionMS > productionMS {
		boostedProductionMS = productionMS
	}
	bonusFactor := decimal.Zero
	if boostedProductionMS > 0 {
		bonusFactor = catalog.OfflinePolicy().BurstSpeed.Sub(decimal.One)
		if !bonusFactor.IsStateValue() || !bonusFactor.Gt(decimal.Zero) {
			return EvaluationResult{}, fmt.Errorf("%w: invalid compute burst speed", ErrInvalidEngineState)
		}
	}

	entries, provisioned, remainders, err := accrueContent(state, catalog, productionMS, boostedProductionMS, efficiency, bonusFactor, contributions, policy)
	if err != nil {
		return EvaluationResult{}, err
	}
	receipt, err := state.Ledger.ApplyAccrual(economy.Transaction{Entries: entries})
	if err != nil {
		return EvaluationResult{}, fmt.Errorf("%w: ledger commit: %v", ErrInvalidEngineState, err)
	}
	state.ComputeCreditMS += banked
	state.ComputeBurstRemainingMS -= boostedWallMS
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

func accrueContent(state *save.State, catalog *economy.Catalog, productionMS, boostedProductionMS int64, efficiency, bonusFactor decimal.Decimal, contributions []multiplier.Contribution, policy *simulationPolicy) ([]economy.Entry, map[string]int64, map[string]int64, error) {
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
	remainingBoostedMS := boostedProductionMS
	accrueSegment := func(segmentMS int64) error {
		if segmentMS <= 0 {
			return nil
		}
		policy.setProductionObservation(efficiency.Gt(decimal.Zero))
		rates, err := ratesWithProvisionedAndPolicy(catalog, state.GeneratorCounts, provisioned, contributions, policy)
		policy.setProductionObservation(false)
		if err != nil {
			return err
		}
		for resourceID, values := range rates {
			bonusMS := segmentMS
			if bonusMS > remainingBoostedMS {
				bonusMS = remainingBoostedMS
			}
			delta, err := accrueBoostedConstant(values, segmentMS, bonusMS, efficiency, bonusFactor)
			if err != nil {
				return fmt.Errorf("%w: accrue %s: %v", ErrInvalidEngineState, resourceID, err)
			}
			if !delta.Eq(decimal.Zero) {
				deltas[resourceID] = append(deltas[resourceID], delta)
			}
		}
		if remainingBoostedMS > segmentMS {
			remainingBoostedMS -= segmentMS
		} else {
			remainingBoostedMS = 0
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
				if err := materializeProvisionBoundaryWithPolicy(catalog, state.GeneratorCounts, provisioned, remainders, policy); err != nil {
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

// accrueBoostedConstant integrates the ordinary and burst portions of one
// fixed-grid segment before the single explicit state quantization boundary.
func accrueBoostedConstant(rates []decimal.Decimal, elapsedMS, boostedMS int64, efficiency, bonusFactor decimal.Decimal) (decimal.Decimal, error) {
	if elapsedMS < 0 || boostedMS < 0 || boostedMS > elapsedMS ||
		!efficiency.IsStateValue() || efficiency.Lt(decimal.Zero) ||
		!bonusFactor.IsStateValue() || bonusFactor.Lt(decimal.Zero) {
		return decimal.NaN, ErrInvalidAccrual
	}
	for _, rate := range rates {
		if !rate.IsStateValue() || rate.Lt(decimal.Zero) {
			return decimal.NaN, ErrInvalidAccrual
		}
	}
	totalRate := decimal.SumDeterministic(rates)
	if !totalRate.IsStateValue() {
		return decimal.NaN, ErrInvalidAccrual
	}
	baseSeconds := decimal.FromFloat64(float64(elapsedMS)).Div(decimal.FromFloat64(1000))
	bonusSeconds := decimal.FromFloat64(float64(boostedMS)).Div(decimal.FromFloat64(1000))
	delta := totalRate.Mul(baseSeconds.Add(bonusSeconds.Mul(bonusFactor))).Mul(efficiency).Quantize(decimal.CanonicalSignificantDigits)
	if !delta.IsStateValue() {
		return decimal.NaN, ErrInvalidAccrual
	}
	return delta, nil
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
	return ratesWithProvisionedAndPolicy(catalog, counts, nil, contributions, nil)
}

func validateContributions(catalog *economy.Catalog, contributions []multiplier.Contribution) (map[string]multiplier.Contribution, error) {
	if catalog == nil {
		return nil, ErrInvalidEngineState
	}
	bySource := make(map[string]multiplier.Contribution, len(contributions))
	for _, contribution := range contributions {
		declarationID := activeDeclarationID(contribution.SourceID, contribution.Target)
		declaration, exists := catalog.MultiplierSource(declarationID)
		if !exists && declarationID != contribution.SourceID {
			declaration, exists = catalog.MultiplierSource(declarationID + "." + contribution.Target)
		}
		if !exists || declaration.Slot != contribution.Slot || declaration.Target != contribution.Target ||
			declarationID != contribution.SourceID && declaration.Provider != "active_play" ||
			!contribution.Factor.IsStateValue() || !contribution.Factor.Gt(decimal.Zero) {
			return nil, fmt.Errorf("%w: invalid multiplier contribution %q", ErrInvalidEngineState, contribution.SourceID)
		}
		identity := contribution.SourceID + "\x00" + contribution.Target
		if _, duplicate := bySource[identity]; duplicate {
			return nil, fmt.Errorf("%w: duplicate multiplier contribution %q", ErrInvalidEngineState, contribution.SourceID)
		}
		bySource[identity] = contribution
	}
	return bySource, nil
}

func ratesWithProvisionedAndPolicy(catalog *economy.Catalog, counts, provisioned map[string]int64, contributions []multiplier.Contribution, policy *simulationPolicy) (map[string][]decimal.Decimal, error) {
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
		if policy.masksGenerator(generator.ID) {
			continue
		}
		total := new(big.Int).Add(big.NewInt(count), big.NewInt(generated))
		rate := generator.Production.BaseRate.Mul(decimal.FromString(total.String()))
		for _, slot := range multiplier.Order {
			sources := make([]string, 0)
			for identity, contribution := range bySource {
				if contribution.Slot == slot && (contribution.Target == "all" || contribution.Target == generator.ID) {
					sources = append(sources, identity)
				}
			}
			for _, identity := range multiplier.OrderedSourceIDs(sources) {
				rate = rate.Mul(bySource[identity].Factor)
				policy.activateSynergySource(bySource[identity].SourceID)
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

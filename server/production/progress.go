package production

import (
	"fmt"

	"cloud-clicker/server/decimal"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/save"
)

func SubProgressValue(catalog *economy.Catalog, state *save.State, tier int) (decimal.Decimal, error) {
	if catalog == nil || state == nil || state.Ledger == nil {
		return decimal.NaN, ErrInvalidEngineState
	}
	definition, exists := catalog.ProgressCoordinate(tier)
	if !exists {
		return decimal.NaN, fmt.Errorf("%w: missing progress coordinate tier %d", ErrInvalidEngineState, tier)
	}
	return evaluateCoordinate(definition, state)
}

func evaluateCoordinate(definition economy.ProgressCoordinateDefinition, state *save.State) (decimal.Decimal, error) {
	switch definition.Kind {
	case economy.ProgressResourceLog:
		return resourceLog(state, definition.ResourceID, definition.Target)
	case economy.ProgressCountFraction:
		return countFraction(state, definition.CountKey, definition.Required)
	case economy.ProgressComposite:
		values := make([]decimal.Decimal, 0, len(definition.Terms))
		for _, term := range definition.Terms {
			var value decimal.Decimal
			var err error
			switch term.Kind {
			case economy.ProgressResourceLog:
				value, err = resourceLog(state, term.ResourceID, term.Target)
			case economy.ProgressCountFraction:
				value, err = countFraction(state, term.CountKey, term.Required)
			default:
				err = ErrInvalidEngineState
			}
			if err != nil {
				return decimal.NaN, err
			}
			values = append(values, value.Mul(term.Weight))
		}
		return clampProgress(decimal.SumDeterministic(values)), nil
	default:
		return decimal.NaN, ErrInvalidEngineState
	}
}

func resourceLog(state *save.State, resourceID string, target decimal.Decimal) (decimal.Decimal, error) {
	value, exists := state.Ledger.Balance(resourceID)
	if !exists || value.Lt(decimal.Zero) || !target.Gt(decimal.Zero) {
		return decimal.NaN, ErrInvalidEngineState
	}
	coordinate := decimal.One.Add(value).Log10().Div(decimal.One.Add(target).Log10())
	if !coordinate.IsStateValue() {
		return decimal.NaN, ErrInvalidEngineState
	}
	return clampProgress(coordinate), nil
}

func countFraction(state *save.State, countKey string, required int64) (decimal.Decimal, error) {
	if countKey != economy.GeneratorTotalOwned || required <= 0 {
		return decimal.NaN, ErrInvalidEngineState
	}
	total := int64(0)
	for _, count := range state.GeneratorCounts {
		if count < 0 || count > decimal.MaxExactInteger-total {
			return decimal.NaN, ErrInvalidEngineState
		}
		total += count
	}
	coordinate := decimal.FromFloat64(float64(total)).Div(decimal.FromFloat64(float64(required)))
	return clampProgress(coordinate), nil
}

func clampProgress(value decimal.Decimal) decimal.Decimal {
	return value.Max(decimal.Zero).Min(decimal.One).Quantize(decimal.CanonicalSignificantDigits)
}

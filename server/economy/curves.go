package economy

import (
	"errors"
	"fmt"

	"cloud-clicker/server/decimal"
)

var ErrInvalidCurveInput = errors.New("invalid cost-curve input")

type AffordabilityResult struct {
	Count        int64
	UsedFallback bool
}

func (c *Catalog) BulkCost(generatorID string, owned, count int64) (decimal.Decimal, error) {
	definition, exists := c.generatorByID[generatorID]
	if !exists {
		return decimal.NaN, fmt.Errorf("%w: unknown generator class %q", ErrInvalidCurveInput, generatorID)
	}
	return BulkCost(definition.Price, owned, count)
}

func (c *Catalog) MaxAffordable(generatorID string, cash decimal.Decimal, owned int64) (int64, error) {
	result, err := c.MaxAffordableDetailed(generatorID, cash, owned)
	return result.Count, err
}

func (c *Catalog) MaxAffordableDetailed(generatorID string, cash decimal.Decimal, owned int64) (AffordabilityResult, error) {
	definition, exists := c.generatorByID[generatorID]
	if !exists {
		return AffordabilityResult{}, fmt.Errorf("%w: unknown generator class %q", ErrInvalidCurveInput, generatorID)
	}
	return MaxAffordableDetailed(definition.Price, cash, owned)
}

func BulkCost(price PriceDefinition, owned, count int64) (decimal.Decimal, error) {
	if err := validateQuoteInput(price, owned, count); err != nil {
		return decimal.NaN, err
	}
	if count == 0 {
		return decimal.Zero, nil
	}

	countValue := decimal.FromFloat64(float64(count))
	var cost decimal.Decimal
	switch price.Curve.Kind {
	case CurveConstant:
		cost = price.Base.Mul(countValue)
	case CurveLinear:
		ownedValue := decimal.FromFloat64(float64(owned))
		first := price.Base.Add(price.Curve.Step.Mul(ownedValue))
		triangle := countValue.
			Mul(decimal.FromFloat64(float64(count - 1))).
			Div(decimal.FromFloat64(2))
		cost = first.Mul(countValue).Add(price.Curve.Step.Mul(triangle))
	case CurveGeometric:
		cost = decimal.SumGeometricSeries(count, price.Base, price.Curve.Ratio, owned)
	default:
		return decimal.NaN, fmt.Errorf("%w: unsupported curve kind %q", ErrInvalidCurveInput, price.Curve.Kind)
	}
	if !cost.IsStateValue() {
		return decimal.NaN, fmt.Errorf("%w: cost is outside the finite Decimal domain", ErrInvalidCurveInput)
	}
	return cost, nil
}

func MaxAffordable(price PriceDefinition, cash decimal.Decimal, owned int64) (int64, error) {
	result, err := MaxAffordableDetailed(price, cash, owned)
	return result.Count, err
}

func MaxAffordableDetailed(price PriceDefinition, cash decimal.Decimal, owned int64) (AffordabilityResult, error) {
	if !cash.IsStateValue() || cash.Lt(decimal.Zero) || owned < 0 || owned > decimal.MaxExactInteger {
		return AffordabilityResult{}, ErrInvalidCurveInput
	}
	if err := validateQuoteInput(price, owned, 0); err != nil {
		return AffordabilityResult{}, err
	}

	maximum := decimal.MaxExactInteger - owned
	if price.Curve.Kind == CurveGeometric {
		candidate, decimalFallback, err := decimal.AffordGeometricSeriesDetailed(cash, price.Base, price.Curve.Ratio, owned)
		if err != nil {
			return AffordabilityResult{}, fmt.Errorf("%w: geometric affordability: %v", ErrInvalidCurveInput, err)
		}
		if candidate > maximum {
			candidate = maximum
		}
		for correction := 0; correction < 8 && candidate > 0 && !isAffordable(price, cash, owned, candidate); correction++ {
			candidate--
		}
		for correction := 0; correction < 8 && candidate < maximum && isAffordable(price, cash, owned, candidate+1); correction++ {
			candidate++
		}
		if affordabilityPostconditions(price, cash, owned, candidate, maximum) {
			return AffordabilityResult{Count: candidate, UsedFallback: decimalFallback}, nil
		}
		return AffordabilityResult{Count: maxAffordableBySearch(price, cash, owned, maximum), UsedFallback: true}, nil
	}

	return AffordabilityResult{Count: maxAffordableBySearch(price, cash, owned, maximum)}, nil
}

func maxAffordableBySearch(price PriceDefinition, cash decimal.Decimal, owned, maximum int64) int64 {
	high := maximum
	low := int64(0)
	for low < high {
		middle := low + (high-low+1)/2
		if isAffordable(price, cash, owned, middle) {
			low = middle
		} else {
			high = middle - 1
		}
	}
	return low
}

func affordabilityPostconditions(price PriceDefinition, cash decimal.Decimal, owned, count, maximum int64) bool {
	if !isAffordable(price, cash, owned, count) {
		return false
	}
	return count == maximum || !isAffordable(price, cash, owned, count+1)
}

func isAffordable(price PriceDefinition, cash decimal.Decimal, owned, count int64) bool {
	cost, err := BulkCost(price, owned, count)
	return err == nil && cost.Lte(cash)
}

func validateQuoteInput(price PriceDefinition, owned, count int64) error {
	if owned < 0 || count < 0 || owned > decimal.MaxExactInteger || count > decimal.MaxExactInteger-owned ||
		!price.Base.IsStateValue() || !price.Base.Gt(decimal.Zero) {
		return ErrInvalidCurveInput
	}
	switch price.Curve.Kind {
	case CurveConstant:
		return nil
	case CurveLinear:
		if !price.Curve.Step.IsStateValue() || price.Curve.Step.Lt(decimal.Zero) {
			return ErrInvalidCurveInput
		}
	case CurveGeometric:
		if !price.Curve.Ratio.IsStateValue() || price.Curve.Ratio.Lt(decimal.One) {
			return ErrInvalidCurveInput
		}
	default:
		return ErrInvalidCurveInput
	}
	return nil
}

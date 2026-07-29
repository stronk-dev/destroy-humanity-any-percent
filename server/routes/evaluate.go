package routes

import (
	"fmt"

	"cloud-clicker/server/decimal"
)

type Context struct {
	ContextVersion        int
	Resources             map[string]decimal.Decimal
	DoctrinesByTransition map[string]string
	StructureID           string
	LedgerFactKinds       map[string]bool
	MeterBands            map[string]int
	RegionTraits          map[string]bool
}

type Resolution struct {
	GateID      string
	RouteID     string
	Requirement []Requirement
}

func EvaluatePredicate(predicate []Condition, context Context) (bool, error) {
	if context.ContextVersion < 1 || context.Resources == nil || context.DoctrinesByTransition == nil || context.LedgerFactKinds == nil {
		return false, ErrInvalidContext
	}
	for _, condition := range predicate {
		matched := false
		switch condition.Kind {
		case ConditionResourceAtLeast:
			value, ok := context.Resources[condition.ResourceID]
			matched = ok && value.IsStateValue() && value.Gte(condition.Value)
		case ConditionResourceAtMost:
			value, ok := context.Resources[condition.ResourceID]
			matched = ok && value.IsStateValue() && value.Lte(condition.Value)
		case ConditionMeterBand:
			value, ok := context.MeterBands[condition.MeterID]
			matched = ok && value >= condition.Min && value <= condition.Max
		case ConditionDoctrineIs:
			matched = context.DoctrinesByTransition[condition.Transition] == condition.DoctrineID
		case ConditionDoctrineIsNot:
			value, ok := context.DoctrinesByTransition[condition.Transition]
			matched = ok && value != condition.DoctrineID
		case ConditionLedgerFact:
			matched = context.LedgerFactKinds[condition.FactKind]
		case ConditionStructureIs:
			matched = context.StructureID == condition.StructureID
		case ConditionRegionTrait:
			matched = context.RegionTraits[condition.TraitID]
		default:
			return false, fmt.Errorf("%w: unknown condition %q", ErrInvalidContext, condition.Kind)
		}
		if !matched {
			return false, nil
		}
	}
	return true, nil
}

func (c *Catalog) Resolve(gateID, routeID string, context Context) (Resolution, bool, error) {
	gate, exists := c.gateByID[gateID]
	if !exists {
		return Resolution{}, false, nil
	}
	if routeID == "" {
		return Resolution{GateID: gate.ID, Requirement: append([]Requirement(nil), gate.Requirement...)}, true, nil
	}
	for _, route := range gate.Routes {
		if route.RouteID != routeID {
			continue
		}
		if !route.Active || context.ContextVersion < route.RequiresContextVersion {
			return Resolution{}, false, nil
		}
		matched, err := EvaluatePredicate(route.Predicate, context)
		if err != nil || !matched {
			return Resolution{}, false, err
		}
		resolution := Resolution{GateID: gate.ID, RouteID: route.RouteID}
		if route.Effect.Kind == EffectDiscount {
			resolution.Requirement = make([]Requirement, len(gate.Requirement))
			for index, requirement := range gate.Requirement {
				resolution.Requirement[index] = Requirement{
					ResourceID: requirement.ResourceID,
					Amount:     requirement.Amount.Mul(route.Effect.Fraction).Quantize(decimal.CanonicalSignificantDigits),
				}
			}
		}
		return resolution, true, nil
	}
	return Resolution{}, false, nil
}

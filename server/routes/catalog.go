// Package routes owns declarative gate alternatives and their pure predicates.
// It deliberately has no dependency on the production engine.
package routes

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
	"sort"

	"cloud-clicker/server/decimal"
)

const (
	CatalogSchemaVersion  = 1
	CurrentContextVersion = 1
)

var (
	ErrInvalidCatalog = errors.New("invalid routes catalog")
	ErrInvalidContext = errors.New("invalid route predicate context")
	idPattern         = regexp.MustCompile(`^[a-z][a-z0-9_]*(?:\.[a-z][a-z0-9_]*)*$`)
)

type ConditionKind string

const (
	ConditionResourceAtLeast ConditionKind = "resource_at_least"
	ConditionResourceAtMost  ConditionKind = "resource_at_most"
	ConditionMeterBand       ConditionKind = "meter_band"
	ConditionDoctrineIs      ConditionKind = "doctrine_is"
	ConditionDoctrineIsNot   ConditionKind = "doctrine_is_not"
	ConditionLedgerFact      ConditionKind = "ledger_fact_present"
	ConditionStructureIs     ConditionKind = "structure_is"
	ConditionRegionTrait     ConditionKind = "region_trait"
)

type EffectKind string

const (
	EffectDiscount   EffectKind = "discount"
	EffectSubstitute EffectKind = "substitute"
)

type Requirement struct {
	ResourceID string
	Amount     decimal.Decimal
}

type Condition struct {
	Kind        ConditionKind
	ResourceID  string
	Value       decimal.Decimal
	MeterID     string
	Min         int
	Max         int
	Transition  string
	DoctrineID  string
	StructureID string
	FactKind    string
	TraitID     string
}

type Effect struct {
	Kind     EffectKind
	Fraction decimal.Decimal
}

type Alternative struct {
	RouteID                string
	HouseName              string
	Active                 bool
	RequiresContextVersion int
	ExclusionSlot          string
	ExclusionValue         string
	Predicate              []Condition
	Effect                 Effect
}

type Gate struct {
	ID          string
	Requirement []Requirement
	Routes      []Alternative
}

type KnowledgePolicy struct {
	RegistryFirstBonus int64
	FounderFirstGrant  int64
	RepeatGrant        int64
	HintCost           int64
}

type Catalog struct {
	contextVersion            int
	depletionDistinctRequired int
	knowledge                 KnowledgePolicy
	gates                     []Gate
	gateByID                  map[string]Gate
	routeByID                 map[string]Alternative
}

type rawCatalog struct {
	SchemaVersion             int          `json:"schema_version"`
	ContextVersion            int          `json:"context_version"`
	DepletionDistinctRequired int          `json:"depletion_distinct_routes_required"`
	Knowledge                 rawKnowledge `json:"knowledge"`
	Gates                     []rawGate    `json:"gates"`
}

type rawKnowledge struct {
	RegistryFirstBonus int64 `json:"registry_first_bonus"`
	FounderFirstGrant  int64 `json:"founder_first_grant"`
	RepeatGrant        int64 `json:"repeat_grant"`
	HintCost           int64 `json:"hint_cost"`
}

type rawGate struct {
	GateID      string           `json:"gate_id"`
	Requirement []rawRequirement `json:"requirement"`
	Routes      []rawAlternative `json:"routes"`
}

type rawRequirement struct {
	ResourceID string `json:"resource_id"`
	Amount     string `json:"amount"`
}

type rawAlternative struct {
	RouteID                string         `json:"route_id"`
	HouseName              string         `json:"house_name"`
	Active                 bool           `json:"active"`
	RequiresContextVersion int            `json:"requires_context_version"`
	ExclusionSlot          string         `json:"exclusion_slot"`
	ExclusionValue         string         `json:"exclusion_value"`
	Predicate              []rawCondition `json:"predicate"`
	Effect                 rawEffect      `json:"effect"`
}

type rawCondition struct {
	Kind        string  `json:"kind"`
	ResourceID  *string `json:"resource_id"`
	Value       *string `json:"value"`
	MeterID     *string `json:"meter_id"`
	Min         *int    `json:"min"`
	Max         *int    `json:"max"`
	Transition  *string `json:"transition"`
	DoctrineID  *string `json:"doctrine_id"`
	StructureID *string `json:"structure_id"`
	FactKind    *string `json:"fact_kind"`
	TraitID     *string `json:"trait_id"`
}

type rawEffect struct {
	Kind     string  `json:"kind"`
	Fraction *string `json:"fraction"`
}

func LoadCatalog(data []byte) (*Catalog, error) {
	var raw rawCatalog
	if err := decodeStrict(data, &raw); err != nil {
		return nil, fmt.Errorf("%w: decode: %v", ErrInvalidCatalog, err)
	}
	if raw.SchemaVersion != CatalogSchemaVersion || raw.ContextVersion != CurrentContextVersion || raw.Gates == nil {
		return nil, fmt.Errorf("%w: unsupported schema/context version or missing gates", ErrInvalidCatalog)
	}
	if raw.DepletionDistinctRequired <= 0 || raw.DepletionDistinctRequired > int(decimal.MaxExactInteger) {
		return nil, fmt.Errorf("%w: invalid depletion route count", ErrInvalidCatalog)
	}
	if raw.Knowledge.RegistryFirstBonus <= 0 || raw.Knowledge.FounderFirstGrant <= 0 || raw.Knowledge.RepeatGrant <= 0 || raw.Knowledge.HintCost <= 0 {
		return nil, fmt.Errorf("%w: knowledge values must be positive", ErrInvalidCatalog)
	}
	catalog := &Catalog{
		contextVersion: raw.ContextVersion, depletionDistinctRequired: raw.DepletionDistinctRequired,
		knowledge: KnowledgePolicy(raw.Knowledge), gateByID: make(map[string]Gate), routeByID: make(map[string]Alternative),
	}
	for gateIndex, source := range raw.Gates {
		gate, err := parseGate(source)
		if err != nil {
			return nil, fmt.Errorf("%w: gates[%d]: %v", ErrInvalidCatalog, gateIndex, err)
		}
		if _, duplicate := catalog.gateByID[gate.ID]; duplicate {
			return nil, fmt.Errorf("%w: duplicate gate %q", ErrInvalidCatalog, gate.ID)
		}
		for _, route := range gate.Routes {
			if _, duplicate := catalog.routeByID[route.RouteID]; duplicate {
				return nil, fmt.Errorf("%w: duplicate route %q", ErrInvalidCatalog, route.RouteID)
			}
			if route.Active && route.RequiresContextVersion > catalog.contextVersion {
				return nil, fmt.Errorf("%w: active route %q requires unavailable context version", ErrInvalidCatalog, route.RouteID)
			}
			catalog.routeByID[route.RouteID] = route
		}
		catalog.gates = append(catalog.gates, gate)
		catalog.gateByID[gate.ID] = gate
	}
	if len(catalog.routeByID) == 0 {
		return nil, fmt.Errorf("%w: at least one route is required", ErrInvalidCatalog)
	}
	maximum := catalog.MaxRoutesPerRun()
	if maximum >= catalog.depletionDistinctRequired {
		return nil, fmt.Errorf("%w: depletion is reachable in one run: maximum %d >= %d", ErrInvalidCatalog, maximum, catalog.depletionDistinctRequired)
	}
	return catalog, nil
}

func parseGate(source rawGate) (Gate, error) {
	if !idPattern.MatchString(source.GateID) || len(source.Requirement) == 0 || source.Routes == nil {
		return Gate{}, errors.New("gate_id, requirement, and routes are required")
	}
	gate := Gate{ID: source.GateID}
	requirements := make(map[string]struct{})
	for index, item := range source.Requirement {
		if !idPattern.MatchString(item.ResourceID) {
			return Gate{}, fmt.Errorf("requirement[%d] has invalid resource_id", index)
		}
		if _, duplicate := requirements[item.ResourceID]; duplicate {
			return Gate{}, fmt.Errorf("duplicate requirement resource %q", item.ResourceID)
		}
		amount, err := decimal.ParseCanonical(item.Amount)
		if err != nil || !amount.Gt(decimal.Zero) {
			return Gate{}, fmt.Errorf("requirement[%d] amount must be positive canonical Decimal", index)
		}
		requirements[item.ResourceID] = struct{}{}
		gate.Requirement = append(gate.Requirement, Requirement{ResourceID: item.ResourceID, Amount: amount})
	}
	for index, item := range source.Routes {
		route, err := parseAlternative(item)
		if err != nil {
			return Gate{}, fmt.Errorf("routes[%d]: %v", index, err)
		}
		gate.Routes = append(gate.Routes, route)
	}
	return gate, nil
}

func parseAlternative(source rawAlternative) (Alternative, error) {
	if !idPattern.MatchString(source.RouteID) || source.HouseName == "" || source.RequiresContextVersion < 1 ||
		!validExclusionSlot(source.ExclusionSlot) || !idPattern.MatchString(source.ExclusionValue) || len(source.Predicate) == 0 {
		return Alternative{}, errors.New("invalid route declaration")
	}
	route := Alternative{
		RouteID: source.RouteID, HouseName: source.HouseName, Active: source.Active,
		RequiresContextVersion: source.RequiresContextVersion, ExclusionSlot: source.ExclusionSlot,
		ExclusionValue: source.ExclusionValue,
	}
	for index, item := range source.Predicate {
		condition, err := parseCondition(item)
		if err != nil {
			return Alternative{}, fmt.Errorf("predicate[%d]: %v", index, err)
		}
		route.Predicate = append(route.Predicate, condition)
	}
	for _, condition := range route.Predicate {
		if (condition.Kind == ConditionMeterBand || condition.Kind == ConditionRegionTrait) && route.RequiresContextVersion < 2 {
			return Alternative{}, errors.New("meter and region conditions require context version 2")
		}
	}
	if !hasExclusionCondition(route) {
		return Alternative{}, errors.New("exclusion slot/value must match an explicit predicate condition")
	}
	effect, err := parseEffect(source.Effect)
	if err != nil {
		return Alternative{}, err
	}
	route.Effect = effect
	return route, nil
}

func hasExclusionCondition(route Alternative) bool {
	if route.ExclusionSlot == "structure" {
		for _, condition := range route.Predicate {
			if condition.Kind == ConditionStructureIs && condition.StructureID == route.ExclusionValue {
				return true
			}
		}
		return false
	}
	transition := route.ExclusionSlot[len("doctrine:"):]
	for _, condition := range route.Predicate {
		if condition.Kind == ConditionDoctrineIs && condition.Transition == transition && condition.DoctrineID == route.ExclusionValue {
			return true
		}
	}
	return false
}

func parseCondition(source rawCondition) (Condition, error) {
	kind := ConditionKind(source.Kind)
	switch kind {
	case ConditionResourceAtLeast, ConditionResourceAtMost:
		if source.ResourceID == nil || source.Value == nil || !onlyConditionFields(source, "resource_id", "value") || !idPattern.MatchString(*source.ResourceID) {
			return Condition{}, errors.New("resource condition requires exactly resource_id and value")
		}
		value, err := decimal.ParseCanonical(*source.Value)
		if err != nil || !value.IsStateValue() {
			return Condition{}, errors.New("resource condition value must be canonical Decimal")
		}
		return Condition{Kind: kind, ResourceID: *source.ResourceID, Value: value}, nil
	case ConditionMeterBand:
		if source.MeterID == nil || source.Min == nil || source.Max == nil || !onlyConditionFields(source, "meter_id", "min", "max") ||
			!idPattern.MatchString(*source.MeterID) || *source.Min < 0 || *source.Max > 100 || *source.Min > *source.Max {
			return Condition{}, errors.New("meter_band requires exactly a mechanical meter_id and inclusive 0..100 min/max")
		}
		return Condition{Kind: kind, MeterID: *source.MeterID, Min: *source.Min, Max: *source.Max}, nil
	case ConditionDoctrineIs, ConditionDoctrineIsNot:
		if source.Transition == nil || source.DoctrineID == nil || !onlyConditionFields(source, "transition", "doctrine_id") ||
			!idPattern.MatchString(*source.Transition) || !idPattern.MatchString(*source.DoctrineID) {
			return Condition{}, errors.New("doctrine condition requires exactly transition and doctrine_id")
		}
		return Condition{Kind: kind, Transition: *source.Transition, DoctrineID: *source.DoctrineID}, nil
	case ConditionLedgerFact:
		if source.FactKind == nil || !onlyConditionFields(source, "fact_kind") || !idPattern.MatchString(*source.FactKind) {
			return Condition{}, errors.New("ledger fact condition requires exactly fact_kind")
		}
		return Condition{Kind: kind, FactKind: *source.FactKind}, nil
	case ConditionStructureIs:
		if source.StructureID == nil || !onlyConditionFields(source, "structure_id") || !idPattern.MatchString(*source.StructureID) {
			return Condition{}, errors.New("structure condition requires exactly structure_id")
		}
		return Condition{Kind: kind, StructureID: *source.StructureID}, nil
	case ConditionRegionTrait:
		if source.TraitID == nil || !onlyConditionFields(source, "trait_id") || !idPattern.MatchString(*source.TraitID) {
			return Condition{}, errors.New("region trait condition requires exactly trait_id")
		}
		return Condition{Kind: kind, TraitID: *source.TraitID}, nil
	default:
		return Condition{}, fmt.Errorf("unknown condition kind %q", source.Kind)
	}
}

func onlyConditionFields(source rawCondition, fields ...string) bool {
	present := map[string]bool{
		"resource_id": source.ResourceID != nil, "value": source.Value != nil, "meter_id": source.MeterID != nil,
		"min": source.Min != nil, "max": source.Max != nil, "transition": source.Transition != nil,
		"doctrine_id": source.DoctrineID != nil, "structure_id": source.StructureID != nil,
		"fact_kind": source.FactKind != nil, "trait_id": source.TraitID != nil,
	}
	wanted := make(map[string]bool, len(fields))
	for _, field := range fields {
		wanted[field] = true
	}
	for field, exists := range present {
		if exists != wanted[field] {
			return false
		}
	}
	return true
}

func parseEffect(source rawEffect) (Effect, error) {
	switch EffectKind(source.Kind) {
	case EffectSubstitute:
		if source.Fraction != nil {
			return Effect{}, errors.New("substitute forbids fraction")
		}
		return Effect{Kind: EffectSubstitute}, nil
	case EffectDiscount:
		if source.Fraction == nil {
			return Effect{}, errors.New("discount requires fraction")
		}
		fraction, err := decimal.ParseCanonical(*source.Fraction)
		if err != nil || !fraction.Gt(decimal.Zero) || !fraction.Lt(decimal.One) {
			return Effect{}, errors.New("discount fraction must be canonical and in (0,1)")
		}
		return Effect{Kind: EffectDiscount, Fraction: fraction}, nil
	default:
		return Effect{}, fmt.Errorf("unknown effect kind %q", source.Kind)
	}
}

func validExclusionSlot(value string) bool {
	if value == "structure" {
		return true
	}
	const prefix = "doctrine:"
	return len(value) > len(prefix) && value[:len(prefix)] == prefix && idPattern.MatchString(value[len(prefix):])
}

func decodeStrict(data []byte, destination any) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return err
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			return errors.New("multiple JSON values")
		}
		return err
	}
	return nil
}

func (c *Catalog) ContextVersion() int              { return c.contextVersion }
func (c *Catalog) DepletionDistinctRequired() int   { return c.depletionDistinctRequired }
func (c *Catalog) KnowledgePolicy() KnowledgePolicy { return c.knowledge }

func (c *Catalog) Gate(id string) (Gate, bool) {
	gate, ok := c.gateByID[id]
	return cloneGate(gate), ok
}

func (c *Catalog) Route(id string) (Alternative, bool) {
	route, ok := c.routeByID[id]
	return cloneAlternative(route), ok
}

func (c *Catalog) Gates() []Gate {
	result := make([]Gate, len(c.gates))
	for index, gate := range c.gates {
		result[index] = cloneGate(gate)
	}
	return result
}

func cloneGate(source Gate) Gate {
	result := source
	result.Requirement = append([]Requirement(nil), source.Requirement...)
	result.Routes = make([]Alternative, len(source.Routes))
	for index, route := range source.Routes {
		result.Routes[index] = cloneAlternative(route)
	}
	return result
}

func cloneAlternative(source Alternative) Alternative {
	result := source
	result.Predicate = append([]Condition(nil), source.Predicate...)
	return result
}

// MaxRoutesPerRun exhaustively enumerates declared exclusion-slot assignments.
func (c *Catalog) MaxRoutesPerRun() int {
	valuesBySlot := make(map[string]map[string]struct{})
	for _, route := range c.routeByID {
		if valuesBySlot[route.ExclusionSlot] == nil {
			valuesBySlot[route.ExclusionSlot] = make(map[string]struct{})
		}
		valuesBySlot[route.ExclusionSlot][route.ExclusionValue] = struct{}{}
	}
	slots := make([]string, 0, len(valuesBySlot))
	for slot := range valuesBySlot {
		slots = append(slots, slot)
	}
	sort.Strings(slots)
	maximum := 0
	var search func(int, map[string]string)
	search = func(index int, assignment map[string]string) {
		if index == len(slots) {
			count := 0
			for _, route := range c.routeByID {
				if assignment[route.ExclusionSlot] == route.ExclusionValue {
					count++
				}
			}
			if count > maximum {
				maximum = count
			}
			return
		}
		values := make([]string, 0, len(valuesBySlot[slots[index]]))
		for value := range valuesBySlot[slots[index]] {
			values = append(values, value)
		}
		sort.Strings(values)
		for _, value := range values {
			assignment[slots[index]] = value
			search(index+1, assignment)
		}
	}
	search(0, make(map[string]string))
	return maximum
}

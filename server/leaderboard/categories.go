package leaderboard

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"
)

type CategoryTimer string

const (
	TimerRTA      CategoryTimer = "rta"
	TimerAttended CategoryTimer = "attended"
)

type PredicateKind string

const (
	PredicateAny           PredicateKind = "any"
	PredicateAllGates      PredicateKind = "all_gates"
	PredicateFactsSuperset PredicateKind = "facts_superset"
	PredicateFactsDisjoint PredicateKind = "facts_disjoint"
	PredicateCountAtMost   PredicateKind = "count_at_most"
	PredicateAllOf         PredicateKind = "all_of"
)

type Predicate struct {
	Kind     PredicateKind
	SetRef   string
	Field    string
	Literal  int64
	Children []Predicate
}

type Category struct {
	ID        string
	NameKey   string
	Timer     CategoryTimer
	Predicate Predicate
}

type CategoryCatalog struct {
	FullGateSet []string
	FactSets    map[string][]string
	Categories  []Category
}

type TerminalFacts struct {
	GatesCrossed             []string
	Facts                    []string
	GeneratorsPurchasedTotal int64
}

type rawCategoryCatalog struct {
	SchemaVersion int                 `json:"schema_version"`
	FullGateSet   []string            `json:"full_gate_set"`
	FactSets      map[string][]string `json:"fact_sets"`
	Categories    []rawCategory       `json:"categories"`
}

type rawCategory struct {
	ID        string       `json:"id"`
	NameKey   string       `json:"name_key"`
	Timer     string       `json:"timer"`
	Predicate rawPredicate `json:"predicate"`
}

type rawPredicate struct {
	Kind    string         `json:"kind"`
	SetRef  *string        `json:"set_ref,omitempty"`
	Field   *string        `json:"field,omitempty"`
	Literal *int64         `json:"literal,omitempty"`
	All     []rawPredicate `json:"all,omitempty"`
}

func LoadCategoryCatalog(data []byte, routeGateIDs []string) (*CategoryCatalog, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var raw rawCategoryCatalog
	if err := decoder.Decode(&raw); err != nil {
		return nil, fmt.Errorf("%w: category catalog: %v", ErrInvalidEpoch, err)
	}
	if err := rejectTrailingCategoryJSON(decoder); err != nil || raw.SchemaVersion != 1 ||
		!sortedUniqueMechanical(raw.FullGateSet) || !sameStrings(raw.FullGateSet, routeGateIDs) || len(raw.FactSets) != 2 || len(raw.Categories) != 4 {
		return nil, fmt.Errorf("%w: category catalog envelope", ErrInvalidEpoch)
	}
	catalog := &CategoryCatalog{FullGateSet: append([]string(nil), raw.FullGateSet...), FactSets: map[string][]string{}, Categories: make([]Category, 0, 4)}
	for _, id := range []string{"completion_set", "forbidden_set"} {
		values, ok := raw.FactSets[id]
		if !ok || !sortedUniqueMechanical(values) {
			return nil, fmt.Errorf("%w: fact set %s", ErrInvalidEpoch, id)
		}
		catalog.FactSets[id] = append([]string(nil), values...)
	}
	seen := map[string]bool{}
	for _, source := range raw.Categories {
		if seen[source.ID] || !mechanicalPattern.MatchString(source.ID) || !mechanicalPattern.MatchString(source.NameKey) ||
			(source.Timer != string(TimerRTA) && source.Timer != string(TimerAttended)) {
			return nil, fmt.Errorf("%w: category row", ErrInvalidEpoch)
		}
		predicate, err := decodePredicate(source.Predicate, catalog.FactSets, 0)
		if err != nil {
			return nil, err
		}
		seen[source.ID] = true
		catalog.Categories = append(catalog.Categories, Category{ID: source.ID, NameKey: source.NameKey, Timer: CategoryTimer(source.Timer), Predicate: predicate})
	}
	expected := []string{"any_percent", "ethical_percent", "hundred_percent", "low_percent"}
	sort.Slice(catalog.Categories, func(i, j int) bool { return catalog.Categories[i].ID < catalog.Categories[j].ID })
	actual := make([]string, len(catalog.Categories))
	for index := range catalog.Categories {
		actual[index] = catalog.Categories[index].ID
	}
	if !sameStrings(actual, expected) || !canonicalPhase0Shapes(catalog) {
		return nil, fmt.Errorf("%w: canonical category rows", ErrInvalidEpoch)
	}
	return catalog, nil
}

func (catalog *CategoryCatalog) Matching(facts TerminalFacts) ([]Category, error) {
	if catalog == nil || !sortedUniqueMechanical(facts.GatesCrossed) || !sortedUniqueMechanical(facts.Facts) || facts.GeneratorsPurchasedTotal < 0 {
		return nil, ErrInvalidEpoch
	}
	result := make([]Category, 0, len(catalog.Categories))
	for _, category := range catalog.Categories {
		matched, err := catalog.evaluate(category.Predicate, facts)
		if err != nil {
			return nil, err
		}
		if matched {
			result = append(result, category)
		}
	}
	return result, nil
}

func (catalog *CategoryCatalog) evaluate(predicate Predicate, facts TerminalFacts) (bool, error) {
	switch predicate.Kind {
	case PredicateAny:
		return true, nil
	case PredicateAllGates:
		return sameStrings(facts.GatesCrossed, catalog.FullGateSet), nil
	case PredicateFactsSuperset:
		return containsAll(facts.Facts, catalog.FactSets[predicate.SetRef]), nil
	case PredicateFactsDisjoint:
		return disjoint(facts.Facts, catalog.FactSets[predicate.SetRef]), nil
	case PredicateCountAtMost:
		return facts.GeneratorsPurchasedTotal <= predicate.Literal, nil
	case PredicateAllOf:
		for _, child := range predicate.Children {
			matched, err := catalog.evaluate(child, facts)
			if err != nil || !matched {
				return matched, err
			}
		}
		return true, nil
	default:
		return false, ErrInvalidEpoch
	}
}

func decodePredicate(source rawPredicate, sets map[string][]string, depth int) (Predicate, error) {
	if depth > 4 {
		return Predicate{}, fmt.Errorf("%w: predicate depth", ErrInvalidEpoch)
	}
	predicate := Predicate{Kind: PredicateKind(source.Kind)}
	switch predicate.Kind {
	case PredicateAny, PredicateAllGates:
		if source.SetRef != nil || source.Field != nil || source.Literal != nil || len(source.All) != 0 {
			return Predicate{}, fmt.Errorf("%w: predicate fields", ErrInvalidEpoch)
		}
	case PredicateFactsSuperset, PredicateFactsDisjoint:
		if source.SetRef == nil || source.Field != nil || source.Literal != nil || len(source.All) != 0 {
			return Predicate{}, fmt.Errorf("%w: set predicate", ErrInvalidEpoch)
		}
		if _, ok := sets[*source.SetRef]; !ok {
			return Predicate{}, fmt.Errorf("%w: unknown fact set", ErrInvalidEpoch)
		}
		predicate.SetRef = *source.SetRef
	case PredicateCountAtMost:
		if source.SetRef != nil || source.Field == nil || *source.Field != "generators_purchased_total" || source.Literal == nil || *source.Literal < 0 || len(source.All) != 0 {
			return Predicate{}, fmt.Errorf("%w: count predicate", ErrInvalidEpoch)
		}
		predicate.Field, predicate.Literal = *source.Field, *source.Literal
	case PredicateAllOf:
		if source.SetRef != nil || source.Field != nil || source.Literal != nil || len(source.All) < 2 || len(source.All) > 8 {
			return Predicate{}, fmt.Errorf("%w: all_of predicate", ErrInvalidEpoch)
		}
		for _, child := range source.All {
			decoded, err := decodePredicate(child, sets, depth+1)
			if err != nil {
				return Predicate{}, err
			}
			predicate.Children = append(predicate.Children, decoded)
		}
	default:
		return Predicate{}, fmt.Errorf("%w: predicate kind", ErrInvalidEpoch)
	}
	return predicate, nil
}

func canonicalPhase0Shapes(catalog *CategoryCatalog) bool {
	byID := map[string]Category{}
	for _, category := range catalog.Categories {
		byID[category.ID] = category
	}
	any := byID["any_percent"]
	ethical := byID["ethical_percent"]
	hundred := byID["hundred_percent"]
	low := byID["low_percent"]
	return any.Timer == TimerRTA && any.Predicate.Kind == PredicateAny &&
		ethical.Timer == TimerAttended && ethical.Predicate.Kind == PredicateFactsDisjoint && ethical.Predicate.SetRef == "forbidden_set" &&
		hundred.Timer == TimerRTA && hundred.Predicate.Kind == PredicateAllOf && len(hundred.Predicate.Children) == 2 &&
		hundred.Predicate.Children[0].Kind == PredicateAllGates && hundred.Predicate.Children[1].Kind == PredicateFactsSuperset && hundred.Predicate.Children[1].SetRef == "completion_set" &&
		low.Timer == TimerRTA && low.Predicate.Kind == PredicateCountAtMost && low.Predicate.Literal == 40
}

func rejectTrailingCategoryJSON(decoder *json.Decoder) error {
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		return ErrInvalidEpoch
	}
	return nil
}

func sortedUniqueMechanical(values []string) bool {
	last := ""
	for _, value := range values {
		if !mechanicalPattern.MatchString(value) || value <= last {
			return false
		}
		last = value
	}
	return true
}

func sameStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func containsAll(values, required []string) bool {
	present := make(map[string]bool, len(values))
	for _, value := range values {
		present[value] = true
	}
	for _, value := range required {
		if !present[value] {
			return false
		}
	}
	return true
}

func disjoint(left, right []string) bool {
	present := make(map[string]bool, len(left))
	for _, value := range left {
		present[value] = true
	}
	for _, value := range right {
		if present[value] {
			return false
		}
	}
	return true
}

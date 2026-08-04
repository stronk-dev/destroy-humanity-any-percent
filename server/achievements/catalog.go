// Package achievements owns permanent achievement IDs and their exact score.
// Achievement score is neither an economy resource nor a Clout mint.
package achievements

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
	maximumConditionDepth = 4
	maximumConditionNodes = 64
)

var (
	ErrInvalidCatalog = errors.New("invalid achievement catalog")
	mechanicalID      = regexp.MustCompile(`^[a-z][a-z0-9_]*(?:\.[a-z][a-z0-9_]*)*$`)
)

type ConditionScope string

const (
	ScopeRun    ConditionScope = "run"
	ScopeCareer ConditionScope = "career"
)

type ConditionKind string

const (
	ConditionFactPresent          ConditionKind = "fact_present"
	ConditionCounterAtLeast       ConditionKind = "counter_at_least"
	ConditionExitCountAtLeast     ConditionKind = "exit_count_at_least"
	ConditionOwnsGeneratorAtLeast ConditionKind = "owns_generator_at_least"
	ConditionAllOf                ConditionKind = "all_of"
)

type ProofKind string

const (
	ProofProvenance ProofKind = "provenance"
	ProofBurn       ProofKind = "burn"
	ProofPossession ProofKind = "possession"
)

type Condition struct {
	Kind        ConditionKind
	FactKind    string
	Counter     string
	Minimum     int64
	GeneratorID string
	Children    []Condition
}

type Proof struct {
	Kind                 ProofKind
	EventKinds           []string
	EventKind            string
	ResourceID           string
	Minimum              string
	JustificationCopyKey string
}

type Definition struct {
	ID             string
	ConditionScope ConditionScope
	Condition      Condition
	Proof          Proof
	ScoreGrant     int64
	CopyKey        string
}

type Registry struct {
	CopyKeys       map[string]bool
	GeneratorIDs   map[string]bool
	EventKinds     map[string]bool
	ResourceIDs    map[string]bool
	RunCounters    map[string]bool
	CareerCounters map[string]bool
}

type Catalog struct {
	Definitions []Definition
	byID        map[string]Definition
}

type rawCatalog struct {
	SchemaVersion int             `json:"schema_version"`
	Achievements  []rawDefinition `json:"achievements"`
}

type rawDefinition struct {
	ID             string          `json:"id"`
	ConditionScope ConditionScope  `json:"condition_scope"`
	Condition      json.RawMessage `json:"condition"`
	Proof          json.RawMessage `json:"proof"`
	ScoreGrant     int64           `json:"score_grant"`
	CopyKey        string          `json:"copy_key"`
}

type rawCondition struct {
	Kind        ConditionKind     `json:"kind"`
	FactKind    string            `json:"fact_kind,omitempty"`
	Counter     string            `json:"counter,omitempty"`
	Minimum     *int64            `json:"minimum,omitempty"`
	Count       *int64            `json:"count,omitempty"`
	GeneratorID string            `json:"generator_id,omitempty"`
	Conditions  []json.RawMessage `json:"conditions,omitempty"`
}

type rawProof struct {
	Kind                 ProofKind `json:"kind"`
	EventKinds           []string  `json:"event_kinds,omitempty"`
	EventKind            string    `json:"event_kind,omitempty"`
	ResourceID           string    `json:"resource_id,omitempty"`
	Minimum              string    `json:"minimum,omitempty"`
	JustificationCopyKey string    `json:"justification_copy_key,omitempty"`
}

func LoadCatalog(data []byte, registry Registry) (*Catalog, error) {
	if !validRegistry(registry) {
		return nil, ErrInvalidCatalog
	}
	var raw rawCatalog
	if err := decodeStrict(data, &raw); err != nil || raw.SchemaVersion != CatalogSchemaVersion || raw.Achievements == nil {
		return nil, fmt.Errorf("%w: root", ErrInvalidCatalog)
	}
	catalog := &Catalog{Definitions: make([]Definition, 0, len(raw.Achievements)), byID: map[string]Definition{}}
	lastID := ""
	for index, source := range raw.Achievements {
		if !mechanicalID.MatchString(source.ID) || source.ID <= lastID || (source.ConditionScope != ScopeRun && source.ConditionScope != ScopeCareer) ||
			source.ScoreGrant < 1 || source.ScoreGrant > decimal.MaxExactInteger || !registry.CopyKeys[source.CopyKey] {
			return nil, fmt.Errorf("%w: achievements[%d]", ErrInvalidCatalog, index)
		}
		condition, nodes, err := parseCondition(source.Condition, source.ConditionScope, registry, 1)
		if err != nil || nodes > maximumConditionNodes {
			return nil, fmt.Errorf("%w: achievements[%d].condition", ErrInvalidCatalog, index)
		}
		proof, err := parseProof(source.Proof, source.ConditionScope, condition, registry)
		if err != nil {
			return nil, fmt.Errorf("%w: achievements[%d].proof", ErrInvalidCatalog, index)
		}
		definition := Definition{ID: source.ID, ConditionScope: source.ConditionScope, Condition: condition, Proof: proof, ScoreGrant: source.ScoreGrant, CopyKey: source.CopyKey}
		catalog.Definitions = append(catalog.Definitions, definition)
		catalog.byID[definition.ID], lastID = definition, definition.ID
	}
	return catalog, nil
}

func parseCondition(data []byte, scope ConditionScope, registry Registry, depth int) (Condition, int, error) {
	if depth > maximumConditionDepth {
		return Condition{}, 0, ErrInvalidCatalog
	}
	var discriminator struct {
		Kind ConditionKind `json:"kind"`
	}
	if json.Unmarshal(data, &discriminator) != nil {
		return Condition{}, 0, ErrInvalidCatalog
	}
	var raw rawCondition
	if err := decodeStrict(data, &raw); err != nil {
		return Condition{}, 0, err
	}
	switch raw.Kind {
	case ConditionFactPresent:
		if !mechanicalID.MatchString(raw.FactKind) || raw.Counter != "" || raw.Minimum != nil || raw.Count != nil || raw.GeneratorID != "" || raw.Conditions != nil {
			return Condition{}, 0, ErrInvalidCatalog
		}
		return Condition{Kind: raw.Kind, FactKind: raw.FactKind}, 1, nil
	case ConditionCounterAtLeast:
		counters := registry.RunCounters
		if scope == ScopeCareer {
			counters = registry.CareerCounters
		}
		if !counters[raw.Counter] || raw.Minimum == nil || *raw.Minimum < 0 || *raw.Minimum > decimal.MaxExactInteger || raw.FactKind != "" || raw.Count != nil || raw.GeneratorID != "" || raw.Conditions != nil {
			return Condition{}, 0, ErrInvalidCatalog
		}
		return Condition{Kind: raw.Kind, Counter: raw.Counter, Minimum: *raw.Minimum}, 1, nil
	case ConditionExitCountAtLeast:
		if scope != ScopeCareer || raw.Count == nil || *raw.Count < 1 || *raw.Count > decimal.MaxExactInteger || raw.FactKind != "" || raw.Counter != "" || raw.Minimum != nil || raw.GeneratorID != "" || raw.Conditions != nil {
			return Condition{}, 0, ErrInvalidCatalog
		}
		return Condition{Kind: raw.Kind, Minimum: *raw.Count}, 1, nil
	case ConditionOwnsGeneratorAtLeast:
		if scope != ScopeRun || !registry.GeneratorIDs[raw.GeneratorID] || raw.Count == nil || *raw.Count < 1 || *raw.Count > decimal.MaxExactInteger || raw.FactKind != "" || raw.Counter != "" || raw.Minimum != nil || raw.Conditions != nil {
			return Condition{}, 0, ErrInvalidCatalog
		}
		return Condition{Kind: raw.Kind, GeneratorID: raw.GeneratorID, Minimum: *raw.Count}, 1, nil
	case ConditionAllOf:
		if len(raw.Conditions) < 2 || len(raw.Conditions) > 16 || raw.FactKind != "" || raw.Counter != "" || raw.Minimum != nil || raw.Count != nil || raw.GeneratorID != "" {
			return Condition{}, 0, ErrInvalidCatalog
		}
		result := Condition{Kind: raw.Kind, Children: make([]Condition, len(raw.Conditions))}
		nodes := 1
		for index, child := range raw.Conditions {
			parsed, childNodes, err := parseCondition(child, scope, registry, depth+1)
			if err != nil {
				return Condition{}, 0, err
			}
			result.Children[index], nodes = parsed, nodes+childNodes
		}
		return result, nodes, nil
	default:
		return Condition{}, 0, ErrInvalidCatalog
	}
}

func parseProof(data []byte, scope ConditionScope, condition Condition, registry Registry) (Proof, error) {
	var raw rawProof
	if err := decodeStrict(data, &raw); err != nil {
		return Proof{}, err
	}
	switch raw.Kind {
	case ProofProvenance:
		if len(raw.EventKinds) == 0 || raw.EventKind != "" || raw.ResourceID != "" || raw.Minimum != "" || raw.JustificationCopyKey != "" || containsKind(condition, ConditionOwnsGeneratorAtLeast) {
			return Proof{}, ErrInvalidCatalog
		}
		last := ""
		for _, kind := range raw.EventKinds {
			if kind <= last || !registry.EventKinds[kind] {
				return Proof{}, ErrInvalidCatalog
			}
			last = kind
		}
		return Proof{Kind: raw.Kind, EventKinds: append([]string(nil), raw.EventKinds...)}, nil
	case ProofBurn:
		if scope != ScopeRun || !registry.EventKinds[raw.EventKind] || !registry.ResourceIDs[raw.ResourceID] || raw.EventKinds != nil || raw.JustificationCopyKey != "" {
			return Proof{}, ErrInvalidCatalog
		}
		minimum, err := decimal.ParseCanonical(raw.Minimum)
		if err != nil || !minimum.IsStateValue() || !minimum.Gt(decimal.Zero) {
			return Proof{}, ErrInvalidCatalog
		}
		return Proof{Kind: raw.Kind, EventKind: raw.EventKind, ResourceID: raw.ResourceID, Minimum: minimum.String()}, nil
	case ProofPossession:
		if scope != ScopeRun || !containsKind(condition, ConditionOwnsGeneratorAtLeast) || !registry.CopyKeys[raw.JustificationCopyKey] || raw.EventKinds != nil || raw.EventKind != "" || raw.ResourceID != "" || raw.Minimum != "" {
			return Proof{}, ErrInvalidCatalog
		}
		return Proof{Kind: raw.Kind, JustificationCopyKey: raw.JustificationCopyKey}, nil
	default:
		return Proof{}, ErrInvalidCatalog
	}
}

func (catalog *Catalog) Definition(id string) (Definition, bool) {
	if catalog == nil {
		return Definition{}, false
	}
	value, ok := catalog.byID[id]
	return cloneDefinition(value), ok
}

func (catalog *Catalog) Score(ids map[string]bool) (int64, error) {
	if catalog == nil || ids == nil {
		return 0, ErrInvalidCatalog
	}
	var total int64
	for id, earned := range ids {
		definition, ok := catalog.byID[id]
		if !earned || !ok || total > decimal.MaxExactInteger-definition.ScoreGrant {
			return 0, ErrInvalidCatalog
		}
		total += definition.ScoreGrant
	}
	return total, nil
}

func containsKind(condition Condition, kind ConditionKind) bool {
	if condition.Kind == kind {
		return true
	}
	for _, child := range condition.Children {
		if containsKind(child, kind) {
			return true
		}
	}
	return false
}

func validRegistry(registry Registry) bool {
	return registry.CopyKeys != nil && registry.GeneratorIDs != nil && registry.EventKinds != nil && registry.ResourceIDs != nil && registry.RunCounters != nil && registry.CareerCounters != nil
}

func decodeStrict(data []byte, target any) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return ErrInvalidCatalog
	}
	return nil
}

func cloneDefinition(source Definition) Definition {
	result := source
	result.Condition = cloneCondition(source.Condition)
	result.Proof.EventKinds = append([]string(nil), source.Proof.EventKinds...)
	return result
}

func cloneCondition(source Condition) Condition {
	result := source
	result.Children = make([]Condition, len(source.Children))
	for index := range source.Children {
		result.Children[index] = cloneCondition(source.Children[index])
	}
	return result
}

func SortedEarnedIDs(ids map[string]bool) []string {
	result := make([]string, 0, len(ids))
	for id, earned := range ids {
		if earned {
			result = append(result, id)
		}
	}
	sort.Strings(result)
	return result
}

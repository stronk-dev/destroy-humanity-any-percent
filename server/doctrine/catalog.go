// Package doctrine owns the immutable doctrine-choice catalog. Doctrine
// effects deliberately live elsewhere; this package only defines which
// write-once choice is legal at a tier transition.
package doctrine

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strconv"

	"cloud-clicker/server/decimal"
	"cloud-clicker/server/routes"
)

const CatalogSchemaVersion = 1

var (
	ErrInvalidCatalog = errors.New("invalid doctrine catalog")
	idPattern         = regexp.MustCompile(`^[a-z][a-z0-9_]*(?:\.[a-z][a-z0-9_]*)*$`)
	transitionPattern = regexp.MustCompile(`^transition\.t([0-9]+)_to_t([0-9]+)$`)
	gatePattern       = regexp.MustCompile(`^gate\.t([0-9]+)_to_t([0-9]+)$`)
)

type Transition struct {
	ID          string
	SourceTier  int64
	GateID      string
	DoctrineIDs []string
}

type Catalog struct {
	transitions []Transition
	byID        map[string]Transition
}

type rawCatalog struct {
	SchemaVersion int             `json:"schema_version"`
	Transitions   []rawTransition `json:"transitions"`
}

type rawTransition struct {
	TransitionID string   `json:"transition_id"`
	SourceTier   int64    `json:"source_tier"`
	GateID       string   `json:"gate_id"`
	DoctrineIDs  []string `json:"doctrine_ids"`
}

func LoadCatalog(data []byte) (*Catalog, error) {
	var raw rawCatalog
	if err := decodeStrict(data, &raw); err != nil {
		return nil, fmt.Errorf("%w: decode: %v", ErrInvalidCatalog, err)
	}
	if raw.SchemaVersion != CatalogSchemaVersion || len(raw.Transitions) == 0 {
		return nil, fmt.Errorf("%w: unsupported schema or empty transitions", ErrInvalidCatalog)
	}
	catalog := &Catalog{byID: make(map[string]Transition, len(raw.Transitions))}
	previous := ""
	for index, source := range raw.Transitions {
		transition, err := parseTransition(source)
		if err != nil {
			return nil, fmt.Errorf("%w: transitions[%d]: %v", ErrInvalidCatalog, index, err)
		}
		if previous != "" && previous >= transition.ID {
			return nil, fmt.Errorf("%w: transitions must be byte-sorted and unique", ErrInvalidCatalog)
		}
		previous = transition.ID
		catalog.transitions = append(catalog.transitions, transition)
		catalog.byID[transition.ID] = transition
	}
	return catalog, nil
}

func parseTransition(source rawTransition) (Transition, error) {
	transitionTier, ok := adjacentBoundary(transitionPattern, source.TransitionID)
	gateTier, gateOK := adjacentBoundary(gatePattern, source.GateID)
	if !ok || !gateOK || transitionTier != source.SourceTier || gateTier != source.SourceTier ||
		source.SourceTier < 0 || source.SourceTier > 8 || len(source.DoctrineIDs) < 2 {
		return Transition{}, errors.New("transition, source tier, gate, and branching doctrine set must describe one adjacent boundary")
	}
	doctrineIDs := append([]string(nil), source.DoctrineIDs...)
	for index, id := range doctrineIDs {
		if !idPattern.MatchString(id) || index > 0 && doctrineIDs[index-1] >= id {
			return Transition{}, errors.New("doctrine_ids must be byte-sorted, unique mechanical IDs")
		}
	}
	return Transition{ID: source.TransitionID, SourceTier: source.SourceTier, GateID: source.GateID, DoctrineIDs: doctrineIDs}, nil
}

func adjacentBoundary(pattern *regexp.Regexp, value string) (int64, bool) {
	match := pattern.FindStringSubmatch(value)
	if len(match) != 3 {
		return 0, false
	}
	from, fromErr := strconv.ParseInt(match[1], 10, 64)
	to, toErr := strconv.ParseInt(match[2], 10, 64)
	return from, fromErr == nil && toErr == nil && from >= 0 && from <= decimal.MaxExactInteger && to == from+1
}

func (catalog *Catalog) Transition(id string) (Transition, bool) {
	if catalog == nil {
		return Transition{}, false
	}
	value, ok := catalog.byID[id]
	value.DoctrineIDs = append([]string(nil), value.DoctrineIDs...)
	return value, ok
}

func (catalog *Catalog) Transitions() []Transition {
	if catalog == nil {
		return nil
	}
	result := make([]Transition, len(catalog.transitions))
	for index, value := range catalog.transitions {
		result[index] = value
		result[index].DoctrineIDs = append([]string(nil), value.DoctrineIDs...)
	}
	return result
}

func (catalog *Catalog) Allows(transitionID, doctrineID string) bool {
	transition, ok := catalog.Transition(transitionID)
	if !ok {
		return false
	}
	index := sort.SearchStrings(transition.DoctrineIDs, doctrineID)
	return index < len(transition.DoctrineIDs) && transition.DoctrineIDs[index] == doctrineID
}

// ValidateRoutes closes the cross-artifact identity boundary. A route may
// consume only a declared doctrine choice, and every doctrine choice gate must
// be present in the same pinned routes artifact.
func (catalog *Catalog) ValidateRoutes(routeCatalog *routes.Catalog) error {
	if catalog == nil || routeCatalog == nil {
		return ErrInvalidCatalog
	}
	for _, transition := range catalog.transitions {
		if _, ok := routeCatalog.Gate(transition.GateID); !ok {
			return fmt.Errorf("%w: doctrine gate %q is absent from routes", ErrInvalidCatalog, transition.GateID)
		}
	}
	for _, gate := range routeCatalog.Gates() {
		for _, alternative := range gate.Routes {
			for _, condition := range alternative.Predicate {
				if condition.Kind != routes.ConditionDoctrineIs && condition.Kind != routes.ConditionDoctrineIsNot {
					continue
				}
				if !catalog.Allows(condition.Transition, condition.DoctrineID) {
					return fmt.Errorf("%w: route %q references undeclared doctrine %s/%s", ErrInvalidCatalog, alternative.RouteID, condition.Transition, condition.DoctrineID)
				}
			}
		}
	}
	return nil
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

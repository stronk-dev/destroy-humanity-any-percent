// Package economy implements the configuration-driven economy kernel.
package economy

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"

	"cloud-clicker/server/decimal"
)

const CatalogSchemaVersion = 2

var (
	ErrInvalidCatalog = errors.New("invalid economy catalog")
	idPattern         = regexp.MustCompile(`^[a-z][a-z0-9_]*(?:\.[a-z][a-z0-9_]*)*$`)
)

type Scope string

const (
	ScopeCompany Scope = "company"
	ScopeFounder Scope = "founder"
	ScopeWorld   Scope = "world"
	ScopeGuild   Scope = "guild"
)

type CurveKind string

const (
	CurveConstant  CurveKind = "constant"
	CurveLinear    CurveKind = "linear"
	CurveGeometric CurveKind = "geometric"
)

type Hardcap struct {
	Amount    decimal.Decimal
	ReasonKey string
}

type ResourceDefinition struct {
	ID          string
	Scope       Scope
	NumericKind string
	Initial     decimal.Decimal
	Minimum     decimal.Decimal
	Hardcap     *Hardcap
}

type CostCurve struct {
	Kind  CurveKind
	Step  decimal.Decimal
	Ratio decimal.Decimal
}

type PriceDefinition struct {
	ResourceID string
	Base       decimal.Decimal
	Curve      CostCurve
}

type GeneratorClassDefinition struct {
	ID         string
	Price      PriceDefinition
	Production *ProductionDefinition
}

type ProductionDefinition struct {
	ResourceID string
	BaseRate   decimal.Decimal
}

type Catalog struct {
	resources     []ResourceDefinition
	resourceByID  map[string]ResourceDefinition
	generators    []GeneratorClassDefinition
	generatorByID map[string]GeneratorClassDefinition
}

type rawCatalog struct {
	SchemaVersion    int                 `json:"schema_version"`
	Resources        []rawResource       `json:"resources"`
	GeneratorClasses []rawGeneratorClass `json:"generator_classes"`
}

type rawResource struct {
	ID          string          `json:"id"`
	Scope       string          `json:"scope"`
	NumericKind string          `json:"numeric_kind"`
	Initial     string          `json:"initial"`
	Minimum     string          `json:"minimum"`
	Hardcap     json.RawMessage `json:"hardcap"`
}

type rawHardcap struct {
	Amount    string `json:"amount"`
	ReasonKey string `json:"reason_key"`
}

type rawGeneratorClass struct {
	ID         string         `json:"id"`
	Price      rawPrice       `json:"price"`
	Production *rawProduction `json:"production"`
}

type rawProduction struct {
	ResourceID string `json:"resource_id"`
	BaseRate   string `json:"base_rate"`
}

type rawPrice struct {
	ResourceID string   `json:"resource_id"`
	Base       string   `json:"base"`
	Curve      rawCurve `json:"curve"`
}

type rawCurve struct {
	Kind  string  `json:"kind"`
	Step  *string `json:"step"`
	Ratio *string `json:"ratio"`
}

func LoadCatalog(data []byte) (*Catalog, error) {
	var raw rawCatalog
	if err := decodeStrict(data, &raw); err != nil {
		return nil, catalogError("decode", err)
	}
	if raw.SchemaVersion != 1 && raw.SchemaVersion != CatalogSchemaVersion {
		return nil, catalogError("schema_version", fmt.Errorf("got %d, supported versions are 1 and %d", raw.SchemaVersion, CatalogSchemaVersion))
	}
	if raw.Resources == nil {
		return nil, catalogError("resources", errors.New("field is required"))
	}
	if raw.GeneratorClasses == nil {
		return nil, catalogError("generator_classes", errors.New("field is required"))
	}

	catalog := &Catalog{
		resources:     make([]ResourceDefinition, 0, len(raw.Resources)),
		resourceByID:  make(map[string]ResourceDefinition, len(raw.Resources)),
		generators:    make([]GeneratorClassDefinition, 0, len(raw.GeneratorClasses)),
		generatorByID: make(map[string]GeneratorClassDefinition, len(raw.GeneratorClasses)),
	}

	for index, source := range raw.Resources {
		definition, err := parseResource(source)
		if err != nil {
			return nil, catalogError(fmt.Sprintf("resources[%d]", index), err)
		}
		if _, exists := catalog.resourceByID[definition.ID]; exists {
			return nil, catalogError("resources", fmt.Errorf("duplicate id %q", definition.ID))
		}
		catalog.resources = append(catalog.resources, definition)
		catalog.resourceByID[definition.ID] = definition
	}

	for index, source := range raw.GeneratorClasses {
		definition, err := parseGenerator(source, raw.SchemaVersion)
		if err != nil {
			return nil, catalogError(fmt.Sprintf("generator_classes[%d]", index), err)
		}
		if _, exists := catalog.generatorByID[definition.ID]; exists {
			return nil, catalogError("generator_classes", fmt.Errorf("duplicate id %q", definition.ID))
		}
		if _, exists := catalog.resourceByID[definition.Price.ResourceID]; !exists {
			return nil, catalogError("generator_classes", fmt.Errorf("%q references unknown resource %q", definition.ID, definition.Price.ResourceID))
		}
		if definition.Production != nil {
			output, exists := catalog.resourceByID[definition.Production.ResourceID]
			if !exists {
				return nil, catalogError("generator_classes", fmt.Errorf("%q references unknown production resource %q", definition.ID, definition.Production.ResourceID))
			}
			price := catalog.resourceByID[definition.Price.ResourceID]
			if output.Scope != price.Scope {
				return nil, catalogError("generator_classes", fmt.Errorf("%q crosses scopes from %q to %q", definition.ID, price.Scope, output.Scope))
			}
		}
		catalog.generators = append(catalog.generators, definition)
		catalog.generatorByID[definition.ID] = definition
	}

	return catalog, nil
}

func (c *Catalog) Resource(id string) (ResourceDefinition, bool) {
	definition, ok := c.resourceByID[id]
	return cloneResource(definition), ok
}

func (c *Catalog) Resources() []ResourceDefinition {
	definitions := make([]ResourceDefinition, len(c.resources))
	for index, definition := range c.resources {
		definitions[index] = cloneResource(definition)
	}
	return definitions
}

func (c *Catalog) GeneratorClass(id string) (GeneratorClassDefinition, bool) {
	definition, ok := c.generatorByID[id]
	return cloneGenerator(definition), ok
}

func (c *Catalog) GeneratorClasses() []GeneratorClassDefinition {
	definitions := make([]GeneratorClassDefinition, len(c.generators))
	for index, definition := range c.generators {
		definitions[index] = cloneGenerator(definition)
	}
	return definitions
}

func (c *Catalog) GeneratorClassesForScope(scope Scope) []GeneratorClassDefinition {
	definitions := make([]GeneratorClassDefinition, 0)
	for _, definition := range c.generators {
		if definition.Production == nil {
			continue
		}
		output := c.resourceByID[definition.Production.ResourceID]
		if output.Scope == scope {
			definitions = append(definitions, cloneGenerator(definition))
		}
	}
	return definitions
}

func validScope(scope Scope) bool {
	return scope == ScopeCompany || scope == ScopeFounder || scope == ScopeWorld || scope == ScopeGuild
}

func parseResource(source rawResource) (ResourceDefinition, error) {
	if !validID(source.ID) {
		return ResourceDefinition{}, fmt.Errorf("invalid id %q", source.ID)
	}
	scope := Scope(source.Scope)
	if !validScope(scope) {
		return ResourceDefinition{}, fmt.Errorf("unsupported scope %q", source.Scope)
	}
	if source.NumericKind != "decimal" {
		return ResourceDefinition{}, fmt.Errorf("unsupported numeric_kind %q", source.NumericKind)
	}
	initial, err := parseCatalogDecimal(source.Initial)
	if err != nil {
		return ResourceDefinition{}, fmt.Errorf("initial: %w", err)
	}
	minimum, err := parseCatalogDecimal(source.Minimum)
	if err != nil {
		return ResourceDefinition{}, fmt.Errorf("minimum: %w", err)
	}
	if initial.Lt(minimum) {
		return ResourceDefinition{}, errors.New("initial is below minimum")
	}

	hardcap, err := parseHardcap(source.Hardcap)
	if err != nil {
		return ResourceDefinition{}, err
	}
	if hardcap != nil {
		if hardcap.Amount.Lt(minimum) {
			return ResourceDefinition{}, errors.New("hardcap is below minimum")
		}
		if initial.Gt(hardcap.Amount) {
			return ResourceDefinition{}, errors.New("initial exceeds hardcap")
		}
	}

	return ResourceDefinition{
		ID:          source.ID,
		Scope:       scope,
		NumericKind: source.NumericKind,
		Initial:     initial,
		Minimum:     minimum,
		Hardcap:     hardcap,
	}, nil
}

func parseHardcap(data json.RawMessage) (*Hardcap, error) {
	if data == nil {
		return nil, errors.New("hardcap field is required and must be an object or null")
	}
	if strings.TrimSpace(string(data)) == "null" {
		return nil, nil
	}
	var source rawHardcap
	if err := decodeStrict(data, &source); err != nil {
		return nil, fmt.Errorf("hardcap: %w", err)
	}
	amount, err := parseCatalogDecimal(source.Amount)
	if err != nil {
		return nil, fmt.Errorf("hardcap amount: %w", err)
	}
	if !validID(source.ReasonKey) {
		return nil, fmt.Errorf("invalid hardcap reason_key %q", source.ReasonKey)
	}
	return &Hardcap{Amount: amount, ReasonKey: source.ReasonKey}, nil
}

func parseGenerator(source rawGeneratorClass, schemaVersion int) (GeneratorClassDefinition, error) {
	if !validID(source.ID) {
		return GeneratorClassDefinition{}, fmt.Errorf("invalid id %q", source.ID)
	}
	if !validID(source.Price.ResourceID) {
		return GeneratorClassDefinition{}, fmt.Errorf("invalid price resource_id %q", source.Price.ResourceID)
	}
	base, err := parseCatalogDecimal(source.Price.Base)
	if err != nil || !base.Gt(decimal.Zero) {
		return GeneratorClassDefinition{}, errors.New("price base must be a positive canonical decimal")
	}
	curve, err := parseCurve(source.Price.Curve)
	if err != nil {
		return GeneratorClassDefinition{}, err
	}
	definition := GeneratorClassDefinition{
		ID: source.ID,
		Price: PriceDefinition{
			ResourceID: source.Price.ResourceID,
			Base:       base,
			Curve:      curve,
		},
	}
	if schemaVersion == 1 {
		if source.Production != nil {
			return GeneratorClassDefinition{}, errors.New("catalog version 1 forbids production")
		}
		return definition, nil
	}
	if source.Production == nil {
		return GeneratorClassDefinition{}, errors.New("production is required")
	}
	if !validID(source.Production.ResourceID) {
		return GeneratorClassDefinition{}, fmt.Errorf("invalid production resource_id %q", source.Production.ResourceID)
	}
	baseRate, err := parseCatalogDecimal(source.Production.BaseRate)
	if err != nil || !baseRate.Gt(decimal.Zero) {
		return GeneratorClassDefinition{}, errors.New("production base_rate must be a positive canonical decimal")
	}
	definition.Production = &ProductionDefinition{ResourceID: source.Production.ResourceID, BaseRate: baseRate}
	return definition, nil
}

func parseCurve(source rawCurve) (CostCurve, error) {
	switch CurveKind(source.Kind) {
	case CurveConstant:
		if source.Step != nil || source.Ratio != nil {
			return CostCurve{}, errors.New("constant curve forbids step and ratio")
		}
		return CostCurve{Kind: CurveConstant}, nil
	case CurveLinear:
		if source.Step == nil || source.Ratio != nil {
			return CostCurve{}, errors.New("linear curve requires step and forbids ratio")
		}
		step, err := parseCatalogDecimal(*source.Step)
		if err != nil || step.Lt(decimal.Zero) {
			return CostCurve{}, errors.New("linear step must be a non-negative canonical decimal")
		}
		return CostCurve{Kind: CurveLinear, Step: step}, nil
	case CurveGeometric:
		if source.Ratio == nil || source.Step != nil {
			return CostCurve{}, errors.New("geometric curve requires ratio and forbids step")
		}
		ratio, err := parseCatalogDecimal(*source.Ratio)
		if err != nil || ratio.Lt(decimal.One) {
			return CostCurve{}, errors.New("geometric ratio must be a canonical decimal greater than or equal to one")
		}
		return CostCurve{Kind: CurveGeometric, Ratio: ratio}, nil
	default:
		return CostCurve{}, fmt.Errorf("unsupported curve kind %q", source.Kind)
	}
}

func parseCatalogDecimal(source string) (decimal.Decimal, error) {
	value, err := decimal.ParseCanonical(source)
	if err != nil || !value.IsStateValue() {
		return decimal.NaN, fmt.Errorf("%q is not an RFC-0001 canonical state decimal", source)
	}
	return value, nil
}

func validID(value string) bool {
	return idPattern.MatchString(value)
}

func cloneResource(definition ResourceDefinition) ResourceDefinition {
	if definition.Hardcap != nil {
		hardcap := *definition.Hardcap
		definition.Hardcap = &hardcap
	}
	return definition
}

func cloneGenerator(definition GeneratorClassDefinition) GeneratorClassDefinition {
	if definition.Production != nil {
		production := *definition.Production
		definition.Production = &production
	}
	return definition
}

func decodeStrict(data []byte, destination any) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		if err == nil {
			return errors.New("multiple JSON values")
		}
		return err
	}
	return nil
}

func catalogError(path string, err error) error {
	return fmt.Errorf("%w: %s: %v", ErrInvalidCatalog, path, err)
}

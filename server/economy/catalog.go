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
	"cloud-clicker/server/multiplier"
)

const CatalogSchemaVersion = 3

var (
	ErrInvalidCatalog        = errors.New("invalid economy catalog")
	idPattern                = regexp.MustCompile(`^[a-z][a-z0-9_]*(?:\.[a-z][a-z0-9_]*)*$`)
	minimumResourceLogTarget = decimal.New(5, -15)
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

type MultiplierSlot = multiplier.Slot

const (
	SlotUpgrades   = multiplier.SlotUpgrades
	SlotMilestones = multiplier.SlotMilestones
	SlotFaction    = multiplier.SlotFaction
	SlotDoctrine   = multiplier.SlotDoctrine
	SlotCommons    = multiplier.SlotCommons
	SlotTrust      = multiplier.SlotTrust
	SlotEventBuffs = multiplier.SlotEventBuffs
	SlotPrestige   = multiplier.SlotPrestige
)

var MultiplierSlotOrder = multiplier.Order

type ProgressKind string

const (
	ProgressResourceLog   ProgressKind = "resource_log"
	ProgressCountFraction ProgressKind = "count_fraction"
	ProgressComposite     ProgressKind = "composite"
)

const GeneratorTotalOwned = "generators.total_owned"

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

type ManualOutputDefinition struct {
	ResourceID      string
	AmountPerAction decimal.Decimal
}

type ManualActionDefinition struct {
	ID     string
	Output ManualOutputDefinition
}

type MultiplierSourceDefinition struct {
	ID       string
	Slot     MultiplierSlot
	Target   string
	Provider string
}

type ProgressTerm struct {
	Weight     decimal.Decimal
	Kind       ProgressKind
	ResourceID string
	Target     decimal.Decimal
	CountKey   string
	Required   int64
}

type ProgressCoordinateDefinition struct {
	Tier       int
	Kind       ProgressKind
	ResourceID string
	Target     decimal.Decimal
	CountKey   string
	Required   int64
	Terms      []ProgressTerm
}

type ManualPolicy struct {
	RefillMilliPerMS int64
	BucketCapMilli   int64
}

type OfflinePolicy struct {
	Efficiency           decimal.Decimal
	AccrualCapMS         int64
	BankRatioNumerator   int64
	BankRatioDenominator int64
	BankCapMS            int64
	BurstSpeed           decimal.Decimal
	BurstMaxDurationMS   int64
}

type Catalog struct {
	resources      []ResourceDefinition
	resourceByID   map[string]ResourceDefinition
	generators     []GeneratorClassDefinition
	generatorByID  map[string]GeneratorClassDefinition
	manualActions  []ManualActionDefinition
	manualByID     map[string]ManualActionDefinition
	multipliers    []MultiplierSourceDefinition
	multiplierByID map[string]MultiplierSourceDefinition
	progress       []ProgressCoordinateDefinition
	progressByTier map[int]ProgressCoordinateDefinition
	manualPolicy   ManualPolicy
	offlinePolicy  OfflinePolicy
}

type rawCatalog struct {
	SchemaVersion       int                     `json:"schema_version"`
	Resources           []rawResource           `json:"resources"`
	GeneratorClasses    []rawGeneratorClass     `json:"generator_classes"`
	ManualActions       []rawManualAction       `json:"manual_actions"`
	MultiplierSources   []rawMultiplierSource   `json:"multiplier_sources"`
	ProgressCoordinates []rawProgressCoordinate `json:"progress_coordinates"`
	ManualPolicy        *rawManualPolicy        `json:"manual_policy"`
	OfflinePolicy       *rawOfflinePolicy       `json:"offline_policy"`
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

type rawManualAction struct {
	ID     string          `json:"id"`
	Output rawManualOutput `json:"output"`
}

type rawManualOutput struct {
	ResourceID      string `json:"resource_id"`
	AmountPerAction string `json:"amount_per_action"`
}

type rawMultiplierSource struct {
	ID       string `json:"id"`
	Slot     string `json:"slot"`
	Target   string `json:"target"`
	Provider string `json:"provider"`
}

type rawProgressCoordinate struct {
	Tier       int               `json:"tier"`
	Kind       string            `json:"kind"`
	ResourceID string            `json:"resource"`
	Target     string            `json:"target"`
	CountKey   string            `json:"count"`
	Required   int64             `json:"required"`
	Terms      []rawProgressTerm `json:"terms"`
}

type rawProgressTerm struct {
	Weight     string `json:"weight"`
	Kind       string `json:"kind"`
	ResourceID string `json:"resource"`
	Target     string `json:"target"`
	CountKey   string `json:"count"`
	Required   int64  `json:"required"`
}

type rawManualPolicy struct {
	RefillMilliPerMS int64 `json:"refill_milli_per_ms"`
	BucketCapMilli   int64 `json:"bucket_cap_milli"`
}

type rawOfflinePolicy struct {
	Efficiency           string `json:"efficiency"`
	AccrualCapMS         int64  `json:"accrual_cap_ms"`
	BankRatioNumerator   int64  `json:"bank_ratio_numerator"`
	BankRatioDenominator int64  `json:"bank_ratio_denominator"`
	BankCapMS            int64  `json:"bank_cap_ms"`
	BurstSpeed           string `json:"burst_speed"`
	BurstMaxDurationMS   int64  `json:"burst_max_duration_ms"`
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
	if raw.SchemaVersion < 1 || raw.SchemaVersion > CatalogSchemaVersion {
		return nil, catalogError("schema_version", fmt.Errorf("got %d, supported versions are 1 through %d", raw.SchemaVersion, CatalogSchemaVersion))
	}
	if raw.Resources == nil {
		return nil, catalogError("resources", errors.New("field is required"))
	}
	if raw.GeneratorClasses == nil {
		return nil, catalogError("generator_classes", errors.New("field is required"))
	}

	catalog := &Catalog{
		resources:      make([]ResourceDefinition, 0, len(raw.Resources)),
		resourceByID:   make(map[string]ResourceDefinition, len(raw.Resources)),
		generators:     make([]GeneratorClassDefinition, 0, len(raw.GeneratorClasses)),
		generatorByID:  make(map[string]GeneratorClassDefinition, len(raw.GeneratorClasses)),
		manualActions:  make([]ManualActionDefinition, 0, len(raw.ManualActions)),
		manualByID:     make(map[string]ManualActionDefinition, len(raw.ManualActions)),
		multipliers:    make([]MultiplierSourceDefinition, 0, len(raw.MultiplierSources)),
		multiplierByID: make(map[string]MultiplierSourceDefinition, len(raw.MultiplierSources)),
		progress:       make([]ProgressCoordinateDefinition, 0, len(raw.ProgressCoordinates)),
		progressByTier: make(map[int]ProgressCoordinateDefinition, len(raw.ProgressCoordinates)),
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

	if raw.SchemaVersion < 3 {
		if raw.ManualActions != nil || raw.MultiplierSources != nil || raw.ProgressCoordinates != nil || raw.ManualPolicy != nil || raw.OfflinePolicy != nil {
			return nil, catalogError("schema_version", errors.New("catalog versions 1 and 2 forbid production-engine fields"))
		}
		return catalog, nil
	}
	if raw.ManualActions == nil || raw.MultiplierSources == nil || raw.ProgressCoordinates == nil || raw.ManualPolicy == nil || raw.OfflinePolicy == nil {
		return nil, catalogError("production_engine", errors.New("manual_actions, multiplier_sources, progress_coordinates, manual_policy, and offline_policy are required"))
	}

	for index, source := range raw.ManualActions {
		definition, err := parseManualAction(source)
		if err != nil {
			return nil, catalogError(fmt.Sprintf("manual_actions[%d]", index), err)
		}
		if _, exists := catalog.manualByID[definition.ID]; exists {
			return nil, catalogError("manual_actions", fmt.Errorf("duplicate id %q", definition.ID))
		}
		resource, exists := catalog.resourceByID[definition.Output.ResourceID]
		if !exists || resource.Scope != ScopeCompany {
			return nil, catalogError("manual_actions", fmt.Errorf("%q output must reference a company resource", definition.ID))
		}
		catalog.manualActions = append(catalog.manualActions, definition)
		catalog.manualByID[definition.ID] = definition
	}

	singleProviderSlots := make(map[MultiplierSlot]string)
	for index, source := range raw.MultiplierSources {
		definition, err := parseMultiplierSource(source)
		if err != nil {
			return nil, catalogError(fmt.Sprintf("multiplier_sources[%d]", index), err)
		}
		if _, exists := catalog.multiplierByID[definition.ID]; exists {
			return nil, catalogError("multiplier_sources", fmt.Errorf("duplicate id %q", definition.ID))
		}
		if definition.Target != "all" {
			if _, exists := catalog.generatorByID[definition.Target]; !exists {
				return nil, catalogError("multiplier_sources", fmt.Errorf("%q references unknown target %q", definition.ID, definition.Target))
			}
		}
		if definition.Slot == SlotCommons || definition.Slot == SlotTrust {
			if prior, exists := singleProviderSlots[definition.Slot]; exists {
				return nil, catalogError("multiplier_sources", fmt.Errorf("slot %q is single-provider (got %q and %q)", definition.Slot, prior, definition.ID))
			}
			singleProviderSlots[definition.Slot] = definition.ID
		}
		catalog.multipliers = append(catalog.multipliers, definition)
		catalog.multiplierByID[definition.ID] = definition
	}

	for index, source := range raw.ProgressCoordinates {
		definition, err := parseProgressCoordinate(source, catalog.resourceByID)
		if err != nil {
			return nil, catalogError(fmt.Sprintf("progress_coordinates[%d]", index), err)
		}
		if _, exists := catalog.progressByTier[definition.Tier]; exists {
			return nil, catalogError("progress_coordinates", fmt.Errorf("duplicate tier %d", definition.Tier))
		}
		catalog.progress = append(catalog.progress, definition)
		catalog.progressByTier[definition.Tier] = definition
	}
	for tier := 0; tier <= 3; tier++ {
		if _, exists := catalog.progressByTier[tier]; !exists {
			return nil, catalogError("progress_coordinates", fmt.Errorf("tier %d is required", tier))
		}
	}

	manualPolicy, err := parseManualPolicy(*raw.ManualPolicy)
	if err != nil {
		return nil, catalogError("manual_policy", err)
	}
	offlinePolicy, err := parseOfflinePolicy(*raw.OfflinePolicy)
	if err != nil {
		return nil, catalogError("offline_policy", err)
	}
	catalog.manualPolicy = manualPolicy
	catalog.offlinePolicy = offlinePolicy

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

func (c *Catalog) ManualAction(id string) (ManualActionDefinition, bool) {
	definition, ok := c.manualByID[id]
	return definition, ok
}

func (c *Catalog) ManualActions() []ManualActionDefinition {
	return append([]ManualActionDefinition(nil), c.manualActions...)
}

func (c *Catalog) MultiplierSource(id string) (MultiplierSourceDefinition, bool) {
	definition, ok := c.multiplierByID[id]
	return definition, ok
}

func (c *Catalog) MultiplierSources() []MultiplierSourceDefinition {
	return append([]MultiplierSourceDefinition(nil), c.multipliers...)
}

func (c *Catalog) ProgressCoordinate(tier int) (ProgressCoordinateDefinition, bool) {
	definition, ok := c.progressByTier[tier]
	return cloneProgress(definition), ok
}

func (c *Catalog) ProgressCoordinates() []ProgressCoordinateDefinition {
	definitions := make([]ProgressCoordinateDefinition, len(c.progress))
	for index, definition := range c.progress {
		definitions[index] = cloneProgress(definition)
	}
	return definitions
}

func (c *Catalog) ManualPolicy() ManualPolicy { return c.manualPolicy }

func (c *Catalog) OfflinePolicy() OfflinePolicy { return c.offlinePolicy }

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

func parseManualAction(source rawManualAction) (ManualActionDefinition, error) {
	if !validID(source.ID) || !validID(source.Output.ResourceID) {
		return ManualActionDefinition{}, errors.New("id and output resource_id must be mechanical ids")
	}
	amount, err := parseCatalogDecimal(source.Output.AmountPerAction)
	if err != nil || !amount.Gt(decimal.Zero) {
		return ManualActionDefinition{}, errors.New("amount_per_action must be a positive canonical decimal")
	}
	return ManualActionDefinition{ID: source.ID, Output: ManualOutputDefinition{
		ResourceID: source.Output.ResourceID, AmountPerAction: amount,
	}}, nil
}

func parseMultiplierSource(source rawMultiplierSource) (MultiplierSourceDefinition, error) {
	if !validID(source.ID) || !validID(source.Provider) {
		return MultiplierSourceDefinition{}, errors.New("id and provider must be mechanical ids")
	}
	slot := MultiplierSlot(source.Slot)
	if !validMultiplierSlot(slot) {
		return MultiplierSourceDefinition{}, fmt.Errorf("unsupported slot %q", source.Slot)
	}
	if source.Target != "all" && !validID(source.Target) {
		return MultiplierSourceDefinition{}, fmt.Errorf("invalid target %q", source.Target)
	}
	return MultiplierSourceDefinition{ID: source.ID, Slot: slot, Target: source.Target, Provider: source.Provider}, nil
}

func validMultiplierSlot(slot MultiplierSlot) bool {
	return multiplier.ValidSlot(slot)
}

func parseProgressCoordinate(source rawProgressCoordinate, resources map[string]ResourceDefinition) (ProgressCoordinateDefinition, error) {
	if source.Tier < 0 || source.Tier > 3 {
		return ProgressCoordinateDefinition{}, errors.New("tier must be in the in-scope range 0..3")
	}
	definition := ProgressCoordinateDefinition{Tier: source.Tier, Kind: ProgressKind(source.Kind)}
	switch definition.Kind {
	case ProgressResourceLog:
		term, err := parseProgressTerm(rawProgressTerm{Kind: source.Kind, ResourceID: source.ResourceID, Target: source.Target}, resources, false)
		if err != nil || source.CountKey != "" || source.Required != 0 || source.Terms != nil {
			if err == nil {
				err = errors.New("resource_log forbids count, required, and terms")
			}
			return ProgressCoordinateDefinition{}, err
		}
		definition.ResourceID, definition.Target = term.ResourceID, term.Target
	case ProgressCountFraction:
		term, err := parseProgressTerm(rawProgressTerm{Kind: source.Kind, CountKey: source.CountKey, Required: source.Required}, resources, false)
		if err != nil || source.ResourceID != "" || source.Target != "" || source.Terms != nil {
			if err == nil {
				err = errors.New("count_fraction forbids resource, target, and terms")
			}
			return ProgressCoordinateDefinition{}, err
		}
		definition.CountKey, definition.Required = term.CountKey, term.Required
	case ProgressComposite:
		if source.ResourceID != "" || source.Target != "" || source.CountKey != "" || source.Required != 0 || len(source.Terms) == 0 {
			return ProgressCoordinateDefinition{}, errors.New("composite requires terms and forbids direct resource/count fields")
		}
		weights := make([]decimal.Decimal, 0, len(source.Terms))
		definition.Terms = make([]ProgressTerm, 0, len(source.Terms))
		for _, rawTerm := range source.Terms {
			term, err := parseProgressTerm(rawTerm, resources, true)
			if err != nil {
				return ProgressCoordinateDefinition{}, err
			}
			definition.Terms = append(definition.Terms, term)
			weights = append(weights, term.Weight)
		}
		if !decimal.SumDeterministic(weights).Eq(decimal.One) {
			return ProgressCoordinateDefinition{}, errors.New("composite weights must sum exactly to 1e0")
		}
	default:
		return ProgressCoordinateDefinition{}, fmt.Errorf("unsupported progress kind %q", source.Kind)
	}
	return definition, nil
}

func parseProgressTerm(source rawProgressTerm, resources map[string]ResourceDefinition, weighted bool) (ProgressTerm, error) {
	term := ProgressTerm{Kind: ProgressKind(source.Kind)}
	if weighted {
		weight, err := parseCatalogDecimal(source.Weight)
		if err != nil || !weight.Gt(decimal.Zero) || weight.Gt(decimal.One) {
			return ProgressTerm{}, errors.New("term weight must be a canonical Decimal in (0,1]")
		}
		term.Weight = weight
	} else if source.Weight != "" {
		return ProgressTerm{}, errors.New("non-composite progress forbids weight")
	}
	switch term.Kind {
	case ProgressResourceLog:
		resource, exists := resources[source.ResourceID]
		if !exists || resource.Scope != ScopeCompany || source.CountKey != "" || source.Required != 0 {
			return ProgressTerm{}, errors.New("resource_log requires a company resource and forbids count fields")
		}
		target, err := parseCatalogDecimal(source.Target)
		if err != nil || target.Lt(minimumResourceLogTarget) {
			return ProgressTerm{}, errors.New("resource_log target must be a canonical decimal greater than or equal to 5e-15")
		}
		denominator := decimal.One.Add(target).Log10()
		if !denominator.IsStateValue() || !denominator.Gt(decimal.Zero) {
			return ProgressTerm{}, errors.New("resource_log target must produce a finite positive logarithm")
		}
		term.ResourceID, term.Target = source.ResourceID, target
	case ProgressCountFraction:
		if source.CountKey != GeneratorTotalOwned || source.Required <= 0 || source.Required > decimal.MaxExactInteger || source.ResourceID != "" || source.Target != "" {
			return ProgressTerm{}, errors.New("count_fraction requires generators.total_owned and a positive safe-integer required value")
		}
		term.CountKey, term.Required = source.CountKey, source.Required
	default:
		return ProgressTerm{}, fmt.Errorf("unsupported progress term kind %q", source.Kind)
	}
	return term, nil
}

func parseManualPolicy(source rawManualPolicy) (ManualPolicy, error) {
	if source.RefillMilliPerMS <= 0 || source.BucketCapMilli <= 0 || source.RefillMilliPerMS > decimal.MaxExactInteger || source.BucketCapMilli > decimal.MaxExactInteger {
		return ManualPolicy{}, errors.New("manual policy values must be positive safe integers")
	}
	return ManualPolicy{RefillMilliPerMS: source.RefillMilliPerMS, BucketCapMilli: source.BucketCapMilli}, nil
}

func parseOfflinePolicy(source rawOfflinePolicy) (OfflinePolicy, error) {
	efficiency, err := parseCatalogDecimal(source.Efficiency)
	if err != nil || efficiency.Lt(decimal.Zero) || efficiency.Gt(decimal.One) {
		return OfflinePolicy{}, errors.New("efficiency must be a canonical Decimal in [0,1]")
	}
	burstSpeed, err := parseCatalogDecimal(source.BurstSpeed)
	if err != nil || !burstSpeed.Gt(decimal.Zero) {
		return OfflinePolicy{}, errors.New("burst_speed must be a positive canonical Decimal")
	}
	values := []int64{source.AccrualCapMS, source.BankRatioNumerator, source.BankRatioDenominator, source.BankCapMS, source.BurstMaxDurationMS}
	for _, value := range values {
		if value <= 0 || value > decimal.MaxExactInteger {
			return OfflinePolicy{}, errors.New("offline integer fields must be positive safe integers")
		}
	}
	if source.BankRatioNumerator > source.BankRatioDenominator {
		return OfflinePolicy{}, errors.New("bank ratio may not exceed one")
	}
	return OfflinePolicy{
		Efficiency: efficiency, AccrualCapMS: source.AccrualCapMS,
		BankRatioNumerator: source.BankRatioNumerator, BankRatioDenominator: source.BankRatioDenominator,
		BankCapMS: source.BankCapMS, BurstSpeed: burstSpeed, BurstMaxDurationMS: source.BurstMaxDurationMS,
	}, nil
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

func cloneProgress(definition ProgressCoordinateDefinition) ProgressCoordinateDefinition {
	definition.Terms = append([]ProgressTerm(nil), definition.Terms...)
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

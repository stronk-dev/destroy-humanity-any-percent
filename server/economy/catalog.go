// Package economy implements the configuration-driven economy kernel.
package economy

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"

	"cloud-clicker/server/decimal"
	"cloud-clicker/server/multiplier"
	"cloud-clicker/server/routes"
)

const CatalogSchemaVersion = 4

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
	ID                 string
	Tier               int
	Category           string
	Price              PriceDefinition
	Production         *ProductionDefinition
	Provision          *ProvisionDefinition
	ProvisionedHardcap *ProvisionedHardcap
	Ladder             []LadderRung
	Roles              []GeneratorRole
}

type ProvisionDefinition struct {
	GeneratorID string
	RatePPM     int64
}

type ProvisionedHardcap struct {
	Count     int64
	ReasonKey string
}

type LadderRung struct {
	PurchasedAt   int64
	MultiplierPPM int64
}

type GeneratorRoleKind string

const (
	RoleProvision    GeneratorRoleKind = "provision"
	RoleSynergyFeed  GeneratorRoleKind = "synergy_feed"
	RoleManualOutput GeneratorRoleKind = "manual_output"
	RoleStockRate    GeneratorRoleKind = "stock_rate"
)

type GeneratorRole struct {
	Kind            GeneratorRoleKind
	GeneratorID     string
	PoolID          string
	ActionID        string
	PerPurchasedPPM int64
}

type AvailabilityWindow struct {
	FromGate string
	ToGate   string
}

type UpgradeEffect struct {
	SourceID string
	Slot     MultiplierSlot
	Target   string
	Factor   decimal.Decimal
}

type UpgradeDefinition struct {
	ID       string
	Cost     UpgradeCost
	Window   AvailabilityWindow
	Requires []routes.Condition
	Effects  []UpgradeEffect
	Roles    []string
	CopyKey  string
}

type UpgradeCost struct {
	ResourceID string
	Amount     decimal.Decimal
}

type SynergyCurve string

const (
	SynergyLinear SynergyCurve = "linear"
	SynergyLog    SynergyCurve = "log"
)

type SynergySourceKind string

const (
	SynergyGenerator SynergySourceKind = "generator"
	SynergyUpgrade   SynergySourceKind = "upgrade"
)

type SynergySource struct {
	Kind        SynergySourceKind
	ID          string
	PerCountPPM int64
}

type SynergyPoolDefinition struct {
	ID      string
	Sources []SynergySource
	Slot    MultiplierSlot
	Curve   SynergyCurve
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
	resources       []ResourceDefinition
	resourceByID    map[string]ResourceDefinition
	generators      []GeneratorClassDefinition
	generatorByID   map[string]GeneratorClassDefinition
	upgrades        []UpgradeDefinition
	upgradeByID     map[string]UpgradeDefinition
	synergyPools    []SynergyPoolDefinition
	synergyByID     map[string]SynergyPoolDefinition
	manualActions   []ManualActionDefinition
	manualByID      map[string]ManualActionDefinition
	multipliers     []MultiplierSourceDefinition
	multiplierByID  map[string]MultiplierSourceDefinition
	progress        []ProgressCoordinateDefinition
	progressByTier  map[int]ProgressCoordinateDefinition
	manualPolicy    ManualPolicy
	offlinePolicy   OfflinePolicy
	provisionTickMS int64
}

type rawCatalog struct {
	SchemaVersion       int                     `json:"schema_version"`
	Resources           []rawResource           `json:"resources"`
	GeneratorClasses    []rawGeneratorClass     `json:"generator_classes"`
	Upgrades            []rawUpgrade            `json:"upgrades"`
	SynergyPools        []rawSynergyPool        `json:"synergy_pools"`
	ProvisionTickMS     int64                   `json:"provision_tick_ms"`
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
	ID                 string                 `json:"id"`
	Tier               *int                   `json:"tier"`
	Category           *string                `json:"category"`
	Price              rawPrice               `json:"price"`
	Production         *rawProduction         `json:"production"`
	Provisions         *rawProvision          `json:"provisions"`
	ProvisionedHardcap *rawProvisionedHardcap `json:"provisioned_hardcap"`
	Ladder             []rawLadderRung        `json:"ladder"`
	Roles              []rawGeneratorRole     `json:"roles"`
}

type rawProvision struct {
	GeneratorID string `json:"generator_id"`
	RatePPM     int64  `json:"rate_ppm"`
}

type rawProvisionedHardcap struct {
	Count     int64  `json:"count"`
	ReasonKey string `json:"reason_key"`
}

type rawLadderRung struct {
	PurchasedAt   int64 `json:"purchased_at"`
	MultiplierPPM int64 `json:"multiplier_ppm"`
}

type rawGeneratorRole struct {
	Kind            string  `json:"kind"`
	GeneratorID     *string `json:"generator_id"`
	PoolID          *string `json:"pool_id"`
	ActionID        *string `json:"action_id"`
	PerPurchasedPPM *int64  `json:"per_purchased_ppm"`
}

type rawUpgrade struct {
	ID       string             `json:"id"`
	Cost     rawUpgradeCost     `json:"cost"`
	Window   rawWindow          `json:"window"`
	Requires json.RawMessage    `json:"requires"`
	Effects  []rawUpgradeEffect `json:"effects"`
	Roles    []string           `json:"roles"`
	CopyKey  string             `json:"copy_key"`
}

type rawUpgradeCost struct {
	ResourceID string `json:"resource"`
	Amount     string `json:"amount"`
}

type rawWindow struct {
	FromGate json.RawMessage `json:"from_gate"`
	ToGate   json.RawMessage `json:"to_gate"`
}

type rawUpgradeEffect struct {
	SourceID string `json:"source_id"`
	Slot     string `json:"slot"`
	Target   string `json:"target"`
	Factor   string `json:"factor"`
}

type rawSynergyPool struct {
	ID      string             `json:"id"`
	Sources []rawSynergySource `json:"sources"`
	Slot    string             `json:"slot"`
	Curve   string             `json:"curve"`
}

type rawSynergySource struct {
	Kind        string `json:"kind"`
	ID          string `json:"id_or_class"`
	PerCountPPM int64  `json:"per_count_ppm"`
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
		upgrades:       make([]UpgradeDefinition, 0, len(raw.Upgrades)),
		upgradeByID:    make(map[string]UpgradeDefinition, len(raw.Upgrades)),
		synergyPools:   make([]SynergyPoolDefinition, 0, len(raw.SynergyPools)),
		synergyByID:    make(map[string]SynergyPoolDefinition, len(raw.SynergyPools)),
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
		if raw.ManualActions != nil || raw.MultiplierSources != nil || raw.ProgressCoordinates != nil || raw.ManualPolicy != nil || raw.OfflinePolicy != nil ||
			raw.Upgrades != nil || raw.SynergyPools != nil || raw.ProvisionTickMS != 0 {
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

	if raw.SchemaVersion == 3 {
		if raw.Upgrades != nil || raw.SynergyPools != nil || raw.ProvisionTickMS != 0 {
			return nil, catalogError("schema_version", errors.New("catalog version 3 forbids purchasable-content fields"))
		}
		return catalog, nil
	}
	if raw.Upgrades == nil || raw.SynergyPools == nil || raw.ProvisionTickMS <= 0 || raw.ProvisionTickMS > decimal.MaxExactInteger {
		return nil, catalogError("purchasable_content", errors.New("upgrades, synergy_pools, and a positive safe provision_tick_ms are required"))
	}
	catalog.provisionTickMS = raw.ProvisionTickMS

	for index, source := range raw.Upgrades {
		definition, err := parseUpgrade(source, catalog.resourceByID)
		if err != nil {
			return nil, catalogError(fmt.Sprintf("upgrades[%d]", index), err)
		}
		if _, duplicate := catalog.upgradeByID[definition.ID]; duplicate {
			return nil, catalogError("upgrades", fmt.Errorf("duplicate id %q", definition.ID))
		}
		for _, effect := range definition.Effects {
			if _, duplicate := catalog.multiplierByID[effect.SourceID]; duplicate {
				return nil, catalogError("upgrades", fmt.Errorf("duplicate contribution source %q", effect.SourceID))
			}
			if _, generator := catalog.generatorByID[effect.Target]; !generator {
				if _, manual := catalog.manualByID[effect.Target]; !manual {
					return nil, catalogError("upgrades", fmt.Errorf("%q effect references unknown target %q", definition.ID, effect.Target))
				}
			}
			declaration := MultiplierSourceDefinition{ID: effect.SourceID, Slot: effect.Slot, Target: effect.Target, Provider: definition.ID}
			catalog.multipliers = append(catalog.multipliers, declaration)
			catalog.multiplierByID[declaration.ID] = declaration
		}
		catalog.upgrades = append(catalog.upgrades, definition)
		catalog.upgradeByID[definition.ID] = definition
	}

	for index, source := range raw.SynergyPools {
		definition, err := parseSynergyPool(source, catalog.generatorByID, catalog.upgradeByID)
		if err != nil {
			return nil, catalogError(fmt.Sprintf("synergy_pools[%d]", index), err)
		}
		if _, duplicate := catalog.synergyByID[definition.ID]; duplicate {
			return nil, catalogError("synergy_pools", fmt.Errorf("duplicate id %q", definition.ID))
		}
		declaration, declared := catalog.multiplierByID[definition.ID]
		if !declared || declaration.Slot != definition.Slot || declaration.Provider != definition.ID {
			return nil, catalogError("synergy_pools", fmt.Errorf("%q must map to one matching multiplier source", definition.ID))
		}
		catalog.synergyPools = append(catalog.synergyPools, definition)
		catalog.synergyByID[definition.ID] = definition
	}
	for _, generator := range catalog.generators {
		for _, rung := range generator.Ladder {
			declaration := MultiplierSourceDefinition{ID: LadderSourceID(generator.ID, rung.PurchasedAt), Slot: SlotMilestones, Target: generator.ID, Provider: generator.ID}
			if _, duplicate := catalog.multiplierByID[declaration.ID]; duplicate {
				return nil, catalogError("generator_classes", fmt.Errorf("duplicate ladder contribution source %q", declaration.ID))
			}
			catalog.multipliers = append(catalog.multipliers, declaration)
			catalog.multiplierByID[declaration.ID] = declaration
		}
		for _, role := range generator.Roles {
			if role.Kind != RoleManualOutput {
				continue
			}
			declaration := MultiplierSourceDefinition{ID: ManualRoleSourceID(generator.ID, role.ActionID), Slot: SlotUpgrades, Target: role.ActionID, Provider: generator.ID}
			if _, duplicate := catalog.multiplierByID[declaration.ID]; duplicate {
				return nil, catalogError("generator_classes", fmt.Errorf("duplicate manual role contribution source %q", declaration.ID))
			}
			catalog.multipliers = append(catalog.multipliers, declaration)
			catalog.multiplierByID[declaration.ID] = declaration
		}
	}
	if err := catalog.validateGeneratorContent(); err != nil {
		return nil, catalogError("generator_classes", err)
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

func (c *Catalog) Upgrade(id string) (UpgradeDefinition, bool) {
	definition, ok := c.upgradeByID[id]
	return cloneUpgrade(definition), ok
}

func (c *Catalog) Upgrades() []UpgradeDefinition {
	definitions := make([]UpgradeDefinition, len(c.upgrades))
	for index, definition := range c.upgrades {
		definitions[index] = cloneUpgrade(definition)
	}
	return definitions
}

func (c *Catalog) SynergyPool(id string) (SynergyPoolDefinition, bool) {
	definition, ok := c.synergyByID[id]
	return cloneSynergyPool(definition), ok
}

func (c *Catalog) SynergyPools() []SynergyPoolDefinition {
	definitions := make([]SynergyPoolDefinition, len(c.synergyPools))
	for index, definition := range c.synergyPools {
		definitions[index] = cloneSynergyPool(definition)
	}
	return definitions
}

func (c *Catalog) ProvisionTickMS() int64 { return c.provisionTickMS }

func LadderSourceID(generatorID string, purchasedAt int64) string {
	return generatorID + ".ladder.purchased_" + strconv.FormatInt(purchasedAt, 10)
}

func ManualRoleSourceID(generatorID, actionID string) string {
	return generatorID + ".role.manual_output." + actionID
}

func (c *Catalog) ValidateGateReferences(gateIDs []string) error {
	known := make(map[string]bool, len(gateIDs))
	for _, id := range gateIDs {
		if !validID(id) || known[id] {
			return fmt.Errorf("%w: invalid or duplicate route gate id %q", ErrInvalidCatalog, id)
		}
		known[id] = true
	}
	for _, upgrade := range c.upgrades {
		for _, id := range []string{upgrade.Window.FromGate, upgrade.Window.ToGate} {
			if id != "" && !known[id] {
				return fmt.Errorf("%w: upgrade %q references unknown gate %q", ErrInvalidCatalog, upgrade.ID, id)
			}
		}
		for _, condition := range upgrade.Requires {
			if condition.Kind == routes.ConditionResourceAtLeast || condition.Kind == routes.ConditionResourceAtMost {
				resource, exists := c.resourceByID[condition.ResourceID]
				if !exists || resource.Scope != ScopeCompany {
					return fmt.Errorf("%w: upgrade %q references non-company resource %q", ErrInvalidCatalog, upgrade.ID, condition.ResourceID)
				}
			}
		}
	}
	return nil
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
	if schemaVersion < 4 {
		if source.Tier != nil || source.Category != nil || source.Provisions != nil || source.ProvisionedHardcap != nil || source.Ladder != nil || source.Roles != nil {
			return GeneratorClassDefinition{}, errors.New("catalog versions before 4 forbid purchasable-content fields")
		}
		return definition, nil
	}
	if source.Tier == nil || *source.Tier < 0 || *source.Tier > 3 || source.Category == nil || !validID(*source.Category) ||
		source.Ladder == nil || len(source.Roles) == 0 {
		return GeneratorClassDefinition{}, errors.New("tier, category, ladder, and at least one role are required")
	}
	definition.Tier, definition.Category = *source.Tier, *source.Category
	if source.Provisions != nil {
		if !validID(source.Provisions.GeneratorID) || source.Provisions.RatePPM <= 0 || source.Provisions.RatePPM > decimal.MaxExactInteger {
			return GeneratorClassDefinition{}, errors.New("provisions require a target id and positive safe rate_ppm")
		}
		definition.Provision = &ProvisionDefinition{GeneratorID: source.Provisions.GeneratorID, RatePPM: source.Provisions.RatePPM}
	}
	if source.ProvisionedHardcap != nil {
		if source.ProvisionedHardcap.Count <= 0 || source.ProvisionedHardcap.Count > decimal.MaxExactInteger || !validID(source.ProvisionedHardcap.ReasonKey) {
			return GeneratorClassDefinition{}, errors.New("provisioned_hardcap requires a positive safe count and reason_key")
		}
		definition.ProvisionedHardcap = &ProvisionedHardcap{Count: source.ProvisionedHardcap.Count, ReasonKey: source.ProvisionedHardcap.ReasonKey}
	}
	prior := int64(0)
	for index, sourceRung := range source.Ladder {
		if sourceRung.PurchasedAt <= prior || sourceRung.PurchasedAt > decimal.MaxExactInteger || sourceRung.MultiplierPPM <= 0 || sourceRung.MultiplierPPM == 1_000_000 || sourceRung.MultiplierPPM > decimal.MaxExactInteger {
			return GeneratorClassDefinition{}, fmt.Errorf("ladder[%d] requires increasing purchased_at and positive, non-neutral safe multiplier_ppm", index)
		}
		definition.Ladder = append(definition.Ladder, LadderRung{PurchasedAt: sourceRung.PurchasedAt, MultiplierPPM: sourceRung.MultiplierPPM})
		prior = sourceRung.PurchasedAt
	}
	seenRoles := map[string]bool{}
	for index, sourceRole := range source.Roles {
		role, key, err := parseGeneratorRole(sourceRole)
		if err != nil || seenRoles[key] {
			if err == nil {
				err = errors.New("duplicate role kind/target")
			}
			return GeneratorClassDefinition{}, fmt.Errorf("roles[%d]: %v", index, err)
		}
		seenRoles[key] = true
		definition.Roles = append(definition.Roles, role)
	}
	return definition, nil
}

func parseGeneratorRole(source rawGeneratorRole) (GeneratorRole, string, error) {
	role := GeneratorRole{Kind: GeneratorRoleKind(source.Kind)}
	switch role.Kind {
	case RoleProvision:
		if source.GeneratorID == nil || !validID(*source.GeneratorID) || source.PoolID != nil || source.ActionID != nil || source.PerPurchasedPPM != nil {
			return GeneratorRole{}, "", errors.New("provision requires exactly generator_id")
		}
		role.GeneratorID = *source.GeneratorID
		return role, string(role.Kind) + ":" + role.GeneratorID, nil
	case RoleSynergyFeed:
		if source.PoolID == nil || !validID(*source.PoolID) || source.GeneratorID != nil || source.ActionID != nil || source.PerPurchasedPPM != nil {
			return GeneratorRole{}, "", errors.New("synergy_feed requires exactly pool_id")
		}
		role.PoolID = *source.PoolID
		return role, string(role.Kind) + ":" + role.PoolID, nil
	case RoleManualOutput:
		if source.ActionID == nil || !validID(*source.ActionID) || source.PerPurchasedPPM == nil || *source.PerPurchasedPPM <= 0 || *source.PerPurchasedPPM > decimal.MaxExactInteger ||
			source.GeneratorID != nil || source.PoolID != nil {
			return GeneratorRole{}, "", errors.New("manual_output requires exactly action_id and positive safe per_purchased_ppm")
		}
		role.ActionID, role.PerPurchasedPPM = *source.ActionID, *source.PerPurchasedPPM
		return role, string(role.Kind) + ":" + role.ActionID, nil
	case RoleStockRate:
		if source.PerPurchasedPPM == nil || *source.PerPurchasedPPM <= 0 || *source.PerPurchasedPPM > decimal.MaxExactInteger ||
			source.GeneratorID != nil || source.PoolID != nil || source.ActionID != nil {
			return GeneratorRole{}, "", errors.New("stock_rate requires exactly positive safe per_purchased_ppm")
		}
		role.PerPurchasedPPM = *source.PerPurchasedPPM
		return role, string(role.Kind), nil
	default:
		return GeneratorRole{}, "", fmt.Errorf("unsupported role kind %q", source.Kind)
	}
}

func parseUpgrade(source rawUpgrade, resources map[string]ResourceDefinition) (UpgradeDefinition, error) {
	if !validID(source.ID) || !validID(source.Cost.ResourceID) || !validID(source.CopyKey) || len(source.Effects) == 0 || source.Roles == nil {
		return UpgradeDefinition{}, errors.New("id, cost resource, copy_key, effects, and roles are required")
	}
	resource, exists := resources[source.Cost.ResourceID]
	if !exists || resource.Scope != ScopeCompany {
		return UpgradeDefinition{}, errors.New("cost must reference a company resource")
	}
	amount, err := parseCatalogDecimal(source.Cost.Amount)
	if err != nil || !amount.Gt(decimal.Zero) {
		return UpgradeDefinition{}, errors.New("cost amount must be a positive canonical Decimal")
	}
	window, err := parseWindow(source.Window)
	if err != nil {
		return UpgradeDefinition{}, err
	}
	requires, err := routes.ParseConditions(source.Requires)
	if err != nil {
		return UpgradeDefinition{}, fmt.Errorf("requires: %v", err)
	}
	definition := UpgradeDefinition{ID: source.ID, Cost: UpgradeCost{ResourceID: source.Cost.ResourceID, Amount: amount}, Window: window, Requires: requires, CopyKey: source.CopyKey}
	seenEffects := map[string]bool{}
	for index, sourceEffect := range source.Effects {
		if !validID(sourceEffect.SourceID) || !validID(sourceEffect.Target) || sourceEffect.Slot != string(SlotUpgrades) || seenEffects[sourceEffect.SourceID] {
			return UpgradeDefinition{}, fmt.Errorf("effects[%d] requires a unique source_id, upgrades slot, and mechanical target", index)
		}
		factor, err := parseCatalogDecimal(sourceEffect.Factor)
		if err != nil || !factor.Gt(decimal.Zero) {
			return UpgradeDefinition{}, fmt.Errorf("effects[%d] factor must be a positive canonical Decimal", index)
		}
		seenEffects[sourceEffect.SourceID] = true
		definition.Effects = append(definition.Effects, UpgradeEffect{SourceID: sourceEffect.SourceID, Slot: SlotUpgrades, Target: sourceEffect.Target, Factor: factor})
	}
	seenRoles := map[string]bool{}
	for index, role := range source.Roles {
		kind := GeneratorRoleKind(role)
		if !validRoleKind(kind) || seenRoles[role] {
			return UpgradeDefinition{}, fmt.Errorf("roles[%d] is invalid or duplicate", index)
		}
		seenRoles[role] = true
		definition.Roles = append(definition.Roles, role)
	}
	return definition, nil
}

func parseWindow(source rawWindow) (AvailabilityWindow, error) {
	from, err := nullableMechanicalID(source.FromGate)
	if err != nil {
		return AvailabilityWindow{}, fmt.Errorf("window from_gate: %v", err)
	}
	to, err := nullableMechanicalID(source.ToGate)
	if err != nil {
		return AvailabilityWindow{}, fmt.Errorf("window to_gate: %v", err)
	}
	return AvailabilityWindow{FromGate: from, ToGate: to}, nil
}

func nullableMechanicalID(data json.RawMessage) (string, error) {
	if data == nil {
		return "", errors.New("field is required")
	}
	if strings.TrimSpace(string(data)) == "null" {
		return "", nil
	}
	var value string
	if err := json.Unmarshal(data, &value); err != nil || !validID(value) {
		return "", errors.New("must be null or a mechanical id")
	}
	return value, nil
}

func validRoleKind(kind GeneratorRoleKind) bool {
	return kind == RoleProvision || kind == RoleSynergyFeed || kind == RoleManualOutput || kind == RoleStockRate
}

func parseSynergyPool(source rawSynergyPool, generators map[string]GeneratorClassDefinition, upgrades map[string]UpgradeDefinition) (SynergyPoolDefinition, error) {
	definition := SynergyPoolDefinition{ID: source.ID, Slot: MultiplierSlot(source.Slot), Curve: SynergyCurve(source.Curve)}
	if !validID(definition.ID) || !validMultiplierSlot(definition.Slot) || (definition.Curve != SynergyLinear && definition.Curve != SynergyLog) || len(source.Sources) == 0 {
		return SynergyPoolDefinition{}, errors.New("id, non-empty sources, valid slot, and linear|log curve are required")
	}
	seen := map[string]bool{}
	for index, item := range source.Sources {
		entry := SynergySource{Kind: SynergySourceKind(item.Kind), ID: item.ID, PerCountPPM: item.PerCountPPM}
		key := string(entry.Kind) + ":" + entry.ID
		if !validID(entry.ID) || entry.PerCountPPM <= 0 || entry.PerCountPPM > decimal.MaxExactInteger || seen[key] {
			return SynergyPoolDefinition{}, fmt.Errorf("sources[%d] is invalid or duplicate", index)
		}
		switch entry.Kind {
		case SynergyGenerator:
			if _, exists := generators[entry.ID]; !exists {
				return SynergyPoolDefinition{}, fmt.Errorf("sources[%d] references unknown generator", index)
			}
		case SynergyUpgrade:
			if _, exists := upgrades[entry.ID]; !exists {
				return SynergyPoolDefinition{}, fmt.Errorf("sources[%d] references unknown upgrade", index)
			}
		default:
			return SynergyPoolDefinition{}, fmt.Errorf("sources[%d] has unsupported kind", index)
		}
		seen[key] = true
		definition.Sources = append(definition.Sources, entry)
	}
	return definition, nil
}

func (c *Catalog) validateGeneratorContent() error {
	providerByTarget := map[string]string{}
	for _, generator := range c.generators {
		if generator.Provision == nil {
			continue
		}
		target, exists := c.generatorByID[generator.Provision.GeneratorID]
		if !exists || generator.Tier != target.Tier+1 || target.ProvisionedHardcap == nil {
			return fmt.Errorf("%q provision edge must target a capped generator exactly one tier down", generator.ID)
		}
		if prior := providerByTarget[target.ID]; prior != "" {
			return fmt.Errorf("provision target %q has multiple providers %q and %q", target.ID, prior, generator.ID)
		}
		providerByTarget[target.ID] = generator.ID
	}
	for index := range c.generators {
		generator := c.generators[index]
		for _, role := range generator.Roles {
			switch role.Kind {
			case RoleProvision:
				if generator.Provision == nil || generator.Provision.GeneratorID != role.GeneratorID {
					return fmt.Errorf("%q provision role does not match its edge", generator.ID)
				}
			case RoleSynergyFeed:
				pool, exists := c.synergyByID[role.PoolID]
				if !exists || !poolHasGenerator(pool, generator.ID) {
					return fmt.Errorf("%q synergy_feed role does not match pool %q", generator.ID, role.PoolID)
				}
			case RoleManualOutput:
				if _, exists := c.manualByID[role.ActionID]; !exists {
					return fmt.Errorf("%q manual_output role references unknown action %q", generator.ID, role.ActionID)
				}
			case RoleStockRate:
			default:
				return fmt.Errorf("%q has unsupported role %q", generator.ID, role.Kind)
			}
		}
	}
	return nil
}

func poolHasGenerator(pool SynergyPoolDefinition, id string) bool {
	for _, source := range pool.Sources {
		if source.Kind == SynergyGenerator && source.ID == id {
			return true
		}
	}
	return false
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
	if definition.Provision != nil {
		provision := *definition.Provision
		definition.Provision = &provision
	}
	if definition.ProvisionedHardcap != nil {
		hardcap := *definition.ProvisionedHardcap
		definition.ProvisionedHardcap = &hardcap
	}
	definition.Ladder = append([]LadderRung(nil), definition.Ladder...)
	definition.Roles = append([]GeneratorRole(nil), definition.Roles...)
	return definition
}

func cloneUpgrade(definition UpgradeDefinition) UpgradeDefinition {
	definition.Requires = append([]routes.Condition(nil), definition.Requires...)
	definition.Effects = append([]UpgradeEffect(nil), definition.Effects...)
	definition.Roles = append([]string(nil), definition.Roles...)
	return definition
}

func cloneSynergyPool(definition SynergyPoolDefinition) SynergyPoolDefinition {
	definition.Sources = append([]SynergySource(nil), definition.Sources...)
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

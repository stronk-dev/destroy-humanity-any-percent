// Package garden is Server Garden (rfc/minigame-server-garden.md): the
// `server_garden` artifact grammar (SG1), the Founder garden state (SG2), the
// lazy wall-clock advance (SG3), the deterministic tick (SG4), the pure
// Founder commands (SG5), and the Founder-side harvest math (SG6). It is a
// persistent Founder mechanic, never a platform session (SG0).
package garden

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
)

const (
	SchemaVersion = 1
	// ArtifactName is the pinned artifact name inside a catalog bundle.
	ArtifactName = "server_garden"
	// FaucetOwnerID is the SG-P1 persistent-tenant faucet identity.
	FaucetOwnerID = "server_garden"
	// PayoutScoreFactID is the garden's only score fact (SG1 rule 9).
	PayoutScoreFactID = "garden.harvest_units"

	SoulGateUnrelated  = "unrelated"
	SoulGateHumanHobby = "human_hobby"

	PPM                   = int64(1_000_000)
	MaxGridSide           = int64(6)
	MaxAgeTicks           = int64(1_000_000)
	MaxEffectPPM          = int64(10_000_000)
	MinTickMS             = int64(60_000)
	MaxTickMS             = int64(86_400_000)
	MaxHarvestUnits       = int64(1_000_000_000)
	MaxCatchupCapMS       = int64(604_800_000)
	MaxSubstrateLockoutMS = int64(86_400_000)
	// EvaluationBudget bounds ceil(cap/min tick) × max area (SG1 rule 8).
	EvaluationBudget = int64(2_000_000)
	maxRows          = 256
)

var (
	ErrInvalidCatalog    = errors.New("invalid server_garden artifact")
	ErrInvalidTransition = errors.New("invalid server_garden epoch transition")

	idPattern         = regexp.MustCompile(`^[a-z][a-z0-9_]{0,47}$`)
	mechanicalPattern = regexp.MustCompile(`^[a-z][a-z0-9_]*(?:\.[a-z][a-z0-9_]*)*$`)
)

type DimensionRow struct {
	MinLevel int64 `json:"min_level"`
	Width    int64 `json:"width"`
	Height   int64 `json:"height"`
}

type Substrate struct {
	SubstrateID     string `json:"substrate_id"`
	TickMS          int64  `json:"tick_ms"`
	EffectPPM       int64  `json:"effect_ppm"`
	ChanceFactorPPM int64  `json:"chance_factor_ppm"`
	NameCopyKey     string `json:"name_copy_key"`
	TooltipCopyKey  string `json:"tooltip_copy_key"`
}

type Species struct {
	SpeciesID          string `json:"species_id"`
	MaturationTicks    int64  `json:"maturation_ticks"`
	HarvestUnits       int64  `json:"harvest_units"`
	Starter            bool   `json:"starter"`
	NameCopyKey        string `json:"name_copy_key"`
	DescriptionCopyKey string `json:"description_copy_key"`
}

type Parent struct {
	SpeciesID           string `json:"species_id"`
	MinMatureNeighbours int64  `json:"min_mature_neighbours"`
}

type Recipe struct {
	RecipeID       string   `json:"recipe_id"`
	ChildSpeciesID string   `json:"child_species_id"`
	Parents        []Parent `json:"parents"`
	ChancePPM      int64    `json:"chance_ppm"`
}

// Catalog is a loaded `server_garden` artifact. Arrays keep raw-byte ID order.
type Catalog struct {
	UnlockID             string
	HostGeneratorID      string
	SoulGate             string
	MaxWidth             int64
	MaxHeight            int64
	DimensionRows        []DimensionRow
	GridHardcapReasonKey string
	CatchupCapMS         int64
	CatchupReasonKey     string
	SubstrateLockoutMS   int64
	LockoutReasonKey     string
	DefaultSubstrateID   string
	Substrates           []Substrate
	Species              []Species
	Recipes              []Recipe
	Payout               PayoutPolicy

	substrateIndex map[string]int
	speciesIndex   map[string]int
}

// PayoutPolicy mirrors the platform's six-key payout row (SG1 rule 9).
type PayoutPolicy struct {
	CreditedResourceID string `json:"credited_resource_id"`
	SendsPerDay        int64  `json:"sends_per_day"`
	PerSendCap         int64  `json:"per_send_cap"`
	ConversionPPM      int64  `json:"conversion_ppm"`
	PayoutScoreFactID  string `json:"payout_score_fact_id"`
	CapReasonKey       string `json:"cap_reason_key"`
}

// PayoutValidator is the platform's payout loader, injected so the garden
// never imports the session platform (and the save layer can import the
// garden). It validates the row against the declared score set; the garden
// then decodes the same six keys exactly.
type PayoutValidator func(data []byte, scoreFactIDs map[string]struct{}) error

// Declarations are the cross-artifact authorities SG1 rules 9 and 10 bind to.
type Declarations struct {
	CopyKeys           map[string]struct{}
	ResourceIDs        map[string]struct{}
	FiscalUnlockIDs    map[string]struct{}
	FiscalGeneratorIDs map[string]struct{}
	ValidatePayout     PayoutValidator
}

// Substrate returns the substrate row for id.
func (catalog *Catalog) Substrate(id string) (Substrate, bool) {
	if catalog == nil {
		return Substrate{}, false
	}
	index, ok := catalog.substrateIndex[id]
	if !ok {
		return Substrate{}, false
	}
	return catalog.Substrates[index], true
}

// SpeciesRow returns the species row for id.
func (catalog *Catalog) SpeciesRow(id string) (Species, bool) {
	if catalog == nil {
		return Species{}, false
	}
	index, ok := catalog.speciesIndex[id]
	if !ok {
		return Species{}, false
	}
	return catalog.Species[index], true
}

// Starters returns the starter species IDs in catalog (raw-byte) order.
func (catalog *Catalog) Starters() []string {
	starters := []string{}
	for _, species := range catalog.Species {
		if species.Starter {
			starters = append(starters, species.SpeciesID)
		}
	}
	return starters
}

// Dimension resolves the active grid for a host Fiscal level: the last row
// whose min_level is at most level (SG3).
func (catalog *Catalog) Dimension(level int64) DimensionRow {
	active := catalog.DimensionRows[0]
	for _, row := range catalog.DimensionRows {
		if row.MinLevel <= level {
			active = row
		}
	}
	return active
}

// ContentHash is the pinned identity of the artifact bytes.
func ContentHash(data []byte) string {
	digest := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(digest[:])
}

type wireCatalog struct {
	SchemaVersion      int             `json:"schema_version"`
	UnlockID           string          `json:"unlock_id"`
	HostGeneratorID    string          `json:"host_generator_id"`
	SoulGate           string          `json:"soul_gate"`
	Grid               json.RawMessage `json:"grid"`
	Clock              json.RawMessage `json:"clock"`
	DefaultSubstrateID string          `json:"default_substrate_id"`
	Substrates         json.RawMessage `json:"substrates"`
	Species            json.RawMessage `json:"species"`
	Recipes            json.RawMessage `json:"recipes"`
	Payout             json.RawMessage `json:"payout"`
}

type wireGrid struct {
	MaxWidth         int64           `json:"max_width"`
	MaxHeight        int64           `json:"max_height"`
	DimensionRows    json.RawMessage `json:"dimension_rows"`
	HardcapReasonKey string          `json:"hardcap_reason_key"`
}

type wireClock struct {
	CatchupCapMS       int64  `json:"catchup_cap_ms"`
	CatchupReasonKey   string `json:"catchup_reason_key"`
	SubstrateLockoutMS int64  `json:"substrate_lockout_ms"`
	LockoutReasonKey   string `json:"lockout_reason_key"`
}

// LoadCatalog decodes and validates a `server_garden` artifact against every
// SG1 rule. Any violation fails the whole load.
func LoadCatalog(data []byte, declarations Declarations) (*Catalog, error) {
	if declarations.CopyKeys == nil || declarations.ResourceIDs == nil || declarations.FiscalUnlockIDs == nil || declarations.FiscalGeneratorIDs == nil || declarations.ValidatePayout == nil {
		return nil, fmt.Errorf("%w: declarations are required", ErrInvalidCatalog)
	}
	if !uniqueKeys(data) {
		return nil, fmt.Errorf("%w: duplicate or malformed keys", ErrInvalidCatalog)
	}
	var wire wireCatalog
	if err := exactDecode(data, &wire, "schema_version", "unlock_id", "host_generator_id", "soul_gate", "grid", "clock", "default_substrate_id", "substrates", "species", "recipes", "payout"); err != nil {
		return nil, err
	}
	if wire.SchemaVersion != SchemaVersion {
		return nil, fmt.Errorf("%w: schema_version must be %d", ErrInvalidCatalog, SchemaVersion)
	}
	catalog := &Catalog{UnlockID: wire.UnlockID, HostGeneratorID: wire.HostGeneratorID, SoulGate: wire.SoulGate,
		DefaultSubstrateID: wire.DefaultSubstrateID, substrateIndex: map[string]int{}, speciesIndex: map[string]int{}}
	if !mechanicalPattern.MatchString(catalog.UnlockID) || !mechanicalPattern.MatchString(catalog.HostGeneratorID) {
		return nil, fmt.Errorf("%w: unlock/host ids are not mechanical", ErrInvalidCatalog)
	}
	if _, ok := declarations.FiscalUnlockIDs[catalog.UnlockID]; !ok {
		return nil, fmt.Errorf("%w: unlock_id is not a Fiscal unlock row", ErrInvalidCatalog)
	}
	if _, ok := declarations.FiscalGeneratorIDs[catalog.HostGeneratorID]; !ok {
		return nil, fmt.Errorf("%w: host_generator_id is not a Fiscal generator level row", ErrInvalidCatalog)
	}
	if catalog.SoulGate != SoulGateUnrelated && catalog.SoulGate != SoulGateHumanHobby {
		return nil, fmt.Errorf("%w: unknown soul_gate", ErrInvalidCatalog)
	}
	if err := catalog.loadGrid(wire.Grid, declarations); err != nil {
		return nil, err
	}
	if err := catalog.loadClock(wire.Clock, declarations); err != nil {
		return nil, err
	}
	if err := catalog.loadSubstrates(wire.Substrates, declarations); err != nil {
		return nil, err
	}
	if _, ok := catalog.substrateIndex[catalog.DefaultSubstrateID]; !ok {
		return nil, fmt.Errorf("%w: default_substrate_id does not exist", ErrInvalidCatalog)
	}
	if err := catalog.loadSpecies(wire.Species, declarations); err != nil {
		return nil, err
	}
	if err := catalog.loadRecipes(wire.Recipes); err != nil {
		return nil, err
	}
	if err := catalog.validateReachability(); err != nil {
		return nil, err
	}
	if err := catalog.validateSatisfiability(); err != nil {
		return nil, err
	}
	if err := catalog.validateBudget(); err != nil {
		return nil, err
	}
	scoreFacts := map[string]struct{}{PayoutScoreFactID: {}}
	var payout PayoutPolicy
	if err := declarations.ValidatePayout(wire.Payout, scoreFacts); err != nil ||
		exactDecode(wire.Payout, &payout, "credited_resource_id", "sends_per_day", "per_send_cap", "conversion_ppm", "payout_score_fact_id", "cap_reason_key") != nil ||
		payout.PayoutScoreFactID != PayoutScoreFactID {
		return nil, fmt.Errorf("%w: payout policy", ErrInvalidCatalog)
	}
	catalog.Payout = payout
	return catalog, nil
}

func (catalog *Catalog) loadGrid(data json.RawMessage, declarations Declarations) error {
	var grid wireGrid
	if err := exactDecode(data, &grid, "max_width", "max_height", "dimension_rows", "hardcap_reason_key"); err != nil {
		return err
	}
	if grid.MaxWidth < 1 || grid.MaxWidth > MaxGridSide || grid.MaxHeight < 1 || grid.MaxHeight > MaxGridSide {
		return fmt.Errorf("%w: max grid side outside [1,%d]", ErrInvalidCatalog, MaxGridSide)
	}
	if !hasKey(declarations.CopyKeys, grid.HardcapReasonKey) {
		return fmt.Errorf("%w: grid hardcap reason key", ErrInvalidCatalog)
	}
	var raw []json.RawMessage
	if err := strictDecode(grid.DimensionRows, &raw); err != nil || len(raw) == 0 || len(raw) > maxRows {
		return fmt.Errorf("%w: dimension_rows must be a non-empty bounded array", ErrInvalidCatalog)
	}
	rows := make([]DimensionRow, 0, len(raw))
	for index, item := range raw {
		var row DimensionRow
		if err := exactDecode(item, &row, "min_level", "width", "height"); err != nil {
			return err
		}
		if index == 0 && row.MinLevel != 0 || index > 0 && row.MinLevel <= rows[index-1].MinLevel {
			return fmt.Errorf("%w: dimension min_level must start at 0 and strictly ascend", ErrInvalidCatalog)
		}
		if row.Width < 1 || row.Width > grid.MaxWidth || row.Height < 1 || row.Height > grid.MaxHeight ||
			index > 0 && (row.Width < rows[index-1].Width || row.Height < rows[index-1].Height) {
			return fmt.Errorf("%w: dimension rows must be non-decreasing within the max grid", ErrInvalidCatalog)
		}
		rows = append(rows, row)
	}
	if last := rows[len(rows)-1]; last.Width != grid.MaxWidth || last.Height != grid.MaxHeight {
		return fmt.Errorf("%w: the last dimension row must equal the max grid (visible hardcap)", ErrInvalidCatalog)
	}
	catalog.MaxWidth, catalog.MaxHeight, catalog.DimensionRows, catalog.GridHardcapReasonKey = grid.MaxWidth, grid.MaxHeight, rows, grid.HardcapReasonKey
	return nil
}

func (catalog *Catalog) loadClock(data json.RawMessage, declarations Declarations) error {
	var clock wireClock
	if err := exactDecode(data, &clock, "catchup_cap_ms", "catchup_reason_key", "substrate_lockout_ms", "lockout_reason_key"); err != nil {
		return err
	}
	if clock.CatchupCapMS < 1 || clock.CatchupCapMS > MaxCatchupCapMS || clock.SubstrateLockoutMS < 0 || clock.SubstrateLockoutMS > MaxSubstrateLockoutMS {
		return fmt.Errorf("%w: clock domain", ErrInvalidCatalog)
	}
	if !hasKey(declarations.CopyKeys, clock.CatchupReasonKey) || !hasKey(declarations.CopyKeys, clock.LockoutReasonKey) {
		return fmt.Errorf("%w: clock reason keys", ErrInvalidCatalog)
	}
	catalog.CatchupCapMS, catalog.CatchupReasonKey, catalog.SubstrateLockoutMS, catalog.LockoutReasonKey = clock.CatchupCapMS, clock.CatchupReasonKey, clock.SubstrateLockoutMS, clock.LockoutReasonKey
	return nil
}

func (catalog *Catalog) loadSubstrates(data json.RawMessage, declarations Declarations) error {
	var raw []json.RawMessage
	if err := strictDecode(data, &raw); err != nil || len(raw) == 0 || len(raw) > maxRows {
		return fmt.Errorf("%w: substrates must be a non-empty bounded array", ErrInvalidCatalog)
	}
	for index, item := range raw {
		var row Substrate
		if err := exactDecode(item, &row, "substrate_id", "tick_ms", "effect_ppm", "chance_factor_ppm", "name_copy_key", "tooltip_copy_key"); err != nil {
			return err
		}
		if !idPattern.MatchString(row.SubstrateID) || index > 0 && catalog.Substrates[index-1].SubstrateID >= row.SubstrateID {
			return fmt.Errorf("%w: substrate ids must match the id grammar and be unique and byte-sorted", ErrInvalidCatalog)
		}
		if row.TickMS < MinTickMS || row.TickMS > MaxTickMS || row.EffectPPM < 0 || row.EffectPPM > MaxEffectPPM ||
			row.ChanceFactorPPM < 0 || row.ChanceFactorPPM > MaxEffectPPM {
			return fmt.Errorf("%w: substrate %q domain", ErrInvalidCatalog, row.SubstrateID)
		}
		if !hasKey(declarations.CopyKeys, row.NameCopyKey) || !hasKey(declarations.CopyKeys, row.TooltipCopyKey) {
			return fmt.Errorf("%w: substrate %q copy keys", ErrInvalidCatalog, row.SubstrateID)
		}
		catalog.substrateIndex[row.SubstrateID] = index
		catalog.Substrates = append(catalog.Substrates, row)
	}
	return nil
}

func (catalog *Catalog) loadSpecies(data json.RawMessage, declarations Declarations) error {
	var raw []json.RawMessage
	if err := strictDecode(data, &raw); err != nil || len(raw) == 0 || len(raw) > maxRows {
		return fmt.Errorf("%w: species must be a non-empty bounded array", ErrInvalidCatalog)
	}
	starters := 0
	for index, item := range raw {
		var row Species
		if err := exactDecode(item, &row, "species_id", "maturation_ticks", "harvest_units", "starter", "name_copy_key", "description_copy_key"); err != nil {
			return err
		}
		if !idPattern.MatchString(row.SpeciesID) || index > 0 && catalog.Species[index-1].SpeciesID >= row.SpeciesID {
			return fmt.Errorf("%w: species ids must match the id grammar and be unique and byte-sorted", ErrInvalidCatalog)
		}
		if row.MaturationTicks < 1 || row.MaturationTicks > MaxAgeTicks || row.HarvestUnits < 0 || row.HarvestUnits > MaxHarvestUnits {
			return fmt.Errorf("%w: species %q domain", ErrInvalidCatalog, row.SpeciesID)
		}
		if !hasKey(declarations.CopyKeys, row.NameCopyKey) || !hasKey(declarations.CopyKeys, row.DescriptionCopyKey) {
			return fmt.Errorf("%w: species %q copy keys", ErrInvalidCatalog, row.SpeciesID)
		}
		if row.Starter {
			starters++
		}
		catalog.speciesIndex[row.SpeciesID] = index
		catalog.Species = append(catalog.Species, row)
	}
	if starters == 0 {
		return fmt.Errorf("%w: at least one species must be a starter", ErrInvalidCatalog)
	}
	return nil
}

func (catalog *Catalog) loadRecipes(data json.RawMessage) error {
	var raw []json.RawMessage
	if err := strictDecode(data, &raw); err != nil || len(raw) > maxRows {
		return fmt.Errorf("%w: recipes must be a bounded array", ErrInvalidCatalog)
	}
	for index, item := range raw {
		var fields struct {
			RecipeID       string          `json:"recipe_id"`
			ChildSpeciesID string          `json:"child_species_id"`
			Parents        json.RawMessage `json:"parents"`
			ChancePPM      int64           `json:"chance_ppm"`
		}
		if err := exactDecode(item, &fields, "recipe_id", "child_species_id", "parents", "chance_ppm"); err != nil {
			return err
		}
		if !idPattern.MatchString(fields.RecipeID) || index > 0 && catalog.Recipes[index-1].RecipeID >= fields.RecipeID {
			return fmt.Errorf("%w: recipe ids must match the id grammar and be unique and byte-sorted", ErrInvalidCatalog)
		}
		if _, ok := catalog.speciesIndex[fields.ChildSpeciesID]; !ok {
			return fmt.Errorf("%w: recipe %q child does not exist", ErrInvalidCatalog, fields.RecipeID)
		}
		if fields.ChancePPM < 1 || fields.ChancePPM > PPM {
			return fmt.Errorf("%w: recipe %q chance_ppm domain", ErrInvalidCatalog, fields.RecipeID)
		}
		var rawParents []json.RawMessage
		if err := strictDecode(fields.Parents, &rawParents); err != nil || len(rawParents) < 1 || len(rawParents) > 2 {
			return fmt.Errorf("%w: recipe %q needs 1-2 parents", ErrInvalidCatalog, fields.RecipeID)
		}
		recipe := Recipe{RecipeID: fields.RecipeID, ChildSpeciesID: fields.ChildSpeciesID, ChancePPM: fields.ChancePPM}
		sum := int64(0)
		for parentIndex, rawParent := range rawParents {
			var parent Parent
			if err := exactDecode(rawParent, &parent, "species_id", "min_mature_neighbours"); err != nil {
				return err
			}
			if _, ok := catalog.speciesIndex[parent.SpeciesID]; !ok {
				return fmt.Errorf("%w: recipe %q parent does not exist", ErrInvalidCatalog, fields.RecipeID)
			}
			if parentIndex > 0 && recipe.Parents[parentIndex-1].SpeciesID >= parent.SpeciesID {
				return fmt.Errorf("%w: recipe %q parents must be distinct and byte-sorted", ErrInvalidCatalog, fields.RecipeID)
			}
			if parent.MinMatureNeighbours < 1 || parent.MinMatureNeighbours > 8 {
				return fmt.Errorf("%w: recipe %q min_mature_neighbours domain", ErrInvalidCatalog, fields.RecipeID)
			}
			sum += parent.MinMatureNeighbours
			recipe.Parents = append(recipe.Parents, parent)
		}
		if sum > 8 {
			return fmt.Errorf("%w: recipe %q parent counts exceed 8 neighbours", ErrInvalidCatalog, fields.RecipeID)
		}
		catalog.Recipes = append(catalog.Recipes, recipe)
	}
	return nil
}

// validateReachability is SG1 rule 5: the fixpoint from the starters must
// discover every species (the completable Pokédex).
func (catalog *Catalog) validateReachability() error {
	reachable := map[string]bool{}
	for _, species := range catalog.Species {
		if species.Starter {
			reachable[species.SpeciesID] = true
		}
	}
	for changed := true; changed; {
		changed = false
		for _, recipe := range catalog.Recipes {
			if reachable[recipe.ChildSpeciesID] {
				continue
			}
			all := true
			for _, parent := range recipe.Parents {
				all = all && reachable[parent.SpeciesID]
			}
			if all {
				reachable[recipe.ChildSpeciesID], changed = true, true
			}
		}
	}
	for _, species := range catalog.Species {
		if !reachable[species.SpeciesID] {
			return fmt.Errorf("%w: species %q is unreachable from the starters", ErrInvalidCatalog, species.SpeciesID)
		}
	}
	return nil
}

// validateSatisfiability is SG1 rule 6: every recipe's parent-count sum fits
// the largest Moore neighbourhood inside the max grid.
func (catalog *Catalog) validateSatisfiability() error {
	largest := MooreCapacity(catalog.MaxWidth, catalog.MaxHeight)
	for _, recipe := range catalog.Recipes {
		sum := int64(0)
		for _, parent := range recipe.Parents {
			sum += parent.MinMatureNeighbours
		}
		if sum > largest {
			return fmt.Errorf("%w: recipe %q cannot be satisfied inside the max grid", ErrInvalidCatalog, recipe.RecipeID)
		}
	}
	return nil
}

// MooreCapacity is the largest number of in-grid Moore neighbours any cell of
// a width×height grid has.
func MooreCapacity(width, height int64) int64 {
	across := min(width, 3) - 1
	down := min(height, 3) - 1
	return (across+1)*(down+1) - 1
}

// validateBudget is SG1 rule 8: ceil(catchup_cap_ms / min tick) × max area
// bounds per-command work.
func (catalog *Catalog) validateBudget() error {
	minimum := catalog.Substrates[0].TickMS
	for _, substrate := range catalog.Substrates {
		minimum = min(minimum, substrate.TickMS)
	}
	ticks := (catalog.CatchupCapMS + minimum - 1) / minimum
	if ticks*catalog.MaxWidth*catalog.MaxHeight > EvaluationBudget {
		return fmt.Errorf("%w: clock evaluation budget exceeds %d", ErrInvalidCatalog, EvaluationBudget)
	}
	return nil
}

// ValidateTransition is SG1 rule 11: species, substrate, and recipe IDs are
// append-only across epochs.
func ValidateTransition(current, next *Catalog) error {
	if current == nil {
		return nil
	}
	if next == nil {
		return fmt.Errorf("%w: server_garden cannot disappear between epochs", ErrInvalidTransition)
	}
	for _, species := range current.Species {
		if _, ok := next.speciesIndex[species.SpeciesID]; !ok {
			return fmt.Errorf("%w: species %q was removed", ErrInvalidTransition, species.SpeciesID)
		}
	}
	for _, substrate := range current.Substrates {
		if _, ok := next.substrateIndex[substrate.SubstrateID]; !ok {
			return fmt.Errorf("%w: substrate %q was removed", ErrInvalidTransition, substrate.SubstrateID)
		}
	}
	nextRecipes := map[string]bool{}
	for _, recipe := range next.Recipes {
		nextRecipes[recipe.RecipeID] = true
	}
	for _, recipe := range current.Recipes {
		if !nextRecipes[recipe.RecipeID] {
			return fmt.Errorf("%w: recipe %q was removed", ErrInvalidTransition, recipe.RecipeID)
		}
	}
	return nil
}

func hasKey(keys map[string]struct{}, key string) bool {
	if !mechanicalPattern.MatchString(key) {
		return false
	}
	_, ok := keys[key]
	return ok
}

// exactDecode decodes one JSON object requiring exactly keys (no missing,
// unknown, or duplicate keys) with safe-integer numbers.
func exactDecode(data []byte, target any, keys ...string) error {
	var fields map[string]json.RawMessage
	if err := strictDecode(data, &fields); err != nil || fields == nil || len(fields) != len(keys) || !uniqueKeys(data) {
		return fmt.Errorf("%w: object keys are not exact", ErrInvalidCatalog)
	}
	for _, key := range keys {
		value, ok := fields[key]
		if !ok {
			return fmt.Errorf("%w: missing key %q", ErrInvalidCatalog, key)
		}
		if !safeNumbers(value) {
			return fmt.Errorf("%w: key %q is not an exact safe integer", ErrInvalidCatalog, key)
		}
	}
	if err := strictDecode(data, target); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidCatalog, err)
	}
	return nil
}

// safeNumbers rejects any JSON number token that is not an exact integer in
// [-(2^53-1), 2^53-1] anywhere inside value.
func safeNumbers(value json.RawMessage) bool {
	decoder := json.NewDecoder(bytes.NewReader(value))
	decoder.UseNumber()
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			return true
		}
		if err != nil {
			return false
		}
		number, ok := token.(json.Number)
		if !ok {
			continue
		}
		parsed, err := number.Int64()
		if err != nil || parsed > maxExactInteger || parsed < -maxExactInteger {
			return false
		}
	}
}

const maxExactInteger = int64(9_007_199_254_740_991)

func strictDecode(data []byte, target any) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	if _, err := decoder.Token(); err != io.EOF {
		return fmt.Errorf("%w: trailing data", ErrInvalidCatalog)
	}
	return nil
}

// uniqueKeys rejects duplicate object keys at any depth.
func uniqueKeys(data []byte) bool {
	decoder := json.NewDecoder(bytes.NewReader(data))
	var walk func() bool
	walk = func() bool {
		token, err := decoder.Token()
		if err != nil {
			return false
		}
		delim, ok := token.(json.Delim)
		if !ok {
			return true
		}
		switch delim {
		case '{':
			seen := map[string]bool{}
			for decoder.More() {
				keyToken, err := decoder.Token()
				key, isString := keyToken.(string)
				if err != nil || !isString || seen[key] {
					return false
				}
				seen[key] = true
				if !walk() {
					return false
				}
			}
			_, err = decoder.Token()
			return err == nil
		case '[':
			for decoder.More() {
				if !walk() {
					return false
				}
			}
			_, err = decoder.Token()
			return err == nil
		}
		return false
	}
	if !walk() {
		return false
	}
	_, err := decoder.Token()
	return err == io.EOF
}

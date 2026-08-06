// Package fiscal owns the Founder-scoped fiscal-period catalog and its pure,
// replayable arithmetic. It deliberately has no store or production imports.
package fiscal

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"regexp"
	"sort"

	"cloud-clicker/server/decimal"
	"cloud-clicker/server/determinism"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/multiplier"
	"cloud-clicker/server/runidentity"
)

const (
	CatalogSchemaVersion  = 1
	EarlyHarvestSubstream = "fiscal.early_harvest.v1"
	PPM                   = int64(1_000_000)
)

var (
	ErrInvalidCatalog  = errors.New("invalid fiscal catalog")
	ErrInvalidState    = errors.New("invalid fiscal state")
	ErrClockRegression = errors.New("fiscal wall clock regressed")
	ErrNotRipe         = errors.New("fiscal period not ripe")
	ErrUnknownTarget   = errors.New("unknown fiscal target")
	ErrAlreadyUnlocked = errors.New("fiscal target already unlocked")
	ErrUnaffordable    = errors.New("insufficient fiscal credit")
	ErrCapExceeded     = errors.New("fiscal target cap exceeded")
	idPattern          = regexp.MustCompile(`^[a-z][a-z0-9_]*(?:\.[a-z][a-z0-9_]*)*$`)
)

type ClockPolicy struct {
	EarlyMS         int64
	GuaranteedMS    int64
	AutoMS          int64
	EarlySuccessPPM int64
}

type CreditPolicy struct {
	CreditPerPeriod  int64
	Hardcap          int64
	HardcapReasonKey string
}

type HoardPolicy struct {
	PPMPerCredit int64
	CapCredits   int64
	Slot         multiplier.Slot
	SourceID     string
	Target       string
}

type GeneratorLevelRow struct {
	GeneratorID      string
	PPMPerLevel      int64
	LevelHardcap     int64
	HardcapReasonKey string
	Slot             multiplier.Slot
	SourceID         string
}

type UnlockRow struct {
	UnlockID string
	Cost     int64
}

type Catalog struct {
	Clock         ClockPolicy
	Credit        CreditPolicy
	Hoard         HoardPolicy
	generatorRows []GeneratorLevelRow
	unlockRows    []UnlockRow
	generatorByID map[string]GeneratorLevelRow
	unlockByID    map[string]UnlockRow
}

type rawCatalog struct {
	SchemaVersion      *int                   `json:"schema_version"`
	ClockPolicy        rawClockPolicy         `json:"clock_policy"`
	CreditPolicy       rawCreditPolicy        `json:"credit_policy"`
	HoardPolicy        rawHoardPolicy         `json:"hoard_policy"`
	GeneratorLevelRows []rawGeneratorLevelRow `json:"generator_level_rows"`
	UnlockRows         []rawUnlockRow         `json:"unlock_rows"`
}

type rawClockPolicy struct {
	EarlyMS         *int64 `json:"early_ms"`
	GuaranteedMS    *int64 `json:"guaranteed_ms"`
	AutoMS          *int64 `json:"auto_ms"`
	EarlySuccessPPM *int64 `json:"early_success_ppm"`
}

type rawCreditPolicy struct {
	CreditPerPeriod  *int64 `json:"credit_per_period"`
	Hardcap          *int64 `json:"hardcap"`
	HardcapReasonKey string `json:"hardcap_reason_key"`
}

type rawHoardPolicy struct {
	PPMPerCredit *int64 `json:"ppm_per_credit"`
	CapCredits   *int64 `json:"cap_credits"`
	Slot         string `json:"slot"`
	SourceID     string `json:"source_id"`
	Target       string `json:"target"`
}

type rawGeneratorLevelRow struct {
	GeneratorID      string `json:"generator_id"`
	PPMPerLevel      *int64 `json:"ppm_per_level"`
	LevelHardcap     *int64 `json:"level_hardcap"`
	HardcapReasonKey string `json:"hardcap_reason_key"`
	Slot             string `json:"slot"`
	SourceID         string `json:"source_id"`
}

type rawUnlockRow struct {
	UnlockID string `json:"unlock_id"`
	Cost     *int64 `json:"cost"`
}

func LoadCatalog(data []byte, economyCatalog *economy.Catalog) (*Catalog, error) {
	if economyCatalog == nil {
		return nil, ErrInvalidCatalog
	}
	var raw rawCatalog
	if err := decodeStrict(data, &raw); err != nil {
		return nil, fmt.Errorf("%w: decode: %v", ErrInvalidCatalog, err)
	}
	if raw.SchemaVersion == nil || *raw.SchemaVersion != CatalogSchemaVersion ||
		raw.ClockPolicy.EarlyMS == nil || raw.ClockPolicy.GuaranteedMS == nil || raw.ClockPolicy.AutoMS == nil || raw.ClockPolicy.EarlySuccessPPM == nil ||
		raw.CreditPolicy.CreditPerPeriod == nil || raw.CreditPolicy.Hardcap == nil ||
		raw.HoardPolicy.PPMPerCredit == nil || raw.HoardPolicy.CapCredits == nil ||
		len(raw.GeneratorLevelRows) == 0 || len(raw.UnlockRows) == 0 {
		return nil, ErrInvalidCatalog
	}
	clock := ClockPolicy{*raw.ClockPolicy.EarlyMS, *raw.ClockPolicy.GuaranteedMS, *raw.ClockPolicy.AutoMS, *raw.ClockPolicy.EarlySuccessPPM}
	credit := CreditPolicy{*raw.CreditPolicy.CreditPerPeriod, *raw.CreditPolicy.Hardcap, raw.CreditPolicy.HardcapReasonKey}
	hoard := HoardPolicy{*raw.HoardPolicy.PPMPerCredit, *raw.HoardPolicy.CapCredits, multiplier.Slot(raw.HoardPolicy.Slot), raw.HoardPolicy.SourceID, raw.HoardPolicy.Target}
	if !validClock(clock) || credit.CreditPerPeriod <= 0 || credit.CreditPerPeriod > credit.Hardcap || credit.Hardcap > decimal.MaxExactInteger ||
		!idPattern.MatchString(credit.HardcapReasonKey) || hoard.PPMPerCredit <= 0 || hoard.PPMPerCredit > PPM ||
		hoard.CapCredits <= 0 || hoard.CapCredits > credit.Hardcap || hoard.Target != "all" || !validDeclaration(economyCatalog, hoard.SourceID, hoard.Slot, hoard.Target) {
		return nil, ErrInvalidCatalog
	}
	catalog := &Catalog{Clock: clock, Credit: credit, Hoard: hoard, generatorByID: map[string]GeneratorLevelRow{}, unlockByID: map[string]UnlockRow{}}
	previous := ""
	for index, source := range raw.GeneratorLevelRows {
		row := GeneratorLevelRow{source.GeneratorID, valueOrZero(source.PPMPerLevel), valueOrZero(source.LevelHardcap), source.HardcapReasonKey, multiplier.Slot(source.Slot), source.SourceID}
		_, generatorExists := economyCatalog.GeneratorClass(row.GeneratorID)
		if source.PPMPerLevel == nil || source.LevelHardcap == nil || !idPattern.MatchString(row.GeneratorID) || !idPattern.MatchString(row.HardcapReasonKey) ||
			row.PPMPerLevel <= 0 || row.PPMPerLevel > PPM || row.LevelHardcap <= 0 || row.LevelHardcap > decimal.MaxExactInteger ||
			previous >= row.GeneratorID && previous != "" || !generatorExists || !validDeclaration(economyCatalog, row.SourceID, row.Slot, row.GeneratorID) {
			return nil, fmt.Errorf("%w: generator_level_rows[%d]", ErrInvalidCatalog, index)
		}
		previous = row.GeneratorID
		catalog.generatorRows = append(catalog.generatorRows, row)
		catalog.generatorByID[row.GeneratorID] = row
	}
	previous = ""
	for index, source := range raw.UnlockRows {
		row := UnlockRow{source.UnlockID, valueOrZero(source.Cost)}
		if source.Cost == nil || !idPattern.MatchString(row.UnlockID) || row.Cost <= 0 || row.Cost > credit.Hardcap || previous >= row.UnlockID && previous != "" {
			return nil, fmt.Errorf("%w: unlock_rows[%d]", ErrInvalidCatalog, index)
		}
		previous = row.UnlockID
		catalog.unlockRows = append(catalog.unlockRows, row)
		catalog.unlockByID[row.UnlockID] = row
	}
	return catalog, nil
}

func validClock(clock ClockPolicy) bool {
	return clock.EarlyMS > 0 && clock.EarlyMS <= clock.GuaranteedMS && clock.GuaranteedMS <= clock.AutoMS &&
		clock.AutoMS <= decimal.MaxExactInteger && decimal.MaxExactInteger/clock.EarlyMS < decimal.MaxExactInteger &&
		clock.EarlySuccessPPM >= 0 && clock.EarlySuccessPPM <= PPM
}

func validDeclaration(catalog *economy.Catalog, sourceID string, slot multiplier.Slot, target string) bool {
	if !idPattern.MatchString(sourceID) || !multiplier.ValidSlot(slot) {
		return false
	}
	declaration, ok := catalog.MultiplierSource(sourceID)
	return ok && multiplier.Slot(declaration.Slot) == slot && declaration.Target == target && declaration.Provider == "fiscal"
}

func valueOrZero(value *int64) int64 {
	if value == nil {
		return 0
	}
	return *value
}

func (catalog *Catalog) GeneratorLevelRows() []GeneratorLevelRow {
	if catalog == nil {
		return nil
	}
	return append([]GeneratorLevelRow(nil), catalog.generatorRows...)
}

func (catalog *Catalog) UnlockRows() []UnlockRow {
	if catalog == nil {
		return nil
	}
	return append([]UnlockRow(nil), catalog.unlockRows...)
}

func (catalog *Catalog) GeneratorLevel(id string) (GeneratorLevelRow, bool) {
	if catalog == nil {
		return GeneratorLevelRow{}, false
	}
	value, ok := catalog.generatorByID[id]
	return value, ok
}

func (catalog *Catalog) Unlock(id string) (UnlockRow, bool) {
	if catalog == nil {
		return UnlockRow{}, false
	}
	value, ok := catalog.unlockByID[id]
	return value, ok
}

func (catalog *Catalog) HoardFactor(credit int64) (decimal.Decimal, error) {
	if catalog == nil || credit < 0 || credit > catalog.Credit.Hardcap {
		return decimal.NaN, ErrInvalidState
	}
	return ppmFactor(minInt64(credit, catalog.Hoard.CapCredits), catalog.Hoard.PPMPerCredit)
}

func (catalog *Catalog) GeneratorLevelFactor(generatorID string, level int64) (decimal.Decimal, error) {
	row, ok := catalog.GeneratorLevel(generatorID)
	if !ok {
		return decimal.NaN, ErrUnknownTarget
	}
	if level < 0 || level > row.LevelHardcap {
		return decimal.NaN, ErrInvalidState
	}
	return ppmFactor(level, row.PPMPerLevel)
}

func ppmFactor(count, perCountPPM int64) (decimal.Decimal, error) {
	product := new(big.Int).Mul(big.NewInt(count), big.NewInt(perCountPPM))
	value := decimal.FromString(product.String()).Div(decimal.FromFloat64(float64(PPM))).Add(decimal.One).Quantize(decimal.CanonicalSignificantDigits)
	if !value.IsStateValue() || !value.Gt(decimal.Zero) {
		return decimal.NaN, ErrInvalidState
	}
	return value, nil
}

func (catalog *Catalog) GeneratorLevelCost(generatorID string, current, levels int64) (int64, error) {
	row, ok := catalog.GeneratorLevel(generatorID)
	if !ok {
		return 0, ErrUnknownTarget
	}
	if current < 0 || levels <= 0 || current > row.LevelHardcap || levels > row.LevelHardcap-current {
		return 0, ErrCapExceeded
	}
	// levels*(2*current+levels+1)/2, using wide integer intermediates.
	widened := new(big.Int).Mul(big.NewInt(levels), new(big.Int).Add(new(big.Int).Mul(big.NewInt(2), big.NewInt(current)), big.NewInt(levels+1)))
	widened.Div(widened, big.NewInt(2))
	if !widened.IsInt64() || widened.Int64() > decimal.MaxExactInteger {
		return 0, ErrCapExceeded
	}
	return widened.Int64(), nil
}

func EarlyHarvestDraw(founderID string, sequence int64) (int64, error) {
	if founderID == "" || sequence < 0 || sequence > decimal.MaxExactInteger {
		return 0, ErrInvalidState
	}
	base := runidentity.Seed(founderID, sequence)
	return int64(determinism.Substream(base, EarlyHarvestSubstream).Bound(uint64(PPM))), nil
}

func minInt64(left, right int64) int64 {
	if left < right {
		return left
	}
	return right
}

func SortedUnlocks(values []string) ([]string, bool) {
	result := append([]string(nil), values...)
	sort.Strings(result)
	for index, value := range result {
		if !idPattern.MatchString(value) || index > 0 && result[index-1] == value {
			return nil, false
		}
	}
	return result, true
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

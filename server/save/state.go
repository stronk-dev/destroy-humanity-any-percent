package save

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"cloud-clicker/server/decimal"
	"cloud-clicker/server/economy"
)

const (
	CurrentVersion           = 4
	millisecondCursorVersion = 4
)

var ErrInvalidState = errors.New("invalid saved state")

type State struct {
	Ledger                *economy.Ledger
	GeneratorCounts       map[string]int64
	EvaluatedThrough      time.Time
	ComputeCreditMS       int64
	ManualTokenMilli      int64
	ManualTokenRefilledAt time.Time
}

type stateV1 struct {
	Balances map[string]string `json:"balances"`
}

type stateV2 struct {
	Balances         map[string]string `json:"balances"`
	Generators       map[string]int64  `json:"generators"`
	EvaluatedThrough string            `json:"evaluated_through"`
}

type stateV4 struct {
	Balances              map[string]string `json:"balances"`
	Generators            map[string]int64  `json:"generators"`
	EvaluatedThrough      string            `json:"evaluated_through"`
	ComputeCreditMS       int64             `json:"compute_credit_ms"`
	ManualTokenMilli      int64             `json:"manual_token_milli"`
	ManualTokenRefilledAt string            `json:"manual_token_refilled_at"`
}

func ConstantsHash(catalogBytes []byte) string {
	digest := sha256.Sum256(catalogBytes)
	return "sha256:" + hex.EncodeToString(digest[:])
}

func EncodeState(state *State) ([]byte, error) {
	if state == nil || state.Ledger == nil {
		return nil, fmt.Errorf("%w: nil state or ledger", ErrInvalidState)
	}
	if state.GeneratorCounts == nil {
		return nil, fmt.Errorf("%w: generators are required", ErrInvalidState)
	}
	for id, count := range state.GeneratorCounts {
		if count < 0 || count > decimal.MaxExactInteger {
			return nil, fmt.Errorf("%w: invalid generator count for %q", ErrInvalidState, id)
		}
	}
	if state.ComputeCreditMS < 0 || state.ComputeCreditMS > decimal.MaxExactInteger ||
		state.ManualTokenMilli < 0 || state.ManualTokenMilli > decimal.MaxExactInteger {
		return nil, fmt.Errorf("%w: production integers exceed the exact domain", ErrInvalidState)
	}
	cursor, err := formatCursor(state.EvaluatedThrough)
	if err != nil {
		return nil, err
	}
	refilledAt, err := formatCursor(state.ManualTokenRefilledAt)
	if err != nil {
		return nil, fmt.Errorf("%w: manual_token_refilled_at is required", ErrInvalidState)
	}
	if state.ManualTokenRefilledAt.After(state.EvaluatedThrough) {
		return nil, fmt.Errorf("%w: manual_token_refilled_at exceeds evaluated_through", ErrInvalidState)
	}
	encoded, err := json.Marshal(stateV4{
		Balances: state.Ledger.Snapshot(), Generators: state.GeneratorCounts, EvaluatedThrough: cursor,
		ComputeCreditMS: state.ComputeCreditMS, ManualTokenMilli: state.ManualTokenMilli,
		ManualTokenRefilledAt: refilledAt,
	})
	if err != nil {
		return nil, fmt.Errorf("%w: encode: %v", ErrInvalidState, err)
	}
	return encoded, nil
}

func RestoreState(data []byte, version int, catalog *economy.Catalog, scope economy.Scope, migrationBaseline time.Time) (*State, error) {
	if version > CurrentVersion {
		return nil, fmt.Errorf("%w: save version %d is newer than supported version %d", ErrInvalidState, version, CurrentVersion)
	}
	if version < 1 {
		return nil, fmt.Errorf("%w: unsupported save version %d", ErrInvalidState, version)
	}
	if catalog == nil {
		return nil, fmt.Errorf("%w: nil catalog", ErrInvalidState)
	}

	var source stateV4
	if version == 1 {
		var legacy stateV1
		if err := decodeState(data, &legacy); err != nil {
			return nil, err
		}
		cursor, err := formatCursor(CanonicalServerTime(migrationBaseline))
		if err != nil {
			return nil, fmt.Errorf("%w: version-1 migration baseline: %v", ErrInvalidState, err)
		}
		source = stateV4{
			Balances: legacy.Balances, Generators: zeroGeneratorCounts(catalog, scope), EvaluatedThrough: cursor,
		}
	} else if version == 2 {
		var previous stateV2
		if err := decodeState(data, &previous); err != nil {
			return nil, err
		}
		source = stateV4{Balances: previous.Balances, Generators: previous.Generators, EvaluatedThrough: previous.EvaluatedThrough}
	} else if err := decodeState(data, &source); err != nil {
		return nil, err
	}

	if source.Balances == nil || source.Generators == nil {
		return nil, fmt.Errorf("%w: balances and generators are required", ErrInvalidState)
	}
	ledger, err := economy.RestoreLedger(catalog, scope, source.Balances)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidState, err)
	}
	counts, err := validateGeneratorCounts(catalog, scope, source.Generators)
	if err != nil {
		return nil, err
	}
	cursor, err := restoreCursor(source.EvaluatedThrough, version)
	if err != nil {
		return nil, err
	}
	if version < 3 {
		source.ComputeCreditMS = 0
		source.ManualTokenMilli = catalog.ManualPolicy().BucketCapMilli
		source.ManualTokenRefilledAt = source.EvaluatedThrough
	}
	refilledAt, err := restoreCursor(source.ManualTokenRefilledAt, version)
	if err != nil {
		return nil, fmt.Errorf("%w: manual_token_refilled_at: %v", ErrInvalidState, err)
	}
	if refilledAt.After(cursor) {
		return nil, fmt.Errorf("%w: manual_token_refilled_at exceeds evaluated_through", ErrInvalidState)
	}
	if err := validateProductionState(catalog, scope, source.ComputeCreditMS, source.ManualTokenMilli); err != nil {
		return nil, err
	}
	return &State{
		Ledger: ledger, GeneratorCounts: counts, EvaluatedThrough: cursor,
		ComputeCreditMS: source.ComputeCreditMS, ManualTokenMilli: source.ManualTokenMilli,
		ManualTokenRefilledAt: refilledAt,
	}, nil
}

func validateProductionState(catalog *economy.Catalog, scope economy.Scope, creditMS, tokenMilli int64) error {
	if creditMS < 0 || creditMS > decimal.MaxExactInteger || tokenMilli < 0 || tokenMilli > decimal.MaxExactInteger {
		return fmt.Errorf("%w: production integers exceed the exact domain", ErrInvalidState)
	}
	if scope != economy.ScopeCompany {
		if creditMS != 0 || tokenMilli != 0 {
			return fmt.Errorf("%w: production state is company-scoped", ErrInvalidState)
		}
		return nil
	}
	offlinePolicy, manualPolicy := catalog.OfflinePolicy(), catalog.ManualPolicy()
	if offlinePolicy.BankCapMS == 0 && creditMS != 0 || offlinePolicy.BankCapMS > 0 && creditMS > offlinePolicy.BankCapMS {
		return fmt.Errorf("%w: compute_credit_ms exceeds catalog cap", ErrInvalidState)
	}
	if manualPolicy.BucketCapMilli == 0 && tokenMilli != 0 || manualPolicy.BucketCapMilli > 0 && tokenMilli > manualPolicy.BucketCapMilli {
		return fmt.Errorf("%w: manual_token_milli exceeds catalog cap", ErrInvalidState)
	}
	return nil
}

func validateGeneratorCounts(catalog *economy.Catalog, scope economy.Scope, source map[string]int64) (map[string]int64, error) {
	expected := catalog.GeneratorClassesForScope(scope)
	if len(source) != len(expected) {
		return nil, fmt.Errorf("%w: generators contain %d entries, want %d", ErrInvalidState, len(source), len(expected))
	}
	counts := make(map[string]int64, len(expected))
	for _, definition := range expected {
		count, ok := source[definition.ID]
		if !ok || count < 0 || count > decimal.MaxExactInteger {
			return nil, fmt.Errorf("%w: invalid generator count for %q", ErrInvalidState, definition.ID)
		}
		counts[definition.ID] = count
	}
	return counts, nil
}

func zeroGeneratorCounts(catalog *economy.Catalog, scope economy.Scope) map[string]int64 {
	counts := make(map[string]int64)
	for _, definition := range catalog.GeneratorClassesForScope(scope) {
		counts[definition.ID] = 0
	}
	return counts
}

// CanonicalServerTime returns the only timestamp representation permitted for
// authoritative production cursors: UTC truncated to an exact millisecond.
func CanonicalServerTime(value time.Time) time.Time {
	return value.UTC().Truncate(time.Millisecond)
}

func formatCursor(value time.Time) (string, error) {
	if value.IsZero() {
		return "", fmt.Errorf("%w: evaluated_through is required", ErrInvalidState)
	}
	if !isCanonicalServerTime(value) {
		return "", fmt.Errorf("%w: production cursor must be canonical UTC whole milliseconds", ErrInvalidState)
	}
	return value.Format(time.RFC3339Nano), nil
}

func parseCursor(source string) (time.Time, error) {
	value, err := time.Parse(time.RFC3339Nano, source)
	if err != nil || value.Location() != time.UTC || value.Format(time.RFC3339Nano) != source {
		return time.Time{}, fmt.Errorf("%w: evaluated_through must be canonical UTC RFC3339Nano", ErrInvalidState)
	}
	return value, nil
}

func restoreCursor(source string, version int) (time.Time, error) {
	value, err := parseCursor(source)
	if err != nil {
		return time.Time{}, err
	}
	if version < millisecondCursorVersion {
		return CanonicalServerTime(value), nil
	}
	if !isCanonicalServerTime(value) {
		return time.Time{}, fmt.Errorf("%w: production cursor must be canonical UTC whole milliseconds", ErrInvalidState)
	}
	return value, nil
}

func isCanonicalServerTime(value time.Time) bool {
	return value.Location() == time.UTC && value.Nanosecond()%int(time.Millisecond) == 0
}

func decodeState(data []byte, destination any) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return fmt.Errorf("%w: decode: %v", ErrInvalidState, err)
	}
	return ensureJSONEnd(decoder)
}

func ensureJSONEnd(decoder *json.Decoder) error {
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			return fmt.Errorf("%w: trailing JSON value", ErrInvalidState)
		}
		return fmt.Errorf("%w: trailing data: %v", ErrInvalidState, err)
	}
	return nil
}

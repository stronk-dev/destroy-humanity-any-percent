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

const CurrentVersion = 2

var ErrInvalidState = errors.New("invalid saved state")

type State struct {
	Ledger           *economy.Ledger
	GeneratorCounts  map[string]int64
	EvaluatedThrough time.Time
}

type stateV1 struct {
	Balances map[string]string `json:"balances"`
}

type stateV2 struct {
	Balances         map[string]string `json:"balances"`
	Generators       map[string]int64  `json:"generators"`
	EvaluatedThrough string            `json:"evaluated_through"`
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
	cursor, err := formatCursor(state.EvaluatedThrough)
	if err != nil {
		return nil, err
	}
	encoded, err := json.Marshal(stateV2{
		Balances: state.Ledger.Snapshot(), Generators: state.GeneratorCounts, EvaluatedThrough: cursor,
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

	var source stateV2
	if version == 1 {
		var legacy stateV1
		if err := decodeState(data, &legacy); err != nil {
			return nil, err
		}
		cursor, err := formatCursor(migrationBaseline)
		if err != nil {
			return nil, fmt.Errorf("%w: version-1 migration baseline: %v", ErrInvalidState, err)
		}
		source = stateV2{
			Balances: legacy.Balances, Generators: zeroGeneratorCounts(catalog, scope), EvaluatedThrough: cursor,
		}
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
	cursor, err := parseCursor(source.EvaluatedThrough)
	if err != nil {
		return nil, err
	}
	return &State{Ledger: ledger, GeneratorCounts: counts, EvaluatedThrough: cursor}, nil
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

func formatCursor(value time.Time) (string, error) {
	if value.IsZero() {
		return "", fmt.Errorf("%w: evaluated_through is required", ErrInvalidState)
	}
	return value.UTC().Format(time.RFC3339Nano), nil
}

func parseCursor(source string) (time.Time, error) {
	value, err := time.Parse(time.RFC3339Nano, source)
	if err != nil || value.Location() != time.UTC || value.Format(time.RFC3339Nano) != source {
		return time.Time{}, fmt.Errorf("%w: evaluated_through must be canonical UTC RFC3339Nano", ErrInvalidState)
	}
	return value, nil
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

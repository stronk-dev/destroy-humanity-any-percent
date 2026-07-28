package save

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"cloud-clicker/server/economy"
)

const CurrentVersion = 1

var ErrInvalidState = errors.New("invalid saved state")

type StateV1 struct {
	Balances map[string]string `json:"balances"`
}

func ConstantsHash(catalogBytes []byte) string {
	digest := sha256.Sum256(catalogBytes)
	return "sha256:" + hex.EncodeToString(digest[:])
}

func EncodeLedger(ledger *economy.Ledger) ([]byte, error) {
	if ledger == nil {
		return nil, fmt.Errorf("%w: nil ledger", ErrInvalidState)
	}
	encoded, err := json.Marshal(StateV1{Balances: ledger.Snapshot()})
	if err != nil {
		return nil, fmt.Errorf("%w: encode: %v", ErrInvalidState, err)
	}
	return encoded, nil
}

func RestoreLedger(data []byte, version int, catalog *economy.Catalog, scope economy.Scope) (*economy.Ledger, error) {
	if version > CurrentVersion {
		return nil, fmt.Errorf("%w: save version %d is newer than supported version %d", ErrInvalidState, version, CurrentVersion)
	}
	if version < 1 {
		return nil, fmt.Errorf("%w: unsupported save version %d", ErrInvalidState, version)
	}
	migrated, err := migrateState(data, version)
	if err != nil {
		return nil, err
	}

	decoder := json.NewDecoder(bytes.NewReader(migrated))
	decoder.DisallowUnknownFields()
	var state StateV1
	if err := decoder.Decode(&state); err != nil {
		return nil, fmt.Errorf("%w: decode: %v", ErrInvalidState, err)
	}
	if err := ensureJSONEnd(decoder); err != nil {
		return nil, err
	}
	if state.Balances == nil {
		return nil, fmt.Errorf("%w: balances is required", ErrInvalidState)
	}
	ledger, err := economy.RestoreLedger(catalog, scope, state.Balances)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidState, err)
	}
	return ledger, nil
}

func migrateState(data []byte, version int) ([]byte, error) {
	current := append([]byte(nil), data...)
	for version < CurrentVersion {
		return nil, fmt.Errorf("%w: missing migration from version %d", ErrInvalidState, version)
	}
	return current, nil
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

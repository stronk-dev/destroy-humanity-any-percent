// Package minigameapi owns the pinned structural policy that activates the
// authenticated minigame surface for a catalog bundle.
package minigameapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"regexp"
)

const SchemaVersion = 1

var (
	ErrInvalidCatalog = errors.New("invalid minigame API catalog")
	mechanicalID      = regexp.MustCompile(`^[a-z][a-z0-9_]*(?:\.[a-z][a-z0-9_]*)*$`)
	engineVersion     = regexp.MustCompile(`^[1-9][0-9]*\.[0-9]+\.[0-9]+$`)
)

type Operation struct {
	OperationID string `json:"operation_id"`
	Version     int    `json:"version"`
}

type Tenant struct {
	EngineRef     string `json:"engine_ref"`
	EngineVersion string `json:"engine_version"`
	MinigameID    string `json:"minigame_id"`
}

type Catalog struct {
	SchemaVersion int         `json:"schema_version"`
	Operations    []Operation `json:"operations"`
	Tenants       []Tenant    `json:"tenants"`
}

var requiredOperations = [...]string{
	"create_minigame_session",
	"get_current_minigame_session",
	"play_minigame_command",
	"resolve_minigame_session",
}

func LoadCatalog(data []byte) (*Catalog, error) {
	if !uniqueJSONKeys(data) {
		return nil, ErrInvalidCatalog
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var catalog Catalog
	if decoder.Decode(&catalog) != nil {
		return nil, ErrInvalidCatalog
	}
	var trailing any
	if !errors.Is(decoder.Decode(&trailing), io.EOF) || catalog.SchemaVersion != SchemaVersion ||
		len(catalog.Operations) != len(requiredOperations) || len(catalog.Tenants) == 0 {
		return nil, ErrInvalidCatalog
	}
	for index, operation := range catalog.Operations {
		if operation.OperationID != requiredOperations[index] || operation.Version != 1 {
			return nil, ErrInvalidCatalog
		}
	}
	prior := ""
	for _, tenant := range catalog.Tenants {
		key := tenant.EngineRef + "\x00" + tenant.EngineVersion + "\x00" + tenant.MinigameID
		if !mechanicalID.MatchString(tenant.EngineRef) || !engineVersion.MatchString(tenant.EngineVersion) ||
			!mechanicalID.MatchString(tenant.MinigameID) || key <= prior {
			return nil, ErrInvalidCatalog
		}
		prior = key
	}
	return &catalog, nil
}

func (catalog *Catalog) SupportsTenant(minigameID, engineRef, version string) bool {
	if catalog == nil {
		return false
	}
	for _, tenant := range catalog.Tenants {
		if tenant.MinigameID == minigameID && tenant.EngineRef == engineRef && tenant.EngineVersion == version {
			return true
		}
	}
	return false
}

func uniqueJSONKeys(data []byte) bool {
	decoder := json.NewDecoder(bytes.NewReader(data))
	var readValue func() bool
	readValue = func() bool {
		token, err := decoder.Token()
		if err != nil {
			return false
		}
		delimiter, compound := token.(json.Delim)
		if !compound {
			return true
		}
		switch delimiter {
		case '{':
			seen := map[string]struct{}{}
			for decoder.More() {
				keyToken, keyErr := decoder.Token()
				key, valid := keyToken.(string)
				if keyErr != nil || !valid {
					return false
				}
				if _, duplicate := seen[key]; duplicate {
					return false
				}
				seen[key] = struct{}{}
				if !readValue() {
					return false
				}
			}
			end, endErr := decoder.Token()
			return endErr == nil && end == json.Delim('}')
		case '[':
			for decoder.More() {
				if !readValue() {
					return false
				}
			}
			end, endErr := decoder.Token()
			return endErr == nil && end == json.Delim(']')
		default:
			return false
		}
	}
	if !readValue() {
		return false
	}
	var trailing any
	return errors.Is(decoder.Decode(&trailing), io.EOF)
}

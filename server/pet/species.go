package pet

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"

	"cloud-clicker/server/decimal"
)

// Pet Adoption v1 (rfc/pet-adoption-v1.md PA2/PA4): the pet_species artifact
// and the grinding-proof, nonce-derived adoption draws.

const (
	SpeciesSchemaVersion   = 1
	AvailabilityStarter    = "starter"
	VisualFamilyCat        = "cat"
	AdoptionNonceBytes     = 16
	maxSpeciesRows         = 64
	maxSpeciesListElements = 256
)

var (
	ErrInvalidSpecies = errors.New("invalid pet_species artifact")
	ErrInvalidDraw    = errors.New("invalid pet adoption draw")

	// TemperamentOrder is the canonical combat Temperament order (PA2,
	// PC4/C15): lazy, playful, curious, sassy, shy, chaotic.
	TemperamentOrder = []string{"lazy", "playful", "curious", "sassy", "shy", "chaotic"}

	speciesIDPattern = regexp.MustCompile(`^pet_species\.[a-z][a-z0-9_]*$`)
	paletteIDPattern = regexp.MustCompile(`^pet_palette\.[a-z0-9_]+$`)
	nameKeyPattern   = regexp.MustCompile(`^pet\.name\.[a-z0-9_]+\.[a-z0-9_]+$`)
	copyKeyPattern   = regexp.MustCompile(`^[a-z][a-z0-9_]*(?:\.[a-z][a-z0-9_]*)*$`)
)

type SpeciesRow struct {
	SpeciesID           string
	Availability        string
	VisualFamily        string
	AllowedTemperaments []string
	PaletteIDs          []string
	NameKeys            []string
	NameCopyKey         string
	DescriptionCopyKey  string
}

type SpeciesCatalog struct {
	MaxPetsPerFounder int64
	Species           []SpeciesRow
	byID              map[string]int
}

// SpeciesDeclarations are the copy registries the artifact is validated
// against: every name, species-name and description key must exist and use
// the companion tone (PA2, PA8.6).
type SpeciesDeclarations struct {
	CopyKeys      map[string]struct{}
	CompanionKeys map[string]struct{}
}

type rawSpeciesCatalog struct {
	SchemaVersion     *int              `json:"schema_version"`
	MaxPetsPerFounder *int64            `json:"max_pets_per_founder"`
	Species           []json.RawMessage `json:"species"`
}

type rawSpeciesRow struct {
	SpeciesID           string   `json:"species_id"`
	Availability        string   `json:"availability"`
	VisualFamily        string   `json:"visual_family"`
	AllowedTemperaments []string `json:"allowed_temperaments"`
	PaletteIDs          []string `json:"palette_ids"`
	NameKeys            []string `json:"name_keys"`
	NameCopyKey         string   `json:"name_copy_key"`
	DescriptionCopyKey  string   `json:"description_copy_key"`
}

func LoadSpeciesCatalog(data []byte, declarations SpeciesDeclarations) (*SpeciesCatalog, error) {
	if !uniqueStateJSONKeys(data) {
		return nil, fmt.Errorf("%w: duplicate keys", ErrInvalidSpecies)
	}
	var top map[string]json.RawMessage
	if json.Unmarshal(data, &top) != nil || !hasRawKeys(top, []string{"schema_version", "max_pets_per_founder", "species"}) {
		return nil, fmt.Errorf("%w: top-level keys are not exact", ErrInvalidSpecies)
	}
	var raw rawSpeciesCatalog
	if err := strictSpeciesDecode(data, &raw); err != nil {
		return nil, err
	}
	if raw.SchemaVersion == nil || *raw.SchemaVersion != SpeciesSchemaVersion || raw.MaxPetsPerFounder == nil ||
		*raw.MaxPetsPerFounder < 1 || *raw.MaxPetsPerFounder > decimal.MaxExactInteger || len(raw.Species) == 0 || len(raw.Species) > maxSpeciesRows {
		return nil, fmt.Errorf("%w: invalid schema, cap, or species list", ErrInvalidSpecies)
	}
	catalog := &SpeciesCatalog{MaxPetsPerFounder: *raw.MaxPetsPerFounder, byID: map[string]int{}}
	starters := 0
	for index, rowBytes := range raw.Species {
		var rowKeys map[string]json.RawMessage
		if json.Unmarshal(rowBytes, &rowKeys) != nil || !hasRawKeys(rowKeys, []string{"species_id", "availability", "visual_family", "allowed_temperaments", "palette_ids", "name_keys", "name_copy_key", "description_copy_key"}) {
			return nil, fmt.Errorf("%w: species row keys are not exact", ErrInvalidSpecies)
		}
		var row rawSpeciesRow
		if err := strictSpeciesDecode(rowBytes, &row); err != nil {
			return nil, err
		}
		if !speciesIDPattern.MatchString(row.SpeciesID) || index > 0 && catalog.Species[index-1].SpeciesID >= row.SpeciesID {
			return nil, fmt.Errorf("%w: species ids must be namespaced, sorted, and unique", ErrInvalidSpecies)
		}
		if row.Availability != AvailabilityStarter || row.VisualFamily != VisualFamilyCat {
			return nil, fmt.Errorf("%w: unknown availability or visual family", ErrInvalidSpecies)
		}
		starters++
		if err := validateTemperaments(row.AllowedTemperaments); err != nil {
			return nil, err
		}
		if !sortedUniqueMatching(row.PaletteIDs, paletteIDPattern) || !sortedUniqueMatching(row.NameKeys, nameKeyPattern) {
			return nil, fmt.Errorf("%w: palette_ids or name_keys are not sorted, unique, and well-formed", ErrInvalidSpecies)
		}
		for _, key := range append(append([]string{}, row.NameKeys...), row.NameCopyKey, row.DescriptionCopyKey) {
			if !copyKeyPattern.MatchString(key) {
				return nil, fmt.Errorf("%w: copy key %q is not mechanical", ErrInvalidSpecies, key)
			}
			if _, ok := declarations.CopyKeys[key]; !ok {
				return nil, fmt.Errorf("%w: copy key %q is not registered", ErrInvalidSpecies, key)
			}
			if _, ok := declarations.CompanionKeys[key]; !ok {
				return nil, fmt.Errorf("%w: copy key %q is not companion tone", ErrInvalidSpecies, key)
			}
		}
		catalog.byID[row.SpeciesID] = index
		catalog.Species = append(catalog.Species, SpeciesRow{SpeciesID: row.SpeciesID, Availability: row.Availability, VisualFamily: row.VisualFamily,
			AllowedTemperaments: append([]string{}, row.AllowedTemperaments...), PaletteIDs: append([]string{}, row.PaletteIDs...),
			NameKeys: append([]string{}, row.NameKeys...), NameCopyKey: row.NameCopyKey, DescriptionCopyKey: row.DescriptionCopyKey})
	}
	if starters != 1 {
		return nil, fmt.Errorf("%w: exactly one starter row is required", ErrInvalidSpecies)
	}
	return catalog, nil
}

func (catalog *SpeciesCatalog) Row(speciesID string) (SpeciesRow, bool) {
	if catalog == nil {
		return SpeciesRow{}, false
	}
	index, ok := catalog.byID[speciesID]
	if !ok {
		return SpeciesRow{}, false
	}
	return catalog.Species[index], true
}

func (row SpeciesRow) HasName(nameKey string) bool { return contains(row.NameKeys, nameKey) }

func (row SpeciesRow) HasPalette(paletteID string) bool { return contains(row.PaletteIDs, paletteID) }

func (row SpeciesRow) HasTemperament(temperament string) bool {
	return contains(row.AllowedTemperaments, temperament)
}

func validateTemperaments(values []string) error {
	if len(values) == 0 || len(values) > len(TemperamentOrder) {
		return fmt.Errorf("%w: allowed_temperaments must be non-empty", ErrInvalidSpecies)
	}
	cursor := 0
	for _, value := range values {
		found := false
		for cursor < len(TemperamentOrder) {
			candidate := TemperamentOrder[cursor]
			cursor++
			if candidate == value {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("%w: temperament %q is unknown, duplicated, or out of canonical order", ErrInvalidSpecies, value)
		}
	}
	return nil
}

func sortedUniqueMatching(values []string, pattern *regexp.Regexp) bool {
	if len(values) == 0 || len(values) > maxSpeciesListElements {
		return false
	}
	for index, value := range values {
		if !pattern.MatchString(value) || index > 0 && values[index-1] >= value {
			return false
		}
	}
	return true
}

func contains(values []string, value string) bool {
	for _, candidate := range values {
		if candidate == value {
			return true
		}
	}
	return false
}

func strictSpeciesDecode(data []byte, target any) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidSpecies, err)
	}
	var trailing any
	if !errors.Is(decoder.Decode(&trailing), io.EOF) {
		return fmt.Errorf("%w: trailing data", ErrInvalidSpecies)
	}
	return nil
}

// AdoptionDraws are the values PA4.3–PA4.5 derive from the frozen server
// nonce. The client-chosen intent_id never enters any hash input.
type AdoptionDraws struct {
	PetID       string
	Temperament string
	PaletteID   string
}

// DrawAdoption derives the pet ID, temperament and palette. founderID is the
// canonical UUID string of the adopting Founder; nonceHex is 32 lowercase hex
// characters; serverTimeMS is the command's server time in unix ms.
func DrawAdoption(founderID string, nonceHex string, serverTimeMS int64, row SpeciesRow) (AdoptionDraws, error) {
	founder, err := uuidBytes(founderID)
	if err != nil {
		return AdoptionDraws{}, err
	}
	nonce, err := hex.DecodeString(nonceHex)
	if err != nil || len(nonce) != AdoptionNonceBytes || hex.EncodeToString(nonce) != nonceHex {
		return AdoptionDraws{}, fmt.Errorf("%w: nonce must be %d lowercase hex bytes", ErrInvalidDraw, AdoptionNonceBytes)
	}
	if serverTimeMS < 0 || serverTimeMS > 0xffffffffffff || len(row.AllowedTemperaments) == 0 || len(row.PaletteIDs) == 0 {
		return AdoptionDraws{}, fmt.Errorf("%w: invalid time or species row", ErrInvalidDraw)
	}
	idDigest := sha256.Sum256(append(append([]byte("pet.id.v1"), founder...), nonce...))
	var id [16]byte
	for index := 0; index < 6; index++ {
		id[index] = byte(uint64(serverTimeMS) >> (8 * (5 - index)))
	}
	id[6] = 0x70 | idDigest[0]&0x0f
	id[7] = idDigest[1]
	id[8] = 0x80 | idDigest[2]&0x3f
	copy(id[9:], idDigest[3:10])
	return AdoptionDraws{
		PetID:       formatUUID(id),
		Temperament: row.AllowedTemperaments[labelledIndex("pet.temperament.v1", nonce, len(row.AllowedTemperaments))],
		PaletteID:   row.PaletteIDs[labelledIndex("pet.palette.v1", nonce, len(row.PaletteIDs))],
	}, nil
}

func labelledIndex(label string, nonce []byte, size int) int {
	digest := sha256.Sum256(append([]byte(label), nonce...))
	return int(binary.BigEndian.Uint64(digest[:8]) % uint64(size))
}

func uuidBytes(value string) ([]byte, error) {
	if len(value) != 36 || value[8] != '-' || value[13] != '-' || value[18] != '-' || value[23] != '-' {
		return nil, fmt.Errorf("%w: founder id is not a canonical UUID", ErrInvalidDraw)
	}
	compact := value[0:8] + value[9:13] + value[14:18] + value[19:23] + value[24:36]
	decoded, err := hex.DecodeString(compact)
	if err != nil || hex.EncodeToString(decoded) != compact {
		return nil, fmt.Errorf("%w: founder id is not a lowercase canonical UUID", ErrInvalidDraw)
	}
	return decoded, nil
}

func formatUUID(id [16]byte) string {
	text := hex.EncodeToString(id[:])
	return text[0:8] + "-" + text[8:12] + "-" + text[12:16] + "-" + text[16:20] + "-" + text[20:32]
}

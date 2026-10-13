package combat

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"unicode/utf8"
)

var ErrInvalidSpecies = errors.New("invalid combat species record")

// SpeciesStats owns C2's five signed-int32 identity ceilings. This structural
// boundary does not invent balance values, minimum stats or care modifiers.
type SpeciesStats struct {
	HP         int32 `json:"hp"`
	ATK        int32 `json:"atk"`
	DEF        int32 `json:"def"`
	SPD        int32 `json:"spd"`
	StaminaMax int32 `json:"stamina_max"`
}

type SpeciesRecord struct {
	ID          string       `json:"id"`
	Temperament Temperament  `json:"temperament"`
	Runtime     Runtime      `json:"runtime"`
	Stats       SpeciesStats `json:"stats"`
	Moves       [4]string    `json:"moves"`
	Tags        []string     `json:"tags"`
}

// DecodeSpeciesRecord consumes C2's exact species row, with moves declared by
// the caller's validated pinned catalog. It does not create that catalog or
// infer ownership, effect legality, synergy vocabulary or a stat balance policy.
func DecodeSpeciesRecord(data []byte, moves map[string]struct{}) (SpeciesRecord, error) {
	fields, err := speciesObject(data, "id", "temperament", "runtime", "stats", "moves", "tags")
	if err != nil {
		return SpeciesRecord{}, ErrInvalidSpecies
	}
	var record SpeciesRecord
	if !speciesString(fields["id"], &record.ID) || !speciesString(fields["temperament"], (*string)(&record.Temperament)) ||
		!speciesString(fields["runtime"], (*string)(&record.Runtime)) {
		return SpeciesRecord{}, ErrInvalidSpecies
	}
	if _, err := NewIdentity(record.Temperament, record.Runtime); err != nil {
		return SpeciesRecord{}, ErrInvalidSpecies
	}
	stats, err := speciesObject(fields["stats"], "hp", "atk", "def", "spd", "stamina_max")
	if err != nil {
		return SpeciesRecord{}, ErrInvalidSpecies
	}
	for key, target := range map[string]*int32{"hp": &record.Stats.HP, "atk": &record.Stats.ATK,
		"def": &record.Stats.DEF, "spd": &record.Stats.SPD, "stamina_max": &record.Stats.StaminaMax} {
		raw := bytes.TrimSpace(stats[key])
		if bytes.Equal(raw, []byte("null")) || json.Unmarshal(raw, target) != nil {
			return SpeciesRecord{}, ErrInvalidSpecies
		}
	}
	moveIDs, err := speciesStrings(fields["moves"])
	if err != nil || len(moveIDs) != len(record.Moves) {
		return SpeciesRecord{}, ErrInvalidSpecies
	}
	for index, id := range moveIDs {
		if _, known := moves[id]; !known {
			return SpeciesRecord{}, ErrInvalidSpecies
		}
		record.Moves[index] = id
	}
	if record.Tags, err = speciesStrings(fields["tags"]); err != nil {
		return SpeciesRecord{}, ErrInvalidSpecies
	}
	return record, nil
}

func speciesObject(data []byte, keys ...string) (map[string]json.RawMessage, error) {
	if !utf8.Valid(data) {
		return nil, ErrInvalidSpecies
	}
	allowed := make(map[string]bool, len(keys))
	for _, key := range keys {
		allowed[key] = true
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	start, err := decoder.Token()
	if err != nil || start != json.Delim('{') {
		return nil, ErrInvalidSpecies
	}
	fields := make(map[string]json.RawMessage, len(keys))
	for decoder.More() {
		token, err := decoder.Token()
		key, valid := token.(string)
		if err != nil || !valid || !allowed[key] {
			return nil, ErrInvalidSpecies
		}
		if _, duplicate := fields[key]; duplicate {
			return nil, ErrInvalidSpecies
		}
		var raw json.RawMessage
		if decoder.Decode(&raw) != nil {
			return nil, ErrInvalidSpecies
		}
		fields[key] = raw
	}
	end, err := decoder.Token()
	if err != nil || end != json.Delim('}') || len(fields) != len(keys) {
		return nil, ErrInvalidSpecies
	}
	var trailing any
	if !errors.Is(decoder.Decode(&trailing), io.EOF) {
		return nil, ErrInvalidSpecies
	}
	return fields, nil
}

func speciesString(data []byte, target *string) bool {
	var value *string
	if json.Unmarshal(data, &value) != nil || value == nil || !wellFormedReference(data) {
		return false
	}
	*target = *value
	return true
}

func speciesStrings(data []byte) ([]string, error) {
	var raw []json.RawMessage
	if json.Unmarshal(data, &raw) != nil || raw == nil {
		return nil, ErrInvalidSpecies
	}
	values := make([]string, len(raw))
	for index, item := range raw {
		if !speciesString(item, &values[index]) {
			return nil, ErrInvalidSpecies
		}
	}
	return values, nil
}

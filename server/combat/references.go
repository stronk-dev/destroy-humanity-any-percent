package combat

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"strconv"
	"unicode/utf8"
)

var ErrInvalidReferences = errors.New("invalid combat loadout references")

// DecodeTeamReferences checks C2's ordered, three-species reference envelope.
// declarations must come from the caller's validated, pinned catalog. This is
// not pet ownership, instance selection, or engine-specific loadout legality.
func DecodeTeamReferences(data []byte, declarations map[string]struct{}) ([3]string, error) {
	var team [3]string
	ids, err := decodeReferences(data, declarations, len(team))
	if err != nil {
		return team, err
	}
	copy(team[:], ids)
	return team, nil
}

// DecodeDeckReferences checks C2's ordered, eight-card reference envelope.
// It preserves slots; it does not sort, deduplicate, or choose cycle legality.
func DecodeDeckReferences(data []byte, declarations map[string]struct{}) ([8]string, error) {
	var deck [8]string
	ids, err := decodeReferences(data, declarations, len(deck))
	if err != nil {
		return deck, err
	}
	copy(deck[:], ids)
	return deck, nil
}

func decodeReferences(data []byte, declarations map[string]struct{}, count int) ([]string, error) {
	if !utf8.Valid(data) {
		return nil, ErrInvalidReferences
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	start, err := decoder.Token()
	if err != nil || start != json.Delim('{') {
		return nil, ErrInvalidReferences
	}
	fields := make(map[string]json.RawMessage, 2)
	for decoder.More() {
		token, err := decoder.Token()
		key, ok := token.(string)
		if err != nil || !ok || (key != "v" && key != "ids") {
			return nil, ErrInvalidReferences
		}
		if _, duplicate := fields[key]; duplicate {
			return nil, ErrInvalidReferences
		}
		var raw json.RawMessage
		if decoder.Decode(&raw) != nil {
			return nil, ErrInvalidReferences
		}
		fields[key] = raw
	}
	end, err := decoder.Token()
	if err != nil || end != json.Delim('}') || len(fields) != 2 || string(bytes.TrimSpace(fields["v"])) != "1" {
		return nil, ErrInvalidReferences
	}
	var trailing any
	if !errors.Is(decoder.Decode(&trailing), io.EOF) {
		return nil, ErrInvalidReferences
	}
	var rows []json.RawMessage
	if json.Unmarshal(fields["ids"], &rows) != nil || len(rows) != count {
		return nil, ErrInvalidReferences
	}
	ids := make([]string, count)
	for index, raw := range rows {
		var id *string
		if json.Unmarshal(raw, &id) != nil || id == nil || !wellFormedReference(raw) {
			return nil, ErrInvalidReferences
		}
		if _, known := declarations[*id]; !known {
			return nil, ErrInvalidReferences
		}
		ids[index] = *id
	}
	return ids, nil
}

// encoding/json otherwise replaces lone UTF-16 surrogates with U+FFFD. Do not
// turn malformed wire text into a different, possibly declared reference.
// json.Unmarshal has already checked the string's syntax and escape lengths.
func wellFormedReference(raw []byte) bool {
	for index := 0; index < len(raw); index++ {
		if raw[index] != '\\' {
			continue
		}
		index++
		if raw[index] != 'u' {
			continue
		}
		value, _ := strconv.ParseUint(string(raw[index+1:index+5]), 16, 16)
		index += 4
		if value >= 0xdc00 && value <= 0xdfff {
			return false
		}
		if value >= 0xd800 && value <= 0xdbff {
			if index+6 >= len(raw) || raw[index+1] != '\\' || raw[index+2] != 'u' {
				return false
			}
			low, _ := strconv.ParseUint(string(raw[index+3:index+7]), 16, 16)
			if low < 0xdc00 || low > 0xdfff {
				return false
			}
			index += 6
		}
	}
	return true
}

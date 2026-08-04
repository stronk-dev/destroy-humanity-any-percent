package publicapi

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"regexp"
)

type BoardVariables struct {
	Advisor  int     `json:"advisor"`
	Commons  int     `json:"commons"`
	Faction  *string `json:"faction"`
	Glitched int     `json:"glitched"`
}

var factionIDPattern = regexp.MustCompile(`^[a-z][a-z0-9_]*(?:\.[a-z][a-z0-9_]*)*$`)

func EncodeBoardVariables(value BoardVariables) (string, error) {
	if !validBoardVariables(value) {
		return "", ErrInvalidCursor
	}
	data, err := json.Marshal(value)
	if err != nil {
		return "", ErrInvalidCursor
	}
	return base64.RawURLEncoding.EncodeToString(data), nil
}

func DecodeBoardVariables(encoded string) (BoardVariables, error) {
	if len(encoded) == 0 || len(encoded) > 512 {
		return BoardVariables{}, ErrInvalidCursor
	}
	data, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return BoardVariables{}, ErrInvalidCursor
	}
	var value BoardVariables
	if err := decodeCursorStrict(data, &value); err != nil || !validBoardVariables(value) {
		return BoardVariables{}, ErrInvalidCursor
	}
	canonical, _ := json.Marshal(value)
	if !bytes.Equal(canonical, data) {
		return BoardVariables{}, ErrInvalidCursor
	}
	return value, nil
}

func validBoardVariables(value BoardVariables) bool {
	return (value.Advisor == 0 || value.Advisor == 1) && (value.Commons == 0 || value.Commons == 1) &&
		(value.Glitched == 0 || value.Glitched == 1) && (value.Faction == nil || factionIDPattern.MatchString(*value.Faction))
}

package publicapi

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"regexp"
)

type BoardVariables struct {
	Advisor  int     `json:"advisor"`
	Commons  int     `json:"commons"`
	Faction  *string `json:"faction"`
	Glitched int     `json:"glitched"`
}

type BoardFilter struct {
	Category     string
	Variables    BoardVariables
	EpochID      int64
	MandateLevel int
	Limit        int
}

type boardFilterWire struct {
	Category     string         `json:"category"`
	EpochID      int64          `json:"epoch"`
	Limit        int            `json:"limit"`
	MandateLevel int            `json:"mandate"`
	Variables    BoardVariables `json:"variables"`
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

func EncodeBoardFilter(value BoardFilter) ([]byte, error) {
	if !factionIDPattern.MatchString(value.Category) || !validBoardVariables(value.Variables) || value.EpochID < 1 || value.MandateLevel < 0 || value.MandateLevel > 20 || value.Limit < 1 || value.Limit > 100 {
		return nil, ErrInvalidCursor
	}
	encoded, err := json.Marshal(boardFilterWire{Category: value.Category, EpochID: value.EpochID, Limit: value.Limit, MandateLevel: value.MandateLevel, Variables: value.Variables})
	if err != nil || !canonicalJSON(encoded) {
		return nil, ErrInvalidCursor
	}
	return encoded, nil
}

func BoardFilterSHA256(value BoardFilter) (string, error) {
	encoded, err := EncodeBoardFilter(value)
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256(encoded)
	return hex.EncodeToString(digest[:]), nil
}

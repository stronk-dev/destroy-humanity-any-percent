package minigame

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"sort"
)

var ErrInvalidCatalog = errors.New("invalid minigame catalog")

// Catalog is the activation surface owned by the pinned minigames artifact.
// Content rows remain absent until a balance mint; the artifact still closes
// the ID and rating-season domains used by Founder save v17.
type Catalog struct {
	minigameIDs   []string
	ratingSeasons []string
}

type catalogWire struct {
	SchemaVersion int      `json:"schema_version"`
	MinigameIDs   []string `json:"minigame_ids"`
	RatingSeasons []string `json:"rating_seasons"`
}

func LoadCatalog(data []byte) (*Catalog, error) {
	if !uniqueJSONKeys(data) || !hasExactJSONKeys(data, "minigame_ids", "rating_seasons", "schema_version") {
		return nil, ErrInvalidCatalog
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var wire catalogWire
	if decoder.Decode(&wire) != nil || wire.SchemaVersion != 1 {
		return nil, ErrInvalidCatalog
	}
	var trailing any
	if !errors.Is(decoder.Decode(&trailing), io.EOF) || !sortedMechanical(wire.MinigameIDs) || !sortedMechanical(wire.RatingSeasons) {
		return nil, ErrInvalidCatalog
	}
	return &Catalog{minigameIDs: append([]string(nil), wire.MinigameIDs...), ratingSeasons: append([]string(nil), wire.RatingSeasons...)}, nil
}

func (catalog *Catalog) MinigameIDs() []string {
	if catalog == nil {
		return nil
	}
	return append([]string(nil), catalog.minigameIDs...)
}

func (catalog *Catalog) RatingSeasons() []string {
	if catalog == nil {
		return nil
	}
	return append([]string(nil), catalog.ratingSeasons...)
}

func (catalog *Catalog) HasRatingSeason(value string) bool {
	if catalog == nil {
		return false
	}
	index := sort.SearchStrings(catalog.ratingSeasons, value)
	return index < len(catalog.ratingSeasons) && catalog.ratingSeasons[index] == value
}

func sortedMechanical(values []string) bool {
	if values == nil {
		return false
	}
	for index, value := range values {
		if !mechanicalPattern.MatchString(value) || index > 0 && values[index-1] >= value {
			return false
		}
	}
	return true
}

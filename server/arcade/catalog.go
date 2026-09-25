// Package arcade owns the pinned Demo Disc Arcade artifact and the pure
// transition engines of its cover-disc toys (rfc/minigame-demo-disc-arcade.md
// AR1–AR4): the `mine_grid` deduction grid and `snake`.
package arcade

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"regexp"
)

const (
	SchemaVersion = 1
	EngineVersion = "1.0.0"
	// StageTierMax bounds a stage's min_tier to the shipped tier range.
	StageTierMax = 9
)

var (
	ErrInvalidCatalog = errors.New("invalid arcade catalog")
	idPattern         = regexp.MustCompile(`^[a-z][a-z0-9_]*(?:\.[a-z][a-z0-9_]*)*$`)
)

type Stage struct {
	StageID      string   `json:"stage_id"`
	MinTier      int64    `json:"min_tier"`
	TitleCopyKey string   `json:"title_copy_key"`
	Toys         []string `json:"toys"`
}

type Container struct {
	Stages []Stage `json:"stages"`
}

type Preset struct {
	PresetID string `json:"preset_id"`
	Width    int64  `json:"width"`
	Height   int64  `json:"height"`
	Mines    int64  `json:"mines"`
	CopyKey  string `json:"copy_key"`
}

type MineGridContent struct {
	Presets []Preset `json:"presets"`
}

type SnakeContent struct {
	Width              int64 `json:"width"`
	Height             int64 `json:"height"`
	StartLength        int64 `json:"start_length"`
	GrowthPerFood      int64 `json:"growth_per_food"`
	MaxTicksPerAdvance int64 `json:"max_ticks_per_advance"`
	// PresentationTickMS is presentation-only: validated, never read (AR4.1).
	PresentationTickMS int64 `json:"presentation_tick_ms"`
}

type Catalog struct {
	SchemaVersion int             `json:"schema_version"`
	Container     Container       `json:"container"`
	MineGrid      MineGridContent `json:"mine_grid"`
	Snake         SnakeContent    `json:"snake"`
}

type Declarations struct {
	CopyKeys map[string]struct{}
}

// LoadCatalog enforces AR1.2/AR3.1/AR4.1 exactly: closed keys at every
// level, byte-sorted rows, and the stated loader bounds. Cross-artifact
// checks (toy IDs against minigames definitions) belong to composition.
func LoadCatalog(data []byte, declarations Declarations) (*Catalog, error) {
	if len(declarations.CopyKeys) == 0 || !uniqueJSONKeys(data) ||
		!hasExactJSONKeys(data, "schema_version", "container", "mine_grid", "snake") {
		return nil, ErrInvalidCatalog
	}
	var raw struct {
		Container json.RawMessage `json:"container"`
		MineGrid  json.RawMessage `json:"mine_grid"`
		Snake     json.RawMessage `json:"snake"`
	}
	if json.Unmarshal(data, &raw) != nil || !hasExactJSONKeys(raw.Container, "stages") || !hasExactJSONKeys(raw.MineGrid, "presets") ||
		!hasExactJSONKeys(raw.Snake, "width", "height", "start_length", "growth_per_food", "max_ticks_per_advance", "presentation_tick_ms") {
		return nil, ErrInvalidCatalog
	}
	var rows struct {
		Container struct {
			Stages []json.RawMessage `json:"stages"`
		} `json:"container"`
		MineGrid struct {
			Presets []json.RawMessage `json:"presets"`
		} `json:"mine_grid"`
	}
	if json.Unmarshal(data, &rows) != nil {
		return nil, ErrInvalidCatalog
	}
	for _, row := range rows.Container.Stages {
		if !hasExactJSONKeys(row, "stage_id", "min_tier", "title_copy_key", "toys") {
			return nil, ErrInvalidCatalog
		}
	}
	for _, row := range rows.MineGrid.Presets {
		if !hasExactJSONKeys(row, "preset_id", "width", "height", "mines", "copy_key") {
			return nil, ErrInvalidCatalog
		}
	}
	var catalog Catalog
	if strictDecode(data, &catalog) != nil || catalog.SchemaVersion != SchemaVersion ||
		!validStages(catalog.Container.Stages, declarations.CopyKeys) || !validPresets(catalog.MineGrid.Presets, declarations.CopyKeys) ||
		!validSnake(catalog.Snake) {
		return nil, ErrInvalidCatalog
	}
	return &catalog, nil
}

func validStages(stages []Stage, keys map[string]struct{}) bool {
	if len(stages) == 0 {
		return false
	}
	seen := map[string]bool{}
	for index, stage := range stages {
		if !idPattern.MatchString(stage.StageID) || seen[stage.StageID] || stage.MinTier < 0 || stage.MinTier > StageTierMax ||
			index > 0 && stages[index-1].MinTier >= stage.MinTier || !hasKey(keys, stage.TitleCopyKey) || len(stage.Toys) == 0 {
			return false
		}
		seen[stage.StageID] = true
		for toyIndex, toy := range stage.Toys {
			if !idPattern.MatchString(toy) || toyIndex > 0 && stage.Toys[toyIndex-1] >= toy {
				return false
			}
		}
	}
	return true
}

func validPresets(presets []Preset, keys map[string]struct{}) bool {
	if len(presets) == 0 {
		return false
	}
	for index, preset := range presets {
		if !idPattern.MatchString(preset.PresetID) || index > 0 && presets[index-1].PresetID >= preset.PresetID ||
			preset.Width < 5 || preset.Width > 30 || preset.Height < 5 || preset.Height > 30 ||
			preset.Mines < 1 || preset.Mines > preset.Width*preset.Height-9 ||
			preset.CopyKey != "arcade.mine_grid.preset."+preset.PresetID || !hasKey(keys, preset.CopyKey) {
			return false
		}
	}
	return true
}

func validSnake(snake SnakeContent) bool {
	return snake.Width >= 5 && snake.Width <= 30 && snake.Height >= 5 && snake.Height <= 30 &&
		snake.StartLength >= 2 && snake.StartLength <= snake.Width/2 && snake.GrowthPerFood >= 1 && snake.GrowthPerFood <= 8 &&
		snake.MaxTicksPerAdvance >= 1 && snake.MaxTicksPerAdvance <= 256 &&
		snake.PresentationTickMS >= 50 && snake.PresentationTickMS <= 1000
}

// Preset returns the named mine_grid preset.
func (catalog *Catalog) Preset(id string) (Preset, bool) {
	for _, preset := range catalog.MineGrid.Presets {
		if preset.PresetID == id {
			return preset, true
		}
	}
	return Preset{}, false
}

func hasKey(keys map[string]struct{}, key string) bool {
	if !idPattern.MatchString(key) {
		return false
	}
	_, ok := keys[key]
	return ok
}

func ContentHash(data []byte) string {
	digest := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(digest[:])
}

func strictDecode(data []byte, destination any) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return err
	}
	var trailing any
	if !errors.Is(decoder.Decode(&trailing), io.EOF) {
		return ErrInvalidCatalog
	}
	return nil
}

func hasExactJSONKeys(data []byte, expected ...string) bool {
	var object map[string]json.RawMessage
	if json.Unmarshal(data, &object) != nil || object == nil || len(object) != len(expected) {
		return false
	}
	for _, key := range expected {
		if _, ok := object[key]; !ok {
			return false
		}
	}
	return true
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
				key, ok := keyToken.(string)
				if keyErr != nil || !ok {
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
	_, err := decoder.Token()
	return errors.Is(err, io.EOF)
}

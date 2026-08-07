package minigame

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"sort"

	"cloud-clicker/server/decimal"
)

var ErrInvalidCatalog = errors.New("invalid minigame catalog")

type RatingPolicy struct {
	StartingElo      int64  `json:"starting_elo"`
	EloFloor         int64  `json:"elo_floor"`
	EloCeiling       int64  `json:"elo_ceiling"`
	ProvisionalGames int64  `json:"provisional_games"`
	SeasonMember     string `json:"season_member"`
}

type UnlockCondition struct {
	Kind     string          `json:"kind"`
	FactID   string          `json:"fact_id,omitempty"`
	Value    json.RawMessage `json:"value,omitempty"`
	UnlockID string          `json:"unlock_id,omitempty"`
}

type Definition struct {
	MinigameID         string
	EngineRef          string
	EngineVersion      string
	Modes              []Mode
	ResultScoreFactIDs []string
	Scaling            *ScalingPolicy
	Payout             PayoutPolicy
	Fallback           FallbackPolicy
	OfflineQuality     OfflineQualityPolicy
	Rating             RatingPolicy
	Unlock             UnlockCondition
	SoulGate           string
}

// Catalog is the complete immutable policy surface owned by the pinned
// minigames artifact. There is deliberately no second, process-local policy
// registry: live resolution and replay both resolve Definition rows here.
type Catalog struct {
	schemaVersion int
	definitions   []Definition
	ratingSeasons []string
}

type catalogWire struct {
	SchemaVersion int               `json:"schema_version"`
	RatingSeasons []string          `json:"rating_seasons"`
	Minigames     []json.RawMessage `json:"minigames"`
}

type definitionWire struct {
	MinigameID         string          `json:"minigame_id"`
	EngineRef          string          `json:"engine_ref"`
	EngineVersion      string          `json:"engine_version"`
	Modes              []Mode          `json:"modes"`
	ResultScoreFactIDs []string        `json:"result_score_fact_ids"`
	Scaling            json.RawMessage `json:"scaling"`
	Payout             json.RawMessage `json:"payout"`
	Fallback           json.RawMessage `json:"fallback"`
	OfflineQuality     json.RawMessage `json:"offline_quality"`
	RatingPolicy       json.RawMessage `json:"rating_policy"`
	UnlockCondition    json.RawMessage `json:"unlock_condition"`
	SoulGate           string          `json:"soul_gate"`
}

func LoadCatalog(data []byte) (*Catalog, error) {
	if !uniqueJSONKeys(data) || !hasExactJSONKeys(data, "minigames", "rating_seasons", "schema_version") {
		return nil, ErrInvalidCatalog
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var wire catalogWire
	if decoder.Decode(&wire) != nil || wire.SchemaVersion != 2 && wire.SchemaVersion != 3 {
		return nil, ErrInvalidCatalog
	}
	var trailing any
	if !errors.Is(decoder.Decode(&trailing), io.EOF) || !sortedMechanical(wire.RatingSeasons) || wire.Minigames == nil {
		return nil, ErrInvalidCatalog
	}
	catalog := &Catalog{schemaVersion: wire.SchemaVersion, definitions: make([]Definition, len(wire.Minigames)), ratingSeasons: append([]string(nil), wire.RatingSeasons...)}
	priorID := ""
	for index, raw := range wire.Minigames {
		definition, err := loadDefinition(raw, catalog.ratingSeasons, wire.SchemaVersion)
		if err != nil || index > 0 && priorID >= definition.MinigameID {
			return nil, ErrInvalidCatalog
		}
		priorID = definition.MinigameID
		catalog.definitions[index] = definition
	}
	return catalog, nil
}

func loadDefinition(data []byte, ratingSeasons []string, schemaVersion int) (Definition, error) {
	keys := []string{"engine_ref", "engine_version", "fallback", "minigame_id", "modes",
		"offline_quality", "payout", "rating_policy", "result_score_fact_ids", "scaling", "unlock_condition"}
	if schemaVersion >= 3 {
		keys = append(keys, "soul_gate")
	}
	if !hasExactJSONKeys(data, keys...) {
		return Definition{}, ErrInvalidCatalog
	}
	var wire definitionWire
	if decodeExact(data, &wire) != nil || !mechanicalPattern.MatchString(wire.MinigameID) ||
		!mechanicalPattern.MatchString(wire.EngineRef) || !versionPattern.MatchString(wire.EngineVersion) ||
		!sortedModes(wire.Modes) || !sortedMechanical(wire.ResultScoreFactIDs) ||
		schemaVersion == 2 && wire.SoulGate != "" || schemaVersion == 3 && wire.SoulGate != "human_hobby" && wire.SoulGate != "unrelated" {
		return Definition{}, ErrInvalidCatalog
	}
	scaling, err := LoadScalingPolicy(wire.Scaling, true)
	if err != nil {
		return Definition{}, ErrInvalidCatalog
	}
	declaredFacts := make(map[string]struct{}, len(wire.ResultScoreFactIDs))
	for _, id := range wire.ResultScoreFactIDs {
		declaredFacts[id] = struct{}{}
	}
	// Resource/copy ownership is validated again against the owning economy and
	// copy registries at bundle composition. This loader still closes the row
	// grammar and cross-references every fact within this immutable artifact.
	var payoutProbe PayoutPolicy
	if decodeExact(wire.Payout, &payoutProbe) != nil {
		return Definition{}, ErrInvalidCatalog
	}
	payout, err := LoadPayoutPolicy(wire.Payout, PayoutDeclarations{
		ResourceIDs:  map[string]struct{}{payoutProbe.CreditedResourceID: {}},
		ScoreFactIDs: declaredFacts, CopyKeys: map[string]struct{}{payoutProbe.CapReasonKey: {}},
	})
	if err != nil {
		return Definition{}, ErrInvalidCatalog
	}
	fallback, err := LoadFallbackPolicy(wire.Fallback)
	if err != nil {
		return Definition{}, ErrInvalidCatalog
	}
	var qualityProbe OfflineQualityPolicy
	if decodeExact(wire.OfflineQuality, &qualityProbe) != nil {
		return Definition{}, ErrInvalidCatalog
	}
	quality, err := LoadOfflineQualityPolicy(wire.OfflineQuality, OfflineQualityDeclarations{
		ScoreFactIDs: declaredFacts, AutomationDestinations: map[string]struct{}{qualityProbe.AutomationDestination: {}},
	})
	if err != nil {
		return Definition{}, ErrInvalidCatalog
	}
	rating, err := loadRatingPolicy(wire.RatingPolicy, ratingSeasons)
	if err != nil {
		return Definition{}, err
	}
	unlock, err := loadUnlockCondition(wire.UnlockCondition)
	if err != nil {
		return Definition{}, err
	}
	return Definition{MinigameID: wire.MinigameID, EngineRef: wire.EngineRef, EngineVersion: wire.EngineVersion,
		Modes: append([]Mode(nil), wire.Modes...), ResultScoreFactIDs: append([]string(nil), wire.ResultScoreFactIDs...),
		Scaling: scaling, Payout: payout, Fallback: fallback, OfflineQuality: quality, Rating: rating, Unlock: unlock, SoulGate: wire.SoulGate}, nil
}

func (catalog *Catalog) SchemaSupportsSoul() bool {
	if catalog == nil || catalog.schemaVersion != 3 {
		return false
	}
	for _, definition := range catalog.definitions {
		if definition.SoulGate != "human_hobby" && definition.SoulGate != "unrelated" {
			return false
		}
	}
	return true
}

func loadRatingPolicy(data []byte, seasons []string) (RatingPolicy, error) {
	if !hasExactJSONKeys(data, "elo_ceiling", "elo_floor", "provisional_games", "season_member", "starting_elo") {
		return RatingPolicy{}, ErrInvalidCatalog
	}
	var policy RatingPolicy
	if decodeExact(data, &policy) != nil || policy.EloFloor < -decimal.MaxExactInteger ||
		policy.EloCeiling > decimal.MaxExactInteger || policy.EloFloor > policy.StartingElo ||
		policy.StartingElo > policy.EloCeiling || policy.ProvisionalGames < 0 ||
		policy.ProvisionalGames > decimal.MaxExactInteger || !containsSorted(seasons, policy.SeasonMember) {
		return RatingPolicy{}, ErrInvalidCatalog
	}
	return policy, nil
}

func loadUnlockCondition(data []byte) (UnlockCondition, error) {
	var header struct {
		Kind string `json:"kind"`
	}
	if decodeExactHeader(data, &header) != nil {
		return UnlockCondition{}, ErrInvalidCatalog
	}
	switch header.Kind {
	case "always":
		if !hasExactJSONKeys(data, "kind") {
			return UnlockCondition{}, ErrInvalidCatalog
		}
		return UnlockCondition{Kind: header.Kind}, nil
	case "fact_equals":
		if !hasExactJSONKeys(data, "fact_id", "kind", "value") {
			return UnlockCondition{}, ErrInvalidCatalog
		}
		var wire struct {
			Kind   string          `json:"kind"`
			FactID string          `json:"fact_id"`
			Value  json.RawMessage `json:"value"`
		}
		if decodeExact(data, &wire) != nil || !mechanicalPattern.MatchString(wire.FactID) || !validUnlockValue(wire.Value) {
			return UnlockCondition{}, ErrInvalidCatalog
		}
		return UnlockCondition{Kind: wire.Kind, FactID: wire.FactID, Value: bytes.Clone(wire.Value)}, nil
	case "fiscal_unlock":
		if !hasExactJSONKeys(data, "kind", "unlock_id") {
			return UnlockCondition{}, ErrInvalidCatalog
		}
		var wire struct {
			Kind     string `json:"kind"`
			UnlockID string `json:"unlock_id"`
		}
		if decodeExact(data, &wire) != nil || !mechanicalPattern.MatchString(wire.UnlockID) {
			return UnlockCondition{}, ErrInvalidCatalog
		}
		return UnlockCondition{Kind: wire.Kind, UnlockID: wire.UnlockID}, nil
	default:
		return UnlockCondition{}, ErrInvalidCatalog
	}
}

func decodeExactHeader(data []byte, destination any) error {
	if !uniqueJSONKeys(data) {
		return ErrInvalidCatalog
	}
	return json.Unmarshal(data, destination)
}

func validUnlockValue(data []byte) bool {
	if len(data) == 0 || !json.Valid(data) {
		return false
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	var value any
	if decoder.Decode(&value) != nil {
		return false
	}
	switch typed := value.(type) {
	case bool, string:
		return true
	case json.Number:
		parsed, err := typed.Int64()
		return err == nil && parsed >= -decimal.MaxExactInteger && parsed <= decimal.MaxExactInteger
	default:
		return false
	}
}

func sortedModes(values []Mode) bool {
	if len(values) == 0 {
		return false
	}
	for index, value := range values {
		if value != ModeSolo && value != ModeAsyncSnapshot || index > 0 && values[index-1] >= value {
			return false
		}
	}
	return true
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

func containsSorted(values []string, value string) bool {
	index := sort.SearchStrings(values, value)
	return index < len(values) && values[index] == value
}

func (catalog *Catalog) MinigameIDs() []string {
	if catalog == nil {
		return nil
	}
	result := make([]string, len(catalog.definitions))
	for index := range catalog.definitions {
		result[index] = catalog.definitions[index].MinigameID
	}
	return result
}

func (catalog *Catalog) Definition(id string) (Definition, bool) {
	if catalog == nil {
		return Definition{}, false
	}
	index := sort.Search(len(catalog.definitions), func(index int) bool { return catalog.definitions[index].MinigameID >= id })
	if index == len(catalog.definitions) || catalog.definitions[index].MinigameID != id {
		return Definition{}, false
	}
	value := catalog.definitions[index]
	value.Modes = append([]Mode(nil), value.Modes...)
	value.ResultScoreFactIDs = append([]string(nil), value.ResultScoreFactIDs...)
	return value, true
}

func (catalog *Catalog) RatingSeasons() []string {
	if catalog == nil {
		return nil
	}
	return append([]string(nil), catalog.ratingSeasons...)
}

func (catalog *Catalog) HasRatingSeason(value string) bool {
	return catalog != nil && containsSorted(catalog.ratingSeasons, value)
}

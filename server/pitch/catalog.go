// Package pitch owns the immutable content and pure transition engine for The Pitch.
package pitch

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
	"sort"

	"cloud-clicker/server/decimal"
)

const (
	SchemaVersion = 1
	EngineVersion = "1.0.0"
)

var (
	ErrInvalidCatalog = errors.New("invalid Pitch catalog")
	idPattern         = regexp.MustCompile(`^[a-z][a-z0-9_]*(?:\.[a-z][a-z0-9_]*)*$`)
	instancePattern   = regexp.MustCompile(`^([a-z][a-z0-9_]*(?:\.[a-z][a-z0-9_]*)*)#([1-9][0-9]*)$`)
)

type Policy struct {
	HandSize            int64  `json:"hand_size"`
	PlaySize            int64  `json:"play_size"`
	HandsPerRound       int64  `json:"hands_per_round"`
	HackSlots           int64  `json:"hack_slots"`
	StartCurrency       int64  `json:"start_currency"`
	ShopSize            int64  `json:"shop_size"`
	RoundClearCurrency  int64  `json:"round_clear_currency"`
	HandSizeReasonKey   string `json:"hand_size_reason_key"`
	PlaySizeReasonKey   string `json:"play_size_reason_key"`
	HandsReasonKey      string `json:"hands_per_round_reason_key"`
	HackSlotsReasonKey  string `json:"hack_slots_reason_key"`
	RoundsReasonKey     string `json:"rounds_reason_key"`
	ExponentReasonKey   string `json:"exponent_reason_key"`
	BestExponentHardcap int64  `json:"best_exponent_hardcap"`
}

type MetricCard struct {
	CardID     string `json:"card_id"`
	BaseMetric string `json:"base_metric"`
	Copies     int64  `json:"copies"`
	CopyKey    string `json:"copy_key"`
}

type Effect struct {
	Kind          string `json:"kind"`
	Amount        string `json:"amount,omitempty"`
	Factor        string `json:"factor,omitempty"`
	Shape         string `json:"shape,omitempty"`
	PartnerHackID string `json:"partner_hack_id,omitempty"`
}

type GrowthHack struct {
	HackID      string `json:"hack_id"`
	Price       int64  `json:"price"`
	DraftWeight int64  `json:"draft_weight"`
	Effect      Effect `json:"effect"`
	CopyKey     string `json:"copy_key"`
}

type FundingRow struct {
	Round         int64  `json:"round"`
	FundingTarget string `json:"funding_target"`
}

type Catalog struct {
	SchemaVersion int          `json:"schema_version"`
	Policy        Policy       `json:"policy"`
	MetricCards   []MetricCard `json:"metric_cards"`
	GrowthHacks   []GrowthHack `json:"growth_hacks"`
	FundingCurve  []FundingRow `json:"funding_curve"`
	cardByID      map[string]MetricCard
	hackByID      map[string]GrowthHack
}

type Declarations struct {
	CopyKeys map[string]struct{}
}

func LoadCatalog(data []byte, declarations Declarations) (*Catalog, error) {
	if !uniqueJSONKeys(data) || !hasExactJSONKeys(data, "funding_curve", "growth_hacks", "metric_cards", "policy", "schema_version") ||
		len(declarations.CopyKeys) == 0 {
		return nil, ErrInvalidCatalog
	}
	var exact struct {
		Policy       json.RawMessage   `json:"policy"`
		MetricCards  []json.RawMessage `json:"metric_cards"`
		GrowthHacks  []json.RawMessage `json:"growth_hacks"`
		FundingCurve []json.RawMessage `json:"funding_curve"`
	}
	if json.Unmarshal(data, &exact) != nil || !hasExactJSONKeys(exact.Policy, "best_exponent_hardcap", "exponent_reason_key",
		"hack_slots", "hack_slots_reason_key", "hand_size", "hand_size_reason_key", "hands_per_round",
		"hands_per_round_reason_key", "play_size", "play_size_reason_key", "round_clear_currency", "rounds_reason_key",
		"shop_size", "start_currency") {
		return nil, ErrInvalidCatalog
	}
	for _, row := range exact.MetricCards {
		if !hasExactJSONKeys(row, "base_metric", "card_id", "copies", "copy_key") {
			return nil, ErrInvalidCatalog
		}
	}
	for _, row := range exact.GrowthHacks {
		if !hasExactJSONKeys(row, "copy_key", "draft_weight", "effect", "hack_id", "price") {
			return nil, ErrInvalidCatalog
		}
		var effect struct {
			Effect json.RawMessage `json:"effect"`
		}
		if json.Unmarshal(row, &effect) != nil || !validEffectKeys(effect.Effect) {
			return nil, ErrInvalidCatalog
		}
	}
	for _, row := range exact.FundingCurve {
		if !hasExactJSONKeys(row, "funding_target", "round") {
			return nil, ErrInvalidCatalog
		}
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var catalog Catalog
	if decoder.Decode(&catalog) != nil {
		return nil, ErrInvalidCatalog
	}
	var trailing any
	if !errors.Is(decoder.Decode(&trailing), io.EOF) || catalog.SchemaVersion != SchemaVersion ||
		len(catalog.MetricCards) != 12 || len(catalog.GrowthHacks) != 8 || len(catalog.FundingCurve) != 8 || !validPolicy(catalog.Policy, declarations.CopyKeys) {
		return nil, ErrInvalidCatalog
	}
	catalog.cardByID = make(map[string]MetricCard, len(catalog.MetricCards))
	catalog.hackByID = make(map[string]GrowthHack, len(catalog.GrowthHacks))
	prior := ""
	for _, card := range catalog.MetricCards {
		metric, metricErr := decimal.ParseCanonical(card.BaseMetric)
		if !idPattern.MatchString(card.CardID) || prior >= card.CardID && prior != "" || card.Copies != 2 ||
			metricErr != nil || metric.Lt(decimal.Zero) || card.CopyKey != "pitch.card."+card.CardID {
			return nil, ErrInvalidCatalog
		}
		if _, ok := declarations.CopyKeys[card.CopyKey]; !ok {
			return nil, ErrInvalidCatalog
		}
		prior, catalog.cardByID[card.CardID] = card.CardID, card
	}
	prior = ""
	for _, hack := range catalog.GrowthHacks {
		if !idPattern.MatchString(hack.HackID) || prior >= hack.HackID && prior != "" || hack.Price < 0 ||
			hack.Price > decimal.MaxExactInteger || hack.DraftWeight <= 0 || hack.DraftWeight > decimal.MaxExactInteger ||
			hack.CopyKey != "pitch.hack."+hack.HackID || !validEffect(hack.Effect) {
			return nil, ErrInvalidCatalog
		}
		if _, ok := declarations.CopyKeys[hack.CopyKey]; !ok {
			return nil, ErrInvalidCatalog
		}
		prior, catalog.hackByID[hack.HackID] = hack.HackID, hack
	}
	for _, hack := range catalog.GrowthHacks {
		if hack.Effect.Kind == "chain_factor" {
			if _, ok := catalog.hackByID[hack.Effect.PartnerHackID]; !ok || hack.Effect.PartnerHackID == hack.HackID {
				return nil, ErrInvalidCatalog
			}
		}
	}
	for index, row := range catalog.FundingCurve {
		target, targetErr := decimal.ParseCanonical(row.FundingTarget)
		if row.Round != int64(index+1) || targetErr != nil || !target.Gt(decimal.Zero) ||
			index > 0 && !target.Gt(decimal.FromString(catalog.FundingCurve[index-1].FundingTarget)) {
			return nil, ErrInvalidCatalog
		}
	}
	return &catalog, nil
}

func validEffectKeys(data []byte) bool {
	var header struct {
		Kind string `json:"kind"`
	}
	if json.Unmarshal(data, &header) != nil {
		return false
	}
	switch header.Kind {
	case "flat_add":
		return hasExactJSONKeys(data, "amount", "kind")
	case "card_factor":
		return hasExactJSONKeys(data, "factor", "kind")
	case "shape_factor":
		return hasExactJSONKeys(data, "factor", "kind", "shape")
	case "chain_factor":
		return hasExactJSONKeys(data, "factor", "kind", "partner_hack_id")
	default:
		return false
	}
}

func validPolicy(policy Policy, keys map[string]struct{}) bool {
	if policy.HandSize != 7 || policy.PlaySize != 4 || policy.HandsPerRound != 3 || policy.HackSlots != 4 ||
		policy.StartCurrency != 4 || policy.ShopSize != 3 || policy.RoundClearCurrency != 3 || policy.BestExponentHardcap != 1_000_000 {
		return false
	}
	for _, key := range []string{policy.HandSizeReasonKey, policy.PlaySizeReasonKey, policy.HandsReasonKey,
		policy.HackSlotsReasonKey, policy.RoundsReasonKey, policy.ExponentReasonKey} {
		if !idPattern.MatchString(key) {
			return false
		}
		if _, ok := keys[key]; !ok {
			return false
		}
	}
	return policy.ExponentReasonKey == "cap.pitch_exponent"
}

func validEffect(effect Effect) bool {
	parsePositive := func(value string) bool {
		parsed, err := decimal.ParseCanonical(value)
		return err == nil && parsed.Gt(decimal.Zero)
	}
	switch effect.Kind {
	case "flat_add":
		return effect.Factor == "" && effect.Shape == "" && effect.PartnerHackID == "" && parsePositive(effect.Amount)
	case "card_factor":
		return effect.Amount == "" && effect.Shape == "" && effect.PartnerHackID == "" && parsePositive(effect.Factor)
	case "shape_factor":
		return effect.Amount == "" && effect.PartnerHackID == "" && (effect.Shape == "pair" || effect.Shape == "full_hand") && parsePositive(effect.Factor)
	case "chain_factor":
		return effect.Amount == "" && effect.Shape == "" && idPattern.MatchString(effect.PartnerHackID) && parsePositive(effect.Factor)
	default:
		return false
	}
}

func (catalog *Catalog) Card(id string) (MetricCard, bool) {
	if catalog == nil {
		return MetricCard{}, false
	}
	card, ok := catalog.cardByID[id]
	return card, ok
}

func (catalog *Catalog) Hack(id string) (GrowthHack, bool) {
	if catalog == nil {
		return GrowthHack{}, false
	}
	hack, ok := catalog.hackByID[id]
	return hack, ok
}

func (catalog *Catalog) FundingTarget(round int64) (string, bool) {
	if catalog == nil || round < 1 || round > int64(len(catalog.FundingCurve)) {
		return "", false
	}
	return catalog.FundingCurve[round-1].FundingTarget, true
}

func (catalog *Catalog) CardInstances() []string {
	if catalog == nil {
		return nil
	}
	result := make([]string, 0, len(catalog.MetricCards)*2)
	for _, card := range catalog.MetricCards {
		for copyOrdinal := int64(1); copyOrdinal <= card.Copies; copyOrdinal++ {
			result = append(result, fmt.Sprintf("%s#%d", card.CardID, copyOrdinal))
		}
	}
	return result
}

func BaseCardID(instanceID string) (string, bool) {
	match := instancePattern.FindStringSubmatch(instanceID)
	returnValue := ""
	if len(match) == 3 {
		returnValue = match[1]
	}
	return returnValue, returnValue != ""
}

func ContentHash(data []byte) string {
	digest := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(digest[:])
}

func (catalog *Catalog) HackIDs() []string {
	if catalog == nil {
		return nil
	}
	result := make([]string, len(catalog.GrowthHacks))
	for index, row := range catalog.GrowthHacks {
		result[index] = row.HackID
	}
	return result
}

func sortedUnique(values []string) bool {
	return sort.StringsAreSorted(values) && len(values) == len(uniqueStrings(values))
}

func uniqueStrings(values []string) map[string]struct{} {
	result := make(map[string]struct{}, len(values))
	for _, value := range values {
		result[value] = struct{}{}
	}
	return result
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
	var trailing any
	return errors.Is(decoder.Decode(&trailing), io.EOF)
}

func hasExactJSONKeys(data []byte, expected ...string) bool {
	var value map[string]json.RawMessage
	if json.Unmarshal(data, &value) != nil || len(value) != len(expected) {
		return false
	}
	for _, key := range expected {
		if _, ok := value[key]; !ok {
			return false
		}
	}
	return true
}

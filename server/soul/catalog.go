package soul

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"regexp"
	"sort"

	"cloud-clicker/server/decimal"
)

var ErrInvalidCatalog = errors.New("invalid soul catalog")

type BandMember string

const (
	BandWhole    BandMember = "whole"
	BandDimming  BandMember = "dimming"
	BandHollow   BandMember = "hollow"
	BandNearZero BandMember = "near_zero"
)

type OwnerKind string

const (
	OwnerEvent     OwnerKind = "event"
	OwnerLongevity OwnerKind = "longevity"
	OwnerContract  OwnerKind = "contract"
	OwnerFixture   OwnerKind = "fixture"
)

type EndingVariant string

const (
	EndingEarnestAscension EndingVariant = "earnest_ascension"
	EndingTrainingData     EndingVariant = "training_data"
)

type Policy struct {
	Floor   int64 `json:"soul_floor"`
	Initial int64 `json:"soul_initial"`
	Max     int64 `json:"soul_max"`
}

type Band struct {
	Member             BandMember `json:"band_member"`
	MinInclusive       int64      `json:"min_inclusive"`
	MaxInclusive       int64      `json:"max_inclusive"`
	HumanContentLocked bool       `json:"human_content_locked"`
	ReasonKey          string     `json:"reason_key"`
}

type DebitSource struct {
	SourceID       string    `json:"source_id"`
	OwnerKind      OwnerKind `json:"owner_kind"`
	Amount         int64     `json:"amount"`
	MayExhaust     bool      `json:"may_exhaust"`
	SingleUse      bool      `json:"single_use"`
	CurtainCopyKey string    `json:"curtain_copy_key"`
}

type RecoveryActivity struct {
	ActivityID         string `json:"activity_id"`
	DurationAttendedMS int64  `json:"duration_attended_ms"`
	RecoveryAmount     int64  `json:"recovery_amount"`
	ReasonKey          string `json:"reason_key"`
}

type EndingPolicy struct {
	WholeVariant    EndingVariant `json:"whole_variant"`
	DepletedVariant EndingVariant `json:"depleted_variant"`
}

type Catalog struct {
	SchemaVersion      int                `json:"schema_version"`
	Policy             Policy             `json:"policy"`
	Bands              []Band             `json:"bands"`
	DebitSources       []DebitSource      `json:"debit_sources"`
	RecoveryActivities []RecoveryActivity `json:"recovery_activities"`
	EndingPolicy       EndingPolicy       `json:"ending_policy"`
}

type Declarations struct {
	CopyKeys    map[string]struct{}
	EpochSeeded bool
}

var mechanicalID = regexp.MustCompile(`^[a-z][a-z0-9_]*(?:\.[a-z][a-z0-9_]*)*$`)

func LoadCatalog(data []byte, declarations Declarations) (*Catalog, error) {
	if !uniqueJSONKeys(data) || !exactSoulCatalogKeys(data) || len(declarations.CopyKeys) == 0 {
		return nil, ErrInvalidCatalog
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var catalog Catalog
	if decoder.Decode(&catalog) != nil || catalog.SchemaVersion != 1 {
		return nil, ErrInvalidCatalog
	}
	var trailing any
	if !errors.Is(decoder.Decode(&trailing), io.EOF) || validateCatalog(&catalog, declarations) != nil {
		return nil, ErrInvalidCatalog
	}
	return &catalog, nil
}

func exactSoulCatalogKeys(data []byte) bool {
	if !exactKeys(data, "bands", "debit_sources", "ending_policy", "policy", "recovery_activities", "schema_version") {
		return false
	}
	var root map[string]json.RawMessage
	if json.Unmarshal(data, &root) != nil || !exactKeys(root["policy"], "soul_floor", "soul_initial", "soul_max") ||
		!exactKeys(root["ending_policy"], "depleted_variant", "whole_variant") {
		return false
	}
	return exactRows(root["bands"], "band_member", "human_content_locked", "max_inclusive", "min_inclusive", "reason_key") &&
		exactRows(root["debit_sources"], "amount", "curtain_copy_key", "may_exhaust", "owner_kind", "single_use", "source_id") &&
		exactRows(root["recovery_activities"], "activity_id", "duration_attended_ms", "reason_key", "recovery_amount")
}

func exactRows(data []byte, keys ...string) bool {
	var rows []json.RawMessage
	if json.Unmarshal(data, &rows) != nil || rows == nil {
		return false
	}
	for _, row := range rows {
		if !exactKeys(row, keys...) {
			return false
		}
	}
	return true
}

func validateCatalog(catalog *Catalog, declarations Declarations) error {
	if catalog.Policy.Floor < 0 || catalog.Policy.Max > decimal.MaxExactInteger || catalog.Policy.Floor > catalog.Policy.Initial || catalog.Policy.Initial > catalog.Policy.Max ||
		len(catalog.Bands) != 4 || catalog.DebitSources == nil || catalog.RecoveryActivities == nil ||
		catalog.EndingPolicy.WholeVariant != EndingEarnestAscension || catalog.EndingPolicy.DepletedVariant != EndingTrainingData {
		return ErrInvalidCatalog
	}
	wantMembers := [...]BandMember{BandNearZero, BandHollow, BandDimming, BandWhole}
	next := catalog.Policy.Floor
	for index, band := range catalog.Bands {
		if band.Member != wantMembers[index] || band.MinInclusive != next || band.MinInclusive > band.MaxInclusive ||
			band.MaxInclusive > catalog.Policy.Max || band.HumanContentLocked != (band.Member == BandNearZero) || !copyKey(declarations, band.ReasonKey) {
			return ErrInvalidCatalog
		}
		if band.MaxInclusive == decimal.MaxExactInteger {
			next = band.MaxInclusive
		} else {
			next = band.MaxInclusive + 1
		}
	}
	if catalog.Bands[len(catalog.Bands)-1].MaxInclusive != catalog.Policy.Max {
		return ErrInvalidCatalog
	}
	prior := ""
	for _, source := range catalog.DebitSources {
		if !mechanicalID.MatchString(source.SourceID) || source.SourceID <= prior || !validOwner(source.OwnerKind) || source.Amount < 1 || source.Amount > decimal.MaxExactInteger ||
			source.MayExhaust != source.SingleUse || !copyKey(declarations, source.CurtainCopyKey) || declarations.EpochSeeded && source.OwnerKind == OwnerFixture {
			return ErrInvalidCatalog
		}
		prior = source.SourceID
	}
	prior = ""
	for _, activity := range catalog.RecoveryActivities {
		if !mechanicalID.MatchString(activity.ActivityID) || activity.ActivityID <= prior || activity.DurationAttendedMS < 1 || activity.DurationAttendedMS > decimal.MaxExactInteger ||
			activity.RecoveryAmount < 1 || activity.RecoveryAmount > decimal.MaxExactInteger || !copyKey(declarations, activity.ReasonKey) {
			return ErrInvalidCatalog
		}
		prior = activity.ActivityID
	}
	if !declarations.EpochSeeded && (len(catalog.DebitSources) == 0 || len(catalog.RecoveryActivities) == 0) {
		return ErrInvalidCatalog
	}
	return nil
}

func validOwner(value OwnerKind) bool {
	return value == OwnerEvent || value == OwnerLongevity || value == OwnerContract || value == OwnerFixture
}

func copyKey(declarations Declarations, value string) bool {
	_, ok := declarations.CopyKeys[value]
	return ok
}

func (catalog *Catalog) BandFor(value int64) (Band, bool) {
	if catalog == nil || value < catalog.Policy.Floor || value > catalog.Policy.Max {
		return Band{}, false
	}
	index := sort.Search(len(catalog.Bands), func(index int) bool { return catalog.Bands[index].MaxInclusive >= value })
	if index == len(catalog.Bands) {
		return Band{}, false
	}
	return catalog.Bands[index], true
}

func (catalog *Catalog) HumanContentLocked(value int64) (bool, error) {
	band, ok := catalog.BandFor(value)
	if !ok {
		return false, ErrInvalidCatalog
	}
	return band.HumanContentLocked, nil
}

func (catalog *Catalog) DebitSource(id string) (DebitSource, bool) {
	index := sort.Search(len(catalog.DebitSources), func(index int) bool { return catalog.DebitSources[index].SourceID >= id })
	if index == len(catalog.DebitSources) || catalog.DebitSources[index].SourceID != id {
		return DebitSource{}, false
	}
	return catalog.DebitSources[index], true
}

func (catalog *Catalog) RecoveryActivity(id string) (RecoveryActivity, bool) {
	index := sort.Search(len(catalog.RecoveryActivities), func(index int) bool { return catalog.RecoveryActivities[index].ActivityID >= id })
	if index == len(catalog.RecoveryActivities) || catalog.RecoveryActivities[index].ActivityID != id {
		return RecoveryActivity{}, false
	}
	return catalog.RecoveryActivities[index], true
}

func exactKeys(data []byte, keys ...string) bool {
	var object map[string]json.RawMessage
	if json.Unmarshal(data, &object) != nil || len(object) != len(keys) {
		return false
	}
	for _, key := range keys {
		if _, ok := object[key]; !ok {
			return false
		}
	}
	return true
}

func uniqueJSONKeys(data []byte) bool {
	decoder := json.NewDecoder(bytes.NewReader(data))
	var visit func() bool
	visit = func() bool {
		token, err := decoder.Token()
		if err != nil {
			return false
		}
		delimiter, ok := token.(json.Delim)
		if !ok {
			return true
		}
		switch delimiter {
		case '{':
			seen := map[string]bool{}
			for decoder.More() {
				keyToken, err := decoder.Token()
				key, ok := keyToken.(string)
				if err != nil || !ok || seen[key] {
					return false
				}
				seen[key] = true
				if !visit() {
					return false
				}
			}
			end, err := decoder.Token()
			return err == nil && end == json.Delim('}')
		case '[':
			for decoder.More() {
				if !visit() {
					return false
				}
			}
			end, err := decoder.Token()
			return err == nil && end == json.Delim(']')
		default:
			return false
		}
	}
	return visit()
}

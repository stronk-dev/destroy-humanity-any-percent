package combat

import (
	"bytes"
	"encoding/json"
	"errors"
)

var ErrInvalidLaneRecord = errors.New("invalid combat lane record")

type LaneTargets string

const (
	LaneTargetsUnits     LaneTargets = "units"
	LaneTargetsBuildings LaneTargets = "buildings"
	LaneTargetsAny       LaneTargets = "any"
)

// LaneRecord is C2's common unit/building/spell row representation, not an
// effect union, balance policy, full catalog or executable battle definition.
// Preserve the optional bypass marker's presence; cross-row bypass coverage
// and the cheap-cycle threshold belong to the complete catalog contract.
type LaneRecord struct {
	ID          string      `json:"id"`
	Temperament Temperament `json:"temperament"`
	Cost        int32       `json:"cost"`
	HP          int32       `json:"hp"`
	ATK         int32       `json:"atk"`
	SPD         int32       `json:"spd"`
	Range       int32       `json:"range"`
	Targets     LaneTargets `json:"targets"`
	Tags        []string    `json:"tags"`
	Bypass      *bool       `json:"bypass,omitempty"`
}

func DecodeLaneRecord(data []byte) (LaneRecord, error) {
	keys := []string{"id", "temperament", "cost", "hp", "atk", "spd", "range", "targets", "tags"}
	// Reuse the existing raw-key/Unicode authority for either exact shape.
	fields, err := speciesObject(data, keys...)
	if err != nil {
		fields, err = speciesObject(data, append(keys, "bypass")...)
	}
	if err != nil {
		return LaneRecord{}, ErrInvalidLaneRecord
	}
	var record LaneRecord
	if !speciesString(fields["id"], &record.ID) ||
		!speciesString(fields["temperament"], (*string)(&record.Temperament)) ||
		!speciesString(fields["targets"], (*string)(&record.Targets)) {
		return LaneRecord{}, ErrInvalidLaneRecord
	}
	if _, valid := temperamentIndex(record.Temperament); !valid {
		return LaneRecord{}, ErrInvalidLaneRecord
	}
	if record.Targets != LaneTargetsUnits && record.Targets != LaneTargetsBuildings && record.Targets != LaneTargetsAny {
		return LaneRecord{}, ErrInvalidLaneRecord
	}
	for key, target := range map[string]*int32{"cost": &record.Cost, "hp": &record.HP,
		"atk": &record.ATK, "spd": &record.SPD, "range": &record.Range} {
		raw := bytes.TrimSpace(fields[key])
		if bytes.Equal(raw, []byte("null")) || json.Unmarshal(raw, target) != nil {
			return LaneRecord{}, ErrInvalidLaneRecord
		}
	}
	if record.Tags, err = speciesStrings(fields["tags"]); err != nil {
		return LaneRecord{}, ErrInvalidLaneRecord
	}
	if raw, present := fields["bypass"]; present {
		if json.Unmarshal(raw, &record.Bypass) != nil || record.Bypass == nil {
			return LaneRecord{}, ErrInvalidLaneRecord
		}
	}
	return record, nil
}

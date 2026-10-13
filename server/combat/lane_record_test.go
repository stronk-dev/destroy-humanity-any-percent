package combat

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"testing"
)

func TestLaneRecordVectors(t *testing.T) {
	data, err := os.ReadFile("../../testdata/combat/lane-record-vectors.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Version int `json:"version"`
		Cases   []struct {
			Name     string      `json:"name"`
			Source   string      `json:"source"`
			Expected *LaneRecord `json:"expected"`
		} `json:"cases"`
	}
	if err := json.Unmarshal(data, &fixture); err != nil || fixture.Version != 1 || len(fixture.Cases) != 100 {
		t.Fatalf("invalid complete lane record vectors: %v", err)
	}
	for _, vector := range fixture.Cases {
		t.Run(vector.Name, func(t *testing.T) {
			source := []byte(vector.Source)
			actual, err := DecodeLaneRecord(source)
			if string(source) != vector.Source {
				t.Fatal("decoder rewrote source bytes")
			}
			if vector.Expected == nil {
				if !errors.Is(err, ErrInvalidLaneRecord) || !reflect.DeepEqual(actual, LaneRecord{}) {
					t.Fatalf("malformed row admitted or partial state returned: %+v err=%v", actual, err)
				}
				return
			}
			if err != nil || !reflect.DeepEqual(actual, *vector.Expected) {
				t.Fatalf("lane row=%+v want=%+v err=%v", actual, *vector.Expected, err)
			}
			encoded, err := json.Marshal(actual)
			if err != nil {
				t.Fatal(err)
			}
			decoded, err := DecodeLaneRecord(encoded)
			if err != nil || !reflect.DeepEqual(decoded, actual) {
				t.Fatalf("lane row failed round-trip: %v", err)
			}
		})
	}
}

func TestLaneRecordsUseAllTemperamentsTargetsAndOwnTheirStorage(t *testing.T) {
	marker := true
	row := LaneRecord{ID: "lane", Cost: 2, HP: 100, ATK: 64, SPD: 10, Range: 3,
		Tags: []string{"tag_b", "tag_a", "tag_b"}, Bypass: &marker}
	for _, temperament := range temperamentOrder {
		for _, targets := range []LaneTargets{LaneTargetsUnits, LaneTargetsBuildings, LaneTargetsAny} {
			row.Temperament, row.Targets = temperament, targets
			for _, presence := range []*bool{nil, boolPointer(false), boolPointer(true)} {
				row.Bypass = presence
				source, _ := json.Marshal(row)
				actual, err := DecodeLaneRecord(source)
				if err != nil || !reflect.DeepEqual(actual, row) {
					t.Fatalf("declared axes/marker failed: %s/%s %+v err=%v", temperament, targets, actual, err)
				}
			}
		}
	}
	source, _ := json.Marshal(row)
	before := bytes.Clone(source)
	first, err := DecodeLaneRecord(source)
	if err != nil {
		t.Fatal(err)
	}
	first.Tags[0], first.HP, *first.Bypass = "changed", 0, false
	second, err := DecodeLaneRecord(source)
	if err != nil || !reflect.DeepEqual(second, row) || !bytes.Equal(before, source) {
		t.Fatal("caller mutation aliased source or a subsequent record")
	}
	invalid := bytes.Replace(source, []byte(`"lane"`), []byte{'"', 0xff, '"'}, 1)
	if actual, err := DecodeLaneRecord(invalid); !errors.Is(err, ErrInvalidLaneRecord) || !reflect.DeepEqual(actual, LaneRecord{}) {
		t.Fatal("malformed UTF-8 repaired or partial state returned")
	}
}

func boolPointer(value bool) *bool { return &value }

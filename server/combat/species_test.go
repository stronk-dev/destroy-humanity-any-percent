package combat

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"testing"
)

func TestSpeciesRecordVectors(t *testing.T) {
	data, err := os.ReadFile("../../testdata/combat/species-vectors.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Version      int      `json:"version"`
		Declarations []string `json:"declarations"`
		Cases        []struct {
			Name     string         `json:"name"`
			Source   string         `json:"source"`
			Expected *SpeciesRecord `json:"expected"`
		} `json:"cases"`
	}
	if err := json.Unmarshal(data, &fixture); err != nil || fixture.Version != 1 || len(fixture.Cases) != 75 {
		t.Fatalf("invalid complete species vectors: %v", err)
	}
	declarations := make(map[string]struct{}, len(fixture.Declarations))
	for _, id := range fixture.Declarations {
		declarations[id] = struct{}{}
	}
	for _, vector := range fixture.Cases {
		t.Run(vector.Name, func(t *testing.T) {
			source := []byte(vector.Source)
			actual, err := DecodeSpeciesRecord(source, declarations)
			if string(source) != vector.Source {
				t.Fatal("species decoder rewrote source bytes")
			}
			if vector.Expected == nil {
				if !errors.Is(err, ErrInvalidSpecies) || !reflect.DeepEqual(actual, SpeciesRecord{}) {
					t.Fatalf("malformed species admitted or partial state returned: %+v, err=%v", actual, err)
				}
				return
			}
			if err != nil || !reflect.DeepEqual(actual, *vector.Expected) {
				t.Fatalf("species=%+v want=%+v error=%v", actual, *vector.Expected, err)
			}
			roundtrip, err := json.Marshal(actual)
			if err != nil {
				t.Fatal(err)
			}
			decoded, err := DecodeSpeciesRecord(roundtrip, declarations)
			if err != nil || !reflect.DeepEqual(decoded, actual) {
				t.Fatalf("species row did not round-trip: %v", err)
			}
		})
	}
}

func TestSpeciesRecordsUseAllIdentityPairsAndOwnTheirStorage(t *testing.T) {
	declarations := map[string]struct{}{"alpha": {}, "beta": {}, "gamma": {}, "delta": {}}
	row := SpeciesRecord{ID: "species", Stats: SpeciesStats{100, 64, 20, 10, 8},
		Moves: [4]string{"alpha", "beta", "gamma", "delta"}, Tags: []string{"tag_b", "tag_a"}}
	for _, temperament := range temperamentOrder {
		for _, runtime := range []Runtime{RuntimeContainer, RuntimeVM, RuntimeBareMetal, RuntimeServerless, RuntimeEdge, RuntimeMainframe} {
			row.Temperament, row.Runtime = temperament, runtime
			source, err := json.Marshal(row)
			if err != nil {
				t.Fatal(err)
			}
			decoded, err := DecodeSpeciesRecord(source, declarations)
			if err != nil || !reflect.DeepEqual(decoded, row) {
				t.Fatalf("declared identity species refused: %s/%s %v", temperament, runtime, err)
			}
		}
	}
	source, _ := json.Marshal(row)
	before := bytes.Clone(source)
	first, err := DecodeSpeciesRecord(source, declarations)
	if err != nil {
		t.Fatal(err)
	}
	first.Tags[0], first.Moves[0], first.Stats.HP = "changed", "changed", 1
	second, err := DecodeSpeciesRecord(source, declarations)
	if err != nil || !reflect.DeepEqual(second, row) || !bytes.Equal(before, source) {
		t.Fatal("caller mutation aliased source or a subsequent record")
	}
	delete(declarations, "alpha")
	if _, err := DecodeSpeciesRecord(source, declarations); !errors.Is(err, ErrInvalidSpecies) {
		t.Fatal("removed move declaration still admitted")
	}
	if _, err := DecodeSpeciesRecord(source, nil); !errors.Is(err, ErrInvalidSpecies) {
		t.Fatal("missing catalog declarations admitted")
	}
	// A malformed byte must not turn into an otherwise legal replacement character.
	invalid := bytes.Replace(source, []byte(`"species"`), []byte{'"', 0xff, '"'}, 1)
	if _, err := DecodeSpeciesRecord(invalid, map[string]struct{}{"alpha": {}, "beta": {}, "gamma": {}, "delta": {}}); !errors.Is(err, ErrInvalidSpecies) {
		t.Fatal("malformed UTF-8 silently repaired")
	}
}

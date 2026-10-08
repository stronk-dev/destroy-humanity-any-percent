package combat

import (
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"testing"
)

func TestLoadoutReferenceVectors(t *testing.T) {
	data, err := os.ReadFile("../../testdata/combat/reference-vectors.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Version      int      `json:"version"`
		Declarations []string `json:"declarations"`
		Cases        []struct {
			Name     string   `json:"name"`
			Kind     string   `json:"kind"`
			Source   string   `json:"source"`
			Expected []string `json:"expected"`
		} `json:"cases"`
	}
	if err := json.Unmarshal(data, &fixture); err != nil || fixture.Version != 1 || len(fixture.Cases) != 32 {
		t.Fatalf("invalid shared reference vectors: %v", err)
	}
	declarations := make(map[string]struct{}, len(fixture.Declarations))
	for _, id := range fixture.Declarations {
		declarations[id] = struct{}{}
	}
	for _, vector := range fixture.Cases {
		t.Run(vector.Name, func(t *testing.T) {
			var actual []string
			var err error
			switch vector.Kind {
			case "team":
				team, decodeErr := DecodeTeamReferences([]byte(vector.Source), declarations)
				actual, err = team[:], decodeErr
			case "deck":
				deck, decodeErr := DecodeDeckReferences([]byte(vector.Source), declarations)
				actual, err = deck[:], decodeErr
			default:
				t.Fatal("unknown vector kind")
			}
			if vector.Expected == nil {
				if !errors.Is(err, ErrInvalidReferences) {
					t.Fatalf("malformed references admitted: %q, err=%v", actual, err)
				}
				for _, id := range actual {
					if id != "" {
						t.Fatal("refusal returned partial references")
					}
				}
			} else if err != nil || !reflect.DeepEqual(actual, vector.Expected) {
				t.Fatalf("references=%q, want=%q, err=%v", actual, vector.Expected, err)
			}
		})
	}
}

func TestLoadoutReferencesDoNotAliasDeclarationsOrRepairInput(t *testing.T) {
	declarations := map[string]struct{}{"alpha": {}, "beta": {}, "gamma": {}}
	input := []byte(`{"v":1,"ids":["gamma","alpha","beta"]}`)
	before := string(input)
	team, err := DecodeTeamReferences(input, declarations)
	if err != nil {
		t.Fatal(err)
	}
	team[0] = "changed"
	delete(declarations, "alpha")
	if string(input) != before {
		t.Fatal("decoder changed source bytes")
	}
	if _, err := DecodeTeamReferences(input, declarations); !errors.Is(err, ErrInvalidReferences) {
		t.Fatal("removed declaration still admitted")
	}
	if _, err := DecodeTeamReferences(input, nil); !errors.Is(err, ErrInvalidReferences) {
		t.Fatal("missing catalog declarations admitted")
	}
	invalidUTF8 := append([]byte(`{"v":1,"ids":["alpha","`), 0xff)
	invalidUTF8 = append(invalidUTF8, []byte(`","gamma"]}`)...)
	if _, err := DecodeTeamReferences(invalidUTF8, map[string]struct{}{"alpha": {}, "�": {}, "gamma": {}}); !errors.Is(err, ErrInvalidReferences) {
		t.Fatal("invalid UTF-8 aliased a declared replacement character")
	}
}

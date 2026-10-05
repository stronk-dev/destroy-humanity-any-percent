package garden

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

// Literal replacements preserve JSON token spelling. Parsing and re-marshalling
// these inputs would erase exactly the decimal/exponent defect being tested.
func TestGardenRawCatalogAdmission(t *testing.T) {
	var corpus struct {
		Version int    `json:"version"`
		Valid   string `json:"valid"`
		Cases   []struct {
			Name        string  `json:"name"`
			Reject      bool    `json:"reject"`
			Raw         *string `json:"raw"`
			From        string  `json:"from"`
			To          string  `json:"to"`
			Occurrences int     `json:"occurrences"`
			Prefix      string  `json:"prefix"`
			Suffix      string  `json:"suffix"`
		} `json:"cases"`
	}
	if err := json.Unmarshal(readRepo(t, "testdata/garden/catalog-raw-fixtures-v1.json"), &corpus); err != nil || corpus.Version != 1 || len(corpus.Cases) != 36 {
		t.Fatalf("raw corpus is incomplete or invalid: %v", err)
	}
	valid := string(readRepo(t, corpus.Valid))
	declarations := fixtureDeclarations(t, loadCorpus(t))
	for _, test := range corpus.Cases {
		t.Run(test.Name, func(t *testing.T) {
			input := valid
			if test.Raw != nil {
				input = *test.Raw
			} else if test.From != "" {
				if count := strings.Count(input, test.From); count != test.Occurrences {
					t.Fatalf("raw replacement has %d matches, expected %d", count, test.Occurrences)
				}
				input = strings.Replace(input, test.From, test.To, 1)
			}
			input = test.Prefix + input + test.Suffix
			catalog, err := LoadCatalog([]byte(input), declarations)
			if test.Reject {
				if catalog != nil || !errors.Is(err, ErrInvalidCatalog) {
					t.Fatalf("invalid raw catalog admitted: catalog=%v error=%v", catalog != nil, err)
				}
			} else if err != nil || catalog == nil {
				t.Fatalf("valid raw catalog rejected: %v", err)
			}
		})
	}
}

// SG1 has no nullable fields. SG2 deliberately does: changing their shared
// number scanner must not turn the catalog correction into a save migration.
func TestGardenCatalogCorrectionPreservesNullableState(t *testing.T) {
	catalog := fixtureCatalog(t)
	for _, planted := range []bool{false, true} {
		state := NewState(catalog)
		if planted {
			state.Plots = []Plot{{Row: 0, Col: 0, SpeciesID: catalog.Starters()[0], AgeTicks: 0}}
		}
		encoded, err := EncodeState(state)
		if err != nil {
			t.Fatal(err)
		}
		decoded, err := DecodeState(encoded)
		if err != nil || !state.Equal(decoded) {
			t.Fatalf("SG2 nullable state no longer round-trips: %s: %v", encoded, err)
		}
	}
}

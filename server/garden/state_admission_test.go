package garden

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

// Shared literal states exercise the actual save codec, then its pinned-catalog
// binding. The valid base is canonical Go field order, not a second serializer.
func TestGardenStateAdmission(t *testing.T) {
	var corpus struct {
		Version int    `json:"version"`
		Valid   string `json:"valid"`
		Cases   []struct {
			Name        string `json:"name"`
			Reject      bool   `json:"reject"`
			From        string `json:"from"`
			To          string `json:"to"`
			Occurrences int    `json:"occurrences"`
		} `json:"cases"`
	}
	if err := json.Unmarshal(readRepo(t, "testdata/garden/state-admission-v1.json"), &corpus); err != nil || corpus.Version != 1 || len(corpus.Cases) != 31 {
		t.Fatalf("state admission corpus is incomplete or invalid: %v", err)
	}
	catalog := fixtureCatalog(t)
	for _, test := range corpus.Cases {
		t.Run(test.Name, func(t *testing.T) {
			input := corpus.Valid
			if test.From != "" {
				if count := strings.Count(input, test.From); count != test.Occurrences {
					t.Fatalf("replacement has %d matches, expected %d", count, test.Occurrences)
				}
				input = strings.Replace(input, test.From, test.To, 1)
			}
			state, err := DecodeState([]byte(input))
			if err == nil {
				err = ValidateAgainst(catalog, state)
			}
			if test.Reject {
				if !errors.Is(err, ErrInvalidState) {
					t.Fatalf("invalid state admitted: state=%v error=%v", state != nil, err)
				}
				return
			}
			if err != nil || state == nil {
				t.Fatalf("valid state rejected: %v", err)
			}
			encoded, err := EncodeState(state)
			// -0 is intentionally canonicalized to 0 by both codecs.
			expected := strings.ReplaceAll(input, ":-0", ":0")
			if err != nil || string(encoded) != expected {
				t.Fatalf("canonical bytes changed: got %s want %s error=%v", encoded, expected, err)
			}
		})
	}
}

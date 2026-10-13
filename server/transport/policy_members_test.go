package transport

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestPolicyRejectsNoncanonicalAndDuplicateMembers(t *testing.T) {
	data, err := os.ReadFile("../../balance/transport/phase0.json")
	if err != nil {
		t.Fatal(err)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil || len(fields) != 12 {
		t.Fatalf("invalid current transport policy fixture: %v", err)
	}
	for key, value := range fields {
		t.Run(key, func(t *testing.T) {
			canonical := `"` + key + `"`
			alias := `"` + strings.ToUpper(key) + `"`
			escaped := fmt.Sprintf(`"\u%04x%s"`, key[0], key[1:])
			for name, source := range map[string]string{
				"case replacement":  strings.Replace(string(data), canonical, alias, 1),
				"case collision":    strings.Replace(string(data), "{", "{"+alias+":"+string(value)+",", 1),
				"duplicate":         strings.Replace(string(data), "{", "{"+canonical+":"+string(value)+",", 1),
				"escaped duplicate": strings.Replace(string(data), "{", "{"+escaped+":"+string(value)+",", 1),
			} {
				t.Run(name, func(t *testing.T) {
					if policy, err := LoadPolicy([]byte(source)); !errors.Is(err, ErrInvalidPolicy) || policy != nil {
						t.Fatalf("noncanonical/duplicate member admitted: policy=%+v err=%v", policy, err)
					}
				})
			}
			// Escaping a canonical name is legal JSON, not a case alias.
			before, err := LoadPolicy(data)
			if err != nil {
				t.Fatal(err)
			}
			policy, err := LoadPolicy([]byte(strings.Replace(string(data), canonical, escaped, 1)))
			if err != nil || !reflect.DeepEqual(policy, before) {
				t.Fatalf("legal escaped canonical key changed the policy: %+v err=%v", policy, err)
			}
		})
	}
}

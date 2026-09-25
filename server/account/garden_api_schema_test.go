package account

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

// Server Garden SG9/AC10: every Go-projected garden view validates against the
// registered get_current_garden response, and the schema refuses a salt.
func TestGardenViewFixturesValidate(t *testing.T) {
	registry, err := PrivateAPIRegistry()
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile("../../testdata/garden/view-fixtures-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	var views map[string]json.RawMessage
	if err := json.Unmarshal(data, &views); err != nil || len(views) != 4 {
		t.Fatalf("view fixtures: %v", err)
	}
	for name, view := range views {
		if err := registry.ValidateResponse("get_current_garden", 200, view); err != nil {
			t.Fatalf("%s view fails the schema: %v", name, err)
		}
	}
	leaked := strings.Replace(string(views["active"]), `"kind": "active"`, `"kind": "active", "salt_hex": "0123456789abcdef"`, 1)
	if err := registry.ValidateResponse("get_current_garden", 200, []byte(leaked)); err == nil {
		t.Fatal("the response schema accepts a salt field")
	}
}

package minigameapi

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestCandidateCatalogIsClosedAndSupportsPitch(t *testing.T) {
	data, err := os.ReadFile("../../balance/testdata/minigame-api-candidate-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := LoadCatalog(data)
	if err != nil {
		t.Fatal(err)
	}
	if !catalog.SupportsTenant("pitch", "pitch", "1.0.0") || catalog.SupportsTenant("pitch", "pitch", "2.0.0") {
		t.Fatal("tenant policy did not remain closed")
	}
	for _, invalid := range [][]byte{
		[]byte(`{"schema_version":1,"operations":[],"tenants":[]}`),
		[]byte(`{"schema_version":1,"operations":[{"operation_id":"create_minigame_session","version":1},{"operation_id":"get_current_minigame_session","version":1},{"operation_id":"play_minigame_command","version":1},{"operation_id":"resolve_minigame_session","version":1}],"tenants":[{"engine_ref":"pitch","engine_version":"1.0.0","minigame_id":"pitch","extra":true}]}`),
		[]byte(`{"schema_version":1,"schema_version":1,"operations":[],"tenants":[]}`),
	} {
		if _, err := LoadCatalog(invalid); err == nil {
			t.Fatalf("invalid catalog accepted: %s", invalid)
		}
	}
}

func TestRawCatalogAdmissionMatchesTheClosedWire(t *testing.T) {
	data, err := os.ReadFile("../../balance/testdata/minigame-api-candidate-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	var base bytes.Buffer
	if err := json.Compact(&base, data); err != nil {
		t.Fatal(err)
	}
	data, err = os.ReadFile("../../testdata/minigame/api-catalog-raw-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	var vectors struct {
		Cases []struct {
			Name, Find, Replace string
			Valid               bool
		}
	}
	if err := json.Unmarshal(data, &vectors); err != nil || len(vectors.Cases) == 0 {
		t.Fatalf("invalid raw vectors: %v", err)
	}
	for _, vector := range vectors.Cases {
		t.Run(vector.Name, func(t *testing.T) {
			if strings.Count(base.String(), vector.Find) != 1 {
				t.Fatal("raw mutation must bind exactly one source occurrence")
			}
			raw := strings.Replace(base.String(), vector.Find, vector.Replace, 1)
			catalog, err := LoadCatalog([]byte(raw))
			if (err == nil) != vector.Valid {
				t.Fatalf("admitted=%t, want %t, err=%v", err == nil, vector.Valid, err)
			}
			if vector.Valid && !catalog.SupportsTenant("pitch", "pitch", "1.0.0") {
				t.Fatal("valid raw form changed the tenant binding")
			}
		})
	}
}

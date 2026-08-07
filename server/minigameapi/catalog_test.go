package minigameapi

import (
	"os"
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

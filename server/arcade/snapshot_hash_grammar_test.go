package arcade

import (
	"encoding/json"
	"errors"
	"testing"

	"cloud-clicker/server/minigame"
)

func TestArcadeSnapshotHashGrammar(t *testing.T) {
	var fixture struct {
		Version int `json:"version"`
		Cases   []struct {
			Name  string `json:"name"`
			Valid bool   `json:"valid"`
			Hash  string `json:"hash"`
		} `json:"cases"`
	}
	if err := json.Unmarshal(readFile(t, "../../testdata/arcade/snapshot-hash-grammar-v1.json"), &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.Version != 1 || len(fixture.Cases) != 12 {
		t.Fatal("unexpected shared hash grammar population")
	}
	for _, tenant := range []minigame.Tenant{NewMineGridTenant(), NewSnakeTenant()} {
		t.Run(tenant.Descriptor().EngineRef, func(t *testing.T) {
			s := newSession(t, tenant.Descriptor().EngineRef, 7, fixturePath)
			if err := tenant.ValidateSnapshot(s.snapshot); err != nil {
				t.Fatalf("actual engine genesis must remain valid: %v", err)
			}
			for _, row := range fixture.Cases {
				t.Run(row.Name, func(t *testing.T) {
					var fields map[string]json.RawMessage
					if err := json.Unmarshal(s.snapshot, &fields); err != nil {
						t.Fatal(err)
					}
					hash, err := json.Marshal(row.Hash)
					if err != nil {
						t.Fatal(err)
					}
					fields["arcade_content_hash"] = hash
					candidate, err := json.Marshal(fields)
					if err != nil {
						t.Fatal(err)
					}
					err = tenant.ValidateSnapshot(candidate)
					if row.Valid && err != nil || !row.Valid && !errors.Is(err, minigame.ErrInvalidTenant) {
						t.Fatalf("snapshot hash %q: valid=%v, got %v", row.Hash, row.Valid, err)
					}
				})
			}
		})
	}
}

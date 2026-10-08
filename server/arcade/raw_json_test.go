package arcade

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"cloud-clicker/server/minigame"
)

func TestArcadeRawJSONParity(t *testing.T) {
	var vectors struct {
		Version int
		Catalog []struct {
			Name, Find, Replace string
			Valid               bool
		}
		Commands []struct{ Name, Engine, Raw, Expect string }
	}
	if err := json.Unmarshal(readFile(t, "../../testdata/arcade/raw-json-v1.json"), &vectors); err != nil {
		t.Fatal(err)
	}
	if vectors.Version != 1 || len(vectors.Catalog) != 12 || len(vectors.Commands) != 20 {
		t.Fatal("incomplete raw JSON population")
	}
	var compact bytes.Buffer
	if err := json.Compact(&compact, readFile(t, candidatePath)); err != nil {
		t.Fatal(err)
	}
	for _, row := range vectors.Catalog {
		t.Run("catalog/"+row.Name, func(t *testing.T) {
			if strings.Count(compact.String(), row.Find) != 1 {
				t.Fatal("mutation must target exactly one token")
			}
			content := []byte(strings.Replace(compact.String(), row.Find, row.Replace, 1))
			loaded, err := LoadCatalog(content, declarations())
			if row.Valid {
				if err != nil || loaded == nil {
					t.Fatalf("valid catalog refused: %v", err)
				}
			} else if !errors.Is(err, ErrInvalidCatalog) || loaded != nil {
				t.Fatalf("invalid catalog admitted: %v", err)
			}
			for _, tenant := range []minigame.Tenant{NewMineGridTenant(), NewSnakeTenant()} {
				destination := MineGridScalingDestination
				if tenant.Descriptor().EngineRef == SnakeEngineRef {
					destination = SnakeScalingDestination
				}
				output, err := tenant.Create(minigame.CreateInput{Mode: minigame.ModeSolo, Seed: 7,
					Content: content, ContentHash: ContentHash(content), ContentSchemaVersion: SchemaVersion,
					ScalingInputs: map[string]int64{destination: 1}})
				if row.Valid {
					if err != nil || len(output) == 0 {
						t.Fatalf("valid catalog did not create: %v", err)
					}
				} else if err == nil || len(output) != 0 {
					t.Fatalf("invalid catalog created: %s %v", output, err)
				}
			}
		})
	}
	for _, row := range vectors.Commands {
		t.Run("command/"+row.Name, func(t *testing.T) {
			s := newSession(t, row.Engine, 7, candidatePath)
			var tenant minigame.Tenant = NewSnakeTenant()
			if row.Engine == MineGridEngineRef {
				tenant = NewMineGridTenant()
				if err := s.apply(`{"kind":"choose_board","preset_id":"small"}`); err != nil {
					t.Fatal(err)
				}
			}
			before := bytes.Clone(s.snapshot)
			command := json.RawMessage(row.Raw)
			validation := tenant.ValidateCommand(command)
			output, err := tenant.Apply(minigame.ApplyInput{Mode: minigame.ModeSolo, Seed: s.seed,
				Revision: s.revision, Snapshot: s.snapshot, Command: command, ScalingInputs: s.scaling,
				Content: s.content, ContentHash: s.hash, ContentSchemaVersion: SchemaVersion})
			if row.Expect == "applied" {
				var state struct{ Revision int64 }
				if validation != nil || err != nil || json.Unmarshal(output.Snapshot, &state) != nil || state.Revision != s.revision+1 {
					t.Fatalf("valid raw command refused: %v %v", validation, err)
				}
			} else if rejectionCode(validation) != row.Expect || rejectionCode(err) != row.Expect || len(output.Snapshot) != 0 || output.Result != nil {
				t.Fatalf("expected %s without output, got %v %v %+v", row.Expect, validation, err, output)
			}
			if !bytes.Equal(s.snapshot, before) || string(command) != row.Raw {
				t.Fatal("input mutated")
			}
		})
	}
}

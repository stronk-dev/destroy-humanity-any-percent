package typer

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"

	"cloud-clicker/server/minigame"
)

type rawMutation struct {
	Name, Find, Replace string
	Valid               bool
}

func TestTyperRawJSONParity(t *testing.T) {
	data, err := os.ReadFile("../../testdata/typer/raw-json-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	var vectors struct {
		Catalog, Snapshot []rawMutation
		Commands          []struct{ Name, Phase, Raw, Expect string }
	}
	if err := json.Unmarshal(data, &vectors); err != nil {
		t.Fatal(err)
	}
	if len(vectors.Catalog) != 8 || len(vectors.Snapshot) != 7 || len(vectors.Commands) != 11 {
		t.Fatal("raw JSON population is incomplete")
	}
	replace := func(t *testing.T, source []byte, row rawMutation) json.RawMessage {
		t.Helper()
		if strings.Count(string(source), row.Find) != 1 {
			t.Fatalf("%s must target exactly one token", row.Name)
		}
		return json.RawMessage(strings.Replace(string(source), row.Find, row.Replace, 1))
	}
	for _, row := range vectors.Catalog {
		t.Run("catalog/"+row.Name, func(t *testing.T) {
			content := replace(t, fixtureBytes(t), row)
			catalog, err := LoadCatalog(content, declarations())
			if row.Valid {
				if err != nil || catalog == nil {
					t.Fatalf("valid raw catalog refused: %v", err)
				}
			} else if !errors.Is(err, ErrInvalidCatalog) || catalog != nil {
				t.Fatalf("malformed raw catalog admitted: %v", err)
			}
			output, err := NewTenant().Create(minigame.CreateInput{Mode: minigame.ModeSolo, Seed: 42,
				ScalingInputs: map[string]int64{ScalingDestination: 1}, Content: content,
				ContentHash: ContentHash(content), ContentSchemaVersion: SchemaVersion})
			if row.Valid {
				if err != nil || len(output) == 0 {
					t.Fatalf("valid raw catalog did not create: %s %v", output, err)
				}
			} else if !errors.Is(err, minigame.ErrInvalidTenant) || len(output) != 0 {
				t.Fatalf("malformed catalog created a snapshot: %s %v", output, err)
			}
		})
	}
	for _, row := range vectors.Snapshot {
		t.Run("snapshot/"+row.Name, func(t *testing.T) {
			h := newHarness(t, 42)
			if err := h.apply(`{"kind":"begin","assist_level":"untimed"}`, 1); err != nil {
				t.Fatal(err)
			}
			if err := h.apply(`{"kind":"submit_line","text":"miss"}`, 2); err != nil {
				t.Fatal(err)
			}
			snapshot := replace(t, h.snapshot, row)
			before := bytes.Clone(snapshot)
			validation := NewTenant().ValidateSnapshot(snapshot)
			output, err := NewTenant().Apply(minigame.ApplyInput{Mode: minigame.ModeSolo, Seed: h.seed,
				Revision: h.revision, Snapshot: snapshot, Command: json.RawMessage(`{"kind":"end_run"}`),
				ScalingInputs: map[string]int64{ScalingDestination: 1}, Content: h.content,
				ContentHash: h.hash, ContentSchemaVersion: SchemaVersion, ServerTimeMs: 3})
			if row.Valid {
				if validation != nil || err != nil || output.Result == nil || output.Result.Outcome != OutcomeEndedEarly {
					t.Fatalf("valid raw snapshot refused: %+v %v %v", output, validation, err)
				}
			} else if !errors.Is(validation, minigame.ErrInvalidTenant) || !errors.Is(err, minigame.ErrTenantDivergence) || len(output.Snapshot) != 0 || output.Result != nil {
				t.Fatalf("malformed snapshot must refuse without output: %+v %v %v", output, validation, err)
			}
			if !bytes.Equal(snapshot, before) {
				t.Fatal("snapshot input mutated")
			}
		})
	}
	for _, row := range vectors.Commands {
		t.Run("command/"+row.Name, func(t *testing.T) {
			h := newHarness(t, 42)
			if row.Phase == PhaseTyping {
				if err := h.apply(`{"kind":"begin","assist_level":"untimed"}`, 1); err != nil {
					t.Fatal(err)
				}
			}
			before, revision := bytes.Clone(h.snapshot), h.revision
			err := h.apply(row.Raw, 2)
			if row.Expect == "applied" {
				if err != nil || h.state().Revision != revision+1 {
					t.Fatalf("valid raw command did not advance exactly once: %v", err)
				}
			} else if rejectionCode(err) != row.Expect || h.revision != revision || !bytes.Equal(h.snapshot, before) {
				t.Fatalf("wrong refusal or state changed: wanted %s, got %v", row.Expect, err)
			}
		})
	}
}

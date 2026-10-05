package arcade

import (
	"bytes"
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	"cloud-clicker/server/minigame"
)

func TestMineGridSharedSnapshotValueGrammar(t *testing.T) {
	var fixture struct {
		Version   int    `json:"version"`
		Seed      string `json:"seed"`
		PresetID  string `json:"preset_id"`
		FirstCell int64  `json:"first_cell"`
		Cases     []struct {
			Name  string `json:"name"`
			Stage string `json:"stage"`
			Path  []any  `json:"path"`
			Value any    `json:"value"`
		} `json:"cases"`
	}
	if err := json.Unmarshal(readFile(t, "../../testdata/arcade/snapshot-value-negatives-v1.json"), &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.Version != 1 || fixture.Seed != "7" || len(fixture.Cases) != 19 {
		t.Fatal("unexpected shared grammar population")
	}
	for _, row := range fixture.Cases {
		t.Run(row.Name, func(t *testing.T) {
			s := newSession(t, MineGridEngineRef, 7, fixturePath)
			if err := s.apply(`{"kind":"choose_board","preset_id":"` + fixture.PresetID + `"}`); err != nil {
				t.Fatal(err)
			}
			if row.Stage != "unplaced" {
				command, _ := json.Marshal(map[string]any{"kind": "reveal", "cell": fixture.FirstCell})
				if err := s.apply(string(command)); err != nil {
					t.Fatal(err)
				}
				if value := s.mineGrid(); value.Phase != MineGridPlaying || len(value.Revealed) == 0 {
					t.Fatal("placed positive control did not reach nonterminal revealed state")
				}
			}
			if row.Stage == "terminal" {
				if err := s.apply(`{"kind":"quit"}`); err != nil {
					t.Fatal(err)
				}
			}
			clean := append(json.RawMessage(nil), s.snapshot...)
			var root any
			if err := json.Unmarshal(clean, &root); err != nil {
				t.Fatal(err)
			}
			target := root
			for _, key := range row.Path[:len(row.Path)-1] {
				switch value := target.(type) {
				case map[string]any:
					target = value[key.(string)]
				case []any:
					target = value[int(key.(float64))]
				default:
					t.Fatal("invalid shared mutation path")
				}
			}
			target.(map[string]any)[row.Path[len(row.Path)-1].(string)] = row.Value
			forged, err := json.Marshal(root)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := decodeMineGridSnapshot(forged); !errors.Is(err, minigame.ErrInvalidTenant) {
				t.Errorf("decoder accepted %s: %v", row.Name, err)
			}
			s.snapshot = forged
			input := minigame.ApplyInput{Mode: minigame.ModeSolo, Seed: s.seed, Revision: s.revision,
				Snapshot: forged, Command: json.RawMessage(`{"kind":"quit"}`), ScalingInputs: s.scaling,
				Content: s.content, ContentHash: s.hash, ContentSchemaVersion: SchemaVersion}
			if _, err := NewMineGridTenant().Apply(input); !errors.Is(err, minigame.ErrTenantDivergence) {
				t.Errorf("direct tenant apply accepted %s: %v", row.Name, err)
			}
			if err := s.apply(`{"kind":"quit"}`); !errors.Is(err, minigame.ErrInvalidTenant) && !errors.Is(err, minigame.ErrTenantDivergence) {
				t.Errorf("registry apply accepted %s: %v", row.Name, err)
			}
			s.snapshot = clean
			// The forged call must not advance the platform session revision.
			value, err := decodeMineGridSnapshot(clean)
			if err != nil || value.Revision != s.revision {
				t.Fatal("forged input corrupted the clean control")
			}
			if row.Stage == "terminal" {
				if code := rejectionCode(s.apply(`{"kind":"quit"}`)); code != "illegal_phase" {
					t.Fatalf("clean terminal lost its existing refusal: %s", code)
				}
			} else {
				input.Snapshot = clean
				baseline, err := NewMineGridTenant().Apply(input)
				if err != nil {
					t.Fatalf("clean direct positive control: %v", err)
				}
				if err := s.apply(`{"kind":"quit"}`); err != nil {
					t.Fatalf("clean registry positive control: %v", err)
				}
				if !bytes.Equal(s.snapshot, baseline.Snapshot) || !reflect.DeepEqual(s.result, baseline.Result) {
					t.Fatal("clean registry output diverges from direct positive control")
				}
			}
		})
	}
}

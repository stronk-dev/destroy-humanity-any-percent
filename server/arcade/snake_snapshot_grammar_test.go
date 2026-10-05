package arcade

import (
	"bytes"
	"encoding/json"
	"errors"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"cloud-clicker/server/minigame"
)

func TestSnakeSharedSnapshotGrammar(t *testing.T) {
	type edit struct {
		Before string `json:"before"`
		After  string `json:"after"`
	}
	type mutation struct {
		Name  string `json:"name"`
		Stage string `json:"stage"`
		Edits []edit `json:"edits"`
	}
	var fixture struct {
		Version     int `json:"version"`
		Populations map[string]struct {
			Scenario    string `json:"scenario"`
			ThroughStep int    `json:"through_step"`
		} `json:"populations"`
		Cases         []mutation `json:"cases"`
		PositiveCases []mutation `json:"positive_cases"`
	}
	if err := json.Unmarshal(readFile(t, "../../testdata/arcade/snake-snapshot-negatives-v1.json"), &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.Version != 1 || len(fixture.Cases) != 45 || len(fixture.PositiveCases) != 7 {
		t.Fatal("unexpected predeclared Snake grammar population")
	}
	var corpus contentCorpus
	if err := json.Unmarshal(readFile(t, corpusPath), &corpus); err != nil {
		t.Fatal(err)
	}
	population := func(t *testing.T, stage string) *session {
		ref, ok := fixture.Populations[stage]
		if !ok {
			t.Fatalf("unknown population %s", stage)
		}
		for _, scenario := range corpus.Scenarios {
			if scenario.Name != ref.Scenario {
				continue
			}
			if scenario.Engine != SnakeEngineRef || ref.ThroughStep < 0 || ref.ThroughStep > len(scenario.Steps) {
				t.Fatal("invalid real Snake population reference")
			}
			seed, err := strconv.ParseUint(scenario.Seed, 10, 64)
			if err != nil {
				t.Fatal(err)
			}
			s := newSession(t, SnakeEngineRef, seed, fixturePath)
			if string(s.snapshot) != scenario.GenesisBytes {
				t.Fatal("actual genesis differs from Go corpus")
			}
			for _, step := range scenario.Steps[:ref.ThroughStep] {
				got := "applied"
				if err := s.apply(string(step.Command)); err != nil {
					got = rejectionCode(err)
				}
				if got != step.Expect || string(s.snapshot) != step.SnapshotBytes {
					t.Fatal("actual command population differs from Go corpus")
				}
			}
			s.snake()
			return s
		}
		t.Fatal("missing actual Snake scenario")
		return nil
	}
	mutate := func(t *testing.T, snapshot []byte, edits []edit) []byte {
		value := string(snapshot)
		for _, edit := range edits {
			if edit.Before == "" || strings.Count(value, edit.Before) != 1 {
				t.Fatalf("raw mutation must match exactly once: %q", edit.Before)
			}
			value = strings.Replace(value, edit.Before, edit.After, 1)
		}
		return []byte(value)
	}
	inputFor := func(s *session, snapshot []byte) minigame.ApplyInput {
		return minigame.ApplyInput{Mode: minigame.ModeSolo, Seed: s.seed, Revision: s.revision,
			Snapshot: snapshot, Command: json.RawMessage(`{"kind":"quit"}`), ScalingInputs: s.scaling,
			Content: s.content, ContentHash: s.hash, ContentSchemaVersion: SchemaVersion}
	}
	for _, row := range fixture.Cases {
		t.Run("refuse/"+row.Name, func(t *testing.T) {
			s := population(t, row.Stage)
			forged := mutate(t, s.snapshot, row.Edits)
			if _, err := decodeSnakeSnapshot(forged); !errors.Is(err, minigame.ErrInvalidTenant) {
				t.Errorf("decoder admitted %s: %v", row.Name, err)
			}
			if output, err := NewSnakeTenant().Apply(inputFor(s, forged)); !errors.Is(err, minigame.ErrTenantDivergence) {
				t.Errorf("direct apply did not refuse grammar %s: err=%v snapshot=%s result=%+v", row.Name, err, output.Snapshot, output.Result)
			}
			if row.Stage == "terminal" {
				if code := rejectionCode(s.apply(`{"kind":"quit"}`)); code != "illegal_phase" {
					t.Fatal("clean terminal lost its phase refusal")
				}
			} else if _, err := NewSnakeTenant().Apply(inputFor(s, s.snapshot)); err != nil {
				t.Fatal("clean actual state cannot quit")
			}
		})
	}
	for _, row := range fixture.PositiveCases {
		t.Run("accept/"+row.Name, func(t *testing.T) {
			s := population(t, row.Stage)
			legal := mutate(t, s.snapshot, row.Edits)
			if _, err := decodeSnakeSnapshot(legal); err != nil {
				t.Fatalf("legal spelling/boundary refused: %v", err)
			}
			output, err := NewSnakeTenant().Apply(inputFor(s, legal))
			if err != nil {
				t.Fatal(err)
			}
			if row.Name == "tick_safe_max" {
				value, err := decodeSnakeSnapshot(output.Snapshot)
				if err != nil || value.Tick != maxSafeInteger || output.Result == nil || output.Result.ScoreFacts[1].Value != maxSafeInteger {
					t.Fatal("legal safe tick boundary or certified fact changed")
				}
			} else {
				baseline, err := NewSnakeTenant().Apply(inputFor(s, s.snapshot))
				if err != nil || !bytes.Equal(output.Snapshot, baseline.Snapshot) || !reflect.DeepEqual(output.Result, baseline.Result) {
					t.Fatal("legal spelling changed canonical execution")
				}
			}
		})
	}
}

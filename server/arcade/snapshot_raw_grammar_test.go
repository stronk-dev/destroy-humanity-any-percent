package arcade

import (
	"bytes"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"

	"cloud-clicker/server/minigame"
)

func TestMineGridSharedRawSnapshotGrammar(t *testing.T) {
	type mutation struct {
		Name   string `json:"name"`
		Stage  string `json:"stage"`
		Before string `json:"before"`
		After  string `json:"after"`
	}
	var fixture struct {
		Version       int        `json:"version"`
		Cases         []mutation `json:"cases"`
		PositiveCases []mutation `json:"positive_cases"`
	}
	if err := json.Unmarshal(readFile(t, "../../testdata/arcade/snapshot-raw-negatives-v1.json"), &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.Version != 1 || len(fixture.Cases) != 25 || len(fixture.PositiveCases) != 6 {
		t.Fatal("unexpected shared raw grammar population")
	}
	population := func(t *testing.T, stage string) *session {
		s := newSession(t, MineGridEngineRef, 7, fixturePath)
		if stage != "genesis" {
			if err := s.apply(`{"kind":"choose_board","preset_id":"large"}`); err != nil {
				t.Fatal(err)
			}
		}
		if stage == "placed" || stage == "terminal" {
			if err := s.apply(`{"kind":"reveal","cell":40}`); err != nil {
				t.Fatal(err)
			}
			if value := s.mineGrid(); value.Phase != MineGridPlaying || len(value.Revealed) == 0 {
				t.Fatal("placed positive control did not reach revealed playing state")
			}
		}
		if stage == "terminal" {
			if err := s.apply(`{"kind":"quit"}`); err != nil {
				t.Fatal(err)
			}
		}
		s.mineGrid() // The exact clean source must decode before any mutation.
		return s
	}
	inputFor := func(s *session, snapshot []byte) minigame.ApplyInput {
		return minigame.ApplyInput{Mode: minigame.ModeSolo, Seed: s.seed, Revision: s.revision,
			Snapshot: snapshot, Command: json.RawMessage(`{"kind":"quit"}`), ScalingInputs: s.scaling,
			Content: s.content, ContentHash: s.hash, ContentSchemaVersion: SchemaVersion}
	}
	for _, row := range fixture.Cases {
		t.Run("refuse/"+row.Name, func(t *testing.T) {
			s := population(t, row.Stage)
			if count := strings.Count(string(s.snapshot), row.Before); count != 1 {
				t.Fatalf("mutation must match exactly once, got %d", count)
			}
			forged := []byte(strings.Replace(string(s.snapshot), row.Before, row.After, 1))
			if _, err := decodeMineGridSnapshot(forged); !errors.Is(err, minigame.ErrInvalidTenant) {
				t.Errorf("decoder accepted %s: %v", row.Name, err)
			}
			if output, err := NewMineGridTenant().Apply(inputFor(s, forged)); !errors.Is(err, minigame.ErrTenantDivergence) {
				t.Errorf("direct apply accepted %s: err=%v output=%s", row.Name, err, output.Snapshot)
			}
			if row.Stage == "terminal" {
				if code := rejectionCode(s.apply(`{"kind":"quit"}`)); code != "illegal_phase" {
					t.Fatalf("clean terminal lost its refusal: %s", code)
				}
			} else {
				baseline, err := NewMineGridTenant().Apply(inputFor(s, s.snapshot))
				if err != nil {
					t.Fatal(err)
				}
				output, err := NewMineGridTenant().Apply(inputFor(s, s.snapshot))
				if err != nil || !bytes.Equal(output.Snapshot, baseline.Snapshot) || !reflect.DeepEqual(output.Result, baseline.Result) {
					t.Fatal("clean pure execution changed")
				}
			}
		})
	}
	for _, row := range fixture.PositiveCases {
		t.Run("accept/"+row.Name, func(t *testing.T) {
			s := population(t, row.Stage)
			if count := strings.Count(string(s.snapshot), row.Before); count != 1 {
				t.Fatalf("positive mutation must match exactly once, got %d", count)
			}
			legal := []byte(strings.Replace(string(s.snapshot), row.Before, row.After, 1))
			if _, err := decodeMineGridSnapshot(legal); err != nil {
				t.Fatalf("valid spelling rejected: %v", err)
			}
			output, err := NewMineGridTenant().Apply(inputFor(s, legal))
			if err != nil {
				t.Fatal(err)
			}
			if row.Name == "flag_coordinates_at_bounds" {
				value, err := decodeMineGridSnapshot(output.Snapshot)
				if err != nil || !reflect.DeepEqual(value.Flags, []int64{0, 80}) {
					t.Fatal("valid coordinate boundaries changed")
				}
			} else {
				baseline, err := NewMineGridTenant().Apply(inputFor(s, s.snapshot))
				if err != nil || !bytes.Equal(output.Snapshot, baseline.Snapshot) || !reflect.DeepEqual(output.Result, baseline.Result) {
					t.Fatal("valid spelling changed canonical execution")
				}
			}
		})
	}
}

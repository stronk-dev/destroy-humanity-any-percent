package arcade

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"
)

type snakeAtomicCase struct {
	Name    string `json:"name"`
	Through string `json:"through"`
	Turns   []struct {
		Tick      string `json:"tick"`
		Direction string `json:"direction"`
	} `json:"turns"`
	Expected string `json:"expected"`
}

func TestSnakeRejectionAtomicity(t *testing.T) {
	var fixture struct {
		Version int               `json:"version"`
		Cases   []snakeAtomicCase `json:"cases"`
	}
	if err := json.Unmarshal(readFile(t, "../../testdata/arcade/snake-rejection-atomicity-v1.json"), &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.Version != 1 || len(fixture.Cases) != 6 {
		t.Fatal("unexpected shared rejection population")
	}
	var content Catalog
	if err := json.Unmarshal(readFile(t, oddFixturePath), &content); err != nil {
		t.Fatal(err)
	}
	for _, moved := range []bool{false, true} {
		stage := "genesis"
		if moved {
			stage = "moved"
		}
		for _, row := range fixture.Cases {
			t.Run(stage+"/"+row.Name, func(t *testing.T) {
				s := newSession(t, SnakeEngineRef, 455, oddFixturePath)
				if moved {
					if err := s.apply(`{"kind":"advance","through_tick":1,"turns":[]}`); err != nil {
						t.Fatal(err)
					}
				}
				state := s.snake()
				times := map[string]int64{"current_tick": state.Tick, "next_tick": state.Tick + 1,
					"second_tick": state.Tick + 2, "after_death": state.Tick + state.Width - state.Body[0]%state.Width + 1}
				resolve := func(key string) int64 {
					value, ok := times[key]
					if !ok {
						t.Fatalf("unknown time marker %s", key)
					}
					return value
				}
				turns := []map[string]any{}
				for _, turn := range row.Turns {
					turns = append(turns, map[string]any{"tick": resolve(turn.Tick), "direction": turn.Direction})
				}
				command, err := json.Marshal(map[string]any{"kind": "advance", "through_tick": resolve(row.Through), "turns": turns})
				if err != nil {
					t.Fatal(err)
				}
				before := append([]byte(nil), s.snapshot...)
				revision := s.revision
				if code := rejectionCode(s.apply(string(command))); code != row.Expected {
					t.Fatalf("expected %s, got %s", row.Expected, code)
				}
				if !bytes.Equal(s.snapshot, before) || s.revision != revision || s.result != nil {
					t.Fatal("registry rejection changed snapshot/revision/result")
				}
				// Also observe the actual transition's input object, not merely an
				// immutable encoded string owned by the caller of Apply.
				decoded, err := decodeSnakeCommand(command)
				if err != nil {
					t.Fatal(err)
				}
				stateBytes, err := json.Marshal(state)
				if err != nil {
					t.Fatal(err)
				}
				result, transitionErr := snakeTransition(&state, decoded, content.Snake)
				after, err := json.Marshal(state)
				if err != nil {
					t.Fatal(err)
				}
				if rejectionCode(transitionErr) != row.Expected || result != nil || !bytes.Equal(after, stateBytes) {
					t.Fatal("direct transition rejection changed its input or error/result")
				}
				if err := s.apply(fmt.Sprintf(`{"kind":"advance","through_tick":%d,"turns":[]}`, state.Tick+1)); err != nil {
					t.Fatal("valid advance refused after rejection:", err)
				}
				if s.revision != revision+1 || s.snake().Tick != state.Tick+1 || s.snake().Body[0] != state.Body[0]+1 {
					t.Fatal("successful positive control did not really advance")
				}
				if err := s.apply(`{"kind":"quit"}`); err != nil || s.result == nil || s.result.Outcome != SnakeQuit {
					t.Fatal("quit unusable after rejected/accepted advances:", err)
				}
				terminal := append([]byte(nil), s.snapshot...)
				terminalRevision := s.revision
				if rejectionCode(s.apply(`{"kind":"quit"}`)) != "illegal_phase" || !bytes.Equal(s.snapshot, terminal) || s.revision != terminalRevision || s.result.Outcome != SnakeQuit {
					t.Fatal("post-terminal refusal mutated the terminal")
				}
			})
		}
	}
}

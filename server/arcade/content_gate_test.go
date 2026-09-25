package arcade

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
	"testing"

	"cloud-clicker/server/minigame"
)

var updateArcadeCorpus = flag.Bool("update-arcade-corpus", false, "regenerate the checked-in arcade content corpus")

const corpusPath = "../../testdata/arcade/content-gate-v1.json"

type corpusStep struct {
	Command json.RawMessage `json:"command"`
	Expect  string          `json:"expect"`
}

type corpusScenario struct {
	Name             string           `json:"name"`
	Engine           string           `json:"engine"`
	Seed             string           `json:"seed"`
	Steps            []corpusStep     `json:"steps"`
	ExpectedTerminal json.RawMessage  `json:"expected_terminal"`
	ExpectedResult   *minigame.Result `json:"expected_result"`
}

type contentCorpus struct {
	Version           int              `json:"version"`
	ArcadeContentHash string           `json:"arcade_content_hash"`
	TransitionBudget  int              `json:"transition_budget"`
	Scenarios         []corpusScenario `json:"scenarios"`
}

// builder drives the real tenants through the registry and records every
// step with its observed outcome: the corpus is generated, never hand-written.
type builder struct {
	s        *session
	scenario corpusScenario
}

func newBuilder(t *testing.T, name, engine string, seed uint64) *builder {
	return &builder{s: newSession(t, engine, seed, fixturePath), scenario: corpusScenario{Name: name, Engine: engine, Seed: strconv.FormatUint(seed, 10)}}
}

func (b *builder) step(command string) string {
	b.s.t.Helper()
	expect := "applied"
	if err := b.s.apply(command); err != nil {
		expect = rejectionCode(err)
		if expect == "" {
			b.s.t.Fatalf("%s: non-rejection error %v for %s", b.scenario.Name, err, command)
		}
	}
	b.scenario.Steps = append(b.scenario.Steps, corpusStep{Command: canonical(b.s.t, command), Expect: expect})
	return expect
}

func (b *builder) must(command string) {
	b.s.t.Helper()
	if got := b.step(command); got != "applied" {
		b.s.t.Fatalf("%s: %s rejected with %s", b.scenario.Name, command, got)
	}
}

func (b *builder) expectReject(command, code string) {
	b.s.t.Helper()
	if got := b.step(command); got != code {
		b.s.t.Fatalf("%s: %s expected %s, got %s", b.scenario.Name, command, code, got)
	}
}

func (b *builder) finish(outcome string) corpusScenario {
	b.s.t.Helper()
	if b.s.result == nil || b.s.result.Outcome != outcome {
		b.s.t.Fatalf("%s: expected terminal outcome %s, got %+v", b.scenario.Name, outcome, b.s.result)
	}
	b.scenario.ExpectedTerminal, b.scenario.ExpectedResult = b.s.snapshot, b.s.result
	return b.scenario
}

func cellCommand(kind string, cell int64) string {
	return fmt.Sprintf(`{"kind":%q,"cell":%d}`, kind, cell)
}

// mineGridScenarios covers AR7's mine_grid list on the corpus fixture.
func mineGridScenarios(t *testing.T) []corpusScenario {
	var scenarios []corpusScenario
	// Every preset, with a centre first reveal, then quit.
	for _, preset := range []string{"large", "medium", "small"} {
		b := newBuilder(t, "mine_grid_preset_"+preset, MineGridEngineRef, 101)
		b.must(fmt.Sprintf(`{"kind":"choose_board","preset_id":%q}`, preset))
		state := b.s.mineGrid()
		b.must(cellCommand("reveal", (state.Height/2)*state.Width+state.Width/2))
		if b.s.result == nil {
			b.must(`{"kind":"quit"}`)
			scenarios = append(scenarios, b.finish(MineGridQuit))
		} else {
			scenarios = append(scenarios, b.finish(MineGridCleared))
		}
	}
	// Clear-board win on the 1-mine small board.
	b := newBuilder(t, "mine_grid_clear_small", MineGridEngineRef, 202)
	b.must(`{"kind":"choose_board","preset_id":"small"}`)
	b.must(cellCommand("reveal", 12))
	mines := int64Set(MineCells(5, 5, 1, 12, 202))
	for cell := int64(0); cell < 25 && b.s.result == nil; cell++ {
		if _, open := revealedSet(ptr(b.s.mineGrid()))[cell]; !open && !mines[cell] {
			b.must(cellCommand("reveal", cell))
		}
	}
	scenarios = append(scenarios, b.finish(MineGridCleared))
	// Flood that stops at a flag, flag toggling, and every rejection code.
	b = newBuilder(t, "mine_grid_flags_and_rejections", MineGridEngineRef, 303)
	b.expectReject(cellCommand("reveal", 0), "illegal_phase")
	b.expectReject(`{"kind":"choose_board","preset_id":"huge"}`, "unknown_preset")
	b.must(`{"kind":"choose_board","preset_id":"large"}`)
	b.expectReject(`{"kind":"choose_board","preset_id":"large"}`, "illegal_phase")
	b.expectReject(cellCommand("reveal", 81), "cell_out_of_range")
	b.must(cellCommand("toggle_flag", 0))
	b.must(cellCommand("toggle_flag", 0))
	b.must(cellCommand("toggle_flag", 0))
	b.expectReject(cellCommand("reveal", 0), "cell_flagged")
	b.must(cellCommand("reveal", 40))
	b.expectReject(cellCommand("reveal", 40), "cell_revealed")
	b.expectReject(cellCommand("toggle_flag", 40), "cell_revealed")
	b.expectReject(cellCommand("chord", 1), "chord_unsatisfied")
	if flags := b.s.mineGrid().Flags; len(flags) != 1 || flags[0] != 0 {
		t.Fatalf("the flood must stop at the flag: %v", flags)
	}
	b.must(`{"kind":"quit"}`)
	b.expectReject(`{"kind":"quit"}`, "illegal_phase")
	scenarios = append(scenarios, b.finish(MineGridQuit))
	// Quit from setup.
	b = newBuilder(t, "mine_grid_quit_setup", MineGridEngineRef, 404)
	b.must(`{"kind":"quit"}`)
	scenarios = append(scenarios, b.finish(MineGridQuit))
	// Chord success, then chord detonate on a wrong flag.
	scenarios = append(scenarios, chordScenario(t, "mine_grid_chord_success", true), chordScenario(t, "mine_grid_chord_detonate", false))
	return scenarios
}

func ptr(value MineGridSnapshot) *MineGridSnapshot { return &value }

// chordScenario searches seeds on the large board for a revealed numbered cell
// with a hidden safe neighbor, then flags either its true mines (success) or
// the same number of safe neighbors (detonate).
func chordScenario(t *testing.T, name string, honest bool) corpusScenario {
	for seed := uint64(1); seed < 500; seed++ {
		probe := newSession(t, MineGridEngineRef, seed, fixturePath)
		if probe.apply(`{"kind":"choose_board","preset_id":"large"}`) != nil || probe.apply(cellCommand("reveal", 40)) != nil || probe.result != nil {
			continue
		}
		state := probe.mineGrid()
		mines := int64Set(MineCells(9, 9, 10, 40, seed))
		revealed := revealedSet(&state)
		for _, row := range state.Revealed {
			if row.Adjacent == 0 {
				continue
			}
			var mineNeighbors, safeHidden []int64
			for _, neighbor := range Neighbors(9, 9, row.Cell) {
				if _, open := revealed[neighbor]; open {
					continue
				}
				if mines[neighbor] {
					mineNeighbors = append(mineNeighbors, neighbor)
				} else {
					safeHidden = append(safeHidden, neighbor)
				}
			}
			flags := mineNeighbors
			if !honest {
				if int64(len(safeHidden)) < row.Adjacent {
					continue
				}
				flags = safeHidden[:row.Adjacent]
			} else if len(safeHidden) == 0 {
				continue
			}
			b := newBuilder(t, name, MineGridEngineRef, seed)
			b.must(`{"kind":"choose_board","preset_id":"large"}`)
			b.must(cellCommand("reveal", 40))
			for _, flag := range flags {
				b.must(cellCommand("toggle_flag", flag))
			}
			b.must(cellCommand("chord", row.Cell))
			if honest {
				if b.s.result != nil {
					return b.finish(MineGridCleared)
				}
				b.must(`{"kind":"quit"}`)
				return b.finish(MineGridQuit)
			}
			return b.finish(MineGridDetonated)
		}
	}
	t.Fatalf("%s: no seed produced a chord scenario", name)
	return corpusScenario{}
}

// hamiltonian is a 6x5 cycle: column 0 down, a column serpentine over rows
// 1..4, column 5 up to row 0, and row 0 back west. It contains the genesis
// head's column-3 upward edge, so a snake that turns up at tick 1 can follow
// it forever without colliding and eventually fills the board.
func hamiltonian() []int64 {
	cells := []int64{}
	add := func(x, y int64) { cells = append(cells, y*6+x) }
	for y := int64(0); y <= 4; y++ {
		add(0, y)
	}
	for x := int64(1); x <= 4; x++ {
		if x%2 == 1 {
			for y := int64(4); y >= 1; y-- {
				add(x, y)
			}
		} else {
			for y := int64(1); y <= 4; y++ {
				add(x, y)
			}
		}
	}
	for y := int64(4); y >= 0; y-- {
		add(5, y)
	}
	for x := int64(4); x >= 1; x-- {
		add(x, 0)
	}
	return cells
}

func directionBetween(width, from, to int64) string {
	switch {
	case to == from-width:
		return "up"
	case to == from+width:
		return "down"
	case to == from-1:
		return "left"
	default:
		return "right"
	}
}

// drive plays the snake with a steering policy, batching at most
// max_ticks_per_advance ticks per advance and flushing exactly at a terminal
// tick or when stop holds, as an honest client does (AR6.3).
func drive(b *builder, policy func(SnakeSnapshot) string, stop func(SnakeSnapshot) bool, maxTicks int64) {
	for spent := int64(0); spent < maxTicks && b.s.result == nil; {
		local := cloneSnake(b.s.snake())
		start := local.Tick
		foodSeed, _ := strconv.ParseUint(local.FoodSeed, 10, 64)
		type turn struct {
			Tick      int64  `json:"tick"`
			Direction string `json:"direction"`
		}
		turns := []turn{}
		through := local.Tick
		for local.Tick-start < 64 {
			t := local.Tick + 1
			if next := policy(local); next != local.Direction {
				turns = append(turns, turn{Tick: t, Direction: next})
				local.Direction = next
			}
			outcome := snakeStep(&local, t, local.Width*local.Height, 1, foodSeed)
			through = t
			if outcome != "" || stop(local) {
				break
			}
		}
		command, _ := json.Marshal(map[string]any{"kind": "advance", "through_tick": through, "turns": turns})
		b.must(string(command))
		spent += through - start
		if stop(b.s.snake()) {
			return
		}
	}
}

func followCycle(cycle []int64) func(SnakeSnapshot) string {
	position := map[int64]int{}
	for index, cell := range cycle {
		position[cell] = index
	}
	return func(s SnakeSnapshot) string {
		head := s.Body[0]
		if s.Tick == 0 {
			return "up"
		}
		return directionBetween(s.Width, head, cycle[(position[head]+1)%len(cycle)])
	}
}

func snakeScenarios(t *testing.T) []corpusScenario {
	var scenarios []corpusScenario
	never := func(SnakeSnapshot) bool { return false }
	// Wall crashes: east straight on, then north, south, and west.
	walls := []struct {
		name  string
		turns string
	}{
		{"snake_wall_east", `[]`},
		{"snake_wall_north", `[{"direction":"up","tick":1}]`},
		{"snake_wall_south", `[{"direction":"down","tick":1}]`},
		{"snake_wall_west", `[{"direction":"up","tick":1},{"direction":"left","tick":2}]`},
	}
	for _, wall := range walls {
		b := newBuilder(t, wall.name, SnakeEngineRef, 505)
		var parsed []map[string]any
		_ = json.Unmarshal([]byte(wall.turns), &parsed)
		policy := func(s SnakeSnapshot) string {
			for _, turn := range parsed {
				if int64(turn["tick"].(float64)) == s.Tick+1 {
					return turn["direction"].(string)
				}
			}
			return s.Direction
		}
		drive(b, policy, never, 64)
		scenarios = append(scenarios, b.finish(SnakeCrashed))
	}
	// Eat, grow, and clear the whole board by following the Hamiltonian cycle.
	b := newBuilder(t, "snake_cleared", SnakeEngineRef, 606)
	drive(b, followCycle(hamiltonian()), never, 5000)
	scenarios = append(scenarios, b.finish(SnakeCleared))
	// A legal tail chase and a self-collision: grow along the cycle, then circle
	// a 2x2 block. Length 4 chases its own tail legally; length 5 bites itself.
	for _, target := range []struct {
		name   string
		length int
		expect string
	}{{"snake_tail_chase", 4, SnakeQuit}, {"snake_self_collision", 5, SnakeCrashed}} {
		scenario, ok := loopScenario(t, target.name, target.length, target.expect)
		if !ok {
			t.Fatalf("%s: no seed produced the scenario", target.name)
		}
		scenarios = append(scenarios, scenario)
	}
	// Every rejection, then quit.
	b = newBuilder(t, "snake_rejections", SnakeEngineRef, 707)
	b.expectReject(`{"kind":"advance","through_tick":0,"turns":[]}`, "advance_window")
	b.expectReject(`{"kind":"advance","through_tick":65,"turns":[]}`, "advance_window")
	b.expectReject(`{"kind":"advance","through_tick":2,"turns":[{"direction":"up","tick":3}]}`, "advance_window")
	b.expectReject(`{"kind":"advance","through_tick":2,"turns":[{"direction":"up","tick":2},{"direction":"left","tick":2}]}`, "turns_not_ascending")
	b.expectReject(`{"kind":"advance","through_tick":1,"turns":[{"direction":"right","tick":1}]}`, "invalid_turn")
	b.expectReject(`{"kind":"advance","through_tick":1,"turns":[{"direction":"left","tick":1}]}`, "invalid_turn")
	b.expectReject(`{"kind":"advance","through_tick":5,"turns":[]}`, "advance_past_terminal")
	b.must(`{"kind":"advance","through_tick":1,"turns":[]}`)
	b.must(`{"kind":"quit"}`)
	b.expectReject(`{"kind":"quit"}`, "illegal_phase")
	scenarios = append(scenarios, b.finish(SnakeQuit))
	return scenarios
}

// loopScenario follows the cycle until the body reaches length with no
// pending growth, then circles the 2x2 block the head can turn into, for
// eight ticks; a food-free loop is required so the outcome is the length's.
func loopScenario(t *testing.T, name string, length int, expect string) (corpusScenario, bool) {
	cycle := hamiltonian()
	for seed := uint64(1); seed < 400; seed++ {
		b := newBuilder(t, name, SnakeEngineRef, seed)
		drive(b, followCycle(cycle), func(s SnakeSnapshot) bool { return len(s.Body) == length && s.PendingGrowth == 0 }, 2000)
		state := b.s.snake()
		if b.s.result != nil || len(state.Body) != length || state.PendingGrowth != 0 {
			continue
		}
		// Clockwise turn sequence starting from the current direction.
		clockwise := map[string]string{"up": "right", "right": "down", "down": "left", "left": "up"}
		counter := map[string]string{"up": "left", "left": "down", "down": "right", "right": "up"}
		for _, rotate := range []map[string]string{clockwise, counter} {
			local := cloneSnake(state)
			foodSeed, _ := strconv.ParseUint(local.FoodSeed, 10, 64)
			var turns []map[string]any
			outcome := ""
			for step := int64(1); step <= 8 && outcome == ""; step++ {
				t := local.Tick + 1
				next := rotate[local.Direction]
				turns = append(turns, map[string]any{"tick": t, "direction": next})
				local.Direction = next
				outcome = snakeStep(&local, t, local.Width*local.Height, 1, foodSeed)
			}
			wantCrash := expect == SnakeCrashed
			if outcome == SnakeCrashed && wantCrash && local.Score == state.Score || outcome == "" && !wantCrash && local.Score == state.Score {
				command, _ := json.Marshal(map[string]any{"kind": "advance", "through_tick": local.Tick, "turns": turns})
				b.must(string(command))
				if !wantCrash {
					b.must(`{"kind":"quit"}`)
				}
				return b.finish(expect), true
			}
		}
	}
	return corpusScenario{}, false
}

func generateCorpus(t *testing.T) contentCorpus {
	content := readFile(t, fixturePath)
	corpus := contentCorpus{Version: 1, ArcadeContentHash: ContentHash(content)}
	corpus.Scenarios = append(mineGridScenarios(t), snakeScenarios(t)...)
	for _, scenario := range corpus.Scenarios {
		for _, step := range scenario.Steps {
			if step.Expect == "applied" {
				corpus.TransitionBudget++
			}
		}
	}
	return corpus
}

func TestArcadeContentGate(t *testing.T) {
	corpus := generateCorpus(t)
	encoded, err := json.MarshalIndent(corpus, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	encoded = append(encoded, '\n')
	if *updateArcadeCorpus {
		if err := os.WriteFile(corpusPath, encoded, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	existing, err := os.ReadFile(corpusPath)
	if err != nil || !bytes.Equal(existing, encoded) {
		t.Fatalf("checked-in arcade corpus is stale; run make arcade-corpus (%v)", err)
	}
	outcomes, codes := map[string]bool{}, map[string]bool{}
	for _, scenario := range corpus.Scenarios {
		outcomes[scenario.Engine+":"+scenario.ExpectedResult.Outcome] = true
		for _, step := range scenario.Steps {
			codes[scenario.Engine+":"+step.Expect] = true
		}
	}
	for _, want := range []string{"mine_grid:cleared", "mine_grid:detonated", "mine_grid:quit", "snake:cleared", "snake:crashed", "snake:quit"} {
		if !outcomes[want] {
			t.Fatalf("corpus misses outcome %s", want)
		}
	}
	for _, code := range mineGridErrors {
		if !codes["mine_grid:"+code] {
			t.Fatalf("corpus misses mine_grid rejection %s", code)
		}
	}
	for _, code := range snakeErrors {
		if !codes["snake:"+code] {
			t.Fatalf("corpus misses snake rejection %s", code)
		}
	}
}

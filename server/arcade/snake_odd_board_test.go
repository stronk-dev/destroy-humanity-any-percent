package arcade

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"testing"
)

const oddFixturePath = "../../testdata/arcade/snake-5x5-fixture-v1.json"
const oddCorpusPath = "../../testdata/arcade/snake-5x5-gate-v1.json"

var updateOddCorpus = flag.Bool("update-arcade-odd-corpus", false, "regenerate the exact 5x5 Snake witness")

// A cycle on columns 1..4, with two two-cell detours through column 0.
// Cell 0 is deliberately excluded, not silently removed from the board.
func oddCycle() []int64 {
	return []int64{1, 6, 5, 10, 11, 16, 15, 20, 21, 22, 17, 12, 7, 8, 13, 18, 23, 24, 19, 14, 9, 4, 3, 2}
}

func adjacent5(from, to int64) bool {
	x, y := from%5-to%5, from/5-to/5
	if x < 0 {
		x = -x
	}
	if y < 0 {
		y = -y
	}
	return x+y == 1
}

func validOddCycle(route []int64) bool {
	if len(route) != 24 {
		return false
	}
	seen := map[int64]bool{}
	for i, cell := range route {
		if cell < 1 || cell >= 25 || seen[cell] || !adjacent5(cell, route[(i+1)%len(route)]) {
			return false
		}
		seen[cell] = true
	}
	return true
}

func oddDirection(state SnakeSnapshot) (string, bool) {
	if state.FoodCell == 0 {
		if len(state.Body) == 24 && state.PendingGrowth > 0 && adjacent5(state.Body[0], 0) {
			return directionBetween(5, state.Body[0], 0), true
		}
		return "", false // Observed failure of this strategy, not an impossible board.
	}
	route := oddCycle()
	for i, cell := range route {
		if cell == state.Body[0] {
			return directionBetween(5, cell, route[(i+1)%len(route)]), true
		}
	}
	return "", false
}

type oddObservation struct {
	Seed       uint64 `json:"seed"`
	Outcome    string `json:"outcome"`
	Tick       int64  `json:"tick"`
	BodyLength int    `json:"body_length"`
	Score      int64  `json:"score"`
}

type oddTrace struct {
	directions []string
	states     []SnakeSnapshot
}

type oddCorpus struct {
	Version             int              `json:"version"`
	ArcadeContentHash   string           `json:"arcade_content_hash"`
	SeedPopulation      int              `json:"seed_population"`
	Outcomes            map[string]int   `json:"outcomes"`
	LargestObservedTick int64            `json:"largest_observed_tick"`
	Observations        []oddObservation `json:"observations"`
	TransitionBudget    int              `json:"transition_budget"`
	Scenario            corpusScenario   `json:"scenario"`
}

func probeOddSeed(t *testing.T, seed uint64) (oddObservation, oddTrace) {
	s := newSession(t, SnakeEngineRef, seed, oddFixturePath)
	trace := oddTrace{}
	for steps := 0; steps < 578; steps++ {
		state := s.snake()
		direction, usable := oddDirection(state)
		if !usable {
			return oddObservation{seed, "excluded_food", state.Tick, len(state.Body), state.Score}, trace
		}
		turns := "[]"
		if direction != state.Direction {
			turns = fmt.Sprintf(`[{"tick":%d,"direction":%q}]`, state.Tick+1, direction)
		}
		if err := s.apply(fmt.Sprintf(`{"kind":"advance","through_tick":%d,"turns":%s}`, state.Tick+1, turns)); err != nil {
			t.Fatalf("seed %d: real one-tick command rejected: %v", seed, err)
		}
		actual := s.snake()
		trace.directions = append(trace.directions, direction)
		trace.states = append(trace.states, actual)
		if s.result != nil {
			return oddObservation{seed, s.result.Outcome, actual.Tick, len(actual.Body), actual.Score}, trace
		}
	}
	t.Fatalf("seed %d exhausted the derived 578-tick research bound; observation invalid", seed)
	return oddObservation{}, oddTrace{}
}

func recordOddTrace(t *testing.T, seed uint64, trace oddTrace) corpusScenario {
	s := newSession(t, SnakeEngineRef, seed, oddFixturePath)
	b := &builder{s: s, scenario: corpusScenario{Name: "snake_5x5_cleared", Engine: SnakeEngineRef,
		Seed: strconv.FormatUint(seed, 10), ExpectedGenesis: append(json.RawMessage(nil), s.snapshot...), GenesisBytes: string(s.snapshot)}}
	for start := 0; start < len(trace.states); {
		end := start
		state := s.snake()
		var turns []map[string]any
		direction := state.Direction
		for end < len(trace.states) && end-start < 64 {
			if trace.directions[end] != direction {
				direction = trace.directions[end]
				turns = append(turns, map[string]any{"tick": end + 1, "direction": direction})
			}
			end++
			if trace.states[end-1].Score != state.Score || trace.states[end-1].Phase == SnakeTerminal {
				break
			}
		}
		if turns == nil {
			turns = []map[string]any{}
		}
		command := func(through int) string {
			encoded, err := json.Marshal(map[string]any{"kind": "advance", "through_tick": through, "turns": turns})
			if err != nil {
				t.Fatal(err)
			}
			return string(encoded)
		}
		if end == len(trace.states) {
			if end-start >= 64 {
				t.Fatal("terminal batch leaves no legal overshoot window")
			}
			before := string(s.snapshot)
			b.expectReject(command(end+1), "advance_past_terminal")
			if string(s.snapshot) != before || s.result != nil {
				t.Fatal("overlong command changed actual state/result")
			}
		}
		b.must(command(end))
		start = end
	}
	b.expectReject(`{"kind":"quit"}`, "illegal_phase")
	return b.finish(SnakeCleared)
}

func validOddWitness(scenario corpusScenario) bool {
	var genesis, terminal SnakeSnapshot
	if json.Unmarshal(scenario.ExpectedGenesis, &genesis) != nil || json.Unmarshal(scenario.ExpectedTerminal, &terminal) != nil ||
		genesis.Width != 5 || genesis.Height != 5 || terminal.Width != 5 || terminal.Height != 5 ||
		!reflect.DeepEqual(genesis.Body, []int64{12, 11}) || terminal.Phase != SnakeTerminal || terminal.FoodCell != -1 ||
		len(terminal.Body) != 25 || terminal.Score != 24 || scenario.ExpectedResult == nil || scenario.ExpectedResult.Outcome != SnakeCleared {
		return false
	}
	seen := map[int64]bool{}
	for _, cell := range terminal.Body {
		if cell < 0 || cell >= 25 || seen[cell] {
			return false
		}
		seen[cell] = true
	}
	facts := scenario.ExpectedResult.ScoreFacts
	return len(facts) == 2 && facts[0].Kind == "snake.food_eaten" && facts[0].Value == terminal.Score &&
		facts[1].Kind == "snake.ticks_survived" && facts[1].Value == terminal.Tick
}

func generateOddCorpus(t *testing.T) oddCorpus {
	if !validOddCycle(oddCycle()) {
		t.Fatal("declared 24-cell route is invalid")
	}
	var old, fixture Catalog
	if err := json.Unmarshal(readFile(t, fixturePath), &old); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(readFile(t, oddFixturePath), &fixture); err != nil {
		t.Fatal(err)
	}
	old.Snake.Width = 5
	if !reflect.DeepEqual(old, fixture) {
		t.Fatal("5x5 fixture changed more than Snake width")
	}
	corpus := oddCorpus{Version: 1, ArcadeContentHash: ContentHash(readFile(t, oddFixturePath)), SeedPopulation: 1024,
		Outcomes: map[string]int{"cleared": 0, "crashed": 0, "excluded_food": 0}}
	var selected uint64
	var selectedTrace oddTrace
	for seed := uint64(1); seed <= 1024; seed++ {
		observation, trace := probeOddSeed(t, seed)
		if _, known := corpus.Outcomes[observation.Outcome]; !known {
			t.Fatal("unclassified research outcome")
		}
		corpus.Observations = append(corpus.Observations, observation)
		corpus.Outcomes[observation.Outcome]++
		if observation.Tick > corpus.LargestObservedTick {
			corpus.LargestObservedTick = observation.Tick
		}
		if observation.Outcome == SnakeCleared && selected == 0 {
			selected, selectedTrace = seed, trace
		}
	}
	if selected == 0 {
		t.Fatal("bounded strategy found no clearing witness; do not substitute 6x5 or infer impossibility")
	}
	corpus.Scenario = recordOddTrace(t, selected, selectedTrace)
	corpus.TransitionBudget = len(corpus.Scenario.Steps)
	if !validOddWitness(corpus.Scenario) {
		t.Fatal("actual generated trace is not the exact 5x5 clearing witness")
	}
	return corpus
}

func TestSnakeOddBoardContentGate(t *testing.T) {
	corpus := generateOddCorpus(t)
	encoded, err := json.MarshalIndent(corpus, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	encoded = append(encoded, '\n')
	if *updateOddCorpus {
		if err := os.WriteFile(oddCorpusPath, encoded, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if existing, err := os.ReadFile(oddCorpusPath); err != nil || !bytes.Equal(existing, encoded) {
		t.Fatalf("exact 5x5 corpus missing/stale; run make arcade-corpus (%v)", err)
	}
	t.Logf("all %d seeds observed: outcomes=%v max_tick=%d selected_seed=%s attempts=%d", corpus.SeedPopulation,
		corpus.Outcomes, corpus.LargestObservedTick, corpus.Scenario.Seed, corpus.TransitionBudget)
}

func TestSnakeOddBoardWitnessControls(t *testing.T) {
	var stored oddCorpus
	if err := json.Unmarshal(readFile(t, oddCorpusPath), &stored); err != nil {
		t.Fatal(err)
	}
	if !validOddWitness(stored.Scenario) {
		t.Fatal("clean witness refused")
	}
	var historical contentCorpus
	if err := json.Unmarshal(readFile(t, corpusPath), &historical); err != nil {
		t.Fatal(err)
	}
	foundHistorical := false
	for _, row := range historical.Scenarios {
		if row.Name == "snake_cleared" {
			foundHistorical = true
			if validOddWitness(row) {
				t.Fatal("6x5 proof masquerades as 5x5")
			}
		}
	}
	if !foundHistorical {
		t.Fatal("missing 6x5 negative substitute")
	}
	clone := func() corpusScenario {
		data, err := json.Marshal(stored.Scenario)
		if err != nil {
			t.Fatal(err)
		}
		var value corpusScenario
		if err := json.Unmarshal(data, &value); err != nil {
			t.Fatal(err)
		}
		return value
	}
	for _, defect := range []string{"short_body", "false_outcome", "false_food_fact", "false_tick_fact"} {
		bad := clone()
		if defect == "short_body" {
			var terminal SnakeSnapshot
			if err := json.Unmarshal(bad.ExpectedTerminal, &terminal); err != nil {
				t.Fatal(err)
			}
			terminal.Body = terminal.Body[:24]
			bad.ExpectedTerminal, _ = json.Marshal(terminal)
		} else if defect == "false_outcome" {
			bad.ExpectedResult.Outcome = SnakeQuit
		} else if defect == "false_food_fact" {
			bad.ExpectedResult.ScoreFacts[0].Value++
		} else {
			bad.ExpectedResult.ScoreFacts[1].Value++
		}
		if validOddWitness(bad) {
			t.Fatalf("witness validator admits %s", defect)
		}
	}
	for _, route := range [][]int64{append([]int64{1}, oddCycle()...), oddCycle()[:23],
		{1, 5, 6, 10, 11, 16, 15, 20, 21, 22, 17, 12, 7, 8, 13, 18, 23, 24, 19, 14, 9, 4, 3, 2}} {
		if validOddCycle(route) {
			t.Fatal("invalid cycle admitted")
		}
	}
}

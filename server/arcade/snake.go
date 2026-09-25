package arcade

import (
	"encoding/json"
	"strconv"

	"cloud-clicker/server/determinism"
	"cloud-clicker/server/minigame"
)

const (
	SnakeEngineRef          = "snake"
	SnakeScalingDestination = "minigame.arcade.snake"
	snakeRunSubstream       = "snake.run.v1"
	snakeFoodSeedSubstream  = "snake.food_seed.v1"
	snakeFoodSubstream      = "snake.food.v1"

	SnakePlaying  = "playing"
	SnakeTerminal = "terminal"

	SnakeCleared = "cleared"
	SnakeCrashed = "crashed"
	SnakeQuit    = "quit"
)

// snakeErrors is AR4.7's sorted rejection taxonomy.
var snakeErrors = []string{"advance_past_terminal", "advance_window", "illegal_phase", "invalid_turn", "turns_not_ascending"}

var snakeFactKinds = []string{"snake.food_eaten", "snake.ticks_survived"}

var snakeDeltas = map[string][2]int64{"up": {0, -1}, "down": {0, 1}, "left": {-1, 0}, "right": {1, 0}}

var snakeOpposites = map[string]string{"up": "down", "down": "up", "left": "right", "right": "left"}

// SnakeSnapshot is AR4.3: exactly fourteen keys. FoodSeed is published as a
// canonical decimal string so an honest client can place food locally.
type SnakeSnapshot struct {
	ArcadeContentHash   string  `json:"arcade_content_hash"`
	ArcadeSchemaVersion int     `json:"arcade_schema_version"`
	Phase               string  `json:"phase"`
	Tick                int64   `json:"tick"`
	Direction           string  `json:"direction"`
	Body                []int64 `json:"body"`
	PendingGrowth       int64   `json:"pending_growth"`
	FoodCell            int64   `json:"food_cell"`
	FoodIndex           int64   `json:"food_index"`
	FoodSeed            string  `json:"food_seed"`
	Score               int64   `json:"score"`
	Width               int64   `json:"width"`
	Height              int64   `json:"height"`
	Revision            int64   `json:"revision"`
}

type SnakeTenant struct{}

func NewSnakeTenant() SnakeTenant { return SnakeTenant{} }

func (SnakeTenant) Descriptor() minigame.Descriptor {
	return minigame.Descriptor{EngineRef: SnakeEngineRef, EngineVersion: EngineVersion,
		CommandSchema: "snake.command.v1", SnapshotSchema: "snake.snapshot.v1", ResultSchema: "minigame.result.v1",
		Modes: []minigame.Mode{minigame.ModeSolo}, ErrorTaxonomy: append([]string(nil), snakeErrors...),
		Destinations: map[string]minigame.DestinationClass{SnakeScalingDestination: minigame.DestinationBreadth}}
}

func (SnakeTenant) ValidateCommand(data json.RawMessage) error {
	_, rejection := decodeSnakeCommand(data)
	return rejection
}

func (SnakeTenant) ValidateSnapshot(data json.RawMessage) error {
	_, err := decodeSnakeSnapshot(data)
	return err
}

func (SnakeTenant) ValidateResult(result *minigame.Result) error {
	if result == nil {
		return nil
	}
	if result.Outcome != SnakeCleared && result.Outcome != SnakeCrashed && result.Outcome != SnakeQuit ||
		result.RatingDelta != nil || len(result.ScoreFacts) != len(snakeFactKinds) {
		return minigame.ErrInvalidTenant
	}
	for index, fact := range result.ScoreFacts {
		if fact.Kind != snakeFactKinds[index] || fact.Value < 0 {
			return minigame.ErrInvalidTenant
		}
	}
	return nil
}

func (SnakeTenant) Create(input minigame.CreateInput) (json.RawMessage, error) {
	catalog, err := catalogForInput(input.Content, input.ContentHash, input.ContentSchemaVersion)
	if err != nil || input.Mode != minigame.ModeSolo || !validLiteralScaling(input.ScalingInputs, SnakeScalingDestination) {
		return nil, minigame.ErrInvalidTenant
	}
	snapshot := SnakeGenesis(catalog.Snake, input.Seed)
	snapshot.ArcadeContentHash, snapshot.ArcadeSchemaVersion = input.ContentHash, input.ContentSchemaVersion
	return encodeCanonical(snapshot)
}

// SnakeGenesis is AR4.3: head at (width/2, height/2), the body extending
// west to start_length, direction right, and food 0 placed.
func SnakeGenesis(content SnakeContent, seed uint64) SnakeSnapshot {
	head := (content.Height/2)*content.Width + content.Width/2
	body := make([]int64, 0, content.StartLength)
	for offset := int64(0); offset < content.StartLength; offset++ {
		body = append(body, head-offset)
	}
	foodSeed := SnakeFoodSeed(seed)
	snapshot := SnakeSnapshot{Phase: SnakePlaying, Direction: "right", Body: body, FoodSeed: strconv.FormatUint(foodSeed, 10),
		Width: content.Width, Height: content.Height, Revision: 1}
	snapshot.FoodCell = placeFood(snapshot.Body, content.Width*content.Height, foodSeed, 0)
	snapshot.FoodIndex = 1
	return snapshot
}

// SnakeFoodSeed is AR4.2's published food seed for a session seed.
func SnakeFoodSeed(seed uint64) uint64 {
	runSeed := determinism.Substream(seed, snakeRunSubstream).Next()
	return determinism.Substream(runSeed, snakeFoodSeedSubstream).Next()
}

// placeFood puts food number k on the ascending list of empty cells, or
// returns -1 when no empty cell remains.
func placeFood(body []int64, cells int64, foodSeed uint64, k int64) int64 {
	occupied := int64Set(body)
	empty := make([]int64, 0, cells)
	for cell := int64(0); cell < cells; cell++ {
		if !occupied[cell] {
			empty = append(empty, cell)
		}
	}
	if len(empty) == 0 {
		return -1
	}
	return empty[determinism.Substream(foodSeed^uint64(k), snakeFoodSubstream).Bound(uint64(len(empty)))]
}

func (SnakeTenant) Apply(input minigame.ApplyInput) (minigame.ApplyOutput, error) {
	catalog, err := catalogForInput(input.Content, input.ContentHash, input.ContentSchemaVersion)
	if err != nil || input.Mode != minigame.ModeSolo || !validLiteralScaling(input.ScalingInputs, SnakeScalingDestination) {
		return minigame.ApplyOutput{}, minigame.ErrInvalidTenant
	}
	snapshot, err := decodeSnakeSnapshot(input.Snapshot)
	if err != nil || snapshot.Revision != input.Revision || snapshot.ArcadeContentHash != input.ContentHash ||
		snapshot.ArcadeSchemaVersion != input.ContentSchemaVersion || snapshot.Width != catalog.Snake.Width ||
		snapshot.Height != catalog.Snake.Height || snapshot.FoodSeed != strconv.FormatUint(SnakeFoodSeed(input.Seed), 10) {
		return minigame.ApplyOutput{}, minigame.ErrTenantDivergence
	}
	command, rejection := decodeSnakeCommand(input.Command)
	if rejection != nil {
		return minigame.ApplyOutput{}, rejection
	}
	result, rejection := snakeTransition(&snapshot, command, catalog.Snake)
	if rejection != nil {
		return minigame.ApplyOutput{}, rejection
	}
	snapshot.Revision = input.Revision + 1
	encoded, err := encodeCanonical(snapshot)
	if err != nil {
		return minigame.ApplyOutput{}, minigame.ErrTenantDivergence
	}
	return minigame.ApplyOutput{Snapshot: encoded, Result: result}, nil
}

type snakeTurn struct {
	Tick      int64
	Direction string
}

type snakeCommand struct {
	Kind        string
	ThroughTick int64
	Turns       []snakeTurn
}

// snakeTransition simulates on a copy so a rejected command mutates nothing.
func snakeTransition(snapshot *SnakeSnapshot, command snakeCommand, content SnakeContent) (*minigame.Result, error) {
	if snapshot.Phase != SnakePlaying {
		return nil, reject("illegal_phase", command.Kind+" requires playing phase")
	}
	if command.Kind == "quit" {
		return terminateSnake(snapshot, SnakeQuit), nil
	}
	if command.ThroughTick <= snapshot.Tick || command.ThroughTick > snapshot.Tick+content.MaxTicksPerAdvance {
		return nil, reject("advance_window", "through_tick is outside the advance window")
	}
	for index, turn := range command.Turns {
		if turn.Tick <= snapshot.Tick || turn.Tick > command.ThroughTick {
			return nil, reject("advance_window", "turn tick is outside the advance window")
		}
		if index > 0 && command.Turns[index-1].Tick >= turn.Tick {
			return nil, reject("turns_not_ascending", "turn ticks must be strictly ascending")
		}
	}
	work := cloneSnake(*snapshot)
	foodSeed, _ := strconv.ParseUint(work.FoodSeed, 10, 64)
	cells := work.Width * work.Height
	turnIndex := 0
	for t := work.Tick + 1; t <= command.ThroughTick; t++ {
		if turnIndex < len(command.Turns) && command.Turns[turnIndex].Tick == t {
			next := command.Turns[turnIndex].Direction
			if next == work.Direction || next == snakeOpposites[work.Direction] {
				return nil, reject("invalid_turn", "a turn must be a 90 degree change")
			}
			work.Direction = next
			turnIndex++
		}
		outcome := snakeStep(&work, t, cells, content.GrowthPerFood, foodSeed)
		if outcome != "" {
			if t != command.ThroughTick {
				return nil, reject("advance_past_terminal", "the game ended before through_tick")
			}
			*snapshot = work
			return terminateSnake(snapshot, outcome), nil
		}
	}
	*snapshot = work
	return nil, nil
}

// snakeStep is AR4.4 in its exact order; it returns a terminal outcome or "".
func snakeStep(snapshot *SnakeSnapshot, t, cells, growthPerFood int64, foodSeed uint64) string {
	delta := snakeDeltas[snapshot.Direction]
	head := snapshot.Body[0]
	x, y := head%snapshot.Width+delta[0], head/snapshot.Width+delta[1]
	if x < 0 || y < 0 || x >= snapshot.Width || y >= snapshot.Height {
		snapshot.Tick = t
		return SnakeCrashed
	}
	next := y*snapshot.Width + x
	growing := snapshot.PendingGrowth > 0
	kept := snapshot.Body
	if !growing {
		kept = snapshot.Body[:len(snapshot.Body)-1]
	}
	for _, cell := range kept {
		if cell == next {
			snapshot.Tick = t
			return SnakeCrashed
		}
	}
	snapshot.Body = append([]int64{next}, kept...)
	if growing {
		snapshot.PendingGrowth--
	}
	if next == snapshot.FoodCell {
		snapshot.Score++
		snapshot.PendingGrowth += growthPerFood
		snapshot.FoodCell = placeFood(snapshot.Body, cells, foodSeed, snapshot.FoodIndex)
		snapshot.FoodIndex++
		if snapshot.FoodCell == -1 {
			snapshot.Tick = t
			return SnakeCleared
		}
	}
	snapshot.Tick = t
	return ""
}

func terminateSnake(snapshot *SnakeSnapshot, outcome string) *minigame.Result {
	snapshot.Phase = SnakeTerminal
	return &minigame.Result{Outcome: outcome, RatingDelta: nil, ScoreFacts: []minigame.ScoreFact{
		{Kind: "snake.food_eaten", Value: snapshot.Score},
		{Kind: "snake.ticks_survived", Value: snapshot.Tick},
	}}
}

func cloneSnake(value SnakeSnapshot) SnakeSnapshot {
	value.Body = append([]int64(nil), value.Body...)
	return value
}

func decodeSnakeCommand(data []byte) (snakeCommand, error) {
	if !uniqueJSONKeys(data) {
		return snakeCommand{}, reject("illegal_phase", "command keys are not unique")
	}
	var header struct {
		Kind string `json:"kind"`
	}
	if json.Unmarshal(data, &header) != nil {
		return snakeCommand{}, reject("illegal_phase", "command is not an object")
	}
	switch header.Kind {
	case "advance":
		var wire struct {
			Kind        string            `json:"kind"`
			ThroughTick json.RawMessage   `json:"through_tick"`
			Turns       []json.RawMessage `json:"turns"`
		}
		if !hasExactJSONKeys(data, "kind", "through_tick", "turns") || strictDecode(data, &wire) != nil || wire.Turns == nil {
			return snakeCommand{}, reject("advance_window", "advance schema mismatch")
		}
		through, ok := exactInteger(wire.ThroughTick)
		if !ok {
			return snakeCommand{}, reject("advance_window", "through_tick must be a safe integer")
		}
		command := snakeCommand{Kind: header.Kind, ThroughTick: through, Turns: make([]snakeTurn, 0, len(wire.Turns))}
		for _, raw := range wire.Turns {
			var turn struct {
				Tick      json.RawMessage `json:"tick"`
				Direction string          `json:"direction"`
			}
			if !hasExactJSONKeys(raw, "tick", "direction") || strictDecode(raw, &turn) != nil {
				return snakeCommand{}, reject("advance_window", "turn schema mismatch")
			}
			tick, tickOK := exactInteger(turn.Tick)
			if !tickOK {
				return snakeCommand{}, reject("advance_window", "turn tick must be a safe integer")
			}
			if _, known := snakeDeltas[turn.Direction]; !known {
				return snakeCommand{}, reject("invalid_turn", "unknown direction")
			}
			command.Turns = append(command.Turns, snakeTurn{Tick: tick, Direction: turn.Direction})
		}
		return command, nil
	case "quit":
		if !hasExactJSONKeys(data, "kind") {
			return snakeCommand{}, reject("illegal_phase", "quit schema mismatch")
		}
		return snakeCommand{Kind: header.Kind}, nil
	default:
		return snakeCommand{}, reject("illegal_phase", "unknown command kind")
	}
}

func decodeSnakeSnapshot(data []byte) (SnakeSnapshot, error) {
	if !uniqueJSONKeys(data) || !hasExactJSONKeys(data, "arcade_content_hash", "arcade_schema_version", "phase", "tick", "direction", "body",
		"pending_growth", "food_cell", "food_index", "food_seed", "score", "width", "height", "revision") {
		return SnakeSnapshot{}, minigame.ErrInvalidTenant
	}
	var value SnakeSnapshot
	if strictDecode(data, &value) != nil || value.ArcadeSchemaVersion != SchemaVersion || !validHash(value.ArcadeContentHash) ||
		value.Phase != SnakePlaying && value.Phase != SnakeTerminal || value.Revision < 1 || value.Tick < 0 ||
		value.Width < 5 || value.Width > 30 || value.Height < 5 || value.Height > 30 || len(value.Body) < 2 ||
		value.PendingGrowth < 0 || value.Score < 0 || value.FoodIndex < 1 || value.FoodIndex != value.Score+1 {
		return SnakeSnapshot{}, minigame.ErrInvalidTenant
	}
	if _, ok := snakeDeltas[value.Direction]; !ok {
		return SnakeSnapshot{}, minigame.ErrInvalidTenant
	}
	parsed, err := strconv.ParseUint(value.FoodSeed, 10, 64)
	if err != nil || strconv.FormatUint(parsed, 10) != value.FoodSeed {
		return SnakeSnapshot{}, minigame.ErrInvalidTenant
	}
	cells := value.Width * value.Height
	seen := map[int64]bool{}
	for _, cell := range value.Body {
		if cell < 0 || cell >= cells || seen[cell] {
			return SnakeSnapshot{}, minigame.ErrInvalidTenant
		}
		seen[cell] = true
	}
	if value.FoodCell < -1 || value.FoodCell >= cells || value.FoodCell >= 0 && seen[value.FoodCell] ||
		value.FoodCell == -1 && value.Phase != SnakeTerminal {
		return SnakeSnapshot{}, minigame.ErrInvalidTenant
	}
	return value, nil
}

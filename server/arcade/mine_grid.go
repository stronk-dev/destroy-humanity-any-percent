package arcade

import (
	"encoding/json"
	"sort"
	"strconv"
	"strings"

	"cloud-clicker/server/copykeys"
	"cloud-clicker/server/determinism"
	"cloud-clicker/server/minigame"
)

const (
	MineGridEngineRef          = "mine_grid"
	MineGridScalingDestination = "minigame.arcade.mine_grid"
	mineGridRunSubstream       = "mine_grid.run.v1"
	mineGridMinesSubstream     = "mine_grid.mines.v1"

	MineGridSetup    = "setup"
	MineGridPlaying  = "playing"
	MineGridTerminal = "terminal"

	MineGridCleared   = "cleared"
	MineGridDetonated = "detonated"
	MineGridQuit      = "quit"
)

// mineGridErrors is AR3.7's sorted rejection taxonomy.
var mineGridErrors = []string{"cell_flagged", "cell_out_of_range", "cell_revealed", "chord_unsatisfied", "illegal_phase", "unknown_preset"}

var mineGridFactKinds = []string{"mine_grid.cells_revealed", "mine_grid.cleared"}

// RevealedCell fields are declared in byte order so the nested object is
// canonical.
type RevealedCell struct {
	Adjacent int64 `json:"adjacent"`
	Cell     int64 `json:"cell"`
}

// MineGridSnapshot is AR3.4: exactly thirteen keys. MineCells stays empty in
// every non-terminal snapshot (the hidden-information invariant).
type MineGridSnapshot struct {
	ArcadeContentHash   string         `json:"arcade_content_hash"`
	ArcadeSchemaVersion int            `json:"arcade_schema_version"`
	Phase               string         `json:"phase"`
	PresetID            *string        `json:"preset_id"`
	Width               int64          `json:"width"`
	Height              int64          `json:"height"`
	Mines               int64          `json:"mines"`
	FirstCell           int64          `json:"first_cell"`
	Revealed            []RevealedCell `json:"revealed"`
	Flags               []int64        `json:"flags"`
	ExplodedCell        int64          `json:"exploded_cell"`
	MineCells           []int64        `json:"mine_cells"`
	Revision            int64          `json:"revision"`
}

type MineGridTenant struct{}

func NewMineGridTenant() MineGridTenant { return MineGridTenant{} }

func (MineGridTenant) Descriptor() minigame.Descriptor {
	return minigame.Descriptor{EngineRef: MineGridEngineRef, EngineVersion: EngineVersion,
		CommandSchema: "mine_grid.command.v1", SnapshotSchema: "mine_grid.snapshot.v1", ResultSchema: "minigame.result.v1",
		Modes: []minigame.Mode{minigame.ModeSolo}, ErrorTaxonomy: append([]string(nil), mineGridErrors...),
		Destinations: map[string]minigame.DestinationClass{MineGridScalingDestination: minigame.DestinationBreadth}}
}

func (MineGridTenant) ValidateCommand(data json.RawMessage) error {
	_, rejection := decodeMineGridCommand(data)
	return rejection
}

func (MineGridTenant) ValidateSnapshot(data json.RawMessage) error {
	_, err := decodeMineGridSnapshot(data)
	return err
}

func (MineGridTenant) ValidateResult(result *minigame.Result) error {
	if result == nil {
		return nil
	}
	if result.Outcome != MineGridCleared && result.Outcome != MineGridDetonated && result.Outcome != MineGridQuit ||
		result.RatingDelta != nil || len(result.ScoreFacts) != len(mineGridFactKinds) {
		return minigame.ErrInvalidTenant
	}
	for index, fact := range result.ScoreFacts {
		if fact.Kind != mineGridFactKinds[index] || fact.Value < 0 {
			return minigame.ErrInvalidTenant
		}
	}
	cleared := result.ScoreFacts[1].Value
	if cleared > 1 || (cleared == 1) != (result.Outcome == MineGridCleared) {
		return minigame.ErrInvalidTenant
	}
	return nil
}

func (MineGridTenant) Create(input minigame.CreateInput) (json.RawMessage, error) {
	if _, err := catalogForInput(input.Content, input.ContentHash, input.ContentSchemaVersion); err != nil ||
		input.Mode != minigame.ModeSolo || !validLiteralScaling(input.ScalingInputs, MineGridScalingDestination) {
		return nil, minigame.ErrInvalidTenant
	}
	return encodeCanonical(MineGridSnapshot{ArcadeContentHash: input.ContentHash, ArcadeSchemaVersion: input.ContentSchemaVersion,
		Phase: MineGridSetup, FirstCell: -1, ExplodedCell: -1, Revealed: []RevealedCell{}, Flags: []int64{}, MineCells: []int64{}, Revision: 1})
}

func (MineGridTenant) Apply(input minigame.ApplyInput) (minigame.ApplyOutput, error) {
	catalog, err := catalogForInput(input.Content, input.ContentHash, input.ContentSchemaVersion)
	if err != nil || input.Mode != minigame.ModeSolo || !validLiteralScaling(input.ScalingInputs, MineGridScalingDestination) {
		return minigame.ApplyOutput{}, minigame.ErrInvalidTenant
	}
	snapshot, err := decodeMineGridSnapshot(input.Snapshot)
	if err != nil || snapshot.Revision != input.Revision || snapshot.ArcadeContentHash != input.ContentHash ||
		snapshot.ArcadeSchemaVersion != input.ContentSchemaVersion || validateMineGridAgainstCatalog(snapshot, catalog, input.Seed) != nil {
		return minigame.ApplyOutput{}, minigame.ErrTenantDivergence
	}
	command, rejection := decodeMineGridCommand(input.Command)
	if rejection != nil {
		return minigame.ApplyOutput{}, rejection
	}
	result, rejection := mineGridTransition(&snapshot, command, catalog, input.Seed)
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

type mineGridCommand struct {
	Kind     string
	PresetID string
	Cell     int64
}

func mineGridTransition(snapshot *MineGridSnapshot, command mineGridCommand, catalog *Catalog, seed uint64) (*minigame.Result, error) {
	if command.Kind == "quit" {
		if snapshot.Phase == MineGridTerminal {
			return nil, reject("illegal_phase", "quit requires a non-terminal phase")
		}
		return terminateMineGrid(snapshot, seed, MineGridQuit), nil
	}
	if command.Kind == "choose_board" {
		if snapshot.Phase != MineGridSetup {
			return nil, reject("illegal_phase", "choose_board requires setup phase")
		}
		preset, ok := catalog.Preset(command.PresetID)
		if !ok {
			return nil, reject("unknown_preset", "preset is not in the pinned catalog")
		}
		id := preset.PresetID
		snapshot.PresetID, snapshot.Width, snapshot.Height, snapshot.Mines = &id, preset.Width, preset.Height, preset.Mines
		snapshot.Phase = MineGridPlaying
		return nil, nil
	}
	if snapshot.Phase != MineGridPlaying {
		return nil, reject("illegal_phase", command.Kind+" requires playing phase")
	}
	if command.Cell < 0 || command.Cell >= snapshot.Width*snapshot.Height {
		return nil, reject("cell_out_of_range", "cell is outside the board")
	}
	revealed := revealedSet(snapshot)
	flags := int64Set(snapshot.Flags)
	switch command.Kind {
	case "reveal":
		if flags[command.Cell] {
			return nil, reject("cell_flagged", "flagged cells cannot be revealed")
		}
		if _, ok := revealed[command.Cell]; ok {
			return nil, reject("cell_revealed", "cell is already revealed")
		}
		if snapshot.FirstCell == -1 {
			snapshot.FirstCell = command.Cell
		}
		mines := int64Set(MineCells(snapshot.Width, snapshot.Height, snapshot.Mines, snapshot.FirstCell, seed))
		if mines[command.Cell] {
			snapshot.ExplodedCell = command.Cell
			return terminateMineGrid(snapshot, seed, MineGridDetonated), nil
		}
		floodReveal(snapshot, revealed, flags, mines, command.Cell)
	case "toggle_flag":
		if _, ok := revealed[command.Cell]; ok {
			return nil, reject("cell_revealed", "revealed cells cannot be flagged")
		}
		if flags[command.Cell] {
			delete(flags, command.Cell)
		} else {
			flags[command.Cell] = true
		}
		snapshot.Flags = sortedCells(flags)
		return nil, nil
	case "chord":
		adjacent, ok := revealed[command.Cell]
		if !ok || adjacent == 0 {
			return nil, reject("chord_unsatisfied", "chord requires a revealed numbered cell")
		}
		neighbors := Neighbors(snapshot.Width, snapshot.Height, command.Cell)
		flagged := int64(0)
		for _, neighbor := range neighbors {
			if flags[neighbor] {
				flagged++
			}
		}
		if flagged != adjacent {
			return nil, reject("chord_unsatisfied", "flagged neighbors do not match the count")
		}
		mines := int64Set(MineCells(snapshot.Width, snapshot.Height, snapshot.Mines, snapshot.FirstCell, seed))
		for _, neighbor := range neighbors {
			if _, open := revealed[neighbor]; !open && !flags[neighbor] && mines[neighbor] {
				snapshot.ExplodedCell = neighbor
				return terminateMineGrid(snapshot, seed, MineGridDetonated), nil
			}
		}
		for _, neighbor := range neighbors {
			if _, open := revealed[neighbor]; !open && !flags[neighbor] {
				floodReveal(snapshot, revealed, flags, mines, neighbor)
			}
		}
	}
	if int64(len(snapshot.Revealed)) == snapshot.Width*snapshot.Height-snapshot.Mines {
		return terminateMineGrid(snapshot, seed, MineGridCleared), nil
	}
	return nil, nil
}

// floodReveal is the closure from start: each cell is revealed with its
// adjacent-mine count, zeros expand to hidden unflagged neighbors, and flags
// are never auto-revealed. The stored list is sorted.
func floodReveal(snapshot *MineGridSnapshot, revealed map[int64]int64, flags, mines map[int64]bool, start int64) {
	stack := []int64{start}
	for len(stack) > 0 {
		cell := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if _, open := revealed[cell]; open || flags[cell] || mines[cell] {
			continue
		}
		count := int64(0)
		neighbors := Neighbors(snapshot.Width, snapshot.Height, cell)
		for _, neighbor := range neighbors {
			if mines[neighbor] {
				count++
			}
		}
		revealed[cell] = count
		if count == 0 {
			stack = append(stack, neighbors...)
		}
	}
	cells := make([]int64, 0, len(revealed))
	for cell := range revealed {
		cells = append(cells, cell)
	}
	sort.Slice(cells, func(i, j int) bool { return cells[i] < cells[j] })
	snapshot.Revealed = make([]RevealedCell, 0, len(cells))
	for _, cell := range cells {
		snapshot.Revealed = append(snapshot.Revealed, RevealedCell{Cell: cell, Adjacent: revealed[cell]})
	}
}

func terminateMineGrid(snapshot *MineGridSnapshot, seed uint64, outcome string) *minigame.Result {
	snapshot.Phase = MineGridTerminal
	snapshot.MineCells = []int64{}
	if snapshot.FirstCell >= 0 {
		snapshot.MineCells = MineCells(snapshot.Width, snapshot.Height, snapshot.Mines, snapshot.FirstCell, seed)
	}
	cleared := int64(0)
	if outcome == MineGridCleared {
		cleared = 1
	}
	return &minigame.Result{Outcome: outcome, RatingDelta: nil, ScoreFacts: []minigame.ScoreFact{
		{Kind: "mine_grid.cells_revealed", Value: int64(len(snapshot.Revealed))},
		{Kind: "mine_grid.cleared", Value: cleared},
	}}
}

// Neighbors are the up-to-8 cells at Chebyshev distance 1, ascending (AR3.2).
func Neighbors(width, height, cell int64) []int64 {
	x, y := cell%width, cell/width
	result := []int64{}
	for dy := int64(-1); dy <= 1; dy++ {
		for dx := int64(-1); dx <= 1; dx++ {
			nx, ny := x+dx, y+dy
			if dx == 0 && dy == 0 || nx < 0 || ny < 0 || nx >= width || ny >= height {
				continue
			}
			result = append(result, ny*width+nx)
		}
	}
	return result
}

// MineCells is AR3.3: a pure function of (seed, board, first_cell), never
// stored. The eligible list excludes first_cell and its neighbors; a
// descending Fisher–Yates with rejection-sampled bounds shuffles it and the
// first `mines` entries, sorted, are the mines.
func MineCells(width, height, mines, firstCell int64, seed uint64) []int64 {
	excluded := int64Set(append(Neighbors(width, height, firstCell), firstCell))
	eligible := []int64{}
	for cell := int64(0); cell < width*height; cell++ {
		if !excluded[cell] {
			eligible = append(eligible, cell)
		}
	}
	runSeed := determinism.Substream(seed, mineGridRunSubstream).Next()
	random := determinism.Substream(runSeed, mineGridMinesSubstream)
	for index := len(eligible) - 1; index > 0; index-- {
		swap := int(random.Bound(uint64(index + 1)))
		eligible[index], eligible[swap] = eligible[swap], eligible[index]
	}
	chosen := append([]int64(nil), eligible[:mines]...)
	sort.Slice(chosen, func(i, j int) bool { return chosen[i] < chosen[j] })
	return chosen
}

func revealedSet(snapshot *MineGridSnapshot) map[int64]int64 {
	result := make(map[int64]int64, len(snapshot.Revealed))
	for _, row := range snapshot.Revealed {
		result[row.Cell] = row.Adjacent
	}
	return result
}

func int64Set(values []int64) map[int64]bool {
	result := make(map[int64]bool, len(values))
	for _, value := range values {
		result[value] = true
	}
	return result
}

func sortedCells(values map[int64]bool) []int64 {
	result := make([]int64, 0, len(values))
	for value := range values {
		result = append(result, value)
	}
	sort.Slice(result, func(i, j int) bool { return result[i] < result[j] })
	return result
}

func decodeMineGridCommand(data []byte) (mineGridCommand, error) {
	if !uniqueJSONKeys(data) {
		return mineGridCommand{}, reject("illegal_phase", "command keys are not unique")
	}
	var header struct {
		Kind string `json:"kind"`
	}
	if json.Unmarshal(data, &header) != nil {
		return mineGridCommand{}, reject("illegal_phase", "command is not an object")
	}
	switch header.Kind {
	case "choose_board":
		var wire struct {
			Kind     string `json:"kind"`
			PresetID string `json:"preset_id"`
		}
		if !hasExactJSONKeys(data, "kind", "preset_id") || strictDecode(data, &wire) != nil {
			return mineGridCommand{}, reject("unknown_preset", "choose_board requires a preset_id string")
		}
		return mineGridCommand{Kind: header.Kind, PresetID: wire.PresetID}, nil
	case "reveal", "toggle_flag", "chord":
		var wire struct {
			Kind string          `json:"kind"`
			Cell json.RawMessage `json:"cell"`
		}
		if !hasExactJSONKeys(data, "kind", "cell") || strictDecode(data, &wire) != nil {
			return mineGridCommand{}, reject("cell_out_of_range", "cell command requires an integer cell")
		}
		cell, ok := exactInteger(wire.Cell)
		if !ok {
			return mineGridCommand{}, reject("cell_out_of_range", "cell must be a safe integer")
		}
		return mineGridCommand{Kind: header.Kind, Cell: cell}, nil
	case "quit":
		if !hasExactJSONKeys(data, "kind") {
			return mineGridCommand{}, reject("illegal_phase", "quit schema mismatch")
		}
		return mineGridCommand{Kind: header.Kind}, nil
	default:
		return mineGridCommand{}, reject("illegal_phase", "unknown command kind")
	}
}

// exactInteger accepts a JSON integer literal within the safe-integer range;
// the platform canonicalizes commands before tenants see them.
func exactInteger(raw json.RawMessage) (int64, bool) {
	text := strings.TrimSpace(string(raw))
	value, err := strconv.ParseInt(text, 10, 64)
	if err != nil || strconv.FormatInt(value, 10) != text || value < -maxSafeInteger || value > maxSafeInteger {
		return 0, false
	}
	return value, true
}

const maxSafeInteger = 9_007_199_254_740_991

func decodeMineGridSnapshot(data []byte) (MineGridSnapshot, error) {
	if !uniqueJSONKeys(data) || !hasExactJSONKeys(data, "arcade_content_hash", "arcade_schema_version", "phase", "preset_id", "width", "height",
		"mines", "first_cell", "revealed", "flags", "exploded_cell", "mine_cells", "revision") {
		return MineGridSnapshot{}, minigame.ErrInvalidTenant
	}
	var value MineGridSnapshot
	if strictDecode(data, &value) != nil || value.ArcadeSchemaVersion != SchemaVersion || !validHash(value.ArcadeContentHash) ||
		value.Phase != MineGridSetup && value.Phase != MineGridPlaying && value.Phase != MineGridTerminal || value.Revision < 1 ||
		value.Revealed == nil || value.Flags == nil || value.MineCells == nil || value.FirstCell < -1 || value.ExplodedCell < -1 {
		return MineGridSnapshot{}, minigame.ErrInvalidTenant
	}
	cells := value.Width * value.Height
	if value.PresetID == nil {
		if value.Width != 0 || value.Height != 0 || value.Mines != 0 || value.FirstCell != -1 || len(value.Revealed) != 0 ||
			len(value.Flags) != 0 || value.ExplodedCell != -1 || value.Phase == MineGridPlaying {
			return MineGridSnapshot{}, minigame.ErrInvalidTenant
		}
	} else if value.Phase == MineGridSetup || value.Width < 5 || value.Width > 30 || value.Height < 5 || value.Height > 30 ||
		value.Mines < 1 || value.Mines > cells-9 || value.FirstCell >= cells || value.ExplodedCell >= cells {
		return MineGridSnapshot{}, minigame.ErrInvalidTenant
	}
	// The hidden-information invariant: no mine position before terminal.
	if value.Phase != MineGridTerminal && (len(value.MineCells) != 0 || value.ExplodedCell != -1) {
		return MineGridSnapshot{}, minigame.ErrInvalidTenant
	}
	for index, row := range value.Revealed {
		if row.Cell < 0 || row.Cell >= cells || row.Adjacent < 0 || row.Adjacent > 8 || index > 0 && value.Revealed[index-1].Cell >= row.Cell {
			return MineGridSnapshot{}, minigame.ErrInvalidTenant
		}
	}
	for _, list := range [][]int64{value.Flags, value.MineCells} {
		for index, cell := range list {
			if cell < 0 || cell >= cells || index > 0 && list[index-1] >= cell {
				return MineGridSnapshot{}, minigame.ErrInvalidTenant
			}
		}
	}
	return value, nil
}

// validateMineGridAgainstCatalog re-derives the hidden mines and proves the
// stored state is exactly what the engine could have produced.
func validateMineGridAgainstCatalog(value MineGridSnapshot, catalog *Catalog, seed uint64) error {
	if value.PresetID == nil {
		return nil
	}
	preset, ok := catalog.Preset(*value.PresetID)
	if !ok || preset.Width != value.Width || preset.Height != value.Height || preset.Mines != value.Mines {
		return minigame.ErrInvalidTenant
	}
	if value.FirstCell == -1 {
		if len(value.Revealed) != 0 || len(value.MineCells) != 0 {
			return minigame.ErrInvalidTenant
		}
		return nil
	}
	mineList := MineCells(value.Width, value.Height, value.Mines, value.FirstCell, seed)
	mines := int64Set(mineList)
	for _, row := range value.Revealed {
		count := int64(0)
		for _, neighbor := range Neighbors(value.Width, value.Height, row.Cell) {
			if mines[neighbor] {
				count++
			}
		}
		if mines[row.Cell] || count != row.Adjacent {
			return minigame.ErrInvalidTenant
		}
	}
	if value.Phase == MineGridTerminal {
		if len(value.MineCells) != len(mineList) {
			return minigame.ErrInvalidTenant
		}
		for index, cell := range mineList {
			if value.MineCells[index] != cell {
				return minigame.ErrInvalidTenant
			}
		}
		if value.ExplodedCell != -1 && !mines[value.ExplodedCell] {
			return minigame.ErrInvalidTenant
		}
	}
	return nil
}

func validHash(value string) bool {
	return strings.HasPrefix(value, "sha256:") && len(value) == 71
}

func encodeCanonical(value any) (json.RawMessage, error) {
	encoded, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	var canonical map[string]json.RawMessage
	if json.Unmarshal(encoded, &canonical) != nil {
		return nil, minigame.ErrTenantDivergence
	}
	return json.Marshal(canonical)
}

func catalogForInput(data []byte, hash string, schemaVersion int) (*Catalog, error) {
	if len(data) == 0 || hash != ContentHash(data) || schemaVersion != SchemaVersion {
		return nil, ErrInvalidCatalog
	}
	keys := map[string]struct{}{}
	for _, key := range copykeys.All() {
		keys[key] = struct{}{}
	}
	return LoadCatalog(data, Declarations{CopyKeys: keys})
}

// validLiteralScaling is AR2's only scaling input: the literal breadth 1.
func validLiteralScaling(values map[string]int64, destination string) bool {
	value, ok := values[destination]
	return len(values) == 1 && ok && value == 1
}

func reject(code, detail string) *minigame.Rejection {
	return &minigame.Rejection{Code: code, Detail: detail}
}

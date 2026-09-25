package arcade

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"cloud-clicker/server/minigame"
)

// session drives one tenant through the real registry so every transition
// is canonicalized and validated exactly as the platform does it.
type session struct {
	t        *testing.T
	registry *minigame.TenantRegistry
	engine   string
	content  []byte
	hash     string
	seed     uint64
	scaling  map[string]int64
	snapshot json.RawMessage
	revision int64
	result   *minigame.Result
}

func newSession(t *testing.T, engine string, seed uint64, contentPath string) *session {
	t.Helper()
	registry, err := minigame.NewTenantRegistry(NewMineGridTenant(), NewSnakeTenant())
	if err != nil {
		t.Fatal(err)
	}
	content := readFile(t, contentPath)
	destination := MineGridScalingDestination
	if engine == SnakeEngineRef {
		destination = SnakeScalingDestination
	}
	s := &session{t: t, registry: registry, engine: engine, content: content, hash: ContentHash(content), seed: seed,
		scaling: map[string]int64{destination: 1}, revision: 1}
	snapshot, err := registry.Create(engine, EngineVersion, minigame.CreateInput{Mode: minigame.ModeSolo, Seed: seed,
		ScalingInputs: s.scaling, Content: content, ContentHash: s.hash, ContentSchemaVersion: SchemaVersion})
	if err != nil {
		t.Fatalf("create %s: %v", engine, err)
	}
	s.snapshot = snapshot
	return s
}

func (s *session) apply(command string) error {
	s.t.Helper()
	output, err := s.registry.Apply(s.engine, EngineVersion, minigame.ApplyInput{Mode: minigame.ModeSolo, Seed: s.seed,
		Revision: s.revision, Snapshot: s.snapshot, Command: canonical(s.t, command), ScalingInputs: s.scaling,
		Content: s.content, ContentHash: s.hash, ContentSchemaVersion: SchemaVersion})
	if err != nil {
		return err
	}
	s.snapshot, s.result = output.Snapshot, output.Result
	s.revision++
	return nil
}

func canonical(t *testing.T, command string) json.RawMessage {
	t.Helper()
	var value map[string]any
	if err := json.Unmarshal([]byte(command), &value); err != nil {
		t.Fatalf("command %s: %v", command, err)
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return encoded
}

func rejectionCode(err error) string {
	var rejection *minigame.Rejection
	if errors.As(err, &rejection) && rejection != nil {
		return rejection.Code
	}
	return ""
}

func (s *session) mineGrid() MineGridSnapshot {
	s.t.Helper()
	value, err := decodeMineGridSnapshot(s.snapshot)
	if err != nil {
		s.t.Fatal(err)
	}
	return value
}

func (s *session) snake() SnakeSnapshot {
	s.t.Helper()
	value, err := decodeSnakeSnapshot(s.snapshot)
	if err != nil {
		s.t.Fatal(err)
	}
	return value
}

func TestMineGridNeverExposesMinesBeforeTerminal(t *testing.T) {
	s := newSession(t, MineGridEngineRef, 7, candidatePath)
	if err := s.apply(`{"kind":"choose_board","preset_id":"small"}`); err != nil {
		t.Fatal(err)
	}
	if err := s.apply(`{"kind":"reveal","cell":40}`); err != nil {
		t.Fatal(err)
	}
	if snapshot := s.mineGrid(); snapshot.Phase != MineGridPlaying || len(snapshot.MineCells) != 0 ||
		strings.Contains(string(s.snapshot), `"mine_cells":[0`) {
		t.Fatalf("mines leaked before terminal: %s", s.snapshot)
	}
	// Failing case: a stored non-terminal snapshot carrying mines is refused.
	leaked := strings.Replace(string(s.snapshot), `"mine_cells":[]`, `"mine_cells":[1]`, 1)
	if _, err := decodeMineGridSnapshot([]byte(leaked)); !errors.Is(err, minigame.ErrInvalidTenant) {
		t.Fatalf("a non-terminal snapshot with mine positions must be invalid, got %v", err)
	}
	if err := s.apply(`{"kind":"quit"}`); err != nil {
		t.Fatal(err)
	}
	final := s.mineGrid()
	if final.Phase != MineGridTerminal || int64(len(final.MineCells)) != final.Mines {
		t.Fatalf("terminal snapshot must hold every mine: %s", s.snapshot)
	}
}

func TestMineGridFirstRevealIsAlwaysSafeAndPlacementIsPure(t *testing.T) {
	for seed := uint64(1); seed <= 64; seed++ {
		first := int64(seed % 81)
		mines := MineCells(9, 9, 10, first, seed)
		excluded := int64Set(append(Neighbors(9, 9, first), first))
		if len(mines) != 10 {
			t.Fatalf("seed %d: %d mines", seed, len(mines))
		}
		for index, cell := range mines {
			if excluded[cell] || index > 0 && mines[index-1] >= cell {
				t.Fatalf("seed %d: mine %d violates first-reveal exclusion or order: %v", seed, cell, mines)
			}
		}
		again := MineCells(9, 9, 10, first, seed)
		for index := range mines {
			if mines[index] != again[index] {
				t.Fatalf("seed %d: placement is not pure", seed)
			}
		}
	}
}

func TestMineGridRejectionsMutateNothing(t *testing.T) {
	s := newSession(t, MineGridEngineRef, 3, candidatePath)
	cases := []struct{ command, code string }{
		{`{"kind":"reveal","cell":0}`, "illegal_phase"},
		{`{"kind":"choose_board","preset_id":"huge"}`, "unknown_preset"},
	}
	for _, c := range cases {
		before := string(s.snapshot)
		if code := rejectionCode(s.apply(c.command)); code != c.code || string(s.snapshot) != before {
			t.Fatalf("%s: got %q and snapshot change=%v", c.command, code, string(s.snapshot) != before)
		}
	}
	if err := s.apply(`{"kind":"choose_board","preset_id":"small"}`); err != nil {
		t.Fatal(err)
	}
	for _, c := range []struct{ command, code string }{
		{`{"kind":"reveal","cell":81}`, "cell_out_of_range"},
		{`{"kind":"reveal","cell":-1}`, "cell_out_of_range"},
		{`{"kind":"chord","cell":0}`, "chord_unsatisfied"},
	} {
		before := string(s.snapshot)
		if code := rejectionCode(s.apply(c.command)); code != c.code || string(s.snapshot) != before {
			t.Fatalf("%s: got %q", c.command, code)
		}
	}
}

func TestSnakeRejectsAdvancePastTerminalWithoutMutation(t *testing.T) {
	s := newSession(t, SnakeEngineRef, 5, fixturePath)
	// The 5x5 head starts at x=2 facing right: it crashes into the east wall at tick 3.
	before := string(s.snapshot)
	if code := rejectionCode(s.apply(`{"kind":"advance","through_tick":5,"turns":[]}`)); code != "advance_past_terminal" || string(s.snapshot) != before {
		t.Fatalf("advance past a mid-window terminal must reject without mutation, got %q", code)
	}
	if err := s.apply(`{"kind":"advance","through_tick":3,"turns":[]}`); err != nil {
		t.Fatal(err)
	}
	if snapshot := s.snake(); snapshot.Phase != SnakeTerminal || snapshot.Tick != 3 || s.result.Outcome != SnakeCrashed {
		t.Fatalf("expected a crash at tick 3: %s", s.snapshot)
	}
}

func TestSnakeGenesisPublishesFoodSeedAndOffersPureFood(t *testing.T) {
	s := newSession(t, SnakeEngineRef, 11, candidatePath)
	snapshot := s.snake()
	if snapshot.FoodSeed == "" || snapshot.FoodIndex != 1 || snapshot.Body[0] != 10*20+10 || len(snapshot.Body) != 3 || snapshot.Body[2] != 10*20+8 {
		t.Fatalf("genesis differs from AR4.3: %s", s.snapshot)
	}
	again := newSession(t, SnakeEngineRef, 11, candidatePath)
	if string(again.snapshot) != string(s.snapshot) {
		t.Fatal("genesis is not a pure function of the seed")
	}
}

func TestArcadeTenantsRejectWrongScaling(t *testing.T) {
	content := readFile(t, candidatePath)
	registry, err := minigame.NewTenantRegistry(NewMineGridTenant(), NewSnakeTenant())
	if err != nil {
		t.Fatal(err)
	}
	for _, value := range []int64{0, 2} {
		_, err := registry.Create(MineGridEngineRef, EngineVersion, minigame.CreateInput{Mode: minigame.ModeSolo, Seed: 1,
			ScalingInputs: map[string]int64{MineGridScalingDestination: value}, Content: content, ContentHash: ContentHash(content), ContentSchemaVersion: 1})
		if err == nil {
			t.Fatalf("literal breadth %d must be refused", value)
		}
	}
}

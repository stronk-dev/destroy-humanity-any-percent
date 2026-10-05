package arcade

import (
	"bytes"
	"encoding/json"
	"flag"
	"hash/fnv"
	"os"
	"strconv"
	"testing"

	"cloud-clicker/server/determinism"
)

var updateMineGridSampling = flag.Bool("update-arcade-sampling-corpus", false, "regenerate the real-seed Mine Grid rejection-sampling witness")

const samplingPath = "../../testdata/arcade/mine-grid-sampling-v1.json"

type mineGridSamplingWitness struct {
	Version          int            `json:"version"`
	ContentHash      string         `json:"arcade_content_hash"`
	Seed             string         `json:"seed"`
	RunSeed          string         `json:"run_seed"`
	Bound            uint64         `json:"bound"`
	Threshold        uint64         `json:"threshold"`
	FirstDraw        string         `json:"first_draw"`
	AcceptedDraw     string         `json:"accepted_draw"`
	Mines            []int64        `json:"mines"`
	TransitionBudget int            `json:"transition_budget"`
	Scenario         corpusScenario `json:"scenario"`
}

// Invert SplitMix64's bijective mixing, then its increment and both published
// substream XORs. This constructs one real seed; no search ceiling or fake RNG.
func mineGridZeroDrawSeed() uint64 {
	labelHash := func(label string) uint64 {
		h := fnv.New64a()
		_, _ = h.Write([]byte(label))
		return h.Sum64()
	}
	unxor := func(value uint64, shift uint) uint64 {
		x := value
		for i := 0; i < 7; i++ {
			x = value ^ (x >> shift)
		}
		return x
	}
	inverse := func(odd uint64) uint64 {
		x := odd
		for i := 0; i < 6; i++ {
			x *= 2 - odd*x
		}
		return x
	}
	increment := uint64(0x9e3779b97f4a7c15)
	runSeed := (uint64(0) - increment) ^ labelHash(mineGridMinesSubstream)
	value := unxor(runSeed, 31) * inverse(0x94d049bb133111eb)
	value = unxor(value, 27) * inverse(0xbf58476d1ce4e5b9)
	return (unxor(value, 30) - increment) ^ labelHash(mineGridRunSubstream)
}

func TestMineGridRejectionSamplingContentGate(t *testing.T) {
	seed := mineGridZeroDrawSeed()
	if seed != 15581846558861750132 {
		t.Fatal("inverse construction changed its independently derived seed")
	}
	bound := uint64(72) // 9×9 minus first cell 40 and its eight neighbors.
	threshold := (-bound) % bound
	runSeed := determinism.Substream(seed, mineGridRunSubstream).Next()
	raw := determinism.Substream(runSeed, mineGridMinesSubstream)
	first, accepted := raw.Next(), raw.Next()
	if runSeed != 2387092019343320515 || threshold != 16 || first != 0 || accepted != 16294208416658607535 || first >= threshold || accepted < threshold {
		t.Fatal("actual RNG does not exercise exactly the declared rejected/accepted draws")
	}
	bounded := determinism.Substream(runSeed, mineGridMinesSubstream)
	if bounded.Bound(bound) != accepted%bound || bounded.Next() != raw.Next() {
		t.Fatal("actual Bound did not consume the rejected first and accepted second draws")
	}
	b := newBuilder(t, "mine_grid_rejected_first_shuffle_draw", MineGridEngineRef, seed)
	b.must(`{"kind":"choose_board","preset_id":"large"}`)
	b.must(cellCommand("reveal", 40))
	state := b.s.mineGrid()
	if state.Width != 9 || state.Height != 9 || state.Mines != 10 || state.FirstCell != 40 || state.Phase != MineGridPlaying || len(Neighbors(9, 9, 40))+1 != 9 {
		t.Fatal("actual transition did not reach the pinned existing board population")
	}
	b.must(`{"kind":"quit"}`)
	b.expectReject(`{"kind":"quit"}`, "illegal_phase")
	scenario := b.finish(MineGridQuit)
	final := b.s.mineGrid()
	if len(final.MineCells) != 10 || final.Revision != 4 || len(scenario.Steps) != 4 {
		t.Fatal("wrong actual terminal/command population")
	}
	witness := mineGridSamplingWitness{Version: 1, ContentHash: b.s.hash, Seed: strconv.FormatUint(seed, 10),
		RunSeed: strconv.FormatUint(runSeed, 10), Bound: bound, Threshold: threshold,
		FirstDraw: strconv.FormatUint(first, 10), AcceptedDraw: strconv.FormatUint(accepted, 10),
		Mines: final.MineCells, TransitionBudget: len(scenario.Steps), Scenario: scenario}
	encoded, err := json.MarshalIndent(witness, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	encoded = append(encoded, '\n')
	if *updateMineGridSampling {
		if err := os.WriteFile(samplingPath, encoded, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if existing, err := os.ReadFile(samplingPath); err != nil || !bytes.Equal(existing, encoded) {
		t.Fatalf("real-seed sampling witness stale: run make arcade-corpus (%v)", err)
	}
	t.Logf("real seed %d rejects draw %d below threshold %d at bound %d; four attempts/five literal states", seed, first, threshold, bound)
}

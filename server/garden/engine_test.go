package garden

import (
	"bytes"
	"encoding/json"
	"flag"
	"math/rand"
	"os"
	"testing"
)

var updateGardenCorpus = flag.Bool("update-garden-corpus", false, "regenerate the checked-in garden engine corpus")

const engineCorpusPath = "testdata/garden/engine-corpus-v1.json"

type corpusStep struct {
	Op          string          `json:"op"`
	ServerMS    int64           `json:"server_ms,omitempty"`
	HostLevel   int64           `json:"host_level,omitempty"`
	Unlocked    bool            `json:"unlocked,omitempty"`
	Salt        string          `json:"salt,omitempty"`
	Row         int64           `json:"row,omitempty"`
	Col         int64           `json:"col,omitempty"`
	SpeciesID   string          `json:"species_id,omitempty"`
	SubstrateID string          `json:"substrate_id,omitempty"`
	IntentID    string          `json:"intent_id,omitempty"`
	Plots       []HarvestTarget `json:"plots,omitempty"`
}

type corpusRejection struct {
	Category string `json:"category"`
	Detail   string `json:"detail"`
}

type corpusOutcome struct {
	Result    json.RawMessage  `json:"result"`
	Rejection *corpusRejection `json:"rejection"`
	State     *State           `json:"state"`
}

type corpusScenario struct {
	Name       string          `json:"name"`
	CatalogOps []fixtureOp     `json:"catalog_ops"`
	Initial    *State          `json:"initial"`
	Steps      []corpusStep    `json:"steps"`
	Outcomes   []corpusOutcome `json:"outcomes"`
}

type engineCorpus struct {
	Version   int              `json:"version"`
	Scenarios []corpusScenario `json:"scenarios"`
}

func setOp(value any, path ...any) fixtureOp {
	data, _ := json.Marshal(value)
	return fixtureOp{Op: "set", Path: path, Value: data}
}

func scenarioCatalog(t testing.TB, ops []fixtureOp) *Catalog {
	t.Helper()
	corpus := loadCorpus(t)
	data := applyCase(t, readRepo(t, corpus.Valid), fixtureCase{Ops: normalizeOps(ops)})
	catalog, err := LoadCatalog(data, fixtureDeclarations(t, corpus))
	if err != nil {
		t.Fatalf("scenario catalog: %v", err)
	}
	return catalog
}

// normalizeOps round-trips ops through JSON so integer path segments are
// float64, exactly as they are after reading the corpus back.
func normalizeOps(ops []fixtureOp) []fixtureOp {
	data, _ := json.Marshal(ops)
	var normalized []fixtureOp
	_ = json.Unmarshal(data, &normalized)
	return normalized
}

// runStep drives one step through the real engine.
func runStep(catalog *Catalog, state *State, step corpusStep) (any, error) {
	switch step.Op {
	case "advance":
		return catalog.Advance(state, AdvanceInput{ServerMS: step.ServerMS, HostLevel: step.HostLevel, Unlocked: step.Unlocked, Salt: step.Salt})
	case "plant":
		return catalog.Plant(state, step.HostLevel, step.Row, step.Col, step.SpeciesID)
	case "uproot":
		return catalog.Uproot(state, step.Row, step.Col)
	case "set_substrate":
		return catalog.SetSubstrate(state, step.ServerMS, step.SubstrateID)
	case "harvest":
		return catalog.Harvest(state, step.IntentID, step.Plots)
	}
	panic("unknown corpus op " + step.Op)
}

func record(t testing.TB, catalog *Catalog, state *State, step corpusStep) corpusOutcome {
	t.Helper()
	working := state.Clone()
	result, err := runStep(catalog, working, step)
	outcome := corpusOutcome{Result: json.RawMessage("null")}
	if err != nil {
		rejection, ok := AsRejection(err)
		if !ok {
			t.Fatalf("step %+v: non-rejection error %v", step, err)
		}
		outcome.Rejection = &corpusRejection{Category: rejection.Category, Detail: rejection.Detail}
		outcome.State = state.Clone()
		return outcome
	}
	data, marshalErr := json.Marshal(result)
	if marshalErr != nil {
		t.Fatal(marshalErr)
	}
	outcome.Result, outcome.State = data, working
	*state = *working.Clone()
	return outcome
}

type scenarioBuilder struct {
	t        testing.TB
	catalog  *Catalog
	state    *State
	scenario corpusScenario
}

func newScenario(t testing.TB, name string, ops []fixtureOp, initial *State) *scenarioBuilder {
	catalog := scenarioCatalog(t, ops)
	if initial == nil {
		initial = NewState(catalog)
	}
	if err := ValidateAgainst(catalog, initial); err != nil {
		t.Fatalf("%s initial state: %v", name, err)
	}
	return &scenarioBuilder{t: t, catalog: catalog, state: initial.Clone(), scenario: corpusScenario{Name: name, CatalogOps: normalizeOps(ops), Initial: initial.Clone()}}
}

func (b *scenarioBuilder) do(step corpusStep) corpusOutcome {
	b.t.Helper()
	outcome := record(b.t, b.catalog, b.state, step)
	b.scenario.Steps = append(b.scenario.Steps, step)
	b.scenario.Outcomes = append(b.scenario.Outcomes, outcome)
	return outcome
}

func (b *scenarioBuilder) expectReject(step corpusStep, detail string) {
	b.t.Helper()
	if outcome := b.do(step); outcome.Rejection == nil || outcome.Rejection.Detail != detail {
		b.t.Fatalf("%s: %+v expected rejection %s, got %+v", b.scenario.Name, step, detail, outcome.Rejection)
	}
}

func (b *scenarioBuilder) advance(serverMS, level int64) Advance {
	b.t.Helper()
	step := corpusStep{Op: "advance", ServerMS: serverMS, HostLevel: level, Unlocked: true}
	if NeedsSalt(b.state, true) {
		step.Salt = "0123456789abcdef"
	}
	outcome := b.do(step)
	if outcome.Rejection != nil {
		b.t.Fatalf("advance rejected: %+v", outcome.Rejection)
	}
	var advance Advance
	if err := json.Unmarshal(outcome.Result, &advance); err != nil {
		b.t.Fatal(err)
	}
	return advance
}

func ppmPointer(value int64) *int64 { return &value }

func matureState(t testing.TB, plots ...Plot) *State {
	t.Helper()
	state := &State{SubstrateID: "bare_metal", SeedCollection: []string{"strain_a", "strain_b"}, Plots: plots}
	salt, anchor := "0123456789abcdef", int64(1_000_000)
	state.SaltHex, state.TickAnchorWallMS = &salt, &anchor
	return state
}

const tick = int64(300_000)

// buildEngineCorpus is AC3/AC4/AC5's shared vector set, generated by driving
// the real engine; TS replays it byte-for-byte.
func buildEngineCorpus(t testing.TB) engineCorpus {
	scenarios := []corpusScenario{}

	// Maturation, clock anchor, and spread with the salt drawn on first advance.
	b := newScenario(t, "maturation_then_spread", []fixtureOp{setOp(900_000, "recipes", 0, "chance_ppm")}, nil)
	b.do(corpusStep{Op: "advance", ServerMS: 5_000, HostLevel: 0, Unlocked: false})
	b.advance(10_000, 0)
	b.do(corpusStep{Op: "plant", HostLevel: 0, Row: 0, Col: 0, SpeciesID: "strain_a"})
	if a := b.advance(10_000+3*tick, 0); len(a.Matured) != 1 || a.TicksApplied != 3 {
		t.Fatalf("maturation vector %+v", a)
	}
	if a := b.advance(10_000+6*tick+123, 0); len(a.Spawned) == 0 {
		t.Fatalf("spread vector %+v", a)
	}
	scenarios = append(scenarios, b.scenario)

	// Two-parent cross where strain_b is only a diagonal neighbour (OD-6).
	crossOps := []fixtureOp{setOp(1, "recipes", 0, "chance_ppm"), setOp(1, "recipes", 1, "chance_ppm"), setOp(999_990, "recipes", 2, "chance_ppm")}
	b = newScenario(t, "diagonal_only_cross", crossOps, matureState(t,
		Plot{Row: 0, Col: 0, SpeciesID: "strain_a", AgeTicks: 3, MaturedEffectPPM: ppmPointer(1_000_000)},
		Plot{Row: 1, Col: 2, SpeciesID: "strain_b", AgeTicks: 4, MaturedEffectPPM: ppmPointer(1_000_000)}))
	if a := b.advance(1_000_000+tick, 1); len(a.Spawned) == 0 || a.Spawned[0].RecipeID != "r_c_cross" || a.Spawned[0].Col != 1 || a.Spawned[0].Row != 0 {
		t.Fatalf("diagonal cross vector %+v", a)
	}
	scenarios = append(scenarios, b.scenario)

	// min_mature_neighbours > 1: three mature strain_a around one empty plot.
	b = newScenario(t, "cluster_min_three", []fixtureOp{setOp(1, "recipes", 0, "chance_ppm"), setOp(999_990, "recipes", 4, "chance_ppm")}, matureState(t,
		Plot{Row: 0, Col: 0, SpeciesID: "strain_a", AgeTicks: 3, MaturedEffectPPM: ppmPointer(1_000_000)},
		Plot{Row: 0, Col: 1, SpeciesID: "strain_a", AgeTicks: 3, MaturedEffectPPM: ppmPointer(1_000_000)},
		Plot{Row: 1, Col: 0, SpeciesID: "strain_a", AgeTicks: 3, MaturedEffectPPM: ppmPointer(1_000_000)}))
	if a := b.advance(1_000_000+tick, 0); len(a.Spawned) != 1 || a.Spawned[0].SpeciesID != "strain_e" {
		t.Fatalf("cluster vector %+v", a)
	}
	scenarios = append(scenarios, b.scenario)

	// Dormant plots never grow and never count as neighbours.
	b = newScenario(t, "dormant_plots", []fixtureOp{setOp(1_000_000, "recipes", 0, "chance_ppm")}, matureState(t,
		Plot{Row: 0, Col: 2, SpeciesID: "strain_a", AgeTicks: 3, MaturedEffectPPM: ppmPointer(1_000_000)},
		Plot{Row: 2, Col: 0, SpeciesID: "strain_b", AgeTicks: 0}))
	if a := b.advance(1_000_000+5*tick, 0); a.TicksApplied != 5 || len(a.Spawned) != 0 || len(a.Matured) != 0 {
		t.Fatalf("dormant vector %+v", a)
	}
	b.expectReject(corpusStep{Op: "plant", HostLevel: 0, Row: 0, Col: 2, SpeciesID: "strain_a"}, DetailPlotDormant)
	scenarios = append(scenarios, b.scenario)

	// Chaos factor and the cumulative clamp in canonical recipe order (OD-21).
	chaos := []fixtureOp{setOp(500_000, "recipes", 0, "chance_ppm"), setOp(500_000, "recipes", 4, "chance_ppm")}
	b = newScenario(t, "chaos_cumulative_clamp", chaos, matureState(t,
		Plot{Row: 0, Col: 0, SpeciesID: "strain_a", AgeTicks: 3, MaturedEffectPPM: ppmPointer(1_000_000)},
		Plot{Row: 0, Col: 1, SpeciesID: "strain_a", AgeTicks: 3, MaturedEffectPPM: ppmPointer(1_000_000)},
		Plot{Row: 1, Col: 0, SpeciesID: "strain_a", AgeTicks: 3, MaturedEffectPPM: ppmPointer(1_000_000)}))
	b.state.SubstrateID = "chaos_monkey"
	b.scenario.Initial.SubstrateID = "chaos_monkey"
	if a := b.advance(1_000_000+tick, 0); len(a.Spawned) != 1 || a.Spawned[0].RecipeID != "r_a_spread" {
		t.Fatalf("clamped chaos vector must always pick the first recipe: %+v", a)
	}
	scenarios = append(scenarios, b.scenario)

	// Fixed-point skip: nothing grows and nothing is eligible.
	b = newScenario(t, "fixed_point_skip", nil, matureState(t,
		Plot{Row: 0, Col: 0, SpeciesID: "strain_e", AgeTicks: 5, MaturedEffectPPM: ppmPointer(1_000_000)}))
	if a := b.advance(1_000_000+80_000_000, 0); a.TicksApplied != 266 || a.TickSeqAfter != 266 || a.Visible() {
		t.Fatalf("fixed-point vector %+v", a)
	}
	scenarios = append(scenarios, b.scenario)

	// A retune lowering maturation below age matures on the next tick.
	b = newScenario(t, "retune_below_age", []fixtureOp{setOp(3, "species", 3, "maturation_ticks")}, matureState(t,
		Plot{Row: 0, Col: 0, SpeciesID: "strain_d", AgeTicks: 5}))
	if a := b.advance(1_000_000+tick, 0); len(a.Matured) != 1 {
		t.Fatalf("retune vector %+v", a)
	}
	scenarios = append(scenarios, b.scenario)

	// Catch-up hardcap (AC5), clock regression, and the locked garden.
	b = newScenario(t, "catchup_cap_and_regression", nil, matureState(t, Plot{Row: 0, Col: 0, SpeciesID: "strain_a", AgeTicks: 0}))
	if a := b.advance(1_000_000+30*3_600_000, 0); a.CatchupForfeitedMS != 6*3_600_000 || a.CatchupReasonKey == nil || !a.Visible() {
		t.Fatalf("catch-up vector %+v", a)
	}
	if a := b.advance(500, 0); a.TicksApplied != 0 {
		t.Fatalf("regression vector %+v", a)
	}
	b.do(corpusStep{Op: "advance", ServerMS: 999_999_999, HostLevel: 0, Unlocked: false})
	scenarios = append(scenarios, b.scenario)

	// Effect frozen at maturity (OD-8), then harvest math and seeds.
	b = newScenario(t, "effect_frozen_harvest", nil, matureState(t,
		Plot{Row: 0, Col: 0, SpeciesID: "strain_a", AgeTicks: 2},
		Plot{Row: 0, Col: 1, SpeciesID: "strain_c", AgeTicks: 6, MaturedEffectPPM: ppmPointer(750_000)},
		Plot{Row: 1, Col: 1, SpeciesID: "strain_e", AgeTicks: 5, MaturedEffectPPM: ppmPointer(1_000_000)}))
	b.state.SubstrateID, b.scenario.Initial.SubstrateID = "mainframe", "mainframe"
	if a := b.advance(1_000_000+900_000, 0); len(a.Matured) != 1 {
		t.Fatalf("mainframe maturity %+v", a)
	}
	b.do(corpusStep{Op: "set_substrate", ServerMS: 1_000_000 + 900_000 + 1_000, SubstrateID: "bare_metal"})
	b.expectReject(corpusStep{Op: "harvest", IntentID: "01986666-0000-7000-8000-000000000001", Plots: []HarvestTarget{{Row: 5, Col: 5}}}, DetailUnknownPlot)
	outcome := b.do(corpusStep{Op: "harvest", IntentID: "01986666-0000-7000-8000-000000000002", Plots: []HarvestTarget{{Row: 0, Col: 0}, {Row: 0, Col: 1}, {Row: 1, Col: 1}}})
	var harvest Harvest
	if err := json.Unmarshal(outcome.Result, &harvest); err != nil || harvest.TotalUnits != 6+15+0 || len(harvest.SeedsDiscovered) != 2 {
		t.Fatalf("harvest vector %+v %v", harvest, err)
	}
	scenarios = append(scenarios, b.scenario)

	// Every command rejection detail (AC7's pure half).
	b = newScenario(t, "command_rejections", nil, matureState(t,
		Plot{Row: 0, Col: 0, SpeciesID: "strain_a", AgeTicks: 1}))
	b.expectReject(corpusStep{Op: "plant", HostLevel: 0, Row: 3, Col: 0, SpeciesID: "strain_a"}, DetailPlotDormant)
	b.expectReject(corpusStep{Op: "plant", HostLevel: 0, Row: 0, Col: 0, SpeciesID: "strain_a"}, DetailPlotOccupied)
	b.expectReject(corpusStep{Op: "plant", HostLevel: 0, Row: 0, Col: 1, SpeciesID: "strain_zz"}, DetailUnknownSpecies)
	b.expectReject(corpusStep{Op: "plant", HostLevel: 0, Row: 0, Col: 1, SpeciesID: "strain_c"}, DetailSeedNotCollected)
	b.expectReject(corpusStep{Op: "uproot", Row: 1, Col: 1}, DetailUnknownPlot)
	b.expectReject(corpusStep{Op: "harvest", IntentID: "01986666-0000-7000-8000-000000000003", Plots: []HarvestTarget{{Row: 0, Col: 0}}}, DetailNotMature)
	b.expectReject(corpusStep{Op: "set_substrate", ServerMS: 2_000_000, SubstrateID: "quantum"}, DetailUnknownSubstrate)
	b.expectReject(corpusStep{Op: "set_substrate", ServerMS: 2_000_000, SubstrateID: "bare_metal"}, DetailUnchanged)
	b.do(corpusStep{Op: "set_substrate", ServerMS: 2_000_000, SubstrateID: "containerized"})
	b.expectReject(corpusStep{Op: "set_substrate", ServerMS: 2_000_000 + 599_999, SubstrateID: "mainframe"}, DetailLockout)
	b.do(corpusStep{Op: "set_substrate", ServerMS: 2_000_000 + 600_000, SubstrateID: "mainframe"})
	b.do(corpusStep{Op: "plant", HostLevel: 8, Row: 5, Col: 5, SpeciesID: "strain_b"})
	b.do(corpusStep{Op: "uproot", Row: 0, Col: 0})
	scenarios = append(scenarios, b.scenario)

	return engineCorpus{Version: 1, Scenarios: scenarios}
}

func TestGardenEngineCorpus(t *testing.T) {
	corpus := buildEngineCorpus(t)
	data, err := json.MarshalIndent(corpus, "", " ")
	if err != nil {
		t.Fatal(err)
	}
	data = append(data, '\n')
	path := repositoryRoot + engineCorpusPath
	if *updateGardenCorpus {
		if err := os.WriteFile(path, data, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	stored, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(stored, data) {
		t.Fatal("garden engine corpus drifted from the engine; run make garden-corpus and review the diff")
	}
	// Replaying the stored corpus reproduces every outcome.
	var decoded engineCorpus
	if err := json.Unmarshal(stored, &decoded); err != nil {
		t.Fatal(err)
	}
	for _, scenario := range decoded.Scenarios {
		catalog := scenarioCatalog(t, scenario.CatalogOps)
		state := scenario.Initial.Clone()
		for index, step := range scenario.Steps {
			got := record(t, catalog, state, step)
			want := scenario.Outcomes[index]
			gotJSON, _ := json.Marshal(got)
			wantJSON, _ := json.Marshal(want)
			if !bytes.Equal(gotJSON, wantJSON) {
				t.Fatalf("%s step %d diverged:\n got %s\nwant %s", scenario.Name, index, gotJSON, wantJSON)
			}
		}
	}
}

// TestGardenPartitionInvariance is AC3: below the cap, advancing t0→t2 in one
// step equals t0→t1→t2 for every t1, across seeds and substrates.
func TestGardenPartitionInvariance(t *testing.T) {
	catalog := scenarioCatalog(t, []fixtureOp{setOp(200_000, "recipes", 0, "chance_ppm"), setOp(200_000, "recipes", 1, "chance_ppm"), setOp(100_000, "recipes", 2, "chance_ppm")})
	random := rand.New(rand.NewSource(7))
	for trial := 0; trial < 120; trial++ {
		start := matureState(t, Plot{Row: 0, Col: 0, SpeciesID: "strain_a"}, Plot{Row: 2, Col: 2, SpeciesID: "strain_b"})
		salt := []byte("0123456789abcdef")
		for index := range salt {
			salt[index] = "0123456789abcdef"[random.Intn(16)]
		}
		saltText := string(salt)
		start.SaltHex = &saltText
		start.SubstrateID = []string{"bare_metal", "containerized", "mainframe", "chaos_monkey"}[trial%4]
		level := int64(random.Intn(9))
		end := int64(1_000_000) + int64(random.Intn(int(catalog.CatchupCapMS)-1)) + 1
		whole := start.Clone()
		if _, err := catalog.Advance(whole, AdvanceInput{ServerMS: end, HostLevel: level, Unlocked: true}); err != nil {
			t.Fatal(err)
		}
		for split := 0; split < 4; split++ {
			middle := int64(1_000_000) + int64(random.Intn(int(end-1_000_000)))
			parts := start.Clone()
			for _, at := range []int64{middle, end} {
				if _, err := catalog.Advance(parts, AdvanceInput{ServerMS: at, HostLevel: level, Unlocked: true}); err != nil {
					t.Fatal(err)
				}
			}
			if !whole.Equal(parts) {
				t.Fatalf("trial %d split %d: partition diverged at t1=%d", trial, split, middle)
			}
		}
	}
}

// TestGardenCursorPRNGBreaksPartition is AC3's failing case: a variant that
// stores a PRNG cursor (continuing one stream across advances instead of
// keying draws by absolute tick) diverges under a split.
func TestGardenCursorPRNGBreaksPartition(t *testing.T) {
	catalog := scenarioCatalog(t, []fixtureOp{setOp(200_000, "recipes", 0, "chance_ppm")})
	diverged := false
	for seed := 0; seed < 40 && !diverged; seed++ {
		run := func(stops []int64) *State {
			state := matureState(t, Plot{Row: 0, Col: 0, SpeciesID: "strain_a", AgeTicks: 3, MaturedEffectPPM: ppmPointer(1_000_000)})
			cursor := uint64(seed)
			for _, stop := range stops {
				cursorAdvance(catalog, state, stop, &cursor)
			}
			return state
		}
		whole := run([]int64{1_000_000 + 40*tick})
		split := run([]int64{1_000_000 + 13*tick, 1_000_000 + 40*tick})
		diverged = !whole.Equal(split)
	}
	if !diverged {
		t.Fatal("a stored-cursor PRNG variant must break partition invariance on some seed")
	}
}

// cursorAdvance is the deliberately wrong variant: one draw stream continues
// across advances, so where the advance boundaries fall changes the draws.
func cursorAdvance(catalog *Catalog, state *State, serverMS int64, cursor *uint64) {
	dimension := catalog.Dimension(0)
	substrate, _ := catalog.Substrate(state.SubstrateID)
	ticks := (serverMS - *state.TickAnchorWallMS) / substrate.TickMS
	for step := int64(1); step <= ticks; step++ {
		*cursor++
		scratch := Advance{}
		catalog.tick(state, dimension, substrate, *cursor*uint64(step), 0, &scratch)
	}
	anchor := *state.TickAnchorWallMS + ticks*substrate.TickMS
	state.TickAnchorWallMS = &anchor
	state.TickSeq += ticks
}

func TestGardenAdvanceSaltContract(t *testing.T) {
	catalog := fixtureCatalog(t)
	state := NewState(catalog)
	if _, err := catalog.Advance(state, AdvanceInput{ServerMS: 1, Unlocked: true}); err == nil {
		t.Fatal("an initializing advance without a salt must fail")
	}
	if _, err := catalog.Advance(state, AdvanceInput{ServerMS: 1, Unlocked: false, Salt: "0123456789abcdef"}); err == nil {
		t.Fatal("a salt on a locked advance must fail")
	}
	if _, err := catalog.Advance(state, AdvanceInput{ServerMS: 1, Unlocked: true, Salt: "0123456789ABCDEF"}); err == nil {
		t.Fatal("an uppercase salt must fail")
	}
	if _, err := catalog.Advance(state, AdvanceInput{ServerMS: 1, Unlocked: true, Salt: "0123456789abcdef"}); err != nil || state.SaltHex == nil || *state.TickAnchorWallMS != 1 {
		t.Fatalf("initializing advance: %v %+v", err, state)
	}
	if _, err := catalog.Advance(state, AdvanceInput{ServerMS: 2, Unlocked: true, Salt: "0123456789abcdef"}); err == nil {
		t.Fatal("a second salt must fail")
	}
}

// timelineEvent is one Founder command in AC6's property: a garden command, a
// host-level purchase (spend_fiscal_credit), or any other Founder command.
type timelineEvent struct {
	at   int64
	kind string
}

// replayTimeline advances the garden before each command whose kind is in
// triggers, applying level purchases after their own pre-step, then closes
// with a garden command at end.
func replayTimeline(t *testing.T, catalog *Catalog, events []timelineEvent, end int64, triggers map[string]bool) *State {
	t.Helper()
	state := matureState(t, Plot{Row: 0, Col: 0, SpeciesID: "strain_a", AgeTicks: 3, MaturedEffectPPM: ppmPointer(1_000_000)},
		Plot{Row: 1, Col: 1, SpeciesID: "strain_b", AgeTicks: 4, MaturedEffectPPM: ppmPointer(1_000_000)})
	level := int64(0)
	for _, event := range append(append([]timelineEvent{}, events...), timelineEvent{at: end, kind: "garden"}) {
		if triggers[event.kind] {
			if _, err := catalog.Advance(state, AdvanceInput{ServerMS: event.at, HostLevel: level, Unlocked: true}); err != nil {
				t.Fatal(err)
			}
		}
		if event.kind == "level_up" {
			level = min(level+1, 8)
		}
	}
	return state
}

// TestGardenTriggerSetIsSufficient is AC6: advancing before every Founder
// command equals advancing only before the SG3 set (garden commands and
// spend_fiscal_credit), because no other command changes an advance input.
func TestGardenTriggerSetIsSufficient(t *testing.T) {
	catalog := scenarioCatalog(t, []fixtureOp{setOp(150_000, "recipes", 0, "chance_ppm"), setOp(150_000, "recipes", 1, "chance_ppm"), setOp(80_000, "recipes", 2, "chance_ppm")})
	random := rand.New(rand.NewSource(11))
	every := map[string]bool{"garden": true, "level_up": true, "other": true}
	set := map[string]bool{"garden": true, "level_up": true}
	for trial := 0; trial < 150; trial++ {
		events := []timelineEvent{}
		at := int64(1_000_000)
		for index := 0; index < 12; index++ {
			at += int64(random.Intn(3_600_000)) + 1
			events = append(events, timelineEvent{at: at, kind: []string{"garden", "level_up", "other", "other"}[random.Intn(4)]})
		}
		end := at + int64(random.Intn(3_600_000)) + 1
		if !replayTimeline(t, catalog, events, end, every).Equal(replayTimeline(t, catalog, events, end, set)) {
			t.Fatalf("trial %d: the SG3 trigger set diverged from advancing before every command", trial)
		}
	}
}

// TestGardenTriggerSetWithoutSpendDiverges is AC6's failing case: dropping
// spend_fiscal_credit from the set diverges once a mid-interval level
// purchase widens the grid.
func TestGardenTriggerSetWithoutSpendDiverges(t *testing.T) {
	catalog := scenarioCatalog(t, []fixtureOp{setOp(400_000, "recipes", 0, "chance_ppm"), setOp(400_000, "recipes", 1, "chance_ppm")})
	events := []timelineEvent{{at: 1_000_000 + 2*3_600_000, kind: "level_up"}}
	end := int64(1_000_000 + 4*3_600_000)
	withSpend := replayTimeline(t, catalog, events, end, map[string]bool{"garden": true, "level_up": true})
	withoutSpend := replayTimeline(t, catalog, events, end, map[string]bool{"garden": true})
	if withSpend.Equal(withoutSpend) {
		t.Fatal("removing spend_fiscal_credit from the trigger set must diverge when a level purchase widens the grid")
	}
}

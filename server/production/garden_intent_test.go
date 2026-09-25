package production

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"cloud-clicker/server/economy"
	"cloud-clicker/server/garden"
	"cloud-clicker/server/save"
)

const gardenCorpusPath = "../../testdata/replay/garden-v1.json"

type gardenRunner struct {
	t        *testing.T
	catalogs CatalogBundle
	bundle   string
	state    *save.State
	revision int64
	cases    []reputationCorpusCase
}

// command runs one Founder command through the live resolved-input builder
// and the shared ApplyFounderLogged, appending a corpus case.
func (runner *gardenRunner) command(name, kind, body string, now time.Time) FounderLoggedTransition {
	t := runner.t
	t.Helper()
	intentID := fmt.Sprintf("01986666-6d%02d-7000-8000-000000000001", len(runner.cases)+1)
	request, err := ParseIntent([]byte(fmt.Sprintf(`{"intent_id":%q,"kind":%q,"expected_revision":%d,%s}`, intentID, kind, runner.revision, body)))
	if err != nil {
		t.Fatal(err)
	}
	pre := mustEncodeState(t, runner.state)
	command := save.FounderReplayCommand{IntentID: intentID, FounderStreamID: "01986666-5d00-4000-8000-000000000001",
		FounderID: petFixtureFounderID, Revision: runner.revision, FounderLogSeq: runner.revision, ServerTSMS: now.UnixMilli()}
	var resolved any = founderInvalidResolved{Kind: "invalid", Detail: request.InvalidDetail}
	if request.InvalidDetail == "" {
		switch {
		case kind == IntentSpendFiscalCredit:
			salt, saltErr := liveGardenSalt(runner.catalogs, runner.state)
			if saltErr != nil {
				t.Fatal(saltErr)
			}
			cost, costErr := resolvedFiscalCost(runner.catalogs.Fiscal, fiscalStateFromSave(runner.state), request.FiscalTarget)
			if costErr != nil {
				cost = 0
			}
			resolved = founderFiscalSpendResolved{Kind: kind, Target: fiscalTargetFromRequest(request), ResolvedCost: cost, GardenSaltHex: salt}
		default:
			live, liveErr := liveGardenResolved(runner.catalogs, runner.state, kind, now.UnixMilli())
			if liveErr != nil {
				t.Fatal(liveErr)
			}
			resolved = live
		}
	}
	inputs, err := save.MarshalFounderReplayInputs(command, resolved)
	if err != nil {
		t.Fatal(err)
	}
	transition, err := ApplyFounderLogged(runner.state, request.CanonicalPayload, runner.catalogs, inputs)
	if err != nil {
		t.Fatalf("%s: %v", name, err)
	}
	if transition.Outcome == save.IntentApplied {
		runner.revision++
	}
	post := mustEncodeState(t, runner.state)
	runner.cases = append(runner.cases, reputationCorpusCase{Name: name, Bundle: runner.bundle, StateVersion: save.VersionForState(runner.state),
		PreState: pre, CanonicalPayload: request.CanonicalPayload, ReplayInputs: inputs, Outcome: string(transition.Outcome),
		ReceiptJSON: canonicalFixtureJSON(t, transition.Receipt), EventsJSON: canonicalFixtureValue(t, fixtureEvents(transition.Events)), PostStateJSON: canonicalFixtureJSON(t, post)})
	return transition
}

func (runner *gardenRunner) reject(name, kind, body string, now time.Time, category, detail string) {
	t := runner.t
	t.Helper()
	before := mustEncodeState(t, runner.state)
	transition := runner.command(name, kind, body, now)
	gotCategory, gotDetail := rejectionOf(t, transition.Receipt)
	if gotCategory != category || gotDetail != detail || len(transition.Events) != 0 || !bytes.Equal(before, mustEncodeState(t, runner.state)) {
		t.Fatalf("%s: rejection %s/%s events=%d receipt=%s", name, gotCategory, gotDetail, len(transition.Events), transition.Receipt)
	}
}

func (runner *gardenRunner) apply(name, kind, body string, now time.Time) FounderLoggedTransition {
	t := runner.t
	t.Helper()
	transition := runner.command(name, kind, body, now)
	if transition.Outcome != save.IntentApplied {
		t.Fatalf("%s: not applied: %s", name, transition.Receipt)
	}
	return transition
}

func eventKinds(events []save.EventWrite) []save.EventKind {
	kinds := []save.EventKind{}
	for _, event := range events {
		kinds = append(kinds, event.Kind)
	}
	return kinds
}

func pinGardenSalt(t *testing.T) {
	t.Helper()
	previous := gardenSaltDrawer
	gardenSaltDrawer = func() (string, error) { return "5eed5eed5eed5eed", nil }
	t.Cleanup(func() { gardenSaltDrawer = previous })
}

func buildGardenCorpus(t *testing.T) cosmeticCorpus {
	t.Helper()
	pinGardenSalt(t)
	shop := cosmeticsContentBundle(t)
	grown := gardenContentBundle(t)
	now := time.Date(2026, 9, 25, 12, 0, 0, 0, time.UTC)
	plant := func(row, col int, species string) string {
		return fmt.Sprintf(`"row":%d,"col":%d,"species_id":%q`, row, col, species)
	}
	cell := func(row, col int) string { return fmt.Sprintf(`"row":%d,"col":%d`, row, col) }

	inactive := &gardenRunner{t: t, catalogs: shop, bundle: "shop", state: reputationFounderState(t, shop, 24, now, 1), revision: 1}
	inactive.reject("rejects-inactive-below-v25", IntentGardenPlant, plant(0, 0, "strain_a"), now, "not_eligible", garden.DetailInactive)

	state := reputationFounderState(t, grown, 25, now, 1)
	state.FiscalCredit = 60
	runner := &gardenRunner{t: t, catalogs: grown, bundle: "grown", state: state, revision: 1}
	// AC12: a client-authored tick/time/outcome/salt field is terminal invalid.
	runner.reject("rejects-client-tick-seq", IntentGardenPlant, plant(0, 0, "strain_a")+`,"tick_seq":3`, now, "invalid", "garden_plant.fields")
	runner.reject("rejects-client-salt", IntentGardenSetSubstrate, `"substrate_id":"mainframe","salt_hex":"0123456789abcdef"`, now, "invalid", "garden_set_substrate.fields")
	runner.reject("rejects-row-out-of-range", IntentGardenPlant, plant(6, 0, "strain_a"), now, "invalid", "row")
	runner.reject("rejects-locked", IntentGardenPlant, plant(0, 0, "strain_a"), now, "not_eligible", garden.DetailUnlockRequired)
	if state.ServerGarden.SaltHex != nil {
		t.Fatal("a locked advance drew the hidden salt")
	}
	unlock := runner.apply("buys-the-unlock", IntentSpendFiscalCredit, `"target":{"kind":"unlock","unlock_id":"minigame.server_garden"}`, now)
	if state.ServerGarden.SaltHex != nil || state.ServerGarden.TickAnchorWallMS != nil || !strings.Contains(string(unlock.Receipt), `"garden_advance"`) {
		t.Fatalf("the unlock's pre-step ran on the pre-purchase (locked) garden: %+v %s", state.ServerGarden, unlock.Receipt)
	}
	start := now.Add(time.Minute)
	planted := runner.apply("plants-and-anchors", IntentGardenPlant, plant(0, 0, "strain_a"), start)
	if state.ServerGarden.SaltHex == nil || *state.ServerGarden.TickAnchorWallMS != start.UnixMilli() || !strings.HasSuffix(fmt.Sprint(eventKinds(planted.Events)), "garden_planted.v1]") ||
		strings.Contains(fmt.Sprint(eventKinds(planted.Events)), "garden_advanced") {
		t.Fatalf("first unlocked command must draw the salt and anchor: %+v %v", state.ServerGarden, eventKinds(planted.Events))
	}
	runner.reject("rejects-occupied", IntentGardenPlant, plant(0, 0, "strain_b"), start, "not_eligible", garden.DetailPlotOccupied)
	runner.reject("rejects-dormant", IntentGardenPlant, plant(2, 0, "strain_b"), start, "not_eligible", garden.DetailPlotDormant)
	runner.reject("rejects-unknown-species", IntentGardenPlant, plant(0, 1, "strain_zz"), start, "unknown_id", garden.DetailUnknownSpecies)
	runner.reject("rejects-uncollected-seed", IntentGardenPlant, plant(0, 1, "strain_c"), start, "not_eligible", garden.DetailSeedNotCollected)
	runner.reject("rejects-uproot-empty", IntentGardenUproot, cell(1, 1), start, "unknown_id", garden.DetailUnknownPlot)
	runner.reject("rejects-unknown-substrate", IntentGardenSetSubstrate, `"substrate_id":"quantum"`, start, "unknown_id", garden.DetailUnknownSubstrate)
	runner.reject("rejects-unchanged-substrate", IntentGardenSetSubstrate, `"substrate_id":"bare_metal"`, start, "not_eligible", garden.DetailUnchanged)
	runner.apply("plants-second-starter", IntentGardenPlant, plant(1, 1, "strain_b"), start)
	grown1h := start.Add(time.Hour + 7*time.Second)
	switched := runner.apply("switches-substrate-after-growth", IntentGardenSetSubstrate, `"substrate_id":"containerized"`, grown1h)
	if kinds := fmt.Sprint(eventKinds(switched.Events)); !strings.HasPrefix(kinds, "[") || !strings.Contains(kinds, "garden_advanced.v1 garden_substrate_set.v1") {
		t.Fatalf("an hour of growth must precede the substrate event: %s", kinds)
	}
	runner.reject("rejects-substrate-lockout", IntentGardenSetSubstrate, `"substrate_id":"mainframe"`, grown1h.Add(5*time.Minute), "not_eligible", garden.DetailLockout)
	// The level-up widens the grid only after its own pre-step (AC6 order).
	levelled := runner.apply("buys-a-host-level", IntentSpendFiscalCredit, `"target":{"kind":"generator_level","generator_id":"generator.beige_tower","levels":1}`, grown1h.Add(20*time.Minute))
	if !strings.Contains(string(levelled.Receipt), `"garden_advance"`) {
		t.Fatalf("spend_fiscal_credit must carry the garden pre-step: %s", levelled.Receipt)
	}
	runner.apply("plants-in-the-widened-row", IntentGardenPlant, plant(0, 2, "strain_a"), grown1h.Add(21*time.Minute))
	runner.apply("uproots", IntentGardenUproot, cell(0, 2), grown1h.Add(22*time.Minute))
	away := grown1h.Add(31 * time.Hour)
	caught := runner.apply("catches-up-past-the-cap", IntentGardenSetSubstrate, `"substrate_id":"mainframe"`, away)
	var catchReceipt struct {
		GardenAdvance garden.Advance `json:"garden_advance"`
	}
	if err := json.Unmarshal(caught.Receipt, &catchReceipt); err != nil || catchReceipt.GardenAdvance.CatchupForfeitedMS <= 0 || catchReceipt.GardenAdvance.CatchupReasonKey == nil {
		t.Fatalf("a 31h absence must report visible forfeiture: %s", caught.Receipt)
	}

	cases := append(append([]reputationCorpusCase{}, inactive.cases...), runner.cases...)
	exitCases := []reputationExitCase{makeReputationPlanExitCase(t, "exit-activates-founder-v25", shop, grown, 24, 4, nil, now)}
	return cosmeticCorpus{Version: 1, Cases: cases, ExitCases: exitCases, Bundles: map[string]reputationCorpusBundle{
		"shop":  {ConstantsHash: shop.ConstantsHash, Artifacts: stringArtifacts(shop.Artifacts)},
		"grown": {ConstantsHash: grown.ConstantsHash, Artifacts: stringArtifacts(grown.Artifacts)},
	}}
}

// TestGardenCorpus is AC7 + AC9's Founder half: every SG5 detail through the
// shared ApplyFounderLogged, pinned as the corpus the TS replay must
// byte-match. GARDEN_UPDATE_FIXTURE=1 regenerates it.
func TestGardenCorpus(t *testing.T) {
	encoded, err := json.MarshalIndent(buildGardenCorpus(t), "", " ")
	if err != nil {
		t.Fatal(err)
	}
	encoded = append(encoded, '\n')
	if os.Getenv("GARDEN_UPDATE_FIXTURE") == "1" {
		if err := os.WriteFile(gardenCorpusPath, encoded, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	pinned, err := os.ReadFile(gardenCorpusPath)
	if err != nil {
		t.Fatalf("pinned corpus: %v", err)
	}
	if !bytes.Equal(pinned, encoded) {
		t.Fatal("garden corpus drifted from the Go transition; regenerate with GARDEN_UPDATE_FIXTURE=1 and review")
	}
}

// AC7's failing case: replay rejects a recorded command whose resolved advance
// no longer matches the recomputation, and a rejected command commits nothing
// (the pre-step rolls back with it).
func TestGardenReplayRejectsTamperedAdvance(t *testing.T) {
	corpus := buildGardenCorpus(t)
	grown := gardenContentBundle(t)
	for _, testCase := range corpus.Cases {
		if testCase.Name != "switches-substrate-after-growth" {
			continue
		}
		var inputs map[string]any
		if err := json.Unmarshal(testCase.ReplayInputs, &inputs); err != nil {
			t.Fatal(err)
		}
		inputs["resolved"].(map[string]any)["advance"].(map[string]any)["ticks_applied"] = 1
		tampered, _ := json.Marshal(inputs)
		state, err := save.RestoreState(testCase.PreState, testCase.StateVersion, grown.Economy, economy.ScopeFounder, time.Time{})
		if err != nil {
			t.Fatal(err)
		}
		before := mustEncodeState(t, state)
		if _, err := ApplyFounderLogged(state, testCase.CanonicalPayload, grown, tampered); !errors.Is(err, ErrInvalidReplayInputs) || !bytes.Equal(before, mustEncodeState(t, state)) {
			t.Fatalf("a tampered advance replayed: err=%v", err)
		}
		return
	}
	t.Fatal("corpus case missing")
}

// checkGardenTransition's law: only the trigger set may change the garden.
func TestTransitionLayerRejectsGardenMutatingArms(t *testing.T) {
	grown := gardenContentBundle(t)
	fresh := garden.NewState(grown.Garden)
	changed := fresh.Clone()
	changed.TickSeq = 1
	if checkGardenTransition(fresh, changed, false) == nil || checkGardenTransition(nil, changed, false) == nil {
		t.Fatal("illegal garden transitions accepted")
	}
	if checkGardenTransition(fresh, changed, true) != nil || checkGardenTransition(nil, fresh, false) != nil || checkGardenTransition(fresh, fresh.Clone(), false) != nil {
		t.Fatal("legal garden transitions rejected")
	}
}

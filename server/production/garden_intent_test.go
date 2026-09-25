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

	"cloud-clicker/server/decimal"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/garden"
	"cloud-clicker/server/meters"
	"cloud-clicker/server/minigame"
	"cloud-clicker/server/save"
)

const gardenCorpusPath = "../../testdata/replay/garden-v1.json"

type gardenRunner struct {
	t             *testing.T
	catalogs      CatalogBundle
	bundle        string
	state         *save.State
	revision      int64
	cases         []reputationCorpusCase
	creditHarvest bool
}

type gardenCorpus struct {
	Version      int                               `json:"version"`
	Bundles      map[string]reputationCorpusBundle `json:"bundles"`
	Cases        []reputationCorpusCase            `json:"cases"`
	ExitCases    []reputationExitCase              `json:"exit_cases"`
	CompanyCases []reputationCorpusCase            `json:"company_cases"`
}

// gardenAttendance is the frozen Founder attendance sample a harvest carries.
func gardenAttendance(catalogs CatalogBundle, state *save.State) FounderAttendanceSample {
	return FounderAttendanceSample{CompanyStreamID: "01986666-1900-7000-8000-000000000901", RunSeq: 1, CompanyRevision: 1,
		CompanyConstantsHash: catalogs.ConstantsHash, CompletedAttendedMS: state.AgeMS, CurrentRunPartialAttendedMS: 400_000,
		EffectiveFounderAttendedMS: state.AgeMS + 400_000}
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
		case kind == IntentGardenHarvest:
			resolvedKind := IntentGardenHarvest
			if runner.creditHarvest {
				resolvedKind = gardenHarvestCreditedKind
			}
			live, liveErr := liveGardenHarvestResolved(runner.catalogs, runner.state, resolvedKind, now.UnixMilli(), gardenAttendance(runner.catalogs, runner.state))
			if liveErr != nil {
				t.Fatal(liveErr)
			}
			resolved = live
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

func buildGardenCorpus(t *testing.T) gardenCorpus {
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

	// SG6 Founder side: a Founder-only harvest must reject; the coordinator's
	// credited arm must apply, remove the plants, and collect new seeds.
	runner.reject("rejects-harvest-empty-plot", IntentGardenHarvest, `"plots":[{"row":5,"col":5}]`, away.Add(time.Minute), "unknown_id", garden.DetailUnknownPlot)
	mature := []string{}
	for _, plot := range state.ServerGarden.Plots {
		if plot.Mature() {
			mature = append(mature, fmt.Sprintf(`{"row":%d,"col":%d}`, plot.Row, plot.Col))
		}
	}
	if len(mature) < 2 {
		t.Fatalf("the corpus garden needs mature plants to harvest: %+v", state.ServerGarden.Plots)
	}
	harvestBody := `"plots":[` + strings.Join(mature, ",") + `]`
	runner.creditHarvest = true
	harvested := runner.apply("applies-credited-harvest", IntentGardenHarvest, harvestBody, away.Add(2*time.Minute))
	var harvestReceipt founderGardenHarvestReceipt
	if err := json.Unmarshal(harvested.Receipt, &harvestReceipt); err != nil || harvestReceipt.Harvest.TotalUnits <= 0 || len(harvestReceipt.Harvest.SeedsDiscovered) == 0 {
		t.Fatalf("credited harvest receipt: %s", harvested.Receipt)
	}
	runner.creditHarvest = false

	cases := append(append([]reputationCorpusCase{}, inactive.cases...), runner.cases...)
	exitCases := []reputationExitCase{makeReputationPlanExitCase(t, "exit-activates-founder-v25", shop, grown, 24, 4, nil, now)}
	return gardenCorpus{Version: 1, Cases: cases, ExitCases: exitCases, CompanyCases: buildGardenCompanyCases(t, grown, now), Bundles: map[string]reputationCorpusBundle{
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

// gardenCompanyState is a pinned Company run under the garden bundle.
func gardenCompanyState(t *testing.T, bundle CatalogBundle, now time.Time) *save.State {
	t.Helper()
	company := replayFixtureState(t, bundle.Economy, now)
	meterState, err := meters.NewRunState(bundle.Meters, 0)
	if err != nil {
		t.Fatal(err)
	}
	company.WireVersion, company.MeterBands = 18, nil
	company.MeterValues, company.MeterDecayRemainders, company.MeterInputRemainders = meterState.Values, meterState.DecayRemainders, meterState.InputRemainders
	company.AchievementsEarnedRun = map[string]bool{}
	company.RunStartedAt = now
	if _, err := initializeActivePlayState(company, bundle.Opportunities, petFixtureFounderID); err != nil {
		t.Fatal(err)
	}
	return company
}

// buildGardenCompanyCases pins the Company run-log arm (SG6): a normal
// credit, the per-send cap, the exhausted daily sends (forfeit with the reason
// key), and a zero harvest (no window touched), through the real ApplyLogged.
func buildGardenCompanyCases(t *testing.T, grown CatalogBundle, now time.Time) []reputationCorpusCase {
	t.Helper()
	policy := grown.Garden.Payout
	base := mustEncodeState(t, gardenCompanyState(t, grown, now))
	cases := []reputationCorpusCase{}
	build := func(name string, selected, quotaBefore, remainderBefore int64) {
		state := replayFixtureStateFromEncoded(t, grown, base)
		intentID := fmt.Sprintf("01986666-6e%02d-7000-8000-000000000001", len(cases)+1)
		harvestHash, err := garden.HarvestHash(intentID, []garden.HarvestedPlot{{Row: 0, Col: 0, SpeciesID: "strain_c", Units: selected}}, selected)
		if err != nil {
			t.Fatal(err)
		}
		var faucet *minigameFaucetWire
		credit := decimal.Zero
		forfeited := int64(0)
		var reason *string
		if selected > 0 {
			conversion, convertErr := minigame.ConvertPayout(selected, 0, policy.ConversionPPM, remainderBefore)
			if convertErr != nil {
				t.Fatal(convertErr)
			}
			credited, quotaAfter := int64(0), quotaBefore
			if quotaBefore < policy.SendsPerDay {
				quotaAfter++
				credited = min(conversion.ConvertedUnits, policy.PerSendCap)
			}
			forfeited = conversion.ConvertedUnits - credited
			capReason := ""
			if forfeited > 0 {
				capReason = policy.CapReasonKey
				reason = &capReason
			}
			faucet = &minigameFaucetWire{AttendedDay: 0, QuotaBefore: quotaBefore, QuotaAfter: quotaAfter, RemainderBeforePPM: remainderBefore,
				RemainderAfterPPM: conversion.ConversionRemainderPPM, ReducedScore: conversion.ReducedScore, ConvertedUnits: conversion.ConvertedUnits,
				CreditedUnits: credited, ForfeitedUnits: forfeited, CapReasonKey: capReason}
			credit = decimal.FromString(fmt.Sprint(credited))
		}
		probe := replayFixtureStateFromEncoded(t, grown, base)
		ledger, err := probe.Ledger.ApplyAccrual(economy.Transaction{Entries: []economy.Entry{{ResourceID: policy.CreditedResourceID, Delta: credit}}})
		if err != nil {
			t.Fatal(err)
		}
		creditedDelta, err := minigameCreditedDelta(ledger, policy.CreditedResourceID)
		if err != nil {
			t.Fatal(err)
		}
		command := save.ReplayCommand{IntentID: intentID, CompanyStreamID: "01986666-1900-7000-8000-000000000901", FounderID: petFixtureFounderID,
			Revision: 1, RunSeq: state.RunSeq, RunLogSeq: 1}
		resolved := gardenCompanyResolved{Kind: gardenCompanyCreditKind, IntentID: intentID, HarvestHash: harvestHash, PolicyHash: gardenPolicyHash(policy),
			SelectedUnits: selected, Faucet: faucet, Credited: creditedDelta, ForfeitedUnits: forfeited, CapReasonKey: reason, FaucetApplied: selected > 0,
			FounderLog:      minigameLogCoordinate{StreamID: "01986666-5d00-4000-8000-000000000001", Revision: 7, Sequence: 6},
			CompanyRevision: 2, FounderRevision: 7}
		inputs, err := json.Marshal(replayInputsWire{Version: save.ReplayInputsVersion, Command: command, EvaluatedAtMS: now.UnixMilli(), EvaluationMode: ModeOnline, Resolved: mustJSON(resolved)})
		if err != nil {
			t.Fatal(err)
		}
		payload, err := normalizeReplayJSON(mustJSON(gardenHarvestCreditPayload{Kind: gardenCompanyCreditKind, IntentID: intentID, HarvestHash: harvestHash}))
		if err != nil {
			t.Fatal(err)
		}
		pre := mustEncodeState(t, state)
		transition, err := ApplyLogged(state, payload, grown, inputs)
		if err != nil || transition.Outcome != save.IntentApplied {
			t.Fatalf("%s: %v %s", name, err, transition.Receipt)
		}
		cases = append(cases, reputationCorpusCase{Name: name, Bundle: "grown", StateVersion: save.VersionForState(state), PreState: pre,
			CanonicalPayload: payload, ReplayInputs: inputs, Outcome: string(transition.Outcome), ReceiptJSON: canonicalFixtureJSON(t, transition.Receipt),
			EventsJSON: canonicalFixtureValue(t, fixtureEvents(transition.Events)), PostStateJSON: canonicalFixtureJSON(t, mustEncodeState(t, state))})
	}
	build("credits-a-harvest", 21, 0, 0)
	build("caps-a-single-send", 500, 1, 250_000)
	build("forfeits-past-the-daily-sends", 40, policy.SendsPerDay, 0)
	build("zero-harvest-touches-no-window", 0, 0, 0)
	return cases
}

// The Founder-only harvest arm refuses to apply: applied harvests commit only
// through the SG-P2 coordinator, so a Founder-only credit can never happen.
func TestFounderOnlyHarvestRefusesToApply(t *testing.T) {
	pinGardenSalt(t)
	grown := gardenContentBundle(t)
	now := time.Date(2026, 9, 25, 12, 0, 0, 0, time.UTC)
	state := reputationFounderState(t, grown, 25, now, 1)
	state.FiscalUnlocks[grown.Garden.UnlockID] = true
	effect := int64(1_000_000)
	salt, anchor := "0123456789abcdef", now.UnixMilli()
	state.ServerGarden.SaltHex, state.ServerGarden.TickAnchorWallMS = &salt, &anchor
	state.ServerGarden.Plots = []garden.Plot{{Row: 0, Col: 0, SpeciesID: "strain_a", AgeTicks: 3, MaturedEffectPPM: &effect}}
	runner := &gardenRunner{t: t, catalogs: grown, bundle: "grown", state: state, revision: 1}
	request, err := ParseIntent([]byte(`{"intent_id":"01986666-6f01-7000-8000-000000000001","kind":"garden_harvest","expected_revision":1,"plots":[{"row":0,"col":0}]}`))
	if err != nil {
		t.Fatal(err)
	}
	resolved, err := liveGardenHarvestResolved(grown, runner.state, IntentGardenHarvest, now.UnixMilli(), gardenAttendance(grown, runner.state))
	if err != nil {
		t.Fatal(err)
	}
	command := save.FounderReplayCommand{IntentID: request.IntentID, FounderStreamID: "01986666-5d00-4000-8000-000000000001", FounderID: petFixtureFounderID,
		Revision: 1, FounderLogSeq: 1, ServerTSMS: now.UnixMilli()}
	inputs, err := save.MarshalFounderReplayInputs(command, resolved)
	if err != nil {
		t.Fatal(err)
	}
	before := mustEncodeState(t, state)
	if _, err := ApplyFounderLogged(state, request.CanonicalPayload, grown, inputs); !errors.Is(err, errGardenHarvestMustCredit) || !bytes.Equal(before, mustEncodeState(t, state)) {
		t.Fatalf("a Founder-only harvest applied or mutated state: %v", err)
	}
}

// TestCompanyArmEnvelopesParse is the finding fixed alongside SG6:
// parseReplayInputs rejected the minigame resolution's Company envelope, so
// ApplyLogged (and VerifyReplayRun) could never reach
// applyCompanyMinigameResolution. Both Company arms' envelopes must parse.
func TestCompanyArmEnvelopesParse(t *testing.T) {
	now := time.Date(2026, 8, 5, 17, 0, 0, 0, time.UTC)
	_, _, companyCase, _ := makeMinigameResolutionReplayFixture(t, now)
	if _, err := parseReplayInputs(companyCase.ReplayInputs); err != nil {
		t.Fatalf("the minigame resolution Company envelope must parse: %v", err)
	}
	corpus := buildGardenCorpus(t)
	for _, testCase := range corpus.CompanyCases {
		if _, err := parseReplayInputs(testCase.ReplayInputs); err != nil {
			t.Fatalf("%s: the garden credit envelope must parse: %v", testCase.Name, err)
		}
	}
}

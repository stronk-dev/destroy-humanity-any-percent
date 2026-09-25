package production

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"cloud-clicker/server/copykeys"
	"cloud-clicker/server/curriculum"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/meters"
	"cloud-clicker/server/pet"
	"cloud-clicker/server/reputation"
	"cloud-clicker/server/save"
)

const reputationCorpusPath = "../../testdata/replay/reputation-tree-v1.json"

type reputationCorpusBundle struct {
	ConstantsHash string            `json:"constants_hash"`
	Artifacts     map[string]string `json:"artifacts"`
}

type reputationCorpusCase struct {
	Name             string          `json:"name"`
	Bundle           string          `json:"bundle"`
	StateVersion     int             `json:"state_version"`
	PreState         json.RawMessage `json:"pre_state"`
	CanonicalPayload json.RawMessage `json:"canonical_payload"`
	ReplayInputs     json.RawMessage `json:"replay_inputs"`
	Outcome          string          `json:"outcome"`
	ReceiptJSON      string          `json:"receipt_json"`
	EventsJSON       string          `json:"events_json"`
	PostStateJSON    string          `json:"post_state_json"`
}

type reputationCorpus struct {
	Version int                               `json:"version"`
	Bundles map[string]reputationCorpusBundle `json:"bundles"`
	Cases   []reputationCorpusCase            `json:"cases"`
	// Exit is the R4/R7 terminal witness: a scripted-first burnout Exit on a
	// tree bundle whose Founder owns starter nodes (AC8, run_started v2).
	Exit crossRuntimeActiveExit `json:"exit"`
}

// reputationFounderState is an encodable Founder at the bundle's floor with
// every v17–v21 collection initialized, plus optional v22 tree state.
func reputationFounderState(t *testing.T, catalogs CatalogBundle, version int, now time.Time, level int64) *save.State {
	t.Helper()
	state := replayFounderFixtureState(t, catalogs, now)
	state.WireVersion = version
	state.MinigameRatings = map[string]save.MinigameRatingState{"pitch": {Elo: 1000, SeasonMember: "s1"}}
	state.MinigameOfflineQuality = map[string]save.MinigameOfflineQualityState{"pitch": {GradePPM: 200_000}}
	state.Pets = map[string]pet.CareState{}
	state.FiscalPeriodOpenedWallMS, state.FiscalGeneratorLevels, state.FiscalUnlocks = now.UnixMilli(), map[string]int64{}, map[string]bool{}
	for _, row := range catalogs.Fiscal.GeneratorLevelRows() {
		state.FiscalGeneratorLevels[row.GeneratorID] = 0
	}
	state.Soul, state.SoulExhaustedSourceIDs = catalogs.Soul.Policy.Initial, []string{}
	state.ReputationLevel = level
	if version >= 22 {
		state.ReputationNodesOwned = []string{}
	}
	if err := catalogs.ValidateFoundationState(state); err != nil {
		t.Fatalf("reputation Founder fixture invalid: %v", err)
	}
	return state
}

type reputationRunner struct {
	t        *testing.T
	catalogs CatalogBundle
	bundle   string
	state    *save.State
	revision int64
	now      time.Time
	cases    []reputationCorpusCase
}

// purchase drives one logged Founder command through the live resolver and
// ApplyFounderLogged, records the corpus case, and returns the transition.
func (runner *reputationRunner) purchase(name, body string, serverTS time.Time) FounderLoggedTransition {
	t := runner.t
	t.Helper()
	intentID := fmt.Sprintf("01986666-5a%02d-7000-8000-000000000001", len(runner.cases)+1)
	request, err := ParseIntent([]byte(fmt.Sprintf(`{"intent_id":%q,"kind":"purchase_reputation_node","expected_revision":%d,%s}`, intentID, runner.revision, body)))
	if err != nil {
		t.Fatal(err)
	}
	pre := mustEncodeState(t, runner.state)
	command := save.FounderReplayCommand{IntentID: intentID, FounderStreamID: "01986666-5c00-4000-8000-000000000001",
		FounderID: "01986666-5d00-7000-8000-000000000001", Revision: runner.revision, FounderLogSeq: runner.revision, ServerTSMS: serverTS.UnixMilli()}
	var resolved any = founderInvalidResolved{Kind: "invalid", Detail: request.InvalidDetail}
	if request.InvalidDetail == "" {
		resolved, err = resolveReputationPurchase(runner.catalogs, runner.state, request.ReputationNodeID)
		if err != nil {
			t.Fatal(err)
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
	events := fixtureEvents(transition.Events)
	runner.cases = append(runner.cases, reputationCorpusCase{Name: name, Bundle: runner.bundle, StateVersion: save.VersionForState(runner.state),
		PreState: pre, CanonicalPayload: request.CanonicalPayload, ReplayInputs: inputs, Outcome: string(transition.Outcome),
		ReceiptJSON: canonicalFixtureJSON(t, transition.Receipt), EventsJSON: canonicalFixtureValue(t, events), PostStateJSON: canonicalFixtureJSON(t, post)})
	return transition
}

func rejectionOf(t *testing.T, receipt json.RawMessage) (string, string) {
	t.Helper()
	var value struct {
		Outcome   string `json:"outcome"`
		Rejection struct {
			Category string `json:"category"`
			Detail   string `json:"detail"`
		} `json:"rejection"`
	}
	if err := json.Unmarshal(receipt, &value); err != nil || value.Outcome != "rejected" {
		t.Fatalf("not a rejection: %s", receipt)
	}
	return value.Rejection.Category, value.Rejection.Detail
}

func buildReputationCorpus(t *testing.T) reputationCorpus {
	t.Helper()
	tree := reputationContentBundle(t)
	live := activeContentBundle(t)
	now := time.Date(2026, 9, 25, 12, 0, 0, 0, time.UTC)

	// AC3: every R5 rejection row is recorded with no mutation and no event.
	rejections := &reputationRunner{t: t, catalogs: tree, bundle: "tree", state: reputationFounderState(t, tree, 22, now, 3), revision: 1, now: now}
	inactive := &reputationRunner{t: t, catalogs: live, bundle: "live", state: reputationFounderState(t, live, 21, now, 3), revision: 1, now: now}
	for _, row := range []struct {
		runner           *reputationRunner
		name, body       string
		category, detail string
	}{
		{inactive, "rejects-inactive-tree", `"node_id":"reputation.unlock.p05"`, "not_eligible", "reputation_tree_inactive"},
		{rejections, "rejects-invalid-fields", `"node_id":"reputation.unlock.p05","extra":1`, "invalid", "purchase_reputation_node.fields"},
		{rejections, "rejects-invalid-node-id", `"node_id":"Not Mechanical"`, "invalid", "node_id"},
		{rejections, "rejects-unknown-node", `"node_id":"reputation.unlock.p99"`, "unknown_id", "reputation.unlock.p99"},
		{rejections, "rejects-missing-prerequisite", `"node_id":"reputation.starter.generated_beige_tower"`, "not_eligible", "requires"},
		{rejections, "rejects-ladder-prerequisite", `"node_id":"reputation.unlock.p25"`, "not_eligible", "requires"},
	} {
		before := mustEncodeState(t, row.runner.state)
		transition := row.runner.purchase(row.name, row.body, now)
		category, detail := rejectionOf(t, transition.Receipt)
		if category != row.category || detail != row.detail || len(transition.Events) != 0 || !bytes.Equal(before, mustEncodeState(t, row.runner.state)) {
			t.Fatalf("%s: rejection %s/%s events=%d mutated=%t", row.name, category, detail, len(transition.Events), !bytes.Equal(before, mustEncodeState(t, row.runner.state)))
		}
	}
	// Owned and unaffordable need prior state: buy p05 (cost 1), then retry it,
	// then attempt cash_small (cost 2) with only 2 available -> applied, then
	// generated_beige_tower (cost 3) with 0 available -> unaffordable.
	rejections.purchase("applies-first-unlock", `"node_id":"reputation.unlock.p05"`, now)
	if category, detail := rejectionOf(t, rejections.purchase("rejects-owned", `"node_id":"reputation.unlock.p05"`, now).Receipt); category != "not_eligible" || detail != "owned" {
		t.Fatalf("owned rejection %s/%s", category, detail)
	}
	rejections.purchase("applies-starter-with-exact-budget", `"node_id":"reputation.starter.cash_small"`, now)
	if category, detail := rejectionOf(t, rejections.purchase("rejects-unaffordable-starter", `"node_id":"reputation.starter.generated_beige_tower"`, now).Receipt); category != "unaffordable" || detail != "reputation" {
		t.Fatalf("unaffordable rejection %s/%s", category, detail)
	}
	// Boundary: cost exactly one more than available (level 1, cost 2).
	boundary := &reputationRunner{t: t, catalogs: tree, bundle: "tree", state: reputationFounderState(t, tree, 22, now, 1), revision: 1, now: now}
	if category, detail := rejectionOf(t, boundary.purchase("rejects-cost-one-over-available", `"node_id":"reputation.starter.cash_small"`, now).Receipt); category != "unaffordable" || detail != "reputation" {
		t.Fatalf("boundary rejection %s/%s", category, detail)
	}
	rejections.cases = append(rejections.cases, boundary.cases...)
	if rejections.state.ReputationSpent != 3 || rejections.state.ReputationUnlockPPM != 50_000 {
		t.Fatalf("after two purchases spent=%d unlock=%d", rejections.state.ReputationSpent, rejections.state.ReputationUnlockPPM)
	}

	// A purchase chain across all nine nodes at the full tree cost, with one
	// purchase carrying an automatic Fiscal sweep.
	chain := &reputationRunner{t: t, catalogs: tree, bundle: "tree", state: reputationFounderState(t, tree, 22, now, 552), revision: 1, now: now}
	order := []string{"reputation.unlock.p05", "reputation.starter.cash_small", "reputation.starter.generated_beige_tower",
		"reputation.unlock.p25", "reputation.starter.upgrade_continuous_feed_paper", "reputation.starter.cash_large",
		"reputation.unlock.p50", "reputation.unlock.p75", "reputation.unlock.p100"}
	for index, id := range order {
		serverTS := now
		if index >= 3 {
			serverTS = now.Add(time.Second)
		}
		transition := chain.purchase("chain-"+id, fmt.Sprintf(`"node_id":%q`, id), serverTS)
		if transition.Outcome != save.IntentApplied {
			t.Fatalf("chain %s rejected: %s", id, transition.Receipt)
		}
		if index == 3 && (len(transition.Events) != 2 || transition.Events[0].Kind != save.EventFiscalPeriodHarvested) {
			t.Fatalf("sweep purchase events: %+v", transition.Events)
		}
	}
	if chain.state.ReputationSpent != 552 || chain.state.ReputationUnlockPPM != 1_000_000 || len(chain.state.ReputationNodesOwned) != 9 {
		t.Fatalf("full chain spent=%d unlock=%d owned=%v", chain.state.ReputationSpent, chain.state.ReputationUnlockPPM, chain.state.ReputationNodesOwned)
	}

	cases := append(append(append([]reputationCorpusCase{}, inactive.cases...), rejections.cases...), chain.cases...)
	return reputationCorpus{Version: 1, Exit: makeReputationExitFixture(t, now), Bundles: map[string]reputationCorpusBundle{
		"tree": {ConstantsHash: tree.ConstantsHash, Artifacts: stringArtifacts(tree.Artifacts)},
		"live": {ConstantsHash: live.ConstantsHash, Artifacts: stringArtifacts(live.Artifacts)},
	}, Cases: cases}
}

// TestReputationPurchaseCorpus is AC3 (the rejection table) and the Go half of
// AC4: it pins the Go-authored corpus the TS replay must byte-match.
// REPUTATION_UPDATE_FIXTURE=1 regenerates it.
func TestReputationPurchaseCorpus(t *testing.T) {
	corpus := buildReputationCorpus(t)
	encoded, err := json.MarshalIndent(corpus, "", " ")
	if err != nil {
		t.Fatal(err)
	}
	encoded = append(encoded, '\n')
	if os.Getenv("REPUTATION_UPDATE_FIXTURE") == "1" {
		if err := os.WriteFile(reputationCorpusPath, encoded, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	pinned, err := os.ReadFile(reputationCorpusPath)
	if err != nil {
		t.Fatalf("pinned corpus: %v", err)
	}
	if !bytes.Equal(pinned, encoded) {
		t.Fatal("reputation replay corpus drifted from the Go transition; regenerate with REPUTATION_UPDATE_FIXTURE=1 and review")
	}
}

// TestReputationPurchaseReplayRejectsTamperedResolvedInputs proves replay
// recomputes the frozen inputs rather than trusting them.
func TestReputationPurchaseReplayRejectsTamperedResolvedInputs(t *testing.T) {
	tree := reputationContentBundle(t)
	now := time.Date(2026, 9, 25, 12, 0, 0, 0, time.UTC)
	for name, mutate := range map[string]func(*founderReputationPurchaseResolved){
		"cost":        func(value *founderReputationPurchaseResolved) { value.ResolvedCost++ },
		"level":       func(value *founderReputationPurchaseResolved) { value.ReputationLevel++ },
		"spent":       func(value *founderReputationPurchaseResolved) { value.ReputationSpentBefore = 1 },
		"owned":       func(value *founderReputationPurchaseResolved) { value.OwnedBefore = []string{"reputation.unlock.p05"} },
		"node":        func(value *founderReputationPurchaseResolved) { value.NodeID = "reputation.unlock.p25" },
		"rejectCost0": func(value *founderReputationPurchaseResolved) { value.ResolvedCost = 0 },
	} {
		state := reputationFounderState(t, tree, 22, now, 10)
		request, err := ParseIntent([]byte(`{"intent_id":"01986666-5e00-7000-8000-000000000001","kind":"purchase_reputation_node","expected_revision":1,"node_id":"reputation.unlock.p05"}`))
		if err != nil {
			t.Fatal(err)
		}
		resolved, err := resolveReputationPurchase(tree, state, request.ReputationNodeID)
		if err != nil {
			t.Fatal(err)
		}
		mutate(&resolved)
		command := save.FounderReplayCommand{IntentID: request.IntentID, FounderStreamID: "01986666-5c00-4000-8000-000000000001",
			FounderID: "01986666-5d00-7000-8000-000000000001", Revision: 1, FounderLogSeq: 1, ServerTSMS: now.UnixMilli()}
		inputs, err := save.MarshalFounderReplayInputs(command, resolved)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := ApplyFounderLogged(state, request.CanonicalPayload, tree, inputs); err == nil {
			t.Fatalf("tampered %s resolved inputs replayed", name)
		}
	}
}

// makeReputationExitFixture mirrors makeCurriculumExitFixture on a tree
// bundle: the burnout curriculum starter (10 generated beige towers) plus the
// owned generated_beige_tower node (5) must yield exactly 15 generated and 0
// purchased, with cash_small's grant and run_started v2's summary.
func makeReputationExitFixture(t *testing.T, now time.Time) crossRuntimeActiveExit {
	t.Helper()
	current := reputationContentBundle(t)
	artifacts := cloneArtifactMap(current.Artifacts)
	curriculumBytes, err := os.ReadFile("../../balance/testdata/t0-t1/curriculum-v2.json")
	if err != nil {
		t.Fatal(err)
	}
	artifacts["curriculum"] = curriculumBytes
	nextHash, err := save.ConstantsHashArtifacts(artifacts)
	if err != nil {
		t.Fatal(err)
	}
	next := current
	next.ConstantsHash, next.Artifacts, next.Next = nextHash, artifacts, nil
	keys := map[string]struct{}{}
	for _, key := range copykeys.All() {
		keys[key] = struct{}{}
	}
	gateIDs := map[string]struct{}{}
	for _, gate := range next.Routes.Gates() {
		gateIDs[gate.ID] = struct{}{}
	}
	if next.Curriculum, err = curriculum.Load(curriculumBytes, curriculum.Declarations{Economy: next.Economy, CopyKeys: keys, GateIDs: gateIDs}); err != nil {
		t.Fatal(err)
	}
	if next.ReputationTree, err = reputation.LoadTree(artifacts["reputation_tree"], reputation.Declarations{Economy: next.Economy, Curriculum: next.Curriculum, CopyKeys: keys}); err != nil {
		t.Fatal(err)
	}
	if !next.valid(nextHash) {
		t.Fatal("next reputation bundle is invalid")
	}
	current.Next = &next
	founderID := "01986666-6d00-7000-8000-000000000001"
	company := replayFixtureState(t, current.Economy, now.Add(-15*time.Minute))
	company.WireVersion, company.Tier = 18, 1
	company.GatesCrossed["gate.t0_to_t1"] = true
	company.MeterBands = nil
	meterState, err := meters.NewRunState(current.Meters, 0)
	if err != nil {
		t.Fatal(err)
	}
	company.MeterValues, company.MeterDecayRemainders, company.MeterInputRemainders = meterState.Values, meterState.DecayRemainders, meterState.InputRemainders
	company.AchievementsEarnedRun = map[string]bool{}
	if _, err := initializeActivePlayState(company, current.Opportunities, founderID); err != nil {
		t.Fatal(err)
	}
	advanceActivePlayFixtureAttendance(t, company, current.Opportunities, current.Prestige, founderID, now)
	company.ManualTokenRefilledAt = now
	founder := reputationFounderState(t, current, 22, now, 6)
	founder.ReputationSpent, founder.ReputationUnlockPPM = 6, 50_000
	founder.ReputationNodesOwned = []string{"reputation.starter.cash_small", "reputation.starter.generated_beige_tower", "reputation.unlock.p05"}
	if err := current.ValidateFoundationState(founder); err != nil {
		t.Fatal(err)
	}
	preState := mustEncodeState(t, company)
	request, err := ParseIntent([]byte(`{"intent_id":"01986666-6d01-7000-8000-000000000001","kind":"perform_manual_batch","expected_revision":1,"action_id":"manual.click","count":1,"window_ms":1}`))
	if err != nil {
		t.Fatal(err)
	}
	carry := founderCarry(founder)
	carry.FounderRevision, carry.FounderConstantsHash = 1, current.ConstantsHash
	minigameActive := false
	activeEvidence, err := resolveActivePlaySchedule(company, current.Opportunities, current.Prestige, founderID, now)
	if err != nil {
		t.Fatal(err)
	}
	nextSpawn, err := next.Opportunities.Spawn(founderID, company.RunSeq+1, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	branch, err := next.Curriculum.SelectBranch(company, current.Economy)
	if err != nil || branch.Branch != "burnout" {
		t.Fatalf("reputation exit branch=%q err=%v", branch.Branch, err)
	}
	command := save.ReplayCommand{IntentID: request.IntentID, CompanyStreamID: "01986666-6e00-7000-8000-000000000001", FounderID: founderID, Revision: 1, RunSeq: 1, RunLogSeq: 1}
	inputs, err := buildReplayInputs(replayBuild{Command: command, Mode: ModeOnline, Now: now, IntentKind: request.Kind,
		RouteContextVersion: current.Routes.ContextVersion(), FounderCarry: &carry, Terminal: true, ExecutedRouteIDs: []string{},
		SelectedExitType: "scripted_first", SelectedBranch: &branch.Branch, SelectedTerms: json.RawMessage(`{}`), NextConstantsHash: next.ConstantsHash,
		ActivePlay: &activeEvidence, NextActivePlay: spawnEvidence(nextSpawn), MinigameSessionActive: &minigameActive})
	if err != nil {
		t.Fatal(err)
	}
	transition, err := ApplyLoggedExit(company, request.CanonicalPayload, current, inputs)
	if err != nil {
		t.Fatal(err)
	}
	newCompany := transition.Decision.NewCompanyState
	cash, _ := newCompany.Ledger.Balance("company.cash")
	if newCompany.GeneratorProvisioned["generator.beige_tower"] != 15 || newCompany.GeneratorCounts["generator.beige_tower"] != 0 || cash.String() != "1e3" {
		t.Fatalf("AC8 starters generated=%d purchased=%d cash=%s", newCompany.GeneratorProvisioned["generator.beige_tower"], newCompany.GeneratorCounts["generator.beige_tower"], cash)
	}
	started := transition.Decision.CompanyStartedEvents[0]
	if started.SchemaVersion != 2 || !bytes.Contains(started.Payload, []byte(`"applied_starter_node_ids":["reputation.starter.cash_small","reputation.starter.generated_beige_tower"]`)) ||
		!bytes.Contains(started.Payload, []byte(`"bonus_factor":"1.003e0"`)) {
		t.Fatalf("run_started v2 payload schema=%d %s", started.SchemaVersion, started.Payload)
	}
	company = replayFixtureStateFromEncoded(t, current, preState)
	result := executeTerminalFixture(t, "reputation-starters-after-burnout", current, company, preState, request, inputs, carry)
	return crossRuntimeActiveExit{ConstantsHash: current.ConstantsHash, Artifacts: stringArtifacts(current.Artifacts),
		NextConstantsHash: next.ConstantsHash, NextArtifacts: stringArtifacts(next.Artifacts), Case: result}
}

func replayFixtureStateFromEncoded(t *testing.T, catalogs CatalogBundle, encoded json.RawMessage) *save.State {
	t.Helper()
	state, err := save.RestoreState(encoded, 18, catalogs.Economy, economy.ScopeCompany, time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	return state
}

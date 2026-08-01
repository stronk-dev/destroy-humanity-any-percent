package production

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"cloud-clicker/server/decimal"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/epochseed"
	"cloud-clicker/server/multiplier"
	prestigecore "cloud-clicker/server/prestige"
	"cloud-clicker/server/save"
)

var updateReplayFixture = flag.Bool("update-replay-fixture", false, "rewrite the shared ApplyLogged fixture")

type crossRuntimeFixture struct {
	Version       int                        `json:"version"`
	ConstantsHash string                     `json:"constants_hash"`
	Artifacts     map[string]string          `json:"artifacts"`
	Cases         []crossRuntimeFixtureCase  `json:"cases"`
	TerminalCases []crossRuntimeTerminalCase `json:"terminal_cases"`
	Additional    []crossRuntimeBundleCase   `json:"additional_bundles"`
	FullRun       crossRuntimeFullRun        `json:"full_run"`
	RejectedExit  crossRuntimeFullRun        `json:"rejected_exit_run"`
}

type crossRuntimeFullRun struct {
	Genesis        json.RawMessage            `json:"genesis"`
	Entries        []crossRuntimeFullRunEntry `json:"entries"`
	FinalStateJSON string                     `json:"final_state_json"`
}

type crossRuntimeFullRunEntry struct {
	Seq              int64           `json:"seq"`
	CanonicalPayload json.RawMessage `json:"canonical_payload"`
	ReplayInputs     json.RawMessage `json:"replay_inputs"`
	ReceiptJSON      string          `json:"receipt_json"`
	EventsJSON       string          `json:"events_json"`
	Terminal         bool            `json:"terminal"`
}

type crossRuntimeBundleCase struct {
	ConstantsHash string                  `json:"constants_hash"`
	Artifacts     map[string]string       `json:"artifacts"`
	Case          crossRuntimeFixtureCase `json:"case"`
}

type crossRuntimeFixtureCase struct {
	Name             string          `json:"name"`
	PreState         json.RawMessage `json:"pre_state"`
	CanonicalPayload json.RawMessage `json:"canonical_payload"`
	ReplayInputs     json.RawMessage `json:"replay_inputs"`
	Outcome          string          `json:"outcome"`
	Receipt          json.RawMessage `json:"receipt"`
	Events           []fixtureEvent  `json:"events"`
	PostState        json.RawMessage `json:"post_state"`
	ReceiptJSON      string          `json:"receipt_json"`
	EventsJSON       string          `json:"events_json"`
	PostStateJSON    string          `json:"post_state_json"`
}

type fixtureEvent struct {
	Kind          string          `json:"kind"`
	SchemaVersion int             `json:"schema_version"`
	IntentID      string          `json:"intent_id"`
	Payload       json.RawMessage `json:"payload"`
}

type crossRuntimeTerminalCase struct {
	Name                 string          `json:"name"`
	PreState             json.RawMessage `json:"pre_state"`
	CanonicalPayload     json.RawMessage `json:"canonical_payload"`
	ReplayInputs         json.RawMessage `json:"replay_inputs"`
	Outcome              string          `json:"outcome"`
	Receipt              json.RawMessage `json:"receipt"`
	FounderOutput        any             `json:"founder_output"`
	FinalCompany         json.RawMessage `json:"final_company"`
	NewCompany           json.RawMessage `json:"new_company"`
	FounderEvents        []fixtureEvent  `json:"founder_events"`
	CompanyEndedEvents   []fixtureEvent  `json:"company_ended_events"`
	CompanyStartedEvents []fixtureEvent  `json:"company_started_events"`
	ReceiptJSON          string          `json:"receipt_json"`
	FounderOutputJSON    string          `json:"founder_output_json"`
	FinalCompanyJSON     string          `json:"final_company_json"`
	NewCompanyJSON       string          `json:"new_company_json"`
	FounderEventsJSON    string          `json:"founder_events_json"`
	CompanyEndedJSON     string          `json:"company_ended_events_json"`
	CompanyStartedJSON   string          `json:"company_started_events_json"`
}

func TestApplyLoggedCrossRuntimeFixture(t *testing.T) {
	fixture := makeCrossRuntimeFixture(t)
	path := filepath.Join("..", "..", "testdata", "replay", "apply-logged-v1.json")
	encoded, err := json.MarshalIndent(fixture, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	encoded = append(encoded, '\n')
	if *updateReplayFixture {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, encoded, 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	committed, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(committed, encoded) {
		t.Fatal("shared ApplyLogged fixture is stale; run the replay fixture generation target")
	}
}

func makeCrossRuntimeFixture(t *testing.T) crossRuntimeFixture {
	t.Helper()
	bundleBytes, err := epochseed.Load("../..")
	if err != nil {
		t.Fatal(err)
	}
	catalogs := loadReplayTestBundle(t, bundleBytes.Hash, bundleBytes.Artifacts)
	artifacts := make(map[string]string, len(bundleBytes.Artifacts))
	for name, data := range bundleBytes.Artifacts {
		artifacts[name] = string(data)
	}
	baseNow := time.Date(2026, 8, 1, 8, 0, 0, 0, time.UTC)
	cases := []struct {
		name          string
		payload       string
		advance       time.Duration
		mode          EvaluationMode
		configure     func(*save.State)
		contributions []multiplier.Contribution
		weight        *int64
		carry         *replayFounderCarry
		founderID     string
	}{
		{name: "manual-online", payload: `{"intent_id":"01986666-0001-7000-8000-000000000001","kind":"perform_manual_batch","expected_revision":1,"action_id":"manual.click","count":3,"window_ms":10}`, advance: time.Second, mode: ModeOnline},
		{name: "manual-offline-accrual", payload: `{"intent_id":"01986666-0002-7000-8000-000000000002","kind":"perform_manual_batch","expected_revision":1,"action_id":"manual.click","count":2,"window_ms":100}`, advance: 48 * time.Hour, mode: ModeOffline, configure: func(state *save.State) { state.GeneratorCounts["generator.beige_tower"] = 3 }},
		{name: "buy-generator", payload: `{"intent_id":"01986666-0003-7000-8000-000000000003","kind":"buy_generator","expected_revision":1,"generator_id":"generator.beige_tower","count":{"mode":"exact","value":4}}`, mode: ModeOnline, configure: func(state *save.State) { setCash(t, state, "1e4") }},
		{name: "buy-generator-max", payload: `{"intent_id":"01986666-0014-7000-8000-000000000014","kind":"buy_generator","expected_revision":1,"generator_id":"generator.beige_tower","count":{"mode":"max"}}`, mode: ModeOnline, configure: func(state *save.State) { setCash(t, state, "1e1000") }},
		{name: "cross-gate", payload: `{"intent_id":"01986666-0004-7000-8000-000000000004","kind":"cross_gate","expected_revision":1,"gate_id":"gate.t2_to_t3","route_id":null}`, mode: ModeOnline, configure: func(state *save.State) { state.Tier = 2; setCash(t, state, "1e10") }, carry: &replayFounderCarry{FounderRevision: 1, FounderConstantsHash: bundleBytes.Hash, NetworkSlots: []save.NetworkSlot{}, LedgerFactKinds: []string{}, ExitHistoryCount: 0}},
		{name: "cross-gate-offer-spawn", payload: `{"intent_id":"01986666-0011-7000-8000-000000000011","kind":"cross_gate","expected_revision":1,"gate_id":"gate.t2_to_t3","route_id":null}`, mode: ModeOnline,
			configure: func(state *save.State) {
				state.Tier = 2
				state.LifetimeValue = decimal.New(8, 12)
				setCash(t, state, "1e10")
			},
			carry:     &replayFounderCarry{FounderRevision: 1, FounderConstantsHash: bundleBytes.Hash, NetworkSlots: []save.NetworkSlot{}, LedgerFactKinds: []string{}, ExitHistoryCount: 1},
			founderID: offerFixtureFounder(t, catalogs.Prestige.SpawnGatePPM[3])},
		{name: "sign-compact", payload: `{"intent_id":"01986666-0005-7000-8000-000000000005","kind":"sign_compact","expected_revision":1,"tithe_ppm":110000}`, mode: ModeOnline},
		{name: "leave-compact", payload: `{"intent_id":"01986666-0006-7000-8000-000000000006","kind":"leave_compact","expected_revision":1}`, mode: ModeOnline, configure: func(state *save.State) { state.CompactMember, state.CompactTithePPM = true, 110_000 }, weight: int64Pointer(800_000)},
		{name: "incorporate-open-source", payload: `{"intent_id":"01986666-0007-7000-8000-000000000007","kind":"incorporate","expected_revision":1,"faction_id":"open_source"}`, mode: ModeOnline, configure: func(state *save.State) { state.Tier = 2 }},
		{name: "decline-offer", payload: `{"intent_id":"01986666-0008-7000-8000-000000000008","kind":"decline_exit_offer","expected_revision":1,"offer_id":"01986666-0008-7000-8000-000000000099"}`, mode: ModeOnline, configure: func(state *save.State) {
			state.OfferState = &save.ExitOfferState{OfferID: "01986666-0008-7000-8000-000000000099", ExitType: "acquisition", TermsJSON: json.RawMessage(`{"market_modifier_ppm":1000000,"payout_preview":{"reputation_delta":0,"network_slot_unlocks":[],"route_knowledge":0,"clout_reach_note":"clout.reach.preserved"}}`), SpawnedAt: baseNow.Add(-time.Minute), ExpiresAt: baseNow.Add(time.Minute)}
		}},
		{name: "offer-expires-during-manual-batch", payload: `{"intent_id":"01986666-0015-7000-8000-000000000015","kind":"perform_manual_batch","expected_revision":1,"action_id":"manual.click","count":1,"window_ms":1}`, mode: ModeOnline, configure: func(state *save.State) {
			state.OfferState = &save.ExitOfferState{OfferID: "01986666-0015-7000-8000-000000000099", ExitType: "acquisition", TermsJSON: json.RawMessage(`{"market_modifier_ppm":1000000,"payout_preview":{"reputation_delta":0,"network_slot_unlocks":[],"route_knowledge":0,"clout_reach_note":"clout.reach.preserved"}}`), SpawnedAt: baseNow.Add(-time.Minute), ExpiresAt: baseNow}
		}},
		{name: "skip-ahead-gate", payload: `{"intent_id":"01986666-0016-7000-8000-000000000016","kind":"cross_gate","expected_revision":1,"gate_id":"gate.t2_to_t3","route_id":null}`, mode: ModeOnline, configure: func(state *save.State) { state.Tier = 1; setCash(t, state, "1e10") }, carry: &replayFounderCarry{FounderRevision: 1, FounderConstantsHash: bundleBytes.Hash, NetworkSlots: []save.NetworkSlot{}, LedgerFactKinds: []string{}, ExitHistoryCount: 0}},
		{name: "lower-gate-after-higher", payload: `{"intent_id":"01986666-0017-7000-8000-000000000017","kind":"cross_gate","expected_revision":1,"gate_id":"gate.t2_to_t3","route_id":null}`, mode: ModeOnline, configure: func(state *save.State) { state.Tier = 4; setCash(t, state, "1e10") }, carry: &replayFounderCarry{FounderRevision: 1, FounderConstantsHash: bundleBytes.Hash, NetworkSlots: []save.NetworkSlot{}, LedgerFactKinds: []string{}, ExitHistoryCount: 0}},
		{name: "closed-hook-chain", payload: `{"intent_id":"01986666-0009-7000-8000-000000000009","kind":"perform_manual_batch","expected_revision":1,"action_id":"manual.click","count":1,"window_ms":60000}`, advance: time.Minute, mode: ModeOnline,
			configure: func(state *save.State) {
				state.GeneratorCounts["generator.beige_tower"] = 2
				state.FactionID = "open_source"
				state.IncorporatedAt = baseNow.Add(-time.Minute)
				state.CompactMember, state.CompactTithePPM = true, 130_000
			},
			contributions: []multiplier.Contribution{{Slot: multiplier.SlotFaction, SourceID: "guild.stock_consumption", Target: "all", Factor: decimal.One}, {Slot: multiplier.SlotCommons, SourceID: "commons.member", Target: "all", Factor: decimal.New(11, -1)}}, weight: int64Pointer(812_345)},
		{name: "invalid-manual", payload: `{"intent_id":"01986666-0010-7000-8000-000000000010","kind":"perform_manual_batch","expected_revision":1,"action_id":"manual.click","count":0,"window_ms":1}`, mode: ModeOnline},
		{name: "invalid-buy-generator-id", payload: `{"intent_id":"01986666-0101-7000-8000-000000000101","kind":"buy_generator","expected_revision":1,"generator_id":"INVALID","count":{"mode":"exact","value":1}}`, mode: ModeOnline},
		{name: "invalid-buy-count-object", payload: `{"intent_id":"01986666-0102-7000-8000-000000000102","kind":"buy_generator","expected_revision":1,"generator_id":"generator.beige_tower","count":1}`, mode: ModeOnline},
		{name: "invalid-buy-count-max", payload: `{"intent_id":"01986666-0103-7000-8000-000000000103","kind":"buy_generator","expected_revision":1,"generator_id":"generator.beige_tower","count":{"mode":"max","value":1}}`, mode: ModeOnline},
		{name: "invalid-buy-count-exact", payload: `{"intent_id":"01986666-0104-7000-8000-000000000104","kind":"buy_generator","expected_revision":1,"generator_id":"generator.beige_tower","count":{"mode":"exact","value":0}}`, mode: ModeOnline},
		{name: "invalid-buy-count-mode", payload: `{"intent_id":"01986666-0105-7000-8000-000000000105","kind":"buy_generator","expected_revision":1,"generator_id":"generator.beige_tower","count":{"mode":"some"}}`, mode: ModeOnline},
		{name: "invalid-manual-action", payload: `{"intent_id":"01986666-0106-7000-8000-000000000106","kind":"perform_manual_batch","expected_revision":1,"action_id":"INVALID","count":1,"window_ms":1}`, mode: ModeOnline},
		{name: "invalid-manual-window-last", payload: `{"intent_id":"01986666-0107-7000-8000-000000000107","kind":"perform_manual_batch","expected_revision":1,"action_id":"INVALID","count":0,"window_ms":0}`, mode: ModeOnline},
		{name: "invalid-cross-gate-id", payload: `{"intent_id":"01986666-0108-7000-8000-000000000108","kind":"cross_gate","expected_revision":1,"gate_id":"INVALID","route_id":null}`, mode: ModeOnline},
		{name: "invalid-cross-route-last", payload: `{"intent_id":"01986666-0109-7000-8000-000000000109","kind":"cross_gate","expected_revision":1,"gate_id":"INVALID","route_id":"INVALID"}`, mode: ModeOnline},
		{name: "invalid-compact-tithe", payload: `{"intent_id":"01986666-0110-7000-8000-000000000110","kind":"sign_compact","expected_revision":1,"tithe_ppm":1000001}`, mode: ModeOnline},
		{name: "invalid-incorporate-faction", payload: `{"intent_id":"01986666-0111-7000-8000-000000000111","kind":"incorporate","expected_revision":1,"faction_id":"INVALID"}`, mode: ModeOnline},
		{name: "invalid-decline-offer", payload: `{"intent_id":"01986666-0112-7000-8000-000000000112","kind":"decline_exit_offer","expected_revision":1,"offer_id":"INVALID"}`, mode: ModeOnline},
		{name: "invalid-accept-offer-last", payload: `{"intent_id":"01986666-0113-7000-8000-000000000113","kind":"accept_exit_offer","expected_revision":1,"expected_founder_revision":0,"offer_id":"INVALID"}`, mode: ModeOnline},
		{name: "invalid-wind-down-founder", payload: `{"intent_id":"01986666-0114-7000-8000-000000000114","kind":"wind_down","expected_revision":1,"expected_founder_revision":0}`, mode: ModeOnline},
		{name: "invalid-leave-fields", payload: `{"intent_id":"01986666-0115-7000-8000-000000000115","kind":"leave_compact","expected_revision":1,"extra":true}`, mode: ModeOnline},
		{name: "unknown-buy-rejects-before-accrual", payload: `{"intent_id":"01986666-0012-7000-8000-000000000012","kind":"buy_generator","expected_revision":1,"generator_id":"generator.unknown","count":{"mode":"exact","value":1}}`, advance: time.Minute, mode: ModeOnline,
			configure: func(state *save.State) { state.GeneratorCounts["generator.beige_tower"] = 3 }},
		{name: "existing-compact-rejects-before-accrual", payload: `{"intent_id":"01986666-0013-7000-8000-000000000013","kind":"sign_compact","expected_revision":1,"tithe_ppm":110000}`, advance: time.Minute, mode: ModeOnline,
			configure: func(state *save.State) {
				state.GeneratorCounts["generator.beige_tower"] = 3
				state.CompactMember, state.CompactTithePPM = true, 110_000
			}, weight: int64Pointer(800_000)},
	}
	result := crossRuntimeFixture{Version: 1, ConstantsHash: bundleBytes.Hash, Artifacts: artifacts, Cases: make([]crossRuntimeFixtureCase, 0, len(cases))}
	for index, definition := range cases {
		state := replayFixtureState(t, catalogs.Economy, baseNow)
		if definition.configure != nil {
			definition.configure(state)
		}
		preState, err := save.EncodeState(state)
		if err != nil {
			t.Fatalf("%s pre-state: %v", definition.name, err)
		}
		request, err := ParseIntent([]byte(definition.payload))
		if err != nil {
			t.Fatalf("%s parse: %v", definition.name, err)
		}
		founderID := definition.founderID
		if founderID == "" {
			founderID = "01986666-2000-7000-8000-000000000001"
		}
		command := save.ReplayCommand{IntentID: request.IntentID, CompanyStreamID: "01986666-1000-7000-8000-000000000001", FounderID: founderID, Revision: 1, RunSeq: 1, RunLogSeq: int64(index + 1)}
		build := replayBuild{Command: command, Mode: definition.mode, Now: baseNow.Add(definition.advance), IntentKind: request.Kind,
			Contributions: definition.contributions, CommonsWeightPPM: definition.weight, RouteContextVersion: catalogs.Routes.ContextVersion(), FounderCarry: definition.carry}
		inputs, err := buildReplayInputs(build)
		if err != nil {
			t.Fatalf("%s inputs: %v", definition.name, err)
		}
		transition, err := ApplyLogged(state, request.CanonicalPayload, catalogs, inputs)
		if err != nil {
			t.Fatalf("%s transition: %v", definition.name, err)
		}
		postState, err := save.EncodeState(state)
		if err != nil {
			t.Fatalf("%s post-state: %v", definition.name, err)
		}
		events := make([]fixtureEvent, len(transition.Events))
		for eventIndex, event := range transition.Events {
			events[eventIndex] = fixtureEvent{Kind: string(event.Kind), SchemaVersion: event.SchemaVersion, IntentID: event.IntentID, Payload: event.Payload}
		}
		result.Cases = append(result.Cases, crossRuntimeFixtureCase{Name: definition.name, PreState: preState, CanonicalPayload: request.CanonicalPayload,
			ReplayInputs: inputs, Outcome: string(transition.Outcome), Receipt: transition.Receipt, Events: events, PostState: postState,
			ReceiptJSON: canonicalFixtureJSON(t, transition.Receipt), EventsJSON: canonicalFixtureValue(t, events), PostStateJSON: canonicalFixtureJSON(t, postState)})
	}
	result.TerminalCases = []crossRuntimeTerminalCase{
		makeTerminalFixtureCase(t, catalogs, bundleBytes.Hash, baseNow),
		makeAcceptedOfferFixtureCase(t, catalogs, bundleBytes.Hash, baseNow),
		makeScriptedGateFixtureCase(t, catalogs, bundleBytes.Hash, baseNow),
	}
	result.Additional = []crossRuntimeBundleCase{makeFallbackInvariantFixture(t, bundleBytes.Artifacts, baseNow)}
	result.FullRun = makeFullRunFixture(t, catalogs, bundleBytes.Hash, baseNow)
	result.RejectedExit = makeRejectedExitFixture(t, catalogs, bundleBytes.Hash, baseNow)
	return result
}

func makeRejectedExitFixture(t *testing.T, catalogs CatalogBundle, constantsHash string, now time.Time) crossRuntimeFullRun {
	t.Helper()
	assertPerEntryNextCatalog(t, catalogs, constantsHash, now)
	state := replayFixtureState(t, catalogs.Economy, now)
	state.Tier = 0
	genesis, err := save.EncodeState(state)
	if err != nil {
		t.Fatal(err)
	}
	founderID := "01987778-2000-7000-8000-000000000001"
	request, err := parseLoggedIntent([]byte(`{"kind":"wind_down","expected_revision":1,"expected_founder_revision":1}`), "01987778-0001-7000-8000-000000000001")
	if err != nil {
		t.Fatal(err)
	}
	command := save.ReplayCommand{IntentID: request.IntentID, CompanyStreamID: "01987778-1000-7000-8000-000000000001", FounderID: founderID, Revision: 1, RunSeq: 1, RunLogSeq: 1}
	carry := replayFounderCarry{FounderRevision: 1, FounderConstantsHash: constantsHash, NetworkSlots: []save.NetworkSlot{}, LedgerFactKinds: []string{}}
	inputs, err := buildReplayInputs(replayBuild{Command: command, Mode: ModeOnline, Now: now, IntentKind: request.Kind, RouteContextVersion: catalogs.Routes.ContextVersion(), FounderCarry: &carry, Terminal: true, ExecutedRouteIDs: []string{}, SelectedExitType: "scripted_first", SelectedTerms: json.RawMessage(`{}`), NextConstantsHash: constantsHash})
	if err != nil {
		t.Fatal(err)
	}
	exit, err := ApplyLoggedExit(state, request.CanonicalPayload, catalogs, inputs)
	if err != nil || exit.Decision.Outcome != save.IntentRejected {
		t.Fatalf("rejected exit outcome=%s err=%v", exit.Decision.Outcome, err)
	}
	entries := []crossRuntimeFullRunEntry{{Seq: 1, CanonicalPayload: request.CanonicalPayload, ReplayInputs: inputs, ReceiptJSON: canonicalFixtureJSON(t, exit.Decision.Receipt), EventsJSON: "[]", Terminal: true}}
	request, err = parseLoggedIntent([]byte(`{"kind":"perform_manual_batch","expected_revision":1,"action_id":"manual.click","count":1,"window_ms":1000}`), "01987778-0002-7000-8000-000000000002")
	if err != nil {
		t.Fatal(err)
	}
	command = save.ReplayCommand{IntentID: request.IntentID, CompanyStreamID: command.CompanyStreamID, FounderID: founderID, Revision: 1, RunSeq: 1, RunLogSeq: 2}
	inputs, err = buildReplayInputs(replayBuild{Command: command, Mode: ModeOnline, Now: now.Add(time.Second), IntentKind: request.Kind, RouteContextVersion: catalogs.Routes.ContextVersion()})
	if err != nil {
		t.Fatal(err)
	}
	ordinary, err := ApplyLogged(state, request.CanonicalPayload, catalogs, inputs)
	if err != nil {
		t.Fatal(err)
	}
	entries = append(entries, crossRuntimeFullRunEntry{Seq: 2, CanonicalPayload: request.CanonicalPayload, ReplayInputs: inputs, ReceiptJSON: canonicalFixtureJSON(t, ordinary.Receipt), EventsJSON: canonicalFixtureValue(t, fixtureEvents(ordinary.Events))})
	full := crossRuntimeFullRun{Genesis: genesis, Entries: entries, FinalStateJSON: canonicalFixtureJSON(t, mustEncodeState(t, state))}
	verifiedEntries := []ReplayLogEntry{
		{Sequence: 1, CanonicalPayload: entries[0].CanonicalPayload, ReplayInputs: entries[0].ReplayInputs, ReceiptJSON: []byte(entries[0].ReceiptJSON), EventsJSON: []byte(entries[0].EventsJSON), Terminal: true},
		{Sequence: 2, CanonicalPayload: entries[1].CanonicalPayload, ReplayInputs: entries[1].ReplayInputs, ReceiptJSON: []byte(entries[1].ReceiptJSON), EventsJSON: []byte(entries[1].EventsJSON)},
	}
	if verdict := VerifyReplayRun(genesis, save.CurrentVersion, catalogs, verifiedEntries, constantsHash, false); verdict != ReplayLogGap {
		t.Fatalf("rejected-exit continuation verdict=%s", verdict)
	}
	tampered := append([]ReplayLogEntry(nil), verifiedEntries...)
	tampered[1].ReceiptJSON = []byte(`{}`)
	if verdict := VerifyReplayRun(genesis, save.CurrentVersion, catalogs, tampered, constantsHash, false); verdict != ReplayStateDivergence {
		t.Fatalf("rejected-exit continuation tamper verdict=%s", verdict)
	}
	if verdict := VerifyReplayRun(genesis, save.CurrentVersion, catalogs, verifiedEntries[:1], constantsHash, false); verdict != ReplayLogGap {
		t.Fatalf("final rejected exit verdict=%s", verdict)
	}
	clock := append([]ReplayLogEntry(nil), verifiedEntries...)
	var wire map[string]any
	if err := json.Unmarshal(clock[0].ReplayInputs, &wire); err != nil {
		t.Fatal(err)
	}
	wire["evaluated_at_ms"] = 1
	clock[0].ReplayInputs, _ = json.Marshal(wire)
	if verdict := VerifyReplayRun(genesis, save.CurrentVersion, catalogs, clock, constantsHash, false); verdict != ReplayClockViolation {
		t.Fatalf("rejected exit clock verdict=%s", verdict)
	}
	return full
}

func assertPerEntryNextCatalog(t *testing.T, catalogs CatalogBundle, constantsHash string, now time.Time) {
	t.Helper()
	alternate := func(suffix byte) CatalogBundle {
		artifacts := make(map[string][]byte, len(catalogs.Artifacts))
		for name, data := range catalogs.Artifacts {
			artifacts[name] = append([]byte(nil), data...)
		}
		// JSON whitespace changes immutable artifact identity while preserving the
		// already-loaded policy objects used by this boundary test.
		artifacts["economy"] = append(artifacts["economy"], '\n', ' ', suffix)
		hash, err := save.ConstantsHashArtifacts(artifacts)
		if err != nil {
			t.Fatal(err)
		}
		result := catalogs
		result.ConstantsHash, result.Artifacts, result.Next = hash, artifacts, nil
		return result
	}
	nextA, nextB := alternate(' '), alternate('\t')
	state := replayFixtureState(t, catalogs.Economy, now)
	state.Tier = 0
	genesis := mustEncodeState(t, state)
	entries := make([]ReplayLogEntry, 0, 2)
	for index, next := range []CatalogBundle{nextA, nextB} {
		intentID := fmt.Sprintf("01987779-000%d-7000-8000-%012d", index+1, index+1)
		request, err := parseLoggedIntent([]byte(`{"kind":"wind_down","expected_revision":1,"expected_founder_revision":1}`), intentID)
		if err != nil {
			t.Fatal(err)
		}
		command := save.ReplayCommand{IntentID: intentID, CompanyStreamID: "01987779-1000-7000-8000-000000000001",
			FounderID: "01987779-2000-7000-8000-000000000001", Revision: 1, RunSeq: 1, RunLogSeq: int64(index + 1)}
		carry := replayFounderCarry{FounderRevision: 1, FounderConstantsHash: constantsHash, NetworkSlots: []save.NetworkSlot{}, LedgerFactKinds: []string{}}
		inputs, err := buildReplayInputs(replayBuild{Command: command, Mode: ModeOnline, Now: now.Add(time.Duration(index) * time.Second),
			IntentKind: request.Kind, RouteContextVersion: catalogs.Routes.ContextVersion(), FounderCarry: &carry, Terminal: true,
			ExecutedRouteIDs: []string{}, SelectedExitType: "scripted_first", SelectedTerms: json.RawMessage(`{}`), NextConstantsHash: next.ConstantsHash})
		if err != nil {
			t.Fatal(err)
		}
		executionCatalogs := catalogs
		executionCatalogs.Next = &next
		transition, err := ApplyLoggedExit(state, request.CanonicalPayload, executionCatalogs, inputs)
		if err != nil || transition.Decision.Outcome != save.IntentRejected {
			t.Fatalf("per-entry exit %d outcome=%s err=%v", index, transition.Decision.Outcome, err)
		}
		entryNext := next
		entries = append(entries, ReplayLogEntry{Sequence: int64(index + 1), CanonicalPayload: request.CanonicalPayload,
			ReplayInputs: inputs, ReceiptJSON: []byte(canonicalFixtureJSON(t, transition.Decision.Receipt)), EventsJSON: []byte(`[]`),
			Terminal: true, NextCatalog: &entryNext})
	}
	// A run-wide slot deliberately points at the last row. The first row must
	// still resolve its own bundle; prior last-write-wins behavior failed here.
	catalogs.Next = &nextB
	if verdict := VerifyReplayRun(genesis, save.CurrentVersion, catalogs, entries, constantsHash, false); verdict != ReplayLogGap {
		t.Fatalf("per-entry next-catalog verdict=%s", verdict)
	}
	wrong := append([]ReplayLogEntry(nil), entries...)
	wrong[0].NextCatalog = &nextB
	if verdict := VerifyReplayRun(genesis, save.CurrentVersion, catalogs, wrong, constantsHash, false); verdict != ReplayStateDivergence {
		t.Fatalf("wrong per-entry next-catalog verdict=%s", verdict)
	}
}

func mustEncodeState(t *testing.T, state *save.State) json.RawMessage {
	t.Helper()
	encoded, err := save.EncodeState(state)
	if err != nil {
		t.Fatal(err)
	}
	return encoded
}

func makeFullRunFixture(t *testing.T, catalogs CatalogBundle, constantsHash string, startedAt time.Time) crossRuntimeFullRun {
	t.Helper()
	founderID := offerFixtureFounder(t, catalogs.Prestige.SpawnGatePPM[3])
	state := replayFixtureState(t, catalogs.Economy, startedAt)
	state.Tier = 2
	state.LifetimeValue = decimal.New(8, 12)
	setCash(t, state, "1e10")
	genesis, err := save.EncodeState(state)
	if err != nil {
		t.Fatal(err)
	}
	entries := make([]crossRuntimeFullRunEntry, 0, 51)
	revision := int64(1)
	now := startedAt
	ordinary := []string{
		`{"kind":"perform_manual_batch","expected_revision":%d,"action_id":"manual.click","count":2,"window_ms":1000}`,
		`{"kind":"buy_generator","expected_revision":%d,"generator_id":"generator.beige_tower","count":{"mode":"exact","value":3}}`,
		`{"kind":"sign_compact","expected_revision":%d,"tithe_ppm":110000}`,
		`{"kind":"leave_compact","expected_revision":%d}`,
		`{"kind":"incorporate","expected_revision":%d,"faction_id":"vc_funded"}`,
		`{"kind":"cross_gate","expected_revision":%d,"gate_id":"gate.t2_to_t3","route_id":null}`,
	}
	for index := 0; index < 50; index++ {
		mode := ModeOnline
		if index == 10 {
			now = now.Add(48 * time.Hour)
			mode = ModeOffline
		} else {
			now = now.Add(time.Second)
		}
		payload := fmt.Sprintf(`{"kind":"perform_manual_batch","expected_revision":%d,"action_id":"manual.click","count":1,"window_ms":1000}`, revision)
		if index < len(ordinary) {
			payload = fmt.Sprintf(ordinary[index], revision)
		}
		if index == 6 {
			payload = fmt.Sprintf(`{"kind":"decline_exit_offer","expected_revision":%d,"offer_id":"%s"}`, revision, prestigecore.OfferID(founderID, 1, 3, 0, now.Add(-time.Second)))
		}
		intentID := fmt.Sprintf("01987777-%04d-7000-8000-%012d", index+1, index+1)
		request, err := parseLoggedIntent([]byte(payload), intentID)
		if err != nil {
			t.Fatal(err)
		}
		command := save.ReplayCommand{IntentID: intentID, CompanyStreamID: "01987777-1000-7000-8000-000000000001", FounderID: founderID, Revision: revision, RunSeq: 1, RunLogSeq: int64(index + 1)}
		var carry *replayFounderCarry
		if request.Kind == IntentCrossGate {
			carry = &replayFounderCarry{FounderRevision: 1, FounderConstantsHash: constantsHash, NetworkSlots: []save.NetworkSlot{}, LedgerFactKinds: []string{}, ExitHistoryCount: 1}
		}
		var weight *int64
		if state.CompactMember {
			weight = int64Pointer(800_000)
		}
		inputs, err := buildReplayInputs(replayBuild{Command: command, Mode: mode, Now: now, IntentKind: request.Kind, RouteContextVersion: catalogs.Routes.ContextVersion(), FounderCarry: carry, CommonsWeightPPM: weight})
		if err != nil {
			t.Fatal(err)
		}
		transition, err := ApplyLogged(state, request.CanonicalPayload, catalogs, inputs)
		if err != nil || transition.Outcome != save.IntentApplied {
			t.Fatalf("full run step %d outcome=%s err=%v", index+1, transition.Outcome, err)
		}
		entries = append(entries, crossRuntimeFullRunEntry{Seq: int64(index + 1), CanonicalPayload: request.CanonicalPayload, ReplayInputs: inputs, ReceiptJSON: canonicalFixtureJSON(t, transition.Receipt), EventsJSON: canonicalFixtureValue(t, fixtureEvents(transition.Events))})
		revision++
	}
	now = now.Add(time.Second)
	intentID := "01987777-0051-7000-8000-000000000051"
	request, err := parseLoggedIntent([]byte(fmt.Sprintf(`{"kind":"wind_down","expected_revision":%d,"expected_founder_revision":1}`, revision)), intentID)
	if err != nil {
		t.Fatal(err)
	}
	command := save.ReplayCommand{IntentID: intentID, CompanyStreamID: "01987777-1000-7000-8000-000000000001", FounderID: founderID, Revision: revision, RunSeq: 1, RunLogSeq: 51}
	carry := replayFounderCarry{FounderRevision: 1, FounderConstantsHash: constantsHash, NetworkSlots: []save.NetworkSlot{}, LedgerFactKinds: []string{}, ExitHistoryCount: 1}
	inputs, err := buildReplayInputs(replayBuild{Command: command, Mode: ModeOnline, Now: now, IntentKind: request.Kind, RouteContextVersion: catalogs.Routes.ContextVersion(), FounderCarry: &carry, Terminal: true, ExecutedRouteIDs: []string{}, SelectedExitType: "collapse", SelectedTerms: json.RawMessage(`{}`), NextConstantsHash: constantsHash})
	if err != nil {
		t.Fatal(err)
	}
	transition, err := ApplyLoggedExit(state, request.CanonicalPayload, catalogs, inputs)
	if err != nil || transition.Decision.Outcome != save.IntentApplied {
		t.Fatalf("full run terminal outcome=%s err=%v", transition.Decision.Outcome, err)
	}
	allEvents := append(fixtureEvents(transition.Decision.FounderEvents), fixtureEvents(transition.Decision.CompanyEndedEvents)...)
	allEvents = append(allEvents, fixtureEvents(transition.Decision.CompanyStartedEvents)...)
	entries = append(entries, crossRuntimeFullRunEntry{Seq: 51, CanonicalPayload: request.CanonicalPayload, ReplayInputs: inputs, ReceiptJSON: canonicalFixtureJSON(t, transition.Decision.Receipt), EventsJSON: canonicalFixtureValue(t, allEvents), Terminal: true})
	finalState, err := save.EncodeState(transition.Company)
	if err != nil {
		t.Fatal(err)
	}
	full := crossRuntimeFullRun{Genesis: genesis, Entries: entries, FinalStateJSON: canonicalFixtureJSON(t, finalState)}
	assertFullRunVerifier(t, full, catalogs, constantsHash)
	return full
}

func assertFullRunVerifier(t *testing.T, full crossRuntimeFullRun, catalogs CatalogBundle, constantsHash string) {
	t.Helper()
	convert := func(source []crossRuntimeFullRunEntry) []ReplayLogEntry {
		result := make([]ReplayLogEntry, len(source))
		for index, entry := range source {
			result[index] = ReplayLogEntry{Sequence: entry.Seq, CanonicalPayload: entry.CanonicalPayload, ReplayInputs: entry.ReplayInputs, ReceiptJSON: []byte(entry.ReceiptJSON), EventsJSON: []byte(entry.EventsJSON), Terminal: entry.Terminal}
		}
		return result
	}
	entries := convert(full.Entries)
	verdict, finalState := verifyReplayRunDetailed(full.Genesis, save.CurrentVersion, catalogs, entries, constantsHash, false)
	if verdict != ReplayVerified {
		t.Fatalf("full run verdict=%s", verdict)
	}
	encodedFinal, err := save.EncodeState(finalState)
	if err != nil {
		t.Fatalf("encode verified final state: %v", err)
	}
	if got := canonicalFixtureJSON(t, encodedFinal); got != full.FinalStateJSON {
		t.Fatalf("verified final state differs\ngot:  %s\nwant: %s", got, full.FinalStateJSON)
	}
	gap := append([]ReplayLogEntry(nil), entries[:19]...)
	gap = append(gap, entries[20:]...)
	if verdict := VerifyReplayRun(full.Genesis, save.CurrentVersion, catalogs, gap, constantsHash, false); verdict != ReplayLogGap {
		t.Fatalf("gap verdict=%s", verdict)
	}
	if verdict := VerifyReplayRun(full.Genesis, save.CurrentVersion, catalogs, entries, "sha256:"+strings.Repeat("f", 64), false); verdict != ReplayConstantsMismatch {
		t.Fatalf("constants verdict=%s", verdict)
	}
	if verdict := VerifyReplayRun(full.Genesis, save.CurrentVersion, catalogs, entries, constantsHash, true); verdict != ReplayEngineMismatch {
		t.Fatalf("engine verdict=%s", verdict)
	}
	tampered := append([]ReplayLogEntry(nil), entries...)
	tampered[10].ReceiptJSON = []byte(`{}`)
	if verdict := VerifyReplayRun(full.Genesis, save.CurrentVersion, catalogs, tampered, constantsHash, false); verdict != ReplayStateDivergence {
		t.Fatalf("state verdict=%s", verdict)
	}
	corrupt := append([]ReplayLogEntry(nil), entries...)
	corrupt[10].ReceiptJSON = []byte(`not-json`)
	if verdict := VerifyReplayRun(full.Genesis, save.CurrentVersion, catalogs, corrupt, constantsHash, false); verdict != ReplayStateDivergence {
		t.Fatalf("corrupt receipt verdict=%s", verdict)
	}
	eventTamper := append([]ReplayLogEntry(nil), entries...)
	eventTamper[1].EventsJSON = []byte(`[]`)
	if verdict := VerifyReplayRun(full.Genesis, save.CurrentVersion, catalogs, eventTamper, constantsHash, false); verdict != ReplayStateDivergence {
		t.Fatalf("event tamper verdict=%s", verdict)
	}
	clock := append([]ReplayLogEntry(nil), entries...)
	var wire map[string]any
	if err := json.Unmarshal(clock[10].ReplayInputs, &wire); err != nil {
		t.Fatal(err)
	}
	wire["evaluated_at_ms"] = 1
	clock[10].ReplayInputs, _ = json.Marshal(wire)
	if verdict := VerifyReplayRun(full.Genesis, save.CurrentVersion, catalogs, clock, constantsHash, false); verdict != ReplayClockViolation {
		t.Fatalf("clock verdict=%s", verdict)
	}
}

func makeFallbackInvariantFixture(t *testing.T, source map[string][]byte, now time.Time) crossRuntimeBundleCase {
	t.Helper()
	artifacts := make(map[string][]byte, len(source))
	for name, data := range source {
		artifacts[name] = bytes.Clone(data)
	}
	var economyArtifact map[string]any
	if err := json.Unmarshal(artifacts["economy"], &economyArtifact); err != nil {
		t.Fatal(err)
	}
	resources := economyArtifact["resources"].([]any)
	resources[0].(map[string]any)["hardcap"].(map[string]any)["amount"] = "1e4000000000000000"
	modified, err := json.Marshal(economyArtifact)
	if err != nil {
		t.Fatal(err)
	}
	artifacts["economy"] = modified
	hash, err := save.ConstantsHashArtifacts(artifacts)
	if err != nil {
		t.Fatal(err)
	}
	catalogs := loadReplayTestBundle(t, hash, artifacts)
	state := replayFixtureState(t, catalogs.Economy, now)
	setCash(t, state, "1e4000000000000000")
	state.OfferState = &save.ExitOfferState{OfferID: "01986666-0201-7000-8000-000000000299", ExitType: "acquisition", TermsJSON: json.RawMessage(`{"market_modifier_ppm":1000000,"payout_preview":{"reputation_delta":0,"network_slot_unlocks":[],"route_knowledge":0,"clout_reach_note":"clout.reach.preserved"}}`), SpawnedAt: now.Add(-time.Minute), ExpiresAt: now}
	preState, err := save.EncodeState(state)
	if err != nil {
		t.Fatal(err)
	}
	request, err := ParseIntent([]byte(`{"intent_id":"01986666-0201-7000-8000-000000000201","kind":"buy_generator","expected_revision":1,"generator_id":"generator.beige_tower","count":{"mode":"max"}}`))
	if err != nil {
		t.Fatal(err)
	}
	command := save.ReplayCommand{IntentID: request.IntentID, CompanyStreamID: "01986666-1000-7000-8000-000000000001", FounderID: "01986666-2000-7000-8000-000000000001", Revision: 1, RunSeq: 1, RunLogSeq: 1}
	inputs, err := buildReplayInputs(replayBuild{Command: command, Mode: ModeOnline, Now: now, IntentKind: request.Kind, RouteContextVersion: catalogs.Routes.ContextVersion()})
	if err != nil {
		t.Fatal(err)
	}
	transition, err := ApplyLogged(state, request.CanonicalPayload, catalogs, inputs)
	if err != nil {
		t.Fatal(err)
	}
	if len(transition.Invariants) != 1 || transition.Invariants[0].Kind != InvariantAffordFallback {
		t.Fatalf("fallback invariant missing: %+v", transition.Invariants)
	}
	postState, err := save.EncodeState(state)
	if err != nil {
		t.Fatal(err)
	}
	events := fixtureEvents(transition.Events)
	fixtureArtifacts := make(map[string]string, len(artifacts))
	for name, data := range artifacts {
		fixtureArtifacts[name] = string(data)
	}
	return crossRuntimeBundleCase{ConstantsHash: hash, Artifacts: fixtureArtifacts, Case: crossRuntimeFixtureCase{
		Name: "buy-generator-max-fallback-invariant", PreState: preState, CanonicalPayload: request.CanonicalPayload,
		ReplayInputs: inputs, Outcome: string(transition.Outcome), Receipt: transition.Receipt, Events: events, PostState: postState,
		ReceiptJSON: canonicalFixtureJSON(t, transition.Receipt), EventsJSON: canonicalFixtureValue(t, events), PostStateJSON: canonicalFixtureJSON(t, postState),
	}}
}

func makeTerminalFixtureCase(t *testing.T, catalogs CatalogBundle, constantsHash string, now time.Time) crossRuntimeTerminalCase {
	t.Helper()
	company := replayFixtureState(t, catalogs.Economy, now.Add(-20*time.Minute))
	company.Tier = 1
	company.LifetimeValue = decimal.New(8, 12)
	company.GeneratorCounts["generator.beige_tower"] = 2
	preState, err := save.EncodeState(company)
	if err != nil {
		t.Fatal(err)
	}
	request, err := ParseIntent([]byte(`{"intent_id":"01986666-0100-7000-8000-000000000100","kind":"wind_down","expected_revision":1,"expected_founder_revision":1}`))
	if err != nil {
		t.Fatal(err)
	}
	command := save.ReplayCommand{IntentID: request.IntentID, CompanyStreamID: "01986666-1000-7000-8000-000000000001", FounderID: "01986666-2000-7000-8000-000000000001", Revision: 1, RunSeq: 1, RunLogSeq: 100}
	carry := replayFounderCarry{FounderRevision: 1, FounderConstantsHash: constantsHash, NetworkSlots: []save.NetworkSlot{}, LedgerFactKinds: []string{}, ExitHistoryCount: 0}
	inputs, err := buildReplayInputs(replayBuild{Command: command, Mode: ModeOnline, Now: now, IntentKind: request.Kind, RouteContextVersion: catalogs.Routes.ContextVersion(), FounderCarry: &carry, Terminal: true, ExecutedRouteIDs: []string{}, SelectedExitType: "scripted_first", SelectedTerms: json.RawMessage(`{}`), NextConstantsHash: constantsHash})
	if err != nil {
		t.Fatal(err)
	}
	transition, err := ApplyLoggedExit(company, request.CanonicalPayload, catalogs, inputs)
	if err != nil {
		t.Fatalf("wind-down-scripted-first transition: %v", err)
	}
	finalCompany, err := save.EncodeState(transition.Company)
	if err != nil {
		t.Fatal(err)
	}
	newCompany, err := save.EncodeState(transition.Decision.NewCompanyState)
	if err != nil {
		t.Fatal(err)
	}
	return withTerminalRawJSON(t, crossRuntimeTerminalCase{Name: "wind-down-scripted-first", PreState: preState, CanonicalPayload: request.CanonicalPayload, ReplayInputs: inputs,
		Outcome: string(transition.Decision.Outcome), Receipt: transition.Decision.Receipt, FounderOutput: replayFounderOutput(transition.Founder, carry),
		FinalCompany: finalCompany, NewCompany: newCompany, FounderEvents: fixtureEvents(transition.Decision.FounderEvents),
		CompanyEndedEvents: fixtureEvents(transition.Decision.CompanyEndedEvents), CompanyStartedEvents: fixtureEvents(transition.Decision.CompanyStartedEvents)})
}

func makeAcceptedOfferFixtureCase(t *testing.T, catalogs CatalogBundle, constantsHash string, now time.Time) crossRuntimeTerminalCase {
	t.Helper()
	company := replayFixtureState(t, catalogs.Economy, now.Add(-20*time.Minute))
	company.Tier = 3
	company.LifetimeValue = decimal.New(27, 12)
	terms := json.RawMessage(`{"market_modifier_ppm":1100000,"payout_preview":{"reputation_delta":5,"network_slot_unlocks":[],"route_knowledge":0,"clout_reach_note":"clout.reach.preserved"}}`)
	company.OfferState = &save.ExitOfferState{OfferID: "01986666-0200-7000-8000-000000000200", ExitType: "acquisition", TermsJSON: terms, SpawnedAt: now.Add(-time.Minute), ExpiresAt: now.Add(time.Minute)}
	preState, err := save.EncodeState(company)
	if err != nil {
		t.Fatal(err)
	}
	request, err := ParseIntent([]byte(`{"intent_id":"01986666-0200-7000-8000-000000000201","kind":"accept_exit_offer","expected_revision":1,"expected_founder_revision":2,"offer_id":"01986666-0200-7000-8000-000000000200"}`))
	if err != nil {
		t.Fatal(err)
	}
	command := save.ReplayCommand{IntentID: request.IntentID, CompanyStreamID: "01986666-1000-7000-8000-000000000001", FounderID: "01986666-2000-7000-8000-000000000001", Revision: 1, RunSeq: 1, RunLogSeq: 101}
	carry := replayFounderCarry{FounderRevision: 2, FounderConstantsHash: constantsHash, ReputationLevel: 1, NetworkSlots: []save.NetworkSlot{}, LedgerFactKinds: []string{}, ExitHistoryCount: 1}
	inputs, err := buildReplayInputs(replayBuild{Command: command, Mode: ModeOnline, Now: now, IntentKind: request.Kind, RouteContextVersion: catalogs.Routes.ContextVersion(), FounderCarry: &carry, Terminal: true, ExecutedRouteIDs: []string{}, SelectedExitType: "acquisition", SelectedTerms: terms, NextConstantsHash: constantsHash})
	if err != nil {
		t.Fatal(err)
	}
	return executeTerminalFixture(t, "accept-stored-offer", catalogs, company, preState, request, inputs, carry)
}

func makeScriptedGateFixtureCase(t *testing.T, catalogs CatalogBundle, constantsHash string, now time.Time) crossRuntimeTerminalCase {
	t.Helper()
	company := replayFixtureState(t, catalogs.Economy, now.Add(-20*time.Minute))
	company.EvaluatedThrough = now
	company.ManualTokenRefilledAt = now
	company.Tier = 2
	company.LifetimeValue = decimal.New(8, 12)
	setCash(t, company, "1e10")
	preState, err := save.EncodeState(company)
	if err != nil {
		t.Fatal(err)
	}
	request, err := ParseIntent([]byte(`{"intent_id":"01986666-0300-7000-8000-000000000300","kind":"cross_gate","expected_revision":1,"gate_id":"gate.t2_to_t3","route_id":null}`))
	if err != nil {
		t.Fatal(err)
	}
	command := save.ReplayCommand{IntentID: request.IntentID, CompanyStreamID: "01986666-1000-7000-8000-000000000001", FounderID: "01986666-2000-7000-8000-000000000001", Revision: 1, RunSeq: 1, RunLogSeq: 102}
	carry := replayFounderCarry{FounderRevision: 1, FounderConstantsHash: constantsHash, NetworkSlots: []save.NetworkSlot{}, LedgerFactKinds: []string{}, ExitHistoryCount: 0}
	inputs, err := buildReplayInputs(replayBuild{Command: command, Mode: ModeOnline, Now: now, IntentKind: request.Kind, RouteContextVersion: catalogs.Routes.ContextVersion(), FounderCarry: &carry, Terminal: true, ExecutedRouteIDs: []string{}, SelectedExitType: "scripted_first", SelectedTerms: json.RawMessage(`{}`), NextConstantsHash: constantsHash})
	if err != nil {
		t.Fatal(err)
	}
	return executeTerminalFixture(t, "scripted-cross-gate", catalogs, company, preState, request, inputs, carry)
}

func executeTerminalFixture(t *testing.T, name string, catalogs CatalogBundle, company *save.State, preState json.RawMessage, request IntentRequest, inputs json.RawMessage, carry replayFounderCarry) crossRuntimeTerminalCase {
	t.Helper()
	transition, err := ApplyLoggedExit(company, request.CanonicalPayload, catalogs, inputs)
	if err != nil {
		t.Fatalf("%s transition: %v", name, err)
	}
	finalCompany, err := save.EncodeState(transition.Company)
	if err != nil {
		t.Fatal(err)
	}
	newCompany, err := save.EncodeState(transition.Decision.NewCompanyState)
	if err != nil {
		t.Fatal(err)
	}
	return withTerminalRawJSON(t, crossRuntimeTerminalCase{Name: name, PreState: preState, CanonicalPayload: request.CanonicalPayload, ReplayInputs: inputs,
		Outcome: string(transition.Decision.Outcome), Receipt: transition.Decision.Receipt, FounderOutput: replayFounderOutput(transition.Founder, carry),
		FinalCompany: finalCompany, NewCompany: newCompany, FounderEvents: fixtureEvents(transition.Decision.FounderEvents),
		CompanyEndedEvents: fixtureEvents(transition.Decision.CompanyEndedEvents), CompanyStartedEvents: fixtureEvents(transition.Decision.CompanyStartedEvents)})
}

func withTerminalRawJSON(t *testing.T, value crossRuntimeTerminalCase) crossRuntimeTerminalCase {
	t.Helper()
	value.ReceiptJSON = canonicalFixtureJSON(t, value.Receipt)
	value.FounderOutputJSON = canonicalFixtureValue(t, value.FounderOutput)
	value.FinalCompanyJSON = canonicalFixtureJSON(t, value.FinalCompany)
	value.NewCompanyJSON = canonicalFixtureJSON(t, value.NewCompany)
	value.FounderEventsJSON = canonicalFixtureValue(t, value.FounderEvents)
	value.CompanyEndedJSON = canonicalFixtureValue(t, value.CompanyEndedEvents)
	value.CompanyStartedJSON = canonicalFixtureValue(t, value.CompanyStartedEvents)
	return value
}

func canonicalFixtureJSON(t *testing.T, raw json.RawMessage) string {
	t.Helper()
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err != nil {
		t.Fatal(err)
	}
	return canonicalFixtureValue(t, value)
}

func canonicalFixtureValue(t *testing.T, value any) string {
	t.Helper()
	encoded, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	decoder := json.NewDecoder(bytes.NewReader(encoded))
	decoder.UseNumber()
	var canonical any
	if err := decoder.Decode(&canonical); err != nil {
		t.Fatal(err)
	}
	encoded, err = json.Marshal(canonical)
	if err != nil {
		t.Fatal(err)
	}
	return string(encoded)
}

func replayFounderOutput(state *save.State, carry replayFounderCarry) any {
	facts := make([]string, 0, len(state.LedgerFactKinds))
	for fact := range state.LedgerFactKinds {
		facts = append(facts, fact)
	}
	sort.Strings(facts)
	return map[string]any{"founder_revision": carry.FounderRevision, "founder_constants_hash": carry.FounderConstantsHash,
		"reputation_level": state.ReputationLevel, "route_knowledge_balance": state.RouteKnowledgeBalance, "age_ms": state.AgeMS,
		"notoriety": state.Notoriety, "advisor_mode": state.AdvisorMode, "network_slots": state.NetworkSlots,
		"ledger_fact_kinds": facts, "exit_history_count": len(state.ExitHistory)}
}

func fixtureEvents(events []save.EventWrite) []fixtureEvent {
	result := make([]fixtureEvent, len(events))
	for index, event := range events {
		result[index] = fixtureEvent{Kind: string(event.Kind), SchemaVersion: event.SchemaVersion, IntentID: event.IntentID, Payload: event.Payload}
	}
	return result
}

func replayFixtureState(t *testing.T, catalog *economy.Catalog, now time.Time) *save.State {
	t.Helper()
	ledger, err := economy.NewLedger(catalog, economy.ScopeCompany)
	if err != nil {
		t.Fatal(err)
	}
	return &save.State{Ledger: ledger, GeneratorCounts: map[string]int64{"generator.beige_tower": 0}, EvaluatedThrough: now,
		ManualTokenMilli: 50_000, ManualTokenRefilledAt: now, GatesCrossed: map[string]bool{}, RunSeq: 1,
		DoctrinesByTransition: map[string]string{}, LedgerFactKinds: map[string]bool{}, MeterBands: map[string]int{}, RegionTraits: map[string]bool{},
		HintsUnlocked: map[string]bool{}, CompactSamples: []save.CompactSample{}, LifetimeValue: decimal.Zero, RunStartedAt: now,
		OfflineSpans: []save.OfflineSpan{}, NetworkSlots: []save.NetworkSlot{}, ExitHistory: []save.ExitRecord{}}
}

func setCash(t *testing.T, state *save.State, amount string) {
	t.Helper()
	value, err := decimal.ParseCanonical(amount)
	if err != nil {
		t.Fatal(err)
	}
	current, _ := state.Ledger.Balance("company.cash")
	if _, err := state.Ledger.Apply(economy.Transaction{Entries: []economy.Entry{{ResourceID: "company.cash", Delta: value.Sub(current)}}}); err != nil {
		t.Fatal(err)
	}
}

func int64Pointer(value int64) *int64 { return &value }

func offerFixtureFounder(t *testing.T, threshold int64) string {
	t.Helper()
	for candidate := int64(1); candidate <= 1_000_000; candidate++ {
		founderID := fmt.Sprintf("01986666-2000-7000-8000-%012d", candidate)
		spawn, _, _ := prestigecore.OfferDraws(founderID, 1, 3, 0)
		if spawn < threshold {
			return founderID
		}
	}
	t.Fatal("could not find deterministic offer-spawn founder")
	return ""
}

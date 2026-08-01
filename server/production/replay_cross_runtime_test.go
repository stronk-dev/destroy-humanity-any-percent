package production

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
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
		{name: "closed-hook-chain", payload: `{"intent_id":"01986666-0009-7000-8000-000000000009","kind":"perform_manual_batch","expected_revision":1,"action_id":"manual.click","count":1,"window_ms":60000}`, advance: time.Minute, mode: ModeOnline,
			configure: func(state *save.State) {
				state.GeneratorCounts["generator.beige_tower"] = 2
				state.FactionID = "open_source"
				state.IncorporatedAt = baseNow.Add(-time.Minute)
				state.CompactMember, state.CompactTithePPM = true, 130_000
			},
			contributions: []multiplier.Contribution{{Slot: multiplier.SlotFaction, SourceID: "guild.stock_consumption", Target: "all", Factor: decimal.One}, {Slot: multiplier.SlotCommons, SourceID: "commons.member", Target: "all", Factor: decimal.New(11, -1)}}, weight: int64Pointer(812_345)},
		{name: "invalid-manual", payload: `{"intent_id":"01986666-0010-7000-8000-000000000010","kind":"perform_manual_batch","expected_revision":1,"action_id":"manual.click","count":0,"window_ms":1}`, mode: ModeOnline},
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
			ReplayInputs: inputs, Outcome: string(transition.Outcome), Receipt: transition.Receipt, Events: events, PostState: postState})
	}
	result.TerminalCases = []crossRuntimeTerminalCase{
		makeTerminalFixtureCase(t, catalogs, bundleBytes.Hash, baseNow),
		makeAcceptedOfferFixtureCase(t, catalogs, bundleBytes.Hash, baseNow),
		makeScriptedGateFixtureCase(t, catalogs, bundleBytes.Hash, baseNow),
	}
	return result
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
	return crossRuntimeTerminalCase{Name: "wind-down-scripted-first", PreState: preState, CanonicalPayload: request.CanonicalPayload, ReplayInputs: inputs,
		Outcome: string(transition.Decision.Outcome), Receipt: transition.Decision.Receipt, FounderOutput: replayFounderOutput(transition.Founder, carry),
		FinalCompany: finalCompany, NewCompany: newCompany, FounderEvents: fixtureEvents(transition.Decision.FounderEvents),
		CompanyEndedEvents: fixtureEvents(transition.Decision.CompanyEndedEvents), CompanyStartedEvents: fixtureEvents(transition.Decision.CompanyStartedEvents)}
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
	return crossRuntimeTerminalCase{Name: name, PreState: preState, CanonicalPayload: request.CanonicalPayload, ReplayInputs: inputs,
		Outcome: string(transition.Decision.Outcome), Receipt: transition.Decision.Receipt, FounderOutput: replayFounderOutput(transition.Founder, carry),
		FinalCompany: finalCompany, NewCompany: newCompany, FounderEvents: fixtureEvents(transition.Decision.FounderEvents),
		CompanyEndedEvents: fixtureEvents(transition.Decision.CompanyEndedEvents), CompanyStartedEvents: fixtureEvents(transition.Decision.CompanyStartedEvents)}
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

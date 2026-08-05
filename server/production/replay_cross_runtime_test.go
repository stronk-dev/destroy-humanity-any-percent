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

	"cloud-clicker/server/achievements"
	"cloud-clicker/server/decimal"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/epochseed"
	"cloud-clicker/server/guild"
	"cloud-clicker/server/meters"
	"cloud-clicker/server/minigame"
	"cloud-clicker/server/multiplier"
	"cloud-clicker/server/pet"
	prestigecore "cloud-clicker/server/prestige"
	"cloud-clicker/server/save"
)

var updateReplayFixture = flag.Bool("update-replay-fixture", false, "rewrite the shared ApplyLogged fixture")

type crossRuntimeFixture struct {
	Version         int                        `json:"version"`
	ConstantsHash   string                     `json:"constants_hash"`
	Artifacts       map[string]string          `json:"artifacts"`
	Cases           []crossRuntimeFixtureCase  `json:"cases"`
	TerminalCases   []crossRuntimeTerminalCase `json:"terminal_cases"`
	Additional      []crossRuntimeBundleCase   `json:"additional_bundles"`
	ActiveExit      crossRuntimeActiveExit     `json:"active_foundation_exit"`
	FullRun         crossRuntimeFullRun        `json:"full_run"`
	RejectedExit    crossRuntimeFullRun        `json:"rejected_exit_run"`
	FounderHash     string                     `json:"founder_constants_hash"`
	FounderFiles    map[string]string          `json:"founder_artifacts"`
	FounderCases    []crossRuntimeFounderCase  `json:"founder_cases"`
	FounderRun      crossRuntimeFounderRun     `json:"founder_run"`
	PetFounderHash  string                     `json:"pet_founder_constants_hash"`
	PetFounderFiles map[string]string          `json:"pet_founder_artifacts"`
	PetFounderCases []crossRuntimeFounderCase  `json:"pet_founder_cases"`
	MinigameHash    string                     `json:"minigame_constants_hash"`
	MinigameFiles   map[string]string          `json:"minigame_artifacts"`
	MinigameCompany crossRuntimeFixtureCase    `json:"minigame_company_case"`
	MinigameFounder crossRuntimeFounderCase    `json:"minigame_founder_case"`
}

type crossRuntimeFounderCase struct {
	Name                string          `json:"name"`
	StateVersion        int             `json:"state_version"`
	PreState            json.RawMessage `json:"pre_state"`
	CanonicalPayload    json.RawMessage `json:"canonical_payload"`
	ReplayInputs        json.RawMessage `json:"replay_inputs"`
	Outcome             string          `json:"outcome"`
	Receipt             json.RawMessage `json:"receipt"`
	Events              []fixtureEvent  `json:"events"`
	PostState           json.RawMessage `json:"post_state"`
	ResultConstantsHash string          `json:"result_constants_hash"`
	ReceiptJSON         string          `json:"receipt_json"`
	EventsJSON          string          `json:"events_json"`
	PostStateJSON       string          `json:"post_state_json"`
}

type crossRuntimeFounderRun struct {
	FounderStreamID string                        `json:"founder_stream_id"`
	FounderID       string                        `json:"founder_id"`
	GenesisRevision int64                         `json:"genesis_revision"`
	GenesisVersion  int                           `json:"genesis_version"`
	GenesisHash     string                        `json:"genesis_constants_hash"`
	Genesis         json.RawMessage               `json:"genesis"`
	Entries         []crossRuntimeFounderRunEntry `json:"entries"`
	HeadRevision    int64                         `json:"head_revision"`
	HeadVersion     int                           `json:"head_version"`
	HeadHash        string                        `json:"head_constants_hash"`
	HeadState       json.RawMessage               `json:"head_state"`
}

type crossRuntimeFounderRunEntry struct {
	Sequence         int64                  `json:"seq"`
	IntentID         string                 `json:"intent_id"`
	ConstantsHash    string                 `json:"constants_hash"`
	CanonicalPayload json.RawMessage        `json:"canonical_payload"`
	ReplayInputs     json.RawMessage        `json:"replay_inputs"`
	ReceiptJSON      string                 `json:"receipt_json"`
	EventsJSON       string                 `json:"events_json"`
	AppliedRevision  *int64                 `json:"applied_revision"`
	ServerTSMS       int64                  `json:"server_ts_ms"`
	Source           *save.FounderLogSource `json:"source"`
}

type crossRuntimeActiveExit struct {
	ConstantsHash     string                   `json:"constants_hash"`
	Artifacts         map[string]string        `json:"artifacts"`
	NextConstantsHash string                   `json:"next_constants_hash"`
	NextArtifacts     map[string]string        `json:"next_artifacts"`
	Case              crossRuntimeTerminalCase `json:"case"`
}

type crossRuntimeFullRun struct {
	ConstantsHash  string                     `json:"constants_hash"`
	Artifacts      map[string]string          `json:"artifacts"`
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
		{name: "purchase-total-cap-precedes-affordability", payload: `{"intent_id":"01986666-0018-7000-8000-000000000018","kind":"buy_generator","expected_revision":1,"generator_id":"generator.beige_tower","count":{"mode":"exact","value":1}}`, mode: ModeOnline,
			configure: func(state *save.State) { state.GeneratorPurchasedTotal = decimal.MaxExactInteger }},
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
	result.Additional = append([]crossRuntimeBundleCase{makeFallbackInvariantFixture(t, bundleBytes.Artifacts, baseNow)}, makeFoundationFixtures(t, bundleBytes.Artifacts, baseNow)...)
	result.Additional = append(result.Additional,
		makeActiveFoundationReplayFixture(t, baseNow, 5001*time.Millisecond, "active-foundation-offline-5001ms"),
		makeActiveFoundationReplayFixture(t, baseNow, 25*time.Hour, "active-foundation-offline-25h"),
		makeActiveFoundationBandReplayFixture(t, baseNow),
		makeActiveFoundationBurnReplayFixture(t, baseNow),
	)
	result.ActiveExit = makeActiveFoundationExitFixture(t, baseNow)
	result.FullRun = makeFullRunFixture(t, catalogs, bundleBytes.Hash, baseNow)
	result.RejectedExit = makeRejectedExitFixture(t, catalogs, bundleBytes.Hash, baseNow)
	result.FounderHash, result.FounderFiles, result.FounderCases, result.FounderRun = makeFounderReplayFixture(t, baseNow)
	result.PetFounderHash, result.PetFounderFiles, result.PetFounderCases = makePetFounderReplayFixture(t, baseNow)
	result.MinigameHash, result.MinigameFiles, result.MinigameCompany, result.MinigameFounder = makeMinigameResolutionReplayFixture(t, baseNow)
	return result
}

func makeMinigameResolutionReplayFixture(t *testing.T, now time.Time) (string, map[string]string, crossRuntimeFixtureCase, crossRuntimeFounderCase) {
	t.Helper()
	_, active := foundationTestBundles(t)
	artifact, err := os.ReadFile("../../testdata/minigame/catalog-v2.json")
	if err != nil {
		t.Fatal(err)
	}
	artifacts := cloneArtifactMap(active.Artifacts)
	artifacts["minigames"] = artifact
	hash, err := save.ConstantsHashArtifacts(artifacts)
	if err != nil {
		t.Fatal(err)
	}
	catalogs := active
	catalogs.ConstantsHash, catalogs.Artifacts = hash, artifacts
	catalogs.Minigames, err = minigame.LoadCatalog(artifact)
	if err != nil || !catalogs.valid(hash) {
		t.Fatalf("load minigame replay bundle: %v", err)
	}
	result := minigame.Result{Outcome: "completed", RatingDelta: int64Pointer(25), ScoreFacts: []minigame.ScoreFact{{Kind: "score.total", Value: 400}}}
	resultBytes, _ := json.Marshal(result)
	payload, _ := json.Marshal(minigameResolutionPayload{Kind: minigameResolutionKind, SessionID: "01986666-0900-7000-8000-000000000901", Result: resultBytes})
	payload, _ = normalizeReplayJSON(payload)
	definition, _ := catalogs.Minigames.Definition("fixture.counter")
	attendance := FounderAttendanceSample{CompanyStreamID: "01986666-1900-7000-8000-000000000901", RunSeq: 1, CompanyRevision: 1,
		CompanyConstantsHash: hash, CompletedAttendedMS: 0, CurrentRunPartialAttendedMS: 60_000, EffectiveFounderAttendedMS: 60_000}
	oldRating := save.MinigameRatingState{Elo: 1000, SeasonMember: "ranked", GamesCounted: 0}
	newRating := save.MinigameRatingState{Elo: 1025, SeasonMember: "ranked", GamesCounted: 1}
	oldQuality := save.MinigameOfflineQualityState{GradePPM: 500_000, LastFounderAttendedMS: 0, DecayRemainderPPM: 0}
	newQuality := save.MinigameOfflineQualityState{GradePPM: 750_000, LastFounderAttendedMS: 60_000, DecayRemainderPPM: 0}
	rating := minigameRatingChangeReceipt{Rated: true, OldElo: 1000, NewElo: 1025, SeasonMember: "ranked", GamesBefore: 0, GamesAfter: 1}
	quality := minigameQualityChangeReceipt{Old: oldQuality, New: newQuality}
	certifiedHash := certifiedResultHash(resultBytes)
	faucet := minigameFaucetWire{AttendedDay: 0, QuotaBefore: 0, QuotaAfter: 1, RemainderBeforePPM: 0,
		RemainderAfterPPM: 0, ReducedScore: 400, ConvertedUnits: 50, CreditedUnits: 50}
	companyResolved := minigameCompanyResolved{Kind: minigameResolutionKind, SessionID: "01986666-0900-7000-8000-000000000901",
		MinigameID: "fixture.counter", CertifiedResultHash: certifiedHash, PayoutPolicy: definition.Payout, SelectedScore: 400,
		Faucet: faucet, CreditedDelta: "5e1", FounderLog: minigameLogCoordinate{StreamID: "01986666-2900-7000-8000-000000000901", Revision: 2, Sequence: 1},
		CompanyRevision: 2, FounderRevision: 2, RatingChange: rating, QualityChange: quality}
	companyCommand := save.ReplayCommand{IntentID: "01986666-0900-7000-8000-000000000901", CompanyStreamID: attendance.CompanyStreamID,
		FounderID: "01986666-3900-7000-8000-000000000901", Revision: 1, RunSeq: 1, RunLogSeq: 1}
	companyInputs, _ := json.Marshal(replayInputsWire{Version: save.ReplayInputsVersion, Command: companyCommand, EvaluatedAtMS: now.UnixMilli(), EvaluationMode: ModeOnline, Resolved: mustJSON(companyResolved)})
	company := replayFixtureState(t, catalogs.Economy, now)
	company.WireVersion = 16
	company.MeterBands = nil
	meterState, err := meters.NewRunState(catalogs.Meters, 0)
	if err != nil {
		t.Fatal(err)
	}
	company.MeterValues, company.MeterDecayRemainders, company.MeterInputRemainders = meterState.Values, meterState.DecayRemainders, meterState.InputRemainders
	company.AchievementsEarnedRun = map[string]bool{}
	companyPre := mustEncodeState(t, company)
	companyTransition, err := applyCompanyMinigameResolution(company, payload, catalogs, replayInputsWire{Version: save.ReplayInputsVersion,
		Command: companyCommand, EvaluatedAtMS: now.UnixMilli(), EvaluationMode: ModeOnline, Resolved: mustJSON(companyResolved)})
	if err != nil {
		t.Fatal(err)
	}
	companyPost := mustEncodeState(t, company)
	companyEvents := fixtureEvents(companyTransition.Events)
	companyCase := crossRuntimeFixtureCase{Name: "minigame-resolution-company", PreState: companyPre, CanonicalPayload: payload, ReplayInputs: companyInputs,
		Outcome: string(companyTransition.Outcome), Receipt: companyTransition.Receipt, Events: companyEvents, PostState: companyPost,
		ReceiptJSON: canonicalFixtureJSON(t, companyTransition.Receipt), EventsJSON: canonicalFixtureValue(t, companyEvents), PostStateJSON: canonicalFixtureJSON(t, companyPost)}
	founder := replayFounderFixtureState(t, catalogs, now)
	founder.MinigameRatings["fixture.counter"] = oldRating
	founder.MinigameOfflineQuality["fixture.counter"] = oldQuality
	founderPre := mustEncodeState(t, founder)
	founderCommand := save.FounderReplayCommand{IntentID: companyCommand.IntentID, FounderStreamID: companyResolved.FounderLog.StreamID,
		FounderID: companyCommand.FounderID, Revision: 1, FounderLogSeq: 1, ServerTSMS: now.UnixMilli()}
	founderResolved := minigameFounderResolved{Kind: minigameResolutionKind, SessionID: companyCommand.IntentID, MinigameID: "fixture.counter",
		CertifiedResultHash: certifiedHash, RatingBefore: oldRating, RatingAfter: newRating, QualityBefore: oldQuality, QualityAfter: newQuality, Attendance: attendance}
	founderInputs, _ := save.MarshalFounderReplayInputs(founderCommand, founderResolved)
	founderTransition, err := applyFounderMinigameResolution(founder, payload, catalogs, founderReplayInputsWire{Version: 1, Command: founderCommand, EvaluatedAtMS: now.UnixMilli(), Resolved: mustJSON(founderResolved)})
	if err != nil {
		t.Fatal(err)
	}
	founderPost := mustEncodeState(t, founder)
	founderEvents := fixtureEvents(founderTransition.Events)
	founderCase := crossRuntimeFounderCase{Name: "minigame-resolution-founder", StateVersion: 17, PreState: founderPre, CanonicalPayload: payload,
		ReplayInputs: founderInputs, Outcome: string(founderTransition.Outcome), Receipt: founderTransition.Receipt, Events: founderEvents, PostState: founderPost,
		ResultConstantsHash: founderTransition.ResultConstantsHash, ReceiptJSON: canonicalFixtureJSON(t, founderTransition.Receipt),
		EventsJSON: canonicalFixtureValue(t, founderEvents), PostStateJSON: canonicalFixtureJSON(t, founderPost)}
	return hash, stringArtifacts(artifacts), companyCase, founderCase
}

func makeFounderReplayFixture(t *testing.T, now time.Time) (string, map[string]string, []crossRuntimeFounderCase, crossRuntimeFounderRun) {
	t.Helper()
	_, catalogs := foundationTestBundles(t)
	definitions := []struct {
		name      string
		payload   string
		configure func(*save.State)
		resolved  func(*save.State) any
	}{
		{name: "invalid-route-hint", payload: `{"intent_id":"01986666-0700-7000-8000-000000000701","kind":"buy_route_hint","expected_revision":1,"route_id":"INVALID"}`,
			resolved: func(*save.State) any { return founderInvalidResolved{Kind: "invalid", Detail: "route_id"} }},
		{name: "route-hint-applied", payload: `{"intent_id":"01986666-0700-7000-8000-000000000702","kind":"buy_route_hint","expected_revision":1,"route_id":"route.ipo_sequence_break"}`,
			configure: func(state *save.State) { state.RouteKnowledgeBalance = 125 },
			resolved: func(state *save.State) any {
				return founderRouteHintResolved{Kind: string(IntentBuyRouteHint), RouteContextVersion: catalogs.Routes.ContextVersion(), RouteKnowledgeBalance: state.RouteKnowledgeBalance}
			}},
		{name: "exit-rejected", payload: `{"intent_id":"01986666-0700-7000-8000-000000000703","kind":"wind_down","expected_revision":1,"expected_founder_revision":1}`,
			configure: func(state *save.State) { state.AgeMS = 100 },
			resolved: func(state *save.State) any {
				return founderExitResolvedWire{Kind: founderExitResolvedKind, Outcome: string(save.IntentRejected), CompanyStreamID: "01986666-1700-7000-8000-000000000703", RunSeq: 1, RunLogSeq: 3, ResultConstantsHash: catalogs.ConstantsHash, AgeMSBefore: state.AgeMS, AgeMSAfter: state.AgeMS, AddedNetworkSlots: []save.NetworkSlot{}, AddedLedgerFactKinds: []string{}, AddedLifetimeAchievements: []string{}, ResultFounderWireVersion: save.VersionForState(state), Rejection: &founderAuditRejection{Category: "not_eligible", Detail: "tier"}}
			}},
		{name: "exit-applied", payload: `{"intent_id":"01986666-0700-7000-8000-000000000704","kind":"wind_down","expected_revision":1,"expected_founder_revision":1}`,
			configure: func(state *save.State) { state.RouteKnowledgeBalance, state.AgeMS = 75, 100 },
			resolved: func(state *save.State) any {
				return founderExitResolvedWire{Kind: founderExitResolvedKind, Outcome: string(save.IntentApplied), CompanyStreamID: "01986666-1700-7000-8000-000000000704", RunSeq: 1, RunLogSeq: 4, ResultConstantsHash: catalogs.ConstantsHash, ReputationDelta: 3, RouteKnowledgeDelta: 5, AttendedMS: 600_000, AgeMSBefore: state.AgeMS, AgeMSAfter: state.AgeMS + 600_000, AddedNetworkSlots: []save.NetworkSlot{{Slot: "network.board", CarriedRef: "contact.board"}}, AddedLedgerFactKinds: []string{"exit.collapse"}, AddedLifetimeAchievements: []string{}, ExitRecord: &founderExitRecordWire{RunID: 1, ExitType: "collapse", OccurredAtMS: now.UnixMilli(), ReputationDelta: 3}, ResultFounderWireVersion: save.VersionForState(state)}
			}},
	}
	cases := make([]crossRuntimeFounderCase, 0, len(definitions))
	for index, definition := range definitions {
		state := replayFounderFixtureState(t, catalogs, now)
		if definition.configure != nil {
			definition.configure(state)
		}
		pre := mustEncodeState(t, state)
		request, err := ParseIntent([]byte(definition.payload))
		if err != nil {
			t.Fatalf("%s parse: %v", definition.name, err)
		}
		command := save.FounderReplayCommand{IntentID: request.IntentID, FounderStreamID: "01986666-2700-7000-8000-000000000001", FounderID: "01986666-3700-7000-8000-000000000001", Revision: request.ExpectedRevision, FounderLogSeq: int64(index + 1), ServerTSMS: now.Add(time.Duration(index) * time.Second).UnixMilli()}
		inputs, err := save.MarshalFounderReplayInputs(command, definition.resolved(state))
		if err != nil {
			t.Fatalf("%s inputs: %v", definition.name, err)
		}
		transition, err := ApplyFounderLogged(state, request.CanonicalPayload, catalogs, inputs)
		if err != nil {
			t.Fatalf("%s transition: %v", definition.name, err)
		}
		post := mustEncodeState(t, state)
		events := fixtureEvents(transition.Events)
		cases = append(cases, crossRuntimeFounderCase{Name: definition.name, StateVersion: save.VersionForState(state), PreState: pre, CanonicalPayload: request.CanonicalPayload, ReplayInputs: inputs,
			Outcome: string(transition.Outcome), Receipt: transition.Receipt, Events: events, PostState: post, ResultConstantsHash: transition.ResultConstantsHash,
			ReceiptJSON: canonicalFixtureJSON(t, transition.Receipt), EventsJSON: canonicalFixtureValue(t, events), PostStateJSON: canonicalFixtureJSON(t, post)})
	}
	return catalogs.ConstantsHash, stringArtifacts(catalogs.Artifacts), cases, makeFounderFullRunFixture(t, catalogs, now)
}

func makePetFounderReplayFixture(t *testing.T, now time.Time) (string, map[string]string, []crossRuntimeFounderCase) {
	t.Helper()
	_, active := foundationTestBundles(t)
	_, catalogs := founderFeatureBundles(t, active)
	state := replayFounderFixtureState(t, catalogs, now)
	const petID = "01986666-7700-7000-8000-000000000001"
	state.Pets[petID] = pet.CareState{
		StatsPPM:                map[pet.StatID]int64{pet.StatHunger: 600_000, pet.StatEnergy: 800_000, pet.StatCleanliness: 800_000, pet.StatAffection: 800_000},
		StatDecayRemaindersPPM:  map[pet.StatID]int64{pet.StatHunger: 0, pet.StatEnergy: 0, pet.StatCleanliness: 0, pet.StatAffection: 0},
		CooldownUntilAttendedMS: map[string]int64{"care.feed": 0}, TrustPPM: 500_000,
		BehaviorState: pet.BehaviorIdle, BehaviorQueue: []pet.BehaviorQueueEntry{},
	}
	pre := mustEncodeState(t, state)
	request, err := ParseIntent([]byte(`{"intent_id":"01986666-0700-7000-8000-000000000705","kind":"care_action","expected_revision":1,"pet_id":"01986666-7700-7000-8000-000000000001","action_id":"care.feed"}`))
	if err != nil {
		t.Fatal(err)
	}
	command := save.FounderReplayCommand{IntentID: request.IntentID, FounderStreamID: "01986666-2700-7000-8000-000000000005",
		FounderID: "01986666-3700-7000-8000-000000000005", Revision: 1, FounderLogSeq: 1, ServerTSMS: now.UnixMilli()}
	inputs, err := save.MarshalFounderReplayInputs(command, founderCareResolved{Kind: IntentCareAction, PetAttendedBeforeMS: 0,
		Attendance: FounderAttendanceSample{CompanyStreamID: "01986666-1700-7000-8000-000000000705", RunSeq: 1,
			CompanyRevision: 1, CompanyConstantsHash: catalogs.ConstantsHash, CompletedAttendedMS: 0,
			CurrentRunPartialAttendedMS: 120_000, EffectiveFounderAttendedMS: 120_000}})
	if err != nil {
		t.Fatal(err)
	}
	transition, err := ApplyFounderLogged(state, request.CanonicalPayload, catalogs, inputs)
	if err != nil {
		t.Fatal(err)
	}
	post := mustEncodeState(t, state)
	events := fixtureEvents(transition.Events)
	return catalogs.ConstantsHash, stringArtifacts(catalogs.Artifacts), []crossRuntimeFounderCase{{Name: "care-action-applied",
		StateVersion: 18, PreState: pre, CanonicalPayload: request.CanonicalPayload, ReplayInputs: inputs,
		Outcome: string(transition.Outcome), Receipt: transition.Receipt, Events: events, PostState: post,
		ResultConstantsHash: transition.ResultConstantsHash, ReceiptJSON: canonicalFixtureJSON(t, transition.Receipt),
		EventsJSON: canonicalFixtureValue(t, events), PostStateJSON: canonicalFixtureJSON(t, post)}}
}

func makeFounderFullRunFixture(t *testing.T, catalogs CatalogBundle, now time.Time) crossRuntimeFounderRun {
	t.Helper()
	const streamID = "01986666-2800-7000-8000-000000000001"
	const founderID = "01986666-3800-7000-8000-000000000001"
	state := replayFounderFixtureState(t, catalogs, now)
	state.RouteKnowledgeBalance = 125
	genesis := mustEncodeState(t, state)
	history := save.FounderHistory{FounderStreamID: streamID, FounderID: founderID,
		Genesis: save.FounderGenesis{FounderStreamID: streamID, Revision: 1, State: genesis, Version: save.VersionForState(state), ConstantsHash: catalogs.ConstantsHash}, Entries: []save.FounderHistoryEntry{}}
	result := crossRuntimeFounderRun{FounderStreamID: streamID, FounderID: founderID, GenesisRevision: 1,
		GenesisVersion: save.VersionForState(state), GenesisHash: catalogs.ConstantsHash, Genesis: genesis, Entries: []crossRuntimeFounderRunEntry{}}
	definitions := []struct {
		payload  string
		resolved func(*save.State) any
		source   *save.FounderLogSource
	}{
		{payload: `{"intent_id":"01986666-0800-7000-8000-000000000801","kind":"buy_route_hint","expected_revision":1,"route_id":"route.ipo_sequence_break"}`,
			resolved: func(state *save.State) any {
				return founderRouteHintResolved{Kind: string(IntentBuyRouteHint), RouteContextVersion: catalogs.Routes.ContextVersion(), RouteKnowledgeBalance: state.RouteKnowledgeBalance}
			}},
		{payload: `{"intent_id":"01986666-0800-7000-8000-000000000802","kind":"wind_down","expected_revision":1,"expected_founder_revision":2}`,
			resolved: func(state *save.State) any {
				return founderExitResolvedWire{Kind: founderExitResolvedKind, Outcome: string(save.IntentRejected), CompanyStreamID: "01986666-1800-7000-8000-000000000001", RunSeq: 1, RunLogSeq: 2, ResultConstantsHash: catalogs.ConstantsHash, AgeMSBefore: state.AgeMS, AgeMSAfter: state.AgeMS, AddedNetworkSlots: []save.NetworkSlot{}, AddedLedgerFactKinds: []string{}, AddedLifetimeAchievements: []string{}, ResultFounderWireVersion: save.VersionForState(state), Rejection: &founderAuditRejection{Category: "not_eligible", Detail: "tier"}}
			},
			source: &save.FounderLogSource{CompanyStreamID: "01986666-1800-7000-8000-000000000001", RunSeq: 1, RunLogSeq: 2}},
		{payload: `{"intent_id":"01986666-0800-7000-8000-000000000803","kind":"wind_down","expected_revision":1,"expected_founder_revision":2}`,
			resolved: func(state *save.State) any {
				return founderExitResolvedWire{Kind: founderExitResolvedKind, Outcome: string(save.IntentApplied), CompanyStreamID: "01986666-1800-7000-8000-000000000001", RunSeq: 1, RunLogSeq: 3, ResultConstantsHash: catalogs.ConstantsHash, ReputationDelta: 2, RouteKnowledgeDelta: 5, AttendedMS: 900_000, AgeMSBefore: state.AgeMS, AgeMSAfter: state.AgeMS + 900_000, AddedNetworkSlots: []save.NetworkSlot{}, AddedLedgerFactKinds: []string{"exit.collapse"}, AddedLifetimeAchievements: []string{}, ExitRecord: &founderExitRecordWire{RunID: 1, ExitType: "collapse", OccurredAtMS: now.Add(3 * time.Second).UnixMilli(), ReputationDelta: 2}, ResultFounderWireVersion: save.VersionForState(state)}
			},
			source: &save.FounderLogSource{CompanyStreamID: "01986666-1800-7000-8000-000000000001", RunSeq: 1, RunLogSeq: 3}},
	}
	revision := int64(1)
	for index, definition := range definitions {
		request, err := ParseIntent([]byte(definition.payload))
		if err != nil {
			t.Fatal(err)
		}
		command := save.FounderReplayCommand{IntentID: request.IntentID, FounderStreamID: streamID, FounderID: founderID, Revision: revision, FounderLogSeq: int64(index + 1), ServerTSMS: now.Add(time.Duration(index+1) * time.Second).UnixMilli()}
		inputs, err := save.MarshalFounderReplayInputs(command, definition.resolved(state))
		if err != nil {
			t.Fatal(err)
		}
		transition, err := ApplyFounderLogged(state, request.CanonicalPayload, catalogs, inputs)
		if err != nil {
			t.Fatalf("Founder full run step %d: %v", index, err)
		}
		var applied *int64
		if transition.Outcome == save.IntentApplied {
			revision++
			value := revision
			applied = &value
		}
		events := fixtureEvents(transition.Events)
		history.Entries = append(history.Entries, save.FounderHistoryEntry{Sequence: int64(index + 1), IntentID: request.IntentID, ConstantsHash: catalogs.ConstantsHash,
			CanonicalPayload: request.CanonicalPayload, ReplayInputs: inputs, Receipt: transition.Receipt, AppliedRevision: applied, ServerTSMS: command.ServerTSMS, Source: definition.source, Events: transition.Events})
		result.Entries = append(result.Entries, crossRuntimeFounderRunEntry{Sequence: int64(index + 1), IntentID: request.IntentID, ConstantsHash: catalogs.ConstantsHash,
			CanonicalPayload: request.CanonicalPayload, ReplayInputs: inputs, ReceiptJSON: canonicalFixtureJSON(t, transition.Receipt), EventsJSON: canonicalFixtureValue(t, events), AppliedRevision: applied, ServerTSMS: command.ServerTSMS, Source: definition.source})
	}
	head := mustEncodeState(t, state)
	history.HeadRevision, history.HeadVersion, history.HeadConstants, history.HeadState = revision, save.VersionForState(state), catalogs.ConstantsHash, head
	if verdict := VerifyFounderHistory(history, ReplayCatalogSet{catalogs.ConstantsHash: catalogs}); verdict != ReplayVerified {
		t.Fatalf("Founder full-run verdict=%s", verdict)
	}
	result.HeadRevision, result.HeadVersion, result.HeadHash, result.HeadState = revision, save.VersionForState(state), catalogs.ConstantsHash, head
	return result
}

func replayFounderFixtureState(t *testing.T, catalogs CatalogBundle, now time.Time) *save.State {
	t.Helper()
	ledger, err := economy.NewLedger(catalogs.Economy, economy.ScopeFounder)
	if err != nil {
		t.Fatal(err)
	}
	counts, provisioned, remainders := map[string]int64{}, map[string]int64{}, map[string]int64{}
	for _, generator := range catalogs.Economy.GeneratorClassesForScope(economy.ScopeFounder) {
		counts[generator.ID], provisioned[generator.ID] = 0, 0
		if generator.Provision != nil {
			remainders[generator.Provision.GeneratorID] = 0
		}
	}
	state := &save.State{WireVersion: save.LatestSupportedVersion, Ledger: ledger, GeneratorCounts: counts, UpgradesOwned: map[string]bool{}, GeneratorProvisioned: provisioned,
		ProvisionRemaindersPPM: remainders, EvaluatedThrough: now, ManualTokenRefilledAt: now, GatesCrossed: map[string]bool{}, DoctrinesByTransition: map[string]string{},
		LedgerFactKinds: map[string]bool{}, MeterValues: map[string]int{}, MeterDecayRemainders: map[string]int64{}, MeterInputRemainders: map[string]int64{}, AchievementsEarnedRun: map[string]bool{},
		AchievementsEarnedLifetime: map[string]bool{}, RegionTraits: map[string]bool{}, HintsUnlocked: map[string]bool{}, CompactSamples: []save.CompactSample{}, LifetimeValue: decimal.Zero,
		OfflineSpans: []save.OfflineSpan{}, NetworkSlots: []save.NetworkSlot{}, ExitHistory: []save.ExitRecord{}}
	if catalogs.Minigames != nil {
		state.WireVersion = 17
		state.MinigameRatings = map[string]save.MinigameRatingState{}
		state.MinigameOfflineQuality = map[string]save.MinigameOfflineQualityState{}
	}
	if catalogs.Pets != nil {
		state.WireVersion = 18
		state.Pets = map[string]pet.CareState{}
	}
	if _, err := save.EncodeState(state); err != nil {
		t.Fatalf("Founder fixture base state: %v", err)
	}
	return state
}

func makeActiveFoundationExitFixture(t *testing.T, now time.Time) crossRuntimeActiveExit {
	t.Helper()
	_, current := foundationTestBundles(t)
	next := retunedAchievementBundle(t, current, current.Achievements.Definitions[0].ScoreGrant+1)
	current.Next = &next
	company := replayFixtureState(t, current.Economy, now.Add(-20*time.Minute))
	company.WireVersion = save.LatestSupportedVersion
	meterState, err := meters.NewRunState(current.Meters, 17)
	if err != nil {
		t.Fatal(err)
	}
	company.MeterBands = nil
	company.MeterValues, company.MeterDecayRemainders, company.MeterInputRemainders = meterState.Values, meterState.DecayRemainders, meterState.InputRemainders
	company.AchievementsEarnedRun = map[string]bool{}
	company.Tier = 3
	company.LifetimeValue = decimal.New(27, 12)
	terms := json.RawMessage(`{"market_modifier_ppm":1100000,"payout_preview":{"reputation_delta":5,"network_slot_unlocks":[],"route_knowledge":0,"clout_reach_note":"clout.reach.preserved"}}`)
	company.OfferState = &save.ExitOfferState{OfferID: "01986666-0600-7000-8000-000000000600", ExitType: "acquisition", TermsJSON: terms, SpawnedAt: now.Add(-time.Minute), ExpiresAt: now.Add(time.Minute)}
	preState := mustEncodeState(t, company)
	request, err := ParseIntent([]byte(`{"intent_id":"01986666-0600-7000-8000-000000000601","kind":"accept_exit_offer","expected_revision":1,"expected_founder_revision":2,"offer_id":"01986666-0600-7000-8000-000000000600"}`))
	if err != nil {
		t.Fatal(err)
	}
	definition := current.Achievements.Definitions[0]
	carry := replayFounderCarry{FounderRevision: 2, FounderConstantsHash: current.ConstantsHash, ReputationLevel: 1, Notoriety: 17, NetworkSlots: []save.NetworkSlot{}, LedgerFactKinds: []string{}, ExitHistoryCount: 1,
		AchievementsEarnedLifetime: []string{definition.ID}, AchievementScoreLifetime: definition.ScoreGrant}
	command := save.ReplayCommand{IntentID: request.IntentID, CompanyStreamID: "01986666-1600-7000-8000-000000000001", FounderID: "01986666-2600-7000-8000-000000000001", Revision: 1, RunSeq: 1, RunLogSeq: 1}
	inputs, err := buildReplayInputs(replayBuild{Command: command, Mode: ModeOnline, Now: now, IntentKind: request.Kind, RouteContextVersion: current.Routes.ContextVersion(), FounderCarry: &carry, Terminal: true, ExecutedRouteIDs: []string{}, SelectedExitType: "acquisition", SelectedTerms: terms, NextConstantsHash: next.ConstantsHash})
	if err != nil {
		t.Fatal(err)
	}
	result := executeTerminalFixture(t, "active-foundation-exit", current, company, preState, request, inputs, carry)
	return crossRuntimeActiveExit{ConstantsHash: current.ConstantsHash, Artifacts: stringArtifacts(current.Artifacts), NextConstantsHash: next.ConstantsHash, NextArtifacts: stringArtifacts(next.Artifacts), Case: result}
}

func stringArtifacts(source map[string][]byte) map[string]string {
	result := make(map[string]string, len(source))
	for name, data := range source {
		result[name] = string(data)
	}
	return result
}

func makeActiveFoundationReplayFixture(t *testing.T, now time.Time, gap time.Duration, name string) crossRuntimeBundleCase {
	t.Helper()
	_, catalogs := foundationTestBundles(t)
	state := replayFixtureState(t, catalogs.Economy, now.Add(-gap))
	state.WireVersion = save.LatestSupportedVersion
	meterState, err := meters.NewRunState(catalogs.Meters, 17)
	if err != nil {
		t.Fatal(err)
	}
	state.MeterBands = nil
	state.MeterValues, state.MeterDecayRemainders, state.MeterInputRemainders = meterState.Values, meterState.DecayRemainders, meterState.InputRemainders
	state.MeterValues["doom.probability"] = 71
	state.AchievementsEarnedRun = map[string]bool{}
	state.Tier = 1
	preState := mustEncodeState(t, state)
	request, err := ParseIntent([]byte(`{"intent_id":"01986666-0400-7000-8000-000000000401","kind":"perform_manual_batch","expected_revision":1,"action_id":"manual.click","count":1,"window_ms":1}`))
	if err != nil {
		t.Fatal(err)
	}
	definition := catalogs.Achievements.Definitions[1]
	carry := replayFounderCarry{FounderRevision: 1, FounderConstantsHash: catalogs.ConstantsHash, NetworkSlots: []save.NetworkSlot{}, LedgerFactKinds: []string{}, ExitHistoryCount: 1,
		AchievementsEarnedLifetime: []string{definition.ID}, AchievementScoreLifetime: definition.ScoreGrant}
	command := save.ReplayCommand{IntentID: request.IntentID, CompanyStreamID: "01986666-1400-7000-8000-000000000001", FounderID: "01986666-2400-7000-8000-000000000001", Revision: 1, RunSeq: 1, RunLogSeq: 1}
	inputs, err := buildReplayInputs(replayBuild{Command: command, Mode: ModeOnline, Now: now, IntentKind: request.Kind, RouteContextVersion: catalogs.Routes.ContextVersion(), FounderCarry: &carry})
	if err != nil {
		t.Fatal(err)
	}
	transition, err := ApplyLogged(state, request.CanonicalPayload, catalogs, inputs)
	if err != nil {
		t.Fatal(err)
	}
	postState := mustEncodeState(t, state)
	artifacts := make(map[string]string, len(catalogs.Artifacts))
	for name, data := range catalogs.Artifacts {
		artifacts[name] = string(data)
	}
	events := fixtureEvents(transition.Events)
	return crossRuntimeBundleCase{ConstantsHash: catalogs.ConstantsHash, Artifacts: artifacts, Case: crossRuntimeFixtureCase{
		Name: name, PreState: preState, CanonicalPayload: request.CanonicalPayload, ReplayInputs: inputs,
		Outcome: string(transition.Outcome), Receipt: transition.Receipt, Events: events, PostState: postState,
		ReceiptJSON: canonicalFixtureJSON(t, transition.Receipt), EventsJSON: canonicalFixtureValue(t, events), PostStateJSON: canonicalFixtureJSON(t, postState),
	}}
}

func makeActiveFoundationBandReplayFixture(t *testing.T, now time.Time) crossRuntimeBundleCase {
	t.Helper()
	_, catalogs := foundationTestBundles(t)
	state := replayFixtureState(t, catalogs.Economy, now.Add(-4*time.Second))
	state.WireVersion = save.LatestSupportedVersion
	meterState, err := meters.NewRunState(catalogs.Meters, 17)
	if err != nil {
		t.Fatal(err)
	}
	state.MeterBands = nil
	state.MeterValues, state.MeterDecayRemainders, state.MeterInputRemainders = meterState.Values, meterState.DecayRemainders, meterState.InputRemainders
	state.MeterValues["doom.probability"] = 70
	state.MeterDecayRemainders["doom.probability"] = 3_596_000
	state.AchievementsEarnedRun = map[string]bool{}
	state.Tier = 1
	preState := mustEncodeState(t, state)
	request, err := ParseIntent([]byte(`{"intent_id":"01986666-0400-7000-8000-000000000403","kind":"perform_manual_batch","expected_revision":1,"action_id":"manual.click","count":1,"window_ms":1}`))
	if err != nil {
		t.Fatal(err)
	}
	definition := catalogs.Achievements.Definitions[1]
	carry := replayFounderCarry{FounderRevision: 1, FounderConstantsHash: catalogs.ConstantsHash, NetworkSlots: []save.NetworkSlot{}, LedgerFactKinds: []string{}, ExitHistoryCount: 1,
		AchievementsEarnedLifetime: []string{definition.ID}, AchievementScoreLifetime: definition.ScoreGrant}
	command := save.ReplayCommand{IntentID: request.IntentID, CompanyStreamID: "01986666-1400-7000-8000-000000000003", FounderID: "01986666-2400-7000-8000-000000000003", Revision: 1, RunSeq: 1, RunLogSeq: 1}
	inputs, err := buildReplayInputs(replayBuild{Command: command, Mode: ModeOnline, Now: now, IntentKind: request.Kind, RouteContextVersion: catalogs.Routes.ContextVersion(), FounderCarry: &carry})
	if err != nil {
		t.Fatal(err)
	}
	transition, err := ApplyLogged(state, request.CanonicalPayload, catalogs, inputs)
	if err != nil {
		t.Fatal(err)
	}
	if len(transition.Events) != 2 || transition.Events[0].Kind != save.EventMeterBandChanged || transition.Events[1].Kind != save.EventAchievementEarned {
		t.Fatalf("foundation event order=%+v", transition.Events)
	}
	postState := mustEncodeState(t, state)
	events := fixtureEvents(transition.Events)
	return crossRuntimeBundleCase{ConstantsHash: catalogs.ConstantsHash, Artifacts: stringArtifacts(catalogs.Artifacts), Case: crossRuntimeFixtureCase{
		Name: "active-foundation-band-crossing", PreState: preState, CanonicalPayload: request.CanonicalPayload, ReplayInputs: inputs,
		Outcome: string(transition.Outcome), Receipt: transition.Receipt, Events: events, PostState: postState,
		ReceiptJSON: canonicalFixtureJSON(t, transition.Receipt), EventsJSON: canonicalFixtureValue(t, events), PostStateJSON: canonicalFixtureJSON(t, postState),
	}}
}

func makeActiveFoundationBurnReplayFixture(t *testing.T, now time.Time) crossRuntimeBundleCase {
	t.Helper()
	_, catalogs := foundationTestBundles(t)
	artifact := []byte(`{"schema_version":1,"achievements":[{"id":"achievement.gate_burn","condition_scope":"run","condition":{"kind":"counter_at_least","counter":"tier","minimum":3},"proof":{"kind":"burn","event_kind":"gate_crossed","resource_id":"company.cash","minimum":"1e9"},"score_grant":3,"copy_key":"category.any_percent"}]}`)
	catalog, err := achievements.LoadCatalog(artifact, FoundationAchievementRegistry(catalogs.Economy))
	if err != nil {
		t.Fatal(err)
	}
	artifacts := make(map[string][]byte, len(catalogs.Artifacts))
	for name, data := range catalogs.Artifacts {
		artifacts[name] = append([]byte(nil), data...)
	}
	artifacts["achievements"] = artifact
	hash, err := save.ConstantsHashArtifacts(artifacts)
	if err != nil {
		t.Fatal(err)
	}
	catalogs.Artifacts, catalogs.ConstantsHash, catalogs.Achievements = artifacts, hash, catalog
	if !catalogs.valid(hash) {
		t.Fatal("burn-proof bundle is not internally valid")
	}
	state := replayFixtureState(t, catalogs.Economy, now.Add(-time.Second))
	state.WireVersion = save.LatestSupportedVersion
	meterState, err := meters.NewRunState(catalogs.Meters, 0)
	if err != nil {
		t.Fatal(err)
	}
	state.MeterBands = nil
	state.MeterValues, state.MeterDecayRemainders, state.MeterInputRemainders = meterState.Values, meterState.DecayRemainders, meterState.InputRemainders
	state.AchievementsEarnedRun = map[string]bool{}
	state.Tier = 2
	state.GeneratorCounts["generator.beige_tower"] = 1_000_000_000
	setCash(t, state, "1e9")
	preState := mustEncodeState(t, state)
	request, err := ParseIntent([]byte(`{"intent_id":"01986666-0400-7000-8000-000000000402","kind":"cross_gate","expected_revision":1,"gate_id":"gate.t2_to_t3","route_id":null}`))
	if err != nil {
		t.Fatal(err)
	}
	carry := replayFounderCarry{FounderRevision: 1, FounderConstantsHash: catalogs.ConstantsHash, NetworkSlots: []save.NetworkSlot{}, LedgerFactKinds: []string{}, AchievementsEarnedLifetime: []string{}}
	command := save.ReplayCommand{IntentID: request.IntentID, CompanyStreamID: "01986666-1400-7000-8000-000000000002", FounderID: "01986666-2400-7000-8000-000000000002", Revision: 1, RunSeq: 1, RunLogSeq: 1}
	inputs, err := buildReplayInputs(replayBuild{Command: command, Mode: ModeOnline, Now: now, IntentKind: request.Kind, RouteContextVersion: catalogs.Routes.ContextVersion(), FounderCarry: &carry})
	if err != nil {
		t.Fatal(err)
	}
	transition, err := ApplyLogged(state, request.CanonicalPayload, catalogs, inputs)
	if err != nil {
		t.Fatal(err)
	}
	if !state.AchievementsEarnedRun["achievement.gate_burn"] {
		t.Fatal("action-only debit did not satisfy burn proof")
	}
	postState := mustEncodeState(t, state)
	events := fixtureEvents(transition.Events)
	return crossRuntimeBundleCase{ConstantsHash: catalogs.ConstantsHash, Artifacts: stringArtifacts(catalogs.Artifacts), Case: crossRuntimeFixtureCase{
		Name: "active-foundation-burn-after-accrual", PreState: preState, CanonicalPayload: request.CanonicalPayload, ReplayInputs: inputs,
		Outcome: string(transition.Outcome), Receipt: transition.Receipt, Events: events, PostState: postState,
		ReceiptJSON: canonicalFixtureJSON(t, transition.Receipt), EventsJSON: canonicalFixtureValue(t, events), PostStateJSON: canonicalFixtureJSON(t, postState),
	}}
}

func makeRejectedExitFixture(t *testing.T, catalogs CatalogBundle, constantsHash string, now time.Time) crossRuntimeFullRun {
	t.Helper()
	assertPerEntryNextCatalog(t, catalogs, constantsHash, now)
	state := replayFixtureState(t, catalogs.Economy, now)
	state.Tier = 3
	state.GatesCrossed["gate.t2_to_t3"] = true
	state.FactionID = "open_source"
	state.IncorporatedAt = now.Add(-time.Minute)
	state.StockUnits = 3
	genesis, err := save.EncodeState(state)
	if err != nil {
		t.Fatal(err)
	}
	founderID := "01987778-2000-7000-8000-000000000001"
	request, err := parseLoggedIntent([]byte(`{"kind":"buy_generator","expected_revision":1,"generator_id":"generator.beige_tower","count":{"mode":"exact","value":1}}`), "01987778-0001-7000-8000-000000000001")
	if err != nil {
		t.Fatal(err)
	}
	command := save.ReplayCommand{IntentID: request.IntentID, CompanyStreamID: "01987778-1000-7000-8000-000000000001", FounderID: founderID, Revision: 1, RunSeq: 1, RunLogSeq: 1}
	inputs, err := buildReplayInputs(replayBuild{Command: command, Mode: ModeOnline, Now: now.Add(time.Second), IntentKind: request.Kind, RouteContextVersion: catalogs.Routes.ContextVersion()})
	if err != nil {
		t.Fatal(err)
	}
	ordinaryRejected, err := ApplyLogged(state, request.CanonicalPayload, catalogs, inputs)
	if err != nil || ordinaryRejected.Outcome != save.IntentRejected {
		t.Fatalf("late rejected intent outcome=%s err=%v", ordinaryRejected.Outcome, err)
	}
	if after := mustEncodeState(t, state); !bytes.Equal(genesis, after) {
		t.Fatalf("late rejected intent mutated replay state\nbefore=%s\nafter=%s", genesis, after)
	}
	entries := []crossRuntimeFullRunEntry{{Seq: 1, CanonicalPayload: request.CanonicalPayload, ReplayInputs: inputs, ReceiptJSON: canonicalFixtureJSON(t, ordinaryRejected.Receipt), EventsJSON: "[]"}}

	request, err = parseLoggedIntent([]byte(`{"kind":"cross_gate","expected_revision":1,"gate_id":"gate.t2_to_t3","route_id":null}`), "01987778-0002-7000-8000-000000000002")
	if err != nil {
		t.Fatal(err)
	}
	command = save.ReplayCommand{IntentID: request.IntentID, CompanyStreamID: command.CompanyStreamID, FounderID: founderID, Revision: 1, RunSeq: 1, RunLogSeq: 2}
	carry := replayFounderCarry{FounderRevision: 1, FounderConstantsHash: constantsHash, NetworkSlots: []save.NetworkSlot{}, LedgerFactKinds: []string{}}
	inputs, err = buildReplayInputs(replayBuild{Command: command, Mode: ModeOnline, Now: now.Add(2 * time.Second), IntentKind: request.Kind,
		RouteContextVersion: catalogs.Routes.ContextVersion(), FounderCarry: &carry, Terminal: true, ExecutedRouteIDs: []string{},
		SelectedExitType: "scripted_first", SelectedTerms: json.RawMessage(`{}`), NextConstantsHash: constantsHash,
		GuildSettlementBatch: guild.SettlementBatch{GuildID: "01987778-3000-7000-8000-000000000001", BaseSeq: 0,
			Settlements: []guild.Settlement{{GuildID: "01987778-3000-7000-8000-000000000001", BoundarySeq: 1, DebitUnits: 1, CreditUnits: 2}}}})
	if err != nil {
		t.Fatal(err)
	}
	exit, err := ApplyLoggedExit(state, request.CanonicalPayload, catalogs, inputs)
	if err != nil || exit.Decision.Outcome != save.IntentRejected {
		t.Fatalf("rejected exit outcome=%s err=%v", exit.Decision.Outcome, err)
	}
	if after := mustEncodeState(t, state); !bytes.Equal(genesis, after) {
		t.Fatalf("rejected terminal preflight mutated replay state\nbefore=%s\nafter=%s", genesis, after)
	}
	entries = append(entries, crossRuntimeFullRunEntry{Seq: 2, CanonicalPayload: request.CanonicalPayload, ReplayInputs: inputs, ReceiptJSON: canonicalFixtureJSON(t, exit.Decision.Receipt), EventsJSON: "[]", Terminal: true})

	request, err = parseLoggedIntent([]byte(`{"kind":"perform_manual_batch","expected_revision":1,"action_id":"manual.click","count":1,"window_ms":1000}`), "01987778-0003-7000-8000-000000000003")
	if err != nil {
		t.Fatal(err)
	}
	command = save.ReplayCommand{IntentID: request.IntentID, CompanyStreamID: command.CompanyStreamID, FounderID: founderID, Revision: 1, RunSeq: 1, RunLogSeq: 3}
	inputs, err = buildReplayInputs(replayBuild{Command: command, Mode: ModeOnline, Now: now.Add(3 * time.Second), IntentKind: request.Kind, RouteContextVersion: catalogs.Routes.ContextVersion()})
	if err != nil {
		t.Fatal(err)
	}
	ordinary, err := ApplyLogged(state, request.CanonicalPayload, catalogs, inputs)
	if err != nil {
		t.Fatal(err)
	}
	entries = append(entries, crossRuntimeFullRunEntry{Seq: 3, CanonicalPayload: request.CanonicalPayload, ReplayInputs: inputs, ReceiptJSON: canonicalFixtureJSON(t, ordinary.Receipt), EventsJSON: canonicalFixtureValue(t, fixtureEvents(ordinary.Events))})
	full := crossRuntimeFullRun{Genesis: genesis, Entries: entries, FinalStateJSON: canonicalFixtureJSON(t, mustEncodeState(t, state))}
	verifiedEntries := []ReplayLogEntry{
		{Sequence: 1, CanonicalPayload: entries[0].CanonicalPayload, ReplayInputs: entries[0].ReplayInputs, ReceiptJSON: []byte(entries[0].ReceiptJSON), EventsJSON: []byte(entries[0].EventsJSON)},
		{Sequence: 2, CanonicalPayload: entries[1].CanonicalPayload, ReplayInputs: entries[1].ReplayInputs, ReceiptJSON: []byte(entries[1].ReceiptJSON), EventsJSON: []byte(entries[1].EventsJSON), Terminal: true},
		{Sequence: 3, CanonicalPayload: entries[2].CanonicalPayload, ReplayInputs: entries[2].ReplayInputs, ReceiptJSON: []byte(entries[2].ReceiptJSON), EventsJSON: []byte(entries[2].EventsJSON)},
	}
	if verdict := VerifyReplayRun(genesis, save.CurrentVersion, catalogs, verifiedEntries, constantsHash, false); verdict != ReplayLogGap {
		t.Fatalf("rejected-exit continuation verdict=%s", verdict)
	}
	tampered := append([]ReplayLogEntry(nil), verifiedEntries...)
	tampered[2].ReceiptJSON = []byte(`{}`)
	if verdict := VerifyReplayRun(genesis, save.CurrentVersion, catalogs, tampered, constantsHash, false); verdict != ReplayStateDivergence {
		t.Fatalf("rejected-exit continuation tamper verdict=%s", verdict)
	}
	if verdict := VerifyReplayRun(genesis, save.CurrentVersion, catalogs, verifiedEntries[:2], constantsHash, false); verdict != ReplayLogGap {
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
		t.Fatalf("encode state version=%d pets=%d: %v", save.VersionForState(state), len(state.Pets), err)
	}
	return encoded
}

func makeFullRunFixture(t *testing.T, catalogs CatalogBundle, constantsHash string, startedAt time.Time) crossRuntimeFullRun {
	t.Helper()
	boundaryArtifacts := make(map[string][]byte, len(catalogs.Artifacts))
	for name, data := range catalogs.Artifacts {
		boundaryArtifacts[name] = bytes.Clone(data)
	}
	var economyArtifact map[string]any
	if err := json.Unmarshal(boundaryArtifacts["economy"], &economyArtifact); err != nil {
		t.Fatal(err)
	}
	resources := economyArtifact["resources"].([]any)
	resources[0].(map[string]any)["hardcap"].(map[string]any)["amount"] = "1e4000000000000000"
	boundaryArtifacts["economy"], _ = json.Marshal(economyArtifact)
	var err error
	constantsHash, err = save.ConstantsHashArtifacts(boundaryArtifacts)
	if err != nil {
		t.Fatal(err)
	}
	catalogs = loadReplayTestBundle(t, constantsHash, boundaryArtifacts)
	fixtureArtifacts := make(map[string]string, len(boundaryArtifacts))
	for name, data := range boundaryArtifacts {
		fixtureArtifacts[name] = string(data)
	}
	founderID := offerFixtureFounder(t, catalogs.Prestige.SpawnGatePPM[3])
	state := replayFixtureState(t, catalogs.Economy, startedAt)
	state.Tier = 2
	state.LifetimeValue = decimal.New(8, 12)
	state.DoctrinesByTransition["transition.t3_to_t4"] = "doctrine.capture"
	// The sequential corpus deliberately begins at the declared hardcap. Its
	// max purchase reaches the closed-form boundary and exercises the invariant
	// event path without swapping catalogs mid-run.
	setCash(t, state, "1e4000000000000000")
	genesis, err := save.EncodeState(state)
	if err != nil {
		t.Fatal(err)
	}
	entries := make([]crossRuntimeFullRunEntry, 0, 51)
	revision := int64(1)
	now := startedAt
	ordinary := []string{
		`{"kind":"perform_manual_batch","expected_revision":%d,"action_id":"manual.click","count":2,"window_ms":1000}`,
		`{"kind":"buy_generator","expected_revision":%d,"generator_id":"generator.beige_tower","count":{"mode":"max"}}`,
		`{"kind":"sign_compact","expected_revision":%d,"tithe_ppm":110000}`,
		`{"kind":"incorporate","expected_revision":%d,"faction_id":"open_source"}`,
		`{"kind":"perform_manual_batch","expected_revision":%d,"action_id":"manual.click","count":1,"window_ms":1000}`,
		`{"kind":"cross_gate","expected_revision":%d,"gate_id":"gate.t2_to_t3","route_id":null}`,
	}
	for index := 0; index < 50; index++ {
		mode := ModeOnline
		if index == 10 {
			now = now.Add(48 * time.Hour)
			mode = ModeOffline
		} else {
			now = now.Add(5 * time.Second)
		}
		if index == 6 {
			if state.OfferState == nil {
				t.Fatal("full run gate did not spawn the offer required by the expiry step")
			}
			now = state.OfferState.ExpiresAt
		}
		payload := fmt.Sprintf(`{"kind":"perform_manual_batch","expected_revision":%d,"action_id":"manual.click","count":1,"window_ms":1000}`, revision)
		if index < len(ordinary) {
			payload = fmt.Sprintf(ordinary[index], revision)
		}
		if index == 7 {
			payload = fmt.Sprintf(`{"kind":"cross_gate","expected_revision":%d,"gate_id":"gate.t4_to_t5","route_id":"route.ipo_sequence_break"}`, revision)
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
		if index == 20 {
			inputs = withReplaySettlement(t, inputs, replayGuildSettlementBatch{GuildID: "01987777-3000-7000-8000-000000000001", BaseSeq: 0,
				Settlements: []replayGuildSettlement{{BoundarySeq: 1, DebitUnits: 1, CreditUnits: 7}}})
		}
		transition, err := ApplyLogged(state, request.CanonicalPayload, catalogs, inputs)
		if err != nil || transition.Outcome != save.IntentApplied {
			t.Fatalf("full run step %d outcome=%s err=%v", index+1, transition.Outcome, err)
		}
		if index == 3 && !hasEventKind(transition.Events, save.EventCompactTitheRaised) {
			t.Fatal("full run existing-member Open Source incorporation omitted compact_tithe_raised")
		}
		if index == 1 {
			var receipt struct {
				AppliedCount int64 `json:"applied_count"`
			}
			if err := json.Unmarshal(transition.Receipt, &receipt); err != nil || request.CountMode != "max" || receipt.AppliedCount < 1 || !hasEventKind(transition.Events, save.EventInvariantReported) {
				t.Fatalf("full run max/invariant step missing: mode=%q count=%d events=%+v err=%v", request.CountMode, receipt.AppliedCount, transition.Events, err)
			}
		}
		if index == 6 && !hasEventKind(transition.Events, save.EventExitOfferExpired) {
			t.Fatalf("full run offer expiry missing: events=%+v", transition.Events)
		}
		if index == 7 && !hasEventKind(transition.Events, save.EventRouteExecuted) {
			t.Fatal("full run discounted route crossing omitted route_executed")
		}
		if index == 20 {
			parsed, parseErr := parseReplayInputs(inputs)
			var resolved replayAccrualResolved
			decodeErr := decodeReplayStrict(parsed.Resolved, &resolved)
			batch := resolved.Accrual.GuildSettlementBatch
			if parseErr != nil || decodeErr != nil || batch.GuildID != "01987777-3000-7000-8000-000000000001" ||
				batch.BaseSeq != 0 || len(batch.Settlements) != 1 || batch.Settlements[0].BoundarySeq != 1 ||
				batch.Settlements[0].DebitUnits != 1 || batch.Settlements[0].CreditUnits != 7 || state.GuildBoundarySeq != 1 ||
				state.GuildConsumedWindow != 7 || state.ConsumedStockUnits != 7 {
				t.Fatalf("full run guild settlement batch missing or changed: %+v", resolved.Accrual.GuildSettlementBatch)
			}
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
	inputs, err := buildReplayInputs(replayBuild{Command: command, Mode: ModeOnline, Now: now, IntentKind: request.Kind, RouteContextVersion: catalogs.Routes.ContextVersion(), FounderCarry: &carry, CommonsWeightPPM: int64Pointer(800_000), Terminal: true, ExecutedRouteIDs: []string{"route.ipo_sequence_break"}, SelectedExitType: "collapse", SelectedTerms: json.RawMessage(`{}`), NextConstantsHash: constantsHash})
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
	full := crossRuntimeFullRun{ConstantsHash: constantsHash, Artifacts: fixtureArtifacts, Genesis: genesis, Entries: entries, FinalStateJSON: canonicalFixtureJSON(t, finalState)}
	assertFullRunVerifier(t, full, catalogs, constantsHash)
	return full
}

func hasEventKind(events []save.EventWrite, kind save.EventKind) bool {
	for _, event := range events {
		if event.Kind == kind {
			return true
		}
	}
	return false
}

func withReplaySettlement(t *testing.T, inputs json.RawMessage, settlement replayGuildSettlementBatch) json.RawMessage {
	t.Helper()
	var wire map[string]json.RawMessage
	if err := json.Unmarshal(inputs, &wire); err != nil {
		t.Fatal(err)
	}
	var resolved map[string]json.RawMessage
	if err := json.Unmarshal(wire["resolved"], &resolved); err != nil {
		t.Fatal(err)
	}
	var accrual map[string]json.RawMessage
	if err := json.Unmarshal(resolved["accrual"], &accrual); err != nil {
		t.Fatal(err)
	}
	batch, err := json.Marshal(settlement)
	if err != nil {
		t.Fatal(err)
	}
	accrual["guild_settlement_batch"] = batch
	resolved["accrual"], err = json.Marshal(accrual)
	if err != nil {
		t.Fatal(err)
	}
	wire["resolved"], err = json.Marshal(resolved)
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(wire)
	if err != nil {
		t.Fatal(err)
	}
	return encoded
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
	if verdict := VerifyReplayRun(full.Genesis, 11, catalogs, entries, constantsHash, false); verdict != ReplayConstantsMismatch {
		t.Fatalf("pre-genesis-version-floor verdict=%s", verdict)
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

func makeFoundationFixtures(t *testing.T, source map[string][]byte, now time.Time) []crossRuntimeBundleCase {
	t.Helper()
	artifacts := make(map[string][]byte, len(source))
	for name, data := range source {
		artifacts[name] = bytes.Clone(data)
	}
	economyBytes, err := os.ReadFile("../../testdata/economy-foundation-v4.json")
	if err != nil {
		t.Fatal(err)
	}
	var economyArtifact map[string]any
	if err := json.Unmarshal(economyBytes, &economyArtifact); err != nil {
		t.Fatal(err)
	}
	economyArtifact["upgrades"].([]any)[0].(map[string]any)["window"].(map[string]any)["to_gate"] = "gate.t2_to_t3"
	artifacts["economy"], err = json.Marshal(economyArtifact)
	if err != nil {
		t.Fatal(err)
	}
	hash, err := save.ConstantsHashArtifacts(artifacts)
	if err != nil {
		t.Fatal(err)
	}
	catalogs := loadReplayTestBundle(t, hash, artifacts)
	fixtureArtifacts := make(map[string]string, len(artifacts))
	for name, data := range artifacts {
		fixtureArtifacts[name] = string(data)
	}
	definitions := []struct {
		name      string
		payload   string
		advance   time.Duration
		configure func(*save.State)
	}{
		{name: "foundation-buy-upgrade", payload: `{"intent_id":"01986666-0301-7000-8000-000000000301","kind":"buy_upgrade","expected_revision":1,"upgrade_id":"upgrade.click"}`, configure: func(state *save.State) { setCash(t, state, "1e3") }},
		{name: "foundation-manual-role", payload: `{"intent_id":"01986666-0302-7000-8000-000000000302","kind":"perform_manual_batch","expected_revision":1,"action_id":"manual.click","count":1,"window_ms":1}`, configure: func(state *save.State) { state.GeneratorCounts["generator.low"] = 10 }},
		{name: "foundation-provision-grid", payload: `{"intent_id":"01986666-0303-7000-8000-000000000303","kind":"perform_manual_batch","expected_revision":1,"action_id":"manual.click","count":1,"window_ms":1}`, advance: 180 * time.Second, configure: func(state *save.State) { state.GeneratorCounts["generator.high"] = 2 }},
		{name: "foundation-combined-content", payload: `{"intent_id":"01986666-0304-7000-8000-000000000304","kind":"perform_manual_batch","expected_revision":1,"action_id":"manual.click","count":1,"window_ms":1}`, advance: time.Second, configure: func(state *save.State) {
			state.GeneratorCounts["generator.low"] = 10
			state.UpgradesOwned["upgrade.click"] = true
		}},
	}
	result := make([]crossRuntimeBundleCase, 0, len(definitions))
	for index, definition := range definitions {
		state := replayFixtureState(t, catalogs.Economy, now)
		if definition.configure != nil {
			definition.configure(state)
		}
		preState := mustEncodeState(t, state)
		request, err := ParseIntent([]byte(definition.payload))
		if err != nil {
			t.Fatal(err)
		}
		command := save.ReplayCommand{IntentID: request.IntentID, CompanyStreamID: "01986666-1300-7000-8000-000000000001", FounderID: "01986666-2300-7000-8000-000000000001", Revision: 1, RunSeq: 1, RunLogSeq: int64(index + 1)}
		inputs, err := buildReplayInputs(replayBuild{Command: command, Mode: ModeOnline, Now: now.Add(definition.advance), IntentKind: request.Kind, RouteContextVersion: catalogs.Routes.ContextVersion()})
		if err != nil {
			t.Fatal(err)
		}
		transition, err := ApplyLogged(state, request.CanonicalPayload, catalogs, inputs)
		if err != nil {
			t.Fatalf("%s: %v", definition.name, err)
		}
		postState := mustEncodeState(t, state)
		events := fixtureEvents(transition.Events)
		result = append(result, crossRuntimeBundleCase{ConstantsHash: hash, Artifacts: fixtureArtifacts, Case: crossRuntimeFixtureCase{
			Name: definition.name, PreState: preState, CanonicalPayload: request.CanonicalPayload, ReplayInputs: inputs,
			Outcome: string(transition.Outcome), Receipt: transition.Receipt, Events: events, PostState: postState,
			ReceiptJSON: canonicalFixtureJSON(t, transition.Receipt), EventsJSON: canonicalFixtureValue(t, events), PostStateJSON: canonicalFixtureJSON(t, postState),
		}})
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
		"ledger_fact_kinds": facts, "exit_history_count": len(state.ExitHistory),
		"achievements_earned_lifetime": sortedBoolKeys(state.AchievementsEarnedLifetime),
		"achievement_score_lifetime":   state.AchievementScoreLifetime}
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
	counts := map[string]int64{}
	provisioned := map[string]int64{}
	remainders := map[string]int64{}
	for _, generator := range catalog.GeneratorClassesForScope(economy.ScopeCompany) {
		counts[generator.ID] = 0
		provisioned[generator.ID] = 0
		if generator.Provision != nil {
			remainders[generator.Provision.GeneratorID] = 0
		}
	}
	return &save.State{Ledger: ledger, GeneratorCounts: counts, UpgradesOwned: map[string]bool{},
		GeneratorProvisioned: provisioned, ProvisionRemaindersPPM: remainders, EvaluatedThrough: now,
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

package production

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"cloud-clicker/server/achievements"
	"cloud-clicker/server/activeplay"
	"cloud-clicker/server/decimal"
	"cloud-clicker/server/doctrine"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/guild"
	"cloud-clicker/server/meters"
	"cloud-clicker/server/minigame"
	"cloud-clicker/server/minigameapi"
	"cloud-clicker/server/multiplier"
	"cloud-clicker/server/pet"
	prestigecore "cloud-clicker/server/prestige"
	"cloud-clicker/server/routes"
	"cloud-clicker/server/save"
)

var updateReplayFixture = flag.Bool("update-replay-fixture", false, "rewrite the shared ApplyLogged fixture")

type crossRuntimeFixture struct {
	Version            int                           `json:"version"`
	ConstantsHash      string                        `json:"constants_hash"`
	Artifacts          map[string]string             `json:"artifacts"`
	Cases              []crossRuntimeFixtureCase     `json:"cases"`
	TerminalCases      []crossRuntimeTerminalCase    `json:"terminal_cases"`
	Additional         []crossRuntimeBundleCase      `json:"additional_bundles"`
	ActiveExit         crossRuntimeActiveExit        `json:"active_foundation_exit"`
	ActivePlayExit     crossRuntimeActiveExit        `json:"active_play_exit"`
	FirstContentExit   crossRuntimeActiveExit        `json:"first_content_exit"`
	FullRun            crossRuntimeFullRun           `json:"full_run"`
	DoctrineRun        crossRuntimeFullRun           `json:"doctrine_run"`
	ActivePlayRun      crossRuntimeFullRun           `json:"active_play_run"`
	RejectedExit       crossRuntimeFullRun           `json:"rejected_exit_run"`
	FounderHash        string                        `json:"founder_constants_hash"`
	FounderFiles       map[string]string             `json:"founder_artifacts"`
	FounderCases       []crossRuntimeFounderCase     `json:"founder_cases"`
	FounderRun         crossRuntimeFounderRun        `json:"founder_run"`
	PetFounderHash     string                        `json:"pet_founder_constants_hash"`
	PetFounderFiles    map[string]string             `json:"pet_founder_artifacts"`
	PetFounderCases    []crossRuntimeFounderCase     `json:"pet_founder_cases"`
	MinigameHash       string                        `json:"minigame_constants_hash"`
	MinigameFiles      map[string]string             `json:"minigame_artifacts"`
	MinigameCompany    crossRuntimeFixtureCase       `json:"minigame_company_case"`
	MinigameFounder    crossRuntimeFounderCase       `json:"minigame_founder_case"`
	SoulHash           string                        `json:"soul_constants_hash"`
	SoulFiles          map[string]string             `json:"soul_artifacts"`
	SoulCompany        crossRuntimeFixtureCase       `json:"soul_company_case"`
	SoulFounder        crossRuntimeFounderCase       `json:"soul_founder_case"`
	MinigameStartHash  string                        `json:"minigame_start_constants_hash"`
	MinigameStartFiles map[string]string             `json:"minigame_start_artifacts"`
	MinigameStart      crossRuntimeFounderCase       `json:"minigame_start_founder_case"`
	MinigameExitReset  crossRuntimeFounderCase       `json:"minigame_exit_reset_founder_case"`
	MinigameActiveExit crossRuntimeFixtureCase       `json:"minigame_active_exit_case"`
	MinigameActivation crossRuntimeFounderActivation `json:"minigame_activation"`
}

type crossRuntimeFounderActivation struct {
	ConstantsHash     string                  `json:"constants_hash"`
	Artifacts         map[string]string       `json:"artifacts"`
	NextConstantsHash string                  `json:"next_constants_hash"`
	NextArtifacts     map[string]string       `json:"next_artifacts"`
	Case              crossRuntimeFounderCase `json:"case"`
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
	catalogs := epoch5TestBundle(t)
	artifacts := make(map[string]string, len(catalogs.Artifacts))
	for name, data := range catalogs.Artifacts {
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
		{name: "session-boundary-offline-catchup", payload: `{"intent_id":"01986666-0019-7000-8000-000000000019","kind":"perform_manual_batch","expected_revision":1,"action_id":"manual.click","count":2,"window_ms":100}`, advance: 48 * time.Hour, mode: ModeOnline, configure: func(state *save.State) { state.GeneratorCounts["generator.beige_tower"] = 3 }},
		{name: "buy-generator", payload: `{"intent_id":"01986666-0003-7000-8000-000000000003","kind":"buy_generator","expected_revision":1,"generator_id":"generator.beige_tower","count":{"mode":"exact","value":4}}`, mode: ModeOnline, configure: func(state *save.State) { setCash(t, state, "1e4") }},
		{name: "buy-generator-max", payload: `{"intent_id":"01986666-0014-7000-8000-000000000014","kind":"buy_generator","expected_revision":1,"generator_id":"generator.beige_tower","count":{"mode":"max"}}`, mode: ModeOnline, configure: func(state *save.State) { setCash(t, state, "1e1000") }},
		{name: "purchase-total-cap-precedes-affordability", payload: `{"intent_id":"01986666-0018-7000-8000-000000000018","kind":"buy_generator","expected_revision":1,"generator_id":"generator.beige_tower","count":{"mode":"exact","value":1}}`, mode: ModeOnline,
			configure: func(state *save.State) { state.GeneratorPurchasedTotal = decimal.MaxExactInteger }},
		{name: "cross-gate", payload: `{"intent_id":"01986666-0004-7000-8000-000000000004","kind":"cross_gate","expected_revision":1,"gate_id":"gate.t2_to_t3","route_id":null}`, mode: ModeOnline, configure: func(state *save.State) { state.Tier = 2; setCash(t, state, "1e10") }, carry: &replayFounderCarry{FounderRevision: 1, FounderConstantsHash: catalogs.ConstantsHash, NetworkSlots: []save.NetworkSlot{}, LedgerFactKinds: []string{}, ExitHistoryCount: 0}},
		{name: "cross-gate-offer-spawn", payload: `{"intent_id":"01986666-0011-7000-8000-000000000011","kind":"cross_gate","expected_revision":1,"gate_id":"gate.t2_to_t3","route_id":null}`, mode: ModeOnline,
			configure: func(state *save.State) {
				state.Tier = 2
				state.LifetimeValue = decimal.New(8, 12)
				setCash(t, state, "1e10")
			},
			carry:     &replayFounderCarry{FounderRevision: 1, FounderConstantsHash: catalogs.ConstantsHash, NetworkSlots: []save.NetworkSlot{}, LedgerFactKinds: []string{}, ExitHistoryCount: 1},
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
		{name: "skip-ahead-gate", payload: `{"intent_id":"01986666-0016-7000-8000-000000000016","kind":"cross_gate","expected_revision":1,"gate_id":"gate.t2_to_t3","route_id":null}`, mode: ModeOnline, configure: func(state *save.State) { state.Tier = 1; setCash(t, state, "1e10") }, carry: &replayFounderCarry{FounderRevision: 1, FounderConstantsHash: catalogs.ConstantsHash, NetworkSlots: []save.NetworkSlot{}, LedgerFactKinds: []string{}, ExitHistoryCount: 0}},
		{name: "lower-gate-after-higher", payload: `{"intent_id":"01986666-0017-7000-8000-000000000017","kind":"cross_gate","expected_revision":1,"gate_id":"gate.t2_to_t3","route_id":null}`, mode: ModeOnline, configure: func(state *save.State) { state.Tier = 4; setCash(t, state, "1e10") }, carry: &replayFounderCarry{FounderRevision: 1, FounderConstantsHash: catalogs.ConstantsHash, NetworkSlots: []save.NetworkSlot{}, LedgerFactKinds: []string{}, ExitHistoryCount: 0}},
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
	result := crossRuntimeFixture{Version: 1, ConstantsHash: catalogs.ConstantsHash, Artifacts: artifacts, Cases: make([]crossRuntimeFixtureCase, 0, len(cases))}
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
		build.OfflineCatchup, err = buildOfflineCatchup(state, catalogs.Economy, definition.mode, baseNow.Add(definition.advance))
		if err != nil {
			t.Fatalf("%s offline catchup: %v", definition.name, err)
		}
		inputs, err := buildReplayInputs(build)
		if err != nil {
			t.Fatalf("%s inputs: %v", definition.name, err)
		}
		if definition.name == "manual-online" || definition.name == "buy-generator" {
			var historical replayInputsWire
			if err := decodeReplayStrict(inputs, &historical); err != nil {
				t.Fatal(err)
			}
			if definition.name == "manual-online" {
				historical.Version = 5
			} else {
				historical.Version = 6
			}
			inputs, err = json.Marshal(historical)
			if err != nil {
				t.Fatal(err)
			}
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
		makeTerminalFixtureCase(t, catalogs, catalogs.ConstantsHash, baseNow),
		makeAcceptedOfferFixtureCase(t, catalogs, catalogs.ConstantsHash, baseNow),
		makeScriptedGateFixtureCase(t, catalogs, catalogs.ConstantsHash, baseNow),
	}
	result.Additional = append([]crossRuntimeBundleCase{makeFallbackInvariantFixture(t, catalogs.Artifacts, baseNow)}, makeFoundationFixtures(t, catalogs.Artifacts, baseNow)...)
	result.Additional = append(result.Additional,
		makeActiveFoundationReplayFixture(t, baseNow, 5001*time.Millisecond, "active-foundation-offline-5001ms"),
		makeActiveFoundationReplayFixture(t, baseNow, 25*time.Hour, "active-foundation-offline-25h"),
		makeActiveFoundationBandReplayFixture(t, baseNow),
		makeActiveFoundationBurnReplayFixture(t, baseNow),
	)
	result.ActiveExit = makeActiveFoundationExitFixture(t, baseNow)
	result.ActivePlayExit = makeActivePlayExitFixture(t, baseNow)
	result.FirstContentExit = makeFirstContentExitFixture(t, baseNow)
	result.FullRun = makeFullRunFixture(t, catalogs, catalogs.ConstantsHash, baseNow)
	result.DoctrineRun = makeDoctrineReplayRunFixture(t, baseNow)
	result.ActivePlayRun = makeActivePlayReplayRunFixture(t, baseNow)
	result.RejectedExit = makeRejectedExitFixture(t, catalogs, catalogs.ConstantsHash, baseNow)
	result.FounderHash, result.FounderFiles, result.FounderCases, result.FounderRun = makeFounderReplayFixture(t, baseNow)
	result.PetFounderHash, result.PetFounderFiles, result.PetFounderCases = makePetFounderReplayFixture(t, baseNow)
	result.MinigameHash, result.MinigameFiles, result.MinigameCompany, result.MinigameFounder = makeMinigameResolutionReplayFixture(t, baseNow)
	result.SoulHash, result.SoulFiles, result.SoulCompany, result.SoulFounder = makeSoulRecoveryReplayFixture(t, baseNow)
	result.MinigameStartHash, result.MinigameStartFiles, result.MinigameStart, result.MinigameExitReset, result.MinigameActiveExit = makeMinigameStartReplayFixture(t, baseNow)
	result.MinigameActivation = makeMinigameActivationFixture(t, baseNow)
	return result
}

func makeMinigameActivationFixture(t *testing.T, now time.Time) crossRuntimeFounderActivation {
	t.Helper()
	_, current := foundationTestBundles(t)
	artifact, err := os.ReadFile("../../testdata/minigame/pitch-v3.json")
	if err != nil {
		t.Fatal(err)
	}
	next := current
	next.Artifacts = cloneArtifactMap(current.Artifacts)
	next.Artifacts["minigames"] = artifact
	next.Minigames, err = minigame.LoadCatalog(artifact)
	if err != nil {
		t.Fatal(err)
	}
	next.ConstantsHash, err = save.ConstantsHashArtifacts(next.Artifacts)
	if err != nil || !next.valid(next.ConstantsHash) {
		t.Fatalf("minigame activation fixture bundle err=%v", err)
	}

	founder := replayFounderFixtureState(t, current, now)
	founder.WireVersion = 16
	founder.AgeMS = 12_345
	pre := mustEncodeState(t, founder)
	current.Next = &next
	command := save.FounderReplayCommand{IntentID: "01986666-4b00-7000-8000-000000000001", FounderStreamID: "01986666-4b00-4000-8000-000000000002",
		FounderID: "01986666-4b00-4000-8000-000000000003", Revision: 1, FounderLogSeq: 1, ServerTSMS: now.UnixMilli()}
	resolved := founderExitResolvedWire{Kind: founderExitResolvedKind, Outcome: string(save.IntentApplied), CompanyStreamID: "01986666-4b00-4000-8000-000000000004",
		RunSeq: 1, RunLogSeq: 1, ResultConstantsHash: next.ConstantsHash, AgeMSBefore: founder.AgeMS, AgeMSAfter: founder.AgeMS,
		AddedNetworkSlots: []save.NetworkSlot{}, AddedLedgerFactKinds: []string{}, AddedLifetimeAchievements: []string{},
		ExitRecord: &founderExitRecordWire{RunID: 1, ExitType: "collapse", OccurredAtMS: now.UnixMilli()}, ResultFounderWireVersion: 17}
	inputs, err := save.MarshalFounderReplayInputs(command, resolved)
	if err != nil {
		t.Fatal(err)
	}
	payload := []byte(`{"kind":"wind_down","expected_revision":1,"expected_founder_revision":1}`)
	transition, err := applyFounderExitResolved(founder, command, IntentRequest{}, current, resolved)
	if err != nil {
		t.Fatal(err)
	}
	post := mustEncodeState(t, transition.State)
	events := fixtureEvents(transition.Events)
	caseValue := crossRuntimeFounderCase{Name: "activate-content-bearing-minigame-catalog", StateVersion: 16, PreState: pre, CanonicalPayload: payload,
		ReplayInputs: inputs, Outcome: string(transition.Outcome), Receipt: transition.Receipt, Events: events, PostState: post,
		ResultConstantsHash: transition.ResultConstantsHash, ReceiptJSON: canonicalFixtureJSON(t, transition.Receipt),
		EventsJSON: canonicalFixtureValue(t, events), PostStateJSON: canonicalFixtureJSON(t, post)}
	return crossRuntimeFounderActivation{ConstantsHash: current.ConstantsHash, Artifacts: artifactStrings(current.Artifacts),
		NextConstantsHash: next.ConstantsHash, NextArtifacts: artifactStrings(next.Artifacts), Case: caseValue}
}

func artifactStrings(artifacts map[string][]byte) map[string]string {
	result := make(map[string]string, len(artifacts))
	for name, data := range artifacts {
		result[name] = string(data)
	}
	return result
}

func makeMinigameStartReplayFixture(t *testing.T, now time.Time) (string, map[string]string, crossRuntimeFounderCase, crossRuntimeFounderCase, crossRuntimeFixtureCase) {
	t.Helper()
	catalogs := pitchFeatureBundle(t)
	apiBytes, err := os.ReadFile("../../balance/testdata/minigame-api-candidate-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	catalogs.Artifacts = cloneArtifactMap(catalogs.Artifacts)
	catalogs.Artifacts["minigame_api"] = apiBytes
	catalogs.MinigameAPI, err = minigameapi.LoadCatalog(apiBytes)
	if err != nil {
		t.Fatal(err)
	}
	catalogs.ConstantsHash, err = save.ConstantsHashArtifacts(catalogs.Artifacts)
	if err != nil || !catalogs.valid(catalogs.ConstantsHash) {
		t.Fatalf("minigame start fixture bundle err=%v", err)
	}
	state := replayFounderFixtureState(t, catalogs, now)
	state.WireVersion, state.MinigameSessionSeq = 21, 7
	state.MinigameRatings = map[string]save.MinigameRatingState{"pitch": {Elo: 1000, SeasonMember: "s1"}}
	state.MinigameOfflineQuality = map[string]save.MinigameOfflineQualityState{"pitch": {GradePPM: 200_000}}
	state.Pets = map[string]pet.CareState{}
	state.FiscalPeriodOpenedWallMS, state.FiscalGeneratorLevels, state.FiscalUnlocks = now.UnixMilli(), map[string]int64{}, map[string]bool{"minigame.pitch": true}
	for _, row := range catalogs.Fiscal.GeneratorLevelRows() {
		state.FiscalGeneratorLevels[row.GeneratorID] = 0
	}
	state.Soul, state.SoulExhaustedSourceIDs = 50, []string{}
	pre := mustEncodeState(t, state)
	const sessionID = "01986666-2b00-7000-8000-000000000001"
	const startIntentID = "01986666-2b01-7000-8000-000000000001"
	const founderID = "01986666-3b00-7000-8000-000000000001"
	payload, _ := json.Marshal(startMinigameSessionPayload{Kind: startMinigameSessionKind, SessionID: sessionID, MinigameID: "pitch"})
	command := save.FounderReplayCommand{IntentID: startIntentID, FounderStreamID: "01986666-2c00-4000-8000-000000000001",
		FounderID: founderID, Revision: 1, FounderLogSeq: 1, ServerTSMS: now.UnixMilli()}
	resolved := startMinigameSessionResolved{Kind: startMinigameSessionKind, CompanyStreamID: "01986666-1b00-4000-8000-000000000001",
		RunSeq: 1, SequenceBefore: 7, SequenceAfter: 8, Seed: minigameSessionSeed(founderID, 1, 8)}
	inputs, err := save.MarshalFounderReplayInputs(command, resolved)
	if err != nil {
		t.Fatal(err)
	}
	transition, err := ApplyFounderLogged(state, payload, catalogs, inputs)
	if err != nil {
		t.Fatal(err)
	}
	post := mustEncodeState(t, state)
	events := fixtureEvents(transition.Events)
	founderCase := crossRuntimeFounderCase{
		Name: "start-minigame-session-founder", StateVersion: 21, PreState: pre, CanonicalPayload: payload,
		ReplayInputs: inputs, Outcome: string(transition.Outcome), Receipt: transition.Receipt, Events: events,
		PostState: post, ResultConstantsHash: transition.ResultConstantsHash,
		ReceiptJSON: canonicalFixtureJSON(t, transition.Receipt), EventsJSON: canonicalFixtureValue(t, events),
		PostStateJSON: canonicalFixtureJSON(t, post),
	}

	company := replayFixtureState(t, catalogs.Economy, now)
	company.WireVersion, company.MeterBands = 16, nil
	meterState, meterErr := meters.NewRunState(catalogs.Meters, 0)
	if meterErr != nil {
		t.Fatal(meterErr)
	}
	company.MeterValues, company.MeterDecayRemainders, company.MeterInputRemainders =
		meterState.Values, meterState.DecayRemainders, meterState.InputRemainders
	company.AchievementsEarnedRun = map[string]bool{}
	companyPre := mustEncodeState(t, company)
	exitRequest, err := ParseIntent([]byte(`{"intent_id":"01986666-2b00-7000-8000-000000000002","kind":"wind_down","expected_revision":1,"expected_founder_revision":1}`))
	if err != nil {
		t.Fatal(err)
	}
	exitPayload := json.RawMessage(exitRequest.CanonicalPayload)
	exitCommand := save.ReplayCommand{IntentID: exitRequest.IntentID, CompanyStreamID: resolved.CompanyStreamID,
		FounderID: founderID, Revision: 1, RunSeq: 1, RunLogSeq: 1}
	carry := founderCarry(state)
	carry.FounderRevision, carry.FounderConstantsHash = 1, catalogs.ConstantsHash
	active := true
	exitInputs, err := buildReplayInputs(replayBuild{Command: exitCommand, Mode: ModeOnline, Now: now,
		IntentKind: exitRequest.Kind, RouteContextVersion: catalogs.Routes.ContextVersion(), FounderCarry: &carry,
		Terminal: true, ExecutedRouteIDs: []string{}, SelectedExitType: "scripted_first",
		SelectedTerms: json.RawMessage(`{}`), NextConstantsHash: catalogs.ConstantsHash,
		MinigameSessionActive: &active})
	if err != nil {
		t.Fatal(err)
	}
	exitTransition, err := ApplyLoggedExit(company, exitPayload, catalogs, exitInputs)
	if err != nil || exitTransition.Decision.Outcome != save.IntentRejected {
		t.Fatalf("active minigame Exit outcome=%s err=%v", exitTransition.Decision.Outcome, err)
	}
	companyPost := mustEncodeState(t, company)
	exitEvents := []fixtureEvent{}
	exitCase := crossRuntimeFixtureCase{Name: "minigame-active-exit-rejected", PreState: companyPre,
		CanonicalPayload: exitPayload, ReplayInputs: exitInputs, Outcome: string(exitTransition.Decision.Outcome),
		Receipt: exitTransition.Decision.Receipt, Events: exitEvents, PostState: companyPost,
		ReceiptJSON: canonicalFixtureJSON(t, exitTransition.Decision.Receipt), EventsJSON: canonicalFixtureValue(t, exitEvents),
		PostStateJSON: canonicalFixtureJSON(t, companyPost)}
	resetPre := mustEncodeState(t, state)
	resetCommand := save.FounderReplayCommand{IntentID: "01986666-2b02-7000-8000-000000000001",
		FounderStreamID: command.FounderStreamID, FounderID: founderID, Revision: 2, FounderLogSeq: 2, ServerTSMS: now.Add(time.Second).UnixMilli()}
	resetResolved := founderExitResolvedWire{Kind: founderExitResolvedKind, Outcome: string(save.IntentApplied),
		CompanyStreamID: resolved.CompanyStreamID, RunSeq: 1, RunLogSeq: 2, ResultConstantsHash: catalogs.ConstantsHash,
		AgeMSBefore: state.AgeMS, AgeMSAfter: state.AgeMS, AddedNetworkSlots: []save.NetworkSlot{}, AddedLedgerFactKinds: []string{},
		AddedLifetimeAchievements: []string{}, ExitRecord: &founderExitRecordWire{RunID: 1, ExitType: "collapse",
			OccurredAtMS: resetCommand.ServerTSMS}, ResultFounderWireVersion: 21}
	resetInputs, err := save.MarshalFounderReplayInputs(resetCommand, resetResolved)
	if err != nil {
		t.Fatal(err)
	}
	resetRequest, err := ParseIntent([]byte(`{"intent_id":"01986666-2b02-7000-8000-000000000001","kind":"wind_down","expected_revision":1,"expected_founder_revision":2}`))
	if err != nil {
		t.Fatal(err)
	}
	resetPayload := json.RawMessage(resetRequest.CanonicalPayload)
	resetTransition, err := ApplyFounderLogged(state, resetPayload, catalogs, resetInputs)
	if err != nil {
		t.Fatal(err)
	}
	resetPost := mustEncodeState(t, state)
	resetEvents := fixtureEvents(resetTransition.Events)
	resetCase := crossRuntimeFounderCase{Name: "exit-resets-minigame-session-sequence", StateVersion: 21,
		PreState: resetPre, CanonicalPayload: resetPayload, ReplayInputs: resetInputs, Outcome: string(resetTransition.Outcome),
		Receipt: resetTransition.Receipt, Events: resetEvents, PostState: resetPost,
		ResultConstantsHash: resetTransition.ResultConstantsHash, ReceiptJSON: canonicalFixtureJSON(t, resetTransition.Receipt),
		EventsJSON: canonicalFixtureValue(t, resetEvents), PostStateJSON: canonicalFixtureJSON(t, resetPost)}
	if state.MinigameSessionSeq != 0 {
		t.Fatalf("v21 Exit preserved minigame session sequence %d", state.MinigameSessionSeq)
	}
	return catalogs.ConstantsHash, stringArtifacts(catalogs.Artifacts), founderCase, resetCase, exitCase
}

func makeDoctrineReplayRunFixture(t *testing.T, now time.Time) crossRuntimeFullRun {
	t.Helper()
	catalogs := doctrineReplayBundle(t)
	artifacts, hash := catalogs.Artifacts, catalogs.ConstantsHash
	state := replayFixtureState(t, catalogs.Economy, now)
	state.WireVersion, state.Tier, state.ComputeCreditMS = 17, 3, 5_000
	state.GeneratorCounts["generator.beige_tower"] = 1
	setCash(t, state, "1e13")
	meterState, err := meters.NewRunState(catalogs.Meters, 0)
	if err != nil {
		t.Fatal(err)
	}
	state.MeterBands = nil
	state.MeterValues, state.MeterDecayRemainders, state.MeterInputRemainders = meterState.Values, meterState.DecayRemainders, meterState.InputRemainders
	state.AchievementsEarnedRun = map[string]bool{}
	genesis := mustEncodeState(t, state)
	definitions := []struct {
		payload string
		at      time.Duration
	}{
		{payload: `{"intent_id":"01986666-0a01-7000-8000-000000000001","kind":"cross_gate","expected_revision":1,"gate_id":"gate.t3_to_t4","route_id":null}`},
		{payload: `{"intent_id":"01986666-0a02-7000-8000-000000000002","kind":"pick_doctrine","expected_revision":1,"transition_id":"transition.t3_to_t4","doctrine_id":"doctrine.capture"}`},
		{payload: `{"intent_id":"01986666-0a03-7000-8000-000000000003","kind":"pick_doctrine","expected_revision":2,"transition_id":"transition.t3_to_t4","doctrine_id":"doctrine.ethical"}`},
		{payload: `{"intent_id":"01986666-0a04-7000-8000-000000000004","kind":"spend_compute_credit","expected_revision":2,"amount_ms":1500,"target":"accelerate"}`},
		{payload: `{"intent_id":"01986666-0a05-7000-8000-000000000005","kind":"spend_compute_credit","expected_revision":3,"amount_ms":1000,"target":"accelerate"}`},
		{payload: `{"intent_id":"01986666-0a06-7000-8000-000000000006","kind":"perform_manual_batch","expected_revision":3,"action_id":"manual.click","count":1,"window_ms":1}`, at: 500 * time.Millisecond},
		{payload: `{"intent_id":"01986666-0a07-7000-8000-000000000007","kind":"perform_manual_batch","expected_revision":4,"action_id":"manual.click","count":1,"window_ms":1}`, at: 1500 * time.Millisecond},
		{payload: `{"intent_id":"01986666-0a08-7000-8000-000000000008","kind":"spend_compute_credit","expected_revision":5,"amount_ms":4000,"target":"accelerate"}`, at: 1500 * time.Millisecond},
		{payload: `{"intent_id":"01986666-0a09-7000-8000-000000000009","kind":"spend_compute_credit","expected_revision":5,"amount_ms":14400001,"target":"accelerate"}`, at: 1500 * time.Millisecond},
		{payload: `{"intent_id":"01986666-0a10-7000-8000-000000000010","kind":"spend_compute_credit","expected_revision":5,"amount_ms":0,"target":"accelerate"}`, at: 1500 * time.Millisecond},
		{payload: `{"intent_id":"01986666-0a11-7000-8000-000000000011","kind":"cross_gate","expected_revision":5,"gate_id":"gate.t3_to_t4","route_id":null}`, at: 1500 * time.Millisecond},
	}
	carry := replayFounderCarry{FounderRevision: 1, FounderConstantsHash: hash, NetworkSlots: []save.NetworkSlot{}, LedgerFactKinds: []string{}, AchievementsEarnedLifetime: []string{}}
	revision := int64(1)
	entries := make([]crossRuntimeFullRunEntry, 0, len(definitions))
	for index, definition := range definitions {
		request, parseErr := ParseIntent([]byte(definition.payload))
		if parseErr != nil {
			t.Fatal(parseErr)
		}
		if request.ExpectedRevision != revision {
			t.Fatalf("doctrine fixture expected revision=%d live=%d", request.ExpectedRevision, revision)
		}
		command := save.ReplayCommand{IntentID: request.IntentID, CompanyStreamID: "01986666-1a00-7000-8000-000000000001", FounderID: "01986666-2a00-7000-8000-000000000001", Revision: revision, RunSeq: 1, RunLogSeq: int64(index + 1)}
		inputs, inputErr := buildReplayInputs(replayBuild{Command: command, Mode: ModeOnline, Now: now.Add(definition.at), IntentKind: request.Kind, RouteContextVersion: catalogs.Routes.ContextVersion(), FounderCarry: &carry})
		if inputErr != nil {
			t.Fatal(inputErr)
		}
		transition, transitionErr := ApplyLogged(state, request.CanonicalPayload, catalogs, inputs)
		if transitionErr != nil {
			t.Fatalf("doctrine fixture row %d: %v", index+1, transitionErr)
		}
		if transition.Outcome == save.IntentApplied {
			revision++
		}
		events := fixtureEvents(transition.Events)
		entries = append(entries, crossRuntimeFullRunEntry{Seq: int64(index + 1), CanonicalPayload: request.CanonicalPayload, ReplayInputs: inputs,
			ReceiptJSON: canonicalFixtureJSON(t, transition.Receipt), EventsJSON: canonicalFixtureValue(t, events)})
	}
	finalState := mustEncodeState(t, state)
	return crossRuntimeFullRun{ConstantsHash: hash, Artifacts: stringArtifacts(artifacts), Genesis: genesis, Entries: entries, FinalStateJSON: canonicalFixtureJSON(t, finalState)}
}

func doctrineReplayBundle(t *testing.T) CatalogBundle {
	t.Helper()
	_, catalogs := foundationTestBundles(t)
	artifacts := cloneArtifactMap(catalogs.Artifacts)
	var routeRoot map[string]any
	if err := json.Unmarshal(artifacts["routes"], &routeRoot); err != nil {
		t.Fatal(err)
	}
	gates := routeRoot["gates"].([]any)
	doctrineGate := map[string]any{"gate_id": "gate.t3_to_t4", "requirement": []any{map[string]any{"resource_id": "company.cash", "amount": "1e12"}}, "routes": []any{}}
	routeRoot["gates"] = append(append(append([]any{}, gates[:1]...), doctrineGate), gates[1:]...)
	routeArtifact, err := json.Marshal(routeRoot)
	if err != nil {
		t.Fatal(err)
	}
	var categoryRoot map[string]any
	if err := json.Unmarshal(artifacts["categories"], &categoryRoot); err != nil {
		t.Fatal(err)
	}
	categoryGates := categoryRoot["full_gate_set"].([]any)
	categoryRoot["full_gate_set"] = append(append(append([]any{}, categoryGates[:1]...), "gate.t3_to_t4"), categoryGates[1:]...)
	categoryArtifact, err := json.Marshal(categoryRoot)
	if err != nil {
		t.Fatal(err)
	}
	doctrineArtifact := []byte(`{"schema_version":1,"transitions":[{"transition_id":"transition.t3_to_t4","source_tier":3,"gate_id":"gate.t3_to_t4","doctrine_ids":["doctrine.capture","doctrine.ethical"]}]}`)
	routeCatalog, err := routes.LoadCatalog(routeArtifact)
	if err != nil {
		t.Fatal(err)
	}
	doctrineCatalog, err := doctrine.LoadCatalog(doctrineArtifact)
	if err != nil {
		t.Fatal(err)
	}
	artifacts["categories"], artifacts["routes"], artifacts["doctrines"] = categoryArtifact, routeArtifact, doctrineArtifact
	hash, err := save.ConstantsHashArtifacts(artifacts)
	if err != nil {
		t.Fatal(err)
	}
	catalogs.Artifacts, catalogs.ConstantsHash, catalogs.Routes, catalogs.Doctrines = artifacts, hash, routeCatalog, doctrineCatalog
	if !catalogs.valid(hash) {
		t.Fatal("doctrine fixture bundle is not internally valid")
	}
	return catalogs
}

func makeActivePlayReplayRunFixture(t *testing.T, now time.Time) crossRuntimeFullRun {
	t.Helper()
	catalogs := activePlayReplayBundle(t)
	state := replayFixtureState(t, catalogs.Economy, now)
	state.WireVersion, state.Tier, state.ComputeCreditMS = 18, 3, 5_000
	state.GeneratorCounts["generator.beige_tower"] = 100
	setCash(t, state, "9.99999999999e999")
	meterState, err := meters.NewRunState(catalogs.Meters, 0)
	if err != nil {
		t.Fatal(err)
	}
	state.MeterBands = nil
	state.MeterValues, state.MeterDecayRemainders, state.MeterInputRemainders = meterState.Values, meterState.DecayRemainders, meterState.InputRemainders
	state.AchievementsEarnedRun = map[string]bool{}
	founderID := activePlayFixtureFounder(t, catalogs.Opportunities)
	initial, err := initializeActivePlayState(state, catalogs.Opportunities, founderID)
	if err != nil {
		t.Fatal(err)
	}
	genesis := mustEncodeState(t, state)
	hash := catalogs.ConstantsHash
	carry := replayFounderCarry{FounderRevision: 1, FounderConstantsHash: hash, NetworkSlots: []save.NetworkSlot{}, LedgerFactKinds: []string{}, AchievementsEarnedLifetime: []string{}}
	revision := int64(1)
	sequence := int64(0)
	entries := make([]crossRuntimeFullRunEntry, 0, 12)
	appendCommand := func(kind, fields string, at time.Time) save.IntentOutcome {
		t.Helper()
		sequence++
		intentID := fmt.Sprintf("01987777-%04x-7000-8000-%012x", sequence, sequence)
		payload := fmt.Sprintf(`{"intent_id":%q,"kind":%q,"expected_revision":%d%s}`, intentID, kind, revision, fields)
		request, parseErr := ParseIntent([]byte(payload))
		if parseErr != nil {
			t.Fatalf("active fixture row %d parse: %v", sequence, parseErr)
		}
		command := save.ReplayCommand{IntentID: request.IntentID, CompanyStreamID: "01987777-1a00-7000-8000-000000000001", FounderID: founderID,
			Revision: revision, RunSeq: state.RunSeq, RunLogSeq: sequence}
		activeEvidence, resolveErr := resolveActivePlaySchedule(state, catalogs.Opportunities, catalogs.Prestige, founderID, at)
		if resolveErr != nil {
			t.Fatalf("active fixture row %d schedule: %v", sequence, resolveErr)
		}
		build := replayBuild{Command: command, Mode: ModeOnline, Now: at, IntentKind: request.Kind, RouteContextVersion: catalogs.Routes.ContextVersion(),
			FounderCarry: &carry, ActivePlay: &activeEvidence}
		if request.Kind == IntentClaimOpportunity {
			accrual, accrualErr := makeReplayAccrual(nil, nil, guild.SettlementBatch{}, build.RouteContextVersion)
			if accrualErr != nil {
				t.Fatal(accrualErr)
			}
			claim, claimErr := resolveActiveClaimEvidence(state, catalogs, save.Revision{StreamID: command.CompanyStreamID, OwnerID: founderID,
				Number: revision, ConstantsHash: hash, RunLogSequence: sequence}, request, ModeOnline, at, accrual, activeEvidence)
			if claimErr != nil {
				t.Fatalf("active fixture row %d claim: %v", sequence, claimErr)
			}
			build.ActivePlay.Claim = claim
		}
		inputs, inputErr := buildReplayInputs(build)
		if inputErr != nil {
			t.Fatalf("active fixture row %d inputs: %v", sequence, inputErr)
		}
		transition, transitionErr := ApplyLogged(state, request.CanonicalPayload, catalogs, inputs)
		if transitionErr != nil {
			t.Fatalf("active fixture row %d transition: %v", sequence, transitionErr)
		}
		if transition.Outcome == save.IntentApplied {
			revision++
		}
		entries = append(entries, crossRuntimeFullRunEntry{Seq: sequence, CanonicalPayload: request.CanonicalPayload, ReplayInputs: inputs,
			ReceiptJSON: canonicalFixtureJSON(t, transition.Receipt), EventsJSON: canonicalFixtureValue(t, fixtureEvents(transition.Events))})
		return transition.Outcome
	}
	manual := func(at time.Time) save.IntentOutcome {
		return appendCommand(IntentPerformManualBatch, `,"action_id":"manual.click","count":1,"window_ms":1`, at)
	}
	claim := func(id string, at time.Time) save.IntentOutcome {
		return appendCommand(IntentClaimOpportunity, fmt.Sprintf(`,"opportunity_id":%q`, id), at)
	}
	started := now
	firstAt := started.Add(time.Duration(initial.SpawnedAttendedMS) * time.Millisecond)
	if manual(firstAt) != save.IntentApplied || state.PendingOpportunity == nil || state.PendingOpportunity.EffectRowID != "active.production" {
		t.Fatal("first production opportunity did not materialize")
	}
	if claim(state.PendingOpportunity.OpportunityID, firstAt) != save.IntentApplied {
		t.Fatal("first production opportunity was not claimed")
	}
	secondAt := started.Add(time.Duration(state.NextOpportunityAttendedMS) * time.Millisecond)
	if manual(secondAt) != save.IntentApplied || state.PendingOpportunity == nil || state.PendingOpportunity.EffectRowID != "active.building" {
		t.Fatal("building-special opportunity did not materialize")
	}
	if claim(state.PendingOpportunity.OpportunityID, secondAt) != save.IntentApplied || len(state.ActiveBuffs) != 2 {
		t.Fatal("overlapping production buffs were not recorded")
	}
	if !strings.Contains(entries[len(entries)-1].ReceiptJSON, `"cap_reason_key":"cap.active_combo"`) ||
		!strings.Contains(entries[len(entries)-1].EventsJSON, `"hardcap_reason_key":"cap.active_combo"`) {
		t.Fatal("cross-target saturation did not surface its combo hardcap reason")
	}
	if manual(secondAt.Add(time.Millisecond)) != save.IntentApplied {
		t.Fatal("overlapping buff command did not apply")
	}
	thirdAt := started.Add(time.Duration(state.NextOpportunityAttendedMS) * time.Millisecond)
	if manual(thirdAt) != save.IntentApplied || state.PendingOpportunity == nil || state.PendingOpportunity.EffectRowID != "active.lucky" {
		t.Fatal("Lucky opportunity did not materialize")
	}
	if claim(state.PendingOpportunity.OpportunityID, thirdAt) != save.IntentApplied {
		t.Fatal("Lucky opportunity was not claimed")
	}
	fourthAt := started.Add(time.Duration(state.NextOpportunityAttendedMS) * time.Millisecond)
	if manual(fourthAt) != save.IntentApplied || state.PendingOpportunity == nil {
		t.Fatal("miss candidate did not materialize")
	}
	missedID := state.PendingOpportunity.OpportunityID
	offlineReturn := fourthAt.Add(time.Hour)
	if manual(offlineReturn) != save.IntentApplied || state.PendingOpportunity == nil || state.PendingOpportunity.OpportunityID != missedID {
		t.Fatal("offline wall gap advanced the attended scheduler")
	}
	expiryAt := offlineReturn.Add(time.Duration(catalogs.Opportunities.Schedule.LifetimeMS) * time.Millisecond)
	if manual(expiryAt.Add(-time.Millisecond)) != save.IntentApplied || state.PendingOpportunity == nil || state.PendingOpportunity.OpportunityID != missedID {
		t.Fatal("pre-expiry command did not preserve the pending opportunity")
	}
	if claim(missedID, expiryAt) != save.IntentRejected || state.PendingOpportunity == nil || state.PendingOpportunity.OpportunityID != missedID {
		t.Fatal("expired claim did not reject and roll back scheduler cleanup")
	}
	nextAfterMiss, err := catalogs.Opportunities.Spawn(founderID, state.RunSeq, state.OpportunitySpawnSeq, state.PendingOpportunity.ExpiresAttendedMS)
	if err != nil {
		t.Fatal(err)
	}
	compoundAt := expiryAt.Add(time.Duration(nextAfterMiss.SampledIntervalMS) * time.Millisecond)
	compoundOutcome := manual(compoundAt)
	if compoundOutcome != save.IntentApplied || state.PendingOpportunity == nil || state.PendingOpportunity.OpportunityID != nextAfterMiss.OpportunityID {
		t.Fatalf("next applied command did not persist compound miss+spawn: outcome=%s pending=%+v expected=%+v", compoundOutcome, state.PendingOpportunity, nextAfterMiss)
	}
	if claim(state.PendingOpportunity.OpportunityID, compoundAt) != save.IntentApplied || state.PendingOpportunity != nil {
		t.Fatal("compound successor claim did not schedule the self-miss fixture")
	}
	untilNext := state.NextOpportunityAttendedMS - nextAfterMiss.SpawnedAttendedMS
	if untilNext <= 0 {
		t.Fatal("invalid self-miss fixture interval")
	}
	preSpawnAt := compoundAt.Add(time.Duration(untilNext-1) * time.Millisecond)
	if manual(preSpawnAt) != save.IntentApplied || state.PendingOpportunity != nil {
		t.Fatal("pre-spawn command materialized the self-miss opportunity early")
	}
	selfMissAt := preSpawnAt.Add(time.Duration(catalogs.Opportunities.Schedule.LifetimeMS+1) * time.Millisecond)
	if manual(selfMissAt) != save.IntentApplied || state.PendingOpportunity != nil || state.NextOpportunityAttendedMS <= 0 {
		t.Fatal("spawn-then-self-miss compound transition did not persist")
	}
	finalState := mustEncodeState(t, state)
	return crossRuntimeFullRun{ConstantsHash: hash, Artifacts: stringArtifacts(catalogs.Artifacts), Genesis: genesis, Entries: entries,
		FinalStateJSON: canonicalFixtureJSON(t, finalState)}
}

func activePlayReplayBundle(t *testing.T) CatalogBundle {
	t.Helper()
	catalogs := doctrineReplayBundle(t)
	artifacts := cloneArtifactMap(catalogs.Artifacts)
	var economyRoot map[string]any
	if err := json.Unmarshal(artifacts["economy"], &economyRoot); err != nil {
		t.Fatal(err)
	}
	sources := economyRoot["multiplier_sources"].([]any)
	economyRoot["multiplier_sources"] = append(sources,
		map[string]any{"id": "active.building.generator.beige_tower", "slot": "event_buffs", "target": "generator.beige_tower", "provider": "active_play"},
		map[string]any{"id": "active.click", "slot": "event_buffs", "target": "manual.click", "provider": "active_play"},
		map[string]any{"id": "active.production", "slot": "event_buffs", "target": "all", "provider": "active_play"},
	)
	economyArtifact, err := json.Marshal(economyRoot)
	if err != nil {
		t.Fatal(err)
	}
	economyCatalog, err := economy.LoadCatalog(economyArtifact)
	if err != nil {
		t.Fatal(err)
	}
	fixture, err := os.ReadFile("../../balance/testdata/active-play-foundation-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	var root struct {
		Baseline map[string]any `json:"baseline"`
	}
	if err := json.Unmarshal(fixture, &root); err != nil {
		t.Fatal(err)
	}
	root.Baseline["combo_policy"].(map[string]any)["combo_cap"] = "1e1"
	root.Baseline["schedule_policy"].(map[string]any)["minimum_interval_ms"] = float64(2_500)
	opportunitiesArtifact, err := json.Marshal(root.Baseline)
	if err != nil {
		t.Fatal(err)
	}
	opportunities, err := activeplay.LoadCatalog(opportunitiesArtifact, economyCatalog)
	if err != nil {
		t.Fatal(err)
	}
	artifacts["economy"], artifacts["opportunities"] = economyArtifact, opportunitiesArtifact
	hash, err := save.ConstantsHashArtifacts(artifacts)
	if err != nil {
		t.Fatal(err)
	}
	catalogs.Artifacts, catalogs.ConstantsHash, catalogs.Economy, catalogs.Opportunities = artifacts, hash, economyCatalog, opportunities
	if !catalogs.valid(hash) {
		t.Fatal("active-play fixture bundle is not internally valid")
	}
	return catalogs
}

func activePlayFixtureFounder(t *testing.T, catalog *activeplay.Catalog) string {
	t.Helper()
	for candidate := int64(1); candidate <= 1_000_000; candidate++ {
		founderID := fmt.Sprintf("01987777-2000-7000-8000-%012d", candidate)
		first, e0 := catalog.Spawn(founderID, 1, 0, 0)
		second, e1 := catalog.Spawn(founderID, 1, 1, first.SpawnedAttendedMS)
		third, e2 := catalog.Spawn(founderID, 1, 2, second.SpawnedAttendedMS)
		fourth, e3 := catalog.Spawn(founderID, 1, 3, third.SpawnedAttendedMS)
		fifth, e4 := catalog.Spawn(founderID, 1, 4, fourth.ExpiresAttendedMS)
		sixth, e5 := catalog.Spawn(founderID, 1, 5, fifth.SpawnedAttendedMS)
		if e0 == nil && e1 == nil && e2 == nil && e3 == nil && e4 == nil && e5 == nil && first.EffectRowID == "active.production" && second.EffectRowID == "active.building" &&
			third.EffectRowID == "active.lucky" && fourth.EffectRowID == "active.click" && first.SampledIntervalMS <= 5_000 && second.SampledIntervalMS < 5_000 &&
			third.SampledIntervalMS <= 5_000 && fourth.SampledIntervalMS <= 5_000 && fifth.SampledIntervalMS < 5_000 && sixth.SampledIntervalMS < 5_000 {
			return founderID
		}
	}
	t.Fatal("no bounded Active-Play fixture seed")
	return ""
}

func makeActivePlayExitFixture(t *testing.T, now time.Time) crossRuntimeActiveExit {
	t.Helper()
	catalogs := activePlayReplayBundle(t)
	founderID := activePlayFixtureFounder(t, catalogs.Opportunities)
	company := replayFixtureState(t, catalogs.Economy, now)
	company.WireVersion, company.Tier = 18, 2
	company.RunStartedAt = now.Add(-20 * time.Minute)
	meterState, err := meters.NewRunState(catalogs.Meters, 0)
	if err != nil {
		t.Fatal(err)
	}
	company.MeterBands = nil
	company.MeterValues, company.MeterDecayRemainders, company.MeterInputRemainders = meterState.Values, meterState.DecayRemainders, meterState.InputRemainders
	company.AchievementsEarnedRun = map[string]bool{}
	if _, err := initializeActivePlayState(company, catalogs.Opportunities, founderID); err != nil {
		t.Fatal(err)
	}
	attended := int64((20 * time.Minute) / time.Millisecond)
	probe, err := catalogs.Opportunities.Spawn(founderID, 1, 0, 0)
	if err != nil || probe.SampledIntervalMS > attended {
		t.Fatal("invalid Active-Play Exit fixture interval")
	}
	pending, err := catalogs.Opportunities.Spawn(founderID, 1, 0, attended-probe.SampledIntervalMS)
	if err != nil || pending.SpawnedAttendedMS != attended {
		t.Fatal("invalid Active-Play Exit fixture pending row")
	}
	company.OpportunitySpawnSeq, company.NextOpportunityAttendedMS = 1, 0
	company.PendingOpportunity = &save.PendingOpportunity{OpportunityID: pending.OpportunityID, SpawnedAttendedMS: pending.SpawnedAttendedMS,
		ExpiresAttendedMS: pending.ExpiresAttendedMS, EffectRowID: pending.EffectRowID, SelectedGeneratorID: activeNullableString(pending.SelectedGenerator)}
	company.ActiveBuffs = []save.ActiveBuff{{BuffInstanceID: catalogs.Opportunities.BuffID(founderID, 1, 0, 0), EffectRowID: "active.production",
		ActivatedAttendedMS: attended - 1_000, ExpiresAttendedMS: attended + 5_000}}
	company.LifetimeValue = decimal.New(8, 12)
	setCash(t, company, "1e10")
	preState := mustEncodeState(t, company)
	request, err := ParseIntent([]byte(`{"intent_id":"01987777-0300-7000-8000-000000000300","kind":"cross_gate","expected_revision":1,"gate_id":"gate.t2_to_t3","route_id":null}`))
	if err != nil {
		t.Fatal(err)
	}
	command := save.ReplayCommand{IntentID: request.IntentID, CompanyStreamID: "01987777-1000-7000-8000-000000000001", FounderID: founderID, Revision: 1, RunSeq: 1, RunLogSeq: 1}
	carry := replayFounderCarry{FounderRevision: 1, FounderConstantsHash: catalogs.ConstantsHash, NetworkSlots: []save.NetworkSlot{}, LedgerFactKinds: []string{}, ExitHistoryCount: 0,
		AchievementsEarnedLifetime: []string{}}
	activeEvidence, err := resolveActivePlaySchedule(company, catalogs.Opportunities, catalogs.Prestige, founderID, now)
	if err != nil {
		t.Fatal(err)
	}
	nextSpawn, err := catalogs.Opportunities.Spawn(founderID, 2, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	inputs, err := buildReplayInputs(replayBuild{Command: command, Mode: ModeOnline, Now: now, IntentKind: request.Kind,
		RouteContextVersion: catalogs.Routes.ContextVersion(), FounderCarry: &carry, Terminal: true, ExecutedRouteIDs: []string{}, SelectedExitType: "scripted_first",
		SelectedTerms: json.RawMessage(`{}`), NextConstantsHash: catalogs.ConstantsHash, ActivePlay: &activeEvidence, NextActivePlay: spawnEvidence(nextSpawn)})
	if err != nil {
		t.Fatal(err)
	}
	result := executeTerminalFixture(t, "active-play-exit-reset", catalogs, company, preState, request, inputs, carry)
	var nextWire map[string]any
	if err := json.Unmarshal(result.NewCompany, &nextWire); err != nil {
		t.Fatal(err)
	}
	if nextWire["pending_opportunity"] != nil || len(nextWire["active_buffs"].([]any)) != 0 || nextWire["opportunity_spawn_seq"] != float64(0) {
		t.Fatal("Exit did not discard prior Active-Play state")
	}
	return crossRuntimeActiveExit{ConstantsHash: catalogs.ConstantsHash, Artifacts: stringArtifacts(catalogs.Artifacts),
		NextConstantsHash: catalogs.ConstantsHash, NextArtifacts: stringArtifacts(catalogs.Artifacts), Case: result}
}

func TestExitResetsComputeBurst(t *testing.T) {
	legacy, _ := foundationTestBundles(t)
	catalogs := doctrineReplayBundle(t)
	startedAt := time.Date(2026, 8, 5, 12, 0, 0, 0, time.UTC)
	founder := foundationScopeState(t, legacy.Economy, economy.ScopeFounder)
	company := foundationScopeState(t, legacy.Economy, economy.ScopeCompany)
	company.RunSeq = 1
	activated, err := prestigecore.NewRunState(catalogs.Economy, company, founder, startedAt)
	if err != nil {
		t.Fatal(err)
	}
	if err := settleAndActivateFoundations(legacy, catalogs, founder, company, activated); err != nil {
		t.Fatal(err)
	}
	activated.ComputeBurstRemainingMS = 60_000
	founderRevision := save.Revision{StreamID: "01986666-2a00-7000-8000-000000000010", OwnerID: "01986666-2a00-7000-8000-000000000011", Number: 1, ConstantsHash: catalogs.ConstantsHash}
	companyRevision := save.Revision{StreamID: "01986666-1a00-7000-8000-000000000010", OwnerID: founderRevision.OwnerID, Number: 1, ConstantsHash: catalogs.ConstantsHash}
	decision, err := finishExitResolved(IntentRequest{IntentID: "01986666-0a12-7000-8000-000000000012"}, founder, founderRevision, activated, companyRevision,
		startedAt.Add(time.Second), "collapse", prestigecore.Terms{}, nil, []string{}, catalogs, catalogs, nil)
	if err != nil {
		t.Fatal(err)
	}
	if activated.ComputeBurstRemainingMS != 0 || decision.NewCompanyState.ComputeBurstRemainingMS != 0 {
		t.Fatalf("Exit retained compute burst: terminal=%d new_run=%d", activated.ComputeBurstRemainingMS, decision.NewCompanyState.ComputeBurstRemainingMS)
	}
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

func makeSoulRecoveryReplayFixture(t *testing.T, now time.Time) (string, map[string]string, crossRuntimeFixtureCase, crossRuntimeFounderCase) {
	t.Helper()
	catalogs := recoverySoulBundle(t)
	const sessionID = "01986666-0950-7000-8000-000000000951"
	const companyStreamID = "01986666-1950-7000-8000-000000000951"
	const founderStreamID = "01986666-2950-7000-8000-000000000951"
	const founderID = "01986666-3950-7000-8000-000000000951"
	payload, _ := json.Marshal(soulRecoveryPayload{Kind: soulRecoveryResolveKind, SessionID: sessionID})
	company := replayFixtureState(t, catalogs.Economy, now)
	company.WireVersion = 16
	company.GeneratorCounts["generator.beige_tower"] = 4
	company.MeterBands = nil
	meterState, err := meters.NewRunState(catalogs.Meters, 0)
	if err != nil {
		t.Fatal(err)
	}
	company.MeterValues, company.MeterDecayRemainders, company.MeterInputRemainders = meterState.Values, meterState.DecayRemainders, meterState.InputRemainders
	company.AchievementsEarnedRun = map[string]bool{}
	companyPre := mustEncodeState(t, company)
	companyCommand := save.ReplayCommand{IntentID: sessionID, CompanyStreamID: companyStreamID, FounderID: founderID,
		Revision: 1, RunSeq: 1, RunLogSeq: 1}
	suppression := soulSuppression{FromEvaluatedMS: now.UnixMilli(), ToEvaluatedMS: now.Add(5 * time.Second).UnixMilli(),
		FounderAttendedStart: 0, FounderAttendedEnd: 5_000, SessionID: sessionID}
	companyInputs, err := buildSoulSuppressionInputs(companyCommand, soulRecoveryResolveKind, suppression, nil, nil,
		catalogs.Routes.ContextVersion(), nil)
	if err != nil {
		t.Fatal(err)
	}
	companyTransition, err := ApplyLogged(company, payload, catalogs, companyInputs)
	if err != nil {
		t.Fatal(err)
	}
	companyPost := mustEncodeState(t, company)
	companyEvents := fixtureEvents(companyTransition.Events)
	companyCase := crossRuntimeFixtureCase{Name: "soul-recovery-company", PreState: companyPre, CanonicalPayload: payload,
		ReplayInputs: companyInputs, Outcome: string(companyTransition.Outcome), Receipt: companyTransition.Receipt,
		Events: companyEvents, PostState: companyPost, ReceiptJSON: canonicalFixtureJSON(t, companyTransition.Receipt),
		EventsJSON: canonicalFixtureValue(t, companyEvents), PostStateJSON: canonicalFixtureJSON(t, companyPost)}

	founder := replayFounderFixtureState(t, catalogs, now)
	founder.WireVersion = 20
	founder.FiscalPeriodOpenedWallMS = now.UnixMilli()
	founder.FiscalGeneratorLevels = make(map[string]int64, len(catalogs.Fiscal.GeneratorLevelRows()))
	for _, row := range catalogs.Fiscal.GeneratorLevelRows() {
		founder.FiscalGeneratorLevels[row.GeneratorID] = 0
	}
	founder.FiscalUnlocks = map[string]bool{}
	founder.Soul = 50
	founder.SoulExhaustedSourceIDs = []string{}
	founderPre := mustEncodeState(t, founder)
	activity, _ := catalogs.Soul.RecoveryActivity("touch_grass.fixture")
	beforeBand, _ := catalogs.Soul.BandFor(50)
	afterBand, _ := catalogs.Soul.BandFor(65)
	founderResolved := founderSoulRecoveryResolved{Kind: "soul_recovery", Action: soulRecoveryResolveKind,
		SessionID: sessionID, ActivityID: activity.ActivityID, CompanyStreamID: companyStreamID, RunSeq: 1,
		FounderAttendedStartMS: 0, FounderAttendedEndMS: 6_000, RecoveryAmount: activity.RecoveryAmount,
		SoulBefore: 50, SoulAfter: 65, BandBefore: beforeBand.Member, BandAfter: afterBand.Member,
		ReasonKey: activity.ReasonKey}
	founderCommand := save.FounderReplayCommand{IntentID: sessionID, FounderStreamID: founderStreamID,
		FounderID: founderID, Revision: 1, FounderLogSeq: 1, ServerTSMS: now.Add(5 * time.Second).UnixMilli()}
	founderInputs, err := save.MarshalFounderReplayInputs(founderCommand, founderResolved)
	if err != nil {
		t.Fatal(err)
	}
	founderTransition, err := ApplyFounderLogged(founder, payload, catalogs, founderInputs)
	if err != nil {
		t.Fatal(err)
	}
	founderPost := mustEncodeState(t, founder)
	founderEvents := fixtureEvents(founderTransition.Events)
	founderCase := crossRuntimeFounderCase{Name: "soul-recovery-founder", StateVersion: 20, PreState: founderPre,
		CanonicalPayload: payload, ReplayInputs: founderInputs, Outcome: string(founderTransition.Outcome),
		Receipt: founderTransition.Receipt, Events: founderEvents, PostState: founderPost,
		ResultConstantsHash: founderTransition.ResultConstantsHash, ReceiptJSON: canonicalFixtureJSON(t, founderTransition.Receipt),
		EventsJSON: canonicalFixtureValue(t, founderEvents), PostStateJSON: canonicalFixtureJSON(t, founderPost)}
	return catalogs.ConstantsHash, stringArtifacts(catalogs.Artifacts), companyCase, founderCase
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

func makeFirstContentExitFixture(t *testing.T, now time.Time) crossRuntimeActiveExit {
	t.Helper()
	catalogs := activeContentBundle(t)
	company := replayFixtureState(t, catalogs.Economy, now.Add(-48*time.Hour))
	company.WireVersion = 18
	company.ActiveBuffs = []save.ActiveBuff{}
	company.MeterBands = nil
	meterState, err := meters.NewRunState(catalogs.Meters, 9)
	if err != nil {
		t.Fatal(err)
	}
	company.MeterValues, company.MeterDecayRemainders, company.MeterInputRemainders = meterState.Values, meterState.DecayRemainders, meterState.InputRemainders
	company.AchievementsEarnedRun = map[string]bool{}
	company.Tier = 1
	if err := catalogs.ValidateFoundationState(company); err != nil {
		t.Fatalf("first-content company: %v", err)
	}
	founder := replayFounderFixtureState(t, catalogs, now)
	founder.WireVersion = 21
	if err := activateMinigameState(founder, catalogs.Minigames); err != nil {
		t.Fatal(err)
	}
	founder.MinigameRatings["pitch"] = save.MinigameRatingState{Elo: 1017, SeasonMember: "s1", GamesCounted: 2}
	founder.MinigameOfflineQuality["pitch"] = save.MinigameOfflineQualityState{GradePPM: 700_000, LastFounderAttendedMS: 12_000, DecayRemainderPPM: 3}
	founder.Pets = map[string]pet.CareState{}
	founder.FiscalCredit, founder.FiscalPeriodOpenedWallMS, founder.FiscalPeriodSequence = 2, now.UnixMilli(), 1
	founder.FiscalGeneratorLevels = make(map[string]int64, len(catalogs.Fiscal.GeneratorLevelRows()))
	for _, row := range catalogs.Fiscal.GeneratorLevelRows() {
		founder.FiscalGeneratorLevels[row.GeneratorID] = 0
	}
	founder.FiscalUnlocks = map[string]bool{"minigame.pitch": true}
	founder.Soul, founder.SoulExhaustedSourceIDs, founder.MinigameSessionSeq = 80, []string{}, 9
	founder.ExitHistory = []save.ExitRecord{{RunID: 0, ExitType: "collapse", OccurredAt: now.Add(-time.Hour)}}
	if err := catalogs.ValidateFoundationState(founder); err != nil {
		t.Fatalf("first-content Founder: %v", err)
	}
	preState := mustEncodeState(t, company)
	request, err := ParseIntent([]byte(`{"intent_id":"01986666-0c00-7000-8000-000000000001","kind":"wind_down","expected_revision":1,"expected_founder_revision":2}`))
	if err != nil {
		t.Fatal(err)
	}
	carry := founderCarry(founder)
	carry.FounderRevision, carry.FounderConstantsHash = 2, catalogs.ConstantsHash
	minigameActive := false
	activeEvidence, err := resolveActivePlaySchedule(company, catalogs.Opportunities, catalogs.Prestige,
		"01986666-2c00-7000-8000-000000000001", now)
	if err != nil {
		t.Fatal(err)
	}
	nextSpawn, err := catalogs.Opportunities.Spawn("01986666-2c00-7000-8000-000000000001", company.RunSeq+1, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	command := save.ReplayCommand{IntentID: request.IntentID, CompanyStreamID: "01986666-1c00-7000-8000-000000000001",
		FounderID: "01986666-2c00-7000-8000-000000000001", Revision: 1, RunSeq: 1, RunLogSeq: 1}
	catchup, err := buildOfflineCatchup(company, catalogs.Economy, ModeOnline, now)
	if err != nil || catchup == nil {
		t.Fatalf("first-content Exit catchup=%+v err=%v", catchup, err)
	}
	inputs, err := buildReplayInputs(replayBuild{Command: command, Mode: ModeOnline, Now: now, IntentKind: request.Kind,
		RouteContextVersion: catalogs.Routes.ContextVersion(), FounderCarry: &carry, Terminal: true, ExecutedRouteIDs: []string{},
		SelectedExitType: "collapse", SelectedTerms: json.RawMessage(`{}`), NextConstantsHash: catalogs.ConstantsHash,
		ActivePlay: &activeEvidence, NextActivePlay: spawnEvidence(nextSpawn), MinigameSessionActive: &minigameActive, OfflineCatchup: catchup})
	if err != nil {
		t.Fatal(err)
	}
	result := executeTerminalFixture(t, "first-content-same-epoch-exit", catalogs, company, preState, request, inputs, carry)
	return crossRuntimeActiveExit{ConstantsHash: catalogs.ConstantsHash, Artifacts: stringArtifacts(catalogs.Artifacts),
		NextConstantsHash: catalogs.ConstantsHash, NextArtifacts: stringArtifacts(catalogs.Artifacts), Case: result}
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
	catchup, err := buildOfflineCatchup(state, catalogs.Economy, ModeOnline, now)
	if err != nil {
		t.Fatal(err)
	}
	inputs, err := buildReplayInputs(replayBuild{Command: command, Mode: ModeOnline, Now: now, IntentKind: request.Kind,
		RouteContextVersion: catalogs.Routes.ContextVersion(), FounderCarry: &carry, OfflineCatchup: catchup})
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
	result := executeTerminalFixture(t, "accept-stored-offer", catalogs, company, preState, request, inputs, carry)
	resolvedIndex, endedIndex := -1, -1
	for index, current := range result.CompanyEndedEvents {
		if current.Kind == string(save.EventExitOfferResolved) {
			resolvedIndex = index
			if string(current.Payload) != `{"offer_id":"01986666-0200-7000-8000-000000000200","resolution":"accepted"}` {
				t.Fatalf("accepted resolution payload=%s", current.Payload)
			}
		}
		if current.Kind == string(save.EventRunEnded) {
			endedIndex = index
		}
	}
	if resolvedIndex < 0 || endedIndex < 0 || resolvedIndex >= endedIndex {
		t.Fatalf("accepted terminal event order=%+v", result.CompanyEndedEvents)
	}
	return result
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
	result := founderCarry(state)
	result.FounderRevision, result.FounderConstantsHash = carry.FounderRevision, carry.FounderConstantsHash
	return result
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

package production

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"cloud-clicker/server/decimal"
	"cloud-clicker/server/guild"
	"cloud-clicker/server/save"
)

// Complete canonical JSON strings avoid storing two authorities for the same
// expected output. These are observations of actual ApplyLogged transitions.
type axisSequenceAction struct {
	Name          string          `json:"name"`
	PreStateJSON  string          `json:"pre_state_json"`
	Payload       json.RawMessage `json:"canonical_payload"`
	Inputs        json.RawMessage `json:"replay_inputs"`
	Outcome       string          `json:"outcome"`
	ReceiptJSON   string          `json:"receipt_json"`
	EventsJSON    string          `json:"events_json"`
	PostStateJSON string          `json:"post_state_json"`
}
type axisSequenceRow struct {
	ID           string               `json:"id"`
	Effect       string               `json:"effect"`
	FounderID    string               `json:"founder_id"`
	Gap          int64                `json:"gap_ms"`
	Modes        []EvaluationMode     `json:"modes"`
	FirstSpawnMS int64                `json:"first_spawn_attended_ms"`
	PreludeCount int                  `json:"prelude_count"`
	Actions      []axisSequenceAction `json:"actions"`
}
type axisSequenceReport struct {
	Version          int                    `json:"version"`
	Acceptance       string                 `json:"acceptance_status"`
	Sources          map[string]string      `json:"source_sha256"`
	Bundle           reputationCorpusBundle `json:"bundle"`
	Enumerated       int                    `json:"seed_candidates_enumerated"`
	ProvisionTickMS  int64                  `json:"provision_tick_ms"`
	CatchupCeilingMS int64                  `json:"catchup_ceiling_ms"`
	Selected         map[string]string      `json:"selected_first_founders"`
	Rows             []axisSequenceRow      `json:"rows"`
	Negatives        int                    `json:"negative_cases"`
	CommandAttempts  int                    `json:"command_attempts"`
}

var axisSequenceEffects = []string{"active.building", "active.click", "active.lucky", "active.production"}

func axisSequenceFounders(t *testing.T, bundle CatalogBundle) map[string]string {
	t.Helper()
	selected := map[string]string{}
	for index := 1; index <= 4096; index++ {
		id := fmt.Sprintf("01986666-2000-7000-8000-%012x", index)
		spawn, err := bundle.Opportunities.Spawn(id, 2, 0, 0)
		if err != nil {
			t.Fatal(err)
		}
		if selected[spawn.EffectRowID] == "" {
			selected[spawn.EffectRowID] = id
		}
	}
	if len(selected) != 4 {
		t.Fatal("all4096 seed population lacks four effects")
	}
	for _, effect := range axisSequenceEffects {
		if selected[effect] == "" {
			t.Fatalf("missing declared effect %s", effect)
		}
	}
	return selected
}

func axisSequenceStep(t *testing.T, bundle CatalogBundle, row *axisSequenceRow, state *save.State, carry replayFounderCarry, kind, fields string, mode EvaluationMode, now time.Time, want save.IntentOutcome) *save.State {
	t.Helper()
	pre := mustEncodeState(t, state)
	state = activationRestore(t, bundle, pre, 19, now)
	index := len(row.Actions) + 1
	revision := int64(index)
	intentID := fmt.Sprintf("01986666-1f%02x-7000-8000-%012x", index, index)
	request, err := ParseIntent([]byte(fmt.Sprintf(`{"intent_id":%q,"kind":%q,"expected_revision":%d%s}`, intentID, kind, revision, fields)))
	if err != nil || request.InvalidDetail != "" {
		t.Fatalf("sequence intent: %v %s", err, request.InvalidDetail)
	}
	command := save.ReplayCommand{IntentID: request.IntentID, CompanyStreamID: "01986666-1000-7000-8000-000000000001", FounderID: row.FounderID, Revision: revision, RunSeq: 2, RunLogSeq: int64(index)}
	active, err := resolveActivePlaySchedule(state, bundle.Opportunities, bundle.Prestige, row.FounderID, now)
	if err != nil {
		t.Fatal(err)
	}
	catchup, err := buildOfflineCatchup(state, bundle.Economy, mode, now)
	if err != nil {
		t.Fatal(err)
	}
	if kind == IntentClaimOpportunity {
		accrual, err := makeReplayAccrual(nil, nil, guild.SettlementBatch{}, bundle.Routes.ContextVersion())
		if err != nil {
			t.Fatal(err)
		}
		active.Claim, err = resolveActiveClaimEvidence(state, bundle, save.Revision{StreamID: command.CompanyStreamID, OwnerID: row.FounderID, Number: revision, ConstantsHash: bundle.ConstantsHash, RunLogSequence: int64(index)}, request, mode, now, accrual, active)
		if err != nil {
			t.Fatal(err)
		}
		if active.Claim == nil {
			t.Fatal("actual pending claim did not resolve")
		}
	}
	inputs, err := buildReplayInputs(replayBuild{Command: command, Mode: mode, Now: now, IntentKind: kind, RouteContextVersion: bundle.Routes.ContextVersion(), FounderCarry: &carry, ActivePlay: &active, OfflineCatchup: catchup})
	if err != nil {
		t.Fatal(err)
	}
	transition, err := ApplyLogged(state, request.CanonicalPayload, bundle, inputs)
	if err != nil || transition.Outcome != want {
		t.Fatalf("%s/step%d: %v outcome%s receipt%s", row.ID, index, err, transition.Outcome, transition.Receipt)
	}
	post := mustEncodeState(t, transition.State)
	activationRestore(t, bundle, post, 19, now)
	if want == save.IntentRejected && !bytes.Equal(pre, post) {
		t.Fatal("rejected rebuy changed full initial state")
	}
	row.Actions = append(row.Actions, axisSequenceAction{Name: fmt.Sprintf("%s/step%d/%s", row.ID, index, kind), PreStateJSON: canonicalFixtureJSON(t, pre), Payload: request.CanonicalPayload, Inputs: inputs, Outcome: string(transition.Outcome), ReceiptJSON: canonicalFixtureJSON(t, transition.Receipt), EventsJSON: canonicalFixtureValue(t, fixtureEvents(transition.Events)), PostStateJSON: canonicalFixtureJSON(t, post)})
	return transition.State
}

func axisSequenceCase(t *testing.T, bundle CatalogBundle, effect, founderID string, gap int64, modes []EvaluationMode, start time.Time) axisSequenceRow {
	t.Helper()
	row := axisSequenceRow{ID: fmt.Sprintf("%s/%d/%s-%s-%s", effect, gap, modes[0], modes[1], modes[2]), Effect: effect, FounderID: founderID, Gap: gap, Modes: modes}
	state := axisTimingState(t, bundle, start)
	state.GeneratorCounts["generator.beige_tower"], state.GeneratorCounts["generator.beige_tower_v2"], state.GeneratorCounts["generator.legal_dept"] = 99, 10, 1
	state.GeneratorPurchasedTotal, state.ComputeCreditMS = 110, 5000
	setCash(t, state, "1e9")
	spawn, err := initializeActivePlayState(state, bundle.Opportunities, founderID)
	if err != nil || spawn.EffectRowID != effect {
		t.Fatalf("wrong actual seed effect: %v", err)
	}
	carry := tier2FounderCarry(t, bundle, start)
	carry.AchievementsEarnedLifetime, carry.AchievementScoreLifetime = []string{}, 0
	for _, definition := range bundle.Achievements.Definitions {
		carry.AchievementsEarnedLifetime = append(carry.AchievementsEarnedLifetime, definition.ID)
		carry.AchievementScoreLifetime += definition.ScoreGrant
	}
	first := start.Add(time.Duration(spawn.SpawnedAttendedMS) * time.Millisecond)
	row.FirstSpawnMS = spawn.SpawnedAttendedMS
	manual := `,"action_id":"manual.click","count":1,"window_ms":1000`
	for cursor := start; first.Sub(cursor).Milliseconds() > bundle.Prestige.CatchupCeilingMS; {
		cursor = cursor.Add(time.Duration(bundle.Prestige.CatchupCeilingMS) * time.Millisecond)
		state = axisSequenceStep(t, bundle, &row, state, carry, IntentPerformManualBatch, manual, ModeOnline, cursor, save.IntentApplied)
		row.PreludeCount++
	}
	buy := `,"generator_id":"generator.beige_tower","count":{"mode":"exact","value":1}`
	state = axisSequenceStep(t, bundle, &row, state, carry, IntentBuyGenerator, buy, ModeOnline, first, save.IntentApplied)
	if state.AttainmentScoreRun != 12 || state.GeneratorCounts["generator.beige_tower"] != 100 || state.PendingOpportunity == nil || state.PendingOpportunity.EffectRowID != effect || state.AchievementScoreRun != 0 {
		t.Fatalf("%s initial purchase: attainment=%d purchased=%d pending=%+v earned=%d spawn=%+v first=%s", row.ID, state.AttainmentScoreRun, state.GeneratorCounts["generator.beige_tower"], state.PendingOpportunity, state.AchievementScoreRun, spawn, first)
	}
	state = axisSequenceStep(t, bundle, &row, state, carry, IntentClaimOpportunity, fmt.Sprintf(`,"opportunity_id":%q`, state.PendingOpportunity.OpportunityID), ModeOnline, first, save.IntentApplied)
	if state.PendingOpportunity != nil {
		t.Fatal("claim failed to clear actual pending")
	}
	originalBuff := ""
	if effect != "active.lucky" {
		if len(state.ActiveBuffs) != 1 || state.ActiveBuffs[0].EffectRowID != effect {
			t.Fatal("claim did not create actual expected buff")
		}
		originalBuff = state.ActiveBuffs[0].BuffInstanceID
		values, err := activePlayContributions(state, bundle.Opportunities, spawn.SpawnedAttendedMS)
		if err != nil || len(values) != 1 || !values[0].Factor.Gt(decimal.One) {
			t.Fatalf("actual buff missing non-neutral contribution: %v %+v", err, values)
		}
	} else if len(state.ActiveBuffs) != 0 {
		t.Fatal("Lucky was forged into buff")
	}
	plusOne := first.Add(time.Millisecond)
	state = axisSequenceStep(t, bundle, &row, state, carry, IntentBuyUpgrade, `,"upgrade_id":"upgrade.pr_intern_2"`, ModeOnline, plusOne, save.IntentApplied)
	if !state.UpgradesOwned["upgrade.pr_intern_2"] {
		t.Fatal("actual second PR purchase did not own row")
	}
	state = axisSequenceStep(t, bundle, &row, state, carry, IntentSpendComputeCredit, `,"amount_ms":3000,"target":"accelerate"`, ModeOnline, plusOne, save.IntentApplied)
	if state.ComputeCreditMS != 2000 || state.ComputeBurstRemainingMS != 3000 {
		t.Fatal("actual burst did not start with exact debit")
	}
	state = axisSequenceStep(t, bundle, &row, state, carry, IntentBuyGenerator, buy, modes[0], first.Add(500*time.Millisecond), save.IntentApplied)
	returned := first.Add(time.Duration(500+gap) * time.Millisecond)
	state = axisSequenceStep(t, bundle, &row, state, carry, IntentPerformManualBatch, manual, modes[1], returned, save.IntentApplied)
	final := returned.Add(time.Duration(bundle.Prestige.CatchupCeilingMS) * time.Millisecond)
	state = axisSequenceStep(t, bundle, &row, state, carry, IntentPerformManualBatch, manual, modes[2], final, save.IntentApplied)
	for _, buff := range state.ActiveBuffs {
		if buff.BuffInstanceID == originalBuff {
			t.Fatalf("%s original buff remains: %+v offline=%+v final=%s", row.ID, buff, state.OfflineSpans, final)
		}
	}
	permits, ok := state.Ledger.Balance("company.permits")
	provision := state.GeneratorProvisioned["generator.beige_tower"]
	if !ok || !permits.Gt(decimal.Zero) || (gap == 3114 && provision != 0) || (gap == 90000000 && provision < 1) || state.GeneratorCounts["generator.beige_tower"] != 101 {
		t.Fatalf("%s declared population: permits=%s present=%v provision=%d purchased=%d", row.ID, permits.String(), ok, state.GeneratorProvisioned["generator.beige_tower"], state.GeneratorCounts["generator.beige_tower"])
	}
	state = axisSequenceStep(t, bundle, &row, state, carry, IntentBuyUpgrade, `,"upgrade_id":"upgrade.pr_intern_2"`, ModeOnline, final, save.IntentRejected)
	var receipt struct {
		Rejection struct {
			Category string `json:"category"`
			Detail   string `json:"detail"`
		} `json:"rejection"`
	}
	if err := json.Unmarshal([]byte(row.Actions[row.PreludeCount+7].ReceiptJSON), &receipt); err != nil || receipt.Rejection.Category != "not_eligible" || receipt.Rejection.Detail != "owned" {
		t.Fatal("rebuy is not the owned refusal")
	}
	return row
}

func TestAxisActionBuffModeReplayResearch(t *testing.T) {
	bundle := axisContentBundle(t)
	report := axisSequenceReport{Version: 1, Acceptance: "NOT_PROVEN: bounded action/buff/mode sequences; production AC6 remains red", Enumerated: 4096, Selected: axisSequenceFounders(t, bundle), Bundle: reputationCorpusBundle{ConstantsHash: bundle.ConstantsHash, Artifacts: stringArtifacts(bundle.Artifacts)}, Sources: sourceResearchHashes(t, []string{"server/production/axis_sequence_research_test.go", "client/test/sequence-research.test.ts", "server/production/axis_activation_research_test.go", "server/production/axis_timing_test.go", "server/production/axis_stack_test.go", "server/production/replay_cross_runtime_test.go", "server/production/tier2_candidate_test.go", "server/production/replay.go", "server/production/active_play.go", "server/production/intents.go", "server/production/engine.go", "server/save/state.go", "client/src/replay.ts", "client/src/economy.ts", "balance/testdata/axis-stack/economy-v5-fixture.json", "balance/opportunities/t0-t1.json"})}
	report.ProvisionTickMS = bundle.Economy.ProvisionTickMS()
	report.CatchupCeilingMS = bundle.Prestige.CatchupCeilingMS
	if report.CatchupCeilingMS != 5000 {
		t.Fatal("sequence requires pinned5000ms attended-clock boundary")
	}
	for _, path := range []string{"server/prestige/runtime.go", "balance/prestige/phase0.json"} {
		report.Sources[path] = sourceResearchHashes(t, []string{path})[path]
	}
	start := time.Date(2026, 10, 7, 12, 0, 0, 0, time.UTC)
	for _, effect := range axisSequenceEffects {
		for _, gap := range []int64{3114, 90000000} {
			for _, modes := range [][]EvaluationMode{{ModeOnline, ModeOffline, ModeOnline}, {ModeOffline, ModeOnline, ModeOffline}} {
				report.Rows = append(report.Rows, axisSequenceCase(t, bundle, effect, report.Selected[effect], gap, modes, start))
			}
		}
	}
	for _, row := range report.Rows {
		claim := row.Actions[row.PreludeCount+1]
		for _, fault := range []string{"missing", "effect", "next"} {
			var wire replayInputsWire
			var resolved replayAccrualResolved
			if err := json.Unmarshal(claim.Inputs, &wire); err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal(wire.Resolved, &resolved); err != nil {
				t.Fatal(err)
			}
			if resolved.ActivePlay == nil || resolved.ActivePlay.Claim == nil {
				t.Fatal("claim negative lacks actual evidence")
			}
			switch fault {
			case "missing":
				resolved.ActivePlay.Claim = nil
			case "effect":
				resolved.ActivePlay.Claim.EffectRowID = "active.invalid"
			case "next":
				resolved.ActivePlay.Claim.NextOpportunityAttendedMS++
			}
			var err error
			wire.Resolved, err = json.Marshal(resolved)
			if err != nil {
				t.Fatal(err)
			}
			bad, err := json.Marshal(wire)
			if err != nil {
				t.Fatal(err)
			}
			state, err := save.RestoreState([]byte(claim.PreStateJSON), 19, bundle.Economy, "company", start)
			if err != nil {
				t.Fatal(err)
			}
			before := mustEncodeState(t, state)
			_, err = ApplyLogged(state, claim.Payload, bundle, bad)
			if !errors.Is(err, ErrInvalidReplayInputs) || !bytes.Equal(before, mustEncodeState(t, state)) {
				t.Fatalf("%s/%s refusal/rollback failed: %v", row.ID, fault, err)
			}
			report.Negatives++
		}
	}
	if len(report.Rows) != 16 || report.Negatives != 48 {
		t.Fatal("sequence observation truncated")
	}
	for _, row := range report.Rows {
		if row.PreludeCount != int((row.FirstSpawnMS-1)/report.CatchupCeilingMS) || len(row.Actions) != 8+row.PreludeCount {
			t.Fatal("sequence actions truncated")
		}
		report.CommandAttempts += len(row.Actions)
	}
	sourceResearchArtifact(t, "../../testdata/axis-stack/sequence-research-v1.json", "UPDATE_SEQUENCE_RESEARCH", report)
	t.Logf("16 actual sequences/eight core commands plus clock preludes/%d command attempts/48 unchanged-state claim refusals; not SQL/natural progression/AC6", report.CommandAttempts)
}

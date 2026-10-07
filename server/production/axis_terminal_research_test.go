package production

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"cloud-clicker/server/save"
)

// Actual terminal and next-run producers over admitted synthetic inventory;
// not SQL transactions, natural progression or partition acceptance.
type terminalResearchRow struct {
	Profile loggedPolicyProfile      `json:"profile"`
	Exit    crossRuntimeTerminalCase `json:"exit"`
	Next    crossRuntimeFixtureCase  `json:"next"`
}

type terminalResearchReport struct {
	Version    int                    `json:"version"`
	Acceptance string                 `json:"acceptance_status"`
	Sources    map[string]string      `json:"source_sha256"`
	Bundle     reputationCorpusBundle `json:"bundle"`
	Rows       []terminalResearchRow  `json:"rows"`
	Negatives  int                    `json:"negative_cases"`
}

func terminalResearchProfiles() []loggedPolicyProfile {
	rows := []loggedPolicyProfile{}
	for _, elapsed := range []int64{0, 3114, 120001} {
		for _, profile := range []loggedPolicyProfile{{ID: "quiet"}, {ID: "combined", Burst: 1500, Provider: 10, Legal: 1}} {
			profile.ID, profile.Mode, profile.Elapsed = fmt.Sprintf("%s/online/%d", profile.ID, elapsed), ModeOnline, elapsed
			rows = append(rows, profile)
		}
	}
	return rows
}

func terminalResearchCase(t *testing.T, bundle CatalogBundle, p loggedPolicyProfile, start time.Time) terminalResearchRow {
	t.Helper()
	company := axisTimingState(t, bundle, start)
	company.ComputeBurstRemainingMS = p.Burst
	company.GeneratorCounts["generator.beige_tower_v2"], company.GeneratorCounts["generator.legal_dept"] = p.Provider, p.Legal
	company.GeneratorPurchasedTotal = 1 + p.Provider + p.Legal
	pre := mustEncodeState(t, company)
	company = loggedPolicyRestore(t, bundle, pre, start)
	carry := tier2FounderCarry(t, bundle, start)
	carry.AchievementsEarnedLifetime, carry.AchievementScoreLifetime = []string{}, 0
	for _, definition := range bundle.Achievements.Definitions {
		carry.AchievementsEarnedLifetime = append(carry.AchievementsEarnedLifetime, definition.ID)
		carry.AchievementScoreLifetime += definition.ScoreGrant
	}
	request, err := ParseIntent([]byte(`{"intent_id":"01986666-0e01-7000-8000-000000000001","kind":"wind_down","expected_revision":1,"expected_founder_revision":2}`))
	if err != nil || request.InvalidDetail != "" {
		t.Fatalf("terminal research request: %v %s", err, request.InvalidDetail)
	}
	command := save.ReplayCommand{IntentID: request.IntentID, CompanyStreamID: "01986666-1000-7000-8000-000000000001",
		FounderID: "01986666-2000-7000-8000-000000000001", Revision: 1, RunSeq: 2, RunLogSeq: 1}
	now := start.Add(time.Duration(p.Elapsed) * time.Millisecond)
	active, err := resolveActivePlaySchedule(company, bundle.Opportunities, bundle.Prestige, command.FounderID, now)
	if err != nil {
		t.Fatal(err)
	}
	spawn, err := bundle.Opportunities.Spawn(command.FounderID, 3, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	minigame := false
	inputs, err := buildReplayInputs(replayBuild{Command: command, Mode: ModeOnline, Now: now, IntentKind: request.Kind,
		RouteContextVersion: bundle.Routes.ContextVersion(), FounderCarry: &carry, Terminal: true, ExecutedRouteIDs: []string{},
		SelectedExitType: "collapse", SelectedTerms: json.RawMessage(`{}`), NextConstantsHash: bundle.ConstantsHash,
		ActivePlay: &active, NextActivePlay: spawnEvidence(spawn), MinigameSessionActive: &minigame})
	if err != nil {
		t.Fatal(err)
	}
	exit := executeTerminalFixture(t, p.ID, bundle, company, pre, request, inputs, carry)
	if exit.Outcome != string(save.IntentApplied) {
		t.Fatalf("terminal research outcome: %s %s", exit.Outcome, exit.Receipt)
	}
	final := loggedPolicyRestore(t, bundle, exit.FinalCompany, start)
	if final.AttainmentScoreRun < 6 || len(final.AchievementsAttainedRun) < 2 {
		t.Fatal("terminal discarded run attainment history")
	}
	next := loggedPolicyRestore(t, bundle, exit.NewCompany, now)
	if next.WireVersion != 19 || next.RunSeq != 3 || len(next.AchievementsAttainedRun) != 0 || next.AttainmentScoreRun != 0 ||
		next.PendingOpportunity != nil || len(next.ActiveBuffs) != 0 || next.OpportunitySpawnSeq != 0 || next.NextOpportunityAttendedMS != spawn.SpawnedAttendedMS {
		t.Fatal("terminal next-run reset/floor/scheduler mismatch")
	}
	var nextCarry replayFounderCarry
	if err := json.Unmarshal([]byte(exit.FounderOutputJSON), &nextCarry); err != nil {
		t.Fatal(err)
	}
	nextCarry.FounderRevision++
	manual, err := ParseIntent([]byte(`{"intent_id":"01986666-0e02-7000-8000-000000000002","kind":"perform_manual_batch","expected_revision":3,"action_id":"manual.click","count":1,"window_ms":1000}`))
	if err != nil || manual.InvalidDetail != "" {
		t.Fatalf("next-run manual request: %v %s", err, manual.InvalidDetail)
	}
	command.IntentID, command.Revision, command.RunSeq, command.RunLogSeq = manual.IntentID, 3, 3, 1
	manualNow := now.Add(time.Second)
	active, err = resolveActivePlaySchedule(next, bundle.Opportunities, bundle.Prestige, command.FounderID, manualNow)
	if err != nil {
		t.Fatal(err)
	}
	manualInputs, err := buildReplayInputs(replayBuild{Command: command, Mode: ModeOnline, Now: manualNow, IntentKind: manual.Kind,
		RouteContextVersion: bundle.Routes.ContextVersion(), FounderCarry: &nextCarry, ActivePlay: &active})
	if err != nil {
		t.Fatal(err)
	}
	transition, err := ApplyLogged(next, manual.CanonicalPayload, bundle, manualInputs)
	if err != nil || transition.Outcome != save.IntentApplied {
		t.Fatalf("next-run manual: %v %s %s", err, transition.Outcome, transition.Receipt)
	}
	post := mustEncodeState(t, transition.State)
	loggedPolicyRestore(t, bundle, post, manualNow)
	nextCase := crossRuntimeFixtureCase{Name: p.ID + "/next-manual", PreState: exit.NewCompany, CanonicalPayload: manual.CanonicalPayload,
		ReplayInputs: manualInputs, Outcome: string(transition.Outcome), Receipt: transition.Receipt, Events: fixtureEvents(transition.Events), PostState: post,
		ReceiptJSON: canonicalFixtureJSON(t, transition.Receipt), EventsJSON: canonicalFixtureValue(t, fixtureEvents(transition.Events)), PostStateJSON: canonicalFixtureJSON(t, post)}
	return terminalResearchRow{Profile: p, Exit: exit, Next: nextCase}
}

func TestAxisTerminalReplayResearch(t *testing.T) {
	bundle := researchBundleCap(t, axisContentBundle(t), "1e100")
	profiles := terminalResearchProfiles()
	if len(profiles) != 6 {
		t.Fatal("incomplete terminal research population")
	}
	report := terminalResearchReport{Version: 1, Acceptance: "NOT_PROVEN: bounded terminal research; production AC6 remains red",
		Bundle: reputationCorpusBundle{ConstantsHash: bundle.ConstantsHash, Artifacts: stringArtifacts(bundle.Artifacts)},
		Sources: sourceResearchHashes(t, []string{"server/production/axis_terminal_research_test.go", "client/test/terminal-research.test.ts",
			"server/production/axis_logged_policy_research_test.go", "server/production/axis_timing_test.go", "server/production/axis_stack_test.go",
			"server/production/tier2_candidate_test.go", "server/production/replay_cross_runtime_test.go", "server/production/replay.go",
			"server/production/prestige.go", "server/production/foundations.go", "server/production/active_play.go", "server/save/state.go",
			"balance/testdata/axis-stack/economy-v5-fixture.json", "client/src/replay.ts", "client/src/economy.ts"})}
	start := time.Date(2026, 10, 7, 12, 0, 0, 0, time.UTC)
	seen := map[string]bool{}
	for _, p := range profiles {
		if seen[p.ID] {
			t.Fatal("duplicate terminal research profile")
		}
		seen[p.ID] = true
		report.Rows = append(report.Rows, terminalResearchCase(t, bundle, p, start))
	}
	row := report.Rows[0]
	for _, fault := range []string{"current", "next", "sequence"} {
		var inputs replayInputsWire
		var resolved replayExitResolved
		if err := json.Unmarshal(row.Exit.ReplayInputs, &inputs); err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(inputs.Resolved, &resolved); err != nil {
			t.Fatal(err)
		}
		if resolved.ActivePlay == nil || resolved.NextActivePlay == nil {
			t.Fatal("terminal negative population missing schedules")
		}
		switch fault {
		case "current":
			resolved.ActivePlay = nil
		case "next":
			resolved.NextActivePlay = nil
		case "sequence":
			resolved.NextActivePlay.Sequence++
		}
		var err error
		inputs.Resolved, err = json.Marshal(resolved)
		if err != nil {
			t.Fatal(err)
		}
		bad, err := json.Marshal(inputs)
		if err != nil {
			t.Fatal(err)
		}
		state := loggedPolicyRestore(t, bundle, row.Exit.PreState, start)
		_, err = ApplyLoggedExit(state, row.Exit.CanonicalPayload, bundle, bad)
		if !errors.Is(err, ErrInvalidReplayInputs) {
			t.Fatalf("terminal %s not invalid-input refusal: %v", fault, err)
		}
		if !bytes.Equal(row.Exit.PreState, mustEncodeState(t, state)) {
			t.Fatal("terminal refusal changed complete initial state")
		}
		report.Negatives++
	}
	if len(report.Rows) != 6 || report.Negatives != 3 {
		t.Fatal("truncated terminal observation")
	}
	sourceResearchRegressionArtifact(t, "../../testdata/axis-stack/terminal-research-v1.json", "UPDATE_TERMINAL_RESEARCH", &report, &report.Sources)
	t.Log("CV4 actual Go:6 terminal/next-manual pairs,3 unchanged-state refusals; TS/SQL/AC6 acceptance not inferred")
}

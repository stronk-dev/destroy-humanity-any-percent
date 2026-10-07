package production

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"cloud-clicker/server/save"
)

const activationFounderID = "01986666-2000-7000-8000-000000000001"

type activationResearchRow struct {
	ID   string                    `json:"id"`
	Old  crossRuntimeFixtureCase   `json:"old"`
	Exit crossRuntimeTerminalCase  `json:"exit"`
	Next []crossRuntimeFixtureCase `json:"next"`
}

type activationResearchReport struct {
	Version    int                     `json:"version"`
	Acceptance string                  `json:"acceptance_status"`
	Sources    map[string]string       `json:"source_sha256"`
	Current    reputationCorpusBundle  `json:"current"`
	Next       reputationCorpusBundle  `json:"next"`
	Rows       []activationResearchRow `json:"rows"`
}

func activationRestore(t *testing.T, bundle CatalogBundle, raw json.RawMessage, version int, now time.Time) *save.State {
	t.Helper()
	state, err := save.RestoreState(raw, version, bundle.Economy, "company", now)
	if err != nil {
		t.Fatal(err)
	}
	if err := bundle.ValidateFoundationState(state); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(raw, mustEncodeState(t, state)) {
		t.Fatal("activation restore changed complete state")
	}
	return state
}

func activationInitial(t *testing.T, bundle CatalogBundle, now time.Time, veteran bool) (*save.State, *save.State) {
	t.Helper()
	company := tier2Company(t, bundle, now)
	company.GeneratorCounts["generator.beige_tower"], company.GeneratorPurchasedTotal = 1, 1
	if _, err := initializeActivePlayState(company, bundle.Opportunities, activationFounderID); err != nil {
		t.Fatal(err)
	}
	founder := reputationFounderState(t, bundle, 21, now, 0)
	founder.ExitHistory = []save.ExitRecord{{RunID: 1, ExitType: "collapse", OccurredAt: now.Add(-time.Hour)}}
	if veteran {
		for _, definition := range bundle.Achievements.Definitions {
			founder.AchievementsEarnedLifetime[definition.ID] = true
			founder.AchievementScoreLifetime += definition.ScoreGrant
		}
	}
	if err := bundle.ValidateFoundationState(company); err != nil {
		t.Fatal(err)
	}
	if err := bundle.ValidateFoundationState(founder); err != nil {
		t.Fatal(err)
	}
	return company, founder
}

func activationOrdinary(t *testing.T, name string, bundle CatalogBundle, state *save.State, carry replayFounderCarry, command save.ReplayCommand, body string, now time.Time) (crossRuntimeFixtureCase, *save.State) {
	t.Helper()
	pre := mustEncodeState(t, state)
	state = activationRestore(t, bundle, pre, save.VersionForState(state), now)
	request, err := ParseIntent([]byte(body))
	if err != nil || request.InvalidDetail != "" {
		t.Fatalf("activation intent: %v %s", err, request.InvalidDetail)
	}
	command.IntentID = request.IntentID
	active, err := resolveActivePlaySchedule(state, bundle.Opportunities, bundle.Prestige, command.FounderID, now)
	if err != nil {
		t.Fatal(err)
	}
	catchup, err := buildOfflineCatchup(state, bundle.Economy, ModeOnline, now)
	if err != nil {
		t.Fatal(err)
	}
	inputs, err := buildReplayInputs(replayBuild{Command: command, Mode: ModeOnline, Now: now, IntentKind: request.Kind,
		RouteContextVersion: bundle.Routes.ContextVersion(), FounderCarry: &carry, ActivePlay: &active, OfflineCatchup: catchup})
	if err != nil {
		t.Fatal(err)
	}
	transition, err := ApplyLogged(state, request.CanonicalPayload, bundle, inputs)
	if err != nil || transition.Outcome != save.IntentApplied {
		t.Fatalf("%s: %v %s %s", name, err, transition.Outcome, transition.Receipt)
	}
	post := mustEncodeState(t, transition.State)
	activationRestore(t, bundle, post, save.VersionForState(transition.State), now)
	return crossRuntimeFixtureCase{Name: name, PreState: pre, CanonicalPayload: request.CanonicalPayload, ReplayInputs: inputs,
		Outcome: string(transition.Outcome), Receipt: transition.Receipt, Events: fixtureEvents(transition.Events), PostState: post,
		ReceiptJSON: canonicalFixtureJSON(t, transition.Receipt), EventsJSON: canonicalFixtureValue(t, fixtureEvents(transition.Events)), PostStateJSON: canonicalFixtureJSON(t, post)}, transition.State
}

func activationResearchCase(t *testing.T, current, next CatalogBundle, veteran bool, gap int64, start time.Time) activationResearchRow {
	t.Helper()
	profile := "fresh"
	if veteran {
		profile = "veteran"
	}
	row := activationResearchRow{ID: fmt.Sprintf("%s/online/%d", profile, gap)}
	company, founder := activationInitial(t, current, start, veteran)
	carry := founderCarry(founder)
	carry.FounderRevision, carry.FounderConstantsHash = 1, current.ConstantsHash
	command := save.ReplayCommand{CompanyStreamID: "01986666-1000-7000-8000-000000000001", FounderID: activationFounderID, Revision: 1, RunSeq: 2, RunLogSeq: 1}
	manualTime := start.Add(time.Second)
	row.Old, company = activationOrdinary(t, row.ID+"/old-manual", current, company, carry, command,
		`{"intent_id":"01986666-0f01-7000-8000-000000000001","kind":"perform_manual_batch","expected_revision":1,"action_id":"manual.click","count":1,"window_ms":1000}`, manualTime)
	if company.WireVersion != 18 || company.AchievementsAttainedRun != nil || company.AttainmentScoreRun != 0 {
		t.Fatal("old manual acquired new axis")
	}
	exitTime := manualTime.Add(time.Duration(gap) * time.Millisecond)
	request, err := ParseIntent([]byte(`{"intent_id":"01986666-0f02-7000-8000-000000000002","kind":"wind_down","expected_revision":2,"expected_founder_revision":1}`))
	if err != nil {
		t.Fatal(err)
	}
	command.IntentID, command.Revision, command.RunLogSeq = request.IntentID, 2, 2
	active, err := resolveActivePlaySchedule(company, current.Opportunities, current.Prestige, command.FounderID, exitTime)
	if err != nil {
		t.Fatal(err)
	}
	catchup, err := buildOfflineCatchup(company, current.Economy, ModeOnline, exitTime)
	if err != nil {
		t.Fatal(err)
	}
	spawn, err := next.Opportunities.Spawn(command.FounderID, 3, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	inactive := false
	inputs, err := buildReplayInputs(replayBuild{Command: command, Mode: ModeOnline, Now: exitTime, IntentKind: request.Kind,
		RouteContextVersion: current.Routes.ContextVersion(), FounderCarry: &carry, Terminal: true, ExecutedRouteIDs: []string{},
		SelectedExitType: "collapse", SelectedTerms: json.RawMessage(`{}`), NextConstantsHash: next.ConstantsHash,
		ActivePlay: &active, NextActivePlay: spawnEvidence(spawn), MinigameSessionActive: &inactive, OfflineCatchup: catchup})
	if err != nil {
		t.Fatal(err)
	}
	current.Next = &next
	row.Exit = executeTerminalFixture(t, row.ID, current, company, mustEncodeState(t, company), request, inputs, carry)
	if row.Exit.Outcome != string(save.IntentApplied) {
		t.Fatalf("activation exit rejected: %s", row.Exit.Receipt)
	}
	final := activationRestore(t, current, row.Exit.FinalCompany, 18, exitTime)
	if final.WireVersion != 18 || final.AchievementsAttainedRun != nil || final.AttainmentScoreRun != 0 {
		t.Fatal("activation retrofitted old final state")
	}
	company = activationRestore(t, next, row.Exit.NewCompany, 19, exitTime)
	if company.WireVersion != 19 || company.RunSeq != 3 || len(company.AchievementsAttainedRun) != 0 || company.AttainmentScoreRun != 0 || company.AchievementScoreRun != 0 || len(company.AchievementsEarnedRun) != 0 {
		t.Fatal("activation initial reset/floor mismatch")
	}
	if err := json.Unmarshal([]byte(row.Exit.FounderOutputJSON), &carry); err != nil {
		t.Fatal(err)
	}
	carry.FounderRevision, carry.FounderConstantsHash = 2, next.ConstantsHash
	command.Revision, command.RunSeq, command.RunLogSeq = 4, 3, 1
	action, company := activationOrdinary(t, row.ID+"/next-manual", next, company, carry, command,
		`{"intent_id":"01986666-0f03-7000-8000-000000000003","kind":"perform_manual_batch","expected_revision":4,"action_id":"manual.click","count":10,"window_ms":1000}`, exitTime.Add(time.Second))
	row.Next = append(row.Next, action)
	command.Revision, command.RunLogSeq = 5, 2
	action, company = activationOrdinary(t, row.ID+"/next-purchase", next, company, carry, command,
		`{"intent_id":"01986666-0f04-7000-8000-000000000004","kind":"buy_generator","expected_revision":5,"generator_id":"generator.beige_tower","count":{"mode":"exact","value":1}}`, exitTime.Add(1001*time.Millisecond))
	row.Next = append(row.Next, action)
	if company.GeneratorPurchasedTotal != 1 || company.AttainmentScoreRun != 2 || len(company.AchievementsAttainedRun) != 1 || !company.AchievementsAttainedRun["achievement.generators_purchased_1"] || company.AchievementScoreRun != 0 || len(company.AchievementsEarnedRun) != 0 {
		t.Fatal("next purchase did not independently re-attain without lifetime duplication")
	}
	return row
}

func TestAxisActivationReplayResearch(t *testing.T) {
	current, next := activeContentBundle(t), axisContentBundle(t)
	if current.ConstantsHash == next.ConstantsHash {
		t.Fatal("activation bundle identities identical")
	}
	report := activationResearchReport{Version: 1, Acceptance: "NOT_PROVEN: bounded pinned activation; production AC6 remains red",
		Current: reputationCorpusBundle{ConstantsHash: current.ConstantsHash, Artifacts: stringArtifacts(current.Artifacts)},
		Next:    reputationCorpusBundle{ConstantsHash: next.ConstantsHash, Artifacts: stringArtifacts(next.Artifacts)},
		Sources: sourceResearchHashes(t, []string{"server/production/axis_activation_research_test.go", "client/test/activation-research.test.ts", "server/production/axis_activation_integration_test.go",
			"server/production/prestige.go", "server/production/foundations.go", "server/production/active_play.go", "server/production/replay.go", "server/production/founder_replay.go", "server/production/founder_exit_replay.go",
			"server/production/first_content_epoch_test.go", "server/production/axis_stack_test.go", "server/production/tier2_candidate_test.go", "server/production/reputation_purchase_test.go",
			"server/production/replay_cross_runtime_test.go", "server/production/reputation_career_integration_test.go", "server/production/reputation_plan_atomicity_integration_test.go",
			"server/save/state.go", "server/save/exit.go", "server/save/genesis.go", "server/production/verifier.go", "client/src/replay.ts", "balance/testdata/axis-stack/economy-v5-fixture.json"})}
	start := time.Date(2026, 10, 7, 12, 0, 0, 0, time.UTC)
	for _, veteran := range []bool{false, true} {
		for _, gap := range []int64{0, 3114, 90000000} {
			report.Rows = append(report.Rows, activationResearchCase(t, current, next, veteran, gap, start))
		}
	}
	if len(report.Rows) != 6 {
		t.Fatal("activation population truncated")
	}
	sourceResearchRegressionArtifact(t, "../../testdata/axis-stack/activation-research-v1.json", "UPDATE_ACTIVATION_RESEARCH", &report, &report.Sources)
	t.Log("six complete actual old/manual→activation→new/manual/purchase sequences; not natural progression or AC6")
}

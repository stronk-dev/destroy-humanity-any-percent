package production

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"cloud-clicker/server/decimal"
	"cloud-clicker/server/save"
)

// Synthetic admitted inventory and an actual logged producer, not natural
// progression, SQL persistence, an adopted anchor or partition acceptance.
type loggedPolicyProfile struct {
	ID       string         `json:"id"`
	Mode     EvaluationMode `json:"mode"`
	Elapsed  int64          `json:"elapsed_ms"`
	Burst    int64          `json:"initial_burst_ms"`
	Provider int64          `json:"provider_count"`
	Legal    int64          `json:"legal_count"`
}

type loggedPolicyObservation struct {
	Profile loggedPolicyProfile     `json:"profile"`
	Case    crossRuntimeFixtureCase `json:"case"`
}

type loggedPolicyReport struct {
	Version       int                       `json:"version"`
	Acceptance    string                    `json:"acceptance_status"`
	Sources       map[string]string         `json:"source_sha256"`
	Bundle        reputationCorpusBundle    `json:"bundle"`
	Rows          []loggedPolicyObservation `json:"rows"`
	NegativeCases int                       `json:"negative_cases"`
}

func loggedPolicyProfiles() []loggedPolicyProfile {
	profiles := []loggedPolicyProfile{}
	for _, mode := range []EvaluationMode{ModeOnline, ModeOffline} {
		for _, end := range []int64{3114, 59999, 120001, 90000000} {
			for _, p := range []struct {
				name                   string
				burst, provider, legal int64
			}{
				{"quiet", 0, 0, 0}, {"boosted", 1500, 0, 0}, {"combined", 1500, 10, 1},
			} {
				profiles = append(profiles, loggedPolicyProfile{fmt.Sprintf("%s/%s/%d", p.name, mode, end), mode, end, p.burst, p.provider, p.legal})
			}
		}
	}
	return profiles
}

func loggedPolicyRestore(t *testing.T, bundle CatalogBundle, encoded json.RawMessage, start time.Time) *save.State {
	t.Helper()
	state, err := save.RestoreState(encoded, 19, bundle.Economy, "company", start)
	if err != nil {
		t.Fatalf("logged policy restore: %v", err)
	}
	if err := bundle.ValidateFoundationState(state); err != nil {
		t.Fatalf("logged policy admission: %v", err)
	}
	if !bytes.Equal(encoded, mustEncodeState(t, state)) {
		t.Fatal("logged policy full restore changed bytes")
	}
	return state
}

func loggedPolicyCase(t *testing.T, bundle CatalogBundle, profile loggedPolicyProfile, start time.Time) crossRuntimeFixtureCase {
	t.Helper()
	state := axisTimingState(t, bundle, start)
	state.ComputeBurstRemainingMS = profile.Burst
	state.GeneratorCounts["generator.beige_tower_v2"] = profile.Provider
	state.GeneratorCounts["generator.legal_dept"] = profile.Legal
	state.GeneratorPurchasedTotal = 1 + profile.Provider + profile.Legal
	pre := mustEncodeState(t, state)
	state = loggedPolicyRestore(t, bundle, pre, start)
	request, err := ParseIntent([]byte(`{"intent_id":"01986666-0901-7000-8000-000000000901","kind":"buy_generator","expected_revision":1,"generator_id":"generator.beige_tower","count":{"mode":"exact","value":1}}`))
	if err != nil {
		t.Fatal(err)
	}
	command := save.ReplayCommand{IntentID: request.IntentID, CompanyStreamID: "01986666-1000-7000-8000-000000000001",
		FounderID: "01986666-2000-7000-8000-000000000001", Revision: 1, RunSeq: 2, RunLogSeq: 1}
	carry := tier2FounderCarry(t, bundle, start)
	carry.AchievementsEarnedLifetime, carry.AchievementScoreLifetime = []string{}, 0
	for _, definition := range bundle.Achievements.Definitions {
		carry.AchievementsEarnedLifetime = append(carry.AchievementsEarnedLifetime, definition.ID)
		carry.AchievementScoreLifetime += definition.ScoreGrant
	}
	now := start.Add(time.Duration(profile.Elapsed) * time.Millisecond)
	active, err := resolveActivePlaySchedule(state, bundle.Opportunities, bundle.Prestige, command.FounderID, now)
	if err != nil {
		t.Fatal(err)
	}
	catchup, err := buildOfflineCatchup(state, bundle.Economy, profile.Mode, now)
	if err != nil {
		t.Fatal(err)
	}
	wantCatchup := profile.Mode == ModeOnline && profile.Elapsed == 90000000
	if (catchup != nil) != wantCatchup {
		t.Fatal("incorrect diagnostic catchup presence")
	}
	inputs, err := buildReplayInputs(replayBuild{Command: command, Mode: profile.Mode, Now: now, IntentKind: request.Kind,
		RouteContextVersion: bundle.Routes.ContextVersion(), FounderCarry: &carry, ActivePlay: &active, OfflineCatchup: catchup})
	if err != nil {
		t.Fatal(err)
	}
	transition, err := ApplyLogged(state, request.CanonicalPayload, bundle, inputs)
	if err != nil || transition.Outcome != save.IntentApplied {
		t.Fatalf("%s: outcome%s error%v receipt%s", profile.ID, transition.Outcome, err, transition.Receipt)
	}
	if state.GeneratorCounts["generator.beige_tower"] != 2 || state.GeneratorPurchasedTotal != 2+profile.Provider+profile.Legal ||
		state.GeneratorCounts["generator.beige_tower_v2"] != profile.Provider || state.GeneratorCounts["generator.legal_dept"] != profile.Legal || state.ComputeBurstRemainingMS != 0 {
		t.Fatalf("%s purchase/burst binding failed", profile.ID)
	}
	permits, present := state.Ledger.Balance("company.permits")
	if !present || (profile.Legal > 0 && !permits.Gt(decimal.Zero)) || (profile.Legal == 0 && permits != decimal.Zero) {
		t.Fatal("nonzero/zero permits population binding failed")
	}
	post := mustEncodeState(t, transition.State)
	loggedPolicyRestore(t, bundle, post, start)
	return crossRuntimeFixtureCase{Name: profile.ID, PreState: pre, CanonicalPayload: request.CanonicalPayload, ReplayInputs: inputs,
		Outcome: string(transition.Outcome), Receipt: transition.Receipt, Events: fixtureEvents(transition.Events), PostState: post,
		ReceiptJSON: canonicalFixtureJSON(t, transition.Receipt), EventsJSON: canonicalFixtureValue(t, fixtureEvents(transition.Events)), PostStateJSON: canonicalFixtureJSON(t, post)}
}

func TestAxisLoggedPolicyReplayResearch(t *testing.T) {
	bundle := researchBundleCap(t, axisContentBundle(t), "1e100")
	profiles := loggedPolicyProfiles()
	if len(profiles) != 24 {
		t.Fatal("incomplete logged policy population")
	}
	report := loggedPolicyReport{Version: 1, Acceptance: "NOT_PROVEN: bounded logged replay only; production AC6 remains red",
		Bundle: reputationCorpusBundle{ConstantsHash: bundle.ConstantsHash, Artifacts: stringArtifacts(bundle.Artifacts)},
		Sources: sourceResearchHashes(t, []string{"server/production/axis_logged_policy_research_test.go", "client/test/logged-policy-research.test.ts",
			"server/production/axis_timing_test.go", "server/production/axis_stack_test.go", "server/production/axis_partition_research_test.go",
			"server/production/axis_rate_source_research_test.go", "balance/testdata/axis-stack/economy-v5-fixture.json",
			"server/production/replay.go", "server/production/engine.go", "server/production/content.go", "server/production/accrual.go",
			"server/save/state.go", "client/src/replay.ts", "client/src/economy.ts"})}
	start := time.Date(2026, 10, 7, 12, 0, 0, 0, time.UTC)
	seen := map[string]bool{}
	for _, profile := range profiles {
		if seen[profile.ID] {
			t.Fatal("duplicate logged policy profile")
		}
		seen[profile.ID] = true
		row := loggedPolicyCase(t, bundle, profile, start)
		report.Rows = append(report.Rows, loggedPolicyObservation{profile, row})
		if profile.Mode != ModeOnline || profile.Elapsed != 90000000 {
			continue
		}
		for _, fault := range []string{"missing", "from", "to"} {
			var inputs replayInputsWire
			if err := json.Unmarshal(row.ReplayInputs, &inputs); err != nil {
				t.Fatal(err)
			}
			if inputs.OfflineCatchup == nil {
				t.Fatal("negative population missing catchup")
			}
			switch fault {
			case "missing":
				inputs.OfflineCatchup = nil
			case "from":
				inputs.OfflineCatchup.OfflineSpan.FromMS++
			case "to":
				inputs.OfflineCatchup.OfflineSpan.ToMS--
			}
			bad, err := json.Marshal(inputs)
			if err != nil {
				t.Fatal(err)
			}
			state := loggedPolicyRestore(t, bundle, row.PreState, start)
			_, err = ApplyLogged(state, row.CanonicalPayload, bundle, bad)
			if !errors.Is(err, ErrInvalidReplayInputs) {
				t.Fatalf("%s/%s not explicitly refused: %v", profile.ID, fault, err)
			}
			if !bytes.Equal(row.PreState, mustEncodeState(t, state)) {
				t.Fatal("invalid catchup changed initial state")
			}
			report.NegativeCases++
		}
	}
	if len(report.Rows) != 24 || report.NegativeCases != 9 {
		t.Fatal("truncated logged policy observation")
	}
	sourceResearchRegressionArtifact(t, "../../testdata/axis-stack/logged-policy-research-v1.json", "UPDATE_LOGGED_POLICY_RESEARCH", &report, &report.Sources)
	t.Log("R-012 actual Go ApplyLogged:24 admitted purchases,9 explicit unmodified catchup refusals; TS comparison and AC6 not inferred")
}

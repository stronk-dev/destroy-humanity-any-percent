package production

import (
	"bytes"
	"os"
	"testing"
	"time"

	"cloud-clicker/server/epochseed"
	"cloud-clicker/server/leaderboard"
	"cloud-clicker/server/meters"
	"cloud-clicker/server/pet"
	"cloud-clicker/server/routes"
	"cloud-clicker/server/save"
)

// tier2CandidateArtifacts is the live epoch-8 bundle with the three Tier-2
// candidates (rfc/tier2-content.md §A1, §A2, §B) swapped in. It is fixture-first:
// no epoch pins these bytes until the owner ratifies them by SHA.
func tier2CandidateArtifacts(t *testing.T) map[string][]byte {
	t.Helper()
	seed, err := epochseed.Load("../..")
	if err != nil {
		t.Fatal(err)
	}
	artifacts := map[string][]byte{}
	for name, data := range seed.Artifacts {
		artifacts[name] = data
	}
	for name, path := range map[string]string{
		"economy":    "../../balance/testdata/t2/economy-candidate-v1.json",
		"routes":     "../../balance/testdata/t2/routes-candidate-v1.json",
		"categories": "../../balance/testdata/t2/categories-candidate-v1.json",
	} {
		if artifacts[name], err = os.ReadFile(path); err != nil {
			t.Fatal(err)
		}
	}
	return artifacts
}

func tier2CandidateBundle(t *testing.T) CatalogBundle {
	t.Helper()
	artifacts := tier2CandidateArtifacts(t)
	hash, err := save.ConstantsHashArtifacts(artifacts)
	if err != nil {
		t.Fatal(err)
	}
	// The TypeScript loader consumes the same pinned identity (tier2-candidates.test.ts).
	pinned, err := os.ReadFile("../../balance/testdata/t2/candidate-bundle-hash.txt")
	if err != nil || string(bytes.TrimSpace(pinned)) != hash {
		t.Fatalf("candidate bundle hash %s differs from the pinned %q (err=%v)", hash, pinned, err)
	}
	return loadCompleteReplayTestBundle(t, hash, artifacts)
}

// applyTier2Command runs one Company intent through ApplyLogged on a run-2
// Company (one prior Exit), so the run-1 scripted curriculum never intercepts.
func applyTier2Command(t *testing.T, bundle CatalogBundle, state *save.State, payload string, revision int64, now time.Time) (save.IntentOutcome, []byte) {
	t.Helper()
	request, err := ParseIntent([]byte(payload))
	if err != nil {
		t.Fatal(err)
	}
	command := save.ReplayCommand{IntentID: request.IntentID, CompanyStreamID: "01986666-1000-7000-8000-000000000001",
		FounderID: "01986666-2000-7000-8000-000000000001", Revision: revision, RunSeq: 2, RunLogSeq: revision}
	carry := tier2FounderCarry(t, bundle, now)
	activeEvidence, err := resolveActivePlaySchedule(state, bundle.Opportunities, bundle.Prestige, command.FounderID, now)
	if err != nil {
		t.Fatal(err)
	}
	catchup, err := buildOfflineCatchup(state, bundle.Economy, ModeOnline, now)
	if err != nil {
		t.Fatal(err)
	}
	inputs, err := buildReplayInputs(replayBuild{Command: command, Mode: ModeOnline, Now: now, IntentKind: request.Kind,
		RouteContextVersion: bundle.Routes.ContextVersion(), FounderCarry: &carry, ActivePlay: &activeEvidence, OfflineCatchup: catchup})
	if err != nil {
		t.Fatal(err)
	}
	transition, err := ApplyLogged(state, request.CanonicalPayload, bundle, inputs)
	if err != nil {
		t.Fatalf("%s: %v", payload, err)
	}
	if transition.Outcome == save.IntentRejected {
		t.Logf("%s rejected: %s", request.Kind, transition.Receipt)
	}
	return transition.Outcome, transition.Receipt
}

// tier2FounderCarry is a foundation-complete Founder with one prior Exit at the
// bundle's Founder floor, frozen as the Company command's carry.
func tier2FounderCarry(t *testing.T, bundle CatalogBundle, now time.Time) replayFounderCarry {
	t.Helper()
	founder := replayFounderFixtureState(t, bundle, now)
	founder.WireVersion, _ = bundle.versionFloors()
	if err := activateMinigameState(founder, bundle.Minigames); err != nil {
		t.Fatal(err)
	}
	founder.Pets = map[string]pet.CareState{}
	founder.FiscalCredit, founder.FiscalPeriodOpenedWallMS, founder.FiscalPeriodSequence = 0, now.UnixMilli(), 1
	founder.FiscalGeneratorLevels = make(map[string]int64, len(bundle.Fiscal.GeneratorLevelRows()))
	for _, row := range bundle.Fiscal.GeneratorLevelRows() {
		founder.FiscalGeneratorLevels[row.GeneratorID] = 0
	}
	founder.FiscalUnlocks = map[string]bool{}
	founder.Soul, founder.SoulExhaustedSourceIDs = 100, []string{}
	founder.ExitHistory = []save.ExitRecord{{RunID: 1, ExitType: "collapse", OccurredAt: now.Add(-time.Hour)}}
	if err := bundle.ValidateFoundationState(founder); err != nil {
		t.Fatalf("tier-2 Founder: %v", err)
	}
	carry := founderCarry(founder)
	carry.FounderRevision, carry.FounderConstantsHash = 2, bundle.ConstantsHash
	return carry
}

func tier2Company(t *testing.T, bundle CatalogBundle, now time.Time) *save.State {
	t.Helper()
	state := replayFixtureState(t, bundle.Economy, now)
	state.WireVersion = 18
	state.ActiveBuffs = []save.ActiveBuff{}
	state.MeterBands = nil
	meterState, err := meters.NewRunState(bundle.Meters, 9)
	if err != nil {
		t.Fatal(err)
	}
	state.MeterValues, state.MeterDecayRemainders, state.MeterInputRemainders = meterState.Values, meterState.DecayRemainders, meterState.InputRemainders
	state.AchievementsEarnedRun = map[string]bool{}
	state.RunSeq, state.Tier = 2, 1
	state.GatesCrossed["gate.t0_to_t1"] = true
	setCash(t, state, "2e7")
	if err := bundle.ValidateFoundationState(state); err != nil {
		t.Fatalf("tier-2 Company: %v", err)
	}
	return state
}

const (
	tier2CrossPayload       = `{"intent_id":"01986666-0901-7000-8000-000000000901","kind":"cross_gate","expected_revision":1,"gate_id":"gate.t1_to_t2","route_id":null}`
	tier2IncorporatePayload = `{"intent_id":"01986666-0902-7000-8000-000000000902","kind":"incorporate","expected_revision":2,"faction_id":"open_source"}`
)

func TestTier2CandidateReachesTierTwoAndIncorporation(t *testing.T) {
	bundle := tier2CandidateBundle(t)
	gate, ok := bundle.Routes.Gate("gate.t1_to_t2")
	if !ok || len(gate.Requirement) != 1 || gate.Requirement[0].Amount.String() != "1e7" {
		t.Fatalf("candidate gate = %+v exists=%v", gate, ok)
	}
	now := time.Date(2026, 9, 25, 12, 0, 0, 0, time.UTC)
	state := tier2Company(t, bundle, now)
	if outcome, _ := applyTier2Command(t, bundle, state, tier2CrossPayload, 1, now); outcome != save.IntentApplied || state.Tier != 2 || !state.GatesCrossed["gate.t1_to_t2"] {
		t.Fatalf("cross outcome=%s tier=%d gates=%v", outcome, state.Tier, state.GatesCrossed)
	}
	if outcome, _ := applyTier2Command(t, bundle, state, tier2IncorporatePayload, 2, now.Add(time.Second)); outcome != save.IntentApplied || state.FactionID != "open_source" {
		t.Fatalf("incorporate outcome=%s faction=%q", outcome, state.FactionID)
	}
}

// The AC0 failing case: the live epoch-8 bundle has no gate.t1_to_t2, so the
// same crossing is an unknown gate and a Tier-1 Company cannot incorporate.
func TestTier2EpochEightBundleCannotReachTierTwo(t *testing.T) {
	bundle := activeContentBundle(t)
	now := time.Date(2026, 9, 25, 12, 0, 0, 0, time.UTC)
	state := tier2Company(t, bundle, now)
	if outcome, receipt := applyTier2Command(t, bundle, state, tier2CrossPayload, 1, now); outcome != save.IntentRejected || state.Tier != 1 ||
		!bytes.Contains(receipt, []byte(`"rejection":{"category":"unknown_id","detail":"gate.t1_to_t2"}`)) {
		t.Fatalf("epoch-8 cross outcome=%s tier=%d receipt=%s", outcome, state.Tier, receipt)
	}
	incorporate := `{"intent_id":"01986666-0903-7000-8000-000000000903","kind":"incorporate","expected_revision":1,"faction_id":"open_source"}`
	if outcome, receipt := applyTier2Command(t, bundle, state, incorporate, 1, now); outcome != save.IntentRejected || state.FactionID != "" ||
		!bytes.Contains(receipt, []byte(`"rejection":{"category":"not_eligible","detail":"tier"}`)) {
		t.Fatalf("epoch-8 incorporate outcome=%s faction=%q receipt=%s", outcome, state.FactionID, receipt)
	}
}

func TestTier2CandidatesAreInsertionOnlyOverEpochEight(t *testing.T) {
	seed, err := epochseed.Load("../..")
	if err != nil {
		t.Fatal(err)
	}
	candidate := tier2CandidateArtifacts(t)
	// Economy and routes only insert lines: every live line survives in order.
	for _, name := range []string{"economy", "routes"} {
		lines, index := bytes.Split(seed.Artifacts[name], []byte("\n")), 0
		for _, line := range bytes.Split(candidate[name], []byte("\n")) {
			// A live line may only gain the trailing comma a following insertion needs.
			if index < len(lines) && (bytes.Equal(line, lines[index]) || bytes.Equal(line, append(append([]byte{}, lines[index]...), ','))) {
				index++
			}
		}
		if index != len(lines) {
			t.Fatalf("%s candidate deletes or edits live line %d: %q", name, index+1, lines[index])
		}
	}
	// Categories change exactly one byte range: the gate set gains the new gate.
	expected := bytes.Replace(seed.Artifacts["categories"], []byte(`["gate.t0_to_t1", `), []byte(`["gate.t0_to_t1", "gate.t1_to_t2", `), 1)
	if !bytes.Equal(expected, candidate["categories"]) {
		t.Fatal("categories candidate differs from epoch 8 beyond the gate-set insertion")
	}
	routeCatalog, err := routes.LoadCatalog(candidate["routes"])
	if err != nil {
		t.Fatal(err)
	}
	gateIDs := []string{}
	for _, gate := range routeCatalog.Gates() {
		gateIDs = append(gateIDs, gate.ID)
	}
	if _, err := leaderboard.LoadCategoryCatalog(candidate["categories"], gateIDs); err != nil {
		t.Fatalf("candidate categories: %v", err)
	}
	// AC1 subset failing case: the epoch-8 categories omit gate.t1_to_t2.
	if _, err := leaderboard.LoadCategoryCatalog(seed.Artifacts["categories"], gateIDs); err == nil {
		t.Fatal("epoch-8 categories accepted a gate set missing gate.t1_to_t2")
	}
}

package production

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"cloud-clicker/server/achievements"
	"cloud-clicker/server/decimal"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/epochseed"
	"cloud-clicker/server/meters"
	"cloud-clicker/server/minigame"
	"cloud-clicker/server/multiplier"
	"cloud-clicker/server/pet"
	prestigecore "cloud-clicker/server/prestige"
	"cloud-clicker/server/save"
)

type achievementParityFixture struct {
	Baseline json.RawMessage `json:"baseline"`
	Registry struct {
		CopyKeys          []string            `json:"copy_keys"`
		GeneratorIDs      []string            `json:"generator_ids"`
		EventKinds        []string            `json:"event_kinds"`
		ResourceIDs       []string            `json:"resource_ids"`
		RunCounters       []string            `json:"run_counters"`
		CareerCounters    []string            `json:"career_counters"`
		ProvenanceSources map[string][]string `json:"provenance_sources"`
	} `json:"registry"`
}

func foundationAchievementRegistry(fixture achievementParityFixture) achievements.Registry {
	return achievements.Registry{
		CopyKeys: toFoundationSet(fixture.Registry.CopyKeys), GeneratorIDs: toFoundationSet(fixture.Registry.GeneratorIDs),
		EventKinds: toFoundationSet(fixture.Registry.EventKinds), ResourceIDs: toFoundationSet(fixture.Registry.ResourceIDs),
		RunCounters: toFoundationSet(fixture.Registry.RunCounters), CareerCounters: toFoundationSet(fixture.Registry.CareerCounters),
		ProvenanceSources: fixture.Registry.ProvenanceSources,
	}
}

func foundationTestBundles(t *testing.T) (CatalogBundle, CatalogBundle) {
	t.Helper()
	seed, err := epochseed.Load("../..")
	if err != nil {
		t.Fatal(err)
	}
	legacy := loadReplayTestBundle(t, seed.Hash, seed.Artifacts)
	metersFixture, err := os.ReadFile("../../balance/testdata/meters-catalog-parity-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	var meterEnvelope struct {
		Baseline json.RawMessage `json:"baseline"`
	}
	if json.Unmarshal(metersFixture, &meterEnvelope) != nil {
		t.Fatal("decode meter fixture")
	}
	meterCatalog, err := meters.LoadCatalog(meterEnvelope.Baseline)
	if err != nil {
		t.Fatal(err)
	}
	achievementArtifact := []byte(`{"schema_version":1,"achievements":[{"id":"achievement.first_gate","condition_scope":"run","condition":{"kind":"counter_at_least","counter":"tier","minimum":1},"proof":{"kind":"provenance","event_kinds":["gate_crossed"]},"score_grant":4,"copy_key":"category.any_percent"},{"id":"achievement.old_hand","condition_scope":"career","condition":{"kind":"exit_count_at_least","count":1},"proof":{"kind":"provenance","event_kinds":["founder_advanced","run_ended"]},"score_grant":6,"copy_key":"category.any_percent"}]}`)
	achievementCatalog, err := achievements.LoadCatalog(achievementArtifact, FoundationAchievementRegistry(legacy.Economy))
	if err != nil {
		t.Fatal(err)
	}
	artifacts := make(map[string][]byte, len(seed.Artifacts)+2)
	for name, data := range seed.Artifacts {
		artifacts[name] = append([]byte(nil), data...)
	}
	artifacts["meters"] = append([]byte(nil), meterEnvelope.Baseline...)
	artifacts["achievements"] = append([]byte(nil), achievementArtifact...)
	hash, err := save.ConstantsHashArtifacts(artifacts)
	if err != nil {
		t.Fatal(err)
	}
	active := legacy
	active.ConstantsHash, active.Artifacts, active.Meters, active.Achievements = hash, artifacts, meterCatalog, achievementCatalog
	if !active.valid(hash) {
		t.Fatal("foundation bundle is not internally valid")
	}
	return legacy, active
}

func founderFeatureBundles(t *testing.T, active CatalogBundle) (CatalogBundle, CatalogBundle) {
	t.Helper()
	minigameBytes := []byte(`{"schema_version":1,"minigame_ids":[],"rating_seasons":[]}`)
	minigameCatalog, err := minigame.LoadCatalog(minigameBytes)
	if err != nil {
		t.Fatal(err)
	}
	withMinigames := active
	withMinigames.Artifacts = cloneArtifactMap(active.Artifacts)
	withMinigames.Artifacts["minigames"] = minigameBytes
	withMinigames.ConstantsHash, err = save.ConstantsHashArtifacts(withMinigames.Artifacts)
	if err != nil {
		t.Fatal(err)
	}
	withMinigames.Minigames = minigameCatalog
	petBytes := []byte(`{"schema_version":1,"stat_policy":{"grid_ms":60000,"stats":[{"stat_id":"hunger","initial_ppm":800000,"floor_ppm":100000,"decay_ppm_per_grid":1000},{"stat_id":"energy","initial_ppm":800000,"floor_ppm":100000,"decay_ppm_per_grid":1000},{"stat_id":"cleanliness","initial_ppm":800000,"floor_ppm":100000,"decay_ppm_per_grid":1000},{"stat_id":"affection","initial_ppm":800000,"floor_ppm":100000,"decay_ppm_per_grid":1000}],"diminishing_threshold_ppm":700000,"diminishing_factor_ppm":500000},"actions":[{"action_id":"care.feed","stat_id":"hunger","delta_ppm":100000,"cooldown_attended_ms":60000,"min_eligible_ppm":100000}],"trust_policy":{"initial_ppm":500000,"neutral_ppm":500000,"floor_ppm":100000,"cap_ppm":1000000,"gain_ppm_per_effective_action":1000,"decay_ppm_per_grid":100},"mood_policy":[{"mood_member":"withdrawn","floor_ppm":0},{"mood_member":"restless","floor_ppm":250000},{"mood_member":"neutral","floor_ppm":500000},{"mood_member":"engaged","floor_ppm":750000}],"behavior_policy":[]}`)
	petCatalog, err := pet.LoadCatalog(petBytes)
	if err != nil {
		t.Fatal(err)
	}
	withPets := withMinigames
	withPets.Artifacts = cloneArtifactMap(withMinigames.Artifacts)
	withPets.Artifacts["pets"] = petBytes
	withPets.ConstantsHash, err = save.ConstantsHashArtifacts(withPets.Artifacts)
	if err != nil {
		t.Fatal(err)
	}
	withPets.Pets = petCatalog
	if !withMinigames.valid(withMinigames.ConstantsHash) || !withPets.valid(withPets.ConstantsHash) {
		t.Fatal("Founder feature bundles invalid")
	}
	return withMinigames, withPets
}

func cloneArtifactMap(source map[string][]byte) map[string][]byte {
	result := make(map[string][]byte, len(source)+1)
	for key, value := range source {
		result[key] = append([]byte(nil), value...)
	}
	return result
}

func TestFounderFeatureVersionsAreReachableWithoutChangingCompanyAxis(t *testing.T) {
	legacy, active := foundationTestBundles(t)
	minigames, pets := founderFeatureBundles(t, active)
	founder := foundationScopeState(t, active.Economy, economy.ScopeFounder)
	company := foundationScopeState(t, active.Economy, economy.ScopeCompany)
	firstCompany := foundationScopeState(t, active.Economy, economy.ScopeCompany)
	if err := settleAndActivateFoundations(legacy, active, founder, company, firstCompany); err != nil {
		t.Fatal(err)
	}
	secondCompany := foundationScopeState(t, active.Economy, economy.ScopeCompany)
	if err := settleAndActivateFoundations(active, minigames, founder, firstCompany, secondCompany); err != nil {
		t.Fatal(err)
	}
	if save.VersionForState(founder) != 17 || save.VersionForState(secondCompany) != 16 {
		t.Fatalf("mixed axes founder=%d company=%d", save.VersionForState(founder), save.VersionForState(secondCompany))
	}
	partialCarry := foundationScopeState(t, active.Economy, economy.ScopeFounder)
	partialCarry.WireVersion, partialCarry.AchievementsEarnedLifetime = 16, map[string]bool{}
	if err := validateFoundationHookInputs(minigames, secondCompany, partialCarry); err != nil {
		t.Fatalf("v17 bundle rejected partial Company-transition carry: %v", err)
	}
	thirdCompany := foundationScopeState(t, active.Economy, economy.ScopeCompany)
	if err := settleAndActivateFoundations(minigames, pets, founder, secondCompany, thirdCompany); err != nil {
		t.Fatal(err)
	}
	if save.VersionForState(founder) != 18 || save.VersionForState(thirdCompany) != 16 || founder.Pets == nil {
		t.Fatalf("v18 axes/state founder=%d company=%d pets=%v", save.VersionForState(founder), save.VersionForState(thirdCompany), founder.Pets)
	}
}

func TestFounderExitReplayActivatesV17AndV18FromPinnedNextBundle(t *testing.T) {
	legacy, active := foundationTestBundles(t)
	minigames, pets := founderFeatureBundles(t, active)
	founder := foundationScopeState(t, active.Economy, economy.ScopeFounder)
	company := foundationScopeState(t, active.Economy, economy.ScopeCompany)
	newCompany := foundationScopeState(t, active.Economy, economy.ScopeCompany)
	if err := settleAndActivateFoundations(legacy, active, founder, company, newCompany); err != nil {
		t.Fatal(err)
	}
	command := save.FounderReplayCommand{IntentID: "01986666-1900-7000-8000-000000000001", FounderStreamID: "01986666-1900-7000-8000-000000000002", FounderID: "01986666-1900-7000-8000-000000000003", Revision: 1, FounderLogSeq: 1, ServerTSMS: 1}
	active.Next = &minigames
	resolved := founderExitResolvedWire{Kind: founderExitResolvedKind, Outcome: string(save.IntentApplied), CompanyStreamID: "01986666-1900-7000-8000-000000000004", RunSeq: 1, RunLogSeq: 1, ResultConstantsHash: minigames.ConstantsHash, AgeMSBefore: founder.AgeMS, AgeMSAfter: founder.AgeMS, AddedNetworkSlots: []save.NetworkSlot{}, AddedLedgerFactKinds: []string{}, AddedLifetimeAchievements: []string{}, ExitRecord: &founderExitRecordWire{RunID: 1, ExitType: "collapse", OccurredAtMS: 1}, ResultFounderWireVersion: 17}
	transition, err := applyFounderExitResolved(founder, command, IntentRequest{}, active, resolved)
	if err != nil || save.VersionForState(transition.State) != 17 || transition.State.MinigameRatings == nil {
		t.Fatalf("v17 replay transition=%+v err=%v", transition.State, err)
	}
	minigames.Next = &pets
	command.Revision, command.FounderLogSeq, command.IntentID = 2, 2, "01986666-1900-7000-8000-000000000005"
	resolved.RunSeq, resolved.RunLogSeq, resolved.ResultConstantsHash, resolved.ResultFounderWireVersion = 2, 2, pets.ConstantsHash, 18
	resolved.ExitRecord = &founderExitRecordWire{RunID: 2, ExitType: "collapse", OccurredAtMS: 2}
	transition, err = applyFounderExitResolved(founder, command, IntentRequest{}, minigames, resolved)
	if err != nil || save.VersionForState(transition.State) != 18 || transition.State.Pets == nil {
		t.Fatalf("v18 replay transition=%+v err=%v", transition.State, err)
	}
}

func retunedAchievementBundle(t *testing.T, active CatalogBundle, scoreGrant int64) CatalogBundle {
	t.Helper()
	var raw map[string]any
	if err := json.Unmarshal(active.Artifacts["achievements"], &raw); err != nil {
		t.Fatal(err)
	}
	rows := raw["achievements"].([]any)
	rows[0].(map[string]any)["score_grant"] = scoreGrant
	artifact, err := json.Marshal(raw)
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := achievements.LoadCatalog(artifact, FoundationAchievementRegistry(active.Economy))
	if err != nil {
		t.Fatal(err)
	}
	artifacts := make(map[string][]byte, len(active.Artifacts))
	for name, data := range active.Artifacts {
		artifacts[name] = append([]byte(nil), data...)
	}
	artifacts["achievements"] = artifact
	hash, err := save.ConstantsHashArtifacts(artifacts)
	if err != nil {
		t.Fatal(err)
	}
	active.Artifacts, active.ConstantsHash, active.Achievements = artifacts, hash, catalog
	if !active.valid(hash) {
		t.Fatal("retuned foundation bundle is not internally valid")
	}
	return active
}

func toFoundationSet(values []string) map[string]bool {
	result := make(map[string]bool, len(values))
	for _, value := range values {
		result[value] = true
	}
	return result
}

func foundationScopeState(t *testing.T, catalog *economy.Catalog, scope economy.Scope) *save.State {
	t.Helper()
	ledger, err := economy.NewLedger(catalog, scope)
	if err != nil {
		t.Fatal(err)
	}
	return &save.State{Ledger: ledger, WireVersion: save.CurrentVersion}
}

func TestFoundationActivationIsAtomicNewRunOnly(t *testing.T) {
	legacy, active := foundationTestBundles(t)
	founder := foundationScopeState(t, legacy.Economy, economy.ScopeFounder)
	founder.Notoriety = 100
	company := foundationScopeState(t, legacy.Economy, economy.ScopeCompany)
	company.MeterBands = map[string]int{"trust.users.standing": 99}
	newCompany := foundationScopeState(t, active.Economy, economy.ScopeCompany)

	if err := settleAndActivateFoundations(legacy, active, founder, company, newCompany); err != nil {
		t.Fatal(err)
	}
	if save.VersionForState(founder) != 16 || save.VersionForState(newCompany) != 16 || save.VersionForState(company) != 14 {
		t.Fatalf("versions founder=%d old=%d new=%d", save.VersionForState(founder), save.VersionForState(company), save.VersionForState(newCompany))
	}
	if newCompany.MeterValues["trust.users.standing"] != 55 || newCompany.MeterValues["trust.users.grievance"] != 50 ||
		newCompany.MeterValues["trust.users.standing"] == company.MeterBands["trust.users.standing"] {
		t.Fatalf("new meter state=%v legacy=%v", newCompany.MeterValues, company.MeterBands)
	}
	if len(newCompany.AchievementsEarnedRun) != 0 || newCompany.AchievementScoreRun != 0 || len(founder.AchievementsEarnedLifetime) != 0 {
		t.Fatalf("activation retroactively earned achievements: founder=%+v company=%+v", founder, newCompany)
	}
	if err := active.ValidateFoundationState(founder); err != nil {
		t.Fatal(err)
	}
	if err := active.ValidateFoundationState(newCompany); err != nil {
		t.Fatal(err)
	}
	context, err := routeContext(newCompany, active.Routes.ContextVersion())
	if err != nil || context.MeterBands["trust.users.standing"] != 55 {
		t.Fatalf("v16 route context meters=%v err=%v", context.MeterBands, err)
	}
}

func TestFoundationExitSettlesAndDerivesAchievementScore(t *testing.T) {
	legacy, active := foundationTestBundles(t)
	founder := foundationScopeState(t, legacy.Economy, economy.ScopeFounder)
	oldCompany := foundationScopeState(t, legacy.Economy, economy.ScopeCompany)
	company := foundationScopeState(t, active.Economy, economy.ScopeCompany)
	if err := settleAndActivateFoundations(legacy, active, founder, oldCompany, company); err != nil {
		t.Fatal(err)
	}
	definition := active.Achievements.Definitions[0]
	company.AchievementsEarnedRun[definition.ID] = true
	company.AchievementScoreRun = definition.ScoreGrant
	nextCompany := foundationScopeState(t, active.Economy, economy.ScopeCompany)
	if err := settleAndActivateFoundations(active, active, founder, company, nextCompany); err != nil {
		t.Fatal(err)
	}
	if !founder.AchievementsEarnedLifetime[definition.ID] || founder.AchievementScoreLifetime != definition.ScoreGrant ||
		len(nextCompany.AchievementsEarnedRun) != 0 || nextCompany.AchievementScoreRun != 0 {
		t.Fatalf("settlement founder=%+v next=%+v", founder, nextCompany)
	}
	company.AchievementScoreRun++
	if err := active.ValidateFoundationState(company); err == nil {
		t.Fatal("non-derived run score was accepted")
	}
}

func TestFoundationExitRederivesLifetimeScoreUnderNextCatalog(t *testing.T) {
	legacy, active := foundationTestBundles(t)
	founder := foundationScopeState(t, legacy.Economy, economy.ScopeFounder)
	oldCompany := foundationScopeState(t, legacy.Economy, economy.ScopeCompany)
	company := foundationScopeState(t, active.Economy, economy.ScopeCompany)
	if err := settleAndActivateFoundations(legacy, active, founder, oldCompany, company); err != nil {
		t.Fatal(err)
	}
	definition := active.Achievements.Definitions[0]
	company.AchievementsEarnedRun[definition.ID] = true
	company.AchievementScoreRun = definition.ScoreGrant
	next := retunedAchievementBundle(t, active, definition.ScoreGrant+1)
	nextCompany := foundationScopeState(t, next.Economy, economy.ScopeCompany)
	if err := settleAndActivateFoundations(active, next, founder, company, nextCompany); err != nil {
		t.Fatalf("honest Exit across score retune failed: %v", err)
	}
	if founder.AchievementScoreLifetime != definition.ScoreGrant+1 {
		t.Fatalf("lifetime score = %d, want next-catalog grant %d", founder.AchievementScoreLifetime, definition.ScoreGrant+1)
	}
	if err := next.ValidateFoundationState(founder); err != nil {
		t.Fatalf("next-catalog Founder state rejected: %v", err)
	}
}

func TestFoundationArtifactPairAndDowngradeFailClosed(t *testing.T) {
	legacy, active := foundationTestBundles(t)
	broken := active
	broken.Achievements = nil
	if broken.valid(active.ConstantsHash) {
		t.Fatal("single foundation artifact was accepted")
	}
	founder := foundationScopeState(t, legacy.Economy, economy.ScopeFounder)
	oldCompany := foundationScopeState(t, legacy.Economy, economy.ScopeCompany)
	company := foundationScopeState(t, active.Economy, economy.ScopeCompany)
	if err := settleAndActivateFoundations(legacy, active, founder, oldCompany, company); err != nil {
		t.Fatal(err)
	}
	if err := settleAndActivateFoundations(active, legacy, founder, company, foundationScopeState(t, legacy.Economy, economy.ScopeCompany)); err == nil {
		t.Fatal("foundation artifact downgrade was accepted")
	}
}

func TestFoundationTransitionOrdersMetersBeforeAchievementsAndLatches(t *testing.T) {
	_, active := foundationTestBundles(t)
	now := time.Date(2026, 8, 4, 15, 0, 0, 0, time.UTC)
	state := replayFixtureState(t, active.Economy, now.Add(-time.Hour))
	state.WireVersion = save.LatestSupportedVersion
	meterState, err := meters.NewRunState(active.Meters, 0)
	if err != nil {
		t.Fatal(err)
	}
	state.MeterBands = nil
	state.MeterValues, state.MeterDecayRemainders, state.MeterInputRemainders = meterState.Values, meterState.DecayRemainders, meterState.InputRemainders
	state.MeterValues["doom.probability"] = 71
	state.AchievementsEarnedRun = map[string]bool{}
	state.Tier = 1
	before, err := cloneReplayState(state, active.Economy)
	if err != nil {
		t.Fatal(err)
	}
	state.EvaluatedThrough = now
	founder := foundationScopeState(t, active.Economy, economy.ScopeFounder)
	founder.WireVersion = save.LatestSupportedVersion
	founder.AchievementsEarnedLifetime = map[string]bool{}
	events := []save.EventWrite{}
	request := IntentRequest{IntentID: "01986666-0600-7000-8000-000000000601"}
	revision := save.Revision{StreamID: "01986666-1600-7000-8000-000000000001", OwnerID: "01986666-2600-7000-8000-000000000001"}
	if err := applyFoundationTransition(active, before, state, founder, revision, request, now, nil, map[string]string{}, false, &events); err != nil {
		t.Fatal(err)
	}
	if len(events) != 2 || events[0].Kind != save.EventMeterBandChanged || events[1].Kind != save.EventAchievementEarned {
		t.Fatalf("foundation event order=%+v", events)
	}
	if state.MeterValues["doom.probability"] != 69 || !state.AchievementsEarnedRun["achievement.first_gate"] || state.AchievementScoreRun != 4 {
		t.Fatalf("foundation state meter=%d earned=%v score=%d", state.MeterValues["doom.probability"], state.AchievementsEarnedRun, state.AchievementScoreRun)
	}
	before, err = cloneReplayState(state, active.Economy)
	if err != nil {
		t.Fatal(err)
	}
	events = nil
	if err := applyFoundationTransition(active, before, state, founder, revision, request, now, nil, map[string]string{}, false, &events); err != nil {
		t.Fatal(err)
	}
	if len(events) != 0 || state.AchievementScoreRun != 4 {
		t.Fatalf("latched transition events=%+v score=%d", events, state.AchievementScoreRun)
	}
}

func TestFoundationTransitionBurnProofRequiresSameBatchDebit(t *testing.T) {
	_, active := foundationTestBundles(t)
	artifact := []byte(`{"schema_version":1,"achievements":[{"id":"achievement.burn","condition_scope":"run","condition":{"kind":"counter_at_least","counter":"tier","minimum":1},"proof":{"kind":"burn","event_kind":"gate_crossed","resource_id":"company.cash","minimum":"1e1"},"score_grant":3,"copy_key":"category.any_percent"}]}`)
	catalog, err := achievements.LoadCatalog(artifact, FoundationAchievementRegistry(active.Economy))
	if err != nil {
		t.Fatal(err)
	}
	artifacts := make(map[string][]byte, len(active.Artifacts))
	for name, data := range active.Artifacts {
		artifacts[name] = append([]byte(nil), data...)
	}
	artifacts["achievements"] = artifact
	hash, err := save.ConstantsHashArtifacts(artifacts)
	if err != nil {
		t.Fatal(err)
	}
	active.Artifacts, active.ConstantsHash, active.Achievements = artifacts, hash, catalog
	now := time.Date(2026, 8, 4, 16, 0, 0, 0, time.UTC)
	makeState := func() (*save.State, *save.State, *save.State) {
		state := replayFixtureState(t, active.Economy, now)
		state.WireVersion = save.LatestSupportedVersion
		meterState := meters.NewState(active.Meters)
		state.MeterBands = nil
		state.MeterValues, state.MeterDecayRemainders, state.MeterInputRemainders = meterState.Values, meterState.DecayRemainders, meterState.InputRemainders
		state.AchievementsEarnedRun = map[string]bool{}
		state.Tier = 1
		setCash(t, state, "1e2")
		before, cloneErr := cloneReplayState(state, active.Economy)
		if cloneErr != nil {
			t.Fatal(cloneErr)
		}
		founder := foundationScopeState(t, active.Economy, economy.ScopeFounder)
		founder.WireVersion = save.LatestSupportedVersion
		founder.AchievementsEarnedLifetime = map[string]bool{}
		return before, state, founder
	}
	request := IntentRequest{IntentID: "01986666-0600-7000-8000-000000000602"}
	revision := save.Revision{StreamID: "01986666-1600-7000-8000-000000000002", OwnerID: "01986666-2600-7000-8000-000000000002"}
	for _, testCase := range []struct {
		name       string
		debit      bool
		withEvent  bool
		wantEarned bool
	}{{"complete-proof", true, true, true}, {"missing-debit", false, true, false}, {"missing-event", true, false, false}} {
		t.Run(testCase.name, func(t *testing.T) {
			before, state, founder := makeState()
			actionDebits := map[string]string{}
			if testCase.debit {
				actionDebits["company.cash"] = "1e1"
			}
			events := []save.EventWrite{}
			if testCase.withEvent {
				events = append(events, save.EventWrite{Kind: save.EventGateCrossed, IntentID: request.IntentID})
			}
			if err := applyFoundationTransition(active, before, state, founder, revision, request, now, nil, actionDebits, false, &events); err != nil {
				t.Fatal(err)
			}
			if state.AchievementsEarnedRun["achievement.burn"] != testCase.wantEarned {
				t.Fatalf("earned=%v events=%+v", state.AchievementsEarnedRun, events)
			}
		})
	}
}

func TestActionDebitsSeparateAccrualFromTheActionSink(t *testing.T) {
	debits, err := actionDebits(map[string]string{"company.cash": "2e2"}, map[string]string{"company.cash": "1.9e2"})
	if err != nil {
		t.Fatal(err)
	}
	if debits["company.cash"] != "1e1" {
		t.Fatalf("action debit = %q, want 1e1", debits["company.cash"])
	}
}

func TestFoundationTransitionUsesTheCanonicalOfflineLedger(t *testing.T) {
	_, active := foundationTestBundles(t)
	for _, duration := range []time.Duration{5001 * time.Millisecond, 25 * time.Hour} {
		t.Run(duration.String(), func(t *testing.T) {
			now := time.Date(2026, 8, 4, 17, 0, 0, 0, time.UTC)
			before := replayFixtureState(t, active.Economy, now.Add(-duration))
			before.WireVersion = save.LatestSupportedVersion
			meterState, err := meters.NewRunState(active.Meters, 0)
			if err != nil {
				t.Fatal(err)
			}
			before.MeterBands = nil
			before.MeterValues, before.MeterDecayRemainders, before.MeterInputRemainders = meterState.Values, meterState.DecayRemainders, meterState.InputRemainders
			before.MeterValues["doom.probability"] = 71
			before.AchievementsEarnedRun = map[string]bool{}
			state, err := cloneReplayState(before, active.Economy)
			if err != nil {
				t.Fatal(err)
			}
			state.EvaluatedThrough = now
			if err := prestigecore.RecordOfflineSpan(state, before.EvaluatedThrough, now, 5000); err != nil {
				t.Fatal(err)
			}
			founder := foundationScopeState(t, active.Economy, economy.ScopeFounder)
			founder.WireVersion = save.LatestSupportedVersion
			founder.AchievementsEarnedLifetime = map[string]bool{}
			events := []save.EventWrite{}
			request := IntentRequest{IntentID: "01986666-0600-7000-8000-000000000603"}
			revision := save.Revision{StreamID: "01986666-1600-7000-8000-000000000003", OwnerID: "01986666-2600-7000-8000-000000000003"}
			if err := applyFoundationTransition(active, before, state, founder, revision, request, now, nil, map[string]string{}, false, &events); err != nil {
				t.Fatal(err)
			}
			if state.MeterValues["doom.probability"] != 71 || len(events) != 0 {
				t.Fatalf("offline return changed meter=%d events=%+v", state.MeterValues["doom.probability"], events)
			}
		})
	}
}

func TestFoundationTransitionDerivesFactAndContributionInputs(t *testing.T) {
	_, active := foundationTestBundles(t)
	now := time.Date(2026, 8, 4, 18, 0, 0, 0, time.UTC)
	makeState := func(start time.Time) (*save.State, *save.State, *save.State) {
		before := replayFixtureState(t, active.Economy, start)
		before.WireVersion = save.LatestSupportedVersion
		meterState, err := meters.NewRunState(active.Meters, 0)
		if err != nil {
			t.Fatal(err)
		}
		before.MeterBands = nil
		before.MeterValues, before.MeterDecayRemainders, before.MeterInputRemainders = meterState.Values, meterState.DecayRemainders, meterState.InputRemainders
		before.AchievementsEarnedRun = map[string]bool{}
		state, err := cloneReplayState(before, active.Economy)
		if err != nil {
			t.Fatal(err)
		}
		founder := foundationScopeState(t, active.Economy, economy.ScopeFounder)
		founder.WireVersion = save.LatestSupportedVersion
		founder.AchievementsEarnedLifetime = map[string]bool{}
		return before, state, founder
	}
	request := IntentRequest{IntentID: "01986666-0600-7000-8000-000000000604"}
	revision := save.Revision{StreamID: "01986666-1600-7000-8000-000000000004", OwnerID: "01986666-2600-7000-8000-000000000004"}

	before, state, founder := makeState(now)
	before.MeterValues["doom.probability"], state.MeterValues["doom.probability"] = 67, 67
	state.LedgerFactKinds["externality.emitted"] = true
	events := []save.EventWrite{}
	if err := applyFoundationTransition(active, before, state, founder, revision, request, now, nil, map[string]string{}, false, &events); err != nil {
		t.Fatal(err)
	}
	if state.MeterValues["doom.probability"] != 70 || len(events) != 1 || events[0].Kind != save.EventMeterBandChanged {
		t.Fatalf("fact input meter=%d events=%+v", state.MeterValues["doom.probability"], events)
	}

	before, state, founder = makeState(now.Add(-4 * time.Second))
	before.MeterValues["doom.probability"], state.MeterValues["doom.probability"] = 70, 70
	before.MeterInputRemainders[meters.InputRemainderKey("doom.probability", 1)] = 3_596_000
	state.MeterInputRemainders[meters.InputRemainderKey("doom.probability", 1)] = 3_596_000
	state.EvaluatedThrough = now
	events = nil
	contributions := []multiplier.Contribution{{Slot: multiplier.SlotUpgrades, SourceID: "generator.example", Target: "all", Factor: decimal.New(2, 0)}}
	if err := applyFoundationTransition(active, before, state, founder, revision, request, now, contributions, map[string]string{}, false, &events); err != nil {
		t.Fatal(err)
	}
	if state.MeterValues["doom.probability"] != 69 || len(events) != 1 || events[0].Kind != save.EventMeterBandChanged {
		t.Fatalf("contribution input meter=%d events=%+v", state.MeterValues["doom.probability"], events)
	}
}

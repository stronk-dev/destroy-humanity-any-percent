package production

import (
	"bytes"
	"encoding/json"
	"os"
	"reflect"
	"sort"
	"testing"
	"time"

	"cloud-clicker/server/achievements"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/epochseed"
	"cloud-clicker/server/meters"
	"cloud-clicker/server/save"
)

const attainmentVectorsPath = "../../testdata/axis-stack/attainment-vectors-v1.json"

// axisContentBundle is the live epoch-8 bundle with the economy replaced by
// the Clout v1 schema-5 fixture (fixture-first; no epoch pins it yet).
func axisContentBundle(t *testing.T) CatalogBundle {
	t.Helper()
	return axisContentBundleWithInput(t, "achievement_attainment_run")
}

func axisContentBundleWithInput(t *testing.T, input string) CatalogBundle {
	t.Helper()
	seed, err := epochseed.Load("../..")
	if err != nil {
		t.Fatal(err)
	}
	artifacts := map[string][]byte{}
	for name, data := range seed.Artifacts {
		artifacts[name] = data
	}
	fixture, err := os.ReadFile("../../balance/testdata/axis-stack/economy-v5-fixture.json")
	if err != nil {
		t.Fatal(err)
	}
	var root map[string]any
	if err := json.Unmarshal(fixture, &root); err != nil {
		t.Fatal(err)
	}
	root["axis_stack"].(map[string]any)["input"] = input
	if artifacts["economy"], err = json.Marshal(root); err != nil {
		t.Fatal(err)
	}
	hash, err := save.ConstantsHashArtifacts(artifacts)
	if err != nil {
		t.Fatal(err)
	}
	bundle := activeContentBundle(t)
	bundle.ConstantsHash, bundle.Artifacts = hash, artifacts
	if bundle.Economy, err = economy.LoadCatalog(artifacts["economy"]); err != nil {
		t.Fatal(err)
	}
	if !bundle.valid(hash) {
		t.Fatal("axis bundle is not valid")
	}
	return bundle
}

func TestAxisStackOwnsCompanyV19Activation(t *testing.T) {
	active := activeContentBundle(t)
	axis := axisContentBundle(t)
	if _, company := axis.versionFloors(); company != 19 {
		t.Fatalf("axis bundle company floor = %d", company)
	}
	if _, company := active.versionFloors(); company != 18 {
		t.Fatalf("epoch-8 company floor = %d", company)
	}
	// Run boundary from a pre-foundation run into the axis bundle (the same
	// harness the Reputation activation test uses).
	legacy := epoch5TestBundle(t)
	now := time.Date(2026, 9, 25, 12, 0, 0, 0, time.UTC)
	founder := foundationScopeState(t, legacy.Economy, economy.ScopeFounder)
	company := foundationScopeState(t, legacy.Economy, economy.ScopeCompany)
	company.RunStartedAt = now.Add(-2 * time.Hour)
	next := foundationScopeState(t, axis.Economy, economy.ScopeCompany)
	next.RunStartedAt = now.Add(-time.Hour)
	next.RunSeq = 2
	next.AchievementsAttainedRun, next.AttainmentScoreRun = map[string]bool{"achievement.first_gate": true}, 2
	// Exit order: foundations settle first, then the active-play scheduler
	// initializes; the latter must not lower the v19 floor back to v18.
	if err := settleAndActivateFoundations(legacy, axis, founder, company, next); err != nil {
		t.Fatal(err)
	}
	if _, err := initializeActivePlayState(next, axis.Opportunities, "01986666-0000-7000-8000-000000000001"); err != nil {
		t.Fatal(err)
	}
	if save.VersionForState(next) != 19 || next.AchievementsAttainedRun == nil || len(next.AchievementsAttainedRun) != 0 || next.AttainmentScoreRun != 0 {
		t.Fatalf("new run = v%d attained=%v score=%d", save.VersionForState(next), next.AchievementsAttainedRun, next.AttainmentScoreRun)
	}
}

func TestAttainmentDerivationIsValidated(t *testing.T) {
	bundle := axisContentBundle(t)
	company := foundationScopeState(t, bundle.Economy, economy.ScopeCompany)
	company.AchievementsAttainedRun = map[string]bool{"achievement.first_gate": true}
	company.AttainmentScoreRun = 2
	if err := validateAttainmentState(bundle, company); err != nil {
		t.Fatal(err)
	}
	for name, mutate := range map[string]func(*save.State){
		"tampered_score": func(state *save.State) { state.AttainmentScoreRun = 3 },
		"career_definition": func(state *save.State) {
			state.AchievementsAttainedRun["achievement.old_hand"] = true
			state.AttainmentScoreRun = 6
		},
		"unknown_definition":    func(state *save.State) { state.AchievementsAttainedRun["achievement.unknown"] = true },
		"earned_not_attained":   func(state *save.State) { state.AchievementsEarnedRun["achievement.generators_purchased_1"] = true },
		"nil_set_at_validation": func(state *save.State) { state.AchievementsAttainedRun = nil },
	} {
		t.Run(name, func(t *testing.T) {
			state := foundationScopeState(t, bundle.Economy, economy.ScopeCompany)
			state.AchievementsEarnedRun = map[string]bool{}
			state.AchievementsAttainedRun = map[string]bool{"achievement.first_gate": true}
			state.AttainmentScoreRun = 2
			mutate(state)
			if validateAttainmentState(bundle, state) == nil {
				t.Fatal("invalid attainment state accepted")
			}
		})
	}
}

// attainmentVector is one shared Go/TS case for the CV2 second pass.
type attainmentVector struct {
	Name            string           `json:"name"`
	AttainedBefore  []string         `json:"attained_before"`
	EarnedNow       []string         `json:"earned_now"`
	Counters        map[string]int64 `json:"counters"`
	Generators      map[string]int64 `json:"generators"`
	EventKinds      []string         `json:"event_kinds"`
	CashDebit       *string          `json:"cash_debit"`
	AttainedAfter   []string         `json:"attained_after"`
	ScoreAfter      int64            `json:"score_after"`
	ReattainedOrder []string         `json:"reattained_order"`
}

func attainmentCases() []attainmentVector {
	debit := "1e9"
	return []attainmentVector{
		{Name: "veteran_reattains_first_purchase", Counters: map[string]int64{"generators_purchased_total": 1, "tier": 0}, EventKinds: []string{"generator_purchased"}},
		{Name: "newly_earned_attains_without_reattained_event", EarnedNow: []string{"achievement.generators_purchased_1"}, Counters: map[string]int64{"generators_purchased_total": 1, "tier": 0}, EventKinds: []string{"generator_purchased"}},
		{Name: "latched_id_grants_once", AttainedBefore: []string{"achievement.generators_purchased_1"}, Counters: map[string]int64{"generators_purchased_total": 30, "tier": 0}, EventKinds: []string{"generator_purchased"}},
		{Name: "provenance_without_event_does_not_attain", Counters: map[string]int64{"generators_purchased_total": 1, "tier": 0}},
		{Name: "burn_without_debit_does_not_attain", Counters: map[string]int64{"generators_purchased_total": 0, "tier": 3}, EventKinds: []string{"gate_crossed"}},
		{Name: "burn_with_debit_attains", Counters: map[string]int64{"generators_purchased_total": 0, "tier": 3}, EventKinds: []string{"gate_crossed"}, CashDebit: &debit},
		{Name: "possession_attains_on_count", Counters: map[string]int64{"generators_purchased_total": 0, "tier": 0}, Generators: map[string]int64{"generator.beige_tower": 100}},
		{Name: "career_definitions_never_attain", Counters: map[string]int64{"generators_purchased_total": 0, "tier": 0, "age_ms": 1e9}, EventKinds: []string{"founder_advanced", "run_ended"}},
		{Name: "byte_order_of_several", Counters: map[string]int64{"generators_purchased_total": 25, "tier": 1}, EventKinds: []string{"gate_crossed", "generator_purchased"}},
	}
}

func runAttainmentVector(t *testing.T, bundle CatalogBundle, vector attainmentVector) attainmentVector {
	t.Helper()
	state := foundationScopeState(t, bundle.Economy, economy.ScopeCompany)
	state.AchievementsEarnedRun = map[string]bool{}
	state.AchievementsAttainedRun = map[string]bool{}
	for _, id := range vector.AttainedBefore {
		definition, _ := bundle.Achievements.Definition(id)
		state.AchievementsAttainedRun[id] = true
		state.AttainmentScoreRun += definition.ScoreGrant
	}
	earned := map[string]bool{}
	for _, id := range vector.EarnedNow {
		earned[id] = true
		state.AchievementsEarnedRun[id] = true
	}
	generators := map[string]int64{}
	for id, count := range vector.Generators {
		generators[id] = count
	}
	run := achievements.Observation{Facts: map[string]bool{}, Counters: vector.Counters, Generators: generators}
	events := make([]save.EventWrite, 0, len(vector.EventKinds))
	for _, kind := range vector.EventKinds {
		events = append(events, save.EventWrite{Kind: save.EventKind(kind)})
	}
	debits := map[string]string{}
	if vector.CashDebit != nil {
		debits["company.cash"] = *vector.CashDebit
	}
	proofEvents := append([]save.EventWrite(nil), events...)
	runID := map[string]any{"company_stream_id": "01986666-0000-7000-8000-000000000002", "run_seq": 2}
	if err := attainRun(bundle, state, run, earned, func(definition achievements.Definition) bool {
		return achievementProofSatisfied(definition, debits, proofEvents)
	}, runID, "01986666-0000-7000-8000-000000000003", &events); err != nil {
		t.Fatalf("%s: %v", vector.Name, err)
	}
	result := vector
	result.AttainedAfter = []string{}
	for id := range state.AchievementsAttainedRun {
		result.AttainedAfter = append(result.AttainedAfter, id)
	}
	sort.Strings(result.AttainedAfter)
	result.ScoreAfter = state.AttainmentScoreRun
	result.ReattainedOrder = []string{}
	for _, event := range events[len(vector.EventKinds):] {
		if event.Kind != save.EventAchievementReattained {
			t.Fatalf("%s: unexpected event %s", vector.Name, event.Kind)
		}
		var payload struct {
			AchievementID string `json:"achievement_id"`
		}
		_ = json.Unmarshal(event.Payload, &payload)
		result.ReattainedOrder = append(result.ReattainedOrder, payload.AchievementID)
	}
	return result
}

// TestAttainmentVectors pins the Go-authored CV2 corpus the TS runtime
// replays byte-for-byte (AC5). Regenerate with UPDATE_AXIS_VECTORS=1.
func TestAttainmentVectors(t *testing.T) {
	bundle := axisContentBundle(t)
	results := make([]attainmentVector, 0)
	for _, vector := range attainmentCases() {
		results = append(results, runAttainmentVector(t, bundle, vector))
	}
	encoded, err := json.MarshalIndent(map[string]any{"schema_version": 1, "economy": "balance/testdata/axis-stack/economy-v5-fixture.json", "vectors": results}, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	encoded = append(encoded, '\n')
	if os.Getenv("UPDATE_AXIS_VECTORS") == "1" {
		if err := os.WriteFile(attainmentVectorsPath, encoded, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	committed, err := os.ReadFile(attainmentVectorsPath)
	if err != nil || !bytes.Equal(committed, encoded) {
		t.Fatalf("attainment vectors drifted (err=%v); regenerate with UPDATE_AXIS_VECTORS=1 and review", err)
	}
	byName := map[string]attainmentVector{}
	for _, result := range results {
		byName[result.Name] = result
	}
	// Semantic pins that do not depend on the golden file.
	if got := byName["veteran_reattains_first_purchase"]; !reflect.DeepEqual(got.ReattainedOrder, []string{"achievement.generators_purchased_1"}) || got.ScoreAfter != 2 {
		t.Fatalf("veteran = %+v", got)
	}
	if got := byName["newly_earned_attains_without_reattained_event"]; len(got.ReattainedOrder) != 0 || got.ScoreAfter != 2 {
		t.Fatalf("newly earned = %+v", got)
	}
	if got := byName["latched_id_grants_once"]; !reflect.DeepEqual(got.ReattainedOrder, []string{"achievement.generators_purchased_25"}) {
		t.Fatalf("latched = %+v", got)
	}
	if got := byName["career_definitions_never_attain"]; len(got.AttainedAfter) != 0 {
		t.Fatalf("career = %+v", got)
	}
	if got := byName["burn_without_debit_does_not_attain"]; !reflect.DeepEqual(got.AttainedAfter, []string{"achievement.first_gate"}) {
		t.Fatalf("burn without debit must attain only the provenance row = %+v", got)
	}
	if got := byName["burn_with_debit_attains"]; !reflect.DeepEqual(got.AttainedAfter, []string{"achievement.first_gate", "achievement.gate_burn_t3"}) {
		t.Fatalf("burn with debit = %+v", got)
	}
}

const axisFormulaVectorsPath = "../../testdata/axis-stack/formula-vectors-v1.json"

type axisFormulaVector struct {
	Name               string   `json:"name"`
	Input              string   `json:"input"`
	AttainmentScoreRun int64    `json:"attainment_score_run"`
	AchievementScore   int64    `json:"achievement_score_run"`
	Owned              []string `json:"owned"`
	ExtraInternPPM     int64    `json:"extra_intern_ppm"`
	X                  int64    `json:"x"`
	Saturated          bool     `json:"saturated"`
	Contributions      []string `json:"contributions"`
	Product            string   `json:"product"`
}

// axisVectorCatalog is the fixture economy, optionally with a third
// test-only intern (pr_intern_9) so the three-factor product is covered.
func axisVectorCatalog(t *testing.T, input string, extraPPM int64) *economy.Catalog {
	t.Helper()
	data, err := os.ReadFile("../../balance/testdata/axis-stack/economy-v5-fixture.json")
	if err != nil {
		t.Fatal(err)
	}
	var root map[string]any
	if err := json.Unmarshal(data, &root); err != nil {
		t.Fatal(err)
	}
	root["axis_stack"].(map[string]any)["input"] = input
	if extraPPM > 0 {
		root["upgrades"] = append(root["upgrades"].([]any), map[string]any{"id": "upgrade.pr_intern_9", "cost": map[string]any{"resource": "company.cash", "amount": "1e9"},
			"window": map[string]any{"from_gate": "gate.t0_to_t1", "to_gate": nil}, "requires": []any{map[string]any{"kind": "axis_at_least", "minimum": 1}},
			"effects": []any{map[string]any{"source_id": "upgrade.pr_intern_9.axis", "slot": "axis_stack", "target": "all", "factor_ppm": extraPPM}}, "roles": []any{"synergy_feed"}, "copy_key": "upgrade.pr_intern_9"})
		root["multiplier_sources"] = append(root["multiplier_sources"].([]any), map[string]any{"id": "upgrade.pr_intern_9.axis", "slot": "axis_stack", "target": "all", "provider": "axis_stack"})
	}
	encoded, _ := json.Marshal(root)
	catalog, err := economy.LoadCatalog(encoded)
	if err != nil {
		t.Fatal(err)
	}
	return catalog
}

func axisFormulaCases() []axisFormulaVector {
	both := []string{"upgrade.pr_intern_1", "upgrade.pr_intern_2"}
	return []axisFormulaVector{
		{Name: "x_zero_is_neutral", Input: "achievement_attainment_run", Owned: []string{"upgrade.pr_intern_1"}},
		{Name: "x_at_threshold", Input: "achievement_attainment_run", AttainmentScoreRun: 6, Owned: []string{"upgrade.pr_intern_1"}},
		{Name: "x_eight_one_intern", Input: "achievement_attainment_run", AttainmentScoreRun: 8, Owned: []string{"upgrade.pr_intern_1"}},
		{Name: "x_at_cap_two_interns", Input: "achievement_attainment_run", AttainmentScoreRun: 44, Owned: both},
		{Name: "x_out_of_bounds_clamps", Input: "achievement_attainment_run", AttainmentScoreRun: 50, Owned: both},
		{Name: "three_interns_multiply_in_byte_order", Input: "achievement_attainment_run", AttainmentScoreRun: 12, Owned: append(append([]string{}, both...), "upgrade.pr_intern_9"), ExtraInternPPM: 15_000},
		{Name: "score_run_contrast_arm_reads_score", Input: "achievement_score_run", AttainmentScoreRun: 0, AchievementScore: 10, Owned: both},
		{Name: "unowned_interns_contribute_nothing", Input: "achievement_attainment_run", AttainmentScoreRun: 12},
	}
}

// TestAxisFormulaVectors pins the Go-authored CV3 vectors the TS runtime must
// reproduce byte-for-byte (AC2). Regenerate with UPDATE_AXIS_VECTORS=1.
func TestAxisFormulaVectors(t *testing.T) {
	results := make([]axisFormulaVector, 0)
	for _, vector := range axisFormulaCases() {
		catalog := axisVectorCatalog(t, vector.Input, vector.ExtraInternPPM)
		state := foundationScopeState(t, catalog, economy.ScopeCompany)
		state.GeneratorCounts = map[string]int64{}
		for _, generator := range catalog.GeneratorClassesForScope(economy.ScopeCompany) {
			state.GeneratorCounts[generator.ID] = 0
		}
		state.UpgradesOwned = map[string]bool{}
		for _, id := range vector.Owned {
			state.UpgradesOwned[id] = true
		}
		state.AchievementsAttainedRun, state.AttainmentScoreRun, state.AchievementScoreRun = map[string]bool{}, vector.AttainmentScoreRun, vector.AchievementScore
		x, saturated, err := AxisInput(state, catalog)
		if err != nil {
			t.Fatal(err)
		}
		contributions, err := contentContributions(state, catalog)
		if err != nil {
			t.Fatal(err)
		}
		result := vector
		result.X, result.Saturated, result.Contributions = x, saturated, []string{}
		for _, contribution := range contributions {
			if contribution.Slot == economy.SlotAxisStack {
				result.Contributions = append(result.Contributions, contribution.SourceID+"="+contribution.Factor.String())
			}
		}
		product, err := contributionFactorForTarget(catalog, "all", contributions)
		if err != nil {
			t.Fatal(err)
		}
		result.Product = product.String()
		results = append(results, result)
	}
	encoded, err := json.MarshalIndent(map[string]any{"schema_version": 1, "vectors": results}, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	encoded = append(encoded, '\n')
	if os.Getenv("UPDATE_AXIS_VECTORS") == "1" {
		if err := os.WriteFile(axisFormulaVectorsPath, encoded, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	committed, err := os.ReadFile(axisFormulaVectorsPath)
	if err != nil || !bytes.Equal(committed, encoded) {
		t.Fatalf("axis formula vectors drifted (err=%v); regenerate with UPDATE_AXIS_VECTORS=1 and review", err)
	}
	for _, result := range results {
		switch result.Name {
		case "x_zero_is_neutral":
			if result.Product != "1e0" || !reflect.DeepEqual(result.Contributions, []string{"upgrade.pr_intern_1.axis=1e0"}) {
				t.Fatalf("x=0 = %+v", result)
			}
		case "x_eight_one_intern":
			if result.Product != "1.2e0" {
				t.Fatalf("x=8 = %+v (RFC CV8: x1.20)", result)
			}
		case "x_out_of_bounds_clamps":
			if result.X != 44 || !result.Saturated {
				t.Fatalf("clamp = %+v", result)
			}
		}
	}
}

// TestCarryRuleIsStructural is AC3: byte-identical Company states under
// different Founders (lifetime sets, Clout) produce identical attainment and
// axis rates under the ruled attainment input; the same veteran under the
// shipped achievement_score_run input diverges (DG-3 as a discriminating fact).
func TestCarryRuleIsStructural(t *testing.T) {
	now := time.Date(2026, 9, 25, 15, 0, 0, 0, time.UTC)
	run := func(bundle CatalogBundle, founder *save.State) (*save.State, []save.EventWrite) {
		state := replayFixtureState(t, bundle.Economy, now.Add(-time.Hour))
		meterState, err := meters.NewRunState(bundle.Meters, 0)
		if err != nil {
			t.Fatal(err)
		}
		state.WireVersion = 19
		state.MeterBands = nil
		state.MeterValues, state.MeterDecayRemainders, state.MeterInputRemainders = meterState.Values, meterState.DecayRemainders, meterState.InputRemainders
		state.AchievementsEarnedRun, state.AchievementsAttainedRun = map[string]bool{}, map[string]bool{}
		state.ActiveBuffs = []save.ActiveBuff{}
		state.Tier, state.GeneratorPurchasedTotal = 1, 1
		state.UpgradesOwned["upgrade.pr_intern_1"] = true
		before, err := cloneReplayState(state, bundle.Economy)
		if err != nil {
			t.Fatal(err)
		}
		state.EvaluatedThrough = now
		events := []save.EventWrite{}
		request := IntentRequest{IntentID: "01986666-0600-7000-8000-000000000701"}
		revision := save.Revision{StreamID: "01986666-1600-7000-8000-000000000002", OwnerID: "01986666-2600-7000-8000-000000000002"}
		if err := applyFoundationTransition(bundle, before, state, founder, revision, request, now, nil, map[string]string{}, false, &events); err != nil {
			t.Fatal(err)
		}
		return state, events
	}
	founderWith := func(bundle CatalogBundle, lifetime map[string]bool, clout int64) *save.State {
		founder := foundationScopeState(t, bundle.Economy, economy.ScopeFounder)
		founder.WireVersion = save.LatestSupportedVersion
		founder.AchievementsEarnedLifetime, founder.CloutLifetime = lifetime, clout
		return founder
	}
	veteranSet := map[string]bool{"achievement.first_gate": true, "achievement.generators_purchased_1": true}
	axisProduct := func(bundle CatalogBundle, state *save.State) string {
		contributions, err := contentContributions(state, bundle.Economy)
		if err != nil {
			t.Fatal(err)
		}
		product, err := contributionFactorForTarget(bundle.Economy, "all", contributions)
		if err != nil {
			t.Fatal(err)
		}
		return product.String()
	}

	attainment := axisContentBundle(t)
	fresh, freshEvents := run(attainment, founderWith(attainment, map[string]bool{}, 0))
	veteran, veteranEvents := run(attainment, founderWith(attainment, veteranSet, 999))
	if !reflect.DeepEqual(fresh.AchievementsAttainedRun, veteran.AchievementsAttainedRun) || fresh.AttainmentScoreRun != veteran.AttainmentScoreRun || fresh.AttainmentScoreRun != 4 {
		t.Fatalf("attainment diverged: fresh=%v/%d veteran=%v/%d", fresh.AchievementsAttainedRun, fresh.AttainmentScoreRun, veteran.AchievementsAttainedRun, veteran.AttainmentScoreRun)
	}
	if a, b := axisProduct(attainment, fresh), axisProduct(attainment, veteran); a != b {
		t.Fatalf("axis product diverged under attainment: fresh=%s veteran=%s", a, b)
	}
	kinds := func(events []save.EventWrite) []string {
		result := []string{}
		for _, event := range events {
			result = append(result, string(event.Kind))
		}
		return result
	}
	if got := kinds(veteranEvents); !reflect.DeepEqual(got, []string{"achievement_reattained.v1", "achievement_reattained.v1"}) {
		t.Fatalf("veteran events = %v", got)
	}
	if got := kinds(freshEvents); !reflect.DeepEqual(got, []string{"achievement_earned.v1", "achievement_earned.v1"}) {
		t.Fatalf("fresh events = %v", got)
	}

	// The failing case: the shipped score input shrinks for the veteran.
	score := axisContentBundleWithInput(t, "achievement_score_run")
	freshScore, _ := run(score, founderWith(score, map[string]bool{}, 0))
	veteranScore, _ := run(score, founderWith(score, veteranSet, 999))
	if axisProduct(score, freshScore) == axisProduct(score, veteranScore) {
		t.Fatal("achievement_score_run did not diverge for a veteran; AC3's discriminating case is vacuous")
	}
}

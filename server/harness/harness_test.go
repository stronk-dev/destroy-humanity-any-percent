package harness

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"cloud-clicker/server/economy"
	"cloud-clicker/server/epochseed"
	"cloud-clicker/server/replaycatalog"
	"cloud-clicker/server/save"
)

func TestRoleActivationCountsAreCanonicalAndComplete(t *testing.T) {
	counts := map[string]RoleActivationCount{
		"generator.low\x00stock_rate\x00faction.stock": {GeneratorID: "generator.low", Kind: economy.RoleStockRate, TargetID: "faction.stock", Count: 3},
		"generator.high\x00provision\x00generator.low": {GeneratorID: "generator.high", Kind: economy.RoleProvision, TargetID: "generator.low", Count: 2},
	}
	got := sortedRoleActivations(counts)
	if len(got) != 2 || got[0].GeneratorID != "generator.high" || got[0].Count != 2 || got[1].GeneratorID != "generator.low" || got[1].Count != 3 {
		t.Fatalf("role activation counts=%+v", got)
	}
}

func TestSuiteConstantsHashUsesEpochManifest(t *testing.T) {
	root := filepath.Join("..", "..")
	suite, err := LoadSuite(root, "testdata/harness/scenarios/phase0-production.json")
	if err != nil {
		t.Fatal(err)
	}
	bundle, err := epochseed.Load(root)
	if err != nil {
		t.Fatal(err)
	}
	if suite.ConstantsHash != bundle.Hash || len(bundle.Artifacts) != len(bundle.Seed.Artifacts) {
		t.Fatalf("suite=%s manifest=%s artifacts=%d declarations=%d", suite.ConstantsHash, bundle.Hash, len(bundle.Artifacts), len(bundle.Seed.Artifacts))
	}
}

func TestSplitMix64AndBoundArePinned(t *testing.T) {
	random := NewSplitMix64(0)
	if got := random.Next(); got != 0xe220a8397b1dcdaf {
		t.Fatalf("first draw = %016x", got)
	}
	if got := random.Next(); got != 0x6e789e6aa1b965f4 {
		t.Fatalf("second draw = %016x", got)
	}
	bounded := NewSplitMix64(42)
	for index := 0; index < 10_000; index++ {
		if got := bounded.Bound(7); got >= 7 {
			t.Fatalf("bounded draw = %d", got)
		}
	}
}

func TestDeterministicUUIDV7(t *testing.T) {
	first := NewUUIDStream(7)
	second := NewUUIDStream(7)
	left, err := first.Next(Epoch.UnixMilli())
	if err != nil {
		t.Fatal(err)
	}
	right, err := second.Next(Epoch.UnixMilli())
	if err != nil || left != right || left[14] != '7' || left[19] < '8' || left[19] > 'b' {
		t.Fatalf("uuid left=%s right=%s err=%v", left, right, err)
	}
}

func TestSmallRunIsByteDeterministic(t *testing.T) {
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	suite, err := LoadSuite(root, "testdata/harness/scenarios/phase0-production.json")
	if err != nil {
		t.Fatal(err)
	}
	suite.Scenario.Milestones = suite.Scenario.Milestones[:3]
	spec := RunSpec{PolicyID: "casual.phase0", PolicyVersion: 1, SeedStart: "0", SeedCount: 1, HorizonMS: 300_000}
	first := suite.run(spec, 0)
	second := suite.run(spec, 0)
	left, _ := CanonicalJSON(first)
	right, _ := CanonicalJSON(second)
	if string(left) != string(right) || first.Outcome != "completed" {
		t.Fatalf("determinism/outcome mismatch\n%s\n%s", left, right)
	}
}

func TestRunTaskDispatchExecutesEveryCompleteKeyExactlyOnce(t *testing.T) {
	const workerCount = 4
	const taskCount = 17
	suite := Suite{Scenario: Scenario{ID: "scenario.dispatch", Version: 3},
		ScenarioHash: "sha256:scenario", ConstantsHash: "sha256:constants"}
	tasks := make([]runTask, taskCount)
	for index := range tasks {
		spec := RunSpec{PolicyID: "policy.dispatch", PolicyVersion: 2, HorizonMS: 1}
		seed := uint64(index + 100)
		tasks[index] = runTask{spec: spec, seed: seed, key: suite.runKey(spec, seed)}
	}

	started := make(chan struct{}, workerCount)
	release := make(chan struct{})
	go func() {
		for range workerCount {
			<-started
		}
		close(release)
	}()
	var arrived atomic.Int64
	counts := make(map[RunKey]int, taskCount)
	var countsMu sync.Mutex
	reports, err := dispatchRunTasks(tasks, workerCount, func(task runTask) RunReport {
		if arrived.Add(1) <= workerCount {
			started <- struct{}{}
			<-release
		}
		countsMu.Lock()
		counts[task.key]++
		countsMu.Unlock()
		return RunReport{Key: task.key}
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(reports) != taskCount || len(counts) != taskCount {
		t.Fatalf("reports=%d distinct executions=%d want=%d", len(reports), len(counts), taskCount)
	}
	for _, task := range tasks {
		if counts[task.key] != 1 {
			t.Fatalf("run key %+v executions=%d want=1", task.key, counts[task.key])
		}
	}
}

func TestRunTaskDispatchRejectsMismatchedResultKey(t *testing.T) {
	key := RunKey{HarnessSchemaVersion: 1, ScenarioID: "scenario.dispatch", ScenarioVersion: 1,
		ScenarioHash: "sha256:scenario", PolicyID: "policy.dispatch", PolicyVersion: 1,
		Seed: "0", ConstantsHash: "sha256:constants"}
	_, err := dispatchRunTasks([]runTask{{key: key}}, 1, func(runTask) RunReport {
		return RunReport{Key: RunKey{}}
	})
	if err == nil || !strings.Contains(err.Error(), "mismatched run key") {
		t.Fatalf("mismatched result key error=%v", err)
	}
}

func TestBaselineDriftThresholdsUseIntegerCrossMultiplication(t *testing.T) {
	baseline := AggregateReport{Values: []AggregateValue{{PolicyID: "casual.phase0", Milestone: "m", Statistic: "p50", ValueMS: 100}}}
	warning := AggregateReport{Values: []AggregateValue{{PolicyID: "casual.phase0", Milestone: "m", Statistic: "p50", ValueMS: 111}}}
	warnings, failures := CompareBaseline(warning, baseline)
	if len(warnings) != 1 || len(failures) != 0 {
		t.Fatalf("11%% drift warnings=%v failures=%v", warnings, failures)
	}
	failing := AggregateReport{Values: []AggregateValue{{PolicyID: "casual.phase0", Milestone: "m", Statistic: "p50", ValueMS: 126}}}
	warnings, failures = CompareBaseline(failing, baseline)
	if len(warnings) != 0 || len(failures) != 1 {
		t.Fatalf("26%% drift warnings=%v failures=%v", warnings, failures)
	}
	_, failures = CompareBaseline(AggregateReport{}, baseline)
	if len(failures) != 1 {
		t.Fatalf("removed baseline key failures=%v", failures)
	}
}

func TestLoadRatifiedFirstContentCandidateSuite(t *testing.T) {
	root := filepath.Join("..", "..")
	suite, identity, err := LoadCandidateSuite(root, "testdata/harness/scenarios/phase0-production.json",
		"planning/first-content-epoch/promotion-manifest.candidate.v1.json")
	if err != nil {
		t.Fatal(err)
	}
	if identity.ConstantsHash != "sha256:1a4463bcf67440ce1ba01e6c6eb850c0614329cac63064ef07725d042c7cf21a" ||
		len(identity.ArtifactNames) != 16 || suite.ConstantsHash != identity.ConstantsHash {
		t.Fatalf("identity=%+v suite hash=%s", identity, suite.ConstantsHash)
	}
	if _, ok := suite.Catalog.Resource("company.permits"); !ok {
		t.Fatal("candidate harness did not load the ratified Economy bytes")
	}
	found := false
	for _, gate := range suite.RoutesCatalog.Gates() {
		found = found || gate.ID == "gate.t3_to_t4"
	}
	if !found {
		t.Fatal("candidate harness did not load the ratified Routes bytes")
	}
}

func TestLoadRatifiedT0T1CandidateManifest(t *testing.T) {
	root := filepath.Join("..", "..")
	manifestPath := "planning/t0-t1-content/promotion-manifest.candidate.v1.json"
	manifest, artifacts, productionPaths, err := loadCandidateManifest(root, manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]struct {
		schema int
		source string
		path   string
		hash   string
		gate   string
	}{
		"achievements":  {1, "balance/testdata/first-content/achievements-v1.json", "balance/achievements/first-content.json", "1a11d6c5a0c044ff8077574bb71f1c893bde93a050e20a91e0d776c7e79f8903", "make test-go GO_PACKAGES='./achievements ./replaycatalog' && make test-client"},
		"categories":    {1, "balance/testdata/t0-t1/categories-v1.json", "balance/categories/phase0.json", "ff63b341ff8a7439e48cbfa7cb91dcf51089809fcbb0e6e54201965e5911b9a5", "make test-go GO_PACKAGES='./leaderboard ./replaycatalog'"},
		"commons":       {1, "balance/commons/phase0.json", "balance/commons/phase0.json", "33d4e73a32e12c973acf9633a1e829fd4da2de0753c6004821fb93ff14208c93", "make verify-schema"},
		"doctrines":     {1, "balance/testdata/first-content/doctrines-v1.json", "balance/doctrines/first-content.json", "a3bca5f7eb07fb3b5bf185ce6191771c044a033b47c6bba390582dd7e1745672", "make test-go GO_PACKAGES='./doctrine ./replaycatalog' && make test-client"},
		"economy":       {4, "balance/testdata/t0-t1/economy-v4.json", "balance/catalogs/phase0.json", "fb75e5cf32f545d9470cc8512a8c63f45ed9edd96c68ba65cfeabe0ce2c7f37d", "make verify-schema && make t0-t1-branch-check"},
		"factions":      {1, "balance/factions/phase0.json", "balance/factions/phase0.json", "e44f461eca6cc6c048edebc42356915e6d4be16f480b4795a1fcc458855005fe", "make verify-schema"},
		"fiscal":        {1, "balance/testdata/first-content/fiscal-v1.json", "balance/fiscal/first-content.json", "3847236f8001ed7e29ab41054fbeef38c5e5ea8b838e478d2c4057fdc417f2a9", "make test-go GO_PACKAGES='./fiscal ./replaycatalog' && make test-client"},
		"guilds":        {1, "balance/guilds/phase0.json", "balance/guilds/phase0.json", "e70e644fd62be3c37e0ae465ea55eb104dfc83f810f2d66f11806328d18366fa", "make verify-schema"},
		"meters":        {1, "balance/testdata/first-content/meters-v1.json", "balance/meters/first-content.json", "320deca9ccbe70c1822f0d2664ea75dfd7627d7f098dfd1243ef432bea7bb485", "make test-go GO_PACKAGES='./meters ./replaycatalog' && make test-client"},
		"minigame_api":  {1, "balance/testdata/minigame-api-candidate-v1.json", "balance/minigame-api/first-content.json", "b16b5e0eb6f9426c8b1b94255e2d8e04f53f78b391fdbbb348ad7438d7bab31c", "make test-go GO_PACKAGES='./minigameapi ./replaycatalog' && make test-client"},
		"minigames":     {3, "testdata/minigame/pitch-v3.json", "balance/minigames/first-content.json", "f08fd3ab1959da66f389ef918b936f81d8a2562762055e7b27f4f9e771ff0862", "make test-go GO_PACKAGES='./minigame ./replaycatalog' && make test-client"},
		"opportunities": {1, "balance/testdata/t0-t1/opportunities-v1.json", "balance/opportunities/t0-t1.json", "d2aab242f2e5b9a73c11b32981fdb3971b4cbf6d96df3fbf0277abfc1ebc7974", "make test-go GO_PACKAGES='./activeplay ./replaycatalog'"},
		"pets":          {2, "balance/testdata/first-content/pets-v2.json", "balance/pets/first-content.json", "5c1f27006871ddbd688cdb36e673a64ef5080c92950d22df486576dfae4aa1c1", "make test-go GO_PACKAGES='./pet ./replaycatalog' && make test-client"},
		"pitch":         {1, "balance/testdata/pitch-v1.json", "balance/pitch.json", "bd4218199c5ef00eaa2851020f6d77fcf826a30eee1d399a371a711b9b0ee10f", "make test-go GO_PACKAGES='./pitch ./replaycatalog' && make test-client"},
		"prestige":      {1, "balance/prestige/phase0.json", "balance/prestige/phase0.json", "1873090781bed666c8f989169a9e59990547b1f713ac2f9a8215f51d3f0ea7ec", "make verify-schema"},
		"relevance":     {2, "balance/testdata/t0-t1/relevance-policy-t1-v2.json", "balance/relevance/t0-t1.json", "f513360cc421e9b5a4ca624c977fdde055104b52052952e28a4aa5d5443554ef", "make t0-t1-relevance-all && make test-go GO_PACKAGES='./relevancepolicy ./replaycatalog'"},
		"routes":        {1, "balance/testdata/t0-t1/routes-v1.json", "balance/routes/phase0.json", "a84cce06ae67a68817174b99cfe7191e3c2f9bf47c1c20b4ebab1704baf99cfa", "make verify-schema && make test-go GO_PACKAGES='./routes ./replaycatalog'"},
		"soul":          {1, "balance/testdata/first-content/soul-v1.json", "balance/soul/first-content.json", "a57798f94892a86fd6ea727b76d5bfa663db27c4abd10180204c26ea83587de4", "make test-go GO_PACKAGES='./soul ./replaycatalog' && make test-client"},
	}
	if len(manifest.Artifacts) != len(want) || len(artifacts) != len(want) || len(productionPaths) != len(want) {
		t.Fatalf("epoch-7 candidate artifacts=%d loaded=%d paths=%d want=%d", len(manifest.Artifacts), len(artifacts), len(productionPaths), len(want))
	}
	for _, row := range manifest.Artifacts {
		expected, ok := want[row.Name]
		if !ok || row.SchemaVersion != expected.schema || row.SourcePath != expected.source || row.ProductionPath != expected.path ||
			row.SHA256 != expected.hash || row.ContentGate != expected.gate {
			t.Fatalf("unexpected epoch-7 row %+v", row)
		}
	}
	hash, err := save.ConstantsHashArtifacts(artifacts)
	if err != nil || hash != manifest.ConstantsHash || hash != "sha256:6c7fab29c24fae68e3067c883177bc78fe61b9d91704b6d936b3e4f3cfd8f789" {
		t.Fatalf("epoch-7 constants hash=%s manifest=%s err=%v", hash, manifest.ConstantsHash, err)
	}
	bundle, err := replaycatalog.Load(manifest.ConstantsHash, artifacts)
	if err != nil || bundle.Opportunities == nil || bundle.Relevance == nil {
		t.Fatalf("epoch-7 replay bundle opportunities=%v relevance=%v err=%v", bundle.Opportunities != nil, bundle.Relevance != nil, err)
	}
}

func TestCandidatePacingReportSeparatesFindingsFromInvariants(t *testing.T) {
	baseline := AggregateReport{ConstantsHash: "sha256:baseline", Values: []AggregateValue{{PolicyID: "p", Milestone: "m", Statistic: "p50", ValueMS: 100}}}
	current := AggregateReport{ScenarioID: "scenario", ScenarioHash: "sha256:scenario", ConstantsHash: "sha256:candidate", RunCount: 1,
		Values:   []AggregateValue{{PolicyID: "p", Milestone: "m", Statistic: "p50", ValueMS: 130}},
		Failures: []string{"envelope p/m/p50=130 outside bounds", "schema=1/policy=p:resource_bounds"}}
	report := BuildCandidatePacingReport(CandidateIdentity{ManifestPath: "manifest.json", ArtifactNames: []string{"economy"}, ConstantsHash: current.ConstantsHash}, current, baseline)
	if len(report.PacingFindings) != 2 || len(report.InvariantFailures) != 1 || len(report.Values) != 1 || report.Values[0].DeltaMS != 30 ||
		report.Values[0].RelativeDeltaPPM == nil || *report.Values[0].RelativeDeltaPPM != 300_000 {
		t.Fatalf("report=%+v", report)
	}
	if report.PacingWarnings == nil || report.PacingFindings == nil || report.InvariantFailures == nil {
		t.Fatalf("candidate report lists must use closed empty arrays: %+v", report)
	}
}

func TestT0ProgressObservationParticipatesInBaselineDrift(t *testing.T) {
	baseline := AggregateReport{Values: []AggregateValue{{PolicyID: "casual.phase0", Milestone: "milestone.t0_progress_1", Statistic: "p50", ValueMS: 337_000}}}
	warning := AggregateReport{Values: []AggregateValue{{PolicyID: "casual.phase0", Milestone: "milestone.t0_progress_1", Statistic: "p50", ValueMS: 371_000}}}
	warnings, failures := CompareBaseline(warning, baseline)
	if len(warnings) != 1 || len(failures) != 0 {
		t.Fatalf("T0 observation warning=%v failures=%v", warnings, failures)
	}
	failing := AggregateReport{Values: []AggregateValue{{PolicyID: "casual.phase0", Milestone: "milestone.t0_progress_1", Statistic: "p50", ValueMS: 422_000}}}
	warnings, failures = CompareBaseline(failing, baseline)
	if len(warnings) != 0 || len(failures) != 1 {
		t.Fatalf("T0 observation warning=%v failures=%v", warnings, failures)
	}
}

func TestBaselineOnlyRewriteFailsChangeGuard(t *testing.T) {
	if err := ValidateBaselineCommit([]string{baselinePath}, nil, "BALANCE-CHANGE: retune"); err == nil {
		t.Fatal("baseline-only rewrite passed")
	}
	inputs := []string{"balance/catalogs/phase0.json"}
	if err := ValidateBaselineCommit([]string{baselinePath}, inputs, "ordinary commit"); err == nil {
		t.Fatal("baseline rewrite without BALANCE-CHANGE subject passed")
	}
	if err := ValidateBaselineCommit([]string{baselinePath, goldenPath}, inputs, "BALANCE-CHANGE: phase0 retune"); err != nil {
		t.Fatalf("valid balance change failed: %v", err)
	}
	if err := ValidateBaselineCommit([]string{baselinePath, "server/code.go"}, inputs, "BALANCE-CHANGE: smuggle"); err == nil {
		t.Fatal("balance label authorized a code change")
	}
	if err := ValidateBaselineCommit([]string{baselinePath, "balance/catalogs/phase0.json"}, inputs, "BALANCE-CHANGE: same commit"); err == nil {
		t.Fatal("same commit input and baseline change passed")
	}
	if err := ValidateBaselineCommit([]string{baselinePath}, []string{"balance/commons/phase0.json"}, "BALANCE-CHANGE: Commons retune"); err != nil {
		t.Fatalf("Commons input was not recognized: %v", err)
	}
	if err := ValidateBaselineCommit([]string{baselinePath}, []string{"balance/routes/phase0.json"}, "BALANCE-CHANGE: Routes retune"); err != nil {
		t.Fatalf("Routes input was not recognized: %v", err)
	}
	if err := ValidateBaselineCommit([]string{baselinePath}, []string{"balance/prestige/phase0.json"}, "BALANCE-CHANGE: Prestige retune"); err != nil {
		t.Fatalf("Prestige input was not recognized: %v", err)
	}
	if err := ValidateBaselineCommit([]string{baselinePath}, []string{"balance/guilds/phase0.json"}, "BALANCE-CHANGE: Guild retune"); err != nil {
		t.Fatalf("Guild input was not recognized: %v", err)
	}
	if err := ValidateBaselineCommit([]string{relevanceGoldenPath}, []string{"balance/testdata/t0-t1/relevance-scenario-v2.json"}, "BALANCE-CHANGE: candidate scenario"); err != nil {
		t.Fatalf("candidate scenario input was not recognized: %v", err)
	}
	if err := ValidateBaselineCommit([]string{baselinePath}, []string{"balance/categories/phase0.json"}, "BALANCE-CHANGE: category retune"); err != nil {
		t.Fatalf("category input was not recognized: %v", err)
	}
	if err := ValidateBaselineCommit([]string{baselinePath, goldenPath}, nil, "CONSTANTS-IDENTITY: repair hash domain"); err != nil {
		t.Fatal(err)
	}
}

func TestConstantsIdentityGuardAllowsOnlyManifestHash(t *testing.T) {
	baseline := AggregateReport{SchemaVersion: 1, ScenarioID: "scenario", ScenarioHash: "sha256:scenario", ConstantsHash: "sha256:old", RunCount: 1,
		Values: []AggregateValue{{PolicyID: "policy", Milestone: "milestone", Statistic: "p50", ValueMS: 10}}}
	golden := GoldenReport{SchemaVersion: 1, Runs: []RunReport{{Key: RunKey{ConstantsHash: "sha256:old"}, Outcome: "completed"}}}
	beforeBaseline, _ := json.Marshal(baseline)
	beforeGolden, _ := json.Marshal(golden)
	want := "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	baseline.ConstantsHash = want
	golden.Runs[0].Key.ConstantsHash = want
	afterBaseline, _ := json.Marshal(baseline)
	afterGolden, _ := json.Marshal(golden)
	if err := validateConstantsIdentityBlobs(beforeBaseline, afterBaseline, beforeGolden, afterGolden, want); err != nil {
		t.Fatal(err)
	}
	baseline.Values[0].ValueMS++
	changedBaseline, _ := json.Marshal(baseline)
	if err := validateConstantsIdentityBlobs(beforeBaseline, changedBaseline, beforeGolden, afterGolden, want); err == nil {
		t.Fatal("identity-only guard accepted pacing drift")
	}
	unknownBaseline := append(afterBaseline[:len(afterBaseline)-1], []byte(`,"unknown":true}`)...)
	if err := validateConstantsIdentityBlobs(beforeBaseline, unknownBaseline, beforeGolden, afterGolden, want); err == nil {
		t.Fatal("identity-only guard accepted an unknown field")
	}
}

func TestCheckedReportsContainNoJSONFloats(t *testing.T) {
	assertTypeHasNoFloat(t, reflect.TypeOf(SuiteReport{}), "SuiteReport")
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	suite, err := LoadSuite(root, "testdata/harness/scenarios/phase0-production.json")
	if err != nil {
		t.Fatal(err)
	}
	suite.Scenario.Milestones = suite.Scenario.Milestones[:3]
	spec := RunSpec{PolicyID: "casual.phase0", PolicyVersion: 1, SeedStart: "0", SeedCount: 1, HorizonMS: 300_000}
	run := suite.run(spec, 0)
	report := SuiteReport{SchemaVersion: 1, Runs: []RunReport{run}, Aggregate: suite.aggregate([]RunReport{run})}
	data, err := json.Marshal(report)
	if err != nil || !json.Valid(data) {
		t.Fatalf("report JSON=%s err=%v", data, err)
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	var decoded any
	if err := decoder.Decode(&decoded); err != nil {
		t.Fatal(err)
	}
	assertJSONNumbersAreIntegers(t, decoded)
}

func TestUnknownMilestoneKindFailsRuntimeValidation(t *testing.T) {
	err := validateMilestones([]Milestone{{ID: "milestone.future", Kind: "future_kind", MustReach: true}})
	if err == nil || !strings.Contains(err.Error(), "unknown milestone kind") {
		t.Fatalf("unknown milestone err=%v", err)
	}
}

func TestObservationMatrixRejectsInvalidReferencesDuplicatesAndMissingCoverage(t *testing.T) {
	runs := []RunSpec{{PolicyID: "casual.phase0"}}
	milestones := []Milestone{{ID: "milestone.first"}}
	complete := []Envelope{
		{PolicyID: "casual.phase0", Milestone: "milestone.first", Statistic: "p50"},
		{PolicyID: "casual.phase0", Milestone: "milestone.first", Statistic: "p95"},
	}
	tests := []struct {
		name      string
		envelopes []Envelope
		contains  string
	}{
		{name: "unknown policy", envelopes: []Envelope{{PolicyID: "unknown", Milestone: "milestone.first", Statistic: "p50"}}, contains: "unknown policy"},
		{name: "unknown milestone", envelopes: []Envelope{{PolicyID: "casual.phase0", Milestone: "unknown", Statistic: "p50"}}, contains: "unknown milestone"},
		{name: "duplicate tuple", envelopes: append(append([]Envelope{}, complete...), complete[0]), contains: "duplicate envelope"},
		{name: "missing p95", envelopes: complete[:1], contains: "missing pacing observation"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := validateObservationMatrix(runs, milestones, test.envelopes); err == nil || !strings.Contains(err.Error(), test.contains) {
				t.Fatalf("error=%v want containing %q", err, test.contains)
			}
		})
	}
	if err := validateObservationMatrix(runs, milestones, complete); err != nil {
		t.Fatalf("complete observation matrix: %v", err)
	}
}

func TestPhase0ObservationMatrixIsCompleteAndOrdered(t *testing.T) {
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	suite, err := LoadSuite(root, "testdata/harness/scenarios/phase0-production.json")
	if err != nil {
		t.Fatal(err)
	}
	expected := make([]string, 0, len(suite.Scenario.Runs)*len(suite.Scenario.Milestones)*2)
	seenPolicies := make(map[string]bool)
	for _, run := range suite.Scenario.Runs {
		if seenPolicies[run.PolicyID] {
			continue
		}
		seenPolicies[run.PolicyID] = true
		for _, milestone := range suite.Scenario.Milestones {
			for _, statistic := range []string{"p50", "p95"} {
				expected = append(expected, run.PolicyID+"/"+milestone.ID+"/"+statistic)
			}
		}
	}
	actual := make([]string, 0, len(suite.Scenario.Envelopes))
	for _, envelope := range suite.Scenario.Envelopes {
		actual = append(actual, envelope.PolicyID+"/"+envelope.Milestone+"/"+envelope.Statistic)
	}
	if !reflect.DeepEqual(actual, expected) || len(actual) != 16 {
		t.Fatalf("observation order/count\nactual=%v\nexpected=%v", actual, expected)
	}
}

func TestAggregateInvariantFailureCarriesCompleteRunKey(t *testing.T) {
	key := RunKey{HarnessSchemaVersion: 1, ScenarioID: "scenario.test", ScenarioVersion: 2,
		ScenarioHash: "sha256:scenario", PolicyID: "policy.test", PolicyVersion: 3,
		Seed: "42", ConstantsHash: "sha256:constants"}
	suite := Suite{Scenario: Scenario{ID: key.ScenarioID}, ScenarioHash: key.ScenarioHash, ConstantsHash: key.ConstantsHash}
	aggregate := suite.aggregate([]RunReport{{Key: key, InvariantFailures: []string{"numeric_domain"}}})
	if len(aggregate.Failures) != 1 {
		t.Fatalf("failures=%v", aggregate.Failures)
	}
	for _, value := range []string{"schema=1", "scenario=scenario.test@2", "scenario_hash=sha256:scenario", "policy=policy.test@3", "seed=42", "constants_hash=sha256:constants", "numeric_domain"} {
		if !strings.Contains(aggregate.Failures[0], value) {
			t.Fatalf("failure %q omits %q", aggregate.Failures[0], value)
		}
	}
}

func assertTypeHasNoFloat(t *testing.T, value reflect.Type, path string) {
	t.Helper()
	switch value.Kind() {
	case reflect.Float32, reflect.Float64:
		t.Fatalf("%s contains %s", path, value)
	case reflect.Pointer, reflect.Slice, reflect.Array:
		assertTypeHasNoFloat(t, value.Elem(), path)
	case reflect.Struct:
		for index := 0; index < value.NumField(); index++ {
			field := value.Field(index)
			assertTypeHasNoFloat(t, field.Type, path+"."+field.Name)
		}
	}
}

func assertJSONNumbersAreIntegers(t *testing.T, value any) {
	t.Helper()
	switch typed := value.(type) {
	case json.Number:
		if strings.ContainsAny(typed.String(), ".eE") {
			t.Fatalf("report contains JSON float %q", typed)
		}
	case []any:
		for _, item := range typed {
			assertJSONNumbersAreIntegers(t, item)
		}
	case map[string]any:
		for _, item := range typed {
			assertJSONNumbersAreIntegers(t, item)
		}
	}
}

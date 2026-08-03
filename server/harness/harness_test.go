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

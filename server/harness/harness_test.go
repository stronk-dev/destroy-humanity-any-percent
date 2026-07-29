package harness

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

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

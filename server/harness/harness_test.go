package harness

import (
	"encoding/json"
	"path/filepath"
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
}

func TestBaselineOnlyRewriteFailsChangeGuard(t *testing.T) {
	if err := ValidateBaselineChange([]string{"M\t" + baselinePath}, "BALANCE-CHANGE: retune"); err == nil {
		t.Fatal("baseline-only rewrite passed")
	}
	valid := []string{"M\t" + baselinePath, "M\tbalance/catalogs/phase0.json"}
	if err := ValidateBaselineChange(valid, "ordinary commit"); err == nil {
		t.Fatal("baseline rewrite without BALANCE-CHANGE subject passed")
	}
	if err := ValidateBaselineChange(valid, "BALANCE-CHANGE: phase0 retune"); err != nil {
		t.Fatalf("valid balance change failed: %v", err)
	}
	if err := ValidateBaselineChange([]string{"A\t" + baselinePath}, "harness: initial baseline"); err != nil {
		t.Fatalf("initial baseline failed: %v", err)
	}
}

func TestCheckedReportsContainNoJSONFloats(t *testing.T) {
	data, err := json.Marshal(AggregateReport{SchemaVersion: 1, RunCount: 1,
		Values: []AggregateValue{{PolicyID: "p", Milestone: "m", Statistic: "p50", ValueMS: 1}}})
	if err != nil || !json.Valid(data) {
		t.Fatalf("report JSON=%s err=%v", data, err)
	}
}

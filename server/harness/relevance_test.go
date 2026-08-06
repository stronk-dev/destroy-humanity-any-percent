package harness

import (
	"reflect"
	"strings"
	"testing"

	"cloud-clicker/server/decimal"
)

func TestRelevanceRunBudgetAndReachedEncoding(t *testing.T) {
	budget, err := ComputeRelevanceRunBudget(2, 1, 3, 4, true)
	if err != nil || budget != 28 {
		t.Fatalf("budget=%d err=%v", budget, err)
	}
	baseline, ablated := int64(100), int64(175)
	both := MakeRelevanceDelta("milestone.test", &baseline, &ablated, 500)
	unreached := MakeRelevanceDelta("milestone.test", &baseline, nil, 500)
	if both.Status != "both_reached" || both.DeltaMS == nil || *both.DeltaMS != 75 ||
		unreached.Status != "ablated_unreached" || unreached.DeltaMS == nil || *unreached.DeltaMS != 400 {
		t.Fatalf("both=%+v unreached=%+v", both, unreached)
	}
	if value, err := ceilDecimalRatio(decimal.New(11, 0), decimal.New(2, 0)); err != nil || value != 6 {
		t.Fatalf("ceil ratio=%d err=%v", value, err)
	}
}

func TestRelevanceFixtureRunsDeterministicallyThroughProduction(t *testing.T) {
	suite, err := LoadRelevanceSuite("../..", "testdata/harness/relevance/scenario-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	first, err := suite.RunRelevance()
	if err != nil {
		t.Fatal(err)
	}
	second, err := suite.RunRelevance()
	if err != nil {
		t.Fatal(err)
	}
	firstBytes, _ := CanonicalJSON(first)
	secondBytes, _ := CanonicalJSON(second)
	if !reflect.DeepEqual(firstBytes, secondBytes) {
		t.Fatal("relevance report is not byte deterministic")
	}
	if first.RunBudget.DeclaredRuns != 12 || first.RunBudget.ExecutedRuns != 12 ||
		first.RunBudget.DeclaredTransitions != first.RunBudget.ExecutedTransitions || len(first.Items) != 3 || len(first.Groups) != 4 {
		t.Fatalf("report cardinality=%+v items=%d groups=%d", first.RunBudget, len(first.Items), len(first.Groups))
	}
	joined := strings.Join(first.Failures, ",")
	if !strings.Contains(joined, "role_floor:generator.alpha") {
		t.Fatalf("fixture did not discriminate roleless generator: %v", first.Failures)
	}
	if err := ValidateRelevanceReport(first); err != nil {
		t.Fatal(err)
	}
	invalid := first
	invalid.Groups = nil
	if err := ValidateRelevanceReport(invalid); err == nil {
		t.Fatal("report accepted a missing row family")
	}
	invalid = first
	invalid.Items = append([]RelevanceItemReport(nil), first.Items...)
	invalid.Items[0], invalid.Items[1] = invalid.Items[1], invalid.Items[0]
	if err := ValidateRelevanceReport(invalid); err == nil {
		t.Fatal("report accepted unsorted item rows")
	}
	invalid = first
	invalid.Items = append([]RelevanceItemReport(nil), first.Items...)
	invalid.Items[0].IndividualDeltas = append([]RelevanceDelta(nil), first.Items[0].IndividualDeltas...)
	invalid.Items[0].IndividualDeltas[0].Status = "both_reached"
	invalid.Items[0].IndividualDeltas[0].AblatedMS = nil
	if err := ValidateRelevanceReport(invalid); err == nil {
		t.Fatal("report accepted an illegal delta null union")
	}
}

func TestRelevanceFailsBeforeDispatchWhenRunBudgetIsTooSmall(t *testing.T) {
	suite, err := LoadRelevanceSuite("../..", "testdata/harness/relevance/scenario-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	suite.Scenario.RelevanceBudgetMaxRuns = 11
	if _, err := suite.RunRelevance(); err == nil || !strings.Contains(err.Error(), "run budget") {
		t.Fatalf("budget error=%v", err)
	}
}

func TestRelevanceRegistryIsFailClosedForActiveCatalogs(t *testing.T) {
	if err := ValidateRelevanceRegistry("../.."); err != nil {
		t.Fatal(err)
	}
}

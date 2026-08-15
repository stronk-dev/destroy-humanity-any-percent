package harness

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFirstHourPolicyV3RatifiedBytes(t *testing.T) {
	repositoryRoot := filepath.Clean(filepath.Join("..", ".."))
	registry, data, hash, err := LoadFirstHourPolicy(repositoryRoot, "balance/testdata/t0-t1/first-hour-policy-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	if hash != RatifiedFirstHourPolicyHash {
		t.Fatalf("policy hash=%s", hash)
	}
	if len(data) == 0 || registry.Version != 3 {
		t.Fatalf("policy version=%d bytes=%d", registry.Version, len(data))
	}
	for _, id := range []string{"chaos.t0_t1", "casual.t0_t1", "reference.greedy"} {
		if _, ok := registry.Policy(id, 1); !ok {
			t.Fatalf("missing policy %s", id)
		}
	}
}

func TestFirstHourPolicyRejectsGrammarDrift(t *testing.T) {
	repositoryRoot := filepath.Clean(filepath.Join("..", ".."))
	path := filepath.Join(repositoryRoot, "balance/testdata/t0-t1/first-hour-policy-v1.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name string
		from string
		to   string
	}{
		{name: "unknown field", from: `"schema_version": 1,`, to: `"schema_version": 1, "ambient_default": true,`},
		{name: "wait at zero", from: `"legal_only_when_production_income_positive"`, to: `"always_legal"`},
		{name: "decision stream collision", from: `"session_jitter": "jitter"`, to: `"session_jitter": "decision"`},
		{name: "exit threshold", from: `"attended_ms": 2700000`, to: `"attended_ms": 2699999`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mutated := strings.Replace(string(data), test.from, test.to, 1)
			if mutated == string(data) {
				t.Fatal("mutation did not apply")
			}
			temporary := filepath.Join(t.TempDir(), "policy.json")
			if err := os.WriteFile(temporary, []byte(mutated), 0o600); err != nil {
				t.Fatal(err)
			}
			if _, _, _, err := LoadFirstHourPolicy("/", strings.TrimPrefix(temporary, "/")); err == nil {
				t.Fatal("mutated policy accepted")
			}
		})
	}
}

func TestFirstHourDrawPinned(t *testing.T) {
	tests := []struct {
		domain  string
		policy  string
		seed    uint64
		runSeq  int64
		ordinal int64
		want    uint64
	}{
		{domain: "decision", policy: "chaos.t0_t1", seed: 0, runSeq: 1, ordinal: 0, want: 0x01e0f5ff82497bb6},
		{domain: "decision", policy: "chaos.t0_t1", seed: 63, runSeq: 2, ordinal: 900, want: 0xb7d5e818d42dc65e},
		{domain: "jitter", policy: "casual.t0_t1", seed: 31, runSeq: 0, ordinal: 3, want: 0xa75e555246923c4d},
	}
	for _, test := range tests {
		if got := firstHourDraw(test.domain, test.policy, 1, test.seed, test.runSeq, test.ordinal); got != test.want {
			t.Fatalf("draw %s/%s/%d/%d/%d=%#x want %#x", test.domain, test.policy, test.seed, test.runSeq, test.ordinal, got, test.want)
		}
	}
	if _, err := firstHourBoundedDraw("decision", "chaos.t0_t1", 1, 0, 1, 0, 0); err == nil {
		t.Fatal("zero bound accepted")
	}
}

func TestFirstHourSuiteBindsPolicyIntoScenarioIdentity(t *testing.T) {
	repositoryRoot := filepath.Clean(filepath.Join("..", ".."))
	suite, err := LoadFirstHourSuite(repositoryRoot,
		"balance/testdata/t0-t1/harness-scenario-v1.json",
		"balance/testdata/t0-t1/first-hour-policy-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	if suite.PolicyHash != RatifiedFirstHourPolicyHash || suite.ConstantsHash != "sha256:6c7fab29c24fae68e3067c883177bc78fe61b9d91704b6d936b3e4f3cfd8f789" {
		t.Fatalf("policy=%s constants=%s", suite.PolicyHash, suite.ConstantsHash)
	}
	if suite.ScenarioHash != FirstHourScenarioIdentity(suite.ScenarioBytes, suite.PolicyBytes) {
		t.Fatalf("scenario identity=%s", suite.ScenarioHash)
	}
	if got := firstHourPolicyIDs(suite.Scenario); got != "casual.t0_t1,chaos.t0_t1,reference.greedy" {
		t.Fatalf("policy ids=%s", got)
	}
	if got := suite.GateIDs(); len(got) != 1 || got[0] != "gate.t0_to_t1" {
		t.Fatalf("gate scope=%v", got)
	}
	if _, ok := firstHourMilestoneByID(suite.Scenario, "milestone.first_elective_exit"); !ok {
		t.Fatal("elective exit milestone missing")
	}
}

func TestFirstHourRunnerUsesRatifiedPolicyAndRealTransitions(t *testing.T) {
	repositoryRoot := filepath.Clean(filepath.Join("..", ".."))
	suite, err := LoadFirstHourSuite(repositoryRoot,
		"balance/testdata/t0-t1/harness-scenario-v1.json",
		"balance/testdata/t0-t1/first-hour-policy-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	experiment := FirstHourExperiment{AcquihirePurchasedMinimum: 20, BurnoutPriceFactor: "1e0",
		RouteKnowledgeBonus: 50, SeedCapital: "1e2", GeneratedBeigeTowers: 10}
	result := suite.RunExperiment(suite.Scenario.Runs[0], 0, experiment)
	if result.Key.ScenarioHash != suite.ScenarioHash || result.PolicyHash != RatifiedFirstHourPolicyHash || result.TransitionCount == 0 {
		t.Fatalf("identity/transitions result=%+v", result)
	}
	if result.Ending == nil {
		t.Fatalf("scripted ending missing: %+v", result)
	}
	if result.Ending.Branch != "acquihire" && result.Ending.Branch != "burnout" && result.Ending.Branch != "pivot" {
		t.Fatalf("unknown branch %+v", result.Ending)
	}
	if result.Outcome != "completed" {
		t.Fatalf("run failed: %+v", result)
	}
}

func TestFirstHourReferenceReusesProjectedTimeRanker(t *testing.T) {
	repositoryRoot := filepath.Clean(filepath.Join("..", ".."))
	suite, err := LoadFirstHourSuite(repositoryRoot,
		"balance/testdata/t0-t1/harness-scenario-v1.json",
		"balance/testdata/t0-t1/first-hour-policy-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	experiment := FirstHourExperiment{AcquihirePurchasedMinimum: 200, BurnoutPriceFactor: "2e0",
		RouteKnowledgeBonus: 50, SeedCapital: "1e4", GeneratedBeigeTowers: 10}
	result := suite.RunExperiment(suite.Scenario.Runs[2], 0, experiment)
	if result.Outcome != "completed" || result.Ending == nil {
		t.Fatalf("reference run failed: %+v", result)
	}
	if result.Ending.RunOneGateMS != 356_000 || result.Ending.RunTwoGateMS <= 0 || result.Ending.RunTwoGateMS >= result.Ending.RunOneGateMS {
		t.Fatalf("reference gate relation=%+v", result.Ending)
	}
}

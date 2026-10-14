package save

import (
	"encoding/json"
	"errors"
	"os"
	"testing"
)

type runStartedReputationFieldCase struct {
	Name        string `json:"name"`
	Raw         string `json:"raw"`
	Valid       bool   `json:"valid"`
	ParsedValid bool   `json:"parsed_valid"`
}

func runStartedReputationFieldCases(t *testing.T) []runStartedReputationFieldCase {
	t.Helper()
	data, err := os.ReadFile("../../testdata/reputation/run-started-fields-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Version int                             `json:"version"`
		Cases   []runStartedReputationFieldCase `json:"cases"`
	}
	if err := json.Unmarshal(data, &fixture); err != nil || fixture.Version != 1 || len(fixture.Cases) != 14 {
		t.Fatalf("invalid event-field fixture: %v", err)
	}
	return fixture.Cases
}

func runStartedReputationDecision(raw string) IntentDecision {
	return IntentDecision{Outcome: IntentApplied, Receipt: json.RawMessage(`{"outcome":"applied"}`), Events: []EventWrite{{
		Kind: EventRunStarted, SchemaVersion: 2, IntentID: testIntentID,
		Payload: json.RawMessage(`{"founder_id":"01985555-1111-7111-8111-111111111111","run_id":{"company_stream_id":"01985555-2222-7222-8222-222222222222","run_seq":2},"started_at_ms":1800000000000,"assisted":{"advisor":false,"commons":false},"reputation_tree":` + raw + `}`),
	}}}
}

func TestRunStartedReputationFieldAdmission(t *testing.T) {
	for _, row := range runStartedReputationFieldCases(t) {
		t.Run(row.Name, func(t *testing.T) {
			err := validateIntentDecision(runStartedReputationDecision(row.Raw), testIntentID)
			if row.Valid && err != nil || !row.Valid && !errors.Is(err, ErrInvalidStream) {
				t.Fatalf("valid=%v, admission error=%v for %s", row.Valid, err, row.Raw)
			}
		})
	}
}

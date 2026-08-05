package pet

import (
	"encoding/json"
	"os"
	"testing"
)

type careTransitionFixture struct {
	Catalog json.RawMessage `json:"catalog"`
	Cases   []struct {
		Name     string               `json:"name"`
		State    CareState            `json:"state"`
		Input    CareTransitionInput  `json:"input"`
		Expected CareTransitionResult `json:"expected"`
	} `json:"cases"`
}

func TestCareTransitionSharedVectors(t *testing.T) {
	raw, err := os.ReadFile("../../testdata/pet/care-transition-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture careTransitionFixture
	if err := json.Unmarshal(raw, &fixture); err != nil {
		t.Fatal(err)
	}
	catalog, err := LoadCatalog(fixture.Catalog)
	if err != nil {
		t.Fatal(err)
	}
	for _, testCase := range fixture.Cases {
		t.Run(testCase.Name, func(t *testing.T) {
			actual, err := ApplyCareTransition(testCase.State, catalog, testCase.Input)
			if err != nil {
				t.Fatal(err)
			}
			actualJSON, _ := json.Marshal(actual)
			expectedJSON, _ := json.Marshal(testCase.Expected)
			if string(actualJSON) != string(expectedJSON) {
				t.Fatalf("actual=%s\nexpected=%s", actualJSON, expectedJSON)
			}
		})
	}
}

func TestCareTransitionIsPartitionInvariant(t *testing.T) {
	raw, err := os.ReadFile("../../testdata/pet/care-transition-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture careTransitionFixture
	if err := json.Unmarshal(raw, &fixture); err != nil {
		t.Fatal(err)
	}
	catalog, err := LoadCatalog(fixture.Catalog)
	if err != nil {
		t.Fatal(err)
	}
	state := fixture.Cases[0].State
	whole, err := ApplyCareTransition(state, catalog, CareTransitionInput{ActionID: "unknown", AttendedBeforeMS: 0, AttendedAfterMS: 2500})
	if err != nil {
		t.Fatal(err)
	}
	first, err := ApplyCareTransition(state, catalog, CareTransitionInput{ActionID: "unknown", AttendedBeforeMS: 0, AttendedAfterMS: 1250})
	if err != nil {
		t.Fatal(err)
	}
	second, err := ApplyCareTransition(first.State, catalog, CareTransitionInput{ActionID: "unknown", AttendedBeforeMS: 1250, AttendedAfterMS: 2500})
	if err != nil {
		t.Fatal(err)
	}
	wholeDecay := []any{whole.State.StatsPPM, whole.State.StatDecayRemaindersPPM, whole.State.TrustPPM, whole.State.TrustDecayRemainderPPM}
	splitDecay := []any{second.State.StatsPPM, second.State.StatDecayRemaindersPPM, second.State.TrustPPM, second.State.TrustDecayRemainderPPM}
	if mustJSON(wholeDecay) != mustJSON(splitDecay) {
		t.Fatalf("whole=%s split=%s", mustJSON(wholeDecay), mustJSON(splitDecay))
	}
}

func mustJSON(value any) string {
	encoded, _ := json.Marshal(value)
	return string(encoded)
}

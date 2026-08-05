package pet

import (
	"encoding/json"
	"os"
	"testing"
)

type combatInputFixture struct {
	Version int `json:"version"`
	Cases   []struct {
		Name        string        `json:"name"`
		PetTrustPPM int64         `json:"pet_trust_ppm"`
		Soul        int64         `json:"soul"`
		Valid       bool          `json:"valid"`
		Expected    *CombatInputs `json:"expected"`
	} `json:"cases"`
}

func TestCombatInputsSharedVectors(t *testing.T) {
	data, err := os.ReadFile("../../testdata/pet/combat-inputs-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture combatInputFixture
	if err := json.Unmarshal(data, &fixture); err != nil || fixture.Version != 1 || len(fixture.Cases) == 0 {
		t.Fatalf("fixture version=%d cases=%d err=%v", fixture.Version, len(fixture.Cases), err)
	}
	for _, testCase := range fixture.Cases {
		t.Run(testCase.Name, func(t *testing.T) {
			actual, err := NewCombatInputs(testCase.PetTrustPPM, testCase.Soul)
			if (err == nil) != testCase.Valid {
				t.Fatalf("valid=%v actual=%+v err=%v", testCase.Valid, actual, err)
			}
			if testCase.Valid && (testCase.Expected == nil || actual != *testCase.Expected) {
				t.Fatalf("actual=%+v expected=%+v", actual, testCase.Expected)
			}
		})
	}
}

func TestCombatInputsReadReplayOwnedTrust(t *testing.T) {
	actual, err := CombatInputsForState(CareState{TrustPPM: 850_000}, -12)
	if err != nil || actual != (CombatInputs{PetTrustPPM: 850_000, Soul: -12}) {
		t.Fatalf("actual=%+v err=%v", actual, err)
	}
}

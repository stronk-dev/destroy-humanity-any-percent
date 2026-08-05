package minigame

import (
	"errors"
	"testing"
)

var offlineQualityDeclarations = OfflineQualityDeclarations{
	ScoreFactIDs:           map[string]struct{}{"score.total": {}},
	AutomationDestinations: map[string]struct{}{"automation.compute": {}},
}

const validOfflinePolicy = `{"score_fact":"score.total","grade_curve":[{"score_threshold":100,"grade_ppm":400000},{"score_threshold":500,"grade_ppm":700000},{"score_threshold":1000,"grade_ppm":1000000}],"decay_grid_ms":60000,"decay_ppm_per_grid":25000,"neutral_floor_ppm":400000,"automation_destination":"automation.compute"}`

func TestLoadOfflineQualityPolicyAndSelectGrade(t *testing.T) {
	policy, err := LoadOfflineQualityPolicy([]byte(validOfflinePolicy), offlineQualityDeclarations)
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range []struct{ score, want int64 }{{0, 400000}, {99, 400000}, {100, 400000}, {499, 400000}, {500, 700000}, {999, 700000}, {1000, 1000000}} {
		got, gradeErr := OfflineGradeForScore(policy, item.score)
		if gradeErr != nil || got != item.want {
			t.Fatalf("score %d grade=%d err=%v want=%d", item.score, got, gradeErr, item.want)
		}
	}
	if err := ValidateOfflineQualityState(OfflineQualityState{GradePPM: 700000, LastFounderAttendedMS: 1234, DecayRemainderPPM: 999999}); err != nil {
		t.Fatal(err)
	}
}

func TestLoadOfflineQualityPolicyRejectsNoncanonicalRows(t *testing.T) {
	cases := map[string]string{
		"unknown score":         `{"score_fact":"score.other","grade_curve":[{"score_threshold":0,"grade_ppm":400000}],"decay_grid_ms":60000,"decay_ppm_per_grid":25000,"neutral_floor_ppm":400000,"automation_destination":"automation.compute"}`,
		"unknown destination":   `{"score_fact":"score.total","grade_curve":[{"score_threshold":0,"grade_ppm":400000}],"decay_grid_ms":60000,"decay_ppm_per_grid":25000,"neutral_floor_ppm":400000,"automation_destination":"automation.other"}`,
		"extra curve key":       `{"score_fact":"score.total","grade_curve":[{"score_threshold":0,"grade_ppm":400000,"label":"low"}],"decay_grid_ms":60000,"decay_ppm_per_grid":25000,"neutral_floor_ppm":400000,"automation_destination":"automation.compute"}`,
		"duplicate nested":      `{"score_fact":"score.total","grade_curve":[{"score_threshold":0,"score_threshold":1,"grade_ppm":400000}],"decay_grid_ms":60000,"decay_ppm_per_grid":25000,"neutral_floor_ppm":400000,"automation_destination":"automation.compute"}`,
		"descending thresholds": `{"score_fact":"score.total","grade_curve":[{"score_threshold":500,"grade_ppm":400000},{"score_threshold":100,"grade_ppm":700000}],"decay_grid_ms":60000,"decay_ppm_per_grid":25000,"neutral_floor_ppm":400000,"automation_destination":"automation.compute"}`,
		"descending grades":     `{"score_fact":"score.total","grade_curve":[{"score_threshold":0,"grade_ppm":700000},{"score_threshold":100,"grade_ppm":600000}],"decay_grid_ms":60000,"decay_ppm_per_grid":25000,"neutral_floor_ppm":700000,"automation_destination":"automation.compute"}`,
		"floor mismatch":        `{"score_fact":"score.total","grade_curve":[{"score_threshold":100,"grade_ppm":500000}],"decay_grid_ms":60000,"decay_ppm_per_grid":25000,"neutral_floor_ppm":400000,"automation_destination":"automation.compute"}`,
		"empty curve":           `{"score_fact":"score.total","grade_curve":[],"decay_grid_ms":60000,"decay_ppm_per_grid":25000,"neutral_floor_ppm":400000,"automation_destination":"automation.compute"}`,
		"extra root":            `{"score_fact":"score.total","grade_curve":[{"score_threshold":0,"grade_ppm":400000}],"decay_grid_ms":60000,"decay_ppm_per_grid":25000,"neutral_floor_ppm":400000,"automation_destination":"automation.compute","timezone":"utc"}`,
	}
	for name, raw := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := LoadOfflineQualityPolicy([]byte(raw), offlineQualityDeclarations); !errors.Is(err, ErrInvalidOfflineQualityPolicy) {
				t.Fatalf("err=%v", err)
			}
		})
	}
	for _, state := range []OfflineQualityState{{GradePPM: -1}, {GradePPM: 1000001}, {LastFounderAttendedMS: -1}, {DecayRemainderPPM: 1000000}} {
		if err := ValidateOfflineQualityState(state); !errors.Is(err, ErrInvalidOfflineQualityPolicy) {
			t.Fatalf("state %+v err=%v", state, err)
		}
	}
}

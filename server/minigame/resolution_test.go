package minigame

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"
)

func TestApplyFounderResolutionUsesCertifiedDeltaAndAttendedGrid(t *testing.T) {
	data, err := os.ReadFile("../../testdata/minigame/catalog-v2.json")
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := LoadCatalog(data)
	definition, ok := catalog.Definition("fixture.counter")
	if err != nil || !ok {
		t.Fatal(err)
	}
	delta := int64(-2_000)
	result := &Result{Outcome: "complete", ScoreFacts: []ScoreFact{{Kind: "score.total", Value: 100}}, RatingDelta: &delta}
	transition, err := ApplyFounderResolution(RatingState{Elo: 1000, SeasonMember: "ranked", GamesCounted: 2},
		OfflineQualityState{GradePPM: 800_000, LastFounderAttendedMS: 0, DecayRemainderPPM: 1}, result, definition, 60_001)
	if err != nil {
		t.Fatal(err)
	}
	if transition.RatingAfter.Elo != 0 || transition.RatingAfter.GamesCounted != 3 ||
		transition.QualityAfter.GradePPM != 750_000 || transition.QualityAfter.LastFounderAttendedMS != 60_001 ||
		transition.QualityAfter.DecayRemainderPPM != 0 {
		t.Fatalf("unexpected transition: %#v", transition)
	}
}

func TestApplyFounderResolutionUnratedStillUpdatesQuality(t *testing.T) {
	data, _ := os.ReadFile("../../testdata/minigame/catalog-v2.json")
	catalog, _ := LoadCatalog(data)
	definition, _ := catalog.Definition("fixture.counter")
	result := &Result{Outcome: "complete", ScoreFacts: []ScoreFact{{Kind: "score.total", Value: 0}}}
	before := RatingState{Elo: 1000, SeasonMember: "ranked", GamesCounted: 2}
	transition, err := ApplyFounderResolution(before, OfflineQualityState{GradePPM: 750_000}, result, definition, 30_000)
	if err != nil || transition.RatingAfter != before || transition.QualityAfter.GradePPM != 500_000 {
		t.Fatalf("unexpected unrated transition: %#v %v", transition, err)
	}
}

func TestApplyFounderResolutionSharedBoundaryVectors(t *testing.T) {
	catalogData, err := os.ReadFile("../../testdata/minigame/catalog-v2.json")
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := LoadCatalog(catalogData)
	definition, ok := catalog.Definition("fixture.counter")
	if err != nil || !ok {
		t.Fatal(err)
	}
	data, err := os.ReadFile("../../testdata/minigame/resolution-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Version int `json:"version"`
		Cases   []struct {
			Name              string              `json:"name"`
			Rating            RatingState         `json:"rating"`
			Quality           OfflineQualityState `json:"quality"`
			Result            Result              `json:"result"`
			FounderAttendedMS int64               `json:"founder_attended_ms"`
			ExpectedRating    RatingState         `json:"expected_rating"`
			ExpectedQuality   OfflineQualityState `json:"expected_quality"`
		} `json:"cases"`
	}
	if json.Unmarshal(data, &fixture) != nil || fixture.Version != 1 || len(fixture.Cases) != 6 {
		t.Fatal("invalid resolution fixture")
	}
	for _, row := range fixture.Cases {
		t.Run(row.Name, func(t *testing.T) {
			transition, applyErr := ApplyFounderResolution(row.Rating, row.Quality, &row.Result, definition, row.FounderAttendedMS)
			if applyErr != nil || !reflect.DeepEqual(transition.RatingAfter, row.ExpectedRating) ||
				!reflect.DeepEqual(transition.QualityAfter, row.ExpectedQuality) {
				t.Fatalf("unexpected transition: %#v %v", transition, applyErr)
			}
		})
	}
}

package pet

import (
	"errors"
	"os"
	"testing"
)

func TestLoadCatalogGrammarSharedFixture(t *testing.T) {
	data, err := os.ReadFile("../../testdata/pet/catalog-grammar-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	grammar, err := LoadCatalogGrammar(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(grammar.MoodThresholds) != 4 || grammar.MoodThresholds[2].MoodMember != MoodNeutral || grammar.MoodThresholds[2].FloorPPM != 500000 {
		t.Fatalf("mood thresholds=%+v", grammar.MoodThresholds)
	}
	if len(grammar.BehaviorCandidates) != 3 || grammar.BehaviorCandidates[1].ToState != BehaviorActive || grammar.BehaviorCandidates[1].DurationGridTicks != 2 {
		t.Fatalf("behavior candidates=%+v", grammar.BehaviorCandidates)
	}
}

func TestLoadCatalogGrammarRejectsAmbiguousRows(t *testing.T) {
	cases := map[string]string{
		"missing mood":         `{"mood_thresholds":[{"mood_member":"withdrawn","floor_ppm":0}],"behavior_candidates":[]}`,
		"duplicate mood":       `{"mood_thresholds":[{"mood_member":"withdrawn","floor_ppm":0},{"mood_member":"withdrawn","floor_ppm":1},{"mood_member":"neutral","floor_ppm":2},{"mood_member":"engaged","floor_ppm":3}],"behavior_candidates":[]}`,
		"descending floor":     `{"mood_thresholds":[{"mood_member":"withdrawn","floor_ppm":0},{"mood_member":"restless","floor_ppm":2},{"mood_member":"neutral","floor_ppm":1},{"mood_member":"engaged","floor_ppm":3}],"behavior_candidates":[]}`,
		"unknown mood":         `{"mood_thresholds":[{"mood_member":"gone","floor_ppm":0},{"mood_member":"restless","floor_ppm":1},{"mood_member":"neutral","floor_ppm":2},{"mood_member":"engaged","floor_ppm":3}],"behavior_candidates":[]}`,
		"extra mood key":       `{"mood_thresholds":[{"mood_member":"withdrawn","floor_ppm":0,"label":"x"},{"mood_member":"restless","floor_ppm":1},{"mood_member":"neutral","floor_ppm":2},{"mood_member":"engaged","floor_ppm":3}],"behavior_candidates":[]}`,
		"unknown state":        `{"mood_thresholds":[{"mood_member":"withdrawn","floor_ppm":0},{"mood_member":"restless","floor_ppm":1},{"mood_member":"neutral","floor_ppm":2},{"mood_member":"engaged","floor_ppm":3}],"behavior_candidates":[{"from_state":"gone","event":"grid_tick","to_state":"idle","duration_grid_ticks":1}]}`,
		"unknown event":        `{"mood_thresholds":[{"mood_member":"withdrawn","floor_ppm":0},{"mood_member":"restless","floor_ppm":1},{"mood_member":"neutral","floor_ppm":2},{"mood_member":"engaged","floor_ppm":3}],"behavior_candidates":[{"from_state":"idle","event":"offline_tick","to_state":"active","duration_grid_ticks":1}]}`,
		"zero duration":        `{"mood_thresholds":[{"mood_member":"withdrawn","floor_ppm":0},{"mood_member":"restless","floor_ppm":1},{"mood_member":"neutral","floor_ppm":2},{"mood_member":"engaged","floor_ppm":3}],"behavior_candidates":[{"from_state":"idle","event":"grid_tick","to_state":"active","duration_grid_ticks":0}]}`,
		"duplicate candidate":  `{"mood_thresholds":[{"mood_member":"withdrawn","floor_ppm":0},{"mood_member":"restless","floor_ppm":1},{"mood_member":"neutral","floor_ppm":2},{"mood_member":"engaged","floor_ppm":3}],"behavior_candidates":[{"from_state":"idle","event":"grid_tick","to_state":"active","duration_grid_ticks":1},{"from_state":"idle","event":"grid_tick","to_state":"active","duration_grid_ticks":2}]}`,
		"nested duplicate key": `{"mood_thresholds":[{"mood_member":"withdrawn","floor_ppm":0,"floor_ppm":1},{"mood_member":"restless","floor_ppm":2},{"mood_member":"neutral","floor_ppm":3},{"mood_member":"engaged","floor_ppm":4}],"behavior_candidates":[]}`,
	}
	for name, raw := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := LoadCatalogGrammar([]byte(raw)); !errors.Is(err, ErrInvalidCatalogGrammar) {
				t.Fatalf("err=%v", err)
			}
		})
	}
}

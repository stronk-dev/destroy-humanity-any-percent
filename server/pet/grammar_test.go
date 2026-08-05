package pet

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"
)

type grammarFixture struct {
	Version              int                   `json:"version"`
	StatIDs              []StatID              `json:"stat_ids"`
	StatusBands          []StatusBand          `json:"status_bands"`
	Moods                []Mood                `json:"moods"`
	BehaviorStates       []BehaviorState       `json:"behavior_states"`
	BehaviorEvents       []BehaviorEvent       `json:"behavior_events"`
	CareRejectionDetails []CareRejectionDetail `json:"care_rejection_details"`
	BehaviorQueueHardcap int                   `json:"behavior_queue_hardcap"`
	BehaviorPRNGLabel    string                `json:"behavior_prng_label"`
	ValidQueueLengths    []int                 `json:"valid_queue_lengths"`
	InvalidQueueLengths  []int                 `json:"invalid_queue_lengths"`
}

func TestGrammarMatchesSharedFixture(t *testing.T) {
	data, err := os.ReadFile("../../testdata/pet/grammar-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture grammarFixture
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.Version != 1 || !reflect.DeepEqual(fixture.StatIDs, StatIDs()) ||
		!reflect.DeepEqual(fixture.StatusBands, StatusBands()) || !reflect.DeepEqual(fixture.Moods, Moods()) ||
		!reflect.DeepEqual(fixture.BehaviorStates, BehaviorStates()) || !reflect.DeepEqual(fixture.BehaviorEvents, BehaviorEvents()) ||
		!reflect.DeepEqual(fixture.CareRejectionDetails, CareRejectionDetails()) ||
		fixture.BehaviorQueueHardcap != BehaviorQueueHardcap || fixture.BehaviorPRNGLabel != BehaviorPRNGLabel {
		t.Fatalf("pet grammar drift: %+v", fixture)
	}
	for _, length := range fixture.ValidQueueLengths {
		if !ValidBehaviorQueueLength(length) {
			t.Fatalf("valid queue length rejected: %d", length)
		}
	}
	for _, length := range fixture.InvalidQueueLengths {
		if ValidBehaviorQueueLength(length) {
			t.Fatalf("invalid queue length accepted: %d", length)
		}
	}
}

func TestGrammarRejectsUnknownMembers(t *testing.T) {
	if ValidStatID("health") || ValidStatusBand("critical") || ValidMood("happy") ||
		ValidBehaviorState("gone") || ValidBehaviorEvent("offline_tick") || ValidCareRejectionDetail("cost") {
		t.Fatal("invented pet grammar member accepted")
	}
}

func TestGrammarAccessorsDoNotExposeMutableAuthority(t *testing.T) {
	stats := StatIDs()
	stats[0] = "invented"
	if got := StatIDs()[0]; got != StatHunger {
		t.Fatalf("stat authority mutated through accessor: %q", got)
	}
	details := CareRejectionDetails()
	details[len(details)-1] = "invented"
	if got := CareRejectionDetails()[len(details)-1]; got != RejectionUnknownAction {
		t.Fatalf("rejection authority mutated through accessor: %q", got)
	}
}

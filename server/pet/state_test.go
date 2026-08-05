package pet

import (
	"encoding/json"
	"os"
	"testing"
)

type stateFixture struct {
	Version      int `json:"version"`
	Declarations struct {
		ActionIDs   []string `json:"action_ids"`
		BehaviorIDs []string `json:"behavior_ids"`
	} `json:"declarations"`
	Pets json.RawMessage `json:"pets"`
}

func TestCareStateMatchesSharedFixture(t *testing.T) {
	fixture := loadStateFixture(t)
	declarations := StateDeclarations{ActionIDs: fixture.Declarations.ActionIDs, BehaviorIDs: fixture.Declarations.BehaviorIDs}
	states, err := DecodeCareStates(fixture.Pets, declarations)
	if err != nil || fixture.Version != 1 || len(states) != 1 {
		t.Fatalf("states=%+v version=%d err=%v", states, fixture.Version, err)
	}
	state := states["01986666-7000-7000-8000-000000000001"]
	if state.StatsPPM[StatHunger] != 800000 || state.TrustPPM != 850000 ||
		state.BehaviorState != BehaviorActive || len(state.BehaviorQueue) != 2 || state.BehaviorPRNGCursor != 0 ||
		state.EvaluatedThroughAttendedMS != 123456 {
		t.Fatalf("state=%+v", state)
	}
	state.StatsPPM[StatHunger] = 0
	decoded, err := DecodeCareStates(fixture.Pets, declarations)
	if err != nil || decoded["01986666-7000-7000-8000-000000000001"].StatsPPM[StatHunger] != 800000 {
		t.Fatal("decoder exposed mutable state authority")
	}
}

func TestCareStateRejectsUnruledOrIncompleteState(t *testing.T) {
	fixture := loadStateFixture(t)
	declarations := StateDeclarations{ActionIDs: []string{"care.feed"}, BehaviorIDs: []string{"behavior.nap"}}
	valid := CareState{
		StatsPPM:                map[StatID]int64{StatHunger: 1, StatEnergy: 1, StatCleanliness: 1, StatAffection: 1},
		StatDecayRemaindersPPM:  map[StatID]int64{StatHunger: 0, StatEnergy: 0, StatCleanliness: 0, StatAffection: 0},
		CooldownUntilAttendedMS: map[string]int64{"care.feed": 0}, TrustPPM: 1, BehaviorState: BehaviorIdle,
		BehaviorQueue: []BehaviorQueueEntry{{BehaviorID: "behavior.nap", DueAttendedMS: 0}},
	}
	petID := "01986666-7000-7000-8000-000000000001"
	cases := map[string]map[string]CareState{}
	missingStat := cloneCareStates(map[string]CareState{petID: valid})
	delete(missingStat[petID].StatsPPM, StatHunger)
	cases["missing stat"] = missingStat
	unknownCooldown := cloneCareStates(map[string]CareState{petID: valid})
	unknownCooldown[petID].CooldownUntilAttendedMS["care.unknown"] = 1
	cases["unknown cooldown"] = unknownCooldown
	unknownBehavior := cloneCareStates(map[string]CareState{petID: valid})
	unknownBehavior[petID].BehaviorQueue[0].BehaviorID = "behavior.unknown"
	cases["unknown behavior"] = unknownBehavior
	badRemainder := cloneCareStates(map[string]CareState{petID: valid})
	badRemainderState := badRemainder[petID]
	badRemainderState.TrustDecayRemainderPPM = maxExactInteger + 1
	badRemainder[petID] = badRemainderState
	cases["bad remainder"] = badRemainder
	badCursor := cloneCareStates(map[string]CareState{petID: valid})
	badCursorState := badCursor[petID]
	badCursorState.BehaviorPRNGCursor = 1
	badCursor[petID] = badCursorState
	cases["vestigial prng cursor moved"] = badCursor
	badQueueOrder := cloneCareStates(map[string]CareState{petID: valid})
	badQueueState := badQueueOrder[petID]
	badQueueState.BehaviorQueue = []BehaviorQueueEntry{{BehaviorID: "behavior.nap", DueAttendedMS: 2}, {BehaviorID: "behavior.nap", DueAttendedMS: 1}}
	badQueueOrder[petID] = badQueueState
	cases["noncanonical queue"] = badQueueOrder
	longQueue := cloneCareStates(map[string]CareState{petID: valid})
	longQueueState := longQueue[petID]
	longQueueState.BehaviorQueue = make([]BehaviorQueueEntry, 9)
	longQueue[petID] = longQueueState
	cases["long queue"] = longQueue
	cases["bad pet id"] = map[string]CareState{"pet.one": valid}
	for name, states := range cases {
		t.Run(name, func(t *testing.T) {
			if ValidateCareStates(states, declarations) == nil {
				t.Fatal("invalid care state accepted")
			}
		})
	}
	duplicate := []byte(`{"01986666-7000-7000-8000-000000000001":{"stats_ppm":{"hunger":1,"hunger":2}}}`)
	if _, err := DecodeCareStates(duplicate, declarations); err == nil {
		t.Fatal("duplicate nested pet state key accepted")
	}
	withMood := []byte(`{"01986666-7000-7000-8000-000000000001":{"stats_ppm":{"hunger":1,"energy":1,"cleanliness":1,"affection":1},"stat_decay_remainders_ppm":{"hunger":0,"energy":0,"cleanliness":0,"affection":0},"cooldown_until_attended_ms":{},"trust_ppm":1,"trust_decay_remainder_ppm":0,"behavior_state":"idle","behavior_entered_at_attended_ms":0,"behavior_queue":[],"behavior_prng_cursor":0,"mood":"neutral"}}`)
	if _, err := DecodeCareStates(withMood, declarations); err == nil {
		t.Fatal("derived mood accepted as stored state")
	}
	var missingScalar map[string]map[string]json.RawMessage
	if err := json.Unmarshal(fixture.Pets, &missingScalar); err != nil {
		t.Fatal(err)
	}
	delete(missingScalar[petID], "trust_ppm")
	encodedMissingScalar, err := json.Marshal(missingScalar)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := DecodeCareStates(encodedMissingScalar, StateDeclarations{
		ActionIDs: fixture.Declarations.ActionIDs, BehaviorIDs: fixture.Declarations.BehaviorIDs,
	}); err == nil {
		t.Fatal("missing scalar pet state key accepted")
	}
}

func TestCareStateRemainderIsPinnedGridBound(t *testing.T) {
	catalog, err := LoadCatalog([]byte(fullCatalogFixture))
	if err != nil {
		t.Fatal(err)
	}
	state := CareState{
		StatsPPM:                map[StatID]int64{StatHunger: 1, StatEnergy: 1, StatCleanliness: 1, StatAffection: 1},
		StatDecayRemaindersPPM:  map[StatID]int64{StatHunger: catalog.StatPolicy.GridMS, StatEnergy: 0, StatCleanliness: 0, StatAffection: 0},
		CooldownUntilAttendedMS: map[string]int64{}, TrustPPM: 1, BehaviorState: BehaviorIdle, BehaviorQueue: []BehaviorQueueEntry{},
	}
	if ValidateCareStatesForCatalog(map[string]CareState{"01986666-7000-7000-8000-000000000001": state}, catalog) == nil {
		t.Fatal("remainder at pinned grid accepted")
	}
}

func loadStateFixture(t *testing.T) stateFixture {
	t.Helper()
	data, err := os.ReadFile("../../testdata/pet/state-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture stateFixture
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	return fixture
}

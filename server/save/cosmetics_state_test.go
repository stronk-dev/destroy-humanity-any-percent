package save

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"cloud-clicker/server/cosmetic"
	"cloud-clicker/server/economy"
)

func v24FounderState(t *testing.T) *State {
	t.Helper()
	state := v23FounderState(t)
	state.WireVersion = 24
	state.Cosmetics = &cosmetic.State{Owned: []string{"horse_armor"}, Equipped: map[string]string{v23PetID: "horse_armor"}}
	return state
}

// AC3: the Founder v24 codec round-trips cosmetics and rejects every listed
// shape violation. Catalog-aware rejections (unknown id, bundle without the
// artifact) are production.ValidateFoundationState's half.
func TestFounderV24CosmeticsRoundTripAndInvariants(t *testing.T) {
	catalog := stateCatalog(t)
	state := v24FounderState(t)
	encoded, err := EncodeState(state)
	if err != nil {
		t.Fatal(err)
	}
	restored, err := RestoreState(encoded, 24, catalog, economy.ScopeFounder, time.Time{})
	if err != nil || !restored.Cosmetics.Equal(state.Cosmetics) {
		t.Fatalf("v24 restore=%+v err=%v", restored, err)
	}
	var object map[string]json.RawMessage
	if err := json.Unmarshal(encoded, &object); err != nil {
		t.Fatal(err)
	}
	if string(object["cosmetics"]) != `{"owned":["horse_armor"],"equipped":{"`+v23PetID+`":"horse_armor"}}` {
		t.Fatalf("canonical cosmetics bytes = %s", object["cosmetics"])
	}
	load := func(mutate func(map[string]json.RawMessage), version int, scope economy.Scope) error {
		candidate := map[string]json.RawMessage{}
		for name, value := range object {
			candidate[name] = value
		}
		mutate(candidate)
		data, _ := json.Marshal(candidate)
		_, err := RestoreState(data, version, catalog, scope, time.Time{})
		return err
	}
	set := func(raw string) func(map[string]json.RawMessage) {
		return func(value map[string]json.RawMessage) { value["cosmetics"] = json.RawMessage(raw) }
	}
	rejections := map[string]error{
		"v24 without cosmetics":   load(func(value map[string]json.RawMessage) { delete(value, "cosmetics") }, 24, economy.ScopeFounder),
		"cosmetics at v23":        load(func(map[string]json.RawMessage) {}, 23, economy.ScopeFounder),
		"Company v24":             load(func(map[string]json.RawMessage) {}, 24, economy.ScopeCompany),
		"null owned":              load(set(`{"owned":null,"equipped":{}}`), 24, economy.ScopeFounder),
		"null equipped":           load(set(`{"owned":[],"equipped":null}`), 24, economy.ScopeFounder),
		"null cosmetics":          load(set(`null`), 24, economy.ScopeFounder),
		"unsorted owned":          load(set(`{"owned":["zebra_armor","horse_armor"],"equipped":{}}`), 24, economy.ScopeFounder),
		"duplicate owned":         load(set(`{"owned":["horse_armor","horse_armor"],"equipped":{}}`), 24, economy.ScopeFounder),
		"non-mechanical owned":    load(set(`{"owned":["Horse"],"equipped":{}}`), 24, economy.ScopeFounder),
		"equipped by a non-pet":   load(set(`{"owned":["horse_armor"],"equipped":{"01986666-bbbb-7bbb-8bbb-bbbbbbbbbbbb":"horse_armor"}}`), 24, economy.ScopeFounder),
		"equipped not owned":      load(set(`{"owned":[],"equipped":{"`+v23PetID+`":"horse_armor"}}`), 24, economy.ScopeFounder),
		"extra cosmetics key":     load(set(`{"owned":[],"equipped":{},"price":0}`), 24, economy.ScopeFounder),
		"missing equipped":        load(set(`{"owned":[]}`), 24, economy.ScopeFounder),
		"cosmetics as an array":   load(set(`[]`), 24, economy.ScopeFounder),
		"owned carries an amount": load(set(`{"owned":[{"id":"horse_armor","amount":0}],"equipped":{}}`), 24, economy.ScopeFounder),
	}
	for name, err := range rejections {
		if err == nil {
			t.Fatalf("%s accepted", name)
		}
	}
	for name, mutate := range map[string]func(*State){
		"cosmetics before v24": func(value *State) { value.WireVersion = 23 },
		"nil cosmetics at v24": func(value *State) { value.Cosmetics = nil },
		"nil owned":            func(value *State) { value.Cosmetics.Owned = nil },
		"equipped non-pet":     func(value *State) { value.Cosmetics.Equipped = map[string]string{"x": "horse_armor"} },
	} {
		candidate := v24FounderState(t)
		mutate(candidate)
		if _, err := EncodeState(candidate); !errors.Is(err, ErrInvalidState) {
			t.Fatalf("encode accepted %s: %v", name, err)
		}
	}
	empty := v24FounderState(t)
	empty.Cosmetics = cosmetic.NewState()
	encodedEmpty, err := EncodeState(empty)
	if err != nil {
		t.Fatal(err)
	}
	var emptyObject map[string]json.RawMessage
	if json.Unmarshal(encodedEmpty, &emptyObject) != nil || string(emptyObject["cosmetics"]) != `{"owned":[],"equipped":{}}` {
		t.Fatalf("empty cosmetics must encode canonical []/{}: %s", emptyObject["cosmetics"])
	}
}

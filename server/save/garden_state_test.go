package save

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"cloud-clicker/server/economy"
	"cloud-clicker/server/garden"
)

func v25FounderState(t *testing.T) *State {
	t.Helper()
	state := v24FounderState(t)
	state.WireVersion = 25
	salt, anchor, effect := "0123456789abcdef", int64(1_000_000), int64(1_000_000)
	state.ServerGarden = &garden.State{SaltHex: &salt, TickAnchorWallMS: &anchor, TickSeq: 7, SubstrateID: "bare_metal",
		Plots: []garden.Plot{{Row: 0, Col: 1, SpeciesID: "strain_a", AgeTicks: 3, MaturedEffectPPM: &effect}}, SeedCollection: []string{"strain_a", "strain_b"}}
	return state
}

// AC2 (codec half): the Founder v25 codec round-trips server_garden with
// exact keys; v24 rejects garden state and v25 rejects its absence.
func TestFounderV25GardenRoundTripAndInvariants(t *testing.T) {
	catalog := stateCatalog(t)
	state := v25FounderState(t)
	encoded, err := EncodeState(state)
	if err != nil {
		t.Fatal(err)
	}
	restored, err := RestoreState(encoded, 25, catalog, economy.ScopeFounder, time.Time{})
	if err != nil || !restored.ServerGarden.Equal(state.ServerGarden) || !restored.Cosmetics.Equal(state.Cosmetics) {
		t.Fatalf("v25 restore=%+v err=%v", restored, err)
	}
	var object map[string]json.RawMessage
	if err := json.Unmarshal(encoded, &object); err != nil {
		t.Fatal(err)
	}
	want := `{"salt_hex":"0123456789abcdef","tick_anchor_wall_ms":1000000,"tick_seq":7,"substrate_id":"bare_metal","substrate_set_wall_ms":null,"plots":[{"row":0,"col":1,"species_id":"strain_a","age_ticks":3,"matured_effect_ppm":1000000}],"seed_collection":["strain_a","strain_b"]}`
	if string(object["server_garden"]) != want {
		t.Fatalf("canonical server_garden bytes = %s", object["server_garden"])
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
		return func(value map[string]json.RawMessage) { value["server_garden"] = json.RawMessage(raw) }
	}
	replace := func(old, new string) func(map[string]json.RawMessage) {
		return set(strings.Replace(want, old, new, 1))
	}
	rejections := map[string]error{
		"v25 without server_garden": load(func(value map[string]json.RawMessage) { delete(value, "server_garden") }, 25, economy.ScopeFounder),
		"server_garden at v24":      load(func(map[string]json.RawMessage) {}, 24, economy.ScopeFounder),
		"Company v25":               load(func(map[string]json.RawMessage) {}, 25, economy.ScopeCompany),
		"null server_garden":        load(set(`null`), 25, economy.ScopeFounder),
		"missing tick_seq key":      load(replace(`"tick_seq":7,`, ``), 25, economy.ScopeFounder),
		"missing plot stage key":    load(replace(`,"matured_effect_ppm":1000000`, ``), 25, economy.ScopeFounder),
		"extra key":                 load(replace(`"tick_seq":7,`, `"tick_seq":7,"cursor":1,`), 25, economy.ScopeFounder),
		"uppercase salt":            load(replace(`0123456789abcdef`, `0123456789ABCDEF`), 25, economy.ScopeFounder),
		"unsorted plots":            load(replace(`"plots":[{"row":0,"col":1,`, `"plots":[{"row":1,"col":0,"species_id":"strain_a","age_ticks":0,"matured_effect_ppm":null},{"row":0,"col":1,`), 25, economy.ScopeFounder),
		"plot outside max grid":     load(replace(`"col":1,`, `"col":6,`), 25, economy.ScopeFounder),
		"effect over max":           load(replace(`"matured_effect_ppm":1000000`, `"matured_effect_ppm":10000001`), 25, economy.ScopeFounder),
		"unsorted seeds":            load(replace(`["strain_a","strain_b"]`, `["strain_b","strain_a"]`), 25, economy.ScopeFounder),
		"null plots":                load(replace(`"plots":[{"row":0,"col":1,"species_id":"strain_a","age_ticks":3,"matured_effect_ppm":1000000}]`, `"plots":null`), 25, economy.ScopeFounder),
		"fractional tick":           load(replace(`"tick_seq":7`, `"tick_seq":7.5`), 25, economy.ScopeFounder),
	}
	for name, err := range rejections {
		if !errors.Is(err, ErrInvalidState) {
			t.Errorf("%s: expected ErrInvalidState, got %v", name, err)
		}
	}
	stale := v24FounderState(t)
	stale.ServerGarden = state.ServerGarden.Clone()
	if _, err := EncodeState(stale); !errors.Is(err, ErrInvalidState) {
		t.Fatalf("a v24 Founder carrying server_garden encoded: %v", err)
	}
}

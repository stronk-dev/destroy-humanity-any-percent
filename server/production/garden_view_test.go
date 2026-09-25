package production

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"cloud-clicker/server/garden"
	"cloud-clicker/server/save"
)

func gardenViewFounder(t *testing.T, grown CatalogBundle, now time.Time) *save.State {
	t.Helper()
	state := reputationFounderState(t, grown, 25, now, 1)
	state.FiscalUnlocks[grown.Garden.UnlockID] = true
	salt, anchor := "0123456789abcdef", now.UnixMilli()
	state.ServerGarden.SaltHex, state.ServerGarden.TickAnchorWallMS = &salt, &anchor
	effect := int64(1_000_000)
	state.ServerGarden.Plots = []garden.Plot{{Row: 0, Col: 0, SpeciesID: "strain_a", AgeTicks: 3, MaturedEffectPPM: &effect},
		{Row: 1, Col: 1, SpeciesID: "strain_b", AgeTicks: 0}, {Row: 5, Col: 5, SpeciesID: "strain_a", AgeTicks: 0}}
	return state
}

// AC15: at equal server_ms, the projected view equals the garden the next
// command commits (projection = advance on a discarded clone).
func TestGardenViewMatchesTheNextCommit(t *testing.T) {
	pinGardenSalt(t)
	grown := gardenContentBundle(t)
	now := time.Date(2026, 9, 25, 12, 0, 0, 0, time.UTC)
	later := now.Add(5 * time.Hour)
	state := gardenViewFounder(t, grown, now)
	before, _ := save.EncodeState(state)
	encoded, err := ProjectGardenView(grown, state, 1, later.UnixMilli())
	if err != nil {
		t.Fatal(err)
	}
	if after, _ := save.EncodeState(state); string(after) != string(before) {
		t.Fatal("the read projection mutated the Founder state")
	}
	var view struct {
		Kind   string `json:"kind"`
		Garden struct {
			TickSeq int64 `json:"tick_seq"`
			Plots   []struct {
				Row, Col  int64
				SpeciesID string `json:"species_id"`
				Stage     string `json:"stage"`
				Dormant   bool   `json:"dormant"`
			} `json:"plots"`
		} `json:"garden"`
	}
	if err := json.Unmarshal(encoded, &view); err != nil || view.Kind != "active" {
		t.Fatalf("view %s err=%v", encoded, err)
	}
	runner := &gardenRunner{t: t, catalogs: grown, bundle: "grown", state: state, revision: 1}
	runner.apply("commits-at-the-same-instant", IntentGardenSetSubstrate, `"substrate_id":"mainframe"`, later)
	committed := state.ServerGarden
	if view.Garden.TickSeq != committed.TickSeq || len(view.Garden.Plots) != len(committed.Plots) {
		t.Fatalf("projection tick_seq=%d plots=%d, commit tick_seq=%d plots=%d", view.Garden.TickSeq, len(view.Garden.Plots), committed.TickSeq, len(committed.Plots))
	}
	for index, plot := range committed.Plots {
		got := view.Garden.Plots[index]
		if got.Row != plot.Row || got.Col != plot.Col || got.SpeciesID != plot.SpeciesID || (got.Stage == "mature") != plot.Mature() {
			t.Fatalf("plot %d projection %+v commit %+v", index, got, plot)
		}
	}
	if last := view.Garden.Plots[len(view.Garden.Plots)-1]; !last.Dormant {
		t.Fatalf("a plot outside the level-0 grid must project as dormant: %+v", last)
	}
}

// AC10: the projection never carries the hidden salt, the derived base, or a
// draw; locked and inactive gardens project their exact shapes; every shape
// validates against the registered response schema.
func TestGardenViewHidesTheSaltAndValidates(t *testing.T) {
	grown := gardenContentBundle(t)
	now := time.Date(2026, 9, 25, 12, 0, 0, 0, time.UTC)
	state := gardenViewFounder(t, grown, now)
	active, err := ProjectGardenView(grown, state, 3, now.Add(time.Hour).UnixMilli())
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"0123456789abcdef", "salt", "base", "draw"} {
		if strings.Contains(string(active), forbidden) {
			t.Fatalf("the garden view leaks %q: %s", forbidden, active)
		}
	}
	locked := gardenViewFounder(t, grown, now)
	delete(locked.FiscalUnlocks, grown.Garden.UnlockID)
	lockedView, err := ProjectGardenView(grown, locked, 3, now.UnixMilli())
	if err != nil || string(lockedView) != `{"kind":"locked","unlock_id":"minigame.server_garden"}` {
		t.Fatalf("locked view %s err=%v", lockedView, err)
	}
	shop := cosmeticsContentBundle(t)
	inactive, err := ProjectGardenView(shop, reputationFounderState(t, shop, 24, now, 1), 1, now.UnixMilli())
	if err != nil || string(inactive) != `{"kind":"inactive"}` {
		t.Fatalf("inactive view %s err=%v", inactive, err)
	}
	unsalted := gardenViewFounder(t, grown, now)
	unsalted.ServerGarden.SaltHex, unsalted.ServerGarden.TickAnchorWallMS = nil, nil
	fresh, err := ProjectGardenView(grown, unsalted, 1, now.UnixMilli())
	if err != nil || strings.Contains(string(fresh), projectionSalt) {
		t.Fatalf("unsalted view %s err=%v", fresh, err)
	}
	// The account package validates these exact bytes against the registered
	// response schema (account imports production, so the check lives there).
	fixture, err := json.MarshalIndent(map[string]json.RawMessage{"active": active, "fresh": fresh, "inactive": inactive, "locked": lockedView}, "", " ")
	if err != nil {
		t.Fatal(err)
	}
	fixture = append(fixture, '\n')
	if os.Getenv("GARDEN_UPDATE_FIXTURE") == "1" {
		if err := os.WriteFile("../../testdata/garden/view-fixtures-v1.json", fixture, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	pinned, err := os.ReadFile("../../testdata/garden/view-fixtures-v1.json")
	if err != nil || !bytes.Equal(pinned, fixture) {
		t.Fatalf("garden view fixtures drifted; regenerate with GARDEN_UPDATE_FIXTURE=1 (err=%v)", err)
	}
}

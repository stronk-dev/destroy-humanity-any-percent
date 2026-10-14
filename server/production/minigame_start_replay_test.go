package production

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"testing"
	"time"

	"cloud-clicker/server/economy"
	"cloud-clicker/server/minigame"
	"cloud-clicker/server/minigameapi"
	"cloud-clicker/server/save"
)

func TestMinigameStartReplayRequiresSoloPolicy(t *testing.T) {
	data, err := os.ReadFile("../../testdata/replay/apply-logged-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture crossRuntimeFixture
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	data, err = os.ReadFile("../../testdata/minigame/start-mode-replay-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	var population struct {
		Version int `json:"schema_version"`
		Cases   []struct {
			Name      string   `json:"name"`
			Modes     []string `json:"modes"`
			Accepted  bool     `json:"accepted"`
			FiscalDue bool     `json:"fiscal_due"`
		} `json:"cases"`
	}
	if err := json.Unmarshal(data, &population); err != nil || population.Version != 1 || len(population.Cases) != 4 {
		t.Fatalf("incomplete mode population: %v", err)
	}
	for _, row := range population.Cases {
		t.Run(row.Name, func(t *testing.T) {
			artifacts := make(map[string][]byte, len(fixture.MinigameStartFiles))
			for name, source := range fixture.MinigameStartFiles {
				artifacts[name] = []byte(source)
			}
			var root map[string]json.RawMessage
			if err := json.Unmarshal(artifacts["minigames"], &root); err != nil {
				t.Fatal(err)
			}
			var definitions []map[string]json.RawMessage
			if err := json.Unmarshal(root["minigames"], &definitions); err != nil || len(definitions) != 1 || string(definitions[0]["minigame_id"]) != `"pitch"` {
				t.Fatalf("expected one Pitch definition: %v", err)
			}
			definitions[0]["modes"], err = json.Marshal(row.Modes)
			if err != nil {
				t.Fatal(err)
			}
			root["minigames"], err = json.Marshal(definitions)
			if err != nil {
				t.Fatal(err)
			}
			artifacts["minigames"], err = json.Marshal(root)
			if err != nil {
				t.Fatal(err)
			}
			hash, err := save.ConstantsHashArtifacts(artifacts)
			if err != nil {
				t.Fatal(err)
			}
			bundle := pitchFeatureBundle(t)
			for name, source := range bundle.Artifacts {
				if !bytes.Equal(source, []byte(fixture.MinigameStartFiles[name])) {
					t.Fatalf("stored %s fixture differs from the typed Pitch context", name)
				}
			}
			bundle.Artifacts, bundle.ConstantsHash = artifacts, hash
			bundle.Minigames, err = minigame.LoadCatalog(artifacts["minigames"])
			if err != nil {
				t.Fatal(err)
			}
			bundle.MinigameAPI, err = minigameapi.LoadCatalog(artifacts["minigame_api"])
			if err != nil || !bundle.valid(hash) {
				t.Fatalf("mode fixture bundle invalid: %v", err)
			}
			original := fixture.MinigameStart
			state, err := save.RestoreState(original.PreState, original.StateVersion, bundle.Economy, economy.ScopeFounder, time.Time{})
			if err != nil {
				t.Fatal(err)
			}
			before := mustEncodeState(t, state)
			wire, err := parseFounderReplayInputs(original.ReplayInputs)
			if err != nil {
				t.Fatal(err)
			}
			inputs := original.ReplayInputs
			if row.FiscalDue {
				wire.Command.ServerTSMS = state.FiscalPeriodOpenedWallMS + bundle.Fiscal.Clock.AutoMS
				probe := fiscalStateFromSave(state)
				swept, err := bundle.Fiscal.Sweep(&probe, wire.Command.ServerTSMS)
				if err != nil || swept.Periods == 0 || probe.Credit == state.FiscalCredit {
					t.Fatalf("due-sweep control did not change isolated Fiscal state: %v", err)
				}
				inputs, err = save.MarshalFounderReplayInputs(wire.Command, wire.Resolved)
				if err != nil {
					t.Fatal(err)
				}
			}
			transition, replayErr := ApplyFounderLogged(state, original.CanonicalPayload, bundle, inputs)
			if row.Accepted {
				if replayErr != nil || transition.Outcome != save.IntentApplied || transition.ResultConstantsHash != hash ||
					canonicalFixtureJSON(t, transition.Receipt) != original.ReceiptJSON ||
					canonicalFixtureValue(t, fixtureEvents(transition.Events)) != original.EventsJSON ||
					canonicalFixtureJSON(t, mustEncodeState(t, state)) != original.PostStateJSON {
					t.Fatalf("accepted mode changed the committed golden: err=%v", replayErr)
				}
			} else {
				if !errors.Is(replayErr, ErrInvalidReplayInputs) {
					t.Errorf("async-only policy admitted a solo start: outcome=%s err=%v", transition.Outcome, replayErr)
				}
				if !jsonSemanticallyEqual(before, mustEncodeState(t, state)) {
					t.Errorf("invalid mode changed sequence or lazy Fiscal state: sequence=%d credit=%d", state.MinigameSessionSeq, state.FiscalCredit)
				}
			}
			if row.FiscalDue {
				return // A separate clock variant is not the committed golden history.
			}
			if original.EventsJSON != "[]" {
				t.Fatal("the stored start-history fixture has unexpected events")
			}
			appliedRevision := wire.Command.Revision + 1
			history := save.FounderHistory{FounderStreamID: wire.Command.FounderStreamID, FounderID: wire.Command.FounderID,
				Genesis: save.FounderGenesis{FounderStreamID: wire.Command.FounderStreamID, Revision: wire.Command.Revision,
					State: original.PreState, Version: original.StateVersion, ConstantsHash: hash},
				Entries: []save.FounderHistoryEntry{{Sequence: wire.Command.FounderLogSeq, IntentID: wire.Command.IntentID,
					CanonicalPayload: original.CanonicalPayload, ReplayInputs: inputs, Receipt: original.Receipt,
					AppliedRevision: &appliedRevision, ConstantsHash: hash, ServerTSMS: wire.Command.ServerTSMS, Events: []save.EventWrite{}}},
				HeadRevision: appliedRevision, HeadVersion: original.StateVersion, HeadConstants: hash, HeadState: original.PostState}
			want := ReplayStateDivergence
			if row.Accepted {
				want = ReplayVerified
			}
			if verdict := VerifyFounderHistory(history, ReplayCatalogSet{hash: bundle}); verdict != want {
				t.Errorf("complete Founder-history verdict=%s want=%s", verdict, want)
			}
		})
	}
}

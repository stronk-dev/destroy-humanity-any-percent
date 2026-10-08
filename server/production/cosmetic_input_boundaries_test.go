package production

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"testing"
	"time"

	"cloud-clicker/server/economy"
	"cloud-clicker/server/save"
)

// Both replay runtimes consume these literal admission boundaries. Invalid
// recorded context must fail as invalid replay, not become a gameplay refusal.
func TestCosmeticReplayInputBoundaries(t *testing.T) {
	data, err := os.ReadFile("../../testdata/replay/cosmetic-input-boundaries.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []struct {
		Name     string          `json:"name"`
		Base     string          `json:"base"`
		Field    string          `json:"field"`
		Resolved bool            `json:"resolved"`
		Omit     bool            `json:"omit"`
		Value    json.RawMessage `json:"value"`
		Valid    bool            `json:"valid"`
	}
	if err := json.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}
	corpusData, err := os.ReadFile(cosmeticCorpusPath)
	if err != nil {
		t.Fatal(err)
	}
	var corpus cosmeticCorpus
	if err := json.Unmarshal(corpusData, &corpus); err != nil {
		t.Fatal(err)
	}
	catalogs := cosmeticsContentBundle(t)
	for _, row := range cases {
		t.Run(row.Name, func(t *testing.T) {
			var base *reputationCorpusCase
			for i := range corpus.Cases {
				if corpus.Cases[i].Name == row.Base {
					base = &corpus.Cases[i]
					break
				}
			}
			if base == nil {
				t.Fatalf("missing base %s", row.Base)
			}
			inputs := reputationShapeObject(t, base.ReplayInputs)
			resolved := reputationShapeObject(t, inputs["resolved"])
			if row.Field != "" {
				target := resolved
				if !row.Resolved {
					target = reputationShapeObject(t, resolved["active_company"])
				}
				if row.Omit {
					delete(target, row.Field)
				} else {
					target[row.Field] = row.Value
				}
				if !row.Resolved {
					resolved["active_company"] = reputationShapeJSON(t, target)
				}
			}
			inputs["resolved"] = reputationShapeJSON(t, resolved)
			state, err := save.RestoreState(base.PreState, base.StateVersion, catalogs.Economy, economy.ScopeFounder, time.Time{})
			if err != nil {
				t.Fatal(err)
			}
			before := mustEncodeState(t, state)
			payload := []byte(canonicalFixtureJSON(t, base.CanonicalPayload))
			transition, err := ApplyFounderLogged(state, payload, catalogs, reputationShapeJSON(t, inputs))
			if row.Valid {
				if err != nil || transition.Outcome != save.IntentApplied {
					t.Fatalf("valid context outcome=%s err=%v", transition.Outcome, err)
				}
				if canonicalFixtureJSON(t, transition.Receipt) != base.ReceiptJSON || canonicalFixtureValue(t, fixtureEvents(transition.Events)) != base.EventsJSON ||
					canonicalFixtureJSON(t, mustEncodeState(t, state)) != base.PostStateJSON {
					t.Fatal("valid context changed canonical receipt, event or final state")
				}
			} else if !errors.Is(err, ErrInvalidReplayInputs) || !bytes.Equal(before, mustEncodeState(t, state)) {
				t.Fatalf("invalid context outcome=%s err=%v restored=%t", transition.Outcome, err, bytes.Equal(before, mustEncodeState(t, state)))
			}
		})
	}
}

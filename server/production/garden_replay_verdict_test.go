package production

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"cloud-clicker/server/economy"
	"cloud-clicker/server/save"
)

// SG6/SG8/AC9: use the committed four-shape corpus, never a regenerated
// expectation. These runs are deliberately nonterminal; honest verdict is
// log_gap, whereas a mismatched hash must diverge before the terminal check.
func TestGardenCompanyHashVerdicts(t *testing.T) {
	encoded, err := os.ReadFile(gardenCorpusPath)
	if err != nil {
		t.Fatal(err)
	}
	var corpus gardenCorpus
	if err := json.Unmarshal(encoded, &corpus); err != nil {
		t.Fatal(err)
	}
	if len(corpus.CompanyCases) != 4 {
		t.Fatalf("Company population=%d, want four pinned shapes", len(corpus.CompanyCases))
	}
	catalogs := gardenContentBundle(t)
	if corpus.Bundles["grown"].ConstantsHash != catalogs.ConstantsHash {
		t.Fatal("committed replay corpus identity differs from loaded Garden bundle")
	}
	falseHash := "sha256:" + strings.Repeat("0", 64)
	for _, row := range corpus.CompanyCases {
		t.Run(row.Name, func(t *testing.T) {
			// The corpus stores a JSON object in an indented document. Feed
			// canonical command bytes, just as the TS fixture reader does.
			payload, err := normalizeReplayJSON(row.CanonicalPayload)
			if err != nil {
				t.Fatal(err)
			}
			state, err := save.RestoreState(row.PreState, row.StateVersion, catalogs.Economy, economy.ScopeCompany, time.Time{})
			if err != nil {
				t.Fatal(err)
			}
			transition, err := ApplyLogged(state, payload, catalogs, row.ReplayInputs)
			if err != nil || string(transition.Outcome) != row.Outcome || !canonicalJSONEqual(transition.Receipt, []byte(row.ReceiptJSON)) ||
				!canonicalJSONEqual(marshalReplayEvents(transition.Events), []byte(row.EventsJSON)) ||
				!canonicalJSONEqual(mustEncodeState(t, transition.State), []byte(row.PostStateJSON)) {
				t.Fatalf("honest pinned transition diverged: %v", err)
			}
			entry := ReplayLogEntry{Sequence: 1, CanonicalPayload: payload, ReplayInputs: row.ReplayInputs,
				ReceiptJSON: []byte(row.ReceiptJSON), EventsJSON: []byte(row.EventsJSON)}
			verify := func(entry ReplayLogEntry) ReplayVerdict {
				return VerifyReplayRun(row.PreState, row.StateVersion, catalogs, []ReplayLogEntry{entry}, catalogs.ConstantsHash, false)
			}
			if verdict := verify(entry); verdict != ReplayLogGap {
				t.Fatalf("honest nonterminal verdict=%s, want log_gap", verdict)
			}
			for _, target := range []string{"payload", "resolved"} {
				t.Run(target, func(t *testing.T) {
					poisoned := entry
					if target == "payload" {
						var payload gardenHarvestCreditPayload
						if err := json.Unmarshal(entry.CanonicalPayload, &payload); err != nil || payload.HarvestHash == falseHash {
							t.Fatalf("payload probe is not discriminating: %v", err)
						}
						payload.HarvestHash = falseHash
						poisoned.CanonicalPayload, err = normalizeReplayJSON(mustJSON(payload))
					} else {
						var wire replayInputsWire
						if err := json.Unmarshal(entry.ReplayInputs, &wire); err != nil {
							t.Fatal(err)
						}
						var resolved gardenCompanyResolved
						if err := json.Unmarshal(wire.Resolved, &resolved); err != nil || resolved.HarvestHash == falseHash {
							t.Fatalf("resolved probe is not discriminating: %v", err)
						}
						resolved.HarvestHash = falseHash
						wire.Resolved = mustJSON(resolved)
						poisoned.ReplayInputs, err = json.Marshal(wire)
					}
					if err != nil {
						t.Fatal(err)
					}
					if verdict := verify(poisoned); verdict != ReplayStateDivergence {
						t.Fatalf("tampered %s verdict=%s, want state_divergence", target, verdict)
					}
					if verdict := verify(entry); verdict != ReplayLogGap {
						t.Fatalf("probe mutated original evidence: %s", verdict)
					}
				})
			}
		})
	}
}

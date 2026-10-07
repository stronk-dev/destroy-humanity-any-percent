package production

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"cloud-clicker/server/economy"
	"cloud-clicker/server/save"
)

const permitsReplayPath = "../../testdata/replay/permits-gate-v1.json"

// This corpus isolates the ratified Permits contract. It does not stand in for
// the full content epoch, persistence, or a mounted player journey.
func TestPermitsGateReplayCorpus(t *testing.T) {
	artifacts := cloneArtifactMap(epoch5TestBundle(t).Artifacts)
	for name, path := range map[string]string{
		"categories": "../../balance/testdata/first-content/categories-v1.json",
		"economy":    "../../balance/testdata/valid/permits-economy-candidate-v1.json",
		"routes":     "../../balance/testdata/permits-t3-gate-candidate-v1.json",
	} {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		artifacts[name] = data
	}
	hash, err := save.ConstantsHashArtifacts(artifacts)
	if err != nil {
		t.Fatal(err)
	}
	bundle := loadReplayTestBundle(t, hash, artifacts)
	corpus := struct {
		Version       int                       `json:"version"`
		StateVersion  int                       `json:"state_version"`
		ConstantsHash string                    `json:"constants_hash"`
		Artifacts     map[string]string         `json:"artifacts"`
		Cases         []crossRuntimeFixtureCase `json:"cases"`
	}{Version: 1, StateVersion: 14, ConstantsHash: hash, Artifacts: artifactStrings(artifacts)}
	now := time.Date(2026, 8, 8, 12, 0, 0, 0, time.UTC)
	for index, arm := range []struct {
		name, cash, permits, afterCash, afterPermits, missing string
	}{
		{"cash-short", "9.99999999999e11", "1.2e1", "9.99999999999e11", "1.2e1", "company.cash"},
		{"permits-short", "1e12", "1.1999e1", "1e12", "1.1999e1", "company.permits"},
		{"both-short-cash-first", "9.99999999999e11", "1.1999e1", "9.99999999999e11", "1.1999e1", "company.cash"},
		{"exact-requirements", "1e12", "1.2e1", "0", "0", ""},
		{"surplus-preserved", "1.5e12", "2.4e1", "5e11", "1.2e1", ""},
	} {
		t.Run(arm.name, func(t *testing.T) {
			state := replayFixtureState(t, bundle.Economy, now)
			state.Tier = 3
			state.Ledger, err = economy.RestoreLedger(bundle.Economy, economy.ScopeCompany, map[string]string{
				"company.cash": arm.cash, "company.permits": arm.permits,
			})
			if err != nil {
				t.Fatal(err)
			}
			pre := mustEncodeState(t, state)
			request, err := ParseIntent([]byte(fmt.Sprintf(`{"intent_id":"01989999-%04d-7000-8000-%012d","kind":"cross_gate","expected_revision":1,"gate_id":"gate.t3_to_t4","route_id":null}`, index+1, index+1)))
			if err != nil {
				t.Fatal(err)
			}
			command := save.ReplayCommand{IntentID: request.IntentID, CompanyStreamID: "01989999-1000-7000-8000-000000000001",
				FounderID: "01989999-2000-7000-8000-000000000001", Revision: 1, RunSeq: 1, RunLogSeq: 1}
			inputs, err := buildReplayInputs(replayBuild{Command: command, Mode: ModeOnline, Now: now,
				IntentKind: request.Kind, RouteContextVersion: bundle.Routes.ContextVersion(),
				FounderCarry: &replayFounderCarry{FounderRevision: 1, FounderConstantsHash: hash,
					NetworkSlots: []save.NetworkSlot{}, LedgerFactKinds: []string{}, ExitHistoryCount: 0}})
			if err != nil {
				t.Fatal(err)
			}
			// Independently execute the live gate handler, then restore its input
			// and reproduce the same receipt/events/state through logged replay.
			live, err := TransitionWithRoutes(request, state, bundle.Economy, bundle.Routes,
				save.Revision{StreamID: command.CompanyStreamID, OwnerID: command.FounderID, Number: 1, ConstantsHash: hash}, ModeOnline, now, nil, nil)
			if err != nil {
				t.Fatal(err)
			}
			livePost := mustEncodeState(t, state)
			state, err = save.RestoreState(pre, corpus.StateVersion, bundle.Economy, economy.ScopeCompany, time.Time{})
			if err != nil {
				t.Fatal(err)
			}
			result, err := ApplyLogged(state, request.CanonicalPayload, bundle, inputs)
			if err != nil {
				t.Fatal(err)
			}
			post := mustEncodeState(t, state)
			if result.Outcome != live.Outcome || !bytes.Equal(result.Receipt, live.Receipt) ||
				canonicalFixtureValue(t, fixtureEvents(result.Events)) != canonicalFixtureValue(t, fixtureEvents(live.Events)) || !bytes.Equal(post, livePost) {
				t.Fatal("logged replay differs from the live gate handler")
			}
			balances := state.Ledger.Snapshot()
			if balances["company.cash"] != arm.afterCash || balances["company.permits"] != arm.afterPermits {
				t.Fatalf("incorrect exact debit: %v", balances)
			}
			if arm.missing != "" {
				category, detail := rejectionOf(t, result.Receipt)
				if result.Outcome != save.IntentRejected || category != "requirement_not_met" || detail != arm.missing ||
					!bytes.Equal(pre, post) || len(result.Events) != 0 {
					t.Fatal("refusal changed state/events or returned the wrong typed reason")
				}
			} else if result.Outcome != save.IntentApplied || state.Tier != 4 || !state.GatesCrossed[request.GateID] ||
				len(result.Events) != 1 || result.Events[0].Kind != save.EventGateCrossed {
				t.Fatal("successful debit did not cross exactly one gate")
			}
			corpus.Cases = append(corpus.Cases, crossRuntimeFixtureCase{Name: arm.name, PreState: pre,
				CanonicalPayload: request.CanonicalPayload, ReplayInputs: inputs, Outcome: string(result.Outcome),
				Receipt: result.Receipt, Events: fixtureEvents(result.Events), PostState: post,
				ReceiptJSON: canonicalFixtureJSON(t, result.Receipt), EventsJSON: canonicalFixtureValue(t, fixtureEvents(result.Events)),
				PostStateJSON: canonicalFixtureJSON(t, post)})
		})
	}
	if t.Failed() {
		return // Never regenerate a corpus whose independent assertions failed.
	}
	encoded, err := json.MarshalIndent(corpus, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	encoded = append(encoded, '\n')
	if *updateReplayFixture {
		if err := os.WriteFile(permitsReplayPath, encoded, 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	pinned, err := os.ReadFile(permitsReplayPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(pinned, encoded) {
		t.Fatal("Permits replay corpus differs from executed Go behavior; regenerate explicitly and review")
	}
}

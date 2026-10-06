package production

import (
	"bytes"
	"encoding/json"
	"os"
	"slices"
	"testing"
	"time"

	"cloud-clicker/server/decimal"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/save"
)

const reputationBoundaryCorpusPath = "../../testdata/replay/reputation-exit-boundary-v1.json"

type reputationBoundaryReplayRow struct {
	Name     string                   `json:"name"`
	Command  string                   `json:"command"`
	Profile  string                   `json:"profile"`
	Current  string                   `json:"current"`
	Next     string                   `json:"next"`
	Category string                   `json:"category"`
	Detail   string                   `json:"detail"`
	Company  crossRuntimeTerminalCase `json:"company"`
	Founder  reputationCorpusCase     `json:"founder"`
}

type reputationBoundaryReplayCorpus struct {
	Version int                               `json:"version"`
	Bundles map[string]reputationCorpusBundle `json:"bundles"`
	Cases   []reputationBoundaryReplayRow     `json:"cases"`
}

// R8 portable supplement: diagnostic genesis, not SQL persistence, naturally
// earned progression or a minted release. Both historical corpora stay intact.
func buildReputationBoundaryReplayCorpus(t *testing.T) reputationBoundaryReplayCorpus {
	t.Helper()
	live, tree := activeContentBundle(t), reputationContentBundle(t)
	now := time.Date(2026, 10, 6, 12, 0, 0, 0, time.UTC)
	good := []string{"reputation.unlock.p05", "reputation.starter.cash_small", "reputation.starter.generated_beige_tower"}
	profiles := []struct {
		name, category, detail string
		plan                   []string
	}{
		{"activate-plan", "", "", good},
		{"activate-absent", "", "", nil},
		{"activate-empty", "", "", []string{}},
		{"next-inactive", "not_eligible", "reputation_plan.tree_inactive", good[:1]},
		{"unknown-prefix", "unknown_id", "reputation_plan.unknown_id", []string{good[0], "reputation.missing"}},
		{"owned-prefix", "not_eligible", "reputation_plan.owned", []string{good[1], good[0]}},
		{"requires-prefix", "not_eligible", "reputation_plan.requires", []string{good[0], good[2]}},
		{"unaffordable-prefix", "unaffordable", "reputation_plan.reputation", append(slices.Clone(good), "reputation.unlock.p25")},
		{"inactive-no-plan", "", "", nil},
	}
	corpus := reputationBoundaryReplayCorpus{Version: 1, Bundles: map[string]reputationCorpusBundle{
		"live": {ConstantsHash: live.ConstantsHash, Artifacts: stringArtifacts(live.Artifacts)},
		"tree": {ConstantsHash: tree.ConstantsHash, Artifacts: stringArtifacts(tree.Artifacts)},
	}, Cases: []reputationBoundaryReplayRow{}}
	for _, command := range []string{"wind_down", "acquihire", "acquisition"} {
		for _, profile := range profiles {
			current, next, version, currentName, nextName := tree, tree, 22, "tree", "tree"
			activation := profile.name == "activate-plan" || profile.name == "activate-absent" || profile.name == "activate-empty"
			if activation {
				current, version, currentName = live, 21, "live"
			}
			if profile.name == "next-inactive" || profile.name == "inactive-no-plan" {
				current, next, version, currentName, nextName = live, live, 21, "live", "live"
			}
			name := command + "-" + profile.name
			var founderPre json.RawMessage
			source := makeReputationPlanExitCase(t, name, current, next, version, 6, profile.plan, now, func(company, founder *save.State) {
				company.Tier, company.LifetimeValue = 3, decimal.Zero
				if profile.name == "owned-prefix" {
					founder.ReputationSpent, founder.ReputationUnlockPPM, founder.ReputationNodesOwned = 1, 50000, []string{good[0]}
				}
				if command != "wind_down" {
					company.OfferState = &save.ExitOfferState{OfferID: "01986666-e201-7000-8000-000000000001", ExitType: command,
						SpawnedAt: now.Add(-time.Second), ExpiresAt: now.Add(time.Minute),
						TermsJSON: json.RawMessage(`{"market_modifier_ppm":1000000,"payout_preview":{"reputation_delta":0,"network_slot_unlocks":[],"route_knowledge":0,"clout_reach_note":"clout.reach.preserved"}}`)}
				}
				founderPre = mustEncodeState(t, founder)
			})
			linked := current
			if current.ConstantsHash != next.ConstantsHash {
				linked.Next = &next
			}
			result := source.Company.Case
			company := replayFixtureStateFromEncoded(t, current, result.PreState)
			transition, err := ApplyLoggedExit(company, result.CanonicalPayload, linked, result.ReplayInputs)
			if err != nil {
				t.Fatalf("%s Company: %v", name, err)
			}
			if profile.category != "" {
				category, detail := rejectionOf(t, transition.Decision.Receipt)
				if category != profile.category || detail != profile.detail || source.Founder != nil || !bytes.Equal(mustEncodeState(t, company), result.PreState) || transition.Decision.NewCompanyState != nil || len(transition.Decision.FounderEvents) != 0 || len(transition.Decision.CompanyEndedEvents) != 0 || len(transition.Decision.CompanyStartedEvents) != 0 {
					t.Fatalf("%s wrong whole-Exit first failure: %s", name, transition.Decision.Receipt)
				}
				var wire replayInputsWire
				if err := json.Unmarshal(result.ReplayInputs, &wire); err != nil {
					t.Fatal(err)
				}
				var resolved replayExitResolved
				if err := json.Unmarshal(wire.Resolved, &resolved); err != nil {
					t.Fatal(err)
				}
				result = withTerminalRawJSON(t, crossRuntimeTerminalCase{Name: name, PreState: result.PreState, CanonicalPayload: result.CanonicalPayload,
					ReplayInputs: result.ReplayInputs, Outcome: result.Outcome, Receipt: transition.Decision.Receipt,
					FounderOutput: replayFounderOutput(transition.Founder, resolved.FounderCarry), FinalCompany: mustEncodeState(t, company), NewCompany: json.RawMessage(`null`),
					FounderEvents: fixtureEvents(transition.Decision.FounderEvents), CompanyEndedEvents: fixtureEvents(transition.Decision.CompanyEndedEvents), CompanyStartedEvents: fixtureEvents(transition.Decision.CompanyStartedEvents)})
				founder, err := save.RestoreState(founderPre, version, current.Economy, economy.ScopeFounder, time.Time{})
				if err != nil {
					t.Fatal(err)
				}
				audit, _, err := buildFounderExitAudit(wire.Command, save.Revision{OwnerID: wire.Command.FounderID, Number: 1, ConstantsHash: current.ConstantsHash}, founder, transition.Founder, transition.Decision, linked)
				if err != nil {
					t.Fatal(err)
				}
				inputs, err := save.MarshalFounderReplayInputs(save.FounderReplayCommand{IntentID: wire.Command.IntentID, FounderStreamID: "01986666-8c00-4000-8000-000000000001", FounderID: wire.Command.FounderID, Revision: 1, FounderLogSeq: 1, ServerTSMS: now.UnixMilli()}, audit)
				if err != nil {
					t.Fatal(err)
				}
				founderResult, err := ApplyFounderLogged(founder, result.CanonicalPayload, linked, inputs)
				if err != nil {
					t.Fatal(err)
				}
				category, detail = rejectionOf(t, founderResult.Receipt)
				if category != profile.category || detail != profile.detail || founderResult.Outcome != save.IntentRejected || founderResult.ResultConstantsHash != current.ConstantsHash || !bytes.Equal(mustEncodeState(t, founder), founderPre) || len(founderResult.Events) != 0 {
					t.Fatalf("%s Founder refusal differs", name)
				}
				source.Founder = &reputationCorpusCase{Name: name + "-founder", Bundle: currentName, StateVersion: version, PreState: founderPre, CanonicalPayload: result.CanonicalPayload, ReplayInputs: inputs,
					Outcome: string(founderResult.Outcome), ReceiptJSON: canonicalFixtureJSON(t, founderResult.Receipt), EventsJSON: canonicalFixtureValue(t, fixtureEvents(founderResult.Events)), PostStateJSON: canonicalFixtureJSON(t, mustEncodeState(t, founder))}
			} else {
				if result.Outcome != "applied" || source.Founder == nil {
					t.Fatalf("%s valid Exit failed: %s", name, result.Receipt)
				}
				postVersion := 21
				if activation {
					postVersion = 22
				}
				founder, err := save.RestoreState([]byte(source.Founder.PostStateJSON), postVersion, next.Economy, economy.ScopeFounder, time.Time{})
				if err != nil {
					t.Fatal(err)
				}
				planned := profile.name == "activate-plan"
				spent, unlock, factor, owned, starters := int64(0), int64(0), "1e0", []string{}, []string{}
				if planned {
					spent, unlock, factor, owned, starters = 6, 50000, "1.003e0", []string{good[1], good[2], good[0]}, good[1:]
				}
				if save.VersionForState(founder) != postVersion || founder.ReputationLevel != 6 || founder.ReputationSpent != spent || founder.ReputationUnlockPPM != unlock || !slices.Equal(founder.ReputationNodesOwned, owned) || transition.Decision.NewConstantsHash != next.ConstantsHash {
					t.Fatalf("%s independent activation/accounting differs", name)
				}
				newCompany := transition.Decision.NewCompanyState
				if newCompany.RunSeq != 3 || newCompany.OfferState != nil {
					t.Fatal("wrong next Company identity")
				}
				if planned {
					cash, ok := newCompany.Ledger.Balance("company.cash")
					if !ok || cash.String() != "1e3" || newCompany.GeneratorProvisioned["generator.beige_tower"] != 5 || newCompany.GeneratorCounts["generator.beige_tower"] != 0 || newCompany.GeneratorPurchasedTotal != 0 || len(result.FounderEvents) != 4 || result.FounderEvents[0].Kind != "founder_advanced" {
						t.Fatal("wrong independent starters/events")
					}
					for index, id := range good {
						var purchase struct {
							NodeID string `json:"node_id"`
							Cost   int64  `json:"cost"`
							Source string `json:"source"`
						}
						if err := json.Unmarshal(result.FounderEvents[index+1].Payload, &purchase); err != nil || purchase.NodeID != id || purchase.Cost != int64(index+1) || purchase.Source != "exit_plan" || result.FounderEvents[index+1].Kind != string(save.EventReputationNodePurchased) || result.FounderEvents[index+1].SchemaVersion != 1 {
							t.Fatal("wrong independent ordered purchase event")
						}
					}
				}
				if len(result.CompanyStartedEvents) != 1 {
					t.Fatal("missing start event")
				}
				var start struct {
					ReputationTree *reputationRunStarted `json:"reputation_tree"`
				}
				if err := json.Unmarshal(result.CompanyStartedEvents[0].Payload, &start); err != nil {
					t.Fatal(err)
				}
				if activation {
					if result.CompanyStartedEvents[0].SchemaVersion != 2 || start.ReputationTree == nil || start.ReputationTree.BonusFactor != factor || !slices.Equal(start.ReputationTree.AppliedStarterNodeIDs, starters) {
						t.Fatal("wrong independent activation start summary")
					}
				} else if start.ReputationTree != nil {
					t.Fatal("absent-tree no-plan Exit activated Reputation")
				}
			}
			corpus.Cases = append(corpus.Cases, reputationBoundaryReplayRow{Name: name, Command: command, Profile: profile.name, Current: currentName, Next: nextName,
				Category: profile.category, Detail: profile.detail, Company: result, Founder: *source.Founder})
		}
	}
	return corpus
}

func TestReputationBoundaryReplayCorpus(t *testing.T) {
	corpus := buildReputationBoundaryReplayCorpus(t)
	applied := 0
	for _, row := range corpus.Cases {
		if row.Company.Outcome == "applied" {
			applied++
		}
	}
	if len(corpus.Cases) != 27 || applied != 12 {
		t.Fatal("boundary corpus population must be27/12/15")
	}
	for _, offset := range []int{0, 9, 18} {
		absent, empty := corpus.Cases[offset+1].Company, corpus.Cases[offset+2].Company
		if bytes.Equal(absent.CanonicalPayload, empty.CanonicalPayload) {
			t.Fatal("absent/empty request identity collapsed")
		}
		for _, pair := range [][2]string{{absent.ReceiptJSON, empty.ReceiptJSON}, {absent.FounderOutputJSON, empty.FounderOutputJSON}, {absent.FinalCompanyJSON, empty.FinalCompanyJSON}, {absent.NewCompanyJSON, empty.NewCompanyJSON}, {absent.FounderEventsJSON, empty.FounderEventsJSON}, {absent.CompanyEndedJSON, empty.CompanyEndedJSON}, {absent.CompanyStartedJSON, empty.CompanyStartedJSON}} {
			if pair[0] != pair[1] {
				t.Fatal("absent/empty semantic output differs")
			}
		}
	}
	if t.Failed() {
		return
	}
	encoded, err := json.MarshalIndent(corpus, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	encoded = append(encoded, '\n')
	if *updateReplayFixture {
		if err := os.WriteFile(reputationBoundaryCorpusPath, encoded, 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	pinned, err := os.ReadFile(reputationBoundaryCorpusPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(pinned, encoded) {
		t.Fatal("boundary corpus differs from executed Go transitions; explicit generation/review required")
	}
}

func TestReputationBoundaryReplayRefusesCopiedEvidence(t *testing.T) {
	corpus := buildReputationBoundaryReplayCorpus(t)
	live, tree := activeContentBundle(t), reputationContentBundle(t)
	bundles := map[string]CatalogBundle{"live": live, "tree": tree}
	controls := 0
	for _, row := range corpus.Cases {
		if row.Company.Outcome != "rejected" && row.Current == row.Next {
			continue
		}
		t.Run(row.Name, func(t *testing.T) {
			current, next := bundles[row.Current], bundles[row.Next]
			if row.Current != row.Next {
				current.Next = &next
			}
			var inputs map[string]any
			if err := json.Unmarshal(row.Founder.ReplayInputs, &inputs); err != nil {
				t.Fatal(err)
			}
			resolved := inputs["resolved"].(map[string]any)
			if row.Company.Outcome == "rejected" {
				resolved["reputation_delta"] = float64(1)
			} else {
				resolved["result_constants_hash"] = string(bytes.Repeat([]byte("0"), 64))
			}
			encoded, err := json.Marshal(inputs)
			if err != nil {
				t.Fatal(err)
			}
			founder, err := save.RestoreState(row.Founder.PreState, row.Founder.StateVersion, current.Economy, economy.ScopeFounder, time.Time{})
			if err != nil {
				t.Fatal(err)
			}
			if _, err := ApplyFounderLogged(founder, row.Founder.CanonicalPayload, current, encoded); err == nil {
				t.Fatal("copied delta/result hash replayed")
			}
			controls++
			if row.Profile == "activate-plan" {
				if err := json.Unmarshal(row.Founder.ReplayInputs, &inputs); err != nil {
					t.Fatal(err)
				}
				inputs["resolved"].(map[string]any)["reputation_purchases"].([]any)[0].(map[string]any)["resolved_cost"] = float64(2)
				encoded, err = json.Marshal(inputs)
				if err != nil {
					t.Fatal(err)
				}
				founder, err = save.RestoreState(row.Founder.PreState, row.Founder.StateVersion, current.Economy, economy.ScopeFounder, time.Time{})
				if err != nil {
					t.Fatal(err)
				}
				if _, err := ApplyFounderLogged(founder, row.Founder.CanonicalPayload, current, encoded); err == nil {
					t.Fatal("copied planned cost replayed")
				}
				controls++
			}
		})
	}
	if controls != 27 {
		t.Fatalf("copied evidence controls=%d want27", controls)
	}
}

package production

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"testing"
	"time"

	"cloud-clicker/server/decimal"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/save"
)

const reputationOfferCorpusPath = "../../testdata/replay/reputation-offer-plan-v1.json"

type reputationOfferReplayCase struct {
	Name    string                   `json:"name"`
	Kind    string                   `json:"kind"`
	Level   int64                    `json:"level"`
	Company crossRuntimeTerminalCase `json:"company"`
	Founder *reputationCorpusCase    `json:"founder"`
}

type reputationOfferReplayCorpus struct {
	Version int                         `json:"version"`
	Bundle  reputationCorpusBundle      `json:"bundle"`
	Cases   []reputationOfferReplayCase `json:"cases"`
}

// Stored offers here are diagnostic replay inputs. Live offer generation and
// SQL retries/rollback have their own integration population; this is R8 parity.
func buildReputationOfferReplayCorpus(t *testing.T) reputationOfferReplayCorpus {
	t.Helper()
	bundle := reputationContentBundle(t)
	now := time.Date(2026, 10, 6, 12, 0, 0, 0, time.UTC)
	good := []string{"reputation.unlock.p05", "reputation.starter.cash_small", "reputation.starter.generated_beige_tower"}
	bad := append(slices.Clone(good), "reputation.unlock.p25", "reputation.starter.cash_large")
	corpus := reputationOfferReplayCorpus{Version: 1, Bundle: reputationCorpusBundle{
		ConstantsHash: bundle.ConstantsHash, Artifacts: stringArtifacts(bundle.Artifacts)}, Cases: []reputationOfferReplayCase{}}
	for _, offer := range []struct {
		kind, factor    string
		level, modifier int64
	}{{"acquihire", "1.009e0", 18, 900000}, {"acquisition", "1.01e0", 20, 1000000}} {
		if modifier, ok := bundle.Prestige.Modifier(offer.kind); !ok || modifier != offer.modifier {
			t.Fatal("diagnostic payout requires unchanged pinned modifiers")
		}
		for _, arm := range []struct {
			name        string
			plan        []string
			promiseOnly bool
		}{{"payout-funded", good, false}, {"last-unaffordable", bad, false}, {"absent", nil, false}, {"empty", []string{}, false}, {"promise-floor", good, true}} {
			name := offer.kind + "-" + arm.name
			row := makeReputationPlanExitCase(t, name, bundle, bundle, 22, 0, arm.plan, now, func(company, founder *save.State) {
				company.Tier = 3
				company.LifetimeValue = bundle.Prestige.ThresholdValue().Mul(decimal.New(8, 3))
				if arm.promiseOnly {
					company.LifetimeValue = decimal.Zero
				}
				terms := json.RawMessage(fmt.Sprintf(`{"market_modifier_ppm":1000000,"payout_preview":{"reputation_delta":%d,"network_slot_unlocks":[],"route_knowledge":0,"clout_reach_note":"clout.reach.preserved"}}`, offer.level))
				company.OfferState = &save.ExitOfferState{OfferID: "01986666-d001-7000-8000-000000000001", ExitType: offer.kind,
					TermsJSON: terms, SpawnedAt: now.Add(-time.Minute), ExpiresAt: now.Add(time.Minute)}
			})
			result := row.Company.Case
			if arm.name == "last-unaffordable" {
				category, detail := rejectionOf(t, result.Receipt)
				if category != "unaffordable" || detail != "reputation_plan.reputation" || row.Founder != nil {
					t.Fatalf("%s: incorrect whole-plan refusal %s", name, result.Receipt)
				}
				state := replayFixtureStateFromEncoded(t, bundle, result.PreState)
				transition, err := ApplyLoggedExit(state, result.CanonicalPayload, bundle, result.ReplayInputs)
				if err != nil {
					t.Fatal(err)
				}
				if !bytes.Equal(mustEncodeState(t, state), result.PreState) || transition.Decision.NewCompanyState != nil ||
					len(transition.Decision.FounderEvents) != 0 || len(transition.Decision.CompanyEndedEvents) != 0 || len(transition.Decision.CompanyStartedEvents) != 0 {
					t.Fatalf("%s: refusal changed state or emitted events", name)
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
					FounderOutput: replayFounderOutput(transition.Founder, resolved.FounderCarry), FinalCompany: mustEncodeState(t, state), NewCompany: json.RawMessage(`null`),
					FounderEvents: fixtureEvents(transition.Decision.FounderEvents), CompanyEndedEvents: fixtureEvents(transition.Decision.CompanyEndedEvents), CompanyStartedEvents: fixtureEvents(transition.Decision.CompanyStartedEvents)})
				founder := reputationFounderState(t, bundle, 22, now, 0)
				founder.ExitHistory = []save.ExitRecord{{RunID: 1, ExitType: "collapse", OccurredAt: now.Add(-time.Hour)}}
				pre := mustEncodeState(t, founder)
				audit, _, err := buildFounderExitAudit(wire.Command, save.Revision{OwnerID: wire.Command.FounderID, Number: 1, ConstantsHash: bundle.ConstantsHash}, founder, transition.Founder, transition.Decision, bundle)
				if err != nil {
					t.Fatal(err)
				}
				founderInputs, err := save.MarshalFounderReplayInputs(save.FounderReplayCommand{IntentID: wire.Command.IntentID,
					FounderStreamID: "01986666-8c00-4000-8000-000000000001", FounderID: wire.Command.FounderID,
					Revision: 1, FounderLogSeq: 1, ServerTSMS: now.UnixMilli()}, json.RawMessage(audit))
				if err != nil {
					t.Fatal(err)
				}
				founderResult, err := ApplyFounderLogged(founder, result.CanonicalPayload, bundle, founderInputs)
				if err != nil {
					t.Fatal(err)
				}
				category, detail = rejectionOf(t, founderResult.Receipt)
				if category != "unaffordable" || detail != "reputation_plan.reputation" || founderResult.Outcome != save.IntentRejected ||
					founderResult.ResultConstantsHash != bundle.ConstantsHash || !bytes.Equal(mustEncodeState(t, founder), pre) || len(founderResult.Events) != 0 {
					t.Fatalf("%s: Founder rejection changed full state/events/pin", name)
				}
				row.Founder = &reputationCorpusCase{Name: name + "-founder", Bundle: "tree", StateVersion: 22, PreState: pre,
					CanonicalPayload: result.CanonicalPayload, ReplayInputs: founderInputs, Outcome: string(founderResult.Outcome),
					ReceiptJSON: canonicalFixtureJSON(t, founderResult.Receipt), EventsJSON: canonicalFixtureValue(t, fixtureEvents(founderResult.Events)),
					PostStateJSON: canonicalFixtureJSON(t, mustEncodeState(t, founder))}
			} else {
				if result.Outcome != "applied" || row.Founder == nil {
					t.Fatalf("%s: valid Exit did not apply", name)
				}
				founder, err := save.RestoreState([]byte(row.Founder.PostStateJSON), 22, bundle.Economy, economy.ScopeFounder, time.Time{})
				if err != nil {
					t.Fatal(err)
				}
				spent, unlock, factor := int64(0), int64(0), "1e0"
				owned := []string{}
				if len(arm.plan) != 0 {
					spent, unlock, factor = 6, 50000, offer.factor
					owned = []string{"reputation.starter.cash_small", "reputation.starter.generated_beige_tower", "reputation.unlock.p05"}
				}
				if founder.ReputationLevel != offer.level || founder.ReputationSpent != spent || founder.ReputationUnlockPPM != unlock || !slices.Equal(founder.ReputationNodesOwned, owned) {
					t.Fatalf("%s: independent payout/plan accounting differs", name)
				}
				company := replayFixtureStateFromEncoded(t, bundle, result.NewCompany)
				if company.RunSeq != 3 || company.OfferState != nil {
					t.Fatalf("%s: wrong new-run identity", name)
				}
				if len(arm.plan) != 0 {
					cash, ok := company.Ledger.Balance("company.cash")
					if !ok || cash.String() != "1e3" || company.GeneratorProvisioned["generator.beige_tower"] != 5 || company.GeneratorCounts["generator.beige_tower"] != 0 || company.GeneratorPurchasedTotal != 0 {
						t.Fatalf("%s: wrong starter assembly", name)
					}
					if len(result.FounderEvents) != 4 || result.FounderEvents[0].Kind != "founder_advanced" {
						t.Fatalf("%s: wrong Founder event count/order", name)
					}
					for index, id := range good {
						var purchase struct {
							NodeID string `json:"node_id"`
							Cost   int64  `json:"cost"`
							Source string `json:"source"`
						}
						if err := json.Unmarshal(result.FounderEvents[index+1].Payload, &purchase); err != nil {
							t.Fatal(err)
						}
						if result.FounderEvents[index+1].Kind != string(save.EventReputationNodePurchased) || result.FounderEvents[index+1].SchemaVersion != 1 || purchase.NodeID != id || purchase.Cost != int64(index+1) || purchase.Source != "exit_plan" {
							t.Fatalf("%s: wrong ordered purchase %d", name, index)
						}
					}
				}
				resolvedIndex, endedIndex := -1, -1
				for index, event := range result.CompanyEndedEvents {
					if event.Kind == "exit_offer_resolved" {
						resolvedIndex = index
						if canonicalFixtureJSON(t, event.Payload) != `{"offer_id":"01986666-d001-7000-8000-000000000001","resolution":"accepted"}` {
							t.Fatalf("%s: wrong offer resolution", name)
						}
					}
					if event.Kind == "run_ended" {
						endedIndex = index
					}
				}
				if resolvedIndex < 0 || endedIndex <= resolvedIndex {
					t.Fatalf("%s: wrong resolution/end order", name)
				}
				var started struct {
					ReputationTree reputationRunStarted `json:"reputation_tree"`
				}
				if len(result.CompanyStartedEvents) == 0 {
					t.Fatal("missing run_started")
				}
				if err := json.Unmarshal(result.CompanyStartedEvents[0].Payload, &started); err != nil {
					t.Fatal(err)
				}
				starterIDs := []string{}
				if len(arm.plan) != 0 {
					starterIDs = good[1:]
				}
				if result.CompanyStartedEvents[0].SchemaVersion != 2 || started.ReputationTree.BonusFactor != factor || !slices.Equal(started.ReputationTree.AppliedStarterNodeIDs, starterIDs) {
					t.Fatalf("%s: wrong next-run summary", name)
				}
			}
			corpus.Cases = append(corpus.Cases, reputationOfferReplayCase{Name: name, Kind: offer.kind, Level: offer.level, Company: result, Founder: row.Founder})
		}
	}
	return corpus
}

func TestReputationOfferReplayCorpus(t *testing.T) {
	corpus := buildReputationOfferReplayCorpus(t)
	applied, founderArms := 0, 0
	for _, row := range corpus.Cases {
		if row.Company.Outcome == "applied" {
			applied++
		}
		if row.Founder != nil {
			founderArms++
		}
	}
	if len(corpus.Cases) != 10 || applied != 8 || founderArms != 10 {
		t.Fatal("offered-plan population must remain10/8/10")
	}
	for _, offset := range []int{0, 5} {
		absent, empty := corpus.Cases[offset+2].Company, corpus.Cases[offset+3].Company
		if bytes.Equal(absent.CanonicalPayload, empty.CanonicalPayload) {
			t.Fatal("absent and empty must preserve distinct canonical bytes")
		}
		for _, pair := range [][2]string{{absent.ReceiptJSON, empty.ReceiptJSON}, {absent.FounderOutputJSON, empty.FounderOutputJSON}, {absent.FinalCompanyJSON, empty.FinalCompanyJSON}, {absent.NewCompanyJSON, empty.NewCompanyJSON}, {absent.FounderEventsJSON, empty.FounderEventsJSON}, {absent.CompanyEndedJSON, empty.CompanyEndedJSON}, {absent.CompanyStartedJSON, empty.CompanyStartedJSON}} {
			if pair[0] != pair[1] {
				t.Fatal("empty plan changed absent-plan semantic output")
			}
		}
	}
	if t.Failed() {
		return
	} // Never regenerate away a failed independent control.
	encoded, err := json.MarshalIndent(corpus, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	encoded = append(encoded, '\n')
	if *updateReplayFixture {
		if err := os.WriteFile(reputationOfferCorpusPath, encoded, 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	pinned, err := os.ReadFile(reputationOfferCorpusPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(pinned, encoded) {
		t.Fatal("offered-plan corpus differs from executed Go transitions; explicit generation and review required")
	}
}

func TestReputationOfferReplayRefusesCopiedEvidence(t *testing.T) {
	corpus := buildReputationOfferReplayCorpus(t)
	bundle := reputationContentBundle(t)
	for _, row := range corpus.Cases {
		if row.Company.Outcome == "rejected" {
			t.Run(row.Name+"-founder-delta", func(t *testing.T) {
				var inputs map[string]any
				if err := json.Unmarshal(row.Founder.ReplayInputs, &inputs); err != nil {
					t.Fatal(err)
				}
				inputs["resolved"].(map[string]any)["reputation_delta"] = float64(1)
				tampered, err := json.Marshal(inputs)
				if err != nil {
					t.Fatal(err)
				}
				founder, err := save.RestoreState(row.Founder.PreState, 22, bundle.Economy, economy.ScopeFounder, time.Time{})
				if err != nil {
					t.Fatal(err)
				}
				if _, err := ApplyFounderLogged(founder, row.Founder.CanonicalPayload, bundle, tampered); err == nil {
					t.Fatal("rejected Founder arm with credited delta replayed")
				}
			})
			continue
		}
		if row.Founder == nil || !bytes.Contains(row.Company.CanonicalPayload, []byte(`"reputation.unlock.p05"`)) {
			continue
		}
		t.Run(row.Name, func(t *testing.T) {
			var inputs map[string]any
			if err := json.Unmarshal(row.Founder.ReplayInputs, &inputs); err != nil {
				t.Fatal(err)
			}
			purchases := inputs["resolved"].(map[string]any)["reputation_purchases"].([]any)
			purchases[0].(map[string]any)["resolved_cost"] = float64(2)
			tampered, err := json.Marshal(inputs)
			if err != nil {
				t.Fatal(err)
			}
			founder, err := save.RestoreState(row.Founder.PreState, 22, bundle.Economy, economy.ScopeFounder, time.Time{})
			if err != nil {
				t.Fatal(err)
			}
			if _, err := ApplyFounderLogged(founder, row.Founder.CanonicalPayload, bundle, tampered); err == nil {
				t.Fatal("copied wrong cost replayed")
			}
			if err := json.Unmarshal(row.Company.ReplayInputs, &inputs); err != nil {
				t.Fatal(err)
			}
			terms := inputs["resolved"].(map[string]any)["selected_terms"].(map[string]any)
			terms["payout_preview"].(map[string]any)["reputation_delta"] = float64(row.Level + 1)
			tampered, err = json.Marshal(inputs)
			if err != nil {
				t.Fatal(err)
			}
			company := replayFixtureStateFromEncoded(t, bundle, row.Company.PreState)
			if _, err := ApplyLoggedExit(company, row.Company.CanonicalPayload, bundle, tampered); err == nil {
				t.Fatal("copied mismatched stored promise replayed")
			}
			var payload map[string]any
			if err := json.Unmarshal(row.Company.CanonicalPayload, &payload); err != nil {
				t.Fatal(err)
			}
			plan := payload["reputation_plan"].([]any)
			plan[1], plan[2] = plan[2], plan[1]
			wire, err := parseReplayInputs(row.Company.ReplayInputs)
			if err != nil {
				t.Fatal(err)
			}
			// Stored canonical requests omit intent_id; ParseIntent consumes the
			// external request and strips that transport identity again.
			payload["intent_id"] = wire.Command.IntentID
			payloadBytes, err := json.Marshal(payload)
			if err != nil {
				t.Fatal(err)
			}
			request, err := ParseIntent(payloadBytes)
			if err != nil {
				t.Fatal(err)
			}
			company = replayFixtureStateFromEncoded(t, bundle, row.Company.PreState)
			transition, err := ApplyLoggedExit(company, request.CanonicalPayload, bundle, row.Company.ReplayInputs)
			if err != nil {
				t.Fatal(err)
			}
			category, detail := rejectionOf(t, transition.Decision.Receipt)
			if category != "not_eligible" || detail != "reputation_plan.requires" || !bytes.Equal(mustEncodeState(t, company), row.Company.PreState) || transition.Decision.NewCompanyState != nil || len(transition.Decision.FounderEvents) != 0 || len(transition.Decision.CompanyEndedEvents) != 0 || len(transition.Decision.CompanyStartedEvents) != 0 {
				t.Fatal("reordered copied prerequisite plan did not refuse without mutation")
			}
		})
	}
}

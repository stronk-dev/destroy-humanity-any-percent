package production

import (
	"bytes"
	"encoding/json"
	"slices"
	"testing"
	"time"

	"cloud-clicker/server/economy"
	"cloud-clicker/server/meters"
	"cloud-clicker/server/reputation"
	"cloud-clicker/server/save"
)

// Retune only fixture bytes. Every ID remains present, and the new price
// cannot become authority for the cost recorded by a historical purchase.
func reputationRetunedEpoch(t *testing.T, current CatalogBundle, unlock int64) CatalogBundle {
	t.Helper()
	if unlock == 50_000 {
		return current
	}
	next := current
	next.Next, next.Artifacts = nil, cloneArtifactMap(current.Artifacts)
	var root map[string]any
	if err := json.Unmarshal(next.Artifacts["reputation_tree"], &root); err != nil {
		t.Fatal(err)
	}
	for _, raw := range root["nodes"].([]any) {
		node := raw.(map[string]any)
		if node["node_id"] == "reputation.unlock.p05" {
			node["unlock_ppm"], node["cost"] = unlock, 3
		}
	}
	next.Artifacts["reputation_tree"] = reputationStarterJSON(t, root)
	var err error
	next.ReputationTree, err = reputation.LoadTree(next.Artifacts["reputation_tree"], reputation.Declarations{
		Economy: next.Economy, Curriculum: next.Curriculum, CopyKeys: reputationStarterCopyKeys(),
	})
	if err != nil {
		t.Fatal(err)
	}
	next.ConstantsHash, err = save.ConstantsHashArtifacts(next.Artifacts)
	if err != nil || !next.valid(next.ConstantsHash) || next.ConstantsHash == current.ConstantsHash {
		t.Fatalf("invalid retuned fixture: %v", err)
	}
	return next
}

func TestReputationOwnedEffectRebindsAtEpochBoundary(t *testing.T) {
	current := reputationContentBundle(t)
	now := time.Date(2026, 10, 6, 12, 0, 0, 0, time.UTC)
	for _, row := range []struct {
		name   string
		unlock int64
		factor string
	}{{"unchanged", 50_000, "1.0055e0"}, {"higher", 60_000, "1.0066e0"}, {"lower", 40_000, "1.0044e0"}} {
		t.Run(row.name, func(t *testing.T) {
			next := reputationRetunedEpoch(t, current, row.unlock)
			source := current
			source.Next = &next
			initial := reputationFounderState(t, current, 22, now, 11)
			initial.ReputationSpent, initial.ReputationUnlockPPM = 4, 50_000
			initial.ReputationNodesOwned = []string{"reputation.retired.unknown", "reputation.unlock.p05"}
			initial.AgeMS, initial.RouteKnowledgeBalance = 12_345, 9
			if err := current.ValidateFoundationState(initial); err != nil {
				t.Fatal(err)
			}
			before := mustEncodeState(t, initial)
			restore := func() *save.State {
				state, err := save.RestoreState(before, 22, current.Economy, economy.ScopeFounder, time.Time{})
				if err != nil {
					t.Fatal(err)
				}
				return state
			}
			results := map[string][]byte{}
			for _, arm := range []string{"live", "founder_replay"} {
				t.Run(arm, func(t *testing.T) {
					state := restore()
					if arm == "live" {
						company := replayFixtureState(t, current.Economy, now.Add(-time.Hour))
						_, company.WireVersion = current.versionFloors()
						company.MeterBands = nil
						meterState, err := meters.NewRunState(current.Meters, 0)
						if err != nil {
							t.Fatal(err)
						}
						company.MeterValues, company.MeterDecayRemainders, company.MeterInputRemainders = meterState.Values, meterState.DecayRemainders, meterState.InputRemainders
						company.AchievementsEarnedRun = map[string]bool{}
						if err := current.ValidateFoundationState(company); err != nil {
							t.Fatal(err)
						}
						created := foundationScopeState(t, next.Economy, economy.ScopeCompany)
						created.RunStartedAt = now
						if err := settleAndActivateFoundations(source, next, state, company, created); err != nil {
							t.Fatalf("live boundary rejected an append-only effect retune: %v", err)
						}
						state.ExitHistory = append(state.ExitHistory, save.ExitRecord{RunID: 1, ExitType: "collapse", OccurredAt: now})
					} else {
						command := save.FounderReplayCommand{IntentID: "01986666-8d01-7000-8000-000000000001",
							FounderStreamID: "01986666-8c00-4000-8000-000000000001", FounderID: "01986666-8d00-7000-8000-000000000001",
							Revision: 1, FounderLogSeq: 1, ServerTSMS: now.UnixMilli()}
						resolved := founderExitResolvedWire{Kind: founderExitResolvedKind, Outcome: string(save.IntentApplied),
							CompanyStreamID: "01986666-8e00-7000-8000-000000000001", RunSeq: 1, RunLogSeq: 1,
							ResultConstantsHash: next.ConstantsHash, AgeMSBefore: 12_345, AgeMSAfter: 12_345,
							AddedNetworkSlots: []save.NetworkSlot{}, AddedLedgerFactKinds: []string{}, AddedLifetimeAchievements: []string{},
							ExitRecord: &founderExitRecordWire{RunID: 1, ExitType: "collapse", OccurredAtMS: now.UnixMilli()}, ResultFounderWireVersion: 22}
						inputs, err := save.MarshalFounderReplayInputs(command, resolved)
						if err != nil {
							t.Fatal(err)
						}
						request, err := ParseIntent([]byte(`{"intent_id":"01986666-8d01-7000-8000-000000000001","kind":"wind_down","expected_revision":1,"expected_founder_revision":1}`))
						if err != nil || request.InvalidDetail != "" {
							t.Fatalf("invalid fixture request: %v/%s", err, request.InvalidDetail)
						}
						result, err := ApplyFounderLogged(state, request.CanonicalPayload, source, inputs)
						if err != nil || result.Outcome != save.IntentApplied || result.ResultConstantsHash != next.ConstantsHash {
							t.Fatalf("Founder rejected an append-only effect retune: %v/%s", err, result.Outcome)
						}
					}
					if state.ReputationLevel != 11 || state.ReputationSpent != 4 || state.ReputationUnlockPPM != row.unlock ||
						!slices.Equal(state.ReputationNodesOwned, initial.ReputationNodesOwned) || state.AgeMS != 12_345 || state.RouteKnowledgeBalance != 9 {
						t.Fatal("epoch rebinding changed historical accounting, ownership or metadata")
					}
					if available, err := reputation.Available(state.ReputationLevel, state.ReputationSpent); err != nil || available != 7 {
						t.Fatalf("available=%d/%v", available, err)
					}
					if err := next.ValidateFoundationState(state); err != nil {
						t.Fatal(err)
					}
					frozen, err := FrozenFounderContributions(next, state)
					if err != nil {
						t.Fatal(err)
					}
					found := false
					for _, contribution := range frozen {
						if contribution.SourceID == "reputation.founder_bonus" {
							found = true
							if contribution.Factor != row.factor {
								t.Fatalf("next frozen factor=%s want=%s", contribution.Factor, row.factor)
							}
						}
					}
					if !found {
						t.Fatal("missing next-run Reputation row")
					}
					encoded := mustEncodeState(t, state)
					decoded, err := save.RestoreState(encoded, 22, next.Economy, economy.ScopeFounder, time.Time{})
					if err != nil {
						t.Fatal(err)
					}
					if err := next.ValidateFoundationState(decoded); err != nil {
						t.Fatal(err)
					}
					results[arm] = encoded
				})
			}
			if !t.Failed() && !bytes.Equal(results["live"], results["founder_replay"]) {
				t.Fatal("live and Founder replay full state bytes differ")
			}
		})
	}
}

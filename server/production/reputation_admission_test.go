package production

import (
	"bytes"
	"errors"
	"testing"
	"time"

	"cloud-clicker/server/save"
)

func reputationAdmissionCommand(t *testing.T, state *save.State, missingNode bool) (IntentRequest, []byte) {
	t.Helper()
	body := `{"intent_id":"01986666-5e00-7000-8000-000000000001","kind":"purchase_reputation_node","expected_revision":1`
	if !missingNode {
		body += `,"node_id":"reputation.starter.cash_small"`
	}
	request, err := ParseIntent([]byte(body + "}"))
	if err != nil || (request.InvalidDetail != "") != missingNode {
		t.Fatalf("unexpected request validity: %v/%s", err, request.InvalidDetail)
	}
	var resolved any = founderInvalidResolved{Kind: "invalid", Detail: request.InvalidDetail}
	if !missingNode {
		// Independently supplied audit evidence must not bypass the public
		// input boundary merely because the live resolver refuses corruption.
		resolved = founderReputationPurchaseResolved{Kind: IntentPurchaseReputationNode, NodeID: request.ReputationNodeID,
			ResolvedCost: 2, ReputationLevel: state.ReputationLevel, ReputationSpentBefore: state.ReputationSpent,
			OwnedBefore: append([]string{}, state.ReputationNodesOwned...)}
	}
	command := save.FounderReplayCommand{IntentID: request.IntentID, FounderStreamID: "01986666-5c00-4000-8000-000000000001",
		FounderID: "01986666-5d00-7000-8000-000000000001", Revision: 1, FounderLogSeq: 1,
		ServerTSMS: time.Date(2026, 10, 6, 12, 0, 0, 0, time.UTC).UnixMilli()}
	inputs, err := save.MarshalFounderReplayInputs(command, resolved)
	if err != nil {
		t.Fatal(err)
	}
	return request, inputs
}

func TestReputationPinnedInputAdmission(t *testing.T) {
	tree := reputationContentBundle(t)
	now := time.Date(2026, 10, 6, 12, 0, 0, 0, time.UTC)
	for _, source := range []struct {
		name   string
		owned  []string
		spent  int64
		mirror int64
	}{{"too_high", []string{}, 0, 50_000}, {"too_low", []string{"reputation.unlock.p05"}, 1, 0}} {
		for _, arm := range []string{"live_resolver", "public_purchase", "public_recorded_invalid"} {
			t.Run(source.name+"/"+arm, func(t *testing.T) {
				state := reputationFounderState(t, tree, 22, now, 10)
				state.ReputationNodesOwned, state.ReputationSpent, state.ReputationUnlockPPM = source.owned, source.spent, source.mirror
				before := mustEncodeState(t, state) // Structurally legal, but pinned-invalid.
				if !errors.Is(tree.ValidateFoundationState(state), ErrInvalidEngineState) {
					t.Fatal("counterfact is not invalid under the pinned tree")
				}
				if arm == "live_resolver" {
					if _, err := resolveReputationPurchase(tree, state, "reputation.starter.cash_small"); !errors.Is(err, ErrInvalidEngineState) {
						t.Fatalf("live resolver admitted false mirror: %v", err)
					}
				} else {
					request, inputs := reputationAdmissionCommand(t, state, arm == "public_recorded_invalid")
					result, err := ApplyFounderLogged(state, request.CanonicalPayload, tree, inputs)
					if !errors.Is(err, ErrInvalidEngineState) || len(result.Receipt) != 0 || len(result.Events) != 0 {
						t.Fatalf("public boundary admitted false mirror: outcome=%s receipt=%s events=%d err=%v", result.Outcome, result.Receipt, len(result.Events), err)
					}
				}
				if !bytes.Equal(before, mustEncodeState(t, state)) {
					t.Fatal("corruption was repaired or command mutated the pre-state")
				}
			})
		}
	}
}

func TestReputationPinnedOutputAdmissionRollsBack(t *testing.T) {
	tree := reputationContentBundle(t)
	now := time.Date(2026, 10, 6, 12, 0, 0, 0, time.UTC)
	state := reputationFounderState(t, tree, 22, now, 10)
	before := mustEncodeState(t, state)
	request, inputs := reputationAdmissionCommand(t, state, false)
	if founderTransitionTestArm != nil {
		t.Fatal("another test left the transition hook active")
	}
	founderTransitionTestArm = func(candidate *save.State) { candidate.ReputationUnlockPPM = 50_000 }
	defer func() { founderTransitionTestArm = nil }()
	result, err := ApplyFounderLogged(state, request.CanonicalPayload, tree, inputs)
	if !errors.Is(err, ErrInvalidEngineState) || len(result.Receipt) != 0 || len(result.Events) != 0 || !bytes.Equal(before, mustEncodeState(t, state)) {
		t.Fatalf("invalid successful output escaped or did not roll back: outcome=%s receipt=%s events=%d err=%v", result.Outcome, result.Receipt, len(result.Events), err)
	}
}

func TestReputationPinnedAdmissionPreservesValidAndInactiveStates(t *testing.T) {
	tree := reputationContentBundle(t)
	now := time.Date(2026, 10, 6, 12, 0, 0, 0, time.UTC)
	for _, row := range []struct {
		name   string
		owned  []string
		spent  int64
		mirror int64
	}{{"empty", []string{}, 0, 0}, {"known_unlock", []string{"reputation.unlock.p05"}, 1, 50_000},
		{"retired_unknown", []string{"reputation.retired.old"}, 4, 0}} {
		t.Run(row.name, func(t *testing.T) {
			state := reputationFounderState(t, tree, 22, now, 10)
			state.ReputationNodesOwned, state.ReputationSpent, state.ReputationUnlockPPM = row.owned, row.spent, row.mirror
			if err := tree.ValidateFoundationState(state); err != nil {
				t.Fatal(err)
			}
			if resolved, err := resolveReputationPurchase(tree, state, "reputation.starter.cash_small"); err != nil || resolved.ResolvedCost != 2 {
				t.Fatalf("valid live purchase: %v", err)
			}
			request, inputs := reputationAdmissionCommand(t, state, false)
			result, err := ApplyFounderLogged(state, request.CanonicalPayload, tree, inputs)
			if err != nil || result.Outcome != save.IntentApplied || state.ReputationLevel != 10 || state.ReputationSpent != row.spent+2 || state.ReputationUnlockPPM != row.mirror {
				t.Fatalf("valid/retired accounting changed: outcome=%s err=%v", result.Outcome, err)
			}
			if err := tree.ValidateFoundationState(state); err != nil {
				t.Fatal(err)
			}
		})
	}
	t.Run("pre_v22_inactive", func(t *testing.T) {
		state := reputationFounderState(t, activeContentBundle(t), 21, now, 10)
		before := mustEncodeState(t, state)
		request, _ := reputationAdmissionCommand(t, state, false)
		resolved, err := resolveReputationPurchase(tree, state, request.ReputationNodeID)
		if err != nil || resolved.ResolvedCost != 0 {
			t.Fatalf("inactive resolver: %v", err)
		}
		command := save.FounderReplayCommand{IntentID: request.IntentID, FounderStreamID: "01986666-5c00-4000-8000-000000000001",
			FounderID: "01986666-5d00-7000-8000-000000000001", Revision: 1, FounderLogSeq: 1, ServerTSMS: now.UnixMilli()}
		inputs, err := save.MarshalFounderReplayInputs(command, resolved)
		if err != nil {
			t.Fatal(err)
		}
		result, err := ApplyFounderLogged(state, request.CanonicalPayload, tree, inputs)
		if err != nil || result.Outcome != save.IntentRejected || len(result.Events) != 0 || !bytes.Equal(before, mustEncodeState(t, state)) {
			t.Fatalf("inactive taxonomy/state changed: %v", err)
		}
		category, detail := rejectionOf(t, result.Receipt)
		if category != "not_eligible" || detail != "reputation_tree_inactive" {
			t.Fatalf("inactive rejection=%s/%s", category, detail)
		}
	})
}

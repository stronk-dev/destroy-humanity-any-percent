package production

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"

	"cloud-clicker/server/economy"
	"cloud-clicker/server/reputation"
	"cloud-clicker/server/save"
)

// founderReputationPurchaseResolved is R5's frozen resolved-input arm. Replay
// recomputes every field from the pinned tree and Founder state.
type founderReputationPurchaseResolved struct {
	Kind                  string   `json:"kind"`
	NodeID                string   `json:"node_id"`
	ResolvedCost          int64    `json:"resolved_cost"`
	ReputationLevel       int64    `json:"reputation_level"`
	ReputationSpentBefore int64    `json:"reputation_spent_before"`
	OwnedBefore           []string `json:"owned_before"`
}

// reputationTreeActive is R5 step 3: the pinned Founder bundle carries the
// tree and the Founder has activated v22.
func reputationTreeActive(catalogs CatalogBundle, state *save.State) bool {
	return catalogs.ReputationTree != nil && save.VersionForState(state) >= 22
}

// resolveReputationPurchase derives the resolved inputs for a Founder state.
// resolved_cost is the node cost only when the purchase would apply, else 0.
func resolveReputationPurchase(catalogs CatalogBundle, state *save.State, nodeID string) (founderReputationPurchaseResolved, error) {
	owned := append([]string{}, state.ReputationNodesOwned...)
	resolved := founderReputationPurchaseResolved{Kind: IntentPurchaseReputationNode, NodeID: nodeID,
		ReputationLevel: state.ReputationLevel, ReputationSpentBefore: state.ReputationSpent, OwnedBefore: owned}
	if !reputationTreeActive(catalogs, state) {
		return resolved, nil
	}
	applied, rejection, err := catalogs.ReputationTree.Purchase(state.ReputationLevel, state.ReputationSpent, owned, nodeID)
	if err != nil {
		return founderReputationPurchaseResolved{}, err
	}
	if rejection == nil {
		resolved.ResolvedCost = applied.Node.Cost
	}
	return resolved, nil
}

func (s *Service) handleFounderReputation(ctx context.Context, companyStreamID string, request IntentRequest) (HandleResult, error) {
	if s.replayCatalogs == nil {
		return HandleResult{}, fmt.Errorf("%w: Founder replay runtime unavailable", ErrInvalidIntent)
	}
	founder, err := s.store.LoadSiblingLatest(ctx, companyStreamID, economy.ScopeFounder)
	if err != nil {
		return HandleResult{}, err
	}
	result, err := s.store.ApplyFounderLogged(ctx, founder.Revision.StreamID, request.ExpectedRevision,
		request.IntentID, request.RequestHash, request.CanonicalPayload,
		func(state *save.State, revision save.Revision, command save.FounderReplayCommand) (save.IntentDecision, json.RawMessage, error) {
			bundle, ok := s.replayCatalogs.ResolveReplayCatalogs(revision.ConstantsHash)
			if !ok {
				return save.IntentDecision{}, nil, fmt.Errorf("%w: replay catalog bundle unavailable", ErrInvalidIntent)
			}
			var resolved any = founderInvalidResolved{Kind: "invalid", Detail: request.InvalidDetail}
			if request.InvalidDetail == "" {
				purchase, resolveErr := resolveReputationPurchase(bundle, state, request.ReputationNodeID)
				if resolveErr != nil {
					return save.IntentDecision{}, nil, resolveErr
				}
				resolved = purchase
			}
			replayInputs, marshalErr := marshalFounderReplayInputs(command, resolved)
			if marshalErr != nil {
				return save.IntentDecision{}, nil, marshalErr
			}
			transition, transitionErr := ApplyFounderLogged(state, request.CanonicalPayload, bundle, replayInputs)
			return save.IntentDecision{Outcome: transition.Outcome, Receipt: transition.Receipt,
				Events: transition.Events}, replayInputs, transitionErr
		})
	if err != nil {
		var conflict *save.RevisionConflict
		switch {
		case errors.As(err, &conflict):
			return HandleResult{Receipt: marshalRejection(request.IntentID, conflict.Current,
				"revision_conflict", "expected_revision")}, nil
		case errors.Is(err, save.ErrIdempotencyConflict):
			current := request.ExpectedRevision
			if loaded, loadErr := s.store.LoadLatest(ctx, founder.Revision.StreamID); loadErr == nil {
				current = loaded.Revision.Number
			}
			return HandleResult{Receipt: marshalRejection(request.IntentID, current,
				"idempotency_conflict", request.IntentID)}, nil
		default:
			return HandleResult{}, err
		}
	}
	if err := s.projectCommittedEvents(ctx, result.Events); err != nil {
		return HandleResult{}, err
	}
	return HandleResult{Receipt: result.Receipt, Replay: result.Replay}, nil
}

// applyFounderReputationPurchaseResolved is R5's replay arm (live and
// history share it): steps 3–8, recorded rejections, the applied receipt, and
// reputation_node_purchased.v1 with source "direct".
func applyFounderReputationPurchaseResolved(state *save.State, request IntentRequest, revision save.Revision,
	catalogs CatalogBundle, resolvedJSON json.RawMessage) (FounderLoggedTransition, error) {
	if request.Kind != IntentPurchaseReputationNode || request.ExpectedRevision != revision.Number || request.InvalidDetail != "" {
		return FounderLoggedTransition{}, fmt.Errorf("%w: Reputation purchase command", ErrInvalidReplayInputs)
	}
	var resolved founderReputationPurchaseResolved
	if err := decodeReplayStrict(resolvedJSON, &resolved); err != nil || resolved.OwnedBefore == nil {
		return FounderLoggedTransition{}, fmt.Errorf("%w: Reputation purchase inputs", ErrInvalidReplayInputs)
	}
	expected, err := resolveReputationPurchase(catalogs, state, request.ReputationNodeID)
	if err != nil || resolved.Kind != expected.Kind || resolved.NodeID != expected.NodeID || resolved.ResolvedCost != expected.ResolvedCost ||
		resolved.ReputationLevel != expected.ReputationLevel || resolved.ReputationSpentBefore != expected.ReputationSpentBefore ||
		!slices.Equal(resolved.OwnedBefore, expected.OwnedBefore) {
		return FounderLoggedTransition{}, fmt.Errorf("%w: Reputation purchase resolution", ErrInvalidReplayInputs)
	}
	if !reputationTreeActive(catalogs, state) {
		decision, decisionErr := rejectedDecision(request, revision.Number, "not_eligible", "reputation_tree_inactive")
		return founderDecisionTransition(state, decision, catalogs.ConstantsHash, decisionErr)
	}
	availableBefore, _ := reputation.Available(state.ReputationLevel, state.ReputationSpent)
	applied, rejection, err := catalogs.ReputationTree.Purchase(state.ReputationLevel, state.ReputationSpent, state.ReputationNodesOwned, request.ReputationNodeID)
	if err != nil {
		return FounderLoggedTransition{}, fmt.Errorf("%w: Reputation purchase state", ErrInvalidReplayInputs)
	}
	if rejection != nil {
		decision, decisionErr := rejectedDecision(request, revision.Number, rejection.Category, rejection.Detail)
		return founderDecisionTransition(state, decision, catalogs.ConstantsHash, decisionErr)
	}
	spentBefore := state.ReputationSpent
	state.ReputationSpent, state.ReputationNodesOwned, state.ReputationUnlockPPM = applied.SpentAfter, applied.OwnedAfter, applied.UnlockPPMAfter
	receipt, err := json.Marshal(map[string]any{"intent_id": request.IntentID, "outcome": string(save.IntentApplied),
		"founder_revision": revision.Number + 1, "fiscal_sweep": nil, "node_id": applied.Node.NodeID,
		"resolved_cost": applied.Node.Cost, "reputation_available_before": availableBefore,
		"reputation_available_after": availableBefore - applied.Node.Cost, "reputation_unlock_ppm_after": applied.UnlockPPMAfter,
		"effective_from": "next_run"})
	if err != nil {
		return FounderLoggedTransition{}, err
	}
	payload, err := json.Marshal(map[string]any{"node_id": applied.Node.NodeID, "node_kind": applied.Node.Kind, "cost": applied.Node.Cost,
		"reputation_level": state.ReputationLevel, "reputation_spent_before": spentBefore,
		"reputation_spent_after": applied.SpentAfter, "unlock_ppm_after": applied.UnlockPPMAfter, "source": "direct"})
	if err != nil {
		return FounderLoggedTransition{}, err
	}
	return FounderLoggedTransition{State: state, Outcome: save.IntentApplied, Receipt: receipt,
		Events: []save.EventWrite{{Kind: save.EventReputationNodePurchased, SchemaVersion: 1,
			IntentID: request.IntentID, Payload: payload}}, ResultConstantsHash: catalogs.ConstantsHash}, nil
}

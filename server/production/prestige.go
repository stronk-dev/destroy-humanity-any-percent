package production

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"time"

	"cloud-clicker/server/decimal"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/multiplier"
	prestigecore "cloud-clicker/server/prestige"
	"cloud-clicker/server/save"
)

var adjacentGatePattern = regexp.MustCompile(`^gate\.t([0-9]+)_to_t([0-9]+)$`)

func requireFounderCatalogCoherence(founder, company save.Revision) error {
	if founder.ConstantsHash == "" || founder.ConstantsHash != company.ConstantsHash {
		return ErrInvalidEngineState
	}
	return nil
}

func (s *Service) declineExitOffer(request IntentRequest, state *save.State, catalog *economy.Catalog, revision save.Revision, mode EvaluationMode, now time.Time, contributions []multiplier.Contribution, hook AccrualHook) (save.IntentDecision, error) {
	if state == nil || state.Ledger == nil || state.Ledger.Scope() != economy.ScopeCompany {
		return save.IntentDecision{}, ErrInvalidEngineState
	}
	if state.OfferState == nil || state.OfferState.OfferID != request.OfferID {
		return rejectedDecision(request, revision.Number, "not_eligible", "exit_offer")
	}
	if !state.OfferState.ExpiresAt.After(save.CanonicalServerTime(now)) {
		return rejectedDecision(request, revision.Number, "offer_expired", request.OfferID)
	}
	before := state.Ledger.Snapshot()
	result, err := Evaluate(state, catalog, now, mode, contributions)
	if err != nil {
		return save.IntentDecision{}, err
	}
	events, err := runAccrualHook(hook, request.IntentID, state, catalog, revision, result, contributions)
	if err != nil {
		return save.IntentDecision{}, err
	}
	payload, _ := json.Marshal(map[string]any{"offer_id": request.OfferID, "run_seq": state.RunSeq})
	state.OfferState = nil
	events = append(events, save.EventWrite{Kind: save.EventExitOfferDeclined, SchemaVersion: 1, IntentID: request.IntentID, Payload: payload})
	return appliedDecision(request, state, revision.Number+1, 1, before, events, nil)
}

func afterPrestigeTransitionResolved(policy *prestigecore.Policy, request IntentRequest, state *save.State, revision save.Revision, now time.Time, decision *save.IntentDecision, founder *save.State, declinedOffers int64) error {
	if state == nil || decision == nil {
		return ErrInvalidEngineState
	}
	now = save.CanonicalServerTime(now)
	if state.OfferState != nil && !state.OfferState.ExpiresAt.After(now) {
		payload, _ := json.Marshal(map[string]string{"offer_id": state.OfferState.OfferID})
		decision.Events = append(decision.Events, save.EventWrite{Kind: save.EventExitOfferExpired, SchemaVersion: 1, IntentID: request.IntentID, Payload: payload})
		state.OfferState = nil
	}
	if request.Kind == IntentCrossGate && state.OfferState == nil {
		if founder == nil || policy == nil {
			return ErrInvalidEngineState
		}
		if len(founder.ExitHistory) == 0 {
			return refreshAppliedSnapshot(decision, state)
		}
		if state.Tier < 0 || state.Tier >= int64(len(policy.SpawnGatePPM)) {
			return ErrInvalidEngineState
		}
		draw, exitType, driftUp := prestigecore.OfferDraws(revision.OwnerID, state.RunSeq, state.Tier, declinedOffers)
		if policy.SpawnGatePPM[state.Tier] > draw {
			terms, err := prestigecore.ComputeTerms(state, founder, policy, exitType)
			if err != nil {
				return err
			}
			modifier := prestigecore.MarketModifierPPM(declinedOffers, policy.DeclineDriftPPM, driftUp)
			terms = prestigecore.ApplyMarketModifier(terms, modifier)
			termsJSON, _ := json.Marshal(prestigecore.StoredOfferTerms{PayoutPreview: terms, MarketModifierPPM: modifier})
			offerID := prestigecore.OfferID(revision.OwnerID, state.RunSeq, state.Tier, declinedOffers, now)
			expires := now.Add(time.Duration(policy.OfferDurationMS) * time.Millisecond)
			state.OfferState = &save.ExitOfferState{OfferID: offerID, ExitType: exitType, TermsJSON: termsJSON, SpawnedAt: now, ExpiresAt: expires}
			payload, _ := json.Marshal(map[string]any{"offer_id": offerID, "exit_type": exitType, "expires_at_ms": expires.UnixMilli(), "payout_preview": terms})
			decision.Events = append(decision.Events, save.EventWrite{Kind: save.EventExitOfferSpawned, SchemaVersion: 1, IntentID: request.IntentID, Payload: payload})
		}
	}
	return refreshAppliedSnapshot(decision, state)
}

func refreshAppliedSnapshot(decision *save.IntentDecision, state *save.State) error {
	var receipt map[string]json.RawMessage
	if err := json.Unmarshal(decision.Receipt, &receipt); err != nil {
		return err
	}
	snapshot, err := json.Marshal(wireSnapshot(state))
	if err != nil {
		return err
	}
	receipt["snapshot"] = snapshot
	decision.Receipt, err = json.Marshal(receipt)
	return err
}

func tierForGate(gateID string) (int64, bool) {
	match := adjacentGatePattern.FindStringSubmatch(gateID)
	if match == nil {
		return 0, false
	}
	var from, to int64
	if _, err := fmt.Sscanf(match[1], "%d", &from); err != nil {
		return 0, false
	}
	if _, err := fmt.Sscanf(match[2], "%d", &to); err != nil || to != from+1 || to > 9 {
		return 0, false
	}
	return to, true
}

func (s *Service) handleExit(ctx context.Context, streamID string, mode EvaluationMode, now time.Time, request IntentRequest) (HandleResult, error) {
	if s.prestigePolicies == nil {
		return HandleResult{}, fmt.Errorf("%w: prestige runtime unavailable", ErrInvalidIntent)
	}
	executedRoutes, err := s.executedRoutesAt(ctx, streamID, request.ExpectedRevision)
	if err != nil {
		return HandleResult{}, err
	}
	result, err := s.store.ApplyExitTransactionLogged(ctx, streamID, request.ExpectedRevision, request.ExpectedFounderRevision, request.IntentID, request.RequestHash, request.CanonicalPayload,
		func(founder *save.State, founderRevision save.Revision, company *save.State, companyRevision save.Revision, command save.ReplayCommand) (save.ExitDecision, json.RawMessage, error) {
			return s.applyLoggedExit(ctx, request, founder, founderRevision, company, companyRevision, command, mode, now, executedRoutes)
		}, nil)
	if err != nil {
		return s.exitErrorReceipt(ctx, streamID, request, err)
	}
	return s.finishExitResult(ctx, result)
}

func (s *Service) scriptedExitDue(ctx context.Context, companyStreamID string, now time.Time) (bool, int64, error) {
	company, err := s.store.LoadLatest(ctx, companyStreamID)
	if err != nil {
		return false, 0, err
	}
	founder, err := s.store.LoadSiblingLatest(ctx, companyStreamID, economy.ScopeFounder)
	if err != nil {
		return false, 0, err
	}
	if len(founder.State.ExitHistory) != 0 || company.State.RunStartedAt.IsZero() {
		return false, founder.Revision.Number, nil
	}
	copyState := *company.State
	copyState.OfflineSpans = append([]save.OfflineSpan(nil), company.State.OfflineSpans...)
	effectiveNow := save.CanonicalServerTime(now)
	if effectiveNow.After(copyState.EvaluatedThrough) {
		policy, ok := s.prestigePolicies.ResolvePrestige(company.Revision.ConstantsHash)
		if !ok {
			return false, 0, ErrInvalidEngineState
		}
		if err := prestigecore.RecordOfflineSpan(&copyState, copyState.EvaluatedThrough, effectiveNow, policy.CatchupCeilingMS); err != nil {
			return false, 0, err
		}
	}
	attended, err := prestigecore.AttendedMS(&copyState, effectiveNow)
	return attended >= 900_000, founder.Revision.Number, err
}

func (s *Service) handleScriptedCrossGateExit(ctx context.Context, streamID string, mode EvaluationMode, now time.Time, request IntentRequest, expectedFounderRevision int64) (HandleResult, error) {
	executedRoutes, err := s.executedRoutesAt(ctx, streamID, request.ExpectedRevision)
	if err != nil {
		return HandleResult{}, err
	}
	result, err := s.store.ApplyExitTransactionLogged(ctx, streamID, request.ExpectedRevision, expectedFounderRevision, request.IntentID, request.RequestHash, request.CanonicalPayload,
		func(founder *save.State, founderRevision save.Revision, company *save.State, companyRevision save.Revision, command save.ReplayCommand) (save.ExitDecision, json.RawMessage, error) {
			return s.applyLoggedExit(ctx, request, founder, founderRevision, company, companyRevision, command, mode, now, executedRoutes)
		}, nil)
	if err != nil {
		return s.exitErrorReceipt(ctx, streamID, request, err)
	}
	return s.finishExitResult(ctx, result)
}

func (s *Service) finishExit(request IntentRequest, founder *save.State, founderRevision save.Revision, company *save.State, companyRevision save.Revision, now time.Time, exitType string, terms prestigecore.Terms, endedPrefix []save.EventWrite, executedRoutes []string) (save.ExitDecision, error) {
	nextCatalog := s.mustCatalog(s.currentConstantsHash)
	if nextCatalog == nil {
		return save.ExitDecision{}, ErrInvalidEngineState
	}
	return finishExitResolved(request, founder, founderRevision, company, companyRevision, now, exitType, terms, endedPrefix, executedRoutes, nextCatalog, s.currentConstantsHash)
}

func finishExitResolved(request IntentRequest, founder *save.State, founderRevision save.Revision, company *save.State, companyRevision save.Revision, now time.Time, exitType string, terms prestigecore.Terms, endedPrefix []save.EventWrite, executedRoutes []string, nextCatalog *economy.Catalog, nextConstantsHash string) (save.ExitDecision, error) {
	now = save.CanonicalServerTime(now)
	attended, err := prestigecore.AttendedMS(company, now)
	if err != nil {
		return save.ExitDecision{}, err
	}
	if terms.ReputationDelta > decimal.MaxExactInteger-founder.ReputationLevel {
		terms.ReputationDelta = decimal.MaxExactInteger - founder.ReputationLevel
	}
	if terms.RouteKnowledge > decimal.MaxExactInteger-founder.RouteKnowledgeBalance {
		return save.ExitDecision{}, ErrInvalidEngineState
	}
	founder.ReputationLevel += terms.ReputationDelta
	founder.RouteKnowledgeBalance += terms.RouteKnowledge
	founder.AgeMS += attended
	if founder.AgeMS > decimal.MaxExactInteger {
		return save.ExitDecision{}, ErrInvalidEngineState
	}
	for fact := range company.LedgerFactKinds {
		if founder.LedgerFactKinds == nil {
			founder.LedgerFactKinds = map[string]bool{}
		}
		founder.LedgerFactKinds[fact] = true
	}
	founder.NetworkSlots = mergeNetworkSlots(founder.NetworkSlots, terms.NetworkSlotUnlocks)
	founder.ExitHistory = append(founder.ExitHistory, save.ExitRecord{RunID: company.RunSeq, ExitType: exitType, OccurredAt: now, ReputationDelta: terms.ReputationDelta})
	if nextCatalog == nil || nextConstantsHash == "" {
		return save.ExitDecision{}, ErrInvalidEngineState
	}
	newCompany, err := prestigecore.NewRunState(nextCatalog, company, founder, now)
	if err != nil {
		return save.ExitDecision{}, err
	}
	runID := map[string]any{"company_stream_id": companyRevision.StreamID, "run_seq": company.RunSeq}
	assisted := map[string]bool{"commons": company.CompactMember, "advisor": founder.AdvisorMode}
	var factionID *string
	if company.FactionID != "" {
		value := company.FactionID
		factionID = &value
	}
	endedPayload, _ := json.Marshal(map[string]any{"founder_id": companyRevision.OwnerID, "run_id": runID, "exit_type": exitType,
		"started_at_ms": company.RunStartedAt.UnixMilli(), "ended_at_ms": now.UnixMilli(), "rta_ms": now.Sub(company.RunStartedAt).Milliseconds(),
		"attended_ms": attended, "pre_timer": company.RunPreTimer, "terminal_seq": companyRevision.RunLogSequence, "payout": terms, "tier": company.Tier,
		"lifetime_value": company.LifetimeValue.String(), "ledger_fact_kinds": sortedBoolKeys(company.LedgerFactKinds), "executed_routes": executedRoutes,
		"gates_crossed": sortedBoolKeys(company.GatesCrossed), "generators_purchased_total": company.GeneratorPurchasedTotal,
		"assisted": assisted, "faction": factionID})
	startedPayload, _ := json.Marshal(map[string]any{"founder_id": companyRevision.OwnerID, "run_id": map[string]any{"company_stream_id": companyRevision.StreamID, "run_seq": newCompany.RunSeq}, "started_at_ms": now.UnixMilli(), "assisted": map[string]bool{"commons": false, "advisor": founder.AdvisorMode}})
	advancedPayload, _ := json.Marshal(map[string]any{"founder_id": companyRevision.OwnerID, "run_id": runID, "exit_type": exitType, "reputation_delta": terms.ReputationDelta, "route_knowledge": terms.RouteKnowledge, "occurred_at_ms": now.UnixMilli()})
	receipt, _ := json.Marshal(map[string]any{"intent_id": request.IntentID, "outcome": "applied", "applied_count": 1, "receipt": map[string]any{"changes": []any{}}, "new_revision": companyRevision.Number + 2, "founder_revision": founderRevision.Number + 1, "evaluated_at": now.Format(time.RFC3339Nano), "snapshot": wireSnapshot(newCompany)})
	endedEvents := append([]save.EventWrite(nil), endedPrefix...)
	endedEvents = append(endedEvents, save.EventWrite{Kind: save.EventRunEnded, SchemaVersion: 2, IntentID: request.IntentID, Payload: endedPayload})
	return save.ExitDecision{Outcome: save.IntentApplied, Receipt: receipt, FinalCompanyState: company, NewCompanyState: newCompany, NewConstantsHash: nextConstantsHash,
		FounderEvents:      []save.EventWrite{{Kind: save.EventFounderAdvanced, SchemaVersion: 1, IntentID: request.IntentID, Payload: advancedPayload}},
		CompanyEndedEvents: endedEvents, CompanyStartedEvents: []save.EventWrite{{Kind: save.EventRunStarted, SchemaVersion: 1, IntentID: request.IntentID, Payload: startedPayload}}}, nil
}

func (s *Service) executedRoutesAt(ctx context.Context, streamID string, expectedRevision int64) ([]string, error) {
	loaded, err := s.store.LoadLatest(ctx, streamID)
	if err != nil {
		return nil, err
	}
	if loaded.Revision.Number != expectedRevision {
		// ApplyExitTransaction remains the authority for the typed revision
		// conflict. No facts from a different revision enter the terminal event.
		return []string{}, nil
	}
	return s.store.ExecutedRouteIDs(ctx, streamID, loaded.State.RunSeq, expectedRevision)
}

func rejectedExitDecision(request IntentRequest, revision int64, category, detail string) save.ExitDecision {
	return save.ExitDecision{Outcome: save.IntentRejected, Receipt: marshalRejection(request.IntentID, revision, category, detail)}
}

func (s *Service) applyLoggedExit(ctx context.Context, request IntentRequest, founder *save.State, founderRevision save.Revision, company *save.State, companyRevision save.Revision, command save.ReplayCommand, mode EvaluationMode, now time.Time, executedRoutes []string) (save.ExitDecision, json.RawMessage, error) {
	if err := requireFounderCatalogCoherence(founderRevision, companyRevision); err != nil {
		return save.ExitDecision{}, nil, err
	}
	if s.replayCatalogs == nil {
		return save.ExitDecision{}, nil, fmt.Errorf("%w: replay catalog bundle unavailable", ErrInvalidIntent)
	}
	current, ok := s.replayCatalogs.ResolveReplayCatalogs(companyRevision.ConstantsHash)
	if !ok {
		return save.ExitDecision{}, nil, fmt.Errorf("%w: current replay catalog bundle unavailable", ErrInvalidIntent)
	}
	next, ok := s.replayCatalogs.ResolveReplayCatalogs(s.currentConstantsHash)
	if !ok {
		return save.ExitDecision{}, nil, fmt.Errorf("%w: next replay catalog bundle unavailable", ErrInvalidIntent)
	}
	current.Next = &next
	contributions, settlements, err := s.resolveReplayAccrual(ctx, company, companyRevision, s.mustCatalog(companyRevision.ConstantsHash), current.Faction.StockCap, request)
	if err != nil {
		return save.ExitDecision{}, nil, err
	}
	carry := founderCarry(founder)
	carry.FounderRevision = founderRevision.Number
	carry.FounderConstantsHash = founderRevision.ConstantsHash
	selectedType := "collapse"
	selectedTerms := json.RawMessage(`{}`)
	switch request.Kind {
	case IntentCrossGate:
		selectedType = "scripted_first"
	case IntentWindDown:
		if len(founder.ExitHistory) == 0 {
			selectedType = "scripted_first"
		}
	case IntentAcceptExitOffer:
		selectedType = "unresolved"
		if company.OfferState != nil && company.OfferState.OfferID == request.OfferID {
			selectedType = company.OfferState.ExitType
			selectedTerms = append(json.RawMessage(nil), company.OfferState.TermsJSON...)
		}
	default:
		selectedType = "unresolved"
	}
	build := replayBuild{Command: command, Mode: mode, Now: now, IntentKind: request.Kind, Contributions: contributions,
		GuildSettlementBatch: settlements,
		RouteContextVersion:  current.Routes.ContextVersion(), FounderCarry: &carry, Terminal: true,
		ExecutedRouteIDs: executedRoutes, SelectedExitType: selectedType, SelectedTerms: selectedTerms, NextConstantsHash: s.currentConstantsHash}
	if company.CompactMember {
		weight, weightErr := s.resolveCommonsReplayWeight(ctx, companyRevision.StreamID, companyRevision.OwnerID, companyRevision.ConstantsHash)
		if weightErr != nil {
			return save.ExitDecision{}, nil, weightErr
		}
		build.CommonsWeightPPM = &weight
	}
	replayInputs, err := buildReplayInputs(build)
	if err != nil {
		return save.ExitDecision{}, nil, err
	}
	transition, err := ApplyLoggedExit(company, request.CanonicalPayload, current, replayInputs)
	if err != nil {
		return save.ExitDecision{}, nil, err
	}
	if transition.Decision.Outcome == save.IntentApplied {
		if err := applyFounderReplayOutput(founder, transition.Founder); err != nil {
			return save.ExitDecision{}, nil, err
		}
	}
	return transition.Decision, replayInputs, nil
}

func applyFounderReplayOutput(target, replayed *save.State) error {
	if target == nil || replayed == nil || len(replayed.ExitHistory) != len(target.ExitHistory)+1 {
		return ErrInvalidEngineState
	}
	target.ReputationLevel = replayed.ReputationLevel
	target.RouteKnowledgeBalance = replayed.RouteKnowledgeBalance
	target.AgeMS = replayed.AgeMS
	target.NetworkSlots = append([]save.NetworkSlot(nil), replayed.NetworkSlots...)
	target.LedgerFactKinds = cloneBools(replayed.LedgerFactKinds)
	target.ExitHistory = append(target.ExitHistory, replayed.ExitHistory[len(replayed.ExitHistory)-1])
	return nil
}

func (s *Service) mustCatalog(hash string) *economy.Catalog {
	catalog, _ := s.catalogs.Resolve(hash)
	return catalog
}

func (s *Service) exitErrorReceipt(ctx context.Context, streamID string, request IntentRequest, err error) (HandleResult, error) {
	var conflict *save.ExitRevisionConflict
	switch {
	case errors.As(err, &conflict):
		return HandleResult{Receipt: marshalRejection(request.IntentID, conflict.Current, "revision_conflict", string(conflict.Stream))}, nil
	case errors.Is(err, save.ErrIdempotencyConflict):
		current := request.ExpectedRevision
		if loaded, loadErr := s.store.LoadLatest(ctx, streamID); loadErr == nil {
			current = loaded.Revision.Number
		}
		return HandleResult{Receipt: marshalRejection(request.IntentID, current, "idempotency_conflict", request.IntentID)}, nil
	default:
		return HandleResult{}, err
	}
}

func (s *Service) finishExitResult(ctx context.Context, result save.IntentResult) (HandleResult, error) {
	if err := s.projectCommittedEvents(ctx, result.Events); err != nil {
		return HandleResult{}, err
	}
	return HandleResult{Receipt: result.Receipt, Replay: result.Replay}, nil
}

func mergeNetworkSlots(existing, grants []save.NetworkSlot) []save.NetworkSlot {
	bySlot := make(map[string]save.NetworkSlot, len(existing)+len(grants))
	for _, slot := range existing {
		bySlot[slot.Slot] = slot
	}
	for _, slot := range grants {
		bySlot[slot.Slot] = slot
	}
	keys := make([]string, 0, len(bySlot))
	for key := range bySlot {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	result := make([]save.NetworkSlot, 0, len(keys))
	for _, key := range keys {
		result = append(result, bySlot[key])
	}
	return result
}

func setTierFromGate(state *save.State, gateID string) error {
	tier, ok := tierForGate(gateID)
	if !ok {
		return ErrInvalidEngineState
	}
	if tier > state.Tier {
		state.Tier = tier
	}
	return nil
}

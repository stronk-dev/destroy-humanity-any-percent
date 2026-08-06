package production

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"
	"time"

	"cloud-clicker/server/decimal"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/pet"
	"cloud-clicker/server/save"
)

type FounderLoggedTransition struct {
	State               *save.State
	Outcome             save.IntentOutcome
	Receipt             json.RawMessage
	Events              []save.EventWrite
	ResultConstantsHash string
}

type founderReplayInputsWire struct {
	Version       int                       `json:"v"`
	Command       save.FounderReplayCommand `json:"command"`
	EvaluatedAtMS int64                     `json:"evaluated_at_ms"`
	Resolved      json.RawMessage           `json:"resolved"`
}

// ApplyFounderLogged is the projection-free Founder transition used by live
// Founder commands and career replay. It reads only its four arguments.
func ApplyFounderLogged(state *save.State, canonicalPayload []byte, catalogs CatalogBundle, replayInputs []byte) (result FounderLoggedTransition, resultErr error) {
	if state == nil || !catalogs.valid(catalogs.ConstantsHash) {
		return FounderLoggedTransition{}, fmt.Errorf("%w: Founder catalog bundle", ErrInvalidReplayInputs)
	}
	wire, err := parseFounderReplayInputs(replayInputs)
	if err != nil {
		return FounderLoggedTransition{}, err
	}
	if isMinigameResolutionPayload(canonicalPayload) {
		return applyFounderMinigameResolution(state, canonicalPayload, catalogs, wire)
	}
	request, err := parseLoggedIntent(canonicalPayload, wire.Command.IntentID)
	if err != nil || !bytes.Equal(request.CanonicalPayload, canonicalPayload) {
		return FounderLoggedTransition{}, fmt.Errorf("%w: Founder canonical command", ErrInvalidReplayInputs)
	}
	stateBefore, err := cloneFounderReplayState(state, catalogs.Economy)
	if err != nil {
		return FounderLoggedTransition{}, err
	}
	defer func() {
		if resultErr != nil || result.Outcome != save.IntentApplied {
			*state = *stateBefore
		}
	}()
	var header struct {
		Kind string `json:"kind"`
	}
	if err := json.Unmarshal(wire.Resolved, &header); err != nil || header.Kind == "" {
		return FounderLoggedTransition{}, fmt.Errorf("%w: Founder resolved kind", ErrInvalidReplayInputs)
	}
	revision := save.Revision{StreamID: wire.Command.FounderStreamID, OwnerID: wire.Command.FounderID,
		Number: wire.Command.Revision, ConstantsHash: catalogs.ConstantsHash}
	switch header.Kind {
	case "invalid":
		var resolved founderInvalidResolved
		if err := decodeReplayStrict(wire.Resolved, &resolved); err != nil || request.InvalidDetail == "" ||
			resolved.Detail != request.InvalidDetail {
			return FounderLoggedTransition{}, fmt.Errorf("%w: invalid Founder arm", ErrInvalidReplayInputs)
		}
		decision, err := rejectedDecision(request, revision.Number, "invalid", resolved.Detail)
		if err != nil {
			return FounderLoggedTransition{}, err
		}
		return FounderLoggedTransition{State: state, Outcome: decision.Outcome, Receipt: decision.Receipt,
			Events: []save.EventWrite{}, ResultConstantsHash: catalogs.ConstantsHash}, nil
	case string(IntentBuyRouteHint):
		if request.Kind != IntentBuyRouteHint || request.ExpectedRevision != wire.Command.Revision ||
			request.InvalidDetail != "" {
			return FounderLoggedTransition{}, fmt.Errorf("%w: route-hint Founder command", ErrInvalidReplayInputs)
		}
		var resolved founderRouteHintResolved
		if err := decodeReplayStrict(wire.Resolved, &resolved); err != nil ||
			resolved.RouteContextVersion != catalogs.Routes.ContextVersion() ||
			resolved.RouteKnowledgeBalance < 0 || resolved.RouteKnowledgeBalance > decimal.MaxExactInteger {
			return FounderLoggedTransition{}, fmt.Errorf("%w: route-hint Founder inputs", ErrInvalidReplayInputs)
		}
		state.RouteKnowledgeBalance = resolved.RouteKnowledgeBalance
		decision, err := TransitionWithPolicies(request, state, catalogs.Economy, catalogs.Routes, nil, nil,
			revision, ModeOnline, time.UnixMilli(wire.EvaluatedAtMS).UTC(), nil, nil, nil)
		if err != nil {
			return FounderLoggedTransition{}, err
		}
		return FounderLoggedTransition{State: state, Outcome: decision.Outcome, Receipt: decision.Receipt,
			Events: decision.Events, ResultConstantsHash: catalogs.ConstantsHash}, nil
	case IntentCareAction:
		return applyFounderCareResolved(state, request, revision, catalogs, wire.Resolved)
	case founderExitResolvedKind:
		if request.Kind != IntentCrossGate && request.ExpectedFounderRevision != wire.Command.Revision {
			return FounderLoggedTransition{}, fmt.Errorf("%w: Exit Founder revision", ErrInvalidReplayInputs)
		}
		var resolved founderExitResolvedWire
		if err := decodeReplayStrict(wire.Resolved, &resolved); err != nil {
			return FounderLoggedTransition{}, fmt.Errorf("%w: Exit Founder facts", ErrInvalidReplayInputs)
		}
		return applyFounderExitResolved(state, wire.Command, request, catalogs, resolved)
	default:
		return FounderLoggedTransition{}, fmt.Errorf("%w: unknown Founder resolved arm", ErrInvalidReplayInputs)
	}
}

func applyFounderCareResolved(state *save.State, request IntentRequest, revision save.Revision, catalogs CatalogBundle,
	resolvedJSON json.RawMessage) (FounderLoggedTransition, error) {
	if request.Kind != IntentCareAction || request.ExpectedRevision != revision.Number || request.InvalidDetail != "" ||
		catalogs.Pets == nil || state.Pets == nil {
		return FounderLoggedTransition{}, fmt.Errorf("%w: care Founder command", ErrInvalidReplayInputs)
	}
	var resolved founderCareResolved
	if err := decodeReplayStrict(resolvedJSON, &resolved); err != nil || resolved.Kind != IntentCareAction ||
		resolved.PetAttendedBeforeMS < 0 || resolved.PetAttendedBeforeMS > decimal.MaxExactInteger ||
		resolved.Attendance.CompanyConstantsHash != catalogs.ConstantsHash ||
		ValidateFounderAttendanceSample(state, revision.Number, request.ExpectedRevision, resolved.Attendance) != nil {
		return FounderLoggedTransition{}, fmt.Errorf("%w: care Founder inputs", ErrInvalidReplayInputs)
	}
	care, exists := state.Pets[request.PetID]
	if !exists {
		decision, err := rejectedDecision(request, revision.Number, "unknown_id", "unknown_pet")
		return founderDecisionTransition(state, decision, catalogs.ConstantsHash, err)
	}
	if care.EvaluatedThroughAttendedMS != resolved.PetAttendedBeforeMS {
		return FounderLoggedTransition{}, fmt.Errorf("%w: stale care cursor", ErrInvalidReplayInputs)
	}
	priorBand, _, err := pet.CareStatus(care, catalogs.Pets)
	if err != nil {
		return FounderLoggedTransition{}, err
	}
	careResult, err := pet.ApplyCareTransition(care, catalogs.Pets, pet.CareTransitionInput{
		ActionID: request.ActionID, AttendedBeforeMS: resolved.PetAttendedBeforeMS,
		AttendedAfterMS: resolved.Attendance.EffectiveFounderAttendedMS,
	})
	if err != nil {
		return FounderLoggedTransition{}, err
	}
	if !careResult.Applied {
		category := "not_eligible"
		if careResult.RejectionDetail == pet.RejectionUnknownAction {
			category = "unknown_id"
		}
		decision, err := rejectedDecision(request, revision.Number, category, string(careResult.RejectionDetail))
		return founderDecisionTransition(state, decision, catalogs.ConstantsHash, err)
	}
	state.Pets[request.PetID] = careResult.State
	receipt, err := json.Marshal(founderCareReceipt{
		IntentID: request.IntentID, Outcome: string(save.IntentApplied), FounderRevision: revision.Number + 1,
		PetID: request.PetID, ActionID: request.ActionID, StatID: careResult.StatID,
		BeforePPM: careResult.BeforePPM, AppliedPPM: careResult.AppliedPPM, AfterPPM: careResult.AfterPPM,
		TrustBeforePPM: careResult.TrustBeforePPM, TrustAfterPPM: careResult.TrustAfterPPM,
		Mood: careResult.Mood, StatusBand: careResult.StatusBand,
		NextEligibleAttendedMS: careResult.NextEligibleAttendedMS,
	})
	if err != nil {
		return FounderLoggedTransition{}, err
	}
	payload, _ := json.Marshal(map[string]any{
		"pet_id": request.PetID, "action_id": request.ActionID, "stat_id": careResult.StatID,
		"before_ppm": careResult.BeforePPM, "applied_ppm": careResult.AppliedPPM,
		"after_ppm": careResult.AfterPPM, "trust_before_ppm": careResult.TrustBeforePPM,
		"trust_after_ppm": careResult.TrustAfterPPM, "mood": careResult.Mood,
		"status_band": careResult.StatusBand, "next_eligible_attended_ms": careResult.NextEligibleAttendedMS,
	})
	events := []save.EventWrite{{Kind: save.EventPetCareApplied, SchemaVersion: 1,
		IntentID: request.IntentID, Payload: payload}}
	if careResult.StatusChanged {
		statusPayload, _ := json.Marshal(map[string]any{
			"pet_id": request.PetID, "from_status_band": priorBand, "to_status_band": careResult.StatusBand,
		})
		events = append(events, save.EventWrite{Kind: save.EventPetStatusChanged, SchemaVersion: 1,
			IntentID: request.IntentID, Payload: statusPayload})
	}
	return FounderLoggedTransition{State: state, Outcome: save.IntentApplied, Receipt: receipt,
		Events: events, ResultConstantsHash: catalogs.ConstantsHash}, nil
}

func founderDecisionTransition(state *save.State, decision save.IntentDecision, hash string, err error) (FounderLoggedTransition, error) {
	if err != nil {
		return FounderLoggedTransition{}, err
	}
	return FounderLoggedTransition{State: state, Outcome: decision.Outcome, Receipt: decision.Receipt,
		Events: decision.Events, ResultConstantsHash: hash}, nil
}

func cloneFounderReplayState(state *save.State, catalog *economy.Catalog) (*save.State, error) {
	encoded, err := save.EncodeState(state)
	if err != nil {
		return nil, err
	}
	return save.RestoreState(encoded, save.VersionForState(state), catalog, economy.ScopeFounder, time.Time{})
}

func applyFounderExitResolved(state *save.State, command save.FounderReplayCommand, request IntentRequest, catalogs CatalogBundle, resolved founderExitResolvedWire) (FounderLoggedTransition, error) {
	inputHash := catalogs.ConstantsHash
	if resolved.Kind != founderExitResolvedKind || resolved.CompanyStreamID == "" || resolved.RunSeq < 1 ||
		resolved.RunLogSeq < 1 || resolved.AgeMSBefore != state.AgeMS || resolved.AgeMSAfter < resolved.AgeMSBefore ||
		resolved.AttendedMS != resolved.AgeMSAfter-resolved.AgeMSBefore || resolved.AgeMSAfter > decimal.MaxExactInteger ||
		resolved.ResultConstantsHash == "" || resolved.ResultFounderWireVersion < 1 {
		return FounderLoggedTransition{}, ErrInvalidReplayInputs
	}
	if resolved.Outcome == string(save.IntentRejected) {
		if resolved.ResultConstantsHash != inputHash || resolved.Rejection == nil ||
			resolved.ReputationDelta != 0 || resolved.RouteKnowledgeDelta != 0 || resolved.AttendedMS != 0 ||
			resolved.AchievementScoreDelta != 0 || len(resolved.AddedNetworkSlots) != 0 ||
			len(resolved.AddedLedgerFactKinds) != 0 || len(resolved.AddedLifetimeAchievements) != 0 ||
			resolved.ExitRecord != nil || resolved.ResultFounderWireVersion != save.VersionForState(state) {
			return FounderLoggedTransition{}, ErrInvalidReplayInputs
		}
		current := command.Revision
		receipt, _ := json.Marshal(founderExitAuditReceipt{IntentID: command.IntentID, Outcome: resolved.Outcome,
			CurrentRevision: &current, Rejection: resolved.Rejection})
		return FounderLoggedTransition{State: state, Outcome: save.IntentRejected, Receipt: receipt,
			Events: []save.EventWrite{}, ResultConstantsHash: inputHash}, nil
	}
	if resolved.Outcome != string(save.IntentApplied) || resolved.Rejection != nil || resolved.ExitRecord == nil ||
		resolved.ExitRecord.RunID != resolved.RunSeq || resolved.ExitRecord.ReputationDelta != resolved.ReputationDelta ||
		resolved.ReputationDelta < 0 || resolved.RouteKnowledgeDelta < 0 || resolved.AchievementScoreDelta < 0 ||
		resolved.ReputationDelta > decimal.MaxExactInteger-state.ReputationLevel ||
		resolved.RouteKnowledgeDelta > decimal.MaxExactInteger-state.RouteKnowledgeBalance ||
		resolved.AchievementScoreDelta > decimal.MaxExactInteger-state.AchievementScoreLifetime ||
		!sortedUniqueNetworkSlots(resolved.AddedNetworkSlots) ||
		!sortedUniqueMechanical(resolved.AddedLedgerFactKinds) || !sortedUniqueMechanical(resolved.AddedLifetimeAchievements) {
		return FounderLoggedTransition{}, ErrInvalidReplayInputs
	}
	state.ReputationLevel += resolved.ReputationDelta
	state.RouteKnowledgeBalance += resolved.RouteKnowledgeDelta
	state.AgeMS = resolved.AgeMSAfter
	state.AchievementScoreLifetime += resolved.AchievementScoreDelta
	state.NetworkSlots = mergeNetworkSlots(state.NetworkSlots, resolved.AddedNetworkSlots)
	if state.LedgerFactKinds == nil {
		state.LedgerFactKinds = map[string]bool{}
	}
	for _, key := range resolved.AddedLedgerFactKinds {
		state.LedgerFactKinds[key] = true
	}
	if state.AchievementsEarnedLifetime == nil {
		state.AchievementsEarnedLifetime = map[string]bool{}
	}
	for _, key := range resolved.AddedLifetimeAchievements {
		state.AchievementsEarnedLifetime[key] = true
	}
	state.ExitHistory = append(state.ExitHistory, save.ExitRecord{RunID: resolved.ExitRecord.RunID,
		ExitType: resolved.ExitRecord.ExitType, OccurredAt: time.UnixMilli(resolved.ExitRecord.OccurredAtMS).UTC(),
		ReputationDelta: resolved.ExitRecord.ReputationDelta})
	resultCatalogs := catalogs
	if resolved.ResultConstantsHash != inputHash {
		if catalogs.Next == nil || !catalogs.Next.valid(resolved.ResultConstantsHash) {
			return FounderLoggedTransition{}, ErrInvalidReplayInputs
		}
		resultCatalogs = *catalogs.Next
	}
	wantFounderVersion, _ := resultCatalogs.versionFloors()
	if resolved.ResultFounderWireVersion != wantFounderVersion || activateFounderFeatureState(state, resultCatalogs, resolved.ResultFounderWireVersion, command.ServerTSMS) != nil {
		return FounderLoggedTransition{}, ErrInvalidReplayInputs
	}
	state.WireVersion = resolved.ResultFounderWireVersion
	if err := resultCatalogs.ValidateFoundationState(state); err != nil {
		return FounderLoggedTransition{}, err
	}
	next := command.Revision + 1
	receipt, _ := json.Marshal(founderExitAuditReceipt{IntentID: command.IntentID, Outcome: resolved.Outcome,
		FounderRevision: &next, ResultConstantsHash: resolved.ResultConstantsHash})
	eventPayload, _ := json.Marshal(map[string]any{"founder_id": command.FounderID,
		"run_id":    map[string]any{"company_stream_id": resolved.CompanyStreamID, "run_seq": resolved.RunSeq},
		"exit_type": resolved.ExitRecord.ExitType, "reputation_delta": resolved.ReputationDelta,
		"route_knowledge": resolved.RouteKnowledgeDelta, "occurred_at_ms": resolved.ExitRecord.OccurredAtMS})
	events := []save.EventWrite{{Kind: save.EventFounderAdvanced, SchemaVersion: 1,
		IntentID: command.IntentID, Payload: eventPayload}}
	return FounderLoggedTransition{State: state, Outcome: save.IntentApplied, Receipt: receipt,
		Events: events, ResultConstantsHash: resolved.ResultConstantsHash}, nil
}

func activateFounderFeatureState(state *save.State, catalogs CatalogBundle, resultVersion int, serverTSMS int64) error {
	current := save.VersionForState(state)
	if resultVersion < current {
		return ErrInvalidReplayInputs
	}
	if resultVersion >= 17 && current < 17 {
		if catalogs.Minigames == nil || len(catalogs.Minigames.MinigameIDs()) != 0 {
			return ErrInvalidReplayInputs
		}
		state.MinigameRatings = map[string]save.MinigameRatingState{}
		state.MinigameOfflineQuality = map[string]save.MinigameOfflineQualityState{}
	}
	if resultVersion >= 18 && current < 18 {
		if catalogs.Pets == nil {
			return ErrInvalidReplayInputs
		}
		state.Pets = map[string]pet.CareState{}
	}
	if resultVersion >= 19 && current < 19 {
		if catalogs.Fiscal == nil || serverTSMS <= 0 || serverTSMS > decimal.MaxExactInteger {
			return ErrInvalidReplayInputs
		}
		state.FiscalCredit, state.FiscalPeriodOpenedWallMS, state.FiscalPeriodSequence = 0, serverTSMS, 0
		state.FiscalGeneratorLevels = make(map[string]int64, len(catalogs.Fiscal.GeneratorLevelRows()))
		for _, row := range catalogs.Fiscal.GeneratorLevelRows() {
			state.FiscalGeneratorLevels[row.GeneratorID] = 0
		}
		state.FiscalUnlocks = map[string]bool{}
	}
	return nil
}

func parseFounderReplayInputs(data []byte) (founderReplayInputsWire, error) {
	var wire founderReplayInputsWire
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&wire); err != nil {
		return wire, ErrInvalidReplayInputs
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return wire, ErrInvalidReplayInputs
	}
	if wire.Version != save.FounderReplayInputsVersion || wire.EvaluatedAtMS != wire.Command.ServerTSMS ||
		wire.Command.Revision < 1 || wire.Command.FounderLogSeq < 1 || len(wire.Resolved) < 2 {
		return wire, ErrInvalidReplayInputs
	}
	if _, err := save.ValidateFounderReplayInputs(data, wire.Command); err != nil {
		return wire, err
	}
	return wire, nil
}

func sortedUniqueNetworkSlots(values []save.NetworkSlot) bool {
	copyValues := append([]save.NetworkSlot(nil), values...)
	sort.Slice(copyValues, func(i, j int) bool {
		if copyValues[i].Slot != copyValues[j].Slot {
			return copyValues[i].Slot < copyValues[j].Slot
		}
		return copyValues[i].CarriedRef < copyValues[j].CarriedRef
	})
	for index := range values {
		if values[index] != copyValues[index] || index > 0 && values[index] == values[index-1] {
			return false
		}
	}
	return true
}

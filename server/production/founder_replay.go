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
	"cloud-clicker/server/fiscal"
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
	stateBefore, err := cloneFounderReplayState(state, catalogs.Economy)
	if err != nil {
		return FounderLoggedTransition{}, err
	}
	var sweep *fiscal.SweepResult
	if catalogs.Fiscal != nil {
		if save.VersionForState(state) < 19 {
			return FounderLoggedTransition{}, fmt.Errorf("%w: inactive Fiscal state", ErrInvalidReplayInputs)
		}
		fiscalState := fiscalStateFromSave(state)
		sweep, err = catalogs.Fiscal.Sweep(&fiscalState, wire.Command.ServerTSMS)
		if err != nil {
			return FounderLoggedTransition{}, fmt.Errorf("%w: Fiscal sweep", ErrInvalidReplayInputs)
		}
		fiscalStateToSave(state, fiscalState)
	}
	defer func() {
		if resultErr != nil || result.Outcome != save.IntentApplied {
			*state = *stateBefore
			return
		}
		if sweep != nil {
			if err := decorateFounderFiscalSweep(&result, wire.Command.IntentID, sweep); err != nil {
				*state = *stateBefore
				result = FounderLoggedTransition{}
				resultErr = err
			}
		}
	}()
	if isMinigameResolutionPayload(canonicalPayload) {
		return applyFounderMinigameResolution(state, canonicalPayload, catalogs, wire)
	}
	request, err := parseLoggedIntent(canonicalPayload, wire.Command.IntentID)
	if err != nil || !bytes.Equal(request.CanonicalPayload, canonicalPayload) {
		return FounderLoggedTransition{}, fmt.Errorf("%w: Founder canonical command", ErrInvalidReplayInputs)
	}
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
	case IntentHarvestFiscalPeriod:
		return applyFounderFiscalHarvestResolved(state, request, revision, catalogs, wire.Command.ServerTSMS, wire.Resolved)
	case IntentSpendFiscalCredit:
		return applyFounderFiscalSpendResolved(state, request, revision, catalogs, wire.Command.ServerTSMS, wire.Resolved)
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

type founderFiscalSweepWire struct {
	Periods          int64  `json:"periods"`
	CreditBefore     int64  `json:"credit_before"`
	Credited         int64  `json:"credited"`
	CreditAfter      int64  `json:"credit_after"`
	OpenedBeforeMS   int64  `json:"opened_before_ms"`
	OpenedAfterMS    int64  `json:"opened_after_ms"`
	SeqBefore        int64  `json:"seq_before"`
	SeqAfter         int64  `json:"seq_after"`
	Saturated        bool   `json:"saturated"`
	HardcapReasonKey string `json:"hardcap_reason_key"`
}

type fiscalTargetWire struct {
	Kind        string `json:"kind"`
	GeneratorID string `json:"generator_id,omitempty"`
	Levels      int64  `json:"levels,omitempty"`
	UnlockID    string `json:"unlock_id,omitempty"`
}

type founderFiscalHarvestResolved struct {
	Kind                     string `json:"kind"`
	NowWallMS                int64  `json:"now_wall_ms"`
	PeriodOpenedWallMSBefore int64  `json:"period_opened_wall_ms_before"`
	PeriodsSwept             int64  `json:"periods_swept"`
	SeqBefore                int64  `json:"seq_before"`
	DrawPPM                  *int64 `json:"draw_ppm"`
	Outcome                  string `json:"outcome"`
}

type founderFiscalSpendResolved struct {
	Kind         string           `json:"kind"`
	Target       fiscalTargetWire `json:"target"`
	ResolvedCost int64            `json:"resolved_cost"`
}

func fiscalStateFromSave(state *save.State) fiscal.State {
	levels := make(map[string]int64, len(state.FiscalGeneratorLevels))
	for id, level := range state.FiscalGeneratorLevels {
		levels[id] = level
	}
	unlocks := make([]string, 0, len(state.FiscalUnlocks))
	for id, value := range state.FiscalUnlocks {
		if value {
			unlocks = append(unlocks, id)
		}
	}
	sort.Strings(unlocks)
	return fiscal.State{Credit: state.FiscalCredit, PeriodOpenedWallMS: state.FiscalPeriodOpenedWallMS,
		PeriodSequence: state.FiscalPeriodSequence, GeneratorLevels: levels, Unlocks: unlocks}
}

func fiscalStateToSave(state *save.State, value fiscal.State) {
	state.FiscalCredit = value.Credit
	state.FiscalPeriodOpenedWallMS = value.PeriodOpenedWallMS
	state.FiscalPeriodSequence = value.PeriodSequence
	state.FiscalGeneratorLevels = value.GeneratorLevels
	state.FiscalUnlocks = make(map[string]bool, len(value.Unlocks))
	for _, id := range value.Unlocks {
		state.FiscalUnlocks[id] = true
	}
}

func fiscalSweepWire(value *fiscal.SweepResult) founderFiscalSweepWire {
	return founderFiscalSweepWire{Periods: value.Periods, CreditBefore: value.CreditBefore,
		Credited: value.Credited, CreditAfter: value.CreditAfter, OpenedBeforeMS: value.OpenedBeforeMS,
		OpenedAfterMS: value.OpenedAfterMS, SeqBefore: value.SequenceBefore, SeqAfter: value.SequenceAfter,
		Saturated: value.Saturated, HardcapReasonKey: value.HardcapReasonKey}
}

func decorateFounderFiscalSweep(result *FounderLoggedTransition, intentID string, sweep *fiscal.SweepResult) error {
	var receipt map[string]json.RawMessage
	if err := json.Unmarshal(result.Receipt, &receipt); err != nil || receipt == nil {
		return fmt.Errorf("%w: Fiscal receipt", ErrInvalidReplayInputs)
	}
	currentSweep, hasSweep := receipt["fiscal_sweep"]
	if hasSweep && string(currentSweep) != "null" {
		return fmt.Errorf("%w: Fiscal receipt", ErrInvalidReplayInputs)
	}
	encodedSweep, err := json.Marshal(fiscalSweepWire(sweep))
	if err != nil {
		return err
	}
	receipt["fiscal_sweep"] = encodedSweep
	result.Receipt, err = json.Marshal(receipt)
	if err != nil {
		return err
	}
	payload, err := json.Marshal(struct {
		Source string `json:"source"`
		founderFiscalSweepWire
	}{Source: "automatic", founderFiscalSweepWire: fiscalSweepWire(sweep)})
	if err != nil {
		return err
	}
	event := save.EventWrite{Kind: save.EventFiscalPeriodHarvested, SchemaVersion: 1,
		IntentID: intentID, Payload: payload}
	result.Events = append([]save.EventWrite{event}, result.Events...)
	return nil
}

func applyFounderFiscalHarvestResolved(state *save.State, request IntentRequest, revision save.Revision,
	catalogs CatalogBundle, commandWallMS int64, resolvedJSON json.RawMessage) (FounderLoggedTransition, error) {
	if request.Kind != IntentHarvestFiscalPeriod || request.ExpectedRevision != revision.Number ||
		request.InvalidDetail != "" || catalogs.Fiscal == nil || save.VersionForState(state) < 19 {
		return FounderLoggedTransition{}, fmt.Errorf("%w: Fiscal harvest command", ErrInvalidReplayInputs)
	}
	var resolved founderFiscalHarvestResolved
	if err := decodeReplayStrict(resolvedJSON, &resolved); err != nil || resolved.Kind != IntentHarvestFiscalPeriod ||
		resolved.NowWallMS != commandWallMS || resolved.NowWallMS < resolved.PeriodOpenedWallMSBefore || resolved.NowWallMS > decimal.MaxExactInteger ||
		resolved.SeqBefore < 0 || resolved.SeqBefore > decimal.MaxExactInteger || resolved.PeriodsSwept < 0 {
		return FounderLoggedTransition{}, fmt.Errorf("%w: Fiscal harvest inputs", ErrInvalidReplayInputs)
	}
	beforeCredit := state.FiscalCredit
	result := fiscal.HarvestResult{NowWallMS: resolved.NowWallMS, PeriodOpenedBeforeWallMS: resolved.PeriodOpenedWallMSBefore,
		PeriodsSwept: resolved.PeriodsSwept, SequenceBefore: resolved.SeqBefore, DrawPPM: resolved.DrawPPM,
		Outcome: fiscal.HarvestOutcome(resolved.Outcome), CreditBefore: beforeCredit, CreditAfter: state.FiscalCredit}
	if resolved.Outcome == string(fiscal.HarvestConsumedByAuto) {
		if resolved.PeriodsSwept < 1 || state.FiscalPeriodSequence != resolved.SeqBefore+resolved.PeriodsSwept ||
			state.FiscalPeriodOpenedWallMS != resolved.PeriodOpenedWallMSBefore+resolved.PeriodsSwept*catalogs.Fiscal.Clock.AutoMS || resolved.DrawPPM != nil {
			return FounderLoggedTransition{}, fmt.Errorf("%w: consumed Fiscal harvest", ErrInvalidReplayInputs)
		}
	} else {
		fiscalState := fiscalStateFromSave(state)
		applied, err := catalogs.Fiscal.Harvest(&fiscalState, revision.OwnerID, resolved.NowWallMS)
		if errors.Is(err, fiscal.ErrNotRipe) && resolved.Outcome == "rejected" && resolved.DrawPPM == nil && resolved.PeriodsSwept == 0 {
			decision, decisionErr := rejectedDecision(request, revision.Number, "not_eligible", "period_not_ripe")
			return founderDecisionTransition(state, decision, catalogs.ConstantsHash, decisionErr)
		}
		if err != nil || applied.PeriodOpenedBeforeWallMS != resolved.PeriodOpenedWallMSBefore ||
			applied.SequenceBefore != resolved.SeqBefore || applied.PeriodsSwept != resolved.PeriodsSwept ||
			!equalInt64Pointer(applied.DrawPPM, resolved.DrawPPM) || string(applied.Outcome) != resolved.Outcome {
			return FounderLoggedTransition{}, fmt.Errorf("%w: Fiscal harvest resolution", ErrInvalidReplayInputs)
		}
		fiscalStateToSave(state, fiscalState)
		result = applied
	}
	receipt, err := json.Marshal(map[string]any{"intent_id": request.IntentID, "outcome": string(save.IntentApplied),
		"founder_revision": revision.Number + 1, "fiscal_sweep": nil, "source": "manual",
		"fiscal_credit_before": result.CreditBefore, "fiscal_credit_after": state.FiscalCredit,
		"period_opened_wall_ms": state.FiscalPeriodOpenedWallMS, "periods_swept": result.PeriodsSwept,
		"seq_before": result.SequenceBefore, "seq_after": state.FiscalPeriodSequence, "draw_ppm": result.DrawPPM,
		"harvest_outcome": result.Outcome, "saturated": result.Saturated})
	if err != nil {
		return FounderLoggedTransition{}, err
	}
	events := []save.EventWrite{}
	if result.Outcome != fiscal.HarvestConsumedByAuto {
		payload, _ := json.Marshal(map[string]any{"source": "manual", "outcome": result.Outcome,
			"credit_before": result.CreditBefore, "credit_after": state.FiscalCredit,
			"period_opened_wall_ms_before": result.PeriodOpenedBeforeWallMS,
			"period_opened_wall_ms_after":  state.FiscalPeriodOpenedWallMS, "seq_before": result.SequenceBefore,
			"seq_after": state.FiscalPeriodSequence, "draw_ppm": result.DrawPPM, "saturated": result.Saturated})
		events = append(events, save.EventWrite{Kind: save.EventFiscalPeriodHarvested, SchemaVersion: 1,
			IntentID: request.IntentID, Payload: payload})
	}
	return FounderLoggedTransition{State: state, Outcome: save.IntentApplied, Receipt: receipt, Events: events,
		ResultConstantsHash: catalogs.ConstantsHash}, nil
}

func applyFounderFiscalSpendResolved(state *save.State, request IntentRequest, revision save.Revision,
	catalogs CatalogBundle, nowWallMS int64, resolvedJSON json.RawMessage) (FounderLoggedTransition, error) {
	if request.Kind != IntentSpendFiscalCredit || request.ExpectedRevision != revision.Number ||
		request.InvalidDetail != "" || catalogs.Fiscal == nil || save.VersionForState(state) < 19 {
		return FounderLoggedTransition{}, fmt.Errorf("%w: Fiscal spend command", ErrInvalidReplayInputs)
	}
	var resolved founderFiscalSpendResolved
	if err := decodeReplayStrict(resolvedJSON, &resolved); err != nil || resolved.Kind != IntentSpendFiscalCredit ||
		resolved.Target != fiscalTargetFromRequest(request) || resolved.ResolvedCost < 0 {
		return FounderLoggedTransition{}, fmt.Errorf("%w: Fiscal spend inputs", ErrInvalidReplayInputs)
	}
	fiscalState := fiscalStateFromSave(state)
	expectedCost, costErr := resolvedFiscalCost(catalogs.Fiscal, fiscalState, request.FiscalTarget)
	if costErr == nil && expectedCost != resolved.ResolvedCost || costErr != nil && resolved.ResolvedCost != 0 {
		return FounderLoggedTransition{}, fmt.Errorf("%w: Fiscal spend cost", ErrInvalidReplayInputs)
	}
	before := fiscalState.Credit
	if nowWallMS < 1 || nowWallMS > decimal.MaxExactInteger {
		return FounderLoggedTransition{}, fmt.Errorf("%w: Fiscal spend timestamp", ErrInvalidReplayInputs)
	}
	applied, err := catalogs.Fiscal.Spend(&fiscalState, nowWallMS, request.FiscalTarget)
	if err != nil {
		category, detail := "invalid", "fiscal_target"
		switch {
		case errors.Is(err, fiscal.ErrUnknownTarget):
			category, detail = "unknown_id", fiscalTargetID(request.FiscalTarget)
		case errors.Is(err, fiscal.ErrAlreadyUnlocked):
			category, detail = "not_eligible", "already_unlocked"
		case errors.Is(err, fiscal.ErrUnaffordable):
			category, detail = "unaffordable", "fiscal_credit"
		case errors.Is(err, fiscal.ErrCapExceeded):
			category, detail = "cap_exceeded", fiscalTargetID(request.FiscalTarget)
		default:
			return FounderLoggedTransition{}, err
		}
		decision, decisionErr := rejectedDecision(request, revision.Number, category, detail)
		return founderDecisionTransition(state, decision, catalogs.ConstantsHash, decisionErr)
	}
	if applied.ResolvedCost != resolved.ResolvedCost || applied.Target != request.FiscalTarget {
		return FounderLoggedTransition{}, fmt.Errorf("%w: Fiscal spend resolution", ErrInvalidReplayInputs)
	}
	fiscalStateToSave(state, fiscalState)
	receipt, err := json.Marshal(map[string]any{"intent_id": request.IntentID, "outcome": string(save.IntentApplied),
		"founder_revision": revision.Number + 1, "fiscal_sweep": nil, "target": resolved.Target,
		"resolved_cost": applied.ResolvedCost, "fiscal_credit_before": before, "fiscal_credit_after": fiscalState.Credit})
	if err != nil {
		return FounderLoggedTransition{}, err
	}
	payload, _ := json.Marshal(map[string]any{"target": resolved.Target, "resolved_cost": applied.ResolvedCost,
		"fiscal_credit_before": before, "fiscal_credit_after": fiscalState.Credit})
	return FounderLoggedTransition{State: state, Outcome: save.IntentApplied, Receipt: receipt,
		Events: []save.EventWrite{{Kind: save.EventFiscalCreditSpent, SchemaVersion: 1,
			IntentID: request.IntentID, Payload: payload}}, ResultConstantsHash: catalogs.ConstantsHash}, nil
}

func resolvedFiscalCost(catalog *fiscal.Catalog, state fiscal.State, target fiscal.SpendTarget) (int64, error) {
	switch target.Kind {
	case "generator_level":
		current, ok := state.GeneratorLevels[target.GeneratorID]
		if !ok {
			return 0, fiscal.ErrUnknownTarget
		}
		return catalog.GeneratorLevelCost(target.GeneratorID, current, target.Levels)
	case "unlock":
		row, ok := catalog.Unlock(target.UnlockID)
		if !ok {
			return 0, fiscal.ErrUnknownTarget
		}
		if stateUnlockContains(state.Unlocks, target.UnlockID) {
			return 0, fiscal.ErrAlreadyUnlocked
		}
		return row.Cost, nil
	default:
		return 0, fiscal.ErrUnknownTarget
	}
}

func stateUnlockContains(values []string, target string) bool {
	index := sort.SearchStrings(values, target)
	return index < len(values) && values[index] == target
}

func fiscalTargetID(target fiscal.SpendTarget) string {
	if target.Kind == "generator_level" {
		return target.GeneratorID
	}
	return target.UnlockID
}

func fiscalTargetFromRequest(request IntentRequest) fiscalTargetWire {
	return fiscalTargetWire{Kind: request.FiscalTarget.Kind, GeneratorID: request.FiscalTarget.GeneratorID,
		Levels: request.FiscalTarget.Levels, UnlockID: request.FiscalTarget.UnlockID}
}

func equalInt64Pointer(left, right *int64) bool {
	return left == nil && right == nil || left != nil && right != nil && *left == *right
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
			resolved.ExitRecord != nil || resolved.ResultFounderWireVersion != save.VersionForState(state) || resolved.NextSoul != nil {
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
	if resolved.ResultFounderWireVersion != wantFounderVersion || activateFounderFeatureState(state, resultCatalogs, resolved.ResultFounderWireVersion, command.ServerTSMS, resolved.NextSoul) != nil {
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

func activateFounderFeatureState(state *save.State, catalogs CatalogBundle, resultVersion int, serverTSMS int64, nextSoul *nextSoulWire) error {
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
	if resultVersion >= 20 && current < 20 {
		if catalogs.Soul == nil || nextSoul == nil || nextSoul.SoulInitial != catalogs.Soul.Policy.Initial {
			return ErrInvalidReplayInputs
		}
		band, ok := catalogs.Soul.BandFor(nextSoul.SoulInitial)
		if !ok || nextSoul.BandMember != string(band.Member) {
			return ErrInvalidReplayInputs
		}
		state.Soul, state.SoulExhaustedSourceIDs = nextSoul.SoulInitial, []string{}
	} else if nextSoul != nil {
		return ErrInvalidReplayInputs
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

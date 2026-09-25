package production

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"cloud-clicker/server/decimal"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/garden"
	"cloud-clicker/server/minigame"
	"cloud-clicker/server/save"
)

// Server Garden SG6: garden_harvest through the SG-P2 multi-stream coordinator.
//
// Two Founder resolved kinds keep the logs self-describing:
//   - `garden_harvest` is Founder-only and must end rejected (a rejection
//     touches neither the Company nor the faucet);
//   - `garden_harvest_credited` is written by the coordinator, carries a
//     Company source coordinate, and must end applied.
const (
	gardenHarvestCreditedKind = "garden_harvest_credited"
	gardenCompanyCreditKind   = "credit_garden_harvest"
)

var (
	errGardenHarvestConflict = errors.New("garden harvest expected revision is stale")
	errGardenHarvestRaced    = errors.New("garden harvest outcome changed under the lock")
	// errGardenHarvestMustCredit marks a Founder-only harvest arm that would
	// apply: applied harvests commit only through the coordinator.
	errGardenHarvestMustCredit = errors.New("an applied harvest must be credited through the coordinator")
)

type founderGardenHarvestResolved struct {
	Kind          string                  `json:"kind"`
	ServerMS      int64                   `json:"server_ms"`
	GardenSaltHex string                  `json:"garden_salt_hex,omitempty"`
	Advance       garden.Advance          `json:"advance"`
	Attendance    FounderAttendanceSample `json:"attendance"`
}

type founderGardenHarvestReceipt struct {
	IntentID        string         `json:"intent_id"`
	Outcome         string         `json:"outcome"`
	FounderRevision int64          `json:"founder_revision"`
	Kind            string         `json:"kind"`
	Harvest         garden.Harvest `json:"harvest"`
}

// applyFounderGardenHarvestResolved is SG6's Founder side, shared by the live
// coordinator, the Founder-only rejection path, and history replay.
func applyFounderGardenHarvestResolved(state *save.State, request IntentRequest, revision save.Revision, catalogs CatalogBundle,
	serverMS int64, resolvedJSON json.RawMessage, advance *garden.Advance) (FounderLoggedTransition, error) {
	if request.Kind != IntentGardenHarvest || request.ExpectedRevision != revision.Number || request.InvalidDetail != "" {
		return FounderLoggedTransition{}, fmt.Errorf("%w: garden harvest command", ErrInvalidReplayInputs)
	}
	var resolved founderGardenHarvestResolved
	if err := decodeReplayStrict(resolvedJSON, &resolved); err != nil || resolved.Kind != IntentGardenHarvest && resolved.Kind != gardenHarvestCreditedKind ||
		resolved.ServerMS != serverMS || resolved.Attendance.CompanyConstantsHash != catalogs.ConstantsHash ||
		ValidateFounderAttendanceSample(state, revision.Number, revision.Number, resolved.Attendance) != nil {
		return FounderLoggedTransition{}, fmt.Errorf("%w: garden harvest inputs", ErrInvalidReplayInputs)
	}
	recomputed := garden.Advance{Matured: []garden.PlotRef{}, Spawned: []garden.Spawn{}}
	if advance != nil {
		recomputed = *advance
	}
	if !advanceEqual(recomputed, resolved.Advance) {
		return FounderLoggedTransition{}, fmt.Errorf("%w: garden advance diverged from resolved inputs", ErrInvalidReplayInputs)
	}
	credited := resolved.Kind == gardenHarvestCreditedKind
	reject := func(category, detail string) (FounderLoggedTransition, error) {
		if credited {
			return FounderLoggedTransition{}, fmt.Errorf("%w: a credited harvest cannot be rejected", ErrInvalidReplayInputs)
		}
		decision, err := rejectedDecision(request, revision.Number, category, detail)
		return founderDecisionTransition(state, decision, catalogs.ConstantsHash, err)
	}
	if !gardenActive(catalogs, state) {
		return reject("not_eligible", garden.DetailInactive)
	}
	_, unlocked := gardenInputs(catalogs.Garden, state)
	humanLocked := false
	if catalogs.Garden.SoulGate == garden.SoulGateHumanHobby {
		if catalogs.Soul == nil {
			return FounderLoggedTransition{}, fmt.Errorf("%w: human_hobby garden without Soul", ErrInvalidReplayInputs)
		}
		locked, err := catalogs.Soul.HumanContentLocked(state.Soul)
		if err != nil {
			return FounderLoggedTransition{}, err
		}
		humanLocked = locked
	}
	if err := catalogs.Garden.Gate(unlocked, humanLocked); err != nil {
		rejection, _ := garden.AsRejection(err)
		return reject(rejection.Category, rejection.Detail)
	}
	harvest, err := catalogs.Garden.Harvest(state.ServerGarden, request.IntentID, request.GardenPlots)
	if err != nil {
		rejection, ok := garden.AsRejection(err)
		if !ok {
			return FounderLoggedTransition{}, err
		}
		return reject(rejection.Category, rejection.Detail)
	}
	if !credited {
		return FounderLoggedTransition{}, fmt.Errorf("%w: %w", ErrInvalidReplayInputs, errGardenHarvestMustCredit)
	}
	receipt, err := json.Marshal(founderGardenHarvestReceipt{IntentID: request.IntentID, Outcome: string(save.IntentApplied),
		FounderRevision: revision.Number + 1, Kind: request.Kind, Harvest: harvest})
	if err != nil {
		return FounderLoggedTransition{}, err
	}
	payload, err := json.Marshal(harvest)
	if err != nil {
		return FounderLoggedTransition{}, err
	}
	return FounderLoggedTransition{State: state, Outcome: save.IntentApplied, Receipt: receipt,
		Events:              []save.EventWrite{{Kind: save.EventGardenHarvested, SchemaVersion: 1, IntentID: request.IntentID, Payload: payload}},
		ResultConstantsHash: catalogs.ConstantsHash}, nil
}

// liveGardenHarvestResolved builds a harvest's resolved arm at serverMS.
func liveGardenHarvestResolved(bundle CatalogBundle, state *save.State, kind string, serverMS int64, attendance FounderAttendanceSample) (founderGardenHarvestResolved, error) {
	base, err := liveGardenResolved(bundle, state, kind, serverMS)
	if err != nil {
		return founderGardenHarvestResolved{}, err
	}
	return founderGardenHarvestResolved{Kind: kind, ServerMS: serverMS, GardenSaltHex: base.GardenSaltHex, Advance: base.Advance, Attendance: attendance}, nil
}

// gardenHarvestWouldApply probes the Founder side on a clone to route the
// request: an applied harvest must go through the coordinator, a rejection is
// logged Founder-only. Both authoritative paths re-run and assert the outcome.
func gardenHarvestWouldApply(bundle CatalogBundle, loaded save.Loaded, request IntentRequest, serverMS int64, attendance FounderAttendanceSample) (bool, error) {
	probe, err := cloneFounderReplayState(loaded.State, bundle.Economy)
	if err != nil {
		return false, err
	}
	resolved, err := liveGardenHarvestResolved(bundle, probe, IntentGardenHarvest, serverMS, attendance)
	if err != nil {
		return false, err
	}
	command := save.FounderReplayCommand{IntentID: request.IntentID, FounderStreamID: loaded.Revision.StreamID, FounderID: loaded.Key.OwnerID,
		Revision: loaded.Revision.Number, FounderLogSeq: 1, ServerTSMS: serverMS}
	inputs, err := save.MarshalFounderReplayInputs(command, resolved)
	if err != nil {
		return false, err
	}
	_, err = ApplyFounderLogged(probe, request.CanonicalPayload, bundle, inputs)
	if errors.Is(err, errGardenHarvestMustCredit) {
		return true, nil
	}
	return false, err
}

func (s *Service) handleGardenHarvest(ctx context.Context, companyStreamID string, request IntentRequest, now time.Time) (HandleResult, error) {
	founder, err := s.store.LoadSiblingLatest(ctx, companyStreamID, economy.ScopeFounder)
	if err != nil {
		return HandleResult{}, err
	}
	bundle, ok := s.replayCatalogs.ResolveReplayCatalogs(founder.Revision.ConstantsHash)
	if !ok {
		return HandleResult{}, fmt.Errorf("%w: replay catalog bundle unavailable", ErrInvalidIntent)
	}
	attendance, err := s.ResolveFounderAttendance(ctx, founder.Revision.StreamID, companyStreamID, now)
	if err != nil {
		return HandleResult{}, err
	}
	serverMS := save.CanonicalServerTime(now).UnixMilli()
	applies := false
	if founder.Revision.Number == request.ExpectedRevision {
		if applies, err = gardenHarvestWouldApply(bundle, founder, request, serverMS, attendance); err != nil {
			return HandleResult{}, err
		}
	}
	if applies {
		result, err := s.creditGardenHarvest(ctx, companyStreamID, founder, bundle, request, serverMS, attendance, nil)
		if errors.Is(err, errGardenHarvestConflict) || errors.Is(err, errGardenHarvestRaced) || errors.Is(err, ErrFounderAttendanceStale) {
			current := request.ExpectedRevision
			if loaded, loadErr := s.store.LoadLatest(ctx, founder.Revision.StreamID); loadErr == nil {
				current = loaded.Revision.Number
			}
			return HandleResult{Receipt: marshalRejection(request.IntentID, current, "revision_conflict", "expected_revision")}, nil
		}
		return result, err
	}
	return s.rejectGardenHarvest(ctx, founder, request, attendance)
}

// rejectGardenHarvest logs a rejected harvest on the Founder stream only.
func (s *Service) rejectGardenHarvest(ctx context.Context, founder save.Loaded, request IntentRequest, attendance FounderAttendanceSample) (HandleResult, error) {
	result, err := s.store.ApplyFounderLogged(ctx, founder.Revision.StreamID, request.ExpectedRevision,
		request.IntentID, request.RequestHash, request.CanonicalPayload,
		func(state *save.State, revision save.Revision, command save.FounderReplayCommand) (save.IntentDecision, json.RawMessage, error) {
			bundle, ok := s.replayCatalogs.ResolveReplayCatalogs(revision.ConstantsHash)
			if !ok {
				return save.IntentDecision{}, nil, fmt.Errorf("%w: replay catalog bundle unavailable", ErrInvalidIntent)
			}
			resolved, resolveErr := liveGardenHarvestResolved(bundle, state, IntentGardenHarvest, command.ServerTSMS, attendance)
			if resolveErr != nil {
				return save.IntentDecision{}, nil, resolveErr
			}
			replayInputs, marshalErr := marshalFounderReplayInputs(command, resolved)
			if marshalErr != nil {
				return save.IntentDecision{}, nil, marshalErr
			}
			transition, transitionErr := ApplyFounderLogged(state, request.CanonicalPayload, bundle, replayInputs)
			if transitionErr != nil {
				return save.IntentDecision{}, nil, transitionErr
			}
			return save.IntentDecision{Outcome: transition.Outcome, Receipt: transition.Receipt, Events: transition.Events}, replayInputs, nil
		})
	if err != nil {
		var conflict *save.RevisionConflict
		switch {
		case errors.As(err, &conflict):
			return HandleResult{Receipt: marshalRejection(request.IntentID, conflict.Current, "revision_conflict", "expected_revision")}, nil
		case errors.Is(err, save.ErrIdempotencyConflict):
			return HandleResult{Receipt: marshalRejection(request.IntentID, request.ExpectedRevision, "idempotency_conflict", request.IntentID)}, nil
		case errors.Is(err, errGardenHarvestMustCredit):
			// The harvest became applicable between the probe and the lock.
			return HandleResult{Receipt: marshalRejection(request.IntentID, request.ExpectedRevision, "revision_conflict", "expected_revision")}, nil
		default:
			return HandleResult{}, err
		}
	}
	if err := s.projectCommittedEvents(ctx, result.Events); err != nil {
		return HandleResult{}, err
	}
	return HandleResult{Receipt: result.Receipt, Replay: result.Replay}, nil
}

type gardenHarvestCreditPayload struct {
	Kind        string `json:"kind"`
	IntentID    string `json:"intent_id"`
	HarvestHash string `json:"harvest_hash"`
}

type gardenCompanyResolved struct {
	Kind            string                `json:"kind"`
	IntentID        string                `json:"intent_id"`
	HarvestHash     string                `json:"harvest_hash"`
	PolicyHash      string                `json:"policy_hash"`
	SelectedUnits   int64                 `json:"selected_units"`
	Faucet          *minigameFaucetWire   `json:"faucet"`
	Credited        string                `json:"credited"`
	ForfeitedUnits  int64                 `json:"forfeited_units"`
	CapReasonKey    *string               `json:"cap_reason_key"`
	FaucetApplied   bool                  `json:"faucet_applied"`
	FounderLog      minigameLogCoordinate `json:"founder_log_coordinate"`
	CompanyRevision int64                 `json:"company_revision"`
	FounderRevision int64                 `json:"founder_revision"`
}

type gardenCompanyReceipt struct {
	IntentID           string  `json:"intent_id"`
	Outcome            string  `json:"outcome"`
	HarvestHash        string  `json:"harvest_hash"`
	CreditedResourceID string  `json:"credited_resource_id"`
	Credited           string  `json:"credited"`
	ForfeitedUnits     int64   `json:"forfeited_units"`
	CapReasonKey       *string `json:"cap_reason_key"`
	FaucetApplied      bool    `json:"faucet_applied"`
	CompanyRevision    int64   `json:"company_revision"`
	FounderRevision    int64   `json:"founder_revision"`
}

type gardenHarvestAPIReceipt struct {
	IntentID           string         `json:"intent_id"`
	Outcome            string         `json:"outcome"`
	Kind               string         `json:"kind"`
	FounderRevision    int64          `json:"founder_revision"`
	CompanyRevision    int64          `json:"company_revision"`
	Harvest            garden.Harvest `json:"harvest"`
	CreditedResourceID string         `json:"credited_resource_id"`
	Credited           string         `json:"credited"`
	ForfeitedUnits     int64          `json:"forfeited_units"`
	CapReasonKey       *string        `json:"cap_reason_key"`
	FaucetApplied      bool           `json:"faucet_applied"`
	GardenAdvance      garden.Advance `json:"garden_advance"`
}

func gardenMinigamePolicy(policy garden.PayoutPolicy) minigame.PayoutPolicy {
	return minigame.PayoutPolicy{CreditedResourceID: policy.CreditedResourceID, SendsPerDay: policy.SendsPerDay, PerSendCap: policy.PerSendCap,
		ConversionPPM: policy.ConversionPPM, PayoutScoreFactID: policy.PayoutScoreFactID, CapReasonKey: policy.CapReasonKey}
}

// gardenPolicyHash binds the Company arm to the pinned payout row.
func gardenPolicyHash(policy garden.PayoutPolicy) string {
	encoded, _ := json.Marshal(policy)
	digest := sha256.Sum256(encoded)
	return "sha256:" + hex.EncodeToString(digest[:])
}

// creditGardenHarvest is SG6's coordinator: Founder → Company under one
// transaction, the faucet window keyed (founder, "server_garden", day), both
// logs bound by harvest_hash, and the Founder intent record as the
// exactly-once authority.
func (s *Service) creditGardenHarvest(ctx context.Context, companyStreamID string, founderLoaded save.Loaded, bundle CatalogBundle,
	request IntentRequest, serverMS int64, attendance FounderAttendanceSample, fault save.ExitFaultInjector,
) (HandleResult, error) {
	result, err := s.store.ApplyMinigameResolutionTransaction(ctx, save.MinigameResolutionRequest{
		SessionID: request.IntentID, FounderID: founderLoaded.Key.OwnerID, CompanyStreamID: companyStreamID,
		RequestHash: request.RequestHash, CanonicalPayload: request.CanonicalPayload, ServerTSMS: serverMS, FounderIdempotency: true,
	}, func(ctx context.Context, tx *sql.Tx, founder *save.State, founderRevision save.Revision,
		company *save.State, companyRevision save.Revision, companyCommand save.ReplayCommand,
		founderCommand save.FounderReplayCommand, companyNext, founderNext int64,
	) (save.MinigameResolutionDecision, error) {
		if founderRevision.Number != request.ExpectedRevision {
			return save.MinigameResolutionDecision{}, errGardenHarvestConflict
		}
		if companyRevision.ConstantsHash != bundle.ConstantsHash || attendance.CompanyRevision != companyRevision.Number ||
			attendance.CompanyConstantsHash != companyRevision.ConstantsHash ||
			ValidateFounderAttendanceSample(founder, founderRevision.Number, founderRevision.Number, attendance) != nil {
			return save.MinigameResolutionDecision{}, ErrFounderAttendanceStale
		}
		resolved, err := liveGardenHarvestResolved(bundle, founder, gardenHarvestCreditedKind, founderCommand.ServerTSMS, attendance)
		if err != nil {
			return save.MinigameResolutionDecision{}, err
		}
		founderInputs, err := save.MarshalFounderReplayInputs(founderCommand, resolved)
		if err != nil {
			return save.MinigameResolutionDecision{}, err
		}
		transition, err := ApplyFounderLogged(founder, request.CanonicalPayload, bundle, founderInputs)
		if err != nil || transition.Outcome != save.IntentApplied {
			return save.MinigameResolutionDecision{}, errGardenHarvestRaced
		}
		var founderReceipt struct {
			Harvest       garden.Harvest `json:"harvest"`
			GardenAdvance garden.Advance `json:"garden_advance"`
		}
		if err := json.Unmarshal(transition.Receipt, &founderReceipt); err != nil {
			return save.MinigameResolutionDecision{}, err
		}
		harvest := founderReceipt.Harvest
		policy := bundle.Garden.Payout
		var faucet *minigameFaucetWire
		credit := decimal.Zero
		forfeited := int64(0)
		var capReason *string
		if harvest.TotalUnits > 0 {
			applied, faucetErr := minigame.ApplyPersistentFaucetWindowTx(ctx, tx, founderLoaded.Key.OwnerID, garden.FaucetOwnerID,
				attendance.EffectiveFounderAttendedMS, gardenMinigamePolicy(policy), harvest.TotalUnits)
			if faucetErr != nil {
				return save.MinigameResolutionDecision{}, faucetErr
			}
			if fault != nil {
				if faultErr := fault("faucet_window"); faultErr != nil {
					return save.MinigameResolutionDecision{}, faultErr
				}
			}
			wire := faucetWire(applied)
			faucet = &wire
			credit = decimal.FromString(strconv.FormatInt(applied.CreditedUnits, 10))
			forfeited = applied.ForfeitedUnits
			if applied.ConfiguredCapReasonKey != "" {
				reason := applied.ConfiguredCapReasonKey
				capReason = &reason
			}
		}
		ledgerReceipt, err := company.Ledger.ApplyAccrual(economy.Transaction{Entries: []economy.Entry{{ResourceID: policy.CreditedResourceID, Delta: credit}}})
		if err != nil {
			return save.MinigameResolutionDecision{}, err
		}
		creditedDelta, err := minigameCreditedDelta(ledgerReceipt, policy.CreditedResourceID)
		if err != nil {
			return save.MinigameResolutionDecision{}, err
		}
		companyResolved := gardenCompanyResolved{Kind: gardenCompanyCreditKind, IntentID: request.IntentID, HarvestHash: harvest.HarvestHash,
			PolicyHash: gardenPolicyHash(policy), SelectedUnits: harvest.TotalUnits, Faucet: faucet, Credited: creditedDelta,
			ForfeitedUnits: forfeited, CapReasonKey: capReason, FaucetApplied: harvest.TotalUnits > 0,
			FounderLog:      minigameLogCoordinate{StreamID: founderCommand.FounderStreamID, Revision: founderNext, Sequence: founderCommand.FounderLogSeq},
			CompanyRevision: companyNext, FounderRevision: founderNext}
		companyInputs, err := json.Marshal(replayInputsWire{Version: save.ReplayInputsVersion, Command: companyCommand,
			EvaluatedAtMS: founderCommand.ServerTSMS, EvaluationMode: ModeOnline, Resolved: mustJSON(companyResolved)})
		if err != nil {
			return save.MinigameResolutionDecision{}, err
		}
		companyPayload, err := normalizeReplayJSON(mustJSON(gardenHarvestCreditPayload{Kind: gardenCompanyCreditKind, IntentID: request.IntentID, HarvestHash: harvest.HarvestHash}))
		if err != nil {
			return save.MinigameResolutionDecision{}, err
		}
		companyReceipt, companyEvent, err := gardenCompanyReceiptAndEvent(companyResolved, policy.CreditedResourceID)
		if err != nil {
			return save.MinigameResolutionDecision{}, err
		}
		apiReceipt, err := normalizeReplayJSON(mustJSON(gardenHarvestAPIReceipt{IntentID: request.IntentID, Outcome: string(save.IntentApplied),
			Kind: IntentGardenHarvest, FounderRevision: founderNext, CompanyRevision: companyNext, Harvest: harvest,
			CreditedResourceID: policy.CreditedResourceID, Credited: creditedDelta, ForfeitedUnits: forfeited, CapReasonKey: capReason,
			FaucetApplied: harvest.TotalUnits > 0, GardenAdvance: founderReceipt.GardenAdvance}))
		if err != nil {
			return save.MinigameResolutionDecision{}, err
		}
		return save.MinigameResolutionDecision{Receipt: apiReceipt, CompanyLogReceipt: companyReceipt, FounderReceipt: transition.Receipt,
			CompanyReplayInputs: companyInputs, FounderReplayResolved: mustJSON(resolved), CompanyCanonicalPayload: companyPayload,
			CompanyEvents: []save.EventWrite{companyEvent}, FounderEvents: transition.Events}, nil
	}, fault)
	if err != nil {
		return HandleResult{}, err
	}
	if err := s.projectCommittedEvents(ctx, result.Events); err != nil {
		return HandleResult{}, err
	}
	return HandleResult{Receipt: result.Receipt, Replay: result.Replay}, nil
}

func gardenCompanyReceiptAndEvent(resolved gardenCompanyResolved, resourceID string) (json.RawMessage, save.EventWrite, error) {
	receipt, err := normalizeReplayJSON(mustJSON(gardenCompanyReceipt{IntentID: resolved.IntentID, Outcome: string(save.IntentApplied),
		HarvestHash: resolved.HarvestHash, CreditedResourceID: resourceID, Credited: resolved.Credited, ForfeitedUnits: resolved.ForfeitedUnits,
		CapReasonKey: resolved.CapReasonKey, FaucetApplied: resolved.FaucetApplied, CompanyRevision: resolved.CompanyRevision, FounderRevision: resolved.FounderRevision}))
	if err != nil {
		return nil, save.EventWrite{}, err
	}
	payload := mustJSON(map[string]any{"harvest_hash": resolved.HarvestHash, "selected_units": resolved.SelectedUnits, "credited": resolved.Credited,
		"forfeited_units": resolved.ForfeitedUnits, "cap_reason_key": resolved.CapReasonKey, "faucet_applied": resolved.FaucetApplied})
	return receipt, save.EventWrite{Kind: save.EventGardenHarvestCredited, SchemaVersion: 1, IntentID: resolved.IntentID, Payload: payload}, nil
}

func isGardenHarvestCreditPayload(data []byte) bool {
	var header struct {
		Kind string `json:"kind"`
	}
	return json.Unmarshal(data, &header) == nil && header.Kind == gardenCompanyCreditKind
}

// applyCompanyGardenHarvest is the Company run-log replay of
// credit_garden_harvest: the faucet arithmetic is re-derived from the pinned
// payout row, and the credit re-applied through the saturating ledger.
func applyCompanyGardenHarvest(state *save.State, canonicalPayload []byte, catalogs CatalogBundle, wire replayInputsWire) (LoggedTransition, error) {
	var payload gardenHarvestCreditPayload
	if decodeReplayStrict(canonicalPayload, &payload) != nil || payload.Kind != gardenCompanyCreditKind || payload.IntentID != wire.Command.IntentID || catalogs.Garden == nil {
		return LoggedTransition{}, ErrInvalidReplayInputs
	}
	normalized, err := normalizeReplayJSON(canonicalPayload)
	if err != nil || !bytes.Equal(normalized, canonicalPayload) {
		return LoggedTransition{}, ErrInvalidReplayInputs
	}
	var resolved gardenCompanyResolved
	policy := catalogs.Garden.Payout
	if decodeReplayStrict(wire.Resolved, &resolved) != nil || resolved.Kind != gardenCompanyCreditKind || resolved.IntentID != payload.IntentID ||
		resolved.HarvestHash != payload.HarvestHash || resolved.PolicyHash != gardenPolicyHash(policy) ||
		resolved.CompanyRevision != wire.Command.Revision+1 || resolved.FounderRevision != resolved.FounderLog.Revision ||
		resolved.FounderLog.Sequence < 1 || !founderAttendanceStreamPattern.MatchString(resolved.FounderLog.StreamID) ||
		resolved.SelectedUnits < 0 || resolved.FaucetApplied != (resolved.SelectedUnits > 0) || resolved.FaucetApplied != (resolved.Faucet != nil) {
		return LoggedTransition{}, ErrInvalidReplayInputs
	}
	credit := decimal.Zero
	expectedForfeit := int64(0)
	var expectedReason *string
	if resolved.Faucet != nil {
		if validatePersistentFaucetReplay(*resolved.Faucet, policy, resolved.SelectedUnits) != nil {
			return LoggedTransition{}, ErrInvalidReplayInputs
		}
		credit = decimal.FromString(strconv.FormatInt(resolved.Faucet.CreditedUnits, 10))
		expectedForfeit = resolved.Faucet.ForfeitedUnits
		if resolved.Faucet.CapReasonKey != "" {
			reason := resolved.Faucet.CapReasonKey
			expectedReason = &reason
		}
	}
	if resolved.ForfeitedUnits != expectedForfeit || (resolved.CapReasonKey == nil) != (expectedReason == nil) ||
		resolved.CapReasonKey != nil && *resolved.CapReasonKey != *expectedReason {
		return LoggedTransition{}, ErrInvalidReplayInputs
	}
	ledgerReceipt, err := state.Ledger.ApplyAccrual(economy.Transaction{Entries: []economy.Entry{{ResourceID: policy.CreditedResourceID, Delta: credit}}})
	if err != nil {
		return LoggedTransition{}, ErrInvalidReplayInputs
	}
	if creditedDelta, creditErr := minigameCreditedDelta(ledgerReceipt, policy.CreditedResourceID); creditErr != nil || creditedDelta != resolved.Credited {
		return LoggedTransition{}, ErrInvalidReplayInputs
	}
	receipt, event, err := gardenCompanyReceiptAndEvent(resolved, policy.CreditedResourceID)
	if err != nil {
		return LoggedTransition{}, err
	}
	return LoggedTransition{State: state, Outcome: save.IntentApplied, Receipt: receipt, Events: []save.EventWrite{event}}, nil
}

// validatePersistentFaucetReplay re-derives the window arithmetic for a solo
// persistent tenant (zero fallback rate reduction).
func validatePersistentFaucetReplay(value minigameFaucetWire, policy garden.PayoutPolicy, score int64) error {
	definition := minigame.Definition{Payout: gardenMinigamePolicy(policy)}
	return validateFaucetReplay(value, definition, score)
}

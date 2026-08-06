package production

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"cloud-clicker/server/economy"
	"cloud-clicker/server/minigame"
	"cloud-clicker/server/save"
	"cloud-clicker/server/soul"
)

type StartSoulRecoveryRequest struct {
	SessionID       string
	FounderID       string
	CompanyStreamID string
	ActivityID      string
}

// StartMinigameSession is the composed Soul gate in front of the authoritative
// minigame service. The client never supplies the Founder revision, Soul
// value, or constants hash used by this decision.
func (s *Service) StartMinigameSession(ctx context.Context, platform *minigame.Service,
	request minigame.StartRequest, now time.Time,
) (minigame.Session, error) {
	if s == nil || platform == nil || s.replayCatalogs == nil {
		return minigame.Session{}, ErrInvalidIntent
	}
	if s.soulRecoveries != nil {
		active, activeErr := s.soulRecoveries.HasActive(ctx, request.FounderID)
		if activeErr != nil {
			return minigame.Session{}, activeErr
		}
		if active {
			return minigame.Session{}, minigame.ErrExclusiveActivity
		}
	}
	company, err := s.store.LoadLatest(ctx, request.CompanyStreamID)
	if err != nil {
		return minigame.Session{}, err
	}
	founder, err := s.store.LoadSiblingLatest(ctx, request.CompanyStreamID, economy.ScopeFounder)
	if err != nil {
		return minigame.Session{}, err
	}
	if company.Key.OwnerID != request.FounderID || founder.Key.OwnerID != request.FounderID ||
		company.State.RunSeq != request.RunSeq || company.Revision.ConstantsHash != request.ConstantsHash ||
		founder.Revision.ConstantsHash != request.ConstantsHash {
		return minigame.Session{}, ErrInvalidIntent
	}
	bundle, ok := s.replayCatalogs.ResolveReplayCatalogs(request.ConstantsHash)
	if !ok || bundle.Minigames == nil {
		return minigame.Session{}, ErrInvalidIntent
	}
	definition, ok := bundle.Minigames.Definition(request.MinigameID)
	if !ok {
		return minigame.Session{}, ErrInvalidIntent
	}
	if definition.SoulGate == "human_hobby" {
		if bundle.Soul == nil || save.VersionForState(founder.State) < 20 {
			return minigame.Session{}, ErrInvalidIntent
		}
		locked, lockErr := bundle.Soul.HumanContentLocked(founder.State.Soul)
		if lockErr != nil {
			return minigame.Session{}, lockErr
		}
		if locked {
			return minigame.Session{}, fmt.Errorf("%w: human_content_locked", ErrInvalidIntent)
		}
	}
	_ = now // reserved for server-resolved session timing; never client input.
	return platform.Start(ctx, request)
}

type FinishSoulRecoveryRequest struct {
	SessionID string
	FounderID string
}

func (s *Service) applyExitTransactionLogged(ctx context.Context, companyStreamID string,
	expectedCompanyRevision, expectedFounderRevision int64, intentID, requestHash string,
	canonicalPayload []byte, mutate save.LoggedExitMutation,
) (save.IntentResult, error) {
	if s.soulRecoveries == nil {
		return s.store.ApplyExitTransactionLogged(ctx, companyStreamID, expectedCompanyRevision, expectedFounderRevision,
			intentID, requestHash, canonicalPayload, mutate, nil)
	}
	return s.store.ApplyExitTransactionLoggedGuarded(ctx, companyStreamID, expectedCompanyRevision, expectedFounderRevision,
		intentID, requestHash, canonicalPayload, s.soulRecoveries.RequireInactiveTx, mutate, nil)
}

type soulRecoveryReceipt struct {
	IntentID        string          `json:"intent_id"`
	Outcome         string          `json:"outcome"`
	Action          string          `json:"action"`
	SessionID       string          `json:"session_id"`
	ActivityID      string          `json:"activity_id"`
	CompanyRevision int64           `json:"company_revision"`
	FounderRevision int64           `json:"founder_revision"`
	SoulBefore      int64           `json:"soul_before"`
	SoulAfter       int64           `json:"soul_after"`
	BandBefore      soul.BandMember `json:"band_before"`
	BandAfter       soul.BandMember `json:"band_after"`
}

// StartSoulRecovery is a coordinator command, not a Company or Founder
// ApplyLogged kind. The repository atomically owns the session and start event.
func (s *Service) StartSoulRecovery(ctx context.Context, request StartSoulRecoveryRequest, now time.Time) (HandleResult, error) {
	if s == nil || s.soulRecoveries == nil || s.replayCatalogs == nil {
		return HandleResult{}, ErrInvalidIntent
	}
	company, err := s.store.LoadLatest(ctx, request.CompanyStreamID)
	if err != nil {
		return HandleResult{}, err
	}
	founder, err := s.store.LoadSiblingLatest(ctx, request.CompanyStreamID, economy.ScopeFounder)
	if err != nil {
		return HandleResult{}, err
	}
	if request.FounderID != company.Key.OwnerID || founder.Key.OwnerID != request.FounderID || save.VersionForState(founder.State) < 20 {
		return HandleResult{}, ErrInvalidIntent
	}
	bundle, ok := s.replayCatalogs.ResolveReplayCatalogs(company.Revision.ConstantsHash)
	if !ok || bundle.Soul == nil || founder.Revision.ConstantsHash != company.Revision.ConstantsHash {
		return HandleResult{}, ErrInvalidIntent
	}
	activity, ok := bundle.Soul.RecoveryActivity(request.ActivityID)
	if !ok {
		return HandleResult{}, ErrInvalidIntent
	}
	attendance, err := s.ResolveFounderAttendance(ctx, founder.Revision.StreamID, request.CompanyStreamID, now)
	if err != nil {
		return HandleResult{}, err
	}
	payload, _ := json.Marshal(map[string]any{"kind": "start_soul_recovery", "session_id": request.SessionID,
		"founder_id": request.FounderID, "company_stream_id": request.CompanyStreamID, "activity_id": request.ActivityID})
	hash := soulRequestHash(payload)
	session, err := s.soulRecoveries.Start(ctx, soul.StartRecovery{SessionID: request.SessionID, FounderID: request.FounderID,
		FounderStreamID: founder.Revision.StreamID, CompanyStreamID: request.CompanyStreamID, RunSeq: company.State.RunSeq,
		ConstantsHash: company.Revision.ConstantsHash, ActivityID: request.ActivityID,
		FounderAttendedStartMS: attendance.EffectiveFounderAttendedMS, RequiredDurationMS: activity.DurationAttendedMS,
		FounderRevision: founder.Revision.Number, CompanyRevision: company.Revision.Number, RequestHash: hash})
	if err != nil {
		return HandleResult{}, err
	}
	receipt, _ := json.Marshal(map[string]any{"intent_id": request.SessionID, "outcome": string(save.IntentApplied),
		"action": "start_soul_recovery", "session_id": request.SessionID, "activity_id": request.ActivityID,
		"founder_attended_start_ms": session.FounderAttendedStartMS, "required_duration_ms": session.RequiredDurationMS})
	return HandleResult{Receipt: receipt}, nil
}

func (s *Service) CancelSoulRecovery(ctx context.Context, request FinishSoulRecoveryRequest, now time.Time,
	fault save.ExitFaultInjector,
) (HandleResult, error) {
	return s.finishSoulRecovery(ctx, request, now, soulRecoveryCancelKind, fault)
}

func (s *Service) ResolveSoulRecovery(ctx context.Context, request FinishSoulRecoveryRequest, now time.Time,
	fault save.ExitFaultInjector,
) (HandleResult, error) {
	return s.finishSoulRecovery(ctx, request, now, soulRecoveryResolveKind, fault)
}

func (s *Service) finishSoulRecovery(ctx context.Context, request FinishSoulRecoveryRequest, now time.Time, kind string,
	fault save.ExitFaultInjector,
) (HandleResult, error) {
	if s == nil || s.soulRecoveries == nil || s.replayCatalogs == nil ||
		(kind != soulRecoveryResolveKind && kind != soulRecoveryCancelKind) {
		return HandleResult{}, ErrInvalidIntent
	}
	session, err := s.soulRecoveries.Load(ctx, request.FounderID, request.SessionID)
	if err != nil {
		return HandleResult{}, err
	}
	attendance, err := s.ResolveFounderAttendance(ctx, session.FounderStreamID, session.CompanyStreamID, now)
	if err != nil {
		return HandleResult{}, err
	}
	payload, _ := json.Marshal(soulRecoveryPayload{Kind: kind, SessionID: request.SessionID})
	hash := soulRequestHash(payload)
	result, err := s.store.ApplyMinigameResolutionTransaction(ctx, save.MinigameResolutionRequest{
		SessionID: request.SessionID, FounderID: request.FounderID, CompanyStreamID: session.CompanyStreamID,
		RequestHash: hash, CanonicalPayload: payload, ServerTSMS: save.CanonicalServerTime(now).UnixMilli(), CompanyLogFirst: true,
	}, func(ctx context.Context, tx *sql.Tx, founder *save.State, founderRevision save.Revision,
		company *save.State, companyRevision save.Revision, companyCommand save.ReplayCommand,
		founderCommand save.FounderReplayCommand, companyNext, founderNext int64,
	) (save.MinigameResolutionDecision, error) {
		claimed, claimErr := s.soulRecoveries.ClaimTx(ctx, tx, request.FounderID, request.SessionID, hash)
		if claimErr != nil {
			return save.MinigameResolutionDecision{}, claimErr
		}
		if claimed.Status == soul.RecoveryResolved || claimed.Status == soul.RecoveryCancelled {
			return save.MinigameResolutionDecision{}, soul.ErrRecoveryClaimLost
		}
		bundle, ok := s.replayCatalogs.ResolveReplayCatalogs(companyRevision.ConstantsHash)
		if !ok || bundle.Soul == nil || claimed.ConstantsHash != companyRevision.ConstantsHash ||
			claimed.FounderStreamID != founderRevision.StreamID || claimed.CompanyStreamID != companyRevision.StreamID ||
			claimed.RunSeq != company.RunSeq || attendance.CompanyRevision != companyRevision.Number ||
			attendance.CompanyConstantsHash != companyRevision.ConstantsHash ||
			ValidateFounderAttendanceSample(founder, founderRevision.Number, founderRevision.Number, attendance) != nil {
			return save.MinigameResolutionDecision{}, ErrFounderAttendanceStale
		}
		activity, ok := bundle.Soul.RecoveryActivity(claimed.ActivityID)
		if !ok || activity.DurationAttendedMS != claimed.RequiredDurationMS {
			return save.MinigameResolutionDecision{}, ErrInvalidIntent
		}
		if kind == soulRecoveryResolveKind && attendance.EffectiveFounderAttendedMS-claimed.FounderAttendedStartMS < claimed.RequiredDurationMS {
			return save.MinigameResolutionDecision{}, fmt.Errorf("%w: soul_recovery_not_ready", ErrInvalidIntent)
		}
		var commonsWeight *int64
		if company.CompactMember {
			weight, weightErr := s.resolveCommonsReplayWeight(ctx, companyRevision.StreamID, companyRevision.OwnerID, companyRevision.ConstantsHash)
			if weightErr != nil {
				return save.MinigameResolutionDecision{}, weightErr
			}
			commonsWeight = &weight
		}
		suppression := soulSuppression{FromEvaluatedMS: company.EvaluatedThrough.UnixMilli(),
			ToEvaluatedMS: save.CanonicalServerTime(now).UnixMilli(), FounderAttendedStart: claimed.FounderAttendedStartMS,
			FounderAttendedEnd: attendance.EffectiveFounderAttendedMS, SessionID: claimed.SessionID}
		var activeEvidence *activePlayScheduleEvidence
		if company.WireVersion == 18 {
			resolved, activeErr := resolveActivePlaySchedule(company, bundle.Opportunities, bundle.Prestige,
				companyRevision.OwnerID, now)
			if activeErr != nil {
				return save.MinigameResolutionDecision{}, activeErr
			}
			activeEvidence = &resolved
		}
		companyInputs, inputErr := buildSoulSuppressionInputs(companyCommand, kind, suppression, nil, commonsWeight,
			bundle.Routes.ContextVersion(), activeEvidence)
		if inputErr != nil {
			return save.MinigameResolutionDecision{}, inputErr
		}
		companyTransition, transitionErr := ApplySuppressedLogged(company, payload, bundle, companyInputs)
		if transitionErr != nil {
			return save.MinigameResolutionDecision{}, transitionErr
		}
		beforeSoul := founder.Soul
		beforeBand, ok := bundle.Soul.BandFor(beforeSoul)
		if !ok {
			return save.MinigameResolutionDecision{}, ErrInvalidIntent
		}
		afterSoul := beforeSoul
		if kind == soulRecoveryResolveKind {
			if activity.RecoveryAmount > bundle.Soul.Policy.Max-afterSoul {
				afterSoul = bundle.Soul.Policy.Max
			} else {
				afterSoul += activity.RecoveryAmount
			}
		}
		afterBand, ok := bundle.Soul.BandFor(afterSoul)
		if !ok {
			return save.MinigameResolutionDecision{}, ErrInvalidIntent
		}
		founderResolved := founderSoulRecoveryResolved{Kind: "soul_recovery", Action: kind, SessionID: claimed.SessionID,
			ActivityID: claimed.ActivityID, CompanyStreamID: claimed.CompanyStreamID, RunSeq: claimed.RunSeq,
			FounderAttendedStartMS: claimed.FounderAttendedStartMS,
			FounderAttendedEndMS:   attendance.EffectiveFounderAttendedMS, RecoveryAmount: activity.RecoveryAmount,
			SoulBefore: beforeSoul, SoulAfter: afterSoul, BandBefore: beforeBand.Member, BandAfter: afterBand.Member,
			ReasonKey: activity.ReasonKey}
		founderInputs, marshalErr := save.MarshalFounderReplayInputs(founderCommand, founderResolved)
		if marshalErr != nil {
			return save.MinigameResolutionDecision{}, marshalErr
		}
		founderTransition, transitionErr := ApplyFounderLogged(founder, payload, bundle, founderInputs)
		if transitionErr != nil {
			return save.MinigameResolutionDecision{}, transitionErr
		}
		receipt := soulRecoveryReceipt{IntentID: request.SessionID, Outcome: string(save.IntentApplied), Action: kind,
			SessionID: claimed.SessionID, ActivityID: claimed.ActivityID, CompanyRevision: companyNext,
			FounderRevision: founderNext, SoulBefore: beforeSoul, SoulAfter: afterSoul,
			BandBefore: beforeBand.Member, BandAfter: afterBand.Member}
		receiptBytes, _ := json.Marshal(receipt)
		terminalStatus := soul.RecoveryResolved
		if kind == soulRecoveryCancelKind {
			terminalStatus = soul.RecoveryCancelled
		}
		if _, finishErr := s.soulRecoveries.FinishTx(ctx, tx, claimed.SessionID, claimed.ClaimToken,
			terminalStatus, hash, receiptBytes); finishErr != nil {
			return save.MinigameResolutionDecision{}, finishErr
		}
		if fault != nil {
			if faultErr := fault("soul_session_terminal"); faultErr != nil {
				return save.MinigameResolutionDecision{}, faultErr
			}
		}
		founderEvents := append([]save.EventWrite(nil), founderTransition.Events...)
		return save.MinigameResolutionDecision{Receipt: receiptBytes, FounderReceipt: founderTransition.Receipt,
			CompanyReplayInputs: companyInputs, FounderReplayResolved: mustJSON(founderResolved),
			CompanyEvents: companyTransition.Events, FounderEvents: founderEvents}, nil
	}, fault)
	if err != nil {
		if errors.Is(err, soul.ErrRecoveryIdempotency) {
			return HandleResult{}, save.ErrIdempotencyConflict
		}
		return HandleResult{}, err
	}
	if err := s.projectCommittedEvents(ctx, result.Events); err != nil {
		return HandleResult{}, err
	}
	return HandleResult{Receipt: result.Receipt, Replay: result.Replay}, nil
}

func soulRequestHash(payload []byte) string {
	digest := sha256.Sum256(payload)
	return "sha256:" + hex.EncodeToString(digest[:])
}

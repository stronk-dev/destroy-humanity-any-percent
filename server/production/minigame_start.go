package production

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"cloud-clicker/server/decimal"
	"cloud-clicker/server/determinism"
	"cloud-clicker/server/minigame"
	"cloud-clicker/server/runidentity"
	"cloud-clicker/server/save"
)

var (
	ErrMinigameFiscalUnlockRequired = errors.New("minigame fiscal unlock required")
	ErrMinigameHumanContentLocked   = errors.New("minigame human content locked")
)

const startMinigameSessionKind = "start_minigame_session"

type StartMinigameAPIRequest struct {
	SessionID       string
	IntentID        string
	FounderID       string
	CompanyStreamID string
	MinigameID      string
	IdempotencyKey  string
}

type startMinigameSessionPayload struct {
	Kind       string `json:"kind"`
	SessionID  string `json:"session_id"`
	MinigameID string `json:"minigame_id"`
}

type startMinigameSessionResolved struct {
	Kind            string `json:"kind"`
	CompanyStreamID string `json:"company_stream_id"`
	RunSeq          int64  `json:"run_seq"`
	SequenceBefore  int64  `json:"sequence_before"`
	SequenceAfter   int64  `json:"sequence_after"`
	Seed            string `json:"seed"`
}

func isStartMinigameSessionPayload(data []byte) bool {
	var header struct {
		Kind string `json:"kind"`
	}
	return json.Unmarshal(data, &header) == nil && header.Kind == startMinigameSessionKind
}

func applyFounderStartMinigameSession(state *save.State, canonicalPayload []byte, catalogs CatalogBundle,
	wire founderReplayInputsWire,
) (FounderLoggedTransition, error) {
	var payload startMinigameSessionPayload
	var resolved startMinigameSessionResolved
	if decodeReplayStrict(canonicalPayload, &payload) != nil || decodeReplayStrict(wire.Resolved, &resolved) != nil ||
		payload.Kind != startMinigameSessionKind ||
		resolved.Kind != startMinigameSessionKind || resolved.CompanyStreamID == "" || resolved.RunSeq < 1 ||
		resolved.RunSeq > decimal.MaxExactInteger || resolved.SequenceBefore < 0 ||
		resolved.SequenceBefore >= decimal.MaxExactInteger || resolved.SequenceAfter != resolved.SequenceBefore+1 ||
		state.MinigameSessionSeq != resolved.SequenceBefore || save.VersionForState(state) < 21 ||
		catalogs.MinigameAPI == nil || catalogs.Minigames == nil {
		return FounderLoggedTransition{}, fmt.Errorf("%w: minigame start inputs", ErrInvalidReplayInputs)
	}
	definition, ok := catalogs.Minigames.Definition(payload.MinigameID)
	if !ok || !catalogs.MinigameAPI.SupportsTenant(payload.MinigameID, definition.EngineRef, definition.EngineVersion) {
		return FounderLoggedTransition{}, fmt.Errorf("%w: minigame start tenant", ErrInvalidReplayInputs)
	}
	wantSeed := minigameSessionSeed(wire.Command.FounderID, resolved.RunSeq, resolved.SequenceAfter)
	if resolved.Seed != wantSeed {
		return FounderLoggedTransition{}, fmt.Errorf("%w: minigame start seed", ErrInvalidReplayInputs)
	}
	state.MinigameSessionSeq = resolved.SequenceAfter
	receipt, _ := json.Marshal(map[string]any{
		"founder_revision": wire.Command.Revision + 1,
		"intent_id":        wire.Command.IntentID,
		"minigame_id":      payload.MinigameID,
		"outcome":          string(save.IntentApplied),
		"seed":             resolved.Seed,
		"sequence_after":   resolved.SequenceAfter,
		"sequence_before":  resolved.SequenceBefore,
		"session_id":       payload.SessionID,
	})
	return FounderLoggedTransition{State: state, Outcome: save.IntentApplied, Receipt: receipt,
		Events: []save.EventWrite{}, ResultConstantsHash: catalogs.ConstantsHash}, nil
}

func minigameSessionSeed(founderID string, runSeq, sequence int64) string {
	seed := runidentity.Seed(founderID, runSeq) ^ uint64(sequence)
	return strconv.FormatUint(determinism.Substream(seed, "minigame.session.v1").Next(), 10)
}

// StartMinigameAPISession resolves every client-hidden input and commits the
// Founder v21 sequence, immutable Founder log, tenant genesis/session, and
// create-idempotency response in one transaction.
func (s *Service) StartMinigameAPISession(ctx context.Context, platform *minigame.Service,
	request StartMinigameAPIRequest, _ time.Time, fault save.ExitFaultInjector,
) (save.IntentResult, error) {
	if s == nil || s.store == nil || s.replayCatalogs == nil || platform == nil {
		return save.IntentResult{}, ErrInvalidIntent
	}
	requestIdentity, err := json.Marshal(map[string]any{
		"idempotency_key": request.IdempotencyKey,
		"minigame_id":     request.MinigameID,
	})
	if err != nil {
		return save.IntentResult{}, ErrInvalidIntent
	}
	canonicalPayload, err := json.Marshal(startMinigameSessionPayload{Kind: startMinigameSessionKind,
		SessionID: request.SessionID, MinigameID: request.MinigameID})
	if err != nil {
		return save.IntentResult{}, ErrInvalidIntent
	}
	canonicalPayload, err = normalizeReplayJSON(canonicalPayload)
	if err != nil {
		return save.IntentResult{}, ErrInvalidIntent
	}
	return s.store.ApplyMinigameStartTransaction(ctx, save.MinigameStartRequest{
		SessionID: request.SessionID, IntentID: request.IntentID, FounderID: request.FounderID, CompanyStreamID: request.CompanyStreamID,
		IdempotencyKey: request.IdempotencyKey, RequestHash: soulRequestHash(requestIdentity),
		CanonicalPayload: canonicalPayload,
	}, func(ctx context.Context, tx *sql.Tx, founder *save.State, founderRevision save.Revision,
		company *save.State, companyRevision save.Revision, founderCommand save.FounderReplayCommand,
		_ int64,
	) (save.MinigameStartDecision, error) {
		bundle, ok := s.replayCatalogs.ResolveReplayCatalogs(companyRevision.ConstantsHash)
		if !ok || bundle.MinigameAPI == nil || bundle.Minigames == nil || bundle.Pitch == nil ||
			founderRevision.ConstantsHash != companyRevision.ConstantsHash || company.RunSeq < 1 ||
			save.VersionForState(founder) < 21 {
			return save.MinigameStartDecision{}, ErrInvalidIntent
		}
		definition, ok := bundle.Minigames.Definition(request.MinigameID)
		if !ok || !bundle.MinigameAPI.SupportsTenant(request.MinigameID, definition.EngineRef, definition.EngineVersion) ||
			len(definition.Modes) == 0 || definition.Modes[0] != minigame.ModeSolo {
			return save.MinigameStartDecision{}, ErrInvalidIntent
		}
		if definition.Unlock.Kind == "fiscal_unlock" {
			if bundle.Fiscal == nil {
				return save.MinigameStartDecision{}, ErrInvalidIntent
			}
			if _, declared := bundle.Fiscal.Unlock(definition.Unlock.UnlockID); !declared || !founder.FiscalUnlocks[definition.Unlock.UnlockID] {
				return save.MinigameStartDecision{}, fmt.Errorf("%w: %w", ErrInvalidIntent, ErrMinigameFiscalUnlockRequired)
			}
		}
		if definition.SoulGate == "human_hobby" {
			if bundle.Soul == nil || save.VersionForState(founder) < 20 {
				return save.MinigameStartDecision{}, ErrInvalidIntent
			}
			locked, lockErr := bundle.Soul.HumanContentLocked(founder.Soul)
			if lockErr != nil {
				return save.MinigameStartDecision{}, lockErr
			}
			if locked {
				return save.MinigameStartDecision{}, fmt.Errorf("%w: %w", ErrInvalidIntent, ErrMinigameHumanContentLocked)
			}
		}
		scaling, scalingErr := definition.Scaling.Resolve(minigame.ScalingContext{
			Tier: company.Tier, PurchasedGeneratorCounts: company.GeneratorCounts,
			FounderCarryCounters: map[string]int64{
				"achievement_score_lifetime": founder.AchievementScoreLifetime,
				"age_ms":                     founder.AgeMS,
				"exit_history_count":         int64(len(founder.ExitHistory)),
				"notoriety":                  founder.Notoriety,
				"reputation_level":           founder.ReputationLevel,
				"route_knowledge_balance":    founder.RouteKnowledgeBalance,
			},
			AttendedQualityGrades: minigameQualityGrades(founder.MinigameOfflineQuality),
		})
		if scalingErr != nil {
			return save.MinigameStartDecision{}, scalingErr
		}
		before := founder.MinigameSessionSeq
		if before < 0 || before >= decimal.MaxExactInteger {
			return save.MinigameStartDecision{}, ErrInvalidIntent
		}
		after := before + 1
		seed := minigameSessionSeed(request.FounderID, company.RunSeq, after)
		resolved := startMinigameSessionResolved{Kind: startMinigameSessionKind,
			CompanyStreamID: companyRevision.StreamID, RunSeq: company.RunSeq,
			SequenceBefore: before, SequenceAfter: after, Seed: seed}
		resolvedBytes, _ := json.Marshal(resolved)
		founderInputs, inputsErr := save.MarshalFounderReplayInputs(founderCommand, resolved)
		if inputsErr != nil {
			return save.MinigameStartDecision{}, inputsErr
		}
		transition, transitionErr := ApplyFounderLogged(founder, canonicalPayload, bundle, founderInputs)
		if transitionErr != nil || transition.Outcome != save.IntentApplied || founder.MinigameSessionSeq != after {
			if transitionErr != nil {
				return save.MinigameStartDecision{}, transitionErr
			}
			return save.MinigameStartDecision{}, ErrInvalidIntent
		}
		prepared, prepareErr := platform.PrepareStart(minigame.StartRequest{
			SessionID: request.SessionID, MinigameID: request.MinigameID, FounderID: request.FounderID,
			CompanyStreamID: companyRevision.StreamID, RunSeq: company.RunSeq, EngineRef: definition.EngineRef,
			EngineVersion: definition.EngineVersion, ConstantsHash: companyRevision.ConstantsHash,
			ScalingInputs: scaling, Seed: seed, Mode: minigame.ModeSolo,
		})
		if prepareErr != nil {
			return save.MinigameStartDecision{}, prepareErr
		}
		session, createErr := minigame.CreateTx(ctx, tx, prepared)
		if createErr != nil {
			return save.MinigameStartDecision{}, createErr
		}
		apiReceipt, marshalErr := normalizeReplayJSON(mustJSON(map[string]any{
			"constants_hash": session.ConstantsHash,
			"engine_ref":     session.EngineRef,
			"engine_version": session.EngineVersion,
			"minigame_id":    session.MinigameID,
			"mode":           session.Mode,
			"revision":       session.Revision,
			"session_id":     session.SessionID,
			"snapshot":       session.State,
			"status":         session.Status,
		}))
		if marshalErr != nil {
			return save.MinigameStartDecision{}, marshalErr
		}
		if insertErr := minigame.InsertCreateReceiptTx(ctx, tx, request.FounderID, request.IdempotencyKey,
			soulRequestHash(requestIdentity), request.SessionID, apiReceipt); insertErr != nil {
			return save.MinigameStartDecision{}, insertErr
		}
		return save.MinigameStartDecision{Receipt: apiReceipt, FounderReceipt: transition.Receipt,
			FounderReplayResolved: resolvedBytes, FounderEvents: transition.Events}, nil
	}, fault)
}

func minigameQualityGrades(values map[string]save.MinigameOfflineQualityState) map[string]int64 {
	result := make(map[string]int64, len(values))
	for id, value := range values {
		result[id] = value.GradePPM
	}
	return result
}

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
	"io"
	"reflect"
	"strconv"
	"time"

	"cloud-clicker/server/decimal"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/minigame"
	"cloud-clicker/server/save"
)

const minigameResolutionKind = "resolve_minigame_session"

type minigameResolutionPayload struct {
	Kind      string          `json:"kind"`
	SessionID string          `json:"session_id"`
	Result    json.RawMessage `json:"result"`
}

type minigameFaucetWire struct {
	AttendedDay        int64  `json:"attended_day"`
	QuotaBefore        int64  `json:"quota_before"`
	QuotaAfter         int64  `json:"quota_after"`
	RemainderBeforePPM int64  `json:"remainder_before_ppm"`
	RemainderAfterPPM  int64  `json:"remainder_after_ppm"`
	ReducedScore       int64  `json:"reduced_score"`
	ConvertedUnits     int64  `json:"converted_units"`
	CreditedUnits      int64  `json:"credited_units"`
	ForfeitedUnits     int64  `json:"forfeited_units"`
	CapReasonKey       string `json:"cap_reason_key"`
}

type minigameLogCoordinate struct {
	StreamID string `json:"stream_id"`
	Revision int64  `json:"revision"`
	Sequence int64  `json:"sequence"`
}

type minigameCompanyResolved struct {
	Kind                string                       `json:"kind"`
	SessionID           string                       `json:"session_id"`
	MinigameID          string                       `json:"minigame_id"`
	CertifiedResultHash string                       `json:"certified_result_hash"`
	PayoutPolicy        minigame.PayoutPolicy        `json:"payout_policy"`
	SelectedScore       int64                        `json:"selected_score"`
	Faucet              minigameFaucetWire           `json:"faucet"`
	CreditedDelta       string                       `json:"credited_delta"`
	FounderLog          minigameLogCoordinate        `json:"founder_log"`
	CompanyRevision     int64                        `json:"company_revision"`
	FounderRevision     int64                        `json:"founder_revision"`
	RatingChange        minigameRatingChangeReceipt  `json:"rating_change"`
	QualityChange       minigameQualityChangeReceipt `json:"quality_change"`
}

type minigameFounderResolved struct {
	Kind                string                           `json:"kind"`
	SessionID           string                           `json:"session_id"`
	MinigameID          string                           `json:"minigame_id"`
	CertifiedResultHash string                           `json:"certified_result_hash"`
	RatingBefore        save.MinigameRatingState         `json:"rating_before"`
	RatingAfter         save.MinigameRatingState         `json:"rating_after"`
	QualityBefore       save.MinigameOfflineQualityState `json:"quality_before"`
	QualityAfter        save.MinigameOfflineQualityState `json:"quality_after"`
	Attendance          FounderAttendanceSample          `json:"attendance"`
}

type minigameRatingChangeReceipt struct {
	Rated        bool   `json:"rated"`
	OldElo       int64  `json:"old_elo"`
	NewElo       int64  `json:"new_elo"`
	SeasonMember string `json:"season_member"`
	GamesBefore  int64  `json:"games_before"`
	GamesAfter   int64  `json:"games_after"`
}

type minigameQualityChangeReceipt struct {
	Old save.MinigameOfflineQualityState `json:"old"`
	New save.MinigameOfflineQualityState `json:"new"`
}

type minigameResolutionReceipt struct {
	IntentID            string                       `json:"intent_id"`
	Outcome             string                       `json:"outcome"`
	SessionID           string                       `json:"session_id"`
	MinigameID          string                       `json:"minigame_id"`
	CertifiedResultHash string                       `json:"certified_result_hash"`
	CompanyRevision     int64                        `json:"company_revision"`
	FounderRevision     int64                        `json:"founder_revision"`
	CreditedResourceID  string                       `json:"credited_resource_id"`
	CreditedDelta       string                       `json:"credited_delta"`
	ConfiguredForfeit   int64                        `json:"configured_cap_forfeit_units"`
	CapReasonKey        string                       `json:"cap_reason_key"`
	RatingChange        minigameRatingChangeReceipt  `json:"rating_change"`
	QualityChange       minigameQualityChangeReceipt `json:"quality_change"`
}

type minigameFounderReceipt struct {
	IntentID            string                       `json:"intent_id"`
	Outcome             string                       `json:"outcome"`
	FounderRevision     int64                        `json:"founder_revision"`
	SessionID           string                       `json:"session_id"`
	CertifiedResultHash string                       `json:"certified_result_hash"`
	RatingChange        minigameRatingChangeReceipt  `json:"rating_change"`
	QualityChange       minigameQualityChangeReceipt `json:"quality_change"`
}

// ResolveMinigameSession composes the already-certified tenant result. It is
// server-only: no public intent decoder constructs minigameResolutionPayload.
func (s *Service) ResolveMinigameSession(ctx context.Context, platform *minigame.Service,
	resolution *minigame.CertifiedResolution, now time.Time, fault save.ExitFaultInjector,
) (HandleResult, error) {
	if s == nil || s.store == nil || s.replayCatalogs == nil || platform == nil || resolution == nil {
		return HandleResult{}, ErrInvalidIntent
	}
	view, err := resolution.View()
	if err != nil {
		return HandleResult{}, err
	}
	bundle, ok := s.replayCatalogs.ResolveReplayCatalogs(view.ConstantsHash)
	if !ok || bundle.Minigames == nil {
		return HandleResult{}, fmt.Errorf("%w: pinned minigame policy unavailable", ErrInvalidIntent)
	}
	definition, ok := bundle.Minigames.Definition(view.MinigameID)
	if !ok {
		return HandleResult{}, fmt.Errorf("%w: minigame is not pinned", ErrInvalidIntent)
	}
	founderLoaded, err := s.store.LoadSiblingLatest(ctx, view.CompanyStreamID, economy.ScopeFounder)
	if err != nil {
		return HandleResult{}, err
	}
	attendance, err := s.ResolveFounderAttendance(ctx, founderLoaded.Revision.StreamID, view.CompanyStreamID, now)
	if err != nil {
		return HandleResult{}, err
	}
	payload, err := json.Marshal(minigameResolutionPayload{Kind: minigameResolutionKind, SessionID: view.SessionID, Result: view.ResultBytes})
	if err != nil {
		return HandleResult{}, err
	}
	payload, err = normalizeReplayJSON(payload)
	if err != nil {
		return HandleResult{}, err
	}
	digest := sha256.Sum256(payload)
	requestHash := "sha256:" + hex.EncodeToString(digest[:])
	result, err := s.store.ApplyMinigameResolutionTransaction(ctx, save.MinigameResolutionRequest{
		SessionID: view.SessionID, FounderID: view.FounderID, CompanyStreamID: view.CompanyStreamID,
		RequestHash: requestHash, CanonicalPayload: payload,
	}, func(ctx context.Context, tx *sql.Tx, founder *save.State, founderRevision save.Revision,
		company *save.State, companyRevision save.Revision, companyCommand save.ReplayCommand,
		founderCommand save.FounderReplayCommand, companyNext, founderNext int64,
	) (save.MinigameResolutionDecision, error) {
		if companyRevision.ConstantsHash != view.ConstantsHash || company.RunSeq != view.RunSeq ||
			attendance.CompanyRevision != companyRevision.Number || attendance.CompanyConstantsHash != companyRevision.ConstantsHash ||
			ValidateFounderAttendanceSample(founder, founderRevision.Number, founderRevision.Number, attendance) != nil {
			return save.MinigameResolutionDecision{}, ErrFounderAttendanceStale
		}
		prepared, prepareErr := platform.PrepareResolutionTx(ctx, tx, resolution, definition, attendance.EffectiveFounderAttendedMS)
		if prepareErr != nil {
			return save.MinigameResolutionDecision{}, fmt.Errorf("prepare certified minigame resolution: %w", prepareErr)
		}
		if fault != nil {
			if faultErr := fault("faucet_window"); faultErr != nil {
				return save.MinigameResolutionDecision{}, faultErr
			}
		}
		oldRating, ratingOK := founder.MinigameRatings[definition.MinigameID]
		oldQuality, qualityOK := founder.MinigameOfflineQuality[definition.MinigameID]
		if !ratingOK || !qualityOK {
			return save.MinigameResolutionDecision{}, fmt.Errorf("%w: inactive minigame Founder state", ErrInvalidIntent)
		}
		transition, transitionErr := minigame.ApplyFounderResolution(
			minigame.RatingState{Elo: oldRating.Elo, SeasonMember: oldRating.SeasonMember, GamesCounted: oldRating.GamesCounted},
			minigame.OfflineQualityState{GradePPM: oldQuality.GradePPM, LastFounderAttendedMS: oldQuality.LastFounderAttendedMS, DecayRemainderPPM: oldQuality.DecayRemainderPPM},
			view.Result, definition, attendance.EffectiveFounderAttendedMS)
		if transitionErr != nil {
			return save.MinigameResolutionDecision{}, transitionErr
		}
		newRating := save.MinigameRatingState{Elo: transition.RatingAfter.Elo, SeasonMember: transition.RatingAfter.SeasonMember, GamesCounted: transition.RatingAfter.GamesCounted}
		newQuality := save.MinigameOfflineQualityState{GradePPM: transition.QualityAfter.GradePPM,
			LastFounderAttendedMS: transition.QualityAfter.LastFounderAttendedMS, DecayRemainderPPM: transition.QualityAfter.DecayRemainderPPM}
		founder.MinigameRatings[definition.MinigameID] = newRating
		founder.MinigameOfflineQuality[definition.MinigameID] = newQuality
		requested := decimal.FromString(strconv.FormatInt(prepared.Faucet.CreditedUnits, 10))
		ledgerReceipt, ledgerErr := company.Ledger.ApplyAccrual(economy.Transaction{Entries: []economy.Entry{{
			ResourceID: definition.Payout.CreditedResourceID, Delta: requested,
		}}})
		if ledgerErr != nil || len(ledgerReceipt.Changes) != 1 {
			return save.MinigameResolutionDecision{}, ledgerErr
		}
		creditedDelta := ledgerReceipt.Changes[0].Delta
		certifiedHash := certifiedResultHash(view.ResultBytes)
		ratingChange := minigameRatingChangeReceipt{Rated: view.Result.RatingDelta != nil,
			OldElo: oldRating.Elo, NewElo: newRating.Elo, SeasonMember: newRating.SeasonMember,
			GamesBefore: oldRating.GamesCounted, GamesAfter: newRating.GamesCounted}
		qualityChange := minigameQualityChangeReceipt{Old: oldQuality, New: newQuality}
		receipt := minigameResolutionReceipt{IntentID: view.SessionID, Outcome: string(save.IntentApplied),
			SessionID: view.SessionID, MinigameID: definition.MinigameID, CertifiedResultHash: certifiedHash,
			CompanyRevision: companyNext, FounderRevision: founderNext,
			CreditedResourceID: definition.Payout.CreditedResourceID, CreditedDelta: creditedDelta,
			ConfiguredForfeit: prepared.Faucet.ForfeitedUnits, CapReasonKey: prepared.Faucet.ConfiguredCapReasonKey,
			RatingChange: ratingChange, QualityChange: qualityChange}
		receiptBytes, marshalErr := json.Marshal(receipt)
		if marshalErr != nil {
			return save.MinigameResolutionDecision{}, marshalErr
		}
		receiptBytes, marshalErr = normalizeReplayJSON(receiptBytes)
		if marshalErr != nil {
			return save.MinigameResolutionDecision{}, marshalErr
		}
		if _, finalizeErr := platform.FinalizeResolutionTx(ctx, tx, resolution, receiptBytes, companyNext, founderNext); finalizeErr != nil {
			return save.MinigameResolutionDecision{}, fmt.Errorf("finalize certified minigame resolution: %w", finalizeErr)
		}
		if fault != nil {
			if faultErr := fault("session_terminal"); faultErr != nil {
				return save.MinigameResolutionDecision{}, faultErr
			}
		}
		selectedScore, _ := minigame.SelectPayoutScore(view.Result, definition.Payout)
		faucet := faucetWire(prepared.Faucet)
		companyResolved := minigameCompanyResolved{Kind: minigameResolutionKind, SessionID: view.SessionID,
			MinigameID: definition.MinigameID, CertifiedResultHash: certifiedHash, PayoutPolicy: definition.Payout,
			SelectedScore: selectedScore, Faucet: faucet, CreditedDelta: creditedDelta,
			FounderLog:      minigameLogCoordinate{StreamID: founderCommand.FounderStreamID, Revision: founderNext, Sequence: founderCommand.FounderLogSeq},
			CompanyRevision: companyNext, FounderRevision: founderNext, RatingChange: ratingChange, QualityChange: qualityChange}
		companyInputs, marshalErr := json.Marshal(replayInputsWire{Version: save.ReplayInputsVersion, Command: companyCommand,
			EvaluatedAtMS: founderCommand.ServerTSMS, EvaluationMode: ModeOnline, Resolved: mustJSON(companyResolved)})
		if marshalErr != nil {
			return save.MinigameResolutionDecision{}, marshalErr
		}
		founderResolved := minigameFounderResolved{Kind: minigameResolutionKind, SessionID: view.SessionID,
			MinigameID: definition.MinigameID, CertifiedResultHash: certifiedHash,
			RatingBefore: oldRating, RatingAfter: newRating, QualityBefore: oldQuality, QualityAfter: newQuality, Attendance: attendance}
		founderReceiptBytes, _ := json.Marshal(minigameFounderReceipt{IntentID: view.SessionID, Outcome: string(save.IntentApplied),
			FounderRevision: founderNext, SessionID: view.SessionID, CertifiedResultHash: certifiedHash,
			RatingChange: ratingChange, QualityChange: qualityChange})
		companyEventPayload, _ := json.Marshal(map[string]any{"session_id": view.SessionID, "minigame_id": definition.MinigameID,
			"certified_result_hash": certifiedHash, "credited_resource_id": definition.Payout.CreditedResourceID,
			"credited_delta": creditedDelta, "configured_cap_forfeit_units": prepared.Faucet.ForfeitedUnits,
			"cap_reason_key": prepared.Faucet.ConfiguredCapReasonKey, "founder_revision": founderNext})
		founderEventPayload, _ := json.Marshal(map[string]any{"session_id": view.SessionID, "minigame_id": definition.MinigameID,
			"certified_result_hash": certifiedHash, "old_elo": oldRating.Elo, "new_elo": newRating.Elo,
			"season_member": newRating.SeasonMember, "old_quality": oldQuality, "new_quality": newQuality})
		return save.MinigameResolutionDecision{Receipt: receiptBytes, FounderReceipt: founderReceiptBytes,
			CompanyReplayInputs: companyInputs, FounderReplayResolved: mustJSON(founderResolved),
			CompanyEvents: []save.EventWrite{{Kind: save.EventMinigameResolved, SchemaVersion: 1, IntentID: view.SessionID, Payload: companyEventPayload}},
			FounderEvents: []save.EventWrite{{Kind: save.EventMinigameRatingChanged, SchemaVersion: 1, IntentID: view.SessionID, Payload: founderEventPayload}}}, nil
	}, fault)
	if err != nil {
		return HandleResult{}, err
	}
	if err := s.projectCommittedEvents(ctx, result.Events); err != nil {
		return HandleResult{}, err
	}
	return HandleResult{Receipt: result.Receipt, Replay: result.Replay}, nil
}

func certifiedResultHash(data []byte) string {
	digest := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(digest[:])
}

func faucetWire(value minigame.FaucetApplication) minigameFaucetWire {
	return minigameFaucetWire{AttendedDay: value.AttendedDay, QuotaBefore: value.QuotaBefore,
		QuotaAfter: value.QuotaAfter, RemainderBeforePPM: value.RemainderBeforePPM,
		RemainderAfterPPM: value.RemainderAfterPPM, ReducedScore: value.ReducedScore,
		ConvertedUnits: value.ConvertedUnits, CreditedUnits: value.CreditedUnits,
		ForfeitedUnits: value.ForfeitedUnits, CapReasonKey: value.ConfiguredCapReasonKey}
}

func mustJSON(value any) json.RawMessage {
	encoded, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return encoded
}

func parseMinigameResolutionPayload(data []byte) (minigameResolutionPayload, error) {
	var payload minigameResolutionPayload
	if err := decodeReplayStrict(data, &payload); err != nil || payload.Kind != minigameResolutionKind ||
		!intentUUIDV7Pattern.MatchString(payload.SessionID) || len(payload.Result) < 2 || payload.Result[0] != '{' {
		return payload, ErrInvalidReplayInputs
	}
	normalized, err := normalizeReplayJSON(payload.Result)
	if err != nil || !bytes.Equal(normalized, payload.Result) {
		return payload, ErrInvalidReplayInputs
	}
	return payload, nil
}

func normalizeReplayJSON(data []byte) ([]byte, error) {
	var value any
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	if err := decoder.Decode(&value); err != nil {
		return nil, err
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return nil, ErrInvalidReplayInputs
	}
	return json.Marshal(value)
}

func isMinigameResolutionPayload(data []byte) bool {
	var header struct {
		Kind string `json:"kind"`
	}
	return json.Unmarshal(data, &header) == nil && header.Kind == minigameResolutionKind
}

func applyCompanyMinigameResolution(state *save.State, canonicalPayload []byte, catalogs CatalogBundle,
	wire replayInputsWire,
) (LoggedTransition, error) {
	payload, err := parseMinigameResolutionPayload(canonicalPayload)
	if err != nil || payload.SessionID != wire.Command.IntentID || catalogs.Minigames == nil {
		return LoggedTransition{}, ErrInvalidReplayInputs
	}
	var result minigame.Result
	if decodeReplayStrict(payload.Result, &result) != nil {
		return LoggedTransition{}, ErrInvalidReplayInputs
	}
	var resolved minigameCompanyResolved
	if decodeReplayStrict(wire.Resolved, &resolved) != nil || resolved.Kind != minigameResolutionKind ||
		resolved.SessionID != payload.SessionID || resolved.CertifiedResultHash != certifiedResultHash(payload.Result) ||
		resolved.CompanyRevision != wire.Command.Revision+1 || resolved.FounderRevision != resolved.FounderLog.Revision ||
		resolved.FounderLog.Sequence < 1 || !founderAttendanceStreamPattern.MatchString(resolved.FounderLog.StreamID) {
		return LoggedTransition{}, ErrInvalidReplayInputs
	}
	definition, ok := catalogs.Minigames.Definition(resolved.MinigameID)
	if !ok || !reflect.DeepEqual(definition.Payout, resolved.PayoutPolicy) {
		return LoggedTransition{}, ErrInvalidReplayInputs
	}
	score, err := minigame.SelectPayoutScore(&result, definition.Payout)
	if err != nil || score != resolved.SelectedScore || validateFaucetReplay(resolved.Faucet, definition, score) != nil {
		return LoggedTransition{}, ErrInvalidReplayInputs
	}
	delta, err := decimal.ParseCanonical(resolved.CreditedDelta)
	if err != nil || delta.Lt(decimal.Zero) {
		return LoggedTransition{}, ErrInvalidReplayInputs
	}
	ledgerReceipt, err := state.Ledger.ApplyAccrual(economy.Transaction{Entries: []economy.Entry{{
		ResourceID: definition.Payout.CreditedResourceID, Delta: delta,
	}}})
	if err != nil || len(ledgerReceipt.Changes) != 1 || ledgerReceipt.Changes[0].Delta != resolved.CreditedDelta {
		return LoggedTransition{}, ErrInvalidReplayInputs
	}
	receipt := minigameResolutionReceipt{IntentID: payload.SessionID, Outcome: string(save.IntentApplied),
		SessionID: payload.SessionID, MinigameID: resolved.MinigameID, CertifiedResultHash: resolved.CertifiedResultHash,
		CompanyRevision: resolved.CompanyRevision, FounderRevision: resolved.FounderRevision,
		CreditedResourceID: definition.Payout.CreditedResourceID, CreditedDelta: resolved.CreditedDelta,
		ConfiguredForfeit: resolved.Faucet.ForfeitedUnits, CapReasonKey: resolved.Faucet.CapReasonKey,
		RatingChange: resolved.RatingChange, QualityChange: resolved.QualityChange}
	receiptBytes, _ := json.Marshal(receipt)
	eventPayload, _ := json.Marshal(map[string]any{"session_id": payload.SessionID, "minigame_id": resolved.MinigameID,
		"certified_result_hash": resolved.CertifiedResultHash, "credited_resource_id": definition.Payout.CreditedResourceID,
		"credited_delta": resolved.CreditedDelta, "configured_cap_forfeit_units": resolved.Faucet.ForfeitedUnits,
		"cap_reason_key": resolved.Faucet.CapReasonKey, "founder_revision": resolved.FounderRevision})
	return LoggedTransition{State: state, Outcome: save.IntentApplied, Receipt: receiptBytes,
		Events: []save.EventWrite{{Kind: save.EventMinigameResolved, SchemaVersion: 1, IntentID: payload.SessionID, Payload: eventPayload}}}, nil
}

func applyFounderMinigameResolution(state *save.State, requestPayload []byte, catalogs CatalogBundle,
	wire founderReplayInputsWire,
) (FounderLoggedTransition, error) {
	payload, err := parseMinigameResolutionPayload(requestPayload)
	if err != nil || payload.SessionID != wire.Command.IntentID || catalogs.Minigames == nil || state.MinigameRatings == nil || state.MinigameOfflineQuality == nil {
		return FounderLoggedTransition{}, ErrInvalidReplayInputs
	}
	var result minigame.Result
	var resolved minigameFounderResolved
	if decodeReplayStrict(payload.Result, &result) != nil || decodeReplayStrict(wire.Resolved, &resolved) != nil ||
		resolved.Kind != minigameResolutionKind || resolved.SessionID != payload.SessionID ||
		resolved.CertifiedResultHash != certifiedResultHash(payload.Result) || resolved.Attendance.CompanyConstantsHash != catalogs.ConstantsHash ||
		ValidateFounderAttendanceSample(state, wire.Command.Revision, wire.Command.Revision, resolved.Attendance) != nil {
		return FounderLoggedTransition{}, ErrInvalidReplayInputs
	}
	definition, ok := catalogs.Minigames.Definition(resolved.MinigameID)
	oldRating, ratingOK := state.MinigameRatings[resolved.MinigameID]
	oldQuality, qualityOK := state.MinigameOfflineQuality[resolved.MinigameID]
	if !ok || !ratingOK || !qualityOK || oldRating != resolved.RatingBefore || oldQuality != resolved.QualityBefore {
		return FounderLoggedTransition{}, ErrInvalidReplayInputs
	}
	transition, err := minigame.ApplyFounderResolution(minigame.RatingState{Elo: oldRating.Elo,
		SeasonMember: oldRating.SeasonMember, GamesCounted: oldRating.GamesCounted}, minigame.OfflineQualityState{
		GradePPM: oldQuality.GradePPM, LastFounderAttendedMS: oldQuality.LastFounderAttendedMS,
		DecayRemainderPPM: oldQuality.DecayRemainderPPM}, &result, definition, resolved.Attendance.EffectiveFounderAttendedMS)
	if err != nil {
		return FounderLoggedTransition{}, err
	}
	newRating := save.MinigameRatingState{Elo: transition.RatingAfter.Elo, SeasonMember: transition.RatingAfter.SeasonMember,
		GamesCounted: transition.RatingAfter.GamesCounted}
	newQuality := save.MinigameOfflineQualityState{GradePPM: transition.QualityAfter.GradePPM,
		LastFounderAttendedMS: transition.QualityAfter.LastFounderAttendedMS, DecayRemainderPPM: transition.QualityAfter.DecayRemainderPPM}
	if newRating != resolved.RatingAfter || newQuality != resolved.QualityAfter {
		return FounderLoggedTransition{}, ErrInvalidReplayInputs
	}
	state.MinigameRatings[resolved.MinigameID], state.MinigameOfflineQuality[resolved.MinigameID] = newRating, newQuality
	ratingChange := minigameRatingChangeReceipt{Rated: result.RatingDelta != nil, OldElo: oldRating.Elo,
		NewElo: newRating.Elo, SeasonMember: newRating.SeasonMember, GamesBefore: oldRating.GamesCounted, GamesAfter: newRating.GamesCounted}
	qualityChange := minigameQualityChangeReceipt{Old: oldQuality, New: newQuality}
	receipt, _ := json.Marshal(minigameFounderReceipt{IntentID: payload.SessionID, Outcome: string(save.IntentApplied),
		FounderRevision: wire.Command.Revision + 1, SessionID: payload.SessionID, CertifiedResultHash: resolved.CertifiedResultHash,
		RatingChange: ratingChange, QualityChange: qualityChange})
	eventPayload, _ := json.Marshal(map[string]any{"session_id": payload.SessionID, "minigame_id": resolved.MinigameID,
		"certified_result_hash": resolved.CertifiedResultHash, "old_elo": oldRating.Elo, "new_elo": newRating.Elo,
		"season_member": newRating.SeasonMember, "old_quality": oldQuality, "new_quality": newQuality})
	return FounderLoggedTransition{State: state, Outcome: save.IntentApplied, Receipt: receipt,
		Events:              []save.EventWrite{{Kind: save.EventMinigameRatingChanged, SchemaVersion: 1, IntentID: payload.SessionID, Payload: eventPayload}},
		ResultConstantsHash: catalogs.ConstantsHash}, nil
}

func validateFaucetReplay(value minigameFaucetWire, definition minigame.Definition, score int64) error {
	if value.AttendedDay < 0 || value.QuotaBefore < 0 || value.QuotaAfter < value.QuotaBefore ||
		value.QuotaAfter > value.QuotaBefore+1 || value.RemainderBeforePPM < 0 || value.RemainderBeforePPM >= 1_000_000 ||
		value.RemainderAfterPPM < 0 || value.RemainderAfterPPM >= 1_000_000 || value.CreditedUnits < 0 ||
		value.ForfeitedUnits < 0 || value.ConvertedUnits != value.CreditedUnits+value.ForfeitedUnits {
		return ErrInvalidReplayInputs
	}
	converted, err := minigame.ConvertPayout(score, definition.Fallback.RateReductionPPM,
		definition.Payout.ConversionPPM, value.RemainderBeforePPM)
	if err != nil || converted.ReducedScore != value.ReducedScore || converted.ConvertedUnits != value.ConvertedUnits ||
		converted.ConversionRemainderPPM != value.RemainderAfterPPM {
		return ErrInvalidReplayInputs
	}
	canCredit := value.QuotaBefore < definition.Payout.SendsPerDay
	expectedCredit := int64(0)
	if canCredit {
		expectedCredit = value.ConvertedUnits
		if expectedCredit > definition.Payout.PerSendCap {
			expectedCredit = definition.Payout.PerSendCap
		}
	}
	expectedQuota := value.QuotaBefore
	if canCredit {
		expectedQuota++
	}
	expectedForfeit := value.ConvertedUnits - expectedCredit
	expectedReason := ""
	if expectedForfeit > 0 {
		expectedReason = definition.Payout.CapReasonKey
	}
	if value.QuotaAfter != expectedQuota || value.CreditedUnits != expectedCredit || value.ForfeitedUnits != expectedForfeit || value.CapReasonKey != expectedReason {
		return ErrInvalidReplayInputs
	}
	return nil
}

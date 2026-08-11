package save

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"

	"cloud-clicker/server/economy"
)

var minigameOpaqueIDPattern = regexp.MustCompile(`^[A-Za-z0-9-]{1,64}$`)

// MinigameStartRequest is the server-authored MA-C11 coordinator envelope.
// RequestHash belongs to the canonical operation identity (route minigame ID
// plus public idempotency key), while canonical payload is the immutable
// Founder replay command.
type MinigameStartRequest struct {
	SessionID        string
	IntentID         string
	FounderID        string
	CompanyStreamID  string
	IdempotencyKey   string
	RequestHash      string
	CanonicalPayload json.RawMessage
}

type MinigameStartDecision struct {
	Receipt               json.RawMessage
	FounderReceipt        json.RawMessage
	FounderReplayResolved json.RawMessage
	FounderEvents         []EventWrite
}

type MinigameStartMutation func(ctx context.Context, tx *sql.Tx, founder *State,
	founderRevision Revision, company *State, companyRevision Revision,
	founderCommand FounderReplayCommand, founderNextRevision int64,
) (MinigameStartDecision, error)

// ApplyMinigameStartTransaction owns the Founder→Company→session lock and
// commit order. It advances only the replay-owned Founder sequence; Company is
// locked and validated as the pinned run authority but must remain byte-equal.
func (s *Store) ApplyMinigameStartTransaction(ctx context.Context, request MinigameStartRequest,
	mutate MinigameStartMutation, fault ExitFaultInjector,
) (IntentResult, error) {
	if s == nil || !uuidV7Pattern.MatchString(request.SessionID) || !uuidV7Pattern.MatchString(request.IntentID) || request.IntentID == request.SessionID || !uuidPattern.MatchString(request.FounderID) ||
		!uuidPattern.MatchString(request.CompanyStreamID) || !minigameOpaqueIDPattern.MatchString(request.IdempotencyKey) ||
		!hashPattern.MatchString(request.RequestHash) || !jsonObject(request.CanonicalPayload) || mutate == nil {
		return IntentResult{}, fmt.Errorf("%w: invalid minigame start request", ErrInvalidStream)
	}
	var founderStreamID string
	if err := s.db.QueryRowContext(ctx, `SELECT id FROM save_streams WHERE owner_kind='founder' AND owner_id=$1 AND scope='founder' AND archived_at IS NULL`, request.FounderID).Scan(&founderStreamID); errors.Is(err, sql.ErrNoRows) {
		return IntentResult{}, ErrNotFound
	} else if err != nil {
		return IntentResult{}, err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return IntentResult{}, err
	}
	defer tx.Rollback()
	var activeFounder bool
	if err := tx.QueryRowContext(ctx, `SELECT true FROM account_founders WHERE founder_id=$1 AND archived_at IS NULL FOR UPDATE`, request.FounderID).Scan(&activeFounder); errors.Is(err, sql.ErrNoRows) {
		return IntentResult{}, ErrNotFound
	} else if err != nil {
		return IntentResult{}, err
	}
	for _, streamID := range []string{founderStreamID, request.CompanyStreamID} {
		var archived sql.NullTime
		var ownerID string
		if err := tx.QueryRowContext(ctx, `SELECT archived_at,owner_id FROM save_streams WHERE id=$1 FOR UPDATE`, streamID).Scan(&archived, &ownerID); err != nil {
			return IntentResult{}, err
		}
		if archived.Valid || ownerID != request.FounderID {
			return IntentResult{}, ErrArchived
		}
	}

	var recordedHash string
	var recordedResponse []byte
	err = tx.QueryRowContext(ctx, `SELECT request_hash,response::text FROM minigame_create_receipts WHERE founder_id=$1 AND idempotency_key=$2`,
		request.FounderID, request.IdempotencyKey).Scan(&recordedHash, &recordedResponse)
	if err == nil {
		if recordedHash != request.RequestHash {
			return IntentResult{}, ErrIdempotencyConflict
		}
		recordedResponse, err = normalizeJSON(recordedResponse)
		if err != nil || !jsonObject(recordedResponse) {
			return IntentResult{}, ErrInvalidState
		}
		if err := tx.Commit(); err != nil {
			return IntentResult{}, err
		}
		return IntentResult{Outcome: IntentApplied, Receipt: cloneRaw(recordedResponse), Replay: true}, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return IntentResult{}, err
	}

	founder, founderRevision, founderHash, founderStateBytes, err := s.loadExitState(ctx, tx, founderStreamID, request.FounderID, economy.ScopeFounder)
	if err != nil {
		return IntentResult{}, err
	}
	company, companyRevision, companyHash, _, err := s.loadExitState(ctx, tx, request.CompanyStreamID, request.FounderID, economy.ScopeCompany)
	if err != nil {
		return IntentResult{}, err
	}
	if founderHash != companyHash || founderRevision.ConstantsHash != companyRevision.ConstantsHash ||
		company.RunSeq < 1 || VersionForState(founder) < 21 {
		return IntentResult{}, fmt.Errorf("%w: minigame start streams do not share an active v21 run", ErrInvalidState)
	}
	if err := requireRunEpochTx(ctx, tx, request.CompanyStreamID, company.RunSeq, companyHash); err != nil {
		return IntentResult{}, err
	}
	companyBefore, err := s.validatedState(companyHash, economy.ScopeCompany, company)
	if err != nil {
		return IntentResult{}, err
	}
	founderLogSequence, err := nextFounderLogSequence(ctx, tx, founderStreamID)
	if err != nil {
		return IntentResult{}, err
	}
	serverTSMS, err := founderServerTimestamp(ctx, tx)
	if err != nil {
		return IntentResult{}, err
	}
	command := FounderReplayCommand{IntentID: request.IntentID, FounderStreamID: founderStreamID,
		FounderID: request.FounderID, Revision: founderRevision.Number, FounderLogSeq: founderLogSequence,
		ServerTSMS: serverTSMS}
	if founderLogSequence == 1 {
		if err := InsertFounderGenesisTx(ctx, tx, FounderGenesis{FounderStreamID: founderStreamID,
			Revision: founderRevision.Number, State: founderStateBytes, Version: founderRevision.Version,
			ConstantsHash: founderHash}); err != nil {
			return IntentResult{}, err
		}
		if err := runExitFault(fault, "founder_genesis"); err != nil {
			return IntentResult{}, err
		}
	}
	founderNext := founderRevision.Number + 1
	decision, err := mutate(ctx, tx, founder, founderRevision, company, companyRevision, command, founderNext)
	if err != nil {
		return IntentResult{}, err
	}
	decision.Receipt, err = normalizeJSON(decision.Receipt)
	if err != nil || !jsonObject(decision.Receipt) {
		return IntentResult{}, fmt.Errorf("%w: invalid minigame start receipt", ErrInvalidStream)
	}
	decision.FounderReceipt, err = normalizeJSON(decision.FounderReceipt)
	if err != nil || !jsonObject(decision.FounderReceipt) {
		return IntentResult{}, fmt.Errorf("%w: invalid Founder minigame start receipt", ErrInvalidStream)
	}
	var committedHash, committedSession string
	var committedResponse []byte
	if err := tx.QueryRowContext(ctx, `SELECT request_hash,session_id,response::text FROM minigame_create_receipts WHERE founder_id=$1 AND idempotency_key=$2`,
		request.FounderID, request.IdempotencyKey).Scan(&committedHash, &committedSession, &committedResponse); err != nil {
		return IntentResult{}, err
	}
	committedResponse, err = normalizeJSON(committedResponse)
	if err != nil || committedHash != request.RequestHash || committedSession != request.SessionID || !bytes.Equal(committedResponse, decision.Receipt) {
		return IntentResult{}, fmt.Errorf("%w: minigame start receipt was not committed atomically", ErrInvalidState)
	}
	founderEnvelope, err := MarshalFounderReplayInputs(command, decision.FounderReplayResolved)
	if err != nil {
		return IntentResult{}, err
	}
	decision.FounderReplayResolved, err = ValidateFounderReplayInputs(founderEnvelope, command)
	if err != nil {
		return IntentResult{}, err
	}
	if err := validateIntentDecision(IntentDecision{Outcome: IntentApplied, Receipt: decision.FounderReceipt, Events: decision.FounderEvents}, request.IntentID); err != nil {
		return IntentResult{}, err
	}
	if VersionForState(founder) != migratedWriteVersion(founderRevision.Version) ||
		VersionForState(company) != migratedWriteVersion(companyRevision.Version) {
		return IntentResult{}, fmt.Errorf("%w: minigame start changed save identity", ErrInvalidState)
	}
	companyAfter, err := s.validatedState(companyHash, economy.ScopeCompany, company)
	if err != nil || !bytes.Equal(companyBefore, companyAfter) {
		return IntentResult{}, fmt.Errorf("%w: minigame start mutated Company state", ErrInvalidState)
	}
	founderEncoded, err := s.validatedState(founderHash, economy.ScopeFounder, founder)
	if err != nil {
		return IntentResult{}, err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO save_revisions(stream_id,revision,version,state,constants_hash) VALUES($1,$2,$3,$4,$5)`,
		founderStreamID, founderNext, VersionForState(founder), founderEncoded, founderHash); err != nil {
		return IntentResult{}, err
	}
	if err := runExitFault(fault, "founder_revision"); err != nil {
		return IntentResult{}, err
	}
	recordedEvents, err := insertExitEvents(ctx, tx, founderStreamID, request.FounderID, founderNext, founderHash, decision.FounderEvents)
	if err != nil {
		return IntentResult{}, err
	}
	if err := runExitFault(fault, "founder_events"); err != nil {
		return IntentResult{}, err
	}
	if err := insertFounderLog(ctx, tx, command, founderHash, request.CanonicalPayload,
		decision.FounderReplayResolved, decision.FounderReceipt, &founderNext, nil); err != nil {
		return IntentResult{}, err
	}
	if err := runExitFault(fault, "founder_log"); err != nil {
		return IntentResult{}, err
	}
	if err := pruneSaveRevisionsTx(ctx, tx, founderStreamID, founderNext-5); err != nil {
		return IntentResult{}, err
	}
	if err := runExitFault(fault, "retention"); err != nil {
		return IntentResult{}, err
	}
	if err := tx.Commit(); err != nil {
		return IntentResult{}, err
	}
	return IntentResult{Outcome: IntentApplied, Receipt: cloneRaw(decision.Receipt), Events: recordedEvents}, nil
}

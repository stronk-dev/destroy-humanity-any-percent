package save

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"cloud-clicker/server/economy"
)

// MinigameResolutionDecision is the persistence-neutral output of the one
// live/replay transition pair. The callback may finalize the claimed session
// and faucet window through its owning package using the supplied transaction;
// Store owns every save/log/event/outbox write around it.
type MinigameResolutionDecision struct {
	Receipt json.RawMessage
	// CompanyLogReceipt is the receipt reproduced by the Company replay
	// boundary. When absent, Receipt remains the shared API/log receipt used by
	// ordinary minigame resolutions. Cross-stream coordinators may expose a
	// richer durable API receipt without corrupting Company replay parity.
	CompanyLogReceipt     json.RawMessage
	FounderReceipt        json.RawMessage
	CompanyReplayInputs   json.RawMessage
	FounderReplayResolved json.RawMessage
	CompanyEvents         []EventWrite
	FounderEvents         []EventWrite
}

type MinigameResolutionMutation func(ctx context.Context, tx *sql.Tx, founder *State,
	founderRevision Revision, company *State, companyRevision Revision, companyCommand ReplayCommand,
	founderCommand FounderReplayCommand, companyNextRevision, founderNextRevision int64,
) (MinigameResolutionDecision, error)

type MinigameResolutionRequest struct {
	SessionID        string
	FounderID        string
	CompanyStreamID  string
	RequestHash      string
	CanonicalPayload json.RawMessage
	// ServerTSMS is resolved by the composed server command before the
	// transaction and becomes the Founder replay timestamp. It is never client
	// supplied.
	ServerTSMS int64
	// CompanyLogFirst is reserved for server-authored suppression commands
	// whose audit contract requires the Company replay row to precede the
	// Founder audit row. Minigame resolution retains its established order.
	CompanyLogFirst bool
}

// ApplyMinigameResolutionTransaction is the C38 Founder→Company coordinator.
// SessionID is both idempotency key and internal intent ID. A retry returns the
// committed intent receipt before invoking mutate, so tenant replay and faucet
// accounting run at most once.
func (s *Store) ApplyMinigameResolutionTransaction(ctx context.Context, request MinigameResolutionRequest,
	mutate MinigameResolutionMutation, fault ExitFaultInjector,
) (IntentResult, error) {
	if s == nil || !uuidV7Pattern.MatchString(request.SessionID) || !uuidPattern.MatchString(request.FounderID) ||
		!uuidPattern.MatchString(request.CompanyStreamID) || !hashPattern.MatchString(request.RequestHash) ||
		request.ServerTSMS <= 0 || request.ServerTSMS > 9007199254740991 || mutate == nil ||
		validateCanonicalPayload(request.CanonicalPayload, request.RequestHash) != nil {
		return IntentResult{}, fmt.Errorf("%w: invalid minigame resolution request", ErrInvalidStream)
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
	var recordedOutcome IntentOutcome
	var recordedReceipt []byte
	err = tx.QueryRowContext(ctx, `SELECT request_hash,outcome,receipt FROM intent_records WHERE stream_id=$1 AND intent_id=$2`,
		request.CompanyStreamID, request.SessionID).Scan(&recordedHash, &recordedOutcome, &recordedReceipt)
	if err == nil {
		if recordedHash != request.RequestHash {
			return IntentResult{}, ErrIdempotencyConflict
		}
		recordedReceipt, err = normalizeJSON(recordedReceipt)
		if err != nil {
			return IntentResult{}, err
		}
		events, eventsErr := eventsForExitIntent(ctx, tx, []string{founderStreamID, request.CompanyStreamID}, request.FounderID, request.SessionID)
		if eventsErr != nil {
			return IntentResult{}, eventsErr
		}
		if err := tx.Commit(); err != nil {
			return IntentResult{}, err
		}
		return IntentResult{Outcome: recordedOutcome, Receipt: cloneRaw(recordedReceipt), Replay: true, Events: events}, nil
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
	if founderHash != companyHash || founderRevision.ConstantsHash != companyRevision.ConstantsHash || company.RunSeq < 1 {
		return IntentResult{}, fmt.Errorf("%w: minigame streams do not share a pinned run", ErrInvalidState)
	}
	if err := requireRunEpochTx(ctx, tx, request.CompanyStreamID, company.RunSeq, companyHash); err != nil {
		return IntentResult{}, err
	}
	runLogSequence, err := nextRunLogSequence(ctx, tx, request.CompanyStreamID, company.RunSeq)
	if err != nil {
		return IntentResult{}, err
	}
	companyRevision.RunLogSequence = runLogSequence
	companyCommand := ReplayCommand{IntentID: request.SessionID, CompanyStreamID: request.CompanyStreamID,
		FounderID: request.FounderID, Revision: companyRevision.Number, RunSeq: company.RunSeq, RunLogSeq: runLogSequence}
	founderLogSequence, err := nextFounderLogSequence(ctx, tx, founderStreamID)
	if err != nil {
		return IntentResult{}, err
	}
	founderCommand := FounderReplayCommand{IntentID: request.SessionID, FounderStreamID: founderStreamID,
		FounderID: request.FounderID, Revision: founderRevision.Number, FounderLogSeq: founderLogSequence, ServerTSMS: request.ServerTSMS}
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
	founderNext, companyNext := founderRevision.Number+1, companyRevision.Number+1
	decision, err := mutate(ctx, tx, founder, founderRevision, company, companyRevision, companyCommand,
		founderCommand, companyNext, founderNext)
	if err != nil {
		return IntentResult{}, err
	}
	decision.Receipt, err = normalizeJSON(decision.Receipt)
	if err != nil || !jsonObject(decision.Receipt) {
		return IntentResult{}, fmt.Errorf("%w: invalid minigame resolution receipt", ErrInvalidStream)
	}
	companyLogReceipt := decision.Receipt
	if len(decision.CompanyLogReceipt) != 0 {
		companyLogReceipt, err = normalizeJSON(decision.CompanyLogReceipt)
		if err != nil || !jsonObject(companyLogReceipt) {
			return IntentResult{}, fmt.Errorf("%w: invalid Company replay receipt", ErrInvalidStream)
		}
	}
	decision.FounderReceipt, err = normalizeJSON(decision.FounderReceipt)
	if err != nil || !jsonObject(decision.FounderReceipt) {
		return IntentResult{}, fmt.Errorf("%w: invalid Founder minigame receipt", ErrInvalidStream)
	}
	decision.CompanyReplayInputs, err = ValidateReplayInputs(decision.CompanyReplayInputs, companyCommand)
	if err != nil {
		return IntentResult{}, err
	}
	founderEnvelope, err := MarshalFounderReplayInputs(founderCommand, decision.FounderReplayResolved)
	if err != nil {
		return IntentResult{}, err
	}
	decision.FounderReplayResolved, err = ValidateFounderReplayInputs(founderEnvelope, founderCommand)
	if err != nil {
		return IntentResult{}, err
	}
	for _, events := range [][]EventWrite{decision.FounderEvents, decision.CompanyEvents} {
		if err := validateIntentDecision(IntentDecision{Outcome: IntentApplied, Receipt: decision.Receipt, Events: events}, request.SessionID); err != nil {
			return IntentResult{}, err
		}
	}
	if VersionForState(founder) != migratedWriteVersion(founderRevision.Version) ||
		VersionForState(company) != migratedWriteVersion(companyRevision.Version) || company.RunSeq != companyCommand.RunSeq {
		return IntentResult{}, fmt.Errorf("%w: minigame resolution changed save identity", ErrInvalidState)
	}
	founderEncoded, err := s.validatedState(founderHash, economy.ScopeFounder, founder)
	if err != nil {
		return IntentResult{}, err
	}
	companyEncoded, err := s.validatedState(companyHash, economy.ScopeCompany, company)
	if err != nil {
		return IntentResult{}, err
	}
	recorded := make([]EventRecord, 0, len(decision.FounderEvents)+len(decision.CompanyEvents))
	if request.CompanyLogFirst {
		if _, err := tx.ExecContext(ctx, `INSERT INTO save_revisions(stream_id,revision,version,state,constants_hash) VALUES($1,$2,$3,$4,$5)`,
			request.CompanyStreamID, companyNext, VersionForState(company), companyEncoded, companyHash); err != nil {
			return IntentResult{}, err
		}
		if err := runExitFault(fault, "company_revision"); err != nil {
			return IntentResult{}, err
		}
		companyEvents, err := insertExitEvents(ctx, tx, request.CompanyStreamID, request.FounderID, companyNext, companyHash, decision.CompanyEvents)
		if err != nil {
			return IntentResult{}, err
		}
		recorded = append(recorded, companyEvents...)
		if err := runExitFault(fault, "company_events"); err != nil {
			return IntentResult{}, err
		}
		if err := insertRunLog(ctx, tx, request.CompanyStreamID, company.RunSeq, runLogSequence, request.SessionID,
			request.CanonicalPayload, decision.CompanyReplayInputs, companyLogReceipt, &companyNext); err != nil {
			return IntentResult{}, err
		}
		if err := runExitFault(fault, "run_log"); err != nil {
			return IntentResult{}, err
		}
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO save_revisions(stream_id,revision,version,state,constants_hash) VALUES($1,$2,$3,$4,$5)`,
		founderStreamID, founderNext, VersionForState(founder), founderEncoded, founderHash); err != nil {
		return IntentResult{}, err
	}
	if err := runExitFault(fault, "founder_revision"); err != nil {
		return IntentResult{}, err
	}
	founderEvents, err := insertExitEvents(ctx, tx, founderStreamID, request.FounderID, founderNext, founderHash, decision.FounderEvents)
	if err != nil {
		return IntentResult{}, err
	}
	recorded = append(recorded, founderEvents...)
	if err := runExitFault(fault, "founder_events"); err != nil {
		return IntentResult{}, err
	}
	if !request.CompanyLogFirst {
		if _, err := tx.ExecContext(ctx, `INSERT INTO save_revisions(stream_id,revision,version,state,constants_hash) VALUES($1,$2,$3,$4,$5)`,
			request.CompanyStreamID, companyNext, VersionForState(company), companyEncoded, companyHash); err != nil {
			return IntentResult{}, err
		}
		if err := runExitFault(fault, "company_revision"); err != nil {
			return IntentResult{}, err
		}
		companyEvents, err := insertExitEvents(ctx, tx, request.CompanyStreamID, request.FounderID, companyNext, companyHash, decision.CompanyEvents)
		if err != nil {
			return IntentResult{}, err
		}
		recorded = append(recorded, companyEvents...)
		if err := runExitFault(fault, "company_events"); err != nil {
			return IntentResult{}, err
		}
		if err := insertRunLog(ctx, tx, request.CompanyStreamID, company.RunSeq, runLogSequence, request.SessionID,
			request.CanonicalPayload, decision.CompanyReplayInputs, companyLogReceipt, &companyNext); err != nil {
			return IntentResult{}, err
		}
		if err := runExitFault(fault, "run_log"); err != nil {
			return IntentResult{}, err
		}
	}
	source := &FounderLogSource{CompanyStreamID: request.CompanyStreamID, RunSeq: company.RunSeq, RunLogSeq: runLogSequence}
	if err := insertFounderLog(ctx, tx, founderCommand, founderHash, request.CanonicalPayload,
		decision.FounderReplayResolved, decision.FounderReceipt, &founderNext, source); err != nil {
		return IntentResult{}, err
	}
	if err := runExitFault(fault, "founder_log"); err != nil {
		return IntentResult{}, err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO intent_records(stream_id,intent_id,request_hash,outcome,receipt) VALUES($1,$2,$3,$4,$5)`,
		request.CompanyStreamID, request.SessionID, request.RequestHash, IntentApplied, decision.Receipt); err != nil {
		return IntentResult{}, err
	}
	if err := insertReceiptOutbox(ctx, tx, request.FounderID, request.CompanyStreamID, request.SessionID,
		economy.ScopeCompany, companyNext, companyHash, decision.Receipt); err != nil {
		return IntentResult{}, err
	}
	if err := runExitFault(fault, "intent_record"); err != nil {
		return IntentResult{}, err
	}
	if err := pruneSaveRevisionsTx(ctx, tx, founderStreamID, founderNext-5); err != nil {
		return IntentResult{}, err
	}
	if err := pruneSaveRevisionsTx(ctx, tx, request.CompanyStreamID, companyNext-5); err != nil {
		return IntentResult{}, err
	}
	if err := runExitFault(fault, "retention"); err != nil {
		return IntentResult{}, err
	}
	if err := tx.Commit(); err != nil {
		return IntentResult{}, err
	}
	return IntentResult{Outcome: IntentApplied, Receipt: cloneRaw(decision.Receipt), Events: recorded}, nil
}

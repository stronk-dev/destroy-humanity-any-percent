package save

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"time"

	"cloud-clicker/server/decimal"
	"cloud-clicker/server/economy"
)

var ErrIdempotencyConflict = errors.New("idempotency key reused with a different request")

var (
	uuidV7Pattern       = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-7[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
	mechanicalIDPattern = regexp.MustCompile(`^[a-z][a-z0-9_]*(?:\.[a-z][a-z0-9_]*)*$`)
)

type IntentOutcome string

const (
	IntentApplied  IntentOutcome = "applied"
	IntentRejected IntentOutcome = "rejected"
)

type EventKind string

const (
	EventGeneratorPurchased EventKind = "generator_purchased"
	EventInvariantReported  EventKind = "invariant_reported"
	EventCompensation       EventKind = "compensation"
)

type EventWrite struct {
	Kind          EventKind
	SchemaVersion int
	IntentID      string
	Payload       json.RawMessage
}

type IntentDecision struct {
	Outcome IntentOutcome
	Receipt json.RawMessage
	Events  []EventWrite
}

type IntentMutation func(state *State, revision Revision) (IntentDecision, error)

type IntentResult struct {
	Outcome IntentOutcome
	Receipt json.RawMessage
	Replay  bool
}

type RevisionConflict struct {
	Expected int64
	Current  int64
}

func (e *RevisionConflict) Error() string {
	return fmt.Sprintf("%v: got %d, current %d", ErrConflict, e.Expected, e.Current)
}

func (e *RevisionConflict) Unwrap() error { return ErrConflict }

func (s *Store) ApplyIntent(
	ctx context.Context,
	streamID string,
	expectedRevision int64,
	intentID string,
	requestHash string,
	mutate IntentMutation,
) (IntentResult, error) {
	if !uuidPattern.MatchString(streamID) || expectedRevision < 1 || !uuidV7Pattern.MatchString(intentID) ||
		!hashPattern.MatchString(requestHash) || mutate == nil {
		return IntentResult{}, ErrInvalidStream
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return IntentResult{}, err
	}
	defer tx.Rollback()

	var scope economy.Scope
	var archivedAt sql.NullTime
	if err := tx.QueryRowContext(ctx, `SELECT scope, archived_at FROM save_streams WHERE id=$1 FOR UPDATE`, streamID).Scan(&scope, &archivedAt); errors.Is(err, sql.ErrNoRows) {
		return IntentResult{}, ErrNotFound
	} else if err != nil {
		return IntentResult{}, err
	}
	if archivedAt.Valid {
		return IntentResult{}, ErrArchived
	}

	var recordedHash string
	var recordedOutcome IntentOutcome
	var recordedReceipt []byte
	err = tx.QueryRowContext(ctx, `
		SELECT request_hash,outcome,receipt FROM intent_records
		WHERE stream_id=$1 AND intent_id=$2`, streamID, intentID,
	).Scan(&recordedHash, &recordedOutcome, &recordedReceipt)
	if err == nil {
		if recordedHash != requestHash {
			return IntentResult{}, ErrIdempotencyConflict
		}
		recordedReceipt, err = normalizeJSON(recordedReceipt)
		if err != nil {
			return IntentResult{}, fmt.Errorf("%w: stored intent receipt: %v", ErrInvalidState, err)
		}
		if err := tx.Commit(); err != nil {
			return IntentResult{}, err
		}
		return IntentResult{Outcome: recordedOutcome, Receipt: cloneRaw(recordedReceipt), Replay: true}, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return IntentResult{}, err
	}

	var revision Revision
	var stateBytes []byte
	err = tx.QueryRowContext(ctx, `
		SELECT revision,version,constants_hash,created_at,state
		FROM save_revisions WHERE stream_id=$1 ORDER BY revision DESC LIMIT 1`, streamID,
	).Scan(&revision.Number, &revision.Version, &revision.ConstantsHash, &revision.CreatedAt, &stateBytes)
	if errors.Is(err, sql.ErrNoRows) {
		return IntentResult{}, ErrNotFound
	}
	if err != nil {
		return IntentResult{}, err
	}
	revision.StreamID = streamID
	if revision.Number != expectedRevision {
		return IntentResult{}, &RevisionConflict{Expected: expectedRevision, Current: revision.Number}
	}
	catalog, ok := s.catalogs.Resolve(revision.ConstantsHash)
	if !ok {
		return IntentResult{}, fmt.Errorf("%w: unknown catalog %s", ErrInvalidState, revision.ConstantsHash)
	}
	state, err := RestoreState(stateBytes, revision.Version, catalog, scope, revision.CreatedAt)
	if err != nil {
		return IntentResult{}, err
	}
	decision, err := mutate(state, revision)
	if err != nil {
		return IntentResult{}, err
	}
	if err := validateIntentDecision(decision, intentID); err != nil {
		return IntentResult{}, err
	}
	decision.Receipt, err = normalizeJSON(decision.Receipt)
	if err != nil {
		return IntentResult{}, fmt.Errorf("%w: normalize intent receipt: %v", ErrInvalidStream, err)
	}

	if decision.Outcome == IntentRejected {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO intent_records (stream_id,intent_id,request_hash,outcome,receipt)
			VALUES ($1,$2,$3,$4,$5)`, streamID, intentID, requestHash, decision.Outcome, decision.Receipt); err != nil {
			return IntentResult{}, err
		}
		if err := tx.Commit(); err != nil {
			return IntentResult{}, err
		}
		return IntentResult{Outcome: decision.Outcome, Receipt: cloneRaw(decision.Receipt)}, nil
	}

	encodedState, err := s.validatedState(revision.ConstantsHash, scope, state)
	if err != nil {
		return IntentResult{}, err
	}
	newRevision := revision.Number + 1
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO save_revisions (stream_id,revision,version,state,constants_hash)
		VALUES ($1,$2,$3,$4,$5)`, streamID, newRevision, CurrentVersion, encodedState, revision.ConstantsHash); err != nil {
		return IntentResult{}, err
	}
	for _, event := range decision.Events {
		var eventIntent any
		if event.IntentID != "" {
			eventIntent = event.IntentID
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO events (stream_id,revision,schema_version,kind,intent_id,constants_hash,payload)
			VALUES ($1,$2,$3,$4,$5,$6,$7)`, streamID, newRevision, event.SchemaVersion,
			event.Kind, eventIntent, revision.ConstantsHash, event.Payload); err != nil {
			return IntentResult{}, err
		}
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO intent_records (stream_id,intent_id,request_hash,outcome,receipt)
		VALUES ($1,$2,$3,$4,$5)`, streamID, intentID, requestHash, decision.Outcome, decision.Receipt); err != nil {
		return IntentResult{}, err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM save_revisions WHERE stream_id=$1 AND revision <= $2`, streamID, newRevision-5); err != nil {
		return IntentResult{}, err
	}
	if err := tx.Commit(); err != nil {
		return IntentResult{}, err
	}
	return IntentResult{Outcome: decision.Outcome, Receipt: cloneRaw(decision.Receipt)}, nil
}

func (s *Store) PruneIntentRecords(ctx context.Context, cutoff time.Time) (int64, error) {
	if cutoff.IsZero() {
		return 0, ErrInvalidStream
	}
	result, err := s.db.ExecContext(ctx, `DELETE FROM intent_records WHERE created_at < $1`, cutoff.UTC())
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func validateIntentDecision(decision IntentDecision, intentID string) error {
	if decision.Outcome != IntentApplied && decision.Outcome != IntentRejected {
		return ErrInvalidStream
	}
	var receipt map[string]json.RawMessage
	if err := decodeStrictJSON(decision.Receipt, &receipt); err != nil || receipt == nil {
		return fmt.Errorf("%w: invalid intent receipt", ErrInvalidStream)
	}
	if decision.Outcome == IntentRejected && len(decision.Events) != 0 {
		return fmt.Errorf("%w: rejected intent cannot emit events", ErrInvalidStream)
	}
	for _, event := range decision.Events {
		if event.SchemaVersion != 1 || event.IntentID != "" && !uuidV7Pattern.MatchString(event.IntentID) {
			return fmt.Errorf("%w: invalid event envelope", ErrInvalidStream)
		}
		if event.Kind != EventCompensation && event.IntentID != intentID {
			return fmt.Errorf("%w: event intent does not match mutation", ErrInvalidStream)
		}
		if err := validateEventPayload(event); err != nil {
			return err
		}
	}
	return nil
}

func validateEventPayload(event EventWrite) error {
	switch event.Kind {
	case EventGeneratorPurchased:
		var payload struct {
			GeneratorID    string `json:"generator_id"`
			Count          int64  `json:"count"`
			CostResourceID string `json:"cost_resource_id"`
			Cost           string `json:"cost"`
		}
		if err := decodeStrictJSON(event.Payload, &payload); err != nil || !mechanicalIDPattern.MatchString(payload.GeneratorID) ||
			!mechanicalIDPattern.MatchString(payload.CostResourceID) || payload.Count <= 0 || payload.Count > decimal.MaxExactInteger {
			return fmt.Errorf("%w: invalid generator_purchased payload", ErrInvalidStream)
		}
		if value, err := decimal.ParseCanonical(payload.Cost); err != nil || value.Lt(decimal.Zero) {
			return fmt.Errorf("%w: invalid generator_purchased cost", ErrInvalidStream)
		}
	case EventInvariantReported:
		var payload struct {
			InvariantKind string `json:"invariant_kind"`
			Detail        string `json:"detail"`
		}
		if err := decodeStrictJSON(event.Payload, &payload); err != nil || payload.Detail == "" ||
			(payload.InvariantKind != "afford_fallback" && payload.InvariantKind != "residual_clamp" && payload.InvariantKind != "residual_abort") {
			return fmt.Errorf("%w: invalid invariant_reported payload", ErrInvalidStream)
		}
	case EventCompensation:
		var payload struct {
			CompensatesEventID string `json:"compensates_event_id"`
			ReasonKey          string `json:"reason_key"`
		}
		if err := decodeStrictJSON(event.Payload, &payload); err != nil || !uuidPattern.MatchString(payload.CompensatesEventID) || !mechanicalIDPattern.MatchString(payload.ReasonKey) {
			return fmt.Errorf("%w: invalid compensation payload", ErrInvalidStream)
		}
	default:
		return fmt.Errorf("%w: unknown event kind %q", ErrInvalidStream, event.Kind)
	}
	return nil
}

func decodeStrictJSON(data []byte, destination any) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return err
	}
	return ensureJSONEnd(decoder)
}

func cloneRaw(source []byte) json.RawMessage {
	return append(json.RawMessage(nil), source...)
}

func normalizeJSON(source []byte) (json.RawMessage, error) {
	decoder := json.NewDecoder(bytes.NewReader(source))
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err != nil {
		return nil, err
	}
	if err := ensureJSONEnd(decoder); err != nil {
		return nil, err
	}
	return json.Marshal(value)
}

package save

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"sort"

	"cloud-clicker/server/economy"
)

type ExitDecision struct {
	Outcome              IntentOutcome
	Receipt              json.RawMessage
	FinalCompanyState    *State
	NewCompanyState      *State
	FounderEvents        []EventWrite
	CompanyEndedEvents   []EventWrite
	CompanyStartedEvents []EventWrite
}

type ExitMutation func(founder *State, founderRevision Revision, company *State, companyRevision Revision) (ExitDecision, error)
type ExitFaultInjector func(step string) error

type ExitRevisionConflict struct {
	Stream   economy.Scope
	Expected int64
	Current  int64
}

func (conflict *ExitRevisionConflict) Error() string {
	return fmt.Sprintf("%v: %s got %d, current %d", ErrConflict, conflict.Stream, conflict.Expected, conflict.Current)
}

func (conflict *ExitRevisionConflict) Unwrap() error { return ErrConflict }

func (s *Store) ApplyExitTransaction(
	ctx context.Context,
	companyStreamID string,
	expectedCompanyRevision, expectedFounderRevision int64,
	intentID, requestHash string,
	mutate ExitMutation,
	fault ExitFaultInjector,
) (IntentResult, error) {
	if !uuidPattern.MatchString(companyStreamID) || expectedCompanyRevision < 1 || expectedFounderRevision < 1 ||
		!uuidV7Pattern.MatchString(intentID) || !hashPattern.MatchString(requestHash) || mutate == nil {
		return IntentResult{}, ErrInvalidStream
	}
	var founderStreamID string
	var ownerID string
	if err := s.db.QueryRowContext(ctx, `
		SELECT founder.id,company.owner_id
		FROM save_streams company
		JOIN save_streams founder ON founder.owner_kind='founder' AND founder.owner_id=company.owner_id AND founder.scope='founder' AND founder.archived_at IS NULL
		WHERE company.id=$1 AND company.owner_kind='founder' AND company.scope='company'`, companyStreamID).Scan(&founderStreamID, &ownerID); errors.Is(err, sql.ErrNoRows) {
		return IntentResult{}, ErrNotFound
	} else if err != nil {
		return IntentResult{}, err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return IntentResult{}, err
	}
	defer tx.Rollback()

	// The normative lock order is Founder, then Company. There is exactly one of
	// each scope; IDs only break ties in future extensions with multiple streams.
	for _, streamID := range []string{founderStreamID, companyStreamID} {
		var archived sql.NullTime
		if err := tx.QueryRowContext(ctx, `SELECT archived_at FROM save_streams WHERE id=$1 FOR UPDATE`, streamID).Scan(&archived); err != nil {
			return IntentResult{}, err
		}
		if archived.Valid {
			return IntentResult{}, ErrArchived
		}
	}

	var recordedHash string
	var recordedOutcome IntentOutcome
	var recordedReceipt []byte
	err = tx.QueryRowContext(ctx, `SELECT request_hash,outcome,receipt FROM intent_records WHERE stream_id=$1 AND intent_id=$2`, companyStreamID, intentID).Scan(&recordedHash, &recordedOutcome, &recordedReceipt)
	if err == nil {
		if recordedHash != requestHash {
			return IntentResult{}, ErrIdempotencyConflict
		}
		recordedReceipt, err = normalizeJSON(recordedReceipt)
		if err != nil {
			return IntentResult{}, err
		}
		events, err := eventsForExitIntent(ctx, tx, []string{founderStreamID, companyStreamID}, ownerID, intentID)
		if err != nil {
			return IntentResult{}, err
		}
		if err := tx.Commit(); err != nil {
			return IntentResult{}, err
		}
		return IntentResult{Outcome: recordedOutcome, Receipt: cloneRaw(recordedReceipt), Replay: true, Events: events}, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return IntentResult{}, err
	}

	founder, founderRevision, founderHash, err := s.loadExitState(ctx, tx, founderStreamID, ownerID, economy.ScopeFounder)
	if err != nil {
		return IntentResult{}, err
	}
	company, companyRevision, companyHash, err := s.loadExitState(ctx, tx, companyStreamID, ownerID, economy.ScopeCompany)
	if err != nil {
		return IntentResult{}, err
	}
	if founderRevision.Number != expectedFounderRevision {
		return IntentResult{}, &ExitRevisionConflict{Stream: economy.ScopeFounder, Expected: expectedFounderRevision, Current: founderRevision.Number}
	}
	if companyRevision.Number != expectedCompanyRevision {
		return IntentResult{}, &ExitRevisionConflict{Stream: economy.ScopeCompany, Expected: expectedCompanyRevision, Current: companyRevision.Number}
	}
	if founderHash != companyHash {
		return IntentResult{}, fmt.Errorf("%w: Founder and Company constants differ", ErrInvalidState)
	}

	decision, err := mutate(founder, founderRevision, company, companyRevision)
	if err != nil {
		return IntentResult{}, err
	}
	if err := validateExitDecision(decision, intentID); err != nil {
		return IntentResult{}, err
	}
	decision.Receipt, err = normalizeJSON(decision.Receipt)
	if err != nil {
		return IntentResult{}, err
	}
	if decision.Outcome == IntentRejected {
		if _, err := tx.ExecContext(ctx, `INSERT INTO intent_records(stream_id,intent_id,request_hash,outcome,receipt) VALUES($1,$2,$3,$4,$5)`, companyStreamID, intentID, requestHash, decision.Outcome, decision.Receipt); err != nil {
			return IntentResult{}, err
		}
		if err := tx.Commit(); err != nil {
			return IntentResult{}, err
		}
		return IntentResult{Outcome: IntentRejected, Receipt: cloneRaw(decision.Receipt)}, nil
	}

	founderEncoded, err := s.validatedState(founderHash, economy.ScopeFounder, founder)
	if err != nil {
		return IntentResult{}, err
	}
	finalEncoded, err := s.validatedState(companyHash, economy.ScopeCompany, decision.FinalCompanyState)
	if err != nil {
		return IntentResult{}, err
	}
	newEncoded, err := s.validatedState(companyHash, economy.ScopeCompany, decision.NewCompanyState)
	if err != nil {
		return IntentResult{}, err
	}

	recorded := make([]EventRecord, 0, len(decision.FounderEvents)+len(decision.CompanyEndedEvents)+len(decision.CompanyStartedEvents))
	founderNext := founderRevision.Number + 1
	if _, err := tx.ExecContext(ctx, `INSERT INTO save_revisions(stream_id,revision,version,state,constants_hash) VALUES($1,$2,$3,$4,$5)`, founderStreamID, founderNext, CurrentVersion, founderEncoded, founderHash); err != nil {
		return IntentResult{}, err
	}
	if err := runExitFault(fault, "founder_revision"); err != nil {
		return IntentResult{}, err
	}
	founderRecords, err := insertExitEvents(ctx, tx, founderStreamID, ownerID, founderNext, founderHash, decision.FounderEvents)
	if err != nil {
		return IntentResult{}, err
	}
	recorded = append(recorded, founderRecords...)
	if err := runExitFault(fault, "founder_events"); err != nil {
		return IntentResult{}, err
	}

	companyFinal := companyRevision.Number + 1
	if _, err := tx.ExecContext(ctx, `INSERT INTO save_revisions(stream_id,revision,version,state,constants_hash) VALUES($1,$2,$3,$4,$5)`, companyStreamID, companyFinal, CurrentVersion, finalEncoded, companyHash); err != nil {
		return IntentResult{}, err
	}
	if err := runExitFault(fault, "company_final_revision"); err != nil {
		return IntentResult{}, err
	}
	endedRecords, err := insertExitEvents(ctx, tx, companyStreamID, ownerID, companyFinal, companyHash, decision.CompanyEndedEvents)
	if err != nil {
		return IntentResult{}, err
	}
	recorded = append(recorded, endedRecords...)
	if err := runExitFault(fault, "company_ended_events"); err != nil {
		return IntentResult{}, err
	}

	companyNext := companyRevision.Number + 2
	if _, err := tx.ExecContext(ctx, `INSERT INTO save_revisions(stream_id,revision,version,state,constants_hash) VALUES($1,$2,$3,$4,$5)`, companyStreamID, companyNext, CurrentVersion, newEncoded, companyHash); err != nil {
		return IntentResult{}, err
	}
	if err := runExitFault(fault, "company_started_revision"); err != nil {
		return IntentResult{}, err
	}
	startedRecords, err := insertExitEvents(ctx, tx, companyStreamID, ownerID, companyNext, companyHash, decision.CompanyStartedEvents)
	if err != nil {
		return IntentResult{}, err
	}
	recorded = append(recorded, startedRecords...)
	if err := runExitFault(fault, "company_started_events"); err != nil {
		return IntentResult{}, err
	}

	if _, err := tx.ExecContext(ctx, `INSERT INTO intent_records(stream_id,intent_id,request_hash,outcome,receipt) VALUES($1,$2,$3,$4,$5)`, companyStreamID, intentID, requestHash, decision.Outcome, decision.Receipt); err != nil {
		return IntentResult{}, err
	}
	if err := runExitFault(fault, "intent_record"); err != nil {
		return IntentResult{}, err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM save_revisions WHERE stream_id=$1 AND revision <= $2`, founderStreamID, founderNext-5); err != nil {
		return IntentResult{}, err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM save_revisions WHERE stream_id=$1 AND revision <= $2`, companyStreamID, companyNext-5); err != nil {
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

func (s *Store) loadExitState(ctx context.Context, tx *sql.Tx, streamID, ownerID string, scope economy.Scope) (*State, Revision, string, error) {
	var revision Revision
	var data []byte
	if err := tx.QueryRowContext(ctx, `SELECT revision,version,constants_hash,created_at,state FROM save_revisions WHERE stream_id=$1 ORDER BY revision DESC LIMIT 1`, streamID).Scan(&revision.Number, &revision.Version, &revision.ConstantsHash, &revision.CreatedAt, &data); err != nil {
		return nil, Revision{}, "", err
	}
	revision.StreamID, revision.OwnerID = streamID, ownerID
	catalog, ok := s.catalogs.Resolve(revision.ConstantsHash)
	if !ok {
		return nil, Revision{}, "", fmt.Errorf("%w: unknown catalog", ErrInvalidState)
	}
	state, err := RestoreState(data, revision.Version, catalog, scope, revision.CreatedAt)
	return state, revision, revision.ConstantsHash, err
}

func validateExitDecision(decision ExitDecision, intentID string) error {
	base := IntentDecision{Outcome: decision.Outcome, Receipt: decision.Receipt}
	base.Events = append(base.Events, decision.FounderEvents...)
	base.Events = append(base.Events, decision.CompanyEndedEvents...)
	base.Events = append(base.Events, decision.CompanyStartedEvents...)
	if err := validateIntentDecision(base, intentID); err != nil {
		return err
	}
	if decision.Outcome == IntentApplied {
		if decision.FinalCompanyState == nil || decision.NewCompanyState == nil || len(decision.CompanyEndedEvents) == 0 || len(decision.CompanyStartedEvents) == 0 {
			return ErrInvalidStream
		}
	} else if decision.FinalCompanyState != nil || decision.NewCompanyState != nil {
		return ErrInvalidStream
	}
	return nil
}

func insertExitEvents(ctx context.Context, tx *sql.Tx, streamID, ownerID string, revision int64, constantsHash string, events []EventWrite) ([]EventRecord, error) {
	records := make([]EventRecord, 0, len(events))
	for _, event := range events {
		var eventIntent any
		if event.IntentID != "" {
			eventIntent = event.IntentID
		}
		var record EventRecord
		var storedIntent sql.NullString
		if err := tx.QueryRowContext(ctx, `INSERT INTO events(stream_id,revision,schema_version,kind,intent_id,constants_hash,payload) VALUES($1,$2,$3,$4,$5,$6,$7) RETURNING event_id,kind,intent_id,occurred_at,payload`, streamID, revision, event.SchemaVersion, event.Kind, eventIntent, constantsHash, event.Payload).Scan(&record.EventID, &record.Kind, &storedIntent, &record.OccurredAt, &record.Payload); err != nil {
			return nil, err
		}
		record.StreamID, record.OwnerID, record.Revision, record.ConstantsHash = streamID, ownerID, revision, constantsHash
		if storedIntent.Valid {
			record.IntentID = storedIntent.String
		}
		records = append(records, record)
	}
	return records, nil
}

func eventsForExitIntent(ctx context.Context, tx *sql.Tx, streams []string, ownerID, intentID string) ([]EventRecord, error) {
	if len(streams) != 2 {
		return nil, ErrInvalidStream
	}
	rows, err := tx.QueryContext(ctx, `SELECT event_id,stream_id,revision,kind,intent_id,constants_hash,occurred_at,payload FROM events WHERE stream_id IN ($1,$2) AND intent_id=$3 ORDER BY occurred_at,event_id`, streams[0], streams[1], intentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var events []EventRecord
	for rows.Next() {
		var record EventRecord
		var storedIntent sql.NullString
		if err := rows.Scan(&record.EventID, &record.StreamID, &record.Revision, &record.Kind, &storedIntent, &record.ConstantsHash, &record.OccurredAt, &record.Payload); err != nil {
			return nil, err
		}
		record.OwnerID = ownerID
		if storedIntent.Valid {
			record.IntentID = storedIntent.String
		}
		events = append(events, record)
	}
	sort.SliceStable(events, func(i, j int) bool {
		left, right := exitEventOrder(events[i].Kind), exitEventOrder(events[j].Kind)
		if left != right {
			return left < right
		}
		if events[i].Revision != events[j].Revision {
			return events[i].Revision < events[j].Revision
		}
		if events[i].StreamID != events[j].StreamID {
			return events[i].StreamID < events[j].StreamID
		}
		return events[i].EventID < events[j].EventID
	})
	return events, rows.Err()
}

func exitEventOrder(kind EventKind) int {
	switch kind {
	case EventFounderAdvanced:
		return 0
	case EventRunEnded:
		return 1
	case EventRunStarted:
		return 2
	default:
		return 3
	}
}

func runExitFault(inject ExitFaultInjector, step string) error {
	if inject == nil {
		return nil
	}
	return inject(step)
}

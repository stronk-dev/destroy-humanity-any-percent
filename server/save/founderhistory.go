package save

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
)

// FounderHistory is one repeatable-read snapshot of the immutable Founder
// replay evidence and the authoritative head it must reproduce.
type FounderHistory struct {
	FounderStreamID string
	FounderID       string
	Genesis         FounderGenesis
	Entries         []FounderHistoryEntry
	HeadRevision    int64
	HeadVersion     int
	HeadConstants   string
	HeadState       []byte
}

type FounderHistoryEntry struct {
	Sequence         int64
	IntentID         string
	CanonicalPayload []byte
	ReplayInputs     []byte
	Receipt          []byte
	AppliedRevision  *int64
	ConstantsHash    string
	ServerTSMS       int64
	Source           *FounderLogSource
	Events           []EventWrite
}

// LoadFounderHistory reads genesis, log, events, and the current head from one
// repeatable-read snapshot so verification never mixes two committed heads.
func (s *Store) LoadFounderHistory(ctx context.Context, founderStreamID string) (FounderHistory, error) {
	if s == nil || !uuidPattern.MatchString(founderStreamID) {
		return FounderHistory{}, ErrInvalidStream
	}
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true, Isolation: sql.LevelRepeatableRead})
	if err != nil {
		return FounderHistory{}, err
	}
	defer tx.Rollback()
	history := FounderHistory{FounderStreamID: founderStreamID, Entries: []FounderHistoryEntry{}}
	var scope string
	if err := tx.QueryRowContext(ctx, `SELECT owner_id,scope FROM save_streams
		WHERE id=$1 AND owner_kind='founder'`, founderStreamID).
		Scan(&history.FounderID, &scope); errors.Is(err, sql.ErrNoRows) {
		return FounderHistory{}, ErrNotFound
	} else if err != nil {
		return FounderHistory{}, err
	}
	if scope != "founder" || !uuidPattern.MatchString(history.FounderID) {
		return FounderHistory{}, ErrInvalidStream
	}
	history.Genesis.FounderStreamID = founderStreamID
	if err := tx.QueryRowContext(ctx, `SELECT revision,state,version,constants_hash
		FROM founder_genesis WHERE founder_stream_id=$1`, founderStreamID).
		Scan(&history.Genesis.Revision, &history.Genesis.State, &history.Genesis.Version, &history.Genesis.ConstantsHash); errors.Is(err, sql.ErrNoRows) {
		return FounderHistory{}, ErrFounderGenesisUnavailable
	} else if err != nil {
		return FounderHistory{}, err
	}
	if err := tx.QueryRowContext(ctx, `SELECT revision,version,constants_hash,state::text
		FROM save_revisions WHERE stream_id=$1 ORDER BY revision DESC LIMIT 1`, founderStreamID).
		Scan(&history.HeadRevision, &history.HeadVersion, &history.HeadConstants, &history.HeadState); err != nil {
		return FounderHistory{}, err
	}
	rows, err := tx.QueryContext(ctx, `SELECT seq,intent_id,canonical_payload,replay_inputs::text,receipt::text,
		applied_revision,constants_hash,server_ts_ms,source_company_stream_id,source_run_seq,source_run_log_seq
		FROM founder_log WHERE founder_stream_id=$1 ORDER BY seq`, founderStreamID)
	if err != nil {
		return FounderHistory{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var entry FounderHistoryEntry
		var replayInputs, receipt string
		var applied, sourceRun, sourceLog sql.NullInt64
		var sourceStream sql.NullString
		if err := rows.Scan(&entry.Sequence, &entry.IntentID, &entry.CanonicalPayload, &replayInputs, &receipt,
			&applied, &entry.ConstantsHash, &entry.ServerTSMS, &sourceStream, &sourceRun, &sourceLog); err != nil {
			return FounderHistory{}, err
		}
		entry.ReplayInputs, entry.Receipt = []byte(replayInputs), []byte(receipt)
		if applied.Valid {
			value := applied.Int64
			entry.AppliedRevision = &value
		}
		if sourceStream.Valid || sourceRun.Valid || sourceLog.Valid {
			if !sourceStream.Valid || !sourceRun.Valid || !sourceLog.Valid {
				return FounderHistory{}, fmt.Errorf("%w: partial Founder source coordinates", ErrInvalidState)
			}
			entry.Source = &FounderLogSource{CompanyStreamID: sourceStream.String, RunSeq: sourceRun.Int64, RunLogSeq: sourceLog.Int64}
		}
		if !json.Valid(entry.CanonicalPayload) || !json.Valid(entry.ReplayInputs) || !json.Valid(entry.Receipt) ||
			!hashPattern.MatchString(entry.ConstantsHash) || entry.ServerTSMS < 1 {
			return FounderHistory{}, fmt.Errorf("%w: malformed Founder history row", ErrInvalidState)
		}
		history.Entries = append(history.Entries, entry)
	}
	if err := rows.Err(); err != nil {
		return FounderHistory{}, err
	}
	if err := rows.Close(); err != nil {
		return FounderHistory{}, err
	}
	for index := range history.Entries {
		events, err := loadFounderHistoryEvents(ctx, tx, founderStreamID, history.Entries[index].IntentID)
		if err != nil {
			return FounderHistory{}, err
		}
		history.Entries[index].Events = events
	}
	if !json.Valid(history.Genesis.State) || !json.Valid(history.HeadState) ||
		history.Genesis.Revision < 1 || history.HeadRevision < history.Genesis.Revision ||
		!hashPattern.MatchString(history.Genesis.ConstantsHash) || !hashPattern.MatchString(history.HeadConstants) {
		return FounderHistory{}, fmt.Errorf("%w: malformed Founder history boundary", ErrInvalidState)
	}
	if err := tx.Commit(); err != nil {
		return FounderHistory{}, err
	}
	return history, nil
}

func loadFounderHistoryEvents(ctx context.Context, tx *sql.Tx, streamID, intentID string) ([]EventWrite, error) {
	rows, err := tx.QueryContext(ctx, `SELECT kind,schema_version,intent_id,payload::text
		FROM events WHERE stream_id=$1 AND intent_id=$2 ORDER BY event_seq,event_id`, streamID, intentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []EventWrite{}
	for rows.Next() {
		var value EventWrite
		var payload string
		if err := rows.Scan(&value.Kind, &value.SchemaVersion, &value.IntentID, &payload); err != nil {
			return nil, err
		}
		value.Payload = json.RawMessage(payload)
		if !json.Valid(value.Payload) {
			return nil, fmt.Errorf("%w: malformed Founder event", ErrInvalidState)
		}
		result = append(result, value)
	}
	return result, rows.Err()
}

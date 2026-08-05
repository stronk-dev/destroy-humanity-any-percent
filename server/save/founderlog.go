package save

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"cloud-clicker/server/decimal"
)

const FounderReplayInputsVersion = 1

type FounderReplayCommand struct {
	IntentID        string `json:"intent_id"`
	FounderStreamID string `json:"founder_stream_id"`
	FounderID       string `json:"founder_id"`
	Revision        int64  `json:"revision"`
	FounderLogSeq   int64  `json:"founder_log_seq"`
	ServerTSMS      int64  `json:"server_ts_ms"`
}

type founderReplayInputsEnvelope struct {
	Version       int                  `json:"v"`
	Command       FounderReplayCommand `json:"command"`
	EvaluatedAtMS int64                `json:"evaluated_at_ms"`
	Resolved      json.RawMessage      `json:"resolved"`
}

// ValidateFounderReplayInputs owns the reusable Founder envelope. Feature
// packages remain responsible for the exact closed union under resolved.
func ValidateFounderReplayInputs(data []byte, expected FounderReplayCommand) (json.RawMessage, error) {
	var envelope founderReplayInputsEnvelope
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&envelope); err != nil || ensureJSONEnd(decoder) != nil {
		return nil, fmt.Errorf("%w: invalid Founder replay inputs", ErrInvalidStream)
	}
	resolved, resolvedErr := normalizeJSON(envelope.Resolved)
	if envelope.Version != FounderReplayInputsVersion || envelope.Command != expected ||
		!uuidV7Pattern.MatchString(envelope.Command.IntentID) ||
		!uuidPattern.MatchString(envelope.Command.FounderStreamID) ||
		!uuidPattern.MatchString(envelope.Command.FounderID) ||
		envelope.Command.Revision < 1 || envelope.Command.Revision > decimal.MaxExactInteger ||
		envelope.Command.FounderLogSeq < 1 || envelope.Command.FounderLogSeq > decimal.MaxExactInteger ||
		envelope.Command.ServerTSMS <= 0 || envelope.Command.ServerTSMS > decimal.MaxExactInteger ||
		envelope.EvaluatedAtMS != envelope.Command.ServerTSMS ||
		resolvedErr != nil || len(resolved) < 2 || resolved[0] != '{' {
		return nil, fmt.Errorf("%w: invalid Founder replay inputs", ErrInvalidStream)
	}
	normalized, err := normalizeJSON(data)
	if err != nil || !jsonObject(normalized) {
		return nil, fmt.Errorf("%w: invalid Founder replay inputs", ErrInvalidStream)
	}
	return normalized, nil
}

func founderServerTimestamp(ctx context.Context, tx *sql.Tx) (int64, error) {
	var serverTSMS int64
	if err := tx.QueryRowContext(ctx, `SELECT (extract(epoch FROM clock_timestamp())*1000)::bigint`).Scan(&serverTSMS); err != nil {
		return 0, err
	}
	if serverTSMS <= 0 || serverTSMS > decimal.MaxExactInteger || time.UnixMilli(serverTSMS).IsZero() {
		return 0, errors.New("invalid Founder command server timestamp")
	}
	return serverTSMS, nil
}

func nextFounderLogSequence(ctx context.Context, tx *sql.Tx, founderStreamID string) (int64, error) {
	var current sql.NullInt64
	if err := tx.QueryRowContext(ctx, `SELECT max(seq) FROM founder_log WHERE founder_stream_id=$1`, founderStreamID).Scan(&current); err != nil {
		return 0, err
	}
	next := int64(1)
	if current.Valid {
		if current.Int64 >= decimal.MaxExactInteger {
			return 0, ErrInvalidState
		}
		next = current.Int64 + 1
	}
	return next, nil
}

func insertFounderLog(ctx context.Context, tx *sql.Tx, command FounderReplayCommand, constantsHash string,
	canonicalPayload, replayInputs, receipt []byte, appliedRevision *int64,
) error {
	if !uuidPattern.MatchString(command.FounderStreamID) || !uuidPattern.MatchString(command.FounderID) ||
		!uuidV7Pattern.MatchString(command.IntentID) || command.Revision < 1 || command.FounderLogSeq < 1 ||
		command.ServerTSMS < 1 || command.ServerTSMS > decimal.MaxExactInteger ||
		!hashPattern.MatchString(constantsHash) || len(canonicalPayload) == 0 || len(replayInputs) == 0 {
		return ErrInvalidStream
	}
	var revision any
	if appliedRevision != nil {
		if *appliedRevision < 1 || *appliedRevision > decimal.MaxExactInteger {
			return ErrInvalidStream
		}
		revision = *appliedRevision
	}
	result, err := tx.ExecContext(ctx, `
		INSERT INTO founder_log(founder_stream_id,seq,intent_id,canonical_payload,replay_inputs,receipt,
			applied_revision,constants_hash,server_ts_ms)
		VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9)`, command.FounderStreamID, command.FounderLogSeq,
		command.IntentID, canonicalPayload, replayInputs, receipt, revision, constantsHash, command.ServerTSMS)
	if err != nil {
		return err
	}
	if affected, rowsErr := result.RowsAffected(); rowsErr != nil || affected != 1 {
		if rowsErr != nil {
			return rowsErr
		}
		return ErrInvalidStream
	}
	return nil
}

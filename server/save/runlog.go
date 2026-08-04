package save

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"cloud-clicker/server/decimal"
)

const ReplayInputsVersion = 4

type ReplayCommand struct {
	IntentID        string `json:"intent_id"`
	CompanyStreamID string `json:"company_stream_id"`
	FounderID       string `json:"founder_id"`
	Revision        int64  `json:"revision"`
	RunSeq          int64  `json:"run_seq"`
	RunLogSeq       int64  `json:"run_log_seq"`
}

type replayInputsEnvelope struct {
	Version        int             `json:"v"`
	Command        ReplayCommand   `json:"command"`
	EvaluatedAtMS  int64           `json:"evaluated_at_ms"`
	EvaluationMode string          `json:"evaluation_mode"`
	Resolved       json.RawMessage `json:"resolved"`
}

// ValidateReplayInputs validates the persistence-owned envelope and returns
// normalized JSON. The production kernel owns strict validation of the closed
// per-intent resolved union; persistence still refuses any non-object or
// mismatched authoritative command coordinates.
func ValidateReplayInputs(data []byte, expected ReplayCommand) (json.RawMessage, error) {
	var envelope replayInputsEnvelope
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&envelope); err != nil || ensureJSONEnd(decoder) != nil {
		return nil, fmt.Errorf("%w: invalid replay inputs", ErrInvalidStream)
	}
	resolved, resolvedErr := normalizeJSON(envelope.Resolved)
	if envelope.Version != ReplayInputsVersion || envelope.Command != expected ||
		!uuidV7Pattern.MatchString(envelope.Command.IntentID) ||
		!uuidPattern.MatchString(envelope.Command.CompanyStreamID) ||
		!uuidPattern.MatchString(envelope.Command.FounderID) ||
		envelope.Command.Revision < 1 || envelope.Command.Revision > decimal.MaxExactInteger ||
		envelope.Command.RunSeq < 1 || envelope.Command.RunSeq > decimal.MaxExactInteger ||
		envelope.Command.RunLogSeq < 1 || envelope.Command.RunLogSeq > decimal.MaxExactInteger ||
		envelope.EvaluatedAtMS <= 0 || envelope.EvaluatedAtMS > decimal.MaxExactInteger ||
		(envelope.EvaluationMode != "online" && envelope.EvaluationMode != "offline") ||
		resolvedErr != nil || len(resolved) < 2 || resolved[0] != '{' {
		return nil, fmt.Errorf("%w: invalid replay inputs", ErrInvalidStream)
	}
	normalized, err := normalizeJSON(data)
	if err != nil || !jsonObject(normalized) {
		return nil, fmt.Errorf("%w: invalid replay inputs", ErrInvalidStream)
	}
	return normalized, nil
}

func validateCanonicalPayload(payload []byte, requestHash string) error {
	if len(payload) < 2 || !jsonObject(payload) {
		return ErrInvalidStream
	}
	digest := sha256.Sum256(payload)
	if "sha256:"+hex.EncodeToString(digest[:]) != requestHash {
		return ErrInvalidStream
	}
	return nil
}

func jsonObject(payload []byte) bool {
	normalized, err := normalizeJSON(payload)
	return err == nil && len(normalized) == len(payload) && string(normalized) == string(payload) && payload[0] == '{'
}

func nextRunLogSequence(ctx context.Context, tx *sql.Tx, companyStreamID string, runSeq int64) (int64, error) {
	if runSeq < 1 || runSeq > decimal.MaxExactInteger {
		return 0, ErrInvalidState
	}
	var current sql.NullInt64
	if err := tx.QueryRowContext(ctx, `SELECT max(seq) FROM run_log WHERE company_stream_id=$1 AND run_seq=$2`, companyStreamID, runSeq).Scan(&current); err != nil {
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

func insertRunLog(ctx context.Context, tx *sql.Tx, companyStreamID string, runSeq, sequence int64, intentID string, canonicalPayload, replayInputs, receipt []byte, appliedRevision *int64) error {
	if sequence < 1 || runSeq < 1 || !uuidPattern.MatchString(companyStreamID) || !uuidV7Pattern.MatchString(intentID) || len(canonicalPayload) == 0 || len(replayInputs) == 0 {
		return ErrInvalidStream
	}
	var revision any
	if appliedRevision != nil {
		if *appliedRevision < 1 {
			return ErrInvalidStream
		}
		revision = *appliedRevision
	}
	var serverTSMS int64
	err := tx.QueryRowContext(ctx, `
		INSERT INTO run_log(company_stream_id,run_seq,seq,intent_id,canonical_payload,replay_inputs,receipt,applied_revision,server_ts_ms)
		VALUES($1,$2,$3,$4,$5,$6,$7,$8,(extract(epoch FROM clock_timestamp())*1000)::bigint)
		RETURNING server_ts_ms`, companyStreamID, runSeq, sequence, intentID, canonicalPayload, replayInputs, receipt, revision).Scan(&serverTSMS)
	if err != nil {
		return err
	}
	if serverTSMS <= 0 || time.UnixMilli(serverTSMS).IsZero() {
		return errors.New("invalid run-log server timestamp")
	}
	return nil
}

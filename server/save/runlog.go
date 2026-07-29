package save

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"time"

	"cloud-clicker/server/decimal"
)

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

func insertRunLog(ctx context.Context, tx *sql.Tx, companyStreamID string, runSeq, sequence int64, intentID string, canonicalPayload, receipt []byte, appliedRevision *int64) error {
	if sequence < 1 || runSeq < 1 || !uuidPattern.MatchString(companyStreamID) || !uuidV7Pattern.MatchString(intentID) || len(canonicalPayload) == 0 {
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
		INSERT INTO run_log(company_stream_id,run_seq,seq,intent_id,canonical_payload,receipt,applied_revision,server_ts_ms)
		VALUES($1,$2,$3,$4,$5,$6,$7,(extract(epoch FROM clock_timestamp())*1000)::bigint)
		RETURNING server_ts_ms`, companyStreamID, runSeq, sequence, intentID, canonicalPayload, receipt, revision).Scan(&serverTSMS)
	if err != nil {
		return err
	}
	if serverTSMS <= 0 || time.UnixMilli(serverTSMS).IsZero() {
		return errors.New("invalid run-log server timestamp")
	}
	return nil
}

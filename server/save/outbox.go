package save

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"
	"unicode/utf8"
)

type ReceiptOutboxItem struct {
	ID            int64
	ClaimToken    string
	FounderID     string
	CompanyStream string
	IntentID      string
	Revision      int64
	ConstantsHash string
	Receipt       json.RawMessage
	OccurredAt    time.Time
	AttemptCount  int
}

const MaxReceiptOutboxBytes = 60 * 1024

func insertReceiptOutbox(ctx context.Context, tx *sql.Tx, founderID, companyStreamID, intentID string, revision int64, constantsHash string, receipt json.RawMessage) error {
	if tx == nil || !uuidPattern.MatchString(founderID) || !uuidPattern.MatchString(companyStreamID) || !uuidV7Pattern.MatchString(intentID) ||
		revision < 1 || !hashPattern.MatchString(constantsHash) || len(receipt) == 0 || len(receipt) > MaxReceiptOutboxBytes {
		return ErrInvalidStream
	}
	var insertedID int64
	err := tx.QueryRowContext(ctx, `
		INSERT INTO transport_receipt_outbox(founder_id,company_stream_id,intent_id,revision,constants_hash,receipt)
		SELECT $1,$2,$3,$4,$5,$6
		WHERE octet_length($6::jsonb::text) <= $7
		RETURNING outbox_id`, founderID, companyStreamID, intentID, revision, constantsHash, receipt, MaxReceiptOutboxBytes).Scan(&insertedID)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrInvalidStream
	}
	return err
}

func (s *Store) ClaimReceiptOutbox(ctx context.Context, limit int, lease time.Duration) ([]ReceiptOutboxItem, error) {
	if limit < 1 || limit > 1_000 || lease < time.Second || lease > 5*time.Minute {
		return nil, ErrInvalidStream
	}
	rows, err := s.db.QueryContext(ctx, `
		WITH selected AS (
			SELECT outbox_id
			FROM transport_receipt_outbox AS candidate
			WHERE published_at IS NULL AND dead_lettered_at IS NULL
			  AND (claimed_until IS NULL OR claimed_until <= clock_timestamp())
			  AND NOT EXISTS (
			    SELECT 1 FROM transport_receipt_outbox AS earlier
			    WHERE earlier.founder_id=candidate.founder_id
			      AND earlier.published_at IS NULL AND earlier.dead_lettered_at IS NULL
			      AND earlier.outbox_id<candidate.outbox_id
			  )
			ORDER BY outbox_id
			FOR UPDATE SKIP LOCKED
			LIMIT $1
		)
		UPDATE transport_receipt_outbox AS outbox
		SET claim_token=gen_random_uuid(), claimed_until=clock_timestamp()+$2::interval
		FROM selected
		WHERE outbox.outbox_id=selected.outbox_id
		RETURNING outbox.outbox_id,outbox.claim_token,outbox.founder_id,outbox.company_stream_id,
		          outbox.intent_id,outbox.revision,outbox.constants_hash,outbox.receipt,outbox.occurred_at,
		          outbox.attempt_count`, limit, intervalLiteral(lease))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]ReceiptOutboxItem, 0, limit)
	for rows.Next() {
		var item ReceiptOutboxItem
		if err := rows.Scan(&item.ID, &item.ClaimToken, &item.FounderID, &item.CompanyStream, &item.IntentID,
			&item.Revision, &item.ConstantsHash, &item.Receipt, &item.OccurredAt, &item.AttemptCount); err != nil {
			return nil, err
		}
		normalized, err := normalizeJSON(item.Receipt)
		if err != nil {
			return nil, fmt.Errorf("%w: outbox receipt: %v", ErrInvalidState, err)
		}
		item.Receipt = normalized
		item.OccurredAt = item.OccurredAt.UTC()
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	sort.Slice(items, func(left, right int) bool { return items[left].ID < items[right].ID })
	return items, nil
}

func (s *Store) MarkReceiptPublished(ctx context.Context, id int64, claimToken string) error {
	if id < 1 || !uuidPattern.MatchString(claimToken) {
		return ErrInvalidStream
	}
	result, err := s.db.ExecContext(ctx, `
		UPDATE transport_receipt_outbox
		SET published_at=clock_timestamp(),claim_token=NULL,claimed_until=NULL
		WHERE outbox_id=$1 AND claim_token=$2 AND published_at IS NULL`, id, claimToken)
	if err != nil {
		return err
	}
	changed, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if changed != 1 {
		return ErrConflict
	}
	return nil
}

func (s *Store) ReleaseReceiptClaim(ctx context.Context, id int64, claimToken string) error {
	if id < 1 || !uuidPattern.MatchString(claimToken) {
		return ErrInvalidStream
	}
	result, err := s.db.ExecContext(ctx, `
		UPDATE transport_receipt_outbox SET claim_token=NULL,claimed_until=NULL
		WHERE outbox_id=$1 AND claim_token=$2 AND published_at IS NULL`, id, claimToken)
	if err != nil {
		return err
	}
	changed, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if changed != 1 {
		return ErrConflict
	}
	return nil
}

func (s *Store) FailReceiptClaim(ctx context.Context, id int64, claimToken, detail string, maxAttempts int) (bool, error) {
	detail = strings.TrimSpace(detail)
	if id < 1 || !uuidPattern.MatchString(claimToken) || detail == "" || !utf8.ValidString(detail) || utf8.RuneCountInString(detail) > 512 || maxAttempts < 1 || maxAttempts > 1000 {
		return false, ErrInvalidStream
	}
	var dead bool
	err := s.db.QueryRowContext(ctx, `
		UPDATE transport_receipt_outbox
		SET attempt_count=attempt_count+1,
		    last_error=$3,
		    dead_lettered_at=CASE WHEN attempt_count+1 >= $4 THEN clock_timestamp() ELSE NULL END,
		    claim_token=NULL,
		    claimed_until=NULL
		WHERE outbox_id=$1 AND claim_token=$2 AND published_at IS NULL AND dead_lettered_at IS NULL
		RETURNING dead_lettered_at IS NOT NULL`, id, claimToken, detail, maxAttempts).Scan(&dead)
	if errors.Is(err, sql.ErrNoRows) {
		return false, ErrConflict
	}
	return dead, err
}

func intervalLiteral(duration time.Duration) string {
	return fmt.Sprintf("%d milliseconds", duration.Milliseconds())
}

func (s *Store) PendingReceiptCount(ctx context.Context) (int64, error) {
	var count int64
	err := s.db.QueryRowContext(ctx, `SELECT count(*) FROM transport_receipt_outbox WHERE published_at IS NULL AND dead_lettered_at IS NULL`).Scan(&count)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, nil
	}
	return count, err
}

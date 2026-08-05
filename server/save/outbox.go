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

	"cloud-clicker/server/economy"
)

type PlayerOutboxItem struct {
	ID            int64
	ClaimToken    string
	FounderID     string
	StreamID      string
	MessageKind   string
	SourceID      string
	Scope         string
	Revision      int64
	ConstantsHash string
	Payload       json.RawMessage
	OccurredAt    time.Time
	AttemptCount  int
}

const MaxPlayerOutboxBytes = 60 * 1024

func insertReceiptOutbox(ctx context.Context, tx *sql.Tx, founderID, streamID, intentID string, scope economy.Scope, revision int64, constantsHash string, receipt json.RawMessage) error {
	if tx == nil || !uuidPattern.MatchString(founderID) || !uuidPattern.MatchString(streamID) || !uuidV7Pattern.MatchString(intentID) ||
		(scope != economy.ScopeCompany && scope != economy.ScopeFounder) ||
		revision < 1 || !hashPattern.MatchString(constantsHash) || len(receipt) == 0 || len(receipt) > MaxPlayerOutboxBytes {
		return ErrInvalidStream
	}
	var insertedID int64
	err := tx.QueryRowContext(ctx, `
		INSERT INTO transport_player_outbox(founder_id,stream_id,message_kind,source_id,scope,revision,constants_hash,payload)
		SELECT $1,$2,'receipt',$3,$4,$5,$6,$7
		WHERE octet_length($7::jsonb::text) <= $8
		RETURNING outbox_id`, founderID, streamID, intentID, scope, revision, constantsHash, receipt, MaxPlayerOutboxBytes).Scan(&insertedID)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrInvalidStream
	}
	return err
}

func (s *Store) ClaimPlayerOutbox(ctx context.Context, limit int, lease time.Duration) ([]PlayerOutboxItem, error) {
	if limit < 1 || limit > 1_000 || lease < time.Second || lease > 5*time.Minute {
		return nil, ErrInvalidStream
	}
	rows, err := s.db.QueryContext(ctx, `
		WITH selected AS (
			SELECT outbox_id
			FROM transport_player_outbox AS candidate
			WHERE published_at IS NULL AND dead_lettered_at IS NULL
			  AND (claimed_until IS NULL OR claimed_until <= clock_timestamp())
			  AND NOT EXISTS (
			    SELECT 1 FROM transport_player_outbox AS earlier
			    WHERE earlier.founder_id=candidate.founder_id
			      AND earlier.published_at IS NULL AND earlier.dead_lettered_at IS NULL
			      AND earlier.outbox_id<candidate.outbox_id
			  )
			ORDER BY outbox_id
			FOR UPDATE SKIP LOCKED
			LIMIT $1
		)
		UPDATE transport_player_outbox AS outbox
		SET claim_token=gen_random_uuid(), claimed_until=clock_timestamp()+$2::interval
		FROM selected
		WHERE outbox.outbox_id=selected.outbox_id
		RETURNING outbox.outbox_id,outbox.claim_token,outbox.founder_id,outbox.stream_id,
		          outbox.message_kind,outbox.source_id,outbox.scope,outbox.revision,outbox.constants_hash,outbox.payload,outbox.occurred_at,
		          outbox.attempt_count`, limit, intervalLiteral(lease))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]PlayerOutboxItem, 0, limit)
	for rows.Next() {
		var item PlayerOutboxItem
		if err := rows.Scan(&item.ID, &item.ClaimToken, &item.FounderID, &item.StreamID, &item.MessageKind, &item.SourceID,
			&item.Scope, &item.Revision, &item.ConstantsHash, &item.Payload, &item.OccurredAt, &item.AttemptCount); err != nil {
			return nil, err
		}
		normalized, err := normalizeJSON(item.Payload)
		if err != nil {
			return nil, fmt.Errorf("%w: player outbox payload: %v", ErrInvalidState, err)
		}
		if item.MessageKind != "receipt" && item.MessageKind != "event" || item.Scope != "company" && item.Scope != "founder" {
			return nil, fmt.Errorf("%w: player outbox envelope", ErrInvalidState)
		}
		item.Payload = normalized
		item.OccurredAt = item.OccurredAt.UTC()
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	sort.Slice(items, func(left, right int) bool { return items[left].ID < items[right].ID })
	return items, nil
}

func (s *Store) MarkPlayerPublished(ctx context.Context, id int64, claimToken string) error {
	if id < 1 || !uuidPattern.MatchString(claimToken) {
		return ErrInvalidStream
	}
	result, err := s.db.ExecContext(ctx, `
		UPDATE transport_player_outbox
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

func (s *Store) ReleasePlayerClaim(ctx context.Context, id int64, claimToken string) error {
	if id < 1 || !uuidPattern.MatchString(claimToken) {
		return ErrInvalidStream
	}
	result, err := s.db.ExecContext(ctx, `
		UPDATE transport_player_outbox SET claim_token=NULL,claimed_until=NULL
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

func (s *Store) DeferPlayerClaim(ctx context.Context, id int64, claimToken, detail string, delay time.Duration) error {
	detail = strings.TrimSpace(detail)
	if id < 1 || !uuidPattern.MatchString(claimToken) || detail == "" || !utf8.ValidString(detail) || utf8.RuneCountInString(detail) > 512 ||
		delay < 100*time.Millisecond || delay > 5*time.Minute {
		return ErrInvalidStream
	}
	result, err := s.db.ExecContext(ctx, `
		UPDATE transport_player_outbox
		SET claimed_until=clock_timestamp()+$3::interval,last_error=$4
		WHERE outbox_id=$1 AND claim_token=$2 AND published_at IS NULL AND dead_lettered_at IS NULL`,
		id, claimToken, intervalLiteral(delay), detail)
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

func (s *Store) FailPlayerClaim(ctx context.Context, id int64, claimToken, detail string, maxAttempts int) (bool, error) {
	detail = strings.TrimSpace(detail)
	if id < 1 || !uuidPattern.MatchString(claimToken) || detail == "" || !utf8.ValidString(detail) || utf8.RuneCountInString(detail) > 512 || maxAttempts < 1 || maxAttempts > 1000 {
		return false, ErrInvalidStream
	}
	var dead bool
	err := s.db.QueryRowContext(ctx, `
		UPDATE transport_player_outbox
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

func (s *Store) PendingPlayerCount(ctx context.Context) (int64, error) {
	var count int64
	err := s.db.QueryRowContext(ctx, `SELECT count(*) FROM transport_player_outbox WHERE published_at IS NULL AND dead_lettered_at IS NULL`).Scan(&count)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, nil
	}
	return count, err
}

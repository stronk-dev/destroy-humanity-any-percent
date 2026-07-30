package guild

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type PresenceItem struct {
	ID            int64
	ClaimToken    string
	GuildID       string
	AccountID     string
	Kind          string
	GuildRevision int64
	ActiveCount   int64
	OccurredAt    time.Time
}

func (service *Service) ClaimPresence(ctx context.Context, limit int, lease time.Duration) ([]PresenceItem, error) {
	if service == nil || limit < 1 || limit > 1_000 || lease < time.Second || lease > 5*time.Minute {
		return nil, ErrInvalidIntent
	}
	rows, err := service.db.QueryContext(ctx, `
		WITH selected AS (
		  SELECT outbox_id FROM guild_presence_outbox
		  WHERE published_at IS NULL AND (claimed_until IS NULL OR claimed_until<=clock_timestamp())
		  ORDER BY outbox_id FOR UPDATE SKIP LOCKED LIMIT $1
		)
		UPDATE guild_presence_outbox AS outbox
		SET claim_token=gen_random_uuid(),claimed_until=clock_timestamp()+$2::interval
		FROM selected WHERE outbox.outbox_id=selected.outbox_id
		RETURNING outbox.outbox_id,outbox.claim_token,outbox.guild_id,outbox.account_id,outbox.kind,
		          outbox.guild_revision,outbox.active_count,outbox.occurred_at`, limit, fmt.Sprintf("%d milliseconds", lease.Milliseconds()))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]PresenceItem, 0, limit)
	for rows.Next() {
		var item PresenceItem
		if err := rows.Scan(&item.ID, &item.ClaimToken, &item.GuildID, &item.AccountID, &item.Kind,
			&item.GuildRevision, &item.ActiveCount, &item.OccurredAt); err != nil {
			return nil, err
		}
		item.OccurredAt = item.OccurredAt.UTC()
		items = append(items, item)
	}
	return items, rows.Err()
}

func (service *Service) MarkPresencePublished(ctx context.Context, id int64, claimToken string) error {
	if service == nil || id < 1 || !uuidPattern.MatchString(claimToken) {
		return ErrInvalidIntent
	}
	result, err := service.db.ExecContext(ctx, `UPDATE guild_presence_outbox
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
		return sql.ErrNoRows
	}
	return nil
}

func (service *Service) ReleasePresence(ctx context.Context, id int64, claimToken string) error {
	if service == nil || id < 1 || !uuidPattern.MatchString(claimToken) {
		return ErrInvalidIntent
	}
	result, err := service.db.ExecContext(ctx, `UPDATE guild_presence_outbox SET claim_token=NULL,claimed_until=NULL
		WHERE outbox_id=$1 AND claim_token=$2 AND published_at IS NULL`, id, claimToken)
	if err != nil {
		return err
	}
	changed, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if changed != 1 {
		return errors.New("guild presence claim conflict")
	}
	return nil
}

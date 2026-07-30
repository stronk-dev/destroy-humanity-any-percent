package guild

import (
	"context"
	"encoding/json"
	"time"
)

func (service *Service) SweepDisbanded(ctx context.Context, now time.Time, limit int) (int, error) {
	if service == nil || now.IsZero() || limit < 1 || limit > 1_000 {
		return 0, ErrInvalidIntent
	}
	now = now.UTC().Truncate(time.Millisecond)
	tx, err := service.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	rows, err := tx.QueryContext(ctx, `SELECT guild_id,revision FROM guilds
		WHERE disbanded_at IS NULL AND below_floor_since IS NOT NULL
		  AND below_floor_since <= $1::timestamptz-make_interval(days => $2::int)
		ORDER BY guild_id FOR UPDATE SKIP LOCKED LIMIT $3`, now, service.catalog.GraceDays, limit)
	if err != nil {
		return 0, err
	}
	type candidate struct {
		id       string
		revision int64
	}
	var candidates []candidate
	for rows.Next() {
		var value candidate
		if err := rows.Scan(&value.id, &value.revision); err != nil {
			rows.Close()
			return 0, err
		}
		candidates = append(candidates, value)
	}
	if err := rows.Close(); err != nil {
		return 0, err
	}
	disbanded := 0
	for _, value := range candidates {
		memberRows, err := tx.QueryContext(ctx, `SELECT account_id FROM guild_members WHERE guild_id=$1 AND left_at IS NULL ORDER BY account_id FOR UPDATE`, value.id)
		if err != nil {
			return 0, err
		}
		var accounts []string
		for memberRows.Next() {
			var account string
			if err := memberRows.Scan(&account); err != nil {
				memberRows.Close()
				return 0, err
			}
			accounts = append(accounts, account)
		}
		if err := memberRows.Close(); err != nil {
			return 0, err
		}
		if len(accounts) >= service.catalog.MinMembers {
			if _, err := tx.ExecContext(ctx, `UPDATE guilds SET below_floor_since=NULL WHERE guild_id=$1`, value.id); err != nil {
				return 0, err
			}
			continue
		}
		value.revision++
		if _, err := tx.ExecContext(ctx, `UPDATE guild_members SET left_at=$2 WHERE guild_id=$1 AND left_at IS NULL`, value.id, now); err != nil {
			return 0, err
		}
		if _, err := tx.ExecContext(ctx, `UPDATE guilds SET revision=$2,disbanded_at=$3 WHERE guild_id=$1`, value.id, value.revision, now); err != nil {
			return 0, err
		}
		payload, _ := json.Marshal(map[string]any{"reason": "below_member_floor"})
		if _, err := tx.ExecContext(ctx, `INSERT INTO guild_events(guild_id,revision,kind,payload) VALUES($1,$2,'guild_disbanded',$3)`, value.id, value.revision, payload); err != nil {
			return 0, err
		}
		for _, account := range accounts {
			if err := insertPresence(ctx, tx, value.id, account, "left", value.revision, now); err != nil {
				return 0, err
			}
		}
		disbanded++
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return disbanded, nil
}

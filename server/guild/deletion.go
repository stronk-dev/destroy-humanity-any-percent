package guild

import (
	"context"
	"database/sql"
	"sort"
	"time"
)

func (service *Service) PrepareAccountDeletion(ctx context.Context, tx *sql.Tx, accountID string, now time.Time) error {
	if service == nil || tx == nil || !uuidPattern.MatchString(accountID) || now.IsZero() {
		return ErrInvalidIntent
	}
	rows, err := tx.QueryContext(ctx, `SELECT guild_id FROM guild_members WHERE account_id=$1 AND left_at IS NULL ORDER BY guild_id`, accountID)
	if err != nil {
		return err
	}
	var guildIDs []string
	for rows.Next() {
		var guildID string
		if err := rows.Scan(&guildID); err != nil {
			rows.Close()
			return err
		}
		guildIDs = append(guildIDs, guildID)
	}
	if err := rows.Close(); err != nil {
		return err
	}
	sort.Strings(guildIDs)
	for _, guildID := range guildIDs {
		revision, _, err := lockGuild(ctx, tx, guildID)
		if err != nil {
			return err
		}
		var membershipID, role string
		if err := tx.QueryRowContext(ctx, `SELECT membership_id,role FROM guild_members WHERE guild_id=$1 AND account_id=$2 AND left_at IS NULL FOR UPDATE`, guildID, accountID).Scan(&membershipID, &role); err != nil {
			return err
		}
		revision++
		if _, err := tx.ExecContext(ctx, `UPDATE guild_members SET left_at=$2,account_id=NULL WHERE membership_id=$1`, membershipID, now); err != nil {
			return err
		}
		if err := insertGuildEvent(ctx, tx, guildID, revision, "member_left", accountID, accountID, "", map[string]any{"reason": "account_deleted"}); err != nil {
			return err
		}
		if role == "leader" {
			var successor string
			err := tx.QueryRowContext(ctx, `SELECT account_id FROM guild_members WHERE guild_id=$1 AND left_at IS NULL
				ORDER BY CASE role WHEN 'officer' THEN 0 ELSE 1 END,joined_at,membership_id LIMIT 1 FOR UPDATE`, guildID).Scan(&successor)
			if err == nil {
				if _, err := tx.ExecContext(ctx, `UPDATE guild_members SET role='leader' WHERE guild_id=$1 AND account_id=$2 AND left_at IS NULL`, guildID, successor); err != nil {
					return err
				}
				if err := insertGuildEvent(ctx, tx, guildID, revision, "role_changed", accountID, successor, "", map[string]any{"role": "leader", "reason": "account_deleted"}); err != nil {
					return err
				}
			} else if err == sql.ErrNoRows {
				if _, err := tx.ExecContext(ctx, `UPDATE guilds SET disbanded_at=$2 WHERE guild_id=$1`, guildID, now); err != nil {
					return err
				}
				if err := insertGuildEvent(ctx, tx, guildID, revision, "guild_disbanded", accountID, accountID, "", map[string]any{"reason": "account_deleted"}); err != nil {
					return err
				}
			} else {
				return err
			}
		}
		if err := setGuildRevision(ctx, tx, guildID, revision); err != nil {
			return err
		}
		if err := service.refreshFloorState(ctx, tx, guildID, now); err != nil {
			return err
		}
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM guild_applications WHERE account_id=$1`, accountID); err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `DELETE FROM guild_invitations WHERE account_id=$1`, accountID)
	return err
}

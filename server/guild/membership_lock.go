package guild

import (
	"context"
	"database/sql"
)

// Lock before account revisions, Guilds or memberships. Admissions may affect
// another account, so the actor's revision lock alone cannot protect deletion's
// membership census. Sort and deduplicate the actual advisory keys (including
// possible hash collisions), rather than relying on UUID ordering.
func lockMembershipAccounts(ctx context.Context, tx *sql.Tx, accounts ...string) error {
	var nonempty []string
	for _, account := range accounts {
		if account != "" {
			nonempty = append(nonempty, account)
		}
	}
	rows, err := tx.QueryContext(ctx, `SELECT DISTINCT
		hashtextextended('cloud-clicker.guild-membership:' || account_id::uuid::text,0) AS lock_key
		FROM unnest($1::text[]) AS ids(account_id) ORDER BY lock_key`, nonempty)
	if err != nil {
		return err
	}
	var keys []int64
	for rows.Next() {
		var key int64
		if err := rows.Scan(&key); err != nil {
			rows.Close()
			return err
		}
		keys = append(keys, key)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return err
	}
	if err := rows.Close(); err != nil {
		return err
	}
	for _, key := range keys {
		if _, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock($1)`, key); err != nil {
			return err
		}
	}
	return nil
}

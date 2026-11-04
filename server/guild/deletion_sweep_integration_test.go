package guild

import (
	"context"
	"testing"
	"time"
)

func TestGuildAccountDeletionOverlappingSweepIntegration(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	db := guildIntegrationDatabase(t, ctx)
	now := time.Date(2026, 11, 4, 12, 0, 0, 0, time.UTC)
	account := "018f0000-0000-4000-8000-000000000911"
	if _, err := db.ExecContext(ctx, `INSERT INTO accounts(account_id,recovery_hash,created_at) VALUES($1,'test',$2)`, account, now); err != nil {
		t.Fatal(err)
	}
	catalog, err := LoadCatalog([]byte(phase0Catalog))
	if err != nil {
		t.Fatal(err)
	}
	service, err := NewService(db, catalog, integrationNames{}, func() time.Time { return now })
	if err != nil {
		t.Fatal(err)
	}
	created, err := service.Handle(ctx, account, []byte(guildIntent("018f0000-0000-7000-8000-000000000911", "create_guild", 1, `,"name":"Small Systems Sweep","join_policy":"open"`)))
	if err != nil || receiptOutcome(t, created.Receipt) != "applied" {
		t.Fatalf("create=%s err=%v", created.Receipt, err)
	}
	guildID := receiptGuildID(t, created.Receipt)
	due := now.Add(7 * 24 * time.Hour)
	sweeping, sweepName := namedGuildDatabase(t, ctx, "sweep")
	deleting, deletionName := namedGuildDatabase(t, ctx, "delete")
	sweepService, err := NewService(sweeping, catalog, integrationNames{}, service.clock)
	if err != nil {
		t.Fatal(err)
	}
	blocker, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer blocker.Rollback()
	if _, err := blocker.ExecContext(ctx, `SELECT membership_id FROM guild_members WHERE guild_id=$1 AND account_id=$2 AND left_at IS NULL FOR UPDATE`, guildID, account); err != nil {
		t.Fatal(err)
	}
	type sweepResult struct {
		count int
		err   error
	}
	swept := make(chan sweepResult, 1)
	go func() {
		count, err := sweepService.SweepDisbanded(ctx, due, 10)
		swept <- sweepResult{count, err}
	}()
	// Sweep owns the Guild row and is waiting on the membership row. Deletion
	// therefore observes that active membership before waiting for the Guild.
	waitGuildDatabaseBlocked(t, ctx, db, sweepName)
	deleted := make(chan error, 1)
	go func() {
		tx, err := deleting.BeginTx(ctx, nil)
		if err == nil {
			defer tx.Rollback()
			err = service.PrepareAccountDeletion(ctx, tx, account, due)
			if err == nil {
				_, err = tx.ExecContext(ctx, `DELETE FROM accounts WHERE account_id=$1`, account)
			}
			if err == nil {
				err = tx.Commit()
			}
		}
		deleted <- err
	}()
	waitGuildDatabaseBlocked(t, ctx, db, deletionName)
	if err := blocker.Commit(); err != nil {
		t.Fatal(err)
	}
	select {
	case observed := <-swept:
		if observed.err != nil || observed.count != 1 || ctx.Err() != nil {
			t.Fatalf("sweep=%d err=%v", observed.count, observed.err)
		}
	case <-ctx.Done():
		t.Fatal("sweep did not finish within guard")
	}
	select {
	case err := <-deleted:
		if err != nil || ctx.Err() != nil {
			t.Fatalf("account deletion after completed sweep: %v", err)
		}
	case <-ctx.Done():
		t.Fatal("deletion did not finish within guard")
	}
	var accounts, closedAnonymous, departures, disbands, presences int
	var disbanded bool
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM accounts WHERE account_id=$1`, account).Scan(&accounts); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM guild_members WHERE guild_id=$1 AND account_id IS NULL AND left_at=$2`, guildID, due).Scan(&closedAnonymous); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT count(*) FILTER (WHERE kind='member_left'),count(*) FILTER (WHERE kind='guild_disbanded') FROM guild_events WHERE guild_id=$1`, guildID).Scan(&departures, &disbands); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM guild_presence_outbox WHERE guild_id=$1 AND kind='left'`, guildID).Scan(&presences); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT disbanded_at=$2 FROM guilds WHERE guild_id=$1`, guildID, due).Scan(&disbanded); err != nil {
		t.Fatal(err)
	}
	if accounts != 0 || closedAnonymous != 1 || departures != 1 || disbands != 1 || presences != 1 || !disbanded {
		t.Fatalf("post-sweep deletion: accounts=%d closed=%d leaves=%d disbands=%d presences=%d disbanded=%v", accounts, closedAnonymous, departures, disbands, presences, disbanded)
	}
}

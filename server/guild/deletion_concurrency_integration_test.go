package guild

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"net/url"
	"os"
	"testing"
	"time"

	"cloud-clicker/server/save"

	"github.com/jackc/pgx/v5/pgconn"
)

func TestGuildAccountDeletionOverlappingAdmissionIntegration(t *testing.T) {
	for _, kind := range []string{"join_guild", "admit_member"} {
		t.Run(kind, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			db := guildIntegrationDatabase(t, ctx)
			now := time.Date(2026, 11, 4, 12, 0, 0, 0, time.UTC)
			leader := "018f0000-0000-4000-8000-000000000901"
			target := "018f0000-0000-4000-8000-000000000902"
			for _, account := range []string{leader, target} {
				if _, err := db.ExecContext(ctx, `INSERT INTO accounts(account_id,recovery_hash,created_at) VALUES($1,'test',$2)`, account, now); err != nil {
					t.Fatal(err)
				}
			}
			catalog, err := LoadCatalog([]byte(phase0Catalog))
			if err != nil {
				t.Fatal(err)
			}
			service, err := NewService(db, catalog, integrationNames{}, func() time.Time { return now })
			if err != nil {
				t.Fatal(err)
			}
			policy := "open"
			if kind == "admit_member" {
				policy = "apply"
			}
			created, err := service.Handle(ctx, leader, []byte(guildIntent("018f0000-0000-7000-8000-000000000901", "create_guild", 1, `,"name":"Small Systems","join_policy":"`+policy+`"`)))
			if err != nil || receiptOutcome(t, created.Receipt) != "applied" {
				t.Fatalf("create=%s err=%v", created.Receipt, err)
			}
			guildID := receiptGuildID(t, created.Receipt)
			actor, revision, extra := target, 1, `,"guild_id":"`+guildID+`"`
			if kind == "admit_member" {
				applied, err := service.Handle(ctx, target, []byte(guildIntent("018f0000-0000-7000-8000-000000000902", "apply_guild", 1, extra)))
				if err != nil || receiptOutcome(t, applied.Receipt) != "applied" {
					t.Fatalf("apply=%s err=%v", applied.Receipt, err)
				}
				actor, revision, extra = leader, 2, `,"account_id":"`+target+`"`
			}
			admitter, admissionName := namedGuildDatabase(t, ctx, "admit")
			deleting, deletionName := namedGuildDatabase(t, ctx, "delete")
			admissionService, err := NewService(admitter, catalog, integrationNames{}, service.clock)
			if err != nil {
				t.Fatal(err)
			}
			blocker, err := db.BeginTx(ctx, nil)
			if err != nil {
				t.Fatal(err)
			}
			defer blocker.Rollback()
			// Stop the real admission after it has started, before membership
			// insertion. Admission of another account must coordinate too.
			if kind == "join_guild" {
				_, err = blocker.ExecContext(ctx, `SELECT guild_id FROM guilds WHERE guild_id=$1 FOR UPDATE`, guildID)
			} else {
				_, err = blocker.ExecContext(ctx, `SELECT guild_id FROM guild_applications WHERE guild_id=$1 AND account_id=$2 FOR UPDATE`, guildID, target)
			}
			if err != nil {
				t.Fatal(err)
			}
			type admissionResult struct {
				result HandleResult
				err    error
			}
			admitted := make(chan admissionResult, 1)
			go func() {
				result, err := admissionService.Handle(ctx, actor, []byte(guildIntent("018f0000-0000-7000-8000-000000000903", kind, revision, extra)))
				admitted <- admissionResult{result, err}
			}()
			waitGuildDatabaseBlocked(t, ctx, db, admissionName)
			deleted := make(chan error, 1)
			go func() {
				tx, err := deleting.BeginTx(ctx, nil)
				if err == nil {
					defer tx.Rollback()
					err = service.PrepareAccountDeletion(ctx, tx, target, now.Add(time.Minute))
					if err == nil {
						_, err = tx.ExecContext(ctx, `DELETE FROM accounts WHERE account_id=$1`, target)
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
			case observed := <-admitted:
				if observed.err != nil || ctx.Err() != nil || receiptOutcome(t, observed.result.Receipt) != "applied" {
					t.Fatalf("admission=%s err=%v", observed.result.Receipt, observed.err)
				}
			case <-ctx.Done():
				t.Fatal("admission did not finish within guard")
			}
			select {
			case err := <-deleted:
				if err != nil || ctx.Err() != nil {
					t.Fatalf("deletion did not finish within guard: %v", err)
				}
			case <-ctx.Done():
				t.Fatal("deletion did not finish within guard")
			}
			var activeAnonymous, closedAnonymous, activeLeader, accounts int
			if err := db.QueryRowContext(ctx, `SELECT
				count(*) FILTER (WHERE account_id IS NULL AND left_at IS NULL),
				count(*) FILTER (WHERE account_id IS NULL AND left_at IS NOT NULL),
				count(*) FILTER (WHERE account_id=$2 AND role='leader' AND left_at IS NULL)
				FROM guild_members WHERE guild_id=$1`, guildID, leader).Scan(&activeAnonymous, &closedAnonymous, &activeLeader); err != nil {
				t.Fatal(err)
			}
			if err := db.QueryRowContext(ctx, `SELECT count(*) FROM accounts WHERE account_id=$1`, target).Scan(&accounts); err != nil {
				t.Fatal(err)
			}
			if activeAnonymous != 0 || closedAnonymous != 1 || activeLeader != 1 || accounts != 0 {
				t.Fatalf("deleted account escaped Guild cleanup: active anonymous=%d closed anonymous=%d active leader=%d accounts=%d", activeAnonymous, closedAnonymous, activeLeader, accounts)
			}
			var departures, pending int
			if err := db.QueryRowContext(ctx, `SELECT count(*) FROM guild_events WHERE guild_id=$1 AND kind='member_left' AND payload->>'reason'='account_deleted'`, guildID).Scan(&departures); err != nil {
				t.Fatal(err)
			}
			if err := db.QueryRowContext(ctx, `SELECT count(*) FROM guild_applications WHERE account_id=$1`, target).Scan(&pending); err != nil {
				t.Fatal(err)
			}
			if departures != 1 || pending != 0 {
				t.Fatalf("deletion events/requests: departures=%d applications=%d", departures, pending)
			}

			// Reverse the order: deletion has completed its census but has
			// not removed the account yet. A new request must wait until the
			// deletion transaction ends, then refuse rather than resurrect it.
			later := "018f0000-0000-4000-8000-000000000904"
			if _, err := db.ExecContext(ctx, `INSERT INTO accounts(account_id,recovery_hash,created_at) VALUES($1,'test',$2)`, later, now); err != nil {
				t.Fatal(err)
			}
			actor, revision, extra = later, 1, `,"guild_id":"`+guildID+`"`
			if kind == "admit_member" {
				applied, err := service.Handle(ctx, later, []byte(guildIntent("018f0000-0000-7000-8000-000000000904", "apply_guild", 1, extra)))
				if err != nil || receiptOutcome(t, applied.Receipt) != "applied" {
					t.Fatalf("later apply=%s err=%v", applied.Receipt, err)
				}
				actor, revision, extra = leader, 3, `,"account_id":"`+later+`"`
			}
			deletion, err := deleting.BeginTx(ctx, nil)
			if err != nil {
				t.Fatal(err)
			}
			defer deletion.Rollback()
			if err := service.PrepareAccountDeletion(ctx, deletion, later, now.Add(time.Minute)); err != nil {
				t.Fatal(err)
			}
			// The mutex must not take an account parent-row lock that would
			// invert existing Guild -> FK-reference ordering.
			keyShare, err := db.BeginTx(ctx, nil)
			if err != nil {
				t.Fatal(err)
			}
			defer keyShare.Rollback()
			if _, err := keyShare.ExecContext(ctx, `SELECT account_id FROM accounts WHERE account_id=$1 FOR KEY SHARE NOWAIT`, later); err != nil {
				t.Fatalf("membership mutex blocks foreign-key access: %v", err)
			}
			if err := keyShare.Rollback(); err != nil {
				t.Fatal(err)
			}
			go func() {
				result, err := admissionService.Handle(ctx, actor, []byte(guildIntent("018f0000-0000-7000-8000-000000000905", kind, revision, extra)))
				admitted <- admissionResult{result, err}
			}()
			waitGuildDatabaseBlocked(t, ctx, db, admissionName)
			if _, err := deletion.ExecContext(ctx, `DELETE FROM accounts WHERE account_id=$1`, later); err != nil {
				t.Fatal(err)
			}
			if err := deletion.Commit(); err != nil {
				t.Fatal(err)
			}
			select {
			case observed := <-admitted:
				if ctx.Err() != nil {
					t.Fatal("post-deletion admission exceeded guard")
				}
				if kind == "join_guild" {
					var postgres *pgconn.PgError
					if !errors.As(observed.err, &postgres) || postgres.Code != "23503" {
						t.Fatalf("deleted actor must fail the account FK: %s err=%v", observed.result.Receipt, observed.err)
					}
				} else if observed.err != nil || receiptOutcome(t, observed.result.Receipt) != "rejected" || receiptCategory(t, observed.result.Receipt) != "not_applicant" {
					t.Fatalf("deleted target must refuse as not_applicant: %s err=%v", observed.result.Receipt, observed.err)
				}
			case <-ctx.Done():
				t.Fatal("post-deletion admission did not finish within guard")
			}
		})
	}
}

func namedGuildDatabase(t *testing.T, ctx context.Context, role string) (*sql.DB, string) {
	t.Helper()
	parsed, err := url.Parse(os.Getenv("TEST_DATABASE_URL"))
	if err != nil {
		t.Fatal(err)
	}
	application := "guild_deletion_" + role + "_" + rand.Text()
	query := parsed.Query()
	query.Set("application_name", application)
	parsed.RawQuery = query.Encode()
	db, err := save.OpenPostgres(ctx, parsed.String())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db, application
}

func waitGuildDatabaseBlocked(t *testing.T, ctx context.Context, db *sql.DB, application string) {
	t.Helper()
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	for {
		var blocked int
		if err := db.QueryRowContext(ctx, `SELECT count(*) FROM pg_stat_activity
			WHERE datname=current_database() AND application_name=$1 AND state='active'
			AND wait_event_type='Lock' AND cardinality(pg_blocking_pids(pid))>0`, application).Scan(&blocked); err != nil {
			t.Fatal(err)
		}
		if ctx.Err() != nil {
			t.Fatal("database handshake exceeded observation guard")
		}
		if blocked == 1 {
			return
		}
		select {
		case <-ctx.Done():
			t.Fatal("operation never reached observed database lock")
		case <-ticker.C:
		}
	}
}

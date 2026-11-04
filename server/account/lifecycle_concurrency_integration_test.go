package account

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"errors"
	"net/url"
	"os"
	"testing"
	"time"

	"cloud-clicker/server/save"
)

func TestAccountDeletionSerializesLifecycleIntegration(t *testing.T) {
	for _, operation := range []string{"create_session", "refresh_session", "new_founder", "import_founder"} {
		t.Run(operation, func(t *testing.T) {
			now := time.Date(2026, 11, 4, 0, 0, 0, 0, time.UTC)
			base, database, _ := bootstrapRepository(t, &now)
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			created, err := base.CreateAccount(ctx)
			if err != nil {
				t.Fatal(err)
			}
			initial, err := base.CreateSession(ctx, created.AccountID, created.RecoveryCode)
			if err != nil {
				t.Fatal(err)
			}
			deletion, deletionName := namedLifecycleRepository(t, ctx, base, "delete")
			rival, rivalName := namedLifecycleRepository(t, ctx, base, "rival")
			state, err := base.ActiveCompanyState(ctx, created.AccountID)
			if err != nil {
				t.Fatal(err)
			}
			var imported json.RawMessage = append([]byte(nil), state.State...)
			blocker, err := database.BeginTx(ctx, nil)
			if err != nil {
				t.Fatal(err)
			}
			defer blocker.Rollback()
			deleteFirst := operation == "create_session" || operation == "refresh_session"
			if deleteFirst {
				_, err = blocker.ExecContext(ctx, `SELECT founder_id FROM account_founders WHERE founder_id=$1 FOR UPDATE`, created.FounderID)
			} else {
				_, err = blocker.ExecContext(ctx, `SELECT id FROM save_streams WHERE owner_kind='founder' AND owner_id=$1 AND scope='company' FOR UPDATE`, created.FounderID)
			}
			if err != nil {
				t.Fatal(err)
			}
			type result struct {
				pair    TokenPair
				founder Founder
				err     error
			}
			deleted, competed := make(chan result, 1), make(chan result, 1)
			deleteAccount := func() { deleted <- result{err: deletion.DeleteAccount(ctx, created.AccountID)} }
			compete := func() {
				var observed result
				switch operation {
				case "create_session":
					observed.pair, observed.err = rival.CreateSession(ctx, created.AccountID, created.RecoveryCode)
				case "refresh_session":
					observed.pair, observed.err = rival.RefreshSession(ctx, initial.RefreshToken)
				case "new_founder":
					// PostgreSQL accepts this UUID spelling for the same account;
					// it must not select another lifecycle mutex.
					observed.founder, observed.err = rival.NewFounder(ctx, "{"+created.AccountID+"}")
				case "import_founder":
					observed.founder, observed.err = rival.ImportFounder(ctx, created.AccountID, state.ConstantsHash, state.Version, imported)
				}
				competed <- observed
			}
			if deleteFirst {
				go deleteAccount()
				waitLifecycleBlocked(t, ctx, database, deletionName)
				go compete()
				waitLifecycleBlocked(t, ctx, database, rivalName)
			} else {
				go compete()
				waitLifecycleBlocked(t, ctx, database, rivalName)
				go deleteAccount()
				waitLifecycleBlocked(t, ctx, database, deletionName)
			}
			if err := blocker.Commit(); err != nil {
				t.Fatal(err)
			}
			await := func(done <-chan result) result {
				t.Helper()
				select {
				case observed := <-done:
					if ctx.Err() != nil {
						t.Fatal("lifecycle completed after the observation guard")
					}
					return observed
				case <-ctx.Done():
					t.Fatal("concurrent lifecycle did not reach its outcome")
					return result{}
				}
			}
			deleteResult, other := await(deleted), await(competed)
			if deleteResult.err != nil {
				t.Fatalf("ordered account deletion failed: %v", deleteResult.err)
			}
			if deleteFirst {
				if !errors.Is(other.err, ErrAuthentication) || other.pair != (TokenPair{}) {
					t.Fatalf("deletion did not precede credential issuance: %v", other.err)
				}
			} else if other.err != nil || other.founder.ID == "" || other.founder.ID == created.FounderID || other.founder.Imported != (operation == "import_founder") {
				t.Fatalf("Founder transition did not precede deletion: %v", other.err)
			}
			var accounts, linked, liveStreams, founders, streams, credentials int
			if err := database.QueryRowContext(ctx, `SELECT
				(SELECT count(*) FROM accounts),
				(SELECT count(*) FROM account_founders WHERE account_id IS NOT NULL OR archived_at IS NULL),
				(SELECT count(*) FROM save_streams WHERE archived_at IS NULL),
				(SELECT count(*) FROM account_founders),
				(SELECT count(*) FROM save_streams),
				(SELECT count(*) FROM session_families)+(SELECT count(*) FROM sessions)+(SELECT count(*) FROM access_tokens)`).Scan(
				&accounts, &linked, &liveStreams, &founders, &streams, &credentials); err != nil {
				t.Fatal(err)
			}
			wantFounders := 1
			if !deleteFirst {
				wantFounders = 2
			}
			if accounts != 0 || linked != 0 || liveStreams != 0 || credentials != 0 || founders != wantFounders || streams != 2*wantFounders {
				t.Fatalf("incomplete account deletion: accounts=%d linked=%d live_streams=%d credentials=%d founders=%d streams=%d", accounts, linked, liveStreams, credentials, founders, streams)
			}
			if _, err := base.Authenticate(ctx, initial.AccessToken); !errors.Is(err, ErrAuthentication) {
				t.Fatal("deleted account retained access")
			}
			if _, err := base.RefreshSession(ctx, initial.RefreshToken); !errors.Is(err, ErrAuthentication) {
				t.Fatal("deleted account retained refresh")
			}
		})
	}
}

func TestAccountLifecyclePreservesForeignKeyAccessIntegration(t *testing.T) {
	now := time.Date(2026, 11, 4, 0, 0, 0, 0, time.UTC)
	base, database, _ := bootstrapRepository(t, &now)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	first, err := base.CreateAccount(ctx)
	if err != nil {
		t.Fatal(err)
	}
	second, err := base.CreateAccount(ctx)
	if err != nil {
		t.Fatal(err)
	}
	writer, name := namedLifecycleRepository(t, ctx, base, "writer")
	blocker, err := database.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer blocker.Rollback()
	if _, err := blocker.ExecContext(ctx, `SELECT id FROM save_streams WHERE owner_kind='founder' AND owner_id=$1 AND scope='company' FOR UPDATE`, first.FounderID); err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { _, err := writer.NewFounder(ctx, first.AccountID); done <- err }()
	waitLifecycleBlocked(t, ctx, database, name)
	foreignKey, err := database.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer foreignKey.Rollback()
	var accountID string
	if err := foreignKey.QueryRowContext(ctx, `SELECT account_id FROM accounts WHERE account_id=$1 FOR KEY SHARE NOWAIT`, first.AccountID).Scan(&accountID); err != nil || accountID != first.AccountID {
		t.Fatalf("lifecycle mutex blocks existing foreign-key access: %v", err)
	}
	if err := foreignKey.Rollback(); err != nil {
		t.Fatal(err)
	}
	if next, err := base.NewFounder(ctx, second.AccountID); err != nil || next.ID == second.FounderID {
		t.Fatalf("one account's lifecycle blocked another account: %v", err)
	}
	if err := blocker.Commit(); err != nil {
		t.Fatal(err)
	}
	select {
	case err := <-done:
		if err != nil || ctx.Err() != nil {
			t.Fatalf("first account did not finish within guard: %v", err)
		}
	case <-ctx.Done():
		t.Fatal("first account did not finish")
	}
}

func namedLifecycleRepository(t *testing.T, ctx context.Context, base *Repository, name string) (*Repository, string) {
	t.Helper()
	parsed, err := url.Parse(os.Getenv("TEST_DATABASE_URL"))
	if err != nil {
		t.Fatal(err)
	}
	application := "account_lifecycle_" + name + "_" + rand.Text()
	query := parsed.Query()
	query.Set("application_name", application)
	parsed.RawQuery = query.Encode()
	database, err := save.OpenPostgres(ctx, parsed.String())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close() })
	repository, err := NewRepository(database, base.catalogs, base.constantsHash, base.keys, base.clock, nil)
	if err != nil {
		t.Fatal(err)
	}
	return repository, application
}

func waitLifecycleBlocked(t *testing.T, ctx context.Context, database *sql.DB, application string) {
	t.Helper()
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	for {
		var blocked int
		if err := database.QueryRowContext(ctx, `SELECT count(*) FROM pg_stat_activity
			WHERE datname=current_database() AND application_name=$1 AND state='active'
			AND wait_event_type='Lock' AND cardinality(pg_blocking_pids(pid))>0`, application).Scan(&blocked); err != nil {
			t.Fatal(err)
		}
		if ctx.Err() != nil {
			t.Fatal("lock handshake crossed observation guard")
		}
		if blocked == 1 {
			return
		}
		select {
		case <-ctx.Done():
			t.Fatal("operation never reached its observed database lock")
		case <-ticker.C:
		}
	}
}

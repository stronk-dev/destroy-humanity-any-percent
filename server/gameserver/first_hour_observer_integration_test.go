package gameserver

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"cloud-clicker/server/save"
)

func TestFirstHourVerificationObserverIntegration(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}
	ctx := context.Background()
	db, err := save.OpenPostgres(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := save.Migrate(ctx, db); err != nil {
		t.Fatal(err)
	}
	const clean = `TRUNCATE bootstrap_receipts,accounts,save_streams,catalog_sets,epochs RESTART IDENTITY CASCADE`
	if _, err := db.ExecContext(ctx, clean); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if _, err := db.ExecContext(context.Background(), clean); err != nil {
			t.Errorf("clean observer fixture: %v", err)
		}
	})
	const stream = "0198cccc-0000-7000-8000-000000000001"
	// These rows exercise the observer's SQL predicate only. The composed
	// first-hour journey remains the proof of real replay/archive production.
	for _, statement := range []string{
		`INSERT INTO save_streams(id,owner_kind,owner_id,scope) VALUES('` + stream + `','founder','0198cccc-0000-7000-8000-000000000002','company')`,
		`INSERT INTO catalog_sets(constants_hash) VALUES('` + testConstantsHash + `')`,
		`INSERT INTO epochs(name,started_at,changelog_ref) VALUES('observer fixture',now(),'changelog/epoch-1.md')`,
		`INSERT INTO epoch_hashes(epoch_id,constants_hash) SELECT epoch_id,'` + testConstantsHash + `' FROM epochs`,
	} {
		if _, err := db.ExecContext(ctx, statement); err != nil {
			t.Fatal(err)
		}
	}

	for index, test := range []struct {
		name, status, verdict, lastError, wantError string
		archive, deadline                           bool
	}{
		{name: "complete", status: "verified", verdict: "verified", archive: true},
		{name: "claimed without producer", status: "claimed", deadline: true},
		{name: "pending retry is not terminal", status: "pending", lastError: "retryable catalog read", deadline: true},
		{name: "version deferral is not terminal", status: "pending", lastError: "engine version unavailable", deadline: true},
		{name: "archive alone is incomplete", status: "claimed", archive: true, deadline: true},
		{name: "verified without archive", status: "verified", verdict: "verified", wantError: "committed verified row"},
		{name: "verified wrong verdict", status: "verified", verdict: "state_divergence", archive: true, wantError: "committed verified row"},
		{name: "deterministic refusal", status: "dead", verdict: "state_divergence", wantError: "dead-letter/poison"},
		{name: "poison refusal", status: "dead", lastError: "exhausted attempts", wantError: "dead-letter/poison"},
	} {
		t.Run(test.name, func(t *testing.T) {
			run := int64(index + 1)
			tx, err := db.BeginTx(ctx, nil)
			if err != nil {
				t.Fatal(err)
			}
			defer tx.Rollback()
			if _, err := tx.ExecContext(ctx, `INSERT INTO run_epochs(company_stream_id,run_seq,epoch_id,constants_hash,engine_version,build_vcs_hash,seed)
				SELECT $1,$2,epoch_id,$3,'1.0.0','observer fixture','1' FROM epochs`, stream, run, testConstantsHash); err != nil {
				t.Fatal(err)
			}
			if _, err := tx.ExecContext(ctx, `INSERT INTO run_genesis(company_stream_id,run_seq,state,version,constants_hash)
				VALUES($1,$2,'{}',1,$3)`, stream, run, testConstantsHash); err != nil {
				t.Fatal(err)
			}
			if _, err := tx.ExecContext(ctx, `INSERT INTO verification_queue(company_stream_id,run_seq,status,verdict,attempts,last_error)
				VALUES($1,$2,$3,NULLIF($4,''),1,NULLIF($5,''))`, stream, run, test.status, test.verdict, test.lastError); err != nil {
				t.Fatal(err)
			}
			if test.archive {
				if _, err := tx.ExecContext(ctx, `INSERT INTO run_log_archive(run_id,company_stream_id,run_seq,terminal_seq,encoding,bytes,sha256)
					VALUES($1,$2,$3,1,'gzip+json.v1',$4,$5)`, fmt.Sprintf("%s:%d", stream, run), stream, run, []byte("observer fixture"), testConstantsHash); err != nil {
					t.Fatal(err)
				}
			}
			if err := tx.Commit(); err != nil {
				t.Fatal(err)
			}
			bound := time.Second
			if test.deadline {
				bound = 100 * time.Millisecond
			}
			caller, cancel := context.WithTimeout(ctx, bound)
			defer cancel()
			observed, err := observeFirstHourVerification(caller, db, nil, stream, run)
			if observed.status != test.status || observed.verdict != test.verdict || observed.lastError != test.lastError || observed.attempts != 1 {
				t.Fatalf("lost queue diagnostics: %+v", observed)
			}
			if test.deadline {
				if !errors.Is(err, context.DeadlineExceeded) {
					t.Fatalf("incomplete producer must reach caller bound: %v", err)
				}
			} else if test.wantError != "" {
				if err == nil || errors.Is(err, context.DeadlineExceeded) || !strings.Contains(err.Error(), test.wantError) {
					t.Fatalf("required immediate refusal %q: %v", test.wantError, err)
				}
			} else if err != nil || observed.archiveHash != testConstantsHash || observed.terminalSeq != 1 {
				t.Fatalf("durable-success predicate: %+v err=%v", observed, err)
			}
		})
	}
	t.Run("absent run cannot borrow another archive", func(t *testing.T) {
		caller, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
		defer cancel()
		observed, err := observeFirstHourVerification(caller, db, nil, stream, 100)
		if !errors.Is(err, context.DeadlineExceeded) || observed.status != "" || observed.archiveHash != "" {
			t.Fatalf("absent run passed or borrowed state: %+v err=%v", observed, err)
		}
	})

	for _, mode := range []string{"caller cancellation", "worker failure", "original guard"} {
		t.Run("blocked SQL/"+mode, func(t *testing.T) {
			lock, err := db.BeginTx(ctx, nil)
			if err != nil {
				t.Fatal(err)
			}
			defer lock.Rollback()
			if _, err := lock.ExecContext(ctx, `LOCK TABLE verification_queue IN ACCESS EXCLUSIVE MODE`); err != nil {
				t.Fatal(err)
			}
			caller, cancel := context.WithCancel(ctx)
			defer cancel()
			failures := make(chan error, 1)
			workerErr := errors.New("observer control: verification worker stopped")
			result := make(chan error, 1)
			go func() {
				_, err := observeFirstHourVerification(caller, db, failures, stream, 1)
				result <- err
			}()
			// Establish the real blocked query before injecting cancellation. No
			// sleep-only assumption that the observer actually crossed the DB.
			waitFirstHourObserverSQLLock(t, db)
			want := error(context.DeadlineExceeded)
			containment := 7 * time.Second
			switch mode {
			case "caller cancellation":
				want = context.Canceled
				cancel()
				containment = time.Second
			case "worker failure":
				want = workerErr
				failures <- workerErr
				containment = time.Second
			}
			select {
			case err := <-result:
				if !errors.Is(err, want) {
					t.Fatalf("blocked query returned %v, want %v", err, want)
				}
				// The original exclusive lock is still held on this connection.
			case <-time.After(containment):
				_ = lock.Rollback() // contain broken controls without leaving SQL running
				cancel()
				select {
				case <-result:
				case <-time.After(time.Second):
					t.Error("observer did not exit after lock release and cancellation")
				}
				t.Fatal("observer did not return with original exclusive lock held")
			}
		})
	}
}

func waitFirstHourObserverSQLLock(t *testing.T, db *sql.DB) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	for {
		var blocked bool
		if err := db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM pg_stat_activity
			WHERE datname=current_database() AND wait_event_type='Lock'
			AND query LIKE '%FROM verification_queue queue LEFT JOIN run_log_archive%')`).Scan(&blocked); err != nil {
			t.Fatal(err)
		}
		if blocked {
			return
		}
		select {
		case <-ctx.Done():
			t.Fatal("observer never reached the real SQL lock wait")
		case <-time.After(10 * time.Millisecond):
		}
	}
}

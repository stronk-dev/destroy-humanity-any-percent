package gameserver

import (
	"context"
	"database/sql"
	"errors"
	"io"
	"testing"
	"time"
)

func TestComposedDrainBoundsBlockedPlayerRelayIntegration(t *testing.T) {
	for _, callerBound := range []bool{true, false} {
		name := "configured_bound"
		if callerBound {
			name = "shorter_caller_bound"
		}
		t.Run(name, func(t *testing.T) { testComposedDrainBoundsBlockedPlayerRelay(t, callerBound) })
	}
}

func testComposedDrainBoundsBlockedPlayerRelay(t *testing.T, callerBound bool) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	database := newCredentialCleanupDatabase(t, ctx)
	defer database.Close()
	clock := &mutableClock{now: time.Now().UTC()}
	_, composition := newCredentialCleanupComposition(t, ctx, database, clock, io.Discard)
	serverCtx, stopServer := context.WithCancel(ctx)
	defer stopServer()
	if err := composition.Server.Start(serverCtx); err != nil {
		t.Fatal(err)
	}
	defer func() {
		stopServer()
		cleanup, stop := context.WithTimeout(context.Background(), 5*time.Second)
		defer stop()
		if err := composition.Node.Shutdown(cleanup); err != nil {
			t.Errorf("node cleanup: %v", err)
		}
	}()
	lock, err := database.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer lock.Rollback()
	if _, err := lock.ExecContext(ctx, "LOCK TABLE transport_player_outbox IN ACCESS EXCLUSIVE MODE"); err != nil {
		t.Fatal(err)
	}
	// Wait for the real background relay, not a fake Flush or a direct lock on
	// its application mutex. The table lock remains held through the outcome.
	const blockedRelay = `SELECT a.pid FROM pg_stat_activity a JOIN pg_locks l USING(pid)
		WHERE a.datname=current_database() AND l.relation='transport_player_outbox'::regclass
		AND NOT l.granted AND a.query LIKE '%UPDATE transport_player_outbox AS outbox%'`
	var relayPID int
	for {
		err := database.QueryRowContext(ctx, blockedRelay).Scan(&relayPID)
		if err == nil {
			break
		}
		if err != sql.ErrNoRows {
			t.Fatal(err)
		}
		select {
		case <-ctx.Done():
			t.Fatal("actual background relay never waited on the held table lock")
		case <-time.After(10 * time.Millisecond):
		}
	}
	drainCtx := ctx
	bound := composition.Node.DrainTimeout()
	if callerBound {
		bound = 2 * time.Second
		var stopDrain context.CancelFunc
		drainCtx, stopDrain = context.WithTimeout(ctx, bound)
		defer stopDrain()
	}
	drained := make(chan error, 1)
	go func() { drained <- composition.Server.Drain(drainCtx, clock.Time()) }()
	finished := false
	defer func() {
		if !finished {
			// Containment for the broken baseline: only after a failed assertion
			// may cleanup release the query. Join Drain before closing its DB.
			stopServer()
			_ = lock.Rollback()
			select {
			case <-drained:
			case <-time.After(5 * time.Second):
				t.Error("drain did not finish during fixture cleanup")
			}
		}
	}()
	select {
	case err := <-drained:
		finished = true
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("blocked drain must report its expired bound: %v", err)
		}
	case <-time.After(bound + time.Second):
		t.Fatal("drain remained behind the player relay after its deadline")
	}
	if composition.Server.ready.Load() {
		t.Fatal("timed-out drain restored readiness")
	}
	// Pgx may close an interrupted connection asynchronously after QueryContext
	// returns. Observe worker exit AND backend cancellation within the same
	// existing cleanup allowance; Drain is not an instantaneous SQL barrier.
	settled, stopSettled := context.WithTimeout(ctx, time.Second)
	defer stopSettled()
	select {
	case <-composition.Server.relayDone:
	case <-settled.Done():
		t.Fatal("background relay survived bounded drain")
	}
	for {
		var waiting bool
		if err := database.QueryRowContext(settled, "SELECT EXISTS(SELECT 1 FROM pg_locks WHERE pid=$1 AND NOT granted)", relayPID).Scan(&waiting); err != nil {
			t.Fatalf("observe relay cancellation with exclusive lock held: %v", err)
		}
		if !waiting {
			break
		}
		select {
		case <-settled.Done():
			t.Fatal("relay backend lock wait survived cleanup allowance")
		case <-time.After(10 * time.Millisecond):
		}
	}
}

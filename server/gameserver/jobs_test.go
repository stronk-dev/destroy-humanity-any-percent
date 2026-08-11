package gameserver

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"cloud-clicker/server/account"
)

type credentialPrunerFixture struct {
	sessionCalls   int
	bootstrapCalls int
	sessionErr     error
}

func (fixture *credentialPrunerFixture) PruneExpiredSessions(context.Context, time.Time, int) (account.SessionGCResult, error) {
	fixture.sessionCalls++
	return account.SessionGCResult{}, fixture.sessionErr
}
func (fixture *credentialPrunerFixture) PruneExpiredBootstrapReceipts(context.Context, time.Time, int) (int64, error) {
	fixture.bootstrapCalls++
	return 0, nil
}

func TestPeriodicJobPrimeAndScheduledTickUseSameCallback(t *testing.T) {
	var calls atomic.Int64
	job, err := NewPeriodicJob(time.Millisecond, func(context.Context) error {
		calls.Add(1)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := job.Prime(context.Background()); err != nil || calls.Load() != 1 {
		t.Fatalf("prime calls=%d err=%v", calls.Load(), err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- job.Run(ctx) }()
	deadline := time.Now().Add(time.Second)
	for calls.Load() < 2 && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	cancel()
	if err := <-done; err != nil || calls.Load() < 2 {
		t.Fatalf("scheduled calls=%d err=%v", calls.Load(), err)
	}
}

func TestCredentialGCAlwaysRunsSessionThenBootstrapReceiptPruning(t *testing.T) {
	fixture := &credentialPrunerFixture{}
	if err := pruneExpiredCredentials(context.Background(), fixture, time.Unix(1, 0), 1_000); err != nil || fixture.sessionCalls != 1 || fixture.bootstrapCalls != 1 {
		t.Fatalf("credential GC calls sessions=%d bootstrap=%d err=%v", fixture.sessionCalls, fixture.bootstrapCalls, err)
	}
	fixture = &credentialPrunerFixture{sessionErr: errors.New("transient")}
	if err := pruneExpiredCredentials(context.Background(), fixture, time.Unix(1, 0), 1_000); err == nil || fixture.sessionCalls != 1 || fixture.bootstrapCalls != 0 {
		t.Fatalf("credential GC did not stop on session error: sessions=%d bootstrap=%d err=%v", fixture.sessionCalls, fixture.bootstrapCalls, err)
	}
}

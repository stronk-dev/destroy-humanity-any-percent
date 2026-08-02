package gameserver

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

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

package transport

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"cloud-clicker/server/save"
)

type heldPlayerSource struct {
	memoryPlayerSource
	entered chan struct{}
	release chan struct{}
	calls   atomic.Int32
}

func (source *heldPlayerSource) ClaimPlayerOutbox(ctx context.Context, _ int, _ time.Duration) ([]save.PlayerOutboxItem, error) {
	if source.calls.Add(1) == 1 {
		close(source.entered)
		select {
		case <-source.release:
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	return nil, nil
}

func TestPlayerRelayFlushCancellationWhileContended(t *testing.T) {
	for _, deadline := range []bool{false, true} {
		name := "cancelled"
		if deadline {
			name = "deadline"
		}
		t.Run(name, func(t *testing.T) {
			source := &heldPlayerSource{entered: make(chan struct{}), release: make(chan struct{})}
			relay, err := NewPlayerRelay(source, &memoryPublisher{}, &memoryRelaySink{})
			if err != nil {
				t.Fatal(err)
			}
			owner := make(chan error, 1)
			waiter := make(chan error, 1)
			waiterStarted, waiterFinished := false, false
			ownerFinished, released := false, false
			go func() { _, err := relay.Flush(context.Background()); owner <- err }()
			t.Cleanup(func() {
				if !released {
					close(source.release)
				}
				if !ownerFinished {
					select {
					case err := <-owner:
						if err != nil {
							t.Errorf("unrelated owner flush: %v", err)
						}
					case <-time.After(time.Second):
						t.Error("owner flush did not finish after release")
					}
				}
				if waiterStarted && !waiterFinished {
					select {
					case <-waiter:
					case <-time.After(time.Second):
						t.Error("waiter did not finish after fixture release")
					}
				}
			})
			select {
			case <-source.entered:
			case <-time.After(time.Second):
				t.Fatal("first flush did not enter source")
			}
			ctx, cancel := context.WithCancel(context.Background())
			want := context.Canceled
			if deadline {
				cancel()
				ctx, cancel = context.WithTimeout(context.Background(), 50*time.Millisecond)
				want = context.DeadlineExceeded
			} else {
				cancel()
			}
			defer cancel()
			waiterStarted = true
			go func() { _, err := relay.Flush(ctx); waiter <- err }()
			select {
			case err := <-waiter:
				waiterFinished = true
				if !errors.Is(err, want) {
					t.Fatalf("contended flush: %v, want %v", err, want)
				}
			case <-time.After(time.Second):
				t.Fatal("cancelled flush remained behind another caller's source query")
			}
			if calls := source.calls.Load(); calls != 1 {
				t.Fatalf("cancelled waiter reached source: calls=%d", calls)
			}
			select {
			case err := <-owner:
				ownerFinished = true
				t.Fatalf("waiter cancelled the unrelated owner: %v", err)
			default:
			}
			close(source.release)
			released = true
			select {
			case err := <-owner:
				ownerFinished = true
				if err != nil {
					t.Fatal(err)
				}
			case <-time.After(time.Second):
				t.Fatal("owner did not finish after release")
			}
			// Cancellation must not consume the serializer's only slot.
			healthyCtx, stop := context.WithTimeout(context.Background(), time.Second)
			defer stop()
			if count, err := relay.Flush(healthyCtx); count != 0 || err != nil || source.calls.Load() != 2 {
				t.Fatalf("healthy flush after cancellation: count=%d err=%v calls=%d", count, err, source.calls.Load())
			}
		})
	}
}

func TestPlayerRelayFlushAlreadyCancelledDoesNotClaim(t *testing.T) {
	source := &heldPlayerSource{entered: make(chan struct{}), release: make(chan struct{})}
	close(source.release)
	relay, err := NewPlayerRelay(source, &memoryPublisher{}, &memoryRelaySink{})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if count, err := relay.Flush(ctx); count != 0 || !errors.Is(err, context.Canceled) || source.calls.Load() != 0 {
		t.Fatalf("cancelled flush claimed work: count=%d err=%v calls=%d", count, err, source.calls.Load())
	}
}

package main

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"
)

func TestWaitForStopOwnsSignalListenerAndWorkerLifecycles(t *testing.T) {
	t.Run("signal context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		listenerErr, runtimeErr := waitForStop(ctx, make(chan error), make(chan error))
		if listenerErr != nil || runtimeErr != nil {
			t.Fatalf("listener=%v runtime=%v", listenerErr, runtimeErr)
		}
	})
	t.Run("worker failure", func(t *testing.T) {
		injected := errors.New("clearing worker stopped")
		failures := make(chan error, 1)
		failures <- injected
		listenerErr, runtimeErr := waitForStop(context.Background(), make(chan error), failures)
		if listenerErr != nil || !errors.Is(runtimeErr, injected) {
			t.Fatalf("listener=%v runtime=%v", listenerErr, runtimeErr)
		}
	})
	t.Run("listener close", func(t *testing.T) {
		serveErr := make(chan error, 1)
		serveErr <- http.ErrServerClosed
		listenerErr, runtimeErr := waitForStop(context.Background(), serveErr, make(chan error))
		if listenerErr != nil || runtimeErr != nil {
			t.Fatalf("listener=%v runtime=%v", listenerErr, runtimeErr)
		}
	})
	t.Run("listener failure", func(t *testing.T) {
		injected := errors.New("listener failed")
		serveErr := make(chan error, 1)
		serveErr <- injected
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		listenerErr, runtimeErr := waitForStop(ctx, serveErr, make(chan error))
		if !errors.Is(listenerErr, injected) || runtimeErr != nil {
			t.Fatalf("listener=%v runtime=%v", listenerErr, runtimeErr)
		}
	})
}

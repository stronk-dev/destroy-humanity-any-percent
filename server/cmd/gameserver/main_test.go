package main

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"os"
	"syscall"
	"testing"
	"time"

	"cloud-clicker/server/deploymentconfig"
)

func TestCompositionKeysPreserveCurrentAndPreviousRuntimeMaterial(t *testing.T) {
	runtime := deploymentconfig.Config{
		JWT:       deploymentconfig.KeyPair{CurrentID: "jwt-current", Current: bytes.Repeat([]byte{1}, 32), PreviousID: "jwt-previous", Previous: bytes.Repeat([]byte{2}, 32)},
		Bootstrap: deploymentconfig.KeyPair{CurrentID: "bootstrap-current", Current: bytes.Repeat([]byte{3}, 32), PreviousID: "bootstrap-previous", Previous: bytes.Repeat([]byte{4}, 32)},
	}
	jwt, bootstrap := compositionKeys(runtime)
	if jwt.CurrentID != runtime.JWT.CurrentID || jwt.PreviousID != runtime.JWT.PreviousID || !bytes.Equal(jwt.Current, runtime.JWT.Current) || !bytes.Equal(jwt.Previous, runtime.JWT.Previous) {
		t.Fatalf("JWT composition lost rotation material: %+v", jwt)
	}
	if bootstrap.CurrentID != runtime.Bootstrap.CurrentID || !bytes.Equal(bootstrap.Current, runtime.Bootstrap.Current) || !bytes.Equal(bootstrap.Previous[runtime.Bootstrap.PreviousID], runtime.Bootstrap.Previous) {
		t.Fatalf("bootstrap composition lost rotation material: %+v", bootstrap)
	}
}

func TestShutdownContextReceivesSIGTERM(t *testing.T) {
	ctx, stop := shutdownContext(context.Background())
	defer stop()
	process, err := os.FindProcess(os.Getpid())
	if err != nil {
		t.Fatal(err)
	}
	if err := process.Signal(syscall.SIGTERM); err != nil {
		t.Fatal(err)
	}
	select {
	case <-ctx.Done():
	case <-time.After(time.Second):
		t.Fatal("SIGTERM did not cancel gameserver context")
	}
}

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

package gameserver

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"testing"
	"time"

	"cloud-clicker/server/operations"
)

func TestObserveOperationDoesNotAlarmOnOrderlyCancellation(t *testing.T) {
	var logs bytes.Buffer
	registry, err := operations.NewRegistryWithLogger(unavailableObservationDatabase{}, slog.New(slog.NewJSONHandler(&logs, nil)))
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	now := time.Unix(1_700_000_000, 0)
	called := false
	observed := observeOperation(registry, "credential_cleanup", func() time.Time { return now }, func(actual context.Context) error {
		called = true
		if actual != ctx {
			t.Fatal("observer replaced job context")
		}
		return ctx.Err()
	})
	if err := observed(ctx); err != context.Canceled || !called {
		t.Fatal("observer changed callback execution or returned error")
	}
	requireCredentialCleanupMetric(t, registry, "failure", 0)
	requireCredentialCleanupMetric(t, registry, "success", 0)
	requireCredentialCleanupMetric(t, registry, "canceled", 1)
	if logs.Len() != 0 {
		t.Fatal("orderly cancellation emitted a failure alarm")
	}
}

type unavailableObservationDatabase struct{}

func (unavailableObservationDatabase) PingContext(context.Context) error {
	return errors.New("unused observation database")
}
func (unavailableObservationDatabase) QueryRowContext(context.Context, string, ...any) *sql.Row {
	return nil // Collection stops on the failed ping; no database subject here.
}

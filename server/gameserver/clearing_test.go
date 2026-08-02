package gameserver

import (
	"context"
	"errors"
	"testing"

	"cloud-clicker/server/guild"
)

func TestClearingMembershipRaceRetriesThenCommits(t *testing.T) {
	attempts := 0
	committed, err := retryClearingSnapshot(context.Background(), func() error {
		attempts++
		if attempts < 3 {
			return guild.ErrClearingSnapshotChanged
		}
		return nil
	})
	if err != nil || !committed || attempts != 3 {
		t.Fatalf("committed=%t attempts=%d err=%v", committed, attempts, err)
	}
}

func TestClearingMembershipChurnDefersWithoutKillingWorker(t *testing.T) {
	attempts := 0
	committed, err := retryClearingSnapshot(context.Background(), func() error {
		attempts++
		return guild.ErrClearingSnapshotChanged
	})
	if err != nil || committed || attempts != 3 {
		t.Fatalf("committed=%t attempts=%d err=%v", committed, attempts, err)
	}
	injected := errors.New("database unavailable")
	if _, err := retryClearingSnapshot(context.Background(), func() error { return injected }); !errors.Is(err, injected) {
		t.Fatalf("non-membership error=%v", err)
	}
}

func TestClearingRequiresOnePinnedStockCapPerBoundary(t *testing.T) {
	cap, err := mergePinnedStockCap(0, 100_000)
	if err != nil || cap != 100_000 {
		t.Fatalf("first cap=%d err=%v", cap, err)
	}
	cap, err = mergePinnedStockCap(cap, 100_000)
	if err != nil || cap != 100_000 {
		t.Fatalf("shared cap=%d err=%v", cap, err)
	}
	if cap, err = mergePinnedStockCap(cap, 100_001); !errors.Is(err, ErrClearingDriver) || cap != 0 {
		t.Fatalf("mixed cap=%d err=%v", cap, err)
	}
}

func TestClearingEmptyMembershipIsRetryable(t *testing.T) {
	committed, err := retryClearingSnapshot(context.Background(), func() error {
		return guild.ErrClearingSnapshotChanged
	})
	if err != nil || committed {
		t.Fatalf("committed=%t err=%v", committed, err)
	}
}

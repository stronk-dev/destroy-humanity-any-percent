package account

import (
	"testing"
	"time"
)

func TestSoulRecoveryProgressLimiterLiteralPolicy(t *testing.T) {
	buckets := newRecoveryBuckets(2)
	now := time.UnixMilli(1_000)
	for index := 0; index < 6; index++ {
		if !buckets.allow("session-a", now, 1_000) {
			t.Fatalf("burst token %d rejected", index)
		}
	}
	if buckets.allow("session-a", now, 1_000) {
		t.Fatal("seventh same-instant progress accepted")
	}
	if !buckets.allow("session-a", now.Add(time.Second), 1_000) {
		t.Fatal("one refill interval did not restore one token")
	}
	if buckets.allow("session-a", now.Add(500*time.Millisecond), 1_000) {
		t.Fatal("clock regression refilled the limiter")
	}
	buckets.remove("session-a")
	for index := 0; index < 6; index++ {
		if !buckets.allow("session-a", now, 1_000) {
			t.Fatalf("terminal eviction did not reset burst at %d", index)
		}
	}
}

func TestSoulRecoveryProgressLimiterEvictsAfterIdleTTL(t *testing.T) {
	buckets := newRecoveryBuckets(1)
	now := time.UnixMilli(1_000)
	if !buckets.allow("session-a", now, 1_000) || !buckets.allow("session-b", now.Add(15*time.Minute), 1_000) {
		t.Fatal("idle eviction did not admit replacement session")
	}
}

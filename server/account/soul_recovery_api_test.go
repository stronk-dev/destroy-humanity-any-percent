package account

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"
)

func TestSoulRecoveryProgressSchemaIncludesOnlyWatchdogTerminal(t *testing.T) {
	registry, err := PrivateAPIRegistry()
	if err != nil {
		t.Fatal(err)
	}
	const terminal = `{"action":"cancel","activity_id":"repot","band_after":"dimming","band_before":"dimming","cancelled_by":"watchdog","company_revision":3,"founder_revision":4,"intent_id":"01986666-0000-7000-8000-00000000000a","outcome":"applied","session_id":"01986666-0000-7000-8000-00000000000a","soul_after":40,"soul_before":40}`
	if err := registry.ValidateResponse("progress_soul_recovery", http.StatusOK, []byte(terminal)); err != nil {
		t.Fatalf("watchdog terminal rejected: %v", err)
	}
	for _, name := range []string{"player cancellation", "resolution", "missing reason", "mixed progress", "capability leak", "unknown field"} {
		t.Run(name, func(t *testing.T) {
			var value map[string]any
			if err := json.Unmarshal([]byte(terminal), &value); err != nil {
				t.Fatal(err)
			}
			switch name {
			case "player cancellation":
				value["cancelled_by"] = "player"
			case "resolution":
				value["action"] = "resolve"
			case "missing reason":
				delete(value, "cancelled_by")
			case "mixed progress":
				value["eligible"] = true
			case "capability leak":
				value["progress_token"] = "01986666-0000-4000-8000-000000000001"
			case "unknown field":
				value["extra"] = true
			}
			body, err := json.Marshal(value)
			if err != nil {
				t.Fatal(err)
			}
			if err := registry.ValidateResponse("progress_soul_recovery", http.StatusOK, body); err == nil {
				t.Fatal("progress response accepted an invalid terminal arm")
			}
		})
	}
}

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

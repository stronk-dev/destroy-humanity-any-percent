package production

import (
	"io"
	"log/slog"
	"math/rand"
	"testing"
	"time"

	"cloud-clicker/server/decimal"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/save"
)

func TestParseIntentCanonicalHashAndSemantics(t *testing.T) {
	first, err := parseIntent([]byte(`{
      "intent_id":"018f6b7c-9abc-7def-8abc-0123456789ab",
      "kind":"buy_generator","expected_revision":41,
      "generator_id":"generator.example","count":{"mode":"exact","value":3}
    }`))
	if err != nil {
		t.Fatal(err)
	}
	second, err := parseIntent([]byte(`{"count": {"value": 3, "mode": "exact"}, "generator_id":"generator.example", "expected_revision":41, "kind":"buy_generator", "intent_id":"018f6b7c-9abc-7def-8abc-0123456789ab"}`))
	if err != nil {
		t.Fatal(err)
	}
	if first.RequestHash != second.RequestHash || first.InvalidDetail != "" || first.Count != 3 {
		t.Fatalf("first=%+v second=%+v", first, second)
	}

	invalid, err := parseIntent([]byte(`{"intent_id":"018f6b7c-9abc-7def-8abc-0123456789ab","kind":"perform_manual_batch","expected_revision":1,"action_id":"manual.click","count":0,"window_ms":1}`))
	if err != nil || invalid.InvalidDetail != "count" {
		t.Fatalf("invalid=%+v err=%v", invalid, err)
	}
	if _, err := parseIntent([]byte(`{"intent_id":"not-v7","kind":"buy_generator","expected_revision":1}`)); err == nil {
		t.Fatal("malformed envelope was accepted")
	}
	if _, err := parseIntent([]byte(`{"intent_id":"018f6b7c-9abc-7def-8abc-0123456789ab","kind":"buy_generator","expected_revision":1} {}`)); err == nil {
		t.Fatal("trailing JSON value was accepted")
	}
}

func TestManualTokenRefillAdvancesAtCap(t *testing.T) {
	cursor := time.Date(2026, 7, 28, 8, 0, 0, 0, time.UTC)
	state := &save.State{ManualTokenMilli: 50_000, ManualTokenRefilledAt: cursor}
	policy := economy.ManualPolicy{RefillMilliPerMS: 25, BucketCapMilli: 50_000}
	refillManualTokens(state, policy, cursor.Add(time.Second))
	if state.ManualTokenMilli != 50_000 || !state.ManualTokenRefilledAt.Equal(cursor.Add(time.Second)) {
		t.Fatalf("full refill state = %+v", state)
	}
	state.ManualTokenMilli -= 10_000
	refillManualTokens(state, policy, cursor.Add(time.Second))
	if state.ManualTokenMilli != 40_000 {
		t.Fatalf("spent tokens refilled from stale full-bucket time: %d", state.ManualTokenMilli)
	}
	refillManualTokens(state, policy, cursor.Add(1400*time.Millisecond))
	if state.ManualTokenMilli != 50_000 {
		t.Fatalf("refilled tokens = %d", state.ManualTokenMilli)
	}
}

func TestIntentPolicyPropertyTwentyFourHoursTwoHundredSeeds(t *testing.T) {
	catalog := phase0Catalog(t)
	service := &Service{logger: slog.New(slog.NewTextHandler(io.Discard, nil))}
	for seed := int64(0); seed < 200; seed++ {
		random := rand.New(rand.NewSource(seed))
		state := engineState(t, catalog, "0", 0)
		revision := int64(1)
		for step := 1; step <= 288; step++ {
			now := engineCursor.Add(time.Duration(step) * 5 * time.Minute)
			candidate := clonePolicyState(t, catalog, state)
			var decision save.IntentDecision
			var err error
			if step == 1 || random.Intn(2) == 0 {
				decision, err = service.performManualBatch(parsedIntent{
					IntentID: "018f6b7c-9abc-7def-8abc-999999999999", Kind: IntentPerformManualBatch,
					ExpectedRevision: revision, ActionID: "manual.click", Count: int64(random.Intn(80) + 1), WindowMS: 300_000,
				}, candidate, catalog, save.Revision{Number: revision}, ModeOnline, now, nil)
			} else {
				reports := make([]invariantReport, 0)
				decision, err = service.buyGenerator(parsedIntent{
					IntentID: "018f6b7c-9abc-7def-8abc-999999999999", Kind: IntentBuyGenerator,
					ExpectedRevision: revision, GeneratorID: "generator.beige_tower", CountMode: "max",
				}, candidate, catalog, save.Revision{Number: revision}, ModeOnline, now, nil, &reports)
			}
			if err != nil {
				t.Fatalf("seed=%d step=%d error=%v", seed, step, err)
			}
			if decision.Outcome == save.IntentApplied {
				state = candidate
				revision++
			}
			if _, err := save.EncodeState(state); err != nil {
				t.Fatalf("seed=%d step=%d invalid state: %v", seed, step, err)
			}
			cash, _ := state.Ledger.Balance("company.cash")
			if cash.Lt(decimal.Zero) || !cash.IsStateValue() {
				t.Fatalf("seed=%d step=%d poisoned cash=%s", seed, step, cash.String())
			}
		}
		progress, err := SubProgressValue(catalog, state, 0)
		if err != nil || !progress.IsStateValue() || state.GeneratorCounts["generator.beige_tower"] == 0 {
			t.Fatalf("seed=%d soft lock: progress=%s generators=%d err=%v", seed, progress.String(), state.GeneratorCounts["generator.beige_tower"], err)
		}
	}
}

func clonePolicyState(t *testing.T, catalog *economy.Catalog, source *save.State) *save.State {
	t.Helper()
	ledger, err := economy.RestoreLedger(catalog, economy.ScopeCompany, source.Ledger.Snapshot())
	if err != nil {
		t.Fatal(err)
	}
	counts := make(map[string]int64, len(source.GeneratorCounts))
	for id, count := range source.GeneratorCounts {
		counts[id] = count
	}
	return &save.State{
		Ledger: ledger, GeneratorCounts: counts, EvaluatedThrough: source.EvaluatedThrough,
		ComputeCreditMS: source.ComputeCreditMS, ManualTokenMilli: source.ManualTokenMilli,
		ManualTokenRefilledAt: source.ManualTokenRefilledAt,
	}
}

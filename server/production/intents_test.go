package production

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"math/rand"
	"testing"
	"time"

	"cloud-clicker/server/decimal"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/routes"
	"cloud-clicker/server/save"
)

type fakeInvariantSink struct {
	reports []InvariantReport
}

func (sink *fakeInvariantSink) Report(report InvariantReport) {
	sink.reports = append(sink.reports, report)
}

type fakeInvariantMetrics map[string]int

func (metrics fakeInvariantMetrics) Increment(kind string) {
	metrics[kind]++
}

func TestParseIntentCanonicalHashAndSemantics(t *testing.T) {
	first, err := ParseIntent([]byte(`{
      "intent_id":"018f6b7c-9abc-7def-8abc-0123456789ab",
      "kind":"buy_generator","expected_revision":41,
      "generator_id":"generator.example","count":{"mode":"exact","value":3}
    }`))
	if err != nil {
		t.Fatal(err)
	}
	second, err := ParseIntent([]byte(`{"count": {"value": 3, "mode": "exact"}, "generator_id":"generator.example", "expected_revision":41, "kind":"buy_generator", "intent_id":"018f6b7c-9abc-7def-8abc-0123456789ab"}`))
	if err != nil {
		t.Fatal(err)
	}
	if first.RequestHash != second.RequestHash || first.InvalidDetail != "" || first.Count != 3 {
		t.Fatalf("first=%+v second=%+v", first, second)
	}

	invalid, err := ParseIntent([]byte(`{"intent_id":"018f6b7c-9abc-7def-8abc-0123456789ab","kind":"perform_manual_batch","expected_revision":1,"action_id":"manual.click","count":0,"window_ms":1}`))
	if err != nil || invalid.InvalidDetail != "count" {
		t.Fatalf("invalid=%+v err=%v", invalid, err)
	}
	if _, err := ParseIntent([]byte(`{"intent_id":"not-v7","kind":"buy_generator","expected_revision":1}`)); err == nil {
		t.Fatal("malformed envelope was accepted")
	}
	if _, err := ParseIntent([]byte(`{"intent_id":"018f6b7c-9abc-7def-8abc-0123456789ab","kind":"buy_generator","expected_revision":1} {}`)); err == nil {
		t.Fatal("trailing JSON value was accepted")
	}
	cross, err := ParseIntent([]byte(`{"intent_id":"018f6b7c-9abc-7def-8abc-0123456789ab","kind":"cross_gate","expected_revision":1,"gate_id":"gate.t2_to_t3","route_id":null}`))
	if err != nil || cross.InvalidDetail != "" || cross.GateID != "gate.t2_to_t3" || cross.RouteID != "" {
		t.Fatalf("cross=%+v err=%v", cross, err)
	}
	hint, err := ParseIntent([]byte(`{"intent_id":"018f6b7c-9abc-7def-8abc-0123456789ab","kind":"buy_route_hint","expected_revision":1,"route_id":"route.nonprofit_wrapper_zip"}`))
	if err != nil || hint.InvalidDetail != "" || hint.RouteID != "route.nonprofit_wrapper_zip" {
		t.Fatalf("hint=%+v err=%v", hint, err)
	}
}

func TestCrossGateDiscountSubstituteAndRejections(t *testing.T) {
	economyCatalog, routeCatalog := phase0Catalog(t), phase0Routes(t)
	revision := save.Revision{StreamID: "11111111-1111-4111-8111-111111111111", OwnerID: "22222222-2222-4222-8222-222222222222", Number: 1}
	state := engineState(t, economyCatalog, "1e9", 0)
	state.RunSeq = 1
	state.GatesCrossed = map[string]bool{}
	state.DoctrinesByTransition = map[string]string{"transition.t3_to_t4": "doctrine.capture"}
	state.LedgerFactKinds = map[string]bool{}
	state.MeterBands = map[string]int{}
	state.RegionTraits = map[string]bool{}
	request := IntentRequest{IntentID: "018f6b7c-9abc-7def-8abc-0123456789ab", Kind: IntentCrossGate, GateID: "gate.t2_to_t3", RouteID: "route.ipo_sequence_break"}
	decision, err := TransitionWithRoutes(request, state, economyCatalog, routeCatalog, revision, ModeOnline, engineCursor, nil, nil)
	if err != nil || decision.Outcome != save.IntentApplied || state.Ledger.Snapshot()["company.cash"] != "6e8" || !state.GatesCrossed[request.GateID] || len(decision.Events) != 2 {
		t.Fatalf("discount decision=%+v state=%+v err=%v", decision, state, err)
	}
	if decision.Events[0].Kind != save.EventGateCrossed || decision.Events[1].Kind != save.EventRouteExecuted {
		t.Fatalf("events=%+v", decision.Events)
	}
	already, err := TransitionWithRoutes(request, state, economyCatalog, routeCatalog, save.Revision{Number: 2, StreamID: revision.StreamID, OwnerID: revision.OwnerID}, ModeOnline, engineCursor, nil, nil)
	if err != nil || already.Outcome != save.IntentRejected || !bytes.Contains(already.Receipt, []byte("gate_already_crossed")) {
		t.Fatalf("already=%s err=%v", already.Receipt, err)
	}

	substitute := engineState(t, economyCatalog, "0", 0)
	substitute.RunSeq = 1
	substitute.GatesCrossed = map[string]bool{}
	substitute.DoctrinesByTransition = map[string]string{}
	substitute.StructureID = "structure.nonprofit"
	substitute.LedgerFactKinds = map[string]bool{}
	substitute.MeterBands = map[string]int{}
	substitute.RegionTraits = map[string]bool{}
	request.GateID, request.RouteID = "gate.t4_to_t5", "route.nonprofit_wrapper_zip"
	decision, err = TransitionWithRoutes(request, substitute, economyCatalog, routeCatalog, revision, ModeOnline, engineCursor, nil, nil)
	if err != nil || decision.Outcome != save.IntentApplied || substitute.Ledger.Snapshot()["company.cash"] != "0" {
		t.Fatalf("substitute=%+v err=%v", decision, err)
	}

	unmet := engineState(t, economyCatalog, "1e9", 0)
	unmet.RunSeq = 1
	unmet.GatesCrossed = map[string]bool{}
	unmet.DoctrinesByTransition = map[string]string{}
	unmet.LedgerFactKinds = map[string]bool{}
	unmet.MeterBands = map[string]int{}
	unmet.RegionTraits = map[string]bool{}
	request.GateID, request.RouteID = "gate.t2_to_t3", "route.ipo_sequence_break"
	decision, err = TransitionWithRoutes(request, unmet, economyCatalog, routeCatalog, revision, ModeOnline, engineCursor, nil, nil)
	if err != nil || decision.Outcome != save.IntentRejected || !bytes.Contains(decision.Receipt, []byte("route_predicate_unmet")) {
		t.Fatalf("unmet=%s err=%v", decision.Receipt, err)
	}
}

func TestBuyRouteHintDoesNotAffectPredicateEvaluation(t *testing.T) {
	economyCatalog, routeCatalog := phase0Catalog(t), phase0Routes(t)
	ledger, err := economy.RestoreLedger(economyCatalog, economy.ScopeFounder, map[string]string{})
	if err != nil {
		t.Fatal(err)
	}
	state := &save.State{Ledger: ledger, GeneratorCounts: map[string]int64{}, EvaluatedThrough: engineCursor, ManualTokenRefilledAt: engineCursor, GatesCrossed: map[string]bool{}, DoctrinesByTransition: map[string]string{}, LedgerFactKinds: map[string]bool{}, MeterBands: map[string]int{}, RegionTraits: map[string]bool{}, RouteKnowledgeBalance: 60, HintsUnlocked: map[string]bool{}}
	route, _ := routeCatalog.Route("route.nonprofit_wrapper_zip")
	context, err := routeContext(state, routeCatalog.ContextVersion())
	if err != nil {
		t.Fatal(err)
	}
	before, err := routes.EvaluatePredicate(route.Predicate, context)
	if err != nil {
		t.Fatal(err)
	}
	request := IntentRequest{IntentID: "018f6b7c-9abc-7def-8abc-0123456789ab", Kind: IntentBuyRouteHint, RouteID: route.RouteID}
	decision, err := TransitionWithRoutes(request, state, economyCatalog, routeCatalog, save.Revision{Number: 1}, ModeOnline, engineCursor, nil, nil)
	if err != nil || decision.Outcome != save.IntentApplied || state.RouteKnowledgeBalance != 10 || !state.HintsUnlocked[route.RouteID] || len(decision.Events) != 1 {
		t.Fatalf("decision=%+v state=%+v err=%v", decision, state, err)
	}
	afterContext, _ := routeContext(state, routeCatalog.ContextVersion())
	after, err := routes.EvaluatePredicate(route.Predicate, afterContext)
	if err != nil || before != after {
		t.Fatalf("predicate changed before=%v after=%v err=%v", before, after, err)
	}
	again, err := TransitionWithRoutes(request, state, economyCatalog, routeCatalog, save.Revision{Number: 2}, ModeOnline, engineCursor, nil, nil)
	if err != nil || !bytes.Contains(again.Receipt, []byte("already_unlocked")) {
		t.Fatalf("again=%s err=%v", again.Receipt, err)
	}
}

func TestInvariantSinkEventsAndOutcomeReporting(t *testing.T) {
	intentID := "018f6b7c-9abc-7def-8abc-0123456789ab"
	sink := &fakeInvariantSink{}
	reportInvariant(sink, InvariantReport{Kind: InvariantAffordFallback, IntentID: intentID, Detail: "generator.example"})
	if len(sink.reports) != 1 || sink.reports[0].Kind != InvariantAffordFallback {
		t.Fatalf("exported sink reports = %+v", sink.reports)
	}

	catalog := phase0Catalog(t)
	state := engineState(t, catalog, "1e2", 0)
	request := IntentRequest{IntentID: intentID, Kind: IntentBuyGenerator}
	decision, err := appliedDecision(request, state, 2, 1, state.Ledger.Snapshot(), nil, []InvariantReport{
		{Kind: InvariantAffordFallback, IntentID: intentID, Detail: "generator.example"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(decision.Events) != 1 || decision.Events[0].Kind != save.EventInvariantReported {
		t.Fatalf("invariant events = %+v", decision.Events)
	}
	var payload struct {
		Kind   string `json:"invariant_kind"`
		Detail string `json:"detail"`
	}
	if err := json.Unmarshal(decision.Events[0].Payload, &payload); err != nil ||
		payload.Kind != string(InvariantAffordFallback) || payload.Detail != "generator.example" {
		t.Fatalf("payload = %+v err=%v", payload, err)
	}
	if _, err := appliedDecision(request, state, 2, 1, state.Ledger.Snapshot(), nil, []InvariantReport{
		{Kind: InvariantAffordFallback, IntentID: "018f6b7c-9abc-7def-8abc-999999999999", Detail: "generator.example"},
	}); err != ErrInvalidEngineState {
		t.Fatalf("mismatched report error = %v", err)
	}

	metrics := fakeInvariantMetrics{}
	var logs bytes.Buffer
	service := &Service{metrics: metrics, logger: slog.New(slog.NewJSONHandler(&logs, nil))}
	reports := []InvariantReport{
		{Kind: InvariantAffordFallback, IntentID: intentID, Detail: "generator.example"},
		{Kind: InvariantResidualAbort, IntentID: intentID, Detail: "generator.example"},
	}
	service.recordCommittedInvariants(save.IntentResult{Outcome: save.IntentRejected}, reports[:1])
	service.recordCommittedInvariants(save.IntentResult{Outcome: save.IntentRejected, Replay: true}, reports[:1])
	service.recordAbortedInvariants(reports)
	if metrics[string(InvariantAffordFallback)] != 1 || metrics[string(InvariantResidualAbort)] != 1 {
		t.Fatalf("metrics = %+v", metrics)
	}
	if got := bytes.Count(logs.Bytes(), []byte(`"msg":"production invariant"`)); got != 2 {
		t.Fatalf("audit records = %d, logs=%s", got, logs.String())
	}
}

func TestWireChangesRejectsMalformedCanonicalValues(t *testing.T) {
	if _, err := wireChanges(map[string]string{"company.cash": "not-canonical"}, map[string]string{"company.cash": "1e0"}); !errors.Is(err, ErrInvalidEngineState) {
		t.Fatalf("malformed before error = %v", err)
	}
	if _, err := wireChanges(map[string]string{"company.cash": "1e0"}, map[string]string{"company.cash": "not-canonical"}); !errors.Is(err, ErrInvalidEngineState) {
		t.Fatalf("malformed after error = %v", err)
	}
	if _, err := wireChanges(map[string]string{"company.cash": "not-canonical"}, map[string]string{"company.cash": "not-canonical"}); !errors.Is(err, ErrInvalidEngineState) {
		t.Fatalf("equal malformed values error = %v", err)
	}
	if _, err := wireChanges(map[string]string{"company.cash": "1e0"}, map[string]string{}); !errors.Is(err, ErrInvalidEngineState) {
		t.Fatalf("missing resource error = %v", err)
	}
	changes, err := wireChanges(map[string]string{"company.cash": "1e0"}, map[string]string{"company.cash": "2e0"})
	if err != nil || len(changes) != 1 || changes[0]["delta"] != "1e0" {
		t.Fatalf("valid changes=%+v err=%v", changes, err)
	}
}

func TestExportedTransitionMatchesServiceMutationCore(t *testing.T) {
	catalog := phase0Catalog(t)
	left := engineState(t, catalog, "1e2", 0)
	right := clonePolicyState(t, catalog, left)
	request, err := ParseIntent([]byte(`{"intent_id":"018f6b7c-9abc-7def-8abc-0123456789ab","kind":"buy_generator","expected_revision":1,"generator_id":"generator.beige_tower","count":{"mode":"exact","value":2}}`))
	if err != nil {
		t.Fatal(err)
	}
	first, err := Transition(request, left, catalog, save.Revision{Number: 1}, ModeOnline, engineCursor, nil, &invariantCollector{})
	if err != nil {
		t.Fatal(err)
	}
	second, err := (&Service{}).buyGenerator(request, right, catalog, save.Revision{Number: 1}, ModeOnline, engineCursor, nil, &invariantCollector{})
	if err != nil {
		t.Fatal(err)
	}
	leftState, _ := save.EncodeState(left)
	rightState, _ := save.EncodeState(right)
	if string(first.Receipt) != string(second.Receipt) || string(leftState) != string(rightState) {
		t.Fatalf("transition differs from service core\n%s\n%s\n%s\n%s", first.Receipt, second.Receipt, leftState, rightState)
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

func TestManualTokenRefillMillisecondBoundaries(t *testing.T) {
	cursor := time.Date(2026, 7, 28, 8, 0, 0, 100_000_000, time.UTC)
	state := &save.State{ManualTokenMilli: 0, ManualTokenRefilledAt: cursor}
	policy := economy.ManualPolicy{RefillMilliPerMS: 25, BucketCapMilli: 50_000}

	refillManualTokens(state, policy, cursor.Add(999*time.Microsecond))
	if state.ManualTokenMilli != 0 || !state.ManualTokenRefilledAt.Equal(cursor) {
		t.Fatalf("sub-millisecond refill state = %+v", state)
	}
	refillManualTokens(state, policy, cursor.Add(time.Millisecond))
	if state.ManualTokenMilli != 25 || !state.ManualTokenRefilledAt.Equal(cursor.Add(time.Millisecond)) {
		t.Fatalf("exact-millisecond refill state = %+v", state)
	}
	refillManualTokens(state, policy, cursor.Add(500*time.Microsecond))
	if state.ManualTokenMilli != 25 || !state.ManualTokenRefilledAt.Equal(cursor.Add(time.Millisecond)) {
		t.Fatalf("rollback refill state = %+v", state)
	}
}

func TestManualIntentRepairsMigratedCursorPhaseMismatch(t *testing.T) {
	catalog := phase0Catalog(t)
	legacy := []byte(`{
      "balances":{"company.cash":"0"},
      "generators":{"generator.beige_tower":0},
      "evaluated_through":"2026-07-28T08:00:00.1009Z",
      "compute_credit_ms":0,
      "manual_token_milli":50000,
      "manual_token_refilled_at":"2026-07-28T08:00:00.1001Z"
    }`)
	state, err := save.RestoreState(legacy, 3, catalog, economy.ScopeCompany, time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	wantMigrated := time.Date(2026, 7, 28, 8, 0, 0, 100_000_000, time.UTC)
	if !state.EvaluatedThrough.Equal(wantMigrated) || !state.ManualTokenRefilledAt.Equal(wantMigrated) {
		t.Fatalf("migrated cursors = %s / %s", state.EvaluatedThrough, state.ManualTokenRefilledAt)
	}

	now := time.Date(2026, 7, 28, 8, 0, 0, 101_500_000, time.UTC)
	service := &Service{}
	decision, err := service.performManualBatch(IntentRequest{
		IntentID: "018f6b7c-9abc-7def-8abc-0123456789ab", Kind: IntentPerformManualBatch,
		ExpectedRevision: 1, ActionID: "manual.click", Count: 1, WindowMS: 1,
	}, state, catalog, save.Revision{Number: 1}, ModeOnline, now, nil)
	if err != nil {
		t.Fatal(err)
	}
	wantAdvanced := time.Date(2026, 7, 28, 8, 0, 0, 101_000_000, time.UTC)
	if decision.Outcome != save.IntentApplied || !state.EvaluatedThrough.Equal(wantAdvanced) ||
		!state.ManualTokenRefilledAt.Equal(wantAdvanced) {
		t.Fatalf("decision=%+v cursors=%s/%s", decision, state.EvaluatedThrough, state.ManualTokenRefilledAt)
	}
	encoded, err := save.EncodeState(state)
	if err != nil {
		t.Fatal(err)
	}
	if !json.Valid(encoded) {
		t.Fatalf("encoded state is invalid JSON: %s", encoded)
	}
	var receipt struct {
		EvaluatedAt string `json:"evaluated_at"`
		Snapshot    struct {
			EvaluatedThrough      string `json:"evaluated_through"`
			ManualTokenRefilledAt string `json:"manual_token_refilled_at"`
		} `json:"snapshot"`
	}
	if err := json.Unmarshal(decision.Receipt, &receipt); err != nil {
		t.Fatal(err)
	}
	if receipt.EvaluatedAt != "2026-07-28T08:00:00.101Z" ||
		receipt.Snapshot.EvaluatedThrough != receipt.EvaluatedAt ||
		receipt.Snapshot.ManualTokenRefilledAt != receipt.EvaluatedAt {
		t.Fatalf("receipt cursors = %+v", receipt)
	}
}

func TestManualIntentCursorOrderingProperty(t *testing.T) {
	catalog := phase0Catalog(t)
	service := &Service{}
	base := time.Date(2026, 7, 28, 8, 0, 0, 100_000_000, time.UTC)
	for seed := int64(0); seed < 200; seed++ {
		random := rand.New(rand.NewSource(seed))
		evaluatedPhase := random.Intn(1_000_000)
		manualPhase := random.Intn(evaluatedPhase + 1)
		legacy, err := json.Marshal(map[string]any{
			"balances":                 map[string]string{"company.cash": "0"},
			"generators":               map[string]int64{"generator.beige_tower": 0},
			"evaluated_through":        base.Add(time.Duration(evaluatedPhase) * time.Nanosecond).Format(time.RFC3339Nano),
			"compute_credit_ms":        0,
			"manual_token_milli":       50_000,
			"manual_token_refilled_at": base.Add(time.Duration(manualPhase) * time.Nanosecond).Format(time.RFC3339Nano),
		})
		if err != nil {
			t.Fatal(err)
		}
		state, err := save.RestoreState(legacy, 3, catalog, economy.ScopeCompany, time.Time{})
		if err != nil {
			t.Fatalf("seed=%d restore: %v", seed, err)
		}
		now := base
		for step := 0; step < 200; step++ {
			now = now.Add(time.Duration(random.Int63n(4_000_000)-500_000) * time.Nanosecond)
			decision, err := service.performManualBatch(IntentRequest{
				IntentID: "018f6b7c-9abc-7def-8abc-0123456789ab", Kind: IntentPerformManualBatch,
				ExpectedRevision: int64(step + 1), ActionID: "manual.click", Count: 1, WindowMS: 1,
			}, state, catalog, save.Revision{Number: int64(step + 1)}, ModeOnline, now, nil)
			if err != nil || decision.Outcome != save.IntentApplied {
				t.Fatalf("seed=%d step=%d decision=%+v err=%v", seed, step, decision, err)
			}
			if state.ManualTokenRefilledAt.After(state.EvaluatedThrough) {
				t.Fatalf("seed=%d step=%d manual=%s > evaluated=%s", seed, step, state.ManualTokenRefilledAt, state.EvaluatedThrough)
			}
			if _, err := save.EncodeState(state); err != nil {
				t.Fatalf("seed=%d step=%d encode: %v", seed, step, err)
			}
		}
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
				decision, err = service.performManualBatch(IntentRequest{
					IntentID: "018f6b7c-9abc-7def-8abc-999999999999", Kind: IntentPerformManualBatch,
					ExpectedRevision: revision, ActionID: "manual.click", Count: int64(random.Intn(80) + 1), WindowMS: 300_000,
				}, candidate, catalog, save.Revision{Number: revision}, ModeOnline, now, nil)
			} else {
				collector := &invariantCollector{}
				decision, err = service.buyGenerator(IntentRequest{
					IntentID: "018f6b7c-9abc-7def-8abc-999999999999", Kind: IntentBuyGenerator,
					ExpectedRevision: revision, GeneratorID: "generator.beige_tower", CountMode: "max",
				}, candidate, catalog, save.Revision{Number: revision}, ModeOnline, now, nil, collector)
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

package production

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"regexp"
	"sort"
	"time"

	"cloud-clicker/server/decimal"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/multiplier"
	"cloud-clicker/server/save"
)

var ErrInvalidIntent = errors.New("invalid production intent")

var (
	intentUUIDV7Pattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-7[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
	intentIDPattern     = regexp.MustCompile(`^[a-z][a-z0-9_]*(?:\.[a-z][a-z0-9_]*)*$`)
)

const (
	IntentBuyGenerator       = "buy_generator"
	IntentPerformManualBatch = "perform_manual_batch"
)

type ContributionProvider interface {
	Contributions(state *save.State, catalog *economy.Catalog) ([]multiplier.Contribution, error)
}

type InvariantMetrics interface {
	Increment(kind string)
}

type Service struct {
	store         *save.Store
	catalogs      save.CatalogResolver
	contributions ContributionProvider
	metrics       InvariantMetrics
	logger        *slog.Logger
}

type HandleResult struct {
	Receipt json.RawMessage
	Replay  bool
}

type parsedIntent struct {
	IntentID         string
	Kind             string
	ExpectedRevision int64
	RequestHash      string
	InvalidDetail    string
	GeneratorID      string
	CountMode        string
	Count            int64
	ActionID         string
	WindowMS         int64
}

type invariantReport struct {
	Kind   string
	Detail string
}

func NewService(
	store *save.Store,
	catalogs save.CatalogResolver,
	contributions ContributionProvider,
	metrics InvariantMetrics,
	logger *slog.Logger,
) (*Service, error) {
	if store == nil || catalogs == nil {
		return nil, ErrInvalidIntent
	}
	if logger == nil {
		logger = slog.New(slog.DiscardHandler)
	}
	return &Service{store: store, catalogs: catalogs, contributions: contributions, metrics: metrics, logger: logger}, nil
}

func (s *Service) Handle(
	ctx context.Context,
	streamID string,
	mode EvaluationMode,
	now time.Time,
	requestBytes []byte,
) (HandleResult, error) {
	request, err := parseIntent(requestBytes)
	if err != nil {
		return HandleResult{}, err
	}
	reports := make([]invariantReport, 0)
	result, err := s.store.ApplyIntent(ctx, streamID, request.ExpectedRevision, request.IntentID, request.RequestHash,
		func(state *save.State, revision save.Revision) (save.IntentDecision, error) {
			catalog, ok := s.catalogs.Resolve(revision.ConstantsHash)
			if !ok {
				return save.IntentDecision{}, fmt.Errorf("%w: unknown catalog %s", ErrInvalidIntent, revision.ConstantsHash)
			}
			if request.InvalidDetail != "" {
				return rejectedDecision(request, revision.Number, "invalid", request.InvalidDetail)
			}
			var contributions []multiplier.Contribution
			if s.contributions != nil {
				var err error
				contributions, err = s.contributions.Contributions(state, catalog)
				if err != nil {
					return save.IntentDecision{}, err
				}
			}
			switch request.Kind {
			case IntentBuyGenerator:
				return s.buyGenerator(request, state, catalog, revision, mode, now, contributions, &reports)
			case IntentPerformManualBatch:
				return s.performManualBatch(request, state, catalog, revision, mode, now, contributions)
			default:
				return rejectedDecision(request, revision.Number, "invalid", request.Kind)
			}
		})
	if err != nil {
		for _, report := range reports {
			if report.Kind == "residual_abort" {
				s.recordInvariant(request.IntentID, report)
			}
		}
		var conflict *save.RevisionConflict
		switch {
		case errors.As(err, &conflict):
			return HandleResult{Receipt: marshalRejection(request.IntentID, conflict.Current, "revision_conflict", "expected_revision")}, nil
		case errors.Is(err, save.ErrIdempotencyConflict):
			current := request.ExpectedRevision
			if loaded, loadErr := s.store.LoadLatest(ctx, streamID); loadErr == nil {
				current = loaded.Revision.Number
			}
			return HandleResult{Receipt: marshalRejection(request.IntentID, current, "idempotency_conflict", request.IntentID)}, nil
		default:
			return HandleResult{}, err
		}
	}
	if !result.Replay {
		for _, report := range reports {
			s.recordInvariant(request.IntentID, report)
		}
	}
	return HandleResult{Receipt: result.Receipt, Replay: result.Replay}, nil
}

func (s *Service) buyGenerator(
	request parsedIntent,
	state *save.State,
	catalog *economy.Catalog,
	revision save.Revision,
	mode EvaluationMode,
	now time.Time,
	contributions []multiplier.Contribution,
	reports *[]invariantReport,
) (save.IntentDecision, error) {
	generator, exists := catalog.GeneratorClass(request.GeneratorID)
	if !exists || generator.Production == nil {
		return rejectedDecision(request, revision.Number, "unknown_id", request.GeneratorID)
	}
	owned, exists := state.GeneratorCounts[request.GeneratorID]
	if !exists {
		return save.IntentDecision{}, ErrInvalidEngineState
	}
	before := state.Ledger.Snapshot()
	if _, err := Evaluate(state, catalog, now, mode, contributions); err != nil {
		return save.IntentDecision{}, err
	}
	cash, exists := state.Ledger.Balance(generator.Price.ResourceID)
	if !exists {
		return save.IntentDecision{}, ErrInvalidEngineState
	}
	count := request.Count
	if request.CountMode == "max" {
		affordability, err := catalog.MaxAffordableDetailed(request.GeneratorID, cash, owned)
		if err != nil {
			return save.IntentDecision{}, err
		}
		count = affordability.Count
		if affordability.UsedFallback {
			*reports = append(*reports, invariantReport{Kind: "afford_fallback", Detail: request.GeneratorID})
		}
	}
	if count <= 0 {
		return rejectedDecision(request, revision.Number, "unaffordable", request.GeneratorID)
	}
	if count > decimal.MaxExactInteger-owned {
		return rejectedDecision(request, revision.Number, "cap_exceeded", request.GeneratorID)
	}
	cost, err := catalog.BulkCost(request.GeneratorID, owned, count)
	if err != nil {
		return rejectedDecision(request, revision.Number, "invalid", request.GeneratorID)
	}
	if cost.Gt(cash) {
		return rejectedDecision(request, revision.Number, "unaffordable", request.GeneratorID)
	}
	residual := cash.Sub(cost).Quantize(decimal.CanonicalSignificantDigits)
	if residual.Lt(decimal.Zero) {
		unit := decimal.New(1, cash.Exponent()-int64(decimal.CanonicalSignificantDigits-1))
		if unit.IsStateValue() && residual.Abs().Lte(unit) {
			cost = cash
			*reports = append(*reports, invariantReport{Kind: "residual_clamp", Detail: request.GeneratorID})
		} else {
			*reports = append(*reports, invariantReport{Kind: "residual_abort", Detail: request.GeneratorID})
			return save.IntentDecision{}, ErrInvalidEngineState
		}
	}
	if _, err := state.Ledger.Apply(economy.Transaction{Entries: []economy.Entry{{
		ResourceID: generator.Price.ResourceID, Delta: cost.Neg(),
	}}}); err != nil {
		return save.IntentDecision{}, err
	}
	state.GeneratorCounts[request.GeneratorID] = owned + count
	return appliedDecision(request, state, revision.Number+1, count, before, []save.EventWrite{
		generatorPurchasedEvent(request, generator.Price.ResourceID, count, cost),
	}, *reports)
}

func (s *Service) performManualBatch(
	request parsedIntent,
	state *save.State,
	catalog *economy.Catalog,
	revision save.Revision,
	mode EvaluationMode,
	now time.Time,
	contributions []multiplier.Contribution,
) (save.IntentDecision, error) {
	action, exists := catalog.ManualAction(request.ActionID)
	if !exists {
		return rejectedDecision(request, revision.Number, "unknown_id", request.ActionID)
	}
	before := state.Ledger.Snapshot()
	if _, err := Evaluate(state, catalog, now, mode, contributions); err != nil {
		return save.IntentDecision{}, err
	}
	policy := catalog.ManualPolicy()
	refillManualTokens(state, policy, now)
	applied := request.Count
	available := state.ManualTokenMilli / 1000
	if applied > available {
		applied = available
	}
	state.ManualTokenMilli -= applied * 1000
	if applied > 0 {
		amount := action.Output.AmountPerAction.Mul(decimal.FromFloat64(float64(applied)))
		if _, err := state.Ledger.Apply(economy.Transaction{Entries: []economy.Entry{{
			ResourceID: action.Output.ResourceID, Delta: amount,
		}}}); err != nil {
			if errors.Is(err, economy.ErrAboveHardcap) {
				return rejectedDecision(request, revision.Number, "cap_exceeded", request.ActionID)
			}
			return save.IntentDecision{}, err
		}
	}
	return appliedDecision(request, state, revision.Number+1, applied, before, nil, nil)
}

func refillManualTokens(state *save.State, policy economy.ManualPolicy, now time.Time) {
	if !now.After(state.ManualTokenRefilledAt) {
		return
	}
	elapsedMS := now.Sub(state.ManualTokenRefilledAt).Milliseconds()
	if elapsedMS <= 0 {
		return
	}
	if state.ManualTokenMilli < policy.BucketCapMilli {
		remaining := policy.BucketCapMilli - state.ManualTokenMilli
		if elapsedMS >= (remaining+policy.RefillMilliPerMS-1)/policy.RefillMilliPerMS {
			state.ManualTokenMilli = policy.BucketCapMilli
		} else {
			state.ManualTokenMilli += elapsedMS * policy.RefillMilliPerMS
		}
	}
	state.ManualTokenRefilledAt = state.ManualTokenRefilledAt.Add(time.Duration(elapsedMS) * time.Millisecond)
}

func appliedDecision(
	request parsedIntent,
	state *save.State,
	newRevision int64,
	appliedCount int64,
	before map[string]string,
	events []save.EventWrite,
	reports []invariantReport,
) (save.IntentDecision, error) {
	for _, report := range reports {
		if report.Kind == "residual_abort" {
			continue
		}
		payload, _ := json.Marshal(map[string]any{"invariant_kind": report.Kind, "detail": report.Detail})
		events = append(events, save.EventWrite{
			Kind: save.EventInvariantReported, SchemaVersion: 1, IntentID: request.IntentID, Payload: payload,
		})
	}
	receipt := map[string]any{
		"intent_id": request.IntentID, "outcome": "applied", "applied_count": appliedCount,
		"receipt":      map[string]any{"changes": wireChanges(before, state.Ledger.Snapshot())},
		"new_revision": newRevision, "evaluated_at": state.EvaluatedThrough.UTC().Format(time.RFC3339Nano),
		"snapshot": wireSnapshot(state),
	}
	encoded, err := json.Marshal(receipt)
	if err != nil {
		return save.IntentDecision{}, err
	}
	return save.IntentDecision{Outcome: save.IntentApplied, Receipt: encoded, Events: events}, nil
}

func rejectedDecision(request parsedIntent, currentRevision int64, category, detail string) (save.IntentDecision, error) {
	return save.IntentDecision{
		Outcome: save.IntentRejected,
		Receipt: marshalRejection(request.IntentID, currentRevision, category, detail),
	}, nil
}

func marshalRejection(intentID string, currentRevision int64, category, detail string) json.RawMessage {
	encoded, _ := json.Marshal(map[string]any{
		"intent_id": intentID, "outcome": "rejected", "current_revision": currentRevision,
		"rejection": map[string]string{"category": category, "detail": detail},
	})
	return encoded
}

func generatorPurchasedEvent(request parsedIntent, resourceID string, count int64, cost decimal.Decimal) save.EventWrite {
	payload, _ := json.Marshal(map[string]any{
		"generator_id": request.GeneratorID, "count": count, "cost_resource_id": resourceID, "cost": cost.String(),
	})
	return save.EventWrite{
		Kind: save.EventGeneratorPurchased, SchemaVersion: 1, IntentID: request.IntentID, Payload: payload,
	}
}

func wireChanges(before, after map[string]string) []map[string]string {
	ids := make([]string, 0, len(after))
	for id := range after {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	changes := make([]map[string]string, 0)
	for _, id := range ids {
		if before[id] == after[id] {
			continue
		}
		beforeValue, _ := decimal.ParseCanonical(before[id])
		afterValue, _ := decimal.ParseCanonical(after[id])
		changes = append(changes, map[string]string{
			"resource_id": id, "before": before[id],
			"delta": afterValue.Sub(beforeValue).Quantize(decimal.CanonicalSignificantDigits).String(), "after": after[id],
		})
	}
	return changes
}

func wireSnapshot(state *save.State) map[string]any {
	return map[string]any{
		"balances": state.Ledger.Snapshot(), "generators": state.GeneratorCounts,
		"evaluated_through": state.EvaluatedThrough.UTC().Format(time.RFC3339Nano),
		"compute_credit_ms": state.ComputeCreditMS, "manual_token_milli": state.ManualTokenMilli,
		"manual_token_refilled_at": state.ManualTokenRefilledAt.UTC().Format(time.RFC3339Nano),
	}
}

func (s *Service) recordInvariant(intentID string, report invariantReport) {
	s.logger.Error("production invariant", "intent_id", intentID, "kind", report.Kind, "detail", report.Detail)
	if s.metrics != nil {
		s.metrics.Increment(report.Kind)
	}
}

func parseIntent(data []byte) (parsedIntent, error) {
	var root map[string]json.RawMessage
	decoder := json.NewDecoder(bytes.NewReader(data))
	if err := decoder.Decode(&root); err != nil || root == nil {
		return parsedIntent{}, ErrInvalidIntent
	}
	if err := ensureIntentJSONEnd(decoder); err != nil {
		return parsedIntent{}, ErrInvalidIntent
	}
	var request parsedIntent
	if err := json.Unmarshal(root["intent_id"], &request.IntentID); err != nil || !intentUUIDV7Pattern.MatchString(request.IntentID) {
		return parsedIntent{}, ErrInvalidIntent
	}
	if err := json.Unmarshal(root["kind"], &request.Kind); err != nil || request.Kind == "" {
		return parsedIntent{}, ErrInvalidIntent
	}
	if !parsePositiveSafeInt(root["expected_revision"], &request.ExpectedRevision) {
		return parsedIntent{}, ErrInvalidIntent
	}
	request.RequestHash = canonicalRequestHash(root)
	if request.RequestHash == "" {
		return parsedIntent{}, ErrInvalidIntent
	}

	switch request.Kind {
	case IntentBuyGenerator:
		if !hasExactKeys(root, "intent_id", "kind", "expected_revision", "generator_id", "count") {
			request.InvalidDetail = "buy_generator.fields"
			return request, nil
		}
		if err := json.Unmarshal(root["generator_id"], &request.GeneratorID); err != nil || !intentIDPattern.MatchString(request.GeneratorID) {
			request.InvalidDetail = "generator_id"
		}
		var count map[string]json.RawMessage
		if err := json.Unmarshal(root["count"], &count); err != nil {
			request.InvalidDetail = "count"
			return request, nil
		}
		_ = json.Unmarshal(count["mode"], &request.CountMode)
		if request.CountMode == "max" {
			if !hasExactKeys(count, "mode") {
				request.InvalidDetail = "count.max"
			}
		} else if request.CountMode == "exact" {
			if !hasExactKeys(count, "mode", "value") || !parsePositiveSafeInt(count["value"], &request.Count) {
				request.InvalidDetail = "count.exact"
			}
		} else {
			request.InvalidDetail = "count.mode"
		}
	case IntentPerformManualBatch:
		if !hasExactKeys(root, "intent_id", "kind", "expected_revision", "action_id", "count", "window_ms") {
			request.InvalidDetail = "perform_manual_batch.fields"
			return request, nil
		}
		if err := json.Unmarshal(root["action_id"], &request.ActionID); err != nil || !intentIDPattern.MatchString(request.ActionID) {
			request.InvalidDetail = "action_id"
		}
		if !parsePositiveSafeInt(root["count"], &request.Count) {
			request.InvalidDetail = "count"
		}
		if !parsePositiveSafeInt(root["window_ms"], &request.WindowMS) {
			request.InvalidDetail = "window_ms"
		}
	default:
		request.InvalidDetail = request.Kind
	}
	return request, nil
}

func canonicalRequestHash(root map[string]json.RawMessage) string {
	copyRoot := make(map[string]any, len(root)-1)
	for key, raw := range root {
		if key != "intent_id" {
			decoder := json.NewDecoder(bytes.NewReader(raw))
			decoder.UseNumber()
			var value any
			if decoder.Decode(&value) != nil || ensureIntentJSONEnd(decoder) != nil {
				return ""
			}
			copyRoot[key] = value
		}
	}
	encoded, err := json.Marshal(copyRoot)
	if err != nil {
		return ""
	}
	digest := sha256.Sum256(encoded)
	return "sha256:" + hex.EncodeToString(digest[:])
}

func parsePositiveSafeInt(raw json.RawMessage, destination *int64) bool {
	if len(raw) == 0 || json.Unmarshal(raw, destination) != nil {
		return false
	}
	return *destination > 0 && *destination <= decimal.MaxExactInteger
}

func hasExactKeys(values map[string]json.RawMessage, expected ...string) bool {
	if len(values) != len(expected) {
		return false
	}
	for _, key := range expected {
		if _, exists := values[key]; !exists {
			return false
		}
	}
	return true
}

func ensureIntentJSONEnd(decoder *json.Decoder) error {
	var trailing any
	err := decoder.Decode(&trailing)
	if errors.Is(err, io.EOF) {
		return nil
	}
	if err == nil {
		return errors.New("multiple JSON values")
	}
	return err
}

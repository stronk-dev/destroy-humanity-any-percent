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
	"cloud-clicker/server/routes"
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
	IntentCrossGate          = "cross_gate"
	IntentBuyRouteHint       = "buy_route_hint"
)

type RouteCatalogResolver interface {
	ResolveRoutes(constantsHash string) (*routes.Catalog, bool)
}

type RouteProjector interface {
	Project(context.Context, []save.EventRecord) error
	RepairFounder(context.Context, string, *save.State) error
}

type ServiceOption func(*Service) error

func WithRouteCatalogs(resolver RouteCatalogResolver) ServiceOption {
	return func(service *Service) error {
		if resolver == nil {
			return ErrInvalidIntent
		}
		service.routeCatalogs = resolver
		return nil
	}
}

func WithRouteProjector(projector RouteProjector) ServiceOption {
	return func(service *Service) error {
		if projector == nil {
			return ErrInvalidIntent
		}
		service.routeProjector = projector
		return nil
	}
}

type ContributionProvider interface {
	Contributions(state *save.State, catalog *economy.Catalog) ([]multiplier.Contribution, error)
}

type InvariantMetrics interface {
	Increment(kind string)
}

type InvariantKind string

const (
	InvariantAffordFallback InvariantKind = "afford_fallback"
	InvariantResidualClamp  InvariantKind = "residual_clamp"
	InvariantResidualAbort  InvariantKind = "residual_abort"
)

type InvariantReport struct {
	Kind     InvariantKind
	IntentID string
	Detail   string
}

type InvariantSink interface {
	Report(InvariantReport)
}

type Service struct {
	store          *save.Store
	catalogs       save.CatalogResolver
	contributions  ContributionProvider
	metrics        InvariantMetrics
	logger         *slog.Logger
	routeCatalogs  RouteCatalogResolver
	routeProjector RouteProjector
}

type HandleResult struct {
	Receipt json.RawMessage
	Replay  bool
}

type IntentRequest struct {
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
	GateID           string
	RouteID          string
}

type invariantCollector struct {
	reports []InvariantReport
}

func (collector *invariantCollector) Report(report InvariantReport) {
	collector.reports = append(collector.reports, report)
}

func NewService(
	store *save.Store,
	catalogs save.CatalogResolver,
	contributions ContributionProvider,
	metrics InvariantMetrics,
	logger *slog.Logger,
	options ...ServiceOption,
) (*Service, error) {
	if store == nil || catalogs == nil {
		return nil, ErrInvalidIntent
	}
	if logger == nil {
		logger = slog.New(slog.DiscardHandler)
	}
	service := &Service{store: store, catalogs: catalogs, contributions: contributions, metrics: metrics, logger: logger}
	for _, option := range options {
		if option == nil || option(service) != nil {
			return nil, ErrInvalidIntent
		}
	}
	return service, nil
}

func (s *Service) Handle(
	ctx context.Context,
	streamID string,
	mode EvaluationMode,
	now time.Time,
	requestBytes []byte,
) (HandleResult, error) {
	request, err := ParseIntent(requestBytes)
	if err != nil {
		return HandleResult{}, err
	}
	collector := &invariantCollector{}
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
			if s.contributions != nil && request.Kind != IntentBuyRouteHint {
				var err error
				contributions, err = s.contributions.Contributions(state, catalog)
				if err != nil {
					return save.IntentDecision{}, err
				}
			}
			var routeCatalog *routes.Catalog
			if request.Kind == IntentCrossGate || request.Kind == IntentBuyRouteHint {
				if s.routeCatalogs == nil || s.routeProjector == nil {
					return save.IntentDecision{}, fmt.Errorf("%w: route runtime unavailable", ErrInvalidIntent)
				}
				var ok bool
				routeCatalog, ok = s.routeCatalogs.ResolveRoutes(revision.ConstantsHash)
				if !ok {
					return save.IntentDecision{}, fmt.Errorf("%w: unknown routes catalog %s", ErrInvalidIntent, revision.ConstantsHash)
				}
				if request.Kind == IntentBuyRouteHint {
					if err := s.routeProjector.RepairFounder(ctx, revision.OwnerID, state); err != nil {
						return save.IntentDecision{}, err
					}
				}
			}
			return TransitionWithRoutes(request, state, catalog, routeCatalog, revision, mode, now, contributions, collector)
		})
	if err != nil {
		s.recordAbortedInvariants(collector.reports)
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
	if s.routeProjector != nil {
		if err := s.routeProjector.Project(ctx, result.Events); err != nil {
			return HandleResult{}, err
		}
	}
	s.recordCommittedInvariants(result, collector.reports)
	return HandleResult{Receipt: result.Receipt, Replay: result.Replay}, nil
}

// Transition is the single deterministic intent transition used by persisted
// service execution and the in-memory balance harness. It performs no I/O and
// mutates only the supplied working state.
func Transition(
	request IntentRequest,
	state *save.State,
	catalog *economy.Catalog,
	revision save.Revision,
	mode EvaluationMode,
	now time.Time,
	contributions []multiplier.Contribution,
	sink InvariantSink,
) (save.IntentDecision, error) {
	return TransitionWithRoutes(request, state, catalog, nil, revision, mode, now, contributions, sink)
}

func TransitionWithRoutes(
	request IntentRequest,
	state *save.State,
	catalog *economy.Catalog,
	routeCatalog *routes.Catalog,
	revision save.Revision,
	mode EvaluationMode,
	now time.Time,
	contributions []multiplier.Contribution,
	sink InvariantSink,
) (save.IntentDecision, error) {
	service := Service{}
	switch request.Kind {
	case IntentBuyGenerator:
		return service.buyGenerator(request, state, catalog, revision, mode, now, contributions, sink)
	case IntentPerformManualBatch:
		return service.performManualBatch(request, state, catalog, revision, mode, now, contributions)
	case IntentCrossGate:
		return service.crossGate(request, state, catalog, routeCatalog, revision, mode, now, contributions)
	case IntentBuyRouteHint:
		return service.buyRouteHint(request, state, routeCatalog, revision)
	default:
		return rejectedDecision(request, revision.Number, "invalid", request.Kind)
	}
}

func (s *Service) crossGate(request IntentRequest, state *save.State, catalog *economy.Catalog, routeCatalog *routes.Catalog, revision save.Revision, mode EvaluationMode, now time.Time, contributions []multiplier.Contribution) (save.IntentDecision, error) {
	if routeCatalog == nil || state == nil || state.Ledger == nil || state.Ledger.Scope() != economy.ScopeCompany || revision.OwnerID == "" || state.RunSeq < 1 {
		return save.IntentDecision{}, ErrInvalidEngineState
	}
	_, exists := routeCatalog.Gate(request.GateID)
	if !exists {
		return rejectedDecision(request, revision.Number, "unknown_id", request.GateID)
	}
	if state.GatesCrossed[request.GateID] {
		return rejectedDecision(request, revision.Number, "gate_already_crossed", request.GateID)
	}
	if request.RouteID != "" {
		if _, exists := routeCatalog.Route(request.RouteID); !exists {
			return rejectedDecision(request, revision.Number, "unknown_id", request.RouteID)
		}
	}
	before := state.Ledger.Snapshot()
	if _, err := Evaluate(state, catalog, now, mode, contributions); err != nil {
		return save.IntentDecision{}, err
	}
	context, err := routeContext(state, routeCatalog.ContextVersion())
	if err != nil {
		return save.IntentDecision{}, err
	}
	resolution, matched, err := routeCatalog.Resolve(request.GateID, request.RouteID, context)
	if err != nil {
		return save.IntentDecision{}, err
	}
	if !matched {
		return rejectedDecision(request, revision.Number, "route_predicate_unmet", request.RouteID)
	}
	entries := make([]economy.Entry, 0, len(resolution.Requirement))
	for _, requirement := range resolution.Requirement {
		balance, exists := state.Ledger.Balance(requirement.ResourceID)
		if !exists {
			return save.IntentDecision{}, ErrInvalidEngineState
		}
		if balance.Lt(requirement.Amount) {
			return rejectedDecision(request, revision.Number, "requirement_not_met", requirement.ResourceID)
		}
		entries = append(entries, economy.Entry{ResourceID: requirement.ResourceID, Delta: requirement.Amount.Neg()})
	}
	if len(entries) > 0 {
		if _, err := state.Ledger.Apply(economy.Transaction{Entries: entries}); err != nil {
			return save.IntentDecision{}, err
		}
	}
	if state.GatesCrossed == nil {
		state.GatesCrossed = map[string]bool{}
	}
	state.GatesCrossed[request.GateID] = true
	events := []save.EventWrite{gateCrossedEvent(request, revision, state.RunSeq)}
	if request.RouteID != "" {
		events = append(events, routeExecutedEvent(request, revision, state.RunSeq))
	}
	return appliedDecision(request, state, revision.Number+1, 1, before, events, nil)
}

func routeContext(state *save.State, version int) (routes.Context, error) {
	resources := make(map[string]decimal.Decimal)
	for id, raw := range state.Ledger.Snapshot() {
		value, err := decimal.ParseCanonical(raw)
		if err != nil {
			return routes.Context{}, ErrInvalidEngineState
		}
		resources[id] = value
	}
	return routes.Context{ContextVersion: version, Resources: resources, DoctrinesByTransition: cloneStrings(state.DoctrinesByTransition), StructureID: state.StructureID, LedgerFactKinds: cloneBools(state.LedgerFactKinds), MeterBands: cloneInts(state.MeterBands), RegionTraits: cloneBools(state.RegionTraits)}, nil
}

func (s *Service) buyRouteHint(request IntentRequest, state *save.State, routeCatalog *routes.Catalog, revision save.Revision) (save.IntentDecision, error) {
	if routeCatalog == nil || state == nil || state.Ledger == nil || state.Ledger.Scope() != economy.ScopeFounder {
		return save.IntentDecision{}, ErrInvalidEngineState
	}
	if _, exists := routeCatalog.Route(request.RouteID); !exists {
		return rejectedDecision(request, revision.Number, "unknown_id", request.RouteID)
	}
	if state.HintsUnlocked[request.RouteID] {
		return rejectedDecision(request, revision.Number, "already_unlocked", request.RouteID)
	}
	cost := routeCatalog.KnowledgePolicy().HintCost
	if state.RouteKnowledgeBalance < cost {
		return rejectedDecision(request, revision.Number, "insufficient_route_knowledge", request.RouteID)
	}
	state.RouteKnowledgeBalance -= cost
	if state.HintsUnlocked == nil {
		state.HintsUnlocked = map[string]bool{}
	}
	state.HintsUnlocked[request.RouteID] = true
	payload, _ := json.Marshal(map[string]any{"route_id": request.RouteID, "cost": cost})
	return appliedDecision(request, state, revision.Number+1, 1, state.Ledger.Snapshot(), []save.EventWrite{{Kind: save.EventRouteHintPurchased, SchemaVersion: 1, IntentID: request.IntentID, Payload: payload}}, nil)
}

func (s *Service) buyGenerator(
	request IntentRequest,
	state *save.State,
	catalog *economy.Catalog,
	revision save.Revision,
	mode EvaluationMode,
	now time.Time,
	contributions []multiplier.Contribution,
	sink InvariantSink,
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
			reportInvariant(sink, InvariantReport{Kind: InvariantAffordFallback, IntentID: request.IntentID, Detail: request.GeneratorID})
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
			reportInvariant(sink, InvariantReport{Kind: InvariantResidualClamp, IntentID: request.IntentID, Detail: request.GeneratorID})
		} else {
			reportInvariant(sink, InvariantReport{Kind: InvariantResidualAbort, IntentID: request.IntentID, Detail: request.GeneratorID})
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
	}, collectorReports(sink))
}

func (s *Service) performManualBatch(
	request IntentRequest,
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
	effectiveNow := save.CanonicalServerTime(now)
	if !effectiveNow.After(state.ManualTokenRefilledAt) {
		return
	}
	elapsedMS := effectiveNow.Sub(state.ManualTokenRefilledAt).Milliseconds()
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
	state.ManualTokenRefilledAt = effectiveNow
}

func appliedDecision(
	request IntentRequest,
	state *save.State,
	newRevision int64,
	appliedCount int64,
	before map[string]string,
	events []save.EventWrite,
	reports []InvariantReport,
) (save.IntentDecision, error) {
	for _, report := range reports {
		if report.IntentID != request.IntentID || report.Detail == "" {
			return save.IntentDecision{}, ErrInvalidEngineState
		}
		if report.Kind == InvariantResidualAbort {
			continue
		}
		if report.Kind != InvariantAffordFallback && report.Kind != InvariantResidualClamp {
			return save.IntentDecision{}, ErrInvalidEngineState
		}
		payload, _ := json.Marshal(map[string]any{"invariant_kind": report.Kind, "detail": report.Detail})
		events = append(events, save.EventWrite{
			Kind: save.EventInvariantReported, SchemaVersion: 1, IntentID: request.IntentID, Payload: payload,
		})
	}
	changes, err := wireChanges(before, state.Ledger.Snapshot())
	if err != nil {
		return save.IntentDecision{}, err
	}
	receipt := map[string]any{
		"intent_id": request.IntentID, "outcome": "applied", "applied_count": appliedCount,
		"receipt":      map[string]any{"changes": changes},
		"new_revision": newRevision, "evaluated_at": state.EvaluatedThrough.UTC().Format(time.RFC3339Nano),
		"snapshot": wireSnapshot(state),
	}
	encoded, err := json.Marshal(receipt)
	if err != nil {
		return save.IntentDecision{}, err
	}
	return save.IntentDecision{Outcome: save.IntentApplied, Receipt: encoded, Events: events}, nil
}

func rejectedDecision(request IntentRequest, currentRevision int64, category, detail string) (save.IntentDecision, error) {
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

func generatorPurchasedEvent(request IntentRequest, resourceID string, count int64, cost decimal.Decimal) save.EventWrite {
	payload, _ := json.Marshal(map[string]any{
		"generator_id": request.GeneratorID, "count": count, "cost_resource_id": resourceID, "cost": cost.String(),
	})
	return save.EventWrite{
		Kind: save.EventGeneratorPurchased, SchemaVersion: 1, IntentID: request.IntentID, Payload: payload,
	}
}

func gateCrossedEvent(request IntentRequest, revision save.Revision, runSeq int64) save.EventWrite {
	var routeID any
	if request.RouteID != "" {
		routeID = request.RouteID
	}
	payload, _ := json.Marshal(map[string]any{"gate_id": request.GateID, "route_id": routeID, "run_id": map[string]any{"company_stream_id": revision.StreamID, "run_seq": runSeq}, "founder_id": revision.OwnerID})
	return save.EventWrite{Kind: save.EventGateCrossed, SchemaVersion: 1, IntentID: request.IntentID, Payload: payload}
}

func routeExecutedEvent(request IntentRequest, revision save.Revision, runSeq int64) save.EventWrite {
	payload, _ := json.Marshal(map[string]any{"route_id": request.RouteID, "gate_id": request.GateID, "run_id": map[string]any{"company_stream_id": revision.StreamID, "run_seq": runSeq}, "founder_id": revision.OwnerID})
	return save.EventWrite{Kind: save.EventRouteExecuted, SchemaVersion: 1, IntentID: request.IntentID, Payload: payload}
}

func wireChanges(before, after map[string]string) ([]map[string]string, error) {
	idSet := make(map[string]struct{}, len(before)+len(after))
	for id := range before {
		idSet[id] = struct{}{}
	}
	for id := range after {
		idSet[id] = struct{}{}
	}
	ids := make([]string, 0, len(idSet))
	for id := range idSet {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	changes := make([]map[string]string, 0)
	for _, id := range ids {
		beforeRaw, beforeExists := before[id]
		afterRaw, afterExists := after[id]
		if !beforeExists || !afterExists {
			return nil, fmt.Errorf("%w: receipt resource set changed at %s", ErrInvalidEngineState, id)
		}
		beforeValue, err := decimal.ParseCanonical(beforeRaw)
		if err != nil {
			return nil, fmt.Errorf("%w: receipt before value for %s: %v", ErrInvalidEngineState, id, err)
		}
		afterValue, err := decimal.ParseCanonical(afterRaw)
		if err != nil {
			return nil, fmt.Errorf("%w: receipt after value for %s: %v", ErrInvalidEngineState, id, err)
		}
		if beforeRaw == afterRaw {
			continue
		}
		delta := afterValue.Sub(beforeValue).Quantize(decimal.CanonicalSignificantDigits)
		changes = append(changes, map[string]string{
			"resource_id": id, "before": beforeRaw,
			"delta": delta.String(), "after": afterRaw,
		})
	}
	return changes, nil
}

func wireSnapshot(state *save.State) map[string]any {
	return map[string]any{
		"balances": state.Ledger.Snapshot(), "generators": state.GeneratorCounts,
		"evaluated_through": state.EvaluatedThrough.UTC().Format(time.RFC3339Nano),
		"compute_credit_ms": state.ComputeCreditMS, "manual_token_milli": state.ManualTokenMilli,
		"manual_token_refilled_at": state.ManualTokenRefilledAt.UTC().Format(time.RFC3339Nano),
		"gates_crossed":            state.GatesCrossed, "run_seq": state.RunSeq,
		"doctrines_by_transition": state.DoctrinesByTransition, "structure_id": state.StructureID,
		"ledger_fact_kinds": sortedBoolKeys(state.LedgerFactKinds), "meter_bands": state.MeterBands,
		"region_traits": sortedBoolKeys(state.RegionTraits), "route_knowledge_balance": state.RouteKnowledgeBalance,
		"hints_unlocked": sortedBoolKeys(state.HintsUnlocked),
	}
}

func cloneStrings(source map[string]string) map[string]string {
	result := make(map[string]string, len(source))
	for key, value := range source {
		result[key] = value
	}
	return result
}
func cloneBools(source map[string]bool) map[string]bool {
	result := make(map[string]bool, len(source))
	for key, value := range source {
		result[key] = value
	}
	return result
}
func cloneInts(source map[string]int) map[string]int {
	result := make(map[string]int, len(source))
	for key, value := range source {
		result[key] = value
	}
	return result
}
func sortedBoolKeys(source map[string]bool) []string {
	keys := make([]string, 0, len(source))
	for key, value := range source {
		if value {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	return keys
}

func collectorReports(sink InvariantSink) []InvariantReport {
	collector, ok := sink.(*invariantCollector)
	if !ok {
		return nil
	}
	return collector.reports
}

func reportInvariant(sink InvariantSink, report InvariantReport) {
	if sink != nil {
		sink.Report(report)
	}
}

func (s *Service) recordInvariant(report InvariantReport) {
	s.logger.Error("production invariant", "intent_id", report.IntentID, "kind", report.Kind, "detail", report.Detail)
	if s.metrics != nil {
		s.metrics.Increment(string(report.Kind))
	}
}

func (s *Service) recordCommittedInvariants(result save.IntentResult, reports []InvariantReport) {
	if result.Replay {
		return
	}
	for _, report := range reports {
		s.recordInvariant(report)
	}
}

func (s *Service) recordAbortedInvariants(reports []InvariantReport) {
	for _, report := range reports {
		if report.Kind == InvariantResidualAbort {
			s.recordInvariant(report)
		}
	}
}

func ParseIntent(data []byte) (IntentRequest, error) {
	var root map[string]json.RawMessage
	decoder := json.NewDecoder(bytes.NewReader(data))
	if err := decoder.Decode(&root); err != nil || root == nil {
		return IntentRequest{}, ErrInvalidIntent
	}
	if err := ensureIntentJSONEnd(decoder); err != nil {
		return IntentRequest{}, ErrInvalidIntent
	}
	var request IntentRequest
	if err := json.Unmarshal(root["intent_id"], &request.IntentID); err != nil || !intentUUIDV7Pattern.MatchString(request.IntentID) {
		return IntentRequest{}, ErrInvalidIntent
	}
	if err := json.Unmarshal(root["kind"], &request.Kind); err != nil || request.Kind == "" {
		return IntentRequest{}, ErrInvalidIntent
	}
	if !parsePositiveSafeInt(root["expected_revision"], &request.ExpectedRevision) {
		return IntentRequest{}, ErrInvalidIntent
	}
	request.RequestHash = canonicalRequestHash(root)
	if request.RequestHash == "" {
		return IntentRequest{}, ErrInvalidIntent
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
	case IntentCrossGate:
		if !hasExactKeys(root, "intent_id", "kind", "expected_revision", "gate_id", "route_id") {
			request.InvalidDetail = "cross_gate.fields"
			return request, nil
		}
		if err := json.Unmarshal(root["gate_id"], &request.GateID); err != nil || !intentIDPattern.MatchString(request.GateID) {
			request.InvalidDetail = "gate_id"
		}
		if string(root["route_id"]) != "null" {
			if err := json.Unmarshal(root["route_id"], &request.RouteID); err != nil || !intentIDPattern.MatchString(request.RouteID) {
				request.InvalidDetail = "route_id"
			}
		}
	case IntentBuyRouteHint:
		if !hasExactKeys(root, "intent_id", "kind", "expected_revision", "route_id") {
			request.InvalidDetail = "buy_route_hint.fields"
			return request, nil
		}
		if err := json.Unmarshal(root["route_id"], &request.RouteID); err != nil || !intentIDPattern.MatchString(request.RouteID) {
			request.InvalidDetail = "route_id"
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

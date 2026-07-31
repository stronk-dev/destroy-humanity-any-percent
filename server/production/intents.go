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
	"strings"
	"time"

	"cloud-clicker/server/accrualhook"
	"cloud-clicker/server/decimal"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/faction"
	"cloud-clicker/server/guild"
	"cloud-clicker/server/multiplier"
	prestigecore "cloud-clicker/server/prestige"
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
	IntentSignCompact        = "sign_compact"
	IntentLeaveCompact       = "leave_compact"
	IntentAcceptExitOffer    = "accept_exit_offer"
	IntentDeclineExitOffer   = "decline_exit_offer"
	IntentWindDown           = "wind_down"
	IntentFileIPO            = "file_ipo"
	IntentIncorporate        = "incorporate"
)

type PrestigePolicyResolver interface {
	ResolvePrestige(constantsHash string) (*prestigecore.Policy, bool)
}

type CompactPolicyResolver interface {
	CompactTitheBand(constantsHash string) (minimumPPM, maximumPPM int64, ok bool)
}

type FactionCatalogResolver interface {
	ResolveFaction(constantsHash string) (*faction.Catalog, bool)
}

type ProgressionRuntimeResolver interface {
	PrestigePolicyResolver
	FactionCatalogResolver
}

type AccrualHook = accrualhook.Hook

type CommonsWeightResolver interface {
	CompactWeightPPM(founderID string) (int64, bool)
}

type RouteCatalogResolver interface {
	ResolveRoutes(constantsHash string) (*routes.Catalog, bool)
}

type RouteProjector interface {
	Project(context.Context, []save.EventRecord) error
	RepairFounder(context.Context, string, *save.State) error
}

type EventProjector interface {
	Project(context.Context, []save.EventRecord) error
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
		service.projectors = append(service.projectors, projector)
		return nil
	}
}

func WithEventProjector(projector EventProjector) ServiceOption {
	return func(service *Service) error {
		if projector == nil {
			return ErrInvalidIntent
		}
		service.projectors = append(service.projectors, projector)
		return nil
	}
}

func WithCompactPolicies(resolver CompactPolicyResolver) ServiceOption {
	return func(service *Service) error {
		if resolver == nil {
			return ErrInvalidIntent
		}
		service.compactPolicies = resolver
		return nil
	}
}

// WithCommonsWeightResolver binds the projection-derived Commons input that
// must be frozen into replay_inputs before the closed hook chain executes.
func WithCommonsWeightResolver(resolver CommonsWeightResolver) ServiceOption {
	return func(service *Service) error {
		if resolver == nil {
			return ErrInvalidIntent
		}
		service.commonsWeights = resolver
		return nil
	}
}

func WithReplayCatalogs(resolver ReplayCatalogResolver) ServiceOption {
	return func(service *Service) error {
		if resolver == nil {
			return ErrInvalidIntent
		}
		service.replayCatalogs = resolver
		return nil
	}
}

func WithProgressionRuntime(resolver ProgressionRuntimeResolver) ServiceOption {
	return func(service *Service) error {
		if resolver == nil {
			return ErrInvalidIntent
		}
		service.prestigePolicies = resolver
		service.factionCatalogs = resolver
		return nil
	}
}

func WithGuildRuntime(resolver guild.CatalogResolver) ServiceOption {
	return func(service *Service) error {
		if resolver == nil {
			return ErrInvalidIntent
		}
		service.guildCatalogs = resolver
		return nil
	}
}

// WithCurrentConstantsHash binds the process's authoritative balance identity.
// Existing runs continue under their pinned hash; only a Prestige transition
// uses this value to assemble and pin the next run.
func WithCurrentConstantsHash(constantsHash string) ServiceOption {
	return func(service *Service) error {
		if len(constantsHash) != len("sha256:")+64 || !strings.HasPrefix(constantsHash, "sha256:") {
			return ErrInvalidIntent
		}
		if _, ok := service.catalogs.Resolve(constantsHash); !ok {
			return ErrInvalidIntent
		}
		service.currentConstantsHash = constantsHash
		return nil
	}
}

type ContributionProvider interface {
	Contributions(context.Context, *save.State, *economy.Catalog, save.Revision) ([]multiplier.Contribution, error)
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
	store                *save.Store
	catalogs             save.CatalogResolver
	contributions        ContributionProvider
	metrics              InvariantMetrics
	logger               *slog.Logger
	routeCatalogs        RouteCatalogResolver
	routeProjector       RouteProjector
	compactPolicies      CompactPolicyResolver
	commonsWeights       CommonsWeightResolver
	replayCatalogs       ReplayCatalogResolver
	factionCatalogs      FactionCatalogResolver
	projectors           []EventProjector
	guildCatalogs        guild.CatalogResolver
	prestigePolicies     PrestigePolicyResolver
	currentConstantsHash string
}

type HandleResult struct {
	Receipt json.RawMessage
	Replay  bool
}

type IntentRequest struct {
	IntentID                string
	Kind                    string
	ExpectedRevision        int64
	RequestHash             string
	CanonicalPayload        []byte
	InvalidDetail           string
	GeneratorID             string
	CountMode               string
	Count                   int64
	ActionID                string
	WindowMS                int64
	GateID                  string
	RouteID                 string
	TithePPM                int64
	ExpectedFounderRevision int64
	OfferID                 string
	FactionID               string
}

type CompactTitheBand struct {
	MinimumPPM int64
	MaximumPPM int64
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
	if service.prestigePolicies != nil && service.currentConstantsHash == "" {
		return nil, ErrInvalidIntent
	}
	if service.prestigePolicies != nil {
		if _, ok := service.catalogs.(save.StatePolicyValidator); !ok {
			return nil, ErrInvalidIntent
		}
		policy, policyOK := service.prestigePolicies.ResolvePrestige(service.currentConstantsHash)
		_, factionOK := service.factionCatalogs.ResolveFaction(service.currentConstantsHash)
		if !policyOK || !factionOK || policy.CatchupCeilingMS <= 0 {
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
	if request.InvalidDetail == "" && (request.Kind == IntentAcceptExitOffer || request.Kind == IntentWindDown || request.Kind == IntentFileIPO) {
		return s.handleExit(ctx, streamID, mode, now, request)
	}
	if request.InvalidDetail == "" && request.Kind == IntentCrossGate && s.prestigePolicies != nil {
		trigger, founderRevision, err := s.scriptedExitDue(ctx, streamID, now)
		if err != nil {
			return HandleResult{}, err
		}
		if trigger {
			return s.handleScriptedCrossGateExit(ctx, streamID, mode, now, request, founderRevision)
		}
	}
	var prestigeFounder *save.Loaded
	var declinedOffers int64
	if request.Kind == IntentCrossGate && s.prestigePolicies != nil {
		loaded, err := s.store.LoadSiblingLatest(ctx, streamID, economy.ScopeFounder)
		if err != nil {
			return HandleResult{}, err
		}
		prestigeFounder = &loaded
		company, err := s.store.LoadLatest(ctx, streamID)
		if err != nil {
			return HandleResult{}, err
		}
		declinedOffers, err = s.store.CountRunEvents(ctx, streamID, save.EventExitOfferDeclined, company.State.RunSeq)
		if err != nil {
			return HandleResult{}, err
		}
	}
	collector := &invariantCollector{}
	result, err := s.store.ApplyIntentLogged(ctx, streamID, request.ExpectedRevision, request.IntentID, request.RequestHash, request.CanonicalPayload,
		func(state *save.State, revision save.Revision, command save.ReplayCommand) (decision save.IntentDecision, replayInputs json.RawMessage, resultErr error) {
			if s.replayCatalogs == nil {
				return save.IntentDecision{}, nil, fmt.Errorf("%w: replay catalog bundle unavailable", ErrInvalidIntent)
			}
			bundle, ok := s.replayCatalogs.ResolveReplayCatalogs(revision.ConstantsHash)
			if !ok {
				return save.IntentDecision{}, nil, fmt.Errorf("%w: replay catalog bundle unavailable", ErrInvalidIntent)
			}
			build := replayBuild{Command: command, Mode: mode, Now: now, IntentKind: request.Kind,
				DeclinedExitOfferCount: declinedOffers, RouteContextVersion: bundle.Routes.ContextVersion()}
			if prestigeFounder != nil {
				value := founderCarry(prestigeFounder.State)
				value.FounderRevision = prestigeFounder.Revision.Number
				build.FounderCarry = &value
			}
			catalog, ok := s.catalogs.Resolve(revision.ConstantsHash)
			if !ok {
				return save.IntentDecision{}, nil, fmt.Errorf("%w: unknown catalog %s", ErrInvalidIntent, revision.ConstantsHash)
			}
			if request.InvalidDetail != "" {
				build.Contributions = []multiplier.Contribution{}
			}
			var contributions []multiplier.Contribution
			if s.contributions != nil && request.Kind != IntentBuyRouteHint {
				var err error
				contributions, err = s.contributions.Contributions(ctx, state, catalog, revision)
				if err != nil {
					return save.IntentDecision{}, nil, err
				}
			}
			build.Contributions = contributions
			if state.CompactMember {
				if s.commonsWeights == nil {
					return save.IntentDecision{}, nil, fmt.Errorf("%w: commons replay input unavailable", ErrInvalidIntent)
				}
				weight, ok := s.commonsWeights.CompactWeightPPM(revision.OwnerID)
				if !ok || weight < 0 || weight > 1_000_000 {
					return save.IntentDecision{}, nil, fmt.Errorf("%w: commons replay input unavailable", ErrInvalidIntent)
				}
				build.CommonsWeightPPM = &weight
			}
			var routeCatalog *routes.Catalog
			if request.Kind == IntentCrossGate || request.Kind == IntentBuyRouteHint {
				if s.routeCatalogs == nil || s.routeProjector == nil {
					return save.IntentDecision{}, nil, fmt.Errorf("%w: route runtime unavailable", ErrInvalidIntent)
				}
				var ok bool
				routeCatalog, ok = s.routeCatalogs.ResolveRoutes(revision.ConstantsHash)
				if !ok {
					return save.IntentDecision{}, nil, fmt.Errorf("%w: unknown routes catalog %s", ErrInvalidIntent, revision.ConstantsHash)
				}
				build.RouteContextVersion = routeCatalog.ContextVersion()
				if err := ValidateRouteCatalogResources(catalog, routeCatalog); err != nil {
					return save.IntentDecision{}, nil, err
				}
				if request.Kind == IntentBuyRouteHint {
					if err := s.routeProjector.RepairFounder(ctx, revision.OwnerID, state); err != nil {
						return save.IntentDecision{}, nil, err
					}
				}
			}
			if request.Kind == IntentSignCompact || request.Kind == IntentLeaveCompact || request.Kind == IntentIncorporate {
				if s.compactPolicies == nil {
					return save.IntentDecision{}, nil, fmt.Errorf("%w: compact runtime unavailable", ErrInvalidIntent)
				}
				minimum, maximum, ok := s.compactPolicies.CompactTitheBand(revision.ConstantsHash)
				if !ok {
					return save.IntentDecision{}, nil, fmt.Errorf("%w: unknown commons catalog %s", ErrInvalidIntent, revision.ConstantsHash)
				}
				if minimum < 0 || maximum > 1_000_000 || minimum > maximum {
					return save.IntentDecision{}, nil, ErrInvalidEngineState
				}
			}
			var factionCatalog *faction.Catalog
			if request.Kind == IntentIncorporate || request.Kind == IntentLeaveCompact || state.FactionID != "" {
				if s.factionCatalogs == nil {
					return save.IntentDecision{}, nil, fmt.Errorf("%w: faction runtime unavailable", ErrInvalidIntent)
				}
				var ok bool
				factionCatalog, ok = s.factionCatalogs.ResolveFaction(revision.ConstantsHash)
				if !ok {
					return save.IntentDecision{}, nil, fmt.Errorf("%w: unknown faction catalog %s", ErrInvalidIntent, revision.ConstantsHash)
				}
				if state.FactionID != "" {
					member, exists := factionCatalog.Faction(state.FactionID)
					if !exists {
						return save.IntentDecision{}, nil, ErrInvalidEngineState
					}
					state.FactionStockResource = member.Produces
				}
			}
			if command.RunLogSeq == 0 {
				decision, resultErr = TransitionWithPolicies(request, state, catalog, routeCatalog, nil, factionCatalog, revision, mode, now, contributions, collector, nil)
				return decision, nil, resultErr
			}
			replayInputs, resultErr = buildReplayInputs(build)
			if resultErr != nil {
				return save.IntentDecision{}, nil, resultErr
			}
			transition, resultErr := ApplyLogged(state, request.CanonicalPayload, bundle, replayInputs)
			collector.reports = append(collector.reports, transition.Invariants...)
			if resultErr != nil {
				return save.IntentDecision{}, nil, resultErr
			}
			decision = save.IntentDecision{Outcome: transition.Outcome, Receipt: transition.Receipt, Events: transition.Events}
			return decision, replayInputs, nil
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
	for _, projector := range s.projectors {
		if err := projector.Project(ctx, result.Events); err != nil {
			return HandleResult{}, err
		}
	}
	s.recordCommittedInvariants(result, collector.reports)
	return HandleResult{Receipt: result.Receipt, Replay: result.Replay}, nil
}

func (s *Service) resolveCommonsReplayWeight(founderID string) (int64, error) {
	if s.commonsWeights == nil {
		return 0, fmt.Errorf("%w: commons replay input unavailable", ErrInvalidIntent)
	}
	weight, ok := s.commonsWeights.CompactWeightPPM(founderID)
	if !ok || weight < 0 || weight > 1_000_000 {
		return 0, fmt.Errorf("%w: commons replay input unavailable", ErrInvalidIntent)
	}
	return weight, nil
}

func ValidateRouteCatalogResources(catalog *economy.Catalog, routeCatalog *routes.Catalog) error {
	if catalog == nil || routeCatalog == nil {
		return ErrInvalidEngineState
	}
	validate := func(id string) error {
		resource, exists := catalog.Resource(id)
		if !exists || resource.Scope != economy.ScopeCompany {
			return fmt.Errorf("%w: routes catalog references unknown company resource %q", ErrInvalidEngineState, id)
		}
		return nil
	}
	for _, gate := range routeCatalog.Gates() {
		for _, requirement := range gate.Requirement {
			if err := validate(requirement.ResourceID); err != nil {
				return err
			}
		}
		for _, route := range gate.Routes {
			for _, condition := range route.Predicate {
				if condition.Kind == routes.ConditionResourceAtLeast || condition.Kind == routes.ConditionResourceAtMost {
					if err := validate(condition.ResourceID); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
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
	return TransitionWithPolicies(request, state, catalog, routeCatalog, nil, nil, revision, mode, now, contributions, sink, nil)
}

func TransitionWithPolicies(request IntentRequest, state *save.State, catalog *economy.Catalog, routeCatalog *routes.Catalog, compactBand *CompactTitheBand, factionCatalog *faction.Catalog, revision save.Revision, mode EvaluationMode, now time.Time, contributions []multiplier.Contribution, sink InvariantSink, hook AccrualHook) (save.IntentDecision, error) {
	service := Service{}
	switch request.Kind {
	case IntentBuyGenerator:
		return service.buyGenerator(request, state, catalog, revision, mode, now, contributions, sink, hook)
	case IntentPerformManualBatch:
		return service.performManualBatch(request, state, catalog, revision, mode, now, contributions, hook)
	case IntentCrossGate:
		return service.crossGate(request, state, catalog, routeCatalog, revision, mode, now, contributions, hook)
	case IntentBuyRouteHint:
		return service.buyRouteHint(request, state, routeCatalog, revision)
	case IntentSignCompact:
		return service.signCompact(request, state, catalog, compactBand, revision, mode, now, contributions, hook)
	case IntentLeaveCompact:
		return service.leaveCompact(request, state, catalog, compactBand, factionCatalog, revision, mode, now, contributions, hook)
	case IntentIncorporate:
		return service.incorporate(request, state, catalog, compactBand, factionCatalog, revision, mode, now, contributions, hook)
	case IntentDeclineExitOffer:
		return service.declineExitOffer(request, state, catalog, revision, mode, now, contributions, hook)
	default:
		return rejectedDecision(request, revision.Number, "invalid", request.Kind)
	}
}

func (s *Service) signCompact(request IntentRequest, state *save.State, catalog *economy.Catalog, band *CompactTitheBand, revision save.Revision, mode EvaluationMode, now time.Time, contributions []multiplier.Contribution, hook AccrualHook) (save.IntentDecision, error) {
	if state == nil || state.Ledger == nil || state.Ledger.Scope() != economy.ScopeCompany || band == nil || band.MinimumPPM < 0 || band.MaximumPPM > 1_000_000 || band.MinimumPPM > band.MaximumPPM || revision.OwnerID == "" || state.RunSeq < 1 {
		return save.IntentDecision{}, ErrInvalidEngineState
	}
	if state.CompactMember {
		return rejectedDecision(request, revision.Number, "already_member", "compact")
	}
	if request.TithePPM < band.MinimumPPM || request.TithePPM > band.MaximumPPM {
		return rejectedDecision(request, revision.Number, "invalid", "tithe_ppm")
	}
	before := state.Ledger.Snapshot()
	result, err := Evaluate(state, catalog, now, mode, contributions)
	if err != nil {
		return save.IntentDecision{}, err
	}
	events, err := runAccrualHook(hook, request.IntentID, state, catalog, revision, result, contributions)
	if err != nil {
		return save.IntentDecision{}, err
	}
	state.CompactMember, state.CompactTithePPM = true, request.TithePPM
	state.CompactSolidarityPPM, state.CompactSamples = 0, []save.CompactSample{}
	payload, _ := json.Marshal(map[string]any{"founder_id": revision.OwnerID, "run_id": map[string]any{"company_stream_id": revision.StreamID, "run_seq": state.RunSeq}, "tithe_ppm": request.TithePPM, "prior_member": false, "new_member": true})
	events = append(events, save.EventWrite{Kind: save.EventCompactSigned, SchemaVersion: 1, IntentID: request.IntentID, Payload: payload})
	return appliedDecision(request, state, revision.Number+1, 1, before, events, nil)
}

func (s *Service) leaveCompact(request IntentRequest, state *save.State, catalog *economy.Catalog, band *CompactTitheBand, factionCatalog *faction.Catalog, revision save.Revision, mode EvaluationMode, now time.Time, contributions []multiplier.Contribution, hook AccrualHook) (save.IntentDecision, error) {
	if state == nil || state.Ledger == nil || state.Ledger.Scope() != economy.ScopeCompany || band == nil || revision.OwnerID == "" || state.RunSeq < 1 {
		return save.IntentDecision{}, ErrInvalidEngineState
	}
	if state.FactionID != "" {
		if factionCatalog == nil {
			return save.IntentDecision{}, ErrInvalidEngineState
		}
		member, ok := factionCatalog.Faction(state.FactionID)
		if !ok {
			return save.IntentDecision{}, ErrInvalidEngineState
		}
		if member.Compact != nil && member.Compact.AutoSign {
			return rejectedDecision(request, revision.Number, "faction_bound", state.FactionID)
		}
	}
	if !state.CompactMember {
		return rejectedDecision(request, revision.Number, "not_member", "compact")
	}
	before := state.Ledger.Snapshot()
	result, err := Evaluate(state, catalog, now, mode, contributions)
	if err != nil {
		return save.IntentDecision{}, err
	}
	events, err := runAccrualHook(hook, request.IntentID, state, catalog, revision, result, contributions)
	if err != nil {
		return save.IntentDecision{}, err
	}
	priorTithe := state.CompactTithePPM
	state.CompactMember, state.CompactTithePPM, state.CompactSolidarityPPM = false, 0, 0
	state.CompactSamples = []save.CompactSample{}
	payload, _ := json.Marshal(map[string]any{"founder_id": revision.OwnerID, "run_id": map[string]any{"company_stream_id": revision.StreamID, "run_seq": state.RunSeq}, "tithe_ppm": priorTithe, "prior_member": true, "new_member": false})
	events = append(events, save.EventWrite{Kind: save.EventCompactLeft, SchemaVersion: 1, IntentID: request.IntentID, Payload: payload})
	return appliedDecision(request, state, revision.Number+1, 1, before, events, nil)
}

func (s *Service) incorporate(request IntentRequest, state *save.State, catalog *economy.Catalog, band *CompactTitheBand, factionCatalog *faction.Catalog, revision save.Revision, mode EvaluationMode, now time.Time, contributions []multiplier.Contribution, hook AccrualHook) (save.IntentDecision, error) {
	if state == nil || state.Ledger == nil || state.Ledger.Scope() != economy.ScopeCompany || factionCatalog == nil || revision.OwnerID == "" || state.RunSeq < 1 {
		return save.IntentDecision{}, ErrInvalidEngineState
	}
	chosen, ok := factionCatalog.Faction(request.FactionID)
	if !ok {
		return rejectedDecision(request, revision.Number, "unknown_id", request.FactionID)
	}
	if state.Tier < 2 {
		return rejectedDecision(request, revision.Number, "not_eligible", "tier")
	}
	if state.FactionID != "" {
		return rejectedDecision(request, revision.Number, "already_incorporated", state.FactionID)
	}
	if chosen.Compact != nil {
		if band == nil || chosen.Compact.TithePPM < band.MinimumPPM || chosen.Compact.TithePPM > band.MaximumPPM {
			return save.IntentDecision{}, ErrInvalidEngineState
		}
	}
	before := state.Ledger.Snapshot()
	result, err := Evaluate(state, catalog, now, mode, contributions)
	if err != nil {
		return save.IntentDecision{}, err
	}
	events, err := runAccrualHook(hook, request.IntentID, state, catalog, revision, result, contributions)
	if err != nil {
		return save.IntentDecision{}, err
	}
	state.FactionID = chosen.ID
	state.FactionStockResource = chosen.Produces
	state.IncorporatedAt = state.EvaluatedThrough
	payload, _ := json.Marshal(map[string]any{
		"founder_id": revision.OwnerID, "run_id": map[string]any{"company_stream_id": revision.StreamID, "run_seq": state.RunSeq},
		"faction_id": chosen.ID, "stock_resource": chosen.Produces, "incorporated_at_ms": state.IncorporatedAt.UnixMilli(), "compact_auto_signed": chosen.Compact != nil,
	})
	events = append(events, save.EventWrite{Kind: save.EventIncorporated, SchemaVersion: 1, IntentID: request.IntentID, Payload: payload})
	if chosen.Compact != nil {
		if state.CompactMember {
			priorTithe := state.CompactTithePPM
			if state.CompactTithePPM < chosen.Compact.TithePPM {
				state.CompactTithePPM = chosen.Compact.TithePPM
			}
			compactPayload, _ := json.Marshal(map[string]any{"founder_id": revision.OwnerID, "run_id": map[string]any{"company_stream_id": revision.StreamID, "run_seq": state.RunSeq}, "prior_tithe_ppm": priorTithe, "new_tithe_ppm": state.CompactTithePPM})
			events = append(events, save.EventWrite{Kind: save.EventCompactTitheRaised, SchemaVersion: 1, IntentID: request.IntentID, Payload: compactPayload})
		} else {
			state.CompactMember, state.CompactTithePPM = true, chosen.Compact.TithePPM
			state.CompactSolidarityPPM, state.CompactSamples = 0, []save.CompactSample{}
			compactPayload, _ := json.Marshal(map[string]any{"founder_id": revision.OwnerID, "run_id": map[string]any{"company_stream_id": revision.StreamID, "run_seq": state.RunSeq}, "tithe_ppm": chosen.Compact.TithePPM, "prior_member": false, "new_member": true})
			events = append(events, save.EventWrite{Kind: save.EventCompactSigned, SchemaVersion: 1, IntentID: request.IntentID, Payload: compactPayload})
		}
	}
	return appliedDecision(request, state, revision.Number+1, 1, before, events, nil)
}

func runAccrualHook(hook AccrualHook, intentID string, state *save.State, catalog *economy.Catalog, revision save.Revision, result EvaluationResult, contributions []multiplier.Contribution) ([]save.EventWrite, error) {
	if hook == nil || result.ElapsedMS == 0 {
		return nil, nil
	}
	events, err := hook.AfterAccrual(state, catalog, revision, accrualhook.Result{Receipt: result.Receipt, ElapsedMS: result.ElapsedMS, ProductionMS: result.ProductionMS, BankedCreditMS: result.BankedCreditMS, ProgressDeltaPPM: result.ProgressDeltaPPM}, contributions)
	if err != nil {
		return nil, err
	}
	for index := range events {
		if events[index].IntentID != "" && events[index].IntentID != intentID {
			return nil, ErrInvalidEngineState
		}
		events[index].IntentID = intentID
	}
	return events, nil
}

type accrualHookChain []AccrualHook

func appendAccrualHook(existing, next AccrualHook) AccrualHook {
	if existing == nil {
		return next
	}
	if chain, ok := existing.(accrualHookChain); ok {
		return append(chain, next)
	}
	return accrualHookChain{existing, next}
}

func (chain accrualHookChain) AfterAccrual(state *save.State, catalog *economy.Catalog, revision save.Revision, result accrualhook.Result, contributions []multiplier.Contribution) ([]save.EventWrite, error) {
	var events []save.EventWrite
	for _, hook := range chain {
		produced, err := hook.AfterAccrual(state, catalog, revision, result, contributions)
		if err != nil {
			return nil, err
		}
		events = append(events, produced...)
	}
	return events, nil
}

func (s *Service) crossGate(request IntentRequest, state *save.State, catalog *economy.Catalog, routeCatalog *routes.Catalog, revision save.Revision, mode EvaluationMode, now time.Time, contributions []multiplier.Contribution, hook AccrualHook) (save.IntentDecision, error) {
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
	result, err := Evaluate(state, catalog, now, mode, contributions)
	if err != nil {
		return save.IntentDecision{}, err
	}
	accrualEvents, err := runAccrualHook(hook, request.IntentID, state, catalog, revision, result, contributions)
	if err != nil {
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
	if err := setTierFromGate(state, request.GateID); err != nil {
		return save.IntentDecision{}, err
	}
	events := append(accrualEvents, gateCrossedEvent(request, revision, state.RunSeq))
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
	hook AccrualHook,
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
	result, err := Evaluate(state, catalog, now, mode, contributions)
	if err != nil {
		return save.IntentDecision{}, err
	}
	accrualEvents, err := runAccrualHook(hook, request.IntentID, state, catalog, revision, result, contributions)
	if err != nil {
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
	accrualEvents = append(accrualEvents, generatorPurchasedEvent(request, generator.Price.ResourceID, count, cost))
	return appliedDecision(request, state, revision.Number+1, count, before, accrualEvents, collectorReports(sink))
}

func (s *Service) performManualBatch(
	request IntentRequest,
	state *save.State,
	catalog *economy.Catalog,
	revision save.Revision,
	mode EvaluationMode,
	now time.Time,
	contributions []multiplier.Contribution,
	hook AccrualHook,
) (save.IntentDecision, error) {
	action, exists := catalog.ManualAction(request.ActionID)
	if !exists {
		return rejectedDecision(request, revision.Number, "unknown_id", request.ActionID)
	}
	before := state.Ledger.Snapshot()
	result, err := Evaluate(state, catalog, now, mode, contributions)
	if err != nil {
		return save.IntentDecision{}, err
	}
	events, err := runAccrualHook(hook, request.IntentID, state, catalog, revision, result, contributions)
	if err != nil {
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
	return appliedDecision(request, state, revision.Number+1, applied, before, events, nil)
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
		"compact_member": state.CompactMember, "compact_tithe_ppm": state.CompactTithePPM,
		"compact_solidarity_ppm": state.CompactSolidarityPPM,
		"tier":                   state.Tier, "lifetime_value": state.LifetimeValue.String(),
		"offer_state": wireOfferState(state.OfferState), "run_started_at_ms": wireTimeMS(state.RunStartedAt), "run_pre_timer": state.RunPreTimer,
		"offline_spans":        wireOfflineSpans(state.OfflineSpans),
		"collapsed_offline_ms": state.CollapsedOfflineMS,
		"faction_id":           nullableString(state.FactionID), "incorporated_at_ms": nullableTimeMS(state.IncorporatedAt),
		"stock_resource": nullableString(state.FactionStockResource), "stock_units": state.StockUnits,
		"stock_progress_ms": state.StockProgressMS, "consumed_stock_units": state.ConsumedStockUnits,
		"guild_tithe_carry_ppm": state.GuildTitheCarryPPM, "guild_boundary_seq": state.GuildBoundarySeq,
		"guild_boundary_guild_id":     nullableString(state.GuildBoundaryGuildID),
		"guild_consumed_window_units": state.GuildConsumedWindow,
	}
}

func nullableString(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func nullableTimeMS(value time.Time) any {
	if value.IsZero() {
		return nil
	}
	return value.UnixMilli()
}

func wireTimeMS(value time.Time) int64 {
	if value.IsZero() {
		return 0
	}
	return value.UnixMilli()
}

func wireOfferState(offer *save.ExitOfferState) any {
	if offer == nil {
		return nil
	}
	return map[string]any{"offer_id": offer.OfferID, "exit_type": offer.ExitType, "terms_json": json.RawMessage(offer.TermsJSON), "spawned_at_ms": offer.SpawnedAt.UnixMilli(), "expires_at_ms": offer.ExpiresAt.UnixMilli()}
}

func wireOfflineSpans(spans []save.OfflineSpan) []map[string]int64 {
	result := make([]map[string]int64, len(spans))
	for index, span := range spans {
		result[index] = map[string]int64{"from_ms": span.From.UnixMilli(), "to_ms": span.To.UnixMilli()}
	}
	return result
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
	request.CanonicalPayload, request.RequestHash = canonicalRequest(root)
	if request.RequestHash == "" || len(request.CanonicalPayload) == 0 {
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
	case IntentSignCompact:
		if !hasExactKeys(root, "intent_id", "kind", "expected_revision", "tithe_ppm") {
			request.InvalidDetail = "sign_compact.fields"
			return request, nil
		}
		if !parseNonNegativeSafeInt(root["tithe_ppm"], &request.TithePPM) || request.TithePPM > 1_000_000 {
			request.InvalidDetail = "tithe_ppm"
		}
	case IntentLeaveCompact:
		if !hasExactKeys(root, "intent_id", "kind", "expected_revision") {
			request.InvalidDetail = "leave_compact.fields"
		}
	case IntentIncorporate:
		if !hasExactKeys(root, "intent_id", "kind", "expected_revision", "faction_id") {
			request.InvalidDetail = "incorporate.fields"
			return request, nil
		}
		if err := json.Unmarshal(root["faction_id"], &request.FactionID); err != nil || !intentIDPattern.MatchString(request.FactionID) {
			request.InvalidDetail = "faction_id"
		}
	case IntentAcceptExitOffer:
		if !hasExactKeys(root, "intent_id", "kind", "expected_revision", "expected_founder_revision", "offer_id") {
			request.InvalidDetail = "accept_exit_offer.fields"
			return request, nil
		}
		if !parsePositiveSafeInt(root["expected_founder_revision"], &request.ExpectedFounderRevision) {
			request.InvalidDetail = "expected_founder_revision"
		}
		if err := json.Unmarshal(root["offer_id"], &request.OfferID); err != nil || !intentUUIDV7Pattern.MatchString(request.OfferID) {
			request.InvalidDetail = "offer_id"
		}
	case IntentWindDown, IntentFileIPO:
		if !hasExactKeys(root, "intent_id", "kind", "expected_revision", "expected_founder_revision") {
			request.InvalidDetail = request.Kind + ".fields"
			return request, nil
		}
		if !parsePositiveSafeInt(root["expected_founder_revision"], &request.ExpectedFounderRevision) {
			request.InvalidDetail = "expected_founder_revision"
		}
	case IntentDeclineExitOffer:
		if !hasExactKeys(root, "intent_id", "kind", "expected_revision", "offer_id") {
			request.InvalidDetail = "decline_exit_offer.fields"
			return request, nil
		}
		if err := json.Unmarshal(root["offer_id"], &request.OfferID); err != nil || !intentUUIDV7Pattern.MatchString(request.OfferID) {
			request.InvalidDetail = "offer_id"
		}
	default:
		request.InvalidDetail = request.Kind
	}
	return request, nil
}

func parseNonNegativeSafeInt(raw json.RawMessage, destination *int64) bool {
	return len(raw) > 0 && json.Unmarshal(raw, destination) == nil && *destination >= 0 && *destination <= decimal.MaxExactInteger
}

func canonicalRequest(root map[string]json.RawMessage) ([]byte, string) {
	copyRoot := make(map[string]any, len(root)-1)
	for key, raw := range root {
		if key != "intent_id" {
			decoder := json.NewDecoder(bytes.NewReader(raw))
			decoder.UseNumber()
			var value any
			if decoder.Decode(&value) != nil || ensureIntentJSONEnd(decoder) != nil {
				return nil, ""
			}
			copyRoot[key] = value
		}
	}
	encoded, err := json.Marshal(copyRoot)
	if err != nil {
		return nil, ""
	}
	digest := sha256.Sum256(encoded)
	return encoded, "sha256:" + hex.EncodeToString(digest[:])
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

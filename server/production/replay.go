package production

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"cloud-clicker/server/accrualhook"
	"cloud-clicker/server/achievements"
	"cloud-clicker/server/activeplay"
	"cloud-clicker/server/curriculum"
	"cloud-clicker/server/decimal"
	"cloud-clicker/server/doctrine"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/faction"
	"cloud-clicker/server/fiscal"
	"cloud-clicker/server/guild"
	"cloud-clicker/server/meters"
	"cloud-clicker/server/minigame"
	"cloud-clicker/server/minigameapi"
	"cloud-clicker/server/multiplier"
	"cloud-clicker/server/pet"
	"cloud-clicker/server/pitch"
	prestigecore "cloud-clicker/server/prestige"
	"cloud-clicker/server/relevancepolicy"
	"cloud-clicker/server/routes"
	"cloud-clicker/server/save"
	"cloud-clicker/server/soul"
)

var ErrInvalidReplayInputs = errors.New("invalid replay inputs")

type CatalogBundle struct {
	ConstantsHash string
	Artifacts     map[string][]byte
	Economy       *economy.Catalog
	Routes        *routes.Catalog
	Commons       ReplayCommonsPolicy
	Prestige      *prestigecore.Policy
	Faction       *faction.Catalog
	Guild         *guild.Catalog
	Doctrines     *doctrine.Catalog
	Meters        *meters.Catalog
	Achievements  *achievements.Catalog
	Minigames     *minigame.Catalog
	MinigameAPI   *minigameapi.Catalog
	Pets          *pet.Catalog
	Fiscal        *fiscal.Catalog
	Soul          *soul.Catalog
	Pitch         *pitch.Catalog
	Opportunities *activeplay.Catalog
	Relevance     *relevancepolicy.RelevancePolicy
	Curriculum    *curriculum.Catalog
	Next          *CatalogBundle
}

type ReplayCommonsPolicy interface {
	MinimumTithePPM() int64
	MaximumTithePPM() int64
	ResolvedAccrualHook(weightPPM *int64) accrualhook.Hook
}

func (bundle CatalogBundle) ResolvePrestige(constantsHash string) (*prestigecore.Policy, bool) {
	return bundle.Prestige, bundle.valid(constantsHash)
}

func (bundle CatalogBundle) ResolveFaction(constantsHash string) (*faction.Catalog, bool) {
	return bundle.Faction, bundle.valid(constantsHash)
}

func (bundle CatalogBundle) ResolveGuild(constantsHash string) (*guild.Catalog, bool) {
	return bundle.Guild, bundle.valid(constantsHash)
}

func (bundle CatalogBundle) ResolveCatchupCeilingMS(constantsHash string) (int64, bool) {
	if !bundle.valid(constantsHash) || bundle.Prestige.CatchupCeilingMS <= 0 {
		return 0, false
	}
	return bundle.Prestige.CatchupCeilingMS, true
}

type ReplayCatalogResolver interface {
	ResolveReplayCatalogs(constantsHash string) (CatalogBundle, bool)
}

type ReplayCatalogSet map[string]CatalogBundle

func (set ReplayCatalogSet) ResolveReplayCatalogs(constantsHash string) (CatalogBundle, bool) {
	bundle, ok := set[constantsHash]
	return bundle, ok && bundle.valid(constantsHash)
}

func (set ReplayCatalogSet) ResolvePrestige(constantsHash string) (*prestigecore.Policy, bool) {
	bundle, ok := set.ResolveReplayCatalogs(constantsHash)
	if !ok {
		return nil, false
	}
	return bundle.Prestige, true
}

func (set ReplayCatalogSet) ResolveFaction(constantsHash string) (*faction.Catalog, bool) {
	bundle, ok := set.ResolveReplayCatalogs(constantsHash)
	if !ok {
		return nil, false
	}
	return bundle.Faction, true
}

func (set ReplayCatalogSet) ResolveTenantContent(constantsHash, engineRef, engineVersion string) (minigame.TenantContent, bool) {
	bundle, ok := set.ResolveReplayCatalogs(constantsHash)
	if !ok || engineRef != pitch.EngineRef || engineVersion != pitch.EngineVersion || bundle.Pitch == nil {
		return minigame.TenantContent{}, false
	}
	data := bundle.Artifacts["pitch"]
	if len(data) == 0 {
		return minigame.TenantContent{}, false
	}
	return minigame.TenantContent{Bytes: bytes.Clone(data), Hash: pitch.ContentHash(data), SchemaVersion: pitch.SchemaVersion}, true
}

func (bundle CatalogBundle) valid(constantsHash string) bool {
	withFoundations := bundle.Meters != nil || bundle.Achievements != nil
	withDoctrines := bundle.Doctrines != nil
	withMinigames := bundle.Minigames != nil
	withPets := bundle.Pets != nil
	withFiscal := bundle.Fiscal != nil
	withSoul := bundle.Soul != nil
	withPitch := bundle.Pitch != nil
	withMinigameAPI := bundle.MinigameAPI != nil
	withOpportunities := bundle.Opportunities != nil
	withRelevance := bundle.Relevance != nil
	withCurriculum := bundle.Curriculum != nil
	expectedArtifacts := 7
	if withFoundations {
		expectedArtifacts = 9
	}
	if withDoctrines {
		expectedArtifacts++
	}
	if withMinigames {
		expectedArtifacts++
	}
	if withPets {
		expectedArtifacts++
	}
	if withFiscal {
		expectedArtifacts++
	}
	if withSoul {
		expectedArtifacts++
	}
	if withPitch {
		expectedArtifacts++
	}
	if withMinigameAPI {
		expectedArtifacts++
	}
	if withOpportunities {
		expectedArtifacts++
	}
	if withRelevance {
		expectedArtifacts++
	}
	if withCurriculum {
		expectedArtifacts++
	}
	if constantsHash == "" || bundle.ConstantsHash != constantsHash || len(bundle.Artifacts) != expectedArtifacts || bundle.Economy == nil ||
		bundle.Routes == nil || bundle.Commons == nil || bundle.Prestige == nil || bundle.Faction == nil || bundle.Guild == nil {
		return false
	}
	for _, name := range [...]string{"categories", "commons", "economy", "factions", "guilds", "prestige", "routes"} {
		if len(bundle.Artifacts[name]) == 0 {
			return false
		}
	}
	if withFoundations && (bundle.Meters == nil || bundle.Achievements == nil || len(bundle.Artifacts["meters"]) == 0 || len(bundle.Artifacts["achievements"]) == 0) {
		return false
	}
	if withDoctrines && (!withFoundations || len(bundle.Artifacts["doctrines"]) == 0) {
		return false
	}
	if withDoctrines && bundle.Doctrines.ValidateRoutes(bundle.Routes) != nil {
		return false
	}
	if withMinigames && (!withFoundations || len(bundle.Artifacts["minigames"]) == 0) || withPets && (!withMinigames || len(bundle.Artifacts["pets"]) == 0) ||
		withFiscal && (!withPets || len(bundle.Artifacts["fiscal"]) == 0) ||
		withSoul && (!withFiscal || len(bundle.Artifacts["soul"]) == 0 || !bundle.Minigames.SchemaSupportsSoul() || !bundle.Pets.SchemaSupportsSoul()) ||
		withPitch && (!withSoul || len(bundle.Artifacts["pitch"]) == 0) ||
		withMinigameAPI && (!withPitch || len(bundle.Artifacts["minigame_api"]) == 0) ||
		withOpportunities && (!withDoctrines || len(bundle.Artifacts["opportunities"]) == 0) ||
		withRelevance && (!withOpportunities || len(bundle.Artifacts["relevance"]) == 0) ||
		withCurriculum && (!withRelevance || len(bundle.Artifacts["curriculum"]) == 0) {
		return false
	}
	if withOpportunities && (bundle.Opportunities.Schedule.MinimumIntervalMS > decimal.MaxExactInteger-bundle.Opportunities.Schedule.LifetimeMS ||
		bundle.Opportunities.Schedule.MinimumIntervalMS+bundle.Opportunities.Schedule.LifetimeMS <= bundle.Prestige.CatchupCeilingMS) {
		return false
	}
	computed, err := save.ConstantsHashArtifacts(bundle.Artifacts)
	return err == nil && computed == constantsHash
}

func (bundle CatalogBundle) versionFloors() (founder, company int) {
	founder, company = save.CurrentVersion, save.CurrentVersion
	if bundle.foundationsActive() {
		founder, company = 16, 16
	}
	if bundle.Doctrines != nil {
		company = 17
	}
	if bundle.Minigames != nil {
		founder = 17
	}
	if bundle.Pets != nil {
		founder = 18
	}
	if bundle.Fiscal != nil {
		founder = 19
	}
	if bundle.Soul != nil {
		founder = 20
	}
	if bundle.MinigameAPI != nil {
		founder = 21
	}
	if bundle.Opportunities != nil {
		company = 18
	}
	return founder, company
}

func exitVersionFloors(current, next CatalogBundle) save.ExitVersionFloors {
	currentFounder, currentCompany := current.versionFloors()
	nextFounder, nextCompany := next.versionFloors()
	return save.ExitVersionFloors{CurrentFounder: currentFounder, CurrentCompany: currentCompany, NextFounder: nextFounder, NextCompany: nextCompany}
}

type replayContribution struct {
	Slot     multiplier.Slot `json:"slot"`
	SourceID string          `json:"source_id"`
	Target   string          `json:"target"`
	Factor   string          `json:"factor"`
}

type replayGuildSettlement struct {
	BoundarySeq int64 `json:"boundary_seq"`
	DebitUnits  int64 `json:"debit_units"`
	CreditUnits int64 `json:"credit_units"`
}

type replayGuildSettlementBatch struct {
	GuildID     string                  `json:"guild_id"`
	BaseSeq     int64                   `json:"base_seq"`
	Settlements []replayGuildSettlement `json:"settlements"`
}

type replayAccrual struct {
	Contributions        []replayContribution       `json:"contributions"`
	CommonsWeightPPM     *int64                     `json:"commons_weight_ppm"`
	GuildSettlementBatch replayGuildSettlementBatch `json:"guild_settlement_batch"`
	RouteContextVersion  int                        `json:"route_context_version"`
}

type replayFounderCarry struct {
	FounderRevision            int64                    `json:"founder_revision"`
	FounderConstantsHash       string                   `json:"founder_constants_hash"`
	ReputationLevel            int64                    `json:"reputation_level"`
	RouteKnowledgeBalance      int64                    `json:"route_knowledge_balance"`
	AgeMS                      int64                    `json:"age_ms"`
	Notoriety                  int64                    `json:"notoriety"`
	AdvisorMode                bool                     `json:"advisor_mode"`
	NetworkSlots               []save.NetworkSlot       `json:"network_slots"`
	LedgerFactKinds            []string                 `json:"ledger_fact_kinds"`
	ExitHistoryCount           int                      `json:"exit_history_count"`
	AchievementsEarnedLifetime []string                 `json:"achievements_earned_lifetime"`
	AchievementScoreLifetime   int64                    `json:"achievement_score_lifetime"`
	FounderExtensions          *replayFounderExtensions `json:"founder_extensions,omitempty"`
}

// replayFounderExtensions closes the v17-v21 Founder state carried through
// the Company replay boundary. Earlier replay-input versions intentionally
// omit it and remain valid only for bundles whose Founder floor is at most 16.
type replayFounderExtensions struct {
	MinigameRatings        map[string]save.MinigameRatingState         `json:"minigame_ratings"`
	MinigameOfflineQuality map[string]save.MinigameOfflineQualityState `json:"minigame_offline_quality"`
	Pets                   map[string]pet.CareState                    `json:"pets"`
	FiscalCredit           int64                                       `json:"fiscal_credit"`
	FiscalPeriodOpenedMS   int64                                       `json:"fiscal_period_opened_wall_ms"`
	FiscalPeriodSequence   int64                                       `json:"fiscal_period_seq"`
	FiscalGeneratorLevels  map[string]int64                            `json:"fiscal_generator_levels"`
	FiscalUnlocks          []string                                    `json:"fiscal_unlocks"`
	Soul                   int64                                       `json:"soul"`
	SoulExhaustedSourceIDs []string                                    `json:"soul_exhausted_source_ids"`
	MinigameSessionSeq     int64                                       `json:"minigame_session_seq"`
}

type replayInputsWire struct {
	Version        int                   `json:"v"`
	Command        save.ReplayCommand    `json:"command"`
	EvaluatedAtMS  int64                 `json:"evaluated_at_ms"`
	EvaluationMode EvaluationMode        `json:"evaluation_mode"`
	OfflineCatchup *replayOfflineCatchup `json:"offline_catchup,omitempty"`
	Resolved       json.RawMessage       `json:"resolved"`
}

type replayOfflineCatchup struct {
	OpenedAtMS  int64             `json:"opened_at_ms"`
	OfflineSpan replayOfflineSpan `json:"offline_span"`
}

type replayOfflineSpan struct {
	FromMS int64 `json:"from_ms"`
	ToMS   int64 `json:"to_ms"`
}

type LoggedTransition struct {
	State      *save.State
	Outcome    save.IntentOutcome
	Receipt    json.RawMessage
	Events     []save.EventWrite
	Invariants []InvariantReport
}

type LoggedExitTransition struct {
	Founder    *save.State
	Company    *save.State
	Decision   save.ExitDecision
	Invariants []InvariantReport
}

var ReplayHookOrder = [...]string{"prestige", "faction", "guild", "commons"}

var ErrReplayClockViolation = errors.New("replay clock violation")

// ApplyLogged is the only replayable non-terminal Company mutation boundary.
// It consumes no projection, clock, or catalog resolver outside its four data
// arguments. Invariant diagnostics are deterministic transition output.
func ApplyLogged(state *save.State, canonicalPayload []byte, catalogs CatalogBundle, replayInputs []byte) (result LoggedTransition, resultErr error) {
	collector := &invariantCollector{}
	defer func() { result.Invariants = append([]InvariantReport(nil), collector.reports...) }()
	if state == nil || !catalogs.valid(catalogs.ConstantsHash) {
		return LoggedTransition{}, fmt.Errorf("%w: catalog bundle", ErrInvalidReplayInputs)
	}
	wire, err := parseReplayInputs(replayInputs)
	if err != nil {
		return LoggedTransition{}, fmt.Errorf("%w: decode envelope: %v", ErrInvalidReplayInputs, err)
	}
	if catalogs.foundationsActive() && wire.Version < 3 {
		return LoggedTransition{}, fmt.Errorf("%w: active foundations require replay inputs v3+", ErrInvalidReplayInputs)
	}
	if isSoulRecoveryPayload(canonicalPayload) {
		suppressed, suppressErr := ApplySuppressedLogged(state, canonicalPayload, catalogs, replayInputs)
		if suppressErr != nil {
			return LoggedTransition{}, suppressErr
		}
		return LoggedTransition{State: suppressed.State, Outcome: save.IntentApplied, Receipt: suppressed.Receipt, Events: suppressed.Events}, nil
	}
	if isMinigameResolutionPayload(canonicalPayload) {
		return applyCompanyMinigameResolution(state, canonicalPayload, catalogs, wire)
	}
	request, err := parseLoggedIntent(canonicalPayload, wire.Command.IntentID)
	if err != nil || request.IntentID != wire.Command.IntentID || request.ExpectedRevision != wire.Command.Revision ||
		wire.Command.RunSeq != state.RunSeq || !bytes.Equal(request.CanonicalPayload, canonicalPayload) {
		return LoggedTransition{}, fmt.Errorf("%w: command envelope intent=%t revision=%t run=%t canonical=%t parse=%v", ErrInvalidReplayInputs,
			request.IntentID == wire.Command.IntentID, request.ExpectedRevision == wire.Command.Revision,
			wire.Command.RunSeq == state.RunSeq, bytes.Equal(request.CanonicalPayload, canonicalPayload), err)
	}
	if request.Kind == IntentBuyRouteHint {
		return LoggedTransition{}, fmt.Errorf("%w: founder-scope intent", ErrInvalidReplayInputs)
	}
	revision := save.Revision{StreamID: wire.Command.CompanyStreamID, OwnerID: wire.Command.FounderID,
		Number: wire.Command.Revision, ConstantsHash: catalogs.ConstantsHash, RunLogSequence: wire.Command.RunLogSeq}
	if err := deriveFactionStockResource(state, catalogs.Faction); err != nil {
		return LoggedTransition{}, err
	}
	stateBefore, err := cloneReplayState(state, catalogs.Economy)
	if err != nil {
		return LoggedTransition{}, fmt.Errorf("%w: snapshot state: %v", ErrInvalidReplayInputs, err)
	}
	defer func() {
		if resultErr != nil || result.Outcome != save.IntentApplied {
			*state = *stateBefore
		}
	}()
	if request.InvalidDetail != "" {
		decision, decisionErr := rejectedDecision(request, revision.Number, "invalid", request.InvalidDetail)
		if decisionErr != nil {
			return LoggedTransition{}, decisionErr
		}
		return LoggedTransition{State: state, Outcome: decision.Outcome, Receipt: decision.Receipt, Events: []save.EventWrite{}}, nil
	}
	now := time.UnixMilli(wire.EvaluatedAtMS).UTC()
	if now.Before(state.EvaluatedThrough) {
		return LoggedTransition{}, ErrReplayClockViolation
	}
	var accrual replayAccrual
	var founder *save.State
	var declined int64
	var activeEvidence *activePlayScheduleEvidence
	if request.Kind == IntentCrossGate {
		var resolved replayCrossGateResolved
		if err := decodeReplayStrict(wire.Resolved, &resolved); err != nil || resolved.IntentKind != request.Kind {
			return LoggedTransition{}, fmt.Errorf("%w: cross-gate resolved union", ErrInvalidReplayInputs)
		}
		accrual, declined = resolved.Accrual, resolved.DeclinedExitOfferCount
		activeEvidence = resolved.ActivePlay
		if resolved.FounderCarry != nil {
			if !validFounderCarry(*resolved.FounderCarry, wire.Version, catalogs) || resolved.FounderCarry.FounderConstantsHash != catalogs.ConstantsHash {
				return LoggedTransition{}, fmt.Errorf("%w: founder carry", ErrInvalidReplayInputs)
			}
			founder, err = stateFromFounderCarry(*resolved.FounderCarry, catalogs)
			if err != nil {
				return LoggedTransition{}, fmt.Errorf("%w: founder carry state", ErrInvalidReplayInputs)
			}
		}
	} else {
		var resolved replayAccrualResolved
		if err := decodeReplayStrict(wire.Resolved, &resolved); err != nil || resolved.IntentKind != request.Kind {
			return LoggedTransition{}, fmt.Errorf("%w: accrual resolved union", ErrInvalidReplayInputs)
		}
		accrual = resolved.Accrual
		activeEvidence = resolved.ActivePlay
		if resolved.FounderCarry != nil {
			if !validFounderCarry(*resolved.FounderCarry, wire.Version, catalogs) || resolved.FounderCarry.FounderConstantsHash != catalogs.ConstantsHash {
				return LoggedTransition{}, fmt.Errorf("%w: founder carry", ErrInvalidReplayInputs)
			}
			founder, err = stateFromFounderCarry(*resolved.FounderCarry, catalogs)
			if err != nil {
				return LoggedTransition{}, fmt.Errorf("%w: founder carry state", ErrInvalidReplayInputs)
			}
		}
	}
	if catalogs.foundationsActive() && wire.Version >= 4 && founder == nil {
		return LoggedTransition{}, fmt.Errorf("%w: active foundation founder carry", ErrInvalidReplayInputs)
	}
	if (state.WireVersion == 18) != (activeEvidence != nil) || state.WireVersion == 18 && wire.Version < 5 {
		return LoggedTransition{}, fmt.Errorf("%w: active-play resolved presence", ErrInvalidReplayInputs)
	}
	externalContributions, err := contributionsFromReplay(accrual)
	if err != nil || accrual.RouteContextVersion != catalogs.Routes.ContextVersion() {
		return LoggedTransition{}, fmt.Errorf("%w: accrual inputs", ErrInvalidReplayInputs)
	}
	if err := applyReplayGuildSettlements(state, accrual.GuildSettlementBatch, catalogs.Faction.StockCap); err != nil {
		return LoggedTransition{}, fmt.Errorf("%w: guild settlement inputs", ErrInvalidReplayInputs)
	}
	baseContributions, err := assembleContributions(state, catalogs.Economy, externalContributions)
	if err != nil {
		return LoggedTransition{}, err
	}
	if state.CompactMember != (accrual.CommonsWeightPPM != nil) {
		return LoggedTransition{}, fmt.Errorf("%w: commons weight presence", ErrInvalidReplayInputs)
	}
	hook := closedReplayAccrualHook(catalogs, accrual.CommonsWeightPPM)
	catchupEvents, err := applyReplayOfflineCatchup(state, catalogs.Economy, wire, baseContributions, hook, revision, request.IntentID)
	if err != nil {
		return LoggedTransition{}, err
	}
	var activeEvents []save.EventWrite
	if activeEvidence != nil {
		activeEvents, err = applyActivePlaySchedule(state, catalogs.Opportunities, catalogs.Prestige, revision.OwnerID, now, *activeEvidence)
		if err != nil {
			return LoggedTransition{}, fmt.Errorf("%w: active-play schedule", ErrInvalidReplayInputs)
		}
		for index := range activeEvents {
			activeEvents[index].IntentID = request.IntentID
		}
	}
	contributionInputs := append([]multiplier.Contribution(nil), externalContributions...)
	if state.WireVersion == 18 {
		activeContributions, activeErr := activePlayContributions(state, catalogs.Opportunities, activeEvidence.AttendedNowMS)
		if activeErr != nil {
			return LoggedTransition{}, activeErr
		}
		contributionInputs = append(contributionInputs, activeContributions...)
	}
	contributions, err := assembleContributions(state, catalogs.Economy, contributionInputs)
	if err != nil {
		return LoggedTransition{}, err
	}
	band := &CompactTitheBand{MinimumPPM: catalogs.Commons.MinimumTithePPM(), MaximumPPM: catalogs.Commons.MaximumTithePPM()}
	var decision save.IntentDecision
	if request.Kind == IntentClaimOpportunity {
		if activeEvidence == nil {
			return LoggedTransition{}, fmt.Errorf("%w: missing active-play schedule resolution", ErrInvalidReplayInputs)
		}
		decision, err = applyClaimOpportunity(request, state, catalogs, revision, wire.EvaluationMode, now, contributions, hook,
			activeEvidence.AttendedNowMS, activeEvidence.MissedOpportunityID, activeEvidence.Claim, false)
	} else {
		if activeEvidence != nil && activeEvidence.Claim != nil {
			return LoggedTransition{}, fmt.Errorf("%w: unexpected active-play claim resolution", ErrInvalidReplayInputs)
		}
		decision, err = transitionWithSimulationPolicy(request, state, catalogs.Economy, catalogs.Routes, catalogs.Doctrines, band, catalogs.Faction,
			revision, wire.EvaluationMode, now, contributions, collector, hook, nil)
	}
	if err != nil {
		return LoggedTransition{}, err
	}
	if decision.Outcome == save.IntentApplied {
		decision.Events = append(append(catchupEvents, activeEvents...), decision.Events...)
		if err := afterPrestigeTransitionResolved(catalogs.Prestige, catalogs.Economy, request, state, revision, now, &decision, founder, declined); err != nil {
			return LoggedTransition{}, err
		}
		if catalogs.foundationsActive() && wire.Version >= 4 {
			if err := validateFoundationHookInputs(catalogs, state, founder); err != nil {
				return LoggedTransition{}, err
			}
			if err := applyFoundationTransition(catalogs, stateBefore, state, founder, revision, request, now, contributions, decision.ActionDebits, false, &decision.Events); err != nil {
				return LoggedTransition{}, err
			}
			if err := refreshAppliedSnapshot(&decision, state, catalogs.Economy); err != nil {
				return LoggedTransition{}, err
			}
		}
	}
	return LoggedTransition{State: state, Outcome: decision.Outcome, Receipt: decision.Receipt, Events: decision.Events}, nil
}

// ApplyLoggedExit is the terminal arm of the same owned boundary. Founder
// input is reconstructed solely from the frozen carry view; the returned
// Founder state is applied by the live transaction but cross-run Founder
// verification remains outside this RFC.
func ApplyLoggedExit(company *save.State, canonicalPayload []byte, catalogs CatalogBundle, replayInputs []byte) (result LoggedExitTransition, resultErr error) {
	collector := &invariantCollector{}
	defer func() { result.Invariants = append([]InvariantReport(nil), collector.reports...) }()
	if company == nil || !catalogs.valid(catalogs.ConstantsHash) {
		return LoggedExitTransition{}, fmt.Errorf("%w: catalog bundle", ErrInvalidReplayInputs)
	}
	wire, err := parseReplayInputs(replayInputs)
	if err != nil {
		return LoggedExitTransition{}, fmt.Errorf("%w: decode envelope: %v", ErrInvalidReplayInputs, err)
	}
	if catalogs.foundationsActive() && wire.Version < 3 {
		return LoggedExitTransition{}, fmt.Errorf("%w: active foundations require replay inputs v3+", ErrInvalidReplayInputs)
	}
	request, err := parseLoggedIntent(canonicalPayload, wire.Command.IntentID)
	if err != nil || request.ExpectedRevision != wire.Command.Revision || wire.Command.RunSeq != company.RunSeq || !bytes.Equal(request.CanonicalPayload, canonicalPayload) {
		return LoggedExitTransition{}, fmt.Errorf("%w: terminal command envelope", ErrInvalidReplayInputs)
	}
	var resolved replayExitResolved
	if err := decodeReplayStrict(wire.Resolved, &resolved); err != nil || resolved.Kind != "exit" || resolved.IntentKind != request.Kind || resolved.NextConstantsHash == "" {
		return LoggedExitTransition{}, fmt.Errorf("%w: terminal resolved union", ErrInvalidReplayInputs)
	}
	next := catalogs.Next
	if resolved.NextConstantsHash == catalogs.ConstantsHash {
		next = &catalogs
	}
	if next == nil || !next.valid(resolved.NextConstantsHash) {
		return LoggedExitTransition{}, fmt.Errorf("%w: next catalog bundle", ErrInvalidReplayInputs)
	}
	if next.foundationsActive() && wire.Version < 3 {
		return LoggedExitTransition{}, fmt.Errorf("%w: foundation activation requires replay inputs v3+", ErrInvalidReplayInputs)
	}
	now := time.UnixMilli(wire.EvaluatedAtMS).UTC()
	if now.Before(company.EvaluatedThrough) {
		return LoggedExitTransition{}, ErrReplayClockViolation
	}
	if err := deriveFactionStockResource(company, catalogs.Faction); err != nil {
		return LoggedExitTransition{}, err
	}
	companyBefore, err := cloneReplayState(company, catalogs.Economy)
	if err != nil {
		return LoggedExitTransition{}, fmt.Errorf("%w: snapshot state: %v", ErrInvalidReplayInputs, err)
	}
	defer func() {
		if resultErr != nil || result.Decision.Outcome != save.IntentApplied {
			*company = *companyBefore
		}
	}()
	if (catalogs.MinigameAPI != nil) != (resolved.MinigameSessionActive != nil) {
		return LoggedExitTransition{}, fmt.Errorf("%w: terminal minigame activity evidence", ErrInvalidReplayInputs)
	}
	if (company.WireVersion == 18) != (resolved.ActivePlay != nil) || company.WireVersion == 18 && wire.Version < 5 ||
		(next.Opportunities != nil) != (resolved.NextActivePlay != nil) {
		return LoggedExitTransition{}, fmt.Errorf("%w: terminal active-play evidence", ErrInvalidReplayInputs)
	}
	externalContributions, err := contributionsFromReplay(resolved.Accrual)
	if err != nil || resolved.Accrual.RouteContextVersion != catalogs.Routes.ContextVersion() {
		return LoggedExitTransition{}, fmt.Errorf("%w: terminal accrual inputs", ErrInvalidReplayInputs)
	}
	if company.CompactMember != (resolved.Accrual.CommonsWeightPPM != nil) || !validFounderCarry(resolved.FounderCarry, wire.Version, catalogs) ||
		resolved.FounderCarry.FounderConstantsHash != catalogs.ConstantsHash || !sortedUniqueMechanical(resolved.ExecutedRouteIDs) {
		return LoggedExitTransition{}, fmt.Errorf("%w: terminal frozen inputs", ErrInvalidReplayInputs)
	}
	if err := applyReplayGuildSettlements(company, resolved.Accrual.GuildSettlementBatch, catalogs.Faction.StockCap); err != nil {
		return LoggedExitTransition{}, fmt.Errorf("%w: terminal guild settlement inputs", ErrInvalidReplayInputs)
	}
	revision := save.Revision{StreamID: wire.Command.CompanyStreamID, OwnerID: wire.Command.FounderID,
		Number: wire.Command.Revision, ConstantsHash: catalogs.ConstantsHash, RunLogSequence: wire.Command.RunLogSeq}
	baseContributions, err := assembleContributions(company, catalogs.Economy, externalContributions)
	if err != nil {
		return LoggedExitTransition{}, err
	}
	hook := closedReplayAccrualHook(catalogs, resolved.Accrual.CommonsWeightPPM)
	catchupEvents, err := applyReplayOfflineCatchup(company, catalogs.Economy, wire, baseContributions, hook, revision, request.IntentID)
	if err != nil {
		return LoggedExitTransition{}, err
	}
	var activeEvents []save.EventWrite
	if resolved.ActivePlay != nil {
		activeEvents, err = applyActivePlaySchedule(company, catalogs.Opportunities, catalogs.Prestige, wire.Command.FounderID, now, *resolved.ActivePlay)
		if err != nil || resolved.ActivePlay.Claim != nil {
			return LoggedExitTransition{}, fmt.Errorf("%w: terminal active-play schedule", ErrInvalidReplayInputs)
		}
		for index := range activeEvents {
			activeEvents[index].IntentID = request.IntentID
		}
	}
	contributionInputs := append([]multiplier.Contribution(nil), externalContributions...)
	if resolved.ActivePlay != nil {
		activeContributions, activeErr := activePlayContributions(company, catalogs.Opportunities, resolved.ActivePlay.AttendedNowMS)
		if activeErr != nil {
			return LoggedExitTransition{}, activeErr
		}
		contributionInputs = append(contributionInputs, activeContributions...)
	}
	contributions, err := assembleContributions(company, catalogs.Economy, contributionInputs)
	if err != nil {
		return LoggedExitTransition{}, err
	}
	founder, err := stateFromFounderCarry(resolved.FounderCarry, catalogs)
	if err != nil {
		return LoggedExitTransition{}, fmt.Errorf("%w: founder carry state: %v", ErrInvalidReplayInputs, err)
	}
	if resolved.MinigameSessionActive != nil && *resolved.MinigameSessionActive {
		decision := rejectedExitDecision(request, wire.Command.Revision, "not_eligible", "minigame_session_active")
		return LoggedExitTransition{Founder: founder, Company: company, Decision: decision}, nil
	}
	var prefix []save.EventWrite
	var actionDebits map[string]string
	var exitType string
	var terms prestigecore.Terms
	if request.Kind == IntentCrossGate {
		transition, transitionErr := transitionWithSimulationPolicy(request, company, catalogs.Economy, catalogs.Routes, catalogs.Doctrines, nil, nil,
			revision, wire.EvaluationMode, now, contributions, collector, hook, nil)
		if transitionErr != nil {
			return LoggedExitTransition{}, transitionErr
		}
		if transition.Outcome == save.IntentRejected {
			return LoggedExitTransition{Founder: founder, Company: company, Decision: save.ExitDecision{Outcome: save.IntentRejected, Receipt: transition.Receipt}}, nil
		}
		actionDebits = transition.ActionDebits
		attended, attendedErr := prestigecore.AttendedMS(company, save.CanonicalServerTime(now))
		minimumAttended := int64(900_000)
		if resolved.SelectedBranch != nil {
			if next.Curriculum == nil || request.GateID != next.Curriculum.FirstFailure.GateID || company.RunSeq != next.Curriculum.FirstFailure.RunSeq {
				return LoggedExitTransition{}, fmt.Errorf("%w: scripted curriculum trigger", ErrInvalidReplayInputs)
			}
			minimumAttended = next.Curriculum.FirstFailure.AttendedMS
		}
		if attendedErr != nil || attended < minimumAttended || len(founder.ExitHistory) != 0 {
			return LoggedExitTransition{}, ErrInvalidEngineState
		}
		exitType, prefix = "scripted_first", append(append(catchupEvents, activeEvents...), transition.Events...)
		terms, err = prestigecore.ComputeTerms(company, founder, catalogs.Prestige, exitType)
	} else {
		scripted := resolved.SelectedBranch != nil
		if request.Kind == IntentFileIPO && !scripted {
			decision := rejectedExitDecision(request, revision.Number, "not_eligible", "ipo_chain")
			return LoggedExitTransition{Founder: founder, Company: company, Decision: decision}, nil
		}
		if request.Kind == IntentWindDown && company.Tier < 1 && !scripted {
			decision := rejectedExitDecision(request, revision.Number, "not_eligible", "tier")
			return LoggedExitTransition{Founder: founder, Company: company, Decision: decision}, nil
		}
		exitType = "collapse"
		if scripted || request.Kind == IntentWindDown && len(founder.ExitHistory) == 0 {
			exitType = "scripted_first"
		}
		var promised *prestigecore.StoredOfferTerms
		if request.Kind == IntentAcceptExitOffer {
			if company.OfferState == nil || company.OfferState.OfferID != request.OfferID {
				decision := rejectedExitDecision(request, revision.Number, "not_eligible", "exit_offer")
				return LoggedExitTransition{Founder: founder, Company: company, Decision: decision}, nil
			}
			if !company.OfferState.ExpiresAt.After(save.CanonicalServerTime(now)) {
				decision := rejectedExitDecision(request, revision.Number, "offer_expired", request.OfferID)
				return LoggedExitTransition{Founder: founder, Company: company, Decision: decision}, nil
			}
			if !jsonSemanticallyEqual(company.OfferState.TermsJSON, resolved.SelectedTerms) {
				return LoggedExitTransition{}, fmt.Errorf("%w: selected offer terms", ErrInvalidReplayInputs)
			}
			stored, storedErr := prestigecore.DecodeStoredOfferTerms(resolved.SelectedTerms)
			if storedErr != nil {
				return LoggedExitTransition{}, storedErr
			}
			promised, exitType = &stored, company.OfferState.ExitType
		}
		evaluation, evaluationErr := Evaluate(company, catalogs.Economy, now, wire.EvaluationMode, contributions)
		if evaluationErr != nil {
			return LoggedExitTransition{}, evaluationErr
		}
		prefix, err = runAccrualHook(hook, request.IntentID, company, catalogs.Economy, revision, evaluation, contributions)
		prefix = append(append(catchupEvents, activeEvents...), prefix...)
		if err == nil && resolved.SelectedBranch != nil && request.Kind != IntentWindDown {
			attended, attendedErr := prestigecore.AttendedMS(company, save.CanonicalServerTime(now))
			if next.Curriculum == nil || attendedErr != nil || company.RunSeq != next.Curriculum.FirstFailure.RunSeq || len(founder.ExitHistory) != 0 ||
				!company.GatesCrossed[next.Curriculum.FirstFailure.GateID] || attended < next.Curriculum.FirstFailure.AttendedMS {
				return LoggedExitTransition{}, fmt.Errorf("%w: scripted curriculum trigger", ErrInvalidReplayInputs)
			}
		}
		if err == nil {
			terms, err = prestigecore.ComputeTerms(company, founder, catalogs.Prestige, exitType)
		}
		if err == nil && promised != nil {
			terms = prestigecore.PromiseTerms(promised.PayoutPreview, prestigecore.ApplyMarketModifier(terms, promised.MarketModifierPPM))
		}
	}
	if err != nil || resolved.SelectedExitType != exitType {
		if err != nil {
			return LoggedExitTransition{}, err
		}
		return LoggedExitTransition{}, fmt.Errorf("%w: selected exit type", ErrInvalidReplayInputs)
	}
	var selectedBranch *curriculum.Branch
	if resolved.SelectedBranch != nil {
		if exitType != "scripted_first" || next.Curriculum == nil || request.Kind == IntentCrossGate {
			return LoggedExitTransition{}, fmt.Errorf("%w: selected branch", ErrInvalidReplayInputs)
		}
		branch, branchErr := next.Curriculum.SelectBranch(company, catalogs.Economy)
		if branchErr != nil || branch.Branch != *resolved.SelectedBranch {
			return LoggedExitTransition{}, fmt.Errorf("%w: selected branch", ErrInvalidReplayInputs)
		}
		selectedBranch = &branch
	} else if exitType == "scripted_first" && next.Curriculum != nil {
		return LoggedExitTransition{}, fmt.Errorf("%w: missing selected branch", ErrInvalidReplayInputs)
	}
	if catalogs.foundationsActive() && wire.Version >= 4 {
		if err := validateFoundationHookInputs(catalogs, company, founder); err != nil {
			return LoggedExitTransition{}, err
		}
		if err := applyFoundationTransition(catalogs, companyBefore, company, founder, revision, request, now, contributions, actionDebits, true, &prefix); err != nil {
			return LoggedExitTransition{}, err
		}
	}
	founderRevision := save.Revision{StreamID: "", OwnerID: wire.Command.FounderID, Number: resolved.FounderCarry.FounderRevision,
		ConstantsHash: catalogs.ConstantsHash}
	decision, err := finishExitResolved(request, founder, founderRevision, company, revision, now, exitType, terms, selectedBranch, prefix,
		resolved.ExecutedRouteIDs, catalogs, *next, resolved.NextActivePlay)
	if err != nil {
		return LoggedExitTransition{}, err
	}
	return LoggedExitTransition{Founder: founder, Company: company, Decision: decision}, nil
}

// previewCurriculumBranch runs the exact frozen terminal accrual on a clone and
// stops before the requested command mutation. The selected branch can then be
// persisted in replay inputs without introducing a second accrual model.
func previewCurriculumBranch(company *save.State, current, next CatalogBundle, request IntentRequest, build replayBuild) (curriculum.Branch, error) {
	if company == nil || next.Curriculum == nil || request.Kind == IntentCrossGate {
		return curriculum.Branch{}, ErrInvalidEngineState
	}
	preview, err := cloneReplayState(company, current.Economy)
	if err != nil {
		return curriculum.Branch{}, err
	}
	if err := deriveFactionStockResource(preview, current.Faction); err != nil {
		return curriculum.Branch{}, err
	}
	accrual, err := makeReplayAccrual(build.Contributions, build.CommonsWeightPPM, build.GuildSettlementBatch, build.RouteContextVersion)
	if err != nil {
		return curriculum.Branch{}, err
	}
	if err := applyReplayGuildSettlements(preview, accrual.GuildSettlementBatch, current.Faction.StockCap); err != nil {
		return curriculum.Branch{}, err
	}
	externalContributions, err := contributionsFromReplay(accrual)
	if err != nil {
		return curriculum.Branch{}, err
	}
	baseContributions, err := assembleContributions(preview, current.Economy, externalContributions)
	if err != nil {
		return curriculum.Branch{}, err
	}
	wire := replayInputsWire{Version: save.ReplayInputsVersion, Command: build.Command, EvaluatedAtMS: build.Now.UnixMilli(), EvaluationMode: build.Mode, OfflineCatchup: build.OfflineCatchup}
	revision := save.Revision{StreamID: build.Command.CompanyStreamID, OwnerID: build.Command.FounderID, Number: build.Command.Revision, ConstantsHash: current.ConstantsHash, RunLogSequence: build.Command.RunLogSeq}
	hook := closedReplayAccrualHook(current, accrual.CommonsWeightPPM)
	if _, err := applyReplayOfflineCatchup(preview, current.Economy, wire, baseContributions, hook, revision, request.IntentID); err != nil {
		return curriculum.Branch{}, err
	}
	if build.ActivePlay != nil {
		if _, err := applyActivePlaySchedule(preview, current.Opportunities, current.Prestige, build.Command.FounderID, save.CanonicalServerTime(build.Now), *build.ActivePlay); err != nil {
			return curriculum.Branch{}, err
		}
	}
	contributionInputs := append([]multiplier.Contribution(nil), externalContributions...)
	if build.ActivePlay != nil {
		activeContributions, activeErr := activePlayContributions(preview, current.Opportunities, build.ActivePlay.AttendedNowMS)
		if activeErr != nil {
			return curriculum.Branch{}, activeErr
		}
		contributionInputs = append(contributionInputs, activeContributions...)
	}
	contributions, err := assembleContributions(preview, current.Economy, contributionInputs)
	if err != nil {
		return curriculum.Branch{}, err
	}
	evaluation, err := Evaluate(preview, current.Economy, build.Now, build.Mode, contributions)
	if err != nil {
		return curriculum.Branch{}, err
	}
	if _, err := runAccrualHook(hook, request.IntentID, preview, current.Economy, revision, evaluation, contributions); err != nil {
		return curriculum.Branch{}, err
	}
	return next.Curriculum.SelectBranch(preview, current.Economy)
}

func jsonSemanticallyEqual(left, right []byte) bool {
	var leftValue, rightValue any
	leftDecoder, rightDecoder := json.NewDecoder(bytes.NewReader(left)), json.NewDecoder(bytes.NewReader(right))
	leftDecoder.UseNumber()
	rightDecoder.UseNumber()
	if leftDecoder.Decode(&leftValue) != nil || rightDecoder.Decode(&rightValue) != nil {
		return false
	}
	leftCanonical, leftErr := json.Marshal(leftValue)
	rightCanonical, rightErr := json.Marshal(rightValue)
	return leftErr == nil && rightErr == nil && bytes.Equal(leftCanonical, rightCanonical)
}

func parseLoggedIntent(canonicalPayload []byte, intentID string) (IntentRequest, error) {
	var root map[string]json.RawMessage
	decoder := json.NewDecoder(bytes.NewReader(canonicalPayload))
	if err := decoder.Decode(&root); err != nil || root == nil {
		return IntentRequest{}, ErrInvalidReplayInputs
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) || root["intent_id"] != nil {
		return IntentRequest{}, ErrInvalidReplayInputs
	}
	encodedIntentID, _ := json.Marshal(intentID)
	root["intent_id"] = encodedIntentID
	full, err := json.Marshal(root)
	if err != nil {
		return IntentRequest{}, err
	}
	return ParseIntent(full)
}

func closedReplayAccrualHook(catalogs CatalogBundle, commonsWeightPPM *int64) AccrualHook {
	var chain AccrualHook
	chain = appendAccrualHook(chain, prestigecore.AccrualHook{Policies: catalogs})
	chain = appendAccrualHook(chain, faction.AccrualHook{Catalogs: catalogs, Policies: catalogs})
	chain = appendAccrualHook(chain, guild.AccrualHook{Catalogs: catalogs})
	chain = appendAccrualHook(chain, catalogs.Commons.ResolvedAccrualHook(commonsWeightPPM))
	return chain
}

func contributionsFromReplay(accrual replayAccrual) ([]multiplier.Contribution, error) {
	if accrual.Contributions == nil || accrual.GuildSettlementBatch.Settlements == nil || accrual.RouteContextVersion < 0 {
		return nil, ErrInvalidReplayInputs
	}
	batch := accrual.GuildSettlementBatch
	if batch.GuildID == "" {
		if batch.BaseSeq != 0 || len(batch.Settlements) != 0 {
			return nil, ErrInvalidReplayInputs
		}
	} else if !intentUUIDV7Pattern.MatchString(batch.GuildID) || batch.BaseSeq < 0 || batch.BaseSeq > decimal.MaxExactInteger {
		return nil, ErrInvalidReplayInputs
	}
	lastSettlement := batch.BaseSeq
	for _, settlement := range batch.Settlements {
		if settlement.BoundarySeq <= lastSettlement || settlement.BoundarySeq > decimal.MaxExactInteger || settlement.DebitUnits < 0 ||
			settlement.DebitUnits > decimal.MaxExactInteger || settlement.CreditUnits < 0 || settlement.CreditUnits > decimal.MaxExactInteger {
			return nil, ErrInvalidReplayInputs
		}
		lastSettlement = settlement.BoundarySeq
	}
	result := make([]multiplier.Contribution, len(accrual.Contributions))
	lastKey := ""
	for index, item := range accrual.Contributions {
		factor, err := decimal.ParseCanonical(item.Factor)
		key := string(item.Slot) + "\x00" + item.SourceID + "\x00" + item.Target
		if err != nil || !multiplier.ValidSlot(item.Slot) || item.SourceID == "" || item.Target == "" ||
			!factor.IsStateValue() || !factor.Gt(decimal.Zero) || index > 0 && key <= lastKey {
			return nil, ErrInvalidReplayInputs
		}
		lastKey = key
		result[index] = multiplier.Contribution{Slot: item.Slot, SourceID: item.SourceID, Target: item.Target, Factor: factor}
	}
	return result, nil
}

func applyReplayGuildSettlements(state *save.State, encoded replayGuildSettlementBatch, stockCap int64) error {
	if state == nil {
		return ErrInvalidReplayInputs
	}
	batch := guild.SettlementBatch{GuildID: encoded.GuildID, BaseSeq: encoded.BaseSeq, Settlements: make([]guild.Settlement, len(encoded.Settlements))}
	for index, settlement := range encoded.Settlements {
		batch.Settlements[index] = guild.Settlement{GuildID: encoded.GuildID, BoundarySeq: settlement.BoundarySeq,
			DebitUnits: settlement.DebitUnits, CreditUnits: settlement.CreditUnits}
	}
	if err := guild.ApplySettlements(state, batch, stockCap); err != nil {
		return err
	}
	return nil
}

// resolvedSettlementState exists only for the live input-resolution read in
// resolveReplayAccrual. ApplyLogged rejection rollback uses cloneReplayState
// so every transition-owned field, not merely these settlement fields, rolls
// back atomically.
type resolvedSettlementState struct {
	stockUnits          int64
	consumedStockUnits  int64
	consumedWindowUnits int64
	boundaryGuildID     string
	boundarySeq         int64
}

func (before resolvedSettlementState) restore(state *save.State) {
	state.StockUnits = before.stockUnits
	state.ConsumedStockUnits = before.consumedStockUnits
	state.GuildConsumedWindow = before.consumedWindowUnits
	state.GuildBoundaryGuildID = before.boundaryGuildID
	state.GuildBoundarySeq = before.boundarySeq
}

func cloneReplayState(state *save.State, catalog *economy.Catalog) (*save.State, error) {
	encoded, err := save.EncodeState(state)
	if err != nil {
		return nil, err
	}
	cloned, err := save.RestoreState(encoded, save.VersionForState(state), catalog, economy.ScopeCompany, time.Time{})
	if err != nil {
		return nil, err
	}
	cloned.FactionStockResource = state.FactionStockResource
	return cloned, nil
}

func validFounderCarry(carry replayFounderCarry, wireVersion int, catalogs CatalogBundle) bool {
	foundationsActive := catalogs.foundationsActive()
	if carry.FounderRevision < 1 || carry.FounderRevision > decimal.MaxExactInteger ||
		len(carry.FounderConstantsHash) != len("sha256:")+64 || !strings.HasPrefix(carry.FounderConstantsHash, "sha256:") ||
		carry.ReputationLevel < 0 || carry.ReputationLevel > decimal.MaxExactInteger ||
		carry.RouteKnowledgeBalance < 0 || carry.RouteKnowledgeBalance > decimal.MaxExactInteger ||
		carry.AgeMS < 0 || carry.AgeMS > decimal.MaxExactInteger || carry.Notoriety < 0 ||
		carry.Notoriety > decimal.MaxExactInteger || carry.ExitHistoryCount < 0 || int64(carry.ExitHistoryCount) > decimal.MaxExactInteger ||
		carry.AchievementScoreLifetime < 0 || carry.AchievementScoreLifetime > decimal.MaxExactInteger || carry.NetworkSlots == nil ||
		carry.LedgerFactKinds == nil || wireVersion == 2 && (foundationsActive || len(carry.AchievementsEarnedLifetime) != 0 || carry.AchievementScoreLifetime != 0) ||
		wireVersion >= 3 && carry.AchievementsEarnedLifetime == nil {
		return false
	}
	founderFloor, _ := catalogs.versionFloors()
	if wireVersion < 6 {
		if carry.FounderExtensions != nil || founderFloor > 16 {
			return false
		}
	} else if (founderFloor >= 17) != (carry.FounderExtensions != nil) {
		return false
	}
	last := ""
	for _, fact := range carry.LedgerFactKinds {
		if fact <= last {
			return false
		}
		last = fact
	}
	last = ""
	for _, slot := range carry.NetworkSlots {
		if slot.Slot <= last || slot.CarriedRef == "" {
			return false
		}
		last = slot.Slot
	}
	if carry.AchievementsEarnedLifetime == nil {
		return wireVersion == 2 && !foundationsActive && carry.AchievementScoreLifetime == 0
	}
	return sortedUniqueMechanical(carry.AchievementsEarnedLifetime) && (foundationsActive || len(carry.AchievementsEarnedLifetime) == 0 && carry.AchievementScoreLifetime == 0)
}

func sortedUniqueMechanical(values []string) bool {
	if values == nil {
		return false
	}
	last := ""
	for _, value := range values {
		if value <= last {
			return false
		}
		last = value
	}
	return true
}

func stateFromFounderCarry(carry replayFounderCarry, catalogs CatalogBundle) (*save.State, error) {
	facts := make(map[string]bool, len(carry.LedgerFactKinds))
	for _, fact := range carry.LedgerFactKinds {
		facts[fact] = true
	}
	ledger, err := economy.NewLedger(catalogs.Economy, economy.ScopeFounder)
	if err != nil {
		return nil, err
	}
	history := make([]save.ExitRecord, carry.ExitHistoryCount)
	state := &save.State{Ledger: ledger, ReputationLevel: carry.ReputationLevel, RouteKnowledgeBalance: carry.RouteKnowledgeBalance,
		AgeMS: carry.AgeMS, Notoriety: carry.Notoriety, AdvisorMode: carry.AdvisorMode,
		NetworkSlots: append([]save.NetworkSlot(nil), carry.NetworkSlots...), LedgerFactKinds: facts, ExitHistory: history}
	if !catalogs.foundationsActive() {
		if len(carry.AchievementsEarnedLifetime) != 0 || carry.AchievementScoreLifetime != 0 {
			return nil, ErrInvalidReplayInputs
		}
		return state, nil
	}
	founderFloor, _ := catalogs.versionFloors()
	state.WireVersion = founderFloor
	state.AchievementsEarnedLifetime = make(map[string]bool, len(carry.AchievementsEarnedLifetime))
	for _, id := range carry.AchievementsEarnedLifetime {
		state.AchievementsEarnedLifetime[id] = true
	}
	state.AchievementScoreLifetime = carry.AchievementScoreLifetime
	if founderFloor < 17 {
		if carry.FounderExtensions != nil {
			return nil, ErrInvalidReplayInputs
		}
		if err := validateFounderCarryFoundationState(catalogs, state); err != nil {
			return nil, err
		}
		return state, nil
	}
	extensions := carry.FounderExtensions
	if extensions == nil {
		return nil, ErrInvalidReplayInputs
	}
	state.MinigameRatings = cloneMinigameRatingsForReplay(extensions.MinigameRatings)
	state.MinigameOfflineQuality = cloneMinigameQualityForReplay(extensions.MinigameOfflineQuality)
	if founderFloor >= 18 {
		state.Pets = clonePetStatesForReplay(extensions.Pets)
	} else if len(extensions.Pets) != 0 {
		return nil, ErrInvalidReplayInputs
	}
	if founderFloor >= 19 {
		state.FiscalCredit = extensions.FiscalCredit
		state.FiscalPeriodOpenedWallMS = extensions.FiscalPeriodOpenedMS
		state.FiscalPeriodSequence = extensions.FiscalPeriodSequence
		state.FiscalGeneratorLevels = cloneInt64Counts(extensions.FiscalGeneratorLevels)
		state.FiscalUnlocks = boolMapFromSorted(extensions.FiscalUnlocks)
	} else if extensions.FiscalCredit != 0 || extensions.FiscalPeriodOpenedMS != 0 || extensions.FiscalPeriodSequence != 0 ||
		len(extensions.FiscalGeneratorLevels) != 0 || len(extensions.FiscalUnlocks) != 0 {
		return nil, ErrInvalidReplayInputs
	}
	if founderFloor >= 20 {
		state.Soul = extensions.Soul
		state.SoulExhaustedSourceIDs = append([]string{}, extensions.SoulExhaustedSourceIDs...)
	} else if extensions.Soul != 0 || len(extensions.SoulExhaustedSourceIDs) != 0 {
		return nil, ErrInvalidReplayInputs
	}
	if founderFloor >= 21 {
		state.MinigameSessionSeq = extensions.MinigameSessionSeq
	} else if extensions.MinigameSessionSeq != 0 {
		return nil, ErrInvalidReplayInputs
	}
	if err := catalogs.ValidateFoundationState(state); err != nil {
		return nil, err
	}
	return state, nil
}

type replayAccrualResolved struct {
	Kind         string                      `json:"kind"`
	IntentKind   string                      `json:"intent_kind"`
	Accrual      replayAccrual               `json:"accrual"`
	FounderCarry *replayFounderCarry         `json:"founder_carry,omitempty"`
	ActivePlay   *activePlayScheduleEvidence `json:"active_play,omitempty"`
}

type replayCrossGateResolved struct {
	Kind                   string                      `json:"kind"`
	IntentKind             string                      `json:"intent_kind"`
	Accrual                replayAccrual               `json:"accrual"`
	DeclinedExitOfferCount int64                       `json:"declined_exit_offer_count"`
	FounderCarry           *replayFounderCarry         `json:"founder_carry"`
	ActivePlay             *activePlayScheduleEvidence `json:"active_play,omitempty"`
}

type replayExitResolved struct {
	Kind                  string                      `json:"kind"`
	IntentKind            string                      `json:"intent_kind"`
	Accrual               replayAccrual               `json:"accrual"`
	FounderCarry          replayFounderCarry          `json:"founder_carry"`
	ExecutedRouteIDs      []string                    `json:"executed_route_ids"`
	SelectedExitType      string                      `json:"selected_exit_type"`
	SelectedBranch        *string                     `json:"selected_branch,omitempty"`
	SelectedTerms         json.RawMessage             `json:"selected_terms"`
	NextConstantsHash     string                      `json:"next_constants_hash"`
	ActivePlay            *activePlayScheduleEvidence `json:"active_play,omitempty"`
	NextActivePlay        *activePlaySpawnEvidence    `json:"next_active_play,omitempty"`
	MinigameSessionActive *bool                       `json:"minigame_session_active,omitempty"`
}

type replayBuild struct {
	Command                save.ReplayCommand
	Mode                   EvaluationMode
	Now                    time.Time
	IntentKind             string
	Contributions          []multiplier.Contribution
	CommonsWeightPPM       *int64
	GuildSettlementBatch   guild.SettlementBatch
	RouteContextVersion    int
	DeclinedExitOfferCount int64
	FounderCarry           *replayFounderCarry
	Terminal               bool
	ExecutedRouteIDs       []string
	SelectedExitType       string
	SelectedBranch         *string
	SelectedTerms          json.RawMessage
	NextConstantsHash      string
	ActivePlay             *activePlayScheduleEvidence
	NextActivePlay         *activePlaySpawnEvidence
	MinigameSessionActive  *bool
	OfflineCatchup         *replayOfflineCatchup
}

func buildReplayInputs(input replayBuild) (json.RawMessage, error) {
	if input.Command.RunLogSeq == 0 {
		return nil, nil
	}
	if input.IntentKind == IntentBuyRouteHint {
		return nil, ErrInvalidReplayInputs
	}
	if input.Mode != ModeOnline && input.Mode != ModeOffline {
		return nil, ErrInvalidReplayInputs
	}
	accrual, err := makeReplayAccrual(input.Contributions, input.CommonsWeightPPM, input.GuildSettlementBatch, input.RouteContextVersion)
	if err != nil {
		return nil, err
	}
	var resolved []byte
	switch {
	case input.Terminal:
		if input.FounderCarry == nil || input.SelectedExitType == "" || len(input.SelectedTerms) == 0 || input.NextConstantsHash == "" {
			return nil, ErrInvalidReplayInputs
		}
		routes := append([]string(nil), input.ExecutedRouteIDs...)
		if routes == nil {
			routes = []string{}
		}
		sort.Strings(routes)
		carry := normalizedFounderCarry(*input.FounderCarry)
		resolved, err = json.Marshal(replayExitResolved{Kind: "exit", IntentKind: input.IntentKind, Accrual: accrual,
			FounderCarry: carry, ExecutedRouteIDs: routes, SelectedExitType: input.SelectedExitType,
			SelectedBranch: input.SelectedBranch,
			SelectedTerms:  append(json.RawMessage(nil), input.SelectedTerms...), NextConstantsHash: input.NextConstantsHash,
			ActivePlay: input.ActivePlay, NextActivePlay: input.NextActivePlay,
			MinigameSessionActive: input.MinigameSessionActive})
	case input.IntentKind == IntentCrossGate:
		var carry *replayFounderCarry
		if input.FounderCarry != nil {
			value := normalizedFounderCarry(*input.FounderCarry)
			carry = &value
		}
		resolved, err = json.Marshal(replayCrossGateResolved{Kind: "cross_gate", IntentKind: input.IntentKind, Accrual: accrual,
			DeclinedExitOfferCount: input.DeclinedExitOfferCount, FounderCarry: carry, ActivePlay: input.ActivePlay})
	default:
		var carry *replayFounderCarry
		if input.FounderCarry != nil {
			value := normalizedFounderCarry(*input.FounderCarry)
			carry = &value
		}
		resolved, err = json.Marshal(replayAccrualResolved{Kind: "accrual", IntentKind: input.IntentKind, Accrual: accrual, FounderCarry: carry, ActivePlay: input.ActivePlay})
	}
	if err != nil {
		return nil, err
	}
	wire, err := json.Marshal(replayInputsWire{Version: save.ReplayInputsVersion, Command: input.Command,
		EvaluatedAtMS: save.CanonicalServerTime(input.Now).UnixMilli(), EvaluationMode: input.Mode,
		OfflineCatchup: input.OfflineCatchup, Resolved: resolved})
	if err != nil {
		return nil, err
	}
	if _, err := parseReplayInputs(wire); err != nil {
		return nil, err
	}
	return wire, nil
}

func buildOfflineCatchup(state *save.State, catalog *economy.Catalog, mode EvaluationMode, now time.Time) (*replayOfflineCatchup, error) {
	if state == nil || catalog == nil || state.EvaluatedThrough.IsZero() || mode != ModeOnline {
		return nil, nil
	}
	opened := save.CanonicalServerTime(now)
	if !opened.After(state.EvaluatedThrough) {
		return nil, nil
	}
	capMS := catalog.OfflinePolicy().AccrualCapMS
	if capMS <= 0 {
		return nil, ErrInvalidEngineState
	}
	if opened.Sub(state.EvaluatedThrough).Milliseconds() <= capMS {
		return nil, nil
	}
	return &replayOfflineCatchup{OpenedAtMS: opened.UnixMilli(), OfflineSpan: replayOfflineSpan{
		FromMS: state.EvaluatedThrough.UnixMilli(), ToMS: opened.UnixMilli(),
	}}, nil
}

func applyReplayOfflineCatchup(state *save.State, catalog *economy.Catalog, wire replayInputsWire,
	contributions []multiplier.Contribution, hook AccrualHook, revision save.Revision, intentID string,
) ([]save.EventWrite, error) {
	needed, err := buildOfflineCatchup(state, catalog, wire.EvaluationMode, time.UnixMilli(wire.EvaluatedAtMS).UTC())
	if err != nil {
		return nil, err
	}
	if wire.Version < 7 {
		if wire.OfflineCatchup != nil {
			return nil, fmt.Errorf("%w: legacy offline catchup", ErrInvalidReplayInputs)
		}
		return []save.EventWrite{}, nil
	}
	if (needed == nil) != (wire.OfflineCatchup == nil) {
		return nil, fmt.Errorf("%w: offline catchup presence", ErrInvalidReplayInputs)
	}
	if needed == nil {
		return []save.EventWrite{}, nil
	}
	if *needed != *wire.OfflineCatchup || wire.OfflineCatchup.OpenedAtMS != wire.EvaluatedAtMS ||
		wire.OfflineCatchup.OfflineSpan.FromMS >= wire.OfflineCatchup.OfflineSpan.ToMS {
		return nil, fmt.Errorf("%w: offline catchup coordinates", ErrInvalidReplayInputs)
	}
	evaluation, err := Evaluate(state, catalog, time.UnixMilli(wire.OfflineCatchup.OpenedAtMS).UTC(), ModeOffline, contributions)
	if err != nil {
		return nil, err
	}
	events, err := runAccrualHook(hook, intentID, state, catalog, revision, evaluation, contributions)
	if err != nil {
		return nil, err
	}
	return events, nil
}

func normalizedFounderCarry(carry replayFounderCarry) replayFounderCarry {
	if carry.AchievementsEarnedLifetime == nil {
		carry.AchievementsEarnedLifetime = []string{}
	}
	if carry.FounderExtensions != nil {
		carry.FounderExtensions = cloneFounderExtensions(carry.FounderExtensions)
	}
	return carry
}

func makeReplayAccrual(contributions []multiplier.Contribution, weight *int64, settlements guild.SettlementBatch, routeContextVersion int) (replayAccrual, error) {
	if routeContextVersion < 0 || weight != nil && (*weight < 0 || *weight > 1_000_000) {
		return replayAccrual{}, ErrInvalidReplayInputs
	}
	result := replayAccrual{Contributions: make([]replayContribution, len(contributions)), CommonsWeightPPM: weight,
		GuildSettlementBatch: replayGuildSettlementBatch{GuildID: settlements.GuildID, BaseSeq: settlements.BaseSeq, Settlements: make([]replayGuildSettlement, len(settlements.Settlements))}, RouteContextVersion: routeContextVersion}
	for index, settlement := range settlements.Settlements {
		if settlement.GuildID != settlements.GuildID {
			return replayAccrual{}, ErrInvalidReplayInputs
		}
		result.GuildSettlementBatch.Settlements[index] = replayGuildSettlement{BoundarySeq: settlement.BoundarySeq, DebitUnits: settlement.DebitUnits, CreditUnits: settlement.CreditUnits}
	}
	for index, contribution := range contributions {
		if !multiplier.ValidSlot(contribution.Slot) || contribution.SourceID == "" || contribution.Target == "" ||
			!contribution.Factor.IsStateValue() || !contribution.Factor.Gt(decimal.Zero) {
			return replayAccrual{}, ErrInvalidReplayInputs
		}
		result.Contributions[index] = replayContribution{Slot: contribution.Slot, SourceID: contribution.SourceID, Target: contribution.Target, Factor: contribution.Factor.String()}
	}
	sort.Slice(result.Contributions, func(left, right int) bool {
		if result.Contributions[left].Slot != result.Contributions[right].Slot {
			return result.Contributions[left].Slot < result.Contributions[right].Slot
		}
		if result.Contributions[left].SourceID != result.Contributions[right].SourceID {
			return result.Contributions[left].SourceID < result.Contributions[right].SourceID
		}
		return result.Contributions[left].Target < result.Contributions[right].Target
	})
	if _, err := contributionsFromReplay(result); err != nil {
		return replayAccrual{}, err
	}
	return result, nil
}

func founderCarry(state *save.State) replayFounderCarry {
	facts := make([]string, 0, len(state.LedgerFactKinds))
	for key, present := range state.LedgerFactKinds {
		if present {
			facts = append(facts, key)
		}
	}
	sort.Strings(facts)
	slots := append([]save.NetworkSlot(nil), state.NetworkSlots...)
	if slots == nil {
		slots = []save.NetworkSlot{}
	}
	sort.Slice(slots, func(left, right int) bool { return slots[left].Slot < slots[right].Slot })
	earned := sortedBoolKeys(state.AchievementsEarnedLifetime)
	if earned == nil {
		earned = []string{}
	}
	result := replayFounderCarry{ReputationLevel: state.ReputationLevel, RouteKnowledgeBalance: state.RouteKnowledgeBalance,
		AgeMS: state.AgeMS, Notoriety: state.Notoriety, AdvisorMode: state.AdvisorMode, NetworkSlots: slots,
		LedgerFactKinds: facts, ExitHistoryCount: len(state.ExitHistory), AchievementsEarnedLifetime: earned,
		AchievementScoreLifetime: state.AchievementScoreLifetime}
	if save.VersionForState(state) >= 17 {
		unlocks := sortedBoolKeys(state.FiscalUnlocks)
		if unlocks == nil {
			unlocks = []string{}
		}
		exhausted := append([]string{}, state.SoulExhaustedSourceIDs...)
		if exhausted == nil {
			exhausted = []string{}
		}
		extensions := &replayFounderExtensions{
			MinigameRatings: cloneMinigameRatingsForReplay(state.MinigameRatings), MinigameOfflineQuality: cloneMinigameQualityForReplay(state.MinigameOfflineQuality),
			Pets: clonePetStatesForReplay(state.Pets), FiscalCredit: state.FiscalCredit, FiscalPeriodOpenedMS: state.FiscalPeriodOpenedWallMS,
			FiscalPeriodSequence: state.FiscalPeriodSequence, FiscalGeneratorLevels: cloneInt64Counts(state.FiscalGeneratorLevels), FiscalUnlocks: unlocks,
			Soul: state.Soul, SoulExhaustedSourceIDs: exhausted, MinigameSessionSeq: state.MinigameSessionSeq,
		}
		if extensions.MinigameRatings == nil {
			extensions.MinigameRatings = map[string]save.MinigameRatingState{}
		}
		if extensions.MinigameOfflineQuality == nil {
			extensions.MinigameOfflineQuality = map[string]save.MinigameOfflineQualityState{}
		}
		if extensions.Pets == nil {
			extensions.Pets = map[string]pet.CareState{}
		}
		if extensions.FiscalGeneratorLevels == nil {
			extensions.FiscalGeneratorLevels = map[string]int64{}
		}
		result.FounderExtensions = extensions
	}
	return result
}

func cloneFounderExtensions(source *replayFounderExtensions) *replayFounderExtensions {
	if source == nil {
		return nil
	}
	return &replayFounderExtensions{
		MinigameRatings: cloneMinigameRatingsForReplay(source.MinigameRatings), MinigameOfflineQuality: cloneMinigameQualityForReplay(source.MinigameOfflineQuality),
		Pets: clonePetStatesForReplay(source.Pets), FiscalCredit: source.FiscalCredit, FiscalPeriodOpenedMS: source.FiscalPeriodOpenedMS,
		FiscalPeriodSequence: source.FiscalPeriodSequence, FiscalGeneratorLevels: cloneInt64Counts(source.FiscalGeneratorLevels),
		FiscalUnlocks: append([]string{}, source.FiscalUnlocks...), Soul: source.Soul,
		SoulExhaustedSourceIDs: append([]string{}, source.SoulExhaustedSourceIDs...), MinigameSessionSeq: source.MinigameSessionSeq,
	}
}

func cloneMinigameRatingsForReplay(source map[string]save.MinigameRatingState) map[string]save.MinigameRatingState {
	if source == nil {
		return nil
	}
	result := make(map[string]save.MinigameRatingState, len(source))
	for id, value := range source {
		result[id] = value
	}
	return result
}

func cloneMinigameQualityForReplay(source map[string]save.MinigameOfflineQualityState) map[string]save.MinigameOfflineQualityState {
	if source == nil {
		return nil
	}
	result := make(map[string]save.MinigameOfflineQualityState, len(source))
	for id, value := range source {
		result[id] = value
	}
	return result
}

func clonePetStatesForReplay(source map[string]pet.CareState) map[string]pet.CareState {
	if source == nil {
		return nil
	}
	result := make(map[string]pet.CareState, len(source))
	for id, value := range source {
		stats := make(map[pet.StatID]int64, len(value.StatsPPM))
		for statID, amount := range value.StatsPPM {
			stats[statID] = amount
		}
		remainders := make(map[pet.StatID]int64, len(value.StatDecayRemaindersPPM))
		for statID, amount := range value.StatDecayRemaindersPPM {
			remainders[statID] = amount
		}
		cooldowns := make(map[string]int64, len(value.CooldownUntilAttendedMS))
		for actionID, cursor := range value.CooldownUntilAttendedMS {
			cooldowns[actionID] = cursor
		}
		value.StatsPPM, value.StatDecayRemaindersPPM, value.CooldownUntilAttendedMS = stats, remainders, cooldowns
		value.BehaviorQueue = append([]pet.BehaviorQueueEntry{}, value.BehaviorQueue...)
		result[id] = value
	}
	return result
}

func boolMapFromSorted(values []string) map[string]bool {
	if values == nil {
		return nil
	}
	result := make(map[string]bool, len(values))
	for _, value := range values {
		result[value] = true
	}
	return result
}

func parseReplayInputs(data []byte) (replayInputsWire, error) {
	var wire replayInputsWire
	if err := decodeReplayStrict(data, &wire); err != nil || (wire.Version != 2 && wire.Version != 3 && wire.Version != 4 && wire.Version != 5 && wire.Version != 6 && wire.Version != 7 && wire.Version != save.ReplayInputsVersion) ||
		(wire.EvaluationMode != ModeOnline && wire.EvaluationMode != ModeOffline) || wire.EvaluatedAtMS <= 0 {
		return replayInputsWire{}, ErrInvalidReplayInputs
	}
	if wire.Version < 7 && wire.OfflineCatchup != nil || wire.OfflineCatchup != nil &&
		(wire.OfflineCatchup.OpenedAtMS <= 0 || wire.OfflineCatchup.OpenedAtMS != wire.EvaluatedAtMS ||
			wire.OfflineCatchup.OfflineSpan.FromMS <= 0 || wire.OfflineCatchup.OfflineSpan.ToMS != wire.EvaluatedAtMS ||
			wire.OfflineCatchup.OfflineSpan.FromMS >= wire.OfflineCatchup.OfflineSpan.ToMS) {
		return replayInputsWire{}, ErrInvalidReplayInputs
	}
	var discriminator struct {
		Kind string `json:"kind"`
	}
	if err := json.Unmarshal(wire.Resolved, &discriminator); err != nil {
		return replayInputsWire{}, ErrInvalidReplayInputs
	}
	switch discriminator.Kind {
	case "accrual":
		var value replayAccrualResolved
		if err := decodeReplayStrict(wire.Resolved, &value); err != nil || value.IntentKind == "" {
			return replayInputsWire{}, ErrInvalidReplayInputs
		}
	case "cross_gate":
		var value replayCrossGateResolved
		if err := decodeReplayStrict(wire.Resolved, &value); err != nil || value.IntentKind != IntentCrossGate || value.DeclinedExitOfferCount < 0 {
			return replayInputsWire{}, ErrInvalidReplayInputs
		}
	case "exit":
		var value replayExitResolved
		if err := decodeReplayStrict(wire.Resolved, &value); err != nil || value.IntentKind == "" || value.SelectedExitType == "" ||
			!jsonObjectValue(value.SelectedTerms) || len(value.NextConstantsHash) != len("sha256:")+64 || !strings.HasPrefix(value.NextConstantsHash, "sha256:") {
			return replayInputsWire{}, ErrInvalidReplayInputs
		}
	case "soul_recovery_suppression":
		var value soulSuppressionResolved
		if err := decodeReplayStrict(wire.Resolved, &value); err != nil ||
			(value.IntentKind != soulRecoveryResolveKind && value.IntentKind != soulRecoveryCancelKind) {
			return replayInputsWire{}, ErrInvalidReplayInputs
		}
	default:
		return replayInputsWire{}, ErrInvalidReplayInputs
	}
	return wire, nil
}

func deriveFactionStockResource(state *save.State, catalog *faction.Catalog) error {
	if state.FactionID == "" {
		if state.FactionStockResource != "" {
			return ErrInvalidEngineState
		}
		return nil
	}
	member, ok := catalog.Faction(state.FactionID)
	if !ok {
		return ErrInvalidEngineState
	}
	state.FactionStockResource = member.Produces
	return nil
}

func jsonObjectValue(data []byte) bool {
	var value map[string]json.RawMessage
	return decodeReplayStrict(data, &value) == nil && value != nil
}

func decodeReplayStrict(data []byte, target any) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			return ErrInvalidReplayInputs
		}
		return err
	}
	return nil
}

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
	"cloud-clicker/server/decimal"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/faction"
	"cloud-clicker/server/guild"
	"cloud-clicker/server/multiplier"
	prestigecore "cloud-clicker/server/prestige"
	"cloud-clicker/server/routes"
	"cloud-clicker/server/save"
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

func (bundle CatalogBundle) valid(constantsHash string) bool {
	if constantsHash == "" || bundle.ConstantsHash != constantsHash || len(bundle.Artifacts) != 7 || bundle.Economy == nil ||
		bundle.Routes == nil || bundle.Commons == nil || bundle.Prestige == nil || bundle.Faction == nil || bundle.Guild == nil {
		return false
	}
	for _, name := range [...]string{"categories", "commons", "economy", "factions", "guilds", "prestige", "routes"} {
		if len(bundle.Artifacts[name]) == 0 {
			return false
		}
	}
	computed, err := save.ConstantsHashArtifacts(bundle.Artifacts)
	return err == nil && computed == constantsHash
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
	FounderRevision       int64              `json:"founder_revision"`
	FounderConstantsHash  string             `json:"founder_constants_hash"`
	ReputationLevel       int64              `json:"reputation_level"`
	RouteKnowledgeBalance int64              `json:"route_knowledge_balance"`
	AgeMS                 int64              `json:"age_ms"`
	Notoriety             int64              `json:"notoriety"`
	AdvisorMode           bool               `json:"advisor_mode"`
	NetworkSlots          []save.NetworkSlot `json:"network_slots"`
	LedgerFactKinds       []string           `json:"ledger_fact_kinds"`
	ExitHistoryCount      int                `json:"exit_history_count"`
}

type replayInputsWire struct {
	Version        int                `json:"v"`
	Command        save.ReplayCommand `json:"command"`
	EvaluatedAtMS  int64              `json:"evaluated_at_ms"`
	EvaluationMode EvaluationMode     `json:"evaluation_mode"`
	Resolved       json.RawMessage    `json:"resolved"`
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
	if request.Kind == IntentCrossGate {
		var resolved replayCrossGateResolved
		if err := decodeReplayStrict(wire.Resolved, &resolved); err != nil || resolved.IntentKind != request.Kind {
			return LoggedTransition{}, fmt.Errorf("%w: cross-gate resolved union", ErrInvalidReplayInputs)
		}
		accrual, declined = resolved.Accrual, resolved.DeclinedExitOfferCount
		if resolved.FounderCarry != nil {
			if !validFounderCarry(*resolved.FounderCarry) || resolved.FounderCarry.FounderConstantsHash != catalogs.ConstantsHash {
				return LoggedTransition{}, fmt.Errorf("%w: founder carry", ErrInvalidReplayInputs)
			}
			founder = stateFromFounderCarry(*resolved.FounderCarry)
		}
	} else {
		var resolved replayAccrualResolved
		if err := decodeReplayStrict(wire.Resolved, &resolved); err != nil || resolved.IntentKind != request.Kind {
			return LoggedTransition{}, fmt.Errorf("%w: accrual resolved union", ErrInvalidReplayInputs)
		}
		accrual = resolved.Accrual
	}
	contributions, err := contributionsFromReplay(accrual)
	if err != nil || accrual.RouteContextVersion != catalogs.Routes.ContextVersion() {
		return LoggedTransition{}, fmt.Errorf("%w: accrual inputs", ErrInvalidReplayInputs)
	}
	settlementBefore, err := applyReplayGuildSettlements(state, accrual.GuildSettlementBatch, catalogs.Faction.StockCap)
	if err != nil {
		return LoggedTransition{}, fmt.Errorf("%w: guild settlement inputs", ErrInvalidReplayInputs)
	}
	defer func() {
		if resultErr != nil || result.Outcome != save.IntentApplied {
			settlementBefore.restore(state)
		}
	}()
	if state.CompactMember != (accrual.CommonsWeightPPM != nil) {
		return LoggedTransition{}, fmt.Errorf("%w: commons weight presence", ErrInvalidReplayInputs)
	}
	hook := closedReplayAccrualHook(catalogs, accrual.CommonsWeightPPM)
	band := &CompactTitheBand{MinimumPPM: catalogs.Commons.MinimumTithePPM(), MaximumPPM: catalogs.Commons.MaximumTithePPM()}
	decision, err := TransitionWithPolicies(request, state, catalogs.Economy, catalogs.Routes, band, catalogs.Faction,
		revision, wire.EvaluationMode, now, contributions, collector, hook)
	if err != nil {
		return LoggedTransition{}, err
	}
	if decision.Outcome == save.IntentApplied {
		if err := afterPrestigeTransitionResolved(catalogs.Prestige, request, state, revision, now, &decision, founder, declined); err != nil {
			return LoggedTransition{}, err
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
	now := time.UnixMilli(wire.EvaluatedAtMS).UTC()
	if now.Before(company.EvaluatedThrough) {
		return LoggedExitTransition{}, ErrReplayClockViolation
	}
	contributions, err := contributionsFromReplay(resolved.Accrual)
	if err != nil || resolved.Accrual.RouteContextVersion != catalogs.Routes.ContextVersion() {
		return LoggedExitTransition{}, fmt.Errorf("%w: terminal accrual inputs", ErrInvalidReplayInputs)
	}
	if company.CompactMember != (resolved.Accrual.CommonsWeightPPM != nil) || !validFounderCarry(resolved.FounderCarry) ||
		resolved.FounderCarry.FounderConstantsHash != catalogs.ConstantsHash || !sortedUniqueMechanical(resolved.ExecutedRouteIDs) {
		return LoggedExitTransition{}, fmt.Errorf("%w: terminal frozen inputs", ErrInvalidReplayInputs)
	}
	settlementBefore, err := applyReplayGuildSettlements(company, resolved.Accrual.GuildSettlementBatch, catalogs.Faction.StockCap)
	if err != nil {
		return LoggedExitTransition{}, fmt.Errorf("%w: terminal guild settlement inputs", ErrInvalidReplayInputs)
	}
	defer func() {
		if resultErr != nil || result.Decision.Outcome != save.IntentApplied {
			settlementBefore.restore(company)
		}
	}()
	revision := save.Revision{StreamID: wire.Command.CompanyStreamID, OwnerID: wire.Command.FounderID,
		Number: wire.Command.Revision, ConstantsHash: catalogs.ConstantsHash, RunLogSequence: wire.Command.RunLogSeq}
	founder := stateFromFounderCarry(resolved.FounderCarry)
	hook := closedReplayAccrualHook(catalogs, resolved.Accrual.CommonsWeightPPM)
	var prefix []save.EventWrite
	var exitType string
	var terms prestigecore.Terms
	if request.Kind == IntentCrossGate {
		transition, transitionErr := TransitionWithPolicies(request, company, catalogs.Economy, catalogs.Routes, nil, nil,
			revision, wire.EvaluationMode, now, contributions, collector, hook)
		if transitionErr != nil {
			return LoggedExitTransition{}, transitionErr
		}
		if transition.Outcome == save.IntentRejected {
			return LoggedExitTransition{Founder: founder, Company: company, Decision: save.ExitDecision{Outcome: save.IntentRejected, Receipt: transition.Receipt}}, nil
		}
		attended, attendedErr := prestigecore.AttendedMS(company, save.CanonicalServerTime(now))
		if attendedErr != nil || attended < 900_000 || len(founder.ExitHistory) != 0 {
			return LoggedExitTransition{}, ErrInvalidEngineState
		}
		exitType, prefix = "scripted_first", transition.Events
		terms, err = prestigecore.ComputeTerms(company, founder, catalogs.Prestige, exitType)
	} else {
		if request.Kind == IntentFileIPO {
			decision := rejectedExitDecision(request, revision.Number, "not_eligible", "ipo_chain")
			return LoggedExitTransition{Founder: founder, Company: company, Decision: decision}, nil
		}
		if request.Kind == IntentWindDown && company.Tier < 1 {
			decision := rejectedExitDecision(request, revision.Number, "not_eligible", "tier")
			return LoggedExitTransition{Founder: founder, Company: company, Decision: decision}, nil
		}
		exitType = "collapse"
		if request.Kind == IntentWindDown && len(founder.ExitHistory) == 0 {
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
	founderRevision := save.Revision{StreamID: "", OwnerID: wire.Command.FounderID, Number: resolved.FounderCarry.FounderRevision,
		ConstantsHash: catalogs.ConstantsHash}
	decision, err := finishExitResolved(request, founder, founderRevision, company, revision, now, exitType, terms, prefix,
		resolved.ExecutedRouteIDs, next.Economy, resolved.NextConstantsHash)
	if err != nil {
		return LoggedExitTransition{}, err
	}
	return LoggedExitTransition{Founder: founder, Company: company, Decision: decision}, nil
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

type replaySettlementState struct {
	stockUnits          int64
	consumedStockUnits  int64
	consumedWindowUnits int64
	boundaryGuildID     string
	boundarySeq         int64
}

func applyReplayGuildSettlements(state *save.State, encoded replayGuildSettlementBatch, stockCap int64) (replaySettlementState, error) {
	if state == nil {
		return replaySettlementState{}, ErrInvalidReplayInputs
	}
	before := replaySettlementState{stockUnits: state.StockUnits, consumedStockUnits: state.ConsumedStockUnits,
		consumedWindowUnits: state.GuildConsumedWindow, boundaryGuildID: state.GuildBoundaryGuildID, boundarySeq: state.GuildBoundarySeq}
	batch := guild.SettlementBatch{GuildID: encoded.GuildID, BaseSeq: encoded.BaseSeq, Settlements: make([]guild.Settlement, len(encoded.Settlements))}
	for index, settlement := range encoded.Settlements {
		batch.Settlements[index] = guild.Settlement{GuildID: encoded.GuildID, BoundarySeq: settlement.BoundarySeq,
			DebitUnits: settlement.DebitUnits, CreditUnits: settlement.CreditUnits}
	}
	if err := guild.ApplySettlements(state, batch, stockCap); err != nil {
		return replaySettlementState{}, err
	}
	return before, nil
}

func (before replaySettlementState) restore(state *save.State) {
	state.StockUnits = before.stockUnits
	state.ConsumedStockUnits = before.consumedStockUnits
	state.GuildConsumedWindow = before.consumedWindowUnits
	state.GuildBoundaryGuildID = before.boundaryGuildID
	state.GuildBoundarySeq = before.boundarySeq
}

func validFounderCarry(carry replayFounderCarry) bool {
	if carry.FounderRevision < 1 || carry.FounderRevision > decimal.MaxExactInteger ||
		len(carry.FounderConstantsHash) != len("sha256:")+64 || !strings.HasPrefix(carry.FounderConstantsHash, "sha256:") ||
		carry.ReputationLevel < 0 || carry.ReputationLevel > decimal.MaxExactInteger ||
		carry.RouteKnowledgeBalance < 0 || carry.RouteKnowledgeBalance > decimal.MaxExactInteger ||
		carry.AgeMS < 0 || carry.AgeMS > decimal.MaxExactInteger || carry.Notoriety < 0 ||
		carry.Notoriety > decimal.MaxExactInteger || carry.ExitHistoryCount < 0 || int64(carry.ExitHistoryCount) > decimal.MaxExactInteger || carry.NetworkSlots == nil || carry.LedgerFactKinds == nil {
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
	return true
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

func stateFromFounderCarry(carry replayFounderCarry) *save.State {
	facts := make(map[string]bool, len(carry.LedgerFactKinds))
	for _, fact := range carry.LedgerFactKinds {
		facts[fact] = true
	}
	history := make([]save.ExitRecord, carry.ExitHistoryCount)
	return &save.State{ReputationLevel: carry.ReputationLevel, RouteKnowledgeBalance: carry.RouteKnowledgeBalance,
		AgeMS: carry.AgeMS, Notoriety: carry.Notoriety, AdvisorMode: carry.AdvisorMode,
		NetworkSlots: append([]save.NetworkSlot(nil), carry.NetworkSlots...), LedgerFactKinds: facts, ExitHistory: history}
}

type replayAccrualResolved struct {
	Kind       string        `json:"kind"`
	IntentKind string        `json:"intent_kind"`
	Accrual    replayAccrual `json:"accrual"`
}

type replayCrossGateResolved struct {
	Kind                   string              `json:"kind"`
	IntentKind             string              `json:"intent_kind"`
	Accrual                replayAccrual       `json:"accrual"`
	DeclinedExitOfferCount int64               `json:"declined_exit_offer_count"`
	FounderCarry           *replayFounderCarry `json:"founder_carry"`
}

type replayExitResolved struct {
	Kind              string             `json:"kind"`
	IntentKind        string             `json:"intent_kind"`
	Accrual           replayAccrual      `json:"accrual"`
	FounderCarry      replayFounderCarry `json:"founder_carry"`
	ExecutedRouteIDs  []string           `json:"executed_route_ids"`
	SelectedExitType  string             `json:"selected_exit_type"`
	SelectedTerms     json.RawMessage    `json:"selected_terms"`
	NextConstantsHash string             `json:"next_constants_hash"`
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
	SelectedTerms          json.RawMessage
	NextConstantsHash      string
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
		resolved, err = json.Marshal(replayExitResolved{Kind: "exit", IntentKind: input.IntentKind, Accrual: accrual,
			FounderCarry: *input.FounderCarry, ExecutedRouteIDs: routes, SelectedExitType: input.SelectedExitType,
			SelectedTerms: append(json.RawMessage(nil), input.SelectedTerms...), NextConstantsHash: input.NextConstantsHash})
	case input.IntentKind == IntentCrossGate:
		var carry *replayFounderCarry
		if input.FounderCarry != nil {
			value := *input.FounderCarry
			carry = &value
		}
		resolved, err = json.Marshal(replayCrossGateResolved{Kind: "cross_gate", IntentKind: input.IntentKind, Accrual: accrual,
			DeclinedExitOfferCount: input.DeclinedExitOfferCount, FounderCarry: carry})
	default:
		resolved, err = json.Marshal(replayAccrualResolved{Kind: "accrual", IntentKind: input.IntentKind, Accrual: accrual})
	}
	if err != nil {
		return nil, err
	}
	wire, err := json.Marshal(replayInputsWire{Version: save.ReplayInputsVersion, Command: input.Command,
		EvaluatedAtMS: save.CanonicalServerTime(input.Now).UnixMilli(), EvaluationMode: input.Mode, Resolved: resolved})
	if err != nil {
		return nil, err
	}
	if _, err := parseReplayInputs(wire); err != nil {
		return nil, err
	}
	return wire, nil
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
	return replayFounderCarry{ReputationLevel: state.ReputationLevel, RouteKnowledgeBalance: state.RouteKnowledgeBalance,
		AgeMS: state.AgeMS, Notoriety: state.Notoriety, AdvisorMode: state.AdvisorMode, NetworkSlots: slots,
		LedgerFactKinds: facts, ExitHistoryCount: len(state.ExitHistory)}
}

func parseReplayInputs(data []byte) (replayInputsWire, error) {
	var wire replayInputsWire
	if err := decodeReplayStrict(data, &wire); err != nil || wire.Version != save.ReplayInputsVersion ||
		(wire.EvaluationMode != ModeOnline && wire.EvaluationMode != ModeOffline) || wire.EvaluatedAtMS <= 0 {
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

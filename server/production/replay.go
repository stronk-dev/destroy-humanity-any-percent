package production

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"sort"
	"time"

	"cloud-clicker/server/decimal"
	"cloud-clicker/server/multiplier"
	"cloud-clicker/server/save"
)

var ErrInvalidReplayInputs = errors.New("invalid replay inputs")

type replayContribution struct {
	Slot     multiplier.Slot `json:"slot"`
	SourceID string          `json:"source_id"`
	Target   string          `json:"target"`
	Factor   string          `json:"factor"`
}

type replayGuildSettlement struct {
	GuildID     string `json:"guild_id"`
	BoundarySeq int64  `json:"boundary_seq"`
	Units       int64  `json:"units"`
}

type replayAccrual struct {
	Contributions        []replayContribution    `json:"contributions"`
	CommonsWeightPPM     *int64                  `json:"commons_weight_ppm"`
	GuildSettlementBatch []replayGuildSettlement `json:"guild_settlement_batch"`
	RouteContextVersion  int                     `json:"route_context_version"`
}

type replayFounderCarry struct {
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

type replayAccrualResolved struct {
	Kind       string        `json:"kind"`
	IntentKind string        `json:"intent_kind"`
	Accrual    replayAccrual `json:"accrual"`
}

type replayRouteHintResolved struct {
	Kind                string `json:"kind"`
	IntentKind          string `json:"intent_kind"`
	RouteContextVersion int    `json:"route_context_version"`
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
	if input.Mode != ModeOnline && input.Mode != ModeOffline {
		return nil, ErrInvalidReplayInputs
	}
	accrual, err := makeReplayAccrual(input.Contributions, input.CommonsWeightPPM, input.RouteContextVersion)
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
		sort.Strings(routes)
		resolved, err = json.Marshal(replayExitResolved{Kind: "exit", IntentKind: input.IntentKind, Accrual: accrual,
			FounderCarry: *input.FounderCarry, ExecutedRouteIDs: routes, SelectedExitType: input.SelectedExitType,
			SelectedTerms: append(json.RawMessage(nil), input.SelectedTerms...), NextConstantsHash: input.NextConstantsHash})
	case input.IntentKind == IntentBuyRouteHint:
		resolved, err = json.Marshal(replayRouteHintResolved{Kind: "route_hint", IntentKind: input.IntentKind, RouteContextVersion: input.RouteContextVersion})
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

func makeReplayAccrual(contributions []multiplier.Contribution, weight *int64, routeContextVersion int) (replayAccrual, error) {
	if routeContextVersion < 0 || weight != nil && (*weight < 0 || *weight > 1_000_000) {
		return replayAccrual{}, ErrInvalidReplayInputs
	}
	result := replayAccrual{Contributions: make([]replayContribution, len(contributions)), CommonsWeightPPM: weight,
		GuildSettlementBatch: []replayGuildSettlement{}, RouteContextVersion: routeContextVersion}
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
	case "route_hint":
		var value replayRouteHintResolved
		if err := decodeReplayStrict(wire.Resolved, &value); err != nil || value.IntentKind != IntentBuyRouteHint || value.RouteContextVersion < 0 {
			return replayInputsWire{}, ErrInvalidReplayInputs
		}
	case "cross_gate":
		var value replayCrossGateResolved
		if err := decodeReplayStrict(wire.Resolved, &value); err != nil || value.IntentKind != IntentCrossGate || value.DeclinedExitOfferCount < 0 {
			return replayInputsWire{}, ErrInvalidReplayInputs
		}
	case "exit":
		var value replayExitResolved
		if err := decodeReplayStrict(wire.Resolved, &value); err != nil || value.IntentKind == "" || value.SelectedExitType == "" || len(value.SelectedTerms) == 0 {
			return replayInputsWire{}, ErrInvalidReplayInputs
		}
	default:
		return replayInputsWire{}, ErrInvalidReplayInputs
	}
	return wire, nil
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

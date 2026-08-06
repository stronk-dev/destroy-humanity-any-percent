package production

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"

	"cloud-clicker/server/decimal"
	"cloud-clicker/server/save"
)

const founderExitResolvedKind = "exit.v1"

type founderExitRecordWire struct {
	RunID           int64  `json:"run_id"`
	ExitType        string `json:"exit_type"`
	OccurredAtMS    int64  `json:"occurred_at_ms"`
	ReputationDelta int64  `json:"reputation_delta"`
}

type founderExitResolvedWire struct {
	Kind                      string                 `json:"kind"`
	Outcome                   string                 `json:"outcome"`
	CompanyStreamID           string                 `json:"company_stream_id"`
	RunSeq                    int64                  `json:"run_seq"`
	RunLogSeq                 int64                  `json:"run_log_seq"`
	ResultConstantsHash       string                 `json:"result_constants_hash"`
	ReputationDelta           int64                  `json:"reputation_delta"`
	RouteKnowledgeDelta       int64                  `json:"route_knowledge_delta"`
	AttendedMS                int64                  `json:"attended_ms"`
	AgeMSBefore               int64                  `json:"age_ms_before"`
	AgeMSAfter                int64                  `json:"age_ms_after"`
	AchievementScoreDelta     int64                  `json:"achievement_score_delta"`
	AddedNetworkSlots         []save.NetworkSlot     `json:"added_network_slots"`
	AddedLedgerFactKinds      []string               `json:"added_ledger_fact_kinds"`
	AddedLifetimeAchievements []string               `json:"added_lifetime_achievements"`
	ExitRecord                *founderExitRecordWire `json:"exit_record"`
	ResultFounderWireVersion  int                    `json:"result_founder_wire_version"`
	Rejection                 *founderAuditRejection `json:"rejection"`
	NextSoul                  *nextSoulWire          `json:"next_soul,omitempty"`
}

type nextSoulWire struct {
	SoulInitial int64  `json:"soul_initial"`
	BandMember  string `json:"band_member"`
}

type founderAuditRejection struct {
	Category string `json:"category"`
	Detail   string `json:"detail"`
}

type founderExitAuditReceipt struct {
	IntentID            string                 `json:"intent_id"`
	Outcome             string                 `json:"outcome"`
	FounderRevision     *int64                 `json:"founder_revision,omitempty"`
	CurrentRevision     *int64                 `json:"current_revision,omitempty"`
	ResultConstantsHash string                 `json:"result_constants_hash,omitempty"`
	Rejection           *founderAuditRejection `json:"rejection,omitempty"`
}

func buildFounderExitAudit(command save.ReplayCommand, founderRevision save.Revision, before, after *save.State, decision save.ExitDecision, catalogs CatalogBundle) (json.RawMessage, json.RawMessage, error) {
	if before == nil || after == nil || command.CompanyStreamID == "" || command.RunSeq < 1 || command.RunLogSeq < 1 {
		return nil, nil, ErrInvalidEngineState
	}
	resultHash := founderRevision.ConstantsHash
	facts := founderExitResolvedWire{Kind: founderExitResolvedKind, CompanyStreamID: command.CompanyStreamID,
		RunSeq: command.RunSeq, RunLogSeq: command.RunLogSeq, ResultConstantsHash: resultHash,
		Outcome:     string(decision.Outcome),
		AgeMSBefore: before.AgeMS, AgeMSAfter: before.AgeMS, AddedNetworkSlots: []save.NetworkSlot{},
		AddedLedgerFactKinds: []string{}, AddedLifetimeAchievements: []string{},
		ResultFounderWireVersion: save.VersionForState(before)}
	receipt := founderExitAuditReceipt{IntentID: command.IntentID, Outcome: string(decision.Outcome)}
	if decision.Outcome == save.IntentRejected {
		category, detail, err := rejectionFromReceipt(decision.Receipt, command.IntentID)
		if err != nil {
			return nil, nil, err
		}
		current := founderRevision.Number
		receipt.CurrentRevision, receipt.Rejection = &current, &founderAuditRejection{Category: category, Detail: detail}
		facts.Rejection = &founderAuditRejection{Category: category, Detail: detail}
	} else if decision.Outcome == save.IntentApplied {
		if decision.NewConstantsHash == "" || len(after.ExitHistory) != len(before.ExitHistory)+1 ||
			after.ReputationLevel < before.ReputationLevel || after.RouteKnowledgeBalance < before.RouteKnowledgeBalance ||
			after.AgeMS < before.AgeMS || after.AchievementScoreLifetime < before.AchievementScoreLifetime {
			return nil, nil, ErrInvalidEngineState
		}
		facts.ResultConstantsHash = decision.NewConstantsHash
		facts.ReputationDelta = after.ReputationLevel - before.ReputationLevel
		facts.RouteKnowledgeDelta = after.RouteKnowledgeBalance - before.RouteKnowledgeBalance
		facts.AttendedMS = after.AgeMS - before.AgeMS
		facts.AgeMSAfter = after.AgeMS
		facts.AchievementScoreDelta = after.AchievementScoreLifetime - before.AchievementScoreLifetime
		facts.AddedNetworkSlots = addedNetworkSlots(before.NetworkSlots, after.NetworkSlots)
		facts.AddedLedgerFactKinds = addedBoolKeys(before.LedgerFactKinds, after.LedgerFactKinds)
		facts.AddedLifetimeAchievements = addedBoolKeys(before.AchievementsEarnedLifetime, after.AchievementsEarnedLifetime)
		last := after.ExitHistory[len(after.ExitHistory)-1]
		facts.ExitRecord = &founderExitRecordWire{RunID: last.RunID, ExitType: last.ExitType,
			OccurredAtMS: last.OccurredAt.UnixMilli(), ReputationDelta: last.ReputationDelta}
		facts.ResultFounderWireVersion = save.VersionForState(after)
		if save.VersionForState(before) < 20 && facts.ResultFounderWireVersion >= 20 {
			resultCatalogs := catalogs
			if decision.NewConstantsHash != catalogs.ConstantsHash {
				if catalogs.Next == nil || catalogs.Next.ConstantsHash != decision.NewConstantsHash {
					return nil, nil, ErrInvalidEngineState
				}
				resultCatalogs = *catalogs.Next
			}
			if resultCatalogs.Soul == nil {
				return nil, nil, ErrInvalidEngineState
			}
			band, ok := resultCatalogs.Soul.BandFor(after.Soul)
			if !ok || after.Soul != resultCatalogs.Soul.Policy.Initial {
				return nil, nil, ErrInvalidEngineState
			}
			facts.NextSoul = &nextSoulWire{SoulInitial: after.Soul, BandMember: string(band.Member)}
		}
		next := founderRevision.Number + 1
		receipt.FounderRevision, receipt.ResultConstantsHash = &next, decision.NewConstantsHash
	} else {
		return nil, nil, ErrInvalidEngineState
	}
	if facts.AgeMSBefore < 0 || facts.AgeMSAfter < facts.AgeMSBefore || facts.AgeMSAfter > decimal.MaxExactInteger ||
		facts.AttendedMS != facts.AgeMSAfter-facts.AgeMSBefore {
		return nil, nil, ErrInvalidEngineState
	}
	resolved, err := json.Marshal(facts)
	if err != nil {
		return nil, nil, err
	}
	auditReceipt, err := json.Marshal(receipt)
	return resolved, auditReceipt, err
}

func verifyFounderExitLiveParity(command save.ReplayCommand, founderRevision save.Revision, request IntentRequest,
	before, wantAfter *save.State, decision save.ExitDecision, resolvedJSON, receiptJSON json.RawMessage,
	catalogs CatalogBundle,
) error {
	var resolved founderExitResolvedWire
	if err := decodeReplayStrict(resolvedJSON, &resolved); err != nil {
		return err
	}
	working, err := cloneFounderReplayState(before, catalogs.Economy)
	if err != nil {
		return err
	}
	founderCommand := save.FounderReplayCommand{IntentID: command.IntentID,
		FounderStreamID: founderRevision.StreamID, FounderID: founderRevision.OwnerID,
		Revision: founderRevision.Number, FounderLogSeq: 1, ServerTSMS: 1}
	replayed, err := applyFounderExitResolved(working, founderCommand, request, catalogs, resolved)
	if err != nil || replayed.Outcome != decision.Outcome || !jsonSemanticallyEqual(replayed.Receipt, receiptJSON) ||
		len(replayed.Events) != len(decision.FounderEvents) {
		return fmt.Errorf("%w: live Founder Exit parity", ErrInvalidEngineState)
	}
	if decision.Outcome == save.IntentApplied {
		wantState, cloneErr := cloneFounderReplayState(before, catalogs.Economy)
		if cloneErr != nil || applyFounderReplayOutput(wantState, wantAfter) != nil {
			return fmt.Errorf("%w: live Founder Exit expected state", ErrInvalidEngineState)
		}
		got, gotErr := save.EncodeState(replayed.State)
		want, wantErr := save.EncodeState(wantState)
		if gotErr != nil || wantErr != nil || !bytes.Equal(got, want) {
			return fmt.Errorf("%w: live Founder Exit state parity at %s (got=%v want=%v)", ErrInvalidEngineState,
				firstStateDifference(got, want), gotErr, wantErr)
		}
	}
	for index := range replayed.Events {
		if replayed.Events[index].Kind != decision.FounderEvents[index].Kind ||
			replayed.Events[index].SchemaVersion != decision.FounderEvents[index].SchemaVersion ||
			replayed.Events[index].IntentID != decision.FounderEvents[index].IntentID ||
			!jsonSemanticallyEqual(replayed.Events[index].Payload, decision.FounderEvents[index].Payload) {
			return fmt.Errorf("%w: live Founder Exit event parity", ErrInvalidEngineState)
		}
	}
	return nil
}

func firstStateDifference(left, right []byte) string {
	var leftFields, rightFields map[string]json.RawMessage
	if json.Unmarshal(left, &leftFields) != nil || json.Unmarshal(right, &rightFields) != nil {
		return "invalid_json"
	}
	keys := make([]string, 0, len(leftFields)+len(rightFields))
	seen := map[string]bool{}
	for key := range leftFields {
		seen[key], keys = true, append(keys, key)
	}
	for key := range rightFields {
		if !seen[key] {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	for _, key := range keys {
		if !bytes.Equal(leftFields[key], rightFields[key]) {
			return key
		}
	}
	return "encoding"
}

func rejectionFromReceipt(data json.RawMessage, intentID string) (string, string, error) {
	var wire struct {
		IntentID        string                `json:"intent_id"`
		Outcome         string                `json:"outcome"`
		CurrentRevision int64                 `json:"current_revision"`
		Rejection       founderAuditRejection `json:"rejection"`
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&wire); err != nil || wire.IntentID != intentID || wire.Outcome != "rejected" ||
		wire.Rejection.Category == "" || wire.Rejection.Detail == "" {
		return "", "", fmt.Errorf("%w: rejected Exit receipt", ErrInvalidEngineState)
	}
	return wire.Rejection.Category, wire.Rejection.Detail, nil
}

func addedBoolKeys(before, after map[string]bool) []string {
	result := make([]string, 0)
	for key, enabled := range after {
		if enabled && !before[key] {
			result = append(result, key)
		}
	}
	sort.Strings(result)
	return result
}

func addedNetworkSlots(before, after []save.NetworkSlot) []save.NetworkSlot {
	present := make(map[save.NetworkSlot]bool, len(before))
	for _, slot := range before {
		present[slot] = true
	}
	result := make([]save.NetworkSlot, 0)
	for _, slot := range after {
		if !present[slot] {
			result = append(result, slot)
		}
	}
	sort.Slice(result, func(left, right int) bool {
		if result[left].Slot != result[right].Slot {
			return result[left].Slot < result[right].Slot
		}
		return result[left].CarriedRef < result[right].CarriedRef
	})
	return result
}

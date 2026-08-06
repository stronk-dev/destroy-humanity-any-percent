package production

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"cloud-clicker/server/activeplay"
	"cloud-clicker/server/decimal"
	"cloud-clicker/server/economy"
	"cloud-clicker/server/multiplier"
	prestigecore "cloud-clicker/server/prestige"
	"cloud-clicker/server/save"
)

type activePlayExpiredBuff struct {
	BuffInstanceID string `json:"buff_instance_id"`
}

type activePlaySpawnEvidence struct {
	Sequence            int64   `json:"sequence"`
	SampledIntervalMS   int64   `json:"sampled_interval_ms"`
	EffectDraw          string  `json:"effect_draw"`
	GeneratorDraw       *string `json:"generator_draw"`
	EffectRowID         string  `json:"effect_row_id"`
	SelectedGeneratorID *string `json:"selected_generator_id"`
	OpportunityID       string  `json:"opportunity_id"`
	SpawnedAttendedMS   int64   `json:"spawned_attended_ms"`
	ExpiresAttendedMS   int64   `json:"expires_attended_ms"`
}

type activePlayScheduleEvidence struct {
	AttendedNowMS           int64                    `json:"attended_now_ms"`
	BeforeSequence          int64                    `json:"before_sequence"`
	BeforeNextOpportunityMS int64                    `json:"before_next_opportunity_attended_ms"`
	AfterSequence           int64                    `json:"after_sequence"`
	AfterNextOpportunityMS  int64                    `json:"after_next_opportunity_attended_ms"`
	ExpiredBuffs            []activePlayExpiredBuff  `json:"expired_buffs"`
	MissedOpportunityID     *string                  `json:"missed_opportunity_id"`
	Spawned                 *activePlaySpawnEvidence `json:"spawned"`
	Claim                   *activePlayClaimEvidence `json:"claim"`
}

type activePlayClaimEvidence struct {
	OpportunityID             string  `json:"opportunity_id"`
	EffectRowID               string  `json:"effect_row_id"`
	SelectedTarget            *string `json:"selected_target"`
	BuffInstanceID            *string `json:"buff_instance_id"`
	RequestedDelta            *string `json:"requested_delta"`
	ActualCreditedDelta       *string `json:"actual_credited_delta"`
	Saturated                 *bool   `json:"saturated"`
	CapReasonKey              *string `json:"cap_reason_key"`
	NextSampledIntervalMS     int64   `json:"next_sampled_interval_ms"`
	NextOpportunityAttendedMS int64   `json:"next_opportunity_attended_ms"`
}

func initializeActivePlayState(state *save.State, catalog *activeplay.Catalog, founderID string) (*activePlaySpawnEvidence, error) {
	if state == nil || catalog == nil || founderID == "" || state.RunSeq < 1 {
		return nil, ErrInvalidEngineState
	}
	spawn, err := catalog.Spawn(founderID, state.RunSeq, 0, 0)
	if err != nil {
		return nil, err
	}
	state.WireVersion = 18
	state.OpportunitySpawnSeq = 0
	state.NextOpportunityAttendedMS = spawn.SpawnedAttendedMS
	state.PendingOpportunity = nil
	state.ActiveBuffs = []save.ActiveBuff{}
	return spawnEvidence(spawn), nil
}

func spawnEvidence(spawn activeplay.Spawn) *activePlaySpawnEvidence {
	return &activePlaySpawnEvidence{Sequence: spawn.Sequence, SampledIntervalMS: spawn.SampledIntervalMS, EffectDraw: strconv.FormatUint(spawn.EffectDraw, 10),
		GeneratorDraw: uint64StringPointer(spawn.GeneratorDraw), EffectRowID: spawn.EffectRowID, SelectedGeneratorID: activeNullableString(spawn.SelectedGenerator),
		OpportunityID: spawn.OpportunityID, SpawnedAttendedMS: spawn.SpawnedAttendedMS, ExpiresAttendedMS: spawn.ExpiresAttendedMS}
}

func resolveActivePlaySchedule(state *save.State, catalog *activeplay.Catalog, policy *prestigecore.Policy, founderID string, now time.Time) (activePlayScheduleEvidence, error) {
	if state == nil || catalog == nil || policy == nil || founderID == "" || state.WireVersion != 18 {
		return activePlayScheduleEvidence{}, ErrInvalidEngineState
	}
	clone := *state
	clone.OfflineSpans = append([]save.OfflineSpan(nil), state.OfflineSpans...)
	clone.ActiveBuffs = cloneActiveBuffRows(state.ActiveBuffs)
	clone.PendingOpportunity = clonePendingOpportunityRow(state.PendingOpportunity)
	effectiveNow := save.CanonicalServerTime(now)
	if effectiveNow.After(clone.EvaluatedThrough) {
		if err := prestigecore.RecordOfflineSpan(&clone, clone.EvaluatedThrough, effectiveNow, policy.CatchupCeilingMS); err != nil {
			return activePlayScheduleEvidence{}, err
		}
	}
	attended, err := prestigecore.AttendedMS(&clone, effectiveNow)
	if err != nil {
		return activePlayScheduleEvidence{}, err
	}
	evidence := activePlayScheduleEvidence{AttendedNowMS: attended, BeforeSequence: state.OpportunitySpawnSeq,
		BeforeNextOpportunityMS: state.NextOpportunityAttendedMS, ExpiredBuffs: []activePlayExpiredBuff{}}
	working := *state
	working.ActiveBuffs = cloneActiveBuffRows(state.ActiveBuffs)
	working.PendingOpportunity = clonePendingOpportunityRow(state.PendingOpportunity)
	if err := advanceActivePlaySchedule(&working, catalog, founderID, evidence.AttendedNowMS, &evidence, true); err != nil {
		return activePlayScheduleEvidence{}, err
	}
	evidence.AfterSequence = working.OpportunitySpawnSeq
	evidence.AfterNextOpportunityMS = working.NextOpportunityAttendedMS
	return evidence, nil
}

func applyActivePlaySchedule(state *save.State, catalog *activeplay.Catalog, policy *prestigecore.Policy, founderID string, now time.Time, evidence activePlayScheduleEvidence) ([]save.EventWrite, error) {
	if state == nil || catalog == nil || policy == nil || founderID == "" || state.WireVersion != 18 ||
		evidence.BeforeSequence != state.OpportunitySpawnSeq || evidence.BeforeNextOpportunityMS != state.NextOpportunityAttendedMS {
		return nil, ErrInvalidReplayInputs
	}
	clone := *state
	clone.OfflineSpans = append([]save.OfflineSpan(nil), state.OfflineSpans...)
	effectiveNow := save.CanonicalServerTime(now)
	if effectiveNow.After(clone.EvaluatedThrough) {
		if err := prestigecore.RecordOfflineSpan(&clone, clone.EvaluatedThrough, effectiveNow, policy.CatchupCeilingMS); err != nil {
			return nil, err
		}
	}
	attended, err := prestigecore.AttendedMS(&clone, effectiveNow)
	if err != nil || attended != evidence.AttendedNowMS {
		return nil, ErrInvalidReplayInputs
	}
	if err := advanceActivePlaySchedule(state, catalog, founderID, attended, &evidence, false); err != nil {
		return nil, err
	}
	if state.OpportunitySpawnSeq != evidence.AfterSequence || state.NextOpportunityAttendedMS != evidence.AfterNextOpportunityMS {
		return nil, ErrInvalidReplayInputs
	}
	return activePlayScheduleEvents(evidence), nil
}

func advanceActivePlaySchedule(state *save.State, catalog *activeplay.Catalog, founderID string, attendedNow int64, evidence *activePlayScheduleEvidence, generate bool) error {
	if state == nil || evidence == nil || attendedNow < 0 || state.RunSeq < 1 {
		return ErrInvalidReplayInputs
	}
	kept := make([]save.ActiveBuff, 0, len(state.ActiveBuffs))
	expired := make([]activePlayExpiredBuff, 0)
	for _, buff := range state.ActiveBuffs {
		if buff.ExpiresAttendedMS <= attendedNow {
			expired = append(expired, activePlayExpiredBuff{BuffInstanceID: buff.BuffInstanceID})
		} else {
			kept = append(kept, buff)
		}
	}
	sort.Slice(expired, func(i, j int) bool { return expired[i].BuffInstanceID < expired[j].BuffInstanceID })
	if generate {
		evidence.ExpiredBuffs = expired
	} else if !equalExpiredBuffs(expired, evidence.ExpiredBuffs) {
		return ErrInvalidReplayInputs
	}
	state.ActiveBuffs = kept

	transitions := int64(0)
	var missed *string
	var spawned *activePlaySpawnEvidence
	for {
		if transitions >= catalog.Schedule.MaxDueTransitions {
			return fmt.Errorf("%w: active-play due-transition ceiling", ErrInvalidReplayInputs)
		}
		if state.PendingOpportunity != nil {
			if state.PendingOpportunity.ExpiresAttendedMS > attendedNow {
				break
			}
			value := state.PendingOpportunity.OpportunityID
			missed = &value
			from := state.PendingOpportunity.ExpiresAttendedMS
			state.PendingOpportunity = nil
			next, err := catalog.Spawn(founderID, state.RunSeq, state.OpportunitySpawnSeq, from)
			if err != nil {
				return err
			}
			state.NextOpportunityAttendedMS = next.SpawnedAttendedMS
			transitions++
			continue
		}
		if state.NextOpportunityAttendedMS == 0 || state.NextOpportunityAttendedMS > attendedNow {
			break
		}
		sequence := state.OpportunitySpawnSeq
		probe, err := catalog.Spawn(founderID, state.RunSeq, sequence, 0)
		if err != nil || probe.SampledIntervalMS > state.NextOpportunityAttendedMS {
			return ErrInvalidReplayInputs
		}
		from := state.NextOpportunityAttendedMS - probe.SampledIntervalMS
		spawn, err := catalog.Spawn(founderID, state.RunSeq, sequence, from)
		if err != nil || spawn.SpawnedAttendedMS != state.NextOpportunityAttendedMS || sequence >= decimal.MaxExactInteger {
			return ErrInvalidReplayInputs
		}
		selected := activeNullableString(spawn.SelectedGenerator)
		candidate := activePlaySpawnEvidence{Sequence: spawn.Sequence, SampledIntervalMS: spawn.SampledIntervalMS,
			EffectDraw: strconv.FormatUint(spawn.EffectDraw, 10), GeneratorDraw: uint64StringPointer(spawn.GeneratorDraw), EffectRowID: spawn.EffectRowID,
			SelectedGeneratorID: selected, OpportunityID: spawn.OpportunityID, SpawnedAttendedMS: spawn.SpawnedAttendedMS,
			ExpiresAttendedMS: spawn.ExpiresAttendedMS}
		if generate {
			spawned = &candidate
		} else if evidence.Spawned == nil || !equalSpawnEvidence(candidate, *evidence.Spawned) {
			return ErrInvalidReplayInputs
		} else {
			spawned = &candidate
		}
		state.PendingOpportunity = &save.PendingOpportunity{OpportunityID: spawn.OpportunityID, SpawnedAttendedMS: spawn.SpawnedAttendedMS,
			ExpiresAttendedMS: spawn.ExpiresAttendedMS, EffectRowID: spawn.EffectRowID, SelectedGeneratorID: selected}
		state.OpportunitySpawnSeq++
		state.NextOpportunityAttendedMS = 0
		transitions++
		continue
	}
	if generate {
		evidence.MissedOpportunityID = missed
		evidence.Spawned = spawned
	} else if !equalNullableString(missed, evidence.MissedOpportunityID) || (spawned == nil) != (evidence.Spawned == nil) {
		return ErrInvalidReplayInputs
	}
	return nil
}

func activePlayScheduleEvents(evidence activePlayScheduleEvidence) []save.EventWrite {
	events := make([]save.EventWrite, 0, len(evidence.ExpiredBuffs)+2)
	for _, expired := range evidence.ExpiredBuffs {
		payload, _ := json.Marshal(map[string]any{"buff_instance_id": expired.BuffInstanceID, "attended_ms": evidence.AttendedNowMS})
		events = append(events, save.EventWrite{Kind: save.EventBuffExpired, SchemaVersion: 1, Payload: payload})
	}
	if evidence.MissedOpportunityID != nil {
		payload, _ := json.Marshal(map[string]any{"opportunity_id": *evidence.MissedOpportunityID, "attended_ms": evidence.AttendedNowMS})
		events = append(events, save.EventWrite{Kind: save.EventOpportunityExpired, SchemaVersion: 1, Payload: payload})
	}
	if evidence.Spawned != nil {
		payload, _ := json.Marshal(map[string]any{"opportunity_id": evidence.Spawned.OpportunityID, "spawned_attended_ms": evidence.Spawned.SpawnedAttendedMS,
			"expires_attended_ms": evidence.Spawned.ExpiresAttendedMS, "effect_row_id": evidence.Spawned.EffectRowID,
			"selected_generator_id": evidence.Spawned.SelectedGeneratorID})
		events = append(events, save.EventWrite{Kind: save.EventOpportunitySpawned, SchemaVersion: 1, Payload: payload})
	}
	return events
}

func activePlayContributions(state *save.State, catalog *activeplay.Catalog, attendedNow int64) ([]multiplier.Contribution, error) {
	if state == nil || catalog == nil || attendedNow < 0 {
		return nil, ErrInvalidEngineState
	}
	result := make([]multiplier.Contribution, 0)
	for _, buff := range state.ActiveBuffs {
		if buff.ExpiresAttendedMS <= attendedNow {
			continue
		}
		effect, ok := catalog.Effect(buff.EffectRowID)
		if !ok {
			return nil, ErrInvalidEngineState
		}
		source := "active_play." + effect.ID + "." + buff.BuffInstanceID
		switch effect.Kind {
		case "production_frenzy":
			result = append(result, multiplier.Contribution{Slot: multiplier.SlotEventBuffs, SourceID: source, Target: "all", Factor: effect.Factor})
		case "click_frenzy":
			for _, action := range effect.ActionIDs {
				result = append(result, multiplier.Contribution{Slot: multiplier.SlotEventBuffs, SourceID: source, Target: action, Factor: effect.Factor})
			}
		case "building_special":
			if buff.SelectedTarget == nil {
				return nil, ErrInvalidEngineState
			}
			owned, ok := state.GeneratorCounts[*buff.SelectedTarget]
			if !ok || owned < 0 {
				return nil, ErrInvalidEngineState
			}
			factor, err := countPPMFactor(owned, effect.PerOwnedPPM)
			if err != nil {
				return nil, err
			}
			result = append(result, multiplier.Contribution{Slot: multiplier.SlotEventBuffs, SourceID: source, Target: *buff.SelectedTarget, Factor: factor})
		default:
			return nil, ErrInvalidEngineState
		}
	}
	sort.Slice(result, func(i, j int) bool { return contributionKey(result[i]) < contributionKey(result[j]) })
	return clampActiveContributionProducts(result, catalog.Combo.Cap), nil
}

func resolveActiveClaimEvidence(state *save.State, catalogs CatalogBundle, revision save.Revision, request IntentRequest,
	mode EvaluationMode, now time.Time, accrual replayAccrual, schedule activePlayScheduleEvidence) (*activePlayClaimEvidence, error) {
	if request.Kind != IntentClaimOpportunity {
		return nil, nil
	}
	clone, err := cloneReplayState(state, catalogs.Economy)
	if err != nil {
		return nil, err
	}
	if _, err := applyActivePlaySchedule(clone, catalogs.Opportunities, catalogs.Prestige, revision.OwnerID, now, schedule); err != nil {
		return nil, err
	}
	external, err := contributionsFromReplay(accrual)
	if err != nil {
		return nil, err
	}
	if err := applyReplayGuildSettlements(clone, accrual.GuildSettlementBatch, catalogs.Faction.StockCap); err != nil {
		return nil, err
	}
	active, err := activePlayContributions(clone, catalogs.Opportunities, schedule.AttendedNowMS)
	if err != nil {
		return nil, err
	}
	contributions, err := assembleContributions(clone, catalogs.Economy, append(external, active...))
	if err != nil {
		return nil, err
	}
	hook := closedReplayAccrualHook(catalogs, accrual.CommonsWeightPPM)
	generated := &activePlayClaimEvidence{}
	decision, err := applyClaimOpportunity(request, clone, catalogs, revision, mode, now, contributions, hook,
		schedule.AttendedNowMS, schedule.MissedOpportunityID, generated, true)
	if err != nil {
		return nil, err
	}
	if decision.Outcome == save.IntentRejected {
		return nil, nil
	}
	return generated, nil
}

func applyClaimOpportunity(request IntentRequest, state *save.State, catalogs CatalogBundle, revision save.Revision, mode EvaluationMode,
	now time.Time, contributions []multiplier.Contribution, hook AccrualHook, attendedNow int64, missedOpportunityID *string,
	evidence *activePlayClaimEvidence, generate bool) (save.IntentDecision, error) {
	if state == nil || state.Ledger == nil || catalogs.Opportunities == nil || attendedNow < 0 {
		return save.IntentDecision{}, ErrInvalidEngineState
	}
	before := state.Ledger.Snapshot()
	evaluation, err := evaluateWithSimulationPolicy(state, catalogs.Economy, now, mode, contributions, nil)
	if err != nil {
		return save.IntentDecision{}, err
	}
	events, err := runAccrualHook(hook, request.IntentID, state, catalogs.Economy, revision, evaluation, contributions)
	if err != nil {
		return save.IntentDecision{}, err
	}
	postAccrual := state.Ledger.Snapshot()
	if state.PendingOpportunity == nil {
		if missedOpportunityID != nil && *missedOpportunityID == request.OpportunityID {
			return rejectedDecision(request, revision.Number, "not_eligible", "opportunity_expired")
		}
		return rejectedDecision(request, revision.Number, "not_eligible", "opportunity_not_pending")
	}
	if state.PendingOpportunity.OpportunityID != request.OpportunityID {
		return rejectedDecision(request, revision.Number, "unknown_id", "opportunity_id")
	}
	pending := *state.PendingOpportunity
	effect, ok := catalogs.Opportunities.Effect(pending.EffectRowID)
	if !ok || pending.ExpiresAttendedMS <= attendedNow {
		return save.IntentDecision{}, ErrInvalidEngineState
	}
	if evidence == nil {
		return save.IntentDecision{}, fmt.Errorf("%w: missing applied active-play claim resolution", ErrInvalidReplayInputs)
	}
	claim := activePlayClaimEvidence{OpportunityID: pending.OpportunityID, EffectRowID: pending.EffectRowID,
		SelectedTarget: cloneStringPointer(pending.SelectedGeneratorID)}
	state.PendingOpportunity = nil
	next, nextErr := catalogs.Opportunities.Spawn(revision.OwnerID, state.RunSeq, state.OpportunitySpawnSeq, attendedNow)
	if nextErr != nil {
		return save.IntentDecision{}, nextErr
	}
	state.NextOpportunityAttendedMS = next.SpawnedAttendedMS
	claim.NextSampledIntervalMS = next.SampledIntervalMS
	claim.NextOpportunityAttendedMS = next.SpawnedAttendedMS
	claimPayload := map[string]any{"opportunity_id": pending.OpportunityID, "effect_row_id": pending.EffectRowID,
		"selected_target": pending.SelectedGeneratorID}
	var effectEvents []save.EventWrite
	switch effect.Kind {
	case "production_frenzy", "click_frenzy", "building_special":
		buffID := catalogs.Opportunities.BuffID(revision.OwnerID, state.RunSeq, state.OpportunitySpawnSeq-1, attendedNow)
		claim.BuffInstanceID = &buffID
		target := cloneStringPointer(pending.SelectedGeneratorID)
		buff := save.ActiveBuff{BuffInstanceID: buffID, EffectRowID: effect.ID, SelectedTarget: target,
			ActivatedAttendedMS: attendedNow, ExpiresAttendedMS: attendedNow + effect.DurationMS}
		state.ActiveBuffs = append(state.ActiveBuffs, buff)
		sort.Slice(state.ActiveBuffs, func(i, j int) bool { return state.ActiveBuffs[i].BuffInstanceID < state.ActiveBuffs[j].BuffInstanceID })
		claimPayload["buff_instance_id"] = buffID
		payload, _ := json.Marshal(map[string]any{"buff_instance_id": buffID, "effect_row_id": effect.ID, "selected_target": target,
			"activated_attended_ms": attendedNow, "expires_attended_ms": buff.ExpiresAttendedMS})
		effectEvents = append(effectEvents, save.EventWrite{Kind: save.EventBuffStarted, SchemaVersion: 1, IntentID: request.IntentID, Payload: payload})
	case "lucky_payout":
		bank, exists := state.Ledger.Balance(effect.ResourceID)
		if !exists {
			return save.IntentDecision{}, ErrInvalidEngineState
		}
		rates, rateErr := ratesWithProvisionedAndPolicy(catalogs.Economy, state.GeneratorCounts, state.GeneratorProvisioned, contributions, nil)
		if rateErr != nil {
			return save.IntentDecision{}, rateErr
		}
		rate := decimal.SumDeterministic(rates[effect.ResourceID])
		bankTerm := effect.LuckyBankFrac.Mul(bank).Quantize(decimal.CanonicalSignificantDigits)
		rateTerm := effect.LuckyRateCap.Mul(rate).Quantize(decimal.CanonicalSignificantDigits)
		minimum := bankTerm
		if rateTerm.Lt(minimum) {
			minimum = rateTerm
		}
		requested := minimum.Add(effect.Epsilon).Quantize(decimal.CanonicalSignificantDigits)
		receipt, applyErr := state.Ledger.ApplyAccrual(economy.Transaction{Entries: []economy.Entry{{ResourceID: effect.ResourceID, Delta: requested}}})
		if applyErr != nil {
			return save.IntentDecision{}, applyErr
		}
		actual := decimal.Zero
		for _, change := range receipt.Changes {
			if change.ResourceID == effect.ResourceID {
				actual, applyErr = decimal.ParseCanonical(change.Delta)
			}
		}
		if applyErr != nil {
			return save.IntentDecision{}, applyErr
		}
		saturated := !actual.Eq(requested)
		claim.RequestedDelta, claim.ActualCreditedDelta, claim.Saturated = stringPointer(requested.String()), stringPointer(actual.String()), boolPointer(saturated)
		if saturated {
			claim.CapReasonKey = stringPointer(effect.HardcapReasonKey)
		}
		claimPayload["requested_delta"], claimPayload["actual_credited_delta"], claimPayload["saturated"], claimPayload["cap_reason_key"] = requested.String(), actual.String(), saturated, claim.CapReasonKey
	default:
		return save.IntentDecision{}, ErrInvalidEngineState
	}
	if generate {
		*evidence = claim
	} else if !equalClaimEvidence(claim, *evidence) {
		return save.IntentDecision{}, ErrInvalidReplayInputs
	}
	claimEncoded, _ := json.Marshal(claimPayload)
	events = append(events, save.EventWrite{Kind: save.EventOpportunityClaimed, SchemaVersion: 1, IntentID: request.IntentID, Payload: claimEncoded})
	events = append(events, effectEvents...)
	decision, err := appliedDecisionWithActionDebits(request, state, catalogs.Economy, revision.Number+1, 1, before, postAccrual, events, nil)
	if err != nil {
		return save.IntentDecision{}, err
	}
	decision.Receipt, err = addOpportunityReceipt(decision.Receipt, claim)
	return decision, err
}

func addOpportunityReceipt(receipt json.RawMessage, claim activePlayClaimEvidence) (json.RawMessage, error) {
	var root map[string]any
	if json.Unmarshal(receipt, &root) != nil {
		return nil, ErrInvalidEngineState
	}
	inner, ok := root["receipt"].(map[string]any)
	if !ok {
		return nil, ErrInvalidEngineState
	}
	inner["opportunity"] = claim
	return json.Marshal(root)
}

func equalClaimEvidence(left, right activePlayClaimEvidence) bool {
	return left.OpportunityID == right.OpportunityID && left.EffectRowID == right.EffectRowID && equalNullableString(left.SelectedTarget, right.SelectedTarget) &&
		equalNullableString(left.BuffInstanceID, right.BuffInstanceID) && equalNullableString(left.RequestedDelta, right.RequestedDelta) &&
		equalNullableString(left.ActualCreditedDelta, right.ActualCreditedDelta) && equalBoolPointer(left.Saturated, right.Saturated) && equalNullableString(left.CapReasonKey, right.CapReasonKey) &&
		left.NextSampledIntervalMS == right.NextSampledIntervalMS && left.NextOpportunityAttendedMS == right.NextOpportunityAttendedMS
}

func equalBoolPointer(left, right *bool) bool {
	return left == nil && right == nil || left != nil && right != nil && *left == *right
}
func cloneStringPointer(value *string) *string {
	if value == nil {
		return nil
	}
	result := *value
	return &result
}
func stringPointer(value string) *string { result := value; return &result }
func boolPointer(value bool) *bool       { result := value; return &result }

func clampActiveContributionProducts(values []multiplier.Contribution, cap decimal.Decimal) []multiplier.Contribution {
	result := append([]multiplier.Contribution(nil), values...)
	byTarget := map[string][]int{}
	for index := range result {
		byTarget[result[index].Target] = append(byTarget[result[index].Target], index)
	}
	for _, indexes := range byTarget {
		product := decimal.One
		for _, index := range indexes {
			candidate := product.Mul(result[index].Factor).Quantize(decimal.CanonicalSignificantDigits)
			if candidate.Gt(cap) {
				result[index].Factor = cap.Div(product).Quantize(decimal.CanonicalSignificantDigits)
				for _, later := range indexes {
					if later > index {
						result[later].Factor = decimal.One
					}
				}
				break
			}
			product = candidate
		}
	}
	return result
}

func activeDeclarationID(sourceID, target string) string {
	if !strings.HasPrefix(sourceID, "active_play.") {
		return sourceID
	}
	rest := strings.TrimPrefix(sourceID, "active_play.")
	parts := strings.Split(rest, ".")
	if len(parts) < 2 {
		return sourceID
	}
	// The UUID is always the first segment containing a dash.
	uuidIndex := -1
	for index, part := range parts {
		if strings.Contains(part, "-") {
			uuidIndex = index
			break
		}
	}
	if uuidIndex < 1 {
		return sourceID
	}
	effect := strings.Join(parts[:uuidIndex], ".")
	return effect
}

func activeNullableString(value string) *string {
	if value == "" {
		return nil
	}
	result := value
	return &result
}
func equalNullableString(left, right *string) bool {
	return left == nil && right == nil || left != nil && right != nil && *left == *right
}
func equalExpiredBuffs(left, right []activePlayExpiredBuff) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}
func equalSpawnEvidence(left, right activePlaySpawnEvidence) bool {
	return left.Sequence == right.Sequence && left.SampledIntervalMS == right.SampledIntervalMS && left.EffectDraw == right.EffectDraw && equalNullableString(left.GeneratorDraw, right.GeneratorDraw) && left.EffectRowID == right.EffectRowID && equalNullableString(left.SelectedGeneratorID, right.SelectedGeneratorID) && left.OpportunityID == right.OpportunityID && left.SpawnedAttendedMS == right.SpawnedAttendedMS && left.ExpiresAttendedMS == right.ExpiresAttendedMS
}
func uint64StringPointer(value *uint64) *string {
	if value == nil {
		return nil
	}
	result := strconv.FormatUint(*value, 10)
	return &result
}
func clonePendingOpportunityRow(value *save.PendingOpportunity) *save.PendingOpportunity {
	if value == nil {
		return nil
	}
	result := *value
	result.SelectedGeneratorID = activeNullableString(pointerString(value.SelectedGeneratorID))
	return &result
}
func cloneActiveBuffRows(values []save.ActiveBuff) []save.ActiveBuff {
	if values == nil {
		return nil
	}
	result := make([]save.ActiveBuff, len(values))
	copy(result, values)
	for i := range result {
		result[i].SelectedTarget = activeNullableString(pointerString(values[i].SelectedTarget))
	}
	return result
}
func pointerString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

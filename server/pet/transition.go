package pet

import (
	"errors"
	"math/big"
	"sort"
	"strconv"
	"strings"

	"cloud-clicker/server/fixedgrid"
)

var (
	ErrInvalidCareTransition = errors.New("invalid pet care transition")
	ErrStaleCareTransition   = errors.New("stale pet care transition")
)

type CareTransitionInput struct {
	ActionID         string `json:"action_id"`
	AttendedBeforeMS int64  `json:"attended_before_ms"`
	AttendedAfterMS  int64  `json:"attended_after_ms"`
}

type CareTransitionResult struct {
	State                  CareState           `json:"state"`
	Applied                bool                `json:"applied"`
	RejectionDetail        CareRejectionDetail `json:"rejection_detail"`
	StatID                 StatID              `json:"stat_id"`
	BeforePPM              int64               `json:"before_ppm"`
	AppliedPPM             int64               `json:"applied_ppm"`
	AfterPPM               int64               `json:"after_ppm"`
	TrustBeforePPM         int64               `json:"trust_before_ppm"`
	TrustAfterPPM          int64               `json:"trust_after_ppm"`
	Mood                   Mood                `json:"mood"`
	StatusBand             StatusBand          `json:"status_band"`
	StatusChanged          bool                `json:"status_changed"`
	NextEligibleAttendedMS int64               `json:"next_eligible_attended_ms"`
	EligibleActionIDs      []string            `json:"eligible_action_ids"`
}

func ApplyCareTransition(state CareState, catalog *Catalog, input CareTransitionInput) (CareTransitionResult, error) {
	if catalog == nil {
		return CareTransitionResult{}, ErrInvalidCareTransition
	}
	if input.AttendedBeforeMS != state.EvaluatedThroughAttendedMS || input.AttendedAfterMS < input.AttendedBeforeMS ||
		!exactNonnegative(input.AttendedAfterMS) {
		return CareTransitionResult{}, ErrStaleCareTransition
	}
	const fixturePetID = "01986666-0000-7000-8000-000000000001"
	if ValidateCareStatesForCatalog(map[string]CareState{fixturePetID: state}, catalog) != nil {
		return CareTransitionResult{}, ErrInvalidCareTransition
	}
	next := cloneCareStates(map[string]CareState{fixturePetID: state})[fixturePetID]
	statusBefore, _, err := CareStatus(next, catalog)
	if err != nil {
		return CareTransitionResult{}, err
	}
	if err := applyCareDecay(&next, catalog, input.AttendedAfterMS-input.AttendedBeforeMS); err != nil {
		return CareTransitionResult{}, err
	}
	if err := advanceBehavior(&next, catalog, input.AttendedBeforeMS, input.AttendedAfterMS); err != nil {
		return CareTransitionResult{}, err
	}
	next.EvaluatedThroughAttendedMS = input.AttendedAfterMS
	result := CareTransitionResult{State: next, RejectionDetail: RejectionUnknownAction}
	action, found := findAction(catalog, input.ActionID)
	if !found {
		return finishCareTransition(result, statusBefore, catalog, EventCareRejected)
	}
	result.StatID = action.StatID
	result.BeforePPM = next.StatsPPM[action.StatID]
	result.AfterPPM = result.BeforePPM
	result.TrustBeforePPM = next.TrustPPM
	result.TrustAfterPPM = next.TrustPPM
	result.NextEligibleAttendedMS = next.CooldownUntilAttendedMS[action.ActionID]
	switch {
	case result.NextEligibleAttendedMS > input.AttendedAfterMS:
		result.RejectionDetail = RejectionCooldown
		return finishCareTransition(result, statusBefore, catalog, EventCareRejected)
	case result.BeforePPM < action.MinEligiblePPM:
		result.RejectionDetail = RejectionIneligible
		return finishCareTransition(result, statusBefore, catalog, EventCareRejected)
	}
	effective := action.DeltaPPM
	if result.BeforePPM >= catalog.StatPolicy.DiminishingThresholdPPM {
		value := new(big.Int).Mul(big.NewInt(action.DeltaPPM), big.NewInt(catalog.StatPolicy.DiminishingFactorPPM))
		value.Quo(value, big.NewInt(1_000_000))
		effective = value.Int64()
	}
	headroom := int64(1_000_000) - result.BeforePPM
	if effective > headroom {
		result.AppliedPPM = headroom
	} else {
		result.AppliedPPM = effective
	}
	if result.AppliedPPM == 0 {
		result.RejectionDetail = RejectionSaturated
		return finishCareTransition(result, statusBefore, catalog, EventCareRejected)
	}
	if input.AttendedAfterMS > maxExactInteger-action.CooldownAttendedMS {
		return CareTransitionResult{}, ErrInvalidCareTransition
	}
	result.Applied = true
	result.RejectionDetail = ""
	result.AfterPPM = result.BeforePPM + result.AppliedPPM
	next.StatsPPM[action.StatID] = result.AfterPPM
	next.CooldownUntilAttendedMS[action.ActionID] = input.AttendedAfterMS + action.CooldownAttendedMS
	result.NextEligibleAttendedMS = next.CooldownUntilAttendedMS[action.ActionID]
	grant := catalog.TrustPolicy.GainPPMPerEffectiveAction
	if grant > catalog.TrustPolicy.CapPPM-next.TrustPPM {
		grant = catalog.TrustPolicy.CapPPM - next.TrustPPM
	}
	next.TrustPPM += grant
	result.TrustAfterPPM = next.TrustPPM
	result.State = next
	return finishCareTransition(result, statusBefore, catalog, EventCareApplied)
}

func applyCareDecay(state *CareState, catalog *Catalog, elapsedMS int64) error {
	for _, row := range catalog.StatPolicy.Stats {
		integrated, err := fixedgrid.Integrate(elapsedMS, row.DecayPPMPerGrid,
			state.StatDecayRemaindersPPM[row.StatID], catalog.StatPolicy.GridMS)
		if err != nil {
			return ErrInvalidCareTransition
		}
		headroom := state.StatsPPM[row.StatID] - row.FloorPPM
		if integrated.Whole.Cmp(big.NewInt(headroom)) >= 0 {
			state.StatsPPM[row.StatID], state.StatDecayRemaindersPPM[row.StatID] = row.FloorPPM, 0
		} else {
			state.StatsPPM[row.StatID] -= integrated.Whole.Int64()
			state.StatDecayRemaindersPPM[row.StatID] = integrated.Remainder
		}
	}
	integrated, err := fixedgrid.Integrate(elapsedMS, catalog.TrustPolicy.DecayPPMPerGrid,
		state.TrustDecayRemainderPPM, catalog.StatPolicy.GridMS)
	if err != nil {
		return ErrInvalidCareTransition
	}
	distance := state.TrustPPM - catalog.TrustPolicy.NeutralPPM
	if distance < 0 {
		distance = -distance
	}
	if integrated.Whole.Cmp(big.NewInt(distance)) >= 0 {
		state.TrustPPM, state.TrustDecayRemainderPPM = catalog.TrustPolicy.NeutralPPM, 0
	} else if state.TrustPPM > catalog.TrustPolicy.NeutralPPM {
		state.TrustPPM -= integrated.Whole.Int64()
		state.TrustDecayRemainderPPM = integrated.Remainder
	} else if state.TrustPPM < catalog.TrustPolicy.NeutralPPM {
		state.TrustPPM += integrated.Whole.Int64()
		state.TrustDecayRemainderPPM = integrated.Remainder
	} else {
		state.TrustDecayRemainderPPM = 0
	}
	return nil
}

func finishCareTransition(result CareTransitionResult, prior StatusBand, catalog *Catalog, event BehaviorEvent) (CareTransitionResult, error) {
	if err := applyBehaviorEvent(&result.State, catalog, event, result.State.EvaluatedThroughAttendedMS); err != nil {
		return CareTransitionResult{}, err
	}
	status, mood, err := CareStatus(result.State, catalog)
	if err != nil {
		return CareTransitionResult{}, err
	}
	result.StatusBand, result.Mood, result.StatusChanged = status, mood, status != prior
	result.EligibleActionIDs = EligibleCareActions(result.State, catalog, result.State.EvaluatedThroughAttendedMS)
	return result, nil
}

func CareStatus(state CareState, catalog *Catalog) (StatusBand, Mood, error) {
	if catalog == nil || len(catalog.MoodPolicy) != len(statusBands) {
		return "", "", ErrInvalidCareTransition
	}
	scalar := int64(1_000_000)
	for _, statID := range statIDs {
		if state.StatsPPM[statID] < scalar {
			scalar = state.StatsPPM[statID]
		}
	}
	selected := 0
	for index, row := range catalog.MoodPolicy {
		if row.FloorPPM <= scalar {
			selected = index
		}
	}
	return statusBands[selected], catalog.MoodPolicy[selected].MoodMember, nil
}

func EligibleCareActions(state CareState, catalog *Catalog, attendedMS int64) []string {
	if catalog == nil {
		return []string{}
	}
	result := make([]string, 0, len(catalog.Actions))
	for _, action := range catalog.Actions {
		if state.CooldownUntilAttendedMS[action.ActionID] <= attendedMS &&
			state.StatsPPM[action.StatID] >= action.MinEligiblePPM && state.StatsPPM[action.StatID] < 1_000_000 {
			result = append(result, action.ActionID)
		}
	}
	return result
}

func findAction(catalog *Catalog, actionID string) (ActionPolicy, bool) {
	index := sort.Search(len(catalog.Actions), func(index int) bool { return catalog.Actions[index].ActionID >= actionID })
	if index == len(catalog.Actions) || catalog.Actions[index].ActionID != actionID {
		return ActionPolicy{}, false
	}
	return catalog.Actions[index], true
}

func advanceBehavior(state *CareState, catalog *Catalog, before, after int64) error {
	if err := drainDue(state, before); err != nil {
		return err
	}
	nextGrid, ok := nextGridBoundary(before, catalog.StatPolicy.GridMS)
	if !ok {
		if after != before {
			return ErrInvalidCareTransition
		}
		return drainDue(state, after)
	}
	type cyclePoint struct {
		at      int64
		entered int64
	}
	seen := map[string]cyclePoint{}
	for nextGrid <= after {
		if !hasBehaviorCandidate(catalog, state.BehaviorState, EventGridTick) {
			due, exists := earliestDue(state.BehaviorQueue)
			if !exists || due > after {
				break
			}
			jump, valid := gridAtOrAfter(due, catalog.StatPolicy.GridMS)
			if !valid || jump > after {
				break
			}
			nextGrid = jump
		}
		if err := drainDue(state, nextGrid); err != nil || applyBehaviorEvent(state, catalog, EventGridTick, nextGrid) != nil {
			return ErrInvalidCareTransition
		}
		signature := behaviorCycleSignature(*state, nextGrid)
		if prior, exists := seen[signature]; exists {
			cycleMS := nextGrid - prior.at
			if cycleMS > 0 {
				cycles := (after - nextGrid) / cycleMS
				shiftEntered := state.BehaviorEnteredAtAttendedMS-prior.entered == cycleMS
				if state.BehaviorEnteredAtAttendedMS != prior.entered && !shiftEntered {
					return ErrInvalidCareTransition
				}
				for _, entry := range state.BehaviorQueue {
					if allowed := (maxExactInteger - entry.DueAttendedMS) / cycleMS; cycles > allowed {
						cycles = allowed
					}
				}
				if shiftEntered {
					if allowed := (maxExactInteger - state.BehaviorEnteredAtAttendedMS) / cycleMS; cycles > allowed {
						cycles = allowed
					}
				}
				if cycles > 0 {
					shift := cycles * cycleMS
					shiftBehaviorTimes(state, shift, shiftEntered)
					nextGrid += shift
				}
			}
		} else {
			seen[signature] = cyclePoint{at: nextGrid, entered: state.BehaviorEnteredAtAttendedMS}
		}
		if nextGrid > maxExactInteger-catalog.StatPolicy.GridMS {
			break
		}
		nextGrid += catalog.StatPolicy.GridMS
	}
	return drainDue(state, after)
}

func applyBehaviorEvent(state *CareState, catalog *Catalog, event BehaviorEvent, at int64) error {
	for _, row := range catalog.BehaviorPolicy {
		if row.FromState != state.BehaviorState || row.Event != event {
			continue
		}
		duration := new(big.Int).Mul(big.NewInt(row.DurationGridTicks), big.NewInt(catalog.StatPolicy.GridMS))
		due := new(big.Int).Add(big.NewInt(at), duration)
		if !due.IsInt64() || due.Int64() > maxExactInteger {
			return ErrInvalidCareTransition
		}
		queue := state.BehaviorQueue[:0]
		for _, entry := range state.BehaviorQueue {
			if entry.BehaviorID != string(row.ToState) {
				queue = append(queue, entry)
			}
		}
		if len(queue) >= BehaviorQueueHardcap {
			return ErrInvalidCareTransition
		}
		state.BehaviorQueue = append(queue, BehaviorQueueEntry{BehaviorID: string(row.ToState), DueAttendedMS: due.Int64()})
		sortBehaviorQueue(state.BehaviorQueue)
		return nil
	}
	return nil
}

func drainDue(state *CareState, through int64) error {
	sortBehaviorQueue(state.BehaviorQueue)
	index := 0
	for index < len(state.BehaviorQueue) && state.BehaviorQueue[index].DueAttendedMS <= through {
		entry := state.BehaviorQueue[index]
		candidate := BehaviorState(entry.BehaviorID)
		if !ValidBehaviorState(candidate) {
			return ErrInvalidCareTransition
		}
		state.BehaviorState, state.BehaviorEnteredAtAttendedMS = candidate, entry.DueAttendedMS
		index++
	}
	state.BehaviorQueue = append([]BehaviorQueueEntry(nil), state.BehaviorQueue[index:]...)
	return nil
}

func sortBehaviorQueue(queue []BehaviorQueueEntry) {
	sort.Slice(queue, func(left, right int) bool {
		if queue[left].DueAttendedMS != queue[right].DueAttendedMS {
			return queue[left].DueAttendedMS < queue[right].DueAttendedMS
		}
		return queue[left].BehaviorID < queue[right].BehaviorID
	})
}

func hasBehaviorCandidate(catalog *Catalog, state BehaviorState, event BehaviorEvent) bool {
	for _, row := range catalog.BehaviorPolicy {
		if row.FromState == state && row.Event == event {
			return true
		}
	}
	return false
}

func earliestDue(queue []BehaviorQueueEntry) (int64, bool) {
	if len(queue) == 0 {
		return 0, false
	}
	return queue[0].DueAttendedMS, true
}

func nextGridBoundary(value, grid int64) (int64, bool) {
	quotient := value / grid
	if quotient >= maxExactInteger/grid {
		return 0, false
	}
	return (quotient + 1) * grid, true
}

func gridAtOrAfter(value, grid int64) (int64, bool) {
	if value%grid == 0 {
		return value, true
	}
	if value > maxExactInteger-(grid-value%grid) {
		return 0, false
	}
	return value + grid - value%grid, true
}

func behaviorCycleSignature(state CareState, at int64) string {
	var builder strings.Builder
	builder.WriteString(string(state.BehaviorState))
	for _, entry := range state.BehaviorQueue {
		builder.WriteByte('|')
		builder.WriteString(entry.BehaviorID)
		builder.WriteByte(':')
		builder.WriteString(strconv.FormatInt(entry.DueAttendedMS-at, 10))
	}
	return builder.String()
}

func shiftBehaviorTimes(state *CareState, shift int64, shiftEntered bool) {
	if shiftEntered {
		state.BehaviorEnteredAtAttendedMS += shift
	}
	for index := range state.BehaviorQueue {
		state.BehaviorQueue[index].DueAttendedMS += shift
	}
}
